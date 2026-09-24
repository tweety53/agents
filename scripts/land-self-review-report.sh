#!/usr/bin/env bash
# land-self-review-report.sh — the one landing chain for a self-review
# report or context bundle (kan-523), extracted from the prose shells at
# skills/flow-self-review/SKILL.md step 5 and skills/flow/archive.md step 9
# so the chain is correct once and harness-covered instead of
# per-prose-copy. The chain, in order: re-assert the branch (F3 — nothing
# at all runs on a mismatch), git add the report, optionally git rm the
# context bundle, skip cleanly when nothing is staged, refuse loudly when
# the staged set is anything beyond the chain's own paths (kan-657 — a
# bare git commit takes the whole index, so a shared checkout's foreign
# staged work would be swept in), re-assert the branch again before the
# commit and once more before the pull/push pair (a shared checkout can
# be switched mid-run; the start-of-run assert cannot be trusted across
# the run), commit the subject, and — only under --push <base> — git
# pull --rebase origin <base> then git push origin <base>, both inside
# the same guard (F4 — they can never act on a branch other than the
# asserted one). Every git call goes through git -C <repo>, so whatever
# the caller has checked out is never consulted.
#
# Usage:
#   land-self-review-report.sh <repo> <branch> <subject> <add-path> [<rm-path>] [--push <base>]
#
# Exit codes:
#   0  landed, or nothing to land (one LAND-NOTHING-TO-COMMIT line)
#   1  branch mismatch — one LAND-BRANCH-MISMATCH line; past the
#      start-of-run check, nothing further runs
#   2  usage
#   3  foreign staged work — one LAND-FOREIGN-STAGED line naming it; the
#      commit is refused and nothing is rolled back (the chain's own
#      add/rm stay staged beside the foreign paths)
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

# The start-of-run assert cannot be trusted across the run: a concurrent
# session on a shared checkout can switch the branch between two steps
# (observed, kan-657). The commit and the pull/push pair each re-check
# before an irreversible step runs on the wrong branch.
assert_branch() {
  local stage="$1" found
  found="$(g branch --show-current)"
  [ "$found" = "$BRANCH" ] && return 0
  echo "LAND-BRANCH-MISMATCH: expected $BRANCH, found $found $stage" >&2
  exit 1
}

assert_branch "— nothing added, committed, pulled or pushed"

g add -- "$ADD_PATH"
if [ -n "$RM_PATH" ]; then
  g rm -- "$RM_PATH"
fi

if g diff --cached --quiet; then
  echo "LAND-NOTHING-TO-COMMIT: nothing staged under $REPO — no commit, no pull, no push"
  exit 0
fi

# git commit takes the whole index, and the chain staged only its own
# paths: a shared checkout holding foreign staged work must not be swept
# into the commit (kan-657, where 121 foreign paths landed as 66ae176 and
# reverted a just-merged change). Refuse loudly, index untouched; the own-paths set is deliberately stated twice — in the add/rm calls and in EXPECTED below — because the staging verbs and this assertion must agree, and each side reads it in its own grammar.
STAGED="$(g diff --cached --name-only --no-renames | sort)"
if [ -n "$RM_PATH" ]; then
  EXPECTED="$(printf '%s\n%s\n' "$ADD_PATH" "$RM_PATH" | sort)"
else
  EXPECTED="$(printf '%s\n' "$ADD_PATH" | sort)"
fi
if [ "$STAGED" != "$EXPECTED" ]; then
  FOREIGN="$(comm -13 <(printf '%s\n' "$EXPECTED") <(printf '%s\n' "$STAGED") | tr '\n' ' ')"
  echo "LAND-FOREIGN-STAGED: expected only $(printf '%s' "$EXPECTED" | tr '\n' ' ')— foreign staged: ${FOREIGN}— missing from the index: $(comm -23 <(printf '%s\n' "$EXPECTED") <(printf '%s\n' "$STAGED") | tr '\n' ' ')— nothing committed, pulled or pushed; clear the staging or land from a clean checkout" >&2
  exit 3
fi

assert_branch "before the commit — nothing committed, pulled or pushed"

g commit -m "$SUBJECT" -- "$ADD_PATH" ${RM_PATH:+"$RM_PATH"}

if [ -n "$PUSH_BASE" ]; then
  assert_branch "before the pull/push — nothing pulled or pushed"
  g pull --rebase origin "$PUSH_BASE"
  g push origin "$PUSH_BASE"
fi
