## Context

The existing `stats/cmd/flow/` subcommands (`state`, `stage`, `settings`,
`record`) resolve a change against flowd's HTTP store via a project key —
none of them touch a worktree's filesystem. Ticking `tasks.md` is inherently
a filesystem edit inside a change's own worktree, which today only the bash
guards in `skills/flow/scripts/` know how to locate (`scripts/lib/spec-root.sh`
+ `scripts/lib/change-plan.sh`, including `link.md`/peer indirection for
satellite worktrees checking someone else's plan).

## Decisions

### Implementation site: Go CLI subcommand, not a bash guard

**ID:** tasks-tick-cli-subcommand
**Status:** active
**Chosen:** `flow tasks tick` as a new `stats/cmd/flow/tasks.go` subcommand —
**Considered:** a `skills/flow/scripts/tasks-tick.sh` guard sourcing the
existing bash libs directly, no porting needed. Ruled out because the ticket
frames this as "resolving the change the same way every other myflow
subcommand does" and the operator confirmed the Go CLI is the intended home
— it gets the same install/versioning and `go test` coverage as every other
`flow` subcommand, and is available to any harness the CLI ships to.

### Plan-file resolution: spec-root probe reimplemented in Go, no link.md/peer indirection

**ID:** tasks-tick-no-peer-indirection
**Status:** active
**Chosen:** Port `spec-root.sh`'s `spec_root_leaf` probe (spectre-first, then
openspec, default spectre) directly into Go — `<dir>/<spec-root>/changes/<change>/tasks.md`,
where `-C dir` (default cwd) is used directly as the worktree root.
**Considered:** also porting `change-plan.sh`'s `link.md`/peer resolution for
satellite worktrees. Ruled out (YAGNI) — that indirection exists for guards
checking a *different* worktree's plan from outside it; `/flow` only ever
ticks a task inside the worktree that owns the plan, so there is no caller
for the indirected case yet.

### Refuse rather than no-op on an already-ticked task

**ID:** tasks-tick-refuse-on-double-tick
**Status:** active
**Chosen:** exit 1 when the target task's own checkbox is already `[x]`.
**Considered:** silently no-op (matches idempotent-command conventions
elsewhere). Ruled out per the ticket's own stated requirement — a double-tick
must be visible, not silent, since a silent no-op is exactly the kind of
quiet failure mode this change exists to remove.

### Task/step boundary follows `build-green.md`'s Placement rule exactly

**ID:** tasks-tick-placement-rule
**Status:** active
**Chosen:** a task is a column-0 `- [ ] <flat-integer>. …` line; its body runs
to the next task line, the next `## `/`### ` heading, or EOF; a step is a
`  - [ ] **Step N: …**` line (two-column indent) within that body. Ticking
flips the task line and every step line in its body, and nothing else.
**Considered:** nothing else — this is the plan grammar every other guard in
this repository (`check-task-commit-fields.sh`, `check-plan-shape.sh`)
already reads, so a second grammar here would be a new source of drift.

## Open questions

None.
