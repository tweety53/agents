import { expect, test, type Page } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1 } from "./support";

// Weekly costs (KAN-1057): one stacked-bar chart, a bar per Wednesday week,
// under a Period / Split by / Project chip bar, four tiles and a What moved
// table comparing the last full week with the week before. The window is
// computed from the browser's own clock (WeeklyCosts.tsx's
// weekWindow(n, Date.now())), so the clock is pinned before the first
// navigation to an instant just after the UI-test fixture's own data
// (cmd/uitest-seed/seed.go's fixtureNow, 2026-08-15T09:00:00Z): a Sunday,
// so the current week is 2026-08-12, day 5 of 7, the last full week
// 2026-08-05 and the week it is compared with 2026-07-29, on every run.
//
// The fixture's runs fall in the weeks of 2026-07-15, 2026-07-22 and
// 2026-08-12, so neither compared week carries spend of its own:
// withComparedWeeks adds a 2026-07-29 and a 2026-08-05 row to the stack's
// real weekly-costs and weekly-model-effort-costs responses, chosen so What
// moved holds a rise (Implementation), a fall (Review) and a series at $0 in
// both weeks (Self-review, left out), and so the model · effort split carries
// seven pairs, folding the cheapest into Other models.
// The empty path pins the clock years before the fixture, so the view has no
// recorded run and reads Panel's not-recorded banner.
const AFTER_FIXTURE = new Date("2026-08-16T12:00:00Z");
const BEFORE_FIXTURE = new Date("2020-01-01T12:00:00Z");

test.use({ viewport: { width: 1280, height: 1024 } });

const PHASE_ROWS = [
  { week: "2026-07-29", planningUsd: 1, implementationUsd: 2, reviewUsd: 3, selfReviewUsd: null, visualVerifyUsd: null, otherUsd: 0.5 },
  { week: "2026-08-05", planningUsd: 1.5, implementationUsd: 5, reviewUsd: 1, selfReviewUsd: null, visualVerifyUsd: null, otherUsd: 0.5 },
].map((r) => ({ ...r, overallUsd: r.planningUsd + r.implementationUsd + r.reviewUsd + r.otherUsd }));

const MODEL_ROWS = [
  { week: "2026-07-29", model: "claude-opus-5-5", effort: "medium", costUsd: 4 },
  { week: "2026-08-05", model: "claude-opus-5-5", effort: "medium", costUsd: 6 },
  { week: "2026-08-05", model: "claude-opus-5-5", effort: "high", costUsd: 0.05 },
  { week: "2026-07-29", model: "claude-opus-5-5", effort: "session", costUsd: 2.5 },
];

async function stub(page: Page, view: string, rows: { week: string }[]) {
  await page.route(`**/api/v1/stats/${view}?*`, async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    body.rows = [...body.rows, ...rows].sort((a: { week: string }, b: { week: string }) => a.week.localeCompare(b.week));
    await route.fulfill({ response, json: body });
  });
}

async function withComparedWeeks(page: Page) {
  await stub(page, "weekly-costs", PHASE_ROWS);
  await stub(page, "weekly-model-effort-costs", MODEL_ROWS);
}

async function open(page: Page, time: Date) {
  await page.clock.install({ time });
  await page.goto("/#/weekly-costs");
  await expect(page.getByRole("heading", { name: "Weekly costs" })).toBeVisible();
}

const chip = (page: Page, group: string, name: string) =>
  page.getByRole("group", { name: group }).getByRole("button", { name, exact: true });
const segment = (page: Page, week: string) => page.locator(`[data-testid="wc-segment"][data-week="${week}"]`);
const whatMoved = (page: Page) => page.getByText("What moved, week of 2026-08-05 vs 2026-07-29");

