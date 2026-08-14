// ViewFrame is the page-level chrome a recomposed dashboard (task 20)
// still needs and Panel does not provide: the page heading (an <h2>,
// distinct from Panel's own nested <h3>) and its description, wrapping
// whatever panels the view composes beneath them.
//
// Before task 20 this component also carried the loading/error/not-recorded
// branches -- the same branches Panel.tsx now owns. Duplicating that branch
// structure across both files was deliberate *while the recomposition was
// in flight* (Panel.tsx's own header comment on why), but once every view
// renders through Panel instead, keeping a second copy here would be the
// exact defect this round's own instructions call out: "the not-recorded
// branch must survive the recomposition, as exactly one implementation."
// Panel is that one implementation now; this file carries none of it.
import type { ReactNode } from "react";

export interface ViewFrameProps {
  title: string;
  description: string;
  children: ReactNode;
}

export function ViewFrame({ title, description, children }: ViewFrameProps) {
  return (
    <section aria-label={title} className="dashboard">
      <h2>{title}</h2>
      <p className="view-description">{description}</p>
      {children}
    </section>
  );
}
