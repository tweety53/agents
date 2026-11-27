package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// flow-guard parallel-groups — which implementer groups of skills/flow/implement.md's **4. Execute**
// may run at once, and the merge-back of the ones that ran in copies.
//
// Usage: flow-guard parallel-groups next <canonical-worktree>
//
//	flow-guard parallel-groups merge <canonical-worktree> <copy>…
//
// next reads <canonical-worktree>/.superpowers/sdd/decision.json's `groups` and the plan at
// <worktree>/<spec root>/changes/<name>/tasks.md, <name> the worktree's spectre/<name> branch, and
// prints `GROUPS: <g>[ <g>…]` — 1-based decision indexes — or `GROUPS: none` when every group's tasks
// are ticked. The first group with an open task always heads the batch: that alone is the serial
// order. Each next group in decision order joins while all of these hold, and the batch stops at the
// first that fails one:
//   - it has an open task, and every **After:** task outside it is ticked;
//   - it and every group already in the batch declare **After:** and a non-empty **Files:** on every
//     task, no Files: entry matching the project's `ui paths` (its measure runs against the
//     canonical worktree), and no Files: entry of it equal to or under one of the batch's;
//   - the batch holds fewer than pgLimit groups.
//
// A plan sorted under more than one repository, or a canonical worktree carrying a Gradle or Maven
// build at its root, is serial: one `SERIAL: <why>` line, then the head group alone.
//
// merge applies each copy's commits — a `flow-guard throwaway-worktree create <canonical-worktree>
// <copy> --sdd` copy, on flow-copy/<basename of copy>, whose implementer committed there — onto the
// canonical worktree's branch, in argument order (plan order), by cherry-pick, printing
// `PICKED: <copy sha> <new sha>` per commit, then removes the copy with its reports folded back
// (`throwaway-worktree remove … --fold-back`). The canonical worktree's tracked files must be clean.
// A copy with uncommitted changes, or whose commits do not apply, prints `BLOCKED: <copy> — <why>`,
// aborts the cherry-pick, and keeps that copy and every later one on disk with its branch, so no
// commit is lost; copies merged before it are already removed.
//
// Exit 0 printed or merged; 1 BLOCKED; 2 usage, an unreadable plan or decision, a worktree not on
// spectre/<name>, a dirty canonical worktree, or a failed git step.
//
// Go, in flow-guard (stats/internal/guard/parallelgroups.go).
func init() { Registry["parallel-groups"] = parallelGroups }

// pgLimit is the batch cap, matching flowd's default heavy-run slot count (FLOWD_HEAVY_SLOTS).
// ponytail: a constant; read the live slot count if a project raises FLOWD_HEAVY_SLOTS.
const pgLimit = 2

// pgJVMBuild is a root file whose build the project allows one of at a time.
var pgJVMBuild = []string{"gradlew", "build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts", "pom.xml"}

const pgUsage = "usage: flow-guard parallel-groups next <canonical-worktree>\n" +
	"       flow-guard parallel-groups merge <canonical-worktree> <copy>…\n"

func parallelGroups(args []string, env Env, stdout, stderr io.Writer) int {
	switch {
	case len(args) == 2 && args[0] == "next":
		return pgNext(env, pcAbs(env, args[1]), stdout, stderr)
	case len(args) >= 3 && args[0] == "merge":
		return pgMerge(env, pcAbs(env, args[1]), args[2:], stdout, stderr)
	}
	fmt.Fprint(stderr, pgUsage)
	Reason(env, "usage")
	return 2
}

func pgFail(env Env, stderr io.Writer, id, format string, a ...any) int {
	fmt.Fprintf(stderr, "parallel-groups: "+format+"\n", a...)
	Reason(env, id)
	return 2
}

func pgNext(env Env, wt string, stdout, stderr io.Writer) int {
	branch, _ := capture(envGit(env)("-C", wt, "rev-parse", "--abbrev-ref", "HEAD"))
	name, ok := strings.CutPrefix(strings.TrimSpace(branch), "spectre/")
	if !ok {
		return pgFail(env, stderr, "not-a-change-worktree", "%s is not on a spectre/<name> branch", wt)
	}
	plan := filepath.Join(wt, specRootLeaf(wt, stderr), "changes", name, "tasks.md")
	lines, err := tcfReadPlan(env, plan)
	if err != nil {
		return pgFail(env, stderr, "plan-unreadable", "%v", err)
	}
	var d dgpvDecision
	raw, err := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "decision.json"))
	if err == nil {
		err = json.Unmarshal(raw, &d)
	}
	if err != nil {
		return pgFail(env, stderr, "decision-unreadable", "decision.json: %v", err)
	}
	cfg, _ := os.ReadFile(filepath.Join(wt, ".flow", "project.md"))
	var globs []string
	for _, g := range strings.Split(vvCells(string(cfg))["ui paths"], ",") {
		if g = strings.Trim(strings.TrimSpace(g), "`"); g != "" {
			globs = append(globs, g)
		}
	}
	ui := func(f string) bool {
		return slices.ContainsFunc(globs, func(g string) bool { return vtGlobRegexp(g).MatchString(vtBytesAsRunes(f)) })
	}
	limit := pgLimit
	tasks := pdbParseTasks(lines)
	repoAt := pdgRepositories(env, plan, lines)
	for _, t := range tasks {
		if repoAt[t.taskLine] != repoAt[tasks[0].taskLine] {
			fmt.Fprintln(stdout, "SERIAL: plan spans more than one repository")
			limit = 1
			break
		}
	}
	if limit > 1 && slices.ContainsFunc(pgJVMBuild, func(f string) bool { return isFile(filepath.Join(wt, f)) }) {
		fmt.Fprintln(stdout, "SERIAL: JVM build — one build at a time")
		limit = 1
	}
	groups := make([][]string, len(d.Groups))
	for i, g := range d.Groups {
		groups[i] = g.Tasks
	}
	batch := pgSelect(groups, tasks, ui, limit)
	if len(batch) == 0 {
		fmt.Fprintln(stdout, "GROUPS: none")
		return 0
	}
	out := make([]string, len(batch))
	for i, g := range batch {
		out[i] = fmt.Sprint(g)
	}
	fmt.Fprintln(stdout, "GROUPS: "+strings.Join(out, " "))
	return 0
}

