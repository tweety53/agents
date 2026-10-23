## Context

Self-review fix change for `kan-1028-mutation-failure-modes-extra-dispatch`, planned on that
change's own branch `spectre/kan-1028-mutation-failure-modes-extra-dispatch` in its worktree at
`d7933ad6`. Five non-`big` findings, each carried by one task in `tasks.md`; each was traced at
`d7933ad6` before its fix was designed, and the branch's history since its base `53b0dfcd` holds
no earlier fix. Every question was auto-resolved to its recommended option (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). The brainstorm design, with the
per-finding trace, is
`.superpowers/sdd/2026-10-23-self-review-kan-1028-mutation-failure-modes-extra-dispatch-design.md`
in this worktree. Every reworded or added run-loaded sentence is listed in this change's
`verbatim-moves.txt`, checked with `check-verbatim-moves.sh` (exit 0) with all five tasks' diffs
applied.

## Decisions

### A dropped role leaves its decided bundle

**ID:** dropped-role-leaves-bundle
**Status:** active
**Chosen:** one sentence after the roster's drop rule in `skills/flow/review-panel.md`: a dropped role leaves its `panel.dispatches` bundle too, the bundle dispatches the roles it still names, and a bundle left naming none is not dispatched.
**Considered:** the sentence at the dispatch site (`write-panel-diff.sh`'s paragraph) — the drop rule is where a reader learns a role leaves, and a second site restates it; rewriting a stored decision's `panel.dispatches` — a recorded decision is never rewritten, and the defect is latent, with nothing written wrong.

### The status template prints the Learning path

**ID:** status-template-learning-line
**Status:** active
**Chosen:** `skills/flow-contracts/handoff-blocks.md`'s `IN_PROGRESS` template carries `**Learning:** (run-only) <the learning note's absolute path, when composed — …>` in the line where the section stood, after `**Decisions:**`, as the run's own block places it.
**Considered:** dropping the line from the template — the template is the definition `/flow-status` regenerates from, and the run prints the line.

### Retire failure-modes and mutation from ValidReviewers

**ID:** retire-extra-dispatch-reviewers
**Status:** active
**Chosen:** remove both ids from `ValidReviewers` (`stats/internal/store/settings.go`), strip them from a stored `flow_settings` list with `0044_flow_settings_drop_extra_dispatch_roles.sql` modelled on `0033`, update the tests that wrote them as valid, add one test for the strip and one for the refusal, and drop `failure-modes` from `skills/flow-settings/SKILL.md`'s example warning line. The skill reads its offer from `ValidReviewers`, so its "every id is dispatched" sentence is true again with no edit.
**Considered:** keeping the ids and adding a sentence to `/flow-settings` that they are never dispatched — the offer would still present two choices that do nothing; filtering them in the skill alone — a copy of the vocabulary outside its one canonical place.

### Store repair for the settings row

**ID:** settings-row-no-repair
**Status:** active
**Chosen:** no repair: the stored row is `["primary","principles"]` (`flow settings get`), so the migration is a no-op on it and nothing was written wrong.
**Considered:** an explicit `flow settings set` — writes the same list back.

### A session token ends in a UTC timestamp

**ID:** session-token-utc-suffix
**Status:** active
**Chosen:** **Stage marks** (`skills/flow-contracts/pipeline.md`) requires the token to end in the run's start time, UTC to the second, read once with `date -u +%Y%m%dT%H%M%SZ` and typed as a literal; `skills/flow/SKILL.md` and `skills/flow-fast/SKILL.md` cite that shape instead of restating it. `skills/flow-plan/SKILL.md` already cites it.
**Considered:** a CLI refusal of a token already bound to another session — the binding is resolved after the fact from transcripts, so the CLI cannot know at mark time; a random suffix — an agent typing a "random" literal repeats itself, where a clock read does not.

### A task added mid-run is re-decided

**ID:** mid-run-task-re-decided
**Status:** active
**Chosen:** one paragraph after **A pivot reconciles the three artifacts together** in `skills/flow/implement.md`: whenever `tasks.md` gains a task once the decision is recorded — a pivot, an operator's mid-run request, any append outside a fix run's — the parent re-runs **Decide** and records the new row as `document-fix.md` re-decides an appended plan, before the added task is implemented.
**Considered:** a guard comparing the newest row's `inputs` to `plan-class.sh` at verify — a second mechanism for a rule the parent applies at the one moment it appends.

### kan-1028's decision row is corrected

**ID:** k1028-decision-row-repair
**Status:** active
**Chosen:** record a second decision row for `kan-1028-mutation-failure-modes-extra-dispatch` from its worktree's `decision.json` with `inputs.tasks` and `inputs.files` set to the figures `plan-class.sh` prints over the archived `tasks.md`, every other field unchanged; class, panel and rolls already match.
**Considered:** adding `extra_dispatch_mutation` to the row — no such dispatch ran on that run, and the row would claim a pair nothing used; leaving the row — a forward-only fix leaves the record wrong (**3. Fix every finding that is not `big`**, `skills/flow-self-review/SKILL.md`).

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 3 retires a validator's vocabulary behind a migration, and Task 5 writes a row to the persistent store.
**Considered:** `flow-low` — three of the five tasks are one-sentence prose, but a migration that strips the wrong id would empty the operator's panel.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-1028-mutation-failure-modes-extra-dispatch-fix`, Tasks 1 to 5 in order; the tasks are file-disjoint.
**Considered:** one fixer per task — five dispatches for five small diffs.

## Open questions

## Live check

Task 5 writes a decision row and Task 3 adds a migration, the two runtime touches. Before:
`flow record decisions -change kan-1028-mutation-failure-modes-extra-dispatch -C <this worktree> | jq 'length, .[-1].decision.inputs'`
prints a length of one and inputs `tasks 5, files 20`; `flow settings get` prints
`["primary","principles"]`.
<!-- measured: flow record decisions -change kan-1028-mutation-failure-modes-extra-dispatch -C . @ branch spectre/kan-1028-mutation-failure-modes-extra-dispatch (live store read, d7933ad6) -->
After Task 5: a length of two, and the newest row's inputs `tasks 6, files 23`, class `regular`.
<!-- measured: scripts/plan-class.sh spectre/changes/archive/kan-1028-mutation-failure-modes-extra-dispatch/tasks.md 1 . 53b0dfcd @ branch spectre/kan-1028-mutation-failure-modes-extra-dispatch -->
Task 3's migration runs only when `flowd` next starts, which no step of this change does; its
effect is proved by `TestSettingsStore_MigrationDropsExtraDispatchRoles` on a throwaway database.
Not working looks like a length of one, or a newest row still reading `tasks 5`. Task 5's own
steps run it; no separate verification stage runs on a self-review branch.
