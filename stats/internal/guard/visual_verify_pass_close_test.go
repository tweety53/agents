package guard

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// A die after the stage opened ends the stage rather than leaving it open.
func TestVisualVerifyPassCloseDieEndsStage(t *testing.T) {
	t.Parallel()
	f := vvcFixture(t, "")
	writeExec(t, filepath.Join(f.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+f.log+"\ncase \"$1 $2 $3\" in \"record dispatch end\") echo store down; exit 1;; esac\n")
	f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
	rc, out, calls := f.close("visual-verify")
	if rc != 2 || !strings.Contains(calls, "stage end -C "+f.wt+" -command /flow -stage flow.visual-verify -outcome stopped demo") {
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
		!strings.Contains(calls, "-outcome blocked -cause missing-fixture") || !strings.Contains(calls, "-stage flow.visual-verify -outcome stopped") || !strings.Contains(out, "STAGE: stopped") {
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

// TestVisualVerifyPassCloseClosesTheLoop pins decisions closes-loop-writer,
// closes-loop-derived-in-close, capture-freshness-by-mtime,
// freshness-derived-in-close, closing-frame-list,
// frame-map-change-keeps-full-run, prototype-change-keeps-full-run and
// close-prints-closes-loop-line: a scoped row ends -closes-loop with
// `CLOSES-LOOP: yes` only when every frame of the closing list was
// re-captured after the round file's instant, every changed frame has a line
// of its own, no sidecar moved and the prototype renders match the last
// verifier's snapshot; an unscoped row prints no CLOSES-LOOP: line.
func TestVisualVerifyPassCloseClosesTheLoop(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	const judged = "- f1 sweeps: text done\n"
	vvDir := func(f *vvpFix) string { return filepath.Join(f.wt, "spectre/changes/demo/visual-verification") }
	proto := func(f *vvpFix) {
		writeFile(t, filepath.Join(vvDir(f), "prototype.mockups"), "frame none\nhome.png f1\n")
		writeFile(t, filepath.Join(vvDir(f), "prototype/f1.png"), "png\n")
	}
	snap := func(dir string) string {
		s := ""
		for _, name := range []string{"prototype.mockups", "prototype/f1.png"} {
			b, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			s += fmt.Sprintf("%x %s\n", sha256.Sum256(b), name)
		}
		return s
	}
	const mockupsCell = "| `mockups` | `design/mockups` |\n"
	// sidecar commits the home spec's frame map, so a mockups cell's frame list
	// has a sidecar line for f1 and no sidecar reads changed.
	sidecar := func(f *vvpFix, body string) {
		writeFile(t, filepath.Join(f.wt, "tests/visual/home.spec.ts.mockups"), body)
		gitRun(t, f.wt, "add", "tests/visual/home.spec.ts.mockups")
		gitRun(t, f.wt, "commit", "-qm", "sidecar")
	}
	snapFile := func(f *vvpFix) string { return filepath.Join(f.wt, ".superpowers/sdd/visual-prototype.txt") }
	// fix commits the round's one source edit and returns its blob.
	fix := func(f *vvpFix) string {
		writeFile(t, filepath.Join(f.wt, "src/ui/home.tsx"), "fixed\n")
		gitRun(t, f.wt, "add", "src/ui/home.tsx")
		gitRun(t, f.wt, "commit", "-qm", "fix")
		return tcfGit(t, f.wt, "rev-parse", "HEAD:src/ui/home.tsx")
	}
	cases := []struct {
		name     string
		setup    func(f *vvpFix)
		lines    string
		stale    bool
		unscoped bool
		want     []string
		closes   bool
		isProto  bool
		extra    string
		noPairs  bool
		after    func(f *vvpFix) // runs once the round base is read
	}{
		{name: "fresh, the changed frame judged", lines: judged, want: []string{"CLOSES-LOOP: yes"}, closes: true},
		{name: "an unchanged frame needs no line", setup: func(f *vvpFix) {
			gitRun(t, f.wt, "add", "tests/visual/home.spec.ts-home.png")
			gitRun(t, f.wt, "commit", "-qm", "capture")
		}, want: []string{"CLOSES-LOOP: yes"}, closes: true},
		{name: "a stale capture", lines: judged, stale: true, want: []string{"CLOSES-LOOP: no — f1: "}},
		{name: "a stale capture the round's recapture passed on HEAD's sources", lines: judged, stale: true, after: func(f *vvpFix) {
			vvpReceipt(t, f.wt, "tests/visual/home.spec.ts", "src/ui/home.tsx", fix(f))
		}, want: []string{"CLOSES-LOOP: yes"}, closes: true},
		{name: "a stale capture a recapture passed on an earlier draft", lines: judged, stale: true, after: func(f *vvpFix) {
			fix(f)
			vvpReceipt(t, f.wt, "tests/visual/home.spec.ts", "src/ui/home.tsx", "0000000000000000000000000000000000000000")
		}, want: []string{"CLOSES-LOOP: no — f1: "}},
		{name: "a changed frame on the frames line alone", lines: "- frames: 0/1 — f1: capture unchanged since x\n", want: []string{"CLOSES-LOOP: no — f1: "}},
		{name: "a frame the prototype map lists and no pair carries", setup: func(f *vvpFix) {
			writeFile(t, filepath.Join(vvDir(f), "prototype.mockups"), "x.png f2\n")
		}, lines: judged, want: []string{"CLOSES-LOOP: no — f2: "}},
		{name: "a sidecar changed since the round base", setup: func(f *vvpFix) {
			writeFile(t, filepath.Join(f.wt, "tests/visual/home.spec.ts.mockups"), "home.png f1\n")
		}, lines: judged, want: []string{"CLOSES-LOOP: no — ", "home.spec.ts.mockups"}},
		{name: "unscoped", lines: judged, unscoped: true},
		{name: "a prototype render unchanged since the last completed verifier", setup: func(f *vvpFix) {
			proto(f)
			writeFile(t, snapFile(f), snap(vvDir(f)))
		}, lines: judged, want: []string{"CLOSES-LOOP: yes"}, closes: true, isProto: true},
		{name: "a prototype render changed since the last completed verifier", setup: func(f *vvpFix) {
			proto(f)
			writeFile(t, snapFile(f), snap(vvDir(f)))
			writeFile(t, filepath.Join(vvDir(f), "prototype/f1.png"), "new\n")
		}, lines: judged, want: []string{"CLOSES-LOOP: no — ", "visual-prototype.txt"}, isProto: true},
		{name: "a prototype map with no snapshot", setup: proto, lines: judged, want: []string{"CLOSES-LOOP: no — ", "visual-prototype.txt"}, isProto: true},
		{name: "a declared frame no pair carries", extra: mockupsCell, setup: func(f *vvpFix) {
			writeFile(t, filepath.Join(f.wt, "design/mockups/f2.png"), "png\n")
			writeFile(t, filepath.Join(f.wt, "spectre/changes/demo/design.md"), "# design\n\nFrame f1 is the home screen.\nFrame f2 is the menu.\n")
			sidecar(f, "home.png f1\nmenu.png f2\n")
		}, lines: "- frames: 2/2\n" + judged + "- f2 sweeps: text done\n", want: []string{"CLOSES-LOOP: no — f2: "}},
		{name: "a mockups-root file changed since the round base", extra: mockupsCell, setup: func(f *vvpFix) {
			sidecar(f, "home.png f1\n")
			writeFile(t, filepath.Join(f.wt, "design/mockups/f1.png"), "new\n")
		}, lines: "- frames: 1/1\n" + judged, want: []string{"CLOSES-LOOP: no — design/mockups/f1.png changed since"}},
		{name: "an empty closing list", noPairs: true, want: []string{"CLOSES-LOOP: no — no frame on the change's list"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := vvcFixture(t, c.extra)
			if c.setup != nil {
				c.setup(f)
			}
			base := vgHead(t, f.wt)
			if c.after != nil {
				c.after(f)
			}
			at := time.Now().Add(-time.Minute)
			capture := filepath.Join(f.wt, "tests/visual/home.spec.ts-home.png")
			mtime := time.Now()
			if c.stale {
				mtime = at.Add(-time.Hour)
			}
			if err := os.Chtimes(capture, mtime, mtime); err != nil {
				t.Fatal(err)
			}
			sdd := filepath.Join(f.wt, ".superpowers/sdd")
			if !c.unscoped {
				writeFile(t, filepath.Join(sdd, "visual-round-visual-verify.txt"), base+" "+at.UTC().Format(time.RFC3339Nano))
			}
			pairs := `[{"spec":"` + filepath.Join(f.wt, "tests/visual/home.spec.ts") + `","frame":"f1","capture":"` + capture + `"}]`
			if c.noPairs {
				pairs = "[]"
			}
			writeFile(t, filepath.Join(sdd, "visual-manifest-visual-verify.json"), `{"pairs":`+pairs+`}`)
			f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n"+c.lines)
			rc, out, calls := f.close("visual-verify")
			if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: done") {
				t.Fatalf("exit %d, want 0 ending NEXT: done\n%s\n%s", rc, out, calls)
			}
			for _, w := range c.want {
				if !strings.Contains(out, w) {
					t.Errorf("output lacks %q", w)
				}
			}
			if c.want == nil && strings.Contains(out, "CLOSES-LOOP:") {
				t.Errorf("an unscoped row printed a CLOSES-LOOP: line")
			}
			if got := strings.Contains(calls, "-closes-loop"); got != c.closes {
				t.Errorf("-closes-loop in the calls %v, want %v", got, c.closes)
			}
			if c.isProto {
				b, err := os.ReadFile(snapFile(f))
				if err != nil || string(b) != snap(vvDir(f)) {
					t.Errorf("visual-prototype.txt %q (%v), want the map and renders now %q", b, err, snap(vvDir(f)))
				}
			}
			if t.Failed() {
				t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
			}
		})
	}
}

// TestVisualVerifyPassCloseDemandsDefectLines pins decision
// defects-close-check: a frame the round file names under defects: needs a
// per-frame line of its own before close ends the row -closes-loop, even when
// its capture is unchanged since the round base (kan-1122 self-review, F1).
func TestVisualVerifyPassCloseDemandsDefectLines(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name, lines, want string
		closes            bool
	}{
		{"no line of its own", "", "CLOSES-LOOP: no — f1: named under defects: and the report has no line of its own for it", false},
		{"a line of its own", "- f1 sweeps: text done\n", "CLOSES-LOOP: yes", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := vvcFixture(t, "")
			// The capture committed at the round base: unchanged, so only defects: asks for its line.
			gitRun(t, f.wt, "add", "tests/visual/home.spec.ts-home.png")
			gitRun(t, f.wt, "commit", "-qm", "capture")
			base := vgHead(t, f.wt)
			at := time.Now().Add(-time.Minute)
			capture := filepath.Join(f.wt, "tests/visual/home.spec.ts-home.png")
			now := time.Now()
			if err := os.Chtimes(capture, now, now); err != nil {
				t.Fatal(err)
			}
			sdd := filepath.Join(f.wt, ".superpowers/sdd")
			writeFile(t, filepath.Join(sdd, "visual-round-visual-verify.txt"), base+" "+at.UTC().Format(time.RFC3339Nano)+" f1")
			writeFile(t, filepath.Join(sdd, "visual-manifest-visual-verify.json"),
				`{"pairs":[{"spec":"`+filepath.Join(f.wt, "tests/visual/home.spec.ts")+`","frame":"f1","capture":"`+capture+`"}]}`)
			f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n"+c.lines)
			rc, out, calls := f.close("visual-verify")
			if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: done") || !strings.Contains(out, c.want) ||
				strings.Contains(calls, "-closes-loop") != c.closes {
				t.Fatalf("exit %d, want 0 with %q and -closes-loop %v\n%s\n%s", rc, c.want, c.closes, out, calls)
			}
		})
	}
}

func TestVisualVerifyPassCloseNoReport(t *testing.T) {
	t.Parallel()
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
		writeFile(t, filepath.Join(h.wt, ".flow/project.md"), "# project\n\n"+vvpSection+"| `regression checkout` | `"+regress+"` |\n| `regression repo` | `x/regress` |\n")
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

// TestVisualVerifyPassCloseShards pins decisions shard-key-segment,
// shard-unscoped-single-worktree, record-parts-merged-by-close and
// closes-serial-no-lock (KAN-1121): a shard's blocked count, remainder and
// redispatch keys are its own index's, its report guards run on its slice,
// the stage ends only once every index's newest row is closed, and
// visual-verification.md is the shards' parts joined in order.
func TestVisualVerifyPassCloseShards(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f := vvcFixture(t, "")
	f.appendRow("visual-verify-s1of3", f.wt)
	f.appendRow("visual-verify-s2of3", f.wt)
	writeFile(t, f.stagePath(), f.stageFile()+"blocked visual-verify-s1of3\n")
	f.report("visual-verify-s2of3", f.wt, vvcVerdict("1", vvcBlocked, "environment"), "- steps: "+vvcBlocked+"\n")
	if rc, out, _ := f.close("visual-verify-s2of3"); rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: redispatch visual-verify-2-s2of3") {
		t.Fatalf("another index's block counted: exit %d\n%s", rc, out)
	}

	// A shard names the spec it needs a merge-base capture for; close asks
	// the parent once per spec, before the remainder.
	o := vvcFixture(t, "")
	o.appendRow("visual-verify-s1of2", o.wt)
	owed := "4 done | 8 done | 9 done | 10 blocked — merge-base capture owed | 11 done | motion n/a — no motion named"
	o.report("visual-verify-s1of2", o.wt, vvcVerdict("0", owed, ""), "- merge-base capture owed: /abs/a.spec.ts\n- merge-base capture owed: /abs/a.spec.ts\n- steps: "+owed+"\n")
	if rc, out, _ := o.close("visual-verify-s1of2"); rc != 0 || strings.Count(out, "MERGE-BASE-SPEC: /abs/a.spec.ts\n") != 1 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: rest visual-verify-rest-s1of2") {
		t.Fatalf("an owed merge-base capture: exit %d\n%s", rc, out)
	}

	// A shard of a set of several worktrees: the repository name, then the
	// shard segment, on its remainder and its redispatch alike.
	m := vvcFixture(t, "")
	writeFile(t, m.stagePath(), strings.Replace(m.stageFile(), "set 1\n", "set 2\n", 1))
	m.appendRow("visual-verify-demo-s2of3", m.wt)
	m.report("visual-verify-demo-s2of3", m.wt, vvcVerdict("0", "10 not done", ""), "- steps: 10 not done\n")
	if rc, out, _ := m.close("visual-verify-demo-s2of3"); rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: rest visual-verify-rest-demo-s2of3") {
		t.Fatalf("a suffixed shard's remainder: exit %d\n%s", rc, out)
	}
	m.appendRow("visual-verify-rest-demo-s2of3", m.wt)
	m.report("visual-verify-rest-demo-s2of3", m.wt, vvcVerdict("1", vvcBlocked, "environment"), "- steps: "+vvcBlocked+"\n")
	if rc, out, _ := m.close("visual-verify-rest-demo-s2of3"); rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: redispatch visual-verify-2-demo-s2of3") {
		t.Fatalf("a suffixed shard's redispatch: exit %d\n%s", rc, out)
	}

	g := vvcFixture(t, "| `mockups` | `design/mockups` |\n")
	vv := filepath.Join(g.wt, "spectre/changes/demo/visual-verification")
	writeFile(t, filepath.Join(g.wt, "spectre/changes/demo/design.md"), "# design\n\nFrames f1 and f2.\n")
	writeFile(t, filepath.Join(g.wt, "design/mockups/f2.png"), "png\n")
	writeFile(t, filepath.Join(g.wt, "tests/visual/home.spec.ts.mockups"), "one.png f1\ntwo.png f2\n")
	writeFile(t, filepath.Join(vv, "record-s1of2.md"), "# s1\n")
	writeFile(t, filepath.Join(vv, "record-s2of2.md"), "# s2\n")
	writeFile(t, g.stagePath(), strings.Replace(g.stageFile(), "row visual-verify "+g.wt+"\n", "", 1)+
		"slice s1of2 f1\nslice s2of2 f2\nrow visual-verify-s1of2 "+g.wt+"\nrow visual-verify-s2of2 "+g.wt+"\n")
	undone := strings.Replace(vvcSteps, "10 done", "10 not done", 1)
	g.report("visual-verify-s2of2", g.wt, vvcVerdict("0", undone, ""), "- frames: 1/1\n- f2 note: composed\n- steps: "+undone+"\n")
	if rc, out, _ := g.close("visual-verify-s2of2"); rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: rest visual-verify-rest-s2of2") {
		t.Fatalf("a shard's remainder: exit %d\n%s", rc, out)
	}
	g.appendRow("visual-verify-rest-s2of2", g.wt)
	g.report("visual-verify-rest-s2of2", g.wt, vvcVerdict("0", vvcSteps, ""), "- frames: 1/1\n- f2 note: composed\n- steps: "+vvcSteps+"\n")
	rc, out, calls := g.close("visual-verify-rest-s2of2")
	if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: done") || strings.Contains(out, "STAGE:") || strings.Contains(calls, "stage end") {
		t.Fatalf("the last index closed first ended the stage, or its slice's report was refused: exit %d\n%s\n%s", rc, out, calls)
	}
	g.report("visual-verify-s1of2", g.wt, vvcVerdict("0", vvcSteps, ""), "- frames: 1/1\n- f1 note: composed\n- steps: "+vvcSteps+"\n")
	rc, out, calls = g.close("visual-verify-s1of2")
	if rc != 0 || !strings.Contains(out, "STAGE: completed") || !strings.Contains(calls, "stage end -C "+g.wt+" -command /flow -stage flow.visual-verify -outcome completed demo") {
		t.Fatalf("every index closed: exit %d\n%s\n%s", rc, out, calls)
	}
	if b, _ := os.ReadFile(filepath.Join(filepath.Dir(vv), "visual-verification.md")); string(b) != "# s1\n\n# s2\n" {
		t.Fatalf("visual-verification.md = %q, want the parts joined in shard order", b)
	}
}

