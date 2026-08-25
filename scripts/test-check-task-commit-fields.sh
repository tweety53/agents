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

# new_repo [change-name] -> sets REPO to a fresh throwaway git repo with one
# root commit, so every case's task commit has a real parent to diff
# against. change-name defaults to "change-a"; cases 17-23 pass a real
# change name because check_commit_scope derives the change name from the
# fixture's own directory, per tasks.md task 1.
new_repo() {
  CHANGE_NAME="${1:-change-a}"
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/task-commit-fields-test.XXXXXX")"
  git -C "$REPO" init -q
  git -C "$REPO" config user.email "test@example.com"
  git -C "$REPO" config user.name "Test"
  mkdir -p "$REPO/spectre/changes/$CHANGE_NAME"
  printf 'root\n' > "$REPO/root.txt"
  git -C "$REPO" add root.txt
  git -C "$REPO" commit -q -m "root"
}

# write_tasks_md <repo> <content> -> writes the change's tasks.md
write_tasks_md() {
  printf '%s' "$2" > "$1/spectre/changes/$CHANGE_NAME/tasks.md"
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
write_tasks_md "$REPO" '- [ ] 1. Clean task

**Files:** `alpha.txt`, `beta.txt`
**Tests:** `test_alpha`
**Commit:** add alpha and beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf '# test_alpha covers alpha\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 1: files subset of declared passes" || fail "case 1: rc=$RC out=$OUT"

# ===========================================================================
# Case 2: commit touches a file not in declared Files: -> exit 1.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 2. Undeclared file

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha only
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf '# test_alpha covers alpha\n' > "$REPO/alpha.txt"
printf 'gamma\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt gamma.txt
git -C "$REPO" commit -q -m "add alpha only"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 1 ] && pass "case 2: undeclared file fails" || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"gamma.txt"*"not declared"*) pass "case 2: names the undeclared file" ;;
  *) fail "case 2: expected message naming gamma.txt as undeclared, out=$OUT" ;;
esac

# ===========================================================================
# Case 3: declared test name found in the commit's diff -> exit 0.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 3. Test present

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 3 "$SHA"
[ "$RC" -eq 0 ] && pass "case 3: declared test name found in diff passes" || fail "case 3: rc=$RC out=$OUT"

# ===========================================================================
# Case 4: declared test name missing from the commit's diff -> exit 1.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 4. Test missing

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'no tests here\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 4 "$SHA"
[ "$RC" -eq 1 ] && pass "case 4: missing declared test fails" || fail "case 4: rc=$RC out=$OUT"
case "$OUT" in
  *"test_alpha"*"not found in the diff"*) pass "case 4: names the missing test" ;;
  *) fail "case 4: expected message naming test_alpha as missing, out=$OUT" ;;
esac

# ===========================================================================
# Case 5: commit subject matches declared Commit: -> exit 0.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 5. Subject matches

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha for real
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha for real"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 5 "$SHA"
[ "$RC" -eq 0 ] && pass "case 5: commit subject matches declared Commit: passes" || fail "case 5: rc=$RC out=$OUT"

# ===========================================================================
# Case 6: commit subject does not match declared Commit: -> exit 1.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 6. Subject mismatch

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha for real
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha, not quite right"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 6 "$SHA"
[ "$RC" -eq 1 ] && pass "case 6: commit subject mismatch fails" || fail "case 6: rc=$RC out=$OUT"
case "$OUT" in
  *"subject"*"does not match"*) pass "case 6: reports the subject mismatch" ;;
  *) fail "case 6: expected message reporting subject mismatch, out=$OUT" ;;
esac

# ===========================================================================
# Case 7: extra path covered by Allowed-collateral: glob -> exit 0.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 7. Collateral covered

**Files:** `alpha.txt`
**Allowed-collateral:** `docs/*.md`
**Tests:** `test_alpha`
**Commit:** add alpha and sweep docs
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
mkdir -p "$REPO/docs"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
printf 'swept\n' > "$REPO/docs/notes.md"
git -C "$REPO" add alpha.txt docs/notes.md
git -C "$REPO" commit -q -m "add alpha and sweep docs"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 7 "$SHA"
[ "$RC" -eq 0 ] && pass "case 7: extra path covered by Allowed-collateral: glob passes" || fail "case 7: rc=$RC out=$OUT"

# ===========================================================================
# Case 8: Tests: written as real free prose naming cases by number (the
# shape every real Tests: field in this plan's own tasks.md actually uses,
# e.g. "Case 1: files subset of declared passes; Case 2: undeclared file
# fails; ..."), with both cases' "# Case N:" markers present in the diff ->
# exit 0. Proves the guard does not require backtick-quoted identifiers.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 8. Prose cases present

**Files:** `guard_test.sh`
**Tests:** Case 1: files subset of declared passes; Case 2: undeclared file
fails
**Commit:** add guard test cases
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
{
  printf '# Case 1: files subset of declared passes\n'
  printf '# Case 2: undeclared file fails\n'
} > "$REPO/guard_test.sh"
git -C "$REPO" add guard_test.sh
git -C "$REPO" commit -q -m "add guard test cases"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 8 "$SHA"
[ "$RC" -eq 0 ] && pass "case 8: prose Case N: labels found in diff passes" || fail "case 8: rc=$RC out=$OUT"

# ===========================================================================
# Case 9: same prose shape as case 8, but the diff only carries "# Case 1:",
# never "Case 2" -> exit 1, naming "Case 2" (not the whole sentence).
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 9. Prose case missing

**Files:** `guard_test.sh`
**Tests:** Case 1: files subset of declared passes; Case 2: undeclared file
fails
**Commit:** add guard test case one only
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf '# Case 1: files subset of declared passes\n' > "$REPO/guard_test.sh"
git -C "$REPO" add guard_test.sh
git -C "$REPO" commit -q -m "add guard test case one only"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 9 "$SHA"
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
write_tasks_md "$REPO" '- [ ] 10. No checkable tests declared

**Files:** `guard_test.sh`
**Tests:** the 7 cases listed in task 3.2, run for the first time against
the wrapper this task adds
**Commit:** add the wrapper
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'wrapper body, no case markers at all\n' > "$REPO/guard_test.sh"
git -C "$REPO" add guard_test.sh
git -C "$REPO" commit -q -m "add the wrapper"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 10 "$SHA"
[ "$RC" -eq 0 ] && pass "case 10: prose with no Case N: or backticks declares nothing, never false-fails" || fail "case 10: rc=$RC out=$OUT"

# ===========================================================================
# Case 11: Regression: passes — reverting the task commit makes its
# declared test fail (RESULT alpha_test: fail), and un-reverting restores
# suite.txt to passing state and HEAD to the task commit.
# ===========================================================================
new_repo
write_test_runner "$REPO"
write_tasks_md "$REPO" '- [ ] 11. Regression passes

**Files:** `suite.txt`
**Tests:** `alpha_test`
**Baseline:** not applicable — covered by case 13
**Commit:** add alpha_test
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'alpha_test\n' >> "$REPO/suite.txt"
git -C "$REPO" add suite.txt
git -C "$REPO" commit -q -m "add alpha_test"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 11 "$SHA"
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
write_tasks_md "$REPO" '- [ ] 12. Regression skip

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 12 "$SHA"
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
write_tasks_md "$REPO" '- [ ] 13. Baseline passes

**Files:** `suite.txt`
**Tests:** `alpha_test`
**Baseline:** before=0 after=1
**Commit:** add alpha_test
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'alpha_test\n' >> "$REPO/suite.txt"
git -C "$REPO" add suite.txt
git -C "$REPO" commit -q -m "add alpha_test"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 13 "$SHA"
[ "$RC" -eq 0 ] && pass "case 13: baseline check passes (parent/task counts match declared before=/after=)" || fail "case 13: rc=$RC out=$OUT"
[ "$(git -C "$REPO" rev-parse HEAD)" = "$SHA" ] && pass "case 13: HEAD unchanged after baseline check" || fail "case 13: HEAD moved"

# ===========================================================================
# Case 14: Baseline: skips (not fails) when the ## test command's output
# carries no parseable COUNT: line.
# ===========================================================================
new_repo
write_unsupported_test_runner "$REPO"
write_tasks_md "$REPO" '- [ ] 14. Baseline skip

**Files:** `alpha.txt`
**Baseline:** before=0 after=1
**Commit:** add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'no tests here\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 14 "$SHA"
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
write_tasks_md "$REPO" '- [ ] 15. Fence guard

**Files:** `alpha.txt`

Example of the field grammar, not a real field:

```
**Files:** `evil.txt`
```

**Tests:** `test_alpha`
**Commit:** add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 15 "$SHA"
[ "$RC" -eq 0 ] && pass "case 15: field-looking line inside a fenced block is not parsed as real field data" || fail "case 15: rc=$RC out=$OUT"

