## Context

A standalone self-review fix loop over five held tickets. The triage evidence for each was re-read
against the trees at plan time:

- KAN-836, KAN-834 — `skills/flow/visual-verify-verifier.md` step 8 carries only the `expectRoom`
  on-screen assertion; no rule in `skills/` or `rules/` covers a positive control, a parked pointer
  or a dimming layer.
- KAN-833 — `skills/flow/engineering-principles.md` `### Testing principles` lists no
  measured-bound violation; the PIXEL PROBE paragraph of `skills/flow/review-panel.md` binds only
  panel fix subagents.
- KAN-832 — gymie-frontend's `build.gradle.kts` registers no `verifyChange`, and its
  `.flow/project.md` has no `## test` or `## lint`. Its root applies the detekt and ktlint plugins
  with `apply false`, so the root project has neither task; every subproject gets both from the
  `subprojects {}` block. With the planned task applied, `./gradlew verifyChange :shared:desktopTest --tests 'com.gymie.shared.ui.HomeScrollReachUiTest'`
  ran as one invocation and exited 0, its dry run listing `ktlintCheck` and `detekt` for all four
  of `:shared`, `:androidApp`, `:desktopApp`, `:webApp`.
  <!-- measured: ./gradlew -q verifyChange --dry-run and the one-invocation run above, with the Task 4 build.gradle.kts hunk applied @ gymie-frontend 755c74ea → 4 subprojects listed, exit 0 -->
- KAN-827 — the CLI journals a mark on any store failure, including a timeout and a non-success
  answer (`stats/cmd/flow/stage.go`, `skills/flow-contracts/state-file.md` **The pipeline never
  blocks**). flowd replays at startup and in `reconcile.Watch` only on a failing-to-succeeding ping;
  `flow journal flush` is called by no skill. A new `TestWatchReplaysOnHealthyTick` (always-healthy
  ping, entry appended after Watch starts) fails at `02d83a03` and passes with the planned change;
  the whole `./internal/reconcile` package passes with it.
  <!-- measured: go test ./internal/reconcile -run 'TestWatchReplaysOnHealthyTick' -count=1 @ 02d83a03 → FAIL "Watch never replayed…"; go test ./internal/reconcile -count=1 with the Task 5 diff applied @ 02d83a03 → ok -->

The three run-loaded prose edits (Tasks 1, 2, 3) and the reworded contract sentence (Task 5) trip
`scripts/check-verbatim-moves.sh` — six FAIL lines with every edit applied at `02d83a03`. This
change's `verbatim-moves.txt`, committed with the plan, lists each sentence exactly as the FAIL
line prints it, so the fixer writes nothing under `spectre/changes/`.
<!-- measured: scripts/check-verbatim-moves.sh with all four tasks' doc edits applied @ 02d83a03 → 6 violation(s), exit 1 -->

## Decisions

### Click-reached captures: positive control and parked pointer in one sentence

**ID:** click-capture-positive-control
**Status:** active
**Chosen:** one bolded sentence right after step 8's `expectRoom` sentence, naming both mechanisms (an assertion on a drawn, state-dependent property; `page.mouse.move(0, 0)` after the last click) — they guard the same capture against the same mis-click.
**Considered:** changing gymie-playwright's `clickTag` to park the pointer itself — a cross-repo helper change that still leaves the positive control unstated, and the rule must reach every project's verifier, not one suite.

### Dimmed captures: structural assertion, no threshold change

**ID:** scrim-structural-assertion
**Status:** active
**Chosen:** a step 8 sentence requiring tag presence/absence beneath any scrim, dimming or alpha layer (covered tags included) and a recapture of a dimmed baseline that predates the change.
**Considered:** lowering gymie-playwright's per-pixel `threshold` — re-flags antialiasing on every capture; tolerance tuning, not the seam.

### Measured-bound violation in engineering-principles

**ID:** bounds-violation-in-principles
**Status:** active
**Chosen:** append to `### Testing principles`' **Violation looks like:** list — the file implementers read as REQUIRED READING and the principles reviewer applies.
**Considered:** amending `skills/flow/review-panel.md`'s PIXEL PROBE paragraph — reaches panel fix subagents only, never the implementer who writes the test first.

