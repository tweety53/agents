import { expect, test, type Page } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, FIXTURE_NOW } from "./support";

// Captures of the Flow guards view (grouped panels, no tiles): at rest at
// 1280px, ranked by Never fired, ranked by Run time, and the first Project
// chip after All pressed -- every one full page. Its Period chips are a
// rolling window ending now, so the clock is pinned to the UI-test fixture's
// FIXTURE_NOW. The fixture's projects are uitest-alpha then uitest-beta, so
// the first Project chip after All is "uitest-alpha"; project-get and
// run-verify-list are helpers, drawn with blank fire columns.

const rankChip = (page: Page, name: string) => page.getByRole("group", { name: "Rank by" }).getByRole("button", { name, exact: true });

async function openAndAssert(page: Page) {
  await page.clock.setFixedTime(FIXTURE_NOW);
  await page.goto("/#/flow-health");
  await expect(page.getByText("Loading…")).toHaveCount(0);

  await expect(page.getByRole("link", { name: "Flow guards" })).toHaveAttribute("aria-current", "page");
  const bar = page.getByRole("toolbar", { name: "Filters" });
  for (const g of ["Period", "Project", "Rank by"]) await expect(bar.getByRole("group", { name: g })).toBeVisible();
  await expect(bar.getByRole("group", { name: "Period" }).getByRole("button", { name: "7d" })).toHaveAttribute("aria-pressed", "true");
  await expect(rankChip(page, "Fires")).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByRole("button", { name: "Clear filters" })).toHaveCount(0);

  await expect(page.getByText("Which guards catch something, which never fire, and what does each cost in run time?")).toBeVisible();
  await expect(page.getByText("Every guard, most fires first.")).toBeVisible();
  await expect(page.getByRole("cell", { name: /^check-guard-calls/ })).toBeVisible();

  expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
}

test.beforeEach(async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
});

test("flow health at rest", async ({ page }) => {
  await expect(page).toHaveScreenshot("flow-health.png", { fullPage: true });
});

for (const [rank, title, file] of [
  ["Never fired", "never-fired first, most runs first.", "flow-health-never-fired.png"],
  ["Run time", "most run time first.", "flow-health-speed-up.png"],
]) {
  test(`flow health ranked by ${rank}`, async ({ page }) => {
    await rankChip(page, rank).click();
    await expect(rankChip(page, rank)).toHaveAttribute("aria-pressed", "true");
    await expect(page.getByText(`Every guard, ${title}`)).toBeVisible();
    await expect(page).toHaveScreenshot(file, { fullPage: true });
  });
}

test("flow health project pressed", async ({ page }) => {
  const chip = page.getByRole("group", { name: "Project" }).getByRole("button").nth(1);
  await chip.click();
  await expect(chip).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByRole("button", { name: "Clear filters" })).toBeVisible();
  await expect(page).toHaveScreenshot("flow-health-project.png", { fullPage: true });
});
