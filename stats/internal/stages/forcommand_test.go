package stages

import (
	"slices"
	"testing"
)

// TestForCommandKeepsTableOrder pins that ForCommand lists a command's
// stages in Table order and leaves out the stages it never runs.
func TestForCommandKeepsTableOrder(t *testing.T) {
	keys := ForCommand(Flow)
	if len(keys) == 0 || keys[0] != "flow.plan" {
		t.Fatalf("ForCommand(/flow) = %v, want flow.plan first", keys)
	}
	if i, j := slices.Index(keys, "flow.implement"), slices.Index(keys, "flow.review"); i < 0 || j < 0 || i > j {
		t.Errorf("ForCommand(/flow) = %v, want flow.implement before flow.review", keys)
	}
	if ForCommand("nope") == nil || len(ForCommand("nope")) != 0 {
		t.Errorf("ForCommand(unknown) = %v, want empty", ForCommand("nope"))
	}
}
