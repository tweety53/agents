#!/usr/bin/env bash
# Assertion harness for hooks/flow-active-change.py. Starts a stub store (a tiny
# HTTP server answering GET /api/v1/stage-runs with a canned body) under a sandboxed
# mktemp directory, points the hook at it through FLOW_ADDR, and pipes fixture stdin
# JSON to the hook, asserting its stdout and exit code — never one alone. Modeled on
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
trap 'rm -rf "$SANDBOX"; [ -n "${STUB_PID:-}" ] && kill "$STUB_PID" 2>/dev/null || true' EXIT

# A free port picked at runtime, not a hard-coded one: an occupied port would
# fail all nine cases for a reason the hook does not own.
STUB_PORT="$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()')"
STUB_BODY="$SANDBOX/stub-body.json"
STUB_STATUS="$SANDBOX/stub-status.txt"
STUB_LOG="$SANDBOX/stub-log.txt"
: >"$STUB_LOG"
printf '200' >"$STUB_STATUS"

# The stub reads its body and status per request, so a case rewrites either without
# restarting the server; every request's path is appended to STUB_LOG for the
# query-shape assertions below.
cat >"$SANDBOX/stub.py" <<'PYEOF'
import os
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer

BODY = os.environ["STUB_BODY"]
STATUS = os.environ["STUB_STATUS"]
LOG = os.environ["STUB_LOG"]


class Stub(BaseHTTPRequestHandler):
    def do_GET(self):
        with open(LOG, "a") as f:
            f.write(self.path + "\n")
        with open(STATUS) as f:
            status = int(f.read().strip())
        with open(BODY, "rb") as f:
            body = f.read()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *args):
        pass


HTTPServer(("127.0.0.1", int(sys.argv[1])), Stub).serve_forever()
PYEOF

STUB_BODY="$STUB_BODY" STUB_STATUS="$STUB_STATUS" STUB_LOG="$STUB_LOG" \
  python3 "$SANDBOX/stub.py" "$STUB_PORT" &
STUB_PID=$!

# Wait for the stub to accept connections before the first case: the stub
# interpreter's own startup leaves a window where the port still refuses
# (measured 145 ms idle; longer under run-guard-tests.sh's parallel load), and
# case 1's urlopen fires inside it — the hook fails open on the refusal, so the
# case reads as an empty store answer rather than as the startup race it is.
python3 - "$STUB_PORT" <<'PYEOF'
import socket
import sys
import time

deadline = time.time() + 10
while True:
    try:
        socket.create_connection(("127.0.0.1", int(sys.argv[1])), timeout=0.25).close()
        sys.exit(0)
    except OSError:
        if time.time() > deadline:
            sys.exit(1)
        time.sleep(0.05)
PYEOF

ADDR="http://127.0.0.1:$STUB_PORT"
DEAD_ADDR="http://127.0.0.1:1"

# run_hook <payload-json> [addr] -> sets RC and OUT
run_hook() {
  set +e
  OUT="$(printf '%s' "$1" | FLOW_ADDR="${2:-$ADDR}" python3 "$HOOK" 2>&1)"
  RC=$?
  set -e
}

# set_body <rows-json> -> the stub's next stageRuns answer
set_body() {
  printf '{"total":%s,"stageRuns":[%s]}' "$1" "$2" >"$STUB_BODY"
}

set_status() {
  printf '%s' "$1" >"$STUB_STATUS"
}

reset_log() {
  : >"$STUB_LOG"
}

# row <change-name-or-empty> -> one stage-run row with (or without) a changeName
row() {
  if [ -n "$1" ]; then
    printf '{"stageRunId":7,"command":"/flow","stage":"flow.kickoff","startedAt":"2026-09-17T09:00:00Z","changeName":"%s"}' "$1"
  else
    printf '{"stageRunId":7,"command":"/flow","stage":"plan.session","startedAt":"2026-09-17T09:00:00Z"}'
  fi
}

# ===========================================================================
# Case 1: the store holds the session's /flow mark — the context line names
# its change, and the query filters by the session id derived from the
# rollout transcript's filename, newest mark first, one row
# ===========================================================================
set_body 1 "$(row fixture-one)"
reset_log
run_hook "$(printf '{"prompt":"the button is misaligned","transcript_path":"%s/model-io-sess_abc.jsonl"}' "$SANDBOX")"
[ "$RC" -eq 0 ] && pass "case 1: exit 0" || fail "case 1: rc=$RC out=$OUT"
case "$OUT" in
  *"marked change fixture-one"*) pass "case 1: context line names fixture-one" ;;
  *) fail "case 1: expected 'marked change fixture-one' in out=$OUT" ;;
