#!/usr/bin/env bash
# Assertion harness for check-vocabulary.sh. Builds fixture files under a
# sandboxed mktemp directory and asserts the guard's exit status and, where
# the case names one, the presence of the expected hit text. Never touches
# the real repository tree.
#
# Modeled on test-check-task-build-green.sh's fixture-driven pattern: fixtures
# live under mktemp -d, the guard is invoked via a thin run_guard helper that
# captures RC/OUT, and every case ends with an explicit pass/fail assertion —
# never a bare "it didn't crash".
#
# check-vocabulary.sh takes explicit path arguments directly (unlike some
# other guards in this repo, it needs no env-var root override): passing it
# a fixture directory scans exactly that directory, never the real repo.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-vocabulary.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# run_guard <path> -> sets RC and OUT
run_guard() {
  set +e
  OUT="$("$GUARD" "$1" 2>&1)"
  RC=$?
  set -e
}

new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/check-vocabulary-test.XXXXXX")"
  FIXTURE_FILE="$FIXTURE/notes.md"
}

# ===========================================================================
# Case 1: the retired command's own name, written as its own token in
# backticks immediately followed by the word `skill` -- this is the harness
# skill, not the retired command, and must pass clean.
# ===========================================================================
new_fixture
printf 'the harness'"'"'s `code-review` skill handles this panel slot.\n' > "$FIXTURE_FILE"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 1: backtick-fenced + skill is exempt" || fail "case 1: rc=$RC out=$OUT"

# ===========================================================================
# Case 2: the same backtick-fenced token, but followed by `command` instead
# of `skill` -- this is exactly the retired command in its normal written
# form and must still be caught.
# ===========================================================================
new_fixture
printf 'Run the `code-review` command to advance the stage.\n' > "$FIXTURE_FILE"  # vocab-guard:allow
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 2: backtick-fenced + command is still caught" || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"code-review"*) pass "case 2: hit names the token" ;;  # vocab-guard:allow
  *) fail "case 2: expected the hit to name code-review, out=$OUT" ;;  # vocab-guard:allow
esac

# ===========================================================================
# Case 3: the bare token, no backticks at all -- must still be caught.
# ===========================================================================
new_fixture
printf 'The old code-review stage is gone.\n' > "$FIXTURE_FILE"  # vocab-guard:allow
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 3: bare token is still caught" || fail "case 3: rc=$RC out=$OUT"

# ===========================================================================
# Case 4: `requesting-code-review` -- the real external Superpowers skill,
# structurally excluded by the pattern's own prefix class exactly as before
# round one's fix, restored in this round.
# ===========================================================================
new_fixture
printf 'Use superpowers:requesting-code-review before merging.\n' > "$FIXTURE_FILE"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 4: requesting-code-review structurally excluded" || fail "case 4: rc=$RC out=$OUT"

# ===========================================================================
# Case 5: one line carrying BOTH an exempt occurrence and a bare retired
# occurrence -- the exempt one must not hide the genuine one. Exit 1, and
# the reported hit must still show the non-exempt occurrence.
# ===========================================================================
new_fixture
printf 'the harness'"'"'s `code-review` skill remains, but code-review itself was retired.\n' > "$FIXTURE_FILE"  # vocab-guard:allow
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 5: exempt occurrence does not hide a genuine one on the same line" \
  || fail "case 5: rc=$RC out=$OUT"
case "$OUT" in
  *"but code-review itself was retired"*) pass "case 5: reported line still carries the bare occurrence" ;;  # vocab-guard:allow
  *) fail "case 5: expected the reported line to show the bare occurrence, out=$OUT" ;;
esac

# ===========================================================================
# Case 6: a sampling of pre-existing retired tokens, proving the fix broke
# nothing else the guard already caught. Each in its own fixture, each
# expected to fail (RC=1).
# ===========================================================================
retired_tokens=(
  'awaiting-review'  # vocab-guard:allow
  'awaiting-test'  # vocab-guard:allow
  'myflow-full'  # vocab-guard:allow
  'skip-review'  # vocab-guard:allow
  'automerge'  # vocab-guard:allow
  'Gate B'  # vocab-guard:allow
)
i=0
for token in "${retired_tokens[@]}"; do
  i=$((i + 1))
  new_fixture
  printf 'Old text mentioning %s here.\n' "$token" > "$FIXTURE_FILE"
  run_guard "$FIXTURE"
  [ "$RC" -eq 1 ] && pass "case 6.$i: pre-existing token '$token' still caught" \
    || fail "case 6.$i: rc=$RC out=$OUT for token '$token'"
done

# ===========================================================================
# Case 7: a clean file with no retired vocabulary at all -> exit 0.
# ===========================================================================
new_fixture
printf 'Ordinary prose about the current three-state pipeline.\n' > "$FIXTURE_FILE"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 7: clean file passes" || fail "case 7: rc=$RC out=$OUT"

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
