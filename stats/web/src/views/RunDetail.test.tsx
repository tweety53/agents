// RunDetail opens one change on its own dashboard of stage runs
// (the stats-views requirement "One change opens on its own
// dashboard"). These tests cover the route's own non-negotiable rules:
// every stage run is shown attributed to its command, an unmeasured
// metric on an otherwise-measured run reads as unavailable rather than
// zero, an open run (no endedAt) reads as still running rather than
// zero-length, and the header totals come from the server's own
// cost-per-change aggregate -- never from summing the (possibly partial)
// page of stage runs the interface holds.
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  VIEW_NAMES,
  type CostPerChangeRow,
  type ListStageRunsResponse,
  type StageRunDTO,
  type StatsResponse,
} from "../api";
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

const PROJECT = "kan-16-stats-app";
const CHANGE = "kan-16-stats-app";

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
          command: "/flow",
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
          command: "/flow",
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

    // No sessionId: both runs group under their command, as one session.
    expect(await screen.findByRole("row", { name: /^Session 1 · other/ })).toBeInTheDocument();
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
          command: "/flow",
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
    expect(within(row).getByText("5m")).toBeInTheDocument();
    // Cost and input tokens are both unmeasured on this run.
    expect(within(row).getAllByTestId("unavailable").length).toBeGreaterThanOrEqual(2);
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
          command: "/flow",
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

    const stageRuns = within(await screen.findByRole("region", { name: "Stage runs" }));
    const row = (await stageRuns.findByRole("cell", { name: "SDD + TDD per task" })).closest("tr") as HTMLElement;
    expect(within(row).getByTestId("run-open")).toHaveTextContent(/^running$/);
    expect(within(row).queryByText("0m")).not.toBeInTheDocument();
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
            command: "/flow",
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
            command: "/flow",
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
          command: "/flow",
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

    const totalCost = await screen.findByRole("group", { name: "Total cost" });
    expect(within(totalCost).getByText("$99.99")).toBeInTheDocument();
    expect(within(totalCost).queryByText("$2.00")).not.toBeInTheDocument();
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
          command: "/flow",
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
          command: "/flow",
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
            command: "/flow",
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
          command: "/flow",
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

    const totalCostPanel = await screen.findByRole("group", { name: "Total cost" });
    expect(await within(totalCostPanel).findByText("$99.99")).toBeInTheDocument();
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
            command: "/flow",
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
          command: "/flow",
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
    expect(screen.getByText(/^Project agents/)).toBeInTheDocument();
    expect(screen.queryByText(new RegExp(`Project ${HASHED_PROJECT}`))).not.toBeInTheDocument();
  });
});

// Task 5 (kan-201): a stage run opens onto its own dispatches --
// the stats-views requirement, "A stage run opens onto its own
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
    command: "/flow",
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
    const row = within(await screen.findByRole("table")).getByText("5. The review panel").closest("tr") as HTMLElement;
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
    // Its cost is unmeasured.
    expect(within(noTokensRow as HTMLElement).getAllByTestId("unavailable").length).toBeGreaterThanOrEqual(1);
    expect(within(noTokensRow as HTMLElement).queryByText("0")).not.toBeInTheDocument();
    expect(within(noTokensRow as HTMLElement).queryByText("$0.0000")).not.toBeInTheDocument();
  });

  it("adds no view", async () => {
    listStageRunsMock.mockResolvedValue(stageRunsResponse([runWithDispatches]));
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    within(await screen.findByRole("table")).getByText("5. The review panel");
    // The three surviving statistics views from design.md's `stats-ui-cut`
    // (cache efficiency and runs cut by KAN-989) plus `weekly-costs`,
    // `reviewers`, `decisions`, `flow-health`, `self-review` and
    // `review-dynamics` (KAN-1046).
    expect(VIEW_NAMES).toHaveLength(9);
    // Per-dispatch rows appear only under an expanded stage run.
    expect(screen.queryByTestId("dispatch-row")).not.toBeInTheDocument();
  });
});

