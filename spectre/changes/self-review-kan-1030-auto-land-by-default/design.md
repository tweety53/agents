## Context

Self-review fix change for `kan-1030-auto-land-by-default`, planned on the change's own branch
`spectre/kan-1030-auto-land-by-default` at `c0075269` (its archive commit), in its own worktree;
the fixes ride that change's landing route. F1, F2 and F4 as `proposal.md` names them are carried by
Tasks 1, 2 and 3; F3 by its own planning commit. Each finding was traced before its fix was designed, and none was disproved:

- F1: `sync-panel-base` calls `rebaseOntoTip` with `aside` false
  (`stats/internal/guard/syncpanelbase.go:62`), a bare `git rebase`; git 2.50.1 refuses it over an
  uncommitted tracked edit ("cannot rebase: You have unstaged changes", spiked in a scratch repo
  with a dirty `spectre/changes/c/tasks.md`); `close-task.sh` ticks `tasks.md` and pushes without
  committing it. kan-1030's own creating run hit exactly this (its narrative's third bullet).
- F2: `verify-and-handoff.md`'s "`true` prints the block below … Anything else prints no block"
  drops the block's `Decisions:` lines; integrate's handoffs (`handoff-blocks.md`) have no such field.
- F3: `scripts/close-task.sh` is present at `ebfb5504` and `6c9c7d53`; `scripts/stage-task-files.sh`
  is absent at `ebfb5504`. `live-verification.md` records `flowd` built at the merge base and the
  branch, each on port 4199 over a scratch database.
- F4: **Before this stage ends, start the stack** and **A refused start is relayed, never resolved**
  carry no exception for a run that lands straight away.

Every fix is prose; each task's exact text is its `patches/task-<n>.diff`, applied in task order
(`git apply --check` passed for each on `c0075269` with its predecessors and the F3 correction applied), and every added
or reworded run-loaded sentence is listed in `verbatim-moves.txt`. Every question was auto-resolved
to its recommended option (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). The brainstorm design is
`.superpowers/sdd/2026-10-24-self-review-kan-1030-auto-land-by-default-design.md` in this worktree.

## Decisions

### The before-review planning commit opens the base sync

**ID:** planning-commit-before-base-sync
**Status:** active
**Chosen:** a paragraph opening **Check base movement first** makes the before-review planning commit before `sync-panel-base.sh` runs in any worktree; one sentence in **Panel re-runs** keeps it first at a round boundary; `git-boundaries.md`'s "Before every reviewer dispatch" row adds the base sync. The later reviewer-dispatch planning commit stays — its empty delta skips itself.
**Considered:** passing `aside` true in `syncpanelbase.go`, as `sync-onto-base` does — the ticks are committed at the next boundary anyway, and setting them aside restores uncommitted planning edits over a rebased tree for no gain; moving the later planning commit up instead of adding one — that commit also brackets the round's plan-tree snapshot, which must follow the sync.

### An open question stops the run; auto-resolutions print before integrate

**ID:** open-question-stops-auto-land
**Status:** active
**Chosen:** a run leaving a `?` line prints the `IN_PROGRESS` block and ends whatever `reviewBeforeLanding` holds; one with 🤖 lines and no `?` prints them under a `**Decisions:**` heading, then continues into integrate. `pipeline.md`'s **States** sentence and its three transition-table cells, and `skills/flow/SKILL.md`'s fix-run line, add the open-question stop.
**Considered:** adding a `Decisions:` field to integrate's handoffs (`handoff-blocks.md`) — a third block shape and a `/flow-status` regeneration for lines the continuing run can print itself; leaving the always-on trigger summaries (`rules/flow-manual-review.mdc`, `AGENTS.md`, `templates/AGENTS.md`, `skills/README.md`) and the `integrate.md`/`state-file.md` lines as they are — chosen: they state the flag's default effect and point at `pipeline.md`, and none says "only".

### Correct the archived narrative in a planning commit

**ID:** narrative-corrected-in-planning-commit
**Status:** active
**Chosen:** the planner corrects the two lines of `spectre/changes/archive/kan-1030-auto-land-by-default/narrative.md` — the script line names `stage-task-files.sh` alone, the live-check line points at `live-verification.md` — in its own commit, `chore(spectre): correct kan-1030's narrative on its live check and the missing script`, made before the plan commit; the review reads it in `git diff c0075269...spectre/kan-1030-auto-land-by-default`. No task carries F3.
**Considered:** a fixer task — `check-plan-shape.sh` refuses a task whose every `**Files:**` path lies under `spectre/changes/` ("make the record a planning commit with no task"), and the fixer never writes `spectre/changes/`; leaving the lines with a note here (kan-1012's `k1012-live-verification-line-left`) — the record stays wrong for every later reader.

### A run that lands straight away starts no stack

**ID:** no-stack-start-on-auto-land
**Status:** active
**Chosen:** a paragraph under **Before this stage ends, start the stack**: when `reviewBeforeLanding` is not `true` and the run leaves no `?` line, the start, the freshness check and the refused-start stop are skipped. The run-instruction resolution itself still runs; it writes nothing.
**Considered:** skipping only the refused-start stop — still rebuilds a stack nobody reviews; skipping on `reviewBeforeLanding` alone — a run that then stops on an open question would hand off with no stack, against **The `Running:` section is never omitted**.

### Prose fixes rest on the existing guards

**ID:** prose-guards-cover
**Status:** active
**Chosen:** no new test: every fix is contract prose. `check-verbatim-moves.sh` fails each added or reworded sentence unless this change's `verbatim-moves.txt` lists it, and `check-references.sh`, `check-markdown-integrity.sh` and `check-installed-citations.sh` check every cited section and path.
**Considered:** a Go test pinning sentence order in `review-panel.md` — a bespoke guard for one paragraph's position, which the next reflow breaks.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — three short prose patches and a two-line record correction whose exact text the plan fixes and the verbatim guard checks; no code changes.
**Considered:** `flow-medium` — F2 and F4 change when a run stops, but the review judges two short paragraphs against decisions already recorded here.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-1030-auto-land-by-default-fix`, Tasks 1 to 3 in order: three small patches built in that order, Tasks 2 and 3 sharing `verify-and-handoff.md`.
**Considered:** one fixer per task — three dispatches for three `git apply` runs.

## Open questions

## Live check

None — every change is `/flow` contract prose and one archived narrative; no service or store changes.
