import type { StateBoardRow } from "../api";

/** The plan task a running dispatch -- main's inline one or a subagent's -- works on, as "Task 6: <title>", or null when none does. */
export function runningTask(row: StateBoardRow): string | null {
  for (const d of row.subagents ?? []) {
    const t = d.taskId === undefined ? undefined : row.tasks?.find((t) => String(t.n) === d.taskId);
    if (t) return `Task ${t.n}: ${t.title}`;
  }
  return null;
}
