import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { DataTable, type Column } from "./DataTable";
import { defaultPeriod } from "./PeriodPicker";

interface Row {
  id: string;
  name: string;
  cost: number;
}

const rows: Row[] = [
  { id: "b", name: "Beta", cost: 30 },
  { id: "a", name: "Alpha", cost: 10 },
  { id: "c", name: "Gamma", cost: 20 },
];

const columns: Column<Row>[] = [
  { key: "name", header: "Name", sortable: true, accessor: (r) => r.name, filterable: true },
  { key: "cost", header: "Cost", sortable: true, accessor: (r) => r.cost },
];

function bodyRowTexts() {
  const table = screen.getByRole("table");
  const body = within(table).getAllByRole("rowgroup")[1];
  return within(body)
    .getAllByRole("row")
    .map((r) => within(r).getAllByRole("cell").map((c) => c.textContent));
}

describe("DataTable", () => {
  it("renders every row's actual cell values, in the table's default order", () => {
    render(<DataTable columns={columns} rows={rows} rowKey={(r) => r.id} />);
    expect(bodyRowTexts()).toEqual([
      ["Beta", "30"],
      ["Alpha", "10"],
      ["Gamma", "20"],
    ]);
  });

  it("sorts ascending then descending on header click, by the column's real value", async () => {
    const user = userEvent.setup();
    render(<DataTable columns={columns} rows={rows} rowKey={(r) => r.id} />);

    await user.click(screen.getByRole("button", { name: /Cost/ }));
    expect(bodyRowTexts().map((r) => r[1])).toEqual(["10", "20", "30"]);

    await user.click(screen.getByRole("button", { name: /Cost/ }));
    expect(bodyRowTexts().map((r) => r[1])).toEqual(["30", "20", "10"]);
  });

  it("orders a row missing the sorted value as absent, distinct from a recorded zero", async () => {
    interface Nullable {
      id: string;
      cost: number | null;
    }
    const nullableRows: Nullable[] = [
      { id: "zero", cost: 0 },
      { id: "missing", cost: null },
      { id: "five", cost: 5 },
    ];
    const nullableColumns: Column<Nullable>[] = [
      { key: "id", header: "Id", accessor: (r) => r.id },
      { key: "cost", header: "Cost", sortable: true, accessor: (r) => r.cost },
    ];
    const user = userEvent.setup();
    render(<DataTable columns={nullableColumns} rows={nullableRows} rowKey={(r) => r.id} />);

    await user.click(screen.getByRole("button", { name: /Cost/ }));
    expect(bodyRowTexts().map((r) => r[0])).toEqual(["zero", "five", "missing"]);
  });

  it("filters rows by the free-text search box across accessor values", async () => {
    const user = userEvent.setup();
    render(<DataTable columns={columns} rows={rows} rowKey={(r) => r.id} />);

    await user.type(screen.getByRole("searchbox"), "alpha");
    expect(bodyRowTexts()).toEqual([["Alpha", "10"]]);
  });

  it("filters rows by an exact-match column filter", async () => {
    const user = userEvent.setup();
    render(<DataTable columns={columns} rows={rows} rowKey={(r) => r.id} />);

    await user.selectOptions(screen.getByLabelText("Filter by Name"), "Gamma");
    expect(bodyRowTexts()).toEqual([["Gamma", "20"]]);
  });

  it("pages results without dropping or duplicating a row", async () => {
    const user = userEvent.setup();
    render(<DataTable columns={columns} rows={rows} rowKey={(r) => r.id} pageSize={2} />);

    expect(bodyRowTexts()).toHaveLength(2);
    await user.click(screen.getByRole("button", { name: "Next" }));
    expect(bodyRowTexts()).toHaveLength(1);
    expect(screen.getByRole("button", { name: "Next" })).toBeDisabled();
  });

  it("shows a detail panel for a row only when its toggle is expanded", async () => {
    const user = userEvent.setup();
    render(
      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(r) => r.id}
        renderDetail={(r) => <div data-testid={`detail-${r.id}`}>repo detail for {r.name}</div>}
      />,
    );

    expect(screen.queryByTestId("detail-a")).not.toBeInTheDocument();
    await user.click(screen.getAllByRole("button", { name: /Show Details/ })[1]); // Alpha row, default order
    expect(screen.getByTestId("detail-a")).toBeInTheDocument();
    expect(screen.getByText("repo detail for Alpha")).toBeInTheDocument();
  });

  it("renders the empty-period message rather than an empty table when there are no rows", () => {
    render(<DataTable columns={columns} rows={[]} rowKey={(r) => r.id} emptyMessage="Nothing here" />);
    expect(screen.getByText("Nothing here")).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });
});

