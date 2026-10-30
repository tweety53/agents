// Experiments draws the allowed-pairs experiments (KAN-1067, decision
// experiments-view) exactly as the approved prototype
// .superpowers/sdd/experiments-prototype.html does: the roles table in
// force with open overrides struck through, the experiment list, and the
// selected experiment's override, per-metric stat panels, weekly charts
// with the experiment window and before/after means, and its summary.
// Lower is better for every metric (store.ExperimentStats).
import { useEffect, useState } from "react";
import {
  fetchExperimentStats,
  fetchExperiments,
  fetchRoles,
  type ExperimentDTO,
  type ExperimentMetric,
  type ExperimentStats,
  type RoleEntry,
} from "../api";

const METRIC_LABELS: Record<string, string> = {
  reviewRounds: "Review rounds / change",
  round0CritImp: "Round-0 Crit+Imp",
  planCorrections: "Plan corrections",
  tokens: "Tokens / change (k)",
  wallClockMin: "Wall-clock (min)",
  verifyReds: "Verify reds",
  findingsPerTask: "Findings vs commits",
  findingsPerDispatch: "Findings / dispatch",
  lateRoundFindings: "Late-round findings",
  vvMismatches: "Visual mismatches",
  vvReruns: "Visual re-runs",
};

const MAIN_PAIR = "session's own pair";

const label = (name: string) => METRIC_LABELS[name] ?? name;
// Tokens are drawn in thousands, the label's (k).
const scale = (name: string, v: number) => (name === "tokens" ? v / 1000 : v);
const fmt = (v: number) => (v < 10 ? v.toFixed(1) : String(Math.round(v)));
const tipFmt = (v: number) => (v < 10 ? v.toFixed(2) : String(Math.round(v)));
// The daemon writes instants in its own zone; dates are read in UTC.
const day = (iso: string) => new Date(iso).toISOString().slice(0, 10);

function tablePairs(roles: RoleEntry[], role: string): string {
  if (role === "main") return MAIN_PAIR;
  return roles.find((r) => r.role === role && r.marker === "")?.pairs.join(", ") ?? "";
}

