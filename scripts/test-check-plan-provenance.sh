#!/usr/bin/env bash
# Assertion harness for check-plan-provenance.sh. Builds fixture trees under a
# sandboxed TMPDIR and asserts the guard's exit status and output. Never
# touches the real repository tree.
#
# READ THIS BEFORE ADDING OR "FIXING" A FIXTURE. This suite has four times
# encoded one of the guard's own defects as its specification — asserting the
# buggy output as correct, which then made the defect look verified and hid it
# through several further review passes. Check a fixture against the CommonMark
# spec, with the section cited in a comment, before believing it has found a
# bug; observed output is never the justification for an assertion.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-plan-provenance.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# run_guard <fixture_dir> -> sets RC and OUT
#
# Points the guard at the fixture via CHECK_PLAN_PROVENANCE_ROOT rather than
# `cd`-ing into it. The guard resolves REPO_ROOT from its own BASH_SOURCE by
# default (so a real invocation from any cwd still scans this repo, never
# silently nothing) — CHECK_PLAN_PROVENANCE_ROOT is the one explicit, opt-in
# override that lets this harness exercise it against a sandboxed fixture
# tree without that real-repo default getting in the way.
run_guard() {
  set +e
  OUT="$(CHECK_PLAN_PROVENANCE_ROOT="$1" "$GUARD" 2>&1)"
  RC=$?
  set -e
}

new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-test.XXXXXX")"
  mkdir -p "$FIXTURE/openspec/changes/demo-change" \
    "$FIXTURE/openspec/changes/archive/old-change" \
    "$FIXTURE/skills/foo"
}

# ===========================================================================
# SECTION: Provenance syntax — verified:/unverified: tag vocabulary
# (resumes at case 21 with the fence-tag left-boundary check)
# ===========================================================================

# 1. A block tagged verified: passes.
new_fixture
{
  printf '```bash verified:ran it locally\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "verified: block passes" || fail "verified: block: rc=$RC out=$OUT"

# 2. A block tagged unverified: passes.
new_fixture
{
  printf '```bash unverified:confirm the flag name\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "unverified: block passes" || fail "unverified: block: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Fence detection — a fenced block with no tag is flagged
# (resumes at case 10 with nested-fence handling, case 14 with indentation,
# and case 22 onward with full container-prefix parsing)
# ===========================================================================

# 3. An untagged fenced block fails, naming file and line.
new_fixture
{
  printf 'intro\n'
  printf '```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "untagged block fails" || fail "untagged block: expected non-zero"
case "$OUT" in
  *"openspec/changes/demo-change/tasks.md:2"*) pass "untagged block reports file:line" ;;
  *) fail "untagged block file:line: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Claim boundaries — numeric provenance (measured:/predicted:)
# (resumes at case 11 with over-firing guards and the lookahead window, and
# case 15 with provenance-comment negation and CLAIM_RE's left boundary)
# ===========================================================================

# 4. A numeric claim with a measured: comment on the next line passes.
new_fixture
{
  printf 'Baseline: 197 tests\n'
  printf '<!-- measured: ./gradlew test @ c515c42 -->\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "measured: comment satisfies numeric claim" || fail "measured: out=$OUT"

# 5. A numeric claim with no provenance comment fails.
new_fixture
printf 'Baseline: 197 tests\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "unattributed numeric claim fails" || fail "unattributed numeric claim: expected non-zero"
case "$OUT" in
  *"openspec/changes/demo-change/tasks.md:1"*) pass "unattributed claim reports file:line" ;;
  *) fail "unattributed claim file:line: out=$OUT" ;;
esac

