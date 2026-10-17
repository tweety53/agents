// One test per view asserting it renders the actual numbers a fixture
// response carries (never merely that the component mounted -- the
// highest-risk vacuous shape this task's own instructions name), plus the
// absence-vs-zero distinction and the state-board-is-default-route
// requirement.
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { DecisionRow, HealthViewSlug, StatsResponse, StatsViewSlug, ViewName } from "../api";
import { App } from "../App";
import { Decisions } from "./Decisions";
import { FlowHealth } from "./FlowHealth";
import { Reviewers } from "./Reviewers";
import { SelfReview } from "./SelfReview";
import { StageLeaderboard } from "./StageLeaderboard";
import { StateBoard } from "./StateBoard";
import { Trend } from "./Trend";

const { fetchStatsViewMock } = vi.hoisted(() => ({ fetchStatsViewMock: vi.fn() }));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return { ...actual, fetchStatsView: fetchStatsViewMock };
});

const period = { from: new Date("2026-01-01T00:00:00Z"), to: new Date("2026-02-01T00:00:00Z") };

function envelope<Row>(view: StatsViewSlug, rows: Row, recorded = true, unmeasured = false): StatsResponse<Row> {
  return {
    view,
    from: period.from.toISOString(),
    to: period.to.toISOString(),
    boundaryConvention: "a stage run is attributed to the period containing its start instant",
    recorded,
    unmeasured,
    rows,
  };
}

