package guard

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/plan-dispatch-bundles.sh shim (KAN-1058).
//
// This file exists so an operator's muscle memory invoking this exact
// filename, and skills/flow/implement.md's own call site, keep working.
// Scope is one file per call:
//
//   - no arguments: scan every live change's tasks.md under
//     <spec-root>/changes/*/tasks.md (never the archive/ entry: run 1
//     drops a change's directory, so nothing in flight sits there),
//     aggregating exit codes — non-zero if ANY file reports a violation or
//     cannot be read, printing each file's own bundles or violations;
//   - one argument: treat it as an explicit tasks.md path and scan only
//     that file.
//
// The no-argument scan's root is PLAN_DISPATCH_BUNDLES_ROOT when set (a test
// harness's sandboxed fixture tree), else FLOW_GUARD_REPO_ROOT — exported by scripts/flow-guard — as the checkout whose changes are
// scanned.

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
// A joined bundle heavier than pdgMaxWeight (a `big` task weighs 3, every
// other task 1) is then cut, in plan order, into consecutive bundles of at
// most that weight, so no implementer receives more (kan-1091). A run of
// tasks a **Squash-with:** pair straddles is never split -- the pair commits
// together -- and such a run heavier than the cap is a bundle of its own.
// Every cut bundle after the first also waits on every task of the earlier
// ones, since they share files.
//
// Arguments: one tasks.md path computes that file; none scans every live
// change's tasks.md under <root>/<spec leaf>/changes/*/tasks.md,
// aggregating exit codes -- 2 beats 1 beats 0 -- and printing
// each file's own bundles or violations. <root> is PLAN_DISPATCH_BUNDLES_ROOT
// when set (a test's sandboxed fixture tree), else FLOW_GUARD_REPO_ROOT, which scripts/flow-guard exports from its own real location.
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
//	   value: <value>` -- or a **Size:** value other than `small`,
//	   `medium` or `big` -- `<path>:<task line>: task <id> has an unknown
//	   **Size:** value: <value>` -- every missing field first, then every
//	   malformed value, then every unknown size, never stopping at the first;
//	   or, with every field well formed, bundles that wait on each other --
//	   a bundle's after-set naming a task of another bundle whose after-set
//	   leads back to it -- printed as one line `<path>:<task line>: bundles
//	   <a> (tasks <ids>) and <b> (task <ids>) wait on each other — joined by
//	   shared **Files:** path <p>[, <p>…]; move it to **Allowed-collateral:**
//	   or reorder the tasks`, the first cycle a depth-first search in bundle
//	   order finds (kan-1090 self-review, F1).
//	2  invocation error -- wrong argument count, the file cannot be read, or
//	   it cannot be decoded as text.
//
// Every piece of tasks.md grammar this guard reads is planparse.go's
// pdbParseTasks, which plan-dispatch-groups reads through too.
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
		fmt.Fprintln(stderr, "usage: flow-guard plan-dispatch-bundles [path-to-tasks.md]")
		Reason(env, "usage")
		return 2
	}
	root := env.Getenv("PLAN_DISPATCH_BUNDLES_ROOT")
	if root == "" {
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprintln(stderr, "plan-dispatch-bundles: FLOW_GUARD_REPO_ROOT is unset — run flow-guard plan-dispatch-bundles, which sets it")
			Reason(env, "repo-root-unset")
			return 2
		}
	}
	files, ok := tcfScanPlans(env, root, stderr)
	if !ok {
		Reason(env, "scan-failed")
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
		Reason(env, "cannot-read-file")
		return nil, nil, 2
	}
	bundles, afterSets, violations := pdbComputeBundles(pdbParseTasks(lines))
	for _, v := range violations {
		fmt.Fprintf(stderr, "%s:%s\n", path, v)
	}
	if len(violations) > 0 {
		Reason(env, "bundle-violations")
		return nil, nil, 1
	}
	return bundles, afterSets, 0
}

