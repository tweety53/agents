#!/usr/bin/env bash
# Assertion harness for check-python-suppressions.sh. Builds fixture files
# under a sandboxed mktemp directory and asserts the guard's exit status and,
# where the case names one, the presence of the expected hit text. Never
# touches the real repository tree.
#
# Modeled on test-check-vocabulary.sh's fixture-driven pattern: fixtures live
# under mktemp -d, the guard is invoked via a thin run_guard helper that
# captures RC/OUT, and every case ends with an explicit pass/fail assertion —
# never a bare "it didn't crash".
#
# check-python-suppressions.sh takes explicit path arguments (that is the seam
# the harness uses): passing it a fixture directory scans exactly that
# directory, never the real repo.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-python-suppressions.sh"
FAILURES=0

# Every fixture directory is registered here and removed by a single EXIT
# trap, matching test-check-vocabulary.sh's own hygiene: a harness that
# leaks its fixtures litters the machine it proves things on.
FIXTURES=()
cleanup_fixtures() { [ "${#FIXTURES[@]}" -eq 0 ] || rm -rf "${FIXTURES[@]}"; }
trap cleanup_fixtures EXIT

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# run_guard <path> -> sets RC and OUT
run_guard() {
  set +e
  OUT="$("$GUARD" "$1" 2>&1)"
  RC=$?
  set -e
}

# run_guard_from <cwd> [guard args...] -> sets RC and OUT, with the guard's
# working directory set to <cwd> instead of this harness's own
run_guard_from() {
  local from="$1"; shift
  set +e
  OUT="$(cd "$from" && "$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/check-python-suppressions-test.XXXXXX")"
  FIXTURES+=("$FIXTURE")
}

# ===========================================================================
# Case 1: a clean .py file passes.
# ===========================================================================
new_fixture
printf 'import sys\n\nsys.exit(0)\n' > "$FIXTURE/clean.py"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 1: clean file passes" || fail "case 1: rc=$RC out=$OUT"

# ===========================================================================
# Case 2: a bare `# noqa` is caught and the hit names the file.
# ===========================================================================
new_fixture
printf 'try:\n    risky()\nexcept Exception:  # noqa\n    pass\n' > "$FIXTURE/bare_noqa.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 2: bare noqa caught" || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"bare_noqa.py"*) pass "case 2: hit names the file" ;;
  *) fail "case 2: expected the hit to name bare_noqa.py, out=$OUT" ;;
esac

# ===========================================================================
# Case 3: a coded `# noqa: E402` is caught too.
# ===========================================================================
new_fixture
printf 'import os  # noqa: E402\n' > "$FIXTURE/coded_noqa.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 3: coded noqa caught" || fail "case 3: rc=$RC out=$OUT"

# ===========================================================================
# Case 4: `# type: ignore` (mypy/pyright) is caught.
# ===========================================================================
new_fixture
printf 'x: int = something()  # type: ignore\n' > "$FIXTURE/type_ignore.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 4: type: ignore caught" || fail "case 4: rc=$RC out=$OUT"

# ===========================================================================
# Case 5: `# pylint: disable` is caught.
# ===========================================================================
new_fixture
printf 'def f():  # pylint: disable=missing-docstring\n    return 1\n' > "$FIXTURE/pylint_off.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 5: pylint disable caught" || fail "case 5: rc=$RC out=$OUT"

# ===========================================================================
# Case 6: `# mypy: ignore-errors` is caught.
# ===========================================================================
new_fixture
printf '# mypy: ignore-errors\n' > "$FIXTURE/mypy_off.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 6: mypy ignore-errors caught" || fail "case 6: rc=$RC out=$OUT"

# ===========================================================================
# Case 7: `# nolint` (golangci/eslint-style) is caught.
# ===========================================================================
new_fixture
printf 'value = load()  # nolint\n' > "$FIXTURE/nolint.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 7: nolint caught" || fail "case 7: rc=$RC out=$OUT"

# ===========================================================================
# Case 8: the match is case-insensitive.
# ===========================================================================
new_fixture
printf 'import json  # NOQA\n' > "$FIXTURE/upper_noqa.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 8: NOQA uppercase caught" || fail "case 8: rc=$RC out=$OUT"

# ===========================================================================
# Case 9: a non-.py file carrying the same markers is not scanned.
# ===========================================================================
new_fixture
printf '# noqa\n# type: ignore\n' > "$FIXTURE/notes.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 9: non-python file skipped" || fail "case 9: rc=$RC out=$OUT"

