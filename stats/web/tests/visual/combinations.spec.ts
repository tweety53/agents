import { expect, test } from "@playwright/test";
import { DARK_SURFACE_1, FIXTURE_NOW, bodyBackground } from "./support";

// task 12's own required combinations (KAN-171 part 4) -- each control
// above is correct alone; KAN-16 shipped when a *composition* of correct
// parts was wrong, so each spec below proves the composed behaviour, not
// just the two halves in isolation, and takes a screenshot as well.

test.describe("project chip plus state tile", () => {
  test("a Project chip and a state tile narrow the board together, not just either alone", async ({ page }) => {
    await page.clock.setFixedTime(FIXTURE_NOW);
    await page.goto("/#/state-board");
    const table = page.getByRole("region", { name: "Changes" }).getByRole("table");
    await expect(table.getByRole("link", { name: "kan-102-fix-widget-bug" })).toBeVisible();

    // The chip alone keeps both uitest-alpha changes, in two states.
    await page.getByRole("group", { name: "Project" }).getByRole("button", { name: /^uitest-alpha/ }).click();
    await expect(table.getByRole("link", { name: "kan-101-add-widget" })).toBeVisible();
    await expect(table.getByRole("link", { name: "kan-102-fix-widget-bug" })).toBeVisible();

    // The tile narrows further within the chip's project.
    await page.getByRole("button", { name: /^Planning & implementing/ }).click();
    await expect(table.getByRole("link", { name: "kan-101-add-widget" })).toBeVisible();
    await expect(table.getByRole("link", { name: "kan-102-fix-widget-bug" })).toHaveCount(0);
    await expect(table.getByRole("link", { name: "kan-201-refactor-thing" })).toHaveCount(0);

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
    await expect(page).toHaveScreenshot("combination-chips-plus-tile.png", { fullPage: true });
  });
});
