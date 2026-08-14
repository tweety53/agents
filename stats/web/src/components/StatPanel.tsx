// StatPanel renders one large number with a label and an optional
// sparkline -- the "stat panels showing one large number" half of
// proposal.md's Fix round ("the interface adopts Grafana's dashboard
// model"). It is the single most dangerous place in this app to coerce
// an absence: rendered on its own, with no surrounding row or column to
// contradict it, a fabricated 0 in place of "not measured" reads as a
// real, confident answer. It therefore delegates the value itself to
// Unavailable.tsx -- the one place this app already turns a `null` into
// "unavailable" rather than a zero -- instead of re-deriving that rule
// here (this app's own Single Source of Truth principle;
// engineering-principles.md). The three cases `0`, `null` and
// `undefined` are tested separately (Panel.test.tsx) precisely because
// they must not collapse into one another.
import { Unavailable } from "./Unavailable";

export interface StatPanelSparklinePoint {
  day: string;
  value: number | null;
}

export interface StatPanelProps {
  label: string;
  value: number | null | undefined;
  format?: (value: number) => string;
  unit?: string;
  sparkline?: StatPanelSparklinePoint[];
}

const SPARKLINE_WIDTH = 120;
const SPARKLINE_HEIGHT = 28;

/**
 * A minimal inline-SVG sparkline -- deliberately not shared with
 * TimeSeriesPanel.tsx, which draws a full chart with axes and a legible
 * point per day. A sparkline has neither: it is a glance-sized trend
 * line with no scale, and giving it TimeSeriesPanel's own axis/label
 * machinery would be exactly the kind of forced-shared abstraction WET
 * warns against for two shapes that only coincidentally both draw lines
 * (engineering-principles.md).
 */
function sparklinePath(points: StatPanelSparklinePoint[]): string | null {
  const measured = points
    .map((p, i) => (p.value === null || p.value === undefined ? null : { i, value: p.value }))
    .filter((p): p is { i: number; value: number } => p !== null);
  if (measured.length < 2) return null;

  const values = measured.map((p) => p.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;
  const lastIndex = points.length - 1 || 1;

  return measured
    .map(({ i, value }) => {
      const x = (i / lastIndex) * SPARKLINE_WIDTH;
      const y = SPARKLINE_HEIGHT - ((value - min) / range) * SPARKLINE_HEIGHT;
      return `${x},${y}`;
    })
    .join(" ");
}

export function StatPanel({ label, value, format, unit, sparkline }: StatPanelProps) {
  const path = sparkline ? sparklinePath(sparkline) : null;

  // The label is not rendered here as its own heading -- the Panel (or
  // RunPanel, in RunDetail.tsx) already wraps every StatPanel with its
  // own <h3>, and the two are always the same string at every call
  // site (task 20). Rendering it twice was task 27's defect 2. The
  // panel heading is the better owner: every other panel type gets its
  // title that way, so a stat panel carrying its own would be the one
  // inconsistent case (Single Source of Truth,
  // engineering-principles.md). `label` is still taken as a prop and
  // still used for the sparkline's own aria-label below.
  return (
    <div className="stat-panel">
      <div className="stat-panel-value">
        <Unavailable value={value} format={format} unit={unit} />
      </div>
      {path && (
        <svg
          className="stat-panel-sparkline"
          role="img"
          aria-label={`${label} trend`}
          viewBox={`0 0 ${SPARKLINE_WIDTH} ${SPARKLINE_HEIGHT}`}
          width={SPARKLINE_WIDTH}
          height={SPARKLINE_HEIGHT}
        >
          <polyline points={path} fill="none" className="time-series-line" data-testid="stat-panel-sparkline-line" />
        </svg>
      )}
    </div>
  );
}
