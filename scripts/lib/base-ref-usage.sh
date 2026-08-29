# scripts/lib/base-ref-usage.sh — base_ref_usage_message, shared by the
# finish-preflight and base-moved test harnesses (test-check-finish-preflight.sh,
# test-check-base-moved.sh) so the golden usage text each asserts against is
# written once. KAN-298 fix round 2 (panel-report-1-principles.md, F3):
# extracted after the two harnesses' EXPECTED_USAGE literals — identical
# apart from the guard name — were found to differ only in a wording edit
# waiting to happen twice.
#
# TEST-ONLY, DELIBERATELY IN scripts/lib/, matching test-git-shim.sh's own
# placement and its header's reasoning: this file's only two consumers are
# test harnesses, not production guards, but nothing in this repository
# scopes scripts/lib/ to production-only members.
#
# THE TEXT HERE IS AN INDEPENDENT GOLDEN VALUE, NOT A DERIVATION FROM EITHER
# GUARD. It must never be produced by reading, sourcing, or executing
# check-finish-preflight.sh or check-base-moved.sh — doing so would make
# each harness's assertion tautological (it would pass even if the guard's
# real usage message were wrong), which is exactly the regression coverage
# rounds 0 and 1 were spent catching.
#
# Not meant to be executed directly — a caller sources it and calls
# base_ref_usage_message; it sets no `set -euo pipefail` of its own and
# relies on the sourcing script's.

# base_ref_usage_message <guard-name> — print to stdout the five-line usage
# message both check-finish-preflight.sh and check-base-moved.sh emit on
# their missing-arguments path, with <guard-name> substituted into line 1
# (e.g. "check-finish-preflight.sh" or "check-base-moved.sh").
base_ref_usage_message() {
  local guard_name="$1"
  printf 'usage: %s <worktree> <base-ref> <recorded-merge-base|->\n' "$guard_name"
  printf '  <base-ref>  the base branch name, bare (main) or remote-tracking\n'
  printf '              (origin/main). The guard prefers refs/remotes/origin/<base-ref>\n'
  printf '              when it resolves, so a bare name is never tested against a\n'
  printf '              stale local branch.'
}
