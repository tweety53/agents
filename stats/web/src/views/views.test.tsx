// One test per view asserting it renders the actual numbers a fixture
// response carries (never merely that the component mounted -- the
// highest-risk vacuous shape this task's own instructions name), plus the
// absence-vs-zero distinction and the state-board-is-default-route
// requirement.
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { StatsResponse, ViewName } from "../api";
import { App } from "../App";
import { CacheEfficiency } from "./CacheEfficiency";
import { StageLeaderboard } from "./StageLeaderboard";
import { StateBoard } from "./StateBoard";
import { Trend } from "./Trend";

const { fetchStatsViewMock } = vi.hoisted(() => ({ fetchStatsViewMock: vi.fn() }));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return { ...actual, fetchStatsView: fetchStatsViewMock };
});

const period = { from: new Date("2026-01-01T00:00:00Z"), to: new Date("2026-02-01T00:00:00Z") };

function envelope<Row>(view: ViewName, rows: Row, recorded = true, unmeasured = false): StatsResponse<Row> {
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
            nextCommand: "/myflow-finish",
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
            nextCommand: "/myflow-finish",
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
            nextCommand: "/myflow-finish",
          },
          {
            projectKey: KEY_TWO,
            name: "kan-2",
            state: "STARTED",
            updatedAt: "2026-01-15T11:00:00Z",
            updatedBy: "bob",
            nextCommand: "/myflow-do",
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
