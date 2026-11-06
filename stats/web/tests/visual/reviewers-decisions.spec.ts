import { expect, test, type Page } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, FIXTURE_NOW } from "./support";

// Captures of the reworked Reviewers view (KAN-1053): the default view at
// 1280px, each other Rank by chip, the same at 390px and a Project chip's
// hover state -- every one full page. Its Period chips are a rolling window
// ending now, so the clock is pinned to the UI-test fixture's FIXTURE_NOW.
//
// Values below are read from the live UI-test stack rather than assumed:
// GET /api/v1/stats/reviewers?from=2026-07-16T09:00:00Z&to=2026-08-15T09:00:00Z&projects=1
// (the default 30d window ending at FIXTURE_NOW) returns thirteen rows, one
// dispatch each -- kan-102's primary ($0.15, critical 1, minor 1) and
// principles ($0.1, important 1), and kan-104's mutation ($0.125, important 3,
// minor 2, withdrawn 1), exp-test-sanity ($0.125, no findings: the
// experimental badge and "found nothing"), call-efficiency ($0.15, critical 1,
// withdrawn 2), failure-modes ($0.15, important 1, withdrawn 1) and
// test-sanity (unpriced, critical 1, minor 2), plus six unpriced slots with
// no withdrawn finding -- security (important 2), data-integrity (important
// 1, minor 1), api-contract (minor 3), concurrency (important 1),
// performance (minor 1) and docs (minor 2) -- and projects
// [{uitest-alpha, 13}], so the first Project chip after All is "uitest-alpha".
// Every slot but primary and principles carries a description. Each Rank by
// chip orders the rows differently (VALUE, SEVERITY, ALARMS below).

const TILES = ["Cost · best value", "Severity · most found", "False alarms · most withdrawn"];
const VALUE = ["mutation", "primary", "call-efficiency", "principles", "failure-modes", "api-contract", "concurrency", "data-integrity", "docs", "exp-test-sanity", "performance", "security", "test-sanity"];
const SEVERITY = ["mutation", "test-sanity", "primary", "call-efficiency", "security", "data-integrity", "concurrency", "failure-modes", "principles", "api-contract", "docs", "performance", "exp-test-sanity"];
const ALARMS = ["call-efficiency", "failure-modes", "mutation", "api-contract", "concurrency", "data-integrity", "docs", "performance", "primary", "principles", "security", "test-sanity", "exp-test-sanity"];

const slotOrder = (page: Page) => page.locator("tbody .rv-slot > span:first-child");

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
  await expect(slotOrder(page)).toHaveText(VALUE);
  await expect(page.getByRole("img", { name: "experimental reviewer" })).toHaveCount(1);
  await expect(page.getByRole("table").getByText("found nothing", { exact: true })).toBeVisible();
  await expect(page.getByText("found nothing: exp-test-sanity ($0.13)")).toBeVisible();

  expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
}

async function rankBy(page: Page, name: string, title: string, order: string[]) {
  await page.getByRole("group", { name: "Rank by" }).getByRole("button", { name }).click();
  await expect(page.getByText(title)).toBeVisible();
  await expect(slotOrder(page)).toHaveText(order);
}

test("reviewers at rest", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await expect(page).toHaveScreenshot("reviewers.png", { fullPage: true });
});

test("reviewers ranked by severity", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await rankBy(page, "Severity", "Every slot, most severe first", SEVERITY);
  await expect(page).toHaveScreenshot("reviewers-severity.png", { fullPage: true });
});

test("reviewers ranked by false alarms", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await rankBy(page, "False alarms", "Every slot, most false alarms first", ALARMS);
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
