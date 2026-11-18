// RunTimings is one run's timings and costs page (KAN-1132, the
// prototype's view B), drawn from GET /api/v1/stats/run/{project}/{change}.
// getRunTimings is mocked at the module boundary, the way RunDetail.test.tsx
// mocks ../api, so these tests assert the page's own drawing.
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { RunMetric, RunTimings as RunTimingsDTO } from "../api";
import { RunTimings } from "./RunTimings";

const { runTimingsMock } = vi.hoisted(() => ({ runTimingsMock: vi.fn() }));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return { ...actual, getRunTimings: runTimingsMock };
});

const P = "gymie-7c1f238a";
const C = "kan-1091-body-calories-measurements";
const H = 3_600_000;
const FROM = Date.parse("2026-11-16T16:47:00Z");
const at = (h: number) => new Date(FROM + h * H).toISOString();
const m = <T,>(value: T, definition: string, kind: RunMetric<T>["kind"]): RunMetric<T> => ({ value, definition, kind });

const columns: RunTimingsDTO["breakdowns"]["category"]["columns"] = [
  { name: "Time share", definition: "Each instant split equally among the rows active in it; rows sum to wall clock.", kind: "derived" },
  { name: "Busy", definition: "Union of the row's own intervals; not additive.", kind: "derived" },
  { name: "Cost", definition: "Sum of the rows' priced cost_usd.", kind: "measured" },
];
const tail = [
  { label: "parent commands", rows: 412, shareMs: 6 * H, busyMs: 6 * H, cost: 0, noTokens: 0 },
  { label: "idle / parent", rows: 0, shareMs: 12 * H, busyMs: 12 * H, cost: 8.51, noTokens: 0 },
];
const total = { label: "Total", rows: 460, shareMs: 49 * H, busyMs: 0, cost: 348.51, noTokens: 2 };

function fixture(): RunTimingsDTO {
  return {
    project: P,
    change: C,
    from: at(0),
    to: at(49.35),
    open: false,
    totals: {
      wallClockMs: m(49.35 * H, "First stage begin to last stage end (now while a stage is open).", "derived"),
      agentTimeMs: m(30 * H, "Union of every dispatch row's [begin, end).", "derived"),
      parentCommandsMs: m(6.2 * H, "Union of recorded parent-side runs.", "measured"),
      idleMs: m(12.47 * H, "Wall clock minus the union of dispatches and parent commands.", "derived"),
      cost: m(348.51, "Sum of every dispatch row's priced cost_usd plus the parent session's own cost.", "measured"),
      parentSessionCost: m(8.51, "The parent session's own cost.", "measured"),
      visualVerifyShare: m(0.28, "Union of visual verifier rows over wall clock.", "derived"),
      dispatches: 183,
      noTokens: 2,
      parentCommandRuns: 412,
      unclaimedGuardRuns: 0,
    },
    lanes: [
      { label: "flow.brainstorm", kind: "stage", intervals: [{ from: at(0), to: at(2), class: "completed" }] },
      { label: "planned implementer", kind: "category", intervals: [{ from: at(3), to: at(6), class: "completed" }] },
      { label: "visual verifier", kind: "category", intervals: [{ from: at(33), to: at(34), class: "completed" }] },
      { label: "parent commands", kind: "parent", intervals: [{ from: at(2), to: at(3), class: "parent" }] },
      { label: "idle / parent", kind: "idle", intervals: [{ from: at(6), to: at(7), class: "idle" }] },
    ],
    breakdowns: {
      category: { columns, rows: [{ label: "planned implementer", rows: 31, shareMs: 7 * H, busyMs: 7 * H, cost: 95.47, noTokens: 0 }, ...tail], total },
      role: { columns, rows: [{ label: "implementer", rows: 58, shareMs: 24 * H, busyMs: 21 * H, cost: 235.69, noTokens: 0 }, ...tail], total },
      model: { columns, rows: [...tail], total },
      repo: { columns, rows: [{ label: "gymie-frontend", rows: 51, shareMs: 9 * H, busyMs: 8 * H, cost: 112.6, noTokens: 1 }, ...tail], total },
      round: { columns, rows: [...tail], total },
    },
    costPerTask: [
      { task: "31", rows: 3, shareMs: 1 * H, cost: 9.88, estimated: false },
      { task: "13", rows: 5, shareMs: 0.5 * H, cost: 3.68, estimated: true, group: "13+14+16+20" },
    ],
    costPerVerifyRound: [
      { round: "verify", rows: 2, shareMs: 1.5 * H, cost: 6.2 },
      { round: "visual-verify-fix-1-rest", rows: 3, shareMs: 2 * H, cost: 13.95 },
    ],
    counts: {
      fixRounds: m({ verify: 13, panel: 3 }, "Distinct verify-fix-N and visual-verify-fix-N rounds, plus panel-fix rounds.", "derived"),
      departures: m([{ round: "visual-verify-fix-1-rest-gymie", departures: 9 }, { round: "visual-verify-fix-2-gymie", departures: 0 }], "The departures: line of each visual-verify report, by round.", "measured"),
      findings: m({ critical: 2, important: 11, minor: 19 }, "Recorded review findings by canonical severity.", "measured"),
      flakyRetries: m({ retries: 7, runs: 55 }, "Tests that passed only on their one retry.", "measured"),
      rows: m({ open: 0, aborted: 5, blocked: 15 }, "Dispatch rows by outcome.", "measured"),
      journalPending: m(0, "Record writes for this change still in the local journal.", "measured"),
    },
    dataQuality: [
      { id: "open-stage-runs", label: "open stage runs", count: 0, detail: [] },
      { id: "superseded-stage-runs", label: "superseded stage runs", count: 0, detail: [] },
      { id: "unlinked", label: "dispatch rows with no stage link", count: 0, detail: [] },
      { id: "no-tokens", label: "dispatch rows with no tokens — reason stated", count: 2, detail: ["task-13+14-reviewer · transcript harvested late", "verify-fix-13-gymie · skipped, 0 s"] },
      { id: "aborted-ran", label: "aborted rows that ran", count: 0, detail: [] },
      { id: "uncategorised", label: "uncategorised dispatch keys", count: 0, detail: [] },
      { id: "unreadable-openers", label: "unreadable openers", count: 0, detail: [] },
      { id: "unreconciled-cost", label: "unreconciled cost", count: 1, detail: [] },
      { id: "journal", label: "record writes pending in the journal", count: 0, detail: [] },
      { id: "retimed", label: "retimed dispatch rows", count: 0, detail: [] },
      { id: "unclaimed-guard-runs", label: "unclaimed guard runs", count: 0, detail: [] },
    ],
    definitions: [],
  };
}

