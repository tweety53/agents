## Context

`review-panel.md`'s roster table spawns `subagent_type: flow-<model>-<effort>` on
`REVIEW_PANEL_TOGGLE: dynamic`, or `general-purpose` on `default`, for the panel-bundle and
panel-fix rows. The `flow-<model>-<effort>` agent definitions in `agents/` carry no `tools:`
frontmatter, so Claude Code grants them every tool the harness exposes, `Agent` included. KAN-459
observed a panel slot using that access to fork two further subagents instead of doing its own
review reading — the same pattern KAN-449 already named for panel-fix. Every dispatch prompt already
carries a "NO DELEGATION: never call the `Agent` tool" paragraph, but a prompt sentence is advisory,
not a capability limit. (The verifier row is unaffected by any toggle — it always dispatches
`general-purpose`, per `verify-and-handoff.md`.)

Claude Code's agent `tools:` frontmatter is a positive allowlist only — surveyed every agent
definition on this machine (this repo's own, the installed plugins', `Explore`'s) and found no
exclusion/negation syntax anywhere. The fix has to enumerate what to keep, not what to drop.

## Decisions

### Strip `Agent` from the whole `flow-<model>-<effort>` family, not just the panel-review roster

**ID:** strip-agent-whole-family
**Status:** active
**Chosen:** add `tools: Read, Glob, Grep, Bash, Write, Edit, ToolSearch` to all 9
`agents/flow-<model>-<effort>.md` files — one change removes `Agent` for every role dispatched on
this family (reviewer, implementer, panel-fix), matching the NO DELEGATION prose every one of their
dispatch prompts already carries.
**Considered:** a second, read-only-only agent family (e.g. `flow-review-<model>-<effort>.md`,
mirroring the ticket's own "use `Explore`" example) scoped to the pure-reading review roles alone.
Ruled out: `bugbot`/`mutation` mutate code in their throwaway worktree and would still need a third,
edit-capable variant; the ticket's own audit item ("check whether the same gap applies to panel-fix
and the verifier dispatch") is left open under that split, where the whole-family change closes it
in the same edit. The whole-family change is also the smaller diff — each existing file touched
once, versus an equal number of new files plus a `review-panel.md` dispatch-mechanics rewrite
naming the new type per role.

### Tool list is `Read, Glob, Grep, Bash, Write, Edit, ToolSearch` — nothing broader

**ID:** minimal-tool-list
**Status:** active
**Chosen:** these seven. Every dispatch prompt across `implement.md`, `review-panel.md` and
`verify-and-handoff.md` already carries the TOOLS paragraph: "every tool you need that is not
already listed in your tool set — `SendMessage`, `Monitor`, an MCP tool — is loaded in one
`select:<name>,<name>` ToolSearch in your first turn." That paragraph already assumes a narrow base
set augmented on demand; `Read`/`Glob`/`Grep`/`Bash` cover every role's search-and-verify work,
`Write`/`Edit` cover report/reproducer authoring (reviewer roles) and source changes (implementer,
panel-fix, bugbot, mutation), and `ToolSearch` is what the TOOLS paragraph itself requires be
present.
**Considered:** also listing `WebFetch`, `WebSearch`, `AskUserQuestion`, `SendMessage`, `Monitor`,
`NotebookEdit` explicitly. Ruled out: the TOOLS paragraph already documents that these (and any MCP
tool) are meant to be loaded via `ToolSearch` rather than granted up front — listing them directly
would contradict that documented mechanism and widen the base tool set for no role that needs it
outright.

### Document the exclusion accurately, naming which rows it covers and which it does not

**ID:** document-exclusion-accurately
**Status:** active
**Chosen:** add one paragraph to `implement.md`'s "Dispatch sites — the parent's closed list"
naming that the `tools:` change covers the panel-bundle and panel-fix rows under
`REVIEW_PANEL_TOGGLE: dynamic`, and that the verifier row is unaffected regardless of any toggle
(it dispatches `general-purpose` unconditionally, per `verify-and-handoff.md`). An earlier draft of
this paragraph claimed the exclusion covered "every one of these four rows," which is false for the
verifier row and conflated the implementer row's own toggle with `REVIEW_PANEL_TOGGLE` — caught by
this run's own review panel (F2) and corrected before landing.
**Considered:** leaving `implement.md` unchanged, since the ticket's scope item naming it was framed
as "may need updating." Ruled out: the closed-list table is the one place this pipeline states
which dispatch sites exist at all, and leaving it silent about the fix a sibling section just
documented is exactly the kind of prose drift `## Global constraints`-style change ledgers exist to
prevent.

### Close the auto-pick "architectural decision" loophole this run's own mistake surfaced

**ID:** close-autopick-loophole
**Status:** active
**Chosen:** add one paragraph to `skills/flow-fast/brainstorm.md` section B stating that a round
carrying a recommended default is auto-picked regardless of the decision's architectural weight —
at the operator's explicit request, riding along in this same run's commits rather than opening a
separate change, since it is a small, self-contained prose fix directly caused by this run's own
brainstorming.
**Considered:** filing it as a separate follow-up change instead. Ruled out by the operator's own
instruction ("commit this fix as a part of current task").

## Open questions

None — every decision point above was resolved by asking or by the review panel; nothing here
needed leaving open.
