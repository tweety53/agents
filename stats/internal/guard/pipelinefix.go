package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// pipeline-fix.sh: the fixed mechanics of a pipeline-fix or self-review fix
// loop (skills/flow-contracts/pipeline.md **Pipeline defects found
// mid-run**), so no step is retyped or skipped. The shim's header is the
// contract.
func init() {
	Registry["pipeline-fix"] = pipelineFix
}

const pfUsage = `usage: pipeline-fix.sh start <name> [<k>]
       pipeline-fix.sh end <key> <outcome>
       pipeline-fix.sh end <review-key> <report-file>|aborted
       pipeline-fix.sh escalate <key>
       pipeline-fix.sh verify <worktree>
       pipeline-fix.sh land <worktree>
       pipeline-fix.sh drop <worktree>
`

// pfEnv is what every subcommand resolves first.
type pfEnv struct {
	git, main, projectGet string
}

func pipelineFix(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "pipeline-fix"
	if len(args) == 0 {
		fmt.Fprint(stderr, pfUsage)
		return 2
	}
	repo, projectGet := env.Getenv("FLOW_GUARD_REPO_ROOT"), env.Getenv("FLOW_GUARD_PROJECT_GET")
	if repo == "" || projectGet == "" {
		fmt.Fprintf(stderr, "%s: FLOW_GUARD_REPO_ROOT or FLOW_GUARD_PROJECT_GET is unset — run scripts/pipeline-fix.sh, which sets them\n", self)
		return 2
	}
	git, ok := panelResolveGit(env, self, stderr)
	if !ok {
		return 2
	}
	main, ok := lpMain(env, git, repo)
	if !ok {
		fmt.Fprintf(stderr, "%s: %s is not a git checkout\n", self, repo)
		return 2
	}
	pe := pfEnv{git: git, main: main, projectGet: projectGet}
	sub, rest := args[0], args[1:]
	switch {
	case sub == "start" && (len(rest) == 1 || len(rest) == 2):
		return pfStart(env, pe, rest, stdout, stderr)
	case sub == "end" && len(rest) == 2:
		return pfEnd(env, pe, rest[0], rest[1], stdout, stderr)
	case sub == "escalate" && len(rest) == 1:
		return pfEscalate(env, pe, rest[0], stdout, stderr)
	case sub == "verify" && len(rest) == 1:
		return pfVerify(env, pe, smcAbs(env, rest[0]), stdout, stderr)
	case sub == "land" && len(rest) == 1:
		return pfLand(env, pe, smcAbs(env, rest[0]), stdout, stderr)
	case sub == "drop" && len(rest) == 1:
		return pfDrop(env, pe, smcAbs(env, rest[0]), stdout, stderr)
	}
	fmt.Fprint(stderr, pfUsage)
	return 2
}

// pfState writes the loop change's state record (skills/flow-contracts/state-file.md).
func pfState(env Env, dir, change string, rec map[string]any) (int, string) {
	flow, ok := lookPath(env, "flow")
	if !ok {
		return 2, "no flow binary on PATH\n"
	}
	body, _ := json.Marshal(rec)
	cmd := exec.Command(flow, "state", "set", "-C", dir, change)
	cmd.Dir, cmd.Stdin = env.Dir, bytes.NewReader(body)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	return exitCode(cmd.Run()), out.String()
}

func pfRecord(state, branch string, worktrees map[string]string) map[string]any {
	return map[string]any{"state": state, "branch": branch, "worktrees": worktrees,
		"artifactUrl": nil, "jiraIssue": nil, "prUrl": nil, "updatedBy": "/flow"}
}

