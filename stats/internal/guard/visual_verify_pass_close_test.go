package guard

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// vvcFixture is a prepare fixture whose stage has begun one row, its key
// `visual-verify`, with a design.md, a capture spec and its PNG, a full app
// suite and its zip, and a visual-verification/ directory left uncommitted
// under a .gitignore holding spectre/changes/, as the real repository has —
// its one file, uncommitted.txt, a name no component reads, so no fixture
// stands for the record at a depth no reader uses.
func vvcFixture(t *testing.T, extra string) *vvpFix {
	t.Helper()
	f := vvpFixture(t, extra)
	f.environ = append(f.environ, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	writeFile(t, filepath.Join(f.wt, "spectre/changes/demo/design.md"), "# design\n\nFrame f1 is the home screen.\n")
	if strings.Contains(extra, "`mockups`") {
		writeFile(t, filepath.Join(f.wt, "design/mockups/f1.png"), "png\n")
	}
	writeFile(t, filepath.Join(f.wt, ".gitignore"), "spectre/changes/\n")
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "add", "-f", "spectre/changes/demo/design.md")
	gitRun(t, f.wt, "commit", "-qm", "design")
	f.green()
	if rc, out, _ := f.prepare(); rc != 0 {
		t.Fatalf("prepare: exit %d\n%s", rc, out)
	}
	f.captures(f.wt)
	writeFile(t, filepath.Join(f.wt, "spectre/changes/demo/visual-verification/uncommitted.txt"), "# vv\n")
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
		"- home: " + shots + "/home.spec.ts-home.png\n" +
		"- per view, no frame composed: home sweeps: text done\n" + body
	path := filepath.Join(f.wt, ".superpowers/sdd/verify-report-"+key+".md")
	writeFile(f.t, path, r)
	f.manifest(path, shots)
	return path
}

