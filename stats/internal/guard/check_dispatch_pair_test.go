package guard

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
)

// The real table, read from this checkout.
const mpRepo = "../../.."

func TestPolicyAllowsMarkedRowOnlyWithMarker(t *testing.T) {
	t.Parallel()
	rows, err := policyTable(noFlowEnv(t), mpRepo)
	if err != nil {
		t.Fatal(err)
	}
	if err := policyAllows(rows, "panel-fix", "flow-high", "opus", "", "Role: panel-fix\nfix it"); err == nil {
		t.Error("panel-fix on opus/flow-high admitted without the marker")
	}
	// The marker counts only as a whole line.
	if err := policyAllows(rows, "panel-fix", "flow-high", "opus", "", "Role: panel-fix\nnot a No-progress escalation here"); err == nil {
		t.Error("panel-fix on opus/flow-high admitted off a marker inside a line")
	}
	if err := policyAllows(rows, "panel-fix", "flow-high", "opus", "", "No-progress escalation\nRole: panel-fix"); err != nil {
		t.Errorf("panel-fix on opus/flow-high with the marker refused: %v", err)
	}
	if err := policyAllows(rows, "explorer", "Explore", "opus", "medium", ""); err != nil {
		t.Errorf("explorer on opus/Explore@medium refused: %v", err)
	}
	if err := policyAllows(rows, "explorer", "Explore", "opus", "high", ""); err == nil {
		t.Error("explorer on opus/Explore@high admitted")
	}
}

func TestPolicyTableNamesNoBannedPair(t *testing.T) {
	t.Parallel()
	rows, err := policyTable(noFlowEnv(t), mpRepo)
	if err != nil {
		t.Fatal(err)
	}
	for role, rs := range rows {
		for _, r := range rs {
			for _, p := range r.Pairs {
				if strings.Contains(p, "haiku") || strings.Contains(p, "fable") || strings.Contains(p, "xhigh") {
					t.Errorf("role %s names banned pair %s", role, p)
				}
			}
		}
	}
}

func TestCheckDispatchPairExitCodes(t *testing.T) {
	t.Parallel()
	env := envWith(t, nil)
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
		{"missing repo", "Role: panel-merger\n", []string{"/nonexistent", "flow-low", "sonnet"}, 2, ""},
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

// TestPolicyTableRefusesMalformedRows: a row naming two roles, no pair or two
// markers exits the guard 2 -- never a row loaded fail-open. The parse cases
// themselves are modelpolicy's TestParse.
func TestPolicyTableRefusesMalformedRows(t *testing.T) {
	t.Parallel()
	for name, row := range map[string]string{
		"two roles":   "| `a`, `b` | `opus/flow-low` | — |",
		"no pair":     "| `a` | — | — |",
		"two markers": "| `a` | `opus/flow-high` | `X`, `Y` |",
		"two cells":   "| `a` | `opus/flow-low` |",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			repo := t.TempDir()
			writeFile(t, repo+"/"+mpFile, "| Role | Allowed pairs | Requires marker |\n|---|---|---|\n"+row+"\n")
			e := envWith(t, nil)
			e.Stdin = strings.NewReader("Role: a\n")
			if r := runGuard("check-dispatch-pair", []string{repo, "flow-high", "opus"}, e); r.rc != 2 {
				t.Fatalf("rc=%d out=%q, want 2", r.rc, r.out)
			}
		})
	}
}

// TestCheckDispatchPairRoleOnly: --role-only answers whether the Role line
// names a role of the table, for a harness whose pair is a fixed mapping.
func TestCheckDispatchPairRoleOnly(t *testing.T) {
	t.Parallel()
	for prompt, rc := range map[string]int{"Role: visual-verifier\n": 0, "Role: nobody\n": 1, "verify\n": 1} {
		e := envWith(t, nil)
		e.Stdin = strings.NewReader(prompt)
		if r := runGuard("check-dispatch-pair", []string{mpRepo, "--role-only"}, e); r.rc != rc {
			t.Errorf("%q: rc=%d out=%q, want %d", prompt, r.rc, r.out, rc)
		}
	}
}

