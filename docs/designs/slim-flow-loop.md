# slim-flow-loop
> Design record of change slim-flow-loop, written at landing. Each decision's own **Status:** line is the status.

## Context

The ticket names eleven cuts to the `/flow` loop; the operator kept items 1 (the `/clear` handoff
between planning and implementing) and 5 (the review panel's composition), withdrew cut 4 (same-run
cleanup on the open-PR route) in planning round 1, and added a guard-cut scope mid-planning. The cuts
are one change because they rewrite the same pipeline files — `skills/flow/*.md`,
`skills/flow-contracts/pipeline.md`, the `integrate-change` and `visual-verify-pass` guards, the
stage vocabulary — and each cut changes what another cut's text says: the fix procedure writes the
new stage marks, the self-review move removes the land gate the stage vocabulary would otherwise
rename, and the run-instructions removal deletes the state gate the stage retirement would break.

Constraints that shape the plan:

- `flow-guard check-guard-calls` rule A refuses any skill, rule, hook or `CLAUDE.md` text naming an
  unregistered guard, and rule C requires a `_test.go` per registered `check-*` guard. A guard leaves
  the registry in the same task that removes its last prose citation.
- `flow-guard check-references` refuses a citation of a missing file or section; a deleted file
  leaves with every citation of it.
- `stages.Validate` refuses any key outside `stages.Table` for a new mark;
  `TestStagesMatchReadmeLevelOne` keeps the Table and `README.md`'s Level 1 table row for row; the
  SPA's `phase.test.ts` reads `stats/internal/stages/names.go` and requires a phase for every key.
- `stats/internal/api/changes.go` decodes the change body with `DisallowUnknownFields`, so a legacy
  `reviewBeforeLanding` key must be accepted explicitly to keep an old journal entry replayable.
- Role rows live in flowd's `role_pairs` table; a role is retired by a new migration, never by
  editing an applied one. The newest applied migration is `0075`.
- `stats/internal/store/aggregate.go`'s planning category already keeps two retired keys
  (`flow.design-approval`, `flow.plan-review`) so their cost still counts — the precedent for
  retiring keys without remapping stored rows.

## Decisions

### One fix procedure file

**ID:** one-fix-procedure-file
**Status:** active
**Chosen:** `skills/flow/fix.md` replaces `skills/flow/document-fix.md` and `skills/flow/verify-fix-loop.md`; the loop's rules an implementer needs (shared-source fixes with recapture, per-capture `--update-snapshots` re-baselining, the `drift-<slug>` decision, no per-screen offsets, probes, `measure-visual-properties`, the running-clock `--repeat-each 5` rule, the "passing test is not the spec" tie-breaker, MUTATION PROOF) move into its step-2 brief — one file serves the in-run fix and the fix run.
**Considered:** keeping both files trimmed — two files describing one procedure drift apart and double the load set.

### The fix procedure serves fix runs too

**ID:** fix-procedure-serves-fix-runs
**Status:** active
**Chosen:** a fix run at `IN_PROGRESS` (`/flow <fix instructions>` or a plain message) runs the same four steps; its step 3 runs the lint/test list, then visual verify only when the fix touched `ui paths` (operator, round 1 Q4: "Same procedure").
**Considered:** a separate fix-run procedure — two procedures for one job; re-running visual verify always on a fix run — a capture for a fix that touched no UI path proves nothing.

### Fix instructions and fix-start shas live in scratch

**ID:** fix-start-in-scratch
**Status:** active
**Chosen:** step 1 writes `<abs-worktree>/.superpowers/sdd/fix-<k>.md` with the fix instructions verbatim and each worktree's `git rev-parse HEAD` as its fix-start sha; nothing is appended to `proposal.md` or `tasks.md`; no planner, no `spectre validate`, no Jira description sync.
**Considered:** appending an `In-run fix <k>` task to `tasks.md` — it needs a planner and `spectre validate`, the cost cut 7 removes.

### The fix commit carries a Flow-Key and no Task-Id

**ID:** fix-commit-flow-key
**Status:** active
**Chosen:** the fix implementer's commit carries `Flow-Key: fix-<k>-implementer` (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`) and no `Task-Id:`, so no `close-task` runs.
**Considered:** a synthetic task id — `close-task` and `check-unfinished-work` would look for a `tasks.md` task that does not exist.

