# kan-443-redundant-verifier-dispatches-ran-after-an — design

## Context

`RenderLedger` in `stats/internal/records/render.go` (lines 321–361 at `main` `ed5c016`) writes
one `## Dispatch <seq> — <role>` block per `records.Dispatch`, with `Task`, `Role`, `Slot`
(conditional), `Model`, `Commit`, `Diff base` (conditional), `Outcome`, `Started`, `Tokens`,
`Notes` (conditional). `records.Dispatch.Key` (`stats/internal/records/types.go` line 87) is
populated from the `dispatch_key` column by the shared `dispatchColumns` scan in
`stats/internal/store/records.go`, so the renderer already receives it and simply does not print
it.

The ledger is what `flow.self-review`'s subagent reads (`skills/flow/archive.md` step 9). Its
`myflow-cost` angle misread KAN-423's three verifier rows as redundant because nothing on the
row said `verify-<worktree>` versus `visual-verify-<worktree>` versus `verify-2`. The
evidence, and the closure of that premise, are in `proposal.md` `## Why` and are not restated
here.

No file under `skills/` or `spectre/specs/` enumerates the ledger's per-row lines
(`grep -rn "Diff base" skills/ spectre/specs/` finds nothing), so no doc or spec edit accompanies
the code change.

## Change

### 1. The renderer

In `RenderLedger`, immediately after the `Slot` conditional and before `Model`:

```go unverified:to be written on the change branch
if strings.TrimSpace(d.Key) != "" {
	fmt.Fprintf(&b, "- Key: %s\n", neutraliseMarkers(d.Key))
}
```

`neutraliseMarkers` is applied for the same reason it is on `Slot` and `Diff base`: a key is
free text supplied by the dispatcher and must not be able to forge a marker line in the rendered
document. The line is conditional, not `orElse`-defaulted, for the reason the `Diff base` comment
gives — rows recorded before keys existed would otherwise each carry a `- Key: no key` line that
says nothing.

### 2. The test

`stats/internal/records/render_test.go` gains `TestRenderLedgerNamesDispatchKey`, in the shape of
`TestRenderLedgerNamesDiffBase` (lines 419–445): a `records.Run` with two dispatches, the first
keyed `verify-gymie-worktrees`, the second with an empty `Key`. It asserts
`- Slot: <slot>\n- Key: verify-gymie-worktrees\n` renders for the first (placement after `Slot`)
and that the second dispatch's block contains no `- Key:` line. The package's test count moves
from 38 to 39.

## Decisions

### Add the dispatch key to the ledger rather than gate or merge verifier dispatches

**ID:** ledger-key-line-over-dispatch-gate
**Status:** active
**Chosen:** render `- Key: <dispatch_key>` per ledger row — the KAN-423 rows show three distinct,
non-overlapping dispatches; the misread came from the ledger hiding the one field that
distinguishes them, so exposing it fixes the cause and costs a conditional `Fprintf` plus one test.
**Considered:**
- Close KAN-443 as not-a-defect with no code change — leaves the ledger ambiguous, so the next
  self-review can produce the same finding again.
- Gate verifier dispatch on the prior review round being clean, as the issue asks — there is no
  such gate to add: every verifier already runs only after the panel closes clean, and the count
  is one per worktree per stage, not a fixed number.
- Merge `flow.verify` and `flow.visual-verify` into one dispatch per worktree — saves roughly one
  dispatch's warm-up (about 300K cache-read tokens per change) while coupling two stages with
  different block rules and report shapes; not worth the coupling.

### Close the premise in the proposal, not in Jira

**ID:** evidence-in-proposal-not-jira
**Status:** active
**Chosen:** the KAN-423 dispatch evidence lives in `proposal.md` `## Why`, archived with the
change; no Jira comment is posted on KAN-443 — no extra Jira write and no new pipeline step.
**Considered:** post the same evidence as a Jira comment via the Atlassian MCP before writing the
artifacts — an outward write the pipeline's `jira-integration.md` contract does not otherwise
make, for a record the archive already keeps.

## Open questions

None.
