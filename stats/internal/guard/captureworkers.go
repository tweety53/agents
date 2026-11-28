package guard

// captureWorkers is the Playwright worker count a capture run takes, printed on the visual-verify
// bundle's `workers:` line for `--workers=<n>`: half the machine's logical cores, at least 1 and at
// most captureWorkersCap. Capture specs are independent, so one run spreads them across that many
// workers instead of one spec after another (skills/flow/visual-verify-verifier.md, step 8). The
// cap is the capture suite's own; an ad-hoc probe that opens several pages by hand keeps its
// project's lower limit.
func captureWorkers(ncpu int) int {
	return min(max(ncpu/2, 1), captureWorkersCap)
}

// captureWorkersCap bounds the count: ten pages of a canvas app at once froze
// the operator's machine, and six leaves the stack its own cores.
const captureWorkersCap = 6
