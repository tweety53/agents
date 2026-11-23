#!/usr/bin/env python3
"""PreToolUse hook: keep agents from writing into a repository's main checkout while it sits on
the default branch.

Every pipeline run works in a worktree. The main checkout on main/master/develop is the landing
target and a place to read from, never a workbench. A session that edits, stages, commits, stashes
or resets there leaves residue no later session owns — the agents repo carried 45 files of a
reset-away commit across sessions for a day, and both gymie repos carried the reverse image of a
landing as staged changes. This hook denies the tools that create such residue:

  * Edit / Write / MultiEdit / NotebookEdit on a path inside such a checkout;
  * Bash commands that run a mutating git verb there (`git commit`, `git -C <main> add`,
    `cd <main> && git reset ...`), or write into it with `sed -i`, `tee`, a `>`/`>>`
    redirect, `cp` destinations, `mv`/`rm` sources and destinations, or `trash` of a path git
    tracks there; removing a whole clone of a local working checkout is not a write to a main
    checkout (kan-1107 self-review, F5). A `cd` inside `( … )` or `$( … )` ends at its close, as bash scopes it
    (kan-1083 self-review, F2). A heredoc body is
    scanned like any other line, except where the whole command is one plain data write —
    `cat > ABS <<EOF`, `cat >> ABS <<EOF` or `tee [-a] ABS <<EOF` to a literal absolute path,
    its body and closing line, nothing after — where only ABS is judged — and except the body of
    any `cat`/`tee` heredoc inside a longer command, which is that writer's data, never commands.
    A `flow` heredoc body (`flow record summary -file - <<'EOF'`) is data when its opener writes
    no file and nothing else in the command runs code (kan-1025). A body no shell runs — a `python`/`python3`/`node` program on stdin, or a file the rest of
    the command runs only through one of those or `npx playwright test` — is that language's
    text: only the git invocations its lines name are judged (kan-1026: `d>0` read as a write).

"Main checkout" is the worktree whose git dir IS the common dir — a linked worktree's `--git-dir`
is `<common>/worktrees/<name>`, so `<repo>-worktrees/<change>` beside the main checkout passes,
as does a pre-migration `.worktrees/<change>` inside it. A path into a worktree that does not exist
yet passes too, when its nearest existing ancestor is either root. The deny reason suggests the
sibling `<dirname top>/<basename top>-worktrees/<change>` only. "Default branch" is what `origin/HEAD` points at; when the remote head is unknown, a checkout
of main, master or develop counts. A main checkout on a feature branch is not protected — leaving
the default branch is how work gets a branch of its own.

It also denies, in a linked worktree, a `git checkout … -- <path>` or `git restore <path>` naming
a path with unstaged edits: the restore reverts it to its index state and an uncommitted fix dies
with it. The reason points at flow-guard break-and-prove, which restores from a
byte snapshot.

A worktree listed in its change's state record (`flow state get`) is exempt from that denial.

It denies `git stash pop`, `apply`, `drop` and `branch` unless the stash entry they act on was made
on the current branch. A stash list is shared by every worktree of a repository, so a bare
`git stash pop` after a failed `git stash push` pops another change's work into this worktree
(kan-1092: kan-1091's WIP landed as conflict markers in four files). A call with no explicit
`stash@{N}` is denied outright; one naming an entry is allowed only when `git stash list` shows
"On <current branch>:" for it. The reason tells the agent to stop and report BLOCKED.

Allowed everywhere: reads, `git pull --ff-only`, `git fetch`, `git checkout -b`, `git switch -c`,
`git worktree add`, `git push`, and the landing script `flow-guard refresh-main-checkout` — it is not a
`git` verb in the command text, and owns its own main-checkout rules.

Denies, never rewrites: the PreToolUse contract cannot change tool input, so the deny reason
names the worktree command to use instead.

Fails open: any unexpected input, parse error, git failure or internal fault exits 0 and lets
the call through. A guardrail on hygiene must never be the reason work stops.

ponytail: Bash coverage is by token scan — a heredoc piped into python, a script that edits a
file, or a `find -exec` still writes. Widen the verb and writer lists when a new shape shows up
in a repo's reflog, rather than trying to model the shell. Variable expansion is the same
approximation, one step further: `NAME=value` assignments collected positionally from the
command text only, values kept literal, within-line use before an assignment left unexpanded —
never environment state, which a one-shot hook process cannot know.
"""

import glob
import json
import os
import re
import shlex
import subprocess
import sys

EDIT_TOOLS = {"Edit", "Write", "MultiEdit", "NotebookEdit"}

ASSIGNMENT_RE = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*=")

# git verbs that change the index, the worktree, a ref or the stash list.
DENIED_GIT_VERBS = {
    "add", "am", "apply", "cherry-pick", "clean", "commit", "merge", "mv", "rebase",
    "reset", "restore", "revert", "rm", "stash", "update-index", "update-ref",
}

FALLBACK_DEFAULT_BRANCHES = {"main", "master", "develop"}

# A bare `cat > <file>` with no input of its own: it reads stdin, which a Bash call
# never has, so it waits until the call times out.
# ponytail: a line scan — a heredoc body line opening with `cat > <file>` is denied
# too; parse heredocs out first if that ever bites.
BARE_CAT_RE = re.compile(r"(?:^|[;&(]|\|\|)\s*cat\s*>>?\s*[^\s<>|;&()]+\s*(?=$|[;&)]|\|\|)", re.M)
BARE_CAT_REASON = (
    "a bare `cat > <file>` reads stdin, which a Bash call never has, so it waits until the "
    "call times out — write the file with the Write tool, or give cat a heredoc "
    "(`cat > <file> <<'EOF'`)"
)
# A `go test` over a whole heavy package of the agents repo — the guard package (~5 minutes), the flow
# command or `./...` — with no `-run`: a session ran the whole guard package to check a six-second
# change. The operator's own `!` run is not a tool call, so it is never judged; a session names the
# tests it touched with `-run`.
FULL_GO_TEST_RE = re.compile(r"(?:^|\s)\./(?:internal/guard|cmd/flow|\.\.\.)/?(?=\s|$)")
FULL_GO_TEST_REASON = (
    "`go test` over a whole heavy package (./internal/guard, ./cmd/flow or ./...) takes minutes — name "
    "the tests the change touches with `-run '<TestName|OtherTest>'`. A full package run is the "
    "operator's to ask for, and runs from their own `!` prompt."
)


