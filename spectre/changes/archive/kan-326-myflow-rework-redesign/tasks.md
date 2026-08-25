# Tasks

> **Execution:** `/myflow-do` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Design record: `design.md` in this change (decisions/open questions) and
`docs/superpowers/research/kan-326-myflow-rework.md` (full stage-by-stage audit). Open questions
`flow-file-split`, `write-in-progress-fold`, `research-staging-mechanism`, and
`stats-backend-untangle` are resolved by whichever task below they land in — see each task's notes.

Tasks 1-6 are the settings-store, state-schema, and review-panel groundwork; 7-8 are pure renames,
independent of everything else; 9-10 build the new `/flow` command surface and depend on 1-6;
11-12 are the stats UI cut, independent of 1-10; 13 wires cross-repo references after the rename
lands; 14 deletes the old command surface, last, once 9 replaces it; 15 is the rework-driven MD
pass, last of all. `/flow-fast` (a lighter, shorter-brainstorming variant) was designed and then
cancelled before implementation started — see `design.md`'s `flow-fast-cancelled` — so no task
here builds it.

- [x] 1. Settings store: schema and Go types in the stats app

  Add a `flow_settings` table (or equivalent store-backed record) holding the global defaults
  `/flow-settings` will manage: `models.default` (string, one of a fixed enum this task defines —
  the harness's known model identifiers) and `reviewers` (a list of reviewer-slot identifiers,
  validated against the fixed 3-required-slot vocabulary plus Bugbot/Security). Extend
  `stats/internal/store/` with the migration and the Go struct; extend `stats/internal/api/` with
  the validation function that rejects a value off either enum, returning the specific bad value in
  the error.

**Build:** green
**Files:** `stats/internal/store/settings.go`, `stats/internal/store/settings_test.go`, `stats/internal/store/migrations/0015_flow_settings.sql`, `stats/internal/api/settings.go`
**Allowed-collateral:** `stats/internal/store/migrations/*.sql`
**Tests:** `TestSettingsStore_RoundTrip`, `TestSettingsStore_RoundTripUpdates`, `TestSettingsStore_RejectsUnknownModel`, `TestSettingsStore_RejectsUnknownReviewer`
**Regression:** reverting this task removes the settings table and its validation; any test above fails to compile or fails immediately since the package it targets no longer exists.
**Baseline:** before=<N> after=<N+3> <!-- predicted: cd stats && make test @ 2da91eb -->
**Commit:** `feat(stats): add flow settings store with strict validation`

- [x] 2. Settings store: HTTP API endpoints

  Add `GET /api/v1/settings` and `PUT /api/v1/settings` to the stats daemon's router, backed by
  task 1's store methods. `PUT` returns 400 with the rejected value named on a validation failure,
  never a silent partial write.

**Build:** green
**Files:** `stats/internal/api/server.go`, `stats/internal/api/settings.go`, `stats/internal/api/settings_test.go`
**Allowed-collateral:** call-site updates to `api.New(...)`'s new store parameter — `stats/cmd/myflowd/main.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`, `stats/internal/api/changes_test.go`, `stats/internal/api/stages_test.go`, `stats/internal/api/stats_test.go`
**Tests:** `TestSettingsAPI_Get`, `TestSettingsAPI_Put_Valid`, `TestSettingsAPI_Put_RejectsInvalidValue`
**Regression:** reverting this task removes the two routes; the three tests above 404 or fail to route.
**Baseline:** before=<N> after=<N+3> <!-- predicted: cd stats && make test @ 2da91eb, after task 1 -->
**Commit:** `feat(stats): expose flow settings over HTTP`

- [x] 3. `myflow settings` CLI subcommand

  Add `settings get`/`settings set` to `stats/cmd/myflow/`, calling task 2's endpoints, following
  the existing `state`/`stage`/`record` subcommands' pattern (never-block-on-store-failure where
  that pattern already applies; here, unlike `state`/`stage`, a write that fails validation is a
  caller mistake and DOES surface as a non-zero exit and printed reason — there is no fallback
  value to "record" for an invalid model or reviewer name).

**Build:** red
**Squash-with:** Task 2
**Files:** `stats/cmd/myflow/settings.go`, `stats/cmd/myflow/settings_test.go`, `stats/internal/client/client.go`
**Allowed-collateral:** `stats/cmd/myflow/main.go` (dispatch/usage wiring for the new subcommand)
**Tests:** `TestSettingsCmd_Get`, `TestSettingsCmd_Set_PrintsRejectionReason`
**Regression:** reverting this task removes the CLI surface; `/flow-settings` (task 5) has nothing to call.
**Baseline:** before=<N> after=<N+2> <!-- predicted: cd stats && make test @ 2da91eb, after task 2 -->
**Commit:** `feat(myflow): add settings CLI subcommand`

