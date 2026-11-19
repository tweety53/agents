import { expect, it } from "vitest";
import { canonicalStage } from "./phase";
import { stageLabel } from "./stageLabel";

it("names a stage in plain words", () => {
  expect(stageLabel("flow.implement")).toBe("Implementing");
});

it("names a retired stage by the slim stage that replaced it", () => {
  expect(canonicalStage("flow.sdd-tdd")).toBe("flow.implement");
  expect(canonicalStage("flow.write-in-progress")).toBe("flow.verify");
  expect(canonicalStage("flow.land")).toBe("flow.land");
  expect(stageLabel("flow.plan-review")).toBe("Planning the change");
});