// Every slug a view fetches: the navigable views, less "flow-health", which
// fetches the three health slugs instead of one of its own.
const fixtures: Record<Exclude<ViewName, "flow-health"> | HealthViewSlug, StatsResponse<unknown>> = {
  "state-board": envelope("state-board", [
    {
      projectKey: "kan-16-stats-app",
      name: "kan-16-stats-app",
      state: "IN_PROGRESS",
      updatedAt: "2026-01-15T10:00:00Z",
      updatedBy: "alice",
      nextCommand: "/flow-fast",
      plannedTasks: 22,
      currentTasks: 46,
      runCount: 7,
      costUsd: 12.5,
      wallClockMs: 5400000,
      errors: 3,
    },
    {
      projectKey: "kan-16-stats-app",
      name: "kan-29-never-observed",
      state: "STARTED",
      updatedAt: "2026-01-15T09:00:00Z",
      updatedBy: "bob",
      nextCommand: "/flow",
      runCount: 0,
      costUsd: 0,
      wallClockMs: 0,
      errors: 0,
    },
    {
      projectKey: "kan-16-stats-app",
      name: "kan-31-replan-fold",
      state: "FINISHED",
      updatedAt: "2026-01-15T08:00:00Z",
      updatedBy: "carol",
      nextCommand: "/flow",
      plannedTasks: 46,
      currentTasks: 22,
      runCount: 2,
      costUsd: 1.25,
      wallClockMs: 45000,
      errors: 11,
    },
    {
      projectKey: "kan-16-stats-app",
      name: "pipeline-fix-x-14",
      state: "FINISHED",
      updatedAt: "2026-01-15T07:00:00Z",
      updatedBy: "dave",
      nextCommand: "",
      runCount: 1,
      costUsd: 0.5,
      wallClockMs: 1000,
      errors: 0,
    },
  ]),
  "stage-leaderboard": envelope("stage-leaderboard", [
    {
      command: "/flow",
      stage: "SDD + TDD per task",
      runCount: 3,
      meanCostUsd: 1.5,
      medianCostUsd: 1.2,
      p90CostUsd: 2.9,
      totalCostUsd: 4.5,
    },
  ]),
  trend: envelope("trend", [{ day: "2026-01-05", runCount: 2, totalCostUsd: 5.25 }]),
  reviewers: envelope("reviewers", [
    {
      slot: "primary",
      experimental: false,
      description: "",
      dispatches: 2,
      changes: 1,
      critical: 1,
      important: 1,
      minor: 1,
    },
    {
      slot: "exp-failure-modes",
      experimental: true,
      description: "What the diff does under error, timeout and partial write",
      dispatches: 1,
      changes: 1,
      critical: 1,
      important: 0,
      minor: 1,
    },
  ]),
  decisions: envelope("decisions", [
    {
      project: "kan",
      change: "kan-1",
      recordedAt: "2026-06-10T09:10:00Z",
      class: "regular",
      overridden: true,
      execution: "inline",
      implementerModel: "skipped — inline",
      implementerEffort: "",
      fixer: "skipped — inline",
      rerunDispatch: "",
      rosterSize: 2,
      compact: true,
      experimentalSlot: "exp-failure-modes",
      rerun: "delta",
      grouping: "free",
      dispatches: "primary+exp-failure-modes",
      implementerGroups: "",
      wallClockSeconds: 600,
      inputTokens: 100,
      outputTokens: 50,
      cacheReadTokens: 20,
      costUsd: 1.25,
      critical: 1,
      important: 1,
      minor: 1,
      fixRounds: 2,
      fallbacks: 1,
      timedOut: 1,
    },
    {
      project: "kan",
      change: "kan-2",
      recordedAt: "2026-06-11T09:10:00Z",
      class: "regular",
      overridden: false,
      execution: "sdd",
      implementerModel: "opus",
      implementerEffort: "high",
      fixer: "sonnet/medium",
      rerunDispatch: "haiku/low",
      rosterSize: 5,
      compact: false,
      experimentalSlot: "",
      rerun: "full",
      grouping: "static",
      dispatches: "primary+principles · bugbot+mutation",
      implementerGroups: "1+2 · 3",
      wallClockSeconds: 1200,
      inputTokens: 200,
      outputTokens: 100,
      cacheReadTokens: 40,
      costUsd: 3.75,
      critical: 0,
      important: 2,
      minor: 1,
      fixRounds: 1,
      fallbacks: 0,
      timedOut: 0,
    },
  ]),
  "flow-summary": envelope("flow-summary", {
    changesStarted: 4,
    changesFinished: 3,
    totalCostUsd: 42.5,
    meanCostPerFinished: 9.25,
    medianLeadTimeMs: 2 * 86400000 + 43200000,
    meanPanelRounds: 2.5,
    critical: 1,
    important: 2,
    minor: 4,
    guardFires: 10,
  }),
  "self-review": envelope("self-review", [
    { recordedAt: "2026-01-12T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-fix", note: "guard misses a case", disposition: "fixed", ref: "abc1234", blastRadius: 3, commitUrl: "https://github.com/o/r/commit/abc1234" },
    { recordedAt: "2026-01-11T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-cost", note: "cheaper reviewer", disposition: "fixed", ref: "def5678", blastRadius: null, commitUrl: "https://github.com/o/r/commit/def5678" },
    { recordedAt: "2026-01-10T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-speed", note: "needs a redesign", disposition: "filed", ref: "KAN-9", blastRadius: null, commitUrl: "" },
    { recordedAt: "2026-01-09T10:00:00Z", project: "agents", change: "kan-2", angle: "flow-improvement", note: "not worth it", disposition: "declined", ref: "", blastRadius: null, commitUrl: "" },
    { recordedAt: "2026-01-08T10:00:00Z", project: "agents", change: "kan-2", angle: "flow-improvement", note: "panel caught it early", disposition: "noted", ref: "", blastRadius: null, commitUrl: "" },
  ]),
};

beforeEach(() => {
  fetchStatsViewMock.mockReset();
  fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures) => Promise.resolve(fixtures[view]));
});

