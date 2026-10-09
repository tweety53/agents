# Project configuration

**This file is the canonical definition of `<project>/.flow/project.md`.** Skills reference it by
name; none of them restate the format. If a skill and this file ever disagree, this file
wins.

flow is installed globally and runs in any repository, so it must never carry one project's
apps, ports, task names, or credentials in its own files. Everything project-specific lives in
an optional Markdown file **in the project being worked on**:

```text
<main checkout>/.flow/project.md
```

It is a normal Markdown document, committed with the project, read by whichever skill needs it.
Sections are `## <key>` headings; their bodies are prose, tables, or fenced command blocks —
whatever reads best for a human, since agents and humans read the same file.
Some keys are not: a key whose row says its body is resolved and validated rather than read holds
exactly what that row specifies and nothing else. The row is authoritative.

| Key | Supplies |
|-----|----------|
| `## apps` | Every application in the change's blast radius: display name, **absolute** repo root of the main checkout, kind/stack, local URL, and when to consider it in scope. |
| `## run` | How to start the stack and individual services locally — the actual commands, including any parameter that must point at another app's root. |
| `## stop` | Optional. The command that stops the project's local stack, run by bare `/flow` before its stack-stopped check so nothing holds a port or file handle open in a worktree about to be removed. Same shape as `## worktree setup`: one command per line inside a fence, read literally and run in order from the worktree; prose outside a fence is never run. **Absent means that check is skipped, not failed** — cleanup proceeds on the strength of the other two checks. A body with no fenced command declares no command and is skipped the same way. A project with no stack should say so here rather than omit the key, so its absence is a recorded fact rather than an oversight. |
| `## credentials` | Local-development-only sign-in credentials (which app, user, password) and what seeds them. Never deployed-environment secrets. |
| `## test` | The command(s) that run the project's tests. |
| `## worktree setup` | Optional. A fenced command block run once per worktree, from the worktree root, immediately after the run creates the worktree — **2. Isolate the workspace** (`skills/flow/implement.md`) on `/flow`, the kickoff section (`skills/flow-fast/SKILL.md`) on a `/flow-fast` creating run — and before anything else touches the tree — the place for a build that a gitignored, embedded artifact needs before the project's first `go test`/`go build` can succeed. Absent means nothing runs. Same shape as `## lint` and `## test`: one command per line inside the fence, read literally and run in order; resolved through `project-get.sh <worktree> "worktree setup"`, whose exit 1 is the absent case. |
| `## lint` | The command(s) that verify lint, and the auto-fix command to run first. |
| `## standards` | The project's own written standards: the files the principles reviewer receives, plus any opt-in shared rule the project has adopted. This same list is both the opt-in list and the reviewer's standards list. Absent, no standards resolve: the reviewer gets an empty list, never auto-detected files. |
| `## jira` | Optional. The project's Jira project key(s), or the literal `none` — this body holds those and nothing else, never free-form prose. Each key must match the `[A-Z]{2,10}` shape **in its entirety**, as required under **Follow-up issues** (`skills/flow-contracts/jira-followups.md`), which also states how the body is split into candidate keys and what becomes of one that does not match — this value reaches a JQL query, so it is constrained like the attacker-influenced input it is. Governs whether `/flow`'s creating run asks about an issue at all — see **Jira integration** (`jira-integration.md`). |
| `## default landing route` | Optional. One of the literals `pull request`, `merge and push` or `manual` — the value is the row literal the body's first non-blank line equals or leads, per the resolution paragraph below; the rest of that line, like the lines below it, is documentation for the reader, never read. Used as `skills/flow/integrate.md`'s landing question's own `(default, recommended)` option; absent, or a head matching none of the three literals, is reported by name and dropped, falling back to `pull request`. |
| `## handoff` | Optional. One of the literals `required` or `none` — the value is the row literal the body's first non-blank line equals or leads, per the resolution paragraph below; the rest of that line, like the lines below it, is documentation for the reader, never read. Read by `/flow-fast` alone (`skills/flow-fast/SKILL.md`): `none` lands the change in the same invocation, straight after the change summary; `required`, or absent, stops after the summary so the operator reviews the branch first and re-runs `/flow-fast <name>` bare to land it. `/flow` always hands off at `IN_PROGRESS` and never reads this key. A head matching neither literal is reported by name and dropped, resolving as if the key were absent. |
| `## decisions` | Optional. The literal `recommended` — the value is the row literal the body's first non-blank line equals or leads, per the resolution paragraph below; the rest of that line, like the lines below it, is documentation for the reader, never read. `recommended` widens the run's recommended-defaults mode to the planning asks — never the pivot, an outward-facing or irreversible action, or an integrate or cleanup prompt (**Auto-resolution**, `skills/flow-contracts/operator-prompts.md`, canonical for what the mode covers and how a taken default is recorded); absent, the key is dormant and every prompt behaves exactly as its call site states. A head matching none of the literals is reported by name and dropped, resolving as if the key were absent. |
| `## progress` | Optional. The literal `quiet` — the value is the row literal the body's first non-blank line equals or leads, per the resolution paragraph below; the rest of that line, like the lines below it, is documentation for the reader, never read. `quiet` turns on `/flow`'s quiet-progress mode (**Quiet progress**, `skills/flow-contracts/pipeline.md`, canonical for what the mode silences and what still prints); absent, the key is dormant and the run reports progress exactly as it does today. A head matching none of the literals is reported by name and dropped, resolving as if the key were absent. |
| `## workspace isolation` | Optional. The resources an apply worktree runs against its own copy of: for each one, the environment variable that carries it and the value it falls back to, plus the commands that create them, remove them and report which of them survived. Most of those values are derived from the workspace id, and one is not — the cache index is claimed at run time and is not derived from it. Its rows are resolved and validated rather than read, so the two tables specified in `skills/flow-contracts/project-configuration-isolation.md` are the whole of what this body means to the resolver — prose beside them is for the reader, and is where a project records what it has deliberately **not** isolated. Absent means the project is not isolated, which is a supported state and not a misconfiguration — again, see below. Which resources there are, and how each derived value is derived, is stated once under **What the id derives** (`skills/flow-contracts/workspace-isolation.md`). |
| `## visual verification` | Optional. What the `flow.visual-verify` stage validates before it runs: which UI paths in a change's diff trigger it, where captured screenshots land, the `setup`/`verify`/`capture`/`fingerprint`/`specs` commands, and an optional `regression checkout` naming the repository the spec and its PNGs commit to, with `regression repo` recording the identity — never an authorisation — that checkout's real `origin` must equal. Its rows are resolved and validated rather than read, so the two tables specified in `skills/flow-contracts/project-configuration-visual.md` are the whole of what this body means to the resolver — prose beside them is for the reader. Absent means the stage is not configured for this project, a supported state and not a misconfiguration. Mechanically enforced by `<agents repo>/scripts/check-visual-verification.sh`, canonical for what it checks. |
| `## review panel citation check` | Optional. A single fenced command block and nothing else in the section — unlike `## workspace isolation` and `## visual verification` above, this key gets no dedicated parsing guard: the whole of what it means to the resolver is "a fenced block exists, read literally". **Absent means the whole pre-panel step is skipped**, exactly like every other optional key. The command runs from the apply worktree only when `check-panel-citation-trigger.sh` exits 0 (**Guard resolution**, `skills/flow-contracts/pipeline.md`, is how `skills/flow/review-panel.md` names it, which is also canonical for the full wiring procedure); its combined stdout+stderr is captured verbatim and its exit code is never gating. |

