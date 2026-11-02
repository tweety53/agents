// Weekly costs -- "What did flow spend this week, and what moved it from
// last week?" (decision weekly-costs-question), built as the approved
// prototype draws it (.superpowers/sdd/weekly-costs-prototype.html).
//
// The view owns a KAN-1048 chip bar (chip-filters-period-split-project):
// Period 8w / 12w / 26w, whole Wednesday-00:00-UTC weeks ending with the
// current one; Split by Phase / Model · effort; Project All plus one chip per
// project with spend in the window, from weekly-costs' projects=1 answer. App
// renders no DashboardBar for it. The window is computed when a Period chip
// is chosen and once at mount, never per render: a per-render `new Date()`
// would refetch forever. The model · effort rows are requested only under
// their split (fetch-split-on-demand).
//
// One stacked-bar chart, a bar per week and a segment per series
// (one-stacked-bar-chart), at the prototype's fixed size, never scaled
// (prototype-app-shell-fixed-chart), with the Y axis on tickStep
// (y-axis-ticks-scale). The window's last week is the partial current one,
// drawn translucent and dashed as "(so far)"; the one before is the last full
// week, compared with the week before it by the tiles and the What moved table
// (headline-last-full-week). Series colours run --series-1.. in split order,
// model · effort keeping its six costliest pairs and folding the rest into
// Other models (series-colours-and-other-models). Money is formatUsd
// (money-format-usd); a rise is red ▲, a fall green ▼ (rise-red-fall-green);
// a $0 change reads —, a week with no costed run reads "no data", and a series
// at $0 in both compared weeks is left out of What moved
// (zero-change-and-no-data).
import { useState } from "react";
import type { WeeklyCostPoint, WeeklyModelEffortPoint } from "../api";
import { Chip, ChipGroup } from "../components/Chip";
import { Panel } from "../components/Panel";
import { DAY_MS, formatUsd } from "../format";
import { useStatsView, type StatsViewState } from "../hooks/useStatsView";
import { projectLabel } from "../lib/projectLabel";
import { weekWindow } from "../lib/weekWindow";
import type { ViewProps } from "../viewTypes";

const PERIOD_WEEKS = { "8w": 8, "12w": 12, "26w": 26 } as const;
type PeriodChip = keyof typeof PERIOD_WEEKS;
const SPLITS = ["Phase", "Model · effort"] as const;
type Split = (typeof SPLITS)[number];

interface Filters {
  period: PeriodChip;
  split: Split;
  project: string | null;
  from: Date;
  to: Date;
  weeks: string[];
}

// Computed only when a Period chip is chosen.
function windowFor(period: PeriodChip): Pick<Filters, "period" | "from" | "to" | "weeks"> {
  return { period, ...weekWindow(PERIOD_WEEKS[period], Date.now()) };
}

const phases: { key: keyof Omit<WeeklyCostPoint, "week" | "overallUsd">; name: string }[] = [
  { key: "planningUsd", name: "Planning" },
  { key: "implementationUsd", name: "Implementation" },
  { key: "reviewUsd", name: "Review" },
  { key: "selfReviewUsd", name: "Self-review" },
  { key: "visualVerifyUsd", name: "Visual verify" },
  { key: "otherUsd", name: "Other" },
];

/** One series: a name, its colour and its cost per week of the window. */
interface Series {
  name: string;
  color: string;
  v: number[];
}

const colored = (series: Omit<Series, "color">[]): Series[] => series.map((s, i) => ({ ...s, color: `var(--series-${(i % 7) + 1})` }));
const sum = (xs: number[]) => xs.reduce((a, x) => a + x, 0);

function phaseSeries(weeks: string[], rows: WeeklyCostPoint[]): Series[] {
  const byWeek = new Map(rows.map((r) => [r.week, r]));
  return colored(phases.map(({ key, name }) => ({ name, v: weeks.map((w) => byWeek.get(w)?.[key] ?? 0) })));
}

/** A model · effort series' label: the model id without its "claude-"
 * prefix, then the effort or "session" (decision model-label). */
