package guard

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runHookInProcess runs `flow-guard hook <name>` in-process on stdin, with the store at a dead
// port and the incident write stubbed, so no case reads or writes the dev store.
func runHookInProcess(t *testing.T, name, stdin string, vars ...string) (string, int) {
	t.Helper()
	m := map[string]string{"FLOW_ADDR": "http://127.0.0.1:1", "FLOW_HOOK_INCIDENT_CMD": "true", "PATH": os.Getenv("PATH")}
	for _, kv := range vars {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	var out, errb bytes.Buffer
	env := Env{Getenv: func(k string) string { return m[k] }, Dir: t.TempDir(), Stdin: strings.NewReader(stdin)}
	rc := Registry["hook"]([]string{name}, env, &out, &errb)
	return out.String(), rc
}

// TestHookEnforceAgentBaselineBytes pins the hook's deny output byte for byte against what
// hooks/enforce-agent-baseline.py printed before its port: Python's json.dump spacing, its
// \u escaping of every non-ASCII character, and no trailing newline.
func TestHookEnforceAgentBaselineBytes(t *testing.T) {
	t.Parallel()
	const baseline = "Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent."
	transcript := filepath.Join(t.TempDir(), "t.jsonl")
	if err := os.WriteFile(transcript, []byte(`{"type":"user","message":{"content":"<command-name>/flow</command-name>"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := func(in map[string]any, flow bool) string {
		p := map[string]any{"tool_name": "Agent", "tool_input": in}
		if flow {
			p["transcript_path"] = transcript
		}
		b, _ := json.Marshal(p)
		return string(b)
	}
	cases := []struct{ name, stdin, want string }{
		{"no pointer", payload(map[string]any{"prompt": "do <it> & go", "subagent_type": "flow-high"}, false),
			`{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny", "permissionDecisionReason": "Dispatch blocked: the subagent prompt does not carry the agent-baseline pointer.\n\nA dispatched agent starts from an empty context. It has never read ~/.claude/CLAUDE.md, and neither will anything it dispatches. Without the pointer it operates under no rules at all \u2014 production access, lint suppression, pushes to main, dependency versions.\n\nRe-send this call with the following two sentences included verbatim in ` + "`prompt`" + `:\n\nBefore anything else, read ` + "`~/.claude/rules/agent-baseline.md`" + ` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent.\n\nThe second sentence is not optional: it is what carries the rules past depth 1.\nExempt: subagent_type \"fork\" outside a flow session, which inherits this session's context already."}}`},
		{"fork naming a flow skill", payload(map[string]any{"prompt": "read skills/flow/x", "subagent_type": "fork"}, false),
			`{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny", "permissionDecisionReason": "Dispatch blocked: no flow phase, and no loop a phase or the self-review runs, dispatches subagent_type \"fork\" (skills/flow/implement.md, KAN-1000).\n\nRe-send this call as a fresh agent on the pair its contract names, its prompt carrying what it needs -- paths to the run's artifacts, reports and records -- never this session's context."}}`},
		{"no effort in a flow session", payload(map[string]any{"prompt": baseline, "subagent_type": "general-purpose"}, true),
			`{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny", "permissionDecisionReason": "Dispatch blocked: ` + "`subagent_type: general-purpose`" + ` carries no effort and the call sets none, so it would run at the model's built-in effort.\n\nRun ` + "`flow-guard pair-judge <role> <prompt file>`" + ` and re-send on the pair it prints (` + "`skills/flow-contracts/model-policy.md`" + `, **Model and effort**): a ` + "`flow-<effort>`" + ` type, or an ` + "`Explore`/`Plan`" + ` type with the pair's ` + "`effort`" + ` set on the call."}}`},
		{"loop key without its Key line", payload(map[string]any{"prompt": baseline, "description": "pipeline-fix-x go", "subagent_type": "flow-low"}, true),
			`{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny", "permissionDecisionReason": "Dispatch blocked: ` + "`pipeline-fix-x`" + ` is a loop key \u2014 its prompt carries ` + "`Key: pipeline-fix-x`" + ` as a line of its own, and hooks/enforce-loop-dispatch.py records its row under the loop's change off that line. Add the line and re-send; never ` + "`flow record dispatch begin`" + ` it under the run's change."}}`},
		{"pointer outside a flow session", payload(map[string]any{"prompt": baseline, "subagent_type": "flow-low"}, false), ""},
		{"no JSON", "{not json", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			out, rc := runHookInProcess(t, "enforce-agent-baseline", c.stdin)
			if rc != 0 || out != c.want {
				t.Errorf("rc=%d\n got: %s\nwant: %s", rc, out, c.want)
			}
		})
	}
}

// TestHookUnknownName: an unknown hook name is a usage error, exit 2.
func TestHookUnknownName(t *testing.T) {
	t.Parallel()
	if _, rc := runHookInProcess(t, "no-such-hook", "{}"); rc != 2 {
		t.Errorf("rc=%d, want 2", rc)
	}
}

// TestHookAwaitOwnAgentsBytes pins the block hooks/await-own-agents.py printed before its port,
// byte for byte, trailing newline included: a child still running at a zero wait.
func TestHookAwaitOwnAgentsBytes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "agent-c1.meta.json"), []byte(`{"parentAgentId":"p","finished":false}`), 0o644); err != nil {
		t.Fatal(err)
	}
	stdin := `{"agent_id":"p","agent_transcript_path":"` + filepath.Join(dir, "agent-p.jsonl") + `"}`
	out, rc := runHookInProcess(t, "await-own-agents", stdin, "AWAIT_CHILDREN_SECONDS=0")
	want := `{"decision": "block", "reason": "Agents you launched are still running (c1). Stop again to keep waiting for them -- ending your turn on your own would lose your 5-minute prompt cache."}` + "\n"
	if rc != 0 || out != want {
		t.Errorf("rc=%d\n got: %q\nwant: %q", rc, out, want)
	}
}