### The one reviewer is the panel's primary slot

**ID:** single-reviewer-is-primary
**Status:** active
**Chosen:** step 4 runs `flow-guard write-panel-diff late-fix` from the fix-start shas and one `primary` slot (`primary-reviewer-prompt.md`, role `panel-reviewer`, key `fix-<k>-review`) over `late-fix.diff`.
**Considered:** the gated reviewer — it is task-scoped and needs a `tasks.md` task id the fix no longer has.

### A reviewer's Critical or Important finding opens the next fix

**ID:** review-findings-open-next-fix
**Status:** active
**Chosen:** a Critical or Important finding opens fix `<k+1>` through the same procedure; Minors are fixed by that next fix; the no-progress rule's "three failed rounds" counts fix rounds.
**Considered:** escalating to the full decided panel — the re-run of the whole panel is what cut 2 removes.

### The re-run check keeps its -fix-k keying

**ID:** keep-fix-k-keying
**Status:** active
**Chosen:** step 3 re-runs the stage that reported the defect under today's `-fix-<k>` keys and narrowing; only the `-fix-<k>-full` run goes.
**Considered:** new keys — `flow record dispatch begin`, `check-verify-green`'s suffix folding and the round-base logic already read `-fix-<k>`.

### A visual fix re-runs lint and test first

**ID:** visual-fix-reruns-lint-test
**Status:** active
**Chosen:** when visual verify failed and the fix changed code since the last green lint/test run, the lint/test list re-runs before visual verify, so the landing gate is unchanged (operator, round 1 Q5: "Re-run lint/test too").
**Considered:** re-running visual verify alone — code a visual fix changed would land without a green lint/test row.

### Fix steps mark the phase they belong to

**ID:** fix-marks-by-step
**Status:** active
**Chosen:** step 2 runs under `flow.implement`, step 3 under `flow.verify`, step 4 under `flow.review`; `autoStageFor` maps a `fix-<k>-review` key to `flow.review`.
**Considered:** a fix-only mark — an eighth mark the seven-phase cut does not have.

### Every change starts short and escalates

**ID:** start-short-escalate
**Status:** active
**Chosen:** plan triage goes; every new `/flow` change starts with the short planner (`planner-short`, key `plan-short-<n>`); a short round that estimates `large` returns `ESCALATE`; `/flow <key> full` forces full planning (operator, round 1 Q2: "Start short, escalate").
**Considered:** keeping triage — a headless call before planning whose only job the short planner's own size estimate already does; always planning full — the cost short planning exists to save.

### triage.md is the escalation record

**ID:** triage-file-as-escalation-record
**Status:** active
**Chosen:** `.superpowers/sdd/planner/triage.md` is written `TRIAGE: full — <reason>` only on `ESCALATE` or a `full` argument; a resumed run plans full exactly when it exists.
**Considered:** a new file name — every resume reader already reads `triage.md`.

### Plan review and Decide merge into plan-check

