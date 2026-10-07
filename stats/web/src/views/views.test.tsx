// One test per view asserting it renders the actual numbers a fixture
// response carries (never merely that the component mounted -- the
// highest-risk vacuous shape this task's own instructions name), plus the
// absence-vs-zero distinction and the state-board-is-default-route
// requirement.
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { HealthViewSlug, StatsResponse, StatsViewSlug, ViewName } from "../api";
import { App } from "../App";
import { CacheEfficiency } from "./CacheEfficiency";
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
    },
    {
      projectKey: "kan-16-stats-app",
      name: "kan-29-never-observed",
      state: "STARTED",
      updatedAt: "2026-01-15T09:00:00Z",
      updatedBy: "bob",
      nextCommand: "/flow",
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
    },
  ]),
  trend: envelope("trend", [{ day: "2026-01-05", runCount: 2, totalCostUsd: 5.25 }]),
  "cache-efficiency": envelope("cache-efficiency", [
    { command: "/flow", stage: "measured-zero", cacheReadTotal: 0, cacheCreationTotal: 1000, ratio: 0 },
    { command: "/flow", stage: "never-measured", cacheReadTotal: null, cacheCreationTotal: null, ratio: null },
  ]),
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
      findingsPerDispatch: 1.5,
      withdrawnShare: 0,
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
      findingsPerDispatch: 2.0,
      withdrawnShare: 0.5,
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
  runs: envelope("runs", []),
  guards: envelope("guards", [
    {
      guard: "check-plan-shape",
      runs: 40,
      fired: 10,
      cannotAnswer: 3,
      lastRunAt: "2026-01-20T10:00:00Z",
      lastFiredAt: "2026-01-19T10:00:00Z",
      medianDurationMs: 850,
      verdicts: 0,
      falsePositives: 0,
    },
    {
      guard: "check-quiet",
      runs: 12,
      fired: 0,
      cannotAnswer: 0,
      lastRunAt: "2026-01-20T10:00:00Z",
      lastFiredAt: null,
      medianDurationMs: 40,
      verdicts: 0,
      falsePositives: 0,
    },
    {
      guard: "check-unfinished-work",
      runs: 0,
      fired: 0,
      cannotAnswer: 0,
      lastRunAt: null,
      lastFiredAt: null,
      medianDurationMs: null,
      verdicts: 6,
      falsePositives: 2,
    },
  ]),
  "stage-redo": envelope("stage-redo", [
    {
      command: "/flow",
      stage: "flow.review-panel",
      runs: 9,
      changes: 5,
      reentries: 4,
      reenteredChanges: 3,
      medianSeconds: 90,
      p90Seconds: null,
    },
  ]),
  "self-review": envelope("self-review", [
    { recordedAt: "2026-01-12T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-fix", note: "guard misses a case", disposition: "fixed", ref: "abc1234", blastRadius: 3, commitUrl: "https://github.com/o/r/commit/abc1234" },
    { recordedAt: "2026-01-11T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-cost", note: "cheaper reviewer", disposition: "fixed", ref: "def5678", blastRadius: null, commitUrl: "https://github.com/o/r/commit/def5678" },
    { recordedAt: "2026-01-10T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-speed", note: "needs a redesign", disposition: "filed", ref: "KAN-9", blastRadius: null, commitUrl: "" },
    { recordedAt: "2026-01-09T10:00:00Z", project: "agents", change: "kan-2", angle: "flow-improvement", note: "not worth it", disposition: "declined", ref: "", blastRadius: null, commitUrl: "" },
  ]),
  "panel-rounds": envelope("panel-rounds", [
    { project: "agents", change: "kan-1", startedAt: "2026-01-10T10:00:00Z", rounds: 3, findings: 7, critical: 1, important: 2, minor: 4 },
    { project: "agents", change: "kan-2", startedAt: "2026-01-09T10:00:00Z", rounds: 2, findings: 0, critical: 0, important: 0, minor: 0 },
    { project: "agents", change: "kan-3", startedAt: "2026-01-08T10:00:00Z", rounds: null, findings: 0, critical: 0, important: 0, minor: 0 },
  ]),
};

beforeEach(() => {
  fetchStatsViewMock.mockReset();
  fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures) => Promise.resolve(fixtures[view]));
});

