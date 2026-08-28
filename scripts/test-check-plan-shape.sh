#!/usr/bin/env bash
# Assertion harness for check-plan-shape.sh / check-plan-shape.py.
#
# Modeled on test-check-task-build-green.sh's fixture-driven pattern:
# fixtures live under mktemp -d, the guard is invoked via a thin run_guard
# helper that captures RC/OUT, and every case ends with an explicit
# pass/fail assertion. Runs the REAL scripts/check-plan-shape.sh against
# REAL fixture files on disk — never a copy of its logic, never its Python
# internals imported and asserted on directly.
#
# One case per finding (F1-F6), a clean-plan case, a no-arg aggregation
# case, and three exit-2 cases (missing sibling module, unreadable file,
# usage error). Additionally, for each of F1-F6, a MUTATION case: a
# throwaway copy of check-plan-shape.py with that finding's own check
# disabled (via sed against a line this file tags `# F<n>`) is run against
# the SAME fixture, and the finding's message must then be ABSENT — proving
# the fixture's failure actually depends on that specific check, not on
# some other check incidentally catching the same fixture. Mirrors
# test-check-worktree-processes.sh's case 7 (KAN-197 mutation) and the
# requirement scripts/test-check-reproduce-not-read.sh already meets: a
# suite that cannot detect the guard's own defect is not a suite.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-plan-shape.sh"
PY_GUARD="$SCRIPT_DIR/check-plan-shape.py"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do
    rm -rf "$d"
  done
}
trap cleanup EXIT

# run_guard <tasks.md-path> -> sets RC and OUT
run_guard() {
  set +e
  OUT="$("$GUARD" "$1" 2>&1)"
  RC=$?
  set -e
}

# run_guard_root <root> -> sets RC and OUT; the no-argument aggregation
# scan, pointed at <root> via CHECK_PLAN_SHAPE_ROOT rather than a real cwd
# — the same override convention CHECK_TASK_BUILD_GREEN_ROOT uses.
run_guard_root() {
  set +e
  OUT="$(CHECK_PLAN_SHAPE_ROOT="$1" "$GUARD" 2>&1)"
  RC=$?
  set -e
}

new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/plan-shape-test.XXXXXX")"
  DIRS+=("$FIXTURE")
  TASKS_MD="$FIXTURE/tasks.md"
}

# assert_finding_absent_when_mutated <finding-tag> <fixture-path> <message-substring> <case-label>
#
# Copies check-plan-shape.py into a throwaway directory alongside its real
# siblings (check-task-commit-fields.sh's .py and lib/plan_grammar.py,
# symlinked in place so the mutant still resolves them through its own
# SCRIPT_DIR), disables the tagged check with sed, and asserts the
# finding's message no longer appears in the mutant's output against the
# SAME fixture. Asserts the sed edit actually applied first, so a mutation
# that stopped matching (because the guard was reworded) is caught rather
# than silently passing.
assert_mutation_removes_finding() {
  local tag="$1" fixture_path="$2" message_substring="$3" label="$4"
  local mutant_dir mutant_py mutant_sh out rc

  mutant_dir="$(mktemp -d "${TMPDIR:-/tmp}/plan-shape-mutant.XXXXXX")"
  DIRS+=("$mutant_dir")
  mkdir -p "$mutant_dir/lib"
  ln -s "$SCRIPT_DIR/check-task-commit-fields.py" "$mutant_dir/check-task-commit-fields.py"
  ln -s "$SCRIPT_DIR/lib/plan_grammar.py" "$mutant_dir/lib/plan_grammar.py"
  cp "$SCRIPT_DIR/check-plan-shape.sh" "$mutant_dir/check-plan-shape.sh"
  chmod +x "$mutant_dir/check-plan-shape.sh"
  mutant_py="$mutant_dir/check-plan-shape.py"

  sed "/# ${tag}\$/s/^\( *\)if .*/\1if False:  # ${tag} (mutated)/" \
    "$PY_GUARD" > "$mutant_py"

  if cmp -s "$PY_GUARD" "$mutant_py"; then
    fail "$label (mutation): the mutation did not apply — no line tagged '# ${tag}' matched, so nothing was mutated"
    return
  fi
  pass "$label (mutation): the mutation applied"

  set +e
  out="$("$mutant_dir/check-plan-shape.sh" "$fixture_path" 2>&1)"
  rc=$?
  set -e

  case "$out" in
    *"$message_substring"*)
      fail "$label (mutation): the finding still fires with its own check disabled, out=$out"
      ;;
    *)
      pass "$label (mutation): disabling the check removes the finding"
      ;;
  esac
}

