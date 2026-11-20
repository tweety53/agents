package repotest

// TestEnforceAgentBaselineHook ports the retired shell harness,
// the harness for hooks/enforce-agent-baseline.py. It pipes PreToolUse payloads
// to the hook and asserts its decision, one subtest per harness case:
//
//	 1. no baseline pointer                                 -> deny (baseline reason)
//	 2. pointer, outside a flow session, any type/model     -> allow
//	10. a fork dispatch naming a flow skill                 -> deny (names a flow skill)
//	10b. a fork naming no flow skill, no transcript         -> allow
//	10c. the same fork in a session whose transcript records
//	     a /flow command                                    -> deny
//	10d. the same fork, transcript recording only /flow-status -> allow
//	10e. the same fork, the /flow tag only inside a tool result -> allow
//	11. malformed JSON                                      -> allow, exit 0
//	12. prompts outside a flow session, no Role line        -> allow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// Cost: thirteen runs of the real hook, each a python3 start; the shared
// flow-guard build is TestMain's, not this test's. Cutting a run drops a
// dispatch shape the gate proves.
// smoke: core — the hook that gates every dispatch, through the real guard and CLI
func TestEnforceAgentBaselineHook(t *testing.T) {
	t.Parallel()
	const baseline = "Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent."
	const verifier = "Role: visual-verifier\n" + baseline + " Read /x/skills/flow/visual-verify-verifier.md and run its steps."

	sandbox := t.TempDir()
	transcript := func(name, line string) string {
		p := filepath.Join(sandbox, name)
		if err := os.WriteFile(p, []byte(line+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	flowT := transcript("flow.jsonl", `{"type":"user","message":{"content":"<command-name>/flow</command-name>"}}`)
	statusT := transcript("status.jsonl", `{"type":"user","message":{"content":"<command-name>/flow-status</command-name>"}}`)
	interactiveT := transcript("interactive.jsonl", `{"type":"user","message":{"content":"<command-name>/flow-interactive</command-name>"}}`)
	toolT := transcript("tool.jsonl", `{"type":"user","message":{"content":[{"type":"tool_result","content":"<command-name>/flow</command-name>"}]}}`)
	skillT := transcript("skill.jsonl", `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Skill","input":{"skill":"flow-fast"}}]}}`)
	otherSkillT := transcript("other-skill.jsonl", `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Skill","input":{"skill":"flow-status"}}]}}`)

	// payload builds an Agent payload; model "" omits that field, transcript "" omits
	// transcript_path.
	payload := func(prompt, typ, model, transcript string) string {
		ti := map[string]string{"prompt": prompt, "subagent_type": typ}
		if model != "" {
			ti["model"] = model
		}
		p := map[string]any{"tool_name": "Agent", "tool_input": ti}
		if transcript != "" {
			p["transcript_path"] = transcript
		}
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	const hook = "enforce-agent-baseline.py"
	cases := []struct {
		name, hook, stdin string
		deny              []string // nil: allow; else deny, reason carrying each
	}{
		{"1 no pointer", hook, payload("do the thing", "flow-high", "opus", ""), []string{"agent-baseline pointer"}},
		{"2 outside a flow session", hook, payload(baseline+" implement task 3", "flow-medium", "sonnet", ""), nil},
		{"10 fork naming a flow skill", hook, payload(verifier, "fork", "", ""), []string{`subagent_type \"fork\"`}},
		{"10b fork outside a flow session", hook, payload("summarise the thread above", "fork", "", ""), nil},
		{"10c fork in a /flow session", hook, payload("run the loop above", "fork", "", flowT), []string{"KAN-1000"}},
		{"10d fork after /flow-status alone", hook, payload("run the loop above", "fork", "", statusT), nil},
		{"10e fork whose transcript holds the tag only in a tool result", hook, payload("run the loop above", "fork", "", toolT), nil},
		{"10f fork in a run the model started through the Skill tool", hook, payload("run the loop above", "fork", "", skillT), []string{"KAN-1000"}},
		{"10g fork after the Skill tool ran /flow-status alone", hook, payload("run the loop above", "fork", "", otherSkillT), nil},
		{"10h fork in a /flow-interactive session", hook, payload("run the loop above", "fork", "", interactiveT), []string{"KAN-1000"}},
		{"11 malformed JSON", hook, "{not json", nil},
		{"12 a planner citing the verifier file, no marker line", hook, payload("Key: self-review-x-plan\n"+baseline+" Finding 2 cites /x/skills/flow/visual-verify-verifier.md:120.", "flow-medium", "opus", ""), nil},
		{"12b an implementer citing both files, no marker line", hook, payload(baseline+" Task 4 re-reads .superpowers/sdd/tooling-analysis-1.md and step 8 of skills/flow/visual-verify-verifier.md.", "flow-low", "sonnet", ""), nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			out, _, rc := runHook(t, c.hook, []byte(c.stdin))
			out = strings.TrimRight(out, "\n")
			if c.deny == nil {
				if rc != 0 || out != "" {
					t.Errorf("rc=%d out=%s", rc, out)
				}
				return
			}
			if rc != 0 || !strings.Contains(out, `"permissionDecision": "deny"`) {
				t.Fatalf("rc=%d out=%s", rc, out)
			}
			for _, s := range c.deny {
				if !strings.Contains(out, s) {
					t.Errorf("reason lacks %s (out=%s)", s, out)
				}
			}
		})
	}
}

// openRowFixture serves a store answering /api/v1/stage-runs with change c1, session token mf-t1
// and stage flow.plan -- or the contents of a `stage` file beside the called log, when a
// test writes one -- (status 500 when fail), and a fake `flow` first on PATH that appends its argv, one line
// per call, to the returned called file. It answers `record dispatches` with rows (exiting 1 when
// rows is ""), `record dispatch begin` with exit 0, and every other call with exit 1. It returns the env the hook runs
// under and that argv log.
func openRowFixture(t *testing.T, rows string, fail bool, projectKey ...string) (env []string, called string) {
	t.Helper()
	bin := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail || r.URL.Path != "/api/v1/stage-runs" || r.URL.Query().Get("session_id") != "s1" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stage := "flow.plan"
		if b, err := os.ReadFile(filepath.Join(bin, "stage")); err == nil {
			stage = string(b)
		}
		run := map[string]string{"changeName": "c1", "sessionToken": "mf-t1", "stage": stage}
		if len(projectKey) > 0 {
			run["projectKey"] = projectKey[0]
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"stageRuns": []any{run}})
	}))
	t.Cleanup(srv.Close)
	called = filepath.Join(bin, "called")
	rowsAnswer := "exit 1"
	if rows != "" {
		if err := os.WriteFile(filepath.Join(bin, "rows.json"), []byte(rows), 0o644); err != nil {
			t.Fatal(err)
		}
		rowsAnswer = "cat " + filepath.Join(bin, "rows.json")
	}
	script := "printf '%s\\n' \"$*\" >> " + called + "\ncase \"$*\" in\n" +
		"'record dispatches'*) " + rowsAnswer + " ;;\n" +
		"'record dispatch begin'*) exit 0 ;;\n" +
		"*) exit 1 ;;\nesac\n"
	writeStub(t, filepath.Join(bin, "flow"), script)
	// FLOW_HOOK_UNBOUNDED lifts the hook's wall-clock bounds: under the full -race suite a stub
	// starved past them failed a deny case open (kan-1033).
	return []string{"FLOW_ADDR=" + srv.URL, "FLOW_HOOK_UNBOUNDED=1", "PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH")}, called
}

