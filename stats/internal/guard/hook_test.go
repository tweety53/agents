package guard

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/records"
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
	env := core.Env{Getenv: func(k string) string { return m[k] }, Dir: t.TempDir(), Stdin: strings.NewReader(stdin)}
	rc := Registry["hook"]([]string{name}, env, &out, &errb)
	return out.String(), errb.String(), rc
}

// TestHookEnforceAgentBaselineBytes pins the hook's deny output byte for byte against what
// hooks/enforce-agent-baseline.py printed before its port, written by hookJSON (encoding/json, no
// HTML escaping), and no trailing newline.
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
			`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Dispatch blocked (pointer-missing): the subagent prompt does not carry the agent-baseline pointer.\n\nA dispatched agent starts from an empty context. It has never read ~/.claude/CLAUDE.md, and neither will anything it dispatches. Without the pointer it operates under no rules at all — production access, lint suppression, pushes to main, dependency versions.\n\nRe-send this call with the following two sentences included verbatim in ` + "`prompt`" + `:\n\nBefore anything else, read ` + "`~/.claude/rules/agent-baseline.md`" + ` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent.\n\nThe second sentence is not optional: it is what carries the rules past depth 1.\nExempt: subagent_type \"fork\" outside a flow session, which inherits this session's context already."}}`},
		{"fork naming a flow skill", payload(map[string]any{"prompt": "read skills/flow/x", "subagent_type": "fork"}, false),
			`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Dispatch blocked (fork-refused): no flow phase, and no loop a phase or the self-review runs, dispatches subagent_type \"fork\" (skills/flow/implement.md, KAN-1000).\n\nRe-send this call as a fresh agent on the pair its contract names, its prompt carrying what it needs -- paths to the run's artifacts, reports and records -- never this session's context."}}`},
		{"no effort in a flow session", payload(map[string]any{"prompt": baseline, "subagent_type": "general-purpose"}, true),
			`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Dispatch blocked (effort-missing): ` + "`subagent_type: general-purpose`" + ` carries no effort and the call sets none, so it would run at the model's built-in effort.\n\nRun ` + "`flow-guard pair-judge <role> <prompt file>`" + ` and re-send on the pair it prints (` + "`skills/flow-contracts/model-policy.md`" + `, **Model and effort**): a ` + "`flow-<effort>`" + ` type, or an ` + "`Explore`/`Plan`" + ` type with the pair's ` + "`effort`" + ` set on the call."}}`},
		{"loop key without its Key line", payload(map[string]any{"prompt": baseline, "description": "pipeline-fix-x go", "subagent_type": "flow-low"}, true),
			`{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Dispatch blocked (loop-key-line-missing): ` + "`pipeline-fix-x`" + ` is a loop key — its prompt carries ` + "`Key: pipeline-fix-x`" + ` as a line of its own, and flow-guard hook enforce-loop-dispatch records its row under the loop's change off that line. Add the line and re-send; never ` + "`flow record dispatch begin`" + ` it under the run's change."}}`},
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
	want := `{"decision":"block","reason":"Agents you launched are still running (c1). Stop again to keep waiting for them -- ending your turn on your own would lose your 5-minute prompt cache."}` + "\n"
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
// port, path and body byte for byte: the session id quoted with no safe character, the body
// written by hookJSON, the detail cut to 500 UTF-8 bytes on a character boundary,
// and a subagent's tool event marked "subagent": true, so flowd lets it only end a permission wait.
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
		put + `a%20b%2Fc/activity application/json {"state":"working","detail":"Bash: héllo — x"}`,
		put + `s/activity application/json {"state":"question","detail":"Pick 😀?"}`,
		put + `s/activity application/json {"state":"working","detail":"Edit z.go"}`,
		put + `s/activity application/json {"state":"permission","detail":"Claude needs your Permission"}`,
		put + `s/activity application/json {"state":"idle","detail":""}`,
		put + `s/activity application/json {"state":"working","detail":"Read","subagent":true}`,
		put + `s/activity application/json {"state":"working","detail":"Bash: ` + strings.Repeat(`é`, 247) + `"}`,
	}
	mu.Lock()
	defer mu.Unlock()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}

