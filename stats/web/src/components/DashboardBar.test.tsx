// DashboardBar.test.tsx covers the range and variable handlers, the
// model variable's population from the server and its disabled state on
// the live state board (task 19's own Tests entry). `fetchModels` and
// `listChanges` are mocked at the module boundary -- the same pattern
// views.test.tsx already uses for `fetchStatsView` -- so these tests
// assert this component's own wiring rather than the network layer.
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { DashboardBar } from "./DashboardBar";
import type { ChangeDTO, ModelsResponse } from "../api";

const { fetchModelsMock, listChangesMock } = vi.hoisted(() => ({
  fetchModelsMock: vi.fn(),
  listChangesMock: vi.fn(),
}));

vi.mock("../api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api")>();
  return { ...actual, fetchModels: fetchModelsMock, listChanges: listChangesMock };
});

const period = { from: new Date("2026-01-01T00:00:00Z"), to: new Date("2026-02-01T00:00:00Z") };

function modelsResponse(models: string[]): ModelsResponse {
  return { from: period.from.toISOString(), to: period.to.toISOString(), models };
}

const changeA: ChangeDTO = {
  projectKey: "kan-16-myflow-stats-app",
  name: "kan-16-myflow-stats-app",
  state: "IN_PROGRESS",
  updatedAt: "2026-01-15T10:00:00Z",
  updatedBy: "alice",
};
const changeB: ChangeDTO = {
  projectKey: "kan-16-myflow-stats-app",
  name: "kan-7-myflow-principles-panel-jira",
  state: "STARTED",
  updatedAt: "2026-01-10T10:00:00Z",
  updatedBy: "alice",
};

beforeEach(() => {
  fetchModelsMock.mockReset();
  fetchModelsMock.mockResolvedValue(modelsResponse(["claude-opus-5", "claude-sonnet-5"]));
  listChangesMock.mockReset();
  listChangesMock.mockResolvedValue({ total: 2, changes: [changeA, changeB] });
});

function baseProps() {
  return {
    period,
    onPeriodChange: vi.fn(),
    project: undefined as string | undefined,
    onProjectChange: vi.fn(),
    model: undefined as string | undefined,
    onModelChange: vi.fn(),
    modelDisabled: false,
    modelDisabledReason: "The live state board's rows are changes, not stage runs, so a model restriction cannot apply here.",
  };
}

describe("DashboardBar: time range", () => {
  it("calls onPeriodChange with the new range when the start is changed", () => {
    const props = baseProps();
    render(<DashboardBar {...props} />);

    const start = screen.getByLabelText("Period start");
    fireEvent.change(start, { target: { value: "2026-01-05T00:00" } });

    expect(props.onPeriodChange).toHaveBeenCalledTimes(1);
    const [lastCall] = props.onPeriodChange.mock.calls[0];
    expect(lastCall.to).toEqual(period.to);
    expect(lastCall.from).toEqual(new Date("2026-01-05T00:00"));
  });
});

describe("DashboardBar: the model variable", () => {
  it("is populated from GET /api/v1/models for the current period and project", async () => {
    const props = baseProps();
    props.project = "kan-16-myflow-stats-app";
    render(<DashboardBar {...props} />);

    await waitFor(() => expect(fetchModelsMock).toHaveBeenCalled());
    expect(fetchModelsMock).toHaveBeenCalledWith(
      expect.objectContaining({ from: period.from, to: period.to, project: "kan-16-myflow-stats-app" }),
    );
    expect(await screen.findByRole("option", { name: "claude-opus-5" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "claude-sonnet-5" })).toBeInTheDocument();
  });

  it("is enabled on a stage-run dashboard", () => {
    const props = baseProps();
    props.modelDisabled = false;
    render(<DashboardBar {...props} />);
    expect(screen.getByLabelText("Model")).not.toBeDisabled();
  });

  it("is disabled, with a stated reason, on the live state board", () => {
    const props = baseProps();
    props.modelDisabled = true;
    render(<DashboardBar {...props} />);
    expect(screen.getByLabelText("Model")).toBeDisabled();
    expect(screen.getByText(props.modelDisabledReason)).toBeInTheDocument();
  });

  it("does not fetch the model list at all while disabled", async () => {
    const props = baseProps();
    props.modelDisabled = true;
    render(<DashboardBar {...props} />);
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(fetchModelsMock).not.toHaveBeenCalled();
  });
});

describe("DashboardBar: the change variable navigates rather than filters", () => {
  it("selecting a change calls onSelect with that change's project and name, not a filter handler", async () => {
    // ChangeVariable is rendered by DashboardBar with no injected onSelect
    // (its production default), so this exercises the real navigation
    // path via window.location.hash rather than a filter callback --
    // proving the control is a navigation variable, not a scoping one.
    const user = userEvent.setup();
    const props = baseProps();
    render(<DashboardBar {...props} />);

    const select = await screen.findByLabelText("Go to change");
    await user.selectOptions(select, screen.getByRole("option", { name: `${changeB.projectKey}/${changeB.name}` }));

    expect(window.location.hash).toBe(
      `#/run/${encodeURIComponent(changeB.projectKey)}/${encodeURIComponent(changeB.name)}`,
    );
    // Neither onProjectChange nor onModelChange -- the controls that
    // actually filter the dashboard -- were touched by this selection.
    expect(props.onProjectChange).not.toHaveBeenCalled();
    expect(props.onModelChange).not.toHaveBeenCalled();

    window.location.hash = "";
  });
});