describe("views render their fixture response's actual values", () => {
  it("state board shows the change's state and run totals", async () => {
    render(<StateBoard period={period} project={undefined} />);
    // The state also appears as a filter-dropdown <option>, which is not a
    // table cell -- scoping to role "cell" is what makes this assertion
    // about the rendered row rather than about either element.
    expect(await screen.findByRole("cell", { name: "IN_PROGRESS" })).toBeInTheDocument();
    expect(screen.queryByRole("columnheader", { name: /^By/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("columnheader", { name: /Next command/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("cell", { name: "alice" })).not.toBeInTheDocument();
    // Whole-life run totals (design.md's board-run-totals).
    const row = screen.getByRole("link", { name: "kan-16-stats-app" }).closest("tr")!;
    within(row).getByRole("cell", { name: "7" });
    within(row).getByRole("cell", { name: "$12.50" });
    within(row).getByRole("cell", { name: "90.0 min" });
    within(row).getByRole("cell", { name: "3" });
    // Plan growth (KAN-415): the observed change shows planned -> current
    // with the appended figure; the never-observed one reads as the
    // absence dash, never as a count of zero.
    expect(screen.getByRole("cell", { name: "22 → 46 (+24)" })).toBeInTheDocument();
    // A shrinking series carries its own sign, never a malformed "+-24"
    // (panel round 0, F1).
    expect(screen.getByRole("cell", { name: "46 → 22 (-24)" })).toBeInTheDocument();
    // (Scoped to that row's Plan growth cell: its Cost and Wall clock also
    // read "—" now, design.md's no-runs-dash.)
    const unobserved = screen.getByRole("link", { name: "kan-29-never-observed" }).closest("tr")!;
    expect(within(unobserved).getAllByRole("cell", { name: "—" })).toHaveLength(3);
    // Its own stat panel: a count of the rows shown, never a fabricated
    // total or mean for a categorical "state" column.
    expect(screen.getByRole("heading", { name: "Changes" })).toBeInTheDocument();
    within(screen.getByRole("region", { name: "Changes" })).getByText("3");
  });

  it("state board shows a dash for cost and wall clock of a change with no runs", async () => {
    render(<StateBoard period={period} project={undefined} />);
    // design.md's no-runs-dash: a change never measured reads as absent for
    // its cost and wall clock, while its run and error counts stay a measured 0.
    const row = (await screen.findByRole("link", { name: "kan-29-never-observed" })).closest("tr")!;
    const headers = screen.getAllByRole("columnheader").map((h) => h.textContent ?? "");
    const cells = within(row).getAllByRole("cell");
    const cellUnder = (header: RegExp) => cells[headers.findIndex((h) => header.test(h))];
    expect(cellUnder(/^Cost/)).toHaveTextContent(/^—$/);
    expect(cellUnder(/^Wall clock/)).toHaveTextContent(/^—$/);
    expect(cellUnder(/^Runs/)).toHaveTextContent(/^0$/);
    expect(cellUnder(/^Errors/)).toHaveTextContent(/^0$/);
  });

  it("state board hides pipeline-fix changes", async () => {
    render(<StateBoard period={period} project={undefined} />);
    expect(await screen.findByRole("link", { name: "kan-16-stats-app" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "pipeline-fix-x-14" })).not.toBeInTheDocument();
  });

  it("stage leaderboard shows total, mean, median and p90 cost, and its own stat panels", async () => {
    render(<StageLeaderboard period={period} project={undefined} />);
    expect(await screen.findByText("$1.50")).toBeInTheDocument();
    expect(screen.getByText("$1.20")).toBeInTheDocument();
    expect(screen.getByText("$2.90")).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: /Total cost/ })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "$4.50" })).toBeInTheDocument();
    within(screen.getByRole("region", { name: "Stages" })).getByText("1");
    within(screen.getByRole("region", { name: "Total runs" })).getByText("3");
  });

  it("trend shows the day's run count and total cost, its time-series panel, and its own stat panels", async () => {
    render(<Trend period={period} project={undefined} />);
    expect(await screen.findByText("2026-01-05")).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "$5.25" })).toBeInTheDocument();
    // Step 2: the daily points render through TimeSeriesPanel now, not the
    // view's own bespoke bar chart.
    expect(screen.getByTestId("time-series-point")).toBeInTheDocument();
    within(screen.getByRole("region", { name: "Days" })).getByText("1");
    within(screen.getByRole("region", { name: "Total cost" })).getByText("$5.25");
  });

  it("reviewers shows severity counts and badges an exp- slot with its description", async () => {
    render(<Reviewers period={period} project={undefined} />);
    expect(await screen.findByRole("cell", { name: /primary/ })).toBeInTheDocument();
    // primary's own severity counts, one dispatch/change/finding count per
    // column -- distinct values so a swap or a dropped field cannot hide.
    const primaryRow = screen.getByRole("cell", { name: /primary/ }).closest("tr")!;
    expect(within(primaryRow).getByText("2")).toBeInTheDocument(); // dispatches

    const expRow = screen.getByRole("cell", { name: /exp-failure-modes/ }).closest("tr")!;
    const badge = within(expRow).getByRole("img", { name: "experimental reviewer" });
    expect(badge).toHaveAttribute("title", "What the diff does under error, timeout and partial write");
    expect(screen.queryByRole("columnheader", { name: /Findings\/dispatch/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("columnheader", { name: /Withdrawn/ })).not.toBeInTheDocument();

    within(screen.getByRole("region", { name: "Slots" })).getByText("2");
    within(screen.getByRole("region", { name: "Total dispatches" })).getByText("3");
  });

  it("reviewers efficiency weighs critical 10, important 4, minor 1 per dispatch", async () => {
    const row = { experimental: false, description: "", changes: 1 };
    fetchStatsViewMock.mockResolvedValueOnce(
      envelope("reviewers", [
        { ...row, slot: "primary", dispatches: 166, critical: 4, important: 36, minor: 70 },
        { ...row, slot: "idle", dispatches: 0, critical: 0, important: 0, minor: 0 },
      ]),
    );
    render(<Reviewers period={period} project={undefined} />);
    expect(await screen.findByRole("columnheader", { name: /Efficiency/ })).toBeInTheDocument();
    const primaryRow = screen.getByRole("cell", { name: "primary" }).closest("tr")!;
    expect(within(primaryRow).getByRole("cell", { name: "1.53" })).toBeInTheDocument();
    const idleRow = screen.getByRole("cell", { name: "idle" }).closest("tr")!;
    expect(within(idleRow).getByRole("cell", { name: "—" })).toBeInTheDocument();
  });

  it("renders self-review fixes", async () => {
    const user = userEvent.setup();
    const { unmount } = render(<SelfReview period={period} project={undefined} />);

    // A fixed finding's ref links to its commit; a filed one's key is text.
    const link = await screen.findByRole("link", { name: "abc1234" });
    expect(link).toHaveAttribute("href", "https://github.com/o/r/commit/abc1234");
    expect(screen.getByRole("cell", { name: "KAN-9" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "KAN-9" })).not.toBeInTheDocument();

    within(screen.getByRole("region", { name: "Fixed" })).getByText("2");
    within(screen.getByRole("region", { name: "Filed" })).getByText("1");

    // design.md's self-review-drop-per-change: no per-change block.
    expect(screen.queryByRole("region", { name: "Per change" })).not.toBeInTheDocument();

    // The outcome column filters the findings table.
    await user.selectOptions(screen.getByRole("combobox", { name: "Filter by Outcome" }), "declined");
    expect(screen.getByRole("cell", { name: "not worth it" })).toBeInTheDocument();
    expect(screen.queryByRole("cell", { name: "guard misses a case" })).not.toBeInTheDocument();
    expect(screen.queryByRole("cell", { name: "panel caught it early" })).not.toBeInTheDocument();
    unmount();

    fetchStatsViewMock.mockResolvedValueOnce(envelope("self-review", []));
    render(<SelfReview period={period} project={undefined} />);
    expect(await screen.findByText("No self-review findings in this period.")).toBeInTheDocument();
  });

  // Also covers kan-472 task 23's grouping/dispatches/implementer-groups columns, asserted below.
  it("decisions shows a class × execution summary and no per-decision table", async () => {
    render(<Decisions period={period} project={undefined} />);

    // The class × execution summary groups the two runs into their own
    // rows -- "regular / inline" and "regular / sdd" -- each with its own
    // mean cost.
    const summaryRegion = await screen.findByRole("region", { name: "By class and execution" });
    expect(within(summaryRegion).getByText("regular / inline")).toBeInTheDocument();
    expect(within(summaryRegion).getByText("regular / sdd")).toBeInTheDocument();
    expect(within(summaryRegion).getByText("$1.25")).toBeInTheDocument();
    expect(within(summaryRegion).getByText("$3.75")).toBeInTheDocument();

    // design.md's decisions-by-implementer-model: the per-decision table
    // and the Runs tile are gone.
    expect(screen.queryByRole("region", { name: "Every decision" })).not.toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "Runs" })).not.toBeInTheDocument();
  });

  it("decisions groups runs by implementer model", async () => {
    const [inline, sdd] = fixtures.decisions.rows as DecisionRow[];
    fetchStatsViewMock.mockResolvedValueOnce(
      envelope("decisions", [inline, sdd, { ...sdd, change: "kan-3", costUsd: 1.25, wallClockSeconds: 600 }]),
    );
    render(<Decisions period={period} project={undefined} />);

    const region = await screen.findByRole("region", { name: "By implementer model" });
    const rows = within(region).getAllByRole("row").slice(1);
    expect(rows).toHaveLength(2);
    // opus first: two runs, mean cost (3.75 + 1.25) / 2, mean wall-clock
    // (1200 s + 600 s) / 2, mean findings 3, mean fix rounds 1.
    within(rows[0]).getByRole("cell", { name: "opus" });
    within(rows[0]).getByRole("cell", { name: "2" });
    within(rows[0]).getByRole("cell", { name: "$2.50" });
    within(rows[0]).getByRole("cell", { name: "15.0 min" });
    within(rows[0]).getByRole("cell", { name: "3.00" });
    within(rows[0]).getByRole("cell", { name: "1.00" });
    within(rows[1]).getByRole("cell", { name: "skipped — inline" });
    within(rows[1]).getByRole("cell", { name: "1" });
  });

});

