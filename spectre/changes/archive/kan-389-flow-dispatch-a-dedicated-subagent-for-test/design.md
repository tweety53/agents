# kan-389-flow-dispatch-a-dedicated-subagent-for-test — design

## Context

KAN-389: `flow.verify` and `flow.visual-verify` (`skills/flow/verify-and-handoff.md`) run the
project's `## lint`/`## test` commands and the Playwright `setup`/`verify`/`capture` steps inline
in the conductor, which `skills/flow/implement.md` dispatches on `DEFAULT_MODEL`. Those runs are
mechanical — run a command, read output, report — and inherit whatever model the operator picked
for implementation. This change moves them into a dedicated `verifier` subagent pinned to the
literal `sonnet`.

The operator chose "extract into a new subagent" over "policy only" and "pin an existing
dispatch". No staged research note existed (`docs/superpowers/research/kan-389.md`, `kan-389-*.md`
absent). No capability spec under `spectre/specs/` covers these stages, so none is edited.

Measured against `main` at `d4df668`:

- `stats/cmd/flow/record.go:44` — `recordRoles` is a closed list: `implementer, reviewer,
  panel-fix, red-partner, planner, conductor`; `record_test.go:441` pins it. The store column
  `dispatches.role` is plain `TEXT` (`stats/internal/store/migrations/0010_run_records.sql:62`),
  so a new role needs no migration.
- `stats/internal/store/settings.go:19` — `ValidModels` is `sonnet, opus, haiku, fable`.
- Review-panel reproducers run inline in the conductor via `run-reproducer.sh`
  (`skills/flow/review-panel.md`); the archive self-review subagent (`skills/flow/archive.md`
  step 9) runs no tests. Neither is in scope.

## 1. The verifier dispatch shape

Both stages dispatch the same role, `verifier`: `subagent_type: general-purpose`, Agent-tool
`model: sonnet` — the literal, never `DEFAULT_MODEL`, never a plain-language session override.
One dispatch per worktree per stage, each a fresh agent; nothing is resumed between the two stages.

Every verifier prompt carries, verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

and the relay contract: the verifier has no operator channel, fixes nothing, edits no source, runs
every command in the foreground, and ends its turn with a single `## Report` block. The first line
of its first reply is `Model: <the model named in its own system prompt>`.

**Recording.** The conductor records each dispatch as a pair, in the shape section 4 of
`skills/flow/implement.md` states for an implementer:

```bash verified:shape of section 4 of skills/flow/implement.md @ d4df668
flow record dispatch begin -change <name> -role verifier -model sonnet \
  -key <key> -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
flow record dispatch end -change <name> -key <key> -session-token mf-<literal-token> \
  -outcome completed -ended-at <ts> -agent-id <id>
```

`-task` is omitted. `<key>` is `verify` for `flow.verify` and `visual-verify` for
`flow.visual-verify`; on a change whose resolved worktree set holds more than one worktree, each
key is suffixed `-<worktree basename>` so keys stay unique within the session token.

**Handshake.** The conductor compares the `Model:` line against `sonnet`. A match proceeds. A
mismatch records the model that answered and continues on the running agent — no re-dispatch, the
conductor's own rule (`skills/flow/implement.md`, **Dispatch the conductor**):

```bash verified:shape of the conductor handshake in skills/flow/implement.md @ d4df668
flow record dispatch end -change <name> -key <key> -session-token mf-<literal-token> \
  -outcome fallback -ended-at <ts>
flow record dispatch begin -change <name> -role verifier -model <the model the handshake named> \
  -key <key>-<that model, lowercased> -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
```

A mark or a record never blocks. A verifier that ends without a `## Report`, or whose agent dies,
is closed `-outcome aborted` and the stage blocks the handoff exactly as a failed command would,
naming the death.

## 2. `flow.verify`

Conductor, unchanged: the stage marks, `prepare-workspace.sh` per worktree with its exit-1/exit-2
stops, the cache-index claim, the manual fallback when the script cannot be located, the ledger
render, and the block decision.

New: after `prepare-workspace.sh` exits 0, the conductor resolves the commands
`project-get.sh <worktree> lint` and `project-get.sh <worktree> test` print (auto-detect on
exit 1) and dispatches one verifier per worktree whose prompt states: the absolute worktree path;
the printed `KEY=value` lines to export before every command; the lint commands, then the test
commands, in the order printed; and the report shape. The verifier runs each command in order and
does not stop at the first failure — every command's outcome is reported.

The `## Report` block, one entry per command:

```text verified:authored in this change for the verifier report
## Report
- `<command>` — exit <n>
  <the command's output, verbatim, or its last 40 lines when longer, stated as truncated>
```

The conductor shows the report as this stage's output and blocks the handoff on any non-zero
exit, as today.

## 3. `flow.visual-verify`

Conductor, unchanged: the stage marks; step 1 (resolve the section, `Visual: not configured`);
step 2 (`check-visual-trigger.sh`, `Visual: no UI paths touched`, the exit-2 skip); step 9 (commit
the spec and PNGs, the `regression checkout` push command); and the **Blocking** rule.

