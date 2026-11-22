package guard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
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

const pcReport = "**F1** [finding] tasks.md:3 — a gap\n**F2** [advisory] design.md:1 — a nit\nPLAN-CHECK-FINDINGS 2\n"

func pcAnswer(report, decision string) string {
	return report + "===DECISION===\n" + decision + "\n"
}

// pcFixture is dgpvFixture plus the plan-check call's arguments.
func pcFixture(t *testing.T) (args []string, wt, out string) {
	t.Helper()
	wt, root, out := dgpvFixture(t)
	return []string{wt, root, out}, wt, out
}

// TestPlanCheckWritesReportAndDecision: the answer is split at the
// ===DECISION=== line; the report lands in the worktree's plan-check.md, the
// decision in today's schema.
func TestPlanCheckWritesReportAndDecision(t *testing.T) {
	t.Parallel()
	args, wt, out := pcFixture(t)
	stub := newPJStub(t, pjJSON(pcAnswer(pcReport, dgpvGood), ""))
	r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub))
	if r.rc != 0 {
		t.Fatalf("rc=%d stdout=%q err=%q", r.rc, r.stdout, r.err)
	}
	b, _ := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "plan-check.md"))
	if !strings.HasSuffix(strings.TrimSpace(string(b)), "PLAN-CHECK-FINDINGS 2") || strings.Contains(string(b), "===DECISION===") {
		t.Errorf("report %q", b)
	}
	if d := dgpvRead(t, out); len(d.Groups) != 2 || d.Visual.Verify != "skipped" || d.Overrides == nil {
		t.Errorf("decision %+v", d)
	}
	if !strings.Contains(r.stdout, "plan-check (judge):") {
		t.Errorf("stdout %q does not name the judge as the source", r.stdout)
	}
}

// TestPlanCheckFallsBackOnTimeout: a runner past the limit writes the fixed
// decision and a NOT-RUN report, exit 0.
func TestPlanCheckFallsBackOnTimeout(t *testing.T) {
	t.Parallel()
	args, wt, out := pcFixture(t)
	stub := newPJStub(t, "exec sleep 5\n")
	env := pjEnv(t, pjPolicy(dgpvPolicy), stub)
	env.PairJudgeTimeout = 200_000_000
	r := runGuard("plan-check", args, env)
	if r.rc != 0 {
		t.Fatalf("rc=%d err=%q", r.rc, r.err)
	}
	b, _ := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "plan-check.md"))
	if lines := strings.Split(strings.TrimSpace(string(b)), "\n"); lines[len(lines)-1] != "PLAN-CHECK-NOT-RUN" {
		t.Errorf("report %q", b)
	}
	if d := dgpvRead(t, out); d.Groups[0].Reason != "fallback" {
		t.Errorf("decision %+v", d)
	}
	if !strings.Contains(r.err, "using the fallback decision") {
		t.Errorf("stderr %q does not say the fallback decision is used", r.err)
	}
}

// TestPlanCheckPartialFallbacks: a bad last report line falls back the report
// alone; a rejected decision falls back the decision alone; no separator both.
func TestPlanCheckPartialFallbacks(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		answer                       string
		reportNotRun, decisionFallbk bool
	}{
		"bad last line":   {pcAnswer("findings\nmaybe\n", dgpvGood), true, false},
		"bad decision":    {pcAnswer(pcReport, `{"groups":[]}`), false, true},
		"no separator":    {pcReport + dgpvGood, true, true},
		"clean report ok": {pcAnswer("PLAN-CHECK-CLEAN\n", dgpvGood), false, false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			args, wt, out := pcFixture(t)
			stub := newPJStub(t, pjJSON(c.answer, ""))
			if r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub)); r.rc != 0 {
				t.Fatalf("rc=%d err=%q", r.rc, r.err)
			}
			b, _ := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "plan-check.md"))
			if got := strings.HasSuffix(strings.TrimSpace(string(b)), "PLAN-CHECK-NOT-RUN"); got != c.reportNotRun {
				t.Errorf("not-run=%v, want %v: %q", got, c.reportNotRun, b)
			}
			if got := dgpvRead(t, out).Groups[0].Reason == "fallback"; got != c.decisionFallbk {
				t.Errorf("decision fallback=%v, want %v", got, c.decisionFallbk)
			}
		})
	}
}

