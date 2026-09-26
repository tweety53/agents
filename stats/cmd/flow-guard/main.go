// Command flow-guard runs the Go ports of the flow guard scripts. Each
// ported scripts/<name>.sh keeps its header comment -- the guard's
// canonical contract -- and runs `flow-guard <name> "$@"`, built from its
// own checkout by scripts/lib/flow-guard.sh, so arguments,
// environment overrides, verdict lines, the stdout/stderr split and exit
// codes are the script's, unchanged. The guards themselves live in
// internal/guard, one file each, registered by basename in
// guard.Registry; this command only dispatches to them.
//
// Exit 2 on a missing or unregistered guard name: a usage error no shim can
// make, since each passes its own registered name, so it is never read as a
// guard's answer. Past the name, a failure of this command is the named
// guard's own cannot-answer code (cannotAnswer) -- not a flat 2, which
// run-reproducer's contract reads as "refused".
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/tweety53/agents/stats/internal/guard"
)

const usage = "usage: flow-guard <name> [arguments]\n\nruns the Go port of scripts/<name>.sh; see that script's header for its contract\n"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	fn, ok := guard.Registry[args[0]]
	if !ok {
		fmt.Fprintf(stderr, "flow-guard: unknown guard %s\n", args[0])
		return 2
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "flow-guard: cannot read the working directory: %v\n", err)
		return cannotAnswer(args[0])
	}
	return fn(args[1:], guard.Env{Getenv: os.Getenv, Dir: cwd}, stdout, stderr)
}

// cannotAnswer is guard name's "could not answer" exit code: 4 for
// run-reproducer, whose 2 means "refused"; 2 for every other guard.
func cannotAnswer(name string) int {
	if name == "run-reproducer" {
		return 4
	}
	return 2
}
