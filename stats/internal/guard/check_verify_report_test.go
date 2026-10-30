package guard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckVerifyReport(t *testing.T) {
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
			r := runGuard("check-verify-report", []string{path, tc.motions}, Env{Getenv: os.Getenv})
			if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
				t.Fatalf("rc=%d out=%q; want rc=%d containing %q", r.rc, r.out, tc.rc, tc.want)
			}
		})
	}
	for _, args := range [][]string{nil, {"x"}, {"x", "-1"}, {"x", "two"}, {"x", "0", "2"}} {
		if r := runGuard("check-verify-report", args, Env{Getenv: os.Getenv}); r.rc != 2 || !strings.Contains(r.out, "usage") {
			t.Fatalf("args %q: rc=%d out=%q; want usage exit 2", args, r.rc, r.out)
		}
	}
}

// TestCheckVerifyReportFingerprintReason pins kan-1010's self-review: a
// verifier wrote a bare `fingerprint: not declared` for a project declaring
// the row, its parent's step-6 result never named in the prompt. A bare one
// is an undone item; one carrying step 6's reason, or an exit, is not.
func TestCheckVerifyReportFingerprintReason(t *testing.T) {
	const steps = "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n"
	for _, tc := range []struct {
		line string
		rc   int
		want string
	}{
		{"- fingerprint: not declared\n", 1, "fingerprint: `not declared` with no reason"},
		{"- fingerprint: not declared — no fingerprint row in .flow/project.md\n", 0, "COMPLETE"},
		{"- fingerprint: exit 0\n", 0, "COMPLETE"},
	} {
		path := filepath.Join(t.TempDir(), "verify-report-visual-verify.md")
		if err := os.WriteFile(path, []byte("## Report\n"+tc.line+steps), 0o644); err != nil {
			t.Fatal(err)
		}
		r := runGuard("check-verify-report", []string{path, "0"}, Env{Getenv: os.Getenv})
		if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
			t.Fatalf("%q: rc=%d out=%q; want rc=%d containing %q", tc.line, r.rc, r.out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportPreExistingNeedsMergeBaseEvidence pins kan-1015's
// self-review: the kan-466 weight-tab button was attributed pre-existing from
// the code path alone and its baseline re-captured over the defect. A
// pre-existing attribution must name its merge-base evidence.
func TestCheckVerifyReportPreExistingNeedsMergeBaseEvidence(t *testing.T) {
	const steps = "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n"
	for _, tc := range []struct {
		line string
		rc   int
		want string
	}{
		{"  - kan-466 weight-tab: button fill differs. Attributed as pre-existing from the code path only\n", 1, "10 pre-existing attribution with no merge-base evidence"},
		{"  - kan-466 weight-tab: pre-existing — merge-base baseline: 4f8f8344:kan-466-weigh-in-history.spec.ts-snapshots/kan466-weight-tab-darwin.png\n", 0, "COMPLETE"},
		{"  - kan-466 weight-tab: present at merge base — merge-base capture: /tmp/x.png\n", 0, "COMPLETE"},
		{"  - templates T2: not pre-existing — this change's fill token\n", 0, "COMPLETE"},
	} {
		path := filepath.Join(t.TempDir(), "verify-report-visual-verify.md")
		if err := os.WriteFile(path, []byte("## Report\n"+steps+tc.line), 0o644); err != nil {
			t.Fatal(err)
		}
		r := runGuard("check-verify-report", []string{path, "0"}, Env{Getenv: os.Getenv})
		if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
			t.Fatalf("%q: rc=%d out=%q; want rc=%d containing %q", tc.line, r.rc, r.out, tc.rc, tc.want)
		}
	}
}

// kan-1017 (sr12): every verifier wrote `no mockup source` while design.md
// cited a handoff page that existed, and left 7 of 11 matrix columns n/a.
func TestCheckVerifyReportNoMockupSourceWithCitedSource(t *testing.T) {
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
		writeFile(t, filepath.Join(root, "design.md"), "Mockups: `handoff/Flow.html`.\n")
		if tc.cited {
			writeFile(t, filepath.Join(wt, "handoff/Flow.html"), "<html>")
		}
		if tc.inVV {
			writeFile(t, filepath.Join(root, "visual-verification.md"), "- AD1: no mockup source\n")
		}
		body := "## Report\n" + steps
		if tc.inReport {
			body += "- matrix: no mockup source\n"
		}
		path := filepath.Join(wt, ".superpowers/sdd/verify-report-visual-verify.md")
		writeFile(t, path, body)
		args := []string{path, "0"}
		if tc.design {
			args = append(args, "--design", filepath.Join(root, "design.md"))
		}
		r := runGuard("check-verify-report", args, Env{Getenv: os.Getenv})
		if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.rc, r.out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportMatrixLinePerFrame pins kan-1016's self-review: one
// `matrix:` line covered nine frames, and the warm-up sheet's frame had no
// matrix until round 8. A `done` step 10 needs a matrix line per composed frame.
func TestCheckVerifyReportMatrixLinePerFrame(t *testing.T) {
	const frames = "- templates-W1: /tmp/w1.composite.png diff=0.1\n- templates-A4b: /tmp/a4b.composite.png diff=0.2\n- templates-W1 matrix: 12 rows\n" +
		"- templates-W1 bands: 3/3 paired\n- templates-W1 seams: 2/2 paired\n- templates-W1 over-tolerance: 0\n" +
		"- templates-A4b bands: 3/3 paired\n- templates-A4b seams: 2/2 paired\n- templates-A4b over-tolerance: 0\n"
	for _, tc := range []struct {
		name, body string
		rc         int
		want       string
	}{
		{"missing", "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n" + frames, 1, "templates-A4b has no matrix: line"},
		{"present", "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n" + frames + "- templates-A4b matrix: 9 rows\n", 0, "COMPLETE"},
		{"not done", "- steps: 4 done | 7 done | 8 done | 9 done | 10 n/a — no mockups declared | 11 done | motion n/a — none\n" + frames, 0, "COMPLETE"},
	} {
		dir := t.TempDir()
		vrtManifest(t, dir, "visual-verify", "templates-W1")
		vrtManifest(t, dir, "visual-verify-a4b", "templates-A4b") // a second pair, merged below
		vrtMerge(t, dir)
		path := filepath.Join(dir, "verify-report-visual-verify.md")
		if err := os.WriteFile(path, []byte("## Report\n"+tc.body), 0o644); err != nil {
			t.Fatal(err)
		}
		r := runGuard("check-verify-report", []string{path, "0"}, Env{Getenv: os.Getenv})
		if r.rc != tc.rc || !strings.Contains(r.out, tc.want) || (tc.rc == 1 && strings.Contains(r.out, "templates-W1 has")) {
			t.Fatalf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.rc, r.out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportStep10WithACitedPrototype pins that step 10 reported
// n/a beside an approved prototype design.md cites is undone: the prototype is
// rendered as each capture's frame (kan-1038).
func TestCheckVerifyReportStep10WithACitedPrototype(t *testing.T) {
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
		writeFile(t, filepath.Join(root, "design.md"), "## Context\n\nPrototype: `.superpowers/sdd/"+tc.cite+"`.\n")
		if tc.exists {
			writeFile(t, filepath.Join(wt, ".superpowers/sdd", tc.cite), "<html>")
		}
		path := filepath.Join(wt, ".superpowers/sdd/verify-report-visual-verify.md")
		writeFile(t, path, "## Report\n"+tc.report)
		args := []string{path, "0"}
		if tc.design {
			args = append(args, "--design", filepath.Join(root, "design.md"))
		}
		r := runGuard("check-verify-report", args, Env{Getenv: os.Getenv})
		if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.rc, r.out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportAccountsForEveryCapture pins that every PNG the
// resolver prints for the report's `- spec:` names its own view line, and
// every view no frame is composed for its per-view sweeps line
// (group-allowed-pairs-roles self-review).
func TestCheckVerifyReportAccountsForEveryCapture(t *testing.T) {
	const allDone = "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n"
	for _, tc := range []struct {
		name, drop, shots string
		rc                int
		want              string
	}{
		{"every capture accounted", "", "tests/visual", 0, "VERIFY-REPORT-COMPLETE"},
		{"a capture unnamed", "- a-open: ", "tests/visual", 1, "a-open-darwin.png is on no"},
		{"no per-view sweeps line", "per view, no frame composed", "tests/visual", 1, "a-open has no frame and no"},
		{"frame lines removed", "/vv/a-roles.png", "tests/visual", 1, "a-roles has no frame and no"},
		{"no spec line", "- spec: ", "tests/visual", 0, "VERIFY-REPORT-COMPLETE"},
		{"resolver finds nothing", "", "nowhere", 0, "VERIFY-REPORT-COMPLETE"},
	} {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, ".flow/project.md"), rvsSection(tc.shots))
		snaps := filepath.Join(root, "tests/visual/a.spec.ts-snapshots")
		writeFile(t, filepath.Join(snaps, "a-open-darwin.png"), "")
		writeFile(t, filepath.Join(snaps, "a-roles-darwin.png"), "")
		lines := []string{
			allDone,
			"- spec: <root>/tests/visual/a.spec.ts\n",
			"- a-open: " + filepath.Join(snaps, "a-open-darwin.png") + " — header and table read\n",
			"- a-roles: " + filepath.Join(snaps, "a-roles-darwin.png") + " — roles panel read\n",
			"- mock-frame-03: <root>/vv/a-roles.png diff=n/a — prototype frame\n",
			"- mock-frame-03 matrix: 4 rows\n- mock-frame-03 bands: 2/2 paired\n- mock-frame-03 seams: 1/1 paired\n- mock-frame-03 over-tolerance: 0\n",
			"- per view, no frame composed: a-open sweeps: text done | overflow done\n",
		}
		var body strings.Builder
		for _, l := range lines {
			drop := tc.drop != "" && strings.Contains(l, tc.drop)
			if tc.name == "frame lines removed" && strings.Contains(l, "mock-frame-03 matrix:") {
				drop = true
			}
			if !drop {
				body.WriteString(strings.ReplaceAll(l, "<root>", root))
			}
		}
		vrtManifest(t, filepath.Join(root, ".superpowers/sdd"), "visual-verify", "mock-frame-03")
		path := filepath.Join(root, ".superpowers/sdd/verify-report-visual-verify.md")
		writeFile(t, path, "## Report\n"+body.String())
		r := runGuard("check-verify-report", []string{path, "0"}, Env{Getenv: os.Getenv})
		if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.rc, r.out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportTakesNoScopedSpecs pins decision verifier-runs-no-verify:
// the verifier no longer runs `verify`, so the guard takes no scoped-spec count
// and step 7 is no required step — an item still naming it is ignored.
func TestCheckVerifyReportTakesNoScopedSpecs(t *testing.T) {
	const usage = "check-verify-report.sh <report file> <motions named> [--design <design.md>]"
	path := filepath.Join(t.TempDir(), "verify-report-visual-verify.md")
	r := runGuard("check-verify-report", []string{path, "0", "2"}, Env{Getenv: os.Getenv})
	if r.rc != 2 || !strings.Contains(r.out, usage) {
		t.Fatalf("scoped count: rc=%d out=%q; want exit 2 with %q", r.rc, r.out, usage)
	}
	for _, steps := range []string{
		"- steps: 4 done | 8 done | 9 done | 10 done | 11 done | motion n/a — no motions named\n",
		"- steps: 4 done | 7 not done | 8 done | 9 done | 10 done | 11 done | motion n/a — no motions named\n",
	} {
		if err := os.WriteFile(path, []byte("## Report\n"+steps), 0o644); err != nil {
			t.Fatal(err)
		}
		if r := runGuard("check-verify-report", []string{path, "0"}, Env{Getenv: os.Getenv}); r.rc != 0 || !strings.Contains(r.out, "VERIFY-REPORT-COMPLETE") {
			t.Fatalf("%q: rc=%d out=%q; want exit 0 COMPLETE", steps, r.rc, r.out)
		}
	}
}

// TestCheckVerifyReportReadsReportAlone pins decision verdict-only-handback:
// a `## Verdict` block's `- steps:` line never stands in for the `## Report`'s.
func TestCheckVerifyReportReadsReportAlone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "verify-report-visual-verify.md")
	body := "## Verdict\n- verify: exit 0 — clean\n- steps: 4 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n\n" +
		"## Report\n- verify: exit 1\n- steps: 4 done | 8 done | 9 done | 11 done | motion n/a — none\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runGuard("check-verify-report", []string{path, "0"}, Env{Getenv: os.Getenv})
	if r.rc != 1 || !strings.Contains(r.out, "10 absent") {
		t.Fatalf("rc=%d out=%q; want exit 1 naming 10 absent", r.rc, r.out)
	}
}

const vrtDone = "- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — none\n"

// vrtFrame is one composed frame's full set of lines but its over-tolerance one.
func vrtFrame(id string) string {
	return "- " + id + ": /c/" + id + ".png diff=0.1 — match\n- " + id + " bands: 3/3 paired\n- " + id +
		" seams: 2/2 paired\n- " + id + " sweeps: text done | order n/a — one column\n- " + id + " matrix: 3 elements\n"
}

// vrtManifest writes the manifest `visual-verify-pass measure` writes for key,
// built from its own types, with frame's pair carrying over.
func vrtManifest(t *testing.T, sdd, key, frame string, over ...any) {
	t.Helper()
	m := vvmManifestOut{Captures: []string{}, Identical: []string{}, Steps: map[string]vvmStep{},
		Pairs: []vvmPair{{Frame: frame, Bands: []any{}, Seams: []any{}, Over: over}}}
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sdd, "visual-manifest-"+key+".json"), string(body))
}

func vrtOver(row int, status, keys string) any {
	return map[string]any{"row": row, "status": status, "keys": keys}
}

// TestCheckVerifyReportOverToleranceLine pins decision over-tolerance-line:
// every row the manifest's `over` list names for a composed frame is named on
// its `over-tolerance:` line with the manifest's count.
func TestCheckVerifyReportOverToleranceLine(t *testing.T) {
	for _, tc := range []struct {
		name, key, line string
		manifest        bool
		rc              int
		want            string
	}{
		{"every row named", "visual-verify", "- W1 over-tolerance: 2 — #1: departure; #3: annotation\n", true, 0, "COMPLETE"},
		{"missing line", "visual-verify", "", true, 1, "W1 has no over-tolerance: line"},
		{"wrong count", "visual-verify", "- W1 over-tolerance: 1 — #1: departure; #3: annotation\n", true, 1, "W1 over-tolerance: 1, the manifest's over count is 2"},
		{"unnamed row", "visual-verify", "- W1 over-tolerance: 2 — #1: departure\n", true, 1, "row #3 unnamed"},
		{"rest report reads its key's manifest", "visual-verify-rest-2", "- W1 over-tolerance: 2 — #1: departure; #3: annotation\n", true, 0, "COMPLETE"},
		{"no manifest", "visual-verify", "- W1 over-tolerance: 0\n", false, 1, "no readable manifest"},
	} {
		sdd := filepath.Join(t.TempDir(), ".superpowers/sdd")
		if tc.manifest {
			vrtManifest(t, sdd, "visual-verify", "W1", vrtOver(1, "missing", ""), vrtOver(3, "paired", "box.width"))
		}
		path := filepath.Join(sdd, "verify-report-"+tc.key+".md")
		writeFile(t, path, "## Report\n"+vrtDone+vrtFrame("W1")+tc.line)
		r := runGuard("check-verify-report", []string{path, "0"}, Env{Getenv: os.Getenv})
		if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.rc, r.out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportOverToleranceCauses pins step 10's closed cause list.
func TestCheckVerifyReportOverToleranceCauses(t *testing.T) {
	for _, tc := range []struct {
		cause, status, keys string
		ok                  bool
	}{
		{"departure — a 4px shift", "paired", "box.width", true},
		{"annotation", "extra", "", true},
		{"transcription", "paired", "content.width", true},
		{"text width", "paired", "content.width", true},
		{"decision thin-parent-prepare-close", "paired", "box.width", true},
		{"present at the merge base — merge-base capture: /c/base.png", "paired", "box.width", true},
		{"mockup source — 8px from the CSS", "paired", "unresolved", true},
		{"looks fine", "paired", "box.width", false},
		{"decision no-such-id", "paired", "box.width", false},
		{"present at the merge base", "paired", "box.width", false},
		{"mockup source", "paired", "box.width", false},
	} {
		wt := t.TempDir()
		sdd := filepath.Join(wt, ".superpowers/sdd")
		design := filepath.Join(wt, "spectre/changes/x/design.md")
		writeFile(t, design, "## Decisions\n\n**ID:** thin-parent-prepare-close\n")
		vrtManifest(t, sdd, "visual-verify", "W1", vrtOver(2, tc.status, tc.keys))
		path := filepath.Join(sdd, "verify-report-visual-verify.md")
		writeFile(t, path, "## Report\n"+vrtDone+vrtFrame("W1")+"- W1 over-tolerance: 1 — #2: "+tc.cause+"\n")
		r := runGuard("check-verify-report", []string{path, "0", "--design", design}, Env{Getenv: os.Getenv})
		// An unevidenced merge-base attribution is refused by the line-level
		// evidence check before the cause list is read.
		if got := r.rc == 0; got != tc.ok || (!tc.ok && !strings.Contains(r.out, "off step 10's closed list") && !strings.Contains(r.out, "no merge-base evidence")) {
			t.Errorf("%q: rc=%d out=%q; want admitted=%v", tc.cause, r.rc, r.out, tc.ok)
		}
	}
}

// TestCheckVerifyReportComposedFrameLines pins decision by-hand-check-automated:
// a composed frame needs its bands:, seams: and matrix: lines.
func TestCheckVerifyReportComposedFrameLines(t *testing.T) {
	for _, drop := range []string{"bands:", "seams:", "matrix:"} {
		sdd := filepath.Join(t.TempDir(), ".superpowers/sdd")
		vrtManifest(t, sdd, "visual-verify", "W1")
		var body strings.Builder
		for _, l := range strings.SplitAfter(vrtFrame("W1"), "\n") {
			if !strings.Contains(l, drop) {
				body.WriteString(l)
			}
		}
		path := filepath.Join(sdd, "verify-report-visual-verify.md")
		writeFile(t, path, "## Report\n"+vrtDone+body.String()+"- W1 over-tolerance: 0\n")
		r := runGuard("check-verify-report", []string{path, "0"}, Env{Getenv: os.Getenv})
		if r.rc != 1 || !strings.Contains(r.out, "W1 has no "+drop) {
			t.Errorf("drop %s: rc=%d out=%q", drop, r.rc, r.out)
		}
	}
}

// TestCheckVerifyReportMatrixCells pins decision matrix-kind-opens-with-class:
// a matrix cell is n/a only in a column that does not apply to its row's class.
func TestCheckVerifyReportMatrixCells(t *testing.T) {
	const head = "| element | kind | content | alignment | border | fill | radius | shadow |\n|---|---|---|---|---|---|---|---|\n"
	for _, tc := range []struct {
		name, row string
		rc        int
		want      string
	}{
		{"text run n/a in its dashes", "| title | text run heading | 1 | 2 | n/a — no box | n/a — none | — | — |", 0, "COMPLETE"},
		{"text run n/a in content", "| title | text run heading | n/a — skipped | 2 | n/a | n/a | n/a | n/a |", 1, "title.content reads"},
		{"empty cell", "| btn | control save | 1 | 2 | 3 |  | 5 | 6 |", 1, "btn.fill is empty"},
		{"rule n/a in its dashes", "| hr | rule divider | — | — | — | 1 | — | — |", 0, "COMPLETE"},
		{"no mockup source", "| btn | control save | 1 | 2 | n/a — no mockup source | 4 | 5 | 6 |", 0, "COMPLETE"},
		{"unknown class", "| x | blob | 1 | 2 | n/a — none | 4 | 5 | 6 |", 1, "x.border reads"},
		// kan-938: a control's border left n/a when the measure failed.
		{"kan-938 control border", "| btn | control save | 1 | 2 | n/a — measure failed | 4 | 5 | 6 |", 1, "visual-verification.md:5: btn.border reads"},
	} {
		wt := t.TempDir()
		design := filepath.Join(wt, "spectre/changes/x/design.md")
		writeFile(t, design, "## Context\n")
		writeFile(t, filepath.Join(wt, "spectre/changes/x/visual-verification.md"), "# vv\n\n"+head+tc.row+"\n")
		path := filepath.Join(wt, ".superpowers/sdd/verify-report-visual-verify.md")
		writeFile(t, path, "## Report\n"+vrtDone)
		r := runGuard("check-verify-report", []string{path, "0", "--design", design}, Env{Getenv: os.Getenv})
		if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
			t.Errorf("%s: rc=%d out=%q; want rc=%d containing %q", tc.name, r.rc, r.out, tc.rc, tc.want)
		}
	}
}

// TestCheckVerifyReportSweepItems pins that a sweep item reads `done` or `n/a — <why>`.
func TestCheckVerifyReportSweepItems(t *testing.T) {
	for _, tc := range []struct {
		line string
		rc   int
	}{
		{"- per view, no frame composed: home sweeps: text done | order n/a — one column | rows done — 2 rows\n", 0},
		{"- per view, no frame composed: home sweeps: text done | order skipped\n", 1},
		{"- per view, no frame composed: home sweeps: text done | order n/a\n", 1},
	} {
		path := filepath.Join(t.TempDir(), "verify-report-visual-verify.md")
		writeFile(t, path, "## Report\n"+vrtDone+tc.line)
		r := runGuard("check-verify-report", []string{path, "0"}, Env{Getenv: os.Getenv})
		if r.rc != tc.rc || (tc.rc == 1 && !strings.Contains(r.out, "reads neither")) {
			t.Errorf("%q: rc=%d out=%q; want rc=%d", tc.line, r.rc, r.out, tc.rc)
		}
	}
}

// vrtMerge folds visual-manifest-visual-verify-a4b.json's pairs into
// visual-manifest-visual-verify.json: one manifest, two composed frames.
func vrtMerge(t *testing.T, dir string) {
	t.Helper()
	var a, b vvmManifestOut
	for p, m := range map[string]*vvmManifestOut{"visual-verify": &a, "visual-verify-a4b": &b} {
		body, err := os.ReadFile(filepath.Join(dir, "visual-manifest-"+p+".json"))
		if err != nil || json.Unmarshal(body, m) != nil {
			t.Fatal(err)
		}
	}
	a.Pairs = append(a.Pairs, b.Pairs...)
	body, _ := json.Marshal(a)
	writeFile(t, filepath.Join(dir, "visual-manifest-visual-verify.json"), string(body))
}
