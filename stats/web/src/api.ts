// api.ts is the typed client for flowd's HTTP surface (stats/internal/api).
// It carries exactly the parameters the server actually accepts --
// "from"/"to"/"project"/"breakdown"/"change"/"command"/"stage"/"model" on
// the statistics views, and "q"/"sort"/"limit"/"offset" plus arbitrary
// allowlisted filter fields on the two list endpoints -- and nothing
// more. It deliberately does not filter, sort, search or paginate
// anything client-side: the server already does that work
// (internal/store's query allowlist), and doing it twice is exactly what
// would make a paginated list wrong, per this task's own non-negotiable
// requirement.
//
// queryFields.json (imported below) is the server's own field vocabulary
// for SortKey.field and ListQuery.filters, not a hand-typed guess at it:
// stats/internal/store/queryfields_test.go writes
// store.AllowedChangeFields()/AllowedStageRunFields() to that file and
// fails when it disagrees with the committed copy (the same
// generated-fixture-plus-drift-test discipline
// internal/harvest/wireshape_test.go already uses for the metrics wire
// shape). Task 26's defect -- ChangeVariable.tsx and useRunDetail.ts
// sorting by "updatedAt"/"startedAt", DTO field names the server's
// allowlist has never accepted (it takes "updated_at"/"started_at") --
// was exactly a private copy of this vocabulary going stale; importing it
// here instead closes that gap for every future call site, not just
// these two.
import queryFields from "./testdata/queryFields.json";

/** Every field name GET /api/v1/changes accepts in a Filter or a SortKey,
 * mirrored from store.AllowedChangeFields() via queryFields.json -- see
 * this file's header comment. */
export const CHANGE_QUERY_FIELDS: readonly string[] = queryFields.changeFields;

/** Every field name GET /api/v1/stage-runs accepts in a Filter or a
 * SortKey, mirrored from store.AllowedStageRunFields() via
 * queryFields.json -- see this file's header comment. */
export const STAGE_RUN_QUERY_FIELDS: readonly string[] = queryFields.stageRunFields;

/** One of the four surviving statistics views' URL slugs (design.md's
 * `stats-ui-cut`: `CostPerChange`, `ModelComparison`, `PanelEconomics` and
 * `ReworkRate` were cut). */
export type ViewName =
  | "state-board"
  | "stage-leaderboard"
  | "trend"
  | "weekly-costs"
  | "reviewers"
  | "decisions"
  | "flow-health"
  | "self-review"
  | "review-dynamics"
  | "experiments";

export const VIEW_NAMES: readonly ViewName[] = [
  "state-board",
  "stage-leaderboard",
  "trend",
  "weekly-costs",
  "reviewers",
  "decisions",
  "flow-health",
  "self-review",
  "review-dynamics",
  "experiments",
];

/** The "cost-per-change" statistics endpoint slug -- not a navigable
 * `ViewName` (its standalone dashboard was cut, design.md's `stats-ui-cut`),
 * but still called directly by `useRunDetail.ts` for a change's own
 * aggregate, so `fetchStatsView` keeps accepting it alongside `ViewName`. */
type CostPerChangeViewSlug = "cost-per-change";

/** The statistics endpoint the "flow-health" dashboard reads
 * (internal/api/health.go) -- not a navigable view of its own. */
export type HealthViewSlug = "flow-summary";

/** The statistics endpoint the Weekly costs view's model·effort chart reads
 * (decision model-effort-view) -- not a navigable view of its own. */
export type WeeklyModelEffortViewSlug = "weekly-model-effort-costs";

/** Every slug GET /api/v1/stats/{view} is called with from this client. */
/** Experiments reads its own endpoints (fetchRoles, fetchExperiments,
 * fetchExperimentStats), never /api/v1/stats/{view}. */
export type StatsViewSlug = Exclude<ViewName, "experiments"> | CostPerChangeViewSlug | HealthViewSlug | WeeklyModelEffortViewSlug;