// manifest writes the manifest of report's key over shots' capture spec,
// with pairs.
func (f *vvpFix) manifest(report, shots string, pairs ...vvmPair) {
	m, err := json.Marshal(vvmManifestOut{Specs: []string{shots + "/home.spec.ts"}, Captures: []string{}, Identical: []string{}, Steps: map[string]vvmStep{}, Pairs: append([]vvmPair{}, pairs...)})
	if err != nil {
		f.t.Fatal(err)
	}
	writeFile(f.t, vrManifestPath(report), string(m))
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

// smoke: core — a visual-verify pass closing done
func TestVisualVerifyPassCloseDone(t *testing.T) {
	t.Parallel()
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
	smoke.Heavy(t)
	t.Parallel()
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
	t.Parallel()
	f := vvcFixture(t, "")
	before := vgHead(t, f.wt)
	f.report("visual-verify", f.wt, vvcVerdict("1", vvcBlocked, "test-failure"), "- steps: "+vvcBlocked+"\n")
	rc, out, calls := f.close("visual-verify")
	if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: redispatch visual-verify-2") || !strings.Contains(calls, "-outcome blocked -cause test-failure") ||
		vgHead(t, f.wt) != before || strings.Contains(calls, "stage end") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

// Any non-zero exit blocks, and a blocked report naming no cause is an
// environment block.
func TestVisualVerifyPassCloseBlockedDefaults(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "")
	f.report("visual-verify", f.wt, vvcVerdict("2", vvcBlocked, ""), "- steps: "+vvcBlocked+"\n")
	rc, out, calls := f.close("visual-verify")
	if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: redispatch visual-verify-2") || !strings.Contains(calls, "-outcome blocked -cause environment\n") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

// A die marks no stage: the guard makes no stage call.
func TestVisualVerifyPassCloseDieMarksNoStage(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "")
	writeExec(t, filepath.Join(f.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+f.log+"\ncase \"$1 $2 $3\" in \"record dispatch end\") echo store down; exit 1;; esac\n")
	f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
	rc, out, calls := f.close("visual-verify")
	if rc != 2 || strings.Contains(calls, "stage ") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassCloseLaterBlocks(t *testing.T) {
	t.Parallel()
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
		!strings.Contains(calls, "-outcome blocked -cause missing-fixture") || !strings.Contains(out, "STAGE: stopped") {
		t.Fatalf("missing-fixture: exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassCloseJudge(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
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

// vvcDepartures is a verdict naming n departures, every exit zero.
func vvcDepartures(n int) string {
	return "- exits: setup not declared, capture 0, full app suite 0, spec reach 0\n- steps: " + vvcSteps +
		fmt.Sprintf("\n- departures: %d\n- home: the card is 3px narrower — this change", n)
}

// TestVisualVerifyPassCloseEscalates pins the ESCALATE: line (kan-1091: fix
// rounds 4 to 8 ran on the base pair while the departures went 9, 19, 26,
// 24): a round whose report names more departures than the newest visual
// opener of its round, or any failed round from the third on, makes the next
// round a No-progress escalation.
func TestVisualVerifyPassCloseEscalates(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f := vvcFixture(t, "")
	writeFile(t, filepath.Join(f.wt, "spectre/changes/demo/tasks.md"),
		"- [x] 1. In-run fix 1 — flow.visual-verify: two departures\n**Opener:** visual-verify\n\n"+
			"- [ ] 2. In-run fix 2 — flow.verify: one test fails\n**Opener:** verify-demo\n")
	f.report("visual-verify", f.wt, vvcDepartures(2), "- steps: "+vvcSteps+"\n")

	// Round 2 opened from a verify report: its baseline is round 1's visual opener.
	f.appendRow("visual-verify-fix-2", f.wt)
	f.report("visual-verify-fix-2", f.wt, vvcDepartures(3), "- steps: "+vvcSteps+"\n")
	rc, out, _ := f.close("visual-verify-fix-2")
	if rc != 0 || !strings.Contains(out, "ESCALATE: departures 2 → 3 since visual-verify — the next round is a No-progress escalation\nNEXT: judge") {
		t.Fatalf("grown: exit %d\n%s", rc, out)
	}

	// As many departures as the opener is no escalation in round 2.
	f.appendRow("visual-verify-fix-2-full", f.wt)
	f.report("visual-verify-fix-2-full", f.wt, vvcDepartures(2), "- steps: "+vvcSteps+"\n")
	if rc, out, _ := f.close("visual-verify-fix-2-full"); rc != 0 || strings.Contains(out, "ESCALATE:") || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: judge") {
		t.Fatalf("equal: exit %d\n%s", rc, out)
	}

	// The third failed round escalates whatever its count, on judge and on loop.
	f.appendRow("visual-verify-fix-3", f.wt)
	f.report("visual-verify-fix-3", f.wt, vvcDepartures(1), "- steps: "+vvcSteps+"\n")
	if rc, out, _ := f.close("visual-verify-fix-3"); rc != 0 || !strings.Contains(out, "ESCALATE: round 3 is the loop's third failed round or later — the next round is a No-progress escalation\nNEXT: judge") {
		t.Fatalf("third, judge: exit %d\n%s", rc, out)
	}
	f.appendRow("visual-verify-fix-4", f.wt)
	f.report("visual-verify-fix-4", f.wt, vvcVerdict("1", vvcBlocked, "test-failure"), "- steps: "+vvcBlocked+"\n")
	if rc, out, _ := f.close("visual-verify-fix-4"); rc != 0 || strings.Contains(out, "ESCALATE:") || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: redispatch visual-verify-2-fix-4") {
		t.Fatalf("first block is a redispatch, no round failed: exit %d\n%s", rc, out)
	}
	f.appendRow("visual-verify-2-fix-4", f.wt)
	f.report("visual-verify-2-fix-4", f.wt, vvcVerdict("1", vvcBlocked, "test-failure"), "- steps: "+vvcBlocked+"\n")
	if rc, out, _ := f.close("visual-verify-2-fix-4"); rc != 0 || !strings.Contains(out, "ESCALATE: round 4 is the loop's third failed round or later — the next round is a No-progress escalation\nNEXT: loop") {
		t.Fatalf("fourth, loop: exit %d\n%s", rc, out)
	}
}

func TestVisualVerifyPassCloseNoReport(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "")
	rc, out, calls := f.close("visual-verify")
	if rc != 1 || !strings.Contains(out, "STOP: no report") || !strings.Contains(calls, "-outcome aborted") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
	if rc, _, calls := f.close("visual-verify-9"); rc != 2 || calls != "" {
		t.Fatalf("a key the stage file has no row for: exit %d, calls %q", rc, calls)
	}
}

func TestVisualVerifyPassCloseCommitsPathspec(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
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
		"Flow-Key: flow.visual-verify"} {
		if !strings.Contains(files, want) {
			t.Errorf("commit lacks %q:\n%s", want, files)
		}
	}
	if strings.Contains(files, "visual-verification") {
		t.Errorf("the commit carries the gitignored visual-verification directory:\n%s", files)
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
	// checkout printing PUSH:, its worktree on spectre/demo not.
	for _, inWorktree := range []bool{false, true} {
		h := vvcFixture(t, "")
		regress := filepath.Join(filepath.Dir(h.wt), "regress")
		writeFile(t, filepath.Join(regress, "README"), "r\n")
		fixtureRepo(t, regress, repoShape{Branch: "main"})
		gitRun(t, regress, "add", "-A")
		gitRun(t, regress, "commit", "-qm", "r")
		capture := regress
		if inWorktree {
			capture = filepath.Join(filepath.Dir(h.wt), "regress-worktrees", "demo")
			gitRun(t, regress, "worktree", "add", "-q", "-b", "spectre/demo", capture)
		}
		h.captures(capture)
		h.report("visual-verify", capture, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
		writeFile(t, filepath.Join(h.wt, ".flow/project.md"), "# project\n\n"+vvpSect(t, "| `regression checkout` | `"+regress+"` |\n| `regression repo` | `x/regress` |\n"))
		rc, out, _ := h.close("visual-verify")
		if rc != 0 || vvcLog(t, capture, "%s") != "test: record the visual-verify visual verification" ||
			strings.Contains(out, "PUSH:") == inWorktree ||
			(!inWorktree && !strings.Contains(out, "PUSH: git -C "+regress+" push")) {
			t.Fatalf("regression checkout (worktree %v): exit %d\n%s", inWorktree, rc, out)
		}
	}
}

func TestVisualVerifyPassCloseSkipsIgnoredPaths(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "")
	png := filepath.Join(f.wt, "spectre/changes/demo/visual-verification/home.png")
	writeFile(t, png, "png\n")
	f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n- verifier capture: "+png+"\n")
	rc, out, _ := f.close("visual-verify")
	if rc != 0 {
		t.Fatalf("exit %d\n%s", rc, out)
	}
	if !strings.Contains(out, "ROUND-BASE:") {
		t.Errorf("no ROUND-BASE:\n%s", out)
	}
	show, _ := exec.Command("git", "-C", f.wt, "show", "--name-only", "HEAD").Output()
	if !strings.Contains(string(show), "tests/visual/home.spec.ts-home.png\n") {
		t.Errorf("commit lacks the tracked capture:\n%s", show)
	}
	if strings.Contains(string(show), "visual-verification") {
		t.Errorf("commit carries the ignored capture:\n%s", show)
	}
}

func TestVisualVerifyPassCloseUsage(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "")
	for _, args := range [][]string{
		{f.wt, "demo", "visual-verify", "x", "--session-token", "mf-tok"},
		{f.wt, "demo", "visual-verify", "0"},
		{f.wt, "demo", "visual-verify", "0", "--session-token", "mf-tok", "--harness", "claude"},
	} {
		if rc, out, calls := f.run("close", args...); rc != 2 || !strings.Contains(out, "usage: flow-guard visual-verify-pass") || calls != "" {
			t.Errorf("%q: exit %d, calls %q\n%s", args, rc, calls, out)
		}
	}
}

// TestVerifierExitsTemplateOmitsMeasure pins kan-1060's task-12 review: close
// reads any non-zero `exits:` number as a failed command, and `measure` exits
// 1 on findings step 10 explains, so its exit lives on the `- measure:` line.
func TestVerifierExitsTemplateOmitsMeasure(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile("../../../skills/flow/visual-verify-verifier.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "- exits: ") && strings.Contains(l, "measure") {
			t.Errorf("the verdict's exits template names measure: %q", l)
		}
	}
}

// TestVisualVerifyPassCloseKeepsTheRound pins decision
// redispatch-key-keeps-round: a blocked `-fix-<k>` row's redispatch keeps the
// segment `flow record` requires while In-run fix k is open.
func TestVisualVerifyPassCloseKeepsTheRound(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "")
	f.appendRow("visual-verify-fix-3", f.wt)
	f.report("visual-verify-fix-3", f.wt, vvcVerdict("1", vvcBlocked, "test-failure"), "- steps: "+vvcBlocked+"\n")
	if rc, out, _ := f.close("visual-verify-fix-3"); rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: redispatch visual-verify-2-fix-3") {
		t.Fatalf("exit %d\n%s", rc, out)
	}
}

