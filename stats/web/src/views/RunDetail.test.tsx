// RunDetail opens one change on its own dashboard of stage runs
// (specs/myflow-stats-views/spec.md, "One change opens on its own
// dashboard"). These tests cover the route's own non-negotiable rules:
// every stage run is shown attributed to its command, an unmeasured
// metric on an otherwise-measured run reads as unavailable rather than
// zero, an open run (no endedAt) reads as still running rather than
// zero-length, and the header totals come from the server's own
// cost-per-change aggregate -- never from summing the (possibly partial)
// page of stage runs the interface holds.
import { render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CostPerChangeRow, ListStageRunsResponse, StageRunDTO, StatsResponse } from "../api";
import { RunDetail } from "./RunDetail";

const { listStageRunsMock, fetchStatsViewMock } = vi.hoisted(() => ({
  listStageRunsMock: vi.fn(),
  fetchStatsViewMock: vi.fn(),
}));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return { ...actual, listStageRuns: listStageRunsMock, fetchStatsView: fetchStatsViewMock };
});

function stageRunsResponse(stageRuns: StageRunDTO[], total?: number): ListStageRunsResponse {
  return { total: total ?? stageRuns.length, stageRuns };
}

function aggregateEnvelope(rows: CostPerChangeRow[]): StatsResponse<CostPerChangeRow[]> {
  return {
    view: "cost-per-change",
    from: "2020-01-01T00:00:00Z",
    to: "2026-01-01T00:00:00Z",
    boundaryConvention: "a stage run is attributed to the period containing its start instant",
    recorded: true,
    // Unanticipated edit, task 5: StatsResponse gained a required
    // `unmeasured` field (the third arm of the absence distinction), and
    // this fixture predates it. This route always requests an
    // already-recorded period, so `false` is the correct fixture value,
    // not merely a compiling one -- mechanical addition, no logic change.
    unmeasured: false,
    rows,
  };
}

const PROJECT = "kan-16-myflow-stats-app";
const CHANGE = "kan-16-myflow-stats-app";

beforeEach(() => {
  listStageRunsMock.mockReset();
  fetchStatsViewMock.mockReset();
});

describe("a change with runs across several commands", () => {
  it("lists every stage run, each attributed to the command it ran under", async () => {
    listStageRunsMock.mockResolvedValue(
      stageRunsResponse([
        {
          stageRunId: 1,
          harness: "claude-code",
          command: "/myflow-start",
          stage: "1. Brainstorming",
          attempt: 1,
          startedAt: "2026-01-01T00:00:00Z",
          endedAt: "2026-01-01T00:10:00Z",
          outcome: "committed",
          metrics: { cost_usd: 1.1 },
        },
        {
          stageRunId: 2,
          harness: "claude-code",
          command: "/myflow-do",
          stage: "SDD + TDD per task",
          attempt: 1,
          startedAt: "2026-01-01T00:10:00Z",
          endedAt: "2026-01-01T00:30:00Z",
          outcome: "committed",
          metrics: { cost_usd: 2.2 },
        },
      ]),
    );
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    expect(await screen.findByRole("cell", { name: "/myflow-start" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "/myflow-do" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "1. Brainstorming" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "SDD + TDD per task" })).toBeInTheDocument();
  });
});

describe("a stage run that recorded no token metrics", () => {
  it("shows the duration and reads its token and currency figures as unavailable, not zero", async () => {
    listStageRunsMock.mockResolvedValue(
      stageRunsResponse([
        {
          stageRunId: 3,
          harness: "claude-code",
          command: "/myflow-do",
          stage: "SDD + TDD per task",
          attempt: 1,
          startedAt: "2026-01-01T00:00:00Z",
          endedAt: "2026-01-01T00:05:00Z",
          outcome: "committed",
          metrics: {}, // wall-clock time is derivable from started/endedAt, but no metrics bag at all
        },
      ]),
    );
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    const row = (await screen.findByRole("cell", { name: "SDD + TDD per task" })).closest("tr") as HTMLElement;
    expect(within(row).getByText("5.0 min")).toBeInTheDocument();
    // Cost, input tokens and output tokens are all unmeasured on this run.
    expect(within(row).getAllByTestId("unavailable").length).toBeGreaterThanOrEqual(3);
    expect(within(row).queryByText("0")).not.toBeInTheDocument();
    expect(within(row).queryByText("$0.0000")).not.toBeInTheDocument();
  });
});

