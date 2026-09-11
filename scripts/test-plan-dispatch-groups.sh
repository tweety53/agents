#!/usr/bin/env bash
# Assertion harness for plan-dispatch-groups.sh. Builds a fixture tasks.md
# under a sandboxed TMPDIR for every case; never touches this repository's
# own spectre/changes/ tree.
#
# Modeled on test-plan-dispatch-bundles.sh's fixture-driven pattern: a
# fixture directory per case via new_fixture, the guard invoked through a
# thin run_guard helper that captures RC/OUT, and every case ending with an
# explicit pass/fail assertion.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/plan-dispatch-groups.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# run_guard <tasks.md-path> -> sets RC and OUT
run_guard() {
  set +e
  OUT="$("$GUARD" "$1" 2>&1)"
  RC=$?
  set -e
}

FIXTURES=()
cleanup() {
  [ "${#FIXTURES[@]}" -eq 0 ] && return 0
  for fixture in "${FIXTURES[@]}"; do
    rm -rf "$fixture"
  done
}
trap cleanup EXIT

new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/plan-dispatch-groups-test.XXXXXX")"
  FIXTURES+=("$FIXTURE")
  TASKS_MD="$FIXTURE/tasks.md"
}

task() {
  # task <id> <files-path> [after-value] — one unchecked task, one
  # **Files:** entry, an optional **After:** field.
  local id="$1" files="$2" after="${3:-}"
  printf -- '- [ ] %s. Task %s\n\n' "$id" "$id"
  printf '**Files:**\n- Create: `%s`\n\n' "$files"
  if [ -n "$after" ]; then
    printf '**After:** %s\n\n' "$after"
  fi
  printf -- '  - [ ] **Step 1: do it**\n\n'
}

# ===========================================================================
# Case 1: serial-default plan, five disjoint-file tasks, no **After:**
# field on any of them (the serial default) — one group [1,2,3,4,5].
# ===========================================================================
new_fixture
{
  task 1 a.txt
  task 2 b.txt
  task 3 c.txt
  task 4 d.txt
  task 5 e.txt
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 1: serial-default exits 0" || fail "case 1: rc=$RC out=$OUT"
EXPECTED=$'group 1: 1 2 3 4 5'
[ "$OUT" = "$EXPECTED" ] && pass "case 1: one group for a fully serial plan" || fail "case 1: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 2: two independent chains (1->3, 2->4) plus a join point (5, after
# 3 and 4) that straddles both chains and opens its own group.
# ===========================================================================
new_fixture
{
  task 1 a.txt none
  task 2 b.txt none
  task 3 c.txt "Task 1"
  task 4 d.txt "Task 2"
  task 5 e.txt "Task 3, 4"
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 2: two chains plus a join point exits 0" || fail "case 2: rc=$RC out=$OUT"
EXPECTED=$'group 1: 1 3\ngroup 2: 2 4\ngroup 3: 5'
[ "$OUT" = "$EXPECTED" ] && pass "case 2: chains stay split, join point opens its own group" || fail "case 2: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 3: four after:none singletons then a fifth after all four — chain
# merge gives four singletons plus one group; the fold pass merges the four
# singletons (identical empty ready set) down to two.
# ===========================================================================
new_fixture
{
  task 1 a.txt none
  task 2 b.txt none
  task 3 c.txt none
  task 4 d.txt none
  task 5 e.txt "Task 1, 2, 3, 4"
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 3: four singletons plus a join exits 0" || fail "case 3: rc=$RC out=$OUT"
EXPECTED=$'group 1: 1 3\ngroup 2: 2 4\ngroup 3: 5'
[ "$OUT" = "$EXPECTED" ] && pass "case 3: fold to two, alternating, join point untouched" || fail "case 3: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 4: the design's known ceiling — 1, 2->1, 3->1, 4->(2,3) — all merge
# into one group even though 2 and 3 are file-disjoint and mutually ready.
# ===========================================================================
new_fixture
{
  task 1 a.txt
  task 2 b.txt "Task 1"
  task 3 c.txt "Task 1"
  task 4 d.txt "Task 2, 3"
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 4: the ceiling case exits 0" || fail "case 4: rc=$RC out=$OUT"
EXPECTED=$'group 1: 1 2 3 4'
[ "$OUT" = "$EXPECTED" ] && pass "case 4: one group, the documented ceiling" || fail "case 4: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 5: an unchecked task with no **Files:** field passes
# plan-dispatch-bundles.sh's own violation line through unchanged.
# ===========================================================================
new_fixture
{
  task 1 a.txt
  printf -- '- [ ] 2. Fieldless task\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 5: missing Files field exits 1" || fail "case 5: rc=$RC out=$OUT"
case "$OUT" in
  *"task 2 has no **Files:** field"*) pass "case 5: violation message passed through" ;;
  *) fail "case 5: expected a passed-through violation, got [$OUT]" ;;
esac

# ===========================================================================
# Case 6: zero unchecked tasks — every task already checked — prints
# nothing and exits 0.
# ===========================================================================
new_fixture
{
  printf -- '- [x] 1. Done task\n\n'
  printf '**Files:**\n- Modify: `shared.txt`\n\n'
  printf -- '  - [x] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "case 6: zero unchecked tasks prints nothing, exits 0" || fail "case 6: rc=$RC out=$OUT"

# ===========================================================================
# Case 7: wrong argument count exits 2.
# ===========================================================================
set +e
OUT="$("$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 7: no arguments exits 2" || fail "case 7: rc=$RC out=$OUT"

# ===========================================================================
if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
exit 0
