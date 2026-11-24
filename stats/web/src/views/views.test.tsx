// One test per view asserting it renders the actual numbers a fixture
// response carries (never merely that the component mounted -- the
// highest-risk vacuous shape this task's own instructions name), plus the
// absence-vs-zero distinction and the state-board-is-default-route
// requirement.
import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, onTestFinished, vi } from "vitest";
import type {
  SelfReviewRow,
  ConfidenceRow,
  TrendChange,
  HealthViewSlug,
  StageLeaderboardRow,
  StatsResponse,
  StatsViewSlug,
  ViewName,
  WeeklyModelEffortViewSlug,
} from "../api";
import { App } from "../App";
import { AgentRoles } from "./AgentRoles";
import { FlowHealth, nested, panelOf } from "./FlowHealth";
import { DailyCosts } from "./DailyCosts";
import { ReviewDynamics } from "./ReviewDynamics";
import { Reviewers } from "./Reviewers";
import { SelfReview } from "./SelfReview";
import { Confidence } from "./Confidence";
import { StageLeaderboard } from "./StageLeaderboard";
import { StateBoard } from "./StateBoard";
import { Trend } from "./Trend";
import { WeeklyCosts } from "./WeeklyCosts";

const { fetchStatsViewMock, fetchProjectsMock, fetchRolesMock, postRolesEditMock, reviewMock, pageMock, confidenceMock, confidenceReviewMock } = vi.hoisted(() => ({
  confidenceMock: vi.fn(),
  confidenceReviewMock: vi.fn(),
  reviewMock: vi.fn(() => Promise.resolve()),
  pageMock: vi.fn(),
  postRolesEditMock: vi.fn(),
  fetchStatsViewMock: vi.fn(),
  fetchProjectsMock: vi.fn(),
  fetchRolesMock: vi.fn(),
}));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return {
    ...actual,
    fetchStatsView: fetchStatsViewMock,
    fetchProjects: fetchProjectsMock,
    fetchRoles: fetchRolesMock,
    postRolesEdit: postRolesEditMock,
    reviewSelfReviewFinding: reviewMock,
    fetchSelfReviewPage: pageMock,
    fetchConfidencePage: confidenceMock,
    reviewConfidenceEstimate: confidenceReviewMock,
  };
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

// One guard-health row: runs, fired, cannot-answer, median and summed run time, last fired instant.
const gh = (guard: string, runs: number, fired: number, cannotAnswer: number, medianMs: number, totalMs: number, lastFiredAt: string | null) => ({
  guard,
  runs,
  fired,
  cannotAnswer,
  medianMs,
  totalMs,
  lastFiredAt,
});

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
// fetches its health slug, "guard-health", instead of one of its own name.
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
    projects: ["agents-a740d89c", "gymie-7c1f238a"],
  };
}

// The stage leaderboard's fixture: twelve costed stages across two commands
// in the current window (one stage id stageLabel does not map), and a
// previous window missing flow.brainstorm, carrying flow.cleanup only there,
// and in which flow.self-review grew most ($20 -> $200).
const LEADERBOARD_CURRENT: [string, string, number, number | null][] = [
  ["/flow", "flow.sdd-tdd", 500, 400],
  ["/flow", "flow.review-panel", 300, 600],
  ["/flow", "flow.self-review", 200, 20],
  ["/flow-fast", "SDD + TDD per task", 100, 90],
  ["/flow", "flow.verify", 50, 40],
  ["/flow", "flow.brainstorm", 40, null],
  ["/flow", "flow.writing-plans", 30, 35],
  ["/flow", "flow.create-artifacts", 20, 25],
  ["/flow", "flow.kickoff", 10, 12],
  ["/flow", "flow.load-context", 5, 6],
  ["/flow-fast", "flow.verify", 3, 2],
  ["/flow-fast", "flow.cleanup", 2, 1],
];
const LEADERBOARD_PREVIOUS: [string, string, number][] = [
  ...LEADERBOARD_CURRENT.flatMap(([c, s, , prev]): [string, string, number][] => (prev === null ? [] : [[c, s, prev]])),
  ["/flow", "flow.cleanup", 9],
];
const LEADERBOARD_PROJECTS = [
  { projectKey: "gymie-7c1f238a", totalCostUsd: 1000 },
  { projectKey: "agents-a740d89c", totalCostUsd: 260 },
];
function leaderboardRows(rows: readonly (readonly [string, string, number, ...unknown[]])[]): StageLeaderboardRow[] {
  return rows.map(([command, stage, total], i) => ({
    command,
    stage,
    runCount: 10 + i,
    meanCostUsd: total / (10 + i),
    medianCostUsd: 1.25 + i,
    p90CostUsd: 3.5 + i,
    totalCostUsd: total,
    mainCostUsd: null,
    mainSharePct: i === 0 ? 45.6 : null,
  }));
}

// Trend rows at TREND_NOW (current week Aug 12; 12w opens May 27), listed
// out of week order. /flow fills Jul 1, Jul 8, Jul 22 and Jul 29 -- Jul 15
// empty between two filled weeks -- with Jul 22 skewed (median 3, mean 15);
// /flow-fast has more changes than /flow, so a most-frequent default fails.
const TREND_NOW = new Date("2026-08-15T09:00:00Z");
const AG = "agents-0a1b2c3d";
const GY = "gymie-7c1f238a";
const trendRow = (week: string, command: string, projectKey: string, costUsd: number): TrendChange => ({ week, command, projectKey, costUsd });
const TREND_ROWS: TrendChange[] = [
  trendRow("2026-07-22", "/flow", AG, 40),
  trendRow("2026-08-12", "/flow-fast", GY, 0.25),
  trendRow("2026-07-01", "/flow", AG, 10),
  trendRow("2026-07-29", "/flow", GY, 4),
  trendRow("2026-08-05", "/flow-fast", AG, 0.25),
  trendRow("2026-07-22", "/flow", AG, 2),
  trendRow("2026-08-12", "/flow-fast", GY, 0.25),
  trendRow("2026-07-08", "/flow", GY, 8),
  trendRow("2026-08-05", "/flow-fast", AG, 0.25),
  trendRow("2026-07-01", "/flow", AG, 12),
  trendRow("2026-08-12", "/flow-fast", AG, 0.25),
  trendRow("2026-07-22", "/flow", AG, 3),
  trendRow("2026-08-12", "/flow-fast", AG, 0.25),
  trendRow("2026-08-12", "/flow-fast", GY, 0.25),
  trendRow("2026-08-12", "/flow-fast", GY, 0.25),
];

// Weekly costs: three weeks of phase rows (Overall the sum of the
// categories, as the store builds it) and model · effort rows whose window
// totals order the pairs neither alphabetically nor by the last full week.
const WEEKLY_PROJECTS = [
  { projectKey: "agents-0123abcd", totalCostUsd: 80 },
  { projectKey: "gymie-89abcdef", totalCostUsd: 24 },
];
const WEEKLY_ROWS = [
  { ...noPhase, week: "2026-06-03", overallUsd: 75, planningUsd: 10, implementationUsd: 40, reviewUsd: 20, otherUsd: 5 },
  { ...noPhase, week: "2026-06-10", overallUsd: 92, planningUsd: 30, implementationUsd: 25, reviewUsd: 20, selfReviewUsd: 12, otherUsd: 5 },
  { ...noPhase, week: "2026-06-17", overallUsd: 12, planningUsd: 4, implementationUsd: 7, visualVerifyUsd: 1 },
];
const me = (week: string, model: string, effort: string, costUsd: number) => ({ week, model, effort, costUsd });
const WEEKLY_MODEL_ROWS = [
  me("2026-06-03", "claude-opus-5-5", "session", 20),
  me("2026-06-10", "claude-opus-5-5", "session", 30),
  me("2026-06-03", "claude-sonnet-5-5", "medium", 5),
  me("2026-06-10", "claude-sonnet-5-5", "medium", 35),
  me("2026-06-10", "claude-opus-5-5", "medium", 25),
  me("2026-06-10", "claude-opus-5", "high", 15),
  me("2026-06-10", "claude-opus-5-5", "low", 9),
  me("2026-06-17", "claude-opus-5", "low", 3),
  me("2026-06-03", "claude-sonnet-5-5", "session", 2),
  me("2026-06-03", "claude-sonnet-5", "low", 1),
];

// The Reviewers fixture, listed in an order no rank uses; beta and able tie on
// weighted per dispatch (2.50), which critical count and name break differently.
const rv = (slot: string, dispatches: number, pricedDispatches: number, costUsd: number, critical: number, important: number, minor: number, withdrawn: number) => ({
  slot,
  experimental: false,
  description: "",
  dispatches,
  changes: 1,
  critical,
  important,
  minor,
  costUsd,
  pricedDispatches,
  withdrawn,
});
const REVIEWER_ROWS = [
  rv("gamma", 5, 5, 2, 0, 0, 2, 1),
  { ...rv("exp-delta", 2, 2, 0.6, 0, 0, 0, 0), experimental: true, description: "D desc" },
  rv("alpha", 10, 10, 5, 0, 5, 0, 5),
  rv("eps", 3, 0, 0, 0, 1, 0, 0),
  rv("able", 2, 2, 1, 0, 0, 5, 0),
  rv("beta", 4, 2, 0.4, 1, 0, 0, 0),
];
const REVIEWER_PROJECTS = [
  { projectKey: "gymie-7c1f238a", dispatches: 20 },
  { projectKey: "agents-a740d89c", dispatches: 6 },
];

const fixtures: Record<
  Exclude<ViewName, "flow-health" | "agent-roles" | "role-costs" | "confidence"> | HealthViewSlug | WeeklyModelEffortViewSlug,
  StatsResponse<unknown, unknown>
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
      fixRounds: 5,
      reviewerRuns: 14,
      selfReviewFixRounds: 1,
      selfReviewReviewerRuns: 2,
      costUsd: 12.5,
      cost24h: 0,
      cost7d: 1,
      cost30d: 2,
      wallClockMs: 5400000,
      errors: 3,
      currentStage: "flow.sdd-tdd",
      currentStageStartedAt: "2026-01-15T11:30:00Z",
      doneTasks: 20,
      selfReviewTasks: [
        { n: 1, title: "Fix A", state: "done" },
        { n: 2, title: "Fix B", state: "pending" },
      ],
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
  "stage-leaderboard": { ...envelope("stage-leaderboard", leaderboardRows(LEADERBOARD_CURRENT)), projects: LEADERBOARD_PROJECTS },
  trend: envelope("trend", TREND_ROWS),
  "weekly-costs": { ...envelope("weekly-costs", WEEKLY_ROWS), projects: WEEKLY_PROJECTS },
  "daily-costs": envelope("daily-costs", []),
  "weekly-model-effort-costs": envelope("weekly-model-effort-costs", WEEKLY_MODEL_ROWS),
  reviewers: { ...envelope("reviewers", REVIEWER_ROWS), projects: REVIEWER_PROJECTS },
  "guard-health": envelope("guard-health", [
    gh("check-plan-shape", 6794, 3403, 37, 102, 3798824, "2026-11-04T10:00:00Z"),
    gh("check-references", 2246, 110, 0, 618, 2056971, "2026-11-04T09:00:00Z"),
    gh("check-no-bash", 35, 0, 0, 95, 3421, null),
  ]),
  "self-review": envelope("self-review", [
    { id: 1, recordedAt: "2026-01-12T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-fix", note: "[theme: guard-false-pass] guard misses a case", disposition: "fixed", ref: "abc1234", blastRadius: 3, severity: "important", detail: "## Why\nthe guard skips empty input", reviewStatus: "", reason: "", theme: "guard-false-pass", effort: "", commitUrl: "https://github.com/o/r/commit/abc1234" },
    { id: 2, recordedAt: "2026-01-11T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-cost", note: "cheaper reviewer", disposition: "fixed", ref: "def5678", blastRadius: null, severity: "minor", detail: "", reviewStatus: "", commitUrl: "https://github.com/o/r/commit/def5678" },
    { id: 3, recordedAt: "2026-01-10T10:00:00Z", project: "agents", change: "kan-1", angle: "flow-speed", note: "[theme: guard-false-pass] cache the guard", disposition: "suggested", ref: "", blastRadius: null, severity: "minor", detail: "## Proposal\nCache the guard result.\n\n## What happened\nThe guard re-ran on every call.\n\n## Examples\n- task 3 — ran 40 times\n\n## Effort and risk\nEffort: S\nTouches: 1 file\nRisk: Low\nGuard: extends check-references\n\n## Done when\n- [ ] faster", reviewStatus: "to-review", reason: "", theme: "guard-false-pass", effort: "S", commitUrl: "" },
    { id: 4, recordedAt: "2026-01-09T10:00:00Z", project: "agents", change: "kan-2", angle: "flow-improvement", note: "split the brief", disposition: "suggested", ref: "", blastRadius: null, severity: "minor", detail: "## Why\nlong", reviewStatus: "declined", reason: "too noisy", theme: "", effort: "", commitUrl: "" },
    { id: 5, recordedAt: "2026-01-08T10:00:00Z", project: "agents", change: "kan-2", angle: "flow-improvement", note: "panel caught it early", disposition: "noted", ref: "", blastRadius: null, severity: "", detail: "", reviewStatus: "", commitUrl: "" },
  ]),
  "review-dynamics": envelope("review-dynamics", reviewDynamicsRows()),
};

// The board filters by Updated relative to now; every case runs at one
// pinned instant so which fixture rows fall inside 24h/7d is fixed.
const BOARD_NOW = new Date("2026-01-15T12:00:00Z");

// The stage leaderboard's request log and window-keyed answers.
const LB_NOW = BOARD_NOW.getTime();
const calls = () => fetchStatsViewMock.mock.calls.filter((c) => c[0] === "stage-leaderboard").map((c) => c[1] as { from: Date; to: Date; project?: string });
const span = (p: { from: Date; to: Date }) => [p.from.getTime(), p.to.getTime()];
// The mock answers by window: the current window ends LB_NOW, the previous one before it.
const answer = (current: StageLeaderboardRow[], previous: StageLeaderboardRow[]) =>
  fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures, p: { from: Date; to: Date }) =>
    Promise.resolve(
      view !== "stage-leaderboard"
        ? fixtures[view]
        : { ...envelope("stage-leaderboard", p.to.getTime() === LB_NOW ? current : previous), projects: LEADERBOARD_PROJECTS },
    ),
  );
const tile = (cls: string) => within(document.querySelector<HTMLElement>(`.${cls}`)!);

const answerLeaderboard = () => answer(leaderboardRows(LEADERBOARD_CURRENT), leaderboardRows(LEADERBOARD_PREVIOUS));

