package visual

import (
	"github.com/tweety53/agents/stats/internal/guard/core"
	_ "github.com/tweety53/agents/stats/internal/guard/gitops"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	_ "github.com/tweety53/agents/stats/internal/guard/hook"
	_ "github.com/tweety53/agents/stats/internal/guard/plans"
)

// runGuard runs the guard named name in process — a guard folded into
// visual-verify-pass or check-identical-baselines, which no Registry key
// names, first, then a Registry one; the blank imports register
// the other families' guards these tests call.
func runGuard(name string, args []string, env core.Env) guardtest.GuardResult {
	fn := FoldedGuards[name]
	if name == "check-identical-baselines" {
		fn = CheckIdenticalBaselines
	}
	if fn == nil {
		fn = core.Registry[name]
	}
	return guardtest.RunFunc(name, fn, args, env)
}
