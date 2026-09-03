# Integrate (run 1)

Loaded by `skills/flow/SKILL.md` on a **bare** invocation at `IN_PROGRESS` — no argument.

## Deciding which run this is

**`skills/flow-contracts/finish-contract-run1.md` is canonical for every procedure below** — the
base-branch resolution, the preflight checks, the removal sequence and their rationales live in
that file, unchanged by this rework. This file carries only what is specific to *executing* it
under `/flow`.

**Load `skills/flow-contracts/worktree-resolution.md`** too.

**Check guard presence** per **Guard presence check** (`skills/flow-contracts/pipeline.md`),
already run at the top of this invocation.

Run `check-finish-preflight.sh` once per worktree in the set found by **Resolving a change's
worktrees** (`skills/flow-contracts/finish-contract-run1.md`) — never a raw read of the state file's
`worktrees` map. Its `<base-ref>` argument is `origin/$BASE`, `$BASE` being what
`resolve-base-branch.sh` prints for that worktree.

- **`RUN1`** → this file (integrate)
- **`RUN2`** from every worktree → `skills/flow/archive.md` <!-- refs-guard:allow -->
- **`REFUSE`** → stop, report what the script reported, and ask the operator
- **A resolved set that comes back empty** → stop and ask, exactly as `REFUSE`
- **No verdict line at all, and exit 2** → treat exactly as `REFUSE`

**On a `RUN1` verdict**, mark `flow.preflight` (closed immediately, since the verdict is already in
hand) then `flow.unfinished-work-gate`:

```bash
flow stage begin -command '/flow' -stage flow.preflight -harness <harness> -session-token mf-<literal-token> <name>
flow stage end   -command '/flow' -stage flow.preflight -outcome completed <name>
flow stage begin -command '/flow' -stage flow.unfinished-work-gate -harness <harness> -session-token mf-<literal-token> <name>
```

## 1. Check for unfinished work

Run `check-unfinished-work.sh <worktree> <name> <canonical-worktree>` once per worktree in the
resolved set — before the landing question and before any git action. `<canonical-worktree>` is the
one member of the resolved set whose own `<project>/<spec-root>/changes/<name>/tasks.md` exists — the same
worktree for every call in this run, so a satellite worktree's call still resolves its plan through
the link instead of reporting an absence.

- **`CLEAR:` from every worktree** → continue to **2** with no extra prompt.
- **A resolved set that comes back empty** → stop and ask the operator.
- **`OUTSTANDING:`** → show the breakdown, and offer exactly three courses, shape per Operator
  prompts (`skills/flow-contracts/operator-prompts.md`):

  > **This change carries unfinished work — how should integration proceed?**
  > - **Stop — I'll finish it first** *(recommended)*
  > - **Continue — integrate anyway**
  > - **File or join a Jira follow-up, then continue**

  There is no fourth, and in particular none that hands back to `skills/flow/implement.md` inline.
- **No verdict line at all, and a non-zero exit** → stop and ask.

**Stop** exits leaving the change at `IN_PROGRESS` with nothing staged, committed or pushed.
**Continue** carries the outstanding list into **3**'s planning commit and into the handoff. **File
or join a Jira follow-up** puts the outstanding items on a follow-up issue and continues. See
**Follow-up issues** (`skills/flow-contracts/jira-followups.md`) for the search, the
confirmation, and how it is labelled; a filing that fails is one skipped-with-reason line and the
run still continues.

```bash
flow stage end -command '/flow' -stage flow.unfinished-work-gate -outcome completed <name>
```

(`completed` on **Continue** or **File or join…**, `stopped` on **Stop**.)

## 2. Ask how the branch should land

```bash
flow stage begin -command '/flow' -stage flow.landing-question -harness <harness> -session-token mf-<literal-token> <name>
```

**Check whether the base branch has moved first.** Run `check-base-moved.sh` once per worktree in
the resolved set, report every verdict, and ask only on overlap, per **Finish contract**
(`skills/flow-contracts/finish-contract-run1.md`). On an overlap from any worktree, one aggregated
prompt, shape per Operator prompts (`skills/flow-contracts/operator-prompts.md`):

> **The base branch has moved and touches paths this change also touched — how should
> integration proceed?**
> - **Stop — I'll rebase or reorder first** *(recommended)*
> - **Rebase onto `<base>` now, then continue**
> - **Continue — land anyway**

**Stop** exits leaving the change at `IN_PROGRESS` with nothing staged, committed or pushed, and
closes the mark:

```bash
flow stage end -command '/flow' -stage flow.landing-question -outcome stopped <name>
```

**Continue** carries the reported movement into the handoff and proceeds to the landing question
below. No overlap anywhere → report the counts and go straight to the landing question, with no
extra prompt.

