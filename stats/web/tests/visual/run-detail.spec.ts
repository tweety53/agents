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
      "Session 1 · Other07:00–07:05 · 5m · $2.75",
      "Session 2 · Other08:00–08:05 · 5m · $0.03",
    ]);
    await expect(table.getByRole("row", { name: /SDD \+ TDD per task/ })).toContainText("opus 5 · high");
    await expect(table.getByRole("row", { name: /integrate/ })).toContainText("sonnet 5 · low");
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
    // The timeline is collapsed by default (32701733).
    await page.getByRole("button", { name: "Show timeline" }).click();
    const svg = page.getByRole("img", { name: "Stage runs over time" });
    // Each lane label carries its raw stage key in a <title> child too.
    await expect(svg.locator("text.lb-lane")).toHaveText([/SDD \+ TDD per task$/, /integrate$/]);
    const bars = svg.getByTestId("timeline-bar");
    await expect(bars).toHaveCount(2);
    for (const bar of await bars.all()) await expect(bar).toHaveAttribute("fill", "var(--status-ok)");
    // Both runs ended: no now marker.
    await expect(svg.locator("line.lb-now")).toHaveCount(0);

    await bars.last().hover();
    await expect(page.getByRole("tooltip")).toHaveText("integrate integrateSession 2completed · 5m · $0.03");
  });

  // The visual verifier's timeline-tooltip strip caught the tooltip one paint
  // behind the pointer (kan-1048 in-run fix 1); the prototype hides it in the
  // same handler. Read on the running clock: the next animation frame after
  // the pointer leaves the bar has no tooltip.
  test("the tooltip is gone in the paint the pointer leaves the bar", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    await page.getByRole("button", { name: "Show timeline" }).click();
    const bar = page.getByRole("img", { name: "Stage runs over time" }).getByTestId("timeline-bar").last();
    await bar.hover();
    await expect(page.getByRole("tooltip")).toBeVisible();
    const box = (await bar.boundingBox())!;
    // Arm a probe that reads the DOM in the first animation frame after the
    // pointer leaves, then move the real pointer off the bar.
    await page.evaluate(() => {
      const w = window as unknown as { tipAfterLeave?: boolean };
      document.addEventListener(
        "mouseout",
        () => requestAnimationFrame(() => (w.tipAfterLeave = document.querySelector('[role="tooltip"]') !== null)),
        { once: true },
      );
    });
    await page.mouse.move(box.x + box.width / 2, box.y - 40);
    await page.waitForFunction(() => (window as unknown as { tipAfterLeave?: boolean }).tipAfterLeave !== undefined);
    const shown = await page.evaluate(() => (window as unknown as { tipAfterLeave?: boolean }).tipAfterLeave);
    expect(shown).toBe(false);
  });
});
