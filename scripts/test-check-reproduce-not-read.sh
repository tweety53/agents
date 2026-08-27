#!/usr/bin/env bash
# Assertion harness for check-reproduce-not-read.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the guard at each
# with CHECK_REPRODUCE_NOT_READ_ROOT, invoking the REAL
# scripts/check-reproduce-not-read.sh as a subprocess against REAL fixture
# files on disk — never a copy of its logic, and never a hand-written
# expected-output string asserted without having run the guard to produce
# it. Never edits this repository's own skills/flow/*.md.
#
# Seven cases, matching the plan's minimum: both sites correct; the label
# absent from review-panel.md; the label present but missing a shared
# phrase; implement.md carrying only one block; implement.md carrying two
# blocks of the same variant (the guard must match variants, not just count
# blocks); a scoped file missing; a scoped file that is a symlink.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-reproduce-not-read.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# Every case leaves one sandbox directory behind, removed on exit including
# on a failed assertion. An indexed array, not a space-separated string:
# mktemp paths under TMPDIR may contain spaces.
DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do
    rm -rf "$d"
  done
}
trap cleanup EXIT

# new_root -> sets ROOT to a fresh sandbox directory carrying
# skills/flow/, matching the required-site table's scan-root-relative
# paths.
new_root() {
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-reproduce-not-read-test.XXXXXX")"
  DIRS+=("$ROOT")
  mkdir -p "$ROOT/skills/flow"
}

# run_guard -> sets RC and OUT, running the real guard against $ROOT.
run_guard() {
  set +e
  OUT="$(CHECK_REPRODUCE_NOT_READ_ROOT="$ROOT" "$GUARD" 2>&1)"
  RC=$?
  set -e
}

# The two variants, reproduced verbatim from design.md's Part 1 blockquotes.
REVIEWER_BLOCK='> **REPRODUCE, DON'"'"'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.'

IMPLEMENTER_BLOCK='> **REPRODUCE, DON'"'"'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one test you write MUST exercise the real
> thing. A test backed by a fake or a hand-built value passes while the real integration is broken:
> the shape you construct by hand is not the shape the real producer emits. Build the value the way
> production builds it, or assert against the real boundary.'

# A reviewer block with its "exercise the real thing" phrase gutted — case 3.
REVIEWER_BLOCK_NO_SHARED_PHRASE='> **REPRODUCE, DON'"'"'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST do the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.'

write_site() {
  local relpath="$1" content="$2"
  printf '%s\n' "$content" > "$ROOT/$relpath"
}

# ===========================================================================
# Case 1: both required sites correct — exit 0.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 0 ] && pass "case 1: both sites correct exits 0" \
  || fail "case 1: expected exit 0, got rc=$RC out=$OUT"

# ===========================================================================
# Case 2: the label is absent from review-panel.md — exit 1, names the file.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "No REPRODUCE, DON'T READ paragraph here at all."
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 2: exits 1" || fail "case 2: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*) pass "case 2: names review-panel.md" ;;
  *) fail "case 2: expected review-panel.md named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 3: the label is present but a shared phrase is missing — exit 1,
# names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK_NO_SHARED_PHRASE"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 3: exits 1" || fail "case 3: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"exercise the real thing"*) pass "case 3: names the missing phrase" ;;
  *) fail "case 3: expected 'exercise the real thing' named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 4: implement.md carries only one block — exit 1.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 4: exits 1" || fail "case 4: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*) pass "case 4: names implement.md" ;;
  *) fail "case 4: expected implement.md named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 5: implement.md carries two blocks of the SAME variant — exit 1. The
# reviewer/implementer pair is the requirement, not the count alone.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$REVIEWER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 5: two same-variant blocks still exits 1" \
  || fail "case 5: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*"implementer"*) pass "case 5: names implement.md's missing implementer variant" ;;
  *) fail "case 5: expected implement.md's missing implementer variant named, got: $OUT" ;;
esac

# ===========================================================================
# Case 6: a scoped file is missing entirely — exit 2.
# ===========================================================================
new_root
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
# review-panel.md is never written.
run_guard
[ "$RC" -eq 2 ] && pass "case 6: a missing scoped file exits 2" \
  || fail "case 6: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*) pass "case 6: names the missing file" ;;
  *) fail "case 6: expected review-panel.md named in output, got: $OUT" ;;
esac
# Pin the message to its OWN code path, not merely to the exit code. Exit 2 is
# reached here through four redundant layers — the existence check, the
# regular-file check, the readability check, and finally grep's own rc>=2 — so
# asserting rc==2 alone leaves the existence branch a SURVIVING MUTANT:
# deleting that branch outright still left this harness printing
# `all cases passed`, because a later layer caught the same fixture and
# reported a different reason. Found by mutation-testing during task 1's
# per-task review. Asserting the branch's own wording is what makes deleting
# it fail here.
case "$OUT" in
  *"does not exist"*) pass "case 6: reports it through the existence branch" ;;
  *) fail "case 6: expected the 'does not exist' message, got: $OUT" ;;
esac

# ===========================================================================
# Case 7: a scoped file is a symlink — exit 2.
# ===========================================================================
new_root
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
OUTSIDE="$(mktemp "${TMPDIR:-/tmp}/check-reproduce-not-read-test-outside.XXXXXX")"
DIRS+=("$OUTSIDE")
printf '%s\n' "$REVIEWER_BLOCK" > "$OUTSIDE"
ln -s "$OUTSIDE" "$ROOT/skills/flow/review-panel.md"
run_guard
[ "$RC" -eq 2 ] && pass "case 7: a symlinked scoped file exits 2" \
  || fail "case 7: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*) pass "case 7: names the symlinked file" ;;
  *) fail "case 7: expected review-panel.md named in output, got: $OUT" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
