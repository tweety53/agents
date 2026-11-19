# Finish contract — run 2 (the branch is merged)

**This file is canonical for `/flow`'s cleanup run** — run 2's procedure and worktree cleanup.
Run 1 already ran its self-review pass on the change branch (**Run the self-review pass**,
`skills/flow-contracts/finish-contract-run1.md`), so run 2 commits and pushes nothing: it saves the
change's plan and self-review bundle to flowd and cleans up once the branch has merged.

bare `/flow` is the only command that loads this file whole.

**Load `skills/flow-contracts/finish-hand-fallbacks.md` only when** the guard presence check named
one of the scripts this file calls missing, or a call finds the script absent — every by-hand
procedure for them is there.

### Run 2 — the branch is merged

1. **Verify the merge.** The branch is merged when a PR CLI usable for the host reports its PR
   merged, or when `git merge-base --is-ancestor` holds, whatever the PR CLI says. **Not merged →
   this is not run 2.**
   On the merge-and-push continuation, run 1's push already landed the branch on `origin/<base>`,
   so the same test holds there.
2. **Clean up the worktrees, the local branch and the remote branch, then remove the workspace's
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
3. **Remove the proposal artifact source** from the state directory, per its row in
   **Temporary artifacts registry** (`artifacts-registry.md`); a change that published none has
   nothing to remove, and says so.
4. **Write `FINISHED`**, clearing from `worktrees` **only the entries whose removal actually
   succeeded** — see **Worktree cleanup**
   (`skills/flow-contracts/finish-contract-run2.md`) below — and carry every other field
   forward.
5. **Bring the main checkout forward** — after `FINISHED` is written, once per repository, with
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

Once per repository, from outside every worktree in the set, with `<merge-base>` the one the state
file's `worktrees` map records for it (`-` when none):

```bash
flow-guard remove-change-worktrees <repo> <name> <merge-base|->
```

Exit 0: every check passed; each `REMOVED:` entry is one step 4 clears, and each `REMOTE-*` line is
the handoff's **Remote branch:** — deleted, already gone, or not deleted. Exit 1: a `REFUSED:`
line — a failed check, nothing removed; or a failed removal, whose entry stays in
`worktrees`. Exit 2: stop and report. Ignored files go with the worktree's `--force` removal,
never asked about.

**`/flow-fast` runs none of these checks.** Its cleanup (**8. Clean up**, `skills/flow-fast/SKILL.md`)
removes its worktree without `--force`, so git itself refuses one holding modified or untracked
files.

- **Any failed check leaves every worktree alone** and reports why. There is no partial cleanup.
  This includes check 5: a `## stop` command that **exits non-zero, is not found, or has to be
  interrupted** is a *failed* check, not an absent one — only an undeclared key is skipped. Give it
  a bounded wait rather than letting it hang the run.
- **Verify each removal actually succeeded** before writing state. If any `git worktree remove`
  fails for a reason the checks did not predict — a file lock, a permission error — report it and
  leave that worktree's entry in `worktrees`.

**Wave-group copies go with the apply worktree they were copied from.** For each `$WT`, every
entry `git -C "$REPO" worktree list --porcelain` lists at `$WT-wave-group-<g>`, `detached` or on its own branch under `flow-copy`, is a
copy **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) made and a run that died mid-wave never
removed. Check 5 above runs on each such `$COPY` in `$WT`'s place and stays a gate. Checks 1–3
do not: a copy starts from the canonical worktree's uncommitted state, and run 1's reshape folds
every picked commit into one, so a clean tree or a patch-id match proves nothing about it. Every
task the copy ran is either ticked in `tasks.md` and landed through the reshape, or unticked and
re-run by a later implement run, so each copy is removed without asking, before `$WT`, with
`git -C "$REPO" worktree remove --force "$COPY"` and no `git branch -d` (a copy has no branch); on
a failed check 5, every worktree is left alone, exactly as above.

The stack-stopped check reads the optional `## stop` key from `<project>/.flow/project.md` —
see **Project configuration** in `skills/flow-contracts/project-configuration.md`. When the key
or the file is absent the check is **skipped, not failed**, and cleanup proceeds on the strength of
the other checks. A body with no fenced command is that key's same skip, and prints
`SKIPPED: check 5 — ## stop declares no fenced command`; relay the line word for word.

