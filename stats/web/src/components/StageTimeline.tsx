// StageTimeline draws each stage run as a bar positioned by startedAt and
// endedAt across the change's own span, one lane per stage in first-seen
// order (design.md's timeline-lane-per-stage): bars filled by outcome, an
// hour-tick axis, a dashed now marker while a run is open, and a hover
// tooltip. It is presentation over the same rows the stage-run table has: it
// issues no request of its own, so the timeline and the table can never
// disagree.
//
// An open run (no endedAt) draws to now and is hatched rather than given a
// fabricated end -- the same absence-is-never-zero rule the table applies to
// its own metrics, applied here to "when did this run finish".
import { useEffect, useRef, useState } from "react";
import { flushSync } from "react-dom";
import type { StageRunDTO } from "../api";
import { formatClock, formatSpan, formatUsd } from "../format";
import { readCostUsd } from "../metrics";
import { stageLabel } from "../lib/stageLabel";

const L = 210;
const RH = 18;
const TOP = 22;
const HOUR = 3600e3;
const TICK_STEPS = [1 / 12, 1 / 4, 1 / 2, 1, 2, 3, 6, 12, 24].map((h) => h * HOUR);
const MAX_TICKS = 12;
// Bars in one lane closer than this many pixels are drawn joined, so a burst
// of short runs reads as one stretch rather than a row of slivers.
const JOIN_GAP = 8;

/** A run's outcome as the run view colours it: open while running, failed, or completed. */
export function outcomeClass(run: StageRunDTO): "open" | "failed" | "completed" {
  if (!run.endedAt) return "open";
  return run.outcome === "failed" || run.outcome === "aborted" ? "failed" : "completed";
}

const FILL = { completed: "var(--status-ok)", failed: "var(--status-critical)", open: "url(#lb-hatch)" };

/** Draws each stage run as a bar on its stage's lane; `upcoming` adds a muted lane, with no bar, for every stage the change has still to run. */
export function StageTimeline({
  stageRuns,
  sessionOf,
  upcoming = [],
}: {
  stageRuns: StageRunDTO[];
  sessionOf: (run: StageRunDTO) => string;
  upcoming?: string[];
}) {
  const [tip, setTip] = useState<{ run: StageRunDTO; x: number; y: number } | null>(null);
  // Drawn at the container's real pixel width, so text and bar heights keep
  // their size whatever the page width.
  const ref = useRef<SVGSVGElement>(null);
  const [W, setW] = useState(900);
  useEffect(() => {
    const measure = () => ref.current && ref.current.clientWidth > 0 && setW(ref.current.clientWidth);
    measure();
    window.addEventListener("resize", measure);
    return () => window.removeEventListener("resize", measure);
  }, []);
  if (stageRuns.length === 0) return null;

  const now = Date.now();
  const end = (r: StageRunDTO) => (r.endedAt ? Date.parse(r.endedAt) : now);
  const anyOpen = stageRuns.some((r) => !r.endedAt);
  const t0 = Math.min(...stageRuns.map((r) => Date.parse(r.startedAt)));
  const t1 = Math.max(t0 + 1, anyOpen ? now : Math.max(...stageRuns.map(end)));
  const x = (t: number) => L + ((t - t0) / (t1 - t0)) * (W - L - 10);

  const lanes = [...new Set(stageRuns.map((r) => r.stage))];
  const pending = upcoming.filter((s) => !lanes.includes(s));
  const H = TOP + (lanes.length + pending.length) * RH + 6;
  // ponytail: hour ticks widen to 2/3/6/12/24h so a days-long change keeps at most 12 labels.
  const step = TICK_STEPS.find((s) => (t1 - t0) / s <= MAX_TICKS) ?? TICK_STEPS[TICK_STEPS.length - 1];
  const ticks: number[] = [];
  for (let t = Math.ceil(t0 / step) * step; t <= t1; t += step) ticks.push(t);

  return (
    <>
      <svg ref={ref} viewBox={`0 0 ${W} ${H}`} width="100%" height={H} className="lb-timeline" role="img" aria-label="Stage runs over time">
        <defs>
          <pattern id="lb-hatch" width="6" height="6" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
            <rect width="6" height="6" fill="var(--st-progress)" opacity=".35" />
            <line x1="0" y1="0" x2="0" y2="6" stroke="var(--st-progress)" strokeWidth="3" />
          </pattern>
        </defs>
        {ticks.map((t) => (
          <g key={t}>
            <line className="lb-grid" x1={x(t)} x2={x(t)} y1={TOP - 6} y2={H} />
            <text x={x(t)} y={11} textAnchor="middle">
              {formatClock(t)}
            </text>
          </g>
        ))}
        {lanes.map((s, i) => (
          <text key={s} className="lb-lane" x={0} y={TOP + i * RH + 12}>
            <title>{s}</title>
            {stageLabel(s)}
          </text>
        ))}
        {pending.map((s, i) => (
          <text key={s} className="lb-lane upcoming" data-testid="upcoming-lane" x={0} y={TOP + (lanes.length + i) * RH + 12}>
            <title>{`${s} — not run yet`}</title>
            {stageLabel(s)}
          </text>
        ))}
        <polyline
          className="lb-path"
          fill="none"
          points={[...stageRuns]
            .sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt))
            .map((r) => {
              const y = TOP + lanes.indexOf(r.stage) * RH + 8;
              return `${x(Date.parse(r.startedAt))},${y} ${Math.max(x(Date.parse(r.startedAt)) + 6, x(end(r)))},${y}`;
            })
            .join(" ")}
        />
        {stageRuns.map((r) => {
          const x0 = x(Date.parse(r.startedAt));
          const next = Math.min(
            ...stageRuns.filter((o) => o.stage === r.stage && Date.parse(o.startedAt) > Date.parse(r.startedAt)).map((o) => x(Date.parse(o.startedAt))),
          );
          const x1 = Math.max(x0 + 6, x(end(r)));
          return (
            <rect
              key={r.stageRunId}
              className="lb-bar"
              data-testid="timeline-bar"
              x={x0}
              y={TOP + lanes.indexOf(r.stage) * RH + 3}
              width={(next > x1 && next - x1 <= JOIN_GAP ? next : x1) - x0}
              height={10}
              rx={2}
              fill={FILL[outcomeClass(r)]}
              onMouseMove={(e) => setTip({ run: r, x: e.clientX, y: e.clientY })}
              // Hidden in the same paint the pointer leaves, as the prototype's
              // handler hides it; a deferred render left it one frame behind.
              onMouseLeave={() => flushSync(() => setTip(null))}
            />
          );
        })}
        {anyOpen && (
          <>
            <line className="lb-now" x1={x(now)} x2={x(now)} y1={TOP - 6} y2={H} />
            <text x={x(now) - 4} y={TOP - 9} textAnchor="end" className="lb-now-label">
              now
            </text>
          </>
        )}
      </svg>
      {tip && (
        <div className="lb-tip" role="tooltip" style={{ left: Math.min(tip.x + 12, window.innerWidth - 290), top: tip.y + 12 }}>
          <b>{stageLabel(tip.run.stage)}</b> <span className="lb-muted">{tip.run.stage}</span>
          <br />
          {sessionOf(tip.run)}
          <br />
          {outcomeClass(tip.run) === "open" ? "running" : (tip.run.outcome ?? "completed")} ·{" "}
          {formatSpan(end(tip.run) - Date.parse(tip.run.startedAt))}
          {readCostUsd(tip.run.metrics) !== null && ` · ${formatUsd(readCostUsd(tip.run.metrics)!)}`}
        </div>
      )}
    </>
  );
}
