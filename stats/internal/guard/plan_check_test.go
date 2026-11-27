package guard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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

const pcReport = "**F1** [finding] tasks.md:3 — a gap\n**F2** [advisory] design.md:1 — a nit\nPLAN-CHECK-FINDINGS 1\n"

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
	if !strings.HasSuffix(strings.TrimSpace(string(b)), "PLAN-CHECK-FINDINGS 1") || strings.Contains(string(b), "===DECISION===") {
		t.Errorf("report %q", b)
	}
	if d := dgpvRead(t, out); len(d.Groups) != 2 || d.Visual.Verify != "skipped" || d.Overrides == nil {
		t.Errorf("decision %+v", d)
	}
	if !strings.Contains(r.stdout, "plan-check (judge):") {
		t.Errorf("stdout %q does not name the judge as the source", r.stdout)
	}
	if !strings.Contains(r.stdout, " · 2 groups · gated review on · ") {
		t.Errorf("stdout %q does not turn the gated review on for two groups", r.stdout)
	}
}

// TestPlanCheckBlocksOnTimeout: a runner past the limit, asked twice, blocks: exit 1, one stderr line, nothing written.
func TestPlanCheckBlocksOnTimeout(t *testing.T) {
	t.Parallel()
	args, wt, out := pcFixture(t)
	stub := newPJStub(t, "exec sleep 5\n")
	env := pjEnv(t, pjPolicy(dgpvPolicy), stub)
	env.PairJudgeTimeout = 200_000_000
	r := runGuard("plan-check", args, env)
	if r.rc != 1 || r.stdout != "" || !strings.Contains(r.err, "the run is blocked") {
		t.Fatalf("rc=%d stdout=%q err=%q, want 1, nothing on stdout and the block named", r.rc, r.stdout, r.err)
	}
	for _, p := range []string{out, filepath.Join(wt, ".superpowers", "sdd", "plan-check.md")} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("a blocked check wrote %s", p)
		}
	}
}

// TestPlanCheckPartialFallbacks: a bad last report line falls back the report
// alone; a rejected decision or no separator blocks.
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
			r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub))
			if c.decisionFallbk { // a bad decision blocks, never falls back
				if _, err := os.Stat(out); r.rc != 1 || !os.IsNotExist(err) {
					t.Errorf("rc=%d err=%q, want 1 and no decision written", r.rc, r.err)
				}
				return
			}
			if r.rc != 0 {
				t.Fatalf("rc=%d err=%q", r.rc, r.err)
			}
			b, _ := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "plan-check.md"))
			if got := strings.HasSuffix(strings.TrimSpace(string(b)), "PLAN-CHECK-NOT-RUN"); got != c.reportNotRun {
				t.Errorf("not-run=%v, want %v: %q", got, c.reportNotRun, b)
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
	if d := dgpvRead(t, out); strings.HasPrefix(d.Groups[0].Reason, "fallback") {
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
// a rule blocks: exit 1, nothing written, one stderr line naming why.
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
			if r.rc != 1 || r.stdout != "" || !strings.Contains(r.err, c.cause) || strings.Count(r.err, "\n") != 1 {
				t.Fatalf("rc=%d stdout=%q err=%q, want a block and one stderr line naming %q", r.rc, r.stdout, r.err, c.cause)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Errorf("a blocked check wrote %s", out)
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

// TestPlanCheckDigestKeepsTheRubricOnAnOversizedPlan: a design.md past the digest cap truncates the
// artifacts' tail, never the rubric or the visual declaration.
func TestPlanCheckDigestKeepsTheRubricOnAnOversizedPlan(t *testing.T) {
	t.Parallel()
	args, _, _ := pcFixture(t)
	writeFile(t, args[1]+"/design.md", strings.Repeat("x", pjDigestCap+10))
	stub := newPJStub(t, pjJSON(pcAnswer(pcReport, dgpvGood), ""))
	if r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub)); r.rc != 0 {
		t.Fatalf("rc=%d err=%q", r.rc, r.err)
	}
	got := stub.stdin(t)
	for _, want := range []string{"## plan-review-rubric.md\n\n", "`## visual verification`: yes\n", pjTruncated} {
		if !strings.Contains(got, want) {
			t.Errorf("digest lacks %q", want)
		}
	}
}

// TestDgpvCatalogMatchesTheDecisionTable: the slot catalog the decision is checked against is the
// table plan-check-decision.md shows the model, row for row.
func TestDgpvCatalogMatchesTheDecisionTable(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "skills", "flow", "plan-check-decision.md"))
	if err != nil {
		t.Fatal(err)
	}
	var table []string
	for _, m := range regexp.MustCompile("(?m)^\\| `([a-z][a-z0-9-]*)` \\|").FindAllStringSubmatch(string(b), -1) {
		table = append(table, m[1])
	}
	if !slices.Equal(table, dgpvSlots) {
		t.Errorf("plan-check-decision.md's slots %q, dgpvSlots %q", table, dgpvSlots)
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
	if got := dgpvInvalid(d, tasks, make([]string, 6), false); got != "group 1 holds 4 tasks" {
		t.Errorf("dgpvInvalid = %q", got)
	}
}

