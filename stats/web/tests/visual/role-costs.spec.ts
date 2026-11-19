import { readFileSync } from "node:fs";
import { expect, test } from "@playwright/test";
import { FIXTURE_NOW } from "./support";

// The Role costs view (#/role-costs, RoleCosts.tsx), drawn as the approved
// prototype (.superpowers/sdd/kan-1130-interactive-session-2026-11-17/
// role-costs-prototype.html): roles grouped by stage with a model · effort
// mix, one role opened on its pair list, and the weekly spend by stage.
// The seed carries no priced, keyed dispatches, so GET
// /api/v1/stats/role-costs is answered from role-costs.fixture.json -- a
// 30-day snapshot of the dev store's response, the prototype's own data.
// The clock is pinned to FIXTURE_NOW because the Period chips end at now.
const FIXTURE = JSON.parse(readFileSync(new URL("./role-costs.fixture.json", import.meta.url), "utf8"));

test("role costs at rest and one role opened", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await page.clock.setFixedTime(FIXTURE_NOW);
  await page.route("**/api/v1/stats/role-costs?*", (route) => route.fulfill({ json: FIXTURE }));
  await page.goto("/#/role-costs");
  await expect(page.getByRole("link", { name: "Role costs", exact: true })).toHaveAttribute("aria-current", "page");
  const table = page.getByRole("table", { name: "Spend and agent time per role" });
  await expect(table.getByRole("row", { name: /Writes the code/ })).toBeVisible();
  await expect(page).toHaveScreenshot("role-costs.png", { fullPage: true });
  await table.getByRole("row", { name: /Writes the code/ }).click();
  await expect(page.getByTestId("role-pairs")).toBeVisible();
  await expect(page).toHaveScreenshot("role-costs-open.png", { fullPage: true });
});
