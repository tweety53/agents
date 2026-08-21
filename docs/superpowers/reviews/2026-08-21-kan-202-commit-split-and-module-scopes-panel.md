# Final review panel — kan-202-commit-split-and-module-scopes

Roster: **light** — Primary, Principles (Merged lens), Code review (low effort). Panel model
`sonnet` for every slot. Five passes; passes 2, 3, 4 and 5 were **Full** re-runs, each escalated by
the ladder (a guard's behaviour changed, and more than three fix rounds ran).

Diff size measured before every pass by the diff-size guard, under the cap each time: 478, 784,
1260, 1411, 1426 changed lines. Exit 0 on all five, so the operator was never asked the
over-cap question.

## Optional slots

All four triggers fired against pass 1's diff — Security (path and file handling), Adversarial
(>300 changed lines, behaviour changes to guards with harnesses, three test files modified), Lens B
(>200 changed lines) and Lens C (error handling). Under the `light` preset they went to the operator
as one multi-select prompt.

**The operator selected none.** All four are recorded **declined**, which is distinct from a trigger
never firing. They were not offered again in later passes, and no conditional slot ran.

## Findings

| ID | Pass | Slot | Severity | Location | Note |
|---|---|---|---|---|---|
| F1 | 1 | Primary | Minor | `scripts/check-contract-budget.sh:76` | Task 5's own commit set the budget to exactly the file's size, zero headroom, against the guard's documented plus-25% formula. |
| F2 | 1 | Code review | Major | `scripts/gather-self-review-context.sh:562` | Resolves a change's commits by grepping the change name as the commit scope — the handle tasks 5 and 6 remove — so two of three lookups silently return nothing for every future change. |
| F3 | 1 | Code review | Minor | `scripts/test-check-guard-symlinks.sh:742` | A section numbered 8 sat physically before section 7, breaking a monotonic sequence. |
| F4 | 2 | Primary | Major | `scripts/gather-self-review-context.sh:633` | With the implementation commit skipped, an unrelated merge commit from another PR was reported as this change's implementation commit. |
| F5 | 2 | Primary | Minor | `scripts/gather-self-review-context.sh:616` | The archived pathspec earned nothing — `git log -- <path>` already matches across renames — and kept a duplicated if/else alive. |
| F6 | 2 | Principles | Important | `scripts/gather-self-review-context.sh:598` | The archive-shape regex hand-copied at four sites while a variable proving it works was used at one. |
| F7 | 2 | Principles | Minor | `scripts/gather-self-review-context.sh:622` | The plan-resolution pipeline duplicated verbatim across two branches differing only in pathspec. |
| F8 | 2 | Principles | Important | `scripts/test-gather-self-review-context.sh` | The refuse-rather-than-guess branch had no coverage at all; deleting the whole guard failed no test. |
| F9 | 2 | Code review | Major | `scripts/check-task-commit-fields.py:112` | The Conventional Commits breaking-change form `type(scope)!:` bypassed the scope check entirely. |
| F10 | 2 | Code review | Major | `scripts/check-guard-symlinks.sh:341` | Rule 5 iterated a scan set that excludes `skills/myflow-contracts/`, so a symlink there went undetected. |
| F11 | 2 | Code review | Major | `scripts/gather-self-review-context.sh:622` | Any later commit touching the plan path outranked the real planning commit by recency. |
| F12 | 2 | Code review | Major | `scripts/check-task-commit-fields.py:334` | The Jira-key extractor returned the name up to its first digit-only segment rather than a leading key. |
| F13 | 2 | Code review | Minor | `scripts/check-task-commit-fields.py:366` | Case-sensitive comparison let `feat(KAN-202)` pass — the shape a human is most likely to write. |
| F14 | 2 | Code review | Minor | `scripts/check-task-commit-fields.py:577` | The change name derived from a path has no defence against an archive-directory shape. |
| F15 | 2 | Code review | Minor | `scripts/gather-self-review-context.sh:638` | Two git subprocesses where one would return both values. |
| F16 | 2 | Code review | Minor | `scripts/test-check-task-commit-fields.sh:121` | Fixture sites hardcoded a change name that only happened to equal the default. |
| F17 | 2 | Code review | Minor | `scripts/gather-self-review-context.sh:596` | Two resolution strategies required a subject alternation kept in sync across three places. |
| F18 | 3 | Primary | Important | `scripts/gather-self-review-context.sh:590` | The old-era fallback's stated rationale was factually wrong and it was unreachable for any genuine historical commit; its only coverage came from empty-commit fixtures. |
| F19 | 3 | Principles | Minor | `scripts/gather-self-review-context.sh:672` | The deliberately redundant merge check read as an oversight with nothing saying it was intentional. |
| F20 | 3 | Principles | Minor | `scripts/gather-self-review-context.sh:606` | The fallback carried no stated retirement trigger, risking permanent cruft. |
| F21 | 3 | Code review | Minor | `scripts/gather-self-review-context.sh:667` | A root commit as the planning commit's parent made the diff command fail, silenced by a redirect, dropping a genuine implementation commit. |
| F22 | 4 | Principles | Minor | `scripts/gather-self-review-context.sh:640` | The acceptance logic had grown to four inlined conditions and warranted a named predicate. |
| F23 | 5 | Principles | Important | `scripts/gather-self-review-context.sh:508` | The predicate's comment did not record that three of its four conditions are deliberately dead today, leaving the next mutation pass no documented reason not to delete them. |

findings-total: 23
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
finding-status: F14 withdrawn operator-accepted as latent and unreachable through the guard's sole caller, which only ever passes the live path
finding-status: F15 withdrawn operator-accepted as a performance nit in a one-shot advisory script with no behavioural effect
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed

reproducers-total: 23
finding-reproducer: F23 none — documentation gap, the required text was quoted by the slot that raised it
finding-reproducer: F22 none — readability observation, no runtime defect
finding-reproducer: F21 scripts/test-gather-self-review-context.sh
finding-reproducer: F20 none — a missing retirement note, no runtime defect
finding-reproducer: F19 none — a missing comment, no runtime defect
finding-reproducer: F18 scripts/test-gather-self-review-context.sh
finding-reproducer: F17 none — maintainability observation, no runtime defect
finding-reproducer: F16 scripts/test-check-task-commit-fields.sh
finding-reproducer: F15 none — performance nit, no runtime defect
finding-reproducer: F14 none — latent, unreachable through the current sole caller
finding-reproducer: F13 scripts/test-check-task-commit-fields.sh
finding-reproducer: F12 scripts/test-check-task-commit-fields.sh
finding-reproducer: F11 scripts/test-gather-self-review-context.sh
finding-reproducer: F10 scripts/test-check-guard-symlinks.sh
finding-reproducer: F9 scripts/test-check-task-commit-fields.sh
finding-reproducer: F8 scripts/test-gather-self-review-context.sh
finding-reproducer: F7 none — duplication observation, no runtime defect
finding-reproducer: F6 none — duplication observation, no runtime defect
finding-reproducer: F5 none — simplicity observation, no runtime defect
finding-reproducer: F4 scripts/test-gather-self-review-context.sh
finding-reproducer: F3 none — section ordering, no runtime defect
finding-reproducer: F2 scripts/test-gather-self-review-context.sh
finding-reproducer: F1 scripts/check-contract-budget.sh

## Pass log

**Pass 1** — full roster, 478-line diff. Primary 1 Minor, Principles clean, Code review 1 Major and
1 Minor. The Major was the only one, and the cheapest slot found it.

**Pass 2** — Full, 784 lines. Nine findings, five Major. Every one was reproduced by the parent
before being dispatched to a fix round.

**Pass 3** — Full, 1260 lines. Four findings, **no Major**. Both substantive ones were found by
checking the code against reality — this repository's own archived history, and a git command run
directly — rather than against its tests, which were green throughout.

**Pass 4** — Full, 1411 lines. Primary clean, Code review clean, Principles 1 Minor.

**Pass 5** — Full, 1426 lines. Primary clean, Code review clean, Principles 1 Important. The
principles slot ruled on the parent's own mutation evidence rather than accepting its framing: the
parent showed one condition decides every case and proposed simplification; the slot verified all
three underlying claims, then ruled **keep all four**, because the redundancy of three of them rests
on contracts in files this function cannot see and has no test coupling to.

fix-mutation: `scripts/check-task-commit-fields.py` — optional `!` removed from the subject-scope pattern — 2 failures
fix-mutation: `scripts/check-task-commit-fields.py` — scope comparison returned to case-sensitive — 4 failures
fix-mutation: `scripts/check-guard-symlinks.sh` — rule 5 returned to the command-skills scan set — 2 failures
fix-mutation: `scripts/gather-self-review-context.sh` — plan-resolution subject filter removed — 10 failures
fix-mutation: `scripts/gather-self-review-context.sh` — outside-path requirement dropped — 1 failure
fix-mutation: `scripts/gather-self-review-context.sh` — both merge gates removed — 1 failure at the time, later shown over-determined by git's own merge default
fix-mutation: `scripts/gather-self-review-context.sh` — root-commit fix reverted — 1 failure naming the cause
fix-mutation: `scripts/gather-self-review-context.sh` — containment refusal disabled — 19 failures
fix-mutations-total: 8

**Two mutants were found by the parent and repaired in-round rather than raised as findings**, per
the fix-round rule: the first path-based resolution's fixtures left the live pathspec unexercised,
and the finding-G fixture merged an empty branch so its parent failed two conditions at once,
isolating neither.

**Three mutation results this run were false negatives caused by the mutation never applying** — a
`sed` with a bad delimiter, and two regexes that matched nothing. Each reported identically to a
genuine survivor. They were caught only by re-running through a parser that asserts the target text
was found first.

## Handback

The last change after pass 5's clean results was F23's fix, which is **comment-only**: 32
insertions and 1 deletion in one file, with **zero** non-comment lines changed, verified by diffing
the previous branch tip against the new one. The operator was asked whether to run a sixth pass on
that basis and chose to hand off, on the reasoning that the staleness rule exists to catch
unreviewed behaviour change and there is provably none.