// TestDgpvInvalidMutationNeedsTwoGroups: the review threshold — a decision of one implementer group
// and no big task turns the gated per-task reviewer off (its pass would re-read the diff the panel's primary reads)
// and refuses the mutation slot; two groups turn both on.
func TestDgpvInvalidMutationNeedsTwoGroups(t *testing.T) {
	t.Parallel()
	tasks := []pdbTask{{id: "1", size: "small", taskLine: 1}, {id: "2", size: "small", taskLine: 2}}
	panel := []dgpvEntry{{Slots: []string{"primary"}}, {Slots: []string{"mutation"}}}
	one := dgpvDecision{Groups: []dgpvEntry{{Tasks: []string{"1", "2"}}}, Panel: panel}
	one.Visual.Verify = "not configured"
	two := one
	two.Groups = []dgpvEntry{{Tasks: []string{"1"}}, {Tasks: []string{"2"}}}
	if dgpvGatedReview(one, tasks) || !dgpvGatedReview(two, tasks) {
		t.Errorf("gated review: one group %v, two groups %v; want off, on", dgpvGatedReview(one, tasks), dgpvGatedReview(two, tasks))
	}
	if got := dgpvInvalid(one, tasks, make([]string, 4), false); got != "mutation needs more than one group or a big task" {
		t.Errorf("one group: dgpvInvalid = %q", got)
	}
	if got := dgpvInvalid(two, tasks, make([]string, 4), false); got != "" {
		t.Errorf("two groups: dgpvInvalid = %q", got)
	}
}

// TestDgpvInvalidMutationAdmitsABigTask: one group holding a `**Size:** big` task is large enough
// for the gated reviewer and the mutation slot.
func TestDgpvInvalidMutationAdmitsABigTask(t *testing.T) {
	t.Parallel()
	tasks := []pdbTask{{id: "1", size: "big", taskLine: 1}}
	d := dgpvDecision{Groups: []dgpvEntry{{Tasks: []string{"1"}}}, Panel: []dgpvEntry{{Slots: []string{"primary"}}, {Slots: []string{"mutation"}}}}
	d.Visual.Verify = "not configured"
	if got := dgpvInvalid(d, tasks, make([]string, 3), false); got != "" {
		t.Errorf("one group with a big task: dgpvInvalid = %q", got)
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
// and every line after it — a `Lowered by: F<k>` line spelled out as the finding it cites, the rest verbatim — and the stdout line names the number.
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
	want := []string{"Lowered by F1 — tasks.md:3 — a gap",
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

// pcSeq is a judge stub answering its first call with first and every later one with then, counting calls.
func pcSeq(t *testing.T, first, then string) pjStub {
	t.Helper()
	body := "n=$(cat \"$(dirname \"$0\")/count\" 2>/dev/null || echo 0)\necho $((n+1)) > \"$(dirname \"$0\")/count\"\n" +
		"if [ \"$n\" = 0 ]; then\n" + first + "else\n" + then + "fi\n"
	return newPJStub(t, body)
}

func (s pjStub) calls() string {
	b, _ := os.ReadFile(s.dir + "/count")
	return strings.TrimSpace(string(b))
}

// TestPlanCheckRetriesAReportWithNoTerminalLine (KAN-1130): the first answer whose report ends in no
// PLAN-CHECK-* line is not accepted — the judge is asked once more with the terminal line named, and
// a usable second answer is the result.
func TestPlanCheckRetriesAReportWithNoTerminalLine(t *testing.T) {
	t.Parallel()
	args, wt, _ := pcFixture(t)
	stub := pcSeq(t, pjJSON(pcAnswer("some prose, no terminal line\n", dgpvGood), ""), pjJSON(pcAnswer(pcConfReport, dgpvGood), ""))
	r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub))
	if r.rc != 0 || stub.calls() != "2" || !strings.Contains(strings.Join(stub.argv(t), " "), "must be exactly PLAN-CHECK-CLEAN or PLAN-CHECK-FINDINGS") ||
		!strings.Contains(r.stdout, "plan-check (judge):") || !strings.Contains(r.stdout, " · confidence 78%") {
		t.Fatalf("rc=%d calls=%s stdout=%q err=%q", r.rc, stub.calls(), r.stdout, r.err)
	}
	if b, _ := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "plan-check.md")); !strings.Contains(string(b), "CONFIDENCE: 78%") {
		t.Errorf("report %q is not the second answer", b)
	}
	// A first answer that already ends right is asked once.
	args2, _, _ := pcFixture(t)
	one := pcSeq(t, pjJSON(pcAnswer(pcReport, dgpvGood), ""), pjJSON(pcAnswer("bad\n", dgpvGood), ""))
	if r := runGuard("plan-check", args2, pjEnv(t, pjPolicy(dgpvPolicy), one)); r.rc != 0 || one.calls() != "1" {
		t.Errorf("a good first answer: rc=%d calls=%s", r.rc, one.calls())
	}
}