# ===========================================================================
# Case 1: a clean plan -> exit 0, no output.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. A clean task\n\n'
  printf '**Files:** `scripts/foo.sh`\n'
  printf '**Tests:** `scripts/test-foo.sh`\n'
  printf '**Commit:** `feat(scripts): add foo`\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 1: a clean plan passes" || fail "case 1: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 1: no output" || fail "case 1: expected no output, got: $OUT"

# ===========================================================================
# Case 2 (F1): a second **Files:** line in one task body -> exit 1, names
# the task id and the earlier line.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Duplicate Files\n\n'
  printf '**Files:** `a.py`\n\n'
  printf '**Tests:** `test-a.sh`\n\n'
  printf '**Files:** `b.py`\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 2 (F1): duplicate Files line fails" || fail "case 2 (F1): rc=$RC out=$OUT"
case "$OUT" in
  *"task 1 declares a second **Files:**"*) pass "case 2 (F1): names the task and the field" ;;
  *) fail "case 2 (F1): expected message naming task 1's second Files line, out=$OUT" ;;
esac
# Exact file:line prefix, not merely the message substring: the fixture's
# second **Files:** line is physical line 7 (task line 1, blank 2, first
# Files 3, blank 4, Tests 5, blank 6, second Files 7); an off-by-one or
# off-by-two in _check_f1's `file_line = body_start + offset + 1` passes
# every other assertion in this case while naming the wrong line.
case "$OUT" in
  "$TASKS_MD:7: task 1 declares a second"*) pass "case 2 (F1): the exact file:line prefix is line 7" ;;
  *) fail "case 2 (F1): expected the message to be prefixed \"$TASKS_MD:7:\", out=$OUT" ;;
esac
assert_mutation_removes_finding "F1" "$TASKS_MD" "declares a second **Files:**" "case 2"

# ===========================================================================
# Case 3 (F2): no **Files:** line at all -> exit 1.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. No Files line\n\n'
  printf '**Tests:** `test-a.sh`\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 3 (F2): no Files line fails" || fail "case 3 (F2): rc=$RC out=$OUT"
case "$OUT" in
  *"task 1 has no **Files:** line"*) pass "case 3 (F2): names the task" ;;
  *) fail "case 3 (F2): expected message naming task 1, out=$OUT" ;;
esac
assert_mutation_removes_finding "F2" "$TASKS_MD" "has no **Files:** line" "case 3"

# ===========================================================================
# Case 4 (F2): a **Files:** line present but empty -> exit 1, same finding.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Empty Files line\n\n'
  printf '**Files:**\n\n'
  printf '**Tests:** `test-a.sh`\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 4 (F2): an empty Files line fails" || fail "case 4 (F2): rc=$RC out=$OUT"
case "$OUT" in
  *"task 1 has no **Files:** line"*) pass "case 4 (F2): names the task" ;;
  *) fail "case 4 (F2): expected message naming task 1, out=$OUT" ;;
esac

# ===========================================================================
# Case 5 (F3a): a **Files:** line indented past column 0, invisible to
# FIELD_RE -> exit 1, distinct message from F2 even though F2 also fires
# (the real parser sees no Files line either, since it is indented).
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Indented Files line\n\n'
  printf '  **Files:** `a.py`\n\n'
  printf '**Tests:** `test-a.sh`\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 5 (F3a): an indented Files line fails" || fail "case 5 (F3a): rc=$RC out=$OUT"
case "$OUT" in
  *"indented past column 0"*) pass "case 5 (F3a): names the indentation defect" ;;
  *) fail "case 5 (F3a): expected an indentation message, out=$OUT" ;;
