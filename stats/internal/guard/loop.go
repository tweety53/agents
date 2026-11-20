package guard

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-loop-dispatch.sh shim (KAN-1058).
//
// flow-guard check-loop-dispatch <agents-repo> <key> <subagent_type> <model> [<role-id>] — open
// the row of one dispatch of a pipeline-fix or self-review fix loop, <role-id> its
// prompt's `Role:` line, recorded as the row's -policy-role when given
// (hooks/enforce-loop-dispatch.py runs it on every Agent/Task call whose
// prompt carries a `Key: pipeline-fix-<k>-…` or `Key: self-review-…` line).
// It refuses nothing: recording is best-effort, and the hook never denies.
//
// LOOP-DISPATCH-OK, exit 0: the row is opened with `flow record dispatch
// begin` under the loop's change, its session token the loop's branch name —
// except a `pass`, which is opened under the change <name> itself in the
// caller's project, its token self-review-<name>, the parent closing it when
// the pass hands back. A self-review on the change's own branch (no
// self-review-<name> branch in <agents-repo>) records in the flowd project of
// the working directory it runs in — the hook runs it in the payload's cwd.
// LOOP-DISPATCH-NA, exit 0: the key is no loop key.
// Exit 2: nothing was recorded — a loop key whose suffix is no loop suffix, or
// a pipeline-fix key with no pipeline-fix-<name>-<k> branch (flow-guard
// pipeline-fix start makes it), the begin refused, usage.
//
// Go, in flow-guard (stats/internal/guard/loop.go).

// The pipeline-fix and self-review fix loops (skills/flow-contracts/pipeline.md
// **Pipeline defects found mid-run**, skills/flow-self-review/SKILL.md step 3):
// check-loop-dispatch records each loop dispatch in flowd, and flow-guard
// pipeline-fix runs the loop's fixed steps. A loop is its own change in
// <agents repo>'s flowd project, named after its branch -- pipeline-fix-<name>-<k>
// or self-review-<name> -- except a self-review on the change's own branch,
// which has no branch of its own and files its rows under <name> itself -- in
// the flowd project of the caller's working directory, which for a change in
// another project is that project, not <agents repo>. Either way the rows
// carry the loop's branch name as their session token, so one loop's rows
// never mix with the run's own.
func init() {
	Registry["check-loop-dispatch"] = checkLoopDispatch
}

var (
	lpPipelineKey    = regexp.MustCompile(`^pipeline-fix-([0-9]+)-(.+)$`)
	lpDigits         = regexp.MustCompile(`^[0-9]+$`)
	lpPipelineBranch = regexp.MustCompile(`^pipeline-fix-(.+)-([0-9]+)$`)
	lpSuffix         = regexp.MustCompile(`^(?:pass|plan(?:-[0-9]+)?|fix(?:-[0-9]+|-g[0-9]+|-v[0-9]+)?|review-[0-9]+)$`)
)

// lpRole is the allowed-pairs role id (skills/flow-contracts/model-policy.md)
// of a loop key's suffix under prefix (self-review or pipeline-fix), "" when
// the suffix is no loop suffix.
func lpRole(prefix, suffix string) string {
	if !lpSuffix.MatchString(suffix) {
		return ""
	}
	fam := lpFamily(suffix)
	switch {
	case strings.HasPrefix(fam, "fix"):
		fam = "fix"
	case suffix == "review-1":
		fam = "review-1"
	case fam == "review":
		fam = "review-r"
	}
	return prefix + "-" + fam
}

// lpLoop names one loop's records: the change they are filed under, the
// session token (the loop's branch name), the key prefix of its rows, and
// main, the main checkout every flow call and git config read for the loop
// runs with -C.
type lpLoop struct {
	change, token, prefix, main string
}

// lpEffort is the effort a subagent_type names (flow-low, flow-planner-medium), "default" for any other.
func lpEffort(typ string) string {
	effort := strings.TrimPrefix(strings.TrimPrefix(typ, "flow-"), "planner-")
	if !slices.Contains([]string{"low", "medium", "high"}, effort) {
		return "default"
	}
	return effort
}

