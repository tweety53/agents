package repotest

import (
	"strings"
	"testing"
)

// TestVisualVerifyShardProse pins the contract text of a sharded
// flow.visual-verify pass (KAN-1121): the parent runs the shards' closes one
// at a time (decision closes-serial-no-lock), one fix task's opener is the
// pass key without the shard segment (fix-opener-pass-key), a shard's writing
// state check runs on an account it registers fresh
// (shard-state-writes-own-account), and the merge-base capture is a section
// of its own the verifier cites.
func TestVisualVerifyShardProse(t *testing.T) {
	t.Parallel()
	for path, wants := range map[string][]string{
		"skills/flow/visual-verify.md": {
			"The parent runs each shard's `close` as its own Bash call, one after another, never two at once.",
			"## The merge-base capture In each repository",
		},
		"skills/flow/verify-fix-loop.md": {
			"its `**Opener:**` is the pass key without the `-s<i>of<n>` segment",
		},
		"skills/flow/visual-verify-verifier.md": {
			"runs signed in as an account the shard registers fresh for it",
			"(`skills/flow/visual-verify.md`) states",
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
