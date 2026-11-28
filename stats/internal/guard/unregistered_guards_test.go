package guard

import (
	"slices"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/hook"
	"github.com/tweety53/agents/stats/internal/guard/integrate"
	"github.com/tweety53/agents/stats/internal/guard/visual"
	"github.com/tweety53/agents/stats/internal/records"
)

// inProcessGuards are the guards that are no guard.Registry key: each runs only in process from
// its one caller, and the tests reach it through runGuard by its former name.
var inProcessGuards = map[string]core.Func{
	"check-identical-baselines": visual.CheckIdenticalBaselines,   // visual-verify-pass measure
	"check-main-checkout-drift": integrate.CheckMainCheckoutDrift, // integrate-change prepare
	"check-ui-task-measured":    integrate.CheckUITaskMeasured,    // close-task, through utmJudge
	"sweep-dispatch-rows":       hook.SweepDispatchRows,           // the await-own-agents hook
	"end-stale-dispatch-rows":   hook.EndStaleDispatchRows,        // the await-own-agents hook and await-rows
}

func TestInProcessGuardsAreUnregistered(t *testing.T) {
	t.Parallel()
	for n := range inProcessGuards {
		if _, ok := Registry[n]; ok {
			t.Errorf("%s is a guard.Registry key; it runs in process only", n)
		}
	}
}

// TestNoLongerRunGuardsAreHidden: the dispatch-row sweeps (in process only) and capture-workers
// (folded into the visual-verify bundle) record no run any more, so their past guard_runs rows are
// hidden from guard health and prune-candidates' never-fired.
func TestNoLongerRunGuardsAreHidden(t *testing.T) {
	t.Parallel()
	for _, n := range []string{"sweep-dispatch-rows", "end-stale-dispatch-rows", "capture-workers"} {
		if !slices.Contains(records.RetiredGuards, n) {
			t.Errorf("%s is not in records.RetiredGuards", n)
		}
	}
}
