#!/usr/bin/env bash
# Assertion harness for check-self-review-report.sh. Builds fixture files
# under a sandboxed mktemp directory and asserts the guard's exit status and
# a distinctive substring of its output, never one alone, for every case.
# Never touches the real repository tree except in case 7, which by design
# must run the guard bare (see that case's own comment for why).
#
# Modeled on test-check-stage-mark-calls.sh's fixture-driven pattern:
# fixtures live under mktemp -d, the guard is invoked via a thin run_guard
# helper that captures RC/OUT, and every case ends with an explicit
# pass/fail assertion.
#
# THE MUTATION MATRIX (KAN-211). Every classification branch in
# check-self-review-report.sh, plus the post-loop status check on the read
# they all sit inside, was reverted one at a time — the branch's own
# condition forced false, or its violation deleted, with the guard restored
# before the next mutation — this harness re-run after each, and the case
# numbers that failed recorded below. This is the mutation proof KAN-197
# requires of every guard in this repository, written down so a later reader
# can check it against the code instead of taking a commit message's word
# for it. Re-measure it whenever a branch is added, moved or removed; a row
# naming a case that no longer covers its branch is worse than no row.
#
#   branch reverted                            harness cases that failed
#   ------------------------------------------ -------------------------
#   post-loop `read_rc` check                  24
#   `##`-heading label match                   1, 2, 3, 4, 5, 6, 7, 13,
#                                              14, 16, 17, 18, 19, 20, 21
#   other-heading reset (any `#+` level)       12
#   none-marker comparison                     1, 16, 18
#   FINDING_LINE_RE acceptance                 1, 4, 5, 6, 7, 16, 18, 19
#   label-mismatch check                       4
#   disposition split, `declined` arm          1, 7, 18
#   disposition split, neither-arm violation   22
#   empty-issue-key check                      5
#   issue-key shape check                      6, 19
#   loose-shape malformed branch               23 (message assertion only)
#   orphan branch                              12, 15
#   duplicate-section check                    14
#   out-of-order check                         13
#   missing-section check                      2, 17
#   neither-marker-nor-finding check           3, 20, 21
#   both-marker-and-finding check              16
#
# Case 7 appears in several rows because it alone runs the guard bare over
# the repository's real docs/self-review/ corpus, so a broadly-scoped
# reversion shows up there as well as in a fixture case.
#
# TWO ROWS — the disposition neither-arm and the loose-shape malformed
# branch — WERE MEASURED UNCOVERED when this matrix was first written:
# deleting either branch left every one of the harness's then forty-four
# assertions green. Neither was dead code, which was checked rather than
# assumed, so each gained a fixture reaching it directly instead of being
# deleted as a rule nobody wanted. Cases 22 and 23 are those fixtures.
#
# The loose-shape row is annotated `message assertion only` because case
# 23's exit status is NOT a detector for its branch: the same fixture trips
# the post-loop neither-marker-nor-finding check, which holds the exit code
# at 1 with the loose-shape branch removed. Only the message assertion
# fails, and that was observed against a mutated guard, not predicted. Case
# 23's own comment below carries the full reasoning; anything that touches
# its assertions must read it first.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-self-review-report.sh"
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
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/check-self-review-report-test.XXXXXX")"
}

compliant_report() {
  cat <<'EOF'
## Problems and fixes — `myflow-fix`

- **[myflow-fix]** Preflight compares against a stale local base ref — declined

## Cost — `myflow-cost`

- **[myflow-cost]** Every panel slot gathers the same context independently — filed: KAN-201

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
}

# ===========================================================================
# Case 1: a fully compliant report -> exit 0, verdict SELF-REVIEW-REPORT-OK.
# ===========================================================================
new_fixture
compliant_report >"$FIXTURE/fixture-self-review.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 1: compliant report exits 0" || fail "case 1: rc=$RC out=$OUT"
case "$OUT" in
  *"SELF-REVIEW-REPORT-OK"*) pass "case 1: verdict line carries SELF-REVIEW-REPORT-OK" ;;
  *) fail "case 1: expected SELF-REVIEW-REPORT-OK in output, out=$OUT" ;;
esac

# ===========================================================================
# Case 2: the angle-5 section is missing entirely -> named, with the missing
# label, exit 1.
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
## Problems and fixes — `myflow-fix`

- **[myflow-fix]** Preflight compares against a stale local base ref — declined

