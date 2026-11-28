package integrate

import (
	"github.com/tweety53/agents/stats/internal/guard/core"
	_ "github.com/tweety53/agents/stats/internal/guard/gitops"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	_ "github.com/tweety53/agents/stats/internal/guard/hook"
	_ "github.com/tweety53/agents/stats/internal/guard/plans"
	_ "github.com/tweety53/agents/stats/internal/guard/visual"
)

// runGuard runs the guard named name in process — check-main-checkout-drift
// and check-ui-task-measured, which no Registry key names, first, then a
// Registry one; the blank imports register the other families' guards these
// tests call.
func runGuard(name string, args []string, env core.Env) guardtest.GuardResult {
	fn := map[string]core.Func{
		"check-main-checkout-drift": CheckMainCheckoutDrift,
		"check-ui-task-measured":    CheckUITaskMeasured,
	}[name]
	if fn == nil {
		fn = core.Registry[name]
	}
	return guardtest.RunFunc(name, fn, args, env)
}