/**
 * Parameters every statistics view accepts. "from" and "to" are both
 * required by the server (internal/api/stats.go's parsePeriodAndProject) --
 * a view with no period is not a smaller request, it is a different,
 * unbounded one the server does not offer -- so they are required here
 * too, rather than left optional and rejected only at the network hop.
 */
export interface StatsViewParams {
  from: Date;
  to: Date;
  /** Restrict to one project; omitted aggregates across every project. */
  project?: string;
  /** Restrict review dynamics to changes of one plan class. */
  class?: "micro" | "small" | "regular" | "big";
  /**
   * Per-repository breakdown. The server honours this only on
   * "cost-per-change", and only together with "change", "command" and
   * "stage" -- see design.md's API section ("The per-repository breakdown
   * is available on cost-per-change alone"). Requesting it elsewhere is
   * rejected by the server with 400; buildStatsViewQuery rejects the same
   * shape client-side, before a request is even sent, since the rule is
   * fixed and known here.
   */
  breakdown?: "repo";
  /**
   * Scopes the view to one change. Required together with breakdown --
   * see breakdown's own doc comment -- but also meaningful on its own on
   * "cost-per-change": the server accepts "change" alone there, filtering
   * to that one change's own rows server-side (internal/api/stats.go's
   * rowsFor, task 21 step 5), which is what useRunDetail.ts relies on
   * instead of fetching every change in the project and filtering
   * client-side. Sending "change" alone to a view other than
   * cost-per-change is accepted by the server and simply has no effect,
   * the same way "breakdown" itself does on every view but cost-per-change.
   */
  change?: string;
  /** Restrict to one model's own stage runs; rejected by the server on
   * "state-board", whose rows are changes rather than stage runs
   * (the stats-views requirement "A model restriction on the live
   * state board"). */
  model?: string;
  /**
   * The command and stage naming the exact cost-per-change *row* to break
   * down -- required together with breakdown and change, both of them,
   * never one alone. cost-per-change's own rows are grouped by
   * (project, change, command, stage) -- one row per stage
   * (internal/api/stats.go's costPerChangeByRepo doc comment) -- and a
   * request naming the change but not the row would sum every stage's
   * runs together, which is exactly what let two differently-costed rows
   * of the same change render the identical, unreconciling breakdown
   * (post-commit review finding F1, reproduced live against a seeded
   * database). Requiring both here, before a request is even sent, keeps
   * that mistake from being reachable from this client at all.
   */
  command?: string;
  stage?: string;
}

/** The envelope every statistics view answers with. */
export interface StatsResponse<Row = unknown> {
  view: StatsViewSlug;
  /** RFC 3339, the period actually applied -- echoed back, not merely the request's raw string. */
  from: string;
  to: string;
  project?: string;
  /** Echoed back only when a model restriction was applied to this request. */
  model?: string;
  boundaryConvention: string;
  /**
   * False when the requested period lies entirely before this store's
   * earliest recorded stage run -- "not recorded", distinct from "recorded
   * as zero" (design.md, "Starting empty"; the "absence of history"
   * requirement in the stats-views requirement).
   */
  recorded: boolean;
  /**
   * The third arm of the absence distinction (design.md, "the third arm
   * of the absence distinction"): true when stage runs exist for this
   * period and scope, but not one of them carries a measurement --
   * distinct from `recorded: false` ("no runs exist for this period at
   * all") and distinct from a view whose rows carry a real, measured
   * zero. Only ever true when `recorded` is also true -- a period that
   * predates any telemetry is reported through `recorded` alone. A UI
   * must render this as its own state, never folding it into either
   * neighbour: that collapse is exactly the defect task 5 exists to
   * close (a recording misconfiguration read as a quiet week).
   */
  unmeasured: boolean;
  /**
   * How many stage runs in scope recorded no model at all, present only
   * when a model restriction was applied -- absent means "no filter",
   * never "zero were excluded" (task 21, step 3), so a UI reading this
   * must check for the key's presence, not just truthiness (0 is a real,
   * meaningful value here).
   */
  excludedNoModel?: number;
  rows: Row;
}