esac
# Exact file:line prefix, not merely the message substring: the fixture's
# indented **Files:** line is physical line 3 (task line 1, blank 2,
# indented Files 3); an off-by-one or off-by-two in _check_f3a's
# `file_line = body_start + offset + 1` passes every other assertion in
# this case while naming the wrong line.
case "$OUT" in
  "$TASKS_MD:3: task 1 carries a"*) pass "case 5 (F3a): the exact file:line prefix is line 3" ;;
  *) fail "case 5 (F3a): expected the message to be prefixed \"$TASKS_MD:3:\", out=$OUT" ;;
esac
assert_mutation_removes_finding "F3a" "$TASKS_MD" "indented past column 0" "case 5"

# ===========================================================================
# Case 6 (F3b): a task body that opens a fence and never closes it -> exit
# 1, names the fence rather than blaming the swallowed Files field.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Unclosed fence\n\n'
  printf '```\n'
  printf '**Files:** `a.py`\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 6 (F3b): an unclosed fence fails" || fail "case 6 (F3b): rc=$RC out=$OUT"
case "$OUT" in
  *"never closed in its body"*) pass "case 6 (F3b): names the unclosed fence" ;;
  *) fail "case 6 (F3b): expected an unclosed-fence message, out=$OUT" ;;
esac
# design.md: F3b suppresses F1/F2/F4/F6 for the same task — the cause is
# named rather than the consequence. This fixture's body has no
# **Files:** line outside the fence, which would ALSO be F2 (and its
# **Tests:** field is entirely absent, which would ALSO be F6) were F3b
# not an early return; asserting only that "never closed in its body" is
# PRESENT (above) does not prove those are absent — a mutation that turns
# F3b's early return into a plain append (still reporting F3b, but
# falling through to F1/F2/F4/F6 as well) passes that assertion too. So
# also assert the exact line count and that the suppressed findings'
# messages do not appear.
[ "$(printf '%s\n' "$OUT" | wc -l | tr -d ' ')" = "1" ] && pass "case 6 (F3b): exactly one violation line is printed" || fail "case 6 (F3b): expected exactly one violation line, out=$OUT"
case "$OUT" in
  *"has no **Files:** line"*) fail "case 6 (F3b): F2's message leaked through despite the unclosed fence, out=$OUT" ;;
  *) pass "case 6 (F3b): F2 is suppressed" ;;
esac
case "$OUT" in
  *"declares nothing"*) fail "case 6 (F3b): F6's message leaked through despite the unclosed fence, out=$OUT" ;;
  *) pass "case 6 (F3b): F6 is suppressed" ;;
esac
assert_mutation_removes_finding "F3b" "$TASKS_MD" "never closed in its body" "case 6"

# ===========================================================================
# Case 7 (F4): a Tests: field opening with `none` on a task whose Files:
# names a test-shaped path -> exit 1, a contradiction between the opt-out
# and the declared files.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. None with a test-shaped file\n\n'
  printf '**Files:** `scripts/test-check-foo.sh`\n\n'
  printf '**Tests:** none added — nothing to verify\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 7 (F4): none-opening Tests with a test-shaped Files entry fails" || fail "case 7 (F4): rc=$RC out=$OUT"
case "$OUT" in
  *"contradiction between the opt-out"*) pass "case 7 (F4): names the contradiction" ;;
  *) fail "case 7 (F4): expected a contradiction message, out=$OUT" ;;
esac
assert_mutation_removes_finding "F4" "$TASKS_MD" "contradiction between the opt-out" "case 7"

# ===========================================================================
# Case 8: a none-opening Tests field on a task whose Files: does NOT name a
# test-shaped path -> F4 does not fire (F6 must not fire either: `none` is
# a legitimate declaration of zero tests).
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. None with an ordinary file\n\n'
  printf '**Files:** `scripts/foo.sh`\n\n'
  printf '**Tests:** none added — nothing to verify\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 8: none-opening Tests with an ordinary Files entry passes" || fail "case 8: rc=$RC out=$OUT"

