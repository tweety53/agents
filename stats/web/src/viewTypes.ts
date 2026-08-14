// The controls every one of the eight views is driven by (design.md's
// "The views": "Project is a first-class filter on every view"; the API
// section: "from and to are required on every view, including the live
// state board"). Defined once so the eight view components share one
// prop shape instead of eight near-identical redeclarations.
//
// `model` is optional and carried here (task 19) even though no view yet
// reads it -- App.tsx's DashboardBar owns the model variable's state and
// passes it down to whichever dashboard is active, the same way it
// already passes `period` and `project`, so a view that starts honouring
// it (task 20) needs no prop-shape change to do so. A view that has not
// been recomposed yet simply does not destructure it, which is exactly
// why this field is optional rather than required.
import type { Period } from "./components/PeriodPicker";

export interface ViewProps {
  period: Period;
  project: string | undefined;
  model?: string | undefined;
}