// TestVisualVerifyPassCloseRemainderClosesTheLoop pins decision
// rest-closes-the-loop: a `-rest` remainder reads the continued key's round
// file and manifest, so a remainder that completes a scoped pass whose frames
// were all re-captured and judged ends -closes-loop; without the base key's
// files it stays unscoped.
func TestVisualVerifyPassCloseRemainderClosesTheLoop(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name  string
		round bool
		want  bool
	}{
		{"the continued key has a round file", true, true},
		{"no round file for the continued key", false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := vvcFixture(t, "")
			base := vgHead(t, f.wt)
			at := time.Now().Add(-time.Minute)
			capture := filepath.Join(f.wt, "tests/visual/home.spec.ts-home.png")
			now := time.Now()
			if err := os.Chtimes(capture, now, now); err != nil {
				t.Fatal(err)
			}
			sdd := filepath.Join(f.wt, ".superpowers/sdd")
			stage, err := os.ReadFile(filepath.Join(sdd, "visual-verify-stage.txt"))
			if err != nil {
				t.Fatal(err)
			}
			var rows []string
			for _, l := range strings.Split(string(stage), "\n") {
				if strings.Contains(l, "visual-verify ") {
					rows = append(rows, strings.Replace(l, "visual-verify ", "visual-verify-rest ", 1))
				}
			}
			writeFile(t, filepath.Join(sdd, "visual-verify-stage.txt"), string(stage)+strings.Join(rows, "\n")+"\n")
			if c.round {
				writeFile(t, filepath.Join(sdd, "visual-round-visual-verify.txt"), base+" "+at.UTC().Format(time.RFC3339Nano))
			}
			pairs := `[{"spec":"` + filepath.Join(f.wt, "tests/visual/home.spec.ts") + `","frame":"f1","capture":"` + capture + `"}]`
			writeFile(t, filepath.Join(sdd, "visual-manifest-visual-verify.json"), `{"pairs":`+pairs+`}`)
			f.report("visual-verify-rest", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n- f1 sweeps: text done\n")
			rc, out, calls := f.close("visual-verify-rest")
			if rc != 0 {
				t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
			}
			if got := strings.Contains(calls, "-closes-loop"); got != c.want {
				t.Errorf("-closes-loop in the calls %v, want %v\n%s\n%s", got, c.want, out, calls)
			}
		})
	}
}

