// useRunDetail fetches everything the run detail dashboard
// (views/RunDetail.tsx) needs for one change: its stage runs, and the
// cost-per-change aggregate scoped to that change. **The header's totals
// come from the aggregate, never from summing the fetched stage runs** --
// the round's own spec scenario ("a change with more stage runs than one
// page holds") is exactly this, and summing a page is the defect it names
// (the stats-views requirement "One change opens on its own
// dashboard"). Summing cost-per-change's own rows across a change's
// several (command, stage) groupings is a different thing and is fine:
// that endpoint is not paginated, so its rows are already the server's
// complete answer for the period requested, whatever the change's stage
// runs list is paged to.
import { useEffect, useState } from "react";
import { DAY_MS } from "../format";
import {
  fetchStatsView,
  listStageRuns,
  type CostPerChangeRow,
  type StageRunDTO,
  type StateBoardRow,
} from "../api";

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
      /** The change's live state board row -- state, last update, open
       * stage and session activity -- or undefined when the board does
       * not carry it. */
      board: StateBoardRow | undefined;
      /** True when the change has more stage runs than the fetched page
       * held (totalStageRuns > stageRuns.length). Task 25, step 4: a table
       * built only from the fetched page can show far fewer rows than the
       * change actually has, with the header numbers (from the unpaged
       * aggregate) disagreeing with what is visible -- a count that looks
       * complete and is not. This flag is what the panel uses to say so. */
      truncated: boolean;
      /** The server's own total matching (project, change) -- what
       * "showing N of totalStageRuns" reports alongside `truncated`. */
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

export function useRunDetail(project: string, change: string): RunDetailState {
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
      }),
      // The period is half-open: ending a day past now keeps a change
      // updated at this very instant in it.
      fetchStatsView<StateBoardRow[]>("state-board", { from: WIDE_FROM, to: new Date(Date.now() + DAY_MS), project }),
    ])
      .then(([stageRunsResp, aggregateResp, boardResp]) => {
        if (cancelled) return;
        // No client-side changeName filter here (task 25, step 2): the
        // request above already sends "change" and "project", so the
        // server's own scoping is this response's only source of rows.
        const rowsForChange = aggregateResp.recorded ? aggregateResp.rows : [];
        setState({
          status: "ready",
          stageRuns: stageRunsResp.stageRuns,
          summary: summarize(rowsForChange),
          board: boardResp.rows.find((r) => r.name === change),
          truncated: stageRunsResp.total > stageRunsResp.stageRuns.length,
          totalStageRuns: stageRunsResp.total,
        });
      })
      .catch((err: unknown) => {
        if (!cancelled) setState({ status: "error", message: err instanceof Error ? err.message : String(err) });
      });

    return () => {
      cancelled = true;
    };
  }, [project, change]);

  return state;
}
