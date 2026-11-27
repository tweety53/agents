// ReviewDynamics -- the Reviewer findings view (#/review-dynamics): findings
// per dispatch for each current review slot × model · effort, against the
// range just before it of the same length. One Range control picks the last
// 7, 14, 30 or 60 days and steps that window back and forth by its own
// length, never past now; the view carries no other filter.
import { useState } from "react";
import type { HeatCell, ReviewDynamicsRows } from "../api";
import { useStatsView } from "../hooks/useStatsView";

const DAY_MS = 86400000;
const RANGES = [7, 14, 30, 60] as const;
const mmdd = (d: Date) => d.toISOString().slice(5, 10);

// byEffort orders effort names low → medium → high (then anything else by
// name), and any other names alphabetically.
const EFFORTS = ["low", "medium", "high"];
export function byEffort(a: string, b: string): number {
  const rank = (n: string) => {
    const i = EFFORTS.indexOf(n.split(" ").pop()!);
    return i < 0 ? EFFORTS.length : i;
  };
  return rank(a) - rank(b) || a.localeCompare(b);
}

// CURRENT_SLOTS is every slot the heatmap shows: the panel's current slots,
// as the panelSlots internal/api records against -- a retired slot is not
// shown.
const CURRENT_SLOTS = new Set([
  "primary",
  "principles",
  "test-sanity",
  "call-efficiency",
  "failure-modes",
  "mutation",
]);

function Heatmap({ cells, previous }: { cells: HeatCell[]; previous: HeatCell[] }) {
  const rate = (c: HeatCell) => (c.dispatches === 0 ? 0 : (c.critical + c.important + c.minor) / c.dispatches);
  const key = (c: { slot: string; model: string; effort: string }) => `${c.slot}|${c.model}|${c.effort}`;
  // Rows run busiest slot first, as the panel roster does; columns run by
  // model, then effort low → high (heatmap-order).
  const dispatches = new Map<string, number>();
  for (const c of cells) dispatches.set(c.slot, (dispatches.get(c.slot) ?? 0) + c.dispatches);
  const slots = [...dispatches.keys()].sort((a, b) => dispatches.get(b)! - dispatches.get(a)! || a.localeCompare(b));
  const cols = [...new Set(cells.map((c) => `${c.model}|${c.effort}`))]
    .map((k) => k.split("|"))
    .sort(([am, ae], [bm, be]) => am.localeCompare(bm) || byEffort(ae, be));
  const byKey = new Map(cells.map((c) => [key(c), c]));
  const prevByKey = new Map(previous.map((c) => [key(c), c]));
  const max = Math.max(0, ...cells.map(rate)) || 1;

  return (
    <div className="dynamics-panel">
      <h2>Findings per dispatch: slot × model · effort</h2>
      <p className="dynamics-sub">
        Darker = more findings per dispatch. Arrow: change against the range just before, of the same length. Hover a
        cell for its counts.
      </p>
      <div className="dynamics-scroll">
        <table className="dynamics-heat">
          <thead>
            <tr>
              <th className="dynamics-heat-row" />
              {cols.map(([model, effort]) => (
                <th key={`${model}|${effort}`} scope="col">
                  {model}
                  <br />
                  {effort}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {slots.map((slot) => (
              <tr key={slot}>
                <th className="dynamics-heat-row" scope="row">
                  {slot}
                </th>
                {cols.map(([model, effort]) => {
                  const c = byKey.get(key({ slot, model, effort }));
                  if (!c || c.dispatches === 0) {
                    return (
                      <td key={model + effort} className="dynamics-heat-empty">
                        —
                      </td>
                    );
                  }
                  const r = rate(c);
                  const alpha = 0.12 + (0.78 * r) / max;
                  const prev = prevByKey.get(key(c));
                  const change = prev && rate(prev) > 0 ? ((r - rate(prev)) / rate(prev)) * 100 : null;
                  return (
                    <td
                      key={model + effort}
                      style={{
                        background: `rgba(var(--heat), ${alpha.toFixed(2)})`,
                        color: alpha > 0.55 ? "#111217" : "var(--text-primary)",
                      }}
                      title={`${slot} · ${model} ${effort}\n${c.dispatches} dispatches\n${c.critical} critical · ${c.important} important · ${c.minor} minor`}
                    >
                      <div className="dynamics-heat-r">{r.toFixed(2)}</div>
                      {change !== null && (
                        <div className="dynamics-heat-c">
                          {change < 0 ? "▼" : "▲"} {Math.abs(change).toFixed(0)}%
                        </div>
                      )}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function ReviewDynamics() {
  // end is when the view opened: every window is counted back from it, so
  // stepping later returns to exactly the windows stepped through.
  const [end] = useState(() => Date.now());
  const [days, setDays] = useState<(typeof RANGES)[number]>(7);
  const [back, setBack] = useState(0);
  const to = new Date(end - back * days * DAY_MS);
  const from = new Date(to.getTime() - days * DAY_MS);
  const state = useStatsView<ReviewDynamicsRows>("review-dynamics", { from, to });
  const rows = state.status === "ready" ? state.data.rows : null;
  const shown = (c: HeatCell) => CURRENT_SLOTS.has(c.slot);

  return (
    <section aria-label="Reviewer findings" className="dashboard">
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
        <span className="dynamics-sub dynamics-inline">{`${mmdd(from)} – ${mmdd(to)}`}</span>
        <button type="button" aria-label="Later" disabled={back === 0} onClick={() => setBack((b) => b - 1)}>
          ›
        </button>
      </div>
      {state.status === "loading" && <p role="status">Loading…</p>}
      {state.status === "error" && <p role="alert">Failed to load: {state.message}</p>}
      {rows && <Heatmap cells={rows.heat.filter(shown)} previous={rows.previousHeat} />}
    </section>
  );
}
