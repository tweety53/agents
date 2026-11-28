package records

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// DefaultStateRoot is the root of every project's state directory, exactly
// as docs/state-file-internals.md names it; FLOW_STATE_DIR
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
	return filepath.Join(stateRoot(getenv), "hook-runs.jsonl")
}

// GuardRunsPath is the append-only file flow-guard buffers each dispatched
// guard's run in, one GuardRun JSON line, for flowd to ingest into
// guard_runs -- under the same state root as HookRunsPath.
func GuardRunsPath(getenv func(string) string) string {
	return filepath.Join(stateRoot(getenv), "guard-runs.jsonl")
}

func stateRoot(getenv func(string) string) string {
	if root := getenv("FLOW_STATE_DIR"); root != "" {
		return root
	}
	return DefaultStateRoot
}

// AppendJSONLine appends v to path as one JSON line: one O_APPEND write, so
// concurrent writers never interleave, and no flowd call -- flowd ingests
// the file on its own schedule (cmd/flowd/hookruns.go). Every failure is
// swallowed: a measurement never changes the answer of what it measures.
func AppendJSONLine(path string, v any) {
	line, err := json.Marshal(v)
	if err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(line, '\n'))
	_ = f.Close()
}
