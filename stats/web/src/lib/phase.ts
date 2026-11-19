import type { StateBoardRow } from "../api";

/** The run view's phases in pipeline order -- each named by its first stage's place in internal/stages/names.go's Table. */
export const PHASES = ["Planning", "Implementation", "Review", "Verification", "Landing", "Self-review", "Cleanup"] as const;
export type Phase = (typeof PHASES)[number];

/** Every stage key the store can record (internal/stages/names.go's Table), each in exactly one phase. */
export const STAGE_PHASE: Record<string, Phase> = {
  "plan.session": "Planning",
  "flow.plan": "Planning",
  "flow.implement": "Implementation",
  "flow.pipeline-fix": "Implementation",
  "flow.review": "Review",
  "flow.verify": "Verification",
  "flow.land": "Landing",
  "flow.self-review": "Self-review",
  "flow.cleanup": "Cleanup",
};

/** internal/stages/names.go's Retired: a key only historical rows carry, mapped to the slim stage that replaced it. */
export const RETIRED: Record<string, string> = {
  "flow.design-approval": "flow.plan",
  "flow.plan-review": "flow.plan",
  "flow.kickoff": "flow.plan",
  "flow.brainstorm": "flow.plan",
  "flow.create-artifacts": "flow.plan",
  "flow.writing-plans": "flow.plan",
  "flow.decide": "flow.plan",
  "flow.load-context": "flow.implement",
  "flow.isolate-workspace": "flow.implement",
  "flow.document-fix": "flow.implement",
  "flow.sdd-tdd": "flow.implement",
  "flow.review-panel": "flow.review",
  "flow.visual-verify": "flow.verify",
  "flow.stage-diff": "flow.verify",
  "flow.run-instructions": "flow.verify",
  "flow.write-in-progress": "flow.verify",
  "flow.preflight": "flow.land",
  "flow.unfinished-work-gate": "flow.land",
  "flow.landing-question": "flow.land",
  "flow.preserve-sessions": "flow.land",
  "flow.commit-two": "flow.land",
  "flow.landing-routes": "flow.land",
  "flow.verify-merge": "flow.cleanup",
  "flow.write-finished": "flow.cleanup",
  "flow.refresh-main-checkout": "flow.cleanup",
};

/** A stage key as the slim pipeline names it: a retired key becomes its replacement. */
export const canonicalStage = (stage: string): string => RETIRED[stage] ?? stage;

/** The phase a stage belongs to, a retired key counted with its replacement. */
export const stagePhase = (stage: string): Phase | undefined => STAGE_PHASE[canonicalStage(stage)];


/** A change's phase in plain words, derived from its stored state and the stages it ran: STARTED splits into Planning and Implementing; IN_PROGRESS is Integrating while a stage is open (it runs straight on into self-review and landing), Ready for review once none is. */
export function phase(row: Pick<StateBoardRow, "state" | "currentStage" | "lastStage" | "doneTasks">): string {
  switch (row.state) {
    case "IN_PROGRESS":
      return row.currentStage === undefined ? "Ready for review" : "Integrating";
    case "FINISHED":
      return "Finished";
    case "STARTED": {
      const stage = row.currentStage ?? row.lastStage;
      return stage === undefined || (stagePhase(stage) === "Planning" && row.doneTasks === undefined) ? "Planning" : "Implementing";
    }
    default:
      return row.state;
  }
}
