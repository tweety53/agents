# pipeline-fix-kan-1037-stats-weekly-costs-dashboard-34

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` counts tests added by the task; this plan adds none.

**Goal:** a `/flow` run never stops past its plan gate for a question carrying a recommended
option, or for its size; the run's shape is stated once.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit with the `**Commit:**` subject, no trailer; never write
  `spectre/changes/`.
- Insert the paragraphs below verbatim; cut, never paraphrase.

- [x] 1. State the non-stop run shape; make the mockup ask yield to auto-resolution

**Files:** `skills/flow-contracts/pipeline.md`, `rules/design-mockups-are-specs.mdc`, `stats/internal/guard/installedcitations.go`, `stats/internal/guard/references.go`
**Tests:** none — prose; the lint guards in the verify step are the check
**Regression:** reverting this commit drops the run's shape statement and lets "Ask rather than invent" read as licence to ask a recommended-option question mid-run (kan-1037).
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**After:** none
**Commit:** `docs(flow-contracts): state the non-stop run shape past the plan gate`
**Build:** green

**Decision:** shape-in-fewest-operator-actions
**Decision:** plan-gate-clear-is-the-boundary
**Decision:** mockup-ask-yields-in-flow

  - [x] **Step 1: `skills/flow-contracts/pipeline.md`.** Directly under `## Fewest operator actions`,
    before the paragraph opening "**The operator's actions and the run's stops are minimised", insert
    this paragraph and a blank line:

```markdown unverified:prose to insert verbatim; no runtime to apply it against
**Planning is a change's one interactive phase; past the plan gate the run does not stop.**
Brainstorming and the plan gate are the operator's (`skills/flow/brainstorm-planner.md`), and the
plan gate's `/clear` handoff is the boundary. From the `/flow <name>` that resumes the planned
change, implementation, review, verify, integrate and cleanup follow in one invocation until the
landing route ends the run (`skills/flow-contracts/finish-contract-run1.md`). The one stop the
operator chooses is review-before-landing (`reviewBeforeLanding`, `state-file.md`), for manual
testing after implementation; beyond it only **What still stops** ends a turn early. No rule in
another file licenses asking a question that carries a recommended option (kan-1037: the parent
asked a visual verifier's three spec questions under `rules/design-mockups-are-specs.mdc`'s "Ask
rather than invent", then stopped after an in-run fix task citing token cost).
```

  - [x] **Step 2: `rules/design-mockups-are-specs.mdc`.** In the `<!-- core -->` paragraph opening
    "**Ask rather than invent.**", after the sentence "Batch the questions up front; ask mid-build the
    moment a real blocker appears." insert:

```markdown unverified:prose to insert verbatim; no runtime to apply it against
Inside a `/flow` run past its plan gate, a question carrying a recommended option is never asked:
the run takes that option (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`).
```

    Rewrap the paragraph to the file's width; change no other word.
  - [x] **Step 3: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`
    and `scripts/check-installed-citations.sh` exit 0.
  - [x] **Step 4: Commit** the two files with the subject above.
