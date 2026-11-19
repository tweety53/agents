import { expect, test } from "@playwright/test";
import { FIXTURE_NOW } from "./support";

// The Role costs view (#/role-costs, RoleCosts.tsx): each allowed-pairs
// role's spend, agent time, per-dispatch figures, failure rate and model ·
// effort split over the seeded dispatches, and the weekly spend per role.
// The clock is pinned to FIXTURE_NOW because the Period chips end at now.
test("role costs at rest", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await page.clock.setFixedTime(FIXTURE_NOW);
  await page.goto("/#/role-costs");
  await expect(page.getByRole("link", { name: "Role costs", exact: true })).toHaveAttribute("aria-current", "page");
  const table = page.getByRole("table", { name: "Spend and agent time per role" });
  await expect(table.getByRole("row").nth(1)).toBeVisible();
  await expect(page).toHaveScreenshot("role-costs.png", { fullPage: true });
});
