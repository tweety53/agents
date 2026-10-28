// Weekly costs -- where does flow spend go, week by week? Weeks run
// Wednesday 00:00 UTC to the next Wednesday. Two charts cover the last 8
// weeks, the current partial week included, over every project and model:
// the view takes no filter (KAN-1038), so App renders no filter bar for it.
// The first draws each phase, plus Other, as a toggleable line; the second
// (KAN-1041) one line per model · effort pair, the parent session per model
// as its own line. Every point carries a permanent pill with its cost and
// its change from the same series' previous week. A series with no costed
// run that week is a gap in its line, never $0, and the week after a gap,
// like the window's first week, shows its cost alone.
import { useMemo, useState } from "react";
import { Panel } from "../components/Panel";
import { ViewFrame } from "../components/ViewFrame";
import type { WeeklyCostPoint, WeeklyModelEffortPoint } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatUsd } from "../format";
import type { ViewProps } from "../viewTypes";

const DAY_MS = 24 * 60 * 60 * 1000;
const WEEKS = 8;

const categories: { key: keyof Omit<WeeklyCostPoint, "week">; name: string }[] = [
  { key: "overallUsd", name: "Overall" },
  { key: "planningUsd", name: "Planning" },
  { key: "implementationUsd", name: "Implementation" },
  { key: "reviewUsd", name: "Review" },
  { key: "selfReviewUsd", name: "Self-review" },
  { key: "visualVerifyUsd", name: "Visual verify" },
  { key: "otherUsd", name: "Other" },
];

/** One chart line: a name and its cost per week of the window, null where
 * the week carries none. */
interface Series {
  name: string;
  values: (number | null)[];
}

/** The last 8 Wednesday-to-Wednesday weeks: from this week's UTC
 * Wednesday 00:00 minus 7 weeks, to now. */
export function lastEightWeeks(now: Date): { from: Date; to: Date } {
  const midnight = Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate());
  const sinceWednesday = (now.getUTCDay() + 4) % 7;
  return { from: new Date(midnight - (sinceWednesday + (WEEKS - 1) * 7) * DAY_MS), to: now };
}

/** Every week of `from`'s 8-week window, in order, as YYYY-MM-DD. */
function everyWeek(from: Date): string[] {
  return Array.from({ length: WEEKS }, (_, i) => new Date(from.getTime() + i * 7 * DAY_MS).toISOString().slice(0, 10));
}

/** One series per phase category, null in a week the response carries no
 * row for. */
function phaseSeries(weeks: string[], rows: WeeklyCostPoint[]): Series[] {
  const byWeek = new Map(rows.map((r) => [r.week, r]));
  return categories.map(({ key, name }) => ({ name, values: weeks.map((w) => byWeek.get(w)?.[key] ?? null) }));
}

/** A model · effort line's label: the model id without its "claude-"
 * prefix, then the effort or "session" (decision model-label). */
function modelEffortName(model: string, effort: string): string {
  return `${model.replace(/^claude-/, "")} · ${effort}`;
}

/** The long model·effort rows pivoted into one series per pair, ordered by
 * total cost, highest first. */
function modelEffortSeries(weeks: string[], rows: WeeklyModelEffortPoint[]): Series[] {
  const byName = new Map<string, (number | null)[]>();
  for (const r of rows) {
    const week = weeks.indexOf(r.week);
    if (week < 0) continue;
    const name = modelEffortName(r.model, r.effort);
    const values = byName.get(name) ?? weeks.map(() => null);
    values[week] = (values[week] ?? 0) + r.costUsd;
    byName.set(name, values);
  }
  const total = (v: (number | null)[]) => v.reduce<number>((sum, c) => sum + (c ?? 0), 0);
  return [...byName]
    .map(([name, values]) => ({ name, values }))
    .sort((a, b) => total(b.values) - total(a.values));
}

/** The change from `previous` to `value` in percent, or null where either
 * week is unmeasured or the previous week cost nothing. */
function change(value: number, previous: number | null): number | null {
  if (previous === null || previous === 0) return null;
  return ((value - previous) / previous) * 100;
}

function deltaClass(pct: number): string {
  return pct > 0 ? "delta-up" : pct < 0 ? "delta-down" : "delta-flat";
}

function deltaText(pct: number, digits: number, gap: string): string {
  const size = `${Math.abs(pct).toFixed(digits)}%`;
  return pct === 0 ? size : `${pct > 0 ? "▲" : "▼"}${gap}${size}`;
}

/** The Y axis' tick step: the smallest 1, 2, 2.5 or 5 times a power of ten
 * whose five steps reach `max`, so a $3 week fills the plot as a $300 one does. */
