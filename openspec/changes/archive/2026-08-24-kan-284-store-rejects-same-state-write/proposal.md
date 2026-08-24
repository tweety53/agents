## Why

A `/myflow-*` run's first stage mark bootstraps a synthetic `STARTED` change row stamped with
`time.Now()` at nanosecond precision, while every skill writes `updatedAt` truncated to the second.
The store orders same-state writes by that instant, so a truncated instant compares as earlier than
any sub-second instant in the same second and the write is refused — reported as
`store: refused: write would move state backwards`.

The pipeline **requires** same-state writes: a fix never moves the state. Under this mismatch a fix
run cannot record `prUrl`, cannot correct a bad `mergeBase`, and cannot record anything at all until
the state rank advances. KAN-284 attributes this to equal states being treated as backwards; the
store already permits a same-state write with a later instant, so the reported cause is not the real
one and a fix aimed at it would change nothing.

The same ticket reports a second, independent defect: nothing validates a recorded merge base at the
point it is written, so a wrong value survives until `check-finish-preflight.sh` refuses at the
finish gate. The two compound — the value was wrong, and the correction was refused.

## What Changes

- `myflow state set` stamps `updatedAt` itself, at full precision, overwriting whatever the body
  carried. One instant reaches the store row, the on-disk fallback file and the journal entry alike.
- The monotonic write rule is left **unchanged**. It was never wrong; it was being fed values from
  two clocks at two precisions.
- Every `worktrees` map value must be JSON `null` or a 40-character lowercase hex sha. `myflow state
  set` refuses a violation before the store is touched, naming the offending worktree path and the
  rejected value, and exits non-zero without journalling. `PutChange` refuses it as a typed error
  too, covering journal replay of a hand-edited fallback file.
- **BREAKING** for skill authors only: `updatedAt` becomes CLI-owned. Skills no longer read the
  clock for it and no longer emit it. The field remains accepted-and-ignored on the wire, so journal
  entries already on disk still replay.

## Capabilities

### New Capabilities

*(none)*

### Modified Capabilities

- `myflow-state-store`: the recorded instant a write is ordered by becomes CLI-owned rather than
  caller-supplied, and a recorded merge base gains a validated shape enforced at write time.

## Impact

- `stats/cmd/myflow/state.go` — `runStateSet` gains the stamp and the `worktrees` value check.
- `stats/internal/store/changes.go` — a new `ErrInvalidMergeBase`, checked alongside
  `ErrInvalidState`.
- `skills/myflow-contracts/state-file.md` — the `updatedAt` and `worktrees` bullets; the monotonic
  section keeps its rule and gains one sentence.
- No migration. The stray `.claude-485b1ff8` project row and the rows currently parked at synthetic
  `STARTED` are out of scope.
