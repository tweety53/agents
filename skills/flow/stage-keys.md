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
| `skills/flow/review-panel-fix-round.md` | `flow.review-panel` — loaded only once a round has recorded a Critical or Important finding, or a close guard sends the run back to the handback loop; read by section where a citer names it |
| `skills/flow/review-panel-optional-slots.md` | `flow.review-panel` — loaded only for a round whose roster carries `mutation` |
| `skills/flow/review-panel-experimental-slot.md` | `flow.review-panel` — loaded only for a round whose roster carries an `exp-` slot |
| `skills/flow/verify-and-handoff.md` | `flow.verify`, `flow.visual-verify` (steps 1–2), `flow.stage-diff`, `flow.run-instructions`, `flow.write-in-progress` |
| `skills/flow/verify-fix-loop.md` | `flow.document-fix`, `flow.decide`, `flow.sdd-tdd`, `flow.review-panel`, `flow.verify`, `flow.visual-verify` — re-run through their own files; loaded only when a verify stage's final report carries a fixable defect; begins no mark of its own |
| `skills/flow/visual-verify.md` | `flow.visual-verify` from step 3 — loaded only when a worktree's diff matched a `ui paths` glob |
| `skills/flow/visual-verify-tooling-analysis.md` | `flow.visual-verify` — loaded only on a fix run with at least one miss |
| `skills/flow/integrate.md` | `flow.preflight`, `flow.unfinished-work-gate`, `flow.landing-question`, `flow.preserve-sessions`, `flow.commit-two`, `flow.sync-archive`, `flow.commit-archive`, `flow.self-review`, `flow.landing-routes` |
| `skills/flow/unfinished-work-gate.md` | `flow.unfinished-work-gate` — loaded only when a worktree reported `OUTSTANDING` or `VISUAL-VERIFY-MISSING`; begins no mark |
| `skills/flow/sync-onto-base.md` | `flow.landing-question` — loaded only when a worktree's `check-base-moved.sh` verdict is `MOVED`, or the merge-and-push route's push was rejected; closes the enclosing mark, begins none |
| `skills/flow/cleanup.md` | `flow.verify-merge`, `flow.cleanup`, `flow.verify-cleanup`, `flow.write-finished`, `flow.refresh-main-checkout` |
