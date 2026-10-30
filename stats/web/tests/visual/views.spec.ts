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
// GET /api/v1/stats/stage-leaderboard?<PINNED_QUERY> returns its ten
// costliest rows: /flow "SDD + TDD per task" (runCount 2) and
// "flow.review-panel" (1), and the eight /flow-interactive stages of
// seedInteractive's sessions (KAN-1061) -- /flow-fast "integrate" is the
// eleventh and falls off; GET /api/v1/stats/trend?<PINNED_QUERY> returns two
// days (2026-08-14, 2026-08-15).

test.describe("stage leaderboard", () => {
  test("renders every costed stage with data loaded and the dark palette applied", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "Stage leaderboard" })).toBeVisible();

    const tablePanel = page.locator('section.panel[aria-label="Every costed stage"]');
    await expect(tablePanel.getByRole("cell", { name: "/flow", exact: true })).toHaveCount(2);
    await expect(tablePanel.getByRole("cell", { name: "/flow-interactive", exact: true })).toHaveCount(8);
    await expect(tablePanel.getByRole("cell", { name: "SDD + TDD per task", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "flow.review-panel", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "flow.landing-routes", exact: true })).toBeVisible();

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    const stagesPanel = page.locator('section.panel[aria-label="Stages"]');
    await expect(stagesPanel.getByText("Stages", { exact: true })).toHaveCount(1);
    const runsPanel = page.locator('section.panel[aria-label="Total runs"]');
    await expect(runsPanel.getByText("Total runs", { exact: true })).toHaveCount(1);

    await expect(tablePanel.getByRole("button", { name: /^Main session/ })).toBeVisible();
    await expect(tablePanel.getByRole("button", { name: /^Main share/ })).toBeVisible();

    await expect(page).toHaveScreenshot("stage-leaderboard.png", { fullPage: true });
  });

  test("stage leaderboard sorted by Main share", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
    const tablePanel = page.locator('section.panel[aria-label="Every costed stage"]');
    await expect(tablePanel.getByRole("cell", { name: "flow.review-panel", exact: true })).toBeVisible();

    const header = tablePanel.getByRole("button", { name: /^Main share/ });
    await header.click();
    await expect(header).toHaveText("Main share ▲");
    await page.mouse.move(0, 0);

    await expect(page).toHaveScreenshot("stage-leaderboard-main-share-sorted.png", { fullPage: true });
  });

  // StageLeaderboard.tsx declares its own nine sortable columns and its own
  // one filterable column (command) on top of the shared DataTable/
  // FilterBar components baseline.spec.ts's StateBoard spec and
  // run-detail.spec.ts's StageRunTable specs already prove correct in
  // general -- what is untested here is this *view's own* column
  // definitions (its own accessor per column), which is exactly the class
  // of bug a shared, already-correct component cannot catch: a column
  // pointed at the wrong field ties or reorders wrong while the sort
  // mechanism itself keeps working.
  //
  // The fixture's own ten rows (GET
  // /api/v1/stats/stage-leaderboard?<PINNED_QUERY>, this file's own
  // header comment) tie on Command, on Runs (flow.sdd-tdd and
  // flow.load-context 10, flow.review-panel and flow.brainstorm 1) and on
  // Median cost (flow.kickoff and flow.refresh-main-checkout 0.02), so those
  // assertions read only what the ties leave determined. Order assertions
  // read the Stage column, the one value unique to each row.
  //
  // flow.review-panel's one run carries two priced dispatch buckets (0.25
  // of its 0.41, seed.go), so its Main share reads below 100% while every
  // other row reads 100% -- the split the Main share sort captures.
  test.describe("sort and filter", () => {
    const tablePanelSelector = 'section.panel[aria-label="Every costed stage"]';

    test("sorting by Command orders alphabetically", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Command/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Command ▲");
      const ascending = ["/flow", "/flow", ...Array(8).fill("/flow-interactive")];
      await expect(commandCells).toHaveText(ascending);

      await header.click();
      await expect(header).toHaveText("Command ▼");
      await expect(commandCells).toHaveText([...ascending].reverse());
    });

    test("sorting by Stage orders by its own text, not Command's", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Stage/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");
      const stageCells = tablePanel.locator("table tbody tr td:nth-child(2)");

      await header.click();
      await expect(header).toHaveText("Stage ▲");
      // localeCompare order -- which puts the two /flow rows among the
      // /flow-interactive ones, an order Command's own sort above never
      // produces: what proves this column's accessor reads its own field
      // rather than Command's.
      await expect(stageCells).toHaveText([
        "flow.brainstorm", "flow.kickoff", "flow.landing-routes", "flow.load-context", "flow.refresh-main-checkout",
        "flow.review-panel", "flow.sdd-tdd", "flow.verify", "flow.visual-verify", "SDD + TDD per task",
      ]);
      const fi = "/flow-interactive";
      await expect(commandCells).toHaveText([fi, fi, fi, fi, fi, "/flow", fi, fi, fi, "/flow"]);
    });

    test("sorting by Runs orders by run count", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Runs/ });
      const stageCells = tablePanel.locator("table tbody tr td:nth-child(2)");

      await header.click();
      await expect(header).toHaveText("Runs ▲");
      // runCount: flow.review-panel and flow.brainstorm 1 (a tie) lowest,
      // flow.sdd-tdd and flow.load-context 10 (a tie) highest.
      await expect(stageCells.first()).toHaveText(/^flow\.(review-panel|brainstorm)$/);
      await expect(stageCells.last()).toHaveText(/^flow\.(sdd-tdd|load-context)$/);
    });

    test("sorting by Mean cost orders by mean cost", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Mean cost/ });
      const stageCells = tablePanel.locator("table tbody tr td:nth-child(2)");

      await header.click();
      await expect(header).toHaveText("Mean cost ▲");
      // meanCostUsd: flow.kickoff 0.02 lowest ... flow.review-panel 0.41, SDD + TDD per task 1.4645.
      await expect(stageCells).toHaveText([
        "flow.kickoff", "flow.load-context", "flow.refresh-main-checkout", "flow.visual-verify", "flow.landing-routes",
        "flow.verify", "flow.brainstorm", "flow.sdd-tdd", "flow.review-panel", "SDD + TDD per task",
      ]);
    });

    test("sorting by Median cost orders by median cost", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Median cost/ });
      const stageCells = tablePanel.locator("table tbody tr td:nth-child(2)");

      await header.click();
      await expect(header).toHaveText("Median cost ▲");
      // medianCostUsd: flow.load-context 0.015 lowest, flow.kickoff and
      // flow.refresh-main-checkout 0.02 (a tie), then flow.landing-routes 0.06
      // ... SDD + TDD per task ~1.4645 highest.
      await expect(stageCells.first()).toHaveText("flow.load-context");
      await expect(stageCells.nth(3)).toHaveText("flow.landing-routes");
      await expect(stageCells.last()).toHaveText("SDD + TDD per task");
    });

    test("sorting by P90 cost orders by p90 cost", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^P90 cost/ });
      const stageCells = tablePanel.locator("table tbody tr td:nth-child(2)");

      await header.click();
      await expect(header).toHaveText("P90 cost ▲");
      // p90CostUsd: flow.kickoff 0.02 lowest ... flow.review-panel 0.41, SDD + TDD per task 2.4929.
      await expect(stageCells).toHaveText([
        "flow.kickoff", "flow.load-context", "flow.refresh-main-checkout", "flow.visual-verify", "flow.landing-routes",
        "flow.verify", "flow.brainstorm", "flow.sdd-tdd", "flow.review-panel", "SDD + TDD per task",
      ]);
    });

    test("Filter by Command narrows to the matching stage", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      await expect(tablePanel.getByRole("cell", { name: "/flow", exact: true })).toHaveCount(2);

      await page.getByLabel("Filter by Command").selectOption("/flow-interactive");

      await expect(tablePanel.getByRole("cell", { name: "/flow-interactive", exact: true })).toHaveCount(8);
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