# ===========================================================================
# Case 9 (F5): a tasks.md with zero checkbox lines -> exit 1, a vacuous-pass
# defect rather than a silent clean report.
# ===========================================================================
new_fixture
{
  printf '# A plan with no tasks\n\nJust prose, no checkbox lines at all.\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 9 (F5): zero tasks fails" || fail "case 9 (F5): rc=$RC out=$OUT"
case "$OUT" in
  *"zero tasks"*) pass "case 9 (F5): names the vacuous-pass defect" ;;
  *) fail "case 9 (F5): expected a zero-tasks message, out=$OUT" ;;
esac
assert_mutation_removes_finding "F5" "$TASKS_MD" "zero tasks" "case 9"

# ===========================================================================
# Case 10 (F6): a Tests: field that is neither none-opening, nor carries a
# Case N label, nor carries a backtick token -> exit 1, declares nothing.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Vague Tests field\n\n'
  printf '**Files:** `a.py`\n\n'
  printf '**Tests:** the existing harnesses, unchanged, plus a case pinning the fix\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 10 (F6): a vague Tests field fails" || fail "case 10 (F6): rc=$RC out=$OUT"
case "$OUT" in
  *"declares nothing"*) pass "case 10 (F6): names the vacuous field" ;;
  *) fail "case 10 (F6): expected a declares-nothing message, out=$OUT" ;;
esac
assert_mutation_removes_finding "F6" "$TASKS_MD" "declares nothing" "case 10"

# ===========================================================================
# Case 11: a Tests: field carrying a Case N label -> F6 does not fire.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Case-labeled Tests field\n\n'
  printf '**Files:** `a.py`\n\n'
  printf '**Tests:** Case 3 pins the new behaviour\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 11: a Case N label satisfies Tests:" || fail "case 11: rc=$RC out=$OUT"

# ===========================================================================
# Case 12: a Tests: field carrying a backtick token -> F6 does not fire.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Backtick Tests field\n\n'
  printf '**Files:** `a.py`\n\n'
  printf '**Tests:** `scripts/test-a.sh`\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 12: a backtick token satisfies Tests:" || fail "case 12: rc=$RC out=$OUT"

# ===========================================================================
# Case 13: fenced example text is excluded from every scan — a duplicate
# **Files:** line INSIDE a fenced worked example is not a real second
# declaration.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Fenced example, not a real duplicate\n\n'
  printf '**Files:** `a.py`\n\n'
  printf '**Tests:** `test-a.sh`\n\n'
  printf 'Example of the field grammar:\n\n'
  printf '```\n'
  printf '**Files:** `not-a-real-declaration.py`\n'
  printf '```\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 13: a fenced example is not read as a second Files line" || fail "case 13: rc=$RC out=$OUT"

# ===========================================================================
# Case 14: no-argument aggregation scan — one clean tasks.md, one dirty one,
# under a sandboxed root via CHECK_PLAN_SHAPE_ROOT, mirroring
# CHECK_TASK_BUILD_GREEN_ROOT's override convention.
# ===========================================================================
ROOT="$(mktemp -d "${TMPDIR:-/tmp}/plan-shape-root-test.XXXXXX")"
DIRS+=("$ROOT")
mkdir -p "$ROOT/spectre/changes/clean-change" "$ROOT/spectre/changes/dirty-change" \
  "$ROOT/spectre/changes/archive/archived-change"
{
  printf -- '- [ ] 1. Clean\n\n'
  printf '**Files:** `a.py`\n\n'
  printf '**Tests:** `test-a.sh`\n'
} > "$ROOT/spectre/changes/clean-change/tasks.md"
{
  printf -- '- [ ] 1. Dirty\n\n'
  printf '**Tests:** `test-a.sh`\n'
} > "$ROOT/spectre/changes/dirty-change/tasks.md"
{
  printf -- '- [ ] 1. Archived, must not be scanned\n\n'
} > "$ROOT/spectre/changes/archive/archived-change/tasks.md"
run_guard_root "$ROOT"
[ "$RC" -eq 1 ] && pass "case 14: no-arg scan aggregates a violation from the dirty file" || fail "case 14: rc=$RC out=$OUT"
case "$OUT" in
  *"dirty-change/tasks.md:"*"has no **Files:** line"*) pass "case 14: names the dirty file" ;;
  *) fail "case 14: expected the dirty file's violation, out=$OUT" ;;
