# KAN-284 — a same-state write is refused, and `mergeBase` is never validated

Design for change `kan-284-store-rejects-same-state-write`.

## The reported defect, and what is actually wrong

KAN-284 reports that `myflow state set` refused an `IN_PROGRESS` → `IN_PROGRESS` write with
`store: refused: write would move state backwards`, and concludes that equal states are treated as
backwards.

**That conclusion is wrong.** `stats/internal/store/changes.go` accepts a write at the same state
whenever `updated_at` is strictly later:

```sql
WHERE state_rank(EXCLUDED.state) > state_rank(changes.state)
   OR (state_rank(EXCLUDED.state) = state_rank(changes.state)
       AND EXCLUDED.updated_at > changes.updated_at)
```

`TestPutChangeAppliesInOrderWriteAtSameState` pins exactly that behaviour, and its own comment says
so: "this is not a blanket refusal of every same-state write, only of the backwards ones."

The refusal is real; the cause is a **precision mismatch between two writers**.

- Skills write `updatedAt` from `date -u +%Y-%m-%dT%H:%M:%SZ`, per
  `skills/myflow-contracts/state-file.md` — **truncated to the second**.
- `stats/internal/api/stages.go`'s synthetic-change bootstrap writes
  `UpdatedAt: time.Now().UTC()` — **nanosecond precision**.

A truncated instant compares as earlier than any sub-second instant inside the same second, so a
same-state write is refused for up to a full second after a bootstrap, and two same-state writes
inside one wall-clock second are always refused.

Every `/myflow-*` run now fires a stage mark before its first `state set`, so **every change is born
with a nanosecond-precision `STARTED` row** that the first real write then has to beat. Observed
live in this project's store while this design was written:

```text
agents-a740d89c | kan-284-store-rejects-same-state-write | STARTED | 2026-08-23 22:06:29.96468+00 | myflow stage begin (synthetic)
```

Two further rows are parked in that same shape: `kan-315-prod-redis-volume-and-persistence`, and
`kan-239-run-2-asserts-base-branch-and-archives-via-pr` under the stray project key
`.claude-485b1ff8`.

## The second defect: `mergeBase` is unvalidated, and not where the ticket says

The wire `repos` field is not what the store persists. `reposFromWorktrees`
(`stats/internal/api/changes.go`) derives the repository set from the **`worktrees` map's values**,
which the state file contract defines as `{"<absolute worktree path>": "<merge base or null>"}`.
Whatever a skill writes as a value lands in `change_repos.merge_base` unchecked.

Nothing validates it on write. The only thing that ever looks is `check-finish-preflight.sh`, which
refuses with "no merge base recorded — cannot tell an unmerged branch from a merged one" — at the
finish gate, long after and far from the write that recorded it. In KAN-265 the recorded value was a
worktree path, and the two defects compounded: the value was wrong, and the correction was refused.

## Where the decisions live

The chosen fix, the alternatives that were ruled out and why, the contract edits, the testing shape
and what is out of scope are recorded once, in this change's own
`openspec/changes/kan-284-store-rejects-same-state-write/design.md`. This file is the diagnosis that
preceded them and does not restate them.
