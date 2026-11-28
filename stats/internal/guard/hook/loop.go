package hook

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The contract below moved here from the header of the deleted
// scripts/check-loop-dispatch.sh shim (KAN-1058).
//
// flow-guard check-loop-dispatch <agents-repo> <key> <subagent_type> <model> [<role-id>] — open
// the row of one dispatch of a pipeline-fix or self-review fix loop, <role-id> its
// prompt's `Role:` line, recorded as the row's -policy-role when given
// (flow-guard hook enforce-loop-dispatch runs it on every Agent/Task call whose
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
// Go, in flow-guard (stats/internal/guard/hook/loop.go).

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
	core.Registry["check-loop-dispatch"] = CheckLoopDispatch
}

var (
	lpPipelineKey  = regexp.MustCompile(`^pipeline-fix-([0-9]+)-(.+)$`)
	Digits         = regexp.MustCompile(`^[0-9]+$`)
	PipelineBranch = regexp.MustCompile(`^pipeline-fix-(.+)-([0-9]+)$`)
	lpSuffix       = regexp.MustCompile(`^(?:pass|plan(?:-[0-9]+)?|fix(?:-[0-9]+|-[grv][0-9]+)?|review-r?[0-9]+)$`)
)

// LoopRole is the allowed-pairs role id (skills/flow-contracts/model-policy.md)
// of a loop key's suffix -- self-review-pass for the pass, a loop-* role
// shared by the self-review and pipeline-fix loops otherwise -- "" when the
// suffix is no loop suffix.
func LoopRole(suffix string) string {
	if !lpSuffix.MatchString(suffix) {
		return ""
	}
	fam := LoopFamily(suffix)
	switch {
	case fam == "pass":
		return "self-review-pass"
	case strings.HasPrefix(fam, "fix"):
		fam = "fix"
	case suffix == "review-1":
		fam = "review-1"
	case fam == "review":
		fam = "review-r"
	}
	return "loop-" + fam
}

// Loop names one loop's records: the change they are filed under, the
// session token (the loop's branch name), the key prefix of its rows, and
// main, the main checkout every flow call and git config read for the loop
// runs with -C.
type Loop struct {
	Change, Token, prefix, Main string
}

// lpEffort is the effort a subagent_type names (flow-low, flow-planner-medium), "default" for any other.
func lpEffort(typ string) string {
	effort := strings.TrimPrefix(strings.TrimPrefix(typ, "flow-"), "planner-")
	if !slices.Contains([]string{"low", "medium", "high", "xhigh", "max"}, effort) {
		return "default"
	}
	return effort
}