def full_go_test(command):
    for segment in re.split(r"[;&|\n]+", command):
        if (
            re.search(r"\bgo\s+test\b", segment)
            and FULL_GO_TEST_RE.search(segment)
            and not re.search(r"(?:^|\s)-run[\s=]", segment)
        ):
            return True
    return False


_git_cache = {}


def git(cwd, *args):
    """stdout of a git call in cwd, or None on any failure."""
    key = (cwd, args)
    if key not in _git_cache:
        try:
            r = subprocess.run(
                ["git", "-C", cwd, *args], capture_output=True, text=True, timeout=5
            )
            _git_cache[key] = r.stdout.strip() if r.returncode == 0 else None
        except (OSError, subprocess.SubprocessError):
            _git_cache[key] = None
    return _git_cache[key]


def existing_dir(path):
    """The nearest existing directory at or above path (a Write may create the file)."""
    p = os.path.abspath(path)
    while not os.path.isdir(p):
        parent = os.path.dirname(p)
        if parent == p:
            return None
        p = parent
    return p


def is_main_checkout(p):
    """True when p is the toplevel of a main checkout — not a subdirectory, not a linked worktree."""
    top = git(p, "rev-parse", "--show-toplevel")
    git_dir = git(p, "rev-parse", "--absolute-git-dir")
    common = git(p, "rev-parse", "--git-common-dir")
    if not top or not git_dir or not common:
        return False
    if not os.path.isabs(common):
        common = os.path.join(p, common)
    return os.path.realpath(top) == os.path.realpath(p) and os.path.realpath(git_dir) == os.path.realpath(common)


def local_clone(p):
    """True when p is the toplevel of a clone whose origin is a local non-bare checkout —
    a scratch copy, whose whole removal writes no main checkout (kan-1107 self-review, F5)."""
    if not os.path.isdir(p) or not is_main_checkout(p):
        return False
    origin = git(p, "remote", "get-url", "origin") or ""
    if origin.startswith("file://"):
        origin = origin[len("file://") :]
    return os.path.isabs(origin) and os.path.isdir(origin) and git(origin, "rev-parse", "--is-bare-repository") == "false"


def worktree_root(d):
    """True when d is a worktree root: `<main>-worktrees` beside a main checkout `<main>`, or the
    pre-migration `<main>/.worktrees` the hook no longer suggests."""
    name = os.path.basename(d)
    if name == ".worktrees":
        return True
    main = name[: -len("-worktrees")]
    return name.endswith("-worktrees") and bool(main) and is_main_checkout(os.path.join(os.path.dirname(d), main))


class LinkItself(str):
    """A path an `rm` or `mv` source names: those act on a symlink itself, not on its target,
    so a source that is a symlink is judged by the directory holding it."""


def protected(path):
    """(toplevel, branch) when path lies in a main checkout on its default branch, else None.
    path is judged after its symlinks resolve, as the write lands: `abspath` would collapse
    `<link>/..` lexically, and a link into a checkout's subdirectory would hide the checkout."""
    if isinstance(path, LinkItself) and os.path.islink(path):
        path = os.path.realpath(os.path.dirname(path))  # the link's own directory is what changes
    else:
        path = os.path.realpath(path)
    d = existing_dir(path)
    if d is None:
        return None
    if worktree_root(d) and os.path.abspath(path) != d:
        # the nearest existing ancestor is a worktree root, so a missing component
        # below it is a future worktree — never main-checkout content; a
        # not-yet-existing worktree is where work goes. The root itself — path == d —
        # stays protected like any other main-checkout directory.
        return None
    top = git(d, "rev-parse", "--show-toplevel")
    if not top:
        return None
    git_dir = git(d, "rev-parse", "--absolute-git-dir")
    common = git(d, "rev-parse", "--git-common-dir")
    if not git_dir or not common:
        return None
    if not os.path.isabs(common):
        common = os.path.join(d, common)
    if os.path.realpath(git_dir) != os.path.realpath(common):
        return None  # a linked worktree
    branch = git(d, "branch", "--show-current")
    if not branch:
        return None  # detached
    default = default_branch(d)
    if default is None and branch in FALLBACK_DEFAULT_BRANCHES:
        default = branch
    if branch != default:
        return None
    return top, branch


def resolve(path, cwd):
    """path made absolute against cwd, or None when the shell would expand it (`$VAR`, a
    backtick) or cwd itself is unknown — a path the hook cannot know is let through, never
    guessed at: `cd $W; ... >$LOG` once resolved `$LOG` under the literal `$W` and walked up into
    the main checkout."""
    if "$" in path or "`" in path:
        return None
    path = os.path.expanduser(path)
    if os.path.isabs(path):
        return path
    return os.path.join(cwd, path) if cwd else None


def tokenize(command):
    """Shell words with `;`, `&&`, `|` and redirects split out even where they touch a word —
    `cd <worktree>; git commit` must not make `<worktree>;` the directory."""
    try:
        lexer = shlex.shlex(command, posix=True, punctuation_chars=True)
        lexer.whitespace_split = True
        # A run of punctuation is one shlex token: split each paren off it, so `)>` and `;)`
        # still close their group.
        words = [p for w in lexer for p in (re.split(r"([()])", w) if re.fullmatch(r"[();&|<>]+", w) else [w]) if p]
    except ValueError:
        # Unbalanced quotes: raw words, a redirect glued to its target split off it, so `>f`
        # and `2>>f` read as shlex would have split them; a quoted word keeps its quote and
        # never matches.
        return [p for w in command.split() for p in _split_redirect(w)]
    # A backtick span, and a `$(...)` inside a quoted word, still run: each one's command is
    # scanned again as a command of its own, after the line's words.
    spans = [sp for w in words for sp in SUBST_SPAN_RE.findall(w) if sp.startswith("$(")]
    spans += [sp for sp in SUBST_SPAN_RE.findall(command.replace("\\`", "")) if sp.startswith("`")]
    # Each is wrapped in `( … )`: a `cd` inside a substitution ends at its close.
    for span in spans:
        words += [";", "(", *tokenize(span[1:-1] if span.startswith("`") else span[2:-1]), ")", ";"]
    return words


