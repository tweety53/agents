#!/usr/bin/env bash
# check-verify-report.sh — answer whether a flow.visual-verify verifier's
# report is complete, so the parent never accepts a partial one by reading
# its prose (KAN-870: two verifiers ran the tests and the capture, listed
# steps 9-11 under "Not done", and both reports were accepted).
#
# Usage: check-verify-report.sh <report file> <motions named>
#
# <report file> is the relay contract's
# <abs-worktree>/.superpowers/sdd/verify-report-<key>.md
# (skills/flow/visual-verify.md); <motions named> is the count of motions the
# dispatch prompt named, 0 for `motions: none`.
#
# The report's `- steps:` line carries one `<step> <status>` item per required
# step, ` | `-separated, for steps 4, 7, 8, 9, 10, 11 and `motion`
# (skills/flow/visual-verify-verifier.md). A status is admissible when it is
#   done
#   n/a — <reason>       steps 4 and 10 only, and `motion` when <motions named> is 0
#   blocked — <reason>   the reason citing a non-zero `exit <n>`
# and `motion done` with <motions named> above 0 also needs a
# `- motion: <n>/<n>` line, <n> that count. Every other status — absent,
# `not done`, an n/a on a required step, a block citing no failing exit — is
# an undone step.
#
# Prints ONE verdict line to stdout:
#   VERIFY-REPORT-COMPLETE: <report file>
#   VERIFY-REPORT-INCOMPLETE: <report file> — <each undone step and its status, `; `-separated>
#
# Exit 0 complete, 1 incomplete, 2 cannot answer (a usage error, a count
# that is not a non-negative integer, a report file it cannot read).
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
