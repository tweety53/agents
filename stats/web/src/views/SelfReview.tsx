// SelfReview -- every self-review finding recorded in a period: what each
// angle found, whether it was fixed (linked to its commit), filed,
// declined or noted (design.md's self-review-view).
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { ViewFrame } from "../components/ViewFrame";
import type { SelfReviewRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt } from "../format";
import type { ViewProps } from "../viewTypes";

function count(rows: SelfReviewRow[], disposition: SelfReviewRow["disposition"]): number {
  return rows.filter((r) => r.disposition === disposition).length;
}

const columns: Column<SelfReviewRow>[] = [
  { key: "recordedAt", header: "Recorded", sortable: true, accessor: (r) => r.recordedAt, render: (r) => r.recordedAt.slice(0, 10) },
  { key: "change", header: "Change", sortable: true, filterable: true, accessor: (r) => r.change },
  { key: "angle", header: "Angle", sortable: true, filterable: true, accessor: (r) => r.angle },
  { key: "note", header: "Finding", accessor: (r) => r.note },
  { key: "disposition", header: "Outcome", sortable: true, filterable: true, accessor: (r) => r.disposition },
  {
    key: "ref",
    header: "Ref",
    accessor: (r) => r.ref,
    render: (r) => (r.commitUrl ? <a href={r.commitUrl}>{r.ref}</a> : r.ref),
  },
];

export function SelfReview({ period, onPeriodChange, project }: ViewProps) {
  const state = useStatsView<SelfReviewRow[]>("self-review", {
    from: period.from,
    to: period.to,
    project,
  });

  return (
    <ViewFrame title="Self-review fixes" description="Every self-review finding: its angle, whether it was fixed, filed, declined or noted, and the commit that fixed it.">
      <div className="dashboard-stat-row">
        <Panel title="Fixed" state={state}>
          {(data) => <StatPanel label="Fixed" value={count(data.rows, "fixed")} format={formatInt} />}
        </Panel>
        <Panel title="Filed" state={state}>
          {(data) => <StatPanel label="Filed" value={count(data.rows, "filed")} format={formatInt} />}
        </Panel>
      </div>
      <Panel title="Every finding" state={state}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={data.rows}
            rowKey={(r) => `${r.project}/${r.change}/${r.recordedAt}/${r.angle}/${r.note}`}
            emptyMessage="No self-review findings in this period."
            period={period}
            onPeriodChange={onPeriodChange}
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
