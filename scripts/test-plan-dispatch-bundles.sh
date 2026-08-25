#!/usr/bin/env bash
# Assertion harness for plan-dispatch-bundles.sh. Builds a fixture tasks.md
# under a sandboxed TMPDIR for every case; never touches this repository's
# own spectre/changes/ tree.
#
# Modeled on test-check-task-build-green.sh's fixture-driven pattern: a
# fixture directory per case via new_fixture, the guard invoked through a
# thin run_guard helper that captures RC/OUT, and every case ending with an
# explicit pass/fail assertion.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/plan-dispatch-bundles.sh"
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

# Every case leaves one fixture directory behind, removed on exit including
# on a failed assertion. An indexed array, not a space-separated string:
# mktemp paths under TMPDIR may contain spaces, and word-splitting a string
# would leak a fixture whose path split and `rm -rf` the fragments. Mirrors
# test-commit-split.sh's own REPOS array / trap cleanup EXIT pattern.
FIXTURES=()
cleanup() {
  [ "${#FIXTURES[@]}" -eq 0 ] && return 0
  for fixture in "${FIXTURES[@]}"; do
    rm -rf "$fixture"
  done
}
trap cleanup EXIT

new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/plan-dispatch-bundles-test.XXXXXX")"
  FIXTURES+=("$FIXTURE")
  TASKS_MD="$FIXTURE/tasks.md"
}

# ===========================================================================
# Case 1: three unchecked tasks declaring disjoint paths produce three
# bundles, one task per bundle, ordered by lowest task id.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. First task\n\n'
  printf '**Files:**\n- Create: `a.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 2. Second task\n\n'
  printf '**Files:**\n- Create: `b.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 3. Third task\n\n'
  printf '**Files:**\n- Create: `c.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 1: disjoint tasks exit 0" || fail "case 1: rc=$RC out=$OUT"
EXPECTED=$'bundle 1: 1\nbundle 2: 2\nbundle 3: 3'
[ "$OUT" = "$EXPECTED" ] && pass "case 1: three separate bundles" || fail "case 1: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 2: two tasks sharing one declared path form one bundle.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. First task\n\n'
  printf '**Files:**\n- Modify: `shared.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 2. Second task\n\n'
  printf '**Files:**\n- Modify: `shared.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 2: shared path exits 0" || fail "case 2: rc=$RC out=$OUT"
EXPECTED='bundle 1: 1 2'
[ "$OUT" = "$EXPECTED" ] && pass "case 2: one bundle for both tasks" || fail "case 2: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 3: transitivity — task 1 and task 2 share a path, task 2 and task 3
# share a different path; all three land in one bundle.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. First task\n\n'
  printf '**Files:**\n- Create: `x.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 2. Second task\n\n'
  printf '**Files:**\n- Create: `x.txt`\n- Create: `y.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 3. Third task\n\n'
  printf '**Files:**\n- Create: `y.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 3: transitive chain exits 0" || fail "case 3: rc=$RC out=$OUT"
EXPECTED='bundle 1: 1 2 3'
[ "$OUT" = "$EXPECTED" ] && pass "case 3: transitive closure forms one bundle" || fail "case 3: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 4: a task marked `- [x]` on its OWN task line takes no part in any
# bundle, even when it declares a path an unchecked task also declares —
# and the marks on its steps have nothing to say about it. Done-ness lives
# on the task line, which is the same bit `spectre list` counts; the step
# checkboxes beneath a task are body, not tasks. Both halves are pinned
# here against the step-counting rule this guard used to apply: the done
# task's step is still unmarked, and the open task's step is already
# marked, so a guard reading steps would produce exactly the opposite
# bundle.
# ===========================================================================
new_fixture
{
  printf -- '- [x] 1. Done task\n\n'
  printf '**Files:**\n- Modify: `shared.txt`\n\n'
  printf -- '  - [ ] **Step 1: a step left unmarked**\n\n'
  printf -- '- [ ] 2. Open task\n\n'
  printf '**Files:**\n- Modify: `shared.txt`\n\n'
  printf -- '  - [x] **Step 1: a step already marked**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 4: mixed checked/unchecked exits 0" || fail "case 4: rc=$RC out=$OUT"
EXPECTED='bundle 1: 2'
[ "$OUT" = "$EXPECTED" ] && pass "case 4: checked task excluded entirely" || fail "case 4: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 5: an unchecked task with no **Files:** field at all exits 1 and
# names that task.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. First task\n\n'
  printf '**Files:**\n- Create: `a.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 2. Fieldless task\n\n'
  printf 'No Files field here at all.\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 5: missing Files field exits 1" || fail "case 5: rc=$RC out=$OUT"
case "$OUT" in
  *"task 2"*"no **Files:** field"*) pass "case 5: names task 2" ;;
  *) fail "case 5: expected message naming task 2, out=$OUT" ;;