describe("an open stage run", () => {
  it("reads as still running rather than zero-length", async () => {
    listStageRunsMock.mockResolvedValue(
      stageRunsResponse([
        {
          stageRunId: 4,
          harness: "claude-code",
          command: "/myflow-do",
          stage: "SDD + TDD per task",
          attempt: 1,
          startedAt: "2026-01-01T00:00:00Z",
          outcome: undefined,
          metrics: {},
        },
      ]),
    );
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    const row = (await screen.findByRole("cell", { name: "SDD + TDD per task" })).closest("tr") as HTMLElement;
    expect(within(row).getByTestId("run-open")).toHaveTextContent(/still running/i);
    expect(within(row).queryByText("0 ms")).not.toBeInTheDocument();
  });
});

describe("a change with more stage runs than one page holds", () => {
  it("shows the header totals from the server's aggregate rather than from the fetched page of stage runs", async () => {
    // Only two stage runs come back (a short page), but the change really
    // has far more, and the aggregate row already reflects the true total
    // cost -- $99.99 across 50 runs -- while summing just the two returned
    // rows' own cost_usd (1 + 1 = 2) would silently under-report it.
    listStageRunsMock.mockResolvedValue(
      stageRunsResponse(
        [
          {
            stageRunId: 5,
            harness: "claude-code",
            command: "/myflow-do",
            stage: "SDD + TDD per task",
            attempt: 1,
            startedAt: "2026-01-01T00:00:00Z",
            endedAt: "2026-01-01T00:05:00Z",
            outcome: "committed",
            metrics: { cost_usd: 1 },
          },
          {
            stageRunId: 6,
            harness: "claude-code",
            command: "/myflow-do",
            stage: "SDD + TDD per task",
            attempt: 2,
            startedAt: "2026-01-01T00:05:00Z",
            endedAt: "2026-01-01T00:10:00Z",
            outcome: "committed",
            metrics: { cost_usd: 1 },
          },
        ],
        50,
      ),
    );
    fetchStatsViewMock.mockResolvedValue(
      aggregateEnvelope([
        {
          projectKey: PROJECT,
          changeName: CHANGE,
          command: "/myflow-do",
          stage: "SDD + TDD per task",
          runCount: 50,
          measuredRuns: 50,
          totalTokensInput: 500000,
          meanTokensInput: 10000,
          totalCostUsd: 99.99,
          totalDurationMs: 12345678,
          mainTokens: 400000,
          sidechainTokens: 100000,
        },
      ]),
    );

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    expect(await screen.findByText("$99.99")).toBeInTheDocument();
    expect(screen.queryByText("$2.00")).not.toBeInTheDocument();
  });

  it("sums the aggregate across every command/stage row of the change", async () => {
    // The request sends both "project" and "change" (useRunDetail.ts), so
    // a real server never returns a row for another change here -- this
    // mock reflects exactly that scoped response, not a wider one filtered
    // down client-side (task 25, step 2: the client-side changeName filter
    // that used to do this was a no-op the moment the request started
    // sending "change" server-side, and a mock returning rows the server
    // could not is what made that dead filter look necessary).
    listStageRunsMock.mockResolvedValue(stageRunsResponse([]));
    fetchStatsViewMock.mockResolvedValue(
      aggregateEnvelope([
        {
          projectKey: PROJECT,
          changeName: CHANGE,
          command: "/myflow-start",
          stage: "1. Brainstorming",
          runCount: 1,
          measuredRuns: 1,
          totalTokensInput: 1000,
          meanTokensInput: 1000,
          totalCostUsd: 1.5,
          totalDurationMs: 60000,
          mainTokens: 900,
          sidechainTokens: 100,
        },
        {
          projectKey: PROJECT,
          changeName: CHANGE,
          command: "/myflow-do",
          stage: "SDD + TDD per task",
          runCount: 2,
          measuredRuns: 2,
          totalTokensInput: 2000,
          meanTokensInput: 1000,
          totalCostUsd: 2.5,
          totalDurationMs: 120000,
          mainTokens: 1800,
          sidechainTokens: 200,
        },
      ]),
    );

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    // 1.5 + 2.5 = 4.0.
    expect(await screen.findByText("$4.00")).toBeInTheDocument();
  });

  it("still reads the header totals from the server's aggregate after the panel recomposition (task 20)", async () => {
    // Same shape as the test above, but this one pins down *where* the
    // number lives post-recomposition: inside the "Total cost" stat
    // panel specifically, not merely somewhere on the page -- so a later
    // change that re-introduces summing the stage-run page (this file's
    // own regression) would fail here even if some other panel happened
    // to also show "$99.99" by coincidence.
    listStageRunsMock.mockResolvedValue(
      stageRunsResponse(
        [
          {
            stageRunId: 5,
            harness: "claude-code",
            command: "/myflow-do",
            stage: "SDD + TDD per task",
            attempt: 1,
            startedAt: "2026-01-01T00:00:00Z",
            endedAt: "2026-01-01T00:05:00Z",
            outcome: "committed",
            metrics: { cost_usd: 1 },
          },
        ],
        50,
      ),
    );
    fetchStatsViewMock.mockResolvedValue(
      aggregateEnvelope([
        {
          projectKey: PROJECT,
          changeName: CHANGE,
          command: "/myflow-do",
          stage: "SDD + TDD per task",
          runCount: 50,
          measuredRuns: 50,
          totalTokensInput: 500000,
          meanTokensInput: 10000,
          totalCostUsd: 99.99,
          totalDurationMs: 12345678,
          mainTokens: 400000,
          sidechainTokens: 100000,
        },
      ]),
    );

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    const totalCostPanel = await screen.findByRole("region", { name: "Total cost" });
    expect(within(totalCostPanel).getByText("$99.99")).toBeInTheDocument();
    expect(within(totalCostPanel).queryByText("$1.00")).not.toBeInTheDocument();
  });
});