describe("views render their fixture response's actual values", () => {
  it("state board shows the change's state, updater and next command", async () => {
    render(<StateBoard period={period} project={undefined} />);
    // The state also appears as a filter-dropdown <option>, which is not a
    // table cell -- scoping to role "cell" is what makes this assertion
    // about the rendered row rather than about either element.
    expect(await screen.findByRole("cell", { name: "IN_PROGRESS" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "alice" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "/flow-fast" })).toBeInTheDocument();
    // Plan growth (KAN-415): the observed change shows planned -> current
    // with the appended figure; the never-observed one reads as the
    // absence dash, never as a count of zero.
    expect(screen.getByRole("cell", { name: "22 → 46 (+24)" })).toBeInTheDocument();
    // A shrinking series carries its own sign, never a malformed "+-24"
    // (panel round 0, F1).
    expect(screen.getByRole("cell", { name: "46 → 22 (-24)" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "—" })).toBeInTheDocument();
    // Its own stat panel: a count of the rows the server returned, never a
    // fabricated total or mean for a categorical "state" column.
    expect(screen.getByRole("heading", { name: "Changes" })).toBeInTheDocument();
    within(screen.getByRole("region", { name: "Changes" })).getByText("3");
  });

  it("stage leaderboard shows mean, median and p90 cost, and its own stat panels", async () => {
    render(<StageLeaderboard period={period} project={undefined} />);
    expect(await screen.findByText("$1.50")).toBeInTheDocument();
    expect(screen.getByText("$1.20")).toBeInTheDocument();
    expect(screen.getByText("$2.90")).toBeInTheDocument();
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
    expect(within(primaryRow).getByText("1.50")).toBeInTheDocument(); // findings/dispatch

    const expRow = screen.getByRole("cell", { name: /exp-failure-modes/ }).closest("tr")!;
    const badge = within(expRow).getByRole("img", { name: "experimental reviewer" });
    expect(badge).toHaveAttribute("title", "What the diff does under error, timeout and partial write");
    expect(within(expRow).getByText("50%")).toBeInTheDocument(); // withdrawn share

    within(screen.getByRole("region", { name: "Slots" })).getByText("2");
    within(screen.getByRole("region", { name: "Total dispatches" })).getByText("3");
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

    // Per change: fixed / filed / declined, derived from the rows.
    const perChange = screen.getByRole("region", { name: "Per change" });
    const kan1 = within(perChange).getByRole("cell", { name: "kan-1" }).closest("tr")!;
    expect(within(kan1).getAllByRole("cell").map((c) => c.textContent)).toEqual(["kan-1", "2", "1", "0"]);
    const kan2 = within(perChange).getByRole("cell", { name: "kan-2" }).closest("tr")!;
    expect(within(kan2).getAllByRole("cell").map((c) => c.textContent)).toEqual(["kan-2", "0", "0", "1"]);

    // The outcome column filters the findings table.
    await user.selectOptions(screen.getByRole("combobox", { name: "Filter by Outcome" }), "declined");
    expect(screen.getByRole("cell", { name: "not worth it" })).toBeInTheDocument();
    expect(screen.queryByRole("cell", { name: "guard misses a case" })).not.toBeInTheDocument();
    unmount();

    fetchStatsViewMock.mockResolvedValueOnce(envelope("self-review", []));
    render(<SelfReview period={period} project={undefined} />);
    expect(await screen.findByText("No self-review findings in this period.")).toBeInTheDocument();
  });

  // Also covers kan-472 task 23's grouping/dispatches/implementer-groups columns, asserted below.
  it("decisions shows a run's class, execution and cost, and a class × execution summary", async () => {
    render(<Decisions period={period} project={undefined} />);

    const inlineRow = (await screen.findByRole("cell", { name: "kan-1" })).closest("tr")!;
    expect(within(inlineRow).getByText("regular ↑")).toBeInTheDocument();
    expect(within(inlineRow).getByText("inline")).toBeInTheDocument();
    expect(within(inlineRow).getByText("$1.25")).toBeInTheDocument();
    expect(within(inlineRow).getByText("free")).toBeInTheDocument();
    expect(within(inlineRow).getByText("primary+exp-failure-modes")).toBeInTheDocument();

    const sddRow = screen.getByRole("cell", { name: "kan-2" }).closest("tr")!;
    expect(within(sddRow).getByText("sdd")).toBeInTheDocument();
    expect(within(sddRow).getByText("$3.75")).toBeInTheDocument();
    expect(within(sddRow).getByText("static")).toBeInTheDocument();
    expect(within(sddRow).getByText("primary+principles · bugbot+mutation")).toBeInTheDocument();
    expect(within(sddRow).getByText("1+2 · 3")).toBeInTheDocument();
    expect(within(sddRow).getByText("sonnet/medium")).toBeInTheDocument();
    expect(within(sddRow).getByText("haiku/low")).toBeInTheDocument();

    // The class × execution summary groups the two runs into their own
    // rows -- "regular / inline" and "regular / sdd" -- each with its own
    // mean cost, distinct from the per-run table above.
    const summaryRegion = screen.getByRole("region", { name: "By class and execution" });
    expect(within(summaryRegion).getByText("regular / inline")).toBeInTheDocument();
    expect(within(summaryRegion).getByText("regular / sdd")).toBeInTheDocument();

    within(screen.getByRole("region", { name: "Runs" })).getByText("2");
  });

});

