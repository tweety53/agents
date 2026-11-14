package repotest

// TestPanelIndexWording pins KAN-1110: the skill indexes say only a micro
// change's panel comes from the settings store — every other panel is the
// recorded decision's.

import (
	"strings"
	"testing"
)

func TestPanelIndexWording(t *testing.T) {
	t.Parallel()
	for _, f := range []string{
		"AGENTS.md", "CLAUDE.md", "commands-claude/flow.md", "skills/README.md",
		"skills/flow/SKILL.md", "templates/AGENTS.md", "templates/CLAUDE.md",
	} {
		if strings.Contains(repoText(t, f), "review panel resolved from the settings store") {
			t.Errorf("%s still says the review panel is resolved from the settings store", f)
		}
	}
}
