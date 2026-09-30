// RunDetail is the route the eight aggregate views were missing (Fix
// round 1's own rationale, proposal.md): "#/run/<project>/<change>" opens
// one change on a timeline of its stage runs across every command that
// touched it, above a table carrying one row per stage run. Its header
// numbers come from useRunDetail's own cost-per-change aggregate, never
// from summing the stage runs this component renders -- see
// hooks/useRunDetail.ts's own header comment for why that matters.
//
// Recomposed onto the panel primitives (task 20, step 4): the header's
// definition list becomes five stat panels, the timeline and the table
// each get their own panel. This file builds its own minimal panel chrome
// (RunPanel, below) rather than importing components/Panel.tsx: Panel's
// `state` prop is typed to useStatsView's own StatsViewState<Row> --
// {status, data: StatsResponse<Row>} -- and RunDetailState's "ready" case
// carries {stageRuns, summary} instead of one `data` envelope, so the two
// are structurally different unions. Panel.tsx is outside this task's own
// file list, so its shape is not this task's to change; RunPanel
// reproduces the same DOM and the same "panel"/"panel-title"/"panel-body"
// classes Panel itself renders, which is what keeps every dashboard
// visually one system without forcing a shared generic across two
// genuinely different data shapes (WET, engineering-principles.md -- an
// abstraction bent to fit a second, differently-shaped caller is more
// expensive than the small amount of markup repeated here).
//
// **Settling the model variable on this route (task 20's own decision to
// make):** honoured, not disabled. A stage run's `models` bag already
// records which models it used (task 22), and cost-per-change already
// accepts a `model` restriction (task 21) -- both cheap to wire here, so
// the dashboard bar's model control filters this route's stage-run table
// to runs that used the selected model and scopes the header to that
// model's own buckets, via useRunDetail's own `model` parameter, rather
// than being left rendering and doing nothing (this round's own
// non-negotiable: "do not leave it rendering and inert").
import { type ReactNode } from "react";
import { StageRunTable } from "../components/StageRunTable";
import { StageTimeline } from "../components/StageTimeline";
import { StatPanel } from "../components/StatPanel";
import { useRunDetail } from "../hooks/useRunDetail";
import { formatInt, formatMs, formatUsd } from "../format";
import { projectLabel } from "../lib/projectLabel";

export interface RunDetailProps {
  project: string;
  change: string;
  /** The dashboard bar's model variable, honoured on this route -- see
   * this file's own header comment for why. Undefined means "every
   * model", the same as on every other dashboard. */
  model?: string;
}

function RunPanel({ title, description, children }: { title: string; description?: string; children: ReactNode }) {
  return (
    <section className="panel" aria-label={title}>
      <h3 className="panel-title">{title}</h3>
      {description && <p className="panel-description">{description}</p>}
      <div className="panel-body">{children}</div>
    </section>
  );
}

// Per-dispatch cost (task 5, KAN-201) rides StageRunTable's own per-row
// detail toggle -- see components/StageRunTable.tsx's own header comment
// on that table's dispatch section for why it moved there (F3, pass 1 of
// this change's own review panel): it used to be a second, hand-rolled
// expand affordance built locally in this file, rendered as a flat list
// below the whole table with a toggle identifier synthesized from
// `startedAt` because nothing here tied a toggle back to its own row.
// Nested in StageRunTable instead, each toggle is keyed by `stageRunId`
// (already unique per row) through the same mechanism MetricsDetail
// already used, so there is nothing left for this file to build.

export function RunDetail({ project, change, model }: RunDetailProps) {
  const state = useRunDetail(project, change, model);

  return (
    <section aria-label={`Run detail: ${project}/${change}`} className="dashboard">
      <h2>{change}</h2>
      {/* The header names the project (kan-183), it does not key it -- the
          section's own aria-label above and useRunDetail's request both
          keep the full key, since that is the route's identity and not
          this line's to shorten. */}
      <p className="view-description">
        Project {projectLabel(project)}
        {model && ` · Model ${model}`}
      </p>

      {state.status === "loading" && <p role="status">Loading…</p>}
      {state.status === "error" && (
        <p role="alert" className="view-error">
          Failed to load: {state.message}
        </p>
      )}
      {state.status === "ready" && (
        <>
          <div className="dashboard-stat-row">
            <RunPanel title="Runs">
              <StatPanel label="Runs" value={state.summary.runCount} format={formatInt} />
            </RunPanel>
            <RunPanel title="Measured">
              <StatPanel label="Measured" value={state.summary.measuredRuns} format={formatInt} />
            </RunPanel>
            <RunPanel title="Total cost">
              <StatPanel label="Total cost" value={state.summary.totalCostUsd} format={formatUsd} />
            </RunPanel>
            <RunPanel title="Total input tokens">
              <StatPanel label="Total input tokens" value={state.summary.totalTokensInput} format={formatInt} />
            </RunPanel>
            <RunPanel title="Total duration">
              <StatPanel label="Total duration" value={state.summary.totalDurationMs} format={formatMs} />
            </RunPanel>
          </div>

          <RunPanel title="Stage timeline">
            <StageTimeline stageRuns={state.stageRuns} />
          </RunPanel>

          <RunPanel title="Stage runs">
            <StageRunTable
              stageRuns={state.stageRuns}
              truncated={state.truncated}
              totalStageRuns={state.totalStageRuns}
            />
          </RunPanel>
        </>
      )}
    </section>
  );
}
