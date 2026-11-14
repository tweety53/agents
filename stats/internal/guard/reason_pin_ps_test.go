package guard

import (
	"slices"
	"testing"
)

// TestExitSitesTaggedPlanShape is check-plan-shape's range of the reason
// pin (decision pin-per-range): its file and its five sub-guards' files.
func TestExitSitesTaggedPlanShape(t *testing.T) {
	t.Parallel()
	files := []string{"planshape.go", "plansuitesteps.go", "planscriptpaths.go", "planfilespaths.go", "plantaskidrule.go", "planstepfiles.go"}
	for _, v := range exitSiteViolations(reasonPinRange(t, func(n string) bool { return slices.Contains(files, n) }), reasonPinExemptPS) {
		t.Error(v)
	}
}

var reasonPinExemptPS = map[string]string{
	"psWorse":               "combines exit codes its callers recorded",
	"psPlan.afterFieldLine": "returns a line number",
}
