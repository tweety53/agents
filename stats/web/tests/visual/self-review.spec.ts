import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, PINNED_QUERY } from "./support";

// The "Self-review" view (#/self-review, SelfReview.tsx). The UI-test
// fixture (cmd/uitest-seed, seedSelfReviewPair) seeds only fixed findings,
// none to review, so the view opens on its "To review" list empty and this
// spec captures that empty state. The view reads no period: its window runs
// from 2000-01-01 to a day past now.
test.describe("self-review", () => {
  test("renders the empty state with the dark palette applied", async ({ page }) => {
    await page.goto(`/#/self-review?${PINNED_QUERY}`);
    await expect(page.getByRole("link", { name: "Self-review", exact: true })).toHaveAttribute("aria-current", "page");
    await expect(page.getByRole("heading", { name: "Self-review", exact: true })).toBeVisible();
    await expect(page.getByText("Nothing here.")).toBeVisible();
    await expect(page.getByRole("region", { name: "Per change" })).toHaveCount(0);

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("self-review-empty.png", { fullPage: true });
  });
});
