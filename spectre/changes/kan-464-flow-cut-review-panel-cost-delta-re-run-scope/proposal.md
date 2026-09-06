# kan-464-flow-cut-review-panel-cost-delta-re-run-scope

**Jira:** KAN-464

## Why

Across the seven gymie changes measured in
`docs/superpowers/research/flow-gymie-implementation-speedup.md` (thread 2), the review panel ran
16–83 minutes per change, median 32, about 1000 panel-minutes in total. Every one of the seven ran
a fix round, and the minutes concentrate in three places that produce almost no findings.
<!-- measured: transcript analysis of the seven gymie panel runs, recorded in docs/superpowers/research/flow-gymie-implementation-speedup.md thread 2 — not re-runnable as a command -->

**The full-slot re-run after a fix round.** `skills/flow/review-panel.md`'s **Panel re-runs** says
every diff-reading slot re-runs on its delta whenever the round raised anything above Minor —
which every round did. Across all seven changes that re-run produced four stale `tasks.md` fields
and three reviewer-labelled nits, for 1–6 minutes of slots plus the conductor's own dispatch,
<!-- measured: transcript analysis of the seven gymie panel runs, recorded in docs/superpowers/research/flow-gymie-implementation-speedup.md thread 2 — not re-runnable as a command -->
wait and record turns each time. Every high-value finding — kan-422's Critical and three Majors,
kan-454's Critical and High, kan-424's High, kan-455's Major — came from pass 1.

**The conductor running the mutation-proof itself.** **The fix round mutation-proves what it
changed** reads "*you* then mutate each one", so the mutation work sits on the conductor's 200k+
token context by design. In kan-455 that was seven minutes of copying files, running
`:shared:desktopTest` four times and repairing an import in a test the conductor had edited.

**Reproducers reviewers cannot submit.** 11 of 33 findings carried a runnable reproducer. The rest
recorded `none — …` because `check-panel-reproducers.sh` refuses shell metacharacters and a
reviewer that wants a `grep` pipeline has nowhere to put it. In three changes the conductor wrote
the reproducer script itself afterwards — in kan-455 after a ten-minute operator wait on the
refusal — and the guard accepted every path it used. A relative, worktree-contained script path
is already a legal reproducer; nothing tells the slots so.

## What changes

- **A slot re-runs after a fix round only when it raised a finding in the previous round, or the
  previous round raised a new Critical.** This is the clause `skills/flow/review-panel.md`
  already applies to Bugbot, Mutation and Security, widened to every slot; the "every diff-reading
  slot re-runs on its delta" bullet is deleted. The stale-result carve-out that keeps Bugbot's and
  Mutation's results current after a fix they did not raise widens with it, so the handoff's
  zero-stale requirement does not block on the slots the new rule declines to re-run.
- **The fix subagent performs the round's mutation-proof and reports it.** It mutates each
  executable behaviour its fix changed, repairs a surviving mutant before its turn ends, and
  writes the `fix-mutation:` lines into its own report. The parent runs no build: it keeps the
  existing fix-diff hunk walk as its check, and that walk additionally holds the fix subagent to
  the PLAN FIELDS obligation it already carries — the plan-field net that `primary`'s re-run used
  to provide.
- **A new MUTATION PROOF dispatch paragraph** carries that instruction to the fix subagent, with
  its row in `scripts/check-dispatch-paragraphs.sh` and its cases in
  `scripts/test-check-dispatch-paragraphs.sh`, so it cannot be trimmed away silently the way
  KAN-289's instruction was.
- **A reviewer slot may write its reproducer as a script.** Every slot's dispatch prompt states
  that a demonstrating command needing a pipe, a quote, a glob or any other shell metacharacter is
  written to `<abs-worktree>/.superpowers/sdd/reproducers/<round>-<id>-<n>.sh` and recorded as that
  worktree-relative path. No guard changes: `check-panel-reproducers.sh` and `run-reproducer.sh`
  already accept the shape.

Files touched: `skills/flow/review-panel.md`, `scripts/check-dispatch-paragraphs.sh`,
`scripts/test-check-dispatch-paragraphs.sh`.