/** Raised for every non-2xx response. `code` is set only for the handful of machine-distinguishable errors the server names (e.g. an undocumented stage). */
export class ApiError extends Error {
  readonly status: number;
  readonly code?: string;

  constructor(status: number, message: string, code?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

/** Query parameters the two list endpoints (changes, stage-runs) share, mirrored from internal/api's parseChangeQuery / parseStageRunQuery. */
export interface SortKey {
  field: string;
  desc?: boolean;
}

export interface ListQuery {
  /** Free-text term, matched across each endpoint's identity fields. */
  q?: string;
  /** Sort keys applied in order; each becomes "-field" server-side when desc. */
  sort?: SortKey[];
  limit?: number;
  offset?: number;
  /**
   * Equality filters, field name to value -- resolved (or rejected) by
   * the server's own allowlist (internal/store/query.go). This client
   * invents no filtering of its own: every field here is sent as-is and
   * the server decides whether it is known.
   */
  filters?: Record<string, string>;
}

/** Reserved list-query parameter names a filter field must not collide with -- the same set internal/api's reservedQueryParams guards server-side. */
const RESERVED_LIST_PARAMS: ReadonlySet<string> = new Set(["q", "sort", "limit", "offset"]);

export interface RepoDTO {
  repoRoot: string;
  mergeBase?: string;
}

/** Mirrors internal/api/changes.go's changeDTO -- the wire shape GET and PUT exchange. */
export interface ChangeDTO {
  projectKey: string;
  name: string;
  state: string;
  branch?: string;
  worktrees?: unknown;
  artifactUrl?: string;
  jiraIssue?: string;
  planningEffort?: string;
  models?: unknown;
  reviewPanelRoster?: string;
  prUrl?: string;
  repos?: RepoDTO[];
  mainCheckoutPath?: string;
  updatedAt: string;
  updatedBy: string;
}

export interface ListChangesResponse {
  total: number;
  changes: ChangeDTO[];
}

// --- statistics view row DTOs ---
//
// Every type below mirrors its Go counterpart in internal/api/stats.go
// field-for-field, including nullability: a field typed `| null` there
// carries an explicit JSON `null` when unmeasured (no `omitempty` on the Go
// side, per that file's own header comment), and a UI reading it must
// render that null as "unavailable" rather than coercing it to zero -- the
// same absence-is-not-a-value rule the metrics bag draws, carried through
// to the wire and now to the view layer.

/** Mirrors stateBoardRowDTO. */
export interface PlanTask {
  n: number;
  title: string;
  state: "done" | "in progress" | "pending";
  /** A /flow-interactive request's outcome, short landed commit and brainstorm reason (KAN-1061); each absent when unset. */
  outcome?: "landed" | "answered" | "dropped";
  commit?: string;
  reason?: string;
}

/** Mirrors store.ActiveDispatch: a dispatch begun and not yet ended. */
export interface ActiveSubagent {
  role: string;
  taskId?: string;
  key?: string;
  model: string;
  modelVersion?: string;
  effort?: string;
  startedAt: string;
  /** Run by the main session itself, not a subagent. */
  inline?: boolean;
}

export interface StateBoardRow {
  projectKey: string;
  name: string;
  state: string;
  updatedAt: string;
  updatedBy: string;
  nextCommand: string;
  /** The change's first and latest task-count observations (KAN-415); absent
   * when the change was never observed -- absence, never a zero. */
  plannedTasks?: number;
  currentTasks?: number;
  /** Whole-life run totals (store.LiveStateRow). */
  runCount: number;
  costUsd: number;
  /** Cost of the stage runs started within the last 24h, 7d and 30d. */
  cost24h: number;
  cost7d: number;
  cost30d: number;
  wallClockMs: number;
  errors: number;
  /** The open stage run, the done-task count and the open run's session
   * activity; each absent when the store has none. */
  currentStage?: string;
  currentStageStartedAt?: string;
  /** The latest stage run, open or not; lastStageEndedAt is absent while it is open. */
  lastStage?: string;
  lastStageEndedAt?: string;
  doneTasks?: number;
  /** An unfinished change's plan tasks, and its running subagents; each absent when empty. */
  tasks?: PlanTask[];
  /** The self-review planner's own plan (self-review-<name>), absent when it has none. */
  selfReviewTasks?: PlanTask[];
  subagents?: ActiveSubagent[];
  /** When the session last wrote to its transcript; absent for a finished change. */
  sessionSeenAt?: string;
  /** Dispatches begun since the open stage started, ended or not. */
  openStageDispatches?: number;
  activity?: 'working' | 'question' | 'permission' | 'idle';
  activityDetail?: string;
  activitySince?: string;
  /** The command of the change's latest stage run; tasks are its requests when it is /flow-interactive. */
  command?: string;
  /** Main's highest review-panel round and reviewer dispatches, then the self-review's (store.LiveStateRow). */
  fixRounds: number;
  reviewerRuns: number;
  selfReviewFixRounds: number;
  selfReviewReviewerRuns: number;
}

/** Mirrors costPerChangeRowDTO. */
export interface CostPerChangeRow {
  projectKey: string;
  changeName: string;
  command: string;
  stage: string;
  runCount: number;
  measuredRuns: number;
  totalTokensInput: number | null;
  meanTokensInput: number | null;
  totalCostUsd: number | null;
  totalDurationMs: number | null;
  mainTokens: number | null;
  sidechainTokens: number | null;
}

/** Mirrors stageLeaderboardRowDTO. */
export interface StageLeaderboardRow {
  command: string;
  stage: string;
  runCount: number;
  meanCostUsd: number;
  medianCostUsd: number;
  p90CostUsd: number;
  totalCostUsd: number;
}

/** Mirrors flowSummaryDTO (internal/api/health.go): the flow-summary
 * view's one object of period totals. The cost and mean figures are null
 * when nothing in the period was measured -- absence, never a zero. */
export interface FlowSummary {
  changesStarted: number;
  changesFinished: number;
  totalCostUsd: number | null;
  meanCostPerFinished: number | null;
  medianLeadTimeMs: number | null;
  meanPanelRounds: number | null;
  critical: number;
  important: number;
  minor: number;
  guardFires: number;
}

/** Mirrors selfReviewRowDTO. `commitUrl` is "" unless the finding is
 * `fixed` and the daemon knows its checkout's GitHub commit base. */
export interface SelfReviewRow {
  recordedAt: string;
  project: string;
  change: string;
  angle: string;
  note: string;
  disposition: "fixed" | "filed" | "declined" | "noted";
  ref: string;
  blastRadius: number | null;
  commitUrl: string;
}

/** Mirrors trendPointDTO. */
export interface TrendPoint {
  day: string;
  runCount: number;
  totalCostUsd: number | null;
}

/** Mirrors weeklyCostPointDTO: one Wednesday-to-Wednesday week, null
 * where the week carries no costed run in that category. */
export interface WeeklyCostPoint {
  week: string;
  overallUsd: number | null;
  planningUsd: number | null;
  implementationUsd: number | null;
  reviewUsd: number | null;
  selfReviewUsd: number | null;
  visualVerifyUsd: number | null;
  otherUsd: number | null;
}

/** Mirrors weeklyModelEffortPointDTO: one week's cost for one model and
 * effort -- a dispatch effort, "default", or "session". */
export interface WeeklyModelEffortPoint {
  week: string;
  model: string;
  effort: string;
  costUsd: number;
}

/** One value of one Review dynamics series in one bucket (YYYY-MM-DD). */
export interface DynamicsPoint {
  bucket: string;
  metric: string;
  series: string;
  value: number;
  /** The changes or dispatches the value is normalised by. */
  basis: number;
}

/** One slot × model · effort heatmap cell's dispatches and findings. */
export interface HeatCell {
  slot: string;
  model: string;
  effort: string;
  dispatches: number;
  critical: number;
  important: number;
  minor: number;
}

/** The verdict tiles' values over a whole range. */
export interface DynamicsSummary {
  findingsPerChange: number;
  criticalImportantPerChange: number;
  meanFixRounds: number;
  implFindingsPerTask: number | null;
  vvFirstPassShare: number | null;
  changes: number;
}

/** One change to a tunable prompt, drawn as a chart marker. */
export interface PromptChange {
  sha: string;
  at: string;
  kind: "tuning" | "edit";
  area: "review" | "impl" | "vv" | "fix";
  slot: string | null;
  target: string;
  summary: string;
  caseId: string | null;
}

/** The "review-dynamics" view's rows: one object, not an array (KAN-1046). */
export interface ReviewDynamicsRows {
  bucket: "day" | "week";
  series: DynamicsPoint[];
  heat: HeatCell[];
  previousHeat: HeatCell[];
  heatVersion: HeatCell[];
  previousHeatVersion: HeatCell[];
  summary: DynamicsSummary;
  previousSummary: DynamicsSummary;
  markers: PromptChange[];
  lastPromptChangeAt: string | null;
  projects: string[];
}

/** Mirrors reviewerRowDTO. */
export interface ReviewerRow {
  slot: string;
  experimental: boolean;
  description: string;
  dispatches: number;
  changes: number;
  critical: number;
  important: number;
  minor: number;
}

/** Mirrors decisionRowDTO. */
export interface DecisionRow {
  project: string;
  change: string;
  recordedAt: string;
  class: string;
  overridden: boolean;
  execution: string;
  implementerModel: string;
  implementerEffort: string;
  fixer: string;
  rerunDispatch: string;
  rosterSize: number;
  compact: boolean;
  experimentalSlot: string;
  rerun: string;
  grouping: string;
  dispatches: string;
  implementerGroups: string;
  wallClockSeconds: number;
  inputTokens: number;
  outputTokens: number;
  cacheReadTokens: number;
  costUsd: number;
  critical: number;
  important: number;
  minor: number;
  fixRounds: number;
  fallbacks: number;
  timedOut: number;
}

/**
 * Builds the query string for GET /api/v1/stats/{view} from params, exactly
 * the parameters internal/api/stats.go's statsQueryParams allowlist accepts
 * ("from", "to", "project", "breakdown", "change", "command", "stage") and
 * no others.
 *
 * Exported (not just used internally) so a caller -- and this file's own
 * tests -- can assert the constructed query directly, without needing a
 * network layer in the loop.
 */
export function buildStatsViewQuery(params: StatsViewParams): URLSearchParams {
  if (params.breakdown === "repo" && !params.change) {
    throw new Error('breakdown "repo" requires a "change" parameter naming the change to break down');
  }
  // "change" alone (no breakdown) is meaningful too -- task 21, step 5:
  // the server accepts it on cost-per-change to scope the view to one
  // change, server-side, which is what useRunDetail.ts relies on. Only
  // the breakdown=repo pairing rule below stays fixed: command and stage
  // still require breakdown "repo" together with change, never change
  // alone.
  // command and stage name the exact row breakdown=repo scopes to
  // (internal/api/stats.go requires both alongside change -- see
  // StatsViewParams's own doc comment for why "change" alone is not
  // enough). Both or neither: a request naming one but not the other is
  // exactly the half-scoped shape the server now rejects with 400, so this
  // client refuses it before ever sending it.
  if (params.breakdown === "repo" && (!params.command || !params.stage)) {
    throw new Error('breakdown "repo" requires both a "command" and a "stage" parameter naming the row to break down');
  }
  if ((params.command || params.stage) && params.breakdown !== "repo") {
    throw new Error('"command"/"stage" are only meaningful together with breakdown "repo"');
  }

  const qs = new URLSearchParams();
  qs.set("from", params.from.toISOString());
  qs.set("to", params.to.toISOString());
  if (params.project) {
    qs.set("project", params.project);
  }
  if (params.breakdown) {
    qs.set("breakdown", params.breakdown);
  }
  if (params.change) {
    qs.set("change", params.change);
  }
  if (params.command) {
    qs.set("command", params.command);
  }
  if (params.stage) {
    qs.set("stage", params.stage);
  }
  if (params.model) {
    qs.set("model", params.model);
  }
  if (params.class) {
    qs.set("class", params.class);
  }
  return qs;
}

/**
 * Builds the query string the two list endpoints share, from an ListQuery.
 * A filter field colliding with a reserved parameter name is rejected here
 * -- before a request is sent -- rather than silently overwriting "q",
 * "sort", "limit" or "offset" in the resulting query string.
 */
export function buildListQuery(query: ListQuery): URLSearchParams {
  const qs = new URLSearchParams();
  if (query.q) {
    qs.set("q", query.q);
  }
  if (query.sort && query.sort.length > 0) {
    qs.set("sort", query.sort.map((k) => (k.desc ? `-${k.field}` : k.field)).join(","));
  }
  if (query.limit !== undefined) {
    if (query.limit < 0) {
      throw new Error(`limit must not be negative: ${query.limit}`);
    }
    qs.set("limit", String(query.limit));
  }
  if (query.offset !== undefined) {
    if (query.offset < 0) {
      throw new Error(`offset must not be negative: ${query.offset}`);
    }
    qs.set("offset", String(query.offset));
  }
  if (query.filters) {
    for (const [field, value] of Object.entries(query.filters)) {
      if (RESERVED_LIST_PARAMS.has(field)) {
        throw new Error(`"${field}" is a reserved query parameter and cannot also be used as a filter field`);
      }
      qs.set(field, value);
    }
  }
  return qs;
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let message = res.statusText;
    let code: string | undefined;
    try {
      const body = (await res.json()) as { error?: string; code?: string };
      if (body.error) {
        message = body.error;
      }
      code = body.code;
    } catch {
      // Body wasn't JSON (or was empty) -- fall back to statusText.
    }
    throw new ApiError(res.status, message, code);
  }
  return (await res.json()) as T;
}

async function getJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, { ...init, method: "GET" });
  return handleResponse<T>(res);
}

