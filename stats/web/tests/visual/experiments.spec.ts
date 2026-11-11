import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, PINNED_QUERY } from "./support";

// The Experiments view (#/experiments, Experiments.tsx) over
// cmd/uitest-seed's seedExperiments: the open combined experiment
// "planning on xhigh", the closed "implementer sonnet/flow-low" and the
// closed main-cell "main sonnet@low" (seedMainExperiment). The
// view reads no period; the query is pinned so the shell's URL is too.
// experiments.spec.ts.mockups pairs each capture with its frame of the
// approved prototype, .superpowers/sdd/experiments-prototype.html; the
// roles-panel captures are compared against .superpowers/sdd/roles-prototype.html.
async function open(page: import("@playwright/test").Page) {
  await page.goto(`/#/experiments?${PINNED_QUERY}`);
  await expect(page.getByRole("link", { name: "Experiments", exact: true })).toHaveAttribute("aria-current", "page");
  await expect(page.getByText("Loading…")).toHaveCount(0);
  await expect(page.getByText(/^Runs counted:/)).toBeVisible();
}

test.describe("experiments", () => {
  test("experiments open experiment", async ({ page }) => {
    await open(page);
    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
    // Open experiments first: the open one leads the list and opens selected.
    await expect(page.getByRole("region", { name: "Experiments", exact: true }).getByRole("button").first()).toContainText("planning on xhigh");
    await expect(page.getByRole("button", { name: /planning on xhigh/ })).toHaveClass(/\bon\b/);
    await expect(page.getByText("Loading…")).toHaveCount(0);
    // The prototype's measures: a 320px list column, 12px gaps, a 1.6rem
    // stat value and 320×150 charts.
    await expect(page.getByRole("region", { name: "Experiments", exact: true })).toHaveCSS("width", "320px");
    await expect(page.locator(".experiments")).toHaveCSS("row-gap", "12px");
    await expect(page.locator(".experiments .stat .v").first()).toHaveCSS("font-size", "25.6px");
    await expect(page.getByRole("region", { name: "Summary" })).toContainText("Open — the closing agent writes the summary.");
    await expect(page).toHaveScreenshot("experiments-open.png", { fullPage: true });
  });

  test("experiments closed experiment", async ({ page }) => {
    await open(page);
    await page.getByRole("button", { name: /implementer sonnet\/flow-low/ }).click();
    await expect(page.getByRole("region", { name: "Summary" })).toContainText("Negative. Sonnet implementers cut tokens 22%");
    await expect(page.getByText("Loading…")).toHaveCount(0);
    await expect(page).toHaveScreenshot("experiments-closed.png", { fullPage: true });
  });

  test("experiments main-cell metric cards", async ({ page }) => {
    await open(page);
    await page.getByRole("button", { name: /main sonnet@low/ }).click();
    await expect(page.getByRole("region", { name: "Summary" })).toContainText("Positive. Sonnet-low main sessions");
    await expect(page.getByText("Loading…")).toHaveCount(0);
    const cards = page.locator("section.stat");
    const selfReview = page.getByRole("region", { name: "Self-review Crit+Imp", exact: true });
    await expect(selfReview).toBeVisible();
    // The new card sits right after Verify reds.
    const titles = await cards.locator("h3.ttl").allTextContents();
    expect(titles.indexOf("Self-review Crit+Imp")).toBe(titles.indexOf("Verify reds") + 1);
    await expect(page).toHaveScreenshot("experiments-main.png", { fullPage: true });
  });

  test("experiments chart point hover and focus", async ({ page }) => {
    await open(page);
    await page.getByRole("button", { name: /planning on xhigh/ }).click();
    await expect(page.getByText("Loading…")).toHaveCount(0);
    const chart = page.getByRole("img", { name: /^Review rounds \/ change, weekly/ });
    const point = chart.locator("circle.pt").first();
    await point.hover();
    await expect(point).toHaveCSS("r", "5px");
    await expect(page).toHaveScreenshot("experiments-hover.png", { fullPage: true });
    await page.mouse.move(0, 0);
    await point.focus();
    await expect(point).toBeFocused();
    await expect(point).toHaveCSS("r", "5px");
    await expect(page).toHaveScreenshot("experiments-focus.png", { fullPage: true });
  });

  test("experiments roles panel, pair editing draft", async ({ page }) => {
    // Drives the editor without saving, so the shared fixture is unchanged:
    // a malformed draft disables Add, a valid one enables it, adding a chip
    // and removing another enables the row's Save.
    await open(page);
    const roles = page.getByRole("region", { name: "Current roles" });
    const save = roles.getByRole("button", { name: "Save implementer" });
    const row = roles;
    await expect(save).toHaveCount(0);
    await row.getByRole("button", { name: "Choose explicit allow pairs for implementer" }).click();
    const input = row.getByRole("searchbox", { name: "Add explicit allow pair to implementer" });
    await input.fill("not a pair");
    await expect(row.getByText(/^Add “/)).toHaveCount(0);
    await expect(page).toHaveScreenshot("experiments-edit-invalid.png", { fullPage: true });
    await input.fill("opus/flow-high@xhigh");
    await input.press("Enter");
    await expect(row.getByRole("checkbox", { name: "opus/flow-high@xhigh" })).toBeChecked();
    await page.keyboard.press("Escape");
    await expect(row.getByRole("searchbox")).toHaveCount(0);
    await row.getByRole("button", { name: "Choose main pairs for implementer" }).click();
    await row.getByRole("checkbox", { name: "opus/flow-low" }).first().click();
    await page.keyboard.press("Escape");
    await expect(save).toBeVisible();
    await page.mouse.move(0, 0);
    await expect(page).toHaveScreenshot("experiments-edit-dirty.png", { fullPage: true });
  });

  test("experiments roles panel, row hover and field focus", async ({ page }) => {
    await open(page);
    const roles = page.getByRole("region", { name: "Current roles" });
    // The prototype's measures: a 12px table, 2px 8px cells.
    await expect(roles.getByRole("table")).toHaveCSS("font-size", "12px");
    await expect(roles.getByRole("cell").first()).toHaveCSS("padding", "2px 8px");
    await expect(roles).toHaveScreenshot("experiments-roles.png");
    const designer = roles.getByRole("row", { name: /plan-designer/ });
    await designer.hover();
    await expect(roles).toHaveScreenshot("experiments-roles-hover.png");
    await page.mouse.move(0, 0);
    const field = designer.getByRole("button", { name: "Choose main pairs for plan-designer" });
    await field.focus();
    await expect(field).toBeFocused();
    await expect(roles).toHaveScreenshot("experiments-roles-focus.png");
  });
});
