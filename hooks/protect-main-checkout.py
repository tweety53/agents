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
    redirect, `cp` destinations, or `mv`/`rm` sources and destinations.

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
an ad-hoc landing meets this.

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

EDIT_TOOLS = {"Edit", "Write", "MultiEdit", "NotebookEdit"}

ASSIGNMENT_RE = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*=")

# git verbs that change the index, the worktree, a ref or the stash list.
DENIED_GIT_VERBS = {
    "add", "am", "apply", "cherry-pick", "clean", "commit", "merge", "mv", "rebase",
    "reset", "restore", "revert", "rm", "stash", "update-index", "update-ref",
}

FALLBACK_DEFAULT_BRANCHES = {"main", "master", "develop"}

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


def protected(path):
    """(toplevel, branch) when path lies in a main checkout on its default branch, else None."""
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


def bash_hits(command, cwd, landings=None):
    """Paths a Bash command writes into or git-mutates, resolved against cwd and `cd`.

    The command is scanned one logical line at a time — a newline ends an
    argument scan the way `;` does, so a multi-line script's `cp` cannot
    swallow a later line's arguments — while `cd` state carries across lines,
    as it does in bash. Assignments the command text itself makes are expanded
    before path resolution, positionally: a token sees the assignments that
    came before it, in this line and earlier lines, never the ones after —
    and an assignment's value is kept literal, unexpanded."""
    hits = []
    cur = cwd
    env = {}
    for line in logical_lines(command):
        raw_tokens = tokenize(line)
        tokens = []
        for raw in raw_tokens:
            if ASSIGNMENT_RE.match(raw):
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
            if tok == "git":
                target = cur
                j = i + 1
                while j < n and tokens[j].startswith("-"):
                    if tokens[j] == "-C" and j + 1 < n:
                        target = resolve(tokens[j + 1], cur)
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
                    hit = unarchived_landing(args, target)
                    if hit:
                        landings.append(hit)
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
                    hits.extend(resolve(a, cur) for a in args)  # mv removes its sources too
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
                            "permissionDecisionReason": landing_reason(*landings[0]),
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
