#!/usr/bin/env bash
# Assertion harness for check-visual-trigger.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the guard at each,
# invoking the REAL scripts/check-visual-trigger.sh as a subprocess against
# REAL files on disk — never a copy of its logic, and never a hand-written
# expected-output string asserted without having run the guard to produce it.
#
# The changed-paths input is fed on stdin, one path per line, exactly as
# design.md's stage step 2 hands the guard `git diff --name-only` output.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-visual-trigger.sh"
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
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-visual-trigger-test.XXXXXX")"
  DIRS+=("$ROOT")
  mkdir -p "$ROOT/.flow"
  CFG="$ROOT/.flow/project.md"
}

write_cfg() {
  printf '%s\n' "$1" > "$CFG"
}

# write_cfg_bom — identical to write_cfg, but a UTF-8 BOM (EF BB BF) is
# prepended as the file's literal first three bytes, exactly as
# test-check-visual-verification.sh's own copy of this helper does, and
# exactly as a BOM sits in a real file some editors and Windows tooling
# emit by default.
write_cfg_bom() {
  printf '\xef\xbb\xbf%s\n' "$1" > "$CFG"
}

# run_guard <changed-path...> -> sets RC and OUT, feeding each argument as a
# line on stdin to the real guard against $ROOT. No arguments means empty
# stdin (a diff that touched nothing).
run_guard() {
  set +e
  if [ "$#" -eq 0 ]; then
    OUT="$(: | "$GUARD" "$ROOT" 2>&1)"
  else
    OUT="$(printf '%s\n' "$@" | "$GUARD" "$ROOT" 2>&1)"
  fi
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

assert_out_not_contains() {
  local case_name="$1" needle="$2"
  case "$OUT" in
    *"$needle"*) fail "$case_name: expected output NOT to contain '$needle', got: $OUT" ;;
    *) pass "$case_name: output correctly omits '$needle'" ;;
  esac
}

MINIMAL_HEADER='## visual verification

| Setting | Value |
|---------|-------|
| `ui paths` | `stats/web/src/**` |
| `screenshots` | `stats/web/tests/visual` |

| Command | Runs |
|---------|------|
| `verify` | `npm run test:visual` |
| `capture` | `npx playwright test <spec>` |'

# ===========================================================================
# Case 1: no .flow/project.md at all — "not configured" is exit 2, never a
# silent no-match.
# ===========================================================================
new_root
rm -f "$CFG"
run_guard "stats/web/src/App.tsx"
assert_rc "case 1" 2

# ===========================================================================
# Case 2: .flow/project.md exists but declares no `## visual verification`
# section — exit 2, same "cannot answer" as case 1.
# ===========================================================================
new_root
write_cfg "# Project

## run

echo hi
"
run_guard "stats/web/src/App.tsx"
assert_rc "case 2" 2

# ===========================================================================
# Case 3: section declared, `ui paths` absent — exit 2.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/src/App.tsx"
assert_rc "case 3" 2

# ===========================================================================
# Case 4: section declared, `ui paths` present but empty — exit 2.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/src/App.tsx"
assert_rc "case 4" 2

# ===========================================================================
# Case 5: a matching changed path — exit 0.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "stats/web/src/App.tsx"
assert_rc "case 5" 0
assert_out_contains "case 5" "MATCH"

# ===========================================================================
# Case 6: only non-matching changed paths — exit 1, "configured and
# unmatched" is a different answer than case 1/2's "not configured".
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "README.md" "stats/internal/api/handler.go"
assert_rc "case 6" 1

# ===========================================================================
# Case 7: empty stdin (no changed paths at all) — exit 1, trivially no match.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard
assert_rc "case 7" 1

# ===========================================================================
# Case 8 (glob semantics): `**` spans MULTIPLE directory levels, not one
# segment.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "stats/web/src/components/dashboard/Panel.tsx"
assert_rc "case 8" 0

# ===========================================================================
# Case 9 (glob semantics): a leading `./` on the DECLARED glob is normalized
# away.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`./stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/src/App.tsx"
assert_rc "case 9" 0

# ===========================================================================
# Case 10 (glob semantics): a leading `./` on the CHANGED path is normalized
# away too.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "./stats/web/src/App.tsx"
assert_rc "case 10" 0

# ===========================================================================
# Case 11 (glob semantics): an ABSOLUTE declared glob (leading `/`) is read as
# repository-relative, the same as its bare form.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`/stats/web/src/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/src/App.tsx"
assert_rc "case 11" 0

# ===========================================================================
# Case 12 (glob semantics): an ABSOLUTE changed path never matches a relative
# glob — `git diff --name-only` never produces one, so this is a defined
# refusal, not a silent grant.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "/stats/web/src/App.tsx"
assert_rc "case 12" 1

# ===========================================================================
# Case 13 (glob semantics): a changed path outside every declared glob —
# ordinary no-match, not a special case, but pinned per tasks.md.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER"
run_guard "README.md"
assert_rc "case 13" 1

# ===========================================================================
# Case 14 (glob semantics): a glob containing a SPACE is not split on it —
# only the comma is the separator.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/my folder/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "stats/web/my folder/App.tsx"
assert_rc "case 14" 0

# ===========================================================================
# Case 15: two comma-separated globs, only the second one matches.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`gymie-frontend/**, gymie-admin-frontend/**\` |
| \`screenshots\` | \`stats/web/tests/visual\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "gymie-admin-frontend/src/App.tsx"
assert_rc "case 15" 0

# ===========================================================================
# Case 16: the project root does not exist — exit 2.
# ===========================================================================
ROOT="/nonexistent/$$-$RANDOM"
run_guard "stats/web/src/App.tsx"
assert_rc "case 16" 2

# ===========================================================================
# Case 17: `.flow/project.md` is a directory, not a regular file — exit 2.
# ===========================================================================
new_root
rm -f "$CFG"
mkdir -p "$CFG"
run_guard "stats/web/src/App.tsx"
assert_rc "case 17" 2

# ===========================================================================
# Case 18: the file declares `## visual verification` twice — exit 2,
# ambiguous, never picking a winner.
# ===========================================================================
new_root
write_cfg "$MINIMAL_HEADER

$MINIMAL_HEADER"
run_guard "stats/web/src/App.tsx"
assert_rc "case 18" 2

# ===========================================================================
# Case 19 (REPRODUCE, DON'T READ; task 19): `.flow/project.md` begins with a
# UTF-8 BOM, with `## visual verification` as the file's own first line. A
# BOM must not make a present section read as "not configured" — the
# consequence is worse here than in the other two guards, because this is
# the guard that decides whether flow.visual-verify runs at all: task 19's
# own reproduction showed this exact guard exiting 2 ("not configured")
# against a BOM-prefixed declaration before the fix.
#
# NON-VACUOUS BY CONSTRUCTION, not by a substring check on the message: a
# matching path must exit 0 with `MATCH` (case 19), and a non-matching path
# against the SAME BOM-prefixed config must exit 1, not 2 (case 19b) — only
# a parser that actually read the BOM'd section, rather than one still
# defaulting to a fixed "cannot answer" verdict, can produce two different
# exit codes from two different inputs against the same file.
# ===========================================================================
new_root
write_cfg_bom "$MINIMAL_HEADER"
run_guard "stats/web/src/App.tsx"
assert_rc "case 19" 0
assert_out_contains "case 19" "MATCH"
assert_out_not_contains "case 19" "not configured"

new_root
write_cfg_bom "$MINIMAL_HEADER"
run_guard "README.md"
assert_rc "case 19b" 1
assert_out_not_contains "case 19b" "not configured"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
