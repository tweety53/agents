// The live state board -- the default route, laid out as the approved
// prototype (spectre/changes/kan-1048-rework-live-board-and-run-view/design.md):
// a chip filter bar, state tiles that filter by state, a Running now strip of
// one card per change with an open stage run, and the changes grouped by
// state. pipeline-fix-* changes are hidden here, never on the wire.
//
// The board fetches every change ever recorded and applies every filter
// here (design.md's board-client-side-filters): the list holds tens of rows,
// and each chip count must reflect the other filters without a request per
// click. It refetches every five seconds so a session's activity badge
// follows what the session is doing.
import { Fragment, useEffect, useRef, useState, type ReactNode } from "react";
import type { StateBoardRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { DAY_MS, formatInt, formatRelative, formatSpan, formatUsd } from "../format";
import { projectLabel } from "../lib/projectLabel";

const REFRESH_MS = 5000;
const EPOCH = new Date(0);

const STATES = ["IN_PROGRESS", "STARTED", "FINISHED"] as const;
const STATE_LABEL: Record<string, string> = { IN_PROGRESS: "In progress", STARTED: "Started", FINISHED: "Finished" };

export const ACTIVITY_LABEL: Record<string, string> = {
  working: "Working",
  question: "Awaiting your answer",
  permission: "Awaiting permission",
  idle: "Turn ended",
};

const PERIOD_MINUTES = { "24h": 1440, "7d": 10080, "30d": 43200, All: Infinity } as const;
type PeriodChip = keyof typeof PERIOD_MINUTES;

interface Filters {
  project: string | null;
  states: Set<string>;
  period: PeriodChip;
  errors: boolean;
  query: string;
}

const NO_FILTERS: Filters = { project: null, states: new Set(), period: "7d", errors: false, query: "" };

/**
 * The board's only navigation path into a single change's own dashboard.
 * Each segment is percent-encoded independently -- App.tsx's parseRunRoute
 * decodes them the same way -- so a URL-significant character in a project
 * key or change name survives the round trip.
 */
function runDetailHref(row: StateBoardRow): string {
  return `#/run/${encodeURIComponent(row.projectKey)}/${encodeURIComponent(row.name)}`;
}

function openRun(row: StateBoardRow) {
  window.location.hash = runDetailHref(row);
}

function isPipelineFix(row: StateBoardRow): boolean {
  return row.name.startsWith("pipeline-fix-");
}

function needsOperator(row: StateBoardRow): boolean {
  return row.activity !== undefined && row.activity !== "working";
}

export function StatePill({ state }: { state: string }) {
  return <span className={`lb-pill ${state}`}>{state}</span>;
}

/** A session's activity badge, how long it has been in that state, and what it waits on. */
export function ActivityBadge({ row, now }: { row: StateBoardRow; now: number }) {
  if (!row.activity) return null;
  return (
    <div className={`lb-act ${row.activity}`}>
      <span className="lb-actb">{ACTIVITY_LABEL[row.activity]}</span>
      {row.activitySince && <span className="lb-muted">{formatRelative(row.activitySince, now).replace(" ago", "")}</span>}
      {row.activityDetail && (
        <span className="lb-actd" title={row.activityDetail}>
          {row.activityDetail}
        </span>
      )}
    </div>
  );
}

function Chip({ on, onClick, title, children }: { on: boolean; onClick: () => void; title?: string; children: ReactNode }) {
  return (
    <button type="button" className="lb-fchip" aria-pressed={on} onClick={onClick} title={title}>
      {children}
    </button>
  );
}

function RunningCard({ row, now }: { row: StateBoardRow; now: number }) {
  const total = row.currentTasks;
  const done = row.doneTasks ?? 0;
  return (
    <div
      className={`lb-card${needsOperator(row) ? " needs" : ""}`}
      data-testid="running-card"
      role="button"
      tabIndex={0}
      aria-label={row.name}
      onClick={() => openRun(row)}
      onKeyDown={(e) => e.key === "Enter" && openRun(row)}
    >
      <div className="lb-top">
        <span className="lb-name">{row.name}</span>
        <StatePill state={row.state} />
      </div>
      <div className="lb-stage">
        <span className="lb-live-dot" />
        <span>{row.currentStage}</span>
        {row.currentStageStartedAt && <span className="lb-muted">· {formatSpan(now - Date.parse(row.currentStageStartedAt))}</span>}
      </div>
      <ActivityBadge row={row} now={now} />
      <div className="lb-prog">
        <span style={{ width: total ? `${(done / total) * 100}%` : "0%" }} />
      </div>
      <div className="lb-meta">
        <span title={row.projectKey}>{projectLabel(row.projectKey)}</span>
        <span className="lb-num">
          tasks {total === undefined ? "—" : `${done}/${total}`} · {formatUsd(row.costUsd)}
        </span>
      </div>
    </div>
  );
}

/** The latest task count, with its growth over the first when the plan grew or shrank. */
function TasksCell({ row }: { row: StateBoardRow }) {
  const { plannedTasks, currentTasks } = row;
  if (plannedTasks === undefined || currentTasks === undefined) return <span className="lb-muted">—</span>;
  const delta = currentTasks - plannedTasks;
  return (
    <>
      {formatInt(currentTasks)}
      {delta !== 0 && <span className="lb-growth">{delta > 0 ? `+${delta}` : `${delta}`}</span>}
    </>
  );
}

function ChangeRow({ row, now, maxCost }: { row: StateBoardRow; now: number; maxCost: number }) {
  return (
    <tr className="lb-row" onClick={() => openRun(row)}>
      <td>
        {row.currentStage !== undefined && (
          <>
            <span className="lb-live-dot" title="A stage is running" />{" "}
          </>
        )}
        <a href={runDetailHref(row)} onClick={(e) => e.stopPropagation()}>
          {row.name}
        </a>
      </td>
      <td className="lb-proj">
        <span title={row.projectKey}>{projectLabel(row.projectKey)}</span>
      </td>
      <td>
        <StatePill state={row.state} />
      </td>
      <td>
        {row.currentStage === undefined ? (
          <span className="lb-muted">—</span>
        ) : (
          <>
            <span className="lb-muted">{row.currentStage}</span>
            {needsOperator(row) && (
              <>
                {" "}
                <span className={`lb-act ${row.activity} inline`}>
                  <span className="lb-actb">{ACTIVITY_LABEL[row.activity!]}</span>
                </span>
              </>
            )}
          </>
        )}
      </td>
      <td className="r lb-num">
        <TasksCell row={row} />
      </td>
      <td className="r" title={new Date(row.updatedAt).toISOString()}>
        {formatRelative(row.updatedAt, now)}
      </td>
      <td className="r lb-num">{formatInt(row.runCount)}</td>
      <td className="r">
        <span className="lb-costcell">
          <span className="lb-num">{formatUsd(row.costUsd)}</span>
          <span className="lb-costbar">
            <span style={{ width: `${maxCost ? ((row.costUsd / maxCost) * 100).toFixed(0) : 0}%` }} />
          </span>
        </span>
      </td>
      <td className="r lb-num">{row.errors ? <span className="lb-err">{formatInt(row.errors)}</span> : <span className="lb-muted">0</span>}</td>
    </tr>
  );
}

export function StateBoard() {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    // A hidden tab stops refetching; showing it again refetches at once.
    const tick = () => document.hidden || setNow(Date.now());
    const id = setInterval(tick, REFRESH_MS);
    document.addEventListener("visibilitychange", tick);
    return () => {
      clearInterval(id);
      document.removeEventListener("visibilitychange", tick);
    };
  }, []);

  // The period is half-open, so it ends a day past now: a change updated at
  // this very instant is still in it. `now` moving is what refetches.
  const fetched = useStatsView<StateBoardRow[]>("state-board", { from: EPOCH, to: new Date(now + DAY_MS), project: undefined });
  // A refresh keeps showing the last answer while the next one loads.
  const last = useRef<StateBoardRow[] | null>(null);
  if (fetched.status === "ready") last.current = fetched.data.rows.filter((r) => !isPipelineFix(r));

  const [filters, setFilters] = useState<Filters>(NO_FILTERS);
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({ FINISHED: true });

  const rows = last.current;
  const heading = (
    <div>
      <h2>Live state board</h2>
      <p className="lb-sub">Every open change across every project. Click a state tile to filter by state; click a row to open the run.</p>
    </div>
  );
  if (rows === null) {
    return (
      <section aria-label="Live state board" className="dashboard lb">
        {heading}
        <p className="lb-sub">{fetched.status === "error" ? `Could not load the board: ${fetched.message}` : "Loading…"}</p>
      </section>
    );
  }

  const set = (patch: Partial<Filters>) => setFilters((f) => ({ ...f, ...patch }));
  const toggleState = (s: string) => {
    const states = new Set(filters.states);
    if (states.has(s)) {
      states.delete(s);
    } else {
      states.add(s);
      setCollapsed((c) => ({ ...c, [s]: false }));
    }
    set({ states });
  };

  const query = filters.query.trim().toLowerCase();
  const base = rows.filter(
    (r) =>
      (!filters.project || r.projectKey === filters.project) &&
      (now - Date.parse(r.updatedAt)) / 60000 <= PERIOD_MINUTES[filters.period] &&
      (!filters.errors || r.errors > 0) &&
      (!query || r.name.toLowerCase().includes(query)),
  );
  const shown = base.filter((r) => filters.states.size === 0 || filters.states.has(r.state));
  const running = rows
    .filter((r) => r.currentStage !== undefined)
    .sort((a, b) => Number(!needsOperator(a)) - Number(!needsOperator(b)));
  const maxCost = Math.max(0, ...rows.map((r) => r.costUsd));
  const projects = [...new Set(rows.map((r) => r.projectKey))];
  const dirty = filters.project || filters.states.size > 0 || filters.period !== "7d" || filters.errors || filters.query;

  return (
    <section aria-label="Live state board" className="dashboard lb">
      {heading}
      <div className="lb-fbar" role="toolbar" aria-label="Filters">
        <span className="lb-fgroup" role="group" aria-label="Project">
          <span className="lb-fl">Project</span>
          <Chip on={!filters.project} onClick={() => set({ project: null })}>
            All
          </Chip>
          {projects.map((p) => (
            <Chip key={p} on={filters.project === p} onClick={() => set({ project: p })} title={p}>
              {projectLabel(p)}
              <span className="lb-c">{rows.filter((r) => r.projectKey === p).length}</span>
            </Chip>
          ))}
        </span>
        <span className="lb-fgroup" role="group" aria-label="Updated">
          <span className="lb-fl">Updated</span>
          {(Object.keys(PERIOD_MINUTES) as PeriodChip[]).map((p) => (
            <Chip key={p} on={filters.period === p} onClick={() => set({ period: p })}>
              {p}
            </Chip>
          ))}
        </span>
        <span className="lb-fgroup">
          <Chip on={filters.errors} onClick={() => set({ errors: !filters.errors })}>
            With errors only
          </Chip>
        </span>
        {dirty && (
          <button type="button" className="lb-fclear" onClick={() => setFilters(NO_FILTERS)}>
            Clear filters
          </button>
        )}
        <input
          type="search"
          placeholder="Search changes"
          aria-label="Search changes"
          value={filters.query}
          onChange={(e) => set({ query: e.target.value })}
        />
      </div>
      <div className="lb-strip">
        {STATES.map((s) => (
          <button key={s} type="button" className={`lb-tile ${s}`} aria-pressed={filters.states.has(s)} onClick={() => toggleState(s)}>
            <span className="lb-lbl">{STATE_LABEL[s]}</span>
            <span className="lb-val lb-num">{base.filter((r) => r.state === s).length}</span>
          </button>
        ))}
        <div className="lb-tile static">
          <span className="lb-lbl">Running now</span>
          <span className="lb-val lb-num">{running.length}</span>
        </div>
        <div className="lb-tile static">
          <span className="lb-lbl">Cost in period</span>
          <span className="lb-val lb-num">{formatUsd(rows.reduce((a, r) => a + r.costUsd, 0))}</span>
        </div>
      </div>
      <section className="lb-panel" aria-label="Running now">
        <div className="lb-panel-h">
          <span>Running now</span>
          <span className="lb-panel-note">auto-refreshes every 5s</span>
        </div>
        <div className="lb-panel-b lb-cards">
          {running.map((r) => (
            <RunningCard key={`${r.projectKey}/${r.name}`} row={r} now={now} />
          ))}
        </div>
      </section>
      <section className="lb-panel" aria-label="Changes">
        <div className="lb-panel-h">
          <span>Changes</span>
        </div>
        <div className="lb-scroll">
          <table className="lb-table">
            <thead>
              <tr>
                <th>Change</th>
                <th>Project</th>
                <th>State</th>
                <th>Current stage</th>
                <th className="r">Tasks</th>
                <th className="r">Updated</th>
                <th className="r">Runs</th>
                <th className="r">Cost</th>
                <th className="r">Errors</th>
              </tr>
            </thead>
            <tbody>
              {shown.length === 0 && (
                <tr>
                  <td colSpan={9} className="lb-muted">
                    No changes match.
                  </td>
                </tr>
              )}
              {STATES.map((s) => {
                const group = shown.filter((r) => r.state === s).sort((a, b) => Date.parse(b.updatedAt) - Date.parse(a.updatedAt));
                if (group.length === 0) return null;
                const isCollapsed = collapsed[s] ?? false;
                const toggle = () => setCollapsed((c) => ({ ...c, [s]: !isCollapsed }));
                return (
                  <Fragment key={s}>
                    <tr
                      className={`lb-group${isCollapsed ? " collapsed" : ""}`}
                      tabIndex={0}
                      aria-expanded={!isCollapsed}
                      onClick={toggle}
                      onKeyDown={(e) => e.key === "Enter" && toggle()}
                    >
                      <td colSpan={9}>
                        <span className="lb-caret">▾</span>
                        {STATE_LABEL[s]}
                        <span className="lb-cnt">{group.length}</span>
                      </td>
                    </tr>
                    {!isCollapsed && group.map((r) => <ChangeRow key={`${r.projectKey}/${r.name}`} row={r} now={now} maxCost={maxCost} />)}
                  </Fragment>
                );
              })}
            </tbody>
          </table>
        </div>
      </section>
    </section>
  );
}
