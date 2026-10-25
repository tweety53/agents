# self-review-kan-1032-prompt-tuning-angle

## Why

kan-1032's self-review pass (`.superpowers/sdd/self-review-pass.md` in this worktree) raised four
non-`big` pipeline findings, the coordinator added a fifth, and the planner hit a sixth and a
seventh while planning:

- **F-A1** — `run-verify-list` keeps running a whole verify list after the store refused its own
  `begin`, so a refusal the tool can know before any command costs a full re-run (kan-1032's
  integrate: two groups, a 5m29s verifier window, run twice).
- **F-A2** — the mutation-testing brief says nothing about how a reproducer selects its tests, and
  a reproducer filtering by today's test names never sees the killing test a fix adds under a new
  name.
- **F-A3** — every in-run loop dispatch is recorded twice: once by `check-loop-dispatch.sh` under
  the loop's change, stamped, and once by the parent under the run's change, agentless and never
  measurable. The source: `hooks/enforce-agent-baseline.py` denied the first two
  `pipeline-fix-30-plan` launches — whose prompts lacked the `Key:` line — with "no open ledger
  row … record it first with `flow record dispatch begin -change kan-1032-prompt-tuning-angle -key
  pipeline-fix-30-plan`", and the parent kept hand-recording every later loop step.
- **F-A4** — `integrate-change commit` checks only the format of `In-run pipeline fix:` lines that
  exist; a landed in-run fix narrated without its line passes.
- **F-A5** — `check-prompt-tuning-cases` refuses `source.round: 0`, the initial panel's legal
  round, so this pass's own first case can never pass.
- **F-A7** — `hooks/protect-main-checkout.py` denies a `chore` commit as subjectless when an
  earlier part of the same Bash call holds an unbalanced quote (an apostrophe in a heredoc body):
  its raw-word fallback keeps the quote on the subject word. The planner's own plan commit hit it.
- **F-A6** — `hooks/enforce-agent-baseline.py` denies the planner's own `Explore` and `Plan`
  dispatches in a flow session, because their description opens with no row key — the dispatches
  `skills/flow-self-review/SKILL.md`'s PLANNING DELEGATION and `skills/flow/brainstorm-planner.md`
  section B call for.

## What changes

- `run-verify-list` stops before its first command, exit 4, when the store refuses its `begin`.
- The mutation-testing brief requires a reproducer's test command to run every test of the mutated
  code's package, file or module.
- The baseline hook answers a loop-key description without its `Key:` line by asking for the line,
  never for a row under the run's change; `flow record dispatch begin` refuses a `pipeline-fix-`
  key under a change that is not a loop change; the closed list and **Pipeline defects found
  mid-run** say the hook records loop rows; migration 0047 deletes the agentless twins already
  stored (5 rows, all kan-1032's).
- `integrate-change commit` stops `narrative` on a `flow-fix` `fixed` store row whose sha no
  `In-run pipeline fix:` line names.
- `check-prompt-tuning-cases` accepts round 0 and still refuses an absent or negative round.
- The baseline hook lets `Explore` and `Plan` dispatches through its row checks.
- The commit hook judges no message its raw-word fallback split.