// pdbComputeBundles is compute_bundles: bundles in lowest-id order, members
// in plan order, each with its resolved after-set; violations instead when
// any unchecked task lacks **Files:**, carries a malformed **After:**, or
// declares an unknown **Size:**.
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
	for _, t := range unchecked {
		if t.sizeDeclared && !slices.Contains(pdbSizes, t.size) {
			violations = append(violations, fmt.Sprintf("%d: task %s has an unknown **Size:** value: %s", t.taskLine, t.id, t.size))
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
	// Cut every bundle heavier than the cap (kan-1091); prior maps a cut
	// bundle's first member to the tasks of its earlier siblings.
	prior := map[string][]string{}
	var cut [][]pdbTask
	for _, members := range bundles {
		var earlier []string
		for _, c := range pdbCutBundle(members) {
			cut = append(cut, c)
			prior[c[0].id] = slices.Clone(earlier)
			for _, m := range c {
				earlier = append(earlier, m.id)
			}
		}
	}
	bundles = cut
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
		acc := prior[members[0].id]
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
	if v := pdbBundleCycle(bundles, afterSets); v != "" {
		return nil, nil, []string{v}
	}
	return bundles, afterSets, nil
}

// pdbCutBundle cuts one bundle, its members in plan order, into consecutive
// chunks of weight at most pdgMaxWeight. A segment -- the shortest run of
// members no **Squash-with:** pair straddles -- is never split, and a segment
// heavier than the cap is a chunk of its own.
func pdbCutBundle(members []pdbTask) [][]pdbTask {
	if pdgWeight(members) <= pdgMaxWeight {
		return [][]pdbTask{members}
	}
	pos := map[string]int{}
	for i, m := range members {
		pos[m.id] = i
	}
	reach := make([]int, len(members)) // the last member a segment opened here must hold
	for i, m := range members {
		reach[i] = max(reach[i], i)
		for _, p := range m.partners {
			if j, ok := pos[p]; ok {
				lo := min(i, j)
				reach[lo] = max(reach[lo], i, j)
			}
		}
	}
	var chunks [][]pdbTask
	var chunk []pdbTask
	weight, start, end := 0, 0, 0
	for i := range members {
		if end = max(end, reach[i]); end > i {
			continue
		}
		seg := members[start : i+1]
		w := pdgWeight(seg)
		if chunk != nil && weight+w > pdgMaxWeight {
			chunks = append(chunks, chunk)
			chunk, weight = nil, 0
		}
		chunk, weight = append(chunk, seg...), weight+w
		start = i + 1
	}
	return append(chunks, chunk)
}

// pdbBundleCycle searches the bundle graph -- bundle k waits on every bundle
// holding a task of k's after-set that is not k's own -- depth-first in
// bundle order, and
// renders the first cycle as a violation, or "" when there is none
// (kan-1090 self-review, F1).
func pdbBundleCycle(bundles [][]pdbTask, afterSets [][]string) string {
	owner := map[string]int{}
	for i, members := range bundles {
		for _, m := range members {
			owner[m.id] = i
		}
	}
	deps := make([][]int, len(bundles))
	for i, after := range afterSets {
		for _, id := range after {
			if j, ok := owner[id]; ok && j != i && !slices.Contains(deps[i], j) {
				deps[i] = append(deps[i], j)
			}
		}
		slices.Sort(deps[i])
	}
	state := make([]int, len(bundles)) // 0 unseen, 1 on stack, 2 done
	var stack, cycle []int
	var visit func(i int) bool
	visit = func(i int) bool {
		state[i] = 1
		stack = append(stack, i)
		for _, j := range deps[i] {
			if state[j] == 1 {
				cycle = slices.Clone(stack[slices.Index(stack, j):])
				return true
			}
			if state[j] == 0 && visit(j) {
				return true
			}
		}
		stack = stack[:len(stack)-1]
		state[i] = 2
		return false
	}
	for i := range bundles {
		if state[i] == 0 && visit(i) {
			break
		}
	}
	if cycle == nil {
		return ""
	}
	slices.Sort(cycle)
	var names, paths []string
	for _, i := range cycle {
		ids := make([]string, len(bundles[i]))
		seen := map[string]bool{}
		for j, m := range bundles[i] {
			ids[j] = m.id
			for _, p := range m.files {
				if seen[p] && !slices.Contains(paths, p) {
					paths = append(paths, p)
				}
				seen[p] = true
			}
		}
		word := "tasks"
		if len(ids) == 1 {
			word = "task"
		}
		names = append(names, fmt.Sprintf("%d (%s %s)", i+1, word, strings.Join(ids, " ")))
	}
	slices.Sort(paths)
	first := bundles[cycle[0]][0]
	for _, m := range bundles[cycle[0]][1:] {
		if pdbIDCompare(m.id, first.id) < 0 {
			first = m
		}
	}
	list := strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
	return fmt.Sprintf("%d: bundles %s wait on each other — joined by shared **Files:** path %s; move it to **Allowed-collateral:** or reorder the tasks",
		first.taskLine, list, strings.Join(paths, ", "))
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