esac

# ===========================================================================
# Case 6: an **Allowed-collateral:** glob matching another task's declared
# path does not join the two bundles.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. First task\n\n'
  printf '**Files:**\n- Create: `a.txt`\n\n'
  printf '**Allowed-collateral:** `shared.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 2. Second task\n\n'
  printf '**Files:**\n- Create: `shared.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 6: collateral glob exits 0" || fail "case 6: rc=$RC out=$OUT"
EXPECTED=$'bundle 1: 1\nbundle 2: 2'
[ "$OUT" = "$EXPECTED" ] && pass "case 6: collateral does not join bundles" || fail "case 6: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 7: an unreadable path exits 2.
# ===========================================================================
new_fixture
run_guard "$FIXTURE/does-not-exist.md"
[ "$RC" -eq 2 ] && pass "case 7: unreadable path exits 2" || fail "case 7: rc=$RC out=$OUT"

# ===========================================================================
# Case 8: a bullet declaring two comma-separated backtick-quoted paths joins
# a bundle with a task that declares only the second of those paths. Guards
# against _extract_path grabbing only the first backtick token in a
# multi-path bullet (BACKTICK_RE.search instead of .findall).
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. First task\n\n'
  printf '**Files:**\n- Modify: `a.txt`, `b.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 2. Second task\n\n'
  printf '**Files:**\n- Modify: `b.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 8: multi-path bullet exits 0" || fail "case 8: rc=$RC out=$OUT"
EXPECTED='bundle 1: 1 2'
[ "$OUT" = "$EXPECTED" ] && pass "case 8: second declared path in a multi-path bullet still joins the bundle" || fail "case 8: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 9: task ids sort numerically, not lexically — "10" after "2", not
# before it, which is the order sorting their text would give.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 2. Earlier task\n\n'
  printf '**Files:**\n- Create: `x.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 10. Later task\n\n'
  printf '**Files:**\n- Create: `y.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 9: two-digit ids exit 0" || fail "case 9: rc=$RC out=$OUT"
EXPECTED=$'bundle 1: 2\nbundle 2: 10'
[ "$OUT" = "$EXPECTED" ] && pass "case 9: 10 sorts after 2 numerically" || fail "case 9: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 10: a step checkbox directly after a **Files:** bullet block, with no
# blank line between them, must not be consumed as a bogus file-path entry.
# Guards against BULLET_RE (which a checkbox line's `- ` prefix would
# otherwise satisfy) matching a `- [ ]` line while `in_files` is still true.
# The two tasks declare DISJOINT paths and carry the IDENTICAL step text
# immediately after their Files blocks, so they must produce two bundles —
# if the bug is present, both swallow that same step line as a declared
# path, the shared bogus path joins them, and one bundle comes back.
#
# This is the ONE fixture here whose steps are deliberately left UNINDENTED,
# at column 0, where every other fixture indents them two columns as a plan
# now does. That is the whole point of it: an indented step never reaches
# BULLET_RE (which is anchored at `^- `), so indenting this fixture would
# retire the lookahead's only reproducer while the lookahead still has to
# hold for a hand-written plan that gets the indent wrong. Such a line is a
# "malformed task line" finding in `spectre validate`; it must still not
# become a declared path here.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. First\n\n'
  printf '**Files:**\n- Modify: `a.txt`\n'
  printf -- '- [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 2. Second\n\n'
  printf '**Files:**\n- Modify: `b.txt`\n'
  printf -- '- [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 10: adjacent step checkbox after Files block exits 0" || fail "case 10: rc=$RC out=$OUT"
