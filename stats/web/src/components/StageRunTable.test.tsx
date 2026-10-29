// A stage run's dispatch rows read their exact model version, a main-session
// row carries the stage's own share so the rows add up to the stage's cost,
// and an auto-opened run says it was recorded automatically.
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { StageRunTable } from "./StageRunTable";

const run = {
  stageRunId: 1,
  harness: "claude-code",
  command: "/flow",
  stage: "flow.visual-verify",
  attempt: 1,
  startedAt: "2026-01-01T00:00:00Z",
  endedAt: "2026-01-01T00:26:00Z",
  outcome: "stopped",
  autoOpened: true,
  metrics: {
    cost_usd: 3,
    effort: "low",
    models: { "claude-opus-5-5": { tokens: { main: { input: 5 }, sidechain: { input: 9 } } } },
    dispatches: { a1: { model: "opus", modelVersion: "claude-opus-5-5", cost_usd: 2.15, description: "visual-verify-fix-1 verification" } },
  },
};

describe("StageRunTable", () => {
  it("shows the dispatch's model version, main's share and the auto-recorded note", async () => {
    render(<StageRunTable stageRuns={[run]} />);
    expect(screen.getByText("(auto-recorded)")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button"));
    expect(screen.getByTestId("dispatch-row")).toHaveTextContent("opus-5.5");
    const main = screen.getByTestId("main-share-row");
    expect(main).toHaveTextContent("main session");
    expect(main).toHaveTextContent("opus-5.5 · low");
    expect(main).toHaveTextContent("$0.85");
  });
});