# ===========================================================================
# Case 16: the initial `git revert` inside `_commit_reverted` itself fails
# (conflict against dirty local state) -> the guard still leaves the
# worktree exactly as found: HEAD unchanged, no uncommitted or conflicted
# state left behind, rather than mid-conflict.
# ===========================================================================
new_repo
write_test_runner "$REPO"
write_tasks_md "$REPO" '- [ ] 16. Revert conflict

**Files:** `suite.txt`
**Tests:** `alpha_test`
**Baseline:** before=0 after=1
**Commit:** add alpha_test
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'alpha_test\n' >> "$REPO/suite.txt"
git -C "$REPO" add suite.txt
git -C "$REPO" commit -q -m "add alpha_test"
SHA="$(git -C "$REPO" rev-parse HEAD)"
# Dirty the same line the commit touched, uncommitted, so the revert this
# guard attempts cannot cleanly apply and fails with a conflict.
printf 'alpha_test_MODIFIED_LOCALLY\n' > "$REPO/suite.txt"
run_guard "$REPO" 16 "$SHA"
[ "$RC" -ne 0 ] && pass "case 16: guard reports non-zero when the revert itself fails" || fail "case 16: rc=$RC out=$OUT (expected non-zero)"
[ "$(git -C "$REPO" rev-parse HEAD)" = "$SHA" ] && pass "case 16: HEAD unchanged after a failed revert" || fail "case 16: HEAD moved, expected $SHA got $(git -C "$REPO" rev-parse HEAD)"
git -C "$REPO" diff --quiet && git -C "$REPO" diff --cached --quiet && pass "case 16: worktree clean after a failed revert, not left mid-conflict" || fail "case 16: worktree left dirty/conflicted"
[ -z "$(git -C "$REPO" status --porcelain=v1 2>/dev/null | grep '^U')" ] && pass "case 16: no unmerged/conflicted paths left behind" || fail "case 16: unmerged paths remain"

# ===========================================================================
# Case 17: declared Commit: scope equals the change name -> exit 1, naming
# the task and the offending scope. The real commit's subject matches the
# declared field exactly, so the failure can only come from the new scope
# check, never from check_commit_subject.
# ===========================================================================
new_repo "kan-900-some-change"
write_tasks_md "$REPO" '- [ ] 17. Change-name scope

**Files:** `alpha.txt`
**Commit:** feat(kan-900-some-change): add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(kan-900-some-change): add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 17 "$SHA"
[ "$RC" -eq 1 ] && pass "case 17: change-name scope fails" || fail "case 17: rc=$RC out=$OUT"
case "$OUT" in
  *"17"*"kan-900-some-change"*) pass "case 17: message names the task and the offending scope" ;;
  *) fail "case 17: expected message naming task 17 and scope kan-900-some-change, out=$OUT" ;;
esac

# ===========================================================================
# Case 18: declared Commit: scope equals the change name's bare Jira key
# -> exit 1.
# ===========================================================================
new_repo "kan-900-some-change"
write_tasks_md "$REPO" '- [ ] 18. Bare-key scope

**Files:** `alpha.txt`
**Commit:** feat(kan-900): add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(kan-900): add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 18 "$SHA"
[ "$RC" -eq 1 ] && pass "case 18: bare-key scope fails" || fail "case 18: rc=$RC out=$OUT"
case "$OUT" in
  *"18"*"kan-900"*) pass "case 18: message names the task and the offending scope" ;;
  *) fail "case 18: expected message naming task 18 and scope kan-900, out=$OUT" ;;
esac

# ===========================================================================
# Case 19: declared Commit: scope is a numeric task id -> exit 1.
# ===========================================================================
new_repo "kan-900-some-change"
write_tasks_md "$REPO" '- [ ] 19. Numeric task id scope

**Files:** `alpha.txt`
**Commit:** feat(3): add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(3): add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 19 "$SHA"
[ "$RC" -eq 1 ] && pass "case 19: numeric task id scope fails" || fail "case 19: rc=$RC out=$OUT"
case "$OUT" in
  *"19"*"task id"*) pass "case 19: message names the task and reports the task-id shape" ;;
  *) fail "case 19: expected message naming task 19 and a task-id scope, out=$OUT" ;;
esac

# ===========================================================================
# Case 20: declared Commit: scope is a dotted task id -> exit 1.
# ===========================================================================
new_repo "kan-900-some-change"
write_tasks_md "$REPO" '- [ ] 20. Dotted task id scope

**Files:** `alpha.txt`
**Commit:** feat(3.2): add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(3.2): add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 20 "$SHA"
[ "$RC" -eq 1 ] && pass "case 20: dotted task id scope fails" || fail "case 20: rc=$RC out=$OUT"
case "$OUT" in
  *"20"*"task id"*) pass "case 20: message names the task and reports the task-id shape" ;;
  *) fail "case 20: expected message naming task 20 and a task-id scope, out=$OUT" ;;
esac

# ===========================================================================
# Case 21: declared Commit: scope names a real module -> exit 0.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 21. Module scope

**Files:** `alpha.txt`
**Commit:** feat(scripts): add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(scripts): add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 21 "$SHA"
[ "$RC" -eq 0 ] && pass "case 21: module scope passes" || fail "case 21: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 21: clean exit, no scope violation printed" || fail "case 21: expected no output, got: $OUT"

# ===========================================================================
# Case 22: declared Commit: field carries no scope at all -> exit 0, a
# scope is optional.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 22. No scope

**Files:** `alpha.txt`
**Commit:** feat: add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 22 "$SHA"
[ "$RC" -eq 0 ] && pass "case 22: absent scope passes" || fail "case 22: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 22: clean exit, no scope violation printed" || fail "case 22: expected no output, got: $OUT"

# ===========================================================================
# Case 23: declared Commit: scope merely contains the change's Jira key as
# a substring -> exit 0. Proves the check is equality, not substring.
# ===========================================================================
new_repo "kan-900-some-change"
write_tasks_md "$REPO" '- [ ] 23. Scope containing the key

**Files:** `alpha.txt`
**Commit:** feat(kan-900-helpers): add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(kan-900-helpers): add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 23 "$SHA"
[ "$RC" -eq 0 ] && pass "case 23: scope merely containing the key passes (equality, not substring)" || fail "case 23: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 23: clean exit, no scope violation printed" || fail "case 23: expected no output, got: $OUT"

# ===========================================================================
# Case 24: declared Commit: scope merely contains the change name as a
# substring -> exit 0. Proves the change-name check is equality, not
# substring.
# ===========================================================================
new_repo "kan-900-some-change"
write_tasks_md "$REPO" '- [ ] 24. Scope containing the change name

**Files:** `alpha.txt`
**Commit:** feat(kan-900-some-change-helpers): add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(kan-900-some-change-helpers): add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 24 "$SHA"
[ "$RC" -eq 0 ] && pass "case 24: scope merely containing the change name passes (equality, not substring)" || fail "case 24: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 24: clean exit, no scope violation printed" || fail "case 24: expected no output, got: $OUT"

# ===========================================================================
# Case 25 (pass 2, finding A): the Conventional Commits breaking-change
# subject form `<type>(<scope>)!:` names the change as its scope -> exit 1,
# same as the plain `:` form. Before the fix, SUBJECT_SCOPE_RE did not match
# a subject carrying `!` before the colon at all, so check_commit_scope
# silently returned no violation for this exact subject.
# ===========================================================================
new_repo "kan-900-some-change"
write_tasks_md "$REPO" '- [ ] 25. Breaking-change form names the change

**Files:** `alpha.txt`
**Commit:** feat(kan-900-some-change)!: add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(kan-900-some-change)!: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 25 "$SHA"
[ "$RC" -eq 1 ] && pass "case 25: breaking-change '!' form still catches a change-name scope" || fail "case 25: rc=$RC out=$OUT"
case "$OUT" in
  *"25"*"names the change"*) pass "case 25: message reports the change-name shape" ;;
  *) fail "case 25: expected message naming task 25 and the change-name shape, out=$OUT" ;;
esac

# ===========================================================================
# Case 26 (pass 2, finding A, sanity): the breaking-change '!' form with a
# real module scope still passes -> exit 0. Proves the optional '!' did not
# turn every subject into a scope match.
# ===========================================================================
new_repo "kan-900-some-change"
write_tasks_md "$REPO" '- [ ] 26. Breaking-change form, real module

**Files:** `alpha.txt`
**Commit:** feat(scripts)!: add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(scripts)!: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 26 "$SHA"
[ "$RC" -eq 0 ] && pass "case 26: breaking-change '!' form with a real module scope passes" || fail "case 26: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 26: clean exit, no scope violation printed" || fail "case 26: expected no output, got: $OUT"