/** GET /api/v1/stats/{view}?from=&to=&project=&breakdown=&change= */
export async function fetchStatsView<Row = unknown>(
  view: StatsViewSlug,
  params: StatsViewParams,
  init?: RequestInit,
): Promise<StatsResponse<Row>> {
  const qs = buildStatsViewQuery(params);
  return getJSON<StatsResponse<Row>>(`/api/v1/stats/${view}?${qs.toString()}`, init);
}

/** GET /api/v1/changes?q=&sort=&limit=&offset=&<filters> */
export async function listChanges(query: ListQuery = {}, init?: RequestInit): Promise<ListChangesResponse> {
  const qs = buildListQuery(query);
  const suffix = qs.toString();
  return getJSON<ListChangesResponse>(`/api/v1/changes${suffix ? `?${suffix}` : ""}`, init);
}

/**
 * Mirrors internal/api/stats.go's stageRunDTO field for field. `metrics`
 * stays `unknown`, deliberately: it is the open JSONB bag design.md's
 * metrics-bag decision keeps flexible, and closing it into a typed
 * interface here would re-close in the client what the schema went out of
 * its way to leave open. metrics.ts reads it defensively, key by key.
 */
export interface StageRunDTO {
  stageRunId: number;
  repoRoot?: string;
  harness: string;
  sessionId?: string;
  command: string;
  stage: string;
  attempt: number;
  startedAt: string;
  endedAt?: string;
  outcome?: string;
  metrics?: unknown;
  /** True when the store opened the run for a dispatch begun with no stage open. */
  autoOpened?: boolean;
  /** The /flow-interactive request the run worked on (KAN-1061). */
  request?: number;
}