LITERAL_ABS = r"(/[^\s'\"$`\\;&|<>(){}*?\[\]~!#]+)"
DATA_WRITE_RE = re.compile(
    rf"[ \t]*(?:cat[ \t]+>>?[ \t]*{LITERAL_ABS}|tee(?:[ \t]+-a)?[ \t]+{LITERAL_ABS}(?:[ \t]+>[ \t]*/dev/null)?)"
    r"[ \t]+<<[ \t]*(?:'(\w+)'|\"(\w+)\"|(\w+))[ \t]*\r?"
)


def data_write_target(command):
    """ABS when the whole command is one plain heredoc data write — `cat > ABS <<DELIM`,
    `cat >> ABS <<DELIM` or `tee [-a] ABS [>/dev/null] <<DELIM` with ABS a literal absolute
    path, its body, a closing line equal to DELIM (a trailing `\\r` tolerated), then only blank
    lines — else None. Such a body is stdin data, so only ABS is judged; any other command is
    scanned whole, body included, so the hook is never weaker than a plain token scan. An
    unquoted delimiter's body runs `$(...)` and backticks, so one holding either is no data."""
    lines = command.split("\n")
    m = DATA_WRITE_RE.fullmatch(lines[0])
    if not m:
        return None
    target, delim = m.group(1) or m.group(2), m.group(3) or m.group(4) or m.group(5)
    for k, line in enumerate(lines[1:], 1):
        if line.rstrip("\r") == delim:
            body = "\n".join(lines[1:k])
            if m.group(5) and ("$(" in body or "`" in body):
                return None
            return target if all(not rest.strip() for rest in lines[k + 1 :]) else None
    return None


HEREDOC_OPEN_RE = re.compile(r"<<-?[ \t]*(?:'(\w+)'|\"(\w+)\"|(\w+))")
SEGMENT_SPLIT_RE = re.compile(r"&&|\|\||[;|]")
CAT_TARGET_RE = re.compile(r">>?[ \t]*([^\s;&|<>]+)")
SUBST_RE = re.compile(r"[<>]\(|\$\(|`")
COMMAND_CUT_RE = re.compile(r"&&|\|\||[\n;|&(){}]")
SHELL_KEYWORDS = {"do", "then", "else", "elif", "if", "while", "until", "!"}
# A `flow` opener's later words must be plain tokens and none a shell or interpreter name
# (`flow run csh <<EOF`), or its heredoc body stays scanned.
FLOW_PLAIN_WORD_RE = re.compile(r"^[A-Za-z0-9_.,=-]+$")
FLOW_SHELL_WORDS = {
    "sh", "bash", "zsh", "dash", "ksh", "csh", "tcsh", "mksh", "yash", "fish", "ash", "busybox",
    "su", "ssh", "doas", "sudo", "env", "script", "parallel", "xargs", "python", "python3", "perl",
    "ruby", "node",
}

RUNNER_WORDS = {
    "sh", "bash", "zsh", "dash", "ksh", "fish", "source", ".", "eval", "exec", "xargs", "env",
    "command", "builtin", "nohup", "timeout", "nice", "time", "sudo", "python", "python3",
    "perl", "ruby", "node", "osascript", "awk", "gawk", "mawk", "nawk", "make", "gmake", "watch",
    "crontab", "at", "batch",
}


def names_target(word, target):
    """True when a command word may name the heredoc's target: the same spelling, the same file
    name under another path (`"$D"/x.sh` for `$SIB/x.sh`), or a file name an expansion hides."""
    word = word.replace('"', "").replace("'", "")
    name = word.rsplit("/", 1)[-1]
    return word == target or name == target.rsplit("/", 1)[-1] or "$" in name


# Commands known not to execute a file they are given: the heredoc target named only as an
# argument of one of these is data. Anything else naming it may run it, so the body stays scanned.
READER_WORDS = {
    "cat", "tee", "head", "tail", "wc", "grep", "diff", "cmp", "ls", "test", "[", "stat", "cd",
    "echo", "printf",
}
# Shell keywords that open a command without being one: `do x`'s command is `x`.
LEADING_KEYWORDS = {"do", "then", "else", "elif", "!"}
GIT_READ_SUBCOMMANDS = {"diff", "add", "show", "log", "status"}
CHECK_SCRIPT_RE = re.compile(r"check-[\w.-]*\.(?:sh|py)")
# A guard run through the bootstrap: `flow-guard check-…`, bare or as `…/scripts/flow-guard`.
CHECK_GUARD_RE = re.compile(r"check-[\w.-]*")
# Interpreters whose program text is not shell: a body they run is scanned for git verbs only.
PROGRAM_RUNNERS = {"python", "python3", "node"}


def reads_only(words, target):
    """True when the command words are one known not to execute its file arguments: a
    pipeline guard `flow-guard check-*`, `check-*.sh`/`check-*.py` other than the target itself (a written check
    script run afterwards runs the body), a reader in READER_WORDS (`rg` is not one: its
    `--pre` runs a command), `git` with a read subcommand, or `spectre validate`."""
    name = words[0].rsplit("/", 1)[-1]
    if CHECK_SCRIPT_RE.fullmatch(name) and name != target.rsplit("/", 1)[-1]:
        return True
    # The bootstrap runs a compiled guard, never a file it is given; written as the target it runs that body.
    if name == "flow-guard" and name != target.rsplit("/", 1)[-1] and len(words) > 1:
        return bool(CHECK_GUARD_RE.fullmatch(words[1]))
    if name in READER_WORDS:
        return True
    if name == "git":
        return len(words) > 1 and words[1] in GIT_READ_SUBCOMMANDS
    if name == "spectre":
        return len(words) > 1 and words[1] == "validate"
    return False