# ===========================================================================
# Case 27 (pass 2, finding B): declared Commit: scope names the change in
# UPPERCASE -> exit 1. Proves the change-name comparison is
# case-insensitive.
# ===========================================================================
new_repo "kan-900-some-change"
write_tasks_md "$REPO" '- [ ] 27. Uppercase change-name scope

**Files:** `alpha.txt`
**Commit:** feat(KAN-900-SOME-CHANGE): add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(KAN-900-SOME-CHANGE): add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 27 "$SHA"
[ "$RC" -eq 1 ] && pass "case 27: uppercase change-name scope fails (case-insensitive compare)" || fail "case 27: rc=$RC out=$OUT"
case "$OUT" in
  *"27"*"names the change"*) pass "case 27: message reports the change-name shape" ;;
  *) fail "case 27: expected message naming task 27 and the change-name shape, out=$OUT" ;;
esac

# ===========================================================================
# Case 28 (pass 2, finding B): declared Commit: scope names the change's
# bare Jira key in UPPERCASE -> exit 1. Jira keys are conventionally
# written uppercase, so this is the shape a human is most likely to type.
# ===========================================================================
new_repo "kan-900-some-change"
write_tasks_md "$REPO" '- [ ] 28. Uppercase Jira-key scope

**Files:** `alpha.txt`
**Commit:** feat(KAN-900): add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(KAN-900): add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 28 "$SHA"
[ "$RC" -eq 1 ] && pass "case 28: uppercase Jira-key scope fails (case-insensitive compare)" || fail "case 28: rc=$RC out=$OUT"
case "$OUT" in
  *"28"*"Jira key"*) pass "case 28: message reports the Jira-key shape" ;;
  *) fail "case 28: expected message naming task 28 and the Jira-key shape, out=$OUT" ;;
esac

# ===========================================================================
# Case 29 (pass 2, finding C): a change name carrying TWO plausible
# <letters>-<digits> Jira-key-shaped segments (`release-2026-kan-450-
# cleanup`) has no unambiguous leading key at all, per _leading_jira_key's
# fixed definition -> a task scoped to the first candidate passes rather
# than being wrongly flagged as the change's Jira key.
# ===========================================================================
new_repo "release-2026-kan-450-cleanup"
write_tasks_md "$REPO" '- [ ] 29. Ambiguous key-shaped change name

**Files:** `alpha.txt`
**Commit:** feat(release-2026): add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat(release-2026): add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 29 "$SHA"
[ "$RC" -eq 0 ] && pass "case 29: ambiguous key-shaped change name yields no leading key, scope passes" || fail "case 29: rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "case 29: clean exit, no scope violation printed" || fail "case 29: expected no output, got: $OUT"

# ===========================================================================
# Case 30: a folded red task resolves to its partner's commit -> exit 0.
# The fixture is the shape the pipeline actually produces for a `Build: red`
# task: task 1's commit is folded into task 2's, so ONE commit carries both
# tasks' files and task 2's declared subject, and task 1's own declared
# subject exists nowhere. Invoked for the RED task, the guard must take the
# partner's subject and the union of both file sets.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red half

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2

- [ ] 2. Green half

**Files:** `beta.txt`
**Commit:** feat: add alpha and beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "feat: add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 30: folded red task passes on its partner's subject" || fail "case 30: rc=$RC out=$OUT"
case "$OUT" in
  *"does not match declared Commit"*) fail "case 30: red task's own declared subject was still required, out=$OUT" ;;
  *) pass "case 30: the red task's own declared subject is not required" ;;
esac
case "$OUT" in
  *"not declared in Files:"*) fail "case 30: partner's files reported as undeclared collateral, out=$OUT" ;;
  *) pass "case 30: the file set is the union of both tasks" ;;
esac

# ===========================================================================
# Case 31: the same folded commit, invoked for the PARTNER -> exit 0, so
# either task id gives one verdict. The red task's files must not be
# reported as the partner's undeclared collateral either.
# ===========================================================================
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 0 ] && pass "case 31: the partner reaches the same verdict against the folded commit" || fail "case 31: rc=$RC out=$OUT"
case "$OUT" in
  *"not declared in Files:"*) fail "case 31: red task's files reported as undeclared collateral, out=$OUT" ;;
  *) pass "case 31: the union holds when invoked for the partner too" ;;
esac

# ===========================================================================
# Case 32: `Squash-with:` names a task absent from the same plan -> exit 1,
# naming the missing partner. Resolving against a missing partner would
# silently widen the file set to nothing.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red half pointing at nothing

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 9

- [ ] 2. Green half

**Files:** `beta.txt`
**Commit:** feat: add alpha and beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "feat: add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 32: a Squash-with naming a missing partner fails" || fail "case 32: rc=$RC out=$OUT"
case "$OUT" in
  *"Task 9"*) pass "case 32: names the missing partner" ;;
  *) fail "case 32: expected message naming Task 9, out=$OUT" ;;
esac

# ===========================================================================
# Case 33: the named partner is itself `red` -> exit 1, for the same reason
# check-task-build-green.py already fails it: two red halves fold into no
# green commit at all.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red half

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2

- [ ] 2. Also red

**Files:** `beta.txt`
**Commit:** feat: add alpha and beta
**Build:** red

**Squash-with:** Task 1
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "feat: add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 33: a partner that is itself red fails" || fail "case 33: rc=$RC out=$OUT"
case "$OUT" in
  *"itself red"*) pass "case 33: message reports the red partner" ;;
  *) fail "case 33: expected message reporting the red partner, out=$OUT" ;;
esac

# ===========================================================================
# Case 34: a plan whose tasks carry no `Squash-with:` at all is unaffected —
# a commit touching another task's declared file is still undeclared
# collateral. This is the case that stops the resolution being written as an
# unconditional union across every task in the plan.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Ordinary task

**Files:** `alpha.txt`
**Commit:** feat: add alpha
**Build:** green

- [ ] 2. Another ordinary task

**Files:** `beta.txt`
**Commit:** feat: add beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "feat: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 34: without Squash-with, another task's file is still undeclared" || fail "case 34: rc=$RC out=$OUT"
case "$OUT" in
  *"beta.txt"*"not declared"*) pass "case 34: names the other task's file as undeclared" ;;
  *) fail "case 34: expected message naming beta.txt as undeclared, out=$OUT" ;;
esac

# ===========================================================================
# Case 35: the same, for the subject — without `Squash-with:`, a task whose
# commit carries ANOTHER task's declared subject still fails. Pins that the
# partner's subject is taken only for a folded pair.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Ordinary task

**Files:** `alpha.txt`
**Commit:** feat: add alpha
**Build:** green

- [ ] 2. Another ordinary task

**Files:** `beta.txt`
**Commit:** feat: add beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat: add beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 35: without Squash-with, another task's subject is still a mismatch" || fail "case 35: rc=$RC out=$OUT"
case "$OUT" in
  *"does not match declared Commit"*) pass "case 35: reports the subject mismatch" ;;
  *) fail "case 35: expected a subject mismatch message, out=$OUT" ;;
esac

# ===========================================================================
# Case 36 (fix round 1, F2/F3): a red task whose `Squash-with:` names TWO
# green partners. The whole unit folds into ONE commit, so all three tasks'
# files land in it and both partners declare the SAME subject. Every one of
# the three ids must reach the same clean verdict — including a green
# partner asked about its SIBLING partner's file, which is the case the
# green side used never to union.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red third

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3

- [ ] 2. Green third

**Files:** `beta.txt`
**Commit:** feat: add alpha beta and gamma
**Build:** green

- [ ] 3. Sibling green third

**Files:** `gamma.txt`
**Commit:** feat: add alpha beta and gamma
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
printf 'g\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt beta.txt gamma.txt
git -C "$REPO" commit -q -m "feat: add alpha beta and gamma"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 36: the red task unions both partners' files" || fail "case 36: rc=$RC out=$OUT"
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 0 ] && pass "case 36: a green partner unions the red task's and its sibling's files" || fail "case 36 (task 2): rc=$RC out=$OUT"
case "$OUT" in
  *"gamma.txt"*"not declared"*) fail "case 36: sibling partner's file reported as undeclared collateral, out=$OUT" ;;
  *) pass "case 36: the sibling partner's file is inside the union" ;;
esac
run_guard "$REPO" 3 "$SHA"
[ "$RC" -eq 0 ] && pass "case 36: the sibling partner reaches the same verdict" || fail "case 36 (task 3): rc=$RC out=$OUT"
case "$OUT" in
  *"beta.txt"*"not declared"*) fail "case 36: first partner's file reported as undeclared collateral, out=$OUT" ;;
  *) pass "case 36: the first partner's file is inside the union too" ;;
esac

# ===========================================================================
# Case 37 (fix round 1, F3): two partners of the same fold declaring
# DIFFERENT `Commit:` subjects is a plan defect, not a subject mismatch.
# One unit folds into one commit, so one subject; the guard must report the
# disagreement rather than silently taking the first-listed partner's
# subject and blaming the commit for not matching it. The fixture's real
# subject is the LATER-listed partner's, which is exactly the shape that
# produced a false "does not match declared Commit:" before.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red third

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3

