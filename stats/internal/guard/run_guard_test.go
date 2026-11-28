package guard

import (
	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// runGuard runs the guard named name in process — a folded or in-process-only
// guard first, then a Registry one — through runFunc.
func runGuard(name string, args []string, env core.Env) guardtest.GuardResult {
	fn := foldedGuards[name]
	if fn == nil {
		fn = inProcessGuards[name]
	}
	if fn == nil {
		fn = Registry[name]
	}
	return guardtest.RunFunc(name, fn, args, env)
}
