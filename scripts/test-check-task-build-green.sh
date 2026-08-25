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
  printf -- '- [ ] 1.1. First task\n\n'
  printf '**Build:** green\n\n'
  printf -- '- [ ] 1.2. Second task\n\n'
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
  printf -- '- [ ] 2.1. Untagged task\n\n'
  printf 'Some body text, no build tag at all.\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 2: missing tag fails" || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"task 2.1 has no **Build:** tag"*) pass "case 2: names the task id" ;;
  *) fail "case 2: expected message naming task 2.1, out=$OUT" ;;
esac

# ===========================================================================
# Case 3: a task tagged red with a **Squash-with:** field present but naming
# zero ids -> exit 1. The field is present syntactically ("Squash-with:
# Task") but names zero ids, which is what makes this the "no partner named"
# violation rather than the "no Squash-with: field" one (a red task with no
# Squash-with: line at all is a separate violation, case 16).
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 3.1. Red with no partner\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task ,\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 3: red with no partner fails" || fail "case 3: rc=$RC out=$OUT"
case "$OUT" in
  *"task 3.1 is red with no merge partner named"*) pass "case 3: names the task id" ;;
  *) fail "case 3: expected message naming task 3.1, out=$OUT" ;;
esac

# ===========================================================================
# Case 4: Squash-with: Task 9.9, where no task 9.9 exists -> exit 1.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 4.1. Red with nonexistent partner\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 9.9\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 4: nonexistent partner fails" || fail "case 4: rc=$RC out=$OUT"
case "$OUT" in
  *"task 4.1 merges with Task 9.9, which does not exist in this plan"*) pass "case 4: names the missing partner" ;;
  *) fail "case 4: expected message naming missing partner 9.9, out=$OUT" ;;
esac

# ===========================================================================
# Case 5: Squash-with: Task 2, where task 2 is itself tagged red -> exit 1,
# reported against the FIRST task (the one carrying the offending field),
# not the partner.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. First task\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 2\n\n'
  printf -- '- [ ] 2. Second task\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 1\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 5: partner itself red fails" || fail "case 5: rc=$RC out=$OUT"
case "$OUT" in
  *"task 1 merges with Task 2, which is itself red"*) pass "case 5: reported against task 1" ;;
  *) fail "case 5: expected message reported against task 1, out=$OUT" ;;
esac

# ===========================================================================
# Case 6: Squash-with: Task 2, where task 2 is tagged green -> exit 0.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. First task\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 2\n\n'
  printf -- '- [ ] 2. Second task\n\n'
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
  printf -- '- [ ] 7.1. Malformed tag\n\n'
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
  printf -- '- [ ] 8.1. Untagged real task\n\n'
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
# Case 9: a tagged real task whose body ALSO contains a fenced example task
# line + tag (```markdown - [ ] 9.9. Example / **Build:** red / **Squash-
# with:** Task 9.9```) must report clean -- the fenced task line must not
# spawn a phantom task, and must not swallow or corrupt the real task's own
# tag (F2, false positive / body corruption).
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 9.1. Tagged real task\n\n'
  printf 'Example of the wrong way:\n\n'
  printf '```markdown\n'
  printf -- '- [ ] 9.9. Example\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 9.9\n'
  printf '```\n\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 9: fenced task line+tag does not corrupt real task" || fail "case 9: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 9: no output" || fail "case 9: expected no output, got: $OUT"

# ===========================================================================
# Case 10: two tasks sharing the same id must be reported as a duplicate-id
# violation, not silently collapsed by `by_id` lookups (F3).
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 10.1. First\n'
  printf '**Build:** green\n'
  printf -- '- [ ] 10.1. Duplicate id\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 10: duplicate id fails" || fail "case 10: rc=$RC out=$OUT"
case "$OUT" in
  *"task 10.1 is defined more than once (first at line 1)"*) pass "case 10: names the duplicate and first line" ;;
  *) fail "case 10: expected duplicate-id message, out=$OUT" ;;
esac

# ===========================================================================
# Case 11: a task line with an id but no title text after it (`- [ ] 11.1. `
# at end of line) must still be recognised as a task -- proven here by giving
# it no tag at all, so the only way this case can fail is if the task was
# never parsed in the first place (F4). The space after the dot is required,
# here and in spectre both: `- [ ] 11.1.` with nothing after the dot is a
# "malformed task line" finding there and no task at all here.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 11.1. \n'
  printf 'Some body text, no tag.\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 11: titleless task line recognised as a task" || fail "case 11: rc=$RC out=$OUT"
