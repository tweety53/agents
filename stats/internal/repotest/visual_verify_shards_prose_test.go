package repotest

import (
	"strings"
	"testing"
)

// TestVisualVerifyShardProse pins the contract text of a sharded
// flow.visual-verify pass (KAN-1121): the parent runs the shards' closes one
// at a time (decision closes-serial-no-lock), one fix task's opener is the
// pass key without the shard segment (fix-opener-pass-key).
func TestVisualVerifyShardProse(t *testing.T) {
	t.Parallel()
	for path, wants := range map[string][]string{
		"skills/flow/visual-verify.md": {
			"The parent runs each shard's `close` as its own Bash call, one after another, never two at once.",
		},
	} {
		text := repoText(t, path)
		for _, want := range wants {
			if !strings.Contains(text, want) {
				t.Errorf("%s lacks %q", path, want)
			}
		}
	}
}
