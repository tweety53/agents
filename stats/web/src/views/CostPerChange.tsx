// Cost per change -- end-to-end tokens, dollars and wall clock for one
// change, broken down by command and stage. A change spanning several
// repositories is still one row here (design.md, "A change is one unit of
// work"); its repositories are available as a *detail* of that row via
// breakdown=repo, requested only when the row is expanded -- never as
// separate top-level rows, per spec's "A multi-repository change reads as
// one row" requirement.
//
// Recomposed onto the panel primitives (task 20): this view's own question
// -- "what did this period's changes cost, end to end" -- has all three of
// a count, a total and a mean, so it gets all three stat panels, each
// summed from the same rows the table renders (never a second request).
// Every sum stays absence-is-never-zero: a period whose rows are entirely
// unmeasured (totalCostUsd null on every row) reports the stat itself as
// unavailable, not as a fabricated $0 total.
import { useEffect, useState } from "react";
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { Unavailable } from "../components/Unavailable";
import { ViewFrame } from "../components/ViewFrame";
import { fetchStatsView, type CostPerChangeRepoRow, type CostPerChangeRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt, formatMs, formatUsd } from "../format";
import type { ViewProps } from "../viewTypes";

const columns: Column<CostPerChangeRow>[] = [
  { key: "projectKey", header: "Project", sortable: true, accessor: (r) => r.projectKey, filterable: true },
  { key: "changeName", header: "Change", sortable: true, accessor: (r) => r.changeName, filterable: true },
  { key: "command", header: "Command", sortable: true, accessor: (r) => r.command, filterable: true },
  { key: "stage", header: "Stage", sortable: true, accessor: (r) => r.stage },
  { key: "runCount", header: "Runs", sortable: true, accessor: (r) => r.runCount },
  { key: "measuredRuns", header: "Measured", sortable: true, accessor: (r) => r.measuredRuns },
  {
    key: "totalTokensInput",
    header: "Input tokens",
    sortable: true,
    accessor: (r) => r.totalTokensInput,
    render: (r) => <Unavailable value={r.totalTokensInput} format={formatInt} />,
  },
  {
    key: "totalCostUsd",
    header: "Cost",
    sortable: true,
    accessor: (r) => r.totalCostUsd,
    render: (r) => <Unavailable value={r.totalCostUsd} format={formatUsd} />,
  },
  {
    key: "totalDurationMs",
    header: "Duration",
    sortable: true,
    accessor: (r) => r.totalDurationMs,
    render: (r) => <Unavailable value={r.totalDurationMs} format={formatMs} />,
  },
];

// RepoBreakdown is nested under exactly one CostPerChange row and must
// scope to that row alone, never to the change as a whole: cost-per-change
// groups its own rows by (project, change, command, stage) -- one row per
// stage (api.ts's StatsViewParams doc comment; internal/api/stats.go's
// costPerChangeByRepo). Before this, the breakdown summed every stage run
// of the whole change into one panel no matter which row's toggle opened
// it, so two differently-costed rows rendered the identical, unreconciling
// total (post-commit review finding F1, reproduced live). The caption
// below names the row explicitly, so a reader can see at a glance which
// row's figures this panel is meant to reconcile with.
function RepoBreakdown({
  period,
  project,
  changeName,
  command,
  stage,
}: {
  period: ViewProps["period"];
  project: string;
  changeName: string;
  command: string;
  stage: string;
}) {
  const [rows, setRows] = useState<CostPerChangeRepoRow[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetchStatsView<CostPerChangeRepoRow[]>("cost-per-change", {
      from: period.from,
      to: period.to,
      project,
      breakdown: "repo",
      change: changeName,
      command,
      stage,
    })
      .then((data) => {
        if (!cancelled) setRows(data.rows);
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : String(err));
      });
    return () => {
      cancelled = true;
    };
  }, [period.from, period.to, project, changeName, command, stage]);

  if (error) return <p role="alert">Failed to load repository breakdown: {error}</p>;
  if (!rows) return <p role="status">Loading repository breakdown…</p>;
  if (rows.length === 0) return <p>No per-repository detail for this row.</p>;

  return (
    <table className="repo-breakdown">
      <caption>
        Per-repository breakdown for {command} · {stage}
      </caption>
      <thead>
        <tr>
          <th>Repository</th>
          <th>Runs</th>
          <th>Input tokens</th>
          <th>Cost</th>
          <th>Duration</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((r) => (
          <tr key={r.repoRoot ?? "(change-wide)"}>
            <td>{r.repoRoot ?? "(change-wide)"}</td>
            <td>{r.runCount}</td>
            <td>
              <Unavailable value={r.totalTokensInput} format={formatInt} />
            </td>
            <td>
              <Unavailable value={r.totalCostUsd} format={formatUsd} />
            </td>
            <td>
              <Unavailable value={r.totalDurationMs} format={formatMs} />
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

/** Sums a nullable numeric field across rows: null when none of the rows
 * carry it, the sum of the present values otherwise -- the same
 * absence-is-never-zero rule useRunDetail.ts's own summarize() already
 * follows for the run-detail header, applied here to this view's stat
 * panels. Kept local to this file rather than shared: task 20's file list
 * does not include a shared numeric-helpers module, and each view's own
 * "which field, over which rows" differs enough that a shared abstraction
 * would need a callback per call site anyway. */
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

export function CostPerChange({ period, project, model }: ViewProps) {
  const state = useStatsView<CostPerChangeRow[]>("cost-per-change", { from: period.from, to: period.to, project, model });

  return (
    <ViewFrame
      title="Cost per change"
      description="End-to-end tokens, dollars and wall clock for one change, broken down by command and stage."
    >
      <div className="dashboard-stat-row">
        <Panel title="Runs" state={state}>
          {(data) => <StatPanel label="Runs" value={data.rows.reduce((sum, r) => sum + r.runCount, 0)} format={formatInt} />}
        </Panel>
        <Panel title="Total cost" state={state}>
          {(data) => <StatPanel label="Total cost" value={sumNullable(data.rows.map((r) => r.totalCostUsd))} format={formatUsd} />}
        </Panel>
        <Panel title="Mean cost per run" state={state}>
          {(data) => {
            const totalCost = sumNullable(data.rows.map((r) => r.totalCostUsd));
            const totalRuns = data.rows.reduce((sum, r) => sum + r.runCount, 0);
            const mean = totalCost !== null && totalRuns > 0 ? totalCost / totalRuns : null;
            return <StatPanel label="Mean cost per run" value={mean} format={formatUsd} />;
          }}
        </Panel>
      </div>
      <Panel title="By command and stage" state={state}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={data.rows}
            rowKey={(r) => `${r.projectKey}/${r.changeName}/${r.command}/${r.stage}`}
            emptyMessage="No stage runs in this period."
            detailLabel="repository breakdown"
            renderDetail={(r) => (
              <RepoBreakdown period={period} project={r.projectKey} changeName={r.changeName} command={r.command} stage={r.stage} />
            )}
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