case "$OUT" in
  *"task 11.1 has no **Build:** tag"*) pass "case 11: reports the missing tag for 11.1" ;;
  *) fail "case 11: expected missing-tag message naming 11.1, out=$OUT" ;;
esac

# ===========================================================================
# Case 12: a partner id named twice in one Squash-with: field
# ("Squash-with: Task 12.2, 12.2") must produce exactly ONE violation line
# for that pair, not one per occurrence (F10).
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 12.1. First\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 12.2, 12.2\n\n'
  printf -- '- [ ] 12.2. Second\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 12.1\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 12: duplicate partner id fails" || fail "case 12: rc=$RC out=$OUT"
OCCURRENCES="$(printf '%s\n' "$OUT" | grep -c "task 12.1 merges with Task 12.2, which is itself red" || true)"
[ "$OCCURRENCES" -eq 1 ] && pass "case 12: exactly one violation line for the duplicated partner" \
  || fail "case 12: expected exactly one violation line, got $OCCURRENCES: $OUT"

# ===========================================================================
# Case 13: the wrapper's no-argument scan path (F5), exercised as a smoke
# test against this repository's OWN spectre/changes tree. check-task-
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
mkdir -p "$AGG_ROOT/spectre/changes/change-a" \
  "$AGG_ROOT/spectre/changes/change-b" \
  "$AGG_ROOT/spectre/changes/archive/2024-01-01-old-change"
{
  printf -- '- [ ] 1.1. Clean task\n\n'
  printf '**Build:** green\n'
} > "$AGG_ROOT/spectre/changes/change-a/tasks.md"
{
  printf -- '- [ ] 1.1. Untagged task\n\n'
  printf 'No tag here.\n'
} > "$AGG_ROOT/spectre/changes/change-b/tasks.md"
{
  printf -- '- [ ] 1.1. Archived, also untagged\n\n'
  printf 'No tag here either -- must never be scanned.\n'
} > "$AGG_ROOT/spectre/changes/archive/2024-01-01-old-change/tasks.md"

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

# ===========================================================================
# Case 15: a red task's partner is read from a separate **Squash-with:**
# field, not from any inline suffix on the **Build:** line -> exit 0 when
# the named partner is green.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 15.1. Red task\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 15.2\n\n'
  printf -- '- [ ] 15.2. Green partner\n\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 15: partner read from Squash-with: passes" || fail "case 15: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 15: no output" || fail "case 15: expected no output, got: $OUT"

# ===========================================================================
# Case 16: a task tagged **Build:** red with no **Squash-with:** field at
# all -> exit 1, with a message distinct from the "no merge partner named"
# violation (case 3), since it is checking for the absence of an entire
# field rather than an empty one.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 16.1. Red with no Squash-with field\n\n'
  printf '**Build:** red\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 16: red with no Squash-with: field fails" || fail "case 16: rc=$RC out=$OUT"
case "$OUT" in
  *"task 16.1 is red with no **Squash-with:** field"*) pass "case 16: names the missing-field violation" ;;
  *) fail "case 16: expected missing-Squash-with-field message, out=$OUT" ;;
esac

# ===========================================================================
# Case 17: a **Squash-with:** partner must itself be tagged green -> a
# partner tagged red fails (mirrors case 5, restated here under the Tests:
# field's own case numbering).
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 17.1. First task\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 17.2\n\n'
  printf -- '- [ ] 17.2. Second task\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 17.1\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 17: Squash-with: partner must be green fails" || fail "case 17: rc=$RC out=$OUT"
case "$OUT" in
  *"task 17.1 merges with Task 17.2, which is itself red"*) pass "case 17: reported against task 17.1" ;;
  *) fail "case 17: expected message reported against task 17.1, out=$OUT" ;;
esac

