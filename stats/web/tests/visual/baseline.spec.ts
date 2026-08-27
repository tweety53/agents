import { expect, test } from "@playwright/test";

// The visual baseline KAN-171 exists to make automatic: KAN-16 shipped
// three defects that were invisible in a diff and obvious the moment the
// page was opened -- a 400 on the run-detail dashboard, a white page where
// a dark palette was asked for, and every stat panel printing its label
// twice -- and all three survived a five-pass review panel, 296 Go tests
// and 110 SPA tests. Each `test` below carries an explicit assertion for
// one of those three defect classes, so that reintroducing any of them
// fails here even if the pixel diff alone somehow did not (it does, for
// all three, but the assertions are what state that in words rather than
// relying on a reader to infer it from a snapshot).
//
// The dashboard's period is pinned via explicit `from`/`to` query params
// (App.tsx's own periodFromQuery) rather than left to the default 30-day
// window: the fixture's fixed timestamp (cmd/uitest-seed/seed.go's
// fixtureNow, 2026-08-15T09:00:00Z) would otherwise age out of "the last
// 30 days" a few weeks after this baseline is captured, and the dashboard
// would silently render empty for a reason that has nothing to do with
// this suite's own subject.
const DASHBOARD_URL = "/#/state-board?from=2026-08-01T00%3A00%3A00.000Z&to=2026-08-16T00%3A00%3A00.000Z";

// uitest-beta/kan-201-refactor-thing is the fixture's own richest change:
// FINISHED, two stage runs (/myflow-do then /myflow-finish) across two
// models, so the timeline and the table both render more than one row.
// Run detail's own data ignores the shared period control (useRunDetail.ts's
// own WIDE_FROM) -- but the dashboard bar above it still renders that
// control from `period`, which App.tsx seeds from `new Date()` when the URL
// carries no explicit window. Left unpinned, the native datetime-local
// inputs' displayed minute/second ticks over between the run that wrote
// this baseline and the run that compares against it, failing on a pixel
// diff that has nothing to do with this suite's own subject. The same
// explicit query params as the dashboard pin it, for the same reason.
const RUN_DETAIL_URL =
  "/#/run/uitest-beta/kan-201-refactor-thing?from=2026-08-01T00%3A00%3A00.000Z&to=2026-08-16T00%3A00%3A00.000Z";

// The dark surface-1 token (theme.css: --surface-1: #111217), as a browser
// computes it. Reading this rather than a screenshot pixel is what makes
// the assertion state its own failure reason ("the page is not dark")
// instead of just "pixels differ" -- toHaveScreenshot below still catches
// the same regression, but this is the one that says why.
const DARK_SURFACE_1 = "rgb(17, 18, 23)";

async function bodyBackground(page: import("@playwright/test").Page): Promise<string> {
  return page.evaluate(() => getComputedStyle(document.body).backgroundColor);
}

test.describe("dashboard", () => {
  test("renders the live state board with data loaded and the dark palette applied", async ({ page }) => {
    await page.goto(DASHBOARD_URL);
    await expect(page.getByRole("heading", { name: "Live state board" })).toBeVisible();

    // Guards the 400 defect: the table must actually hold the fixture's
    // rows, not fail silently and render an empty table.
    await expect(page.getByRole("cell", { name: "kan-101-add-widget" })).toBeVisible();
    await expect(page.getByRole("cell", { name: "kan-102-fix-widget-bug" })).toBeVisible();
    await expect(page.getByRole("cell", { name: "kan-201-refactor-thing" })).toBeVisible();

    // Guards the white-page defect: the body's rendered background must be
    // the dark token, not the browser's default white document a
    // prefers-color-scheme-gated dark theme would leave behind. This alone
    // does not cover the whole defect class -- see the "dark palette
    // forced via data-theme" describe block below for the combination this
    // assertion, run under playwright.config.ts's colorScheme: "dark", is
    // structurally unable to catch.
    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    // Guards the duplicate-label defect: the "Changes" stat panel's own
    // label is owned by the panel heading; StatPanel must not render it a
    // second time inside the panel body.
    const changesPanel = page.locator('section.panel[aria-label="Changes"]');
    await expect(changesPanel.getByText("Changes", { exact: true })).toHaveCount(1);

    await expect(page).toHaveScreenshot("dashboard.png", { fullPage: true });
  });
});