beforeEach(() => {
  runTimingsMock.mockReset();
  runTimingsMock.mockResolvedValue(fixture());
});

async function page() {
  return screen.findByRole("region", { name: `Run timings and costs: ${P}/${C}` });
}

describe("RunTimings", () => {
  it("run timings page draws the totals strip", async () => {
    render(<RunTimings project={P} change={C} />);
    const root = await page();
    expect(within(root).getByRole("link", { name: "← Run detail" })).toHaveAttribute("href", `#/run/${P}/${C}`);
    expect(within(root).getByRole("heading", { name: `${C} · timings and costs` })).toBeInTheDocument();
    expect(within(root).getByRole("link", { name: "Compare with another run →" })).toHaveAttribute("href", "#/run-costs");
    const totals = within(root).getByRole("region", { name: "Run totals" });
    const labels = [...totals.querySelectorAll(".lb-kpi > .lb-lbl")].map((l) => l.firstChild?.textContent);
    expect(labels).toEqual(["Wall clock", "Agent time", "Parent commands", "Idle / parent", "Cost", "Visual-verify loops"]);
    const vals = [...totals.querySelectorAll(".lb-kpi-val")].map((v) => v.textContent);
    expect(vals).toEqual(["49h 21m", "30h 00m", "6h 12m", "12h 28m", "$348.51", "28%"]);
    const wall = totals.querySelector(".lb-kpi .src");
    expect(wall).toHaveTextContent("derived");
    expect(wall).toHaveAttribute("title", "First stage begin to last stage end (now while a stage is open).");
    expect(within(totals).getByText("183 dispatches · 2 unpriced")).toBeInTheDocument();
  });

  it("run timings timeline draws a lane per category, parent commands and idle", async () => {
    render(<RunTimings project={P} change={C} />);
    const img = await screen.findByRole("img", { name: "Run timeline" });
    const texts = [...img.querySelectorAll("text")].map((t) => t.textContent);
    for (const l of ["flow.brainstorm", "dispatches", "planned implementer", "visual verifier", "parent commands", "idle / parent", "+0h", "+48h"]) {
      expect(texts).toContain(l);
    }
    const titles = [...img.querySelectorAll("rect > title")].map((t) => t.textContent);
    expect(titles).toContain("flow.brainstorm: 2h 00m (+0h 00m → +2h 00m)");
    expect(titles).toContain("planned implementer: 3h 00m (+3h 00m → +6h 00m)");
    expect(titles).toContain("idle / parent: 1h 00m (+6h 00m → +7h 00m)");
    expect(img.querySelector("pattern#hatch")).not.toBeNull();
  });

  it("run timings group-by chips switch the breakdown", async () => {
    render(<RunTimings project={P} change={C} />);
    const panel = await screen.findByRole("region", { name: "Time and cost by" });
    const group = within(panel).getByRole("group", { name: "Group by" });
    expect(within(group).getAllByRole("button").map((b) => b.textContent)).toEqual(["Category", "Role", "Model / effort", "Repo", "Fix round"]);
    expect(within(group).getByRole("button", { name: "Category" })).toHaveAttribute("aria-pressed", "true");
    const table = within(panel).getByRole("table");
    expect(within(table).getAllByRole("columnheader")[0]).toHaveTextContent("Category");
    await userEvent.click(within(group).getByRole("button", { name: "Repo" }));
    expect(within(group).getByRole("button", { name: "Repo" })).toHaveAttribute("aria-pressed", "true");
    expect(within(table).getAllByRole("columnheader")[0]).toHaveTextContent("Repo");
    const rows = within(table).getAllByRole("row").map((r) => [...r.querySelectorAll("td")].map((c) => c.textContent));
    expect(rows[1]).toEqual(["gymie-frontend", "51", "9h 00m", "8h 00m", "$112.60", "1"]);
    expect(rows[2][0]).toBe("parent commands");
    expect(rows[2][4]).toBe("—");
    expect(rows[3][0]).toBe("idle / parent");
    expect(rows[3][4]).toBe("—");
    expect(rows[4]).toEqual(["Total", "460", "49h 00m", "—", "$348.51", "2"]);
  });

  it("run timings data quality lists every gap", async () => {
    render(<RunTimings project={P} change={C} />);
    const panel = await screen.findByRole("region", { name: "Data quality" });
    const items = [...panel.querySelectorAll(":scope .lb-panel-b > ul > li")];
    const labels = items.map((li) => li.querySelector("summary")?.firstChild?.textContent ?? li.children[1].textContent);
    expect(labels).toEqual([
      "open stage runs",
      "dispatch rows with no stage link",
      "dispatch rows with no tokens — reason stated",
      "aborted rows that ran",
      "uncategorised dispatch keys",
      "unreconciled cost",
      "record writes pending in the journal",
      "retimed dispatch rows",
    ]);
    expect(items[0].querySelector(".n")).toHaveClass("ok");
    expect(items[2].querySelector(".n")).toHaveClass("lb-err");
    const detail = [...items[2].querySelectorAll("details li")].map((l) => l.textContent);
    expect(detail).toEqual(["task-13+14-reviewer · transcript harvested late", "verify-fix-13-gymie · skipped, 0 s"]);
  });

  it("run timings cost per task marks a split group estimated", async () => {
    render(<RunTimings project={P} change={C} />);
    const tasks = await screen.findByRole("region", { name: "Cost per task" });
    expect(within(tasks).getByText("top 2 of 2")).toBeInTheDocument();
    const row = within(tasks).getByText("13 · ¼ of group 13+14+16+20").closest("tr")!;
    const est = row.querySelector(".src.estimated");
    expect(est).toHaveTextContent("estimated");
    expect(est).toHaveAttribute("title", "A group dispatch’s cost split evenly across its tasks.");
    const rounds = screen.getByRole("region", { name: "Cost per verify round" });
    expect(within(rounds).getByText("2 rounds")).toBeInTheDocument();
    const cells = within(rounds).getAllByRole("row").slice(1).map((r) => [...r.querySelectorAll("td")].map((c) => c.textContent));
    expect(cells).toEqual([
      ["verify", "verify", "—", "1h 30m", "$6.20"],
      ["1", "visual-verify-fix-1-rest", "9", "2h 00m", "$13.95"],
    ]);
  });
});
