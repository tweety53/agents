import { describe, expect, it } from "vitest";
import { taskCount } from "./taskCount";

describe("taskCount", () => {
  it("derives both counts from the plan when present, over stale stored fields", () =>
    expect(
      taskCount({
        doneTasks: 6,
        currentTasks: 6,
        tasks: [
          { n: 1, title: "a", state: "done" },
          { n: 2, title: "b", state: "in progress" },
          { n: 3, title: "c", state: "pending" },
        ],
      }),
    ).toEqual({ done: 1, total: 3 }));
  it("falls back to the stored fields without a plan", () => expect(taskCount({ doneTasks: 2, currentTasks: 5 })).toEqual({ done: 2, total: 5 }));
  it("leaves the total absent when nothing was observed", () => expect(taskCount({})).toEqual({ done: 0, total: undefined }));
});
