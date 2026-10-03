package guard

// The retired scripts/test-check-self-review-report.sh, ported case for
// case: each subtest is named after that harness's `ok:` label (there was
// never a case 17), followed by the paths the harness never reached and the
// collation pin. Fixtures are small docs/self-review/-shaped directories
// under t.TempDir(); the canonical angle table is the real
// skills/flow-self-review/SKILL.md unless a case overrides it, as the
// harness's was. Case 7 alone runs the guard bare over the repository's own
// docs/self-review/, read only, exactly as the harness did.
//
// Each assertion reads stdout and stderr together, the harness's `2>&1`,
// except where a case pins the split itself.

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

const csrNone = "_none — this angle produced no findings._"

// csrSec is one `##` angle section of a report.
func csrSec(title, label string, body ...string) string {
	return "## " + title + " — `" + label + "`\n\n" + strings.Join(body, "\n\n") + "\n"
}

var (
	csrFix   = csrSec("Problems and fixes", "flow-fix", "- **[flow-fix]** Preflight compares against a stale local base ref — declined")
	csrCost  = csrSec("Cost", "flow-cost", "- **[flow-cost]** Every panel slot gathers the same context independently — filed: KAN-201")
	csrImp   = csrSec("What went well", "flow-improvement", csrNone)
	csrAuto  = csrSec("Automation", "flow-automation", csrNone)
	csrStats = csrSec("Stats app", "flow-stats-app", csrNone)
	csrSpeed = csrSec("Speed", "flow-speed", csrNone)
	// csrFive is a report as written before angle 6 existed.
	csrFive      = strings.Join([]string{csrFix, csrCost, csrImp, csrAuto, csrStats}, "\n")
	csrCompliant = csrFive + "\n" + csrSpeed
	// csrNoneFive is the five original sections, each carrying the marker.
	csrNoneFive = strings.Join([]string{csrSec("Cost", "flow-cost", csrNone), csrImp, csrAuto, csrStats}, "\n")
)

const csrRenamedContract = `   | # | Angle | Label |
   |---|-------|-------|
   | 1 | Problems encountered, and what pipeline change would avoid them | ` + "`flow-regress`" + ` |
   | 2 | Token/time cost, and what would reduce it without quality loss | ` + "`flow-cost`" + ` |
   | 3 | What went well, and how to reproduce it | ` + "`flow-improvement`" + ` |
   | 4 | What could be automated or moved to a script | ` + "`flow-automation`" + ` |
   | 5 | What could move to the Go app or its persistent storage | ` + "`flow-stats-app`" + ` |
`

type csrResult struct {
	rc            int
	out, err, all string
}

var csrRepoRoot = sync.OnceValue(func() string {
	abs, err := filepath.Abs("../../..")
	if err != nil {
		panic(err)
	}
	return abs
})

// csrRun runs the guard in-process. vars override the process environment;
// FLOW_GUARD_REPO_ROOT is the real checkout unless vars set it.
func csrRun(t *testing.T, vars map[string]string, args ...string) csrResult {
	t.Helper()
	lookup := func(k string) (string, bool) {
		if v, ok := vars[k]; ok {
			return v, true
		}
		if k == "FLOW_GUARD_REPO_ROOT" {
			return csrRepoRoot(), true
		}
		return os.LookupEnv(k)
	}
	env := Env{Dir: t.TempDir(), LookupEnv: lookup, Getenv: func(k string) string { v, _ := lookup(k); return v }}
	var out, errb bytes.Buffer
	rc := checkSelfReviewReport(args, env, &out, &errb)
	return csrResult{rc, out.String(), errb.String(), out.String() + errb.String()}
}

// csrDir is a fixture directory holding files (name -> body).
func csrDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		writeFile(t, filepath.Join(dir, name), body)
	}
	return dir
}

func csrReport(t *testing.T, body string) string {
	return csrDir(t, map[string]string{"fixture-self-review.md": body})
}

func csrContract(t *testing.T, body string) map[string]string {
	p := filepath.Join(t.TempDir(), "contract.md")
	writeFile(t, p, body)
	return map[string]string{"CHECK_SELF_REVIEW_ANGLES_CONTRACT": p}
}

