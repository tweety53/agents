#!/usr/bin/env bash
# Assertion harness for check-stage-mark-calls.sh. Builds fixture files under
# a sandboxed mktemp directory and asserts the guard's exit status and, where
# the case names one, the presence of the expected finding text. Never
# touches the real repository tree.
#
# Modeled on test-check-vocabulary.sh's fixture-driven pattern: fixtures live
# under mktemp -d, the guard is invoked via a thin run_guard helper that
# captures RC/OUT, and every case ends with an explicit pass/fail assertion.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-stage-mark-calls.sh"
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
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/check-stage-mark-calls-test.XXXXXX")"
  FIXTURE_FILE="$FIXTURE/SKILL.md"
}

# ===========================================================================
# Case 1: a fully compliant single-line call -> exit 0.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness <harness> -session-token mf-abc123 <name>
myflow stage end   -command '/myflow-do' -stage do.review-panel -outcome completed <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 1: compliant single-line call passes" || fail "case 1: rc=$RC out=$OUT"

# ===========================================================================
# Case 2: -session-token omitted entirely -> caught, -harness present.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness <harness> <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 2: missing -session-token is caught" || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"carries no -session-token"*) pass "case 2: finding names the missing flag" ;;
  *) fail "case 2: expected a -session-token finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 3: -harness omitted entirely -> caught, -session-token present.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -session-token mf-abc123 <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 3: missing -harness is caught" || fail "case 3: rc=$RC out=$OUT"
case "$OUT" in
  *"carries no -harness"*) pass "case 3: finding names the missing flag" ;;
  *) fail "case 3: expected a -harness finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 4: both flags omitted -> two findings, one call.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage 'the review panel' <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 4: both flags missing is caught" || fail "case 4: rc=$RC out=$OUT"
case "$OUT" in
  *"carries no -session-token"*"carries no -harness"*) pass "case 4: both findings reported" ;;
  *) fail "case 4: expected both findings, out=$OUT" ;;
esac

# ===========================================================================
# Case 5: session token is a command substitution "$(...)" -> caught.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness <harness> -session-token "mf-$(date +%s)-$$" <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 5: \$(...) substitution is caught" || fail "case 5: rc=$RC out=$OUT"
case "$OUT" in
  *'command substitution'*) pass "case 5: finding names the substitution shape" ;;
  *) fail "case 5: expected a substitution finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 6: session token contains a backtick -> caught.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness <harness> -session-token "mf-`date +%s`" <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 6: backtick substitution is caught" || fail "case 6: rc=$RC out=$OUT"
case "$OUT" in
  *'backtick'*) pass "case 6: finding names the backtick" ;;
  *) fail "case 6: expected a backtick finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 7: session token contains a shell variable reference ($VAR) -> caught.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness <harness> -session-token "mf-$RANDOM" <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 7: \$VAR substitution is caught" || fail "case 7: rc=$RC out=$OUT"
case "$OUT" in
  *'shell variable reference'*) pass "case 7: finding names the shell-variable shape" ;;
  *) fail "case 7: expected a shell-variable finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 7b: harness is a hardcoded literal ("claude-code") -> caught. This
# skill source installs into three harnesses from one file, so a fixed value
# mislabels two of them -- the guard must reject it the same way it rejects a
# substituted session token.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness claude-code -session-token mf-abc123 <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 7b: hardcoded harness literal is caught" || fail "case 7b: rc=$RC out=$OUT"
case "$OUT" in
  *'hardcoded literal'*) pass "case 7b: finding names the hardcoded literal" ;;
  *) fail "case 7b: expected a hardcoded-literal finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 7c: harness is a hardcoded literal other than "claude-code" (e.g.
# "cursor") -> still caught. The rule is "not a placeholder", not a specific
# forbidden word.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness cursor -session-token mf-abc123 <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 7c: any hardcoded harness literal is caught, not just claude-code" || fail "case 7c: rc=$RC out=$OUT"

# ===========================================================================
# Case 8: a compliant call split across continuation lines -> exit 0. Proves
# the guard reads flags written on a later physical line of the same call.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-fast' \
  -stage do.review-panel \
  -harness <harness> \
  -session-token mf-continued-001 <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 8: compliant multi-line call passes" || fail "case 8: rc=$RC out=$OUT"

