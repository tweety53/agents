#!/usr/bin/env bash
# land-self-review-report.sh — the one landing chain for a self-review
# report or context bundle (kan-523), extracted from the prose shells at
# skills/flow-self-review/SKILL.md step 5 and skills/flow/archive.md step 9
# so the chain is correct once and harness-covered instead of
# per-prose-copy. The chain, in order: re-assert the branch (F3 — nothing
# at all runs on a mismatch), git add the report, optionally git rm the
# context bundle, skip cleanly when nothing is staged, commit the subject,
# and — only under --push <base> — git pull --rebase origin <base> then
# git push origin <base>, both inside the same guard (F4 — they can never
# act on a branch other than the asserted one). Every git call goes
# through git -C <repo>, so whatever the caller has checked out is never
# consulted.
#
# Usage:
#   land-self-review-report.sh <repo> <branch> <subject> <add-path> [<rm-path>] [--push <base>]
#
# Exit codes:
#   0  landed, or nothing to land (one LAND-NOTHING-TO-COMMIT line)
#   1  branch mismatch — one LAND-BRANCH-MISMATCH line, nothing run
#   2  usage
#   otherwise git's own exit code, unmasked: a rejected push leaves the
#   commit local and names itself on stderr; this script never retries it.
#   A pull --rebase that stops on a conflict leaves <repo> mid-rebase with
#   nothing to auto-recover — resolve or `git rebase --abort` by hand; this
#   script never aborts one
set -euo pipefail

usage() {
  echo "usage: land-self-review-report.sh <repo> <branch> <subject> <add-path> [<rm-path>] [--push <base>]" >&2
}

[ "$#" -ge 4 ] || { usage; exit 2; }

REPO="$1"
BRANCH="$2"
SUBJECT="$3"
ADD_PATH="$4"
shift 4

RM_PATH=""
PUSH_BASE=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--push" ]; then
    { [ "$#" -ge 2 ] && [ -n "$2" ]; } || { usage; exit 2; }
    PUSH_BASE="$2"
    shift 2
  else
    [ -z "$RM_PATH" ] || { usage; exit 2; }
    RM_PATH="$1"
    shift
  fi
done

g() { git -C "$REPO" "$@"; }

FOUND="$(g branch --show-current)"
if [ "$FOUND" != "$BRANCH" ]; then
  echo "LAND-BRANCH-MISMATCH: expected $BRANCH, found $FOUND — nothing added, committed, pulled or pushed" >&2
  exit 1
fi

g add -- "$ADD_PATH"
if [ -n "$RM_PATH" ]; then
  g rm -- "$RM_PATH"
fi

if g diff --cached --quiet; then
  echo "LAND-NOTHING-TO-COMMIT: nothing staged under $REPO — no commit, no pull, no push"
  exit 0
fi

g commit -m "$SUBJECT"

if [ -n "$PUSH_BASE" ]; then
  g pull --rebase origin "$PUSH_BASE"
  g push origin "$PUSH_BASE"
fi
