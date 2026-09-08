#!/usr/bin/env bash
# Assertion harness for check-model-statements.sh.
#
# Builds a throwaway fixture tree under TMPDIR and points the guard at it
# with CHECK_MODEL_STATEMENTS_ROOT, invoking the REAL
# scripts/check-model-statements.sh as a subprocess against REAL fixture
# files on disk — never a copy of its logic. Never edits this repository's
# own skills/ trees.
#
# Case 1: every literal present — exit 0. Case 2: the canonical SKILL.md
# literal deleted — exit 1, naming the file. Case 3: a site sentence
# deleted (brainstorm.md) — exit 1. Case 4: a second site sentence deleted
# (review-panel.md) while others stand — exit 1. Case 5: a required file
# missing entirely — exit 2. Case 6: CHECK_MODEL_STATEMENTS_ROOT set but
# empty — exit 2. Case 7: the root a nonexistent directory — exit 2.
set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-model-statements.sh"

PASS=0
FAILURES=0
pass() { printf 'ok: %s\n' "$1"; PASS=$((PASS + 1)); }
fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }

SCRATCH="$(mktemp -d)"
trap 'rm -rf "$SCRATCH"' EXIT

mkfixture() {
  local root="$1"
  mkdir -p "$root/skills/flow" "$root/skills/flow-research"
  printf 'intro\nEvery model-bearing dispatch states its resolved value, immediately\nbefore the dispatch\n' > "$root/skills/flow/SKILL.md"
  printf 'x\nself-review model: <value> (<tier>)\n' > "$root/skills/flow/archive.md"
  printf 'planner model: <PLANNING_MODEL> (<tier>)\n' > "$root/skills/flow/brainstorm.md"
  printf 'conductor model: <value> (<tier>)\nplanner model: <PLANNING_MODEL> (<tier>)\n' > "$root/skills/flow/implement.md"
  printf '<slot> model: <value> (<tier>)\npanel-fix model: <DEFAULT_MODEL> (<tier>)\n' > "$root/skills/flow/review-panel.md"
  printf 'verify model: sonnet (fixed literal)\n' > "$root/skills/flow/verify-and-handoff.md"
  printf 'research model: <value> (<tier>)\n' > "$root/skills/flow-research/SKILL.md"
}

# Case 1: every literal present — exit 0.
FIX="$SCRATCH/case1"; mkfixture "$FIX"
OUT="$(CHECK_MODEL_STATEMENTS_ROOT="$FIX" bash "$GUARD" 2>&1)"; RC=$?
[ "$RC" -eq 0 ] && pass "case 1: all literals present, exits 0" || fail "case 1: expected exit 0, got rc=$RC out=$OUT"

# Case 2: canonical SKILL.md literal deleted — exit 1, names the file.
FIX="$SCRATCH/case2"; mkfixture "$FIX"
grep -v 'Every model-bearing dispatch states' "$FIX/skills/flow/SKILL.md" > "$FIX/skills/flow/SKILL.md.new" && mv "$FIX/skills/flow/SKILL.md.new" "$FIX/skills/flow/SKILL.md"
OUT="$(CHECK_MODEL_STATEMENTS_ROOT="$FIX" bash "$GUARD" 2>&1)"; RC=$?
if [ "$RC" -eq 1 ] && printf '%s' "$OUT" | grep -q 'skills/flow/SKILL.md'; then
  pass "case 2: canonical literal deleted, exits 1 naming SKILL.md"
else
  fail "case 2: expected exit 1 naming SKILL.md, got rc=$RC out=$OUT"
fi

# Case 3: a site sentence deleted (brainstorm.md) — exit 1.
FIX="$SCRATCH/case3"; mkfixture "$FIX"
printf 'other prose\n' > "$FIX/skills/flow/brainstorm.md"
OUT="$(CHECK_MODEL_STATEMENTS_ROOT="$FIX" bash "$GUARD" 2>&1)"; RC=$?
if [ "$RC" -eq 1 ] && printf '%s' "$OUT" | grep -q 'skills/flow/brainstorm.md'; then
  pass "case 3: site sentence deleted, exits 1 naming brainstorm.md"
