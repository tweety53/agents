// DataTable is the one table component task 13 requires: sortable column
// headers, a filter bar, a search box and paging controls, shared by every
// view that lists rows so sorting behaves identically across all eight.
//
// It sorts, filters and pages the `rows` array it is given **in memory**,
// deliberately -- and that is not the client-side filtering the task's own
// non-negotiable requirement forbids. That rule guards GET /api/v1/changes
// and GET /api/v1/stage-runs, whose filter/sort/search/page query
// parameters (task 3.1's allowlist, task 11's list endpoints) exist
// precisely so a caller never re-implements them. The eight statistics
// views this component actually renders are served by
// GET /api/v1/stats/{view} (internal/api/stats.go's statsQueryParams),
// whose *only* parameters are "from", "to", "project", "breakdown" and
// "change" -- the server performs no sort, no filter and no paging on that
// response at all, and returns the whole period-bounded result in one
// shot. There is therefore nothing server-side to duplicate: every value
// DataTable's controls operate on already reached the browser through the
// server's own period/project restriction (never re-derived here), and
// sorting or paging that fixed, already-small array locally is the only
// mechanism available. What must never happen -- and does not, anywhere in
// this file or its callers -- is deriving `from`/`to`/`project` from
// anything other than a request to the server.
import { Fragment, useMemo, useState, type ReactNode } from "react";
import { FilterBar } from "./FilterBar";

export interface Column<T> {
  /** Stable key identifying the column; also the sort/filter key. */
  key: string;
  header: string;
  sortable?: boolean;
  /** The raw value used for sorting, filtering and the default search text. */
  accessor: (row: T) => string | number | null | undefined;
  /** Custom cell content; defaults to String(accessor(row)) when omitted. */
  render?: (row: T) => ReactNode;
  /** When set, FilterBar offers an exact-match dropdown for this column. */
  filterable?: boolean;
}

export interface DataTableProps<T> {
  columns: Column<T>[];
  rows: T[];
  rowKey: (row: T) => string;
  /** Column keys searched by the free-text box; defaults to every column. */
  searchableKeys?: string[];
  pageSize?: number;
  emptyMessage?: string;
  /** Optional expandable detail rendered beneath a row when toggled open. */
  renderDetail?: (row: T) => ReactNode;
  detailLabel?: string;
}

type SortDirection = "asc" | "desc";

function compareValues(a: string | number | null | undefined, b: string | number | null | undefined): number {
  // Absent values sort last regardless of direction -- the caller flips
  // the comparator's sign for desc, and a null/undefined must not migrate
  // to the front just because the sign flipped, per spec: "A record whose
  // metrics lack the key being sorted on SHALL order as absent, distinct
  // from a record whose metric is recorded as zero."
  const aMissing = a === null || a === undefined;
  const bMissing = b === null || b === undefined;
  if (aMissing && bMissing) return 0;
  if (aMissing) return 1;
  if (bMissing) return -1;
  if (typeof a === "number" && typeof b === "number") return a - b;
  return String(a).localeCompare(String(b));
}

