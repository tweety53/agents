// One test per view asserting it renders the actual numbers a fixture
// response carries (never merely that the component mounted -- the
// highest-risk vacuous shape this task's own instructions name), plus the
// absence-vs-zero distinction and the state-board-is-default-route
// requirement.
import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, onTestFinished, vi } from "vitest";
import type {
  TrendChange,
  DecisionRow,
  HealthViewSlug,
  StageLeaderboardRow,
  StatsResponse,
  StatsViewSlug,
  ViewName,
  WeeklyModelEffortViewSlug,
} from "../api";
import { App } from "../App";
import { DynamicsChart } from "../components/DynamicsChart";
import { Decisions } from "./Decisions";
import { Experiments } from "./Experiments";
import { FlowHealth, speedUpCandidates } from "./FlowHealth";
import { ReviewDynamics } from "./ReviewDynamics";
import { Reviewers } from "./Reviewers";
import { SelfReview } from "./SelfReview";
import { StageLeaderboard } from "./StageLeaderboard";
import { StateBoard } from "./StateBoard";
import { Trend } from "./Trend";
import { WeeklyCosts } from "./WeeklyCosts";

const { fetchStatsViewMock, fetchProjectsMock, fetchRolesMock, fetchExperimentsMock, fetchExperimentStatsMock } = vi.hoisted(() => ({
  fetchStatsViewMock: vi.fn(),
  fetchProjectsMock: vi.fn(),
  fetchRolesMock: vi.fn(),
  fetchExperimentsMock: vi.fn(),
  fetchExperimentStatsMock: vi.fn(),
}));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return {
    ...actual,
    fetchStatsView: fetchStatsViewMock,
    fetchProjects: fetchProjectsMock,
    fetchRoles: fetchRolesMock,
    fetchExperiments: fetchExperimentsMock,
    fetchExperimentStats: fetchExperimentStatsMock,
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
    markers: [
      { sha: "abc1234def:skills/flow/primary-reviewer-prompt.md", at: "2026-01-13T10:00:00Z", kind: "edit", area: "review", slot: "primary", target: "skills/flow/primary-reviewer-prompt.md", summary: "re-read every writer", caseId: null },
    ],
    lastPromptChangeAt: "2026-01-13T10:00:00Z",
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
  Exclude<ViewName, "flow-health" | "experiments"> | HealthViewSlug | WeeklyModelEffortViewSlug,
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
  "weekly-model-effort-costs": envelope("weekly-model-effort-costs", WEEKLY_MODEL_ROWS),
  reviewers: { ...envelope("reviewers", REVIEWER_ROWS), projects: REVIEWER_PROJECTS },
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
  "guard-health": envelope("guard-health", [
    gh("check-plan-shape", 6794, 3403, 37, 102, 3798824, "2026-11-04T10:00:00Z"),
    gh("check-references", 2246, 110, 0, 618, 2056971, "2026-11-04T09:00:00Z"),
    gh("check-no-bash", 35, 0, 0, 95, 3421, null),
  ]),
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
    expect(cards[1]).not.toHaveTextContent("Working"); // the live dot already says it
    expect(cards[1]).toHaveTextContent("Implementing tasks · 30m");
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
    expect(row).toHaveTextContent("Implementing tasks");
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
    render(<StageLeaderboard period={period} project={undefined} />);
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
    render(<StageLeaderboard period={period} project={undefined} />);
    await screen.findByRole("table");
    expect(calls().map((p) => (p as { projects?: boolean }).projects)).toEqual([true, undefined]);
  });

  it("stage leaderboard Period chip keeps the Project and Command chips", async () => {
    answerLeaderboard();
    const user = userEvent.setup();
    render(<StageLeaderboard period={period} project={undefined} />);
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
    render(<StageLeaderboard period={period} project={undefined} />);
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
    render(<StageLeaderboard period={period} project={undefined} />);
    await screen.findByRole("table");
    const row = (name: string) => within([...document.querySelectorAll<HTMLElement>("tbody tr")].find((tr) => tr.textContent?.includes(name))!);
    row("Implementing tasks").getByText("new");
    row("Running lint and tests").getByText("+$9.50 (+1.9k%)");
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
    render(<StageLeaderboard period={period} project={undefined} />);
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
    render(<StageLeaderboard period={period} project={undefined} />);
    expect(await screen.findByRole("alert")).toHaveTextContent("previous window failed");
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });

  it("stage leaderboard reads not-recorded and unmeasured as Panel does", async () => {
    fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures) =>
      Promise.resolve(view !== "stage-leaderboard" ? fixtures[view] : envelope("stage-leaderboard", leaderboardRows(LEADERBOARD_CURRENT), false)),
    );
    const { unmount } = render(<StageLeaderboard period={period} project={undefined} />);
    expect(await screen.findByTestId("not-recorded")).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    unmount();

    fetchStatsViewMock.mockImplementation((view: keyof typeof fixtures) =>
      Promise.resolve(view !== "stage-leaderboard" ? fixtures[view] : envelope("stage-leaderboard", leaderboardRows(LEADERBOARD_CURRENT), true, true)),
    );
    render(<StageLeaderboard period={period} project={undefined} />);
    expect(await screen.findByTestId("unmeasured")).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });

  it("stage leaderboard project chips name projects with spend", async () => {
    answerLeaderboard();
    render(<StageLeaderboard period={period} project={undefined} />);
    await screen.findByRole("table");
    const chips = within(screen.getByRole("group", { name: "Project" })).getAllByRole("button");
    expect(chips.map((c) => c.textContent)).toEqual(["All", "gymie$1,000", "agents$260"]);
    expect(chips.map((c) => c.getAttribute("title"))).toEqual([null, "gymie-7c1f238a", "agents-a740d89c"]);
  });

  it("stage leaderboard tiles name stages and the biggest riser", async () => {
    answerLeaderboard();
    render(<StageLeaderboard period={period} project={undefined} />);
    await screen.findByRole("table");
    const spend = tile("sl-t-spend");
    spend.getByText("$1,260");
    expect(spend.getByText(/vs the previous 30d/)).toHaveTextContent("+$20.00 (+2%) vs the previous 30d");

    const top = tile("sl-t-top");
    top.getByText("79%");
    const items = top.getAllByRole("listitem");
    expect(items.map((li) => li.textContent)).toEqual([
      "1Implementing tasks /flowflow.sdd-tdd$50040%",
      "2Code review /flowflow.review-panel$30024%",
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
    render(<StageLeaderboard period={period} project={undefined} />);
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
    first.getByText("Implementing tasks");
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
    const { unmount } = render(<StageLeaderboard period={period} project={undefined} />);
    await screen.findByRole("table");
    const riser = tile("sl-t-riser");
    riser.getByText("—");
    riser.getByText("no stage grew vs the previous 30d");
    unmount();

    answer([], []);
    render(<StageLeaderboard period={period} project={undefined} />);
    expect(await screen.findByText("No costed stage runs in this period.")).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });
  it("reviewers ranks by value, severity and false alarms", async () => {
    const user = userEvent.setup();
    render(<Reviewers period={period} project={undefined} />);
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
    render(<Reviewers period={period} project={undefined} />);
    await screen.findByRole("table");
    expect(document.querySelectorAll(".rv-tiles .lb-tile")[1]).toHaveTextContent("most critical: one, two (2 each)");
  });

  it("reviewers shows no tiles when the rows are not recorded or unmeasured", async () => {
    for (const [recorded, unmeasured] of [[false, false], [true, true]]) {
      fetchStatsViewMock.mockResolvedValueOnce({ ...envelope("reviewers", REVIEWER_ROWS, recorded, unmeasured), projects: [] });
      const { unmount } = render(<Reviewers period={period} project={undefined} />);
      await waitFor(() => expect(fetchStatsViewMock).toHaveBeenCalled());
      await screen.findByRole("toolbar");
      await waitFor(() => expect(screen.queryByText(/loading/i)).not.toBeInTheDocument());
      expect(document.querySelector(".rv-tiles")).toBeNull();
      unmount();
      fetchStatsViewMock.mockClear();
    }
  });

  it("reviewers tiles and footer derive from the rows", async () => {
    render(<Reviewers period={period} project={undefined} />);
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
    const { unmount } = render(<Reviewers period={period} project={undefined} />);
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
    render(<Reviewers period={period} project={undefined} />);
    expect(await screen.findByText("No reviewer dispatches in this period.")).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(document.querySelector(".rv-tiles")).toBeNull();
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

  // A decisions row: the sdd fixture with no findings, fix rounds or fallbacks, then the overrides.
  const dr = (o: Partial<DecisionRow>): DecisionRow => ({
    ...(fixtures.decisions.rows as DecisionRow[])[1],
    critical: 0,
    important: 0,
    minor: 0,
    fixRounds: 0,
    fallbacks: 0,
    ...o,
  });
  const renderDecisions = async (rows: DecisionRow[]) => {
    fetchStatsViewMock.mockResolvedValueOnce(envelope("decisions", rows));
    render(<Decisions period={period} project={undefined} />);
    await screen.findByRole("table");
  };
  // The choice rows under a class's group row, up to the next group row.
  const classRows = (cls: string) => {
    const out: HTMLElement[] = [];
    let inside = false;
    for (const tr of Array.from(document.querySelectorAll<HTMLElement>("tbody tr"))) {
      if (tr.classList.contains("lb-group")) inside = text(tr).startsWith(cls);
      else if (inside && tr.classList.contains("lb-row")) out.push(tr);
    }
    return out;
  };
  const cells = (tr: HTMLElement) => Array.from(tr.querySelectorAll("td"));
  // A cell's text without its caret.
  const text = (el: Element) => el.textContent!.replace("▾", "");
  const choiceRow = (cls: string, name: string) => classRows(cls).find((tr) => text(cells(tr)[0]).startsWith(name))!;

  it("decisions folds every inline spelling into one implementer choice", async () => {
    const reg = (execution: string, implementerModel: string, implementerEffort = "") => dr({ class: "regular", grouping: "static", execution, implementerModel, implementerEffort });
    await renderDecisions([
      reg("inline", "skipped — inline"),
      reg("inline", "skipped - inline"),
      reg("inline", "inline"),
      reg("sdd", "skipped — inline"),
      reg("sdd", "opus"),
      reg("sdd", "default"),
      dr({ class: "", execution: "sdd", implementerModel: "sonnet", implementerEffort: "medium" }),
      reg("", ""),
    ]);
    const regular = classRows("regular").map((tr) => text(cells(tr)[0]));
    expect(regular).toEqual(["inline", "—few runs", "defaultfew runs", "opusfew runs"]);
    expect(cells(choiceRow("regular", "inline"))[2]).toHaveTextContent("4");
    expect(classRows("unclassified").map((tr) => text(cells(tr)[0]))).toEqual(["sonnet · mediumfew runs"]);
  });

  it("decisions pairs implementer choice with review grouping and takes medians", async () => {
    const big = (o: Partial<DecisionRow>) => dr({ class: "big", implementerModel: "opus", implementerEffort: "medium", grouping: "free", ...o });
    await renderDecisions([
      big({ change: "a", costUsd: 20, wallClockSeconds: 600, critical: 1, fixRounds: 2 }),
      big({ change: "b", costUsd: 1, wallClockSeconds: 3600, important: 2, fixRounds: 0 }),
      big({ change: "c", costUsd: 9, wallClockSeconds: 1200, minor: 3, fixRounds: 1 }),
      big({ change: "d", grouping: "static", costUsd: 5 }),
    ]);
    const rows = classRows("big");
    expect(rows).toHaveLength(2);
    expect(cells(rows[0]).map(text)).toEqual(["opus · medium", "free", "3", "$9.00", "20m", "7.0", "1.00", "0"]);
    expect(cells(rows[1])[0]).toHaveTextContent("opus · mediumfew runs");
    expect(cells(rows[1])[1]).toHaveTextContent("static");
    const foot = cells(document.querySelector<HTMLElement>("tfoot tr")!);
    expect(foot[0]).toHaveTextContent("All runs");
    expect(foot[2]).toHaveTextContent("4");
  });

  // A: opus · medium free ×3, B: opus · low static ×3, C: sonnet · medium free ×1 -- the cheapest and fastest.
  const markingRows = () => {
    const big = (o: Partial<DecisionRow>) => dr({ class: "big", ...o });
    const a = (i: number) => big({ change: `a${i}`, implementerModel: "opus", implementerEffort: "medium", grouping: "free", costUsd: 10, wallClockSeconds: 1200, important: 1 });
    const b = (i: number) => big({ change: `b${i}`, implementerModel: "opus", implementerEffort: "low", grouping: "static", costUsd: 12, wallClockSeconds: 600, minor: 1 });
    const reg = (i: number, effort: string, cost: number) => dr({ change: `r${i}`, class: "regular", implementerModel: "opus", implementerEffort: effort, costUsd: cost });
    return [
      a(1), a(2), a(3), b(1), b(2), b(3),
      big({ change: "c", implementerModel: "sonnet", implementerEffort: "medium", grouping: "free", costUsd: 2, wallClockSeconds: 60 }),
      reg(1, "high", 3), reg(2, "high", 3), reg(3, "high", 3), reg(4, "low", 1),
    ];
  };

  it("decisions marks best and worst only among rows of three or more runs", async () => {
    await renderDecisions(markingRows());
    expect(cells(choiceRow("big", "opus · medium"))[3]).toHaveClass("rv-best");
    expect(cells(choiceRow("big", "opus · low"))[3]).toHaveClass("rv-worst");
    const c = choiceRow("big", "sonnet · medium");
    expect(cells(c)[3]).not.toHaveClass("rv-best");
    expect(cells(c)[3]).not.toHaveClass("rv-worst");
    expect(cells(c)[0]).toHaveTextContent("few runs");
    for (const tr of classRows("regular")) for (const td of cells(tr)) expect(td.className).not.toMatch(/rv-(best|worst)/);
  });

  it("decisions tiles name each class's picks or say no choice has 3 runs yet", async () => {
    const small = (i: number) => dr({ change: `s${i}`, class: "small" });
    await renderDecisions([...markingRows(), small(1), small(2), dr({ change: "m", class: "micro" })]);
    const tile = (cls: string) => Array.from(document.querySelectorAll<HTMLElement>(".lb-tile")).find((t) => t.textContent!.startsWith(`${cls} · cheapest`))!;
    const big = tile("big");
    expect(big.querySelector(".lb-val")).toHaveTextContent("$10.00");
    const notes = Array.from(big.querySelectorAll(".sl-note")).map((n) => n.textContent);
    expect(notes).toEqual(["opus · medium · free median per run", "fastest: opus · low · static 10m", "fewest findings: opus · low · static 1.0 / run"]);
    expect(tile("small")).toHaveTextContent("no choice has 3 runs yet · 2 runs");
    expect(tile("micro")).toHaveTextContent(/no choice has 3 runs yet · 1 run$/);
  });

  it("decisions chips narrow by project, class and review grouping", async () => {
    const user = userEvent.setup();
    const reqs = () => fetchStatsViewMock.mock.calls.filter((c) => c[0] === "decisions").map((c) => c[1] as { from: Date; to: Date; project?: string });
    const day = 24 * 60 * 60 * 1000;
    const now = BOARD_NOW.getTime();
    const zeta = "zeta-0a1b2c3d";
    const alpha = "alpha-4e5f6a7b";
    const rows = [
      ...[1, 2, 3].map((i) => dr({ change: `z${i}`, project: zeta, class: "big", grouping: "free" })),
      dr({ change: "z4", project: zeta, class: "regular", grouping: "static" }),
      ...[1, 2].map((i) => dr({ change: `a${i}`, project: alpha, class: "regular", grouping: "free" })),
    ];
    fetchStatsViewMock.mockResolvedValueOnce(envelope("decisions", rows)).mockResolvedValueOnce(envelope("decisions", rows));
    render(<Decisions period={period} project={undefined} />);
    await screen.findByRole("table");
    expect(reqs()).toHaveLength(1);
    expect(reqs()[0]).not.toHaveProperty("project");
    expect([reqs()[0].from.getTime(), reqs()[0].to.getTime()]).toEqual([now - 30 * day, now]);
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();

    const group = (name: string) => within(screen.getByRole("group", { name }));
    const chips = (name: string) => group(name).getAllByRole("button").map((b) => b.textContent);
    const pressed = (g: string, name: string | RegExp) => expect(group(g).getByRole("button", { name })).toHaveAttribute("aria-pressed", "true");
    expect(chips("Project")).toEqual(["All", "zeta4", "alpha2"]);
    expect(group("Project").getByRole("button", { name: /^zeta/ })).toHaveAttribute("title", zeta);

    await user.click(group("Project").getByRole("button", { name: /^zeta/ }));
    expect(chips("Class")).toEqual(["All", "regular1", "big3"]);
    await user.click(group("Class").getByRole("button", { name: /^regular/ }));
    pressed("Project", /^zeta/);
    expect(document.querySelectorAll("tbody tr.lb-group")).toHaveLength(1);
    expect(classRows("regular")).toHaveLength(1);
    expect(chips("Review grouping")).toEqual(["All", "static1"]);

    await user.click(group("Period").getByRole("button", { name: "7d" }));
    await waitFor(() => expect([reqs().at(-1)!.from.getTime(), reqs().at(-1)!.to.getTime()]).toEqual([now - 7 * day, now]));
    await screen.findByRole("table");
    pressed("Project", /^zeta/);
    pressed("Class", /^regular/);

    await user.click(screen.getByRole("button", { name: "Clear filters" }));
    pressed("Period", "30d");
    for (const g of ["Project", "Class", "Review grouping"]) pressed(g, "All");
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();
  });

  it("decisions expands a row to its runs and collapses a class", async () => {
    const user = userEvent.setup();
    await renderDecisions([
      dr({ project: "agents-a740d89c", change: "kan-7-x", class: "big", recordedAt: "2026-01-10T23:30:00Z", costUsd: 4.5, wallClockSeconds: 5400, critical: 1, important: 1, minor: 2, fixRounds: 3 }),
    ]);
    const row = choiceRow("big", "opus · high");
    expect(row).toHaveAttribute("aria-expanded", "false");
    await user.click(row);
    expect(row).toHaveAttribute("aria-expanded", "true");
    const run = document.querySelector<HTMLElement>("tr.dc-run")!;
    expect(within(run).getByRole("link", { name: "kan-7-x" })).toHaveAttribute("href", "#/run/agents-a740d89c/kan-7-x");
    expect(cells(run).map((c) => c.textContent)).toEqual(["kan-7-x agents · 2026-01-10", "", "", "$4.50", "1h 30m", "16 1 · 1 · 2", "3", ""]);

    await user.click(document.querySelector<HTMLElement>("tbody tr.lb-group")!);
    expect(classRows("big")).toHaveLength(0);
    expect(document.querySelector("tr.dc-run")).toBeNull();
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
    expect(r).toHaveTextContent("Implementing tasks · 3m · Edit: stats/web/src/lib/phase.ts");
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
    expect(within(f).getByText("Implementing")).toHaveClass("lb-pill");
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
    const planned = { ...flow, currentStage: undefined, currentStageStartedAt: undefined, lastStage: "flow.decide", lastStageEndedAt: minutesAgo(4), activity: "idle", activitySince: minutesAgo(4), sessionSeenAt: minutesAgo(3) };
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
    render(<WeeklyCosts period={period} project={undefined} />);
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
    render(<WeeklyCosts period={period} project={undefined} />);
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
    render(<Trend period={period} project={undefined} />);
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
    render(<Trend period={period} project={undefined} />);
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
    render(<Trend period={period} project={undefined} />);
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
  const answerGuards = (rows: ReturnType<typeof gh>[], projects: { projectKey: string; runs: number }[] = []) =>
    fetchStatsViewMock.mockResolvedValueOnce({ ...envelope("guard-health", rows), projects });
  const names = () => [...document.querySelectorAll("tbody tr")].map((tr) => tr.querySelector(".rv-slot")!.textContent);
  const pills = () =>
    [...document.querySelectorAll("tbody tr")].filter((tr) => tr.querySelector(".lb-pill.fh-speed")).map((tr) => tr.querySelector(".rv-slot")!.textContent);
  const title = () => document.querySelector(".lb-panel-h span")!.textContent;
  const foot = () => [...document.querySelectorAll("tfoot td")].map((td) => td.textContent);
  const cells = (guard: string) => {
    const tr = [...document.querySelectorAll("tbody tr")].find((r) => r.querySelector(".rv-slot")!.textContent === guard)!;
    return [...tr.querySelectorAll("td")];
  };
  const group = (name: string) => within(screen.getByRole("group", { name }));
  const tileButton = (name: RegExp) => screen.getByRole("button", { name });

  it("speed-up candidates are the fewest guards reaching 80% of run time", () => {
    const six = [
      gh("zeta", 1, 0, 0, 300, 15000, null),
      gh("a", 1, 0, 0, 100, 40000, null),
      gh("f", 1, 0, 0, 600, 2000, null),
      gh("alpha", 1, 0, 0, 300, 15000, null),
      gh("b", 1, 0, 0, 200, 25000, null),
      gh("e", 1, 0, 0, 500, 3000, null),
    ];
    expect(speedUpCandidates(six)).toEqual(new Set(["a", "b", "alpha"]));
    expect(speedUpCandidates(six.map((g) => ({ ...g, totalMs: 0 })))).toEqual(new Set());
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
    render(<FlowHealth period={period} project={undefined} />);
    await screen.findByRole("table");
    const sorted = () => [...document.querySelectorAll("th[aria-sort]")].map((th) => [th.textContent, th.getAttribute("aria-sort")]);
    const expectRank = (order: string[], heading: string, header: string) => {
      expect(names()).toEqual(order);
      expect([...document.querySelectorAll("tbody tr")].map((tr) => tr.querySelector("td")!.textContent)).toEqual(["1", "2", "3", "4", "5"]);
      expect(title()).toBe(heading);
      expect(sorted()).toEqual([[header, "descending"]]);
      expect(pills()).toEqual(["echo"]);
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

  it("flow health tiles name each set and toggle the table to it", async () => {
    const user = userEvent.setup();
    answerGuards([
      gh("fire-big", 1000, 50, 10, 100, 100000, "2026-01-10T00:00:00Z"),
      gh("fire-small", 100, 5, 0, 900, 90000, "2026-01-11T00:00:00Z"),
      // Under the 20-run floor: a higher cannot-answer share than ca-worst, never named worst.
      gh("ca-tiny", 10, 0, 5, 8000, 80000, null),
      gh("ca-worst", 100, 2, 20, 700, 70000, "2026-01-12T00:00:00Z"),
      gh("quiet-1", 500, 0, 0, 120, 60000, null),
      gh("quiet-2", 400, 0, 0, 2, 1000, null),
      gh("quiet-3", 300, 0, 0, 3, 1000, null),
      gh("quiet-4", 200, 0, 0, 5, 1000, null),
    ]);
    render(<FlowHealth period={period} project={undefined} />);
    await screen.findByRole("table");
    const tileTexts = [
      "Catching · fired at least once3 of 8most: fire-big 50 fires",
      "Never fired · ran, caught nothing5quiet-1 (500 runs), quiet-2 (400 runs), quiet-3 (300 runs), +2 more",
      "Cannot answer · share of runs1.3%worst: ca-worst 20%",
      "Speed-up candidates · 80% of run time4 of 7 min totalfire-big 25%, fire-small 22%, ca-tiny 20%, +1 more",
    ];
    const tiles = () => [...document.querySelectorAll<HTMLElement>(".rv-tiles .lb-tile")];
    const expectTiles = () => tiles().forEach((t, i) => expect(t).toHaveTextContent(tileTexts[i]));
    expect(tiles()).toHaveLength(4);
    expectTiles();
    expect(tiles()[1].querySelector(".lb-val")).toHaveClass("rv-worst");
    expect(tiles().map((t) => t.getAttribute("aria-pressed"))).toEqual(["false", "false", "false", "false"]);
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();

    await user.click(tileButton(/^Never fired ·/));
    expect(tileButton(/^Never fired ·/)).toHaveAttribute("aria-pressed", "true");
    expect(names()).toEqual(["quiet-1", "quiet-2", "quiet-3", "quiet-4", "ca-tiny"]);
    expect(title()).toBe("Never-fired guards, most fires first");
    expect(foot()[1]).toBe("5 guards");
    expect(foot()[8]).toBe("35%");
    expect(screen.getByRole("button", { name: "Clear filters" })).toBeInTheDocument();
    expectTiles();

    await user.click(tileButton(/^Speed-up candidates ·/));
    expect(tileButton(/^Speed-up candidates ·/)).toHaveAttribute("aria-pressed", "true");
    expect(tileButton(/^Never fired ·/)).toHaveAttribute("aria-pressed", "false");
    expect(names()).toEqual(pills());
    expect(names().sort()).toEqual(["ca-tiny", "ca-worst", "fire-big", "fire-small"]);
    await user.click(tileButton(/^Speed-up candidates ·/));
    expect(names()).toHaveLength(8);

    await user.click(group("Rank by").getByRole("button", { name: "Run time" }));
    await user.click(tileButton(/^Catching ·/));
    expect(names()).toEqual(["fire-big", "fire-small", "ca-worst"]);
    await user.click(screen.getByRole("button", { name: "Clear filters" }));
    expect(tileButton(/^Catching ·/)).toHaveAttribute("aria-pressed", "false");
    expect(group("Rank by").getByRole("button", { name: "Run time" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.queryByRole("button", { name: "Clear filters" })).not.toBeInTheDocument();
    expect(names()).toHaveLength(8);

    await user.click(tileButton(/^Cannot answer ·/));
    expect(names().sort()).toEqual(["ca-tiny", "ca-worst", "fire-big"]);
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
    render(<FlowHealth period={period} project={undefined} />);
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
    const { unmount } = render(<FlowHealth period={period} project={undefined} />);
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
    const second = render(<FlowHealth period={period} project={undefined} />);
    await screen.findByRole("table");
    expect(text("c-s").slice(6, 8)).toEqual(["1.0 s", "1 min"]);
    expect(text("c-m")[7]).toBe("1 h 0 min");
    expect(text("c-h")[7]).toBe("2 h 0 min");
    second.unmount();

    answerGuards([gh("solo", 10, 1, 0, 450, 1500, "2026-08-14T23:30:00Z")]);
    render(<FlowHealth period={period} project={undefined} />);
    await screen.findByRole("table");
    expect(foot()[1]).toBe("1 guard");
    expect(foot()[8]).toBe("100%");
  });

  it("flow health shows no tiles and the empty message with no guard runs", async () => {
    answerGuards([]);
    const { unmount } = render(<FlowHealth period={period} project={undefined} />);
    expect(await screen.findByText("No guard runs in this period.")).toBeInTheDocument();
    expect(document.querySelector(".lb-tile")).toBeNull();
    unmount();

    fetchStatsViewMock.mockResolvedValueOnce({ ...envelope("guard-health", fixtures["guard-health"].rows, false), projects: [] });
    render(<FlowHealth period={period} project={undefined} />);
    expect(await screen.findByTestId("not-recorded")).toBeInTheDocument();
    expect(document.querySelector(".lb-tile")).toBeNull();
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

describe("Experiments", () => {
  const roles = [
    {
      role: "main", pairs: [], marker: "", group: "Session", does: "The session you talk to: runs the pipeline and sends out every other role.",
      override: { pairs: ["opus@medium"], experimentId: 2 },
    },
    { role: "explorer", pairs: ["opus/Explore@low", "opus/Explore@medium"], marker: "", group: "Planning", does: "Reads the codebase to answer a planner's question." },
    {
      role: "plan-designer", pairs: ["opus/Plan@low", "opus/Plan@medium"], marker: "", group: "Planning",
      does: "Proposes a design or plan for the planner to choose from.", override: { pairs: ["opus/Plan@xhigh"], experimentId: 3 },
    },
    { role: "implementer", pairs: ["opus/flow-low"], marker: "", group: "Implementation", does: "Writes the code for a group of tasks, tests first." },
    { role: "implementer", pairs: ["opus/flow-high"], marker: "No-progress escalation", group: "Implementation", does: "" },
  ];
  const experiments = [
    { id: 3, name: "plan-designer on xhigh", cells: { "plan-designer": ["opus/Plan@xhigh"] }, startedAt: "2026-10-30T09:00:00Z", summary: "", status: "open" },
    { id: 2, name: "main agent opus@medium", cells: { main: ["opus@medium"] }, startedAt: "2026-10-28T09:00:00Z", summary: "", status: "open" },
    {
      id: 1, name: "implementer sonnet/flow-low", cells: { implementer: ["sonnet/flow-low"] }, startedAt: "2026-10-01T09:00:00Z",
      closedAt: "2026-10-21T09:00:00Z", verdict: "negative", summary: "Negative. Table unchanged; experiment closed.", status: "closed",
    },
  ];
  const week = (w: string, value: number) => ({ week: `${w}T00:00:00Z`, value });
  const stats = (id: number, metrics: { name: string; before: number | null; after: number | null }[]) => ({
    experiment: experiments.find((e) => e.id === id),
    beforeFrom: "2026-10-23T09:00:00Z",
    end: "2026-11-06T09:00:00Z",
    runs: { before: 14, after: 4 },
    excluded: 0,
    metrics: metrics.map((m) => ({ ...m, series: [week("2026-10-19", 2.9), week("2026-10-26", 2.4), week("2026-11-02", 1.8)] })),
  });

  beforeEach(() => {
    fetchRolesMock.mockReset().mockResolvedValue(roles);
    fetchExperimentsMock.mockReset().mockResolvedValue(experiments);
    fetchExperimentStatsMock.mockReset().mockImplementation((id: number) =>
      Promise.resolve(
        id === 3
          ? stats(3, [
              { name: "reviewRounds", before: 2.6, after: 1.9 },
              { name: "tokens", before: 412000, after: 488000 },
            ])
          : stats(id, [{ name: "reviewRounds", before: 2.2, after: 3.1 }]),
      ),
    );
  });

  it("Experiments groups the roles, one row per role, folding escalations and linking overrides", async () => {
    render(<Experiments />);
    const table = await screen.findByRole("region", { name: "Current roles" });
    expect(within(table).getAllByRole("columnheader").slice(0, 3).map((h) => h.textContent)).toEqual(["Role", "Does", "Allowed pairs"]);
    const rows = within(table).getAllByRole("row");
    expect(rows.map((r) => r.querySelector("th, td")?.textContent)).toEqual([
      "Role", "Session", "main", "Planning", "explorer", "plan-designer", "Implementation", "implementer",
    ]);
    for (const i of [1, 3, 6]) expect(rows[i].querySelector("th")).toHaveAttribute("colspan", "3");
    const designer = rows[5];
    expect(designer.querySelector("s")).toHaveTextContent("opus/Plan@low, opus/Plan@medium");
    expect(within(designer).getByText("opus/Plan@xhigh").closest("code")).toHaveClass("exp");
    // Each pair is its own unbreakable span, so a long list wraps between pairs, never at a hyphen (kan-1102 self-review, F19).
    expect([...designer.querySelectorAll("s .pair")].map((p) => p.textContent)).toEqual(["opus/Plan@low", "opus/Plan@medium"]);
    expect(within(designer).getByRole("link", { name: "#3" })).toBeInTheDocument();
    expect(within(designer).getByText("Proposes a design or plan for the planner to choose from.")).toBeInTheDocument();
    const implementer = rows[7];
    expect(within(implementer).getByText("opus/flow-low")).toBeInTheDocument();
    expect(implementer.querySelector(".esc")).toHaveTextContent("opus/flow-high with No-progress escalation");
    expect(rows[2].querySelector("s")).toHaveTextContent("session's own pair");
    expect(within(rows[4]).queryByRole("link")).toBeNull();
    expect(rows[4].querySelector("s")).toBeNull();
    expect(within(table).queryByText("Requires marker")).toBeNull();
    expect(within(table).queryByText("Experiment")).toBeNull();
    expect(table).not.toHaveTextContent("From skills/flow-contracts/model-policy.md");

    await userEvent.click(within(rows[2]).getByRole("link", { name: "#2" }));
    expect(screen.getByRole("button", { name: /main agent opus@medium/ })).toHaveClass("on");
  });

  it("Experiments marks each experiment open or closed and selects the newest", async () => {
    render(<Experiments />);
    const list = await screen.findByRole("region", { name: "Experiments" });
    const first = await within(list).findByRole("button", { name: /plan-designer on xhigh/ });
    expect(first).toHaveClass("on");
    expect(within(first).getByText("open")).toHaveClass("pill", "open");
    expect(first).toHaveTextContent("since 2026-10-30");
    const closed = within(list).getByRole("button", { name: /implementer sonnet\/flow-low/ });
    expect(within(closed).getByText("closed")).toHaveClass("pill", "closed");
    expect(closed).toHaveTextContent("since 2026-10-01 · closed 2026-10-21");
  });

  it("Experiments lists open experiments before closed ones, whatever order the API returns", async () => {
    // The closed one started last, so only an open-first sort puts it last.
    fetchExperimentsMock.mockReset().mockResolvedValue([{ ...experiments[2], startedAt: "2026-10-31T09:00:00Z" }, experiments[1], experiments[0]]);
    render(<Experiments />);
    const list = await screen.findByRole("region", { name: "Experiments" });
    const names = (await within(list).findAllByRole("button")).map((b) => b.firstChild?.textContent);
    expect(names).toEqual(["plan-designer on xhigh", "main agent opus@medium", "implementer sonnet/flow-low"]);
  });

  it("Experiments keeps the stat panel's delta line on one line", async () => {
    render(<Experiments />);
    const rounds = await screen.findByRole("region", { name: "Review rounds / change" });
    expect(within(rounds).getByText("▼ 27% vs 2.6 before")).toHaveStyle({ whiteSpace: "nowrap" });
  });

  it("Experiments shows the selected experiment's override, stat panels, charts and summary", async () => {
    render(<Experiments />);
    const override = await screen.findByRole("region", { name: "Override" });
    const row = within(override).getAllByRole("row")[1];
    expect(within(row).getAllByRole("cell").map((c) => c.textContent)).toEqual(["plan-designer", "opus/Plan@low, opus/Plan@medium", "opus/Plan@xhigh"]);
    expect(await within(override).findByText("Runs counted: 14 before · 4 after start")).toBeInTheDocument();

    const rounds = screen.getByRole("region", { name: "Review rounds / change" });
    expect(within(rounds).getByText("1.9")).toHaveClass("v");
    expect(within(rounds).getByText("▼ 27% vs 2.6 before")).toHaveClass("d", "down");
    const tokens = screen.getByRole("region", { name: "Tokens / change (k)" });
    expect(within(tokens).getByText("488")).toBeInTheDocument();
    expect(within(tokens).getByText("▲ 18% vs 412 before")).toHaveClass("d", "up");

    const trend = screen.getByRole("region", { name: "Weekly trend — before vs after start" });
    const charts = within(trend).getAllByRole("img");
    expect(charts).toHaveLength(2);
    expect(charts[0]).toHaveAttribute("aria-label", "Review rounds / change, weekly, before vs after 2026-10-30");
    expect(charts[0].querySelector("[data-testid=start-marker]")).not.toBeNull();
    // Lower is better: rounds fell, so its after mean is --delta-down;
    // tokens rose, so theirs is --delta-up.
    expect(charts[0].querySelector("[data-testid=mean-before]")).toHaveAttribute("stroke", "var(--text-muted)");
    expect(charts[0].querySelector("[data-testid=mean-after]")).toHaveAttribute("stroke", "var(--delta-down)");
    expect(charts[1].querySelector("[data-testid=mean-after]")).toHaveAttribute("stroke", "var(--delta-up)");
    const points = charts[0].querySelectorAll("circle.pt");
    expect(points).toHaveLength(3);
    expect(points[0]).toHaveAttribute("tabindex", "0");
    expect(points[0].querySelector("title")).toHaveTextContent("10-19: 2.90");
    for (const legend of ["weekly value", "mean before", "mean after (better)", "mean after (worse)", "experiment window"]) {
      expect(within(trend).getByText(legend)).toBeInTheDocument();
    }

    expect(within(screen.getByRole("region", { name: "Summary" })).getByText("Open — the closing agent writes the summary.")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /implementer sonnet\/flow-low/ }));
    expect(await within(screen.getByRole("region", { name: "Summary" })).findByText("Negative. Table unchanged; experiment closed.")).toBeInTheDocument();
    expect(fetchExperimentStatsMock).toHaveBeenLastCalledWith(1);
  });

  it("Experiments draws an unchanged metric as neutral, with no arrow and no delta colour", async () => {
    fetchExperimentStatsMock.mockReset().mockResolvedValue(stats(3, [{ name: "reviewRounds", before: 0, after: 0 }, { name: "verifyReds", before: 2, after: 2 }]));
    render(<Experiments />);
    for (const [name, text] of [["Review rounds / change", "= vs 0.0 before"], ["Verify reds", "= vs 2.0 before"]]) {
      const panel = await screen.findByRole("region", { name });
      const d = within(panel).getByText(text);
      expect(d).toHaveClass("d");
      expect(d).not.toHaveClass("up");
      expect(d).not.toHaveClass("down");
    }
    for (const chart of within(screen.getByRole("region", { name: "Weekly trend — before vs after start" })).getAllByRole("img")) {
      expect(chart.querySelector("[data-testid=mean-after]")).toHaveAttribute("stroke", "var(--text-muted)");
    }
  });

  it("Experiments anchors a chart's end x-axis labels inside the plot", async () => {
    render(<Experiments />);
    const trend = await screen.findByRole("region", { name: "Weekly trend — before vs after start" });
    for (const chart of await within(trend).findAllByRole("img")) {
      const labels = [...chart.querySelectorAll("text")].filter((t) => /^\d\d-\d\d$/.test(t.textContent ?? ""));
      expect(labels[0]).toHaveAttribute("text-anchor", "start");
      expect(labels[labels.length - 1]).toHaveAttribute("text-anchor", "end");
    }
  });

  it("Experiments draws a mean-after line with no before mean as neutral", async () => {
    fetchExperimentStatsMock.mockReset().mockResolvedValue(stats(3, [{ name: "tokens", before: null, after: 60000 }]));
    render(<Experiments />);
    const trend = await screen.findByRole("region", { name: "Weekly trend — before vs after start" });
    const lines = (await within(trend).findAllByRole("img"))
      .map((c) => c.querySelector("[data-testid=mean-after]"))
      .filter(Boolean);
    expect(lines.length).toBeGreaterThan(0);
    for (const l of lines) expect(l).toHaveAttribute("stroke", "var(--text-muted)");
  });

  it("Experiments reads — where no run follows the start, and invites an open when none exists", async () => {
    fetchExperimentStatsMock.mockReset().mockResolvedValue({ ...stats(3, [{ name: "reviewRounds", before: 2.6, after: null }]), runs: { before: 14, after: 0 } });
    const { unmount } = render(<Experiments />);
    const rounds = await screen.findByRole("region", { name: "Review rounds / change" });
    expect(within(rounds).getByText("—")).toHaveClass("v");
    unmount();

    fetchExperimentsMock.mockReset().mockResolvedValue([]);
    render(<Experiments />);
    const list = await screen.findByRole("region", { name: "Experiments" });
    expect(await within(list).findByText(/No experiments yet — open one with/)).toHaveTextContent("No experiments yet — open one with flow experiment open.");
    expect(within(list).getByText("flow experiment open").tagName).toBe("CODE");
  });

  it("Experiments is a nav route labelled Experiments", async () => {
    window.history.replaceState(null, "", "#/experiments");
    render(<App />);
    expect(await screen.findByRole("link", { name: "Experiments" })).toHaveAttribute("aria-current", "page");
    expect(await screen.findByRole("region", { name: "Current roles" })).toBeInTheDocument();
  });
});
