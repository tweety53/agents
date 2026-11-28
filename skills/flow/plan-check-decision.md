<!-- The decision rules both plan-check prompts share: flow-guard plan-check appends this file to
skills/flow/plan-check-prompt.md and to skills/flow/plan-check-decide-prompt.md. -->

## The decision

**1. Implementer groups.** Every task belongs to exactly one group; each group is one implementer
dispatch that works its tasks in plan order, one commit per task.

- A group never holds tasks of two repositories (a multi-repository plan sorts its tasks under
  `## <repository>` headings).
- A `**Size:** big` task is a group alone. Any other group holds at most three tasks.
- Groups run one at a time unless their `**After:**` tasks are done and their `**Files:**` are
  disjoint, and each is a cold start — its implementer re-reads its brief, the
  plan artifacts and the code — so fill every group to the cap, splitting only where a big task,
  the cap or a repository forces it; put tasks together first when one builds on another
  (`**After:**`) or they touch the same files.
- List groups in an order every task's `**After:**` tasks come before it.

**2. Review panel.** Every reviewer of this catalog is in every panel: bundle them all, each
bundle one reviewer dispatch; a decision leaving one out is rejected.

| id | Reviews |
|----|---------|
| `primary` | plan alignment and senior code review |
| `principles` | the engineering principles — small functions, clear names, no hidden coupling |
| `test-sanity` | would each test the diff adds or changes fail if its behaviour broke |
| `performance` | call and query cost: repeats, N+1, unbounded reads, over-fetch, indexes, transactions, writes, timeouts, caching, blocking, locks, memory, payload, frontend |
| `failure-modes` | error returns, timeouts, partial writes, concurrent re-entry |
| `mutation` | sabotage-proofing: breaks the code on purpose to prove the tests notice |

`security` is not in this catalog: `flow-guard plan-check` adds it to every panel after your
answer — pairless when a task's `**Files:**` matches the project's `## security paths`, so the pair
judge picks its pair, and on the panel-reviewer role's cheapest allowed pair otherwise.

- A reviewer the change gives little or nothing to read runs on the cheapest allowed pair: a
  `performance` for a change that adds no call, query, write, migration, client, cache, startup or
  request-path work, collection processing or frontend render; a `failure-modes` without I/O,
  persistence or concurrency; a `test-sanity` when no test is added or changed; a `mutation` when
  the change adds no logic a test should catch breaking; every reviewer but `primary` when every
  task's `**Files:**` lists documentation alone (`.md`, `.mdc`, or `.html` under `docs/`).
- `mutation` is a bundle alone. Every other bundle holds one or two reviewers; a reviewer on the
  cheapest pair shares its bundle only with another on it, never with one the change needs.

**3. Visual verification.** `not configured` when the project declares no `## visual
verification` section. Otherwise `required` when anything a user sees can change — a page, a
response a page consumes, the CORS, auth or routing the frontend depends on — and `skipped` when
nothing can. A path inside a UI directory is not by itself a reason for `required`; answer genuine
doubt with `required`.

**4. Pairs.** Each bundle names the `pair` its dispatch runs on: one of the panel-reviewer role's
allowed pairs, copied
verbatim from the user message's `## Allowed pairs`. A group names no pair: the pair judge picks its implementer's at dispatch time. Weigh what the dispatch's work is, not how big it
looks: verification-only, documentation, configuration and test-only work sits at the cheap end;
concurrency, data-model, platform-interop and performance-sensitive work at the dear end. Pick the
cheapest pair near-certain to succeed on the first try. A role listed with no pairs takes no `pair`.

## The answer

End the answer with a line reading exactly `===DECISION===`, then the decision JSON alone, every
`reason` one line, nothing after it, `<pair>` one of the allowed pairs:

```json
{
  "groups": [{ "tasks": ["1", "2"], "reason": "…" }],
  "panel": [
    { "slots": ["primary", "principles"], "pair": "<pair>", "reason": "…" },
    { "slots": ["mutation"], "pair": "<pair>", "reason": "…" }
  ],
  "visual": { "verify": "required", "reason": "…" }
}
```