describe("absence is rendered distinctly from a recorded zero", () => {
  it("a period before the store held anything is stated, not rendered as an empty measured table", async () => {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(envelope(view, [], false)),
    );
    render(<Trend period={period} project={undefined} />);
    // Every one of Trend's panels shares the same underlying `state`
    // (task 20's recomposition: one useStatsView call per view, several
    // Panel instances reading it), so each independently renders the
    // not-recorded banner rather than one page-level banner -- Panel is
    // now the sole implementation of that branch (this file's own header
    // comment, and Panel.tsx's), so seeing it repeated per panel is
    // exactly what "not duplicated across two files" looks like from one.
    const banners = await screen.findAllByTestId("not-recorded");
    expect(banners.length).toBeGreaterThan(0);
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.queryByTestId("time-series-point")).not.toBeInTheDocument();
  });

  // Task 5's third arm, exercised end to end through a real view rather
  // than Panel.test.tsx's isolated fixture: a period whose runs were
  // recorded but never attributed must read as its own state, distinct
  // from both "no data was recorded" (the case immediately above) and
  // from an ordinary measured table -- never collapsing into either.
  it("a period whose runs were recorded but none attributed reports that plainly, not as 'no data was recorded'", async () => {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(envelope(view, [], true, true)),
    );
    render(<Trend period={period} project={undefined} />);
    const banners = await screen.findAllByTestId("unmeasured");
    expect(banners.length).toBeGreaterThan(0);
    for (const banner of banners) {
      expect(banner).toHaveTextContent("Runs were recorded for this period, but none carried measurements.");
    }
    expect(screen.queryByTestId("not-recorded")).not.toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.queryByTestId("time-series-point")).not.toBeInTheDocument();
  });
});

