import { canonicalStage } from "./phase";

// Plain-words names for the pipeline's stage ids, for the board's cards and
// stage column; a retired id reads as the stage that replaced it, and an id
// not listed falls back to itself.
const LABELS: Record<string, string> = {
  "plan.session": "Planning session",
  "flow.plan": "Planning the change",
  "flow.implement": "Implementing",
  "flow.pipeline-fix": "Pipeline fix",
  "flow.review": "Reviewing",
  "flow.verify": "Verifying",
  "flow.land": "Landing the change",
  "flow.self-review": "Self-review",
  "flow.cleanup": "Cleaning up",
};

export function stageLabel(stage: string): string {
  const key = canonicalStage(stage);
  return LABELS[key] ?? key;
}
