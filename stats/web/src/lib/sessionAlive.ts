import type { StateBoardRow } from "../api";

/** How recently a session must have written to its transcript to read as working. */
export const ALIVE_MS = 90_000;

/** Whether the change's session is working right now: an open stage, a running dispatch, or a transcript written within ALIVE_MS. */
export function sessionAlive(row: StateBoardRow, now: number): boolean {
  if (row.currentStage !== undefined || (row.subagents?.length ?? 0) > 0) return true;
  return row.sessionSeenAt !== undefined && now - Date.parse(row.sessionSeenAt) < ALIVE_MS;
}
