// Trend over time -- "Is a finished change getting cheaper as flow
// changes?" (decision trend-question-cost-per-change), built as the approved
// prototype draws it: whole-life cost per finished change, by the Wednesday
// week it finished.
//
// The daily panels, the model filter and the shared dashboard bar are gone
// (trend-drops-daily-views): App skips the bar here, and the view owns a
// KAN-1048 chip bar of Period (8w / 12w / 26w, whole Wednesday weeks ending
// with the current one -- trend-period-weeks), Command (exactly one always
// selected, default /flow, no "All" -- trend-command-chip-single) and
// Project. The window is computed when a Period chip is chosen and once at
// mount, never per render: a per-render `new Date()` would refetch forever.
//
// One request returns every finished change in the window; the chips filter
// it client-side. The tiles compare the window's latest half with its first
// (trend-headline-halves). Flow changes landed counts every command's
// changes finished in the `agents` project, whatever the Command and Project
// chips say (trend-flow-changes-agents-project). A week with no finished
// change is a gap in both lines and "—" in the table, never $0.
import { useLayoutEffect, useRef, useState } from "react";
import { flushSync } from "react-dom";
import type { TrendChange } from "../api";
import { Chip, ChipGroup } from "../components/Chip";
import { Panel } from "../components/Panel";
import { useStatsView } from "../hooks/useStatsView";
import { projectLabel } from "../lib/projectLabel";
import { weekWindow } from "../lib/weekWindow";
import type { ViewProps } from "../viewTypes";

const PERIOD_WEEKS = { "8w": 8, "12w": 12, "26w": 26 } as const;
type PeriodChip = keyof typeof PERIOD_WEEKS;

interface Filters {
  period: PeriodChip;
  command: string;
  project: string | null;
  from: Date;
  to: Date;
  weeks: string[];
}

// Computed only when a Period chip is chosen.
function windowFor(period: PeriodChip): Pick<Filters, "period" | "from" | "to" | "weeks"> {
  return { period, ...weekWindow(PERIOD_WEEKS[period], Date.now()) };
}

const defaults = (): Filters => ({ ...windowFor("12w"), command: "/flow", project: null });

const usd = (v: number) => (v >= 100 ? "$" + Math.round(v).toLocaleString("en-US") : "$" + v.toFixed(2));
const sum = (xs: number[]) => xs.reduce((a, x) => a + x, 0);
function median(xs: number[]): number | null {
  if (!xs.length) return null;
  const s = [...xs].sort((a, b) => a - b);
  const m = s.length >> 1;
  return s.length % 2 ? s[m] : (s[m - 1] + s[m]) / 2;
}
const mean = (xs: number[]) => (xs.length ? sum(xs) / xs.length : null);
const wkLabel = (w: string) => new Date(w + "T00:00:00Z").toLocaleDateString("en-US", { month: "short", day: "numeric", timeZone: "UTC" });

function Delta({ cur, prev }: { cur: number | null; prev: number | null }) {
  if (cur == null || prev == null || prev < 0.005) return <span className="lb-muted">—</span>;
  const d = cur - prev;
  const p = Math.round(Math.abs(d / prev) * 100);
  const s = d > 0 ? "+" : "−";
  return <span className={d > 0 ? "up" : "down"}>{`${s}${usd(Math.abs(d))} (${s}${p}%)`}</span>;
}

interface Point {
  w: string;
  n: number;
  med: number | null;
  mean: number | null;
  total: number;
  fc: number;
}

// Chart geometry, as the prototype draws it.
const W = 960,
  H = 280,
  L = 56,
  R = 16,
  T = 12,
  B = 34,
  FCH = 46;
const plotB = H - B - FCH - 8;