/** Mirrors internal/api/stats.go's listStageRunsResponse. */
export interface ListStageRunsResponse {
  total: number;
  stageRuns: StageRunDTO[];
}

/**
 * One dispatch's contribution to a stage run, derived client-side
 * (RunDetail.tsx is the sole reader) from its metrics bag's
 * `dispatches.<agentId>` key -- internal/harvest.DispatchBucket's own
 * encoding (attribute.go, KAN-201). Every field but `agentId` is optional:
 * `agentType`/`description`/`model` are omitted, never `""`, when that
 * dispatch's meta sidecar was never found, and `tokens`/`costUsd` are
 * omitted, never `0`, when that dispatch (or its recorded model) carries
 * no priceable token figures -- the same absence-is-never-a-value rule
 * every other metrics reader in this SPA already holds
 * (the stats-views requirement "A stage run opens onto its own
 * dispatches").
 *
 * `costUsd` is read directly from that same `dispatches.<agentId>` key's
 * own `cost_usd` -- store.Store.Price (internal/store/pricing.go) prices
 * each dispatch bucket through the identical pricing path it already
 * uses for `models.<model>`, applied to the dispatch's own recorded
 * model and own tokens, never a second, derived calculation on this
 * client.
 */
export interface DispatchRow {
  agentId: string;
  agentType?: string;
  description?: string;
  model?: string;
  tokens?: number;
  costUsd?: number;
}

