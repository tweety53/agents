package plans

import (
	"github.com/tweety53/agents/stats/internal/guard/core"
	_ "github.com/tweety53/agents/stats/internal/guard/gitops"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// runGuard runs the Registry guard named name in process; the gitops import
// registers throwaway-worktree, which the parallel-groups tests call.
func runGuard(name string, args []string, env core.Env) guardtest.GuardResult {
	return guardtest.RunFunc(name, core.Registry[name], args, env)
}
