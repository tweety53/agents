// Runs -- one row per run (a change's /flow-plan, /flow and /flow-fast
// invocations, keyed by session token), grouped under the change with a
// change subtotal, so a plan session sits beside the runs it seeded
// (docs/superpowers/specs/2026-09-12-run-stats-design.md). Expanding a row
// reuses DataTable's renderDetail: the main session first, then one row
// per dispatch with declared versus served model and effort.
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { Unavailable } from "../components/Unavailable";
import { ViewFrame } from "../components/ViewFrame";
import type { ChangeRuns, RunDispatchRow, RunRow, RunTotals } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatDateTime, formatInt, formatMs, formatRatio, formatUsd } from "../format";
import type { ViewProps } from "../viewTypes";

function served(declared: string, servedMap: Record<string, number> | null): string {
  const keys = Object.keys(servedMap ?? {});
  if (keys.length === 0) return declared;
  const others = keys.filter((k) => k !== declared.toLowerCase());
  return others.length === 0 ? declared : `${declared} (served ${others.join(", ")})`;
}

function findingsCell(d: RunDispatchRow): string {
  if (d.findingsRaised === 0) return "0";
  const parts = Object.entries(d.findingsByStatus ?? {}).sort(([a], [b]) => a.localeCompare(b)).map(([k, v]) => `${k} ${v}`);
  return `${d.findingsRaised} (${parts.join(", ")})`;
}

function errorsCell(t: RunTotals) {
  const total = t.toolErrors + t.apiErrors;
  return <span title={`tool errors ${t.toolErrors}, denials ${t.denials}, api errors ${t.apiErrors}`}>{formatInt(total)}</span>;
}

const runColumns: Column<RunRow>[] = [
  { key: "kind", header: "Kind", sortable: true, filterable: true, accessor: (r) => r.kind },
  { key: "command", header: "Command", accessor: (r) => r.command },
  { key: "startedAt", header: "Started", sortable: true, accessor: (r) => r.startedAt, render: (r) => formatDateTime(r.startedAt) },
  { key: "in", header: "Tokens in", sortable: true, accessor: (r) => r.totals.inputTokens, render: (r) => formatInt(r.totals.inputTokens) },
  { key: "out", header: "Tokens out", sortable: true, accessor: (r) => r.totals.outputTokens, render: (r) => formatInt(r.totals.outputTokens) },
  { key: "cache", header: "Cache hit", sortable: true, accessor: (r) => r.totals.cacheHitRatio, render: (r) => <Unavailable value={r.totals.cacheHitRatio} format={formatRatio} /> },
  { key: "cost", header: "Cost", sortable: true, accessor: (r) => r.totals.costUsd, render: (r) => <Unavailable value={r.totals.costUsd} format={formatUsd} /> },
  { key: "wall", header: "Wall clock", sortable: true, accessor: (r) => r.totals.wallClockMs, render: (r) => formatMs(r.totals.wallClockMs) },
  { key: "gate", header: "Human gate", sortable: true, accessor: (r) => r.totals.humanGateMs, render: (r) => formatMs(r.totals.humanGateMs) },
  { key: "compactions", header: "Compactions", sortable: true, accessor: (r) => r.totals.compactions, render: (r) => formatInt(r.totals.compactions) },
  { key: "turns", header: "Turns", sortable: true, accessor: (r) => r.totals.turns, render: (r) => formatInt(r.totals.turns) },
  { key: "tools", header: "Tool calls", sortable: true, accessor: (r) => r.totals.toolCalls, render: (r) => formatInt(r.totals.toolCalls) },
  { key: "errors", header: "Errors", sortable: true, accessor: (r) => r.totals.toolErrors + r.totals.apiErrors, render: (r) => errorsCell(r.totals) },
  { key: "fanOut", header: "Fan-out", sortable: true, accessor: (r) => r.fanOutMax, render: (r) => formatInt(r.fanOutMax) },
  { key: "suite", header: "Suite 1st pass", accessor: (r) => (r.suiteFirstPass === null ? "" : String(r.suiteFirstPass)),
    render: (r) => (r.suiteRuns === 0 ? <Unavailable value={null} /> : <span>{r.suiteFirstPass ? "pass" : "fail"} ({formatInt(r.suiteRuns)})</span>) },
  // Not filterable: the dropdown option and the cell's own text would be
  // indistinguishable to a query scoped to the whole document once a
  // change has only one run (Runs.test.tsx) -- every other filterable
  // column in this view has more than one candidate cell/option pairing
  // in practice, so this is the one column where that collision is real.
  { key: "execution", header: "Execution", accessor: (r) => r.decision?.execution ?? "" },
  { key: "implementer", header: "Implementer", accessor: (r) => r.decision?.implementer ?? "" },
  { key: "panel", header: "Panel", accessor: (r) => r.decision?.panel ?? "" },
];