# ===========================================================================
# Case 18 (fix round 6, F13): the **Squash-with:** grammar this guard reads
# now lives in scripts/lib/plan_grammar.py, shared with
# check-task-commit-fields.py, which used to keep its own copy of it. The
# two behaviours that split apart there are pinned here against THIS guard,
# so the shared module cannot be changed for one guard without the other
# noticing:
#
#   * a value that does not gate as `Task <ids>` is not a Squash-with field
#     at all — a red task carrying one is reported as having none, exactly
#     as before the grammar moved (case 16's message, reached by a different
#     route);
#   * the field is LINE-scoped — a prose line following it, with no blank
#     line between, is not part of its value, so the partner still resolves.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 18.1. Red task whose Squash-with value carries free text\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 18.2 (see step 3)\n\n'
  printf -- '- [ ] 18.2. Green partner\n\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 18: a Squash-with value that does not gate fails" || fail "case 18: rc=$RC out=$OUT"
case "$OUT" in
  *"task 18.1 is red with no **Squash-with:** field"*) pass "case 18: an ungated value is reported as no field at all" ;;
  *) fail "case 18: expected the missing-field message, out=$OUT" ;;
esac

new_fixture
{
  printf -- '- [ ] 18.3. Red task whose Squash-with is followed by prose\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 18.4\n'
  printf 'The fold is described in the paragraph above.\n\n'
  printf -- '- [ ] 18.4. Green partner\n\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 18: the field is line-scoped, so a following prose line is not part of its value" || fail "case 18 (line scope): rc=$RC out=$OUT"

# ===========================================================================
# Case 19 (fix round 8, F19): WHICH line is the **Squash-with:** field, when
# a body carries a non-gating one ahead of a gating one. Field selection now
# lives in lib/plan_grammar.py's select_squash_with, which both guards call
# instead of each looping over the body itself; this case pins that the
# shared selection still skips the non-gating candidate and resolves the
# partner named by the gating one. The same fixture is asserted against
# check-task-commit-fields.sh as its own case 57, which used to take the
# first field-SHAPED line and reach the opposite verdict on this very body.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 19.1. Red task whose first Squash-with line does not gate\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 19.2 (see note below)\n'
  printf '**Squash-with:** Task 19.2\n\n'
  printf -- '- [ ] 19.2. Green partner named by the gating line\n\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 19: the gating line is the field, so the partner resolves" || fail "case 19: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 19: no output" || fail "case 19: expected no output, got: $OUT"

# ===========================================================================
# Case 20 (fix round 8, F18): this wrapper's OWN missing-grammar check, at
# runtime — the check its sibling check-task-commit-fields.sh has carried
# since fix round 6 and this one did not, though the Python guard underneath
# gained the same lib/plan_grammar.py import in the same round. Mirrors that
# harness's case 56: a throwaway copy of the guard pair with no lib/ sibling,
# run against a fixture the shipped guard passes, so both assertions
# discriminate — without the check the copy reaches python3, whose
# `from plan_grammar import ...` raises and exits 1 with a traceback rather
# than 2 with a sentence.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 20.1. Clean task\n\n'
  printf '**Build:** green\n'
} > "$TASKS_MD"
# Resolved through cd/pwd so the path matches the one the wrapper prints,
# which it derives the same way — on macOS $TMPDIR is itself a symlink.
STRIPPED="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/task-build-green-test.XXXXXX")" && pwd)"
cp "$SCRIPT_DIR/check-task-build-green.sh" "$SCRIPT_DIR/check-task-build-green.py" "$STRIPPED/"
set +e
STRIPPED_OUT="$("$STRIPPED/check-task-build-green.sh" "$TASKS_MD" 2>&1)"
STRIPPED_RC=$?
set -e
[ "$STRIPPED_RC" -eq 2 ] && pass "case 20: a guard copy with no lib/ sibling exits 2" || fail "case 20: rc=$STRIPPED_RC out=$STRIPPED_OUT"
case "$STRIPPED_OUT" in
  *"shared grammar module not found: $STRIPPED/lib/plan_grammar.py"*) pass "case 20: the message names the missing module path" ;;
  *) fail "case 20: expected the missing-module path in the message, out=$STRIPPED_OUT" ;;
esac
rm -rf "$STRIPPED"

# ===========================================================================
# Case 21 (fix round 9, F20): WHICH line is the **Build:** tag, when a body
# carries two. Tag selection now lives in lib/plan_grammar.py's
# select_build_tag, which both guards call: the FIRST line-gated tag wins,
# so this task is red and its missing partner is reported. Read
# last-line-wins it would be green and this plan would pass. The same body
# is asserted against check-task-commit-fields.sh as its own case 58, where
# the last-wins reading was what that guard actually did.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 21.1. Red task carrying two Build lines\n\n'
  printf '**Build:** red\n'
  printf '**Build:** green\n\n'
  printf '**Squash-with:** Task 21.9\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 21: the first Build line is the tag, so the task is red" || fail "case 21: rc=$RC out=$OUT"
