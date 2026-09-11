---
name: flow-fast
description: Minimal-ceremony /flow variant — one invocation from Jira key to pushed change. A git worktree for isolation and nothing else, inline implementation, project lint plus targeted tests, then the project's default landing route and cleanup. Marks every flow.* stage /flow marks and keeps the Jira transitions; no spectre artifacts, no state file, no decision record, no review panel, no guards. Use for /flow-fast.
allowed-tools: Bash(flow:*)
license: MIT
---

Do the work the way a careful engineer does it by hand — read, edit, verify, commit, land — and
record it the way `/flow` does: every `flow.*` stage mark below, in this order, under one
session token, so the stats views see a `/flow-fast` run as the same pipeline. Nothing else of
`/flow` survives here. There is no spectre change, no `proposal.md`/`design.md`/`tasks.md`, no
state file and no three states, no decision record, no dispatch of any kind, no review panel, no
staged-diff gate, no archive branch, no `<project>/docs/superpowers/` record, and no guard script. The only
isolation is git's: a worktree on its own branch. `prepare-workspace.sh`, the per-change
database, bucket, ports and cache index of **Workspace isolation**
(`skills/flow-contracts/workspace-isolation.md`) are never set up.

**Announce at start:** "Using flow-fast for change `<name>`."

**No flags.** The only argument is the change description or Jira key on a creating run, or fix
instructions on a re-run; report anything else rather than ignoring it.

**Every mark is a literal call.** `<harness>` is this harness's name, `ff-<literal-token>` is one
token generated once at the start of the run and typed the same at every `stage begin` — never a
shell substitution, per the CLI's own usage text. Two adjacent lines with nothing between them
mark a stage `/flow-fast` has nothing to run for; the mark stays so the run's stage set matches
`/flow`'s.

**Guardrails, the whole list.** Never dispatch a subagent. Never ask a model, planning-effort or
review question. Never write `<project>/spectre/`, `<project>/docs/superpowers/` or a state file.
Never set up workspace isolation and never call a guard script. Never push to a branch other than
the one the landing route names.

## 1. Kickoff

Resolve the Jira key and the change name per **Resolution (how `jiraIssue` is decided)** and
**Change naming** (`skills/flow-contracts/jira-integration.md`), exactly — including the slug
constraints on a summary-derived name. Then transition the issue to **In Progress** per
**Transitions** there (by name, forward-only, one line on failure per **Never blocking**). Then:

```bash
flow stage begin -command '/flow-fast' -stage flow.kickoff -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.kickoff -outcome completed <name>
```

**A re-run is detected, never recorded**: `<project>/.worktrees/<name>` already existing means an
earlier run left the branch unlanded. Reuse it and skip section 3's creation. With an argument,
it is fix instructions and sections 2–5 run again on the branch; bare, sections 2–5 have nothing
to do and mark through, and the run continues at section 6.

## 2. Brainstorm

```bash
flow stage begin -command '/flow-fast' -stage flow.brainstorm -harness <harness> -session-token ff-<literal-token> <name>
```

Read the ask and the code it touches until the change is clear — every file the change has to
touch, the actual flow end to end. Make every judgment call yourself and name it in the summary;
a `/flow-fast` run with `## handoff` `none` is one command from the operator, `/flow-fast
<key>`, and asks nothing after it. Only with `## handoff` `required`, and only where two
readings would lead to materially different work, ask once, batched, through
**AskUserQuestion**. Pick the simplest implementation that meets the ask.

```bash
flow stage end   -command '/flow-fast' -stage flow.brainstorm -outcome completed <name>
```

## 3. Worktree

```bash
flow stage begin -command '/flow-fast' -stage flow.create-artifacts -harness <harness> -session-token ff-<literal-token> <name>
```

On a creating run, from the main checkout, with `<default-branch>` the branch `origin/HEAD` points
at:

```bash
git fetch origin
grep -qx '.worktrees/' .git/info/exclude 2>/dev/null || echo '.worktrees/' >> .git/info/exclude
git worktree add <project>/.worktrees/<name> -b <name> origin/<default-branch>
```

Nothing else: no `## worktree setup` command, no database, no bucket. A project whose build needs
generated files or installed dependencies gets them the moment section 5's first test run asks
for them, in the worktree, by the project's own ordinary commands.

```bash
flow stage end   -command '/flow-fast' -stage flow.create-artifacts -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.writing-plans -harness <harness> -session-token ff-<literal-token> <name>
```

Register the steps of this change with the harness's task-list mechanism — one entry per file or
logical unit you will touch, so the operator can follow along. Nothing is written to disk.

```bash
flow stage end   -command '/flow-fast' -stage flow.writing-plans -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.decide -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.decide -outcome completed <name>
```

## 4. Implement

```bash
flow stage begin -command '/flow-fast' -stage flow.load-context -harness <harness> -session-token ff-<literal-token> <name>
```

Read `<project>/CLAUDE.md`, `<project>/AGENTS.md` where present, and `<project>/.flow/project.md`'s
`## lint`, `## test`, `## handoff` and `## default landing route` sections, each read with
`project-get.sh <project> <key>`: sections 5 and 7 take their commands from the first three,
and `## handoff` says whether the run stops between them.

```bash
flow stage end   -command '/flow-fast' -stage flow.load-context -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.isolate-workspace -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.isolate-workspace -outcome completed <name>
```

On a re-run only:

```bash
flow stage begin -command '/flow-fast' -stage flow.document-fix -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.document-fix -outcome completed <name>
```

Then:

```bash
flow stage begin -command '/flow-fast' -stage flow.sdd-tdd -harness <harness> -session-token ff-<literal-token> <name>
```

