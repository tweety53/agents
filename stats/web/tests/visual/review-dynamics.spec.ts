import { expect, test, type Page } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1 } from "./support";

// The Reviewer findings view (#/review-dynamics, ReviewDynamics.tsx): the
// findings-per-dispatch heatmap and its one Range control, over the eight
// weeks cmd/uitest-seed's seedReviewDynamics fills, 2026-05-04 ..
// 2026-06-29 -- before every other view's pinned window, so this fixture
// moves no other baseline. The clock is pinned to the fixture's end, since
// every range is counted back from now.
async function open(page: Page) {
  await page.clock.setFixedTime(new Date("2026-06-29T00:00:00Z"));
  await page.goto("/#/review-dynamics");
  await expect(page.getByRole("link", { name: "Reviewer findings", exact: true })).toHaveAttribute("aria-current", "page");
  await expect(page.getByText("Loading…")).toHaveCount(0);
  await expect(page.getByRole("heading", { name: "Findings per dispatch: slot × model · effort" })).toBeVisible();
}

test.describe("reviewer findings", () => {
  test("reviewer findings, last 30 days", async ({ page }) => {
    await open(page);
    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
    const range = page.getByRole("group", { name: "Range" });
    await expect(range.getByRole("button", { name: "30 days" })).toHaveAttribute("aria-pressed", "true");
    await expect(range.getByRole("button", { name: "Later" })).toBeDisabled();
    await expect(page).toHaveScreenshot("reviewer-findings.png", { fullPage: true });
  });

  test("reviewer findings, 14 days stepped one range back", async ({ page }) => {
    await open(page);
    const range = page.getByRole("group", { name: "Range" });
    await range.getByRole("button", { name: "14 days" }).click();
    await range.getByRole("button", { name: "Earlier" }).click();
    await expect(range).toContainText("06-01 – 06-15");
    await expect(range.getByRole("button", { name: "Later" })).toBeEnabled();
    await expect(page.getByText("Loading…")).toHaveCount(0);
    await expect(page).toHaveScreenshot("reviewer-findings-earlier.png", { fullPage: true });
  });
});
