import { expect, test } from "@playwright/test";
import { FIXTURE_NOW, PINNED_QUERY } from "./support";

// The full app suite (skills/flow-contracts/project-configuration-visual.md):
// one full-page capture per screen the app has, enumerated from App.tsx's
// VIEW_COMPONENTS and its run-detail route, never from one change's diff.
// Each capture first asserts the screen is the one on show -- the nav link's
// aria-current, or the run's own heading -- so a missed navigation fails
// here instead of becoming the baseline. Periods are pinned for the reason
// baseline.spec.ts's DASHBOARD_URL comment gives.
const VIEWS: { view: string; label: string; query: string }[] = [
  { view: "state-board", label: "Live state board", query: PINNED_QUERY },
  { view: "stage-leaderboard", label: "Stage leaderboard", query: PINNED_QUERY },
  { view: "trend", label: "Trend over time", query: PINNED_QUERY },
  { view: "reviewers", label: "Reviewers", query: PINNED_QUERY },
  { view: "decisions", label: "Decisions", query: PINNED_QUERY },
  { view: "flow-health", label: "Flow health", query: PINNED_QUERY },
  { view: "self-review", label: "Self-review fixes", query: PINNED_QUERY },
  // The eight weeks cmd/uitest-seed's seedReviewDynamics fills.
  { view: "review-dynamics", label: "Review dynamics", query: "from=2026-05-04T00%3A00%3A00.000Z&to=2026-06-29T00%3A00%3A00.000Z" },
];

test.describe("full app suite", () => {
  for (const { view, label, query } of VIEWS) {
    test(`${view} screen`, async ({ page }) => {
      // The live state board reads the browser clock (kan-1048).
      await page.clock.setFixedTime(FIXTURE_NOW);
      await page.goto(`/#/${view}?${query}`);
      await expect(page.getByRole("link", { name: label, exact: true })).toHaveAttribute("aria-current", "page");
      // Every panel's "Loading…" status gone first: the nav link turns
      // current before any fetch resolves, so without this wait the capture
      // starts mid-load and, under verify's concurrent load, the page kept
      // changing until toHaveScreenshot's stability wait ran out (kan-1038).
      await expect(page.getByText("Loading…")).toHaveCount(0);
      await expect(page).toHaveScreenshot(`full-${view}.png`, { fullPage: true });
    });
  }

  // Weekly costs' 8-week table reads the browser clock, so it is pinned
  // the way weekly-costs.spec.ts pins it.
  test("weekly-costs screen", async ({ page }) => {
    await page.clock.install({ time: new Date("2026-08-16T12:00:00Z") });
    await page.goto(`/#/weekly-costs?${PINNED_QUERY}`);
    await expect(page.getByRole("link", { name: "Weekly costs", exact: true })).toHaveAttribute("aria-current", "page");
    await expect(page.getByText("Loading…")).toHaveCount(0);
    await expect(page).toHaveScreenshot("full-weekly-costs.png", { fullPage: true });
  });

  test("run detail screen", async ({ page }) => {
    await page.clock.setFixedTime(FIXTURE_NOW);
    await page.goto(`/#/run/uitest-beta/kan-201-refactor-thing?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "kan-201-refactor-thing" })).toBeVisible();
    await expect(page.getByText("Loading…")).toHaveCount(0);
    await expect(page).toHaveScreenshot("full-run-detail.png", { fullPage: true });
  });
});
