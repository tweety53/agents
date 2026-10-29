// RunDetail opens one change -- "#/run/<project>/<change>" -- laid out as the
// approved prototype's run view (spectre/changes/kan-1048-rework-live-board-and-run-view/design.md):
// a back link, a header with the state pill, project, Jira key and last
// update, one KPI strip, a full-width stage timeline, and the session-grouped
// stage runs beside a Cost by stage rail. It carries no filter of any kind
// (board-chip-filters-run-none).
//
// The KPI strip's cost and input-token figures come from useRunDetail's
// cost-per-change aggregate, never from summing the stage runs this
// component renders -- see hooks/useRunDetail.ts's own header comment for
// why that matters. State, last update and session activity come from the
// change's state board row.
import { useEffect, useMemo, useState, type ReactNode } from "react";
import { StageRunTable, groupSessions, runEnd } from "../components/StageRunTable";
import { StageTimeline } from "../components/StageTimeline";
import { useRunDetail } from "../hooks/useRunDetail";
import { formatModel, formatRelative, formatSpan, formatUsd } from "../format";
import { asObject, readCacheReadTokens, readCostUsd, readEffort, readMainModels, readTotalInputTokens } from "../metrics";
import { projectLabel } from "../lib/projectLabel";
import { describeDispatch } from "../lib/describeDispatch";
import { agentsDone, runningTask } from "../lib/runningTask";
import { effectiveActivity, sessionAlive } from "../lib/sessionAlive";
import { stageLabel } from "../lib/stageLabel";
import { ActivityBadge, StatePill } from "./StateBoard";
import { taskCount } from "../lib/taskCount";
import { SR_STEPS, selfReviewNow, srStep, type SrStep } from "../lib/selfReviewNow";
import { STAGE_PHASE, type Phase } from "../lib/phase";
import { mainNow, mainProgress, mainSummary, type ProgressState } from "../lib/mainNow";
import { findChanges, getRunRecord, listSelfReviewFindings, listStageRuns, listStages, type ActiveSubagent, type PlanTask, type RecordDispatch, type SelfReviewFinding, type StageRunDTO } from "../api";

export interface RunDetailProps {
  project: string;
  change: string;
}

function Kpi({ label, value, hint, children }: { label: string; value: ReactNode; hint?: ReactNode; children?: ReactNode }) {
  return (
    <div className="lb-kpi" role="group" aria-label={label}>
      <span className="lb-lbl">{label}</span>
      <span className="lb-kpi-val">{value}</span>
      {hint && <span className="lb-hint">{hint}</span>}
      {children}
    </div>
  );
}

/** The Jira key a change name leads with (`kan-1048-…` → `KAN-1048`), if any. */
const TASK_MARK: Record<PlanTask["state"], string> = { done: "✔", "in progress": "◼", pending: "◻" };

/** " (title)" of plan task id, or empty when the plan does not list it. */
function taskTitle(tasks: PlanTask[] | undefined, id: string): string {
  const t = tasks?.find((t) => String(t.n) === id);
  return t ? ` (${t.title})` : "";
}

type AgentStatus = "running" | "waiting" | "needs-you" | "idle";
const STATUS_LABEL: Record<AgentStatus, string> = {
  running: "running",
  waiting: "waiting for subagents",
  "needs-you": "needs you",
  idle: "turn ended",
};

/** An agent's status as a dot whose look names it -- pulsing green running, breathing amber waiting, blinking orange needs you, hollow grey idle -- and its word. */
function StatusDot({ status }: { status: AgentStatus }) {
  return (
    <span className="lb-status">
      <span className={`lb-dot ${status}`} />
      {STATUS_LABEL[status]}
    </span>
  );
}

/** What a dispatch does, in plain words from its key, with its plan task's title when it has one. */
function inlineWhat(d: ActiveSubagent, tasks: PlanTask[] | undefined): string {
  const what = d.key === undefined && d.taskId ? `task ${d.taskId}` : describeDispatch(d.key, d.role);
  return `${what}${d.taskId ? taskTitle(tasks, d.taskId) : ""}`;
}

/** How many of the costliest stages Cost by stage lists before folding the rest into one line. */
const TOP_STAGES = 10;

/** How many of the latest stage runs the collapsed Stage runs table shows. */
const LATEST_RUNS = 5;

/** The outcomes a dispatch ends on without finishing its work. */
const FAILED_OUTCOMES = new Set(["failed", "blocked", "aborted", "timed-out", "cancelled", "stopped"]);

type Step = ActiveSubagent & { phase: Phase; agentId?: string; endedAt?: string; outcome?: string };

