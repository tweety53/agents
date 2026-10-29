// runningTask names a running plan task always, and a non-task dispatch only
// between stages, where no stage name already says what is happening.
import { describe, expect, it } from "vitest";
import type { StateBoardRow } from "../api";
import { agentsDone, lastAction, runningTask } from "./runningTask";

const base = { projectKey: "p", name: "c", state: "IN_PROGRESS", updatedAt: "", updatedBy: "", nextCommand: "", runCount: 0, costUsd: 0, cost24h: 0, cost7d: 0, cost30d: 0, wallClockMs: 0, errors: 0 } as StateBoardRow;
const reviewer = { role: "reviewer", key: "panel-2-primary", model: "opus", startedAt: "" };

describe("runningTask", () => {
  it("names a running plan task", () =>
    expect(
      runningTask({ ...base, currentStage: "flow.sdd-tdd", tasks: [{ n: 6, title: "Fix", state: "in progress" }], subagents: [{ ...reviewer, taskId: "6" }] }),
    ).toBe("Task 6: Fix"));
  it("leaves a non-task dispatch to the open stage's name", () =>
    expect(runningTask({ ...base, currentStage: "flow.review-panel", subagents: [reviewer] })).toBeNull());
  it("describes a non-task dispatch between stages", () =>
    expect(runningTask({ ...base, subagents: [reviewer] })).toBe("Code review round 2: primary"));
});

describe("agentsDone", () => {
  it("is true once every dispatch the open stage began has ended", () =>
    expect(agentsDone({ ...base, currentStage: "flow.review-panel", openStageDispatches: 3 })).toBe(true));
  it("is false while one still runs", () =>
    expect(agentsDone({ ...base, currentStage: "flow.review-panel", openStageDispatches: 3, subagents: [reviewer] })).toBe(false));
  it("is false for a stage that dispatched nothing", () => expect(agentsDone({ ...base, currentStage: "flow.verify" })).toBe(false));
});

describe("lastAction", () => {
  it("names main's last action between stages", () =>
    expect(lastAction({ ...base, activity: "working", activityDetail: "Bash: Run the linters" })).toBe("Last: Bash: Run the linters"));
  it("defers to an open stage", () =>
    expect(lastAction({ ...base, currentStage: "flow.verify", activity: "working", activityDetail: "Edit a.ts" })).toBeNull());
  it("defers to a running dispatch", () =>
    expect(lastAction({ ...base, subagents: [reviewer], activity: "working", activityDetail: "Edit a.ts" })).toBeNull());
  it("is null with no detail", () => expect(lastAction({ ...base, activity: "working" })).toBeNull());
});
