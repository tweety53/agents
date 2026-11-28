package guard

import (
	"errors"
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

// flow-guard run-detached <state-dir> -- <command> [args…] — start a command that can outrun the
// shell tool's foreground limit, detached from the caller's session: stdout and stderr to
// <state-dir>/out.log, the supervising process's pid to <state-dir>/pid, and, when the command
// ends, its exit code to <state-dir>/exit (written to exit.tmp, then renamed). Prints
// `STARTED <pid>` and exits 0 at once. A state-dir whose pid is still alive is refused, exit 1
// (`already-running`): the same command is never launched twice.
//
// flow-guard await-detached <state-dir> [-timeout <s>] — one short foreground wait on it: as soon
// as <state-dir>/exit exists prints `EXITED <code>` and the last 40 lines of out.log; at the
// timeout (default 540, max 570) prints `RUNNING <elapsed>s` and the last 10 lines, exit 0; a dead
// pid with no exit file prints `LOST`, exit 1 (`lost`). It starts no process.
//
// skills/flow/verify-and-handoff.md's **Inline verify** is the contract.
//
// Go, in flow-guard (stats/internal/guard/detached.go).
func init() {
	Registry["run-detached"] = runDetached
	Registry["await-detached"] = awaitDetached
}

// detachedSupervisor runs "$@" and records its exit; $0 is the state-dir. An inline sh -c, not a
// script file: it is the one process that outlives flow-guard to write the exit code.
const detachedSupervisor = `"$@" >"$0/out.log" 2>&1; echo $? >"$0/exit.tmp"; mv "$0/exit.tmp" "$0/exit"`

func runDetached(args []string, env core.Env, stdout, stderr io.Writer) int {
	if len(args) < 3 || args[1] != "--" {
		core.Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard run-detached <state-dir> -- <command> [args…]")
		return 2
	}
	dir := args[0]
	if pid, ok := detachedPid(dir); ok && pidAlive(pid) {
		core.Reason(env, "already-running")
		fmt.Fprintf(stderr, "run-detached: %s is still running (pid %d); wait on it with await-detached\n", dir, pid)
		return 1
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		core.Reason(env, "mkdir-failed")
		fmt.Fprintln(stderr, "run-detached:", err)
		return 1
	}
	for _, f := range []string{"exit", "exit.tmp", "out.log", "pid"} {
		_ = os.Remove(filepath.Join(dir, f))
	}
	abs, _ := filepath.Abs(dir)
	cmd := exec.Command("/bin/sh", append([]string{"-c", detachedSupervisor, abs}, args[2:]...)...)
	cmd.Dir = env.Dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		core.Reason(env, "start-failed")
		fmt.Fprintln(stderr, "run-detached:", err)
		return 1
	}
	pid := cmd.Process.Pid
	_ = os.WriteFile(filepath.Join(dir, "pid"), []byte(strconv.Itoa(pid)+"\n"), 0o644)
	go func() { _ = cmd.Wait() }() // reap it while this process lives; init reaps it after
	fmt.Fprintf(stdout, "STARTED %d\n", pid)
	return 0
}

func awaitDetached(args []string, env core.Env, stdout, stderr io.Writer) int {
	timeout, rest, ok := awaitTimeout(args)
	if !ok || len(rest) != 1 {
		core.Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard await-detached <state-dir> [-timeout <seconds, max 570>]")
		return 2
	}
	dir := rest[0]
	log := filepath.Join(dir, "out.log")
	deadline := time.Now().Add(timeout)
	for {
		if b, err := os.ReadFile(filepath.Join(dir, "exit")); err == nil {
			fmt.Fprintf(stdout, "EXITED %s\n", strings.TrimSpace(string(b)))
			io.WriteString(stdout, tailLines(log, 40))
			return 0
		}
		if pid, ok := detachedPid(dir); !ok || !pidAlive(pid) {
			if _, err := os.Stat(filepath.Join(dir, "exit")); err == nil {
				continue // it ended between the two reads
			}
			core.Reason(env, "lost")
			fmt.Fprintln(stdout, "LOST")
			return 1
		}
		if !time.Now().Before(deadline) {
			elapsed := 0
			if fi, err := os.Stat(filepath.Join(dir, "pid")); err == nil {
				elapsed = int(time.Since(fi.ModTime()).Seconds())
			}
			fmt.Fprintf(stdout, "RUNNING %ds\n", elapsed)
			io.WriteString(stdout, tailLines(log, 10))
			return 0
		}
		pause := min(time.Second, time.Until(deadline))
		time.Sleep(pause)
		core.RecordWait(env, pause) // the polling sleep is the run's wait, not its work
	}
}

// awaitTimeout pulls `-timeout <s>` (default 540, max 570) out of args.
func awaitTimeout(args []string) (time.Duration, []string, bool) {
	secs, rest := 540, []string{}
	for i := 0; i < len(args); i++ {
		if args[i] != "-timeout" {
			rest = append(rest, args[i])
			continue
		}
		if i+1 == len(args) {
			return 0, nil, false
		}
		n, err := strconv.Atoi(args[i+1])
		if err != nil || n < 1 || n > 570 {
			return 0, nil, false
		}
		secs = n
		i++
	}
	return time.Duration(secs) * time.Second, rest, true
}

func detachedPid(dir string) (int, bool) {
	b, err := os.ReadFile(filepath.Join(dir, "pid"))
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	return pid, err == nil && pid > 0
}

func pidAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// tailLines is the last n lines of path, each newline-terminated; "" when unreadable.
func tailLines(path string, n int) string {
	b, err := os.ReadFile(path)
	if err != nil || len(b) == 0 {
		return ""
	}
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n") + "\n"
}
