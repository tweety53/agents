package records

import (
	"bytes"
	"encoding/json"
)

// RoleTaskCost is one (role, task) group's totals in the SDD ledger's cost
// section: what kan-450's cost-per-change figure is broken down into. The
// token fields sum both of the metrics bag's buckets (main + sidechain) the
// same way tokenLine renders a single dispatch, over rows whose bag carried
// a "tokens" key at all; CostUSD sums the rows' present CostUSD figures.
// An unmeasured row adds to no token sum and an unpriced row adds to no
// dollar sum -- absence is never zero -- and both surface as the summary's
// counts instead.
type RoleTaskCost struct {
	Role          string
	Task          string
	Dispatches    int
	Input         int64
	Output        int64
	CacheRead     int64
	CacheCreation int64
	CostUSD       float64
}

// CostSummary is CostByRoleTask's result: one RoleTaskCost per (role, task)
// group, in the rows' seq order of first appearance, plus the ledger-level
// absence counts the section renders beneath the table.
type CostSummary struct {
	Groups     []RoleTaskCost
	Unmeasured int
	Unpriced   int
}

// CostByRoleTask derives a change's cost-per-change figure from its own
// ledger rows -- the run record the store returned, never a second query.
// It is pure: the section RenderLedger writes is exactly this summary, and
// nothing here reads or writes anything else.
func CostByRoleTask(r Run) CostSummary {
	var summary CostSummary
	index := make(map[[2]string]int)

	for _, d := range r.Dispatches {
		key := [2]string{d.Role, d.TaskID}
		i, seen := index[key]
		if !seen {
			i = len(summary.Groups)
			index[key] = i
			summary.Groups = append(summary.Groups, RoleTaskCost{Role: d.Role, Task: d.TaskID})
		}
		g := &summary.Groups[i]
		g.Dispatches++

		measured := false
		if len(bytes.TrimSpace(d.Metrics)) > 0 {
			var m dispatchMetrics
			if err := json.Unmarshal(d.Metrics, &m); err == nil && m.Tokens != nil {
				measured = true
				t := *m.Tokens
				g.Input += t.Main.Input + t.Sidechain.Input
				g.Output += t.Main.Output + t.Sidechain.Output
				g.CacheRead += t.Main.CacheRead + t.Sidechain.CacheRead
				g.CacheCreation += t.Main.CacheCreation + t.Sidechain.CacheCreation
			}
		}
		if !measured {
			summary.Unmeasured++
		}

		if d.CostUSD != nil {
			g.CostUSD += *d.CostUSD
		} else {
			summary.Unpriced++
		}
	}
	return summary
}