/** Each dispatch as a step of the phase whose stage run was the latest to start at or before it (production dispatches carry no stage run id). */
function phaseSteps(runs: StageRunDTO[], dispatches: RecordDispatch[]): Step[] {
  const byStart = [...runs].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt));
  const out: Step[] = [];
  for (const d of [...dispatches].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt))) {
    const t = Date.parse(d.startedAt);
    const run = byStart.filter((r) => Date.parse(r.startedAt) <= t).pop();
    const phase = run && STAGE_PHASE[run.stage];
    if (!phase) continue;
    out.push({
      phase,
      agentId: d.agentId,
      role: d.role,
      key: d.key || undefined,
      taskId: d.taskId || undefined,
      model: d.model,
      modelVersion: d.metrics?.modelVersion,
      effort: (d.effort && d.effort !== "default" ? d.effort : undefined) ?? d.metrics?.effort,
      startedAt: d.startedAt,
      endedAt: d.endedAt,
      outcome: d.outcome,
    });
  }
  return out;
}

type PhaseStatus = "running" | "done" | "idle";
const PHASE_STATUS_LABEL: Record<PhaseStatus, string> = { running: "running", done: "done", idle: "not started" };

/** Each priced dispatch's own cost, keyed by agent id, from its stage run's `dispatches.<agentId>.cost_usd` (KAN-212). */
function dispatchCosts(runs: StageRunDTO[]): Map<string, number> {
  const out = new Map<string, number>();
  for (const r of runs) {
    const buckets = asObject(asObject(r.metrics)?.dispatches);
    for (const [id, b] of Object.entries(buckets ?? {})) {
      const c = readCostUsd(b);
      if (c !== null) out.set(id, c);
    }
  }
  return out;
}

/** Cost of whatever was measured, null when nothing was -- never a fabricated 0. */
function measuredSum(values: (number | null | undefined)[]): number | null {
  const known = values.filter((v): v is number => typeof v === "number");
  return known.length ? known.reduce((a, v) => a + v, 0) : null;
}

/** One line per role of a phase's dispatches: how many, how many ran on each model · effort (most first), their summed cost, failures and whether any still runs -- costliest first, an unpriced role last. */
function roleSummary(steps: Step[], costs: Map<string, number>) {
  const byRole = new Map<string, Step[]>();
  for (const d of steps) byRole.set(d.role, [...(byRole.get(d.role) ?? []), d]);
  const costOf = (ds: Step[]) => measuredSum(ds.map((d) => (d.agentId ? costs.get(d.agentId) : undefined)));
  return [...byRole].map(([role, ds]) => {
    const byModel = new Map<string, Step[]>();
    for (const d of ds) byModel.set(modelEffortLabel(d), [...(byModel.get(modelEffortLabel(d)) ?? []), d]);
    const models = [...byModel].sort((a, b) => b[1].length - a[1].length);
    return {
      role,
      count: ds.length,
      models: models.map(([m, xs]) => `${m} ×${xs.length} (${((c) => (c === null ? "—" : formatUsd(c)))(costOf(xs))})`).join(", "),
      effortUnknown: ds.some((d) => !knownEffort(d.effort)),
      cost: costOf(ds),
      failed: ds.filter((d) => d.endedAt && d.outcome && FAILED_OUTCOMES.has(d.outcome)).length,
      running: ds.some((d) => !d.endedAt),
    };
  })
    .sort((a, b) => (b.cost ?? -1) - (a.cost ?? -1));
}

/** An effort a dispatch row really records -- "default" is the recorder's placeholder, never an effort. */
const knownEffort = (e: string | undefined) => (e && e !== "default" ? e : undefined);

/** A dispatch's model as family and version, with its effort when one is known ("opus 5.5 · low"). */
function modelEffortLabel(d: Step): string {
  const model = d.modelVersion || d.model;
  if (!model) return "unknown";
  const e = knownEffort(d.effort);
  return `${formatModel(model)}${e ? ` · ${e}` : ""}`;
}

/**
 * The cost lines under a panel's roles, so the lines add up to its header:
 * "unattributed" for priced dispatch buckets no recorded dispatch of the
 * panel claims, then "main session" for the runs' cost no bucket carries --
 * the main session's own turns. Null where nothing is left.
 */
function residualCosts(runs: StageRunDTO[], steps: Step[], costs: Map<string, number>): { unattributed: number | null; main: number | null } {
  const header = measuredSum(runs.map((r) => readCostUsd(r.metrics)));
  const claimed = new Set(steps.map((d) => d.agentId).filter((a): a is string => !!a));
  let buckets = 0;
  let unattributed = 0;
  for (const [id, c] of costs) {
    buckets += c;
    if (!claimed.has(id)) unattributed += c;
  }
  const keep = (v: number) => (Math.abs(v) >= 0.005 ? v : null);
  return { unattributed: keep(unattributed), main: header === null ? null : keep(header - buckets) };
}

/** Self-review's one-line summary: its plan's done/total, its findings, and its highest fix round when one ran. */
function selfReviewSummary(tasks: PlanTask[] | undefined, findings: number, steps: Step[]): string {
  const round = fixRound(steps);
  return [
    tasks && `plan ${tasks.filter((t) => t.state === "done").length}/${tasks.length}`,
    `findings ${findings}`,
    round !== null && `fixes round ${round}`,
  ]
    .filter(Boolean)
    .join(" · ");
}

