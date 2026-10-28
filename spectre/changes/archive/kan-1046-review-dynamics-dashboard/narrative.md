# kan-1046-review-dynamics-dashboard — session narrative

## 2026-10-28 — creating run

- **Mid-run, the operator asked for model versions to be kept apart on every panel.** They answered "Same + backfill history", which added Tasks 8–10:
  - the harvester stamps `modelVersion`;
  - `model_family` is added, with version-keyed series and heat;
  - `flowd backfill-models` is added.

  The operator then clarified the series key as `model-version · effort`. Prototype v3, which adds the Family/Version toggle, was re-published for their check.
- **Backfill route.** The plan first routed the backfill through an API endpoint. That was abandoned for a direct-database `flowd` subcommand, matching `flowd restamp-dispatch`. The correction is recorded in tasks.md.
- **Panel round 0** raised 10 findings, F1–F10. They were fixed inline in one commit.
  - The fix also fills every range bucket on the x axis.
  - It also auto-maps any `skills/flow/<slot>-reviewer-prompt.md` file to area review, so reviewer slots added later are imported without code changes.
- **Panel rounds 1 and 2:** 7 mutation survivors were left after round 1. Six were killed with tighter tests (author ≠ committer date, a distinct tuning area, an unreadable subdirectory, a padded unknown model, nested reviewer-prompt paths).
  - Writing those tests exposed a real gap: the import took a nested path that git's `*` pathspec matches as a reviewer prompt. It is now filtered.
  - F15, the ignored ls-files error, was withdrawn as an equivalent mutant.
- **Live check.** It ran on a pg_dump copy of the dev store with a scratch flowd on port 5299. Results: 148 rows imported, 2208 dispatches stamped by the backfill, and one hand-counted week matching the API.
  - The hand count first disagreed: 73 changes against 37. The hand query had omitted the panel-slot filter (`slot IS NOT NULL`) that defines the cohort. Once the filter was added, the counts matched.
  - Lesson: the project's `scripts/live-store-copy.sh` exists for exactly this and should be used next time.
- **Visual verify, first pass**, found 7 departures from the approved prototype, plus 3 missing state captures.
  - In-run fix 1 (Task 11) closed them all: the intro type, heatmap packing and order, effort order, the tooltip surface, and the Version, default-preset and cell-hover captures.
  - The default-preset capture pins Playwright's clock to the fixture's end.
  - The heatmap's cell hover is the native `title` tooltip, which no screenshot draws, so the spec asserts its text.
- **Panel round 3** left 3 Important mutation survivors in the ordering logic: the fixture's busiest slot also sorted first by name, and it had a single model. Round 4 killed all of them with a two-model fixture.
- **Guard conflict.** `check-plan-shape` requires an in-run visual fix task to list `visual-verification.md` in **Files:**, and `check-task-records` requires the task commit to touch it. The record was untracked when Task 11 was committed. It was folded into Task 11's commit with a fixup and autosquash, and the branch was force-pushed with lease. The implementer and panel-fix rows were re-pointed to the rewritten shas.
- **Recapture noise.** The closing full verifier run's `=all` recapture rewrote 4 unrelated full-app-suite PNGs with 3×3px anti-aliasing noise. They were stashed rather than committed.

## 2026-10-28 — integrate run

The preflight and the unfinished-work gate were clear, and origin/main had not moved since the merge base, so no rebase was needed. The branch lands by the configured default route, merge and push.

In-run pipeline fix: 2f5f3d50 — land-self-review-report refused a prompt-tuning case committed on the change's own branch, which step 2a puts there, because it looked on origin/main alone (blast radius 2 files)