describe("the state board is the only navigation path into a change's own dashboard", () => {
  it("links a row's change name to its run-detail route, round-tripping a URL-significant character in both segments", async () => {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(
        envelope(view, [
          {
            projectKey: "kan/16",
            name: "kan-16 stats-app?v2",
            state: "IN_PROGRESS",
            updatedAt: "2026-01-15T10:00:00Z",
            updatedBy: "alice",
            nextCommand: "/flow-fast",
          },
        ]),
      ),
    );

    render(<StateBoard period={period} project={undefined} />);

    const link = await screen.findByRole("link", { name: "kan-16 stats-app?v2" });
    expect(link).toHaveAttribute(
      "href",
      `#/run/${encodeURIComponent("kan/16")}/${encodeURIComponent("kan-16 stats-app?v2")}`,
    );
  });
});

// Task 2 (kan-183): the Project column names a project, it does not key
// it. The cell shows the display name, the full key survives as the
// cell's title (the identity the operator still needs when two checkouts
// share a basename), and the row's own link target keeps the full key
// regardless, since that is the route's identity and not this task's to
// shorten.
describe("the Project column names a project instead of keying it", () => {
  const PROJECT_KEY = "agents-a740d89c";

  it("state board's Project cell shows the display name with the full key as its title", async () => {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(
        envelope(view, [
          {
            projectKey: PROJECT_KEY,
            name: "kan-1",
            state: "IN_PROGRESS",
            updatedAt: "2026-01-15T10:00:00Z",
            updatedBy: "alice",
            nextCommand: "/flow-fast",
          },
        ]),
      ),
    );

    render(<StateBoard period={period} project={undefined} />);

    const cell = await screen.findByRole("cell", { name: "agents" });
    expect(cell).toHaveTextContent(/^agents$/);
    expect(cell.querySelector(`[title="${PROJECT_KEY}"]`)).not.toBeNull();

    // The row's link target is the route's identity -- it keeps the full
    // key even though the cell beside it now reads short.
    const link = screen.getByRole("link", { name: "kan-1" });
    expect(link).toHaveAttribute("href", `#/run/${encodeURIComponent(PROJECT_KEY)}/${encodeURIComponent("kan-1")}`);
  });

  // Reversed from task 2's original assertion (F3, panel round 1): the
  // dropdown now lists the full key, not the display name -- FilterBar
  // renders an option's value and its visible text from the same source
  // (DataTable's own `accessor`), so there is no way to keep the cell's
  // short label on the dropdown without also keying the filter itself on
  // it, which is exactly the ambiguity this fix closes. See the "two
  // projects whose keys share a basename" describe block below for why:
  // a control whose job is disambiguating two same-named projects cannot
  // itself compare on the name that fails to disambiguate them.
  it("the Project filter dropdown now lists the full key, not the display name", async () => {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(
        envelope(view, [
          {
            projectKey: PROJECT_KEY,
            name: "kan-1",
            state: "IN_PROGRESS",
            updatedAt: "2026-01-15T10:00:00Z",
            updatedBy: "alice",
            nextCommand: "/flow-fast",
          },
        ]),
      ),
    );

    render(<StateBoard period={period} project={undefined} />);

    await screen.findByRole("cell", { name: "agents" });
    const option = screen.getByRole("option", { name: PROJECT_KEY }) as HTMLOptionElement;
    expect(option.value).toBe(PROJECT_KEY);
    expect(screen.queryByRole("option", { name: "agents" })).not.toBeInTheDocument();
  });
});