// pfBase is dir's origin/HEAD short name, e.g. origin/main.
func pfBase(env Env, git, dir string) (string, bool) {
	rc, out := kwRun(env, "", git, "-C", dir, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	return strings.TrimSpace(out), rc == 0 && strings.HasPrefix(out, "origin/")
}

func pfStart(env Env, pe pfEnv, rest []string, stdout, stderr io.Writer) int {
	const self = "pipeline-fix start"
	name := rest[0]
	if !plainChangeName(name) || (len(rest) == 2 && !lpDigits.MatchString(rest[1])) {
		fmt.Fprintf(stderr, "%s: <name> must be a plain change name and <k> a number\n", self)
		return 2
	}
	branch := "self-review-" + name
	if len(rest) == 2 {
		branch = "pipeline-fix-" + name + "-" + rest[1]
	}
	if rc, out := kwRun(env, "", pe.git, "-C", pe.main, "fetch", "origin"); rc != 0 {
		fmt.Fprint(stderr, out)
		return 1
	}
	base, ok := pfBase(env, pe.git, pe.main)
	if !ok {
		fmt.Fprintf(stderr, "%s: refs/remotes/origin/HEAD is unset (git remote set-head origin --auto)\n", self)
		return 2
	}
	if len(rest) == 2 {
		k, _ := strconv.Atoi(rest[1])
		if want := pfHighestK(env, pe, base) + 1; k != want {
			fmt.Fprintf(stderr, "%s: <k> is %d — one more than the highest <n> of any pipeline-fix-*-<n> branch or change folder, live or archived — not %d\n", self, want, k)
			return 1
		}
	}
	physical, err := filepath.EvalSymlinks(pe.main)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", self, err)
		return 2
	}
	wt := filepath.Join(siblingRoot(physical), branch)
	if rc, out := kwRun(env, "", pe.git, "-C", pe.main, "worktree", "add", "-b", branch, wt, base); rc != 0 {
		fmt.Fprint(stderr, out)
		return 1
	}
	// The loop's change must have a checkout before its first row: flow
	// record refuses a change no worktree carries, and spectre/changes/<loop>/
	// only exists once the planner commits.
	if err := os.MkdirAll(filepath.Join(wt, ".superpowers", "sdd", branch), 0o755); err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", self, err)
		return 1
	}
	rc, out := kwRun(env, "", pe.git, "-C", wt, "rev-parse", "HEAD")
	if rc != 0 {
		fmt.Fprint(stderr, out)
		return 1
	}
	if rc, out := pfState(env, wt, branch, pfRecord("STARTED", branch, map[string]string{wt: strings.TrimSpace(out)})); rc != 0 {
		fmt.Fprint(stderr, out)
		return 1
	}
	if rc := kwSetup(env, pe.projectGet, wt, self, stderr); rc != 0 {
		return rc
	}
	fmt.Fprintf(stdout, "worktree: %s\n", wt)
	return 0
}

var pfNumbered = regexp.MustCompile(`^(?:[0-9]{4}-[0-9]{2}-[0-9]{2}-)?pipeline-fix-.+-([0-9]+)$`)

// pfHighestK is the highest <n> of any pipeline-fix-*-<n> local or remote
// branch, or change folder under spectre/changes/ or its archive/ in base's
// tree or the main checkout; 0 when there is none. <k> one more than it is
// unique across the repository, so two loops never share a key prefix.
func pfHighestK(env Env, pe pfEnv, base string) int {
	var names []string
	_, refs := kwRun(env, "", pe.git, "-C", pe.main, "for-each-ref", "--format=%(refname:lstrip=2)", "refs/heads/", "refs/remotes/")
	for _, r := range strings.Fields(refs) {
		names = append(names, filepath.Base(r))
	}
	for _, dir := range []string{"spectre/changes/", "spectre/changes/archive/"} {
		_, tree := kwRun(env, "", pe.git, "-C", pe.main, "ls-tree", "-d", "--name-only", base, dir)
		for _, d := range strings.Fields(tree) {
			names = append(names, filepath.Base(d))
		}
		entries, _ := os.ReadDir(filepath.Join(pe.main, dir))
		for _, e := range entries {
			names = append(names, e.Name())
		}
	}
	// A dropped loop's branch is gone; drop recorded its <n> here.
	_, used := kwRun(env, "", pe.git, "-C", pe.main, "config", "--get-all", "flow.loop-k-used")
	for _, k := range strings.Fields(used) {
		names = append(names, "pipeline-fix-dropped-"+k)
	}
	highest := 0
	for _, n := range names {
		if m := pfNumbered.FindStringSubmatch(n); m != nil {
			if v, _ := strconv.Atoi(m[1]); v > highest {
				highest = v
			}
		}
	}
	return highest
}

