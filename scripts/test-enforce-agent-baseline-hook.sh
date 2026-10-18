#!/usr/bin/env bash
# Assertion harness for hooks/enforce-agent-baseline.py. Pipes PreToolUse payloads to
# the hook and asserts its decision. Cases:
#
#   1. no baseline pointer                                 -> deny (baseline reason)
#   2. pointer, no pinned-role marker line, any type/model -> allow
#  2b. hand-picked flow-high                             -> deny
#  2c. "No-progress escalation" line on flow-high          -> allow
#  2d. "Operator-approved high" line on flow-high          -> allow
#  2e. a pipeline-fix loop key on flow-high                -> allow
#   3. verifier prompt on flow-low / opus                  -> allow
#   4. verifier prompt on flow-high / opus                 -> deny, names flow-low and opus
#   5. verifier prompt on flow-low with no model           -> deny
#   6. tooling-analyst prompt (also names the verifier
#      file) on flow-high / opus                           -> allow
#   7. tooling-analyst prompt on flow-low / opus           -> deny, names flow-high
#   8. verifier prompt run from a ~/.zcode/hooks path on
#      flow-high / glm-5.3-flash                           -> allow
#   9. the same from ~/.zcode/hooks on flow-low / opus     -> deny, names glm-5.3-flash
#  10. a fork dispatch carrying the verifier marker        -> allow
#  11. malformed JSON                                      -> allow, exit 0
#  12. a planner or implementer prompt citing the verifier
#      or analyst file, with no marker line, off the
#      pinned pair                                         -> allow
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$SCRIPT_DIR/.." && pwd -P)"
HOOK="$REPO/hooks/enforce-agent-baseline.py"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/test-enforce-agent-baseline-hook.XXXXXX")"
trap 'rm -rf "$SANDBOX"' EXIT
mkdir -p "$SANDBOX/.zcode/hooks"
ln -s "$HOOK" "$SANDBOX/.zcode/hooks/enforce-agent-baseline.py"

BASELINE='Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent.'
VERIFIER="Role: visual-verify verifier
$BASELINE Read /x/skills/flow/visual-verify-verifier.md and run its steps."
ANALYST="Role: visual-verify tooling analyst
$BASELINE Name the sweep in /x/skills/flow/visual-verify-verifier.md; write .superpowers/sdd/tooling-analysis-1.md."

# payload <prompt> <subagent_type> [model]
payload() {
  python3 -c 'import json,sys
ti={"prompt":sys.argv[1],"subagent_type":sys.argv[2]}
if len(sys.argv)>3: ti["model"]=sys.argv[3]
print(json.dumps({"tool_name":"Agent","tool_input":ti}))' "$@"
}

# run_hook <hook path> <stdin>: sets RC and OUT
run_hook() {
  set +e
  OUT="$(printf '%s' "$2" | python3 "$1")"
  RC=$?
  set -e
}

expect_allow() {
  if [ "$RC" -eq 0 ] && [ -z "$OUT" ]; then pass "$1"; else fail "$1 (rc=$RC out=$OUT)"; fi
}

# expect_deny <name> <substring>...
expect_deny() {
  local name="$1" s
  shift
  if [ "$RC" -ne 0 ] || ! grep -q '"permissionDecision": "deny"' <<<"$OUT"; then
    fail "$name (rc=$RC out=$OUT)"
    return
  fi
  for s in "$@"; do
    grep -qF -- "$s" <<<"$OUT" || { fail "$name: reason lacks $s (out=$OUT)"; return; }
  done
  pass "$name"
}

run_hook "$HOOK" "$(payload 'do the thing' flow-high opus)"
expect_deny "1 no pointer" "agent-baseline pointer"

run_hook "$HOOK" "$(payload "$BASELINE implement task 3" flow-medium sonnet)"
expect_allow "2 unpinned role"

run_hook "$HOOK" "$(payload "$BASELINE implement task 3" flow-high opus)"
expect_deny "2b hand-picked flow-high" "reserved for the no-progress escalation"

run_hook "$HOOK" "$(payload "No-progress escalation
$BASELINE fix round 3" flow-high opus)"
expect_allow "2c no-progress escalation on flow-high"

run_hook "$HOOK" "$(payload "Operator-approved high
$BASELINE panel pass 1" flow-high opus)"
expect_allow "2d operator-approved panel on flow-high"

run_hook "$HOOK" "$(payload "Key: pipeline-fix-x-plan-2
$BASELINE re-plan" flow-high opus)"
expect_allow "2e pipeline-fix loop escalation on flow-high"

run_hook "$HOOK" "$(payload "$VERIFIER" flow-low opus)"
expect_allow "3 verifier on flow-low/opus"

run_hook "$HOOK" "$(payload "$VERIFIER" flow-high opus)"
expect_deny "4 verifier on flow-high" 'subagent_type: \"flow-low\"' 'model: \"opus\"'

run_hook "$HOOK" "$(payload "$VERIFIER" flow-low)"
expect_deny "5 verifier with no model" 'model: \"opus\"'

run_hook "$HOOK" "$(payload "$ANALYST" flow-medium opus)"
expect_allow "6 analyst on flow-medium/opus"

run_hook "$HOOK" "$(payload "$ANALYST" flow-high opus)"
expect_deny "7 analyst on flow-high" 'subagent_type: \"flow-medium\"'

run_hook "$SANDBOX/.zcode/hooks/enforce-agent-baseline.py" "$(payload "$VERIFIER" flow-high glm-5.3-flash)"
expect_allow "8 zcode verifier on its mapped pair"

run_hook "$SANDBOX/.zcode/hooks/enforce-agent-baseline.py" "$(payload "$VERIFIER" flow-low opus)"
expect_deny "9 zcode verifier on the claude pair" 'model: \"glm-5.3-flash\"'

run_hook "$HOOK" "$(payload "$VERIFIER" fork)"
expect_allow "10 fork exempt"

run_hook "$HOOK" '{not json'
expect_allow "11 malformed JSON"

run_hook "$HOOK" "$(payload "Key: self-review-x-plan
$BASELINE Finding 2 cites /x/skills/flow/visual-verify-verifier.md:120." flow-medium opus)"
expect_allow "12 a planner citing the verifier file, no marker line"

run_hook "$HOOK" "$(payload "$BASELINE Task 4 re-reads .superpowers/sdd/tooling-analysis-1.md and step 8 of skills/flow/visual-verify-verifier.md." flow-low sonnet)"
expect_allow "12b an implementer citing both files, no marker line"

if [ "$FAILURES" -ne 0 ]; then
  printf 'FAIL: test-enforce-agent-baseline-hook.sh (%d)\n' "$FAILURES" >&2
  exit 1
fi
printf 'PASS: test-enforce-agent-baseline-hook.sh\n'
