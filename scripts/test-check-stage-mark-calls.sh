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

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
