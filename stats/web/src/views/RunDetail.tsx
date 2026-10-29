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
import { useMemo, type ReactNode } from "react";
import { StageRunTable, groupSessions, runEnd } from "../components/StageRunTable";
import { StageTimeline } from "../components/StageTimeline";
import { useRunDetail } from "../hooks/useRunDetail";
import { formatModel, formatRelative, formatSpan, formatUsd } from "../format";
import { readCacheReadTokens, readCostUsd, readEffort, readModels, readTotalInputTokens } from "../metrics";
import { projectLabel } from "../lib/projectLabel";
import { runningTask } from "../lib/runningTask";
import { stageLabel } from "../lib/stageLabel";
import { ActivityBadge, StatePill } from "./StateBoard";
import type { ActiveSubagent, PlanTask, StageRunDTO } from "../api";

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

/** What the main session's inline dispatch does: its task with title, else its key, else its role. */
function inlineWhat(d: ActiveSubagent, tasks: PlanTask[] | undefined): string {
  return d.taskId ? `task ${d.taskId}${taskTitle(tasks, d.taskId)}` : (d.key ?? d.role);
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
  const board = state.status === "ready" ? state.board : undefined;

  return (
    <section aria-label={`Run detail: ${project}/${change}`} className="dashboard lb">
      <div className="lb-crumb">
        <a href="#/state-board">← Live state board</a>
      </div>
      <div className="lb-runhead">
        <div className="lb-title">
          <h2>{change}</h2>
          {board && <StatePill state={board.state} />}
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
          const betweenStages = !openRun && lastRun && (inline !== undefined || (board?.activity !== undefined && board.activity !== "idle"));
          const mainRun = openRun ?? (betweenStages ? lastRun : undefined);
          const mainAgent = mainRun && {
            what: inline ? inlineWhat(inline, board?.tasks) : openRun ? stageLabel(openRun.stage) : `after ${stageLabel(mainRun.stage)}`,
            model: inline
              ? formatModel(inline.modelVersion ?? inline.model)
              : (readModels(mainRun.metrics)?.map(formatModel).join(", ") ?? "—"),
            effort: readEffort(mainRun.metrics) ?? undefined,
            since: inline?.startedAt ?? (openRun ? openRun.startedAt : (board?.activitySince ?? mainRun.endedAt ?? mainRun.startedAt)),
          };
          const nowTask = board ? runningTask(board) : null;
          const agentCount = subagents.length + (mainAgent ? 1 : 0);
          const mainStatus: AgentStatus =
            board?.activity === "question" || board?.activity === "permission"
              ? "needs-you"
              : board?.activity === "idle"
                ? "idle"
                : !inline && subagents.length > 0
                  ? "waiting"
                  : "running";
          const failed = runs.filter((r) => r.outcome === "failed");
          const byStage = new Map<string, number>();
          for (const r of runs) byStage.set(r.stage, (byStage.get(r.stage) ?? 0) + (readCostUsd(r.metrics) ?? 0));
          const stageCost = [...byStage].sort((a, b) => b[1] - a[1]);
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
                    ) : nowTask ? (
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
                          board?.currentTasks !== undefined ? ` · tasks ${board.doneTasks ?? 0}/${board.currentTasks}` : ""
                        }`}
                      {nowTask && <div className="lb-task-now">{nowTask}</div>}
                    </>
                  }
                >
                  {(openRun || nowTask) && board && <ActivityBadge row={board} now={now} />}
                </Kpi>
                <Kpi
                  label="Total cost"
                  value={<span className="lb-num">{summary.totalCostUsd === null ? "—" : formatUsd(summary.totalCostUsd)}</span>}
                  hint={`${summary.runCount} stage runs, ${summary.measuredRuns} measured`}
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

              {(board?.tasks || agentCount > 0) && (
                <div className="lb-split">
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
                              {d.taskId ? `task ${d.taskId}${taskTitle(board?.tasks, d.taskId)}` : (d.key ?? "—")}
                            </td>
                            <td>
                              {formatModel(d.modelVersion ?? d.model)}
                              {d.effort && ` · ${d.effort}`}
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
                  <section className="lb-panel" aria-label="Tasks">
                    <div className="lb-panel-h">
                      <span>Tasks</span>
                      <span className="lb-panel-note">
                        {board?.tasks ? `${board.tasks.filter((t) => t.state === "done").length}/${board.tasks.length} done` : "no plan yet"}
                      </span>
                    </div>
                    <ul className="lb-tasks">
                      {(board?.tasks ?? []).map((t) => (
                        <li key={t.n} className={t.state.replace(" ", "-")}>
                          <span className="lb-task-mark">{TASK_MARK[t.state]}</span> {t.n}. {t.title}
                        </li>
                      ))}
                    </ul>
                  </section>
                </div>
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
                  </span>
                </div>
                <div className="lb-panel-b lb-scroll">
                  <StageTimeline stageRuns={runs} sessionOf={sessionOf} />
                </div>
              </section>

              <div className="lb-split">
                <section className="lb-panel" aria-label="Stage runs">
                  <div className="lb-panel-h">
                    <span>Stage runs</span>
                    <span className="lb-panel-note">click a row for its dispatches</span>
                  </div>
                  <StageRunTable stageRuns={runs} truncated={state.truncated} totalStageRuns={state.totalStageRuns} />
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
                  </div>
                </section>
              </div>
            </>
          );
        })()}
    </section>
  );
}
