#!/usr/bin/env bash
# resolve-base-branch.sh — the single owner of base-branch resolution.
#
# Usage: resolve-base-branch.sh <dir>
#
# stdout carries the bare branch name and NOTHING ELSE on success, so a
# caller composes it directly: BASE="$(resolve-base-branch.sh "$WT")". Every
# refusal writes its reason to stderr and leaves stdout empty — an empty
# stdout must never be readable as a resolved base.
#
#   Exit 0   Resolved; the name is on stdout.
#   Exit 1   A named refusal — detached HEAD, no base resolved, base equal
#            to the current branch, or a base name that fails validation.
#            Validation requires the first character to be one of
#            [A-Za-z0-9._] (so a leading '-' or a leading '/' is refused)
#            and every character in the name to be one of [A-Za-z0-9._/-].
#   Exit 2   Cannot answer — the argument is missing, <dir> is absent,
#            unreadable, or not a git worktree, or HEAD's own ref cannot be
#            read (corrupt or permission-denied).
#   Exit 3   The repository has no 'origin' remote at all.
#
# WHY THE FETCH IS WRAPPED. `git remote show origin` against an unreachable
# host blocks for roughly 75s on the default TCP timeout, which would turn a
# correct refusal into a two-minute hang. `-c core.askpass=true` stops it
# prompting for credentials, `2>/dev/null` swallows the chatter, and `|| true`
# means a failed fetch is not this script's failure — a stale origin/HEAD is
# still resolved, just possibly out of date, which the fetch on the next
# invocation corrects.
#
# WHY HEAD@{upstream} IS NEVER CONSULTED. Inside an apply worktree, HEAD *is*
# the change's own branch, openspec/<name>, by construction. Its upstream is
# origin/openspec/<name>, so HEAD@{upstream} would compare the branch with
# itself the moment it is pushed — always "true", never useful. The base
# branch is resolved from origin/HEAD (or the `git remote show origin`
# fallback), never from HEAD's own upstream.
#
# WHY EXIT 3 IS SEPARATE FROM EXIT 2. The finish contract requires a distinct
# no-remote message. A caller that could only tell the two apart by matching
# stderr text would stop being able to the first time the wording changes.
#
# The unreachable-remote hang is deliberately NOT covered by this script's
# test harness: a wall-clock assertion on a TCP timeout is flaky by
# construction. The wrapping above is carried forward verbatim from the
# fenced block this script replaces, and this comment is where its reasoning
# now lives.
set -euo pipefail

# NO REFUSAL BELOW MAY DEPEND ON THE OPERATOR'S LOCALE. `A-Za-z0-9` inside a
# `case` bracket expression is a COLLATING range, not a byte range: under
# `en_US.UTF-8` on this repository's bash 3.2 floor, a base name carrying a
# non-ASCII byte such as `é` collates inside that range and both validation
# blocks below admit it, printing it on stdout — exactly the value the spec's
# unconditional-assertions requirement says must be refused. Under `LC_ALL=C`
# the same value is correctly rejected. `export LC_ALL=C` for the whole
# script — matching check-panel-reproducers.sh's precedent, not
# check-cleanup-complete.sh's enumerated allowlist — because this script
# invokes no downstream project-supplied command whose own locale must be
# preserved: everything after this line is git plumbing and shell builtins,
# so pinning the locale here costs nothing outside this file.
export LC_ALL=C

DIR="${1:-}"
if [ -z "$DIR" ]; then
  echo "usage: resolve-base-branch.sh <dir>" >&2
  exit 2
fi

if [ ! -d "$DIR" ]; then
  echo "resolve-base-branch: $DIR is not a directory" >&2
  exit 2
fi

if ! git -C "$DIR" rev-parse --git-dir >/dev/null 2>&1; then
  echo "resolve-base-branch: $DIR is not a git worktree" >&2
  exit 2
fi

if ! git -C "$DIR" remote get-url origin >/dev/null 2>&1; then
  echo "resolve-base-branch: no 'origin' remote configured in $DIR" >&2
  exit 3
fi

git -C "$DIR" -c core.askpass=true fetch --quiet origin 2>/dev/null || true

# Each resolution attempt is allowed to fail — a missing origin/HEAD or an
# unreachable remote is an ordinary outcome here, not a script error — so
# each is followed by `|| true`. Without it, `set -o pipefail` makes the
# git side of a failing pipe the pipeline's exit status, and `set -e` would
# abort the whole script right here, silently, before either fallback or
# refusal ever runs.
BASE="$(git -C "$DIR" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's#^origin/##')" || true
if [ -z "$BASE" ]; then
  BASE="$(GIT_TERMINAL_PROMPT=0 git -C "$DIR" remote show origin 2>/dev/null | sed -n 's/^ *HEAD branch: //p')" || true
fi

# `branch --show-current` fails (non-zero exit) when HEAD's own ref cannot be read — a corrupt
# or permission-denied ref, for instance — which is a DIFFERENT fact from detached HEAD, where the
# same command succeeds and simply prints nothing. Telling them apart matters: reported as
# "detached", an unreadable ref would be a misleading diagnosis of a tree resolution genuinely
# cannot read. So this is refused as its own case, exit 2 ("cannot answer"), matching how the
# missing-directory and not-a-worktree checks above already answer "cannot read this tree" — not
# folded into the exit-1 detached-HEAD branch below, which is reserved for the empty-but-successful
# case.
if ! CUR="$(git -C "$DIR" branch --show-current 2>/dev/null)"; then
  echo "resolve-base-branch: could not read the current branch in $DIR" >&2
  exit 2
fi

if [ -z "$CUR" ]; then
  echo "resolve-base-branch: HEAD is detached in $DIR" >&2
  exit 1
fi

if [ -z "$BASE" ]; then
  echo "resolve-base-branch: could not resolve a base branch in $DIR" >&2
  exit 1
fi

if [ "$BASE" = "$CUR" ]; then
  echo "resolve-base-branch: base branch '$BASE' is the same as the current branch — refusing to compare a branch with itself" >&2
  exit 1
fi

# BASE arrived from the remote and flows into refs a caller builds
# (origin/$BASE) and into further git commands. Refuse — without echoing the
# untrusted value back — unless it matches this shape: no leading `-` a
# downstream git call would read as an option, and no control character.
case "$BASE" in
  [A-Za-z0-9._]*) ;;
  *)
    echo "resolve-base-branch: the resolved base branch name is invalid" >&2
    exit 1
    ;;
esac
case "$BASE" in
  *[!A-Za-z0-9._/-]*)
    echo "resolve-base-branch: the resolved base branch name is invalid" >&2
    exit 1
    ;;
esac

printf '%s\n' "$BASE"
