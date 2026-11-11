// Experiments draws the allowed-pairs experiments (KAN-1067, decision
// experiments-view) exactly as the approved prototype
// .superpowers/sdd/experiments-prototype.html does: the roles table in
// force with open overrides struck through -- grouped, one row per role, as
// .superpowers/sdd/roles-prototype.html draws it -- the experiment list, and the
// selected experiment's override, per-metric stat panels, weekly charts
// with the experiment window and before/after means, and its summary.
// Lower is better for every metric (store.ExperimentStats). Each role's base
// and opt-in pairs are edited here and saved to flowd, their one source
// (kan-1103); an open experiment can be made permanent.
import { Fragment, useEffect, useRef, useState } from "react";
import {
  PAIR_SHAPE,
  fetchExperimentStats,
  fetchExperiments,
  fetchRoles,
  promoteExperiment,
  putRolePairs,
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
  selfReviewFindings: "Self-review Crit+Imp",
  findingsPerTask: "Findings vs commits",
  findingsPerDispatch: "Findings / dispatch",
  lateRoundFindings: "Late-round findings",
  vvMismatches: "Visual mismatches",
  vvReruns: "Visual re-runs",
};

// Brief plain-English role names for the Does column; the full sentence is the hover.
// ponytail: a role missing here shows its sentence's first four words.
const SHORT_DOES: Record<string, string> = {
  main: "Runs the pipeline",
  explorer: "Reads the codebase",
  "plan-designer": "Proposes a plan",
  planner: "Writes the plan",
  implementer: "Writes the code",
  "panel-fix": "Fixes review findings",
  "operator-asked": "Does what you asked",
  "design-reviewer": "Reviews the plan",
  "gated-reviewer": "Reviews risky tasks",
  "gated-re-reviewer": "Re-checks risky tasks",
  "panel-reviewer": "Reviews the branch",
  "panel-sanity-efficiency": "Checks tests and waste",
  "panel-mutation-failure": "Breaks code on purpose",
  "panel-rerun": "Re-reviews the branch",
  "late-fix": "Reviews a late fix",
  "panel-merger": "Merges findings",
  "visual-verifier": "Compares UI to mockup",
  "tooling-analyst": "Explains a missed defect",
  "self-review-pass": "Reviews the run",
  "self-review-plan": "Plans review fixes",
  "self-review-fix": "Makes a review fix",
  "self-review-review-1": "Reviews that fix",
  "self-review-review-r": "Re-reviews that fix",
  "pipeline-fix-plan": "Plans a tooling fix",
  "pipeline-fix-fix": "Makes a tooling fix",
  "pipeline-fix-review-1": "Reviews that fix",
  "pipeline-fix-review-r": "Re-reviews that fix",
};
const shortDoes = (r: RoleEntry) => SHORT_DOES[r.role] ?? r.does.split(" ").slice(0, 4).join(" ");

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
  return roles.find((r) => r.role === role)?.base.join(", ") ?? "";
}

// A pairs list wraps between its pairs, never inside one at a pair's hyphen.
function PairList({ pairs, className }: { pairs: string[]; className?: string }) {
  return (
    <code className={className}>
      {pairs.map((p, i) => (
        <Fragment key={p}>
          {i > 0 && ", "}
          <span className="pair">{p}</span>
        </Fragment>
      ))}
    </code>
  );
}

