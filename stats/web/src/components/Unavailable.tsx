// Unavailable renders a metric bag key that was never measured -- `value`
// is `null` -- visibly and semantically differently from a measured zero.
// This is the interface half of the constraint the schema (nullable
// columns, no "0" default) and the API (no `omitempty` on a metric field,
// per internal/api/stats.go's DTO comment) already hold: a nil pointer
// crosses the wire as an explicit JSON `null`, and this component is the
// one place that null is turned into something a reader sees, rather than
// being coerced to 0 by a careless `value ?? 0` somewhere in a view.
export interface UnavailableProps {
  value: number | null | undefined;
  /** Formats a present value; defaults to String(value). */
  format?: (value: number) => string;
  unit?: string;
}

export function Unavailable({ value, format, unit }: UnavailableProps) {
  if (value === null || value === undefined) {
    return (
      <span className="metric-unavailable" data-testid="unavailable" title="Not measured for this stage run">
        unavailable
      </span>
    );
  }
  const text = format ? format(value) : String(value);
  return (
    <span className="metric-value" data-testid="measured">
      {text}
      {unit ? ` ${unit}` : ""}
    </span>
  );
}
