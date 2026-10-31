import { expect, test, type Page } from "@playwright/test";
import { FIXTURE_NOW } from "./support";

// Captures of the reworked Trend over time view (KAN-1051, decision
// trend-capture-spec): the default view at 1280px, the same at 390px, and the
// chart's week-hover state -- every one full page. The view's Period chips are
// a rolling window ending now, so the clock is pinned to the UI-test fixture's
// FIXTURE_NOW (its timestamps are fixed at seed time).
//
// Values below are read from the live UI-test stack rather than assumed:
// GET /api/v1/stats/trend?from=2026-05-27T00:00:00.000Z&to=2026-08-15T09:00:00.000Z
// returns 19 finished changes, 16 of them /flow; the latest filled /flow week
// is 2026-07-22 (one uitest-alpha change, $1.50), so the hover lands there.
// The default /flow tiles read Median per change $9.75 (−$23.75, −71% vs the
// 6 weeks before), Changes finished 16 ($511 spent on them), Flow changes
// landed 13.

const HEADERS = ["Week", "Changes", "Median", "vs previous week", "Mean", "Spend", "Flow changes"];

test.use({ colorScheme: "light" });

async function openAndAssert(page: Page) {
  await page.clock.setFixedTime(FIXTURE_NOW);
  await page.goto("/#/trend");
  await expect(page.getByText("Loading…")).toHaveCount(0);

  await expect(page.getByRole("link", { name: "Trend over time" })).toHaveAttribute("aria-current", "page");
  const bar = page.getByRole("toolbar", { name: "Filters" });
  for (const g of ["Period", "Command", "Project"]) await expect(bar.getByRole("group", { name: g })).toBeVisible();
  await expect(bar.getByRole("group", { name: "Period" }).getByRole("button", { name: "12w" })).toHaveAttribute("aria-pressed", "true");
  await expect(bar.getByRole("group", { name: "Command" }).getByRole("button", { name: /^\/flow\d/ })).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByLabel("Period start")).toHaveCount(0);
  await expect(page.getByRole("combobox", { name: "Model" })).toHaveCount(0);

  for (const t of ["Median per change", "Changes finished", "Flow changes landed"]) await expect(page.locator(".tr-tile .lb-lbl", { hasText: new RegExp(`^${t}$`) })).toBeVisible();

  await expect(page.getByRole("table").getByRole("row").first().getByRole("columnheader")).toHaveText(HEADERS);
}

test("trend at rest", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await expect(page).toHaveScreenshot("trend-wide.png", { fullPage: true });
});

test("trend on a phone", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await openAndAssert(page);
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
  await expect(page).toHaveScreenshot("trend-phone.png", { fullPage: true });
});

test("trend chart hover", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await page.getByTestId("trend-week-2026-07-22").hover();
  await expect(page.locator(".tr-tip")).toBeVisible();
  await expect(page).toHaveScreenshot("trend-chart-hover.png", { fullPage: true });
});
