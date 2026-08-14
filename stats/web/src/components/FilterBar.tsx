// FilterBar is DataTable's control strip: the free-text SearchBox plus one
// exact-match dropdown per filterable column, both operating on the rows
// DataTable already holds (see DataTable.tsx's header comment for why that
// is not a duplicate of the server-side filtering task 11 already
// performs on GET /api/v1/changes and GET /api/v1/stage-runs).
import { SearchBox } from "./SearchBox";

export interface FilterBarColumn {
  key: string;
  header: string;
  /** Every distinct value seen for this column, for the dropdown's options. */
  options: string[];
}

export interface FilterBarProps {
  search: string;
  onSearchChange: (value: string) => void;
  columns: FilterBarColumn[];
  filters: Record<string, string>;
  onFilterChange: (key: string, value: string) => void;
}

export function FilterBar({ search, onSearchChange, columns, filters, onFilterChange }: FilterBarProps) {
  return (
    <div className="data-table-controls">
      <SearchBox value={search} onChange={onSearchChange} />
      {columns.map((c) => (
        <label key={c.key} className="data-table-filter">
          {c.header}
          <select
            aria-label={`Filter by ${c.header}`}
            value={filters[c.key] ?? ""}
            onChange={(e) => onFilterChange(c.key, e.target.value)}
          >
            <option value="">All</option>
            {c.options.map((opt) => (
              <option key={opt} value={opt}>
                {opt}
              </option>
            ))}
          </select>
        </label>
      ))}
    </div>
  );
}
