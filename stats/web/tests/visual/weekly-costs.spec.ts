import { expect, test, type Page } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1 } from "./support";

// Weekly costs (KAN-1037, KAN-1038, KAN-1041). The view is two charts of the
// last 8 weeks -- cost by phase, and cost by model · effort -- computed from
// the browser's own clock (WeeklyCosts.tsx's
// lastEightWeeks(new Date())), over every project and model, with no filter
// bar -- so the clock is pinned before the first navigation to an instant
// just after the UI-test fixture's own data (cmd/uitest-seed/seed.go's
// fixtureNow, 2026-08-15T09:00:00Z): the window is then 2026-06-24..2026-08-16
// on every run.
//
// The fixture's runs fall in the weeks of 2026-07-15 (Overall,
// Implementation), 2026-07-22 (Overall, Planning, Implementation,
// Self-review, Visual verify, Other) and 2026-08-12 (Overall, Review and
// Other).
// The first two carry priced dispatches, so the model·effort chart draws
// lines with deltas across them. withPreviousWeek adds one week,
// 2026-08-05, to the stack's real phase response: 2026-08-12's Overall then
// reads up on it and its Review down.
// The empty path pins the clock years before the fixture, so the panel has
// no recorded run and reads Panel's not-recorded banner.
const AFTER_FIXTURE = new Date("2026-08-16T12:00:00Z");
const BEFORE_FIXTURE = new Date("2020-01-01T12:00:00Z");
const SERIES = ["Overall", "Planning", "Implementation", "Review", "Self-review", "Visual verify", "Other"];

test.use({ viewport: { width: 1280, height: 1024 } });

async function withPreviousWeek(page: Page) {
  await page.route("**/api/v1/stats/weekly-costs?*", async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    body.rows = [
      ...body.rows,
      {
        week: "2026-08-05",
        overallUsd: 1.5,
        planningUsd: 0.3,
        implementationUsd: 0.6,
        reviewUsd: 0.6,
        selfReviewUsd: null,
        visualVerifyUsd: null,
        otherUsd: null,
      },
    ].sort((a: { week: string }, b: { week: string }) => a.week.localeCompare(b.week));
    await route.fulfill({ response, json: body });
  });
}

async function open(page: Page, time: Date) {
  await page.clock.install({ time });
  await page.goto("/#/weekly-costs");
  await expect(page.getByRole("heading", { name: "Weekly costs" })).toBeVisible();
}

// The phase chart's panel, so a locator never reads the model·effort chart.
const phases = (page: Page) => page.getByRole("region", { name: "Cost by phase" });
const modelEffort = (page: Page) => page.getByRole("region", { name: "Cost by model · effort" });

// Overall alone is shown at first; this shows the phases too.
async function showPhases(page: Page) {
  for (const name of SERIES.slice(1)) await phases(page).getByRole("button", { name, exact: true }).click();
  await page.mouse.move(0, 0);
}

test.describe("weekly costs", () => {
  test("draws one chart with delta labels in the dark palette", async ({ page }) => {
    await withPreviousWeek(page);
    await open(page, AFTER_FIXTURE);
    await expect(page.getByRole("link", { name: "Weekly costs", exact: true })).toHaveAttribute("aria-current", "page");
    await expect(page.getByRole("toolbar", { name: "Dashboard controls" })).toHaveCount(0);
    await expect(phases(page).getByRole("button")).toHaveText(SERIES);
    for (const name of SERIES) {
      await expect(phases(page).getByRole("button", { name, exact: true })).toHaveAttribute(
        "aria-pressed",
        String(name === "Overall"),
      );
    }
    await showPhases(page);
    await expect(phases(page).getByText("Week starting (Wed, UTC)")).toBeVisible();
    await expect(phases(page).getByText("Cost (USD)")).toBeVisible();
    const delta = (series: string) =>
      page.locator(`[data-testid="delta-label"][data-series="${series}"][data-week="2026-08-12"] tspan`);
    await expect(delta("Overall")).toHaveClass("delta-up");
    await expect(delta("Review")).toHaveClass("delta-down");
    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("weekly-costs.png", { fullPage: true });
  });

  test("draws the series and delta tokens in the light palette", async ({ page }) => {
    await page.emulateMedia({ colorScheme: "light" });
    await withPreviousWeek(page);
    await open(page, AFTER_FIXTURE);
    await showPhases(page);
    await expect(page.locator(".delta-up, .delta-down")).not.toHaveCount(0);
    expect(await bodyBackground(page)).not.toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("weekly-costs-light.png", { fullPage: true });
  });

  test("renders the empty state before any recorded run", async ({ page }) => {
    await open(page, BEFORE_FIXTURE);
    await expect(page.getByTestId("not-recorded")).toHaveCount(2);
    await expect(page.getByTestId("time-series-point")).toHaveCount(0);

    await expect(page).toHaveScreenshot("weekly-costs-empty.png", { fullPage: true });
  });

  test("a toggled-off series leaves the chart", async ({ page }) => {
    await open(page, AFTER_FIXTURE);
    await showPhases(page);
    const overall = phases(page).locator('[data-testid="time-series-point"][data-series="Overall"]');
    // The stack's own response: Overall in its three weeks, eleven points in all.
    await expect(overall).toHaveCount(3);
    await expect(phases(page).getByTestId("time-series-point")).toHaveCount(11);
    await phases(page).getByRole("button", { name: "Overall", exact: true }).click();
    await expect(phases(page).getByRole("button", { name: "Overall", exact: true })).toHaveAttribute(
      "aria-pressed",
      "false",
    );
    await expect(overall).toHaveCount(0);
    // Park the pointer so no hover styling rides the baseline.
    await page.mouse.move(0, 0);

    await expect(page).toHaveScreenshot("weekly-costs-toggled.png", { fullPage: true });
  });

  test("hovering a point shows its tooltip", async ({ page }) => {
    await withPreviousWeek(page);
    await open(page, AFTER_FIXTURE);
    await expect(page.getByRole("tooltip")).toHaveCount(0);
    await phases(page).locator('[data-testid="time-series-point"][data-series="Overall"][data-week="2026-08-12"]').hover();
    const tip = page.getByRole("tooltip");
    await expect(tip).toContainText("Overall");
    await expect(tip).toContainText("$3.37");
    await expect(tip).toContainText("▲ 124.5%");
    await expect(tip).toContainText("vs prev week");

    await expect(page).toHaveScreenshot("weekly-costs-tooltip.png", { fullPage: true });
  });

  test("draws the model·effort chart with cost labels", async ({ page }) => {
    await open(page, AFTER_FIXTURE);
    await expect(modelEffort(page).getByRole("button")).toHaveText([
      "opus-5 · session",
      "sonnet-5 · session",
      "opus-5 · low",
      "opus-5 · medium",
    ]);
    await expect(modelEffort(page).getByTestId("weekly-line")).toHaveCount(4);
    const pill = modelEffort(page).locator('[data-testid="delta-label"][data-series="opus-5 · low"][data-week="2026-07-22"]');
    await expect(pill).toHaveText("$0.30 ▲20%");
    await expect(modelEffort(page).getByTestId("delta-label").first()).toContainText("$");
    await page.mouse.move(0, 0);

    await expect(page).toHaveScreenshot("weekly-costs-model-effort.png", { fullPage: true });
  });
});
