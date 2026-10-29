// runningTask names a running plan task always, and a non-task dispatch only
// between stages, where no stage name already says what is happening.
import { describe, expect, it } from "vitest";
import type { StateBoardRow } from "../api";
import { runningTask } from "./runningTask";

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
