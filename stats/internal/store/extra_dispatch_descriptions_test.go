package store

import (
	"os"
	"strings"
	"testing"
)

// TestExtraDispatchDescriptionsMatchReviewPanel keeps extraDispatchDescriptions
// in step with the always-on extra dispatch's table in
// skills/flow/review-panel.md, the text the map copies.
func TestExtraDispatchDescriptionsMatchReviewPanel(t *testing.T) {
	doc, err := os.ReadFile("../../../skills/flow/review-panel.md")
	if err != nil {
		t.Fatalf("read review-panel.md: %v", err)
	}
	for slot, desc := range extraDispatchDescriptions {
		name, rest, _ := strings.Cut(desc, " — ")
		row := "| `" + slot + "` | **" + name + "** — " + rest + " |"
		if !strings.Contains(string(doc), row) {
			t.Errorf("review-panel.md carries no row %q", row)
		}
	}
}
