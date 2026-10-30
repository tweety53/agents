// RunDetail opens one change -- "#/run/<project>/<change>" -- laid out as the
// approved prototype's run view (spectre/changes/kan-1048-rework-live-board-and-run-view/design.md):
// a back link, a header with the state pill, project, Jira key and last
// update, one KPI strip, a full-width stage timeline, and the session-grouped
// stage runs beside a Cost by stage rail. It carries no filter of any kind
// (board-chip-filters-run-none).
//
// A /flow-interactive session's view (KAN-1061) swaps the Tasks slot for its
// Requests, Main's phases for the followed request's steps, and leads Stage
// runs with a Request column. Its work area is the prototype's 3fr/2fr grid:
// Agents, Main, the timeline and the latest six stage runs left, Requests and
// Cost by stage right; each agent's status is its dot alone.
//
// The KPI strip's cost and input-token figures come from useRunDetail's
// cost-per-change aggregate, never from summing the stage runs this
// component renders -- see hooks/useRunDetail.ts's own header comment for
// why that matters. State, last update and session activity come from the
// change's state board row.
import { Fragment, useEffect, useMemo, useState, type ReactNode } from "react";
import { StageRunTable, groupSessions, runEnd } from "../components/StageRunTable";
import { StageTimeline } from "../components/StageTimeline";
import { useRunDetail } from "../hooks/useRunDetail";
import { formatModel, formatRelative, formatSpan, formatUsd } from "../format";
import { asObject, readCacheReadTokens, readCostUsd, readEffort, readMainModels, readTotalInputTokens } from "../metrics";
import { projectLabel } from "../lib/projectLabel";
import { describeDispatch } from "../lib/describeDispatch";
import { DecisionsPanel, type Decision } from "../components/DecisionsPanel";
import { agentsDone, dispatchTasks, mainAction, runningTask } from "../lib/runningTask";
import { effectiveActivity, sessionAlive } from "../lib/sessionAlive";
import { stageLabel } from "../lib/stageLabel";
import { ActivityBadge, StatePill } from "./StateBoard";
import { taskCount } from "../lib/taskCount";
import { SR_STEPS, selfReviewNow, srStep, type SrStep } from "../lib/selfReviewNow";
import { STAGE_PHASE, type Phase } from "../lib/phase";
import { mainNow, mainProgress, mainSummary, type ProgressState } from "../lib/mainNow";
import { currentRequest, focusRequest, interactiveWaiting, isInteractive, requestStart, requestSteps } from "../lib/interactive";
import { findChanges, getDecisions, getRunRecord, listSelfReviewFindings, listStageRuns, listStages, type ActiveSubagent, type PlanTask, type RecordDispatch, type SelfReviewFinding, type StageRunDTO } from "../api";

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

