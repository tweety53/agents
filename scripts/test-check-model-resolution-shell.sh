#!/usr/bin/env bash
# Assertion harness for check-model-resolution-shell.sh.
#
# The guard extracts and runs skills/flow/SKILL.md's own "## Model
# resolution" bash block, with no override for which file it reads — unlike
# most guards in this repository, so this harness cannot point it at a
# sandboxed fixture. It instead mutates the REAL SKILL.md in place, in a
# scratch copy captured up front, runs the guard, asserts the outcome, then
# restores the original content — via a trap, so the file is restored even
# if an assertion fails partway through. Never leaves the repository tree
# mutated on exit.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GUARD="$SCRIPT_DIR/check-model-resolution-shell.sh"
SKILL_MD="$REPO_ROOT/skills/flow/SKILL.md"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

[ -r "$SKILL_MD" ] || { echo "cannot read $SKILL_MD" >&2; exit 2; }

BACKUP="$(mktemp "${TMPDIR:-/tmp}/skill-md-backup.XXXXXX")"
cp "$SKILL_MD" "$BACKUP"
cleanup() {
  cp "$BACKUP" "$SKILL_MD"
  rm -f "$BACKUP"
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

# 2. Flipping SELF_REVIEW_MODEL's `-z` to `-n` breaks the fallback logic:
# a non-empty resolved value gets forcibly overwritten instead of preserved.
# The guard must catch this and fail.
sed -i.bak 's/\[ -z "\$SELF_REVIEW_MODEL" \] && SELF_REVIEW_MODEL=fable/[ -n "$SELF_REVIEW_MODEL" ] \&\& SELF_REVIEW_MODEL=fable/' "$SKILL_MD"
rm -f "$SKILL_MD.bak"
run_guard
[ "$RC" -eq 1 ] && pass "flipped -z/-n mutation is caught" \
  || fail "flipped -z/-n: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"SELF_REVIEW_MODEL"*) pass "the failure names SELF_REVIEW_MODEL" ;;
  *) fail "failure does not name SELF_REVIEW_MODEL: out=$OUT" ;;
esac

# 3. Restored, the guard passes clean again.
cp "$BACKUP" "$SKILL_MD"
run_guard
[ "$RC" -eq 0 ] && pass "restored block passes again" \
  || fail "restored: rc=$RC out=$OUT"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-model-resolution-shell: all cases pass\n'
