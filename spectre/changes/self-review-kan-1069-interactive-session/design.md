## Context

Found by the self-review pass of `/flow-interactive` change `kan-1069-interactive-session`; the
fixes ride that change's own branch `kan-1069-interactive-session`, based at `origin/main`
(`3deff9fa`). Every finding was checked against history: none is fixed on `origin/main`, and no
archived change rejected a remedy proposed here.

**Finding 1 — the request-end guard checks only the UI check (flow-fix).**
`gateLandedRequest` (`stats/cmd/flow/request.go`) reads only `flow.visual-verify` stage runs, and
returns 0 before any store read when the range touches no `ui paths`. In the dev store only request
1 of `kan-1069-interactive-session` carries a `flow.verify` run; requests 2–12 carry none and all
landed.
<!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "select … from stage_runs … where c.name='kan-1069-interactive-session' and s.stage in ('flow.visual-verify','flow.verify')" @ dev store 2026-10-30 (a live read, not re-runnable at a ref) -->

**Finding 2 — a chained UI check passes the guard (flow-fix).** The store keeps microseconds:
the chained `flow.visual-verify` runs of requests 3 and 9 ended 31 ms after they began, while the
shortest honest UI check of the change took 11 s.
<!-- measured: the same psql read, ended_at - started_at = 00:00:00.03101 and 00:00:00.031117; shortest other 00:00:11.065958 @ dev store 2026-10-30 (a live read) -->
`client.StageRunRow` carries no `endedAt` today, though `GET /api/v1/stage-runs` serves one
(`stageRunDTO.EndedAt`, `stats/internal/api/stats.go`).

**Finding 3 — the run page crashes on an object `overrides` (flow-fix).** `lines()` in
`stats/web/src/components/DecisionsPanel.tsx` runs `(d.overrides ?? []).flatMap(...)`. Decision
rows whose `overrides` is a JSON object: 46 hold `{}` and 2 hold a legacy role-to-reason map; the
newest of them was written 2026-09-28.
<!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "select d.decision->'overrides', count(*) from decisions d where jsonb_typeof(d.decision->'overrides')='object' group by 1" @ dev store 2026-10-30 (a live read) -->

**Finding 4 — a failed finished read is silent (flow-fix).** In
`stats/web/src/views/StateBoard.tsx`, the only error text renders while `rows === null`; once the
open poll answers, `rows` is set and a `finished` read in `status: "error"` leaves
`lastFinished` null, so the Finished tile and "Cost in period" read "—" with no message.

**Finding 5 — gated reviewer findings never reach the store (flow-stats-app).**
**The gated per-task reviewer** (`skills/flow/implement.md`) records the bundle as one dispatch
row and keeps each pass's findings in its report file only. The Findings KPI
(`stats/internal/store/aggregate.go`, lateral `cim`) counts every `findings` row of the change
with no slot filter, so recording the rows is enough for the count. Every panel guard reading
findings is slot-blind, so naively recorded gated rows would be refused:
`check-panel-findings-closed` (a `fixed` Critical/Important with no later `panel-<round>-<slot>`
re-run, no report file, no `regression-test:` line), `uwUnprovedFixRounds` (no mutation row for the
round), `check-panel-reproducers` (an Important with the `none — <reason>` exemption); `flow record
render -kind panel` would list them in the panel record; and `flow record next-round` /
`highestRecordedRound` would move the panel's opening round past 0.

## Decisions

### What the request-end guard additionally requires

**ID:** verify-required-on-landed
**Status:** active
**Chosen:** every `landed` end requires the request's latest `/flow-interactive` `flow.verify` stage run carrying `-request <n>`, by start, to have ended `completed`, whatever the range touches — read after the git checks and before the `check-visual-trigger` branch; with the store unreachable both stage-run reads are skipped under one `⚠` line naming the verify and UI check reads.
**Considered:** requiring every per-request stage (`flow.load-context`, `flow.sdd-tdd`, `flow.landing-routes`, `flow.refresh-main-checkout`) — four more reads for stages whose absence costs only time attribution, while `flow.verify` is the one whose absence lets a red request land; a script wrapping steps 4–6 — the pass's angle 4 judged the guard the simpler close.

### What counts as a chained UI check