/** The highest fix round a self-review dispatch key names (-fix-<n> or -fix-v<n>), null when none does. */
function fixRound(steps: Step[]): number | null {
  const rounds = steps.map((d) => /-fix-v?(\d+)$/.exec(d.key ?? "")).filter((m) => m !== null).map((m) => Number(m[1]));
  return rounds.length ? Math.max(...rounds) : null;
}

/** One step of a progress line: its label and whether it is done, current or still to come. */
type Progress = { label: string; state: ProgressState }[];

/**
 * One of the run view's two work panels, Main or Self-review: a header with
 * its status, active time, cost and summary; a Now line and its progress on
 * one line; a line per role of its dispatches; then its plan, the task being
 * worked on marked. A collapsible panel shows its header alone until opened,
 * open by default while it runs.
 */
function WorkPanel({
  title,
  runs,
  steps,
  now,
  extra,
  nowText,
  progress,
  summary,
  tasks,
  currentTask,
  note,
  collapsible,
  defaultOpen,
}: {
  title: string;
  runs: StageRunDTO[];
  steps: Step[];
  now: number;
  extra?: ReactNode;
  nowText: string;
  progress: Progress;
  summary: string;
  tasks?: PlanTask[];
  currentTask?: number;
  /** A line under the Now line: Main's view of the self-review. */
  note?: ReactNode;
  collapsible?: boolean;
  /** Whether a collapsible panel starts open; it opens while it runs by default. */
  defaultOpen?: boolean;
}) {
  const status = workStatus(runs);
  // Null follows the default -- open while running -- until the operator toggles it.
  const [toggled, setToggled] = useState<boolean | null>(null);
  const open = !collapsible || (toggled ?? defaultOpen ?? status === "running");
  const active = runs.reduce((a, r) => a + runEnd(r, now) - Date.parse(r.startedAt), 0);
  const cost = measuredSum(runs.map((r) => readCostUsd(r.metrics)));
  // Costs are this panel's own runs' priced buckets, so its lines add up to its header.
  const costs = dispatchCosts(runs);
  const roles = roleSummary(steps, costs);
  const rest = residualCosts(runs, steps, costs);
  const slug = title.toLowerCase();
  const tasksTitle = title === "Main" ? "Tasks" : `${title} tasks`;
  return (
    <div className="lb-work-row">
    <section className={`lb-panel lb-phase ${status}`} aria-label={title}>
      <div className="lb-panel-h">
        <span className="lb-phase-name">
          <span className={`lb-dot ${status}`} />
          {title}
          <span className="lb-phase-st">
            {" · "}
            {PHASE_STATUS_LABEL[status]}
            {status !== "idle" && (
              <>
                {" · "}
                <span className="lb-num">{formatSpan(active)}</span>
                {" · "}
                <span className="lb-num lb-phase-cost">{cost === null ? "—" : formatUsd(cost)}</span>
              </>
            )}
            {summary && (
              <>
                {" · "}
                <span data-testid={`${slug}-summary`}>{summary}</span>
              </>
            )}
          </span>
        </span>
        <span className="lb-panel-note lb-phase-meta">
          {extra}
          {collapsible && (
            <button type="button" className="lb-fchip" aria-expanded={open} onClick={() => setToggled(!open)}>
              {open ? "Hide" : "Show"}
            </button>
          )}
        </span>
      </div>
      {open && (
        <div className="lb-phase-main">
          <div className="lb-phase-sum">
            <div className="lb-sr-line">
            <span className="lb-sr-now" data-testid={`${slug}-now`}>
              <strong>Now:</strong> {nowText}
            </span>
            <span className="lb-sr-prog" data-testid={`${slug}-progress`}>
              {progress.map((p, i) => (
                <span key={p.label} className={p.state}>
                  {i > 0 && <span className="lb-muted"> → </span>}
                  {p.label}
                  {p.state === "done" ? " ✓" : p.state === "current" ? " ●" : ""}
                </span>
              ))}
            </span>
            </div>
            {note}
          </div>
          {status !== "idle" && (
            <ul className="lb-roles">
              {roles.map((r) => (
                <li key={r.role} data-testid="phase-role">
                  <span className="lb-role">
                    {r.role} <span className="lb-muted">×{r.count}</span>
                  </span>
                  <span className="lb-role-models" title={r.effortUnknown ? `${r.models} -- effort unknown for some dispatches` : r.models}>
                    {r.models}
                  </span>
                  <span className="lb-num r">{r.cost === null ? <span className="lb-muted">—</span> : formatUsd(r.cost)}</span>
                  <span className="lb-role-flags">
                    {r.failed > 0 && <span className="lb-err">{r.failed} failed</span>}
                    {r.running && (
                      <span className="lb-status">
                        <span className="lb-dot running" />
                        running
                      </span>
                    )}
                  </span>
                </li>
              ))}
              {roles.length === 0 && <li className="lb-muted">No dispatches yet.</li>}
              {rest.unattributed !== null && (
                <li data-testid="phase-role" className="lb-muted" title="priced dispatches with no recorded dispatch row">
                  <span className="lb-role">unattributed</span>
                  <span />
                  <span className="lb-num r">{formatUsd(rest.unattributed)}</span>
                  <span />
                </li>
              )}
              {rest.main !== null && (
                <li data-testid="phase-role" className="lb-muted" title="the main session's own turns: dispatching, reading reports, acting on them">
                  <span className="lb-role">main session</span>
                  <span />
                  <span className="lb-num r">{formatUsd(rest.main)}</span>
                  <span />
                </li>
              )}
            </ul>
          )}
        </div>
      )}
    </section>
    {tasks && (
      // Its own card beside the panel; collapsed with it to its header.
      <section className={`lb-panel lb-tasks-panel${open ? "" : " collapsed"}`} aria-label={tasksTitle}>
        <div className="lb-panel-h">
          <span>{tasksTitle}</span>
          <span className="lb-panel-note">{doneOf(tasks)}</span>
        </div>
        {open && (
          <div className="lb-tasks-scroll">
            <TaskList tasks={tasks} current={currentTask} />
          </div>
        )}
      </section>
    )}
    </div>
  );
}

