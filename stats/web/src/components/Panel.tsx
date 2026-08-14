// Panel is the chrome primitive every dashboard (task 20) composes
// itself from: a bordered region with a heading, an optional
// description, and a body -- "one self-contained result with its own
// heading" (proposal.md's Fix round 1). It carries the loading, error,
// not-recorded and unmeasured branches ViewFrame.tsx used to render, so
// recomposing a view into panels cannot lose either distinction on the
// way: a period entirely before the store held anything must still read
// as "no data was recorded" rather than falling through to whatever a
// child would otherwise render for an empty result, and a period whose
// runs were recorded but never attributed must read as its own third
// state -- never folded into either neighbour (task 5, design.md's "the
// third arm of the absence distinction"). Its own tests assert exactly
// that: a panel handed `recorded: false` renders the not-recorded
// banner, a panel handed `recorded: true, unmeasured: true` renders the
// unmeasured banner, and neither ever calls `children`.
//
// This is deliberately the same branch structure ViewFrame.tsx already
// has, not a coincidentally-similar rewrite: Panel is what a view
// recomposes onto (task 20), and ViewFrame's own callers are unaffected
// until that recomposition lands, so keeping both correct in the
// meantime means neither may drift from the other by design. The two
// stay separate files rather than one sharing wrapper because ViewFrame
// renders a page-level <section> with an <h2>, while Panel renders a
// panel-level <section> with an <h3> nested inside a dashboard -- merging
// them would need a heading-level flag on the only caller left to name
// it, which is exactly the "config switch with one implementation" KISS
// violation until a second heading level actually exists.
import type { ReactNode } from "react";
import type { StatsResponse } from "../api";
import type { StatsViewState } from "../hooks/useStatsView";

export interface PanelProps<Row> {
  title: string;
  description?: string;
  state: StatsViewState<Row>;
  children: (data: StatsResponse<Row>) => ReactNode;
}

export function Panel<Row>({ title, description, state, children }: PanelProps<Row>) {
  return (
    <section className="panel" aria-label={title}>
      <h3 className="panel-title">{title}</h3>
      {description && <p className="panel-description">{description}</p>}
      <div className="panel-body">
        {state.status === "loading" && <p role="status">Loading…</p>}
        {state.status === "error" && (
          <p role="alert" className="panel-error">
            Failed to load: {state.message}
          </p>
        )}
        {state.status === "ready" && !state.data.recorded && (
          <p className="not-recorded" data-testid="not-recorded">
            No data was recorded for this period.
          </p>
        )}
        {state.status === "ready" && state.data.recorded && state.data.unmeasured && (
          <p className="unmeasured" data-testid="unmeasured">
            Runs were recorded for this period, but none carried measurements. Check your recording
            setup -- this is not the same as a quiet period.
          </p>
        )}
        {state.status === "ready" && state.data.recorded && !state.data.unmeasured && children(state.data)}
      </div>
    </section>
  );
}
