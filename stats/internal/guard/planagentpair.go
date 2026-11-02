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
// flow-guard plan-agent-pair — the model and effort of a planning session's Explore or
// Plan dispatch.
//
// Usage: flow-guard plan-agent-pair <seed> <n> <explorer|plan-designer>
//
// Prints one line, `<model> <effort>` — the first pair of the <role> row,
// `explorer` for an Explore dispatch and `plan-designer` for a Plan one, of
// **Allowed pairs** (skills/flow-contracts/model-policy.md), an open
// experiment's override of that row applied. <seed> is the change name — in
// /flow-plan before capture, the session token; <n> counts the session's
// planning dispatches from 1. Exit 0 on the printed line, exit 2 on a usage
// error (<n> not a positive integer, a role other than the two, or an
// argument count other than 3) or an unreadable table.
//
// The logic is the Go port in stats/internal/guard/planagentpair.go; flow-guard
// is built from this checkout by scripts/lib/flow-guard.sh, never taken from
// PATH.

// planAgentPair is scripts/plan-agent-pair.sh: the model and effort of a
// planning session's <n>th Explore or Plan dispatch -- the first pair of the
// <role> row, explorer or plan-designer, of the allowed-pairs table
// (skills/flow-contracts/model-policy.md), read from FLOW_GUARD_REPO_ROOT,
// which the shim sets, with an open experiment's override of that row applied
// (policyTable), as the agent-baseline hook admits it.
// <seed> is the change name, or the session token in /flow-plan before
// capture; <n> counts the session's planning dispatches from 1. Exit 0 on the
// printed line, exit 2 on a usage error or an unreadable table.
func init() {
	Registry["plan-agent-pair"] = planAgentPair
}

var papPair = regexp.MustCompile(`^([a-z0-9.-]+)/[A-Za-z]+@([a-z]+)$`)

func planAgentPair(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: flow-guard plan-agent-pair <seed> <n> <explorer|plan-designer>")
		return 2
	}
	role := args[2]
	if role != "explorer" && role != "plan-designer" {
		fmt.Fprintf(stderr, "plan-agent-pair.sh: <role> must be explorer or plan-designer, got %q\n", role)
		return 2
	}
	if n, err := strconv.Atoi(args[1]); err != nil || n < 1 {
		fmt.Fprintf(stderr, "plan-agent-pair.sh: <n> must be a positive integer, got %q\n", args[1])
		return 2
	}
	rows, err := policyTable(env, env.Getenv("FLOW_GUARD_REPO_ROOT"))
	if err != nil {
		fmt.Fprintf(stderr, "plan-agent-pair.sh: %v\n", err)
		return 2
	}
	var m []string
	if rs := rows[role]; len(rs) > 0 {
		m = papPair.FindStringSubmatch(rs[0].Pairs[0])
	}
	if m == nil {
		fmt.Fprintf(stderr, "plan-agent-pair.sh: %s names no `<model>/<agent>@<effort>` %s pair\n", mpFile, role)
		return 2
	}
	fmt.Fprintln(stdout, m[1]+" "+m[2])
	return 0
}
