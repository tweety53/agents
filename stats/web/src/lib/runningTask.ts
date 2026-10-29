import type { StateBoardRow } from "../api";
import { describeDispatch } from "./describeDispatch";

/** What a running dispatch -- main's inline one first, then a subagent's -- works on: its plan task as "Task 6: <title>", else its key in plain words; null when none runs. */
export function runningTask(row: StateBoardRow): string | null {
  const ds = [...(row.subagents ?? [])].sort((a, b) => Number(!a.inline) - Number(!b.inline));
  for (const d of ds) {
    const t = d.taskId === undefined ? undefined : row.tasks?.find((t) => String(t.n) === d.taskId);
    if (t) return `Task ${t.n}: ${t.title}`;
  }
  return ds[0] ? describeDispatch(ds[0].key, ds[0].role) : null;
}