**ID:** chained-mark-under-one-second
**Status:** active
**Chosen:** a `completed` `flow.visual-verify` run whose `endedAt` is less than one second after its `startedAt` is refused, naming step 7; `client.StageRunRow` gains `EndedAt`. A screenshot read between two tool calls takes seconds; a chained pair measured 31 ms.
<!-- measured: the Context section's psql read of kan-1069-interactive-session's stage_runs @ dev store 2026-10-30 (a live read) -->
**Considered:** `endedAt == startedAt` as the pass report proposed — the store keeps microseconds, so a chained pair is never equal and the check would never fire; ten seconds — the shortest honest check measured 11 s, too close to refuse real work; extending the check to `flow.verify` — the skill states the never-chained rule for the UI check alone, and no evidence shows a chained verify mark.

### Requests already landed without a verify mark

**ID:** no-row-repair-for-landed-requests
**Status:** active
**Chosen:** no store rows are written for requests 2–12 of `kan-1069-interactive-session`: the defect wrote nothing wrong, it let a missing record pass, and a back-filled `flow.verify` row would claim a verify that never ran.
**Considered:** back-filling `flow.verify` rows — fabricates a measurement; ending those requests again — they are already `landed`, and the guard runs before the write.

### Reading `overrides`

**ID:** overrides-array-only
**Status:** active
**Chosen:** `Array.isArray(d.overrides) ? d.overrides : []` in `lines()`, and `Decision.overrides` typed as possibly non-array; the 48 object-shaped rows are the legacy shape and stay as written.
**Considered:** rendering the legacy role-to-reason map's keys as rows — no caller asked for it and the map names no model or effort to show; rewriting the stored rows to `[]` — destroys the two legacy rows' recorded reasons for a reader-side defect; an SPA-wide error boundary — a second change to the shell for one crash the guard closes.

### Where the state board shows a read error

**ID:** board-error-beside-heading
**Status:** active
**Chosen:** once rows render, a `<p className="lb-sub" role="alert">Could not load the board: <message></p>` under the heading whenever the open poll or the finished read is in `status: "error"` — the same text and class the `rows === null` branch already uses; a later successful read clears it.
**Considered:** a retry timer for the finished read — the read already re-runs on every period chip and every change that finishes, and a timer adds a second poll the board was split to avoid; a message on the Finished tile alone — a failed open poll after a first success is just as silent today.

### How a gated finding is recorded

