package guard

import "testing"

// TestExitSitesTaggedRangeD is the reason pin over the guard files from pythonsuppressions.go up to testcost.go
// (decision pin-per-range).
func TestExitSitesTaggedRangeD(t *testing.T) {
	t.Parallel()
	keep := func(n string) bool { return n >= "pythonsuppressions.go" && n < "testcost.go" }
	for _, v := range exitSiteViolations(reasonPinRange(t, keep), reasonPinExemptD) {
		t.Error(v)
	}
}

var reasonPinExemptD = map[string]string{
	"CannotAnswer": "is the cannot-answer code table",
	"rgiRun":       "returns a child process's exit status",
	"rcInHunk":     "counts lines",
}
