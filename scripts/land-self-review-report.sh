#!/usr/bin/env bash
# land-self-review-report.sh — the one landing chain for a self-review
# report or context bundle (kan-523), extracted from the prose shells at
# skills/flow-self-review/SKILL.md step 7 and skills/flow/archive.md step 9
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
#
# The logic is the Go port in stats/internal/guard/landselfreviewreport.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 with the cause when it
# cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "land-self-review-report: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec land-self-review-report 2 "land-self-review-report:" "$@"
