# kan-383-flow-add-a-file-flag-to-flow-state-set

## Why

A compound heredoc-piped `flow state set` command was blocked by the harness's
permission classifier mid-run (self-review finding from KAN-380, filed as
KAN-383), requiring a workaround: write the JSON to a scratch file, then
`flow state set ... < file`. The canonical form the state contract itself
documents is already a pipe — `printf '%s' "$RECORD_JSON" | flow state set
"$NAME" -C "$DIR"` (`skills/flow-contracts/state-file.md`, **Read it, write
it**) — so the canonical write carries the same classifier friction on every
run. A `-file <path>` flag makes the canonical form a single non-compound
command a `Bash(flow state set *)` allowlist rule can cover cleanly, the same
way KAN-224 moved the workspace-id shell derivation into the CLI as `flow
workspace-id`.

## What changes

- `flow state set -file <path> <name>` reads the change's whole record as JSON
  from the named file. Without `-file`, stdin remains the source, unchanged.
- File read errors (missing, unreadable, oversize, not a JSON object) are
  local input errors: reported on stderr, exit 2, never the fallback path —
  the same family non-JSON stdin already belongs to.
- `skills/flow-contracts/state-file.md`'s canonical form becomes the
  single-command `flow state set -file <path> "$NAME" -C "$DIR"`, with the
  piped form kept as the documented stdin alternative, and names
  `Bash(flow state set *)` as the allowlist rule the file form is designed
  for. No config file is touched.
- `stats/cmd/flow/state.go`'s `stateUsage` and `main.go`'s `usage` line name
  the new flag; tests in `stats/cmd/flow/state_test.go` cover the file path
  and the set-only restriction (`state get -file` stays a usage error).
