# Rendering the session records

The outcome table for `myflow record render`, and what each outcome means for the caller.

**Loaded by `/myflow-do`, `/myflow-finish` and `/myflow-fast`** — on the `prUrl` commit path and in run 1.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `skills/flow-contracts/session-records-rationale.md`;
**a `/myflow-*` run never loads it.**

## Rendering the session records

`/myflow-do` reads this table on its `prUrl` commit path, and `/myflow-finish` reads it in
run 1; the invocation of `myflow record render` itself is described by each caller.

| Outcome | What it means | What you do |
|---------|---------------|-------------|
| `rendered: <dest>`, exit 0 | The record was written to that path from the store's rows. | Nothing. Proceed. |
| `MISSING: <kind> — no rows for <change>`, exit 0 | The store holds no rows of that kind for this change. | **Report it.** It means no record exists, never that one was written elsewhere. Proceed. |
| `journalled: <kind>`, exit 0 | The store could not be reached, so nothing was rendered. | **Report it, naming the kind.** Proceed with the integration. |
| A message on stderr, **exit non-zero** | A destination was refused or could not be written. | **Report it, with the command's own message.** Then proceed. |

**A non-zero exit is never silent and never a stop; the remaining kinds are still attempted after
any one failure, and the handoff names which records were rendered and which were not.** **Neither
`MISSING:` nor `journalled:` is non-zero** — non-zero keeps its one meaning, a destination refused or
a write that failed, so a caller branching on exit status never reads an empty record, or an
unreachable store, as a failure. See **Preserving the session records** (`skills/flow-contracts/session-records-rationale.md`) —
which keeps its former name — for why.
