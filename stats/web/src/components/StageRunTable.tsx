// StageRunTable is the run-detail dashboard's own table (task 18, "The run
// detail dashboard, and the board rows that reach it"): one row per stage
// run, with every reader from metrics.ts routed through Unavailable so an
// unmeasured metric on an otherwise-measured run reads as unavailable, not
// zero -- the round's own single most important rule, per-metric and
// per-run. Duration is computed from startedAt/endedAt directly (not read
// from the metrics bag): a run with no endedAt is open, not zero-length,
// and renders as "still running" rather than through Unavailable, since
// that is a run-lifecycle fact, not an absent measurement.
import { DataTable, type Column } from "./DataTable";
import { Unavailable } from "./Unavailable";
import type { DispatchRow, StageRunDTO } from "../api";
import {
  asObject,
  readCacheReadTokens,
  readCostUsd,
  readEffort,
  readFastMode,
  readInputTokens,
  readMainTokens,
  readModels,
  readOtherMetrics,
  readOutputTokens,
  readSidechainTokens,
} from "../metrics";
import { formatDateTime, formatDurationOrOpen, formatInt, formatUsd } from "../format";

// --- per-dispatch cost (task 5, KAN-201; nested here per F3, pass 1 of
// this change's own review panel) ---
//
// Originally a second, hand-rolled expand/collapse affordance rendered as
// a flat list below this table in views/RunDetail.tsx, keyed by a string
// synthesized from `startedAt` because nothing in that flat list tied a
// toggle back to the row it belonged to. Nested here instead, each stage
// run's dispatches ride the SAME per-row detail toggle MetricsDetail
// already uses -- DataTable's own `renderDetail`/`rowKey` mechanism
// (DataTable.tsx:223-250), keyed by `stageRunId` (this file's own
// `rowKey`, below), which is unique per stage run by construction. No
// synthesized identifier, no second toggle implementation: one expand
// affordance per row, reused for whatever detail that row carries.

/** A bucket's own token grand total (main + sidechain), or undefined when
 * neither bucket carried a chargeable field -- metrics.ts's
 * readMainTokens/readSidechainTokens both read `<bag>.tokens.<key>`, which
 * is exactly the shape a dispatch bucket and a model bucket each carry at
 * their own top level, so they are reused here unchanged rather than
 * re-implemented. */
function tokenTotal(bag: unknown): number | undefined {
  const main = readMainTokens(bag);
  const sidechain = readSidechainTokens(bag);
  if (main === null && sidechain === null) return undefined;
  return (main ?? 0) + (sidechain ?? 0);
}

/**
 * Derives one stage run's dispatch rows from its metrics bag's
 * `dispatches.<agentId>` key (internal/harvest.DispatchBucket,
 * attribute.go), sorted by cost descending with an absent cost sorting
 * last -- never treated as zero (myflow-stats-views spec.md's own
 * ordering rule).
 *
 * Per-dispatch cost is read directly from that same key's own `cost_usd`
 * -- store.Store.Price (internal/store/pricing.go) now prices
 * "dispatches.<agentId>" through the identical rate resolution and
 * chargeableTokens.cost arithmetic as "models.<model>", applied to the
 * dispatch's own recorded model and own tokens (myflow-stats-views
 * spec.md, "Per-dispatch cost SHALL be derived through the same pricing
 * path every other cost figure uses"). This is deliberately never an
 * implied average rate scaled from a model bucket's blended total: that
 * approach assumes a dispatch shares its model bucket's token mix, which
 * a cache-heavy or output-heavy dispatch does not, and silently
 * mis-prices it either way. A dispatch with no `cost_usd` (no recorded
 * model, no rate in effect for it, or an unpriceable token composition --
 * pricing.go's own Price doc comment) renders unavailable, never a
 * fabricated cost.
 */