function Chart({ pts }: { pts: Point[] }) {
  const [hover, setHover] = useState<number | null>(null);
  const chartRef = useRef<HTMLDivElement>(null);
  const tipRef = useRef<HTMLDivElement>(null);

  const vals = pts.flatMap((p) => [p.med, p.mean]).filter((v): v is number => v != null);
  const raw = Math.max(1, ...vals) * 1.05;
  const mag = Math.pow(10, Math.floor(Math.log10(raw)));
  const max = [1, 1.2, 1.6, 2, 2.4, 3, 4, 5, 6, 8, 10].map((m) => m * mag).find((m) => m >= raw)!;
  const fcMax = Math.max(1, ...pts.map((p) => p.fc));
  const step = (W - L - R) / pts.length;
  const x = (i: number) => L + step * (i + 0.5);
  const y = (v: number) => T + (plotB - T) * (1 - v / max);
  const line = (k: "med" | "mean") => {
    let d = "";
    let pen = false;
    pts.forEach((p, i) => {
      const v = p[k];
      if (v == null) {
        pen = false;
        return;
      }
      d += (pen ? "L" : "M") + x(i).toFixed(1) + "," + y(v).toFixed(1);
      pen = true;
    });
    return d;
  };
  const ticks = [0, 0.25, 0.5, 0.75, 1].map((t) => max * t);
  const every = pts.length > 12 ? 2 : 1;
  const hp = hover == null ? null : pts[hover];

  // The tip sits beside the guide, flipped left when it would overflow.
  useLayoutEffect(() => {
    const chart = chartRef.current;
    const tip = tipRef.current;
    if (hover == null || !chart || !tip) return;
    const sx = chart.clientWidth / W;
    const left = x(hover) * sx + 12;
    tip.style.left = (left + tip.offsetWidth > chart.clientWidth ? x(hover) * sx - tip.offsetWidth - 12 : left) + "px";
    tip.style.top = "8px";
  });

  return (
    <div className="tr-chart" ref={chartRef}>
      <svg viewBox={`0 0 ${W} ${H}`} role="img" aria-label="Cost per finished change by week">
        <g className="grid">
          {ticks.map((t) => (
            <line key={t} x1={L} x2={W - R} y1={y(t)} y2={y(t)} />
          ))}
        </g>
        <g className="axis">
          {ticks.map((t) => (
            <text key={t} x={L - 6} y={y(t) + 4} textAnchor="end">
              ${Math.round(t)}
            </text>
          ))}
          {pts.map((p, i) =>
            i % every ? null : (
              <text key={p.w} x={x(i)} y={H - 8} textAnchor="middle">
                {wkLabel(p.w)}
              </text>
            ),
          )}
        </g>
        {pts.map((p, i) => {
          const bh = (FCH * p.fc) / fcMax;
          return <rect key={p.w} className="fc" x={x(i) - step * 0.3} y={H - B - bh} width={step * 0.6} height={bh} />;
        })}
        <line className="guide" y1={T} y2={H - B} x1={hover == null ? -10 : x(hover)} x2={hover == null ? -10 : x(hover)} />
        <path className="mean" d={line("mean")} />
        <path className="median" d={line("med")} />
        {pts.map((p, i) => (p.med == null ? null : <circle key={p.w} className="pt" cx={x(i)} cy={y(p.med)} r={3} />))}
        {pts.map((p, i) => (
          <rect
            key={p.w}
            className="hit"
            data-testid={"trend-week-" + p.w}
            x={L + step * i}
            y={T}
            width={step}
            height={H - B - T}
            onMouseEnter={() => flushSync(() => setHover(i))}
            onMouseLeave={() => flushSync(() => setHover(null))}
          />
        ))}
      </svg>
      <div className="tr-tip" ref={tipRef} style={hp ? { display: "block" } : undefined}>
        {hp && (
          <>
            <b>Week of {wkLabel(hp.w)}</b>
            {hp.n ? `${hp.n} changes · median ${usd(hp.med!)} · mean ${usd(hp.mean!)}` : "no finished changes"}
            <br />
            {hp.fc} flow changes landed
          </>
        )}
      </div>
    </div>
  );
}

