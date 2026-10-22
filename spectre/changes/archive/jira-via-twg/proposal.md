# jira-via-twg

## Why

The Atlassian MCP connector resolves a cloud ID through a remote `getAccessibleAtlassianResources`
lookup before every session's first Jira call; in kan-922's run 2 that lookup hung for over five
minutes and was aborted, leaving KAN-922 short of Done. The `twg` CLI, already installed and
authenticated, fetched and transitioned the same issue in about a second. The operator's
instruction: use `twg` for straightforward Jira interactions.

## What changes

- `skills/flow-contracts/jira-integration.md` makes `twg` the first route for every Jira call the
  pipeline makes — get, transitions, transition, update, create, comment, JQL query — whenever
  `command -v twg` finds it, and the MCP tools the route only when `twg` is absent.
