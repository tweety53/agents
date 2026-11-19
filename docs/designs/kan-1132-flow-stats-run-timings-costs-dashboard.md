# kan-1132-flow-stats-run-timings-costs-dashboard
> Design record of change kan-1132-flow-stats-run-timings-costs-dashboard (KAN-1132), written at landing. Each decision's own **Status:** line is the status.

## Context

KAN-1132. The evidence run is kan-1091-body-calories-measurements, project key `gymie-7c1f238a`,
still `STARTED`; its `tasks.md` and verify reports live in its gymie and gymie-playwright
worktrees. Read against the dev store before planning:

- 183 dispatch rows at planning round 1, 4 with a stage link — exactly the 4 that auto-opened a
  run; every session token matches its stage runs. 188 rows at writing, 184 unlinked.
  <!-- measured: curl -s http://127.0.0.1:4173/api/v1/records/gymie-7c1f238a/kan-1091-body-calories-measurements | count dispatches, stageRunId null @ 2026-11-18 dev store -->
- 63 stage runs, 9 `superseded`, 1 `abandoned`.
- 30 rows without tokens: 29 `agentId: inline` verifier rows with `metrics: {}`, and
  `task-13+14+16+20-reviewer` (transcript present, 35 usage entries, inside its window). 4 more
  rows carry tokens but no price.
- 55 `verify_runs` rows, all `run-verify-list`. 14 tasks carry an `**Opener:**`.
- `guard_runs` (`stats/internal/store/migrations/0032_guard_runs.sql`) holds `project_key, guard,
  worktree, exit_code, outcome, duration_ms, recorded_at`; `worktree` is `os.Getwd()` at the
  guard's start (`workingDir`, `stats/cmd/flow-guard/telemetry.go`). Since 2026-11-14 for gymie,
  1,818 runs (64,153 s summed) carry a `…-worktrees/kan-1091-…` path, and 2,732 ran in the main
  checkout `/Users/tweety53/Projects/gymie`, among them 59 `run-verify-list` (15,650 s), 109
  `visual-verify-pass`, 69 `close-task`, 8 `recapture`; those guards take the worktree as an
  argument. Guard runs inside a dispatch's window are mostly the agent's own (665 `run-capped`
  runs, 29,697 s; 124 `break-and-prove`), and `guard_runs` carries no session or agent id.
- `stats/web/src/theme.css` serves a light variant through `@media (prefers-color-scheme: light)`
  and `:root[data-theme="light"]`; bare `:root` is dark and nothing in the SPA sets `data-theme`.
  The `:root[data-theme="light"]` block sets no `--marker` or `--heat`, so a forced-light page
  keeps the dark values, and the media-query block declares both twice.

History since the run (`git log --since=2026-11-16 -- stats hooks`): `3c6a553e`/`c8684612` added
`flow record run` for run-verify-list and recapture only; `4cf2cf26` selects recapture by test but
a title-only task still selects nothing; `9537d2ed`/`384437b0` fixed two harvest paths, not the
two failure shapes above; `c35bb2d9` aborts a superseded open verify row. No gap but gap 9 is
closed. **Gap 9 is already fixed** and planned no task: a journalled dispatch keeps its supplied
instants on replay (`8ca877ee`, pinned by `TestRecordEntryWithAZeroInstantIsStampedOnReplay`'s
"supplied instants are kept", `stats/internal/reconcile/record_test.go`), and `flow record
handoff-lines` prints the journal count on its `**Records:**` line (`5b64e1a9`, pinned by
`TestRecordHandoffLinesFormat`, `stats/cmd/flow/record_test.go`).

