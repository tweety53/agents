#!/usr/bin/env bash
# Assertion harness for check-dispatch-paragraphs.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the guard at each
# with CHECK_DISPATCH_PARAGRAPHS_ROOT, invoking the REAL
# scripts/check-dispatch-paragraphs.sh as a subprocess against REAL fixture
# files on disk — never a copy of its logic, and never a hand-written
# expected-output string asserted without having run the guard to produce
# it. Never edits this repository's own skills/flow/*.md.
#
# Cases 1-7 cover REPRODUCE, DON'T READ, matching the plan's original
# minimum: both sites correct; the label absent from review-panel.md; the
# label present but missing a shared phrase; implement.md carrying only one
# block; implement.md carrying two blocks of the same variant (the guard
# must match variants, not just count blocks); a scoped file missing; a
# scoped file that is a symlink. Case 1's review-panel.md fixture also
# carries a correct VERBATIM REPORT — THE FACT block, since that site now
# requires both paragraphs to exit 0.
#
# Cases 8-11 cover KAN-217's VERBATIM REPORT — THE FACT paragraph, added
# when the guard was generalized from one hard-coded paragraph to a table
# of them: present and clean (case 1, above); the block missing entirely;
# and one case per required phrase, each dropped in turn.
#
# Case 12 covers CHECK_DISPATCH_PARAGRAPHS_ROOT set but empty (distinct
# from unset). Case 13 covers a required site path that is a directory,
# not a regular file. Case 14 covers a required site file that exists but
# is not readable. Case 15 is the block-text-bleed regression: two
# required blocks glued with no blank line between them, where the first
# is deficient and must not pass by absorbing the second's phrases.
#
# Per KAN-197, every check this file targets was mutation-tested by hand
# during authoring: the check was disabled or removed from a throwaway copy
# of the guard, the same fixture re-run, and the case's failure signal
# confirmed absent — proving the fixture's failure depended on that check,
# not on some other check incidentally catching the same fixture. That is
# an authoring-time review practice, not a mechanism this file runs itself
# (contrast test-check-plan-shape.sh, which automates the same idea
# in-suite via a throwaway mutated copy per case). Case 6's inline comment
# below records one such mutation catching a real surviving-mutant bug, and
# cases 2 and 4 assert on the min-blocks violation's own message — pinned
# after hand-mutating SITE_MIN_BLOCKS showed the exit-code-only assertion
# passing regardless of that check.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-dispatch-paragraphs.sh"
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
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-dispatch-paragraphs-test.XXXXXX")"
  DIRS+=("$ROOT")
  mkdir -p "$ROOT/skills/flow"
}

# run_guard -> sets RC and OUT, running the real guard against $ROOT.
run_guard() {
  set +e
  OUT="$(CHECK_DISPATCH_PARAGRAPHS_ROOT="$ROOT" "$GUARD" 2>&1)"
  RC=$?
  set -e
}

# The two REPRODUCE, DON'T READ variants, reproduced verbatim from
# design.md's Part 1 blockquotes.
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

# The VERBATIM REPORT — THE FACT paragraph, reproduced verbatim from
# skills/flow/review-panel.md.
VERBATIM_BLOCK='> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot'"'"'s report was written
> to. That file is the reviewer'"'"'s own report, unedited — read it before you act on the finding.
> The structured block is the dispatcher'"'"'s summary of it: direction on what to work on, and
> never a source of fact. Where the two disagree the report wins. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.'

# Variants of VERBATIM_BLOCK, each with exactly one required phrase dropped
# while staying a plausible paragraph — cases 10-12.
VERBATIM_BLOCK_NO_REVIEWERS_OWN_REPORT='> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot'"'"'s report was written
> to. That file is the report, unedited — read it before you act on the finding.
> The structured block is the dispatcher'"'"'s summary of it: direction on what to work on, and
> never a source of fact. Where the two disagree the report wins. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.'

VERBATIM_BLOCK_NO_NEVER_A_SOURCE_OF_FACT='> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot'"'"'s report was written
> to. That file is the reviewer'"'"'s own report, unedited — read it before you act on the finding.
> The structured block is the dispatcher'"'"'s summary of it: direction on what to work on, and
> is not itself a fact. Where the two disagree the report wins. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.'

