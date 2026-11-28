package gitops

import (
	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// runGuard runs the guard named name in process: check-base-moved, which no
// Registry key names, or a Registry one.
func runGuard(name string, args []string, env core.Env) guardtest.GuardResult {
	fn := core.Registry[name]
	if name == "check-base-moved" {
		fn = CheckBaseMoved
	}
	return guardtest.RunFunc(name, fn, args, env)
}
