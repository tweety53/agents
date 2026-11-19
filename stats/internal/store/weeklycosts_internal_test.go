package store

import (
	"slices"
	"testing"

	"github.com/tweety53/agents/stats/internal/stages"
)

// TestWeeklyCostStagesAreDeclared fails when a WeeklyCosts category names a
// stage key that is neither in the stages Table nor in stages.Retired.
func TestWeeklyCostStagesAreDeclared(t *testing.T) {
	t.Parallel()
	c := weeklyCostStages
	all := append([]string{c.selfReview}, c.implementation...)
	all = append(all, c.review...)
	all = append(all, c.verify...)
	all = append(all, c.planning...)
	for _, k := range all {
		_, inTable := stages.Name(k)
		_, retired := stages.Retired[k]
		if !inTable && !retired {
			t.Errorf("weekly-cost stage key %q is in neither the stages Table nor stages.Retired", k)
		}
	}
}

// TestWeeklyCostCountsRetiredKeys fails when a retired key's cost leaves the
// category its replacement belongs to.
func TestWeeklyCostCountsRetiredKeys(t *testing.T) {
	t.Parallel()
	c := weeklyCostStages
	for old, now := range stages.Retired {
		for _, cat := range [][]string{c.planning, c.implementation, c.review} {
			if slices.Contains(cat, now) && !slices.Contains(cat, old) {
				t.Errorf("category %v counts %q but not its retired key %q", cat, now, old)
			}
		}
	}
	for _, k := range []string{"flow.implement", "flow.review"} {
		if !slices.Contains(append(slices.Clone(c.implementation), c.review...), k) {
			t.Errorf("no weekly-cost category counts %q", k)
		}
	}
}