VERBATIM_BLOCK_NO_REPORT_WINS='> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot'"'"'s report was written
> to. That file is the reviewer'"'"'s own report, unedited — read it before you act on the finding.
> The structured block is the dispatcher'"'"'s summary of it: direction on what to work on, and
> never a source of fact. Where the two disagree the reviewer'"'"'s report is followed. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.'

write_site() {
  local relpath="$1" content="$2"
  printf '%s\n' "$content" > "$ROOT/$relpath"
}

# ===========================================================================
# Case 1: both required sites correct — exit 0. review-panel.md now carries
# both the REPRODUCE reviewer block and the VERBATIM REPORT block.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 0 ] && pass "case 1: both sites correct exits 0" \
  || fail "case 1: expected exit 0, got rc=$RC out=$OUT"

# ===========================================================================
# Case 2: the REPRODUCE label is absent from review-panel.md — exit 1,
# names the file.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "No REPRODUCE, DON'T READ paragraph here at all.

$VERBATIM_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 2: exits 1" || fail "case 2: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*) pass "case 2: names review-panel.md" ;;
  *) fail "case 2: expected review-panel.md named in output, got: $OUT" ;;
esac
# Pin the min-blocks violation's OWN message, not merely the overall exit
# code: with the label wholly absent, block_count is 0 for BOTH the
# min-blocks check and the reviewer-variant check, so an assertion on exit
# code alone cannot tell them apart. Hand-mutating SITE_MIN_BLOCKS's
# review-panel.md entry from 1 to 0 confirmed this exact message
# disappears while the case's RC==1 assertion above still passes (the
# variant violation alone keeps it red) — this line is what actually pins
# the min-blocks check for that site.
case "$OUT" in
  *"requires at least 1 block(s)"*) pass "case 2: names the min-blocks violation" ;;
  *) fail "case 2: expected the min-blocks violation message, got: $OUT" ;;
esac

# ===========================================================================
# Case 3: the REPRODUCE label is present but a shared phrase is missing —
# exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK_NO_SHARED_PHRASE

$VERBATIM_BLOCK"
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
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 4: exits 1" || fail "case 4: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"implement.md"*) pass "case 4: names implement.md" ;;
  *) fail "case 4: expected implement.md named in output, got: $OUT" ;;
esac
# Pin implement.md's own min-blocks violation message (its site requires 2
# blocks). Hand-mutating SITE_MIN_BLOCKS's implement.md entry from 2 to 1
# confirmed this exact message disappears while the case's RC==1 assertion
# above still passes (the missing-implementer-variant violation alone
# keeps it red) — this line is what actually pins the min-blocks check for
# that site.
case "$OUT" in
  *"requires at least 2 block(s)"*) pass "case 4: names the min-blocks violation" ;;
  *) fail "case 4: expected the min-blocks violation message, got: $OUT" ;;
esac

# ===========================================================================
# Case 5: implement.md carries two blocks of the SAME variant — exit 1. The
# reviewer/implementer pair is the requirement, not the count alone.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK"
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
OUTSIDE="$(mktemp "${TMPDIR:-/tmp}/check-dispatch-paragraphs-test-outside.XXXXXX")"
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

# Case 1 above already demonstrates the VERBATIM REPORT — THE FACT
# paragraph present and clean, alongside a correct REPRODUCE block, exiting
# 0 — no separate case is needed for that.

# ===========================================================================
# Case 8: the VERBATIM REPORT — THE FACT block is missing entirely from
# review-panel.md — exit 1, names the file and the required block count.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 8: exits 1" || fail "case 8: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"review-panel.md"*"VERBATIM REPORT"*) pass "case 8: names review-panel.md and the missing VERBATIM REPORT block" ;;
  *) fail "case 8: expected review-panel.md and VERBATIM REPORT named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 9: the VERBATIM REPORT block is present but missing
