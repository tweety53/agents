import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, PINNED_QUERY } from "./support";

// The "Self-review fixes" view (#/self-review, SelfReview.tsx). The UI-test
// fixture (cmd/uitest-seed) seeds no self-review findings, so over
// PINNED_QUERY the API answers `recorded: true, rows: []` and this spec
// captures the view's empty state: both stat panels at 0 and both tables
// showing their empty message. The period is pinned by query params for the
// reason support.ts's PINNED_QUERY comment gives.
test.describe("self-review fixes", () => {
  test("renders the empty state with the dark palette applied", async ({ page }) => {
    await page.goto(`/#/self-review?${PINNED_QUERY}`);
    await expect(page.getByRole("link", { name: "Self-review fixes" })).toHaveAttribute("aria-current", "page");
    await expect(page.getByRole("heading", { name: "Self-review fixes" })).toBeVisible();
    await expect(page.getByText("No self-review findings in this period.")).toBeVisible();
    await expect(page.getByText("No changes with self-review findings in this period.")).toBeVisible();

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    await expect(page).toHaveScreenshot("self-review-empty.png", { fullPage: true });
  });
});