## Cost — `myflow-cost`

- **[myflow-cost]** Every panel slot gathers the same context independently — filed: KAN-201

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 2: report missing the angle-5 section is caught" || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"missing section"*"myflow-stats-app"*) pass "case 2: finding names the missing angle-5 label" ;;
  *) fail "case 2: expected a missing-section finding naming myflow-stats-app, out=$OUT" ;;
esac

# ===========================================================================
# Case 3: a section carries neither a finding line nor the none-marker ->
# named, exit 1.
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
## Problems and fixes — `myflow-fix`

- **[myflow-fix]** Preflight compares against a stale local base ref — declined

## Cost — `myflow-cost`

- **[myflow-cost]** Every panel slot gathers the same context independently — filed: KAN-201

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

Nothing notable to say here, but no marker either.

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 3: a section with neither a finding line nor the none-marker is caught" || fail "case 3: rc=$RC out=$OUT"
case "$OUT" in
  *"myflow-automation"*"neither a finding line nor the none-marker"*) \
    pass "case 3: finding names the empty automation section" ;;
  *) fail "case 3: expected the empty-section finding naming myflow-automation, out=$OUT" ;;
esac

# ===========================================================================
# Case 4: a finding line's label does not match its section's heading label
# -> named, exit 1.
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
## Problems and fixes — `myflow-fix`

- **[myflow-fix]** Preflight compares against a stale local base ref — declined

## Cost — `myflow-cost`

- **[myflow-fix]** Wrong label under the cost section — declined

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 4: a mismatched finding-line label is caught" || fail "case 4: rc=$RC out=$OUT"
case "$OUT" in
  *"label"*"myflow-fix"*"myflow-cost"*) pass "case 4: finding names the mismatch" ;;
  *) fail "case 4: expected a label-mismatch finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 5: a finding marked filed: with no issue key -> named, exit 1.
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
## Problems and fixes — `myflow-fix`

- **[myflow-fix]** Preflight compares against a stale local base ref — declined

## Cost — `myflow-cost`

- **[myflow-cost]** Every panel slot gathers the same context independently — filed:

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 5: a filed finding with no issue key is caught" || fail "case 5: rc=$RC out=$OUT"
case "$OUT" in
  *"filed"*"no issue key"*) pass "case 5: finding names the missing issue key" ;;
  *) fail "case 5: expected a missing-issue-key finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 6: a finding marked filed: with a NON-EMPTY but malformed issue key
# (not PROJECT-NUMBER) -> named, exit 1. Distinct from case 5 (a fully empty
# key): this is the harness gap F1's own review round found — `filed: yes`,
# `filed: 201` and `filed: KAN-` all used to pass silently.
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
## Problems and fixes — `myflow-fix`

- **[myflow-fix]** Preflight compares against a stale local base ref — declined

## Cost — `myflow-cost`

- **[myflow-cost]** Every panel slot gathers the same context independently — filed: yes

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 6: a filed finding with a malformed issue key is caught" || fail "case 6: rc=$RC out=$OUT"
case "$OUT" in
  *"filed"*"malformed issue key"*"yes"*) pass "case 6: finding names the malformed issue key" ;;
  *) fail "case 6: expected a malformed-issue-key finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 7 (declared pre-rule reports): must run the guard BARE, against the
# real repository docs/self-review/ it ships in. This guard's declared list
# is a hardcoded set of THIS REPOSITORY's own basenames (real, pre-KAN-200
# reports); a sandboxed fixture tree never contains one of those names, so
# declaring them there would make every one of them a KAN-197 F3
# "declared but never recorded" violation for a member simply not part of
# that run's corpus — the same reasoning test-check-stage-mark-calls.sh's own
# case 24 states for the identical shape. Gating declaration on "no CLI
# args" (mirroring check-stage-mark-calls.sh) keeps that protection meaning
# something, and this case proves the real, default invocation is clean.
# ===========================================================================
set +e
REAL_OUT="$("$GUARD" 2>&1)"
REAL_RC=$?
set -e
[ "$REAL_RC" -eq 0 ] && pass "case 7: the real repository's own tree is clean" \
  || fail "case 7: rc=$REAL_RC out=$REAL_OUT"
case "$REAL_OUT" in
  *"SELF-REVIEW-REPORT-OK"*) pass "case 7: verdict line carries SELF-REVIEW-REPORT-OK" ;;
  *) fail "case 7: expected SELF-REVIEW-REPORT-OK in output, out=$REAL_OUT" ;;