# 6. A numeric claim with a predicted: comment passes.
new_fixture
{
  printf 'After the deletion: 186 tests\n'
  printf '<!-- predicted: ./gradlew test after task 1 -->\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "predicted: comment satisfies numeric claim" || fail "predicted: out=$OUT"

# A clean, tagged tasks.md, reused by cases 7-9 so each of them exercises
# ONLY the out-of-scope behavior under test — not the "zero tasks.md
# scanned" refusal (case 13 below), which would otherwise mask it with a
# false rc=2.
clean_tasks_md() {
  printf '```bash verified:ran it locally\necho hi\n```\n' \
    > "$FIXTURE/openspec/changes/demo-change/tasks.md"
}

# ===========================================================================
# SECTION: Scope — which files/directories the guard scans
# (resumes at case 17 with the "zero non-archived changes" steady state)
# ===========================================================================

# 7. An untagged block under openspec/changes/archive/ is not scanned.
new_fixture
clean_tasks_md
{
  printf '```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/archive/old-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "archived tasks.md is not scanned" || fail "archive: out=$OUT"

# 8. An untagged block in proposal.md (not tasks.md) is out of scope.
new_fixture
clean_tasks_md
{
  printf '```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/proposal.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "proposal.md is out of scope" || fail "proposal.md: out=$OUT"

# 9. An untagged block in skills/foo/SKILL.md is out of scope.
new_fixture
clean_tasks_md
{
  printf '```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/skills/foo/SKILL.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "SKILL.md is out of scope" || fail "SKILL.md: out=$OUT"

# ===========================================================================
# SECTION: Fence detection (cont'd) — nested/quoted fence boundaries
# ===========================================================================

# 10. A tagged 4-backtick block containing a 3-backtick block passes — the
# inner 3-backtick fence must not be miscounted as a new block boundary, and
# the outer 4-backtick close must be recognised despite the shorter fence
# nested inside it. This is the CommonMark same-character/at-least-as-long
# closing rule.
new_fixture
{
  printf '````markdown verified:authored in-tree for this change\n'
  printf '```kotlin verified:javap intellij.platform.diff.jar\n'
  printf 'override val toolWindowIds: Array<String>\n'
  printf '```\n'
  printf '````\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "4-backtick block containing a 3-backtick block passes" \
  || fail "nested fence: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Claim boundaries (cont'd) — over-firing guards and the lookahead
# window
# ===========================================================================

# 11. Deliberately-not-matched numeric-looking tokens must not trigger a
# false positive: step numbers, line numbers, ticket IDs, version numbers.
# This is the over-firing guard against the failure mode the brief calls out
# by name.
new_fixture
{
  printf 'See step 1.2 for details.\n'
  printf 'Error reported at line 452.\n'
  printf 'Tracked as IU-262 and KAN-14.\n'
  printf 'Upgrade to version 2.1.0 of the library.\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "step/line/ticket/version numbers do not over-fire" \
  || fail "over-firing: rc=$RC out=$OUT"

# 11b. A unit word must not match as a mere PREFIX of a longer word — the
# right-hand boundary must be a non-word character or end of line. Without
# it, "10 filesystems", "5 minuteswalk", and "10 errorsome" all read as
# numeric claims that are then unattributed, and fail. This is the same
# over-firing class check-references.sh's header names (28 false failures)
# and is a RIGHT-side collision, distinct from the LEFT-side collisions in
# case 11 above.
new_fixture
printf 'We migrated 10 filesystems last week.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "'10 filesystems' does not match 'files' as a prefix" \
  || fail "files-prefix over-fire: rc=$RC out=$OUT"

new_fixture
printf 'It took 5 minuteswalk to get there.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "'5 minuteswalk' does not match 'minutes' as a prefix" \
  || fail "minutes-prefix over-fire: rc=$RC out=$OUT"

new_fixture
printf 'Found 10 errorsome behaviors.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "'10 errorsome' does not match 'errors' as a prefix" \
  || fail "errors-prefix over-fire: rc=$RC out=$OUT"

new_fixture
printf 'The batch processed 20 testsuite runs overnight.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "'20 testsuite' does not match 'tests' as a prefix" \
  || fail "tests-prefix over-fire: rc=$RC out=$OUT"

# 11c. The lookahead window: a provenance comment exactly 2 lines after the
# claim satisfies it; one 3 lines after does not. Pins the documented
# boundary so it cannot silently drift — other tasks depend on this window.
new_fixture
{
  printf 'Baseline: 197 tests\n'
  printf 'a blank note line\n'
  printf '<!-- measured: ./gradlew test @ c515c42 -->\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a provenance comment 2 lines after the claim satisfies it" \
  || fail "2-lines-away: rc=$RC out=$OUT"

new_fixture
{
  printf 'Baseline: 197 tests\n'
  printf 'a blank note line\n'
  printf 'another blank note line\n'
  printf '<!-- measured: ./gradlew test @ c515c42 -->\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a provenance comment 3 lines after the claim does not satisfy it" \
  || fail "3-lines-away: expected non-zero, out=$OUT"

# ===========================================================================
# SECTION: Exit codes — environment (exit 2)
# (resumes at case 18 with a missing openspec/changes/)
# ===========================================================================

# 12. A root that is not a directory exits non-zero with an error, never a
# false "clean" — mirrors check-references.sh's contract.
set +e
OUT="$(CHECK_PLAN_PROVENANCE_ROOT="${TMPDIR:-/tmp}/plan-prov-does-not-exist-$$" "$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "a nonexistent CHECK_PLAN_PROVENANCE_ROOT exits 2" \
  || fail "nonexistent root: rc=$RC out=$OUT"

# 13. A tree with an openspec/changes/ directory but zero matching tasks.md
# (e.g. one sitting at the wrong depth) must never report a clean run — a
# false "0 files, all provenance stated, exit 0" would mask a broken find
# glob or a misconfigured root exactly the way check-references.sh's own
# "no Markdown files found" refusal exists to prevent.
new_fixture
mkdir -p "$FIXTURE/openspec/changes/demo-change/nested"
{
  printf '```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/nested/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 2 ] && pass "zero matching tasks.md refuses to report a clean run" \
  || fail "zero scanned: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Fence detection (cont'd) — indentation
# ===========================================================================

# 14. A fence indented under a list item (2-space content column, the shape
# this repo's own kan-8-myflow-updates/tasks.md uses 52 times) is entered and
# its untagged content is flagged — not scanned as invisible prose. Pins the
# fix for the guard's headline defect: indented fences were previously never
# detected at all.
new_fixture
{
  printf '1. Do the thing:\n'
  printf '  ```bash\n'
  printf '  echo hi\n'
  printf '  ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "an indented untagged fence is detected" \
  || fail "indented fence: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:2"*) pass "indented fence reports the correct line" ;;
  *) fail "indented fence line number: out=$OUT" ;;
esac

# 14b. An indented fence that IS tagged passes — indentation alone is not a
# violation, only a missing tag is.
new_fixture
{
  printf '1. Do the thing:\n'
  printf '  ```bash verified:ran it locally\n'
  printf '  echo hi\n'
  printf '  ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "an indented tagged fence passes" \
  || fail "indented tagged fence: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Claim boundaries (cont'd) — provenance-comment negation and
# CLAIM_RE's left boundary
# ===========================================================================

# 15. A comment that NAMES the provenance words while denying them must not
# satisfy the numeric rule — PROVENANCE_RE needs a left boundary anchored on
# the comment opener, not a bare substring search.
new_fixture
{
  printf 'Baseline: 197 tests\n'
  printf '<!-- not actually measured: just a guess -->\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "'not actually measured:' does not satisfy provenance" \
  || fail "negated measured: expected non-zero, out=$OUT"

new_fixture
{
  printf 'Baseline: 197 tests\n'
  printf '<!-- unmeasured: revisit -->\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "'unmeasured:' does not satisfy provenance" \
  || fail "unmeasured: expected non-zero, out=$OUT"

# 16. CLAIM_RE's LEFT boundary: a digit group that is merely the tail of an
# alphanumeric token is not a numeric claim. "sha256 tests" and "utf8 lines"
# must not over-fire.
new_fixture
printf 'Run sha256 tests to confirm the binary.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "'sha256 tests' does not over-fire" \
  || fail "sha256 over-fire: rc=$RC out=$OUT"

new_fixture
printf 'Normalize utf8 lines before comparing.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "'utf8 lines' does not over-fire" \
  || fail "utf8 over-fire: rc=$RC out=$OUT"

new_fixture
printf 'Track ticket KAN256 for the follow-up.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "'KAN256' (no unit word) does not over-fire" \
  || fail "KAN256 over-fire: rc=$RC out=$OUT"

# A genuine, boundary-satisfying numeric claim must still fire — the
# left-boundary fix must not turn CLAIM_RE toothless.
new_fixture
printf 'It found 256 failures overnight.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a genuine boundary-satisfying numeric claim still fires" \
  || fail "genuine claim: expected non-zero, out=$OUT"

# ===========================================================================
# SECTION: Scope (cont'd) — "nothing in flight" steady state
# ===========================================================================

# 17. Zero non-archived changes under an existing openspec/changes/ is this
# repository's ordinary steady state between /myflow-finish and the next
# /myflow-start, and must exit 0 — not the exit-2 "cannot determine" refusal
# reserved for a missing/unreadable changes directory.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "zero non-archived changes exits 0 (nothing in flight)" \
  || fail "zero non-archived changes: rc=$RC out=$OUT"

# 17b. Same shape, but openspec/changes/ holds only the archive/ directory —
# still zero non-archived changes, still exit 0.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/archive/old-change"
printf '```bash\necho hi\n```\n' > "$FIXTURE/openspec/changes/archive/old-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "only archive/ present still exits 0" \
  || fail "only-archive: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Exit codes (cont'd) — environment (exit 2)
# ===========================================================================

# 18. Missing openspec/changes/ entirely is still "cannot determine" — exit 2
# — distinct from case 17/17b above.
new_fixture
rm -rf "$FIXTURE/openspec"
run_guard "$FIXTURE"
[ "$RC" -eq 2 ] && pass "missing openspec/changes/ exits 2 (cannot determine)" \
  || fail "missing changes dir: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Containment / security — symlink escapes (exit 3)
# (resumes at case 37 with the quadratic-DoS regression guard)
# ===========================================================================

# 19. A symlinked change directory is not a scan bypass: `find -L` descends
# into it, so its content is opened and any violation is flagged.
new_fixture
mkdir -p "$FIXTURE/real-changes/evil-change"
printf '```bash\necho hi\n```\n' > "$FIXTURE/real-changes/evil-change/tasks.md"
rm -rf "$FIXTURE/openspec/changes/demo-change"
ln -s "$FIXTURE/real-changes/evil-change" "$FIXTURE/openspec/changes/evil-change"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a symlinked change directory is not bypassed" \
  || fail "symlinked change dir bypass: expected non-zero, out=$OUT"

# 20. A `tasks.md` that is ITSELF a symlink is refused loudly (exit 3) rather
# than opened directly — an out-of-tree read at best, an unbounded-read hang
# at worst (a symlink to /dev/zero never yields EOF).
new_fixture
ln -sf /dev/zero "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "a symlinked tasks.md is refused, not opened (exit 3, containment)" \
  || fail "symlinked tasks.md: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Provenance syntax (cont'd) — fence-tag left boundary
# ===========================================================================

# 21. Fence-tag boundary: `preverified:` must not satisfy the tag rule as a
# substring of `verified:`.
new_fixture
{
  printf '```bash preverified:not a real tag\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "'preverified:' does not satisfy the fence-tag rule" \
  || fail "preverified: expected non-zero, out=$OUT"

# ===========================================================================
# SECTION: Fence detection (cont'd) — container-prefix parsing (blockquote,
# list marker, checkbox, doubled markers, prose-eligibility, compound
# numbering). This is the bulk of the guard's fail-loud abort-path coverage.
# ===========================================================================

# 22. A fence inside a blockquote is entered, not read as prose — the
# blockquote-marker gap from pass 2 of the review panel. An untagged block
# must be flagged, not shipped silently.
new_fixture
{
  printf '> ```bash\n'
  printf '> rm -rf /\n'
  printf '> ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a blockquoted untagged fence is detected" \
  || fail "blockquoted fence: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:1"*) pass "blockquoted fence reports the correct line" ;;
  *) fail "blockquoted fence line number: out=$OUT" ;;
esac

# 22b. The same shape, tagged, passes — the blockquote marker is not itself
# a violation, only a missing tag is.
new_fixture
{
  printf '> ```bash verified:ran it locally\n'
  printf '> echo hi\n'
  printf '> ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a blockquoted tagged fence passes" \
  || fail "blockquoted tagged fence: rc=$RC out=$OUT"

# 23. A fence opened on a list-bullet line must not be missed, and — the
# actual regression this pins — its closing fence must not be misread as a
# fresh opener that then swallows every later violation in the file. A real
# unattributed numeric claim two lines after the (correctly recognised)
# close must still be reported.
new_fixture
{
  printf -- '- ```bash\n'
  printf '  echo hi\n'
  printf '  ```\n'
  printf 'Baseline: 42 tests\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a bulleted untagged fence is detected" \
  || fail "bulleted fence: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
    pass "bulleted fence reports the opening line, not the close" ;;
  *) fail "bulleted fence line number: out=$OUT" ;;
esac
case "$OUT" in
  *"tasks.md:4: numeric claim with no measured:/predicted: provenance comment"*) \
    pass "a claim after the bulleted fence's close is still reported, not swallowed" ;;
  *) fail "claim after bulleted fence: out=$OUT" ;;
esac

# 24. A doubled list marker on the fence line (a bullet nested directly
# under another bullet, both on the fence line itself) is a prefix shape
# this guard does not parse. Per the fail-loud principle it must abort
# rather than silently read the line as prose. At column 0, indentation is
# never the cause (Important 4, pass-10 fix wave) — the remedy must name
# the actual causes (a doubled/glued marker, a bare checkbox with no list
# marker, or a too-wide gap) and never tell the author to dedent, since
# there is nothing here to dedent.
new_fixture
{
  printf -- '- - ```bash\n'
  printf '  echo hi\n'
  printf '  ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "a doubled list marker on a fence line aborts loudly (exit 4, content-classification)" \
  || fail "doubled list marker: expected rc=4, got rc=$RC out=$OUT"
case "$OUT" in
  *"dedent"*) fail "doubled marker at column 0: remedy wrongly says dedent: out=$OUT" ;;
  *"doubled"*"marker"*) pass "doubled marker at column 0: remedy names the doubled marker, never dedent" ;;
  *) fail "doubled marker at column 0: remedy does not name the marker: out=$OUT" ;;
esac

# 25. A doubled ordered-list marker on the fence line, same principle as 24.
new_fixture
{
  printf '1. 2. ```bash\n'
  printf '   echo hi\n'
  printf '   ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "a doubled ordered-list marker on a fence line aborts loudly (exit 4, content-classification)" \
  || fail "doubled ordered marker: expected rc=4, got rc=$RC out=$OUT"

# 25b. Glued-together dashes with NO separating whitespace (`--`) are NOT a
# doubled list marker under this grammar — a marker is one bullet character
# immediately followed by whitespace, and `--` never matches that pattern
# even once. This is deliberately DIFFERENT from case 24/25 above (`- - `,
# two complete, whitespace-terminated markers): `--` fails eligibility
# entirely and is read as ordinary prose, so it opens no fence itself. The
# bare fence run that follows on its own line, later in the file, IS an
# ordinary unprefixed opener — and is left unclosed here on purpose, so the
# guard's genuine "fenced code block never closed" path is what reports it
# (rc=1), not the fence-classification abort (rc=4). Pins the corrected
# reading after the prior round's report wrongly claimed this still aborts.
new_fixture
{
  printf -- '-- ```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "glued '--' with no space is prose, not a doubled marker; the later bare fence is reported unclosed (rc=1)" \
  || fail "glued dashes as prose: expected rc=1, got rc=$RC out=$OUT"

# 26. THE critical regression this round exists to fix: a fence opened on a
# task-checkbox line (`- [ ] ```bash`, the shape every task line in this
# repo's own tasks.md files uses) tagged verified: passes cleanly, its close
# is recognised as a close (not misread as a fresh untagged opener), and a
# real unattributed numeric claim after it is still reported — nothing after
# the checkbox-fence is swallowed. Continuation/closer indented to the TRUE
# content column `- [ ] ` establishes (6 columns — pass-9 fix wave, Critical
# 1; the flat 2-column indentation this fixture used before that fix was
# the blind spot: it happened to fall inside the bare 0-3 cap by
# coincidence and never exercised a genuinely CommonMark-correct deep
# continuation).
new_fixture
{
  printf -- '- [ ] ```bash verified:ran it locally\n'
  printf '      echo hi\n'
  printf '      ```\n'
  printf 'Baseline: 500 tests\n'
  printf '```bash\n'
  printf 'echo second block, untagged\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "checkbox-fence regression: guard still fails the file" \
  || fail "checkbox-fence regression: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:1"*) fail "checkbox-fence regression: tagged fence at line 1 must NOT be flagged: out=$OUT" ;;
  *) pass "checkbox-fence regression: tagged fence at line 1 is not a false positive" ;;
esac
case "$OUT" in
  *"tasks.md:4: numeric claim with no measured:/predicted: provenance comment"*) \
    pass "checkbox-fence regression: the untagged claim on line 4 is reported, not swallowed" ;;
  *) fail "checkbox-fence regression: claim on line 4 missing: out=$OUT" ;;
esac
case "$OUT" in
  *"tasks.md:5: fenced code block has no verified:/unverified: tag"*) \
    pass "checkbox-fence regression: the untagged block opened at line 5 is reported, not swallowed" ;;
  *) fail "checkbox-fence regression: untagged block at line 5 missing: out=$OUT" ;;
esac

# 27. A fence on a `- [ ] ` line with NO tag is itself flagged (the checkbox
# is not a free pass — only a stated verified:/unverified: tag is).
# Continuation/closer at the true content column (6).
new_fixture
{
  printf -- '- [ ] ```bash\n'
  printf '      echo hi\n'
  printf '      ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "an untagged checkbox-fence is detected" \
  || fail "untagged checkbox-fence: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
    pass "untagged checkbox-fence reports the opening line" ;;
  *) fail "untagged checkbox-fence line number: out=$OUT" ;;
esac

# 28. Same shape with a checked box (`- [x] `) and an unverified: tag passes.
# Continuation/closer at the true content column (6).
new_fixture
{
  printf -- '- [x] ```bash unverified:confirm the flag name\n'
  printf '      echo hi\n'
  printf '      ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a tagged checked-checkbox-fence passes" \
  || fail "checked checkbox-fence: rc=$RC out=$OUT"

# 29. A checkbox-fence nested inside a blockquote (`> - [ ] ```bash`), tagged,
# passes — blockquote, list marker, and checkbox stripped together.
# Continuation/closer repeats the blockquote marker, then reaches the true
# content column the list+checkbox established (2 columns for "- " + 4 for
# "[ ] " = 6) before the fence's own residual indentation.
new_fixture
{
  printf -- '> - [ ] ```bash verified:ran it locally\n'
  printf '>       echo hi\n'
  printf '>       ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a blockquoted checkbox-fence, tagged, passes" \
  || fail "blockquoted checkbox-fence: rc=$RC out=$OUT"

# 30. `- > ```bash` — a list item containing a blockquote, the reverse order
# from case 22/29 above. The iterative stripper retries blockquote and list
# markers on every pass regardless of which comes first, so this parses
# naturally: tagged passes.
new_fixture
{
  printf -- '- > ```bash verified:ran it locally\n'
  printf '  > echo hi\n'
  printf '  > ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a list-item-containing-a-blockquote fence, tagged, passes" \
  || fail "list-then-blockquote fence: rc=$RC out=$OUT"

# 31. A genuinely unknown prefix — a bare task checkbox with NO preceding
# list marker — must abort loudly. The checkbox grammar only recognises
# `[ ]`/`[x]`/`[X]` immediately after a list marker; here there is none.
new_fixture
{
  printf -- '[ ] ```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "a bare checkbox with no list marker aborts loudly (exit 4, content-classification)" \
  || fail "bare checkbox: expected rc=4, got rc=$RC out=$OUT"

# 32. A genuinely unrecognised prefix character (`=`) matches NONE of the
# stripper's grammar (not a blockquote marker, not a list marker, not a
# checkbox) — it is exactly as prose as a bare quote or parenthesis, so it
# is read as ordinary content, not aborted. This is a deliberate reversal
# from an earlier round of this guard (which aborted here because every
# non-letter, non-terminated-digit character was treated as a fence
# candidate — the false-abort failure mode this round exists to fix): `=
# ```bash``` is no more a fence than `"```" is the fence marker`. What
# still must abort is a character the grammar DOES assign meaning to, used
# in a shape the grammar doesn't allow — see the bare-checkbox and
# doubled-list-marker cases above, unaffected by this change.
new_fixture
printf '= ```bash``` is an assignment-like example, not a real fence.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "'= \`\`\`bash\`\`\`' is read as prose, not aborted" \
  || fail "unknown prefix character as prose: rc=$RC out=$OUT"

# 33. Prose that merely mentions a fence run mid-sentence must NOT abort and
# must NOT be treated as a fence: the line starts with an ordinary word, so
# it is never a fence-open/close candidate in the first place.
new_fixture
printf 'Note that ```bash``` marks a fence inline, not as a real block.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a mid-sentence mention of a fence run does not abort" \
  || fail "mid-sentence fence mention: rc=$RC out=$OUT"

# 34. A fence opened on a compound-numbered outline line (`3.4. ```bash`),
# tagged, passes — the compound ordered marker (digit, `.`-separated digit
# groups, terminator) must be stripped just like a simple one. Continuation/
# closer at the true content column ("3.4. " is 5 columns).
new_fixture
{
  printf '3.4. ```bash verified:ran it locally\n'
  printf '     echo hi\n'
  printf '     ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a compound-numbered (3.4.) tagged fence passes" \
  || fail "compound-numbered tagged fence: rc=$RC out=$OUT"

# 34b. The same shape, untagged, is flagged rather than silently skipped or
# aborted.
new_fixture
{
  printf '3.4. ```bash\n'
  printf '     echo hi\n'
  printf '     ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a compound-numbered (3.4.) untagged fence is detected" \
  || fail "compound-numbered untagged fence: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
    pass "compound-numbered untagged fence reports the opening line" ;;
  *) fail "compound-numbered untagged fence line number: out=$OUT" ;;
esac

# 34c. A three-level compound marker with a `)` terminator (`1.2.3)`), tagged,
# passes. Continuation/closer at the true content column ("1.2.3) " is 7
# columns).
new_fixture
{
  printf '1.2.3) ```bash verified:ran it locally\n'
  printf '       echo hi\n'
  printf '       ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a compound-numbered (1.2.3)) tagged fence passes" \
  || fail "1.2.3) tagged fence: rc=$RC out=$OUT"

# 34d. A compound number with NO terminator ("3.4 ") followed later in the
# same sentence by a fence run must be treated as prose and must NOT abort —
# it is indistinguishable from ordinary text that happens to start with a
# task number ("3.4 explains how fences work below").
new_fixture
printf '3.4 explains how the ```bash``` fence marker works inline, not as a real block.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "an unterminated compound number mentioning a fence mid-sentence does not abort" \
  || fail "unterminated compound number mid-sentence: rc=$RC out=$OUT"

# 34e. Confirm the compound-marker fixtures above did not disturb the two
# other genuinely-unrecognised-prefix cases. `= ```bash` is now prose (see
# case 32 and its updated reasoning); a bare `[ ] ```bash` (checkbox with no
# preceding list marker) is a shape the grammar DOES recognise and must
# still abort.
new_fixture
{
  printf '= ```bash\n'
  printf 'echo hi\n'
  printf '```bash verified:closes a fresh opener, not the = line\n'
  printf 'more\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "'= \`\`\`bash' is prose; the real fence below it is still parsed normally" \
  || fail "= fence as prose: rc=$RC out=$OUT"

new_fixture
{
  printf -- '[ ] ```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "a bare checkbox with no list marker still aborts loudly (exit 4, content-classification)" \
  || fail "bare checkbox still aborts: expected rc=4, got rc=$RC out=$OUT"

# 35. The four Critical-A prose fixtures: ordinary prose that mentions fence
# syntax behind punctuation must not abort, no matter which punctuation.
# Each of these was verified to abort under the pre-fix eligibility test
# (any non-letter, non-terminated-digit first character was a fence
# candidate), which is exactly the false-abort failure mode this round
# fixes.
new_fixture
printf '`git diff` shows changes; wrap snippets in ```bash fences with a tag.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a line led by a single-backtick code span does not abort" \
  || fail "backtick-led prose: rc=$RC out=$OUT"

new_fixture
printf '"```" is the fence marker used here, see docs.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a line led by a double-quote does not abort" \
  || fail "quote-led prose: rc=$RC out=$OUT"

new_fixture
printf '(wrap snippets in ```lang blocks with a tag)\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a line led by a parenthesis does not abort" \
  || fail "paren-led prose: rc=$RC out=$OUT"

new_fixture
printf '**Note:** wrap the snippet in ```bash blocks with a tag.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a line led by bold markdown (**) does not abort" \
  || fail "bold-markdown-led prose: rc=$RC out=$OUT"

# 36. Critical B: a line inside an already-open, correctly tagged, correctly
# closed fence that mentions a fence run behind punctuation as CONTENT must
# never abort the scan. Before the fix, classify_fence_line's abort path
# was consulted on every line regardless of fence state, so this exact
# shape aborted the whole guard on ordinary code-comment content.
new_fixture
{
  printf '```bash verified:test\n'
  printf 'echo "start"\n'
  printf '# See ```example``` for illustration\n'
  printf 'echo "end"\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a fence run behind punctuation as CONTENT inside an open fence does not abort" \
  || fail "content-line fence mention: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Containment / security (cont'd) — quadratic-DoS regression guard
# ===========================================================================

# 37. Critical C: a line with many repeated blockquote markers ahead of a
# fence run must not be quadratic. Before the fix, strip_container_prefix
# stripped one marker per bash while-loop iteration, re-matching against a
# shrinking string each time (O(N^2) on N markers); a single bounded-
# repetition regex pass is O(N). Timed directly rather than merely asserting
# a exit code, so a regression back to quadratic behavior fails this
# assertion instead of merely running the suite slower.
new_fixture
{
  MANY_MARKERS="$(printf '> %.0s' {1..4000})"
  printf '%s```bash\n' "$MANY_MARKERS"
  printf 'echo hi\n'
  printf '%s```\n' "$MANY_MARKERS"
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
DOS_START=$(date +%s%N)
run_guard "$FIXTURE"
DOS_END=$(date +%s%N)
DOS_MS=$(( (DOS_END - DOS_START) / 1000000 ))
[ "$RC" -ne 0 ] && pass "a 4000-marker blockquote-padded untagged fence is still detected" \
  || fail "blockquote-padded fence: expected non-zero, out=$OUT"
[ "$DOS_MS" -lt 5000 ] && pass "4000 repeated blockquote markers scan in under 5s (took ${DOS_MS}ms)" \
  || fail "blockquote-padding DoS: took ${DOS_MS}ms, expected < 5000ms"

# ===========================================================================
# SECTION: Pass-5 Criticals — added for the Python reshape (fixture
# coverage that was missing when the bash classifier was replaced)
# ===========================================================================

# 38. Pass-5 Critical (bug hunter) — container scoping on the CLOSING test.
# A fence opened with NO container prefix must be closed only by an
# unprefixed line. Before this fix, the closing test blindly stripped
# container syntax from every content line regardless of what the fence
# was opened inside, so a content line that reduces to a bare fence run
# after generic stripping (here, a blockquoted fence run appearing as
# ordinary CONTENT) falsely closed the block — fabricating a "never
# closed" violation for the real close further down, and misreading the
# next content line as prose subject to the numeric-claim check. The
# correctly-tagged, correctly-closed block below must report ZERO
# violations.
new_fixture
{
  printf '```bash verified:test\n'
  printf 'echo "start"\n'
  printf '> ```\n'
  printf 'echo "ran 200 tests"\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a blockquote-shaped fence run as CONTENT inside a bare-opened fence does not falsely close it" \
  || fail "container-scoping-on-close: rc=$RC out=$OUT"

# 39. Pass-5 Critical (adversarial) — false-abort on container-prefixed
# prose that merely mentions a fence run. The distinction is whether the
# fence run begins the line's content immediately after its container
# prefix is stripped, never whether the line starts with punctuation or a
# valid bullet/blockquote marker. Both of these are ordinary prose and
# must pass cleanly, not abort.
new_fixture
printf -- '- Confirmed the fence style uses ```: markers, not tildes.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "bullet-prefixed prose mentioning a fence run does not abort" \
  || fail "bullet-prefixed prose: rc=$RC out=$OUT"

new_fixture
printf '> See the ``` marker used below for illustration.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "blockquote-prefixed prose mentioning a fence run does not abort" \
  || fail "blockquote-prefixed prose: rc=$RC out=$OUT"

# 39b. The same shape, but WITH an accompanying untagged numeric claim on
# the same line: the claim must still be caught, not silently discarded
# by an abort that never should have fired in the first place.
new_fixture
printf -- '- 42 tests pass in the suite, see ```bash for reference\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "bullet-prefixed prose with an accompanying numeric claim is not swallowed by an abort" \
  || fail "bullet-prefixed prose + claim: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:1: numeric claim with no measured:/predicted: provenance comment"*) \
    pass "the '42 tests' claim is still reported" ;;
  *) fail "bullet-prefixed prose + claim: claim not reported: out=$OUT" ;;
esac

# 40. Pass-5 Critical (security) — a second, independent quadratic: bash
# 3.2 (what macOS ships) makes sequential indexed-array access O(N) per
# access, so reading a file into a bash array and walking it by index was
# O(N^2) overall (measured against the bash guard: 10k lines 0.68s, 160k
# lines 47.56s). Python's list indexing is O(1), so this must complete
# in a small, flat fraction of that time regardless of file size. Timed
# directly, not just asserted clean, so a regression back to quadratic
# behaviour fails this assertion rather than merely running the suite
# slower.
new_fixture
{
  for i in $(seq 1 40000); do
    printf 'Task %d: ordinary prose line with no fence and no numeric claim.\n' "$i"
  done
  printf '```bash verified:generated fixture\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
LARGE_START=$(date +%s%N)
run_guard "$FIXTURE"
LARGE_END=$(date +%s%N)
LARGE_MS=$(( (LARGE_END - LARGE_START) / 1000000 ))
[ "$RC" -eq 0 ] && pass "a 40,000-line clean tasks.md scans clean" \
  || fail "large file: rc=$RC out=$OUT"
[ "$LARGE_MS" -lt 5000 ] && pass "a 40,000-line tasks.md scans in under 5s (took ${LARGE_MS}ms)" \
  || fail "large file DoS: took ${LARGE_MS}ms, expected < 5000ms"

# ===========================================================================
# SECTION: Pass-6 fix wave — Critical 1: blockquote depth is a COUNT, not a
# boolean. A fence opened N blockquote levels deep must be closed only by a
# line at that SAME depth — not by any line that merely has "some" blockquote
# prefix. Depths 1, 2 and 3 are asserted so the fix is proven for the general
# case, not just the depth-2 fixture the review panel happened to repro.
# ===========================================================================

# 41. Depth 1: a fence opened one blockquote level deep is not closed by a
# BARE (zero-level) line; it closes correctly at a matching one-level line.
# This is the simplest non-regression case for the count-based fix.
new_fixture
{
  printf '> ```bash verified:test\n'
  printf '> echo "start"\n'
  printf '```\n'
  printf '> echo "ran 111 tests"\n'
  printf '> ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "depth-1 fence: a bare (zero-level) line does not close it; the matching one-level line does" \
  || fail "depth-1 blockquote fence: rc=$RC out=$OUT"

# 42. Depth 2: THE exact repro from the pass-6 brief. A fence opened two
# blockquote levels deep must not be closed by a one-level-deep line; the
# false close previously fabricated an unattributed-claim hit AND a
# never-closed violation. Correctly tagged and correctly (depth-2) closed:
# zero violations.
new_fixture
{
  printf '> > ```bash verified:test\n'
  printf '> > echo "start"\n'
  printf '> ```\n'
  printf '> > echo "ran 999 tests"\n'
  printf '> > ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "depth-2 fence: a one-level-deep line does not falsely close it (the pass-6 brief repro)" \
  || fail "depth-2 blockquote fence: rc=$RC out=$OUT"
[ "$OUT" = "check-plan-provenance: 1 tasks.md file(s) scanned, all provenance stated" ] \
  && pass "depth-2 fence: zero violations reported" \
  || fail "depth-2 fence: expected zero violations, out=$OUT"

# 43. Depth 3: a fence opened three blockquote levels deep is not closed by
# EITHER a one-level or a two-level line; only the matching three-level line
# closes it. Proves the fix generalises past depth 2, not just the reported
# fixture.
new_fixture
{
  printf '> > > ```bash verified:test\n'
  printf '> > > echo "start"\n'
  printf '> ```\n'
  printf '> > ```\n'
  printf '> > > echo "ran 777 tests"\n'
  printf '> > > ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "depth-3 fence: neither a one-level nor a two-level line falsely closes it; the matching three-level line does" \
  || fail "depth-3 blockquote fence: rc=$RC out=$OUT"

# 44. Negative control: the count-based fix must not make the guard blind to
# a REAL untagged fence at depth 2, correctly closed at depth 2. Depth
# tracking must not accidentally loosen the untagged-fence check itself.
new_fixture
{
  printf '> > ```bash\n'
  printf '> > echo hi\n'
  printf '> > ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "depth-2 fence: a genuinely untagged block, correctly closed at depth 2, is still flagged" \
  || fail "depth-2 untagged fence: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
    pass "depth-2 untagged fence reports the opening line" ;;
  *) fail "depth-2 untagged fence line number: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-6 fix wave — Critical 2: a leading UTF-8 BOM defeats the
# guard
# ===========================================================================

# 45. A BOM-prefixed fence opener must not fall through to prose: before the
# fix, `_WS_RE` does not match U+FEFF, so the BOM-prefixed opener line was
# misread as prose, the fence never truly opened, and the guard's own
# (bare) closing "```" line was read as a FRESH untagged opener — silently
# swallowing everything after it, including a genuine unattributed claim.
# After the fix (`encoding="utf-8-sig"` strips the BOM at read time), the
# block opens and closes correctly and the trailing claim IS reported.
new_fixture
{
  printf '\xef\xbb\xbf```bash verified:test\n'
  printf 'echo hi\n'
  printf '```\n'
  printf 'Baseline: 999 tests\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a BOM-prefixed fence opens/closes correctly; the trailing unattributed claim is reported" \
  || fail "BOM-prefixed fixture: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:4: numeric claim with no measured:/predicted: provenance comment"*) \
    pass "BOM fixture: the '999 tests' claim is reported, not silently swallowed" ;;
  *) fail "BOM fixture: claim not reported: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-6 fix wave — Critical 3: quadratic ReDoS in CLAIM_RE
# ===========================================================================

# 46. A large single line with a long comma-separated digit run and NO
# trailing unit word must not cause CLAIM_RE to backtrack quadratically.
# Before the fix ([0-9][0-9,]*[\s]+), measured through the real wrapper:
# 4,000 -> 0.22s, 8,000 -> 0.66s, 16,000 -> 2.39s (see the fix-wave report
# for the exact numbers re-measured on this tree). Timed directly, not just
# asserted clean, so a regression back to quadratic behaviour fails this
# assertion instead of merely running the suite slower.
new_fixture
python3 -c "
n = 16000
print(','.join(str(i) for i in range(1, n)) + ' widgets')
" > "$FIXTURE/openspec/changes/demo-change/tasks.md"
REDOS_START=$(date +%s%N)
run_guard "$FIXTURE"
REDOS_END=$(date +%s%N)
REDOS_MS=$(( (REDOS_END - REDOS_START) / 1000000 ))
[ "$RC" -eq 0 ] && pass "a large comma-separated-digits line with no unit word is not a claim (exit 0)" \
  || fail "ReDoS fixture correctness: rc=$RC out=$OUT"
[ "$REDOS_MS" -lt 2000 ] && pass "a 16,000-number comma-separated line scans well under a second (took ${REDOS_MS}ms)" \
  || fail "ReDoS regression: took ${REDOS_MS}ms, expected < 2000ms"

# ===========================================================================
# SECTION: Pass-6 fix wave — Critical 4: unhandled OSError leaks a traceback
# and exits 1
# ===========================================================================

# 47. `chmod 000` on a tasks.md that has already passed containment must
# produce a classified message and an exit code that is NOT 1 (1 is this
# guard's own documented code for "violations found" — indistinguishable
# from a crash if a PermissionError were allowed to propagate as a
# traceback).
new_fixture
printf 'Baseline: 197 tests\n<!-- measured: ./gradlew test @ c515c42 -->\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
chmod 000 "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
chmod 644 "$FIXTURE/openspec/changes/demo-change/tasks.md"
[ "$RC" -ne 1 ] && pass "chmod 000 on tasks.md exits classified (rc=$RC), never rc=1 (violations-found)" \
  || fail "chmod 000: expected rc != 1, got rc=$RC out=$OUT"
case "$OUT" in
  *"Traceback"*) fail "chmod 000: a Python traceback leaked: out=$OUT" ;;
  *"cannot read file"*) pass "chmod 000: a classified message is printed, not a traceback" ;;
  *) fail "chmod 000: no classified message found: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-6 fix wave — Important 5: missing python3 exits 127, outside
# the taxonomy
# ===========================================================================

# 48. With python3 absent from PATH, the wrapper itself must exit 2 with a
# clear message rather than letting `exec python3 ...` fail with a bare
# "command not found" (bash's own exit 127, outside this guard's taxonomy).
new_fixture
FAKEPATH="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-fakepath.XXXXXX")"
# A minimal PATH that has no python3 anywhere on it, but still has what the
# wrapper itself needs to run at all (bash, for the `#!/usr/bin/env bash`
# shebang, and dirname, for SCRIPT_DIR resolution) — `env -i` plus an
# explicit PATH is what makes this deterministic regardless of the caller's
# real environment.
ln -sf "$(command -v bash)" "$FAKEPATH/bash"
ln -sf "$(command -v dirname)" "$FAKEPATH/dirname"
set +e
NO_PY3_OUT="$(env -i PATH="$FAKEPATH" CHECK_PLAN_PROVENANCE_ROOT="$FIXTURE" "$GUARD" 2>&1)"
NO_PY3_RC=$?
set -e
rm -rf "$FAKEPATH"
[ "$NO_PY3_RC" -eq 2 ] && pass "missing python3 on PATH exits 2" \
  || fail "missing python3: expected rc=2, got rc=$NO_PY3_RC out=$NO_PY3_OUT"
case "$NO_PY3_OUT" in
  *"python3"*) pass "missing python3: message names python3" ;;
  *) fail "missing python3: message does not mention python3: out=$NO_PY3_OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-6 fix wave — Important 6: invalid UTF-8 is silently
# errors="replace"d
# ===========================================================================

# 49. A file the guard cannot faithfully decode as UTF-8 must be refused
# with the same classified exit code chosen for Critical 4 (chmod 000),
# never silently scanned with replacement characters that could mask a
# claim.
new_fixture
printf 'Baseline: 197 tests\n<!-- measured: ./gradlew test @ c515c42 -->\n\xff\xfe not valid utf-8\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 2 ] && pass "invalid UTF-8 exits 2 (same taxonomy code as the chmod-000 OSError case)" \
  || fail "invalid UTF-8: expected rc=2, got rc=$RC out=$OUT"
case "$OUT" in
  *"cannot decode as UTF-8"*) pass "invalid UTF-8: a classified message is printed" ;;
  *) fail "invalid UTF-8: no classified message found: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-6 fix wave — Minor: a numeric claim on a fence's info-string
# line was never checked
# ===========================================================================

# 50. A numeric claim embedded in a fence's own OPENING line (its info
# string, after the verified:/unverified: tag) must be checked like any
# other numeric claim — it was previously invisible because the info string
# is only ever tested against the tag pattern.
new_fixture
printf -- '```bash verified:ran 500 tests\necho hi\n```\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "a numeric claim in a fence's info string is reported" \
  || fail "info-string claim: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:1: numeric claim with no measured:/predicted: provenance comment"*) \
    pass "info-string claim reports the fence's opening line" ;;
  *) fail "info-string claim line number: out=$OUT" ;;
esac

# 50b. The same shape, but with a measured: comment within the lookahead
# window, passes — the info-string claim is subject to the same lookahead
# rule as any other claim, not a stricter one.
new_fixture
{
  printf -- '```bash verified:ran 500 tests\n'
  printf '<!-- measured: ./gradlew test @ c515c42 -->\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "an info-string claim with a measured: comment in the lookahead window passes" \
  || fail "info-string claim with provenance: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-7 fix wave — Critical 1: an unreadable/broken scan entry is
# a scan-integrity failure, never a silent skip. Four shapes, one root
# cause (os.path.isdir/os.path.lexists swallow every OSError and report
# False, indistinguishable from "absent").
# ===========================================================================

# 51. A locked (chmod 000) change directory beside a clean one must not be
# silently dropped: the unattributed claim inside it must never be read as
# "clean" simply because this guard could not look it up.
new_fixture
mkdir -p "$FIXTURE/openspec/changes/locked-change"
printf 'Baseline: 123 tests\n' > "$FIXTURE/openspec/changes/locked-change/tasks.md"
chmod 000 "$FIXTURE/openspec/changes/locked-change"
clean_tasks_md
run_guard "$FIXTURE"
chmod 755 "$FIXTURE/openspec/changes/locked-change"
[ "$RC" -eq 2 ] && pass "a locked (chmod 000) change dir beside a clean one aborts classified, never a clean run" \
  || fail "locked dir: expected rc=2, got rc=$RC out=$OUT"
case "$OUT" in
  *"all provenance stated"*) fail "locked dir: false clean run leaked: out=$OUT" ;;
  *) pass "locked dir: no false 'all provenance stated' clean run" ;;
esac

# 52. A dangling symlink AS the only change directory must not report
# "nothing in flight" — that message is reserved for genuinely ZERO
# non-archived entries, not for an entry this guard cannot classify.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
ln -s "$FIXTURE/openspec/changes/does-not-exist" "$FIXTURE/openspec/changes/dangling-change"
run_guard "$FIXTURE"
[ "$RC" -eq 2 ] && pass "a dangling symlink change dir aborts classified, never 'nothing in flight'" \
  || fail "dangling symlink (alone): expected rc=2, got rc=$RC out=$OUT"
case "$OUT" in
  *"nothing in flight"*) fail "dangling symlink (alone): false 'nothing in flight' leaked: out=$OUT" ;;
  *) pass "dangling symlink (alone): no false 'nothing in flight'" ;;
esac

# 53. A dangling symlink beside a REAL, clean change: the whole scan must
# abort classified rather than silently reporting the real change as the
# entire (clean) picture.
new_fixture
clean_tasks_md
ln -s "$FIXTURE/openspec/changes/does-not-exist" "$FIXTURE/openspec/changes/dangling-change"
run_guard "$FIXTURE"
[ "$RC" -eq 2 ] && pass "a dangling symlink beside a real change aborts classified, never a clean run" \
  || fail "dangling symlink (beside real): expected rc=2, got rc=$RC out=$OUT"
case "$OUT" in
  *"all provenance stated"*) fail "dangling symlink (beside real): false clean run leaked: out=$OUT" ;;
  *) pass "dangling symlink (beside real): no false 'all provenance stated' clean run" ;;
esac

# 54. A circular symlink (ELOOP) change directory must abort classified,
# never silently resolve to "not a directory" and be skipped.
new_fixture
clean_tasks_md
ln -s "$FIXTURE/openspec/changes/circular-a" "$FIXTURE/openspec/changes/circular-b"
ln -s "$FIXTURE/openspec/changes/circular-b" "$FIXTURE/openspec/changes/circular-a"
run_guard "$FIXTURE"
[ "$RC" -eq 2 ] && pass "a circular symlink change dir aborts classified, never silently skipped" \
  || fail "circular symlink: expected rc=2, got rc=$RC out=$OUT"
case "$OUT" in
  *"all provenance stated"*) fail "circular symlink: false clean run leaked: out=$OUT" ;;
  *) pass "circular symlink: no false 'all provenance stated' clean run" ;;
esac

# ===========================================================================
# SECTION: Pass-7 fix wave — Critical 2: an UN-CONTAINED fence's own
# indentation is capped at 3 columns (CommonMark); 4+ columns with no
# container prefix is an indented code block, and a run of backticks/
# tildes there is literal text, never a fence. Depths 0, 3, 4 and 8,
# backtick and tilde, tagged and untagged.
# ===========================================================================

# 55. 0 columns: an ordinary, correctly tagged fence still opens/closes
# normally (non-regression baseline for the cap).
new_fixture
printf -- '```bash verified:ci\necho hi\n```\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "0-column fence still opens/closes normally" \
  || fail "0-column fence: rc=$RC out=$OUT"

# 56. 3 columns: still within CommonMark's un-contained budget — a real
# fence, tagged, must still be recognised as one (no false claim report).
new_fixture
printf -- '   ```bash verified:ci\n   echo hi\n   ```\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "3-column fence is still a real fence" \
  || fail "3-column fence: rc=$RC out=$OUT"

# 57. 4 columns, backtick, tagged info string: THE exact repro from the
# pass-7 brief, RE-POINTED by the pass-8 fix wave's Critical 1 (an operator
# decision, not a rediscovered bug): pass 7 read a BARE 4+-column fence-like
# line as "not a fence, an indented code block" and kept scanning its
# content as prose. Pass 8 deletes the ambient list-column tracker that
# reasoning depended on and refuses to guess instead — a bare line reducing
# to a fence run at 4+ columns now aborts (exit 4) rather than being read
# either way. Once that line aborts, this guard stops reading the rest of
# THIS file (the same "cannot continue" shape an unreadable/undecodable
# file already has — see Critical 2 below): the claim later in this same
# fixture is no longer reached, which is the point, not a regression — see
# the standalone 4-space-indented-PROSE case in the Critical-1 section
# below for the proof that ordinary prose (no fence run on it) is still
# scanned exactly as before.
new_fixture
{
  printf '    ``` verified:ci\n'
  printf '    echo hi\n'
  printf '    Suite finished: ran 500 tests.\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "a 4-column bare fence-like run aborts (content-classification), not read as an indented code block" \
  || fail "4-column pseudo-fence (backtick): expected rc=4, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:1: fence-like run indented 4+ columns"*) \
    pass "4-column pseudo-fence: the abort names the opening line" ;;
  *) fail "4-column pseudo-fence: abort message missing/wrong line: out=$OUT" ;;
esac

# 58. 8 columns, tilde, tagged info string: proves the fix generalises past
# the exact 4-column/backtick fixture the brief happened to quote — deeper
# indentation and the other fence character abort identically.
new_fixture
{
  printf '        ~~~ verified:ci\n'
  printf '        echo hi\n'
  printf '        Suite finished: ran 777 tests.\n'
  printf '        ~~~\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "an 8-column bare fence-like run (tilde) aborts identically to the 4-column backtick case" \
  || fail "8-column pseudo-fence (tilde): expected rc=4, got rc=$RC out=$OUT"

# 59. 4 columns, backtick, UNTAGGED: the pseudo-fence itself must not be
# reported as an untagged FENCE (it never opened one, and the ambiguity
# means this guard no longer claims to know it wasn't one either) — it
# aborts, exactly like the tagged case in 57 above; the tag/no-tag
# distinction is irrelevant to an opener this parser refuses to classify at
# all.
new_fixture
{
  printf '    ```\n'
  printf '    Baseline: ran 321 tests.\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "a 4-column untagged bare fence-like run aborts rather than being misread either as a fence or as prose" \
  || fail "4-column untagged pseudo-fence: expected rc=4, got rc=$RC out=$OUT"
case "$OUT" in
  *"fenced code block has no verified:/unverified: tag"*) \
    fail "4-column untagged pseudo-fence: falsely reported as an untagged FENCE violation: out=$OUT" ;;
  *) pass "4-column untagged pseudo-fence: not misreported as a fence violation" ;;
esac

# ===========================================================================
# SECTION: Pass-7 fix wave — Critical 3: CLAIM_RE's boundary is negative
# (not word-constituent), not an ASCII-punctuation allowlist — em dash,
# curly quote and fullwidth paren, on both the left and right side of a
# claim.
# ===========================================================================

# 60. Left boundary, em dash (U+2014), glued directly to the digit.
new_fixture
printf 'Result—42 tests\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "em dash (left, glued) is caught" || fail "em dash left: rc=$RC out=$OUT"

# 61. Left boundary, curly opening quote (U+2018), glued directly to the
# digit.
new_fixture
printf 'Result‘42 tests\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "curly quote (left, glued) is caught" || fail "curly quote left: rc=$RC out=$OUT"

# 62. Left boundary, fullwidth left parenthesis (U+FF08), glued directly to
# the digit.
new_fixture
printf '（42 tests\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "fullwidth paren (left, glued) is caught" || fail "fullwidth paren left: rc=$RC out=$OUT"

# 63. Right boundary, em dash, glued directly after the unit word.
new_fixture
printf 'Result: 42 tests—confirmed\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "em dash (right, glued) is caught" || fail "em dash right: rc=$RC out=$OUT"

# 64. Right boundary, curly closing quote (U+2019), glued directly after
# the unit word.
new_fixture
printf 'Result: “42 tests’\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "curly quote (right, glued) is caught" || fail "curly quote right: rc=$RC out=$OUT"

# 65. Right boundary, fullwidth right parenthesis (U+FF09), glued directly
# after the unit word.
new_fixture
printf 'Result: (42 tests）\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "fullwidth paren (right, glued) is caught" || fail "fullwidth paren right: rc=$RC out=$OUT"

# 66. Non-regression: the already-working spaced em dash and ASCII hyphen
# shapes from the brief's own table must still be caught.
new_fixture
printf 'Result — 42 tests\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "em dash (left, spaced) is still caught" || fail "em dash spaced: rc=$RC out=$OUT"

new_fixture
printf 'Result-42 tests\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "ASCII hyphen (left, glued) is still caught" || fail "ASCII hyphen: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-7 fix wave — Important 7: the 10 MiB size cap had zero test
# coverage.
# ===========================================================================

# 67. A tasks.md larger than MAX_TASKS_MD_BYTES must be refused (exit 2),
# naming the file, rather than read unbounded.
new_fixture
python3 -c "
import sys
path = sys.argv[1]
with open(path, 'wb') as f:
    f.write(b'x' * (10 * 1024 * 1024 + 1024))
" "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 2 ] && pass "a tasks.md over the 10 MiB cap is refused (exit 2)" \
  || fail "size cap: expected rc=2, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md"*"exceeds the"*"byte cap"*) pass "size cap: the message names the file and the cap" ;;
  *) fail "size cap: no classified message found: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-7 fix wave — Important 8: a per-file classification failure
