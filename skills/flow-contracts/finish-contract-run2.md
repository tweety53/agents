# Finish contract — run 2 (the branch is merged)

**This file is canonical for `/flow`'s archive run** — run 2's procedure and worktree cleanup.

bare `/flow` is the only command that loads this file whole; `/flow-self-review` reads its step 9.

**Load `skills/flow-contracts/finish-hand-fallbacks.md` only when** the guard presence check named
one of the scripts this file calls missing, or a call finds the script absent — every by-hand
procedure for them is there.

### Run 2 — the branch is merged

1. **Verify the merge.** Use a PR CLI when one is usable for the host; otherwise
   `git merge-base --is-ancestor`. **Not merged → this is not run 2.**
   **On the merge-and-push continuation, the evidence is local**: run 1 merged into `<base>` in
   `<landing-worktree>` without pushing it, so neither a PR CLI nor `origin/<base>` can see the
   merge yet — `git -C <landing-worktree> merge-base --is-ancestor spectre/<name> <base>` is the
   test there.
2. **Position the landing worktree on the archive branch, before anything else touches it.** The
   landing worktree — `<project>/.worktrees/_landing-<name>` — is a throwaway worktree run 2 (and
   the merge-and-push route) uses in place of the main checkout, which is never checked out, staged
   or committed by any `/flow` step. Resolve `<base>` with `resolve-base-branch.sh` against the
   apply worktree, exactly as Run 1 does (`skills/flow-contracts/finish-contract-run1.md`),
   run `classify-untracked.sh <project>/.worktrees/_landing-<name>` when the worktree already
   exists, then invoke `prepare-archive-branch.sh <project>/.worktrees/_landing-<name> <base>
   chore/archive-<name>`. The exit contract: `0` positioned — `<landing-worktree>` is on
   `chore/archive-<name>`, cut from a fast-forwarded `<base>`; `1` a named refusal — the landing
   path's parent is not a `.worktrees` directory, a dirty working tree, **on `<base>` or off it**, a
   detached `HEAD`, an existing archive branch not descended from `origin/<base>`, or a `<base>` or
   `<archive-branch>` name that fails the guard's shape check; `2` an argument is missing,
   `<landing-worktree>` could not be created when absent, or — once positioned — is unreadable or
   not a git worktree, `HEAD`'s own ref cannot be read, or a checkout the guard performs fails; `3`
   `<base>` cannot be reconciled with `origin` — it has diverged, `origin/<base>` does not resolve,
   or there is no `origin` remote at all. Anything but exit `0` stops run 2 here, with nothing
   staged, committed, pushed or removed, and the main checkout never checked out, staged or committed.
   On the merge-and-push continuation, `<base>` already holds run 1's unpushed merge: the
   fast-forward is a no-op, and `chore/archive-<name>` is cut on top of that merge.

   **The pre-flight classifies; the guard still refuses.** `classify-untracked.sh` sorts the
   landing worktree's untracked entries into three classes instead of leaving the guard's flat
   dirty-tree refusal to be cleared by hand: capture images — `*.png`, `*.jpg`,
   `*.jpeg` — move to `<project>/.worktrees/_scratchpad/`, outside the worktree step 11
   force-removes; a `.claude` entry at the worktree root is appended to the checkout's local
   exclude (`<project>/.git/info/exclude` — a linked worktree's own exclude file is not read
   by status), never to the committed gitignore; every other entry is an **asset**, reported and
   touched by nothing. The script never refuses: an absent worktree prints nothing and there is
   nothing to classify — the guard creates it fresh and clean; a settled one prints `CLEAN`; an
   asset stays untracked, so the guard's refusal stands until the run settles it — prompted
   to the operator once, **delete** removes it in place, **commit** moves it to the scratchpad so
   positioning can proceed, then restores and commits it onto `chore/archive-<name>` as its own
   commit immediately after positioning, so it cannot ride step 4's `add -A` unremarked.

