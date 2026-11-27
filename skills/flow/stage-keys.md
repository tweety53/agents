# Stage keys

Which phase file marks each key, for **Stage marks** (`skills/flow-contracts/pipeline.md`).

## Stage keys

Every stage `/flow` marks uses one of seven `flow.*` keys, one per phase. A fix round re-runs its
phase under the same key; no sub-step has a mark of its own.

| Mark | Begins | Ends | Retired keys it covers |
|------|--------|------|------------------------|
| `flow.plan` | `skills/flow/brainstorm.md`, at kickoff | the plan gate | `flow.kickoff`, `flow.brainstorm`, `flow.create-artifacts`, `flow.writing-plans`, `flow.decide` |
| `flow.implement` | `skills/flow/implement.md`, at the implementing run's start (also `cross-repo-worktrees.md`, `fix.md` re-runs) | the last task closed | `flow.load-context`, `flow.isolate-workspace`, `flow.sdd-tdd` |
| `flow.review` | `skills/flow/review-panel.md`, at the first pass (also `review-panel-fix-round.md`, `review-panel-optional-slots.md`) | the last fix round | `flow.review-panel` |
| `flow.verify` | `skills/flow/verify-and-handoff.md`, at the lint/test list (visual verify, `visual-verify.md`, runs inside it; `fix.md` re-runs it) | `IN_PROGRESS` written | `flow.visual-verify`, `flow.stage-diff`, `flow.run-instructions`, `flow.write-in-progress` |
| `flow.land` | `flow-guard integrate-change`, through `skills/flow/integrate.md` | the landing route done | `flow.preflight`, `flow.unfinished-work-gate`, `flow.landing-question`, `flow.preserve-sessions`, `flow.commit-two`, `flow.landing-routes` |
| `flow.self-review` | `flow-guard integrate-change`, after land | the self-review pass closed | none |
| `flow.cleanup` | `flow-guard integrate-change`, through `skills/flow/cleanup.md` | `FINISHED` written | `flow.verify-merge`, `flow.write-finished`, `flow.refresh-main-checkout` |

`Flow-Key:` commit trailers keep their own names and are not stage marks.
