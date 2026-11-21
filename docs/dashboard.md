# The stats service and its dashboard

`flowd` on `127.0.0.1:4173` with PostgreSQL behind it is the store every `/flow*` command writes
through the `flow` CLI: change state, stage runs and cost, findings, dispatches, decisions, and Jira
transitions (off until the three `FLOWD_JIRA_*` variables are set). When the daemon is down, writes
go to an on-disk journal and replay later.

The web UI at http://127.0.0.1:4173 has the state board, stage leaderboard, cost trends, reviewers,
role costs, self-review fixes, flow guards and a per-run detail page. The **Agent roles** view is
also where each role's models are edited ([Roles and models](roles-and-models.md)).

Setup, the launchd agent, the UI-test stack, pricing and Jira credentials:
[`stats/README.md`](../stats/README.md).
