# Verify, stage, and hand off

Loaded by `skills/flow/SKILL.md` once `skills/flow/review-panel.md` closes clean. Carries the stage
order design.md's `workspace-export-lint-merge` and `run-instructions-reorder` decisions produce:
**verify → visual-verify → stage-diff → run-instructions → write-in-progress** — the old
`do.run-instructions → do.workspace-export → do.lint-and-test → do.stage-diff → do.write-in-progress`
order, with the middle two merged into one `flow.verify` stage and `run-instructions` moved to
immediately before the state write. `flow.visual-verify` sits between `flow.verify` and
`flow.stage-diff` — a later insertion, not part of either decision above.

## Verify

**Load `skills/flow-contracts/worktree-resolution.md`** before resolving this run's worktree set,
below.

```bash
flow stage begin -command '/flow' \
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
(`skills/flow-contracts/worktree-resolution.md`), report an empty resolved set and do not proceed.

`prepare-workspace.sh` runs `check-workspace-isolation.sh` against the worktree first, then — only
if that passes — derives and exports the variables the project's `## workspace isolation` section
declares, resolved against the workspace id, and prints one `KEY=value` line per exported variable
to stdout. Exit 0 means the printed lines are what to carry forward into `## lint` and `## test`
below (nothing printed means the project declares no `## workspace isolation` section). A non-zero
exit is the dropped-row case (exit 1, relay the script's own lines verbatim and stop) or the
cannot-answer case (exit 2, stop the same way) — stop **before** `## lint` and `## test`, without
writing the state file.

**Load `skills/flow-contracts/workspace-isolation.md` only when `prepare-workspace.sh` exited
non-zero, or exited 0 with stderr naming a `cache index` row** — the procedures for both live
there; the ordinary exit-0 run loads nothing.

**A declared `cache index` row is never among the printed `KEY=value` lines** — the script reports
it by name on stderr instead. On an exit-0 run whose stderr names a `cache index` row, probe the
project's cache here, claim a free index atomically, and record that claim in the cache itself under
an entry naming this workspace, per **The cache index**
(`skills/flow-contracts/workspace-isolation.md`).

**When the script cannot be located**, apply the same rules by hand from **Project configuration** (`skills/flow-contracts/project-configuration.md`)
and **Workspace isolation** (`skills/flow-contracts/workspace-isolation.md`), and say in the handoff that the validation and
export were performed manually and why.

**This step does not call the project's `create` command.** `create` is called by whatever starts
the project's applications, per **Project configuration**
(`skills/flow-contracts/project-configuration.md`), and this step starts none of them — it
exports, lints, tests, and hands off.

Run the commands `project-get.sh <worktree> lint` and `project-get.sh <worktree> test` print
(auto-detect on exit 1) and show the output. **Nothing runs them later** — `/flow`'s integrate phase
has no verification gate — so a non-zero exit blocks this handoff.

**Load `skills/flow-contracts/session-records.md`** before reading the render outcome below.

**Confirm this run recorded a ledger:**

```bash
flow record render -change <name> -kind ledger -repo <abs-worktree>
```

Read the outcome word, not the exit code. `rendered: <dest>` is ordinary. **`MISSING: ledger — no
rows for <name>` means this run recorded no dispatch at all**, reported plainly here rather than
discovered later. `journalled: ledger` and a non-zero exit are reported the same way. None of these
gates or stops the run — unlike the lint and test exits above. The outcome words are the table under
**Rendering the session records** (`skills/flow-contracts/session-records.md`).

```bash
flow stage end -command '/flow' -stage flow.verify -outcome completed <name>
```

## Visual verification

```bash
flow stage begin -command '/flow' \
  -stage flow.visual-verify \
  -harness <harness> \
  -session-token mf-<literal-token> \
  <name>
```

Reads the `## visual verification` section, canonical in
`skills/flow-contracts/project-configuration.md`. This stage owns its procedure — nothing else in
this pipeline restates it. Resolve once per worktree in this run's resolved set, the same set
**Verify** above resolved:

1. **Resolve the section** — read that worktree's own `<project>/.flow/project.md` directly, by
   its own shape and closed vocabulary. A project declaring no section → this worktree prints
   `Visual: not configured` and is skipped for the rest of this stage.
2. **Match the diff — with the guard, not by eye.** Run

   ```bash
   git -C <worktree> diff --name-only <merge-base>..HEAD | check-visual-trigger.sh <worktree>
   ```

   Exit 0 → at least one changed path matched a declared `ui paths` glob; continue. Exit 1 → this
   worktree prints `Visual: no UI paths touched` and is skipped for the rest of this stage. Exit 2 →
   the guard could not answer (this should not happen here, since step 1 already confirmed the
   section resolves) — report its stderr and skip this worktree the same way exit 1 does.
   `check-visual-trigger.sh` owns the glob semantics (`**` spanning directories, a leading
   dot-slash prefix, an absolute glob, a glob with a space); nothing here restates them.
3. **Run `setup`, if declared.** A non-zero exit blocks, printing the command verbatim.
4. **Probe before starting anything.** Probe the URL of each app `ui paths` matched, resolved for
   this worktree per **What the id derives** (`skills/flow-contracts/workspace-isolation.md`) —
   never the project's declared default. If nothing answers, start the stack from `## run` and
   record that this stage started it — needed at step 10.
5. **Run `verify`.** A non-zero exit blocks.
6. **Capture** — author a spec covering the views this change touched, then run `capture` with
   `<spec>` substituted for the spec's path. `screenshots`'s root-not-leaf shape is canonical in
   `skills/flow-contracts/project-configuration.md`; nothing here restates it. **`capture` creates
   this change's baseline**: writing a PNG that does not yet exist is its success path, not a
   failure — `verify` is the regression gate over an already-committed baseline, `capture` is not,
   and only a `capture` failure for some other reason blocks (see **Blocking** below).
7. **Read every captured PNG — resolve their paths with the guard, not by eye.** Run

   ```bash
   resolve-visual-screenshots.sh <worktree> <spec's basename>
   ```

   Exit 0 prints one absolute PNG path per line — that is what selects this run's fresh output from
   the committed baseline PNGs already sitting under the same `screenshots` root, rather than joining
   `screenshots` with a guessed filename. Exit 1 (zero matches) blocks, indistinguishable from a view
   that never rendered; exit 2 (cannot answer) blocks the same way. **Read every printed path** — no
   script can do that — and state, per view, what was seen. An unreadable PNG is reported and blocks
   too.
8. **Write `<changeRoot>/visual-verification.md`** — one entry per view: its absolute screenshot
   path, resolved by the same recursive search step 7 used, and what was seen.
9. **Commit the spec and its PNGs, and stop there.** A declared `regression checkout` receives
   them; with none declared, commit to the change's own branch instead. **Never push** — see
   `no-automatic-push` (design.md): a file inside a repository cannot authorise a push to another
   repository, so no guard here grants one. `regression repo` still records which repository the
   checkout is expected to be, and `check-visual-verification.sh` still reports a mismatch against
   its real `origin`, but that is an identity assertion, not an authorisation. When a commit landed
   in a `regression checkout`, print the push command for the operator to run by hand:

   ```bash
   git -C <regression checkout> push
   ```
10. **Stop the stack only if step 4 started it.** A stack the operator already had running is left
    alone.

**Blocking.** This stage blocks the `IN_PROGRESS` handoff on: a failed `setup`, a failed `verify`,
a genuine `capture` failure — **never a first-run snapshot write, which is `capture`'s own success
path per step 6 above** — a stack that could not be started, an unreadable PNG, and **a defect the
agent sees in a captured screenshot — even when every assertion passed.** That last one is the whole
point of this stage: three defects have shipped invisible to a diff, a five-pass review panel and
both test suites, and obvious the moment the page was opened.

```bash
flow stage end -command '/flow' -stage flow.visual-verify -outcome completed <name>
```

## Stage, excluding the planning paths

```bash
flow stage begin -command '/flow' -stage flow.stage-diff -harness <harness> -session-token mf-<literal-token> <name>
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

**Load `skills/flow-contracts/git-boundaries.md`** before committing below.

**The one push exception.** Every task and fixup commit already sits on the branch, unpushed. If the
state file records a `prUrl`, a PR is already open, so this run also commits
`<project>/spectre/changes/` and `<project>/docs/superpowers/` and pushes everything to the PR
branch; otherwise this step commits and pushes nothing. On that path only — and in this order — run
`flow record render -change <name> -kind all -repo <worktree>`; then `commit-split.sh <worktree>
<name> "<impl-msg>" "chore(spectre): plan and session records"`; then push the branch. `<impl-msg>`
covers working-tree edits the operator made at the human gate without staging them — derive it the
same way a fixup commit's subject is derived — `fix(<module>): <what changed since the last task
commit>`.

The render overwrites in place. `MISSING: <kind>` means the store holds no rows of that kind and
nothing was written — report it. **A non-zero exit means a destination was refused or could not be
written** — report it, and continue committing the fix.

```bash
flow stage end -command '/flow' -stage flow.stage-diff -outcome completed <name>
```

## Resolve the run instructions

```bash
flow stage begin -command '/flow' -stage flow.run-instructions -harness <harness> -session-token mf-<literal-token> <name>
```

Resolve the run instructions for the handoff's `Run it:` section. It writes no file.

- **Every app root is absolute**, resolved from `git worktree list` or the state file's `worktrees`
  keys. Never a relative sibling path, and never a main-checkout path while a worktree holds the
  work.
- **Every start command comes from `<project>/.flow/project.md`'s `## run`**, with every path
  made absolute.
- **Every URL is the one this worktree resolved**, never the project's declared base. Resolve each
  URL from this worktree's workspace id. A project that declares no isolation resolves nothing.
  An application whose port is fixed outside that project's own repository keeps its default,
  named with a short note.
- Apps in scope come from `## apps` in `<project>/.flow/project.md`, or auto-detection.
- **Where the project declares no runnable application**, resolve the `## lint` and `## test`
  commands instead.
- **On a fix run, reload before this stage ends.** A fix run hands the operator a diff and the run
  instructions this stage resolves; if the applications those instructions name are still serving
  pre-fix code, the operator reviews one thing and runs another — measured, not hypothetical, in
  this repository's own `<project>/stats/internal/web/embed.go`, whose `//go:embed all:dist` makes a running
  daemon blind to an SPA source change until it is rebuilt. Before this stage ends, rebuild and
  restart every application the run instructions above name — the ones just resolved, and no
  others — from the project's `## run` commands.

  Resolve the reload from the two keys the project already declares, in order: its `## stop`, when
  it declares a command, then its `## run`. **A `## stop` that deliberately declares no command
  means there is nothing to stop, not that this rule is skipped** — the reload is then whatever
  `## run` does to bring the application up fresh. No new project-configuration key is added for
  this, and none is needed.

  **Never the flow dev stack.** `flowd` on `127.0.0.1:4173`, its `flow-postgres` container and the
  `flow` database inside it are never stopped, restarted or dropped by any run —
  `<project>/CLAUDE.md` states that prohibition and this rule does not weaken it. Where a project's
  own `## run` names that service, the prohibition wins over this reload rule, never the reverse.
  This is separate from **Visual verification**'s own start/stop rule above (step 10): that stage
  stops only the stack it started for its own probe, and that rule is not restated here. This rule
  reloads whatever the run instructions name, on every fix run, regardless of whether that stage ran
  or started anything.

  **Where every application `## apps` names is one this prohibition covers, the reload is
  nothing — stated, not silently skipped.** This repository is that case: its `## apps` names
  exactly one URL-bearing application, the myflow stats daemon on `127.0.0.1:4173`, and that is the
  protected daemon itself. A fix run against this repository therefore reloads nothing before this
  stage ends, and the handoff states which application was skipped and why:

  ```
  Not reloaded: myflow stats daemon (http://127.0.0.1:4173) — protected, see
  <project>/CLAUDE.md's "Never stop the dev workspace's stats service or its storage".
  ```

  `flow.visual-verify`'s own `make ui-test-up`/`make ui-test-down` pair (step 4 and step 10 above)
  is a different mechanism entirely — it starts and stops the disposable UI-test stack on
  `127.0.0.1:4174` for that stage's own probe, and `4174` is not an application `## apps` names at
  all, so it is never this rule's reload target.

  **A reload that fails blocks this stage**, naming the application and what the command printed —
  handing over run instructions that cannot be followed is the failure this rule exists to prevent.
  Where the project declares no runnable application, there is nothing to reload and the rule is
  satisfied by saying so, not by silently skipping it.

  Two worked examples, both real, and they resolve differently:

  ```bash verified:read from /Users/tweety53/Projects/gymie/.flow/project.md and this repository's own .flow/project.md
  # gymie — `## stop` declares a command, so reload is stop-then-run:
  ./gradlew devStop
  docker compose up -d && ./gradlew devStart -PfrontendRoot=<abs> -PadminFrontendRoot=<abs>

  # this repository — `## apps` names only the protected daemon, so the reload is nothing at all;
  # see the paragraph above for the handoff line this produces instead of a command.
  ```

```bash
flow stage end -command '/flow' -stage flow.run-instructions -outcome completed <name>
```

## Write `IN_PROGRESS`

```bash
flow stage begin -command '/flow' -stage flow.write-in-progress -harness <harness> -session-token mf-<literal-token> <name>
```

Write the state file: `IN_PROGRESS` from `STARTED`, otherwise **the state exactly as read**.
Populate `worktrees` with one absolute-path key per affected worktree and its merge base. Carry
`artifactUrl` (always `null` under `/flow`), `jiraIssue`, `planningEffort` (always `null`),
`models.default` (always `null` — `/flow` resolves models from the settings store per run, never
records a value into the per-change state) and `prUrl` forward verbatim. The state file lives
outside the repo — never `git add` it.

```bash
flow stage end -command '/flow' -stage flow.write-in-progress -outcome completed <name>
```

**Produce the handoff's `Records:` count**, one call per affected worktree:

```bash
flow record journal-count -change <name> -C <abs-worktree>
```

**Produce the handoff's `Costs:` line the same way**, one call per affected worktree:

```bash
flow record cost-status -change <name>
```

It exits 0 always — `unknown` included. Render exactly what it printed.

```
## Implementation staged — review and test | Implementation committed — review and test

**Change:** <name>
**Panel:** clean — roster: <the resolved slot list this run dispatched>; substituted: <slot(s) dispatched as general-purpose in place of their own agent type, or "none">; added this run: <slot(s) an explicit operator instruction added beyond the resolved list, or "none — resolved list ran alone">
**Visual:** not configured | no UI paths touched | <view>: <absolute screenshot path>[, <view>: <absolute screenshot path> …][ — push with: git -C <regression checkout> push]
**Staged:** N/N tasks staged and uncommitted | N/N tasks committed on branch | committed, plus one planning-artifacts commit, and pushed to the PR branch
**Records:** all writes reached the store | N write(s) journalled — the store was unreachable | unknown — the journal could not be counted
**Costs:** <the line `flow record cost-status` printed>
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
resuming mid-panel might see, and should not occur in an ordinary run. **The `Panel:` line is
`/flow`'s own** — it states what **Review panel** (`skills/flow/review-panel.md`) actually
dispatched this run: the resolved roster, any slot substituted as a general-purpose agent per
`unspawnable-id-substitutes`, and any slot an explicit operator instruction added beyond the
resolved list.

**The `Records` line is printed on every run of this branch, journalled or not.** **The `Costs:`
line is printed the same way — always, `unknown` included.**

**The `Visual:` line reports `flow.visual-verify`'s own outcome.** Every screenshot path in it is
absolute, per **Handoff output** (`skills/flow-contracts/pipeline.md`)'s every-path-is-absolute
rule — the operator must be able to open the PNG. **Its push clause appears only when step 9
committed to a `regression checkout`** — the stage never pushes itself, per `no-automatic-push`, so
this is the command the operator runs by hand to land that commit.

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
