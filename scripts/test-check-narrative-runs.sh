#!/usr/bin/env bash
# Assertion harness for check-narrative-runs.sh. Runs the guard against a
# stub `flow` on PATH whose `stage sessions` answers per stage from fixture
# files, in a sandboxed mktemp directory.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-narrative-runs.sh"
FAILURES=0
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

mkdir -p "$WORK/bin"
cat > "$WORK/bin/flow" <<STUB
#!/usr/bin/env bash
# flow stage sessions -C <wt> -stage <key> <change>: prints \$WORK/<key>, or fails when absent.
[ "\$1 \$2" = "stage sessions" ] || exit 9
cat "$WORK/\$6" 2>/dev/null || exit 1
STUB
chmod +x "$WORK/bin/flow"
export PATH="$WORK/bin:$PATH"

# setup <fix runs> <integrate runs> [narrative body]
setup() {
  rm -f "$WORK/flow.document-fix" "$WORK/flow.preserve-sessions" "$WORK/narrative.md"
  printf '%s\n' "$1" >"$WORK/flow.document-fix"
  printf '%s\n' "$2" >"$WORK/flow.preserve-sessions"
  [ "$#" -lt 3 ] || printf '%s\n' "$3" >"$WORK/narrative.md"
}
run_guard() {
  set +e
  OUT="$("$GUARD" "$WORK" demo "$WORK/narrative.md" 2>&1)"
  RC=$?
  set -e
}

setup 1 0 "## 2026-10-19 — implement run"
run_guard
[ "$RC" -eq 1 ] && [[ "$OUT" == NARRATIVE-RUNS-MISSING:* ]] && pass "a fix run with no section exits 1" \
  || fail "a fix run with no section exits 1: rc=$RC out=$OUT"

setup 1 1 "## 2026-10-19 — fix run"
run_guard
[ "$RC" -eq 1 ] && [[ "$OUT" == NARRATIVE-RUNS-MISSING:* ]] && pass "an integrate run with no section exits 1" \
  || fail "an integrate run with no section exits 1: rc=$RC out=$OUT"

setup 2 1 "## 2026-10-18 — fix run
## 2026-10-19 — fix run
## 2026-10-19 — integrate run"
run_guard
[ "$RC" -eq 0 ] && [[ "$OUT" == NARRATIVE-RUNS-OK:* ]] && pass "every run with its section exits 0" \
  || fail "every run with its section exits 0: rc=$RC out=$OUT"

setup 0 0
run_guard
[ "$RC" -eq 0 ] && pass "an absent narrative with no runs exits 0" \
  || fail "an absent narrative with no runs exits 0: rc=$RC out=$OUT"

setup 0 0 "## x"
rm -f "$WORK/flow.preserve-sessions"
run_guard
[ "$RC" -eq 2 ] && pass "a failed store read exits 2" \
  || fail "a failed store read exits 2: rc=$RC out=$OUT"

[ "$FAILURES" -eq 0 ] || { printf '%d failure(s)\n' "$FAILURES" >&2; exit 1; }
