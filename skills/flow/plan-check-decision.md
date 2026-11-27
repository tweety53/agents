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

**2. Review panel.** Pick the reviewers the change needs from this catalog and bundle them; each
bundle is one reviewer dispatch.

| id | Reviews |
|----|---------|
| `primary` | plan alignment and senior code review — always runs |
| `principles` | the engineering principles — small functions, clear names, no hidden coupling |
| `test-sanity` | would each test the diff adds or changes fail if its behaviour broke |
| `call-efficiency` | second callers, a read repeated in one request, a query per list item |
| `failure-modes` | error returns, timeouts, partial writes, concurrent re-entry |
| `mutation` | sabotage-proofing: breaks the code on purpose to prove the tests notice |

`security` is not in this catalog: `flow-guard plan-check` adds it after your answer when a task's
`**Files:**` matches the project's `## security paths`.

- Pick a reviewer only when the change gives it something to read: no `call-efficiency` for a
  change that adds no call, query or request; no `failure-modes` without I/O, persistence or
  concurrency; no `test-sanity` when no test is added or changed; no `mutation` when the change
  adds no logic a test should catch breaking.
- `mutation` needs more than one implementer group, or a task marked `**Size:** big`: a decision of
  one group with no big task is reviewed by the panel's other slots alone, and `flow-guard plan-check`
  refuses one naming `mutation`. The same
  threshold turns the gated per-task reviewer on (`skills/flow/implement.md`).
- When every task's `**Files:**` lists documentation alone (`.md`, `.mdc`, or `.html` under
  `docs/`), the panel is `primary` alone.
- `mutation`, when picked, is a bundle alone. Every other bundle holds one or two reviewers.

**3. Visual verification.** `not configured` when the project declares no `## visual
verification` section. Otherwise `required` when anything a user sees can change — a page, a
response a page consumes, the CORS, auth or routing the frontend depends on — and `skipped` when
nothing can. A path inside a UI directory is not by itself a reason for `required`; answer genuine
doubt with `required`.

## The answer

End the answer with a line reading exactly `===DECISION===`, then the decision JSON alone, every
`reason` one line, nothing after it:

```json
{
  "groups": [{ "tasks": ["1", "2"], "reason": "…" }],
  "panel": [
    { "slots": ["primary", "principles"], "reason": "…" },
    { "slots": ["mutation"], "reason": "…" }
  ],
  "visual": { "verify": "required", "reason": "…" }
}
```
