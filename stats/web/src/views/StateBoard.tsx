// The live state board -- the default route, laid out as the approved
// prototype (spectre/changes/kan-1048-rework-live-board-and-run-view/design.md):
// a chip filter bar, state tiles that filter by state, an Active runs strip of
// one card per unfinished change or change with an open stage run, and the changes grouped by
// state. pipeline-fix-* changes are hidden here, never on the wire.
//
// The board fetches every change ever recorded and applies every filter
// here (design.md's board-client-side-filters): the list holds tens of rows,
// and each chip count must reflect the other filters without a request per
// click. It refetches every five seconds so a session's activity badge
// follows what the session is doing.
import { Fragment, useEffect, useRef, useState, type ReactNode } from "react";
import { fetchProjects, type StateBoardRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { DAY_MS, formatInt, formatRelative, formatSpan, formatUsd } from "../format";
import { phase } from "../lib/phase";
import { taskCount } from "../lib/taskCount";
import { projectLabel } from "../lib/projectLabel";
import { agentsDone, lastAction, mainAction, runningTask } from "../lib/runningTask";
import { effectiveActivity, sessionAlive } from "../lib/sessionAlive";
import { stageLabel } from "../lib/stageLabel";
import { currentRequest, interactiveWaiting, isInteractive, lastEndedRequest } from "../lib/interactive";

const REFRESH_MS = 5000;
const EPOCH = new Date(0);

const STATES = ["IN_PROGRESS", "STARTED", "FINISHED"] as const;
const STATE_LABEL: Record<string, string> = { IN_PROGRESS: "Integrating & review", STARTED: "Planning & implementing", FINISHED: "Finished" };

export const ACTIVITY_LABEL: Record<string, string> = {
  working: "Working",
  question: "Awaiting your answer",
  permission: "Awaiting permission",
  idle: "Turn ended",
  ended: "Session closed",
};

const PERIOD_MINUTES = { "24h": 1440, "7d": 10080, "30d": 43200, All: Infinity } as const;
type PeriodChip = keyof typeof PERIOD_MINUTES;
// The cost tile counts only stage runs started within the chosen window,
// whatever change they belong to.
const PERIOD_COST: Record<PeriodChip, (r: StateBoardRow) => number> = {
  "24h": (r) => r.cost24h,
  "7d": (r) => r.cost7d,
  "30d": (r) => r.cost30d,
  All: (r) => r.costUsd,
};

interface Filters {
  project: string | null;
  states: Set<string>;
  period: PeriodChip;
  query: string;
}

const NO_FILTERS: Filters = { project: null, states: new Set(), period: "7d", query: "" };

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

// Pipeline-fix and self-review sub-runs are not changes of their own: they are
// hidden here, and a self-review-<name> run shows inside <name>'s run view, in
// its Self-review phase, which links to it.
function isPipelineFix(row: StateBoardRow): boolean {
  return row.name.startsWith("pipeline-fix-") || row.name.startsWith("self-review-");
}

function needsOperator(row: StateBoardRow): boolean {
  // An interactive session between requests is at its normal rest, not waiting on the operator.
  if (isInteractive(row) && interactiveWaiting(row)) return false;
  const activity = effectiveActivity(row);
  return activity !== undefined && activity !== "working" && activity !== "waiting";
}

/** The change's phase, coloured by its stored state, which the title names; an unfinished interactive session reads Interactive. */
export function StatePill({ row }: { row: Parameters<typeof phase>[0] & Pick<StateBoardRow, "command"> }) {
  return (
    <span className={`lb-pill ${row.state}`} title={`state ${row.state}`}>
      {isInteractive(row) && row.state !== "FINISHED" ? "Interactive" : phase(row)}
    </span>
  );
}

/** A session's activity badge, how long it has been in that state, and what it waits on. */
export function ActivityBadge({ row, now }: { row: StateBoardRow; now: number }) {
  // "Working" only repeats the live dot beside it; the badge speaks for the other states.
  // So does a "turn ended" that is not one: a refused stop, or a turn waiting on its subagents.
  const activity = effectiveActivity(row);
  if (!activity || activity === "working" || activity === "waiting") return null;
  return (
    <div className={`lb-act ${activity}`}>
      <span className="lb-actb">{ACTIVITY_LABEL[activity]}</span>
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
  if (isInteractive(row) && interactiveWaiting(row)) return <WaitingCard row={row} now={now} />;
  const { done, total } = taskCount(row);
  const review = row.selfReviewTasks;
  const request = isInteractive(row) ? currentRequest(row) : undefined;
  const task = request ? `Request ${request.n}: ${request.title}` : runningTask(row);
  const alive = sessionAlive(row, now);
  // The main session while a stage is open or it works between stages, plus its subagents.
  const subagentCount = (row.subagents ?? []).filter((d) => !d.inline).length;
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
        <StatePill row={row} />
      </div>
      {row.currentStage !== undefined ? (
        <div className="lb-stage">
          <span className="lb-live-dot" />
          {/* One text run, cut to one line -- main's action can be a whole Bash command -- and whole on hover. */}
          <span className="lb-clip" title={mainAction(row) ?? undefined}>
            <span title={row.currentStage}>{stageLabel(row.currentStage)}</span>
            {row.currentStageStartedAt && <span className="lb-muted"> · {formatSpan(now - Date.parse(row.currentStageStartedAt))}</span>}
            {mainAction(row) ? <span className="lb-muted"> · {mainAction(row)}</span> : agentsDone(row) && <span className="lb-muted"> · agents done, main wrapping up</span>}
          </span>
        </div>
      ) : task || alive ? (
        <div className="lb-stage">
          <span className="lb-live-dot" />
          <span>Working between stages</span>
        </div>
      ) : (
        <div className="lb-stage">
          <span title={row.lastStage}>{row.lastStage === undefined ? "No stage yet" : `Paused after: ${stageLabel(row.lastStage)}`}</span>
          {row.lastStageEndedAt && <span className="lb-muted">· {formatSpan(now - Date.parse(row.lastStageEndedAt))} ago</span>}
        </div>
      )}
      {(task ?? lastAction(row)) && (
        <div className="lb-task-now" data-testid="running-task" title={task ?? lastAction(row) ?? ""}>
          {task ?? lastAction(row)}
        </div>
      )}
      <ActivityBadge row={row} now={now} />
      {(subagentCount > 0 || alive) && (
        <div className="lb-muted" data-testid="agent-count">
          {subagentCount > 0 ? `main + ${subagentCount} subagent${subagentCount === 1 ? "" : "s"} running` : "main session running"}
        </div>
      )}
      <div className="lb-prog">
        <span style={{ width: total ? `${(done / total) * 100}%` : "0%" }} />
      </div>
      <div className="lb-meta">
        <span title={row.projectKey}>{projectLabel(row.projectKey)}</span>
        <span className="lb-num">
          {isInteractive(row) ? "requests" : "tasks"} {total === undefined ? "—" : `${done}/${total}`}
          {review && ` · self-review tasks ${review.filter((t) => t.state === "done").length}/${review.length}`} · {formatUsd(row.costUsd)}
        </span>
      </div>
    </div>
  );
}

/** An interactive session between requests: a ring, how long since its last stage ended, and its latest ended request. */
function WaitingCard({ row, now }: { row: StateBoardRow; now: number }) {
  const { done, total } = taskCount(row);
  const last = lastEndedRequest(row);
  return (
    <div
      className={`lb-card${row.activity === "ended" ? "" : " awaiting"}`}
      data-testid="running-card"
      role="button"
      tabIndex={0}
      aria-label={row.name}
      onClick={() => openRun(row)}
      onKeyDown={(e) => e.key === "Enter" && openRun(row)}
    >
      <div className="lb-top">
        <span className="lb-name">{row.name}</span>
        <StatePill row={row} />
      </div>
      <div className="lb-stage">
        <span className="lb-ring" />
        <span>{row.activity === "ended" ? "Session closed" : "Waiting for the next request"}</span>
        {row.lastStageEndedAt && <span className="lb-muted">· {formatSpan(now - Date.parse(row.lastStageEndedAt))}</span>}
      </div>
      {last && (
        <div className="lb-task-now lb-muted" data-testid="running-task">
          Last: Request {last.n} {last.outcome} — {last.title}
        </div>
      )}
      <div className="lb-prog">
        <span style={{ width: total ? `${(done / total) * 100}%` : "0%" }} />
      </div>
      <div className="lb-meta">
        <span title={row.projectKey}>{projectLabel(row.projectKey)}</span>
        <span className="lb-num">
          requests {done}/{total ?? 0} · {formatUsd(row.costUsd)}
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

/** A main · self-review pair of counts, each named in its title. */
export function MainAndSelfReview({ main, selfReview }: { main: number; selfReview: number }) {
  return <span title={`main ${main} · self-review ${selfReview}`}>{`${main} · ${selfReview}`}</span>;
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
        <StatePill row={row} />
      </td>
      <td>
        {row.currentStage === undefined ? (
          <span className="lb-muted">—</span>
        ) : (
          <>
            <span className="lb-muted" title={row.currentStage}>{stageLabel(row.currentStage)}</span>
            {needsOperator(row) && (
              <>
                {" "}
                <span className={`lb-act ${effectiveActivity(row)} inline`}>
                  <span className="lb-actb">{ACTIVITY_LABEL[effectiveActivity(row) ?? ""]}</span>
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
      <td className="r lb-num">
        <MainAndSelfReview main={row.fixRounds} selfReview={row.selfReviewFixRounds} />
      </td>
      <td className="r lb-num">
        <MainAndSelfReview main={row.reviewerRuns} selfReview={row.selfReviewReviewerRuns} />
      </td>
      <td className="r">
        <span className="lb-costcell">
          <span className="lb-num">{formatUsd(row.costUsd)}</span>
          <span className="lb-costbar">
            <span style={{ width: `${maxCost ? ((row.costUsd / maxCost) * 100).toFixed(0) : 0}%` }} />
          </span>
        </span>
      </td>
    </tr>
  );
}

const rowKey = (r: StateBoardRow) => `${r.projectKey}/${r.name}`;

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

  // Only open changes are polled: an empty period still lists every
  // unfinished change, or finished one still running a stage, which the
  // store adds whatever the period. Finished
  // ones load on demand -- the Finished group opened, its tile or a Project chip pressed --
  // then once per Updated chip, inside the widest bounded chip unless All is
  // chosen, and again whenever a polled change drops out of the open set,
  // which is a change finishing. Each period is half-open, so
  // it ends a day past its instant: a change updated then is still in it.
  const [filters, setFilters] = useState<Filters>(NO_FILTERS);
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({ FINISHED: true });
  const wantFinished = useRef(false);
  if (!(collapsed.FINISHED ?? false) || filters.states.has("FINISHED") || filters.project) wantFinished.current = true;
  // Every registered project names a chip before the finished changes load,
  // so one whose changes are all finished is not missing; a failed read leaves the in-period keys.
  const [registered, setRegistered] = useState<string[]>([]);
  useEffect(() => {
    fetchProjects().then(setRegistered, () => {});
  }, []);
  const live = new Date(now + DAY_MS);
  const polled = useStatsView<StateBoardRow[]>("state-board", { from: live, to: live, project: undefined });
  const [finishedAt, setFinishedAt] = useState(now);
  const from = filters.period === "All" ? EPOCH : new Date(finishedAt - PERIOD_MINUTES["30d"] * 60000);
  const finished = useStatsView<StateBoardRow[]>("state-board", { from, to: new Date(finishedAt + DAY_MS), project: undefined }, wantFinished.current);
  // A refresh keeps showing the last answer while the next one loads.
  const lastOpen = useRef<StateBoardRow[] | null>(null);
  const lastFinished = useRef<StateBoardRow[] | null>(null);
  const openNames = useRef<Set<string>>(new Set());
  if (polled.status === "ready") lastOpen.current = polled.data.rows;
  if (finished.status === "ready") lastFinished.current = finished.data.rows;
  useEffect(() => {
    if (polled.status !== "ready") return;
    const names = new Set(polled.data.rows.map(rowKey));
    if ([...openNames.current].some((n) => !names.has(n))) setFinishedAt(now);
    openNames.current = names;
  }, [polled, now]);
  const last = useRef<StateBoardRow[] | null>(null);
  // The open changes render as soon as the quick poll answers; the slower finished read joins them when it lands.
  if (lastOpen.current) {
    const open = new Set(lastOpen.current.map(rowKey));
    last.current = [...lastOpen.current, ...(lastFinished.current ?? []).filter((r) => !open.has(rowKey(r)))].filter((r) => !isPipelineFix(r));
  }
  const fetched = polled.status === "error" ? polled : finished;
  // Until the finished changes load, what counts them reads "—" rather than an open-only figure.
  const finishedLoaded = lastFinished.current !== null;

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
  const inPeriod = rows.filter((r) => (now - Date.parse(r.updatedAt)) / 60000 <= PERIOD_MINUTES[filters.period]);
  const base = inPeriod.filter(
    (r) =>
      (!filters.project || r.projectKey === filters.project) &&
      (!query || r.name.toLowerCase().includes(query)),
  );
  const shown = base.filter((r) => filters.states.size === 0 || filters.states.has(r.state));
  const running = rows
    .filter((r) => r.currentStage !== undefined || r.state !== "FINISHED")
    .sort((a, b) => Number(!needsOperator(a)) - Number(!needsOperator(b)));
  const maxCost = Math.max(0, ...rows.map((r) => r.costUsd));
  const projects = [...new Set((finishedLoaded ? [] : registered).concat(inPeriod.map((r) => r.projectKey), filters.project ?? []))];
  const dirty = filters.project || filters.states.size > 0 || filters.period !== "7d" || filters.query;

  return (
    <section aria-label="Live state board" className="dashboard lb">
      {heading}
      {fetched.status === "error" && (
        <p className="lb-sub" role="alert">
          Could not load the board: {fetched.message}
        </p>
      )}
      <div className="lb-fbar" role="toolbar" aria-label="Filters">
        <span className="lb-fgroup" role="group" aria-label="Project">
          <span className="lb-fl">Project</span>
          <Chip on={!filters.project} onClick={() => set({ project: null })}>
            All
          </Chip>
          {projects.map((p) => (
            <Chip key={p} on={filters.project === p} onClick={() => set({ project: p })} title={p}>
              {projectLabel(p)}
              <span className="lb-c">{finishedLoaded ? inPeriod.filter((r) => r.projectKey === p).length : "—"}</span>
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
            <span className="lb-val lb-num">{s === "FINISHED" && !finishedLoaded ? "—" : base.filter((r) => r.state === s).length}</span>
          </button>
        ))}
        <div className="lb-tile static">
          <span className="lb-lbl">Active runs</span>
          <span className="lb-val lb-num">{running.length}</span>
        </div>
        <div className="lb-tile static">
          <span className="lb-lbl">Cost in period</span>
          <span className="lb-val lb-num">{finishedLoaded ? formatUsd(rows.reduce((a, r) => a + PERIOD_COST[filters.period](r), 0)) : "—"}</span>
        </div>
      </div>
      <section className="lb-panel" aria-label="Active runs">
        <div className="lb-panel-h">
          <span>Active runs</span>
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
                <th className="r">Fix rounds</th>
                <th className="r">Reviewer runs</th>
                <th className="r">Cost</th>
              </tr>
            </thead>
            <tbody>
              {shown.length === 0 && (
                <tr>
                  <td colSpan={10} className="lb-muted">
                    No changes match.
                  </td>
                </tr>
              )}
              {STATES.map((s) => {
                const group = shown.filter((r) => r.state === s).sort((a, b) => Date.parse(b.updatedAt) - Date.parse(a.updatedAt));
                const pending = s === "FINISHED" && !finishedLoaded;
                if (group.length === 0 && !pending) return null;
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
                      <td colSpan={10}>
                        <span className="lb-caret">▾</span>
                        {STATE_LABEL[s]}
                        <span className="lb-cnt">{pending ? "—" : group.length}</span>
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
