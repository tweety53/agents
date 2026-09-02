---
name: flow-settings
description: Read and change the harness-wide flow defaults — default model, reviewer slots, self-review model and planning model. Standalone, not a pipeline stage. Use for /flow-settings.
allowed-tools: Bash(flow:*), Bash(jq:*)
license: MIT
compatibility: Requires the flow CLI and jq.
metadata:
  author: gymie
  version: "1.0"
---

Read and change the harness-wide settings record `flow settings get`/`set` manage: the default
model (`defaultModel`) every `/flow` run's implementer, fixer and reviewer roles use unless a
session overrides it, the reviewer slots (`reviewers`) the review panel dispatches by default, the
model (`selfReviewModel`) the archive-phase self-review pass runs on unless a project or session
overrides it, and the model (`planningModel`) the planner dispatch, `flow.document-fix`'s planner
dispatch and `/flow-research`'s research subagent run on unless a project or session overrides it.

**This is a standalone command, not a pipeline stage.** It takes no change name, reads and writes
no per-change state file, and marks no `flow stage` call. It changes the harness-wide store, not
any one change's record.

**No flags.** Per **Command surface** (`skills/flow-contracts/pipeline.md`), no `/flow*` or
`/flow*` command accepts a flag; that rule extends to this command as part of the same family. The
only input is the operator's answers to the questions this skill asks interactively.

**Announce at start:** "Using flow-settings."

## Workflow

### 1. Read current settings

```bash
CURRENT="$(flow settings get)"
```

`flow settings get` prints one line of JSON: `defaultModel` (a string), `reviewers` (an array of
strings), `selfReviewModel` (a string, empty meaning "the store's own default, `fable`") and
`planningModel` (a string, empty meaning "the store's own default, `fable`"). A non-zero exit means the store could
not be reached — report the CLI's stderr verbatim and stop; there is no per-harness fallback file
for this record the way a per-change state file has one.

Print the current values plainly before asking anything:

```
Current flow settings:
  default model:      sonnet
  reviewers:          <comma-separated list from Reviewers, verbatim>
  self-review model:  <selfReviewModel, or "(fable — store default)" when empty>
  planning model:     <planningModel, or "(fable — store default)" when empty>
```

### 2. Offer to change each field

Ask about the four fields the settings store actually holds — `defaultModel`, `reviewers`,
`selfReviewModel` and `planningModel`. Use **AskUserQuestion**, one field at a time, starting from
the current value read in step 1:

- **Default model** — offer the harness's known model identifiers, read from
  `<agents repo>/stats/internal/store/settings.go`'s `ValidModels` map at the time this skill runs,
  plus "keep current". This is the one value every `/flow` run's implementer, fixer and reviewer
  roles default to; a session can still override it via plain-language instruction, recorded per-run
  rather than here.
- **Reviewers** — offer the exact reviewer-slot ids read from `<agents repo>/stats/internal/store/settings.go`'s
  `ValidReviewers` map at the time this skill runs, never a copy of that list written into this
  file — `ValidReviewers` is that store's own enum and the only place it is canonical. **This list is
  the review panel**: every id it holds is dispatched by every `/flow` run until this command changes
  it again, none held back as a fixed floor — resolution is canonical in `skills/flow/SKILL.md`'s
  Model resolution, dispatch in `skills/flow/review-panel.md`'s roster. "On-demand" now names a
  different mechanism: a per-run operator instruction can still add a slot for a single run without
  touching this list, but an id's presence here is what makes every run dispatch it. Offer the full
  set as a multi-select seeded with the current list, plus "keep current", and say plainly when
  asking that the selection made here becomes every subsequent run's panel, not just this one. The
  CLI itself refuses an empty `-reviewers` before the store is ever contacted: its required-flags
  check (`<agents repo>/stats/cmd/flow/settings.go:123`) exits 2, distinct from the store's own
  exit-1 rejection covered in step 3 below — warn the operator before they try it that selecting zero
  slots fails at step 3 with exit 2, rather than turning review off.
- **Self-review model** — offer the same `ValidModels` set as Default model, plus an explicit
  **"Store default (fable)"** option (maps to the empty string, `-self-review-model ""`) and "keep
  current". This is the model the archive-phase self-review pass — the 5-angle retrospective
  `/flow` runs after a change reaches `FINISHED` — dispatches on; an empty value is a legitimate,
  first-class choice, not a fallback born of an unreachable store, so offer it as a named option
  rather than only as "keep current". Empty resolves to the literal `fable`, unless
  `<project>/.flow/project.md`'s `## self review model` key or a per-run session instruction
  overrides it, per **Model resolution** (`skills/flow/SKILL.md`).
- **Planning model** — offer the same `ValidModels` set as Default model, plus an explicit
  **"Store default (fable)"** option (maps to the empty string, `-planning-model ""`) and "keep
  current". This is the model the planner dispatch, `flow.document-fix`'s planner dispatch and
  `/flow-research`'s research subagent run on unless `<project>/.flow/project.md`'s `## planning
  model` key or a per-run session instruction overrides it; an empty value resolves to the literal
  `fable`, per **Model resolution** (`skills/flow/SKILL.md`).

If the operator keeps all four fields unchanged, say so and stop — do not call `settings set` for
a no-op write.

### 3. Write the change

```bash
flow settings set -model "<defaultModel>" -reviewers "<comma,separated,list>" -self-review-model "<selfReviewModel>" -planning-model "<planningModel>"
```

`-model` and `-reviewers` are required by the CLI even when only one field changed — `settings set`
writes the whole record, replacing what was recorded before. `-self-review-model` and
`-planning-model` are both optional on the CLI (omitting either means empty), but this skill always
passes both explicitly — "keep current" resolves to the value read in step 1, exactly as the other
two fields do — so a no-op run of this skill never silently resets either to empty. Pass the value
just confirmed for each field that changed, and the value read in step 1 for each field that did
not.

A non-zero exit means the store rejected the write (an invalid model or reviewer name) or could not
be reached. Print the CLI's stderr verbatim — it names the specific bad value on a rejection — and
do not report success.

On success, report exactly what changed against the values read in step 1: which field(s) changed,
old value → new value. A field left unchanged is not mentioned as a change.

## Guardrails

- **Never** touch a per-change state file, `<project>/spectre/changes/`, or any `flow
  state`/`flow stage` call — this command's write is scoped to the harness-wide settings record
  alone.
- **Never** commit, stage, push, merge, or create a worktree or branch.
- **No flags** — the only input is the operator's interactive answers.
- **Never** call `settings set` with a value the operator did not just confirm, and never call it at
  all when nothing changed.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| View and change the harness-wide flow defaults | `/flow-settings` |
