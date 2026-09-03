#!/usr/bin/env bash
# Assertion harness for check-spec-reach.sh.
#
# Builds throwaway project roots and regression checkouts under TMPDIR and
# runs the REAL scripts/check-spec-reach.sh against them — never a copy of its
# logic. Playwright is not installed in a fixture: a fake `npm` placed first
# on PATH answers `npm run -s <script> -- --list --reporter=json` with the
# canned JSON the fixture wrote as `<checkout>/.list-<script>.json` (a
# `.fail-<script>` file makes it exit 7 instead), so script selection, the
# union across scripts, orphan reporting and every exit code are exercised
# without a browser. `node` is real: the guard needs it to read package.json
# and the reporter JSON, and it is on every machine Playwright runs on.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-spec-reach.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do rm -rf "$d"; done
}
trap cleanup EXIT

FAKE_BIN="$(mktemp -d "${TMPDIR:-/tmp}/check-spec-reach-bin.XXXXXX")"
DIRS+=("$FAKE_BIN")
cat > "$FAKE_BIN/npm" <<'NPM'
#!/usr/bin/env bash
# fake npm: `npm run -s <script> -- --list --reporter=json`, answered from cwd.
script="$3"
if [ -e ".fail-$script" ]; then echo "playwright: boom for $script" >&2; exit 7; fi
cat ".list-$script.json"
NPM
chmod +x "$FAKE_BIN/npm"
export PATH="$FAKE_BIN:$PATH"

# new_root -> ROOT (project root with .flow/) and CHECKOUT (regression checkout).
new_root() {
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-spec-reach-root.XXXXXX")"
  CHECKOUT="$(mktemp -d "${TMPDIR:-/tmp}/check-spec-reach-co.XXXXXX")"
  DIRS+=("$ROOT" "$CHECKOUT")
  mkdir -p "$ROOT/.flow"
  CFG="$ROOT/.flow/project.md"
}

# cfg_with_checkout [bom] — a minimal section naming $CHECKOUT.
cfg_with_checkout() {
  local prefix=""
  [ "${1:-}" = "bom" ] && prefix=$'\xef\xbb\xbf'
  printf '%s## visual verification\n\n| Setting | Value |\n|---|---|\n| `screenshots` | `.` |\n| `regression checkout` | `%s` |\n' "$prefix" "$CHECKOUT" > "$CFG"
}

# package_json <script>=<command> ...
package_json() {
  local entries="" kv
  for kv in "$@"; do
    entries="$entries\"${kv%%=*}\": \"${kv#*=}\", "
  done
  printf '{ "scripts": { %s } }\n' "${entries%, }" > "$CHECKOUT/package.json"
}

# listing <script> <file>... — canned reporter JSON for one script.
listing() {
  local script="$1"; shift
  local suites="" f
  for f in "$@"; do suites="$suites{\"file\": \"$f\", \"suites\": []}, "; done
  printf '{ "config": { "rootDir": "%s" }, "suites": [ %s ] }\n' "$CHECKOUT" "${suites%, }" > "$CHECKOUT/.list-$script.json"
}

spec() { mkdir -p "$(dirname "$CHECKOUT/$1")"; : > "$CHECKOUT/$1"; }

