import { expect, test } from "@playwright/test";
import { PINNED_QUERY } from "./support";

// The run-detail route's own controls -- StageRunTable's sort/filter and
// its per-row expand toggle, and StageTimeline -- covering the gap a
// review of the first task-12 commit found: baseline.spec.ts's own
// run-detail spec and controls.spec.ts's ChangeVariable spec both only
// ever asserted the page loaded, never interacted with anything on it.
//
// uitest-beta/kan-201-refactor-thing is the fixture's own richest change
// (cmd/uitest-seed/seed.go): two stage runs, read live from the UI-test
// stack (GET /api/v1/stage-runs?project=uitest-beta&name=kan-201-refactor-thing):
//
//   stageRunId 2, /myflow-do, "SDD + TDD per task", attempt 1, completed,
//     started 2026-08-14T07:00:00Z, ended 2026-08-14T07:05:00Z (5 min),
//     tokens.main {input: 310000, output: 48000}, cost_usd 2.75,
//     models: {"claude-opus-5": {...}}
//   stageRunId 3, /myflow-finish, "integrate", attempt 1, completed,
//     started 2026-08-15T08:00:00Z, ended 2026-08-15T08:05:00Z (5 min),
//     tokens.main {input: 8000, output: 1200}, cost_usd 0.028,
//     models: {"claude-sonnet-5": {...}}
//
// Neither run's metrics bag records cache tokens, effort, fast_mode or a
// "dispatches" bucket -- the fixture was seeded for the cost/token views,
// not for this table's own richer per-run detail -- so every one of those
// fields is asserted "unavailable" below (Unavailable.tsx's own
// data-testid), which is itself a real, meaningful assertion: a
// regression that turned an absent field into a fabricated zero would
// fail it. Duration and attempt tie between the two runs (both 5 min,
// both attempt 1), so those two columns are not used for order
// assertions below -- a tie is indistinguishable from a broken sort and
// would prove nothing.
const RUN_DETAIL_URL = `/#/run/uitest-beta/kan-201-refactor-thing?${PINNED_QUERY}`;

test.describe("StageRunTable sort buttons", () => {
  test("sorting by Cost reorders the table by its own real cost", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    const tablePanel = page.locator('section.panel[aria-label="Stage runs"]');
    const costHeader = tablePanel.getByRole("button", { name: /^Cost/ });
    const commandCells = tablePanel.locator("table tbody tr td:nth-child(2)");
    await expect(commandCells).toHaveCount(2);

    await costHeader.click();
    await expect(costHeader).toHaveText("Cost ▲");
    await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);

    await costHeader.click();
    await expect(costHeader).toHaveText("Cost ▼");
    await expect(commandCells).toHaveText(["/myflow-do", "/myflow-finish"]);
  });

  test("sorting by Tokens in reorders the table by its own real input tokens", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    const tablePanel = page.locator('section.panel[aria-label="Stage runs"]');
    const tokensHeader = tablePanel.getByRole("button", { name: /^Tokens in/ });
    const commandCells = tablePanel.locator("table tbody tr td:nth-child(2)");

    await tokensHeader.click();
    await expect(tokensHeader).toHaveText("Tokens in ▲");
    // 8,000 (myflow-finish) < 310,000 (myflow-do).
    await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);
  });

  test("sorting by Command reorders alphabetically, not by any numeric column", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    const tablePanel = page.locator('section.panel[aria-label="Stage runs"]');
    const commandHeader = tablePanel.getByRole("button", { name: /^Command/ });
    const commandCells = tablePanel.locator("table tbody tr td:nth-child(2)");

    await commandHeader.click();
    await expect(commandHeader).toHaveText("Command ▲");
    // "/myflow-do" < "/myflow-finish" ("d" < "f") -- ascending on the
    // Command column's own text, unrelated to cost or token order above
    // (which happen to agree here only because the opus run is both
    // pricier and larger; this test's own column is what proves Command
    // is wired to itself, not silently reusing Cost's comparator).
    await expect(commandCells).toHaveText(["/myflow-do", "/myflow-finish"]);

    await commandHeader.click();
    await expect(commandHeader).toHaveText("Command ▼");
    await expect(commandCells).toHaveText(["/myflow-finish", "/myflow-do"]);
  });
});

test.describe("StageRunTable filters", () => {
  test("Filter by Command narrows to the matching stage run", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    const tablePanel = page.locator('section.panel[aria-label="Stage runs"]');
    await expect(tablePanel.getByRole("cell", { name: "/myflow-finish", exact: true })).toBeVisible();

    await page.getByLabel("Filter by Command").selectOption("/myflow-do");

    await expect(tablePanel.getByRole("cell", { name: "/myflow-do", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "/myflow-finish", exact: true })).toHaveCount(0);
  });

  test("Filter by Stage narrows to the matching stage run", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    const tablePanel = page.locator('section.panel[aria-label="Stage runs"]');
    await expect(tablePanel.getByRole("cell", { name: "SDD + TDD per task", exact: true })).toBeVisible();

    await page.getByLabel("Filter by Stage").selectOption("integrate");

    await expect(tablePanel.getByRole("cell", { name: "integrate", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "SDD + TDD per task", exact: true })).toHaveCount(0);
  });

  test("Filter by Model narrows to the run that used that model", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    const tablePanel = page.locator('section.panel[aria-label="Stage runs"]');
    await expect(tablePanel.getByRole("cell", { name: "claude-opus-5", exact: true })).toBeVisible();

    await page.getByLabel("Filter by Model").selectOption("claude-sonnet-5");

    await expect(tablePanel.getByRole("cell", { name: "claude-sonnet-5", exact: true })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "claude-opus-5", exact: true })).toHaveCount(0);
  });
});

