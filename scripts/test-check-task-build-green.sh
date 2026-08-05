#!/usr/bin/env bash
# Assertion harness for check-task-build-green.sh. Builds fixture tasks.md
# files under a sandboxed TMPDIR and asserts the guard's exit status and,
# where the plan below names one, the presence of the expected violation
# message text. Never touches the real repository tree.
#
# Modeled on test-check-plan-provenance.sh's fixture-driven pattern: fixtures
# live under mktemp -d, the guard is invoked via a thin run_guard helper that
# captures RC/OUT, and every case ends with an explicit pass/fail assertion —
# never a bare "it didn't crash".
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-task-build-green.sh"
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

# run_guard_root <root> -> sets RC and OUT; invokes the guard with no
# arguments (its aggregation/no-arg scan mode), pointed at <root> via
# CHECK_TASK_BUILD_GREEN_ROOT rather than a real cwd, the same pattern
# test-check-plan-provenance.sh uses via CHECK_PLAN_PROVENANCE_ROOT.
run_guard_root() {
  set +e
  OUT="$(CHECK_TASK_BUILD_GREEN_ROOT="$1" "$GUARD" 2>&1)"
  RC=$?
  set -e
}

new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/task-build-green-test.XXXXXX")"
  TASKS_MD="$FIXTURE/tasks.md"
}

# ===========================================================================
# Case 1: all tasks tagged green -> exit 0, no output.
# ===========================================================================
new_fixture
{
  printf '## 1 Group\n\n'
  printf '### 1.1 First task\n\n'
  printf '**Build:** green\n\n'
  printf '### 1.2 Second task\n\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 1: all green passes" || fail "case 1: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 1: no output" || fail "case 1: expected no output, got: $OUT"

# ===========================================================================
# Case 2: one task with no **Build:** line -> exit 1, names that task's id.
# ===========================================================================
new_fixture
{
  printf '### 2.1 Untagged task\n\n'
  printf 'Some body text, no build tag at all.\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 2: missing tag fails" || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"task 2.1 has no **Build:** tag"*) pass "case 2: names the task id" ;;
  *) fail "case 2: expected message naming task 2.1, out=$OUT" ;;
esac

# ===========================================================================
# Case 3: a task tagged red with no "merges with Task ..." clause -> exit 1.
# The clause is present syntactically ("merges with Task") but names zero
# ids, which is what makes this the "no partner named" violation rather than
# the "missing tag" one (a bare "**Build:** red" with no clause at all does
# not match the tag grammar and falls to case 2's violation instead).
# ===========================================================================
new_fixture
{
  printf '### 3.1 Red with no partner\n\n'
  printf '**Build:** red — merges with Task ,\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 3: red with no partner fails" || fail "case 3: rc=$RC out=$OUT"
case "$OUT" in
  *"task 3.1 is red with no merge partner named"*) pass "case 3: names the task id" ;;
  *) fail "case 3: expected message naming task 3.1, out=$OUT" ;;
esac

# ===========================================================================
# Case 4: red — merges with Task 9.9, where no task 9.9 exists -> exit 1.
# ===========================================================================
new_fixture
{
  printf '### 4.1 Red with nonexistent partner\n\n'
  printf '**Build:** red — merges with Task 9.9\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 4: nonexistent partner fails" || fail "case 4: rc=$RC out=$OUT"
case "$OUT" in
  *"task 4.1 merges with Task 9.9, which does not exist in this plan"*) pass "case 4: names the missing partner" ;;
  *) fail "case 4: expected message naming missing partner 9.9, out=$OUT" ;;
esac

# ===========================================================================
# Case 5: red — merges with Task 2, where task 2 is itself tagged red ->
# exit 1, reported against the FIRST task (the one carrying the offending
# tag), not the partner.
# ===========================================================================
new_fixture
{
  printf '### 1 First task\n\n'
  printf '**Build:** red — merges with Task 2\n\n'
  printf '### 2 Second task\n\n'
  printf '**Build:** red — merges with Task 1\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 5: partner itself red fails" || fail "case 5: rc=$RC out=$OUT"
case "$OUT" in
  *"task 1 merges with Task 2, which is itself red"*) pass "case 5: reported against task 1" ;;
  *) fail "case 5: expected message reported against task 1, out=$OUT" ;;
