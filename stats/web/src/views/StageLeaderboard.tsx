// Stage leaderboard -- which stages cost the most, mean/median/p90 across
// a period.
//
// Recomposed onto the panel primitives (task 20): "Stages" (a count) and
// "Total runs" (a sum) are meaningful single numbers for this view's own
// question. A blended "mean cost" across every stage is not: this view
// exists precisely because different stages cost differently, and
// averaging their own means together would answer a question nobody asked
// while looking like it answered "what does a run cost" -- exactly the
// invented-number outcome this task's own instructions rule out. The
// table beneath already carries each stage's real mean, median and p90.
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

export function StageLeaderboard({ period, project, model }: ViewProps) {
  const state = useStatsView<StageLeaderboardRow[]>("stage-leaderboard", {
    from: period.from,
    to: period.to,
    project,
    model,
  });

  return (
    <ViewFrame title="Stage leaderboard" description="Which stages cost the most -- mean, median and p90 -- across a period.">
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
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
