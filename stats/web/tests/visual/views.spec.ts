import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, PINNED_QUERY } from "./support";

// baseline.spec.ts already covers the dashboard (state-board) and
// run-detail views -- the two KAN-16 broke; the stage leaderboard has its
// own spec. This file covers the trend view, which
// task 12 requires ("every view reachable"), in baseline.spec.ts's own
// shape: assert the data actually loaded (no error, no stuck "loading",
// the fixture's own rows visible), assert the dark palette actually
// applied, assert no stat-panel label is double-rendered (KAN-16's third
// defect class), then toHaveScreenshot -- a screenshot alone proves the
// page painted, not that it is right, so every one of these assertions
// comes first.
//
// Values below are read from the live UI-test stack rather than assumed:
// GET /api/v1/stats/trend?<PINNED_QUERY> returns two days (2026-08-14,
// 2026-08-15).

test.describe("trend", () => {
  test("renders the daily cost series with data loaded and the dark palette applied", async ({ page }) => {
    await page.goto(`/#/trend?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "Trend over time" })).toBeVisible();

    const tablePanel = page.locator('section.panel[aria-label="By day"]');
    await expect(tablePanel.getByRole("cell", { name: "2026-08-14", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "2026-08-15", exact: true })).toBeVisible();

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    const daysPanel = page.locator('section.panel[aria-label="Days"]');
    await expect(daysPanel.getByText("Days", { exact: true })).toHaveCount(1);
    const costPanel = page.locator('section.panel[aria-label="Total cost"]');
    await expect(costPanel.getByText("Total cost", { exact: true })).toHaveCount(1);

    await expect(page).toHaveScreenshot("trend.png", { fullPage: true });
  });

  // Trend.tsx declares its own three sortable columns (day, runCount,
  // totalCostUsd), no filterable one -- see this file's own header
  // comment on why this is worth testing even though the shared sort
  // mechanism is already proven elsewhere. All three differ between the
  // fixture's own two days (this file's own header comment); since
  // seedInteractive (KAN-1061) 2026-08-15 carries both more runs (52) and
  // more cost (6.607) than 2026-08-14 (1 run, 2.75), so totalCostUsd's order
  // now agrees with day's and runCount's.
  test.describe("sort", () => {
    const tablePanelSelector = 'section.panel[aria-label="By day"]';

    test("sorting by Day orders chronologically", async ({ page }) => {
      await page.goto(`/#/trend?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Day/ });
      const dayCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Day ▲");
      await expect(dayCells).toHaveText(["2026-08-14", "2026-08-15"]);

      await header.click();
      await expect(header).toHaveText("Day ▼");
      await expect(dayCells).toHaveText(["2026-08-15", "2026-08-14"]);
    });

    test("sorting by Runs orders by run count, agreeing with Day here", async ({ page }) => {
      await page.goto(`/#/trend?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Runs/ });
      const dayCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Runs ▲");
      // runCount: 2026-08-14 has 1, 2026-08-15 has 2.
      await expect(dayCells).toHaveText(["2026-08-14", "2026-08-15"]);
    });

    test("sorting by Total cost orders by cost", async ({ page }) => {
      await page.goto(`/#/trend?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Total cost/ });
      const dayCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Total cost ▲");
      // totalCostUsd: 2026-08-14 2.75 (smaller), 2026-08-15 6.607.
      await expect(dayCells).toHaveText(["2026-08-14", "2026-08-15"]);
    });
  });
});