# (bad encoding, unreadable, oversized) must not mask a real violation found
# in ANOTHER file, and must not stop the scan of other change directories.
# ===========================================================================

# 68. One change with a bad-encoding tasks.md, and a SEPARATE change with a
# genuine unattributed claim: the violation in the good file must still be
# reported (exit 1), not masked by the bad file's environment failure.
new_fixture
mkdir -p "$FIXTURE/openspec/changes/bad-encoding-change"
printf '\xff\xfe not valid utf-8\n' > "$FIXTURE/openspec/changes/bad-encoding-change/tasks.md"
printf 'Baseline: 197 tests\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "a bad-encoding file elsewhere does not mask a real violation (exit 1, not 2)" \
  || fail "mixed bad-encoding + violation: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"demo-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment"*) \
    pass "mixed bad-encoding + violation: the real violation is still reported" ;;
  *) fail "mixed bad-encoding + violation: real violation not reported: out=$OUT" ;;
esac
case "$OUT" in
  *"cannot decode as UTF-8"*) pass "mixed bad-encoding + violation: the bad file is still named" ;;
  *) fail "mixed bad-encoding + violation: bad file not named: out=$OUT" ;;
esac

# 69. Two changes, BOTH bad-encoding, no real violations anywhere: must
# still refuse a clean run (exit 2), never exit 0.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/bad-encoding-change"
printf '\xff\xfe not valid utf-8\n' > "$FIXTURE/openspec/changes/bad-encoding-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 2 ] && pass "a bad-encoding file with no other violations still refuses a clean run (exit 2)" \
  || fail "bad-encoding only: expected rc=2, got rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-7 fix wave — Critical 2 REGRESSION FIX: the indentation cap
# must be measured against the fence's CONTAINER content column, not the
# absolute start of the line — a coordinator review caught this after the
# first Critical-2 fix landed (a legitimate, correctly tagged fence inside a
# blockquote or list was being misread as an indented code block, its
# content scanned as prose, and claims inside it falsely reported). Fixed by
# measuring residual indentation AFTER container-prefix stripping, plus an
# ambient top-level list-item content-column tracker for list continuation
# lines, which never repeat the marker.
# ===========================================================================