// TestPlanCheckDecideOnlyLeavesReport: -decide-only rewrites decision.json and
// leaves an existing plan-check.md byte-identical.
func TestPlanCheckDecideOnlyLeavesReport(t *testing.T) {
	t.Parallel()
	args, wt, out := pcFixture(t)
	report := filepath.Join(wt, ".superpowers", "sdd", "plan-check.md")
	writeFile(t, report, "kept\nPLAN-CHECK-CLEAN\n")
	stub := newPJStub(t, pjJSON(pcAnswer("other\nPLAN-CHECK-CLEAN\n", dgpvGood), ""))
	if r := runGuard("plan-check", append(args, "-decide-only"), pjEnv(t, pjPolicy(dgpvPolicy), stub)); r.rc != 0 {
		t.Fatalf("rc=%d err=%q", r.rc, r.err)
	}
	if b, _ := os.ReadFile(report); string(b) != "kept\nPLAN-CHECK-CLEAN\n" {
		t.Errorf("report changed: %q", b)
	}
	if d := dgpvRead(t, out); d.Groups[0].Reason == "fallback" {
		t.Errorf("decision %+v", d)
	}
	a := stub.argv(t)
	if i := slices.Index(a, "--tools"); i < 0 || a[i+1] != "" {
		t.Errorf("decide-only runs with tools: %q", a)
	}
	if i := slices.Index(a, "--system-prompt"); i < 0 || !strings.HasPrefix(a[i+1], "# Plan check — decision only") {
		t.Errorf("decide-only did not use the decision-only prompt: %q", a)
	}
	if i := slices.Index(a, "--system-prompt"); i < 0 || !strings.Contains(a[i+1], "**1. Implementer groups.**") {
		t.Errorf("decide-only prompt lacks the shared decision rules: %q", a)
	}
}

// TestPlanCheckRunsWithReadOnlyTools: the runner gets Read,Grep,Glob and the
// worktree as cwd, whatever directory the guard runs from.
func TestPlanCheckRunsWithReadOnlyTools(t *testing.T) {
	t.Parallel()
	args, wt, _ := pcFixture(t)
	cwd := filepath.Join(t.TempDir(), "cwd")
	stub := newPJStub(t, "pwd -P > '"+cwd+"'\n"+pjJSON(pcAnswer(pcReport, dgpvGood), ""))
	env := pjEnv(t, pjPolicy(dgpvPolicy), stub)
	env.Dir = t.TempDir()
	if r := runGuard("plan-check", args, env); r.rc != 0 {
		t.Fatalf("rc=%d err=%q", r.rc, r.err)
	}
	a := stub.argv(t)
	if i := slices.Index(a, "--tools"); i < 0 || a[i+1] != "Read,Grep,Glob" {
		t.Errorf("argv %q", a)
	}
	if i := slices.Index(a, "--system-prompt"); i < 0 || !strings.Contains(a[i+1], "**1. Implementer groups.**") {
		t.Errorf("plan-check prompt lacks the shared decision rules: %q", a)
	}
	want, _ := filepath.EvalSymlinks(wt)
	if b, _ := os.ReadFile(cwd); strings.TrimSpace(string(b)) != want {
		t.Errorf("cwd %q, want %q", b, want)
	}
}