esac
case "$REAL_OUT" in
  *"docs/self-review/kan-100-myflow-get-rid-of-staging-use-commits-self-review.md 0 (declared: predates the five-angle report shape (KAN-200))"*) \
    pass "case 7: a declared pre-rule report is reported as declared, with its reason" ;;
  *) fail "case 7: expected a declared pre-rule report named with its reason, out=$REAL_OUT" ;;
esac

# ===========================================================================
# Case 8 (KAN-197 regression shape): a report present in the corpus, absent
# from the declared list, for which the guard performed zero section checks
# at all (no recognizable angle heading anywhere in it) -> named as an
# undeclared zero via scripts/lib/coverage.sh, exit 1. Distinct from case 2
# (a report that IS recognizable, just missing one section): this fixture
# carries none of the five headings, so found_count is 0 and no per-section
# "missing" findings fire at all — only the coverage mechanism names it,
# exactly the shape KAN-73 and KAN-197 exist to catch.
# ===========================================================================
new_fixture
printf 'Ordinary prose with nothing self-review shaped in it at all.\n' \
  >"$FIXTURE/garbage-self-review.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 8: a report checked for nothing is an undeclared coverage violation" \
  || fail "case 8: rc=$RC out=$OUT"
case "$OUT" in
  *"garbage-self-review.md"*"0 checked"*"not declared expected-zero"*) \
    pass "case 8: the violation names the report, its zero count, and that it is undeclared" ;;
  *) fail "case 8: expected the report named as an undeclared zero, out=$OUT" ;;
esac

# ===========================================================================
# Case 9: an empty corpus (a real, readable directory with no *.md file at
# all) is a violation, not a vacuous pass -> exit 1.
# ===========================================================================
new_fixture
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 9: an empty corpus is a violation, not a vacuous pass" || fail "case 9: rc=$RC out=$OUT"
case "$OUT" in
  *"no member was ever recorded"*"checked nothing"*"empty corpus"*) \
    pass "case 9: the verdict names the empty corpus explicitly" ;;
  *) fail "case 9: expected an empty-corpus violation, out=$OUT" ;;
esac

# ===========================================================================
# Case 10: an absent docs/self-review/-shaped directory -> exit 2, distinct
# from exit 1 (the guard cannot answer at all rather than found violations).
# ===========================================================================
ABSENT_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-self-review-report-absent.XXXXXX")/does-not-exist"
run_guard "$ABSENT_DIR"
[ "$RC" -eq 2 ] && pass "case 10: an absent directory exits 2, distinct from 1" || fail "case 10: rc=$RC out=$OUT"
case "$OUT" in
  *"no such directory"*) pass "case 10: the failure names the absent directory" ;;
  *) fail "case 10: expected a no-such-directory message, out=$OUT" ;;
esac

# ===========================================================================
# Case 11 (F2 regression): `find`'s own failure exit status is surfaced as
# exit 2 ("cannot answer"), never folded into "empty corpus" (exit 1). A
# permission-denied-but-readable directory is not reliably constructible in
# every sandbox this harness runs under (search-bit denial is not enforced
# uniformly), so a stub `find` placed first on PATH stands in for that real
# failure: it writes nothing to its output and exits 1, exactly what the
# guard's own `find` call sees from an untraversable target.
# ===========================================================================
new_fixture
cat >"$FIXTURE/find" <<'STUB'
#!/usr/bin/env bash
exit 1
STUB
chmod +x "$FIXTURE/find"
mkdir -p "$FIXTURE/target"
printf 'irrelevant\n' >"$FIXTURE/target/x.md"
PATH="$FIXTURE:$PATH" run_guard "$FIXTURE/target"
[ "$RC" -eq 2 ] && pass "case 11: a failing find exits 2, not folded into empty corpus" \
  || fail "case 11: rc=$RC out=$OUT"
case "$OUT" in
  *"find failed"*) pass "case 11: the failure names find's own error" ;;
  *) fail "case 11: expected a find-failure message, out=$OUT" ;;
esac

# ===========================================================================
# Case 12 (F3 regression): a `###` heading ends the current section just as a
# `##` one does, rather than falling through both checks and being folded
# into the previous section's body. The finding-shaped line that follows the
# `###` here must surface as an orphan (no section open), not be silently
# absorbed into `myflow-fix`'s own none-marker above it.
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
## Problems and fixes — `myflow-fix`