function tickStep(max: number): number {
  const raw = max / 5 || 1;
  const magnitude = 10 ** Math.floor(Math.log10(raw));
  return [1, 2, 2.5, 5, 10].find((m) => m * magnitude >= raw)! * magnitude;
}

const WIDTH = 960;
const HEIGHT = 460;
const LEFT = 56;
const RIGHT = 110;
const TOP = 16;
const BOTTOM = 48;
const PLOT_W = WIDTH - LEFT - RIGHT;
const PLOT_H = HEIGHT - TOP - BOTTOM;
const LABEL_GAP = 17;

interface Point {
  series: number;
  week: number;
  value: number;
  y: number;
  pct: number | null;
}

interface Label extends Point {
  labelY: number;
}

/** Per week column, the cost labels sorted by y and pushed apart to
 * LABEL_GAP, then shifted up together where the stack overruns the plot. */
function stackLabels(points: Point[]): Label[] {
  const out: Label[] = [];
  for (let week = 0; week < WEEKS; week++) {
    const column = points.filter((p) => p.week === week).sort((a, b) => a.y - b.y);
    let last = -Infinity;
    const stacked = column.map((p) => {
      last = Math.max(p.y, last + LABEL_GAP);
      return { ...p, labelY: last };
    });
    const over = last - (TOP + PLOT_H);
    out.push(...stacked.map((l) => (over > 0 ? { ...l, labelY: l.labelY - over } : l)));
  }
  return out;
}

interface ChartProps {
  weeks: string[];
  series: Series[];
  ariaLabel: string;
  /** Whether every line starts shown; otherwise the first alone does. */
  allShown?: boolean;
}

