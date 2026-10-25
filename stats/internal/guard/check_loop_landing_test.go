package guard

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// lpRowsJSON is `flow record dispatches` output for the pipeline-fix fixture
// loop: each entry "<suffix> <outcome>", an outcome "open" leaving the row
// unclosed; a leading "other:" files the row under another session token.
func lpRowsJSON(prefix, token string, rows ...string) string {
	var out []string
	for i, r := range rows {
		tok := token
		if rest, ok := strings.CutPrefix(r, "other:"); ok {
			r, tok = rest, "mf-run"
		}
		suffix, outcome, _ := strings.Cut(r, " ")
		ended := `,"endedAt":"2026-10-16T10:00:00Z"`
		if outcome == "open" {
			outcome, ended = "", ""
		}
		out = append(out, fmt.Sprintf(`{"seq":%d,"key":"%s%s","outcome":"%s","sessionToken":"%s"%s}`, i+1, prefix, suffix, outcome, tok, ended))
	}
	return "[" + strings.Join(out, ",") + "]"
}

func TestCheckLoopLanding(t *testing.T) {
	t.Parallel()
	pf := func(rows ...string) string { return lpRowsJSON("pipeline-fix-6-", lpBranch, rows...) }
	for _, tc := range []struct {
		name, branch, rows string
		rc                 int
		out                []string
	}{
		{name: "another branch", branch: "feature/x", rows: `[]`, out: []string{"LOOP-LANDING-NA"}},
		{name: "a change branch no self-review loop ran on", branch: "spectre/kan-1-demo", rows: `[]`, out: []string{"LOOP-LANDING-NA"}},
		{name: "a change branch with flowd unreadable", branch: "spectre/kan-1-demo", rows: "", out: []string{"LOOP-LANDING-NA"}},
		{name: "a change branch whose self-review loop is unreviewed", branch: "spectre/kan-1-demo",
			rows: lpRowsJSON("self-review-kan-1-demo-", "self-review-kan-1-demo", "plan completed", "fix completed"), rc: 1, out: []string{"REVIEW — no self-review-kan-1-demo-review-<r>"}},
		{name: "a change branch whose self-review loop is clean", branch: "spectre/kan-1-demo",
			rows: lpRowsJSON("self-review-kan-1-demo-", "self-review-kan-1-demo", "plan completed", "fix completed", "review-1 clean"), out: []string{"LOOP-LANDING-OK"}},
		{name: "a review key with no numeric round", branch: lpBranch, rows: pf("plan completed", "fix completed", "review-x clean", "review-1 clean"), rc: 1,
			out: []string{"REVIEW — pipeline-fix-6-review-x carries no numeric round"}},
		{name: "a reviewed loop", branch: lpBranch, rows: pf("plan completed", "fix completed", "review-1 fix", "fix-1 completed", "review-2 clean"),
			out: []string{"LOOP-LANDING-OK"}},
		{name: "a Minor-only review closed by its fix", branch: lpBranch, rows: pf("plan completed", "fix-g1 completed", "fix-g2 completed", "review-1 clean", "fix-1 completed"),
			out: []string{"LOOP-LANDING-OK"}},
		{name: "fixes landed with no review", branch: lpBranch, rows: pf("plan completed", "fix completed"), rc: 1,
			out: []string{"REVIEW — no pipeline-fix-6-review-<r>"}},
		{name: "a fixer with no planner", branch: lpBranch, rows: pf("fix completed", "review-1 clean"), rc: 1,
			out: []string{"PLAN — no completed pipeline-fix-6-plan"}},
		{name: "a planner with no fixer", branch: lpBranch, rows: pf("plan completed", "review-1 clean"), rc: 1,
			out: []string{"FIX — no completed pipeline-fix-6-fix"}},
		{name: "the last review raised an Important", branch: lpBranch, rows: pf("plan completed", "fix completed", "review-1 clean", "fix-1 completed", "review-2 fix"), rc: 1,
			out: []string{`REVIEW — pipeline-fix-6-review-2 ended "fix"`}},
		{name: "a red-line fix after the last review", branch: lpBranch, rows: pf("plan completed", "fix completed", "review-1 clean", "fix-v1 completed"), rc: 1,
			out: []string{"REVIEW — a pipeline-fix-6-fix-v<v> dispatch came after pipeline-fix-6-review-1"}},
		{name: "a row left open", branch: lpBranch, rows: pf("plan completed", "fix completed", "review-1 open"), rc: 1,
			out: []string{"OPEN — pipeline-fix-6-review-1", `ended ""`}},
		{name: "the run's own rows do not count", branch: lpBranch, rows: pf("other:plan completed", "other:fix completed", "other:review-1 clean"), rc: 1,
			out: []string{"PLAN —", "FIX —", "REVIEW —"}},
		{name: "a self-review loop", branch: "self-review-kan-1-demo",
			rows: lpRowsJSON("self-review-kan-1-demo-", "self-review-kan-1-demo", "plan completed", "fix completed", "review-1 clean"), out: []string{"LOOP-LANDING-OK"}},
		{name: "flowd unreadable", branch: lpBranch, rows: "", rc: 2, out: []string{"cannot answer"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := lpNew(t, tc.rows)
			var o, e bytes.Buffer
			rc := checkLoopLanding([]string{fx.repo, tc.branch}, fx.env, &o, &e)
			out := o.String() + e.String()
			if rc != tc.rc {
				t.Fatalf("rc %d, want %d: %s", rc, tc.rc, out)
			}
			for _, w := range tc.out {
				if !strings.Contains(out, w) {
					t.Fatalf("output %q carries no %q", out, w)
				}
			}
			if calls := fx.calls(t); tc.rows != `[]` && tc.rows != "" && !strings.HasPrefix(tc.branch, "spectre/") && tc.branch != "feature/x" && !strings.Contains(calls, "-change "+tc.branch) {
				t.Fatalf("flow calls %q never read the loop's change", calls)
			}
		})
	}
}

