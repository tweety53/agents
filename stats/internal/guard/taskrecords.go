package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// check-task-records — lint one repository's in-flight task records against
// the commits on its branch (KAN-410). Ported from check-task-records.py at
// bbfe685e; scripts/check-task-records.sh is its shim.
//
// The per-commit guard, check-task-commit-fields, checks ONE task's fields
// against ONE real commit at the moment that commit is made. Its findings in
// kan-29's self-review (F14, F1 and the malformed **Build:** lines of tasks
// 43-45) were all drift it cannot see, because it runs before the drift
// happens: a checkbox ticked with no commit behind it, a commit that touched
// files its task never declared, a Build tag mangled by a later hand edit.
// This guard is the retroactive, whole-plan pass those findings asked for:
// every live spectre/changes/*/tasks.md, plus the tasks.md of each archived
// change the base does not carry yet (unlanded-archives, called in-process)
// — every task, checked against the branch's commits since the base ref.
// Zero plans in flight is clean — the bare-tree case a lint step runs in.
//
// Usage: check-task-records.sh [repo-root] [base-ref]
//
// With no repo-root the root is the working directory. `base-ref` is the
// branch the change branched from, bare (`main`) or remote-tracking
// (`origin/main`); a bare name prefers refs/remotes/origin/<name> when that
// resolves, so a stale local branch is never compared against. Absent, it is
// the base recorded on the current branch (`branch.<current>.flowBase`, as
// resolve-base-branch.sh reads it), else the branch `refs/remotes/origin/HEAD`
// points at. The commit range is `<base-ref>..HEAD`.
//
// Three checks, per task:
//
//  1. Tick state vs commit existence. A task declaring a `**Commit:**`
//     subject that is ticked must have a commit carrying that subject
//     reachable from HEAD — a landed plan whose commits are already in the
//     base reads clean, which is the state every unarchived plan left behind
//     by a merged change sits in. An archived plan is exempt from that half:
//     integrate's split collapsed its task commits into one implementation
//     commit before it archived, so a ticked task of an archived plan whose
//     subject no commit carries is not a violation. A task that is unticked
//     must have no commit carrying its subject in `<base-ref>..HEAD` (work
//     the branch started that the record does not claim); a commit reachable
//     from HEAD but older than the base is not this branch's business and is
//     not a violation. A task declaring no subject carries no expectation and
//     is skipped here. A cross-repo change's plan lives in its canonical
//     repository while some tasks land in a peer's worktree: when a ticked
//     task's subject is reachable from no commit here and its change
//     directory carries a `link.md`, the change's other worktrees — the
//     `worktrees` map of its state record, `flow state get -C <repo-root>
//     <change>` — are searched too, and a commit found reachable from one's
//     HEAD is judged in that repository by check 2; a listed path that is not
//     a git worktree is skipped. No record, or no `flow` on PATH, searches
//     nothing and the violation stands.
//  2. Declared `**Files:**` vs the commit's real files. For each ticked
//     task's subject, the files `git diff-tree` reports for the newest
//     reachable commit carrying it are checked against the union of the
//     declaring tasks' `**Files:**` plus `**Allowed-collateral:**` globs — in
//     both directions, which is the part the per-commit guard cannot do: a
//     file the commit touched that nothing declared, and a file the record
//     declares that the commit never touched. Tasks sharing one subject (a
//     red fold: every partner declares the folded commit's agreed subject)
//     form one group keyed on that subject, so the fold's union is what the
//     commit is checked against — the same semantics resolve_folded_task
//     enforces at commit time, reached here from the subjects themselves
//     without re-implementing bundle resolution. A merge commit yields no
//     diff-tree file list and is accepted silently; a flow branch carries
//     none.
//  3. `**Build:**` tag format. First column-0 `**Build:**` line per task, via
//     the shared grammar's tcfSelectBuildTag: absent is a violation, and a
//     value opening with neither green nor red is a violation naming the
//     line. This deliberately re-states a check the plan-time lint
//     (check-task-build-green) already makes: a record drifts after plan
//     time — kan-29's tasks 43-45 are the case that motivated this guard —
//     and because both guards read the same shared grammar (planparse.go),
//     the overlap carries no second definition to drift apart.
//
// Parsing is shared, never re-stated: task lines, bodies, Build tags and
// fences, field parsing (continuation joining, backtick stripping,
// collateral globs) and the undeclared-file check are the tcf* twins
// check-task-commit-fields and check-plan-shape read.
//
// Output: one violation per line on stdout, `<tasks.md path relative to the
// repo root>: task <id> ...`, in first-come order. Exit codes:
//
//	0  clean, or nothing to check (no plans in flight, no declaring tasks)
//	1  violations found
//	2  could not judge — usage, an unreadable plan, unlanded-archives or a
//	   git call that cannot evaluate its arguments; never a verdict about a
//	   record. Every exit-2 refusal past usage and unlanded-archives prints
//	   the stderr opening "check-task-records: COULD NOT JUDGE — not a record
//	   verdict:" so a caller mistake is never read as the guard refusing a
//	   record.
//
// A plan that is not UTF-8 is exit 2: the Python let read_text()'s
// UnicodeDecodeError escape as a traceback (exit 1), and the header's "an
// unreadable plan" wins (KAN-918's header-wins-on-exit-code).
func init() { Registry["check-task-records"] = checkTaskRecords }