else
  fail "case 3: expected exit 1 naming brainstorm.md, got rc=$RC out=$OUT"
fi

# Case 4: a second site deleted (review-panel.md) while others stand — exit 1.
FIX="$SCRATCH/case4"; mkfixture "$FIX"
printf 'other prose\n' > "$FIX/skills/flow/review-panel.md"
OUT="$(CHECK_MODEL_STATEMENTS_ROOT="$FIX" bash "$GUARD" 2>&1)"; RC=$?
if [ "$RC" -eq 1 ] && printf '%s' "$OUT" | grep -q 'skills/flow/review-panel.md'; then
  pass "case 4: site sentence deleted, exits 1 naming review-panel.md"
else
  fail "case 4: expected exit 1 naming review-panel.md, got rc=$RC out=$OUT"
fi

# Case 5: a required file missing entirely — exit 2.
FIX="$SCRATCH/case5"; mkfixture "$FIX"
rm "$FIX/skills/flow/verify-and-handoff.md"
OUT="$(CHECK_MODEL_STATEMENTS_ROOT="$FIX" bash "$GUARD" 2>&1)"; RC=$?
[ "$RC" -eq 2 ] && pass "case 5: required file missing, exits 2" || fail "case 5: expected exit 2, got rc=$RC out=$OUT"

# Case 6: CHECK_MODEL_STATEMENTS_ROOT set but empty — exit 2.
OUT="$(CHECK_MODEL_STATEMENTS_ROOT= bash "$GUARD" 2>&1)"; RC=$?
[ "$RC" -eq 2 ] && pass "case 6: root set but empty, exits 2" || fail "case 6: expected exit 2, got rc=$RC out=$OUT"

# Case 7: the root a nonexistent directory — exit 2.
OUT="$(CHECK_MODEL_STATEMENTS_ROOT="$SCRATCH/nope" bash "$GUARD" 2>&1)"; RC=$?
[ "$RC" -eq 2 ] && pass "case 7: nonexistent root, exits 2" || fail "case 7: expected exit 2, got rc=$RC out=$OUT"

# Case 8: every guard row's own literal, deleted one row at a time, must be
# caught. The rows are read from the guard's own `require` lines, so a row
# added to or removed from the guard keeps this case in sync automatically —
# and a row dropped from the guard shrinks this case's coverage with it,
# which is the mutation this case exists to make loud rather than silent.
CASE8_OK=1
CASE8_N=0
while IFS='|' read -r rel lit; do
  [ -n "$rel" ] || continue
  CASE8_N=$((CASE8_N + 1))
  FIX="$SCRATCH/case8-$CASE8_N"; mkfixture "$FIX"
  target="$FIX/$rel"
  # grep -v exits 1 when it selects nothing — never gate the mv on it, or
  # a single-line fixture keeps its literal and the mutation never applies.
  grep -vF -- "$lit" "$target" > "$target.new"
  mv "$target.new" "$target"
  CHECK_MODEL_STATEMENTS_ROOT="$FIX" bash "$GUARD" >/dev/null 2>&1
  [ $? -eq 1 ] || { fail "case 8: deleting '$lit' from $rel was not caught"; CASE8_OK=0; }
done < <(sed -n "s/^require \([^ ]*\) '\(.*\)'$/\1|\2/p" "$GUARD")
[ "$CASE8_N" -eq 9 ] || { fail "case 8: expected 9 guard rows, saw $CASE8_N — a row was dropped from the guard"; CASE8_OK=0; }
[ "$CASE8_OK" -eq 1 ] && pass "case 8: every guard row's literal deletion is caught ($CASE8_N rows)"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf '%s case(s) passed\n' "$PASS"
