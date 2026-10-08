# self-review-kan-758-offline-l5-fe-group-sessions-offline-logging

## panel-round-from-store

### Why

A KAN-758 fix run recorded `flow record pass` notes under round 4, which an earlier run of the same change had already used: the parent picked its round from session memory, and nothing in the store told it which rounds were taken. The notes had to be withdrawn and the run re-panelled as round 5, overwriting the earlier run's `slot-prompt-4` brief.

### What changes

- `flow record next-round -change <name>` prints one more than the highest round any pass, finding or mutation row of the change carries, or 0 when it has none; a failed read exits 1 and prints no number.
- `skills/flow/review-panel.md` takes a run's first round from that command, each later round of the run being one more.

## verify-green-before-visual

### Why

In a KAN-758 fix run the first visual-verify dispatch began before the inline verify was green and had to be aborted — a wasted verifier dispatch. The stage order verify → visual-verify was stated in `skills/flow/verify-and-handoff.md`, but nothing checked it before a verifier was dispatched.

### What changes

- `check-verify-green.sh <worktree> <change> <session-token>` reads the change's dispatch rows and exits 0 only when this run's inline verify rows (role `verifier`, key `verify` or `verify-<basename>`) exist and the latest row per worktree — a fix round's `verify-fix-<k>` superseding the row it re-ran — carries outcome `completed`.
- `skills/flow/visual-verify.md` step 3 runs it before the preflight; exit 1 dispatches no verifier and returns the run to **Verify**, and exit 2 is a failing check that stops the stage.