# ===========================================================================
# Case 10: a marker in a nested subdirectory is still caught.
# ===========================================================================
new_fixture
mkdir -p "$FIXTURE/pkg/deep"
printf 'raise ValueError()  # noqa\n' > "$FIXTURE/pkg/deep/nested.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 10: nested directory scanned" || fail "case 10: rc=$RC out=$OUT"

# ===========================================================================
# Case 11: a path argument that is neither file nor directory is a hard
# error (exit 2), never a silent skip.
# ===========================================================================
new_fixture
run_guard "$FIXTURE/does-not-exist"
[ "$RC" -eq 2 ] && pass "case 11: missing path is exit 2" || fail "case 11: rc=$RC out=$OUT"

# ===========================================================================
# Case 12: a marker anywhere on the line — leading comment, not trailing —
# is caught (file-level `# flake8: noqa` shape).
# ===========================================================================
new_fixture
printf '# flake8: noqa\n' > "$FIXTURE/filelevel.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 12: file-level flake8 noqa caught" || fail "case 12: rc=$RC out=$OUT"

# ===========================================================================
# Case 13: a symlinked .py next to its real file is not scanned itself —
# grep follows it, so the same hit would be reported once per link and the
# real file once more; the guard skips the link and reports the target once.
# ===========================================================================
new_fixture
printf 'x = 1  # noqa\n' > "$FIXTURE/real.py"
ln -s real.py "$FIXTURE/link.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 13: symlinked copy still finds the hit" || fail "case 13: rc=$RC out=$OUT"
[ "$(printf '%s\n' "$OUT" | grep -c 'noqa')" -eq 1 ] &&
  pass "case 13: hit reported exactly once" ||
  fail "case 13: expected exactly one reported hit, out=$OUT"

# ===========================================================================
# Case 14: no-args mode is independent of the caller's cwd — the scan set is
# resolved absolutely from the guard's own repo root, so a run from any
# other directory answers for the repository (exit 0 clean, 1 markers), and
# never 2 "cannot answer" on a healthy repo.
# ===========================================================================
new_fixture
run_guard_from "$FIXTURE"
[ "$RC" -ne 2 ] && pass "case 14: no-args scan independent of caller cwd" || fail "case 14: rc=$RC out=$OUT"

# ===========================================================================
# Case 15: a git ls-files failure is exit 2, never a vacuous clean answer —
# a broken GIT_DIR makes the enumeration fail while the guard's own repo is
# healthy, exercising exactly the status check the loop cannot observe.
# ===========================================================================
new_fixture
set +e
OUT="$(GIT_DIR="$FIXTURE/not-a-repo" "$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 15: git enumeration failure is exit 2" || fail "case 15: rc=$RC out=$OUT"

# ===========================================================================
# Case 16: a find failure in the args-directory branch is exit 2, never a
# clean answer over a partial scan — an unreadable subtree makes find exit
# non-zero. Skipped under root, which no chmod can make unreadable.
# ===========================================================================
if [ "$(id -u)" -eq 0 ]; then
  pass "case 16: skipped under root (no chmod-unreadable directory exists)"
else
  new_fixture
  mkdir -p "$FIXTURE/hidden"
  printf 'y = 2  # noqa\n' > "$FIXTURE/hidden/secret.py"
  printf 'z = 3\n' > "$FIXTURE/visible.py"
  chmod 000 "$FIXTURE/hidden"
  run_guard "$FIXTURE"
  chmod 755 "$FIXTURE/hidden"
  [ "$RC" -eq 2 ] && pass "case 16: find enumeration failure is exit 2" || fail "case 16: rc=$RC out=$OUT"
fi

# ===========================================================================
# Case 17: pyright's own `# pyright: ignore` spelling is caught — the
# header names pyright, so the regex must cover its native marker too.
# ===========================================================================
new_fixture
printf 'x: int = something()  # pyright: ignore\n' > "$FIXTURE/pyright_ignore.py"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 17: pyright: ignore caught" || fail "case 17: rc=$RC out=$OUT"

# ===========================================================================
if [ "$FAILURES" -eq 0 ]; then
  printf 'all cases passed\n'
  exit 0
fi
printf '%d case(s) failed\n' "$FAILURES" >&2
exit 1