function readDispatchRows(bag: unknown): DispatchRow[] {
  const obj = asObject(bag);
  const dispatches = obj ? asObject(obj.dispatches) : null;
  if (!dispatches) return [];

  const rows: DispatchRow[] = Object.entries(dispatches).map(([agentId, raw]) => {
    const d = asObject(raw);
    const model = d && typeof d.model === "string" ? d.model : undefined;
    const tokens = tokenTotal(d);
    const costUsd = d ? readCostUsd(d) : null;

    return {
      agentId,
      agentType: d && typeof d.agent_type === "string" ? d.agent_type : undefined,
      description: d && typeof d.description === "string" ? d.description : undefined,
      model,
      tokens,
      costUsd: costUsd ?? undefined,
    };
  });

  return rows.sort((a, b) => {
    if (a.costUsd === undefined && b.costUsd === undefined) return 0;
    if (a.costUsd === undefined) return 1;
    if (b.costUsd === undefined) return -1;
    return b.costUsd - a.costUsd;
  });
}

/** A string descriptor cell, rendered unavailable rather than as an empty
 * string when the dispatch's sidecar never carried it -- this table's own
 * "Model" column already establishes this local pattern for a
 * string-typed metric Unavailable's own `format` prop cannot express. */
function DescriptorCell({ value }: { value: string | undefined }) {
  return value === undefined ? <Unavailable value={null} /> : <span data-testid="measured">{value}</span>;
}

/** Wrapped in the same `.data-table-scroll` convention DataTable.tsx:203
 * wraps every other table in (styles.css:372, "A wide table's overflow
 * must stay contained rather than widening the panel or scrolling the
 * page body") -- a long free-text description can overflow this table's
 * width even though it never overflows the outer detail row. */