**Rebase** runs `git -C <worktree> rebase origin/$BASE` once per worktree in the resolved set whose
own `check-base-moved.sh` verdict was `MOVED` — never a worktree whose verdict was `CLEAR`, even
though the prompt above is asked once for the whole change per **ask-only-on-overlap**.

- **Clean** (exit 0): the merge base carried forward for the rest of **this run** becomes
  `origin/$BASE`'s resolved tip at rebase time — call it `<rebased-merge-base>` below.
  **This is a this-run-only value, never written to the state file**: it supersedes, for every
  remaining step of this run, every place below that would otherwise read the state file's
  recorded, pre-rebase merge base for this worktree — most concretely **3. Commit the staged
  work**'s reshape, whose own `<recorded-merge-base>` means `<rebased-merge-base>` for a worktree
  this step rebased, and the state file's original recorded value for every other worktree. Re-run
  `check-base-moved.sh` once more against `<rebased-merge-base>`; a fresh `MOVED` overlap re-offers
  this same three-option prompt rather than looping silently. Otherwise, run **Scoped
  re-verification** below, then proceed to the landing question. If this change's verification
  compares against a recorded baseline, recapture it now — a proof taken against the pre-rebase
  base is void.
- **Conflict** (non-zero exit): never auto-abort — and never
  attempt to resolve the conflict yourself, by editing the conflicting files or otherwise. Leave the
  worktree mid-rebase exactly as `git rebase` left it, report the conflicting file(s) from
  `git status`, and hand off `git -C <worktree> rebase --continue` (after the **operator** resolves
  it) or `git -C <worktree> rebase --abort` as the operator's next manual step. State stays
  `IN_PROGRESS`; this run stops here, exactly as **Stop** already does, and closes the mark
  `stopped`.

**Scoped re-verification**: for each path
`check-base-moved.sh` reported under `overlaps:`, look for a discoverable guard test —
`<agents repo>/scripts/test-<basename-without-ext>.sh` beside `<agents repo>/scripts/<name>.sh`,
the same naming `<agents repo>/scripts/run-guard-tests.sh` already discovers by glob — and run it
if found. A path with none is
stated in the handoff as having no verification to run, not silently skipped. A non-zero exit from
any discovered test blocks this stage exactly like any other verify-stage failure: report it, leave
the change `IN_PROGRESS`, stop before the landing question, and close the mark `stopped`. **Never**
re-run the project's whole `## lint`/`## test` list here. A clean rebase whose overlap set clears
this stage proceeds to the landing question and closes the mark `completed`, exactly like
**Continue**.

Run `project-get.sh <main-checkout> "default landing route"` (exit 1: absent), and resolve it
against the three literals `pull request` / `merge and push` /
`manual`, byte-for-byte after trimming leading/trailing whitespace. A body matching none of
them exactly is reported by name and dropped, resolving as absent.

**A resolved default skips the question entirely** — take that route without asking, and say so
in the handoff (`Route: <route> — from this project's configured default, not asked`). Only an
absent or unresolved default falls back to asking:

> **How should this branch land?**
> - **Open a pull request** *(default, recommended)*
> - **Merge and push**
> - **Handle it manually**

Having asked once, run to completion without asking again.

**Report an existing PR before asking.** If a PR exists for this branch — from `prUrl`, or a PR CLI
when one is usable — say so, open or closed-unmerged, before the operator answers.

```bash
flow stage end -command '/flow' -stage flow.landing-question -outcome completed <name>
```

## 3. Commit the staged work

**Load `skills/flow-contracts/git-boundaries.md`** before either commit below.

```bash
flow stage begin -command '/flow' -stage flow.preserve-sessions -harness <harness> -session-token mf-<literal-token> <name>
```

**Before any route commits, reshape the branch.** Run `git -C <worktree> reset --soft
<recorded-merge-base>`, where `<recorded-merge-base>` is the merge base recorded in the state
file's `worktrees` map for this worktree — **or `<rebased-merge-base>` from step 2 above, for a
worktree this run rebased**, never the state file's now-stale pre-rebase value for that worktree.
This collapses every per-task and fixup commit back into the working tree, uncommitted; using the
stale value here would also collapse in the upstream commits the rebase just brought in, silently
smuggling them into the implementation commit below.

All three routes commit — implementation, the `<project>/spectre/changes/` planning artifacts, and
the session records under `<project>/docs/superpowers/` — as **two** commits, never one.

**Load `skills/flow-contracts/session-records.md`** before rendering, below.

**Render the ledger first**, before staging:

```bash
flow record render -change <name> -kind ledger -repo <worktree>
```

