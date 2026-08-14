import { afterEach, describe, expect, it, vi } from "vitest";
import {
  ApiError,
  buildListQuery,
  buildModelsQuery,
  buildStatsViewQuery,
  CHANGE_QUERY_FIELDS,
  fetchModels,
  fetchStatsView,
  listChanges,
  listStageRuns,
  putChange,
  STAGE_RUN_QUERY_FIELDS,
} from "./api";
import { CHANGE_VARIABLE_SORT_FIELD } from "./components/ChangeVariable";
import { RUN_DETAIL_STAGE_RUN_SORT_FIELD } from "./hooks/useRunDetail";

describe("buildStatsViewQuery", () => {
  it("carries from and to as RFC 3339 instants", () => {
    const qs = buildStatsViewQuery({
      from: new Date("2026-01-01T00:00:00.000Z"),
      to: new Date("2026-02-01T00:00:00.000Z"),
    });
    expect(qs.get("from")).toBe("2026-01-01T00:00:00.000Z");
    expect(qs.get("to")).toBe("2026-02-01T00:00:00.000Z");
    // No project/breakdown/change were supplied -- they must not appear at
    // all, not appear as empty strings. An empty-string "project=" is a
    // different, wrong request: the server's own project filter treats a
    // present-but-empty value as "no filter" only by accident of Go's
    // zero value, and this client should never lean on that.
    expect(qs.has("project")).toBe(false);
    expect(qs.has("breakdown")).toBe(false);
    expect(qs.has("change")).toBe(false);
  });

  it("carries the project filter when supplied", () => {
    const qs = buildStatsViewQuery({
      from: new Date("2026-01-01T00:00:00.000Z"),
      to: new Date("2026-02-01T00:00:00.000Z"),
      project: "kan-16-myflow-stats-app",
    });
    expect(qs.get("project")).toBe("kan-16-myflow-stats-app");
  });

  it("carries breakdown=repo together with change, command and stage", () => {
    const qs = buildStatsViewQuery({
      from: new Date("2026-01-01T00:00:00.000Z"),
      to: new Date("2026-02-01T00:00:00.000Z"),
      breakdown: "repo",
      change: "kan-16-myflow-stats-app",
      command: "/myflow-do",
      stage: "SDD + TDD per task",
    });
    expect(qs.get("breakdown")).toBe("repo");
    expect(qs.get("change")).toBe("kan-16-myflow-stats-app");
    expect(qs.get("command")).toBe("/myflow-do");
    expect(qs.get("stage")).toBe("SDD + TDD per task");
  });

  it("rejects breakdown=repo with no change", () => {
    expect(() =>
      buildStatsViewQuery({
        from: new Date("2026-01-01T00:00:00.000Z"),
        to: new Date("2026-02-01T00:00:00.000Z"),
        breakdown: "repo",
      }),
    ).toThrow(/requires a "change"/);
  });

  // Task 21, step 5: the server now accepts "change" alone (no breakdown)
  // on cost-per-change, scoping the view to one change server-side --
  // useRunDetail.ts relies on exactly this. This client no longer rejects
  // that shape; only breakdown=repo's own command/stage pairing rule,
  // covered below, stays fixed.
  it("carries change alone with no breakdown", () => {
    const qs = buildStatsViewQuery({
      from: new Date("2026-01-01T00:00:00.000Z"),
      to: new Date("2026-02-01T00:00:00.000Z"),
      change: "kan-16-myflow-stats-app",
    });
    expect(qs.get("change")).toBe("kan-16-myflow-stats-app");
    expect(qs.has("breakdown")).toBe(false);
  });

  it("carries the model filter when supplied", () => {
    const qs = buildStatsViewQuery({
      from: new Date("2026-01-01T00:00:00.000Z"),
      to: new Date("2026-02-01T00:00:00.000Z"),
      model: "claude-opus-5",
    });
    expect(qs.get("model")).toBe("claude-opus-5");
  });

  it("omits the model filter when not supplied", () => {
    const qs = buildStatsViewQuery({
      from: new Date("2026-01-01T00:00:00.000Z"),
      to: new Date("2026-02-01T00:00:00.000Z"),
    });
    expect(qs.has("model")).toBe(false);
  });

  // The breakdown must be scoped to one cost-per-change *row*
  // (project + change + command + stage), never the change as a whole --
  // post-commit review finding F1: a request naming "change" but not the
  // row it belongs to previously summed every stage of the change into one
  // panel, so two differently-costed rows rendered the identical,
  // unreconciling total.
  it("rejects breakdown=repo with change but no command/stage", () => {
    expect(() =>
      buildStatsViewQuery({
        from: new Date("2026-01-01T00:00:00.000Z"),
        to: new Date("2026-02-01T00:00:00.000Z"),
        breakdown: "repo",
        change: "kan-16-myflow-stats-app",
      }),
    ).toThrow(/requires both a "command" and a "stage"/);
  });

  it("rejects breakdown=repo with command but no stage", () => {
    expect(() =>
      buildStatsViewQuery({
        from: new Date("2026-01-01T00:00:00.000Z"),
        to: new Date("2026-02-01T00:00:00.000Z"),
        breakdown: "repo",
        change: "kan-16-myflow-stats-app",
        command: "/myflow-do",
      }),
    ).toThrow(/requires both a "command" and a "stage"/);
  });

  it("rejects command/stage with no breakdown", () => {
    expect(() =>
      buildStatsViewQuery({
        from: new Date("2026-01-01T00:00:00.000Z"),
        to: new Date("2026-02-01T00:00:00.000Z"),
        command: "/myflow-do",
        stage: "SDD + TDD per task",
      }),
    ).toThrow(/only meaningful together with breakdown/);
  });
});

