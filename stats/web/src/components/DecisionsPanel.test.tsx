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
      "Tasks 1, 2not run yet2 builds on 1",
      "Task 3not run yeta big task",
      "Review: primary, failure-modesnot run yetnew I/O",
      "Visual check: requirednot run yetthe SPA changes",
    ]);
  });
});

describe("DecisionsPanel with a plan-check decision", () => {
  it("shows the pair each point's own dispatch was given, or not run yet, with the reason beneath", () => {
    const d = {
      groups: [
        { tasks: ["1", "2"], reason: "same schema foundation" },
        { tasks: ["3"], reason: "independent" },
      ],
      panel: [{ slots: ["primary", "principles"], reason: "large change" }],
      visual: { verify: "required", reason: "UI changes" },
    };
    const run = (key: string, model: string, effort: string, startedAt: string) => ({ id: 1, key, role: "x", model, effort, startedAt });
    render(
      <DecisionsPanel
        decision={d}
        dispatches={[
          run("task-1+2-implementer", "sonnet", "low", "2026-01-01T00:00:00Z"),
          run("panel-1-primary+principles", "opus", "low", "2026-01-01T01:00:00Z"),
        ]}
      />,
    );
    const rows = screen.getAllByTestId("decision-row").map((r) => r.textContent);
    expect(rows).toEqual([
      "Tasks 1, 2sonnet · lowsame schema foundation",
      "Task 3not run yetindependent",
      "Review: primary, principlesopus · lowlarge change",
      "Visual check: requirednot run yetUI changes",
    ]);
  });

  it("the visual check shows the pair of a verifier dispatched per repository", () => {
    const run = (key: string, startedAt: string) => ({ id: 1, key, role: "verifier", model: "sonnet", effort: "low", startedAt });
    render(<DecisionsPanel decision={{ visual: { verify: "required", reason: "UI" } }} dispatches={[run("visual-verify-gymie", "2026-01-01T00:00:00Z")]} />);
    expect(screen.getAllByTestId("decision-row").map((r) => r.textContent)).toEqual(["Visual check: requiredsonnet · lowUI"]);
  });

  it("a task the run bundled with another still shows the bundle's pair", () => {
    const run = (key: string) => ({ id: 1, key, role: "x", model: "sonnet", effort: "low", startedAt: "2026-01-01T00:00:00Z" });
    render(<DecisionsPanel decision={{ groups: [{ tasks: ["21"], reason: "a" }, { tasks: ["22"], reason: "b" }, { tasks: ["2"], reason: "c" }] }} dispatches={[run("task-21+22-implementer")]} />);
    expect(screen.getAllByTestId("decision-row").map((r) => r.textContent)).toEqual(["Task 21sonnet · lowa", "Task 22sonnet · lowb", "Task 2not run yetc"]);
  });

  it("a review slot dispatched per repository shows its pair", () => {
    const run = (key: string, model: string) => ({ id: 1, key, role: "x", model, effort: "high", startedAt: "2026-01-01T00:00:00Z" });
    render(
      <DecisionsPanel
        decision={{ panel: [{ slots: ["primary", "principles"], reason: "big" }, { slots: ["mutation"], reason: "m" }] }}
        dispatches={[run("panel-0-primary+principles-backend", "opus"), run("panel-0-primary+principles-frontend", "opus"), run("panel-0-mutation", "sonnet")]}
      />,
    );
    expect(screen.getAllByTestId("decision-row").map((r) => r.textContent)).toEqual(["Review: primary, principlesopus · highbig", "Review: mutationsonnet · highm"]);
  });
});
