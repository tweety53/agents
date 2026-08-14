// Panel economics -- findings per token by review panel roster preset,
// whether "full" earns its cost over "light".
//
// Recomposed onto the panel primitives (task 20): "Total findings" and
// "Total tokens" are straightforward sums across roster rows; "Findings
// per M tok (overall)" is the pooled rate -- total findings over total
// tokens, in millions -- the same reasoning CacheEfficiency.tsx's own
// pooled ratio uses, and for the same reason: averaging each roster's own
// per-row rate would weight a lightly-used roster equally with a heavily
// used one, which answers a different question than "how efficient is
// review spend overall".
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { Unavailable } from "../components/Unavailable";
import { ViewFrame } from "../components/ViewFrame";
import type { PanelEconomicsRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt, formatRatio } from "../format";
import type { ViewProps } from "../viewTypes";

const columns: Column<PanelEconomicsRow>[] = [
  { key: "reviewPanelRoster", header: "Roster", sortable: true, accessor: (r) => r.reviewPanelRoster },
  { key: "findingsTotal", header: "Findings", sortable: true, accessor: (r) => r.findingsTotal },
  {
    key: "tokensTotal",
    header: "Tokens",
    sortable: true,
    accessor: (r) => r.tokensTotal,
    render: (r) => <Unavailable value={r.tokensTotal} format={formatInt} />,
  },
  {
    key: "findingsPerMtok",
    header: "Findings / M tok",
    sortable: true,
    accessor: (r) => r.findingsPerMtok,
    render: (r) => <Unavailable value={r.findingsPerMtok} format={formatRatio} />,
  },
];

/** Same absence-is-never-zero sum this view's siblings use -- see
 * CostPerChange.tsx's own doc comment on why it stays local per file. */
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

export function PanelEconomics({ period, project, model }: ViewProps) {
  const state = useStatsView<PanelEconomicsRow[]>("panel-economics", {
    from: period.from,
    to: period.to,
    project,
    model,
  });

  return (
    <ViewFrame
      title="Panel economics"
      description="Findings per token by review panel roster preset -- whether full earns its cost over light."
    >
      <div className="dashboard-stat-row">
        <Panel title="Total findings" state={state}>
          {(data) => <StatPanel label="Total findings" value={data.rows.reduce((sum, r) => sum + r.findingsTotal, 0)} format={formatInt} />}
        </Panel>
        <Panel title="Total tokens" state={state}>
          {(data) => <StatPanel label="Total tokens" value={sumNullable(data.rows.map((r) => r.tokensTotal))} format={formatInt} />}
        </Panel>
        <Panel title="Findings / M tok (overall)" description="Total findings over total tokens, pooled across every roster." state={state}>
          {(data) => {
            const findings = data.rows.reduce((sum, r) => sum + r.findingsTotal, 0);
            const tokens = sumNullable(data.rows.map((r) => r.tokensTotal));
            const rate = tokens !== null && tokens !== 0 ? findings / (tokens / 1_000_000) : null;
            return <StatPanel label="Findings / M tok" value={rate} format={formatRatio} />;
          }}
        </Panel>
      </div>
      <Panel title="Every roster" state={state}>
        {(data) => (
          <DataTable
            columns={columns}
            rows={data.rows}
            rowKey={(r) => r.reviewPanelRoster}
            emptyMessage="No panel-reviewed stage runs in this period."
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
