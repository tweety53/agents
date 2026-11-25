# The stats service and its dashboard

`flowd` on `127.0.0.1:4173` with PostgreSQL behind it is the store every `/flow*` command writes
through the `flow` CLI: change state, stage runs and cost, findings, dispatches, decisions, and Jira
transitions (off until the three `FLOWD_JIRA_*` variables are set). When the daemon is down, writes
go to an on-disk journal and replay later.

The web UI at http://127.0.0.1:4173 has the state board, stage leaderboard, trend, weekly and daily
costs, reviewers, reviewer findings, role costs, self-review, confidence, flow guards, and per-run
detail and run-cost pages. The **Agent roles** view is also where each role's models are edited
([Roles and models](roles-and-models.md)).

The **Confidence** view is a review queue of every estimate below 100% (**To review**, **Approved**,
**Declined**): each shows its gaps, split into those that need a decision and those that close by
running the work, and is approved or declined with an optional reason.

Setup, the launchd agent, the UI-test stack, pricing and Jira credentials:
[`stats/README.md`](../stats/README.md).
