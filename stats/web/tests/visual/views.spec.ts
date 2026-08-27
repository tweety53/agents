import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, PINNED_QUERY } from "./support";

// baseline.spec.ts already covers the dashboard (state-board) and
// run-detail views -- the two KAN-16 broke. This file covers the three
// remaining static views task 12 requires ("every view reachable"):
// stage-leaderboard, trend and cache-efficiency. Each spec follows
// baseline.spec.ts's own shape: assert the data actually loaded (no
// error, no stuck "loading", the fixture's own rows visible), assert the
// dark palette actually applied, assert no stat-panel label is
// double-rendered (KAN-16's third defect class), then toHaveScreenshot --
// a screenshot alone proves the page painted, not that it is right, so
// every one of these assertions comes first.
//
// Values below are read from the live UI-test stack rather than assumed:
// GET /api/v1/stats/stage-leaderboard?<PINNED_QUERY> returns two rows
// (/myflow-do "SDD + TDD per task", runCount 2; /myflow-finish
// "integrate", runCount 1); GET /api/v1/stats/trend?<PINNED_QUERY>
// returns two days (2026-08-14, 2026-08-15); GET
// /api/v1/stats/cache-efficiency?<PINNED_QUERY> returns both stage rows
// with cacheReadTotal/cacheCreationTotal 0 and ratio null -- the fixture
// seeds no cache tokens, so every ratio cell renders through
// Unavailable.tsx as "unavailable", never a fabricated 0/0 or a blank.

