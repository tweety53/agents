#!/usr/bin/env bash
# Assertion harness for resolve-visual-screenshots.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the guard at each,
# invoking the REAL scripts/resolve-visual-screenshots.sh as a subprocess
# against REAL files on disk.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/resolve-visual-screenshots.sh"
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

new_root() {
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/resolve-visual-screenshots-test.XXXXXX")"
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

# run_guard <spec-basename> -> sets RC and OUT.
run_guard() {
  set +e
  OUT="$("$GUARD" "$ROOT" "$1" 2>&1)"
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

assert_out_line_count() {
  local case_name="$1" want="$2" got
  got="$(printf '%s\n' "$OUT" | grep -c '\.png$' || true)"
  [ "$got" -eq "$want" ] && pass "$case_name: $want PNG line(s)" \
    || fail "$case_name: expected $want PNG line(s), got $got: $OUT"
}

section_with_screenshots() {
  printf '## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `stats/web/src/**` |\n| `screenshots` | `%s` |\n\n| Command | Runs |\n|---------|------|\n| `verify` | `npm run test:visual` |\n| `capture` | `npx playwright test <spec>` |\n' "$1"
}

# ===========================================================================
# Case 1: matches under `<spec>-snapshots/`, the real Playwright default
# layout — exit 0, one absolute path printed.
# ===========================================================================
new_root
SS_DIR="$ROOT/tests/visual"
mkdir -p "$SS_DIR/baseline.spec.ts-snapshots"
: > "$SS_DIR/baseline.spec.ts-snapshots/dashboard-darwin.png"
write_cfg "$(section_with_screenshots "tests/visual")"
run_guard "baseline.spec.ts"
assert_rc "case 1" 0
assert_out_contains "case 1" "tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"
assert_out_line_count "case 1" 1

# Every printed path is absolute.
case "$OUT" in
  /*) pass "case 1: printed path is absolute" ;;
  *) fail "case 1: printed path is not absolute: $OUT" ;;
esac

# ===========================================================================
# Case 2: `screenshots` root IS the repository root (`.`) — still finds the
# nested spec snapshot, without erroring on the wide scan.
# ===========================================================================
new_root
mkdir -p "$ROOT/tests/visual/baseline.spec.ts-snapshots"
: > "$ROOT/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"
write_cfg "$(section_with_screenshots ".")"
run_guard "baseline.spec.ts"
assert_rc "case 2" 0
assert_out_line_count "case 2" 1

# ===========================================================================
# Case 3: a `node_modules` tree sits under the (repo-root) screenshots root
# and carries its OWN unrelated PNGs — they must never match a capture
# spec's basename.
# ===========================================================================
new_root
mkdir -p "$ROOT/node_modules/some-pkg/assets"
: > "$ROOT/node_modules/some-pkg/assets/icon.png"
: > "$ROOT/node_modules/some-pkg/assets/baseline.spec.ts.png"
mkdir -p "$ROOT/tests/visual/baseline.spec.ts-snapshots"
: > "$ROOT/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"
write_cfg "$(section_with_screenshots ".")"
run_guard "other.spec.ts"
assert_rc "case 3" 1
assert_out_not_contains "case 3" "node_modules"

# The real spec's own basename still finds ONLY the two matches that
# genuinely contain it (the real snapshot AND the node_modules decoy that
# was deliberately named to contain the string) — proving the filter reads
# the whole path, not a special-cased exclusion of node_modules.
run_guard "baseline.spec.ts"
assert_rc "case 3b" 0
assert_out_line_count "case 3b" 2
assert_out_contains "case 3b" "dashboard-darwin.png"
assert_out_contains "case 3b" "node_modules/some-pkg/assets/baseline.spec.ts.png"

# ===========================================================================
# Case 4: zero matches — exit 1, the case the stage must block on.
# ===========================================================================
new_root
mkdir -p "$ROOT/tests/visual"
write_cfg "$(section_with_screenshots "tests/visual")"
run_guard "baseline.spec.ts"
assert_rc "case 4" 1

# ===========================================================================
# Case 5: the `screenshots` directory does not exist yet at all — this is
# `capture`'s own first-run state, not a guard failure. Zero matches, exit 1.
# ===========================================================================
new_root
write_cfg "$(section_with_screenshots "tests/visual")"
run_guard "baseline.spec.ts"
assert_rc "case 5" 1

# ===========================================================================
# Case 6: `regression checkout` declared — `screenshots` resolves relative to
# THAT checkout, never the project root.
# ===========================================================================
new_root
CHECKOUT_DIR="$ROOT/checkout"
mkdir -p "$CHECKOUT_DIR/tests/visual/baseline.spec.ts-snapshots"
: > "$CHECKOUT_DIR/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"
mkdir -p "$ROOT/tests/visual/baseline.spec.ts-snapshots"
: > "$ROOT/tests/visual/baseline.spec.ts-snapshots/decoy.png"
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`gymie-frontend/**\` |
| \`screenshots\` | \`tests/visual\` |
| \`regression checkout\` | \`$CHECKOUT_DIR\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "baseline.spec.ts"
assert_rc "case 6" 0
assert_out_contains "case 6" "checkout/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"
assert_out_not_contains "case 6" "decoy.png"

# ===========================================================================
# Case 7: no `.flow/project.md` at all — exit 2, cannot answer.
# ===========================================================================
new_root
rm -f "$CFG"
run_guard "baseline.spec.ts"
assert_rc "case 7" 2

# ===========================================================================
# Case 8: config exists, no `## visual verification` section — exit 2.
# ===========================================================================
new_root
write_cfg "# Project

## run

echo hi
"
run_guard "baseline.spec.ts"
assert_rc "case 8" 2

# ===========================================================================
# Case 9: `screenshots` absent — exit 2.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "baseline.spec.ts"
assert_rc "case 9" 2

# ===========================================================================
# Case 10: `screenshots` present but empty — exit 2.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`stats/web/src/**\` |
| \`screenshots\` | |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "baseline.spec.ts"
assert_rc "case 10" 2

# ===========================================================================
# Case 11: the project root does not exist — exit 2.
# ===========================================================================
ROOT="/nonexistent/$$-$RANDOM"
run_guard "baseline.spec.ts"
assert_rc "case 11" 2

# ===========================================================================
# Case 12: usage error — missing spec basename argument — exit 2.
# ===========================================================================
new_root
write_cfg "$(section_with_screenshots "tests/visual")"
set +e
OUT="$("$GUARD" "$ROOT" 2>&1)"
RC=$?
set -e
assert_rc "case 12" 2

# ===========================================================================
# Case 13: `regression checkout` declared but not an existing directory —
# exit 2, cannot resolve where `screenshots` lives.
# ===========================================================================
new_root
write_cfg "## visual verification

| Setting | Value |
|---------|-------|
| \`ui paths\` | \`gymie-frontend/**\` |
| \`screenshots\` | \`tests/visual\` |
| \`regression checkout\` | \`$ROOT/does-not-exist\` |
| \`regression repo\` | \`git@github.com:tweety53/gymie-playwright.git\` |

| Command | Runs |
|---------|------|
| \`verify\` | \`npm run test:visual\` |
| \`capture\` | \`npx playwright test <spec>\` |"
run_guard "baseline.spec.ts"
assert_rc "case 13" 2

# ===========================================================================
# Case 14 (REPRODUCE, DON'T READ): a SUBSTRING COLLISION — `baseline.spec.ts`
# is a substring of `visual-baseline.spec.ts`, but the filter must anchor at
# a path-segment boundary rather than matching anywhere in the path.
# Resolving `baseline.spec.ts` must find the real
# `baseline.spec.ts-snapshots/` match and MUST NOT also return
# `visual-baseline.spec.ts-snapshots/`'s decoy, which sits at a different
# spec's own snapshot directory.
# ===========================================================================
new_root
mkdir -p "$ROOT/tests/visual/baseline.spec.ts-snapshots"
: > "$ROOT/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"
mkdir -p "$ROOT/tests/visual/visual-baseline.spec.ts-snapshots"
: > "$ROOT/tests/visual/visual-baseline.spec.ts-snapshots/decoy.png"
write_cfg "$(section_with_screenshots "tests/visual")"
run_guard "baseline.spec.ts"
assert_rc "case 14" 0
assert_out_line_count "case 14" 1
assert_out_contains "case 14" "baseline.spec.ts-snapshots/dashboard-darwin.png"
assert_out_not_contains "case 14" "visual-baseline.spec.ts-snapshots"
assert_out_not_contains "case 14" "decoy.png"

# ===========================================================================
# Case 15 (REPRODUCE, DON'T READ; task 19): `.flow/project.md` begins with a
# UTF-8 BOM, with `## visual verification` as the file's own first line. A
# BOM must not make a present section read as absent — task 19's own
# reproduction showed this exact guard exiting 2 ("declares no ... section")
# against a BOM-prefixed declaration before the fix.
#
# NON-VACUOUS: asserts the real PNG path is printed (not merely that the
# exit code is 0, which a guard that fell back to some other default root
# could also produce) AND that the exit-2 "declares no" message is absent —
# only a parser that actually read the BOM'd `screenshots` setting resolves
# the correct search root and finds this file.
# ===========================================================================
new_root
mkdir -p "$ROOT/tests/visual/baseline.spec.ts-snapshots"
: > "$ROOT/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"
write_cfg_bom "$(section_with_screenshots "tests/visual")"
run_guard "baseline.spec.ts"
assert_rc "case 15" 0
assert_out_contains "case 15" "tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"
assert_out_not_contains "case 15" "declares no"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
