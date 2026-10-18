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
    redirect, `cp` destinations, or `mv`/`rm` sources and destinations. A heredoc body is
    scanned like any other line, except where the whole command is one plain data write —
    `cat > ABS <<EOF`, `cat >> ABS <<EOF` or `tee [-a] ABS <<EOF` to a literal absolute path,
    its body and closing line, nothing after — where only ABS is judged — and except the body of
    any `cat`/`tee` heredoc inside a longer command, which is that writer's data, never commands.

"Main checkout" is the worktree whose git dir IS the common dir — a linked worktree's `--git-dir`
is `<common>/worktrees/<name>`, so `<repo>-worktrees/<change>` beside the main checkout passes,
as does a pre-migration `.worktrees/<change>` inside it. A path into a worktree that does not exist
yet passes too, when its nearest existing ancestor is either root. The deny reason suggests the
sibling `<dirname top>/<basename top>-worktrees/<change>` only. "Default branch" is what `origin/HEAD` points at; when the remote head is unknown, a checkout
of main, master or develop counts. A main checkout on a feature branch is not protected — leaving
the default branch is how work gets a branch of its own.

It also denies a landing push that would leave its change open: `git push <remote> <src>:<default>`
where `<src>`'s tree still holds `spectre/changes/<branch>/` — `<branch>` being `<src>`, or the
current branch for `HEAD`, with a `/flow` branch's `spectre/` prefix dropped. A change lands
archived (`spectre archive <change>`, then commit), never as an open folder on the default branch
that no later run owns. `/flow`'s run 1 archives before its push, so only a run that skipped it or
an ad-hoc landing meets this. A landing push of a `pipeline-fix-*` or `self-review-*` branch is
denied too unless scripts/check-loop-landing.sh, in the agents checkout this hook is installed
from, admits it — flowd holding the loop's completed plan, fix and clean last review. That one
check fails closed: a checker that cannot answer or cannot run denies the push.

It also denies a push of a `/flow` change branch (`spectre/<name>`) whose unpushed commits —
`origin/spectre/<name>..HEAD` — hold one outside every task with neither a `Task-Id:` nor a
`Flow-Key:` trailer, as scripts/check-untasked-commits.sh judges them (**Commits outside every
task**, skills/flow-contracts/git-boundaries.md; kan-941 pushed four such commits). It checks only a
push of HEAD's commit to a branch origin already holds, and fails open: a checker that cannot
answer, cannot run or times out lets the push through, and run 1's gate still lists the commit.

Both landing checks read the push from the command text, so a push the hook cannot read is denied
rather than let through: one whose repository (`-C <path>`, a `cd`), remote or refspec goes through
a shell expansion the hook cannot resolve — `$W`, `"${B}:main"`, `$(git branch --show-current)`, a
backtick — and whose destination is the default branch or cannot be told (0b055e56 landed a
`pipeline-fix-*` branch with no flowd records through `git -C $W push origin $B:main`). The reason
asks for literal names. A push whose destination is a literal other branch (`git -C $W push origin
spectre/<name>`) passes; an assignment the command itself makes (`B=x; git push origin $B:main`) is
expanded first, so only names from outside the command count. Scripts are untouched — the hook only
sees the typed Bash command, never what a script it calls runs.

Allowed everywhere: reads, `git pull --ff-only`, `git fetch`, `git checkout -b`, `git switch -c`,
`git worktree add`, `git push`, and the landing scripts (`land-self-review-report.sh`,
`refresh-main-checkout.sh`) — they are not `git` verbs in the command text, and each owns its own
main-checkout rules.

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

import json
import os
import re
import shlex
import subprocess
import sys
import time

EDIT_TOOLS = {"Edit", "Write", "MultiEdit", "NotebookEdit"}

ASSIGNMENT_RE = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*=")

# git verbs that change the index, the worktree, a ref or the stash list.
DENIED_GIT_VERBS = {
    "add", "am", "apply", "cherry-pick", "clean", "commit", "merge", "mv", "rebase",
    "reset", "restore", "revert", "rm", "stash", "update-index", "update-ref",
}

FALLBACK_DEFAULT_BRANCHES = {"main", "master", "develop"}

# A fix loop's branch, or a change's own branch a self-review loop may have run on
# (check-loop-landing.sh answers LOOP-LANDING-NA for one with no loop rows).
LOOP_BRANCH_RE = re.compile(r"^(pipeline-fix-.+-[0-9]+|self-review-.+|spectre/.+)$")

