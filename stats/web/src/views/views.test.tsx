// One test per view asserting it renders the actual numbers a fixture
// response carries (never merely that the component mounted -- the
// highest-risk vacuous shape this task's own instructions name), plus the
// absence-vs-zero distinction and the state-board-is-default-route
// requirement.
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { StatsResponse, StatsViewParams, ViewName } from "../api";
import { App } from "../App";
import { CacheEfficiency } from "./CacheEfficiency";
import { CostPerChange } from "./CostPerChange";
import { ModelComparison } from "./ModelComparison";
import { PanelEconomics } from "./PanelEconomics";
import { ReworkRate } from "./ReworkRate";
import { StageLeaderboard } from "./StageLeaderboard";
import { StateBoard } from "./StateBoard";
import { Trend } from "./Trend";

const { fetchStatsViewMock } = vi.hoisted(() => ({ fetchStatsViewMock: vi.fn() }));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return { ...actual, fetchStatsView: fetchStatsViewMock };
});

const period = { from: new Date("2026-01-01T00:00:00Z"), to: new Date("2026-02-01T00:00:00Z") };

function envelope<Row>(view: ViewName, rows: Row, recorded = true): StatsResponse<Row> {
  return {
    view,
    from: period.from.toISOString(),
    to: period.to.toISOString(),
    boundaryConvention: "a stage run is attributed to the period containing its start instant",
    recorded,
    rows,
  };
}

const fixtures: Record<ViewName, StatsResponse<unknown>> = {
  "state-board": envelope("state-board", [
    {
      projectKey: "kan-16-myflow-stats-app",
      name: "kan-16-myflow-stats-app",
      state: "IN_PROGRESS",
      updatedAt: "2026-01-15T10:00:00Z",
      updatedBy: "alice",
      nextCommand: "/myflow-finish",
    },
  ]),
  "cost-per-change": envelope("cost-per-change", [
    {
      projectKey: "kan-16-myflow-stats-app",
      changeName: "kan-16-myflow-stats-app",
      command: "/myflow-do",
      stage: "SDD + TDD per task",
      runCount: 4,
      measuredRuns: 4,
      totalTokensInput: 12000,
      meanTokensInput: 3000,
      totalCostUsd: 4.5678,
      totalDurationMs: 90000,
      mainTokens: 11000,
      sidechainTokens: 1000,
    },
  ]),
  "stage-leaderboard": envelope("stage-leaderboard", [
    {
      command: "/myflow-do",
      stage: "SDD + TDD per task",
      runCount: 3,
      meanCostUsd: 1.5,
      medianCostUsd: 1.2,
      p90CostUsd: 2.9,
    },
  ]),
  trend: envelope("trend", [{ day: "2026-01-05", runCount: 2, totalCostUsd: 5.25 }]),
  "cache-efficiency": envelope("cache-efficiency", [
    { command: "/myflow-do", stage: "measured-zero", cacheReadTotal: 0, cacheCreationTotal: 1000, ratio: 0 },
    { command: "/myflow-do", stage: "never-measured", cacheReadTotal: null, cacheCreationTotal: null, ratio: null },
  ]),
  "panel-economics": envelope("panel-economics", [
    { reviewPanelRoster: "light", findingsTotal: 6, tokensTotal: 60000, findingsPerMtok: 100 },
  ]),
  "model-comparison": envelope("model-comparison", [
    { model: "claude-sonnet-5", command: "/myflow-do", stage: "SDD + TDD per task", runCount: 4, meanCostUsd: 0.75, reworkAttempts: 1 },
  ]),
  "rework-rate": envelope("rework-rate", [
    { command: "/myflow-do", stage: "SDD + TDD per task", totalAttempts: 5, reworkAttempts: 2, abandonedCount: 1 },
  ]),
};

beforeEach(() => {
  fetchStatsViewMock.mockReset();
  fetchStatsViewMock.mockImplementation((view: ViewName) => Promise.resolve(fixtures[view]));
});

