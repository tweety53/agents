// Decisions -- recorded planner decisions (design.md's "The decision
// chain") joined to their own runs' wall-clock, cost and findings, so a
// roster, model or execution-shape choice can be tuned from data rather
// than memory (design.md's "Stats views › decisions"). Two summaries group
// the rows -- by class × execution, and by implementer model (KAN-989's
// decisions-by-implementer-model) -- client-side: the server's rows are
// already the whole, unpaged, period-bounded result (DataTable.tsx's own
// header comment), and there is no new server-side aggregation to
// duplicate by grouping the small set already in the browser.
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
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
function groupBy(rows: DecisionRow[], key: (r: DecisionRow) => string): Map<string, DecisionRow[]> {
  const groups = new Map<string, DecisionRow[]>();
  for (const row of rows) {
    const k = key(row);
    const group = groups.get(k);
    if (group) group.push(row);
    else groups.set(k, [row]);
  }
  return groups;
}

const totalFindings = (r: DecisionRow) => r.critical + r.important + r.minor;

function summarizeByClassAndExecution(rows: DecisionRow[]): ClassExecutionSummaryRow[] {
  return Array.from(groupBy(rows, (r) => `${r.class} ${r.execution} ${r.grouping}`).entries())
    .map(([key, group]) => {
      const [cls, execution, grouping] = key.split(" ");
      const count = group.length;
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

interface ImplementerModelSummaryRow {
  implementerModel: string;
  count: number;
  meanCostUsd: number;
  meanWallClockSeconds: number;
  meanFindings: number;
  meanFixRounds: number;
}

// Grouped on the recorded implementerModel string as-is: an inline or
// default run groups under the string its decision recorded ("skipped —
// inline", "default"), never folded into a model.
function summarizeByImplementerModel(rows: DecisionRow[]): ImplementerModelSummaryRow[] {
  const mean = (group: DecisionRow[], value: (r: DecisionRow) => number) =>
    group.reduce((sum, r) => sum + value(r), 0) / group.length;
  return Array.from(groupBy(rows, (r) => r.implementerModel).entries())
    .map(([implementerModel, group]) => ({
      implementerModel,
      count: group.length,
      meanCostUsd: mean(group, (r) => r.costUsd),
      meanWallClockSeconds: mean(group, (r) => r.wallClockSeconds),
      meanFindings: mean(group, totalFindings),
      meanFixRounds: mean(group, (r) => r.fixRounds),
    }))
    .sort((a, b) => b.count - a.count || a.implementerModel.localeCompare(b.implementerModel));
}

const implementerModelColumns: Column<ImplementerModelSummaryRow>[] = [
  { key: "implementerModel", header: "Implementer model", accessor: (r) => r.implementerModel },
  { key: "count", header: "Runs", accessor: (r) => r.count, render: (r) => formatInt(r.count) },
  { key: "meanCostUsd", header: "Mean cost", accessor: (r) => r.meanCostUsd, render: (r) => formatUsd(r.meanCostUsd) },
  {
    key: "meanWallClockSeconds",
    header: "Mean wall-clock",
    accessor: (r) => r.meanWallClockSeconds,
    render: (r) => formatMs(r.meanWallClockSeconds * 1000),
  },
  { key: "meanFindings", header: "Mean findings", accessor: (r) => r.meanFindings, render: (r) => r.meanFindings.toFixed(2) },
  { key: "meanFixRounds", header: "Mean fix rounds", accessor: (r) => r.meanFixRounds, render: (r) => r.meanFixRounds.toFixed(2) },
];

export function Decisions({ period, project }: ViewProps) {
  const state = useStatsView<DecisionRow[]>("decisions", {
    from: period.from,
    to: period.to,
    project,
  });

  return (
    <ViewFrame
      title="Decisions"
      description="Recorded planner decisions, summarised by class and execution and by implementer model."
    >
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
      <Panel title="By implementer model" state={state}>
        {(data) => (
          <DataTable
            columns={implementerModelColumns}
            rows={summarizeByImplementerModel(data.rows)}
            rowKey={(r) => r.implementerModel}
            emptyMessage="No decisions in this period."
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
