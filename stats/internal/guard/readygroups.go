package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

// flow-guard ready-groups <tasks.md> <worktree> <change-name>
//
// Prints the implementer groups the parent launches now, computed from the
// plan, the newest recorded decision and the dispatch rows -- never from the
// parent's reading of `after <k>:` lines.
//
// Groups are the `groups[].tasks` of the decision with the highest id,
// numbered from 1 in recorded order. A task is landed when its task line is
// ticked or an ended `implementer` row with outcome `completed` and key
// exactly `task-<ids>-implementer` or `task-<ids>-implementer-retry` names it.
// An open row with such a key whose every task the plan ticks is a returned
// implementer whose end has not reached the store yet (close-task exit 3): it
// is neither in flight nor holds its repository. A row is in flight when it has
// no endedAt, role `implementer` and a key opening `task-<ids>-implementer`
// (a fix, re-commit, pick or retry row included); its repository is that of
// the first of its ids the plan carries, and a row naming no task of the plan
// is ignored. A group is ready when it is not landed, no in-flight row names
// one of its tasks, and every id of each task's resolved after-set (its
// declared **After:** ids, else every plan-earlier task) outside the group
// itself is landed.
//
// The budget is pdgMaxInFlight less the in-flight rows. A repository with an
// in-flight row launches nothing; the ready groups of the others are taken
// round-robin in the order the plan first names the repositories until the
// budget is spent. One line per launched group, ordered by group number:
// `launch <g>: <id>+<id>… <repository> direct|throwaway`, <repository> the
// heading or `-`, `direct` when the group is alone in its repository.
//
// Exit codes:
//
//	0  a launch set printed; or none while a row is in flight or every group
//	   is landed (stderr says which).
//	1  no row in flight, no group ready, some group unlanded: one stderr line
//	   per such group naming its unlanded dependencies.
//	2  invocation error -- a wrong argument count, an unreadable plan,
//	   unreadable rows or decisions, or no usable groups.
func init() { Registry["ready-groups"] = readyGroups }

var (
	rgKeyRE     = regexp.MustCompile(`^task-([^-]+)-implementer`)
	rgFirstPass = regexp.MustCompile(`^task-[^-]+-implementer(-retry)?$`)
)

type rgRow struct {
	Key     string  `json:"key"`
	Role    string  `json:"role"`
	Outcome string  `json:"outcome"`
	EndedAt *string `json:"endedAt"`
}

type rgGroup struct {
	n     int
	ids   []string
	repo  string
	ready bool
	waits []string
}

