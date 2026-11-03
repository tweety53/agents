import { expect, test, type Page } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, FIXTURE_NOW, PINNED_QUERY } from "./support";

// Captures of the reworked Reviewers view (KAN-1053): the default view at
// 1280px, each other Rank by chip, the same at 390px and a Project chip's
// hover state -- every one full page. Its Period chips are a rolling window
// ending now, so the clock is pinned to the UI-test fixture's FIXTURE_NOW.
//
// Values below are read from the live UI-test stack rather than assumed:
// GET /api/v1/stats/reviewers?from=2026-07-16T09:00:00Z&to=2026-08-15T09:00:00Z&projects=1
// (the default 30d window ending at FIXTURE_NOW) returns two rows -- primary
// (1 dispatch, 1 priced, costUsd 0.15, critical 1, minor 1, withdrawn 0) and
// principles (1 dispatch, 1 priced, costUsd 0.1, important 1, withdrawn 0) --
// and projects [{uitest-alpha, 2}], so the first Project chip after All is
// "uitest-alpha". primary ranks first on value ($0.15 for 11 weighted) and on
// severity (11 weighted per dispatch against 4).
//
// The decisions describe below follows baseline.spec.ts's shape: assert the
// data loaded, assert the dark palette applied, then toHaveScreenshot.

const TILES = ["Cost · best value", "Severity · most found", "False alarms · most withdrawn"];

async function openAndAssert(page: Page) {
  await page.clock.setFixedTime(FIXTURE_NOW);
  await page.goto("/#/reviewers");
  await expect(page.getByText("Loading…")).toHaveCount(0);

  await expect(page.getByRole("link", { name: "Reviewers" })).toHaveAttribute("aria-current", "page");
  const bar = page.getByRole("toolbar", { name: "Filters" });
  for (const g of ["Period", "Project", "Rank by"]) await expect(bar.getByRole("group", { name: g })).toBeVisible();
  await expect(bar.getByRole("group", { name: "Period" }).getByRole("button", { name: "30d" })).toHaveAttribute("aria-pressed", "true");
  await expect(bar.getByRole("group", { name: "Rank by" }).getByRole("button", { name: "Value" })).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByLabel("Period start")).toHaveCount(0);

  for (const t of TILES) await expect(page.getByText(t, { exact: true })).toBeVisible();
  await expect(page.getByText("Every slot, best value first")).toBeVisible();
  const table = await page.getByRole("table").innerText();
  expect(table.indexOf("primary")).toBeGreaterThanOrEqual(0);
  expect(table.indexOf("primary")).toBeLessThan(table.indexOf("principles"));

  expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
}

async function rankBy(page: Page, name: string, title: string) {
  await page.getByRole("group", { name: "Rank by" }).getByRole("button", { name }).click();
  await expect(page.getByText(title)).toBeVisible();
}

test("reviewers at rest", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await expect(page).toHaveScreenshot("reviewers.png", { fullPage: true });
});

test("reviewers ranked by severity", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await rankBy(page, "Severity", "Every slot, most severe first");
  await expect(page).toHaveScreenshot("reviewers-severity.png", { fullPage: true });
});

test("reviewers ranked by false alarms", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await rankBy(page, "False alarms", "Every slot, most false alarms first");
  await expect(page).toHaveScreenshot("reviewers-false-alarms.png", { fullPage: true });
});

test("reviewers on a phone", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await openAndAssert(page);
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
  await expect(page).toHaveScreenshot("reviewers-phone.png", { fullPage: true });
});

test("reviewers project chip hover", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await page.getByRole("group", { name: "Project" }).getByRole("button").nth(1).hover();
  await expect(page).toHaveScreenshot("reviewers-chip-hover.png", { fullPage: true });
});

test.describe("decisions", () => {
  test("renders the decision chain with data loaded and the dark palette applied", async ({ page }) => {
    await page.goto(`/#/decisions?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "Decisions" })).toBeVisible();

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("decisions.png", { fullPage: true });
  });
});
