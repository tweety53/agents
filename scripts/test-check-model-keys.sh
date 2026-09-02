#!/usr/bin/env bash
# Assertion harness for check-model-keys.sh. Builds fixture project roots
# under a sandboxed TMPDIR and asserts the guard's exit status and output.
# Never touches the real repository tree.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-model-keys.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

ROOTS=()
cleanup() {
  [ "${#ROOTS[@]}" -eq 0 ] && return 0
  for r in "${ROOTS[@]}"; do
    rm -rf "$r"
  done
}
trap cleanup EXIT

# new_root -> sets ROOT to a fresh project root directory (no .flow yet).
new_root() {
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-model-keys-test.XXXXXX")"
  ROOTS+=("$ROOT")
}

run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

# 1. A valid `## self review model` key passes.
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`fable`\n' > "$ROOT/.flow/project.md"
run_guard "$ROOT"
[ "$RC" -eq 0 ] && pass "valid self review model key passes" \
  || fail "valid key: rc=$RC out=$OUT"

# 2. An invalid model value fails with a named violation.
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`bogus-model`\n' > "$ROOT/.flow/project.md"
run_guard "$ROOT"
[ "$RC" -eq 1 ] && pass "invalid model value fails" \
  || fail "invalid value: rc=$RC out=$OUT"
case "$OUT" in
  *"bogus-model"*"is not a ValidModels member"*) pass "invalid value is named" ;;
  *) fail "invalid value not named: out=$OUT" ;;
esac

# 3. A project root with no .flow/project.md at all is a pass, not a
# violation — absence is a supported, valid state.
new_root
run_guard "$ROOT"
[ "$RC" -eq 0 ] && pass "missing .flow/project.md passes" \
  || fail "missing project.md: rc=$RC out=$OUT"

# 4. F5 regression: every project root examined counts toward the summary,
# including one with no .flow/project.md. Mixing one root that has a valid
# project.md with one that has none must report "2 project(s) checked", not
# undercount the one with no file.
new_root
NO_FILE_ROOT="$ROOT"
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`fable`\n' > "$ROOT/.flow/project.md"
WITH_FILE_ROOT="$ROOT"
run_guard "$NO_FILE_ROOT" "$WITH_FILE_ROOT"
[ "$RC" -eq 0 ] && pass "mixed roots exit 0" || fail "mixed roots: rc=$RC out=$OUT"
case "$OUT" in
  *"2 project(s) checked"*) pass "a root with no .flow/project.md still counts toward the summary" ;;
  *) fail "summary undercounts: out=$OUT" ;;
esac

# 5. A non-directory project root exits 2.
run_guard "${TMPDIR:-/tmp}/check-model-keys-does-not-exist-$$"
[ "$RC" -eq 2 ] && pass "a nonexistent project root exits 2" \
  || fail "nonexistent root: rc=$RC out=$OUT"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-model-keys: all cases pass\n'
