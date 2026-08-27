import { expect, type Page, test } from "@playwright/test";
import { EMPTY_QUERY, PINNED_FROM_INPUT, PINNED_QUERY, PINNED_TO_INPUT } from "./support";

const DAY_MS = 24 * 60 * 60 * 1000;

async function periodInputRangeMs(page: Page): Promise<{ fromMs: number; toMs: number }> {
  const fromValue = await page.getByLabel("Period start").inputValue();
  const toValue = await page.getByLabel("Period end").inputValue();
  return { fromMs: new Date(fromValue).getTime(), toMs: new Date(toValue).getTime() };
}

// The datetime-local inputs PeriodPicker.tsx renders carry no seconds
// field (toInputValue's own "T${H}:${Min}"), so a value read back through
// `new Date(inputValue)` always lands on a whole minute -- up to 59.999s
// off whatever instant the preset actually computed. PeriodPicker.tsx's
// own `sameRange` uses the identical 60s tolerance for the identical
// reason (its own header comment: "the `to` bound ... moves forward on
// every re-render"); this file reuses that number rather than a
// re-derived one, plus a few seconds for this test's own execution time.
const INPUT_MINUTE_TRUNCATION_TOLERANCE_MS = 65_000;

/**
 * Asserts PeriodPicker's own "From"/"To" inputs, at the moment this is
 * called, span approximately `spanMs` ending approximately now --
 * PeriodPicker.tsx's own shape for every relative preset except `Today`,
 * which has its own assertion below. Deliberately time-relative rather
 * than pinned: `7 days`/`30 days`/`90 days` are themselves relative-to-now
 * controls (PeriodPicker.tsx's own PRESETS), so pinning this assertion to
 * a fixed calendar date would make it fail the day after it was written
 * for a reason that has nothing to do with the control under test.
 */
async function expectRangeSpan(page: Page, spanMs: number, toleranceMs = INPUT_MINUTE_TRUNCATION_TOLERANCE_MS) {
  const { fromMs, toMs } = await periodInputRangeMs(page);
  expect(Math.abs(toMs - Date.now())).toBeLessThan(toleranceMs);
  expect(Math.abs(toMs - fromMs - spanMs)).toBeLessThan(toleranceMs);
}

const PRESET_NAMES = ["Today", "7 days", "30 days", "90 days", "All time"];

async function expectOnlyPressed(page: Page, pressedName: string) {
  for (const name of PRESET_NAMES) {
    await expect(page.getByRole("button", { name, exact: true })).toHaveAttribute(
      "aria-pressed",
      name === pressedName ? "true" : "false",
    );
  }
}

// task 12 (KAN-171 part 4): every control the stats SPA exposes, exercised
// on its own -- not just the three view-level baselines. Each spec asserts
// the control's effect in the DOM (a row appearing or disappearing, an
// attribute flipping, the URL changing), never a screenshot alone: a
// screenshot proves the page painted, not that the control did what it
// says. Values are read from the live UI-test stack (see support.ts's own
// header comment and views.spec.ts's for the endpoints queried), not
// assumed.
//
// The state-board table is the shared target for most of these: its
// DataTable is the one in this app with the richest column set (five
// sortable columns, two filterable ones), and PINNED_QUERY/EMPTY_QUERY
// give a deterministic "populated" and "empty" starting point without
// depending on real wall-clock time the way clicking most of
// PeriodPicker's own presets would (all-time is the one exception below,
// and is timezone/now-agnostic by construction: it always spans the
// fixture's fixed 2026-08-15 data no matter when this spec runs).

