package guard

import (
	"fmt"
	"io"
	"strconv"
)

// planAgentPair is scripts/plan-agent-pair.sh: the model and effort of a
// planning session's <n>th Explore or Plan dispatch -- `sonnet low` when
// sha256("<seed>plan<n>") mod 100 < 40, else `opus low`, rolled by
// planclass.go's pcRoll so a run is reproducible and auditable. <seed> is the
// change name, or the session token in /flow-plan before capture; <n> counts
// the session's planning dispatches from 1. Exit 0 on the printed line, exit
// 2 on a usage error. Rule: kan-1033 design.md §5, decision low-effort-40-60.
func init() {
	Registry["plan-agent-pair"] = planAgentPair
}

func planAgentPair(args []string, _ Env, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: plan-agent-pair.sh <seed> <n>")
		return 2
	}
	if n, err := strconv.Atoi(args[1]); err != nil || n < 1 {
		fmt.Fprintf(stderr, "plan-agent-pair.sh: <n> must be a positive integer, got %q\n", args[1])
		return 2
	}
	if pcRoll(args[0]+"plan"+args[1]) < 40 {
		fmt.Fprintln(stdout, "sonnet low")
	} else {
		fmt.Fprintln(stdout, "opus low")
	}
	return 0
}
