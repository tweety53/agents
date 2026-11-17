package guard

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckPlanDepartures pins check-plan-departures (kan-1099: its plan
// restyled shipped screens to new frames with none measured first, and the
// departures surfaced one verify round at a time -- 20 rounds, 16 appended
// tasks): a plan naming the sidecar of a spec that already has captures owes
// a plan-time measure of it, and cites every `over` row that measure found.
func TestCheckPlanDepartures(t *testing.T) {
	t.Parallel()
	const day = "- [ ] 1. Restyle the day view\n\n**Files:** `web/day.tsx`, `e2e/day.spec.ts.mockups`\n"
	manifest := func(spec string) string {
		return `{"pairs":[{"spec":"/x/e2e/` + spec + `","frame":"v2/D1","over":[{"row":2},{"row":15}]},{"spec":"/x/e2e/` + spec + `","frame":"v2/D2","over":[]}]}`
	}
	for _, tc := range []struct {
		name, plan, manifest string
		shipped              bool
		rc                   int
		want                 string
	}{
		{"a plan naming no sidecar owes nothing", "- [ ] 1. One\n\n**Files:** `a.go`\n", "", true, 0, "PLAN-DEPARTURES-CITED: 0 row(s) of 0 measured spec(s)"},
		{"a new view has no capture to measure", day, "", false, 0, "PLAN-DEPARTURES-CITED: 0 row(s) of 0 measured spec(s)"},
		{"a shipped view nobody measured", day, "", true, 1, "PLAN-MEASURE-MISSING: day.spec.ts"},
		{"a ticked task owes nothing", "`e2e/day.spec.ts.mockups` above every task\n" + strings.Replace(day, "- [ ]", "- [x]", 1), "", true, 0, "PLAN-DEPARTURES-CITED: 0 row(s) of 0 measured spec(s)"},
		{"a measure of another spec covers nothing", day, manifest("night.spec.ts"), true, 1, "PLAN-MEASURE-MISSING: day.spec.ts"},
		{"an over row no task cites", day + "Close `v2/D1 #2` and v2/D1 #1.\n", manifest("day.spec.ts"), true, 1, "PLAN-DEPARTURES-UNCITED: v2/D1 #15"},
		{"every over row cited", day + "Close v2/D1 #2 and `v2/D1 #15`.\n", manifest("day.spec.ts"), true, 0, "PLAN-DEPARTURES-CITED: 2 row(s) of 1 measured spec(s)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFile(t, filepath.Join(root, ".flow/project.md"), "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/**` |\n| `screenshots` | `e2e` |\n")
			if tc.shipped {
				writeFile(t, filepath.Join(root, "e2e/day.spec.ts-snapshots/day-darwin.png"), "png")
			}
			if tc.manifest != "" {
				writeFile(t, filepath.Join(root, ".superpowers/sdd/visual-manifest-plan.json"), tc.manifest)
			}
			plan := filepath.Join(root, "tasks.md")
			writeFile(t, plan, tc.plan)
			r := runGuard("check-plan-departures", []string{root, "demo", plan}, Env{Dir: root})
			if r.rc != tc.rc || !strings.Contains(r.stdout, tc.want+"\n") {
				t.Errorf("exit %d, want %d and %q\n%s", r.rc, tc.rc, tc.want, r.out)
			}
			if tc.rc == 1 && strings.Contains(tc.want, "MISSING") && !strings.Contains(r.err, "flow-guard visual-verify-pass measure "+root+" demo plan <spec path>") {
				t.Errorf("stderr lacks the measure call\n%s", r.err)
			}
		})
	}
	root := t.TempDir()
	for name, args := range map[string][]string{"usage": {root, "demo"}, "an unreadable plan": {root, "demo", filepath.Join(root, "none.md")}} {
		if r := runGuard("check-plan-departures", args, Env{Dir: root}); r.rc != 2 {
			t.Errorf("%s: exit %d, want 2\n%s", name, r.rc, r.out)
		}
	}
	writeFile(t, filepath.Join(root, "tasks.md"), day)
	writeFile(t, filepath.Join(root, ".superpowers/sdd/visual-manifest-plan.json"), "{")
	if r := runGuard("check-plan-departures", []string{root, "demo", filepath.Join(root, "tasks.md")}, Env{Dir: root}); r.rc != 2 {
		t.Errorf("an unparsable manifest: exit %d, want 2\n%s", r.rc, r.out)
	}
}