func readyGroups(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: flow-guard ready-groups <tasks.md> <worktree> <change-name>")
		Reason(env, "usage")
		return 2
	}
	lines, err := tcfReadPlan(env, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", args[0], err)
		Reason(env, "cannot-read-file")
		return 2
	}
	var rows []rgRow
	var decisions []struct {
		ID       int64 `json:"id"`
		Decision struct {
			Groups []struct {
				Tasks []json.RawMessage `json:"tasks"`
			} `json:"groups"`
		} `json:"decision"`
	}
	raw, err := pfdRead(env, env.Dispatches, "dispatches", args[2], args[1])
	if err == nil {
		err = json.Unmarshal(raw, &rows)
	}
	if err == nil {
		raw, err = pfdRead(env, env.Decisions, "decisions", args[2], args[1])
	}
	if err == nil {
		err = json.Unmarshal(raw, &decisions)
	}
	if err != nil {
		fmt.Fprintf(stderr, "ready-groups: cannot read the records of %s: %v\n", args[2], err)
		Reason(env, "records-unreadable")
		return 2
	}
	best := -1
	for i, d := range decisions {
		if best < 0 || d.ID > decisions[best].ID {
			best = i
		}
	}
	if best < 0 || len(decisions[best].Decision.Groups) == 0 {
		fmt.Fprintf(stderr, "ready-groups: %s has no recorded groups\n", args[2])
		Reason(env, "no-groups")
		return 2
	}

	tasks := pdbParseTasks(lines)
	repoAt := pdgRepositories(env, args[0], lines)
	index := map[string]int{}
	for i, t := range tasks {
		if _, ok := index[t.id]; !ok {
			index[t.id] = i
		}
	}
	landed := map[string]bool{}
	ticked := map[string]bool{}
	for _, t := range tasks {
		if !t.unchecked {
			landed[t.id] = true
			ticked[t.id] = true
		}
	}
	inFlight := map[string]int{} // repository -> open rows
	busy := map[string]bool{}    // task id -> named by an open row
	open := 0
	for _, r := range rows {
		m := rgKeyRE.FindStringSubmatch(r.Key)
		if r.Role != "implementer" || m == nil {
			continue
		}
		ids := strings.Split(m[1], "+")
		firstPass := rgFirstPass.MatchString(r.Key)
		if r.EndedAt != nil {
			if r.Outcome == "completed" && firstPass {
				for _, id := range ids {
					landed[id] = true
				}
			}
			continue
		}
		if firstPass && !slices.ContainsFunc(ids, func(id string) bool { return !ticked[id] }) {
			continue // returned; its end has not reached the store
		}
		first := -1
		for _, id := range ids {
			if i, ok := index[id]; ok {
				first = i
				break
			}
		}
		if first < 0 {
			continue
		}
		open++
		inFlight[repoAt[tasks[first].taskLine]]++
		for _, id := range ids {
			busy[id] = true
		}
	}

	var groups []*rgGroup
	for g, grp := range decisions[best].Decision.Groups {
		if len(grp.Tasks) == 0 {
			fmt.Fprintf(stderr, "ready-groups: group %d names no tasks\n", g+1)
			Reason(env, "no-groups")
			return 2
		}
		gr := &rgGroup{n: g + 1}
		own := map[string]bool{}
		for _, r := range grp.Tasks {
			id := strings.Trim(string(r), `"`)
			gr.ids = append(gr.ids, id)
			own[id] = true
		}
		all := true
		anyBusy := false
		for _, id := range gr.ids {
			i, ok := index[id]
			if !ok {
				continue
			}
			t := tasks[i]
			if gr.repo == "" {
				gr.repo = repoAt[t.taskLine]
			}
			all = all && landed[id]
			anyBusy = anyBusy || busy[id]
			deps := t.afterSet
			if !t.afterDeclared {
				deps = nil
				for _, e := range tasks[:i] {
					deps = append(deps, e.id)
				}
			}
			for _, d := range deps {
				if !own[d] && !landed[d] && !slices.Contains(gr.waits, d) {
					gr.waits = append(gr.waits, d)
				}
			}
		}
		if all {
			continue // landed
		}
		gr.ready = !anyBusy && len(gr.waits) == 0
		groups = append(groups, gr)
	}
	if len(groups) == 0 {
		fmt.Fprintln(stderr, "ready-groups: nothing to launch — every group landed")
		return 0
	}

	// Repositories in the order the plan first names them.
	var order []string
	for _, t := range tasks {
		if r := repoAt[t.taskLine]; !slices.Contains(order, r) {
			order = append(order, r)
		}
	}
	byRepo := map[string][]*rgGroup{}
	anyReady := false
	for _, g := range groups {
		if g.ready && inFlight[g.repo] == 0 {
			byRepo[g.repo] = append(byRepo[g.repo], g)
		}
		anyReady = anyReady || g.ready
	}
	var launch []*rgGroup
	for round := 0; len(launch) < pdgMaxInFlight-open; round++ {
		took := false
		for _, r := range order {
			if round < len(byRepo[r]) && len(launch) < pdgMaxInFlight-open {
				launch = append(launch, byRepo[r][round])
				took = true
			}
		}
		if !took {
			break
		}
	}
	if len(launch) == 0 {
		if open > 0 {
			fmt.Fprintf(stderr, "ready-groups: nothing to launch — %d in flight\n", open)
			return 0
		}
		if !anyReady {
			for _, g := range groups {
				fmt.Fprintf(stderr, "ready-groups: group %d waits on task %s\n", g.n, strings.Join(g.waits, "+"))
			}
			Reason(env, "no-group-can-launch")
			return 1
		}
		return 0
	}
	slices.SortFunc(launch, func(a, b *rgGroup) int { return a.n - b.n })
	perRepo := map[string]int{}
	for _, g := range launch {
		perRepo[g.repo]++
	}
	for _, g := range launch {
		mode, repo := "throwaway", g.repo
		if perRepo[g.repo] == 1 {
			mode = "direct"
		}
		if repo == "" {
			repo = "-"
		}
		fmt.Fprintf(stdout, "launch %d: %s %s %s\n", g.n, strings.Join(g.ids, "+"), repo, mode)
	}
	return 0
}
