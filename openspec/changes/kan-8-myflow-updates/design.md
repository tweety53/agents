## Context

The full design narrative lives at
`/Users/tweety53/Projects/agents/docs/superpowers/specs/2026-07-28-kan-8-myflow-updates-design.md`
(committed as `f93a3d5`). This file is the OpenSpec-shaped view of it.

myflow today has twelve stages, eighteen commands duplicated across two harness command trees,
and nineteen skills. The stage field encodes two facts at once — how far the work got, and who
is waiting — which is why the model needs seven commands that only write state, an
`originStage` field to remember where a fix was raised, and a monotonic-gates contract to keep
those recorded confirmations from being demoted.

Two immediate predecessors matter. KAN-10 (`c93dfac`) moved four contract sections out of the
always-on rule layer into `skills/myflow-contracts/`. `myflow/lazy-load-pipeline` (`2e3f973`)
extended that to the pipeline itself, leaving `rules/myflow-manual-review.mdc` a 339-word stub
and putting the state machine in `skills/myflow-contracts/pipeline.md`. KAN-8 rewrites what
lives beneath that split; it does not revisit the split.

Constraint worth stating: nothing lets a running agent session prefill or preselect text in the
operator's input box in any supported harness. The ticket's "next command suggested as
autocomplete" is therefore not implementable as literally written, and is met instead by output
convention plus a smaller command surface.

## Goals / Non-Goals

**Goals:**

- Three states whose names are the same vocabulary as the commands that produce them.
- One human gate: review the diff and run the apps at a single stop.
- A command surface small enough that the legal next step is obvious without consulting a table.
- No command whose only effect is to record that a human looked at something.
- `/myflow-finish` leaves no residue: archive committed and pushed, worktrees and branches gone.
- Handoffs that contain only what the operator must act on, with absolute paths.

**Non-Goals:**

- Backward compatibility. The ticket explicitly waives it; no aliases, deprecation shims, or
  migration of existing state files are written.
- Revisiting the always-on / on-demand split. `2e3f973` settled it.
- Re-running `setup.sh global` on this machine — an operator action, not a code change.
- Changing the review panel's **roster** or its trigger table. Both are untouched. Only the model
  selection changes (every slot to Sonnet) and the `full-panel` override is removed; the
  targeted-re-run escalation triggers stay exactly as they are.

## Decisions

### What `/myflow-review` does in the five-command flow

**ID:** review-role
**Status:** superseded by three-state-collapse
**Chosen:** Keep commit + push + open PR. The `*-manual-review` marker commands are deleted
outright, and reviewing the staged diff becomes the unmarked human gate of `IN_PROGRESS`. This
is what makes five states map one-to-one onto five commands.
**Considered:** Making `/myflow-review` the manual-review marker and moving commit/push/PR into
`/myflow-finish` — rejected because finish would become a single irreversible step (commit,
push, PR, merge, sync, archive, delete worktrees) with no gate between opening a PR and
destroying the worktree that produced it.

### Where a fix leaves the state

**ID:** fix-re-entry
**Status:** active
**Chosen:** `/myflow-do` leaves the state untouched except `STARTED` → `IN_PROGRESS`. This
removes `originStage` and the entire fix re-entry table; the operator re-runs an earlier
command when they want an earlier gate re-opened.
**Considered:** Always dropping back to `IN_PROGRESS` — safest, since new code could never pass
a gate that ran before it existed, but a one-line fix at `REVIEW` would cost a full re-walk of
test and review. Also considered recording a `staleAfterFix` marker so `/myflow-review` could
warn — rejected as reintroducing exactly the bookkeeping field this decision exists to delete.

### How the manual test is skipped

**ID:** test-skip-mechanism
**Status:** superseded by one-human-gate
**Chosen:** Skipping is not running `/myflow-test`. `/myflow-review` accepts `IN_PROGRESS` or
`TEST` and records `tested: "skipped"` for the former.
**Considered:** Keeping the always-asked skip prompt and writing a guide marked `SKIPPED` —
rejected because the ticket asks for no docs and no analysis when skipping, and because the
prompt's `"skipped"` value is what forced the sticky-gate monotonicity rules.

