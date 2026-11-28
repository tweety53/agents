# Review panel

After the code is written, the whole branch is reviewed by a panel of reviewers, each a separate
agent with one job. The plan review picks the reviewers
([Planning](planning.md#the-decision-how-it-is-built-and-reviewed)); a project configures nothing
by hand.

## Review panel setup

| Reviewer | Looks for |
|----------|-----------|
| `primary` | Plan alignment and senior code review. **Always runs.** |
| `principles` | The engineering principles: small functions, clear names, no hidden coupling. |
| `test-sanity` | Would each test the diff adds fail if its behaviour broke? Picked only when a test is added or changed. |
| `performance` | Second callers, a read repeated in one request, a query per list item. Picked only when the change adds a call, query or request. |
| `failure-modes` | Error returns, timeouts, partial writes, concurrent re-entry. Picked only with I/O, persistence or concurrency. |
| `mutation` | Breaks the code on purpose to prove the tests notice. Always a bundle alone, in its own throwaway copy of the worktree. |
| `security` | Injection, authorisation, secrets, trust boundary. Never picked by the model: added automatically when a task's files match the project's `## security paths`. |

Reviewers are grouped into **bundles** of one or two, and each bundle is one dispatch. A
documentation-only change gets `primary` alone. After a fix round only the reviewers that raised a
finding run again.

## What else holds

- **The panel runs as decided.** Nothing adds a reviewer because a diff is large or touches a
  certain area; beyond the decision, a reviewer joins only when you ask for it, for that run.
- **Reviewers only read.** Their findings go to a separate fix agent, and a defect several
  reviewers raise is fixed once.
- **One review sees everything.** A change spanning several repositories is reviewed as one
  combined diff, so a seam between two repositories is in one reviewer's view.
- **No round limit.** Fix rounds repeat until the panel closes clean.

Canonical: [`skills/flow/review-panel.md`](../skills/flow/review-panel.md) (dispatch),
[`skills/flow/review-panel-fix-round.md`](../skills/flow/review-panel-fix-round.md) (fix rounds),
[`skills/flow/plan-check-decision.md`](../skills/flow/plan-check-decision.md) (how reviewers are
picked) and each reviewer's own `skills/flow/<reviewer>-reviewer-prompt.md` (`mutation`'s brief is in
the dispatch file).
