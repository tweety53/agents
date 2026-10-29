- [x] 7. In-run fix 2 — visual-verify: a finger-speed month-grid swipe still draws the day pressed
**Build:** green
**Files:** `shared/src/commonMain/kotlin/com/gymie/shared/ui/workout/HomeCalendarGrid.kt`, `tests/kan-1059-repeat-keep-awake-swipe.spec.ts`, `spectre/changes/kan-1059-repeat-row-keep-awake-calendar-swipe/visual-verification.md`
**Tests:** gymie-playwright `tests/kan-1059-repeat-keep-awake-swipe.spec.ts` › "the day under the finger draws no press mid-drag"
**Regression:** the paced probe fails on the Task 6 code and passes after the fix, 5 runs of 5.
**After:** Task 6
**Commit:** `fix(home): hold back the day press under the calendar swipe`

Defect, verbatim from the verifier's report: calendar-swipe-month-grid: DEPARTURE — month-grid press NOT fixed. Under a finger-speed drag (30 moves 16ms apart, 150px left, from day 10) frames 002–019 (51–395ms) draw the pressed grey fill (201,200,200 against 243,242,242) on day 10 while the pointer has moved away, fading 020–021. The delivered probe (12 instantaneous moves, 300ms wait) passes over the defect.

Cause: the day cell's `clickable` emits its press at once, because no ancestor is a scrollable container; Compose delays the press (`TapIndicationDelay`) only under one, and the swipe's consumption past slop cancels a delayed press but not an emitted one. Web Compose's `TapIndicationDelay` is zero, so even under one the press draws at once.

  - [x] **Step 1:** Rewrite the probe as a paced drag (30 moves 16ms apart, 150px left) sampling the cell during the drag; run red on the Task 6 code.
  - [x] **Step 2:** The day cell's `clickable` takes its own interaction source and no indication; `rememberHeldBackPresses` forwards a press to `LocalIndication` only once held 100ms (Android's tap timeout), a release before that as press and release at once, a cancel before that as nothing.
  - [x] **Step 3: Verify.** Frontend `verifyChange` and `:shared:desktopTest --tests 'com.gymie.shared.ui.HomeCalendarSwipeUiTest' --tests 'com.gymie.shared.ui.HomeCalendarUiTest' --tests 'com.gymie.shared.ui.WeekStripUiTest'`; the probe 5 runs of 5 green.
  - [x] **Step 4: Commit** with the subject above.

Correction (2026-10-29): Step 2 as planned — a no-op horizontal `scrollable` over the swipe area — left the probe red 5 runs of 12: web Compose's `TapIndicationDelay` is 0, so a scrollable ancestor delays nothing. The press is held back in the cell instead; the probe went green 12 of 12, the full capture spec 3 of 3. **Files:** names `HomeCalendarGrid.kt`, not `HomeScreen.kt`, and the commit subject follows.