/** A panel's status from its stage runs: idle with none, running while one is open, done after. */
function workStatus(runs: StageRunDTO[]): PhaseStatus {
  return runs.length === 0 ? "idle" : runs.some((r) => !r.endedAt) ? "running" : "done";
}

/** A plan's tasks, each marked done, in progress or pending. */
function TaskList({ tasks, current }: { tasks: PlanTask[]; current?: number }) {
  return (
    <ul className="lb-tasks">
      {tasks.map((t) =>
        t.n === current ? (
          <li key={t.n} className="current" aria-current="step" title={`${t.n}. ${t.title}`}>
            <span className="lb-task-mark">●</span> {t.n}. {t.title}
          </li>
        ) : (
          <li key={t.n} className={t.state.replace(" ", "-")} title={`${t.n}. ${t.title}`}>
            <span className="lb-task-mark">{TASK_MARK[t.state]}</span> {t.n}. {t.title}
          </li>
        ),
      )}
    </ul>
  );
}

const doneOf = (tasks: PlanTask[]) => `${tasks.filter((t) => t.state === "done").length}/${tasks.length}`;

/** A self-review's Now line, progress over its steps, summary and current task, from its dispatches, plan and findings. */
function selfReviewView(steps: Step[], tasks: PlanTask[] | undefined, findings: number, stageOpen: boolean, started: boolean, now: number) {
  const sr = selfReviewNow(steps, tasks, stageOpen, now);
  const at = sr.current ? SR_STEPS.indexOf(sr.current) : -1;
  return {
    nowText: started ? sr.text : "Not started",
    progress: SR_STEPS.map((st, i) => ({
      label: st === "fix" && tasks ? `fix ${tasks.filter((t) => t.state === "done").length}/${tasks.length}` : st,
      state: (!started ? "todo" : sr.done || i < at ? "done" : i === at ? "current" : "todo") as ProgressState,
    })),
    summary: selfReviewSummary(tasks, findings, steps),
    currentTask: sr.taskN,
  };
}

/** What a self-review step's dispatches cost under, in Cost by stage. */
const SR_COST_GROUP: Record<SrStep, string> = {
  review: "Review pass",
  plan: "Planning fixes",
  fix: "Fixing the plan",
  "re-review": "Review rounds",
  verify: "Verify fixes",
  land: "Landing and filing",
  file: "Landing and filing",
};

/**
 * A self-review run's cost split by what its stage runs did: each recorded
 * dispatch's cost under its step (a numbered fix round under "Fix rounds"),
 * the rest of a run -- its main session's turns and any earlier dispatch's
 * tail -- under the run's first own dispatch's step, a run with none under
 * landing and filing.
 */
function selfReviewCosts(runs: StageRunDTO[], records: RecordDispatch[]): Map<string, number> {
  const byAgent = new Map(records.filter((d) => d.agentId).map((d) => [d.agentId!, d]));
  const group = (rec: RecordDispatch) => {
    const st = srStep(rec.key);
    return !st ? "Other" : st.step === "fix" && st.round !== undefined ? "Fix rounds" : SR_COST_GROUP[st.step];
  };
  const out = new Map<string, number>();
  const add = (g: string, v: number) => out.set(g, (out.get(g) ?? 0) + v);
  for (const r of runs) {
    let left = readCostUsd(r.metrics) ?? 0;
    const own = Object.entries(asObject(asObject(r.metrics)?.dispatches) ?? {})
      .map(([id, b]) => ({ rec: byAgent.get(id), cost: readCostUsd(b) ?? 0 }))
      .filter((d): d is { rec: RecordDispatch; cost: number } => !!d.rec && Date.parse(d.rec.startedAt) >= Date.parse(r.startedAt) - 1000)
      .sort((a, b) => Date.parse(a.rec.startedAt) - Date.parse(b.rec.startedAt));
    for (const d of own) {
      add(group(d.rec), d.cost);
      left -= d.cost;
    }
    add(own.length ? group(own[0].rec) : "Landing and filing", left);
  }
  return out;
}

