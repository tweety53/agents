// A stage run's row names what its own dispatches did, each with its model and
// effort, and leads with its stage only when its session runs more than one;
// its outcome is the last dispatch's in plain words; its cost splits agents +
// main session on hover; two or more own dispatches open beneath it in time
// order, one is the row itself, and a dispatch's tail from an earlier run is
// never a row of its own.
import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { StageRunTable, foldPrefixes } from "./StageRunTable";
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
  it("names a run by its dispatches' actions, its agents' models in their own column, no stage lead when its session runs that stage alone, and no jargon tooltip", () => {
    render(<StageRunTable stageRuns={[run]} records={records} />);
    const cell = screen.getByRole("button", { name: "Show dispatches" }).closest("td")!;
    expect(cell).toHaveTextContent(/^Fix round 2 → Review round 2$/);
    expect(screen.getByRole("columnheader", { name: "Model · effort" })).toBeInTheDocument();
    expect(screen.getByTestId("run-model")).toHaveTextContent(/^opus 5\.5, sonnet 5\.5$/);
    expect(cell.title).not.toContain("auto-recorded");
    expect(cell.closest("tr")).toHaveTextContent(/running/);
  });

  it("leads with the stage when its session runs more than one", () => {
    render(<StageRunTable stageRuns={[run, other]} records={records} />);
    expect(screen.getByRole("button", { name: "Show dispatches" }).closest("td")!).toHaveTextContent(/^Self-review · Fix round 2 → Review round 2$/);
  });

  it("splits the run's cost into agents and main session on hover", () => {
    render(<StageRunTable stageRuns={[run]} records={records} />);
    const cost = screen.getByTestId("run-cost");
    expect(cost).toHaveTextContent(/^\$3\.00$/);
    expect(cost.title).toBe("agents $2.15 + main session $0.85");
  });

  it("lists dispatch rows in time order, by their description with their real outcome and timing, agent id only in a tooltip, and no main-session row", async () => {
    render(<StageRunTable stageRuns={[run]} records={records} />);
    await userEvent.click(screen.getByRole("button", { name: "Show dispatches" }));
    const rows = screen.getAllByTestId("dispatch-row");
    expect(rows[0]).toHaveTextContent(/^⎿ self-review-kan-9-x-fix-2 minor fixesfixed00:0110m4kopus 5\.5\$2\.00$/);
    expect(rows[1]).toHaveTextContent(/^⎿ self-review-kan-9-x-review-2 re-reviewrunning00:12/);
    expect(screen.queryByText(/a548eaf33f9b0cb85/)).not.toBeInTheDocument();
    expect(within(rows[0]).getByRole("cell", { name: "self-review-kan-9-x-fix-2 minor fixes" }).title).toBe("agent a548eaf33f9b0cb85");
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
    expect(within(row).getAllByRole("cell")[0]).toHaveTextContent(/^Review round 2$/);
    expect(within(row).getByTestId("run-model")).toHaveTextContent(/^sonnet 5\.5$/);
    expect(row).toHaveTextContent("clean");
    expect(screen.getByTestId("run-cost").title).toBe("agents $2.15 + main session $0.85");
  });

  it("names a dispatch's exact model from its served models when the bucket records no version", () => {
    const served = { ...run, metrics: { ...(run.metrics as object), dispatches: { aeb9c995f7b6dd74f: { model: "sonnet", cost_usd: 0.15, signals: { served_models: { "claude-sonnet-5-5": 6 } } } } } };
    render(<StageRunTable stageRuns={[served]} records={records} />);
    expect(screen.getByTestId("run-model")).toHaveTextContent(/^sonnet 5\.5$/);
  });

  it("counts several agents' models in action order, and shows the main session's model muted on a row with no agent", () => {
    const rec = (id: number, agentId: string, model: string, at: string) => ({ id, agentId, key: `panel-1-s${id}`, role: "reviewer", model, startedAt: at, endedAt: at });
    const panel: StageRunDTO = {
      ...run,
      stage: "flow.review-panel",
      metrics: {
        cost_usd: 1,
        dispatches: Object.fromEntries(
          [["p1", "claude-opus-5-5"], ["p2", "claude-sonnet-5-5"], ["p3", "claude-opus-5-5"], ["p4", "claude-opus-5-5"]].map(([id, m]) => [id, { modelVersion: m, cost_usd: 0.1, signals: { served_efforts: { low: 1 } } }]),
        ),
      },
    };
    const panelRecs: RecordDispatch[] = [rec(1, "p1", "opus", "2026-01-01T00:01:00Z"), rec(2, "p2", "sonnet", "2026-01-01T00:02:00Z"), rec(3, "p3", "opus", "2026-01-01T00:03:00Z"), rec(4, "p4", "opus", "2026-01-01T00:04:00Z")];
    const bare: StageRunDTO = { ...run, stageRunId: 9, stage: "flow.verify", startedAt: "2026-01-01T02:00:00Z", metrics: { effort: "low", models: { "claude-opus-5-5": { tokens: { main: { input: 5 } } } } } };
    render(<StageRunTable stageRuns={[panel, bare]} records={panelRecs} />);
    const [agents, main] = screen.getAllByTestId("run-model");
    expect(agents).toHaveTextContent(/^opus 5\.5 · low ×3, sonnet 5\.5 · low$/);
    const muted = within(main).getByText("opus 5.5 · low");
    expect(muted).toHaveClass("lb-muted");
    expect(muted).toHaveAttribute("title", "main session");
  });

  it("opens a run's dispatches as one attached group, says their shared prefix once, and keeps them until their closing animation ends", async () => {
    // No description: the row falls back to its action name.
    const shared = { ...run, stage: "flow.review-panel", metrics: { dispatches: { a548eaf33f9b0cb85: {}, aeb9c995f7b6dd74f: {} } } };
    const recs: RecordDispatch[] = [
      { ...records[0], key: "panel-2-primary" },
      { ...records[1], key: "panel-2-mutation" },
    ];
    render(<StageRunTable stageRuns={[shared]} records={recs} />);
    const toggle = screen.getByRole("button", { name: "Show dispatches" });
    await userEvent.click(toggle);
    expect(toggle.closest("tr")).toHaveClass("open");
    const rows = screen.getAllByTestId("dispatch-row");
    expect(rows.map((r) => r.querySelector("td")!.textContent)).toEqual(["⎿ Code review round 2: primary", "⎿ mutation"]);
    expect(rows[1]).toHaveClass("last");
    await userEvent.click(screen.getByRole("button", { name: "Hide dispatches" }));
    // Closing: still mounted, sliding out, until the animation ends.
    const closing = document.querySelectorAll("tr.lb-dispatch.out");
    expect(closing).toHaveLength(2);
    expect(screen.queryByTestId("dispatch-row")).not.toBeInTheDocument();
    // jsdom's styles carry no "animation" property, so React listens for the prefixed name there; browsers fire animationend.
    fireEvent(closing[0].querySelector(".lb-cell")!, new Event("webkitAnimationEnd", { bubbles: true }));
    expect(document.querySelectorAll("tr.lb-dispatch")).toHaveLength(0);
  });

  it("says a shared action prefix once", () => {
    expect(foldPrefixes(["Code review round 2: primary", "Code review round 2: test sanity", "Running lint and tests", "Code review round 3: primary"])).toEqual([
      "Code review round 2: primary",
      "test sanity",
      "Running lint and tests",
      "Code review round 3: primary",
    ]);
  });

  it("names a dispatch with no record by the key its description leads with", () => {
    render(<StageRunTable stageRuns={[run]} />);
    expect(screen.getByRole("button", { name: "Show dispatches" }).closest("td")!).toHaveTextContent(/^Fix round 2 → Review round 2$/);
  });

  it("shows the row's total tokens in, split agents + main session on hover", () => {
    const counted = { ...run, metrics: { ...(run.metrics as object), tokens: { main: { input: 1000, cache_read: 9000 }, sidechain: { input: 4000 } } } };
    render(<StageRunTable stageRuns={[counted]} records={records} />);
    const tokens = screen.getByTestId("run-tokens");
    expect(tokens).toHaveTextContent(/^14k$/);
    expect(tokens.title).toBe("agents 4k + main session 10k");
  });
});