**ID:** plan-check-merged-pass
**Status:** active
**Chosen:** one headless guard, `flow-guard plan-check <worktree> <changeRoot> <decision.json> [-decide-only]`, replaces the `plan-review-<n>` dispatch and `flow-guard decide-groups-panel-visual`; it writes `.superpowers/sdd/plan-check.md` (findings in today's `F<k> [blocker|finding|advisory]` shape, last line `PLAN-CHECK-CLEAN`, `PLAN-CHECK-FINDINGS <n>`, or `PLAN-CHECK-NOT-RUN` on the findings fallback) and `decision.json` in today's schema; its checks and fixed fallback are `decide-groups-panel-visual`'s, plus a findings fallback of `not run`; `plan-reviewer-prompt.md` and `decider-prompt.md` merge into `skills/flow/plan-check-prompt.md`.
**Considered:** keeping the review dispatch beside Decide — two passes over the same three artifacts.

### plan-check reads the code with read-only tools

**ID:** plan-check-read-only-tools
**Status:** active
**Chosen:** `plan-check` runs headless with `Read`, `Grep`, `Glob` in the worktree under a time limit of its own (operator, round 1 Q3: "Read-only tools").
**Considered:** no tools with an inlined digest — the review's evidence quality drops to what the digest carries.

### pjRun takes the tool list and the time limit

**ID:** pjrun-tools-and-limit
**Status:** active
**Chosen:** `pjRun` (`stats/internal/guard/pairjudge.go`) takes the tool list and the time limit as arguments; the pair judge passes no tools and its 60-second limit.
**Considered:** a second headless runner — a copy of `pjRun` for one argument difference.

### plan-check runs on the decider role

**ID:** plan-check-on-decider-role
**Status:** active
**Chosen:** `plan-check` runs on the `decider` role's pair; `plan-triage` and `design-reviewer` retire by a new migration.
**Considered:** a new role — a third copy of the decider's pairs.

### A triage that edits tasks.md re-decides alone

**ID:** decide-only-rerun
**Status:** active
**Chosen:** on findings a fresh final planner triages `plan-check.md` as today; when its triage changed `tasks.md`, `plan-check -decide-only` re-runs the decision alone — the same call a story split and a gate **No** make; nothing is reviewed twice.
**Considered:** re-reviewing — a second review the old flow never ran; deciding on the stale plan — groups that name tasks the triage changed.

### plan-check sizes a docs-only plan to primary

**ID:** docs-only-in-plan-check
**Status:** active
**Chosen:** `plan-check-prompt.md` sizes the panel to `primary` alone when every task's `**Files:**` is documentation (`.md`, `.mdc`, or `.html` under `docs/`), which is what `check-panel-docs-only` decided after the fact; `check-panel-docs-only` is deleted.
**Considered:** a docs-only diff running the full decided panel — the cost the guard existed to save.

### /flow-fast decides with plan-check -decide-only

**ID:** flow-fast-decide-only
**Status:** active
**Chosen:** `/flow-fast`'s decide step calls `plan-check -decide-only`.
**Considered:** keeping `decide-groups-panel-visual` for `/flow-fast` alone — two guards for one decision.

### The open-PR route keeps its two runs

**ID:** open-pr-keeps-two-runs
**Status:** active
**Chosen:** ticket cut 4 is withdrawn; the open-PR and manual routes end run 1 at `IN_PROGRESS` and a bare `/flow <name>` after the merge runs cleanup, as today (operator, round 1 Q1: "lets have it as is tbh, this is overengineering"; confirmed round 2).
**Considered:** an await-merge watcher in `land` or `cleanup` — machinery the operator called overengineering.

### land no longer calls cleanup

**ID:** land-stops-calling-cleanup
**Status:** active
**Chosen:** `integrate-change land` loses its self-review gate and, on merge-and-push, no longer calls `cleanup()`; the session runs `prepare`, `commit`, `land`, then the self-review pass from `flow self-review bundle`, then on merge-and-push `integrate-change cleanup` in the same invocation.
**Considered:** cleanup before the self-review — cleanup removes the worktree the bundle is built from.

### Cleanup gates on the self-review

**ID:** self-review-gates-cleanup
**Status:** active
**Chosen:** `integrate-change cleanup` refuses while the change holds no completed `flow.self-review` run.
**Considered:** no gate — a run that skipped the pass would remove the worktree and lose the bundle.

### Self-review fixes always land on their own branch

**ID:** self-review-fixes-own-branch
**Status:** active
**Chosen:** fixes go on `self-review-<name>` and land by `flow-guard pipeline-fix land`; the `fixes-ride-the-change` case, `refreshFixedRefs` and the reshape's survival of fixed shas go.
**Considered:** pushing fixes onto the open PR — the branch has already landed on merge-and-push, and a PR-only rule would split the procedure by route.

### reviewBeforeLanding: drop the column, ignore the legacy key

**ID:** drop-column-ignore-legacy-key
**Status:** active
**Chosen:** migration `0076_drop_review_before_landing.sql` drops `changes.review_before_landing` (0045 untouched); the store field, the DTO field and their three tests go; the change-body decoder accepts and ignores a legacy `reviewBeforeLanding` key.
**Considered:** keeping a dead column — a field nothing writes or reads; refusing the legacy key — an old journal entry would no longer replay.

### The IN_PROGRESS run-instructions gate goes

**ID:** drop-in-progress-gate
**Status:** active
**Chosen:** `flow state set`'s `gateInProgressHandoff` and its test go with the run instructions.
**Considered:** re-pointing it at another mark — no step it could require remains between verify and the state write.

### An open question still stops at IN_PROGRESS

**ID:** open-question-still-stops
**Status:** active
**Chosen:** a run whose `Decisions:` carries a `?` still stops at `IN_PROGRESS` with the remaining handoff block (Summary, Decisions, Next), with no `**Running:**` section.
**Considered:** landing past it — an open operator question would land unanswered.

### Every visual departure blocks

**ID:** no-visual-attribution
**Status:** active
**Chosen:** with the merge-base capture gone, step 10 drops attribution: every departure from the mockup or the approved prototype is a blocking item fixed in-run; the 2x mis-measurement check stays. The `--merge-base` diff argument, the `<merge-base>` placeholder of `specs` and `regression checkout` stay.
**Considered:** an operator-visible "unattributed" status — a status with no capture to back it.

### Tooling analysis is a standalone skill

**ID:** tooling-analysis-skill
**Status:** active
**Chosen:** `skills/flow-visual-tooling-analysis/SKILL.md`, run by the operator on a change whose verify reports exist, inline in the operator's session, dispatching nothing; `prepare --miss`, `NEXT: analyst`, the `tooling-analysis-<n>` key shape, the `tooling-analyst` role and `skills/flow/visual-verify-tooling-analysis.md` go.
**Considered:** keeping the prompt file loose — nothing would load it; deleting it — the operator keeps the analysis as a manual tool (ticket cut 9).

### Folded guards keep their checks in place, in-process

**ID:** guard-checks-in-process
**Status:** active
**Chosen:** `check-frame-list`, `check-verify-report`, `check-sidecar-reach`, `check-verify-green`, `check-visual-verification`, `check-visual-preflight` and `check-workspace-isolation` leave `guard.Registry`; each check survives as a Go function called where it runs today — verify-green, sidecar-reach, pre-flight and the `## visual verification` validation inside `visual-verify-pass prepare`, report and frame-list inside `close`, and `prepare-workspace` validating and parsing its isolation table in-process. Exit codes, reason ids and messages are unchanged (operator, round 2 Q1: "Stay in place, in-process").
**Considered:** moving the pre-capture checks to `close` — a red verify or a missing sidecar would be found after a full capture.

### The visual verification section is validated at prepare

**ID:** visual-config-check-at-prepare
**Status:** active
**Chosen:** `check-visual-verification`'s validation runs at the start of `visual-verify-pass prepare` and leaves this repository's `## lint`.
**Considered:** at `close` — the other visual guards assume a validated section from the start.

### prepare-workspace parses isolation in-process

**ID:** isolation-parse-in-process
**Status:** active
**Chosen:** `prepare-workspace` calls the isolation validator and row parser directly; the `#ROW` print contract, `CHECK_WORKSPACE_ISOLATION_PRINT_ROWS` and `FLOW_GUARD_WORKSPACE_ISOLATION` go; an invalid table still exits with the same `ISOLATION-INVALID` lines; `wiValidate`, `wiSpace` and `wiAbs` stay.
**Considered:** shelling out to an unregistered entry point — a subprocess round trip to call a function in the same binary.

### The verifier self-checks with close --check-only

**ID:** verifier-check-only
**Status:** active
**Chosen:** the bundle's two `check:` lines become one, `flow-guard visual-verify-pass close --check-only <report>`, which runs `close`'s two report checks and writes no row, state or verdict.
**Considered:** refusing at `close` and re-dispatching — a full verifier round per gap, losing the verifier's in-context fix.

### break-and-prove is the one mutation guard

**ID:** keep-break-and-prove-only
**Status:** active
**Chosen:** `mutate-and-verify` is deleted; a flip across several files or harnesses becomes one `break-and-prove --patch` call per file and test command; `mvRunBounded` moves into `resolvebasebranch.go`, its remaining caller; `postmutationcheck.go` stays (operator, round 2 Q2: "Keep break-and-prove only").
**Considered:** merging the two — restore semantics differ and the merge is about fourteen files of behaviour risk; keeping both optional — two tools for one proof; a new shared helper file for `mvRunBounded` — one caller.

### Slot briefs are filled by hand

**ID:** slot-brief-by-hand
**Status:** active
**Chosen:** `render-slot-prompt` is deleted; the parent writes each slot brief to the same `slot-prompt-<round>-<slots>[-repo].md` path: the slot's `<id>-reviewer-prompt.md` fenced body with the placeholder table's six values filled, the shared blockquotes of `review-panel.md` copied verbatim, the mutation pass built last from **The mutation-testing brief**, and the FIX-ROUND SCOPE block on a fix round.
**Considered:** a new template document — a second copy of the templates the placeholder table already indexes.

### fix-slot-judge is deleted, not merged

**ID:** fix-slot-judge-deleted
**Status:** active
**Chosen:** `fix-slot-judge` goes with the fix-run panel it narrowed; nothing is merged into `pair-judge`.
**Considered:** merging it into `pair-judge` (the operator's wording) — the one-reviewer fix path leaves it no slot set to narrow.

### Repeated checks run once, in verify

**ID:** repeated-checks-once-in-verify
**Status:** active
**Chosen:** `check-dev-stack-fresh` runs only in verify's test group while a stack is up; its calls in `implement.md`, the Live check's repeat and `visual-verify-pass prepare`'s step-6 fingerprint with its `NEXT: restart-stack` go. `check-spec-reach` runs only as `run-verify-list`'s appended lint call; the verifier's post-capture call and `verify-and-handoff.md`'s separate invocation go.
**Considered:** keeping the verifier's post-capture spec-reach call — a third run of the same check.

### This repository's capture rows keep their freshness prefix

**ID:** project-rows-keep-freshness
**Status:** active
**Chosen:** `.flow/project.md`'s `verify` and `capture` rows keep `check-dev-stack-fresh ../.. || make ui-test-up` — project configuration, not a pipeline call.
**Considered:** editing the project's rows — they are the restart-on-stale this repository relies on.

### Seven phase marks with new names

**ID:** seven-phase-marks
**Status:** active
**Chosen:** `flow.plan`, `flow.implement`, `flow.review`, `flow.verify`, `flow.land`, `flow.self-review`, `flow.cleanup`, adopted by `/flow`, `/flow-fast` and `/flow-interactive` (operator, round 1 Q6: "New phase names"; Q7: "All three"); `/flow-interactive` marks plan, implement, verify and land per request and self-review and cleanup at its stop.
**Considered:** keeping the old key names for the phases (`flow.sdd-tdd`, `flow.review-panel`, `flow.landing-routes`) — names of steps, not phases; adopting them in `/flow` alone — three vocabularies in one chart.

### Retired keys live in one map

**ID:** retired-keys-map
**Status:** active
**Chosen:** retired keys leave `stages.Table` and `README.md`'s Level 1 table, so `Validate` refuses them for a new mark, and stored rows keep them; `stages.Retired` (retired key → the mark that replaced it) is what every Go reader of history uses — `aggregate.go`'s weekly-cost categories, `experimentstats.go`'s cohorts, `TestWeeklyCostStagesAreDeclared`'s exemption (`flow.design-approval` and `flow.plan-review` join the map under `flow.plan`); the SPA's `STAGE_PHASE` and `stageLabel.ts` keep every retired key mapped to its replacement's phase, and `phase.test.ts` checks the Table and the `Retired` keys both.
**Considered:** keeping the retired keys in the Table marked retired — `Validate` would keep accepting them for new marks (revises round 1's retired-keys-declared).

### The visual-verify cost category becomes verify

**ID:** weekly-verify-category
**Status:** active
**Chosen:** weekly-cost categories are planning, implementation, review, verify and self-review; verify covers `flow.verify` and the retired `flow.visual-verify`.
**Considered:** dropping the category — a week's verify cost would fall into overall-only.

### The interactive request gate reads one verify run

**ID:** request-gate-on-verify
**Status:** active
**Chosen:** `flow request`'s landing gate reads the request's latest `flow.verify` run, which, on a request that touched `ui paths`, must carry the UI check's `-screenshot`.
**Considered:** keeping a visual-verify mark for the gate — an eighth mark.

### Other pipelines keep their marks

**ID:** other-pipelines-keep-marks
**Status:** active
**Chosen:** `plan.session` (`/flow-plan`'s research session) and `flow.pipeline-fix` stay; `Flow-Key: flow.commit-two` and `Flow-Key: flow.visual-verify` are commit trailers, not marks, and stay.
**Considered:** folding them into the seven — they mark other pipelines, and the trailers are git history.

### The change-finished signal is a completed cleanup

**ID:** finished-signal-on-cleanup
**Status:** active
**Chosen:** `api/stages.go` and `hooks/flow-active-change.py` treat a completed `flow.cleanup` as the change-finished signal.
**Considered:** keeping `flow.write-finished` — a mark the seven-phase cut retires.

### Jira keeps two transitions

**ID:** jira-two-transitions
**Status:** active
**Chosen:** In Progress at start and Done at cleanup, for `/flow` and `/flow-fast` alike; the `JIRA: transition <KEY> to In Review` line and every In Review step go; the Go `jira` package keeps its four positions, which map a status someone reads and serve the follow-up join set.
**Considered:** dropping In Review from `/flow` only — the Jira contract is shared; trimming the `jira` package to two positions — the join set reads all four.

### /flow-fast keeps its own self-review order

**ID:** fast-self-review-unchanged
**Status:** active
**Chosen:** `/flow-fast`'s self-review order and its `## handoff: required` stop are unchanged; it drops the run-instructions stage only with the seven-mark cut.
**Considered:** moving its self-review too — outside the ticket's cut 3, which names integrate.

### One migration per retirement

**ID:** one-migration-per-retirement
**Status:** active
**Chosen:** each task that retires roles or drops a column adds its own numbered migration: `0076_drop_review_before_landing.sql`, `0077_retire_fix_roles.sql` (`fix-planner`, `fix-run-judge`, `late-fix`), `0078_retire_planning_roles.sql` (`plan-triage`, `design-reviewer`), `0079_retire_tooling_analyst.sql`.
**Considered:** one migration at the end — a task that removes a role's code would leave its rows behind until a later task.

### Seven marks are added before the old keys retire

**ID:** add-then-retire-keys
**Status:** active
**Chosen:** the new keys and `stages.Retired` land first with every reader; emitters and prose move next; the old keys leave the Table last, so every task leaves the tree green.
**Considered:** one rename task — every emitter, reader, test and skill in one commit.

## Open questions

## Recommended approach

Land the cuts in four slices, each green on its own. **Landing** goes first: `reviewBeforeLanding`
and the run instructions are deleted (drop-column-ignore-legacy-key, drop-in-progress-gate,
open-question-still-stops), self-review moves after `land` and gates `cleanup`
(land-stops-calling-cleanup, self-review-gates-cleanup, self-review-fixes-own-branch) with the open-PR
route's two runs kept (open-pr-keeps-two-runs), and In Review leaves Jira (jira-two-transitions).
**Stage marks** follow: the seven keys and `stages.Retired` are added with every reader, emitters and
prose move to them, and the old keys retire last (seven-phase-marks, retired-keys-map,
add-then-retire-keys). **Fix and planning** then rewrite the fix path as one procedure
(one-fix-procedure-file and its decisions) and the planning tail as one headless pass
(plan-check-merged-pass and its decisions), deleting the review-panel guards the slim paths no longer
call. **Visual verify** last: the merge-base capture and in-loop tooling analysis go, the visual and
isolation guards fold in-process, and the repeated checks run once.

Rejected alternatives are recorded per decision; the two shape-level ones are an await-merge watcher
(open-pr-keeps-two-runs) and moving pre-capture checks to `close` (guard-checks-in-process).

## Critical files

- `skills/flow/SKILL.md` — router: fix-run entry, the Review-before-landing section to delete.
- `skills/flow/fix.md` — new: the one fix procedure.
- `skills/flow/document-fix.md`, `skills/flow/verify-fix-loop.md` — deleted.
- `skills/flow/implement.md` — fix-run stage order, dispatch-site rows, dev-stack calls.
- `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md` — triage, plan review and Decide prose.
- `skills/flow/plan-check-prompt.md` — new: merged prompt; replaces `plan-reviewer-prompt.md`,
  `decider-prompt.md`; `plan-triage-prompt.md` deleted.
- `skills/flow/review-panel.md`, `review-panel-fix-round.md`, `review-panel-optional-slots.md` —
  fix-run review, docs-only branches, rendered-brief mandate, MUTATION PROOF.
- `skills/flow/gated-review-fix.md` — the in-run fix task exemption.
- `skills/flow/verify-and-handoff.md` — run instructions, review hold, dev-stack and spec-reach calls.
- `skills/flow/integrate.md`, `skills/flow/cleanup.md` — self-review placement, In Review step.
- `skills/flow/visual-verify.md`, `visual-verify-verifier.md`, `visual-verify-tooling-analysis.md`
  (moved) — merge-base capture, attribution, tooling analysis, guard citations.
- `skills/flow/stage-keys.md`, `skills/flow/SKILL-rationale.md` — marks and rationale.
- `skills/flow-visual-tooling-analysis/SKILL.md` — new standalone skill.
- `skills/flow-contracts/pipeline.md`, `pipeline-rationale.md`, `state-file.md`, `handoff-blocks.md`,
  `handoff-blocks-rationale.md`, `finish-contract-run1.md`, `finish-contract-run2.md`,
  `operator-prompts-auto-resolution.md`, `jira-integration.md`, `jira-integration-finish.md`,
  `jira-integration-rationale.md`, `SKILL.md`, `model-policy.md`, `git-boundaries.md`,
  `known-bugs.md`, `workspace-isolation.md`, `project-configuration.md`,
  `project-configuration-visual.md`, `project-configuration-isolation.md` — contracts.
- `skills/flow-fast/SKILL.md`, `skills/flow-interactive/SKILL.md`, `skills/flow-plan/SKILL.md`,
  `skills/flow-self-review/SKILL.md`, `skills/flow-status/SKILL.md` — sibling commands.
- `rules/flow-manual-review.mdc`, `rules/work-end-to-end.mdc` — review-before-landing mentions.
- `CLAUDE.md`, `README.md`, `.flow/project.md` — skill index, Level 1 table, `## lint`.
- `stats/internal/stages/names.go`, `names_test.go` — the vocabulary and `Retired`.
- `stats/internal/store/aggregate.go`, `experimentstats.go`, `records.go` (`autoStageFor`),
  `changes.go`, `migrations/0076`–`0079` — store.
- `stats/internal/api/changes.go`, `stages.go`, `selfreview.go`, `records.go` — API.
- `stats/cmd/flow/state.go`, `request.go`, `interactive.go`, `record.go`, `stage.go` — CLI.
- `stats/internal/guard/integratechange*.go` — prepare, commit, land, cleanup.
- `stats/internal/guard/plancheck.go` (new), `plantriage.go`, `decidegroupspanelvisual.go`,
  `pairjudge.go` — planning tail.
- `stats/internal/guard/fixslotjudge.go`, `paneldocsonly.go`, `renderslotprompt.go`,
  `mutateandverify.go`, `resolvebasebranch.go`, `shimduties.go`, `guard.go`, `loadsets.go` — cut guards.
- `stats/internal/guard/visualverifypass*.go`, `verifyreport.go`, `framelist.go`, `sidecarreach.go`,
  `verifygreen.go`, `visualpreflight.go`, `visualverification.go`, `workspaceisolation.go`,
  `prepareworkspace.go` — visual and isolation guards.
- `stats/internal/records/types.go`, `category.go`, `testdata/dispatch-keys.json` — guard and key vocabularies.
- `hooks/enforce-agent-baseline.py`, `hooks/flow-active-change.py`, `hooks/block-unsafe-commands.py`.
- `scripts/flow-guard` — exit-4 `case` for `mutate-and-verify`.
- `stats/web/src/lib/phase.ts`, `stageLabel.ts`, `interactive.ts`, `stats/web/src/views/StateBoard.tsx`,
  `RunDetail.tsx`, `Experiments.tsx`; `mods/flow-task-list/hooks/register.tsx` — SPA and mod.

## Existing code to reuse

- `flow-guard write-panel-diff late-fix` (`stats/internal/guard/writepaneldiff.go`) — the fix diff.
- `primary-reviewer-prompt.md` and `review-panel.md`'s placeholder table — the one reviewer's brief.
- `pjRun` (`stats/internal/guard/pairjudge.go`) — the headless runner `plan-check` calls.
- `decide-groups-panel-visual`'s checks and fallback (`decidegroupspanelvisual.go`) — moved into `plancheck.go`.
- `plan-review-rubric.md` — read by `plan-check` unchanged.
- `selfReviewCompleted` (`integratechange_land.go`) — moves to the cleanup gate.
- `flow-guard pipeline-fix land` — lands self-review fixes.
- `aggregate.go`'s retired-key precedent (`flow.design-approval`, `flow.plan-review`) — generalised as `stages.Retired`.
- `wiValidate`, `wiSpace`, `wiAbs` (`workspaceisolation.go`), `vvHeading`… (`visualsection.go`) — kept helpers.
- `postmutationcheck.go` — kept for `break-and-prove`.
- `vvp.guard` (`visualverifypass.go`) — replaced by direct calls to the same functions.

## Verification

- Every task's targeted `go test -run` and `vitest` names, and its lint (`gofmt -l`, `go vet` on the
  touched packages, `check-references`, `check-guard-calls`).
- Once all tasks land, this repository's whole `## lint` list (`.flow/project.md`) exits 0 —
  `check-guard-calls` proves no prose names a deleted or unregistered guard, `check-references`
  proves no citation of a deleted file or section.
- `cd stats && go test ./internal/stages/ ./internal/store/ ./internal/api/ ./cmd/flow/` and
  `cd stats/web && npx --no-install vitest run` pass; the guard package's full run is asked for
  first (operator rule), targeted `-run` otherwise.
- `grep -rn` over `skills/ rules/ hooks/ CLAUDE.md README.md .flow/` for every retired stage key,
  deleted guard, deleted role and `reviewBeforeLanding` returns only `stages.Retired`, the SPA maps,
  the migrations and rationale history.
- The Live check below.

## Non-goals

- The `/clear` handoff between planning and implementing (ticket item 1).
- The review panel's composition (ticket item 5); only the fix path stops re-running it.
- Same-run cleanup on the open-PR route (ticket cut 4, withdrawn in round 1).
- `/flow-fast`'s own self-review order and its `## handoff` stop.
- Editing another project's `## lint` that lists a removed guard.
- Rewriting historical docs under `docs/` that name retired keys or guards.

## Risks

- A one-reviewer fix path can miss what a full panel would catch on a large fix; the Critical or
  Important finding loop and the no-progress rule bound it.
- Dropping merge-base attribution makes every visual departure the change's to fix, including a
  pre-existing one — more in-run fixes on a screen that was already off.
- A headless plan pass reviews within a time limit; a run that hits it falls back to `not run` for
  findings and the fixed decision.
- Starting every change short relies on the short planner's own `large` estimate to escalate.
- A hand-filled slot brief loses `render-slot-prompt`'s checks: an unfilled placeholder, a relative
  path, a missing standards file or a mis-named mutation copy reaches the reviewer.
- Accepted: plan review F5 — hand-filled slot briefs lose `render-slot-prompt`'s checks; a
  `close --check-only`-style lint of the written brief would keep them — declined: it puts back a
  guard this change cuts, and a bad brief surfaces in the reviewer's own report.
- A spec visual verify captures is checked for reach only by the next verify run, not in the run
  that wrote it.
- A broken `## visual verification` or `## workspace isolation` section is found when a run reaches
  visual verify or workspace setup, no longer by this repository's lint.
- A mutation proof needing one mutation across several files or harnesses at once is now one
  `break-and-prove` call per file and harness, without `mutate-and-verify`'s blast-radius report.
- A project whose own `## lint` lists `flow-guard check-workspace-isolation` or `flow-guard
  check-visual-verification` fails its lint until that line is removed; this change can only fix
  this repository's.
- Retired stage keys must stay in `stages.Retired` and the SPA maps, or historical cost and phase
  charts lose rows.
- A change already at `IN_PROGRESS` on the old pipeline when this lands has run its self-review
  inside integrate; its run 2 finds a completed `flow.self-review` row and cleans up as before.

## Live check

After landing and `./setup.sh global && make -C stats restart`: run one `/flow` change through the
PR route and confirm run 1 ends after the self-review and the bare `/flow <name>` after the merge
cleans up; confirm the stats SPA's run detail, phase and weekly-cost charts still render a
pre-change run's retired stage keys under the new phases. Not working looks like a run detail with
unphased stage rows, or a weekly-cost week whose implementation or review total drops to zero.