**The visual spec is the approved prototype**,
`.superpowers/sdd/kan-1132-prototype.html` in this change's worktree, published as
https://claude.ai/artifact/3tBVekQtmEhGxnW7ps8z5T and approved by the operator at design approval.
It draws run detail's header with its one new button, the run's separate timings and costs
page, and the Run costs view with Compare, in the app's light variant on the app's own shell and
`lb-*` classes; `flow.visual-verify` compares against it, and every capture of a view it draws
drives the interactive states it shows (the button's hover, the group-by chips, the period and
project chips, the Compare pickers, a metric's definition tooltip on hover). Run detail's
unchanged panels are not drawn: its header is a region the prototype draws alone, captured as an
element beside the view's full-page capture.

This is one change because every dashboard figure rests on the recording fixes: a timeline with no
stage links, a category table with no categories, or a total with unexplained unattributed rows
would draw numbers the ticket's acceptance line refuses.

## Decisions

### Category table in Go, Opener read live then saved

**ID:** category-table-go-records
**Status:** active
**Chosen:** one ordered table in `stats/internal/records/category.go`, classifying at read time; Opener from the worktree tasks.md, then the saved copy.
**Considered:** a TypeScript classifier beside `describeDispatch.ts` — cannot read N worktrees' tasks.md for the period view; a stored category column — never re-labels history, and the ticket forbids it.

### Categories beyond the ticket's twelve

**ID:** extra-key-families
**Status:** active
**Chosen:** add rows for every other documented key family (full-suite fixer, plan reviewer, self-review planner/reviewer/fixer, pipeline-fix planner/fixer/reviewer, operator request) plus an `other:<role>` fallback listed in data quality.
**Considered:** the twelve only — 7 kan-1091 rows (`full-suite-fix-*`, `plan-review-*`) would carry no category, failing the acceptance line.

### Time share sums to wall clock

**ID:** time-share-sweep
**Status:** active
**Chosen:** split each instant equally among the dispatch rows active in it; parent commands and idle fill the rest; show busy (own union) beside it.
**Considered:** per-row union only — rows would not sum to the total, failing "dashboard totals equal the per-row sums"; per-row sum — counts overlap twice, which the ticket forbids.

### Group key cost split evenly

**ID:** group-cost-split-even
**Status:** active
**Chosen:** a `task-13+14-…` row's cost and time split evenly across its task ids, marked estimated.
**Considered:** showing the group as its own row — leaves "cost per task" unanswered for grouped tasks.

### Repo from the key's suffix

**ID:** repo-from-key-suffix
**Status:** active
**Chosen:** the longest `-<repo>` key segment matching a project root basename; otherwise "change-wide".
**Considered:** storing a repo per dispatch row — a schema change for a value the key already carries.

### Attribution: fix at source, stamp the rest

**ID:** unattributed-reason-in-metrics
**Status:** active
**Chosen:** fix the two kan-1091 failure shapes, and stamp a reason into `metrics` at each remaining give-up site.
**Considered:** derive the reason at read — only two of the causes are derivable from the row.

### Recapture resolves titles, else fails loudly

**ID:** recapture-title-search
**Status:** active
**Chosen:** find specs whose source contains a named title; none → exit non-zero with one line.
**Considered:** Files-adjacent specs — a guess that can recapture unrelated specs; a loud error alone — leaves a correct task unable to recapture.

### Begin auto-closes, never refuses

**ID:** stage-begin-auto-closes
**Status:** active
**Chosen:** keep begin's auto-close of an open run, with its own end instant and warning.
**Considered:** refuse begin — breaks the never-block contract whenever an end was journalled while the store was down.

### Flaky count on verify runs

**ID:** verify-run-flaky-column
**Status:** active
**Chosen:** a nullable `flaky` count on `verify_runs`, written by run-verify-list through `flow record run -flaky`.
**Considered:** show flaky retries as "not recorded" — leaves a ticket count permanently empty.

### Journal counted by flowd

**ID:** flowd-counts-journal
**Status:** active
**Chosen:** flowd counts the change's journal file at its own state root; unreadable → "unknown" in data quality.
**Considered:** the CLI reporting the count to flowd — a write path for a value flowd can read.

### Category doc in the rationale file

**ID:** category-doc-rationale
**Status:** active
**Chosen:** `skills/flow-contracts/session-records-rationale.md` — never loaded by a run.
**Considered:** `session-records.md` — loaded by runs, costing tokens on every run for a maintainer's note.

### Dashboard visuals from the prototype

**ID:** prototype-is-the-spec
**Status:** active
**Chosen:** the prototype file, on the app's own shell and `lb-*` classes, is the spec `flow.visual-verify` compares against once approved.
**Considered:** building from the existing `lb-*` views without a mockup — lost because the operator approved the prototype as the visual spec, and `flow.visual-verify` needs a reference to compare against.

### Go computes the dashboard

**ID:** dashboard-go-computes
**Status:** superseded by run-costs-separate-page
**Chosen:** `GET /api/v1/stats/run/{p}/{c}` and a `run-costs` stats view built from one Go function; RunDetail gains the per-run panels and lanes; a new Run costs view carries across-runs and Compare. Operator, round 1 Q1.
**Considered:** the SPA computing in RunDetail — no access to worktree tasks.md or reports for the across-runs view, totals testable only in the browser; a separate per-run page — two pages drawing one run's timeline can disagree.
**Superseded because:** the operator, after design approval: "current run's page will be too heavy. Need a separate button on run page" (`.superpowers/sdd/planner/answers-3.md`).

### Plan check at hand-back

**ID:** gated-bracket-handback
**Status:** active
**Chosen:** the `SubagentStop` hook runs `flow record dispatch handback`, which checks the plan at the agent's stop and leaves a marker the parent's end trusts; no marker → today's live check. Operator, round 1 Q2.
**Considered:** widening the tolerated diff — a reviewer's own tasks.md edit is indistinguishable from the parent's; the hook ending the row — the outcome, report and commit only the parent knows make a second end.

### Parent time from guard_runs

**ID:** parent-time-guard-runs
**Status:** active
**Chosen:** flowd reads the existing `guard_runs`, attributing a run by its worktree path. Operator, round 1 Q3.
**Considered:** `flow record run` for every guard — duplicates `guard_runs` row for row and still cannot name the change outside a worktree.

### Parent commands are guard time outside every dispatch

**ID:** parent-commands-outside-dispatches
**Status:** active
**Chosen:** parent commands = union of the change's guard runs minus union of its dispatch rows; a run nested in another collapses into it.
**Considered:** every guard run as parent time — `guard_runs` carries no agent id, so an agent's own `run-capped` and the parent's `await-sub-agents` would be counted a second time beside the dispatch they ran in; a parent-only list of guard names — `run-verify-list` and `recapture` run both from the parent and from agents.

### Back-fill by migration

**ID:** stage-link-backfill-migration
**Status:** active
**Chosen:** one idempotent migration `UPDATE` by the same `dispatch_stage_run` rule; ambiguous rows stay NULL, listed in data quality. Operator, round 1 Q4.
**Considered:** deriving history's links at every read — every reader repeats the rule; a manual maintenance verb — history stays unlinked until someone runs it.

### Auto-close keeps `superseded`

**ID:** stage-close-superseded
**Status:** active
**Chosen:** begin's close of an open run keeps outcome `superseded`; `abandoned` stays the sweeper's word. Operator, round 1 Q5, over the ticket's "abandoned" wording.
**Considered:** `abandoned` for new closes — the sweeper's and begin's closes become indistinguishable; re-labelling history — the same, plus a migration over stored outcomes.

### Retime audit in a JSONB column, old verb removed

**ID:** retime-jsonb-drop-old
**Status:** active
**Chosen:** a nullable `retimed` JSONB column `{reason, at, startedAt, endedAt}`, original instants kept across retimes; `flowd retime-dispatch` deleted. Operator, round 1 Q6.
**Considered:** a `dispatch_retimes` table — every intermediate instant kept, a table for an audit the dashboard reads one line of; keeping the old verb — leaves an unaudited path.

### Compare compares two changes

**ID:** compare-two-changes
**Status:** active
**Chosen:** pick any two runs in Run costs; every metric and breakdown row side by side with Δ and Δ%. Operator, round 1 Q7.
**Considered:** two tasks — a per-task view the ticket's dashboard list does not otherwise carry; both — the task picker is a second compare for the same question.

### Prototype in the app's light variant

**ID:** prototype-light-variant
**Status:** active
**Chosen:** the prototype's tokens are `theme.css`'s `:root[data-theme="light"]` values, unchanged otherwise; the new views' visual specs run under `colorScheme: "light"`. Operator, round 1 message "design must be white, not black".
**Considered:** hand-picked white values — a standalone restyle the prototype rule forbids.

### Complete the forced-light tokens

**ID:** theme-light-marker-heat
**Status:** active
**Chosen:** `:root[data-theme="light"]` gains `--marker: #8250df` and `--heat: 31, 111, 235`; the media-query block's duplicate pair is dropped; one test pins the forced-light values.
**Considered:** leaving it — a forced-light page draws markers and heat cells in the dark palette.

### A guard run records the worktree it was pointed at

**ID:** guard-run-worktree-arg
**Status:** active
**Chosen:** flow-guard's telemetry records, as a new `guard_runs` row's `worktree`, the first argument that is an existing directory under a `…-worktrees/<change>` path, else `os.Getwd()` as today; existing rows keep what they hold. Operator, round 2 Q1.
**Considered:** a new nullable `target_worktree` column beside the unchanged `worktree` — a migration and a wider write for the same attribution; leaving telemetry as is — every run driven from the main checkout keeps its command time in idle / parent, so "complete timings" stays short.

### New views follow the app's existing light variant

**ID:** new-views-follow-light-variant
**Status:** active
**Chosen:** every colour of the new panels and the Run costs view is a `theme.css` token, so they are white wherever the app already is (OS light mode, or `data-theme="light"`); no other view changes. Operator, round 2 Q2.
**Considered:** the whole app always light — every existing view turns light and every dark baseline capture is re-recorded; the new views forced light inside a dark app — white panels inside a dark RunDetail.

### Per-run timings and costs on a separate page, opened from run detail

**ID:** run-costs-separate-page
**Status:** active
**Chosen:** Go computes, as `dashboard-go-computes` chose — `GET /api/v1/stats/run/{p}/{c}` and a `run-costs` stats view built from one Go function — but RunDetail gains no panel and no timeline lane: it gains one button, **Timings and costs →**, opening a separate per-run page at `#/run-costs/<project>/<change>` that carries the totals strip, the timeline with category, parent-command and idle lanes, the breakdowns, data quality, counts, cost per task and per verify round. The across-runs **Run costs** view at `#/run-costs` and Compare stay as designed, its runs table linking to each run's page. Operator, after design approval (`.superpowers/sdd/planner/answers-3.md`).
**Considered:** the panels inside RunDetail (`dashboard-go-computes`) — the run page becomes too heavy, the operator's reason; the earlier objection to a separate page, two pages drawing one run's timeline that can disagree, is met by the new page drawing its stage bars from the same Go result as its lanes, while RunDetail's own timeline is left unchanged.

## Open questions

## Recommended approach

**Recording fixes (stats Go, hooks).**

1. **Stage link (gap 1).** A SQL function `dispatch_stage_run(change_id, token, at)`: the
   candidates are the change's stage runs whose half-open window `[started_at, ended_at)` holds
   `at`; exactly one with the same session token wins, else exactly one candidate in total, else
   NULL. `recordDispatchTx` (`stats/internal/store/records.go`) calls it before
   `openAutoStageRun` and auto-opens only on NULL. Matching by the begin instant, not "the run open
   now", keeps a journal replay attached to the run that held it. `DispatchWindowsForSession` drops
   its `stage_run_id IS NULL OR auto_opened` filter, so a linked never-ended row keeps its
   next-begin cap. History is back-filled by a migration running one idempotent
   `UPDATE dispatches SET stage_run_id = dispatch_stage_run(…) WHERE stage_run_id IS NULL` at
   flowd's next start; ambiguous rows stay NULL and are listed in data quality
   (`stage-link-backfill-migration`).
2. **Stage marks (gap 2).** `stats/cmd/flow/stage.go` prints one line,
   `flow: stage <verb>: <error> (flow stage -h for usage)`, on every error, and the usage block
   only on `-h`; `flow stage end` with no open run exits 1. Begin keeps auto-closing an open run
   with its own end instant (the new begin's start), outcome `superseded`, and its warning
   (`stage-begin-auto-closes`, `stage-close-superseded`). `flow record handoff-lines` gains a
   `**Stages:**` line, `0 open` or `N open — <stage> (run <id>), …`; the data-quality panel lists
   open and superseded runs too.
3. **Gated-reviewer bracket (gap 3).** `hooks/await-own-agents.py`, on its non-blocking path, runs
   a new `flow record dispatch handback`, which runs `check-plan-unchanged verify` at the instant
   the agent stops and leaves a marker beside the dispatch's snapshot, keyed by that dispatch row's id, so a re-dispatch under the same key (a new row) never inherits an earlier pass's marker; the parent's later
   `flow record dispatch end` trusts a passing marker instead of re-checking the live plan, and a
   failing marker refuses the end as today's check does. With no marker (a harness with no
   `SubagentStop` hook — ZCode, Muse) the end runs today's live check, and its refusal message says
   to end the row before editing the plan. `aborted` stays only for a dispatch that never ran
   (`gated-bracket-handback`).
4. **Categories (gap 4).** One Go table, `stats/internal/records/category.go`: ordered rows
   `{category id, label, key regexp, opener prefix}`, first match wins, nothing stored per row
   (`category-table-go-records`, `extra-key-families`). A task implementer key is classed by its
   first task id's `**Opener:**` — `visual-verify…` → in-run fix from visual verify, `verify…` →
   in-run fix from verify, none → planned implementer. The Opener is read from the change's live
   worktree `spectre/changes/<name>/tasks.md` first, then the copy `flow change save` kept
   (`readPlanFile` and `savedFile`, `stats/internal/api/stats.go`). Unmatched keys class as
   `other:<role>` and an unreadable tasks.md is its own item, both listed in data quality.
5. **Attribution (gap 5).** Both kan-1091 failure shapes are reproduced as harvest fixtures — an
   inline verify row the settle never patches, and an agent row whose transcript is harvested
   before its id is stamped — and fixed at their source; every remaining give-up site stamps a
   reason into the row's `metrics` bag, as `MarkDispatchesUnattributed` already does
   (`unattributed-reason-in-metrics`). `records.CostReason(d)` becomes the one precedence that
   `CostStatusOf`, the render and the dashboard share.
6. **Parent-side time (gap 6).** flowd reads the existing `guard_runs` (`parent-time-guard-runs`):
   a run belongs to the change whose `-worktrees/<change>` segment its `worktree` path carries, in
   any repository's worktree directory, and starts at `recorded_at` − `duration_ms`. New rows name
   the worktree a guard was pointed at (`guard-run-worktree-arg`, `workingDir` in
   `stats/cmd/flow-guard/telemetry.go`). A run nested wholly inside another of the same change
   collapses into the outer one. Parent commands are the union of the change's guard runs minus
   the union of its dispatch rows, and idle / parent is wall clock minus the union of dispatch rows
   and parent commands (`parent-commands-outside-dispatches`), both computed at read and marked
   derived. A guard run flowd could not record — telemetry is not journalled — lands in idle /
   parent.
7. **Late begins (gap 7).** `flow record dispatch retime`, over a new flowd route, reason
   required, writes the nullable `retimed` JSONB column on `dispatches`: `startedAt`/`endedAt` the
   original instants, written by the first retime and kept across later ones; `reason` and `at`
   the latest retime's (`retime-jsonb-drop-old`). `flowd retime-dispatch` and its store path are
   deleted; `ErrInlineBegunAfterCommit` (`stats/internal/store/records.go`) and close-task's
   message (`stats/internal/guard/closetask.go`) point at the new verb.
8. **Recapture and Files (gap 8).** A task whose Tests/Regression name titles but no path selects
   every spec under the configured spec root whose source contains that title; if still none,
   recapture exits non-zero with one line naming the task (`recapture-title-search`).
   `flow tasks files … add|remove` refuse an empty path or one carrying whitespace or a newline,
   exit 1, one line.
9. **Flaky retries.** run-verify-list, which already parses its flaky count, passes it through
   `flow record run -flaky` into a nullable `flaky` column on `verify_runs`
   (`verify-run-flaky-column`).

**Dashboard (`stats/internal/api`, `stats/web`), as the prototype draws it**
(`run-costs-separate-page`, `prototype-is-the-spec`).

- One Go function computes a run's result from rows flowd already holds;
  `GET /api/v1/stats/run/{project}/{change}` serves it, and a `run-costs` stats view folds it over
  a period across changes, so the two never disagree (`run-costs-separate-page`). Compare fetches
  the one-run result for two changes and joins client-side (`compare-two-changes`).
- **Time share:** a sweep over the dispatch intervals splits each instant equally among the rows
  active in it; parent commands and idle / parent fill the instants no dispatch holds, so every
  breakdown's rows sum exactly to wall clock. Each row also shows its **busy** time — the union of
  its own intervals, not additive, no total (`time-share-sweep`).
- **Cost:** each row's `cost_usd`; the parent session's own cost is stage cost minus the dispatch
  buckets, moving RunDetail's `residualCosts` definition (`stats/web/src/views/RunDetail.tsx`)
  server-side; totals are the row sums by construction, and a data-quality item fires when they
  differ from the stage-run cost by $0.01 or more.
- **Breakdowns:** category, role, model/effort, repo (`repo-from-key-suffix`), fix round (loop ×
  round from the key or the task's Opener); cost per task (`group-cost-split-even`) and per verify
  round (its verify rows plus the in-run fix work whose Opener names it); visual-verify loop share.
- **Counts:** fix rounds; departures per visual round, from the `departures:` line of each verify
  report, worktree first, then the saved `reports.<repo>.…` copy; findings per canonical severity;
  flaky retries; open, aborted and blocked rows; journal pending (`flowd-counts-journal`).
- **Data quality:** open and superseded stage runs, unlinked rows, rows without tokens with their
  reason, aborted rows that ran, uncategorised keys, unreadable openers, unreconciled cost, journal
  pending, retimed rows, and the project's main-checkout guard runs inside the run's window that no
  change could claim.
- **Every metric** carries `definition` and `kind` (`measured|derived|estimated`); the UI shows the
  kind as a tag beside the label with the definition as its tooltip, and a Definitions panel lists
  them all.
- **UI:** RunDetail gains one `lb-fchip` button, **Timings and costs →**, beside its state pill,
  and nothing else. It opens a new per-run page, `#/run-costs/<project>/<change>`, its crumb leading
  back to run detail, carrying a totals strip, a Timeline panel (a new `RunTimeline` component drawn
  at the prototype's geometry: the stage lanes, one lane per category, parent commands, idle; `StageTimeline`
  is unchanged), a group-by breakdown table, data
  quality, counts, cost per task and per verify round, and a link into Compare. A new top-level
  view **Run costs**, `#/run-costs`, carries the period and project chips, the category table
  across runs, the runs table — each run linking to its per-run page — Compare and the Definitions
  panel. No new dependency; hand-rolled SVG as every existing chart. Every colour is a
  `theme.css` token (`new-views-follow-light-variant`); the new views' visual specs run under
  `test.use({ colorScheme: "light" })`, as `stats/web/tests/visual/stage-leaderboard.spec.ts` does
  (`prototype-light-variant`), and `theme.css`'s forced-light block is completed
  (`theme-light-marker-heat`).

**Docs.** The category table's location and how to add a category, one paragraph in
`skills/flow-contracts/session-records-rationale.md` (`category-doc-rationale`); the `**Stages:**`
line in `skills/flow-contracts/handoff-blocks.md`; `README.md`'s Stats app section for the new view;
`stats/README.md` for the new verbs and the removed `flowd retime-dispatch`.

## Critical files

- `stats/internal/store/records.go` — `recordDispatchTx`, `openAutoStageRun`,
  `DispatchWindowsForSession`, `MarkDispatchesUnattributed`, `ErrInlineBegunAfterCommit`, the
  retime store path to delete and the new retime write.
- `stats/internal/store/migrations/` — new migrations: `dispatch_stage_run` and the back-fill, the
  `retimed` column on `dispatches`, the `flaky` column on `verify_runs` (the last is `0068_verify_runs.sql`).
- `stats/cmd/flow/stage.go` — one-line errors, `stage end` exit codes.
- `stats/cmd/flow/record.go` — `handoff-lines`' `**Stages:**` line, `dispatch handback`,
  `dispatch retime`, `run -flaky`, the end's marker check (`verifyPlanTreeForEnd`).
- `stats/cmd/flow/tasks.go` — `tasks files` path validation.
- `stats/cmd/flowd/main.go` — removal of `retime-dispatch`.
- `stats/internal/api/server.go`, `stats/internal/api/records.go` — the retime and hand-back
  routes, the per-run route.
- `stats/internal/api/stats.go` — the `run-costs` view, `readPlanFile`/`savedFile` reuse.
- `stats/internal/api/runtimings.go` (new) — the per-run computation.
- `stats/internal/records/category.go` (new), `stats/internal/records/types.go` — the category
  table, `CostReason`, `CostStatusOf`.
- `stats/internal/harvest/` — the two attribution failure shapes and reason stamping.
- `stats/cmd/flow-guard/telemetry.go` — `workingDir` → worktree argument first.
- `stats/internal/guard/planunchanged.go` — the check the hand-back runs.
- `stats/internal/guard/recapture.go` — title search, loud failure.
- `stats/internal/guard/runverifylist.go` — passes the flaky count.
- `stats/internal/guard/closetask.go` — message names the new retime verb.
- `hooks/await-own-agents.py` — calls `flow record dispatch handback` on its non-blocking path.
- `stats/web/src/views/RunDetail.tsx` — the one button.
- `stats/web/src/views/RunTimings.tsx` (new), `stats/web/src/views/RunCosts.tsx` (new) — the
  per-run page and the across-runs view.
- `stats/web/src/components/RunTimeline.tsx` (new) — the per-run page's timeline, at the prototype's
  geometry (lane pitch, bar height, `+Nh` axis), which `StageTimeline`'s does not share.
- `stats/web/src/App.tsx`, `stats/web/src/api.ts` — the two routes, the nav entry, the fetches.
- `stats/web/src/theme.css` — forced-light `--marker`/`--heat`.
- `stats/cmd/uitest-seed/seed.go` — seeded data for the new panels' captures.
- `stats/web/tests/visual/` — light-scheme capture specs of the new panels and view.
- `skills/flow-contracts/session-records-rationale.md`, `skills/flow-contracts/handoff-blocks.md`,
  `README.md`, `stats/README.md` — docs.

## Existing code to reuse

- `openAutoStageRun` and `recordDispatchTx` (`stats/internal/store/records.go`) — the stage-link
  call slots in before the auto-open.
- `MarkDispatchesUnattributed` (`stats/internal/store/records.go`) — the reason-stamping pattern.
- `CostStatusOf` (`stats/internal/records/types.go`) — folded onto `CostReason`.
- `readPlanFile`, `savedFile`, `planTasks` (`stats/internal/api/stats.go`) — tasks.md live then
  saved, for Openers and per-task rows.
- `verifyPlanTreeForEnd` (`stats/cmd/flow/record.go`) and `check-plan-unchanged`
  (`stats/internal/guard/planunchanged.go`) — the check the hand-back moves earlier.
- `guard_runs` and its write in `stats/cmd/flow-guard/telemetry.go` — parent command time.
- run-verify-list's flaky parse (`stats/internal/guard/runverifylist.go`) and `flow record run`
  (`stats/cmd/flow/record.go`).
- The `GET /api/v1/stats/{view}` view switch (`stats/internal/api/stats.go`, e.g.
  `viewWeeklyCosts`) — the `run-costs` view's slot.
- `StageTimeline` (`stats/web/src/components/StageTimeline.tsx`), `describeDispatch`
  (`stats/web/src/lib/describeDispatch.ts`, the key shapes the category tests cover),
  `residualCosts` (`stats/web/src/views/RunDetail.tsx`), the `lb-*` classes and `theme.css`
  tokens.
- `stats/cmd/uitest-seed` and the light-scheme visual specs (`stage-leaderboard.spec.ts`,
  `review-dynamics.spec.ts`, `weekly-costs.spec.ts`) — the capture pattern.

## Verification

- `cd stats && go test ./...` and the flow-guard contract tests green, every gap's fixture failing
  on the base and passing on the change.
- `internal/api/runtimings_test.go` asserts over every breakdown generically: total == Σ rows for
  count, share and cost, and share total == wall clock; a guard run inside a dispatch window adds
  no parent-command time.
- The Go category table and `describeDispatch.ts` are run over one shared fixture list of dispatch keys (`stats/internal/records/testdata/dispatch-keys.json`), the Go test and a vitest case each asserting the key shapes they parse agree, so the two classifiers cannot drift silently.
- The migration's back-fill links an unambiguous fixture row and leaves an ambiguous one NULL.
- The hand-back marker admits a parent plan edit after it and refuses a reviewer's own edit before
  it.
- `cd stats && make -s web-dist && cd web && npx --no-install tsc -b`, vitest for the per-run
  page, the across-runs view, the breakdown table, Compare and run detail's button reaching the
  per-run route, and the light-scheme capture specs against the UI-test stack on 4174, compared to
  the prototype by `flow.visual-verify`.
- The real run: the `## Live check` below.

## Live check

Against a copy of the dev store, never the dev `flowd`: from this change's worktree run
`scripts/flow-guard live-store-copy up kan-1132-live`, which builds flowd from the change and
starts it on 4175 against `flow_kan-1132-live_live`, running the new migrations — the back-fill
included — on the copied rows.

- **Before** (dev store, read-only `GET http://127.0.0.1:4173/api/v1/records/gymie-7c1f238a/kan-1091-body-calories-measurements`):
  record the dispatch count and the count with `stageRunId` null (188 and 184 at planning).
  <!-- measured: curl -s http://127.0.0.1:4173/api/v1/records/gymie-7c1f238a/kan-1091-body-calories-measurements | count dispatches, stageRunId null @ 2026-11-18 dev store -->
- **After** (copy, `GET http://127.0.0.1:4175/api/v1/stats/run/gymie-7c1f238a/kan-1091-body-calories-measurements`
  and the copy's records route): record the unlinked count, the rows with neither tokens nor a
  stated reason, the open stage runs, the rows classed `other:<role>`, the share total against
  wall clock, and the cost total against Σ rows.
- **Working looks like:** every dispatch linked or listed as ambiguous in data quality; every row
  carrying a category from the table; every row with tokens or a reason; share total == wall
  clock and cost total == Σ rows exactly; the category table reproduced.
- **Not working looks like:** the after unlinked count equal to the before figure (the back-fill
  did not run); rows with no tokens and no reason; a share total differing from wall clock; any
  `other:` row in kan-1091.
- Open stage runs on kan-1091 are counted, not forced to zero: the run is still `STARTED`, and a
  run legitimately open in a live run is what the `**Stages:**` line exists to report.
- `scripts/flow-guard live-store-copy down kan-1132-live` afterwards.

## Non-goals

- No pair-judge data on the dashboard — requested mid-planning, then withdrawn by the operator.
- No stored category, repo or stage-link derivation per read — the decisions above store only
  what the record lacks.
- No change to existing views beyond RunDetail's one button and the forced-light token fix;
  RunDetail's own panels and timeline are untouched (`run-costs-separate-page`), and the app's
  default theme is unchanged (`new-views-follow-light-variant`).
- Existing `guard_runs` rows are not re-attributed; a main-checkout run before this change stays
  unclaimed and is counted in data quality.
- No caching of finished runs' results; each request computes.

## Risks

- **`run-costs` over a long period** computes every run on request; a period holding hundreds of
  runs may approach `statsWriteTimeout` (`stats/internal/api/server.go`). Limit: finished runs can
  be cached later; the view's request is bounded by the existing timeout.
- **The back-fill migration** writes stored rows; it is idempotent and touches only NULL links,
  and ambiguous rows stay NULL, so a wrong link can come only from a unique-window match. It runs
  one per-row `dispatch_stage_run(…)` call over every project's NULL rows inside flowd's start-up
  migration (184 of 188 on kan-1091 alone); a slow or failing back-fill delays or blocks the dev
  `flowd` start. Limit: the live check runs it first against a store copy and times it.
- **Two key parsers** — the Go category table and `describeDispatch.ts` read one key grammar; the
  shared fixture list in `## Verification` fails either side that drifts.
- **A harness with no `SubagentStop` hook** (ZCode, Muse) keeps today's end check; its refusal now
  says to end the row before editing the plan.
- **Guard runs flowd could not record** (store unreachable; telemetry is not journalled) land in
  idle / parent, so idle can over-read during an outage.
- **Openers and reports read from worktrees** vanish once a change is cleaned up; the saved copies
  are the fallback, and an unreadable tasks.md is a data-quality item rather than a guess.