case "$OUT" in
  *"task 21.1 merges with Task 21.9, which does not exist in this plan"*) pass "case 21: the red task's missing partner is reported" ;;
  *) fail "case 21: expected the missing-partner message, out=$OUT" ;;
esac

# ===========================================================================
# Case 22 (fix round 9, F20): the **Build:** tag is LINE-SCOPED. A prose
# line following the tag with no blank line between is not part of it, so
# the task is still red. The same body is asserted against
# check-task-commit-fields.sh as its own case 59, where that line used to be
# joined onto the tag's value and left the task with no tag at all.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 22.1. Red task whose tag line is followed by prose\n\n'
  printf '**Build:** red\n'
  printf 'this sentence explains the tag and is not part of it\n\n'
  printf '**Squash-with:** Task 22.9\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 22: a prose line under the tag does not unset it" || fail "case 22: rc=$RC out=$OUT"
case "$OUT" in
  *"task 22.1 merges with Task 22.9, which does not exist in this plan"*) pass "case 22: the red task's missing partner is reported" ;;
  *) fail "case 22: expected the missing-partner message, out=$OUT" ;;
esac

# ===========================================================================
# Case 23 (fix round 9): a fenced example task line opens no task, so the
# defective fold shown inside it is documentation and not a violation.
# Case 9 pins the same rule for this guard's own body parsing; this body is
# the one asserted against check-task-commit-fields.sh as its own case 60,
# where the fenced task line used to open a real task whose ungated
# `Squash-with:` failed every task in the plan.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 23.1. Real task with a worked example in its body\n\n'
  printf '**Build:** green\n\n'
  printf 'Example of a fold, shown but never declared:\n\n'
  printf '```\n'
  printf -- '- [ ] 23.9. Example red task\n'
  printf '**Build:** red\n'
  printf '**Squash-with:** Task 23.8 (see the note)\n'
  printf '```\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 23: a fenced example task line opens no task" || fail "case 23: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 23: no output" || fail "case 23: expected no output, got: $OUT"

# ===========================================================================
# Case 24 (fix round 9): WHICH task a duplicated id names. Task 24.3's
# partner lookup resolves against the FIRST task line carrying id 24.1 — the
# green one — so no "itself red" violation is reported for it, however many
# later task lines reuse that id. The duplicate itself is still reported (case
# 10's rule). The same body is asserted against
# check-task-commit-fields.sh as its own case 61, where the id used to
# resolve to the LAST task line instead.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 24.1. First task line for this id\n\n'
  printf '**Build:** green\n\n'
  printf -- '- [ ] 24.1. Second task line reusing the id\n\n'
  printf '**Build:** red\n\n'
  printf -- '- [ ] 24.3. Red task folding into Task 24.1\n\n'
  printf '**Build:** red\n'
  printf '**Squash-with:** Task 24.1\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 24: the duplicate id is still reported" || fail "case 24: rc=$RC out=$OUT"
case "$OUT" in
  *"task 24.1 is defined more than once (first at line 1)"*) pass "case 24: names the duplicate and first line" ;;
  *) fail "case 24: expected duplicate-id message, out=$OUT" ;;
esac
case "$OUT" in
  *"task 24.3 merges with Task 24.1, which is itself red"*) fail "case 24: the id resolved to the last task line, not the first, out=$OUT" ;;
  *) pass "case 24: a duplicated id resolves to its first task line" ;;
esac

# ===========================================================================
# Case 25 (fix round 11, F22): a fence a task body OPENS and never CLOSES is
# a violation of its own, reported INSTEAD of the tag and partner checks it
# blinded. An unclosed fence runs to the end of the block, so the
# `**Build:**` tag below it is code and is correctly not read — but calling
# the task untagged names the consequence and hides the cause. The detection
# is lib/plan_grammar.py's `unclosed_fence`, and the same body is asserted
# against check-task-commit-fields.sh as its own case 64, where the
# swallowed `**Files:**` made the commit be blamed for touching the file the
# task really did declare.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Task whose body opens a fence it never closes\n\n'
  printf '~~~\n\n'
  printf '**Build:** green\n'
  printf '**Files:** `alpha.txt`\n'
  printf '**Commit:** test: add alpha\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 25: an unclosed fence in a task body fails" || fail "case 25: rc=$RC out=$OUT"
