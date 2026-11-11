package guard

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// TestDesignSectionsNamedInProse pins the prose copies of the plan-mode
// design sections to psDesignSections, which check-plan-shape enforces: the
// flow-plan list and template and the spectre-planning list each name every
// heading (kan-1104 self-review, F7: a heading added to the guard reached no
// prose copy, and no check noticed).
func TestDesignSectionsNamedInProse(t *testing.T) {
	t.Parallel()
	read := func(p string) string {
		b, err := os.ReadFile("../../../" + p)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	const flowPlan, spectrePlanning = "skills/flow-plan/SKILL.md", "skills/spectre-planning/SKILL.md"
	fp, sp := read(flowPlan), read(spectrePlanning)
	fpLines := strings.Split(fp, "\n")
	for _, h := range psDesignSections {
		if !strings.Contains(fp, "`"+h+"`") {
			t.Errorf("%s: the flow-plan list lacks `%s`", flowPlan, h)
		}
		if !slices.Contains(fpLines, h) {
			t.Errorf("%s: the flow-plan template lacks a %s line", flowPlan, h)
		}
		if !strings.Contains(sp, "`"+h+"`") {
			t.Errorf("%s: the spectre-planning list lacks `%s`", spectrePlanning, h)
		}
	}
}
