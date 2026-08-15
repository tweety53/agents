// PeriodPicker is the "from"/"to" control every one of the eight views is
// driven by -- both parameters are required by the server (api.ts's own
// StatsViewParams doc comment; internal/api/stats.go's
// parsePeriodAndProject), so this component always has a concrete `from`
// and `to` rather than an optional one a view could silently treat as
// "everything".
export interface Period {
  from: Date;
  to: Date;
}

export interface PeriodPickerProps {
  period: Period;
  onChange: (period: Period) => void;
}

const DAY_MS = 24 * 60 * 60 * 1000;

// The dashboard's own default period -- the preceding 30 days ending now
// (App.tsx's own `defaultPeriod`, moved here so the "30 days" preset below
// and the reset button share one definition with it rather than each
// carrying a copy of "30 * DAY_MS" that could drift out of step).
export function defaultPeriod(): Period {
  const to = new Date();
  return { from: new Date(to.getTime() - 30 * DAY_MS), to };
}

function todayMidnight(): Date {
  const midnight = new Date();
  midnight.setHours(0, 0, 0, 0);
  return midnight;
}

// The five presets design.md's "Presets" section specifies, in order.
// Each `range` is a function rather than a precomputed value: it is called
// at click time (to set the period) and at every render (to derive which
// preset, if any, matches the current period) so "now" is always read
// fresh -- a stored range would go stale the instant time moves on.
const PRESETS: { name: string; range: () => Period }[] = [
  { name: "Today", range: () => ({ from: todayMidnight(), to: new Date() }) },
  { name: "7 days", range: () => ({ from: new Date(Date.now() - 7 * DAY_MS), to: new Date() }) },
  { name: "30 days", range: defaultPeriod },
  { name: "90 days", range: () => ({ from: new Date(Date.now() - 90 * DAY_MS), to: new Date() }) },
  { name: "All time", range: () => ({ from: new Date("2020-01-01T00:00:00Z"), to: new Date() }) },
];

// A preset's `range()` and `defaultPeriod()` both read "now" fresh on
// every call (the comment above PRESETS explains why), which means the
// `to` bound they produce moves forward on every re-render while a
// period already set stays exactly where it was snapshotted. Comparing
// for exact equality therefore stops matching the instant any time
// passes at all -- an unrelated re-render a few seconds after picking
// "7 days" would flip that preset's button back to unpressed, and would
// make the reset button appear under a period that is still, as far as
// anyone using this page is concerned, the default one.
//
// Treating two ranges as the same when both bounds fall within a minute
// of each other fixes that without storing which preset was picked (a
// second source of truth this project's rule against storing derivable
// state forbids): a range ending "40 seconds ago" reads as the same
// range as one ending "now" to anyone looking at it, and a minute is far
// longer than a render gap yet far shorter than a human deliberately
// re-picking the same preset. `DataTable.tsx` imports this rather than
// reimplementing it, so the two places that make this comparison can't
// drift apart on what "the same" means.
const SAME_RANGE_TOLERANCE_MS = 60 * 1000;

export function sameRange(a: Period, b: Period): boolean {
  return (
    Math.abs(a.from.getTime() - b.from.getTime()) <= SAME_RANGE_TOLERANCE_MS &&
    Math.abs(a.to.getTime() - b.to.getTime()) <= SAME_RANGE_TOLERANCE_MS
  );
}

// Local-time <input type="datetime-local"> has no timezone of its own; it
// is interpreted (and re-serialised) in the browser's local zone, which is
// fine here since the value only ever round-trips through `new Date(...)`
// before reaching api.ts, which converts to RFC 3339 UTC itself
// (buildStatsViewQuery's qs.set(..., params.from.toISOString())).
function toInputValue(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function PeriodPicker({ period, onChange }: PeriodPickerProps) {
  return (
    // "dashboard-time-range" is the styling hook DashboardBar.tsx uses to
    // place this control on the bar's own right edge, matching Grafana's
    // own top-bar layout (proposal.md's Fix round: "a top bar carrying a
    // time-range picker ... on the right and template-variable dropdowns
    // ... on the left"); the two inner labels carry "dashboard-variable",
    // the same class every other control on the bar carries, so all of
    // them draw from the same input styling in styles.css.
    <fieldset className="period-picker dashboard-time-range">
      <legend>Period</legend>
      <label className="dashboard-variable">
        From
        <input
          type="datetime-local"
          aria-label="Period start"
          value={toInputValue(period.from)}
          onChange={(e) => {
            const next = new Date(e.target.value);
            if (!Number.isNaN(next.getTime())) {
              onChange({ from: next, to: period.to });
            }
          }}
        />
      </label>
      <label className="dashboard-variable">
        To
        <input
          type="datetime-local"
          aria-label="Period end"
          value={toInputValue(period.to)}
          onChange={(e) => {
            const next = new Date(e.target.value);
            if (!Number.isNaN(next.getTime())) {
              onChange({ from: period.from, to: next });
            }
          }}
        />
      </label>
      <div className="period-picker-presets">
        {PRESETS.map(({ name, range }) => (
          <button
            key={name}
            type="button"
            aria-pressed={sameRange(period, range())}
            onClick={() => onChange(range())}
          >
            {name}
          </button>
        ))}
        {!sameRange(period, defaultPeriod()) && (
          <button type="button" onClick={() => onChange(defaultPeriod())}>
            Reset to default
          </button>
        )}
      </div>
    </fieldset>
  );
}