case "$OUT" in
  *":3: task 1 opens a code fence here that is never closed"*) pass "case 25: names the task and the line the fence opened on" ;;
  *) fail "case 25: expected the unclosed-fence violation at line 3, out=$OUT" ;;
esac
case "$OUT" in
  *"has no **Build:** tag"*) fail "case 25: the swallowed tag is still reported as the defect, out=$OUT" ;;
  *) pass "case 25: the untagged-task noise is replaced, not accompanied" ;;
esac

# The no-false-positive half: a fence the body CLOSES is not an unclosed one,
# and the tag above it is read exactly as before.
new_fixture
{
  printf -- '- [ ] 1. Task whose body closes the fence it opens\n\n'
  printf '**Build:** green\n'
  printf '**Files:** `alpha.txt`\n'
  printf '**Commit:** test: add alpha\n\n'
  printf '~~~\n'
  printf 'example\n'
  printf '~~~\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 25: a closed fence is not reported as unclosed" || fail "case 25: rc=$RC out=$OUT"

# ===========================================================================
# Case 26: a task IS the spectre checkbox line `- [ ] <id>. <title>`, and a
# step checkbox beneath it is NOT a task. This is the distinction spectre's
# own parser makes -- it reads `- [ ] <n>. <text>` and nothing else -- and
# the whole reason the task shape moved off the `###` heading: while the two
# grammars disagreed, `spectre list` reported 0/0 for a real plan, `spectre
# validate` reported one false "malformed task line" per step checkbox, and
# `spectre archive` refused with "tasks.md has no tasks".
#
# 26a proves the checkbox line opens a task: the task carries no tag at all,
# so the only way this can pass is if the line was parsed as a task in the
# first place.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 26.1. Untagged task\n\n'
  printf 'Some body text, no build tag at all.\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 26a: a checkbox line opens a task" || fail "case 26a: rc=$RC out=$OUT"
case "$OUT" in
  *"task 26.1 has no **Build:** tag"*) pass "case 26a: names the task id read off the checkbox line" ;;
  *) fail "case 26a: expected missing-tag message naming 26.1, out=$OUT" ;;
esac

# 26b: a `[x]` checkbox is a task too -- the mark carries whether the task is
# DONE, never whether it is a task. An untagged done task is still a
# violation here, exactly as an untagged open one is.
new_fixture
{
  printf -- '- [x] 26.2. Untagged done task\n\n'
  printf 'Some body text, no build tag at all.\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 26b: a checked task line is still a task" || fail "case 26b: rc=$RC out=$OUT"
case "$OUT" in
  *"task 26.2 has no **Build:** tag"*) pass "case 26b: names the checked task's id" ;;
  *) fail "case 26b: expected missing-tag message naming 26.2, out=$OUT" ;;
esac

# 26c: the step checkboxes beneath a task are not tasks. The task itself is
# tagged, so any step read as a task would surface as a second, untagged one.
new_fixture
{
  printf -- '- [ ] 26.3. Tagged task with steps\n\n'
  printf '**Build:** green\n\n'
  printf -- '- [ ] **Step 1: a step, never a task**\n'
  printf -- '- [x] **Step 2: a done step, still never a task**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 26c: step checkboxes are not tasks" || fail "case 26c: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 26c: no output" || fail "case 26c: expected no output, got: $OUT"

# 26d: a `### <id> <title>` heading is NOT a task any more. Accepting both
# shapes would leave this guard and spectre disagreeing exactly as they did
# before, so the heading is inert here. The real task comes FIRST and carries
# the only tag in the file, and the heading below it closes that task's body:
# a grammar that still read the heading as a task would find it untagged and
# fail.
new_fixture
{
  printf -- '- [ ] 26.5. The only real task\n\n'
  printf '**Build:** green\n\n'
  printf '### 26.4 A heading, not a task\n\n'
  printf 'No tag here.\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 26d: a level-3 heading opens no task" || fail "case 26d: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 26d: no output" || fail "case 26d: expected no output, got: $OUT"

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
