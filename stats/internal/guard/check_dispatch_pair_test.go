package guard

import (
	"github.com/tweety53/agents/stats/internal/smoke"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// mpRepo is the agents-repo argument every check-dispatch-pair call takes.
const mpRepo = "../../.."

// testPolicy is `flow policy get` as 0061_role_pairs.sql seeds it: the
// Allowed pairs table as it stood when the pairs moved into flowd
// (testdata/policy.json, which internal/repotest's hook tests read too).
var testPolicy = func() string {
	// Located off this file, not the working directory: a test binary
	// self-exec'd as a guard (runSibling) runs from wherever its caller is.
	_, self, _, _ := runtime.Caller(0)
	b, err := os.ReadFile(filepath.Join(filepath.Dir(self), "testdata", "policy.json"))
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(b))
}()

// policyStub is the arm of a fake `flow` answering `policy get` with
// testPolicy; every fake flow of this package's policy readers opens with it.
var policyStub = "[ \"$1 $2\" = \"policy get\" ] && { printf '%s\\n' '" + testPolicy + "'; exit 0; }\n"

func TestPolicyAllowsMarkedRowOnlyWithMarker(t *testing.T) {
	t.Parallel()
	rows, err := policyTable(noFlowEnv(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := policyAllows(rows, "panel-fix", "flow-high", "opus", "", "Role: panel-fix\nfix it", false); err == nil {
		t.Error("panel-fix on opus/flow-high admitted without the marker")
	}
	// The marker counts only as a whole line.
	if err := policyAllows(rows, "panel-fix", "flow-high", "opus", "", "Role: panel-fix\nnot a No-progress escalation here", false); err == nil {
		t.Error("panel-fix on opus/flow-high admitted off a marker inside a line")
	}
	if err := policyAllows(rows, "panel-fix", "flow-high", "opus", "", "No-progress escalation\nRole: panel-fix", false); err != nil {
		t.Errorf("panel-fix on opus/flow-high with the marker refused: %v", err)
	}
	if err := policyAllows(rows, "explorer", "Explore", "opus", "medium", "", false); err != nil {
		t.Errorf("explorer on opus/Explore@medium refused: %v", err)
	}
	if err := policyAllows(rows, "explorer", "Explore", "opus", "high", "", false); err == nil {
		t.Error("explorer on opus/Explore@high admitted")
	}
	if err := policyAllows(rows, "plan-designer", "Plan", "fable", "xhigh", "", false); err != nil {
		t.Errorf("plan-designer on fable/Plan@xhigh refused: %v", err)
	}
	if err := policyAllows(rows, "planner", "flow-planner-high", "fable", "", "", false); err != nil {
		t.Errorf("planner on fable/flow-planner-high refused: %v", err)
	}
	if err := policyAllows(rows, "plan-designer", "Plan", "opus", "xhigh", "", false); err != nil {
		t.Errorf("plan-designer on opus/Plan@xhigh refused: %v", err)
	}
	if err := policyAllows(rows, "planner", "flow-planner-xhigh", "opus", "", "", false); err != nil {
		t.Errorf("planner on opus/flow-planner-xhigh refused: %v", err)
	}
	if err := policyAllows(rows, "explorer", "Explore", "sonnet", "low", "", false); err == nil {
		t.Error("explorer on sonnet/Explore@low admitted")
	}
	if err := policyAllows(rows, "explorer", "Explore", "fable", "medium", "", false); err == nil {
		t.Error("explorer on fable/Explore@medium admitted")
	}
	if err := policyAllows(rows, "panel-reviewer", "flow-medium", "fable", "", "", false); err != nil {
		t.Errorf("panel-reviewer on fable/flow-medium refused: %v", err)
	}
	if err := policyAllows(rows, "panel-reviewer", "flow-low", "fable", "", "", false); err != nil {
		t.Errorf("panel-reviewer on fable/flow-low refused: %v", err)
	}
	if err := policyAllows(rows, "panel-reviewer", "flow-high", "fable", "", "", false); err == nil {
		t.Error("panel-reviewer on fable/flow-high admitted")
	}
	if err := policyAllows(rows, "implementer", "flow-medium", "fable", "", "", false); err == nil {
		t.Error("implementer on fable/flow-medium admitted")
	}
}

func TestPolicyTableNamesNoBannedPair(t *testing.T) {
	t.Parallel()
	rows, err := policyTable(noFlowEnv(t))
	if err != nil {
		t.Fatal(err)
	}
	for role, rs := range rows {
		for _, r := range rs {
			for _, p := range r.Pairs {
				// fable (medium..xhigh) is admitted for the two planning roles only (operator, 2026-11-09).
				planning := role == "planner" || role == "plan-designer"
				reviewer := role == "panel-reviewer" && strings.HasPrefix(p, "fable/flow-") // fable low and medium only, operator 2026-11-09
				if strings.Contains(p, "haiku") || (strings.Contains(p, "fable") && !planning && !reviewer) ||
					(strings.Contains(p, "xhigh") && !planning) ||
					(planning && strings.Contains(p, "fable") && strings.HasSuffix(p, "low")) {
					t.Errorf("role %s names banned pair %s", role, p)
				}
			}
		}
	}
}

// smoke: core — the hook-facing exit codes of check-dispatch-pair
func TestCheckDispatchPairExitCodes(t *testing.T) {
	t.Parallel()
	env := noFlowEnv(t)
	for _, c := range []struct {
		name, prompt string
		args         []string
		rc           int
		out          string
	}{
		{"allowed", "Role: panel-merger\n", []string{mpRepo, "flow-low", "sonnet"}, 0, "DISPATCH-PAIR-OK panel-merger"},
		{"explore effort", "Role: explorer\n", []string{mpRepo, "Explore", "opus", "low"}, 0, "DISPATCH-PAIR-OK explorer"},
		{"no role line", "Role panel-merger\n", []string{mpRepo, "flow-low", "sonnet"}, 1, "DISPATCH-PAIR-REFUSED: no `Role:` line"},
		{"unknown role", "Role: nobody\n", []string{mpRepo, "flow-low", "sonnet"}, 1, "DISPATCH-PAIR-REFUSED: unknown role"},
		{"wrong pair", "Role: panel-merger\n", []string{mpRepo, "flow-low", "opus"}, 1, "DISPATCH-PAIR-REFUSED: panel-merger"},
		{"usage", "", []string{mpRepo}, 2, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			e := env
			e.Stdin = strings.NewReader(c.prompt)
			r := runGuard("check-dispatch-pair", c.args, e)
			if r.rc != c.rc || !strings.HasPrefix(r.stdout, c.out) {
				t.Errorf("rc=%d out=%q, want %d and %q", r.rc, r.out, c.rc, c.out)
			}
		})
	}
}

// TestCheckDispatchPairNoFlowdNoCacheRefuses: no answer from `flow policy
// get` -- flowd down and no cache, an unparsable answer, or no flow at all --
// exits the guard 2 naming flowd and the cache: never a dispatch admitted
// off pairs nobody read.
func TestCheckDispatchPairNoFlowdNoCacheRefuses(t *testing.T) {
	t.Parallel()
	for name, env := range map[string]Env{
		"flowd down, no cache": fakeFlowEnvBare(t, "echo 'flow: policy get: flowd unreachable (dial) and no usable cache at /x/policy-cache.json' >&2; exit 1"),
		"not json":             fakeFlowEnvBare(t, "echo 'not json'"),
		"no flow":              envWith(t, []string{"PATH=" + t.TempDir()}),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			e := env
			e.Stdin = strings.NewReader("Role: panel-merger\n")
			r := runGuard("check-dispatch-pair", []string{mpRepo, "flow-low", "sonnet"}, e)
			if r.rc != 2 || !strings.Contains(r.out, "flowd unreachable and no policy cache") {
				t.Fatalf("rc=%d out=%q, want 2 naming flowd and the cache", r.rc, r.out)
			}
		})
	}
}

