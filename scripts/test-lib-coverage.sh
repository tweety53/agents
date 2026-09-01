#!/usr/bin/env bash
# Assertion harness for scripts/lib/coverage.sh — the shared per-member
# coverage library tasks 2-5 (check-guard-symlinks.sh, check-references.sh,
# check-vocabulary.sh, check-stage-mark-calls.sh) each source. Sources the
# library directly rather than through any guard, per kan-197's task 1: the
# thing under test is the library's own contract, not any guard's use of it.
#
# The five cases task 1 names: a member with a non-zero count; a member at
# zero AND declared; a member at zero and undeclared (named, non-zero exit);
# an empty corpus (itself a violation, not a vacuous pass); and a member name
# carrying a space or a leading `-`, handled rather than mis-parsed. A short
# "robustness" section beyond those five exercises the library's own
# fail-fast argument validation (bad count, empty member, a duplicate
# record/declare) — reject rather than guess, applied here to this
# library's own arguments rather than to a grep call.
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records:
# indexed arrays only, no associative arrays. Every array here — and every
# array coverage.sh itself keeps — is walked by its indices
# ("${!ARRAY[@]}"), never its values ("${ARRAY[@]}") directly: the latter is
# unbound-variable-under-`set -u` when the array is empty, verified against
# /bin/bash 3.2.57 on this machine (the indices form and `${#ARRAY[@]}` are
# both empty-safe on the same bash; only the values form is not).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB="$SCRIPT_DIR/lib/coverage.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

if [ ! -r "$LIB" ]; then
  echo "test-lib-coverage: cannot read $LIB" >&2
  exit 2
fi
# shellcheck source=lib/coverage.sh
source "$LIB"

# assert_eq <label> <expected> <actual>
assert_eq() {
  if [ "$2" = "$3" ]; then
    pass "$1"
  else
    fail "$1 — expected [$2], got [$3]"
  fi
}

# assert_contains <label> <haystack> <needle>
assert_contains() {
  case "$2" in
    *"$3"*) pass "$1" ;;
    *) fail "$1 — [$2] does not contain [$3]" ;;
  esac
}

# assert_nonzero_rc <label> <rc>
assert_nonzero_rc() {
  if [ "$2" -ne 0 ]; then
    pass "$1"
  else
    fail "$1 — expected a non-zero exit, got 0"
  fi
}

# assert_zero_rc <label> <rc>
assert_zero_rc() {
  if [ "$2" -eq 0 ]; then
    pass "$1"
  else
    fail "$1 — expected exit 0, got $2"
  fi
}

# run_verdict -> sets VERDICT (stdout) and RC. coverage_verdict is expected
# to return non-zero on a violation, which must not abort this harness under
# `set -e` — the same set +e/set -e bracket
# test-check-panel-reproducers.sh's header describes for a command that is
# SUPPOSED to fail sometimes.
run_verdict() {
  set +e
  VERDICT="$(coverage_verdict)"
  RC=$?
  set -e
}

# ---------------------------------------------------------------------------
# 1. A member with a non-zero count renders and passes.
# ---------------------------------------------------------------------------
coverage_reset
coverage_record "myflow-do" 14
assert_eq "case 1: a non-zero count renders as 'name count'" "myflow-do 14" "$(coverage_report)"
run_verdict
assert_eq "case 1: a non-zero count produces no violation lines" "" "$VERDICT"
assert_zero_rc "case 1: a non-zero count exits 0" "$RC"

# ---------------------------------------------------------------------------
# 2. A member at zero AND declared renders as declared and passes — with and
#    without a cited reason (task 5's "declaration should cite that reason").
# ---------------------------------------------------------------------------
coverage_reset
coverage_record "flow-status" 0
coverage_declare "flow-status" "read-only report; writes no stage marks by contract"
assert_eq "case 2: a declared zero renders with its cited reason" \
  "flow-status 0 (declared: read-only report; writes no stage marks by contract)" \
  "$(coverage_report)"
run_verdict
assert_eq "case 2: a declared zero produces no violation lines" "" "$VERDICT"
assert_zero_rc "case 2: a declared zero exits 0" "$RC"

coverage_reset
coverage_record "myflow-start" 0
coverage_declare "myflow-start"
assert_eq "case 2b: a declared zero with no reason renders plain '(declared)'" \
  "myflow-start 0 (declared)" "$(coverage_report)"
run_verdict
assert_zero_rc "case 2b: a declared zero with no reason exits 0" "$RC"

# ---------------------------------------------------------------------------
# 3. A member at zero and undeclared is reported by name and fails.
# ---------------------------------------------------------------------------
coverage_reset
coverage_record "myflow-fast" 0
assert_eq "case 3: an undeclared zero still renders as a plain count" "myflow-fast 0" "$(coverage_report)"
run_verdict
assert_contains "case 3: an undeclared zero is named in the verdict" "$VERDICT" "myflow-fast"
assert_nonzero_rc "case 3: an undeclared zero exits non-zero" "$RC"