function TotalsCells({ t }: { t: RunTotals }) {
  return (
    <>
      <td>{formatInt(t.inputTokens)}</td>
      <td>{formatInt(t.outputTokens)}</td>
      <td><Unavailable value={t.cacheHitRatio} format={formatRatio} /></td>
      <td><Unavailable value={t.costUsd} format={formatUsd} /></td>
      <td>{formatMs(t.wallClockMs)}</td>
      <td>{formatInt(t.turns)}</td>
      <td>{formatInt(t.toolCalls)}</td>
      <td>{errorsCell(t)}</td>
      <td>{formatInt(t.compactions)}</td>
      <td><Unavailable value={t.contextEnd} format={formatInt} /></td>
    </>
  );
}

function RunDetail({ run }: { run: RunRow }) {
  return (
    <div className="data-table-scroll">
      <table className="dispatch-table">
        <thead>
          <tr>
            <th>Session</th><th>Agent type</th><th>Depth</th><th>Model</th><th>Effort</th><th>Findings</th>
            <th>Tokens in</th><th>Tokens out</th><th>Cache hit</th><th>Cost</th><th>Wall clock</th>
            <th>Turns</th><th>Tool calls</th><th>Errors</th><th>Compactions</th><th>Context end</th>
          </tr>
        </thead>
        <tbody>
          <tr data-testid="main-session-row">
            <td>main session</td><td></td><td>0</td><td></td><td></td><td></td>
            <TotalsCells t={run.main} />
          </tr>
          {run.dispatches.map((d: RunDispatchRow) => (
            <tr key={d.seq} data-testid="dispatch-row">
              <td>{d.seq} {d.role}{d.slot ? ` / ${d.slot}` : ""}{d.description ? ` — ${d.description}` : ""}</td>
              <td>{d.agentType}</td>
              <td>{d.depth ?? ""}</td>
              <td className={d.mismatch ? "dispatch-mismatch" : undefined} data-testid={d.mismatch ? "dispatch-mismatch" : undefined}>
                {served(d.declaredModel, d.servedModels)}
              </td>
              <td>{served(d.declaredEffort, d.servedEfforts)}</td>
              <td>{findingsCell(d)}</td>
              <TotalsCells t={d.totals} />
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function ChangePanel({ group, period, onPeriodChange }: { group: ChangeRuns } & Pick<ViewProps, "period" | "onPeriodChange">) {
  const title = group.change ?? `${group.jiraKey} (no change yet)`;
  return (
    <section className="panel" aria-label={title}>
      <h3 className="panel-title">{title}</h3>
      <p className="panel-description">
        {group.jiraKey && group.change ? `${group.jiraKey} · ` : ""}
        cost <Unavailable value={group.totals.costUsd} format={formatUsd} /> · tokens in {formatInt(group.totals.inputTokens)} ·
        idle between runs {formatMs(group.idleBetweenRunsMs)} · fix iterations {formatInt(group.fixIterations)}
      </p>
      <div className="panel-body">
        <DataTable
          columns={runColumns}
          rows={group.runs}
          rowKey={(r) => r.sessionToken || r.startedAt}
          emptyMessage="No runs in this period."
          detailLabel="run details"
          renderDetail={(r) => <RunDetail run={r} />}
          period={period}
          onPeriodChange={onPeriodChange}
        />
      </div>
    </section>
  );
}

export function Runs({ period, onPeriodChange, project }: ViewProps) {
  const state = useStatsView<ChangeRuns[]>("runs", { from: period.from, to: period.to, project });
  return (
    <ViewFrame title="Runs" description="Every run of every change — plan, creating, fix, integrate — with its main session and dispatches.">
      <Panel title="Runs by change" state={state}>
        {(data) => (
          <>
            {data.rows.length === 0 && <p>No runs in this period.</p>}
            {data.rows.map((g) => (
              <ChangePanel key={`${g.project}/${g.change ?? g.jiraKey}`} group={g} period={period} onPeriodChange={onPeriodChange} />
            ))}
          </>
        )}
      </Panel>
    </ViewFrame>
  );
}