- [x] 4. State file schema: collapse `models.*` to `models.default`, drop `reviewPanelRoster`

  Update the Go DTO that reads/writes a change's state file (`stats/internal/api/changes.go`,
  which carries `ReviewPanelRoster`
  <!-- measured: grep -n "ReviewPanelRoster" stats/internal/api/changes.go stats/internal/api/stats.go @ 2da91eb --> )
  to drop that field. Update `skills/myflow-contracts/state-file.md`'s JSON shape to match.
  `stats/internal/api/stats.go`'s own `ReviewPanelRoster` hit is a **different** concern —
  `panelEconomicsRowDTO`, backing the PanelEconomics aggregation view via `store/aggregate.go`'s
  `changes.review_panel_roster` query — and belongs to task 12 (removing the cut views' backend
  support), not this task; leave it untouched here. `models` was already an untyped
  `json.RawMessage` passthrough at every Go layer with no typed three-field struct to mechanically
  collapse — the "collapse to `models.default`" half of this task is a wire-shape convention
  documented in `state-file.md`'s JSON example and prose, not a Go struct change. This is a breaking
  schema change for any change still `IN_PROGRESS` under the old shape — the migration path
  (best-effort read of the old shape, or a hard cutover) is an open question for whoever merges this
  to resolve against how many changes are actually in flight at merge time.

**Build:** red
**Squash-with:** Task 1
**Files:** `stats/internal/api/changes.go`, `stats/internal/api/changes_test.go`, `skills/myflow-contracts/state-file.md`
**Tests:** `TestPutChangeRejectsReviewPanelRosterField`, `TestGetChangeOmitsReviewPanelRoster`
**Regression:** reverting this task restores `reviewPanelRoster` and the three `models.*` sub-fields; the updated tests fail to compile against the new struct shape.
**Baseline:** before=<N> after=<N> <!-- predicted: cd stats && make test @ 2da91eb — field rename, not a test-count change -->
**Commit:** `feat(stats): collapse model roles and drop review panel roster from state schema`

- [x] 5. `/flow-settings` skill and command

  New `skills/flow-settings/SKILL.md` (+ `commands/flow-settings.md`,
  `commands-claude/flow-settings.md`) that reads/writes task 2's API via task 3's CLI. Accepts no
  flags, per the existing "no command accepts a flag" pipeline rule (`pipeline.md:71`) — extended
  to this new command since it is part of the same command family. Prints the current global
  defaults and lets the operator change them via `AskUserQuestion`, one field at a time.

**Build:** green
**Files:** `skills/flow-settings/SKILL.md`, `commands/flow-settings.md`, `commands-claude/flow-settings.md`
**Allowed-collateral:** `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-stage-mark-calls.sh`, `scripts/check-contract-budget.sh`
**Tests:** manual — a skill/prompt file, not compiled code; verified by running the command and confirming a settings read-then-write round trip against a local daemon
**Regression:** reverting this task removes the command; the settings store from tasks 1-3 has no operator-facing way to be changed.
**Baseline:** before=<N> after=<N> — unverified:no automated test for a skill file
**Commit:** `feat(flow-settings): add settings command`

- [x] 6. Review panel: fix at 3 required slots, Bugbot/Security on-demand only

  Update the review-panel section (today `myflow-do/SKILL.md` **5. The review panel**, landing in
  whichever file task 9 places it in) to remove the `light`/`standard`/`full` roster table and its
  trigger-based optional-slot selection, replacing it with: 3 required slots dispatched
  unconditionally, Bugbot and Security dispatched only when named in an explicit operator
  instruction, checked for at the start of the panel stage and at every fix round. Depends on task
  4's schema change (no `reviewPanelRoster` field left to read).

**Build:** red
**Squash-with:** Task 9
**Files:** the review-panel section of task 9's new `/flow` skill file(s)
**Tests:** none automated — this is a prompt-text change; verified by running `/flow` on a small change and confirming exactly 3 slots dispatch with no operator instruction, and a 4th dispatches only when asked
**Regression:** reverting this task restores roster-based selection.
**Baseline:** before=<N> after=<N> — unverified:no automated test for panel prompt text
**Commit:** folded into task 9's commit