function modelEffortName(model: string, effort: string): string {
  return `${model.replace(/^claude-/, "")} · ${effort}`;
}

/** The long rows pivoted into one series per pair, costliest over the window
 * first, the six costliest kept and every other pair summed into Other models. */
function modelEffortSeries(weeks: string[], rows: WeeklyModelEffortPoint[]): Series[] {
  const byName = new Map<string, number[]>();
  for (const r of rows) {
    const week = weeks.indexOf(r.week);
    if (week < 0) continue;
    const name = modelEffortName(r.model, r.effort);
    const v = byName.get(name) ?? weeks.map(() => 0);
    v[week] += r.costUsd;
    byName.set(name, v);
  }
  const ranked = [...byName].map(([name, v]) => ({ name, v })).sort((a, b) => sum(b.v) - sum(a.v));
  const rest = ranked.slice(6);
  const kept = ranked.slice(0, 6);
  if (rest.length) kept.push({ name: "Other models", v: weeks.map((_, w) => sum(rest.map((s) => s.v[w]))) });
  return colored(kept);
}

/** The Y axis' tick step: the smallest 1, 2, 2.5 or 5 times a power of ten
 * whose five steps reach `max`, so a $3 week fills the plot as a $300 one does. */
function tickStep(max: number): number {
  const raw = max / 5 || 1;
  const magnitude = 10 ** Math.floor(Math.log10(raw));
  return [1, 2, 2.5, 5, 10].find((m) => m * magnitude >= raw)! * magnitude;
}

/** A change from `b` to `a`: ▲ / ▼ with its $ size and, when `b` cost
 * something, its whole percentage; — when nothing changed. */
function Delta({ a, b }: { a: number; b: number }) {
  const d = a - b;
  if (Math.round(d * 100) === 0) return <span>—</span>;
  const s = d > 0 ? "+" : "−";
  const pct = b ? ` (${s}${Math.round(Math.abs((d / b) * 100))}%)` : "";
  return <span className={d > 0 ? "up" : "down"}>{`${d > 0 ? "▲" : "▼"} ${s}${formatUsd(Math.abs(d))}${pct}`}</span>;
}

const Swatch = ({ color }: { color: string }) => <span className="wc-sw" style={{ background: color }} />;

// Chart geometry, as the prototype draws it.
const CW = 1140,
  CH = 300,
  L = 60,
  R = 10,
  T = 10,
  B = 34,
  PW = CW - L - R,
  PH = CH - T - B;

