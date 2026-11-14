package repotest

import (
	"strings"
	"testing"
)

// TestCiteSendBackCarriesTheAgentID pins that the -cite send-back resumes the
// failed re-run's agent and its begin passes that agent's id, so its spend is
// the reviewer's, not the parent's (self-review-kan-1110 F17).
func TestCiteSendBackCarriesTheAgentID(t *testing.T) {
	t.Parallel()
	fix := repoText(t, "skills/flow/review-panel-fix-round.md")
	for _, want := range []string{
		"resumes the failed re-run's agent with `SendMessage`",
		"-agent-id <that re-run row's agentId>",
	} {
		if !strings.Contains(fix, want) {
			t.Errorf("review-panel-fix-round.md lacks %q", want)
		}
	}
	if !strings.Contains(repoText(t, "skills/flow/review-panel.md"), "save the `-cite` send-back's") {
		t.Error("review-panel.md's -agent-id rule does not except the -cite send-back")
	}
}
