// DynamicsChart draws one Review dynamics trend panel (design.md's
// dynamics-chart-component): every series on one shared y-scale with three
// ticks, at most six date labels, a dashed marker per prompt change in the
// chart's area, and a hover crosshair whose tooltip lists each series'
// value, what it is based on and the marker at that point. Geometry is the
// approved prototype's chart(): a 480×170 viewBox, padding l34 r10 t12 b22.
import { useState } from "react";

export interface DynamicsSeries {
  name: string;
  /** One value per bucket, null where the series has no point. */
  values: (number | null)[];
  basis: (number | null)[];
}

export interface DynamicsMarker {
  index: number;
  what: string;
  text: string;
}

export interface DynamicsChartProps {
  title: string;
  description: string;
  buckets: string[];
  bucketKind: "day" | "week";
  series: DynamicsSeries[];
  markers: DynamicsMarker[];
  basisLabel: string;
  format?: (v: number) => string;
}

const W = 480;
const H = 170;
const P = { l: 34, r: 10, t: 12, b: 22 };

const mmdd = (bucket: string) => bucket.slice(5, 10);

export function DynamicsChart({
  title,
  description,
  buckets,
  bucketKind,
  series,
  markers,
  basisLabel,
  format = (v) => v.toFixed(1),
}: DynamicsChartProps) {
  const [hover, setHover] = useState<number | null>(null);
  const n = buckets.length;
  const all = series.flatMap((s) => s.values.filter((v): v is number => v !== null));
  const max = Math.max(0, ...all) * 1.12 || 1;
  const x = (i: number) => P.l + (n < 2 ? (W - P.l - P.r) / 2 : (i * (W - P.l - P.r)) / (n - 1));
  const y = (v: number) => H - P.b - (v * (H - P.t - P.b)) / max;
  const step = Math.max(1, Math.ceil(n / 6));
  const band = n < 2 ? W - P.l - P.r : (W - P.l - P.r) / (n - 1);
  const colour = (j: number) => `var(--series-${(j % 7) + 1})`;

  const basisAt = (i: number) => Math.max(0, ...series.map((s) => s.basis[i] ?? 0));
  const tip =
    hover === null ? null : (
      <div className="dynamics-tip" role="tooltip" style={{ left: `${(x(hover) / W) * 100}%` }}>
        <b>{bucketKind === "week" ? `Week of ${mmdd(buckets[hover])}` : mmdd(buckets[hover])}</b>
        {series.map((s, j) => (
          <div key={s.name}>
            <i style={{ background: colour(j) }} />
            {s.name}: {s.values[hover] === null ? "—" : format(s.values[hover]!)}
          </div>
        ))}
        <div className="dynamics-tip-basis">
          based on {basisAt(hover)} {basisLabel}
        </div>
        {markers
          .filter((m) => m.index === hover)
          .map((m) => (
            <div key={m.what + m.text} className="dynamics-tip-marker">
              ◆ {m.what}: {m.text}
            </div>
          ))}
      </div>
    );

  return (
    <section className="dynamics-panel" aria-label={title}>
      <h2>{title}</h2>
      <p className="dynamics-sub">{description}</p>
      <svg viewBox={`0 0 ${W} ${H}`} role="img" aria-label={title} onMouseLeave={() => setHover(null)}>
        <line className="dynamics-axis" x1={P.l} y1={H - P.b} x2={W - P.r} y2={H - P.b} />
        {[0, max / 2.24, max / 1.12].map((t) => (
          <g key={t}>
            <text className="dynamics-tick" x={P.l - 4} y={y(t) + 3} textAnchor="end">
              {format(t)}
            </text>
            <line className="dynamics-axis" x1={P.l} x2={W - P.r} y1={y(t)} y2={y(t)} strokeOpacity={0.35} />
          </g>
        ))}
        {buckets.map((b, i) =>
          i % step === 0 ? (
            <text key={b} className="dynamics-tick" x={x(i)} y={H - 6} textAnchor="middle">
              {mmdd(b)}
            </text>
          ) : null,
        )}
        {markers.map((m) => (
          <g key={`${m.index}-${m.what}-${m.text}`}>
            <line className="dynamics-marker" x1={x(m.index)} x2={x(m.index)} y1={P.t} y2={H - P.b} />
            <circle className="dynamics-marker-dot" cx={x(m.index)} cy={P.t} r={4} />
          </g>
        ))}
        {series.map((s, j) => (
          <g key={s.name}>
            <polyline
              points={s.values.flatMap((v, i) => (v === null ? [] : [`${x(i)},${y(v)}`])).join(" ")}
              fill="none"
              stroke={colour(j)}
              strokeWidth={2}
            />
            {s.values.map((v, i) => (v === null ? null : <circle key={i} cx={x(i)} cy={y(v)} r={2.3} fill={colour(j)} />))}
          </g>
        ))}
        {hover !== null && <line className="dynamics-cross" x1={x(hover)} x2={x(hover)} y1={P.t} y2={H - P.b} />}
        {buckets.map((b, i) => (
          <rect
            key={b}
            data-bucket={b}
            x={Math.max(P.l, x(i) - band / 2)}
            y={0}
            width={Math.min(band, W - P.r - Math.max(P.l, x(i) - band / 2))}
            height={H}
            fill="transparent"
            onMouseEnter={() => setHover(i)}
          />
        ))}
      </svg>
      <div className="dynamics-legend">
        {series.map((s, j) => (
          <span key={s.name}>
            <i style={{ background: colour(j) }} />
            {s.name}
          </span>
        ))}
        {markers.length > 0 && (
          <span>
            <i style={{ background: "var(--marker)" }} />
            prompt change
          </span>
        )}
      </div>
      {tip}
    </section>
  );
}