# The whole budget for check-loop-landing.sh calls in one command (every push it makes), well under the
# 300s setup.sh registers: a hook the harness times out fails open.
LOOP_CHECK_TIMEOUT = 240
LOOP_DEADLINE = None

# The untasked-commits checker and its budget; overridable for the hook's own harness only.
UNTASKED_CHECK = os.environ.get("UNTASKED_COMMITS_CHECK") or os.path.join(
    os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "scripts", "check-untasked-commits.sh"
)
UNTASKED_CHECK_TIMEOUT = 60

# The loop-landing checker; overridable for the hook's own harness only.
LOOP_CHECK = os.environ.get("LOOP_LANDING_CHECK") or os.path.join(
    os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "scripts", "check-loop-landing.sh"
)

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
        # not-yet-existing worktree is where work goes. The root convention is owned
        # by scripts/check-worktree-location.sh; this rule mirrors it for path
        # resolution, and the root itself — path == d — stays protected like any
        # other main-checkout directory.
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
        return list(lexer)
    except ValueError:
        return command.split()


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
RUNNER_WORDS = {
    "sh", "bash", "zsh", "dash", "ksh", "fish", "source", ".", "eval", "exec", "xargs", "env",
    "command", "builtin", "nohup", "timeout", "nice", "time", "sudo", "python", "python3",
    "perl", "ruby", "node", "osascript",
}


def may_run_code(text):
    """True when text — the command outside one heredoc body — could run that body or a file
    it writes: a process or command substitution (`>(sh)`, `bash <(`, `eval "$(`), a word that
    runs code (`sh`, `eval`, `xargs`, ...) anywhere, or a command named by a path or an
    expansion (`"$D"/x.sh`). Over-matching only keeps the body scanned."""
    if SUBST_RE.search(text):
        return True
    for seg in COMMAND_CUT_RE.split(text):
        words = [w.strip("'\"") for w in seg.split()]
        if any(w in RUNNER_WORDS for w in words):
            return True
        while words and ASSIGNMENT_RE.match(words[0]):
            words.pop(0)
        if words and ("/" in words[0] or "$" in words[0] or words[0].startswith("~")):
            return True
    return False


def strip_data_heredocs(command):
    """command with the body of a `cat`/`tee` heredoc to a named file removed — that body is
    data for a writer whose target the opening line already names, so a report line such as
    `a > the` must not read as a redirect (kan-929: subagents' report writes
    `D=...; cat > $D/r.md <<'EOF'` were denied over their own prose). A body stays scanned when
    the writer has no target, its output is piped on, the target is named again after the
    body (written, then run), the target is a device or a process substitution (`/dev/stdout`,
    `>(sh)`), the rest of the command may run code (`may_run_code` — an enclosing `bash <(` or
    `eval "$(`, a later `sh "$D"/x.sh`), the opener is not a plain `cat`/`tee`
    (`python3 - <<EOF`, `x=$(cat <<EOF`), or an unquoted-delimiter body runs `$(...)` or
    backticks."""
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
        if not words or words[0] not in ("cat", "tee"):
            continue
        if words[0] == "cat":
            t = CAT_TARGET_RE.search(seg)
            target = t.group(1) if t else None
        else:
            target = next((w for w in words[1:] if not w.startswith(("-", ">"))), None)
        delim = m.group(1) or m.group(2) or m.group(3)
        tabs = line[m.start() : m.start() + 3] == "<<-"
        end = next(
            (j for j in range(k, len(lines)) if (lines[j].lstrip("\t") if tabs else lines[j]).rstrip("\r") == delim),
            None,
        )
        if not target or end is None or target in "\n".join(lines[end + 1 :]):
            continue
        if target.startswith("/dev/") or "(" in target:
            continue
        if may_run_code("\n".join(lines[:k] + lines[end + 1 :])):
            continue
        if m.group(3) and any(c in "\n".join(lines[k:end]) for c in ("$(", "`")):
            continue
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


def unarchived_landing(args, target):
    """(target, change) when the push args land `<src>:<default>` with
    spectre/changes/<change>/ still open in <src>'s tree, else None."""
    if target is None:
        return None
    default = default_branch(target)
    for a in args:
        if ":" not in a or a.startswith("-"):
            continue
        src, _, dst = a.lstrip("+").partition(":")
        dst = dst.removeprefix("refs/heads/")
        if not src or dst not in ({default} if default else FALLBACK_DEFAULT_BRANCHES):
            continue
        change = git(target, "branch", "--show-current") if src == "HEAD" else src
        change = (change or "").removeprefix("refs/heads/").removeprefix("spectre/")
        if change and git(target, "ls-tree", "-d", "--name-only", src, f"spectre/changes/{change}"):
            return target, change
    return None