_none — this angle produced no findings._

### An h3 aside inside the fix section

- **[myflow-fix]** must not be folded into the section above — declined

## Cost — `myflow-cost`

_none — this angle produced no findings._

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 12: a ### heading resets the current section" || fail "case 12: rc=$RC out=$OUT"
case "$OUT" in
  *"finding line appears before any recognized section heading"*) \
    pass "case 12: the line after the ### is named an orphan, not folded into myflow-fix" ;;
  *) fail "case 12: expected an orphan finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 13 (F4 regression): two present, complete, otherwise-valid sections in
# the wrong relative order (`myflow-cost` before `myflow-fix`) are named as
# out of order. Every section is otherwise well-formed and all five are
# present, so this is the ONLY defect the fixture carries — nothing else
# would trip exit 1 if the ordering check itself were missing.
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
## Cost — `myflow-cost`

_none — this angle produced no findings._

## Problems and fixes — `myflow-fix`

_none — this angle produced no findings._

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 13: an out-of-order section is caught" || fail "case 13: rc=$RC out=$OUT"
case "$OUT" in
  *"myflow-fix"*"out of order"*"myflow-cost"*) pass "case 13: finding names the out-of-order section" ;;
  *) fail "case 13: expected an out-of-order finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 14 (F5 regression): the same angle's `##` heading appears twice. Both
# occurrences are otherwise well-formed (a bare none-marker each), so nothing
# else in this fixture would trip exit 1 if the duplicate-section check
# itself were missing.
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
## Problems and fixes — `myflow-fix`

_none — this angle produced no findings._

## Problems and fixes — `myflow-fix`

_none — this angle produced no findings._

## Cost — `myflow-cost`

_none — this angle produced no findings._

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 14: a duplicate section is caught" || fail "case 14: rc=$RC out=$OUT"
case "$OUT" in
  *"duplicate section for angle"*"myflow-fix"*) pass "case 14: finding names the duplicate section" ;;
  *) fail "case 14: expected a duplicate-section finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 15 (F6 regression): a finding-shaped line appears before any
# recognized heading at all — the shape this guard's own header calls out as
# "previously dropped in silence".
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
- **[myflow-fix]** appears before any heading — declined

## Problems and fixes — `myflow-fix`

_none — this angle produced no findings._

## Cost — `myflow-cost`

_none — this angle produced no findings._

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 15: a pre-heading finding line is caught" || fail "case 15: rc=$RC out=$OUT"
case "$OUT" in
  *"finding line appears before any recognized section heading"*) \
    pass "case 15: finding names the pre-heading line" ;;
  *) fail "case 15: expected a pre-heading finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 16 (F7 regression): a section carries BOTH the none-marker and a real
# finding line — mutually exclusive per the report shape.
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
## Problems and fixes — `myflow-fix`

_none — this angle produced no findings._

- **[myflow-fix]** a finding alongside the none-marker above — declined

## Cost — `myflow-cost`

_none — this angle produced no findings._

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 16: a none-marker alongside a finding line is caught" || fail "case 16: rc=$RC out=$OUT"
case "$OUT" in
  *"carries both the none-marker and"*"finding line"*) \
    pass "case 16: finding names the none-marker-plus-finding conflict" ;;
  *) fail "case 16: expected a both-marker-and-finding finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 17 (F8 regression): a file whose BASENAME matches one of the
# seventeen hardcoded declared pre-rule names, scanned via an EXPLICIT
# directory argument (not this guard's bare, default invocation) — the one
# mode where declare_pre_rule never runs. The declared-basename fast path
# must not apply here: content is missing its fifth section, and that
# SPECIFIC violation must still be named, proving the guard actually scanned
# the file rather than waving it through on basename alone.
# ===========================================================================
new_fixture
cat >"$FIXTURE/kan-73-install-guard-scripts-alongside-skills-self-review.md" <<'EOF'
## Problems and fixes — `myflow-fix`

- **[myflow-fix]** Preflight compares against a stale local base ref — declined

## Cost — `myflow-cost`

- **[myflow-cost]** Every panel slot gathers the same context independently — filed: KAN-201

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 17: a declared basename is still content-checked outside bare mode" \
  || fail "case 17: rc=$RC out=$OUT"