// lpMain is <repo>'s main checkout, the -C every flow call runs with.
func lpMain(env Env, git, repo string) (string, bool) {
	rc, out := kwRun(env, "", git, "-C", repo, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if rc != 0 {
		return "", false
	}
	return filepath.Dir(strings.TrimSpace(out)), true
}

// lpBranchExists is whether refs/heads/<b> exists in repo.
func lpBranchExists(env Env, git, repo, b string) bool {
	rc, _ := kwRun(env, "", git, "-C", repo, "rev-parse", "-q", "--verify", "refs/heads/"+b)
	return rc == 0
}

// lpFromKey splits key into its loop, its suffix and the suffix's role id.
// ok false with a message is a refusal (exit 1); ok false with msg "" means
// the key is no loop key at all.
func lpFromKey(env Env, git, repo, key string) (lpLoop, string, string, string, bool) {
	if m := lpPipelineKey.FindStringSubmatch(key); m != nil {
		k, suffix := m[1], m[2]
		role := lpRole("pipeline-fix", suffix)
		if role == "" {
			return lpLoop{}, "", "", fmt.Sprintf("KEY — %s: suffix %q is no loop suffix", key, suffix), false
		}
		_, out := kwRun(env, "", git, "-C", repo, "for-each-ref", "--format=%(refname:short)", "refs/heads/")
		var hits []string
		for _, b := range strings.Fields(out) {
			if strings.HasPrefix(b, "pipeline-fix-") && strings.HasSuffix(b, "-"+k) && plainChangeName(b) {
				hits = append(hits, b)
			}
		}
		if len(hits) != 1 {
			return lpLoop{}, "", "", fmt.Sprintf("NOT-STARTED — %s: %d branches pipeline-fix-<name>-%s; start the loop with flow-guard pipeline-fix start <name> %s", key, len(hits), k, k), false
		}
		return lpLoop{change: hits[0], token: hits[0], prefix: "pipeline-fix-" + k + "-", main: repo}, suffix, role, "", true
	}
	if !strings.HasPrefix(key, "self-review-") {
		return lpLoop{}, "", "", "", false
	}
	rest := strings.TrimPrefix(key, "self-review-")
	// The longest name wins: a role suffix is the short tail.
	for i := len(rest) - 1; i > 0; i-- {
		if rest[i] != '-' {
			continue
		}
		name, suffix := rest[:i], rest[i+1:]
		role := lpRole("self-review", suffix)
		if role == "" || !plainChangeName(name) {
			continue
		}
		loop := lpLoop{token: "self-review-" + name, prefix: "self-review-" + name + "-", change: name, main: repo}
		if lpBranchExists(env, git, repo, loop.token) {
			loop.change = loop.token
		} else if m, ok := lpMain(env, git, env.Dir); env.Dir != "" && ok {
			// On the change's own branch the change lives in the caller's
			// project -- gymie's for a gymie run -- never necessarily <agents repo>.
			loop.main = m
		}
		return loop, suffix, role, "", true
	}
	return lpLoop{}, "", "", fmt.Sprintf("KEY — %s: its suffix is no loop suffix", key), false
}

// lpFamily is a suffix's place in the loop: plan, fix (the first fixers,
// `fix` and `fix-g<g>`), fix-r (a review's fix), fix-v (a red line's fix),
// review or pass.
func lpFamily(suffix string) string {
	switch {
	case suffix == "pass":
		return "pass"
	case suffix == "plan" || strings.HasPrefix(suffix, "plan-"):
		return "plan"
	case suffix == "fix" || strings.HasPrefix(suffix, "fix-g"):
		return "fix"
	case strings.HasPrefix(suffix, "fix-v"):
		return "fix-v"
	case strings.HasPrefix(suffix, "fix-"):
		return "fix-r"
	case strings.HasPrefix(suffix, "review-"):
		return "review"
	}
	return ""
}

func checkLoopDispatch(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "check-loop-dispatch"
	if (len(args) != 4 && len(args) != 5) || args[0] == "" || args[1] == "" {
		fmt.Fprint(stderr, "usage: flow-guard check-loop-dispatch <agents-repo> <key> <subagent_type> <model> [<role-id>]\n")
		Reason(env, "usage")
		return 2
	}
	repo, key, typ, model := smcAbs(env, args[0]), args[1], args[2], args[3]
	roleID := "" // the prompt's `Role:` line id: Role costs counts the row under it
	if len(args) == 5 {
		roleID = args[4]
	}
	git, ok := panelResolveGit(env, self, stderr)
	if !ok {
		Reason(env, "git-unresolved")
		return 2
	}
	main, ok := lpMain(env, git, repo)
	if !ok {
		fmt.Fprintf(stderr, "%s: %s is not a git checkout\n", self, repo)
		Reason(env, "not-a-git-checkout")
		return 2
	}
	lp, suffix, _, msg, ok := lpFromKey(env, git, main, key)
	if !ok {
		if msg == "" {
			fmt.Fprintf(stdout, "LOOP-DISPATCH-NA %s\n", key)
			return 0
		}
		fmt.Fprintf(stderr, "%s: not recorded: %s\n", self, msg)
		Reason(env, "key-unresolved")
		return 2
	}
	// The pass belongs to the change's own self-review stage, never to the fix
	// loop it may start: it runs in the change's session, before any
	// self-review-<name> change exists, so it is recorded under <name> in the
	// caller's project -- the row the run view reads it from.
	family, change, rowMain := lpFamily(suffix), lp.change, lp.main
	if family == "pass" {
		change, rowMain = lpPassTarget(env, git, lp)
	}
	recordRole := map[string]string{"pass": "reviewer", "plan": "planner", "fix": "implementer", "fix-r": "implementer", "fix-v": "implementer", "review": "reviewer"}[family]
	begin := []string{"record", "dispatch", "begin", "-C", rowMain, "-change", change,
		"-role", recordRole, "-model", model, "-effort", lpEffort(typ), "-key", key, "-session-token", lp.token}
	if roleID != "" {
		begin = append(begin, "-policy-role", roleID)
	}
	out, _, rc := pcFlow(env, true, begin...)
	if rc != 0 {
		fmt.Fprintf(stderr, "%s: flow record dispatch begin failed (exit %d) -- not recorded:\n%s", self, rc, out)
		fmt.Fprintf(stderr, "%s: a loop's rows need its change: flow-guard pipeline-fix start creates it\n", self)
		Reason(env, "loop-begin-failed")
		return 2
	}
	fmt.Fprintf(stdout, "LOOP-DISPATCH-OK %s\n", key)
	return 0
}

// lpPassTarget is the change and checkout the self-review pass is recorded
// under: change <name>, the checkout the caller's directory belongs to when it
// resolves, lp.main otherwise. check-loop-dispatch begins the pass here and
// flow-guard pipeline-fix end closes it here.
func lpPassTarget(env Env, git string, lp lpLoop) (change, main string) {
	change, main = strings.TrimSuffix(strings.TrimPrefix(lp.prefix, "self-review-"), "-"), lp.main
	if m, ok := lpMain(env, git, env.Dir); env.Dir != "" && ok {
		main = m
	}
	return change, main
}

// lpFromBranch is the loop a branch carries: spectre/<name> an own-branch
// self-review, pipeline-fix-<name>-<k> and self-review-<name> a loop branch.
func lpFromBranch(branch string) (lpLoop, bool) {
	if name, ok := strings.CutPrefix(branch, "spectre/"); ok && plainChangeName(name) {
		return lpLoop{change: name, token: "self-review-" + name, prefix: "self-review-" + name + "-"}, true
	}
	if m := lpPipelineBranch.FindStringSubmatch(branch); m != nil && plainChangeName(branch) {
		return lpLoop{change: branch, token: branch, prefix: "pipeline-fix-" + m[2] + "-"}, true
	}
	if name, ok := strings.CutPrefix(branch, "self-review-"); ok && plainChangeName(name) {
		return lpLoop{change: branch, token: branch, prefix: branch + "-"}, true
	}
	return lpLoop{}, false
}

func lpOwnBranch(branch string) bool { return strings.HasPrefix(branch, "spectre/") }
