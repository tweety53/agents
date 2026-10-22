package guard

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// plan-dispatch-bundles groups a tasks.md's unchecked tasks into dispatch
// bundles by the overlap of their declared **Files:** paths.
//
// Rule (canonical definition: this repository's archived kan-109 change's
// dispatch-economy requirement "Requirement: Implementer dispatches are
// bundled by declared file overlap" -- do not restate it here). Two
// unchecked tasks join the same bundle when they declare any common path
// under **Files:**, or when one carries a `**Squash-with:**` field naming the
// other (the build-green contract's "dispatched together as a single unit",
// skills/flow-contracts/build-green.md); the relation is transitive. An
// **Allowed-collateral:** glob never joins two tasks -- it names paths a
// legitimate sweep may also touch, not paths the task owns.
//
// Arguments: one tasks.md path computes that file; none scans every live
// change's tasks.md under <root>/<spec leaf>/changes/*/tasks.md, and that of
// every archived change the base does not carry yet (unlanded-archives, run
// in-process), aggregating exit codes -- 2 beats 1 beats 0 -- and printing
// each file's own bundles or violations. <root> is PLAN_DISPATCH_BUNDLES_ROOT
// when set (a test's sandboxed fixture tree), else FLOW_GUARD_REPO_ROOT,
// which scripts/plan-dispatch-bundles.sh exports as its own REAL parent's
// parent: the shim is reachable from <repo>/scripts/ and from a
// skills/<name>/scripts/ symlink, and going up from the second lands on the
// skill directory, which exists -- a scan of it would return a confident
// empty answer rather than an error.
//
// Exit codes:
//
//	0  bundles computed -- printed on stdout ordered by each bundle's lowest
//	   task id, as `bundle <k>: <ids>` with the ids in plan (document)
//	   order, each immediately followed by `after <k>: <ids>` -- that
//	   bundle's resolved after-set, the union over its members of the
//	   `**After:**` ids they declare, every plan-order earlier task id for a
//	   member carrying no field (the serial default; checked tasks
//	   included), and nothing for a member declaring `none` -- sorted
//	   numerically, or `after <k>: none` when empty. A file with zero
//	   unchecked tasks prints nothing and still exits 0.
//	1  one or more unchecked tasks carry no **Files:** field -- printed on
//	   stderr as `<path>:<task line>: task <id> has no **Files:** field` --
//	   or an `**After:**` value that does not gate as `Task <ids>` or
//	   `none` -- `<path>:<task line>: task <id> has a malformed **After:**
//	   value: <value>` -- every missing field first, then every malformed
//	   value, never stopping at the first.
//	2  invocation error -- wrong argument count, the file cannot be read, or
//	   it cannot be decoded as text.
//
// Every piece of tasks.md grammar this guard reads is planparse.go's
// pdbParseTasks, which generate-relocation-comparison and
// plan-dispatch-groups read through too.
//
// A task is UNCHECKED when the mark on its OWN task line is a space; a
// `- [x]` task takes no part in any bundle and is never reported as missing
// a field. The step checkboxes beneath a task are body and decide nothing.
//
// Path comparison is on the literal extracted text: no filesystem
// resolution, no globbing, no normalization beyond pdbExtractPaths'
// stripping.
func init() { Registry["plan-dispatch-bundles"] = planDispatchBundles }

func planDispatchBundles(args []string, env Env, stdout, stderr io.Writer) int {
	switch len(args) {
	case 1:
		return pdbCheckFile(args[0], env, stdout, stderr)
	case 0:
	default:
		fmt.Fprintln(stderr, "usage: plan-dispatch-bundles.sh [path-to-tasks.md]")
		return 2
	}
	root := env.Getenv("PLAN_DISPATCH_BUNDLES_ROOT")
	if root == "" {
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprintln(stderr, "plan-dispatch-bundles: FLOW_GUARD_REPO_ROOT is unset — run scripts/plan-dispatch-bundles.sh, which sets it")
			return 2
		}
	}
	files, ok := tcfScanPlans(env, root, stderr)
	if !ok {
		return 2
	}
	status := 0
	for _, f := range files {
		status = psWorse(status, pdbCheckFile(f, env, stdout, stderr))
	}
	return status
}

func pdbCheckFile(path string, env Env, stdout, stderr io.Writer) int {
	bundles, afterSets, rc := pdbBundlesOf(path, env, stderr)
	if rc != 0 {
		return rc
	}
	for i, members := range bundles {
		ids := make([]string, len(members))
		for j, m := range members {
			ids[j] = m.id
		}
		after := strings.Join(afterSets[i], " ")
		if after == "" {
			after = "none"
		}
		fmt.Fprintf(stdout, "bundle %d: %s\nafter %d: %s\n", i+1, strings.Join(ids, " "), i+1, after)
	}
	return 0
}

