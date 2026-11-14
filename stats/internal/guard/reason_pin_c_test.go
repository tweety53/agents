package guard

import (
	"slices"
	"testing"
)

// TestExitSitesTaggedRangeC is the reason pin over the guard files from ownedcorpus.go up to pythonsuppressions.go, check-plan-shape's six excluded
// (decision pin-per-range).
func TestExitSitesTaggedRangeC(t *testing.T) {
	t.Parallel()
	keep := func(n string) bool {
		return n >= "ownedcorpus.go" && n < "pythonsuppressions.go" && !slices.Contains(reasonPinPSFiles, n)
	}
	for _, v := range exitSiteViolations(reasonPinRange(t, keep), reasonPinExemptC) {
		t.Error(v)
	}
}

var reasonPinExemptC = map[string]string{
	"ownedCorpus":       "returns a cannot-answer its callers record",
	"checkTreeRestored": "returns a status post-mutation-check records at its own exit",
	"gitExec":           "returns a child process's exit status",
	"ppListMarkerEnd":   "returns a position",
	"ppCheckboxEnd":     "returns a position",
	"ppResolveExitCode": "combines codes recorded where each finding is found",
	"pwExitCode":        "returns a child process's exit status",
	"pfState":           "returns a status its callers record at their own exits",
}
