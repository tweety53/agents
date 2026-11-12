# kan-1107-flow-verify-stop-full-suite-fix-reverify-loop
> Design record of change kan-1107-flow-verify-stop-full-suite-fix-reverify-loop (KAN-1107), written at landing. Each decision's own **Status:** line is the status.

## Context

No prototype: nothing user-visible changes.

The ticket (KAN-1107) names nine items across the agents repository and two gymie repositories.
The operator chose one multi-repo change over splitting the gymie half off, so items 2 and 7 are in
scope here, with worktrees in gymie and gymie-playwright (section H). A later round grew the
change into a restructuring of the Playwright suite (one file per feature area, empty and
maximal-fill smoke captures, gap captures); at about 50 tasks it was split. This change keeps the
ticket's items 1–9 plus the selection mechanics that make the verify tail cheap (K1, I, L) and an
interim three-test smoke set. What moved out is listed once:

- **KAN-1109** (already filed):
  - K4 one-file-per-area merges, with their snapshot moves, sidecar concatenation, `OFFLINE_SPECS`,
    `test:e2e:routes` and `test:visual*` edits, and stale-mention prose;
  - the K4 forward-rule prose: agents `project-configuration-visual.md` and
    `visual-verify-verifier.md` step 8, and gymie `.flow/project.md` l.341-343;
  - the `scripts/spec-names.test.mjs` guard;
  - the 12 empty + maximal-fill smoke tests and their seeding;
  - the `README.md` `## Feature areas` table;
  - K3's workers measurement and the smoke wall-clock row;
  - the agents compose narrowing in `visualverifypass_measure.go`.
- **KAN-1108** (already filed): every gap capture.
- No further Jira issue is filed by this change.

Evidence base: gymie-playwright `7f3b0d2` (Playwright 1.62.1), agents `a5e73bab`. The three smoke
lines and the `package.json` scripts were re-checked at `7f3b0d2`. Not known: the failing test
names of KAN-1086's rounds 1–8 (log overwritten); verify-10 failed in
`fidelity/kan-754-offline-l3.spec.ts:380` and `fidelity/templates.spec.ts:97`.

## Decisions

### Narrowed re-runs through run-verify-list flags

**ID:** narrow-via-run-verify-list-flags
**Status:** active
**Chosen:** two `run-verify-list` flags, `-only <file>` and `-detect-flaky`, substituted into the `verify` cell's existing `<specs>` placeholder — no project config change, and the guard tests can stub a `verify` cell that echoes its argv.
**Considered:** a new per-project `rerun` cell — restates the ~600-char stack bring-up of `verify`, the two drift apart, and the parser is still needed; contract-only, the parent running the narrowed command by hand — the parent rebuilds the cd/ports/env prefix, skips the stack-freshness check and the logs, and no test reaches it.

### One multi-repo change

**ID:** multi-repo-change
**Status:** active
**Chosen:** gymie and gymie-playwright get worktrees on this change; items 2 and 7 are in scope (operator's answer, not the recommended option).
**Considered:** split the gymie half into a follow-up gymie change — the recommendation, declined by the operator; drop items 2 and 7 with no follow-up.

### Split the suite restructuring into KAN-1109

**ID:** scope-split-kan-1109
**Status:** active
**Chosen:** this change carries items 1–9, K1, I, L's selection mechanics and full trigger, and an interim `@smoke` set of three core flows; KAN-1109 carries the rest (the list under `## Context`). The verify-cost fix — the ticket's purpose — lands without waiting on 75 spec moves; the spec-name guard must ship with the merges because it fails on today's `kan-*` files. KAN-1109 inherits the round-4 answers: one merged file per area (operator's choice over the recommended per-area file set, which kept every file renamed with `git mv`), per-ticket functional specs treated the same and `test:e2e:routes` deleted, a multi-area file going whole to its dominant area, the spec-name test plus contract sentences, and workers 5/4/3 ×2 with no sharding; and the round-5 answers on two offline files, journeys staying in `tests-full/`, composing only new or modified PNGs, the 12 area smoke tests, per-test API seeding on a pinned account, the kept group-floor mask and initials avatars.
**Considered:** one change of about 50 tasks — honours the earlier "in this change" answers literally, but puts one review panel over a three-repo diff with about 550 moved PNGs and about 25 new captures, and a red capture check holds the loop fix back; <!-- predicted: round-5 estimate from the K4 merge list and the gap-capture count, confirmed by KAN-1109's own plan --> three changes (loop fix / merges and smoke / gap captures) — the merges and the maximal-fill captures touch the same 19 files back to back, so separating them adds a rebase without saving review.

### Gap captures go to KAN-1108

