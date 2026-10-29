// phase splits STARTED into Planning and Implementing by the stages run, and
// splits IN_PROGRESS into Integrating (a stage open) and Ready for review.
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
  it("reads IN_PROGRESS with no open stage as Ready for review", () => expect(phase({ state: "IN_PROGRESS" })).toBe("Ready for review"));
  it("reads IN_PROGRESS with an open stage as Integrating", () =>
    expect(phase({ state: "IN_PROGRESS", currentStage: "flow.self-review" })).toBe("Integrating"));
  it("reads FINISHED as Finished", () => expect(phase({ state: "FINISHED" })).toBe("Finished"));
});
