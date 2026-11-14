package guard

import (
	"slices"
	"testing"
)

// TestExitSitesTaggedPlanShape is check-plan-shape's range of the reason
// pin (decision pin-per-range): its file and its five sub-guards' files.
func TestExitSitesTaggedPlanShape(t *testing.T) {
	t.Parallel()
	for _, v := range exitSiteViolations(reasonPinRange(t, func(n string) bool { return slices.Contains(reasonPinPSFiles, n) }), reasonPinExemptPS) {
		t.Error(v)
	}
}

// reasonPinPSFiles are check-plan-shape's file and its five sub-guards'.
var reasonPinPSFiles = []string{"planshape.go", "plansuitesteps.go", "planscriptpaths.go", "planfilespaths.go", "plantaskidrule.go", "planstepfiles.go"}

var reasonPinExemptPS = map[string]string{
	"psWorse":               "combines exit codes its callers recorded",
	"psPlan.afterFieldLine": "returns a line number",
}
