# kan-537-flow-serve-the-last-flow-mark-per-session-from — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-537-flow-serve-the-last-flow-mark-per-session-from-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** the dispatch context bundle silently failed to build during panel pre-flight, the reviewers read the plan and decision directly instead, and the rebuilt bundle gave no cause; pre-flight should refuse or name the failure rather than silently downgrade the reviewers' context — filed: KAN-593
- **[flow-fix]** panel F3 (deferred): the `project_key` COALESCE lives as two independent code copies (query.go:302, stageruns.go:667) narrated as equal in three comments with nothing enforcing the equality; hold the expression once so the copies cannot drift — filed: KAN-594

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the dispatch row reads `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
