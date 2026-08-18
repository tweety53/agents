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
# `###` here must surface as an ORPHAN (no section open), not be silently
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
# Case 21 (round 3 Code review/H2): a REAL 0x1F (ASCII Unit Separator) byte
# inside report content must be a named violation, never silently swallowed.
# `read`'s trailing-delimiter trim (the same mechanism case 19/20 exploit for
# tab) applies to 0x1F too: a none-marker line differing from the real one
# only by a trailing 0x1F would otherwise be accepted as compliant. The
# `myflow-improvement` section's none-marker line here carries one genuine
# trailing 0x1F byte, appended via ANSI-C quoting.
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
  *"unsupported control byte"*) pass "case 21: the violation names the unsupported control byte" ;;
  *) fail "case 21: expected an unsupported-control-byte finding, out=$OUT" ;;
esac

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
