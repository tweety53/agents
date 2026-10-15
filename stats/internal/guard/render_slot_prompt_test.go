package guard

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// rspSkillDir is this checkout's real skills/flow, absolute.
func rspSkillDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../../../skills/flow")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// rspFx is a canonical worktree and a peer, each with every diff's touched
// list written as write-panel-diff.sh writes it, and a plan directory.
type rspFx struct{ canon, peer, plan string }

func rspNewFx(t *testing.T) rspFx {
	t.Helper()
	dir := t.TempDir()
	fx := rspFx{canon: dir + "/canon", peer: dir + "/peer", plan: dir + "/plan"}
	for _, wt := range []string{fx.canon, fx.peer} {
		mkdir(t, wt+"/.superpowers/sdd")
	}
	for _, f := range []string{"final-review.diff", "late-fix.diff", "fix-round-2.diff",
		"slot-delta-2-primary.diff", "slot-delta-2-principles.diff"} {
		writeFile(t, fx.canon+"/.superpowers/sdd/"+f+".touched", "# worktree: "+fx.canon+" — merge base abc\nM\t"+f+"-touched.go\n")
	}
	mkdir(t, fx.plan)
	for _, f := range []string{"proposal.md", "design.md", "tasks.md"} {
		writeFile(t, fx.plan+"/"+f, f)
	}
	return fx
}

func (fx rspFx) run(t *testing.T, skill string, args ...string) (guardResult, string) {
	t.Helper()
	r := runGuard("render-slot-prompt", args, Env{Dir: fx.canon, Getenv: os.Getenv})
	path := strings.TrimSuffix(r.stdout, "\n")
	if r.rc != 0 {
		return r, ""
	}
	return r, rbRead(path)
}

// rspTestBlock is the blockquote in file opening `> **<label>:**`, read here
// independently of the renderer: the first such line through the last
// consecutive `>` line.
func rspTestBlock(t *testing.T, file, label string) string {
	t.Helper()
	var out []string
	for _, l := range strings.Split(rbRead(file), "\n") {
		if len(out) == 0 && !strings.HasPrefix(l, "> **"+label+":**") {
			continue
		}
		if !strings.HasPrefix(l, ">") {
			break
		}
		out = append(out, l)
	}
	if len(out) == 0 {
		t.Fatalf("%s carries no %s block", file, label)
	}
	return strings.Join(out, "\n")
}

// rspTablePlaceholders is every placeholder review-panel.md's table defines.
func rspTablePlaceholders(t *testing.T, skill string) []string {
	t.Helper()
	re := regexp.MustCompile("(?m)^\\| `(\\[[A-Z_]+\\])` \\|")
	var out []string
	for _, m := range re.FindAllStringSubmatch(rbRead(skill+"/review-panel.md"), -1) {
		out = append(out, m[1])
	}
	if len(out) < 7 {
		t.Fatalf("review-panel.md's placeholder table yielded %v", out)
	}
	return out
}