// pdbBundlesOf is check_file plus main's refusals: the file's bundles and
// after-sets, or the exit code with its lines already on stderr.
func pdbBundlesOf(path string, env Env, stderr io.Writer) ([][]pdbTask, [][]string, int) {
	lines, err := tcfReadPlan(env, path)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", path, err)
		return nil, nil, 2
	}
	bundles, afterSets, violations := pdbComputeBundles(pdbParseTasks(lines))
	for _, v := range violations {
		fmt.Fprintf(stderr, "%s:%s\n", path, v)
	}
	if len(violations) > 0 {
		return nil, nil, 1
	}
	return bundles, afterSets, 0
}

// pdbComputeBundles is compute_bundles: bundles in lowest-id order, members
// in plan order, each with its resolved after-set; violations instead when
// any unchecked task lacks **Files:** or carries a malformed **After:**.
func pdbComputeBundles(tasks []pdbTask) ([][]pdbTask, [][]string, []string) {
	var unchecked []pdbTask
	for _, t := range tasks {
		if t.unchecked {
			unchecked = append(unchecked, t)
		}
	}
	var violations []string
	for _, t := range unchecked {
		if !t.filesPresent {
			violations = append(violations, fmt.Sprintf("%d: task %s has no **Files:** field", t.taskLine, t.id))
		}
	}
	for _, t := range unchecked {
		if t.afterMalformed != nil {
			violations = append(violations, fmt.Sprintf("%d: task %s has a malformed **After:** value: %s", t.taskLine, t.id, *t.afterMalformed))
		}
	}
	if violations != nil {
		return nil, nil, violations
	}

	// Union-find over task ids; a duplicated id is one node, as the Python's
	// dict-keyed parent map made it.
	parent := map[string]string{}
	for _, t := range unchecked {
		parent[t.id] = t.id
	}
	find := func(x string) string {
		for parent[x] != x {
			x = parent[x]
		}
		return x
	}
	union := func(a, b string) {
		if ra, rb := find(a), find(b); ra != rb {
			parent[rb] = ra
		}
	}
	byPath := map[string]string{}
	for _, t := range unchecked {
		for _, p := range t.files {
			if owner, ok := byPath[p]; ok {
				union(owner, t.id)
			} else {
				byPath[p] = t.id
			}
		}
	}
	for _, t := range unchecked {
		for _, partner := range t.partners {
			if _, ok := parent[partner]; ok {
				union(t.id, partner)
			}
		}
	}
	index := map[string]int{}
	var bundles [][]pdbTask
	for _, t := range unchecked {
		root := find(t.id)
		i, ok := index[root]
		if !ok {
			i = len(bundles)
			index[root] = i
			bundles = append(bundles, nil)
		}
		bundles[i] = append(bundles[i], t)
	}
	lowest := func(members []pdbTask) string {
		low := members[0].id
		for _, m := range members[1:] {
			if pdbIDCompare(m.id, low) < 0 {
				low = m.id
			}
		}
		return low
	}
	slices.SortStableFunc(bundles, func(a, b []pdbTask) int { return pdbIDCompare(lowest(a), lowest(b)) })

	// The serial default's "every plan-order earlier task id" scans ALL
	// tasks in document order, a checked one included: a resumed plan's done
	// tasks still precede the open ones. A duplicated id's position is its
	// last occurrence, as the Python's dict comprehension left it.
	planOrder := make([]string, len(tasks))
	position := map[string]int{}
	for i, t := range tasks {
		planOrder[i] = t.id
		position[t.id] = i
	}
	afterSets := make([][]string, len(bundles))
	for i, members := range bundles {
		var acc []string
		for _, m := range members {
			ids := m.afterSet
			if !m.afterDeclared {
				ids = planOrder[:position[m.id]]
			}
			for _, id := range ids {
				if !slices.Contains(acc, id) {
					acc = append(acc, id)
				}
			}
		}
		slices.SortStableFunc(acc, pdbIDCompare)
		afterSets[i] = acc
	}
	return bundles, afterSets, nil
}

// pdbIDCompare is _id_key's numeric order over task ids, so "10" sorts after
// "2". A dotted id (a declared After value may name one) is no integer: the
// Python crashed with a ValueError traceback on it; here it orders after
// every flat id of the same integer part, by its text.
func pdbIDCompare(a, b string) int {
	ai, aRest, _ := strings.Cut(a, ".")
	bi, bRest, _ := strings.Cut(b, ".")
	ai, bi = strings.TrimLeft(ai, "0"), strings.TrimLeft(bi, "0")
	if len(ai) != len(bi) {
		return len(ai) - len(bi)
	}
	if c := strings.Compare(ai, bi); c != 0 {
		return c
	}
	return strings.Compare(aRest, bRest)
}
