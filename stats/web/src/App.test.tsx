// App.test.tsx covers the hash router's own contract: which view or run
// route a hash resolves to, and (task 2) the period the URL's query
// carries or falls back to. `fetchStatsView` and `listStageRuns` are
// mocked at the module boundary -- the same pattern views.test.tsx and
// RunDetail.test.tsx already use -- so these tests assert this file's own
// routing logic, never the network layer.
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";
import type { CostPerChangeRow, ListStageRunsResponse, StatsResponse } from "./api";

const { fetchStatsViewMock, listStageRunsMock } = vi.hoisted(() => ({
  fetchStatsViewMock: vi.fn(),
  listStageRunsMock: vi.fn(),
}));

vi.mock("./api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./api")>();
  return { ...actual, fetchStatsView: fetchStatsViewMock, listStageRuns: listStageRunsMock };
});

function aggregateEnvelope(rows: CostPerChangeRow[] = []): StatsResponse<CostPerChangeRow[]> {
  return {
    view: "cost-per-change",
    from: "2020-01-01T00:00:00Z",
    to: "2026-01-01T00:00:00Z",
    boundaryConvention: "a stage run is attributed to the period containing its start instant",
    recorded: true,
    unmeasured: false,
    rows,
  };
}

function stageRunsResponse(): ListStageRunsResponse {
  return { total: 0, stageRuns: [] };
}

beforeEach(() => {
  fetchStatsViewMock.mockReset();
  fetchStatsViewMock.mockImplementation((view) => Promise.resolve({ ...aggregateEnvelope(), view }));
  listStageRunsMock.mockReset();
  listStageRunsMock.mockResolvedValue(stageRunsResponse());
});

describe("route resolution ignores the query string (task 1)", () => {
  it("a static view route with a query resolves to that view", async () => {
    window.location.hash = "#/trend?from=2026-01-01T00:00:00Z&to=2026-02-01T00:00:00Z";
    render(<App />);
    expect(await screen.findByRole("heading", { name: "Trend over time" })).toBeInTheDocument();
    window.location.hash = "";
  });

  it("a run route with a query resolves to that run, and the change name does not carry the query", async () => {
    window.location.hash =
      "#/run/kan-16-stats-app/kan-16-stats-app?from=2026-01-01T00:00:00Z&to=2026-02-01T00:00:00Z";
    render(<App />);
    await screen.findByRole("heading", { name: "kan-16-stats-app" });
    expect(listStageRunsMock).toHaveBeenCalledWith(
      expect.objectContaining({ filters: expect.objectContaining({ name: "kan-16-stats-app" }) }),
    );
    window.location.hash = "";
  });

  it("a static view route with no query resolves exactly as today", async () => {
    window.location.hash = "#/trend";
    render(<App />);
    expect(await screen.findByRole("heading", { name: "Trend over time" })).toBeInTheDocument();
    window.location.hash = "";
  });

  it("a run route with no query resolves exactly as today", async () => {
    window.location.hash = "#/run/kan-16-stats-app/kan-16-stats-app";
    render(<App />);
    await screen.findByRole("heading", { name: "kan-16-stats-app" });
    expect(listStageRunsMock).toHaveBeenCalledWith(
      expect.objectContaining({ filters: expect.objectContaining({ name: "kan-16-stats-app" }) }),
    );
    window.location.hash = "";
  });
});

it("no view renders a dashboard bar: every one carries its own chips", async () => {
  for (const [hash, heading] of [["#/weekly-costs", "Weekly costs"], ["#/trend", "Trend over time"]] as const) {
    window.location.hash = hash;
    const { unmount } = render(<App />);
    expect(await screen.findByRole("heading", { name: heading })).toBeInTheDocument();
    expect(screen.queryByRole("toolbar", { name: "Dashboard controls" })).toBeNull();
    unmount();
  }
  window.location.hash = "";
});

it("reviewers carries no dashboard bar", async () => {
  window.location.hash = "#/reviewers";
  render(<App />);
  expect(await screen.findByRole("heading", { name: "Reviewers" })).toBeInTheDocument();
  expect(screen.queryByRole("toolbar", { name: "Dashboard controls" })).toBeNull();
  window.location.hash = "";
});

it("flow health carries no dashboard bar", async () => {
  window.location.hash = "#/flow-health";
  render(<App />);
  expect(await screen.findByRole("heading", { name: "Flow guards" })).toBeInTheDocument();
  expect(screen.queryByRole("toolbar", { name: "Dashboard controls" })).toBeNull();
  window.location.hash = "";
});

describe("the live state board carries its own filters", () => {
  it("renders no dashboard bar on #/state-board", async () => {
    window.location.hash = "#/state-board";
    render(<App />);
    expect(await screen.findByRole("heading", { name: "Live state board" })).toBeInTheDocument();
    expect(screen.queryByRole("toolbar", { name: "Dashboard controls" })).toBeNull();
    expect(screen.queryByLabelText("Period start")).toBeNull();
    window.location.hash = "";
  });
});

it("run-costs route resolves to the run timings page", async () => {
  window.location.hash = "#/run-costs/p/c";
  render(<App />);
  expect(await screen.findByRole("region", { name: "Run timings and costs: p/c" })).toBeInTheDocument();
  expect(screen.queryByRole("toolbar", { name: "Dashboard controls" })).toBeNull();
  window.location.hash = "";
});

it("#/run/<project>/<change> still opens run detail beside the run-costs route", async () => {
  window.location.hash = "#/run/p/c";
  render(<App />);
  expect(await screen.findByRole("region", { name: "Run detail: p/c" })).toBeInTheDocument();
  window.location.hash = "";
});

