# group-allowed-pairs-roles

## Why

The roles table — `## Allowed pairs` in `skills/flow-contracts/model-policy.md`, drawn row for row
by the Experiments view (`stats/web/src/views/Experiments.tsx`) — is one flat list in which a
role's escalation pair takes a row of its own and nothing says what a role does. Planning, for
one, is `explorer` + `plan-designer` + `planner`, yet nothing groups them. The operator asked for
the roles grouped both visually and in what flowd serves, a plain-English line per role, and a
more compact table. The operator also asked that every review-panel bundle be a role row of its
own: today the two always-on extra bundles, `test-sanity+call-efficiency` and
`mutation+failure-modes`, share one `panel-extra` row, and the experiment stats misattribute
dispatches across the shared row.

## What changes

- `## Allowed pairs` becomes one `###` table per pipeline area — Planning, Implementation, Review,
  Visual verification, Self-review, Pipeline fix — each table gaining a `Does` column carrying a
  plain-English line on what the role does.
- `panel-extra` is replaced by `panel-sanity-efficiency` (the `test-sanity+call-efficiency`
  bundle) and `panel-mutation-failure` (the `mutation+failure-modes` bundle), each on today's
  pairs; `plan-class.sh` rolls each bundle's pair off its own row, the dispatch-site table and the
  model-and-effort rules name them, and the experiment stats attribute each bundle's pass-1
  dispatches to its own role and every re-run dispatch to `panel-rerun`.
- flowd parses each role's group and description from the table and serves them on
  `GET /api/v1/roles`, `main` under a Session group.
- The Experiments view's roles table draws one compact row per role under a header row per
  group — columns Role, Does, Allowed pairs; escalation pairs folded in as a muted second line; an
  open experiment's override shown inline with its `#id` link — as the approved prototype draws.
