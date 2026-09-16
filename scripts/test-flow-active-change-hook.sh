#!/usr/bin/env bash
# Assertion harness for hooks/flow-active-change.py. Builds fixture transcripts under a
# sandboxed mktemp directory and pipes fixture stdin JSON to the hook, asserting its
# stdout and exit code — never one alone. Modeled on
# scripts/test-check-self-review-report.sh's fixture-driven pattern: a run_hook helper
# captures RC/OUT, cases are numbered exactly as the plan's **Tests:** field names them.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOK="$REPO/hooks/flow-active-change.py"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/test-flow-active-change-hook.XXXXXX")"
trap 'rm -rf "$SANDBOX"' EXIT

# run_hook <payload-json> -> sets RC and OUT
run_hook() {
  set +e
  OUT="$(printf '%s' "$1" | python3 "$HOOK" 2>&1)"
  RC=$?
  set -e
}

mark_line() {
  # mark_line <change-name> -> one flow stage begin mark line for a /flow run
  printf "flow stage begin -command '/flow' -stage flow.document-fix -harness claude-code -session-token mf-abc123 %s" "$1"
}

# ===========================================================================
# Case 1: one /flow mark — the context line names its change
# ===========================================================================
T="$SANDBOX/case1.jsonl"
printf '{"type":"tool_result","command":"%s"}\n' "$(mark_line fixture-one)" >"$T"
run_hook "$(printf '{"prompt":"the button is misaligned","transcript_path":"%s"}' "$T")"
[ "$RC" -eq 0 ] && pass "case 1: exit 0" || fail "case 1: rc=$RC out=$OUT"
case "$OUT" in
  *"marked change fixture-one"*) pass "case 1: context line names fixture-one" ;;
  *) fail "case 1: expected 'marked change fixture-one' in out=$OUT" ;;
esac

# ===========================================================================
# Case 2: two /flow marks for different changes — the last one wins
# ===========================================================================
T="$SANDBOX/case2.jsonl"
{
  printf '{"type":"tool_result","command":"%s"}\n' "$(mark_line fixture-one)"
  printf '{"type":"tool_result","command":"%s"}\n' "$(mark_line fixture-two)"
} >"$T"
run_hook "$(printf '{"prompt":"still broken","transcript_path":"%s"}' "$T")"
[ "$RC" -eq 0 ] && pass "case 2: exit 0" || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"marked change fixture-two"*) pass "case 2: last mark (fixture-two) wins" ;;
  *) fail "case 2: expected 'marked change fixture-two' in out=$OUT" ;;
esac

# ===========================================================================
# Case 3: only /flow-plan and /flow-fast marks — no output, exit 0
# ===========================================================================
T="$SANDBOX/case3.jsonl"
{
  printf "flow stage begin -command '/flow-plan' -stage flow.research -harness claude-code -session-token mf-x fixture-plan\n"
  printf "flow stage begin -command '/flow-fast' -stage do.implement -harness claude-code -session-token mf-y fixture-fast\n"
} | while IFS= read -r cmd; do printf '{"type":"tool_result","command":"%s"}\n' "$cmd"; done >"$T"
run_hook "$(printf '{"prompt":"any problem here","transcript_path":"%s"}' "$T")"
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "case 3: no /flow mark -> exit 0, empty output" || fail "case 3: rc=$RC out=$OUT"

# ===========================================================================
# Case 4: a prompt starting with / — no output, exit 0
# ===========================================================================
T="$SANDBOX/case4.jsonl"
printf '{"type":"tool_result","command":"%s"}\n' "$(mark_line fixture-one)" >"$T"
run_hook "$(printf '{"prompt":"/flow x","transcript_path":"%s"}' "$T")"
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "case 4: slash prompt -> exit 0, empty output" || fail "case 4: rc=$RC out=$OUT"

# ===========================================================================
# Case 5: malformed stdin — no output, exit 0
# ===========================================================================
set +e
OUT="$(printf 'not json' | python3 "$HOOK" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "case 5: malformed stdin -> exit 0, empty output" || fail "case 5: rc=$RC out=$OUT"

# ===========================================================================
# Case 6: no transcript_path — the rollout under FLOW_ZCODE_ROLLOUTS_DIR
# named by session_id is read
# ===========================================================================
ZDIR="$SANDBOX/zroot"
mkdir -p "$ZDIR"
printf '{"type":"tool_result","command":"%s"}\n' "$(mark_line fixture-six)" >"$ZDIR/model-io-sess_abc.jsonl"
set +e
OUT="$(FLOW_ZCODE_ROLLOUTS_DIR="$ZDIR" printf '{"prompt":"broken again","session_id":"abc"}' | \
  FLOW_ZCODE_ROLLOUTS_DIR="$ZDIR" python3 "$HOOK" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "case 6: exit 0" || fail "case 6: rc=$RC out=$OUT"
case "$OUT" in
  *"marked change fixture-six"*) pass "case 6: zcode rollout resolved via session_id" ;;
  *) fail "case 6: expected 'marked change fixture-six' in out=$OUT" ;;
esac

if [ "$FAILURES" -gt 0 ]; then
  printf 'FAILED: %d assertion(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'PASS: test-flow-active-change-hook.sh\n'
