import { expect, it } from "vitest";
import { stageLabel } from "./stageLabel";

it("names the plan review's stage in plain words", () => {
  expect(stageLabel("flow.plan-review")).toBe("Reviewing the plan");
});
