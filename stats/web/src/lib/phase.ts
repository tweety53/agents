import type { StateBoardRow } from "../api";

// The stages a change runs before any implementation; a STARTED change whose
// open or latest stage is one of these, with no implementer dispatch yet, is
// still planning.
const PLANNING = new Set([
  "plan.session",
  "flow.kickoff",
  "flow.brainstorm",
  "flow.design-approval",
  "flow.create-artifacts",
  "flow.writing-plans",
  "flow.decide",
]);

/** A change's phase in plain words, derived from its stored state and the stages it ran: STARTED splits into Planning and Implementing; IN_PROGRESS is Integrating while a stage is open (it runs straight on into self-review and landing), Ready for review once none is. */
export function phase(row: Pick<StateBoardRow, "state" | "currentStage" | "lastStage" | "doneTasks">): string {
  switch (row.state) {
    case "IN_PROGRESS":
      return row.currentStage === undefined ? "Ready for review" : "Integrating";
    case "FINISHED":
      return "Finished";
    case "STARTED": {
      const stage = row.currentStage ?? row.lastStage;
      return stage === undefined || (PLANNING.has(stage) && row.doneTasks === undefined) ? "Planning" : "Implementing";
    }
    default:
      return row.state;
  }
}
