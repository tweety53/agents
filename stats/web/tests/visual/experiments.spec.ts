import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, PINNED_QUERY } from "./support";

// The Experiments view (#/experiments, Experiments.tsx) over
// cmd/uitest-seed's seedExperiments: the open combined experiment
// "planning on xhigh" and the closed "implementer sonnet/flow-low". The
// view reads no period; the query is pinned so the shell's URL is too.
// experiments.spec.ts.mockups pairs each capture with its frame of the
// approved prototype, .superpowers/sdd/experiments-prototype.html.
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
});
