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

// The contract below moved here from the header of the deleted
// scripts/pipeline-fix.sh shim (KAN-1058).
//
// flow-guard pipeline-fix — the fixed mechanics of a pipeline-fix or self-review fix
// loop (skills/flow-contracts/pipeline.md **Pipeline defects found mid-run**,
// skills/flow-self-review/SKILL.md step 3), run by the parent, so no step is
// retyped or skipped. The loop is its own change in this repository's flowd
// project, named after its branch — pipeline-fix-<name>-<k> or
// self-review-<name> — and flow-guard check-loop-dispatch files every loop dispatch
// under it (a self-review on the change's own branch files under <name>, in
// the flowd project of the caller's working directory: end and escalate of
// such a key run from the run's own project).
//
// Usage:
//   flow-guard pipeline-fix start <name> [<k>]
//       git fetch origin; `git worktree add -b <branch> <repo>-worktrees/<branch>
//       origin/<default>`, <branch> pipeline-fix-<name>-<k> with <k>, else
//       self-review-<name>; refuses a <k> that is not one more than the highest
//       <n> of any pipeline-fix-*-<n> branch (local or remote), change flowd
//       records (`flow state list`, which must answer from its store, complete,
//       else exit 2) or directory under <repo>-worktrees/, so <k> is repo-unique; creates <worktree>/.superpowers/sdd/<branch>/ and the
//       loop's STARTED state record (flow state set); runs ## worktree setup.
//       stdout: `worktree: <path>`.
//   flow-guard pipeline-fix end <key> <outcome>
//   flow-guard pipeline-fix end <review-key> <report-file>|aborted
//       closes the dispatch <key> (flow record dispatch end), `completed` or
//       `aborted`. A review takes its reviewer's report file, and closes
//       `clean` when its `Critical: <n>` and `Important: <n>` lines are both 0,
//       `fix` otherwise; a report missing either line exits 2. An `aborted`
//       key may be dispatched again.
//   flow-guard pipeline-fix escalate <key>
//       records <key>'s no-progress escalation (git config flow.escalation,
//       value `<loop branch>:<key>`; land and drop clear the loop's), which
//       flow-guard check-loop-dispatch requires before admitting <key> on its role's
//       marked `No-progress escalation` row of **Allowed pairs**
//       (skills/flow-contracts/model-policy.md); refused unless that table
//       gives <key>'s role such a row and an earlier round of its kind
//       completed.
//   flow-guard pipeline-fix verify <worktree>
//       runs the ## lint lines and tests the branch's touched paths need,
//       against <worktree>'s own origin/HEAD: every guard line, the `cd stats` lines for
//       touched Go, the `cd stats/web` lines for a touched SPA file; go test
//       of each touched stats package, scripts/test-<n>.sh or
//       scripts/test-<n>-hook.sh of a touched script or hook, the setuptest
//       package for setup.sh, the internal/guard tests that audit every
//       script (TestNoPipeIntoEarlyExitGrep and its kin) for a touched script
//       or setup.sh, vitest of a touched SPA test or the test beside a touched
//       source. One `green:` or `red:` line per command, a red one
//       followed by its last 40 output lines. When the diff from the base
//       branch touches a `ui paths` glob of `## visual verification`, the
//       section's `verify` row runs last, resolved as flow-guard run-verify-list
//       resolves it (from a declared `regression checkout`'s worktree, a
//       `specs` row's paths in `<specs>`); a red one makes `land` refuse.
//   flow-guard pipeline-fix land <worktree>
//       the merge-and-push route: refuses any other `## default landing
//       route`; git fetch origin; a branch whose base moved is rebased onto
//       origin/<default> (a stopped rebase is aborted); verify, always, before
//       anything is pushed; check-loop-landing's refusals; `git push
//       origin <branch>:<default>`; the loop's FINISHED state record; `git
//       pull --ff-only` in the main checkout; removes the worktree and the
//       branch. stdout: `landed: <sha>`. When that pull succeeded and moved
//       a stats/ path (with stats/Makefile and stats/cmd/flowd present), `make -C <main>/stats
//       restart` runs after it: `FLOWD-RESTARTED: <main>/stats`, or
//       `FLOWD-RESTART-FAILED: make -C <main>/stats restart exited <n> — run
//       it by hand` with make's output on stderr, exit 1.
//   flow-guard pipeline-fix drop <worktree>
//       a deferred loop: the FINISHED state record with withdrawn: true, then
//       removes the worktree (--force) and the branch, landing nothing.
//
// Exit codes: 0 done; 1 refused, red, or a step failed (its output relayed);
// 2 usage or it cannot answer (no flowd answer, no git, no origin/HEAD).
//
// Go, in flow-guard (stats/internal/guard/pipelinefix.go).

