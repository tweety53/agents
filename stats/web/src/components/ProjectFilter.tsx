// ProjectFilter is the project-scope control every view shares. It is a
// free-text field rather than a dropdown sourced from the server: no
// endpoint in api.ts lists distinct project keys (GET /api/v1/changes
// lists changes, not projects, and paginates), and inventing a
// client-side aggregation over it would be exactly the kind of
// client-side re-derivation task 13's own non-negotiable requirement
// warns against for the parameters the server already restricts by.
// Leaving the field empty aggregates across every project, per
// design.md's "Project is a first-class filter" ("all projects" scenario:
// no project filter, response identifies which project each row came
// from -- every view's rows already carry projectKey or are project-scoped
// by construction).
export interface ProjectFilterProps {
  project: string | undefined;
  onChange: (project: string | undefined) => void;
}

export function ProjectFilter({ project, onChange }: ProjectFilterProps) {
  return (
    // "dashboard-variable" is the shared styling hook DashboardBar.tsx's
    // other template variables (ModelVariable, ChangeVariable) also carry,
    // so this control reads as one of the bar's variables rather than a
    // visually distinct leftover from the page-per-table chrome (task 19,
    // step 1: "No component hard-codes a colour" -- shared class, shared
    // tokens, one look).
    <label className="dashboard-variable project-filter">
      Project
      <input
        type="text"
        aria-label="Project filter"
        placeholder="All projects"
        value={project ?? ""}
        onChange={(e) => onChange(e.target.value.trim() === "" ? undefined : e.target.value)}
      />
    </label>
  );
}
