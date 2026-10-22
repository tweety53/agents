package guard

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

// kickoffWorktree is scripts/kickoff-worktree.sh: `/flow`'s kickoff steps
// 1–4 (skills/flow/brainstorm.md **A**) in order — the location guard, the
// worktree add at the sibling <project>-worktrees/ root (design.md:
// sibling-worktrees-layout) persisted through `flow state add-worktree`, the
// project's `## worktree setup` commands, the branch push. The shim's header
// is the contract: 0 all four done (`worktree:` and `merge-base:` on
// stdout), 1 a step stopped the run (its lines relayed; the worktree is
// already persisted when step 2 got that far), 2 usage or it cannot answer.
func init() {
	Registry["kickoff-worktree"] = kickoffWorktree
}

// kwRun runs name args from dir (or env's directory), returning its exit
// code and combined output.
func kwRun(env Env, dir, name string, args ...string) (int, string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = env.Dir
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	return exitCode(cmd.Run()), out.String()
}

// kwFenced is every line inside a ``` fence of body, in order: the commands
// of a `## worktree setup` section, never its fence markers or prose.
func kwFenced(body string) []string {
	var cmds []string
	in := false
	for _, l := range strings.Split(body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			in = !in
			continue
		}
		if in && strings.TrimSpace(l) != "" {
			cmds = append(cmds, l)
		}
	}
	return cmds
}

