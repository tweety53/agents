import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { Runs } from "./Runs";
import type { ChangeRuns, StatsResponse } from "../api";

// The same mocking shape views.test.tsx uses: swap fetchStatsView for a
// vi.fn() and hand it a complete envelope.
const { fetchStatsViewMock } = vi.hoisted(() => ({ fetchStatsViewMock: vi.fn() }));
vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return { ...actual, fetchStatsView: fetchStatsViewMock };
});

const period = { from: new Date("2026-09-01T00:00:00Z"), to: new Date("2026-09-02T00:00:00Z") };

function envelope(rows: ChangeRuns[]): StatsResponse<ChangeRuns[]> {
  return {
    view: "runs",
    from: period.from.toISOString(),
    to: period.to.toISOString(),
    boundaryConvention: "a stage run is attributed to the period containing its start instant",
    recorded: true,
    unmeasured: false,
    rows,
  };
}

const totals = (over: Partial<import("../api").RunTotals> = {}) => ({
  inputTokens: 0, outputTokens: 0, thinkingTokens: 0, cacheRead: 0, cacheWrite5m: 0, cacheWrite1h: 0,
  cacheHitRatio: null, costUsd: null, wallClockMs: 0, humanGateMs: 0, compactions: 0, turns: 0,
  toolCalls: 0, toolErrors: 0, denials: 0, apiErrors: 0, contextEnd: null, ...over,
});

const rows: ChangeRuns[] = [{
  project: "p", change: "kan-1-x", jiraKey: "KAN-1", totals: totals({ inputTokens: 600 }), idleBetweenRunsMs: 0, fixIterations: 0,
  runs: [{
    sessionToken: "mf-1", kind: "flow", command: "/flow", startedAt: "2026-09-01T10:00:00Z", endedAt: "2026-09-01T11:00:00Z",
    totals: totals({ inputTokens: 600, costUsd: 2 }), main: totals({ inputTokens: 200, costUsd: 0.75 }),
    decision: { execution: "subagent", implementer: "claude-opus-5/medium", panel: "1 slot, compact" },
    fanOutMax: 1, suiteRuns: 2, suiteFirstPass: false,
    dispatches: [{
      seq: 1, role: "implementer", slot: "", agentId: "a1", agentType: "flow-medium", description: "impl", depth: 1,
      declaredModel: "claude-opus-5", declaredEffort: "medium", servedModels: { "claude-sonnet-5": 5 }, servedEfforts: { medium: 5 },
      mismatch: true, findingsRaised: 3, findingsByStatus: { fixed: 2, open: 1 }, startedAt: "2026-09-01T10:10:00Z", endedAt: "2026-09-01T10:40:00Z", totals: totals({ inputTokens: 400, costUsd: 1.25 }),
    }],
  }],
}];

describe("Runs", () => {
  it("shows one row per run and expands to main plus dispatches", async () => {
    fetchStatsViewMock.mockResolvedValue(envelope(rows));
    render(<Runs period={period} project={undefined} />);
    expect(await screen.findByText("kan-1-x")).toBeInTheDocument();
    expect(screen.getByText("subagent")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /run details/i }));
    expect(screen.getByText("main session")).toBeInTheDocument();
    expect(screen.getByText("claude-opus-5 (served claude-sonnet-5)")).toBeInTheDocument();
    expect(screen.getByTestId("dispatch-mismatch")).toBeInTheDocument();
    expect(screen.getByText("3 (fixed 2, open 1)")).toBeInTheDocument();
  });

  it("gives two token-less runs sharing a startedAt distinct row keys, so expanding one opens only its own detail", async () => {
    const tokenLessRuns: ChangeRuns[] = [{
      project: "p", change: "kan-2-y", jiraKey: "KAN-2", totals: totals(), idleBetweenRunsMs: 0, fixIterations: 0,
      runs: [
        {
          sessionToken: "", kind: "flow-fast", command: "/flow-fast", startedAt: "2026-09-01T10:00:00Z", endedAt: "2026-09-01T10:30:00Z",
          totals: totals(), main: totals(), decision: null, fanOutMax: 0, suiteRuns: 0, suiteFirstPass: null, dispatches: [],
        },
        {
          sessionToken: "", kind: "flow-fast", command: "/flow-fast", startedAt: "2026-09-01T10:00:00Z", endedAt: "2026-09-01T11:00:00Z",
          totals: totals(), main: totals(), decision: null, fanOutMax: 0, suiteRuns: 0, suiteFirstPass: null, dispatches: [],
        },
      ],
    }];
    fetchStatsViewMock.mockResolvedValue(envelope(tokenLessRuns));
    render(<Runs period={period} project={undefined} />);

    const toggles = await screen.findAllByRole("button", { name: /run details/i });
    expect(toggles).toHaveLength(2);

    await userEvent.click(toggles[0]);
    expect(screen.getAllByTestId("main-session-row")).toHaveLength(1);
  });
});
