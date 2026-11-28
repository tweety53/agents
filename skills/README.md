# flow skills

**flow** = spectre + Superpowers **Basic Workflow** bridge with a **three-state** machine: a
kickoff marker, one combined review-and-test gate, and an integrate phase that merges before the
cleanup run cleans up.

The state each phase of `/flow` ends in, and the gate that follows it, are under
**States** (`flow-contracts/pipeline.md`); the pipeline diagram is under **How a change really flows** and the
step table under **The steps a run marks** (`README.md`). This file copies none of them — it is read
on demand, beside the document it would be copying, exactly as `README.md` is. What does carry the
three-line digest is the layer that is loaded into a session before anything reads the pipeline: the
always-on rule, and a project's own `CLAUDE.md`.

See also: `flow-manual-review.mdc` — authored at `rules/flow-manual-review.mdc` in this repo,
installed by `setup.sh global` to `~/.claude/rules/` and inlined into the managed block in
`~/.claude/CLAUDE.md`. It is a **stub**: the pipeline itself lives in
`skills/flow-contracts/pipeline.md`, loaded on demand by `/flow`.

`<name>` is **optional** on `/flow` — if omitted, the sole active
change relevant to that state is used automatically; if there are multiple, you're
asked which.

**Model:** See "Model resolution" in `skills/flow/SKILL.md`, which is canonical for `/flow`; see
"Model policy" in `flow-contracts/model-policy.md` for the per-harness enforcement notes that
still apply.

## Superpowers Basic Workflow map

| Step | Skill | flow command |
|------|-------|----------------|
| **1** | brainstorming | `/flow` (creating run) |
| **3** | writing-plans | `/flow` (creating run) |
| **2** | worktree creation, stated in `skills/flow/implement.md` (no skill) | `/flow` (implementation) |
| **4** | subagent-driven-development | `/flow` (implementation) |
| **5** | test-driven-development | `/flow`, every implementer dispatch |
| **6** | the review panel | `/flow` (implementation) |
| **8** | verification-before-completion | `/flow` (implementation) |

`finishing-a-development-branch` is **never** invoked — integration is `/flow`'s own job.

## Command map

| Command | Skill | What it does |
|---------|-------|--------------|
| `/flow <name>` | `flow` | Single-command pipeline: no state creates the change and writes `STARTED`, then — same invocation — runs brainstorming (fully interactive) and implementation behind the decided review panel, ending at `IN_PROGRESS` — or, unless a question is open, landing straight away in the same invocation. An argument at `IN_PROGRESS` is a fix run; state unchanged, and it lands the same way. Bare at `IN_PROGRESS` lands the branch by the project's `## default landing route`, asking only when none is declared — open PR (default), merge and push, or manual — and, on merge-and-push, continues in the same invocation through cleanup to `FINISHED`. Publishes no proposal artifact. |
| *(gate)* | you | Creating run or fix that leaves a `?` in its `Decisions:`: answer it, then review the staged diff (otherwise it lands straight away). Integrate with open PR or manual: wait for the branch to merge (or finish your manual steps). Merge-and-push: nothing — the state is terminal. |
| `/flow-interactive [key or name]` | `flow-interactive` | Conversational session: one invocation is one change, one Jira issue and one session worktree, and a stream of requests the main session coordinates — each looked at live, then by one subagent implemented test-first, verified with the project's lint and targeted tests, landed by merge and push, reinstalled and, when it touches `ui paths`, checked on the live page — until an explicit stop runs one self-review pass, cleanup and `FINISHED`. Every request is recorded in flowd; no spectre artifacts or state file |
| `/flow-plan` | `flow-plan` | Thinking-partner mode — no implementation; a captured session creates the change at `STARTED` for `/flow` to resume |
| `/flow-learn-quiz` | `flow-learn-quiz` | Teacher mode, only when asked: quizzes you on a change. Standalone |
| *(run by the pipeline)* | `flow-self-review` | Runs a change's self-review pass, a one-shot dispatch — inside `/flow`'s integrate and `/flow-interactive`'s stop, or standalone from the context bundle flowd saved at cleanup. Standalone, not a pipeline stage. Fixes and lands every non-big `flow-fix` finding; suggests the impactful rest (no Jira issue). |

Each row says what a command is *for*. Its steps are stated once under
**The steps a run marks** (`README.md`) and are deliberately not
repeated here.

## Skills

```
skills/
├── flow/               ← /flow (brainstorm, implement behind the review panel, integrate and clean up)
├── flow-interactive/   ← /flow-interactive (a conversational session landing each request)
├── flow-plan/          ← /flow-plan
├── flow-learn-quiz/    ← /flow-learn-quiz
├── flow-self-review/   ← the self-review pass, run by the pipeline
└── flow-contracts/     ← on-demand contracts; `pipeline.md` is canonical for the state machine
```

