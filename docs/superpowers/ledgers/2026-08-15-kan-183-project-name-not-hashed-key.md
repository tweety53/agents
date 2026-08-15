# SDD ledger — kan-183-project-name-not-hashed-key

Command: `/myflow-fast` (creating run). Harness: claude-code. Session token: `mf-kan183-a1`.
Recorded models (myflow-fast defaults): implementation `sonnet`, review panel `sonnet`, panel fixes
`sonnet`. Review panel roster: `light`.

Worktree: `/Users/tweety53/Projects/agents-kan-183` · branch
`openspec/kan-183-project-name-not-hashed-key` · merge base `c6f6dab` · workspace id `kan-183-4ec3`.

## Dispatches

Task 1: complete (commit `9529208`, review clean first time, model: sonnet, review: combined).
`projectLabel` and its nine cases. The reviewer additionally probed a key that is only the suffix,
two suffix-shaped tails, trailing whitespace and a non-ASCII basename, and found no defect.

Task 2: complete (commit `36b83f7`, model: sonnet, review: combined). Original commit `c797183`;
the reviewer found **F1 — the change dropdown's shortened label had no test that actually locked
it**: the existing assertion used a project key whose tail could never match the suffix, so it passed
identically whether or not the shortening worked. The same implementer added a locking test and
proved it non-vacuous by reverting the change in a scratch copy and watching it fail. Rewritten again
by fix round 1's rebase.

Task 3: complete (commit `22f3fd2`, model: sonnet, review: combined). Original `42484a9`, rewritten
by both fix rounds. `check-task-commit-fields.sh` initially failed on three undeclared files —
`server.go` and two test fakes — and the **parent repaired the plan** rather than sending the task
back: growing a store interface forces the interface's own file and every fake implementing it, so
the sweep was declared rather than argued away.

## The review panel

Three passes, two fix rounds, six findings — all fixed, none withdrawn. The record is
`final-review-panel.md`; what belongs here is the shape of what it caught.

**Pass 1** found a documentation-free third copy of the suffix pattern (F1), a method signature
repeated across three interfaces (F2), and — the substantive one — **F3: the Project column's
`accessor` returned the display name, so the client's own filter merged two projects the server
refuses to merge**, answering 400 on that same ambiguity. Fixing it reversed a decision the plan had
made deliberately.

**Pass 2** found the deeper one. Task 3 turned two **pure parsing functions into functions that
perform I/O**, and no call site's error classification was revisited: a database failure during
project resolution surfaced as a **400 carrying raw `store:`-prefixed internal text, unlogged**,
while an identical failure a line later became a logged 500. The fakes even carried an error field to
simulate it that no test ever set. Pass 1 had checked that every caller compiled and that behaviour
was preserved for well-formed input; nobody asked what a *new kind of failure* meant for the existing
"any parse error is a 400" rule. That question is the lesson worth keeping.

**Pass 3** confirmed all six closed, with Primary independently reproducing the pre-fix defect in a
throwaway worktree rather than accepting the fixer's account of it.

## Notes carried to the handoff

- The operator answered the panel-composition question once, on pass 1 — required three slots only —
  and said plainly that re-asking it every pass was unwanted. Passes 2 and 3 ran that roster without
  asking. Security, Adversarial and Lens B all had triggers fire and are recorded as declined by the
  operator, not as untriggered; Security's and Adversarial's grounds were folded into Primary's
  dispatch instead, which is a substitution and not equivalent coverage.
- Task 2's plan predicted 150 → 156 SPA tests; the real figure is 157 after both fix rounds. The
  predictions were left as written, tagged `predicted:`, rather than rewritten after the fact.
