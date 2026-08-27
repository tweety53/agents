// Shared fixtures for the visual specs added under task 12 (KAN-171 part
// 4). `baseline.spec.ts` predates this file and keeps its own copies of
// the same values -- it is the committed baseline these specs must not
// weaken, so it is left untouched rather than refactored onto this module
// (a second caller, not the first one, is what justifies extracting a
// shared helper -- engineering-principles.md's DRY entry).
//
// PINNED_QUERY and EMPTY_QUERY are both explicit `from`/`to` query params
// for the same reason baseline.spec.ts's own DASHBOARD_URL comment gives:
// the UI-test fixture's timestamps are fixed at seed time
// (cmd/uitest-seed/seed.go's fixtureNow, 2026-08-15T09:00:00Z), so a
// window computed from "now" (the default period, or a relative preset
// like "7 days") ages the fixture out of range a few weeks after this
// spec is written, and every assertion below would start failing for a
// reason that has nothing to do with the control under test.
export const PINNED_QUERY = "from=2026-08-01T00%3A00%3A00.000Z&to=2026-08-16T00%3A00%3A00.000Z";

// A period entirely before the fixture's own data -- `recorded: false` on
// every view (verified against the live UI-test stack: GET
// /api/v1/stats/state-board?from=2020-01-01T00:00:00Z&to=2020-01-02T00:00:00Z
// returns `"recorded": false, "rows": []`). Used as a deliberately-empty
// starting point so a spec can change the period *in the browser* (typing
// into PeriodPicker's inputs, or clicking a preset) and assert that the
// resulting reload is what populated the table -- proving the period
// control's effect is visible, not just that the pinned URL happens to
// carry data already.
export const EMPTY_QUERY = "from=2020-01-01T00%3A00%3A00.000Z&to=2020-01-02T00%3A00%3A00.000Z";

// The datetime-local input value PeriodPicker.tsx's own toInputValue
// produces for the two instants PINNED_QUERY names.
//
// These are NOT timezone-independent, and nothing here pins a timezone:
// playwright.config.ts sets no `timezoneId`, so Chromium inherits the
// host's zone -- EEST (UTC+3) on the machine this was written on, not UTC.
// What makes the values safe is the WIDTH of the window, not a matching
// zone: PINNED_QUERY spans 2026-08-01 to 2026-08-16, so an offset of a few
// hours cannot move either endpoint across a day boundary, and both this
// file and PeriodPicker read "local" the same way in whatever zone the run
// happens to use.
//
// If a future assertion needs the input value to be exact to the hour,
// pin `timezoneId` in playwright.config.ts rather than widening this
// comment -- the safety here is a margin, not a guarantee.
export const PINNED_FROM_INPUT = "2026-08-01T00:00";
export const PINNED_TO_INPUT = "2026-08-16T00:00";

// theme.css's dark surface-1 token, as a browser computes it -- the same
// value baseline.spec.ts's own DARK_SURFACE_1 asserts against, kept here
// under its own name so a spec added to this directory can assert the
// white-page defect class without importing from the file it must not
// modify.
export const DARK_SURFACE_1 = "rgb(17, 18, 23)";

export async function bodyBackground(page: import("@playwright/test").Page): Promise<string> {
  return page.evaluate(() => getComputedStyle(document.body).backgroundColor);
}