New: for each worktree that survives steps 1–2, the conductor dispatches one verifier whose prompt
states: the absolute worktree path; the `KEY=value` lines `flow.verify` exported for that
worktree; the `## visual verification` section's resolved `setup`, `verify`, `capture` commands and
`screenshots` root; the worktree-resolved URL of each app `ui paths` matched, per **What the id
derives** (`skills/flow-contracts/workspace-isolation.md`); the project's `## run` commands; the
views this change touched (from the diff step 2 matched); `<changeRoot>`; and the instruction to run
steps 3–8 and 10 of **Visual verification** (`skills/flow/verify-and-handoff.md`) as written —
`setup`, probe, start the stack only if nothing answers, `verify`, author the spec and `capture`,
`resolve-visual-screenshots.sh` and read every printed PNG, write `<changeRoot>/visual-verification.md`,
stop the stack only if it started it. The verifier commits nothing and pushes nothing.

The `## Report` block:

```text verified:authored in this change for the verifier report
## Report
- setup: <not declared | exit <n>>
- stack: <already running | started and stopped | could not be started — <output>>
- verify: exit <n>
  <output, verbatim or last 40 lines>
- capture: exit <n>
  <output, verbatim or last 40 lines>
- spec: <absolute spec path>
- <view>: <absolute PNG path> — <what was seen, including any defect>
- visual-verification.md: written | not written — <reason>
```

Each step's non-zero exit, an unreadable PNG, `resolve-visual-screenshots.sh` exit 1 or 2, and a
defect the verifier names in a screenshot are all carried in the report. The conductor applies
**Blocking** to the report — the decision stays the conductor's, including the rule that a first-run
snapshot write by `capture` is its success path — then runs step 9 on the committed-nothing tree
the verifier left, and the `Visual:` handoff line is built from the report's view entries.

## 4. `skills/flow/SKILL.md` — Model resolution

The resolution block gains one line beside the resolved roles:

```bash verified:literal, fixed by KAN-389
VERIFY_MODEL=sonnet
```

with a paragraph in the shape of the `SELF_REVIEW_MODEL` and `PLANNING_MODEL` paragraphs, stating
the difference: `VERIFY_MODEL` governs the two verifier dispatches (`skills/flow/verify-and-handoff.md`);
it is a fixed literal — not read from the settings store, not read from `<project>/.flow/project.md`,
not overridable by a plain-language session instruction, and it never falls back because it is
never resolved. The `DEFAULT_MODEL` paragraph's "all three roles" sentence is left as it stands: the
verifier is not among them.

## 5. `stats/cmd/flow/record.go` and `implement.md`'s role sentence

- `recordRoles` gains `"verifier"`; its doc comment names the role as the subagent
  `skills/flow/verify-and-handoff.md` dispatches for `flow.verify` and `flow.visual-verify`.
- `record_test.go`: the accepted-role loop in the unrecognised-role test gains `verifier`, and a
  `TestRecordAcceptsVerifierRole` in `TestRecordAcceptsPlannerRole`'s shape pins acceptance.
- `skills/flow/implement.md` section 4: "`-role` is one of `implementer`, `reviewer`, `panel-fix`
  or `red-partner`" gains `verifier`, pointing at `skills/flow/verify-and-handoff.md`.
- The conductor's dispatch prompt (`skills/flow/implement.md`, **Dispatch the conductor**) states
  nothing new: `sonnet` is a literal the conductor reads from `skills/flow/SKILL.md`, not a value
  the parent passes.

## Decisions

### fresh-dispatch-per-stage

**ID:** fresh-dispatch-per-stage
**Status:** active
**Chosen:** one fresh verifier per worktree per stage, two record pairs — no SendMessage resume
between `flow.verify` and `flow.visual-verify`; the second dispatch is handed the same `KEY=value`
lines the first was.
**Considered:** one verifier per worktree resumed across both stages, a single record pair — rejected:
it adds a resume contract for no saved input, since the exported lines are already in the
conductor's hands.

### conductor-keeps-guards-and-commits

**ID:** conductor-keeps-guards-and-commits
**Status:** active
**Chosen:** the conductor keeps `prepare-workspace.sh`, visual steps 1–2 and 9, every stage mark
and every block decision; the verifier runs commands and visual steps 3–8 and 10.
**Considered:** the verifier also commits (step 9) — rejected: commits are a git boundary the
conductor already owns for every other stage. The verifier writing nothing (steps 3–7 and 10 only,
conductor writes `visual-verification.md` from the report) — rejected: the verifier has the paths
and observations in hand, and re-transcribing them through the report adds a copy to get wrong.

### verifier-role-recorded

**ID:** verifier-role-recorded
**Status:** active
**Chosen:** add `verifier` to `recordRoles` so the dispatch is recorded and appears in the ledger
and `cost-status`.
**Considered:** recording nothing, as archive self-review does today — rejected: a dispatch made
for cost predictability that the cost ledger cannot see defeats its own purpose.

### sonnet-mismatch-continues

**ID:** sonnet-mismatch-continues
**Status:** active
**Chosen:** a `Model:` handshake mismatch records the fallback and continues on the running agent.
**Considered:** re-dispatch once, as the planner does — rejected: the planner's re-dispatch exists
because `fable` may be unavailable; `sonnet` is the store's own default and the cheapest model in
`ValidModels`, so a re-dispatch could only cost more.

## Open questions

None.