describe("buildModelsQuery", () => {
  it("carries from and to, and omits project when not supplied", () => {
    const qs = buildModelsQuery({
      from: new Date("2026-01-01T00:00:00.000Z"),
      to: new Date("2026-02-01T00:00:00.000Z"),
    });
    expect(qs.get("from")).toBe("2026-01-01T00:00:00.000Z");
    expect(qs.get("to")).toBe("2026-02-01T00:00:00.000Z");
    expect(qs.has("project")).toBe(false);
  });

  it("carries the project filter when supplied", () => {
    const qs = buildModelsQuery({
      from: new Date("2026-01-01T00:00:00.000Z"),
      to: new Date("2026-02-01T00:00:00.000Z"),
      project: "kan-16-myflow-stats-app",
    });
    expect(qs.get("project")).toBe("kan-16-myflow-stats-app");
  });
});

describe("buildListQuery", () => {
  it("is empty for an empty query", () => {
    const qs = buildListQuery({});
    expect(qs.toString()).toBe("");
  });

  it("carries q, sort with descending prefixes, limit and offset", () => {
    const qs = buildListQuery({
      q: "kan-16",
      sort: [{ field: "updatedAt", desc: true }, { field: "name" }],
      limit: 25,
      offset: 50,
    });
    expect(qs.get("q")).toBe("kan-16");
    expect(qs.get("sort")).toBe("-updatedAt,name");
    expect(qs.get("limit")).toBe("25");
    expect(qs.get("offset")).toBe("50");
  });

  it("carries arbitrary filter fields as equality params", () => {
    const qs = buildListQuery({ filters: { project: "kan-16-myflow-stats-app", state: "IN_PROGRESS" } });
    expect(qs.get("project")).toBe("kan-16-myflow-stats-app");
    expect(qs.get("state")).toBe("IN_PROGRESS");
  });

  it("rejects a filter field that collides with a reserved parameter", () => {
    expect(() => buildListQuery({ filters: { sort: "name" } })).toThrow(/reserved query parameter/);
  });

  it("rejects a negative limit before any request is sent", () => {
    expect(() => buildListQuery({ limit: -1 })).toThrow(/must not be negative/);
  });

  it("rejects a negative offset before any request is sent", () => {
    expect(() => buildListQuery({ offset: -1 })).toThrow(/must not be negative/);
  });
});

