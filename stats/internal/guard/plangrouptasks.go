package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// flow-guard plan-group-tasks <tasks.md> <decision.json>
//
// Maps each recorded implementer group to the task ids the context gather
// takes (skills/flow/sdd-dispatch.md): one `tasks <g>: <id>,<id>…` line per
// entry of the decision's `groups`, in order, each the task ids of its
// `bundles` in bundle order, as plan-dispatch-bundles numbers them. It reads
// the recorded groups -- an override split included -- never the mechanical
// ones plan-dispatch-groups prints, whose output stays as it is.
//
// Bundle ids number unchecked tasks only, so the mapping is right only before
// the plan's first tick; sdd-dispatch.md runs it once, at implement entry.
//
// Exit codes:
//
//	0  every group mapped, printed on stdout.
//	1  plan-dispatch-bundles' own violations, passed through unchanged.
//	2  invocation error -- a wrong argument count, an unreadable or
//	   undecodable decision, a missing, null or empty `groups`, a group with
//	   no bundles, or a bundle
//	   id the plan does not carry (named on stderr); nothing on stdout.
func init() { Registry["plan-group-tasks"] = planGroupTasks }

func planGroupTasks(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: flow-guard plan-group-tasks <tasks.md> <decision.json>")
		Reason(env, "usage")
		return 2
	}
	bundles, _, rc := pdbBundlesOf(args[0], env, stderr)
	if rc != 0 {
		return rc
	}
	raw, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", args[1], err)
		Reason(env, "cannot-read-decision")
		return 2
	}
	var decision struct{ Groups []struct{ Bundles []int } }
	if err := json.Unmarshal(raw, &decision); err != nil {
		fmt.Fprintf(stderr, "%s: cannot decode decision: %v\n", args[1], err)
		Reason(env, "decision-unparseable")
		return 2
	}
	if len(decision.Groups) == 0 {
		fmt.Fprintf(stderr, "%s: the decision records no groups\n", args[1])
		Reason(env, "no-groups")
		return 2
	}
	var out strings.Builder
	for g, group := range decision.Groups {
		if len(group.Bundles) == 0 {
			fmt.Fprintf(stderr, "%s: group %d names no bundles\n", args[1], g+1)
			Reason(env, "empty-group")
			return 2
		}
		var ids []string
		for _, k := range group.Bundles {
			if k < 1 || k > len(bundles) {
				fmt.Fprintf(stderr, "%s: group %d names bundle %d, which %s does not carry\n", args[1], g+1, k, args[0])
				Reason(env, "unknown-bundle")
				return 2
			}
			for _, task := range bundles[k-1] {
				ids = append(ids, task.id)
			}
		}
		fmt.Fprintf(&out, "tasks %d: %s\n", g+1, strings.Join(ids, ","))
	}
	fmt.Fprint(stdout, out.String())
	return 0
}