test.describe("PeriodPicker", () => {
  test("the All time preset broadens the table from empty to the fixture's own rows", async ({ page }) => {
    await page.goto(`/#/state-board?${EMPTY_QUERY}`);
    const tablePanel = page.locator('section.panel[aria-label="Every open change"]');
    await expect(tablePanel.getByTestId("not-recorded")).toBeVisible();

    const allTime = page.getByRole("button", { name: "All time" });
    await allTime.click();
    await expect(allTime).toHaveAttribute("aria-pressed", "true");

    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-201-refactor-thing" })).toBeVisible();
  });

  test("typing explicit From/To values reloads the table to that window", async ({ page }) => {
    await page.goto(`/#/state-board?${EMPTY_QUERY}`);
    const tablePanel = page.locator('section.panel[aria-label="Every open change"]');
    await expect(tablePanel.getByTestId("not-recorded")).toBeVisible();

    await page.getByLabel("Period start").fill(PINNED_FROM_INPUT);
    await page.getByLabel("Period end").fill(PINNED_TO_INPUT);

    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-201-refactor-thing" })).toBeVisible();
  });

  // The remaining four presets, plus the conditional Reset button --
  // PeriodPicker.tsx's own PRESETS array and its `!sameRange(...)` guard
  // around "Reset to default". Each preset is itself relative to "now",
  // so these assert the picker's own computed range (against a
  // same-instant Date.now() read, see expectRangeSpan's own header
  // comment) rather than which fixture rows that range happens to
  // include -- the same reason this file's own header comment gives for
  // why `All time` is the one preset safe to assert fixture rows against.
  test("Today sets the range to midnight through now, and presses only itself", async ({ page }) => {
    await page.goto(`/#/state-board?${EMPTY_QUERY}`);
    await page.getByRole("button", { name: "Today", exact: true }).click();
    await expectOnlyPressed(page, "Today");

    const { fromMs, toMs } = await periodInputRangeMs(page);
    const expectedMidnight = new Date();
    expectedMidnight.setHours(0, 0, 0, 0);
    expect(Math.abs(fromMs - expectedMidnight.getTime())).toBeLessThan(INPUT_MINUTE_TRUNCATION_TOLERANCE_MS);
    expect(Math.abs(toMs - Date.now())).toBeLessThan(INPUT_MINUTE_TRUNCATION_TOLERANCE_MS);
  });

  test("7 days sets a seven-day range ending now, and presses only itself", async ({ page }) => {
    await page.goto(`/#/state-board?${EMPTY_QUERY}`);
    await page.getByRole("button", { name: "7 days", exact: true }).click();
    await expectOnlyPressed(page, "7 days");
    await expectRangeSpan(page, 7 * DAY_MS);
  });

  test("30 days sets the default thirty-day range, and presses only itself", async ({ page }) => {
    await page.goto(`/#/state-board?${EMPTY_QUERY}`);
    await page.getByRole("button", { name: "30 days", exact: true }).click();
    await expectOnlyPressed(page, "30 days");
    await expectRangeSpan(page, 30 * DAY_MS);
  });

  test("90 days sets a ninety-day range ending now, and presses only itself", async ({ page }) => {
    await page.goto(`/#/state-board?${EMPTY_QUERY}`);
    await page.getByRole("button", { name: "90 days", exact: true }).click();
    await expectOnlyPressed(page, "90 days");
    await expectRangeSpan(page, 90 * DAY_MS);
  });

  test("Reset to default only renders when the period differs from default, and resets it", async ({ page }) => {
    await page.goto(`/#/state-board?${EMPTY_QUERY}`);
    const reset = page.getByRole("button", { name: "Reset to default" });
    // EMPTY_QUERY (2020) is nowhere near the last 30 days, so the button
    // PeriodPicker.tsx only renders when `!sameRange(period,
    // defaultPeriod())` must be present here.
    await expect(reset).toBeVisible();

    await reset.click();

    await expectRangeSpan(page, 30 * DAY_MS);
    // Having just been set to the default range, the button's own
    // condition for existing is now false -- it must disappear, not
    // merely stop doing anything.
    await expect(reset).toHaveCount(0);
  });
});

test.describe("SearchBox", () => {
  test("typing a term narrows the table to matching rows only", async ({ page }) => {
    await page.goto(`/#/state-board?${PINNED_QUERY}`);
    const tablePanel = page.locator('section.panel[aria-label="Every open change"]');
    await expect(tablePanel.getByRole("cell", { name: "kan-201-refactor-thing" })).toBeVisible();

    await page.getByLabel("Search").fill("widget");

    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-102-fix-widget-bug" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-201-refactor-thing" })).toHaveCount(0);
  });
});

