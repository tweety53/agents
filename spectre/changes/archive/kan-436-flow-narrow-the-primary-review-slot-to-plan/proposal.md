# kan-436-flow-narrow-the-primary-review-slot-to-plan

**Jira:** KAN-436

## Why

`primary`'s only unique mandate on the review panel is plan alignment — no other slot is briefed on
the branch against `proposal.md`, `design.md` and `tasks.md`. Its code-quality half duplicates
`code-review-low` on the same `final-review.diff`. Across the seven changes
`docs/superpowers/research/flow-speedup.md` section 8 measured, `primary` produced 3 findings, all
prose or formatting and no code defect, against bugbot's 11, `code-review-low`'s 4 (2 Major) and
principles' 2.

The roster row also hands the slot `superpowers:requesting-code-review` — a dispatcher's skill whose
`SKILL.md` tells a *caller* how to brief a reviewer and whose `code-reviewer.md` is a template with
`[DESCRIPTION]/[PLAN]/[BASE_SHA]/[HEAD_SHA]` placeholders. A slot given it is asked to self-brief
with a template meant for whoever briefs it; five of the seven `primary` runs never invoked it and
ran on the conductor's own brief.

## What changes

- `skills/flow/review-panel.md` **The roster** — the `primary` row drops the
  `superpowers:requesting-code-review` reference. The slot is a general-purpose reviewer briefed on
  `final-review.diff` against `proposal.md`, `design.md` and each task's `**Files:**`, `**Tests:**`
  and `**Commit:**` fields in `tasks.md`, and nothing else; code quality is `code-review-low`'s and
  Bugbot's.
- The same file's "subagent-facing file" paragraph drops `code-reviewer.md` (Primary) from the
  list of files the dispatcher resolves by path.
- Unchanged: **The docs-only reduction** still reduces pass 1 to `primary` alone; **Model
  resolution** (`skills/flow/SKILL.md`) still falls back to `primary` on an empty store list; the
  fix-round re-run rule; `ValidReviewers`. `skills/flow/implement.md` and `skills/README.md` keep
  citing `superpowers:requesting-code-review` for the per-task review — that is the research
  note's separate "Per-task review deleted" item, not this change.
