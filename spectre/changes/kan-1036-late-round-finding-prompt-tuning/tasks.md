# kan-1036-late-round-finding-prompt-tuning

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** make the self-review's prompt-tuning angle trace every late-round miss.

**Spec:** `design.md` beside this file.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1036-late-round-finding-prompt-tuning`; every `**Files:**` path is relative to its root. Never touch the main checkout.

- [ ] 1. Trace a late-round miss in angle 8
  - [ ] **Step 1: Add the paragraph** to `skills/flow-self-review/SKILL.md`, directly after the `reviewer` bullet's last line (`` `skills/flow/visual-verify.md`. ``) and before the paragraph opening `The visual-verify step is analysed`, separated by blank lines:

```markdown unverified:applies-after-reviewer-bullet
A **late-round miss** — a finding a slot first raises in round 2 or later on a hunk that same slot
already reviewed, unchanged, in an earlier round — is traced to why the earlier rounds missed it,
whatever its severity: the pass compares that slot's earlier dispatches with the round that raised
it — the prompt, the diff each was given, the findings it returned. A generic instruction gap is a
`reviewer` miss case whose `why_picked` names the rounds that missed it; any other cause — a re-run
shown only the delta, or reviewer variance the bundle cannot explain — is an angle-8 finding naming
that cause, classified by this step's rules.
```

  - [ ] **Step 2: Acknowledge** each new sentence: run `scripts/check-verbatim-moves.sh`, and append each FAIL line's text after `::` to `spectre/changes/kan-1036-late-round-finding-prompt-tuning/verbatim-moves.txt`, one per line.
  - [ ] **Step 3: Verify:** `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh` exit 0.

**Files:** `skills/flow-self-review/SKILL.md`
**Allowed-collateral:** `spectre/changes/kan-1036-late-round-finding-prompt-tuning/verbatim-moves.txt`
**Tests:** none — prose; `check-verbatim-moves.sh` checks the sentences.
**Regression:** none — no test is declared.
**Build:** green
**After:** none
**Commit:** `docs(flow-self-review): trace a late-round review miss in angle 8`

**Decision:** late-round-miss-in-angle-8
**Decision:** late-round-miss-any-severity
**Decision:** late-round-miss-no-schema-field
