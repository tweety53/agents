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
	start := time.Now()
	out := &firstLine{w: os.Stdout}
	code := run(os.Args[1:], out, os.Stderr)
	if name, ok := dispatched(os.Args[1:]); ok {
		if dir := workingDir(); dir != "" {
			recordRun(os.Getenv, dir, name, code, string(out.line), time.Since(start), http.DefaultClient)
		}
	}
	os.Exit(code)
}

func run(args []string, stdout, stderr io.Writer) int {
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
		return guard.CannotAnswer(args[0])
	}
	return fn(args[1:], guard.Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: cwd, Stdin: os.Stdin}, stdout, stderr)
}
