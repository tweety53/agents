package guard

import (
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
		{"no steps line", "## Report\n- verify: exit 0\n", "0", 1, "4 absent; 7 absent; 8 absent; 9 absent; 10 absent; 11 absent; motion absent"},
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
		{"missing report", "", "0", 2, "cannot read"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "verify-report-visual-verify.md")
			if tc.report != "" {
				if err := os.WriteFile(path, []byte(tc.report), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			r := runGuard("check-verify-report", []string{path, tc.motions}, Env{Getenv: os.Getenv})
			if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
				t.Fatalf("rc=%d out=%q; want rc=%d containing %q", r.rc, r.out, tc.rc, tc.want)
			}
		})
	}
	for _, args := range [][]string{nil, {"x"}, {"x", "-1"}, {"x", "two"}} {
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