/**
 * GET /api/v1/stage-runs?q=&sort=&limit=&offset=&<filters> -- the
 * stage-run counterpart of listChanges, built on the same buildListQuery
 * this client already uses for the changes list. "project" and "name" are
 * filterable here via the owning change's join (internal/store/query.go's
 * stageRunFieldColumns = stageRunOwnFieldColumns merged with
 * changeFieldColumns), which is how a caller scopes this to one change.
 */
export async function listStageRuns(query: ListQuery = {}, init?: RequestInit): Promise<ListStageRunsResponse> {
  const qs = buildListQuery(query);
  const suffix = qs.toString();
  return getJSON<ListStageRunsResponse>(`/api/v1/stage-runs${suffix ? `?${suffix}` : ""}`, init);
}

/** GET /api/v1/changes/{project}/{name} */
export async function getChange(project: string, name: string, init?: RequestInit): Promise<ChangeDTO> {
  return getJSON<ChangeDTO>(`/api/v1/changes/${encodeURIComponent(project)}/${encodeURIComponent(name)}`, init);
}

/**
 * Parameters GET /api/v1/models accepts -- exactly internal/api/stats.go's
 * modelsQueryParams ("from", "to", "project"), mirroring
 * StatsViewParams's own required-from/to posture: a model list with no
 * period is a different, unbounded question the server does not offer.
 */