describe("a change whose stage runs exceed one fetched page (task 25, step 4)", () => {
  // Task 20 filters the fetched stage-run page against each run's models
  // keys in the browser. Where a change's stage runs exceed one page, that
  // filter (or even the plain unfiltered table) shows only the runs
  // within the page fetched, with nothing saying so -- so a change with
  // far more runs than fit in one page can display a handful with no
  // indication the list is not complete. This must not happen silently.
  it("states the shortfall when the server's total exceeds the fetched page", async () => {
    listStageRunsMock.mockResolvedValue(
      stageRunsResponse(
        [
          {
            stageRunId: 20,
            harness: "claude-code",
            command: "/myflow-do",
            stage: "SDD + TDD per task",
            attempt: 1,
            startedAt: "2026-01-01T00:00:00Z",
            endedAt: "2026-01-01T00:05:00Z",
            outcome: "committed",
            metrics: {},
          },
        ],
        500, // the change really has 500 stage runs; only one page came back
      ),
    );
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    expect(await screen.findByText(/showing 1 of 500 stage runs/i)).toBeInTheDocument();
  });

  it("states no shortfall when every stage run of the change was fetched", async () => {
    listStageRunsMock.mockResolvedValue(
      stageRunsResponse([
        {
          stageRunId: 21,
          harness: "claude-code",
          command: "/myflow-do",
          stage: "SDD + TDD per task",
          attempt: 1,
          startedAt: "2026-01-01T00:00:00Z",
          endedAt: "2026-01-01T00:05:00Z",
          outcome: "committed",
          metrics: {},
        },
      ]),
    );
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    await screen.findByRole("cell", { name: "SDD + TDD per task" });
    expect(screen.queryByText(/showing .* of .* stage runs/i)).not.toBeInTheDocument();
  });
});

