// ReviewDynamics -- whether the review panel, the implementer and visual
// verify get better or worse over time (KAN-1046). Verdict tiles compare the
// range against the previous range of the same length (previous-range),
// trend charts draw per day or per week with the prompt changes of their own
// area as markers (markers-by-area), and a heatmap shows findings per
// dispatch for each slot × model · effort. The view owns its range presets
// (view-owned-range), and a Family / Version toggle keys every per-model
// chart and the heatmap by model family or by exact model version
// (model-version-split). Layout and copy are the approved prototype's.
import { useEffect, useRef, useState } from "react";
import type { DynamicsSummary, HeatCell, PromptChange, ReviewDynamicsRows } from "../api";
import { DynamicsChart, type DynamicsMarker, type DynamicsSeries } from "../components/DynamicsChart";
import type { Period } from "../components/PeriodPicker";
import { useStatsView } from "../hooks/useStatsView";
import type { ViewProps } from "../viewTypes";

const DAY_MS = 86400000;

type Preset = "1" | "2" | "4" | "8" | "12" | "tune" | "all";

const PRESETS: { id: Preset; label: string }[] = [
  { id: "1", label: "Last week" },
  { id: "2", label: "Last 2 weeks" },
  { id: "4", label: "Last 4 weeks" },
  { id: "8", label: "Last 8 weeks" },
  { id: "12", label: "Last 12 weeks" },
  { id: "tune", label: "Since last prompt change" },
  { id: "all", label: "All time" },
];

// "All time" starts at the epoch: the store holds nothing older.
const ALL_TIME_FROM = new Date("2020-01-01T00:00:00Z");

function presetPeriod(id: Preset, lastPromptChangeAt: string | null): Period | null {
  const to = new Date();
  if (id === "all") return { from: ALL_TIME_FROM, to };
  if (id === "tune") return lastPromptChangeAt ? { from: new Date(lastPromptChangeAt), to } : null;
  return { from: new Date(to.getTime() - Number(id) * 7 * DAY_MS), to };
}

const AREA_NAMES: Record<PromptChange["area"], string> = {
  review: "reviewer",
  impl: "implementer",
  vv: "verifier",
  fix: "panel fixer",
};

function markerWhat(m: PromptChange): string {
  return `${m.kind === "tuning" ? "Prompt tuning" : "Prompt edit"} · ${m.slot ?? AREA_NAMES[m.area]}`;
}

function markerText(m: PromptChange): string {
  return `${m.summary} (${m.sha.slice(0, 7)})`;
}

const mmdd = (iso: string) => iso.slice(5, 10);

function percent(v: number): string {
  return `${(v * 100).toFixed(0)}%`;
}

type Better = "lower" | "higher" | "none";

function Tile({
  label,
  value,
  previous,
  previousChanges,
  better = "lower",
  format = (v) => v.toFixed(2),
}: {
  label: string;
  value: number | null;
  previous: number | null;
  previousChanges: number;
  better?: Better;
  format?: (v: number) => string;
}) {
  let delta = <span className="dynamics-flat">no earlier range</span>;
  if (previousChanges > 0 && value !== null && previous !== null) {
    // From a previous 0 the change has no percentage: any rise reads "from 0".
    const pct = previous === 0 ? (value === 0 ? 0 : Infinity) : ((value - previous) / previous) * 100;
    const arrow = Number.isFinite(pct) ? `${pct < 0 ? "▼" : "▲"} ${Math.abs(pct).toFixed(0)}%` : "▲ from 0";
    const good = better === "lower" ? pct < 0 : pct > 0;
    if (better === "none") delta = <span className="dynamics-flat">{arrow} vs before</span>;
    else if (Math.abs(pct) < 3) delta = <span className="dynamics-flat">≈ unchanged</span>;
    else delta = <span className={good ? "dynamics-good" : "dynamics-bad"}>{`${arrow} · ${good ? "better" : "worse"}`}</span>;
  }
  return (
    <div className="dynamics-tile" data-testid={`tile-${label}`}>
      <div className="dynamics-tile-k">{label}</div>
      <div className="dynamics-tile-v">{value === null ? "—" : format(value)}</div>
      <div className="dynamics-tile-d">{delta}</div>
    </div>
  );
}

