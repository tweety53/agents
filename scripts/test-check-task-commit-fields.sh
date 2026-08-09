#!/usr/bin/env bash
# Assertion harness for check-task-commit-fields.sh / .py. Builds a
# throwaway git repo per case under a sandboxed TMPDIR, makes a real commit
# in it, writes a tasks.md declaring that task's Files:/Tests:/Commit:
# fields, and asserts the guard's exit status and message text against the
# real commit. Never touches this repository's own tree or history.
#
# Modeled on test-check-task-build-green.sh's pattern: fixtures live under
# mktemp -d, the guard is invoked via a thin run_guard helper that captures
# RC/OUT, and every case ends with an explicit pass/fail assertion.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-task-commit-fields.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# run_guard <worktree> <task-id> <commit-sha> [parent-sha] -> sets RC and OUT
run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

# new_repo -> sets REPO to a fresh throwaway git repo with one root commit,
# so every case's task commit has a real parent to diff against.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/task-commit-fields-test.XXXXXX")"
  git -C "$REPO" init -q
  git -C "$REPO" config user.email "test@example.com"
  git -C "$REPO" config user.name "Test"
  mkdir -p "$REPO/openspec/changes/change-a"
  printf 'root\n' > "$REPO/root.txt"
  git -C "$REPO" add root.txt
  git -C "$REPO" commit -q -m "root"
}

# write_tasks_md <repo> <content> -> writes the change's tasks.md
write_tasks_md() {
  printf '%s' "$2" > "$1/openspec/changes/change-a/tasks.md"
}

# write_project_md_test_section <repo> <command-line> -> writes a
# `.myflow/project.md` whose `## test` section names exactly one command —
# the shape check-task-commit-fields.py's read_single_test_command requires
# in order to target Regression:/Baseline: at all.
write_project_md_test_section() {
  mkdir -p "$1/.myflow"
  printf '## test\n\n```\n%s\n```\n' "$2" > "$1/.myflow/project.md"
}

# write_test_runner <repo> -> a synthetic `## test` command
# (run_suite.sh + suite.txt) that speaks check-task-commit-fields.py's own
# minimal test-runner contract: `--only <name>` prints `RESULT <name>:
# pass|fail` depending on whether <name> is a line in suite.txt, and no
# arguments prints `COUNT: <N>`, N being suite.txt's line count. Committed
# once, ahead of the plan and task commits, so a task commit's own revert
# never touches the runner itself — only the suite.txt line(s) it added.
write_test_runner() {
  cat > "$1/run_suite.sh" <<'RUNNER'
#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
SUITE="$DIR/suite.txt"
if [ "${1:-}" = "--only" ]; then
  name="$2"
  if [ -f "$SUITE" ] && grep -qxF "$name" "$SUITE"; then
    echo "RESULT $name: pass"
  else
    echo "RESULT $name: fail"
  fi
  exit 0
fi
count=0
[ -f "$SUITE" ] && count="$(wc -l < "$SUITE" | tr -d ' ')"
echo "COUNT: $count"
RUNNER
  chmod +x "$1/run_suite.sh"
  : > "$1/suite.txt"
  git -C "$1" add run_suite.sh suite.txt
  git -C "$1" commit -q -m "fixture: add run_suite.sh test runner"
  write_project_md_test_section "$1" "./run_suite.sh"
  git -C "$1" add .myflow/project.md
  git -C "$1" commit -q -m "fixture: configure ## test"
}

# write_unsupported_test_runner <repo> -> a single `## test` command that
# never emits a RESULT or COUNT line, whatever it is given — the shape
# Regression:/Baseline: fall back to skipped-not-verified against.
write_unsupported_test_runner() {
  cat > "$1/run_dumb.sh" <<'RUNNER'
#!/usr/bin/env bash
echo "ran the whole suite, however many args you gave me"
RUNNER
  chmod +x "$1/run_dumb.sh"
  git -C "$1" add run_dumb.sh
  git -C "$1" commit -q -m "fixture: add run_dumb.sh test runner"
  write_project_md_test_section "$1" "./run_dumb.sh"
  git -C "$1" add .myflow/project.md
  git -C "$1" commit -q -m "fixture: configure ## test"
}

