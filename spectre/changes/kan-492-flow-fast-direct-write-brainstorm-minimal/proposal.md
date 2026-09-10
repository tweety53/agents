# kan-492-flow-fast-direct-write-brainstorm-minimal

## Why

`/flow-fast`'s own spec still carries most of `/flow`'s ceremony through citation: the full
brainstorming checklist with a specs file, the full `/flow` artifact shape, a two-run
merge-and-push finish with a second landing/archive branch/push, a six-check cleanup, and the
64 KB `review-panel.md` load for three dispatch paragraphs. None of that is read by anything
`/flow-fast` actually runs.

## What changes

- Brainstorm writes `proposal.md`/`design.md`/`tasks.md` directly — no
  `superpowers:brainstorming`, no `superpowers:writing-plans`, no specs file.
- Artifacts shrink to the minimal shape the guards and the task loop actually read (`tasks.md`:
  checkbox + `**Files:**`/`**Tests:**`/`**Commit:**`; `design.md`: prose `## Context`/`##
  Decisions`); no rendered ledger/panel record — the store is the terminal record.
- Finish becomes one run: merge-and-push only, `spectre archive` committed on `<base>`, one push,
  no chore/archive branch, no landing question.
- Cleanup drops checks 1–4, keeping `## stop` and `check-worktree-processes.sh`.
- The review panel's shared dispatch paragraphs move into `skills/flow/panel-dispatch.md`; a
  Critical/Major fix re-runs only `simple-reviewer` on the delta.
- Jira goes straight to Done after the push — no In Review hop.
- Two spec contradictions (`check-plan-shape.sh` vs. the guard set; `/myflow-fast` naming in both
  finish contracts) are fixed in the same change.
