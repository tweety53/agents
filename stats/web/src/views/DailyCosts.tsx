// Daily costs -- "which day cost how much": Weekly costs' categories per
// 24-hour day, the days starting at the client's local midnight. The Range
// control mirrors Reviewer findings': 7 or 14 days, stepped earlier and
// later by its own length, never past now. Below the days, each weekday's mean
// from one daily-costs fetch over the whole recorded span, so it does not move
// with Range.
import { useState } from "react";
import type { WeeklyCostPoint } from "../api";
import { formatUsd } from "../format";
import { StackedBars, type Series } from "../components/StackedBars";
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
  const [range, setRange] = useState<(typeof RANGES)[number]>(7);
  const [back, setBack] = useState(0);
  const to = back === 0 ? new Date(end) : dayOffset(1 - back * range);
  const from = dayOffset(1 - back * range - range);
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
  const dayIsos = (rows ?? []).map((r) => r.week);
  const daySeries: Series[] = parts.map(({ key, name }, i) => ({ name, color: `var(--series-${i + 1})`, v: (rows ?? []).map((r) => r[key] ?? 0) }));
  const dayTotals = (rows ?? []).map((r) => r.overallUsd ?? 0);
  const means = WEEKDAYS.map((w) => meanOf(w) ?? 0);
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
            aria-pressed={range === d}
            onClick={() => {
              setRange(d);
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
        <StackedBars
          title="Spend per day by phase"
          slots={rows.map((r) => ({ key: r.week, label: label(r.week), sub: weekday(r.week) }))}
          series={daySeries}
          totals={dayTotals}
          slotTestId="dc-bar"
          segLabel={(name, i) => `${name}, ${label(dayIsos[i])} ${weekday(dayIsos[i])}`}
          tipTitle={(i) => `${label(dayIsos[i])} ${weekday(dayIsos[i])}`}
          aria={`Daily costs, ${label(rows[0]?.week ?? from.toISOString())} to ${label(rows.at(-1)?.week ?? to.toISOString())}, stacked by ${parts.map((c) => c.name).join(", ")}: ${rows.map((r) => `${label(r.week)} ${usd(r.overallUsd)}`).join(", ")}`}
        />
      )}
      {span && (
        <StackedBars
          title="Average per weekday"
          slots={WEEKDAYS.map((w) => ({ key: w, label: w }))}
          series={[{ name: "Average", color: "var(--series-1)", v: means }]}
          totals={means}
          slotTestId="dc-wbar"
          segLabel={(_, i) => `${WEEKDAYS[i]} average`}
          tipTitle={(i) => `${WEEKDAYS[i]} average`}
          aria={`Average cost per weekday: ${WEEKDAYS.map((w) => `${w} ${mean(w)}`).join(", ")}`}
        />
      )}
    </section>
  );
}