# ===========================================================================
# Case 9: `stage end` carries neither flag by design and must never be
# flagged -- only `stage begin` is checked.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness <harness> -session-token mf-abc123 <name>
myflow stage end   -command '/myflow-do' -stage do.review-panel -outcome completed <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 9: stage end is never checked" || fail "case 9: rc=$RC out=$OUT"

# ===========================================================================
# Case 10 (KAN-197 coverage): a file with no stage marks at all -> zero calls
# checked FOR THAT FILE, undeclared, is now a coverage violation naming it —
# the KAN-73-shaped case: a rule computing nothing to check for one corpus
# member, on an otherwise clean tree. Before per-file coverage this exited 0;
# that was exactly the gap this task closes, so the assertion below is
# deliberately the opposite of what an earlier version of this file checked.
# ===========================================================================
new_fixture
printf 'Ordinary prose with no myflow stage marks in it.\n' >"$FIXTURE_FILE"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 10: a file with no stage marks at all is an undeclared coverage violation" \
  || fail "case 10: rc=$RC out=$OUT"
case "$OUT" in
  *"$FIXTURE_FILE"*"0 checked"*"not declared expected-zero"*) \
    pass "case 10: names the file, the zero count, and that it is undeclared" ;;
  *) fail "case 10: expected the file named as an undeclared zero, out=$OUT" ;;
esac

# ===========================================================================
# Case 11: change argument is `<name-or-best-guess>` -> caught, finding names
# the offending argument. A mark writes, so a guessed name bootstraps a
# change row that outlives the run (kan-182).
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-fast' -stage do.state-gate -harness <harness> -session-token mf-abc123 <name-or-best-guess>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 11: guessed change name is caught" || fail "case 11: rc=$RC out=$OUT"
case "$OUT" in
  *'<name-or-best-guess>'*) pass "case 11: finding names the offending argument" ;;
  *) fail "case 11: expected the argument named in the finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 12: the same shape, written across continuation lines -> caught,
# proving the check reads the assembled command and not the physical line.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-fast' \
  -stage do.state-gate \
  -harness <harness> \
  -session-token mf-abc123 <name-or-best-guess>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 12: guessed change name is caught across continuation lines" || fail "case 12: rc=$RC out=$OUT"
case "$OUT" in
  *'<name-or-best-guess>'*) pass "case 12: finding names the offending argument" ;;
  *) fail "case 12: expected the argument named in the finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 13: change argument is the compliant `<name>` -> exit 0. The check
# must not fire on the form every other skill uses.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-fast' -stage do.state-gate -harness <harness> -session-token mf-abc123 <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 13: compliant <name> argument passes" || fail "case 13: rc=$RC out=$OUT"

# ===========================================================================
# Case 14: `stage end ... <name-or-best-guess>` -> `stage end` is deliberately
# not examined at all, so no guess-placeholder finding is ever reported for
# it. This fixture carries no `stage begin` call whatsoever, so — since
# KAN-197 — it is now ALSO an undeclared coverage violation (case 10's
# shape); the assertion below is scoped to what this case actually tests
# (the guess-check exemption), not to overall exit status.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage end -command '/myflow-fast' -stage do.state-gate -outcome completed <name-or-best-guess>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 14: no stage begin call at all is an undeclared coverage violation" \
  || fail "case 14: rc=$RC out=$OUT"
case "$OUT" in
  *'names a guess'*) fail "case 14: stage end was wrongly checked for a guessed name: out=$OUT" ;;
  *) pass "case 14: stage end with a guessed name is never checked" ;;
esac

# ===========================================================================
# Case 15: change argument is `<name-or-best-guess>` followed by trailing
# whitespace -> still caught. Kept as a regression case: the guard now
# matches the placeholder anywhere in the assembled text, so trailing
# whitespace after it is irrelevant, but a prior extraction-based version
# of this guard mishandled it by splitting off an empty last token.
# ===========================================================================
new_fixture
printf '```bash\nmyflow stage begin -command '"'"'/myflow-fast'"'"' -stage do.state-gate -harness <harness> -session-token mf-abc123 <name-or-best-guess> \n```\n' >"$FIXTURE_FILE"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 15: guessed change name with trailing whitespace is caught" || fail "case 15: rc=$RC out=$OUT"
case "$OUT" in
  *'<name-or-best-guess>'*) pass "case 15: finding names the offending argument" ;;
  *) fail "case 15: expected the argument named in the finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 16: change argument is a QUOTED guess placeholder, single-quoted ->