def outside_single_quotes(text):
    """text with every single-quoted span — `'…'`, and `$'…'` with its backslash escapes —
    blanked, since bash reads a backtick or `$(` inside one as a plain character; a `'` inside
    `"…"` or behind a backslash is literal. None when a quote never closes, or when a
    single-quoted span crosses a newline — an apostrophe in a `#` comment or another heredoc's
    body opens no quote in bash — so the caller reads the text as before."""
    out, i, quote = [], 0, None
    while i < len(text):
        c = text[i]
        if quote in ("'", "$'"):
            if quote == "$'" and c == "\\":
                i += 2
                continue
            if c == "\n":
                return None
            if c == "'":
                quote = None
            i += 1
            continue
        if c == "\\":
            out.append(text[i:i + 2])
            i += 2
            continue
        if quote == '"':
            quote = None if c == '"' else quote
        elif c == '"':
            quote = '"'
        elif c == "'":
            quote = "$'" if text[i - 1:i] == "$" else "'"
            out.append(" ")
            i += 1
            continue
        out.append(c)
        i += 1
    return None if quote else "".join(out)


def command_segments(text):
    """(words, assigned) per command of text — quotes stripped, a leading `do`/`then`/`!`
    dropped, `assigned` when NAME=value words led it — or None when text holds a substitution
    this cannot read: a process substitution, a backtick, a nested `$(`. A plain `$(...)` is a
    command of its own. A literal NAME=value expands later words, so `$D/r.md` reads as the file
    D names; a name no literal assigns stays an expansion `names_target` reads as hiding any."""
    spans = re.findall(r"\$\(([^()]*)\)", text)
    flat = re.sub(r"\$\([^()]*\)", " ", text)
    bare = outside_single_quotes(flat)
    if SUBST_RE.search(flat if bare is None else bare):
        return None
    env, out = {}, []
    for seg in COMMAND_CUT_RE.split(flat + "\n" + "\n".join(spans)):
        words = [w.strip("'\"") for w in seg.split()]
        while words and words[0] in LEADING_KEYWORDS:
            words.pop(0)
        assigned = False
        while words and ASSIGNMENT_RE.match(words[0]):
            name, _, value = words.pop(0).partition("=")
            value = expand_vars(value, env)
            if value and "$" not in value:
                env[name] = value
            else:
                env.pop(name, None)
            assigned = True
        out.append(([expand_vars(w, env) for w in words], assigned))
    return out


def literal_export(args, target):
    """True when every export argument is `NAME=value` whose value can run nothing: no `$` or
    backtick, and not naming target (`names_target`) — with target None, holding no `/`."""
    if not args:
        return False
    for arg in args:
        if not ASSIGNMENT_RE.match(arg):
            return False
        value = arg.split("=", 1)[1]
        if "$" in value or "`" in value:
            return False
        if names_target(value, target) if target else "/" in value:
            return False
    return True


def may_run_code(text, target):
    """True when text — the command outside one heredoc body — could run that body or the file
    target it writes. Fails closed: a process or command substitution (`>(sh)`, `bash <(`,
    `eval "$(`), a word that runs code (RUNNER_WORDS, matched by file name so `/bin/bash`
    counts as `bash`) anywhere, or any command whose words may name target (`names_target` —
    the same spelling, the same file name, a bare `x.sh`, an expansion) unless that command is
    one `reads_only` allows. A command naming another file (`lint.sh /x/o.md`,
    `./gradlew`) leaves the body as data; `php $T`, `setsid $T`, `find . -exec`, `ssh h $T`
    keep it scanned, as does any command behind a leading `NAME=value` or run under
    `env`/`export` — save an `export` whose every argument is `NAME=value` with a literal value:
    no `$` or backtick, and one that cannot name target (with target None, no `/`). With target None no command names it, so substitutions, runner words, env
    assignments and a command word holding `/` or `$` count. Over-matching only keeps the body scanned."""
    segments = command_segments(text)
    if segments is None:
        return True
    for words, assigned in segments:
        if any(w.rsplit("/", 1)[-1] in RUNNER_WORDS for w in words):
            return True
        # An env assignment can make any command run a file (GIT_EXTERNAL_DIFF, PAGER,
        # LD_PRELOAD, …): a command carrying one, or run under `env`/`export`, is never a reader.
        if words and not assigned and words[0] == "export" and literal_export(words[1:], target):
            continue
        if words and (assigned or words[0].rsplit("/", 1)[-1] in {"env", "export"}):
            return True
        # With no target, a command word holding `/` or `$` may run what the opener wrote or
        # read (`/tmp/zz.sh`, `$l`).
        lead = words[:]
        while lead and lead[0] in SHELL_KEYWORDS:
            lead.pop(0)
        if target is None and lead and ("/" in lead[0] or "$" in lead[0]):
            return True
        if target and words and any(names_target(w, target) for w in words) and not reads_only(words, target):
            return True
    return False


def runs_as_program(text, target):
    """True when text — the command outside one heredoc body — runs target only as a non-shell
    program: no substitution, and every command naming target is a reader (`reads_only`), a
    PROGRAM_RUNNERS interpreter given it (`python3 x.py`, `node x.js`) or `npx playwright test`,
    with no other runner word anywhere. Such a body is that language's text, never shell."""
    segments = command_segments(text)
    if segments is None:
        return False
    for words, assigned in segments:
        if not words:
            continue
        if words[0].rsplit("/", 1)[-1] in PROGRAM_RUNNERS or words[:3] == ["npx", "playwright", "test"]:
            if any(w.rsplit("/", 1)[-1] in RUNNER_WORDS for w in words[1:]):
                return False
            continue
        if any(w.rsplit("/", 1)[-1] in RUNNER_WORDS for w in words):
            return False
        if assigned or words[0].rsplit("/", 1)[-1] in {"env", "export"}:
            return False
        if any(names_target(w, target) for w in words) and not reads_only(words, target):
            return False
    return True


PROGRAM_GIT_RE = re.compile(r"(?:^|\s)(git\s.*)$")