describe("absence is rendered distinctly from a recorded zero", () => {
  it("cache efficiency: a real zero ratio reads as a value, an unmeasured ratio reads as unavailable", async () => {
    render(<CacheEfficiency period={period} project={undefined} />);

    const zeroRow = (await screen.findByText("measured-zero")).closest("tr") as HTMLElement;
    const missingRow = screen.getByText("never-measured").closest("tr") as HTMLElement;

    // Columns: command, stage, cache read, cache creation, ratio -- the
    // ratio cell is the last one in each row.
    const zeroCells = within(zeroRow).getAllByRole("cell");
    const missingCells = within(missingRow).getAllByRole("cell");
    const zeroRatioCell = zeroCells[zeroCells.length - 1];
    const missingRatioCell = missingCells[missingCells.length - 1];

    expect(within(zeroRatioCell).getByTestId("measured")).toHaveTextContent("0.00");
    expect(within(zeroRatioCell).queryByTestId("unavailable")).not.toBeInTheDocument();

    expect(within(missingRatioCell).getByTestId("unavailable")).toBeInTheDocument();
    expect(within(missingRatioCell).queryByTestId("measured")).not.toBeInTheDocument();
    // Every cell in the never-measured row reads as unavailable, not as a
    // fabricated zero -- cacheReadTotal and cacheCreationTotal are also
    // null in this fixture.
    expect(within(missingRow).getAllByTestId("unavailable")).toHaveLength(3);

    // Cache efficiency's own stat panels: totals pooled across both rows
    // (1000 read via the zero-ratio row's own cacheReadTotal of 0, plus
    // null from the never-measured row -- sumNullable treats the null row
    // as contributing nothing, not as a zero), and the overall ratio
    // 0 / 1000 = 0, a real measured zero, not "unavailable".
    within(screen.getByRole("region", { name: "Total cache read" })).getByText("0");
    within(screen.getByRole("region", { name: "Total cache creation" })).getByText("1,000");
    within(screen.getByRole("region", { name: "Overall ratio" })).getByText("0.00");
  });

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
  it("shows each guard's fire rate, marks a never-fired guard, and keeps a verdict-only guard's run fields unavailable", async () => {
    render(<FlowHealth period={period} project={undefined} />);
    const shape = (await screen.findByRole("cell", { name: "check-plan-shape" })).closest("tr")!;
    expect(within(shape).getByText("25%")).toBeInTheDocument(); // 10 of 40 fired
    expect(within(shape).getByText("850 ms")).toBeInTheDocument();

    const quiet = screen.getByRole("cell", { name: "check-quiet" }).closest("tr")!;
    expect(within(quiet).getByText("0%")).toBeInTheDocument();
    expect(within(quiet).getByText("never")).toBeInTheDocument();

    // Known only from its verdicts: no fire rate and no runtime, shown as
    // unavailable rather than as a zero.
    const verdictOnly = screen.getByRole("cell", { name: "check-unfinished-work" }).closest("tr")!;
    expect(within(verdictOnly).getAllByTestId("unavailable")).toHaveLength(2);
    expect(within(verdictOnly).getByText("6")).toBeInTheDocument();

    within(screen.getByRole("region", { name: "Guards that fired" })).getByText("1");
    within(screen.getByRole("region", { name: "Guards that ran and never fired" })).getByText("1");
  });

  it("shows stage re-entries and the mean over changes that recorded panel rounds", async () => {
    render(<FlowHealth period={period} project={undefined} />);
    const panelRow = (await screen.findByRole("cell", { name: "flow.review-panel" })).closest("tr")!;
    expect(within(panelRow).getByText("60%")).toBeInTheDocument(); // 3 of 5 changes re-entered
    expect(within(panelRow).getByText("1.5 min")).toBeInTheDocument();
    expect(within(panelRow).getByTestId("unavailable")).toBeInTheDocument(); // p90 not measured

    // (3 + 2) / 2 -- kan-3 recorded no rounds and is not counted as zero.
    within(screen.getByRole("region", { name: "Mean panel rounds" })).getByText("2.50");
    within(screen.getByRole("region", { name: "Total re-entries" })).getByText("4");
    const kan3 = screen.getByRole("cell", { name: "kan-3" }).closest("tr")!;
    expect(within(kan3).getByTestId("unavailable")).toBeInTheDocument();
  });
});
