// App.test.tsx covers the hash router's own contract: which view or run
// route a hash resolves to, and (task 2) the period the URL's query
// carries or falls back to. `fetchStatsView` and `listStageRuns` are
// mocked at the module boundary -- the same pattern views.test.tsx and
// RunDetail.test.tsx already use -- so these tests assert this file's own
// routing and period logic, never the network layer.
import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
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
      "#/run/kan-16-myflow-stats-app/kan-16-myflow-stats-app?from=2026-01-01T00:00:00Z&to=2026-02-01T00:00:00Z";
    render(<App />);
    await screen.findByRole("heading", { name: "kan-16-myflow-stats-app" });
    expect(listStageRunsMock).toHaveBeenCalledWith(
      expect.objectContaining({ filters: expect.objectContaining({ name: "kan-16-myflow-stats-app" }) }),
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
    window.location.hash = "#/run/kan-16-myflow-stats-app/kan-16-myflow-stats-app";
    render(<App />);
    await screen.findByRole("heading", { name: "kan-16-myflow-stats-app" });
    expect(listStageRunsMock).toHaveBeenCalledWith(
      expect.objectContaining({ filters: expect.objectContaining({ name: "kan-16-myflow-stats-app" }) }),
    );
    window.location.hash = "";
  });
});

describe("the period is read from the URL, falling back to the default (task 2)", () => {
  // A fixed clock, not a second `new Date()` at assertion time -- the
  // latter is flaky by construction (the two calls can straddle a
  // millisecond, or in the worst case a day boundary).
  const NOW = new Date("2026-03-01T00:00:00.000Z");
  const DEFAULT_FROM = new Date("2026-01-30T00:00:00.000Z");

  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(NOW);
  });
  afterEach(() => {
    vi.useRealTimers();
    window.location.hash = "";
  });

  async function renderAndReadPeriod(hash: string): Promise<{ from: Date; to: Date }> {
    window.location.hash = hash;
    render(<App />);
    await vi.waitFor(() => expect(fetchStatsViewMock).toHaveBeenCalled());
    const [, params] = fetchStatsViewMock.mock.calls[0] as [string, { from: Date; to: Date }];
    return { from: params.from, to: params.to };
  }

  it("only `from` is present: falls back to the default period", async () => {
    const { from, to } = await renderAndReadPeriod("#/cost-per-change?from=2026-01-01T00:00:00Z");
    expect(from).toEqual(DEFAULT_FROM);
    expect(to).toEqual(NOW);
  });

  it("only `to` is present: falls back to the default period", async () => {
    const { from, to } = await renderAndReadPeriod("#/cost-per-change?to=2026-01-01T00:00:00Z");
    expect(from).toEqual(DEFAULT_FROM);
    expect(to).toEqual(NOW);
  });

  it("an unparsable bound: falls back to the default period", async () => {
    const { from, to } = await renderAndReadPeriod(
      "#/cost-per-change?from=not-a-date&to=2026-01-01T00:00:00Z",
    );
    expect(from).toEqual(DEFAULT_FROM);
    expect(to).toEqual(NOW);
  });

  it("`to` earlier than `from`: falls back to the default period", async () => {
    const { from, to } = await renderAndReadPeriod(
      "#/cost-per-change?from=2026-02-01T00:00:00Z&to=2026-01-01T00:00:00Z",
    );
    expect(from).toEqual(DEFAULT_FROM);
    expect(to).toEqual(NOW);
  });

  it("no query at all: falls back to the default period", async () => {
    const { from, to } = await renderAndReadPeriod("#/cost-per-change");
    expect(from).toEqual(DEFAULT_FROM);
    expect(to).toEqual(NOW);
  });

  it("a valid pair is applied as given", async () => {
    const { from, to } = await renderAndReadPeriod(
      "#/cost-per-change?from=2026-01-05T00:00:00Z&to=2026-01-10T00:00:00Z",
    );
    expect(from).toEqual(new Date("2026-01-05T00:00:00Z"));
    expect(to).toEqual(new Date("2026-01-10T00:00:00Z"));
  });

  it("a period write updates the URL via replaceState and does not push a history entry", async () => {
    const replaceStateSpy = vi.spyOn(window.history, "replaceState");
    const pushStateSpy = vi.spyOn(window.history, "pushState");

    window.location.hash = "#/cost-per-change";
    render(<App />);
    await vi.waitFor(() => expect(fetchStatsViewMock).toHaveBeenCalled());

    expect(replaceStateSpy).toHaveBeenCalled();
    expect(pushStateSpy).not.toHaveBeenCalled();
    expect(window.location.hash).toContain("from=");
    expect(window.location.hash).toContain("to=");

    replaceStateSpy.mockRestore();
    pushStateSpy.mockRestore();
  });

  // Review finding F1: a hashchange that carries an explicit period (a
  // shared deep link pasted into a tab where the app is already mounted --
  // this fires `hashchange` with no reload, so no fresh `useState`
  // initialiser runs) must not be clobbered by the stale in-memory period.
  // The URL is authoritative on arrival, same as it is on first load.
  it("a hashchange carrying an explicit period wins over the stale in-memory one", async () => {
    window.location.hash = "#/trend";
    render(<App />);
    await vi.waitFor(() => expect(fetchStatsViewMock).toHaveBeenCalled());
    fetchStatsViewMock.mockClear();

    window.location.hash = "#/cost-per-change?from=2026-01-05T00:00:00Z&to=2026-01-10T00:00:00Z";
    window.dispatchEvent(new HashChangeEvent("hashchange"));

    await vi.waitFor(() => expect(fetchStatsViewMock).toHaveBeenCalled());
    const [, params] = fetchStatsViewMock.mock.calls[0] as [string, { from: Date; to: Date }];
    expect(params.from).toEqual(new Date("2026-01-05T00:00:00Z"));
    expect(params.to).toEqual(new Date("2026-01-10T00:00:00Z"));
  });

  // Task 5: nothing else in this file asserts defaultPeriod()'s own span
  // and end -- every fallback case above only checks that App *uses* the
  // default, not what the default actually is. A fixed clock, not a
  // second `new Date()` at assertion time, which would be flaky by
  // construction.
  it("the default period is the preceding 30 days ending now", async () => {
    const { from, to } = await renderAndReadPeriod("#/cost-per-change");
    expect(to).toEqual(NOW);
    expect(from).toEqual(DEFAULT_FROM);
    expect(to.getTime() - from.getTime()).toBe(30 * 24 * 60 * 60 * 1000);
  });
});