Every skill above but `flow-plan` and `flow-contracts` requires the `spectre` CLI
(`go install github.com/tweety53/spectre/cmd/spectre@latest`); those two need none — reading a
spectre tree, or a contract file, is reading markdown.

## Phase load table

The files each phase loads, in load order, and the condition under which each is loaded. Every
row indexes the load directive in the file named under **Loaded by**; that directive is canonical.
`TestPhaseLoadTableMatchesTheTree` (`stats/internal/guard/load_table_test.go`) fails when a file
named here does not exist or a conditional load set (`stats/internal/guard/loadsets.go`) lists a
file this table omits.

| Phase | File | Condition | Loaded by |
|---|---|---|---|
| every `/flow` run | `skills/flow/SKILL.md` | always | the harness |
| every `/flow` run | `skills/flow-contracts/pipeline.md` | always, first | `skills/flow/SKILL.md` |
| every `/flow` run | `skills/flow-contracts/progress-visibility.md` | the harness lacks the `flow-task-list` mod, or the run edits it, the board or `tasks.md`'s markers | `skills/flow-contracts/pipeline.md` |
| every `/flow` run | `skills/flow-contracts/quiet-progress.md` | the project's `## progress` mode is on | `skills/flow-contracts/pipeline.md` |
| every `/flow` run | `skills/flow-contracts/pipeline-defects.md` | the run hits a pipeline defect | `skills/flow-contracts/pipeline.md` |
| every `/flow` run | `skills/flow-contracts/guard-verdict-verification.md` | a gate guard prints `DRIFT-DIRTY`, `REFUSE`, `OUTSTANDING` or `MOVED` | `skills/flow-contracts/pipeline.md` |
| every `/flow` run | `skills/flow-contracts/state-file.md` | before reading or writing the state file | `skills/flow-contracts/pipeline.md` |
| every `/flow` run | `skills/flow-contracts/project-configuration.md` | before reading `.flow/project.md` | `skills/flow-contracts/pipeline.md` |
| every `/flow` run | `skills/flow-contracts/project-configuration-visual.md` | `flow.verify` begins, or a `## visual verification` section is written or reviewed | `skills/flow-contracts/project-configuration.md` |
| every `/flow` run | `skills/flow-contracts/jira-integration.md` | before any Jira call | `skills/flow-contracts/pipeline.md` |
| creating run (no state) | `skills/flow/brainstorm.md` | always | `skills/flow/SKILL.md` |
| creating run | `skills/flow/resume.md` | the name resolves to a change already recorded | `skills/flow/brainstorm.md` |
| creating run, planner | `skills/flow/brainstorm-planner.md` | the planner dispatch | `skills/flow/brainstorm.md` |
| creating run, planner | `skills/flow/brainstorm-ui.md` | the change carries UI | `skills/flow/brainstorm-planner.md` |
| creating run, planner | `skills/flow/designer-brief.md` | a `designer` is dispatched | `skills/flow/brainstorm-planner.md` |
| creating run, planner | `skills/flow/plan-triage.md` | the dispatch that triages a plan check | `skills/flow/brainstorm-planner.md` |
| resume at `STARTED` | `skills/flow/resume.md` | always | `skills/flow/SKILL.md` |
| resume at `STARTED` | `skills/flow/brainstorm.md` | the resume point is not `implement.md` | `skills/flow/resume.md` |
| implement | `skills/flow/implement.md` | planning artifacts exist | `skills/flow/SKILL.md` |
| implement | `skills/flow-contracts/worktree-resolution.md` | always | `skills/flow/implement.md` |
| implement | `skills/flow/cross-repo-worktrees.md` | the change links a peer or a second `## apps` repository | `skills/flow/implement.md` |
| implement | `skills/flow/implementer-brief.md`, `skills/flow/fixer-brief.md` | read by the implementer or fixer it dispatches | `skills/flow/implement.md` |
| implement | `skills/flow/fix.md` | a fix run | `skills/flow/implement.md` |
| implement | `skills/flow/visual-verify.md` | `flow-guard check-visual-trigger` matches | `skills/flow/implement.md` |
| verify and handoff | `skills/flow-contracts/session-records.md` | before reading a render outcome | `skills/flow/verify-and-handoff.md` |
| review panel | `skills/flow/review-panel.md` | after `flow.implement` | `skills/flow/SKILL.md` |
| review panel | `skills/flow/slot-brief.md` | read by every panel slot | `flow-guard build-slot-prompt` |
| review panel | `skills/flow/review-panel-optional-slots.md` | a round dispatching `mutation` | `skills/flow/review-panel.md` |
| review panel | `skills/flow/review-panel-fix-round.md` | a round recorded a Critical or Important | `skills/flow/review-panel.md` |
| verify and handoff | `skills/flow/verify-and-handoff.md` | the panel closes clean | `skills/flow/SKILL.md` |
| verify and handoff | `skills/flow-contracts/worktree-resolution.md` | always | `skills/flow/verify-and-handoff.md` |
| verify and handoff | `skills/flow-contracts/workspace-isolation.md` | by section, the one each trigger names | `skills/flow/verify-and-handoff.md` |
| verify and handoff | `skills/flow/fix.md` | a stage's final run is red | `skills/flow/verify-and-handoff.md` |
| verify and handoff | `skills/flow/live-check.md` | the live check runs | `skills/flow/verify-and-handoff.md` |
| verify and handoff | `skills/flow/visual-verify.md` | the visual closing line is not `NEXT: skip` | `skills/flow/verify-and-handoff.md` |
| verify and handoff | `skills/flow/visual-verify-verifier.md` | read by the visual verifier | `skills/flow/visual-verify.md` |
| verify and handoff | `skills/flow-contracts/git-boundaries.md` | before committing | `skills/flow/verify-and-handoff.md` |
| verify and handoff | `skills/flow-contracts/git-boundaries-commit-chain.md` | the state file records a commit chain | `skills/flow/verify-and-handoff.md` |
| fix run (`IN_PROGRESS` with an argument) | `skills/flow/fix.md` | always | `skills/flow/SKILL.md` |
| fix run | `skills/flow/fix-visual.md` | a visual fix | `skills/flow/fix.md` |
| integrate (bare at `IN_PROGRESS`) | `skills/flow/integrate.md` | always | `skills/flow/SKILL.md` |
| integrate | `skills/flow-contracts/finish-contract-run1.md` | a stop cites it | `skills/flow/integrate.md` |
| integrate | `skills/flow/integrate-own-tooling.md` | the canonical repository is `<agents repo>` | `skills/flow/integrate.md` |
| integrate | `skills/flow/unfinished-work-gate.md` | `STOP: unfinished-work` | `skills/flow/integrate.md` |
| integrate | `skills/flow/sync-onto-base.md` | `sync-conflict` or a moved base | `skills/flow/integrate.md` |
| integrate | `skills/flow/sync-onto-base-visual.md` | the guard prints `RECAPTURE:` | `skills/flow/sync-onto-base.md` |
| integrate | `skills/flow-contracts/git-boundaries-commit-chain.md` | before writing the subjects | `skills/flow/integrate.md` |
| integrate | `skills/flow-self-review/SKILL.md` | the self-review pass | `skills/flow-contracts/finish-contract-run1.md` |
| cleanup (run 2) | `skills/flow/cleanup.md` | merge-and-push, or a merged branch | `skills/flow/integrate.md` |
| cleanup | `skills/flow-contracts/finish-contract-run2.md` | always | `skills/flow/cleanup.md` |
| cleanup | `skills/flow-contracts/artifacts-registry.md` | always | `skills/flow-contracts/finish-contract-run2.md` |
| cleanup | `skills/flow-contracts/jira-integration.md` | the Done transition | `skills/flow/cleanup.md` |
| `/flow-interactive` | `skills/flow-interactive/SKILL.md` | always | the harness |
| `/flow-interactive` | `skills/flow-contracts/operator-prompts-auto-resolution.md` | a question with a recommended option | `skills/flow-interactive/SKILL.md` |
| `/flow-interactive` | `skills/flow-contracts/workspace-isolation.md` | a request worktree | `skills/flow-interactive/SKILL.md` |
| `/flow-interactive`, step 4 | `skills/flow/visual-verify.md` | the diff touches `ui paths` | `skills/flow-interactive/SKILL.md` |
| `/flow-interactive`, step 5 | `skills/flow/sync-onto-base.md` | a rebase conflict | `skills/flow-interactive/SKILL.md` |
| `/flow-interactive`, step 3a | `skills/flow/fix.md` | a fix of a red step 4 or a step 7 defect | `skills/flow-interactive/SKILL.md` |
| `/flow-interactive`, stop | `skills/flow-self-review/SKILL.md` | the explicit stop | `skills/flow-interactive/SKILL.md` |
| `/flow-self-review` | `skills/flow-self-review/SKILL.md` | always | the pipeline, or standalone |
| `/flow-self-review` | `skills/flow-self-review/themes.md` | read by the pass | `skills/flow-self-review/SKILL.md` |
| `/flow-self-review` | `skills/flow-contracts/model-policy.md` | every dispatch | `skills/flow-self-review/SKILL.md` |

## Naming a guard in skill prose

**Prose describing this repository's own guard is not an invocation.** A guard invoked by name
is written `flow-guard <name>`. Prose that describes **this repository's own** lint and test
guards — resolved through `<agents repo>/.flow/project.md`'s `## lint` and `## test` lists
rather than through `flow-guard` on `PATH` — names the guard as
`<agents repo>/scripts/flow-guard <name>` instead of a bare repository-relative path: a bare path there
resolves, for a reader standing in an installed project, against that project's own tree, so the
sentence would name a file the reader may be able to write.