# ===========================================================================
# Case 1: commit's changed files are a subset of declared Files: -> exit 0.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 1.1 Clean task

**Files:** `alpha.txt`, `beta.txt`
**Tests:** `test_alpha`
**Commit:** add alpha and beta
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf '# test_alpha covers alpha\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 1: files subset of declared passes" || fail "case 1: rc=$RC out=$OUT"

# ===========================================================================
# Case 2: commit touches a file not in declared Files: -> exit 1.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 2.1 Undeclared file

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha only
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf '# test_alpha covers alpha\n' > "$REPO/alpha.txt"
printf 'gamma\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt gamma.txt
git -C "$REPO" commit -q -m "add alpha only"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 2.1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 2: undeclared file fails" || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"gamma.txt"*"not declared"*) pass "case 2: names the undeclared file" ;;
  *) fail "case 2: expected message naming gamma.txt as undeclared, out=$OUT" ;;
esac

# ===========================================================================
# Case 3: declared test name found in the commit's diff -> exit 0.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 3.1 Test present

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 3.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 3: declared test name found in diff passes" || fail "case 3: rc=$RC out=$OUT"

# ===========================================================================
# Case 4: declared test name missing from the commit's diff -> exit 1.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 4.1 Test missing

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'no tests here\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 4.1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 4: missing declared test fails" || fail "case 4: rc=$RC out=$OUT"
case "$OUT" in
  *"test_alpha"*"not found in the diff"*) pass "case 4: names the missing test" ;;
  *) fail "case 4: expected message naming test_alpha as missing, out=$OUT" ;;
esac

# ===========================================================================
# Case 5: commit subject matches declared Commit: -> exit 0.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 5.1 Subject matches

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha for real
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha for real"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 5.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 5: commit subject matches declared Commit: passes" || fail "case 5: rc=$RC out=$OUT"

# ===========================================================================
# Case 6: commit subject does not match declared Commit: -> exit 1.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 6.1 Subject mismatch

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha for real
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha, not quite right"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 6.1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 6: commit subject mismatch fails" || fail "case 6: rc=$RC out=$OUT"
case "$OUT" in
  *"subject"*"does not match"*) pass "case 6: reports the subject mismatch" ;;
  *) fail "case 6: expected message reporting subject mismatch, out=$OUT" ;;
esac

# ===========================================================================
# Case 7: extra path covered by Allowed-collateral: glob -> exit 0.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 7.1 Collateral covered

**Files:** `alpha.txt`
**Allowed-collateral:** `docs/*.md`
**Tests:** `test_alpha`
**Commit:** add alpha and sweep docs
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
mkdir -p "$REPO/docs"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
printf 'swept\n' > "$REPO/docs/notes.md"
git -C "$REPO" add alpha.txt docs/notes.md
git -C "$REPO" commit -q -m "add alpha and sweep docs"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 7.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 7: extra path covered by Allowed-collateral: glob passes" || fail "case 7: rc=$RC out=$OUT"

# ===========================================================================
# Case 8: Tests: written as real free prose naming cases by number (the
# shape every real Tests: field in this plan's own tasks.md actually uses,
# e.g. "Case 1: files subset of declared passes; Case 2: undeclared file
# fails; ..."), with both cases' "# Case N:" markers present in the diff ->
# exit 0. Proves the guard does not require backtick-quoted identifiers.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 8.1 Prose cases present

**Files:** `guard_test.sh`
**Tests:** Case 1: files subset of declared passes; Case 2: undeclared file
fails
**Commit:** add guard test cases
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
{
  printf '# Case 1: files subset of declared passes\n'
  printf '# Case 2: undeclared file fails\n'
} > "$REPO/guard_test.sh"
git -C "$REPO" add guard_test.sh
git -C "$REPO" commit -q -m "add guard test cases"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 8.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 8: prose Case N: labels found in diff passes" || fail "case 8: rc=$RC out=$OUT"