test.describe("run detail", () => {
  test("renders one change's stage-run history with data loaded and the dark palette applied", async ({ page }) => {
    await page.goto(RUN_DETAIL_URL);
    await expect(page.getByRole("heading", { name: "kan-201-refactor-thing" })).toBeVisible();

    // Guards the 400 defect directly: task 26's own regression (sorting by
    // "startedAt", a DTO field name, instead of "started_at", the server's
    // allowlisted column) leaves this route on "Loading…" forever with an
    // error banner never shown either -- the request just never resolves.
    // Both the "still loading" and "errored" cases are excluded, and the
    // two stage runs the fixture recorded must actually be visible.
    await expect(page.getByRole("status")).toHaveCount(0);
    await expect(page.getByRole("alert")).toHaveCount(0);
    // Scoped to the stage-run table specifically: "/myflow-do" alone is
    // ambiguous -- the same fixture value also appears in the timeline's
    // own SVG label and in the table's command filter dropdown.
    const stageRunsPanel = page.locator('section.panel[aria-label="Stage runs"]');
    await expect(stageRunsPanel.getByRole("cell", { name: "/myflow-do", exact: true })).toBeVisible();
    await expect(stageRunsPanel.getByRole("cell", { name: "/myflow-finish", exact: true })).toBeVisible();

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);

    const runsPanel = page.locator('section.panel[aria-label="Runs"]');
    await expect(runsPanel.getByText("Runs", { exact: true })).toHaveCount(1);

    await expect(page).toHaveScreenshot("run-detail.png", { fullPage: true });
  });
});

// theme.css's own contract has three cases, not two: bare `:root` carries
// dark unconditionally; `@media (prefers-color-scheme: light)` serves a
// real light variant, excluded via `:not([data-theme="dark"])`; and
// `:root[data-theme="light"]` forces light explicitly regardless of the
// OS. The two describe blocks above run under playwright.config.ts's
// colorScheme: "dark" and so only ever exercise the bare-`:root` path --
// a mutation that moves every color token out of bare `:root` into
// `@media (prefers-color-scheme: dark) { :root { … } }` still renders
// correctly under both of those (the dark OS context still matches its
// own media query), so neither catches it. `data-theme="dark"` under a
// *light* OS context is the one combination that mutation actually
// breaks: today, the light media block is excluded by its own
// `:not([data-theme="dark"])`, and bare `:root`'s unconditional dark
// values still apply -- correct. Under the mutation, the light block
// still doesn't match (data-theme is "dark", not absent or "light"), and
// the dark media block doesn't match either (the OS context here is
// light) -- no color tokens apply at all, and the body's
// `background: var(--surface-1)` resolves to nothing rather than a
// specific color, not to the light variant's white either. A computed
// background color, not a screenshot, is what makes that failure
// specific: a screenshot alone would just become the new accepted
// baseline the next time someone runs `--update-snapshots` without
// looking.
test.describe("dark palette forced via data-theme, in a light OS context", () => {
  test.use({ colorScheme: "light" });

  test('data-theme="dark" still renders the dark surface token', async ({ page }) => {
    await page.goto(DASHBOARD_URL);
    await expect(page.getByRole("heading", { name: "Live state board" })).toBeVisible();

    // Set after load, not via an init script run before navigation: CSS
    // custom properties and media/attribute selectors are recomputed
    // live off the DOM's current state, so the timing of the attribute
    // write does not change what steady-state getComputedStyle reports
    // below -- only whether an earlier flash-of-wrong-theme happened,
    // which is not this assertion's subject. No part of this app sets
    // data-theme itself today; this exercises theme.css's own contract
    // directly rather than through a UI control that doesn't exist yet.
    await page.evaluate(() => document.documentElement.setAttribute("data-theme", "dark"));

    expect(await bodyBackground(page)).toBe(DARK_SURFACE_1);
  });
});