func TestRenderSlotPrompt(t *testing.T) {
	t.Parallel()
	skill := rspSkillDir(t)
	panel := skill + "/review-panel.md"
	fx := rspNewFx(t)
	sdd := fx.canon + "/.superpowers/sdd/"

	t.Run("one slot, pass 1", func(t *testing.T) {
		r, got := fx.run(t, skill, skill, "0", fx.canon, fx.plan, "primary", "-diff", "final", "--", fx.canon)
		if r.rc != 0 {
			t.Fatalf("exit %d\n%s", r.rc, r.out)
		}
		if want := sdd + "slot-prompt-0-primary.md\n"; r.stdout != want {
			t.Errorf("stdout %q, want %q", r.stdout, want)
		}
		for _, label := range []string{"TOOLS", "FOREGROUND BUILDS", "NO DELEGATION", "REPRODUCE, DON'T READ"} {
			if b := rspTestBlock(t, panel, label); strings.Count(got, b) != 1 {
				t.Errorf("the %s block is not carried exactly once, verbatim", label)
			}
		}
		for _, want := range []string{
			"`" + fx.canon + "`; `final-review.diff`", // WORKTREES, the set filled in
			rspTestBlock(t, panel, "ENTRY CONTEXT") + "\n\n" + rbRead(sdd+"final-review.diff.touched"),
			"## PASS primary",
			"**Diff file:** " + sdd + "final-review.diff",
			"**Change artifacts:** " + fx.plan + "/proposal.md, " + fx.plan + "/design.md, " + fx.plan + "/tasks.md",
			"**Context bundle:** " + sdd + "dispatch-context.md",
			skill + "/reviewer-calibration.md",
			"`" + sdd + "panel-report-0-primary.md`",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("render lacks %q", want)
			}
		}
		for _, absent := range []string{"INDEPENDENT PASSES", "FIX-ROUND SCOPE", "CITATION CHECK", "<abs-worktree", "\n    You are"} {
			if strings.Contains(got, absent) {
				t.Errorf("render carries %q", absent)
			}
		}
	})

	t.Run("a bundle over two worktrees fills every table placeholder", func(t *testing.T) {
		writeFile(t, fx.peer+"/.superpowers/sdd/citation-check.md", "scan")
		std := fx.plan + "/standards.md"
		writeFile(t, std, "standards")
		r, got := fx.run(t, skill, skill, "0", fx.canon, fx.plan, "primary+principles+failure-modes",
			"-diff", "final", "-standard", std, "--", fx.canon, fx.peer)
		if r.rc != 0 {
			t.Fatalf("exit %d\n%s", r.rc, r.out)
		}
		if !strings.HasSuffix(r.stdout, "slot-prompt-0-primary+principles+failure-modes.md\n") {
			t.Errorf("stdout %q", r.stdout)
		}
		for _, p := range rspTablePlaceholders(t, skill) {
			if strings.Contains(got, p) {
				t.Errorf("%s left unfilled", p)
			}
		}
		citation := strings.Replace(rspTestBlock(t, panel, "CITATION CHECK"), "<abs-worktree>", fx.peer, 1)
		if strings.Count(got, "CITATION CHECK") != 1 || !strings.Contains(got, citation) {
			t.Errorf("want one CITATION CHECK, for the one worktree whose scan exists")
		}
		for _, want := range []string{
			rspTestBlock(t, panel, "INDEPENDENT PASSES"),
			"`" + fx.canon + "`, `" + fx.peer + "`; `final-review.diff`",
			sdd + "dispatch-context.md, " + fx.peer + "/.superpowers/sdd/dispatch-context.md",
			"**Principles:** " + skill + "/engineering-principles.md",
			"**Project standards:** " + std,
			"**Global constraints:** the design.md section of your context bundle",
			"panel-report-0-principles.md", "panel-report-0-failure-modes.md",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("render lacks %q", want)
			}
		}
		p, q, f := strings.Index(got, "## PASS primary"), strings.Index(got, "## PASS principles"), strings.Index(got, "## PASS failure-modes")
		if p < 0 || q < p || f < q {
			t.Errorf("PASS sections out of roster order: %d %d %d", p, q, f)
		}
	})

	for _, tc := range []struct{ kind, file string }{
		{"final", "final-review.diff"},
		{"late-fix", "late-fix.diff"},
		{"delta", "slot-delta-2-primary.diff"},
		{"fix-round", "fix-round-2.diff"},
	} {
		t.Run("-diff "+tc.kind, func(t *testing.T) {
			r, got := fx.run(t, skill, skill, "2", fx.canon, fx.plan, "primary+principles", "-diff", tc.kind, "--", fx.canon)
			if r.rc != 0 {
				t.Fatalf("exit %d\n%s", r.rc, r.out)
			}
			if !strings.Contains(got, "**Diff file:** "+sdd+tc.file+"\n") || !strings.Contains(got, tc.file+"-touched.go") {
				t.Errorf("want %s and its touched list", tc.file)
			}
			// The shared ENTRY CONTEXT and INDEPENDENT PASSES blocks name
			// final-review.diff; a pass reading another diff must not be
			// pointed back at the whole branch.
			if tc.kind != "final" && strings.Contains(got, "`final-review.diff`") {
				t.Errorf("-diff %s render still points a pass at `final-review.diff`", tc.kind)
			}
		})
	}

	t.Run("-fix-report adds FIX-ROUND SCOPE", func(t *testing.T) {
		report := sdd + "panel-fix-report-2.md"
		r, got := fx.run(t, skill, skill, "2", fx.canon, fx.plan, "primary", "-diff", "fix-round", "-fix-report", report, "--", fx.canon)
		if r.rc != 0 {
			t.Fatalf("exit %d\n%s", r.rc, r.out)
		}
		want := strings.Replace(rspTestBlock(t, skill+"/review-panel-fix-round.md", "FIX-ROUND SCOPE"), "<fix report>", report, 1)
		if !strings.Contains(got, want) {
			t.Errorf("render lacks the FIX-ROUND SCOPE block naming %s", report)
		}
	})

	// review-panel-experimental-slot.md: an exp-<name> slot is dispatched
	// like any other, its template skills/flow/experimental/<name>.md. The
	// real tree carries no candidate (kan-959 promoted test-sanity), so a
	// copy of the skill's top-level files gains one probe.
	t.Run("an exp- slot renders from experimental/<name>.md", func(t *testing.T) {
		exp := t.TempDir()
		mds, err := filepath.Glob(skill + "/*.md")
		if err != nil || len(mds) == 0 {
			t.Fatalf("glob %s/*.md: %v", skill, err)
		}
		for _, md := range mds {
			writeFile(t, exp+"/"+filepath.Base(md), rbRead(md))
		}
		mkdir(t, exp+"/experimental")
		writeFile(t, exp+"/experimental/probe.md", "description: a probe\n\n# Probe\n\n```\n    You are a probe reviewer.\n```\n")
		r, got := fx.run(t, exp, exp, "0", fx.canon, fx.plan, "primary+exp-probe", "-diff", "final", "--", fx.canon)
		if r.rc != 0 {
			t.Fatalf("exit %d\n%s", r.rc, r.out)
		}
		for _, want := range []string{
			"## PASS exp-probe\n\nYou are a probe reviewer.",
			"`" + sdd + "panel-report-0-exp-probe.md`",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("render lacks %q", want)
			}
		}
		r, _ = fx.run(t, skill, skill, "0", fx.canon, fx.plan, "exp-no-such-prompt", "-diff", "final", "--", fx.canon)
		if r.rc != 2 || !strings.Contains(r.err, "experimental/no-such-prompt.md") {
			t.Errorf("absent experimental prompt: exit %d stderr %q; want 2 naming it", r.rc, r.err)
		}
	})

	t.Run("-no-bundle", func(t *testing.T) {
		r, got := fx.run(t, skill, skill, "0", fx.canon, fx.plan, "primary+principles", "-diff", "final", "-no-bundle", "--", fx.canon)
		if r.rc != 0 {
			t.Fatalf("exit %d\n%s", r.rc, r.out)
		}
		for _, want := range []string{
			"**Context bundle:** none — the bundle was not built",
			"**Global constraints:** " + fx.plan + "/design.md",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("render lacks %q", want)
			}
		}
	})
}

