package guard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// vvcFixture is a prepare fixture whose stage has begun one row, its key
// `visual-verify`, with a design.md, a capture spec and its PNG, a full app
// suite and its zip, and a visual-verification/ directory to commit.
func vvcFixture(t *testing.T, extra string) *vvpFix {
	t.Helper()
	f := vvpFixture(t, extra)
	f.environ = append(f.environ, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	writeFile(t, filepath.Join(f.wt, "spectre/changes/demo/design.md"), "# design\n\nFrame f1 is the home screen.\n")
	if strings.Contains(extra, "`mockups`") {
		writeFile(t, filepath.Join(f.wt, "design/mockups/f1.png"), "png\n")
	}
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "commit", "-qm", "design")
	f.green()
	if rc, out, _ := f.prepare(); rc != 0 {
		t.Fatalf("prepare: exit %d\n%s", rc, out)
	}
	f.captures(f.wt)
	writeFile(t, filepath.Join(f.wt, "spectre/changes/demo/visual-verification/visual-verification.md"), "# vv\n")
	return f
}

// captures writes a capture spec, its PNG, the full app suite and its zip
// under root's tests/visual.
func (f *vvpFix) captures(root string) {
	for _, p := range []string{"home.spec.ts", "home.spec.ts-home.png", "full-app-suite.spec.ts", "full-app-suite.zip"} {
		writeFile(f.t, filepath.Join(root, "tests/visual", p), p+"\n")
	}
}

const vvcSteps = "4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — no motion named"

// report writes key's report: verdict carries the `## Verdict` lines after
// `- Model:`, root the capture checkout, body any further `## Report` lines.
func (f *vvpFix) report(key, root, verdict, body string) string {
	shots := filepath.Join(root, "tests/visual")
	r := "## Verdict\n- Model: opus\n" + verdict + "\n\n## Report\n- setup: not declared\n- fingerprint: exit 0\n- capture: exit 0\n" +
		"- full app suite: created — " + shots + "/full-app-suite.spec.ts, 1 screens, capture exit 0\n" +
		"- full app suite zip: " + shots + "/full-app-suite.zip, 1 PNGs\n" +
		"- spec: " + shots + "/home.spec.ts\n- home: " + shots + "/home.spec.ts-home.png\n" +
		"- per view, no frame composed: home sweeps: text done\n" + body
	path := filepath.Join(f.wt, ".superpowers/sdd/verify-report-"+key+".md")
	writeFile(f.t, path, r)
	return path
}

func vvcVerdict(capture, steps, cause string, departures ...string) string {
	v := "- exits: setup not declared, capture " + capture + ", full app suite 0, spec reach 0\n- steps: " + steps
	if cause != "" {
		v += "\n- cause: " + cause
	}
	v += "\n- departures: " + string(rune('0'+len(departures)))
	for _, d := range departures {
		v += "\n- " + d
	}
	return v
}

func (f *vvpFix) close(key string) (int, string, string) {
	f.t.Helper()
	return f.run("close", f.wt, "demo", key, "0", "--session-token", "mf-tok")
}