SEPARATORS = ("&&", "||", ";", "|")
# Push flags that push every branch, the default branch among them.
ALL_BRANCH_FLAGS = ("--all", "--mirror", "--branches")


def push_refspecs(args, target):
    """(src, dst) for each refspec of `git push <args>`, dst a bare branch name:
    `<src>:<dst>`, a colon-less `<b>` as `<b>:<b>`, HEAD and no refspec at all as
    the current branch."""
    pos = [a for a in args if not a.startswith("-")]
    specs = pos[1:] or ["HEAD"]
    out = []
    if any(a in ALL_BRANCH_FLAGS for a in args):
        default = default_branch(target)
        out += [(d, d) for d in ({default} if default else FALLBACK_DEFAULT_BRANCHES)]
    for a in specs:
        src, colon, dst = a.lstrip("+").partition(":")
        if not colon:
            dst = src
        if dst == "HEAD":
            dst = git(target, "branch", "--show-current") or ""
        out.append((src, dst.removeprefix("refs/heads/")))
    return out


def unreadable_landing(args, target):
    """The deny reason when the push may land on the default branch and the hook cannot read it:
    its repository is unknown (target None, a `-C $W` or `cd $W`), or any of its words holds a
    shell expansion left unexpanded, and some refspec's destination is the default branch, an
    expansion, or unknown. Else None."""
    if target is not None and not any("$" in a or "`" in a for a in args):
        return None
    default = default_branch(target) if target else None
    defaults = {default} if default else FALLBACK_DEFAULT_BRANCHES
    unknown = any((a.startswith("-") and ("$" in a or "`" in a)) or a in ALL_BRANCH_FLAGS for a in args)
    pos = [a for a in args if not a.startswith("-")]
    for a in pos[1:] or ["HEAD"]:
        src, colon, dst = a.lstrip("+").partition(":")
        dst = (dst if colon else src).removeprefix("refs/heads/")
        if dst == "HEAD":
            dst = git(target, "branch", "--show-current") if target else None
        if not dst or "$" in dst or "`" in dst or dst in defaults:
            unknown = True
    if not unknown:
        return None
    return (
        "Blocked: this push may land on the default branch, and part of it — the repository "
        "(`-C`, `cd`), the remote or a refspec — goes through a shell expansion (`$VAR`, "
        "`${VAR}`, `$(...)`, a backtick) the hook cannot resolve, so its landing checks (an open "
        "spectre change, a fix loop's records) cannot see what it lands. Write the push with "
        "literal names:\n\n"
        "  git -C /abs/path/to/worktree push origin <branch>:<default-branch>\n\n"
        "A fix loop lands with scripts/pipeline-fix.sh land <worktree>."
    )


def loop_landing(args, target):
    """The deny reason when the push lands, on the default branch, a commit that
    contains the tip of a loop branch (or a change branch) not yet on
    origin/<default> which check-loop-landing.sh refuses, else None. The pushed
    <src> is resolved and every such branch it contains is checked, so a
    landing under another spelling -- `@:main`, `<sha>:main`, a temporary
    branch, `pull . <loop>` then `push origin main` -- is checked too."""
    if target is None:
        return None
    global LOOP_DEADLINE
    default = default_branch(target)
    # One budget per command, not per push: the hook is one process per command.
    if LOOP_DEADLINE is None:
        # LOOP_LANDING_TIMEOUT: the hook's own harness only.
        LOOP_DEADLINE = time.monotonic() + float(os.environ.get("LOOP_LANDING_TIMEOUT") or LOOP_CHECK_TIMEOUT)
    deadline = LOOP_DEADLINE
    for src, dst in push_refspecs(args, target):
        if not src or dst not in ({default} if default else FALLBACK_DEFAULT_BRANCHES):
            continue
        # ponytail: a <src> this same command creates (`git branch tmp && git push origin
        # tmp:main`) is not there yet -- HEAD stands in for it, the commit it is made from.
        sha = git(target, "rev-parse", "--verify", "-q", f"{src}^{{commit}}") or git(
            target, "rev-parse", "--verify", "-q", "HEAD"
        )
        if not sha:
            continue
        cmd = ["for-each-ref", "--format=%(refname:short)", "--merged", sha]
        if git(target, "rev-parse", "--verify", "-q", f"refs/remotes/origin/{dst}"):
            cmd += ["--no-merged", f"refs/remotes/origin/{dst}"]
        cmd += ["refs/heads/pipeline-fix-*", "refs/heads/self-review-*", "refs/heads/spectre/"]
        for branch in (git(target, *cmd) or "").split():
            if not LOOP_BRANCH_RE.match(branch):
                continue
            try:
                r = subprocess.run(
                    [LOOP_CHECK, target, branch], capture_output=True, text=True,
                    timeout=max(1, deadline - time.monotonic()),
                )
            except (OSError, subprocess.SubprocessError) as e:
                return f"Blocked: landing {branch} needs check-loop-landing.sh, which could not run ({e})."
            if r.returncode != 0:
                return (
                    f"Blocked: {branch} is a fix loop, and its records do not cover this landing "
                    f"(check-loop-landing.sh exit {r.returncode}):\n{r.stdout}{r.stderr}\n"
                    "Land it with scripts/pipeline-fix.sh land <worktree> once they do."
                )
    return None