/** Findings by severity, Critical first and red while any is open to read. */
function Cim({ c, i, m }: { c: number; i: number; m: number }) {
  return (
    <span className="lb-num lb-cim" title={`${c} critical · ${i} important · ${m} minor`}>
      <span className={c ? "lb-err" : undefined}>C {c}</span> · I {i} · M {m}
    </span>
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

/** An agent's status as a dot whose look names it -- pulsing green running, breathing amber waiting, blinking orange needs you, hollow grey idle -- and its word; bare, the dot alone with its word as its accessible name. */
function StatusDot({ status, bare }: { status: AgentStatus; bare?: boolean }) {
  if (bare) return <span className={`lb-dot ${status}`} role="img" aria-label={STATUS_LABEL[status]} />;
  return (
    <span className="lb-status">
      <span className={`lb-dot ${status}`} />
      {STATUS_LABEL[status]}
    </span>
  );
}

/** What a dispatch does, in plain words from its key, with its plan task's title when it has one. */
function inlineWhat(d: ActiveSubagent, tasks: PlanTask[] | undefined): string {
  const what = d.key === undefined && d.taskId ? `task ${d.taskId.replace(/\+/g, ", ")}` : describeDispatch(d.key, d.role);
  return `${what}${d.taskId ? taskTitle(tasks, d.taskId) : ""}`;
}

/** How many of the costliest stages Cost by stage lists before folding the rest into one line. */
const TOP_STAGES = 10;

/** How many of the latest stage runs the collapsed Stage runs table shows. */
const LATEST_RUNS = 5;
/** An interactive session's Stage runs: the prototype's six latest rows. */
const INTERACTIVE_RUNS = 6;

/** The outcomes a dispatch ends on without finishing its work. */
const FAILED_OUTCOMES = new Set(["failed", "blocked", "aborted", "timed-out", "cancelled", "stopped"]);

type Step = ActiveSubagent & { phase: Phase; agentId?: string; endedAt?: string; outcome?: string; cause?: string; notes?: string };

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
      // The self-review's analysis pass records as a reviewer; "reviewer" names the code-review slots.
      role: d.role === "reviewer" && /-pass$/.test(d.key ?? "") ? "analyzer" : d.role,
      key: d.key || undefined,
      taskId: d.taskId || undefined,
      model: d.model,
      modelVersion: d.metrics?.modelVersion,
      effort: (d.effort && d.effort !== "default" ? d.effort : undefined) ?? d.metrics?.effort,
      startedAt: d.startedAt,
      endedAt: d.endedAt,
      outcome: d.outcome,
      cause: d.cause,
      notes: d.notes,
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
  // The main session ran an inline dispatch itself: its cost is in the "main session" line, never a bucket of its own.
  // An unpriced group says why: run inline, never run (failed before any agent id), or no agent id to match a bucket by.
  const unpriced = (ds: Step[]) =>
    ds.every((d) => d.agentId === "inline")
      ? "inline"
      : ds.every((d) => !d.agentId && d.outcome && FAILED_OUTCOMES.has(d.outcome))
        ? "never ran"
        : ds.every((d) => !d.agentId)
          ? "no agent id"
          : "—";
  const costLabel = (ds: Step[]) => ((c) => (c !== null ? formatUsd(c) : unpriced(ds)))(costOf(ds));
  return [...byRole].map(([role, ds]) => {
    const byModel = new Map<string, Step[]>();
    for (const d of ds) byModel.set(modelEffortLabel(d), [...(byModel.get(modelEffortLabel(d)) ?? []), d]);
    const models = [...byModel].sort((a, b) => b[1].length - a[1].length);
    return {
      role,
      count: ds.length,
      models: models.map(([m, xs]) => `${m} ×${xs.length} (${costLabel(xs)})`).join(", "),
      unpriced: unpriced(ds),
      effortUnknown: ds.some((d) => !knownEffort(d.effort)),
      cost: costOf(ds),
      failures: ds.filter((d) => d.endedAt && d.outcome && FAILED_OUTCOMES.has(d.outcome)),
      running: ds.some((d) => !d.endedAt),
    };
  })
    .sort((a, b) => (b.cost ?? -1) - (a.cost ?? -1));
}

/** A role's cost cell when none of its dispatches is priced, by why. */
const UNPRICED_COST: Record<string, string> = { inline: "in main session", "never ran": "never ran", "no agent id": "unattributed", "—": "—" };

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
 * one line; a line per role of its dispatches. A collapsible panel -- one
 * given onToggle -- shows its header alone while closed.
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
  note,
  open = true,
  onToggle,
  order,
}: {
  title: string;
  runs: StageRunDTO[];
  steps: Step[];
  now: number;
  extra?: ReactNode;
  nowText: string;
  progress: Progress;
  summary: string;
  /** A line under the Now line: Main's view of the self-review. */
  note?: ReactNode;
  open?: boolean;
  onToggle?: () => void;
  /** Its place in the single stack narrow widths fall back to. */
  order: number;
}) {
  const status = workStatus(runs);
  const active = runs.reduce((a, r) => a + runEnd(r, now) - Date.parse(r.startedAt), 0);
  const cost = measuredSum(runs.map((r) => readCostUsd(r.metrics)));
  // Costs are this panel's own runs' priced buckets, so its lines add up to its header.
  const costs = dispatchCosts(runs);
  const roles = roleSummary(steps, costs);
  const rest = residualCosts(runs, steps, costs);
  const slug = title.toLowerCase();
  return (
    <section className={`lb-panel lb-phase ${status}`} aria-label={title} style={{ order }}>
      <div className="lb-panel-h">
        <span className="lb-phase-name">
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
          {onToggle && (
            <button type="button" className="lb-fchip" aria-expanded={open} onClick={onToggle}>
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
                <Fragment key={r.role}>
                <li data-testid="phase-role">
                  <span className="lb-role">
                    {r.role} <span className="lb-muted">×{r.count}</span>
                  </span>
                  <span className="lb-role-models" title={r.effortUnknown ? `${r.models} -- effort unknown for some dispatches` : r.models}>
                    {r.models}
                  </span>
                  <span className="lb-num r">{r.cost !== null ? formatUsd(r.cost) : <span className="lb-muted">{UNPRICED_COST[r.unpriced]}</span>}</span>
                  <span className="lb-role-flags">
                    {r.failures.length > 0 && <span className="lb-err">{r.failures.length} failed</span>}
                    {r.running && (
                      <span className="lb-status">
                        <span className="lb-dot running" />
                        running
                      </span>
                    )}
                  </span>
                </li>
                {r.failures.map((d, i) => (
                  <li key={`${d.key ?? d.agentId}-${i}`} data-testid="phase-fail" className="lb-fail-why">
                    {d.key ?? d.role} <span className="lb-err">{d.outcome}</span>
                    {d.cause && ` (${d.cause})`}
                    {d.notes && `: ${d.notes}`}
                  </li>
                ))}
                </Fragment>
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
  );
}

/** A work panel's plan as its own card, closed to its header while its panel is. */
function TasksPanel({ title, tasks, current, open = true, order }: { title: string; tasks: PlanTask[]; current: number[]; open?: boolean; order: number }) {
  return (
    <section className={`lb-panel lb-tasks-panel${open ? "" : " collapsed"}`} aria-label={title} style={{ order }}>
      <div className="lb-panel-h">
        <span>{title}</span>
        <span className="lb-panel-note">{doneOf(tasks)}</span>
      </div>
      {open && <TaskList tasks={tasks} current={current} />}
    </section>
  );
}

/** A panel's status from its stage runs: idle with none, running while one is open, done after. */
function workStatus(runs: StageRunDTO[]): PhaseStatus {
  return runs.length === 0 ? "idle" : runs.some((r) => !r.endedAt) ? "running" : "done";
}

/** A plan's tasks, each marked done, in progress or pending. */
function TaskList({ tasks, current }: { tasks: PlanTask[]; current: number[] }) {
  return (
    <ul className="lb-tasks">
      {tasks.map((t) =>
        current.includes(t.n) ? (
          <li key={t.n} className="current" aria-current="step">
            <span className="lb-task-mark">●</span> {t.n}. {t.title}
          </li>
        ) : (
          <li key={t.n} className={t.state.replace(" ", "-")}>
            <span className="lb-task-mark">{TASK_MARK[t.state]}</span> {t.n}. {t.title}
          </li>
        ),
      )}
    </ul>
  );
}

/** A request's mark and the class its line takes, by its state and outcome. */
function requestLook(t: PlanTask): { cls: string; mark: string } {
  if (t.state === "in progress") return { cls: "current", mark: "●" };
  if (t.state === "pending") return { cls: "queued", mark: "○" };
  return t.outcome === "dropped" ? { cls: "dropped", mark: "–" } : { cls: t.outcome ?? "landed", mark: "✓" };
}

/** An interactive session's requests in the Tasks slot (KAN-1061): each with its mark, its short commit or outcome, and its brainstorm reason under it. */
function RequestsPanel({ tasks, order }: { tasks: PlanTask[]; order: number }) {
  return (
    <section className="lb-panel lb-tasks-panel" aria-label="Requests" style={{ order }}>
      <div className="lb-panel-h">
        <span>Requests</span>
        <span className="lb-panel-note">{doneOf(tasks)}</span>
      </div>
      <ul className="lb-reqs">
        {tasks.map((t) => {
          const { cls, mark } = requestLook(t);
          return (
            <li key={t.n} className={cls} aria-current={cls === "current" ? "step" : undefined}>
              <span className="m">{mark}</span>
              <span className="t">
                {t.n}. {t.title}
              </span>
              {cls === "landed" ? (
                <span className="lb-num">{t.commit}</span>
              ) : cls === "queued" ? (
                <span className="lb-muted">queued</span>
              ) : (
                <span>{cls === "current" ? "" : cls}</span>
              )}
              {t.reason && <span className="why">brainstormed — {t.reason}</span>}
            </li>
          );
        })}
      </ul>
    </section>
  );
}

/** An interactive session's Main panel: the request it follows, how long since that request's first stage began, and that request's steps. */
function RequestMainPanel({ runs, n, now, order }: { runs: StageRunDTO[]; n: number | undefined; now: number; order: number }) {
  const start = n === undefined ? undefined : requestStart(runs, n);
  return (
    <section className="lb-panel" aria-label="Main" style={{ order }}>
      <div className="lb-panel-h">
        <span>Main</span>
        {n !== undefined && <span className="lb-panel-note">{`Request ${n}${start === undefined ? "" : ` · ${formatSpan(now - start)}`}`}</span>}
      </div>
      {n !== undefined && (
        <div className="lb-panel-b">
          <ul className="lb-steps">
            {requestSteps(runs, n).map((s) => (
              <li key={s.stage} className={s.state === "done" ? "done" : s.state === "current" ? "cur" : undefined}>
                {stageLabel(s.stage)}
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
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
  analyze: "Analysis pass",
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
  // An interactive session's upcoming lanes are its followed request's stages still to run.
  const interactive = board !== undefined && isInteractive(board);
  const focus = interactive ? focusRequest(board, runs0) : undefined;
  const upcoming = interactive
    ? board.state === "FINISHED" || focus === undefined
      ? []
      : requestSteps(runs0, focus).filter((s) => s.state === "todo").map((s) => s.stage)
    : command
      ? order.slice(furthest + 1).filter((s) => !ran.has(s))
      : [];

  // The phase panels' dispatches and the self-review findings, refetched on the board's cadence.
  const [dispatches, setDispatches] = useState<RecordDispatch[]>([]);
  const [findings, setFindings] = useState<SelfReviewFinding[]>([]);
  const [decision, setDecision] = useState<Decision | null>(null);
  useEffect(() => {
    let live = true;
    const load = () => {
      getDecisions(project, change)
        .then((r) => live && setDecision(r?.[0]?.decision ?? null))
        .catch(() => live && setDecision(null));
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
  // The tasks Main works on: every running dispatch's, a task group's each, else the plan's in-progress one.
  const running = mainSteps.flatMap((d) => (d.endedAt ? [] : dispatchTasks(d)));
  const mainCurrentTasks = running.length ? running : (board?.tasks ?? []).filter((t) => t.state === "in progress").map((t) => t.n);
  const srOpen = board?.currentStage === "flow.self-review" || srRuns.some((r) => !r.endedAt);
  const srView = isSub
    ? selfReviewView(steps, board?.tasks, findings.length + parentFindings.length, srOpen, runs0.length > 0, now)
    : selfReviewView(srSteps, board?.selfReviewTasks, allFindings.length, srOpen, srStarted, now);
  const srCost = measuredSum(srRuns.map((r) => readCostUsd(r.metrics)));
  // The self-review panel and its tasks open while it runs, until the operator toggles them.
  const [srToggled, setSrToggled] = useState<boolean | null>(null);
  const srPanelOpen = srToggled ?? srOpen;

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
          const latestRuns = interactive ? INTERACTIVE_RUNS : LATEST_RUNS;
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
          // An interactive session with no request in progress and no stage open waits for the operator's next one.
          const waitingNext = interactive && interactiveWaiting(board);
          // A stage the session just opened carries no harvested metrics yet: main keeps the latest measured run's model and effort.
          const mainMetrics = [mainRun, ...[...runs].sort((a, b) => Date.parse(b.startedAt) - Date.parse(a.startedAt))].find((r) => r && readMainModels(r.metrics))?.metrics;
          const mainAgent = mainRun && {
            what: waitingNext ? "waiting for the next request" : inline ? inlineWhat(inline, board?.tasks) : openRun ? `${stageLabel(openRun.stage)}${board && mainAction(board) ? ` · ${mainAction(board)}` : board && agentsDone(board) ? ", wrapping up after its agents" : ""}` : (board?.activity === "working" && board.activityDetail ? board.activityDetail : `between stages, after ${stageLabel(mainRun.stage)}`),
            model: inline
              ? formatModel(inline.modelVersion ?? inline.model)
              : (readMainModels(mainMetrics)?.map(formatModel).join(", ") ?? "—"),
            effort: readEffort(mainMetrics) ?? undefined,
            since: inline?.startedAt ?? (openRun ? openRun.startedAt : (mainRun.endedAt ?? mainRun.startedAt)),
          };
          const request = interactive ? currentRequest(board) : undefined;
          const nowTask = request ? `Request ${request.n}: ${request.title}` : board ? runningTask(board) : null;
          const count = board && taskCount(board);
          const agentCount = subagents.length + (mainAgent ? 1 : 0);
          const mainStatus: AgentStatus = waitingNext
            ? "idle"
            : board?.activity === "question" || board?.activity === "permission"
              ? "needs-you"
              : board !== undefined && effectiveActivity(board) === "idle"
                ? "idle"
                : !inline && subagents.length > 0
                  ? "waiting"
                  : "running";
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
          const timelinePanel = (
          <section className="lb-panel" aria-label="Timeline" style={interactive ? { order: 5 } : undefined}>
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
          );
          const stageRunsPanel = (
          <section className="lb-panel" aria-label="Stage runs" style={interactive ? { order: 6 } : undefined}>
            <div className="lb-panel-h">
              <span>Stage runs</span>
              <span className="lb-panel-note">
                click a row for its dispatches
                {runs.length > latestRuns && (
                  <>
                    {" · "}
                    <button type="button" className="lb-fchip" onClick={() => setAllRuns(!allRuns)}>
                      {allRuns ? `Show latest ${latestRuns}` : `Show all ${runs.length}`}
                    </button>
                  </>
                )}
              </span>
            </div>
            <StageRunTable
              stageRuns={allRuns ? runs : [...runs].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt)).slice(-latestRuns)}
              truncated={state.truncated}
              totalStageRuns={state.totalStageRuns}
              records={dispatches}
              requests={interactive}
            />
          </section>
          );
          // An interactive session's every request by its stage runs' cost, costliest first -- never folded.
          const byRequest = new Map<number, number>();
          for (const r of runs) if (r.request !== undefined) byRequest.set(r.request, (byRequest.get(r.request) ?? 0) + (readCostUsd(r.metrics) ?? 0));
          const requestCost = [...byRequest].filter(([, v]) => v > 0).sort((a, b) => b[1] - a[1]);
          const maxRequest = Math.max(0, ...requestCost.map(([, v]) => v));
          const requestCostPanel = (
          <section className="lb-panel" aria-label="Cost by request" style={{ order: 7 }}>
            <div className="lb-panel-h">
              <span>Cost by request</span>
            </div>
            <div className="lb-panel-b">
              {requestCost.map(([n, cost]) => (
                <div key={n} className="lb-hbar">
                  <span className="lb-n" title={board?.tasks?.find((t) => t.n === n)?.title}>
                    Request {n}
                  </span>
                  <span className="lb-t">
                    <span style={{ width: `${maxRequest ? (cost / maxRequest) * 100 : 0}%` }} />
                  </span>
                  <span className="r lb-num">{formatUsd(cost)}</span>
                </div>
              ))}
            </div>
          </section>
          );
          const costPanel = (
          <section className="lb-panel" aria-label="Cost by stage" style={interactive ? { order: 7 } : undefined}>
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
          );

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
                    ) : waitingNext ? (
                      <>
                        <span className="lb-ring" /> Waiting for the next request
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
                        !interactive &&
                        `for ${formatSpan(now - Date.parse(openRun.startedAt))}${
                          count?.total !== undefined ? ` · tasks ${count.done}/${count.total}` : ""
                        }`}
                      {nowTask && <div className="lb-task-now">{nowTask}</div>}
                      {board && !waitingNext && (mainAction(board) ? <div className="lb-task-now">{mainAction(board)}</div> : agentsDone(board) && <div className="lb-task-now">Agents done, main is wrapping up the stage</div>)}
                    </>
                  }
                >
                  {(openRun || betweenStages) && !waitingNext && board && <ActivityBadge row={board} now={now} />}
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
              </section>
              {board && (
                <div className="lb-kpi-groups">
                  <section className="lb-panel" aria-label="Main totals">
                    <div className="lb-panel-h">
                      <span>Main</span>
                    </div>
                    <div className="lb-kpis">
                      <Kpi label="Fix rounds" value={<span className="lb-num">{board.fixRounds}</span>} />
                      <Kpi label="Full panel runs" value={<span className="lb-num">{board.fullPanelRuns}</span>} hint={`of ${board.panelRuns} panel rounds`} />
                      <Kpi label="Reviewer runs" value={<span className="lb-num">{board.reviewerRuns}</span>} />
                      <Kpi label="Findings" value={<Cim c={board.critical} i={board.important} m={board.minor} />} />
                    </div>
                  </section>
                  <section className="lb-panel" aria-label="Self-review totals">
                    <div className="lb-panel-h">
                      <span>Self-review</span>
                    </div>
                    <div className="lb-kpis">
                      <Kpi label="Fix rounds" value={<span className="lb-num">{board.selfReviewFixRounds}</span>} />
                      <Kpi label="Reviewer runs" value={<span className="lb-num">{board.selfReviewReviewerRuns}</span>} />
                      <Kpi label="Findings" value={<Cim c={board.selfReviewCritical} i={board.selfReviewImportant} m={board.selfReviewMinor} />} />
                    </div>
                  </section>
                </div>
              )}

              {/* Two independent stacks right under the KPI strip -- agents and work panels left, their plans right -- each card its own height; an interactive session's stage runs and cost rail join them. */}
              <div className={`lb-work${interactive ? " lb-work-interactive" : ""}`}>
                <div className="lb-work-col">
              {agentCount > 0 && (
                  <section className="lb-panel" aria-label="Agents" style={{ order: 0 }}>
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
                              <StatusDot status={mainStatus} bare={interactive} />
                            </td>
                            <td className="lb-clip" title={mainAgent.what}>{mainAgent.what}</td>
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
                              <StatusDot status="running" bare={interactive} />
                            </td>
                            <td className="lb-clip" title={inlineWhat(d, board?.tasks)}>
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
                    <WorkPanel title="Main" runs={runs0} steps={steps} now={now} nowText={srView.nowText} progress={srView.progress} summary={srView.summary} order={1} />
                  ) : interactive ? (
                    <>
                      <RequestMainPanel runs={runs0} n={focus} now={now} order={1} />
                      {srStarted && (
                        <WorkPanel
                          title="Self-review"
                          runs={srRuns}
                          steps={srSteps}
                          now={now}
                          order={3}
                          open={srPanelOpen}
                          onToggle={() => setSrToggled(!srPanelOpen)}
                          nowText={srView.nowText}
                          progress={srView.progress}
                          summary={srView.summary}
                        />
                      )}
                    </>
                  ) : (
                    <>
                      <WorkPanel
                        title="Main"
                        runs={mainRuns}
                        steps={mainSteps}
                        now={now}
                        order={1}
                        nowText={srOpen ? "Running the self-review" : mainNow(mainRuns, mainSteps, board?.tasks, board?.state === "FINISHED", now)}
                        progress={mainProgress(mainRuns).map((x) => ({
                          label: x.phase.toLowerCase(),
                          // The self-review runs in its own sub-run, so Main's own stages never carry it.
                          state: x.phase !== "Self-review" ? x.state : srOpen ? "current" : srStatus === "done" ? "done" : x.state,
                        }))}
                        summary={mainSummary(board?.tasks, mainSteps)}
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
                        order={3}
                        open={srPanelOpen}
                        onToggle={() => setSrToggled(!srPanelOpen)}
                        extra={sub && <a href={`#/run/${encodeURIComponent(sub.project)}/${encodeURIComponent(sub.name)}`}>sub-run</a>}
                        nowText={srView.nowText}
                        progress={srView.progress}
                        summary={srView.summary}
                      />
                    </>
                  )}
                  {interactive && timelinePanel}
                  {interactive && stageRunsPanel}
                </div>
                <div className="lb-work-col">
                  {board?.tasks &&
                    (interactive ? (
                      <RequestsPanel tasks={board.tasks} order={2} />
                    ) : (
                      <TasksPanel title="Tasks" tasks={board.tasks} current={isSub ? [srView.currentTask ?? 0] : mainCurrentTasks} order={2} />
                    ))}
                  {!isSub && board?.selfReviewTasks && (
                    <TasksPanel title="Self-review tasks" tasks={board.selfReviewTasks} current={[srView.currentTask ?? 0]} open={srPanelOpen} order={4} />
                  )}
                  {interactive && requestCostPanel}
                  {interactive && costPanel}
                </div>
              </div>

              {!interactive && (
                <>
                  {timelinePanel}
                  <div className="lb-split">
                    {stageRunsPanel}
                    <div className="lb-work-col">
                      {costPanel}
                      {decision && <DecisionsPanel decision={decision} />}
                    </div>
                  </div>
                </>
              )}
            </>
          );
        })()}
    </section>
  );
}
