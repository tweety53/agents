package guard

import "testing"

// TestExitSitesTaggedRangeA is the reason pin over the guard files from asksilence.go up to integratechange.go
// (decision pin-per-range).
func TestExitSitesTaggedRangeA(t *testing.T) {
	t.Parallel()
	keep := func(n string) bool { return n >= "asksilence.go" && n < "integratechange.go" }
	for _, v := range exitSiteViolations(reasonPinRange(t, keep), reasonPinExemptA) {
		t.Error(v)
	}
}

var reasonPinExemptA = map[string]string{
	"bpRunArgv":          "returns a child process's exit status",
	"exitCode":           "returns a child process's exit status",
	"pcFlow":             "returns the flow CLI's exit status",
	"changePlanStoreDir": "returns a lookup status (0 found, 1 none, 3 ambiguous), not an exit code",
	"changePlanDir":      "returns a lookup status (0 found, 1 none, 3 ambiguous), not an exit code",
	"ccRunSurvivors":     "returns the survivors command's exit status",
}
