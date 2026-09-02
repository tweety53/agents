#!/usr/bin/env bash
# Assertion harness for project-get.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the script at each,
# invoking the REAL scripts/project-get.sh as a subprocess against REAL
# files on disk — never a copy of its logic, and never a hand-written
# expected-output string asserted without having run the script to produce
# it.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="$SCRIPT_DIR/project-get.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do
    rm -rf "$d"
  done
}
trap cleanup EXIT

# new_root -> sets ROOT to a fresh sandbox project root carrying .flow/.
new_root() {
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/project-get-test.XXXXXX")"
  DIRS+=("$ROOT")
  mkdir -p "$ROOT/.flow"
  CFG="$ROOT/.flow/project.md"
}

write_cfg() {
  printf '%s\n' "$1" > "$CFG"
}

# write_cfg_bom — identical to write_cfg, but a UTF-8 BOM (EF BB BF) is
# prepended as the file's literal first three bytes, exactly as a real file
# some editors and Windows tooling emit by default.
write_cfg_bom() {
  printf '\xef\xbb\xbf%s\n' "$1" > "$CFG"
}

# run_get <arg...> -> sets RC and OUT to the real script's exit code and
# combined stdout+stderr.
run_get() {
  set +e
  OUT="$("$BIN" "$@" 2>&1)"
  RC=$?
  set -e
}

assert_rc() {
  local case_name="$1" want="$2"
  [ "$RC" -eq "$want" ] && pass "$case_name: exit $want" \
    || fail "$case_name: expected exit $want, got rc=$RC out=$OUT"
}

assert_out_contains() {
  local case_name="$1" needle="$2"
  case "$OUT" in
    *"$needle"*) pass "$case_name: output names '$needle'" ;;
    *) fail "$case_name: expected output to contain '$needle', got: $OUT" ;;
  esac
}

# ===========================================================================
# Case 1: usage with one argument — exit 2.
# ===========================================================================
run_get "/tmp"
assert_rc "case 1" 2

# ===========================================================================
# Case 2: root not a directory — exit 2.
# ===========================================================================
run_get "/nonexistent/$$-$RANDOM" lint
assert_rc "case 2" 2

# ===========================================================================
# Case 3: no .flow/project.md — exit 1, stderr names the path.
# ===========================================================================
new_root
rm -f "$CFG"
run_get "$ROOT" lint
assert_rc "case 3" 1
assert_out_contains "case 3" "$CFG"

# ===========================================================================
# Case 4: key absent — exit 1, stderr names the key.
# ===========================================================================
new_root
write_cfg "## test

\`go test ./...\`"
run_get "$ROOT" lint
assert_rc "case 4" 1
assert_out_contains "case 4" "lint"

# ===========================================================================
# Case 5: key present — body printed verbatim, including a fenced block and
# a "### " subheading, trailing blank lines removed. Asserted byte for byte
# against a here-doc.
# ===========================================================================
new_root
write_cfg "## lint

\`\`\`bash
scripts/check-vocabulary.sh
\`\`\`

### note

see above


## test

\`go test ./...\`"
run_get "$ROOT" lint
assert_rc "case 5" 0
EXPECTED_CASE5="$(cat <<'EOF'
```bash
scripts/check-vocabulary.sh
```

### note

see above
EOF
)"
if [ "$OUT" = "$EXPECTED_CASE5" ]; then
  pass "case 5: body matches byte for byte"
else
  fail "case 5: body mismatch: got: $OUT"
fi

# ===========================================================================
# Case 6: key declared twice — exit 2, ambiguous, never picking a winner.
# ===========================================================================
new_root
write_cfg "## lint

first

## lint

second"
run_get "$ROOT" lint
assert_rc "case 6" 2

# ===========================================================================
# Case 7: a UTF-8 BOM before the first heading does not hide a "## <key>" on
# line 1.
# ===========================================================================
new_root
write_cfg_bom "## lint

scripts/check-vocabulary.sh"
run_get "$ROOT" lint
assert_rc "case 7" 0
assert_out_contains "case 7" "scripts/check-vocabulary.sh"

# ===========================================================================
# Case 8: a key whose body is empty prints nothing and exits 0.
# ===========================================================================
new_root
write_cfg "## lint

## test

\`go test ./...\`"
run_get "$ROOT" lint
assert_rc "case 8" 0
if [ -z "$OUT" ]; then
  pass "case 8: empty body prints nothing"
else
  fail "case 8: expected empty output, got: $OUT"
fi

# ===========================================================================
# Case 9: a multi-word key passed as one quoted argument resolves.
# ===========================================================================
new_root
write_cfg "## default landing route

open PR"
run_get "$ROOT" "default landing route"
assert_rc "case 9" 0
assert_out_contains "case 9" "open PR"

# ===========================================================================
# Case 10: trailing blank lines are actually trimmed by project_section
# itself, not merely appearing trimmed because command substitution (as
# used by run_get/cases 1-9 above) strips all trailing newlines from
# captured output on its own — which would hide a body whose only trailing
# content is blank lines. Captured via real file redirection instead, then
# compared by byte count (wc -c), so a body ending in untrimmed blank lines
# ("body\n\n\n", 7 bytes) reads different from a trimmed one ("body\n",
# 5 bytes) even though $(...) would collapse both to "body".
# ===========================================================================
new_root
write_cfg "## lint

body


## test

x"
OUT_FILE="$ROOT/case10-out.txt"
set +e
"$BIN" "$ROOT" lint > "$OUT_FILE" 2>&1
RC=$?
set -e
assert_rc "case 10" 0
GOT_BYTES="$(wc -c < "$OUT_FILE" | tr -d ' ')"
if [ "$GOT_BYTES" = "5" ]; then
  pass "case 10: trailing blank lines trimmed (5 bytes: 'body\\n')"
else
  fail "case 10: expected 5 bytes ('body\\n'), got $GOT_BYTES bytes"
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
