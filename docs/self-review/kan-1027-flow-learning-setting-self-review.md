# Self-review report for kan-1027-flow-learning-setting

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `/flow-fast` recorded the learning note with `-C <worktree>` after cleanup removed the worktree, so project-key resolution failed and the three-times cap never applied; now `-C <project>` — fixed: 9c74f7f7
- **[flow-fix]** The guarded commit-per-task chain was cwd-dependent: run from `stats/`, `git add -- <repo-relative paths>` failed with a pathspec error; every run-instructed `git add`/`git commit` now runs as `git -C <abs-worktree>` — fixed: 82b68c9e
- **[flow-fix]** Plans for run-loaded prose omitted `check-verbatim-moves.sh` and `verbatim-moves.txt` (panel F1/F8); `.flow/project.md`'s `## lint` now tells the planner — fixed: c00a9cd7
- **[flow-fix]** The narrative's claim that `check-references.sh` rejects bold tokens starting with a digit is disproved: over 30 numbered bold citations pass in `skills/` — noted
- **[flow-fix]** The plan's `<details>` fold broke the no-HTML handoff rule; the panel caught it and the change fixed it — noted
- **[flow-fix]** The archived `design.md` Live check both names and forbids the dev flowd; the verifier followed the safe half — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Reviewer rows read `output not measured`, the renderer's documented `OutputUnmeasured` harvester limit — noted
- **[flow-cost]** Inline prose task 4 read 3.5M cache tokens, the parent's accumulated context as inline mode expects — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The live check ran against the change's own workspace store with before figures from a merge-base worktree, never the dev flowd — noted
- **[flow-improvement]** Widened `**Files:**` lists were recorded as `Correction:` paragraphs, and F9 was withdrawn with the plan's recorded reason — noted
- **[flow-improvement]** The post-fix re-review ran on sonnet at low effort in 24s; inline tasks took about 2 minutes each — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** An inline `panel-fix-<r>` back-fill took its start from the pre-review planning commit, folding the panel wait into its duration; `flow record dispatch backfill` now floors it at its round's last reviewer end, and row 7260 was retimed and restamped — fixed: dde3fd17

## What can be sped up — `flow-speed`

- **[flow-speed]** Live-check before figures were measured serially after the panel; pass 1's panel now captures them with their merge-base sha, re-measured when a rebase moved it — fixed: 721b8b6a
- **[flow-speed]** The primary reviewer's 8m17s was the panel's long pole, within normal range — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case: Task 1, store the learning setting and topic history (mining skipped this run) — noted
  - Task: `spectre/changes/archive/kan-1027-flow-learning-setting/tasks.md:42-79`
  - Code: `stats/internal/store/learningtopics.go:49-80` at 64ec4d2b — advisory-locked cap check in one transaction
  - Why: a store-enforced cap under concurrency with a clear slug-validation edge set
- **[flow-eval]** review case: F8 Critical, principles slot, `check-verbatim-moves.sh` failing on the change's prose (mining skipped this run) — noted
  - Location: `skills/flow-contracts/learning.md:1`; fixed by `verbatim-moves.txt` in d8854416
  - Why: the reviewer has to run a guard the plan left out instead of trusting the plan's verify list
