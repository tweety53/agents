// One test per view asserting it renders the actual numbers a fixture
// response carries (never merely that the component mounted -- the
// highest-risk vacuous shape this task's own instructions name), plus the
// absence-vs-zero distinction and the state-board-is-default-route
// requirement.
import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, onTestFinished, vi } from "vitest";
import type {
  DecisionRow,
  HealthViewSlug,
  StatsResponse,
  StatsViewSlug,
  ViewName,
  WeeklyModelEffortViewSlug,
} from "../api";
import { App } from "../App";
import { DynamicsChart } from "../components/DynamicsChart";
import { Decisions } from "./Decisions";
import { FlowHealth } from "./FlowHealth";
import { ReviewDynamics } from "./ReviewDynamics";
import { Reviewers } from "./Reviewers";
import { SelfReview } from "./SelfReview";
import { StageLeaderboard } from "./StageLeaderboard";
import { StateBoard } from "./StateBoard";
import { Trend } from "./Trend";
import { WeeklyCosts } from "./WeeklyCosts";

const { fetchStatsViewMock } = vi.hoisted(() => ({ fetchStatsViewMock: vi.fn() }));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return { ...actual, fetchStatsView: fetchStatsViewMock };
});

const period = { from: new Date("2026-01-01T00:00:00Z"), to: new Date("2026-02-01T00:00:00Z") };

function envelope<Row>(view: StatsViewSlug, rows: Row, recorded = true, unmeasured = false): StatsResponse<Row> {
  return {
    view,
    from: period.from.toISOString(),
    to: period.to.toISOString(),
    boundaryConvention: "a stage run is attributed to the period containing its start instant",
    recorded,
    unmeasured,
    rows,
  };
}

// A weekly-costs row with every category null; fixtures spread it and set
// the categories they carry.
const noPhase = {
  planningUsd: null,
  implementationUsd: null,
  reviewUsd: null,
  selfReviewUsd: null,
  visualVerifyUsd: null,
  otherUsd: null,
};

// Every slug a view fetches: the navigable views, less "flow-health", which
// fetches the three health slugs instead of one of its own.
function reviewDynamicsRows() {
  const pt = (bucket: string, metric: string, series: string, value: number, basis: number) => ({ bucket, metric, series, value, basis });
  const weeks = ["2026-01-05", "2026-01-12"];
  const series = weeks.flatMap((w, i) => [
    pt(w, "findings-per-change", "critical", 0.5 - i * 0.25, 4),
    pt(w, "findings-per-change", "important", 1, 4),
    pt(w, "findings-per-change", "minor", 2 + i, 4),
    pt(w, "per-dispatch-slot", "primary", 2, 3),
    pt(w, "fix-rounds", "mean", 1.5, 4),
    pt(w, "fix-rounds", "max", 3, 4),
    pt(w, "per-dispatch-model", "opus", 2.5, 5),
    pt(w, "per-dispatch-model-version", "claude-opus-5-5", 2, 3),
    pt(w, "per-dispatch-model-version", "opus (version unrecorded)", 3, 2),
    pt(w, "per-dispatch-effort", "low", 1, 5),
    pt(w, "changes", "changes", 4, 4),
    pt(w, "impl-findings-per-task", "opus · low", 1.5, 6),
    pt(w, "impl-findings-per-task-version", "claude-opus-5-5 · low", 1.5, 6),
    pt(w, "impl-blocked-share", "blocked", 0.1, 10),
    pt(w, "vv-rounds", "mean", 1, 2),
    pt(w, "vv-rounds", "max", 2, 2),
    pt(w, "vv-first-pass", "first-pass", 0.5, 2),
  ]);
  const heat = (model: string) => [
    { slot: "primary", model, effort: "low", dispatches: 4, critical: 1, important: 2, minor: 5 },
    { slot: "bugbot", model, effort: "low", dispatches: 2, critical: 0, important: 1, minor: 0 },
  ];
  return {
    bucket: "week",
    series,
    heat: heat("opus"),
    previousHeat: [{ slot: "primary", model: "opus", effort: "low", dispatches: 4, critical: 2, important: 4, minor: 6 }],
    heatVersion: heat("claude-opus-5-5"),
    previousHeatVersion: [],
    summary: { findingsPerChange: 0.75, criticalImportantPerChange: 0.98, meanFixRounds: 1.5, implFindingsPerTask: 1.5, vvFirstPassShare: 0.6, changes: 8 },
    previousSummary: { findingsPerChange: 1, criticalImportantPerChange: 1, meanFixRounds: 2, implFindingsPerTask: 2, vvFirstPassShare: 0.5, changes: 6 },
    markers: [
      { sha: "abc1234def:skills/flow/primary-reviewer-prompt.md", at: "2026-01-13T10:00:00Z", kind: "edit", area: "review", slot: "primary", target: "skills/flow/primary-reviewer-prompt.md", summary: "re-read every writer", caseId: null },
    ],
    lastPromptChangeAt: "2026-01-13T10:00:00Z",
    projects: ["agents-a740d89c", "gymie-7c1f238a"],
  };
}

const fixtures: Record<
  Exclude<ViewName, "flow-health"> | HealthViewSlug | WeeklyModelEffortViewSlug,
  StatsResponse<unknown>
