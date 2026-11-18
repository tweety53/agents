// RunCosts is the across-runs Run costs view (KAN-1132, the prototype's
// view C), drawn from GET /api/v1/stats/run-costs with Compare joining two
// runs' one-run results. fetchStatsView, fetchProjects and getRunTimings are
// mocked at the module boundary, so these tests assert the view's own drawing.
import { act, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { RunCostsResult, RunMetric, RunTimings as RunTimingsDTO } from "../api";
import { RunCosts } from "./RunCosts";

const { fetchStatsViewMock, fetchProjectsMock, runTimingsMock } = vi.hoisted(() => ({
  fetchStatsViewMock: vi.fn(),
  fetchProjectsMock: vi.fn(),
  runTimingsMock: vi.fn(),
}));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return { ...actual, fetchStatsView: fetchStatsViewMock, fetchProjects: fetchProjectsMock, getRunTimings: runTimingsMock };
});

const H = 3_600_000;
const NOW = Date.parse("2026-11-19T12:00:00Z");
const P = "gymie-7c1f238a";
const Q = "agents-0a1b2c3d";
const C1 = "kan-1091-body-calories-measurements";
const C2 = "kan-1088-weekly-volume-chart";
const C3 = "kan-1120-panel-fix-single-dispatch";
const m = <T,>(value: T, kind: RunMetric<T>["kind"] = "derived"): RunMetric<T> => ({ value, definition: "d", kind });

const DEFS = [
  { name: "Wall clock", definition: "First stage begin to last stage end; now while a stage is open.", kind: "derived" as const },
  { name: "Agent time", definition: "Union of dispatch rows' [begin, end).", kind: "derived" as const },
  { name: "Parent commands", definition: "Recorded parent-side command runs.", kind: "measured" as const },
  { name: "Idle / parent", definition: "Wall clock minus the union of dispatches and parent commands.", kind: "derived" as const },
  { name: "Cost", definition: "Priced cost_usd from harvested tokens.", kind: "measured" as const },
  { name: "Apportioned cost", definition: "A row whose tokens were shared and split by time.", kind: "estimated" as const },
  { name: "Category", definition: "Read from the dispatch key.", kind: "derived" as const },
];

function run(project: string, change: string, wallH: number, agentH: number, cost: number, tasks: number, verify: number) {
  return {
    project,
    change,
    wallClockMs: m(wallH * H),
    agentTimeMs: m(agentH * H),
    idleMs: m((wallH - agentH) * H),
    cost: m(cost, "measured"),
    tasks: m(tasks),
    fixRounds: m({ verify, panel: 0 }),
    categoryCost: {},
  };
}

function costs(): RunCostsResult {
  return {
    columns: [],
    categories: [
      { category: "planned implementer", runs: 3, rows: 93, seconds: 6.9 * 3600, cost: 95.47, shareOfCost: 0.274 },
      { category: "gated fixer", runs: 2, rows: 21, seconds: 2.41 * 3600, cost: 36.83, shareOfCost: 0.106 },
    ],
    runs: [run(P, C1, 49.35, 30.68, 348.51, 61, 13), run(P, C2, 22.1, 15.02, 141.2, 28, 6), run(Q, C3, 3.2, 2.05, 18.44, 6, 1)],
  };
}

// timings builds one run's result: wall/agent hours, cost, a task count and
// its category costs ("parent commands" and "idle / parent" are not categories).
function timings(change: string, wallH: number, agentH: number, cost: number, tasks: number, verify: number, cats: Record<string, number>): RunTimingsDTO {
  const row = (label: string, c: number) => ({ label, rows: 1, shareMs: H, busyMs: H, cost: c, noTokens: 0 });
  const rows = [...Object.entries(cats).map(([k, v]) => row(k, v)), row("parent commands", 0), row("idle / parent", 1)];
  const bd = { columns: [], rows, total: row("Total", cost) };
  return {
    project: P,
    change,
    from: new Date(NOW - wallH * H).toISOString(),
    to: new Date(NOW).toISOString(),
    open: false,
    totals: {
      wallClockMs: m(wallH * H),
      agentTimeMs: m(agentH * H),
      parentCommandsMs: m(0, "measured"),
      idleMs: m((wallH - agentH) * H),
      cost: m(cost, "measured"),
      parentSessionCost: m(0, "measured"),
      visualVerifyShare: m(0),
      dispatches: 1,
      noTokens: 0,
      parentCommandRuns: 0,
      unclaimedGuardRuns: 0,
    },
    lanes: [],
    breakdowns: { category: bd, role: bd, model: bd, repo: bd, round: bd },
    costPerTask: Array.from({ length: tasks }, (_, i) => ({ task: String(i + 1), rows: 1, shareMs: H, cost: 1, estimated: false })),
    costPerVerifyRound: [],
    counts: {
      fixRounds: m({ verify, panel: 0 }),
      departures: m([]),
      findings: m({ critical: 0, important: 0, minor: 0 }),
      flakyRetries: m({ retries: 0, runs: 0 }),
      rows: m({ open: 0, aborted: 0, blocked: 0 }),
      journalPending: m(0),
    },
    dataQuality: [],
    definitions: DEFS,
  };
}

