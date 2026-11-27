---
name: flow-fast
description: Minimal-ceremony /flow variant — one invocation from Jira key to landed change. A git worktree for isolation and nothing else, implementation and review panel as a headless decider decides, project lint plus targeted tests, then the project's default landing route and cleanup; no spectre artifacts, no state file. Use for /flow-fast.
allowed-tools: Bash(flow:*)
license: MIT
---

Do the work the way a careful engineer does it by hand — read, edit, verify, commit, land — and
record it the way `/flow` does: every `flow.*` stage mark below, in this order, so the stats views see a `/flow-fast` run as the same pipeline. Nothing else of
`/flow` survives here. There is no spectre change, no `proposal.md`/`design.md`, no state file
and no three states, no staged-diff gate, no archive, and no pipeline guard. Dispatch and the review panel
follow the recorded decision (**Dynamic decisions** below). The only isolation is git's: a
worktree on its own branch. `flow-guard prepare-workspace`, the per-change
database, bucket, ports and cache index of **Workspace isolation**
(`skills/flow-contracts/workspace-isolation.md`) are never set up.

**Announce at start:** "Using flow-fast for change `<name>`."

**No flags.** The only argument is the change description or Jira key on a creating run, or fix
instructions on a re-run; report anything else rather than ignoring it. One exception:
`--base <branch>` on a creating run, as `/flow` takes it (`skills/flow/SKILL.md`).

**Every mark is a literal call.** `<harness>` is this harness's name, `ff-<literal-token>` is one
token generated once at the start of the run, shaped as **Stage marks**
(`skills/flow-contracts/pipeline.md`) states, and typed the same in every `flow record` command —
never a shell substitution, per the CLI's own usage text. A mark carries none. A `flow stage mark` line
marks each `-stages` key begun and ended `completed`, in order: the stages `/flow-fast` has nothing
to run for; the mark stays so the run's stage set matches `/flow`'s.

**Guardrails, the whole list.** Never use git stash pop/apply/drop; if a worktree looks wrong, stop and report BLOCKED with the exact `git status` output. Never dispatch a subagent the recorded decision does not name —
an implementer per group and its gated reviewer, the decision's panel dispatches, the panel-fix subagent; never
a planner or a verifier, save the in-run fix loop of **Pipeline defects found mid-run**
(`skills/flow-contracts/pipeline.md`), which a `/flow-fast` run follows too, and the
self-review pass's own dispatches (`skills/flow-self-review/SKILL.md`). Never ask a model,
planning-effort or review question. Never write `<project>/spectre/`,
save that pass's `self-review-<name>` change and an in-run fix's `pipeline-fix-<name>-<k>` change, or a state file. Never set up workspace isolation and never call a
guard script a cited `skills/flow/` section does not call itself. Never push to a branch other
than the one the landing route names.

## Stage keys

Every stage `/flow-fast` marks uses a `flow.*` key — the same keys `/flow` marks, since both
commands drive the same pipeline stages. The full key list, in the order each step below marks
them:

| Step | Keys |
|------|------|
| **1. Kickoff** | `flow.plan` |
| **2. Brainstorm** | `flow.plan` |
| **3. Worktree** | `flow.plan` |
| **4. Implement** | `flow.implement` |
| **5. Verify** | `flow.review`, `flow.verify`, `flow.pipeline-fix`, `flow.self-review` |
| **6. Preflight** | `flow.land` |
| **7. Land** | `flow.land` |
| **8. Clean up** | `flow.cleanup` |

`/flow-fast` never runs a visual verification. `flow.self-review` marks
section 5's in-run self-review pass. `flow.pipeline-fix` is never marked by hand: the store opens it
on its own when a `pipeline-fix-*` dispatch begins with no stage open, and closes it with that
dispatch's last sibling.

## Dynamic decisions

Every dispatch runs on its role's pair, per **Model and effort**
(`skills/flow-contracts/model-policy.md`).

Every dispatch's `description` and `Role:` line follow **Dispatch sites — the parent's closed list** (`skills/flow/implement.md`).

