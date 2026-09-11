# kan-487-flow-panel-review-slots-can-fork-sub-agents

## Why

`/flow`'s review-panel and implementer roles all dispatch on the `flow-<model>-<effort>` agent
family (`agents/flow-*.md`), which carries no `tools:` frontmatter and therefore inherits every
tool this harness exposes, including `Agent`. Every dispatch prompt across `implement.md`,
`review-panel.md` and `verify-and-handoff.md` already tells the dispatched agent "NO DELEGATION:
… never call the `Agent` tool" — but that is a sentence in a prompt, not a structural limit. KAN-459
round 6 showed a panel slot ignoring it and forking two further `general-purpose` subagents to do
its own review legwork — the same shape KAN-449 already named for panel-fix dispatches. Uncontrolled
subagent forking burns tokens outside the conductor's/parent's closed dispatch list, invisible until
spotted live in the agent tree UI.

No role dispatched on this family is ever supposed to call `Agent` — implementer, reviewer, and
panel-fix alike carry the identical NO DELEGATION prose. Removing `Agent` from the family's tool
list closes the gap structurally for every one of those roles in one change, rather than only the
panel-review slots the ticket named.

## What changes

Every `agents/flow-<model>-<effort>.md` file — one per model/effort pair — gains an
explicit `tools:` allowlist — `Read, Glob, Grep, Bash, Write, Edit, ToolSearch` — omitting `Agent`.
This matches the existing TOOLS paragraph every dispatch prompt already carries ("every tool you
need that is not already listed … is loaded in one `select:<name>,<name>` ToolSearch"): the base set
was always meant to be narrow, with anything else (`SendMessage`, `Monitor`, an MCP tool) loaded on
demand — `Agent` simply never belongs in that on-demand set for these roles.

No dispatch mechanics change in `review-panel.md`, `implement.md` or `verify-and-handoff.md` — the
`subagent_type` values dispatched (`flow-<model>-<effort>`, and `general-purpose` on
`REVIEW_PANEL_TOGGLE: default`) are unchanged; only the `flow-*` agent definitions' own tool access
shrinks. `implement.md`'s "Dispatch sites — the parent's closed list" section gains one paragraph
documenting the mechanism the first change adds, for the panel bundle and panel-fix rows the panel
review toggle already covers, and naming that the verifier row is unaffected — prose only, no
mechanics changed. The `general-purpose` dispatch path (used when the review panel toggle is
`default`, not `dynamic`) still has full `Agent` access — out of scope here, since `general-purpose`
is a harness-provided agent type this repository does not define or own.

**A third, unrelated fix rides along in this same run's commits, at the operator's request rather
than the ticket's scope:** during this run's own brainstorming, an `AskUserQuestion` was raised for
a round that in fact carried a recommended default (the tool-allowlist approach above) — a decision
`/flow-fast`'s own auto-pick rule (`skills/flow-fast/brainstorm.md`, section B) already required
picking automatically, since architectural weight is not one of its two grounds to ask. That rule
gains one paragraph closing the loophole this run's own mistake exposed, so a future `/flow-fast`
run does not repeat it.
