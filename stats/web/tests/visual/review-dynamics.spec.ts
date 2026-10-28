import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1 } from "./support";

// The Review dynamics view (#/review-dynamics, ReviewDynamics.tsx) over the
// eight weeks cmd/uitest-seed's seedReviewDynamics fills, 2026-05-04 ..
// 2026-06-29 -- before every other view's pinned window, so this fixture
// moves no other baseline. The range is pinned by query params, so no
// preset is active -- but in the default-preset test, which pins the clock
// to the fixture's end instead.
const DYNAMICS_QUERY = "from=2026-05-04T00%3A00%3A00.000Z&to=2026-06-29T00%3A00%3A00.000Z";

async function open(page: import("@playwright/test").Page) {
  await page.goto(`/#/review-dynamics?${DYNAMICS_QUERY}`);
  await expect(page.getByRole("link", { name: "Review dynamics", exact: true })).toHaveAttribute("aria-current", "page");
  await expect(page.getByRole("heading", { name: "Review dynamics" })).toBeVisible();
  await expect(page.getByText("Loading…")).toHaveCount(0);
  await expect(page.getByRole("region", { name: "Findings per change by severity" })).toBeVisible();
}

test.describe("review dynamics", () => {
  test("review dynamics full page", async ({ page }) => {
    await open(page);
    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
    // The prototype's measures: the intro is a 12px .sub, the heatmap packs
    // 92px cells with no row rules.
    await expect(page.locator("section > p.dynamics-sub").first()).toHaveCSS("font-size", "12px");
    const cell = page.locator(".dynamics-heat td").first();
    await expect(cell).toHaveCSS("width", "92px");
    await expect(cell).toHaveCSS("border-bottom-style", "none");
    await expect(page).toHaveScreenshot("review-dynamics.png", { fullPage: true });
  });

  test("review dynamics by version", async ({ page }) => {
    await open(page);
    await page.getByRole("button", { name: "Version", exact: true }).click();
    await expect(page.getByRole("button", { name: "Version", exact: true })).toHaveAttribute("aria-pressed", "true");
    await expect(page).toHaveScreenshot("review-dynamics-version.png", { fullPage: true });
  });

  test("review dynamics default preset", async ({ page }) => {
    await page.clock.setFixedTime(new Date("2026-06-29T00:00:00Z"));
    await page.goto("/#/review-dynamics");
    await expect(page.getByRole("button", { name: "Last 8 weeks" })).toHaveAttribute("aria-pressed", "true");
    await expect(page.getByText("Loading…")).toHaveCount(0);
    await expect(page.getByRole("region", { name: "Findings per change by severity" })).toBeVisible();
    await expect(page).toHaveScreenshot("review-dynamics-default-preset.png", { fullPage: true });
  });

  test("review dynamics chart hover", async ({ page }) => {
    await open(page);
    const severity = page.getByRole("region", { name: "Findings per change by severity" });
    // The week of 05-18 carries the primary reviewer prompt edit's marker.
    await severity.locator('[data-bucket="2026-05-18"]').hover();
    const tip = severity.getByRole("tooltip");
    await expect(tip).toContainText("◆ Prompt edit · primary");
    await expect(tip).toHaveCSS("background-color", "rgb(34, 37, 43)");
    await expect(page).toHaveScreenshot("review-dynamics-hover.png", { fullPage: true });
  });

  test("review dynamics heatmap cell hover", async ({ page }) => {
    await open(page);
    // The prototype's cell hover is the native title tooltip, which no
    // screenshot draws: its text is asserted, the hovered page captured.
    const cell = page.locator(".dynamics-heat td[title]").first();
    await expect(cell).toHaveAttribute("title", /dispatches\n\d+ critical · \d+ important · \d+ minor$/);
    await cell.hover();
    await expect(page).toHaveScreenshot("review-dynamics-heat-hover.png", { fullPage: true });
  });

  test.describe("light", () => {
    test.use({ colorScheme: "light" });
    test("review dynamics light", async ({ page }) => {
      await open(page);
      await expect(page).toHaveScreenshot("review-dynamics-light.png", { fullPage: true });
    });
  });
});