// PairEditor is one tier's pairs as a multiselect field: the chips are its
// trigger, which opens a filterable checklist over every pair the roles use.
// Typing a well-formed pair (PAIR_SHAPE) not yet listed offers to add it.
// Escape, a click outside or Done closes it.
function PairEditor({ role, tier, pairs, catalog, onChange }: { role: string; tier: string; pairs: string[]; catalog: string[]; onChange: (p: string[]) => void }) {
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState("");
  const box = useRef<HTMLSpanElement>(null);
  // The menu is position: fixed, so the roles panel's scroll box never clips it;
  // it sits under the trigger, or above when the viewport has no room below.
  const [at, setAt] = useState<{ left: number; top?: number; bottom?: number }>({ left: 0 });
  useEffect(() => {
    if (!open) return;
    const r = box.current!.getBoundingClientRect();
    const up = window.innerHeight - r.bottom < 300 && r.top > window.innerHeight - r.bottom;
    setAt(up ? { left: r.left, bottom: window.innerHeight - r.top + 4 } : { left: r.left, top: r.bottom + 4 });
    const shut = () => setOpen(false);
    // A scroll inside the checklist itself keeps it open.
    const scrolled = (ev: Event) => !box.current?.contains(ev.target as Node) && shut();
    window.addEventListener("scroll", scrolled, true);
    window.addEventListener("resize", shut);
    const away = (ev: MouseEvent) => box.current && !box.current.contains(ev.target as Node) && setOpen(false);
    document.addEventListener("mousedown", away);
    return () => {
      document.removeEventListener("mousedown", away);
      window.removeEventListener("scroll", scrolled, true);
      window.removeEventListener("resize", shut);
    };
  }, [open]);
  const q = draft.trim();
  const canAdd = PAIR_SHAPE.test(q) && !pairs.includes(q) && !catalog.includes(q);
  const options = [...new Set([...catalog, ...pairs])].sort().filter((o) => o.toLowerCase().includes(q.toLowerCase()));
  const add = () => {
    onChange([...pairs, q]);
    setDraft("");
  };
  return (
    <span
      className={`ms ${tier}${open ? " open" : ""}`}
      ref={box}
      onKeyDown={(ev) => {
        if (ev.key === "Escape") setOpen(false);
      }}
    >
      <button type="button" className="trig" aria-haspopup="listbox" aria-expanded={open} title={pairs.join(", ")} aria-label={`Choose ${tier} pairs for ${role}`} onClick={() => setOpen(!open)}>
        <span className="lbl">{tier}</span>
        <span className="chips">
          {pairs.length === 0 ? (
            <span className="none">none</span>
          ) : (
            pairs.map((p) => (
              <span className="chip" key={p}>
                <span className="pair">{p}</span>
              </span>
            ))
          )}
        </span>
        {pairs.length > 1 && <span className="cnt">{pairs.length}</span>}
      </button>
      {open && (
        <div className="menu" style={at}>
          <input
            type="search"
            aria-label={`Add ${tier} pair to ${role}`}
            placeholder="Filter or type a new pair…"
            autoFocus
            value={draft}
            onChange={(ev) => setDraft(ev.target.value)}
            onKeyDown={(ev) => {
              if (ev.key === "Enter") {
                ev.preventDefault();
                if (canAdd) add();
              }
            }}
          />
          <div className="opts">
            {canAdd && (
              <label>
                <input type="checkbox" checked={false} onChange={add} />
                <span>Add “{q}”</span>
              </label>
            )}
            {options.map((o) => (
              <label key={o}>
                <input type="checkbox" checked={pairs.includes(o)} onChange={(ev) => onChange(ev.target.checked ? [...pairs, o] : pairs.filter((x) => x !== o))} />
                <span className="pair">{o}</span>
              </label>
            ))}
          </div>
          <div className="foot">
            <span>{pairs.length} selected</span>
            <button type="button" onClick={() => setOpen(false)}>
              Done
            </button>
          </div>
        </div>
      )}
    </span>
  );
}

function RoleRow({ r, catalog, onSelect, onSaved }: { r: RoleEntry; catalog: string[]; onSelect: (id: number) => void; onSaved: () => void }) {
  const [base, setBase] = useState(r.base);
  const [optin, setOptin] = useState(r.optin);
  const [saveError, setSaveError] = useState<string | null>(null);
  const dirty = base.join() !== r.base.join() || optin.join() !== r.optin.join();
  const main = r.role === "main";
  return (
    <tr className="role">
      <td className="role">
        <code>{r.role}</code>
      </td>
      <td className="does" title={r.does}>
        <span className="tag">{r.group}</span> {shortDoes(r)}
      </td>
      <td className="pairs">
        {main ? (
          r.override ? (
            <s>
              <PairList pairs={[MAIN_PAIR]} />
            </s>
          ) : (
            <PairList pairs={[MAIN_PAIR]} />
          )
        ) : r.override ? (
          <s>
            <PairEditor role={r.role} tier="base" pairs={base} catalog={catalog} onChange={setBase} />
          </s>
        ) : (
          <PairEditor role={r.role} tier="base" pairs={base} catalog={catalog} onChange={setBase} />
        )}
        {r.override && (
          <>
            {" "}
            <PairList pairs={r.override.pairs} className="exp" />
            <a
              href="#"
              onClick={(ev) => {
                ev.preventDefault();
                onSelect(r.override!.experimentId);
              }}
            >
              #{r.override.experimentId}
            </a>
          </>
        )}
        {!main && (
          <>
            {" "}
            <PairEditor role={r.role} tier="opt-in" pairs={optin} catalog={catalog} onChange={setOptin} />
            {r.marked.map((x) => (
              <span className="esc" key={x.pair + x.marker}>
                <PairList pairs={[x.pair]} /> with <code>{x.marker}</code>
              </span>
            ))}
            <button
              type="button"
              disabled={!dirty}
              aria-label={`Save ${r.role}`}
              onClick={() => {
                setSaveError(null);
                putRolePairs(r.role, { base, optin })
                  .then(onSaved)
                  .catch((err: Error) => setSaveError(err.message));
              }}
            >
              Save
            </button>
            {saveError && (
              <span role="alert" className="panel-error">
                {saveError}
              </span>
            )}
          </>
        )}
      </td>
    </tr>
  );
}

