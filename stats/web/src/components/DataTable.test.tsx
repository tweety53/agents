import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { DataTable, type Column } from "./DataTable";

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
