# self-review-kan-920-flow-improvement-port-the-markdown-and-lint

## Why

kan-920's self-review pass found six pipeline defects, none `big`; review 1 disproved A1:

- **A1 (flow-fix) — disproved, dropped.** The parent fixed panel round 0's four Minors inline and
  recorded no `panel-fix` dispatch row. **Disproved by review 1:** `skills/flow/review-panel.md` (lines 802-809 at this branch) has the parent fix a round that raised no Critical and no Important inline, with "no dispatch record"; kan-920's round 0 raised four Minors and nothing else, so the missing `panel-fix` row is the contract working, not a defect. The guard commit was dropped and the back-filled `panel-fix-0` row deleted from the dev store.
- **A2 (flow-fix).** Task 5's inline `dispatch begin` ran after the same Bash call had already
  edited `design.md`, so its row recorded 7 seconds for the task. Only the 3-second floor of
  `flow record dispatch end` checks inline timing.
  <!-- measured: flow record dispatches on the dev store, read 2026-10-21 → 22:48:47.687 to 22:48:54.752 +03:00; the floor from implement.md's Records bullet @ branch spectre/kan-920-flow-improvement-port-the-markdown-and-lint -->
- **A3 (flow-fix).** `brainstorm-planner.md` tells Decide to record the parent's effort as
  `unknown` where the harness states none; `model-policy.md`'s **Harness mapping** says `default`,
  and `flow record dispatch begin -effort unknown` is refused, so task 1's begin failed once.
- **A4 (flow-fix).** A subagent transcript writes every message's usage at stream start, each line
  carrying `"stop_reason": null`, and never the final count. The harvester stores that first-chunk
  output as the dispatch's output, so every subagent row under-reports output and every cost view
  under-prices it.
- **S1 (flow-speed).** `run-verify-list.sh` runs the `## test` list's commands one after the
  other: the guard harnesses, then `go test -race`, about six and a half minutes, twice a run.
- **S2 (flow-speed).** The read-only review panel is dispatched only after that full suite has
  finished.

## What changes

- `flow record dispatch begin -agent-id inline` refuses a worktree already carrying tracked
  modifications other than the change's own `tasks.md`; this change's task-5 row gets its real
  start.
- Decide records an unstated parent effort as `default`, `flow decision render` refuses any other
  word outside the `-effort` set, and a migration rewrites the stored `unknown`s.
- The harvester counts a stream-start message under `output_unmeasured` instead of adding its
  output; pricing refuses such a bucket, the ledger and the SPA render its output as not measured,
  and a migration marks the rows already stored.
- `run-verify-list.sh` runs a `test` list's commands concurrently, reported in list order, and the
  SPA build takes a lock so concurrent callers never race `npm ci`.
- Pass 1 of the review panel is dispatched before the full suite, which runs while the slots read.
