import { expect, test } from "@playwright/test";
import { DARK_SURFACE_1, PINNED_FROM_INPUT, PINNED_QUERY, PINNED_TO_INPUT, bodyBackground } from "./support";

// task 12's own required combinations (KAN-171 part 4) -- each control
// above is correct alone; KAN-16 shipped when a *composition* of correct
// parts was wrong, so each spec below proves the composed behaviour, not
// just the two halves in isolation. Every spec takes a screenshot in
// addition to its DOM assertions -- composition is exactly the defect
// class a screenshot alone would have caught (proposal.md), so the same
// discipline baseline.spec.ts already applies to single views applies
// here to their combinations.

test.describe("filter plus search", () => {
  test("a FilterBar column filter and a search term narrow the table together, not just either alone", async ({
    page,
  }) => {
    await page.goto(`/#/state-board?${PINNED_QUERY}`);
    const tablePanel = page.locator('section.panel[aria-label="Every open change"]');

    // Search alone: "widget" matches both uitest-alpha changes (their
    // names both contain it) but not uitest-beta's.
    await page.getByLabel("Search").fill("widget");
    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-102-fix-widget-bug" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-201-refactor-thing" })).toHaveCount(0);

    // Adding the State filter must narrow *further*, not replace or
    // ignore the search already in effect -- the composition this test
    // exists to prove, since either control narrowing the *same* two
    // rows down to one would not distinguish "combined" from "only the
    // filter mattered".
    await page.getByLabel("Filter by State").selectOption("IN_PROGRESS");
    await expect(tablePanel.getByRole("cell", { name: "kan-102-fix-widget-bug" })).toBeVisible();
    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toHaveCount(0);
    await expect(tablePanel.getByRole("cell", { name: "kan-201-refactor-thing" })).toHaveCount(0);

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
    await expect(page).toHaveScreenshot("combination-filter-plus-search.png", { fullPage: true });
  });
});

test.describe("period change plus column sort", () => {
  test("changing the period reloads the table, and the sort already applied still governs its new rows", async ({
    page,
  }) => {
    // A period entirely before the fixture's data -- see support.ts's own
    // EMPTY_QUERY comment -- reached here via the same URL shape rather
    // than the constant itself, so this spec's own intent (start empty,
    // then change the period from inside the browser) reads without a
    // second file open.
    await page.goto("/#/state-board?from=2020-01-01T00%3A00%3A00.000Z&to=2020-01-02T00%3A00%3A00.000Z");
    const tablePanel = page.locator('section.panel[aria-label="Every open change"]');
    await expect(tablePanel.getByTestId("not-recorded")).toBeVisible();

    // DataTable itself is never mounted for a `recorded: false` period
    // (Panel.tsx's own branch never calls `children`), so this locator
    // resolves nothing until the period change below populates the
    // panel -- Playwright locators are lazy, so declaring it now and
    // clicking it once the table exists is what proves the sort applies
    // to a table sorted for the very first time, not one re-sorted after
    // already holding rows.
    const changeHeader = tablePanel.getByRole("button", { name: /^Change/ });
    await page.getByLabel("Period start").fill(PINNED_FROM_INPUT);
    await page.getByLabel("Period end").fill(PINNED_TO_INPUT);

    await expect(tablePanel.getByRole("cell", { name: "kan-101-add-widget" })).toBeVisible();
    // The first click on a column never sorted before goes ascending
    // (DataTable.tsx's toggleSort: a fresh `sortKey` always starts at
    // "asc"), so this is the freshly-mounted table's very first sort, not
    // a toggle away from a prior direction.
    await changeHeader.click();
    await expect(changeHeader).toHaveText("Change ▲");
    const changeCells = tablePanel.locator("tbody tr td:nth-child(2) a");
    await expect(changeCells).toHaveText(["kan-101-add-widget", "kan-102-fix-widget-bug", "kan-201-refactor-thing"]);

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
    await expect(page).toHaveScreenshot("combination-period-plus-sort.png", { fullPage: true });
  });
});

test.describe("ModelVariable disabled on state-board, enabled elsewhere", () => {
  test("the same control is correctly disabled on one route and correctly functional on the next", async ({
    page,
  }) => {
    await page.goto(`/#/state-board?${PINNED_QUERY}`);
    await expect(page.getByLabel("Model")).toBeDisabled();

    // In-app navigation (a real hash change, not a fresh page.goto) is
    // the composition under test: App.tsx's `modelDisabled` is derived
    // from the *route*, and the period must survive the navigation
    // unchanged (App.tsx's own onUrlPeriod comment on why a hashchange
    // carrying no query must not reset it) -- proving both in one
    // sequence is what a page.goto straight to stage-leaderboard would
    // not exercise.
    await page.getByRole("link", { name: "Stage leaderboard" }).click();
    await expect(page.getByRole("heading", { name: "Stage leaderboard" })).toBeVisible();
    const model = page.getByLabel("Model");
    await expect(model).toBeEnabled();
    await expect(model.locator("option")).toHaveText(["All models", "claude-opus-5", "claude-sonnet-5"]);

    const tablePanel = page.locator('section.panel[aria-label="Every costed stage"]');
    await expect(tablePanel.getByRole("cell", { name: "/myflow-finish", exact: true })).toBeVisible();

    await model.selectOption({ label: "claude-opus-5" });

    // GET /api/v1/stats/stage-leaderboard?model=claude-opus-5 (verified
    // against the live UI-test stack) drops the "/myflow-finish
    // integrate" row entirely -- it ran only under claude-sonnet-5 -- and
    // narrows "/myflow-do SDD + TDD per task" to its one opus run.
    await expect(tablePanel.getByRole("cell", { name: "/myflow-finish", exact: true })).toHaveCount(0);
    await expect(tablePanel.getByRole("cell", { name: "/myflow-do", exact: true })).toBeVisible();
    const stagesPanel = page.locator('section.panel[aria-label="Stages"]');
    await expect(stagesPanel.locator('[data-testid="measured"]')).toHaveText("1");

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
    await expect(page).toHaveScreenshot("combination-model-disabled-composition.png", { fullPage: true });
  });
});
