package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// lpFx is one loop fixture: an agents repo carrying the real model-policy.md
// (**Allowed pairs**), a `flow` stub on PATH that answers `record
// dispatches` from rows and logs every call, and the guard's Env.
type lpFx struct {
	repo, log, rows string
	prompt          string // the prompt the hook hands in on stdin
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
	table, err := os.ReadFile("../../../" + mpFile)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, fx.repo+"/"+mpFile, string(table))
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
	writeExec(t, bin+"/flow", "#!/usr/bin/env bash\n"+policyStub+`printf '%s\n' "$*" >>`+fx.log+`
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
	env := fx.env
	env.Stdin = strings.NewReader(fx.prompt)
	rc := checkLoopDispatch(append([]string{fx.repo}, args...), env, &o, &e)
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

// TestLoopSuffixRoles pins the key suffix -> role id map onto the real table:
// every suffix lands on a role the allowed-pairs table carries.
func TestLoopSuffixRoles(t *testing.T) {
	t.Parallel()
	rows, err := policyTable(noFlowEnv(t))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"pass": "-pass", "plan": "-plan", "plan-2": "-plan", "fix": "-fix", "fix-3": "-fix", "fix-g1": "-fix",
		"fix-v3": "-fix", "review-1": "-review-1", "review-2": "-review-r"}
	for _, prefix := range []string{"self-review", "pipeline-fix"} {
		for suffix, role := range want {
			got := lpRole(prefix, suffix)
			if got != prefix+role {
				t.Errorf("%s %s: role %q, want %q", prefix, suffix, got, prefix+role)
			}
			if prefix+role != "pipeline-fix-pass" && len(rows[got]) == 0 {
				t.Errorf("role %q has no row of the allowed-pairs table", got)
			}
		}
	}
	for _, bad := range []string{"fix-x", "review", "plan-a", "pass-1"} {
		if got := lpRole("self-review", bad); got != "" {
			t.Errorf("suffix %q: role %q, want none", bad, got)
		}
	}
}

// TestLoopDispatchReadsPolicyTable: the pair comes from the role's pairs in
// flowd (`flow policy get`), never a table of the loop's own.
func TestLoopDispatchReadsPolicyTable(t *testing.T) {
	t.Parallel()
	rows := lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 fix")
	fx := lpNew(t, rows, lpBranch)
	policy := strings.Replace(testPolicy, `"pipeline-fix-review-r":{"base":["sonnet/flow-low"]`, `"pipeline-fix-review-r":{"base":["opus/flow-low"]`, 1)
	if policy == testPolicy {
		t.Fatal("the seeded pairs carry no pipeline-fix-review-r base to rewrite")
	}
	// A `flow` ahead of the fixture's own answers `policy get` with the
	// rewritten pairs and hands every other call to the fixture's.
	wrap := t.TempDir()
	writeExec(t, wrap+"/flow", "#!/bin/sh\n[ \"$1 $2\" = \"policy get\" ] && { printf '%s\\n' '"+policy+"'; exit 0; }\nexec "+filepath.Dir(fx.repo)+"/bin/flow \"$@\"\n")
	inner := fx.env.Getenv
	fx.env.Getenv = func(k string) string {
		if k == "PATH" {
			return wrap + ":" + inner(k)
		}
		return inner(k)
	}
	if rc, out := fx.run(t, "pipeline-fix-6-review-2", "flow-low", "sonnet"); rc != 1 || !strings.Contains(out, "PAIR —") {
		t.Fatalf("sonnet/flow-low off the fixture row: rc %d: %s", rc, out)
	}
	if rc, out := fx.run(t, "pipeline-fix-6-review-2", "flow-low", "opus"); rc != 0 {
		t.Fatalf("opus/flow-low on the fixture row: rc %d: %s", rc, out)
	}
}

// TestLoopDispatchFirstReviewOffPolicy: a first-review-effort the role's
// allowed pairs do not admit never refuses the pair they do (smoke-test-set,
// pipeline-fix 67).
func TestLoopDispatchFirstReviewOffPolicy(t *testing.T) {
	t.Parallel()
	fx := lpNew(t, lpPlanDone, lpBranch)
	policy := strings.Replace(testPolicy, `"pipeline-fix-review-1":{"base":["opus/flow-medium","opus/flow-low"]`, `"pipeline-fix-review-1":{"base":["opus/flow-low"]`, 1)
	if policy == testPolicy {
		t.Fatal("the seeded pairs carry no pipeline-fix-review-1 base to rewrite")
	}
	wrap := t.TempDir()
	writeExec(t, wrap+"/flow", "#!/bin/sh\n[ \"$1 $2\" = \"policy get\" ] && { printf '%s\\n' '"+policy+"'; exit 0; }\nexec "+filepath.Dir(fx.repo)+"/bin/flow \"$@\"\n")
	inner := fx.env.Getenv
	fx.env.Getenv = func(k string) string {
		if k == "PATH" {
			return wrap + ":" + inner(k)
		}
		return inner(k)
	}
	writeFile(t, fx.repo+"-worktrees/"+lpBranch+"/spectre/changes/"+lpBranch+"/design.md", strings.Replace(lpDesignLow, "flow-low", "flow-medium", 1))
	if rc, out := fx.run(t, "pipeline-fix-6-review-1", "flow-low", "opus"); rc != 0 {
		t.Fatalf("the policy's only pair against a recorded flow-medium: rc %d: %s", rc, out)
	}
}

func TestCheckLoopDispatch(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, tc := range []struct {
		name, rows, key, typ, model string
		prompt                      string // stdin; the escalated fix carries its marker
		branches, escalated         []string
		design                      string // the loop worktree's design.md, written uncommitted when set
		rc                          int
		out, begin                  string // begin "" asserts no begin call
	}{
		{name: "plan admitted and recorded", rows: `[]`, key: "pipeline-fix-6-plan", typ: "flow-planner-medium", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-change " + lpBranch + " -role planner -model opus -effort medium -key pipeline-fix-6-plan -session-token " + lpBranch},
		{name: "fix without a plan", rows: `[]`, key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "PLAN —"},
		{name: "fix after an open plan", rows: lpPlanOpen, key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "OPEN — pipeline-fix-6-plan"},
		{name: "fix after a completed plan", rows: lpPlanDone, key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-role implementer -model opus -effort low -key pipeline-fix-6-fix"},
		{name: "fix on flow-high unescalated", rows: lpPlanDone, key: "pipeline-fix-6-fix", typ: "flow-high", model: "opus", branches: []string{lpBranch}, rc: 1, out: "ESCALATION —"},
		{name: "fix on flow-high escalated", rows: lpFixRound, key: "pipeline-fix-6-fix-1", typ: "flow-high", model: "opus", branches: []string{lpBranch},
			escalated: []string{lpBranch + ":pipeline-fix-6-fix-1"}, prompt: "No-progress escalation\n", out: "LOOP-DISPATCH-OK", begin: "-effort high"},
		{name: "fix on flow-high escalated without its marker", rows: lpFixRound, key: "pipeline-fix-6-fix-1", typ: "flow-high", model: "opus", branches: []string{lpBranch},
			escalated: []string{lpBranch + ":pipeline-fix-6-fix-1"}, rc: 1, out: "PAIR —"},
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
		{name: "a reused key", rows: lpPlanDone, key: "pipeline-fix-6-plan", typ: "flow-planner-medium", model: "opus", branches: []string{lpBranch}, rc: 1, out: "KEY-REUSED —"},
		{name: "no loop branch", rows: `[]`, key: "pipeline-fix-6-plan", typ: "flow-planner-medium", model: "opus", rc: 1, out: "NOT-STARTED —"},
		{name: "an unknown suffix", rows: `[]`, key: "pipeline-fix-6-implementer", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 1, out: "KEY —"},
		{name: "self-review on its own branch files under the change", rows: `[]`, key: "self-review-kan-1-demo-plan", typ: "flow-planner-medium", model: "opus",
			out: "LOOP-DISPATCH-OK", begin: "-change kan-1-demo -role planner"},
		{name: "self-review loop files under its branch", rows: `[]`, key: "self-review-kan-1-demo-plan", typ: "flow-planner-medium", model: "opus", branches: []string{"self-review-kan-1-demo"},
			out: "LOOP-DISPATCH-OK", begin: "-change self-review-kan-1-demo -role planner -model opus -effort medium -key self-review-kan-1-demo-plan -session-token self-review-kan-1-demo"},
		{name: "the pass is recorded under the change itself", rows: `[]`, key: "self-review-kan-1-demo-pass", typ: "flow-medium", model: "opus",
			out: "LOOP-DISPATCH-OK", begin: "-change kan-1-demo -role reviewer -model opus -effort medium -key self-review-kan-1-demo-pass -session-token self-review-kan-1-demo"},
		{name: "the pass stays under the change while its loop branch exists", rows: `[]`, key: "self-review-kan-1-demo-pass", typ: "flow-medium", model: "opus", branches: []string{"self-review-kan-1-demo"},
			out: "LOOP-DISPATCH-OK", begin: "-change kan-1-demo -role reviewer"},
		{name: "the pass off its pair", rows: `[]`, key: "self-review-kan-1-demo-pass", typ: "flow-low", model: "opus", rc: 1, out: "PAIR —"},
		{name: "flowd unreadable", rows: "", key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 2, out: "cannot answer"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := lpNew(t, tc.rows, tc.branches...)
			if tc.design != "" {
				writeFile(t, fx.repo+"-worktrees/"+lpBranch+"/spectre/changes/"+lpBranch+"/design.md", tc.design)
			}
			for _, k := range tc.escalated {
				gitRun(t, fx.repo, "config", "--add", "flow.escalation", k)
			}
			fx.prompt = tc.prompt
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

// TestLoopDispatchPlannerTypes pins the plan row's planner definitions
// (kan-1033, decision planner-carries-agent): flow-planner-medium and
// flow-planner-low are admitted, a leaf flow-medium is refused, and
// flow-planner-high is refused even on a recorded escalation.
func TestLoopDispatchPlannerTypes(t *testing.T) {
	t.Parallel()
	const plan2 = `[{"seq":1,"key":"self-review-x-plan","outcome":"completed","sessionToken":"self-review-x","endedAt":"2026-10-16T10:00:00Z"}]`
	for _, tc := range []struct {
		name, rows, key, typ string
		escalated            bool
		rc                   int
		out                  string
	}{
		{"medium planner", `[]`, "self-review-x-plan", "flow-planner-medium", false, 0, "-effort medium"},
		{"low planner", `[]`, "self-review-x-plan", "flow-planner-low", false, 0, "-effort low"},
		{"leaf flow-medium", `[]`, "self-review-x-plan", "flow-medium", false, 1, "PAIR — self-review-plan runs on opus/flow-planner-medium, opus/flow-planner-low"},
		{"high planner unescalated", plan2, "self-review-x-plan-2", "flow-planner-high", false, 1, "PAIR —"},
		{"high planner escalated", plan2, "self-review-x-plan-2", "flow-planner-high", true, 1, "PAIR —"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := lpNew(t, tc.rows, "self-review-x")
			if tc.escalated {
				gitRun(t, fx.repo, "config", "--add", "flow.escalation", "self-review-x:"+tc.key)
			}
			rc, out := fx.run(t, tc.key, tc.typ, "opus")
			if rc != tc.rc || !strings.Contains(out+fx.calls(t), tc.out) {
				t.Fatalf("rc %d, want %d; output %q, calls %q, want %q", rc, tc.rc, out, fx.calls(t), tc.out)
			}
		})
	}
}

func TestCheckLoopDispatchUnreadableTable(t *testing.T) {
	t.Parallel()
	fx := lpNew(t, `[]`, lpBranch)
	bin := t.TempDir()
	writeExec(t, bin+"/flow", "#!/bin/sh\nexit 1\n")
	fx.env.Getenv = pathEnv(bin)
	if rc, out := fx.run(t, "pipeline-fix-6-plan", "flow-medium", "opus"); rc != 2 {
		t.Fatalf("rc %d, want 2: %s", rc, out)
	}
}

// A self-review on a change's own branch in another project -- no
// self-review-<name> branch in <agents repo> -- files its rows in that
// project's flowd project: flow runs with -C the caller's main checkout, and
// review-1's first-review-effort is read from the loop's own plan,
// spectre/changes/self-review-<name>/, in that project's <name> worktree.
// A self-review-<name> loop branch in <agents repo> still files there.
func TestCheckLoopDispatchOwnBranchInAnotherProject(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := lpNew(t, lpRowsJSON("self-review-kan-1-demo-", "self-review-kan-1-demo", "plan completed"))
	proj := filepath.Dir(fx.repo) + "/gymie"
	writeFile(t, proj+"/README.md", "gymie\n")
	gitRun(t, "", "init", "-q", "-b", "develop", proj)
	gitRun(t, proj, "add", "-A")
	gitRun(t, proj, "commit", "-q", "-m", "init")
	gitRun(t, proj, "branch", "spectre/kan-1-demo")
	writeFile(t, proj+"-worktrees/kan-1-demo/spectre/changes/self-review-kan-1-demo/design.md", lpDesignLow)
	fx.env.Dir = proj
	if rc, out := fx.run(t, "self-review-kan-1-demo-review-1", "flow-low", "opus"); rc != 0 {
		t.Fatalf("rc %d: %s", rc, out)
	}
	if calls := fx.calls(t); !strings.Contains(calls, "record dispatch begin -C "+proj+" -change kan-1-demo -role reviewer") ||
		!strings.Contains(calls, "record dispatches -change kan-1-demo -C "+proj) {
		t.Fatalf("the own-branch loop was not filed in the caller's project:\n%s", calls)
	}
	gitRun(t, fx.repo, "branch", "self-review-kan-1-demo")
	if rc, out := fx.run(t, "self-review-kan-1-demo-fix", "flow-medium", "opus"); rc != 0 ||
		!strings.Contains(fx.calls(t), "record dispatch begin -C "+fx.repo+" -change self-review-kan-1-demo") {
		t.Fatalf("a self-review loop branch left <agents repo>: rc %d: %s\n%s", rc, out, fx.calls(t))
	}
}

// A self-review on a /flow-fast change's own branch, <name> with no spectre/
// prefix, reads review-1's first-review-effort from the loop's plan there.
func TestCheckLoopDispatchOwnFastBranchReadsLoopPlan(t *testing.T) {
	t.Parallel()
	fx := lpNew(t, lpRowsJSON("self-review-kan-1-demo-", "self-review-kan-1-demo", "plan completed", "fix completed"))
	gitRun(t, fx.repo, "branch", "kan-1-demo")
	writeFile(t, fx.repo+"-worktrees/kan-1-demo/spectre/changes/self-review-kan-1-demo/design.md", lpDesignLow)
	if rc, out := fx.run(t, "self-review-kan-1-demo-review-1", "flow-low", "opus"); rc != 0 {
		t.Fatalf("rc %d: %s", rc, out)
	}
}

// review-1's first-review-effort reads the effort as a planner writes it --
// bare or backticked low/medium as the skill names it, or flow-low/flow-medium
// -- from that decision's own entry, and refuses an entry naming none rather
// than falling back to flow-medium.
func TestCheckLoopDispatchFirstReviewEffort(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	const head = "## Decisions\n\n### First review\n\n**ID:** first-review-effort\n**Status:** active\n"
	for _, tc := range []struct {
		name, chosen, typ string
		rc                int
		out               string
		superseded        bool
		status            string
	}{
		{name: "bare low", chosen: "**Chosen:** low — dev-only tooling, no hard seam.\n**Considered:** medium — no seam here.\n", typ: "flow-low", out: "LOOP-DISPATCH-OK"},
		{name: "bare flow-medium", chosen: "**Chosen:** flow-medium — eleven behaviour changes.\n", typ: "flow-medium", out: "LOOP-DISPATCH-OK"},
		{name: "bare low refuses flow-medium", chosen: "**Chosen:** low — dev-only tooling.\n", typ: "flow-medium", rc: 1, out: "runs on flow-low"},
		{name: "backticked low in prose", chosen: "**Chosen:** the first review runs at effort `low` on `opus`: a test-only seam.\n", typ: "flow-low", out: "LOOP-DISPATCH-OK"},
		{name: "backticked medium", chosen: "**Chosen:** `medium` — a gate every run passes.\n**Considered:** `low` — too thin.\n", typ: "flow-medium", out: "LOOP-DISPATCH-OK"},
		{name: "bare value before a backticked one", chosen: "**Chosen:** low — rather than `medium`.\n", typ: "flow-low", out: "LOOP-DISPATCH-OK"},
		{name: "no effort named", chosen: "**Chosen:** the cheap one.\n", typ: "flow-medium", rc: 1, out: "EFFORT —"},
		{name: "a superseded entry is skipped for the active one", chosen: "**Chosen:** low — first take.\n\n### First review, revised\n\n**ID:** first-review-effort\n**Status:** active\n**Chosen:** medium — a hard seam after all.\n", typ: "flow-medium", out: "LOOP-DISPATCH-OK", superseded: true},
		{name: "a status with a trailing clause is still active", chosen: "**Chosen:** low — revised.\n", typ: "flow-low", out: "LOOP-DISPATCH-OK", status: "active, revised"},
		{name: "no Chosen line, the next entry's never read", chosen: "\n### Grouping\n\n**ID:** fixer-grouping\n**Chosen:** one fixer at `low`.\n", typ: "flow-low", rc: 1, out: "EFFORT —"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := lpNew(t, lpPlanDone, lpBranch)
			h := head
			if tc.superseded {
				h = strings.Replace(h, "**Status:** active", "**Status:** superseded", 1)
			}
			if tc.status != "" {
				h = strings.Replace(h, "**Status:** active", "**Status:** "+tc.status, 1)
			}
			writeFile(t, fx.repo+"-worktrees/"+lpBranch+"/spectre/changes/"+lpBranch+"/design.md", h+tc.chosen)
			if rc, out := fx.run(t, "pipeline-fix-6-review-1", tc.typ, "opus"); rc != tc.rc || !strings.Contains(out, tc.out) {
				t.Fatalf("rc %d, want %d; output %q, want it to carry %q", rc, tc.rc, out, tc.out)
			}
		})
	}
}

// TestCheckLoopDispatchSelfReviewFixPairs pins kan-1064's self-review-fix
// row: the fixer runs on opus or sonnet at flow-medium, flow-low is off its
// row, and with no marked row an escalated flow-high fix is refused as a
// pair -- the no-progress answer is a fresh re-plan, never flow-high.
func TestCheckLoopDispatchSelfReviewFixPairs(t *testing.T) {
	t.Parallel()
	const token = "self-review-kan-1-demo"
	for _, tc := range []struct {
		name, key, typ, model, prompt string
		escalated                     bool
		rc                            int
		out                           string
	}{
		{"opus medium", token + "-fix-1", "flow-medium", "opus", "", false, 0, "LOOP-DISPATCH-OK"},
		{"sonnet medium", token + "-fix-1", "flow-medium", "sonnet", "", false, 0, "LOOP-DISPATCH-OK"},
		{"opus low", token + "-fix-1", "flow-low", "opus", "", false, 1, "PAIR —"},
		{"escalated opus high", token + "-fix-1", "flow-high", "opus", "No-progress escalation\n", true, 1, "PAIR —"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := lpNew(t, lpRowsJSON(token+"-", token, "plan completed", "fix completed"), token)
			if tc.escalated {
				gitRun(t, fx.repo, "config", "--add", "flow.escalation", token+":"+tc.key)
			}
			fx.prompt = tc.prompt
			if rc, out := fx.run(t, tc.key, tc.typ, tc.model); rc != tc.rc || !strings.Contains(out, tc.out) {
				t.Fatalf("rc %d, want %d; output %q, want it to carry %q", rc, tc.rc, out, tc.out)
			}
		})
	}
}

// TestCheckLoopDispatchPassRunsOnce pins decision f20-pass-runs-once (kan-1102
// self-review, F20): a pass whose key holds a completed row while its loop
// branch exists is refused, PASS-DONE, and never begun again; an aborted pass
// row, no rows, or no loop branch admit it as before.
func TestCheckLoopDispatchPassRunsOnce(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	const (
		key    = "self-review-kan-1-demo-pass"
		branch = "self-review-kan-1-demo"
		done   = `[{"seq":1,"key":"` + key + `","outcome":"completed","sessionToken":"` + branch + `","endedAt":"2026-11-10T02:00:00Z"}]`
	)
	for _, tc := range []struct {
		name, rows string
		branches   []string
		rc         int
		out        string
	}{
		{name: "a completed pass beside its loop branch", rows: done, branches: []string{branch}, rc: 1, out: "LOOP-DISPATCH-REFUSED: PASS-DONE — " + key + " already completed and its loop branch " + branch + " exists"},
		{name: "a completed pass with no loop branch", rows: done, out: "LOOP-DISPATCH-OK"},
		{name: "an aborted pass beside its loop branch", rows: strings.Replace(done, "completed", "aborted", 1), branches: []string{branch}, out: "LOOP-DISPATCH-OK"},
		{name: "no rows beside the loop branch", rows: `[]`, branches: []string{branch}, out: "LOOP-DISPATCH-OK"},
		{name: "flowd unreadable beside the loop branch", rows: "", branches: []string{branch}, rc: 2, out: "cannot answer"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := lpNew(t, tc.rows, tc.branches...)
			rc, out := fx.run(t, key, "flow-medium", "opus")
			if rc != tc.rc || !strings.Contains(out, tc.out) {
				t.Fatalf("rc %d, want %d; output %q, want it to carry %q", rc, tc.rc, out, tc.out)
			}
			if begun := strings.Contains(fx.calls(t), "dispatch begin"); begun != (tc.rc == 0) {
				t.Fatalf("begin recorded %v, want %v:\n%s", begun, tc.rc == 0, fx.calls(t))
			}
		})
	}
}

// TestCheckLoopDispatchPassNeedsNarrative pins decision f21-narrative-in-run-only
// (kan-1102 self-review, F21): a pass dispatched from the change branch
// (spectre/<name> or <name>) is refused, NARRATIVE, unless the first
// …/bundle.md its prompt names carries a `## Session narrative` line; a pass
// from any other branch is not checked.
func TestCheckLoopDispatchPassNeedsNarrative(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	const key = "self-review-kan-1-demo-pass"
	const bare = "# bundle\n\n## change summary\n\nx\n"
	const told = bare + "\n## Session narrative\n\nThe run planned, built and landed.\n"
	for _, tc := range []struct {
		name, branch, bundle string
		noPath               bool // the prompt names no …/bundle.md path
		rc                   int
		out                  string
	}{
		{name: "an in-run bundle with no narrative", branch: "spectre/kan-1-demo", bundle: bare, rc: 1, out: "LOOP-DISPATCH-REFUSED: NARRATIVE — " + key},
		{name: "an in-run prompt naming no bundle", branch: "spectre/kan-1-demo", bundle: told, noPath: true, rc: 1, out: "LOOP-DISPATCH-REFUSED: NARRATIVE — " + key},
		{name: "an in-run bundle that cannot be read", branch: "spectre/kan-1-demo", rc: 1, out: "/sr/bundle.md carries none"},
		{name: "an in-run bundle with its narrative", branch: "spectre/kan-1-demo", bundle: told, out: "LOOP-DISPATCH-OK"},
		{name: "a standalone pass on the landing branch", branch: "main", bundle: bare, out: "LOOP-DISPATCH-OK"},
		{name: "a /flow-fast pass with its narrative", branch: "kan-1-demo", bundle: told, out: "LOOP-DISPATCH-OK"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := lpNew(t, `[]`)
			dir := filepath.Dir(fx.repo) + "/caller"
			gitRun(t, "", "init", "-q", "-b", tc.branch, dir)
			gitRun(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
			fx.env.Dir = dir
			bundle := filepath.Dir(fx.repo) + "/sr/bundle.md"
			if tc.bundle != "" {
				writeFile(t, bundle, tc.bundle)
			}
			fx.prompt = "Key: " + key + "\nRead the bundle at `" + bundle + "` first.\n"
			if tc.noPath {
				fx.prompt = "Key: " + key + "\nRead the bundle first.\n"
			}
			rc, out := fx.run(t, key, "flow-medium", "opus")
			if rc != tc.rc || !strings.Contains(out, tc.out) {
				t.Fatalf("rc %d, want %d; output %q, want it to carry %q", rc, tc.rc, out, tc.out)
			}
			if begun := strings.Contains(fx.calls(t), "dispatch begin"); begun != (tc.rc == 0) {
				t.Fatalf("begin recorded %v, want %v:\n%s", begun, tc.rc == 0, fx.calls(t))
			}
		})
	}
}