beforeEach(() => {
  vi.useFakeTimers({ toFake: ["Date"], now: BOARD_NOW });
  onTestFinished(() => {
    vi.useRealTimers();
  });
  fetchStatsViewMock.mockReset();
  fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures) => Promise.resolve(fixtures[view]));
  fetchProjectsMock.mockReset();
  fetchProjectsMock.mockResolvedValue([]);
});

describe("views render their fixture response's actual values", () => {
  it("state board filters by chips and state tiles", async () => {
    const user = userEvent.setup();
    render(<StateBoard />);
    const table = await screen.findByRole("table");
    const tile = (name: RegExp) => screen.getByRole("button", { name });
    const visible = (name: string) => within(table).queryByRole("link", { name }) !== null;
    const clear = () => user.click(screen.getByRole("button", { name: "Clear filters" }));
    // ...through past now: a change updated this instant is inside the half-open period.
    expect(fetchStatsViewMock.mock.calls[0][1].to.getTime()).toBeGreaterThan(BOARD_NOW.getTime());
    // Updated defaults to 7d: the fortnight-old change is hidden, and the
    // tiles count only what the other filters leave.
    expect(visible("kan-5-old")).toBe(false);
    expect(tile(/^Planning/)).toHaveTextContent("1");
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();
    await user.click(within(screen.getByRole("group", { name: "Updated" })).getByRole("button", { name: "All" }));
    expect(tile(/^Planning/)).toHaveTextContent("2");
    expect(visible("kan-5-old")).toBe(true);
    // A Project chip narrows rows and tiles.
    await user.click(within(screen.getByRole("group", { name: "Project" })).getByRole("button", { name: /^gymie/ }));
    expect(tile(/^Planning/)).toHaveTextContent("1");
    expect(tile(/^Integrating & review/)).toHaveTextContent("0");
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
    // One card per unfinished change; the one
    // awaiting an answer comes first and is marked as needing the operator.
    expect(cards).toHaveLength(3);
    expect(cards[0]).toHaveTextContent("kan-29-never-observed");
    expect(cards[0]).toHaveClass("needs");
    expect(cards[0]).toHaveTextContent("Awaiting your answer");
    expect(cards[0]).toHaveTextContent("Which layout?");
    expect(cards[0]).toHaveTextContent("Planning the change");
    expect(cards[1]).toHaveTextContent("kan-16-stats-app");
    expect(cards[1]).not.toHaveClass("needs");
    expect(cards[1]).not.toHaveTextContent("Working"); // the live dot already says it
    expect(cards[1]).toHaveTextContent("Implementing · 30m");
    expect(cards[1]).toHaveTextContent("tasks 20/46 · self-review tasks 1/2 · $12.50");
    // The card keeps agents, main's inline work included, to a count; the run view lists them.
    expect(within(cards[1]).getByTestId("agent-count")).toHaveTextContent("main + 2 subagents running");
    // An open stage alone is the main session, one agent; a change between stages has none.
    expect(within(cards[0]).getByTestId("agent-count")).toHaveTextContent("main session running");
    expect(within(cards[2]).getByTestId("agent-count")).toHaveTextContent("main session running");
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
    // Finished starts collapsed but counted on open; opening it shows its rows.
    expect(groups.map((g) => g.textContent)).toEqual(["▾Integrating & review1", "▾Planning & implementing1", "▾Finished1"]);
    expect(table.queryByRole("link", { name: "kan-31-replan-fold" })).not.toBeInTheDocument();
    await user.click(groups[2]);
    expect(await table.findByRole("link", { name: "kan-31-replan-fold" })).toBeInTheDocument();
    // A row: state pill, current stage, tasks with growth, relative update
    // with the ISO time as its title, cost; no errors column.
    const row = table.getByRole("link", { name: "kan-16-stats-app" }).closest("tr")!;
    expect(row).toHaveTextContent("Integrating");
    expect(row).toHaveTextContent("Implementing");
    expect(within(row).getByText("+24")).toBeInTheDocument();
    expect(within(row).getByTitle("2026-01-15T10:00:00.000Z")).toHaveTextContent("2h ago");
    expect(row).toHaveTextContent("$12.50");
    // Fix rounds and reviewer runs, main · self-review.
    expect(table.getByRole("columnheader", { name: "Fix rounds" })).toBeInTheDocument();
    expect(table.getByRole("columnheader", { name: "Reviewer runs" })).toBeInTheDocument();
    expect(within(row).getByTitle("main 5 · self-review 1")).toHaveTextContent("5 · 1");
    expect(within(row).getByTitle("main 14 · self-review 2")).toHaveTextContent("14 · 2");
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
    // It counts the finished changes, which load on open.
    await waitFor(() => expect(cost()).not.toBe("—"));
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

  it("state board polls only open changes, reading finished ones once per period chip", async () => {
    vi.useFakeTimers({ toFake: ["Date", "setInterval", "clearInterval"], now: BOARD_NOW });
    render(<StateBoard />);
    await act(async () => {});
    const periods = () => fetchStatsViewMock.mock.calls.map((c) => [(c[1] as { from: Date }).from.getTime(), (c[1] as { to: Date }).to.getTime()]);
    const DAY = 24 * 60 * 60 * 1000;
    // The open poll asks for an empty period -- the store adds every unfinished change -- and finished ones come from the 30d window.
    expect(periods()).toEqual(expect.arrayContaining([[BOARD_NOW.getTime() + DAY, BOARD_NOW.getTime() + DAY], [BOARD_NOW.getTime() - 30 * DAY, BOARD_NOW.getTime() + DAY]]));
    const before = periods().length;
    await act(async () => vi.advanceTimersByTime(5000));
    expect(periods().slice(before)).toEqual([[BOARD_NOW.getTime() + 5000 + DAY, BOARD_NOW.getTime() + 5000 + DAY]]);
    fireEvent.click(within(screen.getByRole("group", { name: "Updated" })).getByRole("button", { name: "All" }));
    await act(async () => {});
    expect(periods().at(-1)![0]).toBe(0);
  });

  it("state board rereads finished changes once a polled change leaves the open set", async () => {
    vi.useFakeTimers({ toFake: ["Date", "setInterval", "clearInterval"], now: BOARD_NOW });
    const rows = fixtures["state-board"].rows as { name: string; state: string }[];
    let open = rows.filter((r) => r.state !== "FINISHED");
    const isPoll = (p: { from: Date; to: Date }) => p.from.getTime() === p.to.getTime();
    fetchStatsViewMock.mockImplementation((_v: string, p: { from: Date; to: Date }) =>
      Promise.resolve({ ...fixtures["state-board"], rows: isPoll(p) ? open : rows }),
    );
    render(<StateBoard />);
    await act(async () => {});
    fireEvent.click(screen.getByRole("row", { name: /^▾Finished/ }));
    await act(async () => {});
    const reads = () => fetchStatsViewMock.mock.calls.filter((c) => !isPoll(c[1])).length;
    const first = reads();
    await act(async () => vi.advanceTimersByTime(5000));
    expect(reads()).toBe(first);
    open = open.slice(1);
    await act(async () => vi.advanceTimersByTime(5000));
    expect(reads()).toBe(first + 1);
  });

  it("state board renders the open changes without waiting for the finished ones", async () => {
    fetchStatsViewMock.mockImplementation((_v: string, p: { from: Date; to: Date }) =>
      p.from.getTime() === p.to.getTime() ? Promise.resolve(fixtures["state-board"]) : new Promise(() => {}),
    );
    render(<StateBoard />);
    expect(await screen.findAllByTestId("running-card")).not.toHaveLength(0);
    expect(screen.queryByText("Loading…")).not.toBeInTheDocument();
  });

  it("state board counts finished changes on open and names only projects with changes", async () => {
    fetchProjectsMock.mockResolvedValue(["tmp.junk"]);
    render(<StateBoard />);
    await screen.findByRole("table");
    const finished = screen.getByRole("button", { name: /^Finished/ });
    await waitFor(() => expect(finished).toHaveTextContent("Finished1"));
    const group = within(screen.getByRole("group", { name: "Project" }));
    expect(group.queryByRole("button", { name: /^tmp\.junk/ })).not.toBeInTheDocument();
  });

  it("state board shows a failed finished read", async () => {
    fetchStatsViewMock.mockImplementation((_v: string, p: { from: Date; to: Date }) =>
      p.from.getTime() === p.to.getTime() ? Promise.resolve(fixtures["state-board"]) : Promise.reject(new Error("boom")),
    );
    const user = userEvent.setup();
    render(<StateBoard />);
    const table = within(await screen.findByRole("table"));
    await user.click(table.getByRole("row", { name: /^▾Finished/ }));
    expect(await screen.findByText("Could not load the board: boom")).toBeInTheDocument();
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

  it("stage leaderboard chips drive the request and the client filter", async () => {
    answerLeaderboard();
    const user = userEvent.setup();
    render(<StageLeaderboard />);
    await screen.findByRole("table");
    const day = 24 * 60 * 60 * 1000;
    expect(calls().map(span)).toEqual([
      [LB_NOW - 30 * day, LB_NOW],
      [LB_NOW - 60 * day, LB_NOW - 30 * day],
    ]);
    expect(screen.queryByLabelText("Period start")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Model")).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/Search/)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();

    const periodGroup = within(screen.getByRole("group", { name: "Period" }));
    await user.click(periodGroup.getByRole("button", { name: "7d" }));
    await waitFor(() => expect(calls().slice(-2).map(span)).toEqual([
      [LB_NOW - 7 * day, LB_NOW],
      [LB_NOW - 14 * day, LB_NOW - 7 * day],
    ]));
    expect(await screen.findByRole("columnheader", { name: "vs previous 7d" })).toBeInTheDocument();

    const projectGroup = within(screen.getByRole("group", { name: "Project" }));
    await user.click(projectGroup.getByRole("button", { name: /^gymie/ }));
    await waitFor(() => expect(calls().slice(-2).map((p) => p.project)).toEqual(["gymie-7c1f238a", "gymie-7c1f238a"]));

    await screen.findByRole("table");
    const before = calls().length;
    const commandGroup = within(screen.getByRole("group", { name: "Command" }));
    await user.click(commandGroup.getByRole("button", { name: /^\/flow-fast/ }));
    const narrowed = within(screen.getByRole("table"));
    expect(narrowed.getAllByRole("row")).toHaveLength(4);
    expect(narrowed.getAllByRole("cell", { name: "/flow-fast" })).toHaveLength(3);
    expect(calls()).toHaveLength(before);

    await user.click(screen.getByRole("button", { name: "Clear filters" }));
    expect(periodGroup.getByRole("button", { name: "30d" })).toHaveAttribute("aria-pressed", "true");
    expect(projectGroup.getByRole("button", { name: "All" })).toHaveAttribute("aria-pressed", "true");
    expect(commandGroup.getByRole("button", { name: "All" })).toHaveAttribute("aria-pressed", "true");
    await waitFor(() => expect(calls().slice(-2).map((p) => [...span(p), p.project])).toEqual([
      [LB_NOW - 30 * day, LB_NOW, undefined],
      [LB_NOW - 60 * day, LB_NOW - 30 * day, undefined],
    ]));
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();
  });

  it("stage leaderboard asks for projects on its current-window call only", async () => {
    answerLeaderboard();
    render(<StageLeaderboard />);
    await screen.findByRole("table");
    expect(calls().map((p) => (p as { projects?: boolean }).projects)).toEqual([true, undefined]);
  });

  it("stage leaderboard Period chip keeps the Project and Command chips", async () => {
    answerLeaderboard();
    const user = userEvent.setup();
    render(<StageLeaderboard />);
    await screen.findByRole("table");
    const projectGroup = within(screen.getByRole("group", { name: "Project" }));
    await user.click(projectGroup.getByRole("button", { name: /^gymie/ }));
    await screen.findByRole("table");
    const commandGroup = within(screen.getByRole("group", { name: "Command" }));
    await user.click(commandGroup.getByRole("button", { name: /^\/flow-fast/ }));
    await user.click(within(screen.getByRole("group", { name: "Period" })).getByRole("button", { name: "90d" }));
    await screen.findByRole("table");
    expect(within(screen.getByRole("group", { name: "Project" })).getByRole("button", { name: /^gymie/ })).toHaveAttribute("aria-pressed", "true");
    expect(within(screen.getByRole("group", { name: "Command" })).getByRole("button", { name: /^\/flow-fast/ })).toHaveAttribute("aria-pressed", "true");
    expect(calls().at(-1)?.project).toBe("gymie-7c1f238a");
  });

  it("stage leaderboard Spend tile compares the chosen command's previous spend", async () => {
    answerLeaderboard();
    const user = userEvent.setup();
    render(<StageLeaderboard />);
    await screen.findByRole("table");
    await user.click(within(screen.getByRole("group", { name: "Command" })).getByRole("button", { name: /^\/flow-fast/ }));
    const spend = tile("sl-t-spend");
    spend.getByText("$105");
    expect(spend.getByText(/vs the previous 30d/)).toHaveTextContent("+$12.00 (+13%) vs the previous 30d");
  });

  it("stage leaderboard reads a sub-cent previous spend as new, writes a thousandfold rise in k, and drops the command note under a Command chip", async () => {
    answer(
      leaderboardRows([["/flow", "flow.sdd-tdd", 34, null], ["/flow", "flow.verify", 10, null], ["/flow-fast", "flow.cleanup", 2, null]]),
      leaderboardRows([["/flow", "flow.sdd-tdd", 0.002], ["/flow", "flow.verify", 0.5]]),
    );
    const user = userEvent.setup();
    render(<StageLeaderboard />);
    await screen.findByRole("table");
    const row = (name: string) => within([...document.querySelectorAll<HTMLElement>("tbody tr")].find((tr) => tr.textContent?.includes(name))!);
    row("Implementing").getByText("new");
    row("Verifying").getByText("+$9.50 (+1.9k%)");
    expect(document.querySelectorAll(".sl-name .sl-cmd").length).toBeGreaterThan(0);
    await user.click(within(screen.getByRole("group", { name: "Command" })).getByRole("button", { name: /^\/flow\$/ }));
    expect(document.querySelectorAll(".sl-name .sl-cmd")).toHaveLength(0);
  });

  it("stage leaderboard drops stages and projects under half a cent", async () => {
    const cur = leaderboardRows([...LEADERBOARD_CURRENT, ["/flow", "flow.zero", 0.001, null], ["/flow", "flow.half-cent", 0.005, null], ["/flow", "flow.fifth", 0.2, null]]);
    fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures, p: { from: Date; to: Date }) =>
      Promise.resolve(
        view !== "stage-leaderboard"
          ? fixtures[view]
          : {
              ...envelope("stage-leaderboard", p.to.getTime() === LB_NOW ? cur : leaderboardRows(LEADERBOARD_PREVIOUS)),
              projects: [...LEADERBOARD_PROJECTS, { projectKey: "idle-00000000", totalCostUsd: 0.001 }],
            },
      ),
    );
    render(<StageLeaderboard />);
    await screen.findByRole("table");
    // 0.001 is dropped; exactly half a cent and 0.2 are kept.
    screen.getByText("14 stages ran");
    const chips = within(screen.getByRole("group", { name: "Project" })).getAllByRole("button");
    expect(chips.map((c) => c.textContent)).toEqual(["All", "gymie$1,000", "agents$260"]);
  });

  it("stage leaderboard shows a failed previous-window request", async () => {
    fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures, p: { from: Date; to: Date }) =>
      view !== "stage-leaderboard"
        ? Promise.resolve(fixtures[view])
        : p.to.getTime() === LB_NOW
          ? Promise.resolve({ ...envelope("stage-leaderboard", leaderboardRows(LEADERBOARD_CURRENT)), projects: LEADERBOARD_PROJECTS })
          : Promise.reject(new Error("previous window failed")),
    );
    render(<StageLeaderboard />);
    expect(await screen.findByRole("alert")).toHaveTextContent("previous window failed");
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });

  it("stage leaderboard reads not-recorded and unmeasured as Panel does", async () => {
    fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures) =>
      Promise.resolve(view !== "stage-leaderboard" ? fixtures[view] : envelope("stage-leaderboard", leaderboardRows(LEADERBOARD_CURRENT), false)),
    );
    const { unmount } = render(<StageLeaderboard />);
    expect(await screen.findByTestId("not-recorded")).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    unmount();

    fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures) =>
      Promise.resolve(view !== "stage-leaderboard" ? fixtures[view] : envelope("stage-leaderboard", leaderboardRows(LEADERBOARD_CURRENT), true, true)),
    );
    render(<StageLeaderboard />);
    expect(await screen.findByTestId("unmeasured")).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });

  it("stage leaderboard project chips name projects with spend", async () => {
    answerLeaderboard();
    render(<StageLeaderboard />);
    await screen.findByRole("table");
    const chips = within(screen.getByRole("group", { name: "Project" })).getAllByRole("button");
    expect(chips.map((c) => c.textContent)).toEqual(["All", "gymie$1,000", "agents$260"]);
    expect(chips.map((c) => c.getAttribute("title"))).toEqual([null, "gymie-7c1f238a", "agents-a740d89c"]);
  });

  it("stage leaderboard tiles name stages and the biggest riser", async () => {
    answerLeaderboard();
    render(<StageLeaderboard />);
    await screen.findByRole("table");
    const spend = tile("sl-t-spend");
    spend.getByText("$1,260");
    expect(spend.getByText(/vs the previous 30d/)).toHaveTextContent("+$20.00 (+2%) vs the previous 30d");

    const top = tile("sl-t-top");
    top.getByText("79%");
    const items = top.getAllByRole("listitem");
    expect(items.map((li) => li.textContent)).toEqual([
      "1Implementing /flowflow.sdd-tdd$50040%",
      "2Reviewing /flowflow.review-panel$30024%",
      "3Self-review /flowflow.self-review$20016%",
    ]);

    const riser = tile("sl-t-riser");
    riser.getByText("+$180");
    riser.getByText("Self-review");
    riser.getByText("flow.self-review");
    riser.getByText("/flow");
    riser.getByText("$20.00 → $200 vs the previous 30d (+900%)");
  });

  it("stage leaderboard table ranks the top ten with deltas and an Other row", async () => {
    answerLeaderboard();
    render(<StageLeaderboard />);
    const table = within(await screen.findByRole("table"));
    expect(table.getAllByRole("columnheader").map((h) => h.textContent)).toEqual([
      "#", "Stage", "Command", "Spend", "Share", "vs previous 30d", "Runs", "Median run", "P90 run", "Main share",
    ]);
    expect(table.queryAllByRole("button")).toHaveLength(0);
    screen.getByText("12 stages ran");
    const rows = table.getAllByRole("row").slice(1);
    expect(rows).toHaveLength(11);
    expect(rows.map((r) => within(r).getAllByRole("cell")[0].textContent)).toEqual(["1", "2", "3", "4", "5", "6", "7", "8", "9", "10", ""]);

    const first = within(rows[0]);
    first.getByText("Implementing");
    first.getByText("flow.sdd-tdd");
    expect(first.getByText("+$100 (+25%)")).toHaveClass("up");
    first.getByRole("cell", { name: "40%" });
    first.getByRole("cell", { name: "46%" });
    first.getByRole("cell", { name: "$1.25" });
    first.getByRole("cell", { name: "$3.50" });
    expect(within(rows[1]).getByText("−$300 (−50%)")).toHaveClass("down");
    expect(within(rows[3]).getAllByText("SDD + TDD per task")).toHaveLength(1);
    within(rows[5]).getByText("new");

    const other = within(rows[10]);
    expect(rows[10]).toHaveClass("other");
    other.getByText("Other 2 stages");
    other.getByText("$5.00");
    other.getByRole("cell", { name: "0%" });
    other.getByText("+$2.00 (+67%)");
    other.getByRole("cell", { name: "41" });
    expect(other.getAllByRole("cell", { name: "—" })).toHaveLength(3);
  });

  it("stage leaderboard reports no riser and the empty state", async () => {
    const cur = leaderboardRows(LEADERBOARD_CURRENT);
    answer(cur, cur.map((r) => ({ ...r, totalCostUsd: r.totalCostUsd * 2 })));
    const { unmount } = render(<StageLeaderboard />);
    await screen.findByRole("table");
    const riser = tile("sl-t-riser");
    riser.getByText("—");
    riser.getByText("no stage grew vs the previous 30d");
    unmount();

    answer([], []);
    render(<StageLeaderboard />);
    expect(await screen.findByText("No costed stage runs in this period.")).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });
  it("reviewers ranks by value, severity and false alarms", async () => {
    const user = userEvent.setup();
    render(<Reviewers />);
    await screen.findByRole("table");
    const slots = () => [...document.querySelectorAll("tbody .rv-slot")].map((e) => e.firstChild?.textContent);
    const sorted = () => screen.getAllByRole("columnheader").filter((th) => th.hasAttribute("aria-sort"));
    const rank = (name: string) => user.click(within(screen.getByRole("group", { name: "Rank by" })).getByRole("button", { name }));

    expect(slots()).toEqual(["beta", "able", "alpha", "gamma", "eps", "exp-delta"]);
    screen.getByText("Every slot, best value first");
    expect(sorted()).toHaveLength(1);
    expect(sorted()[0]).toHaveTextContent("$ per weighted finding");
    expect(sorted()[0]).toHaveAttribute("aria-sort", "ascending");

    await rank("Severity");
    expect(slots()).toEqual(["beta", "able", "alpha", "eps", "gamma", "exp-delta"]);
    screen.getByText("Every slot, most severe first");
    expect(sorted()).toHaveLength(1);
    expect(sorted()[0]).toHaveTextContent("Weighted per dispatch");
    expect(sorted()[0]).toHaveAttribute("aria-sort", "descending");

    await rank("False alarms");
    expect(slots()).toEqual(["alpha", "gamma", "able", "beta", "eps", "exp-delta"]);
    screen.getByText("Every slot, most false alarms first");
    expect(sorted()).toHaveLength(1);
    expect(sorted()[0]).toHaveTextContent("Withdrawn");
    expect(sorted()[0]).toHaveAttribute("aria-sort", "descending");

    const row = (slot: string) => within(screen.getByText(slot, { selector: ".rv-slot > :first-child" }).closest("tr")!);
    expect(row("exp-delta").getByRole("img", { name: "experimental reviewer" })).toHaveAttribute("title", "D desc");
    const eps = row("eps").getAllByRole("cell");
    expect(eps[3]).toHaveTextContent(/^—$/);
    expect(eps[4]).toHaveTextContent(/^—$/);
    expect(row("exp-delta").getAllByRole("cell")[4]).toHaveTextContent("found nothing");
    expect(row("gamma").getAllByRole("cell")[9]).toHaveTextContent("1 33%");
    expect(row("beta").getAllByRole("cell")[4]).toHaveClass("rv-best");
    expect(row("gamma").getAllByRole("cell")[4]).toHaveClass("rv-worst");
  });

  it("reviewers names every slot tied on most critical, each", async () => {
    fetchStatsViewMock.mockResolvedValueOnce({ ...envelope("reviewers", [rv("one", 2, 2, 1, 2, 0, 0, 0), rv("two", 2, 2, 1, 2, 0, 0, 0)]), projects: [] });
    render(<Reviewers />);
    await screen.findByRole("table");
    expect(document.querySelectorAll(".rv-tiles .lb-tile")[1]).toHaveTextContent("most critical: one, two (2 each)");
  });

  it("reviewers shows no tiles when the rows are not recorded or unmeasured", async () => {
    for (const [recorded, unmeasured] of [[false, false], [true, true]]) {
      fetchStatsViewMock.mockResolvedValueOnce({ ...envelope("reviewers", REVIEWER_ROWS, recorded, unmeasured), projects: [] });
      const { unmount } = render(<Reviewers />);
      await waitFor(() => expect(fetchStatsViewMock).toHaveBeenCalled());
      await screen.findByRole("toolbar");
      await waitFor(() => expect(screen.queryByText(/loading/i)).not.toBeInTheDocument());
      expect(document.querySelector(".rv-tiles")).toBeNull();
      unmount();
      fetchStatsViewMock.mockClear();
    }
  });

  it("reviewers tiles and footer derive from the rows", async () => {
    render(<Reviewers />);
    await screen.findByRole("table");
    const [cost, severity, alarms] = [...document.querySelectorAll<HTMLElement>(".rv-tiles .lb-tile")];
    expect(cost).toHaveTextContent("Cost · best value$0.08beta per weighted findingleast: gamma $1.00found nothing: exp-delta ($0.60)");
    expect(cost.querySelector(".lb-val")).toHaveClass("rv-best");
    expect(severity).toHaveTextContent("Severity · most found2.50beta weighted per dispatchmost critical: beta (1)");
    expect(alarms).toHaveTextContent("False alarms · most withdrawn50%alpha 5 of 10 findingsfewest: eps 0%");
    expect(alarms.querySelector(".lb-val")).toHaveClass("rv-worst");
    const foot = [...document.querySelectorAll("tfoot td")].map((td) => td.textContent);
    expect(foot).toEqual(["All slots", "26", "$9.00 21 of 26 priced", "$0.27", "1", "6", "7", "1.58", "6 30%"]);
  });

  it("reviewers chips request the window and project, and Clear keeps the rank", async () => {
    const user = userEvent.setup();
    const reqs = () => fetchStatsViewMock.mock.calls.filter((c) => c[0] === "reviewers").map((c) => c[1] as { from: Date; to: Date; project?: string; projects?: boolean });
    const day = 24 * 60 * 60 * 1000;
    const now = BOARD_NOW.getTime();
    const { unmount } = render(<Reviewers />);
    await screen.findByRole("table");
    expect(reqs()).toHaveLength(1);
    expect(reqs()[0]).toMatchObject({ projects: true, project: undefined });
    expect([reqs()[0].from.getTime(), reqs()[0].to.getTime()]).toEqual([now - 30 * day, now]);
    expect(screen.queryByLabelText("Period start")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();

    const group = (name: string) => within(screen.getByRole("group", { name }));
    const pressed = (g: string, name: string | RegExp) => expect(group(g).getByRole("button", { name })).toHaveAttribute("aria-pressed", "true");
    await user.click(group("Rank by").getByRole("button", { name: "Severity" }));
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();
    await user.click(group("Period").getByRole("button", { name: "7d" }));
    await waitFor(() => expect([reqs().at(-1)!.from.getTime(), reqs().at(-1)!.to.getTime()]).toEqual([now - 7 * day, now]));
    await screen.findByRole("table");
    pressed("Rank by", "Severity");
    await user.click(group("Project").getByRole("button", { name: /^gymie/ }));
    await waitFor(() => expect(reqs().at(-1)).toMatchObject({ project: "gymie-7c1f238a", projects: true }));
    await screen.findByRole("table");
    pressed("Rank by", "Severity");
    pressed("Period", "7d");
    expect(group("Project").getByRole("button", { name: /^gymie/ })).toHaveAttribute("title", "gymie-7c1f238a");
    expect(group("Project").getByRole("button", { name: /^gymie/ })).toHaveTextContent("gymie20");

    await user.click(screen.getByRole("button", { name: "Clear filters" }));
    pressed("Period", "30d");
    pressed("Project", "All");
    pressed("Rank by", "Severity");
    await waitFor(() => expect(reqs().at(-1)).toMatchObject({ project: undefined }));
    expect([reqs().at(-1)!.from.getTime(), reqs().at(-1)!.to.getTime()]).toEqual([now - 30 * day, now]);
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();
    unmount();

    fetchStatsViewMock.mockResolvedValueOnce({ ...envelope("reviewers", []), projects: [] });
    render(<Reviewers />);
    expect(await screen.findByText("No reviewer dispatches in this period.")).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(document.querySelector(".rv-tiles")).toBeNull();
  });

  // The server's lists over the fixture rows: `show`/`status` pick the list, newest first.
  const servePages = (rows: SelfReviewRow[]) =>
    pageMock.mockImplementation(async (q: { show: string; status?: string; limit: number; offset: number }) => {
      const list = rows.filter((r) => (q.show === "fixed" ? r.disposition === "fixed" : r.reviewStatus === q.status));
      const n = (s: string) => rows.filter((r) => r.reviewStatus === s).length;
      return { total: list.length, toReview: n("to-review"), approved: n("approved"), declined: n("declined"), limit: q.limit, offset: q.offset, rows: list.slice(q.offset, q.offset + q.limit) };
    });
  const fixtureRows = () => (fixtures["self-review"].rows as unknown as SelfReviewRow[]);

  it("self-review: opens on To review with the structured detail; approve sends the reason", async () => {
    servePages(fixtureRows());
    const user = userEvent.setup();
    render(<SelfReview />);

    // The open suggestion, theme split off its title, its sections laid out, the theme seen twice (one fixed row shares it).
    expect(await screen.findByRole("heading", { name: "cache the guard" })).toBeInTheDocument();
    expect(screen.getByText("Cache the guard result.")).toBeInTheDocument();
    expect(screen.getByText("The guard re-ran on every call.")).toBeInTheDocument();
    expect(screen.getByText("2× (theme)")).toBeInTheDocument();
    expect(screen.getByText("extends check-references")).toBeInTheDocument(); // a guard finding says new or an extension
    expect(screen.getByText("guard misses a case", { exact: false })).toBeInTheDocument(); // same theme before
    expect(screen.queryByRole("link", { name: "abc1234" })).not.toBeInTheDocument();

    // The identifier an agent is given, and the copyable line that carries it.
    expect(screen.getByText("kan-1#3")).toBeInTheDocument();
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", { value: { writeText }, configurable: true });
    await user.click(screen.getByRole("button", { name: "Copy for agent" }));
    expect(writeText).toHaveBeenCalledWith(expect.stringContaining("kan-1#3"));
    expect(writeText).toHaveBeenCalledWith(expect.stringContaining("flow self-review finding-get 3"));

    await user.type(screen.getByLabelText("Reason"), "worth it");
    await user.click(screen.getByRole("button", { name: "Approve" }));
    expect(reviewMock).toHaveBeenCalledWith("agents", 3, "approved", "worth it");
  });

  it("self-review: Auto-fixed is read only; Declined shows its reason and moves back to review", async () => {
    servePages(fixtureRows());
    const user = userEvent.setup();
    render(<SelfReview />);
    await user.click(await screen.findByRole("button", { name: /^Auto-fixed/ }));
    expect(await screen.findByRole("link", { name: "abc1234" })).toHaveAttribute("href", "https://github.com/o/r/commit/abc1234");
    expect(screen.queryByRole("button", { name: "Approve" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /^Declined/ }));
    expect(await screen.findByText("Reason: too noisy")).toBeInTheDocument();
    expect(document.querySelector(".sr-raw")).toHaveTextContent("## Why long"); // an older free-form detail stays raw
    await user.click(screen.getByRole("button", { name: "Move back to review" }));
    expect(reviewMock).toHaveBeenCalledWith("agents", 4, "to-review", "");
  });

  it("self-review: every list pages 25 rows at a time", async () => {
    const many = Array.from({ length: 30 }, (_, i) => ({ ...fixtureRows()[2], id: 100 + i, note: `suggestion number ${i}` }));
    servePages(many);
    const user = userEvent.setup();
    render(<SelfReview />);
    expect(await screen.findByRole("button", { name: /suggestion number 0(?!\d)/ })).toBeInTheDocument();
    expect(pageMock).toHaveBeenCalledWith(expect.objectContaining({ show: "to-review", status: "to-review", limit: 25, offset: 0 }));
    expect(screen.queryByRole("button", { name: /suggestion number 25(?!\d)/ })).not.toBeInTheDocument();
    expect(screen.getByText("1–25 of 30")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Next" }));
    expect(await screen.findByRole("button", { name: /suggestion number 25(?!\d)/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /suggestion number 0(?!\d)/ })).not.toBeInTheDocument();
    expect(screen.getByText("26–30 of 30")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Next" })).toBeDisabled();
  });

  it("confidence: lists the estimates to review, 25 a page, and approves, declines and moves one back", async () => {
    const rows: ConfidenceRow[] = [
      { id: 2, project: "agents", change: "kan-9", stage: "implement", task: "2+3", score: 85, scoreAfter: 70, gapsAfter: ["a test pinned the wrong order"], gaps: ["the fixture is missing", "task 3 states no order"], tags: ["missing-fixture"], recordedAt: "2026-11-20T10:00:00Z", reviewStatus: "to-review" },
      { id: 1, project: "agents", change: "kan-9", stage: "plan", score: 70, gaps: ["closes with: operator — column order"], tags: ["open-decision"], disposition: "later", recordedAt: "2026-11-19T10:00:00Z", reviewStatus: "to-review" },
      ...Array.from({ length: 28 }, (_, i) => ({ id: 10 + i, project: "agents", change: `kan-${100 + i}`, stage: "implement" as const, task: "1", score: 90, gaps: ["g"], tags: ["missing-fixture"], recordedAt: "2026-11-01T10:00:00Z", reviewStatus: "to-review" as const })),
    ];
    confidenceMock.mockImplementation(async (q: { status: string; limit: number; offset: number }) => {
      const count = (s: string) => rows.filter((r) => r.reviewStatus === s).length;
      const list = rows.filter((r) => r.reviewStatus === q.status);
      return { total: list.length, toReview: count("to-review"), approved: count("approved"), declined: count("declined"), limit: q.limit, offset: q.offset, rows: list.slice(q.offset, q.offset + q.limit) };
    });
    confidenceReviewMock.mockImplementation(async (_project: string, id: number, status: ConfidenceRow["reviewStatus"], reason: string) => {
      const row = rows.find((r) => r.id === id)!;
      row.reviewStatus = status;
      row.reason = reason || undefined;
    });
    const user = userEvent.setup();
    render(<Confidence />);

    // To review is the list on screen; the newest estimate is open: its gaps in full, its tag.
    expect(await screen.findByRole("heading", { name: "85% → 70% · kan-9" })).toBeInTheDocument();
    expect(screen.getByText("the fixture is missing")).toBeInTheDocument();
    expect(screen.getByText("task 3 states no order")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Still open once its tests were green" })).toBeInTheDocument();
    expect(screen.getByText("a test pinned the wrong order")).toBeInTheDocument();
    expect(confidenceMock).toHaveBeenCalledWith({ status: "to-review", limit: 25, offset: 0 });
    expect(screen.getByText("1–25 of 30")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "To review 30" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getAllByText("missing-fixture").length).toBeGreaterThan(0);
    // Tags are data on a row, never a filter.
    expect(screen.queryByRole("button", { name: /^missing-fixture/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^All/ })).not.toBeInTheDocument();

    // The estimate copies whole, for an agent.
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", { value: { writeText }, configurable: true });
    await user.click(screen.getByRole("button", { name: "Copy for agent" }));
    expect(writeText).toHaveBeenCalledWith(expect.stringContaining("85%"));
    expect(writeText).toHaveBeenCalledWith(expect.stringContaining("- the fixture is missing"));
    expect(writeText).toHaveBeenCalledWith(expect.stringMatching(/Analyze what can be done and where, and say what generic fix would prevent this kind of gap in any future task — a fix specific to this task is useless\.$/));

    // Declining with a reason stores both and takes the row off the list.
    await user.type(screen.getByLabelText("Reason"), " not a real gap ");
    await user.click(screen.getByRole("button", { name: "Decline" }));
    expect(confidenceReviewMock).toHaveBeenCalledWith("agents", 2, "declined", "not a real gap");
    expect(await screen.findByRole("heading", { name: "70% · kan-9" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "To review 29" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Declined 1" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Approve" }));
    expect(confidenceReviewMock).toHaveBeenLastCalledWith("agents", 1, "approved", "");
    expect(await screen.findByRole("button", { name: "Approved 1" })).toBeInTheDocument();

    // The Declined chip reads that list: the reason, and the way back.
    await user.click(screen.getByRole("button", { name: "Declined 1" }));
    expect(await screen.findByText("Reason: not a real gap")).toBeInTheDocument();
    expect(confidenceMock).toHaveBeenLastCalledWith({ status: "declined", limit: 25, offset: 0 });
    expect(screen.queryByRole("button", { name: "Approve" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Move back to review" }));
    expect(confidenceReviewMock).toHaveBeenLastCalledWith("agents", 2, "to-review", "");
    expect(await screen.findByText("Nothing here.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "To review 29" })).toBeInTheDocument();
  });

  it("confidence: splits the gaps into those needing a decision and those closing by running", async () => {
    const gaps = [
      "**F9** [finding] tasks.md:41 — closes with: plan — Task 4 relies on a flag no task checks",
      "Lowered by F3 — the order is unstated",
      "a free-text gap",
      "**F11** [advisory] tasks.md:58 — closes with: running it — cold cache deadline — probe: make bench → under 2s; on failure: cache warm-up",
      "**F12** [finding] tasks.md:60 — closes with: running it — probe: make x → ok; on failure: none",
    ];
    confidenceMock.mockResolvedValue({ total: 1, toReview: 1, approved: 0, declined: 0, limit: 25, offset: 0, rows: [{ id: 1, project: "agents", change: "kan-7", stage: "plan", score: 80, gaps, tags: ["t"], recordedAt: "2026-11-20T10:00:00Z", reviewStatus: "to-review" }] });
    const user = userEvent.setup();
    render(<Confidence />);
    const decide = (await screen.findByRole("heading", { name: "Needs a decision" })).nextElementSibling as HTMLElement;
    const running = screen.getByRole("heading", { name: "Closes by running" }).nextElementSibling as HTMLElement;
    expect(decide.textContent).toContain("Task 4 relies on a flag");
    expect(decide.textContent).toContain("Lowered by F3");
    expect(decide.textContent).toContain("a free-text gap");
    expect(decide.textContent).not.toContain("cold cache");
    expect(running.textContent).toContain("cold cache");
    expect(running.textContent).toContain("F12");
    expect(running.textContent).not.toContain("Task 4");
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", { value: { writeText }, configurable: true });
    await user.click(screen.getByRole("button", { name: "Copy for agent" }));
    expect(writeText).toHaveBeenCalledWith(expect.stringMatching(/Needs a decision:\n- \*\*F9\*\*[\s\S]*Closes by running:\n- \*\*F11\*\*/));
  });

  it("confidence: no rows says so", async () => {
    confidenceMock.mockResolvedValue({ total: 0, toReview: 0, approved: 0, declined: 0, limit: 25, offset: 0, rows: [] });
    render(<Confidence />);
    expect(await screen.findByText("Nothing here.")).toBeInTheDocument();
  });

  it("self-review: no rows says so", async () => {
    servePages([]);
    render(<SelfReview />);
    expect(await screen.findByText("Nothing here.")).toBeInTheDocument();
  });
});

describe("interactive session card", () => {
  const minutesAgo = (m: number) => new Date(BOARD_NOW.getTime() - m * 60000).toISOString();
  const base = { updatedBy: "flow", nextCommand: "/flow", runCount: 1, cost24h: 0, cost7d: 0, cost30d: 0, wallClockMs: 0, errors: 0, updatedAt: minutesAgo(1) };
  const running = {
    ...base,
    projectKey: "agents",
    name: "kan-1072-interactive-session-2026-10-30",
    state: "STARTED",
    command: "/flow-interactive",
    costUsd: 4.12,
    currentStage: "flow.sdd-tdd",
    currentStageStartedAt: minutesAgo(3),
    lastStage: "flow.sdd-tdd",
    activity: "working",
    activityDetail: "Edit: stats/web/src/lib/phase.ts",
    tasks: [
      { n: 1, title: "Why no effort in stage runs?", state: "done", outcome: "landed", commit: "a41c9e2" },
      { n: 2, title: 'Rename "Running now" to "Active runs"', state: "done", outcome: "landed", commit: "0b7d113" },
      { n: 3, title: 'What does "Errors" mean?', state: "done", outcome: "answered" },
      { n: 4, title: "Count cache reads in input tokens", state: "done", outcome: "landed", commit: "9f3c2ab" },
      { n: 5, title: "Move the cost rail to the left", state: "done", outcome: "dropped" },
      { n: 6, title: "Show the task list and the running agents", state: "done", outcome: "landed", commit: "c18a6d0", reason: "new store fields and two panels" },
      { n: 7, title: "Show the stage's model on the Agents row", state: "in progress" },
      { n: 8, title: "Remove the tool errors column", state: "pending" },
    ],
  };
  const waiting = {
    ...base,
    projectKey: "gymie",
    name: "kan-1073-interactive-session-2026-10-30",
    state: "STARTED",
    command: "/flow-interactive",
    costUsd: 1.87,
    lastStage: "flow.visual-verify",
    lastStageEndedAt: minutesAgo(6),
    activity: "idle",
    activitySince: minutesAgo(6),
    tasks: [
      { n: 1, title: "Show the streak", state: "done", outcome: "landed", commit: "1111111" },
      { n: 2, title: "Widen the chart", state: "done", outcome: "landed", commit: "2222222" },
      { n: 3, title: "Hide the empty streak row", state: "done", outcome: "landed", commit: "3333333" },
    ],
  };
  const flow = {
    ...base,
    projectKey: "agents",
    name: "kan-1068-flow-plan-gate-summary",
    state: "STARTED",
    command: "/flow",
    costUsd: 9.4,
    currentStage: "flow.sdd-tdd",
    currentStageStartedAt: minutesAgo(12),
    tasks: [
      { n: 1, title: "One", state: "done" },
      { n: 2, title: "Two", state: "done" },
      { n: 3, title: "Print the gate summary before the decision block", state: "pending" },
      { n: 4, title: "Four", state: "pending" },
    ],
  };

  it("interactive session card reads as the prototype draws it", async () => {
    fetchStatsViewMock.mockImplementation(() => Promise.resolve(envelope("state-board", [running, waiting, flow])));
    render(<StateBoard />);
    const cards = await screen.findAllByTestId("running-card");
    const card = (name: string) => cards.find((c) => c.getAttribute("aria-label") === name)!;

    const r = card(running.name);
    expect(within(r).getByText("Interactive")).toHaveClass("lb-pill");
    expect(within(r).getByTestId("running-task")).toHaveTextContent("Request 7: Show the stage's model on the Agents row");
    expect(r).toHaveTextContent("Implementing · 3m · Edit: stats/web/src/lib/phase.ts");
    expect(r).toHaveTextContent("requests 6/8 · $4.12");

    const w = card(waiting.name);
    expect(within(w).getByText("Interactive")).toHaveClass("lb-pill");
    // Two spans, as the prototype draws them: the flex gap spaces them.
    expect(within(w).getByText("Waiting for the next request").nextSibling).toHaveTextContent("· 6m");
    expect(w.querySelector(".lb-ring")).not.toBeNull();
    expect(within(w).getByTestId("running-task")).toHaveTextContent("Last: Request 3 landed — Hide the empty streak row");
    expect(within(w).getByTestId("running-task")).toHaveClass("lb-muted");
    expect(w).toHaveTextContent("requests 3/3 · $1.87");
    expect(w).not.toHaveClass("needs");
    expect(w).toHaveClass("awaiting");
    expect(w.querySelector(".lb-act")).toBeNull();
    expect(within(w).queryByTestId("agent-count")).toBeNull();
    expect(w).not.toHaveTextContent("Paused after");

    const f = card(flow.name);
    expect(f).toHaveTextContent("tasks 2/4 · $9.40");
    expect(within(f).getByTitle("state STARTED")).toHaveTextContent("Implementing");
    expect(within(f).getByTitle("state STARTED")).toHaveClass("lb-pill");
    expect(f).not.toHaveTextContent("requests");
  });

  it("an interactive session whose harness session ended reads closed, not waiting", async () => {
    fetchStatsViewMock.mockImplementation(() => Promise.resolve(envelope("state-board", [{ ...waiting, activity: "ended" }])));
    render(<StateBoard />);
    const w = await screen.findByTestId("running-card");
    expect(within(w).getByText("Session closed").nextSibling).toHaveTextContent("· 6m");
    expect(w).not.toHaveTextContent("Waiting for the next request");
    expect(w).not.toHaveClass("awaiting");
  });

  it("a /flow change whose planning ended asks for /clear and the implementing run, and needs the operator", async () => {
    const planned = { ...flow, currentStage: undefined, currentStageStartedAt: undefined, lastStage: "flow.plan", lastStageEndedAt: minutesAgo(4), activity: "idle", activitySince: minutesAgo(4), sessionSeenAt: minutesAgo(3) };
    fetchStatsViewMock.mockImplementation(() => Promise.resolve(envelope("state-board", [planned])));
    render(<StateBoard />);
    const c = await screen.findByTestId("running-card");
    expect(c).toHaveTextContent(`Planned — /clear, then /flow ${flow.name}`);
    expect(c).not.toHaveTextContent("Paused after");
    expect(c).toHaveClass("needs");
  });
});

describe("weekly costs", () => {
  // A Friday: current week 2026-06-17, day 3 of 7; last full week 2026-06-10, compared with 2026-06-03.
  const WC_NOW = new Date("2026-06-19T12:00:00Z");
  const wcCalls = (view: string) =>
    fetchStatsViewMock.mock.calls.filter((c) => c[0] === view).map((c) => c[1] as { from: Date; to: Date; project?: string; projects?: boolean });
  const group = (name: string) => within(screen.getByRole("group", { name }));
  const chip = (name: string, label: string) => group(name).getByRole("button", { name: label });
  const pressedIn = (name: string) => group(name).getAllByRole("button", { pressed: true }).map((b) => b.textContent);
  const tileOf = (prefix: string) =>
    within([...document.querySelectorAll<HTMLElement>(".wc .lb-tile")].find((t) => t.querySelector(".lb-lbl")?.textContent?.startsWith(prefix))!);
  const segments = (week: string) => screen.getAllByTestId("wc-segment").filter((s) => s.getAttribute("data-week") === week);
  const texts = (id: string) => screen.getAllByTestId(id).map((e) => e.textContent);
  const legend = () => [...document.querySelectorAll(".wc-legend > span")].map((e) => e.textContent);
  const renderWeekly = async () => {
    vi.useFakeTimers({ now: WC_NOW, toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    render(<WeeklyCosts />);
    await screen.findByText("What moved, week of 2026-06-10 vs 2026-06-03");
  };

  it("weekly costs opens on 8w, Phase and All and requests the phase rows with projects", async () => {
    await renderWeekly();
    expect(screen.getByRole("heading", { name: "Weekly costs" })).toBeInTheDocument();
    screen.getByText("What did flow spend this week, and what moved it from last week?");
    const bar = within(screen.getByRole("toolbar", { name: "Filters" }));
    const labels = (name: string) => within(bar.getByRole("group", { name })).getAllByRole("button").map((b) => b.textContent);
    expect(labels("Period")).toEqual(["8w", "12w", "26w"]);
    expect(labels("Split by")).toEqual(["Phase", "Model · effort"]);
    expect(labels("Project")).toEqual(["All", "agents", "gymie"]);
    expect(pressedIn("Period")).toEqual(["8w"]);
    expect(pressedIn("Split by")).toEqual(["Phase"]);
    expect(pressedIn("Project")).toEqual(["All"]);
    const calls = wcCalls("weekly-costs");
    expect(calls).toHaveLength(1);
    expect(calls[0].projects).toBe(true);
    expect(calls[0].project).toBeUndefined();
    expect(calls[0].from).toEqual(new Date("2026-04-29T00:00:00Z"));
    expect(wcCalls("weekly-model-effort-costs")).toHaveLength(0);
  });

  it("weekly costs stacks a bar per week with its total, the partial week so far and empty weeks as no data", async () => {
    await renderWeekly();
    screen.getByText("Spend per week by phase");
    expect(texts("wc-week")).toEqual(["04-29", "05-06", "05-13", "05-20", "05-27", "06-03", "06-10", "06-17 (so far)"]);
    expect(screen.getAllByTestId("wc-total").map((t) => [t.getAttribute("data-week"), t.textContent])).toEqual([
      ["2026-06-03", "$75.00"],
      ["2026-06-10", "$92.00"],
      ["2026-06-17", "$12.00"],
    ]);
    expect(screen.getAllByText("no data")).toHaveLength(5);
    expect(texts("wc-total")).not.toContain("$0.00");
    for (const s of segments("2026-06-17")) {
      expect(s).toHaveAttribute("fill-opacity", ".45");
      expect(s).toHaveAttribute("stroke-dasharray", "3 2");
    }
    for (const s of segments("2026-06-10")) expect(s).not.toHaveAttribute("stroke-dasharray");
    expect(segments("2026-06-17").map((s) => s.getAttribute("data-series"))).toEqual(["Planning", "Implementation", "Visual verify"]);
    expect(legend()).toEqual(["Planning", "Implementation", "Review", "Self-review", "Visual verify", "Other"]);
  });

  it("weekly costs tiles compare the last full week with the week before and project this week's pace", async () => {
    await renderWeekly();
    const last = tileOf("Last full week");
    last.getByText("Last full week · 2026-06-10");
    last.getByText("$92.00");
    expect(last.getByText("▲ +$17.00 (+23%)")).toHaveClass("up");
    expect(last.getByText(/vs week of/)).toHaveTextContent("▲ +$17.00 (+23%) vs week of 2026-06-03");
    const now = tileOf("This week so far");
    now.getByText("This week so far · day 3 of 7");
    now.getByText("$12.00");
    now.getByText("on pace for $28.00");
    const avg = tileOf("Weekly average");
    avg.getByText("Weekly average · 8w");
    // Averaging over every full week would read $23.86.
    avg.getByText("$83.50");
    avg.getByText("full weeks only");
    const mover = tileOf("Biggest mover");
    mover.getByText("Planning");
    expect(mover.getByText("▲ +$20.00 (+200%)")).toHaveClass("up");
  });

  it("weekly costs What moved sorts by absolute change and leaves out a series at zero in both weeks", async () => {
    await renderWeekly();
    const table = screen.getByRole("table");
    expect(within(table).getAllByRole("columnheader").map((h) => h.textContent)).toEqual([
      "Phase",
      "2026-06-10",
      "2026-06-03",
      "Change",
      "Share of last week",
    ]);
    const rows = within(table).getAllByRole("row").slice(1);
    const cells = rows.map((r) => within(r).getAllByRole("cell").map((c) => c.textContent));
    expect(cells.map((c) => c[0])).toEqual(["Planning", "Implementation", "Self-review", "Review", "Other"]);
    expect(cells[0]).toEqual(["Planning", "$30.00", "$10.00", "▲ +$20.00 (+200%)", "33%"]);
    expect(cells[1][3]).toBe("▼ −$15.00 (−38%)");
    expect(within(rows[1]).getByText("▼ −$15.00 (−38%)")).toHaveClass("down");
    expect(cells[2][3]).toBe("▲ +$12.00");
    expect(cells[3][3]).toBe("—");
    expect(cells[4][3]).toBe("—");
    expect(rows[0].querySelector(".wc-sharebar > span")).toHaveStyle({ width: `${(30 / 92) * 100}%` });
  });

  it("weekly costs Model · effort split requests its rows and folds all but the six costliest pairs into Other models", async () => {
    const user = userEvent.setup();
    await renderWeekly();
    await user.click(chip("Split by", "Model · effort"));
    await screen.findByText("Spend per week by model · effort");
    const calls = wcCalls("weekly-model-effort-costs");
    expect(calls).toHaveLength(1);
    expect(calls[0].from).toEqual(wcCalls("weekly-costs")[0].from);
    expect(calls[0].to).toEqual(wcCalls("weekly-costs")[0].to);
    expect(legend()).toEqual([
      "opus-5-5 · session",
      "sonnet-5-5 · medium",
      "opus-5-5 · medium",
      "opus-5 · high",
      "opus-5-5 · low",
      "opus-5 · low",
      "Other models",
    ]);
    expect(screen.getAllByRole("columnheader")[0]).toHaveTextContent("Model · effort");
    await user.hover(segments("2026-06-03")[0]);
    const tip = within(screen.getByRole("tooltip"));
    expect(tip.getByText("Other models").closest("tr")).toHaveTextContent("$3.00");
  });

  it("weekly costs Project chip narrows both requests and keeps the Period and Split chips", async () => {
    const user = userEvent.setup();
    await renderWeekly();
    await user.click(chip("Period", "12w"));
    await user.click(chip("Split by", "Model · effort"));
    await user.click(await screen.findByRole("button", { name: "gymie" }));
    await waitFor(() => expect(wcCalls("weekly-costs").at(-1)!.project).toBe("gymie-89abcdef"));
    const costs = wcCalls("weekly-costs").at(-1)!;
    const model = wcCalls("weekly-model-effort-costs").at(-1)!;
    expect(costs.from).toEqual(new Date("2026-04-01T00:00:00Z"));
    expect(model.project).toBe("gymie-89abcdef");
    expect(model.from).toEqual(new Date("2026-04-01T00:00:00Z"));
    await screen.findByText("Spend per week by model · effort");
    expect(pressedIn("Period")).toEqual(["12w"]);
    expect(pressedIn("Split by")).toEqual(["Model · effort"]);
    expect(pressedIn("Project")).toEqual(["gymie"]);
  });

  it("weekly costs bar tooltip lists the week's segments and total on hover and focus", async () => {
    const user = userEvent.setup();
    await renderWeekly();
    expect(screen.queryByRole("tooltip")).toBeNull();
    await user.hover(segments("2026-06-10")[0]);
    const tip = screen.getByRole("tooltip");
    expect(tip).toHaveTextContent("Week of 2026-06-10");
    const rows = within(tip).getAllByRole("row").map((r) => r.textContent);
    expect(rows).toEqual(["Planning$30.00", "Implementation$25.00", "Review$20.00", "Self-review$12.00", "Other$5.00", "Total$92.00"]);
    await user.unhover(segments("2026-06-10")[0]);
    expect(screen.queryByRole("tooltip")).toBeNull();
    act(() => segments("2026-06-17")[0].focus());
    expect(screen.getByRole("tooltip")).toHaveTextContent("Week of 2026-06-17 (so far)");
    expect(within(screen.getByRole("tooltip")).getAllByRole("row").at(-1)).toHaveTextContent("Total$12.00");
    act(() => segments("2026-06-17")[0].blur());
    expect(screen.queryByRole("tooltip")).toBeNull();
  });

  it("weekly costs reads not-recorded as Panel does", async () => {
    fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures) =>
      Promise.resolve(view === "weekly-costs" ? envelope("weekly-costs", [], false) : fixtures[view]),
    );
    vi.useFakeTimers({ now: WC_NOW, toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    render(<WeeklyCosts />);
    expect(await screen.findByTestId("not-recorded")).toBeInTheDocument();
    expect(screen.queryAllByTestId("wc-segment")).toHaveLength(0);
  });

  // Serve `rows` / `modelRows` instead of the shared weekly fixtures.
  const serveWeekly = (rows: unknown[], modelRows: unknown[] = WEEKLY_MODEL_ROWS, projects: unknown[] = WEEKLY_PROJECTS) =>
    fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures) =>
      Promise.resolve(
        view === "weekly-costs"
          ? { ...envelope("weekly-costs", rows), projects }
          : view === "weekly-model-effort-costs"
            ? envelope("weekly-model-effort-costs", modelRows)
            : fixtures[view],
      ),
    );

  it("weekly costs Model · effort keeps the six costliest pairs whatever the row order and sums only the rest into Other models", async () => {
    // The cheapest pair comes first in row order; the sixth costliest (opus-5 · medium, $6) must be drawn once.
    serveWeekly(WEEKLY_ROWS, [
      me("2026-06-10", "claude-sonnet-5", "low", 1),
      me("2026-06-10", "claude-opus-5-5", "session", 30),
      me("2026-06-10", "claude-opus-5-5", "high", 20),
      me("2026-06-10", "claude-opus-5-5", "medium", 15),
      me("2026-06-10", "claude-sonnet-5-5", "medium", 12),
      me("2026-06-10", "claude-opus-5-5", "low", 10),
      me("2026-06-10", "claude-opus-5", "medium", 6),
      me("2026-06-10", "claude-sonnet-5", "high", 2),
    ]);
    const user = userEvent.setup();
    await renderWeekly();
    await user.click(chip("Split by", "Model · effort"));
    await screen.findByText("Spend per week by model · effort");
    const names = ["opus-5-5 · session", "opus-5-5 · high", "opus-5-5 · medium", "sonnet-5-5 · medium", "opus-5-5 · low", "opus-5 · medium", "Other models"];
    expect(legend()).toEqual(names);
    const week = segments("2026-06-10");
    expect(week.map((s) => s.getAttribute("data-series"))).toEqual(names);
    expect(week.at(-1)).toHaveAccessibleName("Other models, week of 2026-06-10: $3.00");
    expect(screen.getAllByTestId("wc-total").map((t) => t.textContent)).toEqual(["$96.00"]);
    // Series colours run --series-1.. in split order.
    expect(week.map((s) => s.getAttribute("fill"))).toEqual(names.map((_, i) => `var(--series-${i + 1})`));
    // The Y axis keeps headroom above the tallest week: its top tick is at least 1.08 times that week's total.
    const ticks = [...document.querySelectorAll(".wc-plot .grid text")].map((t) => Number(t.textContent!.replace(/[$,]/g, "")));
    expect(ticks.at(-1)!).toBeGreaterThanOrEqual(96 * 1.08);
  });

  it("weekly costs Biggest mover names nothing when every series cost the same in both compared weeks", async () => {
    const same = { planningUsd: 10, implementationUsd: 40, reviewUsd: 20, otherUsd: 5 };
    serveWeekly([
      { ...noPhase, week: "2026-06-03", overallUsd: 75, ...same },
      { ...noPhase, week: "2026-06-10", overallUsd: 75, ...same },
    ]);
    await renderWeekly();
    const mover = tileOf("Biggest mover");
    expect(mover.getByText("—")).toHaveClass("lb-val");
    expect(mover.queryByText("Planning")).toBeNull();
  });

  it("weekly costs reads a sub-cent float difference as no change", async () => {
    serveWeekly([
      { ...noPhase, week: "2026-06-03", overallUsd: 0.3, planningUsd: 0.3 },
      { ...noPhase, week: "2026-06-10", overallUsd: 0.1 + 0.2, planningUsd: 0.1 + 0.2 },
    ]);
    await renderWeekly();
    const row = within(screen.getByRole("table")).getAllByRole("row")[1];
    expect(within(row).getAllByRole("cell").map((c) => c.textContent)).toEqual(["Planning", "$0.30", "$0.30", "—", "100%"]);
  });

  it("weekly costs tooltip opens leftwards over the last two weeks and rightwards before them", async () => {
    const user = userEvent.setup();
    await renderWeekly();
    const tipLeft = () => parseFloat(screen.getByRole("tooltip").style.left);
    const x = (s: HTMLElement) => parseFloat(s.getAttribute("x")!);
    await user.hover(segments("2026-06-03")[0]);
    expect(tipLeft()).toBeGreaterThan(x(segments("2026-06-03")[0]));
    await user.hover(segments("2026-06-10")[0]);
    expect(tipLeft()).toBeLessThan(x(segments("2026-06-10")[0]));
  });

  it("weekly costs keeps the selected project's chip after a Period change to a window where it had no spend", async () => {
    const user = userEvent.setup();
    await renderWeekly();
    await user.click(chip("Project", "gymie"));
    await waitFor(() => expect(wcCalls("weekly-costs").at(-1)!.project).toBe("gymie-89abcdef"));
    serveWeekly(WEEKLY_ROWS, WEEKLY_MODEL_ROWS, [{ projectKey: "agents-0123abcd", totalCostUsd: 80 }]);
    await user.click(chip("Period", "26w"));
    await waitFor(() => expect(wcCalls("weekly-costs").at(-1)!.from).toEqual(new Date("2025-12-24T00:00:00Z")));
    await waitFor(() => expect(group("Project").getAllByRole("button").map((b) => b.textContent)).toEqual(["All", "agents", "gymie"]));
    expect(pressedIn("Project")).toEqual(["gymie"]);
    expect(wcCalls("weekly-costs").at(-1)!.project).toBe("gymie-89abcdef");
  });
});

describe("trend", () => {
  const trendCalls = () => fetchStatsViewMock.mock.calls.filter((c) => c[0] === "trend").map((c) => c[1] as { from: Date; to: Date; project?: string; model?: string });
  const group = (name: string) => within(screen.getByRole("group", { name }));
  const pressed = (name: string, chip: string | RegExp) => group(name).getByRole("button", { name: chip });
  const tileOf = (label: string) => within([...document.querySelectorAll<HTMLElement>(".tr-tile")].find((t) => t.querySelector(".lb-lbl")?.textContent === label)!);
  const renderTrend = async () => {
    vi.setSystemTime(TREND_NOW);
    render(<Trend />);
    await screen.findByText("Median per change");
  };

  it("trend chips drive the window and the client filters", async () => {
    const user = userEvent.setup();
    await renderTrend();
    expect(trendCalls()).toHaveLength(1);
    const first = trendCalls()[0];
    expect(first.from).toEqual(new Date("2026-05-27T00:00:00Z"));
    expect(first.to).toEqual(TREND_NOW);
    expect(first.project).toBeUndefined();
    expect(first.model).toBeUndefined();
    expect(pressed("Period", "12w")).toHaveAttribute("aria-pressed", "true");
    expect(pressed("Command", /^\/flow7$/)).toHaveAttribute("aria-pressed", "true");
    expect(pressed("Command", /^\/flow-fast8$/)).toHaveAttribute("aria-pressed", "false");
    for (const label of ["Period start", "Model", "Project filter"]) expect(screen.queryByLabelText(label)).toBeNull();
    expect(screen.queryByRole("button", { name: "Clear filters" })).toBeNull();

    await user.click(pressed("Period", "8w"));
    await waitFor(() => expect(trendCalls()).toHaveLength(2));
    expect(trendCalls()[1].from).toEqual(new Date("2026-06-24T00:00:00Z"));
    await screen.findByText("Median per change");

    await user.click(pressed("Project", /^gymie/));
    expect(trendCalls()).toHaveLength(2);
    expect(tileOf("Changes finished").getByText("2")).toBeInTheDocument();
    expect(pressed("Project", /^gymie/)).toHaveAttribute("aria-pressed", "true");

    await user.click(pressed("Command", /^\/flow-fast/));
    expect(pressed("Project", "All")).toHaveAttribute("aria-pressed", "true");
    expect(tileOf("Changes finished").getByText("8")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Clear filters" }));
    await waitFor(() => expect(trendCalls()).toHaveLength(3));
    await screen.findByText("Median per change");
    expect(pressed("Period", "12w")).toHaveAttribute("aria-pressed", "true");
    expect(pressed("Command", /^\/flow7$/)).toHaveAttribute("aria-pressed", "true");
    expect(pressed("Project", "All")).toHaveAttribute("aria-pressed", "true");
    expect(screen.queryByRole("button", { name: "Clear filters" })).toBeNull();
  });

  it("trend tiles compare the window's halves and count flow changes", async () => {
    const user = userEvent.setup();
    await renderTrend();
    // Last six weeks' /flow costs [8, 2, 3, 40, 4] against the first six's [10, 12].
    const med = tileOf("Median per change");
    med.getByText("$4.00");
    const delta = med.getByText("−$7.00 (−64%)");
    expect(delta).toHaveClass("down");
    expect(med.getByText(/vs the 6 weeks before/)).toHaveTextContent("last 6 weeks · −$7.00 (−64%) vs the 6 weeks before");
    tileOf("Changes finished").getByText("7");
    tileOf("Changes finished").getByText("$79.00 spent on them");
    // Every agents change of every command: five /flow, four /flow-fast.
    tileOf("Flow changes landed").getByText("9");
    await user.click(pressed("Project", /^gymie/));
    tileOf("Flow changes landed").getByText("9");
    await user.click(pressed("Command", /^\/flow-fast/));
    tileOf("Flow changes landed").getByText("9");
  });

  it("trend chart draws median and mean with gaps", async () => {
    await renderTrend();
    const median = document.querySelector(".tr-chart path.median")!.getAttribute("d")!;
    const meanD = document.querySelector(".tr-chart path.mean")!.getAttribute("d")!;
    // Jul 1 → Jul 8, then a move-to past the empty Jul 15 to Jul 22 → Jul 29.
    expect(median.match(/M/g)).toHaveLength(2);
    expect(median.match(/[ML]/g)).toHaveLength(4);
    expect(document.querySelectorAll(".tr-chart circle.pt")).toHaveLength(4);
    expect(meanD).not.toEqual(median);
    expect(document.querySelectorAll(".tr-chart rect.fc")).toHaveLength(12);
  });

  it("trend tooltip says 1 change, not 1 changes", async () => {
    const user = userEvent.setup();
    await renderTrend();
    await user.hover(screen.getByTestId("trend-week-2026-07-29"));
    const tip = document.querySelector<HTMLElement>(".tr-tip")!;
    expect(tip).toHaveTextContent("1 change · median $4.00 · mean $4.00");
    expect(tip).not.toHaveTextContent("1 changes");
  });

  it("trend chart is drawn at the prototype's fixed 1029 by 300", async () => {
    await renderTrend();
    const svg = screen.getByRole("img", { name: "Cost per finished change by week" });
    expect(svg).toHaveAttribute("viewBox", "0 0 960 280");
    expect(svg).toHaveAttribute("width", "1029");
    expect(svg).toHaveAttribute("height", "300");
  });

  it("trend chart hover shows the week's tip", async () => {
    const user = userEvent.setup();
    await renderTrend();
    await user.hover(screen.getByTestId("trend-week-2026-07-22"));
    const tip = document.querySelector<HTMLElement>(".tr-tip")!;
    expect(tip).toHaveTextContent("Week of Jul 22");
    expect(tip).toHaveTextContent("3 changes · median $3.00 · mean $15.00");
    expect(tip).toHaveTextContent("3 flow changes landed");
    await user.hover(screen.getByTestId("trend-week-2026-07-15"));
    expect(tip).toHaveTextContent("Week of Jul 15");
    expect(tip).toHaveTextContent("no finished changes");
    expect(tip).toHaveTextContent("0 flow changes landed");
  });

  it("trend table lists weeks newest first with the empty state", async () => {
    await renderTrend();
    const table = screen.getByRole("table");
    expect(within(table).getAllByRole("columnheader").map((h) => h.textContent)).toEqual([
      "Week",
      "Changes",
      "Median",
      "vs previous week",
      "Mean",
      "Spend",
      "Flow changes",
    ]);
    const rows = within(table).getAllByRole("row").slice(1);
    expect(rows).toHaveLength(12);
    expect(rows[0]).toHaveTextContent(/^Aug 12 so far/);
    const jul15 = within(rows.find((r) => r.textContent?.startsWith("Jul 15"))!).getAllByRole("cell");
    expect([jul15[1], jul15[2], jul15[4], jul15[5]].map((c) => c.textContent)).toEqual(["0", "—", "—", "—"]);
    expect(rows[rows.length - 1]).toHaveTextContent(/^May 27/);
  });

  it("trend reads no rows for the selection as its empty state", async () => {
    fetchStatsViewMock.mockImplementation((view: StatsViewSlug) => Promise.resolve(envelope(view, [])));
    await renderTrend();
    screen.getByText("No finished changes in this period.");
    expect(screen.queryByRole("table")).toBeNull();
  });
});

describe("absence is rendered distinctly from a recorded zero", () => {
  it("a period before the store held anything is stated, not rendered as an empty measured table", async () => {
    fetchStatsViewMock.mockImplementation((view: StatsViewSlug) =>
      Promise.resolve(envelope(view, [], false)),
    );
    render(<Trend />);
    // Trend's body reads its one request's absence through Panel, the sole
    // implementation of that branch (Panel.tsx's header comment).
    const banners = await screen.findAllByTestId("not-recorded");
    expect(banners.length).toBeGreaterThan(0);
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.queryByText("Median per change")).not.toBeInTheDocument();
  });

  // Task 5's third arm, exercised end to end through a real view rather
  // than Panel.test.tsx's isolated fixture: a period whose runs were
  // recorded but never attributed must read as its own state, distinct
  // from both "no data was recorded" (the case immediately above) and
  // from an ordinary measured table -- never collapsing into either.
  it("a period whose runs were recorded but none attributed reports that plainly, not as 'no data was recorded'", async () => {
    fetchStatsViewMock.mockImplementation((view: StatsViewSlug) =>
      Promise.resolve(envelope(view, [], true, true)),
    );
    render(<Trend />);
    const banners = await screen.findAllByTestId("unmeasured");
    expect(banners.length).toBeGreaterThan(0);
    for (const banner of banners) {
      expect(banner).toHaveTextContent("Runs were recorded for this period, but none carried measurements.");
    }
    expect(screen.queryByTestId("not-recorded")).not.toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.queryByText("Median per change")).not.toBeInTheDocument();
  });
});

describe("the state board is the only navigation path into a change's own dashboard", () => {
  it("links a row's change name to its run-detail route, round-tripping a URL-significant character in both segments", async () => {
    fetchStatsViewMock.mockImplementation((view: StatsViewSlug) =>
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
    fetchStatsViewMock.mockImplementation((view: StatsViewSlug) =>
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
    fetchStatsViewMock.mockImplementation((view: StatsViewSlug) =>
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
    expect(await screen.findByRole("cell", { name: "Integrating" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Live state board" })).toBeInTheDocument();
  });

  it("an unrecognised route also falls back to the state board", async () => {
    window.location.hash = "#/not-a-real-view";
    render(<App />);
    expect(await screen.findByRole("cell", { name: "Integrating" })).toBeInTheDocument();
    window.location.hash = "";
  });
});

describe("flow health", () => {
  const answerGuards = (rows: (ReturnType<typeof gh> & { helper?: boolean })[], projects: { projectKey: string; runs: number }[] = []) =>
    fetchStatsViewMock.mockResolvedValueOnce({ ...envelope("guard-health", rows), projects });
  const names = () => [...document.querySelectorAll("tbody tr")].map((tr) => tr.querySelector(".rv-slot")!.textContent);
  const rankLine = () => document.querySelectorAll("p.lb-sub")[1].textContent;
  const foot = () => [...document.querySelectorAll("tfoot td")].map((td) => td.textContent);
  const cells = (guard: string) => {
    const tr = [...document.querySelectorAll("tbody tr")].find((r) => r.querySelector(".rv-slot")!.textContent === guard)!;
    return [...tr.querySelectorAll("td")];
  };
  const group = (name: string) => within(screen.getByRole("group", { name }));

  it("flow guards: a known guard shows its what-it-does hint, and approved ones their badge", async () => {
    answerGuards([
      gh("check-guard-calls", 10, 1, 0, 100, 1000, "2026-01-10T00:00:00Z"),
      gh("sweep-dispatch-rows", 5, 0, 0, 10, 50, null),
      gh("some-other-guard", 5, 0, 0, 10, 50, null),
    ]);
    render(<FlowHealth />);
    // One panel per group: Checks, Dispatch bookkeeping and Other each hold one of the three.
    expect(await screen.findAllByRole("table")).toHaveLength(3);
    const row = (name: string) => screen.getByText(name, { selector: "td *" }).closest("tr") as HTMLElement;
    expect(within(row("check-guard-calls")).getByText("approved")).toBeInTheDocument();
    expect(within(row("check-guard-calls")).getByLabelText(/names a real guard/)).toHaveTextContent("i");
    expect(within(row("sweep-dispatch-rows")).getByText("approved")).toBeInTheDocument();
    expect(within(row("sweep-dispatch-rows")).getByLabelText(/stop was missed/)).toHaveTextContent("i");
    expect(within(row("some-other-guard")).queryByText("approved")).not.toBeInTheDocument();
  });

  it("flow guards: one panel per group, no tiles, and no fire columns on a panel of helpers", async () => {
    answerGuards([
      gh("check-guard-calls", 10, 1, 0, 100, 1000, "2026-01-10T00:00:00Z"),
      { ...gh("project-get", 50, 0, 0, 5, 250, null), helper: true },
      { ...gh("kickoff-worktree", 2, 0, 0, 900, 1800, null), helper: true },
    ]);
    render(<FlowHealth />);
    await screen.findAllByRole("table");
    const panel = (name: string) => screen.getByLabelText(name, { selector: ".lb-panel" });
    expect(document.querySelectorAll(".lb-panel")).toHaveLength(2);
    // Checks holds a real guard: fire columns.
    expect(within(panel("Checks")).getByRole("columnheader", { name: "Fired" })).toBeInTheDocument();
    expect(document.querySelectorAll(".rv-tiles")).toHaveLength(0);
    // Workspace and git holds helpers only: no fire columns.
    const ws = panel("Workspace and git");
    expect(within(ws).queryByRole("columnheader", { name: "Fired" })).not.toBeInTheDocument();
    expect(within(ws).queryByRole("columnheader", { name: "Last fired" })).not.toBeInTheDocument();
  });

  it("flow health ranks by fires, fire rate, never fired and run time", async () => {
    const user = userEvent.setup();
    answerGuards([
      // alpha and bravo tie on fires (bravo more runs) and on run time (name order alpha first);
      // alpha and charlie tie on fire rate (charlie more fires); delta and echo never fired (echo more runs).
      gh("alpha", 100, 10, 0, 500, 50000, "2026-01-10T00:00:00Z"),
      gh("bravo", 200, 10, 0, 250, 50000, "2026-01-11T00:00:00Z"),
      gh("charlie", 400, 40, 0, 2, 1000, "2026-01-12T00:00:00Z"),
      gh("delta", 50, 0, 0, 40, 2000, null),
      gh("echo", 300, 0, 0, 1666, 500000, null),
    ]);
    render(<FlowHealth />);
    await screen.findByRole("table");
    const sorted = () => [...document.querySelectorAll("th[aria-sort]")].map((th) => [th.textContent, th.getAttribute("aria-sort")]);
    const expectRank = (order: string[], heading: string, header: string) => {
      expect(names()).toEqual(order);
      expect([...document.querySelectorAll("tbody tr")].map((tr) => tr.querySelector("td")!.textContent)).toEqual(["1", "2", "3", "4", "5"]);
      expect(rankLine()).toContain(heading);
      expect(sorted()).toEqual([[header, "descending"]]);
      const muted = [...document.querySelectorAll("tbody tr.lb-muted")].map((tr) => tr.querySelector(".rv-slot")!.textContent).sort();
      expect(muted).toEqual(["delta", "echo"]);
    };
    expectRank(["charlie", "bravo", "alpha", "echo", "delta"], "Every guard, most fires first", "Fired");
    await user.click(group("Rank by").getByRole("button", { name: "Fire rate" }));
    expectRank(["charlie", "alpha", "bravo", "delta", "echo"], "Every guard, highest fire rate first", "Fire rate");
    await user.click(group("Rank by").getByRole("button", { name: "Never fired" }));
    expectRank(["echo", "delta", "charlie", "bravo", "alpha"], "Every guard, never-fired first, most runs first", "Last fired");
    await user.click(group("Rank by").getByRole("button", { name: "Run time" }));
    expectRank(["echo", "alpha", "bravo", "delta", "charlie"], "Every guard, most run time first", "Run time");
  });

  it("flow health chips request the window and project, and Clear keeps the rank", async () => {
    const user = userEvent.setup();
    const rows = fixtures["guard-health"].rows;
    // Response order, not alphabetical: zeta before alpha.
    const projects = [
      { projectKey: "zeta-1234abcd", runs: 9000 },
      { projectKey: "alpha-0000aaaa", runs: 12 },
    ];
    fetchStatsViewMock.mockImplementation(() => Promise.resolve({ ...envelope("guard-health", rows), projects }));
    const reqs = () => fetchStatsViewMock.mock.calls.filter((c) => c[0] === "guard-health").map((c) => c[1] as { from: Date; to: Date; project?: string; projects?: boolean });
    const day = 24 * 60 * 60 * 1000;
    const now = BOARD_NOW.getTime();
    const pressed = (g: string, name: string | RegExp) => expect(group(g).getByRole("button", { name })).toHaveAttribute("aria-pressed", "true");
    render(<FlowHealth />);
    await screen.findByRole("table");
    expect(reqs()).toHaveLength(1);
    expect(reqs()[0]).toMatchObject({ projects: true, project: undefined });
    expect([reqs()[0].from.getTime(), reqs()[0].to.getTime()]).toEqual([now - 30 * day, now]);
    expect(group("Project").getAllByRole("button").map((b) => b.textContent)).toEqual(["All", "zeta9,000", "alpha12"]);
    expect(group("Project").getByRole("button", { name: /^zeta/ }).querySelector(".lb-c")).toHaveTextContent("9,000");
    pressed("Rank by", "Fires");

    await user.click(group("Rank by").getByRole("button", { name: "Run time" }));
    await user.click(group("Project").getByRole("button", { name: /^alpha/ }));
    await waitFor(() => expect(reqs().at(-1)).toMatchObject({ project: "alpha-0000aaaa", projects: true }));
    await screen.findByRole("table");
    pressed("Rank by", "Run time");
    expect(group("Project").getByRole("button", { name: /^alpha/ })).toHaveAttribute("title", "alpha-0000aaaa");

    await user.click(group("Period").getByRole("button", { name: "7d" }));
    await waitFor(() => expect([reqs().at(-1)!.from.getTime(), reqs().at(-1)!.to.getTime()]).toEqual([now - 7 * day, now]));
    await screen.findByRole("table");
    expect(reqs().at(-1)).toMatchObject({ project: "alpha-0000aaaa" });
    pressed("Project", /^alpha/);

    await user.click(screen.getByRole("button", { name: "Clear filters" }));
    pressed("Period", "30d");
    pressed("Project", "All");
    pressed("Rank by", "Run time");
    await waitFor(() => expect(reqs().at(-1)).toMatchObject({ project: undefined }));
    expect([reqs().at(-1)!.from.getTime(), reqs().at(-1)!.to.getTime()]).toEqual([now - 30 * day, now]);
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();
  });

  it("flow health formats counts, rates and run times", async () => {
    answerGuards([
      gh("f-zero", 12345, 0, 0, 450, 1500, null),
      gh("f-tiny", 2000, 1, 0, 1500, 600000, "2026-08-14T23:30:00Z"),
      gh("f-five", 100, 5, 3, 1500, 4200000, "2026-08-15T00:00:00Z"),
      gh("f-quarter", 100, 25, 0, 1500, 1000, "2026-08-16T00:00:00Z"),
    ]);
    const { unmount } = render(<FlowHealth />);
    await screen.findByRole("table");
    const text = (guard: string) => cells(guard).map((td) => td.textContent);
    // Columns: #, Guard, Runs, Fired, Fire rate, Cannot answer, Median run, Run time, Share, Last fired.
    expect(text("f-zero").slice(2)).toEqual(["12,345", "—", "0%", "—", "450 ms", "1.5 s", "<0.1%", "never"]);
    expect(cells("f-zero")[9].querySelector(".lb-muted")).toHaveTextContent("never");
    expect(text("f-tiny").slice(2)).toEqual(["2,000", "1", "<0.1%", "—", "1.5 s", "10 min", "12%", "2026-08-14"]);
    expect(text("f-five").slice(2)).toEqual(["100", "5", "5.0%", "3.0%", "1.5 s", "1 h 10 min", "87%", "2026-08-15"]);
    expect(text("f-quarter").slice(2, 5)).toEqual(["100", "25", "25%"]);
    unmount();

    // Rounding carries into the next unit rather than showing 60 of the smaller one.
    answerGuards([gh("c-s", 1, 0, 0, 999.6, 59960, null), gh("c-m", 1, 0, 0, 1, 3599000, null), gh("c-h", 1, 0, 0, 1, 7170000, null)]);
    const second = render(<FlowHealth />);
    await screen.findByRole("table");
    expect(text("c-s").slice(6, 8)).toEqual(["1.0 s", "1 min"]);
    expect(text("c-m")[7]).toBe("1 h 0 min");
    expect(text("c-h")[7]).toBe("2 h 0 min");
    second.unmount();

    answerGuards([gh("solo", 10, 1, 0, 450, 1500, "2026-08-14T23:30:00Z")]);
    render(<FlowHealth />);
    await screen.findByRole("table");
    expect(foot()[1]).toBe("1 script");
    expect(foot()[8]).toBe("100%");
  });

  it("flow health shows no tiles and the empty message with no guard runs", async () => {
    answerGuards([]);
    const { unmount } = render(<FlowHealth />);
    expect(await screen.findByText("No guard runs in this period.")).toBeInTheDocument();
    expect(document.querySelector(".lb-tile")).toBeNull();
    unmount();

    fetchStatsViewMock.mockResolvedValueOnce({ ...envelope("guard-health", fixtures["guard-health"].rows, false), projects: [] });
    render(<FlowHealth />);
    expect(await screen.findByTestId("not-recorded")).toBeInTheDocument();
    expect(document.querySelector(".lb-tile")).toBeNull();
  });
});

describe("review dynamics", () => {
  const lastParams = () => fetchStatsViewMock.mock.calls.at(-1)![1] as { from: Date; to: Date };
  const days = (p: { from: Date; to: Date }) => Math.round((p.to.getTime() - p.from.getTime()) / 86400000);

  it("review dynamics shows only the findings heatmap, busiest slot first and effort low to high", async () => {
    const rows = reviewDynamicsRows();
    const cell = (slot: string, model: string, effort: string, dispatches: number) => ({ slot, model, effort, dispatches, critical: 0, important: 1, minor: 1 });
    fetchStatsViewMock.mockResolvedValue(
      envelope("review-dynamics", {
        ...rows,
        // test-sanity is busiest and sorts last by name; sonnet's low sorts before opus's high by effort alone.
        heat: [cell("mutation", "sonnet", "low", 2), cell("test-sanity", "opus", "high", 3), cell("test-sanity", "opus", "low", 3), cell("test-sanity", "opus", "medium", 1)],
      }),
    );
    render(<ReviewDynamics />);
    const heat = (await screen.findByText(/Findings per dispatch: slot/)).closest(".dynamics-panel") as HTMLElement;
    expect(within(heat).getAllByRole("rowheader").map((h) => h.textContent)).toEqual(["test-sanity", "mutation"]);
    expect(within(heat).getAllByRole("columnheader").map((h) => h.textContent)).toEqual(["", "opuslow", "opusmedium", "opushigh", "sonnetlow"]);
    for (const gone of ["Review panel", "Implementer", "Visual verify"]) expect(screen.queryByText(gone)).toBeNull();
    expect(screen.queryByRole("group", { name: "Projects" })).toBeNull();
  });

  it("review dynamics opens on the last 30 days and picks 7, 14, 30 or 60", async () => {
    fetchStatsViewMock.mockResolvedValue(envelope("review-dynamics", reviewDynamicsRows()));
    render(<ReviewDynamics />);
    await screen.findByText(/Findings per dispatch: slot/);
    expect(days(lastParams())).toBe(30);
    expect(Math.abs(lastParams().to.getTime() - Date.now())).toBeLessThan(60_000);
    const range = screen.getByRole("group", { name: "Range" });
    expect(within(range).getAllByRole("button", { pressed: false }).map((b) => b.textContent)).toEqual(["7 days", "14 days", "60 days"]);
    await userEvent.click(within(range).getByRole("button", { name: "7 days" }));
    await waitFor(() => expect(days(lastParams())).toBe(7));
  });

  it("review dynamics steps the range back and forth by its own length, never past now", async () => {
    fetchStatsViewMock.mockResolvedValue(envelope("review-dynamics", reviewDynamicsRows()));
    render(<ReviewDynamics />);
    await screen.findByText(/Findings per dispatch: slot/);
    const range = screen.getByRole("group", { name: "Range" });
    await userEvent.click(within(range).getByRole("button", { name: "7 days" }));
    const later = within(range).getByRole("button", { name: "Later" });
    expect(later).toBeDisabled();
    await waitFor(() => expect(days(lastParams())).toBe(7));
    const now = lastParams();
    await userEvent.click(within(range).getByRole("button", { name: "Earlier" }));
    await waitFor(() => expect(lastParams().to.getTime()).toBe(now.from.getTime()));
    expect(days(lastParams())).toBe(7);
    expect(later).toBeEnabled();
    expect(range).toHaveTextContent(`${lastParams().from.toISOString().slice(5, 10)} – ${lastParams().to.toISOString().slice(5, 10)}`);
    await userEvent.click(later);
    await waitFor(() => expect(lastParams().to.getTime()).toBe(now.to.getTime()));
  });
});

describe("daily costs", () => {
  // Friday 2026-06-19, local noon: the default 7 days run Sat 06-13 .. Fri 06-19.
  const DC_NOW = new Date(2026, 5, 19, 12);
  // The window fetches span at most 14 days; the all-time fetch spans months.
  const windowCalls = () =>
    fetchStatsViewMock.mock.calls.map((c) => c[1] as { from: Date; to: Date }).filter((p) => p.to.getTime() - p.from.getTime() <= 15 * 86400000);
  const lastParams = () => windowCalls().at(-1)!;
  // One row per 24h from `from` to `to`, a day costing its day-of-month in dollars (planning $1 of it).
  // Nothing is recorded before 06-01, and 06-10 is an empty day inside the recorded span (both cost 0).
  const dailyRows = (p: { from: Date; to: Date }) => {
    const rows = [];
    for (let t = p.from.getTime(); t < p.to.getTime(); t += 86400000) {
      const d = new Date(t);
      const cost = d < new Date(2026, 5, 1) || (d.getMonth() === 5 && d.getDate() === 10) ? 0 : d.getDate();
      rows.push({ week: d.toISOString(), overallUsd: cost, planningUsd: cost ? 1 : 0, implementationUsd: cost ? cost - 1 : 0, reviewUsd: null, selfReviewUsd: null, visualVerifyUsd: null, otherUsd: null });
    }
    return rows;
  };
  const renderDaily = async () => {
    vi.useFakeTimers({ now: DC_NOW, toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    fetchStatsViewMock.mockImplementation((_v: string, p: { from: Date; to: Date }) => Promise.resolve(envelope("daily-costs", dailyRows(p))));
    render(<DailyCosts />);
    await screen.findAllByTestId("dc-bar");
    await screen.findAllByTestId("dc-wbar");
  };
  const range = () => screen.getByRole("group", { name: "Range" });
  const press = (name: string) => userEvent.click(within(range()).getByRole("button", { name }));
  const avg = (weekday: string) => screen.getAllByTestId("dc-wbar").find((e) => e.getAttribute("data-key") === weekday)!.querySelector("[data-testid=wc-total]")!.textContent;

  it("daily costs opens on 7 days with a column per day and offers 7 and 14 days only", async () => {
    await renderDaily();
    expect(screen.getAllByTestId("dc-bar")).toHaveLength(7);
    expect(lastParams().from).toEqual(new Date(2026, 5, 13));
    expect(within(range()).getAllByRole("button", { pressed: true }).map((b) => b.textContent)).toEqual(["7 days"]);
    expect(within(range()).getAllByRole("button", { pressed: false }).map((b) => b.textContent)).toEqual(["14 days"]);
    await press("14 days");
    await waitFor(() => expect(screen.getAllByTestId("dc-bar")).toHaveLength(14));
  });

  it("daily costs steps a window earlier and later, never past now", async () => {
    await renderDaily();
    const later = within(range()).getByRole("button", { name: "Later" });
    expect(later).toBeDisabled();
    const now = lastParams();
    await press("Earlier");
    await waitFor(() => expect(lastParams().to.getTime()).toBe(now.from.getTime()));
    expect(screen.getAllByTestId("dc-bar")).toHaveLength(7);
    expect(later).toBeEnabled();
    await userEvent.click(later);
    await waitFor(() => expect(lastParams().to.getTime()).toBe(now.to.getTime()));
  });

  it("daily costs has no total block and names its average block plainly", async () => {
    await renderDaily();
    expect(screen.queryByTestId("dc-total")).toBeNull();
    expect(screen.queryByText("All time")).toBeNull();
    expect(screen.getByText("Average per weekday")).toBeInTheDocument();
    expect(document.body.textContent).not.toMatch(/all[- ]time/i);
    expect(screen.getAllByTestId("dc-bar")[6]).toHaveTextContent("$19.00");
  });

  it("daily costs labels each day with its date and its weekday, for 7 and 14 days", async () => {
    await renderDaily();
    expect(screen.getAllByTestId("wc-week").slice(0, 7).map((e) => e.textContent)).toEqual(["06-13", "06-14", "06-15", "06-16", "06-17", "06-18", "06-19"]);
    expect(screen.getAllByTestId("wc-sub").slice(0, 7).map((e) => e.textContent)).toEqual(["Sat", "Sun", "Mon", "Tue", "Wed", "Thu", "Fri"]);
    await press("14 days");
    await waitFor(() => expect(screen.getAllByTestId("dc-bar")).toHaveLength(14));
    expect(screen.getAllByTestId("wc-sub").slice(0, 14).map((e) => e.textContent).slice(0, 2)).toEqual(["Sat", "Sun"]);
  });

  it("daily costs has no numeric day table", async () => {
    await renderDaily();
    expect(screen.queryByTestId("dc-day")).toBeNull();
    expect(screen.queryByTestId("dc-day-label")).toBeNull();
    expect(screen.queryByRole("table")).toBeNull();
  });

  it("daily costs averages each weekday over the recorded span, an empty day counting as 0", async () => {
    await renderDaily();
    expect(screen.getAllByTestId("dc-wbar").map((e) => e.getAttribute("data-key"))).toEqual(["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]);
    expect(avg("Mon")).toBe("$8.00"); // 06-01, 08, 15
    expect(avg("Wed")).toBe("$6.67"); // 06-03, 10 (empty), 17
    expect(avg("Fri")).toBe("$12.00"); // 06-05, 12, 19
    expect(avg("Sat")).toBe("$9.50"); // 06-06, 13
    await press("14 days");
    await waitFor(() => expect(screen.getAllByTestId("dc-bar")).toHaveLength(14));
    expect(avg("Wed")).toBe("$6.67");
  });

  const heights = (bar: HTMLElement) => within(bar).getAllByTestId("wc-segment").map((e) => Number(e.getAttribute("height")));
  const sum = (xs: number[]) => xs.reduce((a, b) => a + b, 0);

  it("daily costs charts a stacked bar per day of the window, in proportion to its cost", async () => {
    await renderDaily();
    expect(screen.getByRole("img", { name: /^Daily costs, / })).toBeInTheDocument();
    const bars = screen.getAllByTestId("dc-bar");
    expect(bars).toHaveLength(7);
    // 06-13 costs $13 (planning 1 + implementation 12), 06-19 $19.
    expect(bars[0]).toHaveTextContent("$13.00");
    expect(heights(bars[0])).toHaveLength(2);
    expect(sum(heights(bars[6])) / sum(heights(bars[0]))).toBeCloseTo(19 / 13, 5);
    const [plan, impl] = heights(bars[6]);
    expect(plan / (plan + impl)).toBeCloseTo(1 / 19, 5);
    expect(within(bars[0].parentElement!).getAllByTestId("wc-week")).toHaveLength(7);
    await press("14 days");
    await waitFor(() => expect(screen.getAllByTestId("dc-bar")).toHaveLength(14));
  });

  it("daily costs charts each weekday's mean over the recorded span, a dash for one without data", async () => {
    await renderDaily();
    const wd = screen.getAllByTestId("dc-wbar");
    expect(wd.map((e) => e.getAttribute("data-key"))).toEqual(["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]);
    expect(wd[0]).toHaveTextContent("$8.00");
    expect(screen.getByRole("img", { name: /^Average cost per weekday/ })).toBeInTheDocument();
  });

  it("daily costs leaves a weekday without data an empty slot with a dash", async () => {
    vi.useFakeTimers({ now: DC_NOW, toFake: ["Date"] });
    onTestFinished(() => {
      vi.useRealTimers();
    });
    // The all-time span holds a Monday and a Tuesday only.
    fetchStatsViewMock.mockImplementation((_v: string, p: { from: Date; to: Date }) =>
      Promise.resolve(envelope("daily-costs", p.to.getTime() - p.from.getTime() > 15 * 86400000 ? dailyRows(p).filter((r) => [15, 16].includes(new Date(r.week).getDate())) : dailyRows(p))),
    );
    render(<DailyCosts />);
    await screen.findAllByTestId("dc-wbar");
    const wd = screen.getAllByTestId("dc-wbar");
    expect(wd).toHaveLength(7);
    expect(wd[2]).toHaveTextContent("no data");
    expect(within(wd[2]).queryAllByTestId("wc-segment")).toHaveLength(0);
    expect(wd[0]).toHaveTextContent("$15.00");
  });
});

describe("AgentRoles", () => {
  const roles = [
    {
      role: "main", base: [], optin: [], group: "Session", does: "The session you talk to: runs the pipeline and sends out every other role.",
    },
    { role: "explorer", base: ["opus/Explore@low", "opus/Explore@medium"], optin: [], group: "Planning", does: "Reads the codebase to answer a planner's question." },
    {
      role: "plan-designer", base: ["opus/Plan@low", "opus/Plan@medium"], optin: ["fable/Plan@xhigh"], group: "Planning",
      does: "Proposes a design or plan for the planner to choose from.",
    },
    {
      role: "implementer", base: ["opus/flow-low"], optin: [],
      group: "Implementation", does: "Writes the code for a group of tasks, tests first.",
    },
  ];
  beforeEach(() => {
    postRolesEditMock.mockReset().mockResolvedValue(undefined);
    fetchRolesMock.mockReset().mockResolvedValue(roles);
  });

  it("Agent roles groups the roles, one row per role, on Main and Explicit allow columns", async () => {
    render(<AgentRoles />);
    const table = await screen.findByRole("region", { name: "Current roles" });
    expect(within(table).getAllByRole("columnheader").slice(0, 3).map((h) => h.textContent)).toEqual(["Role", "Main", "Explicit allow"]);
    const rows = within(table).getAllByRole("row");
    expect(rows.map((r) => r.querySelector("th, td")?.textContent)).toEqual([
      "Role", "Session", "mainRuns the pipeline", "Planning", "explorerReads the codebase", "plan-designerProposes a plan", "Implementation", "implementerWrites the code",
    ]);
    for (const i of [1, 3, 6]) expect(rows[i].querySelector("th")).toHaveAttribute("colspan", "3");
    const designer = rows[5];
    expect(designer.querySelector("td.role")).toHaveAttribute("title", "Proposes a design or plan for the planner to choose from.");
    expect(designer.querySelector("td.agent")).toBeNull();
    expect([...designer.querySelectorAll("td.pairs")[0].querySelectorAll(".chip .pair")].map((p) => p.textContent)).toEqual(["opus/Plan@low", "opus/Plan@medium"]);
    expect([...designer.querySelectorAll("td.pairs")[1].querySelectorAll(".chip .pair")].map((p) => p.textContent)).toEqual(["fable/Plan@xhigh"]);
    expect(within(designer).queryByRole("link")).toBeNull();
    expect(rows[2].querySelector("td.pairs")).toHaveTextContent("session's own pair");
  });

  it("Agent roles tags the headless roles", async () => {
    fetchRolesMock.mockResolvedValue([
      { role: "pair-judge", base: ["sonnet/flow-low"], optin: [], group: "Pair choice", does: "Picks each dispatch's model and effort." },
      { role: "implementer", base: ["opus/flow-low"], optin: [], group: "Implementation", does: "Writes the code." },
    ]);
    render(<AgentRoles />);
    const table = await screen.findByRole("region", { name: "Current roles" });
    const tagged = within(table).getAllByText("headless");
    expect(tagged).toHaveLength(1);
    expect(tagged[0].closest("td")).toHaveTextContent("pair-judge");
  });

  it("Agent roles edits a role's base pairs", async () => {
    render(<AgentRoles />);
    const table = await screen.findByRole("region", { name: "Current roles" });
    expect(within(table).queryByRole("button", { name: "Save" })).toBeNull();
    await userEvent.click(within(table).getByRole("button", { name: "Choose main pairs for explorer" }));
    const menu = table.querySelector(".menu") as HTMLElement;
    await userEvent.click(within(menu).getByRole("checkbox", { name: "opus/Explore@low" }));
    const add = within(menu).getByRole("searchbox", { name: "Add main pair to explorer" });
    // A malformed pair, or one already listed, is never offered to add.
    await userEvent.type(add, "opus Explore");
    expect(within(menu).queryByText(/^Add “/)).toBeNull();
    await userEvent.clear(add);
    await userEvent.type(add, "opus/Explore@medium");
    expect(within(menu).queryByText(/^Add “/)).toBeNull();
    await userEvent.clear(add);
    await userEvent.type(add, "sonnet/Explore@max{Enter}");
    expect(add).toHaveValue("");
    await userEvent.click(within(menu).getByRole("button", { name: "Done" }));
    expect(table.querySelector(".menu")).toBeNull();
    await userEvent.click(within(table).getByRole("button", { name: "Save" }));
    expect(postRolesEditMock).toHaveBeenCalledWith({ explorer: { base: ["opus/Explore@medium", "sonnet/Explore@max"], optin: [] } });
    await waitFor(() => expect(fetchRolesMock).toHaveBeenCalledTimes(2));
  });

  it("Agent roles closes a pairs field on Escape and on an outside click", async () => {
    render(<AgentRoles />);
    const table = await screen.findByRole("region", { name: "Current roles" });
    const trig = within(table).getByRole("button", { name: "Choose main pairs for explorer" });
    await userEvent.click(trig);
    expect(table.querySelector(".menu")).not.toBeNull();
    await userEvent.keyboard("{Escape}");
    expect(table.querySelector(".menu")).toBeNull();
    await userEvent.click(trig);
    await userEvent.click(document.body);
    expect(table.querySelector(".menu")).toBeNull();
  });

  it("Agent roles adds an opt-in pair", async () => {
    render(<AgentRoles />);
    const table = await screen.findByRole("region", { name: "Current roles" });
    await userEvent.click(within(table).getByRole("button", { name: "Choose explicit allow pairs for plan-designer" }));
    const menu = table.querySelector(".menu") as HTMLElement;
    await userEvent.type(within(menu).getByRole("searchbox", { name: "Add explicit allow pair to plan-designer" }), "fable/Plan@max{Enter}");
    await userEvent.click(within(menu).getByRole("checkbox", { name: "fable/Plan@xhigh" }));
    await userEvent.click(within(table).getByRole("button", { name: "Save" }));
    expect(postRolesEditMock).toHaveBeenCalledWith({ "plan-designer": { base: ["opus/Plan@low", "opus/Plan@medium"], optin: ["fable/Plan@max"] } });
  });

  it("Agent roles saves edits to several roles in one call", async () => {
    render(<AgentRoles />);
    const table = await screen.findByRole("region", { name: "Current roles" });
    await userEvent.click(within(table).getByRole("button", { name: "Choose main pairs for explorer" }));
    await userEvent.click(within(table).getByRole("checkbox", { name: "opus/Explore@low" }));
    await userEvent.click(within(table).getByRole("button", { name: "Done" }));
    await userEvent.click(within(table).getByRole("button", { name: "Choose main pairs for implementer" }));
    await userEvent.click(within(table).getByRole("checkbox", { name: "opus/flow-low" }));
    await userEvent.click(within(table).getByRole("button", { name: "Done" }));
    expect(table.querySelector(".savebar")).toHaveTextContent("2 roles changed: explorer, implementer");
    await userEvent.click(within(table).getByRole("button", { name: "Save" }));
    expect(postRolesEditMock).toHaveBeenCalledTimes(1);
    expect(Object.keys(postRolesEditMock.mock.calls[0][0])).toEqual(["explorer", "implementer"]);
    await waitFor(() => expect(fetchRolesMock).toHaveBeenCalledTimes(2));
  });

  it("Agent roles shows a refused save", async () => {
    postRolesEditMock.mockReset().mockRejectedValue(new Error("role pair \"x\" is not <model>/<type>[@<effort>]"));
    render(<AgentRoles />);
    const table = await screen.findByRole("region", { name: "Current roles" });
    await userEvent.click(within(table).getByRole("button", { name: "Choose main pairs for implementer" }));
    await userEvent.click(within(table).getByRole("checkbox", { name: "opus/flow-low" }));
    await userEvent.click(within(table).getByRole("button", { name: "Save" }));
    expect(await within(table).findByRole("alert")).toHaveTextContent("is not <model>/<type>");
  });

  it("Agent roles is a nav route", async () => {
    window.history.replaceState(null, "", "#/agent-roles");
    render(<App />);
    expect(await screen.findByRole("link", { name: "Agent roles" })).toHaveAttribute("aria-current", "page");
    expect(await screen.findByRole("region", { name: "Current roles" })).toBeInTheDocument();
  });
});

describe("nested guards", () => {
  it("sit under their caller, and stay in place when the caller is not in the panel", () => {
    const g = (guard: string, parent?: string) => ({ guard, parent, runs: 1, fired: 0, cannotAnswer: 0, medianMs: 1, totalMs: 1, lastFiredAt: null });
    const rows = [g("a"), g("kid", "b"), g("b"), g("orphan", "elsewhere"), g("grandkid", "kid")];
    expect(nested(rows).map((r) => r.guard)).toEqual(["a", "b", "kid", "grandkid", "orphan"]);
  });

  it("move into their caller's panel", () => {
    const g = (guard: string, parent?: string) => ({ guard, parent, runs: 1, fired: 0, cannotAnswer: 0, medianMs: 1, totalMs: 1, lastFiredAt: null });
    const rows = [g("run-verify-list"), g("project-get", "run-verify-list"), g("loner", "absent")];
    expect(panelOf(rows[1], rows)).toBe(panelOf(rows[0], rows));
    expect(panelOf(rows[2], rows)).toBe("Other");
  });
});