// openRowPayload is a flow-session Agent payload for session s1 carrying the baseline pointer.
func openRowPayload(t *testing.T, description, prompt string) []byte {
	t.Helper()
	tr := filepath.Join(t.TempDir(), "s1.jsonl")
	if err := os.WriteFile(tr, []byte(`{"type":"user","message":{"content":"<command-name>/flow</command-name>"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(map[string]any{
		"tool_name":       "Agent",
		"session_id":      "s1",
		"cwd":             t.TempDir(),
		"transcript_path": tr,
		"tool_input": map[string]string{
			"description":   description,
			"prompt":        "Role: panel-reviewer\n" + prompt + "\nBefore anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent.",
			"subagent_type": "flow-low",
			"model":         "opus",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestEnforceAgentBaselineHookAllowsAnotherRepoCwd pins the fail-open on a cwd outside the
// change's project: its rows read empty there, so the hook lets the dispatch through and begins
// no row.
func TestEnforceAgentBaselineHookAllowsAnotherRepoCwd(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	env, called := openRowFixture(t, `[]`, false, "gymie-7c1f238a")
	var payload map[string]any
	if err := json.Unmarshal(openRowPayload(t, "panel-4-primary review", "review"), &payload); err != nil {
		t.Fatal(err)
	}
	payload["cwd"] = repo
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	out, _, rc := runHook(t, "enforce-agent-baseline.py", b, env...)
	if rc != 0 || out != "" {
		t.Fatalf("rc=%d out=%s", rc, out)
	}
	if begins := flowCalls(t, called, "record dispatch begin"); len(begins) != 0 {
		t.Errorf("begun %v", begins)
	}
}

// TestEnforceAgentBaselineHookAsksForTheLoopKeyLine pins kan-1032's self-review: a flow dispatch
// whose description opens with a loop key and whose prompt carries no `Key:` line is told to add
// the line, never to begin a row under the run's change, and the store is never read for it.
func TestEnforceAgentBaselineHookAsksForTheLoopKeyLine(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, description, prompt string
		deny                      bool
	}{
		{"pipeline-fix key, no line", "pipeline-fix-7-plan x", "plan", true},
		{"self-review key, no line", "self-review-x-plan y", "plan", true},
		{"pipeline-fix key with its line", "pipeline-fix-7-plan x", "Key: pipeline-fix-7-plan", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			env, called := openRowFixture(t, `[]`, false)
			out, _, rc := runHook(t, "enforce-agent-baseline.py", openRowPayload(t, c.description, c.prompt), env...)
			denied := strings.Contains(out, `"permissionDecision": "deny"`)
			if rc != 0 || denied != c.deny {
				t.Fatalf("rc=%d deny=%v want %v out=%s", rc, denied, c.deny, out)
			}
			if _, err := os.Stat(called); err == nil {
				t.Errorf("flow was called for a loop-key dispatch")
			}
			if !c.deny {
				return
			}
			word := strings.Fields(c.description)[0]
			for _, s := range []string{"Key: " + word, "enforce-loop-dispatch.py"} {
				if !strings.Contains(out, s) {
					t.Errorf("reason lacks %q: %s", s, out)
				}
			}
			if strings.Contains(out, "flow record dispatch begin -change") {
				t.Errorf("reason asks for a begin under the run's change: %s", out)
			}
		})
	}
}

// TestEnforceAgentBaselineHookSkipsAFinishedChange pins the flow-run checks' scope: in a session
// whose change the store reports FINISHED, a loop-key dispatch with no Key line passes; the same
// dispatch while the change is IN_PROGRESS is denied.
func TestEnforceAgentBaselineHookSkipsAFinishedChange(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		state string
		deny  bool
	}{{"FINISHED", false}, {"IN_PROGRESS", true}} {
		t.Run(c.state, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v1/stage-runs":
					_ = json.NewEncoder(w).Encode(map[string]any{"stageRuns": []any{map[string]string{"changeName": "c1", "projectKey": "p"}}})
				case "/api/v1/changes/p/c1":
					_ = json.NewEncoder(w).Encode(map[string]string{"name": "c1", "state": c.state})
				default:
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			t.Cleanup(srv.Close)
			out, _, rc := runHook(t, "enforce-agent-baseline.py", openRowPayload(t, "self-review-x-plan y", "plan"),
				"FLOW_ADDR="+srv.URL, "FLOW_HOOK_UNBOUNDED=1")
			if denied := strings.Contains(out, `"permissionDecision": "deny"`); rc != 0 || denied != c.deny {
				t.Fatalf("rc=%d deny=%v want %v out=%s", rc, denied, c.deny, out)
			}
		})
	}
}

// TestEnforceAgentBaselineHookRecordsADenial pins decision sr4-hook-denial-incident: each deny
// writes one `flow record incident` call, with the session's change where the store names one;
// an allowed dispatch writes none; a failing write leaves the deny JSON as the whole stdout.
func TestEnforceAgentBaselineHookRecordsADenial(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	noPointer, err := json.Marshal(map[string]any{
		"tool_name":  "Agent",
		"cwd":        t.TempDir(),
		"tool_input": map[string]string{"description": "x", "prompt": "do it", "subagent_type": "flow-low", "model": "opus"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run("missing pointer", func(t *testing.T) {
		cmd, log := incidentStub(t, 0)
		out, _, rc := runHook(t, "enforce-agent-baseline.py", noPointer, "FLOW_HOOK_INCIDENT_CMD="+cmd)
		if rc != 0 || !strings.Contains(out, `"permissionDecision": "deny"`) {
			t.Fatalf("rc=%d out=%s", rc, out)
		}
		calls := incidentCalls(t, log)
		if len(calls) != 1 {
			t.Fatalf("incident calls = %d, want 1: %v", len(calls), calls)
		}
		c := calls[0]
		if strings.Join(c[:2], " ") != "record incident" || argAfter(c, "-guard") != "enforce-agent-baseline" || argAfter(c, "-minutes-lost") != "0" {
			t.Errorf("argv = %v", c)
		}
		if sym := argAfter(c, "-symptom"); sym == "" || !strings.Contains(out, sym[:20]) {
			t.Errorf("-symptom %q does not open the deny reason %s", sym, out)
		}
	})
	t.Run("loop key deny carries the change", func(t *testing.T) {
		cmd, log := incidentStub(t, 0)
		env, _ := openRowFixture(t, `[]`, false)
		out, _, rc := runHook(t, "enforce-agent-baseline.py", openRowPayload(t, "self-review-x-plan y", "plan"), append(env, "FLOW_HOOK_INCIDENT_CMD="+cmd)...)
		if rc != 0 || !strings.Contains(out, `"permissionDecision": "deny"`) {
			t.Fatalf("rc=%d out=%s", rc, out)
		}
		calls := incidentCalls(t, log)
		if len(calls) != 1 || argAfter(calls[0], "-change") != "c1" {
			t.Errorf("calls = %v, want one carrying -change c1", calls)
		}
	})
	t.Run("allowed dispatch", func(t *testing.T) {
		cmd, log := incidentStub(t, 0)
		env, _ := openRowFixture(t, `[{"key":"panel-4-primary","outcome":""}]`, false)
		out, _, rc := runHook(t, "enforce-agent-baseline.py", openRowPayload(t, "panel-4-primary review", "review"), append(env, "FLOW_HOOK_INCIDENT_CMD="+cmd)...)
		if rc != 0 || out != "" {
			t.Fatalf("rc=%d out=%s, want an allow", rc, out)
		}
		if calls := incidentCalls(t, log); len(calls) != 0 {
			t.Errorf("an allowed dispatch logged %v", calls)
		}
	})
	t.Run("failing write", func(t *testing.T) {
		cmd, _ := incidentStub(t, 1)
		out, _, rc := runHook(t, "enforce-agent-baseline.py", noPointer, "FLOW_HOOK_INCIDENT_CMD="+cmd)
		var v map[string]any
		if rc != 0 || json.Unmarshal([]byte(out), &v) != nil || !strings.Contains(out, `"permissionDecision": "deny"`) {
			t.Errorf("rc=%d stdout=%q, want the deny JSON alone", rc, out)
		}
	})
}

// flowCalls returns the argv lines the fake `flow` logged to log that open
// with prefix; a log never written holds none.
func flowCalls(t *testing.T, log, prefix string) []string {
	t.Helper()
	b, err := os.ReadFile(log)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	for _, line := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
		if strings.HasPrefix(line, prefix) {
			calls = append(calls, line)
		}
	}
	return calls
}

// beginRepo is a main checkout the hook derives the change's worktree from: it returns the
// checkout, the payload's cwd, and its change worktree path <checkout>-worktrees/c1.
func beginRepo(t *testing.T) (main, wt string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	main = filepath.Join(root, "main")
	if out, err := exec.Command("git", "init", "-q", main).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	return main, main + "-worktrees/c1"
}

// beginEnv is openRowFixture's env: no row open beyond rows, no incident.
func beginEnv(t *testing.T, rows string) (env []string, log string) {
	t.Helper()
	return openRowFixture(t, rows, false)
}

// beginPayload is a /flow-session Agent payload for session s1 run from cwd, its prompt carrying
// the baseline pointer; effort "" omits the field.
func beginPayload(t *testing.T, cwd, description, prompt, typ, model, effort string) []byte {
	t.Helper()
	tr := filepath.Join(t.TempDir(), "s1.jsonl")
	if err := os.WriteFile(tr, []byte(`{"type":"user","message":{"content":"<command-name>/flow</command-name>"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ti := map[string]string{
		"description":   description,
		"prompt":        prompt + "\nBefore anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent.",
		"subagent_type": typ,
		"model":         model,
	}
	if effort != "" {
		ti["effort"] = effort
	}
	b, err := json.Marshal(map[string]any{"tool_name": "Agent", "session_id": "s1", "cwd": cwd, "transcript_path": tr, "tool_input": ti})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// runBegin runs the hook on payload under env and reports whether it denied, failing on a
// non-zero exit.
func runBegin(t *testing.T, payload []byte, env ...string) (denied bool, out string) {
	t.Helper()
	out, _, rc := runHook(t, "enforce-agent-baseline.py", payload, env...)
	if rc != 0 {
		t.Fatalf("rc=%d out=%s", rc, out)
	}
	return strings.Contains(out, `"permissionDecision": "deny"`), out
}

// TestEnforceAgentBaselineHookBeginsTheRow pins decisions hook-same-file and
// hook-token-from-stage-run: a flow dispatch with no open row is begun, not denied, with the stage
// run's token, its Role line's role, and the call's model and effort.
func TestEnforceAgentBaselineHookBeginsTheRow(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	main, _ := beginRepo(t)
	want := "record dispatch begin -C " + main + " -change c1 -role panel-fix -model opus -effort low -key panel-fix-2 -session-token mf-t1 -policy-role panel-fix"
	env, log := beginEnv(t, `[]`)
	if denied, out := runBegin(t, beginPayload(t, main, "panel-fix-2 fix", "Role: panel-fix\nfix", "flow-low", "opus", ""), env...); denied {
		t.Fatalf("denied: %s", out)
	}
	if got := flowCalls(t, log, "record dispatch begin"); len(got) != 1 || got[0] != want {
		t.Errorf("begins = %q, want [%q]", got, want)
	}
}

// TestEnforceAgentBaselineHookReadsTheOpenStageRun pins kan-1099 self-review F12: a session whose
// latest stage run is another project's run, already ended, inside its still-open run of change
// c1 has its dispatch begun under c1 -- the open run names the session's change, not the later
// ended one.
func TestEnforceAgentBaselineHookReadsTheOpenStageRun(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	main, _ := beginRepo(t)
	env, log := beginEnv(t, `[]`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"stageRuns": []any{
			map[string]string{"changeName": "pipeline-fix-c1-72", "projectKey": "other", "sessionToken": "pf", "stage": "flow.pipeline-fix", "endedAt": "2026-11-16T19:47:52Z"},
			map[string]string{"changeName": "c1", "sessionToken": "mf-t1", "stage": "flow.sdd-tdd"},
		}})
	}))
	t.Cleanup(srv.Close)
	env = append(env, "FLOW_ADDR="+srv.URL)
	if denied, out := runBegin(t, beginPayload(t, main, "task-52-implementer-fix-3 finish remainder", "Role: implementer\nKey: task-52-implementer-fix-3\nfix", "flow-low", "opus", ""), env...); denied {
		t.Fatalf("denied: %s", out)
	}
	want := "record dispatch begin -C " + main + " -change c1 -role implementer -model opus -effort low -task 52 -key task-52-implementer-fix-3 -session-token mf-t1 -policy-role implementer"
	if got := flowCalls(t, log, "record dispatch begin"); len(got) != 1 || got[0] != want {
		t.Errorf("begins = %q, want [%q]", got, want)
	}
}

// TestEnforceAgentBaselineHookSkipsALoopsOpenStageRun pins decision skip-loop-runs-by-change-prefix:
// a loop's own open stage run, newer than the run's own, never names the session's change. In
// kan-1132's run, task-3-implementer was denied and task-1+2-reviewer begun under
// pipeline-fix-kan-1132-flow-stats-run-timings-costs-dashboard-80 while the run's flow.sdd-tdd
// was open.
func TestEnforceAgentBaselineHookSkipsALoopsOpenStageRun(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct{ prefix, stage string }{
		{"pipeline-fix-", "flow.pipeline-fix"},
		{"self-review-", "flow.self-review"},
	} {
		t.Run(c.prefix, func(t *testing.T) {
			t.Parallel()
			main, _ := beginRepo(t)
			env, log := beginEnv(t, `[]`)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"stageRuns": []any{
					map[string]string{"changeName": c.prefix + "c1-81", "sessionToken": c.prefix + "c1-81", "stage": c.stage},
					map[string]string{"changeName": "c1", "sessionToken": "mf-t1", "stage": "flow.sdd-tdd"},
				}})
			}))
			t.Cleanup(srv.Close)
			env = append(env, "FLOW_ADDR="+srv.URL)
			if denied, out := runBegin(t, beginPayload(t, main, "task-3-implementer finish", "Role: implementer\nKey: task-3-implementer\nfix", "flow-low", "opus", ""), env...); denied {
				t.Fatalf("denied: %s", out)
			}
			want := "record dispatch begin -C " + main + " -change c1 -role implementer -model opus -effort low -task 3 -key task-3-implementer -session-token mf-t1 -policy-role implementer"
			if got := flowCalls(t, log, "record dispatch begin"); len(got) != 1 || got[0] != want {
				t.Errorf("begins = %q, want [%q]", got, want)
			}
		})
	}
}