Implement in the worktree, in this session. Test first where a test can express the behaviour
(**superpowers:test-driven-development**); a defect gets a failing test before its fix. Commit
one logical unit at a time on the `<name>` branch, subject in Conventional Commits form with the scope
naming the module the commit moved (`~/.claude/rules/commit-scope-is-the-module.md`), no
attribution trailer. Fix every lint hit the project's `## lint` raises on the files you touched
rather than suppressing it.

```bash
flow stage end   -command '/flow-fast' -stage flow.sdd-tdd -outcome completed <name>
```

## 5. Verify

```bash
flow stage begin -command '/flow-fast' -stage flow.review-panel -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.review-panel -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.verify -harness <harness> -session-token ff-<literal-token> <name>
```

Run every command in `## lint`, in the worktree, and the `## test` commands scoped to what the
change touched — the packages, modules or test files the diff names, never the full suite unless
the operator asked for it. A failure is fixed and re-run under section 4's commit rule; the run
never lands red.

```bash
flow stage end   -command '/flow-fast' -stage flow.verify -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.stage-diff -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.stage-diff -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.run-instructions -harness <harness> -session-token ff-<literal-token> <name>
```

Print the change summary: what changed and why, grouped by area, one or two sentences each;
what was verified and how; anything deliberately left out. This is the one report the run
prints.

**Then the handoff, decided by `## handoff`** (**Project configuration**,
`skills/flow-contracts/project-configuration.md`): `none` continues to section 6 in this same
invocation; `required`, or the key absent, ends the run here — worktree and branch kept, the
summary naming the branch and `Re-run /flow-fast <name> to land it` — so the operator reviews
the branch before anything is pushed. The remaining marks below are then made by that bare
re-run, not by this one.

```bash
flow stage end   -command '/flow-fast' -stage flow.run-instructions -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.write-in-progress -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.write-in-progress -outcome completed <name>
```

## 6. Preflight

```bash
flow stage begin -command '/flow-fast' -stage flow.preflight -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.preflight -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.unfinished-work-gate -harness <harness> -session-token ff-<literal-token> <name>
```

`git -C <worktree> status --porcelain` must be empty: every edit is committed. Anything the ask
covered and this run did not finish is named in the summary above, never silently dropped.

```bash
flow stage end   -command '/flow-fast' -stage flow.unfinished-work-gate -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.landing-question -harness <harness> -session-token ff-<literal-token> <name>
```

The route is `## default landing route` when the project declares one (`merge and push`,
`open PR` or `manual`), read without asking. When the project declares none: with `## handoff`
`required`, ask once through **AskUserQuestion** with those three options, `open PR`
recommended; with `## handoff` `none`, take `open PR` without asking.

```bash
flow stage end   -command '/flow-fast' -stage flow.landing-question -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.preserve-sessions -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.preserve-sessions -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.commit-two -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.commit-two -outcome completed <name>
```

## 7. Land

```bash
flow stage begin -command '/flow-fast' -stage flow.landing-routes -harness <harness> -session-token ff-<literal-token> <name>
```

First, on every route, bring the branch up to date so what lands is what was verified:

```bash
git -C <worktree> fetch origin
git -C <worktree> rebase origin/<default-branch>
```

A rebase that conflicts stops the run here, naming the conflicting files — resolve it and re-run.
A rebase that moved the branch re-runs section 5's lint and targeted tests before continuing.

- **merge and push**: `git -C <worktree> push origin <name>:<default-branch>`. A push the
  remote rejects (branch protection, a non-fast-forward) falls back to **open PR** below and
  says so. This route is the one place `/flow-fast` pushes to the default branch; a project whose
  default branch is protected declares `open PR` instead
  (`~/.claude/rules/no-direct-pushes-to-main.md`).
- **open PR**: `git -C <worktree> push -u origin <name>`, then `gh pr create --base
  <default-branch> --head <name>` with the summary from section 5 as the body.
- **manual**: push nothing; print the branch name and the worktree path.

Then transition the Jira issue to In Review per **Transitions**
(`skills/flow-contracts/jira-integration.md`) — on every route that completed, never on one that stopped.

```bash
flow stage end   -command '/flow-fast' -stage flow.landing-routes -outcome completed <name>
```

**`open PR` and `manual` stop here**, worktree and branch kept, and the run ends by naming the
PR or the branch. When the PR is merged, or the manual landing done, re-run `/flow-fast <name>`
bare: section 1 finds the worktree, sections 2–7 have nothing new to do and mark through, and
section 8 runs.

## 8. Clean up

`merge and push` continues here in the same invocation.

```bash
flow stage begin -command '/flow-fast' -stage flow.verify-merge -harness <harness> -session-token ff-<literal-token> <name>
```

```bash
git -C <worktree> fetch origin
git -C <worktree> merge-base --is-ancestor <name> origin/<default-branch>
```

A non-zero exit means the branch is not on the default branch yet — stop and say so, removing
nothing.

```bash
flow stage end   -command '/flow-fast' -stage flow.verify-merge -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.sync-archive -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.sync-archive -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.commit-archive -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.commit-archive -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.cleanup -harness <harness> -session-token ff-<literal-token> <name>
```

```bash
git -C <project> worktree remove <project>/.worktrees/<name>
git -C <project> branch -D <name>
git -C <project> push origin --delete <name>   # only when the branch was pushed
```

Then, only when the main checkout is on `<default-branch>` with an empty `git status
--porcelain`, `git -C <project> pull --ff-only`; otherwise leave it and report one line naming
why.

```bash
flow stage end   -command '/flow-fast' -stage flow.cleanup -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.write-finished -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.write-finished -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.push-archive -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.push-archive -outcome completed <name>
```

Transition the Jira issue to **Done**, the same way as before. End by naming the landed commit
on `<default-branch>`.
