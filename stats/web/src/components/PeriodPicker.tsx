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
    </fieldset>
  );
}
