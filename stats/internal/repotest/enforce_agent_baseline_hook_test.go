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
//	13. description, loop key line and Role line in a /flow session
//
// The allowed-pairs cases are TestEnforceAgentBaselineHookRefusesOffPolicyDispatch and
// TestEnforceAgentBaselineHookZcodePairForKnownRole.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// Cost: 6.37–7.72 s (load 8–13; no idle window) — twenty-eight runs of the
// real hook, each a python3 start plus the git, flow and flow-guard execs the
// hook makes for its case; the shared flow-guard build is TestMain's, not this
// test's. Cutting a run drops a dispatch shape the gate proves.
// smoke: core — the hook that gates every dispatch, through the real guard and CLI
func TestEnforceAgentBaselineHook(t *testing.T) {
	t.Parallel()
	const baseline = "Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent."
	const verifier = "Role: visual-verifier\n" + baseline + " Read /x/skills/flow/visual-verify-verifier.md and run its steps."
	const reviewer = "Role: panel-reviewer\n" + baseline

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

	// described builds an Agent payload; description "" and model "" omit those fields,
	// transcript "" omits transcript_path.
	described := func(description, prompt, typ, model, transcript string) string {
		ti := map[string]string{"prompt": prompt, "subagent_type": typ}
		if description != "" {
			ti["description"] = description
		}
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
	payload := func(prompt, typ, model, transcript string) string {
		return described("", prompt, typ, model, transcript)
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
		{"11 malformed JSON", hook, "{not json", nil},
		{"12 a planner citing the verifier file, no marker line", hook, payload("Key: self-review-x-plan\n"+baseline+" Finding 2 cites /x/skills/flow/visual-verify-verifier.md:120.", "flow-medium", "opus", ""), nil},
		{"12b an implementer citing both files, no marker line", hook, payload(baseline+" Task 4 re-reads .superpowers/sdd/tooling-analysis-1.md and step 8 of skills/flow/visual-verify-verifier.md.", "flow-low", "sonnet", ""), nil},
		{"13 a /flow dispatch described in prose", hook, described("Round-11 primary review", baseline+" review the diff", "flow-low", "opus", flowT), []string{"opens with the row it records"}},
		{"13b a /flow dispatch with no description", hook, payload(baseline+" review the diff", "flow-low", "opus", flowT), []string{"opens with the row it records"}},
		{"13c a /flow dispatch opening with its key", hook, described("panel-11-primary correctness review", reviewer+" review the diff", "flow-low", "opus", flowT), nil},
		{"13d a /flow task dispatch", hook, described("Tasks 3+4+7/22 (review)", "Role: gated-reviewer\n"+baseline+" review tasks 3, 4 and 7", "flow-medium", "opus", flowT), nil},
		{"13e a /flow loop dispatch carrying its Key line", hook, described("Self-review plan", "Key: self-review-x-plan\n"+baseline+" plan", "flow-medium", "opus", flowT), nil},
		{"13f a prose description outside a flow session", hook, described("Round-11 primary review", baseline+" review the diff", "flow-low", "opus", statusT), nil},
		{"13g a bare loop key prefix does not exempt", hook, described("Round-11 primary review", "Key: self-review-\n"+baseline+" plan", "flow-medium", "opus", flowT), []string{"opens with the row it records"}},
		{"13i a /flow dispatch whose first word is off the closed list", hook, described("verify-fix-1 m3 m4 departures", baseline+" fix the departures", "flow-low", "opus", flowT), []string{"opens with the row it records"}},
		{"13j a /flow dispatch opening with an inline key", hook, described("verify live check kan-1029", baseline+" run the live check", "flow-low", "opus", flowT), []string{"opens with the row it records"}},
		{"13k a /flow verifier opening with its repository-suffixed key", hook, described("visual-verify-fix-11-gymie re-verify", verifier, "flow-low", "opus", flowT), nil},
		{"13o a /flow planner dispatch opening with its plan-<n> key", hook, described("plan-2 design round", "Role: planner\nKey: plan-2\n"+baseline+" run section B", "flow-planner-medium", "opus", flowT), nil},
		{"13n a /flow slot's second ceiling-breach re-dispatch", hook, described("panel-3-primary-retry-2 correctness review", reviewer+" review the diff", "flow-low", "opus", flowT), nil},
		{"13l a /flow panel-fix chunk", hook, described("panel-fix-2-3 chunk three", "Role: panel-fix\n"+baseline+" fix", "flow-low", "opus", flowT), nil},
		{"13m a /flow tooling analyst opening with its repository-suffixed key", hook, described("tooling-analysis-1-gymie-frontend sweeps", "Role: tooling-analyst\n"+baseline+" sweep", "flow-medium", "opus", flowT), nil},
		{"13r a /flow fix-run judge opening with its fix-run-judge-<n> key", hook, described("fix-run-judge-2 roster and verify", "Role: fix-run-judge\n"+baseline+" judge", "flow-low", "opus", flowT), nil},
		{"13p a /flow-interactive operator-asked dispatch opening with its operator-<n> key", hook, described("operator-58 research run", "Role: operator-asked\n"+baseline+" run the research", "flow-high", "opus", interactiveT), nil},
		{"13q a /flow parent's operator-<k> dispatch is denied", hook, described("operator-1-live-check run the live check", "Role: operator-asked\n"+baseline+" run the live check", "flow-high", "opus", flowT), []string{"/flow-interactive"}},
		{"13h a loop key line with trailing words does not exempt", hook, described("Round-11 primary review", "Key: self-review-x more words\n"+baseline+" plan", "flow-medium", "opus", flowT), []string{"opens with the row it records"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			// An unreachable store keeps the live one's open experiments out of the pair check,
			// which reads the table alone when `flow experiment overrides` fails.
			out, _, rc := runHook(t, c.hook, []byte(c.stdin), "FLOW_ADDR=http://127.0.0.1:1")
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
// and stage flow.brainstorm -- or the contents of a `stage` file beside the called log, when a
// test writes one -- (status 500 when fail), and a fake `flow` first on PATH that appends its argv, one line
// per call, to the returned called file. It answers `record dispatches` with rows (exiting 1 when
// rows is ""), `record dispatch begin` with exit $FAKE_BEGIN_RC (default 0) after printing
// $FAKE_BEGIN_ERR to stderr, and every other call with exit 1. It returns the env the hook runs
// under and that argv log.
func openRowFixture(t *testing.T, rows string, fail bool, projectKey ...string) (env []string, called string) {
	t.Helper()
	bin := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail || r.URL.Path != "/api/v1/stage-runs" || r.URL.Query().Get("session_id") != "s1" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stage := "flow.brainstorm"
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
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + called + "\ncase \"$*\" in\n" +
		"'record dispatches'*) " + rowsAnswer + " ;;\n" +
		"'record dispatch begin'*) [ -n \"$FAKE_BEGIN_ERR\" ] && echo \"$FAKE_BEGIN_ERR\" >&2; exit \"${FAKE_BEGIN_RC:-0}\" ;;\n" +
		"*) exit 1 ;;\nesac\n"
	if err := os.WriteFile(filepath.Join(bin, "flow"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
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

// TestEnforceAgentBaselineHookRequiresAnOpenRow pins the begun-row check: a flow dispatch the
// store answers for with no open matching row, and whose row the hook cannot begin (its prompt
// carries no Role line), is denied with no begin; every unanswered store, and a loop-key dispatch,
// passes.
func TestEnforceAgentBaselineHookRequiresAnOpenRow(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	cases := []struct {
		name, rows, description, prompt string
		fail, deny                      bool
	}{
		{"open row", `[{"key":"panel-4-primary","outcome":""}]`, "panel-4-primary review", "review", false, false},
		{"open row, null outcome", `[{"key":"panel-4-primary","outcome":null}]`, "panel-4-primary review", "review", false, false},
		{"closed row", `[{"key":"panel-4-primary","outcome":"completed"}]`, "panel-4-primary review", "review", false, true},
		{"no row", `[{"key":"panel-3-primary","outcome":""}]`, "panel-4-primary review", "review", false, true},
		{"open task row", `[{"key":"task-3-implementer","outcome":""}]`, "Task 3/9 (x)", "implement", false, false},
		{"task row of another task", `[{"key":"task-13-implementer","outcome":""}]`, "Task 3/9 (x)", "implement", false, true},
		{"open group row", `[{"key":"task-1+3-implementer","outcome":""}]`, "Tasks 1+3/9 (x)", "implement", false, false},
		{"flow exits 1", "", "panel-4-primary review", "review", false, false},
		{"store answers 500", `[]`, "panel-4-primary review", "review", true, false},
		{"loop key", `[]`, "Self-review fix", "Key: self-review-x-fix", false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			env, called := openRowFixture(t, c.rows, c.fail)
			payload := openRowPayload(t, c.description, c.prompt)
			if c.deny {
				// No Role line: the hook cannot begin the row, so it keeps today's deny.
				payload = []byte(strings.Replace(string(payload), `Role: panel-reviewer\n`, "", 1))
			}
			out, _, rc := runHook(t, "enforce-agent-baseline.py", payload, env...)
			denied := strings.Contains(out, `"permissionDecision": "deny"`)
			if rc != 0 || denied != c.deny {
				t.Fatalf("rc=%d deny=%v want %v out=%s", rc, denied, c.deny, out)
			}
			if c.deny && !strings.Contains(out, "flow record dispatch begin") {
				t.Errorf("reason lacks the begin command: %s", out)
			}
			if begins := flowCalls(t, called, "record dispatch begin"); len(begins) != 0 {
				t.Errorf("begun %v", begins)
			}
			if c.name == "loop key" {
				if _, err := os.Stat(called); err == nil {
					t.Errorf("flow was called for a loop-key dispatch")
				}
			}
		})
	}
}

// TestEnforceAgentBaselineHookAllowsAnotherRepoCwd pins the fail-open on a cwd outside the
// change's project: its rows read empty there, so the hook lets the dispatch through.
func TestEnforceAgentBaselineHookAllowsAnotherRepoCwd(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	env, _ := openRowFixture(t, `[]`, false, "gymie-7c1f238a")
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
}

// TestEnforceAgentBaselineHookAdmitsRestKeys pins the verifier's -rest keys on the closed list: a
// flow dispatch opening with one, its row open, passes.
func TestEnforceAgentBaselineHookAdmitsRestKeys(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, key := range []string{"visual-verify-rest-gymie", "visual-verify-rest-2-gymie"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			env, _ := openRowFixture(t, `[{"key":"`+key+`","outcome":""}]`, false)
			out, _, rc := runHook(t, "enforce-agent-baseline.py", openRowPayload(t, key+" verifier remainder", "verify"), env...)
			if rc != 0 || out != "" {
				t.Fatalf("rc=%d out=%s", rc, out)
			}
		})
	}
}

// TestEnforceAgentBaselineHookAdmitsFullSuiteFixKey pins the full-suite fixer's key on the closed
// list: a flow dispatch opening with full-suite-fix-<n>, its row open, passes; one opening with a
// word off the list is denied for not naming its row.
func TestEnforceAgentBaselineHookAdmitsFullSuiteFixKey(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		description string
		deny        bool
	}{
		{"full-suite-fix-1 fix the failing suite", false},
		{"full-suite-fixes fix the failing suite", true},
	} {
		t.Run(c.description, func(t *testing.T) {
			t.Parallel()
			env, _ := openRowFixture(t, `[{"key":"full-suite-fix-1","outcome":""}]`, false)
			out, _, rc := runHook(t, "enforce-agent-baseline.py", openRowPayload(t, c.description, "fix"), env...)
			if rc != 0 {
				t.Fatalf("rc=%d out=%s", rc, out)
			}
			if !c.deny {
				if out != "" {
					t.Fatalf("admit: out=%s", out)
				}
				return
			}
			if !strings.Contains(out, `"permissionDecision": "deny"`) || !strings.Contains(out, "opens with the row it records") {
				t.Fatalf("deny: out=%s", out)
			}
		})
	}
}

// TestEnforceAgentBaselineHookAdmitsPlanReviewKey pins the plan reviewer's key on the closed list:
// a flow dispatch opening with plan-review-<n>, its row open, passes; one opening with
// plan-review-x is denied for not naming its row.
func TestEnforceAgentBaselineHookAdmitsPlanReviewKey(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		description string
		deny        bool
	}{
		{"plan-review-1 plan review", false},
		{"plan-review-x plan review", true},
	} {
		t.Run(c.description, func(t *testing.T) {
			t.Parallel()
			env, _ := openRowFixture(t, `[{"key":"plan-review-1","outcome":""}]`, false)
			out, _, rc := runHook(t, "enforce-agent-baseline.py", openRowPayload(t, c.description, "review"), env...)
			if rc != 0 {
				t.Fatalf("rc=%d out=%s", rc, out)
			}
			if !c.deny {
				if out != "" {
					t.Fatalf("admit: out=%s", out)
				}
				return
			}
			if !strings.Contains(out, `"permissionDecision": "deny"`) || !strings.Contains(out, "opens with the row it records") {
				t.Fatalf("deny: out=%s", out)
			}
		})
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

// TestEnforceAgentBaselineHookAdmitsPlanningAgents pins kan-1032's self-review: a planner's
// Explore or Plan dispatch with no key the hook begins under passes the row checks in a flow session,
// while any other type with the same description is still denied.
func TestEnforceAgentBaselineHookAdmitsPlanningAgents(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for typ, deny := range map[string]bool{"Explore": false, "Plan": false, "flow-low": true} {
		t.Run(typ, func(t *testing.T) {
			t.Parallel()
			env, _ := openRowFixture(t, `[]`, false)
			var payload map[string]any
			if err := json.Unmarshal(openRowPayload(t, "self-review plan — trace", "trace"), &payload); err != nil {
				t.Fatal(err)
			}
			ti := payload["tool_input"].(map[string]any)
			ti["subagent_type"] = typ
			if typ != "flow-low" {
				// A planning agent names its role and passes its effort apart.
				role := map[string]string{"Explore": "explorer", "Plan": "plan-designer"}[typ]
				ti["prompt"] = strings.Replace(ti["prompt"].(string), "Role: panel-reviewer", "Role: "+role, 1)
				ti["effort"] = "low"
			}
			b, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			out, _, rc := runHook(t, "enforce-agent-baseline.py", b, env...)
			if denied := strings.Contains(out, `"permissionDecision": "deny"`); rc != 0 || denied != deny {
				t.Fatalf("rc=%d deny=%v want %v out=%s", rc, denied, deny, out)
			}
		})
	}
}

// TestEnforceAgentBaselineHookSkipsAFinishedChange pins the row checks' scope: in a session whose
// change the store reports FINISHED, a dispatch with an off-row description passes; the same
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
			out, _, rc := runHook(t, "enforce-agent-baseline.py", openRowPayload(t, "Round-11 review", "review"),
				"FLOW_ADDR="+srv.URL, "FLOW_HOOK_UNBOUNDED=1")
			if denied := strings.Contains(out, `"permissionDecision": "deny"`); rc != 0 || denied != c.deny {
				t.Fatalf("rc=%d deny=%v want %v out=%s", rc, denied, c.deny, out)
			}
		})
	}
}

// offPolicyPayload is a /flow-session Agent payload outside the row checks' reach: a no-session
// transcript leaves the store unread, and the description opens with a closed-list key.
func offPolicyPayload(t *testing.T, prompt, typ, model, effort string) []byte {
	t.Helper()
	tr := filepath.Join(t.TempDir(), "flow.jsonl")
	if err := os.WriteFile(tr, []byte(`{"type":"user","message":{"content":"<command-name>/flow</command-name>"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ti := map[string]string{
		"description":   "panel-1-primary review",
		"prompt":        prompt + "\nBefore anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent.",
		"subagent_type": typ,
	}
	if model != "" {
		ti["model"] = model
	}
	if effort != "" {
		ti["effort"] = effort
	}
	b, err := json.Marshal(map[string]any{"tool_name": "Agent", "transcript_path": tr, "tool_input": ti})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestEnforceAgentBaselineHookRefusesOffPolicyDispatch pins KAN-1045: in a flow run every dispatch
// is checked against the Allowed pairs table through flow-guard check-dispatch-pair.
func TestEnforceAgentBaselineHookRefusesOffPolicyDispatch(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name, prompt, typ, model, effort string
		deny                             string // "" allows
	}{
		{"no Role line", "review the diff", "flow-low", "opus", "", "no `Role:` line"},
		{"unknown role", "Role: nobody\nreview", "flow-low", "opus", "", "unknown role"},
		{"wrong pair", "Role: panel-merger\nmerge", "flow-low", "opus", "", "panel-merger runs on sonnet/flow-low"},
		{"flow-high without its marker", "Role: panel-fix\nfix round 3", "flow-high", "opus", "", "No-progress escalation"},
		{"flow-high with its marker", "No-progress escalation\nRole: panel-fix\nfix round 3", "flow-high", "opus", "", ""},
		{"operator-approved high panel", "Operator-approved high\nRole: panel-reviewer\nreview", "flow-high", "opus", "", ""},
		{"flow-xhigh", "Role: panel-fix\nfix", "flow-xhigh", "opus", "", "panel-fix runs on"},
		{"no model", "Role: visual-verifier\nverify", "flow-low", "", "", "visual-verifier runs on opus/flow-low"},
		{"allowed pair", "Role: panel-merger\nmerge", "flow-low", "sonnet", "", ""},
		{"Explore at its effort", "Role: explorer\ntrace", "Explore", "opus", "medium", ""},
		{"Explore at high", "Role: explorer\ntrace", "Explore", "opus", "high", "explorer runs on"},
		{"a loop dispatch is the loop hook's", "Key: self-review-x-fix\nfix", "flow-high", "opus", "", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			out, _, rc := runHook(t, "enforce-agent-baseline.py", offPolicyPayload(t, c.prompt, c.typ, c.model, c.effort), "FLOW_ADDR=http://127.0.0.1:1", "FLOW_HOOK_UNBOUNDED=1")
			denied := strings.Contains(out, `"permissionDecision": "deny"`)
			if rc != 0 || denied != (c.deny != "") || !strings.Contains(out, c.deny) {
				t.Fatalf("rc=%d out=%s, want deny carrying %q", rc, out, c.deny)
			}
		})
	}
	// A guard that cannot answer (exit 2) denies -- a broken table never disables enforcement.
	stub := filepath.Join(t.TempDir(), "check-dispatch-pair.sh")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho 'check-dispatch-pair: table unreadable' >&2\nexit 2\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, _, rc := runHook(t, "enforce-agent-baseline.py", offPolicyPayload(t, "Role: panel-merger\nmerge", "flow-low", "sonnet", ""),
		"DISPATCH_PAIR_CHECK="+stub, "FLOW_ADDR=http://127.0.0.1:1", "FLOW_HOOK_UNBOUNDED=1"); rc != 0 || !strings.Contains(out, "exit 2") {
		t.Fatalf("a guard exit 2: rc=%d out=%s, want a deny naming exit 2", rc, out)
	}
	// Outside a flow session nothing is checked against the table.
	b, err := json.Marshal(map[string]any{"tool_name": "Agent", "tool_input": map[string]string{
		"prompt":        "review. Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent.",
		"subagent_type": "flow-high", "model": "opus",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if out, _, rc := runHook(t, "enforce-agent-baseline.py", b); rc != 0 || out != "" {
		t.Fatalf("outside a flow session: rc=%d out=%s", rc, out)
	}
}

// TestEnforceAgentBaselineHookZcodePairForKnownRole pins the ZCode mapping: run from ~/.zcode/hooks
// the hook still needs a Role line naming a role of the table, and admits the mapped pair alone.
func TestEnforceAgentBaselineHookZcodePairForKnownRole(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	zhook := filepath.Join(t.TempDir(), ".zcode", "hooks", "enforce-agent-baseline.py")
	if err := os.MkdirAll(filepath.Dir(zhook), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(repoRoot(t), "hooks", "enforce-agent-baseline.py"), zhook); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name, prompt, typ, model, deny string
	}{
		{"mapped pair", "Role: visual-verifier\nverify", "general-purpose", "glm-5.3-flash", ""},
		{"the claude pair", "Role: visual-verifier\nverify", "flow-low", "opus", "glm-5.3-flash"},
		{"a type ZCode does not offer", "Role: gated-reviewer\nreview", "flow-high", "glm-5.3-flash", "general-purpose"},
		{"unknown role", "Role: nobody\nverify", "general-purpose", "glm-5.3-flash", "unknown role"},
		{"no Role line", "verify", "general-purpose", "glm-5.3-flash", "no `Role:` line"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			out, _, rc := runHook(t, zhook, offPolicyPayload(t, c.prompt, c.typ, c.model, ""), "FLOW_ADDR=http://127.0.0.1:1", "FLOW_HOOK_UNBOUNDED=1")
			denied := strings.Contains(out, `"permissionDecision": "deny"`)
			if rc != 0 || denied != (c.deny != "") || !strings.Contains(out, c.deny) {
				t.Fatalf("rc=%d out=%s, want deny carrying %q", rc, out, c.deny)
			}
		})
	}
}

// anyCommandDeny runs a row-described dispatch, with no open ledger row, in a session whose
// transcript typed command, against a store answering the session's command row only to a
// stage-runs query carrying no `command` parameter, and reports whether the hook denied it.
func anyCommandDeny(t *testing.T, command string) bool {
	t.Helper()
	env, _ := openRowFixture(t, `[]`, false)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/stage-runs" || r.URL.Query().Get("session_id") != "s1" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		runs := []any{}
		if !r.URL.Query().Has("command") {
			runs = append(runs, map[string]string{"changeName": "c1", "command": command})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"stageRuns": runs})
	}))
	t.Cleanup(srv.Close)
	env = append(env, "FLOW_ADDR="+srv.URL)

	// A task description with no Key line names no row the hook can begin, so it is denied.
	payload := openRowPayload(t, "Task 3/9 (x)", "implement")
	var p struct {
		TranscriptPath string `json:"transcript_path"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatal(err)
	}
	line := `{"type":"user","message":{"content":"<command-name>` + command + `</command-name>"}}` + "\n"
	if err := os.WriteFile(p.TranscriptPath, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, rc := runHook(t, "enforce-agent-baseline.py", payload, env...)
	if rc != 0 {
		t.Fatalf("rc=%d out=%s", rc, out)
	}
	return strings.Contains(out, `"permissionDecision": "deny"`)
}

// TestEnforceAgentBaselineHookReadsAnyCommand pins KAN-1061's session-change-any-command: the
// begun-row check reads the session's latest stage run of any command, so a /flow-fast
// session's dispatch with no open row is denied rather than failing open.
func TestEnforceAgentBaselineHookReadsAnyCommand(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	if !anyCommandDeny(t, "/flow-fast") {
		t.Errorf("a /flow-fast session's dispatch with no open row was allowed, want denied")
	}
}

// TestEnforceAgentBaselineHookSeesFlowInteractive pins that a typed /flow-interactive is a flow
// session, its dispatches held to the begun-row check like every flow command's.
func TestEnforceAgentBaselineHookSeesFlowInteractive(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	if !anyCommandDeny(t, "/flow-interactive") {
		t.Errorf("a /flow-interactive session's dispatch with no open row was allowed, want denied")
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
	t.Run("open row deny carries the change", func(t *testing.T) {
		cmd, log := incidentStub(t, 0)
		env, _ := openRowFixture(t, `[]`, false)
		out, _, rc := runHook(t, "enforce-agent-baseline.py", openRowPayload(t, "Task 3/9 (x)", "implement"), append(env, "FLOW_HOOK_INCIDENT_CMD="+cmd)...)
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

// flowCalls returns the argv lines the fake `flow` (or the snapshot stub) logged to log that open
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
// checkout, the payload's cwd, and its change worktree <checkout>-worktrees/c1, whose
// .superpowers/sdd/pair-judge/ exists.
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
	wt = main + "-worktrees/c1"
	if err := os.MkdirAll(filepath.Join(wt, ".superpowers", "sdd", "pair-judge"), 0o755); err != nil {
		t.Fatal(err)
	}
	return main, wt
}

// beginEnv is openRowFixture's env plus a check-plan-unchanged stand-in that logs
// `check-plan-unchanged <argv>` to the same argv log and exits snapshotRC.
func beginEnv(t *testing.T, rows string, snapshotRC int) (env []string, log string) {
	t.Helper()
	env, log = openRowFixture(t, rows, false)
	stub := filepath.Join(t.TempDir(), "check-plan-unchanged")
	script := fmt.Sprintf("#!/bin/sh\nprintf 'check-plan-unchanged %%s\\n' \"$*\" >> %s\nexit %d\n", log, snapshotRC)
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return append(env, "DISPATCH_PLAN_SNAPSHOT="+stub), log
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

// TestEnforceAgentBaselineHookBeginsTheRow pins decisions hook-same-file,
// hook-token-from-stage-run and pair-pick-file: a flow dispatch with no open row is begun, not
// denied, with the stage run's token, its Role line's role, the call's model and effort, and the
// key's .pick line as -pair-pick.
func TestEnforceAgentBaselineHookBeginsTheRow(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, pick := range []bool{true, false} {
		t.Run(fmt.Sprintf("pick=%v", pick), func(t *testing.T) {
			t.Parallel()
			main, wt := beginRepo(t)
			want := "record dispatch begin -C " + main + " -change c1 -role panel-fix -model opus -effort low -key panel-fix-2 -session-token mf-t1"
			if pick {
				if err := os.WriteFile(filepath.Join(wt, ".superpowers", "sdd", "pair-judge", "panel-fix-2.pick"), []byte("opus low judge cheap fix\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				want += " -pair-pick judge: cheap fix"
			}
			env, log := beginEnv(t, `[]`, 0)
			if denied, out := runBegin(t, beginPayload(t, main, "panel-fix-2 fix", "Role: panel-fix\nfix", "flow-low", "opus", ""), env...); denied {
				t.Fatalf("denied: %s", out)
			}
			if got := flowCalls(t, log, "record dispatch begin"); len(got) != 1 || got[0] != want {
				t.Errorf("begins = %q, want [%q]", got, want)
			}
		})
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
	env, log := beginEnv(t, `[]`, 0)
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
	want := "record dispatch begin -C " + main + " -change c1 -role implementer -model opus -effort low -task 52 -key task-52-implementer-fix-3 -session-token mf-t1"
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
			env, log := beginEnv(t, `[]`, 0)
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
			want := "record dispatch begin -C " + main + " -change c1 -role implementer -model opus -effort low -task 3 -key task-3-implementer -session-token mf-t1"
			if got := flowCalls(t, log, "record dispatch begin"); len(got) != 1 || got[0] != want {
				t.Errorf("begins = %q, want [%q]", got, want)
			}
		})
	}
}

// TestEnforceAgentBaselineHookPointsAPanelKeyAtPanelRoundBegins pins decision f4-panel-key-deny:
// a panel reviewer key with no open row is denied with a pointer at panel-round-begins, since a
// begin from the hook carries no -slot and flow refuses it; the merger and panel-fix keys are
// begun as before.
func TestEnforceAgentBaselineHookPointsAPanelKeyAtPanelRoundBegins(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	t.Run("reviewer", func(t *testing.T) {
		t.Parallel()
		main, wt := beginRepo(t)
		env, log := beginEnv(t, `[]`, 0)
		denied, out := runBegin(t, beginPayload(t, main, "panel-1-primary+principles review", "Role: panel-reviewer\nreview", "flow-low", "opus", ""), append(env, "DISPATCH_PAIR_CHECK=true")...)
		if !denied {
			t.Fatalf("not denied: %s", out)
		}
		for _, want := range []string{"flow-guard panel-round-begins " + wt + " c1 1 mf-t1 primary+principles:opus:low", "@<repo>"} {
			if !strings.Contains(out, want) {
				t.Errorf("reason lacks %q: %s", want, out)
			}
		}
		if got := flowCalls(t, log, "record dispatch begin"); len(got) != 0 {
			t.Errorf("begins = %q, want none", got)
		}
	})
	t.Run("suffixed reviewer", func(t *testing.T) {
		t.Parallel()
		main, wt := beginRepo(t)
		env, _ := beginEnv(t, `[]`, 0)
		denied, out := runBegin(t, beginPayload(t, main, "panel-1-primary-part2 review", "Role: panel-reviewer\nreview", "flow-low", "opus", ""), append(env, "DISPATCH_PAIR_CHECK=true")...)
		if want := "flow-guard panel-round-begins " + wt + " c1 1 mf-t1 primary@part2:opus:low"; !denied || !strings.Contains(out, want) {
			t.Errorf("denied=%v, reason lacks %q: %s", denied, want, out)
		}
	})
	t.Run("repo-suffixed reviewer", func(t *testing.T) {
		t.Parallel()
		main, wt := beginRepo(t)
		env, _ := beginEnv(t, `[]`, 0)
		denied, out := runBegin(t, beginPayload(t, main, "panel-1-primary-agents review", "Role: panel-reviewer\nreview", "flow-low", "opus", ""), append(env, "DISPATCH_PAIR_CHECK=true")...)
		if want := "flow-guard panel-round-begins " + wt + " c1 1 mf-t1 <slots>@<repo>:opus:low"; !denied || !strings.Contains(out, want) || strings.Contains(out, "primary-agents:") {
			t.Errorf("denied=%v, reason lacks %q: %s", denied, want, out)
		}
	})
	for _, c := range []struct{ name, description, prompt, role string }{
		{"merger", "panel-2-merger merge", "Role: panel-merger\nmerge", "merger"},
		{"panel-fix", "panel-fix-2 fix", "Role: panel-fix\nfix", "panel-fix"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			main, _ := beginRepo(t)
			env, log := beginEnv(t, `[]`, 0)
			if denied, out := runBegin(t, beginPayload(t, main, c.description, c.prompt, "flow-low", "opus", ""), append(env, "DISPATCH_PAIR_CHECK=true")...); denied {
				t.Fatalf("denied: %s", out)
			}
			begins := flowCalls(t, log, "record dispatch begin")
			if len(begins) != 1 || argAfter(strings.Fields(begins[0]), "-role") != c.role {
				t.Errorf("begins = %q, want one with -role %s", begins, c.role)
			}
		})
	}
}

// TestEnforceAgentBaselineHookBeginsEveryRowRole pins ROW_ROLE: every Role id launchRoleID
// (stats/cmd/flow/record.go) can return, and every panel and planner id the table adds, is begun
// under its row role rather than denied for want of an open row. The pair check is stubbed: the
// pairs each role admits are flowd's, not this table's.
func TestEnforceAgentBaselineHookBeginsEveryRowRole(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for id, role := range map[string]string{
		"implementer": "implementer", "operator-asked": "implementer",
		"gated-reviewer": "reviewer", "gated-re-reviewer": "reviewer", "design-reviewer": "reviewer",
		"panel-reviewer": "reviewer", "panel-sanity-efficiency": "reviewer",
		"panel-mutation-failure": "reviewer", "panel-rerun": "reviewer", "late-fix": "reviewer",
		"panel-merger": "merger", "panel-fix": "panel-fix", "visual-verifier": "verifier",
		"planner": "planner", "tooling-analyst": "planner", "fix-run-judge": "planner",
	} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			main, _ := beginRepo(t)
			env, log := beginEnv(t, `[]`, 0)
			if denied, out := runBegin(t, beginPayload(t, main, "panel-fix-3 x", "Role: "+id+"\nwork", "flow-low", "opus", ""), append(env, "DISPATCH_PAIR_CHECK=true")...); denied {
				t.Fatalf("denied: %s", out)
			}
			begins := flowCalls(t, log, "record dispatch begin")
			if len(begins) != 1 || argAfter(strings.Fields(begins[0]), "-role") != role {
				t.Errorf("begins = %q, want one with -role %s", begins, role)
			}
		})
	}
}

// TestEnforceAgentBaselineHookIgnoresAnUndecodablePick: a .pick that is not UTF-8 is read as no
// pick, and the row is still begun.
func TestEnforceAgentBaselineHookIgnoresAnUndecodablePick(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	main, wt := beginRepo(t)
	if err := os.WriteFile(filepath.Join(wt, ".superpowers", "sdd", "pair-judge", "panel-fix-2.pick"), []byte("\xff\xfe\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env, log := beginEnv(t, `[]`, 0)
	if denied, out := runBegin(t, beginPayload(t, main, "panel-fix-2 fix", "Role: panel-fix\nfix", "flow-low", "opus", ""), env...); denied {
		t.Fatalf("denied: %s", out)
	}
	if begins := flowCalls(t, log, "record dispatch begin"); len(begins) != 1 || strings.Contains(begins[0], "-pair-pick") {
		t.Errorf("begins = %q, want one with no -pair-pick", begins)
	}
}

// TestEnforceAgentBaselineHookBeginKeyFromKeyLine pins the key's source: the prompt's `Key:` line
// tells a gated reviewer bundle from its implementer group, both described `Task <ids>/<n>`; a
// task description with no `Key:` line names no row the hook can begin and keeps today's deny.
func TestEnforceAgentBaselineHookBeginKeyFromKeyLine(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name, description, prompt, typ string
		want                           map[string]string // flag -> value; nil: denied, no begin
		noTask                         bool
	}{
		{"gated reviewer", "Task 3/9 (review)", "Role: gated-reviewer\nKey: task-3-reviewer\nreview", "flow-medium",
			map[string]string{"-key": "task-3-reviewer", "-role": "reviewer", "-task": "3", "-effort": "medium"}, false},
		{"implementer group", "Tasks 3+4/9 (x)", "Role: implementer\nKey: task-3+4-implementer\nimplement", "flow-low",
			map[string]string{"-key": "task-3+4-implementer", "-role": "implementer"}, true},
		{"no Key line", "Task 3/9 (x)", "Role: implementer\nimplement", "flow-low", nil, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			main, _ := beginRepo(t)
			env, log := beginEnv(t, `[]`, 0)
			denied, out := runBegin(t, beginPayload(t, main, c.description, c.prompt, c.typ, "opus", ""), env...)
			begins := flowCalls(t, log, "record dispatch begin")
			if c.want == nil {
				if !denied || !strings.Contains(out, "no open ledger row") || len(begins) != 0 {
					t.Fatalf("want today's deny and no begin: out=%s begins=%q", out, begins)
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

// TestEnforceAgentBaselineHookSkipsAnOpenRow pins decisions panel-round-begins-kept and
// explicit-begin-fallback: a row already open is never begun again.
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
			env, log := beginEnv(t, c.rows, 0)
			if denied, out := runBegin(t, beginPayload(t, main, c.description, c.prompt, "flow-low", "opus", ""), env...); denied {
				t.Fatalf("denied: %s", out)
			}
			if begins := flowCalls(t, log, "record dispatch begin"); len(begins) != 0 {
				t.Errorf("an open row begun again: %q", begins)
			}
		})
	}
}

// TestEnforceAgentBaselineHookSnapshotsAGatedReviewer pins decision plan-check-rides-row: a gated
// reviewer's plan tree is snapshotted to plan-snapshot-<key>.txt before its row is begun, and a
// failed snapshot denies the launch with no row begun.
func TestEnforceAgentBaselineHookSnapshotsAGatedReviewer(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, rc := range []int{0, 2} {
		t.Run(fmt.Sprintf("snapshot exit %d", rc), func(t *testing.T) {
			t.Parallel()
			main, wt := beginRepo(t)
			env, log := beginEnv(t, `[]`, rc)
			denied, out := runBegin(t, beginPayload(t, main, "Task 3/9 (review)", "Role: gated-reviewer\nKey: task-3-reviewer\nreview", "flow-medium", "opus", ""), env...)
			snap := "check-plan-unchanged snapshot " + wt + " c1 " + wt + "/.superpowers/sdd/plan-snapshot-task-3-reviewer.txt"
			calls := append(flowCalls(t, log, "check-plan-unchanged"), flowCalls(t, log, "record dispatch begin")...)
			if rc != 0 {
				if !denied || len(calls) != 1 || calls[0] != snap {
					t.Fatalf("want a deny after the snapshot alone: out=%s calls=%q", out, calls)
				}
				return
			}
			b, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			if denied || len(calls) != 2 || calls[0] != snap || strings.Index(string(b), snap) > strings.Index(string(b), "record dispatch begin") {
				t.Fatalf("want the snapshot, then the begin: out=%s log=%s", out, b)
			}
		})
	}
}

// TestEnforceAgentBaselineHookDeniesARefusedBegin pins decision hook-begin-refusal-denies: a begin
// flow refuses denies the launch with flow's stderr verbatim; a begin flow journals (exit 0) lets
// it through.
func TestEnforceAgentBaselineHookDeniesARefusedBegin(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		rc         string
		stderr     string
		wantDenied bool
	}{
		{"2", "flow: session cap", true},
		{"0", "flow: store unreachable, journalled", false},
	} {
		t.Run("exit "+c.rc, func(t *testing.T) {
			t.Parallel()
			main, _ := beginRepo(t)
			env, _ := beginEnv(t, `[]`, 0)
			env = append(env, "FAKE_BEGIN_RC="+c.rc, "FAKE_BEGIN_ERR="+c.stderr)
			denied, out := runBegin(t, beginPayload(t, main, "panel-fix-2 fix", "Role: panel-fix\nfix", "flow-low", "opus", ""), env...)
			if denied != c.wantDenied {
				t.Fatalf("denied=%v want %v: %s", denied, c.wantDenied, out)
			}
			if denied && (!strings.Contains(out, c.stderr) || !strings.Contains(out, "flow refused this dispatch's begin")) {
				t.Errorf("reason lacks flow's stderr: %s", out)
			}
		})
	}
}

// TestEnforceAgentBaselineHookBeginsPlanningAgentsInItsWorktree pins decision
// planning-agents-hook-begun: a /flow planner's Explore and Plan agents, whose prompts name the
// session change's worktree, are begun; a loop planner's, naming only the loop's worktree, are not,
// and neither is one naming the run's own worktree outside flow.brainstorm (decision
// planning-agent-begins-in-brainstorm).
func TestEnforceAgentBaselineHookBeginsPlanningAgentsInItsWorktree(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name, typ, description, role, names string // names: the worktree suffix the prompt names
		stage                               string // the latest stage run's stage; "" serves flow.brainstorm
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
			env, log := beginEnv(t, `[]`, 0)
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

// TestEnforceAgentBaselineHookBeginsAfterThePairCheck pins that an off-policy dispatch the pair
// check denies leaves no begun row.
func TestEnforceAgentBaselineHookBeginsAfterThePairCheck(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	stub := filepath.Join(t.TempDir(), "check-dispatch-pair")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho 'off policy'\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	main, _ := beginRepo(t)
	env, log := beginEnv(t, `[]`, 0)
	denied, out := runBegin(t, beginPayload(t, main, "panel-fix-2 fix", "Role: panel-fix\nfix", "flow-low", "opus", ""), append(env, "DISPATCH_PAIR_CHECK="+stub)...)
	if !denied || !strings.Contains(out, "check-dispatch-pair exit 1") {
		t.Fatalf("want the pair check's deny: %s", out)
	}
	if begins := flowCalls(t, log, "record dispatch begin"); len(begins) != 0 {
		t.Errorf("an off-policy dispatch begun: %q", begins)
	}
}
