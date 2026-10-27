import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, EMPTY_QUERY, PINNED_QUERY } from "./support";

// Weekly costs (KAN-1037). The "Last 8 weeks" table ignores the period
// picker and is computed from the browser's own clock (WeeklyCosts.tsx's
// lastEightWeeks(new Date())), so the clock is pinned before the first
// navigation to an instant just after the UI-test fixture's own data
// (cmd/uitest-seed/seed.go's fixtureNow, 2026-08-15T09:00:00Z): the table's
// window is then 2026-06-24..2026-08-16 on every run. The chart follows the
// period, pinned by PINNED_QUERY for the reason support.ts gives.
//
// The fixture's costed runs all fall in the week of 2026-08-12, so the
// populated path draws one week, measured for Overall and Review only
// (the fixture has no flow.* planning, sdd-tdd or self-review stage), so
// the chart shows two points and the table three Unavailable cells; the empty path pins the
// clock and the period years before the fixture so both panels render
// their empty states.
const AFTER_FIXTURE = new Date("2026-08-16T12:00:00Z");
const BEFORE_FIXTURE = new Date("2020-01-01T12:00:00Z");

test.use({ viewport: { width: 1280, height: 1024 } });

test.describe("weekly costs", () => {
  test("renders the 8-week table and the five-line chart in the dark palette", async ({ page }) => {
    await page.clock.install({ time: AFTER_FIXTURE });
    await page.goto(`/#/weekly-costs?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "Weekly costs" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Weekly costs", exact: true })).toHaveAttribute("aria-current", "page");

    const table = page.locator('section.panel[aria-label="Last 8 weeks"]');
    await expect(table.getByRole("cell", { name: "2026-08-12", exact: true })).toBeVisible();
    const chart = page.locator('section.panel[aria-label="Weekly cost"]');
    await expect(chart.getByTestId("time-series-point")).toHaveCount(2);
    for (const name of ["Overall", "Planning", "Implementation", "Review", "Self-review"]) {
      await expect(chart.getByText(name, { exact: true })).toBeVisible();
    }
    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("weekly-costs.png", { fullPage: true });
  });

  test("draws the series tokens in the light palette", async ({ page }) => {
    await page.emulateMedia({ colorScheme: "light" });
    await page.clock.install({ time: AFTER_FIXTURE });
    await page.goto(`/#/weekly-costs?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "Weekly costs" })).toBeVisible();
    const chart = page.locator('section.panel[aria-label="Weekly cost"]');
    await expect(chart.getByTestId("time-series-point")).toHaveCount(2);
    expect(await bodyBackground(page)).not.toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("weekly-costs-light.png", { fullPage: true });
  });

  test("renders both empty states before any recorded run", async ({ page }) => {
    await page.clock.install({ time: BEFORE_FIXTURE });
    await page.goto(`/#/weekly-costs?${EMPTY_QUERY}`);
    await expect(page.getByRole("heading", { name: "Weekly costs" })).toBeVisible();
    await expect(page.locator('section.panel[aria-label="Last 8 weeks"]')).toBeVisible();
    await expect(page.getByTestId("time-series-point")).toHaveCount(0);

    await expect(page).toHaveScreenshot("weekly-costs-empty.png", { fullPage: true });
  });

  test("the table keeps its 8 weeks when the period empties the chart", async ({ page }) => {
    await page.clock.install({ time: AFTER_FIXTURE });
    await page.goto(`/#/weekly-costs?${PINNED_QUERY}`);
    const chart = page.locator('section.panel[aria-label="Weekly cost"]');
    await expect(chart.getByTestId("time-series-point")).toHaveCount(2);
    await page.goto(`/#/weekly-costs?${EMPTY_QUERY}`);
    await expect(chart.getByText("No data was recorded for this period.")).toBeVisible();
    await expect(chart.getByTestId("time-series-point")).toHaveCount(0);
    const table = page.locator('section.panel[aria-label="Last 8 weeks"]');
    await expect(table.getByRole("cell", { name: "$3.37", exact: true })).toBeVisible();
  });
});