# still caught. Kept as a regression case (kan-182 panel finding F1): the
# guard now searches the whole assembled command text for the placeholder,
# so a surrounding quote pair is just more text around the match and cannot
# hide it, but a prior extraction-based version required stripping quotes
# from the extracted token explicitly.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-fast' -stage do.state-gate -harness <harness> -session-token mf-abc123 '<name-or-best-guess>'
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 16: single-quoted guessed change name is caught" || fail "case 16: rc=$RC out=$OUT"
case "$OUT" in
  *'<name-or-best-guess>'*) pass "case 16: finding names the offending argument" ;;
  *) fail "case 16: expected the argument named in the finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 17: same shape, double-quoted -> still caught.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-fast' -stage do.state-gate -harness <harness> -session-token mf-abc123 "<name-or-best-guess>"
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 17: double-quoted guessed change name is caught" || fail "case 17: rc=$RC out=$OUT"
case "$OUT" in
  *'<name-or-best-guess>'*) pass "case 17: finding names the offending argument" ;;
  *) fail "case 17: expected the argument named in the finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 18: change argument is followed by a trailing inline shell comment ->
# still caught. Kept as a regression case (kan-182 panel finding F2): the
# guard searches the whole assembled text, including any trailing comment,
# for the placeholder -- a `stage begin` line that mentions
# `<name-or-best-guess>` anywhere, comment or not, is a call site to fix. A
# prior extraction-based version of this guard instead had to strip a
# trailing comment first or the comment's last word became "the" token.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-fast' -stage do.state-gate -harness <harness> -session-token mf-abc123 <name-or-best-guess>  # resolve later
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 18: guessed change name behind a trailing comment is caught" || fail "case 18: rc=$RC out=$OUT"
case "$OUT" in
  *'<name-or-best-guess>'*) pass "case 18: finding names the offending argument" ;;
  *) fail "case 18: expected the argument named in the finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 19: a `#` INSIDE a quoted -session-token value must not cause a
# compliant call (no guess placeholder anywhere) to be misread -- the call
# must still pass. The guard no longer does any comment-stripping at all,
# so this case is now a plain regression check that a stray `#` in an
# unrelated flag value never produces a false positive.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-fast' -stage do.state-gate -harness <harness> -session-token "mf-#abc123" <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 19: a # inside a quoted value is not treated as a comment" || fail "case 19: rc=$RC out=$OUT"

# ===========================================================================
# Case 20: change argument is a quoted guess placeholder with an INTERNAL
# space -> still caught. The old space-splitting last-token extraction
# truncated a quoted value at the internal space and never saw the
# placeholder (kan-182 panel finding F6).
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-fast' -stage do.state-gate -harness <harness> -session-token mf-abc123 '<name-or-best guess>'
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 20: quoted guess placeholder with an internal space is caught" || fail "case 20: rc=$RC out=$OUT"
case "$OUT" in
  *'<name-or-best guess>'*) pass "case 20: finding names the offending argument" ;;
  *) fail "case 20: expected the argument named in the finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 21: a -session-token value carries an escaped literal double-quote
# followed later by a `#` -> the guessed change argument at the end must
# still be caught. The old awk quote-tracker closed its quote on the escaped
# `\"`, desynced, and then treated the `#` as an unquoted comment start,
# discarding the rest of the line -- including the guessed argument
# (kan-182 panel finding F7).
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-fast' -stage do.state-gate -harness <harness> -session-token "mf-\"abc#tok" <name-or-best-guess>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 21: guessed change name past an escaped-quote-plus-# session token is caught" || fail "case 21: rc=$RC out=$OUT"
case "$OUT" in
  *'<name-or-best-guess>'*) pass "case 21: finding names the offending argument" ;;
  *) fail "case 21: expected the argument named in the finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 22: the change argument is separated from the rest of the call by a
