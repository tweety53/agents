// Trend over time -- is the pipeline getting cheaper per change as myflow
// changes? One measure (daily total cost) over time is a time series's job.
//
// Recomposed onto the panel primitives (task 20, step 2): the bespoke
// inline-SVG bar chart this view used to draw for itself is replaced by
// TimeSeriesPanel, task 19's shared line-chart primitive -- the same
// null-breaks-the-line rule applies (a day with no costed run draws a gap,
// never a point at zero), now enforced in one place instead of
// re-implemented per view. The table stays beneath it, unchanged in what
// it shows.
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { TimeSeriesPanel } from "../components/TimeSeriesPanel";
import { Unavailable } from "../components/Unavailable";
import { ViewFrame } from "../components/ViewFrame";
import type { TrendPoint } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt, formatUsd } from "../format";
import type { ViewProps } from "../viewTypes";

const columns: Column<TrendPoint>[] = [
  { key: "day", header: "Day", sortable: true, accessor: (r) => r.day },
  { key: "runCount", header: "Runs", sortable: true, accessor: (r) => r.runCount },
  {
    key: "totalCostUsd",
    header: "Total cost",
    sortable: true,
    accessor: (r) => r.totalCostUsd,
    render: (r) => <Unavailable value={r.totalCostUsd} format={formatUsd} />,
  },
];

/** Same absence-is-never-zero sum this view's siblings use for their own
 * stat panels -- see CostPerChange.tsx's own doc comment on why it stays
 * local to each file rather than shared. */
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

export function Trend({ period, onPeriodChange, project, model }: ViewProps) {
  const state = useStatsView<TrendPoint[]>("trend", { from: period.from, to: period.to, project, model });

  return (
    <ViewFrame title="Trend over time" description="Is the pipeline getting cheaper per change as myflow changes?">
      <div className="dashboard-stat-row">
        <Panel title="Days" state={state}>
          {(data) => <StatPanel label="Days" value={data.rows.length} format={formatInt} />}
        </Panel>
        <Panel title="Total cost" state={state}>
          {(data) => <StatPanel label="Total cost" value={sumNullable(data.rows.map((r) => r.totalCostUsd))} format={formatUsd} />}
        </Panel>
      </div>
      <Panel title="Daily cost" description="Each day's total cost -- a gap, not a $0 point, on a day with no costed run." state={state}>
        {(data) => <TimeSeriesPanel points={data.rows.map((r) => ({ day: r.day, value: r.totalCostUsd }))} format={formatUsd} />}
      </Panel>
      <Panel title="By day" state={state}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={data.rows}
            rowKey={(r) => r.day}
            emptyMessage="No stage runs in this period."
            period={period}
            onPeriodChange={onPeriodChange}
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