3. **Archive the change** — `spectre archive <name>` moves it into
   `<project>/spectre/changes/archive/<name>/`. **The archived leaf carries no date prefix**.
   **One call per change, parent and sub-change alike.** A `<name>-fix-N` sub-change is a flat
   sibling under `<project>/spectre/changes/`, never a directory inside its parent — `spectre new`
   refuses an id that is not a single flat directory name — so the parent's call cannot reach it and
   each sub-change is archived by its own call in this same step. Never left behind, never archived
   alone. **There is nothing to sync into
   `<project>/spectre/specs/` first**.
4. **Commit the archive on `chore/archive-<name>` — no push.** There is no merge to do: the change
   branch was already merged, which step 1 proved. Run 2 merges nothing into the base branch before step 10 lands the archive,
   and never commits the archive on the base branch itself. Every commit run 2 makes after step 2 —
   this one and the self-review context bundle at step 9 alike — asserts `chore/archive-<name>` rather than
   assuming it: naming the directory with `git -C <landing-worktree>` fixes the directory, not the
   branch. A finished change never leaves the archive move uncommitted in the working tree. The push
   happens at step 10, after self-review; step 11, which removes the landing worktree, closes the
   run.

   **The staging is `git add -A`, and this commit's diff is verified scoped to
   `<project>/spectre/changes/`
   before it is made** — `check-archive-scope.sh <landing-worktree> "spectre/changes/"`, run between
   the add and the commit. A `SCOPE-VIOLATION` refuses the commit and leaves the change at `IN_PROGRESS`
   rather than let a stray path land on `chore/archive-<name>` unremarked. **An exit 2 with
   nothing on stdout** — a landing worktree that is not a readable git worktree, or no
   allowed-prefix given — refuses the commit the same way: the guard's header is
   explicit that an inability to answer is never reported as a verdict, so an unreadable tree is
   never read as `SCOPE-OK`.

   **Before the `git add -A`, the rendered ledger and panel record are preserved into this
   commit.** The canonical apply worktree's `<abs-worktree>/.superpowers/sdd/ledgers/<name>.md`
   and `<abs-worktree>/.superpowers/sdd/reviews/<name>-panel.md` are copied into
   `<project>/spectre/changes/archive/<name>/` as `ledger.md` and `panel.md` — each when
   present, an absent file copying nothing — where they ride the archive commit under the scope
   the check above verifies. `<canonical-worktree>` is the resolved set's own canonical member
   (**Run 1 — the branch is not merged**, `skills/flow-contracts/finish-contract-run1.md`), and
   still exists here — its removal is step 5, after this step.
5. **Clean up the worktrees, the local branch and the remote branch, then remove the workspace's
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
   | `remove` | The command that removes them. `/flow`'s archive run calls it, and nothing else does. |
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

   **A failed removal does not stop run 2 here; it is reported, and step 7 decides the verdict** —
   from the project's survivor report and never from this command's exit code, per
   **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).
6. **Remove the proposal artifact source** from the state directory, per its row in
   **Temporary artifacts registry** (`artifacts-registry.md`); a change that published none has
   nothing to remove, and says so.
7. **Verify the cleanup.** Run `check-cleanup-complete.sh <repo> <name> <state-dir>` once
   per repository, **after** every removal above — it is there to judge what the run actually left
   behind. `<state-dir>` is the path `flow state dir` prints for this repository — the same
   resolution serving step 6's removal; a run that has not recently resolved it guesses, and
   a guessed path is how this guard answers exit 2 with no verdict at all.

   | Verdict | What run 2 does |
   |---------|-----------------|
   | `COMPLETE:` | report the cleanup as verified, **relay every clause the line carries after ` — ` word for word**, and go on to step 8 |
   | `LEFTOVER:` | name what remains, **do not write `FINISHED`**, and stop at `IN_PROGRESS` |

   **A `SKIPPED:` clause on a `COMPLETE:` line is relayed, never dropped, and the two rows are
   symmetric for that reason.** The guard appends its notes to the verdict after ` — `, and a
   `SKIPPED:` note there says a registry row was *not* verified — reached, for instance, as
   `COMPLETE: <repo> — … — SKIPPED: the workspace survivor verification — '<cmd>' exited 7, so the
   service could not be reached`. **A skip
   is never a pass**: `<agents repo>/scripts/check-cleanup-complete.sh`'s own header is canonical for why, and it
   is the reason the clause is quoted rather than summarised — the row it leaves unverified and the
   reason it could not be verified are both inside it. The relay does **not** block step 8; why an
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
   operator clears the leftover repeats the verification and nothing else. An already-archived change
   directory means step 3 is already done: the archive move is skipped, not repeated, and the run
   continues to cleanup and verification.