// TestPlanCheckDecisionChecks: through plan-check -decide-only, a failed call or a decision breaking
// a rule writes the fallback decision, exit 0, one stderr line naming why.
func TestPlanCheckDecisionChecks(t *testing.T) {
	t.Parallel()
	swap := func(old, new string) string {
		return pjJSON(pcAnswer("PLAN-CHECK-CLEAN\n", strings.Replace(dgpvGood, old, new, 1)), "")
	}
	for name, c := range map[string]struct{ policy, answer, cause string }{
		"no pair":             {pjPolicy(`"pair-judge":{"base":["opus/flow-low"],"optin":[],"marked":[]}`), pjJSON("", ""), "no decider pair"},
		"command fails":       {pjPolicy(dgpvPolicy), "exit 3\n", "exit 3"},
		"not an object":       {pjPolicy(dgpvPolicy), pjJSON(pcAnswer("PLAN-CHECK-CLEAN\n", "no idea"), ""), "not a JSON object"},
		"task missing":        {pjPolicy(dgpvPolicy), swap(`["1","3"]`, `["1"]`), "task 3 is in no group"},
		"big task shares":     {pjPolicy(dgpvPolicy), swap(`["2"]`, `["2","3"]`), "big task 2 shares group 2"},
		"after order broken":  {pjPolicy(dgpvPolicy), pjJSON(pcAnswer("PLAN-CHECK-CLEAN\n", `{"groups":[{"tasks":["3"],"reason":"a"},{"tasks":["1"],"reason":"a"},{"tasks":["2"],"reason":"b"}],"panel":[{"slots":["primary"],"reason":"c"}],"visual":{"verify":"skipped","reason":"e"}}`), ""), "comes before its After task 1"},
		"mutation shares":     {pjPolicy(dgpvPolicy), swap(`["mutation"]`, `["mutation","failure-modes"]`), "mutation shares a bundle"},
		"no primary":          {pjPolicy(dgpvPolicy), swap(`["primary","principles"]`, `["principles"]`), "no bundle carries primary"},
		"unknown task":        {pjPolicy(dgpvPolicy), swap(`["1","3"]`, `["1","3","9"]`), "names task 9, which the plan lacks"},
		"task in two groups":  {pjPolicy(dgpvPolicy), swap(`["2"]`, `["3"]`), "task 3 is in two groups"},
		"too many slots":      {pjPolicy(dgpvPolicy), swap(`["primary","principles"]`, `["primary","principles","failure-modes"]`), "bundle 1 holds 3 slots"},
		"no slots":            {pjPolicy(dgpvPolicy), swap(`["mutation"]`, `[]`), "bundle 2 holds 0 slots"},
		"slot not in catalog": {pjPolicy(dgpvPolicy), swap(`["mutation"]`, `["bogus"]`), "bogus, which the catalog lacks"},
		"slot in two bundles": {pjPolicy(dgpvPolicy), swap(`["mutation"]`, `["primary"]`), "slot primary is in two bundles"},
		"visual disagrees":    {pjPolicy(dgpvPolicy), swap(`"skipped"`, `"not configured"`), "disagrees with the project"},
		"unknown visual word": {pjPolicy(dgpvPolicy), swap(`"skipped"`, `"maybe"`), "is not required, skipped or not configured"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			wt, root, out := dgpvFixture(t)
			r := runGuard("plan-check", []string{wt, root, out, "-decide-only"}, pjEnv(t, c.policy, newPJStub(t, c.answer)))
			if r.rc != 0 || !strings.Contains(r.stdout, "): unchanged · 3 groups · panel primary+principles test-sanity+call-efficiency failure-modes mutation · visual required") ||
				!strings.Contains(r.err, c.cause) || strings.Count(r.err, "\n") != 1 {
				t.Fatalf("rc=%d stdout=%q err=%q, want the fallback and one stderr line naming %q", r.rc, r.stdout, r.err, c.cause)
			}
			if d := dgpvRead(t, out); d.Groups[2].Tasks[0] != "3" || d.Panel[3].Slots[0] != "mutation" || d.Visual.Reason != "fallback" {
				t.Errorf("fallback decision %+v", d)
			}
		})
	}
}

// TestPlanCheckRefuses: usage and an unreadable or task-less
// plan exit 2 with nothing on stdout and no decision written.
func TestPlanCheckRefuses(t *testing.T) {
	t.Parallel()
	wt, root, out := dgpvFixture(t)
	env := pjEnv(t, pjPolicy(dgpvPolicy), newPJStub(t, pjJSON(pcAnswer("PLAN-CHECK-CLEAN\n", dgpvGood), "")))
	empty := t.TempDir()
	writeFile(t, empty+"/tasks.md", "# Tasks\n")
	for name, args := range map[string][]string{
		"usage":           {wt, root},
		"plan unreadable": {wt, t.TempDir(), out},
		"plan holds none": {wt, empty, out},
	} {
		if r := runGuard("plan-check", args, env); r.rc != 2 || r.stdout != "" {
			t.Errorf("%s: rc=%d stdout=%q, want 2 and nothing", name, r.rc, r.stdout)
		}
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("a refusal wrote %s", out)
	}
}