function DispatchTable({ rows }: { rows: DispatchRow[] }) {
  return (
    <div className="data-table-scroll">
      <table className="dispatch-table">
        <thead>
          <tr>
            <th>Description</th>
            <th>Agent type</th>
            <th>Model</th>
            <th>Tokens</th>
            <th>Cost</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.agentId} data-testid="dispatch-row">
              <td>
                <DescriptorCell value={row.description} />
              </td>
              <td>
                <DescriptorCell value={row.agentType} />
              </td>
              <td>
                <DescriptorCell value={row.model} />
              </td>
              <td>
                <Unavailable value={row.tokens ?? null} format={formatInt} />
              </td>
              <td>
                <Unavailable value={row.costUsd ?? null} format={formatUsd} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/** Wall-clock duration in ms, or null when the run is still open (no endedAt). Used for sorting only -- rendering goes through formatDurationOrOpen. */
function durationMs(run: StageRunDTO): number | null {
  if (!run.endedAt) return null;
  return new Date(run.endedAt).getTime() - new Date(run.startedAt).getTime();
}

function DurationCell({ run }: { run: StageRunDTO }) {
  if (!run.endedAt) {
    return (
      <span className="stage-run-open" data-testid="run-open">
        {formatDurationOrOpen(run.startedAt, run.endedAt)}
      </span>
    );
  }
  return <span data-testid="measured">{formatDurationOrOpen(run.startedAt, run.endedAt)}</span>;
}

const columns: Column<StageRunDTO>[] = [
  { key: "command", header: "Command", sortable: true, accessor: (r) => r.command, filterable: true },
  { key: "stage", header: "Stage", sortable: true, accessor: (r) => r.stage, filterable: true },
  { key: "attempt", header: "Attempt", sortable: true, accessor: (r) => r.attempt },
  { key: "outcome", header: "Outcome", sortable: true, accessor: (r) => r.outcome ?? "" },
  {
    key: "duration",
    header: "Duration",
    sortable: true,
    accessor: (r) => durationMs(r),
    render: (r) => <DurationCell run={r} />,
  },
  {
    key: "tokensInput",
    header: "Tokens in",
    sortable: true,
    accessor: (r) => readInputTokens(r.metrics),
    render: (r) => <Unavailable value={readInputTokens(r.metrics)} format={formatInt} />,
  },
  {
    key: "tokensOutput",
    header: "Tokens out",
    sortable: true,
    accessor: (r) => readOutputTokens(r.metrics),
    render: (r) => <Unavailable value={readOutputTokens(r.metrics)} format={formatInt} />,
  },
  {
    key: "tokensCached",
    header: "Tokens cached",
    sortable: true,
    accessor: (r) => readCacheReadTokens(r.metrics),
    render: (r) => <Unavailable value={readCacheReadTokens(r.metrics)} format={formatInt} />,
  },
  {
    key: "cost",
    header: "Cost",
    sortable: true,
    accessor: (r) => readCostUsd(r.metrics),
    render: (r) => <Unavailable value={readCostUsd(r.metrics)} format={formatUsd} />,
  },
  {
    key: "model",
    header: "Model",
    sortable: true,
    filterable: true,
    // A stage run's model column shows every model the run's metrics bag
    // recorded, not one of them: the models bucket is keyed by however
    // many models actually touched the run (readModels' own doc comment),
    // and a review-panel stage commonly used two. Picking one -- the
    // defect this task exists to fix -- would misattribute the run to a
    // model it only partly used.
    accessor: (r) => readModels(r.metrics)?.join(", ") ?? null,
    render: (r) => {
      const models = readModels(r.metrics);
      return models === null ? (
        <Unavailable value={null} />
      ) : (
        <span data-testid="measured">{models.join(", ")}</span>
      );
    },
  },
];

/** Every run's whole raw metrics bag, expanded on demand -- including keys
 * this build does not know about, per this task's own requirement. Also
 * carries the run's own per-dispatch cost breakdown (task 5, KAN-201) when
 * its metrics bag recorded one -- see this file's own header comment on
 * why that rides this same detail rather than a second toggle. */
function MetricsDetail({ run }: { run: StageRunDTO }) {
  const other = readOtherMetrics(run.metrics);
  const effort = readEffort(run.metrics);
  const fastMode = readFastMode(run.metrics);
  const mainTokens = readMainTokens(run.metrics);
  const sidechainTokens = readSidechainTokens(run.metrics);
  const dispatchRows = readDispatchRows(run.metrics);

  return (
    <dl className="stage-run-detail">
      <div>
        <dt>Started</dt>
        <dd>{formatDateTime(run.startedAt)}</dd>
      </div>
      <div>
        <dt>Ended</dt>
        <dd>{run.endedAt ? formatDateTime(run.endedAt) : "still running"}</dd>
      </div>
      <div>
        <dt>Effort</dt>
        <dd>{effort === null ? <Unavailable value={null} /> : <span data-testid="measured">{effort}</span>}</dd>
      </div>
      <div>
        <dt>Fast mode</dt>
        <dd>{fastMode === null ? <Unavailable value={null} /> : <span data-testid="measured">{String(fastMode)}</span>}</dd>
      </div>
      <div>
        <dt>Main tokens</dt>
        <dd>
          <Unavailable value={mainTokens} format={formatInt} />
        </dd>
      </div>
      <div>
        <dt>Sidechain tokens</dt>
        <dd>
          <Unavailable value={sidechainTokens} format={formatInt} />
        </dd>
      </div>
      {Object.keys(other).length > 0 && (
        <div className="stage-run-detail-raw">
          <dt>Other recorded metrics</dt>
          <dd>
            <pre>{JSON.stringify(other, null, 2)}</pre>
          </dd>
        </div>
      )}
      {dispatchRows.length > 0 && (
        <div className="stage-run-detail-raw stage-run-dispatches">
          <dt>Dispatches ({dispatchRows.length})</dt>
          <dd>
            <DispatchTable rows={dispatchRows} />
          </dd>
        </div>
      )}
    </dl>
  );
}

export interface StageRunTableProps {
  stageRuns: StageRunDTO[];
  /** True when the change has more stage runs than the fetched page held
   * (useRunDetail's own `truncated`) -- task 25, step 4: rather than
   * silently rendering a short list that looks complete, this states the
   * shortfall directly above the table. */
  truncated?: boolean;
  /** The server's own total stage-run count for this change, reported
   * alongside `truncated`. */
  totalStageRuns?: number;
}

export function StageRunTable({ stageRuns, truncated, totalStageRuns }: StageRunTableProps) {
  return (
    <>
      {truncated && (
        <p role="status" className="stage-run-table-truncated">
          Showing {stageRuns.length} of {totalStageRuns} stage runs recorded for this change — narrow the
          period or view the change directly for the rest.
        </p>
      )}
      <DataTable
        columns={columns}
        rows={stageRuns}
        rowKey={(r) => String(r.stageRunId)}
        emptyMessage="No stage runs recorded for this change."
        detailLabel="run details"
        renderDetail={(r) => <MetricsDetail run={r} />}
      />
    </>
  );
}
