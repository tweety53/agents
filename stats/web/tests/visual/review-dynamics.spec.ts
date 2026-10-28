import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1 } from "./support";

// The Review dynamics view (#/review-dynamics, ReviewDynamics.tsx) over the
// eight weeks cmd/uitest-seed's seedReviewDynamics fills, 2026-05-04 ..
// 2026-06-29 -- before every other view's pinned window, so this fixture
// moves no other baseline. The range is pinned by query params, so no
// preset is active.
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
    await expect(page).toHaveScreenshot("review-dynamics.png", { fullPage: true });
  });

  test("review dynamics chart hover", async ({ page }) => {
    await open(page);
    const severity = page.getByRole("region", { name: "Findings per change by severity" });
    // The week of 05-18 carries the primary reviewer prompt edit's marker.
    await severity.locator('[data-bucket="2026-05-18"]').hover();
    await expect(severity.getByRole("tooltip")).toContainText("◆ Prompt edit · primary");
    await expect(severity).toHaveScreenshot("review-dynamics-hover.png");
  });

  test.describe("light", () => {
    test.use({ colorScheme: "light" });
    test("review dynamics light", async ({ page }) => {
      await open(page);
      await expect(page).toHaveScreenshot("review-dynamics-light.png", { fullPage: true });
    });
  });
});
