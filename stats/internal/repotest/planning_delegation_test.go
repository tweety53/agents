package repotest

// TestPlanningDelegationCitationHasParagraph pins that a file citing the PLANNING
// DELEGATION paragraph by name finds it: brainstorm.md carries it once as a
// `> **PLANNING DELEGATION:**` blockquote, the text every planner dispatch carries.

import (
	"strings"
	"testing"
)

func TestPlanningDelegationCitationHasParagraph(t *testing.T) {
	t.Parallel()
	text := repoText(t, "skills/flow/brainstorm.md")
	if !strings.Contains(text, "**PLANNING DELEGATION** paragraph") {
		t.Fatal("skills/flow/brainstorm.md no longer cites the PLANNING DELEGATION paragraph; update this test with the citation")
	}
	if !strings.Contains(text, "> **PLANNING DELEGATION:** ") {
		t.Error("skills/flow/brainstorm.md cites the PLANNING DELEGATION paragraph but has no `> **PLANNING DELEGATION:**` blockquote")
	}
}