def program_git_lines(body):
    """The git invocations a non-shell program body may make, one scan line each: quotes and
    brackets become spaces, so `subprocess.run(["git", "-C", d, "commit"])` reads `git -C d
    commit`, and every other word — a `d>0` comparison, an `rm` in a string — is dropped."""
    out = []
    for line in body.split("\n"):
        m = PROGRAM_GIT_RE.search(re.sub(r"[\"'()\[\],<>|;&{}]", " ", line))
        if m:
            out.append(m.group(1))
    return out


# A `$(...)` without a nested paren, or a backtick pair: the spans of an unquoted heredoc body
# that bash runs.
SUBST_SPAN_RE = re.compile(r"\$\([^()]*\)|`[^`]*`")

# The operators that write a file: shlex (punctuation_chars) yields each as a word of its own, so
# a word that merely starts with `>` came out of quotes and is data (`'>Save<'`).
REDIRECT_OPS = (">", ">>", ">|", "&>", "&>>")
# A redirect glued to its target in an unbalanced-quote fallback word: `>f`, `2>>f`, `&>f`; `>&1`
# duplicates a descriptor and writes no file.
ATTACHED_REDIRECT_RE = re.compile(r"^\d*(&>>|&>|>>|>\||>)([^&].*)$")


def subst_spans(body):
    """The `$(...)` and backtick spans an unquoted heredoc body runs, a backtick span rewritten as
    `$(...)`, or None when a span cannot be closed. body has its escaped backticks removed."""
    spans = SUBST_SPAN_RE.findall(body)
    if (
        body.count("$(") != sum(sp.startswith("$(") for sp in spans)
        or body.count("`") % 2
        # `$(case a in a) rm …;; esac)`: the span closes early and drops the rest.
        or body.count(")") != sum(sp.count(")") for sp in spans if sp.startswith("$("))
    ):
        return None
    return ["$(" + sp[1:-1] + ")" if sp.startswith("`") else sp for sp in spans]


# A heredoc opener's segment cut into words so `x=$(bash` yields `bash`.
OPENER_WORD_SPLIT_RE = re.compile(r"[\s=$()`'\"]+")


def _split_redirect(word):
    m = ATTACHED_REDIRECT_RE.match(word)
    return [m.group(1), m.group(2)] if m else [word]


def drop_data_bodies(lines, target):
    """lines with the body of every heredoc whose text the shell runs nothing of removed — a
    quoted delimiter's, or an unquoted one holding no `$(` or backtick — keeping its opener and
    delimiter lines, so a `bash <<'E'` opener still reads as a runner. strip_data_heredocs judges
    one body's rest through this: another heredoc's prose (a lone `.`, an unpaired backtick) is
    no command of the rest (kan-1041 self-review: two report heredocs in one call were denied).
    A body naming `target` is kept: it may be a script, or a list, that runs the target."""
    out = []
    k = 0
    while k < len(lines):
        line = lines[k]
        out.append(line)
        k += 1
        m = HEREDOC_OPEN_RE.search(line)
        if not m:
            continue
        delim = m.group(1) or m.group(2) or m.group(3)
        tabs = line[m.start() : m.start() + 3] == "<<-"
        end = next(
            (j for j in range(k, len(lines)) if (lines[j].lstrip("\t") if tabs else lines[j]).rstrip("\r") == delim),
            None,
        )
        body = "\n".join(lines[k:end])
        if end is None or (target and target in body) or (m.group(3) and any(c in body for c in ("$(", "`"))):
            continue
        out.append(lines[end])
        k = end + 1
    return out


def strip_data_heredocs(command):
    """command with the body of a `cat`/`tee` heredoc to a named file removed — that body is
    data for a writer whose target the opening line already names, so a report line such as
    `a > the` must not read as a redirect (kan-929: subagents' report writes
    `D=...; cat > $D/r.md <<'EOF'` were denied over their own prose). A body stays scanned when
    the writer has no target, its output is piped on, the target is a device or a process
    substitution (`/dev/stdout`, `>(sh)`), the rest of the command may run code (`may_run_code` —
    an enclosing `bash <(` or `eval "$(`, a later `sh "$D"/x.sh` or `"$D"/x.sh`: written, then
    run; the target named again only as an argument is data), the opener is not a plain `cat`/`tee`
    (`python3 - <<EOF`, `x=$(cat <<EOF`). An unquoted-delimiter body that runs `$(...)` or
    backticks keeps only those spans scanned, a backtick span as `$(...)` — bash runs them and
    nothing else of the body — or the whole body when a span cannot be closed. Any other opener's
    body is data only when the opener is `flow` (an allowlist), its line is not continued, it
    redirects no output, and the rest of the command runs no code (`may_run_code`) — kan-1025: `flow record summary -file - <<'EOF'` bodies were denied over
    their prose."""
    lines = command.split("\n")
    out = []
    k = 0
    while k < len(lines):
        line = lines[k]
        out.append(line)
        k += 1
        m = HEREDOC_OPEN_RE.search(line)
        if not m or "|" in line[m.end() :]:
            continue
        seg = SEGMENT_SPLIT_RE.split(line[: m.start()])[-1] + " " + line[m.end() :]
        words = seg.split()
        while words and ASSIGNMENT_RE.match(words[0]):
            words.pop(0)
        delim = m.group(1) or m.group(2) or m.group(3)
        tabs = line[m.start() : m.start() + 3] == "<<-"
        end = next(
            (j for j in range(k, len(lines)) if (lines[j].lstrip("\t") if tabs else lines[j]).rstrip("\r") == delim),
            None,
        )
        if words and words[0].rsplit("/", 1)[-1] in PROGRAM_RUNNERS and end is not None:
            # A python/node program on stdin: its text is not shell (kan-1026: `d>0`).
            body = "\n".join(lines[k:end])
            if m.group(3) and any(c in body for c in ("$(", "`")):
                continue
            out.extend(program_git_lines(body))
            k = end + 1
            continue
        pre = line[: m.start()]
        if (
            end is not None
            # A `<<` in quotes, a `${…}` or a comment opens no heredoc: its "body" is shell.
            and pre.count("'") % 2 == 0
            and pre.count('"') % 2 == 0
            and pre.count("${") <= pre.count("}")
            and "#" not in pre
            and words
            # An allowlist, not a denylist: any other opener may be a stdin interpreter
            # (`csh`, `ssh h`, `sqlite3`), so only `flow` reads its body as data.
            and words[0].lstrip("\\") == "flow"
            and all(FLOW_PLAIN_WORD_RE.match(w) and w not in FLOW_SHELL_WORDS for w in words[1:])
            and not line.rstrip().endswith("\\")  # a continued opener line hides the rest
            and ">" not in seg  # output written to a file may be run later
            and not any(w.rsplit("/", 1)[-1] in RUNNER_WORDS for w in OPENER_WORD_SPLIT_RE.split(seg))
            and not may_run_code("\n".join(drop_data_bodies(lines[:k], None) + drop_data_bodies(lines[end + 1 :], None)), None)
        ):
            body = "\n".join(lines[k:end]).replace("\\`", "")
            if m.group(3) and any(c in body for c in ("$(", "`")):
                spans = subst_spans(body)
                if spans is None:
                    continue
                out.extend(spans)
            k = end + 1
            continue
        if not words or words[0] not in ("cat", "tee"):
            continue
        if words[0] == "cat":
            t = CAT_TARGET_RE.search(seg)
            target = t.group(1) if t else None
        else:
            target = next((w for w in words[1:] if not w.startswith(("-", ">"))), None)
        if not target or end is None:
            continue
        if target.startswith("/dev/") or "(" in target:
            continue
        rest = "\n".join(drop_data_bodies(lines[:k], target) + drop_data_bodies(lines[end + 1 :], target))
        if may_run_code(rest, target):
            if runs_as_program(rest, target) and not (m.group(3) and any(c in "\n".join(lines[k:end]) for c in ("$(", "`"))):
                out.extend(program_git_lines("\n".join(lines[k:end])))
                k = end + 1
            continue
        # An escaped backtick is a literal in an unquoted body; dropped, it cannot pair with a real one.
        body = "\n".join(lines[k:end]).replace("\\`", "")
        if m.group(3) and any(c in body for c in ("$(", "`")):
            # Only the substitutions of an unquoted body run; its other text is data, so the
            # spans alone are kept scanned. A span this cannot close keeps the whole body.
            spans = subst_spans(body)
            if spans is None:
                continue
            out.extend(spans)
        k = end + 1
    return "\n".join(out)


