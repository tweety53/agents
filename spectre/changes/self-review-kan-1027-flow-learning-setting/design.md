## Context

Self-review fix change for `kan-1027-flow-learning-setting`, planned on that change's own branch
`spectre/kan-1027-flow-learning-setting` in its worktree at `a68fd756`. Five non-`big` findings,
each carried by one task in `tasks.md`; each was traced or reproduced at `a68fd756` before its fix
was designed, and the branch's history since its base `c1282bc4` holds no earlier fix. Every
question was auto-resolved to its recommended option (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). The brainstorm design, with the
per-finding trace, is
`.superpowers/sdd/2026-10-23-self-review-kan-1027-flow-learning-setting-design.md` in this
worktree. Every reworded or added run-loaded sentence is listed in this change's
`verbatim-moves.txt`, checked with `check-verbatim-moves.sh` (exit 0) with all five tasks' diffs
applied.

## Decisions

### The learning note records against the main checkout

**ID:** learning-add-project-dir
**Status:** active
**Chosen:** `flow learning add … -C <project>` in `skills/flow-contracts/learning.md`, with one sentence saying why — the project key derives from the git common directory, identical from the main checkout and the worktree, and the main checkout outlives `/flow-fast`'s cleanup.
**Considered:** recording before `/flow-fast`'s cleanup — splits "recorded as the final message prints" across two moments and a run that stops between them spends a showing unseen; making `fallback.ProjectKey` fall back when the directory is gone — a guess at which project a removed path belonged to.

### The commit chain anchors every git call

**ID:** commit-chain-anchored
**Status:** active
**Chosen:** `git -C <abs-worktree>` on the chain's `reset`, `add` and `commit` in `skills/flow/implement.md`'s COMMIT-PER-TASK paragraph, the form `skills/flow-contracts/git-boundaries-commit-chain.md` already uses, and the "Paths are relative to the worktree root" sentence saying what anchors them.
**Considered:** telling the implementer to `cd` to the worktree root first — a step an implementer already in `stats/` for its tests skips, which is how the defect fired; a dispatch-paragraphs guard row for the chain — the paragraph is pinned by nothing today and the fix is one form a reviewer reads directly.

### The verbatim-moves rule lives in the lint prose

**ID:** verbatim-moves-in-lint-prose
**Status:** active
**Chosen:** one sentence in `.flow/project.md`'s `## lint` prose, beside the paragraph naming `check-verbatim-moves.sh` as the guard for run-loaded prose: a plan whose task edits run-loaded prose names that guard in the task's verify step and commits `verbatim-moves.txt` with the plan. `project-get.sh <project> lint` prints the section's prose, so every planner reading the lint list reads it.
**Considered:** a sentence in `skills/flow/brainstorm-planner.md` — the rule is this repository's, and the planner file serves every project; a plan-shape guard — it would need to know which paths are run-loaded in each project.

### A panel-fix back-fill starts no earlier than its round's reports

**ID:** panel-fix-backfill-floor
**Status:** active
**Chosen:** in `flow record dispatch backfill`'s inline `-commit` derivation (`stats/cmd/flow/record.go`), a `-role panel-fix` row keyed `panel-fix-<r>[-<chunk>]` takes as its start the later of the commit before its fix commit and the latest `endedAt` among round `r-1`'s reviewer rows (`panelFixFloor`); a failed record read, or a round with no ended reviewer row, keeps the commit-derived start, so an outage is never a refusal. `skills/flow/implement.md`'s back-fill sentence states the floor.
**Considered:** the finding's proposed refusal on `dispatch begin -role panel-fix` until a reviewer row of the round has ended — the trace shows the row was written by `dispatch backfill`, not `begin`, so the refusal would not have caught it; a floor at any dispatch's end for every inline role — an inline row may legitimately overlap a background dispatch, and no other role showed the defect.

### The mis-stamped panel-fix-1 row is repaired

**ID:** panel-fix-1-row-repair
**Status:** active
**Chosen:** `flowd retime-dispatch 7260 2026-10-23T09:28:18Z 2026-10-23T09:30:03Z cafe4c30a4be51f9b7438dfc181dc3d762008b82` — the start the floor now derives, the end and commit unchanged — then `flowd restamp-dispatch 7260 <the parent transcript>` to attribute the row's tokens over the corrected window; both run from this worktree's `stats/` against the persistent store `postgres://flow:flow@localhost:5433/flow`.
**Considered:** leaving the row — a forward-only fix leaves the ledger's duration and empty tokens wrong (**3. Fix every finding that is not `big`**, `skills/flow-self-review/SKILL.md`); the first edit's own instant (09:29:31Z) — the floor's value is what the fixed command derives, so the repair matches what a re-run would record.

### Before figures are measured during pass 1's panel

**ID:** live-before-in-panel-window
**Status:** active
**Chosen:** `skills/flow/review-panel.md`'s pass-1 window, where the parent already runs the full suite while the read-only slots review, also measures a live check's before figures after the suite and writes them to `<abs-worktree>/.superpowers/sdd/live-before.md`; **Live check** in `skills/flow/verify-and-handoff.md` copies a figure found there into the record and never measures it again. The measured saving is the before half's ~30s of kan-1027's 7m48s verify window; the 6m59s lint and test run must see the final tree and stays where it is.
**Considered:** the sentence in `verify-and-handoff.md` alone, as the finding counted — that file is not loaded during the panel, so nothing would tell the parent to measure early; writing into `live-verification.md` at panel time — the change's directory is tracked and the suite's tree gate reads it (kan-918).

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 4 changes how a store row's instants derive and repairs a row in the persistent store, and Task 5 moves a step across two phase files.
**Considered:** `flow-low` — three of the five tasks are one-sentence prose, but a misread floor in Task 4 would re-stamp every later panel-fix row.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-1027-flow-learning-setting-fix`, Tasks 1 to 5 in order; Tasks 2 and 4 share `skills/flow/implement.md`.
**Considered:** one fixer per task — five dispatches for five small diffs, two of them on the same file.

## Open questions

## Live check

Task 4 repairs a row of the persistent store, the one runtime this change touches. Read the row
before and after the repair with `flow record dispatches -change kan-1027-flow-learning-setting -C
<this worktree> | jq '.[] | select(.id==7260) | {startedAt, endedAt, metrics}'`. Before:
`startedAt` `2026-10-23T12:19:34+03:00`, `metrics` `{}`. After: `startedAt`
`2026-10-23T12:28:18+03:00`, `endedAt` unchanged `2026-10-23T12:30:03+03:00`, `metrics.tokens`
present. Not working looks like the before figures unchanged, or a `startedAt` other than
12:28:18. Task 4's own steps run it; no separate verification stage runs on a self-review branch.
