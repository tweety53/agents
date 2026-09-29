#!/usr/bin/env bash
# Assertion harness for check-model-resolution-shell.sh.
#
# The guard extracts and runs skills/flow/SKILL.md's own "## Model
# resolution" bash block, and drift-checks skills/flow/archive.md for a
# SELF_REVIEW_MODEL resolution (KAN-854: no run resolves it). Its CHECK_MODEL_RESOLUTION_SKILL_MD and
# CHECK_MODEL_RESOLUTION_ARCHIVE_MD overrides (the
# RUN_GUARD_TESTS_ROOT idiom) let this harness point it at scratch copies
# of the real files, so the mutation cases below never write the real tree —
# KAN-376: this harness used to mutate the REAL SKILL.md in place and
# restore it, and a restore landing inside a concurrent test-setup.sh's
# fingerprint window (which hashes stat lines including mtime) failed the
# containment case "the repo's own skills, rules and commands are
# unchanged" under run-guard-tests.sh's concurrent runner. The copy is
# taken from the real file, run clean, mutated, run again, restored and run
# a third time — the same assertions as before, on a file the suite does
# not share.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GUARD="$SCRIPT_DIR/check-model-resolution-shell.sh"
SKILL_MD="$REPO_ROOT/skills/flow/SKILL.md"
ARCHIVE_MD="$REPO_ROOT/skills/flow/archive.md"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

[ -r "$SKILL_MD" ] || { echo "cannot read $SKILL_MD" >&2; exit 2; }
[ -r "$ARCHIVE_MD" ] || { echo "cannot read $ARCHIVE_MD" >&2; exit 2; }

WORK_SKILL_MD="$(mktemp "${TMPDIR:-/tmp}/skill-md-under-test.XXXXXX")"
WORK_ARCHIVE_MD="$(mktemp "${TMPDIR:-/tmp}/archive-md-under-test.XXXXXX")"
cp "$SKILL_MD" "$WORK_SKILL_MD"
cp "$ARCHIVE_MD" "$WORK_ARCHIVE_MD"
CHECK_MODEL_RESOLUTION_SKILL_MD="$WORK_SKILL_MD"
CHECK_MODEL_RESOLUTION_ARCHIVE_MD="$WORK_ARCHIVE_MD"
export CHECK_MODEL_RESOLUTION_SKILL_MD CHECK_MODEL_RESOLUTION_ARCHIVE_MD
cleanup() {
  rm -f "$WORK_SKILL_MD" "$WORK_SKILL_MD.bak" "$WORK_ARCHIVE_MD" "$WORK_ARCHIVE_MD.bak"
}
trap cleanup EXIT

run_guard() {
  set +e
  OUT="$("$GUARD" 2>&1)"
  RC=$?
  set -e
}

# 1. The unmutated block passes.
run_guard
[ "$RC" -eq 0 ] && pass "unmutated block passes" || fail "unmutated: rc=$RC out=$OUT"

# 2. Flipping DEFAULT_MODEL's `-n` to `-z` breaks the fallback logic: a
# non-empty resolved value gets forcibly overwritten with the `opus` literal
# instead of preserved. The guard must catch this and fail.
sed -i.bak 's/^\[ -n "\$DEFAULT_MODEL" \]/[ -z "$DEFAULT_MODEL" ]/' "$WORK_SKILL_MD"
rm -f "$WORK_SKILL_MD.bak"
cmp -s "$SKILL_MD" "$WORK_SKILL_MD" && fail "the -n/-z mutation did not land in the copy"
run_guard
[ "$RC" -eq 1 ] && pass "flipped -n/-z mutation is caught" \
  || fail "flipped -n/-z: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"DEFAULT_MODEL"*) pass "the failure names DEFAULT_MODEL" ;;
  *) fail "failure does not name DEFAULT_MODEL: out=$OUT" ;;
esac
cp "$SKILL_MD" "$WORK_SKILL_MD"

# 3. A SELF_REVIEW_MODEL resolution re-added to archive.md is refused: no
# run resolves it (KAN-854), and the drift check keeps it that way.
printf '   SELF_REVIEW_MODEL=fable\n' >>"$WORK_ARCHIVE_MD"
run_guard
[ "$RC" -eq 2 ] && pass "a re-added SELF_REVIEW_MODEL resolution is refused" \
  || fail "re-added SELF_REVIEW_MODEL: expected rc=2, got rc=$RC out=$OUT"
case "$OUT" in
  *"SELF_REVIEW_MODEL"*) pass "the refusal names SELF_REVIEW_MODEL" ;;
  *) fail "refusal does not name SELF_REVIEW_MODEL: out=$OUT" ;;
esac

# 4. Restored, the guard passes clean again. The pristine sources are the
# real files, read-only as far as this harness is concerned: they are
# copied FROM, never written.
cp "$SKILL_MD" "$WORK_SKILL_MD"
cp "$ARCHIVE_MD" "$WORK_ARCHIVE_MD"
run_guard
[ "$RC" -eq 0 ] && pass "restored files pass again" \
  || fail "restored: rc=$RC out=$OUT"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-model-resolution-shell: all cases pass\n'