def logical_lines(command):
    """The command split into logical lines: a newline outside quotes cuts, a
    backslash-newline outside quotes joins, and a quoted newline stays inside
    its line — bash's own line discipline, close enough for a token scan."""
    lines = []
    buf = []
    quote = None  # None, "'" or '"' while inside that quote
    i = 0
    n = len(command)
    while i < n:
        ch = command[i]
        if quote is None and ch == "\\" and i + 1 < n and command[i + 1] == "\n":
            i += 2  # continuation: the two physical lines are one command
            continue
        if quote is None and ch == "\\" and i + 1 < n:
            buf.append(ch)
            buf.append(command[i + 1])
            i += 2
            continue
        if quote == '"' and ch == "\\" and i + 1 < n and command[i + 1] in ('"', "\\", "$", "`"):
            buf.append(ch)
            buf.append(command[i + 1])
            i += 2
            continue
        if quote is None and ch in ("'", '"'):
            quote = ch
        elif quote == ch:
            quote = None
        elif quote is None and ch == "\n":
            lines.append("".join(buf))
            buf = []
            i += 1
            continue
        buf.append(ch)
        i += 1
    lines.append("".join(buf))
    return lines


VAR_RE = re.compile(r"\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)")


def expand_vars(token, env):
    """$NAME and ${NAME} replaced where NAME is set in env; the name match is
    greedy, so `$Dy` looks up `Dy`, never `D` — an unset name is left for
    resolve() to let through, never guessed at."""
    if "$" not in token or not env:
        return token

    def sub(m):
        name = m.group(1) or m.group(2)
        return env.get(name, m.group(0))

    return VAR_RE.sub(sub, token)


def default_branch(cwd):
    """The repository's default branch: origin/HEAD's target, else None."""
    head = git(cwd, "symbolic-ref", "-q", "--short", "refs/remotes/origin/HEAD")
    return head.split("/", 1)[1] if head and "/" in head else None


SEPARATORS = ("&&", "||", ";", "|")
def in_change_state(target):
    """True when target's worktree is a key of `worktrees` in the state record of the change named
    by its directory (`<project>-worktrees/<change>`), read with `flow state get`. A change spans
    sibling repos' worktrees (`<other>-worktrees/<change>`) but its record lives under one project,
    so each sibling checkout is asked in turn. Fails closed."""
    top = git(target, "rev-parse", "--show-toplevel")
    if not top:
        return False
    top, name = os.path.realpath(top), os.path.basename(top)
    for c in [top] + sorted(glob.glob(os.path.join(os.path.dirname(os.path.dirname(top)), "*-worktrees", name))):
        try:
            out = subprocess.run(["flow", "state", "get", "-C", c, name],
                                 capture_output=True, text=True, timeout=5).stdout
            keys = json.loads(out[out.index("{"):]).get("worktrees") or {}
        except (OSError, subprocess.SubprocessError, ValueError, AttributeError):
            continue
        if top in {os.path.realpath(k) for k in keys}:
            return True
    return False


