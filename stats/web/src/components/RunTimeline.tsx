// RunTimeline draws one run's lanes on the prototype's SVG geometry
// (KAN-1132 view B): stage lanes first, then a "dispatches" label and the
// category lanes nested under it, parent commands, and idle / parent as a
// hatched lane. Each bar's <title> states its length and offsets from the
// run's start.
import type { RunTimings } from "../api";

/** Formats milliseconds as the prototype's `Nh MMm`. */
export function hm(ms: number): string {
  const m = Math.round(ms / 60_000);
  return `${Math.floor(m / 60)}h ${String(m % 60).padStart(2, "0")}m`;
}

const W = 1100;
const L = 210;
const RH = 18;

function fill(kind: string, cls: string, cat: number): string {
  if (cls === "open") return "var(--st-progress)";
  if (cls === "failed" || cls === "aborted") return "var(--status-critical)";
  if (kind === "stage") return "var(--status-ok)";
  if (kind === "parent") return "var(--text-secondary)";
  if (kind === "idle") return "url(#hatch)";
  return `var(--series-${(cat % 7) + 1})`;
}

export function RunTimeline({ run }: { run: RunTimings }) {
  const from = Date.parse(run.from);
  const span = Math.max(1, Date.parse(run.to) - from);
  const hours = span / 3_600_000;
  const stages = run.lanes.filter((l) => l.kind === "stage");
  const rest = run.lanes.filter((l) => l.kind !== "stage");
  const H = (run.lanes.length + 1) * (RH + 4) + 50;
  const x = (ms: number) => L + ((W - L - 10) * (ms - from)) / span;
  const grid = [];
  for (let h = 0; h <= hours; h += 6) grid.push(h);
  let y = 24;
  let cat = 0;
  const lane = (l: RunTimings["lanes"][number], indent: boolean) => {
    const top = y;
    y += RH + 4;
    const c = l.kind === "category" ? cat++ : 0;
    return (
      <g key={`${l.kind}-${l.label}`}>
        <text x={indent ? 20 : 4} y={top + 13}>
          {l.label}
        </text>
        {l.intervals.map((iv, i) => {
          const a = Date.parse(iv.from);
          const b = Date.parse(iv.to);
          return (
            <rect key={i} className="bar" x={x(a)} y={top} width={Math.max(2, x(b) - x(a))} height={RH} rx={l.kind === "idle" ? undefined : 2} fill={fill(l.kind, iv.class, c)}>
              <title>{`${l.label}: ${hm(b - a)} (+${hm(a - from)} → +${hm(b - from)})`}</title>
            </rect>
          );
        })}
      </g>
    );
  };
  const stageLanes = stages.map((l) => lane(l, false));
  const dispatchY = y;
  y += 16;
  const restLanes = rest.map((l) => lane(l, l.kind !== "idle"));
  return (
    <svg width={W} height={H} role="img" aria-label="Run timeline">
      {grid.map((h) => (
        <g key={h}>
          <line className="lb-grid" x1={x(from + h * 3_600_000)} x2={x(from + h * 3_600_000)} y1={18} y2={H - 6} />
          <text x={x(from + h * 3_600_000) + 3} y={12}>{`+${h}h`}</text>
        </g>
      ))}
      {stageLanes}
      <text x={4} y={dispatchY + 10} style={{ fontWeight: 600 }}>
        dispatches
      </text>
      {restLanes}
      <defs>
        <pattern id="hatch" width={5} height={5} patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
          <rect width={2} height={5} fill="var(--text-muted)" />
        </pattern>
      </defs>
    </svg>
  );
}
