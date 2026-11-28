package gitops

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// flow-guard stop-worktree-processes <worktree>... — stop every process whose working directory
// lies inside one of the worktrees, before the worktree is removed (KAN-1156: a finished change's
// backend jars and Gradle daemons kept ~6 GB resident for a day and a half). A dev stack, a Gradle
// daemon, a dev server, a test worker started from a change's worktree has that worktree as its
// cwd; nothing else names it as theirs.
//
// Each is sent SIGTERM; one still alive after swpGrace is sent SIGKILL. This guard's own process
// and every ancestor of it (the shell running the cleanup) are never touched.
//
// Prints one line per process, `STOPPED: <pid> <worktree> — <command>`, then
// `STOP-WORKTREE-PROCESSES-OK: <n> stopped`. Exit 0 always for a stoppable or absent process,
// including `lsof` missing (`STOP-WORKTREE-PROCESSES-NA: lsof is not on PATH`: nothing can be
// found, and a cleanup is never blocked by this); 2 a usage error.
//
// ponytail: a Gradle daemon is reused across projects, so one started from a finished worktree
// and now serving another build dies under it; the client of that build starts a new daemon.
func init() { core.Registry["stop-worktree-processes"] = stopWorktreeProcesses }

// swpGrace is how long a SIGTERMed process has to exit before SIGKILL.
const swpGrace = 5 * time.Second

func stopWorktreeProcesses(args []string, env core.Env, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		core.Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard stop-worktree-processes <worktree>...")
		return 2
	}
	n, na := StopProcesses(env, args, stdout)
	if na != "" {
		fmt.Fprintln(stdout, "STOP-WORKTREE-PROCESSES-NA: "+na)
		return 0
	}
	fmt.Fprintf(stdout, "STOP-WORKTREE-PROCESSES-OK: %d stopped\n", n)
	return 0
}

// StopProcesses stops the processes inside the worktrees and returns how many it stopped, or why it could not look.
func StopProcesses(env core.Env, worktrees []string, stdout io.Writer) (int, string) {
	lsof, ok := core.LookPath(env, "lsof")
	if !ok {
		return 0, "lsof is not on PATH"
	}
	var roots []string
	for _, w := range worktrees {
		a := WorkspaceAbs(env, w)
		if r, err := filepath.EvalSymlinks(a); err == nil {
			a = r
		}
		roots = append(roots, filepath.Clean(a))
	}
	out, err := exec.Command(lsof, "-nP", "-d", "cwd", "-Fpn").Output()
	if len(bytes.TrimSpace(out)) == 0 {
		if err != nil {
			return 0, "lsof failed: " + err.Error()
		}
		return 0, ""
	}
	skip := swpAncestors()
	type proc struct {
		pid  int
		root string
	}
	var found []proc
	pid := 0
	for _, l := range strings.Split(string(out), "\n") {
		switch {
		case strings.HasPrefix(l, "p"):
			pid, _ = strconv.Atoi(l[1:])
		case strings.HasPrefix(l, "n") && pid > 0 && !skip[pid]:
			for _, r := range roots {
				if p := l[1:]; p == r || strings.HasPrefix(p, r+"/") {
					found = append(found, proc{pid, r})
					skip[pid] = true
				}
			}
		}
	}
	for _, p := range found {
		cmd, _ := exec.Command("ps", "-o", "command=", "-p", strconv.Itoa(p.pid)).Output()
		fmt.Fprintf(stdout, "STOPPED: %d %s — %s\n", p.pid, p.root, cut(strings.TrimSpace(string(cmd)), 100))
		_ = syscall.Kill(p.pid, syscall.SIGTERM)
	}
	deadline := time.Now().Add(swpGrace)
	for _, p := range found {
		for syscall.Kill(p.pid, 0) == nil && time.Now().Before(deadline) {
			time.Sleep(50 * time.Millisecond)
		}
		if syscall.Kill(p.pid, 0) == nil {
			_ = syscall.Kill(p.pid, syscall.SIGKILL)
		}
	}
	return len(found), ""
}

// swpAncestors is this process and every ancestor of it.
func swpAncestors() map[int]bool {
	skip := map[int]bool{}
	for pid := os.Getpid(); pid > 1 && !skip[pid]; {
		skip[pid] = true
		out, err := exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid)).Output()
		if err != nil {
			break
		}
		pid, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}
	return skip
}

func cut(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
