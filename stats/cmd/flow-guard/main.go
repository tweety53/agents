// Command flow-guard runs the flow guards. Each is called as
// `flow-guard <name> [arguments]`, through scripts/flow-guard, which builds
// this command from its own checkout and execs it. The guards live in
// internal/guard, one file each, registered by name in guard.Registry; each
// guard's Go file's leading comment is its canonical contract — arguments,
// environment overrides, verdict lines, the stdout/stderr split and exit
// codes. This command only dispatches to them, after the KAN-902 re-exec
// (guard.CwdCheckout): a self-judging guard called from another worktree of
// its checkout's repository runs that worktree's own scripts/flow-guard.
//
// Exit 2 on a missing or unregistered guard name: a usage error, never read
// as a guard's answer. Past the name, a failure of this command is the named
// guard's own cannot-answer code (guard.CannotAnswer) -- not a flat 2, which
// run-reproducer's contract reads as "refused".
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/tweety53/agents/stats/internal/guard"
)

const usage = "usage: flow-guard <name> [arguments]\n\nruns guard <name>; its contract is the leading comment of its Go file under stats/internal/guard/\n"

// registry is guard.Registry, named here so telemetry.go's dispatched and
// run read the same map.
var registry = guard.Registry

func main() {
	if exe, err := os.Executable(); err == nil {
		pruneCache(exe, time.Now())
	}
	if len(os.Args) > 1 {
		if cwd, err := os.Getwd(); err == nil {
			env := guard.Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: cwd}
			if path, ok := guard.CwdCheckout(env, os.Args[1]); ok {
				err := syscall.Exec(path, append([]string{path}, os.Args[1:]...), os.Environ())
				fmt.Fprintf(os.Stderr, "%s: cannot run %s: %v\n", os.Args[1], path, err)
				os.Exit(guard.CannotAnswer(os.Args[1]))
			}
		}
	}
	os.Exit(runAndRecord(os.Args[1:], os.Stdout, os.Stderr, os.Getenv, workingDir, http.DefaultClient))
}

// runAndRecord runs the guard args name and records the run's telemetry,
// passing recordRun the first line the guard printed on stdout.
func runAndRecord(args []string, stdout, stderr io.Writer, getenv func(string) string, workingDir func() string, client *http.Client) int {
	start := time.Now()
	out := &firstLine{w: stdout}
	reasons := &guard.Reasons{}
	code := run(args, reasons, out, stderr)
	if name, ok := dispatched(args); ok {
		if dir := workingDir(); dir != "" {
			wt := guardWorktree(args, dir)
			if _, err := projectKey(context.Background(), wt); err != nil {
				wt = dir // an argument that is no git checkout: record under the cwd
			}
			recordRun(getenv, wt, name, code, string(out.line), reasons.IDs(), reasons.Headless(), time.Since(start), client)
		}
	}
	return code
}

// run dispatches args to its guard, recording the guard's reason ids into
// reasons.
func run(args []string, reasons *guard.Reasons, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	fn, ok := registry[args[0]]
	if !ok {
		fmt.Fprintf(stderr, "flow-guard: unknown guard %s\n", args[0])
		return 2
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "flow-guard: cannot read the working directory: %v\n", err)
		guard.Reason(guard.Env{Reasons: reasons}, "cannot-read-working-dir")
		return guard.CannotAnswer(args[0])
	}
	return fn(args[1:], guard.Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: cwd, Stdin: os.Stdin, Reasons: reasons, LoadGate: true}, stdout, stderr)
}
