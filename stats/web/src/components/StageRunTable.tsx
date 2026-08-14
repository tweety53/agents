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
import type { StageRunDTO } from "../api";
import {
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
 * this build does not know about, per this task's own requirement. */
function MetricsDetail({ run }: { run: StageRunDTO }) {
  const other = readOtherMetrics(run.metrics);
  const effort = readEffort(run.metrics);
  const fastMode = readFastMode(run.metrics);
  const mainTokens = readMainTokens(run.metrics);
  const sidechainTokens = readSidechainTokens(run.metrics);

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
        detailLabel="raw metrics bag"
        renderDetail={(r) => <MetricsDetail run={r} />}
      />
    </>
  );
}