- [ ] 2. Green third

**Files:** `beta.txt`
**Commit:** feat: add beta
**Build:** green

- [ ] 3. Sibling green third

**Files:** `gamma.txt`
**Commit:** feat: add alpha beta and gamma
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
printf 'g\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt beta.txt gamma.txt
git -C "$REPO" commit -q -m "feat: add alpha beta and gamma"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 37: partners declaring different subjects fail" || fail "case 37: rc=$RC out=$OUT"
case "$OUT" in
  *"different Commit:"*) pass "case 37: reports the disagreement between the partners" ;;
  *) fail "case 37: expected a partner-disagreement message, out=$OUT" ;;
esac
case "$OUT" in
  *"does not match declared Commit"*) fail "case 37: blamed the commit instead of the plan, out=$OUT" ;;
  *) pass "case 37: the commit is not blamed for a plan defect" ;;
esac

# ===========================================================================
# Case 38 (fix round 1, F4): a `Squash-with:` naming a missing partner fails
# when the guard is invoked for a GREEN partner of the same fold, not only
# for the red task itself — "either task id gives the same verdict" covers
# the invalid folds too.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red third

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 9

- [ ] 2. Green third

**Files:** `beta.txt`
**Commit:** feat: add alpha and beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "feat: add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 1 ] && pass "case 38: a missing partner fails when checked via a green partner" || fail "case 38: rc=$RC out=$OUT"
case "$OUT" in
  *"Task 9"*) pass "case 38: names the missing partner from the green side" ;;
  *) fail "case 38: expected message naming Task 9, out=$OUT" ;;
esac

# ===========================================================================
# Case 39 (fix round 1, F4): the same, for a second partner that is itself
# `red` — the green side re-validates the whole partner set, so a fold that
# the red task's own id rejects cannot pass through a green partner's id.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red third

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3

- [ ] 2. Green third

**Files:** `beta.txt`
**Commit:** feat: add alpha and beta
**Build:** green

- [ ] 3. Also red

**Files:** `gamma.txt`
**Commit:** test: add gamma
**Build:** red

**Squash-with:** Task 1
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
printf 'g\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt beta.txt gamma.txt
git -C "$REPO" commit -q -m "feat: add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 1 ] && pass "case 39: a red partner fails when checked via a green partner" || fail "case 39: rc=$RC out=$OUT"
case "$OUT" in
  *"itself red"*) pass "case 39: reports the red partner from the green side" ;;
  *) fail "case 39: expected the red-partner message, out=$OUT" ;;
esac

# ===========================================================================
# Case 40 (fix round 2, F6): a fold in which ONE partner declares no
# `Commit:` field at all. A partner that declares nothing contributes no
# subject, so this is NOT a disagreement — only one subject was ever
# declared, and every id must pass on it. Before this round the undeclared
# field was collected as a distinct subject and the fold failed with
# "different Commit: subjects '...', None".
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red third

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3

- [ ] 2. Green third

**Files:** `beta.txt`
**Commit:** feat: add alpha beta and gamma
**Build:** green

- [ ] 3. Sibling green third declaring no subject

**Files:** `gamma.txt`
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
printf 'g\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt beta.txt gamma.txt
git -C "$REPO" commit -q -m "feat: add alpha beta and gamma"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 40: an undeclared Commit: on one partner is not a disagreement" || fail "case 40: rc=$RC out=$OUT"
case "$OUT" in
  *"different Commit:"*) fail "case 40: reported a false disagreement against the undeclared field, out=$OUT" ;;
  *) pass "case 40: no false disagreement is reported" ;;
esac
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 0 ] && pass "case 40: the declaring partner reaches the same verdict" || fail "case 40 (task 2): rc=$RC out=$OUT"
run_guard "$REPO" 3 "$SHA"
[ "$RC" -eq 0 ] && pass "case 40: the non-declaring partner reaches the same verdict" || fail "case 40 (task 3): rc=$RC out=$OUT"

# ===========================================================================
# Case 41 (fix round 2, F5): the same fixture with a commit subject that
# matches NO declared subject. The one subject the fold declares is the
# subject the surviving commit is checked against — from every id, the
# non-declaring partner's included. Before this round that partner's id
# checked the commit against `None`, i.e. against nothing, and passed.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red third

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3

- [ ] 2. Green third

**Files:** `beta.txt`
**Commit:** feat: add alpha beta and gamma
**Build:** green

- [ ] 3. Sibling green third declaring no subject

**Files:** `gamma.txt`
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
printf 'g\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt beta.txt gamma.txt
git -C "$REPO" commit -q -m "chore: something else entirely"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 41: the red id fails on the fold's one declared subject" || fail "case 41: rc=$RC out=$OUT"
case "$OUT" in
  *"does not match declared Commit"*) pass "case 41: the red id reports a subject mismatch" ;;
  *) fail "case 41: expected a subject mismatch from the red id, out=$OUT" ;;
esac
run_guard "$REPO" 3 "$SHA"
[ "$RC" -eq 1 ] && pass "case 41: the non-declaring partner's id fails too" || fail "case 41 (task 3): rc=$RC out=$OUT"
case "$OUT" in
  *"does not match declared Commit"*) pass "case 41: the non-declaring partner checks against the fold's subject, not nothing" ;;
  *) fail "case 41: expected a subject mismatch from task 3, out=$OUT" ;;
esac

# ===========================================================================
# Case 42 (fix round 2, F5): a fold in which NO partner declares a
# `Commit:` field. There is then no subject for the surviving commit to be
# checked against, which is a plan defect the guard reports — naming the
# RED task, where the defective field is — rather than silently vouching
# for a commit it checked against nothing. Both ids give the one verdict.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red half

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2

- [ ] 2. Green half declaring no subject

**Files:** `beta.txt`
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "chore: an entirely unrelated subject"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 42: a fold declaring no subject at all fails" || fail "case 42: rc=$RC out=$OUT"
case "$OUT" in
  *"no partner named by Squash-with: declares a Commit: subject"*) pass "case 42: reports that nothing declares the folded commit's subject" ;;
  *) fail "case 42: expected the no-declared-subject message, out=$OUT" ;;
esac
case "$OUT" in
  *"task 1:"*) pass "case 42: names the red task, where the defective field is" ;;
  *) fail "case 42: expected the message to name task 1, out=$OUT" ;;
esac
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 1 ] && pass "case 42: the green partner's id gives the same verdict" || fail "case 42 (task 2): rc=$RC out=$OUT"
case "$OUT" in
  *"no partner named by Squash-with: declares a Commit: subject"*) pass "case 42: the green side reports the same defect" ;;
  *) fail "case 42: expected the no-declared-subject message from task 2, out=$OUT" ;;
esac

# ===========================================================================
# Case 43 (fix round 2, F5 scope): an ORDINARY task — no `Squash-with:`,
# named by no red task — declaring no `Commit:` field keeps passing. The
# field is optional outside a fold, and tightening that would reach every
# task in every plan rather than the folded ones this change is about.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Ordinary task declaring no subject

**Files:** `alpha.txt`
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "chore: any subject at all"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 43: Commit: stays optional for an ordinary task" || fail "case 43: rc=$RC out=$OUT"

# ===========================================================================
# Case 44 (fix round 3, F8): TWO red tasks naming the SAME green partner.
# Both reds fold into that partner's commit, so all five tasks are one
# combined fold and one commit — and the two reds' partner sets declare
# DIFFERENT subjects. That is a plan defect, and every id in the combined
# fold must report it. Before this round the shared partner's id resolved
# against the FIRST red in document order only, took that fold's subject,
# and exited 0 on a commit both red ids rejected.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red one

**Files:** `one.txt`
**Commit:** test: add one
**Build:** red

**Squash-with:** Task 3, 4

- [ ] 2. Red two

**Files:** `two.txt`
**Commit:** test: add two
**Build:** red

**Squash-with:** Task 3, 5

- [ ] 3. Shared green partner

**Files:** `shared.txt`
**Build:** green

- [ ] 4. Green partner of red one

**Files:** `four.txt`
**Commit:** feat: subject a
**Build:** green

- [ ] 5. Green partner of red two

**Files:** `five.txt`
**Commit:** feat: subject b
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf '1\n' > "$REPO/one.txt"
printf '2\n' > "$REPO/two.txt"
printf 's\n' > "$REPO/shared.txt"
printf '4\n' > "$REPO/four.txt"
printf '5\n' > "$REPO/five.txt"
git -C "$REPO" add one.txt two.txt shared.txt four.txt five.txt
git -C "$REPO" commit -q -m "feat: subject a"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 3 "$SHA"
[ "$RC" -eq 1 ] && pass "case 44: the shared partner no longer passes a commit both red ids reject" || fail "case 44 (task 3): rc=$RC out=$OUT"
case "$OUT" in
  *"different Commit:"*) pass "case 44: the shared partner reports the folds' disagreement" ;;
  *) fail "case 44: expected a subject-disagreement message from task 3, out=$OUT" ;;