# ===========================================================================
# Case 9: same prose shape as case 8, but the diff only carries "# Case 1:",
# never "Case 2" -> exit 1, naming "Case 2" (not the whole sentence).
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 9.1 Prose case missing

**Files:** `guard_test.sh`
**Tests:** Case 1: files subset of declared passes; Case 2: undeclared file
fails
**Commit:** add guard test case one only
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf '# Case 1: files subset of declared passes\n' > "$REPO/guard_test.sh"
git -C "$REPO" add guard_test.sh
git -C "$REPO" commit -q -m "add guard test case one only"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 9.1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 9: missing prose case fails" || fail "case 9: rc=$RC out=$OUT"
case "$OUT" in
  *"Case 2"*"not found in the diff"*) pass "case 9: names Case 2, not the whole sentence" ;;
  *) fail "case 9: expected message naming Case 2, out=$OUT" ;;
esac

# ===========================================================================
# Case 10: a Tests: field with neither Case N: labels nor backtick-quoted
# identifiers (e.g. "the 7 cases listed in task 3.2, run for the first time
# against the wrapper this task adds", this plan's own task 3.3 Tests: field)
# declares no checkable tests at all -> exit 0, never a false-fail against
# the whole sentence.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 10.1 No checkable tests declared

**Files:** `guard_test.sh`
**Tests:** the 7 cases listed in task 3.2, run for the first time against
the wrapper this task adds
**Commit:** add the wrapper
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'wrapper body, no case markers at all\n' > "$REPO/guard_test.sh"
git -C "$REPO" add guard_test.sh
git -C "$REPO" commit -q -m "add the wrapper"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 10.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 10: prose with no Case N: or backticks declares nothing, never false-fails" || fail "case 10: rc=$RC out=$OUT"

# ===========================================================================
# Case 11: Regression: passes — reverting the task commit makes its
# declared test fail (RESULT alpha_test: fail), and un-reverting restores
# suite.txt to passing state and HEAD to the task commit.
# ===========================================================================
new_repo
write_test_runner "$REPO"
write_tasks_md "$REPO" '### 11.1 Regression passes

**Files:** `suite.txt`
**Tests:** `alpha_test`
**Baseline:** not applicable — covered by case 13
**Commit:** add alpha_test
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'alpha_test\n' >> "$REPO/suite.txt"
git -C "$REPO" add suite.txt
git -C "$REPO" commit -q -m "add alpha_test"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 11.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 11: regression check passes (revert makes named test fail, un-revert restores it)" || fail "case 11: rc=$RC out=$OUT"
[ "$(git -C "$REPO" rev-parse HEAD)" = "$SHA" ] && pass "case 11: HEAD unchanged after un-revert" || fail "case 11: HEAD moved, expected $SHA got $(git -C "$REPO" rev-parse HEAD)"
git -C "$REPO" diff --quiet && git -C "$REPO" diff --cached --quiet && pass "case 11: worktree clean after un-revert" || fail "case 11: worktree not restored"
grep -qxF "alpha_test" "$REPO/suite.txt" && pass "case 11: suite.txt restored to contain alpha_test" || fail "case 11: suite.txt not restored"

# ===========================================================================
# Case 12: Regression: skips (not fails) when the project's ## test command
# cannot target a single named test (run_dumb.sh ignores --only entirely).
# ===========================================================================
new_repo
write_unsupported_test_runner "$REPO"
write_tasks_md "$REPO" '### 12.1 Regression skip

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 12.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 12: regression check skips (not fails) when the ## test command can't target a named test" || fail "case 12: rc=$RC out=$OUT"
case "$OUT" in
  *"Regression"*"skipped"*) pass "case 12: reports Regression: skipped, not verified" ;;
  *) fail "case 12: expected a Regression: skipped message, out=$OUT" ;;
esac

# ===========================================================================
# Case 13: Baseline: passes — the ## test command's COUNT: at the task
# commit and, reverted, at its parent match the declared before=/after=.
# ===========================================================================
new_repo
write_test_runner "$REPO"
write_tasks_md "$REPO" '### 13.1 Baseline passes

