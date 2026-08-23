# Verification — kan-212

Task 12's record: every command run, its result, and the before/after figures the proposal's claims
rest on. Measured on 2026-08-23 in the worktree `/Users/tweety53/Projects/agents-kan-212`, against the
live store on `localhost:5433`.

## Lint

```bash
cd stats && gofmt -w . && go vet ./... && gofmt -l .
```

`go vet` exit 0. `gofmt -l .` printed nothing.

## The Go suite

```bash
cd stats && go test ./... -count=1
```

All 15 packages `ok` — `stats`, `cmd/myflow`, `cmd/myflowd`, `cmd/uitest-seed`, `internal/api`,
`internal/client`, `internal/config`, `internal/fallback`, `internal/harvest`, `internal/reconcile`,
`internal/records`, `internal/stages`, `internal/store`, `internal/sweep`, `internal/web`.

## The SPA

```bash
cd stats/web && npx tsc -b
```

Exit 0. This change adds no view; the check confirms it broke none.

## The guards

Every guard `.myflow/project.md`'s `## lint` names, run from the worktree root. All exit 0:

`check-contract-budget` · `check-guard-symlinks` · `check-installed-citations` ·
`check-installed-rules` · `check-plan-provenance` · `check-references` · `check-self-review-report` ·
`check-stage-mark-calls` · `check-task-build-green` · `check-uitest-overrides` · `check-vocabulary` ·
`check-workspace-isolation`

**`check-task-commit-fields.sh` could not run at all, for every task in this change.** It exits **2**
in this worktree because `openspec/changes/` holds three unarchived changes — this one plus `kan-242`
and `kan-295`, both complete but awaiting their `/myflow-finish` run 2 — and it resolves the change by
globbing for a single `tasks.md`. Each task's commit was therefore checked **by hand** against its
declared `Files:`, `Tests:` and `Commit:` fields, per the skill's own fallback. Every one matched. The
guard being unusable in any worktree carrying more than one open change is a defect in the guard, not
in this change, and is filed as a follow-up.

## The live restart

The daemon was rebuilt from this branch and restarted with `make restart`, which stops whatever holds
the live port, rebuilds, and starts the new binary detached with output to `/tmp/myflow-live.log`.
Migration `0013_session_token_giveups.sql` applied to the live store cleanly — the table exists and
started empty, as expected on a first run.

### What the log already showed

The pre-existing log carried this failure three separate times before today — 2026-08-21T13:29,
2026-08-22T01:40, and again on the first restart at 2026-08-23T23:10 — each abandoning 11 or 15 stage
runs:

```text
level=WARN msg="harvest: session token unresolved after the bounded window, giving up"
  stage_run_ids="[...]" cycles=60
```

So this is a recurring mode, not a one-off, and until this change its only trace was a log line on a
stream that on this machine was not being kept at all.

### Restart 1 — the give-up persists, but the retry recovers nothing

`session_token_giveups` gained two rows, `mf-kan302-a3f9` and `mf-kan190-a3f7`, both
`reason = session never bound`, `retries = 0`. The durable record works.

`mf-kan302-a3f9` still bound **0 of 15** stage runs. The retry re-ran the same bounded window and gave
up again. Cause, established rather than guessed:

- `harvest_offsets` for that transcript read **3,147,539 bytes**, exactly the file's full length;
- `matchSessionTokens` scans only commands **newly read** from a transcript in that cycle;
- its own doc comment states the consequence — *"once the owning session's own mark bytes have already
  been read past (the offset only ever advances)…"*.

Recorded as **F14**, and repaired by task 6.2, which added an offset-independent scan. Task 5's unit
tests could not have caught this: their fake binder supplies matches directly, and the real constraint
is the harvest offset, which no fake models. **Only the live restart could surface it, which is what
this step exists for.**

### Restart 2 — after task 6.2

| Token | Stage runs | Bound before | Bound after |
|-------|-----------:|-------------:|------------:|
| `mf-kan302-a3f9` | 15 | 0 | **15** |
| `mf-kan190-a3f7` | 11 | 0 | **11** |

26 stage runs across two changes, recovered from a state that was previously permanent.

**Binding recovers identity, not past cost.** Every one of those 15 stage runs carries its `session_id`
and **no** `tokens` key: the usage behind the attribution offset is deliberately never re-read, which
is what makes double-counting structurally impossible. A future casualty is fully recovered — the scan
binds it while its transcript is still being written, so everything after attributes normally — while
an already-finished run regains identity alone. `design.md`, `proposal.md` and the delta spec claimed
outright recovery and were narrowed to say this; recorded as **F15**.

### The stamps, live

Both stamp paths fired against the real store:

| Change | `metrics.unattributed` | Rows |
|--------|------------------------|-----:|
| `kan-302-panel-code-review-slot-hangs-on-fork` | `{"reason": "session never bound"}` | 12 |
| `kan-212-persist-per-dispatch-cost-tokens-model-and-role` | `{"reason": "matched more than one dispatch", "candidates": 15}` | 3 |

Two things this confirms beyond the unit tests. The session-never-bound stamp carries **no**
`candidates` key, so the omit-where-not-an-ambiguity rule holds in practice. And this run's own panel
dispatches hit the original defect exactly — 15 concurrent windows with no agent id to separate them,
because `begin` is recorded before the harness reports one. Under the old code one of those 15 would
have been credited with the whole group's tokens and reported as measured; it now refuses to guess and
says so.

## The dispatch baseline

The plan's baseline was 57 dispatch rows, 20 costed. Both figures moved during the run because this
change's own 27 dispatches were recorded into the same store, so the honest comparison is per change:

| Change | Dispatches | Costed | Stamped |
|--------|-----------:|-------:|--------:|
| `kan-212-…` *(this run)* | 27 | 20 | 3 |
| `kan-242-devstop-not-running-while-stack-alive` | 14 | 7 | 0 |
| `kan-286-numeric-keyboard-for-reps-and-seconds` | 8 | 1 | 0 |
| `kan-295-cut-pipeline-load-cost-split-by-consumer` | 20 | 12 | 0 |
| `kan-302-panel-code-review-slot-hangs-on-fork` | 12 | 0 | 12 |
| `kan-315-prod-redis-volume-and-persistence` | 16 | 13 | 0 |
| `keyboard-positioning-across-the-app` | 3 | 0 | 0 |

**The historical rows did not gain cost, and were never going to.** Their usage was read past long
before this change existed, and nothing here re-reads it. What changed for them is that 12 previously
indistinguishable rows now say *why* they have no cost. The attribution fixes govern what is recorded
from here on.

## Open question, answered

`kan-302-give-up-trigger` in `design.md` asked why that session exhausted its binding window. The
restarts answer the part that matters: the mechanism is the bounded give-up, and the reason the
**retry** never helped is the harvest offset — which task 6.2 fixed. Why the *original* window was
exhausted while the run was live remains unestablished; it now leaves a durable row in
`session_token_giveups` and a kept log line, so the next occurrence is diagnosable rather than
invisible.