export interface ModelsParams {
  from: Date;
  to: Date;
  project?: string;
}

/** Mirrors internal/api/stats.go's modelsResponse. */
export interface ModelsResponse {
  from: string;
  to: string;
  project?: string;
  /** The distinct models recorded in the period -- ModelVariable.tsx's
   * only source for the dropdown it offers, never a hard-coded list
   * (the stats-views requirement "The models offered"). */
  models: string[];
}

export function buildModelsQuery(params: ModelsParams): URLSearchParams {
  const qs = new URLSearchParams();
  qs.set("from", params.from.toISOString());
  qs.set("to", params.to.toISOString());
  if (params.project) {
    qs.set("project", params.project);
  }
  return qs;
}

/** The fields the run view reads of a records.Dispatch, the run record's dispatch row. */
export interface RecordDispatch {
  id: number;
  /** The harness's subagent id, which keys the dispatch's priced bucket in its stage run's metrics. */
  agentId?: string;
  key?: string;
  taskId?: string;
  role: string;
  model: string;
  effort?: string;
  outcome?: string;
  /** Why a blocked outcome happened, and the dispatcher's free-text notes. */
  cause?: string;
  notes?: string;
  startedAt: string;
  endedAt?: string;
  /** The harvester's figures: the exact model, and an inline row's real effort (its own column records "default"). */
  metrics?: { modelVersion?: string; effort?: string };
}

/** GET /api/v1/records/{project}/{change} -- the change's run record; the run view reads its dispatches. */
export function getRunRecord(project: string, change: string, init?: RequestInit): Promise<{ dispatches: RecordDispatch[] }> {
  return getJSON<{ dispatches: RecordDispatch[] }>(`/api/v1/records/${encodeURIComponent(project)}/${encodeURIComponent(change)}`, init);
}

