import type { StateBoardRow } from "../api";

/** The run view's phases in pipeline order -- each named by its first stage's place in internal/stages/names.go's Table. */
export const PHASES = ["Planning", "Implementation", "Review", "Verification", "Landing", "Self-review", "Cleanup"] as const;
export type Phase = (typeof PHASES)[number];

/** Every stage key the store can record (internal/stages/names.go), each in exactly one phase. */
export const STAGE_PHASE: Record<string, Phase> = {
  "plan.session": "Planning",
  "flow.plan": "Planning",
  "flow.implement": "Implementation",
  "flow.review": "Review",
  "flow.land": "Landing",
  "flow.kickoff": "Planning",
  "flow.brainstorm": "Planning",
  "flow.create-artifacts": "Planning",
  "flow.writing-plans": "Planning",
  "flow.plan-review": "Planning",
  "flow.design-approval": "Planning",
  "flow.decide": "Planning",
  "flow.load-context": "Implementation",
  "flow.isolate-workspace": "Implementation",
  "flow.sdd-tdd": "Implementation",
  "flow.document-fix": "Implementation",
  "flow.pipeline-fix": "Implementation",
  "flow.review-panel": "Review",
  "flow.verify": "Verification",
  "flow.visual-verify": "Verification",
  "flow.stage-diff": "Verification",
  "flow.run-instructions": "Verification",
  "flow.write-in-progress": "Verification",
  "flow.preflight": "Landing",
  "flow.unfinished-work-gate": "Landing",
  "flow.landing-question": "Landing",
  "flow.preserve-sessions": "Landing",
  "flow.commit-two": "Landing",
  "flow.self-review": "Self-review",
  "flow.landing-routes": "Landing",
  "flow.verify-merge": "Cleanup",
  "flow.cleanup": "Cleanup",
  "flow.write-finished": "Cleanup",
  "flow.refresh-main-checkout": "Cleanup",
};

// A STARTED change whose open or latest stage is planning, with
// no implementer dispatch yet, is still planning.
const PLANNING = new Set(Object.keys(STAGE_PHASE).filter((k) => STAGE_PHASE[k] === "Planning"));

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
