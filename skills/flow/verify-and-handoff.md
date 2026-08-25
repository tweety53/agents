# Verify, stage, and hand off

Loaded by `skills/flow/SKILL.md` once `skills/flow/review-panel.md` closes clean. Carries the stage
order design.md's `workspace-export-lint-merge` and `run-instructions-reorder` decisions produce:
**verify → stage-diff → run-instructions → write-in-progress** — the old
`do.run-instructions → do.workspace-export → do.lint-and-test → do.stage-diff → do.write-in-progress`
order, with the middle two merged into one `flow.verify` stage and `run-instructions` moved to
immediately before the state write.

## Verify

**Load `skills/myflow-contracts/worktree-resolution.md`** before resolving this run's worktree set,
below.

```bash
myflow stage begin -command '/flow' \
  -stage flow.verify \
  -harness <harness> \
  -session-token mf-<literal-token> \
  <name>
```

This stage is design.md's `workspace-export-lint-merge`: the old `do.workspace-export` and
`do.lint-and-test` stages, unconditional automated pass/block checks with nothing interactive
between them, merged into one mark.

**First, validate the section and export what it declares — with the script, not by eye.** Run

```bash
prepare-workspace.sh <worktree>
```

once per worktree in this run's resolved set — the same set **2. Isolate the workspace**
(`skills/flow/implement.md`) resolved, non-empty by construction — never a raw read of the state
file's `worktrees` map. Per **Resolving a change's worktrees**
(`skills/myflow-contracts/worktree-resolution.md`), report an empty resolved set and do not proceed.