> = {
  "state-board": envelope("state-board", [
    {
      projectKey: "kan-16-stats-app",
      name: "kan-16-stats-app",
      state: "IN_PROGRESS",
      updatedAt: "2026-01-15T10:00:00Z",
      updatedBy: "alice",
      nextCommand: "/flow-fast",
      plannedTasks: 22,
      currentTasks: 46,
      runCount: 7,
      costUsd: 12.5,
      cost24h: 0,
      cost7d: 1,
      cost30d: 2,
      wallClockMs: 5400000,
      errors: 3,
      currentStage: "flow.sdd-tdd",
      currentStageStartedAt: "2026-01-15T11:30:00Z",
      doneTasks: 20,
      subagents: [
        { role: "implementer", model: "opus", startedAt: "2026-01-15T11:50:00Z" },
        { role: "reviewer", model: "opus", startedAt: "2026-01-15T11:55:00Z" },
        { role: "implementer", model: "opus", startedAt: "2026-01-15T11:56:00Z", inline: true },
      ],
      activity: "working",
      activityDetail: "2 subagents running",
      activitySince: "2026-01-15T11:59:00Z",
    },
    {
      projectKey: "kan-16-stats-app",
      name: "kan-29-never-observed",
      state: "STARTED",
      updatedAt: "2026-01-15T09:00:00Z",
      updatedBy: "bob",
      nextCommand: "/flow",
      runCount: 0,
      costUsd: 0,
      cost24h: 0,
      cost7d: 1,
      cost30d: 2,
      wallClockMs: 0,
      errors: 0,
      currentStage: "flow.brainstorm",
      currentStageStartedAt: "2026-01-15T11:50:00Z",
      activity: "question",
      activityDetail: "Which layout?",
      activitySince: "2026-01-15T11:57:00Z",
    },
    {
      projectKey: "gymie-7c1f238a",
      name: "kan-5-old",
      state: "STARTED",
      tasks: [{ n: 6, title: "In-run fix 1 — visual-verify: divider missing", state: "in progress" }],
      subagents: [{ role: "implementer", taskId: "6", model: "opus", startedAt: "2026-01-15T11:58:00Z", inline: true }],
      lastStage: "flow.writing-plans",
      lastStageEndedAt: "2026-01-15T11:55:00Z",
      updatedAt: "2026-01-01T10:00:00Z",
      updatedBy: "erin",
      nextCommand: "/flow",
      runCount: 1,
      costUsd: 2,
      cost24h: 0,
      cost7d: 1,
      cost30d: 2,
      wallClockMs: 1000,
      errors: 0,
    },
    {
      projectKey: "kan-16-stats-app",
      name: "kan-31-replan-fold",
      state: "FINISHED",
      updatedAt: "2026-01-15T08:00:00Z",
      updatedBy: "carol",
      nextCommand: "/flow",
      plannedTasks: 46,
      currentTasks: 22,
      runCount: 2,
      costUsd: 1.25,
      cost24h: 0,
      cost7d: 1,
      cost30d: 2,
      wallClockMs: 45000,
      errors: 11,
    },
    {
      projectKey: "kan-16-stats-app",
      name: "pipeline-fix-x-14",
      state: "FINISHED",
      updatedAt: "2026-01-15T07:00:00Z",
      updatedBy: "dave",
      nextCommand: "",
      runCount: 1,
      costUsd: 0.5,
      cost24h: 0,
      cost7d: 1,
      cost30d: 2,
      wallClockMs: 1000,
      errors: 0,
    },
    {
      projectKey: "kan-16-stats-app",
      name: "self-review-kan-31-replan-fold",
      state: "FINISHED",
      updatedAt: "2026-01-15T07:00:00Z",
      updatedBy: "dave",
      nextCommand: "",
      runCount: 0,
      costUsd: 0,
      cost24h: 0,
      cost7d: 0,
      cost30d: 0,
      wallClockMs: 0,
      errors: 0,
    },
  ]),
  "stage-leaderboard": envelope("stage-leaderboard", [
    {
      command: "/flow",
      stage: "SDD + TDD per task",
      runCount: 3,
      meanCostUsd: 1.5,
      medianCostUsd: 1.2,
      p90CostUsd: 2.9,
      totalCostUsd: 4.5,
    },
  ]),
  trend: envelope("trend", [{ day: "2026-01-05", runCount: 2, totalCostUsd: 5.25 }]),
  "weekly-costs": envelope("weekly-costs", [
    { ...noPhase, week: "2026-06-03", overallUsd: 5, implementationUsd: 5 },
    {
      ...noPhase,
      week: "2026-06-10",
      overallUsd: 110,
      planningUsd: 10,
      reviewUsd: 30,
      selfReviewUsd: 43,
      visualVerifyUsd: 20,
      otherUsd: 7,
    },
    { ...noPhase, week: "2026-06-17", overallUsd: 11, planningUsd: 4, implementationUsd: 7, otherUsd: 0.5 },
  ]),
  "weekly-model-effort-costs": envelope("weekly-model-effort-costs", [
    { week: "2026-06-03", model: "claude-opus-5-5", effort: "low", costUsd: 2 },
    { week: "2026-06-03", model: "claude-opus-5-5", effort: "session", costUsd: 3 },
    { week: "2026-06-10", model: "claude-opus-5-5", effort: "low", costUsd: 30 },
    { week: "2026-06-10", model: "claude-opus-5-5", effort: "session", costUsd: 60 },
    { week: "2026-06-10", model: "claude-sonnet-5-5", effort: "medium", costUsd: 20 },
    { week: "2026-06-17", model: "claude-opus-5-5", effort: "session", costUsd: 11 },
  ]),
  reviewers: envelope("reviewers", [
    {
      slot: "primary",
      experimental: false,
      description: "",
      dispatches: 2,
      changes: 1,
      critical: 1,
      important: 1,
      minor: 1,
    },
    {
      slot: "exp-failure-modes",
      experimental: true,
      description: "What the diff does under error, timeout and partial write",
      dispatches: 1,
      changes: 1,
      critical: 1,
      important: 0,
      minor: 1,
    },
  ]),
  decisions: envelope("decisions", [
    {
      project: "kan",
      change: "kan-1",
      recordedAt: "2026-06-10T09:10:00Z",
      class: "regular",
      overridden: true,
      execution: "inline",
      implementerModel: "skipped — inline",
      implementerEffort: "",
      fixer: "skipped — inline",
      rerunDispatch: "",
      rosterSize: 2,
      compact: true,
      experimentalSlot: "exp-failure-modes",
      rerun: "delta",
      grouping: "free",
      dispatches: "primary+exp-failure-modes",
      implementerGroups: "",
      wallClockSeconds: 600,
      inputTokens: 100,
      outputTokens: 50,
      cacheReadTokens: 20,
      costUsd: 1.25,
      critical: 1,
      important: 1,
      minor: 1,
      fixRounds: 2,
      fallbacks: 1,
      timedOut: 1,
    },
    {
      project: "kan",
      change: "kan-2",
      recordedAt: "2026-06-11T09:10:00Z",
      class: "regular",
      overridden: false,
      execution: "sdd",
      implementerModel: "opus",
      implementerEffort: "high",
      fixer: "sonnet/medium",
      rerunDispatch: "haiku/low",
      rosterSize: 5,
      compact: false,
      experimentalSlot: "",
      rerun: "full",
      grouping: "static",
      dispatches: "primary+principles · bugbot+mutation",
      implementerGroups: "1+2 · 3",
      wallClockSeconds: 1200,
      inputTokens: 200,
      outputTokens: 100,
      cacheReadTokens: 40,
      costUsd: 3.75,
      critical: 0,
      important: 2,
      minor: 1,
      fixRounds: 1,
      fallbacks: 0,
      timedOut: 0,
    },
  ]),
  "flow-summary": envelope("flow-summary", {
    changesStarted: 4,
    changesFinished: 3,
    totalCostUsd: 42.5,
    meanCostPerFinished: 9.25,
    medianLeadTimeMs: 2 * 86400000 + 43200000,
    meanPanelRounds: 2.5,
    critical: 1,
    important: 2,
    minor: 4,
    guardFires: 10,
  }),
  "self-review": envelope("self-review", [
    { recordedAt: "2026-01-12T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-fix", note: "guard misses a case", disposition: "fixed", ref: "abc1234", blastRadius: 3, commitUrl: "https://github.com/o/r/commit/abc1234" },
    { recordedAt: "2026-01-11T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-cost", note: "cheaper reviewer", disposition: "fixed", ref: "def5678", blastRadius: null, commitUrl: "https://github.com/o/r/commit/def5678" },
    { recordedAt: "2026-01-10T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-speed", note: "needs a redesign", disposition: "filed", ref: "KAN-9", blastRadius: null, commitUrl: "" },
    { recordedAt: "2026-01-09T10:00:00Z", project: "agents", change: "kan-2", angle: "flow-improvement", note: "not worth it", disposition: "declined", ref: "", blastRadius: null, commitUrl: "" },
    { recordedAt: "2026-01-08T10:00:00Z", project: "agents", change: "kan-2", angle: "flow-improvement", note: "panel caught it early", disposition: "noted", ref: "", blastRadius: null, commitUrl: "" },
  ]),
  "review-dynamics": envelope("review-dynamics", reviewDynamicsRows()),
};

// The board filters by Updated relative to now; every case runs at one
// pinned instant so which fixture rows fall inside 24h/7d is fixed.
const BOARD_NOW = new Date("2026-01-15T12:00:00Z");

beforeEach(() => {
  vi.useFakeTimers({ toFake: ["Date"], now: BOARD_NOW });
  onTestFinished(() => {
    vi.useRealTimers();
  });
  fetchStatsViewMock.mockReset();
  fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures) => Promise.resolve(fixtures[view]));
});