function Chart({ weeks, series, totals }: { weeks: string[]; series: Series[]; totals: number[] }) {
  const [hover, setHover] = useState<number | null>(null);
  const cur = weeks.length - 1;
  const step = tickStep(Math.max(0, ...totals) * 1.08);
  const y = (v: number) => T + PH - (PH * v) / (step * 5);
  const bw = PW / weeks.length;
  const so = (w: number) => (w === cur ? " (so far)" : "");

  return (
    <div className="wc-plot">
      <svg viewBox={`0 0 ${CW} ${CH}`} width={CW} height={CH}>
        <g className="grid">
          {Array.from({ length: 6 }, (_, i) => (
            <g key={i}>
              <line x1={L} x2={L + PW} y1={y(i * step)} y2={y(i * step)} />
              <text x={L - 6} y={y(i * step) + 4} textAnchor="end">
                {formatUsd(i * step)}
              </text>
            </g>
          ))}
        </g>
        {weeks.map((wk, w) => {
          const x = L + w * bw + bw * 0.18;
          const ww = bw * 0.64;
          let acc = 0;
          return (
            <g key={wk}>
              {!totals[w] && (
                <text x={x + ww / 2} y={y(0) - 6} textAnchor="middle">
                  no data
                </text>
              )}
              {series.map((s) => {
                const v = s.v[w];
                if (!v) return null;
                const top = y(acc + v);
                const height = y(acc) - top;
                acc += v;
                return (
                  <rect
                    key={s.name}
                    className="wc-bar"
                    tabIndex={0}
                    data-testid="wc-segment"
                    data-week={wk}
                    data-series={s.name}
                    aria-label={`${s.name}, week of ${wk}${so(w)}: ${formatUsd(v)}`}
                    x={x}
                    width={ww}
                    y={top}
                    height={height}
                    fill={s.color}
                    {...(w === cur ? { fillOpacity: ".45", stroke: s.color, strokeDasharray: "3 2" } : {})}
                    onMouseEnter={() => setHover(w)}
                    onMouseLeave={() => setHover(null)}
                    onFocus={() => setHover(w)}
                    onBlur={() => setHover(null)}
                  />
                );
              })}
              {totals[w] > 0 && (
                <text x={x + ww / 2} y={y(acc) - 5} textAnchor="middle" className="wc-total" data-testid="wc-total" data-week={wk}>
                  {formatUsd(acc)}
                </text>
              )}
              <text x={x + ww / 2} y={CH - B + 16} textAnchor="middle" data-testid="wc-week">
                {wk.slice(5) + so(w)}
              </text>
            </g>
          );
        })}
      </svg>
      {hover !== null && (
        <div
          className="wc-tip"
          role="tooltip"
          // The last two weeks open the tip leftwards, so it stays inside the plot.
          style={{ left: L + (hover + 0.85) * bw - (hover >= weeks.length - 2 ? 260 : 0), top: 40 }}
        >
          <strong>
            Week of {weeks[hover]}
            {so(hover)}
          </strong>
          <table>
            <tbody>
              {series
                .filter((s) => s.v[hover])
                .map((s) => (
                  <tr key={s.name}>
                    <td>
                      <Swatch color={s.color} />
                      {s.name}
                    </td>
                    <td>{formatUsd(s.v[hover])}</td>
                  </tr>
                ))}
              <tr>
                <td>
                  <strong>Total</strong>
                </td>
                <td>
                  <strong>{formatUsd(totals[hover])}</strong>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

export function WeeklyCosts(_props: ViewProps) {
  const [f, setF] = useState<Filters>(() => ({ ...windowFor("8w"), split: "Phase", project: null }));
  const params = { from: f.from, to: f.to, project: f.project ?? undefined };
  const byModel = f.split === "Model · effort";
  const state = useStatsView<WeeklyCostPoint[]>("weekly-costs", { ...params, projects: true });
  const modelState = useStatsView<WeeklyModelEffortPoint[]>("weekly-model-effort-costs", params, byModel);
  const ok = (s: StatsViewState<unknown>) => s.status === "ready" && s.data.recorded && !s.data.unmeasured;
  const pending: StatsViewState<unknown> | null = !ok(state) ? state : byModel && !ok(modelState) ? modelState : null;
  const listed = state.status === "ready" ? (state.data.projects ?? []).map((p) => p.projectKey) : [];
  // The selected project keeps its chip in a window where it had no spend.
  const projects = f.project && !listed.includes(f.project) ? [...listed, f.project] : listed;
  const title = `Spend per week by ${f.split.toLowerCase()}`;

  const weeks = f.weeks;
  const series =
    pending || state.status !== "ready"
      ? []
      : byModel && modelState.status === "ready"
        ? modelEffortSeries(weeks, modelState.data.rows)
        : phaseSeries(weeks, state.data.rows);
  const totals = weeks.map((_, w) => sum(series.map((s) => s.v[w])));
  const cur = weeks.length - 1;
  const last = cur - 1;
  const prev = cur - 2;
  const day = Math.floor((f.to.getTime() - Date.parse(weeks[cur] + "T00:00:00Z")) / DAY_MS) + 1;
  const full = totals.slice(0, cur).filter((t) => t > 0);
  const moved = series
    .map((s) => ({ ...s, a: s.v[last], b: s.v[prev] }))
    .filter((s) => s.a || s.b)
    .sort((x, y) => Math.abs(y.a - y.b) - Math.abs(x.a - x.b));
  const big = moved.length && Math.round((moved[0].a - moved[0].b) * 100) !== 0 ? moved[0] : null;
  const share = (a: number) => (totals[last] ? (a / totals[last]) * 100 : 0);

  return (
    <section aria-label="Weekly costs" className="dashboard lb wc">
      <div>
        <h2>Weekly costs</h2>
        <p className="lb-sub">What did flow spend this week, and what moved it from last week?</p>
      </div>
      <div className="lb-fbar" role="toolbar" aria-label="Filters">
        <ChipGroup name="Period">
          {(Object.keys(PERIOD_WEEKS) as PeriodChip[]).map((p) => (
            <Chip key={p} on={f.period === p} onClick={() => setF({ ...f, ...windowFor(p) })}>
              {p}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Split by">
          {SPLITS.map((s) => (
            <Chip key={s} on={f.split === s} onClick={() => setF({ ...f, split: s })}>
              {s}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Project">
          <Chip on={f.project === null} onClick={() => setF({ ...f, project: null })}>
            All
          </Chip>
          {projects.map((projectKey) => (
            <Chip key={projectKey} on={f.project === projectKey} onClick={() => setF({ ...f, project: projectKey })} title={projectKey}>
              {projectLabel(projectKey)}
            </Chip>
          ))}
        </ChipGroup>
      </div>
      {pending ? (
        // Loading, error, not-recorded and unmeasured read as Panel renders them.
        <Panel title={title} state={pending}>
          {() => null}
        </Panel>
      ) : (
        <>
          <div className="lb-strip">
            <div className="lb-tile static">
              <div className="lb-lbl">Last full week · {weeks[last]}</div>
              <div className="lb-val">{formatUsd(totals[last])}</div>
              <div className="wc-sub">
                <Delta a={totals[last]} b={totals[prev]} /> vs week of {weeks[prev]}
              </div>
            </div>
            <div className="lb-tile static">
              <div className="lb-lbl">This week so far · day {day} of 7</div>
              <div className="lb-val">{formatUsd(totals[cur])}</div>
              <div className="wc-sub">on pace for {formatUsd((totals[cur] * 7) / day)}</div>
            </div>
            <div className="lb-tile static">
              <div className="lb-lbl">Weekly average · {f.period}</div>
              <div className="lb-val">{full.length ? formatUsd(sum(full) / full.length) : "—"}</div>
              <div className="wc-sub">full weeks only</div>
            </div>
            <div className="lb-tile static">
              <div className="lb-lbl">Biggest mover</div>
              <div className="lb-val wc-mover">{big ? big.name : "—"}</div>
              <div className="wc-sub">{big && <Delta a={big.a} b={big.b} />}</div>
            </div>
          </div>
          <div className="wc-row">
            <div className="lb-panel wc-chartpanel">
              <div className="lb-panel-h">
                <span>{title}</span>
              </div>
              <div className="lb-panel-b wc-chart">
                <div className="wc-legend">
                  {series.map((s) => (
                    <span key={s.name}>
                      <Swatch color={s.color} />
                      {s.name}
                    </span>
                  ))}
                </div>
                <Chart weeks={weeks} series={series} totals={totals} />
              </div>
            </div>
            <div className="lb-panel wc-moved">
              <div className="lb-panel-h">
                <span>
                  What moved, week of {weeks[last]} vs {weeks[prev]}
                </span>
              </div>
              <div className="lb-scroll">
                <table className="lb-table">
                  <thead>
                    <tr>
                      <th>{f.split}</th>
                      <th>{weeks[last]}</th>
                      <th>{weeks[prev]}</th>
                      <th>Change</th>
                      <th>Share of last week</th>
                    </tr>
                  </thead>
                  <tbody>
                    {moved.map((m) => (
                      <tr key={m.name}>
                        <td>
                          <Swatch color={m.color} />
                          {m.name}
                        </td>
                        <td>{formatUsd(m.a)}</td>
                        <td>{formatUsd(m.b)}</td>
                        <td>
                          <Delta a={m.a} b={m.b} />
                        </td>
                        <td>
                          {totals[last] ? `${share(m.a).toFixed(0)}%` : "—"}
                          <span className="wc-sharebar">
                            <span style={{ width: `${share(m.a)}%` }} />
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </>
      )}
    </section>
  );
}
