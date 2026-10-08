# self-review-kan-758-offline-l5-fe-group-sessions-offline-logging

## panel-round-from-store

### Why

A KAN-758 fix run recorded `flow record pass` notes under round 4, which an earlier run of the same change had already used: the parent picked its round from session memory, and nothing in the store told it which rounds were taken. The notes had to be withdrawn and the run re-panelled as round 5, overwriting the earlier run's `slot-prompt-4` brief.

### What changes

- `flow record next-round -change <name>` prints one more than the highest round any pass, finding or mutation row of the change carries, or 0 when it has none; a failed read exits 1 and prints no number.
- `skills/flow/review-panel.md` takes a run's first round from that command, each later round of the run being one more.
