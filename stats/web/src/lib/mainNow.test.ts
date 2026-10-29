import { describe, expect, it } from "vitest";
import type { PlanTask, StageRunDTO } from "../api";
import { mainNow, mainProgress, mainSummary } from "./mainNow";

const T0 = Date.parse("2026-01-01T00:00:00Z");
const at = (min: number) => new Date(T0 + min * 60000).toISOString();
const run = (stage: string, start: number, end?: number): StageRunDTO => ({
  stageRunId: start,
  harness: "claude-code",
  command: "/flow",
  stage,
  attempt: 1,
  startedAt: at(start),
  endedAt: end === undefined ? undefined : at(end),
});
const d = (key: string, start: number, end?: number, extra = {}) => ({ key, role: "implementer", model: "opus", startedAt: at(start), endedAt: end === undefined ? undefined : at(end), ...extra });
const tasks: PlanTask[] = [
  { n: 1, title: "Schema", state: "done" },
  { n: 3, title: "Run view", state: "in progress" },
];

describe("mainNow", () => {
  it("names a running implementer by its task title, time and model", () =>
    expect(mainNow([run("flow.sdd-tdd", 0)], [d("task-3-implementer", 5, undefined, { taskId: "3", modelVersion: "claude-opus-5-5" })], tasks, false, T0 + 17 * 60000)).toBe(
      'Implementing task 3 "Run view" · 12m · opus 5.5',
    ));
  it("names a running review round", () =>
    expect(mainNow([run("flow.review-panel", 0)], [d("panel-2-primary", 1, undefined, { role: "reviewer" })], tasks, false, T0 + 3 * 60000)).toBe(
      "Code review round 2: primary · 2m · opus",
    ));
  it("names the open stage when no dispatch runs", () => expect(mainNow([run("flow.verify", 0)], [], tasks, false, T0 + 4 * 60000)).toBe("Running lint and tests · 4m"));
  it("reads Done once the change finished", () => expect(mainNow([run("flow.cleanup", 0, 1)], [], tasks, true, T0 + 60000)).toBe("Done"));
  it("reads main between steps after the last thing to end", () =>
    expect(mainNow([run("flow.commit-archive", 0, 2)], [d("task-1-implementer", 0, 1, { taskId: "1" })], tasks, false, T0 + 6 * 60000)).toBe(
      "Main session is between steps (last: Committing the archive, 4m ago)",
    ));
});

describe("mainProgress", () => {
  it("marks ran phases done, the open one current, the rest pending; Fixes only when one ran", () => {
    const p = mainProgress([run("flow.kickoff", 0, 1), run("flow.sdd-tdd", 2, 3), run("flow.review-panel", 4)]);
    expect(p.map((x) => `${x.phase}:${x.state}`)).toEqual([
      "Planning:done",
      "Decisions:todo",
      "Implementation:done",
      "Review:current",
      "Verification:todo",
      "Handoff:todo",
      "Integrate:todo",
      "Landing:todo",
      "Cleanup:todo",
    ]);
    expect(mainProgress([run("flow.document-fix", 0, 1)]).map((x) => x.phase)).toContain("Fixes");
  });
});

describe("mainSummary", () => {
  it("counts plan ticks, review rounds and fix rounds, leaving out what is zero", () => {
    expect(mainSummary(tasks, [d("panel-0-primary", 0), d("panel-0-mutation", 0), d("panel-1-primary", 1), d("panel-fix-1", 1)])).toBe("tasks 1/2 · review rounds 2 · fix rounds 1");
    expect(mainSummary(undefined, [])).toBe("");
  });
});
