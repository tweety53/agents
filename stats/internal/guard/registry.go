// Package guard holds the Go ports of the flow guard scripts. Each guard
// keeps its bash script's CLI contract; scripts/<name>.sh execs flow-guard.
package guard

import (
	"github.com/tweety53/agents/stats/internal/guard/core"

	// Every family registers its guards from init(); the blank imports keep
	// each registered even where guard uses nothing else of it.
	_ "github.com/tweety53/agents/stats/internal/guard/gitops"
	_ "github.com/tweety53/agents/stats/internal/guard/hook"
	_ "github.com/tweety53/agents/stats/internal/guard/integrate"
	_ "github.com/tweety53/agents/stats/internal/guard/plans"
	_ "github.com/tweety53/agents/stats/internal/guard/visual"
)

// Registry is core.Registry, the one map every guard registers into from
// its file's init(); guard imports every package that registers one, so a
// reader of guard.Registry sees every guard.
var Registry = core.Registry
