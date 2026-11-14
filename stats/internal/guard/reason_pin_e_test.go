package guard

import "testing"

// TestExitSitesTaggedRangeE is the reason pin over the guard files from testcost.go to the end
// (decision pin-per-range).
func TestExitSitesTaggedRangeE(t *testing.T) {
	t.Parallel()
	keep := func(n string) bool { return n >= "testcost.go" }
	for _, v := range exitSiteViolations(reasonPinRange(t, keep), reasonPinExemptE) {
		t.Error(v)
	}
}

var reasonPinExemptE = map[string]string{
	"wsExitCode": "returns a child process's exit status",
}