// F3 (panel round 1, Major): DataTable's exact-match filter and free-text
// search both key off a column's `accessor`. Task 2 set the Project
// column's accessor to the *display name*, so the filter dropdown offered
// one option per distinct display name -- and two projects whose keys
// differ only in the disambiguating hash suffix collapsed into that one
// option, silently merging what resolveProjectParam
// (stats/internal/api/stats.go) refuses server-side with a 400 naming the
// ambiguity. The fix returns the accessor to the full key: the dropdown
// now offers one option per key, each of which narrows to exactly the one
// project it names.
describe("the Project filter narrows by key identity, not by display name", () => {
  const BASENAME = "agents";
  const KEY_ONE = `${BASENAME}-a740d89c`;
  const KEY_TWO = `${BASENAME}-b851e9ad`;

  function twoProjectsSharingABasename(): void {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(
        envelope(view, [
          {
            projectKey: KEY_ONE,
            name: "kan-1",
            state: "IN_PROGRESS",
            updatedAt: "2026-01-15T10:00:00Z",
            updatedBy: "alice",
            nextCommand: "/flow-fast",
          },
          {
            projectKey: KEY_TWO,
            name: "kan-2",
            state: "STARTED",
            updatedAt: "2026-01-15T11:00:00Z",
            updatedBy: "bob",
            nextCommand: "/flow",
          },
        ]),
      ),
    );
  }

  it("offers one filter option per project key, not one per display name", async () => {
    twoProjectsSharingABasename();
    render(<StateBoard period={period} project={undefined} />);

    await screen.findByRole("link", { name: "kan-1" });
    const select = screen.getByRole("combobox", { name: "Filter by Project" });
    // Pre-fix, both rows' accessor collapsed to "agents": the dropdown's
    // deduplicated option list held exactly one entry, and selecting it
    // could never narrow the table to either project alone.
    expect(within(select).getAllByRole("option")).toHaveLength(3); // "All" plus one per key
  });

  it("filtering by one key's value selects only that project's row", async () => {
    twoProjectsSharingABasename();
    const user = userEvent.setup();
    render(<StateBoard period={period} project={undefined} />);

    await screen.findByRole("link", { name: "kan-1" });
    const select = screen.getByRole("combobox", { name: "Filter by Project" });

    await user.selectOptions(select, KEY_ONE);
    expect(screen.getByRole("link", { name: "kan-1" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "kan-2" })).not.toBeInTheDocument();

    await user.selectOptions(select, KEY_TWO);
    expect(screen.getByRole("link", { name: "kan-2" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "kan-1" })).not.toBeInTheDocument();
  });
});

