// Reviewers -- per review-panel slot's dispatch and finding record across
// a period: how often each slot ran, what it found by severity, and its
// severity-weighted efficiency. An experimental slot (an
// "exp-" id) is badged and its roster description shown behind an info icon (design.md's "Stats views › reviewers").
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { ViewFrame } from "../components/ViewFrame";
import type { ReviewerRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt } from "../format";
import type { ViewProps } from "../viewTypes";

// Severity-weighted findings per dispatch (design.md's reviewers-efficiency):
// Critical 10, Important 4, Minor 1; null when the slot never ran.
function efficiency(r: ReviewerRow): number | null {
  return r.dispatches === 0 ? null : (10 * r.critical + 4 * r.important + r.minor) / r.dispatches;
}

const columns: Column<ReviewerRow>[] = [
  {
    key: "slot",
    header: "Slot",
    sortable: true,
    filterable: true,
    accessor: (r) => r.slot,
    render: (r) =>
      r.experimental ? (
        <span>
          {r.slot}{" "}
          <span
            role="img"
            aria-label="experimental reviewer"
            title={r.description}
          >
            🧪
          </span>
        </span>
      ) : (
        r.slot
      ),
  },
  { key: "dispatches", header: "Dispatches", sortable: true, accessor: (r) => r.dispatches, render: (r) => formatInt(r.dispatches) },
  { key: "changes", header: "Changes", sortable: true, accessor: (r) => r.changes, render: (r) => formatInt(r.changes) },
  { key: "critical", header: "Critical", sortable: true, accessor: (r) => r.critical, render: (r) => formatInt(r.critical) },
  { key: "important", header: "Important", sortable: true, accessor: (r) => r.important, render: (r) => formatInt(r.important) },
  { key: "minor", header: "Minor", sortable: true, accessor: (r) => r.minor, render: (r) => formatInt(r.minor) },
  {
    key: "efficiency",
    header: "Efficiency",
    sortable: true,
    accessor: (r) => efficiency(r) ?? -1,
    render: (r) => {
      const value = efficiency(r);
      return value === null ? "—" : value.toFixed(2);
    },
  },
];

export function Reviewers({ period, onPeriodChange, project, model }: ViewProps) {
  const state = useStatsView<ReviewerRow[]>("reviewers", {
    from: period.from,
    to: period.to,
    project,
    model,
  });

  return (
    <ViewFrame title="Reviewers" description="Per review-panel slot: how often it ran, what it found, and how efficiently.">
      <div className="dashboard-stat-row">
        <Panel title="Slots" state={state}>
          {(data) => <StatPanel label="Slots" value={data.rows.length} format={formatInt} />}
        </Panel>
        <Panel title="Total dispatches" state={state}>
          {(data) => <StatPanel label="Total dispatches" value={data.rows.reduce((sum, r) => sum + r.dispatches, 0)} format={formatInt} />}
        </Panel>
      </div>
      <Panel title="Every reviewer slot" state={state}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={data.rows}
            rowKey={(r) => r.slot}
            emptyMessage="No reviewer dispatches in this period."
            period={period}
            onPeriodChange={onPeriodChange}
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
