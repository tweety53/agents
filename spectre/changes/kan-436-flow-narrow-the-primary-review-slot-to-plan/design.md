# Design — kan-436-flow-narrow-the-primary-review-slot-to-plan

## Context

`skills/flow/review-panel.md`'s roster spawns `primary` as "**superpowers:requesting-code-review**
with `final-review.diff` + the plan/spec constraints" and names it "plan alignment + code quality";
the paragraph below the table lists superpowers' `code-reviewer.md` as the subagent-facing file the
dispatcher resolves for it. Section 8 of `docs/superpowers/research/flow-speedup.md` decided the slot
narrows to plan alignment alone: the skill is a dispatcher's template, the slot's findings were
never code defects, and its code-quality half duplicates `code-review-low` on the same diff.

## The row

| id | Slot | How to spawn | Model |
|---|------|---------------|-------|
| `primary` | **Primary** — plan alignment | general-purpose reviewer briefed on `final-review.diff` against `proposal.md`, `design.md` and each task's `**Files:**`/`**Tests:**`/`**Commit:**` fields in `tasks.md` — nothing else; code quality is `code-review-low`'s and Bugbot's | `DEFAULT_MODEL` |

The slot keeps reading `final-review.diff`: since KAN-402 that file is the whole-branch combined
diff every reading slot shares, and the fix-round rule under **The fix round** ("`primary` re-runs
on its delta") depends on the slot being one that reads it. The subagent-facing-file paragraph
becomes "Superpowers' `principles-reviewer-prompt.md` and `engineering-principles.md` (Principles),
and `<project>/.flow/project.md`'s standards files …" — Primary no longer has a file of its own.

## What does not change

**The docs-only reduction** and `skills/flow/SKILL.md`'s **Model resolution** both still resolve to
`primary`; only the brief the id carries changes. `ValidReviewers` is untouched.
`scripts/check-vocabulary.sh` excludes the `requesting-code-review` token structurally, so removing
the mention breaks no guard. The per-task review's own citation of the skill
(`skills/flow/implement.md`, `skills/README.md`) is a separate research item.

## Decisions

### A docs-only branch gets plan-alignment review only

**ID:** docs-only-gets-plan-alignment-only
**Status:** active
**Chosen:** one brief for `primary` everywhere — on a docs-only branch pass 1 is `primary` alone,
so no slot reviews prose or formatting there; the implementer's self-review and the vocabulary and
reference guards are what covers it.
**Considered:** a second clause under **The docs-only reduction** briefing the reduced-roster
`primary` additionally for prose/formatting defects — keeps the old coverage on docs branches, at
the cost of one id carrying two briefs, the split the narrowing exists to remove.

## Open questions

None.
