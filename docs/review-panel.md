# Review panel

After the code is written, the whole branch is reviewed by a panel of reviewers, each a separate
agent with one job. Every reviewer runs on every change; the plan review bundles them and picks
each bundle's model ([Planning](planning.md#the-decision-how-it-is-built-and-reviewed)), the cheapest
allowed one for a reviewer the change gives little or nothing to read. A project configures
nothing by hand.

## Review panel setup

| Reviewer | Looks for |
|----------|-----------|
| `primary` | Plan alignment and senior code review. |
| `principles` | The engineering principles: small functions, clear names, no hidden coupling. |
| `test-sanity` | Would each test the diff adds fail if its behaviour broke? On the cheapest model when no test is added or changed. |
| `performance` | Call and query cost: repeats, N+1, unbounded reads, over-fetch, indexes, transactions, writes, timeouts, caching, blocking, locks, memory, payload, frontend. On the cheapest model unless the change adds a call, query, write, migration, client, cache, startup or request-path work, collection processing or frontend render. |
| `failure-modes` | Error returns, timeouts, partial writes, concurrent re-entry. On the cheapest model without I/O, persistence or concurrency. |
| `mutation` | Breaks the code on purpose to prove the tests notice. Always a bundle alone, in its own throwaway copy of the worktree. |
| `security` | Injection, authorisation, secrets, trust boundary. Added automatically to every panel, on the cheapest model unless a task's files match the project's `## security paths`. |

Reviewers are grouped into **bundles** of one or two, and each bundle is one dispatch. A
documentation-only change runs every reviewer but `primary` on the cheapest model. After a fix round only the reviewers that raised a
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