// TestHookFlowActiveChangeBytes pins what hooks/flow-active-change.py asked flowd and printed before
// its port, byte for byte: the stage-runs query (urlencode's quote_plus), a loop's own run skipped,
// the session id from a transcript's stem, and the context line written by hookJSON with no
// trailing newline.
func TestHookFlowActiveChangeBytes(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ stdin, answer, query, want string }{
		{`{"prompt":"hi","session_id":"a b/c+&"}`,
			`{"stageRuns":[{"changeName":"self-review-x"},{"changeName":"kan-1","command":"/flow","stage":"flow.implement"}]}`,
			"/api/v1/stage-runs?session_id=a+b%2Fc%2B%26&sort=-started_at&limit=50",
			`{"hookSpecificOutput":{"additionalContext":"flow: this session's last /flow run marked change kan-1. A plain message reporting a problem or asking for a change to it is a fix run of kan-1 — **A plain message at IN_PROGRESS** (skills/flow/SKILL.md).","hookEventName":"UserPromptSubmit"}}`},
		{`{"prompt":"hi","transcript_path":"/t/sess-9.jsonl"}`,
			`{"stageRuns":[{"changeName":"kan-2","command":"/flow-interactive","stage":"flow.verify"}]}`,
			"/api/v1/stage-runs?session_id=sess-9&sort=-started_at&limit=50",
			"{\"hookSpecificOutput\":{\"additionalContext\":\"flow: this session runs /flow-interactive change `kan-2`. A plain message is its next request — or its stop (**One request**, **Explicit stop**, `skills/flow-interactive/SKILL.md`).\",\"hookEventName\":\"UserPromptSubmit\"}}"},
		{`{"prompt":"hi","session_id":"s"}`,
			`{"stageRuns":[{"changeName":"kan-3","command":"/flow-interactive","stage":"flow.cleanup"}]}`,
			"/api/v1/stage-runs?session_id=s&sort=-started_at&limit=50", ""},
	} {
		var mu sync.Mutex
		var asked string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			asked = r.RequestURI
			mu.Unlock()
			_, _ = io.WriteString(w, c.answer)
		}))
		out, rc := runHookInProcess(t, "flow-active-change", c.stdin, "FLOW_ADDR="+srv.URL)
		srv.Close()
		mu.Lock()
		if rc != 0 || out != c.want || asked != c.query {
			t.Errorf("rc=%d asked %q\n got: %q\nwant: %q", rc, asked, out, c.want)
		}
		mu.Unlock()
	}
}

// TestHookRecordsItsWallTime: with Env.HookRuns naming a file, each hook run
// appends one records.HookRun line -- the hook's name, its wall time in
// microseconds, its end -- and a fault the hook recovers from is still
// recorded.
func TestHookRecordsItsWallTime(t *testing.T) {
	t.Parallel()
	log := filepath.Join(t.TempDir(), "hook-runs.jsonl")
	env := core.Env{Getenv: func(string) string { return "" }, Dir: t.TempDir(), HookRuns: log}
	before := time.Now()
	for _, stdin := range []string{`{"tool_name":"Read"}`, `not json`} {
		env.Stdin = strings.NewReader(stdin)
		if rc := Registry["hook"]([]string{"enforce-agent-baseline"}, env, io.Discard, io.Discard); rc != 0 {
			t.Fatalf("rc %d", rc)
		}
	}
	body, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(body), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %q, want one per run", lines)
	}
	for _, l := range lines {
		var r records.HookRun
		if err := json.Unmarshal([]byte(l), &r); err != nil || r.Hook != "enforce-agent-baseline" || r.DurationUS < 0 || r.At.Before(before) {
			t.Errorf("line %q (%v), want the hook, a duration and its end", l, err)
		}
	}

}

// TestHookAwaitOwnAgentsOwedHandback: plain text from an agent that has not yet called
// SubagentHandback is no final report, so its running background command is waited for, not killed; once it has called it, the stop goes through.
func TestHookAwaitOwnAgentsOwedHandback(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	output := filepath.Join(dir, "b1.output")
	f, err := os.Create(output)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	build := exec.Command("sleep", "60")
	build.Stdout = f
	if err := build.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = build.Process.Kill(); _ = build.Wait() })
	transcript := filepath.Join(dir, "agent-p.jsonl")
	lines := `{"type":"user","message":{"content":"<system-reminder>\nYour final report is delivered through SubagentHandback: call it."}}
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash"}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"t1","content":"Command running in background with ID: b1. Output is being written to: ` + output + `"}]}}
`
	if err := os.WriteFile(transcript, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	out, rc := runHookInProcess(t, "await-own-agents", `{"agent_id":"p","agent_transcript_path":"`+transcript+`","last_assistant_message":"Waiting for the run."}`)
	if rc != 0 || !strings.Contains(out, "still running (b1)") {
		t.Errorf("rc=%d out=%q, want a block naming b1", rc, out)
	}
	lines += `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t2","name":"SubagentHandback"}]}}
`
	if err := os.WriteFile(transcript, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	out, rc = runHookInProcess(t, "await-own-agents", `{"agent_id":"p","agent_transcript_path":"`+transcript+`","last_assistant_message":"Done."}`)
	if rc != 0 || out != "" {
		t.Errorf("rc=%d out=%q, want the stop let through after the handback", rc, out)
	}
}