# 70. Blockquote, relative indent 0: fence marker glued directly to '>',
# no intervening space at all.
new_fixture
{
  printf '>```bash verified:t\n'
  printf '>ran 500 tests\n'
  printf '>```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "blockquote rel indent 0: real fence recognized (no false claim report)" \
  || fail "blockquote rel indent 0: rc=$RC out=$OUT"

# 71. Blockquote, relative indent 2: THE exact repro from the coordinator's
# regression report. '>' consumes one optional trailing space, leaving 2
# columns before the fence — still within the 0-3 budget.
new_fixture
{
  printf '>   ```bash verified:t\n'
  printf '>   ran 500 tests\n'
  printf '>   ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "blockquote rel indent 2 (coordinator's exact repro): real fence recognized" \
  || fail "blockquote rel indent 2: rc=$RC out=$OUT"
[ "$OUT" = "check-plan-provenance: 1 tasks.md file(s) scanned, all provenance stated" ] \
  && pass "blockquote rel indent 2: zero violations reported" \
  || fail "blockquote rel indent 2: expected zero violations, out=$OUT"

# 72. Blockquote, relative indent 3: still within budget (the boundary
# case).
new_fixture
{
  printf '>    ```bash verified:t\n'
  printf '>    ran 500 tests\n'
  printf '>    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "blockquote rel indent 3: real fence recognized" \
  || fail "blockquote rel indent 3: rc=$RC out=$OUT"

# 73. Blockquote, relative indent 4: OVER budget — an indented code block
# relative to the blockquote's own content column, so the claim inside is
# reported, not swallowed.
new_fixture
{
  printf '>     ```bash verified:t\n'
  printf '>     ran 500 tests\n'
  printf '>     ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "blockquote rel indent 4: over budget, claim reported (not a fence)" \
  || fail "blockquote rel indent 4: expected non-zero, out=$OUT"
case "$OUT" in
  *"numeric claim with no measured:/predicted: provenance comment"*) \
    pass "blockquote rel indent 4: the '500 tests' claim is reported" ;;
  *) fail "blockquote rel indent 4: claim not reported: out=$OUT" ;;
esac

# 74. List item, relative indent 0: fence at absolute column 2, across a
# blank line. A BARE line (no container syntax of its own) at 0-3 columns
# is unconditionally a fence under the pass-8 rule, with no need to
# consult any list context — deleting the ambient tracker does not touch
# this case at all.
new_fixture
{
  printf -- '- item\n'
  printf '\n'
  printf '  ```bash verified:t\n'
  printf '  ran 500 tests\n'
  printf '  ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "list rel indent 0 (absolute column 2): bare fence within the 0-3 budget, unconditionally a fence" \
  || fail "list rel indent 0: rc=$RC out=$OUT"

# 75. Pass-8 fix wave, Critical 1 — REVERSED from pass 7's behaviour on
# purpose. Absolute column 5 (list content column 2, "relative" 3) is a
# BARE line at 4+ columns: pass 7's ambient list-content tracker used to
# read this as a legitimate, list-relative fence and report zero
# violations. Pass 8 deletes that tracker — three separate silent misses
# traced back to it in one review pass — and refuses to guess instead:
# whether a list is in scope here is unknowable from this line alone, so
# it now aborts (exit 4), regardless of the fact that a list item happens
# to precede it. This is the operator's fail-loud choice, not a
# regression.
new_fixture
{
  printf -- '- item\n'
  printf '\n'
  printf '     ```bash verified:t\n'
  printf '     echo hi\n'
  printf '     ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "list rel indent 3 (absolute column 5): now aborts (content-classification) instead of silently resolving via list context" \
  || fail "list rel indent 3: expected rc=4, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:3: fence-like run indented 4+ columns"*) \
    pass "list rel indent 3: the abort names the opening line" ;;
  *) fail "list rel indent 3: abort message missing/wrong line: out=$OUT" ;;
esac

# 76. List item, absolute indent 6 (relative 4, over budget even under the
# deleted tracker's own math): a BARE fence-like run at 4+ columns aborts,
# exactly like case 75 — indentation this deep was never going to resolve
# to a fence either way, only now the guard says so explicitly (exit 4)
# instead of quietly reading it as an indented code block and scanning
# past it for a claim.
new_fixture
{
  printf -- '- item\n'
  printf '\n'
  printf '      ```bash verified:t\n'
  printf '      ran 500 tests\n'
  printf '      ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "list rel indent 4 (absolute column 6): aborts (content-classification)" \
  || fail "list rel indent 4: expected rc=4, got rc=$RC out=$OUT"

# 77. Non-regression: un-contained 0 and 3 columns are still real fences;
# 4 and 8 now abort (content-classification) rather than being read as an
# indented code block and scanned past — same reversal as cases 75/76,
# generalised to the case with no enclosing list or blockquote at all.
new_fixture
printf -- '```bash verified:t\nran 500 tests\n```\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "un-contained rel/abs indent 0: still a real fence" \
  || fail "un-contained 0 (regression): rc=$RC out=$OUT"

new_fixture
printf -- '   ```bash verified:t\n   ran 500 tests\n   ```\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "un-contained rel/abs indent 3: still a real fence" \
  || fail "un-contained 3 (regression): rc=$RC out=$OUT"

new_fixture
printf -- '    ```bash verified:t\n    ran 500 tests\n    ```\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "un-contained rel/abs indent 4: aborts (content-classification), no active list/blockquote to be relative to" \
  || fail "un-contained 4 (pass-8 reversal): expected rc=4, got rc=$RC out=$OUT"

new_fixture
printf -- '        ```bash verified:t\n        ran 500 tests\n        ```\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "un-contained rel/abs indent 8: aborts (content-classification)" \
  || fail "un-contained 8 (pass-8 reversal): expected rc=4, got rc=$RC out=$OUT"

# 78. A TAB-indented un-contained fence: a tab expands to 4 columns from
# column 0 (CommonMark's tab-expansion rule), over the 0-3 budget — a BARE
# line at 4+ columns, so this now aborts (content-classification), same
# reversal as cases 75-77.
new_fixture
printf -- '\t```bash verified:t\n\tran 500 tests\n\t```\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "a tab-indented un-contained fence (tab = 4 columns): aborts (content-classification)" \
  || fail "tab-indented fence: expected rc=4, got rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-12 fix wave — Critical 1: a tab AFTER a container prefix
# token (bullet/ordered marker, checkbox, blockquote marker) must expand
# relative to its TRUE absolute column on the line, not relative to the
# start of whatever mid-line slice `strip_container_prefix`/`classify_line`
# happen to be looking at. The one pre-existing tab fixture above is a BARE
# line with the tab at true column 0 — the one case that was never broken.
# These four reproduce the shapes named directly in the fix-wave brief: a
# tab after a bullet marker, after an ordered marker, after a checkbox, and
# after a blockquote marker immediately preceding a SEPARATE untagged fence
# (the silent-miss shape — before this fix, only the numeric claim was
# reported and the untagged fence was never flagged at all).
#
# Content columns below are hand-computed by CommonMark's tab-expansion
# rule (a tab advances to the next multiple-of-4 column, from the tab's
# TRUE position on the line) and cross-checked against
# check-plan-provenance.py's `_advance_col`, not against this guard's own
# prior (buggy) output.
# ===========================================================================

# 78a. Tab after a bullet marker: ' -<TAB>' — leading space (col 0->1), '-'
# (col 1->2), tab at col 2 expands to the next multiple of 4 = col 4. True
# content column is 4.
new_fixture
{
  printf ' -\t```bash verified:t\n'
  printf '    ran 1 tests\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "tab after bullet marker, tagged, true content col 4: exit 0" \
  || fail "tab after bullet marker, tagged: rc=$RC out=$OUT"

new_fixture
{
  printf ' -\t```bash\n'
  printf '    ran 1 tests\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "tab after bullet marker, untagged, true content col 4: exit 1" \
  || fail "tab after bullet marker, untagged: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
    pass "tab after bullet marker, untagged: opening line reported, not 'never closed'" ;;
  *) fail "tab after bullet marker, untagged: wrong/missing message: out=$OUT" ;;
esac

# 78b. Tab after an ordered marker: ' 1.<TAB>' — leading space (col 0->1),
# '1.' (col 1->3), tab at col 3 expands to the next multiple of 4 = col 4.
# True content column is 4 — same as 78a, different marker shape.
new_fixture
{
  printf ' 1.\t```bash verified:t\n'
  printf '    ran 1 tests\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "tab after ordered marker, tagged, true content col 4: exit 0" \
  || fail "tab after ordered marker, tagged: rc=$RC out=$OUT"

new_fixture
{
  printf ' 1.\t```bash\n'
  printf '    ran 1 tests\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "tab after ordered marker, untagged, true content col 4: exit 1" \
  || fail "tab after ordered marker, untagged: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
    pass "tab after ordered marker, untagged: opening line reported, not 'never closed'" ;;
  *) fail "tab after ordered marker, untagged: wrong/missing message: out=$OUT" ;;
esac

# 78c. Tab after a checkbox: '- [ ]<TAB>' — '- ' (col 0->2), '[ ]' (col
# 2->5), tab at col 5 expands to the next multiple of 4 = col 8. True
# content column is 8.
new_fixture
{
  printf -- '- [ ]\t```bash verified:t\n'
  printf '        ran 1 tests\n'
  printf '        ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "tab after checkbox, tagged, true content col 8: exit 0" \
  || fail "tab after checkbox, tagged: rc=$RC out=$OUT"

new_fixture
{
  printf -- '- [ ]\t```bash\n'
  printf '        ran 1 tests\n'
  printf '        ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "tab after checkbox, untagged, true content col 8: exit 1" \
  || fail "tab after checkbox, untagged: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
    pass "tab after checkbox, untagged: opening line reported, not 'never closed'" ;;
  *) fail "tab after checkbox, untagged: wrong/missing message: out=$OUT" ;;
esac

# 78d. Tab after a blockquote marker, immediately followed by a SEPARATE,
# unrelated, untagged fence later in the file — the silent-miss shape from
# the fix-wave brief: before this fix, the blockquote+tab line's true
# content column was over/under-counted, which could misread the file's
# real structure and swallow the later untagged fence entirely, reporting
# only a numeric claim (or nothing) and never flagging the fence itself.
# '>' plus one space (col 0->2 — the one column a blockquote marker is
# credited with, `_consume_bq_run`'s job since pass 13 and `_BQ_RUN_RE`'s
# `\s?` before it), then a tab at col 2 expanding to the next multiple of
# 4 = col 4. This line is blockquote-prefixed PROSE — its text is not a
# fence run at all — never a fence opener. Note where the tab sits: AFTER
# the marker's one space was already consumed, which is the one placement
# the pass-13 Critical could not reach, and why that defect survived here. The real assertion is that the untagged
# fence two paragraphs later is still independently caught.
new_fixture
{
  printf '> \tordinary blockquoted text, not a fence\n'
  printf '\n'
  printf '```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "tab after blockquote marker + later untagged fence: exit 1, not silently swallowed" \
  || fail "tab after blockquote marker + later untagged fence: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:3: fenced code block has no verified:/unverified: tag"*) \
    pass "tab after blockquote marker + later untagged fence: the fence itself is flagged, not silently skipped" ;;
  *) fail "tab after blockquote marker + later untagged fence: fence not reported: out=$OUT" ;;
esac

# 79. An UNTAGGED fence inside a blockquote at relative indent 0 must still
# be flagged as an untagged FENCE — the container-relative cap must not
# accidentally weaken the untagged-fence check itself.
new_fixture
{
  printf '>```bash\n'
  printf '>ran 500 tests\n'
  printf '>```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
case "$OUT" in
  *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
    pass "untagged fence inside a blockquote (rel indent 0) is still flagged" ;;
  *) fail "untagged fence in blockquote: out=$OUT" ;;
esac

# 80. Nested: a blockquote INSIDE a list item, fence at relative indent 0.
# The blockquote marker repeats on every line (its own container syntax),
# so this is resolved directly by per-line stripping — unaffected by
# Critical 1's deletion of the ambient list-content-column tracker, which
# only ever applied to BARE lines with no container syntax of their own.
new_fixture
{
  printf -- '- item\n'
  printf '  >```bash verified:t\n'
  printf '  >ran 500 tests\n'
  printf '  >```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "nested blockquote-inside-list, rel indent 0: real fence recognized" \
  || fail "nested blockquote-in-list: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-8 fix wave — Critical 1: the ambient list-column tracker is
# DELETED; ambiguous list context fails loud (exit 4) instead of guessing.
# This is the operator's fail-loud decision, not a rediscovered bug — see
# check-plan-provenance.py's classify_line docstring. Required harness list
# from the fix-wave brief, in order.
# ===========================================================================

# 81. `- item` then a 2-space-indented tagged fence, with NO blank line in
# between — unambiguous (bare, 0-3 columns), must still work exactly as it
# did before this fix wave.
new_fixture
{
  printf -- '- item\n'
  printf '  ```bash verified:t\n'
  printf '  echo hi\n'
  printf '  ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "'- item' then a 2-space-indented tagged fence: still exit 0 (unambiguous)" \
  || fail "list then 2-space fence: rc=$RC out=$OUT"

# 82. A 3-space-indented BARE tagged fence, with no enclosing list or
# blockquote at all — still within the 0-3 budget, still exit 0.
new_fixture
{
  printf '   ```bash verified:t\n'
  printf '   echo hi\n'
  printf '   ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "a 3-space-indented bare tagged fence: exit 0" \
  || fail "3-space bare fence: rc=$RC out=$OUT"

# 83. A 4-space-indented BARE fence run aborts (exit 4), and the message
# names the exact line.
new_fixture
{
  printf '    ```bash verified:t\n'
  printf '    echo hi\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "a 4-space-indented bare fence run aborts (exit 4)" \
  || fail "4-space bare fence: expected rc=4, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:1: fence-like run indented 4+ columns"*) \
    pass "4-space bare fence abort: message names the opening line" ;;
  *) fail "4-space bare fence abort: message missing/wrong line: out=$OUT" ;;
esac

# 84. `- - -` (a thematic break, NOT a list item) followed by a 4-space
# fence: THE root cause of one of the pass-7 Criticals — `_LIST_RE` matched
# the first `- ` of a thematic break and opened a phantom list scope, which
# then made the following 4-space fence look "in scope" and swallowed its
# claim/tag silently (exit 0, "all provenance stated" — a silent miss). The
# tracker that produced that reading is deleted; this now aborts (exit 4)
# instead, exactly as any other bare 4+-column fence-like run does.
new_fixture
{
  printf -- '- - -\n'
  printf '    ```bash\n'
  printf '    echo hi\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "'- - -' (thematic break) then a 4-space fence: no longer a silent miss — aborts (exit 4)" \
  || fail "thematic break '- - -': expected rc=4, got rc=$RC out=$OUT"

# 84b. `* * *`, the other common thematic-break spelling — same principle.
new_fixture
{
  printf -- '* * *\n'
  printf '    ```bash\n'
  printf '    echo hi\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "'* * *' (thematic break) then a 4-space fence: aborts (exit 4)" \
  || fail "thematic break '* * *': expected rc=4, got rc=$RC out=$OUT"

# 85. `-      item` (a real list marker, but with a WIDE gap of trailing
# whitespace before the item text) followed on a LATER line by a 4-space
# fence: the second pass-7 root cause — the marker's trailing `\s+` is
# greedy and uncapped, so the ambient tracker inflated the tracked content
# column from the wide gap, making the following 4-space fence look
# "in scope" (silent miss, exit 0). Deleting the tracker means the later
# bare 4-space fence is judged purely on its own absolute indentation, with
# no memory of the wide gap at all — it now aborts (exit 4).
new_fixture
{
  printf -- '-      item\n'
  printf '    ```bash\n'
  printf '    echo hi\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "'-      item' (wide gap) then a 4-space fence: no longer a silent miss — aborts (exit 4)" \
  || fail "wide-gap list item: expected rc=4, got rc=$RC out=$OUT"

# 86. Nested list, dedent, fence at 2-space indent: an outer item, a nested
# item, then a dedent back to a BARE fence at absolute column 2 — still
# within the unconditional 0-3 budget, so still exit 0 regardless of the
# nesting or the dedent. This is the required non-regression proof that
# ordinary, shallow list-fence shapes this repository's own plans actually
# use are not swept up by Critical 1's fail-loud rule — only 4+-column
# BARE lines are.
new_fixture
{
  printf -- '- outer\n'
  printf '  - inner\n'
  printf '  ```bash verified:t\n'
  printf '  echo hi\n'
  printf '  ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "nested list, dedent, fence at 2-space indent: exit 0" \
  || fail "nested list dedent: rc=$RC out=$OUT"

# 87. THE case that proves the abort is scoped to fence RUNS, not to
# indentation in general: a 4-space-indented line of ORDINARY prose (no
# backtick/tilde run on it at all) carrying an untagged numeric claim must
# be scanned and reported (exit 1) exactly as before — never an abort.
# Getting this backwards would refuse most real plans, which routinely use
# 4-space (or list-relative) indentation for ordinary continuation prose.
new_fixture
printf '    Baseline: ran 500 tests with no fence in sight.\n' \
  > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "4-space-indented ordinary prose with an untagged claim: exit 1, not an abort" \
  || fail "4-space prose claim: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:1: numeric claim with no measured:/predicted: provenance comment"*) \
    pass "4-space prose claim: the claim is reported" ;;
  *) fail "4-space prose claim: claim not reported: out=$OUT" ;;
esac
case "$OUT" in
  *"fence-like run"*) fail "4-space prose claim: falsely treated as an abort candidate: out=$OUT" ;;
  *) pass "4-space prose claim: no false abort" ;;
esac

# 88. Blockquote- and list-prefixed fences at relative 0-3 exit 0; at
# relative 4 they are still reported as a claim, exactly as before Critical
# 1 (this behaviour lives entirely in the has-container-prefix branch,
# untouched by deleting the ambient tracker). Blockquote relative 0-4 is
# already pinned by cases 70-73 above; this pins the LIST-prefixed side: a
# list marker's own `\s+` consumes all its trailing whitespace, so a fence
# immediately after a list marker on the SAME line is always relative 0,
# and (per case 23) opens/closes normally.
new_fixture
{
  printf -- '- ```bash verified:t\n'
  printf '  echo hi\n'
  printf '  ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "list-marker-prefixed fence (relative 0, same line): exit 0" \
  || fail "list-prefixed fence rel 0: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-8 fix wave — Critical 2: an abort in one change directory
# must not skip a later one, and a real violation elsewhere must still
# outrank the abort's exit code (1 beats 4, exactly as 1 already beats 2).
# Both sort orders are asserted, since the pre-fix bug depended on which
# directory `sorted()` visited first.
# ===========================================================================

# 89. The abort's directory sorts FIRST alphabetically; the violation's
# directory sorts after it. Before the fix, `sys.exit(4)` fired the moment
# the abort directory was scanned and the violation directory was never
# even opened. After the fix: both are scanned, the violation is reported,
# and the violation's exit code (1) outranks the abort's (4).
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/aaa-abort-change" "$FIXTURE/openspec/changes/zzz-violation-change"
{
  printf '    ```bash\n'
  printf '    echo hi\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/aaa-abort-change/tasks.md"
printf 'Baseline: 197 tests\n' > "$FIXTURE/openspec/changes/zzz-violation-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "abort dir sorts first: violation elsewhere still wins (exit 1, not 4)" \
  || fail "abort-first ordering: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"zzz-violation-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment"*) \
    pass "abort dir sorts first: the later directory's violation is still reported, not skipped" ;;
  *) fail "abort-first ordering: violation not reported: out=$OUT" ;;
esac
case "$OUT" in
  *"aaa-abort-change/tasks.md:1: fence-like run indented 4+ columns"*) \
    pass "abort dir sorts first: the abort itself is still reported" ;;
  *) fail "abort-first ordering: abort not reported: out=$OUT" ;;
esac

# 90. Same shapes, reversed sort order: the VIOLATION's directory now sorts
# first, the abort's directory sorts after it. The bug this pins is subtler
# than 89 above: before the fix, even when the violation was found FIRST,
# the abort encountered later still overwrote the exit code to 4 instead of
# leaving it at 1 — "a violation elsewhere still wins" must hold regardless
# of which directory is visited first.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/aaa-violation-change" "$FIXTURE/openspec/changes/zzz-abort-change"
printf 'Baseline: 197 tests\n' > "$FIXTURE/openspec/changes/aaa-violation-change/tasks.md"
{
  printf '    ```bash\n'
  printf '    echo hi\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/zzz-abort-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "violation dir sorts first: the exit code is still 1, not overwritten to 4 by a later abort" \
  || fail "violation-first ordering: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"aaa-violation-change/tasks.md:1: numeric claim with no measured:/predicted: provenance comment"*) \
    pass "violation dir sorts first: the violation is reported" ;;
  *) fail "violation-first ordering: violation not reported: out=$OUT" ;;
esac
case "$OUT" in
  *"zzz-abort-change/tasks.md:1: fence-like run indented 4+ columns"*) \
    pass "violation dir sorts first: the later directory's abort is still scanned and reported" ;;
  *) fail "violation-first ordering: abort not reported: out=$OUT" ;;
esac

# 91. Two directories, ONE of them an abort, NO violations anywhere: the
# final exit code must be 4 (content-classification), not 1 and not 2 —
# proving exit 4 is still reachable as a clean-file's own outcome, not
# merely downgraded to always losing.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/aaa-abort-change" "$FIXTURE/openspec/changes/zzz-clean-change"
{
  printf '    ```bash\n'
  printf '    echo hi\n'
  printf '    ```\n'
} > "$FIXTURE/openspec/changes/aaa-abort-change/tasks.md"
printf '```bash verified:t\necho hi\n```\n' > "$FIXTURE/openspec/changes/zzz-clean-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "an abort beside a CLEAN directory (no violations anywhere): exit 4" \
  || fail "abort beside clean: expected rc=4, got rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-8 fix wave — Important 3: a closing fence is capped at the
# same 0-3 column budget as an opener; an over-indented backtick/tilde run
# no longer closes a fence early.
# ===========================================================================

# 92. THE exact repro from the brief: a bare-opened fence, its would-be
# closer over-indented at 6 columns (past the 0-3 budget), followed by the
# file's REAL closer at column 0. Before the fix, the 6-space run closed
# the fence early, and the real closer below it was then read as a FRESH,
# untagged opener — "fenced code block never closed". After the fix, the
# 6-space run is read as fence CONTENT (not a close), and the real closer
# correctly ends the (verified:) fence — zero violations.
new_fixture
{
  printf '```bash verified:ran it\n'
  printf 'echo start\n'
  printf '      ```\n'
  printf 'Ran 500 tests and it passed.\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "an over-indented (6-column) closer does not close the fence early; the real closer does" \
  || fail "over-indented closer: rc=$RC out=$OUT"
[ "$OUT" = "check-plan-provenance: 1 tasks.md file(s) scanned, all provenance stated" ] \
  && pass "over-indented closer: zero violations reported (no false 'never closed', no false claim)" \
  || fail "over-indented closer: expected zero violations, out=$OUT"

# 92b. Same shape, but the fence is opened inside a blockquote (bq_pre=1):
# the cap must be measured against the RESIDUAL indentation after the
# blockquote marker, not the line's absolute start — an over-indented
# would-be closer still fails to close, and the real (correctly-indented)
# closer still does.
new_fixture
{
  printf '> ```bash verified:ran it\n'
  printf '> echo start\n'
  printf '>       ```\n'
  printf '> Ran 500 tests and it passed.\n'
  printf '> ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "blockquoted over-indented closer does not close the fence early; the real closer does" \
  || fail "blockquoted over-indented closer: rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-9 fix wave — Critical 1: FenceContext now records the list's
# content column (`list_col`), and the closing test requires it exactly,
# not merely the bare 0-3 budget. Every container fixture ABOVE this section
# indents continuations by a flat 2 columns; that is only correct for `- `/
# `* ` (content column 2) and was, for `- [ ] ` (column 6) and the ordered
# markers, coincidentally inside the 0-3 cap rather than actually correct —
# the blind spot this section exists to close. See FenceContext's and
# is_closing_line's docstrings in check-plan-provenance.py for the mechanism.
# ===========================================================================

# marker_width_case <marker> <content_col> <label>
#
# Asserts BOTH directions for one list-marker shape: a fence opened on
# `<marker>```bash verified:...`, with its continuation and closer indented
# to the marker's TRUE content column, passes (exit 0); the same shape with
# no tag is still caught (exit 1, naming the opening line) rather than
# silently misread as closed-too-early or never-closed.
marker_width_case() {
  local marker="$1" col="$2" label="$3" pad
  pad="$(printf '%*s' "$col" '')"

  new_fixture
  {
    printf '%s```bash verified:ran it locally\n' "$marker"
    printf '%secho hi\n' "$pad"
    printf '%s```\n' "$pad"
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  [ "$RC" -eq 0 ] && pass "$label: tagged fence, continuation/closer at true content column ($col): exit 0" \
    || fail "$label: tagged, col=$col: rc=$RC out=$OUT"

  new_fixture
  {
    printf '%s```bash\n' "$marker"
    printf '%secho hi\n' "$pad"
    printf '%s```\n' "$pad"
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  [ "$RC" -ne 0 ] && pass "$label: untagged fence at the same content column is still caught (exit 1)" \
    || fail "$label: untagged, col=$col: expected non-zero, out=$OUT"
  case "$OUT" in
    *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
      pass "$label: untagged fence reports the opening line, not 'never closed'" ;;
    *) fail "$label: untagged, col=$col: wrong/missing message: out=$OUT" ;;
  esac
}

# 93-98. The six marker widths named in the pass-9 fix-wave brief.
marker_width_case '- '      2 "bullet '- '"
marker_width_case '* '      2 "bullet '* '"
marker_width_case '- [ ] '  6 "checkbox '- [ ] '"
marker_width_case '1. '     3 "ordered '1. '"
marker_width_case '10. '    4 "two-digit ordered '10. '"
marker_width_case '1) '     3 "ordered-paren '1) '"

# 99. THE exact false-positive repro this Critical exists to fix: a fence
# opened on a checkbox line (`- [ ] `, true content column 6) with its
# continuation/closer at column 2 — inside the bare 0-3 budget, but NOT the
# marker's actual content column — must NOT be read as closed. Before the
# fix this passed (exit 0) on a fence that was never legitimately closed
# under CommonMark; after the fix it is read as content all the way to EOF.
new_fixture
{
  printf -- '- [ ] ```bash verified:ran it locally\n'
  printf '  echo hi\n'
  printf '  ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "checkbox fence, closer at col 2 (short of true content column 6): no longer a false close" \
  || fail "checkbox short-closer false-positive: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:1: fenced code block never closed"*) \
    pass "checkbox fence, closer at col 2: correctly reported as never closed" ;;
  *) fail "checkbox short-closer false-positive: wrong message: out=$OUT" ;;
esac

# 100. Same defect, ordered two-digit marker (`10. `, true content column 4):
# a closer at column 2 must not falsely close the fence either.
new_fixture
{
  printf '10. ```bash verified:ran it locally\n'
  printf '  echo hi\n'
  printf '  ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "ordered '10. ' fence, closer at col 2 (short of true content column 4): no longer a false close" \
  || fail "ordered short-closer false-positive: expected non-zero, out=$OUT"
case "$OUT" in
  *"tasks.md:1: fenced code block never closed"*) \
    pass "ordered '10. ' fence, closer at col 2: correctly reported as never closed" ;;
  *) fail "ordered short-closer false-positive: wrong message: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-9 fix wave — Critical 2: a 0-3 column gap between a
# blockquote marker and a list marker no longer defeats the classifier.
# Before the fix, the blockquote run tolerated at most ONE extra space
# after `>` (`_BQ_RUN_RE`'s `\s?`, replaced at pass 13 by
# `_consume_bq_run`'s one-COLUMN step);
# a 2+ space gap fell through to unresolved prose, the fence never opened,
# and its closing delimiter was read as a fresh untagged opener that
# silently swallowed everything after it (an unbounded masking, not a
# bounded miss).
# ===========================================================================

# bq_list_gap_case <bq_prefix> <gap_cols> <label>
#
# Builds the exact three-violation shape from the pass-9 brief — an
# untagged fence opened behind `<bq_prefix>` (one or two blockquote levels)
# then <gap_cols> columns of gap then a bullet, an unattributed numeric
# claim after the block closes, and a second untagged bare fence — and
# asserts all three are reported identically regardless of the gap width.
#
# Pass-10 fix wave, Critical 2: the continuation/closer lines repeat
# `<bq_prefix>` then land on the list item's TRUE content column, which is
# the gap PLUS the marker's own width (`gap_pad` + `list_col_pad`) — not
# merely the marker's own width alone. Before this fix, `list_col` dropped
# the gap, so a closer padded with only the marker's width (this fixture's
# previous shape) was accepted as a valid close one-or-more columns short
# of where the list item's content genuinely starts; that is precisely the
# false-early-close this Critical fixes, not a case this fixture should
# exercise. Padding continuation lines to the gap-inclusive column is what
# a real CommonMark-conforming document looks like at this depth, and is
# the shape that must still produce all three violations after the fix.
bq_list_gap_case() {
  local bq_prefix="$1" gap="$2" label="$3" gap_pad list_col_pad content_pad
  gap_pad="$(printf '%*s' "$gap" '')"
  list_col_pad="  "  # "- " is 2 columns wide
  content_pad="${gap_pad}${list_col_pad}"

  new_fixture
  {
    printf '%s%s- ```python\n' "$bq_prefix" "$gap_pad"
    printf '%s%secho untagged fence 1\n' "$bq_prefix" "$content_pad"
    printf '%s%s```\n' "$bq_prefix" "$content_pad"
    printf 'Baseline: 194 tests\n'
    printf '```python\n'
    printf 'echo untagged fence 2\n'
    printf '```\n'
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  [ "$RC" -eq 1 ] && pass "$label (gap=$gap): exit 1, not masked" \
    || fail "$label (gap=$gap): expected rc=1, got rc=$RC out=$OUT"
  case "$OUT" in
    *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
      pass "$label (gap=$gap): first untagged fence reported" ;;
    *) fail "$label (gap=$gap): first untagged fence missing: out=$OUT" ;;
  esac
  case "$OUT" in
    *"tasks.md:4: numeric claim with no measured:/predicted: provenance comment"*) \
      pass "$label (gap=$gap): the '194 tests' claim is reported, not masked" ;;
    *) fail "$label (gap=$gap): claim missing/masked: out=$OUT" ;;
  esac
  case "$OUT" in
    *"tasks.md:5: fenced code block has no verified:/unverified: tag"*) \
      pass "$label (gap=$gap): second untagged fence reported, not swallowed" ;;
    *) fail "$label (gap=$gap): second untagged fence missing: out=$OUT" ;;
  esac
}

# 101-104. Depth 1 (`>`), gap 0/1/2/3 — all four must classify identically.
bq_list_gap_case '> ' 0 "depth-1 blockquote, list-marker gap"
bq_list_gap_case '> ' 1 "depth-1 blockquote, list-marker gap"
bq_list_gap_case '> ' 2 "depth-1 blockquote, list-marker gap"
bq_list_gap_case '> ' 3 "depth-1 blockquote, list-marker gap"

# 105-108. Depth 2 (`> > `), gap 0/1/2/3 — same assertion, one level deeper.
bq_list_gap_case '> > ' 0 "depth-2 blockquote, list-marker gap"
bq_list_gap_case '> > ' 1 "depth-2 blockquote, list-marker gap"
bq_list_gap_case '> > ' 2 "depth-2 blockquote, list-marker gap"
bq_list_gap_case '> > ' 3 "depth-2 blockquote, list-marker gap"

# 108b/108c. THE exact repro from the pass-10 brief's Critical 2: a fence
# opened behind a 1-column gap (`>  - ``` verified:manual`, true content
# column 3 = gap 1 + marker width 2), with its continuation/closer lines at
# the TRUE content column. Before the fix, `list_col` dropped the gap
# (recording 2, the marker's own width only), so a closer at column 2 — one
# short of the true column 3 — was wrongly accepted as a close. After the
# fix, that same one-short candidate must not close, and a candidate at the
# true column 3 must.
new_fixture
{
  printf '>  - ```bash verified:manual\n'
  printf '>    code line one\n'
  printf '>    ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "gap-1 fixture: closer at the TRUE content column (gap+marker) closes correctly, exit 0" \
  || fail "gap-1 fixture (true column): rc=$RC out=$OUT"

# 108c. The same fixture, with a genuine unattributed claim placed AFTER
# the real closer: the fence must be recognised as closed at the right
# line, and the claim afterward reported — not swallowed as still-open
# fence content, and not lost to a false early close either.
new_fixture
{
  printf '>  - ```bash verified:manual\n'
  printf '>    code line one\n'
  printf '>    ```\n'
  printf 'Baseline: 55 tests\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "gap-1 fixture: claim after the real closer is reported, not masked" \
  || fail "gap-1 fixture (claim after real closer): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:4: numeric claim with no measured:/predicted: provenance comment"*) \
    pass "gap-1 fixture: exactly the '55 tests' claim is reported" ;;
  *) fail "gap-1 fixture: claim missing/wrong: out=$OUT" ;;
esac

# 108d. Non-regression: a closer one column SHORT of the true content
# column (matching the OLD, buggy list_col of marker-width-only) must NOT
# be accepted as a close — the fence runs to EOF unclosed instead of being
# silently, falsely closed early.
new_fixture
{
  printf '>  - ```bash verified:manual\n'
  printf '>    code line one\n'
  printf '>   ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -ne 0 ] && pass "gap-1 fixture: a closer one column short of the true content column does not falsely close" \
  || fail "gap-1 fixture (short closer): expected non-zero, out=$OUT"
case "$OUT" in
  *"fenced code block never closed"*) \
    pass "gap-1 fixture: short closer correctly leaves the fence unclosed, not silently closed" ;;
  *) fail "gap-1 fixture (short closer): wrong message: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-10 fix wave — Critical 3: a 4+ column gap between a
# blockquote marker and a list marker is CommonMark-correct as "no fence
# exists" in isolation (Slot 3's reading), but the operator's adjudication
# (recorded in this change's fix-wave brief) is that Slot 1's reading
# governs: since a list is in scope inside this blockquote at all, four
# relative columns is the SAME ambiguity class the bare 4+-column case
# already refuses loudly on (classify_line's "indent" abort) — so this
# case must ALSO refuse loudly (exit 4), never silently decide either way.
# ===========================================================================

# 109. THE exact repro from the brief: a blockquote, a 4+ column gap, then
# a list marker immediately followed by an UNTAGGED fence run. Before this
# fix, the gap defeated `_starts_container_syntax`'s abort-eligibility
# check (it tested the un-stripped remainder, gap still attached, which
# starts with whitespace rather than the list marker itself) and the whole
# line fell through to prose — the fence never opened, and this file's
# scan continued reading everything after it as ordinary lines. After the
# fix, this is recognised as unresolved container syntax behind a gap wider
# than the budget and aborts loudly: exit 4, and scanning of this file
# stops at the abort line (the claim afterward is never reached — a
# consequence of THIS file's scan stopping, not a masking bug; see
# check_file's docstring and the paired-with-a-real-violation cases below
# for the cross-file guarantee that still holds).
new_fixture
{
  printf '>     - ```python\n'
  printf '>     echo x\n'
  printf '>     ```\n'
  printf 'Baseline: 42 tests\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "4+ column gap, list marker, untagged fence: loud abort (exit 4), never silent prose" \
  || fail "4+ column gap after blockquote: expected rc=4, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:1: line has a fence-like run"*"refusing to guess"*) \
    pass "4+ column gap after blockquote: abort names the ambiguous line" ;;
  *) fail "4+ column gap after blockquote: abort message missing: out=$OUT" ;;
esac
case "$OUT" in
  *"all provenance stated"*) fail "4+ column gap after blockquote: false clean run leaked: out=$OUT" ;;
  *) pass "4+ column gap after blockquote: no false clean run" ;;
esac

# 109b. Genuine prose that merely MENTIONS a fence run, sitting behind a
# blockquote and a 4+ column gap with NO list marker in front of it, must
# stay prose — the fix must not over-fire on the general "residual
# whitespace" shape, only on a genuine unresolved container-syntax
# attempt (Critical 3's own text: "keep genuine prose that merely mentions
# a fence classified as prose").
new_fixture
{
  printf '>     Confirmed the fence style uses ``` markers, all good.\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "genuine prose mentioning a fence behind a 4+ column blockquote gap: still prose, no abort" \
  || fail "prose behind wide gap: expected rc=0, got rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-10 fix wave — the harness blind spot. Pass 9 baked a wrong
# shape into its container fixtures; pass 10 found the same class in the
# environment/containment fixtures (cases 51-54 above): every one of them
# pairs the failure with a CLEAN sibling, never a VIOLATING one — the one
# combination that would have caught Critical 1 of this same wave (the
# per-entry/per-file env gates exiting mid-loop instead of deferring to
# main()'s own documented priority). Fixed at the CLASS level: each
# fixture below pairs its failure mode with a sibling directory carrying a
# real, unattributed numeric claim, in BOTH sort orders, and asserts the
# violation is reported and the exit code is 1 — never 2, never 4.
# ===========================================================================

# violating_sibling_case <prefix> <label>
#
# Creates a change directory named "<prefix>-violating-change" (sorts
# either before or after the failure-fixture's own directory name,
# depending on the prefix chosen by the caller) containing one genuine,
# unattributed numeric claim.
violating_sibling_case_dir() {
  local prefix="$1"
  local dir="$FIXTURE/openspec/changes/${prefix}-violating-change"
  mkdir -p "$dir"
  printf 'Baseline: 321 tests\n' > "$dir/tasks.md"
}

# 110-111. Locked (chmod 000) change directory beside a violating sibling,
# both sort orders.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/locked-change"
printf 'irrelevant\n' > "$FIXTURE/openspec/changes/locked-change/tasks.md"
chmod 000 "$FIXTURE/openspec/changes/locked-change"
violating_sibling_case_dir "a"   # "a-violating-change" sorts BEFORE "locked-change"
run_guard "$FIXTURE"
chmod 755 "$FIXTURE/openspec/changes/locked-change"
[ "$RC" -eq 1 ] && pass "locked dir + violating sibling (violation sorts first): exit 1, not masked" \
  || fail "locked dir + violation (violation first): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"321 tests"*"no measured"*|*"numeric claim with no measured"*) \
    pass "locked dir + violation (violation first): violation reported" ;;
  *) fail "locked dir + violation (violation first): violation missing: out=$OUT" ;;
esac

new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/locked-change"
printf 'irrelevant\n' > "$FIXTURE/openspec/changes/locked-change/tasks.md"
chmod 000 "$FIXTURE/openspec/changes/locked-change"
violating_sibling_case_dir "z"   # "z-violating-change" sorts AFTER "locked-change"
run_guard "$FIXTURE"
chmod 755 "$FIXTURE/openspec/changes/locked-change"
[ "$RC" -eq 1 ] && pass "locked dir + violating sibling (violation sorts last): exit 1, not masked" \
  || fail "locked dir + violation (violation last): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "locked dir + violation (violation last): violation reported" ;;
  *) fail "locked dir + violation (violation last): violation missing: out=$OUT" ;;
