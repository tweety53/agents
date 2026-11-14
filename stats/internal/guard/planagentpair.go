package guard

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
)

// The contract below moved here from the header of the deleted
// scripts/plan-agent-pair.sh shim (KAN-1058).
//
// flow-guard plan-agent-pair — the model and effort of a planning session's Explore,
// Plan or planner dispatch.
//
// Usage: flow-guard plan-agent-pair <seed> <n> <explorer|plan-designer|planner> [<effort>]
//
// Prints one line, `<model> <effort>` — the first base pair of <role>,
// `explorer` for an Explore dispatch, `plan-designer` for a Plan one and
// `planner` for the planner, read with `flow policy get`, an open
// experiment's override of it applied. With <effort> — one of
// low|medium|high|xhigh, the task's effort — it prints the row's first pair
// at that effort instead, and the row's first pair when the row holds none
// at it. The planner row's pairs are `<model>/flow-planner-<effort>`,
// printed as `<model> <effort>`. <seed> is the change name — in /flow-plan
// before capture, the session token; <n> counts the session's planning
// dispatches from 1. Exit 0 on the printed line, exit 2 on a usage error
// (<n> not a positive integer, a role other than the three, a fourth
// argument other than an effort, or an argument count other than 3 or 4) or
// no pairs to read.
//
// The logic is the Go port in stats/internal/guard/planagentpair.go.

// planAgentPair is `flow-guard plan-agent-pair`: the model and effort of a
// planning session's <n>th Explore, Plan or planner dispatch -- the first pair of the
// <role>'s base pairs, explorer, plan-designer or planner, read with `flow policy
// get`, with an open experiment's override of them applied (policyTable), as
// the agent-baseline hook admits it; with an <effort>, the row's first pair
// at that effort, else its first pair.
// <seed> is the change name, or the session token in /flow-plan before
// capture; <n> counts the session's planning dispatches from 1. Exit 0 on the
// printed line, exit 2 on a usage error or an unreadable table.
func init() {
	Registry["plan-agent-pair"] = planAgentPair
}

var papPair = regexp.MustCompile(`^([a-z0-9.-]+)/(?:[A-Za-z]+@|flow-planner-)([a-z]+)$`)

// papRoles are the planning roles plan-agent-pair answers for.
var papRoles = map[string]bool{"explorer": true, "plan-designer": true, "planner": true}

// papEfforts are the efforts an optional fourth argument may name.
var papEfforts = map[string]bool{"low": true, "medium": true, "high": true, "xhigh": true}

func planAgentPair(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 && !(len(args) == 4 && papEfforts[args[3]]) {
		fmt.Fprintln(stderr, "usage: flow-guard plan-agent-pair <seed> <n> <explorer|plan-designer|planner> [low|medium|high|xhigh]")
		return 2
	}
	role := args[2]
	if !papRoles[role] {
		fmt.Fprintf(stderr, "plan-agent-pair: <role> must be explorer, plan-designer or planner, got %q\n", role)
		return 2
	}
	if n, err := strconv.Atoi(args[1]); err != nil || n < 1 {
		fmt.Fprintf(stderr, "plan-agent-pair: <n> must be a positive integer, got %q\n", args[1])
		return 2
	}
	rows, err := policyTable(env)
	if err != nil {
		fmt.Fprintf(stderr, "plan-agent-pair: %v\n", err)
		return 2
	}
	var m []string
	if rs := rows[role]; len(rs) > 0 && len(rs[0].Pairs) > 0 {
		m = papPair.FindStringSubmatch(rs[0].Pairs[0])
		if len(args) == 4 {
			for _, p := range rs[0].Pairs {
				if e := papPair.FindStringSubmatch(p); e != nil && e[2] == args[3] {
					m = e
					break
				}
			}
		}
	}
	if m == nil {
		fmt.Fprintf(stderr, "plan-agent-pair: %s names no `<model>/<agent>@<effort>` %s pair\n", mpFile, role)
		return 2
	}
	fmt.Fprintln(stdout, m[1]+" "+m[2])
	return 0
}
