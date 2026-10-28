## Context

Self-review fix change for `kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets`, planned on
`self-review-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets` at agents `3aab64d6` (the
fix-base). Five non-`big` findings, as `proposal.md` names them, each carried by one task in
`tasks.md`: F1-a (Task 1), F1-b (Task 2), F1-c (Task 3), F1-e (Task 4), A8-a (Task 5). Each was
traced before its fix was designed. Every question was auto-resolved to its recommended option
(**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`). The brainstorm
design, with the traces, is
`.superpowers/sdd/2026-10-28-self-review-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-design.md`
in this worktree.

A fix repairs what its defect already wrote: the dispatch rows of the dev `flow` database are
repaired through `flowd` subcommands and `flow record dispatch backfill` in the task that fixes their
cause. The order matters: Task 1's end repair runs first, since the starts Tasks 2 and 3 derive floor
at verifier and reviewer ends Task 1 moves. `flowd` on 4173, the `flow-postgres` container and the
`flow` database are never stopped or dropped.

## Decisions

### A restamp re-ends a closed row

**ID:** restamp-reends
**Status:** active
**Chosen:** `records.DispatchEnd` gains `ReEnd bool` tagged `json:"-"`, so no HTTP body or journal
line can set it; `Store.EndDispatch` writes `ended_at = CASE WHEN <re-end> THEN $7 ELSE
COALESCE(d.ended_at, $7) END`, and `ApplyDispatchEndAt`'s second call sets it. The test runs
`api.ApplyDispatchEndAt` against the real store (`storetest`), in `stats/internal/store/records_test.go`.
A new `Store.RestampStampedEnds(ctx, since, agentEnd)` and `flowd restamp-ends <since>` move every
non-inline agent row ended at or after `since` whose end carries sub-millisecond digits — the daemon's
own stamp; transcript ends are millisecond and CLI ends whole-second — to its agent's last transcript
write (`harvest.AgentTranscriptEnd`), only when that write is before the end and not before the
start, as `agentEndOf` judges a live end. Run once with `since` 2026-10-28T15:30:00Z.
**Considered:** a new `RecordWriter` method — every fake and the replay writer would grow one for a
flag the store alone reads; reverting `6560ecb7` — a re-point after an autosquash would again
overwrite a closed row's end (kan-1046 self-review, 1a); repairing with `retime-dispatch` one row at a
time — 66 candidate rows, each needing its transcript read by hand.

### A live inline row starts no later than its derived start

**ID:** inline-end-derives-start
**Status:** active
**Chosen:** the back-fill's inline derivation (commit time, the commit before it, `panelFixFloor`,
`inlineFloor`) is extracted into one function `inlineDerivedWindow`, its second caller
`runRecordDispatchEnd`: a `completed` end of an inline `task-<n>-implementer*` row with `-commit`
sends the derived start as a new `DispatchEnd.StartedAt` (`json:"startedAt,omitempty"`), and the store
writes `started_at = LEAST(d.started_at, $start)` before its window checks read the start. A row begun
at the work's start keeps it; one begun at the commit instant moves back to the end of the parent's
last recorded step.
**Considered:** a minimum window — an arbitrary number that a 9-second row still passes; prose
telling the parent to begin earlier — `implement.md` already says so, and rows 8449 and 8458 broke it.

### A task's completed commit is recorded once

**ID:** one-completed-row-per-task-commit
**Status:** active
**Chosen:** `flow record dispatch backfill` of a `task-<n>-implementer*` key refuses, exit 2 before
any store call, when the change holds a `completed` row of a `task-<n>-implementer*` key carrying the
same commit; `flow record dispatch begin -agent-id inline` of a `task-<n>-implementer*` key refuses
when the change holds a `completed` row of a `task-<n>-implementer*` key. A failed record read refuses
nothing. A new `Store.DropDispatch(ctx, dropID, keepID)` and `flowd drop-dispatch <drop-id>
<keep-id>` delete the drop row, first moving its `metrics.tokens` onto the keep row when the keep row
has none, refusing two rows of different changes. Repairs: drop 8410 into 8411 (task 7's tokens move
to its real row), drop 8437 into 8434.
**Considered:** extending `fold-dispatch` — its guard (`agent_id IS NULL AND diff_base IS NOT NULL`)
is the diff-base repair's own; a SQL `DELETE` by hand — the project repairs rows through its own
commands.

### Rows 8449 and 8458 start at the parent's last recorded step

**ID:** retime-late-inline-rows
**Status:** active
**Chosen:** after Task 1's repair, each row's start is retimed with `flowd retime-dispatch` to the
latest end, at or before its commit's time, of an inline row or a `visual-verify*` verifier row of the
change — `inlineFloor`'s rule; end and commit unchanged.
**Considered:** the explorer's estimates (15:57:06Z, ~16:29Z) — the same rule computes them from the
rows once Task 1 has moved the verifier ends.

### The gate matches panel-fix trailers to rows

**ID:** gate-panel-fix-trailer-row
**Status:** active
**Chosen:** `check-unfinished-work` gains a signal inside its `anchor == worktree` block: for every
range of the change's worktree set, it lists the `Flow-Key:` trailers of `<base>..<tip>` — `<tip>`
`refs/flow/pre-reshape/<name>` when that ref resolves in the repository, `HEAD` otherwise — and flags
each `panel-fix-<r>[-<n>]` trailer with no dispatch row of the same key, naming the commit and
the key. A trailer one round low is caught by the same match, since the CLI keys the row a round
higher (`panelFixKeyedARoundLow`).
**Considered:** matching by commit sha — a fix spanning two repositories records one sha on one row;
walking `HEAD` only — after the reshape, the kan-1044 commits are reachable from the pin alone.

### The missing panel-fix rows are keyed by their real rounds

**ID:** backfill-corrected-rounds
**Status:** active
**Chosen:** after Task 1's repair, back-fill `panel-fix-4` (`3b9a3985`), `panel-fix-7` (`ae761ec7`)
and `panel-fix-11` (`25394c9a`) with `flow record dispatch backfill -agent-id inline -commit <sha>`
from the frontend worktree, model `opus`, session token `mf-k1044-20261028T121921Z`. The playwright
commit `0d7052b4` is the same fix as `3b9a3985`, same instant, and gets no row of its own. The
change has landed, so its gate never runs again and the trailers stay as written.
**Considered:** the trailer keys `-3`, `-6`, `-10` — the CLI refuses each as a round low; rewriting
the trailers — the commits are history behind a landed reshape.

### Band measurement crops a scrim to its sheet

**ID:** scrim-crop
**Status:** active
**Chosen:** in the page branch of `mvpOpen` (`stats/internal/guard/measurevisualproperties.go`),
after the most common colour is found: when the pixel at the region's top-left differs from it beyond
`--noise` and each of the two colours covers at least 20% of the region, the corner colour is the
scrim; the region is cropped to the bounding box of the pixels not within `--noise` of the scrim,
and the background recounted inside it. `bands`, `seams` and their `delta` pairing all read the
cropped region. A `ponytail:` comment names the ceiling: a flat scrim reaching the corner; a gradient
scrim needs a layer detector. The stale kan-1015 note in `scripts/measure-visual-properties.sh` is
updated.
**Considered:** a `--scrim` flag — every caller would need to know which frames carry one;
segmenting by colour change — rewrites `measureBands` for every frame.

### A running-clock departure closes on repeated running-clock passes

**ID:** running-clock-repro
**Status:** active
**Chosen:** one paragraph after **A task fixing a visual departure runs its capture spec before it
closes** (`skills/flow/verify-fix-loop.md`): when the departure was observed on a running clock — a
live walk, a recorded frame sequence, any read the capture spec's pinned clock
(`skills/flow/visual-verify-verifier.md`, step 8) does not reproduce — the task first adds a capture
test that runs under the observer's conditions (clock running, same viewport and data) and fails
without the fix, and closes only once that test passes five times over in one run under
`run-capped.sh` with one worker (Playwright `--repeat-each 5 --workers 1`), beside the capture spec's
own exit 0.
**Considered:** a guard — nothing machine-readable records which clock observed a departure; one more
running-clock pass — kan-1044's Task 9 passed one and failed four of five later.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Tasks 1–3 change what every run's dispatch end, back-fill, begin and
integrate gate record or refuse, and write the live store.
**Considered:** `flow-low` — the diffs are small, but a refusal wrong by one condition stops every
run's inline record or landing.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** two groups in series: `-fix-g1` Tasks 1–3, the store, CLI, `flowd` and gate work with its
ordered store repairs (Tasks 1 and 2 share `stats/internal/records/types.go`,
`stats/internal/store/records.go`, `stats/internal/store/records_test.go` and
`stats/cmd/flowd/main.go`; Tasks 2 and 3 repair after Task 1's repair); `-fix-g2` Tasks 4–5, the
measurement guard and the loop prose.
**Considered:** one fixer for all five — one context carrying store repairs, image code and prose;
one per task — five dispatches, three sharing files and an ordered repair.

## Open questions

## Live check

Each step runs before landing, with every task committed, in the worktree. Store reads use
`docker exec flow-postgres psql -U flow -d flow -tAc "<query>"`. Refusal checks run against a copy:
`scripts/live-store-copy.sh up k1044sr` from the worktree, its printed `FLOW_ADDR`/`FLOW_RECORDS_ADDR`
exported, then `cd stats && go run ./cmd/flow …`; `scripts/live-store-copy.sh down k1044sr` at the
end.

1. **Stamped ends.** `select count(*) from dispatches where ended_at >= '2026-10-28T15:30:00Z' and
   agent_id is not null and agent_id <> 'inline' and ended_at <> date_trunc('milliseconds', ended_at)`
   — before 66 (measured at plan time); after, below 66 by the count `flowd restamp-ends` printed.
   Panel-10's three reviewer rows of kan-1044 no longer share one end. Not working looks like 66 and
   one shared panel-10 end.
2. **Inline rows.** `select id, started_at, ended_at, metrics->'tokens' from dispatches where id in
   (8410, 8411, 8434, 8437, 8449, 8458)` — before six rows, 8411 without tokens, 8449 starting
   16:09:21, 8458 starting 16:46:14; after, 8410 and 8437 gone, 8411 carrying 8410's tokens, 8449 and
   8458 starting minutes before their commits. Not working looks like any before value.
3. **Panel-fix rows.** `select dispatch_key, commit_sha from dispatches d join changes c on
   c.id = d.change_id where c.name = 'kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets' and
   dispatch_key like 'panel-fix-%'` — before `panel-fix-1`, `-2`; after also `-4`, `-7`, `-11` on the
   three commits.
4. **Gate.** `scripts/check-unfinished-work.sh` on the kan-1044 frontend worktree flags `3b9a3985`,
   `ae761ec7` and `25394c9a`, whose trailers name `panel-fix-3`, `-6` and `-10`, keys no row carries.
   Not working looks like no panel-fix line.
5. **Scrim.** `scripts/measure-visual-properties.sh --props bands` on a kan-938 scrim capture from
   `/Users/tweety53/Projects/gymie` reports more than one band. Not working looks like one 1632px band.

## Recommended approach

Fix each defect where it starts: the store re-ends a row when the restamp asks; the CLI derives an
inline row's start the way the back-fill does and refuses a duplicate task record; the gate reads the
panel-fix trailers it already lists; the measurement crops a scrim before segmenting; the loop prose
names the running-clock reproduction. Rows the defects wrote are repaired by `flowd` and the back-fill
CLI in the same task, Task 1's repair first. Rejected alternatives are under each decision:
`restamp-reends`, `inline-end-derives-start`, `one-completed-row-per-task-commit`,
`gate-panel-fix-trailer-row`, `scrim-crop`, `running-clock-repro`.

## Critical files

- `stats/internal/records/types.go` — `DispatchEnd`'s `ReEnd` and `StartedAt`.
- `stats/internal/store/records.go`, `records_test.go` — `EndDispatch`, new `RestampStampedEnds`, `DropDispatch`.
- `stats/internal/api/records.go` — `ApplyDispatchEndAt`.
- `stats/cmd/flowd/main.go` — `restamp-ends`, `drop-dispatch`.
- `stats/cmd/flow/record.go`, `record_test.go` — the back-fill derivation, `end`, `begin`.
- `stats/internal/guard/unfinishedwork.go`, `check_unfinished_work_test.go` — the panel-fix signal.
- `stats/internal/guard/measurevisualproperties.go`, `measure_visual_properties_test.go`, `scripts/measure-visual-properties.sh`.
- `skills/flow/verify-fix-loop.md` — the running-clock paragraph.

## Existing code to reuse

- `agentEndOf` and `harvest.AgentTranscriptEnd` / `harvest.DefaultTranscriptsRoot` (`stats/cmd/flowd/main.go`'s `agentEnd`).
- `RetimeDispatch` / `flowd retime-dispatch`; `DropPass` / `flowd drop-pass` as the shape for `drop-dispatch`.
- `commitTime`, `commitBefore`, `panelFixFloor`, `inlineFloor`, `callRecord` + `GetRunRecord` in `record.go`.
- `unfinishedwork.go`'s `dispatches` slice, `ranges`, `capture(envGit(env)(…))`, `specRootLeaf`; the pin name of `pinPreReshape` (`integratechange.go`).
- `mvpPNG` and `mvpRun` test helpers; `mvpOpen`'s cropped-region handling.
- `scripts/run-capped.sh`.

## Verification

Beyond each task's verify step: `scripts/run-guard-tests.sh` and `cd stats && make -s web-dist &&
go test -race -count=1 $(go list ./... | grep -v '/internal/guard$')` pass; the agents `## lint` list
passes; the **Live check** above.
