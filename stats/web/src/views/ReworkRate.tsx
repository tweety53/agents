// Rework rate -- how often /myflow-do re-runs as a fix, and how often a
// stage is abandoned. Derived from recorded attempt numbers and outcomes
// (spec: "Rework is counted from attempts"), never from timing.
//
// Recomposed onto the panel primitives (task 20): "Total attempts",
// "Total rework attempts" and "Total abandoned" are sums across every
// stage; "Rework rate" -- rework attempts over total attempts, pooled --
// is this view's single meaningful "mean", the same pooled-rather-than-
// averaged reasoning CacheEfficiency.tsx and PanelEconomics.tsx already
// use for their own overall rates.
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { ViewFrame } from "../components/ViewFrame";
import type { ReworkRateRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt } from "../format";
import type { ViewProps } from "../viewTypes";

const columns: Column<ReworkRateRow>[] = [
  { key: "command", header: "Command", sortable: true, accessor: (r) => r.command, filterable: true },
  { key: "stage", header: "Stage", sortable: true, accessor: (r) => r.stage },
  { key: "totalAttempts", header: "Total attempts", sortable: true, accessor: (r) => r.totalAttempts },
  { key: "reworkAttempts", header: "Rework attempts", sortable: true, accessor: (r) => r.reworkAttempts },
  { key: "abandonedCount", header: "Abandoned", sortable: true, accessor: (r) => r.abandonedCount },
];

/** A ratio in [0, 1] as a percentage string -- kept local to this file,
 * like format.ts's own siblings, since it is the one view that needs it. */
function formatPercent(value: number): string {
  return `${(value * 100).toFixed(1)}%`;
}

export function ReworkRate({ period, onPeriodChange, project, model }: ViewProps) {
  const state = useStatsView<ReworkRateRow[]>("rework-rate", { from: period.from, to: period.to, project, model });

  return (
    <ViewFrame title="Rework rate" description="How often a command re-runs as a fix, and how often a stage is abandoned.">
      <div className="dashboard-stat-row">
        <Panel title="Total attempts" state={state}>
          {(data) => <StatPanel label="Total attempts" value={data.rows.reduce((sum, r) => sum + r.totalAttempts, 0)} format={formatInt} />}
        </Panel>
        <Panel title="Total rework attempts" state={state}>
          {(data) => (
            <StatPanel label="Total rework attempts" value={data.rows.reduce((sum, r) => sum + r.reworkAttempts, 0)} format={formatInt} />
          )}
        </Panel>
        <Panel title="Total abandoned" state={state}>
          {(data) => (
            <StatPanel label="Total abandoned" value={data.rows.reduce((sum, r) => sum + r.abandonedCount, 0)} format={formatInt} />
          )}
        </Panel>
        <Panel
          title="Overall rework rate"
          description="Rework attempts over total attempts, pooled across every stage."
          state={state}
        >
          {(data) => {
            const total = data.rows.reduce((sum, r) => sum + r.totalAttempts, 0);
            const rework = data.rows.reduce((sum, r) => sum + r.reworkAttempts, 0);
            const rate = total > 0 ? rework / total : null;
            return <StatPanel label="Overall rework rate" value={rate} format={formatPercent} />;
          }}
        </Panel>
      </div>
      <Panel title="Every stage" state={state}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={data.rows}
            rowKey={(r) => `${r.command}/${r.stage}`}
            emptyMessage="No stage runs in this period."
            period={period}
            onPeriodChange={onPeriodChange}
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
