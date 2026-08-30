# kan-372-flow-the-router-eager-loads-11-contracts-before — design

## Context

KAN-372 reports that `skills/flow/SKILL.md` eager-loads `pipeline.md` plus 10 further contracts
before a `/flow` run even resolves which state it is in — 62-68% of a run's skill text, paid
before any project code, diff, or artifact is read (measured at `18a89b5`: ~45k tokens of eager
router load). A second, independent cost sits in dispatch: every subagent re-reads
`rules/agent-baseline.md` + `CLAUDE.md` + `<project>/.flow/project.md` at full price in its own
separate prompt cache. KAN-295 (Done) applied lazy-loading to `pipeline.md` alone; this change
generalises that treatment to the router's remaining load list and to the largest remaining
contracts (`finish-contract.md`, `project-configuration.md`, plus a shrink pass over
`jira-followups.md`, `review-panel.md`, `pipeline.md`, and `state-file.md`).

## Decisions

### lazy-router-loads

**ID:** lazy-router-loads
**Status:** active
**Chosen:** Delete `skills/flow/SKILL.md`'s eager-load block (the "Load worktree-resolution.md,
session-records.md, ..." paragraph and the model-policy.md disclaimer's placement). `pipeline.md`
stays eager — the router needs it to resolve state before anything else. For the other 10
contracts: 4 (`worktree-resolution.md`, `session-records.md`, `git-boundaries.md`,
`artifacts-registry.md`) already carry their own explicit "Load X" line at point of use in
`verify-and-handoff.md`, `integrate.md`, `archive.md` — deleting the SKILL.md duplicate is enough;
`implement.md` gets one added for `worktree-resolution.md`. The remaining 5
(`jira-integration.md`, `plan-provenance.md`, `build-green.md`, `workspace-isolation.md`,
`project-configuration.md`) are cited by name in phase files but never carry an explicit "Load"
imperative — add one at each real point of use. `model-policy.md`'s SKILL.md paragraph is
self-contained prose, not a file-read instruction; relocate it (unchanged) to sit directly above
`## Model resolution` instead of the top preamble.
**Considered:** Leaving the eager block and only trimming file sizes — rejected because the
router pays for every contract regardless of which phase runs; only relocating the load points
removes dead weight per run type, per KAN-295's own precedent.

### finish-contract-split-by-run

**ID:** finish-contract-split-by-run
**Status:** active
**Chosen:** Split `finish-contract.md` into `finish-contract-run1.md` (the "Finish contract"
preflight-signals overview, "Run 1 — the branch is not merged", "Resolving a change's worktrees")
and `finish-contract-run2.md` ("Run 2 — the branch is merged", "Worktree cleanup"). Verified: these
four sections have zero shared citers — `worktree-resolution.md`'s "Resolving a change's
worktrees" citation and `integrate.md`'s own use are both Run-1-only; `archive.md`'s "Worktree
cleanup" citation is Run-2-only. `integrate.md` loads only `finish-contract-run1.md`; `archive.md`
loads only `finish-contract-run2.md`. ~8 other files (`jira-followups.md`, `jira-integration.md`,
`state-file.md`, `artifacts-registry.md`, `project-configuration.md`, `pipeline.md`,
`flow-contracts/SKILL.md`) cite specific sections of the old file and are repointed to whichever
half now carries that section.
**Considered:** A three-way split with a shared preamble file — rejected once the citation survey
showed no section is actually needed by both runs; a third file would be pure overhead.

### project-configuration-split-resolution-vs-authoring

**ID:** project-configuration-split-resolution-vs-authoring
**Status:** active
**Chosen:** Split `project-configuration.md` by audience: a passage a **run** applies at
resolution time (the `## <key>` table, standards entry-form resolution, the containment rules, the
workspace-isolation/visual-verification/citation-check key tables) stays in
`project-configuration.md`; a passage that only helps a human **author** or edit their own
`.flow/project.md` (rationale for a key's shape, "how do I declare this section" framing not
consulted by any run) moves to `project-configuration-authoring.md`. The exact section-by-section
boundary is drawn during task-writing/implementation, since the source file interleaves both
audiences throughout rather than segregating them by heading.
**Considered:** Splitting strictly by existing `##` heading — rejected on inspection: the intro
section and "Where the agents repository is" both mix resolver-consulted rules with
author-oriented explanation inline, so a heading-level split would misfile content either way.

### scoped-dispatch-bundle-not-full-project-md

**ID:** scoped-dispatch-bundle-not-full-project-md
**Status:** active
**Chosen:** Extend `gather-dispatch-context.sh` to extract the `## lint`, `## test`, and `## run`
sections from `.flow/project.md` into the dispatch bundle it already produces, and update the
CONTEXT BUNDLE paragraph in `implement.md` and `review-panel.md` to state those commands are
already present in the bundle. This satisfies `CLAUDE.md`'s lint-fix-priority rule (which sends
every subagent to `.flow/project.md`'s `## lint` section) without a full 23.8 KB re-read per
dispatch.
**Considered:** Passing pre-resolved standards paths already avoids one re-read path (confirmed:
`review-panel.md` already resolves `[STANDARDS_PATHS]` before dispatch, so the principles reviewer
never opens `.flow/project.md` for standards) — the remaining, actually-costly path is lint/test
command discovery, which this decision targets directly rather than re-solving an already-solved
problem.

### shrink-remaining-large-contracts

**ID:** shrink-remaining-large-contracts
**Status:** active
**Chosen:** Audit `jira-followups.md`, `review-panel.md`, `pipeline.md`, and `state-file.md` for
passages that restate content canonical elsewhere in the flow-contracts/skill tree, and cut them
per the repository's cut-never-paraphrase rule (relocate or delete, never reword a normative
sentence). Scoped as a per-file audit task during implementation; verified against
`scripts/check-references.sh`.
**Considered:** A byte-budget cap per file — rejected: `pipeline.md`'s own **Artifact brevity**
section already forbids a length guard or byte budget on change artifacts, and the same reasoning
applies to a contract file — a budget rewards dropping facts rather than dropping restatement.

## Open questions

None recorded — the round-3 convergence check closed with no unresolved item.
