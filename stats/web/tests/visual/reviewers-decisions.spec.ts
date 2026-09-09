import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, PINNED_QUERY } from "./support";

// This change (KAN-472's own diff) adds two new static views -- Reviewers
// and Decisions -- that baseline.spec.ts (Dashboard, RunDetail) and
// views.spec.ts (stage leaderboard, trend, cache efficiency) do not cover.
// Each spec follows baseline.spec.ts's own shape: assert the data actually
// loaded, assert the dark palette applied, then toHaveScreenshot.

test.describe("reviewers", () => {
  test("renders every reviewer slot with data loaded and the dark palette applied", async ({ page }) => {
    await page.goto(`/#/reviewers?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "Reviewers" })).toBeVisible();

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("reviewers.png", { fullPage: true });
  });
});

test.describe("decisions", () => {
  test("renders the decision chain with data loaded and the dark palette applied", async ({ page }) => {
    await page.goto(`/#/decisions?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "Decisions" })).toBeVisible();

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("decisions.png", { fullPage: true });
  });
});
