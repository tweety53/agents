# self-review-backlog-flow-9 — session narrative

## 2026-10-19 — planning

- Planned as item 1 of `/flow-self-review` step 3, every question auto-resolved; no flow stage marks, no decision gate.
- KAN-894's premise fell with the KNOWN-BUGS route's retirement: no skill appends to the file any more. KAN-897 was disproved: both cited names end on a whole word within 48 characters, as the rule asks. Both are cited in `design.md` for the parent to move to Done; the spares KAN-896 and KAN-898 took their places.
- KAN-886 and KAN-889 were measured by mutation: each untested refusal removed, its guard's tests stayed green at `373988df`.
- KAN-890's proposed change-wide scope was measured against case 6, which pins two runs each dispatching a round once as clean; the plan reads another token's original only for whether an original exists.
- KAN-896's "positional chunking" reproduced on kan-813's own commits and traced to `vmBlocks` standing a list item's first line alone; the guard's comparison has no positional step.
- `origin/main` checked green before planning: the guard package's tests, vet, gofmt and the prose guards.
- Every task's diff was drafted and checked in a scratch clone at `373988df`: each new test fails under its task's named mutation and passes without it, and the whole `internal/guard` package passes after Task 5.