type trDie struct{ detail string }

type trGroup struct {
	sha, repo, rel    string
	ids               []string
	files, collateral []string
	commitCount       int
}

type trPeer struct {
	path     string
	subjects map[string][]string
}

func checkTaskRecords(args []string, env Env, stdout, stderr io.Writer) (rc int) {
	if len(args) > 2 {
		fmt.Fprintln(stderr, "usage: check-task-records.sh [repo-root] [base-ref]")
		return 2
	}
	arg := env.Dir
	if len(args) > 0 && args[0] != "" {
		arg = args[0]
	}
	var archived bytes.Buffer
	if unlandedArchivesCmd([]string{arg}, env, &archived, stderr) != 0 {
		return 2
	}
	defer func() {
		if r := recover(); r != nil {
			d, ok := r.(trDie)
			if !ok {
				panic(r)
			}
			fmt.Fprintf(stderr, "check-task-records: COULD NOT JUDGE — not a record verdict: %s\n", d.detail)
			rc = 2
		}
	}()
	violations := trCheck(args, env, archived.String())
	for _, v := range violations {
		fmt.Fprintln(stdout, v)
	}
	if len(violations) > 0 {
		return 1
	}
	return 0
}

func trFail(format string, a ...any) { panic(trDie{fmt.Sprintf(format, a...)}) }

// trResolvePath is pathlib's Path.resolve(): absolute against env.Dir, with
// symlinks resolved through the longest leading part that exists.
func trResolvePath(env Env, p string) string {
	if p == "" {
		p = env.Dir
	}
	abs, _ := filepath.Abs(psAbs(env, p))
	rest := ""
	for d := abs; ; d = filepath.Dir(d) {
		if r, err := filepath.EvalSymlinks(d); err == nil {
			return filepath.Join(r, rest)
		}
		if d == filepath.Dir(d) {
			return abs
		}
		rest = filepath.Join(filepath.Base(d), rest)
	}
}

