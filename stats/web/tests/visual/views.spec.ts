import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, PINNED_QUERY } from "./support";

// baseline.spec.ts already covers the dashboard (state-board) and
// run-detail views -- the two KAN-16 broke. This file covers two more
// static views task 12 requires ("every view reachable"):
// stage-leaderboard and trend. Each spec follows
// baseline.spec.ts's own shape: assert the data actually loaded (no
// error, no stuck "loading", the fixture's own rows visible), assert the
// dark palette actually applied, assert no stat-panel label is
// double-rendered (KAN-16's third defect class), then toHaveScreenshot --
// a screenshot alone proves the page painted, not that it is right, so
// every one of these assertions comes first.
//
// Values below are read from the live UI-test stack rather than assumed:
// GET /api/v1/stats/stage-leaderboard?<PINNED_QUERY> returns three rows
// (/flow "SDD + TDD per task", runCount 2; /flow "flow.review-panel",
// runCount 1; /flow-fast "integrate", runCount 1); GET /api/v1/stats/trend?<PINNED_QUERY>
// returns two days (2026-08-14, 2026-08-15).

test.describe("stage leaderboard", () => {
  test("renders every costed stage with data loaded and the dark palette applied", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "Stage leaderboard" })).toBeVisible();

    const tablePanel = page.locator('section.panel[aria-label="Every costed stage"]');
    await expect(tablePanel.getByRole("cell", { name: "/flow", exact: true })).toHaveCount(2);
    await expect(tablePanel.getByRole("cell", { name: "/flow-fast", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "SDD + TDD per task", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "flow.review-panel", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "integrate", exact: true })).toBeVisible();

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    const stagesPanel = page.locator('section.panel[aria-label="Stages"]');
    await expect(stagesPanel.getByText("Stages", { exact: true })).toHaveCount(1);
    const runsPanel = page.locator('section.panel[aria-label="Total runs"]');
    await expect(runsPanel.getByText("Total runs", { exact: true })).toHaveCount(1);

    await expect(page).toHaveScreenshot("stage-leaderboard.png", { fullPage: true });
  });

  // StageLeaderboard.tsx declares its own six sortable columns and its own
  // one filterable column (command) on top of the shared DataTable/
  // FilterBar components baseline.spec.ts's StateBoard spec and
  // run-detail.spec.ts's StageRunTable specs already prove correct in
  // general -- what is untested here is this *view's own* column
  // definitions (its own accessor per column), which is exactly the class
  // of bug a shared, already-correct component cannot catch: a column
  // pointed at the wrong field ties or reorders wrong while the sort
  // mechanism itself keeps working.
  //
  // The fixture's own three rows (GET
  // /api/v1/stats/stage-leaderboard?<PINNED_QUERY>, this file's own
  // header comment) differ on every column but two: Command ties the two
  // /flow rows and Runs ties the two single-run rows, so those two
  // assertions read only what the tie leaves determined. Order assertions
  // read the Stage column, the one value unique to each row.
  test.describe("sort and filter", () => {
    const tablePanelSelector = 'section.panel[aria-label="Every costed stage"]';

    test("sorting by Command orders alphabetically", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Command/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Command ▲");
      await expect(commandCells).toHaveText(["/flow", "/flow", "/flow-fast"]);

      await header.click();
      await expect(header).toHaveText("Command ▼");
      await expect(commandCells).toHaveText(["/flow-fast", "/flow", "/flow"]);
    });

    test("sorting by Stage orders by its own text, not Command's", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Stage/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");
      const stageCells = tablePanel.locator("table tbody tr td:nth-child(2)");

      await header.click();
      await expect(header).toHaveText("Stage ▲");
      // "flow.review-panel" < "integrate" < "SDD + TDD per task"
      // (localeCompare) -- which puts /flow-fast between the two /flow
      // rows, an order Command's own sort above never produces: what
      // proves this column's accessor reads its own field rather than
      // Command's.
      await expect(stageCells).toHaveText(["flow.review-panel", "integrate", "SDD + TDD per task"]);
      await expect(commandCells).toHaveText(["/flow", "/flow-fast", "/flow"]);
    });

    test("sorting by Runs orders by run count", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Runs/ });
      const stageCells = tablePanel.locator("table tbody tr td:nth-child(2)");

      await header.click();
      await expect(header).toHaveText("Runs ▲");
      // runCount: integrate 1, flow.review-panel 1 (a tie), SDD + TDD per task 2.
      await expect(stageCells.last()).toHaveText("SDD + TDD per task");
    });

    test("sorting by Mean cost orders by mean cost", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Mean cost/ });
      const stageCells = tablePanel.locator("table tbody tr td:nth-child(2)");

      await header.click();
      await expect(header).toHaveText("Mean cost ▲");
      // meanCostUsd: integrate 0.028, flow.review-panel 0.41, SDD + TDD per task 1.4645.
      await expect(stageCells).toHaveText(["integrate", "flow.review-panel", "SDD + TDD per task"]);
    });

    test("sorting by Median cost orders by median cost", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Median cost/ });
      const stageCells = tablePanel.locator("table tbody tr td:nth-child(2)");

      await header.click();
      await expect(header).toHaveText("Median cost ▲");
      // medianCostUsd: integrate 0.028, flow.review-panel 0.41, SDD + TDD per task ~1.4645.
      await expect(stageCells).toHaveText(["integrate", "flow.review-panel", "SDD + TDD per task"]);
    });

    test("sorting by P90 cost orders by p90 cost", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^P90 cost/ });
      const stageCells = tablePanel.locator("table tbody tr td:nth-child(2)");

      await header.click();
      await expect(header).toHaveText("P90 cost ▲");
      // p90CostUsd: integrate 0.028, flow.review-panel 0.41, SDD + TDD per task 2.4929.
      await expect(stageCells).toHaveText(["integrate", "flow.review-panel", "SDD + TDD per task"]);
    });

    test("Filter by Command narrows to the matching stage", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      await expect(tablePanel.getByRole("cell", { name: "/flow", exact: true })).toHaveCount(2);

      await page.getByLabel("Filter by Command").selectOption("/flow-fast");

      await expect(tablePanel.getByRole("cell", { name: "/flow-fast", exact: true })).toBeVisible();
      await expect(tablePanel.getByRole("cell", { name: "/flow", exact: true })).toHaveCount(0);
    });
  });
});

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
  // fixture's own two days (this file's own header comment), and
  // totalCostUsd's own order is the *reverse* of day/runCount's --
  // 2026-08-14 costs more (2.75) despite having fewer runs (1) than
  // 2026-08-15 (0.207, 2 runs) -- which is what proves this column reads
  // its own field rather than reusing day's or runCount's.
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

    test("sorting by Total cost orders by cost, the reverse of Day/Runs", async ({ page }) => {
      await page.goto(`/#/trend?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Total cost/ });
      const dayCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Total cost ▲");
      // totalCostUsd: 2026-08-15 0.207 (smaller), 2026-08-14 2.75.
      await expect(dayCells).toHaveText(["2026-08-15", "2026-08-14"]);
    });
  });
});
