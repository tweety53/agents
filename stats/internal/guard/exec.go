// exec.go holds the process-running helpers several guards share.

package guard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// capture is `VAR="$(cmd 2>/dev/null)"`: stdout with trailing newlines
// stripped, and whether the command exited 0.
func capture(cmd *exec.Cmd) (string, bool) {
	out, err := cmd.Output()
	return strings.TrimRight(string(out), "\n"), err == nil
}

// envGit returns a builder for `git <args>` run as the bash guards ran it:
// git resolved over env's PATH, in env.Dir (so a relative worktree argument
// resolves as it did), with extra appended to the process environment.
func envGit(env Env, extra ...string) func(args ...string) *exec.Cmd {
	git := "git"
	if p, ok := lookPath(env, "git"); ok {
		git = p
	}
	return func(args ...string) *exec.Cmd {
		cmd := exec.Command(git, args...)
		cmd.Dir = env.Dir
		if len(extra) > 0 {
			cmd.Env = append(os.Environ(), extra...)
		}
		return cmd
	}
}

// lookPath is `command -v <name>` over env's PATH, not the process's.
func lookPath(env Env, name string) (string, bool) {
	for _, dir := range filepath.SplitList(env.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		p := dir + "/" + name
		if fi, err := os.Stat(p); err == nil && fi.Mode().IsRegular() && fi.Mode()&0o111 != 0 {
			return p, true
		}
	}
	return "", false
}

// gitOut runs git in env.Dir and returns its stdout with trailing newlines
// stripped, as command substitution does.
func gitOut(env Env, args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	cmd.Dir = env.Dir
	out, ok := capture(cmd)
	if !ok {
		return "", false
	}
	return out, true
}

// panelResolveGit is panel_resolve_git: the git binary's path, found by a
// real search of env's PATH (`type -P git`), so every later call runs that
// binary by path. False, with a message on stderr, when PATH has no git.
func panelResolveGit(env Env, prog string, stderr io.Writer) (string, bool) {
	p, ok := lookPath(env, "git")
	if !ok {
		fmt.Fprintf(stderr, "%s: no git binary found on PATH\n", prog)
	}
	return p, ok
}

// exitCode is `$?` after a command: its exit status, 128+n when a signal
// killed it, and 126 — bash's "found but cannot execute" — when it never
// started.
func exitCode(err error) int {
	var ee *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &ee):
		return rrExitCode(ee.ProcessState)
	}
	return 126
}

// rrExitCode is bash's `$?` after `wait`: the exit status, or 128+signal.
func rrExitCode(ps *os.ProcessState) int {
	if ws, ok := ps.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return ps.ExitCode()
}
