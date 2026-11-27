package guard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// flow-guard pre-existing-failure — classify a failing test command as the
// change's own or pre-existing and unrelated (**The sweep**,
// skills/flow-contracts/known-bugs.md).
//
// Usage: flow-guard pre-existing-failure <worktree> <base-ref> '<failing command>' <path>...
//
// <path> is each file or directory, relative to the repository root, the
// failing test covers. The merge-base is `git merge-base HEAD <base-ref>`;
// the change's files are `git diff --name-only <merge-base>` (committed and
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

func preExistingFailure(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "pre-existing-failure"
	if len(args) < 4 {
		fmt.Fprintf(stderr, "%s: usage: flow-guard pre-existing-failure <worktree> <base-ref> '<failing command>' <path>...\n", prog)
		Reason(env, "usage")
		return 2
	}
	wt, base, command, paths := args[0], args[1], args[2], args[3:]
	git := envGit(env)
	if _, ok := capture(git("-C", wt, "rev-parse", "--is-inside-work-tree")); !ok {
		fmt.Fprintf(stderr, "%s: not a git worktree: %s\n", prog, wt)
		Reason(env, "not-a-worktree")
		return 2
	}
	mb, ok := capture(git("-C", wt, "merge-base", "HEAD", base))
	if !ok {
		fmt.Fprintf(stderr, "%s: %s: no merge-base of HEAD and %s\n", prog, wt, base)
		Reason(env, "no-merge-base")
		return 2
	}
	diff, ok := capture(git("-C", wt, "diff", "--no-renames", "--name-only", mb))
	if !ok {
		fmt.Fprintf(stderr, "%s: %s: git diff --name-only %s failed\n", prog, wt, mb)
		Reason(env, "diff-failed")
		return 2
	}
	changed := strings.Split(diff, "\n")
	for _, p := range paths {
		p = strings.TrimSuffix(strings.TrimPrefix(p, "./"), "/")
		for _, f := range changed {
			if f != "" && (p == "." || f == p || strings.HasPrefix(f, p+"/")) {
				fmt.Fprintf(stdout, "CAUSED: %s is in the change's diff\n", p)
				Reason(env, "caused-touched")
				return 1
			}
		}
	}
	tmp, err := os.MkdirTemp("", "pre-existing-failure-")
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", prog, err)
		Reason(env, "tempdir-failed")
		return 2
	}
	defer os.RemoveAll(tmp)
	co := tmp + "/base"
	if out, err := git("-C", wt, "worktree", "add", "-q", "--detach", co, mb).CombinedOutput(); err != nil {
		fmt.Fprintf(stderr, "%s: git worktree add %s failed: %s\n", prog, mb, out)
		Reason(env, "worktree-add-failed")
		return 2
	}
	defer func() { _ = git("-C", wt, "worktree", "remove", "--force", co).Run() }()
	run := exec.Command("sh", "-c", command)
	run.Dir = co
	run.Stdout, run.Stderr = stderr, stderr
	rc := exitCode(run.Run())
	if rc == 0 {
		fmt.Fprintf(stdout, "CAUSED: passes on the merge-base %s\n", mb)
		Reason(env, "base-green")
		return 1
	}
	fmt.Fprintf(stdout, "PRE-EXISTING: fails on the merge-base %s (exit %d) and no named path is in the change's diff\n", mb, rc)
	return 0
}