test.describe("stage leaderboard", () => {
  test("renders every costed stage with data loaded and the dark palette applied", async ({ page }) => {
    await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "Stage leaderboard" })).toBeVisible();

    const tablePanel = page.locator('section.panel[aria-label="Every costed stage"]');
    await expect(tablePanel.getByRole("cell", { name: "/myflow-do", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "/myflow-finish", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "SDD + TDD per task", exact: true })).toBeVisible();
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
  // All six differ between the fixture's own two rows (GET
  // /api/v1/stats/stage-leaderboard?<PINNED_QUERY>, this file's own
  // header comment) -- unlike CacheEfficiency below, nothing here ties,
  // so every column gets a real order assertion.
  test.describe("sort and filter", () => {
    const tablePanelSelector = 'section.panel[aria-label="Every costed stage"]';

    test("sorting by Command orders alphabetically", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Command/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Command ▲");
      await expect(commandCells).toHaveText(["/myflow-do", "/myflow-finish"]);

      await header.click();
      await expect(header).toHaveText("Command ▼");
      await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);
    });

    test("sorting by Stage orders by its own text, not Command's", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Stage/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Stage ▲");
      // "integrate" < "SDD + TDD per task" (verified: "SDD + TDD per
      // task".localeCompare("integrate") === 1) -- the reverse of
      // Command's own ascending order above, which is what proves this
      // column's accessor reads its own field rather than Command's.
      await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);
    });

    test("sorting by Runs orders by run count", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Runs/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Runs ▲");
      // runCount: /myflow-finish 1, /myflow-do 2.
      await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);
    });

    test("sorting by Mean cost orders by mean cost", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Mean cost/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Mean cost ▲");
      // meanCostUsd: /myflow-finish 0.028, /myflow-do 1.4645.
      await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);
    });

    test("sorting by Median cost orders by median cost", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Median cost/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Median cost ▲");
      // medianCostUsd: /myflow-finish 0.028, /myflow-do ~1.4645.
      await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);
    });

    test("sorting by P90 cost orders by p90 cost", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^P90 cost/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("P90 cost ▲");
      // p90CostUsd: /myflow-finish 0.028, /myflow-do 2.4929.
      await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);
    });

    test("Filter by Command narrows to the matching stage", async ({ page }) => {
      await page.goto(`/#/stage-leaderboard?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      await expect(tablePanel.getByRole("cell", { name: "/myflow-do", exact: true })).toBeVisible();

      await page.getByLabel("Filter by Command").selectOption("/myflow-finish");

      await expect(tablePanel.getByRole("cell", { name: "/myflow-finish", exact: true })).toBeVisible();
      await expect(tablePanel.getByRole("cell", { name: "/myflow-do", exact: true })).toHaveCount(0);
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

test.describe("cache efficiency", () => {
  test("renders every stage's cache totals, unmeasured ratios, and the dark palette applied", async ({ page }) => {
    await page.goto(`/#/cache-efficiency?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "Cache efficiency" })).toBeVisible();

    const tablePanel = page.locator('section.panel[aria-label="Every stage"]');
    await expect(tablePanel.getByRole("cell", { name: "/myflow-do", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "/myflow-finish", exact: true })).toBeVisible();
    // The fixture records no cache tokens: every ratio cell must read
    // "unavailable" (Unavailable.tsx's own data-testid), never a
    // fabricated ratio or a silently blank cell -- the same
    // never-coerce-null-to-zero rule StatPanel.tsx's own header comment
    // states for the stat panels above it.
    await expect(tablePanel.locator('[data-testid="unavailable"]')).toHaveCount(2);

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    const readPanel = page.locator('section.panel[aria-label="Total cache read"]');
    await expect(readPanel.getByText("Total cache read", { exact: true })).toHaveCount(1);
    const ratioPanel = page.locator('section.panel[aria-label="Overall ratio"]');
    await expect(ratioPanel.getByText("Overall ratio", { exact: true })).toHaveCount(1);
    // Both totals are measured zeros (the fixture's runs recorded no
    // cache tokens at all, which is "0", not "absent") but the pooled
    // ratio divides by a zero denominator, so it is unavailable even
    // though its own inputs are not -- Unavailable.tsx's null branch,
    // not its measured branch.
    await expect(ratioPanel.locator('[data-testid="unavailable"]')).toHaveCount(1);

    await expect(page).toHaveScreenshot("cache-efficiency.png", { fullPage: true });
  });

  // CacheEfficiency.tsx declares its own five sortable columns and its own
  // one filterable column (command). Only two of the five are worth an
  // order assertion against this fixture: Command and Stage differ
  // between the two rows the same way they do on stage-leaderboard above.
  // Cache read, Cache creation and Read:creation ratio all **tie** --
  // this file's own header comment already establishes why (the fixture
  // records no cache tokens: cacheReadTotal and cacheCreationTotal are
  // both a measured 0 on both rows, and ratio is null on both) -- so a
  // click on any of those three headers cannot move either row, and an
  // order assertion there would pass whether or not the column's own
  // accessor were even wired up. Skipped for the same reason
  // StageRunTable's Attempt/Outcome/Duration were skipped in
  // run-detail.spec.ts: a tied column proves nothing about its own sort.
  test.describe("sort and filter", () => {
    const tablePanelSelector = 'section.panel[aria-label="Every stage"]';

    test("sorting by Command orders alphabetically", async ({ page }) => {
      await page.goto(`/#/cache-efficiency?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Command/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Command ▲");
      await expect(commandCells).toHaveText(["/myflow-do", "/myflow-finish"]);

      await header.click();
      await expect(header).toHaveText("Command ▼");
      await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);
    });

    test("sorting by Stage orders by its own text, not Command's", async ({ page }) => {
      await page.goto(`/#/cache-efficiency?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      const header = tablePanel.getByRole("button", { name: /^Stage/ });
      const commandCells = tablePanel.locator("table tbody tr td:nth-child(1)");

      await header.click();
      await expect(header).toHaveText("Stage ▲");
      // "integrate" < "SDD + TDD per task" -- same collation this file's
      // stage-leaderboard block already verified via localeCompare.
      await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);
    });

    test("Filter by Command narrows to the matching stage", async ({ page }) => {
      await page.goto(`/#/cache-efficiency?${PINNED_QUERY}`);
      const tablePanel = page.locator(tablePanelSelector);
      await expect(tablePanel.getByRole("cell", { name: "/myflow-do", exact: true })).toBeVisible();

      await page.getByLabel("Filter by Command").selectOption("/myflow-finish");

      await expect(tablePanel.getByRole("cell", { name: "/myflow-finish", exact: true })).toBeVisible();
      await expect(tablePanel.getByRole("cell", { name: "/myflow-do", exact: true })).toHaveCount(0);
    });
  });
});