run_guard() {
  set +e
  OUT="$("$GUARD" "$ROOT" 2>&1)"
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

# Case 1: no .flow/project.md — not configured, exit 0.
new_root; rm -f "$CFG"
run_guard; assert_rc "case 1" 0; assert_out_contains "case 1" "Spec reach: not configured"

# Case 2: no `## visual verification` section — exit 0.
new_root; printf '# Project\n\n## run\n\necho hi\n' > "$CFG"
run_guard; assert_rc "case 2" 0; assert_out_contains "case 2" "Spec reach: not configured"

# Case 3: section without a `regression checkout` row — exit 0.
new_root; printf '## visual verification\n\n| Setting | Value |\n|---|---|\n| `screenshots` | `.` |\n' > "$CFG"
run_guard; assert_rc "case 3" 0; assert_out_contains "case 3" "Spec reach: not configured"

# Case 4: section declared twice — ambiguous, exit 2.
new_root; cfg_with_checkout; printf '\n## visual verification\n' >> "$CFG"
run_guard; assert_rc "case 4" 2; assert_out_contains "case 4" "ambiguous"

# Case 5: checkout is not a directory — exit 2.
new_root; cfg_with_checkout; rm -rf "$CHECKOUT"
run_guard; assert_rc "case 5" 2; assert_out_contains "case 5" "not an existing directory"

# Case 6: checkout holds no package.json — exit 2.
new_root; cfg_with_checkout
run_guard; assert_rc "case 6" 2; assert_out_contains "case 6" "no package.json"

# Case 7: no script contains `playwright test` — exit 2.
new_root; cfg_with_checkout; package_json "test=vitest run"
run_guard; assert_rc "case 7" 2; assert_out_contains "case 7" "playwright test"

# Case 8: two scripts, union covers every spec — exit 0; a non-playwright
# script is ignored, a spec under node_modules is not enumerated.
new_root; cfg_with_checkout
package_json "test:e2e=playwright test tests/" "shots=playwright test shots.spec.ts --headed" "lint=eslint ."
spec tests/a.spec.ts; spec shots.spec.ts; spec node_modules/dep/x.spec.ts
listing "test:e2e" tests/a.spec.ts; listing shots shots.spec.ts
run_guard; assert_rc "case 8" 0; assert_out_contains "case 8" "Spec reach: 2 spec(s), all reached"

# Case 9: one orphan — exit 1, named relative to the checkout, reached ones silent.
new_root; cfg_with_checkout
package_json "test:e2e=playwright test tests/"
spec tests/a.spec.ts; spec kan-347-route-e.spec.ts
listing "test:e2e" tests/a.spec.ts
run_guard; assert_rc "case 9" 1
assert_out_contains "case 9" "kan-347-route-e.spec.ts: reached by no package.json script"
assert_out_not_contains "case 9" "tests/a.spec.ts:"
assert_out_contains "case 9" "1 of 2 spec(s)"

# Case 10: a `--list` run fails — exit 2, its output relayed.
new_root; cfg_with_checkout
package_json "test:e2e=playwright test tests/"
spec tests/a.spec.ts; : > "$CHECKOUT/.fail-test:e2e"
run_guard; assert_rc "case 10" 2; assert_out_contains "case 10" "boom for test:e2e"

# Case 11: a `--list` run prints no JSON — exit 2.
new_root; cfg_with_checkout
package_json "test:e2e=playwright test tests/"
spec tests/a.spec.ts; echo "not json" > "$CHECKOUT/.list-test:e2e.json"
run_guard; assert_rc "case 11" 2; assert_out_contains "case 11" "no parsable JSON"

# Case 12: a BOM before the heading still resolves the checkout.
new_root; cfg_with_checkout bom
package_json "test:e2e=playwright test tests/"
spec tests/a.spec.ts; listing "test:e2e" tests/a.spec.ts
run_guard; assert_rc "case 12" 0; assert_out_contains "case 12" "all reached"

# Case 13: project root argument is not a directory — exit 2, clean message.
set +e
OUT="$("$GUARD" "/no/such/root-$$" 2>&1)"
RC=$?
set -e
assert_rc "case 13" 2; assert_out_contains "case 13" "is not a directory"

# Case 14: malformed package.json — exit 2 with the guard's own diagnostic,
# not the downstream node-error-as-script-name garble.
new_root; cfg_with_checkout
package_json "test:e2e=playwright test tests/"
printf '{ not valid json' > "$CHECKOUT/package.json"
spec tests/a.spec.ts
run_guard; assert_rc "case 14" 2
assert_out_contains "case 14" "could not read scripts from"

# Case 15: a second, unrelated `##` section immediately follows
# `## visual verification`'s own table — its rows must not bleed into the
# section above. The sibling section's `regression checkout` row names a
# nonexistent directory; if the boundary test let it through, the guard
# would fail on that row instead of correctly reporting not-configured.
new_root
printf '## visual verification\n\n| Setting | Value |\n|---|---|\n| `screenshots` | `.` |\n\n## unrelated section\n\n| Setting | Value |\n|---|---|\n| `regression checkout` | `/definitely/not/real/path-%s` |\n' "$$" > "$CFG"
run_guard; assert_rc "case 15" 0; assert_out_contains "case 15" "Spec reach: not configured"

# Case 16: two `regression checkout` rows in one `## visual verification`
# section — first-row-wins, current policy locked in (unlike a duplicate
# heading, which is rejected as ambiguous).
new_root
DUP_CHECKOUT="$(mktemp -d "${TMPDIR:-/tmp}/check-spec-reach-co.XXXXXX")"
DIRS+=("$DUP_CHECKOUT")
printf '## visual verification\n\n| Setting | Value |\n|---|---|\n| `regression checkout` | `%s` |\n| `regression checkout` | `%s` |\n' "$CHECKOUT" "$DUP_CHECKOUT" > "$CFG"
package_json "test:e2e=playwright test tests/"
spec tests/a.spec.ts
listing "test:e2e" tests/a.spec.ts
run_guard; assert_rc "case 16" 0; assert_out_contains "case 16" "all reached"

# Case 17: a decoy listing's `"file"` is an absolute path with an extra
# prefix segment that keeps a real orphan spec's absolute path as a
# contiguous substring (not an equal string) — `path.resolve(root, file)`
# returns an absolute `file` unchanged, so the REACHED line is exactly that
# crafted string. `grep -qxF` (exact line) must still report the real spec
# as an orphan; only a `-qF` (substring) mutation would wrongly call it
# reached, masking the very silent-orphan bug this guard exists to catch.
new_root; cfg_with_checkout
package_json "real=playwright test real/" "decoy=playwright test decoy/"
spec real/a.spec.ts
listing real
printf '{ "config": { "rootDir": "%s" }, "suites": [ {"file": "/zzz%s/real/a.spec.ts", "suites": []} ] }\n' \
  "$CHECKOUT" "$CHECKOUT" > "$CHECKOUT/.list-decoy.json"
run_guard; assert_rc "case 17" 1
assert_out_contains "case 17" "real/a.spec.ts: reached by no package.json script"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