// TestVisualVerifyPassCloseIdenticalRecaptureIsFresh pins decision
// identical-recapture-fresh: a capture Playwright left untouched because the
// recapture was byte-identical counts as re-captured when it is unchanged by
// blob since the round base and a test-results entry is newer than the row's
// start; without that entry, or with a changed blob, it stays stale.
func TestVisualVerifyPassCloseIdenticalRecaptureIsFresh(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name     string
		ran      bool
		changed  bool
		wantLoop bool
		oldRun   bool
	}{
		{"identical and a run after the row began", true, false, true, false},
		{"identical and no run output", false, false, false, false},
		{"a run, but the capture differs from the base blob", true, true, false, false},
		{"identical and only a run output older than the row", false, false, false, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := vvcFixture(t, "")
			gitRun(t, f.wt, "add", "tests/visual/home.spec.ts-home.png")
			gitRun(t, f.wt, "commit", "-qm", "capture")
			at := time.Now().Add(-time.Minute)
			capture := filepath.Join(f.wt, "tests/visual/home.spec.ts-home.png")
			if c.changed {
				writeFile(t, capture, "different\n")
			}
			old := at.Add(-time.Hour)
			if err := os.Chtimes(capture, old, old); err != nil {
				t.Fatal(err)
			}
			if c.ran {
				writeFile(t, filepath.Join(f.wt, "tests/visual/test-results/run-1/trace.txt"), "x\n")
				run := filepath.Join(f.wt, "tests/visual/test-results/run-1")
				now := time.Now()
				if err := os.Chtimes(run, now, now); err != nil {
					t.Fatal(err)
				}
			}
			if c.oldRun {
				run := filepath.Join(f.wt, "tests/visual/test-results/run-0")
				writeFile(t, filepath.Join(run, "trace.txt"), "x\n")
				if err := os.Chtimes(run, old, old); err != nil {
					t.Fatal(err)
				}
			}
			sdd := filepath.Join(f.wt, ".superpowers/sdd")
			// The round base is the canonical worktree's sha: it resolves in no other repository.
			writeFile(t, filepath.Join(sdd, "visual-round-visual-verify.txt"), strings.Repeat("0", 40)+" "+at.UTC().Format(time.RFC3339Nano))
			pairs := `[{"spec":"` + filepath.Join(f.wt, "tests/visual/home.spec.ts") + `","frame":"f1","capture":"` + capture + `"}]`
			writeFile(t, filepath.Join(sdd, "visual-manifest-visual-verify.json"), `{"pairs":`+pairs+`}`)
			f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n- f1 sweeps: text done\n")
			rc, out, calls := f.close("visual-verify")
			if rc != 0 {
				t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
			}
			if got := strings.Contains(calls, "-closes-loop"); got != c.wantLoop {
				t.Errorf("-closes-loop %v, want %v\n%s", got, c.wantLoop, out)
			}
		})
	}
}