8. **Write `FINISHED`**, clearing from `worktrees` **only the entries whose removal actually
   succeeded** — see **Worktree cleanup**
   (`skills/flow-contracts/finish-contract-run2.md`) below — and carry every other field
   forward. This step is reached only on `COMPLETE:`.
9. **Save the self-review context bundle** — after `FINISHED` is written. Self-review is always
   deferred: no reasoning pass and no prompt run here, and a failure never moves the change off
   `FINISHED`. The session fetches the
   bundle with
   `flow self-review bundle -change <name>`: flowd assembles the whole bundle — the ledger and the
   panel record rendered from the store, or, when the store yields no render for that record, read
   from the copies step 4 committed onto the archive branch, the archived `tasks.md`,
   `design.md` and `narrative.md`
   read out of the `chore/archive-<name>` branch of the repository the command resolves from its
   own location (the main checkout its working directory sits in — the store carries no repository
   roots for the pipeline's changes), and the `git log --stat` of the implementation, planning and
   archive commits — and prints it as one Markdown document. A source that is absent is reported
   `skipped: <source> (absent)` inside the bundle, never fatal; nothing is rendered into the
   landing worktree first, and no landing-worktree path is passed in or baked into the bundle.
   The session appends `## Session narrative` (one paragraph it writes for run 2 itself; the
   archived `narrative.md` is already a bundle section, and a change predating the narrative rule
   has the bundle report it skipped), writes the whole to
   `<project>/docs/self-review/<name>-context.md` physically under `<landing-worktree>`, and
   commits it on `chore/archive-<name>` with subject `docs(self-review): <name> self-review context
   bundle`; step 10 carries the bundle. The pass then runs in `/flow-self-review <name>`
   (`skills/flow-self-review/SKILL.md`), canonical for the five angles, what may be filed, the
   filing-and-rating prompt and the report, which deletes the bundle in its report commit. A
   deferred pass covers what the bundle holds and nothing beyond it — the report's
   `**Deferred:**` line states that.
10. **Push the archive branch and land it — the route depends on how this run of archive.md was
    reached.**

    | Reached via | Then |
    |---|---|
    | the merge-and-push continuation, same invocation as run 1 | in `<landing-worktree>`: `git checkout <base>`; `git merge --ff-only chore/archive-<name>`; `git push origin <base>` — **the route's one push of `<base>`**, carrying run 1's merge, the archive commit and step 9's output together. `chore/archive-<name>` is not pushed; its local branch stays, since `flow self-review bundle` reads the archive from it |
    | a standalone invocation | in `<landing-worktree>`: push `chore/archive-<name>`; open a pull request against `<base>` via a PR CLI when usable for the host, then **merge it immediately with that same CLI** (`gh pr merge --merge --delete-branch` or the host's equivalent) — no wait for checks or review, and no operator prompt, because everything this PR carries is this pipeline's own mechanical output (the archive move, the self-review context bundle), never code a human review gate exists for. When no PR CLI is usable for the host, print the forge's create-PR URL and ask whether it was opened **and merged** — the same shape Run 1's pull-request route uses, extended to cover the merge this row no longer defers |

    This push carries both the archive commit and step 9's context bundle, so there is no window
    in which the archive lands while the bundle is still unwritten. Run 2
    never pushes anything but `chore/archive-<name>` on the standalone row and `<base>` on the
    merge-and-push row; the standalone row's PR-CLI merge does not push `<base>` directly — the forge's own
    merge does that on the PR CLI's behalf.

    A failed push, a failed merge, or a failed pull-request creation is reported with the
    command's own output. It never moves the change off `FINISHED` — the change is already
    terminal by step 8. On the merge-and-push row nothing is lost: the merge, the archive and
    step 9's output stay in the local `<base>` and `chore/archive-<name>` refs, and the handoff
    prints `git -C <main-checkout> push origin <base>` to land them. On the standalone row, a PR opened but not yet merged — the CLI's merge
    call itself failed, or no CLI was usable and the operator has not yet confirmed it — has the
    handoff name the open PR and print the exact merge command needed to land it by hand; a PR
    that was never even opened falls back to naming the unpushed branch and the create-PR command
    instead. The archive is never reported as landed until this step actually merges it, on either
    row.

    **This split reads no persisted field.** The merge-and-push row is recognized because this run
    of archive.md is executing as `skills/flow/integrate.md`'s own same-invocation continuation —
    a fact already in scope for that one call path, never written to the state file. A standalone
    invocation (the PR and manual routes always defer archiving this way) has no way to know what
    the original route was, or whether the operator merged the PR through some mechanism this
    pipeline never chose — so it always takes the standalone row.
11. **Remove the landing worktree.** Successful or not:

    ```bash
    git -C <main-checkout> worktree remove --force <landing-worktree>
    git -C <main-checkout> worktree prune
    ```

    Then refresh the main checkout:

    ```bash
    refresh-main-checkout.sh <main-checkout> <base>
    ```

    The script hard-resets
    only when the index is byte-for-byte the tree of an earlier `<base>` tip and the worktree
    equals the index — pure staleness, nothing to lose; anything else is refused by name and
    reported, never reset. `REFRESH-REFUSED` is reported in the handoff and does not move the
    change off `FINISHED`.

**The Jira `Done` transition fires before step 9, not after it.**

### Worktree cleanup

Which worktrees those are, and the `git worktree list --porcelain` scan that finds them when
the state file's map is absent or empty, are **Resolving a change's worktrees**
(`skills/flow-contracts/finish-contract-run1.md`). **The landing worktree is never one of them.**
This section resolves and removes the change's own **apply** worktree(s) — the ones
`spectre/<name>` was ever checked out in — a disjoint set from the single throwaway
`<project>/.worktrees/_landing-<name>` that steps 2–4, 9 and 10 above used in the main checkout's
place; step 11, which this section's own call site (step 5) precedes, is what removes that one.

For each worktree, run **every** check below before removing anything:

```bash
# 1. no uncommitted tracked changes — must be empty
git -C "$WT" status --porcelain --untracked-files=no

# 2. no untracked files that git does not already ignore — must be empty.
#    `--others --exclude-standard` lists exactly the files `--force` would destroy and
#    `.gitignore` does NOT cover. It does not make `--force` safe: ignored files are check 4's.
git -C "$WT" ls-files --others --exclude-standard

# 3. no commits that exist only here. Resolve `BASE` fresh for THIS worktree — never reused from
#    another worktree in the set. A multi-repo change has one `origin` and one default branch per
#    repository, so a name resolved against one worktree's `origin` can be the wrong ref, or absent
#    entirely, in another repository. Resolve it here, before this worktree's own removal below.
BASE="$(resolve-base-branch.sh "$WT")" || { echo "cannot resolve the base branch for $WT — stop and ask"; false; }
#    `@{upstream}` ERRORS when no upstream is configured, and an empty capture would read as
#    "nothing unpushed" — so resolve it explicitly and never let a failed lookup pass as success.
if git -C "$WT" merge-base --is-ancestor HEAD "origin/$BASE" 2>/dev/null; then
  :                                            # already merged into base — nothing can be lost
elif UP="$(git -C "$WT" rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null)"; then
  git -C "$WT" log --oneline "$UP..HEAD"       # must be empty
else
  echo "not merged, and no upstream — cannot prove these commits exist anywhere else"; false
fi

# 4. what `--force` WILL destroy: ignored files, split into what a next build, test or
#    dev-stack run regenerates byte-for-byte and everything else. `--exclude-standard` in check 2 hides
#    everything matched by .gitignore, <project>/.git/info/exclude or the global excludes file,
#    and "ignored" is NOT "disposable" in general — a deliberately-ignored .env or a local
#    override config is ignored and irreplaceable. The split below is what tells the two apart.
git -C "$WT" ls-files --others --ignored --exclude-standard

# 5. the project's local stack is stopped — run its `## stop` command if declared. Give it a
#    bounded wait — 60 seconds — and treat a timeout as a FAILED check.

# 6. no process is still running FROM this worktree. Runs whether or not `## stop` is declared,
#    and whatever that command exited: a project that declares no stop command can still have a
#    stack running from its worktree, and a stop command that exits 0 is not evidence that it
#    worked. Verifying that is the whole of what this check adds.
check-worktree-processes.sh "$WT"
```

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
Split what it found into two buckets by path, never by guessing intent:

- **Regeneratable** — a path under a build/cache/log/test-output location a fresh build, test run
  or `devStart` recreates with identical content the next time it runs: any path component named
  `build`, `.gradle`, `.kotlin`, `node_modules`, `dist`, `.next`, `target`, `out`, `coverage` or
  `test-results`; any `*.log`; and this pipeline's own `<abs-worktree>/.superpowers/sdd/` and
  `<abs-worktree>/.dev-stack/` trees, which a session's own next run writes fresh. A `.png`/`.jpg`
  capture is regeneratable only when it sits under a `test-results` directory (or an equivalent
  declared screenshot-output directory) a test run owns end to end — never a capture sitting loose at a
  project root, which could be a hand-saved reference nothing re-creates.
- **Everything else** — unclassified, and it stays unclassified: no allowlist of *names* is
  trusted here, because the operator decides what they ignore, and a path this list does not
  recognize (a `.env`, a local override config, a stray file at the worktree root) is exactly what
  the doubt is for.

An empty unclassified bucket → **show the regeneratable bucket's count and proceed without
asking**. A
non-empty unclassified bucket → **show that bucket in full** (the regeneratable one named only by
count), saying of each entry whether it is irreplaceable and whether it was already preserved,
**and proceed without asking** — by this point the change's own work is committed and step 1
proved it merged, and the rendered ledger and panel record ride step 4's archive commit. **The one
ask:** an entry that is irreplaceable and *unpreserved* → stop and ask for explicit confirmation
before removing that worktree.

**`/flow-fast` runs none of these checks.** Its cleanup (**8. Clean up**, `skills/flow-fast/SKILL.md`)
removes its worktree without `--force`, so git itself refuses one holding modified or untracked
files.

Then, and only then:

```bash
git -C "$REPO" worktree remove --force "$WT"
git -C "$REPO" branch -d "spectre/<name>"
git -C "$REPO" worktree prune
```

- **`--force` destroys every ignored file in the worktree, and no check prevents that.** Checks 1
  and 2 establish only that nothing *tracked-and-modified* and nothing *untracked-and-unignored*
  is at risk. They say nothing about ignored files, because `--exclude-standard` is what hides
  them — and "ignored" is not "disposable". Check 4 exists to make that visible rather than to
  prevent it: it lists exactly what will die, and asks only about an irreplaceable, unpreserved entry. Claiming the checks make
  `--force` safe would be false: a gitignored `.env` passes checks 1 and 2 and is
  destroyed silently.
- **`git branch -d`, never `-D`.** It must be free to refuse an unmerged branch.
- **An already-removed worktree is success**, not an error.
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

Then the change's **remote** branch:

```bash
OUT="$(git -C "$REPO" push origin --delete "spectre/<name>" 2>&1)"; RC=$?
if [ "$RC" -eq 0 ]; then
  echo "remote branch deleted: origin/spectre/<name>"
elif printf '%s' "$OUT" | grep -q 'remote ref does not exist'; then
  # The forge deleted it on merge. Prune the ref it left behind: check-cleanup-complete.sh reads a
  # surviving refs/remotes/origin/spectre/<name> as a leftover, and it would be a real one.
  git -C "$REPO" fetch --prune --quiet origin
  echo "remote branch already gone — the forge deleted it on merge"
else
  echo "remote branch NOT deleted: $OUT"
fi
```

- **The remote branch is deleted without a further prompt.**
- **An already-absent remote branch is success**, not an error, and the outcome is reported either
  way: deleted, already gone, or refused.
- **A refused push is reported, never swallowed.**
- **The remote delete is not gated on the local one succeeding.**

The stack-stopped check reads the optional `## stop` key from `<project>/.flow/project.md` —
see **Project configuration** in `skills/flow-contracts/project-configuration.md`. When the key
or the file is absent the check is **skipped, not failed**, and cleanup proceeds on the strength of
the other checks.

