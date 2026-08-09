## Context

Full technical design: `docs/superpowers/specs/2026-08-09-kan-106-slim-myflow-skills-design.md`.
This file carries the openspec-required `## Decisions` / `## Open questions` sourced from the
brainstorming dialogue; it does not restate the full design.

`myflow-contract-economy` (KAN-95/82/87) already provides the partition mechanism — verbatim-
partition rule, the bounded rule-extraction carve-out, the per-move ledger, and the budget-guard
ratchet. This change applies that mechanism to files it hasn't reached yet (`myflow-do`,
`myflow-finish`, `myflow-start`, `pipeline.md`, several contracts) and removes three duplicated
procedures. It adds no new partition mechanism and no new capability for the split itself.

Two behavior changes to `/myflow-fast` are folded in on top, raised and approved mid-session.

## Goals / Non-Goals

**Goals:**
- Apply the existing rule-extraction carve-out to the anti-restatement boilerplate and remaining
  justification prose across the pipeline skills and contracts KAN-95 left untouched.
- Consolidate the "named options + marked recommendation + safe default" pattern into one contract.
- Remove three duplicated procedures (empty-worktree-set stop, two-commit chain, `planningEffort`
  fallback) behind citations or a script.
- `/myflow-fast`: silent setup defaults, dropped design-approval confirm.

**Non-Goals:**
- No change to `myflow-contract-economy`'s rules themselves — this change is an application of them.
- No table-driving of the oversized guard test suites (deferred to a follow-up).
- No change to the three-state machine, the review panel roster, or any command besides
  `/myflow-fast`'s two named behavior changes.

## Decisions

### myflow-fast-silent-setup-defaults

**ID:** myflow-fast-silent-setup-defaults
**Status:** active
**Chosen:** `/myflow-fast`'s creating-run setup question round (planning effort, 3 model roles,
roster) records the recommended defaults directly, without asking — recorded here because the
operator flagged the current interactive round as friction against a command whose whole point is
speed. An explicit session instruction still overrides a named field.
**Considered:** Keep the `AskUserQuestion` prompts but pre-select the recommendation (rejected —
still an interactive stop on every creating run, which is exactly the friction being removed).

### myflow-fast-drop-design-approval-gate

**ID:** myflow-fast-drop-design-approval-gate
**Status:** active
**Chosen:** `/myflow-fast` drops the explicit post-design "does this look right?" confirm and
proceeds straight into artifact creation once the design is presented. Brainstorming's clarifying
questions and the design presentation itself stay fully interactive — that is where requirements are
gathered and where the operator can redirect. Scoped to `/myflow-fast` only; `/myflow-start` is
unchanged. The human checkpoint this removes is replaced by the one that already exists downstream:
the `IN_PROGRESS` staged-diff-and-run-instructions review.
**Considered:** Keep the confirm (rejected — operator explicitly asked to drop it, given the
downstream `IN_PROGRESS` review already provides a human gate for this command's whole point of
speed); apply the same drop to `/myflow-start` (rejected — out of scope, `/myflow-start` is not the
command being optimized for speed here, and removing its gate was never asked for).

### defer-guard-test-table-driving

**ID:** defer-guard-test-table-driving
**Status:** active
**Chosen:** Item G (table-driving `test-check-plan-provenance.sh` and
`test-check-cleanup-complete.sh`) is deferred to a follow-up Jira issue rather than attempted in this
change, because it needs its own fixture-format design independent of the prose-slimming work here,
and bundling a ~6000-line test rewrite into a prose-slimming change risks scope creep.
**Considered:** Include it now, since the ticket lists it as independent and parallelizable
(rejected — "independent" describes ordering within the ticket, not that it carries no design risk
of its own).

### fold-myflow-fast-changes-into-kan-106

**ID:** fold-myflow-fast-changes-into-kan-106
**Status:** active
**Chosen:** The two `/myflow-fast` behavior changes are folded into this change rather than filed
separately, per the operator's explicit choice when asked.
**Considered:** File as a separate follow-up issue, keeping KAN-106 scoped to prose slimming only
(the alternative offered; not chosen).

## Risks / Trade-offs

- **KAN-84 gap (open, unaffected)** → `check-references.sh` cannot distinguish a stub that kept its
  pointer from one that lost it. Mitigated by the per-move ledger, not fixed.
- **Item C spans three skill files plus one not-yet-located call site** → raises the chance of
  misclassifying call-site-specific rationale vs. shared mechanic during extraction. Mitigated by
  locating all five sites explicitly during planning before any file is edited.
- **`/myflow-fast` loses a stated hard-gate** → accepted knowingly; the downstream `IN_PROGRESS`
  review remains the actual operator checkpoint for this command.

## Open questions

(none — the operator answered every question raised during brainstorming)