case "$OUT" in
  *"missing section"*"myflow-stats-app"*) \
    pass "case 17: the missing-section finding fired, proving the file was actually scanned" ;;
  *) fail "case 17: expected a missing-section finding despite the matching basename, out=$OUT" ;;
esac

# ===========================================================================
# Case 18 (F9 regression): a CRLF-saved report is read the same as an LF one.
# Every line below carries a trailing `\r`; without stripping it, the exact
# string comparison against the none-marker fails on every section, and
# "neither a finding line nor the none-marker" fires for all five.
# ===========================================================================
new_fixture
compliant_report | sed $'s/$/\r/' >"$FIXTURE/fixture-self-review.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 18: a CRLF-saved compliant report exits 0" || fail "case 18: rc=$RC out=$OUT"
case "$OUT" in
  *"SELF-REVIEW-REPORT-OK"*) pass "case 18: verdict line carries SELF-REVIEW-REPORT-OK despite CRLF" ;;
  *) fail "case 18: expected SELF-REVIEW-REPORT-OK despite CRLF, out=$OUT" ;;
esac

# ===========================================================================
# Case 19 (G3 regression, round 3 Primary/H1): a REAL trailing tab byte (not
# a `\t` escape sequence a heredoc would render as two characters) on a
# `filed: <KEY>` line must be caught as a malformed issue key. Fix round 2's
# own protocol comment claims this exact scenario is what the US-delimited
# record protocol fixes — but round 2 added no case exercising a literal tab
# byte anywhere, so a revert back to `IFS=$'\t' read` left every one of its
# 18 cases green (verified independently by round 3's Principles slot). The
# fixture is built from `compliant_report`'s own `filed: KAN-201` line with
# one genuine tab byte appended via ANSI-C quoting (the same technique case
# 18 already uses for a real `\r`), and its presence is asserted before the
# guard ever runs so this case cannot silently degrade into a no-op.
# ===========================================================================
new_fixture
compliant_report | sed $'s/filed: KAN-201$/filed: KAN-201\t/' >"$FIXTURE/fixture-self-review.md"
grep -q $'\t' "$FIXTURE/fixture-self-review.md" \
  || fail "case 19: fixture setup — trailing tab byte not present in fixture"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 19: a trailing literal tab on a filed line is caught" \
  || fail "case 19: rc=$RC out=$OUT"
case "$OUT" in
  *"filed"*"malformed issue key"*) \
    pass "case 19: the trailing tab is carried into the key and named malformed" ;;
  *) fail "case 19: expected a malformed-issue-key finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 20 (G3 regression, round 3 Primary/H1): a body line beginning with a
# REAL literal tab byte must not be silently recognized as a compliant
# finding. The fixture's `myflow-fix` section carries only a tab-indented,
# otherwise well-formed finding line and no none-marker — under the correct
# US-delimited protocol the leading tab means the line matches neither
# FINDING_LINE_RE nor FINDING_SHAPE_RE (both require the line to start with
# `-`), so the section has zero recognized findings and no none-marker and
# must be flagged. Under the reverted `IFS=$'\t' read` protocol the leading
# tab is trimmed as IFS whitespace, the line reads as if un-indented, and it
# is silently accepted as a well-formed finding — masking exactly the defect
# this case exists to catch.
# ===========================================================================
new_fixture
{
  printf '%s\n\n' '## Problems and fixes — `myflow-fix`'
  printf '%s\n\n' $'\t- **[myflow-fix]** indented finding must not be recognized — declined'
  printf '%s\n\n' '## Cost — `myflow-cost`'
  printf '%s\n\n' '_none — this angle produced no findings._'
  printf '%s\n\n' '## What went well — `myflow-improvement`'
  printf '%s\n\n' '_none — this angle produced no findings._'
  printf '%s\n\n' '## Automation — `myflow-automation`'
  printf '%s\n\n' '_none — this angle produced no findings._'
  printf '%s\n\n' '## Stats app — `myflow-stats-app`'
  printf '%s\n' '_none — this angle produced no findings._'
} >"$FIXTURE/fixture-self-review.md"
grep -q $'\t' "$FIXTURE/fixture-self-review.md" \
  || fail "case 20: fixture setup — leading tab byte not present in fixture"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 20: a tab-indented finding-shaped line is not silently accepted" \
  || fail "case 20: rc=$RC out=$OUT"
