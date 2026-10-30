import { expect, type Page, test } from "@playwright/test";
import { EMPTY_QUERY, FIXTURE_NOW, PINNED_FROM_INPUT, PINNED_QUERY, PINNED_TO_INPUT } from "./support";

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
// The stage leaderboard's table is the shared target for the dashboard
// bar's controls and DataTable's: the live state board carries its own chip
// filters instead (kan-1048, exercised under "Live state board chips"
// below). PINNED_QUERY/EMPTY_QUERY give a deterministic "populated" and
// "empty" starting point; All time is the one PeriodPicker preset safe to
// assert fixture rows against, since it always spans the fixture's fixed
// 2026-08-15 data.
//
// Leaderboard rows in PINNED_QUERY's window, its ten costliest: /flow "SDD +
// TDD per task" and "flow.review-panel" (uitest-alpha) and the eight
// /flow-interactive stages seedInteractive's sessions ran (KAN-1061) --
// /flow-fast "integrate" (uitest-beta) is the eleventh and falls off the
// table unless a filter narrows it.
const LEADERBOARD = 'section.panel[aria-label="Every costed stage"]';

test.describe("PeriodPicker", () => {
  test("the All time preset broadens the table from empty to the fixture's own rows", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${EMPTY_QUERY}`);
    const tablePanel = page.locator(LEADERBOARD);
    await expect(tablePanel.getByTestId("not-recorded")).toBeVisible();

    const allTime = page.getByRole("button", { name: "All time" });
    await allTime.click();
    await expect(allTime).toHaveAttribute("aria-pressed", "true");

    await expect(tablePanel.getByRole("cell", { name: "flow.review-panel", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "flow.landing-routes", exact: true })).toBeVisible();
  });

  test("typing explicit From/To values reloads the table to that window", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${EMPTY_QUERY}`);
    const tablePanel = page.locator(LEADERBOARD);
    await expect(tablePanel.getByTestId("not-recorded")).toBeVisible();

    await page.getByLabel("Period start").fill(PINNED_FROM_INPUT);
    await page.getByLabel("Period end").fill(PINNED_TO_INPUT);

    await expect(tablePanel.getByRole("cell", { name: "flow.review-panel", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "flow.landing-routes", exact: true })).toBeVisible();
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
    await page.goto(`/#/stage-leaderboard?${EMPTY_QUERY}`);
    await page.getByRole("button", { name: "Today", exact: true }).click();
    await expectOnlyPressed(page, "Today");

    const { fromMs, toMs } = await periodInputRangeMs(page);
    const expectedMidnight = new Date();
    expectedMidnight.setHours(0, 0, 0, 0);
    expect(Math.abs(fromMs - expectedMidnight.getTime())).toBeLessThan(INPUT_MINUTE_TRUNCATION_TOLERANCE_MS);
    expect(Math.abs(toMs - Date.now())).toBeLessThan(INPUT_MINUTE_TRUNCATION_TOLERANCE_MS);
  });

  test("7 days sets a seven-day range ending now, and presses only itself", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${EMPTY_QUERY}`);
    await page.getByRole("button", { name: "7 days", exact: true }).click();
    await expectOnlyPressed(page, "7 days");
    await expectRangeSpan(page, 7 * DAY_MS);
  });

  test("30 days sets the default thirty-day range, and presses only itself", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${EMPTY_QUERY}`);
    await page.getByRole("button", { name: "30 days", exact: true }).click();
    await expectOnlyPressed(page, "30 days");
    await expectRangeSpan(page, 30 * DAY_MS);
  });

  test("90 days sets a ninety-day range ending now, and presses only itself", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${EMPTY_QUERY}`);
    await page.getByRole("button", { name: "90 days", exact: true }).click();
    await expectOnlyPressed(page, "90 days");
    await expectRangeSpan(page, 90 * DAY_MS);
  });

  test("Reset to default only renders when the period differs from default, and resets it", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${EMPTY_QUERY}`);
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
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
    const tablePanel = page.locator(LEADERBOARD);
    await expect(tablePanel.getByRole("cell", { name: "flow.landing-routes", exact: true })).toBeVisible();

    await page.getByLabel("Search").fill("landing");

    await expect(tablePanel.getByRole("cell", { name: "flow.landing-routes", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "flow.review-panel", exact: true })).toHaveCount(0);
  });
});

test.describe("ProjectFilter", () => {
  test("typing a project scopes the dashboard bar's own project variable", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
    const tablePanel = page.locator(LEADERBOARD);
    await expect(tablePanel.getByRole("cell", { name: "flow.review-panel", exact: true })).toBeVisible();

    await page.getByLabel("Project filter").fill("uitest-beta");

    await expect(tablePanel.getByRole("cell", { name: "integrate", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "flow.review-panel", exact: true })).toHaveCount(0);
  });
});

test.describe("FilterBar", () => {
  test("choosing an exact-match column filter narrows the table to matching rows", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
    const tablePanel = page.locator(LEADERBOARD);
    await expect(tablePanel.getByRole("cell", { name: "flow.review-panel", exact: true })).toBeVisible();

    await page.getByLabel("Filter by Command").selectOption("/flow-interactive");

    await expect(tablePanel.getByRole("cell", { name: "flow.landing-routes", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "flow.review-panel", exact: true })).toHaveCount(0);
  });
});

