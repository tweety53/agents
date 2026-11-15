package fallback

import "testing"

// TestStateRootReadsTheGivenEnvironment pins that the state root is the
// caller's FLOW_STATE_DIR, read through the getenv it is handed (decision
// getenv-explicit-param), and DefaultStateRoot when that is empty.
func TestStateRootReadsTheGivenEnvironment(t *testing.T) {
	t.Parallel()
	if got := StateRoot(func(k string) string {
		if k == "FLOW_STATE_DIR" {
			return "/x"
		}
		return ""
	}); got != "/x" {
		t.Errorf("StateRoot with FLOW_STATE_DIR=/x = %q, want /x", got)
	}
	if got := StateRoot(func(string) string { return "" }); got != DefaultStateRoot {
		t.Errorf("StateRoot with no FLOW_STATE_DIR = %q, want %q", got, DefaultStateRoot)
	}
}
