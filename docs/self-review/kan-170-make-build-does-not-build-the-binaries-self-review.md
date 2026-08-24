# Self-review — kan-170-make-build-does-not-build-the-binaries

**Rating: 4/5** — the panel caught a Critical and a Major that the implementers and the first round
of per-task reviews missed, and every behaviour the fix round changed was mutation-proved. Against
that: two of the four panel findings traced to contradictions inside the plan this run itself wrote,
and a full panel round was spent verifying commit messages that finish run 1 then collapsed away.

Run shape: 4 tasks, 23 dispatches, 3 panel rounds, 4 findings (1 Critical, 1 Major, 2 Minor), all
fixed. Roster `light`; all four conditional slots fired their triggers and the operator declined
them.

## Problems encountered, and what pipeline change would avoid them — `myflow-fix`

- **[myflow-fix]** `tasks.md` collapsed `design.md`'s check → bind → write into a single `Acquire` that checked and wrote before the bind; the contradiction survived the writing-plans stage, both plan guards, an implementer and a per-task review, and was caught only by the second review after task 3 had been built on the wrong API — declined
- **[myflow-fix]** one task specified `/tmp` in its prose and `os.TempDir()` in its own code block; on macOS those differ, and the consequence would have been `ui-test-down`'s unforced `dropdb` racing a live daemon — declined
- **[myflow-fix]** `check-task-commit-fields.sh` reported `Regression: skipped, not verified` and `Baseline: skipped, not verified` on every task here, because the project's `## test` command cannot target a single named test — two of its five fields are never checked in this repository — declined
- **[myflow-fix]** finish run 1's `reset --soft` reshape discards the per-task commits the panel just reviewed; F4 cost a branch replay and a targeted re-run to fix two commit bodies collapsed away minutes later — declined
- **[myflow-fix]** cleanup check 5 runs the project's declared `## stop`, which here is `docker compose down` — a teardown of a container shared by the live stack and every worktree, invoked to clean up one worktree; it needed an operator decision to skip — declined
- **[myflow-fix]** `run-reproducer.sh` cannot express a reproducer that *is* the missing regression test: it exits 0 before the fix because the case does not exist yet, which reads as "defect not demonstrated" and bounces the finding — declined

## Token/time cost, and what would reduce it without quality loss — `myflow-cost`

- **[myflow-cost]** panel round 2 existed only to re-verify two commit messages run 1 then discarded — a full round, two slot dispatches, on artifacts with a 20-minute contractual lifespan — declined
- **[myflow-cost]** the escalation ladder cannot distinguish mechanically forced churn from a fix that wandered: `Release` gaining one `*slog.Logger` parameter forced a one-line call-site change, tripping "touched a file outside the findings' set" and escalating targeted → Full, costing three full-diff re-reads instead of two targeted ones — declined
- **[myflow-cost]** `myflow record cost-status` reported 13 of 23 dispatches unattributed, so over half the run's cost cannot be attributed — declined

## What went well, and how to reproduce it — `myflow-improvement`

- **[myflow-improvement]** requiring reviewers to *construct* a failure rather than argue it: task 2's re-reviewer disproved the package doc's safety claim by building the race it denied, turning a convincing paragraph into a Major finding — declined
- **[myflow-improvement]** telling each reviewer to verify the implementer's own RED claim rather than take its word; task 1's reviewer reverted the Makefile hunk in a scratch copy and tested the inverse defect too, establishing the guard fails in both directions — declined
- **[myflow-improvement]** briefing slots to trace a format change into its consumers in other files — how the Critical was found: `pidfile.Write` changed one line to two and the Makefile still `cat`ed it whole — declined

## What could be automated or moved to a script — `myflow-automation`

- **[myflow-automation]** a guard diffing `tasks.md`'s code blocks against `design.md`'s for contradictory API and ordering claims; both plan defects above are mechanically detectable — declined
- **[myflow-automation]** the per-dispatch `begin`/`end` bookkeeping is hand-typed, and a mistyped merge base produced a guard exit 2 indistinguishable at a glance from a real refusal — filed: KAN-330

## What could move to the Go app or its persistent storage — `myflow-stats-app`

- **[myflow-stats-app]** the review panel's pass log is the last hand-written panel artifact and is destroyed with the worktree, taking with it the only record that the round's seven mutations ran — filed: KAN-331
- **[myflow-stats-app]** record `-agent-id` at dispatch `begin` for every role rather than only for panel slots, which is what left 13 of 23 dispatches unattributable — filed: KAN-332
