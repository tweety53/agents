package guard

import "testing"

// TestExitSitesTaggedRangeB is the reason pin over the guard files from integratechange.go up to ownedcorpus.go
// (decision pin-per-range).
func TestExitSitesTaggedRangeB(t *testing.T) {
	t.Parallel()
	keep := func(n string) bool { return n >= "integratechange.go" && n < "ownedcorpus.go" }
	for _, v := range exitSiteViolations(reasonPinRange(t, keep), reasonPinExemptB) {
		t.Error(v)
	}
}

var reasonPinExemptB = map[string]string{
	"rcOf":         "returns a child process's exit status",
	"mvRunCommand": "returns a child process's exit status",
	"mvpStep":      "returns a step, not an exit code",
}