func trCheck(args []string, env Env, archivedNames string) []string {
	var root string
	if len(args) > 0 {
		root = trResolvePath(env, args[0])
	} else {
		root = trResolvePath(env, "")
	}
	if !isDir(root) {
		trFail("repo root not found: %s", root)
	}
	git := envGit(env)

	changesDir := root + "/spectre/changes"
	archiveDir := changesDir + "/archive"
	type plan struct {
		path     string
		archived bool
	}
	var plans []plan
	if isDir(changesDir) {
		entries, _ := os.ReadDir(changesDir) // sorted by name, as sorted(iterdir()) is
		for _, e := range entries {
			p := changesDir + "/" + e.Name()
			if isDir(p) && isFile(p+"/tasks.md") {
				plans = append(plans, plan{p + "/tasks.md", false})
			}
		}
	}
	for _, name := range tcfSplitLines(archivedNames) {
		if p := archiveDir + "/" + name + "/tasks.md"; name != "" && isFile(p) {
			plans = append(plans, plan{p, true})
		}
	}
	if len(plans) == 0 {
		return nil
	}

	var base string
	if len(args) == 2 {
		base = trResolveBase(git, root, args[1])
	} else {
		base = trResolveBase(git, root, "")
	}
	headSubjects := trSubjectMap(git, root, "HEAD")
	rangeSubjects := trSubjectMap(git, root, base+"..HEAD")

	var violations []string
	// A subject's partners live in one plan, so the first declaring task's
	// plan path is the attribution for the whole group; a commit already
	// reported under one plan is never re-checked under another.
	groups := map[string]*trGroup{}
	var order []string

	for _, p := range plans {
		rel := strings.TrimPrefix(p.path, root+"/")
		var peers []trPeer
		peersRead := false
		lines, err := tcfReadPlan(env, p.path)
		if err != nil {
			trFail("cannot read %s: %v", p.path, err)
		}
		for _, task := range tcfIterTasks(lines) {
			m := tcfTaskLineRE.FindStringSubmatch(lines[task.taskLine-1])
			ticked := m != nil && m[1] == "x"
			if ticked {
				// A ticked task's steps are ticked with it (flow tasks tick
				// does both); kan-1017 ticked six tasks by hand and left
				// their 27 steps open.
				tcfUnfenced(task.lines, func(_ int, line string) bool {
					if strings.HasPrefix(line, "  - [ ] ") {
						violations = append(violations, fmt.Sprintf("%s: task %s ticked but holds an unticked step: '%s'", rel, task.id, strings.TrimSpace(line)))
						return true
					}
					return false
				})
			}

			if tag := tcfSelectBuildTag(task.lines); tag == nil {
				violations = append(violations, fmt.Sprintf("%s: task %s has no **Build:** tag", rel, task.id))
			} else if tag.kind == "" {
				violations = append(violations, fmt.Sprintf("%s: task %s has a malformed **Build:** tag: '%s'", rel, task.id, tcfStrip(tag.value)))
			}

			fields, ok := tcfParseTask(lines, task.id)
			if !ok {
				trFail("task %s found by the grammar but not by the field parser in %s", task.id, p.path)
			}
			if fields.commit == nil || *fields.commit == "" {
				continue
			}
			subject := *fields.commit
			shas := headSubjects[subject]
			if !ticked {
				if len(rangeSubjects[subject]) > 0 {
					violations = append(violations, fmt.Sprintf("%s: task %s not ticked but commit with subject '%s' exists in %s..HEAD (%s)", rel, task.id, subject, base, shas[0][:12]))
				}
				continue
			}
			repo := root
			if len(shas) == 0 && !p.archived {
				if !peersRead {
					peers, peersRead = trPeerHeads(env, git, root, filepath.Dir(p.path)), true
				}
				for _, peer := range peers {
					if s := peer.subjects[subject]; len(s) > 0 {
						repo, shas = peer.path, s
						break
					}
				}
			}
			if len(shas) == 0 {
				if p.archived {
					continue
				}
				violations = append(violations, fmt.Sprintf("%s: task %s ticked but no commit with subject '%s' reachable from HEAD", rel, task.id, subject))
				continue
			}
			g := groups[subject]
			if g == nil {
				g = &trGroup{sha: shas[0], repo: repo, rel: rel, commitCount: len(rangeSubjects[subject])}
				groups[subject] = g
				order = append(order, subject)
			}
			g.ids = trAppendNew(g.ids, task.id)
			for _, f := range fields.files {
				g.files = trAppendNew(g.files, f)
			}
			for _, c := range fields.allowedCollateral {
				g.collateral = trAppendNew(g.collateral, c)
			}
		}
	}

	for _, subject := range order {
		g := groups[subject]
		id := strings.Join(g.ids, "+")
		if g.commitCount > 1 {
			violations = append(violations, fmt.Sprintf("%s: task %s: subject '%s' is carried by %d commits in %s..HEAD; expected one", g.rel, id, subject, g.commitCount, base))
		}
		changed := trCommitFiles(git, g.repo, g.sha)
		if len(changed) == 0 {
			// A merge commit: diff-tree lists no files for it, so there is no
			// file comparison to make — it is accepted silently, as the
			// header states. A flow branch carries none.
			continue
		}
		merged := tcfTask{id: id, files: g.files, allowedCollateral: g.collateral}
		for _, v := range tcfCheckFiles(merged, changed) {
			violations = append(violations, g.rel+": "+v)
		}
		for _, path := range merged.files {
			if !slices.Contains(changed, path) {
				violations = append(violations, fmt.Sprintf("%s: task %s: declared file %s is not touched by its commit %s", g.rel, id, path, g.sha[:12]))
			}
		}
	}
	return violations
}