esac
case "$OUT" in
  *"clean-change"*) fail "case 14: the clean file's path leaked into the output, out=$OUT" ;;
  *) pass "case 14: the clean file reports nothing" ;;
esac
case "$OUT" in
  *"archived-change"*) fail "case 14: an archived change was scanned, out=$OUT" ;;
  *) pass "case 14: archived changes are excluded from the no-arg scan" ;;
esac

# ===========================================================================
# Case 15: too many arguments -> usage error, exit 2.
# ===========================================================================
set +e
OUT="$("$GUARD" a b 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 15: too many arguments exits 2" || fail "case 15: rc=$RC out=$OUT"

# ===========================================================================
# Case 16: an unreadable/nonexistent tasks.md path -> exit 2, never a silent
# skip.
# ===========================================================================
run_guard "$FIXTURE/does-not-exist.md"
[ "$RC" -eq 2 ] && pass "case 16: a nonexistent file exits 2" || fail "case 16: rc=$RC out=$OUT"

# ===========================================================================
# Case 17: a copy of the guard pair with no check-task-commit-fields.py
# sibling -> exit 2, naming the missing module rather than a traceback.
# Mirrors test-check-task-build-green.sh's case 20.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Clean task\n\n'
  printf '**Files:** `a.py`\n\n'
  printf '**Tests:** `test-a.sh`\n'
} > "$TASKS_MD"
STRIPPED="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/plan-shape-stripped-test.XXXXXX")" && pwd)"
DIRS+=("$STRIPPED")
mkdir -p "$STRIPPED/lib"
cp "$SCRIPT_DIR/check-plan-shape.sh" "$SCRIPT_DIR/check-plan-shape.py" "$STRIPPED/"
cp "$SCRIPT_DIR/lib/spec-root.sh" "$STRIPPED/lib/"
cp "$SCRIPT_DIR/lib/plan_grammar.py" "$STRIPPED/lib/"
set +e
STRIPPED_OUT="$("$STRIPPED/check-plan-shape.sh" "$TASKS_MD" 2>&1)"
STRIPPED_RC=$?
set -e
[ "$STRIPPED_RC" -eq 2 ] && pass "case 17: a guard copy with no check-task-commit-fields.py sibling exits 2" || fail "case 17: rc=$STRIPPED_RC out=$STRIPPED_OUT"
case "$STRIPPED_OUT" in
  *"check-task-commit-fields.py"*) pass "case 17: the message names the missing module" ;;
  *) fail "case 17: expected the missing-module path in the message, out=$STRIPPED_OUT" ;;
esac

# ===========================================================================
# Case 18 (KAN-121 panel-fix round, finding 3): the guard invoked through
# its shipped skills/flow/scripts/ symlink -> REPO_ROOT resolves to the
# TRUE repository root, not to skills/flow itself. This is the guard's own
# production path: skills/flow/brainstorm.md step D invokes it bare, and a
# named guard resolves to <skill-dir>/scripts/<name> per Guard resolution
# (skills/flow-contracts/pipeline.md). REPO_ROOT only matters on the
# no-argument aggregation path (the one-argument path never reads it), so
# this case has to exercise that path through a REAL symlinked
# invocation — not CHECK_PLAN_SHAPE_ROOT, which bypasses REPO_ROOT
# resolution entirely and would prove nothing about it.
#
# Built as a full sandbox clone of the real symlink-farm shape: a
# SANDBOX/scripts/ directory holding the guard and its real siblings
# (symlinked, so this always tracks the checked-out sources) plus a
# SANDBOX/skills/flow/scripts/ symlink farm at the SAME relative depth
# (../../../scripts/...) the real repository uses, and a dirty tasks.md
# fixture placed only under the sandbox's own spectre/changes/ — never
# this repository's own, whose contents are outside this suite's control.
# A REPO_ROOT that resolved to skills/flow instead of the sandbox root
# would see no spectre/changes/ directory there at all and exit 0 with no
# output — the exact silent false-clean this case exists to catch.
# ===========================================================================
SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/plan-shape-symlink-test.XXXXXX")"
DIRS+=("$SANDBOX")
mkdir -p "$SANDBOX/scripts" "$SANDBOX/skills/flow/scripts" \
  "$SANDBOX/spectre/changes/dirty-change"