func TestRenderSlotPromptUnresolvedPlaceholder(t *testing.T) {
	t.Parallel()
	real := rspSkillDir(t)
	skill := t.TempDir()
	for _, f := range []string{"review-panel.md", "review-panel-fix-round.md", "reviewer-calibration.md",
		"engineering-principles.md", "primary-reviewer-prompt.md"} {
		b := rbRead(real + "/" + f)
		if f == "primary-reviewer-prompt.md" {
			b = strings.Replace(b, "    **Diff file:** [DIFF_PATH]\n", "    **Diff file:** [DIFF_PATH]\n    **Extra:** [NEW_THING]\n", 1)
		}
		writeFile(t, skill+"/"+f, b)
	}
	fx := rspNewFx(t)
	r, _ := fx.run(t, skill, skill, "0", fx.canon, fx.plan, "primary", "-diff", "final", "--", fx.canon)
	if r.rc != 1 || !strings.Contains(r.err, "[NEW_THING]") {
		t.Errorf("exit %d stderr %q; want 1 naming [NEW_THING]", r.rc, r.err)
	}
	if _, err := os.Stat(fx.canon + "/.superpowers/sdd/slot-prompt-0-primary.md"); !os.IsNotExist(err) {
		t.Errorf("a render with an unfilled placeholder was written: %v", err)
	}

	// A mutation-alone dispatch renders the shared paragraphs and no PASS
	// section — no ENTRY CONTEXT either, mutation reading no diff file — and
	// the parent types the brief beside the rendered path.
	r, _ = fx.run(t, real, real, "0", fx.canon, fx.plan, "mutation", "-diff", "final", "--", fx.canon)
	wantMutation := fx.canon + "/.superpowers/sdd/slot-prompt-0-mutation.md"
	if r.rc != 0 || strings.TrimSpace(r.stdout) != wantMutation {
		t.Fatalf("mutation alone: exit %d stdout %q stderr %q; want 0 and %s", r.rc, r.stdout, r.err, wantMutation)
	}
	mb, err := os.ReadFile(wantMutation)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"**NO DELEGATION:**", "**WORKTREES:**"} {
		if !strings.Contains(string(mb), s) {
			t.Errorf("mutation-alone render lacks %q", s)
		}
	}
	for _, s := range []string{"## PASS", "**ENTRY CONTEXT:**"} {
		if strings.Contains(string(mb), s) {
			t.Errorf("mutation-alone render carries %q", s)
		}
	}
	// A bundle carrying mutation renders every other role and the shared
	// INDEPENDENT PASSES paragraph, with no PASS section for mutation.
	r, _ = fx.run(t, real, real, "0", fx.canon, fx.plan, "failure-modes+mutation", "-diff", "final", "--", fx.canon)
	want := fx.canon + "/.superpowers/sdd/slot-prompt-0-failure-modes+mutation.md"
	if r.rc != 0 || strings.TrimSpace(r.stdout) != want {
		t.Fatalf("failure-modes+mutation: exit %d stdout %q stderr %q; want 0 and %s", r.rc, r.stdout, r.err, want)
	}
	b, err := os.ReadFile(want)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"**INDEPENDENT PASSES:**", "## PASS failure-modes"} {
		if !strings.Contains(string(b), s) {
			t.Errorf("failure-modes+mutation render lacks %q", s)
		}
	}
	if strings.Contains(string(b), "## PASS mutation") {
		t.Errorf("failure-modes+mutation render carries a PASS mutation section")
	}
}

