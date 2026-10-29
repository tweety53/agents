// Plain-words names for the pipeline's stage ids, for the board's cards and
// stage column; an id not listed falls back to itself.
const LABELS: Record<string, string> = {
  "plan.session": "Planning session",
  "flow.kickoff": "Starting up",
  "flow.load-context": "Loading context",
  "flow.brainstorm": "Designing the change",
  "flow.design-approval": "Waiting for design approval",
  "flow.create-artifacts": "Writing the spec",
  "flow.writing-plans": "Writing the task plan",
  "flow.decide": "Choosing models and reviewers",
  "flow.isolate-workspace": "Setting up the worktree",
  "flow.preflight": "Pre-flight checks",
  "flow.sdd-tdd": "Implementing tasks",
  "flow.review-panel": "Code review",
  "flow.verify": "Running lint and tests",
  "flow.visual-verify": "Checking the UI",
  "flow.stage-diff": "Staging the diff",
  "flow.run-instructions": "Writing test instructions",
  "flow.write-in-progress": "Handing over for review",
  "flow.unfinished-work-gate": "Checking for unfinished work",
  "flow.self-review": "Self-review",
  "flow.pipeline-fix": "Pipeline fix",
  "flow.document-fix": "Documenting a fix",
  "flow.landing-question": "Asking how to land",
  "flow.landing-routes": "Landing the change",
  "flow.verify-merge": "Checking the merge",
  "flow.commit-two": "Committing the work",
  "flow.commit-archive": "Committing the archive",
  "flow.push-archive": "Pushing the archive",
  "flow.sync-archive": "Archiving the change",
  "flow.preserve-sessions": "Saving session logs",
  "flow.refresh-main-checkout": "Updating the main checkout",
  "flow.cleanup": "Cleaning up",
  "flow.verify-cleanup": "Checking the cleanup",
  "flow.write-finished": "Finishing",
};

export function stageLabel(stage: string): string {
  return LABELS[stage] ?? stage;
}