func trAppendNew(list []string, s string) []string {
	if slices.Contains(list, s) {
		return list
	}
	return append(list, s)
}

// trGitOut is the Python's git(): stdout, or a COULD NOT JUDGE naming the
// arguments (never the -C root) and git's stderr.
func trGitOut(git func(...string) *exec.Cmd, root string, args ...string) string {
	cmd := git(append([]string{"-C", root}, args...)...)
	var e bytes.Buffer
	cmd.Stderr = &e
	out, err := cmd.Output()
	if err != nil {
		msg := tcfStrip(e.String())
		if msg == "" {
			msg = "non-zero exit"
		}
		trFail("git %s failed: %s", strings.Join(args, " "), msg)
	}
	return string(out)
}

func trResolveBase(git func(...string) *exec.Cmd, root, baseArg string) string {
	if baseArg == "" {
		cur, _ := capture(git("-C", root, "branch", "--show-current"))
		if cur = tcfStrip(cur); cur != "" {
			rec, _ := capture(git("-C", root, "config", "--get", "branch."+cur+".flowBase"))
			baseArg = tcfStrip(rec)
		}
	}
	if baseArg != "" {
		for _, c := range []string{"refs/remotes/origin/" + baseArg, baseArg} {
			if git("-C", root, "rev-parse", "--verify", "--quiet", c).Run() == nil {
				return c
			}
		}
		trFail("base-ref does not resolve: %s", baseArg)
	}
	out, ok := capture(git("-C", root, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"))
	if !ok {
		trFail("refs/remotes/origin/HEAD does not resolve and no base-ref argument was given")
	}
	return tcfStrip(out)
}

// trSubjectMap is subject -> shas reachable from rev, newest first (`git
// log` order). A subject a record declares should identify exactly one
// commit; the NEWEST one is what the file comparison reads (the state a
// rework or re-squash left behind), and more than one is its own violation,
// raised once per subject group.
func trSubjectMap(git func(...string) *exec.Cmd, root, rev string) map[string][]string {
	subjects := map[string][]string{}
	for _, line := range tcfSplitLines(trGitOut(git, root, "log", rev, "--format=%H%x09%s")) {
		if line == "" {
			continue
		}
		sha, subject, _ := strings.Cut(line, "\t")
		subjects[subject] = append(subjects[subject], sha)
	}
	return subjects
}

// trPeerHeads is the change's other worktrees, from its state record's
// `worktrees` map, each with its HEAD's subject map — only for a cross-repo
// change, one whose directory carries a `link.md`. The state record lives in
// the records store, so the read follows `FLOW_RECORDS_ADDR` when set.
func trPeerHeads(env Env, git func(...string) *exec.Cmd, root, changeDir string) []trPeer {
	flow, ok := lookPath(env, "flow")
	if !isFile(changeDir+"/link.md") || !ok {
		return nil
	}
	cmd := exec.Command(flow, "state", "get", "-C", root, filepath.Base(changeDir))
	cmd.Dir = env.Dir
	cmd.Env = icChildEnv(env)
	if a := env.Getenv("FLOW_RECORDS_ADDR"); a != "" {
		cmd.Env = append(cmd.Env, "FLOW_ADDR="+a)
	}
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	// ponytail: only a JSON object's worktrees map is read; a record of any
	// other shape searches nothing, where the Python iterated a list too.
	var rec struct {
		Worktrees map[string]json.RawMessage `json:"worktrees"`
	}
	if json.Unmarshal(out, &rec) != nil {
		return nil
	}
	trees := make([]string, 0, len(rec.Worktrees))
	for tree := range rec.Worktrees {
		trees = append(trees, tree)
	}
	sort.Strings(trees)
	var peers []trPeer
	for _, tree := range trees {
		if _, err := os.Stat(psAbs(env, tree) + "/.git"); err == nil && trResolvePath(env, tree) != root {
			peers = append(peers, trPeer{tree, trSubjectMap(git, tree, "HEAD")})
		}
	}
	return peers
}

func trCommitFiles(git func(...string) *exec.Cmd, repo, sha string) []string {
	var files []string
	for _, l := range tcfSplitLines(trGitOut(git, repo, "diff-tree", "--no-renames", "--root", "-r", "--name-only", "--no-commit-id", sha)) {
		if l != "" {
			files = append(files, l)
		}
	}
	return files
}