export function DataTable<T>({
  columns,
  rows,
  rowKey,
  searchableKeys,
  pageSize = 20,
  emptyMessage = "No rows for this period.",
  renderDetail,
  detailLabel = "Details",
}: DataTableProps<T>) {
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<SortDirection>("asc");
  const [search, setSearch] = useState("");
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [page, setPage] = useState(0);
  const [openDetail, setOpenDetail] = useState<string | null>(null);

  const searchCols = searchableKeys ?? columns.map((c) => c.key);

  const filtered = useMemo(() => {
    let out = rows;
    if (search.trim()) {
      const term = search.trim().toLowerCase();
      out = out.filter((row) =>
        columns
          .filter((c) => searchCols.includes(c.key))
          .some((c) => String(c.accessor(row) ?? "").toLowerCase().includes(term)),
      );
    }
    for (const [field, value] of Object.entries(filters)) {
      if (!value) continue;
      const col = columns.find((c) => c.key === field);
      if (!col) continue;
      out = out.filter((row) => String(col.accessor(row) ?? "") === value);
    }
    return out;
  }, [rows, search, filters, columns, searchCols]);

  const sorted = useMemo(() => {
    if (!sortKey) return filtered;
    const col = columns.find((c) => c.key === sortKey);
    if (!col) return filtered;
    const copy = [...filtered];
    copy.sort((a, b) => {
      const cmp = compareValues(col.accessor(a), col.accessor(b));
      return sortDir === "asc" ? cmp : -cmp;
    });
    return copy;
  }, [filtered, sortKey, sortDir, columns]);

  const pageCount = Math.max(1, Math.ceil(sorted.length / pageSize));
  const clampedPage = Math.min(page, pageCount - 1);
  const paged = sorted.slice(clampedPage * pageSize, clampedPage * pageSize + pageSize);

  function toggleSort(key: string) {
    if (sortKey === key) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortKey(key);
      setSortDir("asc");
    }
    setPage(0);
  }

  const filterableColumns = columns.filter((c) => c.filterable);
  const filterBarColumns = filterableColumns.map((c) => ({
    key: c.key,
    header: c.header,
    options: Array.from(new Set(rows.map((r) => String(c.accessor(r) ?? "")))).filter(Boolean),
  }));

  return (
    <div className="data-table">
      <FilterBar
        search={search}
        onSearchChange={(v) => {
          setSearch(v);
          setPage(0);
        }}
        columns={filterBarColumns}
        filters={filters}
        onFilterChange={(key, value) => {
          setFilters((f) => ({ ...f, [key]: value }));
          setPage(0);
        }}
      />

      {sorted.length === 0 ? (
        <p className="data-table-empty">{emptyMessage}</p>
      ) : (
        // Every table this component renders now sits inside a Panel's
        // bordered, padded body (task 20's recomposition) rather than
        // spanning a whole page -- a wide table (many columns, or a long
        // formatted value) can therefore overflow the panel's own width.
        // Scrolling that overflow inside its own container, rather than
        // letting it push the page wide, is this component's own job:
        // every view composes tables through DataTable, so fixing it here
        // fixes it everywhere at once (DRY, engineering-principles.md).
        <div className="data-table-scroll">
          <table>
            <thead>
              <tr>
                {renderDetail && <th aria-hidden="true" />}
                {columns.map((c) => (
                  <th key={c.key}>
                    {c.sortable ? (
                      <button type="button" onClick={() => toggleSort(c.key)}>
                        {c.header}
                        {sortKey === c.key ? (sortDir === "asc" ? " ▲" : " ▼") : ""}
                      </button>
                    ) : (
                      c.header
                    )}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {paged.map((row) => {
                const key = rowKey(row);
                const isOpen = openDetail === key;
                return (
                  <Fragment key={key}>
                    <tr>
                      {renderDetail && (
                        <td>
                          <button
                            type="button"
                            aria-expanded={isOpen}
                            aria-label={`${isOpen ? "Hide" : "Show"} ${detailLabel}`}
                            onClick={() => setOpenDetail(isOpen ? null : key)}
                          >
                            {isOpen ? "▾" : "▸"}
                          </button>
                        </td>
                      )}
                      {columns.map((c) => (
                        <td key={c.key}>{c.render ? c.render(row) : String(c.accessor(row) ?? "")}</td>
                      ))}
                    </tr>
                    {renderDetail && isOpen && (
                      <tr className="data-table-detail-row">
                        <td colSpan={columns.length + 1}>{renderDetail(row)}</td>
                      </tr>
                    )}
                  </Fragment>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {pageCount > 1 && (
        <div className="data-table-pager">
          <button type="button" disabled={clampedPage === 0} onClick={() => setPage((p) => p - 1)}>
            Previous
          </button>
          <span>
            Page {clampedPage + 1} of {pageCount}
          </span>
          <button type="button" disabled={clampedPage >= pageCount - 1} onClick={() => setPage((p) => p + 1)}>
            Next
          </button>
        </div>
      )}
    </div>
  );
}