// TestVvpBaseInReadsEachCheckoutsRoundBase pins decision
// round-base-per-checkout, which measure's `changed:` line reads: a checkout
// that cannot resolve the round base is compared with the sha the bases file
// recorded for that checkout's own root.
func TestVvpBaseInReadsEachCheckoutsRoundBase(t *testing.T) {
	t.Parallel()
	other := t.TempDir()
	gitRun(t, other, "init", "-q")
	writeFile(t, filepath.Join(other, "fidelity/a.png"), "png\n")
	gitRun(t, other, "add", ".")
	gitRun(t, other, "-c", "user.email=a@b", "-c", "user.name=a", "commit", "-qm", "capture")
	otherBase := vgHead(t, other)
	foreign := strings.Repeat("0", 40)
	v := &vvp{env: Env{Getenv: os.Getenv}, bases: map[string]string{other: otherBase}}
	if got := v.baseIn(filepath.Join(other, "fidelity"), foreign); got != otherBase {
		t.Errorf("baseIn with the checkout's base recorded = %q, want %q", got, otherBase)
	}
	v.bases = nil
	if got := v.baseIn(filepath.Join(other, "fidelity"), foreign); got != foreign {
		t.Errorf("baseIn with no base recorded = %q, want the round base %q", got, foreign)
	}
}

