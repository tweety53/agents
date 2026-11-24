// Daily costs -- "which day cost how much": Weekly costs' categories per
// 24-hour day, the days starting at the client's local midnight. The Range
// control mirrors Reviewer findings': 7 or 14 days, stepped earlier and
// later by its own length, never past now. Below the days, each weekday's mean
// from one daily-costs fetch over the whole recorded span, so it does not move
// with Range.
import { useState } from "react";
import type { WeeklyCostPoint } from "../api";
import { formatUsd } from "../format";
import { useStatsView } from "../hooks/useStatsView";

const RANGES = [7, 14] as const;
const CATEGORIES: { key: keyof Omit<WeeklyCostPoint, "week">; name: string }[] = [
  { key: "overallUsd", name: "Overall" },
  { key: "planningUsd", name: "Planning" },
  { key: "implementationUsd", name: "Implementation" },
  { key: "reviewUsd", name: "Review" },
  { key: "selfReviewUsd", name: "Self-review" },
  { key: "visualVerifyUsd", name: "Visual verify" },
  { key: "otherUsd", name: "Other" },
];
const WEEKDAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
// ponytail: a fixed start before flow recorded anything; leading empty days are trimmed below.
const ALL_TIME_FROM = new Date(2026, 0, 1);

const localDay = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate());
const dayOffset = (n: number) => {
  const t = localDay(new Date());
  t.setDate(t.getDate() + n);
  return t;
};
const weekday = (iso: string) => WEEKDAYS[(new Date(iso).getDay() + 6) % 7];
const label = (iso: string) => {
  const d = new Date(iso);
  return `${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
};
const usd = (v: number | null | undefined) => formatUsd(v ?? 0);

type Bar = { key: string; label: string; testid: string; weekday?: string; total: number | null; segs: { name: string; v: number; color: string }[] };
const CW = 1140, CH = 220, L = 60, T = 24, B = 28, PH = CH - T - B, PW = CW - L - 10;
// The smallest 1, 2 or 5 times a power of ten that holds max, so the plot is scaled to the tallest bar.
const niceTop = (max: number) => {
  const m = 10 ** Math.floor(Math.log10(max || 1));
  return ([1, 2, 5, 10].find((f) => f * m >= max) ?? 10) * m;
};

// Hand-drawn SVG bars: stacked segments, the value above each bar, two gridlines labelled in $.
function BarChart({ bars, aria }: { bars: Bar[]; aria: string }) {
  const top = niceTop(Math.max(0, ...bars.map((b) => b.segs.reduce((a, s) => a + s.v, 0))));
  const y = (v: number) => T + PH - (PH * v) / top;
  const bw = PW / bars.length;
  return (
    <div className="wc-plot wc-chart lb-panel">
      <svg viewBox={`0 0 ${CW} ${CH}`} width={CW} height={CH} role="img" aria-label={aria}>
        <g className="grid">
          {[top / 2, top].map((v) => (
            <g key={v} data-testid="dc-grid">
              <line x1={L} x2={L + PW} y1={y(v)} y2={y(v)} />
              <text x={L - 6} y={y(v) + 4} textAnchor="end">
                {formatUsd(v)}
              </text>
            </g>
          ))}
        </g>
        {bars.map((b, i) => {
          const x = L + i * bw + bw * 0.18;
          const w = bw * 0.64;
          let acc = 0;
          const sum = b.segs.reduce((a, s) => a + s.v, 0);
          return (
            <g key={b.key} data-testid={b.testid} data-weekday={b.weekday}>
              {b.segs.map((s) => {
                if (!s.v) return null;
                const t = y(acc + s.v);
                const h = y(acc) - t;
                acc += s.v;
                return (
                  <rect key={s.name} data-testid="dc-seg" x={x} width={w} y={t} height={h} fill={s.color}>
                    <title>{`${s.name}: ${formatUsd(s.v)}`}</title>
                  </rect>
                );
              })}
              <text x={x + w / 2} y={y(sum) - 5} textAnchor="middle" className="wc-total">
                {b.total === null ? "—" : formatUsd(b.total)}
              </text>
              <text x={x + w / 2} y={CH - B + 16} textAnchor="middle">
                {b.label}
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
}

export function DailyCosts() {
  // end is when the view opened, as in Reviewer findings.
  const [end] = useState(() => Date.now());
  const [days, setDays] = useState<(typeof RANGES)[number]>(7);
  const [back, setBack] = useState(0);
  const to = back === 0 ? new Date(end) : dayOffset(1 - back * days);
  const from = dayOffset(1 - back * days - days);
  const state = useStatsView<WeeklyCostPoint[]>("daily-costs", { from, to });
  const all = useStatsView<WeeklyCostPoint[]>("daily-costs", { from: ALL_TIME_FROM, to: new Date(end) });

  const rows = state.status === "ready" ? state.data.rows : null;
  const span = all.status === "ready" ? all.data.rows : null;
  const first = span ? span.findIndex((r) => r.overallUsd) : -1;
  const recorded = span && first >= 0 ? span.slice(first) : [];
  const meanOf = (wd: string) => {
    const mine = recorded.filter((r) => weekday(r.week) === wd);
    return mine.length ? mine.reduce((a, r) => a + (r.overallUsd ?? 0), 0) / mine.length : null;
  };
  const mean = (wd: string) => {
    const m = meanOf(wd);
    return m === null ? "—" : usd(m);
  };
  const parts = CATEGORIES.slice(1);
  const dayBars: Bar[] = (rows ?? []).map((r) => ({
    key: r.week,
    label: label(r.week),
    testid: "dc-bar",
    total: r.overallUsd ?? 0,
    segs: parts.map(({ key, name }, i) => ({ name, v: r[key] ?? 0, color: `var(--series-${i + 1})` })),
  }));
  const weekdayBars: Bar[] = WEEKDAYS.map((w) => {
    const m = meanOf(w);
    return { key: w, label: w, testid: "dc-wbar", weekday: w, total: m, segs: m ? [{ name: w, v: m, color: "var(--series-1)" }] : [] };
  });

  return (
    <section aria-label="Daily costs" className="dashboard lb">
      <div>
        <h2>Daily costs</h2>
        <p className="lb-sub">Which day cost how much?</p>
      </div>
      <div className="dynamics-bar" role="group" aria-label="Range">
        <span className="dynamics-lbl">Range</span>
        {RANGES.map((d) => (
          <button
            key={d}
            type="button"
            aria-pressed={days === d}
            onClick={() => {
              setDays(d);
              setBack(0);
            }}
          >
            {d} days
          </button>
        ))}
        <button type="button" aria-label="Earlier" onClick={() => setBack((b) => b + 1)}>
          ‹
        </button>
        <span className="dynamics-sub dynamics-inline">{`${label(from.toISOString())} – ${label(to.toISOString())}`}</span>
        <button type="button" aria-label="Later" disabled={back === 0} onClick={() => setBack((b) => b - 1)}>
          ›
        </button>
      </div>
      {state.status === "loading" && <p role="status">Loading…</p>}
      {state.status === "error" && <p role="alert">Failed to load: {state.message}</p>}
      {rows && (
        <BarChart
          bars={dayBars}
          aria={`Daily costs, ${label(rows[0]?.week ?? from.toISOString())} to ${label(rows.at(-1)?.week ?? to.toISOString())}, stacked by ${parts.map((c) => c.name).join(", ")}: ${dayBars.map((b) => `${b.label} ${usd(b.total)}`).join(", ")}`}
        />
      )}
      {rows && (
        <div className="lb-panel">
          <div className="lb-scroll">
            <table className="lb-table">
              <thead>
                <tr>
                  <th />
                  {rows.map((r) => (
                    <th key={r.week} scope="col" data-testid="dc-day-label">
                      {label(r.week)} {weekday(r.week)}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {CATEGORIES.map(({ key, name }) => (
                  <tr key={key}>
                    <th scope="row">{name}</th>
                    {rows.map((r) => (
                      <td key={r.week} data-testid={key === "overallUsd" ? "dc-day" : undefined}>
                        {usd(r[key])}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
      {span && (
        <>
          <BarChart bars={weekdayBars} aria={`Average cost per weekday: ${WEEKDAYS.map((w) => `${w} ${mean(w)}`).join(", ")}`} />
          <div className="lb-panel">
            <div className="lb-panel-h">
              <span>Average per weekday</span>
            </div>
            <div className="lb-scroll">
              <table className="lb-table">
                <thead>
                  <tr>
                    {WEEKDAYS.map((w) => (
                      <th key={w}>{w}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    {WEEKDAYS.map((w) => (
                      <td key={w} data-testid="dc-avg" data-weekday={w}>
                        {mean(w)}
                      </td>
                    ))}
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </section>
  );
}
