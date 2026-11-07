// Flow health -- "Which guards catch something, which never fire, and what
// does each cost in run time?" (decision flow-health-question), built as the
// approved prototype draws it. The eight period totals it used to show are
// gone, each answered by a sibling view (drop-eight-tiles); the name and
// route stay (keep-name-and-route).
//
// One endpoint, internal/api/health.go's guard-health: one row per guard with
// runs in the window and project, flow-guard's helper commands left out.
// The view owns its filters: a KAN-1048 chip bar of Period (7d / 30d / 90d,
// a rolling window computed when the chip is chosen, 30d by default --
// default-period-30d), Project (from the response's `projects`, each with
// its guard-run count) and Rank by (Fires by default -- default-rank-fires),
// replacing the shared dashboard bar, so App's period, project and model
// props are ignored (chip-filters-period-project-rank). Chip state is local.
//
// Every rank, tile, share and footer total is derived here from the rows
// (candidates-derived-in-spa): ties break by guard name (ties-by-guard-name);
// the speed-up candidates are the fewest guards that together take 80% of
// all run time, over every row of the response (speed-up-candidate-rule),
// shown as a pill, a share column and a tile (speed-up-shown-pill-share-tile).
// Each tile is a toggle filtering the table to its set
// (tiles-toggle-table-filter). The panel note is the prototype's text
// (panel-note-as-prototype); formats follow flow-health-display-formats;
// loading, error, not-recorded and an empty window read as Reviewers renders
// them (flow-health-empty-and-loading).
import { type ReactNode, useState } from "react";
import type { GuardHealthRow, GuardProject } from "../api";
import { Chip, ChipGroup } from "../components/Chip";
import { Panel } from "../components/Panel";
import { PERIOD_DAYS, type PeriodChip, windowFor } from "../format";
import { useStatsView } from "../hooks/useStatsView";
import { projectLabel } from "../lib/projectLabel";
import type { ViewProps } from "../viewTypes";

const RANKS = { fires: "Fires", rate: "Fire rate", silent: "Never fired", time: "Run time" } as const;
type Rank = keyof typeof RANKS;
const TITLE: Record<Rank, string> = {
  fires: "Every guard, most fires first",
  rate: "Every guard, highest fire rate first",
  silent: "Every guard, never-fired first, most runs first",
  time: "Every guard, most run time first",
};
const SHOW = {
  firing: "Catching guards",
  silent: "Never-fired guards",
  ca: "Guards with cannot-answer runs",
  speed: "Speed-up candidates",
} as const;
type Show = keyof typeof SHOW;

interface Filters {
  period: PeriodChip;
  project: string | null;
  from: Date;
  to: Date;
}

const defaults = (): Filters => ({ ...windowFor("30d"), project: null });

const int = (v: number) => v.toLocaleString("en-US");
const pct = (v: number) => (v === 0 ? "0%" : v < 0.001 ? "<0.1%" : v < 0.1 ? `${(v * 100).toFixed(1)}%` : `${Math.round(v * 100)}%`);
const dur = (ms: number) => {
  if (Math.round(ms) < 1000) return `${Math.round(ms)} ms`;
  const tenths = Math.round(ms / 100);
  if (tenths < 600) return `${(tenths / 10).toFixed(1)} s`;
  const m = Math.round(ms / 60000);
  if (m < 60) return `${m} min`;
  return `${Math.floor(m / 60)} h ${m % 60} min`;
};
const total = (rows: GuardHealthRow[]) => rows.reduce((s, r) => s + r.totalMs, 0);

const byName = (a: GuardHealthRow, b: GuardHealthRow) => a.guard.localeCompare(b.guard);
const ORDER: Record<Rank, (a: GuardHealthRow, b: GuardHealthRow) => number> = {
  fires: (a, b) => b.fired - a.fired || b.runs - a.runs,
  rate: (a, b) => b.fired / b.runs - a.fired / a.runs || b.fired - a.fired,
  silent: (a, b) => Number(a.fired > 0) - Number(b.fired > 0) || b.runs - a.runs,
  time: (a, b) => b.totalMs - a.totalMs,
};
const ranked = (rows: GuardHealthRow[], rank: Rank) => [...rows].sort((a, b) => ORDER[rank](a, b) || byName(a, b));

/** The fewest guards, by summed run time, that together take 80% of all guard run time:
 * each is added while the running sum is below 80% of the total, so the one crossing the line
 * is a candidate, and a zero total marks none. */
export function speedUpCandidates(rows: GuardHealthRow[]): Set<string> {
  const all = total(rows);
  const out = new Set<string>();
  let acc = 0;
  for (const r of ranked(rows, "time")) {
    if (acc >= 0.8 * all) break;
    out.add(r.guard);
    acc += r.totalMs;
  }
  return out;
}

const KEEP: Record<Show, (r: GuardHealthRow, cand: Set<string>) => boolean> = {
  firing: (r) => r.fired > 0,
  silent: (r) => r.fired === 0,
  ca: (r) => r.cannotAnswer > 0,
  speed: (r, cand) => cand.has(r.guard),
};