describe("views render their fixture response's actual values", () => {
  it("state board shows the change's state, updater and next command", async () => {
    render(<StateBoard period={period} project={undefined} />);
    // The state also appears as a filter-dropdown <option>, which is not a
    // table cell -- scoping to role "cell" is what makes this assertion
    // about the rendered row rather than about either element.
    expect(await screen.findByRole("cell", { name: "IN_PROGRESS" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "alice" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "/myflow-finish" })).toBeInTheDocument();
    // Its own stat panel: a count of the rows the server returned, never a
    // fabricated total or mean for a categorical "state" column.
    expect(screen.getByRole("heading", { name: "Changes" })).toBeInTheDocument();
    within(screen.getByRole("region", { name: "Changes" })).getByText("1");
  });

  it("cost per change shows the run's cost and duration, and its own stat panels summed from the same rows", async () => {
    render(<CostPerChange period={period} project={undefined} />);
    expect(await screen.findByRole("cell", { name: "$4.5678" })).toBeInTheDocument();
    expect(screen.getByText("1.5 min")).toBeInTheDocument();
    // Stat panels: Runs (sum of runCount = 4), Total cost ($4.5678, the
    // one row's own total) and Mean cost per run ($4.5678 / 4).
    within(screen.getByRole("region", { name: "Runs" })).getByText("4");
    within(screen.getByRole("region", { name: "Total cost" })).getByText("$4.5678");
    within(screen.getByRole("region", { name: "Mean cost per run" })).getByText("$1.142");
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

  it("panel economics shows findings and findings-per-Mtok, and its own stat panels", async () => {
    render(<PanelEconomics period={period} project={undefined} />);
    expect(await screen.findByText("light")).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "6" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "100.00" })).toBeInTheDocument();
    within(screen.getByRole("region", { name: "Total findings" })).getByText("6");
    within(screen.getByRole("region", { name: "Total tokens" })).getByText("60,000");
  });

  it("model comparison shows the model, its mean cost and rework attempts, and its own stat panels", async () => {
    render(<ModelComparison period={period} project={undefined} />);
    // "claude-sonnet-5" also names an <option> in the model filter dropdown
    // (the column is filterable); role "cell" scopes this to the row.
    expect(await screen.findByRole("cell", { name: "claude-sonnet-5" })).toBeInTheDocument();
    expect(screen.getByText("$0.75")).toBeInTheDocument();
    within(screen.getByRole("region", { name: "Models compared" })).getByText("1");
    within(screen.getByRole("region", { name: "Total runs" })).getByText("4");
  });

  it("rework rate shows total attempts, rework attempts and abandoned count, and its own stat panels", async () => {
    render(<ReworkRate period={period} project={undefined} />);
    const row = (await screen.findByText("SDD + TDD per task")).closest("tr");
    expect(row).not.toBeNull();
    const cells = within(row as HTMLElement)
      .getAllByRole("cell")
      .map((c) => c.textContent);
    expect(cells).toEqual(["/myflow-do", "SDD + TDD per task", "5", "2", "1"]);
    within(screen.getByRole("region", { name: "Total attempts" })).getByText("5");
    // 2 rework attempts over 5 total attempts.
    within(screen.getByRole("region", { name: "Rework rate" })).getByText("40.0%");
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
});

describe("cost-per-change repository breakdown is scoped to its own row", () => {
  // Post-commit review finding F1: costPerChangeByRepo used to filter only
  // on project + change name, so it summed *every* stage's runs into one
  // panel regardless of which row's toggle asked for it. Two rows of the
  // same change, with different (command, stage) and different totals,
  // each expanded, must each reconcile with its own row -- not with each
  // other and not with the whole change's sum.
  const sddRow = {
    projectKey: "kan-16-myflow-stats-app",
    changeName: "kan-16-myflow-stats-app",
    command: "/myflow-do",
    stage: "SDD + TDD per task",
    runCount: 2,
    measuredRuns: 2,
    totalTokensInput: 1000,
    meanTokensInput: 500,
    totalCostUsd: 2.1,
    totalDurationMs: 60000,
    mainTokens: 900,
    sidechainTokens: 100,
  };
  const panelRow = {
    ...sddRow,
    command: "/myflow-do",
    stage: "5. The review panel",
    runCount: 1,
    measuredRuns: 1,
    totalCostUsd: 3.1,
  };

  beforeEach(() => {
    fetchStatsViewMock.mockImplementation((view: ViewName, params: StatsViewParams) => {
      if (view !== "cost-per-change") return Promise.resolve(fixtures[view]);
      if (params.breakdown !== "repo") {
        return Promise.resolve(envelope("cost-per-change", [sddRow, panelRow]));
      }
      // The row-scoped breakdown: each row's own repo total, distinct from
      // the other row's and from the (2.1 + 3.1 =) 5.2 whole-change sum
      // the pre-fix code would have returned for both.
      const total = params.stage === sddRow.stage ? sddRow.totalCostUsd : panelRow.totalCostUsd;
      return Promise.resolve(
        envelope("cost-per-change", [
          { repoRoot: "/repos/a", runCount: 1, measuredRuns: 1, totalTokensInput: 1, totalCostUsd: total, totalDurationMs: 1 },
        ]),
      );
    });
  });

  it("each row's expanded panel reconciles with that row's own total, and the two differ", async () => {
    const user = userEvent.setup();
    render(<CostPerChange period={period} project={undefined} />);

    const toggles = await screen.findAllByRole("button", { name: /Show repository breakdown/ });
    expect(toggles).toHaveLength(2);

    await user.click(toggles[0]);
    const firstPanel = await screen.findByText(/Per-repository breakdown for/);
    expect(firstPanel.closest("table")).toHaveTextContent("$2.10");
    expect(firstPanel.closest("table")).not.toHaveTextContent("$5.20");

    await user.click(toggles[1]);
    const panels = screen.getAllByText(/Per-repository breakdown for/);
    const secondPanel = panels[panels.length - 1];
    expect(secondPanel.closest("table")).toHaveTextContent("$3.10");
    expect(secondPanel.closest("table")).not.toHaveTextContent("$5.20");
  });
});

describe("the state board is the only navigation path into a change's own dashboard", () => {
  it("links a row's change name to its run-detail route, round-tripping a URL-significant character in both segments", async () => {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(
        envelope(view, [
          {
            projectKey: "kan/16",
            name: "kan-16 myflow-stats-app?v2",
            state: "IN_PROGRESS",
            updatedAt: "2026-01-15T10:00:00Z",
            updatedBy: "alice",
            nextCommand: "/myflow-finish",
          },
        ]),
      ),
    );

    render(<StateBoard period={period} project={undefined} />);

    const link = await screen.findByRole("link", { name: "kan-16 myflow-stats-app?v2" });
    expect(link).toHaveAttribute(
      "href",
      `#/run/${encodeURIComponent("kan/16")}/${encodeURIComponent("kan-16 myflow-stats-app?v2")}`,
    );
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