/** GET /api/v1/changes/find?name= -- every project's record for one change name. */
export function findChanges(name: string, init?: RequestInit): Promise<{ records: ChangeDTO[] }> {
  return getJSON<{ records: ChangeDTO[] }>(`/api/v1/changes/find?name=${encodeURIComponent(name)}`, init);
}

/** Mirrors records.SelfReviewFinding. */
export interface SelfReviewFinding {
  id: number;
  change: string;
  angle: string;
  note: string;
  disposition: SelfReviewRow["disposition"];
  ref?: string;
  blastRadius?: number;
  recordedAt: string;
}

/** GET /api/v1/self-review/{project}/{change}/findings -- the change's findings in recorded order. */
export function listSelfReviewFindings(project: string, change: string, init?: RequestInit): Promise<SelfReviewFinding[]> {
  return getJSON<SelfReviewFinding[]>(`/api/v1/self-review/${encodeURIComponent(project)}/${encodeURIComponent(change)}/findings`, init);
}

/** GET /api/v1/stages?command= -- the command's stage keys in the order a run marks them. */
export function listStages(command: string, init?: RequestInit): Promise<{ stages: string[] }> {
  return getJSON<{ stages: string[] }>(`/api/v1/stages?command=${encodeURIComponent(command)}`, init);
}

/** GET /api/v1/models?from=&to=&project= */
export async function fetchModels(params: ModelsParams, init?: RequestInit): Promise<ModelsResponse> {
  const qs = buildModelsQuery(params);
  return getJSON<ModelsResponse>(`/api/v1/models?${qs.toString()}`, init);
}

/** PUT /api/v1/changes/{project}/{name} -- sends the whole record, per the state file contract's write rule. */
export async function putChange(
  project: string,
  name: string,
  body: ChangeDTO,
  init?: RequestInit,
): Promise<ChangeDTO> {
  const res = await fetch(`/api/v1/changes/${encodeURIComponent(project)}/${encodeURIComponent(name)}`, {
    ...init,
    method: "PUT",
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
    body: JSON.stringify(body),
  });
  return handleResponse<ChangeDTO>(res);
}


/** One row of GET /api/v1/roles: the Allowed pairs table in force, a
 * leading `main` entry, and the open experiment cell replacing a role's
 * unmarked row (KAN-1067). */
export interface RoleEntry {
  role: string;
  pairs: string[];
  marker: string;
  group: string;
  does: string;
  override?: { pairs: string[]; experimentId: number };
}

/** One allowed-pairs experiment, as GET /api/v1/experiments lists it. */
export interface ExperimentDTO {
  id: number;
  name: string;
  cells: Record<string, string[]>;
  startedAt: string;
  closedAt?: string;
  verdict?: string;
  summary: string;
  status: "open" | "closed";
}

/** One metric of GET /api/v1/experiments/{id}/stats: means before and
 * after the start (null with no counted run) and its weekly series. Lower
 * is better for every metric. */
export interface ExperimentMetric {
  name: string;
  before: number | null;
  after: number | null;
  series: { week: string; value: number }[];
}

export interface ExperimentStats {
  experiment: ExperimentDTO;
  beforeFrom: string;
  end: string;
  runs: { before: number; after: number };
  excluded: number;
  metrics: ExperimentMetric[];
}

/** GET /api/v1/roles */
export function fetchRoles(init?: RequestInit): Promise<RoleEntry[]> {
  return getJSON<RoleEntry[]>("/api/v1/roles", init);
}

/** GET /api/v1/experiments -- newest first. */
export function fetchExperiments(init?: RequestInit): Promise<ExperimentDTO[]> {
  return getJSON<ExperimentDTO[]>("/api/v1/experiments", init);
}

/** GET /api/v1/experiments/{id}/stats */
export function fetchExperimentStats(id: number, init?: RequestInit): Promise<ExperimentStats> {
  return getJSON<ExperimentStats>(`/api/v1/experiments/${id}/stats`, init);
}