describe("views render their fixture response's actual values", () => {
  it("state board filters by chips and state tiles", async () => {
    const user = userEvent.setup();
    render(<StateBoard />);
    const table = await screen.findByRole("table");
    const tile = (name: RegExp) => screen.getByRole("button", { name });
    const visible = (name: string) => within(table).queryByRole("link", { name }) !== null;
    const clear = () => user.click(screen.getByRole("button", { name: "Clear filters" }));
    // The board fetches the whole history once and filters here.
    expect(fetchStatsViewMock.mock.calls[0][1].from.getTime()).toBe(0);
    // ...through past now: a change updated this instant is inside the half-open period.
    expect(fetchStatsViewMock.mock.calls[0][1].to.getTime()).toBeGreaterThan(BOARD_NOW.getTime());
    // Updated defaults to 7d: the fortnight-old change is hidden, and the
    // tiles count only what the other filters leave.
    expect(visible("kan-5-old")).toBe(false);
    expect(tile(/^Started/)).toHaveTextContent("1");
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();
    await user.click(within(screen.getByRole("group", { name: "Updated" })).getByRole("button", { name: "All" }));
    expect(tile(/^Started/)).toHaveTextContent("2");
    expect(visible("kan-5-old")).toBe(true);
    // A Project chip carries its count and narrows rows and tiles.
    const gymie = within(screen.getByRole("group", { name: "Project" })).getByRole("button", { name: /^gymie/ });
    expect(gymie).toHaveTextContent("gymie1");
    await user.click(gymie);
    expect(tile(/^Started/)).toHaveTextContent("1");
    expect(tile(/^In progress/)).toHaveTextContent("0");
    expect(visible("kan-16-stats-app")).toBe(false);
    // Clear filters resets every filter, Updated back to 7d.
    await clear();
    expect(visible("kan-5-old")).toBe(false);
    expect(visible("kan-16-stats-app")).toBe(true);
    await user.type(screen.getByRole("searchbox", { name: "Search changes" }), "KAN-29");
    expect(visible("kan-29-never-observed")).toBe(true);
    expect(visible("kan-16-stats-app")).toBe(false);
    await clear();
    // A state tile filters by its state and opens its collapsed group.
    await user.click(tile(/^Finished/));
    expect(tile(/^Finished/)).toHaveAttribute("aria-pressed", "true");
    expect(visible("kan-31-replan-fold")).toBe(true);
    expect(visible("kan-16-stats-app")).toBe(false);
  });

  it("state board running now cards", async () => {
    render(<StateBoard />);
    const cards = await screen.findAllByTestId("running-card");
    // One card per unfinished change or change with an open stage; the one
    // awaiting an answer comes first and is marked as needing the operator.
    expect(cards).toHaveLength(3);
    expect(cards[0]).toHaveTextContent("kan-29-never-observed");
    expect(cards[0]).toHaveClass("needs");
    expect(cards[0]).toHaveTextContent("Awaiting your answer");
    expect(cards[0]).toHaveTextContent("Which layout?");
    expect(cards[0]).toHaveTextContent("Designing the change");
    expect(cards[1]).toHaveTextContent("kan-16-stats-app");
    expect(cards[1]).not.toHaveClass("needs");
    expect(cards[1]).toHaveTextContent("Working");
    expect(cards[1]).toHaveTextContent("Implementing tasks· 30m");
    expect(cards[1]).toHaveTextContent("tasks 20/46 · $12.50");
    // The card keeps agents, main's inline work included, to a count; the run view lists them.
    expect(within(cards[1]).getByTestId("agent-count")).toHaveTextContent("3 agents running");
    // An open stage alone is the main session, one agent; a change between stages has none.
    expect(within(cards[0]).getByTestId("agent-count")).toHaveTextContent("1 agent running");
    expect(within(cards[2]).getByTestId("agent-count")).toHaveTextContent("1 agent running");
    // A change between stages still names its latest stage and what runs next.
    expect(cards[2]).toHaveTextContent("kan-5-old");
    // ...unless a dispatch works a task between stages, which the card names.
    expect(cards[2]).toHaveTextContent("Working between stages");
    expect(within(cards[2]).getByTestId("running-task")).toHaveTextContent("Task 6: In-run fix 1 — visual-verify: divider missing");
    expect(screen.getByText("Active runs", { selector: ".lb-lbl" }).nextSibling).toHaveTextContent("3");
  });

  it("state board groups by state", async () => {
    const user = userEvent.setup();
    render(<StateBoard />);
    const table = within(await screen.findByRole("table"));
    const groups = table.getAllByRole("row").filter((r) => r.classList.contains("lb-group"));
    expect(groups.map((g) => g.textContent)).toEqual(["▾In progress1", "▾Started1", "▾Finished1"]);
    // Finished starts collapsed; its row shows once the group is opened.
    expect(table.queryByRole("link", { name: "kan-31-replan-fold" })).not.toBeInTheDocument();
    await user.click(groups[2]);
    expect(table.getByRole("link", { name: "kan-31-replan-fold" })).toBeInTheDocument();
    // A row: state pill, current stage, tasks with growth, relative update
    // with the ISO time as its title, cost; no errors column.
    const row = table.getByRole("link", { name: "kan-16-stats-app" }).closest("tr")!;
    expect(row).toHaveTextContent("IN_PROGRESS");
    expect(row).toHaveTextContent("Implementing tasks");
    expect(within(row).getByText("+24")).toBeInTheDocument();
    expect(within(row).getByTitle("2026-01-15T10:00:00.000Z")).toHaveTextContent("2h ago");
    expect(row).toHaveTextContent("$12.50");
    expect(table.queryByRole("columnheader", { name: /errors/i })).not.toBeInTheDocument();
    // A session that is not working says so beside its stage.
    const asking = table.getByRole("link", { name: "kan-29-never-observed" }).closest("tr")!;
    expect(asking).toHaveTextContent("Awaiting your answer");
    expect(row).not.toHaveTextContent("Working");
    expect(table.queryByRole("columnheader", { name: /Next command/ })).not.toBeInTheDocument();
  });

  it("state board cost in period follows only the period", async () => {
    const user = userEvent.setup();
    render(<StateBoard />);
    await screen.findByRole("table");
    // The prototype's tile sums every change whatever chip, tile or search is set.
    const cost = () => screen.getByText("Cost in period", { selector: ".lb-lbl" }).nextSibling!.textContent;
    const tile = within(screen.getByRole("group", { name: "Updated" }));
    await user.click(tile.getByRole("button", { name: "All" }));
    const all = cost();
    await user.click(within(screen.getByRole("group", { name: "Project" })).getByRole("button", { name: /^gymie/ }));
    expect(cost()).toBe(all);
    await user.click(screen.getByRole("button", { name: /^Finished/ }));
    expect(cost()).toBe(all);    // It counts only stage runs started within the chosen window.
    await user.click(tile.getByRole("button", { name: "24h" }));
    expect(cost()).toBe("$0.00");
    await user.click(tile.getByRole("button", { name: "30d" }));
    expect(cost()).not.toBe(all);
  });

  it("state board stops refetching while the tab is hidden", async () => {
    vi.useFakeTimers({ toFake: ["Date", "setInterval", "clearInterval"], now: BOARD_NOW });
    const hidden = vi.spyOn(document, "hidden", "get").mockReturnValue(true);
    const { unmount } = render(<StateBoard />);
    await act(async () => {});
    const before = fetchStatsViewMock.mock.calls.length;
    await act(async () => vi.advanceTimersByTime(15000));
    expect(fetchStatsViewMock.mock.calls.length).toBe(before);
    // Showing the tab again refetches at once.
    hidden.mockReturnValue(false);
    vi.setSystemTime(BOARD_NOW.getTime() + 15000);
    await act(async () => document.dispatchEvent(new Event("visibilitychange")));
    expect(fetchStatsViewMock.mock.calls.length).toBe(before + 1);
    // Unmounted, the board no longer listens for the tab coming back.
    const removed = vi.spyOn(document, "removeEventListener");
    unmount();
    expect(removed).toHaveBeenCalledWith("visibilitychange", expect.any(Function));
    removed.mockRestore();
    hidden.mockRestore();
  });

  it("state board hides pipeline-fix and self-review changes", async () => {
    const user = userEvent.setup();
    render(<StateBoard />);
    const table = within(await screen.findByRole("table"));
    await user.click(table.getByRole("row", { name: /^▾Finished/ }));
    expect(table.getByRole("link", { name: "kan-31-replan-fold" })).toBeInTheDocument();
    expect(table.queryByRole("link", { name: "pipeline-fix-x-14" })).not.toBeInTheDocument();
    expect(table.queryByRole("link", { name: "self-review-kan-31-replan-fold" })).not.toBeInTheDocument();
  });

  it("stage leaderboard shows total, mean, median and p90 cost, and its own stat panels", async () => {
    render(<StageLeaderboard period={period} project={undefined} />);
    expect(await screen.findByText("$1.50")).toBeInTheDocument();
    expect(screen.getByText("$1.20")).toBeInTheDocument();
    expect(screen.getByText("$2.90")).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: /Total cost/ })).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "$4.50" })).toBeInTheDocument();
    within(screen.getByRole("region", { name: "Stages" })).getByText("1");
    within(screen.getByRole("region", { name: "Total runs" })).getByText("3");
  });

  it("trend shows the day's run count and total cost, its time-series panel, and its own stat panels", async () => {
    render(<Trend period={period} project={undefined} />);
    expect(await screen.findByText("2026-01-05")).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "$5.25" })).toBeInTheDocument();
    // Step 2: the daily points render through TimeSeriesPanel now, not the
    // view's own bespoke bar chart.
    expect(screen.getByTestId("time-series-point")).toBeInTheDocument();
    within(screen.getByRole("region", { name: "Days" })).getByText("1");
    within(screen.getByRole("region", { name: "Total cost" })).getByText("$5.25");
  });

  // Shows the phase series, which start hidden.
  const showPhases = async () => {
    for (const name of ["Planning", "Implementation", "Review", "Self-review", "Visual verify", "Other"])
      await userEvent.click(screen.getByRole("button", { name }));
  };
  // The phase chart's panel, so a query never reads the model·effort chart.
  const phases = () => within(screen.getByRole("region", { name: "Cost by phase" }));
  const modelEffort = () => within(screen.getByRole("region", { name: "Cost by model · effort" }));

  it("weekly costs draws one 8-week phase chart with seven toggleable series", async () => {
    // The window ends at "now"; pin it to the fixture's weeks.
    vi.useFakeTimers({ now: new Date("2026-06-19T12:00:00Z"), toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    render(<WeeklyCosts period={period} project="p-a" model="m-a" />);
    await screen.findAllByText("Cost (USD)");
    expect(phases().getByText("Cost (USD)")).toBeInTheDocument();
    expect(phases().getByText("Week starting (Wed, UTC)")).toBeInTheDocument();
    expect(screen.queryByRole("table")).toBeNull();
    // The last 8 Wed->Wed weeks over every project and model, whatever the
    // props say -- from a Wednesday 00:00 UTC, 7 to 8 weeks back.
    const [, params] = fetchStatsViewMock.mock.calls[0];
    expect(params.project).toBeUndefined();
    expect(params.model).toBeUndefined();
    const from = params.from as Date;
    expect(from.getUTCDay()).toBe(3);
    expect(from.getTime() % (24 * 60 * 60 * 1000)).toBe(0);
    const weeksBack = (Date.now() - from.getTime()) / (7 * 24 * 60 * 60 * 1000);
    expect(weeksBack).toBeGreaterThanOrEqual(7);
    expect(weeksBack).toBeLessThan(8);
    expect(phases().getAllByTestId("week-tick").map((t) => t.textContent)).toEqual([
      "2026-04-29",
      "2026-05-06",
      "2026-05-13",
      "2026-05-20",
      "2026-05-27",
      "2026-06-03",
      "2026-06-10",
      "2026-06-17",
    ]);
    // Five even steps from $0 reach the fixture's $110 peak.
    expect(phases().getAllByTestId("cost-tick").map((t) => t.textContent)).toEqual([
      "$0.00",
      "$25.00",
      "$50.00",
      "$75.00",
      "$100.00",
      "$125.00",
    ]);
    // Overall alone is shown at first.
    const toggles = phases().getAllByRole("button", { pressed: true });
    expect(toggles.map((b) => b.textContent)).toEqual(["Overall"]);
    expect(phases().getAllByRole("button").map((b) => b.textContent)).toEqual([
      "Overall",
      "Planning",
      "Implementation",
      "Review",
      "Self-review",
      "Visual verify",
      "Other",
    ]);
    const points = (name: string) =>
      phases()
        .queryAllByTestId("time-series-point")
        .filter((c) => c.getAttribute("data-series") === name);
    expect(points("Review")).toHaveLength(0);
    const review = screen.getByRole("button", { name: "Review" });
    await userEvent.click(review);
    expect(review).toHaveAttribute("aria-pressed", "true");
    expect(points("Review")).toHaveLength(1);
    expect(points("Overall")).toHaveLength(3);
    await userEvent.click(review);
    expect(review).toHaveAttribute("aria-pressed", "false");
    expect(points("Review")).toHaveLength(0);
  });

  // A dot's pill, and the delta inside it.
  const label = (series: string, week: string) =>
    document.querySelector(`[data-testid="delta-label"][data-series="${series}"][data-week="${week}"]`);
  const delta = (series: string, week: string) => label(series, week)?.querySelector("tspan") ?? null;

  it("weekly costs labels each point with its change from the previous week", async () => {
    vi.useFakeTimers({ now: new Date("2026-06-19T12:00:00Z"), toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    render(<WeeklyCosts period={period} project={undefined} />);
    await screen.findAllByText("Cost (USD)");
    expect(delta("Overall", "2026-06-10")).toHaveTextContent("▲2100%");
    expect(delta("Overall", "2026-06-10")).toHaveClass("delta-up");
    expect(delta("Overall", "2026-06-17")).toHaveTextContent("▼90%");
    expect(delta("Overall", "2026-06-17")).toHaveClass("delta-down");
    const point = screen
      .getAllByTestId("time-series-point")
      .find((c) => c.getAttribute("data-series") === "Overall" && c.getAttribute("data-week") === "2026-06-17")!;
    fireEvent.mouseEnter(point);
    const tip = screen.getByRole("tooltip");
    expect(tip).toHaveTextContent("Overall");
    expect(tip).toHaveTextContent("$11.00");
    expect(tip).toHaveTextContent("▼ 90.0%");
    fireEvent.mouseLeave(point);
    expect(screen.queryByRole("tooltip")).toBeNull();
  });

  it("weekly costs breaks a line at an unmeasured week", async () => {
    vi.useFakeTimers({ now: new Date("2026-06-19T12:00:00Z"), toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    render(<WeeklyCosts period={period} project={undefined} />);
    await screen.findAllByText("Cost (USD)");
    await showPhases();
    const d = (name: string) =>
      phases()
        .getAllByTestId("weekly-line")
        .find((l) => l.getAttribute("data-series") === name)!
        .getAttribute("d")!;
    // Implementation is Unavailable on 2026-06-10: two runs, never one line across it.
    expect(d("Implementation").match(/M/g)).toHaveLength(2);
    expect(d("Overall").match(/M/g)).toHaveLength(1);
    // Every series with a measured week draws its line.
    expect(
      phases()
        .getAllByTestId("weekly-line")
        .map((l) => l.getAttribute("data-series")),
    ).toEqual(["Overall", "Planning", "Implementation", "Review", "Self-review", "Visual verify", "Other"]);
    const point = screen
      .getAllByTestId("time-series-point")
      .find((c) => c.getAttribute("data-series") === "Overall" && c.getAttribute("data-week") === "2026-06-10")!;
    expect(point).toHaveAttribute("aria-label", "Overall 2026-06-10: $110.00");
  });

  it("weekly costs draws a series with no measured week as no line", async () => {
    vi.useFakeTimers({ now: new Date("2026-06-19T12:00:00Z"), toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    fetchStatsViewMock.mockResolvedValueOnce(
      envelope("weekly-costs", [{ ...noPhase, week: "2026-06-10", overallUsd: 3, planningUsd: 1, implementationUsd: 2 }]),
    );
    render(<WeeklyCosts period={period} project={undefined} />);
    await screen.findAllByText("Cost (USD)");
    await showPhases();
    const series = (testId: string) =>
      phases()
        .queryAllByTestId(testId)
        .map((e) => e.getAttribute("data-series"));
    expect(series("weekly-line")).not.toContain("Review");
    expect(series("time-series-point")).toEqual(["Overall", "Planning", "Implementation"]);
  });

  it("weekly costs labels zero changes, skips a zero previous week and stacks close labels apart", async () => {
    vi.useFakeTimers({ now: new Date("2026-06-19T12:00:00Z"), toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    const row = { ...noPhase, visualVerifyUsd: 1, otherUsd: 1 };
    fetchStatsViewMock.mockResolvedValueOnce(
      envelope("weekly-costs", [
        { ...row, week: "2026-04-29", overallUsd: 10, planningUsd: 0, implementationUsd: 5, reviewUsd: 4, selfReviewUsd: 3 },
        { ...row, week: "2026-05-06", overallUsd: 12, planningUsd: 2, implementationUsd: 5, reviewUsd: 4, selfReviewUsd: 3 },
        { ...row, week: "2026-05-13", overallUsd: 1, planningUsd: 1, implementationUsd: 1, reviewUsd: 1, selfReviewUsd: 1 },
      ]),
    );
    render(<WeeklyCosts period={period} project={undefined} />);
    await screen.findAllByText("Cost (USD)");
    await showPhases();
    // The window's second week reads its change from the first.
    expect(delta("Overall", "2026-05-06")).toHaveTextContent("▲20%");
    // A previous week that cost nothing carries no percentage, only its cost.
    expect(label("Planning", "2026-05-06")).toHaveTextContent(/^\$2\.00$/);
    // An unchanged week is a neutral 0%, no arrow.
    expect(delta("Implementation", "2026-05-06")).toHaveTextContent(/^0%$/);
    expect(delta("Implementation", "2026-05-06")).toHaveClass("delta-flat");
    // Seven labels on one dot height: pushed 17 apart and lifted back inside the plot.
    const ys = ["Overall", "Planning", "Implementation", "Review", "Self-review", "Visual verify", "Other"]
      .map((s) => Number(label(s, "2026-05-13")!.getAttribute("y")))
      .sort((a, b) => a - b);
    ys.slice(1).forEach((y, i) => expect(y - ys[i]).toBeGreaterThanOrEqual(17));
    expect(Math.max(...ys)).toBeLessThanOrEqual(412 + 3.5);
  });

  it("weekly costs draws a Visual verify series", async () => {
    vi.useFakeTimers({ now: new Date("2026-06-19T12:00:00Z"), toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    render(<WeeklyCosts period={period} project={undefined} />);
    await screen.findAllByText("Cost (USD)");
    const toggles = phases().getAllByRole("button");
    expect(toggles).toHaveLength(7);
    expect(toggles.slice(-2).map((b) => b.textContent)).toEqual(["Visual verify", "Other"]);
    await userEvent.click(phases().getByRole("button", { name: "Visual verify" }));
    await userEvent.click(phases().getByRole("button", { name: "Other" }));
    const points = (name: string) =>
      phases()
        .queryAllByTestId("time-series-point")
        .filter((c) => c.getAttribute("data-series") === name)
        .map((c) => c.getAttribute("aria-label"));
    expect(points("Visual verify")).toEqual(["Visual verify 2026-06-10: $20.00"]);
    expect(points("Other")).toEqual(["Other 2026-06-10: $7.00", "Other 2026-06-17: $0.50"]);
  });

  it("weekly costs labels every dot with its cost", async () => {
    vi.useFakeTimers({ now: new Date("2026-06-19T12:00:00Z"), toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    render(<WeeklyCosts period={period} project={undefined} />);
    await screen.findAllByText("Cost (USD)");
    await showPhases();
    // The window's first measured week has no previous: its cost alone.
    expect(label("Overall", "2026-06-03")).toHaveTextContent(/^\$5\.00$/);
    expect(label("Overall", "2026-06-10")).toHaveTextContent("$110.00");
    expect(label("Overall", "2026-06-10")).toHaveTextContent("▲2100%");
    // Implementation's 2026-06-10 is a gap: the week after reads its cost alone.
    expect(label("Implementation", "2026-06-17")).toHaveTextContent(/^\$7\.00$/);
  });

  it("weekly costs matches the prototype header and pill widths", async () => {
    vi.useFakeTimers({ now: new Date("2026-06-19T12:00:00Z"), toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    render(<WeeklyCosts period={period} project={undefined} />);
    await screen.findAllByText("Cost (USD)");
    await showPhases();
    expect(screen.getByText("Wednesday-to-Wednesday cost, last 8 weeks.")).toBeInTheDocument();
    const width = (el: Element | null) => Number(el?.parentElement?.querySelector("rect")?.getAttribute("width"));
    // The prototype's fixed widths, 40 for a cost alone and 84 with a delta,
    // wherever the text fits; a longer text widens its pill to fit.
    const pills = [...document.querySelectorAll('[data-testid="delta-label"]')];
    const fits = (p: Element) => Math.ceil(6 + (p.textContent ?? "").length * 5.5);
    for (const p of pills) {
      const base = p.querySelector("tspan") ? 84 : 40;
      expect(width(p)).toBe(Math.max(base, fits(p)));
    }
    expect(width(label("Overall", "2026-06-03"))).toBe(40);
    expect(pills.some((p) => p.querySelector("tspan") && width(p) === 84)).toBe(true);
  });

  it("weekly costs widens a pill whose figure overflows the prototype's width", async () => {
    vi.useFakeTimers({ now: new Date("2026-06-19T12:00:00Z"), toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    fetchStatsViewMock.mockResolvedValueOnce(
      envelope("weekly-costs", [
        { ...noPhase, week: "2026-06-03", overallUsd: 1000 },
        { ...noPhase, week: "2026-06-10", overallUsd: 12345.67 },
      ]),
    );
    render(<WeeklyCosts period={period} project={undefined} />);
    await screen.findAllByText("Cost (USD)");
    const width = (el: Element | null) => Number(el?.parentElement?.querySelector("rect")?.getAttribute("width"));
    // "$1,000.00" (9 characters) and "$12,345.67 ▲1135%" (17): 6px in, 5.5px a character.
    await waitFor(() => expect(label("Overall", "2026-06-03")).toHaveTextContent(/^\$1,000\.00$/));
    expect(width(label("Overall", "2026-06-03"))).toBe(56);
    expect(label("Overall", "2026-06-10")).toHaveTextContent("$12,345.67 ▲1135%");
    expect(width(label("Overall", "2026-06-10"))).toBe(100);
  });

  it("weekly costs draws a model·effort chart from the second view", async () => {
    vi.useFakeTimers({ now: new Date("2026-06-19T12:00:00Z"), toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    render(<WeeklyCosts period={period} project="p-a" model="m-a" />);
    await screen.findAllByText("Cost (USD)");
    expect(fetchStatsViewMock).toHaveBeenCalledTimes(2);
    const calls = fetchStatsViewMock.mock.calls;
    expect(calls.map(([view]) => view)).toEqual(["weekly-costs", "weekly-model-effort-costs"]);
    expect(calls[1][1].from).toEqual(calls[0][1].from);
    expect(calls[1][1].to).toEqual(calls[0][1].to);
    // Lines ordered by total cost, every one shown at first.
    expect(modelEffort().getAllByRole("button", { pressed: true }).map((b) => b.textContent)).toEqual([
      "opus-5-5 · session",
      "opus-5-5 · low",
      "sonnet-5-5 · medium",
    ]);
    const points = (name: string) =>
      modelEffort()
        .queryAllByTestId("time-series-point")
        .filter((c) => c.getAttribute("data-series") === name);
    expect(points("opus-5-5 · low")).toHaveLength(2);
    await userEvent.click(modelEffort().getByRole("button", { name: "opus-5-5 · low" }));
    expect(points("opus-5-5 · low")).toHaveLength(0);
    expect(points("opus-5-5 · session")).toHaveLength(3);
    expect(phases().queryAllByTestId("time-series-point")).toHaveLength(3);
    const pill = modelEffort()
      .getAllByTestId("delta-label")
      .find((l) => l.getAttribute("data-series") === "opus-5-5 · session" && l.getAttribute("data-week") === "2026-06-10")!;
    expect(pill).toHaveTextContent("$60.00");
    expect(pill).toHaveTextContent("▲1900%");
  });

  it("reviewers shows severity counts and badges an exp- slot with its description", async () => {
    render(<Reviewers period={period} project={undefined} />);
    expect(await screen.findByRole("cell", { name: /primary/ })).toBeInTheDocument();
    // primary's own severity counts, one dispatch/change/finding count per
    // column -- distinct values so a swap or a dropped field cannot hide.
    const primaryRow = screen.getByRole("cell", { name: /primary/ }).closest("tr")!;
    expect(within(primaryRow).getByText("2")).toBeInTheDocument(); // dispatches

    const expRow = screen.getByRole("cell", { name: /exp-failure-modes/ }).closest("tr")!;
    const badge = within(expRow).getByRole("img", { name: "experimental reviewer" });
    expect(badge).toHaveAttribute("title", "What the diff does under error, timeout and partial write");
    expect(screen.queryByRole("columnheader", { name: /Findings\/dispatch/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("columnheader", { name: /Withdrawn/ })).not.toBeInTheDocument();

    within(screen.getByRole("region", { name: "Slots" })).getByText("2");
    within(screen.getByRole("region", { name: "Total dispatches" })).getByText("3");
  });

  it("reviewers efficiency weighs critical 10, important 4, minor 1 per dispatch", async () => {
    const row = { experimental: false, description: "", changes: 1 };
    fetchStatsViewMock.mockResolvedValueOnce(
      envelope("reviewers", [
        { ...row, slot: "primary", dispatches: 166, critical: 4, important: 36, minor: 70 },
        { ...row, slot: "idle", dispatches: 0, critical: 0, important: 0, minor: 0 },
      ]),
    );
    render(<Reviewers period={period} project={undefined} />);
    expect(await screen.findByRole("columnheader", { name: /Efficiency/ })).toBeInTheDocument();
    const primaryRow = screen.getByRole("cell", { name: "primary" }).closest("tr")!;
    expect(within(primaryRow).getByRole("cell", { name: "1.53" })).toBeInTheDocument();
    const idleRow = screen.getByRole("cell", { name: "idle" }).closest("tr")!;
    expect(within(idleRow).getByRole("cell", { name: "—" })).toBeInTheDocument();
  });

  it("renders self-review fixes", async () => {
    const user = userEvent.setup();
    const { unmount } = render(<SelfReview period={period} project={undefined} />);

    // A fixed finding's ref links to its commit; a filed one's key is text.
    const link = await screen.findByRole("link", { name: "abc1234" });
    expect(link).toHaveAttribute("href", "https://github.com/o/r/commit/abc1234");
    expect(screen.getByRole("cell", { name: "KAN-9" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "KAN-9" })).not.toBeInTheDocument();

    within(screen.getByRole("region", { name: "Fixed" })).getByText("2");
    within(screen.getByRole("region", { name: "Filed" })).getByText("1");

    // design.md's self-review-drop-per-change: no per-change block.
    expect(screen.queryByRole("region", { name: "Per change" })).not.toBeInTheDocument();

    // The outcome column filters the findings table.
    await user.selectOptions(screen.getByRole("combobox", { name: "Filter by Outcome" }), "declined");
    expect(screen.getByRole("cell", { name: "not worth it" })).toBeInTheDocument();
    expect(screen.queryByRole("cell", { name: "guard misses a case" })).not.toBeInTheDocument();
    expect(screen.queryByRole("cell", { name: "panel caught it early" })).not.toBeInTheDocument();
    unmount();

    fetchStatsViewMock.mockResolvedValueOnce(envelope("self-review", []));
    render(<SelfReview period={period} project={undefined} />);
    expect(await screen.findByText("No self-review findings in this period.")).toBeInTheDocument();
  });

  // Also covers kan-472 task 23's grouping/dispatches/implementer-groups columns, asserted below.
  it("decisions shows a class × execution summary and no per-decision table", async () => {
    render(<Decisions period={period} project={undefined} />);

    // The class × execution summary groups the two runs into their own
    // rows -- "regular / inline" and "regular / sdd" -- each with its own
    // mean cost.
    const summaryRegion = await screen.findByRole("region", { name: "By class and execution" });
    expect(await within(summaryRegion).findByText("regular / inline")).toBeInTheDocument();
    expect(within(summaryRegion).getByText("regular / sdd")).toBeInTheDocument();
    expect(within(summaryRegion).getByText("$1.25")).toBeInTheDocument();
    expect(within(summaryRegion).getByText("$3.75")).toBeInTheDocument();

    // design.md's decisions-by-implementer-model: the per-decision table
    // and the Runs tile are gone.
    expect(screen.queryByRole("region", { name: "Every decision" })).not.toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "Runs" })).not.toBeInTheDocument();
  });

  it("decisions groups runs by implementer model", async () => {
    const [inline, sdd] = fixtures.decisions.rows as DecisionRow[];
    fetchStatsViewMock.mockResolvedValueOnce(
      envelope("decisions", [inline, sdd, { ...sdd, change: "kan-3", costUsd: 1.25, wallClockSeconds: 600 }]),
    );
    render(<Decisions period={period} project={undefined} />);

    const region = await screen.findByRole("region", { name: "By implementer model" });
    const rows = (await within(region).findAllByRole("row")).slice(1);
    expect(rows).toHaveLength(2);
    // opus first: two runs, mean cost (3.75 + 1.25) / 2, mean wall-clock
    // (1200 s + 600 s) / 2, mean findings 3, mean fix rounds 1.
    within(rows[0]).getByRole("cell", { name: "opus" });
    within(rows[0]).getByRole("cell", { name: "2" });
    within(rows[0]).getByRole("cell", { name: "$2.50" });
    within(rows[0]).getByRole("cell", { name: "15.0 min" });
    within(rows[0]).getByRole("cell", { name: "3.00" });
    within(rows[0]).getByRole("cell", { name: "1.00" });
    within(rows[1]).getByRole("cell", { name: "skipped — inline" });
    within(rows[1]).getByRole("cell", { name: "1" });
  });

});

describe("absence is rendered distinctly from a recorded zero", () => {
  it("a period before the store held anything is stated, not rendered as an empty measured table", async () => {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(envelope(view, [], false)),
    );
    render(<Trend period={period} project={undefined} />);
    // Every one of Trend's panels shares the same underlying `state`
    // (task 20's recomposition: one useStatsView call per view, several
    // Panel instances reading it), so each independently renders the
    // not-recorded banner rather than one page-level banner -- Panel is
    // now the sole implementation of that branch (this file's own header
    // comment, and Panel.tsx's), so seeing it repeated per panel is
    // exactly what "not duplicated across two files" looks like from one.
    const banners = await screen.findAllByTestId("not-recorded");
    expect(banners.length).toBeGreaterThan(0);
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.queryByTestId("time-series-point")).not.toBeInTheDocument();
  });

  // Task 5's third arm, exercised end to end through a real view rather
  // than Panel.test.tsx's isolated fixture: a period whose runs were
  // recorded but never attributed must read as its own state, distinct
  // from both "no data was recorded" (the case immediately above) and
  // from an ordinary measured table -- never collapsing into either.
  it("a period whose runs were recorded but none attributed reports that plainly, not as 'no data was recorded'", async () => {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(envelope(view, [], true, true)),
    );
    render(<Trend period={period} project={undefined} />);
    const banners = await screen.findAllByTestId("unmeasured");
    expect(banners.length).toBeGreaterThan(0);
    for (const banner of banners) {
      expect(banner).toHaveTextContent("Runs were recorded for this period, but none carried measurements.");
    }
    expect(screen.queryByTestId("not-recorded")).not.toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.queryByTestId("time-series-point")).not.toBeInTheDocument();
  });
});

describe("the state board is the only navigation path into a change's own dashboard", () => {
  it("links a row's change name to its run-detail route, round-tripping a URL-significant character in both segments", async () => {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(
        envelope(view, [
          {
            projectKey: "kan/16",
            name: "kan-16 stats-app?v2",
            state: "IN_PROGRESS",
            updatedAt: "2026-01-15T10:00:00Z",
            updatedBy: "alice",
            nextCommand: "/flow-fast",
          },
        ]),
      ),
    );

    render(<StateBoard />);

    const link = await screen.findByRole("link", { name: "kan-16 stats-app?v2" });
    expect(link).toHaveAttribute(
      "href",
      `#/run/${encodeURIComponent("kan/16")}/${encodeURIComponent("kan-16 stats-app?v2")}`,
    );
  });
});

// Task 2 (kan-183): the Project column names a project, it does not key
// it. The cell shows the display name, the full key survives as the
// cell's title (the identity the operator still needs when two checkouts
// share a basename), and the row's own link target keeps the full key
// regardless, since that is the route's identity and not this task's to
// shorten.
describe("the Project column names a project instead of keying it", () => {
  const PROJECT_KEY = "agents-a740d89c";

  it("state board's Project cell shows the display name with the full key as its title", async () => {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(
        envelope(view, [
          {
            projectKey: PROJECT_KEY,
            name: "kan-1",
            state: "IN_PROGRESS",
            updatedAt: "2026-01-15T10:00:00Z",
            updatedBy: "alice",
            nextCommand: "/flow-fast",
          },
        ]),
      ),
    );

    render(<StateBoard />);

    const cell = await screen.findByRole("cell", { name: "agents" });
    expect(cell).toHaveTextContent(/^agents$/);
    expect(cell.querySelector(`[title="${PROJECT_KEY}"]`)).not.toBeNull();

    // The row's link target is the route's identity -- it keeps the full
    // key even though the cell beside it now reads short.
    const link = screen.getByRole("link", { name: "kan-1" });
    expect(link).toHaveAttribute("href", `#/run/${encodeURIComponent(PROJECT_KEY)}/${encodeURIComponent("kan-1")}`);
  });

});

// F3 (panel round 1, Major): DataTable's exact-match filter and free-text
// search both key off a column's `accessor`. Task 2 set the Project
// column's accessor to the *display name*, so the filter dropdown offered
// one option per distinct display name -- and two projects whose keys
// differ only in the disambiguating hash suffix collapsed into that one
// option, silently merging what resolveProjectParam
// (stats/internal/api/stats.go) refuses server-side with a 400 naming the
// ambiguity. The fix returns the accessor to the full key: the dropdown
// now offers one option per key, each of which narrows to exactly the one
// project it names.
describe("the Project filter narrows by key identity, not by display name", () => {
  const BASENAME = "agents";
  const KEY_ONE = `${BASENAME}-a740d89c`;
  const KEY_TWO = `${BASENAME}-b851e9ad`;

  function twoProjectsSharingABasename(): void {
    fetchStatsViewMock.mockImplementation((view: ViewName) =>
      Promise.resolve(
        envelope(view, [
          {
            projectKey: KEY_ONE,
            name: "kan-1",
            state: "IN_PROGRESS",
            updatedAt: "2026-01-15T10:00:00Z",
            updatedBy: "alice",
            nextCommand: "/flow-fast",
          },
          {
            projectKey: KEY_TWO,
            name: "kan-2",
            state: "STARTED",
            updatedAt: "2026-01-15T11:00:00Z",
            updatedBy: "bob",
            nextCommand: "/flow",
          },
        ]),
      ),
    );
  }

  it("offers one Project chip per project key, each narrowing to its own project", async () => {
    twoProjectsSharingABasename();
    const user = userEvent.setup();
    render(<StateBoard />);

    const table = within(await screen.findByRole("table"));
    const chips = within(screen.getByRole("group", { name: "Project" })).getAllByRole("button");
    // "All" plus one chip per key, each titled with its full key: two keys
    // sharing a display name never merge into one filter.
    expect(chips).toHaveLength(3);
    const one = chips.find((c) => c.title === KEY_ONE)!;
    const two = chips.find((c) => c.title === KEY_TWO)!;

    await user.click(one);
    expect(table.getByRole("link", { name: "kan-1" })).toBeInTheDocument();
    expect(table.queryByRole("link", { name: "kan-2" })).not.toBeInTheDocument();

    await user.click(two);
    expect(table.getByRole("link", { name: "kan-2" })).toBeInTheDocument();
    expect(table.queryByRole("link", { name: "kan-1" })).not.toBeInTheDocument();
  });
});

describe("routing", () => {
  it("the live state board is the default route", async () => {
    window.location.hash = "";
    render(<App />);
    expect(await screen.findByRole("cell", { name: "IN_PROGRESS" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Live state board" })).toBeInTheDocument();
  });

  it("an unrecognised route also falls back to the state board", async () => {
    window.location.hash = "#/not-a-real-view";
    render(<App />);
    expect(await screen.findByRole("cell", { name: "IN_PROGRESS" })).toBeInTheDocument();
    window.location.hash = "";
  });
});

describe("flow health", () => {
  it("flow health shows eight totals tiles", async () => {
    render(<FlowHealth period={period} project={undefined} />);
    const tile = (name: string) => within(screen.getByRole("region", { name }));
    expect(await screen.findByRole("region", { name: "Changes started" })).toBeInTheDocument();
    await tile("Changes started").findByText("4");
    tile("Changes finished").getByText("3");
    tile("Total cost").getByText("$42.50");
    tile("Mean cost per finished change").getByText("$9.25");
    tile("Median lead time").getByText("2.5 d");
    tile("Mean panel rounds").getByText("2.50");
    tile("Findings (C/I/Mi)").getByText("1/2/4");
    tile("Guard fires").getByText("10");
    expect(document.querySelectorAll(".stat-panel")).toHaveLength(8);
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });

  it("flow health reads an unfinished period's means as unavailable", async () => {
    fetchStatsViewMock.mockResolvedValueOnce(
      envelope("flow-summary", {
        changesStarted: 1, changesFinished: 0, totalCostUsd: 0, meanCostPerFinished: null,
        medianLeadTimeMs: null, meanPanelRounds: null, critical: 0, important: 0, minor: 0, guardFires: 0,
      }),
    );
    render(<FlowHealth period={period} project={undefined} />);
    const lead = await screen.findByRole("region", { name: "Median lead time" });
    expect(await within(lead).findByTestId("unavailable")).toBeInTheDocument();
  });
});

describe("review dynamics", () => {
  const noRangeInUrl = () => window.history.replaceState(null, "", "#/review-dynamics");

  it("review dynamics orders the heatmap busiest slot first and effort low to high", async () => {
    noRangeInUrl();
    const rows = reviewDynamicsRows();
    const cell = (slot: string, model: string, effort: string, dispatches: number) => ({ slot, model, effort, dispatches, critical: 0, important: 1, minor: 1 });
    const pt = (series: string, metric = "per-dispatch-effort") => ({ bucket: "2026-01-05", metric, series, value: 1, basis: 2 });
    fetchStatsViewMock.mockResolvedValue(
      envelope("review-dynamics", {
        ...rows,
        series: [
          ...rows.series.filter((p) => p.metric !== "per-dispatch-effort" && p.metric !== "impl-findings-per-task"),
          ...["xhigh", "high", "default", "low", "medium"].map((e) => pt(e)),
          pt("opus · high", "impl-findings-per-task"),
          pt("opus · low", "impl-findings-per-task"),
        ],
        // test-sanity is busiest and sorts last by name; sonnet's low sorts before opus's high by effort alone.
        heat: [cell("mutation", "sonnet", "low", 2), cell("test-sanity", "opus", "high", 3), cell("test-sanity", "opus", "low", 3), cell("test-sanity", "opus", "medium", 1)],
      }),
    );
    render(<ReviewDynamics period={period} onPeriodChange={() => {}} project={undefined} />);
    const heat = (await screen.findByText(/Findings per dispatch: slot/)).closest(".dynamics-panel") as HTMLElement;
    expect(within(heat).getAllByRole("rowheader").map((h) => h.textContent)).toEqual(["test-sanity", "mutation"]);
    expect(within(heat).getAllByRole("columnheader").map((h) => h.textContent)).toEqual(["", "opuslow", "opusmedium", "opushigh", "sonnetlow"]);
    await userEvent.click(within(screen.getByRole("group", { name: "Prompt changes" })).getByRole("button", { name: "All" }));
    const effort = screen.getByRole("region", { name: "Findings per dispatch by reviewer effort" });
    expect([...effort.querySelectorAll(".dynamics-legend > *")].map((e) => e.textContent?.trim())).toEqual(["low", "medium", "high", "default", "xhigh", "prompt change"]);
    const impl = screen.getByRole("region", { name: "Panel findings per implementer task" });
    expect([...impl.querySelectorAll(".dynamics-legend > *")].map((e) => e.textContent?.trim())).toEqual(["opus · low", "opus · high"]);
  });

  it("review dynamics draws every section with tiles against the previous range", async () => {
    noRangeInUrl();
    render(<ReviewDynamics period={period} onPeriodChange={() => {}} project={undefined} />);
    for (const heading of ["Review panel", "Implementer", "Visual verify", "Where findings come from"]) {
      expect(await screen.findByText(heading)).toBeInTheDocument();
    }
    for (const title of [
      "Findings per change by severity", "Findings per dispatch by slot", "Fix rounds per change",
      "Findings per dispatch by reviewer model", "Findings per dispatch by reviewer effort", "Changes reviewed",
      "Panel findings per implementer task", "Implementer dispatches blocked or aborted",
      "Visual verify rounds per change", "Visual verify passed first time",
    ]) {
      expect(screen.getByRole("region", { name: title })).toBeInTheDocument();
    }
    const tile = (k: string) => screen.getByTestId(`tile-${k}`);
    expect(tile("Findings per change")).toHaveTextContent("0.75");
    expect(tile("Findings per change")).toHaveTextContent("▼ 25% · better");
    expect(tile("Visual verify passed first time")).toHaveTextContent("60%");
    expect(tile("Visual verify passed first time")).toHaveTextContent("better");
    expect(tile("Changes reviewed")).not.toHaveTextContent(/better|worse/);
    expect(tile("Critical + Important per change")).toHaveTextContent("≈ unchanged");
    // A previous range with changes but a zero value is an earlier range.
    expect(tile("Implementer findings per task")).toHaveTextContent("▼ 25% · better");
  });

  it("review dynamics tile rises from a previous zero", async () => {
    noRangeInUrl();
    const rows = reviewDynamicsRows();
    fetchStatsViewMock.mockResolvedValue(
      envelope("review-dynamics", { ...rows, previousSummary: { ...rows.previousSummary, meanFixRounds: 0 } }),
    );
    render(<ReviewDynamics period={period} onPeriodChange={() => {}} project={undefined} />);
    await screen.findByText("Review panel");
    expect(screen.getByTestId("tile-Mean fix rounds")).toHaveTextContent("▲ from 0 · worse");
  });

  it("review dynamics filters by project, task size and retired reviewers", async () => {
    fetchStatsViewMock.mockResolvedValue(envelope("review-dynamics", reviewDynamicsRows()));
    const onProjectChange = vi.fn();
    render(
      <ReviewDynamics period={period} onPeriodChange={() => {}} project={undefined} onProjectChange={onProjectChange} />,
    );
    await screen.findByText("Review panel");

    const projects = screen.getByRole("group", { name: "Projects" });
    expect(within(projects).getByRole("button", { name: "All" })).toHaveAttribute("aria-pressed", "true");
    await userEvent.click(within(projects).getByRole("button", { name: "gymie" }));
    expect(onProjectChange).toHaveBeenLastCalledWith("gymie-7c1f238a");

    const size = screen.getByRole("group", { name: "Task size" });
    expect(within(size).getByRole("button", { name: "All" })).toHaveAttribute("aria-pressed", "true");
    await userEvent.click(within(size).getByRole("button", { name: "Big" }));
    expect(fetchStatsViewMock.mock.calls.at(-1)![1]).toMatchObject({ class: "big" });

    const heat = (await screen.findByText(/Findings per dispatch: slot/)).closest(".dynamics-panel") as HTMLElement;
    expect(within(heat).queryByRole("rowheader", { name: "bugbot" })).toBeNull();
    await userEvent.click(within(screen.getByRole("group", { name: "Reviewers" })).getByRole("button", { name: "All" }));
    expect(within(heat).getByRole("rowheader", { name: "bugbot" })).toBeInTheDocument();
  });

  it("review dynamics range presets set the period", async () => {
    noRangeInUrl();
    const onPeriodChange = vi.fn();
    render(<ReviewDynamics period={period} onPeriodChange={onPeriodChange} project={undefined} />);
    await screen.findByText("Review panel");
    expect(screen.getByRole("button", { name: "Last 4 weeks" })).toHaveAttribute("aria-pressed", "true");
    const span = (p: { from: Date; to: Date }) => Math.round((p.to.getTime() - p.from.getTime()) / 86400000);
    expect(span(onPeriodChange.mock.calls[0][0])).toBe(28);
    // The first fetch already asks for the 4 weeks: App's default period is never fetched.
    expect(fetchStatsViewMock).toHaveBeenCalledTimes(1);
    expect(span(fetchStatsViewMock.mock.calls[0][1])).toBe(28);

    await userEvent.click(screen.getByRole("button", { name: "Last 2 weeks" }));
    const last = onPeriodChange.mock.calls.at(-1)![0];
    expect(span(last)).toBe(14);
    expect(Math.abs(last.to.getTime() - Date.now())).toBeLessThan(60000);
    expect(screen.getByRole("button", { name: "Last 2 weeks" })).toHaveAttribute("aria-pressed", "true");

    expect(screen.queryByRole("button", { name: "Since last prompt change" })).toBeNull();
  });

  it("review dynamics family and version toggle", async () => {
    noRangeInUrl();
    render(<ReviewDynamics period={period} onPeriodChange={() => {}} project={undefined} />);
    const model = await screen.findByRole("region", { name: "Findings per dispatch by reviewer model" });
    expect(within(model).getByText("opus")).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: /opus/ })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Version" }));
    const byVersion = screen.getByRole("region", { name: "Findings per dispatch by reviewer model version" });
    expect(within(byVersion).getByText("claude-opus-5-5")).toBeInTheDocument();
    expect(within(byVersion).getByText("opus (version unrecorded)")).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: /claude-opus-5-5/ })).toBeInTheDocument();
    expect(screen.getByText("Findings per dispatch: slot × model version · effort")).toBeInTheDocument();
  });

  it("review dynamics filters prompt-change markers by kind", async () => {
    noRangeInUrl();
    render(<ReviewDynamics period={period} onPeriodChange={() => {}} project={undefined} />);
    const severity = await screen.findByRole("region", { name: "Findings per change by severity" });
    const kinds = screen.getByRole("group", { name: "Prompt changes" });
    expect(within(kinds).getAllByRole("button").map((b) => b.textContent)).toEqual([
      "flow-prompt-tuning",
      "All",
      "auto prompt tuning",
    ]);
    expect(within(kinds).getByRole("button", { name: "flow-prompt-tuning" })).toHaveAttribute("aria-pressed", "true");
    const tip = () => {
      fireEvent.mouseEnter(severity.querySelector('[data-bucket="2026-01-12"]')!);
      return within(severity).getByRole("tooltip");
    };
    expect(tip()).not.toHaveTextContent("Prompt edit");
    await userEvent.click(within(kinds).getByRole("button", { name: "auto prompt tuning" }));
    expect(tip()).toHaveTextContent("Prompt edit · primary");
    await userEvent.click(within(kinds).getByRole("button", { name: "All" }));
    expect(tip()).toHaveTextContent("Prompt edit · primary");
  });

  it("dynamics chart labels each tick with the value it sits at", () => {
    const whole = (v: number) => v.toFixed(0);
    const percent = (v: number) => `${Math.round(v * 100)}%`;
    const tenths = (v: number) => v.toFixed(1);
    const cases: [number, (v: number) => string, string[]][] = [
      [0, whole, ["0", "1", "2"]],
      [1, whole, ["0", "1", "2"]],
      [2, whole, ["0", "1", "2"]],
      [0, tenths, ["0.0", "0.5", "1.0"]],
      // An empty share chart tops at 100%, never 200%.
      [0, percent, ["0%", "50%", "100%"]],
      // A 0.05 step would label 0.0, 0.1, 0.1.
      [0.06, tenths, ["0.0", "0.1", "0.2"]],
      [3, whole, ["0", "2", "4"]],
      [5, whole, ["0", "5", "10"]],
      // A share of 1 is 100%, never a 200% tick.
      [1, percent, ["0%", "50%", "100%"]],
    ];
    for (const [peak, format, want] of cases) {
      const { unmount } = render(
        <DynamicsChart
          title="Changes"
          description=""
          buckets={["2026-05-04", "2026-05-11"]}
          bucketKind="week"
          series={[{ name: "changes", values: [peak, 0], basis: [1, 1] }]}
          markers={[]}
          basisLabel="changes"
          format={format}
        />,
      );
      const ticks = [...document.querySelectorAll(".dynamics-tick")].filter((e) => !e.textContent?.includes("-"));
      // Ticks run bottom to top, each label at the height of the value it names.
      expect(ticks.map((e) => e.textContent)).toEqual(want);
      const ys = ticks.map((e) => Number(e.getAttribute("y")));
      expect(ys[0] - ys[1]).toBeCloseTo((ys[0] - ys[2]) / 2);
      // The top tick leaves 12% headroom in the 136px plot.
      expect(ys[0] - ys[2]).toBeCloseTo(136 / 1.12);
      unmount();
    }
  });

  it("review dynamics tooltip lists values, basis and marker", async () => {
    noRangeInUrl();
    render(<ReviewDynamics period={period} onPeriodChange={() => {}} project={undefined} />);
    const severity = await screen.findByRole("region", { name: "Findings per change by severity" });
    await userEvent.click(within(screen.getByRole("group", { name: "Prompt changes" })).getByRole("button", { name: "All" }));
    fireEvent.mouseEnter(severity.querySelector('[data-bucket="2026-01-12"]')!);
    const tip = within(severity).getByRole("tooltip");
    expect(tip).toHaveTextContent("Week of 01-12");
    expect(tip).toHaveTextContent("Critical: 0.3");
    expect(tip).toHaveTextContent("Minor: 3.0");
    expect(tip).toHaveTextContent("based on 4 changes");
    expect(tip).toHaveTextContent("◆ Prompt edit · primary: re-read every writer (abc1234)");

    const cell = screen.getByTitle(/primary · opus low/);
    expect(cell.getAttribute("title")).toBe("primary · opus low\n4 dispatches\n1 critical · 2 important · 5 minor");
  });
});
