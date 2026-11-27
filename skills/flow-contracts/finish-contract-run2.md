# Finish contract — run 2 (the branch is merged)

**This file is canonical for `/flow`'s cleanup run** — run 2's procedure and worktree cleanup.
Run 1 already ran its self-review pass after landing (**Run the self-review pass**,
`skills/flow-contracts/finish-contract-run1.md`), and **`cleanup` refuses — `STOP: self-review-report` — without a completed `flow.self-review` run**; run 2 commits and pushes nothing: it saves the
change's plan and self-review bundle to flowd and cleans up once the branch has merged — which the
preflight's `RUN2` verdict and the worktree cleanup's check 3 (no commit that exists only here) already
establish, so run 2 tests it nowhere else.

bare `/flow` is the only command that loads this file whole.

### Run 2 — the branch is merged

1. **Clean up the worktrees, the local branch and the remote branch, then remove the workspace's
   database and bucket** — the worktree half being **Worktree cleanup**
   (`skills/flow-contracts/finish-contract-run2.md`) below.

   **`BASE` is resolved per worktree, inside the cleanup loop below — never once for the whole
   change.** Anything but exit `0` for a given worktree — stop and ask, exactly as
   Run 1 does, and leave every worktree alone, per **Any failed check leaves every worktree alone**
   below.

   The removal runs the project's `remove` command, read from the command table below, with the workspace id substituted into its text by the token rule below it.

   The command table has up to three rows and two columns:

   | Command | Runs |
   |---------|------|
   | `create` | The command that creates this workspace's resources when they are absent. Whatever starts the project's applications calls it. |
   | `remove` | The command that removes them. `/flow`'s cleanup run calls it, and nothing else does. |
   | `orphans` | Optional. The command that prints, one per line, the id of every workspace on the machine that no live worktree derives. Run 2 calls it from the main checkout after its own `remove`, runs `remove` for each id printed with that id substituted, and reports `WORKSPACE-ORPHAN-REMOVED: <id>` or `WORKSPACE-ORPHAN-REMOVE-FAILED: <id> — exit <rc>`, and `WORKSPACE-ORPHANS-FAILED: <repo> — exit <rc>` when `orphans` itself fails; no failure of either stops run 2. |

   **Each command runs with a repository root as its working directory, and which root is fixed here
   rather than left to the caller.**

   - **`remove` runs from the main checkout** too. Run 2 calls it after the worktree half of its cleanup
     step, per **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`).
   - **`create` runs from the apply worktree**, because of who calls it: whatever starts the project's
     applications does, and that is the worktree whose applications need the resources.

   The tokens `<id>` and `<id_underscored>` are substituted in a command's text too, and that is how the
   workspace id reaches it — one mechanism for both tables, so there is no argument convention to
   remember alongside it. `<value:…>` is **not** substituted in a command: a command that needs a
   derived value reads the exported variable, which is already in its environment. **Which command a
   project names is the project's own decision**, reusing one
   it already ships or adding one, per **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).

   **A failed removal does not stop run 2 here; it is reported.**
2. **Remove the proposal artifact source** from the state directory, per its row in
   **Temporary artifacts registry** (`artifacts-registry.md`); a change that published none has
   nothing to remove, and says so.
3. **Write `FINISHED`**, clearing from `worktrees` **only the entries whose removal actually
   succeeded** — see **Worktree cleanup**
   (`skills/flow-contracts/finish-contract-run2.md`) below — and carry every other field
   forward.
4. **Bring the main checkout forward** — after `FINISHED` is written, once per repository, with
   `<base>` the branch `refs/remotes/origin/HEAD` names, its `origin/` prefix stripped — the
   worktree `flow-guard resolve-base-branch` reads is gone by now:

   ```bash
   flow-guard refresh-main-checkout <main-checkout> <base>
   ```

   A main checkout that has a `--base` change's `<base>` (`skills/flow/SKILL.md`) checked out is
   refreshed onto that `<base>` instead — the branch the merge check resolved for that
   repository's worktree before removing it — so it is brought forward rather than refused.

   The script fast-forwards the main checkout onto `origin/<base>` only when that loses nothing,
   and refuses by name otherwise; its header is canonical for the conditions. `REFRESH-REFUSED`
   and `FLOWD-RESTART-FAILED` are reported in the handoff, the latter with its `make` command for
   the operator to run, and neither moves the change off `FINISHED`.

### Worktree cleanup

Which worktrees those are, and the `git worktree list --porcelain` scan that finds them when
the state file's map is absent or empty, are **Resolving a change's worktrees**
(`skills/flow-contracts/finish-contract-run1.md`).

Once per repository, from outside every worktree in the set:

```bash
flow-guard remove-change-worktrees <repo> <name>
```

Exit 0: every check passed; each `REMOVED:` entry is one step 3 clears, and each `REMOTE-*` line is
the handoff's **Remote branch:** — deleted, already gone, or not deleted. Exit 1: a `REFUSED:`
line — a failed check, nothing removed; or a failed removal, whose entry stays in
`worktrees`. Exit 2: stop and report. Ignored files go with the worktree's `--force` removal,
never asked about.

**`/flow-fast` runs none of these checks.** Its cleanup (**8. Clean up**, `skills/flow-fast/SKILL.md`)
removes its worktree without `--force`, so git itself refuses one holding modified or untracked
files.

- **Any failed check leaves every worktree alone** and reports why. There is no partial cleanup.
- **Verify each removal actually succeeded** before writing state. If any `git worktree remove`
  fails for a reason the checks did not predict — a file lock, a permission error — report it and
  leave that worktree's entry in `worktrees`.

**Throwaway copies go with the apply worktree they were copied from.** For each `$WT`, every
entry `git -C "$REPO" worktree list --porcelain` lists at `$WT-mutation-<round>`, `detached` or on its own branch under `flow-copy`, is a
copy the review panel's mutation slot made and a run that died mid-round never
removed. Checks 1–3 do not apply to it: a copy starts from the canonical worktree's uncommitted state, so a clean tree or a patch-id match proves nothing about it. Each copy is removed without asking, before `$WT`, with
`git -C "$REPO" worktree remove --force "$COPY"` and no `git branch -d` (a copy has no branch).