// TestEnforceAgentBaselineHookBeginsEveryRowRole pins ROW_ROLE: every Role id launchRoleID
// (stats/cmd/flow/record.go) can return, and every panel and planner id the table adds, is begun
// under its row role.
func TestEnforceAgentBaselineHookBeginsEveryRowRole(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for id, role := range map[string]string{
		"implementer": "implementer", "operator-asked": "implementer",
		"gated-reviewer": "reviewer", "gated-re-reviewer": "reviewer",
		"panel-reviewer": "reviewer", "panel-rerun": "reviewer",
		"panel-merger": "merger", "panel-fix": "panel-fix", "minor-fix": "panel-fix", "no-progress-fixer": "panel-fix", "visual-verifier": "verifier", "runner": "verifier",
		"planner": "planner", "planner-short": "planner",
	} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			main, _ := beginRepo(t)
			env, log := beginEnv(t, `[]`)
			if denied, out := runBegin(t, beginPayload(t, main, "panel-fix-3 x", "Role: "+id+"\nwork", "flow-low", "opus", ""), env...); denied {
				t.Fatalf("denied: %s", out)
			}
			begins := flowCalls(t, log, "record dispatch begin")
			if len(begins) != 1 || argAfter(strings.Fields(begins[0]), "-role") != role {
				t.Errorf("begins = %q, want one with -role %s", begins, role)
			}
		})
	}
}

