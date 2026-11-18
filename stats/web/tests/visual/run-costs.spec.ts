import { expect, test, type Page } from "@playwright/test";
import { FIXTURE_NOW } from "./support";

// Captures of the run timings and costs page, the Run costs view and run
// detail's header with its Timings and costs button (KAN-1132), against the
// two runs cmd/uitest-seed's seedRunCosts writes for uitest-beta:
// kan-301-run-costs-a (every stage, a grouped task, a no-tokens reviewer, two
// repos, flaky verify rounds) and kan-302-run-costs-b. The light scheme is the
// one the approved prototype draws. Every capture is full page, and the clock
// is pinned to FIXTURE_NOW because the Run costs Period chips end at now.

const RUN = "kan-301-run-costs-a";
const OTHER = "kan-302-run-costs-b";

test.use({ colorScheme: "light" });

async function open(page: Page, hash: string) {
  await page.setViewportSize({ width: 1280, height: 1024 });
  await page.clock.setFixedTime(FIXTURE_NOW);
  await page.goto(hash);
  await expect(page.getByText("Loading…")).toHaveCount(0);
}

async function openTimings(page: Page) {
  await open(page, `/#/run-costs/uitest-beta/${RUN}`);
  const root = page.getByRole("region", { name: `Run timings and costs: uitest-beta/${RUN}` });
  await expect(root.getByRole("heading", { name: `${RUN} · timings and costs` })).toBeVisible();
  await expect(root.getByRole("link", { name: "← Run detail" })).toBeVisible();
  for (const r of ["Run totals", "Timeline", "Time and cost by", "Data quality", "Counts", "Cost per task", "Cost per verify round"]) {
    await expect(root.getByRole("region", { name: r, exact: true })).toBeVisible();
  }
  // A panel whose figures carry a source badge states its definition in the badge's title.
  for (const r of ["Run totals", "Cost per task", "Cost per verify round"]) {
    await expect(root.getByRole("region", { name: r, exact: true }).locator(".src[title]").first()).toHaveAttribute("title", /\S/);
  }
  await expect(root.getByRole("img", { name: "Run timeline" })).toBeVisible();
  return root;
}

async function openCosts(page: Page) {
  await open(page, "/#/run-costs");
  await expect(page.getByRole("link", { name: "Run costs" })).toHaveAttribute("aria-current", "page");
  const root = page.getByRole("region", { name: "Run costs", exact: true });
  for (const r of ["Time and cost by category, across runs", "Runs", "Compare runs", "Definitions"]) {
    await expect(root.getByRole("region", { name: r, exact: true })).toBeVisible();
  }
  await expect(root.getByRole("region", { name: "Definitions" }).locator(".src").first()).toBeVisible();
  await expect(root.getByRole("checkbox", { name: `Compare ${RUN}` })).toBeVisible();
  return root;
}

async function parkPointer(page: Page) {
  await page.mouse.move(0, 0);
}

test("run timings page at rest", async ({ page }) => {
  await openTimings(page);
  await expect(page).toHaveScreenshot("run-timings.png", { fullPage: true });
});

test("run timings grouped by repo", async ({ page }) => {
  const root = await openTimings(page);
  const chip = root.getByRole("group", { name: "Group by" }).getByRole("button", { name: "Repo" });
  await chip.click();
  await expect(chip).toHaveAttribute("aria-pressed", "true");
  await parkPointer(page);
  await expect(page).toHaveScreenshot("run-timings-repo.png", { fullPage: true });
});

test("run timings timeline bar hover", async ({ page }) => {
  const root = await openTimings(page);
  await root.locator("rect.bar").first().hover();
  await expect(page).toHaveScreenshot("run-timings-bar-hover.png", { fullPage: true });
});

test("run timings data quality expanded", async ({ page }) => {
  const root = await openTimings(page);
  const summary = root.getByRole("region", { name: "Data quality" }).locator("summary", { hasText: "dispatch rows with no tokens — reason stated" });
  await summary.click();
  await expect(summary.locator("xpath=..")).toHaveAttribute("open", "");
  await parkPointer(page);
  await expect(page).toHaveScreenshot("run-timings-dq-open.png", { fullPage: true });
});

test("run costs at rest", async ({ page }) => {
  await openCosts(page);
  await expect(page).toHaveScreenshot("run-costs.png", { fullPage: true });
});

test("run costs ninety days for one project", async ({ page }) => {
  const root = await openCosts(page);
  const period = root.getByRole("group", { name: "Period" }).getByRole("button", { name: "90 days" });
  await period.click();
  await expect(period).toHaveAttribute("aria-pressed", "true");
  const project = root.getByRole("group", { name: "Project" }).getByRole("button", { name: "uitest-beta" });
  await project.click();
  await expect(project).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByText("Loading…")).toHaveCount(0);
  await parkPointer(page);
  await expect(page).toHaveScreenshot("run-costs-filtered.png", { fullPage: true });
});

test("run costs compare picked", async ({ page }) => {
  const root = await openCosts(page);
  const b = root.getByRole("region", { name: "Compare runs" }).locator("label", { hasText: /^B/ }).getByRole("combobox");
  await b.selectOption(`uitest-beta/${OTHER}`);
  await expect(b).toHaveValue(`uitest-beta/${OTHER}`);
  await parkPointer(page);
  await expect(page).toHaveScreenshot("run-costs-compare.png", { fullPage: true });
});

test("run detail header with the timings button", async ({ page }) => {
  await open(page, `/#/run/uitest-beta/${RUN}`);
  await expect(page.getByRole("heading", { name: RUN, exact: true })).toBeVisible();
  const button = page.getByRole("button", { name: "Timings and costs →" });
  await expect(button).toBeVisible();
  await expect(page).toHaveScreenshot("run-detail-timings.png", { fullPage: true });
  const head = page.locator(".lb-runhead");
  await expect(head).toHaveScreenshot("run-detail-header.png");
  await button.hover();
  await expect(head).toHaveScreenshot("run-detail-header-hover.png");
  await parkPointer(page);
});