- [x] 7. Rename `myflow-status` → `flow-status`

  `git mv skills/myflow-status skills/flow-status`, update its `SKILL.md`'s own name/description
  frontmatter, `commands/myflow-status.md` → `commands/flow-status.md`,
  `commands-claude/myflow-status.md` → `commands-claude/flow-status.md`. Independent of every other
  task — no behavior change, pure rename.

**Build:** green
**Files:** `skills/flow-status/SKILL.md`, `skills/flow-status/scripts/resolve-base-branch.sh`, `commands/flow-status.md`, `commands-claude/flow-status.md`
**Tests:** none automated — verified by running the reference-check guard by hand, not a named test
**Regression:** reverting this task restores the old name; `check-references.sh` would then flag any place task 13 updated to point at the new name.
**Baseline:** before=<N> after=<N> — unverified:rename, not a test-count change
**Commit:** `refactor(skills): rename myflow-status to flow-status`

- [x] 8. Rename `myflow-research` → `flow-research`, add staging-write behavior and deeper research

  `git mv skills/myflow-research skills/flow-research` (+ command file renames as in task 7). Add
  the staging-write behavior decided in `design.md`'s `flow-research-staging`: on a research session
  the operator wants captured, write to `docs/superpowers/research/<topic-slug>.md` (this change's
  own note is the worked example) instead of only offering to update an existing change's
  `design.md`. Resolves open question `research-staging-mechanism`'s *write* half — the *discovery*
  half (how `/flow`'s brainstorming finds a staging file for a given topic) is task 10. Also
  implements `design.md`'s `flow-research-depth`: more codebase investigation before treating a
  topic as understood, and more clarifying questions before writing a staging note — design the
  concrete stopping rule (round count, a convergence test, or an explicit offer) as part of this
  task, per that decision's own note that the mechanism was left open. Also adds the step-by-step
  analysis mode `flow-research-depth` names, as the **default** output shape: a structured
  name/what/why/uses breakdown of the topic (this change's own
  `docs/superpowers/research/kan-326-myflow-rework.md`'s stage-audit section is the worked example)
  rather than prose discussion alone, on every research session, not only when asked. Define and
  implement the fixed section structure `design.md`'s `flow-research-artifact-structure` requires —
  source line, one section per topic/thread, and an always-present step-by-step breakdown section —
  and write every staging note through it, never as free-form prose.

**Build:** green
**Files:** `skills/flow-research/SKILL.md`, `skills/myflow-research/SKILL.md`, `commands/flow-research.md`, `commands/myflow-research.md`, `commands-claude/flow-research.md`, `commands-claude/myflow-research.md`
**Tests:** manual — run `/flow-research` on a throwaway topic, confirm the staging file lands at the expected path with no spectre change created, confirm the session visibly investigates more and asks more questions than a single-pass answer would, and confirm a step-by-step breakdown appears by default with no explicit request needed
**Regression:** reverting this task restores the old name and drops both the staging-write behavior and the deeper-research behavior; `/flow-research` reverts to today's `myflow-research` depth and discussion-only output.
**Baseline:** before=<N> after=<N> — unverified:no automated test for a skill file
**Commit:** `feat(flow-research): rename, add staging-note output, and go deeper`

- [x] 9. Build `/flow`: the single command replacing start/do/finish/fast

  The core of this change. Retire `skills/myflow-start`, `skills/myflow-do`, `skills/myflow-finish`
  as user-facing skills; build `skills/flow/` (file split organized by topic — resolves open
  question `flow-file-split`; propose the concrete split against the actual cited-content volume at
  implementation time, per `design.md`'s note that the split itself was left open) carrying:
  `STARTED` redefined as a kickoff marker written at invocation (`started-redefined`); no
  `ask-options` stage; no `publish-proposal` stage; task 6's fixed 3-slot review panel;
  `do.run-instructions` moved immediately before `do.write-in-progress`
  (`run-instructions-reorder`); `finish.move-in-review` folded into `finish.landing-routes`
  (`move-in-review-fold`); `do.workspace-export`/`do.lint-and-test` merged into one verify stage
  (`workspace-export-lint-merge`); model reads from task 1-4's settings store instead of the deleted
  per-change `models.*` fields, with plain-language session override recorded per-run
  (`settings-scope`). Resolve open question `write-in-progress-fold` here, one way or the other, and
  say which in the task's own commit body. `commands/flow.md` and `commands-claude/flow.md` replace
  `commands/myflow-fast.md` and its claude counterpart.

**Build:** green
**Files:** `skills/flow/SKILL.md`, `skills/flow/brainstorm.md`, `skills/flow/implement.md`, `skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md`, `skills/flow/integrate.md`, `skills/flow/archive.md`, `skills/flow/engineering-principles.md`, `skills/flow/principles-reviewer-prompt.md`, `skills/flow/security-reviewer-prompt.md`, `commands/flow.md`, `commands-claude/flow.md`
**Allowed-collateral:** `skills/flow/scripts/*` (guard-script copies/symlinks mirroring `skills/myflow-fast/scripts/`, not this task's own substance)
**Tests:** none automated — verified by hand-running the existing skill-file guard suite (markdown integrity, normative inventory, vocabulary, stage-mark calls) against the new files
**Regression:** reverting this task leaves no working single-command pipeline; the old `myflow-start`/`-do`/`-finish`/`-fast` split (task 13 has not yet deleted it, since this task precedes it) remains the only working path.
**Baseline:** before=<N> after=<N> — unverified:guard scripts report pass/fail, not a count
**Commit:** `feat(flow): build single-command pipeline replacing start/do/finish/fast`

- [x] 10. `/flow` brainstorming: discover and seed from staged research notes

  Resolves the remaining half of open question `research-staging-mechanism`. In `skills/flow/`'s
  brainstorming section, before starting the interactive checklist, check
  `docs/superpowers/research/` for a note matching the change's resolved name or Jira key; if found,
  present its contents as the starting point and proceed through the full interactive round exactly
  as `design.md`'s `flow-research-staging` decision requires — never skip straight to artifact
  writing. Parse the note against task 8's fixed section structure
  (`flow-research-artifact-structure`) rather than treating it as loose prose — a note missing a
  required section is reported, not silently treated as empty. Decide and document what happens to
  the staging file once its content is adopted (left in place cross-referenced, or removed) as part
  of this task, since `design.md` left that specific sub-question open.

**Build:** red
**Squash-with:** Task 9
**Files:** `skills/flow/brainstorm.md`'s "Seed from a staged research note" section (task 9 split the brainstorming stage into its own file rather than keeping it in `SKILL.md`)
**Tests:** manual — run `/flow` on a topic with an existing staged note (this change's own `docs/superpowers/research/kan-326-myflow-rework.md` is the fixture), confirm it seeds the interactive round rather than being ignored or auto-accepted
**Regression:** reverting this task returns `/flow`'s brainstorming to blank-slate on every run, ignoring any staged research note.
**Baseline:** before=<N> after=<N> — unverified:no automated test for a skill file
**Commit:** folded into task 9's commit

- [x] 11. Stats UI: remove the four cut views

  Delete `stats/web/src/views/{CostPerChange,ModelComparison,PanelEconomics,ReworkRate}.tsx` and
  their route/nav wiring; keep `StateBoard.tsx`, `Trend.tsx`, `CacheEfficiency.tsx`, `RunDetail.tsx`
  per `design.md`'s `stats-ui-cut`. Update `stats/web/src/views/views.test.tsx` to drop the removed
  views' cases.

**Build:** green
**Files:** `stats/web/src/views/CostPerChange.tsx` (deleted), `stats/web/src/views/ModelComparison.tsx` (deleted), `stats/web/src/views/PanelEconomics.tsx` (deleted), `stats/web/src/views/ReworkRate.tsx` (deleted), `stats/web/src/views/views.test.tsx`, `stats/web/src/App.tsx`, `stats/web/src/App.test.tsx`, `stats/web/src/api.ts`, `stats/web/src/views/RunDetail.test.tsx`, `stats/web/src/views/Trend.tsx`, `stats/web/src/views/CacheEfficiency.tsx`
**Allowed-collateral:** `RunDetail` (a kept view) depends on the `cost-per-change` backend endpoint for its header aggregate, so `api.ts` narrows the navigable `ViewName` enum to the 4 surviving views while still widening `fetchStatsView`/`StatsResponse` to accept the non-navigable `cost-per-change` slug — not a full removal of that endpoint's client-side plumbing
**Tests:** `stats/web/src/views/views.test.tsx` (updated)
**Regression:** reverting this task restores the four views and their nav entries; the updated test suite fails to find the removed cases' expectations.
**Baseline:** before=<N> after=<N-4-or-more> — unverified:exact current test count not read
**Commit:** `refactor(web): remove cost-per-change, model-comparison, panel-economics, and rework-rate views`

- [x] 12. Stats UI: untangle and remove backend support for the cut views

  Resolves open question `stats-backend-untangle`. Per-symbol review of
  `stats/internal/api/stats.go`, `stats/internal/store/aggregate.go`, `stats/internal/store/pricing.go`,
  `stats/internal/store/stageruns.go`, `stats/internal/sweep/sweep.go` (the five files
  `design.md`'s `stats-backend-untangle` measured touching the four cut views' names) to separate
  what's exclusive to the cut views from what `StateBoard`/`Trend`/`CacheEfficiency`/`RunDetail`
  also depend on. Remove only the exclusive code; leave shared code and its tests untouched. **Task
  11's implementer found `RunDetail` (a kept view) calls the `cost-per-change` backend endpoint
  directly for its header aggregate** — the endpoint/query itself must stay; only whatever backs
  the deleted `CostPerChange` view's own repo-breakdown/full-page behavior, not shared with
  `RunDetail`'s narrower use, is this task's to remove.

**Build:** green
**Files:** `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/api/changes_test.go`, `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/store/harvestshape_test.go`, `stats/internal/sweep/sweep.go`, `stats/internal/client/client_test.go`
**Allowed-collateral:** `stats/internal/store/stageruns.go`, `stats/internal/store/pricing.go` (doc-comment-only touches, no removable code found in either)
**Tests:** none named — full suite is the right bar for this task, not a subset, since the whole risk is breaking a surviving view's data path; verified by hand-running the project's configured Go test command before and after
**Regression:** reverting this task restores the removed backend code; `make test` continues to pass either way since nothing surviving depended on what's removed, by this task's own construction — a test failure post-removal is this task's own defect to fix, not a sign the revert is needed.
**Baseline:** before=<N> after=<N> — unverified:exact current Go test count not read; this task removes production code, not tests, so the count should be unchanged unless a removed view had its own now-orphaned test file
**Commit:** `refactor(stats): remove backend support for the removed dashboard views`

- [x] 13. Update cross-repo references to the renamed/removed commands

  `README.md`'s Level 1/2 stage tables and command map, `setup.sh`'s install lists,
  `scripts/check-references.sh`'s allow-list, `rules/myflow-manual-review.mdc`'s copied three-line
  pipeline digest, and any other file `check-references.sh` or `check-installed-citations.sh` flags
  after tasks 7-10 land. This task is what makes those two guards pass clean again — run them first
  to get the actual list rather than guessing it here. **Also registers the new `flow.*` stage-key
  namespace task 9 minted** — task 9's own implementer found and disclosed that
  `stats/internal/stages/names.go` doesn't recognize any `flow.*` key yet, so every
  `myflow stage begin -stage flow.*` call is currently rejected by the CLI as undocumented; without
  this, `/flow` cannot actually run end to end. Add every `flow.*` key `skills/flow/` uses (read
  them off `skills/flow/SKILL.md`'s own stage-key table) to `names.go`, and add `/flow`'s row to
  `pipeline.md`'s state-transition table and `README.md`'s Level 1 stages table (currently only
  `/myflow-fast`'s row exists there for the composite command). **Also update
  `scripts/check-stage-mark-calls.sh`'s `EXPECTED_ZERO_FILES` list** — task 9's own review found it
  doesn't yet contemplate `skills/flow/SKILL.md` (a legitimate zero-mark router file whose marks
  live in the phase files it cites) and still carries 4 pre-existing violations from tasks 7/8's
  renames; this task is what closes all of that out.

**Build:** red
**Squash-with:** Task 9
**Files:** `README.md`, `setup.sh`, `scripts/check-references.sh`, `scripts/check-stage-mark-calls.sh`, `rules/myflow-manual-review.mdc`, `skills/myflow-contracts/pipeline.md`, `stats/internal/stages/names.go`, `stats/internal/stages/names_test.go`, plus whatever `scripts/check-references.sh`/`scripts/check-installed-citations.sh` name when run — unverified:exact file list depends on tasks 7-10's actual output
**Tests:** none named — verified by hand-running the reference/citation/rules guards and the stats-app stage-name test package
**Regression:** reverting this task leaves stale references to deleted command names; the three guards above fail.
**Baseline:** before=<N> after=<N> — unverified:guard scripts report pass/fail, not a count
**Commit:** folded into task 9's commit

- [x] 14. Delete the old command surface

  `git rm -r skills/myflow-start skills/myflow-do skills/myflow-finish skills/myflow-fast` and
  their `commands/`/`commands-claude/` counterparts, now that task 9 fully replaces them and task
  13 has repointed every reference. Last task before the MD cleanup pass — nothing after this
  depends on the old files existing.

**Build:** green
**Files:** `commands/myflow-start.md`, `commands/myflow-do.md`, `commands/myflow-finish.md`, `commands/myflow-fast.md`, `commands-claude/myflow-start.md`, `commands-claude/myflow-do.md`, `commands-claude/myflow-finish.md`, `commands-claude/myflow-fast.md`, `README.md`, `skills/myflow-contracts/pipeline.md`, `stats/internal/stages/names.go`, `stats/internal/stages/names_test.go` (all deletions/edits directly needed to remove the old command surface); `skills/myflow-start/*` (deleted), `skills/myflow-do/*` (deleted), `skills/myflow-finish/*` (deleted), `skills/myflow-fast/*` (deleted)
**Allowed-collateral:** `skills/myflow-start/**`, `skills/myflow-do/**`, `skills/myflow-finish/**`, `skills/myflow-fast/**` (every file under the four deleted skill directories — SKILL.md, SKILL-rationale.md, reviewer prompts, engineering-principles.md, and each directory's own `scripts/`)
**Tests:** none named — verified by hand-running the reference-check guard (confirms nothing still points at the deleted paths) and the stats-app stage-name test package
**Regression:** reverting this task restores the old skill directories; since task 9 already replaced their function, the revert leaves two working copies of the same logic rather than breaking anything — the guard in **Tests** above is what would need re-running to confirm the restored old references are consistent again.
**Baseline:** before=<N> after=<N> — unverified:deletion, not a test-count change
**Commit:** `refactor(skills): remove myflow-start, myflow-do, myflow-finish, myflow-fast`

- [x] 15. MD cleanup pass (rework-driven only — general pass is out of scope)

  Per `design.md`'s `md-cleanup-sequenced`: sweep the files touched by tasks 1-14 for prose that is
  now dead — references to roster presets, `ask-options`, `publish-proposal`, or the old
  start/do/finish split that survived the renames/deletions above as descriptive text rather than
  as a citation target. This is scoped to what this change itself obsoletes; a general
  verbosity/redundancy pass over files this change did not otherwise touch is explicitly out of
  scope, per the decision.

**Build:** green
**Files:** `CLAUDE.md`, `AGENTS.md`, `README.md`, `skills/README.md`, `skills/myflow-contracts/pipeline.md`, `skills/myflow-contracts/state-file.md`, `rules/myflow-manual-review.mdc`, `rules/commit-scope-is-the-module.mdc`, `.myflow/project.md`, `commands/flow.md`, `commands-claude/flow.md`, `commands/flow-status.md`, `commands-claude/flow-status.md`, `commands/flow-research.md`, `commands-claude/flow-research.md`, `skills/flow-settings/SKILL.md`, `skills/flow/SKILL.md`, `skills/flow/archive.md`, `skills/flow/brainstorm.md`, `skills/flow/implement.md`, `skills/flow/integrate.md`, `skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md`, `scripts/check-contract-budget.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-installed-citations.sh`, `scripts/check-references.sh`
**Allowed-collateral:** `stats/cmd/myflow/stage_test.go`, `stats/internal/api/stages_test.go`, `stats/internal/reconcile/stage_test.go`, `stats/internal/web/embed_test.go` (final-panel fix round, folded into this commit: stale `/myflow-do`/`do.sdd-tdd` test literals and an orphaned reference to types task 12 deleted); `scripts/test-setup.sh`, `scripts/test-check-references.sh`, `scripts/test-check-stage-mark-calls.sh` (do.lint-and-test gate found these three guard test suites still hardcoded old `myflow-finish`/`myflow-research`/`myflow-status` paths after tasks 7/8/14's renames/deletion — fixed as part of getting the project's own mandatory test suite green before handoff)
**Tests:** none named — verified by hand-running the full project lint and test suite end to end: all 33 guard test suites, the full Go race-detector suite, gofmt, go vet, the TypeScript build, and the SPA test suite
**Regression:** reverting this task leaves dead prose in place; the guards above still pass either way, since dead prose is a readability defect, not a structural one — this task's value is not mechanically checkable, which is exactly why `design.md` scoped it to "rework-driven only" rather than promising a checkable completeness bar.
**Baseline:** before=<N> after=<N> — unverified:prose cleanup, not a test-count change
**Commit:** `docs(flow): remove prose obsoleted by the rework`
