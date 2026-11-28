// Shared fixtures for the visual specs added under task 12 (KAN-171 part
// 4). `baseline.spec.ts` predates this file and keeps its own copies of
// the same values -- it is the committed baseline these specs must not
// weaken, so it is left untouched rather than refactored onto this module
// (a second caller, not the first one, is what justifies extracting a
// shared helper -- engineering-principles.md's DRY entry).
//
// PINNED_QUERY is an explicit `from`/`to` query param pair
// for the same reason baseline.spec.ts's own DASHBOARD_URL comment gives:
// the UI-test fixture's timestamps are fixed at seed time
// (cmd/uitest-seed/seed.go's fixtureNow, 2026-08-15T09:00:00Z), so a
// window computed from "now" (the default period, or a relative preset
// like "7 days") ages the fixture out of range a few weeks after this
// spec is written, and every assertion below would start failing for a
// reason that has nothing to do with the control under test.
export const PINNED_QUERY = "from=2026-08-01T00%3A00%3A00.000Z&to=2026-08-16T00%3A00%3A00.000Z";

// theme.css's dark surface-1 token, as a browser computes it -- the same
// value baseline.spec.ts's own DARK_SURFACE_1 asserts against, kept here
// under its own name so a spec added to this directory can assert the
// white-page defect class without importing from the file it must not
// modify.
export const DARK_SURFACE_1 = "rgb(17, 18, 23)";

export async function bodyBackground(page: import("@playwright/test").Page): Promise<string> {
  return page.evaluate(() => getComputedStyle(document.body).backgroundColor);
}

// cmd/uitest-seed's fixtureNow. The live state board filters by Updated
// relative to the browser clock and the run view measures open runs to it,
// so specs over either pin the clock here (kan-1048).
export const FIXTURE_NOW = new Date("2026-08-15T09:00:00Z");
