package guard

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

// plan-dispatch-groups is the mechanical default for grouping a plan's
// dispatch bundles (plan-dispatch-bundles, computed in-process) into
// implementer groups, biased toward fewer, larger groups when bundles aren't
// independently parallel (KAN-513).
//
// Rule (canonical definition: this repository's
// spectre/changes/kan-513-flow-bias-free-implementer-grouping-toward/design.md,
// sections 1-2 -- do not restate the reasoning or the worked examples here).
// Over bundles B_k with members M_k (task ids) and resolved after-set A_k
// (plan-dispatch-bundles' own after-set, which already applies the serial
// default and may name a bundle's own member when one member's declared
// **After:** names a fellow bundle-mate):
//
//	deps(k) = A_k \ M_k -- what bundle k actually waits on.
//
// Two passes, both in plan order:
//
//  1. Chain merge. Walk bundles by id. Bundle k joins the first existing
//     group g such that deps(k) intersects members(g) and deps(k) is a
//     subset of members(g) union ready(g); otherwise it opens a new group.
//     Joining never changes ready(g).
//  2. Fold to at most pdgMaxInFlight (two) per ready-set -- the implementer
//     in-flight cap. Among the chain-merged groups, those sharing an
//     identical ready set are folded down to exactly two, dealt round-robin
//     in plan order (1st and 3rd together, 2nd and 4th together, ...).
//
// Arguments: exactly one tasks.md path -- the Decide step
// (skills/flow/brainstorm-planner.md, step 5) always names one, and there is
// no "every open change" scan.
//
// Exit codes mirror plan-dispatch-bundles exactly, since this computes on
// top of its output:
//
//	0  groups computed -- printed one per line on stdout, ordered by each
//	   group's lowest bundle id, as `group <g>: <bundle ids>` with the ids in
//	   plan (bundle-id) order. A file with zero unchecked tasks (zero
//	   bundles) prints nothing and still exits 0.
//	1  plan-dispatch-bundles' own violations, passed through unchanged.
//	2  invocation error -- wrong argument count, the file cannot be read, or
//	   it cannot be decoded as text.
func init() { Registry["plan-dispatch-groups"] = planDispatchGroups }

// pdgMaxInFlight is the implementer in-flight cap (skills/flow/sdd-dispatch.md's
// **Waves** paragraph): a ready-set folds to this many groups, never more.
const pdgMaxInFlight = 2

func planDispatchGroups(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: plan-dispatch-groups.sh <path-to-tasks.md>")
		return 2
	}
	bundles, afterSets, rc := pdbBundlesOf(args[0], env, stderr)
	if rc != 0 {
		return rc
	}
	for g, ids := range pdgComputeGroups(bundles, afterSets) {
		words := make([]string, len(ids))
		for i, id := range ids {
			words[i] = strconv.Itoa(id)
		}
		fmt.Fprintf(stdout, "group %d: %s\n", g+1, strings.Join(words, " "))
	}
	return 0
}

type pdgGroup struct {
	bundleIDs      []int
	members, ready map[string]bool
}

// pdgComputeGroups is compute_groups: groups as 1-based bundle ids, each in
// bundle-id order, the groups ordered by their lowest bundle id.
func pdgComputeGroups(bundles [][]pdbTask, afterSets [][]string) [][]int {
	var groups []*pdgGroup
	for i, tasks := range bundles {
		members := map[string]bool{}
		for _, t := range tasks {
			members[t.id] = true
		}
		deps := map[string]bool{}
		for _, id := range afterSets[i] {
			if !members[id] {
				deps[id] = true
			}
		}
		var joined *pdgGroup
		for _, g := range groups {
			if len(deps) == 0 {
				break
			}
			meets, covered := false, true
			for d := range deps {
				meets = meets || g.members[d]
				covered = covered && (g.members[d] || g.ready[d])
			}
			if meets && covered {
				joined = g
				break
			}
		}
		if joined != nil {
			joined.bundleIDs = append(joined.bundleIDs, i+1)
			for m := range members {
				joined.members[m] = true
			}
		} else {
			groups = append(groups, &pdgGroup{[]int{i + 1}, members, deps})
		}
	}

	// Fold pass: bucket the chain-merged groups by their ready set, in
	// first-appearance order -- already plan order.
	buckets := map[string][]*pdgGroup{}
	var order []string
	for _, g := range groups {
		ready := make([]string, 0, len(g.ready))
		for id := range g.ready {
			ready = append(ready, id)
		}
		slices.Sort(ready)
		key := strings.Join(ready, "\x00")
		if _, ok := buckets[key]; !ok {
			order = append(order, key)
		}
		buckets[key] = append(buckets[key], g)
	}
	var folded [][]int
	for _, key := range order {
		bucket := buckets[key]
		for lane := 0; lane < min(len(bucket), pdgMaxInFlight); lane++ {
			var ids []int
			for j := lane; j < len(bucket); j += pdgMaxInFlight {
				ids = append(ids, bucket[j].bundleIDs...)
			}
			slices.Sort(ids)
			folded = append(folded, ids)
		}
	}
	slices.SortStableFunc(folded, func(a, b []int) int { return a[0] - b[0] })
	return folded
}
