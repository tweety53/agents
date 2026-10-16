package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lpFx is one loop fixture: an agents repo carrying the real pairs table, a
// `flow` stub on PATH that answers `record dispatches` from rows and logs
// every call, and the guard's Env.
type lpFx struct {
	repo, log, rows string
	env             Env
}

// lpNew builds the fixture with the given branches; rows is the JSON
// `flow record dispatches` prints ("" makes it fail).
func lpNew(t *testing.T, rows string, branches ...string) lpFx {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fx := lpFx{repo: tmp + "/agents", log: tmp + "/flow.log", rows: tmp + "/rows.json"}
	skill, err := os.ReadFile("../../../" + lpSkill)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, fx.repo+"/"+lpSkill, string(skill))
	gitRun(t, "", "init", "-q", "-b", "main", fx.repo)
	gitRun(t, fx.repo, "add", "-A")
	gitRun(t, fx.repo, "commit", "-q", "-m", "init")
	for _, b := range branches {
		gitRun(t, fx.repo, "branch", b)
	}
	if rows != "" {
		writeFile(t, fx.rows, rows)
	}
	bin := tmp + "/bin"
	writeExec(t, bin+"/flow", `#!/usr/bin/env bash
printf '%s\n' "$*" >>`+fx.log+`
case "$1 $2" in
  "record dispatches") [ -f `+fx.rows+` ] && exec cat `+fx.rows+`; exit 1 ;;
  "record dispatch") echo "recorded: dispatch 1" ;;
esac
`)
	fx.env = Env{Getenv: pathEnv(bin), Dir: tmp}
	return fx
}

func (fx lpFx) run(t *testing.T, args ...string) (int, string) {
	t.Helper()
	var o, e bytes.Buffer
	rc := checkLoopDispatch(append([]string{fx.repo}, args...), fx.env, &o, &e)
	return rc, o.String() + e.String()
}

func (fx lpFx) calls(t *testing.T) string {
	t.Helper()
	b, _ := os.ReadFile(fx.log)
	return string(b)
}

const (
	lpBranch   = "pipeline-fix-kan-1-demo-6"
	lpPlanDone = `[{"seq":1,"key":"pipeline-fix-6-plan","outcome":"completed","sessionToken":"pipeline-fix-kan-1-demo-6","endedAt":"2026-10-16T10:00:00Z"}]`
	lpFixRound = `[{"seq":1,"key":"pipeline-fix-6-plan","outcome":"completed","sessionToken":"pipeline-fix-kan-1-demo-6","endedAt":"2026-10-16T10:00:00Z"},` +
		`{"seq":2,"key":"pipeline-fix-6-fix","outcome":"completed","sessionToken":"pipeline-fix-kan-1-demo-6","endedAt":"2026-10-16T10:00:00Z"},` +
		`{"seq":3,"key":"pipeline-fix-6-review-1","outcome":"fix","sessionToken":"pipeline-fix-kan-1-demo-6","endedAt":"2026-10-16T10:00:00Z"}]`
	lpDesignLow = "## Decisions\n\n**ID:** first-review-effort\n\n**Chosen:** the first review runs on `flow-low` (effort `low`)\n"
	lpPlanOpen  = `[{"seq":1,"key":"pipeline-fix-6-plan","sessionToken":"pipeline-fix-kan-1-demo-6"}]`
)

func TestLoopTableIsTheSkillsOwn(t *testing.T) {
	t.Parallel()
	rows, err := lpTable("../../..")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"pass": "flow-medium opus", "plan-2": "flow-medium opus", "fix-g1": "flow-low opus",
		"fix-v3": "flow-low opus", "review-1": "flow-medium flow-low opus", "review-2": "flow-low sonnet"}
	for suffix, pair := range want {
		got := ""
		for _, r := range rows {
			if r.suffix.MatchString(suffix) {
				got = strings.Join(append(append([]string{}, r.types...), r.model), " ")
				break
			}
		}
		if got != pair {
			t.Errorf("suffix %s: pair %q, want %q", suffix, got, pair)
		}
	}
}

