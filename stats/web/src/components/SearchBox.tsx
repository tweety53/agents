// SearchBox drives DataTable's free-text term. It is a controlled input
// with no debounce or local fetch of its own -- DataTable filters the
// array it already holds in memory (see DataTable.tsx's header comment for
// why that is not the server-side search the task's own requirement
// covers), so there is no request to throttle.
export interface SearchBoxProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export function SearchBox({ value, onChange, placeholder = "Search…" }: SearchBoxProps) {
  return (
    <input
      type="search"
      aria-label="Search"
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  );
}