esac

# 112-113. Dangling symlink change directory beside a violating sibling,
# both sort orders.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
ln -s "$FIXTURE/openspec/changes/does-not-exist" "$FIXTURE/openspec/changes/dangling-change"
violating_sibling_case_dir "a"   # sorts before "dangling-change"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "dangling symlink + violating sibling (violation sorts first): exit 1, not masked" \
  || fail "dangling symlink + violation (violation first): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "dangling symlink + violation (violation first): violation reported" ;;
  *) fail "dangling symlink + violation (violation first): violation missing: out=$OUT" ;;
esac

new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
ln -s "$FIXTURE/openspec/changes/does-not-exist" "$FIXTURE/openspec/changes/dangling-change"
violating_sibling_case_dir "z"   # sorts after "dangling-change"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "dangling symlink + violating sibling (violation sorts last): exit 1, not masked" \
  || fail "dangling symlink + violation (violation last): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "dangling symlink + violation (violation last): violation reported" ;;
  *) fail "dangling symlink + violation (violation last): violation missing: out=$OUT" ;;
esac

# 114-115. Circular symlink (ELOOP) change directory beside a violating
# sibling, both sort orders.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
ln -s "$FIXTURE/openspec/changes/circular-a" "$FIXTURE/openspec/changes/circular-b"
ln -s "$FIXTURE/openspec/changes/circular-b" "$FIXTURE/openspec/changes/circular-a"
violating_sibling_case_dir "a"   # sorts before "circular-a"/"circular-b"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "circular symlink + violating sibling (violation sorts first): exit 1, not masked" \
  || fail "circular symlink + violation (violation first): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "circular symlink + violation (violation first): violation reported" ;;
  *) fail "circular symlink + violation (violation first): violation missing: out=$OUT" ;;
esac

new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
ln -s "$FIXTURE/openspec/changes/circular-a" "$FIXTURE/openspec/changes/circular-b"
ln -s "$FIXTURE/openspec/changes/circular-b" "$FIXTURE/openspec/changes/circular-a"
violating_sibling_case_dir "z"   # sorts after "circular-a"/"circular-b"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "circular symlink + violating sibling (violation sorts last): exit 1, not masked" \
  || fail "circular symlink + violation (violation last): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "circular symlink + violation (violation last): violation reported" ;;
  *) fail "circular symlink + violation (violation last): violation missing: out=$OUT" ;;
esac

# 116-117. Unreadable (chmod 000) tasks.md beside a violating sibling, both
# sort orders.
new_fixture
mkdir -p "$FIXTURE/openspec/changes/unreadable-change"
printf 'irrelevant\n' > "$FIXTURE/openspec/changes/unreadable-change/tasks.md"
chmod 000 "$FIXTURE/openspec/changes/unreadable-change/tasks.md"
violating_sibling_case_dir "a"   # sorts before "unreadable-change"
run_guard "$FIXTURE"
chmod 644 "$FIXTURE/openspec/changes/unreadable-change/tasks.md"
[ "$RC" -eq 1 ] && pass "unreadable tasks.md + violating sibling (violation sorts first): exit 1, not masked" \
  || fail "unreadable tasks.md + violation (violation first): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "unreadable tasks.md + violation (violation first): violation reported" ;;
  *) fail "unreadable tasks.md + violation (violation first): violation missing: out=$OUT" ;;
esac

new_fixture
mkdir -p "$FIXTURE/openspec/changes/unreadable-change"
printf 'irrelevant\n' > "$FIXTURE/openspec/changes/unreadable-change/tasks.md"
chmod 000 "$FIXTURE/openspec/changes/unreadable-change/tasks.md"
violating_sibling_case_dir "z"   # sorts after "unreadable-change"
run_guard "$FIXTURE"
chmod 644 "$FIXTURE/openspec/changes/unreadable-change/tasks.md"
[ "$RC" -eq 1 ] && pass "unreadable tasks.md + violating sibling (violation sorts last): exit 1, not masked" \
  || fail "unreadable tasks.md + violation (violation last): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "unreadable tasks.md + violation (violation last): violation reported" ;;
  *) fail "unreadable tasks.md + violation (violation last): violation missing: out=$OUT" ;;
esac

# 118-119. Bad-encoding tasks.md beside a violating sibling, both sort
# orders (Important 8's original guarantee, re-verified in the specific
# shape the harness blind spot demands).
new_fixture
mkdir -p "$FIXTURE/openspec/changes/bad-encoding-change"
printf '\xff\xfe not valid utf-8\n' > "$FIXTURE/openspec/changes/bad-encoding-change/tasks.md"
violating_sibling_case_dir "a"   # sorts before "bad-encoding-change"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "bad encoding + violating sibling (violation sorts first): exit 1, not masked" \
  || fail "bad encoding + violation (violation first): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "bad encoding + violation (violation first): violation reported" ;;
  *) fail "bad encoding + violation (violation first): violation missing: out=$OUT" ;;
esac

new_fixture
mkdir -p "$FIXTURE/openspec/changes/bad-encoding-change"
printf '\xff\xfe not valid utf-8\n' > "$FIXTURE/openspec/changes/bad-encoding-change/tasks.md"
violating_sibling_case_dir "z"   # sorts after "bad-encoding-change"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "bad encoding + violating sibling (violation sorts last): exit 1, not masked" \
  || fail "bad encoding + violation (violation last): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "bad encoding + violation (violation last): violation reported" ;;
  *) fail "bad encoding + violation (violation last): violation missing: out=$OUT" ;;
esac

# 120-121. Oversized (> MAX_TASKS_MD_BYTES) tasks.md beside a violating
# sibling, both sort orders.
new_fixture
mkdir -p "$FIXTURE/openspec/changes/oversized-change"
python3 -c "open('$FIXTURE/openspec/changes/oversized-change/tasks.md', 'wb').write(b'a' * (10 * 1024 * 1024 + 1))"
violating_sibling_case_dir "a"   # sorts before "oversized-change"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "oversized tasks.md + violating sibling (violation sorts first): exit 1, not masked" \
  || fail "oversized tasks.md + violation (violation first): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "oversized tasks.md + violation (violation first): violation reported" ;;
  *) fail "oversized tasks.md + violation (violation first): violation missing: out=$OUT" ;;
esac

new_fixture
mkdir -p "$FIXTURE/openspec/changes/oversized-change"
python3 -c "open('$FIXTURE/openspec/changes/oversized-change/tasks.md', 'wb').write(b'a' * (10 * 1024 * 1024 + 1))"
violating_sibling_case_dir "z"   # sorts after "oversized-change"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "oversized tasks.md + violating sibling (violation sorts last): exit 1, not masked" \
  || fail "oversized tasks.md + violation (violation last): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "oversized tasks.md + violation (violation last): violation reported" ;;
  *) fail "oversized tasks.md + violation (violation last): violation missing: out=$OUT" ;;
