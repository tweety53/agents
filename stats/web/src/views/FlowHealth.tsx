// Flow health -- whether flow's own machinery earns its keep: which guards
// ever fire, which stages get re-entered, and how many review-panel rounds
// a change needs. Every other view asks what the pipeline costs; this one
// asks where it loops, so a guard that never catches anything can be found
// and removed, and a stage that often needs a second pass can be
// tightened.
//
// One dashboard over three endpoints (internal/api/health.go), each in its
// own panel so one failing or empty source never hides the other two. No
// model control applies: none of the three counts per model.
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { Unavailable } from "../components/Unavailable";
import { ViewFrame } from "../components/ViewFrame";
import type { GuardActivityRow, PanelRoundsRow, StageRedoRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatDateTime, formatInt, formatMs, formatRatio } from "../format";
import type { ViewProps } from "../viewTypes";

function formatShare(value: number): string {
  return `${Math.round(value * 100)}%`;
}

/** part/whole as a share, or null when there is no whole to divide by --
 * a guard that never ran has no fire rate, not a rate of zero. */
function share(part: number, whole: number): number | null {
  return whole > 0 ? part / whole : null;
}

function formatSeconds(value: number): string {
  return formatMs(value * 1000);
}

const guardColumns: Column<GuardActivityRow>[] = [
  { key: "guard", header: "Guard", sortable: true, filterable: true, accessor: (r) => r.guard },
  { key: "runs", header: "Runs", sortable: true, accessor: (r) => r.runs, render: (r) => formatInt(r.runs) },
  { key: "fired", header: "Fired", sortable: true, accessor: (r) => r.fired, render: (r) => formatInt(r.fired) },
  {
    key: "fireRate",
    header: "Fire rate",
    sortable: true,
    accessor: (r) => share(r.fired, r.runs),
    render: (r) => <Unavailable value={share(r.fired, r.runs)} format={formatShare} />,
  },
  {
    key: "cannotAnswer",
    header: "Could not answer",
    sortable: true,
    accessor: (r) => r.cannotAnswer,
    render: (r) => formatInt(r.cannotAnswer),
  },
  {
    key: "lastFiredAt",
    header: "Last fired",
    sortable: true,
    accessor: (r) => r.lastFiredAt,
    render: (r) => (r.lastFiredAt ? formatDateTime(r.lastFiredAt) : r.runs > 0 ? "never" : "—"),
  },
  {
    key: "medianDurationMs",
    header: "Median runtime",
    sortable: true,
    accessor: (r) => r.medianDurationMs,
    render: (r) => <Unavailable value={r.medianDurationMs} format={formatMs} />,
  },
  { key: "verdicts", header: "Verdicts", sortable: true, accessor: (r) => r.verdicts, render: (r) => formatInt(r.verdicts) },
  {
    key: "falsePositives",
    header: "False positives",
    sortable: true,
    accessor: (r) => r.falsePositives,
    render: (r) => formatInt(r.falsePositives),
  },
];

const redoColumns: Column<StageRedoRow>[] = [
  { key: "command", header: "Command", sortable: true, filterable: true, accessor: (r) => r.command },
  { key: "stage", header: "Stage", sortable: true, accessor: (r) => r.stage },
  { key: "changes", header: "Changes", sortable: true, accessor: (r) => r.changes, render: (r) => formatInt(r.changes) },
  { key: "runs", header: "Runs", sortable: true, accessor: (r) => r.runs, render: (r) => formatInt(r.runs) },
  {
    key: "reentries",
    header: "Re-entries",
    sortable: true,
    accessor: (r) => r.reentries,
    render: (r) => formatInt(r.reentries),
  },
  {
    key: "reenteredShare",
    header: "Changes re-entered",
    sortable: true,
    accessor: (r) => share(r.reenteredChanges, r.changes),
    render: (r) => <Unavailable value={share(r.reenteredChanges, r.changes)} format={formatShare} />,
  },
  {
    key: "medianSeconds",
    header: "Median time",
    sortable: true,
    accessor: (r) => r.medianSeconds,
    render: (r) => <Unavailable value={r.medianSeconds} format={formatSeconds} />,
  },
  {
    key: "p90Seconds",
    header: "P90 time",
    sortable: true,
    accessor: (r) => r.p90Seconds,
    render: (r) => <Unavailable value={r.p90Seconds} format={formatSeconds} />,
  },
];

