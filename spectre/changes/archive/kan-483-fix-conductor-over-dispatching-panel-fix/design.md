# Design — kan-483-fix-conductor-over-dispatching-panel-fix

## Context

KAN-482 landed (`c0f5a8f`) the full one-panel-fix-dispatch contract in
`skills/flow/review-panel.md`, the same constraint in the conductor's dispatch prompt in
`skills/flow/implement.md`, and `check-panel-fix-single-dispatch.sh` at the panel close. Item 1a
of KAN-483 is therefore already on `main`; this change delivers item 1b (the verification-step
subagents), item 2 (anti-pattern guidance and a pre-dispatch self-check across every
conductor-facing file), and the issue's lint-and-test-inline requirement.

The conductor reads three files — `skills/flow/implement.md` sections 1, 2 and 4,
`skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md` — and today four Agent-tool
dispatch sites are legitimate across them:

| Site | Role | Key shape | Owning section |
|---|---|---|---|
| implementer, one per group | `implementer` | `task-<n>-implementer` | `implement.md` section 4 |
| panel bundle, at most two per round | `reviewer` | `panel-<round>-<slot+slot>` | `review-panel.md`, **Bundled dispatch** |
| panel-fix, exactly one per round | `panel-fix` | `panel-fix-<round>` (`-retry` once) | `review-panel.md`, fix step |
| verifier, one per worktree | `verifier` | `visual-verify` (`-2`, `-retry`) | `verify-and-handoff.md`, **Visual verification** |

`flow.verify`'s verifier is the fifth today and is removed by this change (below).

`spectre/specs/` is empty in this repository — the skill markdown is the contract surface — so
the change edits skill files directly and adds no spec task.

## Sections

### 1. A closed dispatch-site table and the self-check

`skills/flow/implement.md`, **Dispatch the conductor**, gains a subsection **Dispatch sites — the
conductor's closed list** carrying the table above (without the `flow.verify` row) and, in the
same directness the one-fix-subagent rule already has, the prohibition: everything else in the
three conductor files is the conductor's own Bash and Read work — every `check-*.sh`,
`run-reproducer.sh`, `gather-dispatch-context.sh`, `prepare-workspace.sh`, `## lint` and `## test`,
every `flow record` and `flow stage` call, worktree add and remove, every report read and every
diff walk — and it is never delegated: not to a "verify" reader, a "re-verify" or "mutation
re-verify" agent, a helper, a background task, or a subagent of any other name. The self-check:
before any Agent-tool call, the conductor names the table row the call is; a call with no row is
not made. The conductor's dispatch prompt cites the subsection; the fix-dispatch sentence KAN-482
put into the relay contract folds into that citation rather than standing as a second copy.
**Inline — the parent implements** takes the same table minus the implementer and panel-fix rows.

### 2. Per-step "never a subagent" sentences

Where the contract already says "the parent does X", each conductor-scoped script-run-and-read
step gains the sentence that the conductor runs it itself and dispatches no subagent for it:

- `skills/flow/review-panel.md` — the pre-fix `run-reproducer.sh` runs; the post-fix reproducer
  re-runs; the fix-diff walk and the check of the reported `fix-mutation:` lines against the fix
  diff; `flow record status … fixed`; `check-panel-reproducers.sh`, `check-panel-findings-closed.sh`
  and `check-panel-fix-single-dispatch.sh`; throwaway-worktree add and remove; recording findings
  from slot reports.
- `skills/flow/implement.md` section 4 — the guard-and-tick boundary
  (`check-task-commit-fields.sh`, `flow tasks tick`) and the shared-wave FULL SUITE run.
- `skills/flow/verify-and-handoff.md` — `prepare-workspace.sh`, the ledger render, and visual
  verification's steps 1, 2 and 11.

Additions only. `scripts/check-normative-inventory.sh`'s output before the first edit and after
the last may differ only by added lines — no existing normative sentence is cut or reworded.

### 3. `flow.verify` runs inline