esac

# ===========================================================================
# Case 6: red — merges with Task 2, where task 2 is tagged green -> exit 0.
# ===========================================================================
new_fixture
{
  printf '### 1 First task\n\n'
  printf '**Build:** red — merges with Task 2\n\n'
  printf '### 2 Second task\n\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 6: partner green passes" || fail "case 6: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 6: no output" || fail "case 6: expected no output, got: $OUT"

# ===========================================================================
# Case 7: a malformed tag line (e.g. "**Build:** yellow") is treated as NO
# tag -- falls through to case 2's violation, not a separate parse error.
# ===========================================================================
new_fixture
{
  printf '### 7.1 Malformed tag\n\n'
  printf '**Build:** yellow\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 7: malformed tag fails" || fail "case 7: rc=$RC out=$OUT"
case "$OUT" in
  *"task 7.1 has no **Build:** tag"*) pass "case 7: reported as missing tag" ;;
  *) fail "case 7: expected missing-tag message, out=$OUT" ;;
esac

# ===========================================================================
# Case 8: an untagged real task whose body ALSO contains a fenced example
# line that looks like a tag (```markdown fenced **Build:** green```) must
# still report the missing-tag violation -- the fenced line is example text,
# never a real tag (F2, false negative).
# ===========================================================================
new_fixture
{
  printf '### 8.1 Untagged real task\n\n'
  printf 'Example of the wrong way:\n\n'
  printf '```markdown\n'
  printf '**Build:** green\n'
  printf '```\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 8: fenced example tag does not count as real" || fail "case 8: rc=$RC out=$OUT"
case "$OUT" in
  *"task 8.1 has no **Build:** tag"*) pass "case 8: still reports missing tag" ;;
  *) fail "case 8: expected missing-tag message, out=$OUT" ;;
esac

# ===========================================================================
# Case 9: a tagged real task whose body ALSO contains a fenced example
# heading + tag (```markdown ### 9.9 Example / **Build:** red — merges with
# Task 9.9```) must report clean -- the fenced heading must not spawn a
# phantom task, and must not swallow or corrupt the real task's own tag
# (F2, false positive / body corruption).
# ===========================================================================
new_fixture
{
  printf '### 9.1 Tagged real task\n\n'
  printf 'Example of the wrong way:\n\n'
  printf '```markdown\n'
  printf '### 9.9 Example\n'
  printf '**Build:** red — merges with Task 9.9\n'
  printf '```\n\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 9: fenced heading+tag does not corrupt real task" || fail "case 9: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 9: no output" || fail "case 9: expected no output, got: $OUT"

# ===========================================================================
# Case 10: two tasks sharing the same id must be reported as a duplicate-id
# violation, not silently collapsed by `by_id` lookups (F3).
# ===========================================================================
new_fixture
{
  printf '### 10.1 First\n'
  printf '**Build:** green\n'
  printf '### 10.1 Duplicate id\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 10: duplicate id fails" || fail "case 10: rc=$RC out=$OUT"
case "$OUT" in
  *"task 10.1 is defined more than once (first at line 1)"*) pass "case 10: names the duplicate and first line" ;;
  *) fail "case 10: expected duplicate-id message, out=$OUT" ;;
esac

# ===========================================================================
# Case 11: a heading with an id but no trailing title text (`### 11.1` at
# end of line) must still be recognised as a task -- proven here by giving
# it no tag at all, so the only way this case can fail is if the task was
# never parsed in the first place (F4).
# ===========================================================================
new_fixture
{
  printf '### 11.1\n'
  printf 'Some body text, no tag.\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 11: titleless heading recognised as a task" || fail "case 11: rc=$RC out=$OUT"
case "$OUT" in
  *"task 11.1 has no **Build:** tag"*) pass "case 11: reports the missing tag for 11.1" ;;
  *) fail "case 11: expected missing-tag message naming 11.1, out=$OUT" ;;