def untasked_push(args, target):
    """The deny reason when the push sends HEAD's commit to a `spectre/` branch origin holds and
    check-untasked-commits.sh lists an unpushed commit carrying neither trailer, else None."""
    if target is None:
        return None
    head = git(target, "rev-parse", "--verify", "-q", "HEAD")
    for src, dst in push_refspecs(args, target):
        remote = f"refs/remotes/origin/{dst}"
        if not dst.startswith("spectre/") or not git(target, "rev-parse", "--verify", "-q", remote):
            continue
        if not head or git(target, "rev-parse", "--verify", "-q", f"{src}^{{commit}}") != head:
            continue
        try:
            r = subprocess.run(
                [UNTASKED_CHECK, target, remote, "--not-remotes"], capture_output=True, text=True, timeout=UNTASKED_CHECK_TIMEOUT
            )
        except (OSError, subprocess.SubprocessError):
            return None
        if r.returncode == 1:
            return (
                f"Blocked: a commit this push sends to {dst} sits outside every task with neither a "
                f"`Task-Id:` nor a `Flow-Key:` trailer:\n{r.stdout}\n"
                "Each listed commit is on no origin ref. Add `-m \"Flow-Key: <key>\"` to it — `git commit "
                "--amend` for HEAD, a reword for an earlier listed one, never a commit outside this list — `<key>` its round's dispatch key or stage key (**Commits outside every "
                "task**, skills/flow-contracts/git-boundaries.md), then push again."
            )
    return None


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
                if j < n and tokens[j] == "push" and landings is not None:
                    args = []
                    for t in tokens[j + 1 :]:
                        if t in ("&&", "||", ";", "|"):
                            break
                        args.append(t)
                    unreadable = unreadable_landing(args, target)
                    if unreadable:
                        landings.append(unreadable)
                    hit = None if unreadable else unarchived_landing(args, target)
                    if hit:
                        landings.append(landing_reason(*hit))
                    loop = None if unreadable else loop_landing(args, target)
                    if loop:
                        landings.append(loop)
                    untasked = None if unreadable else untasked_push(args, target)
                    if untasked:
                        landings.append(untasked)
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
            if tok in (">", ">>") and i + 1 < n:
                hits.append(resolve(tokens[i + 1], cur))
                i += 2
                continue
            if tok.startswith((">", ">>")) and len(tok) > 2 and tok.lstrip(">") not in ("&1", "&2"):
                hits.append(resolve(tok.lstrip(">"), cur))
                i += 1
                continue
            if tok in ("cp", "mv", "rm"):
                args = []
                for t in tokens[i + 1 :]:
                    if t in ("&&", "||", ";", "|"):
                        break
                    if not t.startswith("-"):
                        args.append(t)
                if tok in ("rm", "mv"):
                    # mv removes its sources too; neither follows a symlink it is given
                    hits.extend(LinkItself(r) for r in (resolve(a, cur) for a in args) if r)
                elif args:
                    hits.append(resolve(args[-1], cur))
                i += 1
                continue
            i += 1
    return [h for h in hits if h]


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


def landing_reason(top, change):
    return (
        f"Blocked: this push lands `spectre/changes/{change}/` on the default branch still open. "
        "A landed change is archived first, in the same branch:\n\n"
        f"  spectre archive --force {change}   # --force only when it has no tasks.md\n"
        f"  git commit -m \"chore(spectre): archive {change}\"\n\n"
        "then push again."
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
        json.dump(
            {
                "hookSpecificOutput": {
                    "hookEventName": "PreToolUse",
                    "permissionDecision": "deny",
                    "permissionDecisionReason": deny_reason(*hit),
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