function tiles(s: DynamicsSummary, p: DynamicsSummary) {
  const common = { previousChanges: p.changes };
  return (
    <div className="dynamics-tiles">
      <Tile label="Findings per change" value={s.findingsPerChange} previous={p.findingsPerChange} {...common} />
      <Tile label="Critical + Important per change" value={s.criticalImportantPerChange} previous={p.criticalImportantPerChange} {...common} />
      <Tile label="Mean fix rounds" value={s.meanFixRounds} previous={p.meanFixRounds} {...common} />
      <Tile label="Implementer findings per task" value={s.implFindingsPerTask} previous={p.implFindingsPerTask} {...common} />
      <Tile label="Visual verify passed first time" value={s.vvFirstPassShare} previous={p.vvFirstPassShare} better="higher" format={percent} {...common} />
      <Tile label="Changes reviewed" value={s.changes} previous={p.changes} better="none" format={(v) => v.toFixed(0)} {...common} />
    </div>
  );
}

// byEffort orders effort names low → medium → high (then anything else by
// name), and any other names alphabetically.
const EFFORTS = ["low", "medium", "high"];
export function byEffort(a: string, b: string): number {
  const rank = (n: string) => {
    const i = EFFORTS.indexOf(n.split(" ").pop()!);
    return i < 0 ? EFFORTS.length : i;
  };
  return rank(a) - rank(b) || a.localeCompare(b);
}