func TestCheckLoopDispatch(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, rows, key, typ, model string
		branches, escalated         []string
		design                      string // the loop branch's design.md, committed when set
		rc                          int
		out, begin                  string // begin "" asserts no begin call
	}{
		{name: "plan admitted and recorded", rows: `[]`, key: "pipeline-fix-6-plan", typ: "flow-medium", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-change " + lpBranch + " -role planner -model opus -effort medium -key pipeline-fix-6-plan -session-token " + lpBranch},
		{name: "fix without a plan", rows: `[]`, key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "PLAN —"},
		{name: "fix after an open plan", rows: lpPlanOpen, key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "OPEN — pipeline-fix-6-plan"},
		{name: "fix after a completed plan", rows: lpPlanDone, key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-role implementer -model opus -effort low -key pipeline-fix-6-fix"},
		{name: "fix on flow-high unescalated", rows: lpPlanDone, key: "pipeline-fix-6-fix", typ: "flow-high", model: "opus", branches: []string{lpBranch}, rc: 1, out: "ESCALATION —"},
		{name: "fix on flow-high escalated", rows: lpFixRound, key: "pipeline-fix-6-fix-1", typ: "flow-high", model: "opus", branches: []string{lpBranch},
			escalated: []string{lpBranch + ":pipeline-fix-6-fix-1"}, out: "LOOP-DISPATCH-OK", begin: "-effort high"},
		{name: "a hand-set marker with no earlier round", rows: lpPlanDone, key: "pipeline-fix-6-fix", typ: "flow-high", model: "opus", branches: []string{lpBranch},
			escalated: []string{lpBranch + ":pipeline-fix-6-fix"}, rc: 1, out: "ESCALATION —"},
		{name: "another loop's marker for the same key", rows: lpFixRound, key: "pipeline-fix-6-fix-1", typ: "flow-high", model: "opus", branches: []string{lpBranch},
			escalated: []string{"pipeline-fix-other-6:pipeline-fix-6-fix-1", "pipeline-fix-6-fix-1"}, rc: 1, out: "ESCALATION —"},
		{name: "an aborted key is begun again", rows: lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix aborted"), key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-key pipeline-fix-6-fix"},
		{name: "fix on sonnet", rows: lpPlanDone, key: "pipeline-fix-6-fix", typ: "flow-low", model: "sonnet", branches: []string{lpBranch}, rc: 1, out: "PAIR —"},
		{name: "re-review on opus", rows: lpPlanDone, key: "pipeline-fix-6-review-2", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "PAIR —"},
		{name: "first review on flow-low as the planner recorded", rows: lpPlanDone, key: "pipeline-fix-6-review-1", typ: "flow-low", model: "opus", branches: []string{lpBranch},
			design: lpDesignLow, out: "LOOP-DISPATCH-OK", begin: "-role reviewer"},
		{name: "first review on flow-medium against a recorded flow-low", rows: lpPlanDone, key: "pipeline-fix-6-review-1", typ: "flow-medium", model: "opus", branches: []string{lpBranch},
			design: lpDesignLow, rc: 1, out: "first-review-effort"},
		{name: "first review on flow-low with none recorded", rows: lpPlanDone, key: "pipeline-fix-6-review-1", typ: "flow-low", model: "opus", branches: []string{lpBranch},
			rc: 1, out: "first-review-effort"},
		{name: "first review on flow-medium with none recorded", rows: lpPlanDone, key: "pipeline-fix-6-review-1", typ: "flow-medium", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-role reviewer"},
		{name: "a reused key", rows: lpPlanDone, key: "pipeline-fix-6-plan", typ: "flow-medium", model: "opus", branches: []string{lpBranch}, rc: 1, out: "KEY-REUSED —"},
		{name: "no loop branch", rows: `[]`, key: "pipeline-fix-6-plan", typ: "flow-medium", model: "opus", rc: 1, out: "NOT-STARTED —"},
		{name: "an unknown suffix", rows: `[]`, key: "pipeline-fix-6-implementer", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "KEY —"},
		{name: "self-review on its own branch files under the change", rows: `[]`, key: "self-review-kan-1-demo-plan", typ: "flow-medium", model: "opus",
			out: "LOOP-DISPATCH-OK", begin: "-change kan-1-demo -role planner"},
		{name: "self-review loop files under its branch", rows: `[]`, key: "self-review-kan-1-demo-plan", typ: "flow-medium", model: "opus", branches: []string{"self-review-kan-1-demo"},
			out: "LOOP-DISPATCH-OK", begin: "-change self-review-kan-1-demo -role planner -model opus -effort medium -key self-review-kan-1-demo-plan -session-token self-review-kan-1-demo"},
		{name: "the pass is checked, never recorded", rows: `[]`, key: "self-review-kan-1-demo-pass", typ: "flow-medium", model: "opus", out: "LOOP-DISPATCH-OK"},
		{name: "the pass off its pair", rows: `[]`, key: "self-review-kan-1-demo-pass", typ: "flow-low", model: "opus", rc: 1, out: "PAIR —"},
		{name: "flowd unreadable", rows: "", key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 2, out: "cannot answer"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := lpNew(t, tc.rows, tc.branches...)
			if tc.design != "" {
				gitRun(t, fx.repo, "checkout", "-q", lpBranch)
				writeFile(t, fx.repo+"/spectre/changes/"+lpBranch+"/design.md", tc.design)
				gitRun(t, fx.repo, "add", "-A")
				gitRun(t, fx.repo, "commit", "-q", "-m", "plan")
				gitRun(t, fx.repo, "checkout", "-q", "main")
			}
			for _, k := range tc.escalated {
				gitRun(t, fx.repo, "config", "--add", "flow.escalation", k)
			}
			rc, out := fx.run(t, tc.key, tc.typ, tc.model)
			if rc != tc.rc || !strings.Contains(out, tc.out) {
				t.Fatalf("rc %d, want %d; output %q, want it to carry %q", rc, tc.rc, out, tc.out)
			}
			calls := fx.calls(t)
			if tc.begin == "" && strings.Contains(calls, "dispatch begin") {
				t.Fatalf("a begin was recorded:\n%s", calls)
			}
			if tc.begin != "" && !strings.Contains(calls, tc.begin) {
				t.Fatalf("flow calls %q carry no %q", calls, tc.begin)
			}
		})
	}
}

func TestCheckLoopDispatchUnreadableTable(t *testing.T) {
	t.Parallel()
	fx := lpNew(t, `[]`, lpBranch)
	if err := os.Remove(fx.repo + "/" + lpSkill); err != nil {
		t.Fatal(err)
	}
	if rc, out := fx.run(t, "pipeline-fix-6-plan", "flow-medium", "opus"); rc != 2 {
		t.Fatalf("rc %d, want 2: %s", rc, out)
	}
}

// A self-review on a change's own branch in another project -- no
// self-review-<name> branch in <agents repo> -- files its rows in that
// project's flowd project: flow runs with -C the caller's main checkout, and
// review-1's first-review-effort is read from that project's spectre/<name>.
// A self-review-<name> loop branch in <agents repo> still files there.
func TestCheckLoopDispatchOwnBranchInAnotherProject(t *testing.T) {
	t.Parallel()
	fx := lpNew(t, lpRowsJSON("self-review-kan-1-demo-", "self-review-kan-1-demo", "plan completed"))
	proj := filepath.Dir(fx.repo) + "/gymie"
	writeFile(t, proj+"/README.md", "gymie\n")
	gitRun(t, "", "init", "-q", "-b", "develop", proj)
	gitRun(t, proj, "add", "-A")
	gitRun(t, proj, "commit", "-q", "-m", "init")
	gitRun(t, proj, "checkout", "-q", "-b", "spectre/kan-1-demo")
	writeFile(t, proj+"/spectre/changes/kan-1-demo/design.md", lpDesignLow)
	gitRun(t, proj, "add", "-A")
	gitRun(t, proj, "commit", "-q", "-m", "plan")
	gitRun(t, proj, "checkout", "-q", "develop")
	fx.env.Dir = proj
	if rc, out := fx.run(t, "self-review-kan-1-demo-review-1", "flow-low", "opus"); rc != 0 {
		t.Fatalf("rc %d: %s", rc, out)
	}
	if calls := fx.calls(t); !strings.Contains(calls, "record dispatch begin -C "+proj+" -change kan-1-demo -role reviewer") ||
		!strings.Contains(calls, "record dispatches -change kan-1-demo -C "+proj) {
		t.Fatalf("the own-branch loop was not filed in the caller's project:\n%s", calls)
	}
	gitRun(t, fx.repo, "branch", "self-review-kan-1-demo")
	if rc, out := fx.run(t, "self-review-kan-1-demo-fix", "flow-low", "opus"); rc != 0 ||
		!strings.Contains(fx.calls(t), "record dispatch begin -C "+fx.repo+" -change self-review-kan-1-demo") {
		t.Fatalf("a self-review loop branch left <agents repo>: rc %d: %s\n%s", rc, out, fx.calls(t))
	}
}
