# kan-224-myflow-add-a-workspace-id-command

## Why

`skills/flow-contracts/workspace-isolation.md`'s workspace-id derivation (SHA-256 prefix, byte-wise
normalisation, segment-bounded prefix) is deterministic and recorded nowhere, by design — so
`/flow`'s finish-contract run 2 re-derives it by hand from the contract every single run, for the
one purpose of substituting it into the project's `remove` command. On KAN-39's run 2 that contract
was the single largest read of the run, for one 6-line value. The Go CLI (`stats/cmd/flow`) already
has an established pure-function-plus-command-file pattern (`tick.go`/`tasks.go`) and already owns
every other piece of pipeline state — it should own this derivation too, so run 2 stops
hand-transcribing it immediately before a destructive removal command.

## What changes

- A new `flow workspace-id <name>` CLI command prints the workspace id for a change name, computed
  in Go rather than re-derived from prose.
- `skills/flow-contracts/finish-contract-run2.md` step 5 is updated to call the new command instead
  of re-deriving the id by hand.
- `skills/flow-contracts/workspace-isolation.md` stays canonical for the derivation itself; nothing
  about the spec changes.
