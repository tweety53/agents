# Clean up (run 2)

Loaded either by `skills/flow/integrate.md`'s merge-and-push route, in the same invocation, or by a
fresh bare `/flow <name>` invocation once `flow-guard check-finish-preflight` returns `RUN2` from every
worktree.

**`skills/flow-contracts/finish-contract-run2.md` is canonical for the full procedure.**
`flow-guard integrate-change` runs it — chained from `land` on the merge-and-push route, or called once
`prepare`'s `NEXT:` names it — with its `flow.cleanup` stage mark; the parent handles each stop, makes the
Jira transition, and prints the handoff. **Load `skills/flow-contracts/finish-contract-run2.md`
only when** a stop below cites one of its sections.

**A self-hosting change refreshes the tooling before cleanup.** Under the condition of **A change
to `<agents repo>`'s own tooling integrates with the branch's copy of it**
(`skills/flow/integrate-own-tooling.md`), once the branch has landed, run from outside every worktree (Bash
`timeout: 600000`):

```bash
flow-guard refresh-main-checkout <agents repo> <agents-base>
make -C <agents repo>/stats restart
```

then re-read this file from the refreshed checkout and continue with cleanup. The installed
`flow` and flowd predate the change, so the save would exit 2 on a command or table they lack, or
the old cleanup would remove the worktree with no save. The cleanup's own later refresh then finds
nothing to move. This is the one stop of flowd `<agents repo>/CLAUDE.md`'s reinstall rule allows,
brought straight back up.

Call `cleanup` with the parent's own shell outside every worktree in the set.

```bash
flow-guard integrate-change cleanup <main-checkout> <name>
```

Run it with a Bash `timeout: 600000`: a refresh that moved `<agents repo>/stats/` restarts flowd inside the call,
and a kill partway leaves it down.

- **`STOP: self-review-report`** → `cleanup` refuses without a completed `flow.self-review` run:
  run the self-review pass of `skills/flow/integrate.md` — beginning `flow.self-review` again first
  when its run was ended `stopped` — end the stage `completed`, re-run.
- **`STOP: change-save`** → `cleanup` runs `flow change save` in the canonical worktree before
  anything removes a worktree; it keeps `tasks.md` and `self-review-context.md` in flowd, and with them every dispatch report and measure manifest each of the change's worktrees holds, as `reports.<repository>.<file>`. Its exit 1 — a missing `tasks.md` or
  `self-review-context.md`, each named — stops cleanup at `IN_PROGRESS`: report the lines, restore
  the file, and re-run. `remove-change-worktrees` refuses until the save ran.
- **`STOP: worktree-cleanup`** → a `REFUSED:` line, or the guard's exit 2: stop at
  `IN_PROGRESS`, leave every worktree alone, and report each line per **Worktree cleanup**
  (`skills/flow-contracts/finish-contract-run2.md`). A `REFUSED: … check 3` line means the branch
  is not merged: this is not run 2, so fall back to `skills/flow/integrate.md`; nothing was removed.
- **`STOP: state`** → report it; the change stays at `IN_PROGRESS`.

Three stops in a row of `cleanup` take **Three failed rounds** (`skills/flow-contracts/pipeline.md`),
applied inline from every earlier stop's lines — only stops the parent can cure itself. A stop
that exists to ask or to hand to the operator — `STOP: worktree-cleanup` — is never counted toward it: it is handled as its bullet above says, every time.

A `WORKSPACE-REMOVE-FAILED:` line is **the one exception to the stop-at-the-first-failure rule**:
the script carries on. Relay each `REMOTE-*` line as the handoff's
**Remote branch:**.

**Load `skills/flow-contracts/jira-integration.md`.** **Transition the issue to Done** on the
`JIRA: transition <KEY> to Done` line, after the state write, per **Jira integration**
(`skills/flow-contracts/jira-integration.md`).

A `REFRESH-REFUSED` line goes into the handoff verbatim, per step 5 of **Run 2 — the branch is
merged** (`skills/flow-contracts/finish-contract-run2.md`).

```
## Finished

**Change:** <name>
**Worktrees:** removed | left alone — <reason>
**Remote branch:** deleted | already gone | not deleted — <reason>
**Main checkout:** fast-forwarded | already current | <the REFRESH-REFUSED line>
**Jira:** <KEY> → Done | none linked | ⚠ Jira: skipped — <reason>
```

A run 2 that **completes step 4** is terminal and names **no** next command.

**The Finished handoff, and the run's final report beside it, carry no open item** — nothing
"still open", unverified, untraced or left as the operator's call. Each one is resolved before the handoff prints: fixed,
disproved with its evidence, or a finding filed at step 5 of
`/flow-self-review` (`skills/flow-self-review/SKILL.md`).

**Every post-landing item is finished before the handoff prints — a "Not done" line is never the way
out.** The parent records the self-review `fixed` rows for every fix commit
that landed on the base (each sha read from `git log` by its subject; `flow self-review findings
-change <name>` must list them), copies each vendored or sibling-repository file the change's finding
names — in a worktree of that repository, pushed to its `main` — and makes every Jira transition. A
stop in the middle (a daemon holding a worktree, a hook refusing the main checkout) is a blocker to
clear, never a reason to print `Not done` and end the turn.

## Guardrails

- **Never** `git add` the state file, and never move it into the change directory.
