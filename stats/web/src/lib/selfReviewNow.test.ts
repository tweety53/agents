import { describe, expect, it } from "vitest";
import type { PlanTask } from "../api";
import { selfReviewNow, srStep } from "./selfReviewNow";

const N = "self-review-kan-9-x";
const T0 = Date.parse("2026-01-01T00:00:00Z");
const at = (min: number) => new Date(T0 + min * 60000).toISOString();
const tasks: PlanTask[] = [
  { n: 1, title: "Fix A", state: "done" },
  { n: 2, title: "Fix B", state: "pending" },
];

describe("srStep", () => {
  it.each([
    [`${N}-pass`, { step: "review" }],
    [`${N}-plan`, { step: "plan" }],
    [`${N}-plan-2`, { step: "plan" }],
    [`${N}-fix`, { step: "fix" }],
    [`${N}-fix-g3`, { step: "fix" }],
    [`${N}-fix-2`, { step: "fix", round: 2 }],
    [`${N}-review-1`, { step: "re-review", round: 1 }],
    [`${N}-fix-v4`, { step: "verify", round: 4 }],
  ])("%s", (key, want) => expect(srStep(key)).toEqual(want));
  it("ignores every other key", () => {
    expect(srStep("task-1-implementer")).toBeNull();
    expect(srStep(undefined)).toBeNull();
  });
});

describe("selfReviewNow", () => {
  const d = (key: string, start: number, end?: number, extra = {}) => ({ key, role: "implementer", model: "opus", startedAt: at(start), endedAt: end === undefined ? undefined : at(end), ...extra });
  it("names a running group fixer by the first unticked task, with its time and model", () =>
    expect(selfReviewNow([d(`${N}-plan`, 0, 10), d(`${N}-fix-g1`, 11, undefined, { modelVersion: "claude-opus-5-5" })], tasks, true, T0 + 20 * 60000)).toEqual({
      text: 'Fixing task 2 "Fix B" · for 9m · opus 5.5',
      current: "fix",
      done: false,
      taskN: 2,
    }));
  it("names the task a fixer carries by id", () => expect(selfReviewNow([d(`${N}-fix`, 0, undefined, { taskId: "1" })], tasks, true, T0).taskN).toBe(1));
  it("reads a red verify fix with its round", () => expect(selfReviewNow([d(`${N}-fix-v2`, 0)], tasks, true, T0).text).toMatch(/^Verifying — red verify fix round 2 · for/));
  it("reads main between steps after the latest dispatch ended", () =>
    expect(selfReviewNow([d(`${N}-review-1`, 0, 5)], tasks, true, T0 + 8 * 60000)).toEqual({
      text: "Main session is between steps (last: Reviewing the fixes, round 1, 3m ago)",
      current: "re-review",
      done: false,
    }));
  it("reads Done once the stage closed", () => expect(selfReviewNow([d(`${N}-pass`, 0, 1)], tasks, false, T0).done).toBe(true));
});
