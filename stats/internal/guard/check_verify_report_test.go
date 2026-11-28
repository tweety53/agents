package guard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

func TestCheckVerifyReport(t *testing.T) {
	t.Parallel()
	const allDone = "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — no motions named\n"
	cases := []struct {
		name    string
		report  string // "" writes no file
		motions string
		rc      int
		want    string // substring of the verdict line
	}{
		{"every step done", "## Report\n" + allDone, "0", 0, "VERIFY-REPORT-COMPLETE"},
		{"indented steps line", "## Report\n  " + allDone, "0", 0, "VERIFY-REPORT-COMPLETE"},
		{"setup and mockups n/a", "- steps: 4 n/a — setup not declared | 7 done | 8 done | 9 done | 10 n/a — mockups not declared | 11 done | motion n/a — no motions named\n", "0", 0, "COMPLETE"},
		{"blocked on a cited exit", "- steps: 4 done | 7 blocked — verify exit 1 | 8 blocked — step 7 exit 1 | 9 blocked — step 7 exit 1 | 10 blocked — step 7 exit 1 | 11 done | motion n/a — no motions named\n", "0", 0, "COMPLETE"},
		{"motions recorded", "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion done\n- motion: 2/2 — swipe: clean; exit: clean\n", "2", 0, "COMPLETE"},
		// KAN-870: the verifier ran tests and capture and listed 9-11 as not done.
		{"steps 9-11 not done", "- steps: 4 done | 7 done | 8 done | 9 not done | 10 not done | 11 not done | motion n/a — no motions named\n", "0", 1, "9 not done; 10 not done; 11 not done"},
		{"no steps line", "## Report\n- verify: exit 0\n", "0", 1, "4 absent; 8 absent; 9 absent; 10 absent; 11 absent; motion absent"},
		{"a step left out", "- steps: 4 done | 7 done | 8 done | 10 done | 11 done | motion n/a — none\n", "0", 1, "9 absent"},
		{"n/a on a required step", "- steps: 4 done | 7 done | 8 done | 9 n/a — out of time | 10 done | 11 done | motion n/a — none\n", "0", 1, "9 n/a — out of time"},
		{"n/a with no reason", "- steps: 4 n/a | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n", "0", 1, "4 n/a"},
		// KAN-1000: step 10 declared blocked on another step's exit while step 8 captured.
		{"step 10 blocked with captures on disk", "- steps: 4 done | 7 blocked — verify exit 1 | 8 done | 9 done | 10 blocked — verify exit 1 | 11 done | motion n/a — none\n", "0", 1, "10 blocked — verify exit 1 (step 8 captured)"},
		{"step 10 blocked on compose's own exit", "- steps: 4 done | 7 done | 8 done | 9 done | 10 blocked — compose-mockup-frames exit 1 | 11 done | motion n/a — none\n", "0", 0, "COMPLETE"},
		{"step 10 n/a without naming mockups", "- steps: 4 done | 7 done | 8 done | 9 done | 10 n/a — out of time | 11 done | motion n/a — none\n", "0", 1, "10 n/a — out of time"},
		{"step 10 n/a, no mockups declared", "- steps: 4 done | 7 done | 8 done | 9 done | 10 n/a — `mockups` not declared | 11 done | motion n/a — none\n", "0", 0, "COMPLETE"},
		{"blocked with no cited exit", "- steps: 4 done | 7 done | 8 done | 9 blocked — ran out of context | 10 done | 11 done | motion n/a — none\n", "0", 1, "9 blocked — ran out of context"},
		{"blocked on exit 0", "- steps: 4 done | 7 done | 8 done | 9 blocked — capture exit 0 | 10 done | 11 done | motion n/a — none\n", "0", 1, "9 blocked — capture exit 0"},
		{"motions named, step n/a", allDone, "1", 1, "motion n/a — no motions named"},
		{"motions named, no motion line", "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion done\n", "2", 1, "motion: no `- motion: 2/2` line"},
		{"motions named, one entry short", "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion done\n- motion: 1/2 — swipe: clean\n", "2", 1, "motion: no `- motion: 2/2` line"},
		// kan-1022: a wire-contract dispatch names no view, so 8-10 are n/a, never `done`.
		{"steps 8-10 n/a beside a wire contract", "- wire contract: GET /api/templates/{id} — shape matched\n- steps: 4 done | 7 done | 8 n/a — wire contract, no view drawn | 9 n/a — wire contract, no view drawn | 10 n/a — wire contract, no view drawn | 11 done | motion n/a — none\n", "0", 0, "COMPLETE"},
		{"steps 8-9 n/a with no wire contract", "- wire contract: n/a — no wire contract named\n- steps: 4 done | 7 done | 8 n/a — no view drawn | 9 n/a — no view drawn | 10 done | 11 done | motion n/a — none\n", "0", 1, "8 n/a — no view drawn; 9 n/a — no view drawn"},
		{"missing report", "", "0", 2, "cannot read"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "verify-report-visual-verify.md")
			if tc.report != "" {
				if err := os.WriteFile(path, []byte("## Report\n"+strings.TrimPrefix(tc.report, "## Report\n")), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			r := runGuard("check-verify-report", []string{path, tc.motions}, core.Env{Getenv: os.Getenv})
			if r.Rc != tc.rc || !strings.Contains(r.Out, tc.want) {
				t.Fatalf("rc=%d out=%q; want rc=%d containing %q", r.Rc, r.Out, tc.rc, tc.want)
			}
		})
	}
	for _, args := range [][]string{nil, {"x"}, {"x", "-1"}, {"x", "two"}, {"x", "0", "2"}} {
		if r := runGuard("check-verify-report", args, core.Env{Getenv: os.Getenv}); r.Rc != 2 || !strings.Contains(r.Out, "usage") {
			t.Fatalf("args %q: rc=%d out=%q; want usage exit 2", args, r.Rc, r.Out)
		}
	}
}

