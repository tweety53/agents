// The run's latest recorded decision, one line per role: the model and
// effort it was given, its reason on hover.
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { DecisionsPanel } from "./DecisionsPanel";

const decision = {
  class: "big",
  execution: "sdd",
  implementer: { model: "opus", effort: "low", reason: "effort roll 20 < 90 → low" },
  groups: [
    { model: "opus", effort: "low", reason: "implementer default", bundles: [1, 3] },
    { model: "sonnet", effort: "medium", reason: "docs only", bundles: [2] },
  ],
  fixer: { model: "opus", effort: "low", reason: "fixes named findings" },
  panel: {
    dispatches: [{ model: "opus", effort: "low", slots: ["primary", "principles"], reason: "pass-1 panel always opus" }],
    rerun_dispatch: { model: "sonnet", effort: "low", reason: "re-run reads a delta" },
  },
  extra_dispatch: { model: "sonnet", effort: "low", slots: ["test-sanity", "call-efficiency"], reason: "roll 17" },
  overrides: [{ from: "opus/flow-planner-medium", to: "opus/flow-planner-xhigh", pairs: ["planner"], instruction: "xhigh effort is advised for planning" }],
};

describe("DecisionsPanel", () => {
  it("lists each role's model and effort, its reason as the tooltip", () => {
    render(<DecisionsPanel decision={decision} />);
    const rows = screen.getAllByTestId("decision-row").map((r) => r.textContent);
    expect(rows).toEqual([
      "Planneropus · xhigh",
      "Tasks 1, 3opus · low",
      "Task 2sonnet · medium",
      "Fixeropus · low",
      "Review: primary, principlesopus · low",
      "Review re-runsonnet · low",
      "Review: test-sanity, call-efficiencysonnet · low",
    ]);
    expect(screen.getByText("Fixer").closest("[data-testid=decision-row]")!.getAttribute("title")).toBe("fixes named findings");
    expect(screen.getByRole("region", { name: "Decisions" })).toHaveTextContent("big · sdd");
  });

  it("falls back to the implementer when the decision records no groups", () => {
    render(<DecisionsPanel decision={{ implementer: { model: "opus", effort: "high", reason: "seam" } }} />);
    expect(screen.getAllByTestId("decision-row").map((r) => r.textContent)).toEqual(["Implementeropus · high"]);
  });

  it("ignores an overrides object", () => {
    render(<DecisionsPanel decision={{ implementer: { model: "opus", effort: "high", reason: "seam" }, overrides: {} }} />);
    expect(screen.getAllByTestId("decision-row").map((r) => r.textContent)).toEqual(["Implementeropus · high"]);
  });
  it("ignores an override whose pairs is prose, not a list", () => {
    render(<DecisionsPanel decision={{ implementer: { model: "opus", effort: "high", reason: "seam" }, overrides: [{ pairs: "planner → opus/flow-planner-xhigh", instruction: "use xhigh planner" }] }} />);
    expect(screen.getAllByTestId("decision-row").map((r) => r.textContent)).toEqual(["Implementeropus · high"]);
  });

  it("lists a plan-check decision's groups, review bundles and visual check with their reasons", () => {
    render(
      <DecisionsPanel
        decision={{
          groups: [{ tasks: ["1", "2"], reason: "2 builds on 1" }, { tasks: ["3"], reason: "a big task" }],
          panel: [{ slots: ["primary", "failure-modes"], reason: "new I/O" }],
          visual: { verify: "required", reason: "the SPA changes" },
          overrides: [],
        }}
      />,
    );
    expect(screen.getAllByTestId("decision-row").map((r) => r.textContent)).toEqual([
      "Tasks 1, 22 builds on 1",
      "Task 3a big task",
      "Review: primary, failure-modesnew I/O",
      "Visual check: requiredthe SPA changes",
    ]);
  });
});