### How the next command is surfaced

**ID:** next-command-hint
**Status:** active
**Chosen:** Print it bare as the final line of every handoff, with nothing after it. Works in
every harness and needs no harness feature.
**Considered:** A Claude Code `Stop` hook printing the next command — rejected: it still cannot
prefill the input box, it is Claude-Code-only, and it would run on every stop in every project,
against this ticket's own simplification goal. Driving autocomplete directly was investigated
and is not possible.

### When `/myflow-review` asks about merging

**ID:** merge-choice-timing
**Status:** superseded by finish-absorbs-integration
**Chosen:** Ask before doing any work, then run to completion unattended.
**Considered:** Asking after the PR is open, so the choice is made with test and coverage
results in hand — rejected because it puts a stop in the middle of a run the operator wanted to
leave alone. Considered dropping auto-merge entirely — rejected as losing the one-command path
for trivial changes.

### Fate of the `/opsx:*` commands

**ID:** opsx-removal
**Status:** active
**Chosen:** Delete the five that duplicate pipeline steps; keep `/opsx:explore`, which has no
myflow equivalent and touches no state.
**Considered:** Keeping all of them — they cost nothing at session time, but each is another
file every future rename must sweep. Considered deleting all six — rejected because explore is
a thinking mode, not a pipeline shortcut.

### Whether any flag survives

**ID:** flag-removal
**Status:** active
**Chosen:** No command accepts any flag. The only argument is the optional change name. This
extends the ticket's "Flags don't needed" from the cleanup step to the whole command surface,
and takes `full-panel` with it.
**Considered:** Keeping `full-panel` on `/myflow-do` on the grounds that it selects reviewer
breadth rather than pipeline shape — rejected because "no flags" is a contract an operator can
hold in their head, while "one flag, on one command, for one reason" is a footnote they have to
look up. The cost is real and accepted: forcing a full whole-branch panel is no longer possible,
so panel breadth rests entirely on the automatic escalation triggers. Also considered replacing
it with a question, as the merge choice was — rejected because it would interrupt every fix
round to ask something whose answer is almost always the default.

### Skill layout

**ID:** skill-layout
**Status:** active
**Chosen:** One skill per command, named after it — nineteen skills to ten.
**Considered:** Minimal churn, rewriting contents while keeping filenames — rejected because
the simplification would then be true of the docs but not of the tree, and `/myflow-do` would
still load one of two skills depending on whether it was a fix. Considered a single dispatching
skill branching on state — rejected because every command would load all five stages'
instructions, directly against the token work `2e3f973` just landed.

### Collapsing to three states

**ID:** three-state-collapse
**Status:** active
**Chosen:** `STARTED`, `IN_PROGRESS`, `FINISHED`, with `/myflow-test` and `/myflow-review` deleted
outright. Supersedes `review-role`, which kept `/myflow-review` as the integration command.
**Considered:** the five-state model this change originally specified (start / do / test / review
/ finish) — rejected because two of its five states existed only to separate work the operator
does in one sitting, and a third existed only to run verification that `/myflow-do` had already
done. Every state that is not a distinct human decision is a state the operator has to remember
the rules for.

### Reviewing and testing are one gate

**ID:** one-human-gate
**Status:** active
**Chosen:** `/myflow-do` emits the staged diff **and** the manual test guide, so the human reviews
and tests at a single stop. Supersedes `test-skip-mechanism`, which made skipping the test a
matter of not running a separate command.
**Considered:** keeping a separate `/myflow-test`, on the grounds that a guide written before the
diff is reviewed may describe behaviour that review changes — rejected because a fix re-runs
`/myflow-do`, which refreshes both surfaces together, so they cannot drift. Skipping the manual
test now needs no mechanism at all: the guide is simply there to use or ignore.

### Finish absorbs integration