// LoopMain is <repo>'s main checkout, the -C every flow call runs with.
func LoopMain(env core.Env, git, repo string) (string, bool) {
	rc, out := core.RunCombined(env, "", git, "-C", repo, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if rc != 0 {
		return "", false
	}
	return filepath.Dir(strings.TrimSpace(out)), true
}

// BranchExists is whether refs/heads/<b> exists in repo.
func BranchExists(env core.Env, git, repo, b string) bool {
	rc, _ := core.RunCombined(env, "", git, "-C", repo, "rev-parse", "-q", "--verify", "refs/heads/"+b)
	return rc == 0
}

// LoopFromKey splits key into its loop, its suffix and the suffix's role id.
// ok false with a message is a refusal (exit 1); ok false with msg "" means
// the key is no loop key at all.
func LoopFromKey(env core.Env, git, repo, key string) (Loop, string, string, string, bool) {
	if m := lpPipelineKey.FindStringSubmatch(key); m != nil {
		k, suffix := m[1], m[2]
		role := LoopRole(suffix)
		if role == "" {
			return Loop{}, "", "", fmt.Sprintf("KEY — %s: suffix %q is no loop suffix", key, suffix), false
		}
		_, out := core.RunCombined(env, "", git, "-C", repo, "for-each-ref", "--format=%(refname:short)", "refs/heads/")
		var hits []string
		for _, b := range strings.Fields(out) {
			if strings.HasPrefix(b, "pipeline-fix-") && strings.HasSuffix(b, "-"+k) && core.PlainChangeName(b) {
				hits = append(hits, b)
			}
		}
		if len(hits) != 1 {
			return Loop{}, "", "", fmt.Sprintf("NOT-STARTED — %s: %d branches pipeline-fix-<name>-%s; start the loop with flow-guard pipeline-fix start <name> %s", key, len(hits), k, k), false
		}
		return Loop{Change: hits[0], Token: hits[0], prefix: "pipeline-fix-" + k + "-", Main: repo}, suffix, role, "", true
	}
	if !strings.HasPrefix(key, "self-review-") {
		return Loop{}, "", "", "", false
	}
	rest := strings.TrimPrefix(key, "self-review-")
	// The longest name wins: a role suffix is the short tail.
	for i := len(rest) - 1; i > 0; i-- {
		if rest[i] != '-' {
			continue
		}
		name, suffix := rest[:i], rest[i+1:]
		role := LoopRole(suffix)
		if role == "" || !core.PlainChangeName(name) {
			continue
		}
		loop := Loop{Token: "self-review-" + name, prefix: "self-review-" + name + "-", Change: name, Main: repo}
		if BranchExists(env, git, repo, loop.Token) {
			loop.Change = loop.Token
		} else if m, ok := LoopMain(env, git, env.Dir); env.Dir != "" && ok {
			// On the change's own branch the change lives in the caller's
			// project -- gymie's for a gymie run -- never necessarily <agents repo>.
			loop.Main = m
		}
		return loop, suffix, role, "", true
	}
	return Loop{}, "", "", fmt.Sprintf("KEY — %s: its suffix is no loop suffix", key), false
}

// LoopFamily is a suffix's place in the loop: plan, fix (the first fixers,
// `fix` and `fix-g<g>`), fix-r (a review's fix), fix-v (a red line's fix),
// review or pass.
func LoopFamily(suffix string) string {
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

func CheckLoopDispatch(args []string, env core.Env, stdout, stderr io.Writer) int {
	const self = "check-loop-dispatch"
	if (len(args) != 4 && len(args) != 5) || args[0] == "" || args[1] == "" {
		fmt.Fprint(stderr, "usage: flow-guard check-loop-dispatch <agents-repo> <key> <subagent_type> <model> [<role-id>]\n")
		core.Reason(env, "usage")
		return 2
	}
	repo, key, typ, model := core.Abs(env, args[0]), args[1], args[2], args[3]
	roleID := "" // the prompt's `Role:` line id: Role costs counts the row under it
	if len(args) == 5 {
		roleID = args[4]
	}
	git, ok := core.ResolveGit(env, self, stderr)
	if !ok {
		core.Reason(env, "git-unresolved")
		return 2
	}
	main, ok := LoopMain(env, git, repo)
	if !ok {
		fmt.Fprintf(stderr, "%s: %s is not a git checkout\n", self, repo)
		core.Reason(env, "not-a-git-checkout")
		return 2
	}
	lp, suffix, _, msg, ok := LoopFromKey(env, git, main, key)
	if !ok {
		if msg == "" {
			fmt.Fprintf(stdout, "LOOP-DISPATCH-NA %s\n", key)
			return 0
		}
		fmt.Fprintf(stderr, "%s: not recorded: %s\n", self, msg)
		core.Reason(env, "key-unresolved")
		return 2
	}
	// The pass belongs to the change's own self-review stage, never to the fix
	// loop it may start: it runs in the change's session, before any
	// self-review-<name> change exists, so it is recorded under <name> in the
	// caller's project -- the row the run view reads it from.
	family, change, rowMain := LoopFamily(suffix), lp.Change, lp.Main
	if family == "pass" {
		change, rowMain = PassTarget(env, git, lp)
	}
	recordRole := map[string]string{"pass": "reviewer", "plan": "planner", "fix": "implementer", "fix-r": "implementer", "fix-v": "implementer", "review": "reviewer"}[family]
	begin := []string{"record", "dispatch", "begin", "-C", rowMain, "-change", change,
		"-role", recordRole, "-model", model, "-effort", lpEffort(typ), "-key", key, "-session-token", lp.Token}
	if roleID != "" {
		begin = append(begin, "-policy-role", roleID)
	}
	out, _, rc := core.Flow(env, true, begin...)
	if rc != 0 {
		fmt.Fprintf(stderr, "%s: flow record dispatch begin failed (exit %d) -- not recorded:\n%s", self, rc, out)
		fmt.Fprintf(stderr, "%s: a loop's rows need its change: flow-guard pipeline-fix start creates it\n", self)
		core.Reason(env, "loop-begin-failed")
		return 2
	}
	fmt.Fprintf(stdout, "LOOP-DISPATCH-OK %s\n", key)
	return 0
}

// PassTarget is the change and checkout the self-review pass is recorded
// under: change <name>, the checkout the caller's directory belongs to when it
// resolves, lp.main otherwise. check-loop-dispatch begins the pass here and
// flow-guard pipeline-fix end closes it here.
func PassTarget(env core.Env, git string, lp Loop) (change, main string) {
	change, main = strings.TrimSuffix(strings.TrimPrefix(lp.prefix, "self-review-"), "-"), lp.Main
	if m, ok := LoopMain(env, git, env.Dir); env.Dir != "" && ok {
		main = m
	}
	return change, main
}

// LoopFromBranch is the loop a branch carries: spectre/<name> an own-branch
// self-review, pipeline-fix-<name>-<k> and self-review-<name> a loop branch.
func LoopFromBranch(branch string) (Loop, bool) {
	if name, ok := strings.CutPrefix(branch, "spectre/"); ok && core.PlainChangeName(name) {
		return Loop{Change: name, Token: "self-review-" + name, prefix: "self-review-" + name + "-"}, true
	}
	if m := PipelineBranch.FindStringSubmatch(branch); m != nil && core.PlainChangeName(branch) {
		return Loop{Change: branch, Token: branch, prefix: "pipeline-fix-" + m[2] + "-"}, true
	}
	if name, ok := strings.CutPrefix(branch, "self-review-"); ok && core.PlainChangeName(name) {
		return Loop{Change: branch, Token: branch, prefix: branch + "-"}, true
	}
	return Loop{}, false
}

func OwnBranch(branch string) bool { return strings.HasPrefix(branch, "spectre/") }