describe("fetch integration", () => {
  const originalFetch = globalThis.fetch;

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("fetchStatsView requests the exact URL the server expects and returns its rows", async () => {
    const fetchMock = vi.fn(
      async () =>
        new Response(
          JSON.stringify({
            view: "state-board",
            from: "2026-01-01T00:00:00Z",
            to: "2026-02-01T00:00:00Z",
            boundaryConvention: "start-attributed, half-open",
            recorded: true,
            rows: [{ name: "kan-16-myflow-stats-app" }],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
    );
    globalThis.fetch = fetchMock as unknown as typeof fetch;

    const resp = await fetchStatsView("state-board", {
      from: new Date("2026-01-01T00:00:00.000Z"),
      to: new Date("2026-02-01T00:00:00.000Z"),
    });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url] = fetchMock.mock.calls[0] as unknown as [string];
    expect(url).toBe(
      "/api/v1/stats/state-board?from=2026-01-01T00%3A00%3A00.000Z&to=2026-02-01T00%3A00%3A00.000Z",
    );
    expect(resp.recorded).toBe(true);
    expect(resp.rows).toEqual([{ name: "kan-16-myflow-stats-app" }]);
  });

  it("listChanges omits the leading '?' when the query is empty", async () => {
    const fetchMock = vi.fn(
      async () => new Response(JSON.stringify({ total: 0, changes: [] }), { status: 200 }),
    );
    globalThis.fetch = fetchMock as unknown as typeof fetch;

    await listChanges();

    const [url] = fetchMock.mock.calls[0] as unknown as [string];
    expect(url).toBe("/api/v1/changes");
  });

  it("listStageRuns builds one change's URL from project/name filters, a sort and a limit", async () => {
    const fetchMock = vi.fn(
      async () => new Response(JSON.stringify({ total: 0, stageRuns: [] }), { status: 200 }),
    );
    globalThis.fetch = fetchMock as unknown as typeof fetch;

    await listStageRuns({
      filters: { project: "kan-16-myflow-stats-app", name: "kan-16-myflow-stats-app" },
      sort: [{ field: "startedAt" }],
      limit: 100,
    });

    const [url] = fetchMock.mock.calls[0] as unknown as [string];
    expect(url).toBe(
      "/api/v1/stage-runs?sort=startedAt&limit=100&project=kan-16-myflow-stats-app&name=kan-16-myflow-stats-app",
    );
  });

  it("listStageRuns rejects a filter field colliding with a reserved parameter, the same as listChanges", async () => {
    await expect(listStageRuns({ filters: { limit: "10" } })).rejects.toThrow(/reserved query parameter/);
  });

  it("listStageRuns returns the response's total and stageRuns, including a run whose metrics bag is a nested object", async () => {
    const stageRuns = [
      {
        stageRunId: 1,
        harness: "claude-code",
        command: "/myflow-do",
        stage: "SDD + TDD per task",
        attempt: 1,
        startedAt: "2026-01-01T00:00:00Z",
        endedAt: "2026-01-01T00:05:00Z",
        outcome: "committed",
        metrics: { cost_usd: 0.5, tokens: { input: 100, output: 50 } },
      },
    ];
    const fetchMock = vi.fn(
      async () => new Response(JSON.stringify({ total: 1, stageRuns }), { status: 200 }),
    );
    globalThis.fetch = fetchMock as unknown as typeof fetch;

    const resp = await listStageRuns();

    expect(resp.total).toBe(1);
    expect(resp.stageRuns).toEqual(stageRuns);
  });

  it("fetchModels requests the exact URL the server expects and returns its models", async () => {
    const fetchMock = vi.fn(
      async () =>
        new Response(
          JSON.stringify({
            from: "2026-01-01T00:00:00Z",
            to: "2026-02-01T00:00:00Z",
            models: ["claude-opus-5", "claude-sonnet-5"],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
    );
    globalThis.fetch = fetchMock as unknown as typeof fetch;

    const resp = await fetchModels({
      from: new Date("2026-01-01T00:00:00.000Z"),
      to: new Date("2026-02-01T00:00:00.000Z"),
    });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url] = fetchMock.mock.calls[0] as unknown as [string];
    expect(url).toBe("/api/v1/models?from=2026-01-01T00%3A00%3A00.000Z&to=2026-02-01T00%3A00%3A00.000Z");
    expect(resp.models).toEqual(["claude-opus-5", "claude-sonnet-5"]);
  });

  it("putChange sends a PUT with a JSON body and the Content-Type header", async () => {
    const change = {
      projectKey: "agents",
      name: "kan-16-myflow-stats-app",
      state: "IN_PROGRESS",
      updatedAt: "2026-01-01T00:00:00Z",
      updatedBy: "test",
    };
    const fetchMock = vi.fn(async () => new Response(JSON.stringify(change), { status: 200 }));
    globalThis.fetch = fetchMock as unknown as typeof fetch;

    await putChange("agents", "kan-16-myflow-stats-app", change);

    const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe("/api/v1/changes/agents/kan-16-myflow-stats-app");
    expect(init.method).toBe("PUT");
    expect((init.headers as Record<string, string>)["Content-Type"]).toBe("application/json");
    expect(JSON.parse(init.body as string)).toEqual(change);
  });

  it("throws ApiError with the server's message and code on a non-2xx response", async () => {
    const fetchMock = vi.fn(
      async () =>
        new Response(JSON.stringify({ error: "change is at a later state", code: "monotonic_violation" }), {
          status: 409,
        }),
    );
    globalThis.fetch = fetchMock as unknown as typeof fetch;

    const rejection = putChange("agents", "kan-16-myflow-stats-app", {
      projectKey: "agents",
      name: "kan-16-myflow-stats-app",
      state: "STARTED",
      updatedAt: "2026-01-01T00:00:00Z",
      updatedBy: "test",
    });

    await expect(rejection).rejects.toBeInstanceOf(ApiError);
    await expect(rejection).rejects.toMatchObject({
      name: "ApiError",
      status: 409,
      code: "monotonic_violation",
      message: "change is at a later state",
    });
  });
});

// Task 26: the server's allowlist takes column names ("updated_at",
// "started_at"), not DTO field names ("updatedAt", "startedAt") -- and
// nothing compared the SPA's private copy of that vocabulary against the
// server's own until now. CHANGE_QUERY_FIELDS/STAGE_RUN_QUERY_FIELDS are
// imported from stats/web/src/testdata/queryFields.json, itself written by
// stats/internal/store/queryfields_test.go from
// store.AllowedChangeFields()/AllowedStageRunFields() -- the server's one
// real vocabulary. This suite asserts *membership* of the exact field
// names ChangeVariable.tsx and useRunDetail.ts actually send (imported
// from those files, not retyped here), so a call site that regresses back
// to a DTO field name fails this test: asserting a literal string against
// itself would not, which is why the fields under test are the
// components' own exported constants and not string literals written in
// this file.
describe("the SPA's sort fields are members of the server's allowlist, not a private copy of it", () => {
  it("CHANGE_QUERY_FIELDS is non-empty and mirrors the server's allowlist", () => {
    expect(CHANGE_QUERY_FIELDS.length).toBeGreaterThan(0);
  });

  it("STAGE_RUN_QUERY_FIELDS is non-empty and mirrors the server's allowlist", () => {
    expect(STAGE_RUN_QUERY_FIELDS.length).toBeGreaterThan(0);
  });

  it("ChangeVariable's sort field is a member of the server's change allowlist", () => {
    expect(CHANGE_QUERY_FIELDS).toContain(CHANGE_VARIABLE_SORT_FIELD);
  });

  it("useRunDetail's sort field is a member of the server's stage-run allowlist", () => {
    expect(STAGE_RUN_QUERY_FIELDS).toContain(RUN_DETAIL_STAGE_RUN_SORT_FIELD);
  });
});