// TestVisualVerifyPassCloseIdenticalRecaptureInAnotherCheckout pins decision
// identical-recapture-fresh: the capture and its test-results live in a
// checkout other than the canonical worktree (the playwright repository), and
// the proof walks up to that checkout's own root.
func TestVisualVerifyPassCloseIdenticalRecaptureInAnotherCheckout(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f := vvcFixture(t, "")
	other := t.TempDir()
	gitRun(t, other, "init", "-q")
	capture := filepath.Join(other, "fidelity/home.spec.ts-snapshots/home-darwin.png")
	writeFile(t, capture, "png\n")
	gitRun(t, other, "add", ".")
	gitRun(t, other, "-c", "user.email=a@b", "-c", "user.name=a", "commit", "-qm", "capture")
	at := time.Now().Add(-time.Minute)
	old := at.Add(-time.Hour)
	if err := os.Chtimes(capture, old, old); err != nil {
		t.Fatal(err)
	}
	run := filepath.Join(other, "test-results/run-1")
	writeFile(t, filepath.Join(run, "trace.txt"), "x\n")
	now := time.Now()
	if err := os.Chtimes(run, now, now); err != nil {
		t.Fatal(err)
	}
	sdd := filepath.Join(f.wt, ".superpowers/sdd")
	writeFile(t, filepath.Join(sdd, "visual-round-visual-verify.txt"), strings.Repeat("0", 40)+" "+at.UTC().Format(time.RFC3339Nano))
	pairs := `[{"spec":"` + filepath.Join(other, "fidelity/home.spec.ts") + `","frame":"f1","capture":"` + capture + `"}]`
	writeFile(t, filepath.Join(sdd, "visual-manifest-visual-verify.json"), `{"pairs":`+pairs+`}`)
	f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n- f1 sweeps: text done\n")
	rc, out, calls := f.close("visual-verify")
	if rc != 0 {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
	if !strings.Contains(calls, "-closes-loop") {
		t.Errorf("a capture in another checkout, identical and run after the row began, did not close the loop\n%s", out)
	}
}

// TestVisualVerifyPassCloseReadsEachCheckoutsRoundBase pins decision
// round-base-per-checkout: a capture in another repository is compared with
// the round base prepare recorded for that checkout (the .bases file), since
// the round file's sha is the canonical worktree's and resolves nowhere else.
// An unchanged capture then needs no line of its own; without the file it
// reads as changed and does.
func TestVisualVerifyPassCloseReadsEachCheckoutsRoundBase(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name  string
		bases bool
		want  bool
	}{
		{"the checkout's own base recorded", true, true},
		{"no base for the checkout", false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := vvcFixture(t, "")
			other := t.TempDir()
			gitRun(t, other, "init", "-q")
			capture := filepath.Join(other, "fidelity/home.spec.ts-snapshots/home-darwin.png")
			writeFile(t, capture, "png\n")
			gitRun(t, other, "add", ".")
			gitRun(t, other, "-c", "user.email=a@b", "-c", "user.name=a", "commit", "-qm", "capture")
			otherBase := vgHead(t, other)
			at := time.Now().Add(-time.Minute)
			now := time.Now()
			if err := os.Chtimes(capture, now, now); err != nil { // fresh by mtime: only the base is under test
				t.Fatal(err)
			}
			sdd := filepath.Join(f.wt, ".superpowers/sdd")
			writeFile(t, filepath.Join(sdd, "visual-round-visual-verify.txt"), strings.Repeat("0", 40)+" "+at.UTC().Format(time.RFC3339Nano))
			if c.bases {
				writeFile(t, filepath.Join(sdd, "visual-round-visual-verify.bases"), other+" "+otherBase+"\n")
			}
			pairs := `[{"spec":"` + filepath.Join(other, "fidelity/home.spec.ts") + `","frame":"f1","capture":"` + capture + `"}]`
			writeFile(t, filepath.Join(sdd, "visual-manifest-visual-verify.json"), `{"pairs":`+pairs+`}`)
			f.report("visual-verify", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
			rc, out, calls := f.close("visual-verify")
			if rc != 0 {
				t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
			}
			if got := strings.Contains(calls, "-closes-loop"); got != c.want {
				t.Errorf("-closes-loop %v, want %v\n%s", got, c.want, out)
			}
		})
	}
}

