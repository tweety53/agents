package guard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dgpvPlan is a three-task plan: task 2 big, task 3 after task 1.
const dgpvPlan = "# Tasks\n\n- [ ] 1. first\n\n**Size:** small\n\n- [ ] 2. second\n\n**Size:** big\n\n- [ ] 3. third\n\n**After:** Task 1\n"

const dgpvPolicy = `"decider":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}`

// dgpvFixture is a worktree declaring visual verification and a change root
// holding the plan and a design.md; out is the decision path.
func dgpvFixture(t *testing.T) (wt, root, out string) {
	t.Helper()
	wt, root = t.TempDir(), t.TempDir()
	writeFile(t, wt+"/.flow/project.md", "## visual verification\n\n| Setting | Value |\n")
	writeFile(t, root+"/tasks.md", dgpvPlan)
	writeFile(t, root+"/design.md", "the design")
	return wt, root, filepath.Join(t.TempDir(), "decision.json")
}

func dgpvRead(t *testing.T, out string) dgpvDecision {
	t.Helper()
	var d dgpvDecision
	b, err := os.ReadFile(out)
	if err != nil || json.Unmarshal(b, &d) != nil {
		t.Fatalf("decision %s: %v %s", out, err, b)
	}
	return d
}

const dgpvGood = `{"groups":[{"tasks":["1","3"],"reason":"a"},{"tasks":["2"],"reason":"b"}],` +
	`"panel":[{"slots":["primary","principles"],"reason":"c"},{"slots":["mutation"],"reason":"d"}],` +
	`"visual":{"verify":"skipped","reason":"e"}}`

// TestDecideGroupsPanelVisual: a valid answer is written as given, overrides
// empty, on the decider row's first pair, the plan, design and visual flag
// on stdin and decider-prompt.md as the system prompt.
func TestDecideGroupsPanelVisual(t *testing.T) {
	t.Parallel()
	wt, root, out := dgpvFixture(t)
	stub := newPJStub(t, pjJSON("```json\n"+dgpvGood+"\n```", ""))
	r := runGuard("decide-groups-panel-visual", []string{wt, root, out}, pjEnv(t, pjPolicy(dgpvPolicy), stub))
	if r.rc != 0 || !strings.HasPrefix(r.stdout, "decided (judge): 2 groups · panel primary+principles mutation · visual skipped") {
		t.Fatalf("rc=%d stdout=%q err=%q", r.rc, r.stdout, r.err)
	}
	b, _ := os.ReadFile(out)
	if d := dgpvRead(t, out); len(d.Groups) != 2 || d.Groups[0].Tasks[1] != "3" || d.Visual.Verify != "skipped" || d.Overrides == nil || !strings.Contains(string(b), `"overrides": []`) {
		t.Errorf("decision %s", b)
	}
	if in := stub.stdin(t); !strings.Contains(in, "- [ ] 2. second") || !strings.Contains(in, "the design") || !strings.Contains(in, "`## visual verification`: yes") {
		t.Errorf("stdin %q", in)
	}
	if sp := stub.systemPrompt(t); !strings.Contains(sp, "You decide how an approved plan is built and reviewed") {
		t.Errorf("system prompt is not decider-prompt.md: %q", sp)
	}
	if a := strings.Join(stub.argv(t), " "); !strings.Contains(a, "--model sonnet --effort low") {
		t.Errorf("argv %q, want the decider row's first pair", a)
	}
}

// TestDecideGroupsPanelVisualFallsBack: a failed call or an answer breaking
// a rule writes the fallback decision, exit 0, one stderr line naming why.
func TestDecideGroupsPanelVisualFallsBack(t *testing.T) {
	t.Parallel()
	swap := func(old, new string) string { return pjJSON(strings.Replace(dgpvGood, old, new, 1), "") }
	for name, c := range map[string]struct{ policy, answer, cause string }{
		"no pair":             {pjPolicy(`"pair-judge":{"base":["opus/flow-low"],"optin":[],"marked":[]}`), pjJSON("", ""), "no decider pair"},
		"command fails":       {pjPolicy(dgpvPolicy), "exit 3\n", "exit 3"},
		"not an object":       {pjPolicy(dgpvPolicy), pjJSON("no idea", ""), "not a JSON object"},
		"task missing":        {pjPolicy(dgpvPolicy), swap(`["1","3"]`, `["1"]`), "task 3 is in no group"},
		"big task shares":     {pjPolicy(dgpvPolicy), swap(`["2"]`, `["2","3"]`), "big task 2 shares group 2"},
		"after order broken":  {pjPolicy(dgpvPolicy), pjJSON(`{"groups":[{"tasks":["3"],"reason":"a"},{"tasks":["1"],"reason":"a"},{"tasks":["2"],"reason":"b"}],"panel":[{"slots":["primary"],"reason":"c"}],"visual":{"verify":"skipped","reason":"e"}}`, ""), "comes before its After task 1"},
		"mutation shares":     {pjPolicy(dgpvPolicy), swap(`["mutation"]`, `["mutation","failure-modes"]`), "mutation shares a bundle"},
		"no primary":          {pjPolicy(dgpvPolicy), swap(`["primary","principles"]`, `["principles"]`), "no bundle carries primary"},
		"unknown visual word": {pjPolicy(dgpvPolicy), swap(`"skipped"`, `"maybe"`), "is not required, skipped or not configured"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			wt, root, out := dgpvFixture(t)
			r := runGuard("decide-groups-panel-visual", []string{wt, root, out}, pjEnv(t, c.policy, newPJStub(t, c.answer)))
			if r.rc != 0 || !strings.HasPrefix(r.stdout, "decided (fallback): 3 groups · panel primary+principles test-sanity+call-efficiency failure-modes mutation · visual required") ||
				!strings.Contains(r.err, c.cause) || strings.Count(r.err, "\n") != 1 {
				t.Fatalf("rc=%d stdout=%q err=%q, want the fallback and one stderr line naming %q", r.rc, r.stdout, r.err, c.cause)
			}
			if d := dgpvRead(t, out); d.Groups[2].Tasks[0] != "3" || d.Panel[3].Slots[0] != "mutation" || d.Visual.Reason != "fallback" {
				t.Errorf("fallback decision %+v", d)
			}
		})
	}
}

// TestDecideGroupsPanelVisualRefuses: usage and an unreadable or task-less
// plan exit 2 with nothing on stdout and no decision written.
func TestDecideGroupsPanelVisualRefuses(t *testing.T) {
	t.Parallel()
	wt, root, out := dgpvFixture(t)
	env := pjEnv(t, pjPolicy(dgpvPolicy), newPJStub(t, pjJSON(dgpvGood, "")))
	empty := t.TempDir()
	writeFile(t, empty+"/tasks.md", "# Tasks\n")
	for name, args := range map[string][]string{
		"usage":           {wt, root},
		"plan unreadable": {wt, t.TempDir(), out},
		"plan holds none": {wt, empty, out},
	} {
		if r := runGuard("decide-groups-panel-visual", args, env); r.rc != 2 || r.stdout != "" {
			t.Errorf("%s: rc=%d stdout=%q, want 2 and nothing", name, r.rc, r.stdout)
		}
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("a refusal wrote %s", out)
	}
}
