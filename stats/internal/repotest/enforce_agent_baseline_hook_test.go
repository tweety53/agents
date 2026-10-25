package repotest

// TestEnforceAgentBaselineHook ports the retired shell harness,
// the harness for hooks/enforce-agent-baseline.py. It pipes PreToolUse payloads
// to the hook and asserts its decision, one subtest per harness case:
//
//	 1. no baseline pointer                                 -> deny (baseline reason)
//	 2. pointer, no pinned-role marker line, any type/model -> allow
//	2b. hand-picked flow-high                               -> deny
//	2c. "No-progress escalation" line on flow-high          -> allow
//	2d. "Operator-approved high" line on flow-high          -> allow
//	2e. a pipeline-fix loop key on flow-high                -> allow
//	 3. verifier prompt on flow-low / opus                  -> allow
//	 4. verifier prompt on flow-high / opus                 -> deny, names flow-low and opus
//	 5. verifier prompt on flow-low with no model           -> deny
//	 6. tooling-analyst prompt (also names the verifier
//	    file) on flow-medium / opus                         -> allow
//	 7. tooling-analyst prompt on flow-high / opus          -> deny, names flow-medium
//	 8. verifier prompt run from a ~/.zcode/hooks path on
//	    general-purpose / glm-5.3-flash                     -> allow
//	8b. the same from ~/.zcode/hooks on flow-high /
//	    glm-5.3-flash, a type ZCode does not offer (KAN-609) -> deny, names general-purpose
//	 9. the same from ~/.zcode/hooks on flow-low / opus     -> deny, names glm-5.3-flash
//	10. a fork dispatch carrying the verifier marker        -> deny (names a flow skill)
//	10b. a fork naming no flow skill, no transcript         -> allow
//	10c. the same fork in a session whose transcript records
//	     a /flow command                                    -> deny
//	10d. the same fork, transcript recording only /flow-status -> allow
//	10e. the same fork, the /flow tag only inside a tool result -> allow
//	11. malformed JSON                                      -> allow, exit 0
//	12. a planner or implementer prompt citing the verifier
//	    or analyst file, with no marker line, off the
//	    pinned pair                                         -> allow
//
// The ZCode cases run the hook through a symlink under a temp
// .zcode/hooks directory, as setup.sh installs it there.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnforceAgentBaselineHook(t *testing.T) {
	t.Parallel()
	const baseline = "Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent."
	const verifier = "Role: visual-verify verifier\n" + baseline + " Read /x/skills/flow/visual-verify-verifier.md and run its steps."
	const analyst = "Role: visual-verify tooling analyst\n" + baseline + " Name the sweep in /x/skills/flow/visual-verify-verifier.md; write .superpowers/sdd/tooling-analysis-1.md."

	sandbox := t.TempDir()
	zhook := filepath.Join(sandbox, ".zcode", "hooks", "enforce-agent-baseline.py")
	if err := os.MkdirAll(filepath.Dir(zhook), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(repoRoot(t), "hooks", "enforce-agent-baseline.py"), zhook); err != nil {
		t.Fatal(err)
	}
	transcript := func(name, line string) string {
		p := filepath.Join(sandbox, name)
		if err := os.WriteFile(p, []byte(line+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	flowT := transcript("flow.jsonl", `{"type":"user","message":{"content":"<command-name>/flow</command-name>"}}`)
	statusT := transcript("status.jsonl", `{"type":"user","message":{"content":"<command-name>/flow-status</command-name>"}}`)
	toolT := transcript("tool.jsonl", `{"type":"user","message":{"content":[{"type":"tool_result","content":"<command-name>/flow</command-name>"}]}}`)

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
		{"2 unpinned role", hook, payload(baseline+" implement task 3", "flow-medium", "sonnet", ""), nil},
		{"2b hand-picked flow-high", hook, payload(baseline+" implement task 3", "flow-high", "opus", ""), []string{"reserved for the no-progress escalation"}},
		{"2c no-progress escalation on flow-high", hook, payload("No-progress escalation\n"+baseline+" fix round 3", "flow-high", "opus", ""), nil},
		{"2d operator-approved panel on flow-high", hook, payload("Operator-approved high\n"+baseline+" panel pass 1", "flow-high", "opus", ""), nil},
		{"2e pipeline-fix loop escalation on flow-high", hook, payload("Key: pipeline-fix-x-plan-2\n"+baseline+" re-plan", "flow-high", "opus", ""), nil},
		{"3 verifier on flow-low/opus", hook, payload(verifier, "flow-low", "opus", ""), nil},
		{"4 verifier on flow-high", hook, payload(verifier, "flow-high", "opus", ""), []string{`subagent_type: \"flow-low\"`, `model: \"opus\"`}},
		{"5 verifier with no model", hook, payload(verifier, "flow-low", "", ""), []string{`model: \"opus\"`}},
		{"6 analyst on flow-medium/opus", hook, payload(analyst, "flow-medium", "opus", ""), nil},
		{"7 analyst on flow-high", hook, payload(analyst, "flow-high", "opus", ""), []string{`subagent_type: \"flow-medium\"`}},
		{"8 zcode verifier on its mapped pair", zhook, payload(verifier, "general-purpose", "glm-5.3-flash", ""), nil},
		{"8b zcode verifier on a type ZCode does not offer", zhook, payload(verifier, "flow-high", "glm-5.3-flash", ""), []string{`subagent_type: \"general-purpose\"`}},
		{"9 zcode verifier on the claude pair", zhook, payload(verifier, "flow-low", "opus", ""), []string{`model: \"glm-5.3-flash\"`}},
		{"10 fork naming a flow skill", hook, payload(verifier, "fork", "", ""), []string{`subagent_type \"fork\"`}},
		{"10b fork outside a flow session", hook, payload("summarise the thread above", "fork", "", ""), nil},
		{"10c fork in a /flow session", hook, payload("run the loop above", "fork", "", flowT), []string{"KAN-1000"}},
		{"10d fork after /flow-status alone", hook, payload("run the loop above", "fork", "", statusT), nil},
		{"10e fork whose transcript holds the tag only in a tool result", hook, payload("run the loop above", "fork", "", toolT), nil},
		{"11 malformed JSON", hook, "{not json", nil},
		{"12 a planner citing the verifier file, no marker line", hook, payload("Key: self-review-x-plan\n"+baseline+" Finding 2 cites /x/skills/flow/visual-verify-verifier.md:120.", "flow-medium", "opus", ""), nil},
		{"12b an implementer citing both files, no marker line", hook, payload(baseline+" Task 4 re-reads .superpowers/sdd/tooling-analysis-1.md and step 8 of skills/flow/visual-verify-verifier.md.", "flow-low", "sonnet", ""), nil},
		{"13 a /flow dispatch described in prose", hook, described("Round-11 primary review", baseline+" review the diff", "flow-low", "opus", flowT), []string{"opens with the row it records"}},
		{"13b a /flow dispatch with no description", hook, payload(baseline+" review the diff", "flow-low", "opus", flowT), []string{"opens with the row it records"}},
		{"13c a /flow dispatch opening with its key", hook, described("panel-11-primary correctness review", baseline+" review the diff", "flow-low", "opus", flowT), nil},
		{"13d a /flow task dispatch", hook, described("Tasks 3+4+7/22 (review)", baseline+" review tasks 3, 4 and 7", "flow-low", "opus", flowT), nil},
		{"13e a /flow loop dispatch carrying its Key line", hook, described("Self-review plan", "Key: self-review-x-plan\n"+baseline+" plan", "flow-medium", "opus", flowT), nil},
		{"13f a prose description outside a flow session", hook, described("Round-11 primary review", baseline+" review the diff", "flow-low", "opus", statusT), nil},
		{"13g a bare loop key prefix does not exempt", hook, described("Round-11 primary review", "Key: self-review-\n"+baseline+" plan", "flow-medium", "opus", flowT), []string{"opens with the row it records"}},
		{"13i a /flow dispatch whose first word is off the closed list", hook, described("verify-fix-1 m3 m4 departures", baseline+" fix the departures", "flow-low", "opus", flowT), []string{"opens with the row it records"}},
		{"13j a /flow dispatch opening with an inline key", hook, described("verify live check kan-1029", baseline+" run the live check", "flow-low", "opus", flowT), []string{"opens with the row it records"}},
		{"13k a /flow verifier opening with its repository-suffixed key", hook, described("visual-verify-fix-11-gymie re-verify", baseline+" verify", "flow-low", "opus", flowT), nil},
		{"13l a /flow panel-fix chunk", hook, described("panel-fix-2-3 chunk three", baseline+" fix", "flow-low", "opus", flowT), nil},
		{"13m a /flow tooling analyst opening with its repository-suffixed key", hook, described("tooling-analysis-1-gymie-frontend sweeps", baseline+" sweep", "flow-low", "opus", flowT), nil},
		{"13h a loop key line with trailing words does not exempt", hook, described("Round-11 primary review", "Key: self-review-x more words\n"+baseline+" plan", "flow-medium", "opus", flowT), []string{"opens with the row it records"}},
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

// openRowFixture serves a store answering /api/v1/stage-runs with change c1 (status 500 when
// fail), and a fake `flow` first on PATH printing rows (exiting 1 when rows is ""). It returns the
// env the hook runs under and a file the fake `flow` touches when called.
func openRowFixture(t *testing.T, rows string, fail bool, projectKey ...string) (env []string, called string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail || r.URL.Path != "/api/v1/stage-runs" || r.URL.Query().Get("session_id") != "s1" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		run := map[string]string{"changeName": "c1"}
		if len(projectKey) > 0 {
			run["projectKey"] = projectKey[0]
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"stageRuns": []any{run}})
	}))
	t.Cleanup(srv.Close)
	bin := t.TempDir()
	called = filepath.Join(bin, "called")
	script := "#!/bin/sh\ntouch " + called + "\n"
	if rows == "" {
		script += "exit 1\n"
	} else {
		if err := os.WriteFile(filepath.Join(bin, "rows.json"), []byte(rows), 0o644); err != nil {
			t.Fatal(err)
		}
		script += "cat " + filepath.Join(bin, "rows.json") + "\n"
	}
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
			"prompt":        prompt + "\nBefore anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task. Include this instruction verbatim in any prompt you write for another agent.",
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
// store answers for with no open matching row is denied; every unanswered store, and a loop-key
// dispatch, passes.
func TestEnforceAgentBaselineHookRequiresAnOpenRow(t *testing.T) {
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
			out, _, rc := runHook(t, "enforce-agent-baseline.py", openRowPayload(t, c.description, c.prompt), env...)
			denied := strings.Contains(out, `"permissionDecision": "deny"`)
			if rc != 0 || denied != c.deny {
				t.Fatalf("rc=%d deny=%v want %v out=%s", rc, denied, c.deny, out)
			}
			if c.deny && !strings.Contains(out, "flow record dispatch begin") {
				t.Errorf("reason lacks the begin command: %s", out)
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
