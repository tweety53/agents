// Model comparison -- the same stage on different models: cost and
// subsequent rework.
//
// Recomposed onto the panel primitives (task 20): "Models compared" (a
// count of the distinct models the period's rows name), "Total runs" and
// "Total rework attempts" (sums) are this view's meaningful single
// numbers. A blended mean cost across every model is not: this view exists
// precisely to keep each model's own cost separate for comparison, so
// averaging them back together would erase the one thing the view is for
// -- the table beneath already carries each model's own mean cost.
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { Unavailable } from "../components/Unavailable";
import { ViewFrame } from "../components/ViewFrame";
import type { ModelComparisonRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt, formatUsd } from "../format";
import type { ViewProps } from "../viewTypes";

const columns: Column<ModelComparisonRow>[] = [
  { key: "model", header: "Model", sortable: true, accessor: (r) => r.model, filterable: true },
  { key: "command", header: "Command", sortable: true, accessor: (r) => r.command, filterable: true },
  { key: "stage", header: "Stage", sortable: true, accessor: (r) => r.stage },
  { key: "runCount", header: "Runs", sortable: true, accessor: (r) => r.runCount },
  {
    key: "meanCostUsd",
    header: "Mean cost",
    sortable: true,
    accessor: (r) => r.meanCostUsd,
    render: (r) => <Unavailable value={r.meanCostUsd} format={formatUsd} />,
  },
  { key: "reworkAttempts", header: "Rework attempts", sortable: true, accessor: (r) => r.reworkAttempts },
];

export function ModelComparison({ period, onPeriodChange, project, model }: ViewProps) {
  const state = useStatsView<ModelComparisonRow[]>("model-comparison", {
    from: period.from,
    to: period.to,
    project,
    model,
  });

  return (
    <ViewFrame title="Model comparison" description="How cost and subsequent rework differ for the same stage across models.">
      <div className="dashboard-stat-row">
        <Panel title="Models compared" state={state}>
          {(data) => <StatPanel label="Models compared" value={new Set(data.rows.map((r) => r.model)).size} format={formatInt} />}
        </Panel>
        <Panel title="Total runs" state={state}>
          {(data) => <StatPanel label="Total runs" value={data.rows.reduce((sum, r) => sum + r.runCount, 0)} format={formatInt} />}
        </Panel>
        <Panel title="Total rework attempts" state={state}>
          {(data) => (
            <StatPanel label="Total rework attempts" value={data.rows.reduce((sum, r) => sum + r.reworkAttempts, 0)} format={formatInt} />
          )}
        </Panel>
      </div>
      <Panel title="By model, command and stage" state={state}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={data.rows}
            rowKey={(r) => `${r.model}/${r.command}/${r.stage}`}
            emptyMessage="No stage runs recording a model in this period."
            period={period}
            onPeriodChange={onPeriodChange}
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