function RolesTable({ roles, onSelect, onSaved }: { roles: RoleEntry[]; onSelect: (id: number) => void; onSaved: () => void }) {
  const catalog = [...new Set(roles.flatMap((r) => [...r.base, ...r.optin, ...r.marked.map((m) => m.pair), ...(r.override?.pairs ?? [])]))];
  // One row per role, groups in order of first appearance.
  const groups: { name: string; roles: RoleEntry[] }[] = [];
  for (const r of roles) {
    let g = groups.find((x) => x.name === r.group);
    if (!g) groups.push((g = { name: r.group, roles: [] }));
    g.roles.push(r);
  }
  return (
    <table>
      <thead>
        <tr>
          <th>Role</th>
          <th>Does</th>
          <th>Allowed pairs</th>
        </tr>
      </thead>
      <tbody>
        {groups.map((g) => (
          <Fragment key={g.name}>
            <tr className="grp">
              <th colSpan={3}>{g.name}</th>
            </tr>
            {g.roles.map((r) => (
              // Keyed on the saved pairs, so a reload after Save resets the draft.
              <RoleRow key={`${r.role} ${r.base.join()} ${r.optin.join()}`} r={r} catalog={catalog} onSelect={onSelect} onSaved={onSaved} />
            ))}
          </Fragment>
        ))}
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
    const pct = b === 0 || d === 0 ? "" : `${Math.abs(d).toFixed(0)}% `;
    // An unchanged value is neither better nor worse: no arrow, no delta colour.
    delta = (
      <div className={d === 0 ? "d" : `d ${d > 0 ? "up" : "down"}`} style={{ whiteSpace: "nowrap" }}>
        {`${d === 0 ? "=" : d > 0 ? "▲" : "▼"} ${pct}vs ${fmt(b)} before`}
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
  const col = after === null || before === null || after === before ? "var(--text-muted)" : after < before ? "var(--delta-down)" : "var(--delta-up)";
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
        <text key={i} x={x(i)} y={H - 6} textAnchor={n > 1 && i === n - 1 ? "end" : n > 1 && i === 0 ? "start" : "middle"} fontSize="9" fill="var(--text-muted)">
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

function Detail({ e, stats, roles, onPromoted }: { e: ExperimentDTO; stats: ExperimentStats | null; roles: RoleEntry[]; onPromoted: () => void }) {
  const metrics = stats?.metrics ?? [];
  const [promoteError, setPromoteError] = useState<string | null>(null);
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
        {e.status === "open" && (
          <button
            type="button"
            onClick={() => {
              setPromoteError(null);
              promoteExperiment(e.id)
                .then(onPromoted)
                .catch((err: Error) => setPromoteError(err.message));
            }}
          >
            Make permanent
          </button>
        )}
        {promoteError && (
          <p role="alert" className="panel-error">
            {promoteError}
          </p>
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

  // Bumped after a save or a promote: reloads the roles and the list.
  const [reload, setReload] = useState(0);
  const refresh = () => setReload((n) => n + 1);

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
  }, [reload]);

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
          <RolesTable roles={roles} onSelect={setSel} onSaved={refresh} />
        </div>
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
      {e && <Detail key={e.id} e={e} stats={stats?.experiment.id === e.id ? stats : null} roles={roles} onPromoted={refresh} />}
    </div>
  );
}