esac

# 122-123. A content-classification abort (the pass-8/9/10 "prefix"/"indent"
# refusal) beside a REAL violation in a separate change directory, both
# sort orders — the abort must stop only ITS OWN file's scan; the
# violation elsewhere must still be reported and win the exit code (exit 1
# outranks exit 4, per main()'s own documented priority).
new_fixture
mkdir -p "$FIXTURE/openspec/changes/abort-change"
printf -- '- - ```bash\necho hi\n```\n' > "$FIXTURE/openspec/changes/abort-change/tasks.md"
violating_sibling_case_dir "a"   # sorts before "abort-change"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "classification abort + violating sibling (violation sorts first): exit 1, not masked by exit 4" \
  || fail "abort + violation (violation first): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "abort + violation (violation first): violation reported" ;;
  *) fail "abort + violation (violation first): violation missing: out=$OUT" ;;
esac

new_fixture
mkdir -p "$FIXTURE/openspec/changes/abort-change"
printf -- '- - ```bash\necho hi\n```\n' > "$FIXTURE/openspec/changes/abort-change/tasks.md"
violating_sibling_case_dir "z"   # sorts after "abort-change"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "classification abort + violating sibling (violation sorts last): exit 1, not masked by exit 4" \
  || fail "abort + violation (violation last): expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "abort + violation (violation last): violation reported" ;;
  *) fail "abort + violation (violation last): violation missing: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-11 fix wave — Critical 1: the outermost leading indentation
# in front of a list marker is now folded into `list_content_col`, not
# discarded. Pass 9 fixed the blockquote-to-list GAP; pass 10 fixed that
# gap's width not reaching `list_content_col`; this is the third layer —
# the indentation BEFORE any of that (in front of the marker itself, when
# no blockquote is present) was measured, then thrown away. Characterised
# by the parent across the full space: marker indent 0-1 correct at every
# closer position; indent 2 breaks at content_col+2; indent 3 breaks at
# +1 and +2; indent 4-5 break at EVERY position, including the exact
# content column. The matrix below covers indents 0-5 at closer offsets
# 0/+1/+2/+3 (all four are inside CommonMark's 0-3 budget once the true
# content column is reached, so all four must close identically), tagged
# (exit 0) and untagged (exit 1, reporting the OPENING line's missing tag —
# never a false "never closed").
#
# CORRECTED at pass 12 (Critical 2): rows for indent=4 and indent=5 used to
# assert exit 0 (a real, accepted list fence) at every closer offset. That
# was itself Critical 2 of this pass encoded as the specification — CommonMark
# (https://spec.commonmark.org/0.30/#indented-code-blocks and
# https://spec.commonmark.org/0.30/#list-items) caps UN-CONTAINED indentation
# at 0-3 columns before ANY container marker is recognised at all: a `-`
# beginning 4+ columns into an otherwise bare line, with no enclosing list
# already in scope, is not a list marker under CommonMark — it is literal
# text inside an indented code block. This parser cannot tell "indent 4
# inside an already-open list" from "indent 4 at the top level" apart (the
# cross-line state that could is deliberately deleted — see classify_line's
# docstring on the pass-7 ambient tracker and its three Criticals), so it
# must refuse (exit 4) at indent 4+ exactly as the BARE case (no marker at
# all) already does at the same indent — never accept it just because a
# marker token happens to be present. Verified against CommonMark, not
# against the guard's own prior output: the parent's earlier "all 24
# correct" verification (pass 11) checked these rows against the fix's own
# intent, not against the spec, which is what let this encode a defect as
# a passing test for an entire fix wave.
# ===========================================================================

# assert_over_indented_marker_message <label>
#
# Minor 6 (pass-13 fix wave): every indent-4/5 row below asserted ONLY the
# exit code, so the abort message they produce was never read by any
# assertion — which is exactly how the wrong message shipped. A line that
# DOES carry a `- ` or `> ` marker but exceeds the outer-indent cap was
# handed the BARE case's text ("with no container prefix on this line …
# Fix by … adding an explicit list/blockquote marker on this line"): both
# clauses false there, and the named remedy is to add a marker the line
# already has. These fixtures all carry a marker, so the message must name
# the outer indentation as the cause, and must NOT claim the line has no
# container prefix.
assert_over_indented_marker_message() {
  local label="$1"
  case "$OUT" in
    *"container prefix starting 4+ columns in"*) \
      pass "$label: abort message names the over-indented container prefix as the cause" ;;
    *) fail "$label: abort message does not name the over-indented container prefix: out=$OUT" ;;
  esac
  case "$OUT" in
    *"with no container prefix on this line"*) \
      fail "$label: abort message wrongly reuses the bare-line text on a line that HAS a marker: out=$OUT" ;;
    *) pass "$label: abort message does not claim the line has no container prefix" ;;
  esac
}

# indent_closer_case <indent> <offset> <label>
#
# Builds a `- ` bullet fence (marker width 2) indented `<indent>` columns,
# with its continuation/closer lines padded to the true content column
# (`indent + 2`) plus `<offset>` extra columns (0-3, all inside the bare
# budget).
#
# indent 0-3: asserts the tagged variant closes cleanly (exit 0) and the
# untagged variant is still caught by name (exit 1, opening line, "no
# verified:/unverified: tag") rather than misread as never-closed (the
# false-refusal direction pass 11's fix addressed) or masking anything
# after it — unchanged behaviour, CommonMark's 0-3 column budget applies
# normally once a marker legitimately starts the container.
#
# indent 4-5: asserts BOTH variants refuse (exit 4) regardless of the tag
# — a classification abort fires before the tag is ever inspected, because
# the outer indent alone already makes the marker's legitimacy
# undecidable (see the CommonMark citation in this section's header).
indent_closer_case() {
  local indent="$1" offset="$2" label="$3" marker_pad content_col pad
  # Minor 10 (pass-12 fix wave): `bullet_marker_width` duplicates
  # knowledge the production code owns (the width of "- ", the exact
  # literal used two lines below to build the fixture) — this harness has
  # no access to check-plan-provenance.py's own `_LIST_RE` match width at
  # bash-script time, so the duplication cannot be removed outright. Named
  # and placed immediately next to the literal it must stay in sync with,
  # rather than left as a bare "+ 2", so a future change to the marker
  # literal below is visibly adjacent to the arithmetic that depends on it.
  local bullet_marker_width=2

  marker_pad="$(printf '%*s' "$indent" '')"
  content_col=$((indent + bullet_marker_width))
  pad="$(printf '%*s' "$((content_col + offset))" '')"

  if [ "$indent" -gt 3 ]; then
    new_fixture
    {
      printf '%s- ```bash verified:ran it locally\n' "$marker_pad"
      printf '%secho hi\n' "$pad"
      printf '%s```\n' "$pad"
    } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
    run_guard "$FIXTURE"
    [ "$RC" -eq 4 ] && pass "$label: tagged, indent=$indent, closer=content_col+$offset: exit 4 (outer indent capped before container recognition)" \
      || fail "$label: tagged, indent=$indent, closer=content_col+$offset: expected rc=4, got rc=$RC out=$OUT"
    assert_over_indented_marker_message "$label: tagged, indent=$indent, closer=content_col+$offset"

    new_fixture
    {
      printf '%s- ```bash\n' "$marker_pad"
      printf '%secho hi\n' "$pad"
      printf '%s```\n' "$pad"
    } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
    run_guard "$FIXTURE"
    [ "$RC" -eq 4 ] && pass "$label: untagged, indent=$indent, closer=content_col+$offset: exit 4 (tag is irrelevant to a refusal)" \
      || fail "$label: untagged, indent=$indent, closer=content_col+$offset: expected rc=4, got rc=$RC out=$OUT"
    assert_over_indented_marker_message "$label: untagged, indent=$indent, closer=content_col+$offset"
    return
  fi

  new_fixture
  {
    printf '%s- ```bash verified:ran it locally\n' "$marker_pad"
    printf '%secho hi\n' "$pad"
    printf '%s```\n' "$pad"
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  [ "$RC" -eq 0 ] && pass "$label: tagged, indent=$indent, closer=content_col+$offset: exit 0" \
    || fail "$label: tagged, indent=$indent, closer=content_col+$offset: rc=$RC out=$OUT"

  new_fixture
  {
    printf '%s- ```bash\n' "$marker_pad"
    printf '%secho hi\n' "$pad"
    printf '%s```\n' "$pad"
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  [ "$RC" -eq 1 ] && pass "$label: untagged, indent=$indent, closer=content_col+$offset: exit 1" \
    || fail "$label: untagged, indent=$indent, closer=content_col+$offset: expected rc=1, got rc=$RC out=$OUT"
  case "$OUT" in
    *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
      pass "$label: untagged, indent=$indent, closer=content_col+$offset: opening line reported, not 'never closed'" ;;
    *) fail "$label: untagged, indent=$indent, closer=content_col+$offset: wrong/missing message: out=$OUT" ;;
  esac
}

for indent in 0 1 2 3 4 5; do
  for offset in 0 1 2 3; do
    indent_closer_case "$indent" "$offset" "marker-indent x closer-column (bullet '- ')"
  done
done

# blockquote_indent_case <indent> <offset> <label>
#
# The blockquote-marker equivalent of indent_closer_case above, named
# explicitly in the fix-wave brief as having "the same defect and no
# coverage at all". Unlike a list marker, a blockquote marker MUST repeat
# on every continuation/closing line under this parser's grammar (no lazy
# continuation — see is_closing_line's docstring), so the content/closer
# lines here repeat the same leading indent plus "> ", then `<offset>`
# columns of residual indentation (0-3) before the fence run itself.
#
# indent 0-3: unchanged from before this pass — CommonMark
# (https://spec.commonmark.org/0.30/#block-quotes) allows a block quote
# marker to be preceded by up to 3 spaces of indentation; the fence is a
# real, container-scoped fence exactly as today.
#
# indent 4-5: CommonMark's indented-code-block rule (same citation as
# indent_closer_case above) applies identically to a ">" marker as to a
# "-" marker — 4+ columns of outer indentation with no enclosing list
# already in scope makes the line an indented code block, and the ">"
# there is literal text, never a real blockquote marker. Both variants
# must refuse (exit 4), the same as the bullet case.
blockquote_indent_case() {
  local indent="$1" offset="$2" label="$3" marker_pad pad

  marker_pad="$(printf '%*s' "$indent" '')"
  pad="$(printf '%*s' "$offset" '')"

  if [ "$indent" -gt 3 ]; then
    new_fixture
    {
      printf '%s> ```bash verified:ran it locally\n' "$marker_pad"
      printf '%s> %secho hi\n' "$marker_pad" "$pad"
      printf '%s> %s```\n' "$marker_pad" "$pad"
    } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
    run_guard "$FIXTURE"
    [ "$RC" -eq 4 ] && pass "$label: tagged, indent=$indent, closer=+$offset: exit 4 (outer indent capped before container recognition)" \
      || fail "$label: tagged, indent=$indent, closer=+$offset: expected rc=4, got rc=$RC out=$OUT"
    assert_over_indented_marker_message "$label: tagged, indent=$indent, closer=+$offset"

    new_fixture
    {
      printf '%s> ```bash\n' "$marker_pad"
      printf '%s> %secho hi\n' "$marker_pad" "$pad"
      printf '%s> %s```\n' "$marker_pad" "$pad"
    } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
    run_guard "$FIXTURE"
    [ "$RC" -eq 4 ] && pass "$label: untagged, indent=$indent, closer=+$offset: exit 4 (tag is irrelevant to a refusal)" \
      || fail "$label: untagged, indent=$indent, closer=+$offset: expected rc=4, got rc=$RC out=$OUT"
    assert_over_indented_marker_message "$label: untagged, indent=$indent, closer=+$offset"
    return
  fi

  new_fixture
  {
    printf '%s> ```bash verified:ran it locally\n' "$marker_pad"
    printf '%s> %secho hi\n' "$marker_pad" "$pad"
    printf '%s> %s```\n' "$marker_pad" "$pad"
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  [ "$RC" -eq 0 ] && pass "$label: tagged, indent=$indent, closer=+$offset: exit 0" \
    || fail "$label: tagged, indent=$indent, closer=+$offset: rc=$RC out=$OUT"

  new_fixture
  {
    printf '%s> ```bash\n' "$marker_pad"
    printf '%s> %secho hi\n' "$marker_pad" "$pad"
    printf '%s> %s```\n' "$marker_pad" "$pad"
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  [ "$RC" -eq 1 ] && pass "$label: untagged, indent=$indent, closer=+$offset: exit 1" \
    || fail "$label: untagged, indent=$indent, closer=+$offset: expected rc=1, got rc=$RC out=$OUT"
  case "$OUT" in
    *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
      pass "$label: untagged, indent=$indent, closer=+$offset: opening line reported, not 'never closed'" ;;
    *) fail "$label: untagged, indent=$indent, closer=+$offset: wrong/missing message: out=$OUT" ;;
  esac
}

for indent in 0 1 2 3 4 5; do
  for offset in 0 1 2 3; do
    blockquote_indent_case "$indent" "$offset" "marker-indent x closer-column (blockquote '> ')"
  done
done

# ===========================================================================
# SECTION: Pass-11 fix wave — Critical 2: a containment refusal (exit 3)
# must not stop the scan, and must still outrank a violation found
# elsewhere. `violating_sibling_case_dir` (pass-10) was used 16 times above
# and NONE of them was a containment case — which is why this Critical
# survived eleven passes. Extended here to every containment branch this
# guard's grammar recognises, in both sort orders: the sibling's violation
# must still be reported, and the exit code must still be 3 (containment
# outranks a violation found elsewhere, per the adjudicated priority
# ladder — see check-plan-provenance.py's module docstring).
#
# NOTE on branch count: `verify_tasks_md_containment` has eight `sys.exit`
# (now `return False`) sites, but they collapse to four DISTINCT
# constructible fixtures — "symlinked change directory" and "directory
# escape" are the SAME source-code branch (`real_dir != expected`): a
# symlinked change-directory entry can never resolve to exactly
# `<changes_dir>/<name>` (a symlink's realpath, by definition, is never the
# symlink's own path), so any symlinked change directory always fails
# containment as a "directory escape", whether it points outside the repo
# entirely or merely to another directory inside `openspec/changes/`. Both
# shapes are exercised below for honesty about what's actually being
# constructed, even though the guard code takes one branch for both. The
# other two `sys.exit(3)` pairs (`cannot resolve containing directory` /
# `cannot resolve {changes_dir}`) require a directory to disappear or
# become unstatable BETWEEN the initial `os.lstat` and containment's own
# `os.stat` calls — a real race, not a fixture buildable with `mktemp -d`
# and ordinary POSIX tools — and are not exercised here; see the module's
# "KNOWN LIMITATION — containment/read TOCTOU" note.
# ===========================================================================

# 124-125. `tasks.md` itself is a symlink, beside a violating sibling, both
# sort orders.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/symlinked-tasks-change"
ln -sf /dev/zero "$FIXTURE/openspec/changes/symlinked-tasks-change/tasks.md"
violating_sibling_case_dir "a"   # sorts before "symlinked-tasks-change"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "symlinked tasks.md + violating sibling (violation sorts first): exit 3, scan continues" \
  || fail "symlinked tasks.md + violation (violation first): expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "symlinked tasks.md + violation (violation first): sibling violation reported" ;;
  *) fail "symlinked tasks.md + violation (violation first): sibling violation missing: out=$OUT" ;;
esac
case "$OUT" in
  *"tasks.md is a symlink"*) pass "symlinked tasks.md + violation (violation first): containment message reported" ;;
  *) fail "symlinked tasks.md + violation (violation first): containment message missing: out=$OUT" ;;
esac

new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/symlinked-tasks-change"
ln -sf /dev/zero "$FIXTURE/openspec/changes/symlinked-tasks-change/tasks.md"
violating_sibling_case_dir "z"   # sorts after "symlinked-tasks-change"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "symlinked tasks.md + violating sibling (violation sorts last): exit 3, scan continues" \
  || fail "symlinked tasks.md + violation (violation last): expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "symlinked tasks.md + violation (violation last): sibling violation reported" ;;
  *) fail "symlinked tasks.md + violation (violation last): sibling violation missing: out=$OUT" ;;
esac

# 126-127. Change directory reached via a symlink pointing OUTSIDE the
# repository's changes tree entirely ("directory escape"), beside a
# violating sibling, both sort orders.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/outside-changes/escaped-change"
printf 'irrelevant\n' > "$FIXTURE/outside-changes/escaped-change/tasks.md"
ln -s "$FIXTURE/outside-changes/escaped-change" "$FIXTURE/openspec/changes/escaped-change"
violating_sibling_case_dir "a"   # sorts before "escaped-change"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "directory escape + violating sibling (violation sorts first): exit 3, scan continues" \
  || fail "directory escape + violation (violation first): expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "directory escape + violation (violation first): sibling violation reported" ;;
  *) fail "directory escape + violation (violation first): sibling violation missing: out=$OUT" ;;
esac
case "$OUT" in
  *"outside"*"symlink escape"*) pass "directory escape + violation (violation first): containment message reported" ;;
  *) fail "directory escape + violation (violation first): containment message missing: out=$OUT" ;;
esac

new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/outside-changes/escaped-change"
printf 'irrelevant\n' > "$FIXTURE/outside-changes/escaped-change/tasks.md"
ln -s "$FIXTURE/outside-changes/escaped-change" "$FIXTURE/openspec/changes/escaped-change"
violating_sibling_case_dir "z"   # sorts after "escaped-change"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "directory escape + violating sibling (violation sorts last): exit 3, scan continues" \
  || fail "directory escape + violation (violation last): expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "directory escape + violation (violation last): sibling violation reported" ;;
  *) fail "directory escape + violation (violation last): sibling violation missing: out=$OUT" ;;
esac

# 128-129. Change directory entry is itself a symlink, but pointing to
# ANOTHER real change directory INSIDE `openspec/changes/` (same source
# branch as 126-127 — see this section's NOTE — different fixture shape),
# beside a violating sibling, both sort orders.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/real-target-change"
printf 'irrelevant\n' > "$FIXTURE/openspec/changes/real-target-change/tasks.md"
ln -s "$FIXTURE/openspec/changes/real-target-change" "$FIXTURE/openspec/changes/aliased-change"
violating_sibling_case_dir "a"   # sorts before "aliased-change"/"real-target-change"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "symlinked change directory (in-tree target) + violating sibling (violation sorts first): exit 3" \
  || fail "symlinked change dir (in-tree) + violation (violation first): expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "symlinked change directory (in-tree target) + violation (violation first): sibling violation reported" ;;
  *) fail "symlinked change dir (in-tree) + violation (violation first): sibling violation missing: out=$OUT" ;;
esac

new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/real-target-change"
printf 'irrelevant\n' > "$FIXTURE/openspec/changes/real-target-change/tasks.md"
ln -s "$FIXTURE/openspec/changes/real-target-change" "$FIXTURE/openspec/changes/aliased-change"
violating_sibling_case_dir "z"   # sorts after "aliased-change"/"real-target-change"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "symlinked change directory (in-tree target) + violating sibling (violation sorts last): exit 3" \
  || fail "symlinked change dir (in-tree) + violation (violation last): expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "symlinked change directory (in-tree target) + violation (violation last): sibling violation reported" ;;
  *) fail "symlinked change dir (in-tree) + violation (violation last): sibling violation missing: out=$OUT" ;;
esac

# 130-131. `tasks.md` is a non-regular file (a FIFO) after containment's own
# directory-escape checks pass, beside a violating sibling, both sort
# orders. `os.stat` on a FIFO never blocks (unlike `open()`), so this is
# safe to construct and check without a reader/writer on the other end.
new_fixture
mkdir -p "$FIXTURE/openspec/changes/fifo-change"
mkfifo "$FIXTURE/openspec/changes/fifo-change/tasks.md"
violating_sibling_case_dir "a"   # sorts before "fifo-change"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "non-regular tasks.md (FIFO) + violating sibling (violation sorts first): exit 3, scan continues" \
  || fail "FIFO tasks.md + violation (violation first): expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "FIFO tasks.md + violation (violation first): sibling violation reported" ;;
  *) fail "FIFO tasks.md + violation (violation first): sibling violation missing: out=$OUT" ;;
