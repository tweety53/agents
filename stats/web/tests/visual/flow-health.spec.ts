import { expect, test, type Page } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, FIXTURE_NOW } from "./support";

// Captures of the reworked Flow health view (KAN-1050): the view at rest at
// 1280px, the Never fired tile hovered, the Catching tile focused, the Never
// fired tile pressed, Rank by Run time with the Speed-up candidates tile
// pressed, and the first Project chip after All pressed -- every one full
// page. Its Period chips are a rolling window ending now, so the clock is
// pinned to the UI-test fixture's FIXTURE_NOW. Each capture is compared
// against its frame of the approved prototype through
// flow-health.spec.ts.mockups.
//
// Values below are read from the live UI-test stack rather than assumed:
// GET /api/v1/stats/guard-health?from=2026-07-16T09:00:00Z&to=2026-08-15T09:00:00Z&projects=1
// (the default 30d window ending at FIXTURE_NOW) returns thirteen guards --
// the helper commands project-get and run-verify-list are excluded -- and
// projects uitest-alpha (205 runs) then uitest-beta (18 runs), so the first
// Project chip after All is "uitest-alpha":
//   catching (5 of 13): check-plan-shape 12 fires, check-references 4,
//     check-verbatim-moves 2, check-dispatch-paragraphs 1, run-guard-tests 1;
//   never fired (8): break-and-prove, check-guard-calls,
//     check-installed-citations, check-markdown-integrity, check-no-bash,
//     check-spec-reach, check-stage-mark-calls, check-task-commit-fields --
//     so the tile's note ends "+5 more";
//   cannot answer: break-and-prove 1 of 10, check-guard-calls 2 of 8,
//     check-task-commit-fields 6 of 25;
//   speed-up candidates (4 of 15,231,970 ms total): run-guard-tests
//     (4,200,000 ms), check-verbatim-moves (3,840,000), break-and-prove
//     (3,000,000), check-installed-citations (2,400,000) -- the fourth
//     crosses 80%, so the tile's note ends "+1 more".

const TILES = ["Catching", "Never fired", "Cannot answer", "Speed-up candidates"];

// A tile by its label; the " ·" keeps "Never fired" off the Rank by chip of
// the same name.
const tile = (page: Page, label: string) => page.getByRole("button", { name: new RegExp(`^${label} ·`) });

async function openAndAssert(page: Page) {
  await page.clock.setFixedTime(FIXTURE_NOW);
  await page.goto("/#/flow-health");
  await expect(page.getByText("Loading…")).toHaveCount(0);

  await expect(page.getByRole("link", { name: "Flow health" })).toHaveAttribute("aria-current", "page");
  const bar = page.getByRole("toolbar", { name: "Filters" });
  for (const g of ["Period", "Project", "Rank by"]) await expect(bar.getByRole("group", { name: g })).toBeVisible();
  await expect(bar.getByRole("group", { name: "Period" }).getByRole("button", { name: "30d" })).toHaveAttribute("aria-pressed", "true");
  await expect(bar.getByRole("group", { name: "Rank by" }).getByRole("button", { name: "Fires" })).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByLabel("Period start")).toHaveCount(0);

  await expect(page.getByText("Which guards catch something, which never fire, and what does each cost in run time?")).toBeVisible();
  for (const t of TILES) await expect(tile(page, t)).toBeVisible();
  await expect(page.getByText("speed-up", { exact: true }).first()).toBeVisible();
  await expect(page.getByRole("cell", { name: /^project-get/ })).toHaveCount(0);
  await expect(page.getByRole("cell", { name: /^run-verify-list/ })).toHaveCount(0);

  expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
}

test("flow health at rest", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await expect(page).toHaveScreenshot("flow-health.png", { fullPage: true });
});

test("flow health tile hover", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await tile(page, "Never fired").hover();
  await expect(page).toHaveScreenshot("flow-health-tile-hover.png", { fullPage: true });
});

test("flow health tile focus", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  await tile(page, "Catching").focus();
  await expect(page).toHaveScreenshot("flow-health-tile-focus.png", { fullPage: true });
});

test("flow health never fired pressed", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  const t = tile(page, "Never fired");
  await t.click();
  await expect(t).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByRole("button", { name: "Clear filters" })).toBeVisible();
  await expect(page).toHaveScreenshot("flow-health-never-fired.png", { fullPage: true });
});

test("flow health speed-up by run time", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  const rank = page.getByRole("group", { name: "Rank by" }).getByRole("button", { name: "Run time" });
  await rank.click();
  await expect(rank).toHaveAttribute("aria-pressed", "true");
  const t = tile(page, "Speed-up candidates");
  await t.click();
  await expect(t).toHaveAttribute("aria-pressed", "true");
  await expect(page).toHaveScreenshot("flow-health-speed-up.png", { fullPage: true });
});

test("flow health project pressed", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await openAndAssert(page);
  const chip = page.getByRole("group", { name: "Project" }).getByRole("button").nth(1);
  await chip.click();
  await expect(chip).toHaveAttribute("aria-pressed", "true");
  await expect(page).toHaveScreenshot("flow-health-project.png", { fullPage: true });
});