const roundsColumns: Column<PanelRoundsRow>[] = [
  { key: "change", header: "Change", sortable: true, filterable: true, accessor: (r) => r.change },
  { key: "project", header: "Project", sortable: true, accessor: (r) => r.project },
  {
    key: "startedAt",
    header: "Panel started",
    sortable: true,
    accessor: (r) => r.startedAt,
    render: (r) => formatDateTime(r.startedAt),
  },
  {
    key: "rounds",
    header: "Rounds",
    sortable: true,
    accessor: (r) => r.rounds,
    render: (r) => <Unavailable value={r.rounds} format={formatInt} />,
  },
  { key: "findings", header: "Findings", sortable: true, accessor: (r) => r.findings, render: (r) => formatInt(r.findings) },
  { key: "critical", header: "Critical", sortable: true, accessor: (r) => r.critical, render: (r) => formatInt(r.critical) },
  { key: "important", header: "Important", sortable: true, accessor: (r) => r.important, render: (r) => formatInt(r.important) },
  { key: "minor", header: "Minor", sortable: true, accessor: (r) => r.minor, render: (r) => formatInt(r.minor) },
];

/** Mean rounds over the changes that recorded any -- null, not zero, when
 * none did. */
function meanRounds(rows: PanelRoundsRow[]): number | null {
  const recorded = rows.flatMap((r) => (r.rounds === null ? [] : [r.rounds]));
  return recorded.length > 0 ? recorded.reduce((a, b) => a + b, 0) / recorded.length : null;
}

export function FlowHealth({ period, onPeriodChange, project }: ViewProps) {
  const params = { from: period.from, to: period.to, project };
  const guards = useStatsView<GuardActivityRow[]>("guards", params);
  const redo = useStatsView<StageRedoRow[]>("stage-redo", params);
  const rounds = useStatsView<PanelRoundsRow[]>("panel-rounds", params);

  return (
    <ViewFrame
      title="Flow health"
      description="Where flow loops rather than what it costs: which guards ever fire, which stages get re-entered, and how many review-panel rounds a change needs."
    >
      <div className="dashboard-stat-row">
        <Panel title="Guards that fired" state={guards}>
          {(data) => <StatPanel label="Guards that fired" value={data.rows.filter((r) => r.fired > 0).length} format={formatInt} />}
        </Panel>
        <Panel title="Guards that ran and never fired" state={guards}>
          {(data) => (
            <StatPanel
              label="Guards that ran and never fired"
              value={data.rows.filter((r) => r.runs > 0 && r.fired === 0).length}
              format={formatInt}
            />
          )}
        </Panel>
        <Panel title="Total re-entries" state={redo}>
          {(data) => <StatPanel label="Total re-entries" value={data.rows.reduce((s, r) => s + r.reentries, 0)} format={formatInt} />}
        </Panel>
        <Panel title="Mean panel rounds" state={rounds}>
          {(data) => <StatPanel label="Mean panel rounds" value={meanRounds(data.rows)} format={formatRatio} />}
        </Panel>
      </div>
      <Panel
        title="Guards"
        description="Runs are recorded by the flow-guard binary for every ported guard; verdicts and false positives by the guards that keep one. A guard that runs and never fires is a candidate for removal."
        state={guards}
      >
        {(data) => (
          <DataTable
            columns={guardColumns}
            rows={data.rows}
            rowKey={(r) => r.guard}
            emptyMessage="No guard ran in this period."
            period={period}
            onPeriodChange={onPeriodChange}
          />
        )}
      </Panel>
      <Panel
        title="Stage re-entries"
        description="A re-entry is the same stage begun again on the same change: a fix run, a resumed session, a retried gate."
        state={redo}
      >
        {(data) => (
          <DataTable
            columns={redoColumns}
            rows={data.rows}
            rowKey={(r) => `${r.command}/${r.stage}`}
            emptyMessage="No stage ran in this period."
            period={period}
            onPeriodChange={onPeriodChange}
          />
        )}
      </Panel>
      <Panel title="Review-panel rounds per change" state={rounds}>
        {(data) => (
          <DataTable
            columns={roundsColumns}
            rows={data.rows}
            rowKey={(r) => `${r.project}/${r.change}`}
            emptyMessage="No change entered the review panel in this period."
            period={period}
            onPeriodChange={onPeriodChange}
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
