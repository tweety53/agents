# kan-436-flow-narrow-the-primary-review-slot-to-plan

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

**Spec:** `design.md` beside this file. One task, `Build: green`, documentation only: two edits in
`skills/flow/review-panel.md` **The roster**. `skills/flow/SKILL.md` (**Model resolution**) and
**The docs-only reduction** need no edit — both still resolve to `primary`, and only the brief the
id carries changes.

- [x] 1. Narrow the `primary` roster row to plan alignment

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** none — documentation; the shipped guards in step 3 are the check
**Regression:** none — no test is added; `scripts/check-references.sh` and `scripts/check-vocabulary.sh` are the mechanical checks and pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ prose; the count is the number of tests this task touches @ branch main -->
<!-- predicted: unchanged after this task, on branch spectre/kan-436-flow-narrow-the-primary-review-slot-to-plan -->
**Commit:** `docs(review-panel): narrow the primary slot to plan alignment`

  - [x] **Step 1: The row.** In `skills/flow/review-panel.md`'s roster table, replace the `primary`
    row (the first data row, reading "**Primary** — plan alignment + code quality |
    **superpowers:requesting-code-review** with `final-review.diff` + the plan/spec constraints")
    with:

```markdown verified:column layout copied from the roster table at skills/flow/review-panel.md:104-111 on main
| `primary` | **Primary** — plan alignment | general-purpose reviewer briefed on `final-review.diff` against `proposal.md`, `design.md` and each task's `**Files:**`/`**Tests:**`/`**Commit:**` fields in `tasks.md` — nothing else; code quality is `code-review-low`'s and Bugbot's | `DEFAULT_MODEL` |
```

  - [x] **Step 2: The subagent-facing-file paragraph.** Directly below the table, the sentence
    opening "Superpowers' `code-reviewer.md` (Primary), `principles-reviewer-prompt.md` and
    `engineering-principles.md` (Principles), …" loses its first item, becoming "Superpowers'
    `principles-reviewer-prompt.md` and `engineering-principles.md` (Principles), and
    `<project>/.flow/project.md`'s standards files are inputs to the slot that reads them; …" —
    the rest of the paragraph is unchanged. Afterwards
    `grep -n "requesting-code-review\|code-reviewer.md" skills/flow/review-panel.md` prints nothing.
    `skills/flow/implement.md:13` and `skills/README.md:36` still cite the skill for the per-task
    review — leave them; that is a separate research item.
  - [x] **Step 3: Verify.** From the repository root run `scripts/check-vocabulary.sh`,
    `scripts/check-references.sh`, `scripts/check-normative-inventory.sh` and
    `scripts/check-contract-budget.sh` — each exits 0. Commit.
