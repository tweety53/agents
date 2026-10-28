import { expect, test } from "@playwright/test";
import { FIXTURE_NOW, PINNED_QUERY } from "./support";

// The opened run's own surfaces (kan-1048): the session-grouped stage-run
// table with its dispatch toggle, the stage-lane timeline, and the absence
// of any filter control.
//
// uitest-beta/kan-201-refactor-thing (cmd/uitest-seed/seed.go): two stage
// runs in two sessions --
//
//   uitest-session-201a, /flow, "SDD + TDD per task", started
//     2026-08-14T07:00:00Z, 5 min, cost_usd 2.75, claude-opus-5, effort high
//   uitest-session-201b, /flow-fast, "integrate", started
//     2026-08-15T08:00:00Z, 5 min, cost_usd 0.028, claude-sonnet-5, effort low
const RUN_DETAIL_URL = `/#/run/uitest-beta/kan-201-refactor-thing?${PINNED_QUERY}`;

test.beforeEach(async ({ page }) => {
  await page.clock.setFixedTime(FIXTURE_NOW);
});

test.describe("stage runs", () => {
  test("group by session and phase, each row carrying its model version and effort", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    const table = page.getByRole("region", { name: "Stage runs" }).getByRole("table");
    await expect(table.locator("tr.lb-group")).toHaveText([
      "Session 1 · other07:00–07:05 · 5m · $2.75",
      "Session 2 · landing08:00–08:05 · 5m · $0.03",
    ]);
    await expect(table.getByRole("row", { name: /SDD \+ TDD per task/ })).toContainText("opus-5 · high");
    await expect(table.getByRole("row", { name: /integrate/ })).toContainText("sonnet-5 · low");
  });

  test("carry no filter control of any kind", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    await expect(page.getByRole("heading", { name: "kan-201-refactor-thing" })).toBeVisible();
    await expect(page.getByRole("region", { name: "Stage runs" })).toBeVisible();
    await expect(page.getByRole("combobox")).toHaveCount(0);
    await expect(page.getByRole("searchbox")).toHaveCount(0);
    await expect(page.getByRole("textbox")).toHaveCount(0);
    await expect(page.getByRole("toolbar")).toHaveCount(0);
  });
});

test.describe("timeline", () => {
  test("draws one lane per stage, coloured by outcome", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    const svg = page.getByRole("img", { name: "Stage runs over time" });
    await expect(svg.locator("text.lb-lane")).toHaveText(["SDD + TDD per task", "integrate"]);
    const bars = svg.getByTestId("timeline-bar");
    await expect(bars).toHaveCount(2);
    for (const bar of await bars.all()) await expect(bar).toHaveAttribute("fill", "var(--status-ok)");
    // Both runs ended: no now marker.
    await expect(svg.locator("line.lb-now")).toHaveCount(0);

    await bars.last().hover();
    await expect(page.getByRole("tooltip")).toHaveText("integrateSession 2completed · 5m · $0.03");
  });
});
