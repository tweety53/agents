package guard

import (
	"bytes"
	"fmt"
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
	fixtureRepo(t, fx.repo, repoShape{Branch: "main"})
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
	rc := checkLoopDispatch(append([]string{fx.repo}, args...), fx.env, &o, &e)
	return rc, o.String() + e.String()
}

func (fx lpFx) calls(t *testing.T) string {
	t.Helper()
	b, _ := os.ReadFile(fx.log)
	return string(b)
}

const lpBranch = "pipeline-fix-kan-1-demo-6"

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

// TestLoopSuffixRoles pins the key suffix -> role id map onto the real table:
// every suffix lands on a role the allowed-pairs table carries.
func TestLoopSuffixRoles(t *testing.T) {
	t.Parallel()
	rows, err := policyTable(noFlowEnv(t))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"pass": "-pass", "plan": "-plan", "plan-2": "-plan", "fix": "-fix", "fix-3": "-fix", "fix-g1": "-fix",
		"fix-v3": "-fix", "fix-r1": "-fix", "review-1": "-review-1", "review-2": "-review-r", "review-r1": "-review-r"}
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
	for suffix, family := range map[string]string{"fix-r1": "fix-r", "review-r1": "review"} {
		if got := lpFamily(suffix); got != family {
			t.Errorf("suffix %q: family %q, want %q", suffix, got, family)
		}
	}
	for _, bad := range []string{"fix-r", "review-r", "fix-x", "review", "plan-a", "pass-1"} {
		if got := lpRole("self-review", bad); got != "" {
			t.Errorf("suffix %q: role %q, want none", bad, got)
		}
	}
}

// TestCheckLoopDispatch pins that check-loop-dispatch refuses nothing: it
// opens the loop dispatch's row, or exits 2 with nothing recorded when the key
// names no loop it can record under.
func TestCheckLoopDispatch(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, tc := range []struct {
		name, key, typ, model string
		branches              []string
		rc                    int
		out, begin            string // begin "" asserts no begin call
	}{
		{name: "plan recorded", key: "pipeline-fix-6-plan", typ: "flow-planner-medium", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-change " + lpBranch + " -role planner -model opus -effort medium -key pipeline-fix-6-plan -session-token " + lpBranch},
		{name: "fix recorded with no plan before it", key: "pipeline-fix-6-fix", typ: "flow-low", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-role implementer -model opus -effort low -key pipeline-fix-6-fix"},
		{name: "fix on flow-high recorded with no escalation", key: "pipeline-fix-6-fix-1", typ: "flow-high", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-effort high"},
		{name: "fix on flow-xhigh recorded at xhigh", key: "pipeline-fix-6-fix-1", typ: "flow-xhigh", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-effort xhigh"},
		{name: "first review recorded on any effort", key: "pipeline-fix-6-review-1", typ: "flow-medium", model: "opus", branches: []string{lpBranch},
			out: "LOOP-DISPATCH-OK", begin: "-role reviewer"},
		{name: "no loop branch", key: "pipeline-fix-6-plan", typ: "flow-planner-medium", model: "opus", rc: 2, out: "not recorded: NOT-STARTED —"},
		{name: "an unknown suffix", key: "pipeline-fix-6-implementer", typ: "flow-low", model: "opus", branches: []string{lpBranch}, rc: 2, out: "not recorded: KEY —"},
		{name: "no loop key", key: "panel-1-primary", typ: "flow-low", model: "opus", out: "LOOP-DISPATCH-NA"},
		{name: "self-review on its own branch files under the change", key: "self-review-kan-1-demo-plan", typ: "flow-planner-medium", model: "opus",
			out: "LOOP-DISPATCH-OK", begin: "-change kan-1-demo -role planner"},
		{name: "self-review loop files under its branch", key: "self-review-kan-1-demo-plan", typ: "flow-planner-medium", model: "opus", branches: []string{"self-review-kan-1-demo"},
			out: "LOOP-DISPATCH-OK", begin: "-change self-review-kan-1-demo -role planner -model opus -effort medium -key self-review-kan-1-demo-plan -session-token self-review-kan-1-demo"},
		{name: "the pass is recorded under the change itself", key: "self-review-kan-1-demo-pass", typ: "flow-medium", model: "opus",
			out: "LOOP-DISPATCH-OK", begin: "-change kan-1-demo -role reviewer -model opus -effort medium -key self-review-kan-1-demo-pass -session-token self-review-kan-1-demo"},
		{name: "the pass stays under the change while its loop branch exists", key: "self-review-kan-1-demo-pass", typ: "flow-medium", model: "opus", branches: []string{"self-review-kan-1-demo"},
			out: "LOOP-DISPATCH-OK", begin: "-change kan-1-demo -role reviewer"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := lpNew(t, `[]`, tc.branches...)
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

// TestCheckLoopDispatchBeginRefused: a begin flowd refuses records nothing
// and exits 2, which the hook logs and lets through.
func TestCheckLoopDispatchBeginRefused(t *testing.T) {
	t.Parallel()
	fx := lpNew(t, `[]`, lpBranch)
	bin := t.TempDir()
	writeExec(t, bin+"/flow", "#!/bin/sh\nexit 1\n")
	fx.env.Getenv = pathEnv(bin)
	if rc, out := fx.run(t, "pipeline-fix-6-plan", "flow-medium", "opus"); rc != 2 || !strings.Contains(out, "not recorded") {
		t.Fatalf("rc %d, want 2: %s", rc, out)
	}
}

// A self-review on a change's own branch in another project -- no
// self-review-<name> branch in <agents repo> -- files its rows in that
// project's flowd project: flow runs with -C the caller's main checkout.
// A self-review-<name> loop branch in <agents repo> still files there.
func TestCheckLoopDispatchOwnBranchInAnotherProject(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := lpNew(t, `[]`)
	proj := filepath.Dir(fx.repo) + "/gymie"
	writeFile(t, proj+"/README.md", "gymie\n")
	fixtureRepo(t, proj, repoShape{Branch: "develop"})
	gitRun(t, proj, "add", "-A")
	gitRun(t, proj, "commit", "-q", "-m", "init")
	gitRun(t, proj, "branch", "spectre/kan-1-demo")
	fx.env.Dir = proj
	if rc, out := fx.run(t, "self-review-kan-1-demo-review-1", "flow-low", "opus"); rc != 0 {
		t.Fatalf("rc %d: %s", rc, out)
	}
	if calls := fx.calls(t); !strings.Contains(calls, "record dispatch begin -C "+proj+" -change kan-1-demo -role reviewer") {
		t.Fatalf("the own-branch loop was not filed in the caller's project:\n%s", calls)
	}
	gitRun(t, fx.repo, "branch", "self-review-kan-1-demo")
	if rc, out := fx.run(t, "self-review-kan-1-demo-fix", "flow-medium", "opus"); rc != 0 ||
		!strings.Contains(fx.calls(t), "record dispatch begin -C "+fx.repo+" -change self-review-kan-1-demo") {
		t.Fatalf("a self-review loop branch left <agents repo>: rc %d: %s\n%s", rc, out, fx.calls(t))
	}
}
