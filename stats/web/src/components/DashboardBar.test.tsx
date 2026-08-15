// DashboardBar.test.tsx covers the range and variable handlers, the
// model variable's population from the server and its disabled state on
// the live state board (task 19's own Tests entry). `fetchModels` and
// `listChanges` are mocked at the module boundary -- the same pattern
// views.test.tsx already uses for `fetchStatsView` -- so these tests
// assert this component's own wiring rather than the network layer.
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { DashboardBar } from "./DashboardBar";
import { defaultPeriod } from "./PeriodPicker";
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
// A key that actually carries the documented -[0-9a-f]{8} suffix
// (kan-183) -- changeA's and changeB's own projectKey ("kan-16-...-app")
// ends in "app", which projectLabel never strips, so an assertion built
// only on those two would pass identically whether or not the option
// label actually shortens anything.
const changeHashed: ChangeDTO = {
  projectKey: "agents-a740d89c",
  name: "kan-1",
  state: "IN_PROGRESS",
  updatedAt: "2026-01-20T10:00:00Z",
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

describe("DashboardBar: period presets (task 3)", () => {
  // A fixed clock, not a second `new Date()` at assertion time -- the
  // latter is flaky by construction (the two calls can straddle a
  // millisecond boundary). App.test.tsx's "task 2" describe block uses
  // the same pattern.
  const NOW = new Date("2026-03-15T14:30:00.000Z");
  const DAY_MS = 24 * 60 * 60 * 1000;

  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(NOW);
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  function todayMidnight(): Date {
    const midnight = new Date(NOW);
    midnight.setHours(0, 0, 0, 0);
    return midnight;
  }

  it.each([
    ["Today", todayMidnight()],
    ["7 days", new Date(NOW.getTime() - 7 * DAY_MS)],
    ["30 days", new Date(NOW.getTime() - 30 * DAY_MS)],
    ["90 days", new Date(NOW.getTime() - 90 * DAY_MS)],
    ["All time", new Date("2020-01-01T00:00:00Z")],
  ])("selecting %s sets the period from that bound to now", (name, expectedFrom) => {
    const props = baseProps();
    render(<DashboardBar {...props} />);

    fireEvent.click(screen.getByRole("button", { name }));

    expect(props.onPeriodChange).toHaveBeenCalledTimes(1);
    const [applied] = props.onPeriodChange.mock.calls[0];
    expect(applied.from).toEqual(expectedFrom);
    expect(applied.to).toEqual(NOW);
  });

  it("marks a preset active when the current period equals that preset's range", () => {
    const props = baseProps();
    props.period = { from: new Date(NOW.getTime() - 7 * DAY_MS), to: NOW };
    render(<DashboardBar {...props} />);

    expect(screen.getByRole("button", { name: "7 days" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "Today" })).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByRole("button", { name: "30 days" })).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByRole("button", { name: "90 days" })).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByRole("button", { name: "All time" })).toHaveAttribute("aria-pressed", "false");
  });

  it("marks no preset active when the period is a hand-typed range", () => {
    const props = baseProps();
    props.period = { from: new Date("2025-06-01T00:00:00Z"), to: new Date("2025-06-15T00:00:00Z") };
    render(<DashboardBar {...props} />);

    for (const name of ["Today", "7 days", "30 days", "90 days", "All time"]) {
      expect(screen.getByRole("button", { name })).toHaveAttribute("aria-pressed", "false");
    }
  });

  // Regression for the bug this task's review found: `range()` and
  // `defaultPeriod()` both read "now" fresh on every call, but the period
  // this bar was given is a fixed snapshot from the moment it was set.
  // Before the fix, comparing the two for exact equality meant a preset
  // read as active only in the instant it was picked -- any later
  // re-render (an unrelated prop change, here) recomputed a `now` a few
  // seconds further on and flipped `aria-pressed` back to false even
  // though the period on screen never changed.
  it("keeps a preset marked active across a re-render once time has moved on", () => {
    const props = baseProps();
    props.period = { from: new Date(NOW.getTime() - 7 * DAY_MS), to: NOW };
    const { rerender } = render(<DashboardBar {...props} />);

    expect(screen.getByRole("button", { name: "7 days" })).toHaveAttribute("aria-pressed", "true");

    vi.setSystemTime(new Date(NOW.getTime() + 5000));
    rerender(<DashboardBar {...props} project="kan-16-myflow-stats-app" />);

    expect(screen.getByRole("button", { name: "7 days" })).toHaveAttribute("aria-pressed", "true");
  });
});

describe("DashboardBar: reset to default (task 3)", () => {
  const NOW = new Date("2026-03-15T14:30:00.000Z");

  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(NOW);
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("is present while the period differs from the default", () => {
    const props = baseProps();
    props.period = { from: new Date("2025-06-01T00:00:00Z"), to: new Date("2025-06-15T00:00:00Z") };
    render(<DashboardBar {...props} />);

    expect(screen.getByRole("button", { name: "Reset to default" })).toBeInTheDocument();
  });

  it("is absent while the period is the default", () => {
    const props = baseProps();
    props.period = defaultPeriod();
    render(<DashboardBar {...props} />);

    expect(screen.queryByRole("button", { name: "Reset to default" })).not.toBeInTheDocument();
  });

  it("restores the default period when clicked", () => {
    const props = baseProps();
    props.period = { from: new Date("2025-06-01T00:00:00Z"), to: new Date("2025-06-15T00:00:00Z") };
    render(<DashboardBar {...props} />);

    fireEvent.click(screen.getByRole("button", { name: "Reset to default" }));

    expect(props.onPeriodChange).toHaveBeenCalledTimes(1);
    const [applied] = props.onPeriodChange.mock.calls[0];
    expect(applied).toEqual(defaultPeriod());
  });

  // Same root cause as the preset regression above, on the other side of
  // the same comparison: before the fix, a period that was the default at
  // mount read as non-default (and so grew a reset button) the moment a
  // later re-render recomputed `defaultPeriod()` against a newer `now`.
  it("stays absent across a re-render once time has moved on, while the period is still the default", () => {
    const props = baseProps();
    props.period = defaultPeriod();
    const { rerender } = render(<DashboardBar {...props} />);

    expect(screen.queryByRole("button", { name: "Reset to default" })).not.toBeInTheDocument();

    vi.setSystemTime(new Date(NOW.getTime() + 5000));
    rerender(<DashboardBar {...props} project="kan-16-myflow-stats-app" />);

    expect(screen.queryByRole("button", { name: "Reset to default" })).not.toBeInTheDocument();
  });
});

describe("DashboardBar: manual entry keeps working (task 3)", () => {
  it("does not call onPeriodChange when the typed value is unparsable", () => {
    const props = baseProps();
    render(<DashboardBar {...props} />);

    const start = screen.getByLabelText("Period start");
    fireEvent.change(start, { target: { value: "not-a-date" } });

    expect(props.onPeriodChange).not.toHaveBeenCalled();
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

  // kan-183: the option label names the project, it does not key it. This
  // is the one assertion in the suite that can actually fail that claim --
  // every other option-label check here uses a projectKey without the
  // documented suffix, so projectLabel is a no-op for them regardless of
  // whether the shortening works.
  it("labels an option with the project's display name, not its raw key, for a key carrying the hash suffix", async () => {
    listChangesMock.mockReset();
    listChangesMock.mockResolvedValue({ total: 1, changes: [changeHashed] });
    render(<DashboardBar {...baseProps()} />);

    expect(await screen.findByRole("option", { name: "agents/kan-1" })).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: `${changeHashed.projectKey}/${changeHashed.name}` })).not.toBeInTheDocument();
  });
});
