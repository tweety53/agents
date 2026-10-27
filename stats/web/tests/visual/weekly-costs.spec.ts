import { expect, test, type Page } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1 } from "./support";

// Weekly costs (KAN-1037, KAN-1038). The view is one chart of the last 8
// weeks, computed from the browser's own clock (WeeklyCosts.tsx's
// lastEightWeeks(new Date())), over every project and model, with no filter
// bar -- so the clock is pinned before the first navigation to an instant
// just after the UI-test fixture's own data (cmd/uitest-seed/seed.go's
// fixtureNow, 2026-08-15T09:00:00Z): the window is then 2026-06-24..2026-08-16
// on every run.
//
// The fixture's runs fall in the weeks of 2026-07-22 (Overall, Planning,
// Implementation, Self-review) and 2026-08-12 (Overall and Review), never in
// two consecutive weeks, so on its own it draws no week-over-week delta.
// withPreviousWeek adds one week, 2026-08-05, to the stack's real response:
// 2026-08-12's Overall then reads up on it and its Review down.
// The empty path pins the clock years before the fixture, so the panel has
// no recorded run and reads Panel's not-recorded banner.
const AFTER_FIXTURE = new Date("2026-08-16T12:00:00Z");
const BEFORE_FIXTURE = new Date("2020-01-01T12:00:00Z");
const SERIES = ["Overall", "Planning", "Implementation", "Review", "Self-review"];

test.use({ viewport: { width: 1280, height: 1024 } });

async function withPreviousWeek(page: Page) {
  await page.route("**/api/v1/stats/weekly-costs?*", async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    body.rows = [
      ...body.rows,
      { week: "2026-08-05", overallUsd: 1.5, planningUsd: 0.3, implementationUsd: 0.6, reviewUsd: 0.6, selfReviewUsd: null },
    ].sort((a: { week: string }, b: { week: string }) => a.week.localeCompare(b.week));
    await route.fulfill({ response, json: body });
  });
}

async function open(page: Page, time: Date) {
  await page.clock.install({ time });
  await page.goto("/#/weekly-costs");
  await expect(page.getByRole("heading", { name: "Weekly costs" })).toBeVisible();
}

test.describe("weekly costs", () => {
  test("draws one chart with delta labels in the dark palette", async ({ page }) => {
    await withPreviousWeek(page);
    await open(page, AFTER_FIXTURE);
    await expect(page.getByRole("link", { name: "Weekly costs", exact: true })).toHaveAttribute("aria-current", "page");
    await expect(page.getByRole("toolbar", { name: "Dashboard controls" })).toHaveCount(0);
    for (const name of SERIES) {
      await expect(page.getByRole("button", { name, exact: true })).toHaveAttribute("aria-pressed", "true");
    }
    await expect(page.getByText("Week starting (Wed, UTC)")).toBeVisible();
    await expect(page.getByText("Cost (USD)")).toBeVisible();
    const delta = (series: string) => page.locator(`[data-testid="delta-label"][data-series="${series}"][data-week="2026-08-12"]`);
    await expect(delta("Overall")).toHaveClass("delta-up");
    await expect(delta("Review")).toHaveClass("delta-down");
    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("weekly-costs.png", { fullPage: true });
  });

  test("draws the series and delta tokens in the light palette", async ({ page }) => {
    await page.emulateMedia({ colorScheme: "light" });
    await withPreviousWeek(page);
    await open(page, AFTER_FIXTURE);
    await expect(page.locator(".delta-up, .delta-down")).not.toHaveCount(0);
    expect(await bodyBackground(page)).not.toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("weekly-costs-light.png", { fullPage: true });
  });

  test("renders the empty state before any recorded run", async ({ page }) => {
    await open(page, BEFORE_FIXTURE);
    await expect(page.getByTestId("not-recorded")).toHaveCount(1);
    await expect(page.getByTestId("time-series-point")).toHaveCount(0);

    await expect(page).toHaveScreenshot("weekly-costs-empty.png", { fullPage: true });
  });

  test("a toggled-off series leaves the chart", async ({ page }) => {
    await open(page, AFTER_FIXTURE);
    const overall = page.locator('[data-testid="time-series-point"][data-series="Overall"]');
    // The stack's own response: Overall in its two weeks, six points in all.
    await expect(overall).toHaveCount(2);
    await expect(page.getByTestId("time-series-point")).toHaveCount(6);
    await page.getByRole("button", { name: "Overall", exact: true }).click();
    await expect(page.getByRole("button", { name: "Overall", exact: true })).toHaveAttribute("aria-pressed", "false");
    await expect(overall).toHaveCount(0);
    // Park the pointer so no hover styling rides the baseline.
    await page.mouse.move(0, 0);

    await expect(page).toHaveScreenshot("weekly-costs-toggled.png", { fullPage: true });
  });

  test("hovering a point shows its tooltip", async ({ page }) => {
    await withPreviousWeek(page);
    await open(page, AFTER_FIXTURE);
    await expect(page.getByRole("tooltip")).toHaveCount(0);
    await page.locator('[data-testid="time-series-point"][data-series="Overall"][data-week="2026-08-12"]').hover();
    const tip = page.getByRole("tooltip");
    await expect(tip).toContainText("Overall");
    await expect(tip).toContainText("$3.37");
    await expect(tip).toContainText("▲ 124.5%");

    await expect(page).toHaveScreenshot("weekly-costs-tooltip.png", { fullPage: true });
  });
});
