#!/usr/bin/env bash
# Assertion harness for check-uitest-overrides.sh.
#
# Builds throwaway Makefile fixtures under a sandboxed TMPDIR and points the
# guard at each with its `[<makefile>]` argument — never edits stats/Makefile
# itself, and never leaves a broken fixture behind.
#
# WHY THIS FILE EXISTS. check-uitest-overrides.sh was, for one review round,
# the only scripts/check-*.sh with no scripts/test-check-*.sh counterpart.
# Its regex matched only a plain `:=` anchored at column 0 and passed every
# one of: a leading space before the variable name, `?=`, `+=`, `=`, and
# `::=` — all of them real, command-line-overridable Make assignments. The
# cases below assert the guard rejects each of those forms; the first four
# were written and run against the OLD regex to confirm they failed there
# before the regex was fixed, which is what makes this suite evidence rather
# than a description of what the guard is hoped to do.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-uitest-overrides.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# Every case leaves one sandbox directory behind, removed on exit including
# on a failed assertion. An indexed array, not a space-separated string:
# mktemp paths under TMPDIR may contain spaces, and word-splitting a string
# would leak a sandbox whose path split and rm -rf the fragments.
DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do
    rm -rf "$d"
  done
}
trap cleanup EXIT

# new_dir -> sets DIR to a fresh sandbox directory.
new_dir() {
  DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-uitest-overrides-test.XXXXXX")"
  DIRS+=("$DIR")
}

# run_guard <makefile> -> sets RC and OUT.
run_guard() {
  set +e
  OUT="$("$GUARD" "$1" 2>&1)"
  RC=$?
  set -e
}

# assert_rejects <name> <makefile-body>
# Writes the body to a fresh Makefile, asserts exit 1 and the
# UITEST-OVERRIDES-FAIL verdict token.
assert_rejects() {
  local name="$1" body="$2"
  new_dir
  printf '%s\n' "$body" > "$DIR/Makefile"
  run_guard "$DIR/Makefile"
  if [ "$RC" -eq 1 ]; then
    pass "$name: exits 1"
  else
    fail "$name: expected exit 1, got rc=$RC out=$OUT"
  fi
  case "$OUT" in
    *"UITEST-OVERRIDES-FAIL:"*) pass "$name: prints UITEST-OVERRIDES-FAIL" ;;
    *) fail "$name: expected UITEST-OVERRIDES-FAIL in output, got: $OUT" ;;
  esac
}

# ===========================================================================
# 1-5: the five evasion forms F1 found — each a real, command-line
# overridable Make assignment that the OLD `^UITEST_[A-Z_]+[[:space:]]*:=`
# regex missed.
# ===========================================================================
assert_rejects "case 1 (leading space, plain :=)" '  UITEST_EVIL := /etc/passwd'
assert_rejects "case 2 (?=)" 'UITEST_EVIL ?= /etc/passwd'
assert_rejects "case 3 (+=)" 'UITEST_EVIL += /etc/passwd'
assert_rejects "case 4 (recursive =)" 'UITEST_EVIL = /etc/passwd'
assert_rejects "case 5 (::=)" 'UITEST_EVIL ::= /etc/passwd'

# ===========================================================================
# 6. A correctly `override`-protected assignment is accepted.
# ===========================================================================
new_dir
printf 'override UITEST_GOOD := ok\n' > "$DIR/Makefile"
run_guard "$DIR/Makefile"
[ "$RC" -eq 0 ] && pass "case 6: override-protected assignment exits 0" \
  || fail "case 6: rc=$RC out=$OUT"
case "$OUT" in
  *"UITEST-OVERRIDES-OK:"*) pass "case 6: prints UITEST-OVERRIDES-OK" ;;
  *) fail "case 6: expected UITEST-OVERRIDES-OK in output, got: $OUT" ;;
esac

# ===========================================================================
# 7. A non-UITEST_ variable, override or not, is never flagged.
# ===========================================================================
new_dir
printf 'OTHER_VAR := x\noverride OTHER_VAR2 := y\n' > "$DIR/Makefile"
run_guard "$DIR/Makefile"
[ "$RC" -eq 0 ] && pass "case 7: non-UITEST_ variables are ignored" \
  || fail "case 7: rc=$RC out=$OUT"