function RolesTable({ roles, onSelect }: { roles: RoleEntry[]; onSelect: (id: number) => void }) {
  return (
    <table>
      <tbody>
        <tr>
          <th>Role</th>
          <th>Allowed pairs</th>
          <th>Requires marker</th>
          <th>Experiment</th>
        </tr>
        {roles.map((r, i) => {
          const pairs = r.role === "main" ? MAIN_PAIR : r.pairs.join(", ");
          return (
            <tr key={`${r.role}-${i}`}>
              <td>
                <code>{r.role}</code>
              </td>
              <td>
                {r.override ? (
                  <>
                    <s>
                      <code>{pairs}</code>
                    </s>{" "}
                    <code className="exp">{r.override.pairs.join(", ")}</code>
                  </>
                ) : (
                  <code>{pairs}</code>
                )}
              </td>
              <td>{r.marker ? <code>{r.marker}</code> : "—"}</td>
              <td>
                {r.override ? (
                  <a
                    href="#"
                    onClick={(ev) => {
                      ev.preventDefault();
                      onSelect(r.override!.experimentId);
                    }}
                  >
                    #{r.override.experimentId}
                  </a>
                ) : (
                  "—"
                )}
              </td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}

function StatCell({ metric }: { metric: ExperimentMetric }) {
  const { name, before, after } = metric;
  let delta = null;
  if (before !== null && after !== null) {
    const b = scale(name, before);
    const a = scale(name, after);
    // ponytail: a zero before has no percentage; the arrow alone says which way.
    const d = b === 0 ? a : ((a - b) / b) * 100;
    const pct = b === 0 ? "" : `${Math.abs(d).toFixed(0)}% `;
    delta = (
      <div className={`d ${d > 0 ? "up" : "down"}`} style={{ whiteSpace: "nowrap" }}>
        {`${d > 0 ? "▲" : "▼"} ${pct}vs ${fmt(b)} before`}
      </div>
    );
  }
  return (
    <section className="panel stat" aria-label={label(name)}>
      <h3 className="ttl">{label(name)}</h3>
      <div className="v">{after === null ? "—" : fmt(scale(name, after))}</div>
      {delta}
    </section>
  );
}

// The prototype's 320×150 chart geometry.
const W = 320;
const H = 150;
const L = 36;
const R = 8;
const T = 10;
const B = 22;

function Chart({ metric, startedAt }: { metric: ExperimentMetric; startedAt: string }) {
  const { name } = metric;
  const v = metric.series.map((p) => scale(name, p.value));
  const lab = metric.series.map((p) => day(p.week).slice(5));
  const n = v.length;
  const start = new Date(startedAt).getTime();
  // The first week that holds the start or follows it; n when none does.
  let s = metric.series.findIndex((p) => new Date(p.week).getTime() + 7 * 864e5 > start);
  if (s === -1) s = n;
  const hi = Math.max(0, ...v) * 1.15 || 1;
  const x = (i: number) => (n > 1 ? L + (i * (W - L - R)) / (n - 1) : L + (W - L - R) / 2);
  const y = (t: number) => T + (1 - t / hi) * (H - T - B);
  const sx = Math.min(Math.max(x(s - 0.5), L), W - R);
  const before = metric.before === null ? null : scale(name, metric.before);
  const after = metric.after === null ? null : scale(name, metric.after);
  const col = after !== null && before !== null && after < before ? "var(--delta-down)" : "var(--delta-up)";
  const meanLine = (a: number, b: number, val: number, stroke: string, testId: string) => (
    <line data-testid={testId} x1={a} x2={b} y1={y(val)} y2={y(val)} stroke={stroke} strokeWidth="1.5" strokeDasharray="5 3" opacity=".8" />
  );
  const xl = [...new Set([0, s, n - 1])].filter((i) => i >= 0 && i < n);
  const area = n ? `M${x(0)},${y(0)} ` + v.map((t, i) => `L${x(i)},${y(t)}`).join(" ") + ` L${x(n - 1)},${y(0)} Z` : "";
  return (
    <svg viewBox={`0 0 ${W} ${H}`} role="img" aria-label={`${label(name)}, weekly, before vs after ${day(startedAt)}`}>
      <rect x={sx} y={T} width={W - R - sx} height={H - T - B} fill="var(--marker)" opacity=".07" />
      {[0, hi / 2, hi].map((t) => (
        <g key={t}>
          <line x1={L} x2={W - R} y1={y(t)} y2={y(t)} stroke="var(--border)" />
          <text x={L - 4} y={y(t) + 3} textAnchor="end" fontSize="9" fill="var(--text-muted)">
            {t < 10 ? t.toFixed(1) : Math.round(t)}
          </text>
        </g>
      ))}
      {xl.map((i) => (
        <text key={i} x={x(i)} y={H - 6} textAnchor="middle" fontSize="9" fill="var(--text-muted)">
          {lab[i]}
        </text>
      ))}
      {n > 0 && <path d={area} fill="var(--series-1)" opacity=".12" />}
      {n > 0 && (
        <polyline
          points={v.map((t, i) => `${x(i)},${y(t)}`).join(" ")}
          fill="none"
          stroke="var(--series-1)"
          strokeWidth="2"
          strokeLinejoin="round"
        />
      )}
      {before !== null && meanLine(L, sx, before, "var(--text-muted)", "mean-before")}
      {after !== null && meanLine(sx, W - R, after, col, "mean-after")}
      <line data-testid="start-marker" x1={sx} x2={sx} y1={T} y2={H - B} stroke="var(--marker)" strokeDasharray="3 3" />
      {v.map((t, i) => (
        <circle key={i} className="pt" cx={x(i)} cy={y(t)} r="3" fill="var(--series-1)" tabIndex={0}>
          <title>{`${lab[i]}: ${tipFmt(t)}`}</title>
        </circle>
      ))}
    </svg>
  );
}

function Detail({ e, stats, roles }: { e: ExperimentDTO; stats: ExperimentStats | null; roles: RoleEntry[] }) {
  const metrics = stats?.metrics ?? [];
  return (
    <section className="right" id="detail">
      <section className="panel" aria-label="Override">
        <h2 className="ttl">Override</h2>
        <div className="wrap">
          <table>
            <tbody>
              <tr>
                <th>Role</th>
                <th>Table pairs</th>
                <th>Experiment pairs</th>
              </tr>
              {Object.keys(e.cells)
                .sort()
                .map((role) => (
                  <tr key={role}>
                    <td>
                      <code>{role}</code>
                    </td>
                    <td>
                      <code>{tablePairs(roles, role)}</code>
                    </td>
                    <td>
                      <code>{e.cells[role].join(", ")}</code>
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
        {stats ? (
          <p className="sub">
            Runs counted: {stats.runs.before} before · {stats.runs.after} after start
          </p>
        ) : (
          <p role="status">Loading…</p>
        )}
      </section>
      <div className="cells">
        {metrics.map((m) => (
          <StatCell key={m.name} metric={m} />
        ))}
      </div>
      <section className="panel" aria-label="Weekly trend — before vs after start">
        <h2 className="ttl">Weekly trend — before vs after start</h2>
        <div className="legend">
          <span>
            <i style={{ background: "var(--series-1)" }} />
            weekly value
          </span>
          <span>
            <i className="dash" style={{ borderColor: "var(--text-muted)" }} />
            mean before
          </span>
          <span>
            <i className="dash" style={{ borderColor: "var(--delta-down)" }} />
            mean after (better)
          </span>
          <span>
            <i className="dash" style={{ borderColor: "var(--delta-up)" }} />
            mean after (worse)
          </span>
          <span>
            <i style={{ background: "var(--marker)", opacity: 0.35 }} />
            experiment window
          </span>
        </div>
        <div className="charts">
          {metrics.map((m) => (
            <figure key={m.name}>
              <figcaption>{label(m.name)}</figcaption>
              <Chart metric={m} startedAt={e.startedAt} />
            </figure>
          ))}
        </div>
      </section>
      <section className="panel" aria-label="Summary">
        <h2 className="ttl">Summary</h2>
        <div className="summary">{e.summary ? e.summary : "Open — the closing agent writes the summary."}</div>
      </section>
    </section>
  );
}

export function Experiments() {
  const [roles, setRoles] = useState<RoleEntry[] | null>(null);
  const [list, setList] = useState<ExperimentDTO[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [sel, setSel] = useState<number | null>(null);
  const [stats, setStats] = useState<ExperimentStats | null>(null);
  const [statsError, setStatsError] = useState<string | null>(null);

  useEffect(() => {
    let live = true;
    Promise.all([fetchRoles(), fetchExperiments()])
      .then(([r, all]) => {
        if (!live) return;
        // Open experiments first, then newest start first, as the prototype lists them.
        const l = [...all].sort((a, b) => Number(b.status === "open") - Number(a.status === "open") || b.startedAt.localeCompare(a.startedAt));
        setRoles(r);
        setList(l);
        setSel((cur) => cur ?? l[0]?.id ?? null);
      })
      .catch((err: Error) => live && setError(err.message));
    return () => {
      live = false;
    };
  }, []);

  useEffect(() => {
    if (sel === null) return;
    let live = true;
    setStats(null);
    setStatsError(null);
    fetchExperimentStats(sel)
      .then((s) => live && setStats(s))
      .catch((err: Error) => live && setStatsError(err.message));
    return () => {
      live = false;
    };
  }, [sel]);

  if (error) {
    return (
      <p role="alert" className="panel-error">
        Failed to load: {error}
      </p>
    );
  }
  if (!roles || !list) return <p role="status">Loading…</p>;
  const e = list.find((x) => x.id === sel);

  return (
    <div className="experiments">
      <section className="panel roles" aria-label="Current roles">
        <h2 className="ttl">Current roles — allowed pairs in force</h2>
        <div className="wrap" id="roles">
          <RolesTable roles={roles} onSelect={setSel} />
        </div>
        <p className="sub">
          From <code>skills/flow-contracts/model-policy.md</code>; a cell an open experiment overrides shows its experiment pairs, the
          table's own pairs struck through.
        </p>
      </section>
      <section className="panel list" aria-label="Experiments">
        <h2 className="ttl">Experiments</h2>
        <div id="list">
          {list.length === 0 && (
            <p className="sub">
              No experiments yet — open one with <code>flow experiment open</code>.
            </p>
          )}
          {list.map((x) => (
            <button key={x.id} type="button" className={x.id === sel ? "on" : ""} onClick={() => setSel(x.id)}>
              {x.name}
              <span className={`pill ${x.status}`}>{x.status}</span>
              <div className="sub">
                since {day(x.startedAt)}
                {x.closedAt ? ` · closed ${day(x.closedAt)}` : ""}
              </div>
            </button>
          ))}
        </div>
      </section>
      {statsError && (
        <p role="alert" className="panel-error">
          Failed to load stats: {statsError}
        </p>
      )}
      {e && <Detail e={e} stats={stats?.experiment.id === e.id ? stats : null} roles={roles} />}
    </div>
  );
}