// TestPlanCheckNamesTheFallbackPart: a report with no verdict line falls back
// the report alone, and the stderr line says so.
func TestPlanCheckNamesTheFallbackPart(t *testing.T) {
	t.Parallel()
	args, _, _ := pcFixture(t)
	r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), newPJStub(t, pjJSON(pcAnswer("no verdict\n", dgpvGood), ""))))
	if r.rc != 0 || !strings.Contains(r.err, "fallback report —") {
		t.Fatalf("rc=%d err=%q", r.rc, r.err)
	}
}

// TestPlanCheckDigest: the model is shown design.md, the rubric and the visual declaration.
func TestPlanCheckDigest(t *testing.T) {
	t.Parallel()
	args, _, _ := pcFixture(t)
	stub := newPJStub(t, pjJSON(pcAnswer(pcReport, dgpvGood), ""))
	if r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub)); r.rc != 0 {
		t.Fatalf("rc=%d err=%q", r.rc, r.err)
	}
	got := stub.stdin(t)
	for _, want := range []string{"## design.md\n\n", "## plan-review-rubric.md\n\n", "`## visual verification`: yes\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("digest lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "plan-review-rubric.md\n\nnone") || strings.Contains(got, "## design.md\n\nnone") {
		t.Errorf("digest carries a missing file:\n%s", got)
	}
}

// TestDgpvInvalidGroupSizeLimit: a group of four tasks is refused.
func TestDgpvInvalidGroupSizeLimit(t *testing.T) {
	t.Parallel()
	var tasks []pdbTask
	var ids []string
	for i, id := range []string{"1", "2", "3", "4"} {
		tasks = append(tasks, pdbTask{id: id, size: "small", taskLine: i + 1})
		ids = append(ids, id)
	}
	d := dgpvDecision{Groups: []dgpvEntry{{Tasks: ids}}}
	if got := dgpvInvalid(d, tasks, make([]string, 6), map[string]bool{}, false); got != "group 1 holds 4 tasks" {
		t.Errorf("dgpvInvalid = %q", got)
	}
}

// TestPlanCheckAddsTheSecuritySlotOnAMatchingFile (kan-1125): a task file
// matching `## security paths` adds a security bundle to the decision whatever
// the decider answered; no match, or no section, adds none.
func TestPlanCheckAddsTheSecuritySlotOnAMatchingFile(t *testing.T) {
	t.Parallel()
	plan := "# Tasks\n\n- [ ] 1. first\n\n**Files:** `api/auth/login.go`\n"
	for name, c := range map[string]struct {
		project string
		want    bool
	}{
		"match":      {"## visual verification\n\n| Setting | Value |\n\n## security paths\n\napi/auth/**, api/secrets/**\n", true},
		"no match":   {"## visual verification\n\n| Setting | Value |\n\n## security paths\n\napi/billing/**\n", false},
		"no section": {"## visual verification\n\n| Setting | Value |\n", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			wt, root, out := dgpvFixture(t)
			writeFile(t, wt+"/.flow/project.md", c.project)
			writeFile(t, root+"/tasks.md", plan)
			good := `{"groups":[{"tasks":["1"],"reason":"a"}],"panel":[{"slots":["primary"],"reason":"c"}],"visual":{"verify":"skipped","reason":"e"}}`
			stub := newPJStub(t, pjJSON(pcAnswer(pcReport, good), ""))
			if r := runGuard("plan-check", []string{wt, root, out}, pjEnv(t, pjPolicy(dgpvPolicy), stub)); r.rc != 0 {
				t.Fatalf("rc=%d err=%q", r.rc, r.err)
			}
			has := false
			for _, b := range dgpvRead(t, out).Panel {
				has = has || slices.Contains(b.Slots, "security")
			}
			if has != c.want {
				t.Errorf("security bundle present = %v, want %v", has, c.want)
			}
		})
	}
}

