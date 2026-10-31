import type { StateBoardRow } from "../api";

/** How recently a session must have written to its transcript to read as working. */
export const ALIVE_MS = 90_000;

/** Whether the change's session is working right now -- never once it ended: an open stage, a running dispatch, or a transcript written within ALIVE_MS. */
export function sessionAlive(row: StateBoardRow, now: number): boolean {
  // A closed session is not working, whatever stage or write it left behind (SessionEnd).
  if (row.activity === "ended") return false;
  if (row.currentStage !== undefined || (row.subagents?.length ?? 0) > 0) return true;
  return row.sessionSeenAt !== undefined && now - Date.parse(row.sessionSeenAt) < ALIVE_MS;
}

/** The session's activity as the board should read it: a reported "idle" stays idle only while
 * nothing runs on -- a turn that ended while its subagents run is "waiting", and one a stop hook
 * refused, which keeps writing its transcript after the report, is still "working". */
export function effectiveActivity(row: StateBoardRow): StateBoardRow["activity"] | "waiting" {
  if (row.activity !== "idle") return row.activity;
  if ((row.subagents ?? []).some((d) => !d.inline)) return "waiting";
  const wroteSince = row.sessionSeenAt !== undefined && row.activitySince !== undefined && Date.parse(row.sessionSeenAt) > Date.parse(row.activitySince);
  return wroteSince ? "working" : "idle";
}
