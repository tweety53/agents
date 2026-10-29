// What a change's main pipeline -- every phase but Self-review -- is doing
// right now, and where it stands, derived from its stage runs, its dispatches
// and its plan.
import type { PlanTask, StageRunDTO } from "../api";
import { formatModel, formatSpan } from "../format";
import { describeDispatch } from "./describeDispatch";
import { PHASES, STAGE_PHASE, type Phase } from "./phase";
import { stageLabel } from "./stageLabel";
import type { SrDispatch } from "./selfReviewNow";

export type ProgressState = "done" | "current" | "todo";

/** A dispatch in plain words, an implementer's task named with its plan title. */
function dispatchLabel(d: SrDispatch, tasks: PlanTask[] | undefined): string {
  const what = describeDispatch(d.key, d.role);
  const task = d.taskId ? tasks?.find((t) => String(t.n) === d.taskId) : undefined;
  return task ? `${what} "${task.title}"` : what;
}

/**
 * The main pipeline's "Now:" line: its running dispatch, with its time and
 * model; else its open stage, with its time; "Done" once the change finished;
 * else the main session between steps, naming the last stage or dispatch to
 * end and how long ago.
 */
export function mainNow(runs: StageRunDTO[], dispatches: SrDispatch[], tasks: PlanTask[] | undefined, finished: boolean, now: number): string {
  const byStart = <T extends { startedAt: string }>(xs: T[]) => [...xs].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt));
  const running = byStart(dispatches.filter((d) => !d.endedAt)).pop();
  if (running) return `${dispatchLabel(running, tasks)} · ${formatSpan(now - Date.parse(running.startedAt))} · ${formatModel(running.modelVersion ?? running.model)}`;
  const open = byStart(runs.filter((r) => !r.endedAt)).pop();
  if (open) return `${stageLabel(open.stage)} · ${formatSpan(now - Date.parse(open.startedAt))}`;
  if (finished) return "Done";
  const ends = [
    ...runs.filter((r) => r.endedAt).map((r) => ({ at: Date.parse(r.endedAt!), what: stageLabel(r.stage) })),
    ...dispatches.filter((d) => d.endedAt).map((d) => ({ at: Date.parse(d.endedAt!), what: dispatchLabel(d, tasks) })),
  ].sort((a, b) => a.at - b.at);
  const last = ends.pop();
  return last ? `Main session is between steps (last: ${last.what}, ${formatSpan(now - last.at)} ago)` : "Main session is between steps (nothing run yet)";
}

/** The main pipeline's phases in order -- Fixes only when one ran -- each done once it ran and closed, current while it holds the open stage. */
export function mainProgress(runs: StageRunDTO[]): { phase: Phase; state: ProgressState }[] {
  const ran = new Set(runs.map((r) => STAGE_PHASE[r.stage]));
  const open = new Set(runs.filter((r) => !r.endedAt).map((r) => STAGE_PHASE[r.stage]));
  return PHASES.filter((p) => p !== "Self-review" && (p !== "Fixes" || ran.has(p))).map((phase) => ({
    phase,
    state: open.has(phase) ? "current" : ran.has(phase) ? "done" : "todo",
  }));
}

/** Main's summary line: plan done/total, code-review rounds and their fix rounds, each left out when there are none. */
export function mainSummary(tasks: PlanTask[] | undefined, dispatches: SrDispatch[]): string {
  const rounds = (re: RegExp) => new Set(dispatches.map((d) => re.exec(d.key ?? "")?.[1]).filter((r) => r !== undefined)).size;
  const review = rounds(/^panel-(\d+)-/);
  const fix = rounds(/^panel-fix-(\d+)$/);
  return [
    tasks?.length && `tasks ${tasks.filter((t) => t.state === "done").length}/${tasks.length}`,
    review > 0 && `review rounds ${review}`,
    fix > 0 && `fix rounds ${fix}`,
  ]
    .filter(Boolean)
    .join(" · ");
}