const Slot = ({ name }: { name: string }) => <span className="rv-slot">{name}</span>;

interface TileProps {
  show: Show | null;
  k: Show;
  label: string;
  val: ReactNode;
  cls?: string;
  note?: ReactNode;
  onToggle: (k: Show) => void;
}

function Tile({ show, k, label, val, cls = "", note, onToggle }: TileProps) {
  return (
    <button type="button" className="lb-tile" aria-pressed={show === k} onClick={() => onToggle(k)}>
      <span className="lb-lbl">{label}</span>
      <span className={`lb-val ${cls}`.trim()}>{val}</span>
      {note && <span className="sl-note">{note}</span>}
    </button>
  );
}

function Tiles({ rows, cand, show, onToggle }: { rows: GuardHealthRow[]; cand: Set<string>; show: Show | null; onToggle: (k: Show) => void }) {
  const firing = rows.filter((r) => r.fired > 0);
  const silent = rows.filter((r) => r.fired === 0).sort((a, b) => b.runs - a.runs || byName(a, b));
  const top = ranked(firing, "fires")[0];
  const runs = rows.reduce((s, r) => s + r.runs, 0);
  const ca = rows.reduce((s, r) => s + r.cannotAnswer, 0);
  const all = total(rows);
  const worstCa = rows.filter((r) => r.runs >= 20).sort((a, b) => b.cannotAnswer / b.runs - a.cannotAnswer / a.runs || byName(a, b))[0];
  const cands = ranked(rows, "time").filter((r) => cand.has(r.guard));
  const more = (n: number) => (n > 3 ? `, +${n - 3} more` : "");
  const list = (rs: GuardHealthRow[], tail: (r: GuardHealthRow) => string) =>
    rs.slice(0, 3).map((r, i) => (
      <span key={r.guard}>
        {i > 0 && ", "}
        <Slot name={r.guard} /> {tail(r)}
      </span>
    ));
  const of = (text: string) => (
    <span className="lb-muted" style={{ fontSize: 14 }}>
      {text}
    </span>
  );
  return (
    <div className="rv-tiles">
      <Tile
        show={show}
        k="firing"
        label="Catching · fired at least once"
        val={
          <>
            {firing.length} {of(`of ${rows.length}`)}
          </>
        }
        note={
          top && (
            <>
              most: <Slot name={top.guard} /> {int(top.fired)} fires
            </>
          )
        }
        onToggle={onToggle}
      />
      <Tile
        show={show}
        k="silent"
        label="Never fired · ran, caught nothing"
        val={String(silent.length)}
        cls={silent.length ? "rv-worst" : ""}
        note={
          silent.length > 0 && (
            <>
              {list(silent, (r) => `(${int(r.runs)} runs)`)}
              {more(silent.length)}
            </>
          )
        }
        onToggle={onToggle}
      />
      <Tile
        show={show}
        k="ca"
        label="Cannot answer · share of runs"
        val={pct(runs ? ca / runs : 0)}
        note={
          worstCa &&
          worstCa.cannotAnswer > 0 && (
            <>
              worst: <Slot name={worstCa.guard} /> <span className="rv-worst">{pct(worstCa.cannotAnswer / worstCa.runs)}</span>
            </>
          )
        }
        onToggle={onToggle}
      />
      <Tile
        show={show}
        k="speed"
        label="Speed-up candidates · 80% of run time"
        val={
          <>
            {cands.length} {of(`of ${dur(all)} total`)}
          </>
        }
        note={
          cands.length > 0 && (
            <>
              {list(cands, (r) => pct(r.totalMs / all))}
              {more(cands.length)}
            </>
          )
        }
        onToggle={onToggle}
      />
    </div>
  );
}

