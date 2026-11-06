import { expect, test, type Page } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, FIXTURE_NOW } from "./support";

// Captures of the reworked Decisions view (KAN-1055): the default view at
// 1280px, a Project chip's hover state, the big class narrowed to the static
// grouping, a row expanded beside a collapsed class, and a row's hover state
// -- every one full page. Its Period chips are a rolling window ending now,
// so the clock is pinned to the UI-test fixture's FIXTURE_NOW. Each capture
// is compared against its frame of the approved prototype through
// decisions.spec.ts.mockups.
//
// Values below are read from the live UI-test stack rather than assumed:
// GET /api/v1/stats/decisions?from=2026-07-16T09:00:00Z&to=2026-08-15T09:00:00Z
// (the default 30d window ending at FIXTURE_NOW) returns twenty rows --
// fourteen in uitest-alpha and six in uitest-beta, so the first Project chip
// after All is "uitest-alpha":
//   big (9): opus · medium free x4 (kan-601..604), opus · low static x3
//     (kan-605..607), sonnet · medium free x1 (kan-608, one fallback), and
//     fixtureData()'s kan-201 opus · high default x1;
//   regular (7): inline free x3 (kan-609 "skipped — inline", kan-610
//     "skipped - inline", kan-611 "inline"), opus · low static x3
//     (kan-612..614) and fixtureData()'s kan-102 sonnet · medium default x1;
//   small (3): inline free x2 (kan-615, 616) and opus · low static x1
//     (kan-617) -- no choice has 3 runs, so its tile reads "no choice has 3
//     runs yet · 3 runs";
//   unclassified (1): inline static x1 (kan-618, no class recorded).

async function openAndAssert(page: Page) {
  await page.clock.setFixedTime(FIXTURE_NOW);
  await page.goto("/#/decisions");
  await expect(page.getByText("Loading…")).toHaveCount(0);

  await expect(page.getByRole("link", { name: "Decisions" })).toHaveAttribute("aria-current", "page");
  const bar = page.getByRole("toolbar", { name: "Filters" });
  for (const g of ["Period", "Project", "Class", "Review grouping"]) await expect(bar.getByRole("group", { name: g })).toBeVisible();
  await expect(bar.getByRole("group", { name: "Period" }).getByRole("button", { name: "30d" })).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByLabel("Period start")).toHaveCount(0);

  await expect(page.getByText("For each class of change, which implementer choice and review grouping deliver it cheapest, fastest and with the fewest findings?")).toBeVisible();
  await expect(page.getByText("big · cheapest", { exact: true })).toBeVisible();
  await expect(page.getByText("no choice has 3 runs yet · 3 runs", { exact: true })).toBeVisible();
  await expect(page.getByText("few runs", { exact: true }).first()).toBeVisible();

  expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
}

// A class's group row, by its accessible name ("▾big9 runs").
const groupRow = (page: Page, cls: string) => page.getByRole("row", { name: new RegExp(`^▾ ?${cls} ?\\d+ runs$`) });

// A pair's row, by its accessible name, which opens with the caret, the
// implementer choice and the review grouping and goes on with the run count.
const pairRow = (page: Page, choice: string, grouping: string, runs: number) =>
  page.getByRole("row", { name: new RegExp(`^▾ ?${choice} ?${grouping} ?${runs} `) });

test("decisions at rest", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await expect(page).toHaveScreenshot("decisions.png", { fullPage: true });
});

test("decisions project chip hover", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await page.getByRole("group", { name: "Project" }).getByRole("button").nth(1).hover();
  await expect(page).toHaveScreenshot("decisions-chip-hover.png", { fullPage: true });
});

test("decisions big static", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  const cls = page.getByRole("group", { name: "Class" }).getByRole("button", { name: /^big/ });
  await cls.click();
  await expect(cls).toHaveAttribute("aria-pressed", "true");
  const grp = page.getByRole("group", { name: "Review grouping" }).getByRole("button", { name: /^static/ });
  await grp.click();
  await expect(grp).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByRole("button", { name: "Clear filters" })).toBeVisible();
  await expect(page).toHaveScreenshot("decisions-big-static.png", { fullPage: true });
});

test("decisions row expanded", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  const row = pairRow(page, "opus · medium", "free", 4);
  await row.click();
  await expect(row).toHaveAttribute("aria-expanded", "true");
  await expect(page.getByRole("link", { name: "kan-601-decisions" })).toBeVisible();
  const small = groupRow(page, "small");
  await small.click();
  await expect(small).toHaveAttribute("aria-expanded", "false");
  await expect(page).toHaveScreenshot("decisions-row-expanded.png", { fullPage: true });
});

test("decisions row hover", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await pairRow(page, "inline", "free", 3).hover();
  await expect(page).toHaveScreenshot("decisions-row-hover.png", { fullPage: true });
});
