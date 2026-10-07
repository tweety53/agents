import { expect, test } from "@playwright/test";
import { PINNED_QUERY, RUNS_QUERY } from "./support";

// The full app suite (skills/flow-contracts/project-configuration-visual.md):
// one full-page capture per screen the app has, enumerated from App.tsx's
// VIEW_COMPONENTS and its run-detail route, never from one change's diff.
// Each capture first asserts the screen is the one on show -- the nav link's
// aria-current, or the run's own heading -- so a missed navigation fails
// here instead of becoming the baseline. Periods are pinned for the reason
// baseline.spec.ts's DASHBOARD_URL comment gives; the runs view reads its
// own isolated fixture window (support.ts's RUNS_QUERY).
const VIEWS: { view: string; label: string; query: string }[] = [
  { view: "state-board", label: "Live state board", query: PINNED_QUERY },
  { view: "stage-leaderboard", label: "Stage leaderboard", query: PINNED_QUERY },
  { view: "trend", label: "Trend over time", query: PINNED_QUERY },
  { view: "cache-efficiency", label: "Cache efficiency", query: PINNED_QUERY },
  { view: "reviewers", label: "Reviewers", query: PINNED_QUERY },
  { view: "decisions", label: "Decisions", query: PINNED_QUERY },
  { view: "runs", label: "Runs", query: RUNS_QUERY },
  { view: "flow-health", label: "Flow health", query: PINNED_QUERY },
  { view: "self-review", label: "Self-review fixes", query: PINNED_QUERY },
];

test.describe("full app suite", () => {
  for (const { view, label, query } of VIEWS) {
    test(`${view} screen`, async ({ page }) => {
      await page.goto(`/#/${view}?${query}`);
      await expect(page.getByRole("link", { name: label, exact: true })).toHaveAttribute("aria-current", "page");
      await expect(page).toHaveScreenshot(`full-${view}.png`, { fullPage: true });
    });
  }

  test("run detail screen", async ({ page }) => {
    await page.goto(`/#/run/uitest-beta/kan-201-refactor-thing?${PINNED_QUERY}`);
    await expect(page.getByRole("heading", { name: "kan-201-refactor-thing" })).toBeVisible();
    await expect(page).toHaveScreenshot("full-run-detail.png", { fullPage: true });
  });
});