// TestVvcParseVerdictIgnoresParenthesisedCounts: `capture 0 (28 passed)` is exit 0; a bare non-zero exit still counts.
func TestVvcParseVerdictIgnoresParenthesisedCounts(t *testing.T) {
	t.Parallel()
	if bad, _, _ := vvcParseVerdict("- exits: setup 0, capture 0 (28 passed), spec reach 0 (132 reached), measure 0/0"); bad {
		t.Error("a parenthesised count read as a non-zero exit")
	}
	if bad, _, _ := vvcParseVerdict("- exits: setup 0, capture 1 (CF5 test failure)"); !bad {
		t.Error("a bare non-zero exit was missed")
	}
}

func TestCloseOwesNoMergeBaseCapture(t *testing.T) {
	// A report carrying a merge-base capture owed line is asked no capture.
	o := vvcFixture(t, "")
	o.appendRow("visual-verify-2", o.wt)
	owed := "4 done | 8 done | 9 done | 10 blocked — merge-base capture owed | 11 done | motion n/a — no motion named"
	o.report("visual-verify-2", o.wt, vvcVerdict("0", owed, ""), "- merge-base capture owed: /abs/a.spec.ts\n- steps: "+owed+"\n")
	if rc, out, _ := o.close("visual-verify-2"); rc != 0 || strings.Contains(out, "MERGE-BASE-SPEC") || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: rest visual-verify-2-rest") {
		t.Fatalf("an owed merge-base capture: exit %d\n%s", rc, out)
	}
}

func TestCloseCheckOnlyWritesNothing(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "")
	path := f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
	good, _, calls := f.run("close", "--check-only", path, "0")
	bad := filepath.Join(f.wt, ".superpowers/sdd/verify-report-bad.md")
	writeFile(t, bad, "## Verdict\n- Model: opus\n")
	rc, out, calls2 := f.run("close", "--check-only", bad, "0")
	if good != 0 || rc != 1 || strings.TrimSpace(out) == "" {
		t.Fatalf("good exit %d, bad exit %d\n%s", good, rc, out)
	}
	if strings.Contains(calls+calls2, "dispatch") || strings.Contains(f.stageFile(), "verdict") {
		t.Errorf("check-only wrote state:\n%s\n%s", calls+calls2, f.stageFile())
	}
}

