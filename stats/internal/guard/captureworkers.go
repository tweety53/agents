package guard

import (
	"fmt"
	"io"
	"runtime"
)

// flow-guard capture-workers
//
// Prints the Playwright worker count a capture run takes, for
// `--workers=$(flow-guard capture-workers)`: half the machine's logical
// cores, at least 1 and at most captureWorkersCap. Capture specs are
// independent, so one run spreads them across that many workers instead of
// one spec after another (skills/flow/visual-verify-verifier.md, step 8).
// The cap is the capture suite's own; an ad-hoc probe that opens several
// pages by hand keeps its project's lower limit.
//
// Exit codes:
//
//	0  the count is on stdout.
//	2  an argument was passed.
func init() { Registry["capture-workers"] = captureWorkersGuard }

// captureWorkersCap bounds the count: ten pages of a canvas app at once froze
// the operator's machine, and six leaves the stack its own cores.
const captureWorkersCap = 6

// captureWorkers is the count for a machine with ncpu logical cores.
func captureWorkers(ncpu int) int {
	return min(max(ncpu/2, 1), captureWorkersCap)
}

func captureWorkersGuard(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: flow-guard capture-workers")
		Reason(env, "usage")
		return 2
	}
	fmt.Fprintln(stdout, captureWorkers(runtime.NumCPU()))
	return 0
}