**ID:** gated-findings-recorded
**Status:** active
**Chosen:** the parent records each gated pass's Critical, Important and Minor findings in one `flow record finding -from <jsonl>` call per bundle, every row under `-slot gated-reviewer -round 0 -task <n> -dispatch-seq <the bundle's seq> -status open -reproducer 'none — gated per-task review; its report file holds the evidence'`, refs from `flow record next-ref`; a clean re-review pass closes its task's rows `fixed`, an inline Minor fix closes its row `fixed`, an unresolvable Minor is `withdrawn <reason>`. Prose in `skills/flow/implement.md` and `skills/flow/gated-review-fix.md`.
**Considered:** a separate gated-findings table — a second store, a second KPI query and a migration for what the `findings` table and its slot column already hold; recording only Critical and Important — the KPI counts Minors, and the panel records every severity.

### Keeping the panel's guards off gated rows

**ID:** gated-rows-skipped-by-panel-guards
**Status:** active
**Chosen:** one exported constant, `records.GatedReviewerSlot = "gated-reviewer"` (`stats/internal/records`), and every panel-scoped reader drops rows carrying it: `check-panel-findings-closed` (all its checks), `uwUnprovedFixRounds`, `check-panel-reproducers`, `flow record render -kind panel`, and `flow record next-round` / `highestRecordedRound`. `check-unfinished-work`'s open-finding check still counts gated rows, so a gated finding left open still stops integrate.
**Considered:** treating a gated row as covered by a clean `task-…-reviewer-fix-<k>` dispatch — needs a task-to-row mapping, panel-style report files and `regression-test:` lines the gated path does not write; a round sentinel such as `-1` outside the panel's range — every guard still reads the row, and the reproducers and render readers ignore rounds; recording gated rows under a separate change — splits one run's findings across two change names.

### History of gated findings

**ID:** no-gated-history-backfill
**Status:** active
**Chosen:** earlier runs' gated findings are not back-filled: they live only in report files under each run's gitignored `.superpowers/sdd/`, removed at cleanup, so there is no record to repair from.
**Considered:** parsing archived self-review bundles — they carry no gated report files.

### The Findings KPI hint

**ID:** findings-kpi-no-hint
**Status:** active
**Chosen:** `RunDetail.tsx`'s Findings `<Kpi>` loses its `hint` prop.
**Considered:** rewording the hint to "panel and gated reviews" — the label already says Findings, and the run's findings are now all recorded.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 5 changes five slot-blind guards and readers whose refusals gate every `/flow` run, and Task 1 reorders a landing guard's reads.
**Considered:** `flow-low` — each task is test-pinned, but a missed reader in Task 5 refuses every run that records a gated finding.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** two fixer dispatches in series — group 1: Tasks 1, 2, 3, 4 (the request-end guard and two SPA one-liners, each small and test-pinned); group 2: Task 5 (the gated findings across prose, guards, CLI and SPA).
**Considered:** one fixer for all five — Task 5 alone spans eight-plus files across four packages and would start on a context the first four already filled; five fixers — four extra dispatches for tasks of a few lines each.

## Open questions

## Recommended approach

Fix each defect where every caller routes through it: the guard function for findings 1–2
(`verify-required-on-landed`, `chained-mark-under-one-second`), the one reader for finding 3
(`overrides-array-only`), the board's single render path for finding 4
(`board-error-beside-heading`), and for finding 5 one recorded slot plus one constant every
panel-scoped reader skips (`gated-findings-recorded`, `gated-rows-skipped-by-panel-guards`). The
rejected alternatives — wider stage requirements, equality as the chained test, an error boundary,
a retry timer, a gated-findings table, a round sentinel — are recorded with each decision.

## Critical files

- `stats/cmd/flow/request.go` — `gateLandedRequest`, the request-end guard.
- `stats/cmd/flow/request_test.go` — the guard's tests and fake daemon.
- `stats/internal/client/client.go` — `StageRunRow`, read by the guard.
- `skills/flow-interactive/SKILL.md` — **Request-end guard** prose.
- `stats/web/src/components/DecisionsPanel.tsx`, `DecisionsPanel.test.tsx` — the run page's decisions panel.
- `stats/web/src/views/StateBoard.tsx`, `stats/web/src/views/views.test.tsx` — the live state board.
- `skills/flow/implement.md`, `skills/flow/gated-review-fix.md` — the gated reviewer's recording prose.
- `stats/internal/records/` — home of `GatedReviewerSlot`; `render.go` renders the panel record.
- `stats/internal/guard/panelfindingsclosed.go`, `unfinishedwork.go`, `panelreproducers.go` and their tests.
- `stats/cmd/flow/record.go`, `record_test.go` — `next-round` and `highestRecordedRound`.
- `stats/web/src/views/RunDetail.tsx` — the Findings KPI.

## Existing code to reuse

- `callRecord` and `client.ListStageRuns` (`stats/cmd/flow/`, `stats/internal/client/client.go`) — the stage-run read the UI check already makes; the verify read is the same call with `flow.verify`.
- The `rows === null` error line of `StateBoard.tsx` — its text and class are reused for the new line.
- `flow record finding -from <jsonl>`, `flow record next-ref`, `flow record status` — the panel's recording verbs, used unchanged for gated rows.
- `newRequestFixture`, `requestDaemon`, `runRequest` (`stats/cmd/flow/request_test.go`) — the guard's test harness.

## Verification

- `cd stats && make -s web-dist && go test -race -count=1 ./cmd/flow/ ./internal/guard/ ./internal/records/ ./internal/client/` and `cd stats/web && npm test` pass.
- `scripts/run-verify-list.sh <worktree> lint` is green, `check-verbatim-moves.sh` acknowledging this change's `verbatim-moves.txt`.
- The **Live check** below runs against a copy of the dev store.

## Live check

Against a copy of the real store, never the dev stack: `scripts/live-store-copy.sh up
sr-kan-1069` from this worktree, exporting the `FLOW_ADDR` / `FLOW_RECORDS_ADDR` it prints, and the
worktree's own `flow` build (`cd stats && go build -o /tmp/flow-sr-kan-1069 ./cmd/flow`).

1. `flow record request end -C <this worktree> -change kan-1069-interactive-session -n 5 -outcome landed -commit 4cd99eb23d2bde12796fda40c7313d231c32eabb -base efc23db3c230b6c6d12fd25a468b3dd224f9df1d` — before: the installed `flow` passes the guard and writes (do not run it against the dev store); after: exit 1, one line naming `flow.verify`. Not working looks like exit 0 or a `recorded:` line.
2. The same for `-n 3 -commit b1d6446f836ac45f9ac9c7dcb9a6551269034262 -base f6edb293284c52f7190ab38f799860ef3353b0f3` after inserting, into the copy only, a completed `flow.verify` row for request 3 — after: exit 1, one line naming the chained UI check. Not working: exit 0.
3. The copy's SPA at the printed address, run page of `agents-a740d89c` / `withdraw-changes-abandoned-before-planning`: renders the Decisions panel, no blank page. Not working: an empty `#root`.
4. `scripts/live-store-copy.sh down sr-kan-1069`.
