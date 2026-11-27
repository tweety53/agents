package records

import (
	"path/filepath"
	"time"
)

// DefaultStateRoot is the root of every project's state directory, exactly
// as skills/flow-contracts/state-file-internals.md names it; FLOW_STATE_DIR
// overrides it (fallback.StateRoot).
const DefaultStateRoot = "/Users/tweety53/Agents/flow/state"

// HookRun is one `flow-guard hook <name>` invocation's wall time, one JSON
// line of the HookRunsPath file the hook appends to and flowd ingests into
// hook_runs (0114_hook_runs.sql). At is the run's end.
type HookRun struct {
	Hook       string    `json:"hook"`
	DurationUS int64     `json:"us"`
	At         time.Time `json:"at"`
}

// HookRunsPath is the append-only file hook runs are buffered in, under
// the state root getenv names: FLOW_STATE_DIR, else DefaultStateRoot.
func HookRunsPath(getenv func(string) string) string {
	root := getenv("FLOW_STATE_DIR")
	if root == "" {
		root = DefaultStateRoot
	}
	return filepath.Join(root, "hook-runs.jsonl")
}
