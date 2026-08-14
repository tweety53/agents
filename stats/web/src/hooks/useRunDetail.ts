// useRunDetail fetches everything the run detail dashboard
// (views/RunDetail.tsx) needs for one change: its stage runs, and the
// cost-per-change aggregate scoped to that change. **The header's totals
// come from the aggregate, never from summing the fetched stage runs** --
// the round's own spec scenario ("a change with more stage runs than one
// page holds") is exactly this, and summing a page is the defect it names
// (specs/myflow-stats-views/spec.md, "One change opens on its own
// dashboard"). Summing cost-per-change's own rows across a change's
// several (command, stage) groupings is a different thing and is fine:
// that endpoint is not paginated, so its rows are already the server's
// complete answer for the period requested, whatever the change's stage
// runs list is paged to.
import { useEffect, useState } from "react";
import { fetchStatsView, listStageRuns, type CostPerChangeRow, type StageRunDTO } from "../api";

export interface RunSummary {
  runCount: number;
  measuredRuns: number;
  totalCostUsd: number | null;
  totalTokensInput: number | null;
  totalDurationMs: number | null;
  mainTokens: number | null;
  sidechainTokens: number | null;
}

export type RunDetailState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | {
      status: "ready";
      stageRuns: StageRunDTO[];
      summary: RunSummary;
      /** True when the change has more stage runs than the fetched page
       * held (totalStageRuns > stageRuns.length, before any model
       * filtering below). Task 25, step 4: a model-filtered table built
       * only from the fetched page can silently show far fewer rows than
       * the change actually has, with the header numbers (from the
       * unpaged aggregate) disagreeing with what is visible -- a count
       * that looks complete and is not. This flag is what the panel uses
       * to say so rather than leave it unstated. */
      truncated: boolean;
      /** The server's own total matching (project, change), independent
       * of any model filter -- what "showing N of totalStageRuns" reports
       * alongside `truncated`. */
      totalStageRuns: number;
    };

// One page of stage runs -- large enough for virtually every real change,
// but deliberately finite. The header numbers must never depend on
// fetching every stage run to be correct, which is exactly the property
// tested by seeding a page shorter than the aggregate's own run count.
const STAGE_RUN_PAGE_SIZE = 200;

/**
 * The field this hook sorts the stage-run list by -- a real column name
 * from the server's own allowlist (store.AllowedStageRunFields(), mirrored
 * into api.ts's STAGE_RUN_QUERY_FIELDS), never a DTO field name. Exported
 * so api.test.ts's membership test asserts against exactly the string this
 * hook actually sends: task 26's defect was this same field sorting by
 * "startedAt" (StageRunDTO's own field name), which the server's allowlist
 * has never accepted and returns 400 for -- the run detail dashboard could
 * not load its stage runs at all.
 */
export const RUN_DETAIL_STAGE_RUN_SORT_FIELD = "started_at";

/** Sums a metric across cost-per-change rows, honouring absence-is-never-zero:
 * null when none of the rows carry the metric, the sum of the present
 * values otherwise (a row with no measured runs for that stage contributes
 * nothing to the sum, the same as it would to a server-side SUM). */
function sumNullable(values: Array<number | null>): number | null {
  let total = 0;
  let any = false;
  for (const v of values) {
    if (v !== null) {
      total += v;
      any = true;
    }
  }
  return any ? total : null;
}

function summarize(rows: CostPerChangeRow[]): RunSummary {
  return {
    runCount: rows.reduce((sum, r) => sum + r.runCount, 0),
    measuredRuns: rows.reduce((sum, r) => sum + r.measuredRuns, 0),
    totalCostUsd: sumNullable(rows.map((r) => r.totalCostUsd)),
    totalTokensInput: sumNullable(rows.map((r) => r.totalTokensInput)),
    totalDurationMs: sumNullable(rows.map((r) => r.totalDurationMs)),
    mainTokens: sumNullable(rows.map((r) => r.mainTokens)),
    sidechainTokens: sumNullable(rows.map((r) => r.sidechainTokens)),
  };
}

// The run detail route shows a change's whole run history (spec: "every
// stage run recorded against it"), not the shared period control's
// window -- so this hook asks cost-per-change for a fixed, wide period of
// its own rather than accepting one from its caller. `to` is captured once
// per fetch (not memoised across renders) so a change that gains a new
// stage run and is revisited is still covered.
const WIDE_FROM = new Date(0);

