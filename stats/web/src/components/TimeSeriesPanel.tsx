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
import type { ReactElement } from "react";

export interface TimeSeriesPoint {
  day: string;
  value: number | null | undefined;
}

export interface TimeSeriesPanelProps {
  points: TimeSeriesPoint[];
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

function measuredPoints(points: TimeSeriesPoint[]): Measured[] {
  return points
    .map((point, index) => (point.value === null || point.value === undefined ? null : { index, point, value: point.value }))
    .filter((m): m is Measured => m !== null);
}

export function TimeSeriesPanel({ points, format }: TimeSeriesPanelProps) {
  const measured = measuredPoints(points);

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
  const lastIndex = Math.max(1, points.length - 1);

  function x(index: number): number {
    return PADDING + (index / lastIndex) * plotWidth;
  }
  function y(value: number): number {
    return PADDING + plotHeight - ((value - minValue) / range) * plotHeight;
  }

  const zeroY = y(0);
  const linePoints = measured.map((m) => `${x(m.index)},${y(m.value)}`).join(" ");

  const circles: ReactElement[] = measured.map((m) => (
    <circle
      key={m.point.day}
      cx={x(m.index)}
      cy={y(m.value)}
      r={2.5}
      className="time-series-point"
      data-testid="time-series-point"
    >
      <title>{`${m.point.day}: ${format ? format(m.value) : String(m.value)}`}</title>
    </circle>
  ));

  return (
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
      <polyline points={linePoints} fill="none" className="time-series-line" data-testid="time-series-line" />
      {circles}
    </svg>
  );
}
