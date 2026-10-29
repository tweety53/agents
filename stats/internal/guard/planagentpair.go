package guard

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
)

// planAgentPair is scripts/plan-agent-pair.sh: the model and effort of a
// planning session's <n>th Explore or Plan dispatch -- the first pair of the
// plan-designer row of the allowed-pairs table (skills/flow-contracts/
// model-policy.md), read from FLOW_GUARD_REPO_ROOT, which the shim sets.
// <seed> is the change name, or the session token in /flow-plan before
// capture; <n> counts the session's planning dispatches from 1. Exit 0 on the
// printed line, exit 2 on a usage error or an unreadable table.
func init() {
	Registry["plan-agent-pair"] = planAgentPair
}

var papPair = regexp.MustCompile(`^([a-z0-9.-]+)/[A-Za-z]+@([a-z]+)$`)

func planAgentPair(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: plan-agent-pair.sh <seed> <n>")
		return 2
	}
	if n, err := strconv.Atoi(args[1]); err != nil || n < 1 {
		fmt.Fprintf(stderr, "plan-agent-pair.sh: <n> must be a positive integer, got %q\n", args[1])
		return 2
	}
	rows, err := policyTable(env.Getenv("FLOW_GUARD_REPO_ROOT"))
	if err != nil {
		fmt.Fprintf(stderr, "plan-agent-pair.sh: %v\n", err)
		return 2
	}
	var m []string
	if rs := rows["plan-designer"]; len(rs) > 0 {
		m = papPair.FindStringSubmatch(rs[0].pairs[0])
	}
	if m == nil {
		fmt.Fprintf(stderr, "plan-agent-pair.sh: %s names no `<model>/Plan@<effort>` plan-designer pair\n", mpFile)
		return 2
	}
	fmt.Fprintln(stdout, m[1]+" "+m[2])
	return 0
}
