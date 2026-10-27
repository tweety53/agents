# kan-1040-copyable-next-command

## Why

Every `/flow*` handoff ends on the next command, `/flow <name>`, as a bare line. In the Claude
mobile app that line has to be selected and copied by hand; a fenced code block gets a one-tap copy
button there, and in any terminal or desktop client that offers one.

A creating `/flow` run whose arguments also asked to "create a jira task" filed the ticket and
ended its turn (2026-10-27, KAN-1040 itself): nothing in **Resolution** says a requested issue is
filed and then planned on, so the run read the filing as the whole request.

## What changes

- Every handoff prints its next command alone in its own fenced code block, the last thing in the
  output. `/clear` stays a bare line above it; the `**Field:**` lines stay rendered Markdown.
- A creating run asked to create a Jira issue files it, links it, and continues straight into the
  In Progress transition and planning — never a stop.
- `/flow-status` is unchanged: its next commands sit in table cells, where a fence cannot go.
