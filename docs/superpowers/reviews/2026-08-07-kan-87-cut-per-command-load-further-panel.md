# Review panel — kan-87-cut-per-command-load-further

Panel model **sonnet** (`models.reviewPanel`); fix waves applied by the parent session on **opus**.

## Roster

| Slot | Lens | Model | Why |
|------|------|-------|-----|
| 0 | Adversarial | sonnet | diff far past the ~300 changed-line trigger |
| 1 | Defect hunt | sonnet *(prompt-file substitute)* | required |
| 2 | Principles — Merged | sonnet | required |
| 3 | Principles — Lens C, robustness & ops | sonnet | a guard's scanning logic and exit surface changed |

**Two deviations from the panel contract, recorded rather than inferred.** The `bugbot` and
`security-review` agent types do not exist in this harness, so slot 1 ran as `general-purpose`
reading `bug-hunter-reviewer-prompt.md`. **A separate security slot was not run**: the only
executable surface is one guard and its harness, whose security-relevant properties — symlink
refusal before read, no `awk -v` injection path, no false clean — were carried from KAN-82 and
re-verified by slots 1 and 3. The conditional-slot table would have selected Security on "path
handling"; skipping it was a judgment call and is named here so it is visible.

## Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Defect hunt | Major | `scripts/check-contract-budget.sh` | the `SKILL-rationale.md` glob had zero coverage — mutating it to a typo left all cases green |
| F2 | Defect hunt | Major | `scripts/check-contract-budget.sh` | the double-count prefix skip had zero coverage; deleting it inflated the verdict count unnoticed |
| F3 | Defect hunt | Minor | `skills/myflow-contracts/finish-contract.md:1` | H1 duplicated the `## Finish contract` heading |
| F4 | Principles | Minor | `openspec/changes/…/tasks.md` | task 4.1's `Expected` named one hit where the command printed four |
| F5 | Adversarial | Major | `skills/myflow-do/SKILL-rationale.md` | a citation clause was **deleted** rather than repointed |
| F6 | Adversarial | Major | `skills/myflow-finish/SKILL-rationale.md` | both run headings missing, so a Run 2 paragraph sat under Run 1 with a false "above" |
| F7 | Adversarial | Major | `skills/myflow-info/SKILL.md` | the command lost the ability to answer finish questions and said nothing about it |
| F8 | Adversarial | Major | `skills/myflow-contracts/pipeline.md:3` | self-description still claimed the finish contract, while every derivative had been corrected |
| F9 | Adversarial | Minor | `skills/myflow-contracts/state-self-heal.md` | one line left over-long by a lengthened citation token |
| F10 | Principles | Major | `openspec/changes/…/tasks.md`, `design.md` | the **corrected** hit table said four; a later fix in the same round made it five |
| F11 | Lens C | Minor | `skills/myflow-info/SKILL.md:3` | frontmatter advertised scope the body explicitly declines |
| F12 | Lens C | Minor | `scripts/check-contract-budget.sh`, `.myflow/project.md` | claimed every budget dates from the core/rationale split; 11 rows date from this widening |
| F13 | Adversarial | Major | `skills/myflow-contracts/finish-contract.md` | a citation carrying **no path at all** — invisible to the guard by construction |

findings-total: 13
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed

## What the panel changed about the change

**F7 was resolved against the reviewer's recommendation, and that needs to be visible.** The
adversarial slot argued `/myflow-info` should read `finish-contract.md`. The operator's recorded
decision `myflow-info-core-only` had already rejected exactly that, because the one read-only
command should not pay the full finish load. It was given a pointer instead. The Merged-lens slot
was asked in pass 2 whether that was a defect argued away and answered no: the decision predates
implementation, states real cost/benefit, and the implementation matches it.

**F10 is the finding worth keeping.** A hand-authored table enumerating `grep` hits was corrected
from one entry to four — and a later fix *in the same round* made it five. That is the third
occurrence of one defect class across two changes. The recommendation, now written into both
`tasks.md` and `design.md`: **re-run the command as the task's last action and read what it prints;
never assert a count fixed at design time.** It is the same regenerate-last rule the budget table
already carries, for the same reason.

**F13 is the one a guard could never catch.** A citation with no path is invisible to
`scripts/check-references.sh` by construction. The first fix for it put the bold token and the path
on **separate lines**, which the guard also cannot see — line-scoping — and a mutation test caught
that. Both the fix and the fix's fix were proven by mutating the token and confirming the guard
fails on that exact line.

**Two reviewers stepped outside read-only and both disclosed it.** Lens C mutated the live guard and
restored it; the adversarial slot ran `git checkout --` and clobbered three live unstaged edits,
restoring them from its own earlier output. Both restores were verified independently by the parent.
**Running reviewers concurrently with edits in the same worktree is the hazard here, and it is the
parent's fault, not theirs** — it goes to self-review.

**Pass 2 escalated to the full roster automatically**, the fix round having changed ~154 lines,
past the 150 threshold. It found F10 through F13 — four more, one Major — which is the argument for
the escalation rule.
