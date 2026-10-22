## Context

Operator-requested pipeline change after kan-922: the MCP route hung on its cloud-ID lookup;
`twg jira workitem get KAN-922` returned in 0.96s and `twg jira workitem transition --id KAN-922
--transition-id Done` fired at once. The contract already named `twg` as the fallback route and
maps every MCP call to its `twg` counterpart, so only the preference order changes.

## Decisions

### Route preference

**ID:** twg-first
**Status:** active
**Chosen:** `twg` whenever `command -v twg` finds it; MCP only when it is absent.
**Considered:** MCP first with a timeout fallback to `twg` — still pays the hang on every session before falling back.

## Open questions