esac
case "$OUT" in
  *"Task 3"*) pass "case 44: the message names the shared task" ;;
  *) fail "case 44: expected the message to name the shared task, out=$OUT" ;;
esac
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 44: the first red id reaches the same verdict" || fail "case 44 (task 1): rc=$RC out=$OUT"
case "$OUT" in
  *"different Commit:"*) pass "case 44: the first red id reports the same disagreement" ;;
  *) fail "case 44: expected the disagreement message from task 1, out=$OUT" ;;
esac
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 1 ] && pass "case 44: the second red id reaches the same verdict" || fail "case 44 (task 2): rc=$RC out=$OUT"
case "$OUT" in
  *"different Commit:"*) pass "case 44: the second red id reports the same disagreement" ;;
  *) fail "case 44: expected the disagreement message from task 2, out=$OUT" ;;
esac

# ===========================================================================
# Case 45 (fix round 3, F8): the same two-red shape with the two folds
# AGREEING on the subject. One combined fold, one commit carrying all five
# tasks' files — so every id must union the WHOLE group, the sibling fold's
# files included, and pass. Before this round each red id saw only its own
# fold's files and failed on the other fold's as undeclared collateral,
# while the shared partner's id passed: three ids, two verdicts.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red one

**Files:** `one.txt`
**Commit:** test: add one
**Build:** red

**Squash-with:** Task 3, 4

- [ ] 2. Red two

**Files:** `two.txt`
**Commit:** test: add two
**Build:** red

**Squash-with:** Task 3, 5

- [ ] 3. Shared green partner

**Files:** `shared.txt`
**Build:** green

- [ ] 4. Green partner of red one

**Files:** `four.txt`
**Commit:** feat: the one folded subject
**Build:** green

- [ ] 5. Green partner of red two

**Files:** `five.txt`
**Commit:** feat: the one folded subject
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf '1\n' > "$REPO/one.txt"
printf '2\n' > "$REPO/two.txt"
printf 's\n' > "$REPO/shared.txt"
printf '4\n' > "$REPO/four.txt"
printf '5\n' > "$REPO/five.txt"
git -C "$REPO" add one.txt two.txt shared.txt four.txt five.txt
git -C "$REPO" commit -q -m "feat: the one folded subject"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 45: the first red id unions the sibling fold's files" || fail "case 45 (task 1): rc=$RC out=$OUT"
case "$OUT" in
  *"five.txt"*"not declared"*) fail "case 45: the sibling fold's file reported as undeclared collateral, out=$OUT" ;;
  *) pass "case 45: the sibling fold's file is inside the combined union" ;;
esac
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 0 ] && pass "case 45: the second red id reaches the same verdict" || fail "case 45 (task 2): rc=$RC out=$OUT"
run_guard "$REPO" 3 "$SHA"
[ "$RC" -eq 0 ] && pass "case 45: the shared partner reaches the same verdict" || fail "case 45 (task 3): rc=$RC out=$OUT"
run_guard "$REPO" 4 "$SHA"
[ "$RC" -eq 0 ] && pass "case 45: a partner of one fold sees the other fold's files too" || fail "case 45 (task 4): rc=$RC out=$OUT"
run_guard "$REPO" 5 "$SHA"
[ "$RC" -eq 0 ] && pass "case 45: and so does a partner of the other" || fail "case 45 (task 5): rc=$RC out=$OUT"

# ===========================================================================
# Case 46 (fix round 3, F7): the no-subject violation's precondition, pinned
# from both sides. It fires when the partners were otherwise valid (case 42
# pins that half against a real commit) and must NOT fire when a partner was
# missing or is itself red — those folds already carry their own violation
# and reporting a second one would name the same single defect twice. The
# guard expresses that precondition as an explicit partner-error flag rather
# than inferring it from "no other violation has been appended yet".
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red half naming a missing partner

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 9
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "chore: an entirely unrelated subject"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 46: a missing partner still fails" || fail "case 46: rc=$RC out=$OUT"
case "$OUT" in
  *"which does not exist in this plan"*) pass "case 46: reports the missing partner" ;;
  *) fail "case 46: expected the missing-partner message, out=$OUT" ;;
esac
case "$OUT" in
  *"no partner named by Squash-with: declares a Commit: subject"*) fail "case 46: reported the same single defect twice, out=$OUT" ;;
  *) pass "case 46: the no-subject message does not fire behind a missing partner" ;;
esac

new_repo
write_tasks_md "$REPO" '- [ ] 1. Red half naming a red partner

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2

- [ ] 2. Partner that is itself red and declares no subject

**Files:** `beta.txt`
**Build:** red
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "chore: an entirely unrelated subject"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 46: a partner that is itself red still fails" || fail "case 46 (red partner): rc=$RC out=$OUT"
case "$OUT" in
  *"which is itself red"*) pass "case 46: reports the red partner" ;;
  *) fail "case 46: expected the red-partner message, out=$OUT" ;;
esac
case "$OUT" in
  *"no partner named by Squash-with: declares a Commit: subject"*) fail "case 46: reported the same single defect twice behind a red partner, out=$OUT" ;;
  *) pass "case 46: the no-subject message does not fire behind a red partner" ;;
esac

# ===========================================================================
# Case 47 (fix round 4, item 1): the disagreement check's precondition,
# pinned the way case 46 pins the no-subject check's. A red task naming
# THREE partners — one absent from the plan, and two present, green and
# declaring DIFFERENT `Commit:` subjects. The partner set never resolved, so
# what the partners that did resolve happened to declare is not a second,
# separate defect: the run reports the missing partner and nothing else.
# Without the `partner_error` gate the same fixture reports two violations
# for one broken field.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red task naming a missing partner and two disagreeing ones

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2, 3, 9

- [ ] 2. Green partner declaring one subject

**Files:** `beta.txt`
**Commit:** feat: subject a
**Build:** green

- [ ] 3. Green partner declaring a different subject

**Files:** `gamma.txt`
**Commit:** feat: subject b
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
printf 'g\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt beta.txt gamma.txt
git -C "$REPO" commit -q -m "chore: an entirely unrelated subject"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 47: a missing partner still fails alongside disagreeing ones" || fail "case 47: rc=$RC out=$OUT"
case "$OUT" in
  *"which does not exist in this plan"*) pass "case 47: reports the missing partner" ;;
  *) fail "case 47: expected the missing-partner message, out=$OUT" ;;
esac
case "$OUT" in
  *"different Commit:"*) fail "case 47: reported the subject disagreement behind an unresolved partner set, out=$OUT" ;;
  *) pass "case 47: the disagreement message does not fire behind a missing partner" ;;
esac
case "$OUT" in
  *"no partner named by Squash-with: declares a Commit: subject"*) fail "case 47: reported the no-subject defect too, out=$OUT" ;;
  *) pass "case 47: the no-subject message does not fire behind a missing partner either" ;;
esac

# ===========================================================================
# Case 48 (fix round 4, item 2): the no-subject rule belongs to the COMBINED
# fold, not to one red's own partner list. Two reds share a green partner,
# so both fold into that partner's one commit — and only the SECOND red's
# partners declare a `Commit:` subject. One commit carries one subject, so
# that is the fold's subject and there is no defect; every id in the group
# must pass on it. Before this round the rule was enforced per red task and
# the first red's id failed with "no partner ... declares a Commit: subject"
# while a subject was sitting in the same fold.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red one, whose partners declare no subject

**Files:** `one.txt`
**Commit:** test: add one
**Build:** red

**Squash-with:** Task 3, 4

- [ ] 2. Red two, whose partner declares the fold subject

**Files:** `two.txt`
**Commit:** test: add two
**Build:** red

**Squash-with:** Task 3, 5

- [ ] 3. Shared green partner declaring no subject

**Files:** `shared.txt`
**Build:** green

- [ ] 4. Green partner of red one declaring no subject

**Files:** `four.txt`
**Build:** green

- [ ] 5. Green partner of red two declaring the fold subject

**Files:** `five.txt`
**Commit:** feat: the one folded subject
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf '1\n' > "$REPO/one.txt"
printf '2\n' > "$REPO/two.txt"
printf 's\n' > "$REPO/shared.txt"
printf '4\n' > "$REPO/four.txt"
printf '5\n' > "$REPO/five.txt"
git -C "$REPO" add one.txt two.txt shared.txt four.txt five.txt
git -C "$REPO" commit -q -m "feat: the one folded subject"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 48: a red whose own partners declare no subject takes the fold's" || fail "case 48 (task 1): rc=$RC out=$OUT"
case "$OUT" in
  *"no partner named by Squash-with: declares a Commit: subject"*) fail "case 48: the no-subject rule was applied per red task, out=$OUT" ;;
  *) pass "case 48: the no-subject rule is asked of the combined fold" ;;
