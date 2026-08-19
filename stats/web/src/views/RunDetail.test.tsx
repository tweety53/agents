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
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { VIEW_NAMES, type CostPerChangeRow, type ListStageRunsResponse, type StageRunDTO, type StatsResponse } from "../api";
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

// Task 5 (kan-201): a stage run opens onto its own dispatches --
// specs/myflow-stats-views/spec.md, "A stage run opens onto its own
// dispatches". The fixture mirrors internal/harvest/attribute.go's real
// DispatchBucket shape (KAN-201's own tasks.md, task 5's header): `tokens`
// carries the dispatch's own token figures (or is absent when a dispatch's
// harvest recorded none), and `agent_type`/`description`/`model`/
// `spawn_depth` are each independently omitted -- never "" or 0 -- when no
// meta sidecar was found for that dispatch. `spawn_depth` is a *string on
// harvest.DispatchBucket itself (`json:"spawn_depth,omitempty"`,
// attribute.go), not a custom-marshalled int -- deliberately:
// jsonb_deep_add SUMS two JSON numbers at one key, so a numeric depth
// re-sent across harvest cycles would accumulate instead of replacing
// (KAN-201 panel finding F12). Every writer converts the plain int
// ReadDispatchMeta reports with strconv.Itoa at the point of
// construction (watcher.go); encoding/json's own struct handling suffices
// to write it as a JSON string, with no MarshalJSON needed (KAN-201 panel
// finding F23 removed the shadow struct and hand-written
// MarshalJSON/UnmarshalJSON this used to require) -- these fixtures carry
// the string form so they keep matching what the producer actually
// writes. "agent-1" carries
// both tokens and descriptors (the ordinary case); "agent-2" carries
// tokens but no descriptors (sidecar never found); "agent-3" carries
// descriptors but no tokens at all.
//
// Per-dispatch cost is a key store.Store.Price (pricing.go) writes
// directly -- "dispatches.<agentId>.cost_usd" -- through the identical
// rate resolution and chargeableTokens.cost arithmetic as "models.
// <model>.cost_usd", applied to that dispatch's own recorded model and
// own tokens. RunDetail.tsx reads it as-is; it is never a second,
// derived calculation (an implied average rate scaled from a model
// bucket's blended total, which is what this route used to do and which
// overstates a cache-heavy dispatch while understating an output-heavy
// one). "agent-1" alone carries both tokens and a priced cost_usd, so it
// alone gets a real cost; "agent-2" has no recorded model (so Price never
// prices it), and "agent-3" has no tokens at all -- both must read as
// unavailable, never as a fabricated 0. The fixture carries no "models"
// bucket at all, which is deliberate: it proves the dispatch row's cost
// comes from the dispatch's own "cost_usd", not from deriving one out of
// a sibling "models" bucket that, here, does not even exist.
describe("a stage run that dispatched three subagents (kan-201)", () => {
  const runWithDispatches: StageRunDTO = {
    stageRunId: 30,
    harness: "claude-code",
    command: "/myflow-do",
    stage: "5. The review panel",
    attempt: 1,
    startedAt: "2026-01-01T00:00:00Z",
    endedAt: "2026-01-01T00:20:00Z",
    outcome: "committed",
    metrics: {
      dispatches: {
        "agent-1": {
          tokens: { main: {}, sidechain: { input: 80000, output: 16000, cache_creation: 0, cache_read: 0, thinking: 0 } },
          agent_type: "general-purpose",
          description: "Review slot A",
          model: "claude-sonnet-5",
          spawn_depth: "1",
          cost_usd: 1.6,
        },
        "agent-2": {
          tokens: { main: {}, sidechain: { input: 20000, output: 4000, cache_creation: 0, cache_read: 0, thinking: 0 } },
          // no agent_type/description/model/cost_usd -- sidecar never found
        },
        "agent-3": {
          agent_type: "general-purpose",
          description: "Review slot C",
          model: "claude-sonnet-5",
          spawn_depth: "1",
          // no "tokens" key at all, so no cost_usd either
        },
      },
    },
  };

  async function renderExpanded() {
    listStageRunsMock.mockResolvedValue(stageRunsResponse([runWithDispatches]));
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    const user = userEvent.setup();
    // The dispatch table rides the stage run's own per-row detail toggle
    // (StageRunTable.tsx's MetricsDetail) rather than a second,
    // dispatch-specific affordance -- see that file's own header comment
    // (F3, pass 1 of this change's own review panel). The row is found by
    // its stage cell, and its toggle by scoping to that row.
    const row = (await screen.findByRole("cell", { name: "5. The review panel" })).closest("tr") as HTMLElement;
    await user.click(within(row).getByRole("button"));
    return screen.getAllByTestId("dispatch-row");
  }

  it("expands a stage run into its dispatches", async () => {
    const rows = await renderExpanded();
    expect(rows).toHaveLength(3);
  });

  it("orders dispatch rows by cost descending", async () => {
    const rows = await renderExpanded();
    // "agent-1" is the only dispatch that carries a "cost_usd" -- it must
    // sort first, ahead of both dispatches with no computable cost
    // (absent sorts last, never as zero).
    expect(within(rows[0]).getByText("Review slot A")).toBeInTheDocument();
  });

  it("reads the dispatch's own cost_usd rather than deriving one", async () => {
    const rows = await renderExpanded();
    // The fixture carries no "models" bucket at all, so any value shown
    // here can only have come from "dispatches.agent-1.cost_usd" itself
    // -- there is nothing else in the metrics bag left to derive it from.
    const agent1Row = rows.find((row) => within(row).queryByText("Review slot A")) as HTMLElement;
    expect(within(agent1Row).getByText("$1.60")).toBeInTheDocument();
  });

  it("renders an unmeasured dispatch as unavailable", async () => {
    const rows = await renderExpanded();
    const noTokensRow = rows.find((row) => within(row).queryByText("Review slot C"));
    expect(noTokensRow).toBeDefined();
    // Tokens and cost are both unmeasured for this dispatch.
    expect(within(noTokensRow as HTMLElement).getAllByTestId("unavailable").length).toBeGreaterThanOrEqual(2);
    expect(within(noTokensRow as HTMLElement).queryByText("0")).not.toBeInTheDocument();
    expect(within(noTokensRow as HTMLElement).queryByText("$0.0000")).not.toBeInTheDocument();
  });

  it("adds no view", async () => {
    listStageRunsMock.mockResolvedValue(stageRunsResponse([runWithDispatches]));
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    await screen.findByRole("cell", { name: "5. The review panel" });
    // The interface still serves the same eight statistics views.
    expect(VIEW_NAMES).toHaveLength(8);
    // Per-dispatch rows appear only under an expanded stage run.
    expect(screen.queryByTestId("dispatch-row")).not.toBeInTheDocument();
  });
});