// pgSelect is next's batch: 1-based indexes into groups, by the rules of the header.
func pgSelect(groups [][]string, tasks []pdbTask, ui func(string) bool, limit int) []int {
	byID := map[string]pdbTask{}
	for _, t := range tasks {
		byID[t.id] = t
	}
	open := func(g []string) bool {
		return slices.ContainsFunc(g, func(id string) bool { return byID[id].unchecked })
	}
	ready := func(g []string) bool {
		for _, id := range g {
			for _, a := range byID[id].afterSet {
				if !slices.Contains(g, a) && byID[a].unchecked {
					return false
				}
			}
		}
		return true
	}
	files := func(g []string) ([]string, bool) {
		var fs []string
		for _, id := range g {
			t := byID[id]
			if !t.afterDeclared || !t.filesPresent || len(t.files) == 0 || slices.ContainsFunc(t.files, ui) {
				return nil, false
			}
			for _, f := range t.files {
				fs = append(fs, strings.TrimSuffix(strings.TrimPrefix(f, "./"), "/"))
			}
		}
		return fs, true
	}
	var batch []int
	var taken []string
	for gi, g := range groups {
		if !open(g) {
			continue
		}
		fs, ok := files(g)
		if len(batch) == 0 {
			batch, taken = []int{gi + 1}, fs
			if !ok {
				return batch
			}
			continue
		}
		if !ok || !ready(g) || len(batch) >= limit || pgOverlap(fs, taken) {
			break
		}
		batch, taken = append(batch, gi+1), append(taken, fs...)
	}
	return batch
}

// pgOverlap is whether a path of a equals, or lies under, a path of b, or the other way round.
func pgOverlap(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if x == y || strings.HasPrefix(x, y+"/") || strings.HasPrefix(y, x+"/") {
				return true
			}
		}
	}
	return false
}

func pgMerge(env Env, wt string, copies []string, stdout, stderr io.Writer) int {
	git := envGit(env)
	if st, ok := capture(git("-C", wt, "status", "--porcelain", "--untracked-files=no")); !ok || strings.TrimSpace(st) != "" {
		return pgFail(env, stderr, "dirty-worktree", "%s has uncommitted tracked changes (or git status failed) — commit them first", wt)
	}
	for _, cp := range copies {
		cp = pcAbs(env, cp)
		if st, ok := capture(git("-C", cp, "status", "--porcelain")); !ok || strings.TrimSpace(st) != "" {
			fmt.Fprintf(stdout, "BLOCKED: %s — uncommitted changes (or git status failed); it and every later copy are kept\n", cp)
			Reason(env, "copy-uncommitted")
			return 1
		}
		branch := "flow-copy/" + filepath.Base(cp)
		fork, ok1 := capture(git("-C", wt, "merge-base", "HEAD", branch))
		list, ok2 := capture(git("-C", wt, "rev-list", "--reverse", strings.TrimSpace(fork)+".."+branch))
		before, ok3 := capture(git("-C", wt, "rev-parse", "HEAD"))
		if !ok1 || !ok2 || !ok3 {
			return pgFail(env, stderr, "git-failed", "cannot read %s's commits from %s", branch, wt)
		}
		picks := strings.Fields(list)
		if len(picks) > 0 {
			if out, err := git(append([]string{"-C", wt, "cherry-pick"}, picks...)...).CombinedOutput(); err != nil {
				_ = git("-C", wt, "cherry-pick", "--abort").Run()
				fmt.Fprintf(stdout, "BLOCKED: %s — its commits do not apply on %s: %s; it and every later copy are kept\n", cp, wt, strings.TrimSpace(string(out)))
				Reason(env, "cherry-pick-conflict")
				return 1
			}
			after, _ := capture(git("-C", wt, "rev-list", "--reverse", strings.TrimSpace(before)+"..HEAD"))
			for i, n := range strings.Fields(after) {
				fmt.Fprintf(stdout, "PICKED: %s %s\n", picks[i], n)
			}
		}
		if err := twRemove(env, wt, cp, true); err != nil {
			return pgFail(env, stderr, "remove-failed", "%s merged but not removed: %v", cp, err)
		}
	}
	return 0
}