`<changeRoot>` is `<abs-worktree>/.superpowers/sdd/<name>/`, and `<project>/.superpowers/` is listed in
`<project>/.git/info/exclude`. Every `skills/flow/` section cited below runs as written, with
these substitutions and no others: `<changeRoot>` as above; `ff-<literal-token>` wherever it
reads `mf-<literal-token>`; the one worktree as the resolved set, its `<merge-base>` the sha
`git -C <worktree> rev-parse origin/<base>` prints right after section 3 creates it — or, once
section 4's base sync rebased the branch, the sha its last `REBASED:` line names —
standing wherever a cited section reads the working notes' merge base; and nothing written to
`<project>/spectre/` or a state file — `flow record render` is
skipped, the finding rows are the record. The scripts a cited section calls run as it says.

- **writing-plans**: write `<changeRoot>/tasks.md` in `/flow`'s plan shape — the task line and
  fields: `- [ ] <n>. <title>`, `**Files:**`, `**Tests:**`,
  `**Commit:**`, `**After:**` — one task per file or logical unit the change touches. A re-run replaces the file with
  the fix's own tasks.
  A minimal task in that shape — every field line at column 0, never indented under the task line,
  and `**Tests:**` naming the test in backticks or opening with `none — <reason>`:

```markdown
- [ ] 1. Reject an empty name in the parser
**Files:** `stats/internal/parse/name.go`, `stats/internal/parse/name_test.go`
**Tests:** `TestParseNameRejectsEmpty` in `stats/internal/parse/name_test.go`
**Commit:** fix(parse): reject an empty name
**After:** none
```

- **decide**: `flow-guard plan-check <worktree> <changeRoot> <abs-worktree>/.superpowers/sdd/decision.json -decide-only` (**Plan check**, `skills/flow/brainstorm.md`), `<changeRoot>` the
  directory holding this run's `tasks.md` and no `design.md`, writing `<abs-worktree>/.superpowers/sdd/decision.json` and printing the
  `## Decision` block.
- **sdd-tdd**: **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) as
  written except its task close: `flow-guard close-task` is not run — its tick writes a spectre plan, which flow-fast never writes. A group closes on the implementer's
  `flow record dispatch end` and `git -C <worktree> push origin <name>`, its gated per-task reviewer
  running at that close, then section 5's lint and tests.
- **review-panel**: `skills/flow/review-panel.md` as
  written, **The panel** through **Panel re-runs**.

The change summary in section 5 then also carries the `## Decision` block and each finding with
its status.

## 1. Kickoff

Resolve the Jira key and the change name per **Resolution (how `jiraIssue` is decided)** and
**Change naming** (`skills/flow-contracts/jira-integration.md`), exactly — including the slug
constraints on a summary-derived name. Then transition the issue to **In Progress** per
**Transitions** there. Then:

**A re-run is detected, never recorded**: `<project>-worktrees/<name>` already existing means an
earlier run left the branch unlanded. Reuse it and skip the creation below. With an argument,
it is fix instructions and sections 2–5 run again on the branch; bare, sections 2–5 have nothing
to do and mark through, and the run continues at section 6.

**On a creating run, create the worktree here, before anything is read** — from the main
checkout, with `<base>` the `--base` value when one was given, else the branch `origin/HEAD`
points at:

```bash
git fetch origin
git worktree add <project>-worktrees/<name> -b <name> origin/<base>
git -C <project>-worktrees/<name> config branch.<name>.flowBase <base>   # --base only
git -C <project>-worktrees/<name> push -u origin <name>
```

Then the project's declared `## worktree setup` (**Project configuration**,
`skills/flow-contracts/project-configuration.md`) runs once, from the worktree root, before
anything else touches the tree — the fenced command lines only, in order, in the foreground; the
key's exit 1 (absent) continues with nothing run, its exit 2 stops the run, and a command's
non-zero exit ends the run naming the command and its output — each only for a cause out of the
run's reach; any other is cured and the setup re-run (**Fewest operator actions**,
`skills/flow-contracts/pipeline.md`). A re-run reusing an existing
worktree runs nothing here — the creating run already ran it.

