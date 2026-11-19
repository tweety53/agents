package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

// flow-guard run-reproducer — run a reviewer's command against a worktree
// without letting it touch that worktree or the machine.
//
// Usage: flow-guard run-reproducer <worktree> [-t <seconds>] -- <cmd...>
//
// The command runs in a throwaway copy of <worktree> — flow-guard
// throwaway-worktree's copy: HEAD, the uncommitted diff, every untracked
// file and `.superpowers/sdd`, where a panel reproducer is written — made under the system temporary directory and removed afterwards,
// so nothing the command writes, reverts or deletes reaches the worktree
// (kan-1040: a reviewer's `git checkout -q -- skills` reverted an uncommitted
// edit). It runs under flow-guard run-capped, the project's memory cap, and is
// killed — its whole process tree, by run-capped — once it outlives the time
// bound: -t seconds, 300 when not given.
//
// Exit codes: the command's own status; 124 when the bound killed it
// (stderr carries `run-reproducer: TIMED OUT after <n>s`); run-capped's 137
// when the memory cap killed it; 2 usage, a worktree that is not a
// directory, or a copy that could not be made.
func init() { Registry["run-reproducer"] = runReproducer }

const rrUsage = "usage: flow-guard run-reproducer <worktree> [-t <seconds>] -- <cmd...>"

func runReproducer(args []string, env Env, stdout, stderr io.Writer) int {
	fail := func(id, msg string) int {
		fmt.Fprintln(stderr, "run-reproducer: "+msg)
		Reason(env, id)
		return 2
	}
	if len(args) < 3 {
		return fail("usage", rrUsage)
	}
	wt, rest := args[0], args[1:]
	bound := 300 * time.Second
	if rest[0] == "-t" && len(rest) > 1 {
		n, err := strconv.Atoi(rest[1])
		if err != nil || n <= 0 {
			return fail("usage", rrUsage)
		}
		bound, rest = time.Duration(n)*time.Second, rest[2:]
	}
	if len(rest) < 2 || rest[0] != "--" {
		return fail("usage", rrUsage)
	}
	if !filepath.IsAbs(wt) {
		wt = filepath.Join(env.Dir, wt)
	}
	if !isDir(wt) {
		return fail("not-a-directory", "not a directory: "+wt)
	}
	tmp, err := os.MkdirTemp("", "run-reproducer-")
	if err != nil {
		return fail("no-temporary-directory", err.Error())
	}
	defer os.RemoveAll(tmp)
	cp := filepath.Join(tmp, filepath.Base(wt))
	defer func() {
		if err := twRemove(env, wt, cp, false); err != nil {
			fmt.Fprintf(stderr, "run-reproducer: warning: the copy %s was not removed: %v\n", cp, err)
		}
	}()
	if err := twCreate(env, wt, cp, true); err != nil {
		return fail("copy-failed", "cannot copy "+wt+": "+err.Error())
	}

	cmd := runSibling(env, "run-capped", rest...)
	cmd.Dir = cp
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Start(); err != nil {
		return fail("exec-failed", err.Error())
	}
	var timedOut atomic.Bool
	timer := time.AfterFunc(bound, func() {
		timedOut.Store(true)
		_ = cmd.Process.Signal(syscall.SIGTERM) // run-capped kills the whole tree
	})
	err = cmd.Wait()
	timer.Stop()
	if timedOut.Load() {
		fmt.Fprintf(stderr, "run-reproducer: TIMED OUT after %ds — the command was killed\n", int(bound/time.Second))
		Reason(env, "timed-out")
		return 124
	}
	return exitCode(err)
}