// pfLoop resolves key's loop for end and escalate.
func pfLoop(env Env, pe pfEnv, key string, stderr io.Writer) (lpLoop, string, lpRow, int) {
	rows, err := lpTable(pe.main)
	if err != nil {
		fmt.Fprintf(stderr, "pipeline-fix: %v\n", err)
		return lpLoop{}, "", lpRow{}, 2
	}
	lp, suffix, row, msg, ok := lpFromKey(env, pe.git, pe.main, key, rows)
	if !ok {
		if msg == "" {
			msg = key + " is no loop key"
		}
		fmt.Fprintf(stderr, "pipeline-fix: %s\n", msg)
		return lpLoop{}, "", lpRow{}, 1
	}
	return lp, suffix, row, 0
}

var (
	pfCritical  = regexp.MustCompile(`(?m)^[\s*-]*Critical[\s*]*:[\s*]*([0-9]+)`)
	pfImportant = regexp.MustCompile(`(?m)^[\s*-]*Important[\s*]*:[\s*]*([0-9]+)`)
)

// pfReviewOutcome is clean or fix, derived from the review report's
// `Critical: <n>` and `Important: <n>` lines -- never the parent's word.
func pfReviewOutcome(report string) (string, error) {
	b, err := os.ReadFile(report)
	if err != nil {
		return "", err
	}
	c, i := pfCritical.FindSubmatch(b), pfImportant.FindSubmatch(b)
	if c == nil || i == nil {
		return "", fmt.Errorf("%s carries no `Critical: <n>` and `Important: <n>` lines", report)
	}
	if string(c[1]) == "0" && string(i[1]) == "0" {
		return "clean", nil
	}
	return "fix", nil
}

func pfEnd(env Env, pe pfEnv, key, outcome string, stdout, stderr io.Writer) int {
	lp, suffix, _, rc := pfLoop(env, pe, key, stderr)
	if rc != 0 {
		return rc
	}
	if lpFamily(suffix) == "review" && outcome != "aborted" {
		// clean: the review raised no Critical and no Important.
		o, err := pfReviewOutcome(outcome)
		if err != nil {
			fmt.Fprintf(stderr, "pipeline-fix end: %s takes its reviewer's report file, or aborted: %v\n", key, err)
			return 2
		}
		outcome = o
	} else if want := []string{"completed", "aborted"}; lpFamily(suffix) != "review" && !slices.Contains(want, outcome) {
		fmt.Fprintf(stderr, "pipeline-fix end: %s takes outcome %s\n", key, strings.Join(want, " or "))
		return 2
	}
	out, _, frc := pcFlow(env, true, "record", "dispatch", "end", "-C", lp.main, "-change", lp.change,
		"-key", key, "-session-token", lp.token, "-outcome", outcome)
	fmt.Fprint(stdout, string(out))
	if frc != 0 {
		return 1
	}
	// An own-branch self-review has no loop worktree for pfRemove to clear
	// its escalation markers with; its final clean review clears them.
	if outcome == "clean" && lp.change != lp.token {
		kwRun(env, "", pe.git, "-C", lp.main, "config", "--unset-all", "flow.escalation", "^"+regexp.QuoteMeta(lp.token)+":")
	}
	return 0
}

