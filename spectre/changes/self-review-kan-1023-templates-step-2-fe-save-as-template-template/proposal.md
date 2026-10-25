# self-review-kan-1023-templates-step-2-fe-save-as-template-template

## Why

gymie kan-1023's self-review pass
(`/Users/tweety53/Projects/gymie-worktrees/kan-1023-templates-step-2-fe-save-as-template-template/.superpowers/sdd/self-review-pass.md`)
raised four non-`big` findings:

- **F1** — `check-unfinished-work`'s verify-failure spec regex captures a leading `(`, `[`,
  backtick or quote, so the guard demands a citation carrying the stray character and a clean
  citation of the same failure never clears it.
- **F2** — a visual-fix task's `**Regression:**` asks for the departure probe alone; panel rounds
  3, 5, 7 and 9 of the gymie run each raised an Important the round's own fix had introduced.
- **F3** — `close-task.sh` left appended Tasks 13–18 unticked and the parent ticked them by hand.
- **F4** — eight ledger rows render `Tokens: not measured` (dispatch rows 7560, 7562, 7567, 7577,
  7613, 7616, 7621, 7624).

## What changes

- **F1:** the spec character class excludes `(`, `[`, backtick, `'` and `"` (Task 1).
- **F2:** `skills/flow/verify-fix-loop.md` requires a test for every new branch a visual fix adds,
  named in the task's `**Regression:**` beside the probe (Task 2).
- **F3:** no change — disproved: the gymie change's class was `big`, where a gated task's tick
  waits for its reviewer by design (`design.md`, decision `close-task-big-waits-by-design`).
- **F4:** the harvester stamps a launch described `Task <ids>/<n>` onto its task row (Task 3); a
  closed agentless row no launch ran is stamped inline and every closed inline row is measured,
  a zero share written as zero (Task 4); `flowd restamp-dispatch` repairs agentless inline rows,
  and the eight rows are repaired (Tasks 3 and 4).
