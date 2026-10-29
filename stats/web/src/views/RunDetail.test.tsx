// RunDetail opens one change on its own dashboard of stage runs
// (the stats-views requirement "One change opens on its own
// dashboard"). These tests cover the route's own non-negotiable rules:
// every stage run is shown attributed to its command, an unmeasured
// metric on an otherwise-measured run reads as unavailable rather than
// zero, an open run (no endedAt) reads as still running rather than
// zero-length, and the header totals come from the server's own
// cost-per-change aggregate -- never from summing the (possibly partial)
// page of stage runs the interface holds.
import { render, screen, waitFor, within } from "@testing-library/react";
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

const { listStageRunsMock, fetchStatsViewMock, listStagesMock, listFindingsMock, runRecordMock, findChangesMock } = vi.hoisted(() => ({
  listStageRunsMock: vi.fn(),
  findChangesMock: vi.fn((_name: string) => Promise.resolve({ records: [] as unknown[] })),
  runRecordMock: vi.fn((_project: string, _change: string) => Promise.resolve({ dispatches: [] as unknown[] })),
  listFindingsMock: vi.fn((_project: string, _change: string) => Promise.resolve([] as unknown[])),
  fetchStatsViewMock: vi.fn(),
  listStagesMock: vi.fn(() => Promise.resolve({ stages: ["flow.kickoff", "flow.brainstorm", "flow.decide", "flow.sdd-tdd", "flow.review-panel", "flow.verify"] })),
}));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return { ...actual, listStageRuns: listStageRunsMock, fetchStatsView: fetchStatsViewMock, listStages: listStagesMock, listSelfReviewFindings: listFindingsMock, getRunRecord: runRecordMock, findChanges: findChangesMock };
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

