// The live state board -- the default route (skills/myflow-do/SKILL.md's
// three-state pipeline, surfaced here instead of a `jq` read of the state
// file). Spec: "it lists every non-archived change with its state, its
// last update, and the command that should run next".
//
// Recomposed onto the panel primitives (task 20): a single "Changes" stat
// panel above the table. A count is the only single number this view's own
// question ("what is the current state of every open change") actually
// has -- "state" is categorical, not something a total or a mean could
// summarise, so no second or third stat panel is invented for it (this
// task's own instruction: "no meaningful single number" gets no panel
// rather than a fabricated one).
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { ViewFrame } from "../components/ViewFrame";
import type { StateBoardRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatDateTime, formatInt } from "../format";
import type { ViewProps } from "../viewTypes";

/**
 * The board's only navigation path into a single change's own dashboard
 * (task 18, "Make it reachable from the board"). Each segment is
 * percent-encoded independently -- App.tsx's parseRunRoute decodes them
 * the same way -- so a project key or change name containing a
 * URL-significant character (a literal "/", "#" or "?") survives the
 * round trip instead of being read as an extra route segment.
 */
function runDetailHref(row: StateBoardRow): string {
  return `#/run/${encodeURIComponent(row.projectKey)}/${encodeURIComponent(row.name)}`;
}

const columns: Column<StateBoardRow>[] = [
  { key: "projectKey", header: "Project", sortable: true, accessor: (r) => r.projectKey, filterable: true },
  {
    key: "name",
    header: "Change",
    sortable: true,
    accessor: (r) => r.name,
    render: (r) => <a href={runDetailHref(r)}>{r.name}</a>,
  },
  { key: "state", header: "State", sortable: true, accessor: (r) => r.state, filterable: true },
  {
    key: "updatedAt",
    header: "Updated",
    sortable: true,
    accessor: (r) => r.updatedAt,
    render: (r) => formatDateTime(r.updatedAt),
  },
  { key: "updatedBy", header: "By", sortable: true, accessor: (r) => r.updatedBy },
  { key: "nextCommand", header: "Next command", accessor: (r) => r.nextCommand || "—" },
];

// The live state board's rows are changes, not stage runs -- the server
// rejects a "model" restriction here with 400 (spec: "A model restriction
// on the live state board"), and App.tsx already disables the dashboard
// bar's model variable on this route for exactly that reason. This view
// therefore never forwards `model` to useStatsView, even if App.tsx ever
// stopped disabling it -- the same defence-in-depth useRunDetail.ts's own
// client-side change filter documents for a different restriction.
export function StateBoard({ period, project }: ViewProps) {
  const state = useStatsView<StateBoardRow[]>("state-board", { from: period.from, to: period.to, project });

  return (
    <ViewFrame title="Live state board" description="Every open change across every project, with its state and its next command.">
      <div className="dashboard-stat-row">
        <Panel title="Changes" state={state}>
          {(data) => <StatPanel label="Changes" value={data.rows.length} format={formatInt} />}
        </Panel>
      </div>
      <Panel title="Every open change" state={state}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={data.rows}
            rowKey={(r) => `${r.projectKey}/${r.name}`}
            emptyMessage="No changes updated in this period."
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