// Two stage runs sharing the same `stage` label AND the same `attempt`
// number are the ordinary case, not a contrived one: /myflow-do records
// one stage run per task, all under the same stage label, and a first
// try on each carries the same attempt number (1). This used to require a
// toggle identifier synthesized from `startedAt`, because the dispatch
// toggles rendered as a flat list with no table-row context to tell two of
// them apart. Nested in StageRunTable's own per-row detail instead (F3,
// pass 1 of this change's own review panel), each toggle lives inside its
// own <tr>, keyed by `stageRunId` -- unique per row by construction, so
// there is no longer an identification problem to test for; what remains
// worth pinning is that expanding one row's toggle never reveals another
// row's dispatches.
describe("two stage runs sharing both stage and attempt (kan-201)", () => {
  const sharedRunA: StageRunDTO = {
    stageRunId: 40,
    harness: "claude-code",
    command: "/myflow-do",
    stage: "SDD + TDD per task",
    attempt: 1,
    startedAt: "2026-01-01T00:00:00Z",
    endedAt: "2026-01-01T00:05:00Z",
    outcome: "committed",
    metrics: {
      dispatches: {
        "agent-a": {
          tokens: { main: { input: 10 } },
          description: "Task A dispatch",
          model: "claude-sonnet-5",
          cost_usd: 1,
        },
      },
    },
  };
  const sharedRunB: StageRunDTO = {
    stageRunId: 41,
    harness: "claude-code",
    command: "/myflow-do",
    stage: "SDD + TDD per task",
    attempt: 1,
    startedAt: "2026-01-01T01:00:00Z",
    endedAt: "2026-01-01T01:05:00Z",
    outcome: "committed",
    metrics: {
      dispatches: {
        "agent-b": {
          tokens: { main: { input: 20 } },
          description: "Task B dispatch",
          model: "claude-sonnet-5",
          cost_usd: 2,
        },
      },
    },
  };

  it("expands only the clicked run's own dispatches, not the other run sharing its stage and attempt", async () => {
    listStageRunsMock.mockResolvedValue(stageRunsResponse([sharedRunA, sharedRunB]));
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    // No sort is applied, so the table rows come back in fetch order:
    // header row first, then sharedRunA, then sharedRunB -- both sharing
    // the same visible stage/attempt text, so only row position (backed by
    // each row's own `stageRunId` as DataTable's rowKey) tells them apart.
    const rows = await screen.findAllByRole("row");
    const [rowA, rowB] = [rows[1], rows[2]];

    const user = userEvent.setup();
    await user.click(within(rowA).getByRole("button"));

    const dispatchRows = screen.getAllByTestId("dispatch-row");
    expect(dispatchRows).toHaveLength(1);
    expect(within(dispatchRows[0]).getByText("Task A dispatch")).toBeInTheDocument();
    expect(screen.queryByText("Task B dispatch")).not.toBeInTheDocument();
    expect(within(rowB).queryByTestId("dispatch-row")).not.toBeInTheDocument();
  });
});