// TestPlanCheckAsksTheModelAtMostTwice: answers with no terminal line and no usable decision cost two
// model calls in all, never four, and the second names why the first could not be used.
func TestPlanCheckAsksTheModelAtMostTwice(t *testing.T) {
	t.Parallel()
	args, _, _ := pcFixture(t)
	bad := pjJSON(pcAnswer("no terminal line\n", `{"groups":[]}`), "")
	stub := pcSeq(t, bad, bad)
	r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub))
	if r.rc != 1 || stub.calls() != "2" || !strings.Contains(strings.Join(stub.argv(t), " "), "Your previous answer could not be used: plan-check decision rejected") {
		t.Fatalf("rc=%d calls=%s err=%q argv=%q", r.rc, stub.calls(), r.err, stub.argv(t))
	}
}

// TestPlanCheckDerivesTheTerminalLine: after the retry, a body that still ends in no terminal line but
// holds finding lines or a confidence score is parsed, never discarded.
func TestPlanCheckDerivesTheTerminalLine(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct{ body, last, conf string }{
		"findings":   {"**F1** [finding] tasks.md:3 — a gap\n**F2** [advisory] design.md:1 — a nit\n", "PLAN-CHECK-FINDINGS 1", " · confidence unknown"},
		"confidence": {"### Execution confidence\n\nCONFIDENCE: 90%\n", "PLAN-CHECK-CLEAN", " · confidence 90%"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			args, wt, _ := pcFixture(t)
			bad := pjJSON(pcAnswer(c.body, dgpvGood), "")
			stub := pcSeq(t, bad, bad)
			r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub))
			b, _ := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "plan-check.md"))
			if r.rc != 0 || stub.calls() != "2" || !strings.HasSuffix(strings.TrimSpace(string(b)), c.last) || !strings.Contains(r.stdout, c.last+" ") || !strings.Contains(r.stdout, c.conf) {
				t.Fatalf("rc=%d calls=%s report=%q stdout=%q", r.rc, stub.calls(), b, r.stdout)
			}
		})
	}
}