// Task 2 (kan-183): the header names the project by its display name, but
// the route itself -- section aria-label and useRunDetail's request -- is
// keyed by the full project key, which is not this task's to shorten.
describe("the header names the project by its display name (kan-183)", () => {
  const HASHED_PROJECT = "agents-a740d89c";

  it("shows the display name, not the raw key, while the section identity keeps the full key", async () => {
    listStageRunsMock.mockResolvedValue(stageRunsResponse([]));
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={HASHED_PROJECT} change={CHANGE} />);

    await screen.findByRole("heading", { name: CHANGE });
    expect(screen.getByText("Project agents")).toBeInTheDocument();
    expect(screen.queryByText(`Project ${HASHED_PROJECT}`)).not.toBeInTheDocument();
  });
});

describe("the run-detail dashboard's model variable (task 20's own decision: honoured, not disabled)", () => {
  const runOnSonnet: StageRunDTO = {
    stageRunId: 10,
    harness: "claude-code",
    command: "/myflow-do",
    stage: "SDD + TDD per task",
    attempt: 1,
    startedAt: "2026-01-01T00:00:00Z",
    endedAt: "2026-01-01T00:05:00Z",
    outcome: "committed",
    metrics: { cost_usd: 1, models: { "claude-sonnet-5": { tokens: { input: 100 } } } },
  };
  const runOnOpus: StageRunDTO = {
    stageRunId: 11,
    harness: "claude-code",
    command: "/myflow-do",
    stage: "5. The review panel",
    attempt: 1,
    startedAt: "2026-01-01T00:10:00Z",
    endedAt: "2026-01-01T00:15:00Z",
    outcome: "committed",
    metrics: { cost_usd: 2, models: { "claude-opus-5": { tokens: { input: 200 } } } },
  };
  const runWithNoModel: StageRunDTO = {
    stageRunId: 12,
    harness: "claude-code",
    command: "/myflow-start",
    stage: "1. Brainstorming",
    attempt: 1,
    startedAt: "2026-01-01T00:20:00Z",
    endedAt: "2026-01-01T00:25:00Z",
    outcome: "committed",
    metrics: {},
  };

  it("filters the stage-run table to runs whose metrics bag recorded the selected model", async () => {
    listStageRunsMock.mockResolvedValue(stageRunsResponse([runOnSonnet, runOnOpus, runWithNoModel]));
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} model="claude-sonnet-5" />);

    expect(await screen.findByRole("cell", { name: "SDD + TDD per task" })).toBeInTheDocument();
    expect(screen.queryByRole("cell", { name: "5. The review panel" })).not.toBeInTheDocument();
    expect(screen.queryByRole("cell", { name: "1. Brainstorming" })).not.toBeInTheDocument();
  });

  it("scopes the header to that model by sending it to the cost-per-change aggregate", async () => {
    listStageRunsMock.mockResolvedValue(stageRunsResponse([runOnSonnet]));
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} model="claude-sonnet-5" />);

    await screen.findByRole("region", { name: "Runs" });
    expect(fetchStatsViewMock).toHaveBeenCalledWith(
      "cost-per-change",
      expect.objectContaining({ project: PROJECT, change: CHANGE, model: "claude-sonnet-5" }),
    );
  });

  it("shows every run, unfiltered, when no model is selected", async () => {
    listStageRunsMock.mockResolvedValue(stageRunsResponse([runOnSonnet, runOnOpus, runWithNoModel]));
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    expect(await screen.findByRole("cell", { name: "SDD + TDD per task" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "5. The review panel" })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "1. Brainstorming" })).toBeInTheDocument();
  });
});