function Heatmap({ cells, previous, byVersion }: { cells: HeatCell[]; previous: HeatCell[]; byVersion: boolean }) {
  const rate = (c: HeatCell) => (c.dispatches === 0 ? 0 : (c.critical + c.important + c.minor) / c.dispatches);
  const key = (c: { slot: string; model: string; effort: string }) => `${c.slot}|${c.model}|${c.effort}`;
  // Rows run busiest slot first, as the panel roster does; columns run by
  // model, then effort low → high (heatmap-order).
  const dispatches = new Map<string, number>();
  for (const c of cells) dispatches.set(c.slot, (dispatches.get(c.slot) ?? 0) + c.dispatches);
  const slots = [...dispatches.keys()].sort((a, b) => dispatches.get(b)! - dispatches.get(a)! || a.localeCompare(b));
  const cols = [...new Set(cells.map((c) => `${c.model}|${c.effort}`))]
    .map((k) => k.split("|"))
    .sort(([am, ae], [bm, be]) => am.localeCompare(bm) || byEffort(ae, be));
  const byKey = new Map(cells.map((c) => [key(c), c]));
  const prevByKey = new Map(previous.map((c) => [key(c), c]));
  const max = Math.max(0, ...cells.map(rate)) || 1;

  return (
    <div className="dynamics-panel">
      <h2>{`Findings per dispatch: slot × ${byVersion ? "model version" : "model"} · effort`}</h2>
      <p className="dynamics-sub">
        Darker = more findings per dispatch. Arrow: change against the previous range of the same length. Hover a cell
        for its counts.
      </p>
      <div className="dynamics-scroll">
        <table className="dynamics-heat">
          <thead>
            <tr>
              <th className="dynamics-heat-row" />
              {cols.map(([model, effort]) => (
                <th key={`${model}|${effort}`} scope="col">
                  {model}
                  <br />
                  {effort}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {slots.map((slot) => (
              <tr key={slot}>
                <th className="dynamics-heat-row" scope="row">
                  {slot}
                </th>
                {cols.map(([model, effort]) => {
                  const c = byKey.get(key({ slot, model, effort }));
                  if (!c || c.dispatches === 0) {
                    return (
                      <td key={model + effort} className="dynamics-heat-empty">
                        —
                      </td>
                    );
                  }
                  const r = rate(c);
                  const alpha = 0.12 + (0.78 * r) / max;
                  const prev = prevByKey.get(key(c));
                  const change = prev && rate(prev) > 0 ? ((r - rate(prev)) / rate(prev)) * 100 : null;
                  return (
                    <td
                      key={model + effort}
                      style={{
                        background: `rgba(var(--heat), ${alpha.toFixed(2)})`,
                        color: alpha > 0.55 ? "#111217" : "var(--text-primary)",
                      }}
                      title={`${slot} · ${model} ${effort}\n${c.dispatches} dispatches\n${c.critical} critical · ${c.important} important · ${c.minor} minor`}
                    >
                      <div className="dynamics-heat-r">{r.toFixed(2)}</div>
                      {change !== null && (
                        <div className="dynamics-heat-c">
                          {change < 0 ? "▼" : "▲"} {Math.abs(change).toFixed(0)}%
                        </div>
                      )}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function ReviewDynamics({ period, onPeriodChange, project }: ViewProps) {
  const [preset, setPreset] = useState<Preset | null>(() => (/[?&]from=/.test(window.location.hash) ? null : "8"));
  const [byVersion, setByVersion] = useState(false);
  // A URL naming no range opens on Last 8 weeks (view-owned-range), and
  // the first fetch already asks for it, so App's default period is never
  // fetched. The child's effect runs before App's, so the hash still shows
  // whether the URL carried a range.
  const [shown, setShown] = useState<Period>(() => (preset === "8" ? presetPeriod("8", null)! : period));
  const mounted = useRef(false);
  useEffect(() => {
    if (preset === "8") onPeriodChange?.(shown);
  }, []); // Mount only: a later preset click sets the period itself.
  useEffect(() => {
    if (mounted.current) setShown(period);
    mounted.current = true;
  }, [period.from.getTime(), period.to.getTime()]);
  const state = useStatsView<ReviewDynamicsRows>("review-dynamics", { from: shown.from, to: shown.to, project });

  const rows = state.status === "ready" ? state.data.rows : null;
  const choose = (id: Preset) => {
    const next = presetPeriod(id, rows?.lastPromptChangeAt ?? null);
    if (!next) return;
    setPreset(id);
    onPeriodChange?.(next);
  };
  const span = shown.to.getTime() - shown.from.getTime();
  const prevFrom = new Date(shown.from.getTime() - span);

  return (
    <section aria-label="Review dynamics" className="dashboard">
      <h2>Review dynamics</h2>
      <p className="dynamics-sub">
        Every rate is per change or per dispatch, so a busy week doesn't look worse than a quiet one. Lower is better
        unless a tile says otherwise. Hover a chart for the values at that point; hover a purple marker for the prompt
        change it stands for.
      </p>
      <div className="dynamics-bar" role="group" aria-label="Range">
        <span className="dynamics-lbl">Range</span>
        {PRESETS.map((p) => (
          <button
            key={p.id}
            type="button"
            aria-pressed={preset === p.id}
            disabled={p.id === "tune" && !rows?.lastPromptChangeAt}
            onClick={() => choose(p.id)}
          >
            {p.label}
          </button>
        ))}
        <span className="dynamics-lbl">vs</span>
        <span className="dynamics-sub dynamics-inline">
          {`${mmdd(prevFrom.toISOString())} – ${mmdd(new Date(shown.from.getTime() - DAY_MS).toISOString())}`}
        </span>
      </div>
      <div className="dynamics-bar" role="group" aria-label="Models">
        <span className="dynamics-lbl">Models</span>
        <button type="button" aria-pressed={!byVersion} onClick={() => setByVersion(false)}>
          Family
        </button>
        <button type="button" aria-pressed={byVersion} onClick={() => setByVersion(true)}>
          Version
        </button>
        <span className="dynamics-sub dynamics-inline">
          Family merges every version of a model; Version keeps each one apart.
        </span>
      </div>
      {state.status === "loading" && <p role="status">Loading…</p>}
      {state.status === "error" && <p role="alert">Failed to load: {state.message}</p>}
      {rows && <Dynamics rows={rows} byVersion={byVersion} period={shown} />}
    </section>
  );
}

// rangeBuckets is every bucket the period spans -- UTC days, or ISO weeks
// from the Monday on or before from -- so a bucket with no data still
// holds its place on the x axis.
function rangeBuckets(period: Period, kind: "day" | "week"): string[] {
  const start = new Date(Date.UTC(period.from.getUTCFullYear(), period.from.getUTCMonth(), period.from.getUTCDate()));
  if (kind === "week") start.setUTCDate(start.getUTCDate() - ((start.getUTCDay() + 6) % 7));
  const step = kind === "week" ? 7 : 1;
  const out: string[] = [];
  for (const d = start; d < period.to; d.setUTCDate(d.getUTCDate() + step)) out.push(d.toISOString().slice(0, 10));
  return out;
}

function Dynamics({ rows, byVersion, period }: { rows: ReviewDynamicsRows; byVersion: boolean; period: Period }) {
  const buckets = [...new Set([...rangeBuckets(period, rows.bucket), ...rows.series.map((p) => p.bucket)])].sort();
  const index = new Map(buckets.map((b, i) => [b, i]));

  const markersFor = (area: PromptChange["area"] | null): DynamicsMarker[] => {
    if (!area) return [];
    return rows.markers.flatMap((m) => {
      if (m.area !== area) return [];
      // A marker belongs to the last bucket starting on or before it.
      const day = m.at.slice(0, 10);
      let i = -1;
      buckets.forEach((b, k) => {
        if (b <= day) i = k;
      });
      return i < 0 ? [] : [{ index: i, what: markerWhat(m), text: markerText(m) }];
    });
  };

  const seriesOf = (metric: string, names?: Record<string, string>): DynamicsSeries[] => {
    const byName = new Map<string, DynamicsSeries>();
    for (const p of rows.series) {
      if (p.metric !== metric) continue;
      const name = names?.[p.series] ?? p.series;
      let s = byName.get(name);
      if (!s) {
        s = { name, values: buckets.map(() => null), basis: buckets.map(() => null) };
        byName.set(name, s);
      }
      s.values[index.get(p.bucket)!] = p.value;
      s.basis[index.get(p.bucket)!] = p.basis;
    }
    const order = names ? Object.values(names) : [...byName.keys()].sort(byEffort);
    return order.flatMap((n) => byName.get(n) ?? []);
  };

  const chart = (
    title: string,
    description: string,
    series: DynamicsSeries[],
    area: PromptChange["area"] | null,
    basisLabel = "changes",
    format?: (v: number) => string,
  ) => (
    <DynamicsChart
      title={title}
      description={description}
      buckets={buckets}
      bucketKind={rows.bucket}
      series={series}
      markers={markersFor(area)}
      basisLabel={basisLabel}
      format={format}
    />
  );

  const model = byVersion ? "model version" : "model";
  return (
    <>
      {tiles(rows.summary, rows.previousSummary)}
      <div className="dynamics-section">Review panel</div>
      <div className="dynamics-grid">
        {chart(
          "Findings per change by severity",
          "Mean over changes whose panel recorded findings in the bucket.",
          seriesOf("findings-per-change", { critical: "Critical", important: "Important", minor: "Minor" }),
          "review",
        )}
        {chart(
          "Findings per dispatch by slot",
          "A bundled dispatch counts once for each slot it names.",
          seriesOf("per-dispatch-slot"),
          "review",
          "dispatches",
        )}
        {chart(
          "Fix rounds per change",
          "Panel rounds until it closed clean.",
          seriesOf("fix-rounds", { mean: "Mean fix rounds", max: "Max fix rounds" }),
          "fix",
        )}
        {chart(
          `Findings per dispatch by reviewer ${model}`,
          "",
          seriesOf(byVersion ? "per-dispatch-model-version" : "per-dispatch-model"),
          "review",
          "dispatches",
        )}
        {chart("Findings per dispatch by reviewer effort", "", seriesOf("per-dispatch-effort"), "review", "dispatches")}
        {chart("Changes reviewed", "Volume, for context on the rates.", seriesOf("changes"), null, "changes", (v) => v.toFixed(0))}
      </div>
      <div className="dynamics-section">Implementer</div>
      <div className="dynamics-grid">
        {chart(
          "Panel findings per implementer task",
          `By the implementer's ${model} · effort. Fewer means cleaner first drafts.`,
          seriesOf(byVersion ? "impl-findings-per-task-version" : "impl-findings-per-task"),
          "impl",
          "tasks",
        )}
        {chart(
          "Implementer dispatches blocked or aborted",
          "Share of implementer dispatches.",
          seriesOf("impl-blocked-share", { blocked: "Blocked or aborted" }),
          "impl",
          "tasks",
          percent,
        )}
      </div>
      <div className="dynamics-section">Visual verify</div>
      <div className="dynamics-grid">
        {chart(
          "Visual verify rounds per change",
          "Verifier re-runs until the captures matched.",
          seriesOf("vv-rounds", { mean: "Mean verify rounds", max: "Max verify rounds" }),
          "vv",
        )}
        {chart(
          "Visual verify passed first time",
          "Share of changes whose first verifier run passed. Higher is better.",
          seriesOf("vv-first-pass", { "first-pass": "Passed first time" }),
          "vv",
          "changes",
          percent,
        )}
      </div>
      <div className="dynamics-section">Where findings come from</div>
      <Heatmap
        cells={byVersion ? rows.heatVersion : rows.heat}
        previous={byVersion ? rows.previousHeatVersion : rows.previousHeat}
        byVersion={byVersion}
      />
      <p className="dynamics-note">
        Weeks start Monday (UTC); the 1- and 2-week ranges plot days. A finding is dated by when it was recorded.
        Severities fold to Critical / Important / Minor; any other spelling counts as Minor. Markers: applied prompt
        tunings plus past edits to the reviewer, implementer, verifier and fixer prompts, imported from git.
      </p>
    </>
  );
}
