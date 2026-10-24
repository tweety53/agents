#!/usr/bin/env bash
# check-verify-report.sh — answer whether a flow.visual-verify verifier's
# report is complete, so the parent never accepts a partial one by reading
# its prose (KAN-870: two verifiers ran the tests and the capture, listed
# steps 9-11 under "Not done", and both reports were accepted).
#
# Usage: check-verify-report.sh <report file> <motions named> [<scoped specs>]
#
# <report file> is the relay contract's
# <abs-worktree>/.superpowers/sdd/verify-report-<key>.md
# (skills/flow/visual-verify.md); <motions named> is the count of motions the
# dispatch prompt named, 0 for `motions: none`; <scoped specs>, given for a
# re-run under a round base alone, is the count of paths `specs` printed for
# that round base (skills/flow/verify-fix-loop.md).
#
# The report's `- steps:` line carries one `<step> <status>` item per required
# step, ` | `-separated, for steps 4, 7, 8, 9, 10, 11 and `motion`
# (skills/flow/visual-verify-verifier.md). A status is admissible when it is
#   done
#   n/a — <reason>       steps 4 and 10 only, and `motion` when <motions named> is 0;
#                        steps 8, 9 and 10 too when a `- wire contract:` line
#                        names an endpoint — the dispatch named no view, so
#                        nothing is captured (kan-1022's self-review)
#   blocked — <reason>   the reason citing a non-zero `exit <n>`; step 10 only
#                        when step 8 is blocked too or the reason cites
#                        `compose-mockup-frames exit <n>` (KAN-1000)
# and `motion done` with <motions named> above 0 also needs a
# `- motion: <n>/<n>` line, <n> that count — a strip recorded beyond the
# named motions goes on a sub-bullet, never into <n>. Every other status — absent,
# `not done`, an n/a on a required step, a block citing no failing exit — is
# an undone step. A `- fingerprint: not declared` line with no ` — <reason>`
# is undone too: step 6 reports its exit 2 with the guard's reason, which the
# prompt carries (kan-1010's self-review).
#
# A `## Report` line attributing a departure as `pre-existing` or `present at
# the merge base` that names neither `merge-base baseline: <sha>:<path>` nor
# `merge-base capture: <path>` leaves step 10 undone (kan-1015: a re-baseline
# over a defect attributed from the code path alone).
#
# Given <scoped specs>, the report's `- verify:` line must read `not run — no
# spec since the round base` when it is 0, else name `<n> mapped specs` with
# <n> equal to it; otherwise step 7 is undone as `ran <k> mapped specs where
# specs printed <n>` or `names no mapped-spec count` (kan-1015: a re-run under
# a round base where `specs` printed 2 ran 95).
#
# Prints ONE verdict line to stdout:
#   VERIFY-REPORT-COMPLETE: <report file>
#   VERIFY-REPORT-INCOMPLETE: <report file> — <each undone step and its status, `; `-separated>
#
# Exit 0 complete, 1 incomplete, 2 cannot answer (a usage error, a count
# or scoped count that is not a non-negative integer, a report file it
# cannot read).
#
# Go, in flow-guard (stats/internal/guard/verifyreport.go).
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-verify-report: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-verify-report 2 "check-verify-report:" "$@"