`prepare-workspace.sh` runs `check-workspace-isolation.sh` against the worktree first, then — only
if that passes — derives and exports the variables the project's `## workspace isolation` section
declares, resolved against the workspace id, and prints one `KEY=value` line per exported variable
to stdout. Exit 0 means the printed lines are what to carry forward into `## lint` and `## test`
below (nothing printed means the project declares no `## workspace isolation` section). A non-zero
exit is the dropped-row case (exit 1, relay the script's own lines verbatim and stop) or the
cannot-answer case (exit 2, stop the same way) — stop **before** `## lint` and `## test`, without
writing the state file.

**A declared `cache index` row is never among the printed `KEY=value` lines** — the script reports
it by name on stderr instead. On an exit-0 run whose stderr names a `cache index` row, probe the
project's cache here, claim a free index atomically, and record that claim in the cache itself under
an entry naming this workspace, per **The cache index**
(`skills/myflow-contracts/workspace-isolation.md`).

**When the script cannot be located**, apply the same rules by hand from **Project configuration** (`skills/myflow-contracts/project-configuration.md`)
and **Workspace isolation** (`skills/myflow-contracts/workspace-isolation.md`), and say in the handoff that the validation and
export were performed manually and why.

**This step does not call the project's `create` command.** `create` is called by whatever starts
the project's applications, per **Project configuration**
(`skills/myflow-contracts/project-configuration.md`), and this step starts none of them — it
exports, lints, tests, and hands off.

Run the `## lint` and `## test` commands from `<project>/.myflow/project.md` (auto-detect if
absent) and show the output. **Nothing runs them later** — `/flow`'s integrate phase has no
verification gate — so a non-zero exit blocks this handoff.

**Load `skills/myflow-contracts/session-records.md`** before reading the render outcome below.

**Confirm this run recorded a ledger:**

```bash
myflow record render -change <name> -kind ledger -repo <abs-worktree>
```

Read the outcome word, not the exit code. `rendered: <dest>` is ordinary. **`MISSING: ledger — no
rows for <name>` means this run recorded no dispatch at all**, reported plainly here rather than
discovered later. `journalled: ledger` and a non-zero exit are reported the same way. None of these
gates or stops the run — unlike the lint and test exits above. The outcome words are the table under
**Rendering the session records** (`skills/myflow-contracts/session-records.md`).

```bash
myflow stage end -command '/flow' -stage flow.verify -outcome completed <name>
```

## Stage, excluding the planning paths

```bash
myflow stage begin -command '/flow' -stage flow.stage-diff -harness <harness> -session-token mf-<literal-token> <name>
```

Confirm every intended task checkbox is `[x]`, and that `git log <merge-base>..HEAD` shows one
commit per completed task, with every fix-round and red-task-partner fixup already folded in via
`git rebase --autosquash` — no stray `fixup!` commit should remain unsquashed, unless a PR already
exists (below).

In **every** affected worktree:

```bash
git -C <worktree> status
git -C <worktree> log <merge-base>..HEAD --oneline
```

> **`<project>/spectre/changes/` and `<project>/docs/superpowers/` are never part of a task
> commit.** `<project>/spectre/specs/` is not one of them — a capability spec belongs in the task
> commit that implements its requirement. This step only confirms nothing slipped in.

**Load `skills/myflow-contracts/git-boundaries.md`** before committing below.

**The one push exception.** Every task and fixup commit already sits on the branch, unpushed. If the
state file records a `prUrl`, a PR is already open, so this run also commits
`<project>/spectre/changes/` and `<project>/docs/superpowers/` and pushes everything to the PR
branch; otherwise this step commits and pushes nothing. On that path only — and in this order — run
`myflow record render -change <name> -kind all -repo <worktree>`; then `commit-split.sh <worktree>
<name> "<impl-msg>" "chore(spectre): plan and session records"`; then push the branch. `<impl-msg>`
covers working-tree edits the operator made at the human gate without staging them — derive it the
same way a fixup commit's subject is derived — `fix(<module>): <what changed since the last task
commit>`.

The render overwrites in place. `MISSING: <kind>` means the store holds no rows of that kind and
nothing was written — report it. **A non-zero exit means a destination was refused or could not be
written** — report it, and continue committing the fix.

```bash
myflow stage end -command '/flow' -stage flow.stage-diff -outcome completed <name>
```

## Resolve the run instructions

```bash
myflow stage begin -command '/flow' -stage flow.run-instructions -harness <harness> -session-token mf-<literal-token> <name>
```

Resolve the run instructions for the handoff's `Run it:` section. It writes no file.

- **Every app root is absolute**, resolved from `git worktree list` or the state file's `worktrees`
  keys. Never a relative sibling path, and never a main-checkout path while a worktree holds the
  work.
- **Every start command comes from `<project>/.myflow/project.md`'s `## run`**, with every path
  made absolute.
- **Every URL is the one this worktree resolved**, never the project's declared base. Resolve each
  URL from this worktree's workspace id. A project that declares no isolation resolves nothing.
  An application whose port is fixed outside that project's own repository keeps its default,
  named with a short note.
- Apps in scope come from `## apps` in `<project>/.myflow/project.md`, or auto-detection.
- **Where the project declares no runnable application**, resolve the `## lint` and `## test`
  commands instead.

```bash
myflow stage end -command '/flow' -stage flow.run-instructions -outcome completed <name>
```

## Write `IN_PROGRESS`

```bash
myflow stage begin -command '/flow' -stage flow.write-in-progress -harness <harness> -session-token mf-<literal-token> <name>
```

Write the state file: `IN_PROGRESS` from `STARTED`, otherwise **the state exactly as read**.
Populate `worktrees` with one absolute-path key per affected worktree and its merge base. Carry
`artifactUrl` (always `null` under `/flow`), `jiraIssue`, `planningEffort` (always `null`),
`models.default` (always `null` — `/flow` resolves models from the settings store per run, never
records a value into the per-change state) and `prUrl` forward verbatim. The state file lives
outside the repo — never `git add` it.

```bash
myflow stage end -command '/flow' -stage flow.write-in-progress -outcome completed <name>
```

**Produce the handoff's `Records:` count**, one call per affected worktree:

```bash
myflow record journal-count -change <name> -C <abs-worktree>
```

**Produce the handoff's `Costs:` line the same way**, one call per affected worktree:

```bash
myflow record cost-status -change <name>
```

It exits 0 always — `unknown` included. Render exactly what it printed.

```
## Implementation staged — review and test | Implementation committed — review and test

**Change:** <name>
**Panel:** clean — required: Primary, Principles, Code review (low); on-demand: <Bugbot and/or Security, or "none — not requested">
**Staged:** N/N tasks staged and uncommitted | N/N tasks committed on branch | committed, plus one planning-artifacts commit, and pushed to the PR branch
**Records:** all writes reached the store | N write(s) journalled — the store was unreachable | unknown — the journal could not be counted
**Costs:** <the line `myflow record cost-status` printed>
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)
**Jira description (pre-edit):** <the text as it stood before the write, verbatim in a fenced block, inside <details> when long> | omitted — this run wrote no description

Worktree:   <absolute worktree path>

Run it:
  <command>          # <app or check name>
  <command>

Review the diff, then run it:
  git -C <absolute worktree path> diff <merge base>..HEAD
  open -na "IntelliJ IDEA" --args "<absolute worktree path>"

Re-run this command to fix anything you find, or bare to move on to integrating it.

Next:
/flow <name>
```

**Heading and `Staged:` line select from whether the worktree carries any commits yet** — a
creating run's very first `IN_PROGRESS` write, reached before any task committed, would be an
anomaly (implementation always commits per task), so in practice this always reads "committed" —
the "staged and uncommitted" alternative is carried only for symmetry with the phrasing an operator
resuming mid-panel might see, and should not occur in an ordinary run. **The `Panel:` line's
on-demand clause is `/flow`'s own** — it replaces the retired roster/preset line with what
**Review panel** (`skills/flow/review-panel.md`) actually decided this run: which of Bugbot and
Security ran, by explicit request, or that neither did.

**The `Records` line is printed on every run of this branch, journalled or not.** **The `Costs:`
line is printed the same way — always, `unknown` included.**

The pre-edit description line is present only on a fix run that synced the description in **3.
Documenting a fix** (`skills/flow/implement.md`), and reproduces that text without summarising or
reflowing it.

## Guardrails

- **Commit per task and per fixup** — never `<project>/spectre/changes/` or
  `<project>/docs/superpowers/` in a task or fixup commit. **Never push, merge, or open a PR** —
  except the `prUrl` exception above.
- **Never** run `finishing-a-development-branch`.
- **Never** create a second worktree for the same change.
- **Never** advance the state from `IN_PROGRESS`; write back what you read.
- **Never** hand off with an open finding of any severity, or a stale clean result.
- **Never** mark a task's checkbox before that task's review passes.