# literal TAB, not a space -> still caught. The old extraction split on a
# literal space only, so a tab-separated tail was never recognized as the
# last token (kan-182 panel finding F8).
# ===========================================================================
new_fixture
printf 'myflow stage begin -command '"'"'/myflow-fast'"'"' -stage do.state-gate -harness <harness> -session-token mf-abc123\t<name-or-best-guess>\n' >"$FIXTURE_FILE"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 22: tab-separated guessed change name is caught" || fail "case 22: rc=$RC out=$OUT"
case "$OUT" in
  *'<name-or-best-guess>'*) pass "case 22: finding names the offending argument" ;;
  *) fail "case 22: expected the argument named in the finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 23 (KAN-197 coverage): a file with one compliant call reports its
# count, and the verdict's own breakdown line names it.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness <harness> -session-token mf-abc123 <name>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 23: a file with one compliant call exits 0" || fail "case 23: rc=$RC out=$OUT"
case "$OUT" in
  *"$FIXTURE_FILE 1"*) pass "case 23: the breakdown names the file with its checked count" ;;
  *) fail "case 23: expected the breakdown to show '$FIXTURE_FILE 1', out=$OUT" ;;
esac

# ===========================================================================
# Case 24 (KAN-197 coverage): a member declared expected-zero in the guard's
# own source reports its zero without failing. This guard has no
# CHECK_*_ROOT override (unlike check-references.sh / check-guard-symlinks.sh),
# so a sandboxed fixture path can never coincide with one of the guard's own
# repo-relative declared names — proving "declared passes" instead requires
# running the guard bare, against the real repository it ships in, which
# tasks 1-4 already require to be clean. skills/myflow-status/SKILL.md is
# declared with its own by-contract reason (task 5's own note); a second,
# ordinary declared member (skills/myflow-contracts/SKILL.md — the guard's
# own default corpus post-F1-narrowing is only SKILL.md/pipeline.md files, so
# this is the smallest remaining ordinary example, not skills/README.md,
# which the narrowed corpus no longer reaches at all) is checked alongside it
# so this case covers both the special-reason and the generic-reason shape.
# ===========================================================================
set +e
REAL_OUT="$("$GUARD" 2>&1)"
REAL_RC=$?
set -e
[ "$REAL_RC" -eq 0 ] && pass "case 24: the real repository's own tree is clean" \
  || fail "case 24: rc=$REAL_RC out=$REAL_OUT"
case "$REAL_OUT" in
  *"skills/myflow-status/SKILL.md 0 (declared: read-only status report"*) \
    pass "case 24: myflow-status's declared zero carries its by-contract reason" ;;
  *) fail "case 24: expected myflow-status's declared-zero reason in the breakdown, out=$REAL_OUT" ;;
esac
case "$REAL_OUT" in
  *"skills/myflow-contracts/SKILL.md 0 (declared:"*) \
    pass "case 24: an ordinary declared-zero member also reports as declared" ;;
  *) fail "case 24: expected skills/myflow-contracts/SKILL.md reported as a declared zero, out=$REAL_OUT" ;;
esac

# ===========================================================================
# Case 25 (KAN-197 F7): a directory target that enumerates to ZERO candidate
# files must not vanish from coverage — before this fix, coverage_record was
# never called at all for such a target, so it was invisible even though the
# run went on to report "clean". Reproduces the panel's own repro: a fixture
# directory holding no SKILL.md/pipeline.md file whatsoever.
# ===========================================================================
EMPTY_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-stage-mark-calls-empty.XXXXXX")"
run_guard "$EMPTY_DIR"
[ "$RC" -eq 1 ] && pass "case 25: an empty target is an undeclared coverage violation, not a vanishing pass" \
  || fail "case 25: rc=$RC out=$OUT"
case "$OUT" in
  *"$EMPTY_DIR"*"0 checked"*"not declared expected-zero"*) \
    pass "case 25: the violation names the empty target itself" ;;
  *) fail "case 25: expected the empty target named as an undeclared zero, out=$OUT" ;;
esac