esac
QUERY="$(tail -1 "$STUB_LOG")"
case "$QUERY" in
  *session_id=abc*) pass "case 1: session_id derived from the rollout filename" ;;
  *) fail "case 1: expected session_id=abc in query=$QUERY" ;;
esac
case "$QUERY" in
  *command=%2Fflow*) pass "case 1: command=/flow filter in the query" ;;
  *) fail "case 1: expected command=%2Fflow in query=$QUERY" ;;
esac
case "$QUERY" in
  *sort=-started_at*limit=1*) pass "case 1: newest-mark-first, one row" ;;
  *) fail "case 1: expected sort=-started_at&limit=1 in query=$QUERY" ;;
esac

# ===========================================================================
# Case 2: two /flow marks for different changes — the store answers newest
# first and the hook reads stageRuns[0], so the last mark wins
# ===========================================================================
set_body 2 "$(row fixture-two),$(row fixture-one)"
run_hook "$(printf '{"prompt":"still broken","session_id":"abc","transcript_path":"%s/model-io-sess_abc.jsonl"}' "$SANDBOX")"
[ "$RC" -eq 0 ] && pass "case 2: exit 0" || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"marked change fixture-two"*) pass "case 2: last mark (fixture-two) wins" ;;
  *) fail "case 2: expected 'marked change fixture-two' in out=$OUT" ;;
esac

# ===========================================================================
# Case 3: no /flow mark in the store (e.g. only /flow-plan and /flow-fast
# marks, which the command filter excludes) — no output, exit 0
# ===========================================================================
set_body 0 ""
run_hook '{"prompt":"any problem here","session_id":"abc"}'
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "case 3: no /flow mark -> exit 0, empty output" || fail "case 3: rc=$RC out=$OUT"

# ===========================================================================
# Case 4: a prompt starting with / — no output, exit 0, and no store call
# ===========================================================================
set_body 1 "$(row fixture-one)"
reset_log
run_hook '{"prompt":"/flow x","session_id":"abc"}'
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "case 4: slash prompt -> exit 0, empty output" || fail "case 4: rc=$RC out=$OUT"
[ ! -s "$STUB_LOG" ] && pass "case 4: store not consulted" || fail "case 4: expected no store call, log=$(cat "$STUB_LOG")"

# ===========================================================================
# Case 5: malformed stdin — no output, exit 0
# ===========================================================================
set +e
OUT="$(printf 'not json' | FLOW_ADDR="$ADDR" python3 "$HOOK" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "case 5: malformed stdin -> exit 0, empty output" || fail "case 5: rc=$RC out=$OUT"

# ===========================================================================
# Case 6: no transcript_path — the session_id from the hook payload is the
# store query's session filter
# ===========================================================================
set_body 1 "$(row fixture-six)"
reset_log
run_hook '{"prompt":"broken again","session_id":"sess-six"}'
[ "$RC" -eq 0 ] && pass "case 6: exit 0" || fail "case 6: rc=$RC out=$OUT"
case "$OUT" in
  *"marked change fixture-six"*) pass "case 6: payload session_id queried" ;;
  *) fail "case 6: expected 'marked change fixture-six' in out=$OUT" ;;
esac
QUERY="$(tail -1 "$STUB_LOG")"
case "$QUERY" in
  *session_id=sess-six*) pass "case 6: query filtered by the payload session id" ;;
  *) fail "case 6: expected session_id=sess-six in query=$QUERY" ;;
esac

# ===========================================================================
# Case 7: the store is unreachable — fail open, no output, exit 0
# ===========================================================================
set_body 1 "$(row fixture-one)"
run_hook '{"prompt":"broken","session_id":"abc"}' "$DEAD_ADDR"
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "case 7: unreachable store -> exit 0, empty output" || fail "case 7: rc=$RC out=$OUT"

# ===========================================================================
# Case 8: the store answers 500 — fail open, no output, exit 0
# ===========================================================================
set_status 500
run_hook '{"prompt":"broken","session_id":"abc"}'
set_status 200
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "case 8: store 500 -> exit 0, empty output" || fail "case 8: rc=$RC out=$OUT"

# ===========================================================================
# Case 9: the newest row is an unattached plan session (no changeName) —
# no change to name, no output, exit 0
# ===========================================================================
set_body 1 "$(row "")"
run_hook '{"prompt":"broken","session_id":"abc"}'
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "case 9: nil changeName -> exit 0, empty output" || fail "case 9: rc=$RC out=$OUT"

if [ "$FAILURES" -gt 0 ]; then
  printf 'FAILED: %d assertion(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'PASS: test-flow-active-change-hook.sh\n'
