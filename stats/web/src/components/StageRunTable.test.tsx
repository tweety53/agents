// A stage run's row names what its dispatches did; its dispatch rows read the
// dispatch's plain words, real outcome and timing, never a raw key or agent
// id, and the exact model as family and version; a main-session row carries
// the stage's own share so the rows add up to the stage's cost; the stage key
// and the auto-recorded note live in the row's tooltip only.
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

describe("StageRunTable", () => {
  it("names a collapsed run by what its dispatches did, keeping the stage key and auto-recorded note to the tooltip", () => {
    render(<StageRunTable stageRuns={[run]} records={records} />);
    const cell = screen.getByRole("cell", { name: /^▸/ });
    expect(cell).toHaveTextContent(/^▸Self-review · fixing, round 2 \(opus\) \+ reviewing fixes, round 2 \(sonnet\)$/);
    expect(cell.title).toContain("flow.self-review");
    expect(cell.title).toContain("auto-recorded");
    expect(screen.queryByText(/auto-recorded/)).not.toBeInTheDocument();
    expect(screen.queryByText("flow.self-review")).not.toBeInTheDocument();
  });

  it("reads dispatch rows in plain words with their real outcome and timing, agent id only in a tooltip", async () => {
    render(<StageRunTable stageRuns={[run]} records={records} />);
    await userEvent.click(screen.getByRole("button", { name: "Show dispatches" }));
    const rows = screen.getAllByTestId("dispatch-row");
    expect(rows[0]).toHaveTextContent(/^Self-review: fixing, round 2completed00:0110m4kopus 5\.5\$2\.00$/);
    expect(rows[1]).toHaveTextContent(/^Self-review: reviewing fixes, round 2running00:12/);
    expect(screen.queryByText(/a548eaf33f9b0cb85/)).not.toBeInTheDocument();
    expect(screen.queryByText(new RegExp(`${KEY}-fix-2`))).not.toBeInTheDocument();
    expect(within(rows[0]).getByRole("cell", { name: "Self-review: fixing, round 2" }).title).toContain("a548eaf33f9b0cb85");
    // main's share: the stage's cost less its dispatches', as a muted one-liner.
    const main = screen.getByTestId("main-share-row");
    expect(main).toHaveTextContent(/^main sessionopus 5\.5 · low\$0\.85$/);
  });
});