EXPECTED=$'bundle 1: 1\nbundle 2: 2'
[ "$OUT" = "$EXPECTED" ] && pass "case 10: the step line is not swallowed as a declared path" || fail "case 10: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 11: with PLAN_DISPATCH_BUNDLES_ROOT unset and no argument, the guard
# must derive its repository root from its OWN resolved location — not from a
# fixed "one level up above $SCRIPT_DIR", which only holds while it lives at
# <repo>/scripts/. Built here: a scratch tree where the guard is reachable at
# two depths, its real home (scratch-repo/scripts/) and a
# skills/myflow-do/scripts/ symlink, mirroring how setup.sh's install carries
# it. Invoked through the symlink with no argument, it must scan THAT tree's
# own spectre/changes/ — never skills/myflow-do/spectre/changes/, which does
# not exist and would silently scan nothing (see design.md, "The
# $SCRIPT_DIR/.. hazard").
# ===========================================================================
new_fixture
mkdir -p "$FIXTURE/scripts/lib" "$FIXTURE/skills/myflow-do/scripts" \
  "$FIXTURE/spectre/changes/some-change"
cp "$GUARD" "$FIXTURE/scripts/plan-dispatch-bundles.sh"
chmod +x "$FIXTURE/scripts/plan-dispatch-bundles.sh"
cp "$SCRIPT_DIR/plan-dispatch-bundles.py" "$FIXTURE/scripts/plan-dispatch-bundles.py"
cp "$SCRIPT_DIR/lib/resolve-file.sh" "$FIXTURE/scripts/lib/resolve-file.sh"
ln -s ../../../scripts/plan-dispatch-bundles.sh \
  "$FIXTURE/skills/myflow-do/scripts/plan-dispatch-bundles.sh"
ln -s ../../../scripts/plan-dispatch-bundles.py \
  "$FIXTURE/skills/myflow-do/scripts/plan-dispatch-bundles.py"
ln -s ../../../scripts/lib "$FIXTURE/skills/myflow-do/scripts/lib"
{
  printf -- '- [ ] 1. Fieldless task\n\n'
  printf 'No Files field here at all.\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$FIXTURE/spectre/changes/some-change/tasks.md"
set +e
OUT="$("$FIXTURE/skills/myflow-do/scripts/plan-dispatch-bundles.sh" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 1 ] && pass "case 11: no-arg default resolves through a skill-dir symlink to the real repo root" \
  || fail "case 11: expected rc=1 (found the fixture's own tasks.md), got rc=$RC out=$OUT"
case "$OUT" in
  *"task 1"*"no **Files:** field"*) pass "case 11: names task 1 from the fixture's own spectre/changes/" ;;
  *) fail "case 11: expected message naming task 1, out=$OUT" ;;
esac

# ===========================================================================
# Case 12: a `### <id> <title>` heading opens no task. This guard keeps its
# own copy of the task-line pattern, mirroring check-task-build-green.py's
# rather than importing it, so the heading shape a myflow plan used to mark
# a task with has to be pinned as inert HERE too — a copy that drifted back
# into accepting both shapes would put this guard and spectre right back
# into the disagreement the checkbox line ended. The heading below declares
# no **Files:** field, so a guard still reading it as an unchecked task
# would report it and exit 1; the one real task above it is the only bundle.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. The only real task\n\n'
  printf '**Files:**\n- Create: `a.txt`\n\n'
  printf '### 2 A heading, not a task\n\n'
  printf 'No Files field here at all.\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 12: a level-3 heading opens no task" || fail "case 12: rc=$RC out=$OUT"
EXPECTED='bundle 1: 1'
[ "$OUT" = "$EXPECTED" ] && pass "case 12: only the checkbox task bundles" || fail "case 12: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 13: a task's id is a FLAT integer here too. `- [ ] 1.1. …` is a
# "malformed task line" finding to spectre and no task to it; this guard
# keeps its own copy of the task-line pattern rather than importing it, so
# the narrowing has to be pinned against THAT copy — check-task-build-
# green.sh's case 27 pins it against the shared grammar. The dotted line
# below declares no **Files:** field, so a copy that still admitted a dotted
# id would read it as an unchecked task and exit 1 naming it.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. The only real task\n\n'
  printf '**Files:**\n- Create: `a.txt`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n\n'
  printf -- '- [ ] 1.1. A dotted task line, which is no task\n\n'
  printf 'No Files field here at all.\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 13: a dotted task line opens no task" || fail "case 13: rc=$RC out=$OUT"
EXPECTED='bundle 1: 1'
[ "$OUT" = "$EXPECTED" ] && pass "case 13: only the flat-id task bundles" || fail "case 13: expected [$EXPECTED], got [$OUT]"

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