# ===========================================================================
# Case 26 (KAN-197 F1 mutation): the narrowed corpus (SKILL.md/pipeline.md
# only) still detects a SKILL.md that lost all its marks — narrowing must not
# have cost detection — AND a co-located file that could never carry a mark
# (any other .md basename) is correctly excluded from the corpus rather than
# forcing a declaration for it.
# ===========================================================================
new_fixture
printf 'Ordinary prose with no myflow stage marks in it.\n' >"$FIXTURE_FILE"
printf 'This file is not named SKILL.md or pipeline.md and can never carry a mark.\n' \
  >"$FIXTURE/notes.md"
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 26: a SKILL.md stripped of its marks still fires after F1's narrowing" \
  || fail "case 26: rc=$RC out=$OUT"
case "$OUT" in
  *"notes.md"*) fail "case 26: a non-SKILL.md/pipeline.md file was wrongly pulled into the corpus, out=$OUT" ;;
  *) pass "case 26: a file that can never carry a mark is excluded from the corpus entirely" ;;
esac

# ===========================================================================
# Case 27 (KAN-197 F8): a rejected coverage_record must not be swallowed. The
# same file passed as two separate targets makes the SECOND coverage_record
# call for it fail (already recorded); before this fix the guard ran under
# `set -uo pipefail` with no `-e` and never read that return, so it printed a
# "clean" verdict whose own breakdown contradicted itself. Now it refuses to
# answer (exit 2) instead.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness <harness> -session-token mf-abc123 <name>
```
EOF
set +e
OUT="$("$GUARD" "$FIXTURE_FILE" "$FIXTURE_FILE" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 27: a rejected coverage_record aborts loudly instead of reporting a contradictory clean verdict" \
  || fail "case 27: rc=$RC out=$OUT"
case "$OUT" in
  *"already recorded"*) pass "case 27: the failure names the real cause" ;;
  *) fail "case 27: expected coverage_record's own already-recorded message, out=$OUT" ;;
esac

# ===========================================================================
# Case 28 (KAN-258): a `myflow record dispatch` call site carrying a session
# token that is a command substitution is rejected and named. `record
# dispatch` takes `-session-token` for exactly the reason `stage begin` does
# — the daemon finds the literal token in the calling session's own
# transcript — so a substitution here is the same defect, recorded
# identically by every caller and discriminating between no two sessions.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow record dispatch -change <name> -task 3 -role implementer -model opus \
  -commit abc1234 -outcome completed -session-token "mf-$(date +%s)" -started-at <ts>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 28: a substituted session token on \`record dispatch\` is caught" \
  || fail "case 28: rc=$RC out=$OUT"
case "$OUT" in
  *'command substitution'*) pass "case 28: the finding names the substitution shape" ;;
  *) fail "case 28: expected a substitution finding, out=$OUT" ;;
esac

# ===========================================================================
# Case 29 (KAN-258): the same call site carrying a literal token passes, and
# is counted — a `record dispatch` call is a checked call site in its own
# right, so a file carrying one and no `stage begin` is a covered member
# rather than an undeclared zero.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow record dispatch -change <name> -task 3 -role implementer -model opus \
  -commit abc1234 -outcome completed -session-token mf-abc123 -started-at <ts>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 29: a literal session token on \`record dispatch\` passes" \
  || fail "case 29: rc=$RC out=$OUT"
case "$OUT" in
  *"$FIXTURE_FILE 1"*) pass "case 29: the breakdown counts the \`record dispatch\` call site" ;;
  *) fail "case 29: expected the breakdown to show '$FIXTURE_FILE 1', out=$OUT" ;;
esac

# ===========================================================================
# Case 30 (KAN-258): `-session-token` omitted entirely from a `record
# dispatch` call site is caught. The flag is required by `myflow record
# dispatch`'s own usage block, and a call site that omits it writes a
# dispatch row no session can ever be bound to.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow record dispatch -change <name> -role reviewer -slot Primary -model sonnet \
  -outcome completed -started-at <ts>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 1 ] && pass "case 30: a \`record dispatch\` with no -session-token is caught" \
  || fail "case 30: rc=$RC out=$OUT"
case "$OUT" in
  *"carries no -session-token"*) pass "case 30: the finding names the missing flag" ;;
  *) fail "case 30: expected a -session-token finding, out=$OUT" ;;
esac

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