// A satellite -- the loop's branch in another repository it carried a task
// into -- lands on the rows the loop filed under the agents checkout's
// project (FLOW_GUARD_REPO_ROOT) when its own repository's project has none;
// a loop with rows nowhere is still refused, and a change's own branch never
// reads another project's rows.
func TestCheckLoopLandingSatellite(t *testing.T) {
	t.Parallel()
	const sr = "self-review-kan-1-demo"
	clean := lpRowsJSON(sr+"-", sr, "plan completed", "fix completed", "review-1 clean")
	for _, tc := range []struct {
		name, branch, agentsRows string
		rc                       int
		out                      string
	}{
		{name: "rows under the agents project", branch: sr, agentsRows: clean, out: "LOOP-LANDING-OK"},
		{name: "rows nowhere", branch: sr, agentsRows: `[]`, rc: 1, out: "PLAN — no completed " + sr + "-plan"},
		{name: "a change branch reads only its own project", branch: "spectre/kan-1-demo", agentsRows: clean, out: "LOOP-LANDING-NA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := lpNew(t, `[]`)
			sat := filepath.Dir(fx.repo) + "/gymie"
			gitRun(t, "", "init", "-q", "-b", "main", sat)
			gitRun(t, sat, "commit", "-q", "--allow-empty", "-m", "init")
			writeFile(t, fx.rows+".agents", tc.agentsRows)
			bin := filepath.Dir(fx.repo) + "/bin"
			writeExec(t, bin+"/flow", `#!/usr/bin/env bash
case "$*" in
  *"-C `+fx.repo+`"*) exec cat `+fx.rows+`.agents ;;
  *) exec cat `+fx.rows+` ;;
esac
`)
			env := fx.env
			env.Getenv = func(k string) string {
				if k == "FLOW_GUARD_REPO_ROOT" {
					return fx.repo
				}
				return fx.env.Getenv(k)
			}
			var o, e bytes.Buffer
			rc := checkLoopLanding([]string{sat, tc.branch}, env, &o, &e)
			if out := o.String() + e.String(); rc != tc.rc || !strings.Contains(out, tc.out) {
				t.Fatalf("rc %d, want %d; output %q, want it to carry %q", rc, tc.rc, out, tc.out)
			}
		})
	}
}

// A self-review-* branch whose only commits add prompt-tuning cases lands
// with no flowd rows; one more commit outside docs/prompt-tuning/ puts it
// back under the loop's row checks.
func TestLoopLandingAdmitsPromptTuningOnly(t *testing.T) {
	t.Parallel()
	const sr = "self-review-x"
	fx := lpNew(t, `[]`)
	gitRun(t, fx.repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	gitRun(t, fx.repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	gitRun(t, fx.repo, "checkout", "-q", "-b", sr)
	writeFile(t, fx.repo+"/docs/prompt-tuning/x-1/case.json", "{}\n")
	gitRun(t, fx.repo, "add", "-A")
	gitRun(t, fx.repo, "commit", "-q", "-m", "docs(prompt-tuning): x cases")

	var o, e bytes.Buffer
	if rc := checkLoopLanding([]string{fx.repo, sr}, fx.env, &o, &e); rc != 0 || !strings.Contains(o.String(), "LOOP-LANDING-OK "+sr+" (prompt-tuning cases only)") {
		t.Fatalf("cases-only branch: rc %d, output %q", rc, o.String()+e.String())
	}

	writeFile(t, fx.repo+"/skills/flow/implement.md", "edit\n")
	gitRun(t, fx.repo, "add", "-A")
	gitRun(t, fx.repo, "commit", "-q", "-m", "fix: an edit")
	o.Reset()
	e.Reset()
	if rc := checkLoopLanding([]string{fx.repo, sr}, fx.env, &o, &e); rc != 1 || !strings.Contains(o.String(), "LOOP-LANDING-REFUSED: PLAN") {
		t.Fatalf("mixed branch: rc %d, output %q", rc, o.String()+e.String())
	}
}

// The admission needs a non-empty diff and a self-review-* branch: an empty
// self-review branch and a pipeline-fix branch carrying only cases both stay
// under the row checks.
func TestLoopLandingPromptTuningAdmissionBounds(t *testing.T) {
	t.Parallel()
	fx := lpNew(t, `[]`)
	gitRun(t, fx.repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	gitRun(t, fx.repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	gitRun(t, fx.repo, "branch", "self-review-empty")
	gitRun(t, fx.repo, "checkout", "-q", "-b", lpBranch)
	writeFile(t, fx.repo+"/docs/prompt-tuning/x-1/case.json", "{}\n")
	gitRun(t, fx.repo, "add", "-A")
	gitRun(t, fx.repo, "commit", "-q", "-m", "docs(prompt-tuning): x cases")
	for _, b := range []string{"self-review-empty", lpBranch} {
		var o, e bytes.Buffer
		if rc := checkLoopLanding([]string{fx.repo, b}, fx.env, &o, &e); rc != 1 || strings.Contains(o.String(), "prompt-tuning cases only") {
			t.Fatalf("%s: rc %d, output %q", b, rc, o.String()+e.String())
		}
	}
}