/** The Stage runs panel's table, every run shown. */
async function stageRunsTable(): Promise<HTMLElement> {
  const panel = await screen.findByRole("region", { name: "Stage runs" });
  const all = within(panel).queryByRole("button", { name: /^Show all/ });
  if (all) await userEvent.click(all);
  return within(panel).getByRole("table");
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
    expect(await screen.findByRole("row", { name: /^Session 1 · Other/ })).toBeInTheDocument();
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
    const row = within(await stageRunsTable()).getByText(/^Review slot A/).closest("tr") as HTMLElement;
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

    within(await stageRunsTable()).getByText(/^Review slot A/);
    // The three surviving statistics views from design.md's `stats-ui-cut`
    // (cache efficiency and runs cut by KAN-989) plus `weekly-costs`,
    // `reviewers`, `decisions`, `flow-health`, `self-review` and
    // `review-dynamics` (KAN-1046) and `experiments` (KAN-1067).
    expect(VIEW_NAMES).toHaveLength(10);
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

  it("names each run sharing a stage and attempt by its own single dispatch, in the row itself", async () => {
    listStageRunsMock.mockResolvedValue(stageRunsResponse([sharedRunA, sharedRunB]));
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));

    render(<RunDetail project={PROJECT} change={CHANGE} />);

    await stageRunsTable();
    const [rowA, rowB] = screen.getAllByRole("row").filter((r) => r.classList.contains("lb-row"));
    expect(within(rowA).getAllByRole("cell")[0]).toHaveTextContent(/^Task A dispatch$/);
    expect(within(rowB).getAllByRole("cell")[0]).toHaveTextContent(/^Task B dispatch$/);
    expect(within(rowA).queryByRole("button")).not.toBeInTheDocument();
    expect(screen.queryByTestId("dispatch-row")).not.toBeInTheDocument();
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
      "Session 1 · Planning00:00–00:40 · 40m · $3.00",
      "Session 2 · Implementation01:00–now · 2h 00m · $3.50",
    ]);
  });

  it("run detail shows a row's agent model and effort, else its main session's, in the Model · effort column", async () => {
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const kickoff = (await screen.findByRole("cell", { name: "Starting up" })).closest("tr")!;
    expect(within(kickoff).getByTestId("run-model")).toHaveTextContent(/^opus 4\.7 · medium$/);
    // A single dispatch is the row itself, with its bucket's model and its most-served effort.
    const one = screen.getByRole("cell", { name: "Task 4/11" }).closest("tr")!;
    expect(within(one).getByTestId("run-model")).toHaveTextContent(/^sonnet 4\.6 · medium$/);
    expect(within(one).getByTestId("run-cost").title).toBe("agents $0.25 + main session $0.25");
  });

  it("run detail has no filter controls", async () => {
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    await screen.findByRole("cell", { name: "Starting up" });
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
    expect(screen.getByRole("heading", { name: CHANGE }).parentElement).toHaveTextContent("Integrating");
    const kpi = (name: string) => screen.getByRole("group", { name });
    expect(kpi("Now running")).toHaveTextContent("Implementing tasks");
    // Both counts come from the plan it lists, never the stale stored 3/11.
    expect(kpi("Now running")).toHaveTextContent("for 1h 00m · tasks 1/3");
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
    expect(subs).toHaveTextContent("opus 5.5 · high");
    expect(subs).toHaveTextContent("10m");
    // The main session leads, marked Main, on its inline task; each subagent is marked Subagent.
    const agentRows = [...subs.querySelectorAll("tbody tr")];
    expect(agentRows).toHaveLength(2);
    expect(agentRows[0]).toHaveTextContent("Mainneeds youImplementing task 3 (Run view)");
    expect(agentRows[1]).toHaveTextContent("⎿ implementerrunningImplementing task 2 (Board card)opus 5.5 · high10m");
    expect(subs.querySelector(".lb-panel-note")).toHaveTextContent("2");
    // The plan is its own Tasks card beside Main, the task being worked on marked ●.
    const tasks = screen.getByRole("region", { name: "Tasks" });
    // Plans stack in their own column, beside the agents and work panels, right under the KPI strip.
    const work = tasks.closest(".lb-work")!;
    expect(work.previousElementSibling).toHaveAttribute("aria-label", "Run totals");
    const [left, right] = [...work.children];
    expect([...left.children].map((e) => e.getAttribute("aria-label"))).toEqual(["Agents", "Main", "Self-review"]);
    expect(right).toContainElement(tasks);
    // Narrow widths fall back to one stack: agents, Main, its tasks, Self-review, its tasks.
    expect([...left.children, ...right.children].map((e) => [e.getAttribute("aria-label"), (e as HTMLElement).style.order])).toEqual([
      ["Agents", "0"],
      ["Main", "1"],
      ["Self-review", "3"],
      ["Tasks", "2"],
    ]);
    expect(within(tasks).getByText("Tasks").parentElement).toHaveTextContent(/^Tasks1\/3$/);
    expect(tasks).toHaveTextContent("✔ 1. Schema");
    expect(tasks).toHaveTextContent("● 2. Board card");
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

  it("run detail shows exactly a Main and a Self-review panel, Main first, the running self-review open and summed up inside Main", async () => {
    const impl3 = { ...implementation[0], metrics: { ...(implementation[0].metrics as object), dispatches: { "agent-impl": { cost_usd: 1 } } } };
    listStageRunsMock.mockResolvedValue(stageRunsResponse([...planning, impl3, run(5, "flow.self-review", "2026-01-01T02:00:00Z", undefined)]));
    runRecordMock.mockResolvedValueOnce({
      dispatches: [
        { id: 1, agentId: "agent-impl", role: "implementer", taskId: "1", key: "task-1-implementer", model: "opus", startedAt: "2026-01-01T01:05:00Z", endedAt: "2026-01-01T01:20:00Z", outcome: "completed" },
        { id: 2, role: "self-review", key: "self-review", model: "opus", effort: "medium", startedAt: "2026-01-01T02:00:00Z", endedAt: "2026-01-01T02:20:00Z", outcome: "completed" },
        { id: 3, role: "implementer", key: "self-review-kan-16-fix-1", model: "opus", startedAt: "2026-01-01T02:30:00Z", endedAt: "2026-01-01T02:40:00Z", outcome: "blocked" },
        { id: 4, role: "implementer", key: "self-review-kan-16-fix-v2", model: "opus", effort: "low", metrics: { modelVersion: "claude-opus-5-5" }, startedAt: "2026-01-01T02:50:00Z" },
      ],
    });
    const reviewing = {
      ...boardRow,
      currentStage: "flow.self-review",
      selfReviewTasks: [
        { n: 1, title: "Fix the divider", state: "done" },
        { n: 2, title: "Fix the count", state: "pending" },
      ],
    };
    fetchStatsViewMock.mockImplementation((view: string) =>
      Promise.resolve(view === "state-board" ? { ...aggregateEnvelope([]), view, rows: [reviewing] } : aggregateEnvelope([])),
    );
    listFindingsMock.mockResolvedValueOnce([
      { id: 1, change: CHANGE, angle: "correctness", note: "Count is stale", disposition: "fixed", ref: "abc1234", recordedAt: "2026-01-01T02:21:00Z" },
    ]);
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const panel = await screen.findByRole("region", { name: "Self-review" });
    // Two work panels and no others, Main always first.
    const work = screen.getAllByRole("region").filter((r) => r.classList.contains("lb-phase")).map((r) => r.getAttribute("aria-label"));
    expect(work).toEqual(["Main", "Self-review"]);
    // A running self-review starts open.
    expect(within(panel).getByRole("button", { name: "Hide" })).toHaveAttribute("aria-expanded", "true");
    expect(screen.queryByRole("region", { name: "Implementation" })).not.toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Agents" })).toBeInTheDocument();
    // One line per role, never a row per dispatch: count, each model · effort's count and cost, cost, failures, running.
    const roles = await within(panel).findAllByTestId("phase-role");
    expect(roles.map((r) => r.textContent)).toEqual(["self-review ×1opus · medium ×1 (—)—", "implementer ×2opus ×1 (—), opus 5.5 · low ×1 (—)—1 failedrunning"]);
    expect(within(panel).getByTestId("self-review-summary")).toHaveTextContent(/^plan 1\/2 · findings 1 · fixes round 2$/);
    // What it does now, unmistakably: the running red-verify fixer, its time and model, and where that sits.
    expect(within(panel).getByTestId("self-review-now")).toHaveTextContent(/^Now: Verifying — red verify fix round 2 · for 10m · opus 5\.5$/);
    expect(within(panel).getByTestId("self-review-progress")).toHaveTextContent("analyze ✓ → plan ✓ → fix 1/2 ✓ → re-review ✓ → verify ● → land → file");
    // Main: everything else -- its header, Now line, progress over the main phases, summary, plan and roles.
    const main = screen.getByRole("region", { name: "Main" });
    // The summary rides the header.
    expect(main.querySelector(".lb-phase-name")).toHaveTextContent(/^Main · done · 1h 10m · \$6\.00 · tasks 1\/3$/);
    // Main alone tells the whole story: the self-review's status and summary in one line.
    expect(within(main).getByTestId("main-self-review")).toHaveTextContent(/^Self-review: running · plan 1\/2 · findings 1 · fixes round 2$/);
    expect(within(main).getByTestId("main-now")).toHaveTextContent(/^Now: Running the self-review$/);
    expect(within(main).getByTestId("main-progress")).toHaveTextContent(
      "planning ✓ → decisions → implementation ✓ → review → verification → handoff → integrate → self-review ● → landing → cleanup",
    );
    expect(within(main).getByTestId("main-summary")).toHaveTextContent(/^tasks 1\/3$/);
    // The lines add up to the header: the priced dispatch, then the main session's own share.
    expect(within(main).getAllByTestId("phase-role").map((r) => r.textContent)).toEqual(["implementer ×1opus ×1 ($1.00)$1.00", "main session$5.00"]);
    expect(screen.getByRole("region", { name: "Tasks" })).toHaveTextContent("● 2. Board card");
    const srTasks = screen.getByRole("region", { name: "Self-review tasks" });
    expect(srTasks).toHaveTextContent("Self-review tasks1/2");
    expect(srTasks).toHaveTextContent("✔ 1. Fix the divider");
    expect(srTasks).toHaveTextContent("◻ 2. Fix the count");
    // Findings are counted in the summary, never listed.
    expect(within(panel).queryByTestId("self-review-finding")).not.toBeInTheDocument();
    expect(listFindingsMock).toHaveBeenCalledWith(PROJECT, CHANGE);
  });

  it("run detail merges the self-review sub-run into the Self-review phase, with its costs and findings", async () => {
    const SUB = `self-review-${CHANGE}`;
    findChangesMock.mockResolvedValue({ records: [{ projectKey: "agents-tooling", name: SUB }] });
    listStageRunsMock.mockImplementation((q: { filters?: { name?: string } }) =>
      Promise.resolve(
        stageRunsResponse(
          q.filters?.name === SUB
            ? [
                run(20, "flow.self-review", "2026-01-01T02:10:00Z", "2026-01-01T02:30:00Z", { metrics: { cost_usd: 2, dispatches: { "agent-plan": { cost_usd: 1.5 } } } }),
                run(21, "flow.self-review", "2026-01-01T02:31:00Z", undefined, { metrics: { cost_usd: 0.5 } }),
              ]
            : [...planning, run(5, "flow.self-review", "2026-01-01T02:00:00Z", undefined)],
        ),
      ),
    );
    runRecordMock.mockImplementation((_p: string, name: string) =>
      Promise.resolve({
        dispatches:
          name === SUB
            ? [
                { id: 7, agentId: "agent-plan", role: "planner", key: `${SUB}-plan`, model: "opus", startedAt: "2026-01-01T02:10:00Z", endedAt: "2026-01-01T02:29:00Z", outcome: "completed" },
                { id: 8, agentId: "agent-fix", role: "implementer", key: `${SUB}-fix-g1`, model: "opus", startedAt: "2026-01-01T02:31:00Z" },
              ]
            : [],
      }),
    );
    listFindingsMock.mockImplementation((_p: string, name: string) =>
      Promise.resolve(name === SUB ? [{ id: 9, change: SUB, angle: "tooling", note: "Guard too strict", disposition: "fixed", ref: "def5678", recordedAt: "2026-01-01T02:40:00Z" }] : []),
    );
    const reviewing = { ...boardRow, currentStage: "flow.self-review", selfReviewTasks: [{ n: 1, title: "Fix the guard", state: "pending" }] };
    fetchStatsViewMock.mockImplementation((view: string) =>
      Promise.resolve(
        view === "state-board"
          ? { ...aggregateEnvelope([]), view, rows: [reviewing] }
          : aggregateEnvelope([
              { projectKey: PROJECT, changeName: CHANGE, command: "/flow", stage: "flow.kickoff", runCount: 3, measuredRuns: 2, totalTokensInput: 1, meanTokensInput: 1, totalCostUsd: 3, totalDurationMs: 1, mainTokens: 1, sidechainTokens: 1 },
            ]),
      ),
    );
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const panel = await screen.findByRole("region", { name: "Self-review" });
    const roles = await within(panel).findAllByTestId("phase-role");
    expect(roles.map((r) => r.textContent)).toEqual(["planner ×1opus ×1 ($1.50)$1.50", "implementer ×1opus ×1 (—)—running", "main session$1.00"]);
    // Every cost counted once: Total = Main + Self-review, and each panel's lines add up to its header.
    const usd = (el: Element | null) => Number((el?.textContent ?? "").replace(/[^\d.]/g, ""));
    const main = screen.getByRole("region", { name: "Main" });
    const lines = (p: HTMLElement) => within(p).getAllByTestId("phase-role").reduce((a, r) => a + usd(r.querySelector(".lb-num")), 0);
    expect(usd(main.querySelector(".lb-phase-cost"))).toBe(3);
    expect(lines(main)).toBeCloseTo(3, 9);
    expect(lines(panel)).toBeCloseTo(usd(panel.querySelector(".lb-phase-cost")), 9);
    const total = screen.getByRole("group", { name: "Total cost" });
    expect(usd(total.querySelector(".lb-kpi-val"))).toBeCloseTo(usd(main.querySelector(".lb-phase-cost")) + usd(panel.querySelector(".lb-phase-cost")), 9);
    expect(total).toHaveTextContent("incl. self-review sub-run $2.50");
    // The phase sums the change's own self-review run (unmeasured) with the sub-run's two.
    expect(panel.querySelector(".lb-phase-cost")).toHaveTextContent("$2.50");
    // A group fixer runs: the first unticked task is the one in progress.
    expect(screen.getByRole("region", { name: "Self-review tasks" })).toHaveTextContent("● 1. Fix the guard");
    expect(within(panel).getByTestId("self-review-now")).toHaveTextContent('Now: Fixing task 1 "Fix the guard"');
    await waitFor(() => expect(within(panel).getByTestId("self-review-summary")).toHaveTextContent("findings 1"));
    expect(within(panel).getByRole("link", { name: "sub-run" })).toHaveAttribute("href", `#/run/agents-tooling/${SUB}`);
    expect(findChangesMock).toHaveBeenCalledWith(SUB);
    // Main's line carries the self-review's cost.
    expect(within(main).getByTestId("main-self-review")).toHaveTextContent(/· \$2\.50$/);
    findChangesMock.mockResolvedValue({ records: [] });
    runRecordMock.mockImplementation(() => Promise.resolve({ dispatches: [] }));
    listFindingsMock.mockImplementation(() => Promise.resolve([]));
  });

  it("run detail's self-review panel says what it waits on before any dispatch", async () => {
    const reviewing = { ...boardRow, currentStage: "flow.self-review" };
    fetchStatsViewMock.mockImplementation((view: string) =>
      Promise.resolve(view === "state-board" ? { ...aggregateEnvelope([]), view, rows: [reviewing] } : aggregateEnvelope([])),
    );
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const panel = await screen.findByRole("region", { name: "Self-review" });
    // Without a plan or a fix round the summary names only what exists.
    expect(within(panel).getByTestId("self-review-summary")).toHaveTextContent(/^findings 0$/);
    expect(within(panel).getByTestId("self-review-now")).toHaveTextContent("Now: Main session is between steps (nothing dispatched yet)");
  });

  it("run detail's collapsed Stage runs shows the latest 5 runs until its toggle shows all", async () => {
    const user = userEvent.setup();
    const stages = ["flow.kickoff", "flow.brainstorm", "flow.decide", "flow.sdd-tdd", "flow.review-panel", "flow.verify", "flow.visual-verify"];
    listStageRunsMock.mockResolvedValue(stageRunsResponse(stages.map((st, i) => run(10 + i, st, `2026-01-01T0${i}:00:00Z`, `2026-01-01T0${i}:30:00Z`))));
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const panel = await screen.findByRole("region", { name: "Stage runs" });
    const rows = () => panel.querySelectorAll("tbody tr:not(.lb-group)");
    await waitFor(() => expect(rows()).toHaveLength(5));
    // The latest five: the two earliest runs are the ones left out.
    expect(within(panel).queryByText("Starting up")).not.toBeInTheDocument();
    expect(within(panel).queryByText("Designing the change")).not.toBeInTheDocument();
    expect(within(panel).getByText("Checking the UI")).toBeInTheDocument();
    await user.click(within(panel).getByRole("button", { name: "Show all 7" }));
    expect(rows()).toHaveLength(7);
    expect(within(panel).getByText("Starting up")).toBeInTheDocument();
    await user.click(within(panel).getByRole("button", { name: "Show latest 5" }));
    expect(rows()).toHaveLength(5);
  });

  it("run detail's Stage runs shows every run and no toggle with 5 or fewer", async () => {
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const panel = await screen.findByRole("region", { name: "Stage runs" });
    await waitFor(() => expect(panel.querySelectorAll("tbody tr:not(.lb-group)")).toHaveLength(4));
    expect(within(panel).queryByRole("button", { name: /^Show (all|latest)/ })).not.toBeInTheDocument();
  });

  it("run detail keeps a self-review that is not running collapsed to its header until opened, its tasks card with it, saying Not started for one yet to run", async () => {
    const user = userEvent.setup();
    const planned = { ...boardRow, selfReviewTasks: [{ n: 1, title: "Fix the guard", state: "pending" }] };
    fetchStatsViewMock.mockImplementation((view: string) => Promise.resolve(view === "state-board" ? { ...aggregateEnvelope([]), view, rows: [planned] } : aggregateEnvelope([])));
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    await screen.findByRole("region", { name: "Main" });
    const work = screen.getAllByRole("region").filter((r) => r.classList.contains("lb-phase")).map((r) => r.getAttribute("aria-label"));
    expect(work).toEqual(["Main", "Self-review"]);
    const sr = screen.getByRole("region", { name: "Self-review" });
    expect(within(sr).getByTestId("self-review-summary")).toHaveTextContent("findings 0");
    expect(within(sr).queryByTestId("self-review-now")).not.toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Main" })).queryByTestId("main-self-review")).not.toBeInTheDocument();
    const srTasks = await screen.findByRole("region", { name: "Self-review tasks" });
    expect(srTasks).toHaveTextContent(/^Self-review tasks0\/1$/);
    await user.click(within(sr).getByRole("button", { name: "Show" }));
    expect(within(sr).getByTestId("self-review-now")).toHaveTextContent(/^Now: Not started$/);
    expect(srTasks).toHaveTextContent("◻ 1. Fix the guard");
    expect(within(screen.getByRole("region", { name: "Main" })).getByTestId("main-now")).toHaveTextContent(/^Now: Implementing tasks · 1h 00m$/);
  });

  it("run detail labels models as family and version, never with a placeholder default effort", async () => {
    runRecordMock.mockResolvedValueOnce({
      dispatches: [
        { id: 1, agentId: "inline", role: "implementer", key: "task-1-implementer", model: "opus", effort: "default", metrics: { modelVersion: "claude-opus-5-5", effort: "medium" }, startedAt: "2026-01-01T01:05:00Z", endedAt: "2026-01-01T01:10:00Z" },
        { id: 2, agentId: "inline", role: "verifier", key: "verify", model: "opus", effort: "default", startedAt: "2026-01-01T01:11:00Z", endedAt: "2026-01-01T01:12:00Z" },
      ],
    });
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const main = await screen.findByRole("region", { name: "Main" });
    await waitFor(() => expect(within(main).getAllByTestId("phase-role").length).toBeGreaterThan(1));
    const text = within(main).getAllByTestId("phase-role").map((r) => r.textContent).join(" | ");
    expect(text).toContain("implementer ×1opus 5.5 · medium");
    expect(text).toContain("verifier ×1opus");
    expect(text).not.toContain("default");
    const verifier = within(main).getAllByTestId("phase-role").find((r) => r.textContent?.startsWith("verifier"))!;
    expect(verifier.querySelector(".lb-role-models")!.getAttribute("title")).toContain("effort unknown");
  });

  it("run detail's Cost by stage lists the ten costliest stages and folds the rest into one line", async () => {
    const stages = ["flow.kickoff", "flow.brainstorm", "flow.design-approval", "flow.create-artifacts", "flow.writing-plans", "flow.decide", "flow.load-context", "flow.isolate-workspace", "flow.sdd-tdd", "flow.review-panel", "flow.verify", "flow.visual-verify"];
    const at = (i: number) => `2026-01-01T00:${String(i * 4).padStart(2, "0")}:00Z`;
    listStageRunsMock.mockResolvedValue(
      stageRunsResponse([
        ...stages.map((st, i) => run(30 + i, st, at(i), at(i), { metrics: { cost_usd: 12 - i } })),
        // A zero-cost and an unmeasured stage: never listed, never in "N more".
        run(50, "flow.stage-diff", at(12), at(12), { metrics: { cost_usd: 0 } }),
        run(51, "flow.run-instructions", at(13), at(13), { metrics: {} }),
      ]),
    );
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const rail = await screen.findByRole("region", { name: "Cost by stage" });
    await waitFor(() => expect(rail.querySelectorAll(".lb-hbar:not([data-testid])")).toHaveLength(10));
    expect(rail.querySelector(".lb-hbar")).toHaveTextContent("Starting up$12.00");
    // The two cheapest, $2 and $1, as one muted line, so the panel still sums to the stages' $78.
    expect(within(rail).getByTestId("stage-cost-rest")).toHaveTextContent(/^2 more\$3\.00$/);
    const sum = [...rail.querySelectorAll(".lb-hbar .lb-num")].reduce((a, e) => a + Number(e.textContent!.replace(/[^\d.]/g, "")), 0);
    expect(sum).toBe(78);
    expect(rail).not.toHaveTextContent("Staging the diff");
    expect(rail).not.toHaveTextContent("Writing test instructions");
  });

  it("run detail counts each role's dispatches per model · effort, the counts summing to the role's and the costs to its cost", async () => {
    const priced = { ...implementation[0], metrics: { ...(implementation[0].metrics as object), dispatches: { r1: { cost_usd: 1 }, r2: { cost_usd: 2 }, r3: { cost_usd: 0.5 } } } };
    listStageRunsMock.mockResolvedValue(stageRunsResponse([...planning, priced]));
    const rev = (id: number, agentId: string, model: string, effort: string) => ({ id, agentId, role: "reviewer", key: `panel-1-r${id}`, model, effort, startedAt: "2026-01-01T01:05:00Z", endedAt: "2026-01-01T01:06:00Z", outcome: "clean" });
    runRecordMock.mockResolvedValueOnce({ dispatches: [rev(1, "r1", "sonnet", "low"), rev(2, "r2", "opus", "low"), rev(3, "r3", "opus", "low"), rev(4, "r4", "", "")] });
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const main = await screen.findByRole("region", { name: "Main" });
    const line = await waitFor(() => within(main).getAllByTestId("phase-role").find((r) => r.textContent?.startsWith("reviewer"))!);
    const models = line.querySelector(".lb-role-models")!.textContent!;
    expect(models).toBe("opus · low ×2 ($2.50), sonnet · low ×1 ($1.00), unknown ×1 (—)");
    const counts = [...models.matchAll(/×(\d+)/g)].reduce((a, m) => a + Number(m[1]), 0);
    expect(line).toHaveTextContent(new RegExp(`^reviewer ×${counts}`));
    const costs = [...models.matchAll(/\(\$([\d.]+)\)/g)].reduce((a, m) => a + Number(m[1]), 0);
    expect(line.querySelector(".lb-num")).toHaveTextContent(`$${costs.toFixed(2)}`);
  });

  it("run detail on a self-review sub-run's own page shows one Main panel of that run's work, its cost split by what each run did", async () => {
    const SUB = `self-review-${CHANGE}`;
    const at = (m: number) => `2026-01-01T02:${String(m).padStart(2, "0")}:00Z`;
    listStageRunsMock.mockResolvedValue(
      stageRunsResponse([
        run(60, "flow.self-review", at(0), at(9), { metrics: { cost_usd: 3, dispatches: { pass: { cost_usd: 2.5 } } } }),
        run(61, "flow.self-review", at(10), at(19), { metrics: { cost_usd: 1, dispatches: { plan: { cost_usd: 0.75 } } } }),
        // A fix round whose run also carries the pass's tail: the tail's cost stays with the round.
        run(62, "flow.self-review", at(20), at(29), { metrics: { cost_usd: 2, dispatches: { fix: { cost_usd: 1.5 }, pass: { cost_usd: 0.1 } } } }),
        run(63, "flow.self-review", at(30), at(39), { metrics: { cost_usd: 0.4, dispatches: { rr: { cost_usd: 0.3 } } } }),
        run(64, "flow.self-review", at(40), at(45), { metrics: { cost_usd: 0.2 } }),
      ]),
    );
    const d = (id: number, agentId: string, role: string, suffix: string, m: number, outcome: string) => ({ id, agentId, role, key: `${SUB}-${suffix}`, model: "opus", startedAt: at(m), endedAt: at(m + 8), outcome });
    runRecordMock.mockImplementation(() =>
      Promise.resolve({ dispatches: [d(1, "pass", "reviewer", "pass", 0, "fix"), d(2, "plan", "planner", "plan", 10, "completed"), d(3, "fix", "implementer", "fix-1", 20, "completed"), d(4, "rr", "reviewer", "review-1", 30, "clean")] }),
    );
    const subBoard = { ...boardRow, name: SUB, state: "FINISHED", currentStage: undefined, subagents: undefined, activity: undefined, tasks: [{ n: 1, title: "Fix the guard", state: "done" }] };
    fetchStatsViewMock.mockImplementation((view: string) => Promise.resolve(view === "state-board" ? { ...aggregateEnvelope([]), view, rows: [subBoard] } : aggregateEnvelope([])));
    findChangesMock.mockImplementation((name: string) => Promise.resolve({ records: name === CHANGE ? [{ projectKey: PROJECT, name: CHANGE }] : [] }));
    listFindingsMock.mockImplementation((_p: string, name: string) =>
      Promise.resolve(name === CHANGE ? [1, 2].map((id) => ({ id, change: CHANGE, angle: "tooling", note: "n", disposition: "fixed", recordedAt: at(50) })) : []),
    );
    render(<RunDetail project={PROJECT} change={SUB} />);
    const main = await screen.findByRole("region", { name: "Main" });
    const work = screen.getAllByRole("region").filter((r) => r.classList.contains("lb-phase")).map((r) => r.getAttribute("aria-label"));
    expect(work).toEqual(["Main"]);
    expect(findChangesMock).not.toHaveBeenCalledWith(`self-review-${SUB}`);
    // Its findings are the reviewed change's.
    expect(findChangesMock).toHaveBeenCalledWith(CHANGE);
    await waitFor(() => expect(within(main).getByTestId("main-summary")).toHaveTextContent(/^plan 1\/1 · findings 2 · fixes round 1$/));
    expect(within(main).getByTestId("main-now")).toHaveTextContent(/^Now: Done$/);
    expect(within(main).getByTestId("main-progress")).toHaveTextContent("analyze ✓ → plan ✓ → fix 1/1 ✓ → re-review ✓ → verify ✓ → land ✓ → file ✓");
    expect(screen.getByRole("region", { name: "Tasks" })).toHaveTextContent("✔ 1. Fix the guard");
    const rail = screen.getByRole("region", { name: "Cost by stage" });
    await waitFor(() => expect(rail).toHaveTextContent("Fix rounds"));
    const bars = [...rail.querySelectorAll(".lb-hbar")].map((b) => b.textContent);
    expect(bars).toEqual(["Analysis pass$3.00", "Fix rounds$2.00", "Planning fixes$1.00", "Review rounds$0.40", "Landing and filing$0.20"]);
    expect(rail).not.toHaveTextContent(/^Self-review/);
    // The analysis pass is the analyzer; reviewer stays the re-review's.
    expect(within(main).getAllByTestId("phase-role").map((r) => r.textContent?.split(" ")[0])).toEqual(expect.arrayContaining(["analyzer", "reviewer"]));
    runRecordMock.mockImplementation(() => Promise.resolve({ dispatches: [] }));
    findChangesMock.mockResolvedValue({ records: [] });
    listFindingsMock.mockImplementation(() => Promise.resolve([]));
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

  it("run detail reads main as waiting, not ended, when its turn ended while a subagent runs", async () => {
    const ended = { ...boardRow, activity: "idle", subagents: boardRow.subagents.filter((d) => !d.inline) };
    fetchStatsViewMock.mockImplementation((view: string) =>
      Promise.resolve(view === "state-board" ? { ...aggregateEnvelope([]), view, rows: [ended] } : aggregateEnvelope([])),
    );
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const main = (await screen.findByRole("cell", { name: "Main" })).closest("tr")!;
    expect(main.querySelector(".lb-dot.waiting")).not.toBeNull();
  });

  it("run detail reads main as running when it wrote its transcript after a refused stop", async () => {
    const refused = { ...boardRow, activity: "idle", activitySince: "2026-01-01T02:58:00Z", sessionSeenAt: "2026-01-01T02:59:30Z", subagents: [] };
    fetchStatsViewMock.mockImplementation((view: string) =>
      Promise.resolve(view === "state-board" ? { ...aggregateEnvelope([]), view, rows: [refused] } : aggregateEnvelope([])),
    );
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const main = (await screen.findByRole("cell", { name: "Main" })).closest("tr")!;
    expect(main.querySelector(".lb-dot.running")).not.toBeNull();
  });

  it("run detail still shows its stage runs when the board row fails to load", async () => {
    fetchStatsViewMock.mockImplementation((view: string) =>
      view === "state-board" ? Promise.reject(new Error("board down")) : Promise.resolve(aggregateEnvelope([])),
    );
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    expect(await screen.findByRole("cell", { name: "Starting up" })).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("stage timeline starts collapsed, its header toggle showing and hiding it", async () => {
    const user = userEvent.setup();
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    const panel = await screen.findByRole("region", { name: "Timeline" });
    expect(within(panel).queryByRole("img", { name: "Stage runs over time" })).not.toBeInTheDocument();
    await user.click(within(panel).getByRole("button", { name: "Show timeline" }));
    expect(within(panel).getByRole("img", { name: "Stage runs over time" })).toBeInTheDocument();
    await user.click(within(panel).getByRole("button", { name: "Hide timeline" }));
    expect(within(panel).queryByRole("img", { name: "Stage runs over time" })).not.toBeInTheDocument();
  });

  it("stage timeline draws a lane per stage coloured by outcome", async () => {
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    await userEvent.click(await screen.findByRole("button", { name: "Show timeline" }));
    const svg = await screen.findByRole("img", { name: "Stage runs over time" });
    // The stages still to run follow as muted lanes, in the command's order; a skipped one (decide) is not "upcoming".
    await waitFor(() => expect(svg.querySelectorAll("[data-testid=upcoming-lane]").length).toBe(2));
    expect([...svg.querySelectorAll("[data-testid=upcoming-lane]")].map((t) => t.lastChild?.textContent)).toEqual(["Code review", "Running lint and tests"]);
    // Each lane reads in plain words, its stage id on hover.
    const lanes = [...svg.querySelectorAll("text.lb-lane:not(.upcoming)")];
    expect(lanes.map((t) => t.lastChild?.textContent)).toEqual(["Starting up", "Designing the change", "Implementing tasks"]);
    expect(lanes.map((t) => t.querySelector("title")?.textContent)).toEqual(["flow.kickoff", "flow.brainstorm", "flow.sdd-tdd"]);
    const fills = [...svg.querySelectorAll("rect.lb-bar")].map((r) => r.getAttribute("fill"));
    expect(fills).toEqual(["var(--status-ok)", "var(--status-ok)", "var(--status-critical)", "url(#lb-hatch)"]);
    expect(svg.querySelector("line.lb-now")).not.toBeNull();
    expect(svg).toHaveAttribute("width", "100%");
  });
});

describe("the run view stays live", () => {
  afterEach(() => vi.useRealTimers());

  it("refetches every five seconds and shows a stage run that appears", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const run = (id: number, stage: string): StageRunDTO => ({
      stageRunId: id,
      harness: "claude-code",
      command: "/flow",
      stage,
      attempt: 1,
      startedAt: "2026-01-01T00:00:00Z",
      metrics: {},
    });
    listStageRunsMock.mockResolvedValue(stageRunsResponse([run(1, "SDD + TDD per task")]));
    fetchStatsViewMock.mockResolvedValue(aggregateEnvelope([]));
    render(<RunDetail project={PROJECT} change={CHANGE} />);
    await screen.findAllByRole("cell", { name: "SDD + TDD per task" });

    listStageRunsMock.mockResolvedValue(stageRunsResponse([run(1, "SDD + TDD per task"), run(2, "Review panel")]));
    await vi.advanceTimersByTimeAsync(5000);

    await screen.findAllByRole("cell", { name: "Review panel" });
    expect(screen.getAllByRole("cell", { name: "SDD + TDD per task" }).length).toBeGreaterThan(0);
  });
});