# ===========================================================================
# 8. A recipe line (leading TAB) that happens to start with UITEST_ text is
#    shell, not a Make assignment, and must not be flagged.
# ===========================================================================
new_dir
{
  printf 'dummy:\n'
  printf '\tUITEST_EVIL := not-a-make-assignment\n'
} > "$DIR/Makefile"
run_guard "$DIR/Makefile"
[ "$RC" -eq 0 ] && pass "case 8: a tab-led recipe line is not flagged" \
  || fail "case 8: rc=$RC out=$OUT"

# ===========================================================================
# 9. A UITEST_* variable moved into an `include`d file is still caught,
#    resolved relative to the including file's directory.
# ===========================================================================
new_dir
printf 'include extra.mk\noverride UITEST_GOOD := ok\n' > "$DIR/Makefile"
printf 'UITEST_HIDDEN := /etc/passwd\n' > "$DIR/extra.mk"
run_guard "$DIR/Makefile"
[ "$RC" -eq 1 ] && pass "case 9: a violation in an included file is caught" \
  || fail "case 9: rc=$RC out=$OUT"
case "$OUT" in
  *"extra.mk"*"UITEST_HIDDEN"*) pass "case 9: names the included file and variable" ;;
  *) fail "case 9: expected extra.mk/UITEST_HIDDEN in output, got: $OUT" ;;
esac

# ===========================================================================
# 10. An include target built from a Make variable ($(...)) cannot be
#     resolved without running Make — reported and skipped, not silently
#     ignored, and not a crash.
# ===========================================================================
new_dir
printf 'include $(EXTRA_FILES)\noverride UITEST_GOOD := ok\n' > "$DIR/Makefile"
run_guard "$DIR/Makefile"
[ "$RC" -eq 0 ] && pass "case 10: unresolved include does not crash the guard" \
  || fail "case 10: rc=$RC out=$OUT"
case "$OUT" in
  *"cannot resolve"*) pass "case 10: reports the unresolved include" ;;
  *) fail "case 10: expected a report of the unresolved include, got: $OUT" ;;
esac

# ===========================================================================
# 11-16: three more evasion forms, found in the round that added
# `override define` to stats/Makefile's UITEST_SAFE_RMRF macro — each beats
# the guard above (or beat it before this round) the same way cases 1-5
# did, and each has a correctly-protected counterpart that must still be
# accepted.
# ===========================================================================
assert_rejects "case 11 (define, no override)" 'define UITEST_EVIL
foo
endef'

new_dir
printf 'override define UITEST_GOOD\nfoo\nendef\n' > "$DIR/Makefile"
run_guard "$DIR/Makefile"
[ "$RC" -eq 0 ] && pass "case 12: override define is accepted" \
  || fail "case 12: rc=$RC out=$OUT"
case "$OUT" in
  *"UITEST-OVERRIDES-OK:"*) pass "case 12: prints UITEST-OVERRIDES-OK" ;;
  *) fail "case 12: expected UITEST-OVERRIDES-OK in output, got: $OUT" ;;
esac

assert_rejects "case 13 (export, no override)" 'export UITEST_EVIL := /etc/passwd'

new_dir
printf 'override export UITEST_GOOD := ok\n' > "$DIR/Makefile"
run_guard "$DIR/Makefile"
[ "$RC" -eq 0 ] && pass "case 14: override export is accepted" \
  || fail "case 14: rc=$RC out=$OUT"
case "$OUT" in
  *"UITEST-OVERRIDES-OK:"*) pass "case 14: prints UITEST-OVERRIDES-OK" ;;
  *) fail "case 14: expected UITEST-OVERRIDES-OK in output, got: $OUT" ;;
esac

assert_rejects "case 15 (\$(eval ...), no override)" '$(eval UITEST_EVIL := /etc/passwd)'

new_dir
printf '$(eval override UITEST_GOOD := ok)\n' > "$DIR/Makefile"
run_guard "$DIR/Makefile"
[ "$RC" -eq 0 ] && pass "case 16: \$(eval override ...) is accepted" \
  || fail "case 16: rc=$RC out=$OUT"
case "$OUT" in
  *"UITEST-OVERRIDES-OK:"*) pass "case 16: prints UITEST-OVERRIDES-OK" ;;
  *) fail "case 16: expected UITEST-OVERRIDES-OK in output, got: $OUT" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-uitest-overrides: all cases pass\n'