/**
 * True when the stage run's metrics bag recorded the given model among the
 * models it used. Reads the metrics bag's "models" object directly --
 * `metrics.models.<model>` (task 22's per-model bucket shape) -- rather
 * than going through metrics.ts's `readModel`, which still reads the
 * retired top-level scalar "model" key and is out of this task's file
 * list to fix (that reader's own repair belongs to whichever later task
 * owns metrics.ts). A run whose bag carries no "models" object, or whose
 * "models" object does not carry this key, never matches -- the same
 * absence-is-never-a-match rule the server's own model restriction
 * applies (specs/myflow-stats-views/spec.md, "Runs that recorded no
 * model": such a run is excluded, never coerced into matching every
 * filter by default).
 */
function stageRunUsedModel(run: StageRunDTO, model: string): boolean {
  if (run.metrics === null || typeof run.metrics !== "object") return false;
  const models = (run.metrics as Record<string, unknown>).models;
  if (models === null || typeof models !== "object" || Array.isArray(models)) return false;
  return Object.prototype.hasOwnProperty.call(models, model);
}

export function useRunDetail(project: string, change: string, model?: string): RunDetailState {
  const [state, setState] = useState<RunDetailState>({ status: "loading" });

  useEffect(() => {
    let cancelled = false;
    setState({ status: "loading" });

    Promise.all([
      listStageRuns({
        filters: { project, name: change },
        sort: [{ field: RUN_DETAIL_STAGE_RUN_SORT_FIELD }],
        limit: STAGE_RUN_PAGE_SIZE,
      }),
      fetchStatsView<CostPerChangeRow[]>("cost-per-change", {
        from: WIDE_FROM,
        to: new Date(),
        project,
        // Task 21, step 5: the server now accepts "change" alone on
        // cost-per-change and scopes the aggregate to this one change,
        // server-side -- so this request no longer relies on fetching
        // every change in the project and filtering to one in the
        // browser to be correct, which stayed right only because this
        // view is unpaged (this file's own header comment on why the
        // header's totals must never come from summing a paged list).
        // Task 25, step 2 removed the client-side changeName filter that
        // used to sit here: it had become a no-op the moment this request
        // started sending "change" server-side, and a no-op filter that
        // looks load-bearing invites the next reader to keep it. The rows
        // this response carries are taken as-is, below.
        change,
        // Task 20, step 4: the run-detail dashboard honours the model
        // template variable rather than leaving it rendering and inert.
        // cost-per-change already accepts "model" (task 21) and restricts
        // server-side against the per-model buckets, so the header's
        // totals become that model's own totals for this change, the same
        // way they already are the server's aggregate rather than a sum
        // of the page below.
        model,
      }),
    ])
      .then(([stageRunsResp, aggregateResp]) => {
        if (cancelled) return;
        // No client-side changeName filter here (task 25, step 2): the
        // request above already sends "change" and "project", so the
        // server's own scoping is this response's only source of rows --
        // a mock that returns rows for a change other than this one is
        // returning something the real server cannot, and is a defect in
        // the mock, not a reason to keep filtering here.
        const rowsForChange = aggregateResp.recorded ? aggregateResp.rows : [];
        // The stage-run table's own filter: unlike the header (scoped
        // server-side, above), the list endpoint this hook calls has no
        // per-model restriction to send -- a model name is not a fixed
        // field the query allowlist can name (internal/store/query.go's
        // own fixed-identifier requirement) -- so this hook filters the
        // page it already fetched. That page is this one change's own
        // stage runs, bounded to STAGE_RUN_PAGE_SIZE, not the paginated
        // aggregate views DataTable.tsx's own header comment forbids
        // client-side filtering on.
        const stageRuns = model
          ? stageRunsResp.stageRuns.filter((r) => stageRunUsedModel(r, model))
          : stageRunsResp.stageRuns;
        // Task 25, step 4: stageRunsResp.total is the server's own count
        // of every stage run matching (project, change) -- unaffected by
        // the model filter above, which never reaches the server. When it
        // exceeds the page actually fetched, the table below (whichever
        // of stageRuns' rows a model filter left) cannot be a complete
        // picture, and must say so rather than render a count that looks
        // complete and is not.
        const truncated = stageRunsResp.total > stageRunsResp.stageRuns.length;
        setState({
          status: "ready",
          stageRuns,
          summary: summarize(rowsForChange),
          truncated,
          totalStageRuns: stageRunsResp.total,
        });
      })
      .catch((err: unknown) => {
        if (!cancelled) setState({ status: "error", message: err instanceof Error ? err.message : String(err) });
      });

    return () => {
      cancelled = true;
    };
  }, [project, change, model]);

  return state;
}
