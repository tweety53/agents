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
    redirect, `cp`/`mv` destinations, or `rm`.

"Main checkout" is the worktree whose git dir IS the common dir — a linked worktree's `--git-dir`
is `<common>/worktrees/<name>`, so `.worktrees/<change>` and `<repo>-worktrees/<change>` alike
pass. "Default branch" is what `origin/HEAD` points at; when the remote head is unknown, a checkout
of main, master or develop counts. A main checkout on a feature branch is not protected — leaving
the default branch is how work gets a branch of its own.

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
in a repo's reflog, rather than trying to model the shell.
"""

import json
import os
import shlex
import subprocess
import sys

EDIT_TOOLS = {"Edit", "Write", "MultiEdit", "NotebookEdit"}

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


def protected(path):
    """(toplevel, branch) when path lies in a main checkout on its default branch, else None."""
    d = existing_dir(path)
    if d is None:
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
    head = git(d, "symbolic-ref", "-q", "--short", "refs/remotes/origin/HEAD")
    default = head.split("/", 1)[1] if head and "/" in head else None
    if default is None and branch in FALLBACK_DEFAULT_BRANCHES:
        default = branch
    if branch != default:
        return None
    return top, branch


def resolve(path, cwd):
    return path if os.path.isabs(path) else os.path.join(cwd, path)


def bash_hits(command, cwd):
    """Paths a Bash command writes into or git-mutates, resolved against cwd and `cd`."""
    try:
        tokens = shlex.split(command)
    except ValueError:
        tokens = command.split()
    hits = []
    cur = cwd
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
            i = j + 1
            continue
        if tok == "sed" and any(t.startswith("-i") for t in tokens[i + 1 : i + 4]):
            for t in tokens[i + 1 :]:
                if t in ("&&", "||", ";", "|"):
                    break
                if not t.startswith("-") and os.path.exists(resolve(t, cur)):
                    hits.append(resolve(t, cur))
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
            if tok == "rm":
                hits.extend(resolve(a, cur) for a in args)
            elif args:
                hits.append(resolve(args[-1], cur))
            i += 1
            continue
        i += 1
    return hits


def deny_reason(top, branch):
    name = os.path.basename(top)
    return (
        f"Blocked: {top} is the main checkout of {name}, and it is on `{branch}`, the default "
        "branch. Agents never edit, stage, commit, stash or reset there — it is the landing "
        "target every pipeline run pushes to, and anything left behind becomes residue no later "
        "session owns.\n\n"
        "Work in a worktree on a branch of its own instead:\n\n"
        f"  git -C {top} worktree add {top}/.worktrees/<change> -b <change> origin/{branch}\n\n"
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
            candidates.extend(bash_hits(command, cwd))
    else:
        return 0

    for path in candidates:
        hit = protected(path)
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
