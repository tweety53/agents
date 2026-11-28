import { expect, type Page, test } from "@playwright/test";
import { FIXTURE_NOW } from "./support";

// Every control the stats SPA exposes, exercised on its own. Each spec
// asserts the control's effect in the DOM (a row appearing or
// disappearing, an attribute flipping), never a screenshot alone.
//
// The other views' own chip filters are exercised in their own specs.

// The live state board's own filters (kan-1048): fixture rows inside the
// default 7d at FIXTURE_NOW are kan-101-add-widget (STARTED, running),
// kan-102-fix-widget-bug (IN_PROGRESS, running) and kan-201-refactor-thing
// (FINISHED, its group collapsed).
test.describe("Live state board chips", () => {
  test.beforeEach(async ({ page }) => {
    await page.clock.setFixedTime(FIXTURE_NOW);
    await page.goto("/#/state-board");
  });

  const table = (page: Page) => page.getByRole("region", { name: "Changes" }).getByRole("table");
  const tile = (page: Page, name: RegExp) => page.getByRole("button", { name });

  test("a Project chip narrows the rows and the state tiles", async ({ page }) => {
    await expect(tile(page, /^Integrating & review/)).toContainText("1");
    await page.getByRole("group", { name: "Project" }).getByRole("button", { name: /^uitest-beta/ }).click();
    await expect(tile(page, /^Integrating & review/)).toContainText("0");
    await expect(tile(page, /^Finished/)).toContainText("1");
    await expect(table(page).getByRole("link", { name: "kan-101-add-widget" })).toHaveCount(0);
  });

  test("a state tile filters by its state and opens its collapsed group", async ({ page }) => {
    await expect(table(page).getByRole("link", { name: "kan-201-refactor-thing" })).toHaveCount(0);
    await tile(page, /^Finished/).click();
    await expect(tile(page, /^Finished/)).toHaveAttribute("aria-pressed", "true");
    await expect(table(page).getByRole("link", { name: "kan-201-refactor-thing" })).toBeVisible();
    await expect(table(page).getByRole("link", { name: "kan-101-add-widget" })).toHaveCount(0);
  });

  test("search narrows the rows by change name", async ({ page }) => {
    await page.getByRole("searchbox", { name: "Search changes" }).fill("fix-widget");
    await expect(table(page).getByRole("link", { name: "kan-102-fix-widget-bug" })).toBeVisible();
    await expect(table(page).getByRole("link", { name: "kan-101-add-widget" })).toHaveCount(0);
  });

  test("Updated All widens past seven days, and Clear filters restores the default", async ({ page }) => {
    // Finished changes load on demand: open their group so the widening is measured on loaded rows.
    await page.getByRole("row", { name: /^▾ ?Finished/ }).click();
    await expect(page.getByRole("row", { name: /^▾ ?Finished/ })).not.toContainText("—");
    const finished = await tile(page, /^Finished/).textContent();
    await page.getByRole("group", { name: "Updated" }).getByRole("button", { name: "All" }).click();
    await expect(tile(page, /^Finished/)).not.toHaveText(finished!);
    await page.getByRole("button", { name: "Clear filters" }).click();
    await expect(tile(page, /^Finished/)).toHaveText(finished!);
    await expect(page.getByRole("button", { name: "Clear filters" })).toHaveCount(0);
  });
});