esac
case "$OUT" in
  *"not a regular file after resolution"*) pass "FIFO tasks.md + violation (violation first): containment message reported" ;;
  *) fail "FIFO tasks.md + violation (violation first): containment message missing: out=$OUT" ;;
esac

new_fixture
mkdir -p "$FIXTURE/openspec/changes/fifo-change"
mkfifo "$FIXTURE/openspec/changes/fifo-change/tasks.md"
violating_sibling_case_dir "z"   # sorts after "fifo-change"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "non-regular tasks.md (FIFO) + violating sibling (violation sorts last): exit 3, scan continues" \
  || fail "FIFO tasks.md + violation (violation last): expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) \
    pass "FIFO tasks.md + violation (violation last): sibling violation reported" ;;
  *) fail "FIFO tasks.md + violation (violation last): sibling violation missing: out=$OUT" ;;
esac

# 132. A containment refusal beside a content-classification abort AND a
# real violation, across several directories, several sort orders — every
# failure kind must still be reported, and containment (exit 3) must still
# win the exit code over the abort (exit 4) and the violation (exit 1).
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/abort-change"
printf -- '- - ```bash\necho hi\n```\n' > "$FIXTURE/openspec/changes/abort-change/tasks.md"
mkdir -p "$FIXTURE/openspec/changes/m-symlinked-tasks-change"
ln -sf /dev/zero "$FIXTURE/openspec/changes/m-symlinked-tasks-change/tasks.md"
violating_sibling_case_dir "a"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "containment + abort + violation, all three directories: exit 3 wins" \
  || fail "containment + abort + violation: expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"numeric claim with no measured"*) pass "containment + abort + violation: violation still reported" ;;
  *) fail "containment + abort + violation: violation missing: out=$OUT" ;;
esac
case "$OUT" in
  *"fence-like run of backticks/tildes behind a prefix"*) pass "containment + abort + violation: abort still reported" ;;
  *) fail "containment + abort + violation: abort missing: out=$OUT" ;;
esac
case "$OUT" in
  *"tasks.md is a symlink"*) pass "containment + abort + violation: containment message still reported" ;;
  *) fail "containment + abort + violation: containment message missing: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Minor 8 (pass-12 fix wave) — every containment combination
# fixture above pairs a containment refusal with a real VIOLATION
# (`violating_sibling_case_dir`). None of them pairs containment with an
# environment-failure-only sibling or an abort-only sibling, so the
# 3-over-2 and 3-over-4 ladder edges (containment still winning the exit
# code over an environment failure, and over a classification abort, with
# NO violation anywhere to also demonstrate exit 1 winning over exit 3 —
# which it must not, since 3 outranks 1 too) were unverified.
# ===========================================================================

# 132a. Containment refusal beside an environment-failure-only sibling (no
# violation, no abort anywhere): exit 3 must still win over exit 2, and
# both messages must still be reported.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/unreadable-change"
printf 'no claims, no fences\n' > "$FIXTURE/openspec/changes/unreadable-change/tasks.md"
chmod 000 "$FIXTURE/openspec/changes/unreadable-change/tasks.md"
mkdir -p "$FIXTURE/openspec/changes/m-symlinked-tasks-change"
ln -sf /dev/zero "$FIXTURE/openspec/changes/m-symlinked-tasks-change/tasks.md"
run_guard "$FIXTURE"
chmod 644 "$FIXTURE/openspec/changes/unreadable-change/tasks.md" 2>/dev/null || true
[ "$RC" -eq 3 ] && pass "containment + environment-failure-only sibling: exit 3 wins" \
  || fail "containment + environment-only: expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"cannot read file"*) pass "containment + environment-only: environment failure still reported" ;;
  *) fail "containment + environment-only: environment failure missing: out=$OUT" ;;
esac
case "$OUT" in
  *"tasks.md is a symlink"*) pass "containment + environment-only: containment message still reported" ;;
  *) fail "containment + environment-only: containment message missing: out=$OUT" ;;
esac

# 132b. Containment refusal beside an abort-only sibling (no violation, no
# environment failure anywhere): exit 3 must still win over exit 4, and
# both messages must still be reported.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/abort-change"
printf -- '- - ```bash\necho hi\n```\n' > "$FIXTURE/openspec/changes/abort-change/tasks.md"
mkdir -p "$FIXTURE/openspec/changes/m-symlinked-tasks-change"
ln -sf /dev/zero "$FIXTURE/openspec/changes/m-symlinked-tasks-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "containment + abort-only sibling: exit 3 wins" \
  || fail "containment + abort-only: expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"fence-like run of backticks/tildes behind a prefix"*) pass "containment + abort-only: abort still reported" ;;
  *) fail "containment + abort-only: abort missing: out=$OUT" ;;
esac
case "$OUT" in
  *"tasks.md is a symlink"*) pass "containment + abort-only: containment message still reported" ;;
  *) fail "containment + abort-only: containment message missing: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Important 4 (pass-12 fix wave) — assert the EXACT wording of all
# four summary banners, not merely that "some" message appeared. The
# pass-11 environment-failure banner
# ("(environment failure — a permission, encoding or size problem on disk,
# see above)") was introduced specifically to stop it being confused with
# the classification banner's near-identical "could not be classified"
# phrasing — before this section, the ONLY occurrence of that exact string
# anywhere in this harness was inside a comment, so a refactor could
# silently restore the ambiguous wording and every assertion above would
# stay green. Each fixture below produces exactly ONE of the four channels
# in isolation, so the banner asserted is unambiguously the one under test.
# ===========================================================================

# 133. Exit 1 (violations) banner, in isolation: an untagged fence with no
# other channel present.
new_fixture
printf 'intro\n```bash\necho hi\n```\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "exit 1 banner fixture: rc=1" || fail "exit 1 banner fixture: rc=$RC out=$OUT"
case "$OUT" in
  *"1 plan-provenance violation(s) found."*) pass "exit 1 banner: exact wording" ;;
  *) fail "exit 1 banner: wording changed or missing: out=$OUT" ;;
esac
case "$OUT" in
  *"For the violations above: fix by adding the missing verified:/unverified: tag or measured:/predicted: comment; never suppress."*) \
    pass "exit 1 remedy line: exact wording" ;;
  *) fail "exit 1 remedy line: wording changed or missing: out=$OUT" ;;
esac

# 134. Exit 4 (content-classification) banner, in isolation: a doubled
# list marker, no violation and no file-read failure anywhere.
new_fixture
printf -- '- - ```bash\necho hi\n```\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "exit 4 banner fixture: rc=4" || fail "exit 4 banner fixture: rc=$RC out=$OUT"
case "$OUT" in
  *"1 line(s) could not be classified (content-classification, see above) — refusing to report a clean run"*) \
    pass "exit 4 banner: exact wording" ;;
  *) fail "exit 4 banner: wording changed or missing: out=$OUT" ;;
esac

# 135. Exit 2 (environment) banner, in isolation: an unreadable tasks.md
# (chmod 000) that has already passed containment, no violation and no
# classification abort anywhere. This is the banner Important 4 exists to
# pin: it must say "environment failure — a permission, encoding or size
# problem on disk", never "could not be classified" (exit 4's phrase).
new_fixture
printf 'no claims, no fences here\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
chmod 000 "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
chmod 644 "$FIXTURE/openspec/changes/demo-change/tasks.md" 2>/dev/null || true
[ "$RC" -eq 2 ] && pass "exit 2 banner fixture: rc=2" || fail "exit 2 banner fixture: rc=$RC out=$OUT"
case "$OUT" in
  *"1 tasks.md file(s) could not be read (environment failure — a permission, encoding or size problem on disk, see above) — refusing to report a clean run"*) \
    pass "exit 2 banner: exact wording, distinct from exit 4's 'could not be classified'" ;;
  *) fail "exit 2 banner: wording changed or missing: out=$OUT" ;;
esac
case "$OUT" in
  *"could not be classified"*) fail "exit 2 banner: wrongly reused exit 4's 'could not be classified' phrasing: out=$OUT" ;;
  *) pass "exit 2 banner: does not reuse exit 4's phrasing" ;;
esac

# 136. Exit 3 (containment/security) banner, in isolation: a symlinked
# tasks.md, no violation, no abort and no unrelated file-read failure
# anywhere else in the fixture.
#
# The noun in this banner is "path(s)", not "tasks.md file(s)", since
# Critical 5 of the pass-14 fix wave: the same counter now also carries a
# refusal on `openspec/changes` ITSELF, which is a directory, not a
# tasks.md. Naming it "tasks.md file(s)" would have made the banner state
# something false for that case — the shape this change exists to police.
# The exit code, the stream and the rest of the sentence are unchanged.
new_fixture
rm -rf "$FIXTURE/openspec/changes/demo-change"
mkdir -p "$FIXTURE/openspec/changes/demo-change"
ln -sf /dev/zero "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "exit 3 banner fixture: rc=3" || fail "exit 3 banner fixture: rc=$RC out=$OUT"
case "$OUT" in
  *"1 path(s) failed containment (symlinked, escaping their change directory, or non-regular — see above) — refusing to report anything but a security failure. Every violation, abort or file error listed above from OTHER directories was still found and still needs fixing; fix the containment escape(s) first."*) \
    pass "exit 3 banner: exact wording" ;;
  *) fail "exit 3 banner: wording changed or missing: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-13 fix wave — Critical 1: a blockquote marker consumes at
# most ONE COLUMN of the whitespace after its `>`, not one whitespace
# CHARACTER. CommonMark
# (https://spec.commonmark.org/0.30/#block-quotes, and the tab rule at
# https://spec.commonmark.org/0.30/#tabs — spec example 6, `>\t\tfoo`,
# renders as an indented code block containing "  foo") is explicit: the
# block quote marker is `>` plus ONE following space, and when that space
# is a tab, only the tab's FIRST column belongs to the marker — its
# remaining columns stay as ordinary indentation of the quoted content.
#
# Every expected column below is derived from that rule by hand and stated
# in the comment above the fixture, never read off this guard's own output.
# The one pre-existing tab-after-blockquote fixture (case 78d) puts the tab
# in the PROSE after "> " — i.e. after the marker's one space was already
# consumed — which is why this survived twelve passes.
# ===========================================================================

# 137a. THE reported silent miss. `>` occupies column 0; the tab spans
# columns 1-3, of which the marker takes column 1 only, leaving columns 2-3
# (2 columns) as quoted-content indentation; the 3 spaces take it to 5
# columns of indentation inside the blockquote. 5 >= 4, so under CommonMark
# line 1 is an INDENTED CODE BLOCK inside the blockquote and its backtick
# run is literal text — not a fence opener at all. Before this fix the tab
# was swallowed whole, the run measured 3 columns, a (tagged) fence opened,
# the genuinely untagged fence on line 3 was swallowed as its content, and
# line 6 closed the phantom fence "cleanly" — exit 0, zero findings.
new_fixture
{
  printf '>\t   ```text verified:x\n'
  printf 'hello\n'
  printf '```bash\n'
  printf 'echo hi\n'
  printf '```\n'
  printf '> ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "tab after '>' + trailing close: exit 1, no longer a silent clean run" \
  || fail "tab after '>' + trailing close: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:3: fenced code block has no verified:/unverified: tag"*) \
    pass "tab after '>' + trailing close: the swallowed untagged fence on line 3 is reported" ;;
  *) fail "tab after '>' + trailing close: line 3 not reported: out=$OUT" ;;
esac
case "$OUT" in
  *"all provenance stated"*) fail "tab after '>' + trailing close: false clean run leaked: out=$OUT" ;;
  *) pass "tab after '>' + trailing close: no false clean run" ;;
esac

# 137b. The control: the same file with the tab replaced by the spaces it
# is worth in COLUMNS — one space for the marker's own column plus the 5
# columns of content indentation derived above ('>' + 6 spaces). This shape
# was already handled correctly before the fix; asserting it beside 137a is
# what makes "same true column width, same verdict" a property of the guard
# rather than a coincidence of which whitespace character was typed.
new_fixture
{
  printf '>      ```text verified:x\n'
  printf 'hello\n'
  printf '```bash\n'
  printf 'echo hi\n'
  printf '```\n'
  printf '> ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "6 spaces after '>' (same column width as the tab shape): exit 1" \
  || fail "6 spaces after '>': expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:3: fenced code block has no verified:/unverified: tag"*) \
    pass "6 spaces after '>': the untagged fence on line 3 is reported, exactly as in the tab shape" ;;
  *) fail "6 spaces after '>': line 3 not reported: out=$OUT" ;;
esac

# 137c. The tab shape with the trailing close removed. Before the fix this
# reported only "never closed" on line 1 (the phantom fence); the untagged
# fence on line 3 was still never named. The verdict must now be the same
# as 137a for the same reason — line 1 is not a fence at all.
new_fixture
{
  printf '>\t   ```text verified:x\n'
  printf 'hello\n'
  printf '```bash\n'
  printf 'echo hi\n'
  printf '```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "tab after '>' without a trailing close: exit 1" \
  || fail "tab after '>' without a trailing close: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:3: fenced code block has no verified:/unverified: tag"*) \
    pass "tab after '>' without a trailing close: line 3's untagged fence is reported, not line 1's phantom fence" ;;
  *) fail "tab after '>' without a trailing close: line 3 not reported: out=$OUT" ;;
esac

# 138. The false-positive half of the same root cause, with NO list marker
# anywhere: `>` + tab + 3 spaces + a bare backtick run is, by the column
# arithmetic in 137a, 5 columns of indentation inside the blockquote —
# indented code, literal text. Before the fix it opened a fence that was
# never closed, inventing a violation on a file that has none.
new_fixture
printf '>\t   ```bash\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "no-list-marker false positive ('>' + tab + 3 spaces + fence run): not a fence, exit 0" \
  || fail "no-list-marker false positive: expected rc=0, got rc=$RC out=$OUT"
case "$OUT" in
  *"never closed"*) fail "no-list-marker false positive: still opens a phantom fence: out=$OUT" ;;
  *) pass "no-list-marker false positive: no phantom fence opened" ;;
esac

# bq_tab_list_closer_case <n_spaces> <expect_rc> <label>
#
# Depth-1 blockquote, tab immediately after the `>`, then a bullet.
# Columns: '>' at col 0; the tab spans cols 1-3, the marker takes col 1, so
# quoted content begins at col 2 and the tab's remaining cols 2-3 are 2
# columns of indentation before the bullet (within CommonMark's 0-3 budget
# for indentation before a list marker); '-' sits at col 4, '- ' is 2
# columns wide, so the list item's content column is col 6 — 4 columns past
# the blockquote's own content column at col 2.
#
# The closing line repeats only the blockquote marker ('> ', content col 2)
# plus <n_spaces>; its backtick run therefore sits at absolute column
# 2 + n_spaces. A closing fence must sit inside the list item (>= col 6)
# and within the fence's own 0-3 column budget (<= col 9) — i.e. n_spaces
# 4 through 7 close it, 3 (one short of the content column) and 8 (one past
# the budget) do not. Before this fix the recorded content column was 4
# rather than 6, shifting that whole window two columns left.
bq_tab_list_closer_case() {
  local n="$1" expect="$2" label="$3" pad
  pad="$(printf '%*s' "$n" '')"
  new_fixture
  {
    printf '>\t- ```text verified:x\n'
    printf '> hello\n'
    printf '> %s```\n' "$pad"
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  [ "$RC" -eq "$expect" ] && pass "$label: closer at bq-relative column $n: exit $expect" \
    || fail "$label: closer at bq-relative column $n: expected rc=$expect, got rc=$RC out=$OUT"
}

# 139. Closers at every width from the true content column (4) to +3.
bq_tab_list_closer_case 4 0 "tab after '>' then a bullet (depth 1)"
bq_tab_list_closer_case 5 0 "tab after '>' then a bullet (depth 1)"
bq_tab_list_closer_case 6 0 "tab after '>' then a bullet (depth 1)"
bq_tab_list_closer_case 7 0 "tab after '>' then a bullet (depth 1)"
# One column short of the content column, and one column past the budget:
# neither is a close, so the fence is reported as never closed.
bq_tab_list_closer_case 3 1 "tab after '>' then a bullet (depth 1)"
bq_tab_list_closer_case 8 1 "tab after '>' then a bullet (depth 1)"

# 140. The same opening shape, untagged: the opening line must be named,
# not misread as never-closed.
new_fixture
{
  printf '>\t- ```text\n'
  printf '> hello\n'
  printf '>     ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "tab after '>' then a bullet, untagged: exit 1" \
  || fail "tab after '>' then a bullet, untagged: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
    pass "tab after '>' then a bullet, untagged: opening line reported, not 'never closed'" ;;
  *) fail "tab after '>' then a bullet, untagged: wrong/missing message: out=$OUT" ;;
esac

# 141. Depth 2, no list marker. `>` at col 0; tab cols 1-3 (marker takes
# col 1, cols 2-3 remain); the second `>` is reached across those 2 columns
# and sits at col 4, so it ends at col 5; the second tab spans cols 5-7 and
# its marker takes col 5, leaving cols 6-7. Quoted content therefore begins
# at col 6, and the following 3 spaces put the backtick run at col 11 — 5
# columns in, an indented code block, not a fence.
new_fixture
printf '>\t>\t   ```bash\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "depth-2 tabs, run 5 columns inside the quote: not a fence, exit 0" \
  || fail "depth-2 tabs, run 5 columns inside the quote: expected rc=0, got rc=$RC out=$OUT"

# 141b. The same depth-2 prefix with the backtick run AT the quoted-content
# column (col 6, relative indent 0) IS a real fence, and closes normally —
# the fix must not make tabs after `>` stop resolving altogether.
new_fixture
{
  printf '>\t> ```text verified:x\n'
  printf '>\t> echo hi\n'
  printf '>\t> ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "depth-2 tabs, run at relative indent 0: real fence, opens and closes, exit 0" \
  || fail "depth-2 tabs, run at relative indent 0: expected rc=0, got rc=$RC out=$OUT"

new_fixture
{
  printf '>\t> ```text\n'
  printf '>\t> echo hi\n'
  printf '>\t> ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
case "$OUT" in
  *"tasks.md:1: fenced code block has no verified:/unverified: tag"*) \
    pass "depth-2 tabs, run at relative indent 0, untagged: opening line reported" ;;
  *) fail "depth-2 tabs, untagged: wrong/missing message: out=$OUT" ;;
esac

# bq2_tab_list_closer_case <n_spaces> <expect_rc> <label>
#
# Depth 2, tab after each `>`, then a bullet. By the column arithmetic in
# 141, quoted content begins at col 6; the following tab spans cols 6-7 (2
# columns of indentation before the marker, within the 0-3 budget), '-' is
# at col 8 and the item's content column is col 10.
#
# The closing line repeats the same '>\t>\t' prefix, which lands the cursor
# at col 6 with the second tab's remaining cols 6-7 still to come: those 2
# columns plus <n_spaces> put the backtick run at col 8 + n_spaces. Inside
# the item (>= col 10) and within the 0-3 budget (<= col 13) means n_spaces
# 2 through 5 close it; 1 does not.
bq2_tab_list_closer_case() {
  local n="$1" expect="$2" label="$3" pad
  pad="$(printf '%*s' "$n" '')"
  new_fixture
  {
    printf '>\t>\t- ```text verified:x\n'
    printf '>\t> hello\n'
    printf '>\t>\t%s```\n' "$pad"
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  [ "$RC" -eq "$expect" ] && pass "$label: closer with $n extra spaces: exit $expect" \
    || fail "$label: closer with $n extra spaces: expected rc=$expect, got rc=$RC out=$OUT"
}

# 142. Depth-2 closers from the true content column to +3, plus one short.
bq2_tab_list_closer_case 2 0 "tabs after both '>'s then a bullet (depth 2)"
bq2_tab_list_closer_case 3 0 "tabs after both '>'s then a bullet (depth 2)"
bq2_tab_list_closer_case 4 0 "tabs after both '>'s then a bullet (depth 2)"
bq2_tab_list_closer_case 5 0 "tabs after both '>'s then a bullet (depth 2)"
bq2_tab_list_closer_case 1 1 "tabs after both '>'s then a bullet (depth 2)"

# ===========================================================================
# SECTION: Pass-13 fix wave — Important 2: the outer-indent cap (0-3
# columns before any container marker is recognised — the same CommonMark
# indented-code-block rule cited in the pass-12 section above) applied only
# to fence OPENING lines. `is_closing_line` never had it, so a closing line
# whose own `>` marker began 4+ columns in was accepted as a valid close of
# a fence opened at outer indent 0. `blockquote_indent_case` above varies
# the indent UNIFORMLY across open, content and close lines, which is why
# no row of it could ever see this. These fixtures deliberately MISMATCH
# the two sides, in both directions.
# ===========================================================================

# 143a. Shallow open (outer 0), deep close (outer 4): the closing line's
# `>` is literal text inside an indented code block, so it does not close
# the fence, which is then correctly reported as never closed.
new_fixture
{
  printf '> ```text verified:x\n'
  printf '> echo hi\n'
  printf '    > ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "open at outer indent 0, close at outer indent 4: not a close, fence reported as never closed" \
  || fail "open outer 0 / close outer 4: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"never closed"*) pass "open outer 0 / close outer 4: reported as never closed, not accepted silently" ;;
  *) fail "open outer 0 / close outer 4: wrong/missing message: out=$OUT" ;;
esac

# 143b. Shallow open (outer 0), close at outer 3 — the last column inside
# the cap. Still a valid close; the cap must not over-fire.
new_fixture
{
  printf '> ```text verified:x\n'
  printf '> echo hi\n'
  printf '   > ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "open at outer indent 0, close at outer indent 3: still a valid close, exit 0" \
  || fail "open outer 0 / close outer 3: expected rc=0, got rc=$RC out=$OUT"

# 143c. The reverse mismatch — deep open (outer 3), shallow close (outer
# 0). CommonMark budgets 0-3 columns of outer indentation on EACH line
# independently; it does not require the two to be equal. This must stay a
# valid close, so the cap is a cap and not a new equality requirement.
new_fixture
{
  printf '   > ```text verified:x\n'
  printf '   > echo hi\n'
  printf '> ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "open at outer indent 3, close at outer indent 0: still a valid close, exit 0" \
  || fail "open outer 3 / close outer 0: expected rc=0, got rc=$RC out=$OUT"

# 143d. Deep open (outer 3), deeper close (outer 4): the cap fires on the
# closing side regardless of how deep the opening line was.
new_fixture
{
  printf '   > ```text verified:x\n'
  printf '   > echo hi\n'
  printf '    > ```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "open at outer indent 3, close at outer indent 4: not a close" \
  || fail "open outer 3 / close outer 4: expected rc=1, got rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-13 fix wave — Minor 6: the abort message for a line that
# HAS a container marker but exceeds the outer-indent cap. See
# `assert_over_indented_marker_message` above for why the indent-4/5 matrix
# rows could not catch this.
# ===========================================================================

# 144. The exact shape from the fix-wave brief: a tab (4 columns from col
# 0) before a `- ` marker, then a tab before the backtick run.
new_fixture
printf -- '\t- \t```bash\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "over-indented line that DOES carry a marker: exit 4" \
  || fail "over-indented line with a marker: expected rc=4, got rc=$RC out=$OUT"
assert_over_indented_marker_message "over-indented line that DOES carry a marker"
case "$OUT" in
  *"adding an explicit list/blockquote marker on this line"*) \
    fail "over-indented line with a marker: remedy tells the operator to add a marker the line already has: out=$OUT" ;;
  *) pass "over-indented line with a marker: remedy does not ask for a marker that is already present" ;;
esac

# 145. The BARE over-indented line keeps its own (correct) message — the
# new text must not have replaced it, or one wrong shared message would
# simply have become another.
new_fixture
printf -- '\t```bash\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "bare over-indented line: exit 4" \
  || fail "bare over-indented line: expected rc=4, got rc=$RC out=$OUT"
case "$OUT" in
  *"with no container prefix on this line"*) \
    pass "bare over-indented line: keeps the bare-case message" ;;
  *) fail "bare over-indented line: bare-case message missing: out=$OUT" ;;
esac

# ===========================================================================
# SECTION: Pass-14 fix wave — Critical 1: whitespace runs bounded WHERE THEY
# ARE USED, not merely measured somewhere.
#
# Every fixture below is justified against CommonMark, not against what the
# guard printed. The two rules in play:
#
#   * List items (https://spec.commonmark.org/0.30/#list-items). Rule 1
#     ("Basic case"): a marker followed by W columns of whitespace, 1 <= W
#     <= 4, gives a content column W past the marker. Rule 2 ("Item
#     starting with indented code"): at W >= 5 the content column is
#     exactly ONE past the marker and the remaining columns are indented
#     code belonging to the item's content.
#   * Block quotes (https://spec.commonmark.org/0.30/#block-quotes): "up
#     to three spaces of initial indent" before the `>`; at 4+ columns the
#     `>` is literal text in an indented code block and the run ends.
#
# Both bounds were absent: `_LIST_RE`/`_CHECKBOX_RE` carried an unbounded
# `\s+` and `_BQ_MARKER_RE` an unbounded `\s*`, so a fence run parked past
# either bound opened a fence CommonMark reads as literal code, and
# everything it then swallowed went unreported.
# ===========================================================================

# 146-151. A bullet followed by 1..6 spaces, then a TAGGED fence, with a
# numeric claim on the next line. Rule 1 holds through W=4: the content
# column is 1+W, the fence sits exactly there (relative indent 0), the
# fence opens, and the claim is fence CONTENT — not scanned — so exit 0.
# At W=5 and W=6 rule 2 applies: the content column is 2, the backtick run
# sits 4+ columns past it and is INDENTED CODE, no fence opens, and the
# claim on the next line is ordinary prose — an unattributed numeric claim,
# exit 1. The claim-alone control (case 152) proves the claim itself is
# always reportable, so a passing W=5 case cannot be an artefact of the
# claim being unrecognised.
for w in 1 2 3 4 5 6; do
  new_fixture
  {
    printf -- '-%*s```bash verified:x\n' "$w" ''
    printf 'took 5 ms\n'
    printf -- '%*s```\n' "$((w + 1))" ''
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  if [ "$w" -le 4 ]; then
    [ "$RC" -eq 0 ] && pass "list marker + $w space(s): rule 1, fence opens, claim is content" \
      || fail "list marker + $w space(s): expected rc=0 (rule 1), got rc=$RC out=$OUT"
  else
    [ "$RC" -eq 1 ] && pass "list marker + $w space(s): rule 2, indented code, claim reported" \
      || fail "list marker + $w space(s): expected rc=1 (rule 2), got rc=$RC out=$OUT"
    case "$OUT" in
      *"tasks.md:2: numeric claim"*) pass "list marker + $w space(s): the claim is named at line 2" ;;
      *) fail "list marker + $w space(s): claim not named: out=$OUT" ;;
    esac
  fi
