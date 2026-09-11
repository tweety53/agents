#!/usr/bin/env bash
# Assertion harness for check-dev-stack-fresh.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the guard at each,
# invoking the REAL scripts/check-dev-stack-fresh.sh as a subprocess against
# REAL files on disk — never a copy of its logic, never a stubbed parser.
#
# THE STALE/FRESH CASES USE A REAL HTTP SERVER, not a fixture standing in for
# one: python3 -m http.server serves a real document from a real socket, and
# the declared fingerprint fetches it with a real curl piped into a real cmp —
# the same boundary the run-instructions stage's fingerprint crosses. A
# hand-built "served" value would prove the plumbing and not the probe.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-dev-stack-fresh.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
SERVER_PID=""
cleanup() {
  [ -n "$SERVER_PID" ] && kill "$SERVER_PID" 2>/dev/null || true
  local d
  for d in "${DIRS[@]:-}"; do
    [ -n "$d" ] && rm -rf "$d"
  done
}
trap cleanup EXIT

new_root() {
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-dev-stack-fresh-test.XXXXXX")"
  DIRS+=("$ROOT")
  mkdir -p "$ROOT/.flow"
  CFG="$ROOT/.flow/project.md"
}

write_cfg() {
  printf '%s\n' "$1" > "$CFG"
}

run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>&1)"
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

# section_with_fp <runs-cell> — a minimal well-formed section whose commands
# table carries exactly one `fingerprint` row. The cell is already escaped
# for a Markdown table by the caller.
section_with_fp() {
  printf '## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `src/**` |\n\n| Command | Runs |\n|---------|------|\n| `fingerprint` | `%s` |\n' "$1"
}

# ===========================================================================
# Case 1: fingerprint exits 0 — fresh.
# ===========================================================================
new_root
printf 'served\n' > "$ROOT/served.html"
printf 'served\n' > "$ROOT/built.html"
write_cfg "$(section_with_fp 'cmp -s served.html built.html')"
run_guard "$ROOT"
assert_rc "case 1" 0
assert_out_contains "case 1" "stack-fresh"

# ===========================================================================
# Case 2: fingerprint exits non-zero — stale.
# ===========================================================================
new_root
printf 'old\n' > "$ROOT/served.html"
printf 'new\n' > "$ROOT/built.html"
write_cfg "$(section_with_fp 'cmp -s served.html built.html')"
run_guard "$ROOT"
assert_rc "case 2" 1
assert_out_contains "case 2" "stack-stale"

