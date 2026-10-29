import type { StateBoardRow } from "../api";

/** A change's done and total task counts: from its plan when the board read one, else the stored implementer-commit count over the latest task-count observation. */
export function taskCount(row: Pick<StateBoardRow, "tasks" | "doneTasks" | "currentTasks">): { done: number; total?: number } {
  if (row.tasks) return { done: row.tasks.filter((t) => t.state === "done").length, total: row.tasks.length };
  return { done: row.doneTasks ?? 0, total: row.currentTasks };
}
