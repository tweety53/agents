# self-review-backlog-flow-1

## Why

Five tickets in KAN's "Backlog Flow" status, filed by earlier self-review passes, are planned here
as one self-review fix change. Two are stale and carry no task (`design.md`, **Context**); three
name defects the tree still has:

- **KAN-539** — a task's guessed icon and its revert were folded into one commit whose message
  still described the icon its tree did not contain. Nothing says a commit's message describes the
  tree it lands.
- **KAN-545** — fidelity specs reached their controls by hardcoded click coordinates, and a layout
  shift broke them twice in one change. Nothing tells a plan or a verifier to reach a control by a
  structural locator.
- **KAN-549** — `design.md`'s declared frame list was never reconciled against the frames the
  verifier counted. The reconciliation is the parent's by-hand reading of the report.

## What changes

- **Revert and back-out commits** (`skills/flow-contracts/git-boundaries.md`) states that a
  commit's message describes the tree it lands, never an attempt undone before it landed; the
  implementer prompt (`skills/flow/implement.md`) cites it.
- Writing plans (`skills/flow/brainstorm-planner.md`, section D) requires every click and capture
  in an end-to-end or fidelity spec on a structural locator, never a page coordinate; the
  verifier's capture step (`skills/flow/visual-verify-verifier.md`, step 8) cites it.
- A new guard, `check-frame-list.sh` (Go, `flow-guard check-frame-list`), derives the frames
  `design.md` declares from the mockups directory and fails on any declared frame with no sidecar
  line, no report line, or a `frames:` count below the declared one; `flow.visual-verify`'s
  parent runs it whenever `mockups` is declared (`skills/flow/visual-verify.md`).