def discards_unstaged(verb, args, target):
    """The deny reason when `git checkout … -- <path>` or `git restore <path>` in a linked worktree
    names a path carrying unstaged edits, else None. Either restore reverts the file to its index
    state, so an uncommitted fix dies with the mutation-proof flip it was meant to undo (kan-1006's
    self-review). `git restore --staged` touches only the index and passes.

    ponytail: a `git checkout <path>` without `--` is not judged — telling a path from a branch
    there needs the ref namespace; add it when that spelling shows up in a reflog."""
    if target is None:
        return None
    if verb == "checkout":
        if "--" not in args:
            return None
        paths = args[args.index("--") + 1 :]
        tree_ish = any(not a.startswith("-") for a in args[: args.index("--")])
    elif verb == "restore":
        if ("--staged" in args or "-S" in args) and "--worktree" not in args and "-W" not in args:
            return None
        tree_ish = any(a in ("-s", "--source") or a.startswith("--source=") for a in args)
        if "--" in args:
            paths = args[args.index("--") + 1 :]
        else:
            paths, skip = [], False
            for a in args:
                if skip:
                    skip = False
                elif a in ("-s", "--source"):
                    skip = True
                elif not a.startswith("-"):
                    paths.append(a)
    else:
        return None
    if not paths:
        return None
    git_dir = git(target, "rev-parse", "--absolute-git-dir")
    common = git(target, "rev-parse", "--git-common-dir")
    if not git_dir or not common:
        return None
    if not os.path.isabs(common):
        common = os.path.join(target, common)
    if os.path.realpath(git_dir) == os.path.realpath(common):
        return None  # not a linked worktree
    if in_change_state(target):
        return None  # a change's own worktree: its fixer restores stray baselines
    # a tree-ish or --source overwrites the index too, so staged edits are lost as well
    dirty = git(target, "diff", "--name-only", *(["HEAD"] if tree_ish else []), "--", *paths)
    if not dirty:
        return None
    files = ", ".join(dirty.splitlines())
    return (
        f"Blocked: this `git {verb}` would discard the uncommitted edits on {files} — it reverts the "
        "file to its index state, and an uncommitted fix on it is lost with the flip a mutation "
        "proof made (**Restoring a path with unstaged edits**, skills/flow-contracts/git-boundaries.md). "
        "Restore a flipped file with `flow-guard break-and-prove`, which restores "
        "from a pre-mutation byte snapshot; edits that must go are committed or stashed first."
    )


def git_args(tokens, start):
    """A git verb's own arguments: the words from `start` up to the first separator, `)` or
    redirect — `2>&1` and `>/dev/null` are the shell's, never a remote, refspec or pathspec."""
    out = []
    for t in tokens[start:]:
        if t in SEPARATORS or t == ")":
            break
        if t.startswith(("<", ">", "&>")):
            if out and out[-1].isdigit():
                out.pop()  # the fd of `2>&1`
            break
        out.append(t)
    return out


def bash_hits(command, cwd, landings=None):
    """Paths a Bash command writes into or git-mutates, resolved against cwd and `cd`.

    The command is scanned one logical line at a time — a newline ends an
    argument scan the way `;` does, so a multi-line script's `cp` cannot
    swallow a later line's arguments — while `cd` state carries across lines,
    as it does in bash. Assignments the command text itself makes are expanded
    before path resolution, positionally: a token sees the assignments that
    came before it, in this line and earlier lines, never the ones after —
    and an assignment's value is kept literal, unexpanded."""
    target = data_write_target(command)
    if target:
        return [target]
    hits = []
    cur = cwd
    saved = []  # the `cur` each open `(` will restore at its `)`
    env = {}
    for line in logical_lines(strip_data_heredocs(command)):
        raw_tokens = tokenize(line)
        tokens = []
        for k, raw in enumerate(raw_tokens):
            if ASSIGNMENT_RE.match(raw):
                # A run of assignments followed by a command word is that command's
                # prefix: bash does not use it when expanding the command's own words.
                rest = [t for t in raw_tokens[k + 1 :] if not ASSIGNMENT_RE.match(t)]
                if not rest or rest[0] in SEPARATORS:
                    name, _, value = raw.partition("=")
                    env[name] = value
            tokens.append(expand_vars(raw, env))
        i = 0
        n = len(tokens)
        while i < n:
            tok = tokens[i]
            if tok == "(":
                saved.append(cur)
            elif tok == ")" and saved:
                cur = saved.pop()  # a `)` with nothing open is a `case` pattern's
            if tok == "cd" and i + 1 < n:
                cur = resolve(tokens[i + 1], cur)
                i += 2
                continue
            expanded = "$" in tok or "`" in tok
            if tok == "git" or (expanded and i + 1 < n and tokens[i + 1] == "push"):
                # `$GIT push` or `git $OPTS push`: the repository cannot be read.
                target = None if expanded else cur
                j = i + 1
                while j < n and (tokens[j].startswith("-") or "$" in tokens[j] or "`" in tokens[j]):
                    if not tokens[j].startswith("-"):
                        target = None
                        j += 1
                        continue
                    if tokens[j] == "-C" and j + 1 < n:
                        target = resolve(tokens[j + 1], cur)
                        j += 2
                        continue
                    if tokens[j] == "-c" and j + 1 < n:
                        j += 2
                        continue
                    j += 1
                if j < n and tokens[j] in DENIED_GIT_VERBS:
                    hits.append(target)
                if j < n and tokens[j] in ("checkout", "restore") and landings is not None:
                    rargs = git_args(tokens, j + 1)
                    discard = discards_unstaged(tokens[j], rargs, target)
                    if discard:
                        landings.append(discard)
                i = j + 1
                continue
            if tok == "sed" and any(t.startswith("-i") for t in tokens[i + 1 : i + 4]):
                for t in tokens[i + 1 :]:
                    if t in ("&&", "||", ";", "|"):
                        break
                    # an empty word is BSD sed's `-i ''` suffix, never a path
                    path = None if not t or t.startswith("-") else resolve(t, cur)
                    if path and os.path.exists(path):
                        hits.append(path)
                i += 1
                continue
            if tok == "tee":
                for t in tokens[i + 1 :]:
                    if t in ("&&", "||", ";", "|"):
                        break
                    if not t.startswith("-"):
                        hits.append(resolve(t, cur))
                i += 1
                continue
            # `>& f` writes stdout and stderr to f; `>& 2` and `>& -` only duplicate or close
            nxt = tokens[i + 1] if i + 1 < n else None
            if nxt is not None and (tok in REDIRECT_OPS or (tok == ">&" and not nxt.isdigit() and nxt != "-")):
                hits.append(resolve(tokens[i + 1], cur))
                i += 2
                continue
            if tok in ("cp", "mv", "rm"):
                args = []
                for t in tokens[i + 1 :]:
                    if t in ("&&", "||", ";", "|", ")"):
                        break
                    if not t.startswith("-"):
                        args.append(t)
                if tok in ("rm", "mv"):
                    # mv removes its sources too; neither follows a symlink it is given
                    hits.extend(
                        LinkItself(r)
                        for r in (resolve(a, cur) for a in args)
                        if r and not (tok == "rm" and local_clone(r))
                    )
                elif args:
                    hits.append(resolve(args[-1], cur))
                i += 1
                continue
            if tok == "trash":
                # a move to the Trash is recoverable, so only a path git tracks is a write
                for t in tokens[i + 1 :]:
                    if t in SEPARATORS or t == ")":
                        break
                    path = None if t.startswith("-") else resolve(t, cur)
                    if path and git(os.path.dirname(path) or "/", "ls-files", "--error-unmatch", os.path.basename(path)) is not None:
                        hits.append(path)
                i += 1
                continue
            i += 1
    return [h for h in hits if h]


