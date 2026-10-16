#!/usr/bin/env bash
# Assertion harness for hooks/enforce-loop-dispatch.py. Points the hook at a stub
# checker through LOOP_DISPATCH_CHECK — the guard itself is covered by
# stats/internal/guard/check_loop_dispatch_test.go — and pipes PreToolUse payloads to
# it, asserting the decision and what the checker was handed. Cases:
#
#   1. a dispatch with no loop key                    -> allow, checker never run
#   2. a loop key the checker admits (exit 0)          -> allow, checker got
#                                                         <repo> <key> <type> <model>
#   3. a loop key the checker refuses (exit 1)         -> deny, its lines in the reason
#   4. a loop key the checker cannot answer (exit 2)   -> deny
#   5. a loop key whose checker cannot run             -> deny (fails closed)
#   6. malformed JSON                                  -> allow, exit 0
#   7. a Bash call carrying a loop key                 -> allow, checker never run
#   8. a prompt naming a loop worktree, no Key line    -> deny, checker never run
#   9. a loop key with no agent-baseline pointer       -> deny, checker never run
#  10. a Key line with trailing text                   -> deny, checker never run
#  11. a pipeline-fix key with a non-numeric <k>       -> deny, checker never run
#  12. a payload carrying a cwd                       -> the checker runs in it
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$SCRIPT_DIR/.." && pwd -P)"
HOOK="$REPO/hooks/enforce-loop-dispatch.py"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/test-enforce-loop-dispatch-hook.XXXXXX")"
trap 'rm -rf "$SANDBOX"' EXIT
LOG="$SANDBOX/args"
RC_FILE="$SANDBOX/rc"
cat >"$SANDBOX/check" <<STUB
#!/usr/bin/env bash
printf '%s\n' "\$@" >"$LOG"
pwd -P >"$SANDBOX/cwd"
echo "LOOP-DISPATCH-REFUSED: PAIR — stub"
exit "\$(cat "$RC_FILE")"
STUB
chmod +x "$SANDBOX/check"

payload() { # <tool> <prompt>
  python3 -c 'import json,sys; print(json.dumps({"tool_name": sys.argv[1], "tool_input": {"prompt": sys.argv[2], "subagent_type": "flow-low", "model": "opus"}}))' "$1" "$2"
}
hook() { # <check> <stdin>
  OUT="$(printf '%s' "$2" | LOOP_DISPATCH_CHECK="$1" python3 "$HOOK")" && RC=0 || RC=$?
}
expect() { # <case> <allow|deny> <needle>
  if [ "$RC" -ne 0 ]; then fail "$1: exit $RC"; return; fi
  case "$2" in
    allow) [ -z "$OUT" ] && pass "$1" || fail "$1: expected allow, got $OUT" ;;
    deny) if grep -q '"permissionDecision": "deny"' <<<"$OUT" && grep -qF -- "$3" <<<"$OUT"; then
            pass "$1"
          else
            fail "$1: expected a deny carrying '$3', got $OUT"
          fi ;;
  esac
}

LOOP_PROMPT=$'Before anything else, read `~/.claude/rules/agent-baseline.md`.\nKey: pipeline-fix-6-fix\nDo it.'

rm -f "$LOG"; echo 0 >"$RC_FILE"
hook "$SANDBOX/check" "$(payload Agent $'Key: task-3-implementer\nwork')"
expect "1 no loop key" allow
[ ! -e "$LOG" ] && pass "1 checker not run" || fail "1 checker ran"

hook "$SANDBOX/check" "$(payload Agent "$LOOP_PROMPT")"
expect "2 admitted" allow
[ "$(tr '\n' ' ' <"$LOG")" = "$REPO pipeline-fix-6-fix flow-low opus " ] \
  && pass "2 checker args" || fail "2 checker args: $(tr '\n' ' ' <"$LOG")"

echo 1 >"$RC_FILE"
hook "$SANDBOX/check" "$(payload Task "$LOOP_PROMPT")"
expect "3 refused" deny "LOOP-DISPATCH-REFUSED: PAIR"

echo 2 >"$RC_FILE"
hook "$SANDBOX/check" "$(payload Agent "$LOOP_PROMPT")"
expect "4 cannot answer" deny "exit 2"

hook "$SANDBOX/missing" "$(payload Agent "$LOOP_PROMPT")"
expect "5 checker cannot run" deny "could not run"

hook "$SANDBOX/check" "not json"
expect "6 malformed JSON" allow

rm -f "$LOG"; echo 1 >"$RC_FILE"
hook "$SANDBOX/check" "$(python3 -c 'import json; print(json.dumps({"tool_name": "Bash", "tool_input": {"command": "echo", "prompt": "Key: pipeline-fix-6-fix"}}))')"
expect "7 not a dispatch" allow
[ ! -e "$LOG" ] && pass "7 checker not run" || fail "7 checker ran"

rm -f "$LOG"; echo 0 >"$RC_FILE"
hook "$SANDBOX/check" "$(payload Agent $'Fix it in /x/agents-worktrees/pipeline-fix-kan-1-demo-6 now.')"
expect "8 loop worktree named, no Key line" deny "carries no"
[ ! -e "$LOG" ] && pass "8 checker not run" || fail "8 checker ran"

hook "$SANDBOX/check" "$(payload Agent $'Key: pipeline-fix-6-fix\nDo it.')"
expect "9 no baseline pointer" deny "agent-baseline.md pointer"
[ ! -e "$LOG" ] && pass "9 checker not run" || fail "9 checker ran"

hook "$SANDBOX/check" "$(payload Agent $'Before anything else, read `~/.claude/rules/agent-baseline.md`.\nKey: pipeline-fix-6-fix (round 2)\nDo it.')"
expect "10 Key line with trailing text" deny "no valid loop key"
[ ! -e "$LOG" ] && pass "10 checker not run" || fail "10 checker ran"

hook "$SANDBOX/check" "$(payload Agent $'Before anything else, read `~/.claude/rules/agent-baseline.md`.\nKey: pipeline-fix-x-fix\nDo it.')"
expect "11 non-numeric <k>" deny "no valid loop key"
[ ! -e "$LOG" ] && pass "11 checker not run" || fail "11 checker ran"

mkdir "$SANDBOX/project"; echo 0 >"$RC_FILE"
hook "$SANDBOX/check" "$(python3 -c 'import json,sys; print(json.dumps({"tool_name": "Agent", "cwd": sys.argv[2], "tool_input": {"prompt": sys.argv[1], "subagent_type": "flow-low", "model": "opus"}}))' "$LOOP_PROMPT" "$SANDBOX/project")"
expect "12 admitted with a cwd" allow
[ "$(cat "$SANDBOX/cwd")" = "$(cd "$SANDBOX/project" && pwd -P)" ] \
  && pass "12 checker ran in the payload's cwd" || fail "12 checker ran in $(cat "$SANDBOX/cwd")"

if [ "$FAILURES" -gt 0 ]; then
  printf '%d case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
