// Daily costs -- "which day cost how much": Weekly costs' categories per
// 24-hour day, the days starting at the client's local midnight. The Range
// control mirrors Reviewer findings': 7 or 14 days, stepped earlier and
// later by its own length, never past now. Below the days, an all-time total
// and each weekday's mean, both from one daily-costs fetch over the whole
// recorded span, so neither moves with Range.
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
  const total = (key: keyof Omit<WeeklyCostPoint, "week">) => recorded.reduce((a, r) => a + (r[key] ?? 0), 0);
  const mean = (wd: string) => {
    const mine = recorded.filter((r) => weekday(r.week) === wd);
    return mine.length ? usd(mine.reduce((a, r) => a + (r.overallUsd ?? 0), 0) / mine.length) : "—";
  };

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
          <div className="lb-panel">
            <div className="lb-panel-h">
              <span>All time</span>
            </div>
            <div className="lb-scroll">
              <table className="lb-table">
                <thead>
                  <tr>
                    {CATEGORIES.map(({ name }) => (
                      <th key={name}>{name}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    {CATEGORIES.map(({ key }) => (
                      <td key={key} data-testid={key === "overallUsd" ? "dc-total" : `dc-total-${key.replace("Usd", "")}`}>
                        {usd(total(key))}
                      </td>
                    ))}
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div className="lb-panel">
            <div className="lb-panel-h">
              <span>Average per day of week, all time</span>
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