// Task 4: an empty view says whether the period is why. A fixed clock, not
// a second `new Date()` at assertion time -- the latter is flaky by
// construction (the two calls can straddle a millisecond boundary).
// DashboardBar.test.tsx's "task 3" describe block uses the same pattern.
describe("DataTable: empty state names the period as a possible cause (task 4)", () => {
  const NOW = new Date("2026-03-15T14:30:00.000Z");

  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(NOW);
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("shows the period hint and a reset control when empty under a non-default period", () => {
    const onPeriodChange = vi.fn();
    const nonDefaultPeriod = { from: new Date("2025-06-01T00:00:00Z"), to: new Date("2025-06-15T00:00:00Z") };
    render(
      <DataTable
        columns={columns}
        rows={[]}
        rowKey={(r) => r.id}
        emptyMessage="No changes updated in this period."
        period={nonDefaultPeriod}
        onPeriodChange={onPeriodChange}
      />,
    );

    // The original wording stays where it is already right.
    expect(screen.getByText("No changes updated in this period.")).toBeInTheDocument();
    expect(screen.getByText(/the period may be why/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /reset/i }));
    expect(onPeriodChange).toHaveBeenCalledExactlyOnceWith(defaultPeriod());
  });

  it("shows neither the period hint nor a reset control when empty under the default period", () => {
    render(
      <DataTable
        columns={columns}
        rows={[]}
        rowKey={(r) => r.id}
        emptyMessage="No changes updated in this period."
        period={defaultPeriod()}
        onPeriodChange={vi.fn()}
      />,
    );

    expect(screen.getByText("No changes updated in this period.")).toBeInTheDocument();
    expect(screen.queryByText(/the period may be why/i)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /reset/i })).not.toBeInTheDocument();
  });

  // Regression for the bug this task's review found: `defaultPeriod()`
  // reads "now" fresh on every call, but `period` here is a fixed
  // snapshot taken when it was set. Before the fix, comparing the two for
  // exact equality meant a period that was the default at mount read as
  // non-default on any later re-render, which surfaced the "period may be
  // why" hint and reset button on a genuinely empty view -- exactly the
  // false blame design.md's "Empty states" rules out.
  it("stays absent across a re-render once time has moved on, while the period is still the default", () => {
    const onPeriodChange = vi.fn();
    const period = defaultPeriod();
    const { rerender } = render(
      <DataTable
        columns={columns}
        rows={[]}
        rowKey={(r) => r.id}
        emptyMessage="No changes updated in this period."
        period={period}
        onPeriodChange={onPeriodChange}
      />,
    );
    expect(screen.queryByText(/the period may be why/i)).not.toBeInTheDocument();

    vi.setSystemTime(new Date(NOW.getTime() + 3000));
    rerender(
      <DataTable
        columns={columns}
        rows={[]}
        rowKey={(r) => r.id}
        emptyMessage="No changes updated in this period."
        period={period}
        onPeriodChange={onPeriodChange}
        pageSize={10}
      />,
    );

    expect(screen.queryByText(/the period may be why/i)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /reset/i })).not.toBeInTheDocument();
  });
});
