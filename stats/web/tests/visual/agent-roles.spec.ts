import { expect, test } from "@playwright/test";
import { PINNED_QUERY } from "./support";

// The Agent roles view (#/agent-roles, AgentRoles.tsx): every role's pairs
// in force, grouped by pipeline stage. The view reads no period; the query
// is pinned so the shell's URL is too. Captures are compared against
// .superpowers/sdd/roles-prototype.html.
async function open(page: import("@playwright/test").Page) {
  await page.goto(`/#/agent-roles?${PINNED_QUERY}`);
  await expect(page.getByRole("link", { name: "Agent roles", exact: true })).toHaveAttribute("aria-current", "page");
  await expect(page.getByText("Loading…")).toHaveCount(0);
  await expect(page.getByRole("region", { name: "Current roles" })).toBeVisible();
}

test.describe("agent roles", () => {
  test("agent roles, pair editing draft", async ({ page }) => {
    // Drives the editor without saving, so the shared fixture is unchanged:
    // a malformed draft disables Add, a valid one enables it, adding a chip
    // and removing another enables the row's Save.
    await open(page);
    const roles = page.getByRole("region", { name: "Current roles" });
    const save = roles.getByRole("button", { name: "Save" });
    const row = roles;
    await expect(save).toHaveCount(0);
    await row.getByRole("button", { name: "Choose explicit allow pairs for implementer" }).click();
    const input = row.getByRole("searchbox", { name: "Add explicit allow pair to implementer" });
    await input.fill("not a pair");
    await expect(row.getByText(/^Add “/)).toHaveCount(0);
    await expect(page).toHaveScreenshot("agent-roles-edit-invalid.png", { fullPage: true });
    // The menu shuts on a window scroll or resize, and a full-page capture
    // leaves it shut: reopen it.
    await page.keyboard.press("Escape");
    await row.getByRole("button", { name: "Choose explicit allow pairs for implementer" }).click();
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
    await expect(page).toHaveScreenshot("agent-roles-edit-dirty.png", { fullPage: true });
  });

  test("agent roles, row hover and field focus", async ({ page }) => {
    await open(page);
    const roles = page.getByRole("region", { name: "Current roles" });
    // The prototype's measures: a 12px table, 2px 8px cells.
    await expect(roles.getByRole("table")).toHaveCSS("font-size", "12px");
    await expect(roles.getByRole("cell").first()).toHaveCSS("padding", "2px 8px");
    await expect(roles).toHaveScreenshot("agent-roles.png");
    const designer = roles.getByRole("row", { name: /plan-designer/ });
    await designer.hover();
    await expect(roles).toHaveScreenshot("agent-roles-hover.png");
    await page.mouse.move(0, 0);
    const field = designer.getByRole("button", { name: "Choose main pairs for plan-designer" });
    await field.focus();
    await expect(field).toBeFocused();
    await expect(roles).toHaveScreenshot("agent-roles-focus.png");
  });
});
