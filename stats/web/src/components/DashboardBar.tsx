// DashboardBar is the Grafana-shaped chrome's one shared control strip
// (proposal.md's Fix round 1: "a top bar carrying a time-range picker
// and template-variable dropdowns that scope every panel beneath them").
// App.tsx renders it exactly once, above whichever dashboard is active,
// and owns the period/project/model state itself; every panel beneath it
// receives those values as props rather than holding a copy of its own.
// That is what makes the spec's "no two panels on a dashboard can be
// showing different periods or different filters at the same moment" a
// property of there being one control, true by construction, rather than
// something an effect has to keep several copies in sync with.
//
// Every one of the nine dashboards this bar sits above is now recomposed
// onto the panel primitives (task 20), including the run-detail route --
// which honours the model variable (RunDetail.tsx's own header comment on
// why), so this bar's model control is never rendered inert anywhere in
// this build. `modelDisabled` stays as the one exception: the live state
// board's rows are changes, not stage runs, and the server rejects the
// restriction there by design.
import { PeriodPicker, type Period } from "./PeriodPicker";
import { ProjectFilter } from "./ProjectFilter";
import { ModelVariable } from "./ModelVariable";
import { ChangeVariable } from "./ChangeVariable";

export interface DashboardBarProps {
  period: Period;
  onPeriodChange: (period: Period) => void;
  project: string | undefined;
  onProjectChange: (project: string | undefined) => void;
  model: string | undefined;
  onModelChange: (model: string | undefined) => void;
  /**
   * True on a dashboard whose rows are not stage runs -- the live state
   * board today -- where the server rejects a "model" restriction with
   * 400 rather than silently ignoring it
   * (specs/myflow-stats-views/spec.md, "A model restriction on the live
   * state board"). The bar must not offer a control whose only possible
   * outcome is that error, so the variable is disabled instead, with
   * `modelDisabledReason` stated alongside it.
   */
  modelDisabled: boolean;
  modelDisabledReason?: string;
}

export function DashboardBar({
  period,
  onPeriodChange,
  project,
  onProjectChange,
  model,
  onModelChange,
  modelDisabled,
  modelDisabledReason,
}: DashboardBarProps) {
  return (
    <div className="dashboard-bar" role="toolbar" aria-label="Dashboard controls">
      <div className="dashboard-bar-variables">
        <ProjectFilter project={project} onChange={onProjectChange} />
        <ModelVariable
          period={period}
          project={project}
          model={model}
          onChange={onModelChange}
          disabled={modelDisabled}
          disabledReason={modelDisabledReason}
        />
        <ChangeVariable project={project} />
      </div>
      <PeriodPicker period={period} onChange={onPeriodChange} />
    </div>
  );
}
