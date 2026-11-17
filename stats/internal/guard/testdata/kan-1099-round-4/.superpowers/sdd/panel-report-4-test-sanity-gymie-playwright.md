Critical: 0
Important: 0
Minor: 0
Unread: none

# Panel report — round 4 — test-sanity — gymie-playwright section

### Summary

Read the `gymie-playwright` section of `final-review.diff` in full (binary snapshot hunks noted only, not opened). Tests added or changed:

- New fidelity specs: `fidelity/kan-1099-auth.spec.ts`, `kan-1099-body.spec.ts`, `kan-1099-home-editor.spec.ts`, `kan-1099-profile.spec.ts`, `kan-1099-public.spec.ts`, `kan-1099-sheets.spec.ts`, `kan-1099-text-lengths.spec.ts` and their `.mockups` sidecars.
- Changed: `add-participant`, `builder`, `copy-fallback` (caretOff before captures; H4 no longer taps the chevron since SH27 keeps the search up), `body-calories-tab` and `copy-session-to-another-day` (`waitForPickedDay`), `full-app-suite` and `tests/auth.spec.ts` (email masks read from the live tag), `kan-754-offline-l3`, `kan-930-day-groups`, `kan-940-cardio-rows`, `nav-scoped-rights`, `profile-body-fields`, `tests/body-weigh-in-history-mockup.spec.ts` (long-press menu, delete-with-Undo), `scripts/specs-for-diff.test.mjs` (smoke line numbers), and the helpers `fidelity/fixtures.ts` (`typeOver`), `fixtures.ts` (`waitForPickedDay`, `yourLinkHost`), `fidelity/offline-l5.ts`, `fidelity/frames.ts`.

Lenses applied:

- Can it fail: every new spec ends in `toHaveScreenshot` against a committed baseline, so each fails on a visual regression. The non-screenshot assertions were checked against the production code they rely on: SH1b's "no login request" and both fields' fill change (`AuthScreens.kt` `broadsheetFieldStyle(accented = error != null)`), SH3b's email-field shift, SE11's "no PUT /api/body/goal", SH13's 422 plus `notification_banner` absence, SH22's `cardio_step_value_helper` absence, SH15/N9's Undo re-log (201). Every tag those negative assertions probe exists in the frontend (`CardioStepSheet.kt`, `NotificationHost.kt`, `AuthScreens.kt`), so the `toBeNull` checks are not vacuous on a misspelt tag.
- `waitForPickedDay` accepts `cal_day_pill_<iso>` or `cal_day_selected_<iso>`; checked against `HomeCalendarGrid.kt:462-464`: `DatePickerSheet` draws HOME style (pick = ring), `CopySessionSheet` draws PICKER style (pick = fill), so each arm is the real pick signal for one caller. The other arm can match before the click only when the target day is today, and no caller picks today (2026-09-12, 2026-09-26, 2026-09-19 against PINNED_TODAY 2026-09-02). That is a theoretical risk only, so not raised.
- `tests/auth.spec.ts` now masks the whole email field box. The invalid-credentials refusal is drawn on the password field (`AuthScreens.kt:948` maps it to `fields["password"]`), outside the mask, so the screenshot still pins the refusal.
- Behaviour vs trivia / Duplicate / Ceremony: the `.mockups` sidecars and `frames.ts` heights are harness data, not tests. SH22b's `expect(RIDE_6X).toContain('6x(')` asserts on the test's own constant, but it is a fixture precondition guard (it fails if `RIDE_LINE` stops carrying `4x(`) and is not named as a behaviour test. Not raised.
- Honest name: the N9 rename ("delete with Undo") matches its new assertions.

No scratch-copy mutation was run: every candidate was settled by reading the production code. Each spec needs the full dev stack (backend, frontend, Playwright), so running one is not cheap.

Coverage: read every non-binary hunk of this section in full. Production files read at the cited lines only: `HomeCalendarGrid.kt`, `AuthScreens.kt`, plus tag greps across `shared/src/commonMain`. Snapshot PNGs were not opened.

### Issues

#### Critical (Must Fix)
None.

#### Important (Should Fix)
None.

#### Minor (Nice to Have)
None.

### Assessment
**Ready for the human gate?** Yes
**Reasoning:** Every added or changed test can fail when the behaviour it names breaks. The negative assertions probe tags that exist, and the changed masks and waits still cover the behaviour their tests name.
