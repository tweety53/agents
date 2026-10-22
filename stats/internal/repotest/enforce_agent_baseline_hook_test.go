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
	"os"
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

	// payload builds an Agent payload; model "" omits it, transcript "" omits transcript_path.
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