func kickoffWorktree(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "kickoff-worktree"
	const usage = "usage: kickoff-worktree.sh [--from <worktree>] <project> <name> [<base>]\n"
	// --from names this change's existing worktree: the base defaults to the
	// flowBase it records, and the state record is found through it.
	from := ""
	if len(args) > 0 && args[0] == "--from" {
		if len(args) < 2 || args[1] == "" {
			fmt.Fprint(stderr, usage)
			return 2
		}
		from, args = smcAbs(env, args[1]), args[2:]
	}
	if len(args) < 2 || len(args) > 3 || args[0] == "" || args[1] == "" || (len(args) == 3 && args[2] == "") {
		fmt.Fprint(stderr, usage)
		return 2
	}
	project, name := smcAbs(env, args[0]), args[1]
	// An optional base names the branch the change is cut from and lands
	// on; it flows into refs and a git config value, so it passes the
	// name validation resolve-base-branch applies.
	base := ""
	if len(args) == 3 {
		base = args[2]
	}
	git, ok := panelResolveGit(env, self, stderr)
	if !ok {
		return 2
	}
	if base == "" && from != "" {
		// Outside a repository `config --get` still reads the global config
		// and exits 1, so the worktree is checked first.
		if rc, _ := kwRun(env, "", git, "-C", from, "rev-parse", "--git-dir"); rc != 0 {
			fmt.Fprintf(stderr, "%s: --from %s is not a git worktree — cannot read its base, so none is added\n", self, from)
			return 2
		}
		// Exit 1 is no flowBase: the default branch, as without --from; any
		// other failure cannot answer the base.
		switch rc, out := kwRun(env, "", git, "-C", from, "config", "--get", "branch.spectre/"+name+".flowBase"); rc {
		case 0:
			base = strings.TrimSpace(out)
		case 1:
		default:
			fmt.Fprint(stderr, out)
			return 2
		}
	}
	if base != "" {
		for i := 0; i < len(base); i++ {
			if !rbbNameByte(base[i], i > 0) {
				fmt.Fprintf(stderr, "%s: the base branch name is invalid\n", self)
				return 2
			}
		}
	}
	location, projectGet := env.Getenv("FLOW_GUARD_WORKTREE_LOCATION"), env.Getenv("FLOW_GUARD_PROJECT_GET")
	if location == "" || projectGet == "" {
		fmt.Fprintf(stderr, "%s: FLOW_GUARD_WORKTREE_LOCATION or FLOW_GUARD_PROJECT_GET is unset — run scripts/kickoff-worktree.sh, which sets them\n", self)
		return 2
	}
	// Before step 1: a worktree step 2 could not persist is never created.
	flow, ok := lookPath(env, "flow")
	if !ok {
		fmt.Fprintf(stderr, "%s: no flow binary on PATH — step 2 could not persist the worktree, so none is created\n", self)
		return 2
	}
	relay := func(rc int, out string) int {
		fmt.Fprint(stderr, out)
		return rc
	}

	// 1. The location guard: its exit 1 or 2 stops the run with its lines.
	if rc, out := kwRun(env, "", location, project); rc != 0 {
		if rc != 1 {
			rc = 2
		}
		return relay(rc, out)
	}

	// 2. Fetch, add in one of two forms, persist before anything else runs.
	// The worktree lives beside the main checkout's physical path — the
	// root the location guard resolves — outside the repository, so nothing
	// needs ignoring; `git worktree add` creates the root on first use.
	physical, err := filepath.EvalSymlinks(project)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot resolve %s: %v\n", self, project, err)
		return 2
	}
	branch := "spectre/" + name
	wt := filepath.Join(siblingRoot(physical), name)
	if rc, out := kwRun(env, "", git, "-C", project, "fetch", "origin"); rc != 0 {
		return relay(1, out)
	}
	// The default branch by name, never HEAD: the main checkout may be on
	// any branch and is never moved.
	defaultBranch := func() (string, bool) {
		rc, out := kwRun(env, "", git, "-C", project, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
		if rc != 0 {
			fmt.Fprintf(stderr, "%s: cannot resolve the default branch — refs/remotes/origin/HEAD is unset (git remote set-head origin --auto)\n", self)
			return "", false
		}
		return strings.TrimSpace(out), true
	}
	// A base this origin lacks is created from its default branch, so every
	// repository of the change records the same base.
	if base != "" {
		if rc, _ := kwRun(env, "", git, "-C", project, "rev-parse", "-q", "--verify", "refs/remotes/origin/"+base); rc != 0 {
			def, ok := defaultBranch()
			if !ok {
				return 2
			}
			if rc, out := kwRun(env, "", git, "-C", project, "push", "origin", "refs/remotes/"+def+":refs/heads/"+base); rc != 0 {
				return relay(1, out)
			}
			if rc, out := kwRun(env, "", git, "-C", project, "fetch", "origin", base); rc != 0 {
				return relay(1, out)
			}
			fmt.Fprintf(stderr, "%s: origin/%s did not exist in %s — created it from %s\n", self, base, project, def)
		}
	}
	add := []string{"-C", project, "worktree", "add", wt, branch}
	if rc, _ := kwRun(env, "", git, "-C", project, "rev-parse", "-q", "--verify", "origin/"+branch); rc != 0 {
		start := "origin/" + base
		if base == "" {
			def, ok := defaultBranch()
			if !ok {
				return 2
			}
			start = def
		}
		add = []string{"-C", project, "worktree", "add", wt, "-b", branch, start}
	} else if base != "" {
		// Resuming a pushed branch: it must already sit on <base>, or
		// flowBase would record a base the branch was never cut from.
		if rc, _ := kwRun(env, "", git, "-C", project, "merge-base", "--is-ancestor", "origin/"+base, "origin/"+branch); rc != 0 {
			fmt.Fprintf(stderr, "%s: origin/%s does not descend from origin/%s in %s — stop and ask the operator: rebase it with git rebase --onto origin/%s <its merge base> %s and push, then re-run\n", self, branch, base, project, base, branch)
			return 1
		}
	}
	if rc, out := kwRun(env, "", git, add...); rc != 0 {
		return relay(1, out)
	}
	rc, out := kwRun(env, "", git, "-C", wt, "rev-parse", "HEAD")
	if rc != 0 {
		return relay(1, out)
	}
	mergeBase := strings.TrimSpace(out)
	stateDir := project
	if from != "" {
		stateDir = from
	}
	if rc, out := kwRun(env, "", flow, "state", "add-worktree", "-C", stateDir, name, wt, mergeBase); rc != 0 {
		return relay(1, out)
	}
	// Recorded on the branch, so resolve-base-branch and every guard it
	// feeds land the change back on <base> rather than on origin/HEAD.
	if base != "" {
		if rc, out := kwRun(env, "", git, "-C", wt, "config", "branch."+branch+".flowBase", base); rc != 0 {
			return relay(1, out)
		}
	}

	// 3. The project's worktree setup, fenced lines only, in order.
	if rc := kwSetup(env, projectGet, wt, self, stderr); rc != 0 {
		return rc
	}

	// 4. The branch exists on the remote from its first minute.
	if rc, out := kwRun(env, "", git, "-C", wt, "push", "-u", "origin", branch); rc != 0 {
		return relay(1, out)
	}
	fmt.Fprintf(stdout, "worktree: %s\nmerge-base: %s\n", wt, mergeBase)
	return 0
}

// kwSetup runs wt's `## worktree setup` fenced lines, in order, from wt:
// 0 done or none declared, 1 a command failed (named, its output relayed),
// 2 project-get cannot answer. stdout alone is the section body; a warning
// project-get prints on stderr is relayed, never parsed as a command.
func kwSetup(env Env, projectGet, wt, self string, stderr io.Writer) int {
	get := exec.Command(projectGet, wt, "worktree setup")
	get.Dir = env.Dir
	var body, getErr bytes.Buffer
	get.Stdout, get.Stderr = &body, &getErr
	switch rc := exitCode(get.Run()); rc {
	case 0:
		fmt.Fprint(stderr, getErr.String())
		for _, c := range kwFenced(body.String()) {
			if rc, cout := kwRun(env, wt, "bash", "-c", c); rc != 0 {
				fmt.Fprintf(stderr, "%s: worktree setup command failed (exit %d): %s\n", self, rc, c)
				fmt.Fprint(stderr, cout)
				return 1
			}
		}
	case 1:
		fmt.Fprintf(stderr, "%s: no ## worktree setup declared — skipped\n", self)
	default:
		fmt.Fprint(stderr, getErr.String())
		return 2
	}
	return 0
}
