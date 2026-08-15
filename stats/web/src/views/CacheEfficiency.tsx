// Cache efficiency -- cache-read against cache-creation per stage, the
// largest single cost lever. Ratio is unavailable (never a fabricated 0 or
// a fake division result) whenever either total is unmeasured or cache
// creation totalled zero (store.CacheEfficiencyRow's own doc comment).
//
// Recomposed onto the panel primitives (task 20): the single number this
// view's own question actually has is the *pooled* ratio -- total cache
// read over total cache creation across every stage -- not an average of
// each stage's own per-row ratio, which would weight a low-volume stage
// the same as a high-volume one and answer a different question than "how
// efficient is the cache overall". That pooled ratio is this view's "mean"
// stat panel; the two totals it is built from are its "total" panels.
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { Unavailable } from "../components/Unavailable";
import { ViewFrame } from "../components/ViewFrame";
import type { CacheEfficiencyRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt, formatRatio } from "../format";
import type { ViewProps } from "../viewTypes";

const columns: Column<CacheEfficiencyRow>[] = [
  { key: "command", header: "Command", sortable: true, accessor: (r) => r.command, filterable: true },
  { key: "stage", header: "Stage", sortable: true, accessor: (r) => r.stage },
  {
    key: "cacheReadTotal",
    header: "Cache read",
    sortable: true,
    accessor: (r) => r.cacheReadTotal,
    render: (r) => <Unavailable value={r.cacheReadTotal} format={formatInt} />,
  },
  {
    key: "cacheCreationTotal",
    header: "Cache creation",
    sortable: true,
    accessor: (r) => r.cacheCreationTotal,
    render: (r) => <Unavailable value={r.cacheCreationTotal} format={formatInt} />,
  },
  {
    key: "ratio",
    header: "Read:creation ratio",
    sortable: true,
    accessor: (r) => r.ratio,
    render: (r) => <Unavailable value={r.ratio} format={formatRatio} />,
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

export function CacheEfficiency({ period, onPeriodChange, project, model }: ViewProps) {
  const state = useStatsView<CacheEfficiencyRow[]>("cache-efficiency", {
    from: period.from,
    to: period.to,
    project,
    model,
  });

  return (
    <ViewFrame title="Cache efficiency" description="cache-read against cache-creation per stage -- the largest single cost lever.">
      <div className="dashboard-stat-row">
        <Panel title="Total cache read" state={state}>
          {(data) => <StatPanel label="Total cache read" value={sumNullable(data.rows.map((r) => r.cacheReadTotal))} format={formatInt} />}
        </Panel>
        <Panel title="Total cache creation" state={state}>
          {(data) => (
            <StatPanel label="Total cache creation" value={sumNullable(data.rows.map((r) => r.cacheCreationTotal))} format={formatInt} />
          )}
        </Panel>
        <Panel title="Overall ratio" description="Total cache read over total cache creation, pooled across every stage." state={state}>
          {(data) => {
            const read = sumNullable(data.rows.map((r) => r.cacheReadTotal));
            const creation = sumNullable(data.rows.map((r) => r.cacheCreationTotal));
            const ratio = read !== null && creation !== null && creation !== 0 ? read / creation : null;
            return <StatPanel label="Overall ratio" value={ratio} format={formatRatio} />;
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