test.describe("DataTable sort buttons", () => {
  test("clicking a sortable header toggles ascending/descending order and its own arrow", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
    const tablePanel = page.locator(LEADERBOARD);
    const stageHeader = tablePanel.getByRole("button", { name: /^Stage/ });
    const stageCells = tablePanel.locator("tbody tr td:nth-child(2)");
    await expect(stageCells).toHaveCount(10);
    const ascending = [
      "flow.brainstorm", "flow.kickoff", "flow.landing-routes", "flow.load-context", "flow.refresh-main-checkout",
      "flow.review-panel", "flow.sdd-tdd", "flow.verify", "flow.visual-verify", "SDD + TDD per task",
    ];

    await stageHeader.click();
    await expect(stageHeader).toHaveText("Stage ▲");
    await expect(stageCells).toHaveText(ascending);

    await stageHeader.click();
    await expect(stageHeader).toHaveText("Stage ▼");
    await expect(stageCells).toHaveText([...ascending].reverse());
  });
});

test.describe("ChangeVariable", () => {
  test("selecting a change navigates to that change's own run-detail page", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);

    await page.getByLabel("Go to change").selectOption({ label: "uitest-beta/kan-201-refactor-thing" });

    await expect(page).toHaveURL(/#\/run\/uitest-beta\/kan-201-refactor-thing/);
    await expect(page.getByRole("heading", { name: "kan-201-refactor-thing" })).toBeVisible();
  });
});

test.describe("ModelVariable", () => {
  test("is disabled on flow health, with its reason stated", async ({ page }) => {
    await page.goto(`/#/flow-health?${PINNED_QUERY}`);
    const model = page.getByLabel("Model");
    await expect(model).toBeDisabled();
    const label = page.locator("label.model-variable");
    await expect(label).toHaveAttribute(
      "title",
      "Flow health totals changes, panel rounds and guard fires across every model, so a model restriction cannot apply here.",
    );
    await expect(label.locator(".dashboard-variable-hint")).toBeVisible();
  });

  test("is enabled with real models on the stage leaderboard", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
    const model = page.getByLabel("Model");
    await expect(model).toBeEnabled();
    await expect(model.locator("option")).toHaveText(["All models", "claude-opus-5", "claude-opus-5-5", "claude-sonnet-5"]);
  });
});

// The live state board's own filters (kan-1048): fixture rows inside the
// default 7d at FIXTURE_NOW are kan-101-add-widget (STARTED, running),
// kan-102-fix-widget-bug (IN_PROGRESS, running) and kan-201-refactor-thing
// (FINISHED, its group collapsed).
test.describe("Live state board chips", () => {
  test.beforeEach(async ({ page }) => {
    await page.clock.setFixedTime(FIXTURE_NOW);
    await page.goto("/#/state-board");
  });

  const table = (page: Page) => page.getByRole("region", { name: "Changes" }).getByRole("table");
  const tile = (page: Page, name: RegExp) => page.getByRole("button", { name });

  test("a Project chip narrows the rows and the state tiles", async ({ page }) => {
    await expect(tile(page, /^Integrating & review/)).toContainText("1");
    await page.getByRole("group", { name: "Project" }).getByRole("button", { name: /^uitest-beta/ }).click();
    await expect(tile(page, /^Integrating & review/)).toContainText("0");
    await expect(tile(page, /^Finished/)).toContainText("1");
    await expect(table(page).getByRole("link", { name: "kan-101-add-widget" })).toHaveCount(0);
  });

  test("a state tile filters by its state and opens its collapsed group", async ({ page }) => {
    await expect(table(page).getByRole("link", { name: "kan-201-refactor-thing" })).toHaveCount(0);
    await tile(page, /^Finished/).click();
    await expect(tile(page, /^Finished/)).toHaveAttribute("aria-pressed", "true");
    await expect(table(page).getByRole("link", { name: "kan-201-refactor-thing" })).toBeVisible();
    await expect(table(page).getByRole("link", { name: "kan-101-add-widget" })).toHaveCount(0);
  });

  test("search narrows the rows by change name", async ({ page }) => {
    await page.getByRole("searchbox", { name: "Search changes" }).fill("fix-widget");
    await expect(table(page).getByRole("link", { name: "kan-102-fix-widget-bug" })).toBeVisible();
    await expect(table(page).getByRole("link", { name: "kan-101-add-widget" })).toHaveCount(0);
  });

  test("Updated All widens past seven days, and Clear filters restores the default", async ({ page }) => {
    const finished = await tile(page, /^Finished/).textContent();
    await page.getByRole("group", { name: "Updated" }).getByRole("button", { name: "All" }).click();
    await expect(tile(page, /^Finished/)).not.toHaveText(finished!);
    await page.getByRole("button", { name: "Clear filters" }).click();
    await expect(tile(page, /^Finished/)).toHaveText(finished!);
    await expect(page.getByRole("button", { name: "Clear filters" })).toHaveCount(0);
  });
});