function jiraKey(change: string): string | null {
  const m = /^([a-z]+-\d+)/i.exec(change);
  return m ? m[1].toUpperCase() : null;
}

/** A failed run counts as retried when a later run of the same stage started after it. */
function retried(runs: StageRunDTO[], failed: StageRunDTO): boolean {
  return runs.some((r) => r.stage === failed.stage && Date.parse(r.startedAt) > Date.parse(failed.startedAt));
}

export function RunDetail({ project, change }: RunDetailProps) {
  const state = useRunDetail(project, change);
  const now = Date.now();
  const sessions = useMemo(() => (state.status === "ready" ? groupSessions(state.stageRuns) : []), [state]);

  const key = jiraKey(change);
  // A self-review sub-run's own page: its work is one panel, Main, with no sub-run of its own.
  const isSub = change.startsWith("self-review-");
  const board = state.status === "ready" ? state.board : undefined;

  // The stages an unfinished change has still to run: its command's stages after the furthest one it ran.
  const runs0 = state.status === "ready" ? state.stageRuns : [];
  const latest = [...runs0].sort((a, b) => Date.parse(b.startedAt) - Date.parse(a.startedAt))[0];
  const command = latest && board?.state !== "FINISHED" ? latest.command : undefined;
  const [order, setOrder] = useState<string[]>([]);
  useEffect(() => {
    if (!command) return;
    let live = true;
    listStages(command)
      .then((r) => live && setOrder(r.stages))
      .catch(() => live && setOrder([]));
    return () => {
      live = false;
    };
  }, [command]);
  const ran = new Set(runs0.map((r) => r.stage));
  const furthest = Math.max(-1, ...order.map((s, i) => (ran.has(s) ? i : -1)));
  const upcoming = command ? order.slice(furthest + 1).filter((s) => !ran.has(s)) : [];

  // The phase panels' dispatches and the self-review findings, refetched on the board's cadence.
  const [dispatches, setDispatches] = useState<RecordDispatch[]>([]);
  const [findings, setFindings] = useState<SelfReviewFinding[]>([]);
  useEffect(() => {
    let live = true;
    const load = () => {
      getRunRecord(project, change)
        .then((r) => live && setDispatches(r?.dispatches ?? []))
        .catch(() => live && setDispatches([]));
      listSelfReviewFindings(project, change)
        .then((r) => live && setFindings(r ?? []))
        .catch(() => live && setFindings([]));
    };
    load();
    const id = setInterval(load, 5000);
    return () => {
      live = false;
      clearInterval(id);
    };
  }, [project, change]);
  const [allRuns, setAllRuns] = useState(false);
  const [timelineOpen, setTimelineOpen] = useState(false);
  // The self-review sub-run, change self-review-<name> in whichever project hosts the tooling: its
  // stage runs, dispatches and findings belong to this change's Self-review phase.
  const [sub, setSub] = useState<{ project: string; name: string; runs: StageRunDTO[]; dispatches: RecordDispatch[]; findings: SelfReviewFinding[] } | null>(null);
  // A sub-run's findings are recorded under the change it reviews, in whichever project hosts that.
  const [parentFindings, setParentFindings] = useState<SelfReviewFinding[]>([]);
  useEffect(() => {
    if (!isSub) return;
    let live = true;
    findChanges(change.slice("self-review-".length))
      .then((r) => {
        const c = r?.records?.[0];
        return c ? listSelfReviewFindings(c.projectKey, c.name) : [];
      })
      .then((f) => live && setParentFindings(f ?? []))
      .catch(() => live && setParentFindings([]));
    return () => {
      live = false;
    };
  }, [change, isSub]);
  useEffect(() => {
    if (isSub) return setSub(null);
    let live = true;
    const load = () =>
      findChanges(`self-review-${change}`)
        .then(async (r) => {
          const c = r?.records?.[0];
          if (!c) return live && setSub(null);
          const [rec, runs, found] = await Promise.all([
            getRunRecord(c.projectKey, c.name).catch(() => ({ dispatches: [] })),
            listStageRuns({ filters: { project: c.projectKey, name: c.name }, limit: 500 }).catch(() => ({ total: 0, stageRuns: [] })),
            listSelfReviewFindings(c.projectKey, c.name).catch(() => []),
          ]);
          if (live) setSub({ project: c.projectKey, name: c.name, runs: runs.stageRuns, dispatches: rec.dispatches ?? [], findings: found ?? [] });
        })
        .catch(() => live && setSub(null));
    load();
    const id = setInterval(load, 5000);
    return () => {
      live = false;
      clearInterval(id);
    };
  }, [change, isSub]);
  const phaseRuns = useMemo(() => [...runs0, ...(sub?.runs ?? [])], [runs0, sub]);
  const steps = useMemo(
    () => [...phaseSteps(runs0, dispatches), ...phaseSteps(sub?.runs ?? [], sub?.dispatches ?? [])].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt)),
    [runs0, dispatches, sub],
  );
  const allFindings = [...findings, ...(sub?.findings ?? [])];
  const srSteps = steps.filter((x) => x.phase === "Self-review");
  const srRuns = phaseRuns.filter((r) => STAGE_PHASE[r.stage] === "Self-review");
  const mainSteps = steps.filter((x) => x.phase !== "Self-review");
  const mainRuns = runs0.filter((r) => STAGE_PHASE[r.stage] !== "Self-review");
  const srStatus = workStatus(srRuns);
  const srStarted = srRuns.length > 0 || board?.currentStage === "flow.self-review";
  // The task Main works on: a running dispatch's, else the plan's in-progress one.
  const mainCurrentTask = Number(mainSteps.filter((d) => !d.endedAt && d.taskId).pop()?.taskId ?? board?.tasks?.find((t) => t.state === "in progress")?.n) || undefined;
  const srOpen = board?.currentStage === "flow.self-review" || srRuns.some((r) => !r.endedAt);
  const srView = isSub
    ? selfReviewView(steps, board?.tasks, findings.length + parentFindings.length, srOpen, runs0.length > 0, now)
    : selfReviewView(srSteps, board?.selfReviewTasks, allFindings.length, srOpen, srStarted, now);
  const srCost = measuredSum(srRuns.map((r) => readCostUsd(r.metrics)));

  return (
    <section aria-label={`Run detail: ${project}/${change}`} className="dashboard lb lb-run">
      <div className="lb-crumb">
        <a href="#/state-board">← Live state board</a>
      </div>
      <div className="lb-runhead">
        <div className="lb-title">
          <h2>{change}</h2>
          {board && <StatePill row={board} />}
        </div>
        {/* The header names the project (kan-183), it does not key it -- the
            section's own aria-label and useRunDetail's request keep the full
            key, since that is the route's identity. */}
        <p className="lb-sub">
          Project {projectLabel(project)}
          {key && ` · ${key}`}
          {board && ` · updated ${formatRelative(board.updatedAt, now)}`}
        </p>
      </div>

      {state.status === "loading" && <p role="status">Loading…</p>}
      {state.status === "error" && (
        <p role="alert" className="view-error">
          Failed to load: {state.message}
        </p>
      )}
      {state.status === "ready" &&
        (() => {
          const runs = state.stageRuns;
          const { summary } = state;
          const openRun = runs.filter((r) => !r.endedAt).sort((a, b) => Date.parse(b.startedAt) - Date.parse(a.startedAt))[0];
          const t0 = runs.length ? Math.min(...runs.map((r) => Date.parse(r.startedAt))) : now;
          const t1 = runs.length ? Math.max(...runs.map((r) => runEnd(r, now))) : now;
          const active = runs.reduce((a, r) => a + runEnd(r, now) - Date.parse(r.startedAt), 0);
          const measured = runs.filter((r) => readTotalInputTokens(r.metrics) !== null);
          const input = measured.reduce((a, r) => a + (readTotalInputTokens(r.metrics) ?? 0), 0);
          const cached = runs.reduce((a, r) => a + (readCacheReadTokens(r.metrics) ?? 0), 0);
          // Subagents are the non-inline dispatches; the main session leads as
          // "main" while a stage is open, on its inline dispatch when it runs one.
          const subagents = (board?.subagents ?? []).filter((d) => !d.inline);
          const inline = board?.subagents?.find((d) => d.inline);
          // Between stages the main session still counts while it works or waits on you.
          const lastRun = [...runs].sort((a, b) => Date.parse(b.startedAt) - Date.parse(a.startedAt))[0];
          const betweenStages = !openRun && lastRun && board !== undefined && (sessionAlive(board, now) || board.activity === "question" || board.activity === "permission");
          const mainRun = openRun ?? (betweenStages ? lastRun : undefined);
          const mainAgent = mainRun && {
            what: inline ? inlineWhat(inline, board?.tasks) : openRun ? `${stageLabel(openRun.stage)}${board && agentsDone(board) ? ", wrapping up after its agents" : ""}` : (board?.activity === "working" && board.activityDetail ? board.activityDetail : `between stages, after ${stageLabel(mainRun.stage)}`),
            model: inline
              ? formatModel(inline.modelVersion ?? inline.model)
              : (readMainModels(mainRun.metrics)?.map(formatModel).join(", ") ?? "—"),
            effort: readEffort(mainRun.metrics) ?? undefined,
            since: inline?.startedAt ?? (openRun ? openRun.startedAt : (mainRun.endedAt ?? mainRun.startedAt)),
          };
          const nowTask = board ? runningTask(board) : null;
          const count = board && taskCount(board);
          const agentCount = subagents.length + (mainAgent ? 1 : 0);
          const mainStatus: AgentStatus =
            board?.activity === "question" || board?.activity === "permission"
              ? "needs-you"
              : board !== undefined && effectiveActivity(board) === "idle"
                ? "idle"
                : !inline && subagents.length > 0
                  ? "waiting"
                  : "running";
          const failed = runs.filter((r) => r.outcome === "failed");
          // The self-review sub-run is this change's work, recorded under its own change: Total carries it once.
          const subCost = measuredSum((sub?.runs ?? []).map((r) => readCostUsd(r.metrics)));
          const totalCost = measuredSum([summary.totalCostUsd, subCost]);
          // A sub-run's every run is Self-review, so its rail splits by what each run did.
          const byStage = isSub ? selfReviewCosts(runs, dispatches) : new Map<string, number>();
          if (!isSub) for (const r of runs) byStage.set(r.stage, (byStage.get(r.stage) ?? 0) + (readCostUsd(r.metrics) ?? 0));
          // A stage that cost nothing, or was never measured, is neither listed nor counted in "N more".
          const allStageCost = [...byStage].filter(([, v]) => v > 0).sort((a, b) => b[1] - a[1]);
          const stageCost = allStageCost.slice(0, TOP_STAGES);
          const restStages = allStageCost.slice(TOP_STAGES);
          const maxStage = Math.max(0, ...stageCost.map(([, v]) => v));
          const sessionOf = (run: StageRunDTO) => sessions.find((s) => s.runs.includes(run))?.label ?? "";

          return (
            <>
              <section className="lb-panel lb-kpis" aria-label="Run totals">
                <Kpi
                  label="Now running"
                  value={
                    openRun ? (
                      <>
                        <span className="lb-live-dot" /> <span title={openRun.stage}>{stageLabel(openRun.stage)}</span>
                      </>
                    ) : betweenStages ? (
                      <>
                        <span className="lb-live-dot" /> Working between stages
                      </>
                    ) : (
                      <span className="lb-muted">—</span>
                    )
                  }
                  hint={
                    <>
                      {openRun &&
                        `for ${formatSpan(now - Date.parse(openRun.startedAt))}${
                          count?.total !== undefined ? ` · tasks ${count.done}/${count.total}` : ""
                        }`}
                      {nowTask && <div className="lb-task-now">{nowTask}</div>}
                      {board && agentsDone(board) && <div className="lb-task-now">Agents done, main is wrapping up the stage</div>}
                    </>
                  }
                >
                  {(openRun || betweenStages) && board && <ActivityBadge row={board} now={now} />}
                </Kpi>
                <Kpi
                  label="Total cost"
                  value={<span className="lb-num">{totalCost === null ? "—" : formatUsd(totalCost)}</span>}
                  hint={`${summary.runCount} stage runs, ${summary.measuredRuns} measured${subCost !== null ? ` · incl. self-review sub-run ${formatUsd(subCost)}` : ""}`}
                />
                <Kpi label="Wall clock" value={<span className="lb-num">{formatSpan(t1 - t0)}</span>} hint={`active ${formatSpan(active)}`} />
                <Kpi
                  label="Input tokens"
                  value={<span className="lb-num">{measured.length === 0 ? "—" : `${(input / 1e6).toFixed(2)}M`}</span>}
                  hint={input > 0 ? `cache reads ${Math.round((cached / input) * 100)}%` : undefined}
                />
                <Kpi
                  label="Failed runs"
                  value={<span className={`lb-num${failed.length ? " lb-err" : ""}`}>{failed.length}</span>}
                  hint={`${failed.filter((f) => retried(runs, f)).length} retried`}
                />
              </section>

              {agentCount > 0 && (
                  <section className="lb-panel" aria-label="Agents">
                    <div className="lb-panel-h">
                      <span>Agents running</span>
                      <span className="lb-panel-note">{agentCount}</span>
                    </div>
                    <table className="lb-table lb-agents">
                      <thead>
                        <tr>
                          <th>Agent</th>
                          <th>Status</th>
                          <th>Doing</th>
                          <th>Model · effort</th>
                          <th className="r">For</th>
                        </tr>
                      </thead>
                      <tbody>
                        {mainAgent && (
                          <tr className="main">
                            <td>
                              Main
                            </td>
                            <td>
                              <StatusDot status={mainStatus} />
                            </td>
                            <td>{mainAgent.what}</td>
                            <td>
                              {mainAgent.model}
                              {mainAgent.effort && ` · ${mainAgent.effort}`}
                            </td>
                            <td className="r lb-num">{formatSpan(now - Date.parse(mainAgent.since))}</td>
                          </tr>
                        )}
                        {subagents.map((d, i) => (
                          <tr key={`${d.key ?? d.role}-${i}`} className="sub">
                            <td>
                              <span className="lb-agent-branch">⎿</span> {d.role}
                            </td>
                            <td>
                              <StatusDot status="running" />
                            </td>
                            <td>
                              {inlineWhat(d, board?.tasks)}
                            </td>
                            <td>
                              {formatModel(d.modelVersion ?? d.model)}
                              {knownEffort(d.effort) && ` · ${d.effort}`}
                            </td>
                            <td className="r lb-num">{formatSpan(now - Date.parse(d.startedAt))}</td>
                          </tr>
                        ))}
                        {agentCount === 0 && (
                          <tr>
                            <td colSpan={5} className="lb-muted">
                              None right now.
                            </td>
                          </tr>
                        )}
                      </tbody>
                    </table>
                  </section>
              )}

              {isSub ? (
                <WorkPanel title="Main" runs={runs0} steps={steps} now={now} tasks={board?.tasks} {...srView} />
              ) : (
                <>
                  <WorkPanel
                    title="Main"
                    runs={mainRuns}
                    steps={mainSteps}
                    now={now}
                    nowText={mainNow(mainRuns, mainSteps, board?.tasks, board?.state === "FINISHED", now)}
                    progress={mainProgress(mainRuns).map((x) => ({ label: x.phase.toLowerCase(), state: x.state }))}
                    summary={mainSummary(board?.tasks, mainSteps)}
                    tasks={board?.tasks}
                    currentTask={mainCurrentTask}
                    note={
                      srStarted && (
                        <div data-testid="main-self-review">
                          <strong>Self-review:</strong> {PHASE_STATUS_LABEL[srStatus]} · {srView.summary}
                          {srCost !== null && ` · ${formatUsd(srCost)}`}
                        </div>
                      )
                    }
                  />
                  <WorkPanel
                    title="Self-review"
                    runs={srRuns}
                    steps={srSteps}
                    now={now}
                    collapsible
                    defaultOpen={srOpen}
                    extra={sub && <a href={`#/run/${encodeURIComponent(sub.project)}/${encodeURIComponent(sub.name)}`}>sub-run</a>}
                    tasks={board?.selfReviewTasks}
                    {...srView}
                  />
                </>
              )}

              <section className="lb-panel" aria-label="Timeline">
                <div className="lb-panel-h">
                  <span>Timeline</span>
                  <span className="lb-legend">
                    <span>
                      <i className="lb-sw completed" />
                      completed
                    </span>
                    <span>
                      <i className="lb-sw failed" />
                      failed
                    </span>
                    <span>
                      <i className="lb-sw open" />
                      running
                    </span>
                    <button type="button" className="lb-fchip" aria-expanded={timelineOpen} onClick={() => setTimelineOpen(!timelineOpen)}>
                      {timelineOpen ? "Hide timeline" : "Show timeline"}
                    </button>
                  </span>
                </div>
                {timelineOpen && (
                  <div className="lb-panel-b lb-scroll">
                    <StageTimeline stageRuns={runs} sessionOf={sessionOf} upcoming={upcoming} />
                  </div>
                )}
              </section>

              <div className="lb-split">
                <section className="lb-panel" aria-label="Stage runs">
                  <div className="lb-panel-h">
                    <span>Stage runs</span>
                    <span className="lb-panel-note">
                      click a row for its dispatches
                      {runs.length > LATEST_RUNS && (
                        <>
                          {" · "}
                          <button type="button" className="lb-fchip" onClick={() => setAllRuns(!allRuns)}>
                            {allRuns ? `Show latest ${LATEST_RUNS}` : `Show all ${runs.length}`}
                          </button>
                        </>
                      )}
                    </span>
                  </div>
                  <StageRunTable
                    stageRuns={allRuns ? runs : [...runs].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt)).slice(-LATEST_RUNS)}
                    truncated={state.truncated}
                    totalStageRuns={state.totalStageRuns}
                    records={dispatches}
                  />
                </section>
                <section className="lb-panel" aria-label="Cost by stage">
                  <div className="lb-panel-h">
                    <span>Cost by stage</span>
                  </div>
                  <div className="lb-panel-b">
                    {stageCost.map(([stage, cost]) => (
                      <div key={stage} className="lb-hbar">
                        <span className="lb-n" title={stage}>
                          <span title={stage}>{stageLabel(stage)}</span>
                        </span>
                        <span className="lb-t">
                          <span style={{ width: `${maxStage ? (cost / maxStage) * 100 : 0}%` }} />
                        </span>
                        <span className="r lb-num">{formatUsd(cost)}</span>
                      </div>
                    ))}
                    {restStages.length > 0 && (
                      <div className="lb-hbar lb-muted" data-testid="stage-cost-rest" title={restStages.map(([s]) => stageLabel(s)).join(", ")}>
                        <span className="lb-n">{restStages.length} more</span>
                        <span />
                        <span className="r lb-num">{formatUsd(restStages.reduce((a, [, v]) => a + v, 0))}</span>
                      </div>
                    )}
                  </div>
                </section>
              </div>
            </>
          );
        })()}
    </section>
  );
}