case "$OUT" in
  *'`myflow-fix`'*"neither a finding line nor the none-marker"*) \
    pass "case 20: the fix section is flagged as neither a finding nor the none-marker" ;;
  *) fail "case 20: expected a neither-finding-nor-none-marker finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 21 (byte fidelity, alongside cases 19 and 20): a REAL 0x1F (ASCII Unit
# Separator) byte inside report content reaches the classifier byte for byte,
# so a line differing from the none-marker only by a trailing 0x1F is NOT
# equal to the none-marker and does not silently satisfy its section. The
# `myflow-improvement` section's none-marker line here carries one genuine
# trailing 0x1F byte, appended via ANSI-C quoting; that section therefore
# carries neither a finding line nor the none-marker and is flagged for that
# reason.
#
# 0x1F is ordinary report content now — it is not itself a violation. It was
# one while the guard serialized its classification as 0x1F-delimited records
# for bash to re-parse (a trailing 0x1F was trimmed by `read` exactly as a
# trailing tab is in cases 19 and 20, silently forgiving the line); KAN-211
# deleted that protocol, and the rule went with it. What this case still
# proves is the same thing it always did: a trailing control byte does not
# buy a non-compliant line a pass.
#
# WHY THIS CASE CARRIES TWO MESSAGE ASSERTIONS, and what each one is for.
# The first states the collapsed guard's behaviour: the section is flagged
# as carrying neither a finding line nor the none-marker. The second — that
# `unsupported control byte` appears nowhere in the output — is the one that
# catches a REVERT of KAN-211, and it is the only assertion in this harness
# that can. That is a measured fact, not a precaution: run against the
# pre-change guard, this fixture produces BOTH violations, the control-byte
# one at the none-marker's own line and the section-level one at the heading,
# because `CTRLBYTE` consumed the none-marker line before it could be
# classified and the post-loop neither-marker-nor-finding check then fired
# for that section anyway. So the first assertion passes in both states and
# would let a restored record protocol through in silence. Asserting the
# retired violation's ABSENCE is what makes the revert loud. Keep both:
# neither one is redundant with the other.
# ===========================================================================
new_fixture
{
  printf '%s\n\n' '## Problems and fixes — `myflow-fix`'
  printf '%s\n\n' '_none — this angle produced no findings._'
  printf '%s\n\n' '## Cost — `myflow-cost`'
  printf '%s\n\n' '_none — this angle produced no findings._'
  printf '%s\n\n' '## What went well — `myflow-improvement`'
  printf '%s\n\n' $'_none — this angle produced no findings._\037'
  printf '%s\n\n' '## Automation — `myflow-automation`'
  printf '%s\n\n' '_none — this angle produced no findings._'
  printf '%s\n\n' '## Stats app — `myflow-stats-app`'
  printf '%s\n' '_none — this angle produced no findings._'
} >"$FIXTURE/fixture-self-review.md"
grep -q $'\037' "$FIXTURE/fixture-self-review.md" \
  || fail "case 21: fixture setup — control byte not present in fixture"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 21: a raw 0x1F byte in report content is caught, not silently swallowed" \
  || fail "case 21: rc=$RC out=$OUT"
case "$OUT" in
  *'`myflow-improvement`'*"neither a finding line nor the none-marker"*) \
    pass "case 21: the improvement section is flagged as neither a finding nor the none-marker" ;;
  *) fail "case 21: expected a neither-finding-nor-none-marker finding, out=$OUT" ;;
esac
case "$OUT" in
  *"unsupported control byte"*) \
    fail "case 21: the retired 0x1F violation is back — a record protocol has been reintroduced, out=$OUT" ;;
  *) pass "case 21: the retired unsupported-control-byte violation is absent" ;;
esac

# ===========================================================================
# Case 22 (KAN-211, mutation-matrix gap): a well-formed finding line whose
# disposition is neither `filed: <KEY>` nor `declined`. The fixture is
# `compliant_report` with the `myflow-cost` finding's disposition replaced by
# `maybe later` — everything else about the line, and about the report, stays
# compliant. The line still matches FINDING_LINE_RE, so FINDING_COUNT
# increments and the section is not flagged for anything else; the ONLY
# defect the fixture carries is the disposition. Delete the neither-arm and
# the guard drops to exit 0 on this fixture, which is what makes the
# exit-status assertion below a genuine detector for that branch.
# ===========================================================================
new_fixture
compliant_report | sed 's/filed: KAN-201$/maybe later/' >"$FIXTURE/fixture-self-review.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 22: an unrecognized finding-line disposition is caught" \
  || fail "case 22: rc=$RC out=$OUT"
