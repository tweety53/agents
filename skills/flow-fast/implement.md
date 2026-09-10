# Implement — inline, TDD, targeted runs only

Loaded by `skills/flow-fast/SKILL.md` once `skills/flow-fast/brainstorm.md` hands off (a creating
run) or on a fix run at `IN_PROGRESS`. Everything below runs in the parent session — **there is no
conductor and no implementer subagent dispatch.**

## 1. Isolate the workspace

```bash
flow stage begin -command '/flow-fast' -stage flow.isolate-workspace -harness <harness> -session-token ff-<literal-token> <name>
```

Run

```bash
prepare-workspace.sh <worktree>
```

once per resolved worktree (the resolved set from `<project>/spectre/changes/<name>`'s own state, non-empty
by construction). `prepare-workspace.sh` runs `check-workspace-isolation.sh` against the worktree
first, then derives and exports the project's `## workspace isolation` variables — cited from
`skills/flow/verify-and-handoff.md`'s own paragraph on this, never restated. Exit 0 with nothing
printed means the project declares no such section. A non-zero exit (a dropped-row refusal, or a
cannot-answer refusal) stops the run before any task starts, relaying the script's own lines.

```bash
flow stage end   -command '/flow-fast' -stage flow.isolate-workspace -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.load-context -harness <harness> -session-token ff-<literal-token> <name>
```

Read `proposal.md`, `design.md` and `tasks.md` from the change root before touching any task.

```bash
flow stage end -command '/flow-fast' -stage flow.load-context -outcome completed <name>
```

## 2. Fix documentation (fix runs only)

On a fix run at `IN_PROGRESS` only — skip this section entirely on a creating run:

```bash
flow stage begin -command '/flow-fast' -stage flow.document-fix -harness <harness> -session-token ff-<literal-token> <name>
```

Document the fix in `proposal.md`/`tasks.md` (or a `<name>-fix-N` sub-change) before implementing
it, exactly as `/flow`'s own `flow.document-fix` (`skills/flow/implement.md`) does — cited, never
restated. Re-record the fixed decision (`skills/flow-fast/brainstorm.md`'s section D) unchanged;
there is nothing to re-roll.

```bash
flow stage end -command '/flow-fast' -stage flow.document-fix -outcome completed <name>
```

## 3. Task loop — TDD, one commit per task

```bash
flow stage begin -command '/flow-fast' -stage flow.sdd-tdd -harness <harness> -session-token ff-<literal-token> <name>
```

For each task in `tasks.md`, in plan order (respecting any `**After:**`
field):

1. **Write the failing test first**, from the task's own `**Tests:**` field. A task whose
   `**Tests:**` opens `**none**` skips this step.
2. **Make it pass** — implement exactly what the task's own description and steps name; no
   speculative scope beyond it.
3. **Run that task's own verify step** — the lint commands its `**Files:**` actually need, plus the
   build tool's own selector for each `**Tests:**` entry (`-run '<name>'` for `go test`, `-t
   '<name>'` for vitest, and the like) — **never** the project's whole `## lint` or `## test` list,
   and never a bare module or repository suite. A full run happens **only** when the operator's own
   instruction text asks for one, at any point in the session — never automatically, and never as
   a default before handoff.
4. **Commit** — one commit per task, subject exactly the task's `**Commit:**` field, scope naming
   the module the task's own `**Files:**` carries. The commit's diff matches the task's `**Files:**`
   (plus any `**Allowed-collateral:**` glob) and its message carries no more than that task's own
   scope. Mark the task's own checkbox in `tasks.md` once its commit lands — this repository's own
   convention, since `check-task-commit-fields.sh` is not run for `/flow-fast`
   (`skills/flow-fast/SKILL.md`'s **Guard set**) but the field shape is kept so a later `/flow`
   resume or a manual audit still reads it correctly.

A task tagged `**Build:** red` is committed on its own, then folded into its `**Squash-with:**`
target's commit per that task's own instruction — the same red/green mechanics
`skills/flow-contracts/build-green.md` defines, cited rather than restated.

**No SDD ceremony runs here.** No spec-delta guard, no plan-provenance guard invocation mid-task.
A task whose `**Files:**` names a `<project>/spectre/specs/<capability>.md` path still writes and commits
that edit in its own commit — the plan already scoped it there in
`skills/flow-fast/brainstorm.md`'s section C — only the guard that would otherwise verify spec reach is
not run.

```bash
flow stage end -command '/flow-fast' -stage flow.sdd-tdd -outcome completed <name>
```

Once every task's checkbox is checked, continue directly into `skills/flow-fast/review.md` — there
is no gate here.
