import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, RUNS_QUERY } from "./support";

test.describe("runs", () => {
  test("renders runs grouped by change, expanded to main session and dispatches", async ({ page }) => {
    await page.goto(`/#/runs?${RUNS_QUERY}`);
    await expect(page.getByRole("heading", { name: "Runs", exact: true })).toBeVisible();
    await expect(page.getByRole("heading", { name: "kan-103-runs-view" })).toBeVisible();
    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    const section = page.getByRole("region", { name: "kan-103-runs-view" });
    await section.getByRole("button", { name: /run details/i }).nth(1).click();
    await expect(section.getByTestId("main-session-row")).toBeVisible();
    await expect(section.getByTestId("dispatch-mismatch")).toBeVisible();

    await expect(page).toHaveScreenshot("runs.png", { fullPage: true });
  });
});