# ---------------------------------------------------------------------------
# 4. An empty corpus is itself a violation, not a vacuous pass.
# ---------------------------------------------------------------------------
coverage_reset
assert_eq "case 4: an empty corpus renders nothing" "" "$(coverage_report)"
run_verdict
if [ -n "$VERDICT" ]; then
  pass "case 4: an empty corpus produces a violation line"
else
  fail "case 4: an empty corpus produced no output — a vacuous pass"
fi
assert_nonzero_rc "case 4: an empty corpus exits non-zero" "$RC"

# ---------------------------------------------------------------------------
# 5. A member name carrying a space or a leading `-` is handled, not
#    mis-parsed — as a rendered count, as an undeclared-zero violation, and
#    as a declared zero that silences that violation.
# ---------------------------------------------------------------------------
ODD_NAME="-weird flag-like name"

coverage_reset
coverage_record "$ODD_NAME" 3
assert_eq "case 5a: a leading-dash, spaced member name renders intact" "$ODD_NAME 3" "$(coverage_report)"

coverage_reset
coverage_record "$ODD_NAME" 0
run_verdict
assert_contains "case 5b: the same odd name is named correctly as a violation" "$VERDICT" "$ODD_NAME"
assert_nonzero_rc "case 5b: an odd-named undeclared zero exits non-zero" "$RC"

coverage_reset
coverage_record "$ODD_NAME" 0
coverage_declare "$ODD_NAME" "test fixture"
run_verdict
assert_eq "case 5c: declaring the same odd name silences the violation" "" "$VERDICT"
assert_zero_rc "case 5c: a declared odd-named zero exits 0" "$RC"

# ---------------------------------------------------------------------------
# 6. Robustness: the library's own argument validation fails fast rather
#    than guessing. Not one of the five named cases, but the same
#    reject-rather-than-guess posture every guard's grep calls hold to,
#    applied here to this library's own inputs.
# ---------------------------------------------------------------------------
coverage_reset
set +e
coverage_record "myflow-do" "not-a-number" 2>/dev/null
RC=$?
set -e
assert_nonzero_rc "case 6a: a non-numeric count is rejected" "$RC"

coverage_reset
set +e
coverage_record "" 3 2>/dev/null
RC=$?
set -e
assert_nonzero_rc "case 6b: an empty member name is rejected" "$RC"

coverage_reset
coverage_record "myflow-do" 5
set +e
coverage_record "myflow-do" 9 2>/dev/null
RC=$?
set -e
assert_nonzero_rc "case 6c: recording the same member twice is rejected" "$RC"

coverage_reset
coverage_declare "flow-status"
set +e
coverage_declare "flow-status" "second reason" 2>/dev/null
RC=$?
set -e
assert_nonzero_rc "case 6d: declaring the same member twice is rejected" "$RC"

# ---------------------------------------------------------------------------
# 7 (KAN-197 F3): a coverage_declare for a member that is never
# coverage_record'd at all is silently inert on today's library — recorded
# here as the exact reproducer the panel used: record one real member,
# declare a ghost that is never recorded, and confirm coverage_verdict now
# complains instead of returning 0. The declaration list can otherwise only
# grow, never self-prune, so a renamed or deleted corpus member leaves a
# stale entry nobody notices.
# ---------------------------------------------------------------------------
coverage_reset
coverage_record "real" 5
coverage_declare "ghost-that-does-not-exist" "stale"
run_verdict
assert_contains "case 7: a declared-but-never-recorded member is named in the verdict" "$VERDICT" "ghost-that-does-not-exist"
assert_nonzero_rc "case 7: a declared-but-never-recorded member exits non-zero" "$RC"

# ---------------------------------------------------------------------------
# 8: a declared-zero member whose recorded count turns out non-zero is a
# violation, not a silent pass — the panel's F9 reproducer (planted a real
# call into a declared-zero router; the guard kept exiting 0). The verdict
# line names the member, its actual count and its now-false reason so the
# stale declaration is loud instead of just quietly dropped from the report.
# ---------------------------------------------------------------------------
coverage_reset
coverage_record "flow.SKILL" 1
coverage_declare "flow.SKILL" "a legitimate zero-mark router"
run_verdict
assert_contains "case 8: a declared member gone non-zero is named in the verdict" "$VERDICT" "flow.SKILL"
assert_contains "case 8: the verdict carries the actual count" "$VERDICT" "1"
assert_contains "case 8: the verdict carries the now-false reason" "$VERDICT" "a legitimate zero-mark router"
assert_nonzero_rc "case 8: a declared member gone non-zero exits non-zero" "$RC"

# ---------------------------------------------------------------------------
if [ "$FAILURES" -eq 0 ]; then
  printf '\n✓ PASS\n'
  exit 0
fi
printf '\n✗ FAIL — %s failure(s)\n' "$FAILURES" >&2
exit 1