// TestVisualVerifyPassCloseRemainderReadsEachCheckoutsRoundBase pins decisions
// rest-closes-the-loop and round-base-per-checkout together: a `-rest` close
// reads the continued key's bases file, so a capture in another checkout is
// compared with that checkout's own round base.
func TestVisualVerifyPassCloseRemainderReadsEachCheckoutsRoundBase(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, c := range []struct {
		name  string
		bases bool
		want  bool
	}{
		{"the checkout's own base recorded", true, true},
		{"no base for the checkout", false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := vvcFixture(t, "")
			other := t.TempDir()
			gitRun(t, other, "init", "-q")
			capture := filepath.Join(other, "fidelity/home.spec.ts-snapshots/home-darwin.png")
			writeFile(t, capture, "png\n")
			gitRun(t, other, "add", ".")
			gitRun(t, other, "-c", "user.email=a@b", "-c", "user.name=a", "commit", "-qm", "capture")
			otherBase := vgHead(t, other)
			at := time.Now().Add(-time.Minute)
			now := time.Now()
			if err := os.Chtimes(capture, now, now); err != nil {
				t.Fatal(err)
			}
			sdd := filepath.Join(f.wt, ".superpowers/sdd")
			stage, err := os.ReadFile(filepath.Join(sdd, "visual-verify-stage.txt"))
			if err != nil {
				t.Fatal(err)
			}
			var rows []string
			for _, l := range strings.Split(string(stage), "\n") {
				if strings.Contains(l, "visual-verify ") {
					rows = append(rows, strings.Replace(l, "visual-verify ", "visual-verify-rest ", 1))
				}
			}
			writeFile(t, filepath.Join(sdd, "visual-verify-stage.txt"), string(stage)+strings.Join(rows, "\n")+"\n")
			writeFile(t, filepath.Join(sdd, "visual-round-visual-verify.txt"), strings.Repeat("0", 40)+" "+at.UTC().Format(time.RFC3339Nano))
			if c.bases {
				writeFile(t, filepath.Join(sdd, "visual-round-visual-verify.bases"), other+" "+otherBase+"\n")
			}
			pairs := `[{"spec":"` + filepath.Join(other, "fidelity/home.spec.ts") + `","frame":"f1","capture":"` + capture + `"}]`
			writeFile(t, filepath.Join(sdd, "visual-manifest-visual-verify.json"), `{"pairs":`+pairs+`}`)
			f.report("visual-verify-rest", f.wt, vvcVerdict("0", vvcSteps, ""), "- steps: "+vvcSteps+"\n")
			rc, out, calls := f.close("visual-verify-rest")
			if rc != 0 {
				t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
			}
			if got := strings.Contains(calls, "-closes-loop"); got != c.want {
				t.Errorf("-closes-loop %v, want %v\n%s", got, c.want, out)
			}
		})
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