test.describe("ProjectFilter", () => {
  test("typing a project scopes the dashboard bar's own project variable", async ({ page }) => {
    await page.goto(`/#/state-board?${PINNED_QUERY}`);
    const tablePanel = page.locator('section.panel[aria-label="Every open change"]');
    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toBeVisible();

    await page.getByLabel("Project filter").fill("uitest-beta");

    await expect(tablePanel.getByRole("cell", { name: "kan-201-refactor-thing" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toHaveCount(0);
    await expect(tablePanel.getByRole("cell", { name: "kan-102-fix-widget-bug" })).toHaveCount(0);
  });
});

test.describe("FilterBar", () => {
  test("choosing an exact-match column filter narrows the table to matching rows", async ({ page }) => {
    await page.goto(`/#/state-board?${PINNED_QUERY}`);
    const tablePanel = page.locator('section.panel[aria-label="Every open change"]');
    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toBeVisible();

    await page.getByLabel("Filter by State").selectOption("FINISHED");

    await expect(tablePanel.getByRole("cell", { name: "kan-201-refactor-thing" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toHaveCount(0);
    await expect(tablePanel.getByRole("cell", { name: "kan-102-fix-widget-bug" })).toHaveCount(0);
  });

  // StateBoard.tsx declares two filterable columns -- Project and State
  // (columns.ts:60-63) -- and the spec above only ever exercised State.
  // This is a distinct select from the dashboard bar's top-level
  // "Project filter" input (the ProjectFilter describe block, above):
  // that one scopes the server request itself, this one is an in-memory
  // exact-match filter over the page DataTable already holds
  // (FilterBar.tsx's own header comment).
  test("choosing the Project column filter narrows the table to matching rows", async ({ page }) => {
    await page.goto(`/#/state-board?${PINNED_QUERY}`);
    const tablePanel = page.locator('section.panel[aria-label="Every open change"]');
    await expect(tablePanel.getByRole("cell", { name: "kan-201-refactor-thing" })).toBeVisible();

    await page.getByLabel("Filter by Project").selectOption("uitest-alpha");

    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-102-fix-widget-bug" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-201-refactor-thing" })).toHaveCount(0);
  });
});

test.describe("DataTable sort buttons", () => {
  test("clicking a sortable header toggles ascending/descending order and its own arrow", async ({ page }) => {
    await page.goto(`/#/state-board?${PINNED_QUERY}`);
    const tablePanel = page.locator('section.panel[aria-label="Every open change"]');
    const changeHeader = tablePanel.getByRole("button", { name: /^Change/ });
    const changeCells = tablePanel.locator("tbody tr td:nth-child(2) a");
    await expect(changeCells).toHaveCount(3);

    await changeHeader.click();
    await expect(changeHeader).toHaveText("Change ▲");
    await expect(changeCells).toHaveText(["kan-101-add-widget", "kan-102-fix-widget-bug", "kan-201-refactor-thing"]);

    await changeHeader.click();
    await expect(changeHeader).toHaveText("Change ▼");
    await expect(changeCells).toHaveText(["kan-201-refactor-thing", "kan-102-fix-widget-bug", "kan-101-add-widget"]);
  });
});

test.describe("ChangeVariable", () => {
  test("selecting a change navigates to that change's own run-detail page", async ({ page }) => {
    await page.goto(`/#/state-board?${PINNED_QUERY}`);

    await page.getByLabel("Go to change").selectOption({ label: "uitest-beta/kan-201-refactor-thing" });

    await expect(page).toHaveURL(/#\/run\/uitest-beta\/kan-201-refactor-thing/);
    await expect(page.getByRole("heading", { name: "kan-201-refactor-thing" })).toBeVisible();
  });
});

test.describe("ModelVariable", () => {
  test("is disabled on the live state board, with its reason stated", async ({ page }) => {
    await page.goto(`/#/state-board?${PINNED_QUERY}`);
    const model = page.getByLabel("Model");
    await expect(model).toBeDisabled();
    const label = page.locator("label.model-variable");
    await expect(label).toHaveAttribute(
      "title",
      "The live state board's rows are changes, not stage runs, so a model restriction cannot apply here.",
    );
    await expect(label.locator(".dashboard-variable-hint")).toBeVisible();
  });

  test("is enabled with real models on the stage leaderboard", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
    const model = page.getByLabel("Model");
    await expect(model).toBeEnabled();
    await expect(model.locator("option")).toHaveText(["All models", "claude-opus-5", "claude-sonnet-5"]);
  });
});
