// Weekly costs -- where does flow spend go, week by week? Weeks run
// Wednesday 00:00 UTC to the next Wednesday. One chart covers the last 8
// weeks, the current partial week included, over every project and model:
// the view takes no filter (KAN-1038), so App renders no filter bar for it.
// Each phase is a toggleable line; every point carries its change from the
// same series' previous week. A category with no costed run that week is a
// gap in its line, never $0, and the week after a gap carries no change.
import { useMemo, useState } from "react";
import { Panel } from "../components/Panel";
import { ViewFrame } from "../components/ViewFrame";
import type { WeeklyCostPoint } from "../api";
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
];

/** The last 8 Wednesday-to-Wednesday weeks: from this week's UTC
 * Wednesday 00:00 minus 7 weeks, to now. */
export function lastEightWeeks(now: Date): { from: Date; to: Date } {
  const midnight = Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate());
  const sinceWednesday = (now.getUTCDay() + 4) % 7;
  return { from: new Date(midnight - (sinceWednesday + (WEEKS - 1) * 7) * DAY_MS), to: now };
}

/** Every week of `from`'s 8-week window, in order: the response's row where
 * it carries one, an all-null row where no costed run fell in that week. */
function everyWeek(from: Date, rows: WeeklyCostPoint[]): WeeklyCostPoint[] {
  const byWeek = new Map(rows.map((r) => [r.week, r]));
  return Array.from({ length: WEEKS }, (_, i) => {
    const week = new Date(from.getTime() + i * 7 * DAY_MS).toISOString().slice(0, 10);
    return (
      byWeek.get(week) ?? {
        week,
        overallUsd: null,
        planningUsd: null,
        implementationUsd: null,
        reviewUsd: null,
        selfReviewUsd: null,
      }
    );
  });
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
const HEIGHT = 440;
const LEFT = 64;
const RIGHT = 60;
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
  pct: number;
  labelY: number;
}

/** Per week column, the delta labels sorted by y and pushed apart to
 * LABEL_GAP, then shifted up together where the stack overruns the plot. */
function stackLabels(points: Point[]): Label[] {
  const out: Label[] = [];
  for (let week = 1; week < WEEKS; week++) {
    const column = points
      .filter((p): p is Point & { pct: number } => p.week === week && p.pct !== null)
      .sort((a, b) => a.y - b.y);
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

function WeeklyCostChart({ rows }: { rows: WeeklyCostPoint[] }) {
  const [shown, setShown] = useState(() => categories.map(() => true));
  const [hovered, setHovered] = useState<{ series: number; week: number } | null>(null);

  const visible = categories
    .flatMap((c, n) => (shown[n] ? rows.map((r) => r[c.key]) : []))
    .filter((v): v is number => v !== null);
  const step = tickStep(Math.max(0, ...visible));
  const x = (week: number) => LEFT + (PLOT_W * week) / (WEEKS - 1);
  const y = (value: number) => TOP + PLOT_H - (PLOT_H * value) / (step * 5);

  const lines = categories.map(({ key }, series) =>
    rows.map((r, week): Point | null => {
      const value = r[key];
      if (value === null) return null;
      return { series, week, value, y: y(value), pct: week ? change(value, rows[week - 1][key]) : null };
    }),
  );
  const drawn = lines.filter((_, n) => shown[n]);
  const points = drawn.flat().filter((p): p is Point => p !== null);
  const labels = stackLabels(points);
  const tip = hovered && points.find((p) => p.series === hovered.series && p.week === hovered.week);
  const color = (series: number) => `var(--series-${series + 1})`;

  return (
    <div className="weekly-chart">
      <div className="weekly-toggles">
        {categories.map(({ name }, n) => (
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
        <svg viewBox={`0 0 ${WIDTH} ${HEIGHT}`} role="img" aria-label="Weekly cost by phase over the last 8 weeks">
          <g className="weekly-axis">
            {Array.from({ length: 6 }, (_, k) => (
              <g key={k}>
                <line x1={LEFT} x2={WIDTH - RIGHT} y1={y(k * step)} y2={y(k * step)} />
                <text x={LEFT - 8} y={y(k * step) + 4} textAnchor="end" data-testid="cost-tick">
                  {formatUsd(k * step)}
                </text>
              </g>
            ))}
            {rows.map((r, week) => (
              <text key={r.week} x={x(week)} y={HEIGHT - BOTTOM + 18} textAnchor="middle" data-testid="week-tick">
                {r.week}
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
            return d && <path key={n} d={d} fill="none" className="weekly-line" style={{ stroke: color(line.find((p) => p)!.series) }} />;
          })}
          {labels.map((l) => {
            const lx = x(l.week) + 8;
            return (
              <g key={`${l.series}-${l.week}`} className="weekly-delta">
                {Math.abs(l.labelY - l.y) > 2 && (
                  <line x1={x(l.week) + 4} y1={l.y} x2={lx} y2={l.labelY} style={{ stroke: color(l.series) }} />
                )}
                <rect x={lx} y={l.labelY - 7.5} width={44} height={15} rx={7.5} style={{ stroke: color(l.series) }} />
                <text
                  x={lx + 22}
                  y={l.labelY + 3.5}
                  textAnchor="middle"
                  className={deltaClass(l.pct)}
                  data-testid="delta-label"
                  data-series={categories[l.series].name}
                  data-week={rows[l.week].week}
                >
                  {deltaText(l.pct, 0, "")}
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
              data-series={categories[p.series].name}
              data-week={rows[p.week].week}
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
            <strong style={{ color: color(tip.series) }}>{categories[tip.series].name}</strong>
            <div>Week of {rows[tip.week].week}</div>
            <strong>{formatUsd(tip.value)}</strong>
            {tip.pct !== null && (
              <div>
                <span className={deltaClass(tip.pct)}>{deltaText(tip.pct, 1, " ")}</span> vs previous week
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export function WeeklyCosts(_props: ViewProps) {
  const eightWeeks = useMemo(() => lastEightWeeks(new Date()), []);
  const state = useStatsView<WeeklyCostPoint[]>("weekly-costs", { ...eightWeeks, project: undefined, model: undefined });

  return (
    <ViewFrame title="Weekly costs" description="Wednesday-to-Wednesday cost by pipeline phase, every project and model.">
      <Panel title="Last 8 weeks" description="Overall includes stages outside the four phases." state={state}>
        {(data) => <WeeklyCostChart rows={everyWeek(eightWeeks.from, data.rows)} />}
      </Panel>
    </ViewFrame>
  );
}