const T: Record<string, RunTimingsDTO> = {
  [C1]: timings(C1, 49.35, 30.68, 348.51, 61, 13, { "planned implementer": 0, "gated fixer": 36.83 }),
  [C2]: timings(C2, 22.1, 15.02, 141.2, 28, 6, { "planned implementer": 40, "gated fixer": 10 }),
  [C3]: timings(C3, 3.2, 2.05, 18.44, 6, 1, { "planned implementer": 12 }),
};

beforeEach(() => {
  vi.useFakeTimers({ now: NOW, toFake: ["Date"] });
  fetchStatsViewMock.mockReset();
  fetchStatsViewMock.mockResolvedValue(costs());
  fetchProjectsMock.mockReset();
  fetchProjectsMock.mockResolvedValue([Q, P]);
  runTimingsMock.mockReset();
  runTimingsMock.mockImplementation((_p: string, c: string) => Promise.resolve(T[c]));
});

afterEach(() => {
  vi.useRealTimers();
});

const pressed = (name: string) => screen.getByRole("button", { name }).getAttribute("aria-pressed");

describe("RunCosts", () => {
  it("run costs period and project chips refetch", async () => {
    render(<RunCosts />);
    await screen.findByRole("button", { name: "gymie" });
    const period = screen.getByRole("group", { name: "Period" });
    expect(within(period).getAllByRole("button").map((b) => b.textContent)).toEqual(["7 days", "30 days", "90 days"]);
    expect(pressed("30 days")).toBe("true");
    const project = screen.getByRole("group", { name: "Project" });
    expect(within(project).getAllByRole("button").map((b) => b.textContent)).toEqual(["All", "agents", "gymie"]);
    expect(pressed("All")).toBe("true");
    expect(fetchStatsViewMock).toHaveBeenLastCalledWith("run-costs", { from: new Date(NOW - 30 * 86_400_000), to: new Date(NOW), project: undefined });

    await userEvent.click(screen.getByRole("button", { name: "90 days" }));
    expect(fetchStatsViewMock).toHaveBeenLastCalledWith("run-costs", { from: new Date(NOW - 90 * 86_400_000), to: new Date(NOW), project: undefined });
    expect(pressed("90 days")).toBe("true");

    await userEvent.click(screen.getByRole("button", { name: "gymie" }));
    expect(fetchStatsViewMock).toHaveBeenLastCalledWith("run-costs", { from: new Date(NOW - 90 * 86_400_000), to: new Date(NOW), project: P });
    expect(pressed("gymie")).toBe("true");
  });

  it("run costs tables draw the category and runs tables", async () => {
    render(<RunCosts />);
    const cat = await screen.findByRole("region", { name: "Time and cost by category, across runs" });
    const catTable = within(cat).getByRole("table");
    expect(within(catTable).getAllByRole("columnheader").map((h) => h.textContent)).toEqual(["Category", "Runs", "Rows", "Time (union)", "Cost", "Share of cost"]);
    expect(within(catTable).getAllByRole("row")[1].textContent).toBe("planned implementer3936h 54m$95.4727.4%");
    const catRow = within(catTable).getAllByRole("row")[1];
    expect(within(catTable).getAllByRole("columnheader").map((h) => h.className)).toEqual(["", "r", "r", "r", "r", "r"]);
    expect([...catRow.children].map((c) => c.className)).toEqual(["", "r lb-num", "r lb-num", "r lb-num", "r lb-num", "r lb-num"]);
    expect(within(cat).getByText("3 finished runs · union per run, summed across runs")).toBeInTheDocument();

    const runs = screen.getByRole("region", { name: "Runs" });
    expect(within(runs).getAllByRole("columnheader").map((h) => h.textContent)).toEqual(["", "Run", "Project", "Wall clock", "Agent time", "Cost", "Tasks", "Fix rounds"]);
    expect(within(runs).getByRole("link", { name: C1 })).toHaveAttribute("href", `#/run-costs/${P}/${C1}`);
    expect(within(runs).getByRole("link", { name: C3 })).toHaveAttribute("href", `#/run-costs/${Q}/${C3}`);
    expect(within(runs).getAllByRole("row")[1].textContent).toBe(`${C1}gymie49h 21m30h 41m$348.516113`);
    const boxes = within(runs).getAllByRole("checkbox");
    expect(boxes.map((b) => (b as HTMLInputElement).checked)).toEqual([true, true, false]);
    expect(boxes[0]).toHaveAccessibleName(`Compare ${C1}`);
  });

  it("run costs compare shows deltas for two ticked runs", async () => {
    render(<RunCosts />);
    const cmp = await screen.findByRole("region", { name: "Compare runs" });
    expect(within(cmp).getByText("B − A")).toBeInTheDocument();
    expect(within(cmp).getByLabelText("A")).toHaveValue(`${P}/${C1}`);
    expect(within(cmp).getByLabelText("B")).toHaveValue(`${P}/${C2}`);
    await within(cmp).findByText("gated fixer · cost");
    const table = within(cmp).getByRole("table");
    expect(within(table).getAllByRole("columnheader").map((h) => h.textContent)).toEqual(["Metric", "A", "B", "Δ", "Δ %"]);
    const rows = within(table).getAllByRole("row").slice(1);
    expect(rows.map((r) => r.firstElementChild?.textContent)).toEqual([
      "Wall clock",
      "Agent time",
      "Idle / parent",
      "Cost",
      "Tasks",
      "Fix rounds",
      "planned implementer · cost",
      "gated fixer · cost",
    ]);
    const cells = (i: number) => [...rows[i].children].map((c) => c.textContent);
    expect(cells(3)).toEqual(["Cost", "$348.51", "$141.20", "−$207.31", "−59%"]);
    expect(rows[3].children[3]).toHaveClass("delta-down");
    expect(cells(6)).toEqual(["planned implementer · cost", "$0.00", "$40.00", "+$40.00", "—"]);
    expect(rows[6].children[3]).toHaveClass("delta-up");

    const runs = screen.getByRole("region", { name: "Runs" });
    await userEvent.click(within(runs).getByRole("checkbox", { name: `Compare ${C3}` }));
    expect(within(cmp).getByLabelText("A")).toHaveValue(`${P}/${C2}`);
    expect(within(cmp).getByLabelText("B")).toHaveValue(`${Q}/${C3}`);
    expect(within(runs).getAllByRole("checkbox").map((b) => (b as HTMLInputElement).checked)).toEqual([false, true, true]);
    expect(runTimingsMock).toHaveBeenCalledWith(Q, C3);
  });

  it("run costs definitions list every metric", async () => {
    render(<RunCosts />);
    const defs = await screen.findByRole("region", { name: "Definitions" });
    await within(defs).findByText("Wall clock");
    const terms = defs.querySelectorAll("dt");
    expect([...terms].map((t) => t.textContent)).toEqual(DEFS.map((d) => `${d.name} ${d.kind}`));
    expect([...defs.querySelectorAll("dd")].map((d) => d.textContent)).toEqual(DEFS.map((d) => d.definition));
    expect(terms[5].querySelector(".src.estimated")).not.toBeNull();
  });

  it("run costs fetches each picked run once and drops a stale error or result", async () => {
    const pending: Record<string, (t: RunTimingsDTO) => void> = {};
    let fail: (e: Error) => void = () => {};
    runTimingsMock.mockImplementation(
      (_p: string, c: string) =>
        new Promise<RunTimingsDTO>((res, rej) => {
          pending[c] = res;
          fail = rej;
        }),
    );
    render(<RunCosts />);
    await vi.waitFor(() => expect(runTimingsMock).toHaveBeenCalledTimes(2));
    pending[C1](T[C1]);
    await screen.findByText("Wall clock", { selector: "dt" });
    const failC2 = fail;
    await userEvent.click(screen.getByRole("checkbox", { name: `Compare ${C3}` }));
    expect(runTimingsMock.mock.calls.map((c) => c[1])).toEqual([C1, C2, C3]);

    await userEvent.click(screen.getByRole("button", { name: "90 days" }));
    await act(async () => {});
    await act(async () => failC2(new Error("late failure")));
    expect(screen.queryByText("late failure")).toBeNull();

    // A late success is dropped too, and its run asked for again when picked.
    await act(async () => pending[C3](T[C3]));
    await userEvent.click(screen.getByRole("checkbox", { name: `Compare ${C3}` }));
    expect(runTimingsMock.mock.calls.map((c) => c[1]).slice(3)).toContain(C3);
  });
});
