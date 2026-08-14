// ChangeVariable is the dashboard bar's "change" template variable, and
// it is deliberately a *navigation* variable, not a filter: selecting a
// change opens that change's own dashboard (#/run/<project>/<change>)
// rather than restricting the rows on the dashboard currently open, the
// same way a Grafana variable can drive a drilldown rather than a query
// restriction (proposal.md's Fix round 1). It lists changes through the
// same GET /api/v1/changes endpoint the state board's own table already
// uses, scoped to the current project filter when one is set, so it
// never invents a second source of truth for "which changes exist".
import { useEffect, useState, type ChangeEvent } from "react";
import { listChanges, type ChangeDTO } from "../api";

export interface ChangeVariableProps {
  project: string | undefined;
  /**
   * Called with the selected change's own project and name. Defaults to
   * navigating there directly (window.location.hash), the same route
   * StateBoard.tsx's own row links already build; overridable so a test
   * can observe the navigation without depending on jsdom's location
   * plumbing.
   */
  onSelect?: (project: string, change: string) => void;
}

const CHANGE_LIST_LIMIT = 200;

/**
 * The field this component sorts the change list by -- a real column name
 * from the server's own allowlist (store.AllowedChangeFields(), mirrored
 * into api.ts's CHANGE_QUERY_FIELDS), never a DTO field name. Exported so
 * api.test.ts's membership test asserts against exactly the string this
 * component actually sends, not a copy of it: task 26's defect was this
 * same field sorting by "updatedAt" (ChangeDTO's own field name), which
 * the server's allowlist has never accepted and returns 400 for.
 */
export const CHANGE_VARIABLE_SORT_FIELD = "updated_at";

function runDetailHash(project: string, change: string): string {
  return `#/run/${encodeURIComponent(project)}/${encodeURIComponent(change)}`;
}

export function ChangeVariable({ project, onSelect }: ChangeVariableProps) {
  const [changes, setChanges] = useState<ChangeDTO[]>([]);

  useEffect(() => {
    let cancelled = false;
    listChanges({
      filters: project ? { project } : undefined,
      sort: [{ field: CHANGE_VARIABLE_SORT_FIELD, desc: true }],
      limit: CHANGE_LIST_LIMIT,
    })
      .then((res) => {
        if (!cancelled) setChanges(res.changes);
      })
      .catch(() => {
        if (!cancelled) setChanges([]);
      });
    return () => {
      cancelled = true;
    };
  }, [project]);

  function handleChange(e: ChangeEvent<HTMLSelectElement>) {
    const index = e.target.value === "" ? -1 : Number(e.target.value);
    const change = changes[index];
    if (!change) return;
    if (onSelect) {
      onSelect(change.projectKey, change.name);
    } else {
      window.location.hash = runDetailHash(change.projectKey, change.name);
    }
    // Reset to the placeholder: this is a navigation trigger, not a value
    // the control should keep holding once it has fired.
    e.target.value = "";
  }

  return (
    <label className="dashboard-variable change-variable">
      Go to change
      <select aria-label="Go to change" value="" onChange={handleChange}>
        <option value="" disabled>
          Select a change…
        </option>
        {changes.map((c, index) => (
          <option key={`${c.projectKey}/${c.name}`} value={index}>
            {c.projectKey}/{c.name}
          </option>
        ))}
      </select>
    </label>
  );
}
