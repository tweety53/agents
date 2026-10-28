- [ ] 5. In-run fix 2 — flow.visual-verify: no capture picks a project or a task size
  - [ ] **Step 1: Seed.** `seedReviewDynamics` records one decision per `uitest-dynamics` change, class `big` for the first change of each week and `small` for the second, so a Task size pick has data to draw.
  - [ ] **Step 2: Spec.** `review dynamics filters` also captures `review-dynamics-picked.png` after picking project `uitest-dynamics` and Task size `Big`, each asserted pressed.
  - [ ] **Step 3: Capture.** Restart the UI-test stack (re-seed), the project's `capture` for `review-dynamics.spec.ts` with `--update-snapshots=all` and for the full app suite, every changed PNG read, `rebuild-full-app-suite-zip.sh`, `visual-verification.md` updated; `npm run test:visual` exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/review-dynamics.spec.ts`, `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/review-dynamics-picked-darwin.png`, `stats/web/tests/visual/full-app-suite.zip`, `spectre/changes/kan-1047-review-dynamics-heatmap-by-project/visual-verification.md`
**Tests:** `review dynamics filters`
**Regression:** `review dynamics filters` fails — the project or Task size pick no longer presses its button or no longer matches its baseline.
**Commit:** test(stats): seed plan classes and capture a project and task-size pick
**After:** Task 4
