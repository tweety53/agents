# flow skills

**flow** = spectre + Superpowers **Basic Workflow** bridge with a **three-state** machine: a
kickoff marker, one combined review-and-test gate, and an integrate phase that merges before the
cleanup run cleans up.

The state each phase of `/flow` ends in, and the gate that follows it, are under
**States** (`flow-contracts/pipeline.md`); the state diagram and the per-command stage table are
under **How the pipeline works** (`README.md`). This file copies none of them — it is read
on demand, beside the document it would be copying, exactly as `README.md` is. What does carry the
three-line digest is the layer that is loaded into a session before anything reads the pipeline: the
always-on rule, and a project's own `CLAUDE.md` / `AGENTS.md`.

See also: `flow-manual-review.mdc` — authored at `rules/flow-manual-review.mdc` in this repo,
installed by `setup.sh global` to `~/.claude/rules/` and inlined into the managed block in
`~/.claude/CLAUDE.md`. It is a **stub**: the pipeline itself lives in
`skills/flow-contracts/pipeline.md`, loaded on demand by `/flow`.

`<name>` is **optional** on `/flow` and on `/flow-status` — if omitted, the sole active
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
| **4** | subagent-driven-development | `/flow` (implementation, `sdd` execution only) |
| **5** | test-driven-development | `/flow`, every implementer dispatch |
| **6** | the review panel | `/flow` (implementation) |
| **8** | verification-before-completion | `/flow` (implementation) |

`finishing-a-development-branch` is **never** invoked — integration is `/flow`'s own job.

## Command map

| Command | Skill | What it does |
|---------|-------|--------------|
| `/flow <name>` | `flow` | Single-command pipeline: no state creates the change and writes `STARTED`, then — same invocation — runs brainstorming (fully interactive) and implementation behind the review panel resolved from the settings store, ending at `IN_PROGRESS`. An argument at `IN_PROGRESS` is a fix run; state unchanged. Bare at `IN_PROGRESS` asks how to land the branch — open PR (default), merge and push, or manual — and, on merge-and-push, continues in the same invocation through cleanup to `FINISHED`. Publishes no proposal artifact. |
| *(gate)* | you | Creating run or fix: review the staged diff — the stack is running. Integrate with open PR or manual: wait for the branch to merge (or finish your manual steps). Merge-and-push: nothing — the state is terminal. |
| `/flow-fast <name>` | `flow-fast` | Minimal-ceremony `/flow` variant — one invocation from Jira key to landed change: a git worktree for isolation only, implementation and review panel as the plan's class decides, project lint plus the tests the change touches, the project's default landing route, cleanup. Marks every `flow.*` stage `/flow` marks; no spectre artifacts or state file. |
| `/flow-status [name]` | `flow-status` | Read-only report of where every open change is |
| `/flow-plan` | `flow-plan` | Thinking-partner mode — no implementation; a captured session creates the change at `STARTED` for `/flow` to resume |
| `/flow-settings` | `flow-settings` | Reads/writes the harness-wide reviewer slots every `/flow` run reads from |
| `/flow-self-review <name>` | `flow-self-review` | Runs a change's self-review pass, inline on this session's model, from the context bundle `/flow` or `/flow-fast` saved. Standalone, not a pipeline stage. Fixes and lands every non-big finding; files the big ones. |

Each row says what a command is *for*. Its stages, in order, are stated once under
**Level 1 — the stages of each command** (`README.md`) and are deliberately not
repeated here.

## Skills

```
skills/
├── flow/               ← /flow (brainstorm, implement behind the review panel, integrate and clean up)
├── flow-fast/          ← /flow-fast (worktree, implement, land, clean up)
├── flow-status/         ← /flow-status (read-only)
├── flow-plan/       ← /flow-plan
├── flow-settings/       ← /flow-settings
├── flow-self-review/   ← /flow-self-review (self-review pass)
└── flow-contracts/    ← on-demand contracts; `pipeline.md` is canonical for the state machine
```

Every skill above but `flow-plan` and `flow-contracts` requires the `spectre` CLI
(`go install github.com/tweety53/spectre/cmd/spectre@latest`); those two need none — reading a
spectre tree, or a contract file, is reading markdown.

## Naming a guard in skill prose

**Prose describing this repository's own guard is not an invocation.** A guard invoked by name
uses the basename form above. Prose that describes **this repository's own** lint and test
guards — resolved through `<agents repo>/.flow/project.md`'s `## lint` and `## test` lists
rather than through `<skill-dir>/scripts/` — names the guard as
`<agents repo>/scripts/<name>` instead of a bare repository-relative path: a bare path there
resolves, for a reader standing in an installed project, against that project's own tree, so the
sentence would name a file the reader may be able to write. See **Guard resolution**
(`skills/flow-contracts/pipeline-rationale.md`) for why carrying the prefix matters.