describe("routing", () => {
  it("the live state board is the default route", async () => {
    window.location.hash = "";
    render(<App />);
    expect(await screen.findByRole("cell", { name: "IN_PROGRESS" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Live state board" })).toBeInTheDocument();
  });

  it("an unrecognised route also falls back to the state board", async () => {
    window.location.hash = "#/not-a-real-view";
    render(<App />);
    expect(await screen.findByRole("cell", { name: "IN_PROGRESS" })).toBeInTheDocument();
    window.location.hash = "";
  });
});

describe("flow health", () => {
  it("flow health shows eight totals tiles", async () => {
    render(<FlowHealth period={period} project={undefined} />);
    const tile = (name: string) => within(screen.getByRole("region", { name }));
    expect(await screen.findByRole("region", { name: "Changes started" })).toBeInTheDocument();
    tile("Changes started").getByText("4");
    tile("Changes finished").getByText("3");
    tile("Total cost").getByText("$42.50");
    tile("Mean cost per finished change").getByText("$9.25");
    tile("Median lead time").getByText("2.5 d");
    tile("Mean panel rounds").getByText("2.50");
    tile("Findings (C/I/Mi)").getByText("1/2/4");
    tile("Guard fires").getByText("10");
    expect(document.querySelectorAll(".stat-panel")).toHaveLength(8);
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });

  it("flow health reads an unfinished period's means as unavailable", async () => {
    fetchStatsViewMock.mockResolvedValueOnce(
      envelope("flow-summary", {
        changesStarted: 1, changesFinished: 0, totalCostUsd: 0, meanCostPerFinished: null,
        medianLeadTimeMs: null, meanPanelRounds: null, critical: 0, important: 0, minor: 0, guardFires: 0,
      }),
    );
    render(<FlowHealth period={period} project={undefined} />);
    const lead = await screen.findByRole("region", { name: "Median lead time" });
    expect(within(lead).getByTestId("unavailable")).toBeInTheDocument();
  });
});