**ID:** finish-absorbs-integration
**Status:** active
**Chosen:** `/myflow-finish` runs twice — once to land the branch (PR by default, merge, or
manual), once to archive after the merge. Supersedes `merge-choice-timing`, which put the same
question at the start of `/myflow-review`.
**Considered:** a separate integration command — rejected as reintroducing the state this change
removes. Considered branching on a recorded "integration started" field — rejected because the
branch's merge status already answers it, and a field would be a second source of truth that can
disagree with git. The consequence is that `/myflow-finish` is the answer to "what do I run next"
twice in a row, which the handoff states explicitly.

### No verification gate before integration

**ID:** no-preflight-verification
**Status:** active
**Chosen:** no tests, linters, or spec-coverage check run before a PR or merge.
**Considered:** keeping the coverage check, since it is the only thing that compared delta-spec
scenarios against actual tests — rejected as the user's explicit instruction, and because it ran
at the point of least leverage: the work is finished, the human has already reviewed it, and a gap
found there routes back to `/myflow-do` anyway. Correctness now rests on TDD per task, per-task
review, the final panel, and the human gate. **Accepted cost:** a change can be merged with a
delta-spec scenario that no test covers, and nothing will say so.

### One model for the whole review panel

**ID:** panel-model-uniform
**Status:** active
**Chosen:** every panel slot runs on Sonnet. The roster and trigger table are unchanged.
**Considered:** the previous split — parent model for the primary and principles slots, an economy
tier resolved from the parent's provider family for the rest — rejected because it made the
panel's cost depend on which model the operator happened to be running, and required a
provider-detection table that had to be maintained as model names changed. A single named model is
predictable and deletes that table.

## Risks / Trade-offs

**A fix does not re-open the gate that already ran** → Accepted deliberately (see `fix-re-entry`).
With one human gate there is no cross-stage staleness to report; the operator re-reviews the diff
or re-runs the apps if they judge the fix warrants it.

**Deleting a delegated-to skill breaks its consumer** → finish delegated into
`openspec-archive-change` and `openspec-sync-specs`; start into `openspec-propose`; do into
`openspec-manual-test-superpowers`. Sequenced in `tasks.md`: inline the used content into the
surviving skill and verify it, then delete the source. Never the reverse order.

**A rename this wide leaves stale references behind** → `scripts/check-references.sh` catches a
reference to a section name that no longer exists; `scripts/check-vocabulary.sh` catches retired
literals once taught them. Neither proves completeness — the vocabulary guard matches literals
only and cannot know about a rename until it is added to it. Manual sweep of each layer is still
what makes the rename done, and the task list enumerates the layers.

**`git worktree remove --force` can destroy uncommitted work** → Four gating checks run first and
all must pass. `--force` still destroys every *ignored* file and no check prevents that, so the
ignored files are **disclosed and confirmed** rather than silently removed — "ignored" is not
"disposable", and this pipeline's own `.superpowers/sdd/` records are ignored. Any failed check
leaves every worktree alone. `git branch -d` is never `-D`, so an unmerged branch refuses
deletion on its own.

**`setup.sh` inlines rule text rather than symlinking it** → An edit to
`rules/myflow-manual-review.mdc` changes nothing for Claude Code or Codex until `setup.sh
global` re-runs. Verification uses a sandboxed `HOME` so the real home directory is never
written to during implementation; re-installing is the operator's call afterwards.

**Existing state files use the old shape** → All nine on this machine are `finished`, so no
migration is written. Self-heal rewrites any file it cannot parse, which is the fallback if one
appears.

## Migration Plan

No runtime migration. The rollout is: land the change, then run `setup.sh global` once to
publish the new commands, skills and rule text to every harness. Rollback is `git revert` of the
merge plus the same re-install, since every artifact is a file in this repository.

## Open Questions

None. The five decisions above were settled during brainstorming; `/opsx:explore`'s survival is
the only judgment call taken without an explicit instruction, and it is called out in the
proposal so it can be reversed cheaply.