test.describe("StageRunTable expand toggle", () => {
  test("opens and closes the per-row metrics detail, revealing content no other view shows", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    const tablePanel = page.locator('section.panel[aria-label="Stage runs"]');
    const row = tablePanel.locator("table tbody tr", { hasText: "/myflow-do" }).first();
    // Located by its own `aria-expanded` attribute, not by its accessible
    // name -- DataTable.tsx flips that name between "Show run details"
    // and "Hide run details" on every click (DataTable.tsx:234), so a
    // name-based locator would stop resolving to this same button the
    // instant it is clicked.
    const toggle = row.locator("button[aria-expanded]");
    await expect(toggle).toHaveAccessibleName("Show run details");

    // The detail row does not exist in the DOM at all until opened --
    // DataTable.tsx only renders it when `renderDetail && isOpen`
    // (DataTable.tsx:245) -- so its absence beforehand is itself part of
    // what this test proves, not just the presence afterward.
    await expect(tablePanel.locator(".stage-run-detail")).toHaveCount(0);

    await toggle.click();
    await expect(toggle).toHaveAttribute("aria-expanded", "true");
    await expect(toggle).toHaveAccessibleName("Hide run details");

    const detail = tablePanel.locator(".stage-run-detail");
    await expect(detail).toBeVisible();
    // Every field this run's own metrics bag never recorded reads as
    // "unavailable" -- not a fabricated zero or a blank cell (metrics.ts's
    // own absence-is-never-zero rule, applied to this run's own bag,
    // which the header comment above confirms carries none of these).
    await expect(detail.locator("dt", { hasText: "Effort" }).locator("xpath=following-sibling::dd[1]").getByTestId("unavailable")).toBeVisible();
    await expect(detail.locator("dt", { hasText: "Fast mode" }).locator("xpath=following-sibling::dd[1]").getByTestId("unavailable")).toBeVisible();
    await expect(detail.locator("dt", { hasText: "Sidechain tokens" }).locator("xpath=following-sibling::dd[1]").getByTestId("unavailable")).toBeVisible();
    // Main tokens IS measured -- tokens.main {input: 310000, output:
    // 48000}, summed by readMainTokens (metrics.ts) to 358000.
    await expect(
      detail.locator("dt", { hasText: "Main tokens" }).locator("xpath=following-sibling::dd[1]").getByTestId("measured"),
    ).toHaveText("358,000");
    // No "dispatches" bucket in this run's metrics bag -- the dispatch
    // breakdown section must not render at all (StageRunTable.tsx:289's
    // own `dispatchRows.length > 0` guard), not render empty.
    await expect(detail.locator(".stage-run-dispatches")).toHaveCount(0);

    await toggle.click();
    await expect(toggle).toHaveAttribute("aria-expanded", "false");
    await expect(toggle).toHaveAccessibleName("Show run details");
    await expect(tablePanel.locator(".stage-run-detail")).toHaveCount(0);
  });
});

test.describe("StageTimeline", () => {
  test("draws one bar per stage run, labelled by command, with the run's own outcome in its tooltip", async ({
    page,
  }) => {
    await page.goto(RUN_DETAIL_URL);
    const timelinePanel = page.locator('section.panel[aria-label="Stage timeline"]');
    const svg = timelinePanel.locator("svg.stage-timeline");
    await expect(svg).toHaveAttribute("aria-label", "Stage runs across the change's span, grouped by command");

    // One lane per distinct command (StageTimeline.tsx's own
    // commandOrder), not one per run -- both fixture runs belong to
    // different commands here, so this also indirectly proves lane
    // grouping runs at all rather than always rendering exactly one lane.
    await expect(svg.locator(".stage-timeline-label")).toHaveText(["/myflow-do", "/myflow-finish"]);

    // Both fixture runs have an endedAt, so neither bar is the "open"
    // variant -- the "still running" hatched bar this component draws
    // for a run with no end instant (StageTimeline.tsx's own header
    // comment) is exercised by nothing in this fixture, and is reported
    // as not covered rather than faked with a synthetic open run.
    const bars = svg.locator('rect[data-testid="stage-timeline-bar"]');
    await expect(bars).toHaveCount(2);
    await expect(svg.locator('rect[data-testid="stage-timeline-open"]')).toHaveCount(0);

    const titles = await bars.locator("title").allTextContents();
    expect(titles).toContain("SDD + TDD per task (attempt 1): 5.0 min -- completed");
    expect(titles).toContain("integrate (attempt 1): 5.0 min -- completed");
  });
});