describe("StageRunTable request column", () => {
  const r = (id: number, stage: string, start: string, request?: number): StageRunDTO => ({
    stageRunId: id, harness: "claude-code", command: "/flow-interactive", stage, attempt: 1, sessionId: `s-${id % 2}`,
    startedAt: start, endedAt: start, outcome: "completed", request,
  });
  const runs = [r(1, "flow.verify", "2026-01-01T10:49:00Z", 6), r(2, "flow.load-context", "2026-01-01T11:01:00Z", 7), r(3, "flow.self-review", "2026-01-01T11:30:00Z")];

  it("request column leads each row with R<n>, newest first and ungrouped, when switched on", () => {
    const { container } = render(<StageRunTable stageRuns={runs} requests />);
    expect(screen.getAllByRole("columnheader")[0]).toHaveTextContent("Request");
    const rows = [...container.querySelectorAll("tbody tr")];
    expect(rows.map((row) => row.querySelector("td")?.textContent)).toEqual(["", "R7", "R6"]);
    expect(rows[1]).toHaveTextContent("Loading context");
    expect(container.querySelector(".lb-group")).toBeNull();
  });
  it("request column stays off by default", () => {
    const { container } = render(<StageRunTable stageRuns={runs} />);
    expect(screen.getAllByRole("columnheader")[0]).toHaveTextContent("Stage");
    expect(container.querySelector(".lb-group")).not.toBeNull();
    expect(container).not.toHaveTextContent("R7");
  });
  it("request column starts every Stage cell with its text, an expander trailing it", () => {
    render(<StageRunTable stageRuns={[run, { ...other, request: 2 }]} records={records} requests />);
    const cells = screen.getAllByRole("row").slice(1).map((row) => row.querySelectorAll("td")[1]);
    for (const td of cells) expect(td.firstChild?.nodeType).toBe(Node.TEXT_NODE);
    expect(screen.getByRole("button", { name: "Show dispatches" }).previousSibling?.nodeType).toBe(Node.TEXT_NODE);
  });
  it("request column widens the empty-table row", () => {
    const { container } = render(<StageRunTable stageRuns={[]} requests />);
    expect(container.querySelector("tbody td")).toHaveAttribute("colspan", "8");
  });
});