**ID:** gap-captures-kan-1108
**Status:** active
**Chosen:** no gap capture in this change or in KAN-1109; KAN-1108, in the KAN-1082 epic, does them once the Gymie V2 design system is fully implemented (operator override).
**Considered:** capture every gap in this change (operator's earlier answer) — 66 of the 79 frames cannot be captured as drawn; capture 13 and list 66 with reasons (round-5 recommendation); list per area and file one follow-up per area (round-4 recommendation).

### Selection diffs against the change's own base

**ID:** selection-base-kickoff-commit
**Status:** active
**Chosen:** the commit `flow.kickoff` branched each worktree from, read from the state record — files merged in from another unmerged branch count only when actually merged into this branch.
**Considered:** `merge-base HEAD origin/<base>` after a fresh fetch — develop may still not contain KAN-1084, so it fails the same way; leave the base and fix only the gymie-side selector.

### "Untouched spec" is any spec outside the specs row's output

**ID:** untouched-spec-outside-specs-output
**Status:** active
**Chosen:** any spec not in the `specs` row's output for this change — reuses the selector the run already ran; a smoke failure therefore counts as "in the change's specs".
**Considered:** any spec whose own file and imported page objects the diff does not modify — needs import tracing in `run-verify-list`; any spec that also fails on the base branch — costs another suite run.

### The detector covers the plain run's selection

**ID:** detector-covers-selection
**Status:** active
**Chosen:** `-detect-flaky` runs the same specs as the failing plain run (the `specs` row's output, or nothing) — matches the ticket's "run once" and catches every flake in that set.
**Considered:** only the failing specs with `--repeat-each 5` — blind to flakes in specs that happened to pass; the whole suite always — about 18 min.

### verify-fix-loop's keys at flow.verify

**ID:** verify-fix-keys-in-run-fix
**Status:** active
**Chosen:** `In-run fix <k>` task and `-fix-<k>` keys at flow.verify; `full-suite-fix-<n>` is implement-phase only; `fix_rounds` counts both.
**Considered:** `full-suite-fix-<n>` / `verify-<k>` everywhere — matches what the KAN-1086 parent did, but loses the In-run fix task trail that `unfinishedwork`, `planshape` and `closetask` guard; keep both and only state which applies where.

### Readers prefer the served model and effort

**ID:** readers-prefer-served-metrics
**Status:** active
**Chosen:** readers prefer `metrics.modelVersion` / `metrics.effort`, the label columns stay, and the ledger shows `label → served` when they differ.
**Considered:** harvest overwrites the `model` / `effort` columns on inline rows — label-vs-served drift can no longer be seen; overwrite the columns and move the label into `metrics.label`.

### Per-round artifacts by round counter

**ID:** per-round-artifacts-round-counter
**Status:** active
**Chosen:** each failing run writes `visual-<key>-r<N>.log` plus `visual-results-r<N>/`, kept until the worktree is removed.
**Considered:** keyed by HEAD sha — two runs at one sha (detector and plain) collide; keep only the last three rounds — loses round 1's evidence on a long loop.

### Dispatch cost on the row's own metrics

**ID:** dispatch-cost-on-row-metrics
**Status:** active
**Chosen:** a new `PriceDispatch` sets the dispatch row's own `metrics.cost_usd` in the existing pricing pass; readers show it only as a breakdown. Covers inline and fixer rows.
**Considered:** readers join the stage-run `dispatches.<agentId>` bucket for every role — inline rows have no agent id and stay unpriced; both, falling back to the bucket when the row has no tokens.

### affected-tests becomes a Gradle task

**ID:** affected-tests-gradle-task
**Status:** active
**Chosen:** a Gradle `affectedTests` task whose mapping is a pure Kotlin function in `buildSrc` with JUnit tests; `## test` becomes `./gradlew affectedTests` — the bash script is ported before it is extended, per "No scripts in bash", and gymie declares no exception.
**Considered:** a Kotlin script run by the `kotlin` CLI — adds a toolchain requirement gymie does not have; declare `affected-tests.sh` a bash exception — only the operator can grant it, and every later extension inherits it.

### Build-logic path routing

**ID:** build-logic-routing
**Status:** active
**Chosen:** `buildSrc/**`, `gradle/*.gradle.kts` and `gradle/wrapper/**` route to buildSrc's own tests only; `gradle/deploy.gradle.kts` also runs `:app:test` (declared an `:app` test input in `src/app/build.gradle.kts` l.103); the version catalog, root `build.gradle.kts` and `settings.gradle.kts` stay full — a catalog bump can break any module.
**Considered:** the ticket's literal wording, all of `buildSrc/` and `gradle/` to buildSrc's tests — a dependency bump then runs no module tests; only `buildSrc/` routed, all of `gradle/` full.

### Seeder tests behind a tag and their own task

**ID:** seeder-tag-and-task
**Status:** active
**Chosen:** `@Tag("seeder")`, excluded from `:app:test` and run by a new `:app:seederTest`, added by `affectedTests` on seeder or migration paths and every full fallback. Known miss: a domain module a seeder calls does not run the seeders until a full run.
**Considered:** exclude only under `-PskipSeeders` — the filter depends on a flag every caller must remember; also run seeders on any `src/app/*` domain change — fewer misses, saves less time.

### Testcontainers: no build change

**ID:** testcontainers-no-change
**Status:** active
**Chosen:** record that `:app` already uses the documented singleton pattern (`AppIntegrationTest.kt` l.63-79); the remaining per-class containers are in small modules that are not the measured hotspot, and the suite records measure them first.
**Considered:** move the per-class `@Container` fields in auth, cardio, groupsession and invite to one singleton per module — about 47 test files, and tests relying on a fresh database per class need an explicit truncate; `withReuse(true)` — parallel forks and worktrees share one database, and a container outlives the suite.

### Suite records

**ID:** suite-records-scope
**Status:** active
**Chosen:** gymie `## test` gets `# suite: gradle-affected`; `run-verify-list` records the visual `verify` row as suite `visual` on plain runs only.
**Considered:** only the gymie line — the ~18-minute Playwright tail stays unrecorded, since `## visual verification` cannot carry a `# suite:` line; also record narrowed and detector runs under `visual-only` and `visual-detect`.

### Non-UI callers followed transitively

**ID:** non-ui-callers-transitive
**Status:** active
**Chosen:** a touched non-UI Kotlin file is followed through UI and non-UI callers until each branch reaches a tagged UI file or runs out — repository → view model → screen is selected.
**Considered:** one hop only — misses that chain; transitive but `commonMain` only — platform source sets stay unselected.

### One sibling-frontend helper

**ID:** shared-sibling-frontend-helper
**Status:** active
**Chosen:** extract `visual-verify-pass prepare`'s sibling rule into one helper used there and by `run-verify-list`.
**Considered:** leave flow.verify scoped to the canonical worktree because `flow.visual-verify` covers the sibling — flow.verify then never runs a spec for a frontend-only diff.

### affectedTests keeps its origin/develop base

**ID:** affected-tests-base-origin-develop
**Status:** active
**Chosen:** `merge-base HEAD origin/develop`, as today — `run-verify-list` runs fence lines literally and the kickoff base lives only in flow's state record; over-selection here costs extra module test runs, not a whole Playwright suite.
**Considered:** the kickoff base through `flow state` — ties gymie's build to the `flow` CLI on PATH and the daemon being up.

### gymie-playwright declares its own test command

**ID:** gymie-playwright-own-test
**Status:** active
**Chosen:** `## test` with `npm run test:scripts` in gymie-playwright's `.flow/project.md`, its "every other key is that project's" sentence naming `## test` as its own.
**Considered:** rely on auto-detect — `package.json` has no `test` script, so it may run nothing or the wrong one.

### Merge order

**ID:** merge-order
**Status:** active
**Chosen:** `.`, `gymie-playwright`, `gymie` — the script lands before the gymie prose that describes it, and since run 1 stops at the first failure, the agents change lands even if a gymie gate fails.
**Considered:** `.`, `gymie`, `gymie-playwright`; `gymie-playwright`, `gymie`, `.` — the gymie side lands before the agents guard that uses it.

### Per-test selection in both projects

**ID:** per-test-selection
**Status:** active
**Chosen:** `specs-for-diff` emits `path:line` per test block reaching a touched tag, `path:<describe line>` for describe-level reach, and the bare path for file-level reach, in both Playwright projects.
**Considered:** per spec as today — a touch reaching one test of the 22-test `full-app-suite` still runs all 22; per test in the fidelity project only — two output shapes for one selector.

### Full suite on demand only

**ID:** full-suite-on-demand
**Status:** active
**Chosen:** the whole suite runs only on demand, `npx playwright test` as `README.md` documents; no scheduler, and no `test:full` alias because that command already is the on-demand run.
**Considered:** a whole-suite run after landing in `/flow`'s run 2 — about 18 min per landed gymie change; a nightly `launchd` job — machine config outside every repo that nothing reads; a `test:full` script — adds nothing over `npx playwright test`.

### Smoke tier by Playwright tag, emitted as tokens

**ID:** smoke-tag-tokens
**Status:** active
**Chosen:** `{ tag: '@smoke' }` on the tests themselves, `npm run test:smoke` = `--grep @smoke`, and `specs-for-diff` emits the smoke tests as `path:line` tokens beside the selection — titles and screenshot names do not change, and `--grep`'s AND with path filters is avoided. Fails under a non-literal tag, which the parser cannot read; the tagged-set test then fails.
**Considered:** a `smoke` Playwright project with a `testMatch` file list — projects select whole files; a `smoke/` directory of separate specs — a copy that drifts.

### Interim smoke set: three core flows

**ID:** interim-smoke-three-flows
**Status:** active
**Chosen:** `tests/auth.spec.ts` l.132 (sign-in), `tests/workout-editor.spec.ts` l.5 (editor persists), `tests/session-player.spec.ts` l.15 (start → log → evaluate → finish), tagged in place; KAN-1109's verbatim moves carry the tag and only widen the tagged-set test.
**Considered:** five whole files, 16 tests (`auth`, `visual-baseline`, `visual-baseline-social`, `session-player`, `workout-editor`) — the operator's round-4 pick, replaced by the two-tier design; the 12 area empty + maximal-fill tests plus the flows — moved to KAN-1109 with the split. <!-- predicted: answers-4.md Q7 and round-5 counts, confirmed by `npx --no-install playwright test --list` on those files -->

### Full trigger on shared rendering code

**ID:** full-trigger-shared-rendering
**Status:** active
**Chosen:** a diff touching shared rendering code (the globs in section L) prints every spec; then the first plain run and every closing gate are full, and `-only` rounds stay narrowed. Otherwise touched + smoke. Theme files already select 92/100 specs, and this closes the 0-spec blind spots (platform, wasm, resources, `AuthenticatedShellContent`).
**Considered:** full only at the closing gate — a shared regression is found after the fix rounds; plus one full run at every closing gate — about +18 min per gymie change, the cost this ticket removes; never automatic — theme 92/100, platform 0.

### Smoke in plain runs only

**ID:** smoke-in-plain-runs-only
**Status:** active
**Chosen:** smoke joins every plain run; `-only` stays failing-only — `-only` runs `--repeat-each 5 --workers 1`, so smoke there multiplies serial time.
**Considered:** `-only` appends the smoke tail, repeated 5× at 1 worker; a second per-round smoke invocation at normal workers — two commands per `-only` round.

### A README task for the agents repo

**ID:** agents-readme-task
**Status:** active
**Chosen:** one task beyond round 6's 27 puts the new `run-verify-list` flags and the verify loop's narrowing in the agents `README.md` Guardrails list, because the change alters a guard's flags and the pipeline and no other task edits that README (**C. Create the change and its artifacts**, `skills/spectre-planning/SKILL.md`).
**Considered:** no README edit — the README rule requires one.

## Open questions

## Recommended approach

The design below is the approved round 6, by section. Each rejected alternative's reason is in its
decision under `## Decisions`.

### A. Narrowed re-runs and the flake detector (items 1, 3, 4)

- **Parsing the failures.** `run-verify-list` (`stats/internal/guard/runverifylist.go`) parses each
  failing visual run's log for Playwright list-reporter failure lines
  (`[project] › <path>.spec.ts:<line>:<col> › <title>`).
  - It writes the deduplicated `<path>:<line>` entries to
    `$(git rev-parse --git-path flow-verify-failed)/visual-failing.txt`.
  - It prints them in two groups: failures in the `specs` row's output ("the change's specs"), and
    failures elsewhere.
- **`-only <file>`.** Replaces `<specs>` with the file's entries, followed by
  `--repeat-each 5 --workers 1`.
  - It skips the `specs` row.
  - It never reads or writes the per-command green cache, and it never counts as the gate.
  - An empty or missing file exits 2.
  - A cell with no `<specs>` placeholder exits 2 and names the cell; the caller then runs full.
- **`-detect-flaky`.** Replaces `<specs>` with the `specs` row's output (or nothing), followed by
  `--retries=1 --fail-on-flaky-tests`.
  - A flaky test exits non-zero and lands in `visual-failing.txt` like any other failure.
  - It is a detector, never a pass: its green never writes the cache and never stands as the gate.
  - `rules/fix-determinism-at-the-source.mdc` is not reworded. The retry exists only to name flakes.
- **Loop algorithm.** It is canonical in `skills/flow/verify-fix-loop.md`; `verify-and-handoff.md`
  and `pipeline.md` point at it.
  1. The first failing plain visual run is followed by one `-detect-flaky` run. One fixer gets the
     union of both runs' failures in one batch.
  2. Each fix round re-runs `-only visual-failing.txt`.
  3. When the narrowed run is green, one plain run is the gate.
  4. If that plain run fails, step 2 repeats on the new failures.
- **A plain run** is one run of the `verify` cell over the `specs` row's output: K1's selection plus
  the smoke tail, or every spec when the full trigger fires (L). It is never the whole suite
  otherwise.
- **Item 4 escalation.** Two consecutive rounds whose failures all fall in the "elsewhere" group
  trigger `-detect-flaky` plus one batched fixer carrying every earlier round's report.
  - That fixer's prompt opens with `No-progress escalation`.
  - This extends `skills/flow-contracts/pipeline.md` l.172-182, beside the same-defect rule.
  - It is not a stop and not a cap; "No cap" in `verify-fix-loop.md` stays.
  - Under the full trigger (L) every spec is in the `specs` row's output, so "elsewhere" is empty
    and this escalation never fires; the same-defect rule still does.
- **"Untouched spec"** means any spec not in the `specs` row's output.
- **Non-visual lists.** Gradle test commands keep their per-command re-run. Narrowing applies to the
  visual command only.

### B. Inline verify waiting (item 5)

- `skills/flow/verify-and-handoff.md` l.106-113 is rewritten: `<verify>` is launched with the Bash
  tool's `run_in_background`, and the parent waits for its completion notification.
- The `nohup` + `until … sleep` text is deleted.
- The rule against hand-chunking `<specs>` stays.

### C. Per-round artifacts (item 6)

- The run-start deletion of `visual-*.log` (`runverifylist.go` l.535-538) is dropped.
- Each failing visual run writes `visual-<key>-r<N>.log`, where `N` is one past the highest existing.
- `keepResults` also copies the newest `test-results/<ts>` directory, from the dir the visual command
  ran in, to `flow-verify-failed/visual-results-r<N>/`.
- Retention ends with the worktree.

### D. Change base for spec selection (item 2, agents side)

- `run-verify-list` passes as `<merge-base>` the commit `flow.kickoff` branched each worktree from.
  It reads that commit from the state record's `worktrees[<abs-path>]` map:
  - `flow.kickoff` writes the entry for the canonical worktree;
  - `flow-guard kickoff-worktree --from` writes it for each added worktree.
- It falls back to `merge-base HEAD origin/<base>` only when the record has no entry for the
  worktree, and says so on stdout.
- This reaches every project's `specs` row and `visualTrigger`. It is the same base
  `visual-verify-pass prepare` already receives as `--merge-base`.

### E. Fix-key family at flow.verify (run deviation)

- At flow.verify, every fix is verify-fix-loop's `In-run fix <k>` task with `-fix-<k>` keys.
- `full-suite-fix-<n>` is stated as implement-phase only (`implement.md` l.1055). `verify-<k>` is
  stated as a re-run with no fix.
- `<k>` is the run's round, not the stage's: `verify-fix-loop.md` l.56-57 read "this stage's round
  in this run", so a flow.verify round 1 and a later flow.visual-verify round 1 both keyed
  `verify-fix-1` and `visual-verify-fix-1` — a collision on `dispatches_key_key` (one key per run,
  `0012_dispatch_key.sql`) that upserts the second round over the first. The line becomes "this
  run's round, counted across both stages", which l.155's "`<k>` the loop's last round" already
  assumes; no guard parses the number (`planshape.go` F24, `unfinishedwork.go`, `closetask.go`
  match the title shape only).
- `store/aggregate.go`'s `fix_rounds` counts both families, with one fix:
  - l.260-261 summed the highest `<k>` of `^verify-fix-(\d+)` and of `^visual-verify-fix-(\d+)`,
    but step 4 re-runs both stages under the same round's `-fix-<k>`, so k rounds counted 2k. The
    two terms become one, `MAX(substring(d.dispatch_key FROM '^(?:visual-)?verify-fix-(\d+)'))`:
    with a per-run `<k>`, the highest round of either family is the run's In-run fix round count;
  - l.264 counts each `implementer` row keyed `^full-suite-fix-[0-9]+$`;
  - the In-run fix task's own implementer row (`task-<n>-implementer`) matches no term, and l.262's
    `panel-fix` count excludes `%verify-fix-%`, so no round counts twice.

### F. Served model and effort (item 8)

- The readers prefer `metrics->>'modelVersion'` and `metrics->>'effort'` over the label columns:
  `aggregate.go` l.212-213, `records/render.go` l.640, `reviewdynamics.go` l.254/261/302.
- The label columns stay as recorded.
- The ledger shows `label → served` when the two differ.

### G. Per-dispatch cost (item 9)

- `PriceDispatch` in `store/pricing.go` prices a dispatch row's own `metrics.tokens` at its served
  model, at the rate in effect at `started_at`.
  - It reuses `combineMainAndSidechain`, `priceCandidate` and `pricingRateInEffect`.
  - It sets `metrics.cost_usd` with `||`, never adding to an existing value.
- It is called after `MergeDispatchMetrics`, and `RepriceSince` also loops over dispatches.
- Readers show it as a breakdown and never add it into a stage total, because `models.<m>` already
  holds those tokens.

### H. Worktrees and landing

- Both gymie worktrees are added at `flow.isolate-workspace`, gymie-playwright first:
  ```bash verified:usage line of stats/internal/guard/kickoffworktree.go l.152 at a5e73bab
  /Users/tweety53/Projects/agents/scripts/flow-guard kickoff-worktree --from /Users/tweety53/Projects/agents-worktrees/kan-1107-flow-verify-stop-full-suite-fix-reverify-loop /Users/tweety53/Projects/gymie-playwright kan-1107-flow-verify-stop-full-suite-fix-reverify-loop
  /Users/tweety53/Projects/agents/scripts/flow-guard kickoff-worktree --from /Users/tweety53/Projects/agents-worktrees/kan-1107-flow-verify-stop-full-suite-fix-reverify-loop /Users/tweety53/Projects/gymie kan-1107-flow-verify-stop-full-suite-fix-reverify-loop
  ```
  - gymie is cut from `origin/develop` and gymie-playwright from `origin/main`.
  - Each call records its kickoff base, runs the repo's `## worktree setup`, pushes with `-u`, and
    appends to `link.md` `## Merge order`.
- **Merge order:** `.`, `gymie-playwright`, `gymie`.
- **Route:** `merge and push`, from agents' `## default landing route`. Each repo merges into its own
  base: agents `main`, gymie `develop`, gymie-playwright `main`.
  - Run 1 stops at the first failure.
  - Run 2 removes all three worktrees.
- **Per-repo verification** runs each worktree's own `flow-guard project-get <worktree> lint|test`.
- **Tasks:** a task never spans two repos, and its `**Files:**` are paths relative to the root of
  the worktree it edits, as `flow-guard check-task-commit-fields` compares them.

### I. Item 2 in gymie-playwright and gymie, and the sibling frontend root

- **Follow callers for non-UI files** (`gymie-playwright/scripts/specs-for-diff.mjs`).
  - A touched non-inert `.kt` file outside the UI tree is walked like a UI file. `reachedBy`'s caller
    pool becomes `ui` plus `others`, followed transitively.
  - A non-UI file whose walk reaches no spec is named through `note(...)` as
    `<path>: no spec reaches it -- none selected for it`.
  - Non-Kotlin files still select nothing through the walk. The full trigger (L) covers the shared
    ones.
  - Tests:
    - repository → view model → tagged screen selects that screen's test;
    - an unreached non-UI file selects nothing and emits the note.
- **gymie `.flow/project.md` l.401-406** is rewritten to what the script does:
  - A UI file no spec reaches, a deleted file whose base text is unknown, a non-UI file whose
    callers reach no spec, and an empty diff each select nothing for themselves, named on stderr.
  - The output is `path` or `path:line` tokens plus the smoke tail. On a shared rendering path it is
    every spec.
  - `run-verify-list.sh` becomes `flow-guard run-verify-list`.
  - The "selects every spec" sentence is deleted.
- **Sibling frontend root (agents).** `visualverifypass_prepare.go`'s rule (l.283-290, l.513-518) is
  extracted into one helper, which it and `run-verify-list` both call. The rule: the first sibling
  worktree whose repo-prefixed diff matches the canonical `ui paths` becomes `<frontend-root>`, and
  that sibling's kickoff base becomes `<merge-base>`.

### J. Item 7 in gymie

- **7a. `affectedTests` Gradle task.** It replaces `scripts/affected-tests.sh`, which is deleted.
  `## test` becomes `./gradlew affectedTests`.
  - **Mapping.** A pure Kotlin function in `buildSrc` with JUnit tests in
    `buildSrc/src/test/kotlin/com/gymie/gradle/`. It keeps today's mapping exactly, untracked files
    included:
    - docs, specs, `.flow`, `scripts` → nothing;
    - `src/gateway` → `:gateway`;
    - `src/app/common` → full;
    - `src/app/src` → `:app`;
    - `src/app/<m>` → `:<m>`;
    - the dependent closure from `project(":x")` lines.
  - **Build-logic paths.**
    - `buildSrc/**`, `gradle/*.gradle.kts` and `gradle/wrapper/**` → a distinct "build logic"
      result with no module tests. Every `./gradlew` invocation already runs buildSrc's tests
      through `jar` finalizedBy `test`. A JUnit test asserts this result so it cannot fold into
      "nothing".
    - `gradle/deploy.gradle.kts` → also `:app:test`.
    - `gradle/libs.versions.toml`, root `build.gradle.kts` and `settings.gradle.kts` → full.
  - **Reading the diff.** The task reads
    `providers.exec { commandLine("git", …) }.standardOutput.asText.get()` eagerly inside
    `register { }` and calls `dependsOn` with the computed paths, because a lazy exec-provider
    `dependsOn` is silently dropped.
    - The base stays `merge-base HEAD origin/develop`.
    - A two-invocation test changes a file between runs and asserts that the second run schedules
      the new module's tests.
- **7b. Seeder tag.**
  - `@Tag("seeder")` goes on `LocalDev{,GroupSession,DevWeek,Friend,Tracking}SeederIntegrationTest`,
    `LocalDevCastSeederOrderIntegrationTest` and `LocalDevSeedWorkoutShapeIntegrationTest`.
  - `:app:test` excludes the tag. A new `:app:seederTest` includes only it.
  - `affectedTests` adds `:app:seederTest` for changes under
    `src/app/src/{main,test}/kotlin/com/gymie/dev/`, `src/app/src/main/resources/db/migration/` or
    `src/app/src/main/kotlin/db/migration/`, and on every full fallback.
  - The operator's full suite becomes `./gradlew test :app:seederTest`, and the `## test` prose says
    so.
- **7c. Testcontainers.** No build change (decision `testcontainers-no-change`).
- **7d. Suite records.**
  - gymie `## test` gets `# suite: gradle-affected`.
  - Agents side: `run-verify-list` wraps the visual `verify` row in `flow suite record -suite visual`
    on plain runs only. That runtime varies with the selection, so the record is per-run cost.

### K1. Per-test selection (`specs-for-diff.mjs`)

- Each spec is split into blocks: every `test(` and `test.describe(` call, top-level or nested, spans
  to its next sibling.
- **Output per spec:**
  - File-level code (imports, top-level helpers and hooks) reaching a touched tag → the bare path.
  - A describe's own body-level code (its hooks, `test.use`, local helpers) reaching it →
    `path:<describe line>`.
  - Otherwise, each test block reaching it → `path:<test line>`.
- Reach through a helper counts for the block that calls it. Both projects use the same code path.
- Tests:
  - one test of a two-test spec reached → `path:<that line>`;
  - a file-level `beforeEach` reached → the bare path;
  - a describe-scoped `beforeEach` reached → `path:<describe line>`.

### L. Selection tiers: smoke tail and full trigger

- **Interim smoke set.** The three tests of decision `interim-smoke-three-flows` get Playwright's
  `{ tag: '@smoke' }` details argument in their current files. No title or screenshot name changes.
- **Script.** `package.json` gains `"test:smoke": "playwright test --grep @smoke"`.
- **Smoke tail.** The K1 parser also finds every test whose call carries a literal `tag: '@smoke'`.
  - Every non-full output is the K1 selection plus each smoke test's `path:line`, deduplicated.
  - An empty diff prints exactly the smoke tail.
  - Tokens are emitted rather than `--grep`, because `--grep` ANDs with path filters.
- **Full trigger.** When the diff touches shared rendering code, the script prints the bare path of
  every spec in the `fidelity` and `functional` projects. It names the triggering file on stderr as
  `full: <path> is shared rendering code`. The shared paths:
  - `shared/src/commonMain/kotlin/com/gymie/shared/ui/theme/**`
  - `…/ui/components/**`
  - `…/ui/nav/**`
  - `…/ui/{AppScreens,AuthenticatedShellContent,ShellDestinationContent,ShellDrawer*,HomeBar,SessionBoot,PressSafeScroll,RevealOnFocus}.kt`
  - `shared/src/commonMain/composeResources/**`
  - `…/shared/platform/**`
  - `shared/src/wasmJsMain/**`
  - `webApp/**`
  - `build.gradle.kts`
  - `shared/build.gradle.kts`
- **Under A:**
  - The first plain run and every closing gate run the selection: touched + smoke, or full on the
    trigger.
  - `-only` rounds run `visual-failing.txt` only, with no smoke.
  - `-detect-flaky` covers the selection.
- **Tests in `scripts/specs-for-diff.test.mjs`:**
  - an empty diff prints the three smoke tokens;
  - a selected smoke test is printed once;
  - a touched `theme/BroadsheetColors.kt` prints every spec;
  - the parsed `@smoke` set equals exactly the three tests above, and any other tag fails the test.
- **Agents side:** no new flag. The tiers reach `run-verify-list` through the `specs` row's output.
- **gymie-playwright `## test`.** `.flow/project.md` gains `## test` with `npm run test:scripts`. Its
  "every other key is that project's" sentence names `## test` as its own.
- **`README.md`** (gymie-playwright) gets one paragraph on verify selection: per-test tokens, the
  smoke tail, `test:smoke`, the full trigger, and `npx playwright test` on demand.

#### Interplay with KAN-1109

- KAN-1109 moves the three tagged tests verbatim into `tests/auth`, `tests/editor` and
  `tests/runner`. The tag moves with each body. The selector finds tags by parsing and the
  describe-level K1 rule already exists, so no selector code changes.
- KAN-1109 changes the tagged-set test from "exactly these three" to "three flows plus one empty +
  maximal-fill test per area file". The line numbers in this design go stale with the merge, which
  is expected.
- KAN-1109's K3 measures the smoke tier's wall time. Until then, the three flows (about 22 s at the
  last measure) are the whole smoke tail.

## Critical files

Agents (`/Users/tweety53/Projects/agents-worktrees/kan-1107-flow-verify-stop-full-suite-fix-reverify-loop`):

- `stats/internal/guard/runverifylist.go` — failure parsing, `-only`, `-detect-flaky`, per-round
  artifacts, kickoff-base lookup, suite record, sibling root.
- `stats/internal/guard/run_verify_list_test.go` — its tests.
- `stats/internal/guard/visualverifypass_prepare.go` — source of the sibling-frontend rule, moved to
  the shared helper; `stats/internal/guard/visual_verify_pass_prepare_test.go` stays green.
- `stats/internal/guard/kickoffworktree.go` — writes the `worktrees[<abs-path>]` kickoff base read by
  section D (read, not edited).
- `skills/flow/verify-fix-loop.md` — canonical loop algorithm, escalation, fix keys.
- `skills/flow/verify-and-handoff.md` — `run_in_background` wait, pointer to the loop, fix keys.
- `skills/flow/implement.md` — `full-suite-fix-<n>` implement-phase only (l.1055), pointer.
- `skills/flow-contracts/pipeline.md` — the no-progress rule, l.172-182.
- `stats/internal/store/aggregate.go` — served model/effort (l.212-213),
  per-dispatch cost breakdown.
- `stats/internal/records/render.go` — ledger `label → served` (l.640), cost breakdown.
- `stats/internal/store/reviewdynamics.go` — served model/effort (l.254/261/302).
- `stats/internal/store/pricing.go` — `PriceDispatch`, `RepriceSince` over dispatches.
- `stats/internal/store/records.go` — `MergeDispatchMetrics`, after which `PriceDispatch` runs.
- `stats/README.md` — the pricing paragraph gains per-dispatch cost.
- `README.md` — Guardrails line for the verify narrowing.

gymie (`/Users/tweety53/Projects/gymie-worktrees/kan-1107-flow-verify-stop-full-suite-fix-reverify-loop`):

- `buildSrc/src/main/kotlin/com/gymie/gradle/` — the affected-tests mapping function.
- `buildSrc/src/test/kotlin/com/gymie/gradle/` — its JUnit tests.
- `build.gradle.kts` — the `affectedTests` task.
- `src/app/build.gradle.kts` — `:app:test` excludes `seeder`; `:app:seederTest`.
- `src/app/src/test/kotlin/com/gymie/dev/LocalDev*Seed*IntegrationTest.kt` (7 files) — `@Tag("seeder")`.
  <!-- measured: ls src/app/src/test/kotlin/com/gymie/dev/LocalDev*Seed*IntegrationTest.kt | wc -l in gymie @ 1a0b0eb9 -->
- `scripts/affected-tests.sh` — deleted.
- `.flow/project.md` — `## test`, `# suite: gradle-affected`, l.401-406 selector prose.
- `src/app/src/test/kotlin/com/gymie/AppIntegrationTest.kt` — the singleton pattern 7c cites (read).

gymie-playwright (`/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1107-flow-verify-stop-full-suite-fix-reverify-loop`):

- `scripts/specs-for-diff.mjs` — K1, I, L.
- `scripts/specs-for-diff.test.mjs` — its tests.
- `tests/auth.spec.ts`, `tests/workout-editor.spec.ts`, `tests/session-player.spec.ts` — `@smoke`.
- `package.json` — `test:smoke`.
- `.flow/project.md` — `## test`.
- `README.md` — verify-selection paragraph.
- `playwright.config.ts` — the `fidelity` / `functional` project definitions the full trigger lists
  (read).

## Existing code to reuse

- `runverifylist.go`: `keepResults` (l.227), the `<frontend-root>` / `<merge-base>` replacer
  (l.340), `visualTrigger`, the per-command green cache.
- `visualverifypass_prepare.go` l.283-290, l.513-518 — the sibling-frontend rule, extracted.
- The state record's `worktrees[<abs-path>]` kickoff base, written by `flow.kickoff` and
  `kickoff-worktree --from`.
- `flow suite record` (`stats/cmd/flow/suite.go`) for the `visual` suite record.
- `store/pricing.go`: `combineMainAndSidechain` (l.221), `priceCandidate` (l.643),
  `pricingRateInEffect` (l.674), `RepriceSince` (l.703).
- `store/records.go`: `MergeDispatchMetrics` (l.1424).
- `specs-for-diff.mjs`: `reachedBy`, `note(...)`, the `UI_FILE` test, the existing tag-reach walk.
- gymie buildSrc's `jar` finalizedBy `test` wiring — runs buildSrc's tests on every `./gradlew` call.
- Playwright 1.62.1 built-ins: `--repeat-each`, `--workers`, `--retries`, `--fail-on-flaky-tests`,
  `--grep`, the `{ tag }` details argument, positional `file:line` filters.

## Verification

Beyond each task's own verify step:

- **Agents:** from `stats/`, `go test ./internal/guard/... ./internal/store/... ./internal/records/...`
  and the lint list of `.flow/project.md` (`gofmt -l`, `go vet`, every `flow-guard check-*`).
- **gymie:** `./gradlew affectedTests` on a branch touching only `buildSrc/` schedules no module
  tests; touching `src/app/src/main/kotlin/com/gymie/dev/` schedules `:app:seederTest`;
  `./gradlew :app:test` runs no `seeder`-tagged test and `./gradlew :app:seederTest` runs exactly the
  seven.
- **gymie-playwright:** `npm run test:scripts` green; `node scripts/specs-for-diff.mjs <gymie
  worktree> <base>` on an empty diff prints exactly the three smoke tokens, and on a diff touching
  `theme/BroadsheetColors.kt` prints every spec with `full: … is shared rendering code` on stderr;
  `npm run test:smoke` runs three tests.
- **End to end (the loop):** in the gymie worktree, with one test made to fail,
  `flow-guard run-verify-list` writes `visual-failing.txt` with that `<path>:<line>` in the change's
  group and `visual-<key>-r1.log` plus `visual-results-r1/`; `-detect-flaky` then runs the
  selection with `--retries=1 --fail-on-flaky-tests`; `-only visual-failing.txt` runs just that
  entry 5× at one worker and writes no cache; with the test fixed, a plain run is the gate and a
  `visual` suite record appears in `flow suite`.
- **Stats:** after `flow pricing reprice -since <date>`, fixer dispatch rows carry `cost_usd`, and the
  ledger shows `label → served` on a row whose served model differs. This runs after landing, once
  the reinstall in this repository's `CLAUDE.md` (`./setup.sh global`, `make -C stats restart`) has
  put the new build behind `flowd`: until then `flowd` serves the pre-change build, and no task
  restarts it (`tasks.md` `## Global constraints`). Tasks 12 and 13 pin the same behaviour on
  `storetest` before landing.
- **End to end (the loop)** is task 23's Step 3.

## Non-goals

- The suite restructuring and gap captures listed under `## Context` (KAN-1109, KAN-1108).
- No round cap or stop on the verify loop; "No cap" stays.
- No wording change to `rules/fix-determinism-at-the-source.mdc`; `--retries=1` names flakes and is
  never a pass.
- No narrowing of non-visual (Gradle) test commands; they keep their per-command re-run.
- No `test:full` script, no scheduled or post-landing full run.
- No Testcontainers build change (decision `testcontainers-no-change`).
- No change to the label `model` / `effort` columns.

## Risks

- **A failure line the parser misses** (another reporter, a changed Playwright format) leaves
  `visual-failing.txt` empty: `-only` exits 2 and the caller runs full, so the loop degrades to
  today's cost, never to a false green. The parser is tested against real list-reporter lines.
- **A state record with no kickoff base** (a worktree created before this change) falls back to
  `merge-base HEAD origin/<base>` and says so on stdout — today's over-selection, visible.
- **Smoke tags read as literals only:** a non-literal tag is missed, and the tagged-set test fails
  rather than silently shrinking the tail.
- **Seeder miss:** a domain module a seeder calls does not run the seeders until a full run
  (decision `seeder-tag-and-task`).
- **Gradle configuration-time diff:** a lazy exec-provider `dependsOn` is silently dropped, so the
  task reads eagerly, and the two-invocation test fails if a stale diff is served.
- **Cost double count:** per-dispatch `cost_usd` overlaps `models.<m>` tokens; readers show it as a
  breakdown only, never summed into a stage total.
- **Three-repo landing:** run 1 stops at the first failing repo, so agents can land while a gymie
  gate is red; the merge order puts the agents tooling first so that state is consistent.
- **No item-4 escalation under the full trigger:** on a change touching shared rendering code,
  "elsewhere" is empty (A), so only the same-defect rule escalates — a loop chasing distinct
  failures in specs the change does not touch runs on without the batched fixer.
- **`affectedTests` over-selects on merged-in branches** (decision
  `affected-tests-base-origin-develop`): diffing from `origin/develop` keeps KAN-1086's extra module
  runs. A `-PaffectedBase=<sha>` property filled by `run-verify-list` from section D's kickoff base
  would remove it without the `flow` CLI, but needs a base placeholder substituted into `## test`
  cells; left for a later change, since the cost is module test runs, never a Playwright suite.
  Accepted: an environment variable `run-verify-list` exports to every fence command (plan-review-3
  F5) needs no placeholder and no `flow` CLI in Gradle, but section D reads the kickoff base inside
  the visual row only, so exporting it means a `flow state get` on every `lint` and `test` list run
  plus a gymie read of the variable — left for the same later change, for the same reason.
- **Full trigger on a backend-only diff:** with no frontend sibling, `<frontend-root>` is the
  canonical worktree (decision `shared-sibling-frontend-helper`), so a gymie diff touching both a
  `ui paths` path and its own root `build.gradle.kts` matches the `build.gradle.kts` glob of L and
  runs every spec. Accepted: the cost is over-selection — more coverage, never less — and task 23's
  end-to-end run takes that path.
- **`fix_rounds` on runs recorded before this change:** a run whose `<k>` restarted per stage
  undercounts — its two stages' rounds share numbers, so the combined `MAX` sees the larger loop
  only. Those runs are history; no backfill.
- **KAN-1109 line drift:** the smoke line numbers here go stale with KAN-1109's merges, which is
  expected; the tag moves with each test body.