// noFlowEnv is an env whose PATH holds no `flow`: the table alone.
func noFlowEnv(t *testing.T) Env {
	t.Helper()
	return envWith(t, []string{"PATH=" + t.TempDir()})
}

// fakeFlowEnv is an env whose PATH holds only a `flow` running script.
func fakeFlowEnv(t *testing.T, script string, over ...string) Env {
	t.Helper()
	bin := t.TempDir()
	writeExec(t, bin+"/flow", "#!/bin/sh\n"+script)
	return envWith(t, append([]string{"PATH=" + bin}, over...))
}

const overridesFlow = `[ "$1 $2" = "experiment overrides" ] || exit 9
printf '%s\n' '{"plan-designer":["opus/Plan@xhigh"],"main":["opus@medium"],"nobody":["opus/flow-low"]}'
`

// TestPolicyTableOverlaysOpenOverrides: an open experiment's cell replaces
// its role's unmarked row whole; marked rows, roles the table lacks and main
// never change the table; a failed or unparsable overrides read, or no flow
// at all, leaves the table alone.
func TestPolicyTableOverlaysOpenOverrides(t *testing.T) {
	t.Parallel()
	alone, err := policyTable(noFlowEnv(t), mpRepo)
	if err != nil {
		t.Fatal(err)
	}
	got, err := policyTable(fakeFlowEnv(t, overridesFlow), mpRepo)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]modelpolicy.Row{}
	for role, rs := range alone {
		want[role] = slices.Clone(rs)
	}
	want["plan-designer"][0].Pairs = []string{"opus/Plan@xhigh"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("overlaid table differs beyond plan-designer's unmarked row:\n got %v\nwant %v", got, want)
	}
	if _, ok := got["main"]; ok {
		t.Fatal("main entered the guard table")
	}

	marked, err := policyTable(fakeFlowEnv(t, `printf '%s\n' '{"panel-fix":["opus/flow-max"]}'`), mpRepo)
	if err != nil {
		t.Fatal(err)
	}
	if rs := marked["panel-fix"]; rs[0].Marker != "" || !reflect.DeepEqual(rs[0].Pairs, []string{"opus/flow-max"}) ||
		rs[1].Marker == "" || !reflect.DeepEqual(rs[1], alone["panel-fix"][1]) {
		t.Fatalf("panel-fix rows = %v; want the unmarked row replaced and the marked row untouched", rs)
	}

	for name, env := range map[string]Env{
		"exit 1":   fakeFlowEnv(t, "echo '{\"plan-designer\":[\"opus/Plan@xhigh\"]}'; exit 1"),
		"not json": fakeFlowEnv(t, "echo 'not json'"),
		"no flow":  noFlowEnv(t),
	} {
		rows, err := policyTable(env, mpRepo)
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
	args := []string{mpRepo, "Plan", "opus", "xhigh"}
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
// opus on flow-planner-medium and -low, nothing else (KAN-1067). No `flow` on
// PATH, so an open experiment in the dev store cannot change the answer.
func TestCheckDispatchPairAdmitsPlanner(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		agent, model string
		rc           int
	}{
		{"flow-planner-medium", "opus", 0},
		{"flow-planner-low", "opus", 0},
		{"flow-planner-high", "opus", 1},
		{"flow-planner-medium", "sonnet", 1},
	} {
		e := noFlowEnv(t)
		e.Stdin = strings.NewReader("Role: planner\n")
		if r := runGuard("check-dispatch-pair", []string{mpRepo, c.agent, c.model}, e); r.rc != c.rc {
			t.Errorf("%s/%s: rc=%d out=%q, want %d", c.model, c.agent, r.rc, r.out, c.rc)
		}
	}
}
