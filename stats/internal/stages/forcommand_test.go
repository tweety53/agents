package stages

import (
	"slices"
	"testing"
)

// TestForCommandKeepsTableOrder pins that ForCommand lists a command's
// stages in Table order and leaves out the stages it never runs.
func TestForCommandKeepsTableOrder(t *testing.T) {
	keys := ForCommand(Flow)
	if len(keys) == 0 || keys[0] != "flow.kickoff" {
		t.Fatalf("ForCommand(/flow) = %v, want flow.kickoff first", keys)
	}
	if i, j := slices.Index(keys, "flow.sdd-tdd"), slices.Index(keys, "flow.review-panel"); i < 0 || j < 0 || i > j {
		t.Errorf("ForCommand(/flow) = %v, want flow.sdd-tdd before flow.review-panel", keys)
	}
	if slices.Contains(ForCommand(FlowFast), "flow.visual-verify") {
		t.Errorf("ForCommand(/flow-fast) carries flow.visual-verify, which /flow-fast never runs")
	}
	if ForCommand("nope") == nil || len(ForCommand("nope")) != 0 {
		t.Errorf("ForCommand(unknown) = %v, want empty", ForCommand("nope"))
	}
}
