// The stacked-bar chart panel Weekly costs and Daily costs share: a titled panel, a legend, a
// fixed-size SVG with a bar per slot and a segment per series, the Y axis on tickStep, the total
// above each bar and a hover tooltip. `partial` is the slot drawn translucent and dashed.
import { useState } from "react";
import { formatUsd } from "../format";

export interface Series {
  name: string;
  color: string;
  v: number[];
}
export interface Slot {
  key: string;
  label: string;
  /** A smaller second line under the label. */
  sub?: string;
}

/** The Y axis' tick step: the smallest 1, 2, 2.5 or 5 times a power of ten
 * whose five steps reach `max`, so a $3 week fills the plot as a $300 one does. */
export function tickStep(max: number): number {
  const raw = max / 5 || 1;
  const magnitude = 10 ** Math.floor(Math.log10(raw));
  return [1, 2, 2.5, 5, 10].find((m) => m * magnitude >= raw)! * magnitude;
}

export const Swatch = ({ color }: { color: string }) => <span className="wc-sw" style={{ background: color }} />;

// Chart geometry, as the prototype draws it.
const CW = 1140,
  CH = 300,
  L = 60,
  R = 10,
  T = 10,
  B = 34,
  PW = CW - L - R,
  PH = CH - T - B;

interface Props {
  title: string;
  slots: Slot[];
  series: Series[];
  totals: number[];
  partial?: number | null;
  /** The aria-label of a segment. */
  segLabel: (series: string, i: number) => string;
  /** The tooltip's heading. */
  tipTitle: (i: number) => string;
  aria?: string;
  /** A test id for each slot's group, with its key in data-key. */
  slotTestId?: string;
}

export function StackedBars({ title, slots, series, totals, partial = null, segLabel, tipTitle, aria, slotTestId }: Props) {
  const [hover, setHover] = useState<number | null>(null);
  const step = tickStep(Math.max(0, ...totals) * 1.08);
  const y = (v: number) => T + PH - (PH * v) / (step * 5);
  const bw = PW / slots.length;

  return (
    <div className="lb-panel wc-chartpanel">
      <div className="lb-panel-h">
        <span>{title}</span>
      </div>
      <div className="lb-panel-b wc-chart">
        <div className="wc-legend">
          {series.map((s) => (
            <span key={s.name}>
              <Swatch color={s.color} />
              {s.name}
            </span>
          ))}
        </div>
        <div className="wc-plot">
          <svg viewBox={`0 0 ${CW} ${CH}`} width={CW} height={CH} {...(aria ? { role: "img", "aria-label": aria } : {})}>
            <g className="grid">
              {Array.from({ length: 6 }, (_, i) => (
                <g key={i}>
                  <line x1={L} x2={L + PW} y1={y(i * step)} y2={y(i * step)} />
                  <text x={L - 6} y={y(i * step) + 4} textAnchor="end">
                    {formatUsd(i * step)}
                  </text>
                </g>
              ))}
            </g>
            {slots.map((slot, w) => {
              const x = L + w * bw + bw * 0.18;
              const ww = bw * 0.64;
              let acc = 0;
              return (
                <g key={slot.key} data-testid={slotTestId} data-key={slotTestId ? slot.key : undefined}>
                  {!totals[w] && (
                    <text x={x + ww / 2} y={y(0) - 6} textAnchor="middle">
                      no data
                    </text>
                  )}
                  {series.map((s) => {
                    const v = s.v[w];
                    if (!v) return null;
                    const top = y(acc + v);
                    const height = y(acc) - top;
                    acc += v;
                    return (
                      <rect
                        key={s.name}
                        className="wc-bar"
                        tabIndex={0}
                        data-testid="wc-segment"
                        data-week={slot.key}
                        data-series={s.name}
                        aria-label={`${segLabel(s.name, w)}: ${formatUsd(v)}`}
                        x={x}
                        width={ww}
                        y={top}
                        height={height}
                        fill={s.color}
                        {...(w === partial ? { fillOpacity: ".45", stroke: s.color, strokeDasharray: "3 2" } : {})}
                        onMouseEnter={() => setHover(w)}
                        onMouseLeave={() => setHover(null)}
                        onFocus={() => setHover(w)}
                        onBlur={() => setHover(null)}
                      />
                    );
                  })}
                  {totals[w] > 0 && (
                    <text x={x + ww / 2} y={y(acc) - 5} textAnchor="middle" className="wc-total" data-testid="wc-total" data-week={slot.key}>
                      {formatUsd(acc)}
                    </text>
                  )}
                  <text x={x + ww / 2} y={CH - B + 16} textAnchor="middle" data-testid="wc-week">
                    {slot.label}
                  </text>
                  {slot.sub && (
                    <text x={x + ww / 2} y={CH - B + 28} textAnchor="middle" className="wc-subl" data-testid="wc-sub">
                      {slot.sub}
                    </text>
                  )}
                </g>
              );
            })}
          </svg>
          {hover !== null && (
            <div
              className="wc-tip"
              role="tooltip"
              // The last two slots open the tip leftwards, so it stays inside the plot.
              style={{ left: L + (hover + 0.85) * bw - (hover >= slots.length - 2 ? 260 : 0), top: 40 }}
            >
              <strong>{tipTitle(hover)}</strong>
              <table>
                <tbody>
                  {series
                    .filter((s) => s.v[hover])
                    .map((s) => (
                      <tr key={s.name}>
                        <td>
                          <Swatch color={s.color} />
                          {s.name}
                        </td>
                        <td>{formatUsd(s.v[hover])}</td>
                      </tr>
                    ))}
                  <tr>
                    <td>
                      <strong>Total</strong>
                    </td>
                    <td>
                      <strong>{formatUsd(totals[hover])}</strong>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