// TestEnforceAgentBaselineHookBeginKeyFromKeyLine pins the key's source: the prompt's `Key:` line
// tells a gated reviewer bundle from its implementer group, both described `Task <ids>/<n>`; a
// task description with no `Key:` line names no row the hook can begin, so it begins none and
// lets the dispatch through.
func TestEnforceAgentBaselineHookBeginKeyFromKeyLine(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name, description, prompt, typ string
		want                           map[string]string // flag -> value; nil: no begin
		noTask                         bool
	}{
		{"gated reviewer", "Task 3/9 (review)", "Role: gated-reviewer\nKey: task-3-reviewer\nreview", "flow-medium",
			map[string]string{"-key": "task-3-reviewer", "-role": "reviewer", "-task": "3", "-effort": "medium"}, false},
		{"implementer group", "Tasks 3+4/9 (x)", "Role: implementer\nKey: task-3+4-implementer\nimplement", "flow-low",
			map[string]string{"-key": "task-3+4-implementer", "-role": "implementer"}, true},
		{"no Key line", "Task 3/9 (x)", "Role: implementer\nimplement", "flow-low", nil, false},
		{"fix implementer", "fix-2-implementer (fix)", "Role: implementer\nKey: fix-2-implementer\nfix", "flow-low",
			map[string]string{"-key": "fix-2-implementer", "-role": "implementer", "-effort": "low"}, true},
		{"fix reviewer", "fix-2-review (review the fix)", "Role: panel-reviewer\nKey: fix-2-review\nreview", "flow-medium",
			map[string]string{"-key": "fix-2-review", "-role": "reviewer", "-effort": "medium"}, true},
		{"short planner", "plan-short-2 design round", "Role: planner-short\nplan", "flow-planner-low",
			map[string]string{"-key": "plan-short-2", "-role": "planner", "-effort": "low"}, true},
		{"minor fix", "panel-2-minor-fix (Minors)", "Role: minor-fix\nKey: panel-2-minor-fix\nfix", "flow-low",
			map[string]string{"-key": "panel-2-minor-fix", "-role": "panel-fix", "-effort": "low"}, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			main, _ := beginRepo(t)
			env, log := beginEnv(t, `[]`)
			denied, out := runBegin(t, beginPayload(t, main, c.description, c.prompt, c.typ, "opus", ""), env...)
			begins := flowCalls(t, log, "record dispatch begin")
			if c.want == nil {
				if denied || len(begins) != 0 {
					t.Fatalf("want an allow and no begin: out=%s begins=%q", out, begins)
				}
				return
			}
			if denied || len(begins) != 1 {
				t.Fatalf("out=%s begins=%q", out, begins)
			}
			argv := strings.Fields(begins[0])
			for flag, v := range c.want {
				if got := argAfter(argv, flag); got != v {
					t.Errorf("%s = %q, want %q (%s)", flag, got, v, begins[0])
				}
			}
			if c.noTask && strings.Contains(begins[0], " -task ") {
				t.Errorf("a group begun with -task: %s", begins[0])
			}
		})
	}
}

