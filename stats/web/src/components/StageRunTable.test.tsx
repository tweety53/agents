// A stage run's row names what its own dispatches did, each with its model and
// effort, and leads with its stage only when its session runs more than one;
// its outcome is the last dispatch's in plain words; its cost splits agents +
// main session on hover; two or more own dispatches open beneath it in time
// order, one is the row itself, and a dispatch's tail from an earlier run is
// never a row of its own.
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { StageRunTable } from "./StageRunTable";
import type { RecordDispatch, StageRunDTO } from "../api";

const KEY = "self-review-kan-9-x";
const run: StageRunDTO = {
  stageRunId: 1,
  harness: "claude-code",
  command: "/flow",
  stage: "flow.self-review",
  attempt: 1,
  startedAt: "2026-01-01T00:00:00Z",
  endedAt: "2026-01-01T00:26:00Z",
  outcome: "completed",
  autoOpened: true,
  metrics: {
    cost_usd: 3,
    effort: "low",
    models: { "claude-opus-5-5": { tokens: { main: { input: 5 }, sidechain: { input: 9 } } } },
    dispatches: {
      a548eaf33f9b0cb85: { model: "opus", modelVersion: "claude-opus-5-5", cost_usd: 2, description: `${KEY}-fix-2 minor fixes`, tokens: { main: { input: 4000 } } },
      aeb9c995f7b6dd74f: { model: "sonnet", modelVersion: "claude-sonnet-5-5", cost_usd: 0.15, description: `${KEY}-review-2 re-review` },
    },
  },
};
const records: RecordDispatch[] = [
  { id: 1, agentId: "a548eaf33f9b0cb85", key: `${KEY}-fix-2`, role: "implementer", model: "opus", outcome: "completed", startedAt: "2026-01-01T00:01:00Z", endedAt: "2026-01-01T00:11:00Z" },
  { id: 2, agentId: "aeb9c995f7b6dd74f", key: `${KEY}-review-2`, role: "reviewer", model: "sonnet", startedAt: "2026-01-01T00:12:00Z" },
];

const other: StageRunDTO = { ...run, stageRunId: 2, stage: "flow.verify", metrics: undefined, startedAt: "2026-01-01T01:00:00Z", endedAt: "2026-01-01T01:01:00Z" };

describe("StageRunTable", () => {
  it("names a run by its dispatches' actions with their models, no stage lead when its session runs that stage alone, and no jargon tooltip", () => {
    render(<StageRunTable stageRuns={[run]} records={records} />);
    const cell = screen.getByRole("cell", { name: /^▸/ });
    expect(cell).toHaveTextContent(/^▸Fix round 2 \(opus 5\.5\) → Review round 2 \(sonnet 5\.5\)$/);
    expect(cell.title).not.toContain("auto-recorded");
    expect(cell.closest("tr")).toHaveTextContent(/running/);
  });

  it("leads with the stage when its session runs more than one", () => {
    render(<StageRunTable stageRuns={[run, other]} records={records} />);
    expect(screen.getByRole("cell", { name: /^▸/ })).toHaveTextContent(/^▸Self-review · Fix round 2 \(opus 5\.5\) → Review round 2/);
  });

  it("splits the run's cost into agents and main session on hover", () => {
    render(<StageRunTable stageRuns={[run]} records={records} />);
    const cost = screen.getByTestId("run-cost");
    expect(cost).toHaveTextContent(/^\$3\.00$/);
    expect(cost.title).toBe("agents $2.15 + main session $0.85");
  });

  it("lists dispatch rows in time order, in plain words with their real outcome and timing, agent id only in a tooltip, and no main-session row", async () => {
    render(<StageRunTable stageRuns={[run]} records={records} />);
    await userEvent.click(screen.getByRole("button", { name: "Show dispatches" }));
    const rows = screen.getAllByTestId("dispatch-row");
    expect(rows[0]).toHaveTextContent(/^Fix round 2fixed00:0110m4k\$2\.00$/);
    expect(rows[1]).toHaveTextContent(/^Review round 2running00:12/);
    expect(screen.queryByText(/a548eaf33f9b0cb85/)).not.toBeInTheDocument();
    expect(within(rows[0]).getByRole("cell", { name: "Fix round 2" }).title).toContain("a548eaf33f9b0cb85");
    expect(screen.queryByText("main session")).not.toBeInTheDocument();
  });

  it("shows a single own dispatch in the row itself, an earlier run's tail neither named nor listed, and the review's clean as the outcome", () => {
    const tailed: RecordDispatch[] = [
      { ...records[0], startedAt: "2025-12-31T23:50:00Z", outcome: "completed" },
      { ...records[1], outcome: "clean", endedAt: "2026-01-01T00:20:00Z" },
    ];
    render(<StageRunTable stageRuns={[run]} records={tailed} />);
    expect(screen.queryByRole("button", { name: "Show dispatches" })).not.toBeInTheDocument();
    const row = screen.getByRole("row", { name: /Review round 2/ });
    expect(within(row).getAllByRole("cell")[0]).toHaveTextContent(/^Review round 2 \(sonnet 5\.5\)$/);
    expect(row).toHaveTextContent("clean");
    expect(screen.getByTestId("run-cost").title).toBe("agents $2.15 + main session $0.85");
  });

  it("names a dispatch with no record by the key its description leads with", () => {
    render(<StageRunTable stageRuns={[run]} />);
    expect(screen.getByRole("cell", { name: /^▸/ })).toHaveTextContent("Fix round 2 (opus 5.5) → Review round 2 (sonnet 5.5)");
  });

  it("shows the row's total tokens in, split agents + main session on hover, and no model column", () => {
    const counted = { ...run, metrics: { ...(run.metrics as object), tokens: { main: { input: 1000, cache_read: 9000 }, sidechain: { input: 4000 } } } };
    render(<StageRunTable stageRuns={[counted]} records={records} />);
    const tokens = screen.getByTestId("run-tokens");
    expect(tokens).toHaveTextContent(/^14k$/);
    expect(tokens.title).toBe("agents 4k + main session 10k");
    expect(screen.queryByRole("columnheader", { name: /model/i })).not.toBeInTheDocument();
  });
});