**From here on, on every run, `<base>` is what `flow-guard resolve-base-branch <worktree>` prints** — the
recorded base when there is one — so a re-run lands where the creating run cut.

Still nothing else: no database, no bucket. Every section from here on
reads and writes the worktree alone; the main checkout is never checked out, staged, committed or
written.

## 2. Brainstorm

```bash
flow stage begin -command '/flow-fast' -stage flow.plan -harness <harness> -session-token mf-<literal-token> <name>
```

Make every judgment call yourself and name it in the summary, each with its rejected
alternative and the one-line reason it lost, however small the call — size never exempts one
(**Decisions**, `skills/flow/brainstorm-planner.md`);
a `/flow-fast` run with `## handoff` `none`, or the key absent, is one command from the operator,
`/flow-fast <key>`, and asks nothing after it. Only with `## handoff` `required`, and only where two
readings would lead to materially different work, ask once, batched, through
**AskUserQuestion**; a question there whose options include a recommended one takes it under
the `## decisions: recommended` mode (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`), recorded the way that section states, and a
genuinely open one still asks. Pick the simplest implementation that meets the ask.

## 3. Worktree

The worktree already exists — section 1 created it. This section only marks through.

**Dynamic decisions**' `writing-plans` step writes the plan.

**Dynamic decisions**' `decide` step runs here.

```bash
flow stage end   -command '/flow-fast' -stage flow.plan -outcome completed <name>
```

## 4. Implement

```bash
flow stage begin -command '/flow-fast' -stage flow.implement -harness <harness> -session-token mf-<literal-token> <name>
```

The branch is rebased onto its base before task 1 runs, as the base-sync step of **1. Load
context and validate the plan** (`skills/flow/implement.md`) states, with no `flow state
add-worktree` call and `<name>` as the pushed branch.

The project's instruction file, `<project>/CLAUDE.md`, is already in this session's context; never
read it here. Read
`<project>/.flow/project.md`'s `## lint`, `## test`, `## handoff` and `## default landing route` sections, the
first two read with `flow-guard project-get <project> <key>`, `## handoff` with
`flow-guard project-get <project> handoff --enum required none` and `## default landing route` with
`flow-guard project-get <project> "default landing route" --enum "pull request" "merge and push" manual`
(exit 1 absent; exit 3 malformed — report its stderr line by name and resolve as absent): sections
5 and 7 take their commands from the first three, and `## handoff` says whether the run stops
between them.

On a re-run only:

Then:

Implement per **Dynamic decisions**' `sdd-tdd` step. Test first where a test can express the behaviour
(**superpowers:test-driven-development**); a defect gets a failing test before its fix. A
refactor of output-producing code follows the PIN BEFORE REFACTOR rule
(`skills/flow/implementer-brief.md`). On a fix run the fix's own condition is mutation-killed before
section 5 closes: after the fix commits, mutate each fix's own condition — flip the gate, drop the
guard, substitute the lagging value — confirm a named test fails, and restore — per the MUTATION
PROOF paragraph (`skills/flow/review-panel.md`) — naming the mutation and its killing test in the
fix commit's body; a mutant nothing kills is an unfinished fix, its killing test added under this
section's commit rule before the run lands. Commit
one logical unit at a time on the `<name>` branch, subject in Conventional Commits form, no
attribution trailer, and `git -C <worktree> push origin <name>` after each one (**Branch
backup**, `skills/flow-contracts/git-boundaries.md`). A branch rewritten by an auto-rebase pushes
with `--force-with-lease`, never a bare `git push`, and never `--force`. A backing-out commit — a reverted prototype,
a withdrawn task — carries the rejection reason in its body (**Revert and back-out commits**,
`skills/flow-contracts/git-boundaries.md`). The commit series is the change's whole
record; section 5's lint and tests are the only close a commit gets. Fix every lint hit the project's `## lint`
raises on the files you touched rather than suppressing it.

```bash
flow stage end   -command '/flow-fast' -stage flow.implement -outcome completed <name>
```

## 5. Verify

```bash
flow stage begin -command '/flow-fast' -stage flow.review -harness <harness> -session-token mf-<literal-token> <name>
```

**Dynamic decisions**' `review-panel` step runs here.

```bash
flow stage end   -command '/flow-fast' -stage flow.review -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.verify -harness <harness> -session-token mf-<literal-token> <name>
```

Run `<agents repo>/scripts/flow-guard run-verify-list <worktree> lint` — it runs every command of
`## lint`'s fence, never the fence lines `flow-guard project-get` prints with them — and the `## test` commands scoped to what the
change touched — the packages, modules or test files the diff names, never the whole list; the full suite,
`<agents repo>/scripts/flow-guard run-verify-list -full <worktree> test`, runs only when the operator asked for it, under the operator's own `FLOW_OPERATOR_FULL_SUITE=1` (**Inline verify**, `skills/flow/verify-and-handoff.md`). When the diff names nothing those commands can scope to, none of them
runs — the change summary below states that empty-scope reason explicitly, the stated reason
being what makes the skip auditable rather than lazy, never a reflexive run of the whole list in its
place. Then `<agents repo>/scripts/flow-guard run-verify-list <worktree> visual` runs the `## visual verification` `verify` row when the diff touches `ui paths` — the committed baselines a `lint` list never checks; a failure that is the change's intended look is recaptured with the section's `capture` row and committed under section 4's commit rule before the run lands. Where the section declares a `regression checkout`, the run first adds that repository's worktree on the change's branch, runs the section's `setup` command in it, and starts there the stack its `verify` row needs (the project's `## run`) — the one exception to the no-isolation rule above; the `visual` list never falls back to the main checkout, and exits 1 naming it while it holds no worktree on the branch. A failure is fixed and re-run under section 4's
commit rule; the run never lands red, a failing test it did not introduce included (**the sweep**,
`skills/flow-contracts/known-bugs.md`).

```bash
flow stage end   -command '/flow-fast' -stage flow.verify -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.self-review -harness <harness> -session-token mf-<literal-token> <name>
```

**Self-review runs here, in this run — this section is canonical for `/flow-fast`'s self-review.**
A `flow.self-review` stage already completed for `<name>` skips the pass: a fix re-run never
records its findings twice.
Assemble the bundle in `<sr-dir>/bundle.md` (**2. Read the bundle and run the five angles**, `skills/flow-self-review/SKILL.md`), never printed, never committed, as, in
order: the stdout of `flow self-review
bundle -change <name>` run from the worktree — of the sources `/flow`'s bundle judges a change
by, a `/flow-fast` run can never have tasks.md, design.md, narrative.md, visual-verification.md and live-verification.md (no spectre
artifacts), so those five report `skipped` by design, never silently;
the git-log source resolves from the main checkout the command resolves, as the change branch's
commits; the ledger and panel render from the store when the run wrote rows.

Over that bundle run steps 2–5 of `/flow-self-review` (`skills/flow-self-review/SKILL.md`),
canonical for the five angles, the fix, where the fixes land — on `<name>` itself
when `<worktree>` is `<agents repo>`'s — the filing and the store rows. No report is written: the findings live in flowd.

The fixes land with the change on every route. **7. Land**'s landing push carries the fix
commits' shas per item 2 of step 3 of `/flow-self-review`, in whichever invocation makes it.

A product defect the pass met stops it at the end of step 2 of `/flow-self-review`, closes the mark below
`-outcome stopped` instead, and continues straight into the fix re-run **the sweep**
(`skills/flow-contracts/known-bugs.md`) names, in the same invocation.

```bash
flow stage end   -command '/flow-fast' -stage flow.self-review -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.verify -harness <harness> -session-token mf-<literal-token> <name>
```

Print the change summary: what changed and why, grouped by area, one or two sentences each;
what was verified and how; anything deliberately left out; the self-review's fixed shas when
section 5 landed any. This is the one report the run prints.

**Then the handoff, decided by `## handoff`** (**Project configuration**,
`skills/flow-contracts/project-configuration.md`): `none`, or the key absent, continues to
section 6 in this same invocation; `required` ends the run here — worktree and branch kept, the
summary naming the branch and `Re-run /flow-fast <name> to land it` — so the operator, who asked
to test by hand, reviews the branch before anything is pushed. The remaining marks below are then
made by that bare re-run, not by this one.

```bash
flow stage end   -command '/flow-fast' -stage flow.verify -outcome completed <name>
```

## 6. Preflight

```bash
flow stage begin -command '/flow-fast' -stage flow.land -harness <harness> -session-token mf-<literal-token> <name>
```

`git -C <worktree> status --porcelain` must be empty: every edit is committed. Anything the ask
covered and this run did not finish is named in the summary above, never silently dropped.

The route is what section 4's `## default landing route` `--enum` call resolved, taken without
asking. When it resolved none (exit 1 or 3): with `## handoff` `required`, ask once through
**AskUserQuestion** with the options `pull request`, `merge and push` and `manual`, `pull request`
recommended; otherwise take `pull request` without asking.

## 7. Land

First, on every route, bring the branch up to date so what lands is what was verified:

```bash
git -C <worktree> fetch origin
git -C <worktree> rebase origin/<base>
```

A rebase that conflicts is resolved in place, automatically, per the **Conflict** bullet of **Sync
the branch onto the base** (`skills/flow/sync-onto-base.md`) — its resolution
rule, its stop-and-ask cases and its handoff line apply as written. A rebase that moved the
branch re-runs section 5's lint and targeted tests before continuing; one that needed resolution
runs the project's whole `## lint` and `## test` lists instead.

- **merge and push**: `git -C <worktree> push origin <name>:<base>`. A push the
  remote rejects (branch protection, a non-fast-forward) falls back to **pull request** below and
  says so. This route is the one place `/flow-fast` pushes to the base branch.
- **pull request**: `git -C <worktree> push --force-with-lease origin <name>`, then `gh pr create --base
  <base> --head <name>` with the summary from section 5 as the body.
- **manual**: `git -C <worktree> push --force-with-lease origin <name>`; print the branch name and the worktree path.

```bash
flow stage end   -command '/flow-fast' -stage flow.land -outcome completed <name>
```

**`pull request` and `manual` stop here**, worktree and branch kept, and the run ends by naming the
PR or the branch. When the PR is merged, or the manual landing done, re-run `/flow-fast <name>`
bare: section 1 finds the worktree, sections 2–7 have nothing new to do and mark through, and
section 8 runs.

## 8. Clean up

`merge and push` continues here in the same invocation.

```bash
flow stage begin -command '/flow-fast' -stage flow.cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

```bash
git -C <worktree> fetch origin
git -C <worktree> merge-base --is-ancestor <name> origin/<base>
```

A non-zero exit means the branch is not on the base branch yet — stop and say so, removing
nothing.

Save the plan and the self-review bundle to flowd first — the worktree holds their only copies;
a non-zero exit stops the run, removing nothing:

```bash
flow change save -C <worktree> <name>
<agents repo>/scripts/flow-guard stop-worktree-processes <project>-worktrees/<name>
git -C <project> worktree remove <project>-worktrees/<name>
git -C <project> branch -D <name>
git -C <project> push origin --delete <name>
```

Then, only when the main checkout is on `<base>` with an empty `git status
--porcelain`, `git -C <project> pull --ff-only`; otherwise leave it and report one line naming
why.

```bash
flow stage end   -command '/flow-fast' -stage flow.cleanup -outcome completed <name>
```

Transition the Jira issue to **Done**, the same way as before. End by naming the landed commit
on `<base>`. The final report carries no open item, per the Finished handoff rule in
`skills/flow/cleanup.md`.
