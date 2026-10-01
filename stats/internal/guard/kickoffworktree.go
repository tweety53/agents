package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// kickoffWorktree is scripts/kickoff-worktree.sh: `/flow`'s kickoff steps
// 1–5 (skills/flow/brainstorm.md **A**) in order — the location guard, the
// .worktrees ignore, the worktree add persisted through `flow state
// add-worktree`, the project's `## worktree setup` commands, the branch
// push. The shim's header is the contract: 0 all five done (`worktree:` and
// `merge-base:` on stdout), 1 a step stopped the run (its lines relayed; the
// worktree is already persisted when step 3 got that far), 2 usage or it
// cannot answer.
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
	if len(args) < 2 || len(args) > 3 || args[0] == "" || args[1] == "" {
		fmt.Fprint(stderr, "usage: kickoff-worktree.sh <project> <name> [<base>]\n")
		return 2
	}
	project, name := smcAbs(env, args[0]), args[1]
	// An optional base names the branch the change is cut from and lands
	// on; it flows into refs and a git config value, so it passes the
	// name validation resolve-base-branch applies.
	base := ""
	if len(args) == 3 {
		base = args[2]
		for i := 0; i < len(base); i++ {
			if !rbbNameByte(base[i], i > 0) {
				fmt.Fprintf(stderr, "%s: the base branch name is invalid\n", self)
				return 2
			}
		}
		if base == "" {
			fmt.Fprint(stderr, "usage: kickoff-worktree.sh <project> <name> [<base>]\n")
			return 2
		}
	}
	location, projectGet := env.Getenv("FLOW_GUARD_WORKTREE_LOCATION"), env.Getenv("FLOW_GUARD_PROJECT_GET")
	if location == "" || projectGet == "" {
		fmt.Fprintf(stderr, "%s: FLOW_GUARD_WORKTREE_LOCATION or FLOW_GUARD_PROJECT_GET is unset — run scripts/kickoff-worktree.sh, which sets them\n", self)
		return 2
	}
	git, ok := panelResolveGit(env, self, stderr)
	if !ok {
		return 2
	}
	// Before step 1: a worktree step 3 could not persist is never created.
	flow, ok := lookPath(env, "flow")
	if !ok {
		fmt.Fprintf(stderr, "%s: no flow binary on PATH — step 3 could not persist the worktree, so none is created\n", self)
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

	// 2. .worktrees ignored, through info/exclude — never a commit.
	if rc, _ := kwRun(env, project, git, "check-ignore", "-q", ".worktrees"); rc != 0 {
		rc, out := kwRun(env, project, git, "rev-parse", "--path-format=absolute", "--git-path", "info/exclude")
		if rc != 0 {
			return relay(1, out)
		}
		exclude := strings.TrimSpace(out)
		b, err := os.ReadFile(exclude)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(stderr, "%s: read %s: %v\n", self, exclude, err)
			return 1
		}
		if !strings.Contains("\n"+string(b), "\n.worktrees/\n") {
			if len(b) > 0 && !bytes.HasSuffix(b, []byte("\n")) {
				b = append(b, '\n')
			}
			if err := os.MkdirAll(filepath.Dir(exclude), 0o755); err == nil {
				err = os.WriteFile(exclude, append(b, ".worktrees/\n"...), 0o644)
			}
			if err != nil {
				fmt.Fprintf(stderr, "%s: append .worktrees/ to %s: %v\n", self, exclude, err)
				return 1
			}
		}
	}

	// 3. Fetch, add in one of two forms, persist before anything else runs.
	branch, wt := "spectre/"+name, filepath.Join(project, ".worktrees", name)
	if rc, out := kwRun(env, "", git, "-C", project, "fetch", "origin"); rc != 0 {
		return relay(1, out)
	}
	if base != "" {
		if rc, _ := kwRun(env, "", git, "-C", project, "rev-parse", "-q", "--verify", "refs/remotes/origin/"+base); rc != 0 {
			fmt.Fprintf(stderr, "%s: origin/%s does not exist — push the base branch first\n", self, base)
			return 1
		}
	}
	add := []string{"-C", project, "worktree", "add", wt, branch}
	if rc, _ := kwRun(env, "", git, "-C", project, "rev-parse", "-q", "--verify", "origin/"+branch); rc != 0 {
		start := "origin/" + base
		if base == "" {
			// The default branch by name, never HEAD: the main checkout may
			// be on any branch and is never moved.
			rc, out := kwRun(env, "", git, "-C", project, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
			if rc != 0 {
				fmt.Fprintf(stderr, "%s: cannot resolve the default branch — refs/remotes/origin/HEAD is unset (git remote set-head origin --auto)\n", self)
				return 2
			}
			start = strings.TrimSpace(out)
		}
		add = []string{"-C", project, "worktree", "add", wt, "-b", branch, start}
	}
	if rc, out := kwRun(env, "", git, add...); rc != 0 {
		return relay(1, out)
	}
	rc, out := kwRun(env, "", git, "-C", wt, "rev-parse", "HEAD")
	if rc != 0 {
		return relay(1, out)
	}
	mergeBase := strings.TrimSpace(out)
	// Recorded on the branch, so resolve-base-branch and every guard it
	// feeds land the change back on <base> rather than on origin/HEAD.
	if base != "" {
		if rc, out := kwRun(env, "", git, "-C", wt, "config", "branch."+branch+".flowBase", base); rc != 0 {
			return relay(1, out)
		}
	}
	if rc, out := kwRun(env, "", flow, "state", "add-worktree", "-C", project, name, wt, mergeBase); rc != 0 {
		return relay(1, out)
	}

	// 4. The project's worktree setup, fenced lines only, in order.
	// stdout alone is the section body; a warning project-get prints on
	// stderr is relayed, never parsed as a command.
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
				return relay(1, cout)
			}
		}
	case 1:
		fmt.Fprintf(stderr, "%s: no ## worktree setup declared — skipped\n", self)
	default:
		return relay(2, getErr.String())
	}

	// 5. The branch exists on the remote from its first minute.
	if rc, out := kwRun(env, "", git, "-C", wt, "push", "-u", "origin", branch); rc != 0 {
		return relay(1, out)
	}
	fmt.Fprintf(stdout, "worktree: %s\nmerge-base: %s\n", wt, mergeBase)
	return 0
}