# ===========================================================================
# Case 3: the real boundary — a real HTTP server serving a real document,
# fetched with a real curl, compared with a real cmp. Stale first, then the
# same stack serving the fresh document flips the verdict.
# ===========================================================================
new_root
SERVE_DIR="$ROOT/site"
mkdir -p "$SERVE_DIR"
printf 'stale-bundle\n' > "$SERVE_DIR/index.html"
printf 'fresh-bundle\n' > "$ROOT/dist.html"
PORT=$(( (RANDOM % 2000) + 20000 ))
# A port another process holds would make this case fail (or pass) for the
# wrong reason, so wait for a genuinely free one before starting, and treat a
# server that never warms up as the case's own failure — never as a verdict
# about the guard.
until [ -z "$(curl -sf -o /dev/null "http://127.0.0.1:$PORT/" 2>/dev/null)" ]; do
  PORT=$(( PORT + 1 ))
done
python3 -m http.server "$PORT" --bind 127.0.0.1 --directory "$SERVE_DIR" >/dev/null 2>&1 &
SERVER_PID=$!
WARMED=0
for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
  if curl -sf "http://127.0.0.1:$PORT/index.html" >/dev/null 2>&1; then WARMED=1; break; fi
  sleep 0.3
done
if [ "$WARMED" -ne 1 ]; then
  fail "case 3: server never warmed up on port $PORT — environment cannot run this case"
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
write_cfg "$(section_with_fp "curl -sf http://127.0.0.1:$PORT/index.html \| cmp -s - dist.html")"
run_guard "$ROOT"
assert_rc "case 3 stale" 1
assert_out_contains "case 3 stale" "stack-stale"
printf 'fresh-bundle\n' > "$SERVE_DIR/index.html"
run_guard "$ROOT"
assert_rc "case 3 fresh" 0
assert_out_contains "case 3 fresh" "stack-fresh"
kill "$SERVER_PID" 2>/dev/null || true
SERVER_PID=""

# ===========================================================================
# Case 4: the command runs with the worktree as its working directory.
# ===========================================================================
new_root
write_cfg "$(section_with_fp 'test -f .flow/project.md')"
run_guard "$ROOT"
assert_rc "case 4" 0
assert_out_contains "case 4" "stack-fresh"

# ===========================================================================
# Case 5: a fingerprint row whose command carries an escaped pipe — the cell
# unescapes to a real shell pipe and the command runs as one.
# ===========================================================================
new_root
write_cfg "$(section_with_fp 'printf a-b\n \| grep -q b')"
run_guard "$ROOT"
assert_rc "case 5" 0
assert_out_contains "case 5" "printf a-b"

# ===========================================================================
# Case 6: a fingerprint row inside a fenced worked example never declares —
# the fence hides it and the absence is exit 2, not a run of example text.
# ===========================================================================
new_root
write_cfg '## visual verification

| Setting | Value |
|---------|-------|
| `ui paths` | `src/**` |

```markdown
| Command | Runs |
|---------|------|
| `fingerprint` | `exit 7` |
```'
run_guard "$ROOT"
assert_rc "case 6" 2
assert_out_contains "case 6" "no readable \`fingerprint\` row"

# ===========================================================================
# Case 7: a real row after the fenced example is still found.
# ===========================================================================
new_root
write_cfg '## visual verification

| Setting | Value |
|---------|-------|
| `ui paths` | `src/**` |

```markdown
| Command | Runs |
|---------|------|
| `fingerprint` | `exit 7` |
```

| Command | Runs |
|---------|------|
| `fingerprint` | `true` |'
run_guard "$ROOT"
assert_rc "case 7" 0
assert_out_contains "case 7" "stack-fresh"

# ===========================================================================
# Case 8: no .flow/project.md at all — exit 2.
# ===========================================================================
new_root
rm -rf "$ROOT/.flow"
run_guard "$ROOT"
assert_rc "case 8" 2
assert_out_contains "case 8" "does not exist"

# ===========================================================================
# Case 9: project.md without the section — exit 2.
# ===========================================================================
new_root
write_cfg '## run

- `make build`'
run_guard "$ROOT"
assert_rc "case 9" 2
assert_out_contains "case 9" "no readable \`fingerprint\` row"

# ===========================================================================
# Case 10: the section present, commands table present, no fingerprint row.
# ===========================================================================
new_root
write_cfg '## visual verification

| Command | Runs |
|---------|------|
| `verify` | `npm run test:visual` |'
run_guard "$ROOT"
assert_rc "case 10" 2
assert_out_contains "case 10" "no readable \`fingerprint\` row"

# ===========================================================================
# Case 11: usage — no arguments, two arguments, a non-directory.
# ===========================================================================
run_guard
assert_rc "case 11 no args" 2
assert_out_contains "case 11 no args" "usage"
run_guard "$ROOT" extra
assert_rc "case 11 two args" 2
run_guard "$ROOT/no-such-dir"
assert_rc "case 11 not a directory" 2

# ===========================================================================
# Case 12: the settings table never arms the commands scan — a `fingerprint`
# row under a `Setting | Value` header is not a command.
# ===========================================================================
new_root
write_cfg '## visual verification

| Setting | Value |
|---------|-------|
| `fingerprint` | `true` |'
run_guard "$ROOT"
assert_rc "case 12" 2

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'all cases passed\n'
