# Guards and hooks

The pipeline never relies on an agent saying it behaved.

## Guards

A guard is a small program whose exit code is the verdict: the diff touches the UI or it does not,
the checks are green on this commit or they are not. Each is Go in `stats/internal/guard/` with its
test beside it, run as `scripts/flow-guard <name>`; the header comment of a guard's source file is
canonical for its arguments and exit codes. Nothing is installed: the first call builds the binary
from the checkout and later calls reuse it.

Every guard run is recorded, with whether it passed, fired or could not answer, and the dashboard's
**Flow guards** view ranks them ([Dashboard](dashboard.md)).

Canonical: the `## flow-guard` section of [`stats/README.md`](../stats/README.md).

## Hooks

Hooks run inside Claude Code itself, before or after a tool call:

- `enforce-agent-baseline` denies a dispatch missing the baseline pointer, and records its row.
- `enforce-loop-dispatch` records a pipeline-fix or self-review loop dispatch.
- `block-unsafe-commands` denies edits, commits and resets on a main checkout's default branch.
- `flow-active-change` routes a plain problem report to the open change as a fix run.
- `await-own-agents` holds a subagent's stop while an agent it launched still runs.
- `flow-session-activity` reports each session's working or waiting state to `flowd`.
- `flow-skill-load` records to `flowd` each skill, rule or doc file of this repository a session loads.

Each is `flow-guard hook <name>` (`stats/internal/guard/hook/`), run through the
`~/.claude/hooks/flow-guard-hook` binary `setup.sh` builds from `stats/cmd/flow-guard-hook`. `setup.sh` prints the snippet that
registers them; it never registers them for you.