STASH_RE = re.compile(r"\bgit((?:\s+-[Cc]\s+\S+)*)\s+stash\s+(pop|apply|drop|branch)\b([^;&|\n)]*)")
STASH_REF_RE = re.compile(r"stash@\{(\d+)\}")
STASH_LIST_RE = re.compile(r"^stash@\{(\d+)\}: (?:WIP on|On) ([^:]+):")
STASH_BLOCKED = (
    " Never use git stash pop/apply/drop/branch: the stash list is shared by every worktree of the "
    "repository, so it can hold another change's work. If a worktree looks wrong, stop and report "
    "BLOCKED with the exact `git status` output."
)


def stash_denial(command, cwd):
    """The reason a `git stash pop|apply|drop|branch` in command acts on a stash not made on the
    current branch (or names none), else None."""
    for m in STASH_RE.finditer(command):
        verb, rest = m.group(2), m.group(3).split()
        target = cwd
        flags = m.group(1).split()
        for k, f in enumerate(flags):
            if f == "-C" and k + 1 < len(flags):
                target = resolve(flags[k + 1], cwd)
        ref = next((STASH_REF_RE.fullmatch(a) for a in rest if STASH_REF_RE.fullmatch(a)), None)
        if ref is None and verb != "branch":
            ref = next((re.fullmatch(r"(\d+)", a) for a in rest if re.fullmatch(r"\d+", a)), None)
        if ref is None:
            return f"Blocked: `git stash {verb}` names no stash entry, so it acts on stash@{{0}}, which may belong to another change." + STASH_BLOCKED
        listing = git(target, "stash", "list") or ""
        current = git(target, "rev-parse", "--abbrev-ref", "HEAD")
        made_on = {int(x.group(1)): x.group(2) for x in map(STASH_LIST_RE.match, listing.splitlines()) if x}
        branch = made_on.get(int(ref.group(1)))
        if branch is None or branch != current:
            return (
                f"Blocked: `git stash {verb} stash@{{{ref.group(1)}}}` — that entry "
                + (f"was made on `{branch}`, not on `{current}`." if branch else "is not in the stash list.")
                + STASH_BLOCKED
            )
    return None


def deny_reason(top, branch):
    name = os.path.basename(top)
    return (
        f"Blocked: {top} is the main checkout of {name}, and it is on `{branch}`, the default "
        "branch. Agents never edit, stage, commit, stash or reset there — it is the landing "
        "target every pipeline run pushes to, and anything left behind becomes residue no later "
        "session owns.\n\n"
        "Work in a worktree on a branch of its own instead:\n\n"
        f"  git -C {top} worktree add {top}-worktrees/<change> -b <change> origin/{branch}\n\n"
        "then edit, commit and push there, and land with a push of `<change>:"
        f"{branch}` or a pull request. To bring this checkout up to date afterwards, "
        f"`git -C {top} pull --ff-only` is allowed. Reads are always allowed. If the user "
        "themselves must act here, hand them the exact command to run with the `!` prefix."
    )


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0
    if not isinstance(payload, dict):
        return 0
    tool = payload.get("tool_name")
    tool_input = payload.get("tool_input")
    if not isinstance(tool_input, dict):
        return 0
    cwd = payload.get("cwd") or os.getcwd()

    candidates = []
    if tool in EDIT_TOOLS:
        path = tool_input.get("file_path") or tool_input.get("notebook_path")
        if isinstance(path, str) and path:
            candidates.append(resolve(path, cwd))
    elif tool == "Bash":
        command = tool_input.get("command")
        if isinstance(command, str) and command:
            landings = []
            if BARE_CAT_RE.search(command):
                landings.append(BARE_CAT_REASON)
            if full_go_test(command):
                landings.append(FULL_GO_TEST_REASON)
            stash = stash_denial(command, cwd)
            if stash:
                landings.append(stash)
            candidates.extend(bash_hits(command, cwd, landings))
            if landings:
                json.dump(
                    {
                        "hookSpecificOutput": {
                            "hookEventName": "PreToolUse",
                            "permissionDecision": "deny",
                            "permissionDecisionReason": landings[0],
                        }
                    },
                    sys.stdout,
                )
                return 0
    else:
        return 0

    for path in candidates:
        hit = protected(path) if path else None
        if hit is None:
            continue
        reason = deny_reason(*hit)
        if tool == "Bash":
            reason += f"\n\nThis command reads as writing {path}."
            if HEREDOC_OPEN_RE.search(command):
                reason += (
                    " A heredoc body the hook could not read as data is scanned as shell — an "
                    "`export`, a `NAME=value` prefix or a code-running word elsewhere in the command "
                    "keeps it scanned. Write the file with the Write tool, or run the heredoc as a "
                    "command of its own."
                )
        json.dump(
            {
                "hookSpecificOutput": {
                    "hookEventName": "PreToolUse",
                    "permissionDecision": "deny",
                    "permissionDecisionReason": reason,
                }
            },
            sys.stdout,
        )
        return 0
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # fail open, never block work on a hook fault
        sys.exit(0)
