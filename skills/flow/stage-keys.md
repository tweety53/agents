# Stage keys

Which phase file marks each key, for **Stage marks** (`skills/flow-contracts/pipeline.md`).

## Stage keys

Every stage `/flow` marks uses a `flow.*` key.

The full key list, in the order each phase file marks them:

| Phase file | Keys |
|------------|------|
| `skills/flow/brainstorm.md` | `flow.kickoff`, `flow.brainstorm`, `flow.design-approval`, `flow.create-artifacts`, `flow.writing-plans`, `flow.decide` |
| `skills/flow/implement.md` | `flow.load-context`, `flow.isolate-workspace`, `flow.document-fix`, `flow.decide` (fix runs), `flow.sdd-tdd` |
| `skills/flow/review-panel.md` | `flow.review-panel` |
| `skills/flow/review-panel-optional-slots.md` | `flow.review-panel` — loaded only for a round whose roster carries `mutation` or an `exp-` slot |
| `skills/flow/verify-and-handoff.md` | `flow.verify`, `flow.visual-verify` (steps 1–2), `flow.stage-diff`, `flow.run-instructions`, `flow.write-in-progress` |
| `skills/flow/visual-verify.md` | `flow.visual-verify` from step 3 — loaded only when a worktree's diff matched a `ui paths` glob |
| `skills/flow/integrate.md` | `flow.preflight`, `flow.unfinished-work-gate`, `flow.landing-question`, `flow.preserve-sessions`, `flow.commit-two`, `flow.landing-routes` |
| `skills/flow/archive.md` | `flow.verify-merge`, `flow.sync-archive`, `flow.commit-archive`, `flow.cleanup`, `flow.verify-cleanup`, `flow.write-finished`, `flow.self-review`, `flow.push-archive` |
