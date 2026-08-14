// Panel.test.tsx covers the panel primitive's loading/error/not-recorded
// branches -- the same branches ViewFrame.tsx already tests, carried
// forward so a view recomposed onto Panel (task 20) cannot lose the
// not-recorded distinction on the way -- plus StatPanel's own
// null-versus-zero rule (task 19, step 3), asserted as three separate
// cases so they cannot collapse into one another.
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Panel } from "./Panel";
import { StatPanel } from "./StatPanel";
import { TimeSeriesPanel } from "./TimeSeriesPanel";
import type { StatsResponse } from "../api";
import type { StatsViewState } from "../hooks/useStatsView";

function readyState<Row>(rows: Row, recorded = true): StatsViewState<Row> {
  return {
    status: "ready",
    data: {
      view: "state-board",
      from: "2026-01-01T00:00:00Z",
      to: "2026-02-01T00:00:00Z",
      boundaryConvention: "a stage run is attributed to the period containing its start instant",
      recorded,
      rows,
    } satisfies StatsResponse<Row>,
  };
}

describe("Panel", () => {
  it("renders 'Loading…' while the fetch is in flight, and never calls children", () => {
    const children = vi.fn();
    render(<Panel title="Runs" description="A panel" state={{ status: "loading" }} children={children} />);
    expect(screen.getByRole("status")).toHaveTextContent("Loading…");
    expect(children).not.toHaveBeenCalled();
  });

  it("renders the error message on a failed fetch, and never calls children", () => {
    const children = vi.fn();
    render(
      <Panel title="Runs" description="A panel" state={{ status: "error", message: "network down" }} children={children} />,
    );
    expect(screen.getByRole("alert")).toHaveTextContent("Failed to load: network down");
    expect(children).not.toHaveBeenCalled();
  });

  it("renders the not-recorded banner when the response says recorded: false, and never calls children", () => {
    const children = vi.fn();
    render(<Panel title="Runs" description="A panel" state={readyState([], false)} children={children} />);
    expect(screen.getByTestId("not-recorded")).toHaveTextContent("No data was recorded for this period.");
    expect(children).not.toHaveBeenCalled();
  });

  it("calls children with the response's data once it is ready and recorded", () => {
    render(
      <Panel title="Runs" description="A panel" state={readyState([{ id: 1 }])}>
        {(data) => <p data-testid="body">{data.rows.length} rows</p>}
      </Panel>,
    );
    expect(screen.getByTestId("body")).toHaveTextContent("1 rows");
    expect(screen.queryByTestId("not-recorded")).not.toBeInTheDocument();
  });

  it("renders the panel's title and description as chrome", () => {
    render(<Panel title="Total cost" description="Summed across the period" state={{ status: "loading" }} children={vi.fn()} />);
    expect(screen.getByRole("heading", { name: "Total cost" })).toBeInTheDocument();
    expect(screen.getByText("Summed across the period")).toBeInTheDocument();
  });
});

describe("StatPanel", () => {
  it("renders a measured zero as '0', not as unavailable", () => {
    render(<StatPanel label="Runs" value={0} />);
    expect(screen.getByTestId("measured")).toHaveTextContent("0");
    expect(screen.queryByTestId("unavailable")).not.toBeInTheDocument();
  });

  it("renders null as 'unavailable', never as 0", () => {
    render(<StatPanel label="Runs" value={null} />);
    expect(screen.getByTestId("unavailable")).toHaveTextContent("unavailable");
    expect(screen.queryByTestId("measured")).not.toBeInTheDocument();
  });

  it("renders undefined as 'unavailable', never as 0", () => {
    render(<StatPanel label="Runs" value={undefined} />);
    expect(screen.getByTestId("unavailable")).toHaveTextContent("unavailable");
    expect(screen.queryByTestId("measured")).not.toBeInTheDocument();
  });

  it("renders a measured positive value formatted, distinctly from both 0 and unavailable", () => {
    render(<StatPanel label="Total cost" value={4.5} format={(v) => `$${v.toFixed(2)}`} />);
    expect(screen.getByTestId("measured")).toHaveTextContent("$4.50");
  });

  it("renders its label exactly once when composed inside a Panel, not doubled with the panel heading", () => {
    render(
      <Panel title="Changes" state={readyState([{ id: 1 }])}>
        {() => <StatPanel label="Changes" value={10} />}
      </Panel>,
    );
    expect(screen.getAllByText("Changes")).toHaveLength(1);
    expect(screen.getByRole("heading", { name: "Changes" })).toBeInTheDocument();
  });
});

describe("TimeSeriesPanel", () => {
  it("draws a point for every measured day, and no point for a null day", () => {
    render(
      <TimeSeriesPanel
        points={[
          { day: "2026-01-01", value: 1 },
          { day: "2026-01-02", value: null },
          { day: "2026-01-03", value: 3 },
        ]}
      />,
    );
    expect(screen.getAllByTestId("time-series-point")).toHaveLength(2);
  });

  it("renders the not-recorded-style empty state when no day in the period is measured", () => {
    render(
      <TimeSeriesPanel
        points={[
          { day: "2026-01-01", value: null },
          { day: "2026-01-02", value: undefined },
        ]}
      />,
    );
    expect(screen.getByTestId("time-series-empty")).toBeInTheDocument();
    expect(screen.queryByTestId("time-series-point")).not.toBeInTheDocument();
  });
});