func pfEscalate(env Env, pe pfEnv, key string, stdout, stderr io.Writer) int {
	lp, suffix, row, rc := pfLoop(env, pe, key, stderr)
	if rc != 0 {
		return rc
	}
	if len(row.escal) == 0 {
		fmt.Fprintf(stderr, "pipeline-fix escalate: the pairs table admits no escalation for %s\n", key)
		return 1
	}
	rows, ok := lpRows(env, lp.main, lp)
	if !ok {
		fmt.Fprintf(stderr, "pipeline-fix escalate: flow record dispatches -change %s failed -- cannot answer\n", lp.change)
		return 2
	}
	// No progress needs an earlier round of the same kind to have made none.
	if !lpEarlierRound(lp, suffix, rows) {
		fmt.Fprintf(stderr, "pipeline-fix escalate: %s has no completed earlier round to escalate from\n", key)
		return 1
	}
	if rc, out := kwRun(env, "", pe.git, "-C", lp.main, "config", "--add", "flow.escalation", lp.token+":"+key); rc != 0 {
		fmt.Fprint(stderr, out)
		return 1
	}
	fmt.Fprintf(stdout, "escalated: %s\n", key)
	return 0
}

// pfTouched is every path wt's branch changed against base.
func pfTouched(env Env, pe pfEnv, wt, base string) ([]string, bool) {
	rc, out := kwRun(env, "", pe.git, "-C", wt, "diff", "--no-renames", "--name-only", base+"...HEAD")
	return strings.Fields(out), rc == 0
}

// pfChecks is the lint lines and tests the touched paths need: every guard
// line of `## lint`, its `cd stats` lines only for touched Go and its
// `cd stats/web` lines only for a touched SPA file; `go test` of each
// touched package, the scripts/test-*.sh harness of a touched script or
// hook, and vitest of a touched SPA test file or the test beside a touched
// source file.
func pfChecks(env Env, pe pfEnv, wt string, touched []string) ([]string, int) {
	get := exec.Command(pe.projectGet, wt, "lint")
	var body bytes.Buffer
	get.Stdout = &body
	if rc := exitCode(get.Run()); rc != 0 {
		return nil, 2
	}
	has := func(pred func(string) bool) bool { return slices.ContainsFunc(touched, pred) }
	goTouched := has(func(p string) bool { return strings.HasPrefix(p, "stats/") && strings.HasSuffix(p, ".go") })
	webTouched := has(func(p string) bool { return strings.HasPrefix(p, "stats/web/") })
	var cmds []string
	for _, l := range kwFenced(body.String()) {
		switch {
		case strings.HasPrefix(l, "cd stats/web"):
			if webTouched {
				cmds = append(cmds, l)
			}
		case strings.HasPrefix(l, "cd stats"):
			if goTouched {
				cmds = append(cmds, l)
			}
		default:
			cmds = append(cmds, l)
		}
	}
	exists := func(rel string) bool { return isFile(filepath.Join(wt, rel)) }
	add := func(c string) {
		if !slices.Contains(cmds, c) {
			cmds = append(cmds, c)
		}
	}
	for _, p := range touched {
		base := filepath.Base(p)
		stem := strings.TrimSuffix(base, filepath.Ext(base))
		switch {
		case strings.HasPrefix(p, "stats/web/"):
			rel := strings.TrimPrefix(p, "stats/web/")
			ext := filepath.Ext(p)
			sibling := strings.TrimSuffix(rel, ext) + ".test" + ext
			switch {
			case strings.Contains(base, ".test.") && exists(p):
				add("cd stats/web && npx vitest run " + rel)
			case (ext == ".ts" || ext == ".tsx") && exists("stats/web/"+sibling):
				add("cd stats/web && npx vitest run " + sibling)
			}
		case strings.HasPrefix(p, "stats/") && strings.HasSuffix(p, ".go"):
			if isDir(filepath.Join(wt, filepath.Dir(p))) {
				add("cd stats && go test ./" + strings.TrimPrefix(filepath.Dir(p), "stats/") + "/ -count=1")
			}
		case strings.HasPrefix(p, "scripts/test-") && strings.HasSuffix(p, ".sh") && exists(p):
			add(p)
		case strings.HasPrefix(p, "scripts/") || strings.HasPrefix(p, "hooks/"):
			for _, h := range []string{"scripts/test-" + stem + ".sh", "scripts/test-" + stem + "-hook.sh"} {
				if exists(h) {
					add(h)
				}
			}
		case p == "setup.sh":
			add("cd stats && go test ./internal/setuptest/ -count=1")
		}
	}
	return cmds, 0
}