// Two stage runs sharing the same `stage` label AND the same `attempt`
// number are the ordinary case, not a contrived one: /flow records
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
    command: "/flow",
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
    command: "/flow",
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

    // Stage runs come back in start order under their session: both share
    // the same visible stage text, so only row position (backed by each
    // row's own `stageRunId` as its key) tells them apart.
    await screen.findByRole("table");
    const [rowA, rowB] = screen.getAllByRole("row").filter((r) => r.classList.contains("lb-row"));

    const user = userEvent.setup();
    await user.click(within(rowA).getByRole("button"));

    const dispatchRows = screen.getAllByTestId("dispatch-row");
    expect(dispatchRows).toHaveLength(1);
    expect(within(dispatchRows[0]).getByText("Task A dispatch")).toBeInTheDocument();
    expect(screen.queryByText("Task B dispatch")).not.toBeInTheDocument();
    expect(within(rowB).queryByTestId("dispatch-row")).not.toBeInTheDocument();
  });
});

// The opened run as the approved prototype draws it (kan-1048).
describe("the reworked run view (kan-1048)", () => {
  const NOW = new Date("2026-01-01T03:00:00Z");
  const run = (id: number, stage: string, start: string, end: string | undefined, extra: Partial<StageRunDTO> = {}): StageRunDTO => ({
    stageRunId: id,
    harness: "claude-code",
    command: "/flow",
    stage,
    attempt: 1,
    startedAt: start,
    endedAt: end,
    outcome: end ? "completed" : undefined,
    metrics: {},
    ...extra,
  });
  const planning = [
    run(1, "flow.kickoff", "2026-01-01T00:00:00Z", "2026-01-01T00:10:00Z", { sessionId: "s-a", metrics: { cost_usd: 1, effort: "medium", models: { "claude-opus-4-7": {} }, tokens: { main: { input: 50000, cache_read: 1000000, cache_creation: 200000 } } } }),
    run(2, "flow.brainstorm", "2026-01-01T00:10:00Z", "2026-01-01T00:40:00Z", { sessionId: "s-a", metrics: { cost_usd: 2, models: { "claude-opus-4-7": {} } } }),
  ];
  const implementation = [
    run(3, "flow.sdd-tdd", "2026-01-01T01:00:00Z", "2026-01-01T01:30:00Z", { sessionId: "s-b", outcome: "failed", metrics: { cost_usd: 3, effort: "high", models: { "claude-opus-4-7": {} } } }),
    run(4, "flow.sdd-tdd", "2026-01-01T02:00:00Z", undefined, {
      sessionId: "s-b",
      metrics: {
        cost_usd: 0.5,
        effort: "high",
        models: { "claude-opus-4-7": {}, "claude-sonnet-4-6": {} },
        dispatches: { "agent-1": { description: "Task 4/11", model: "claude-sonnet-4-6", cost_usd: 0.25, signals: { served_efforts: { low: 1, medium: 3 } } } },
      },
    }),
  ];
  const boardRow = {
    projectKey: PROJECT,
    name: CHANGE,
    state: "IN_PROGRESS",
    updatedAt: "2026-01-01T02:59:00Z",
    updatedBy: "alice",
    nextCommand: "/flow",
    currentTasks: 11,
    doneTasks: 3,
    runCount: 2,
    costUsd: 6.5,
    wallClockMs: 0,
    errors: 0,
    currentStage: "flow.sdd-tdd",
    currentStageStartedAt: "2026-01-01T02:00:00Z",
    activity: "question",
    activityDetail: "Proceed?",
    activitySince: "2026-01-01T02:58:00Z",
    tasks: [
      { n: 1, title: "Schema", state: "done" },
      { n: 2, title: "Board card", state: "in progress" },
      { n: 3, title: "Run view", state: "pending" },
    ],
    subagents: [
      { role: "implementer", taskId: "2", key: "task-2-implementer", model: "opus", modelVersion: "claude-opus-5-5", effort: "high", startedAt: "2026-01-01T02:50:00Z" },
      { role: "implementer", taskId: "3", key: "task-3-implementer", model: "opus", startedAt: "2026-01-01T02:55:00Z", inline: true },
    ],
  };

  beforeEach(() => {
    vi.useFakeTimers({ toFake: ["Date"], now: NOW });
    listStageRunsMock.mockResolvedValue(stageRunsResponse([...planning, ...implementation]));
    fetchStatsViewMock.mockImplementation((view: string) =>
      Promise.resolve(
        view === "state-board"
          ? { ...aggregateEnvelope([]), view, rows: [boardRow] }
          : aggregateEnvelope([
              {
                projectKey: PROJECT,
                changeName: CHANGE,
                command: "/flow",
                stage: "flow.sdd-tdd",
                runCount: 4,
                measuredRuns: 3,
                totalTokensInput: 1250000,
                meanTokensInput: 1,
                totalCostUsd: 6.5,
                totalDurationMs: 1,
                mainTokens: 1,
                sidechainTokens: 1,
              },
            ]),
      ),
    );
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("run detail groups stage runs by session and phase", async () => {
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const groups = (await screen.findAllByRole("row")).filter((r) => r.classList.contains("lb-group"));
    expect(groups.map((g) => g.textContent)).toEqual([
      "Session 1 · planning00:00–00:40 · 40m · $3.00",
      "Session 2 · implementation01:00–now · 2h 00m · $3.50",
    ]);
  });

  it("run detail shows model version and effort", async () => {
    const user = userEvent.setup();
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const kickoff = (await screen.findByRole("cell", { name: "Starting up flow.kickoff" })).closest("tr")!;
    expect(kickoff).toHaveTextContent("opus-4.7 · medium");
    // No effort recorded: the model alone.
    const brainstorm = screen.getByRole("cell", { name: "Designing the change flow.brainstorm" }).closest("tr")!;
    expect(within(brainstorm).getByText("opus-4.7")).toBeInTheDocument();
    const open = screen.getByRole("button", { name: "Show dispatches" });
    expect(open.closest("tr")).toHaveTextContent("opus-4.7, sonnet-4.6 · high");
    // A dispatch row: its bucket's model and its most-served effort.
    await user.click(open);
    const dispatch = screen.getByTestId("dispatch-row");
    expect(dispatch).toHaveTextContent("Task 4/11");
    expect(dispatch).toHaveTextContent("sonnet-4.6 · medium");
    expect(dispatch).toHaveTextContent("$0.25");
  });

  it("run detail has no filter controls", async () => {
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    await screen.findByRole("cell", { name: "Starting up flow.kickoff" });
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
    expect(screen.queryByRole("searchbox")).not.toBeInTheDocument();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(screen.queryByRole("toolbar")).not.toBeInTheDocument();
  });

  it("run detail header and kpis", async () => {
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    expect(await screen.findByRole("link", { name: "← Live state board" })).toHaveAttribute("href", "#/state-board");
    await screen.findByText("Proceed?");
    expect(screen.getByText(/^Project kan-16-stats-app · KAN-16 · updated 1m ago$/)).toBeInTheDocument();
    // The board row is read through past now, so a change updated this instant is found.
    const board = fetchStatsViewMock.mock.calls.find(([view]) => view === "state-board")!;
    expect(board[1].to.getTime()).toBeGreaterThan(NOW.getTime());
    // ...scoped to this one change server-side, never the project's whole board.
    expect(board[1].change).toBe(CHANGE);
    expect(screen.getByRole("heading", { name: CHANGE }).parentElement).toHaveTextContent("IN_PROGRESS");
    const kpi = (name: string) => screen.getByRole("group", { name });
    expect(kpi("Now running")).toHaveTextContent("Implementing tasks");
    expect(kpi("Now running")).toHaveTextContent("for 1h 00m · tasks 3/11");
    expect(kpi("Now running")).toHaveTextContent("Awaiting your answer");
    expect(kpi("Total cost")).toHaveTextContent("$6.50");
    expect(kpi("Total cost")).toHaveTextContent("4 stage runs, 3 measured");
    expect(kpi("Wall clock")).toHaveTextContent("3h 00m");
    expect(kpi("Wall clock")).toHaveTextContent("active 2h 10m");
    // Input counts cache reads and writes too, never just the uncached sliver.
    expect(kpi("Input tokens")).toHaveTextContent("1.25M");
    expect(kpi("Input tokens")).toHaveTextContent("cache reads 80%");
    // Running subagents name what they do and their model, version and effort; the plan lists every task.
    const subs = screen.getByRole("region", { name: "Agents" });
    expect(subs).toHaveTextContent("implementer");
    expect(subs).toHaveTextContent("opus-5.5 · high");
    expect(subs).toHaveTextContent("10m");
    // The main session leads, marked Main, on its inline task; each subagent is marked Subagent.
    const agentRows = [...subs.querySelectorAll("tbody tr")];
    expect(agentRows).toHaveLength(2);
    expect(agentRows[0]).toHaveTextContent("Mainneeds youImplementing task 3 (Run view)");
    expect(agentRows[1]).toHaveTextContent("⎿ implementerrunningImplementing task 2 (Board card)opus-5.5 · high10m");
    expect(subs.querySelector(".lb-panel-note")).toHaveTextContent("2");
    const tasks = screen.getByRole("region", { name: "Tasks" });
    expect(tasks).toHaveTextContent("1/3 done");
    expect(tasks).toHaveTextContent("✔ 1. Schema");
    expect(tasks).toHaveTextContent("◼ 2. Board card");
    expect(tasks).toHaveTextContent("◻ 3. Run view");
    expect(kpi("Failed runs")).toHaveTextContent("1");
    expect(kpi("Failed runs")).toHaveTextContent("1 retried");
    // Cost by stage: summed per stage, largest first, in plain words.
    const rail = screen.getByRole("region", { name: "Cost by stage" });
    expect(within(rail).getAllByText(/^(Implementing tasks|Designing the change|Starting up)$/).map((e) => e.textContent)).toEqual([
      "Implementing tasks",
      "Designing the change",
      "Starting up",
    ]);
    expect(rail).toHaveTextContent("$3.50");
  });

  it("run detail marks main as waiting while only its subagents work", async () => {
    const waiting = { ...boardRow, activity: "working", subagents: boardRow.subagents.filter((d) => !d.inline) };
    fetchStatsViewMock.mockImplementation((view: string) =>
      Promise.resolve(view === "state-board" ? { ...aggregateEnvelope([]), view, rows: [waiting] } : aggregateEnvelope([])),
    );
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const main = (await screen.findByRole("cell", { name: "Main" })).closest("tr")!;
    expect(main).toHaveTextContent("Mainwaiting for subagentsImplementing tasks");
    expect(main.querySelector(".lb-dot.waiting")).not.toBeNull();
  });

  it("run detail lists main while it works between stages", async () => {
    const between = {
      ...boardRow,
      currentStage: undefined,
      activity: "working",
      tasks: [{ n: 4, title: "In-run fix 1 — visual-verify: divider", state: "in progress" }],
      subagents: [{ role: "implementer", taskId: "4", model: "opus", startedAt: "2026-01-01T02:58:00Z", inline: true }],
    };
    listStageRunsMock.mockResolvedValue(stageRunsResponse(planning));
    fetchStatsViewMock.mockImplementation((view: string) =>
      Promise.resolve(view === "state-board" ? { ...aggregateEnvelope([]), view, rows: [between] } : aggregateEnvelope([])),
    );
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const main = (await screen.findByRole("cell", { name: "Main" })).closest("tr")!;
    expect(main).toHaveTextContent("Mainrunningtask 4 (In-run fix 1 — visual-verify: divider)");
    // Now running names the in-run fix even with no stage open.
    const kpi = screen.getByRole("group", { name: "Now running" });
    expect(kpi).toHaveTextContent("Working between stages");
    expect(kpi).toHaveTextContent("Task 4: In-run fix 1 — visual-verify: divider");
  });

  it("run detail still shows its stage runs when the board row fails to load", async () => {
    fetchStatsViewMock.mockImplementation((view: string) =>
      view === "state-board" ? Promise.reject(new Error("board down")) : Promise.resolve(aggregateEnvelope([])),
    );
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    expect(await screen.findByRole("cell", { name: "Starting up flow.kickoff" })).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("stage timeline draws a lane per stage coloured by outcome", async () => {
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const svg = await screen.findByRole("img", { name: "Stage runs over time" });
    // Each lane reads in plain words, its stage id on hover.
    const lanes = [...svg.querySelectorAll("text.lb-lane")];
    expect(lanes.map((t) => t.lastChild?.textContent)).toEqual(["Starting up", "Designing the change", "Implementing tasks"]);
    expect(lanes.map((t) => t.querySelector("title")?.textContent)).toEqual(["flow.kickoff", "flow.brainstorm", "flow.sdd-tdd"]);
    const fills = [...svg.querySelectorAll("rect.lb-bar")].map((r) => r.getAttribute("fill"));
    expect(fills).toEqual(["var(--status-ok)", "var(--status-ok)", "var(--status-critical)", "url(#lb-hatch)"]);
    expect(svg.querySelector("line.lb-now")).not.toBeNull();
    expect(svg).toHaveAttribute("width", "100%");
  });
});
