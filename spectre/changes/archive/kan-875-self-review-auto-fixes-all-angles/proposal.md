# kan-875-self-review-auto-fixes-all-angles

## Why

kan-829 lets a run fix a pipeline defect it hits mid-run, but only angle 1's kind and only while
the run is live. Every other self-review finding — cost, what went well, automation, stats-app,
speed — still becomes a Jira ticket and waits for a later change, though the self-review session
already holds the context to fix it. The operator wants all six angles fixed the same way, the
outcome recorded in the flow store, and the pass kept cheap — by structure, not by a weaker model
(KAN-875).

## What changes

- `/flow-self-review` fixes every finding that is not `big` (kan-829's blast-radius rule) and lands
  the fixes without asking; only `big` findings — and every product-code finding — reach the
  filing prompt, and with none the prompt asks for the rating alone.
- The fixes run in one-shot `opus` subagents, never inline (fix round 1): one fixer for every
  finding, one fresh reviewer per round over the whole branch, one fresh fixer per round of review
  findings — all on one `<agents repo>` worktree branch, one commit per finding, landed once.
- The fix loop runs end to end (fix round 1): a clean review is followed by the agents repo's full
  `## lint` and `## test` on the branch, a red check loops back through fix and review, and only a
  green branch lands on `<agents-base>` — the pass ends with every fix on agents `main`.
- Every finding's outcome — angle, finding, `fixed`/`filed`/`declined`, sha or key, blast radius —
  is a row in a new `self_review_findings` table, written by `flow self-review finding` and read by
  `flow self-review findings`.
- The report carries `fixed: <sha>` as a third disposition, accepted by
  `check-self-review-report.sh`.