esac
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 0 ] && pass "case 48: the declaring red id reaches the same verdict" || fail "case 48 (task 2): rc=$RC out=$OUT"
run_guard "$REPO" 3 "$SHA"
[ "$RC" -eq 0 ] && pass "case 48: the shared partner reaches the same verdict" || fail "case 48 (task 3): rc=$RC out=$OUT"
run_guard "$REPO" 4 "$SHA"
[ "$RC" -eq 0 ] && pass "case 48: the non-declaring partner reaches the same verdict" || fail "case 48 (task 4): rc=$RC out=$OUT"
run_guard "$REPO" 5 "$SHA"
[ "$RC" -eq 0 ] && pass "case 48: the declaring partner reaches the same verdict" || fail "case 48 (task 5): rc=$RC out=$OUT"

# ===========================================================================
# Case 48 (continued): the other half of the same rule. The same joined
# shape with NO red's partners declaring a subject anywhere in the group
# still fails — widening the question to the combined fold must not stop it
# being asked. The message is anchored on the shared task that joined the
# folds, the way case 44 anchors the disagreement, and every id reports it.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red one, whose partners declare no subject

**Files:** `one.txt`
**Commit:** test: add one
**Build:** red

**Squash-with:** Task 3, 4

- [ ] 2. Red two, whose partners declare no subject either

**Files:** `two.txt`
**Commit:** test: add two
**Build:** red

**Squash-with:** Task 3, 5

- [ ] 3. Shared green partner declaring no subject

**Files:** `shared.txt`
**Build:** green

- [ ] 4. Green partner of red one declaring no subject

**Files:** `four.txt`
**Build:** green

- [ ] 5. Green partner of red two declaring no subject

**Files:** `five.txt`
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf '1\n' > "$REPO/one.txt"
printf '2\n' > "$REPO/two.txt"
printf 's\n' > "$REPO/shared.txt"
printf '4\n' > "$REPO/four.txt"
printf '5\n' > "$REPO/five.txt"
git -C "$REPO" add one.txt two.txt shared.txt four.txt five.txt
git -C "$REPO" commit -q -m "chore: an entirely unrelated subject"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 48: a combined fold declaring no subject anywhere still fails" || fail "case 48 (no subject, task 1): rc=$RC out=$OUT"
case "$OUT" in
  *"no partner named by Squash-with: declares a Commit: subject"*) pass "case 48: reports that nothing in the fold declares the commit's subject" ;;
  *) fail "case 48: expected the no-declared-subject message, out=$OUT" ;;
esac
case "$OUT" in
  *"task 3:"*) pass "case 48: the message is anchored on the shared task" ;;
  *) fail "case 48: expected the message to name the shared task, out=$OUT" ;;
esac
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 1 ] && pass "case 48: the second red id reaches the same verdict" || fail "case 48 (no subject, task 2): rc=$RC out=$OUT"
run_guard "$REPO" 3 "$SHA"
[ "$RC" -eq 1 ] && pass "case 48: the shared partner reaches the same verdict" || fail "case 48 (no subject, task 3): rc=$RC out=$OUT"
run_guard "$REPO" 5 "$SHA"
[ "$RC" -eq 1 ] && pass "case 48: a non-declaring partner reaches the same verdict" || fail "case 48 (no subject, task 5): rc=$RC out=$OUT"

# ===========================================================================
# Case 49 (fix round 5, F9): the `Squash-with:` FIELD GATE. The value must be
# `Task <ids>` and nothing else. Without the gate, partner ids were extracted
# with a bare dotted-id pattern over the raw remainder of the line, so any
# digit run in free text became a partner id — and the fold's file union
# silently widened to that unrelated task's declared files, letting a commit
# smuggle an undeclared file past the guard. This is the exact fixture that
# reproduced it: `Task 3 (see step 2)` resolved to partners 3 AND 2, so
# `unrelated.txt` — a file no member of the real fold declares — was never
# flagged and the run exited 0.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red task whose Squash-with carries free text

**Files:** `a.txt`
**Commit:** test: add a
**Build:** red

**Squash-with:** Task 3 (see step 2)

- [ ] 2. An unrelated green task nobody folds with

**Files:** `unrelated.txt`
**Build:** green

- [ ] 3. The real partner

**Files:** `b.txt`
**Commit:** feat: add a and b
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/a.txt"
printf 'u\n' > "$REPO/unrelated.txt"
printf 'b\n' > "$REPO/b.txt"
git -C "$REPO" add a.txt unrelated.txt b.txt
git -C "$REPO" commit -q -m "feat: add a and b"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 49: a malformed Squash-with field fails instead of passing an undeclared file" || fail "case 49: rc=$RC out=$OUT"
case "$OUT" in
  *"Squash-with:"*"is not \`Task"*) pass "case 49: the message names the malformed field" ;;
  *) fail "case 49: expected the malformed-field message, out=$OUT" ;;
esac
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 1 ] && pass "case 49: the unrelated task is not dragged into the fold by its own id" || fail "case 49 (task 2): rc=$RC out=$OUT"

# ===========================================================================
# Case 50 (fix round 5, F9): the discriminating half of the same defect. The
# free text carries a digit that is NOT a task id, and the task that digit
# names does not exist in the plan. Before the gate, `Task 2 (see step 3)`
# resolved to partners 2 and 3 and the run reported "Task 3 ... does not
# exist in this plan" — proof that a word in prose had been read as a merge
# partner. After the gate the field is rejected as malformed, and no partner
# id is invented from its free text at all.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red task naming a partner and mentioning a step number

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2 (see step 3)

- [ ] 2. Green partner

**Files:** `beta.txt`
**Commit:** feat: add alpha and beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "feat: add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 50: free text in Squash-with is a malformed field, not a partner list" || fail "case 50: rc=$RC out=$OUT"
case "$OUT" in
  *"which does not exist in this plan"*) fail "case 50: a digit in free text was still read as a partner id, out=$OUT" ;;
  *) pass "case 50: no partner id is invented from the field's free text" ;;
esac
case "$OUT" in
  *"Squash-with:"*"is not \`Task"*) pass "case 50: the message names the malformed field instead" ;;
  *) fail "case 50: expected the malformed-field message, out=$OUT" ;;
esac

# ===========================================================================
# Case 51 (fix round 5, F9): the gate must not narrow what a WELL-FORMED
# field means. Whitespace-separated ids — `Task 2 3` — resolve exactly as
# `Task 2, 3` does (case 36), from every id in the fold. This is the case
# that stops the gate being written so tightly it only admits a single
# comma-separated list. It used to name its partners `2.1` and `3.4`; a task
# id is a flat integer now, and a dotted partner names no task at all, which
# check-task-build-green.sh pins as its own case 27.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red task with two partners

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2 3

- [ ] 2. Green partner

**Files:** `beta.txt`
**Commit:** feat: add alpha beta and gamma
**Build:** green

- [ ] 3. Sibling green partner

**Files:** `gamma.txt`
**Commit:** feat: add alpha beta and gamma
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
printf 'g\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt beta.txt gamma.txt
git -C "$REPO" commit -q -m "feat: add alpha beta and gamma"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 51: whitespace-separated partners still resolve" || fail "case 51: rc=$RC out=$OUT"
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 0 ] && pass "case 51: the first partner reaches the same verdict" || fail "case 51 (task 2): rc=$RC out=$OUT"
run_guard "$REPO" 3 "$SHA"
[ "$RC" -eq 0 ] && pass "case 51: the sibling partner reaches the same verdict" || fail "case 51 (task 3): rc=$RC out=$OUT"

# ===========================================================================
# Case 52 (fix round 6, F14): the `Squash-with:` field is LINE-SCOPED. Its
# value is what stands on its own `**Squash-with:**` line and nothing else,
# so a prose line following it without an intervening blank line is not part
# of the value. Fix round 5 gated the value AFTER the shared continuation
# join, which slurped that prose into the field and reported a well-formed
# fold as malformed — while check-task-build-green.py, which gates one
# physical line, accepted the very same plan. Both guards now read one
# grammar, from scripts/lib/plan_grammar.py.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red task whose Squash-with is followed by prose

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2
The fold is described in the paragraph above.

- [ ] 2. Green partner

**Files:** `beta.txt`
**Commit:** feat: add alpha and beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "feat: add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 52: an unblanked prose line is not slurped into the Squash-with value" || fail "case 52: rc=$RC out=$OUT"
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 0 ] && pass "case 52: the partner reaches the same verdict" || fail "case 52 (task 2): rc=$RC out=$OUT"

