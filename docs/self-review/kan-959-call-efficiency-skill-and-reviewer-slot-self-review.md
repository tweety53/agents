# Self-review report for kan-959-call-efficiency-skill-and-reviewer-slot

**Deferred:** reasoning pass run on claude-opus-5-5 from the in-run bundle

In-run pass over the bundle (ledger, panel record, tasks, design, narrative, commit log) and nothing beyond it. Two pipeline defects fixed on the change branch; no product defect met; rating 4 — good.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `render-slot-prompt.sh` validated slots against a hardcoded Go map, so main's installed build could not render new slots; now any slot with a `<id>-reviewer-prompt.md` template is accepted — fixed: 668f9e8b
- **[flow-fix]** `check-verbatim-moves.sh` FAIL lines printed `#`-leading sentences raw, which the ack file reads as comments; FAIL lines now print the escaped ack line — fixed: 7d32f7a6
- **[flow-fix]** inline panel-fix row's begin recorded after the work, so its end was refused and the row closed aborted; the begin-first rule exists and was not followed — noted
- **[flow-fix]** plan slips (Task 2 missing `-session-model`, Task 3/5 subtest ownership) and task-order forward citations, all corrected in-run — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** inline tasks re-read the growing parent context (0.36M–3.59M cache read per task; reviewer 3.96M); no quality-neutral cut evident — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** panel caught stale Decide step citations from renumbering, sonnet/low re-review clean; new extra dispatch cheap (124k cache read) — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** a new skill needs expected-zero declarations in three coverage guards, two found only at full lint; a shared declaration waits for a second instance — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

_none — this angle produced no findings._
