package guard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// flow-guard pre-existing-failure — classify a failing test command as the
// change's own or pre-existing and unrelated (**The sweep**,
// skills/flow-contracts/known-bugs.md).
//
// Usage: flow-guard pre-existing-failure <worktree> <base-ref> '<failing command>' <path>...
//
// <path> is each file or directory, relative to the repository root, the
// failing test covers. The merge-base is `git merge-base HEAD <base-ref>`;
// the change's files are `git diff --no-renames --name-only <merge-base>` (committed and
// uncommitted).
//
//	CAUSED: <path> is in the change's diff               exit 1
//	CAUSED: passes on the merge-base <sha>               exit 1
//	PRE-EXISTING: fails on the merge-base <sha> (exit <n>) and no named path is in the change's diff   exit 0
//
// The merge-base run is `sh -c '<failing command>'` in a detached worktree of
// the merge-base, removed afterwards; its output goes to stderr. Exit 2
// cannot answer (usage, not a worktree, an unresolvable base, a failing
// diff or worktree add) — stdout empty, the cause on stderr.
//
// ponytail: the merge-base checkout gets no project setup (installed
// dependencies, workspace database); a command that needs one fails there
// for that reason and reads as PRE-EXISTING — the caller reads the stderr
// output before filing.
func init() { Registry["pre-existing-failure"] = preExistingFailure }

func preExistingFailure(args []string, env core.Env, stdout, stderr io.Writer) int {
	const prog = "pre-existing-failure"
	if len(args) < 4 {
		fmt.Fprintf(stderr, "%s: usage: flow-guard pre-existing-failure <worktree> <base-ref> '<failing command>' <path>...\n", prog)
		core.Reason(env, "usage")
		return 2
	}
	wt, base, command, paths := args[0], args[1], args[2], args[3:]
	git := core.EnvGit(env)
	if _, ok := core.Capture(git("-C", wt, "rev-parse", "--is-inside-work-tree")); !ok {
		fmt.Fprintf(stderr, "%s: not a git worktree: %s\n", prog, wt)
		core.Reason(env, "not-a-worktree")
		return 2
	}
	mb, ok := core.Capture(git("-C", wt, "merge-base", "HEAD", base))
	if !ok {
		fmt.Fprintf(stderr, "%s: %s: no merge-base of HEAD and %s\n", prog, wt, base)
		core.Reason(env, "no-merge-base")
		return 2
	}
	diff, ok := core.Capture(git("-C", wt, "diff", "--no-renames", "--name-only", mb))
	if !ok {
		fmt.Fprintf(stderr, "%s: %s: git diff --no-renames --name-only %s failed\n", prog, wt, mb)
		core.Reason(env, "diff-failed")
		return 2
	}
	changed := strings.Split(diff, "\n")
	for _, p := range paths {
		p = strings.TrimSuffix(strings.TrimPrefix(p, "./"), "/")
		for _, f := range changed {
			if f != "" && (p == "." || f == p || strings.HasPrefix(f, p+"/")) {
				fmt.Fprintf(stdout, "CAUSED: %s is in the change's diff\n", p)
				core.Reason(env, "caused-touched")
				return 1
			}
		}
	}
	tmp, err := os.MkdirTemp("", "pre-existing-failure-")
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", prog, err)
		core.Reason(env, "tempdir-failed")
		return 2
	}
	defer os.RemoveAll(tmp)
	co := tmp + "/base"
	if out, err := git("-C", wt, "worktree", "add", "-q", "--detach", co, mb).CombinedOutput(); err != nil {
		fmt.Fprintf(stderr, "%s: git worktree add %s failed: %s\n", prog, mb, out)
		core.Reason(env, "worktree-add-failed")
		return 2
	}
	defer func() { _ = git("-C", wt, "worktree", "remove", "--force", co).Run() }()
	run := exec.Command("sh", "-c", command)
	run.Dir = co
	run.Stdout, run.Stderr = stderr, stderr
	rc := core.ExitCode(run.Run())
	if rc == 0 {
		fmt.Fprintf(stdout, "CAUSED: passes on the merge-base %s\n", mb)
		core.Reason(env, "base-green")
		return 1
	}
	fmt.Fprintf(stdout, "PRE-EXISTING: fails on the merge-base %s (exit %d) and no named path is in the change's diff\n", mb, rc)
	return 0
}
