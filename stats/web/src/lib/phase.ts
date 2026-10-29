import type { StateBoardRow } from "../api";

/** The run view's phases in pipeline order -- each named by its first stage's place in internal/stages/names.go's Table. */
export const PHASES = [
  "Planning",
  "Decisions",
  "Implementation",
  "Fixes",
  "Review",
  "Verification",
  "Handoff",
  "Integrate",
  "Self-review",
  "Landing",
  "Cleanup",
] as const;
export type Phase = (typeof PHASES)[number];

/** Every stage key the store can record (internal/stages/names.go, plus the retired flow.push-archive older runs carry), each in exactly one phase. */
export const STAGE_PHASE: Record<string, Phase> = {
  "plan.session": "Planning",
  "flow.kickoff": "Planning",
  "flow.brainstorm": "Planning",
  "flow.create-artifacts": "Planning",
  "flow.writing-plans": "Planning",
  "flow.design-approval": "Decisions",
  "flow.decide": "Decisions",
  "flow.load-context": "Implementation",
  "flow.isolate-workspace": "Implementation",
  "flow.sdd-tdd": "Implementation",
  "flow.document-fix": "Fixes",
  "flow.pipeline-fix": "Fixes",
  "flow.review-panel": "Review",
  "flow.verify": "Verification",
  "flow.visual-verify": "Verification",
  "flow.stage-diff": "Handoff",
  "flow.run-instructions": "Handoff",
  "flow.write-in-progress": "Handoff",
  "flow.preflight": "Integrate",
  "flow.unfinished-work-gate": "Integrate",
  "flow.landing-question": "Integrate",
  "flow.preserve-sessions": "Integrate",
  "flow.commit-two": "Integrate",
  "flow.sync-archive": "Integrate",
  "flow.commit-archive": "Integrate",
  "flow.push-archive": "Integrate",
  "flow.self-review": "Self-review",
  "flow.landing-routes": "Landing",
  "flow.verify-merge": "Landing",
  "flow.cleanup": "Cleanup",
  "flow.verify-cleanup": "Cleanup",
  "flow.write-finished": "Cleanup",
  "flow.refresh-main-checkout": "Cleanup",
};

/** The phases in pipeline order, the one holding the open stage first. */
export function phaseOrder(openStage: string | undefined): Phase[] {
  const open = openStage === undefined ? undefined : STAGE_PHASE[openStage];
  return open ? [open, ...PHASES.filter((p) => p !== open)] : [...PHASES];
}

// A STARTED change whose open or latest stage is planning or decisions, with
// no implementer dispatch yet, is still planning.
const PLANNING = new Set(Object.keys(STAGE_PHASE).filter((k) => STAGE_PHASE[k] === "Planning" || STAGE_PHASE[k] === "Decisions"));

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