// pipeline-fix: the fixed mechanics of a pipeline-fix or self-review fix
// loop (skills/flow-contracts/pipeline.md **Pipeline defects found
// mid-run**), so no step is retyped or skipped. The comment above is the contract.
func init() {
	Registry["pipeline-fix"] = pipelineFix
}

const pfUsage = `usage: flow-guard pipeline-fix start <name> [<k>]
       flow-guard pipeline-fix end <key> <outcome>
       flow-guard pipeline-fix end <review-key> <report-file>|aborted
       flow-guard pipeline-fix escalate <key>
       flow-guard pipeline-fix verify <worktree>
       flow-guard pipeline-fix land <worktree>
       flow-guard pipeline-fix drop <worktree>
`

// pfEnv is what every subcommand resolves first.
type pfEnv struct {
	git, main string
}

func pipelineFix(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "pipeline-fix"
	if len(args) == 0 {
		fmt.Fprint(stderr, pfUsage)
		return 2
	}
	repo := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if repo == "" {
		fmt.Fprintf(stderr, "%s: FLOW_GUARD_REPO_ROOT is unset — run flow-guard pipeline-fix, which sets it\n", self)
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
	pe := pfEnv{git: git, main: main}
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
		highest, ok := pfHighestK(env, pe)
		if !ok {
			fmt.Fprintf(stderr, "%s: flowd's change records could not be read in full (flow state list answered from no store, or incompletely) — <k> cannot be checked\n", self)
			return 2
		}
		if want := highest + 1; k != want {
			fmt.Fprintf(stderr, "%s: <k> is %d — one more than the highest <n> of any pipeline-fix-*-<n> branch, flowd change record or worktree — not %d\n", self, want, k)
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
	if rc := kwSetup(env, wt, self, stderr); rc != 0 {
		return rc
	}
	fmt.Fprintf(stdout, "worktree: %s\n", wt)
	return 0
}

var pfNumbered = regexp.MustCompile(`^(?:[0-9]{4}-[0-9]{2}-[0-9]{2}-)?pipeline-fix-.+-([0-9]+)$`)

// pfHighestK is the highest <n> of any pipeline-fix-*-<n> local or remote
// branch, change flowd holds a record of for the project (`flow state
// list`, FINISHED ones included), or directory under <main>-worktrees/;
// 0 when there is none. <k> one more than it is unique across the
// repository, so two loops never share a key prefix. ok false unless flowd
// answered from its store, complete: the on-disk fallback holds no
// pipeline-fix record, so it would admit a used <k>.
func pfHighestK(env Env, pe pfEnv) (int, bool) {
	var names []string
	_, refs := kwRun(env, "", pe.git, "-C", pe.main, "for-each-ref", "--format=%(refname:lstrip=2)", "refs/heads/", "refs/remotes/")
	for _, r := range strings.Fields(refs) {
		names = append(names, filepath.Base(r))
	}
	out, _, rc := pcFlow(env, false, "state", "list", "-C", pe.main)
	var list struct {
		Source   string `json:"source"`
		Complete bool   `json:"complete"`
		Records  []struct {
			Name string `json:"name"`
		} `json:"records"`
	}
	if rc != 0 || json.Unmarshal(out, &list) != nil || list.Source != "store" || !list.Complete {
		return 0, false
	}
	for _, r := range list.Records {
		names = append(names, r.Name)
	}
	entries, _ := os.ReadDir(pe.main + "-worktrees")
	for _, e := range entries {
		names = append(names, e.Name())
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
	return highest, true
}

// pfLoop resolves key's loop for end and escalate.
// role is the suffix's allowed-pairs role id, and escalatable whether the
// table gives that role a marked escalation row.
func pfLoop(env Env, pe pfEnv, key string, stderr io.Writer) (lp lpLoop, suffix string, escalatable bool, rc int) {
	rows, err := policyTable(env, pe.main)
	if err != nil {
		fmt.Fprintf(stderr, "pipeline-fix: %v\n", err)
		return lpLoop{}, "", false, 2
	}
	lp, suffix, role, msg, ok := lpFromKey(env, pe.git, pe.main, key)
	if !ok {
		if msg == "" {
			msg = key + " is no loop key"
		}
		fmt.Fprintf(stderr, "pipeline-fix: %s\n", msg)
		return lpLoop{}, "", false, 1
	}
	return lp, suffix, lpEscalatable(rows, role), 0
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
	main, change := lp.main, lp.change
	if lpFamily(suffix) == "pass" {
		change, main = lpPassTarget(env, pe.git, lp)
	}
	out, _, frc := pcFlow(env, true, "record", "dispatch", "end", "-C", main, "-change", change,
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
	lp, suffix, escalatable, rc := pfLoop(env, pe, key, stderr)
	if rc != 0 {
		return rc
	}
	if !escalatable {
		fmt.Fprintf(stderr, "pipeline-fix escalate: the allowed-pairs table admits no escalation for %s\n", key)
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

// pfGoAudits runs the stats/internal/repotest tests auditing stats/ Go source
// as a whole: TestFlowAddrDeclaration reads stats/cmd/flow, TestGitConfigPins
// stats/internal/guard.
const pfGoAudits = "cd stats && go test ./internal/repotest/ -run '^Test(GitConfigPins|FlowAddrDeclaration)$' -count=1"

// pfScriptAudits runs the stats/internal/guard tests auditing every script as
// a whole — TestNoPipeIntoEarlyExitGrep, TestShimMissingLibCannotAnswer,
// TestEveryFlowGuardShimRequiresLib and TestShimSiblingsDeclared — named after
// no script, so a touched script alone never ran them (kan-1038).
const pfScriptAudits = "cd stats && go test ./internal/guard/ -run '^Test(NoPipeIntoEarlyExitGrep|ShimMissingLibCannotAnswer|EveryFlowGuardShimRequiresLib|ShimSiblingsDeclared)$' -count=1"

// pfChecks is the lint lines and tests the touched paths need: every guard
// line of `## lint`, its `cd stats` lines only for touched Go and its
// `cd stats/web` lines only for a touched SPA file; `go test` of each
// touched package, the scripts/test-*.sh harness of a touched script, the
// stats/internal/repotest package for a touched hook, and vitest of a touched
// SPA test file or the test beside a touched source file. A touched Go file
// also selects pfGoAudits: tests that audit the whole Go tree, named after no
// Go file, which main landed red without twice (KAN-887's self-review). A
// touched script or setup.sh selects pfScriptAudits where the worktree has
// stats/internal/guard. The `## visual verification` row is not among them:
// pfVerify resolves it last through rvlVisualVerify.
func pfChecks(env Env, pe pfEnv, wt string, touched []string) ([]string, int) {
	get := siblingCmd(env, "FLOW_GUARD_PROJECT_GET", "project-get", wt, "lint")
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
	// repotest is <agents repo>'s own package; another project's worktree lacks it
	repotest := isDir(filepath.Join(wt, "stats/internal/repotest"))
	if goTouched && repotest {
		add(pfGoAudits)
	}
	if isDir(filepath.Join(wt, "stats/internal/guard")) && has(func(p string) bool { return p == "setup.sh" || strings.HasPrefix(p, "scripts/") }) {
		add(pfScriptAudits)
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
		case strings.HasPrefix(p, "hooks/") && repotest:
			add("cd stats && go test ./internal/repotest/ -count=1")
		case strings.HasPrefix(p, "scripts/"):
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

// pfVerify runs pfChecks' commands in wt, then the `## visual verification`
// `verify` row when the diff touches a `ui paths` glob, resolved as
// run-verify-list.sh resolves it (decision loop-verify-reuses-visual-row).
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
	fail := func(c string, crc int, out string) {
		red++
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) > 40 {
			lines = lines[len(lines)-40:]
		}
		fmt.Fprintf(stdout, "red: %s (exit %d)\n%s\n", c, crc, strings.Join(lines, "\n"))
	}
	report := func(c, dir string) {
		if crc, out := kwRun(env, dir, "bash", "-c", c); crc != 0 {
			fail(c, crc, out)
			return
		}
		fmt.Fprintf(stdout, "green: %s\n", c)
	}
	for _, c := range cmds {
		report(c, wt)
	}
	if cmd, specs, dir, _, ok := rvlVisualVerify(env, wt, false); ok {
		switch {
		case cmd == "":
			red++
			fmt.Fprint(stdout, "red: visual verify — no merge base to scope the specs row to\n")
		case specs != "":
			src, out := kwRun(env, dir, "bash", "-c", specs)
			paths := strings.Fields(out)
			switch {
			case src != 0:
				fail(specs, src, out)
			case len(paths) == 0:
				fmt.Fprintf(stdout, "green: %s (mapped no spec)\n", specs)
			default:
				report(strings.ReplaceAll(cmd, "<specs>", strings.Join(paths, " ")), dir)
			}
		default:
			report(cmd, dir)
		}
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
			if strings.TrimSpace(out) == "" {
				out = fmt.Sprintf("git %s exited %d\n", strings.Join(a, " "), rc)
			}
			fmt.Fprint(stderr, out)
			return 1
		}
	}
	return 0
}

func pfLand(env Env, pe pfEnv, wt string, stdout, stderr io.Writer) int {
	// wt may be the cwd, and pfRemove deletes it: every command runs from
	// the main checkout, so none after the removal fails to chdir (F13).
	env.Dir = pe.main
	const self = "pipeline-fix land"
	branch, lp, ok := pfLoopWorktree(env, pe, wt, self, stderr)
	if !ok {
		return 2
	}
	get := siblingCmd(env, "FLOW_GUARD_PROJECT_GET", "project-get", wt, "default landing route")
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
	_, old := kwRun(env, "", pe.git, "-C", pe.main, "rev-parse", "HEAD")
	pulled := true
	if rc, out := kwRun(env, "", pe.git, "-C", pe.main, "pull", "--ff-only"); rc != 0 {
		pulled = false
		fmt.Fprintf(stderr, "%s: pushed, but the main checkout did not fast-forward:\n%s", self, out)
	}
	if rc := pfRemove(env, pe, wt, branch, false, stderr); rc != 0 {
		return rc
	}
	fmt.Fprintf(stdout, "landed: %s\n", strings.TrimSpace(sha))
	if pulled {
		_, now := kwRun(env, "", pe.git, "-C", pe.main, "rev-parse", "HEAD")
		if restartFlowdIfStatsMoved(env, pe.main, strings.TrimSpace(old), strings.TrimSpace(now), stdout, stderr) != 0 {
			return 1
		}
	}
	return 0
}

// restartFlowdIfStatsMoved restarts the live flowd -- `make -C <main>/stats
// restart`, which rebuilds and reinstalls the CLI too -- when old..new in the
// main checkout moved a stats/ path and both <main>/stats/Makefile and
// <main>/stats/cmd/flowd exist -- the latter so another project's stats/
// Makefile is never run (design.md: flowd-restart-on-landing). 0 nothing to do or restarted, 1 the
// restart failed, its output relayed.
func restartFlowdIfStatsMoved(env Env, main, old, new string, stdout, stderr io.Writer) int {
	diff, _ := capture(envGit(env)("-C", main, "diff", "--no-renames", "--name-only", old, new))
	moved := false
	for _, p := range strings.Split(diff, "\n") {
		moved = moved || strings.HasPrefix(p, "stats/")
	}
	dir := main + "/stats"
	if _, err := os.Stat(dir + "/Makefile"); !moved || err != nil {
		return 0
	}
	if _, err := os.Stat(dir + "/cmd/flowd"); err != nil {
		return 0
	}
	mk, ok := lookPath(env, "make")
	if !ok {
		mk = "make"
	}
	if rc, out := kwRun(env, "", mk, "-C", dir, "restart"); rc != 0 {
		fmt.Fprint(stderr, out)
		fmt.Fprintf(stdout, "FLOWD-RESTART-FAILED: make -C %s restart exited %d — run it by hand\n", dir, rc)
		return 1
	}
	fmt.Fprintf(stdout, "FLOWD-RESTARTED: %s\n", dir)
	return 0
}

func pfDrop(env Env, pe pfEnv, wt string, stdout, stderr io.Writer) int {
	// wt may be the cwd, and pfRemove deletes it: every command runs from
	// the main checkout, so none after the removal fails to chdir (F13).
	env.Dir = pe.main
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