function WeeklyCostChart({ weeks, series: lineSeries, ariaLabel, allShown = false }: ChartProps) {
  const [shown, setShown] = useState(() => lineSeries.map((_, i) => allShown || i === 0));
  const [hovered, setHovered] = useState<{ series: number; week: number } | null>(null);

  const visible = lineSeries
    .flatMap((s, n) => (shown[n] ? s.values : []))
    .filter((v): v is number => v !== null);
  const step = tickStep(Math.max(0, ...visible));
  const x = (week: number) => LEFT + (PLOT_W * week) / (WEEKS - 1);
  const y = (value: number) => TOP + PLOT_H - (PLOT_H * value) / (step * 5);

  const lines = lineSeries.map(({ values }, series) =>
    values.map((value, week): Point | null => {
      if (value === null) return null;
      return { series, week, value, y: y(value), pct: week ? change(value, values[week - 1]) : null };
    }),
  );
  const drawn = lines.filter((_, n) => shown[n]);
  const points = drawn.flat().filter((p): p is Point => p !== null);
  const labels = stackLabels(points);
  const tip = hovered && points.find((p) => p.series === hovered.series && p.week === hovered.week);
  const color = (series: number) => `var(--series-${(series % 7) + 1})`;
  const name = (series: number) => lineSeries[series].name;

  return (
    <div className="weekly-chart">
      <div className="weekly-toggles">
        {lineSeries.map(({ name }, n) => (
          <button
            key={name}
            type="button"
            aria-pressed={shown[n]}
            onClick={() => setShown(shown.map((on, i) => (i === n ? !on : on)))}
          >
            <span className="weekly-swatch" style={{ background: color(n) }} />
            {name}
          </button>
        ))}
      </div>
      <div className="weekly-plot">
        <svg viewBox={`0 0 ${WIDTH} ${HEIGHT}`} role="img" aria-label={ariaLabel}>
          <g className="weekly-axis">
            {Array.from({ length: 6 }, (_, k) => (
              <g key={k}>
                <line x1={LEFT} x2={WIDTH - RIGHT} y1={y(k * step)} y2={y(k * step)} />
                <text x={LEFT - 8} y={y(k * step) + 4} textAnchor="end" data-testid="cost-tick">
                  {formatUsd(k * step)}
                </text>
              </g>
            ))}
            {weeks.map((week, n) => (
              <text key={week} x={x(n)} y={HEIGHT - BOTTOM + 18} textAnchor="middle" data-testid="week-tick">
                {week}
              </text>
            ))}
            <text x={LEFT + PLOT_W / 2} y={HEIGHT - 6} textAnchor="middle">
              Week starting (Wed, UTC)
            </text>
            <text transform="rotate(-90)" x={-(TOP + PLOT_H / 2)} y={14} textAnchor="middle">
              Cost (USD)
            </text>
          </g>
          {drawn.map((line, n) => {
            // A null week lifts the pen, so the line breaks there.
            const d = line.map((p, week) => (p ? `${week && line[week - 1] ? "L" : "M"}${x(p.week)} ${p.y}` : "")).join("");
            const first = line.find((p) => p);
            // A series with no measured week in the window draws no line.
            if (!first) return null;
            const series = first.series;
            return (
              d && (
                <path
                  key={n}
                  d={d}
                  fill="none"
                  className="weekly-line"
                  style={{ stroke: color(series) }}
                  data-testid="weekly-line"
                  data-series={name(series)}
                />
              )
            );
          })}
          {labels.map((l) => {
            const lx = x(l.week) + 8;
            const cost = formatUsd(l.value);
            const delta = l.pct === null ? "" : ` ${deltaText(l.pct, 0, "")}`;
            // The pill fits its text: about 6px a character at 10px, 6px padding each side.
            const width = 12 + (cost.length + delta.length) * 6;
            return (
              <g key={`${l.series}-${l.week}`} className="weekly-delta">
                {Math.abs(l.labelY - l.y) > 2 && (
                  <line x1={x(l.week) + 4} y1={l.y} x2={lx} y2={l.labelY} style={{ stroke: color(l.series) }} />
                )}
                <rect x={lx} y={l.labelY - 7.5} width={width} height={15} rx={7.5} style={{ stroke: color(l.series) }} />
                <text
                  x={lx + 6}
                  y={l.labelY + 3.5}
                  data-testid="delta-label"
                  data-series={name(l.series)}
                  data-week={weeks[l.week]}
                >
                  {cost}
                  {l.pct !== null && <tspan className={deltaClass(l.pct)}>{delta}</tspan>}
                </text>
              </g>
            );
          })}
          {points.map((p) => (
            <circle
              key={`${p.series}-${p.week}`}
              cx={x(p.week)}
              cy={p.y}
              r={p === tip ? 7 : 4}
              className="weekly-point"
              style={{ fill: color(p.series) }}
              tabIndex={0}
              data-testid="time-series-point"
              data-series={name(p.series)}
              data-week={weeks[p.week]}
              aria-label={`${name(p.series)} ${weeks[p.week]}: ${formatUsd(p.value)}`}
              onMouseEnter={() => setHovered(p)}
              onMouseLeave={() => setHovered(null)}
              onFocus={() => setHovered(p)}
              onBlur={() => setHovered(null)}
            />
          ))}
        </svg>
        {tip && (
          <div
            role="tooltip"
            // The last weeks open the tooltip leftwards, so it stays inside the plot.
            className={tip.week >= WEEKS - 2 ? "weekly-tooltip weekly-tooltip-left" : "weekly-tooltip"}
            style={{ left: `${(x(tip.week) / WIDTH) * 100}%`, top: `${(tip.y / HEIGHT) * 100}%` }}
          >
            <strong style={{ color: color(tip.series) }}>{name(tip.series)}</strong>
            <div>Week of {weeks[tip.week]}</div>
            <strong>{formatUsd(tip.value)}</strong>
            {tip.pct !== null && (
              <div>
                <span className={deltaClass(tip.pct)}>{deltaText(tip.pct, 1, " ")}</span> vs prev week
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export function WeeklyCosts(_props: ViewProps) {
  // Keyed on the UTC day, so a render after midnight moves the window with it.
  const today = new Date().toISOString().slice(0, 10);
  const eightWeeks = useMemo(() => lastEightWeeks(new Date()), [today]);
  const params = { ...eightWeeks, project: undefined, model: undefined };
  const state = useStatsView<WeeklyCostPoint[]>("weekly-costs", params);
  const modelEffort = useStatsView<WeeklyModelEffortPoint[]>("weekly-model-effort-costs", params);
  const weeks = everyWeek(eightWeeks.from);

  return (
    <ViewFrame title="Weekly costs" description="Wednesday-to-Wednesday cost, last 8 weeks, every project and model.">
      <Panel title="Cost by phase" description="The phases and Other add up to Overall." state={state}>
        {(data) => (
          <WeeklyCostChart
            weeks={weeks}
            series={phaseSeries(weeks, data.rows)}
            ariaLabel="Weekly cost by phase over the last 8 weeks"
          />
        )}
      </Panel>
      <Panel
        title="Cost by model · effort"
        description="Each subagent dispatch by the model and effort that served it, plus the parent session per model."
        state={modelEffort}
      >
        {(data) => (
          <WeeklyCostChart
            weeks={weeks}
            series={modelEffortSeries(weeks, data.rows)}
            ariaLabel="Weekly cost by model and effort over the last 8 weeks"
            allShown
          />
        )}
      </Panel>
    </ViewFrame>
  );
}