// TestEnforceAgentBaselineHookSkipsAnOpenRow pins decision explicit-begin-fallback: a row already open is never begun again.
func TestEnforceAgentBaselineHookSkipsAnOpenRow(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct{ name, rows, description, prompt string }{
		{"panel slot", `[{"key":"panel-4-primary","outcome":""}]`, "panel-4-primary review", "Role: panel-reviewer\nreview"},
		{"task row", `[{"key":"task-3-implementer","outcome":""}]`, "Task 3/9 (x)", "Role: implementer\nKey: task-3-implementer\nimplement"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			main, _ := beginRepo(t)
			env, log := beginEnv(t, c.rows)
			if denied, out := runBegin(t, beginPayload(t, main, c.description, c.prompt, "flow-low", "opus", ""), env...); denied {
				t.Fatalf("denied: %s", out)
			}
			if begins := flowCalls(t, log, "record dispatch begin"); len(begins) != 0 {
				t.Errorf("an open row begun again: %q", begins)
			}
		})
	}
}

// TestPlannerDispatchAllowedUnderFlowPlan pins decision
// planning-agents-hook-begun: a /flow planner's Explore and Plan agents, whose prompts name the
// session change's worktree, are begun; a loop planner's, naming only the loop's worktree, are not,
// and neither is one naming the run's own worktree outside flow.plan (decision
// planning-agent-begins-in-brainstorm).
func TestPlannerDispatchAllowedUnderFlowPlan(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name, typ, description, role, names string // names: the worktree suffix the prompt names
		stage                               string // the latest stage run's stage; "" serves flow.plan
		want                                map[string]string
	}{
		{"Explore", "Explore", "explore-3 hooks", "explorer", "c1/hooks", "",
			map[string]string{"-role": "explorer", "-effort": "low", "-key": "explore-3", "-model": "opus"}},
		{"Plan", "Plan", "design-1 options", "plan-designer", "c1", "",
			map[string]string{"-role": "planner", "-effort": "low", "-key": "design-1"}},
		{"loop worktree", "Explore", "explore-3 hooks", "explorer", "pipeline-fix-c1-2/hooks", "", nil},
		{"a longer change name", "Explore", "explore-3 hooks", "explorer", "c10 hooks", "", nil},
		{"loop planner in the run's worktree", "Explore", "explore-3 hooks", "explorer", "c1/hooks", "flow.self-review", nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			main, _ := beginRepo(t)
			env, log := beginEnv(t, `[]`)
			if c.stage != "" {
				if err := os.WriteFile(filepath.Join(filepath.Dir(log), "stage"), []byte(c.stage), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			prompt := "Role: " + c.role + "\nread " + main + "-worktrees/" + c.names
			if denied, out := runBegin(t, beginPayload(t, main, c.description, prompt, c.typ, "opus", "low"), env...); denied {
				t.Fatalf("denied: %s", out)
			}
			begins := flowCalls(t, log, "record dispatch begin")
			if c.want == nil {
				if len(begins) != 0 {
					t.Errorf("a loop planner's agent begun: %q", begins)
				}
				return
			}
			if len(begins) != 1 {
				t.Fatalf("begins = %q, want one", begins)
			}
			argv := strings.Fields(begins[0])
			for flag, v := range c.want {
				if got := argAfter(argv, flag); got != v {
					t.Errorf("%s = %q, want %q (%s)", flag, got, v, begins[0])
				}
			}
		})
	}
}

