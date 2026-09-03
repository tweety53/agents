## Context

`skills/flow/review-panel.md` dispatches every resolved slot on pass 1, whatever the branch's
composition; only re-runs after a fix are delta-scoped (KAN-277). The roster itself comes from the
settings store per `roster-from-settings` (archived `panel-roster-follows-the-settings-list`),
which forbids adding a slot by diff size, touched area or any other automatic trigger, and says
nothing about removing one. `skills/flow/SKILL.md`'s roster resolution already defines a
`primary`-alone shape for an empty store list. `check-panel-citation-trigger.sh` already collects a
branch's own paths (committed since the merge base, staged, unstaged) and tests them against
`\.mdc?$` — the one definition of "documentation path" this repository holds.

The per-task review (`skills/flow/implement.md`) is one combined spec-and-quality reviewer per
task, reading `git diff <task-base>..<task-sha>`. On a branch whose every path is Markdown, the
panel's whole-branch read covers the same text those reviewers read; there is no code seam between
commits for it to find. This is one change because it touches one site — how the panel's pass 1
roster is chosen — and the three contract texts that describe that choice.

No staged research note existed (`docs/superpowers/research/kan-312.md`, `kan-312-*.md` absent).
The approved design is `docs/superpowers/specs/2026-09-03-kan-312-myflow-per-task-review-and-the-panel-duplicate-design.md`;
its sections 1–4 are the plan `tasks.md` carries.

## Decisions

### The signal is diff composition: every touched path is `.md` or `.mdc`

**ID:** docs-only-signal
**Status:** active
**Chosen:** a branch is "docs-only" when every path it touched — committed since the merge base,
staged, or unstaged, the same union `check-panel-citation-trigger.sh` collects — ends `.md` or
`.mdc`. Deterministic, no threshold to tune, spares KAN-302 (all Markdown) and keeps KAN-287 (Go)
on the full roster.
**Considered:** docs-only plus a line-count floor — rejected as a second threshold to justify and
record for no case yet observed; a single task commit only — rejected because it would not have
spared KAN-302's four tasks; more than one task touching a shared file — rejected because four
tasks over two files share files, so it too would not have spared KAN-302.

### A docs-only branch reduces the panel to `primary` alone

**ID:** docs-only-reduces-to-primary
**Status:** active
**Chosen:** on a docs-only verdict pass 1 dispatches `primary` alone — one whole-branch
plan-alignment read — plus any slot the operator's existing per-run instruction names. The panel
record still renders, `check-panel-findings-closed.sh` still gates, the zero-open-findings bar is
untouched. `primary` is the reduced roster even when the resolved list does not carry it, the same
shape `skills/flow/SKILL.md`'s empty-list row already defines. This narrows `roster-from-settings`
rather than superseding it: that decision forbids an automatic addition; this is the one automatic
reduction, and it only ever removes.
**Considered:** skipping the panel entirely — rejected, it leaves no whole-branch read and no
record shape that distinguishes "not run — docs-only" from a silently skipped review; asking the
operator each time — rejected, one more prompt on every docs-only run, which in this repository is
most runs, when the per-run add-slot instruction already covers the override.

### An unanswerable guard runs the full roster

**ID:** guard-exit-2-is-full-roster
**Status:** active
**Chosen:** `check-panel-docs-only.sh` exit 2 reports its stderr and dispatches the resolved roster
unchanged; an empty touched-path set exits 1 and likewise runs the full roster.
**Considered:** treating exit 2 as docs-only — rejected, an unanswered question must never shrink a
review; stopping the run — rejected, the full roster is always a safe answer.

### Fix rounds re-classify

**ID:** fix-rounds-reclassify
**Status:** active
**Chosen:** the guard runs again wherever the re-run cap check runs. A branch that stops being
docs-only dispatches every resolved slot not yet dispatched this run, reading the whole
`final-review.diff` under the existing no-held-sha rule; a branch that stays docs-only keeps the
reduced roster and `primary` re-runs on its delta.
**Considered:** classifying once at pass 1 — rejected, a fix that adds a script would leave code on
the branch that only `primary` ever read.

## Open questions