// TestPlanCheckAdvisoryIsNotAFinding: an `advisory` line is an idea, never owed a fix, so it never
// counts toward `PLAN-CHECK-FINDINGS` — the terminal line the triage dispatch keys on — whether
// the answer's own terminal line counted it or the guard derives one.
func TestPlanCheckAdvisoryIsNotAFinding(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct{ body, last string }{
		"model counted it":  {"**F1** [advisory] design.md:1 — an idea\nCONFIDENCE: 100%\nPLAN-CHECK-FINDINGS 1\n", "PLAN-CHECK-CLEAN"},
		"derived":           {"**F1** [advisory] design.md:1 — an idea\nCONFIDENCE: 100%\n", "PLAN-CHECK-CLEAN"},
		"mixed, model said": {"**F1** [finding] tasks.md:3 — a gap\n**F2** [advisory] design.md:1 — an idea\nPLAN-CHECK-FINDINGS 2\n", "PLAN-CHECK-FINDINGS 1"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			args, wt, _ := pcFixture(t)
			ans := pjJSON(pcAnswer(c.body, dgpvGood), "")
			r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), pcSeq(t, ans, ans)))
			b, _ := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "plan-check.md"))
			if r.rc != 0 || !strings.HasSuffix(strings.TrimSpace(string(b)), c.last) {
				t.Fatalf("rc=%d report=%q stdout=%q", r.rc, b, r.stdout)
			}
		})
	}
}

// TestPlanCheckNoEstimateLine: two answers with no terminal line and nothing to derive one from give the
// explicit `no confidence estimate` line, in the report and on stdout, never a bare not-run.
func TestPlanCheckNoEstimateLine(t *testing.T) {
	t.Parallel()
	args, wt, _ := pcFixture(t)
	bad := pjJSON(pcAnswer("just thoughts\n", dgpvGood), "")
	stub := pcSeq(t, bad, bad)
	r := runGuard("plan-check", args, pjEnv(t, pjPolicy(dgpvPolicy), stub))
	b, _ := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "plan-check.md"))
	if r.rc != 0 || stub.calls() != "2" || !strings.Contains(r.stdout, "PLAN-CHECK-NOT-RUN — no confidence estimate") || strings.Contains(r.stdout, "confidence unknown") ||
		!strings.Contains(string(b), "no confidence estimate exists") {
		t.Fatalf("rc=%d calls=%s stdout=%q report=%q", r.rc, stub.calls(), r.stdout, b)
	}
}

// TestPlanCheckDecideOnlySkipsAnUnchangedShape: after a full check, -decide-only keeps decision.json
// without a model call while tasks.md keeps its tasks and their Files:, After:, Size: and
// repository; a task added or removed, or one of those fields changed, calls the decider again.
func TestPlanCheckDecideOnlySkipsAnUnchangedShape(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		plan string
		asks bool
	}{
		"prose only":   {strings.Replace(dgpvPlan, "first", "first, reworded", 1), false},
		"task added":   {dgpvPlan + "\n- [ ] 4. fourth\n", true},
		"task removed": {strings.Replace(dgpvPlan, "- [ ] 3. third\n\n**After:** Task 1\n", "", 1), true},
		"files added":  {strings.Replace(dgpvPlan, "**Size:** small\n", "**Size:** small\n\n**Files:**\n- a.go\n", 1), true},
		"after moved":  {strings.Replace(dgpvPlan, "**After:** Task 1", "**After:** Task 2", 1), true},
		"size changed": {strings.Replace(dgpvPlan, "**Size:** small", "**Size:** big", 1), true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			wt, root, out := dgpvFixture(t)
			stub := pcSeq(t, pjJSON(pcAnswer(pcReport, dgpvGood), ""), "exit 3\n")
			env := pjEnv(t, pjPolicy(dgpvPolicy), stub)
			if r := runGuard("plan-check", []string{wt, root, out}, env); r.rc != 0 {
				t.Fatalf("full check rc=%d err=%q", r.rc, r.err)
			}
			before, _ := os.ReadFile(out)
			writeFile(t, root+"/tasks.md", c.plan)
			r := runGuard("plan-check", []string{wt, root, out, "-decide-only"}, env)
			if asked := stub.calls() == "3"; asked != c.asks {
				t.Fatalf("decider asked=%v (calls %s), want %v; rc=%d stdout=%q err=%q", asked, stub.calls(), c.asks, r.rc, r.stdout, r.err)
			}
			if c.asks {
				return
			}
			if after, _ := os.ReadFile(out); r.rc != 0 || string(after) != string(before) || !strings.Contains(r.stdout, "plan-check (unchanged):") {
				t.Errorf("rc=%d stdout=%q err=%q, decision kept=%v", r.rc, r.stdout, r.err, string(after) == string(before))
			}
		})
	}
}
