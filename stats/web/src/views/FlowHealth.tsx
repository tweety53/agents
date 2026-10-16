// Flow health -- whether flow's own machinery earns its keep, as eight
// period totals (design.md's flow-health-tiles): how many changes start and
// finish, what they cost and how long they take, how many review-panel
// rounds and findings they need, and how often a guard fires. Every other
// view breaks the pipeline down; this one answers at a glance.
//
// One endpoint (internal/api/health.go's flow-summary), read once and shown
// as tiles only. No model control applies: none of the totals counts per
// model.
import { Panel } from "../components/Panel";
import { StatPanel } from "../components/StatPanel";
import { ViewFrame } from "../components/ViewFrame";
import type { FlowSummary } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import { formatInt, formatMs, formatRatio, formatUsd } from "../format";
import type { ViewProps } from "../viewTypes";

const DAY_MS = 86_400_000;

/** A lead time runs to days; anything shorter reads in formatMs's units. */
function formatLeadTime(ms: number): string {
  return ms >= DAY_MS ? `${(ms / DAY_MS).toFixed(1)} d` : formatMs(ms);
}

export function FlowHealth({ period, project }: ViewProps) {
  const state = useStatsView<FlowSummary>("flow-summary", { from: period.from, to: period.to, project });
  const tile = (label: string, value: (s: FlowSummary) => number | null, format: (v: number) => string) => (
    <Panel title={label} state={state}>
      {(data) => <StatPanel label={label} value={value(data.rows)} format={format} />}
    </Panel>
  );

  return (
    <ViewFrame
      title="Flow health"
      description="Period totals: changes started and finished, what they cost and how long they took, review-panel rounds and findings, and guard fires."
    >
      <div className="dashboard-stat-row">
        {tile("Changes started", (s) => s.changesStarted, formatInt)}
        {tile("Changes finished", (s) => s.changesFinished, formatInt)}
        {tile("Total cost", (s) => s.totalCostUsd, formatUsd)}
        {tile("Mean cost per finished change", (s) => s.meanCostPerFinished, formatUsd)}
      </div>
      <div className="dashboard-stat-row">
        {tile("Median lead time", (s) => s.medianLeadTimeMs, formatLeadTime)}
        {tile("Mean panel rounds", (s) => s.meanPanelRounds, formatRatio)}
        <Panel title="Findings (C/I/Mi)" state={state}>
          {({ rows: s }) => (
            <StatPanel
              label="Findings (C/I/Mi)"
              value={s.critical + s.important + s.minor}
              format={() => `${s.critical}/${s.important}/${s.minor}`}
            />
          )}
        </Panel>
        {tile("Guard fires", (s) => s.guardFires, formatInt)}
      </div>
    </ViewFrame>
  );
}