// TestCloseCheckOnlyFailsOnJudgeGap: a `no sidecar line` frame-list gap is
// judged in a full close, but check-only has no judge and must exit 1 on it.
func TestCloseCheckOnlyFailsOnJudgeGap(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "| `mockups` | `design/mockups` |\n")
	path := f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- frames: 1/1\n- f1 note: composed\n- steps: "+vvcSteps+"\n")
	rc, out, _ := f.run("close", "--check-only", path, "0", "--design", filepath.Join(f.wt, "spectre/changes/demo/design.md"), "--mockups", filepath.Join(f.wt, "design/mockups"))
	if rc != 1 || !strings.Contains(out, "no sidecar line: f1") {
		t.Fatalf("exit %d, want 1 naming the gap\n%s", rc, out)
	}
}

// TestCloseCheckOnlyDerivesMechanicalLines: close reads the spec, each frame's
// composite and diff, and the frame count from the manifest, so a report
// carrying only judgements is complete; a manifest frame with no verdict line,
// nor named on the frames line, still leaves it incomplete.
func TestCloseCheckOnlyDerivesMechanicalLines(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "| `mockups` | `design/mockups` |\n")
	writeFile(t, filepath.Join(f.wt, "tests/visual/home.spec.ts.mockups"), "home.png f1\n")
	design := filepath.Join(f.wt, "spectre/changes/demo/design.md")
	shots := filepath.Join(f.wt, "tests/visual")
	pair := vvmPair{Spec: shots + "/home.spec.ts", Frame: "f1", Capture: shots + "/home.spec.ts-home.png", Diff: "0.01", Bands: []any{}, Seams: []any{}, Over: []any{}}
	args := func(path string) []string {
		f.manifest(path, shots, pair)
		return []string{"--check-only", path, "0", "--design", design, "--mockups", filepath.Join(f.wt, "design/mockups")}
	}
	path := f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- f1: match\n- steps: "+vvcSteps+"\n")
	if rc, out, _ := f.run("close", args(path)...); rc != 0 {
		t.Errorf("a judgement-only report: exit %d, want 0\n%s", rc, out)
	}
	path = f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
	if rc, out, _ := f.run("close", args(path)...); rc != 1 || !strings.Contains(out, "f1 has no `- f1: match | differs` line") {
		t.Errorf("a frame verdict omitted: exit %d, want 1 naming f1\n%s", rc, out)
	}
}

func TestCloseCheckOnlyMockupsNeedsDesign(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "")
	path := f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
	rc, out, _ := f.run("close", "--check-only", path, "0", "--mockups", f.wt)
	if rc != 2 || !strings.Contains(out, "--mockups needs --design") {
		t.Fatalf("exit %d, want usage exit 2\n%s", rc, out)
	}
}

// TestVisualVerifyPassCloseRefusesAStaleRecord (KAN-1155): a record naming another pass, or none, is
// a remainder with a RECORD-STALE verdict; one naming this pass, or no record file at all, closes done.
func TestVisualVerifyPassCloseRefusesAStaleRecord(t *testing.T) {
	t.Parallel()
	record := func(f *vvpFix) string { return filepath.Join(f.wt, "spectre/changes/demo/visual-verification.md") }
	for _, c := range []struct{ name, record, want string }{
		{"another pass", "Pass: fix-13-rest\n\n# views\n", "NEXT: rest visual-verify-rest"},
		{"no pass line", "# views\n", "NEXT: rest visual-verify-rest"},
		{"this pass", "Pass: visual-verify\n\n# views\n", "NEXT: done"},
		{"no record", "", "NEXT: done"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			f := vvcFixture(t, "")
			if c.record != "" {
				writeFile(t, record(f), c.record)
			}
			f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
			rc, out, _ := f.close("visual-verify")
			if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), c.want) || (c.want != "NEXT: done") != strings.Contains(out, "RECORD-STALE: ") {
				t.Fatalf("exit %d, want %s\n%s", rc, c.want, out)
			}
		})
	}
}

func TestVvcReportFramesNamesTheRemaindersOwnFrames(t *testing.T) {
	t.Parallel()
	r := []byte("- frames: v2/LD12 (earlier pass, report a.md), v2/LD6, v2/LD1\n")
	if got := vvcReportFrames(r); got != "v2/LD12,v2/LD6,v2/LD1" {
		t.Fatalf("got %q", got)
	}
	if got := vvcReportFrames([]byte("- steps: x\n")); got != "" {
		t.Fatalf("got %q", got)
	}
}
