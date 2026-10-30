// Stage leaderboard -- the ten stages costing the most in total across a
// period, each with its mean/median/p90 (design.md's
// leaderboard-top10-total). The server returns them ranked by total cost,
// which is the table's order until a reader sorts another column.
//
// Recomposed onto the panel primitives (task 20): "Stages" (a count) and
// "Total runs" (a sum) are meaningful single numbers for this view's own
// question. A blended "mean cost" across every stage is not: this view
// exists precisely because different stages cost differently, and
// averaging their own means together would answer a question nobody asked
// while looking like it answered "what does a run cost" -- exactly the
// invented-number outcome this task's own instructions rule out. The
// table beneath already carries each stage's real mean, median and p90.
//
// "Main session" and "Main share", after "Total cost", are the stage's
// cost outside its dispatched agents and that cost's share of the total
// (decision leaderboard-main-columns); "—" where no run yields a figure.
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { ViewFrame } from "../components/ViewFrame";
import type { StageLeaderboardRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt, formatUsd } from "../format";
import type { ViewProps } from "../viewTypes";

const columns: Column<StageLeaderboardRow>[] = [
  { key: "command", header: "Command", sortable: true, accessor: (r) => r.command, filterable: true },
  { key: "stage", header: "Stage", sortable: true, accessor: (r) => r.stage },
  { key: "runCount", header: "Runs", sortable: true, accessor: (r) => r.runCount },
  {
    key: "totalCostUsd",
    header: "Total cost",
    sortable: true,
    accessor: (r) => r.totalCostUsd,
    render: (r) => formatUsd(r.totalCostUsd),
  },
  {
    key: "mainCostUsd",
    header: "Main session",
    sortable: true,
    accessor: (r) => r.mainCostUsd,
    render: (r) => (r.mainCostUsd === null ? "—" : formatUsd(r.mainCostUsd)),
  },
  {
    key: "mainSharePct",
    header: "Main share",
    sortable: true,
    accessor: (r) => r.mainSharePct,
    render: (r) => (r.mainSharePct === null ? "—" : `${Math.round(r.mainSharePct)}%`),
  },
  {
    key: "meanCostUsd",
    header: "Mean cost",
    sortable: true,
    accessor: (r) => r.meanCostUsd,
    render: (r) => formatUsd(r.meanCostUsd),
  },
  {
    key: "medianCostUsd",
    header: "Median cost",
    sortable: true,
    accessor: (r) => r.medianCostUsd,
    render: (r) => formatUsd(r.medianCostUsd),
  },
  {
    key: "p90CostUsd",
    header: "P90 cost",
    sortable: true,
    accessor: (r) => r.p90CostUsd,
    render: (r) => formatUsd(r.p90CostUsd),
  },
];

export function StageLeaderboard({ period, onPeriodChange, project, model }: ViewProps) {
  const state = useStatsView<StageLeaderboardRow[]>("stage-leaderboard", {
    from: period.from,
    to: period.to,
    project,
    model,
  });

  return (
    <ViewFrame title="Stage leaderboard" description="The ten stages costing the most in total -- with mean, median and p90 -- across a period.">
      <div className="dashboard-stat-row">
        <Panel title="Stages" state={state}>
          {(data) => <StatPanel label="Stages" value={data.rows.length} format={formatInt} />}
        </Panel>
        <Panel title="Total runs" state={state}>
          {(data) => <StatPanel label="Total runs" value={data.rows.reduce((sum, r) => sum + r.runCount, 0)} format={formatInt} />}
        </Panel>
      </div>
      <Panel title="Every costed stage" state={state}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={data.rows}
            rowKey={(r) => `${r.command}/${r.stage}`}
            emptyMessage="No costed stage runs in this period."
            period={period}
            onPeriodChange={onPeriodChange}
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
