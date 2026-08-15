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
# Case 10: a file with no stage marks at all -> exit 0, zero calls checked.
# ===========================================================================
new_fixture
printf 'Ordinary prose with no myflow stage marks in it.\n' >"$FIXTURE_FILE"
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 10: file with no marks passes" || fail "case 10: rc=$RC out=$OUT"

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
# Case 14: `stage end ... <name-or-best-guess>` -> exit 0. `stage end` is
# deliberately not examined at all.
# ===========================================================================
new_fixture
cat >"$FIXTURE_FILE" <<'EOF'
```bash
myflow stage end -command '/myflow-fast' -stage do.state-gate -outcome completed <name-or-best-guess>
```
EOF
run_guard "$FIXTURE"
[ "$RC" -eq 0 ] && pass "case 14: stage end with a guessed name is never checked" || fail "case 14: rc=$RC out=$OUT"

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

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