func pfVerify(env Env, pe pfEnv, wt string, stdout, stderr io.Writer) int {
	// wt's own origin/HEAD: an own-branch self-review verifies a worktree of
	// the change's project, whose default branch need not be <agents repo>'s.
	base, ok := pfBase(env, pe.git, wt)
	if !ok {
		fmt.Fprint(stderr, "pipeline-fix verify: refs/remotes/origin/HEAD is unset\n")
		return 2
	}
	touched, ok := pfTouched(env, pe, wt, base)
	if !ok {
		fmt.Fprintf(stderr, "pipeline-fix verify: cannot diff %s against %s\n", wt, base)
		return 2
	}
	cmds, rc := pfChecks(env, pe, wt, touched)
	if rc != 0 {
		fmt.Fprintf(stderr, "pipeline-fix verify: project-get cannot read %s's ## lint\n", wt)
		return 2
	}
	red := 0
	for _, c := range cmds {
		crc, out := kwRun(env, wt, "bash", "-c", c)
		if crc == 0 {
			fmt.Fprintf(stdout, "green: %s\n", c)
			continue
		}
		red++
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) > 40 {
			lines = lines[len(lines)-40:]
		}
		fmt.Fprintf(stdout, "red: %s (exit %d)\n%s\n", c, crc, strings.Join(lines, "\n"))
	}
	if red > 0 {
		return 1
	}
	return 0
}

// pfLoopWorktree is wt's loop branch, or ok false (named on stderr).
func pfLoopWorktree(env Env, pe pfEnv, wt, self string, stderr io.Writer) (string, lpLoop, bool) {
	rc, out := kwRun(env, "", pe.git, "-C", wt, "branch", "--show-current")
	branch := strings.TrimSpace(out)
	lp, ok := lpFromBranch(branch)
	if rc != 0 || !ok || lpOwnBranch(branch) {
		fmt.Fprintf(stderr, "%s: %s is not on a pipeline-fix-* or self-review-* branch\n", self, wt)
		return "", lpLoop{}, false
	}
	return branch, lp, true
}

// pfRemove removes wt and its branch from the main checkout, and clears the
// loop's escalation markers.
func pfRemove(env Env, pe pfEnv, wt, branch string, force bool, stderr io.Writer) int {
	kwRun(env, "", pe.git, "-C", pe.main, "config", "--unset-all", "flow.escalation", "^"+regexp.QuoteMeta(branch)+":")
	rm := []string{"-C", pe.main, "worktree", "remove", wt}
	if force {
		rm = append(rm, "--force")
	}
	for _, a := range [][]string{rm, {"-C", pe.main, "branch", "-D", branch}} {
		if rc, out := kwRun(env, "", pe.git, a...); rc != 0 {
			fmt.Fprint(stderr, out)
			return 1
		}
	}
	return 0
}