**Files:** `suite.txt`
**Tests:** `alpha_test`
**Baseline:** before=0 after=1
**Commit:** add alpha_test
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'alpha_test\n' >> "$REPO/suite.txt"
git -C "$REPO" add suite.txt
git -C "$REPO" commit -q -m "add alpha_test"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 13.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 13: baseline check passes (parent/task counts match declared before=/after=)" || fail "case 13: rc=$RC out=$OUT"
[ "$(git -C "$REPO" rev-parse HEAD)" = "$SHA" ] && pass "case 13: HEAD unchanged after baseline check" || fail "case 13: HEAD moved"

# ===========================================================================
# Case 14: Baseline: skips (not fails) when the ## test command's output
# carries no parseable COUNT: line.
# ===========================================================================
new_repo
write_unsupported_test_runner "$REPO"
write_tasks_md "$REPO" '### 14.1 Baseline skip

**Files:** `alpha.txt`
**Baseline:** before=0 after=1
**Commit:** add alpha
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'no tests here\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 14.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 14: baseline check skips (not fails) when the count is unparseable" || fail "case 14: rc=$RC out=$OUT"
case "$OUT" in
  *"Baseline"*"skipped"*) pass "case 14: reports Baseline: skipped, not verified" ;;
  *) fail "case 14: expected a Baseline: skipped message, out=$OUT" ;;
esac

# ===========================================================================
# Case 15: a field-looking line inside a fenced code block in the task's own
# body (e.g. a `**Files:**` line shown as illustrative text) must not be
# parsed as a real field. Files: stays `alpha.txt`, not the fenced example's
# `evil.txt`, so a commit touching only alpha.txt still passes.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '### 15.1 Fence guard

**Files:** `alpha.txt`

Example of the field grammar, not a real field:

```
**Files:** `evil.txt`
```

**Tests:** `test_alpha`
**Commit:** add alpha
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 15.1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 15: field-looking line inside a fenced block is not parsed as real field data" || fail "case 15: rc=$RC out=$OUT"

# ===========================================================================
# Case 16: the initial `git revert` inside `_commit_reverted` itself fails
# (conflict against dirty local state) -> the guard still leaves the
# worktree exactly as found: HEAD unchanged, no uncommitted or conflicted
# state left behind, rather than mid-conflict.
# ===========================================================================
new_repo
write_test_runner "$REPO"
write_tasks_md "$REPO" '### 16.1 Revert conflict

**Files:** `suite.txt`
**Tests:** `alpha_test`
**Baseline:** before=0 after=1
**Commit:** add alpha_test
**Build:** green
'
git -C "$REPO" add openspec/changes/change-a/tasks.md
git -C "$REPO" commit -q -m "plan"
printf 'alpha_test\n' >> "$REPO/suite.txt"
git -C "$REPO" add suite.txt
git -C "$REPO" commit -q -m "add alpha_test"
SHA="$(git -C "$REPO" rev-parse HEAD)"
# Dirty the same line the commit touched, uncommitted, so the revert this
# guard attempts cannot cleanly apply and fails with a conflict.
printf 'alpha_test_MODIFIED_LOCALLY\n' > "$REPO/suite.txt"
run_guard "$REPO" 16.1 "$SHA"
[ "$RC" -ne 0 ] && pass "case 16: guard reports non-zero when the revert itself fails" || fail "case 16: rc=$RC out=$OUT (expected non-zero)"
[ "$(git -C "$REPO" rev-parse HEAD)" = "$SHA" ] && pass "case 16: HEAD unchanged after a failed revert" || fail "case 16: HEAD moved, expected $SHA got $(git -C "$REPO" rev-parse HEAD)"
git -C "$REPO" diff --quiet && git -C "$REPO" diff --cached --quiet && pass "case 16: worktree clean after a failed revert, not left mid-conflict" || fail "case 16: worktree left dirty/conflicted"
[ -z "$(git -C "$REPO" status --porcelain=v1 2>/dev/null | grep '^U')" ] && pass "case 16: no unmerged/conflicted paths left behind" || fail "case 16: unmerged paths remain"

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
