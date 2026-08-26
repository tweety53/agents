# Integrate (run 1)

Loaded by `skills/flow/SKILL.md` on a **bare** invocation at `IN_PROGRESS` — no argument. Run 1 of
what was `/myflow-finish`, unchanged in procedure except two folds design.md decides:
`move-in-review-fold` (the Jira "move to In Review" step becomes a sub-step of `flow.landing-routes`
rather than its own mark) and this task's own resolution of open question `write-in-progress-fold`
(below).

## Deciding which run this is

**`skills/myflow-contracts/finish-contract.md` is canonical for every procedure below** — the
base-branch resolution, the preflight checks, the removal sequence and their rationales live in
that file, unchanged by this rework. This file carries only what is specific to *executing* it
under `/flow`.

**Load `skills/myflow-contracts/worktree-resolution.md`** too.

**Check guard presence** per **Guard presence check** (`skills/myflow-contracts/pipeline.md`),
already run at the top of this invocation.

Run `check-finish-preflight.sh` once per worktree in the set found by **Resolving a change's
worktrees** (`skills/myflow-contracts/finish-contract.md`) — never a raw read of the state file's
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
myflow stage begin -command '/flow' -stage flow.preflight -harness <harness> -session-token mf-<literal-token> <name>
myflow stage end   -command '/flow' -stage flow.preflight -outcome completed <name>
myflow stage begin -command '/flow' -stage flow.unfinished-work-gate -harness <harness> -session-token mf-<literal-token> <name>
```

## 1. Check for unfinished work

Run `check-unfinished-work.sh <worktree> <name>` once per worktree in the resolved set — before the
landing question and before any git action.

- **`CLEAR:` from every worktree** → continue to **2** with no extra prompt.
- **A resolved set that comes back empty** → stop and ask the operator.
- **`OUTSTANDING:`** → show the breakdown, and offer exactly three courses, shape per Operator
  prompts (`skills/myflow-contracts/operator-prompts.md`):

  > **This change carries unfinished work — how should integration proceed?**
  > - **Stop — I'll finish it first** *(recommended)*
  > - **Continue — integrate anyway**
  > - **File or join a Jira follow-up, then continue**

  There is no fourth, and in particular none that hands back to `skills/flow/implement.md` inline.
- **No verdict line at all, and a non-zero exit** → stop and ask.

**Stop** exits leaving the change at `IN_PROGRESS` with nothing staged, committed or pushed.
**Continue** carries the outstanding list into **3**'s planning commit and into the handoff. **File
or join a Jira follow-up** puts the outstanding items on a follow-up issue and continues. See
**Follow-up issues** (`skills/myflow-contracts/jira-followups.md`) for the search, the
confirmation, and how it is labelled; a filing that fails is one skipped-with-reason line and the
run still continues.

```bash
myflow stage end -command '/flow' -stage flow.unfinished-work-gate -outcome completed <name>
```

(`completed` on **Continue** or **File or join…**, `stopped` on **Stop**.)

## 2. Ask how the branch should land

```bash
myflow stage begin -command '/flow' -stage flow.landing-question -harness <harness> -session-token mf-<literal-token> <name>
```

> **How should this branch land?**
> - **Open a pull request** *(default, recommended)*
> - **Merge and push**
> - **Handle it manually**

Having asked once, run to completion without asking again.

**Report an existing PR before asking.** If a PR exists for this branch — from `prUrl`, or a PR CLI
when one is usable — say so, open or closed-unmerged, before the operator answers.

```bash
myflow stage end -command '/flow' -stage flow.landing-question -outcome completed <name>
```

## 3. Commit the staged work

**Load `skills/myflow-contracts/git-boundaries.md`** before either commit below.

```bash
myflow stage begin -command '/flow' -stage flow.preserve-sessions -harness <harness> -session-token mf-<literal-token> <name>
```

**Before any route commits, reshape the branch.** Run `git -C <worktree> reset --soft
<recorded-merge-base>`, where `<recorded-merge-base>` is the merge base recorded in the state
file's `worktrees` map for this worktree. This collapses every per-task and fixup commit back into
the working tree, uncommitted.

All three routes commit — implementation, the `<project>/spectre/changes/` planning artifacts, and
the session records under `<project>/docs/superpowers/` — as **two** commits, never one.

**Load `skills/myflow-contracts/session-records.md`** before rendering, below.

**Render the ledger first**, before staging:

```bash
myflow record render -change <name> -kind ledger -repo <worktree>
```

**A change with no dispatch rows reports `MISSING: ledger` and exits 0** — never a failure. **A
non-zero exit means a destination was refused or could not be written:** report it and continue.

```bash
myflow stage end   -command '/flow' -stage flow.preserve-sessions -outcome completed <name>
myflow stage begin -command '/flow' -stage flow.commit-two -harness <harness> -session-token mf-<literal-token> <name>
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
(`skills/myflow-contracts/git-boundaries.md`).

**Implementation first, planning artifacts second.** The second commit's message lists anything the
operator chose to integrate over at **1**. The state file is **not** committed.

```bash
myflow stage end -command '/flow' -stage flow.commit-two -outcome completed <name>
```

## 4. Take the chosen route, write the state, and transition Jira

```bash
myflow stage begin -command '/flow' -stage flow.landing-routes -harness <harness> -session-token mf-<literal-token> <name>
```

This stage carries three sub-steps under one mark, per this task's own resolution of open question
`write-in-progress-fold` and design.md's `move-in-review-fold`: the git route, the state write, and
the Jira transition — three sub-steps that used to be three separate top-level marks
(`finish.landing-routes`, `finish.write-in-progress`, `finish.move-in-review`) now recorded as one.
**This task's own choice**: fold `write-in-progress` in alongside `move-in-review`, rather than
leave it standalone. The write is a genuine no-op (`IN_PROGRESS` → `IN_PROGRESS`, nothing changes
but `prUrl` and `updatedAt`/`updatedBy`) exactly as design.md's open question describes, and by the
time this mark reaches it, the route sub-step immediately above has already decided what `prUrl`
becomes — folding the write in with the route that produces its one real input, and with the Jira
transition that only ever follows a successful route, keeps one mark's three sub-steps in the causal
order they already have to run in, rather than three marks whose middle one records nothing a reader
could not already infer from the other two.

Per **Finish contract** (`skills/myflow-contracts/finish-contract.md`) → run 1. Push with `-u` so
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

**Sub-step: transition the issue to In Review**, whichever route was taken — pull request, merge and
push, or manual — per **Transitions** (`skills/myflow-contracts/jira-integration.md`): after the
state write, never before, never blocking. A run that stopped on a failed push does **not**
transition.

```bash
myflow stage end -command '/flow' -stage flow.landing-routes -outcome completed <name>
```

## No verification gate

**Run no tests, no linters, and no spec-coverage check** — see **Finish contract**
(`skills/myflow-contracts/finish-contract.md`). Correctness was established during
`skills/flow/review-panel.md` and by the human gate.

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
(`skills/myflow-contracts/handoff-blocks.md`) rather than assuming.

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