func vvcLog(t *testing.T, dir, format string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "log", "-1", "--format="+format).Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func TestVisualVerifyPassCloseDone(t *testing.T) {
	f := vvcFixture(t, "")
	f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
	rc, out, calls := f.close("visual-verify")
	head := vgHead(t, f.wt)
	for _, want := range []string{"ROUND-BASE: " + f.wt + "=" + head, "STAGE: completed", "NEXT: done"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q", want)
		}
	}
	for _, want := range []string{"record dispatch end -C " + f.wt + " -change demo -key visual-verify -session-token mf-tok -outcome completed",
		"stage end -C " + f.wt + " -command /flow -stage flow.visual-verify -outcome completed demo",
		"record render -change demo -kind ledger -repo " + f.wt} {
		if !strings.Contains(calls, want) {
			t.Errorf("calls lack %q", want)
		}
	}
	if rc != 0 || vvcLog(t, f.wt, "%s") != "test: record the visual-verify visual verification" || t.Failed() {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassCloseRest(t *testing.T) {
	f := vvcFixture(t, "")
	before := vgHead(t, f.wt)
	undone := strings.Replace(vvcSteps, "10 done", "10 not done", 1)
	f.report("visual-verify", f.wt, vvcVerdict("0", undone, ""), "- steps: "+undone+"\n")
	rc, out, calls := f.close("visual-verify")
	if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: rest visual-verify-rest") || !strings.Contains(calls, "-key visual-verify -session-token mf-tok -outcome aborted") ||
		strings.Contains(calls, "stage end") || vgHead(t, f.wt) != before || !strings.Contains(f.stageFile(), "verdict visual-verify VERIFY-REPORT-INCOMPLETE: ") {
		t.Fatalf("exit %d\n%s\n%s\n%s", rc, out, calls, f.stageFile())
	}
	f.appendRow("visual-verify-rest", f.wt)
	f.report("visual-verify-rest", f.wt, vvcVerdict("0", undone, ""), "- steps: "+undone+"\n")
	if rc, out, _ := f.close("visual-verify-rest"); rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: rest visual-verify-rest-2") {
		t.Fatalf("a second remainder: exit %d\n%s", rc, out)
	}

	// A frame-list gap other than `no sidecar line` is a remainder too.
	g := vvcFixture(t, "| `mockups` | `design/mockups` |\n")
	g.report("visual-verify", g.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
	if rc, out, calls := g.close("visual-verify"); rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: rest visual-verify-rest") || !strings.Contains(g.stageFile(), "verdict visual-verify FRAME-LIST-GAP: ") || !strings.Contains(calls, "-outcome aborted") {
		t.Fatalf("a frame-list gap: exit %d\n%s\n%s", rc, out, g.stageFile())
	}

	// Two worktrees: `-rest` goes before the repository suffix.
	h := vvcFixture(t, "")
	w, _ := h.sibling("gymie-frontend")
	writeFile(t, h.stagePath(), strings.Replace(h.stageFile(), "set 1\n", "set 2\n", 1))
	h.appendRow("visual-verify-demo", h.wt)
	h.report("visual-verify-demo", h.wt, vvcVerdict("0", undone, ""), "- steps: "+undone+"\n")
	if rc, out, _ := h.close("visual-verify-demo"); rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: rest visual-verify-rest-demo") {
		t.Fatalf("a suffixed key: exit %d\n%s", rc, out)
	}
	h.appendRow("visual-verify-gymie-frontend", w)
	writeFile(t, filepath.Join(w, ".superpowers/sdd/verify-report-visual-verify-gymie-frontend.md"), "## Verdict\n- steps: 10 not done\n")
	if rc, out, _ := h.close("visual-verify-gymie-frontend"); rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: rest visual-verify-rest-gymie-frontend") {
		t.Fatalf("a sibling's key: exit %d\n%s", rc, out)
	}
}

func (f *vvpFix) appendRow(key, w string) {
	writeFile(f.t, f.stagePath(), f.stageFile()+"row "+key+" "+w+"\n")
}

const vvcBlocked = "4 done | 7 done | 8 blocked — capture exit 1 | 9 blocked — capture exit 1 | 10 blocked — capture exit 1 | 11 blocked — capture exit 1 | motion n/a — no motion named"