# ===========================================================================
# Case 53 (fix round 6, F14): the other direction of the same scoping rule.
# A partner list wrapped onto a second source line names ONE partner, not
# two — the continuation is outside the field. This is the case that stops
# the gate being re-applied to a joined value: under the joined reading this
# plan resolved partners 2 AND 3 and the commit passed, while
# check-task-build-green.py read partner 2 alone and validated only that
# one. One grammar, one partner list, from either guard; a wrapped field is
# a plan defect the commit check now reports rather than silently widening
# the fold's file union to task 3's declared file.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red task whose partner ids wrap onto a second line

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2,
3

- [ ] 2. Green partner

**Files:** `beta.txt`
**Commit:** feat: add alpha beta and gamma
**Build:** green

- [ ] 3. Sibling green partner named only on the wrapped line

**Files:** `gamma.txt`
**Commit:** feat: add alpha beta and gamma
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
printf 'g\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt beta.txt gamma.txt
git -C "$REPO" commit -q -m "feat: add alpha beta and gamma"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 53: a wrapped partner id is not read as part of the field" || fail "case 53: rc=$RC out=$OUT"
case "$OUT" in
  *"gamma.txt is not declared"*) pass "case 53: the fold's file union is not widened to the wrapped id's task" ;;
  *) fail "case 53: expected gamma.txt to be undeclared collateral, out=$OUT" ;;
esac

# ===========================================================================
# Case 54 (fix round 6, F15): an INVALID Squash-with edge must not join two
# folds. Task 1 is red and names task 9, which is itself red — a plan defect
# reported against task 1. Task 9's own fold with green task 10 is separate,
# valid, and carried by its own commit. The fixed-point growth in
# `_fold_reds` admitted task 1 as a member of task 10's group before
# validating the edge that reached it, so querying task 10 — or task 9 —
# reported task 1's violation and FAILED a commit that is entirely correct.
# An edge that does not resolve is not an edge: it is reported from the red
# task carrying it, and joins nothing.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. A red task naming a partner that is itself red

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 9

- [ ] 9. A red task whose own partner is green

**Files:** `beta.txt`
**Commit:** test: add beta
**Build:** red

**Squash-with:** Task 10

- [ ] 10. The green partner of task 9 fold

**Files:** `gamma.txt`
**Commit:** feat: add beta and gamma
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'b\n' > "$REPO/beta.txt"
printf 'g\n' > "$REPO/gamma.txt"
git -C "$REPO" add beta.txt gamma.txt
git -C "$REPO" commit -q -m "feat: add beta and gamma"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 10 "$SHA"
[ "$RC" -eq 0 ] && pass "case 54: an unrelated valid fold is not contaminated by another red's invalid edge" || fail "case 54 (task 10): rc=$RC out=$OUT"
run_guard "$REPO" 9 "$SHA"
[ "$RC" -eq 0 ] && pass "case 54: the red half of that valid fold reaches the same verdict" || fail "case 54 (task 9): rc=$RC out=$OUT"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 54: the invalid edge is still a violation from the task carrying it" || fail "case 54 (task 1): rc=$RC out=$OUT"
case "$OUT" in
  *"Task 9, which is itself red"*) pass "case 54: the violation names the red partner it could not resolve" ;;
  *) fail "case 54: expected the itself-red message for task 1, out=$OUT" ;;
esac

# ===========================================================================
# Case 55 (fix round 6, F16): a red task whose `Squash-with:` value does not
# gate names no partner the guard may trust, so the guard cannot know which
# commit that task folded into — every fold verdict in the plan is therefore
# unvouched-for, and the malformed field is reported from EVERY id in the
# plan rather than only from the red task's own. Before, the red was simply
# excluded from fold resolution, so querying the green task its free text
# names reported the folded commit's other file as undeclared collateral —
# a true statement about the wrong defect, which sends a reader to the
# `Files:` field instead of the broken `Squash-with:` one.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red task whose Squash-with names its partner in free text

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2 (see step 3)

- [ ] 2. The green partner that free text names

**Files:** `beta.txt`
**Commit:** feat: add alpha and beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "feat: add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 1 ] && pass "case 55: a malformed Squash-with elsewhere in the plan still fails the partner it names" || fail "case 55: rc=$RC out=$OUT"
case "$OUT" in
  *"task 1: Squash-with:"*"is not \`Task"*) pass "case 55: the malformed field is discoverable from the partner id" ;;
  *) fail "case 55: expected task 1 malformed-field message from task 2, out=$OUT" ;;
esac
case "$OUT" in
  *"alpha.txt is not declared"*) fail "case 55: the misleading undeclared-file message is still reported, out=$OUT" ;;
  *) pass "case 55: the misleading undeclared-file message is not reported instead" ;;
esac

# ===========================================================================
# Case 56 (fix round 7, surviving mutant): the wrapper's OWN missing-grammar
# check, at runtime. Replacing `if [ ! -f "$GRAMMAR_MODULE" ]` with
# `if false` broke no assertion in either harness: the line's other purpose
# — being grep-visible to check-guard-symlinks.sh rule 2, which derives a
# guard's required siblings from `$SCRIPT_DIR/<name>` — is covered by that
# guard's own harness, but nothing exercised the exit the check performs.
#
# The fixture is a throwaway copy of the guard PAIR with no `lib/` sibling,
# invoked against an otherwise-clean repo that the shipped guard passes
# (case 1's shape). So both assertions discriminate: with the check removed
# the copy reaches python3, whose `from plan_grammar import ...` raises and
# exits 1 with a traceback rather than 2 with a sentence.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Clean task

**Files:** `alpha.txt`, `beta.txt`
**Tests:** `test_alpha`
**Commit:** add alpha and beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf '# test_alpha covers alpha\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
# Resolved through cd/pwd so the path matches the one the wrapper prints,
# which it derives the same way — on macOS $TMPDIR is itself a symlink.
STRIPPED="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/task-commit-fields-test.XXXXXX")" && pwd)"
cp "$SCRIPT_DIR/check-task-commit-fields.sh" "$SCRIPT_DIR/check-task-commit-fields.py" "$STRIPPED/"
set +e
STRIPPED_OUT="$("$STRIPPED/check-task-commit-fields.sh" "$REPO" 1 "$SHA" 2>&1)"
STRIPPED_RC=$?
set -e
[ "$STRIPPED_RC" -eq 2 ] && pass "case 56: a guard copy with no lib/ sibling exits 2" || fail "case 56: rc=$STRIPPED_RC out=$STRIPPED_OUT"
case "$STRIPPED_OUT" in
  *"shared grammar module not found: $STRIPPED/lib/plan_grammar.py"*) pass "case 56: the message names the missing module path" ;;
  *) fail "case 56: expected the missing-module path in the message, out=$STRIPPED_OUT" ;;
esac
rm -rf "$STRIPPED"

# ===========================================================================
# Case 57 (fix round 8, F19): WHICH line is the `Squash-with:` field. A body
# carrying a non-gating `Squash-with:` line ahead of a gating one has ONE
# field — the gating line — because field selection now lives in
# lib/plan_grammar.py's select_squash_with and both guards call it. Before,
# this guard took the first field-SHAPED line whatever its value, so it read
# the free-text line, resolved no partner, and (through case 55's plan-wide
# reporting) failed every task in the plan — while check-task-build-green.sh
# resolved partner 2 from the second line and exited 0. The same fixture is
# asserted against that guard as its own case 19, so the two verdicts are
# pinned from both sides.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red task whose first Squash-with line does not gate

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red

**Squash-with:** Task 2 (see note below)
**Squash-with:** Task 2

- [ ] 2. The green partner named by the gating line

**Files:** `beta.txt`
**Commit:** feat: add alpha and beta
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'b\n' > "$REPO/beta.txt"
git -C "$REPO" add alpha.txt beta.txt
git -C "$REPO" commit -q -m "feat: add alpha and beta"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 57: the gating line is the field, so the red half of the fold passes" || fail "case 57 (task 1): rc=$RC out=$OUT"
run_guard "$REPO" 2 "$SHA"
[ "$RC" -eq 0 ] && pass "case 57: the green half reaches the same verdict" || fail "case 57 (task 2): rc=$RC out=$OUT"
case "$OUT" in
  *"is not \`Task"*) fail "case 57: the non-gating line was still read as the field, out=$OUT" ;;
  *) pass "case 57: the non-gating line is not reported as the field's value" ;;
esac

# ===========================================================================
# Case 58 (fix round 9, F20): WHICH line is the `**Build:**` tag, when a body
# carries two. This guard used to read `Build` through FIELD_RE's
# alternation, where a later field of the same name overwrote an earlier
# one, so this task was GREEN here and red in check-task-build-green.sh:
# that guard resolved the fold while this one put the task on the ordinary
# single-commit path and checked it against a `Commit:` subject the fold
# would have deleted. Tag selection now lives in lib/plan_grammar.py's
# select_build_tag, which both guards call — the first line-gated tag wins,
# so the task is red and its missing partner is reported. The same body is
# asserted against check-task-build-green.sh as its own case 21.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red task carrying two Build lines

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red
**Build:** green

