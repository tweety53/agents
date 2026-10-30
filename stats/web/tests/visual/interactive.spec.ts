import { expect, test } from "@playwright/test";
import { bodyBackground, DARK_SURFACE_1, FIXTURE_NOW, PINNED_QUERY } from "./support";

// A /flow-interactive session on the live state board and its run view
// (KAN-1061) over cmd/uitest-seed's seedInteractive: the running session
// kan-1072-… (request 7 in progress) and the waiting one kan-1073-….
// interactive.spec.ts.mockups pairs each capture with its frame of the
// approved prototype, .superpowers/sdd/planner/prototype-board.html. Every
// capture is the viewport (the run view the full page), never a clipped element.
const RUNNING = "kan-1072-interactive-session-2026-10-30";
const WAITING = "kan-1073-interactive-session-2026-10-30";

test.beforeEach(async ({ page }) => {
  await page.clock.setFixedTime(FIXTURE_NOW);
});

async function board(page: import("@playwright/test").Page) {
  await page.goto(`/#/state-board?${PINNED_QUERY}`);
  await expect(page.getByRole("heading", { name: "Live state board" })).toBeVisible();
  await expect(page.getByRole("button", { name: RUNNING })).toBeVisible();
  expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
}

test.describe("interactive session", () => {
  test("interactive running card", async ({ page }) => {
    await board(page);
    const card = page.getByRole("button", { name: RUNNING });
    await expect(card.getByTestId("running-task")).toHaveText("Request 7: Show the stage's model on the Agents row");
    await expect(card).toContainText("requests 6/8 · $4.12");
    await expect(page).toHaveScreenshot("interactive-running-card.png");
  });

  test("interactive waiting card", async ({ page }) => {
    await board(page);
    const card = page.getByRole("button", { name: WAITING });
    await expect(card).toContainText("Waiting for the next request");
    await expect(card.getByTestId("running-task")).toHaveText("Last: Request 3 landed — Hide the empty streak row");
    await expect(page).toHaveScreenshot("interactive-waiting-card.png");
  });

  test("interactive running card hovered", async ({ page }) => {
    await board(page);
    const card = page.getByRole("button", { name: RUNNING });
    await card.hover();
    await expect(card).toHaveCSS("border-top-color", "rgb(255, 152, 48)");
    await expect(page).toHaveScreenshot("interactive-running-card-hover.png");
  });

  test("interactive running card focused from the keyboard", async ({ page }) => {
    await board(page);
    const card = page.getByRole("button", { name: RUNNING });
    // Focus the card before it, then Tab onto this one: a keyboard focus.
    await page.getByRole("button", { name: RUNNING }).evaluate((el) => (el.previousElementSibling as HTMLElement | null)?.focus());
    await page.keyboard.press("Tab");
    await expect(card).toBeFocused();
    await expect(card).toHaveCSS("border-top-color", "rgb(255, 152, 48)");
    await expect(page).toHaveScreenshot("interactive-running-card-focus.png");
  });

  test("interactive run view", async ({ page }) => {
    await page.goto(`/#/run/uitest-alpha/${RUNNING}`);
    await expect(page.getByRole("heading", { name: RUNNING })).toBeVisible();
    await expect(page.getByRole("region", { name: "Requests" })).toContainText("6/8");
    await expect(page.getByRole("region", { name: "Main" })).toContainText("Request 7 · 4m");
    await expect(page.getByRole("region", { name: "Stage runs" }).getByRole("columnheader").first()).toHaveText("Request");
    await expect(page.getByText("Loading…")).toHaveCount(0);
    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
    await expect(page).toHaveScreenshot("interactive-run-view.png", { fullPage: true });
  });
});
