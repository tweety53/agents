# Stage keys

Which phase file marks each key, for **Stage marks** (`skills/flow-contracts/pipeline.md`).

## Stage keys

Every stage `/flow` marks uses a `flow.*` key.

The full key list, in the order each phase file marks them:

| Phase file | Keys |
|------------|------|
| `skills/flow/brainstorm.md` | `flow.kickoff`, `flow.brainstorm` (begin), `flow.decide` |
| `skills/flow/brainstorm-planner.md` | `flow.brainstorm` (end), `flow.design-approval`, `flow.create-artifacts`, `flow.writing-plans` |
| `skills/flow/implement.md` | `flow.load-context`, `flow.isolate-workspace`, `flow.sdd-tdd` |
| `skills/flow/document-fix.md` | `flow.document-fix`, `flow.decide` (fix runs) — loaded only on a fix run |
| `skills/flow/cross-repo-worktrees.md` | `flow.isolate-workspace` — loaded only when the change links a peer, an `## apps` entry names a repository holding no worktree for it, `## visual verification` names a `regression checkout`, or the resolved set spans more than one repository |
| `skills/flow/sdd-dispatch.md` | `flow.sdd-tdd` — loaded only when the decision's `execution` is `sdd` |
| `skills/flow/gated-review-fix.md` | `flow.sdd-tdd` — loaded only when a gated reviewer pass closes `fix` |
| `skills/flow/review-panel.md` | `flow.review-panel` |
| `skills/flow/review-panel-late-fix.md` | `flow.review-panel` — loaded only on a fix run |
| `skills/flow/review-panel-fix-round.md` | `flow.review-panel` — loaded only once a round has recorded a Critical or Important finding, or a close guard sends the run back to the handback loop |
| `skills/flow/review-panel-optional-slots.md` | `flow.review-panel` — loaded only for a round whose roster carries `mutation` |
| `skills/flow/review-panel-experimental-slot.md` | `flow.review-panel` — loaded only for a round whose roster carries an `exp-` slot |
| `skills/flow/verify-and-handoff.md` | `flow.verify`, `flow.visual-verify` (steps 1–2), `flow.stage-diff`, `flow.run-instructions`, `flow.write-in-progress` |
| `skills/flow/visual-verify.md` | `flow.visual-verify` from step 3 — loaded only when a worktree's diff matched a `ui paths` glob |
| `skills/flow/integrate.md` | `flow.preflight`, `flow.unfinished-work-gate`, `flow.landing-question`, `flow.preserve-sessions`, `flow.commit-two`, `flow.landing-routes` |
| `skills/flow/archive.md` | `flow.verify-merge`, `flow.sync-archive`, `flow.commit-archive`, `flow.cleanup`, `flow.verify-cleanup`, `flow.write-finished`, `flow.self-review`, `flow.push-archive` |