# "the reviewer's own report" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK_NO_REVIEWERS_OWN_REPORT"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 9: exits 1" || fail "case 9: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"the reviewer's own report"*) pass "case 9: names the missing phrase" ;;
  *) fail "case 9: expected \"the reviewer's own report\" named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 10: the VERBATIM REPORT block is present but missing
# "never a source of fact" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK_NO_NEVER_A_SOURCE_OF_FACT"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 10: exits 1" || fail "case 10: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"never a source of fact"*) pass "case 10: names the missing phrase" ;;
  *) fail "case 10: expected \"never a source of fact\" named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 11: the VERBATIM REPORT block is present but missing
# "the report wins" — exit 1, names the phrase.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK_NO_REPORT_WINS"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 11: exits 1" || fail "case 11: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"the report wins"*) pass "case 11: names the missing phrase" ;;
  *) fail "case 11: expected \"the report wins\" named in output, got: $OUT" ;;
esac

# ===========================================================================
# Case 12: CHECK_DISPATCH_PARAGRAPHS_ROOT is set but empty — exit 2, names
# the refusal (distinct from leaving the variable unset, which defaults to
# scanning this repository itself).
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
set +e
OUT="$(CHECK_DISPATCH_PARAGRAPHS_ROOT="" "$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 12: a set-but-empty root exits 2" \
  || fail "case 12: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"set but empty"*) pass "case 12: names the set-but-empty refusal" ;;
  *) fail "case 12: expected the set-but-empty refusal named, got: $OUT" ;;
esac

# ===========================================================================
# Case 13: a required site path exists but is a directory, not a regular
# file — exit 2, names the refusal.
# ===========================================================================
new_root
mkdir -p "$ROOT/skills/flow/review-panel.md"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 2 ] && pass "case 13: a directory at a site path exits 2" \
  || fail "case 13: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"not a regular file"*) pass "case 13: names the not-a-regular-file refusal" ;;
  *) fail "case 13: expected the not-a-regular-file refusal named, got: $OUT" ;;
esac

# ===========================================================================
# Case 14: a required site path exists as a regular file but is not
# readable — exit 2, names the refusal. Skipped when running as root,
# since root ignores a file's own permission bits.
# ===========================================================================
if [ "$(id -u)" -ne 0 ]; then
  new_root
  write_site "skills/flow/review-panel.md" "$REVIEWER_BLOCK

$VERBATIM_BLOCK"
  write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
  chmod 000 "$ROOT/skills/flow/implement.md"
  run_guard
  chmod 644 "$ROOT/skills/flow/implement.md"
  [ "$RC" -eq 2 ] && pass "case 14: an unreadable site file exits 2" \
    || fail "case 14: expected exit 2, got rc=$RC out=$OUT"
  case "$OUT" in
    *"is not readable"*) pass "case 14: names the not-readable refusal" ;;
    *) fail "case 14: expected the not-readable refusal named, got: $OUT" ;;
  esac
else
  pass "case 14: skipped (running as root, permission bits are not enforced)"
fi

# ===========================================================================
# Case 15: two required blocks glued together with no blank line between
# them — the first (deficient) block must NOT absorb the second block's
# text and pass by borrowing its phrases. Regression for the block-text
# bleed across adjacent blockquote paragraphs found in KAN-217's review:
# extract_block_text used to keep walking through any '>' continuation
# line regardless of whether it started a second required block, so a
# missing-phrase reviewer block glued directly to a correct VERBATIM block
# passed clean by absorbing the VERBATIM block's own phrases.
# ===========================================================================
new_root
write_site "skills/flow/review-panel.md" "> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — please just read the code and trust it, worth less than one you did.
> **VERBATIM REPORT — THE FACT:** the reviewer's own report, never a source of fact, the report wins, exercise the real thing"
write_site "skills/flow/implement.md" "$REVIEWER_BLOCK

$IMPLEMENTER_BLOCK"
run_guard
[ "$RC" -eq 1 ] && pass "case 15: a glued deficient block still exits 1" \
  || fail "case 15: expected exit 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"exercise the real thing"*) pass "case 15: names the phrase the glued block did not actually carry" ;;
  *) fail "case 15: expected 'exercise the real thing' named in output, got: $OUT" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