**Squash-with:** Task 9
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "test: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 58: the first Build line is the tag, so the task is red" || fail "case 58: rc=$RC out=$OUT"
case "$OUT" in
  *"task 1: Squash-with: names Task 9, which does not exist in this plan"*) pass "case 58: the red task's missing partner is reported" ;;
  *) fail "case 58: expected the missing-partner message, out=$OUT" ;;
esac

# ===========================================================================
# Case 59 (fix round 9, F20): the `**Build:**` tag is LINE-SCOPED. A prose
# line following the tag with no blank line between used to be JOINED onto
# its value here, leaving the task with no tag at all while
# check-task-build-green.sh read it as red — the same split verdict as case
# 58, reached the other way. The same body is asserted against that guard as
# its own case 22.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Red task whose tag line is followed by prose

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** red
this sentence explains the tag and is not part of it

**Squash-with:** Task 9
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "test: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 59: a prose line under the tag does not unset it" || fail "case 59: rc=$RC out=$OUT"
case "$OUT" in
  *"task 1: Squash-with: names Task 9, which does not exist in this plan"*) pass "case 59: the red task's missing partner is reported" ;;
  *) fail "case 59: expected the missing-partner message, out=$OUT" ;;
esac

# ===========================================================================
# Case 60 (fix round 9): a fenced example task line opens no task. This
# guard's task scan used to ignore fences entirely, so the worked example
# below became a real task 9 whose ungated `Squash-with:` was reported —
# through case 55's plan-wide rule — against every task in the plan, while
# check-task-build-green.sh saw one clean green task. Task splitting is now
# lib/plan_grammar.py's iter_tasks, which both guards use. The same body is
# asserted against that guard as its own case 23.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Real task with a worked example in its body

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** green

Example of a fold, shown but never declared:

```
- [ ] 9. Example red task
**Build:** red
**Squash-with:** Task 8 (see the note)
```
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "test: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 60: a fenced example task line opens no task" || fail "case 60: rc=$RC out=$OUT"
case "$OUT" in
  *"is not \`Task"*) fail "case 60: the fenced example was read as a real task, out=$OUT" ;;
  *) pass "case 60: the fenced example's Squash-with is not reported" ;;
esac

# ===========================================================================
# Case 61 (fix round 9): WHICH task a duplicated id names. This guard used
# to keep scanning past the first matching task line, so id 1 resolved to the
# LAST task line carrying it, while check-task-build-green.sh resolved every
# lookup to the first. Here that made the two guards read a different
# `Build:` tag, a different `Files:` set and a different `Commit:` subject
# out of one id. `select_task` now answers it for both: the first task line
# wins, so this commit is checked against the green task 1 that declared it.
# The same body is asserted against that guard as its own case 24, where the
# duplicate itself is still reported.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. First task line for this id

**Files:** `alpha.txt`
**Commit:** test: add alpha
**Build:** green

- [ ] 1. Second task line reusing the id

**Files:** `zulu.txt`
**Commit:** test: add zulu
**Build:** red

- [ ] 3. Red task folding into Task 1

**Files:** `gamma.txt`
**Commit:** test: add gamma
**Build:** red

**Squash-with:** Task 1
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
printf 'g\n' > "$REPO/gamma.txt"
git -C "$REPO" add alpha.txt gamma.txt
git -C "$REPO" commit -q -m "test: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 61: a duplicated id resolves to its first task line" || fail "case 61: rc=$RC out=$OUT"
case "$OUT" in
  *"zulu.txt"*) fail "case 61: the id resolved to the last task line's fields, out=$OUT" ;;
  *) pass "case 61: the later task line's fields are not what the commit is checked against" ;;
esac

# ===========================================================================
# Case 62 (fix round 10, F21): a TILDE-fenced example field is not a field.
# This guard's own field loop gated a fence on "^```" while every selector it
# shares with check-task-build-green.sh gates on lib/plan_grammar.py's
# FENCE_RE, which is CommonMark's rule: backticks or tildes, up to three
# columns of indent. A `~~~` example block was therefore invisible to the
# field loop alone, and the `**Files:**` inside it OVERWROTE the real
# declaration above it — so a commit touching the declared file was reported
# undeclared, and a commit touching the example's file passed. Both halves
# are asserted below. The field loop now uses FENCE_RE too.
# ===========================================================================
FENCED_EXAMPLE_TILDE='- [ ] 1. Real task with a tilde-fenced example field

**Files:** `real.txt`
**Commit:** test: add real
**Build:** green

Example of how the field is written:

~~~markdown
**Files:** `example-only.txt`
~~~
'
new_repo
write_tasks_md "$REPO" "$FENCED_EXAMPLE_TILDE"
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'r\n' > "$REPO/real.txt"
git -C "$REPO" add real.txt
git -C "$REPO" commit -q -m "test: add real"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 62: the real Files: declaration survives a tilde-fenced example" || fail "case 62: rc=$RC out=$OUT"

# The discriminating half: the example's file is NOT declared, so a commit
# touching it is undeclared collateral.
new_repo
write_tasks_md "$REPO" "$FENCED_EXAMPLE_TILDE"
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'e\n' > "$REPO/example-only.txt"
git -C "$REPO" add example-only.txt
git -C "$REPO" commit -q -m "test: add real"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 62: a tilde-fenced example file is not declared" || fail "case 62: rc=$RC out=$OUT"
case "$OUT" in
  *example-only.txt*) pass "case 62: the violation names the fenced example's file" ;;
  *) fail "case 62: the undeclared file is not named, out=$OUT" ;;
esac

# ===========================================================================
# Case 63 (fix round 10, F21): the same rule for an INDENTED fence. CommonMark
# allows up to three columns of leading whitespace before a fence delimiter,
# which "^```" also missed. Asserted with backticks so this case discriminates
# on the indent alone, not on the delimiter character case 62 covers. The
# example field inside the block sits at column 0 because FIELD_RE is anchored
# at "^\*\*" — an indented field line is no field to begin with, so an indented
# example body would assert nothing.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Real task with an indented fenced example field

**Files:** `real.txt`
**Commit:** test: add real
**Build:** green

Example of how the field is written:

   ```markdown
**Files:** `example-only.txt`
   ```
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'r\n' > "$REPO/real.txt"
git -C "$REPO" add real.txt
git -C "$REPO" commit -q -m "test: add real"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 63: the real Files: declaration survives an indented fenced example" || fail "case 63: rc=$RC out=$OUT"

# ===========================================================================
# Case 64 (fix round 11, F22): a fence a task body OPENS and never CLOSES is
# reported as the plan defect it is, and REPLACES the downstream noise. In
# CommonMark an unclosed fence runs to the end of the block, so the
# `**Files:**` and `**Commit:**` lines below it genuinely are code — which is
# what a renderer shows and what every selector this guard shares with
# check-task-build-green.sh reads. Swallowing them is therefore correct; what
# was wrong was reporting only the CONSEQUENCE, an empty `Files:` set, which
# made the guard tell the operator that the very file the task declared was
# undeclared collateral. The same body is asserted against
# check-task-build-green.sh as its own case 25, where the swallowed
# `**Build:**` tag used to be reported as an untagged task.
# ===========================================================================
UNCLOSED_FENCE_BODY='- [ ] 1. Task whose body opens a fence it never closes

~~~

**Build:** green
**Files:** `alpha.txt`
**Commit:** test: add alpha
'
new_repo
write_tasks_md "$REPO" "$UNCLOSED_FENCE_BODY"
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "test: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 1 ] && pass "case 64: an unclosed fence in a task body fails" || fail "case 64: rc=$RC out=$OUT"
case "$OUT" in
  *"code fence opened at tasks.md line 3 is never closed"*) pass "case 64: names the task and the line the fence opened on" ;;
  *) fail "case 64: expected the unclosed-fence violation naming line 3, out=$OUT" ;;
esac
# The replacement half: the plan defect is reported INSTEAD of the commit
# being blamed for touching a file the swallowed `**Files:**` declared.
case "$OUT" in
  *"is not declared in Files:"*) fail "case 64: the commit is still blamed for the plan's defect, out=$OUT" ;;
  *) pass "case 64: the undeclared-file noise is replaced, not accompanied" ;;
esac

# The no-false-positive half: a fence the body CLOSES is not an unclosed one,
# and the fields above it are read exactly as before.
new_repo
write_tasks_md "$REPO" '- [ ] 1. Task whose body closes the fence it opens

**Build:** green
**Files:** `alpha.txt`
**Commit:** test: add alpha

~~~
example
~~~
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "test: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 64: a closed fence is not reported as unclosed" || fail "case 64: rc=$RC out=$OUT"

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