export function Trend(_props: ViewProps) {
  const [f, setF] = useState<Filters>(defaults);
  const state = useStatsView<TrendChange[]>("trend", { from: f.from, to: f.to });
  const ready = state.status === "ready" && state.data.recorded && !state.data.unmeasured;

  const inWin = ready ? state.data.rows.filter((c) => f.weeks.includes(c.week)) : [];
  const cmds = [...new Set(inWin.map((c) => c.command))];
  const byCmd = inWin.filter((c) => c.command === f.command);
  const projCounts = new Map<string, number>();
  for (const c of byCmd) projCounts.set(c.projectKey, (projCounts.get(c.projectKey) ?? 0) + 1);
  const rows = byCmd.filter((c) => f.project === null || c.projectKey === f.project);
  const offDefaults = f.period !== "12w" || f.command !== "/flow" || f.project !== null;

  const pts: Point[] = f.weeks.map((w) => {
    const cs = rows.filter((c) => c.week === w).map((c) => c.costUsd);
    const fc = inWin.filter((c) => c.week === w && projectLabel(c.projectKey) === "agents").length;
    return { w, n: cs.length, med: median(cs), mean: mean(cs), total: sum(cs), fc };
  });
  const half = f.weeks.length / 2;
  const firstMed = median(rows.filter((c) => f.weeks.indexOf(c.week) < half).map((c) => c.costUsd));
  const lastMed = median(rows.filter((c) => f.weeks.indexOf(c.week) >= half).map((c) => c.costUsd));
  const fcTotal = sum(pts.map((p) => p.fc));

  return (
    <section aria-label="Trend over time" className="dashboard lb tr">
      <div>
        <h2>Trend over time</h2>
        <p className="lb-sub">Is a finished change getting cheaper as flow changes?</p>
      </div>
      <div className="lb-fbar" role="toolbar" aria-label="Filters">
        <ChipGroup name="Period">
          {(Object.keys(PERIOD_WEEKS) as PeriodChip[]).map((p) => (
            <Chip key={p} on={f.period === p} onClick={() => setF({ ...f, ...windowFor(p) })}>
              {p}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Command">
          {cmds.map((c) => (
            <Chip key={c} on={f.command === c} onClick={() => setF({ ...f, command: c, project: null })}>
              {c}
              <span className="lb-c">{inWin.filter((x) => x.command === c).length}</span>
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Project">
          <Chip on={f.project === null} onClick={() => setF({ ...f, project: null })}>
            All
          </Chip>
          {[...projCounts].map(([p, n]) => (
            <Chip key={p} on={f.project === p} onClick={() => setF({ ...f, project: p })} title={p}>
              {projectLabel(p)}
              <span className="lb-c">{n}</span>
            </Chip>
          ))}
        </ChipGroup>
        {offDefaults && (
          <button type="button" className="lb-fclear" onClick={() => setF(defaults())}>
            Clear filters
          </button>
        )}
      </div>
      {!ready ? (
        // Loading, error, not-recorded and unmeasured read as Panel renders them.
        <Panel title="Cost per finished change, by week" state={state}>
          {() => null}
        </Panel>
      ) : (
        <>
          <div className="tr-tiles">
            <div className="lb-tile static tr-tile">
              <span className="lb-lbl">Median per change</span>
              <span className="lb-val">{lastMed == null ? "—" : usd(lastMed)}</span>
              <span className="tr-note">
                last {half} weeks · <Delta cur={lastMed} prev={firstMed} /> vs the {half} weeks before
              </span>
            </div>
            <div className="lb-tile static tr-tile">
              <span className="lb-lbl">Changes finished</span>
              <span className="lb-val">{rows.length.toLocaleString("en-US")}</span>
              <span className="tr-note">{usd(sum(rows.map((r) => r.costUsd)))} spent on them</span>
            </div>
            <div className="lb-tile static tr-tile">
              <span className="lb-lbl">Flow changes landed</span>
              <span className="lb-val">{fcTotal}</span>
              <span className="tr-note">changes finished in agents, every command</span>
            </div>
          </div>
          <div className="lb-panel">
            <div className="lb-panel-h">
              <span>Cost per finished change, by week</span>
              <span className="lb-panel-note">{`${f.command}${f.project ? " · " + projectLabel(f.project) : ""}`}</span>
            </div>
            <div className="lb-panel-b">
              <div className="tr-legend">
                <span>
                  <i></i>Median
                </span>
                <span>
                  <i className="mean"></i>Mean
                </span>
                <span>
                  <i className="fc"></i>Flow changes landed
                </span>
              </div>
              <Chart pts={pts} />
            </div>
          </div>
          <div className="lb-panel">
            <div className="lb-panel-h">
              <span>By week</span>
              <span className="lb-panel-note">weeks start Wednesday 00:00 UTC</span>
            </div>
            <div className="lb-scroll">
              {!rows.length ? (
                <div className="tr-empty">No finished changes in this period.</div>
              ) : (
                <table className="lb-table">
                  <thead>
                    <tr>
                      <th>Week</th>
                      <th className="r">Changes</th>
                      <th className="r">Median</th>
                      <th className="r">vs previous week</th>
                      <th className="r">Mean</th>
                      <th className="r">Spend</th>
                      <th className="r">Flow changes</th>
                    </tr>
                  </thead>
                  <tbody>
                    {pts
                      .map((p, i) => (
                        <tr key={p.w}>
                          <td>
                            {wkLabel(p.w)}
                            {i === pts.length - 1 && (
                              <>
                                {" "}
                                <span className="lb-muted">so far</span>
                              </>
                            )}
                          </td>
                          <td className="r">{p.n}</td>
                          <td className="r">{p.med == null ? "—" : usd(p.med)}</td>
                          <td className="r">
                            <Delta cur={p.med} prev={i > 0 ? pts[i - 1].med : null} />
                          </td>
                          <td className="r">{p.mean == null ? "—" : usd(p.mean)}</td>
                          <td className="r">{p.n ? usd(p.total) : "—"}</td>
                          <td className="r">{p.fc}</td>
                        </tr>
                      ))
                      .reverse()}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </>
      )}
    </section>
  );
}