**A change with no dispatch rows reports `MISSING: ledger` and exits 0** — never a failure. **A
non-zero exit means a destination was refused or could not be written:** report it and continue.

```bash
flow stage end   -command '/flow' -stage flow.preserve-sessions -outcome completed <name>
flow stage begin -command '/flow' -stage flow.commit-two -harness <harness> -session-token mf-<literal-token> <name>
```

Then stage and commit twice, in this order:

```bash
commit-split.sh <worktree> <name> \
  "<type>(<module>): <what the implementation does>" \
  "chore(spectre): plan and session records"
```

`<type>`, `<module>` and `<what the implementation does>` are derived from the reshaped diff. The
planning message is a **fixed literal**.

**Run that as one command.** The guards, the skipped-empty rule, the stop-on-failure rule and the
symlinked-planning-path case are all under **Git boundaries**
(`skills/flow-contracts/git-boundaries.md`).

**Implementation first, planning artifacts second.** The second commit's message lists anything the
operator chose to integrate over at **1**. The state file is **not** committed.

```bash
flow stage end -command '/flow' -stage flow.commit-two -outcome completed <name>
```

## 4. Take the chosen route, write the state, and transition Jira

```bash
flow stage begin -command '/flow' -stage flow.landing-routes -harness <harness> -session-token mf-<literal-token> <name>
```

This stage carries three sub-steps under one mark: the git route, the state write, and the Jira transition.

Per **Finish contract** (`skills/flow-contracts/finish-contract-run1.md`) → run 1. Push with `-u` so
the branch has an upstream.

**Every git step here can fail, and none of them may fail silently.** A rejected push, a merge
conflict, a commit blocked by a hook, or `gh pr create` erroring must be **reported with the
command's own output**, and the run must **stop** leaving the change at `IN_PROGRESS`.

If there is no remote at all, say exactly that ("this repository has no remote, so there is nothing
to push to or merge into"), not that the base branch failed to resolve.

**Human confirmation is a legitimate substitute for an API probe** on a forge with no usable CLI. If
the answer is No, leave `prUrl` null and say what to do next.

**Sub-step: write the state.** Write `state` unchanged at `IN_PROGRESS`, `prUrl` set if a PR was
opened, and every other field carried forward.

**Load `skills/flow-contracts/jira-integration.md`** — it is canonical for the transition below.

**Sub-step: transition the issue to In Review**, whichever route was taken — pull request, merge and
push, or manual — per **Transitions** (`skills/flow-contracts/jira-integration.md`): after the
state write, never before, never blocking. A run that stopped on a failed push does **not**
transition.

```bash
flow stage end -command '/flow' -stage flow.landing-routes -outcome completed <name>
```

## No verification gate

**Run no tests, no linters, and no spec-coverage check** — see **Finish contract**
(`skills/flow-contracts/finish-contract-run1.md`). Correctness was established during
`skills/flow/review-panel.md` and by the human gate. **One exception:** the scoped
re-verification in step 2 above, triggered only by a rebase this stage itself performed, never a
general re-opening of this rule.

## Handoff

```
## Branch integrated — waiting on the merge | merged and waiting on run 2

**Change:** <name>
**Route:** pull request | merged and pushed | manual
**PR:** <prUrl> | none — merged directly | none — you are handling it
**Outstanding:** <what step 1 reported and the operator integrated over> | none
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)

<what the operator must do before the next run>

Next:
/flow <name>
```

| Route taken in step 4 | Heading |
|--------------------|---------|
| pull request | *waiting on the merge* |
| **merge and push** | *merged and waiting on run 2* — continue below |
| manual | *waiting on the merge* |

Where the route is not certain — a run resumed after a partial failure — take the answer from the
merge-status test in **The block each state renders**
(`skills/flow-contracts/handoff-blocks.md`) rather than assuming.

## After merge-and-push specifically

Continue, within the same invocation and without a further command from the operator, into
`skills/flow/archive.md` exactly as written. Nothing external blocks this route.

## After open PR or manual specifically

Stop after the route completes, printing the handoff above. Each of these two routes needs an
action outside this command's control before archiving can happen. The next bare `/flow <name>`
call, once the branch is integrated, runs the archive phase.

## Guardrails

- **Never** ask how the branch should land before the unfinished-work gate has been answered, and
  never run a git command before it either.
- **Never** mix the implementation and the planning artifacts in one commit.
- **Never** archive a change whose branch has not reached the base branch.
- **Never** run tests, linters, or a coverage check.
- **Never** hardcode `main` or `develop`, and **never** resolve the base branch from `HEAD`'s
  upstream.
- **Never** let a git failure pass silently.
- **Never** stage past a symlinked planning path with a bare `git add -A`.
- **Never** let a Jira call block the archive — one skipped-with-reason line.