function Table({ rows, all, cand, rank }: { rows: GuardHealthRow[]; all: number; cand: Set<string>; rank: Rank }) {
  const sort = (k: Rank) => (k === rank ? "descending" : undefined);
  const share = (v: number) => (
    <>
      {pct(v)}
      <span className="wc-sharebar">
        <span style={{ width: `${(v * 100).toFixed(1)}%` }} />
      </span>
    </>
  );
  const t = rows.reduce(
    (s, r) => ({ runs: s.runs + r.runs, fired: s.fired + r.fired, ca: s.ca + r.cannotAnswer, total: s.total + r.totalMs }),
    { runs: 0, fired: 0, ca: 0, total: 0 },
  );
  return (
    <table className="lb-table">
      <thead>
        <tr>
          <th>#</th>
          <th>Guard</th>
          <th className="r">Runs</th>
          <th className="r" aria-sort={sort("fires")}>
            Fired
          </th>
          <th className="r" aria-sort={sort("rate")}>
            Fire rate
          </th>
          <th className="r">Cannot answer</th>
          <th className="r">Median run</th>
          <th className="r" aria-sort={sort("time")}>
            Run time
          </th>
          <th className="r">Share of run time</th>
          <th aria-sort={sort("silent")}>Last fired</th>
        </tr>
      </thead>
      <tbody>
        {ranked(rows, rank).map((r, i) => (
          <tr key={r.guard} className={r.fired ? undefined : "lb-muted"}>
            <td className="lb-num lb-muted">{i + 1}</td>
            <td>
              <Slot name={r.guard} />
              {cand.has(r.guard) && <span className="lb-pill fh-speed">speed-up</span>}
            </td>
            <td className="r lb-num">{int(r.runs)}</td>
            <td className="r lb-num">{r.fired ? int(r.fired) : "—"}</td>
            <td className="r lb-num">{pct(r.fired / r.runs)}</td>
            <td className="r lb-num">{r.cannotAnswer ? pct(r.cannotAnswer / r.runs) : "—"}</td>
            <td className="r lb-num">{dur(r.medianMs)}</td>
            <td className="r lb-num">{dur(r.totalMs)}</td>
            <td className="r lb-num">{share(all ? r.totalMs / all : 0)}</td>
            <td className="lb-num">{r.lastFiredAt ? r.lastFiredAt.slice(0, 10) : <span className="lb-muted">never</span>}</td>
          </tr>
        ))}
      </tbody>
      <tfoot>
        <tr>
          <td></td>
          <td>
            {rows.length} {rows.length === 1 ? "guard" : "guards"}
          </td>
          <td className="r lb-num">{int(t.runs)}</td>
          <td className="r lb-num">{int(t.fired)}</td>
          <td className="r lb-num">{pct(t.runs ? t.fired / t.runs : 0)}</td>
          <td className="r lb-num">{pct(t.runs ? t.ca / t.runs : 0)}</td>
          <td></td>
          <td className="r lb-num">{dur(t.total)}</td>
          <td className="r lb-num">{pct(all ? t.total / all : 0)}</td>
          <td></td>
        </tr>
      </tfoot>
    </table>
  );
}

export function FlowHealth(_props: ViewProps) {
  const [f, setF] = useState<Filters>(defaults);
  const [rank, setRank] = useState<Rank>("fires");
  const [show, setShow] = useState<Show | null>(null);
  const state = useStatsView<GuardHealthRow[], GuardProject>("guard-health", {
    from: f.from,
    to: f.to,
    project: f.project ?? undefined,
    projects: true,
  });

  const offDefaults = f.period !== "30d" || f.project !== null || show !== null;
  const ready = state.status === "ready" && state.data.recorded && !state.data.unmeasured;
  const rows = state.status === "ready" ? state.data.rows : [];
  const projects = state.status === "ready" ? (state.data.projects ?? []) : [];
  const cand = speedUpCandidates(rows);
  const shown = show ? rows.filter((r) => KEEP[show](r, cand)) : rows;
  const title = show ? `${SHOW[show]}, ${TITLE[rank].replace(/^Every guard, /, "")}` : TITLE[rank];

  return (
    <section aria-label="Flow health" className="dashboard lb sl">
      <div>
        <h2>Flow health</h2>
        <p className="lb-sub">Which guards catch something, which never fire, and what does each cost in run time?</p>
      </div>
      <div className="lb-fbar" role="toolbar" aria-label="Filters">
        <ChipGroup name="Period">
          {(Object.keys(PERIOD_DAYS) as PeriodChip[]).map((p) => (
            <Chip key={p} on={f.period === p} onClick={() => setF({ ...f, ...windowFor(p) })}>
              {p}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Project">
          <Chip on={f.project === null} onClick={() => setF({ ...f, project: null })}>
            All
          </Chip>
          {projects.map((p) => (
            <Chip key={p.projectKey} on={f.project === p.projectKey} onClick={() => setF({ ...f, project: p.projectKey })} title={p.projectKey}>
              {projectLabel(p.projectKey)}
              <span className="lb-c">{int(p.runs)}</span>
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Rank by">
          {(Object.keys(RANKS) as Rank[]).map((k) => (
            <Chip key={k} on={rank === k} onClick={() => setRank(k)}>
              {RANKS[k]}
            </Chip>
          ))}
        </ChipGroup>
        {offDefaults && (
          <button
            type="button"
            className="lb-fclear"
            onClick={() => {
              setF(defaults());
              setShow(null);
            }}
          >
            Clear filters
          </button>
        )}
      </div>
      {ready && rows.length > 0 && <Tiles rows={rows} cand={cand} show={show} onToggle={(k) => setShow(show === k ? null : k)} />}
      <div className="lb-panel">
        <div className="lb-panel-h">
          <span>{title}</span>
          <span className="lb-panel-note">
            fired = the guard refused; speed-up = the fewest guards that together take 80% of all guard run time; helper commands (project-get, workspace, run-capped, …) are not guards and are left out
          </span>
        </div>
        <div className="lb-scroll">
          {!ready ? (
            // Loading, error, not-recorded and unmeasured read as Panel renders them.
            <Panel title={title} state={state}>
              {() => null}
            </Panel>
          ) : !rows.length ? (
            <div className="sl-empty">No guard runs in this period.</div>
          ) : (
            <Table rows={shown} all={total(rows)} cand={cand} rank={rank} />
          )}
        </div>
      </div>
    </section>
  );
}
