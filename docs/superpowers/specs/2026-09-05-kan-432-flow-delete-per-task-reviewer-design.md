# kan-432 — flow: delete the per-task reviewer

Source: KAN-432, decided in `docs/superpowers/research/flow-speedup.md` §8 "Per-task review
deleted" and its lever card "Delete the per-task reviewer".

## Why

Nineteen per-task reviews across seven changes yielded one `fix` outcome (kan-402 task 1), a
fence-regex fidelity gap the reviewer itself rated harmless, and its fix landed after task 2 had
already been implemented and reviewed against the unfixed task 1 — the "caught before the next
task builds on it" rationale did not hold in the one hit. Since KAN-436 the whole-branch `primary`
slot is that reviewer's spec-compliance mandate at branch granularity, `code-review-low` and
Bugbot are its code-quality half, and `check-task-commit-fields.sh` checks Files/Tests/Commit per
task mechanically. Deleting it saves one launch, one record pair and one tick decision per bundle
boundary (~15 s) plus the 55–90 s trailing solo reviewer: ~1–2 min and ~$0.2–0.7 per multi-task
change.

## What changes

Prose only. No script logic, no Go, no store change: `stats/cmd/flow/tick.go` and
`scripts/check-unfinished-work.sh` already know nothing about review, and `record.go` keeps the
`reviewer` role because the panel's slots record under it.

### `skills/flow/implement.md` §4 — the task boundary without a reviewer

`flow tasks tick` fires in the guard's own Bash call, once `check-task-commit-fields.sh` exits 0
for that task's commit. The boundary after bundle N+1 commits becomes:

1. The wait on bundle N+1's implementer report ends.
2. One Bash call: the implementer's `record dispatch end`, the guard on every new sha, `flow tasks
   tick` for every task whose guard passed, and bundle N+2's gather. A guard failure still sends
   the task back to the same implementer before anything launches and is never a fix-round slot.
3. One message launches bundle N+2's implementer; the next Bash call records its `begin`.

Deleted: the pending-fix fold (`--fixup` / `--autosquash` resumption under
`task-<n>-implementer-fix-<k>`), bundle N's re-reviewer, the **Per-task review** paragraph
(`-role reviewer` rows, `-outcome clean`/`fix`, the single combined reviewer), and **The last
bundle's reviewers run alone**. `final-review.diff` is written and the panel's slots dispatched
once the last bundle's guard has passed. The **Review overlaps the next implementer** heading
sentence becomes the boundary above; "one reviewer per commit" and the red/green shared reviewer
go with it. A red task's checkbox is still ticked together with its partner's, on their one
commit's guard pass.

Sentences reworded where they name a reviewer: the step table's row 6 becomes "the review panel |
the final whole-branch panel" (`requesting-code-review` was already dropped from `primary` by
KAN-436); the COMMIT-PER-TASK blockquote's "before the parent dispatches review for it" becomes
"before the guard runs on it"; "wait for every implementer and reviewer launched" becomes "every
implementer launched"; the tick sentence becomes "mark a task's checkbox `[x]` (`flow tasks
tick`) once `check-task-commit-fields.sh` passes on its commit; a step's checkbox tracks the step
and gates nothing". The REPRODUCE, DON'T READ and both FOREGROUND BUILDS blockquotes are the
conductor's and the implementer's own and stay where they are, so
`scripts/check-dispatch-paragraphs.sh`'s table is untouched.

### Sentences elsewhere that become false

- `skills/README.md:36` — step-table row 6, same rewording as `implement.md`'s.
- `skills/flow/brainstorm-planner.md` **D** — the mandated `tasks.md` header line "Mark a task's
  own checkbox when that task passes spec + quality review" becomes "…when
  `check-task-commit-fields.sh` passes on its commit". No guard checks that wording; archived
  plans keep the old header.
- `skills/flow/review-panel.md` — the docs-only reduction's rationale "the whole-branch read
  covers the text every per-task reviewer already read against the same plan" becomes KAN-436's
  own justification: the implementer's self-review and the vocabulary and reference guards cover
  prose; there is no code seam for a second slot to find. "This binds the review panel's fix
  round and not the per-task review's fix in `skills/flow/implement.md`" loses its second clause.
- `skills/flow-contracts/finish-contract-run1.md` — "TDD per task, per-task review, the final
  review panel" drops the middle item.
- Comments only: `scripts/check-dispatch-paragraphs.sh` (the KAN-263 site list and the
  FOREGROUND BUILDS table note name "the per-task reviewer dispatch" for `implement.md`'s second
  block — it is the conductor's own paragraph), `scripts/test-check-dispatch-paragraphs.sh` (case
  1's comment, same phrase), `scripts/check-panel-docs-only.sh` (KAN-312's "every per-task
  reviewer already read" line). The guards' behaviour is unchanged.

## What does not change

`stats/cmd/flow/tick.go`, `stats/cmd/flow/tasks.go`, `scripts/check-unfinished-work.sh`,
`scripts/check-task-commit-fields.sh`, `scripts/plan-dispatch-bundles.sh`, the store's dispatch
schema and `recordRoles`, every guard's table and exit contract, the review panel's roster and
fix round, and `docs/superpowers/research/flow-speedup.md` (a shared research note serving
several tickets; not this change's seed by the exact-filename rule).

## Decisions

### Tick on guard pass, in the guard's own Bash call

**ID:** tick-on-guard-pass
**Status:** active
**Chosen:** `flow tasks tick` rides the same Bash call as `check-task-commit-fields.sh`'s pass
— the guard is the only per-task verdict left, and a separate call would spend a turn on nothing.
**Considered:** tick after the whole-branch panel clears — rejected, a task's checkbox would then
say nothing about progress during the stage and `check-unfinished-work.sh` reads it; keep a
lighter reviewer (spec-only) — rejected, `primary` already carries exactly that mandate for the
whole branch.

### Sweep every sentence the deletion falsifies, comments included

**ID:** sweep-all-stale-mentions
**Status:** active
**Chosen:** every sentence in owned prose that asserts a per-task reviewer exists is fixed in
this change, and the three script comments naming one are corrected too, in one commit.
**Considered:** `implement.md` alone, as the ticket's "Touches" names — rejected, it leaves
`README.md`, `brainstorm-planner.md`'s header template, `review-panel.md` and
`finish-contract-run1.md` describing a step that no longer runs; prose only, leaving the script
comments — rejected, a comment naming a dispatch site that does not exist misdirects the next
guard edit.

### One task

**ID:** one-task
**Status:** active
**Chosen:** one task, one commit, every file — the edits are a single deletion's fallout and
no file is reviewable apart from `implement.md`'s change.
**Considered:** two tasks by file group — rejected, the second group's sentences are only
false once the first lands, so splitting them leaves an intermediate commit that contradicts
itself.

## Open questions

### Does a trailing reviewer make implementers commit more carefully?

**ID:** trailing-reviewer-deterrent
**Status:** open
**Why it is open:** unmeasured — the store records reviewer outcomes, not implementer behaviour
with and without one following; the research note records it as such.
**What it affects:** if the panel's per-change finding rate rises after this lands, the deterrent
was real and a cheaper substitute (a spot-check reviewer, or a stronger implementer self-review)
would be the follow-up; if it stays flat, nothing.
