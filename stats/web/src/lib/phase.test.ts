// phase splits STARTED into Planning and Implementing by the stages run, and
// names IN_PROGRESS for what it is: staged and waiting for review.
import { describe, expect, it } from "vitest";
import { phase } from "./phase";

describe("phase", () => {
  it("reads a STARTED change in planning stages as Planning", () =>
    expect(phase({ state: "STARTED", lastStage: "flow.decide" })).toBe("Planning"));
  it("reads a STARTED change with no stage yet as Planning", () => expect(phase({ state: "STARTED" })).toBe("Planning"));
  it("reads a STARTED change past planning as Implementing", () =>
    expect(phase({ state: "STARTED", currentStage: "flow.review-panel" })).toBe("Implementing"));
  it("reads an in-run re-decide after implementation began as Implementing", () =>
    expect(phase({ state: "STARTED", lastStage: "flow.decide", doneTasks: 5 })).toBe("Implementing"));
  it("reads IN_PROGRESS as Ready for review", () => expect(phase({ state: "IN_PROGRESS" })).toBe("Ready for review"));
  it("reads FINISHED as Finished", () => expect(phase({ state: "FINISHED" })).toBe("Finished"));
});