done

# 152. The control: the same claim with no container syntax at all is a
# violation. Without this, the W>=5 rows above could be passing because the
# claim text is not recognised rather than because it is reachable.
new_fixture
printf 'took 5 ms\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "claim-alone control: reported" \
  || fail "claim-alone control: expected rc=1, got rc=$RC out=$OUT"

# 153. `-` + one tab. The tab starts at column 1 and runs to column 4, so
# W=3 — rule 1, content column 4, fence at relative indent 0. Opens; the
# claim is content; exit 0.
new_fixture
{
  printf -- '-\t```bash verified:x\n'
  printf 'took 5 ms\n'
  printf -- '\t```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "list marker + one tab (W=3): rule 1, fence opens" \
  || fail "list marker + one tab: expected rc=0, got rc=$RC out=$OUT"

# 154. `-` + two tabs — CommonMark's Tabs section, example 7: `-\t\tfoo`
# renders as <ul><li><pre><code>  foo</code></pre></li></ul>. The tabs run
# columns 1-4 and 4-8, so W=7: rule 2, content column 2, and the remaining
# 6 columns are indented code. No fence opens; the claim is prose; exit 1.
new_fixture
{
  printf -- '-\t\t```bash verified:x\n'
  printf 'took 5 ms\n'
  printf -- '\t\t```\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "list marker + two tabs (W=7): rule 2 (spec Tabs ex. 7), claim reported" \
  || fail "list marker + two tabs: expected rc=1, got rc=$RC out=$OUT"

# 155-160. The same bound on a CHECKBOX's trailing whitespace, which this
# grammar treats as an extension of the marker. `- [ ]` ends at column 5;
# W=1..4 keeps the content column at 5+W (fence opens, claim is content),
# W>=5 drops it to 6 with the rest indented code (no fence, claim reported).
for w in 1 2 3 4 5 6; do
  new_fixture
  {
    printf -- '- [ ]%*s```bash verified:x\n' "$w" ''
    printf 'took 5 ms\n'
    printf -- '%*s```\n' "$((w + 5))" ''
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  if [ "$w" -le 4 ]; then
    [ "$RC" -eq 0 ] && pass "checkbox + $w space(s): rule 1, fence opens" \
      || fail "checkbox + $w space(s): expected rc=0, got rc=$RC out=$OUT"
  else
    [ "$RC" -eq 1 ] && pass "checkbox + $w space(s): rule 2, claim reported" \
      || fail "checkbox + $w space(s): expected rc=1, got rc=$RC out=$OUT"
  fi
done

# 161-166. Blockquote depth: `>` + N spaces + `>` + an UNTAGGED fence,
# closed at the same shape. The first `>` credits itself one column, so the
# indentation actually in front of the second `>` is N-1 columns. At N<=4
# that is 0-3 columns — a legal nested blockquote, depth 2, the fence opens
# and its missing tag is reported (exit 1). At N=5 it is 4 columns: the
# second `>` is literal text inside an indented code block, the run ends at
# depth 1, and the guard refuses to guess (exit 4) rather than recording a
# depth CommonMark does not agree with. Before this fix every N recorded
# depth 2 — the width was measured and never bounded.
for n in 0 1 2 3 4 5; do
  new_fixture
  {
    printf -- '>%*s> ```bash\n' "$n" ''
    printf 'x\n'
    printf -- '>%*s> ```\n' "$n" ''
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  if [ "$n" -le 4 ]; then
    [ "$RC" -eq 1 ] && pass "bq gap $n (<=3 cols before the nested >): depth 2, untagged fence reported" \
      || fail "bq gap $n: expected rc=1, got rc=$RC out=$OUT"
  else
    [ "$RC" -eq 4 ] && pass "bq gap $n (4+ cols): run ends at depth 1, guard refuses" \
      || fail "bq gap $n: expected rc=4, got rc=$RC out=$OUT"
  fi
done

# ===========================================================================
# SECTION: Pass-14 fix wave — Criticals 2 and 3: a fence carrying BOTH a
# `bq_pre` and a `bq_post`. The harness had `> - [ ] ``` ` (bq_post = 0) and
# `- > ``` ` (bq_pre = 0) and nothing with both, which is how twelve passes
# missed that no closing line could ever be accepted (C2) and that the one
# whitespace run before the nested `>` was uncapped (C3).
#
# The opener `> - > ```text verified:manual` establishes, per CommonMark:
# `>` at column 0 crediting one column (content at column 2); `- ` there,
# marker plus one space (list content column 2 relative to the quote, i.e.
# absolute column 4); `> ` at column 4 crediting one column (content at
# column 6); the fence run at column 6, relative indent 0.
#
# A closing line must therefore carry `>`, one column, then the list item's
# 2 content columns, then CommonMark's 0-3 columns of indentation before the
# nested `>`, then `> ```. Written as `>` + N spaces + `> ``` `, that is
# N = 1 + 2 + (0..3) = 3..6. N < 3 is indented too little to be inside the
# list item; N > 6 exceeds the 0-3 budget in front of the nested `>`.
# ===========================================================================

# 167-178. The full closer sweep, 0 through 20 columns.
for n in 0 1 2 3 4 5 6 7 8 10 15 20; do
  new_fixture
  {
    printf -- '> - > ```text verified:manual\n'
    printf 'inside\n'
    printf -- '>%*s> ```\n' "$n" ''
  } > "$FIXTURE/openspec/changes/demo-change/tasks.md"
  run_guard "$FIXTURE"
  if [ "$n" -ge 3 ] && [ "$n" -le 6 ]; then
    [ "$RC" -eq 0 ] && pass "bq_pre+bq_post closer at N=$n: legal, fence closes" \
      || fail "bq_pre+bq_post closer at N=$n: expected rc=0, got rc=$RC out=$OUT"
  else
    [ "$RC" -eq 1 ] && pass "bq_pre+bq_post closer at N=$n: illegal, fence stays open" \
      || fail "bq_pre+bq_post closer at N=$n: expected rc=1, got rc=$RC out=$OUT"
    case "$OUT" in
      *"never closed"*) pass "bq_pre+bq_post closer at N=$n: named as unclosed" ;;
      *) fail "bq_pre+bq_post closer at N=$n: wrong message: out=$OUT" ;;
    esac
  fi
done

# 179. What the unclosable fence was SWALLOWING. With a legal closer (N=3)
# an untagged fence and an unattributed claim further down the file are both
# found; before Critical 2's fix neither was, because no closer existed and
# every remaining line was read as this fence's content.
new_fixture
{
  printf -- '> - > ```text verified:manual\n'
  printf 'inside\n'
  printf -- '>   > ```\n'
  printf -- '```untagged\n'
  printf 'x\n'
  printf -- '```\n'
  printf 'took 7 ms\n'
} > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "bq_pre+bq_post: downstream findings survive the fence" \
  || fail "bq_pre+bq_post downstream: expected rc=1, got rc=$RC out=$OUT"
case "$OUT" in
  *"tasks.md:4: fenced code block has no verified"*) pass "bq_pre+bq_post: the later untagged fence is reported" ;;
  *) fail "bq_pre+bq_post: later untagged fence missing: out=$OUT" ;;
esac
case "$OUT" in
  *"tasks.md:7: numeric claim"*) pass "bq_pre+bq_post: the later claim is reported" ;;
  *) fail "bq_pre+bq_post: later claim missing: out=$OUT" ;;
esac

# 180. Critical 3's open/close symmetry, stated as an assertion rather than
# left implicit: the gap in front of the nested `>` is budgeted at the same
# 0-3 columns on BOTH sides. As an OPENER, `> - ` followed by 4+ columns and
# a `>` is refused (the guard will not guess); the closer sweep above
# refuses the mirror shape at N=7. One question, one answer.
new_fixture
printf -- '> -     > ```bash\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 4 ] && pass "bq_post gap 4+ as an OPENER: refused, same budget as the closer" \
  || fail "bq_post gap as opener: expected rc=4, got rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-14 fix wave — Critical 4: the wrapper's interpreter gate.
# This was the only channel of the exit taxonomy with no coverage at all,
# which is why "python3 exists but does not run" shipped as exit 1 —
# "violations found" — on any macOS box without the Command Line Tools.
# ===========================================================================

# 181. A python3 that exists and is executable but fails, exactly as the
# xcrun stub does. Must be exit 2 (environment), never exit 1.
new_fixture
printf 'clean\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
FAKEBIN="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-fakebin.XXXXXX")"
{
  printf '#!/bin/sh\n'
  printf 'echo "xcrun: error: invalid active developer path" >&2\n'
  printf 'exit 1\n'
} > "$FAKEBIN/python3"
chmod +x "$FAKEBIN/python3"
set +e
OUT="$(PATH="$FAKEBIN:$PATH" CHECK_PLAN_PROVENANCE_ROOT="$FIXTURE" "$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "broken python3: exit 2, not exit 1" \
  || fail "broken python3: expected rc=2, got rc=$RC out=$OUT"
case "$OUT" in
  *"failed to run a trivial program"*) pass "broken python3: names what actually happened" ;;
  *) fail "broken python3: message missing: out=$OUT" ;;
esac
case "$OUT" in
  *"invalid active developer path"*) pass "broken python3: the interpreter's own stderr is passed through" ;;
  *) fail "broken python3: interpreter stderr swallowed: out=$OUT" ;;
esac

# 182. The absent case still behaves — the probe must not have displaced the
# `command -v` gate, whose message names a different fix.
MINBIN="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-minbin.XXXXXX")"
for t in dirname pwd; do
  [ -x "/bin/$t" ] && ln -sf "/bin/$t" "$MINBIN/$t"
  [ -x "/usr/bin/$t" ] && ln -sf "/usr/bin/$t" "$MINBIN/$t"
done
set +e
OUT="$(PATH="$MINBIN" CHECK_PLAN_PROVENANCE_ROOT="$FIXTURE" /bin/bash "$GUARD" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "absent python3: still exit 2" \
  || fail "absent python3: expected rc=2, got rc=$RC out=$OUT"
case "$OUT" in
  *"not found on PATH"*) pass "absent python3: keeps its own distinct message" ;;
  *) fail "absent python3: message missing: out=$OUT" ;;
esac

# 183. A working python3 is unaffected — the probe adds a gate, not a
# behaviour change.
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "working python3: unchanged, exit 0" \
  || fail "working python3: expected rc=0, got rc=$RC out=$OUT"

# ===========================================================================
# SECTION: Pass-14 fix wave — Critical 5: `openspec/changes` may itself be a
# symlink. Containment anchored every per-file decision on
# realpath(changes_dir) and never validated that anchor, while `main()`
# reached it through `os.path.isdir`, which follows symlinks. The guard
# refused the strictly weaker escape (a symlinked change directory one level
# deeper) and followed the stronger one.
# ===========================================================================

# 184. `changes/` -> an empty decoy inside the repo, with the real plan
# elsewhere in the tree. This reported exit 0, "nothing in flight" — the
# false clean run the module docstring calls the one impermissible outcome.
FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-test.XXXXXX")"
mkdir -p "$FIXTURE/openspec" "$FIXTURE/decoy" "$FIXTURE/real/plan-change"
{
  printf -- '```untagged\n'
  printf 'x\n'
  printf -- '```\n'
} > "$FIXTURE/real/plan-change/tasks.md"
ln -s "$FIXTURE/decoy" "$FIXTURE/openspec/changes"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "changes/ symlinked to a decoy: exit 3, not a clean run" \
  || fail "changes/ decoy symlink: expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"openspec/changes is a symlink"*) pass "changes/ decoy symlink: names the anchor" ;;
  *) fail "changes/ decoy symlink: message missing: out=$OUT" ;;
esac

# 185. `changes/` -> a directory OUTSIDE the repo root. This read out of
# tree and reported exit 1 with repo-relative paths that concealed it.
FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-test.XXXXXX")"
OUTSIDE="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-outside.XXXXXX")"
mkdir -p "$FIXTURE/openspec" "$OUTSIDE/evil"
{
  printf -- '```untagged\n'
  printf 'x\n'
  printf -- '```\n'
} > "$OUTSIDE/evil/tasks.md"
ln -s "$OUTSIDE" "$FIXTURE/openspec/changes"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "changes/ symlinked outside the root: exit 3" \
  || fail "changes/ outside symlink: expected rc=3, got rc=$RC out=$OUT"

# 186. The one-level-deeper control, which was ALREADY refused — it must
# stay refused, and by its own (different) message, so the new anchor check
# has not simply subsumed the per-file one.
FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-test.XXXXXX")"
OUTSIDE="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-outside.XXXXXX")"
mkdir -p "$FIXTURE/openspec/changes" "$OUTSIDE/evil"
{
  printf -- '```untagged\n'
  printf 'x\n'
  printf -- '```\n'
} > "$OUTSIDE/evil/tasks.md"
ln -s "$OUTSIDE/evil" "$FIXTURE/openspec/changes/evil"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "changes/<name> symlink control: still exit 3" \
  || fail "changes/<name> symlink control: expected rc=3, got rc=$RC out=$OUT"
case "$OUT" in
  *"refusing to scan a path reached through a symlink escape"*) \
    pass "changes/<name> symlink control: keeps its own per-file message" ;;
  *) fail "changes/<name> symlink control: per-file message missing: out=$OUT" ;;
esac

# 187. `openspec/` itself symlinked, with a real `changes` inside it — the
# realpath half of the anchor check rather than the islink half.
FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-test.XXXXXX")"
OUTSIDE="$(mktemp -d "${TMPDIR:-/tmp}/plan-prov-outside.XXXXXX")"
mkdir -p "$OUTSIDE/changes/demo-change"
printf 'clean\n' > "$OUTSIDE/changes/demo-change/tasks.md"
ln -s "$OUTSIDE" "$FIXTURE/openspec"
run_guard "$FIXTURE"
[ "$RC" -eq 3 ] && pass "openspec/ itself symlinked: exit 3 via the realpath check" \
  || fail "openspec/ symlink: expected rc=3, got rc=$RC out=$OUT"

# 188. An ordinary, non-symlinked tree is untouched by the anchor check.
new_fixture
printf 'clean\n' > "$FIXTURE/openspec/changes/demo-change/tasks.md"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "real changes/ directory: anchor check does not fire" \
  || fail "real changes/ directory: expected rc=0, got rc=$RC out=$OUT"

if [ "$FAILURES" -ne 0 ]; then
  printf '\n%d assertion(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf '\nAll check-plan-provenance assertions passed\n'