// TestEnforceAgentBaselineHookRecordsThePairPick pins pair-pick-file: the begin the hook writes
// carries the `<source>: <reason>` of the .pick line flow-guard pair-judge saved for the key, and
// an undecodable .pick is read as no pick while the row is still begun.
func TestEnforceAgentBaselineHookRecordsThePairPick(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name, pick, want string
	}{
		{"judged", "opus low judge cheap fix\n", " -pair-pick judge: cheap fix"},
		{"undecodable", "\xff\xfe\n", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			main, wt := beginRepo(t)
			dir := filepath.Join(wt, ".superpowers", "sdd", "pair-judge")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "panel-fix-2.pick"), []byte(c.pick), 0o644); err != nil {
				t.Fatal(err)
			}
			env, log := beginEnv(t, `[]`)
			if denied, out := runBegin(t, beginPayload(t, main, "panel-fix-2 fix", "Role: panel-fix\nfix", "flow-low", "opus", ""), env...); denied {
				t.Fatalf("denied: %s", out)
			}
			begins := flowCalls(t, log, "record dispatch begin")
			if len(begins) != 1 {
				t.Fatalf("begins = %q, want one", begins)
			}
			if got := strings.Contains(begins[0], "-pair-pick"); got != (c.want != "") || (c.want != "" && !strings.HasSuffix(begins[0], c.want)) {
				t.Errorf("begin = %q, want suffix %q", begins[0], c.want)
			}
		})
	}
}
