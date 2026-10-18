# self-review-backlog-flow-5-oct18

## Why

Five "Backlog Flow" Jira tickets, planned as self-review findings; each holds against the trees at
agents `52438f0d`, gymie `eb06468c` and gymie-playwright `eb0dbb6` (`design.md`, **Context**). Seven
other backlog tickets checked on the way no longer hold and carry no task.

- KAN-913: a workspace `flowd` replays the shared on-disk journal into its own `flow_<id>`
  database; a replayed entry leaves the journal and is dropped with that database at cleanup.
- KAN-773: `setup.sh global` run from a linked worktree only prints a warning, then links the
  operator's `~/.claude` into a tree cleanup will delete.
- KAN-818: a gymie-playwright worktree gets no install step, so Node resolves `@playwright/test`
  from the main checkout's `node_modules`.
- KAN-780: gymie's `## lint` never says a one-invocation verify bundle must hold read-only tasks
  only, so the next bundle author re-derives Gradle's parallel-ordering failure.
- KAN-799: `LocalDevSeederIntegrationTest.seedIsIdempotent` fails whenever `DEV_INVITE_EMAIL` is
  exported, as the documented dev setup does.

## What changes

- `flowd` replays the journal only when it serves the dev workspace's `flow` database.
- `setup.sh global` into the operator's own home from a linked worktree is refused, naming the
  main checkout's command; a sandboxed `$HOME` installs as before.
- gymie-playwright declares `## worktree setup` running `npm ci`.
- gymie's `## lint` states the read-only bundle constraint.
- `LocalDevSeederIntegrationTest` pins `app.dev.invite-email` unset.
