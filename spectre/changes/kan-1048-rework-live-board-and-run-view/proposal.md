# kan-1048-rework-live-board-and-run-view

## Why

The Live state board and the opened-run view of the flow stats app (KAN-1048) read as raw tables:
the shared dashboard bar and the per-column filter dropdowns are clumsy, nothing shows what is
running right now, the run view groups stage runs by the literal command text (`/flow` vs
`/flow kan-1048`), which says nothing, and a stage run's model shows without its version or effort.
The operator wants every dashboard to be useful; these two come first, the other eight follow in
a separate ticket.

## What changes

- **Live state board** — one chip filter bar (Project with counts, Updated 24h/7d/30d/All, With
  errors only, search, Clear filters) replaces the shared dashboard bar and the table's column
  filter dropdowns; clickable state tiles (In progress, Started, Finished) filter by state; a
  Running now strip draws one card per change with an open stage run (stage, elapsed, tasks
  done/total, cost); the table is grouped by state with collapsible sections (Finished collapsed),
  carries a state pill, current stage, task count with plan growth, relative update time, a cost
  bar and red errors. No next-command column.
- **Opened run** — no filters at all, no dashboard bar; a header (back link, change name, state
  pill, project, Jira key, last update); one KPI strip (Now running, Total cost, Wall clock, Input
  tokens with cached share, Failed runs); a full-width timeline with one lane per stage, bars
  coloured by outcome, an hour axis and a now marker; stage runs grouped by session, each group
  labelled `Session N · <phase>` with its span, duration and cost, every row showing
  `model-version · effort`; a Cost by stage rail.
- **Server** — each state-board row gains the change's open stage (`currentStage`,
  `currentStageStartedAt`), its done-task count (`doneTasks`) and the open session's activity
  (`activity`, `activityDetail`, `activitySince`).
- **Session activity** — a Claude Code hook tells flowd when a session is working, awaiting an
  answer to a question, awaiting a permission, or has ended its turn; the Running now cards show
  it, and a card that needs the operator is bordered amber and sorted first.

## Fix 1 — 2026-10-29, in-run fix after `flow.visual-verify`

The visual verifier compared the board and run view against the approved prototype and reported two departures:

- **Cost in period follows the filters.** The board's tile sums only the filtered rows: `$2.66` under the uitest-alpha chip, `$5.44` unfiltered. The prototype sums every change whatever the filter (`changes.reduce((a, c) => a + c.cost, 0)`). No design decision covers it, so the prototype wins and the tile sums every row.
- **The timeline tooltip lingers one paint.** In the timeline-tooltip strip, frame 003 (424ms) still shows the tooltip after the pointer has left the bar. The prototype hides it in the same handler. The tooltip now hides in the same frame the pointer leaves.
