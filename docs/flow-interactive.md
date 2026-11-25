# /flow-interactive

A conversation instead of a plan. One invocation is one session: one change, one Jira issue, one
worktree, and a stream of requests, each shipped to the base branch on its own.

For each message that asks for something, a question included:

1. **Look** at the live state first, the page, the rows behind it, the log, then the code.
2. **Implement** it test-first.
3. **Verify** with the project's lint and the tests the change touches.
4. **Land** it by merge and push, and reinstall.
5. **Check the live page** when the request touched the UI.

A request too big to land as one small step is put to you: run it here, or file it as a Jira task.

The session you talk to only looks, records, dispatches and reports: one subagent per request, on a
pair the [pair judge](roles-and-models.md#the-pair-judge) picks, does steps 2–5. There is no review
panel per request and no plan file. A question that needs no code is just answered. Every request is
recorded, so cost and timeline read as they do for `/flow`.

Saying you are done ends the session: the in-flight request lands, one
[self-review](self-review.md) pass covers the whole session, and cleanup runs.

Canonical: [`skills/flow-interactive/SKILL.md`](../skills/flow-interactive/SKILL.md).