esac

# ===========================================================================
# Case 12: a partner id named twice in one clause ("merges with Task 12.2,
# 12.2") must produce exactly ONE violation line for that pair, not one per
# occurrence (F10).
# ===========================================================================
new_fixture
{
  printf '### 12.1 First\n\n'
  printf '**Build:** red — merges with Task 12.2, 12.2\n\n'
  printf '### 12.2 Second\n\n'
  printf '**Build:** red — merges with Task 12.1\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 12: duplicate partner id fails" || fail "case 12: rc=$RC out=$OUT"
OCCURRENCES="$(printf '%s\n' "$OUT" | grep -c "task 12.1 merges with Task 12.2, which is itself red" || true)"
[ "$OCCURRENCES" -eq 1 ] && pass "case 12: exactly one violation line for the duplicated partner" \
  || fail "case 12: expected exactly one violation line, got $OCCURRENCES: $OUT"

# ===========================================================================
# Case 13: the wrapper's no-argument scan path (F5), exercised as a smoke
# test against this repository's OWN openspec/changes tree. check-task-
# build-green.sh derives REPO_ROOT from its own script location (not the
# invocation cwd) unless CHECK_TASK_BUILD_GREEN_ROOT overrides it, so the
# no-arg default path can only be exercised for real against the real repo
# tree -- which at the time this suite runs contains only archive/ plus this
# change's own tasks.md, all green.
# ===========================================================================
set +e
OUT="$("$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "case 13: no-arg scan of the real repo tree passes" \
  || fail "case 13: rc=$RC out=$OUT"

# ===========================================================================
# Case 14: the wrapper's no-argument aggregation loop (F5), exercised
# against a sandboxed fixture tree via CHECK_TASK_BUILD_GREEN_ROOT (the
# override this fix adds, modeled on check-plan-provenance.sh's
# CHECK_PLAN_PROVENANCE_ROOT). Asserts: aggregation reports non-zero when
# ANY file has a violation, the violating file's own message is surfaced,
# and a tasks.md nested under archive/ (two levels deep, same as a real
# archived change) is excluded from the scan entirely.
# ===========================================================================
AGG_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/task-build-green-agg-test.XXXXXX")"
mkdir -p "$AGG_ROOT/openspec/changes/change-a" \
  "$AGG_ROOT/openspec/changes/change-b" \
  "$AGG_ROOT/openspec/changes/archive/2024-01-01-old-change"
{
  printf '### 1.1 Clean task\n\n'
  printf '**Build:** green\n'
} > "$AGG_ROOT/openspec/changes/change-a/tasks.md"
{
  printf '### 1.1 Untagged task\n\n'
  printf 'No tag here.\n'
} > "$AGG_ROOT/openspec/changes/change-b/tasks.md"
{
  printf '### 1.1 Archived, also untagged\n\n'
  printf 'No tag here either -- must never be scanned.\n'
} > "$AGG_ROOT/openspec/changes/archive/2024-01-01-old-change/tasks.md"

run_guard_root "$AGG_ROOT"
[ "$RC" -ne 0 ] && pass "case 14: aggregation reports non-zero when any file violates" \
  || fail "case 14: rc=$RC out=$OUT"
case "$OUT" in
  *"change-b/tasks.md"*"has no **Build:** tag"*) pass "case 14: surfaces change-b's own violation" ;;
  *) fail "case 14: expected change-b's violation in output, out=$OUT" ;;
esac
case "$OUT" in
  *"archive/"*) fail "case 14: archived tasks.md must be excluded from the scan, out=$OUT" ;;
  *) pass "case 14: archived tasks.md excluded from the scan" ;;
esac

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