// kan-1017 (sr12): every verifier wrote `no mockup source` while design.md
// cited a handoff page that existed, and left 7 of 11 matrix columns n/a.
func TestCheckVerifyReportNoMockupSourceWithCitedSource(t *testing.T) {
	t.Parallel()
	const steps = "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n"
	for _, tc := range []struct {
		name                  string
		inReport, inVV, cited bool
		design                bool
		rc                    int
		want                  string
	}{
		{"report says it", true, false, true, true, 1, "10 `no mockup source` while design.md cites handoff/Flow.html"},
		{"visual-verification.md says it", false, true, true, true, 1, "handoff/Flow.html"},
		{"cited file absent", true, false, false, true, 0, "COMPLETE"},
		{"no --design", true, false, true, false, 0, "COMPLETE"},
	} {
		wt := t.TempDir()
		root := filepath.Join(wt, "spectre/changes/x")
		guardtest.WriteFile(t, filepath.Join(root, "design.md"), "Mockups: `handoff/Flow.html`.\n")
		if tc.cited {
			guardtest.WriteFile(t, filepath.Join(wt, "handoff/Flow.html"), "<html>")
		}
		if tc.inVV {
			guardtest.WriteFile(t, filepath.Join(root, "visual-verification.md"), "- AD1: no mockup source\n")
		}
		body := "## Report\n" + steps
		if tc.inReport {
			body += "- matrix: no mockup source\n"
		}
		path := filepath.Join(wt, ".superpowers/sdd/verify-report-visual-verify.md")
		guardtest.WriteFile(t, path, body)
		args := []string{path, "0"}
		if tc.design {
			args = append(args, "--design", filepath.Join(root, "design.md"))
		}
		r := runGuard("check-verify-report", args, core.Env{Getenv: os.Getenv})
		if r.Rc != tc.rc || !strings.Contains(r.Out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.Rc, r.Out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportStep10WithACitedPrototype pins that step 10 reported
// n/a beside an approved prototype design.md cites is undone: the prototype is
// rendered as each capture's frame (kan-1038).
func TestCheckVerifyReportStep10WithACitedPrototype(t *testing.T) {
	t.Parallel()
	const na = "- steps: 4 done | 7 done | 8 done | 9 done | 10 n/a — mockups not declared | 11 done | motion n/a — none\n"
	const wire = "- wire contract: GET /api/x — shape matched\n- steps: 4 done | 7 done | 8 n/a — wire contract, no view drawn | 9 n/a — wire contract, no view drawn | 10 n/a — wire contract, no view drawn | 11 done | motion n/a — none\n"
	for _, tc := range []struct {
		name, cite, report string
		exists, design     bool
		rc                 int
		want               string
	}{
		{"cited prototype", "proto.html", na, true, true, 1, "10 n/a while design.md cites the prototype .superpowers/sdd/proto.html"},
		{"cited file absent", "proto.html", na, false, true, 0, "COMPLETE"},
		{"no --design", "proto.html", na, true, false, 0, "COMPLETE"},
		{"css only", "proto.css", na, true, true, 0, "COMPLETE"},
		{"wire contract", "proto.html", wire, true, true, 0, "COMPLETE"},
	} {
		wt := t.TempDir()
		root := filepath.Join(wt, "spectre/changes/x")
		guardtest.WriteFile(t, filepath.Join(root, "design.md"), "## Context\n\nPrototype: `.superpowers/sdd/"+tc.cite+"`.\n")
		if tc.exists {
			guardtest.WriteFile(t, filepath.Join(wt, ".superpowers/sdd", tc.cite), "<html>")
		}
		path := filepath.Join(wt, ".superpowers/sdd/verify-report-visual-verify.md")
		guardtest.WriteFile(t, path, "## Report\n"+tc.report)
		args := []string{path, "0"}
		if tc.design {
			args = append(args, "--design", filepath.Join(root, "design.md"))
		}
		r := runGuard("check-verify-report", args, core.Env{Getenv: os.Getenv})
		if r.Rc != tc.rc || !strings.Contains(r.Out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.Rc, r.Out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportAccountsForEveryCapture pins that every PNG the
// resolver prints for the manifest's spec names its own view line
// (group-allowed-pairs-roles self-review), and every manifest frame its own
// verdict line.
func TestCheckVerifyReportAccountsForEveryCapture(t *testing.T) {
	t.Parallel()
	const allDone = "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n"
	for _, tc := range []struct {
		name, drop, shots string
		spec              bool
		rc                int
		want              string
	}{
		{"every capture accounted", "", "tests/visual", true, 0, "VERIFY-REPORT-COMPLETE"},
		{"a capture unnamed", "- a-open: ", "tests/visual", true, 1, "a-open-darwin.png is on no"},
		{"a frame verdict omitted", "- mock-frame-03: ", "tests/visual", true, 1, "mock-frame-03 has no `- mock-frame-03: match | differs` line"},
		{"a frame omitted but named on the frames line", "- mock-frame-03: ", "tests/visual", true, 0, "VERIFY-REPORT-COMPLETE"},
		{"no spec in the manifest", "- a-open: ", "tests/visual", false, 0, "VERIFY-REPORT-COMPLETE"},
		{"resolver finds nothing", "", "nowhere", true, 0, "VERIFY-REPORT-COMPLETE"},
	} {
		root := t.TempDir()
		guardtest.WriteFile(t, filepath.Join(root, ".flow/project.md"), rvsSection(tc.shots))
		snaps := filepath.Join(root, "tests/visual/a.spec.ts-snapshots")
		guardtest.WriteFile(t, filepath.Join(snaps, "a-open-darwin.png"), "")
		guardtest.WriteFile(t, filepath.Join(snaps, "a-roles-darwin.png"), "")
		lines := []string{
			allDone,
			"- a-open: " + filepath.Join(snaps, "a-open-darwin.png") + " — clean\n",
			"- a-roles: " + filepath.Join(snaps, "a-roles-darwin.png") + " — clean\n",
			"- mock-frame-03: match\n",
		}
		if strings.Contains(tc.name, "frames line") {
			lines = append(lines, "- frames: mock-frame-03: capture unchanged since abc1234\n")
		}
		var body strings.Builder
		for _, l := range lines {
			if tc.drop == "" || !strings.Contains(l, tc.drop) {
				body.WriteString(l)
			}
		}
		var specs []string
		if tc.spec {
			specs = []string{filepath.Join(root, "tests/visual/a.spec.ts")}
		}
		vrtManifest(t, filepath.Join(root, ".superpowers/sdd"), "visual-verify", "mock-frame-03", specs)
		path := filepath.Join(root, ".superpowers/sdd/verify-report-visual-verify.md")
		guardtest.WriteFile(t, path, "## Report\n"+body.String())
		r := runGuard("check-verify-report", []string{path, "0"}, core.Env{Getenv: os.Getenv})
		if r.Rc != tc.rc || !strings.Contains(r.Out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.Rc, r.Out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportAccountsForAnUnchangedView pins decision
// unchanged-view-line: in a scoped row — its key has a round file — a view
// whose capture a manifest pair holds may read `capture unchanged since
// <round base>`; the marker is refused in an unscoped row and on a view no
// frame is composed for (KAN-1122).
func TestCheckVerifyReportAccountsForAnUnchangedView(t *testing.T) {
	t.Parallel()
	const allDone = "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n"
	const marker = " — capture unchanged since abc1234"
	for _, tc := range []struct {
		name, open, roles string
		scoped            bool
		rc                int
		want              string
	}{
		{"an unchanged framed view in a scoped row is accounted", " — clean", marker, true, 0, "VERIFY-REPORT-COMPLETE"},
		{"an unscoped row reads every view", " — clean", marker, false, 1, `a-roles reads "capture unchanged"`},
		{"a view no frame is composed for is always read", marker, marker, true, 1, `a-open reads "capture unchanged"`},
	} {
		root := t.TempDir()
		guardtest.WriteFile(t, filepath.Join(root, ".flow/project.md"), rvsSection("tests/visual"))
		snaps := filepath.Join(root, "tests/visual/a.spec.ts-snapshots")
		guardtest.WriteFile(t, filepath.Join(snaps, "a-open-darwin.png"), "")
		guardtest.WriteFile(t, filepath.Join(snaps, "a-roles-darwin.png"), "")
		body := allDone +
			"- a-open: " + filepath.Join(snaps, "a-open-darwin.png") + tc.open + "\n" +
			"- a-roles: " + filepath.Join(snaps, "a-roles-darwin.png") + tc.roles + "\n" +
			"- frames: mock-frame-03: capture unchanged since abc1234\n"
		sdd := filepath.Join(root, ".superpowers/sdd")
		m := vvmManifestOut{Specs: []string{filepath.Join(root, "tests/visual/a.spec.ts")}, Captures: []string{}, Identical: []string{}, Steps: map[string]vvmStep{},
			Pairs: []vvmPair{{Frame: "mock-frame-03", Capture: filepath.Join(snaps, "a-roles-darwin.png")}}}
		mb, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		guardtest.WriteFile(t, filepath.Join(sdd, "visual-manifest-visual-verify.json"), string(mb))
		if tc.scoped {
			guardtest.WriteFile(t, roundFile(sdd, "visual-verify"), "abc1234 "+time.Now().Format(time.RFC3339Nano)+"\n")
		}
		path := filepath.Join(sdd, "verify-report-visual-verify.md")
		guardtest.WriteFile(t, path, "## Report\n"+body)
		r := runGuard("check-verify-report", []string{path, "0"}, core.Env{Getenv: os.Getenv})
		if r.Rc != tc.rc || !strings.Contains(r.Out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.Rc, r.Out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportTakesNoScopedSpecs pins decision verifier-runs-no-verify:
// the verifier no longer runs `verify`, so the guard takes no scoped-spec count
// and step 7 is no required step — an item still naming it is ignored.
func TestCheckVerifyReportTakesNoScopedSpecs(t *testing.T) {
	t.Parallel()
	const usage = "verify-report check, run in-process by visual-verify-pass close: <report file> <motions named> [--design <design.md>]"
	path := filepath.Join(t.TempDir(), "verify-report-visual-verify.md")
	r := runGuard("check-verify-report", []string{path, "0", "2"}, core.Env{Getenv: os.Getenv})
	if r.Rc != 2 || !strings.Contains(r.Out, usage) {
		t.Fatalf("scoped count: rc=%d out=%q; want exit 2 with %q", r.Rc, r.Out, usage)
	}
	for _, steps := range []string{
		"- steps: 4 done | 8 done | 9 done | 10 done | 11 done | motion n/a — no motions named\n",
		"- steps: 4 done | 7 not done | 8 done | 9 done | 10 done | 11 done | motion n/a — no motions named\n",
	} {
		if err := os.WriteFile(path, []byte("## Report\n"+steps), 0o644); err != nil {
			t.Fatal(err)
		}
		if r := runGuard("check-verify-report", []string{path, "0"}, core.Env{Getenv: os.Getenv}); r.Rc != 0 || !strings.Contains(r.Out, "VERIFY-REPORT-COMPLETE") {
			t.Fatalf("%q: rc=%d out=%q; want exit 0 COMPLETE", steps, r.Rc, r.Out)
		}
	}
}

// TestCheckVerifyReportReadsReportAlone pins decision verdict-only-handback:
// a `## Verdict` block's `- steps:` line never stands in for the `## Report`'s.
func TestCheckVerifyReportReadsReportAlone(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "verify-report-visual-verify.md")
	body := "## Verdict\n- verify: exit 0 — clean\n- steps: 4 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n\n" +
		"## Report\n- verify: exit 1\n- steps: 4 done | 8 done | 9 done | 11 done | motion n/a — none\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runGuard("check-verify-report", []string{path, "0"}, core.Env{Getenv: os.Getenv})
	if r.Rc != 1 || !strings.Contains(r.Out, "10 absent") {
		t.Fatalf("rc=%d out=%q; want exit 1 naming 10 absent", r.Rc, r.Out)
	}
}

const vrtDone = "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n"

// vrtFrame is a composed frame's one `match` line.
func vrtFrame(id string) string {
	return "- " + id + ": match\n"
}

// vrtManifest writes the manifest `visual-verify-pass measure` writes for key,
// built from its own types, over specs, with frame's pair carrying over.
func vrtManifest(t *testing.T, sdd, key, frame string, specs []string, over ...any) {
	t.Helper()
	m := vvmManifestOut{Specs: specs, Captures: []string{}, Identical: []string{}, Steps: map[string]vvmStep{},
		Pairs: []vvmPair{{Frame: frame, Bands: []any{}, Seams: []any{}, Over: over}}}
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	guardtest.WriteFile(t, filepath.Join(sdd, "visual-manifest-"+key+".json"), string(body))
}

// TestCheckVerifyReportMatchesAndDifferences pins the difference-only report:
// a composed frame is one `match` line, or a verdict and one `differs:` line
// per difference — with over rows on the manifest, no over-tolerance,
// spot-check, viewed or sweeps line is owed — and a frame with neither shape,
// or with no line at all, is still refused.
func TestCheckVerifyReportMatchesAndDifferences(t *testing.T) {
	t.Parallel()
	over := []any{map[string]any{"row": 1, "status": "missing", "keys": ""}, map[string]any{"row": 3, "status": "paired", "keys": "box.width"}}
	const differs = "- W1: differs\n- W1 differs: Try button — mockup x 40–740, capture x 48–728, inset 8 px a side\n- W1 differs: List label — mockup ink, capture grey #777\n"
	for _, tc := range []struct {
		name, key, body string
		manifest        bool
		rc              int
		want            string
	}{
		{"match line alone", "visual-verify", vrtFrame("W1"), true, 0, "COMPLETE"},
		{"difference lines alone", "visual-verify", differs, true, 0, "COMPLETE"},
		{"rest report reads its key's manifest", "visual-verify-rest-2", vrtFrame("W1"), true, 0, "COMPLETE"},
		{"a verdict with no difference line", "visual-verify", "- W1: differs\n", true, 1, "W1 reads neither match nor lists a `- W1 differs:"},
		{"a match beside a difference line", "visual-verify", vrtFrame("W1") + "- W1 differs: rail — mockup 8 px, capture 6 px\n", true, 1, "W1 reads match and has 1 `- W1 differs:` lines"},
		{"no manifest", "visual-verify", vrtFrame("W1"), false, 1, "no readable manifest"},
		{"a frame in no pair", "visual-verify", vrtFrame("W2"), true, 1, "W2 is in no pair of the manifest"},
	} {
		sdd := filepath.Join(t.TempDir(), ".superpowers/sdd")
		if tc.manifest {
			vrtManifest(t, sdd, "visual-verify", "W1", nil, over...)
		}
		path := filepath.Join(sdd, "verify-report-"+tc.key+".md")
		guardtest.WriteFile(t, path, "## Report\n"+vrtDone+tc.body)
		r := runGuard("check-verify-report", []string{path, "0"}, core.Env{Getenv: os.Getenv})
		if r.Rc != tc.rc || !strings.Contains(r.Out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.Rc, r.Out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportFrames: with --frames a remainder's report owes no
// capture the spec's sidecar maps only to frames outside its list, nor one no
// sidecar line maps.
func TestCheckVerifyReportFrames(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	guardtest.WriteFile(t, filepath.Join(root, ".flow/project.md"), rvsSection("tests/visual"))
	snaps := filepath.Join(root, "tests/visual/a.spec.ts-snapshots")
	for _, png := range []string{"a-open-darwin.png", "a-roles-darwin.png", "a-extra-darwin.png"} {
		guardtest.WriteFile(t, filepath.Join(snaps, png), "")
	}
	spec := filepath.Join(root, "tests/visual/a.spec.ts")
	guardtest.WriteFile(t, spec+".mockups", "frame none\na-open.png fo\na-roles.png fr\n")
	sdd := filepath.Join(root, ".superpowers/sdd")
	vrtManifest(t, sdd, "visual-verify", "fr", []string{spec})
	roles := vrtDone + "- a-roles: " + filepath.Join(snaps, "a-roles-darwin.png") + " — clean\n" +
		"- fr: match\n"
	open := vrtDone + "- a-open: " + filepath.Join(snaps, "a-open-darwin.png") + " — clean\n"
	for _, tc := range []struct {
		name, key, body string
		args            []string
		rc              int
		want            string
	}{
		{"a remainder owes its own frames' captures alone", "visual-verify-rest", roles, []string{"--frames", "fr"}, 0, "VERIFY-REPORT-COMPLETE"},
		{"without --frames every capture is owed", "visual-verify-rest", roles, nil, 1, "a-open-darwin.png is on no"},
		{"a remainder owes no unmapped capture", "visual-verify-rest", open, []string{"--frames", "fo"}, 0, "VERIFY-REPORT-COMPLETE"},
	} {
		path := filepath.Join(sdd, "verify-report-"+tc.key+".md")
		guardtest.WriteFile(t, path, "## Report\n"+tc.body)
		r := runGuard("check-verify-report", append([]string{path, "0"}, tc.args...), core.Env{Getenv: os.Getenv})
		if r.Rc != tc.rc || !strings.Contains(r.Out, tc.want) || tc.rc == 0 && strings.Contains(r.Out, "a-extra") || len(tc.args) > 0 && strings.Contains(r.Out, "a-roles") {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.Rc, r.Out, tc.rc, tc.want)
		}
	}
	if r := runGuard("check-verify-report", []string{filepath.Join(sdd, "verify-report-visual-verify-rest.md"), "0", "--frames"}, core.Env{Getenv: os.Getenv}); r.Rc != 2 {
		t.Errorf("--frames with no value: rc=%d out=%q; want exit 2", r.Rc, r.Out)
	}
}

// TestCheckVerifyReportBlanketDecision: one decision cited by two frames above
// the diff ceiling is refused, each large departure needing its own reason.
func TestCheckVerifyReportBlanketDecision(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, da, db, a, b string
		rc                 int
		want               string
	}{
		{"one decision over both frames", "0.1003", "0.1087", "d1", "d1", 1, "decision d1 excuses 2 frames"},
		{"a reason each", "0.1003", "0.1087", "d1", "d2", 0, "COMPLETE"},
		{"shared decision below the ceiling", "0.01", "0.02", "d1", "d1", 0, "COMPLETE"},
	} {
		sdd := filepath.Join(t.TempDir(), ".superpowers/sdd")
		pair := func(f, diff string) vvmPair {
			return vvmPair{Frame: f, Diff: diff, Bands: []any{}, Seams: []any{}, Over: []any{}}
		}
		body, err := json.Marshal(vvmManifestOut{Captures: []string{}, Identical: []string{}, Steps: map[string]vvmStep{}, Pairs: []vvmPair{pair("A", tc.da), pair("B", tc.db)}})
		if err != nil {
			t.Fatal(err)
		}
		guardtest.WriteFile(t, filepath.Join(sdd, "visual-manifest-visual-verify.json"), string(body))
		design := filepath.Join(sdd, "design.md")
		guardtest.WriteFile(t, design, "**ID:** d1\n\n**ID:** d2\n")
		path := filepath.Join(sdd, "verify-report-visual-verify.md")
		guardtest.WriteFile(t, path, "## Report\n"+vrtDone+"- A: match (decision "+tc.a+")\n- B: match (decision "+tc.b+")\n")
		r := runGuard("check-verify-report", []string{path, "0", "--design", design}, core.Env{Getenv: os.Getenv})
		if r.Rc != tc.rc || !strings.Contains(r.Out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.Rc, r.Out, tc.rc, tc.want)
		}
	}
}
