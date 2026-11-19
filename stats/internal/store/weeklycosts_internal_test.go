package store

import (
	"slices"
	"testing"

	"github.com/tweety53/agents/stats/internal/stages"
)

// TestWeeklyCostStagesAreDeclared fails when a WeeklyCosts category names a
// stage key the stages vocabulary does not declare -- save the retired
// planning keys only historical rows carry, kept so their cost stays
// planning.
func TestWeeklyCostStagesAreDeclared(t *testing.T) {
	t.Parallel()
	retired := []string{"flow.design-approval", "flow.plan-review"}
	c := weeklyCostStages
	for _, k := range append([]string{c.selfReview, c.visualVerify, c.implementation, c.review}, c.planning...) {
		if _, ok := stages.Name(k); !ok && !slices.Contains(retired, k) {
			t.Errorf("weekly-cost stage key %q is not in the stages vocabulary", k)
		}
	}
}

// TestWeeklyCostPlanningCountsPlanReview fails when the plan review's stage
// drops out of the planning category.
func TestWeeklyCostPlanningCountsPlanReview(t *testing.T) {
	t.Parallel()
	if !slices.Contains(weeklyCostStages.planning, "flow.plan-review") {
		t.Errorf("weekly-cost planning stages %v lack flow.plan-review", weeklyCostStages.planning)
	}
}
