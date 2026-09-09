// Decisions -- every recorded planner decision (design.md's "The decision
// chain") joined to its own run's wall-clock, tokens, cost, findings and
// fallback/timed-out dispatch counts, so a roster, model or execution-shape
// choice can be tuned from data rather than memory (design.md's "Stats
// views › decisions"). A summary table above the per-run table groups the
// same rows by class × execution -- count, mean wall-clock, mean cost, mean
// findings -- client-side: the server's rows are already the whole,
// unpaged, period-bounded result (DataTable.tsx's own header comment), and
// there is no new server-side aggregation to duplicate by grouping the
// small set already in the browser.
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { ViewFrame } from "../components/ViewFrame";
import type { DecisionRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt, formatMs, formatUsd } from "../format";
import type { ViewProps } from "../viewTypes";

interface ClassExecutionSummaryRow {
  key: string;
  class: string;
  execution: string;
  grouping: string;
  count: number;
  meanWallClockSeconds: number;
  meanCostUsd: number;
  meanFindings: number;
}

// Grouped by class, execution and grouping -- design.md's "Stats views ›
// decisions" summary is class × execution, and a run's own bundle grouping
// (static / free / default) is a fourth axis the same table now carries: a
// free run and a static run of the same class and execution are not
// comparable runs to average together.
function summarizeByClassAndExecution(rows: DecisionRow[]): ClassExecutionSummaryRow[] {
  const groups = new Map<string, DecisionRow[]>();
  for (const row of rows) {
    const key = `${row.class} ${row.execution} ${row.grouping}`;
    const group = groups.get(key);
    if (group) group.push(row);
    else groups.set(key, [row]);
  }
  return Array.from(groups.entries())
    .map(([key, group]) => {
      const [cls, execution, grouping] = key.split(" ");
      const count = group.length;
      const totalFindings = (r: DecisionRow) => r.critical + r.important + r.minor;
      return {
        key,
        class: cls,
        execution,
        grouping,
        count,
        meanWallClockSeconds: group.reduce((sum, r) => sum + r.wallClockSeconds, 0) / count,
        meanCostUsd: group.reduce((sum, r) => sum + r.costUsd, 0) / count,
        meanFindings: group.reduce((sum, r) => sum + totalFindings(r), 0) / count,
      };
    })
    .sort(
      (a, b) =>
        a.class.localeCompare(b.class) ||
        a.execution.localeCompare(b.execution) ||
        a.grouping.localeCompare(b.grouping),
    );
}

const summaryColumns: Column<ClassExecutionSummaryRow>[] = [
  {
    key: "classExecution",
    header: "Class / execution",
    accessor: (r) => `${r.class} / ${r.execution}`,
  },
  { key: "grouping", header: "Grouping", accessor: (r) => r.grouping },
  { key: "count", header: "Runs", accessor: (r) => r.count, render: (r) => formatInt(r.count) },
  {
    key: "meanWallClockSeconds",
    header: "Mean wall-clock",
    accessor: (r) => r.meanWallClockSeconds,
    render: (r) => formatMs(r.meanWallClockSeconds * 1000),
  },
  {
    key: "meanCostUsd",
    header: "Mean cost",
    accessor: (r) => r.meanCostUsd,
    render: (r) => formatUsd(r.meanCostUsd),
  },
  {
    key: "meanFindings",
    header: "Mean findings",
    accessor: (r) => r.meanFindings,
    render: (r) => r.meanFindings.toFixed(2),
  },
];

const columns: Column<DecisionRow>[] = [
  { key: "change", header: "Change", sortable: true, filterable: true, accessor: (r) => r.change },
  {
    key: "class",
    header: "Class",
    sortable: true,
    filterable: true,
    accessor: (r) => r.class,
    render: (r) => (r.overridden ? `${r.class} ↑` : r.class),
  },
  { key: "execution", header: "Execution", sortable: true, filterable: true, accessor: (r) => r.execution },
  { key: "implementerModel", header: "Implementer", sortable: true, accessor: (r) => r.implementerModel },
  { key: "rosterSize", header: "Roster", sortable: true, accessor: (r) => r.rosterSize, render: (r) => formatInt(r.rosterSize) },
  { key: "rerun", header: "Rerun", sortable: true, accessor: (r) => r.rerun },
  { key: "grouping", header: "Grouping", sortable: true, filterable: true, accessor: (r) => r.grouping },
  { key: "dispatches", header: "Dispatches", accessor: (r) => r.dispatches },
  { key: "implementerGroups", header: "Implementer groups", accessor: (r) => r.implementerGroups },
  {
    key: "wallClockSeconds",
    header: "Wall-clock",
    sortable: true,
    accessor: (r) => r.wallClockSeconds,
    render: (r) => formatMs(r.wallClockSeconds * 1000),
  },
  { key: "costUsd", header: "Cost", sortable: true, accessor: (r) => r.costUsd, render: (r) => formatUsd(r.costUsd) },
  {
    key: "findings",
    header: "Findings (C/I/Mi)",
    accessor: (r) => r.critical + r.important + r.minor,
    render: (r) => `${r.critical}/${r.important}/${r.minor}`,
  },
  { key: "fixRounds", header: "Fix rounds", sortable: true, accessor: (r) => r.fixRounds, render: (r) => formatInt(r.fixRounds) },
  { key: "fallbacks", header: "Fallbacks", sortable: true, accessor: (r) => r.fallbacks, render: (r) => formatInt(r.fallbacks) },
  { key: "timedOut", header: "Timed out", sortable: true, accessor: (r) => r.timedOut, render: (r) => formatInt(r.timedOut) },
];

export function Decisions({ period, onPeriodChange, project }: ViewProps) {
  const state = useStatsView<DecisionRow[]>("decisions", {
    from: period.from,
    to: period.to,
    project,
  });

  return (
    <ViewFrame
      title="Decisions"
      description="Every recorded planner decision joined to its own run's wall-clock, cost and findings."
    >
      <div className="dashboard-stat-row">
        <Panel title="Runs" state={state}>
          {(data) => <StatPanel label="Runs" value={data.rows.length} format={formatInt} />}
        </Panel>
      </div>
      <Panel title="By class and execution" state={state}>
        {(data) => (
          <DataTable
            columns={summaryColumns}
            rows={summarizeByClassAndExecution(data.rows)}
            rowKey={(r) => r.key}
            emptyMessage="No decisions in this period."
          />
        )}
      </Panel>
      <Panel title="Every decision" state={state}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={data.rows}
            rowKey={(r) => `${r.project}/${r.change}/${r.recordedAt}`}
            emptyMessage="No decisions in this period."
            period={period}
            onPeriodChange={onPeriodChange}
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
