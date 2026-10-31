# kan-1062-flow-implement-one-shot-subagents

## Why

KAN-1062: during `flow.sdd-tdd` the main session does work a fresh subagent could do at a fraction of
its context. Over 2026-10-01..31 the board's own `flow.sdd-tdd` main share is 45.8% — $1,049.16 of
$2,289.35 across 184 runs, a mean of $12.44 per run.
<!-- measured: GET http://127.0.0.1:4173/api/v1/stats/stage-leaderboard?from=2026-10-01T00:00:00Z&to=2026-10-31T12:00:00Z (mainSharePct, stats/internal/store/aggregate.go) @ the dev store, read 2026-10-31 by planner plan-1; a live store, not re-runnable at a ref -->

Where it goes, split by each run's decided class. Main cost is apportioned by priced token weight, a
planning estimate rather than the board's own definition:

| command · class · execution | runs | sdd-tdd cost | main share |
|---|---|---|---|
| /flow big · sdd | 79 | $1,763 | 36% |
| /flow regular · inline | 84 | $450 | 80% |
| /flow small · inline | 25 | $70 | 93% |
| /flow micro · inline | 5 | $6 | 93% |

<!-- measured: GET /api/v1/stage-runs?stage=flow.sdd-tdd&limit=500&sort=-started_at filtered to startedAt >= 2026-10-01, joined to GET /api/v1/records/{project}/{change}/decisions @ the dev store, read 2026-10-31 by planner plan-1; a live store, not re-runnable at a ref -->

- **Every class but `big` implements inline.** The parent does every TDD turn on its own context,
  which is 144–157K tokens on entry to the stage and grows past 300K.
  <!-- measured: main transcripts of kan-1060, kan-1061 and gymie kan-1023 in their sdd-tdd windows, parsed by planner plan-1's scratch script, deduplicated by message.id @ machine-local transcripts read 2026-10-31, not re-runnable at a ref -->
- **Under `sdd` the parent still fixes what a child leaves.** A guard refusal, a pick conflict, a
  gated reviewer's `fix`, a clean pass's Minors and every full-suite failure are the parent's own
  inline edits (`skills/flow/implement.md`, `skills/flow/gated-review-fix.md`,
  `skills/flow/sdd-dispatch.md`). That fix and investigation work is 20–35% of a `big` parent's
  cache reads.
  <!-- measured: share of main cache-read by tool call over the same three transcripts @ machine-local transcripts read 2026-10-31, not re-runnable at a ref -->
- **A group boundary takes several parent turns, each at full context** — polls, a separate
  dispatch end, `close-task.sh`, the gather, the begin, snapshots, report reads and the launch.
  `flow record`, `flow stage` and `flow tasks` calls alone are 21–39% of a `big` parent's cache reads.
  <!-- measured: share of main cache-read by tool call over the same three transcripts @ machine-local transcripts read 2026-10-31, not re-runnable at a ref -->
- **Every implementer prompt carries the dispatch paragraphs**, and `implement.md` — 88,669 bytes —
  is read whole on every run and kept for every later stage.
  <!-- measured: git show 5302fcb2:skills/flow/implement.md | wc -c @ 5302fcb2 -->
- **The parent waits by polling** — a `sleep 5` / `test -s` loop (`implement.md`'s **Turn
  discipline**), where the review panel has waited on completion notifications since 983be8de.

## What changes

- **Every class but `micro` runs `sdd`.** `plan-class.sh`'s tree prints `execution sdd` for `small`,
  `regular` and `big`; only `micro` stays `inline`. The gated per-task reviewer stays `big`-only.
- **Under `sdd` the parent edits no code.** A re-commit a guard refusal needs, a wave pick conflict,
  the gated reviewers' `fix` verdicts and open Minors, and every full-suite failure go to a fresh
  one-shot fixer, role `implementer`, on the decision's `fixer` pair — the group's own pair for a
  pick. The parent keeps every judgement: corrections, findings, withdrawals, questions, pivots and
  every record call. `flow record dispatch begin` refuses an `implementer` or `panel-fix` row begun
  `-agent-id inline` on a `task-*` or `full-suite-fix-*` key under an `sdd` decision.
- **The gated reviewers' fixes run once, at the stage's end** — one fixer for every `fix` verdict and
  every open gated Minor, then one re-review bundle. A Minor-only task ticks on its clean verdict.
- **Briefs live on disk; every implementer, fixer and gated reviewer prompt is a stub.** The dispatch
  paragraphs move verbatim to `skills/flow/implementer-brief.md` and
  `skills/flow/gated-reviewer-brief.md`; `implement.md` keeps one literal stub per role.
- **A child's summary is its final message.** An implementer or fixer returns a fixed summary head; a
  gated reviewer bundle returns one line per pass and writes its own findings file, which
  `flow record finding -from` numbers from the change's next free ref.
- **A group boundary is one wake, one Bash call and one launch message.**
- **Waiting is the completion notification.** Every child goes out in the background and the parent
  ends its turn; a background `sleep 3000` keeps its one-hour prompt cache warm while a child runs.
- `hooks/enforce-agent-baseline.py` admits a description opening with a `full-suite-fix-<n>` key.
