package guard

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// runHookInProcess runs `flow-guard hook <name>` in-process on stdin, with the store at a dead
// port and the incident write stubbed, so no case reads or writes the dev store.
func runHookInProcess(t *testing.T, name, stdin string, vars ...string) (string, int) {
	t.Helper()
	out, _, rc := runHookStreams(t, name, stdin, vars...)
	return out, rc
}

// runHookStreams is runHookInProcess returning the hook's stderr as well.
func runHookStreams(t *testing.T, name, stdin string, vars ...string) (string, string, int) {
	t.Helper()
	m := map[string]string{"FLOW_ADDR": "http://127.0.0.1:1", "FLOW_HOOK_INCIDENT_CMD": "true", "PATH": os.Getenv("PATH")}
	for _, kv := range vars {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	var out, errb bytes.Buffer
	env := Env{Getenv: func(k string) string { return m[k] }, Dir: t.TempDir(), Stdin: strings.NewReader(stdin)}
	rc := Registry["hook"]([]string{name}, env, &out, &errb)
	return out.String(), errb.String(), rc
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

// TestHookEnforceLoopDispatchBytes pins what hooks/enforce-loop-dispatch.py logged before its
// port, byte for byte: the checker's arguments, and on its non-zero exit the line naming the key
// and exit code followed by the checker's stdout then stderr; nothing on stdout either way.
func TestHookEnforceLoopDispatchBytes(t *testing.T) {
	t.Parallel()
	check := filepath.Join(t.TempDir(), "chk")
	if err := os.WriteFile(check, []byte("#!/bin/sh\necho \"out:$*\"\necho err >&2\nexit 2\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	stdin := `{"tool_name":"Agent","tool_input":{"prompt":"Key: self-review-a-b\nRole: implementer\n~/.claude/rules/agent-baseline.md","subagent_type":"flow-low","model":"opus"}}`
	out, errOut, rc := runHookStreams(t, "enforce-loop-dispatch", stdin, "LOOP_DISPATCH_CHECK="+check, "FLOW_GUARD_REPO_ROOT=/repo")
	want := "enforce-loop-dispatch: self-review-a-b not recorded (flow-guard check-loop-dispatch exit 2):\nout:/repo self-review-a-b flow-low opus implementer\nerr\n\n"
	if rc != 0 || out != "" || errOut != want {
		t.Errorf("rc=%d out=%q\n got: %q\nwant: %q", rc, out, errOut, want)
	}
}

// TestHookFlowSessionActivityBytes pins the PUTs hooks/flow-session-activity.py sent before its
// port, path and body byte for byte: the session id quoted with no safe character, Python's
// json.dumps spacing and \u escaping, the detail cut to 500 UTF-8 bytes on a character boundary,
// and nothing for a subagent's tool event.
func TestHookFlowSessionActivityBytes(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		defer mu.Unlock()
		got = append(got, r.Method+" "+r.RequestURI+" "+r.Header.Get("Content-Type")+" "+string(b))
	}))
	defer srv.Close()
	for _, stdin := range []string{
		`{"session_id":"a b/c","hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"h\u00e9llo \u2014 x\nsecond"}}`,
		`{"session_id":"s","hook_event_name":"PreToolUse","tool_name":"AskUserQuestion","tool_input":{"questions":[{"question":"Pick \ud83d\ude00?"}]}}`,
		`{"session_id":"s","hook_event_name":"PostToolUse","tool_name":"Edit","tool_input":{"file_path":"/x/y/z.go"}}`,
		`{"session_id":"s","hook_event_name":"Notification","message":"Claude needs your Permission"}`,
		`{"session_id":"s","hook_event_name":"Stop"}`,
		`{"session_id":"s","hook_event_name":"PreToolUse","tool_name":"Read","agent_id":"x"}`,
		`{"session_id":"s","hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"description":"` + strings.Repeat(`\u00e9`, 300) + `"}}`,
	} {
		if out, rc := runHookInProcess(t, "flow-session-activity", stdin, "FLOW_ADDR="+srv.URL); rc != 0 || out != "" {
			t.Errorf("rc=%d out=%q", rc, out)
		}
	}
	const put = "PUT /api/v1/sessions/"
	want := []string{
		put + `a%20b%2Fc/activity application/json {"state": "working", "detail": "Bash: h\u00e9llo \u2014 x"}`,
		put + `s/activity application/json {"state": "question", "detail": "Pick \ud83d\ude00?"}`,
		put + `s/activity application/json {"state": "working", "detail": "Edit z.go"}`,
		put + `s/activity application/json {"state": "permission", "detail": "Claude needs your Permission"}`,
		put + `s/activity application/json {"state": "idle", "detail": ""}`,
		put + `s/activity application/json {"state": "working", "detail": "Bash: ` + strings.Repeat(`\u00e9`, 247) + `"}`,
	}
	mu.Lock()
	defer mu.Unlock()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}
