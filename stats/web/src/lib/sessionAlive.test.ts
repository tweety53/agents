// sessionAlive reads a session as working on an open stage, a running
// dispatch, or a transcript written within ALIVE_MS -- and as paused otherwise.
import { describe, expect, it } from "vitest";
import type { StateBoardRow } from "../api";
import { ALIVE_MS, sessionAlive } from "./sessionAlive";

const NOW = Date.parse("2026-01-01T12:00:00Z");
const base = { projectKey: "p", name: "c", state: "STARTED", updatedAt: "", updatedBy: "", nextCommand: "", runCount: 0, costUsd: 0, cost24h: 0, cost7d: 0, cost30d: 0, wallClockMs: 0, errors: 0 } as StateBoardRow;

describe("sessionAlive", () => {
  it("is alive on an open stage", () => expect(sessionAlive({ ...base, currentStage: "flow.verify" }, NOW)).toBe(true));
  it("is alive on a running dispatch", () =>
    expect(sessionAlive({ ...base, subagents: [{ role: "implementer", model: "opus", startedAt: "" }] }, NOW)).toBe(true));
  it("is alive on a fresh transcript write", () =>
    expect(sessionAlive({ ...base, sessionSeenAt: new Date(NOW - ALIVE_MS + 1000).toISOString() }, NOW)).toBe(true));
  it("is paused on a stale transcript write", () =>
    expect(sessionAlive({ ...base, sessionSeenAt: new Date(NOW - ALIVE_MS - 1000).toISOString() }, NOW)).toBe(false));
  it("is paused with nothing at all", () => expect(sessionAlive(base, NOW)).toBe(false));
});
