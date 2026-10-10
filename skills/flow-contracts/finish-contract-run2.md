# Finish contract — run 2 (the branch is merged)

**This file is canonical for `/flow`'s cleanup run** — run 2's procedure and worktree cleanup.
Run 1 already archived the change and ran its self-review pass on the change branch
(**Archive on the change branch**, `skills/flow-contracts/finish-contract-run1.md`), so run 2
archives, commits and pushes nothing: it cleans up once the branch has merged.

bare `/flow` is the only command that loads this file whole.

**Load `skills/flow-contracts/finish-hand-fallbacks.md` only when** the guard presence check named
one of the scripts this file calls missing, or a call finds the script absent — every by-hand
procedure for them is there.

### Run 2 — the branch is merged

1. **Verify the merge.** Use a PR CLI when one is usable for the host; otherwise
   `git merge-base --is-ancestor`. **Not merged → this is not run 2.**
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

   The command table has three rows and two columns:

   | Command | Runs |
   |---------|------|
   | `create` | The command that creates this workspace's resources when they are absent. Whatever starts the project's applications calls it. |
   | `remove` | The command that removes them. `/flow`'s cleanup run calls it, and nothing else does. |
   | `survivors` | The command that reports which of them still exist. Run 2 calls it after `remove`, and `<agents repo>/scripts/check-cleanup-complete.sh` turns its result into the registry row's verdict. Its output and its exit code are read, so both are specified below. |

   Why a third verb rather than two — why "ran `remove`" is not "verified gone", and why a guard in the
   agents repository cannot ask the question itself — is stated under **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).

   **Each command runs with a repository root as its working directory, and which root is fixed here
   rather than left to the caller.** See
   **Working directory for `survivors`, `remove`, and `create`**
   (`skills/flow-contracts/project-configuration-rationale.md`) for why it must be stated.

   - **`survivors` runs from the main checkout**, and that is not a convention invented here. See
     **Working directory for `survivors`, `remove`, and `create`**
     (`skills/flow-contracts/project-configuration-rationale.md`) for how
     `<agents repo>/scripts/check-cleanup-complete.sh` invokes it.
   - **`remove` runs from the main checkout** too. Run 2 calls it after the worktree half of its cleanup
     step, per **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`). See
     **Working directory for `survivors`, `remove`, and `create`**
     (`skills/flow-contracts/project-configuration-rationale.md`) for why.
   - **`create` runs from the apply worktree**, because of who calls it: whatever starts the project's
     applications does, and that is the worktree whose applications need the resources. See
     **Working directory for `survivors`, `remove`, and `create`**
     (`skills/flow-contracts/project-configuration-rationale.md`) for why the asymmetry is the rule
     working rather than an exception to it.

   The tokens `<id>` and `<id_underscored>` are substituted in a command's text too, and that is how the
   workspace id reaches it — one mechanism for both tables, so there is no argument convention to
   remember alongside it. `<value:…>` is **not** substituted in a command: a command that needs a
   derived value reads the exported variable, which is already in its environment. **Which command a
   project names is the project's own decision**, reusing one
   it already ships or adding one, per **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).

   **A failed removal does not stop run 2 here; it is reported, and step 4 decides the verdict** —
   from the project's survivor report and never from this command's exit code, per
   **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).
3. **Remove the proposal artifact source** from the state directory, per its row in
   **Temporary artifacts registry** (`artifacts-registry.md`); a change that published none has
   nothing to remove, and says so.
4. **Verify the cleanup.** Run `check-cleanup-complete.sh <repo> <name> <state-dir>` once
   per repository, **after** every removal above — it is there to judge what the run actually left
   behind. `<state-dir>` is the path `flow state dir` prints for this repository — the same
   resolution serving step 3's removal; a run that has not recently resolved it guesses, and
   a guessed path is how this guard answers exit 2 with no verdict at all.

   | Verdict | What run 2 does |
   |---------|-----------------|
   | `COMPLETE:` | report the cleanup as verified, **relay every clause the line carries after ` — ` word for word**, and go on to step 5 |
   | `LEFTOVER:` | name what remains, **do not write `FINISHED`**, and stop at `IN_PROGRESS` |

   **A `SKIPPED:` clause on a `COMPLETE:` line is relayed, never dropped, and the two rows are
   symmetric for that reason.** The guard appends its notes to the verdict after ` — `, and a
   `SKIPPED:` note there says a registry row was *not* verified — reached, for instance, as
   `COMPLETE: <repo> — … — SKIPPED: the workspace survivor verification — '<cmd>' exited 7, so the
   service could not be reached`. **A skip
   is never a pass**: `<agents repo>/scripts/check-cleanup-complete.sh`'s own header is canonical for why, and it
   is the reason the clause is quoted rather than summarised — the row it leaves unverified and the
   reason it could not be verified are both inside it. The relay does **not** block step 5; why an
   unreachable service must not strand an already-merged change is stated once under
   **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).

   A non-zero exit with **no verdict line** is the third outcome and not a verdict: report it, leave
   the affected `worktrees` entries in the state file, and treat it exactly as `LEFTOVER` — an
   unverified cleanup is not a verified one. The exit code is checked as well as the line, because a
   caller that greps for `COMPLETE` in empty output finds nothing.

   **None of this is a cue to bring the stack back up.**

   **A leftover blocks the `FINISHED` write, and that is the whole point of having a verdict.**

   **Run 2 is re-entrant, which is what makes that safe.** Every step is remove-or-move *if present*
   and a step whose artifact is already gone is success, not an error — so a re-run after the
   operator clears the leftover repeats the verification and nothing else.

5. **Write `FINISHED`**, clearing from `worktrees` **only the entries whose removal actually
   succeeded** — see **Worktree cleanup**
   (`skills/flow-contracts/finish-contract-run2.md`) below — and carry every other field
   forward. This step is reached only on `COMPLETE:`.
6. **Bring the main checkout forward** — after `FINISHED` is written, once per repository, with
   `<base>` the branch `refs/remotes/origin/HEAD` names, its `origin/` prefix stripped — the
   worktree `resolve-base-branch.sh` reads is gone by now:

   ```bash
   refresh-main-checkout.sh <main-checkout> <base>
   ```

   The script fast-forwards the main checkout onto `origin/<base>` only when that loses nothing,
   and refuses by name otherwise; its header is canonical for the conditions. `REFRESH-REFUSED`
   is reported in the handoff and does not move the change off `FINISHED`.

### Worktree cleanup

Which worktrees those are, and the `git worktree list --porcelain` scan that finds them when
the state file's map is absent or empty, are **Resolving a change's worktrees**
(`skills/flow-contracts/finish-contract-run1.md`).

Once per repository, from outside every worktree in the set, with `<merge-base>` the one the state
file's `worktrees` map records for it (`-` when none):

```bash
remove-change-worktrees.sh <repo> <name> <merge-base|->   # --proceed on the second call only
```

Exit 0: every check passed; each `REMOVED:` entry is one step 5 clears, and each `REMOTE-*` line is
the handoff's **Remote branch:** — deleted, already gone, or not deleted. Exit 1: a `REFUSED:` or
`HELD:` line — a failed check, nothing removed; or a failed removal, whose entry stays in
`worktrees`. Exit 2: stop and report. Exit 3: the disclosure stop — nothing removed and nothing
run; relay its `UNCLASSIFIED:` and `DISCLOSE:` lines per check 4 and the wave-group copies below,
ask only what they ask, then call again with `--proceed`.
An unclassified entry byte-identical to the same path in the repository's main checkout is preserved: its `UNCLASSIFIED:` line ends `— preserved: identical in <main-checkout>`, it never makes the disclosure stop, and it is relayed per check 4 whatever the exit.

**Immediately before check 6, `remove-change-worktrees` stops the worktree's `worktree-lsp`
children** — every language server a `worktree-lsp` wrapper runs at or under the worktree — so a
session's own LSP servers never hold it; check 6 reports any that still do. A check 5 failure on any
worktree removes none, and stops none of their servers.

**Check 6 requires the orchestrating shell's own cwd to be outside every worktree in the resolved
set before it runs.** `cd` out first, for every worktree, before
this check runs for any of them.

**Check 6 is a gate, and both of its bad outcomes are failures.** `HELD:` names each process
holding the worktree, and exit 2 says the guard could not answer at all — the scanning tool is
absent, or the worktree path is not readable. On either, stop
at `IN_PROGRESS`, leave every worktree alone, and report each pid, its working directory, and the
command that reaches it:

```bash
ps -o pid,command -p <pid>
```

**There is no confirmation to proceed past check 6**, unlike check 4's one ask. The remedy is to clear the
process and re-run, while
the worktree still exists and the project's own stop command can still read what it started.

**Check 4 is a disclosure, not a gate — it asks only about an irreplaceable entry nothing preserved.**

An empty unclassified bucket → **show the regeneratable bucket's count and proceed without
asking**. A
non-empty unclassified bucket → **show that bucket in full** (the regeneratable one named only by
count), saying of each entry whether it is irreplaceable and whether it was already preserved,
**and proceed without asking** — by this point the change's own work is committed and step 1
proved it merged, and the rendered ledger and panel record ride run 1's archive commit. **The one
ask:** an entry that is irreplaceable and *unpreserved* → stop and ask for explicit confirmation
before removing that worktree.

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
entry `git -C "$REPO" worktree list --porcelain` lists as `detached` at `$WT-wave-group-<g>` is a
copy **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) made and a run that died mid-wave never
removed. Checks 5 and 6 above run on each such `$COPY` in `$WT`'s place and stay gates. Checks 1–4
do not: a copy starts from the canonical worktree's uncommitted state, and run 1's reshape folds
every picked commit into one, so a clean tree or a patch-id match proves nothing about it. Instead
**disclose** `git -C "$COPY" status --short` and `git -C "$COPY" log --oneline <merge-base>..HEAD`,
then ask once — **Remove the wave-group copies?** — **Yes — remove them** *(recommended)* / **No —
stop and keep them**. Every task the copy ran is either ticked in `tasks.md` and landed through the
reshape, or unticked and re-run by a later implement run, so what the disclosure shows is
diagnostic, not work. On Yes each copy is removed before `$WT`, with
`git -C "$REPO" worktree remove --force "$COPY"` and no `git branch -d` (a copy has no branch); on
No, or a failed check 5 or 6, every worktree is left alone, exactly as above.

The stack-stopped check reads the optional `## stop` key from `<project>/.flow/project.md` —
see **Project configuration** in `skills/flow-contracts/project-configuration.md`. When the key
or the file is absent the check is **skipped, not failed**, and cleanup proceeds on the strength of
the other checks. A body with no fenced command is that key's same skip, and prints
`SKIPPED: check 5 — ## stop declares no fenced command`; relay the line word for word.

