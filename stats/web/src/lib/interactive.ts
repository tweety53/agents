import type { PlanTask, StateBoardRow } from "../api";

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