// TestCheckDispatchPairOptinNeedsLineAndGrant: an opt-in pair is admitted
// only with the whole line `Operator pair: <role>` and a live grant for the
// change on the prompt's Worktree branch -- the line alone, the grant
// alone, or a worktree off a spectre/ branch is refused.
func TestCheckDispatchPairOptinNeedsLineAndGrant(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	wt := t.TempDir()
	for _, args := range [][]string{{"init", "-q", "-b", "spectre/kan-9"}, {"commit", "-q", "--allow-empty", "-m", "x"}} {
		if out, err := exec.Command("git", append([]string{"-C", wt, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	off := t.TempDir()
	if out, err := exec.Command("git", "-C", off, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	policy := `{"planner":{"base":["opus/flow-planner-low"],"optin":["fable/flow-planner-xhigh"],"marked":[]}}`
	stub := func(grants string) Env {
		return fakeFlowEnvBare(t, `case "$1 $2" in
"policy get") printf '%s\n' '`+policy+`' ;;
"policy grants") [ "$4" = kan-9 ] && printf '%s\n' '`+grants+`' || echo '[]' ;;
*) exit 1 ;;
esac
`)
	}
	for _, c := range []struct {
		name, grants, prompt string
		rc                   int
	}{
		{"line and grant", `["planner"]`, "Role: planner\nWorktree: " + wt + "\nOperator pair: planner\n", 0},
		{"no line", `["planner"]`, "Role: planner\nWorktree: " + wt + "\n", 1},
		{"line inside a line", `["planner"]`, "Role: planner\nWorktree: " + wt + "\nnot Operator pair: planner\n", 1},
		{"no grant", `[]`, "Role: planner\nWorktree: " + wt + "\nOperator pair: planner\n", 1},
		{"grant for another role", `["plan-designer"]`, "Role: planner\nWorktree: " + wt + "\nOperator pair: planner\n", 1},
		{"off a spectre branch", `["planner"]`, "Role: planner\nWorktree: " + off + "\nOperator pair: planner\n", 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			e := stub(c.grants)
			e.Stdin = strings.NewReader(c.prompt)
			r := runGuard("check-dispatch-pair", []string{mpRepo, "flow-planner-xhigh", "fable"}, e)
			if r.rc != c.rc {
				t.Fatalf("rc=%d out=%q, want %d", r.rc, r.out, c.rc)
			}
			if c.rc == 1 && !strings.Contains(r.stdout, "Operator pair: planner` and an operator grant") {
				t.Fatalf("refusal %q does not name the line and the grant", r.stdout)
			}
		})
	}
	// The base pair needs neither.
	e := stub(`[]`)
	e.Stdin = strings.NewReader("Role: planner\n")
	if r := runGuard("check-dispatch-pair", []string{mpRepo, "flow-planner-low", "opus"}, e); r.rc != 0 {
		t.Fatalf("base pair: rc=%d out=%q, want 0", r.rc, r.out)
	}
}

// TestCheckDispatchPairRoleOnly: --role-only answers whether the Role line
// names a role of the table, for a harness whose pair is a fixed mapping.
func TestCheckDispatchPairRoleOnly(t *testing.T) {
	t.Parallel()
	for prompt, rc := range map[string]int{"Role: visual-verifier\n": 0, "Role: nobody\n": 1, "verify\n": 1} {
		e := noFlowEnv(t)
		e.Stdin = strings.NewReader(prompt)
		if r := runGuard("check-dispatch-pair", []string{mpRepo, "--role-only"}, e); r.rc != rc {
			t.Errorf("%q: rc=%d out=%q, want %d", prompt, r.rc, r.out, rc)
		}
	}
}

// noFlowEnv is an env whose only `flow` answers `policy get` with
// testPolicy and fails everything else: the seeded pairs, no experiment.
func noFlowEnv(t *testing.T) Env {
	t.Helper()
	return fakeFlowEnv(t, "exit 1\n")
}

// fakeFlowEnv is an env whose PATH holds only a `flow` answering `policy
// get` with testPolicy and running script for everything else.
func fakeFlowEnv(t *testing.T, script string, over ...string) Env {
	t.Helper()
	return fakeFlowEnvBare(t, policyStub+script, over...)
}

// policyFlowEnv is an env whose only `flow` answers `policy get` with
// policy and fails everything else.
func policyFlowEnv(t *testing.T, policy string, over ...string) Env {
	t.Helper()
	return fakeFlowEnvBare(t, "[ \"$1 $2\" = \"policy get\" ] && { printf '%s\\n' '"+policy+"'; exit 0; }\nexit 1\n", over...)
}

// fakeFlowEnvBare is an env whose PATH holds only a `flow` running script.
func fakeFlowEnvBare(t *testing.T, script string, over ...string) Env {
	t.Helper()
	bin := t.TempDir()
	writeExec(t, bin+"/flow", "#!/bin/sh\n"+script)
	return envWith(t, append([]string{"PATH=" + bin}, over...))
}

const overridesFlow = `[ "$1 $2" = "experiment overrides" ] || exit 9
printf '%s\n' '{"plan-designer":["opus/Plan@max"],"main":["opus@medium"],"nobody":["opus/flow-low"]}'
`

// TestPolicyTableOverlaysOpenOverrides: an open experiment's cell replaces
// its role's unmarked row whole, a planning role's replaced pairs appended as
// a row marked `CORE change`; marked rows, roles the table lacks and main
// never change the table; a failed or unparsable overrides read, or no flow
// at all, leaves the table alone.
func TestPolicyTableOverlaysOpenOverrides(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	alone, err := policyTable(noFlowEnv(t))
	if err != nil {
		t.Fatal(err)
	}
	got, err := policyTable(fakeFlowEnv(t, overridesFlow))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]policyRow{}
	for role, rs := range alone {
		want[role] = slices.Clone(rs)
	}
	want["plan-designer"][0].Pairs = []string{"opus/Plan@max"}
	// A planning row keeps its own pairs on an appended `CORE change` row (kan-1102 self-review, F4).
	want["plan-designer"] = append(want["plan-designer"], policyRow{Pairs: alone["plan-designer"][0].Pairs, Marker: "CORE change"})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("overlaid table differs beyond plan-designer's unmarked row:\n got %v\nwant %v", got, want)
	}
	if _, ok := got["main"]; ok {
		t.Fatal("main entered the guard table")
	}

	marked, err := policyTable(fakeFlowEnv(t, `printf '%s\n' '{"panel-fix":["opus/flow-max"]}'`))
	if err != nil {
		t.Fatal(err)
	}
	// A non-planning role gets no `CORE change` row: len unchanged.
	if rs := marked["panel-fix"]; len(rs) != len(alone["panel-fix"]) || rs[0].Marker != "" || !reflect.DeepEqual(rs[0].Pairs, []string{"opus/flow-max"}) ||
		rs[1].Marker == "" || !reflect.DeepEqual(rs[1], alone["panel-fix"][1]) {
		t.Fatalf("panel-fix rows = %v; want the unmarked row replaced and the marked row untouched", rs)
	}

	for name, env := range map[string]Env{
		"exit 1":   fakeFlowEnv(t, "echo '{\"plan-designer\":[\"opus/Plan@max\"]}'; exit 1"),
		"not json": fakeFlowEnv(t, "echo 'not json'"),
		"no flow":  noFlowEnv(t),
	} {
		rows, err := policyTable(env)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !reflect.DeepEqual(rows, alone) {
			t.Errorf("%s: table changed: %v", name, rows["plan-designer"])
		}
	}
}

// TestCheckDispatchPairAdmitsOverride: an open override admits its pair, and
// without it the table refuses it.
func TestCheckDispatchPairAdmitsOverride(t *testing.T) {
	t.Parallel()
	args := []string{mpRepo, "Plan", "opus", "max"}
	e := fakeFlowEnv(t, overridesFlow)
	e.Stdin = strings.NewReader("Role: plan-designer\n")
	if r := runGuard("check-dispatch-pair", args, e); r.rc != 0 || r.stdout != "DISPATCH-PAIR-OK plan-designer\n" {
		t.Fatalf("with the override: rc=%d out=%q, want admitted", r.rc, r.out)
	}
	e = noFlowEnv(t)
	e.Stdin = strings.NewReader("Role: plan-designer\n")
	if r := runGuard("check-dispatch-pair", args, e); r.rc != 1 {
		t.Fatalf("without the override: rc=%d out=%q, want refused", r.rc, r.out)
	}
}

// TestCheckDispatchPairAdmitsPlanner: the real table's `planner` row admits
// opus and fable on flow-planner-medium to -xhigh (opus also -low), nothing else (KAN-1067). No `flow` on
// PATH, so an open experiment in the dev store cannot change the answer.
func TestCheckDispatchPairAdmitsPlanner(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		agent, model string
		rc           int
	}{
		{"flow-planner-medium", "opus", 0},
		{"flow-planner-low", "opus", 0},
		{"flow-planner-high", "opus", 0},
		{"flow-planner-xhigh", "opus", 0},
		{"flow-planner-xhigh", "fable", 0},
		{"flow-planner-low", "fable", 1},
		{"flow-planner-max", "opus", 1},
		{"flow-planner-medium", "sonnet", 1},
	} {
		e := noFlowEnv(t)
		e.Stdin = strings.NewReader("Role: planner\n")
		if r := runGuard("check-dispatch-pair", []string{mpRepo, c.agent, c.model}, e); r.rc != c.rc {
			t.Errorf("%s/%s: rc=%d out=%q, want %d", c.model, c.agent, r.rc, r.out, c.rc)
		}
	}
}

// TestCorePairsSurviveOverride pins decision f4-core-pairs-survive-override
// (kan-1102 self-review, F4): an open experiment overriding a planning row
// keeps the row's own pairs admitted for a prompt carrying the line
// `CORE change`, and plan-agent-pair … core still finds the CORE pair.
func TestCorePairsSurviveOverride(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	env := fakeFlowEnv(t, `[ "$1 $2" = "experiment overrides" ] || exit 9
printf '%s\n' '{"explorer":["sonnet/Explore@low"],"plan-designer":["opus/Plan@medium"]}'
`, "FLOW_GUARD_REPO_ROOT="+root)
	rows, err := policyTable(env)
	if err != nil {
		t.Fatal(err)
	}
	if err := policyAllows(rows, "explorer", "Explore", "opus", "low", "Role: explorer\nCORE change\n", false); err != nil {
		t.Errorf("opus/Explore@low with the CORE change line refused: %v", err)
	}
	if err := policyAllows(rows, "explorer", "Explore", "opus", "low", "Role: explorer\n", false); err == nil {
		t.Error("opus/Explore@low admitted without the CORE change line: the experiment silently lifted")
	}
	if err := policyAllows(rows, "explorer", "Explore", "sonnet", "low", "Role: explorer\n", false); err != nil {
		t.Errorf("the override's sonnet/Explore@low refused: %v", err)
	}
	for role, want := range map[string]string{"explorer": "opus low\n", "plan-designer": "fable xhigh\n"} {
		if r := runGuard("plan-agent-pair", []string{"demo", "1", role, "core"}, env); r.rc != 0 || r.stdout != want {
			t.Errorf("%s core: rc=%d out=%q, want %q", role, r.rc, r.out, want)
		}
	}
}