The conductor runs `## lint`, `## test` and `check-spec-reach.sh <worktree>` itself, per worktree,
in the order `project-get.sh` prints, not stopping at the first failure, with
`prepare-workspace.sh`'s `KEY=value` lines exported before every command. A non-zero exit earns
one inline re-run of that command — the environmental-flake case — and a second non-zero exit
ends the turn with `## Question` naming the command and its output verbatim. The `## Report`
shape is unchanged. The "edits no source and runs none of the `## lint` or `## test` commands
itself" paragraph is rewritten: no source edit after the panel still holds; the commands are now
the conductor's own. The **Guardrails** bullet saying the verifier runs them is updated the same
way. **The verifier dispatch** is rescoped to `flow.visual-verify` only — `-key visual-verify`,
`visual-verify-2`, `visual-verify-<key>-retry` — with its TOOLS and MODEL HANDSHAKE paragraphs
kept in place, since `check-dispatch-paragraphs.sh` pins that site. The inline run is recorded
per worktree as `-role verifier -key verify -model <conductor model> -agent-id inline`, the
convention **Inline — the parent implements** already uses for implementer and panel-fix rows.
`skills/flow/SKILL.md`'s `VERIFY_MODEL` paragraph governs one dispatch, `flow.visual-verify`'s.

### 4. Budgets

`skills/flow/implement.md` sits at 40035 of 41777 budgeted bytes; the new subsection exceeds the
headroom, so its row in `scripts/check-contract-budget.sh` is raised in the same task that adds the
subsection. `review-panel.md` (61655 of 73554), `verify-and-handoff.md` (34354 of 42943) and
`SKILL.md` (16419 of 20520) are expected to stay under budget; the implementer measures after
editing and raises a row only where the guard fails.
<!-- measured: wc -c skills/flow/implement.md skills/flow/review-panel.md skills/flow/verify-and-handoff.md skills/flow/SKILL.md and the budgets() rows of scripts/check-contract-budget.sh @ branch spectre/kan-483-fix-conductor-over-dispatching-panel-fix, 2026-09-10 -->

## Toggles

Resolved `dynamic` for all three (`## execution mode`, `## implementer model`, `## review panel`);
the decision is `tasks.md`'s `## Decision` block and `.superpowers/sdd/decision.json`.

## Decisions

### Enforcement is prose plus a prompt-carried self-check, not a store-side guard

**ID:** prose-plus-self-check
**Status:** active
**Chosen:** a closed dispatch-site table in `implement.md`, cited from the conductor's dispatch
prompt, with a pre-dispatch self-check — the KAN-449 ledger records one `panel-fix-0` row while
four fix subagents and two verification subagents ran, so an unwarranted dispatch is invisible to
`flow record dispatches` and no guard over the store can count it; only the text the conductor
reads before launching can stop it.
**Considered:** a store-side guard like `check-panel-fix-single-dispatch.sh` — counts only recorded
rows, and the drift is exactly the unrecorded ones; a grep guard over the skill text asserting the
table and each "never a subagent" sentence is present — declined by the operator, since it catches
accidental deletion of the wording rather than the drift itself.

### `flow.verify` runs inline in the conductor and its verifier dispatch is dropped

**ID:** verify-inline
**Status:** active
**Chosen:** the conductor runs `## lint`, `## test` and `check-spec-reach.sh` itself — the issue
requires lint and test inline, always, and with those carved out the `flow.verify` verifier would
run one script, the shape section 1 forbids; `flow.visual-verify`'s verifier is untouched, so the
legitimate verifier dispatch the issue names to preserve survives there.
**Considered:** keeping a `flow.verify` verifier that runs only `check-spec-reach.sh` — preserves
the site literally but is a one-script dispatch the new table would otherwise forbid; reading the
issue's line as a run instruction with no contract change — rejected by the operator.

### A failing inline lint or test command is re-run once, then handed back

**ID:** inline-rerun-once
**Status:** active
**Chosen:** one inline re-run of the failing command, then `## Question` with its output verbatim
— the same two-attempt shape the verifier's re-dispatch had, kept for the environmental-flake case
it existed for, with no delegation.
**Considered:** no re-run, first non-zero exit is the handback — loses the flake distinction the
second attempt was added for.

### KAN-482's prompt sentence folds into a citation of the new subsection

**ID:** cite-not-restate
**Status:** active
**Chosen:** the conductor's dispatch prompt cites **Dispatch sites — the conductor's closed list**,
whose panel-fix row carries the one-per-round rule — one statement of the dispatch contract, per
this repository's non-repetition rule.
**Considered:** keeping the KAN-482 sentence beside the citation — two statements of the same
rule, which is the drift the non-repetition rule exists to prevent.

## Open questions

None.