// TestRenderSlotPromptAlwaysOnRoles pins kan-959: the always-on extra
// dispatch's bundle renders one PASS section per role from the real
// templates, each pointed at its own report, and the old exp- id for
// test-sanity no longer resolves.
func TestRenderSlotPromptAlwaysOnRoles(t *testing.T) {
	t.Parallel()
	skill := rspSkillDir(t)
	fx := rspNewFx(t)
	sdd := fx.canon + "/.superpowers/sdd/"
	r, got := fx.run(t, skill, skill, "0", fx.canon, fx.plan, "test-sanity+call-efficiency", "-diff", "final", "--", fx.canon)
	if r.rc != 0 {
		t.Fatalf("exit %d\n%s", r.rc, r.out)
	}
	for _, want := range []string{
		"## PASS test-sanity", "## PASS call-efficiency",
		"`" + sdd + "panel-report-0-test-sanity.md`", "`" + sdd + "panel-report-0-call-efficiency.md`",
		"You are a test-sanity reviewer.", "You are a call-efficiency reviewer.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("render lacks %q", want)
		}
	}
	if r, _ := fx.run(t, skill, skill, "0", fx.canon, fx.plan, "exp-test-sanity", "-diff", "final", "--", fx.canon); r.rc != 2 {
		t.Errorf("exp-test-sanity: exit %d, want 2 — its template moved", r.rc)
	}
}

// TestRenderSlotPromptSlotFromTemplate pins the kan-959 self-review fix: a
// slot is valid when <skill-dir>/<id>-reviewer-prompt.md exists, so a slot
// added to the skill directory renders without a rebuild of the guard; an id
// with no template, or carrying a `/`, is still refused.
func TestRenderSlotPromptSlotFromTemplate(t *testing.T) {
	t.Parallel()
	real := rspSkillDir(t)
	skill := t.TempDir()
	for _, f := range []string{"review-panel.md", "reviewer-calibration.md", "engineering-principles.md"} {
		writeFile(t, skill+"/"+f, rbRead(real+"/"+f))
	}
	writeFile(t, skill+"/brand-new-reviewer-prompt.md", rbRead(real+"/primary-reviewer-prompt.md"))
	fx := rspNewFx(t)
	r, got := fx.run(t, skill, skill, "0", fx.canon, fx.plan, "brand-new", "-diff", "final", "--", fx.canon)
	if r.rc != 0 || !strings.Contains(got, "## PASS brand-new") {
		t.Fatalf("brand-new: exit %d stderr %q; want 0 and a PASS brand-new section", r.rc, r.err)
	}
	for _, slot := range []string{"no-such-slot", "../" + filepath.Base(skill) + "/brand-new"} {
		if r, _ := fx.run(t, skill, skill, "0", fx.canon, fx.plan, slot, "-diff", "final", "--", fx.canon); r.rc != 2 || !strings.Contains(r.err, "no reviewer template") {
			t.Errorf("%s: exit %d stderr %q; want 2 naming no reviewer template", slot, r.rc, r.err)
		}
	}
}
