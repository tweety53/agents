import { expect, test, type Page } from "@playwright/test";
import { FIXTURE_NOW } from "./support";

// Captures of the reworked Stage leaderboard (KAN-1052, decision
// leaderboard-capture-spec): the default view at 1280px, the same at 390px,
// and a Project chip's hover state -- every one full page. The view's Period
// chips are a rolling window ending now, so the clock is pinned to the
// UI-test fixture's FIXTURE_NOW (its timestamps are fixed at seed time).
//
// Values below are read from the live UI-test stack rather than assumed:
// GET /api/v1/stats/stage-leaderboard?from=2026-07-16T09:00:00Z&to=2026-08-15T09:00:00Z
// (the default 30d window ending at FIXTURE_NOW) returns 16 rows, the
// costliest /flow-interactive "flow.sdd-tdd" ($3.23, stageLabel "Implementing
// tasks"), and projects uitest-alpha ($10.149) then uitest-beta ($2.778) --
// so the first Project chip after All is "uitest-alpha". The previous window
// (2026-06-16T09:00:00Z to 2026-07-16T09:00:00Z) returns no rows, so every
// row's growth reads "new" and Biggest riser reads "—" ("no stage grew"),
// as the prototype draws a window with no previous spend.

const HEADERS = ["#", "Stage", "Command", "Spend", "Share", "vs previous 30d", "Runs", "Median run", "P90 run", "Main share"];

test.use({ colorScheme: "light" });

async function openAndAssert(page: Page) {
  await page.clock.setFixedTime(FIXTURE_NOW);
  await page.goto("/#/stage-leaderboard");
  await expect(page.getByText("Loading…")).toHaveCount(0);

  await expect(page.getByRole("link", { name: "Stage leaderboard" })).toHaveAttribute("aria-current", "page");
  const bar = page.getByRole("toolbar", { name: "Filters" });
  for (const g of ["Period", "Project", "Command"]) await expect(bar.getByRole("group", { name: g })).toBeVisible();
  await expect(bar.getByRole("group", { name: "Period" }).getByRole("button", { name: "30d" })).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByLabel("Period start")).toHaveCount(0);
  await expect(page.getByRole("combobox", { name: "Model" })).toHaveCount(0);

  for (const t of ["Spend", "Top 3 stages", "Biggest riser"]) await expect(page.locator(".lb-tile .lb-lbl", { hasText: new RegExp(`^${t}$`) })).toBeVisible();

  await expect(page.getByRole("table").getByRole("row").first().getByRole("columnheader")).toHaveText(HEADERS);
  await expect(page.getByRole("table").getByRole("row", { name: /Implementing tasks/ }).first()).toBeVisible();
}

test("stage leaderboard at rest", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await expect(page).toHaveScreenshot("stage-leaderboard-wide.png", { fullPage: true });
});

test("stage leaderboard on a phone", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await openAndAssert(page);
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
  await expect(page).toHaveScreenshot("stage-leaderboard-phone.png", { fullPage: true });
});

test("stage leaderboard project chip hover", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await page.getByRole("group", { name: "Project" }).getByRole("button").nth(1).hover();
  await expect(page).toHaveScreenshot("stage-leaderboard-chip-hover.png", { fullPage: true });
});