func pfLand(env Env, pe pfEnv, wt string, stdout, stderr io.Writer) int {
	const self = "pipeline-fix land"
	branch, lp, ok := pfLoopWorktree(env, pe, wt, self, stderr)
	if !ok {
		return 2
	}
	get := exec.Command(pe.projectGet, wt, "default landing route")
	var route bytes.Buffer
	get.Stdout = &route
	_ = get.Run()
	if first, _, _ := strings.Cut(strings.TrimSpace(route.String()), "\n"); strings.TrimSpace(first) != "merge and push" {
		fmt.Fprintf(stderr, "%s: the default landing route is %q, not merge and push — land it by that route\n", self, first)
		return 1
	}
	base, ok := pfBase(env, pe.git, pe.main)
	if !ok {
		fmt.Fprintf(stderr, "%s: refs/remotes/origin/HEAD is unset\n", self)
		return 2
	}
	if rc, out := kwRun(env, "", pe.git, "-C", wt, "fetch", "origin"); rc != 0 {
		fmt.Fprint(stderr, out)
		return 1
	}
	if rc, _ := kwRun(env, "", pe.git, "-C", wt, "merge-base", "--is-ancestor", base, "HEAD"); rc != 0 {
		if rc, out := kwRun(env, "", pe.git, "-C", wt, "rebase", base); rc != 0 {
			kwRun(env, "", pe.git, "-C", wt, "rebase", "--abort")
			fmt.Fprintf(stderr, "%s: rebase onto %s stopped — resolve it in %s, then land again\n%s", self, base, wt, out)
			return 1
		}
	}
	// The lint lines and tests run on what is about to land, moved base or not.
	if rc := pfVerify(env, pe, wt, stdout, stderr); rc != 0 {
		fmt.Fprintf(stderr, "%s: red against %s — nothing pushed\n", self, base)
		return 1
	}
	if rc, out := kwRun(env, "", pe.git, "-C", wt, "ls-tree", "-d", "--name-only", "HEAD", "spectre/changes/"+branch); rc != 0 {
		fmt.Fprintf(stderr, "%s: cannot tell whether spectre/changes/%s/ is archived — nothing pushed\n%s", self, branch, out)
		return 2
	} else if strings.TrimSpace(out) != "" {
		fmt.Fprintf(stderr, "%s: spectre/changes/%s/ is not archived — commit chore(spectre): archive %s first\n", self, branch, branch)
		return 1
	}
	rows, ok := lpRows(env, pe.main, lp)
	if !ok {
		fmt.Fprintf(stderr, "%s: flow record dispatches -change %s failed -- cannot answer\n", self, lp.change)
		return 2
	}
	if refusals := lpLandingRefusals(lp, rows); len(refusals) > 0 {
		for _, r := range refusals {
			fmt.Fprintf(stdout, "LOOP-LANDING-REFUSED: %s\n", r)
		}
		return 1
	}
	target := strings.TrimPrefix(base, "origin/")
	if rc, out := kwRun(env, "", pe.git, "-C", wt, "push", "origin", branch+":"+target); rc != 0 {
		fmt.Fprint(stderr, out)
		return 1
	}
	_, sha := kwRun(env, "", pe.git, "-C", wt, "rev-parse", "HEAD")
	if rc, out := pfState(env, pe.main, lp.change, pfRecord("FINISHED", branch, map[string]string{})); rc != 0 {
		fmt.Fprintf(stderr, "%s: pushed, but the loop's state was not written:\n%s", self, out)
	}
	if rc, out := kwRun(env, "", pe.git, "-C", pe.main, "pull", "--ff-only"); rc != 0 {
		fmt.Fprintf(stderr, "%s: pushed, but the main checkout did not fast-forward:\n%s", self, out)
	}
	if rc := pfRemove(env, pe, wt, branch, false, stderr); rc != 0 {
		return rc
	}
	fmt.Fprintf(stdout, "landed: %s\n", strings.TrimSpace(sha))
	return 0
}

func pfDrop(env Env, pe pfEnv, wt string, stdout, stderr io.Writer) int {
	const self = "pipeline-fix drop"
	branch, lp, ok := pfLoopWorktree(env, pe, wt, self, stderr)
	if !ok {
		return 2
	}
	if m := pfNumbered.FindStringSubmatch(branch); m != nil {
		if rc, out := kwRun(env, "", pe.git, "-C", pe.main, "config", "--add", "flow.loop-k-used", m[1]); rc != 0 {
			fmt.Fprint(stderr, out)
			return 1
		}
	}
	rec := pfRecord("FINISHED", branch, map[string]string{})
	rec["withdrawn"] = true
	if rc, out := pfState(env, pe.main, lp.change, rec); rc != 0 {
		fmt.Fprint(stderr, out)
		return 1
	}
	if rc := pfRemove(env, pe, wt, branch, true, stderr); rc != 0 {
		return rc
	}
	fmt.Fprintf(stdout, "dropped: %s\n", branch)
	return 0
}