const pcConfReport = "**F1** [finding] tasks.md:3 — a gap\n\n### Execution confidence\n\nCONFIDENCE: 78%\nLowered by: F1\n\n" +
	"**F2** [finding] tasks.md:9 — closes with: plan — no task checks the flag — the evidence\n" +
	"**F3** [advisory] design.md:1 — closes with: running it — the migration is untried\n\nPLAN-CHECK-FINDINGS 3\n"

// TestPlanCheckWritesConfidence: a well-formed CONFIDENCE line under the
// report's Execution confidence section becomes confidence.json — the score
// and every line after it, verbatim — and the stdout line names the number.
func TestPlanCheckWritesConfidence(t *testing.T) {
	t.Parallel()
	args, wt, _ := pcFixture(t)
	stub := newPJStub(t, pjJSON(pcAnswer(pcConfReport, dgpvGood), ""))
	r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub))
	if r.rc != 0 || !strings.Contains(r.stdout, " · confidence 78%") {
		t.Fatalf("rc=%d stdout=%q err=%q", r.rc, r.stdout, r.err)
	}
	b, err := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "confidence.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Score int      `json:"score"`
		Gaps  []string `json:"gaps"`
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("%v: %s", err, b)
	}
	want := []string{"Lowered by: F1",
		"**F2** [finding] tasks.md:9 — closes with: plan — no task checks the flag — the evidence",
		"**F3** [advisory] design.md:1 — closes with: running it — the migration is untried"}
	if got.Score != 78 || !slices.Equal(got.Gaps, want) {
		t.Errorf("confidence.json = %s", b)
	}
}

// TestPlanCheckDropsStaleConfidence: a report with no well-formed CONFIDENCE
// line removes a confidence.json an earlier check left, says `confidence
// unknown`, and still writes the report and exits 0.
func TestPlanCheckDropsStaleConfidence(t *testing.T) {
	t.Parallel()
	for name, report := range map[string]string{
		"no section":   pcReport,
		"not a number": "### Execution confidence\n\nCONFIDENCE: high\nPLAN-CHECK-CLEAN\n",
		"over 100":     "### Execution confidence\n\nCONFIDENCE: 140%\nPLAN-CHECK-CLEAN\n",
		"zero":         "### Execution confidence\n\nCONFIDENCE: 0%\nPLAN-CHECK-CLEAN\n",
		"no heading":   "CONFIDENCE: 80%\nPLAN-CHECK-CLEAN\n",
	} {
		args, wt, _ := pcFixture(t)
		conf := filepath.Join(wt, ".superpowers", "sdd", "confidence.json")
		writeFile(t, conf, `{"score":50,"gaps":[]}`)
		stub := newPJStub(t, pjJSON(pcAnswer(report, dgpvGood), ""))
		r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub))
		if r.rc != 0 || !strings.Contains(r.stdout, " · confidence unknown") {
			t.Errorf("%s: rc=%d stdout=%q err=%q", name, r.rc, r.stdout, r.err)
		}
		if _, err := os.Stat(conf); !os.IsNotExist(err) {
			t.Errorf("%s: stale confidence.json survived (%v)", name, err)
		}
		if b, _ := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "plan-check.md")); string(b) != report {
			t.Errorf("%s: report = %q", name, b)
		}
	}
}

// TestPlanCheckDecideOnlyLeavesConfidence: -decide-only neither rewrites nor
// removes confidence.json and prints no confidence token.
func TestPlanCheckDecideOnlyLeavesConfidence(t *testing.T) {
	t.Parallel()
	args, wt, _ := pcFixture(t)
	conf := filepath.Join(wt, ".superpowers", "sdd", "confidence.json")
	writeFile(t, conf, `{"score":50,"gaps":["kept"]}`)
	stub := newPJStub(t, pjJSON(pcAnswer(pcConfReport, dgpvGood), ""))
	r := runGuard("plan-check", append(args, "-decide-only"), pjEnv(t, pjPolicy(dgpvPolicy), stub))
	if r.rc != 0 || strings.Contains(r.stdout, "confidence") {
		t.Fatalf("rc=%d stdout=%q err=%q", r.rc, r.stdout, r.err)
	}
	if b, _ := os.ReadFile(conf); string(b) != `{"score":50,"gaps":["kept"]}` {
		t.Errorf("confidence.json changed: %q", b)
	}
}