cp "$SCRIPT_DIR/check-plan-shape.sh" "$SCRIPT_DIR/check-plan-shape.py" "$SANDBOX/scripts/"
chmod +x "$SANDBOX/scripts/check-plan-shape.sh"
ln -s "$SCRIPT_DIR/check-task-commit-fields.py" "$SANDBOX/scripts/check-task-commit-fields.py"
ln -s "$SCRIPT_DIR/lib" "$SANDBOX/scripts/lib"
ln -s "../../../scripts/check-plan-shape.sh" "$SANDBOX/skills/flow/scripts/check-plan-shape.sh"
ln -s "../../../scripts/check-plan-shape.py" "$SANDBOX/skills/flow/scripts/check-plan-shape.py"
ln -s "../../../scripts/check-task-commit-fields.py" "$SANDBOX/skills/flow/scripts/check-task-commit-fields.py"
ln -s "../../../scripts/lib" "$SANDBOX/skills/flow/scripts/lib"
{
  printf -- '- [ ] 1. Dirty\n\n'
  printf '**Tests:** `test-a.sh`\n'
} > "$SANDBOX/spectre/changes/dirty-change/tasks.md"

set +e
SYMLINK_OUT="$("$SANDBOX/skills/flow/scripts/check-plan-shape.sh" 2>&1)"
SYMLINK_RC=$?
set -e
[ "$SYMLINK_RC" -eq 1 ] && pass "case 18: the symlinked invocation resolves REPO_ROOT to the sandbox root and finds the dirty fixture" || fail "case 18: rc=$SYMLINK_RC out=$SYMLINK_OUT"
case "$SYMLINK_OUT" in
  *"dirty-change/tasks.md:"*"has no **Files:** line"*) pass "case 18: names the dirty fixture" ;;
  *) fail "case 18: expected the dirty fixture's violation through the symlinked invocation, out=$SYMLINK_OUT" ;;
esac

# ===========================================================================
# Case 19 (F4/F6): `none` opening on a CONTINUATION line, not on the
# **Tests:** field's own line. parse_task_fields joins a field's continuation
# lines before applying NONE_OPEN_RE, so this IS a none-opening field; a
# guard reading only the field's own physical line sees an empty value and
# reports F6 where F4 is correct. This case is what pins tests_value: without
# it, reverting to the physical-line form passes the whole suite (panel round
# 2, Bugbot).
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Continuation-line none, test-shaped file\n\n'
  printf '**Files:** `scripts/test-check-foo.sh`\n\n'
  printf '**Tests:**\n'
  printf 'none added — nothing to verify\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 1 ] && pass "case 19 (F4): a continuation-line none over a test-shaped Files entry fails" || fail "case 19 (F4): rc=$RC out=$OUT"
case "$OUT" in
  *"contradiction between the opt-out"*) pass "case 19 (F4): names the contradiction, not F6" ;;
  *) fail "case 19 (F4): expected the F4 contradiction message, got out=$OUT" ;;
esac

# The same field over an ordinary Files entry is a legitimate opt-out and
# must pass clean — the direction that fails loudly when tests_value is
# reverted to the field's own physical line.
new_fixture
{
  printf -- '- [ ] 1. Continuation-line none, ordinary file\n\n'
  printf '**Files:** `scripts/foo.sh`\n\n'
  printf '**Tests:**\n'
  printf 'none added — nothing to verify\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 19: a continuation-line none over an ordinary Files entry passes" || fail "case 19: rc=$RC out=$OUT"

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
