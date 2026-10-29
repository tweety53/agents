package guard

import (
	"strings"
	"testing"
)

// The real table, read from this checkout.
const mpRepo = "../../.."

func TestPolicyTableParsesRepo(t *testing.T) {
	t.Parallel()
	rows, err := policyTable(mpRepo)
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{
		"explorer", "plan-designer", "implementer", "implementer-group", "panel-fix", "gated-fix",
		"visual-verify-fix", "panel-reviewer", "panel-extra", "panel-rerun", "gated-reviewer",
		"gated-re-reviewer", "visual-verifier", "late-fix", "panel-merger", "tooling-analyst",
		"self-review-pass", "self-review-plan", "pipeline-fix-plan", "self-review-fix", "pipeline-fix-fix",
		"self-review-review-1", "pipeline-fix-review-1", "self-review-review-r", "pipeline-fix-review-r",
	} {
		if len(rows[role]) == 0 {
			t.Errorf("role %q has no row", role)
		}
	}
}

func TestPolicyAllowsMarkedRowOnlyWithMarker(t *testing.T) {
	t.Parallel()
	rows, err := policyTable(mpRepo)
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
	rows, err := policyTable(mpRepo)
	if err != nil {
		t.Fatal(err)
	}
	for role, rs := range rows {
		for _, r := range rs {
			for _, p := range r.pairs {
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
