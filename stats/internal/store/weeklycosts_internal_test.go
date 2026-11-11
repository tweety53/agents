package store

import (
	"slices"
	"testing"

	"github.com/tweety53/agents/stats/internal/stages"
)

// TestWeeklyCostStagesAreDeclared fails when a WeeklyCosts category names a
// stage key the stages vocabulary does not declare.
func TestWeeklyCostStagesAreDeclared(t *testing.T) {
	c := weeklyCostStages
	for _, k := range append([]string{c.selfReview, c.visualVerify, c.implementation, c.review}, c.planning...) {
		if _, ok := stages.Name(k); !ok {
			t.Errorf("weekly-cost stage key %q is not in the stages vocabulary", k)
		}
	}
}

// TestWeeklyCostPlanningCountsPlanReview fails when the plan review's stage
// drops out of the planning category.
func TestWeeklyCostPlanningCountsPlanReview(t *testing.T) {
	if !slices.Contains(weeklyCostStages.planning, "flow.plan-review") {
		t.Errorf("weekly-cost planning stages %v lack flow.plan-review", weeklyCostStages.planning)
	}
}