case "$OUT" in
  *'finding line disposition is neither `filed: <KEY>` nor `declined`'*) \
    pass "case 22: the violation names the unrecognized disposition rule" ;;
  *) fail "case 22: expected an unrecognized-disposition finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 23 (G2 regression, KAN-211 mutation-matrix gap): a finding-shaped line
# that is malformed — TWO spaces after the leading `-`, so it matches the
# loose FINDING_SHAPE_RE but not the strict FINDING_LINE_RE. This is the
# branch KAN-200 added as its G2 fix, precisely so such a line would stop
# being silently dropped: before it, the line was not counted as a finding,
# not flagged, and invisible to every check below. It has been untested since
# that fix landed, which is why this case exists.
#
# WHY THE MESSAGE ASSERTION IS THE DETECTOR HERE, not the exit status. The
# malformed line leaves its `myflow-fix` section with no recognized finding
# and no none-marker, so the post-loop neither-marker-nor-finding check fires
# on this same fixture and holds the exit code at 1 EVEN WITH the loose-shape
# branch deleted. An assertion on exit status alone therefore passes in both
# states and proves nothing about this branch. Asserting the malformed-line
# message itself is what makes the branch's removal loud — the same
# discipline case 21 documents for its own assertion pair. Keep both: the
# exit-status assertion states the guard's contract, the message assertion
# is the mutation detector.
# ===========================================================================
new_fixture
cat >"$FIXTURE/fixture-self-review.md" <<'EOF'
## Problems and fixes — `myflow-fix`

-  **[myflow-fix]** two spaces after the dash is malformed — declined

## Cost — `myflow-cost`

_none — this angle produced no findings._

## What went well — `myflow-improvement`

_none — this angle produced no findings._

## Automation — `myflow-automation`

_none — this angle produced no findings._

## Stats app — `myflow-stats-app`

_none — this angle produced no findings._
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 23: a malformed finding-shaped line is not silently dropped" \
  || fail "case 23: rc=$RC out=$OUT"
case "$OUT" in
  *"finding-shaped line is malformed"*) \
    pass "case 23: the violation names the line as finding-shaped but malformed" ;;
  *) fail "case 23: expected a malformed finding-shaped-line finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 24 (KAN-211 round-1 F1 regression): a report file the guard cannot
# open. The fixture is an otherwise fully compliant report, `chmod 000` so
# the loop's own `done < "$f"` redirection fails. This is an I/O failure, not
# report content, so it must reach exit 2 ("cannot answer") with a die naming
# the file — never exit 1, which would classify an unread file as a content
# violation.
#
# It is a regression case: collapsing the awk boundary removed the one place
# that checked a per-file read for failure. `scan_report "$f"` returned awk's
# non-zero status and the guard died on it; reading the file directly instead
# left bash printing "Permission denied" to stderr while every per-report
# counter stayed at its reset zero, so the file was recorded with count 0 and
# surfaced as an undeclared-zero coverage violation at exit 1 — the exact
# "reassuring empty result" this guard's header says it does not produce.
#
# BOTH assertions are detectors here, unlike cases 21 and 23. Delete the
# guard's `read_rc` check after `done < "$f"` and the exit status drops from
# 2 to 1 AND the die message disappears, because coverage.sh reports the zero
# count instead. The permission bits are restored before the case ends so the
# fixture directory stays deletable.
#
# What this case does NOT cover, and no case in this harness can: a read that
# fails partway through a file that opened successfully — see **Known limits**
# in KAN-211's `design.md` for why, which is canonical for it.
# ===========================================================================
new_fixture
compliant_report >"$FIXTURE/fixture-self-review.md"
chmod 000 "$FIXTURE/fixture-self-review.md"
run_guard "$FIXTURE"
chmod 644 "$FIXTURE/fixture-self-review.md"
[ "$RC" -eq 2 ] && pass "case 24: an unreadable report file is 'cannot answer', not a violation" \
  || fail "case 24: rc=$RC out=$OUT"
case "$OUT" in
  *"cannot read report file"*"fixture-self-review.md"*) \
    pass "case 24: the die message names the unreadable file" ;;
  *) fail "case 24: expected a cannot-read-report-file die naming the file, out=$OUT" ;;
esac

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
