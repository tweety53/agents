// TimeSeriesPanel renders a line chart from `{ day, value }` points as
// inline SVG (task 19, step 3). No charting dependency is added: a single
// line with axes is well inside what SVG does directly, and pulling in a
// library for it is the "config switch with one implementation" KISS
// violation engineering-principles.md names -- this app already draws
// two other charts this way (Trend.tsx's bar chart, StageTimeline.tsx's
// duration bars), so a third inline-SVG chart is the established pattern,
// not a new one.
//
// A point whose value is `null` or `undefined` breaks the line rather
// than being coerced to 0 or interpolated across -- the same
// absence-is-never-zero rule this app's metrics already follow, applied
// here to "was this day measured at all".
//
// `series`, when given, draws one line per named point list over a shared
// x axis (the union of days, in order) and a shared y range, each in its
// own `--series-<n>` stroke, with a legend of the names.
import type { ReactElement } from "react";

export interface TimeSeriesPoint {
  day: string;
  value: number | null | undefined;
}

export interface TimeSeries {
  name: string;
  points: TimeSeriesPoint[];
}

export interface TimeSeriesPanelProps {
  points?: TimeSeriesPoint[];
  series?: TimeSeries[];
  format?: (value: number) => string;
}

const WIDTH = 480;
const HEIGHT = 160;
const PADDING = 32;

interface Measured {
  index: number;
  point: TimeSeriesPoint;
  value: number;
}

/** Splits one line's measured points into runs of consecutive days, so a
 * day the line does not measure breaks it instead of being drawn across. */
function segments(line: Measured[]): Measured[][] {
  const out: Measured[][] = [];
  for (const m of line) {
    const last = out[out.length - 1];
    if (last && last[last.length - 1].index === m.index - 1) last.push(m);
    else out.push([m]);
  }
  return out;
}

function measuredPoints(points: TimeSeriesPoint[], days: string[]): Measured[] {
  return points
    .map((point) =>
      point.value === null || point.value === undefined ? null : { index: days.indexOf(point.day), point, value: point.value },
    )
    .filter((m): m is Measured => m !== null);
}

export function TimeSeriesPanel({ points, series, format }: TimeSeriesPanelProps) {
  const lines = series ?? [{ name: "", points: points ?? [] }];
  const days = [...new Set(lines.flatMap((s) => s.points.map((p) => p.day)))].sort();
  const measuredLines = lines.map((s) => measuredPoints(s.points, days));
  const measured = measuredLines.flat();

  if (measured.length === 0) {
    return (
      <p className="not-recorded" data-testid="time-series-empty">
        No measured points in this period.
      </p>
    );
  }

  const plotWidth = WIDTH - PADDING * 2;
  const plotHeight = HEIGHT - PADDING * 2;
  const maxValue = Math.max(0, ...measured.map((m) => m.value));
  const minValue = Math.min(0, ...measured.map((m) => m.value));
  const range = maxValue - minValue || 1;
  const lastIndex = Math.max(1, days.length - 1);

  function x(index: number): number {
    return PADDING + (index / lastIndex) * plotWidth;
  }
  function y(value: number): number {
    return PADDING + plotHeight - ((value - minValue) / range) * plotHeight;
  }

  const zeroY = y(0);
  // The single-series path keeps the stylesheet's own stroke; only a
  // multi-series chart colours each line inline.
  const stroke = (n: number) => (series ? `var(--series-${n + 1})` : undefined);

  const drawn: ReactElement[] = measuredLines.flatMap((line, n) => [
    ...segments(line).map((segment) => (
      <polyline
        key={`line-${n}-${segment[0].index}`}
        points={segment.map((m) => `${x(m.index)},${y(m.value)}`).join(" ")}
        fill="none"
        className="time-series-line"
        style={{ stroke: stroke(n) }}
        data-series={lines[n].name || undefined}
        data-testid="time-series-line"
      />
    )),
    ...line.map((m) => (
      <circle
        key={`${n}-${m.point.day}`}
        cx={x(m.index)}
        cy={y(m.value)}
        r={2.5}
        className="time-series-point"
        style={{ fill: stroke(n) }}
        data-series={lines[n].name || undefined}
        data-testid="time-series-point"
      >
        <title>{`${lines[n].name ? `${lines[n].name} ` : ""}${m.point.day}: ${format ? format(m.value) : String(m.value)}`}</title>
      </circle>
    )),
  ]);

  const chart = (
    <svg
      className="time-series-panel"
      role="img"
      aria-label="Trend over the selected period"
      viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
      width={WIDTH}
      height={HEIGHT}
    >
      <line x1={PADDING} y1={zeroY} x2={WIDTH - PADDING} y2={zeroY} className="time-series-axis" />
      <line x1={PADDING} y1={PADDING} x2={PADDING} y2={HEIGHT - PADDING} className="time-series-axis" />
      {drawn}
    </svg>
  );

  if (!series) {
    return chart;
  }
  return (
    <div>
      {chart}
      <ul
        className="time-series-legend"
        data-testid="time-series-legend"
        style={{ display: "flex", gap: "var(--space-4)", listStyle: "none", margin: 0, padding: 0 }}
      >
        {series.map((s, n) => (
          <li key={s.name} style={{ color: stroke(n) }}>
            {s.name}
          </li>
        ))}
      </ul>
    </div>
  );
}
