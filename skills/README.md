# flow skills

**flow** = spectre + Superpowers **Basic Workflow** bridge with a **three-state** machine: a
kickoff marker, one combined review-and-test gate, and an integrate phase that merges before it
archives.

The state each phase of `/flow` ends in, and the gate that follows it, are under
**States** (`flow-contracts/pipeline.md`); the state diagram and the per-command stage table are
under **How the pipeline works** (`README.md`). This file copies none of them — it is read
on demand, beside the document it would be copying, exactly as `README.md` is. What does carry the
three-line digest is the layer that is loaded into a session before anything reads the pipeline: the
always-on rule, and a project's own `CLAUDE.md` / `AGENTS.md`.

See also: `flow-manual-review.mdc` — authored at `rules/flow-manual-review.mdc` in this repo,
installed by `setup.sh global` to `~/.cursor/rules/` and inlined into the managed block in
`~/.claude/CLAUDE.md`. It is a **stub**: the pipeline itself lives in
`skills/flow-contracts/pipeline.md`, loaded on demand by `/flow`.

`<name>` is **optional** on `/flow` and on `/flow-status` — if omitted, the sole active
(non-archived) change relevant to that state is used automatically; if there are multiple, you're
asked which.

**Model:** See "Model resolution" in `skills/flow/SKILL.md`, which is canonical for `/flow`; see
"Model policy" in `flow-contracts/model-policy.md` for the per-harness enforcement notes that
still apply.

## Superpowers Basic Workflow map

| Step | Skill | flow command |
|------|-------|----------------|
| **1** | brainstorming | `/flow` (creating run) |
| **3** | writing-plans | `/flow` (creating run) |
| **2** | using-git-worktrees | `/flow` (implementation) |
| **4** | subagent-driven-development | `/flow` (implementation) |
| **5** | test-driven-development | `/flow`, every implementer dispatch |
| **6** | requesting-code-review + the panel | `/flow` (implementation) |
| **8** | verification-before-completion | `/flow` (implementation) |

`finishing-a-development-branch` is **never** invoked — integration is `/flow`'s own job.

## Command map

| Command | Skill | What it does |
|---------|-------|--------------|
| `/flow <name>` | `flow` | Single-command pipeline: no state creates the change and writes `STARTED`, then — same invocation — runs brainstorming (fully interactive) and implementation behind the fixed 3-slot review panel, ending at `IN_PROGRESS`. An argument at `IN_PROGRESS` is a fix run; state unchanged. Bare at `IN_PROGRESS` asks how to land the branch — open PR (default), merge and push, or manual — and, on merge-and-push, continues in the same invocation through archive to `FINISHED`. Publishes no proposal artifact. |
| *(gate)* | you | Creating run or fix: review the staged diff **and** run the apps. Integrate with open PR or manual: wait for the branch to merge (or finish your manual steps). Merge-and-push: nothing — the state is terminal. |
| `/flow-status [name]` | `flow-status` | Read-only report of where every open change is |
| `/flow-research` | `flow-research` | Thinking-partner mode — no implementation, no state; stages research notes for `/flow`'s brainstorming to seed from |
| `/flow-settings` | `flow-settings` | Reads/writes the harness-wide default model and reviewer slots every `/flow` run reads from |

Each row says what a command is *for*. Its stages, in order, are stated once under
**Level 1 — the stages of each command** (`README.md`) and are deliberately not
repeated here.

## Skills

```
skills/
├── flow/               ← /flow (brainstorm, implement behind the review panel, integrate and archive)
├── flow-status/         ← /flow-status (read-only)
├── flow-research/       ← /flow-research
├── flow-settings/       ← /flow-settings
└── flow-contracts/    ← on-demand contracts; `pipeline.md` is canonical for the state machine
```

Every skill above but `flow-research` and `flow-contracts` requires the `spectre` CLI
(`go install github.com/tweety53/spectre/cmd/spectre@latest`); those two need none — reading a
spectre tree, or a contract file, is reading markdown.