test.describe("weekly costs", () => {
  test("draws the phase split in the dark palette", async ({ page }) => {
    await withComparedWeeks(page);
    await open(page, AFTER_FIXTURE);
    await expect(whatMoved(page)).toBeVisible();
    await expect(page.getByRole("link", { name: "Weekly costs", exact: true })).toHaveAttribute("aria-current", "page");
    await expect(page.getByRole("toolbar", { name: "Dashboard controls" })).toHaveCount(0);
    for (const [group, name] of [["Period", "8w"], ["Split by", "Phase"], ["Project", "All"]])
      await expect(chip(page, group, name)).toHaveAttribute("aria-pressed", "true");
    await expect(page.getByText("This week so far · day 5 of 7")).toBeVisible();
    await expect(page.locator(".wc .lb-table .up").first()).toBeVisible();
    await expect(page.locator(".wc .lb-table .down").first()).toBeVisible();
    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
    await page.mouse.move(0, 0);

    await expect(page).toHaveScreenshot("weekly-costs.png", { fullPage: true });
  });

  test("draws the view in the light palette", async ({ page }) => {
    await page.emulateMedia({ colorScheme: "light" });
    await withComparedWeeks(page);
    await open(page, AFTER_FIXTURE);
    await expect(whatMoved(page)).toBeVisible();
    expect(await bodyBackground(page)).not.toBe(DARK_SURFACE_1);
    await page.mouse.move(0, 0);

    await expect(page).toHaveScreenshot("weekly-costs-light.png", { fullPage: true });
  });

  test("renders the empty state before any recorded run", async ({ page }) => {
    await open(page, BEFORE_FIXTURE);
    await expect(page.getByTestId("not-recorded")).toHaveCount(1);
    await expect(page.getByTestId("wc-segment")).toHaveCount(0);
    await page.mouse.move(0, 0);

    await expect(page).toHaveScreenshot("weekly-costs-empty.png", { fullPage: true });
  });

  test("draws the model · effort split", async ({ page }) => {
    await withComparedWeeks(page);
    await open(page, AFTER_FIXTURE);
    await chip(page, "Split by", "Model · effort").click();
    await expect(page.getByText("Spend per week by model · effort")).toBeVisible();
    await expect(chip(page, "Split by", "Model · effort")).toHaveAttribute("aria-pressed", "true");
    await expect(page.locator(".wc-legend > span").last()).toHaveText("Other models");
    await page.mouse.move(0, 0);

    await expect(page).toHaveScreenshot("weekly-costs-model-effort.png", { fullPage: true });
  });

  test("hovering a bar shows its week's tooltip", async ({ page }) => {
    await withComparedWeeks(page);
    await open(page, AFTER_FIXTURE);
    await expect(page.getByRole("tooltip")).toHaveCount(0);
    await segment(page, "2026-08-05").first().hover();
    const tip = page.getByRole("tooltip");
    await expect(tip).toContainText("Week of 2026-08-05");
    await expect(tip).toContainText("Total");

    await expect(page).toHaveScreenshot("weekly-costs-tooltip.png", { fullPage: true });
  });

  test("focusing a bar from the keyboard shows its tooltip", async ({ page }) => {
    await withComparedWeeks(page);
    await open(page, AFTER_FIXTURE);
    await page.mouse.move(0, 0);
    const first = segment(page, "2026-08-12").first();
    await first.focus();
    await page.keyboard.press("Shift+Tab");
    await page.keyboard.press("Tab");
    await expect(first).toBeFocused();
    await expect(page.getByRole("tooltip")).toContainText("Week of 2026-08-12 (so far)");

    await expect(page).toHaveScreenshot("weekly-costs-focus.png", { fullPage: true });
  });

  test("pressed Period and Project chips narrow the view", async ({ page }) => {
    await withComparedWeeks(page);
    await open(page, AFTER_FIXTURE);
    await chip(page, "Period", "12w").click();
    await chip(page, "Project", "uitest-alpha").click();
    for (const [group, name, on] of [
      ["Period", "12w", "true"],
      ["Project", "uitest-alpha", "true"],
      ["Period", "8w", "false"],
      ["Project", "All", "false"],
    ])
      await expect(chip(page, group, name)).toHaveAttribute("aria-pressed", on);
    await expect(whatMoved(page)).toBeVisible();
    await page.mouse.move(0, 0);

    await expect(page).toHaveScreenshot("weekly-costs-filtered.png", { fullPage: true });
  });

  test("hovering a chip highlights it", async ({ page }) => {
    await withComparedWeeks(page);
    await open(page, AFTER_FIXTURE);
    await expect(whatMoved(page)).toBeVisible();
    await chip(page, "Period", "26w").hover();

    await expect(page).toHaveScreenshot("weekly-costs-chip-hover.png", { fullPage: true });
  });
});