// csrHas is the harness's `case "$OUT" in *a*b*…*)`: every needle, in order.
func csrHas(s string, needles ...string) bool {
	for _, n := range needles {
		i := strings.Index(s, n)
		if i < 0 {
			return false
		}
		s = s[i+len(n):]
	}
	return true
}

func TestCheckSelfReviewReport(t *testing.T) {
	t.Parallel()

	type check struct {
		label string
		ok    func(r csrResult) bool
	}
	rcIs := func(n int) func(csrResult) bool { return func(r csrResult) bool { return r.rc == n } }
	has := func(n ...string) func(csrResult) bool { return func(r csrResult) bool { return csrHas(r.all, n...) } }

	cases := []struct {
		run    func(t *testing.T) csrResult
		checks []check
	}{
		{func(t *testing.T) csrResult { return csrRun(t, nil, csrReport(t, csrCompliant)) }, []check{
			{"case 1: compliant report exits 0", rcIs(0)},
			{"case 1: verdict line carries SELF-REVIEW-REPORT-OK", has("SELF-REVIEW-REPORT-OK")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{csrFix, csrCost, csrImp, csrAuto}, "\n")))
		}, []check{
			{"case 2: report missing the angle-5 section is caught", rcIs(1)},
			{"case 2: finding names the missing angle-5 label", has("missing section", "flow-stats-app")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{csrFix, csrCost, csrImp,
				csrSec("Automation", "flow-automation", "Nothing notable to say here, but no marker either."), csrStats}, "\n")))
		}, []check{
			{"case 3: a section with neither a finding line nor the none-marker is caught", rcIs(1)},
			{"case 3: finding names the empty automation section", has("flow-automation", "neither a finding line nor the none-marker")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{csrFix,
				csrSec("Cost", "flow-cost", "- **[flow-fix]** Wrong label under the cost section — declined"), csrImp, csrAuto, csrStats}, "\n")))
		}, []check{
			{"case 4: a mismatched finding-line label is caught", rcIs(1)},
			{"case 4: finding names the mismatch", has("label", "flow-fix", "flow-cost")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{csrFix,
				csrSec("Cost", "flow-cost", "- **[flow-cost]** Every panel slot gathers the same context independently — filed:"), csrImp, csrAuto, csrStats}, "\n")))
		}, []check{
			{"case 5: a filed finding with no issue key is caught", rcIs(1)},
			{"case 5: finding names the missing issue key", has("filed", "no issue key")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{csrFix,
				csrSec("Cost", "flow-cost", "- **[flow-cost]** Every panel slot gathers the same context independently — filed: yes"), csrImp, csrAuto, csrStats}, "\n")))
		}, []check{
			{"case 6: a filed finding with a malformed issue key is caught", rcIs(1)},
			{"case 6: finding names the malformed issue key", has("filed", "malformed issue key", "yes")},
		}},
		// Bare: the real repository's own docs/self-review/.
		{func(t *testing.T) csrResult { return csrRun(t, nil) }, []check{
			{"case 7: the real repository's own tree is clean", rcIs(0)},
			{"case 7: verdict line carries SELF-REVIEW-REPORT-OK", has("SELF-REVIEW-REPORT-OK")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrDir(t, map[string]string{"garbage-self-review.md": "Ordinary prose with nothing self-review shaped in it at all.\n"}))
		}, []check{
			{"case 8: a report checked for nothing is an undeclared coverage violation", rcIs(1)},
			{"case 8: the violation names the report, its zero count, and that it is undeclared", has("garbage-self-review.md", "0 checked", "not declared expected-zero")},
		}},
		{func(t *testing.T) csrResult { return csrRun(t, nil, t.TempDir()) }, []check{
			{"case 9: an empty corpus is a violation, not a vacuous pass", rcIs(1)},
			{"case 9: the verdict names the empty corpus explicitly", has("no member was ever recorded", "checked nothing", "empty corpus")},
		}},
		{func(t *testing.T) csrResult { return csrRun(t, nil, filepath.Join(t.TempDir(), "does-not-exist")) }, []check{
			{"case 10: an absent directory exits 2, distinct from 1", rcIs(2)},
			{"case 10: the failure names the absent directory", has("no such directory")},
		}},
		// The harness stubbed `find` on PATH; the port walks the directory
		// itself, so the walk is made to fail for real: a directory that can
		// be read but not searched, which is what made the bash's find fail.
		{func(t *testing.T) csrResult {
			dir := csrDir(t, map[string]string{"x.md": "irrelevant\n"})
			if err := os.Chmod(dir, 0o444); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
			r := csrRun(t, nil, dir)
			r.err = strings.ReplaceAll(r.err, dir, "<dir>")
			return r
		}, []check{
			{"case 11: a failing find exits 2, not folded into empty corpus", rcIs(2)},
			{"case 11: the failure names find's own error", func(r csrResult) bool {
				return r.out == "" && r.err == "find: <dir>: Permission denied\n"+
					"check-self-review-report: find failed while scanning '<dir>' (permission denied, or another find error)\n"
			}},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{
				csrSec("Problems and fixes", "flow-fix", csrNone),
				"### An h3 aside inside the fix section\n\n- **[flow-fix]** must not be folded into the section above — declined\n",
				csrNoneFive}, "\n")))
		}, []check{
			{"case 12: a ### heading resets the current section", rcIs(1)},
			{"case 12: the line after the ### is named an orphan, not folded into flow-fix", has("finding line appears before any recognized section heading")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{csrSec("Cost", "flow-cost", csrNone),
				csrSec("Problems and fixes", "flow-fix", csrNone), csrImp, csrAuto, csrStats}, "\n")))
		}, []check{
			{"case 13: an out-of-order section is caught", rcIs(1)},
			{"case 13: finding names the out-of-order section", has("flow-fix", "out of order", "flow-cost")},
		}},
		{func(t *testing.T) csrResult {
			fix := csrSec("Problems and fixes", "flow-fix", csrNone)
			return csrRun(t, nil, csrReport(t, strings.Join([]string{fix, fix, csrNoneFive}, "\n")))
		}, []check{
			{"case 14: a duplicate section is caught", rcIs(1)},
			{"case 14: finding names the duplicate section", has("duplicate section for angle", "flow-fix")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, "- **[flow-fix]** appears before any heading — declined\n\n"+
				strings.Join([]string{csrSec("Problems and fixes", "flow-fix", csrNone), csrNoneFive}, "\n")))
		}, []check{
			{"case 15: a pre-heading finding line is caught", rcIs(1)},
			{"case 15: finding names the pre-heading line", has("finding line appears before any recognized section heading")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{
				csrSec("Problems and fixes", "flow-fix", csrNone, "- **[flow-fix]** a finding alongside the none-marker above — declined"),
				csrNoneFive}, "\n")))
		}, []check{
			{"case 16: a none-marker alongside a finding line is caught", rcIs(1)},
			{"case 16: finding names the none-marker-plus-finding conflict", has("carries both the none-marker and", "finding line")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.ReplaceAll(csrCompliant, "\n", "\r\n")))
		}, []check{
			{"case 18: a CRLF-saved compliant report exits 0", rcIs(0)},
			{"case 18: verdict line carries SELF-REVIEW-REPORT-OK despite CRLF", has("SELF-REVIEW-REPORT-OK")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Replace(csrCompliant, "filed: KAN-201\n", "filed: KAN-201\t\n", 1)))
		}, []check{
			{"case 19: a trailing literal tab on a filed line is caught", rcIs(1)},
			{"case 19: the trailing tab is carried into the key and named malformed", has("filed", "malformed issue key")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{
				csrSec("Problems and fixes", "flow-fix", "\t- **[flow-fix]** indented finding must not be recognized — declined"),
				csrNoneFive}, "\n")))
		}, []check{
			{"case 20: a tab-indented finding-shaped line is not silently accepted", rcIs(1)},
			{"case 20: the fix section is flagged as neither a finding nor the none-marker", has("`flow-fix`", "neither a finding line nor the none-marker")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{
				csrSec("Problems and fixes", "flow-fix", csrNone), csrSec("Cost", "flow-cost", csrNone),
				csrSec("What went well", "flow-improvement", csrNone+"\x1f"), csrAuto, csrStats}, "\n")))
		}, []check{
			{"case 21: a raw 0x1F byte in report content is caught, not silently swallowed", rcIs(1)},
			{"case 21: the improvement section is flagged as neither a finding nor the none-marker", has("`flow-improvement`", "neither a finding line nor the none-marker")},
			{"case 21: the retired unsupported-control-byte violation is absent", func(r csrResult) bool { return !strings.Contains(r.all, "unsupported control byte") }},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Replace(csrCompliant, "filed: KAN-201\n", "maybe later\n", 1)))
		}, []check{
			{"case 22: an unrecognized finding-line disposition is caught", rcIs(1)},
			{"case 22: the violation names the unrecognized disposition rule", has("finding line disposition is neither `filed: <KEY>` nor `declined`")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrReport(t, strings.Join([]string{
				csrSec("Problems and fixes", "flow-fix", "-  **[flow-fix]** two spaces after the dash is malformed — declined"),
				csrNoneFive}, "\n")))
		}, []check{
			{"case 23: a malformed finding-shaped line is not silently dropped", rcIs(1)},
			{"case 23: the violation names the line as finding-shaped but malformed", has("finding-shaped line is malformed")},
		}},
		{func(t *testing.T) csrResult {
			dir := csrReport(t, csrCompliant)
			if err := os.Chmod(filepath.Join(dir, "fixture-self-review.md"), 0); err != nil {
				t.Fatal(err)
			}
			return csrRun(t, nil, dir)
		}, []check{
			{"case 24: an unreadable report file is 'cannot answer', not a violation", rcIs(2)},
			{"case 24: the die message names the unreadable file", has("cannot read report file", "fixture-self-review.md")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrDir(t, map[string]string{"fixture-self-review.md": csrCompliant,
				"fixture-context.md": "# Self-review context bundle for fixture\n"}))
		}, []check{
			{"case 25: a bundle beside a compliant report exits 0", rcIs(0)},
			{"case 25: the bundle's basename is absent from the output", func(r csrResult) bool { return !strings.Contains(r.all, "fixture-context.md") }},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, csrContract(t, csrRenamedContract), csrReport(t, csrCompliant))
		}, []check{
			{"case 26: a renamed canonical label moves the guard off the old spelling", rcIs(1)},
			{"case 26: the missing-section finding names the NEW label", has("missing section for angle", "flow-regress")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, csrContract(t, csrRenamedContract), csrReport(t, strings.ReplaceAll(csrCompliant, "flow-fix", "flow-regress")))
		}, []check{
			{"case 27: a report in the renamed spelling exits 0", rcIs(0)},
		}},
		{func(t *testing.T) csrResult {
			report := strings.Replace(strings.ReplaceAll(csrCompliant, "flow-fix", "flow-regress"),
				"## Cost — `flow-cost`\n", "## Interlude — `not-an-angle`\n\n## Cost — `flow-cost`\n", 1)
			return csrRun(t, csrContract(t, csrRenamedContract+"   | 6 | Docs that taught the operator something new | `flow-docs` |\n"), csrReport(t, report))
		}, []check{
			{"case 28: a sixth canonical angle is demanded", rcIs(1)},
			{"case 28: the missing-section finding names the added angle", has("missing section for angle", "flow-docs")},
		}},
		{func(t *testing.T) csrResult {
			dir := csrReport(t, csrCompliant)
			return csrRun(t, map[string]string{"CHECK_SELF_REVIEW_ANGLES_CONTRACT": filepath.Join(dir, "absent-contract.md")}, dir)
		}, []check{
			{"case 29: an unreadable canonical table is 'cannot answer'", rcIs(2)},
			{"case 29: the die message names the unreadable contract", has("canonical angle table is unreadable", "absent-contract.md")},
		}},
		{func(t *testing.T) csrResult {
			// The harness's `grep -v '^[[:space:]]*| [0-9]'`: every numbered row out.
			var kept []string
			for _, l := range strings.Split(csrRenamedContract, "\n") {
				if !regexp.MustCompile(`^[[:space:]]*\| [0-9]`).MatchString(l) {
					kept = append(kept, l)
				}
			}
			return csrRun(t, csrContract(t, strings.Join(kept, "\n")), csrReport(t, csrCompliant))
		}, []check{
			{"case 30: a label-free canonical table is 'cannot answer'", rcIs(2)},
			{"case 30: the die message names the label-free contract", has("yielded no labels", "contract.md")},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrDir(t, map[string]string{"fixture-self-review.md": csrFive,
				"five-angle-reports.txt": "other-self-review.md\nfixture-self-review.md\n"}))
		}, []check{
			{"case 31: a listed five-angle report exits 0", rcIs(0)},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, csrDir(t, map[string]string{"fixture-self-review.md": csrFive,
				"five-angle-reports.txt": "old-fixture-self-review.md\n"}))
		}, []check{
			{"case 32: an unlisted five-angle report is caught", rcIs(1)},
			{"case 32: the missing-section finding names flow-speed", has("missing section for angle", "flow-speed")},
		}},
		{func(t *testing.T) csrResult {
			dir := csrReport(t, csrCompliant)
			mkdir(t, filepath.Join(dir, "five-angle-reports.txt"))
			return csrRun(t, nil, dir)
		}, []check{
			{"case 33: an unreadable five-angle list is 'cannot answer'", rcIs(2)},
			{"case 33: the die message names the unreadable list", has("cannot read the five-angle report list")},
		}},

		// Beyond the harness: the exact output of a clean run and of a
		// violation, the refusals it never reached, and the collation.
		{func(t *testing.T) csrResult {
			dir := csrReport(t, csrCompliant)
			r := csrRun(t, nil, dir)
			r.out = strings.ReplaceAll(r.out, dir, "<dir>")
			return r
		}, []check{
			{"a clean run prints the verdict and the coverage fragment on stdout", func(r csrResult) bool {
				return r.out == "SELF-REVIEW-REPORT-OK: <dir> — 1 report(s) checked\n  <dir>/fixture-self-review.md 8\n" && r.err == ""
			}},
		}},
		{func(t *testing.T) csrResult {
			dir := csrReport(t, strings.Replace(csrCompliant, "filed: KAN-201\n", "filed: yes\n", 1))
			r := csrRun(t, nil, dir)
			r.out = strings.ReplaceAll(r.out, dir, "<dir>")
			return r
		}, []check{
			{"a violation is one stdout line and the count on stderr", func(r csrResult) bool {
				return r.rc == 1 && r.out == "<dir>/fixture-self-review.md:7: finding line marked filed with a malformed issue key `yes` (want an uppercase project key, a hyphen, digits — e.g. KAN-201)\n" &&
					r.err == "check-self-review-report: 1 violation(s) across 1 report(s) checked\n"
			}},
		}},
		{func(t *testing.T) csrResult { return csrRun(t, nil, "a", "b") }, []check{
			{"two arguments is the usage refusal", func(r csrResult) bool {
				return r.rc == 2 && r.out == "" && r.err == "check-self-review-report: usage: check-self-review-report.sh [dir]\n"
			}},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, nil, filepath.Join(csrReport(t, csrCompliant), "fixture-self-review.md"))
		}, []check{
			{"a file target is not a directory", func(r csrResult) bool { return r.rc == 2 && strings.Contains(r.err, ": not a directory: ") }},
		}},
		{func(t *testing.T) csrResult {
			dir := t.TempDir()
			if err := os.Chmod(dir, 0o311); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
			return csrRun(t, nil, dir)
		}, []check{
			{"an unreadable directory is cannot answer", func(r csrResult) bool { return r.rc == 2 && strings.Contains(r.err, "cannot read directory: ") }},
		}},
		{func(t *testing.T) csrResult {
			return csrRun(t, map[string]string{"FLOW_GUARD_REPO_ROOT": ""}, t.TempDir())
		}, []check{
			{"FLOW_GUARD_REPO_ROOT unset is cannot answer", func(r csrResult) bool {
				return r.rc == 2 && r.err == "check-self-review-report: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-self-review-report.sh, which sets it\n"
			}},
		}},
	}

	// One subtest per label, so parity is a `--- PASS` count; each fixture
	// runs once, in whichever of its labels gets there first. Fixture paths
	// never reach an assertion's needle, so the label naming the running
	// subtest's t.TempDir() changes nothing asserted.
	for _, c := range cases {
		var once sync.Once
		var r csrResult
		for _, ch := range c.checks {
			t.Run(ch.label, func(t *testing.T) {
				t.Parallel()
				once.Do(func() { r = c.run(t) })
				if !ch.ok(r) {
					t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", r.rc, r.out, r.err)
				}
			})
		}
	}

	// find's default -P never descends a symlinked starting point given
	// without a trailing slash: the corpus is empty, as at 9cd35da8.
	t.Run("symlinked target without a trailing slash is an empty corpus", func(t *testing.T) {
		t.Parallel()
		real := csrReport(t, csrCompliant)
		link := filepath.Join(t.TempDir(), "link")
		if err := os.Symlink(real, link); err != nil {
			t.Fatal(err)
		}
		r := csrRun(t, nil, link)
		if r.rc != 1 || !csrHas(r.all, "no member was ever recorded", "0 report(s) checked") {
			t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", r.rc, r.out, r.err)
		}
		if r := csrRun(t, nil, link+"/"); r.rc != 0 {
			t.Errorf("trailing slash: rc=%d\n%s", r.rc, r.all)
		}
	})

	// Under a UTF-8 locale bash's `.` matches no invalid byte, so a finding
	// line carrying one is malformed; under C it is accepted.
	for lc, want := range map[string]int{"en_US.UTF-8": 1, "C": 0} {
		t.Run("invalid UTF-8 finding line/"+lc, func(t *testing.T) {
			t.Parallel()
			body := strings.Replace(csrCompliant, "** ", "** \xff", 1)
			if body == csrCompliant {
				t.Fatal("fixture carries no finding line")
			}
			if r := csrRun(t, map[string]string{"LC_ALL": lc}, csrReport(t, body)); r.rc != want {
				t.Errorf("rc=%d want %d\n%s", r.rc, want, r.all)
			}
		})
	}

	// sort -z keeps a report name carrying a newline one path.
	t.Run("a report name carrying a newline stays one path", func(t *testing.T) {
		t.Parallel()
		dir := csrDir(t, map[string]string{"a\nb-self-review.md": csrCompliant, "c-self-review.md": csrCompliant})
		if r := csrRun(t, nil, dir); r.rc != 0 || !csrHas(r.all, "2 report(s) checked") {
			t.Errorf("rc=%d\n%s", r.rc, r.all)
		}
	})

	// The report order is `sort`'s under the caller's locale, as the bash's
	// `sort -z` was: en_US.UTF-8 folds case, C sorts by byte.
	for lc, want := range map[string]string{
		"en_US.UTF-8": "a-self-review.md:0: 0 checked, and not declared expected-zero (coverage)\nB-self-review.md:0: 0 checked, and not declared expected-zero (coverage)\n",
		"C":           "B-self-review.md:0: 0 checked, and not declared expected-zero (coverage)\na-self-review.md:0: 0 checked, and not declared expected-zero (coverage)\n",
	} {
		t.Run("collation/"+lc, func(t *testing.T) {
			t.Parallel()
			dir := csrDir(t, map[string]string{"B-self-review.md": "x\n", "a-self-review.md": "x\n"})
			r := csrRun(t, map[string]string{"LC_ALL": lc}, dir)
			if got := strings.ReplaceAll(r.out, dir+"/", ""); r.rc != 1 || got != want {
				t.Errorf("rc=%d stdout:\n%s", r.rc, got)
			}
		})
	}
}