### No rationale-appendix lines for the new rules

**ID:** no-rationale-lines
**Status:** active
**Chosen:** add nothing to `skills/flow/SKILL-rationale.md` — each new sentence carries its own reason inline ("so a missed click fails…", "because the dimming can shrink…"), and no existing rule's reasoning changes.
**Considered:** a KAN-836/KAN-834 line beside the KAN-747 one — restates the inline reason in a file no run loads.

### verifyChange over subprojects only

**ID:** verifychange-subprojects
**Status:** active
**Chosen:** `tasks.register("verifyChange")` depending on each subproject's `ktlintCheck` and `detekt`, no `ktlintFormat`, with gymie's KAN-630 comment adapted — the root has neither task, since it applies both plugins with `apply false`.
**Considered:** copying gymie's `dependsOn(":ktlintCheck", ":detekt")` — names root tasks this root does not have; including `ktlintFormat` — a source-mutating task races the checks under the parallel executor (KAN-780).

### Frontend test and lint sections mirror gymie's

**ID:** frontend-test-lint-sections
**Status:** active
**Chosen:** `## test` with the full `./gradlew test` line and the `./gradlew verifyChange :shared:desktopTest --tests '<selectors>'` bundle line; `## lint` with `./gradlew ktlintFormat` as its own pre-step, then `./gradlew verifyChange`, and gymie's read-only-bundle paragraph. Plans written after this read the commands from there; archived plans and `CONTRIBUTING.md` are untouched.
**Considered:** also rewriting `CONTRIBUTING.md`'s `./gradlew ktlintCheck detekt` line — still correct, and outside what the ticket asks.

### Watch replays on every healthy ping

**ID:** watch-replay-every-healthy-tick
**Status:** active
**Chosen:** `reconcile.Watch` calls `Run` on every tick whose ping succeeds and drops the `healthy` transition flag; the package doc, `Watch`'s doc, `cmd/flowd/main.go`'s two comments and `skills/flow-contracts/state-file-internals.md`'s replay sentence follow. `Run` retires entries idempotently and `mu` serializes in-process calls; an empty journal costs one directory walk, and `logReconcileResult` logs only when something applied or was refused, so the steady state is silent.
**Considered:** a CLI-side replay on the next successful store write — every CLI command grows a journal walk and a second replayer process racing flowd's; calling `flow journal flush` from a skill at stage boundaries — a run-side step every skill must remember, where the daemon already holds the timer.

### Refused-replay handoff notice excluded

**ID:** refused-replay-handoff-excluded
**Status:** active
**Chosen:** KAN-827's "name a mark that cannot land in the handoff" is not built here — the CLI exposes no reconcile result to a run, so it would be a new read path; the stranding, the defect the ticket's evidence shows, closes without it. flowd still logs every refused entry.
**Considered:** adding a reconcile-result endpoint and a handoff line — a separate addition outside the defect's root cause.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 5 changes a daemon loop every `flow` call's journal depends on, and Task 4 a build in another repository; both warrant the stronger first look.
**Considered:** `flow-low` — the three prose tasks alone would fit it, but the review covers the whole branch.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** two groups in series — `self-review-self-review-backlog-flow-5-oct18b-fix-g1`: Tasks 1, 2, 3 (agents prose, Task 2 editing the sentence Task 1 adds); `-fix-g2`: Tasks 4, 5 (gymie-frontend build, flowd reconcile — each with a build or test run).
**Considered:** one group of five — one fixer runs a Gradle build, a Go suite and three prose edits in one context; five groups — five dispatches for three one-sentence edits.

## Open questions

## Live check

Task 5 touches flowd. After landing and a flowd restart on the merged code: with flowd running and
its database up, append a stage-mark entry to a test change's `*.journal.stage` file by hand, then
wait one ping interval. Record the journal's pending-entry count before and after, and the store's
stage row. Working: the entry leaves the journal and the row appears within one interval, with an
`applied=1` "flowd journal replay" log line. Not working: the entry stays in the journal past the
interval and no log line appears — the pre-fix behaviour. Tasks 1–4 have no runtime to verify
against: prose, and a Gradle task whose one-invocation run is Task 4's own verify step.