func TestVisualVerifyPassCloseRedispatch(t *testing.T) {
	f := vvcFixture(t, "")
	before := vgHead(t, f.wt)
	f.report("visual-verify", f.wt, vvcVerdict("1", vvcBlocked, "test-failure"), "- steps: "+vvcBlocked+"\n")
	rc, out, calls := f.close("visual-verify")
	if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: redispatch visual-verify-2") || !strings.Contains(calls, "-outcome blocked -cause test-failure") ||
		vgHead(t, f.wt) != before || strings.Contains(calls, "stage end") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassCloseLaterBlocks(t *testing.T) {
	f := vvcFixture(t, "")
	f.report("visual-verify", f.wt, vvcVerdict("1", vvcBlocked, "environment"), "- steps: "+vvcBlocked+"\n")
	f.close("visual-verify")
	f.appendRow("visual-verify-2", f.wt)
	before := vgHead(t, f.wt)
	f.report("visual-verify-2", f.wt, vvcVerdict("1", vvcBlocked, "environment"), "- steps: "+vvcBlocked+"\n")
	rc, out, calls := f.close("visual-verify-2")
	if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: cure-environment visual-verify-3") || vgHead(t, f.wt) != before || strings.Contains(calls, "stage end") {
		t.Fatalf("environment: exit %d\n%s\n%s", rc, out, calls)
	}
	f.appendRow("visual-verify-3", f.wt)
	f.report("visual-verify-3", f.wt, vvcVerdict("1", vvcBlocked, "missing-fixture"), "- steps: "+vvcBlocked+"\n")
	rc, out, calls = f.close("visual-verify-3")
	if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: loop") || vgHead(t, f.wt) == before ||
		!strings.Contains(calls, "-outcome blocked -cause missing-fixture") || !strings.Contains(calls, "-stage flow.visual-verify -outcome stopped") || !strings.Contains(out, "STAGE: stopped") {
		t.Fatalf("missing-fixture: exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassCloseJudge(t *testing.T) {
	f := vvcFixture(t, "")
	f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, "", "home: the card is 3px narrower — departure"), "- steps: "+vvcSteps+"\n")
	before := vgHead(t, f.wt)
	rc, out, calls := f.close("visual-verify")
	if rc != 0 || !strings.Contains(out, "## Verdict") || !strings.Contains(out, "the card is 3px narrower") || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: judge") ||
		vgHead(t, f.wt) == before || strings.Contains(calls, "dispatch end") || strings.Contains(calls, "stage end") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}

	// A `no sidecar line` frame-list gap alone is judged, not a remainder.
	g := vvcFixture(t, "| `mockups` | `design/mockups` |\n")
	g.report("visual-verify", g.wt, vvcVerdict("0", vvcSteps, ""), "- frames: 1/1\n- f1 note: composed\n- steps: "+vvcSteps+"\n")
	if rc, out, calls := g.close("visual-verify"); rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: judge") || !strings.Contains(out, "no sidecar line: f1") || strings.Contains(calls, "dispatch end") {
		t.Fatalf("no sidecar line: exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassCloseNoReport(t *testing.T) {
	f := vvcFixture(t, "")
	rc, out, calls := f.close("visual-verify")
	if rc != 1 || !strings.Contains(out, "STOP: no report") || !strings.Contains(calls, "-outcome aborted") || !strings.Contains(calls, "-stage flow.visual-verify -outcome stopped") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
	if rc, _, calls := f.close("visual-verify-9"); rc != 2 || calls != "" {
		t.Fatalf("a key the stage file has no row for: exit %d, calls %q", rc, calls)
	}
}

