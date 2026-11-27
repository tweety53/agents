package guard

import "testing"

// inProcessGuards are the guards that are no guard.Registry key: each runs only in process from
// its one caller, and the tests reach it through runGuard by its former name.
var inProcessGuards = map[string]Func{
	"check-identical-baselines": checkIdenticalBaselines, // visual-verify-pass measure
	"check-main-checkout-drift": checkMainCheckoutDrift,  // integrate-change prepare
	"check-ui-task-measured":    checkUITaskMeasured,     // close-task, through utmJudge
	"sweep-dispatch-rows":       sweepDispatchRows,       // the await-own-agents hook
	"end-stale-dispatch-rows":   endStaleDispatchRows,    // the await-own-agents hook and await-rows
}

func TestInProcessGuardsAreUnregistered(t *testing.T) {
	t.Parallel()
	for n := range inProcessGuards {
		if _, ok := Registry[n]; ok {
			t.Errorf("%s is a guard.Registry key; it runs in process only", n)
		}
	}
}