**Load `skills/flow-contracts/project-configuration-standards.md`** only when resolving a `## standards` entry.

**Load `skills/flow-contracts/project-configuration-isolation.md`** only when `prepare-workspace.sh` or `check-cleanup-complete.sh` cannot be located, or when writing or reviewing a `## workspace isolation` section.

**Load `skills/flow-contracts/project-configuration-visual.md`** only when `flow.visual-verify` begins, or when writing or reviewing a `## visual verification` section.

**A single-line-literal key's value is the row literal its body's first non-blank line —
whitespace-trimmed, surrounding backticks removed — equals or leads at a word boundary: the head
matches a literal on equality, or when the literal opens the head and the byte after it cannot
extend it into a longer word, and nothing else is normalized** — no case-folding, no synonym list.
Whatever follows the literal on the head line — prose the operator wrote there — and the lines
below it are documentation for the reader, never read. A head that matches no literal is a
malformed row: report it by name (quoting what was found) and drop it, resolving as if the key
were absent. The keys matched this way are `## default landing route`, `## decisions`,
`## progress` and `## handoff`. `project-get.sh <root> <key> --enum <literal>...` performs this resolution: exit 0 prints the
matched literal, exit 1 the key is absent, exit 3 the head matches no literal — its one stderr line
quotes the head.

## Where the agents repository is

`<agents repo>` is the root of the flow agents repository on this machine — the checkout flow
is authored in and installed from. Everything that resolves against it is named here so that no
skill has to work it out for itself.

**`AGENTS_DATA` wins when it is set.** Take it as the root and go no further; it exists so a
machine with an unusual layout can state the answer instead of having one derived.

**Otherwise derive it from the skill you are reading, in two steps.**

1. Take the directory holding that `SKILL.md` — `skills/flow/`, `skills/flow-status/`,
   whichever one you are in — and resolve **that directory** to its physical path, following it if
   it is a symlink.
2. `<agents repo>` is **two levels above** the resolved directory: up out of the skill's own
   directory, then up out of `skills/`.

**The link to resolve is the per-skill one, never the `skills/` directory above it.** See
**The per-skill link, not the `skills/` directory**
(`skills/flow-contracts/project-configuration-rationale.md`)
for the measured global-install layout this guards against.

The same two steps are correct for a project-local install, where nothing is a symlink at all. See
**Project-local installs need no link**
(`skills/flow-contracts/project-configuration-rationale.md`)
for why.

**Confirm the derived root before using it, and treat a miss as an absence rather than a nearer
guess.** Check that the path you are about to use exists under it. That is never a licence to skip
the step, and never a reason to search the filesystem for a checkout that might be one. See
**Confirm the derived root before using it**
(`skills/flow-contracts/project-configuration-rationale.md`)
for the copied-directory case this guards against.

**Roots in `## apps` are main checkouts.** When an apply worktree exists for the change, resolve
that app's root from `git worktree list` in its repo and use the worktree root instead — never a
relative sibling path, and never the main checkout while a worktree holds the work.

**The file is optional, and every key within it is optional.**

- **Absent file, or absent key** → **auto-detect from the repository**: read the build files,
  scripts, and existing docs actually present, and derive what is needed from them.
- **Never fail** because the file is missing. A project with no `<project>/.flow/project.md` is a
  supported, ordinary case, not an error.
- **Never assume another project's layout.** If a value cannot be detected, say so and ask, or
  emit an explicit `TBD` in generated output. Do not fall back to app names, ports, Gradle or
  npm task names, URLs, or credentials remembered from a different repository.