func TestVisualVerifyPassCloseCommitsPathspec(t *testing.T) {
	f := vvcFixture(t, "")
	writeFile(t, filepath.Join(f.wt, "foreign.txt"), "x\n")
	gitRun(t, f.wt, "add", "foreign.txt")
	// A suite PNG the report counts but never names (kan-1060 gated review).
	writeFile(t, filepath.Join(f.wt, "tests/visual/full-app-suite.spec.ts-snapshots/home-darwin.png"), "png\n")
	f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
	if rc, out, _ := f.close("visual-verify"); rc != 0 {
		t.Fatalf("exit %d\n%s", rc, out)
	}
	out, _ := exec.Command("git", "-C", f.wt, "show", "--name-only", "--format=%B", "HEAD").Output()
	files := string(out)
	for _, want := range []string{"tests/visual/home.spec.ts\n", "tests/visual/home.spec.ts-home.png\n", "tests/visual/full-app-suite.spec.ts\n", "tests/visual/full-app-suite.zip\n", "tests/visual/full-app-suite.spec.ts-snapshots/home-darwin.png\n",
		"spectre/changes/demo/visual-verification/visual-verification.md\n", "Flow-Key: flow.visual-verify"} {
		if !strings.Contains(files, want) {
			t.Errorf("commit lacks %q:\n%s", want, files)
		}
	}
	if strings.Contains(files, "foreign.txt") {
		t.Errorf("the commit carries the pre-staged foreign file:\n%s", files)
	}
	if st, _ := exec.Command("git", "-C", f.wt, "diff", "--cached", "--name-only").Output(); string(st) != "foreign.txt\n" {
		t.Errorf("staged after close: %q", st)
	}

	// Nothing to commit: no commit, ROUND-BASE still printed.
	g := vvcFixture(t, "")
	for _, p := range []string{"tests/visual", "spectre/changes/demo/visual-verification"} {
		os.RemoveAll(filepath.Join(g.wt, p))
	}
	g.report("visual-verify", g.wt, vvcVerdict("0", "4 done | 7 done | 8 n/a — wire contract | 9 n/a — wire contract | 10 n/a — wire contract | 11 done | motion n/a — no motion named", ""),
		"- wire contract: GET /x — shape read\n- steps: 4 done | 7 done | 8 n/a — wire contract | 9 n/a — wire contract | 10 n/a — wire contract | 11 done | motion n/a — no motion named\n")
	before := vgHead(t, g.wt)
	if rc, out, _ := g.close("visual-verify"); rc != 0 || vgHead(t, g.wt) != before || !strings.Contains(out, "ROUND-BASE: "+g.wt+"="+before) {
		t.Fatalf("nothing to commit: exit %d\n%s", rc, out)
	}

	// A regression checkout: captures commit where they live — its main
	// checkout printing PUSH:, its worktree on spectre/demo not — and
	// visual-verification/ in the canonical worktree.
	for _, inWorktree := range []bool{false, true} {
		h := vvcFixture(t, "")
		regress := filepath.Join(filepath.Dir(h.wt), "regress")
		writeFile(t, filepath.Join(regress, "README"), "r\n")
		gitRun(t, regress, "init", "-q", "-b", "main")
		gitRun(t, regress, "add", "-A")
		gitRun(t, regress, "commit", "-qm", "r")
		capture := regress
		if inWorktree {
			capture = filepath.Join(filepath.Dir(h.wt), "regress-worktrees", "demo")
			gitRun(t, regress, "worktree", "add", "-q", "-b", "spectre/demo", capture)
		}
		h.captures(capture)
		h.report("visual-verify", capture, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
		writeFile(t, filepath.Join(h.wt, ".flow/project.md"), "# project\n\n"+vvpSection+"| `regression checkout` | `"+regress+"` |\n| `regression repo` | `x/regress` |\n")
		rc, out, _ := h.close("visual-verify")
		if rc != 0 || vvcLog(t, capture, "%s") != "test: record the visual-verify visual verification" ||
			vvcLog(t, h.wt, "%s") != "test: record the visual-verify visual verification" || strings.Contains(out, "PUSH:") == inWorktree ||
			(!inWorktree && !strings.Contains(out, "PUSH: git -C "+regress+" push")) {
			t.Fatalf("regression checkout (worktree %v): exit %d\n%s", inWorktree, rc, out)
		}
	}
}

func TestVisualVerifyPassCloseUsage(t *testing.T) {
	f := vvcFixture(t, "")
	for _, args := range [][]string{
		{f.wt, "demo", "visual-verify", "x", "--session-token", "mf-tok"},
		{f.wt, "demo", "visual-verify", "0"},
		{f.wt, "demo", "visual-verify", "0", "--session-token", "mf-tok", "--harness", "claude"},
	} {
		if rc, out, calls := f.run("close", args...); rc != 2 || !strings.Contains(out, "usage: visual-verify-pass.sh") || calls != "" {
			t.Errorf("%q: exit %d, calls %q\n%s", args, rc, calls, out)
		}
	}
}
