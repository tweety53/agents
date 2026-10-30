import type { PlanTask, StageRunDTO, StateBoardRow } from "../api";
import type { ProgressState } from "./mainNow";

type Row = Pick<StateBoardRow, "command" | "tasks" | "currentStage">;

/** Whether the change is a /flow-interactive session, whose tasks are its requests (KAN-1061). */
export function isInteractive(row: Pick<StateBoardRow, "command">): boolean {
  return row.command === "/flow-interactive";
}

/** The request in progress, if any. */
export function currentRequest(row: Row): PlanTask | undefined {
  return row.tasks?.find((t) => t.state === "in progress");
}

/** The highest-numbered ended request, if any. */
export function lastEndedRequest(row: Row): PlanTask | undefined {
  return (row.tasks ?? []).filter((t) => t.state === "done").reduce<PlanTask | undefined>((a, t) => (a && a.n > t.n ? a : t), undefined);
}

/** Between requests: none in progress and no stage open. */
export function interactiveWaiting(row: Row): boolean {
  return currentRequest(row) === undefined && row.currentStage === undefined;
}

/** A request's own stages, in the order a request runs them (design.md's One request); flow.brainstorm joins after the first only when the request ran one. */
export const REQUEST_STAGES = ["flow.load-context", "flow.sdd-tdd", "flow.verify", "flow.landing-routes", "flow.refresh-main-checkout", "flow.visual-verify"];

/** The request the run view follows: the one in progress, else the highest any stage run carries. */
export function focusRequest(row: Row, runs: StageRunDTO[]): number | undefined {
  return currentRequest(row)?.n ?? runs.reduce<number | undefined>((a, r) => (r.request !== undefined && (a === undefined || r.request > a) ? r.request : a), undefined);
}

/** Request n's steps: each of its stages done once a run of it ended, current while one is open, else todo. */
export function requestSteps(runs: StageRunDTO[], n: number): { stage: string; state: ProgressState }[] {
  const own = runs.filter((r) => r.request === n);
  const stages = own.some((r) => r.stage === "flow.brainstorm") ? [REQUEST_STAGES[0], "flow.brainstorm", ...REQUEST_STAGES.slice(1)] : REQUEST_STAGES;
  return stages.map((stage) => {
    const of = own.filter((r) => r.stage === stage);
    return { stage, state: of.some((r) => !r.endedAt) ? "current" : of.length ? "done" : "todo" };
  });
}

/** When request n's first stage run began, if any. */
export function requestStart(runs: StageRunDTO[], n: number): number | undefined {
  const starts = runs.filter((r) => r.request === n).map((r) => Date.parse(r.startedAt));
  return starts.length ? Math.min(...starts) : undefined;
}
