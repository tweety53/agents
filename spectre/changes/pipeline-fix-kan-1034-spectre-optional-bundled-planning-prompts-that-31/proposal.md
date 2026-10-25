# pipeline-fix-kan-1034-spectre-optional-bundled-planning-prompts-that-31

## Why

`check-task-records.sh` reads a cross-repo change's peer worktrees through `flow state get`, which resolves `FLOW_ADDR`. `/flow`'s verify stage runs the `## lint` list under the `KEY=value` lines `prepare-workspace.sh` exports, where `FLOW_ADDR` is the worktree's own isolated flowd and `FLOW_RECORDS_ADDR` is the dev store. The state read therefore hits a daemon holding no state record, finds no peers, and fails a cross-repo change whose task landed in a sibling worktree (`task 2 ticked but no commit with subject '…' reachable from HEAD`). Found on kan-1034: the guard exits 0 bare and 1 with `FLOW_ADDR=http://127.0.0.1:6783`.

## What changes

`trPeerHeads` (`stats/internal/guard/taskrecords.go`) runs its `flow state get` child with `FLOW_ADDR` set to `FLOW_RECORDS_ADDR` whenever the latter is set and non-empty, so the state read follows the records address as the record family already does. One test case proves it.
