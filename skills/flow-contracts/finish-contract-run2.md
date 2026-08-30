# Finish contract — run 2 (the branch is merged)

**This file is canonical for `/myflow-finish`'s run 2** — run 2's procedure and worktree cleanup.

`/myflow-finish` is the only command that loads this file.

### Run 2 — the branch is merged

1. **Verify the merge.** Use a PR CLI when one is usable for the host; otherwise
   `git merge-base --is-ancestor`. That fallback must stay reachable on its own — it is the only
   merge evidence available on a non-GitHub forge. **Not merged → this is not run 2.**
2. **Position the main checkout on the archive branch, before anything else touches it.** Resolve
   `<base>` with `resolve-base-branch.sh` against the apply worktree, exactly as Run 1 does
   (`skills/flow-contracts/finish-contract-run1.md`),
   then invoke `prepare-archive-branch.sh <main-checkout> <base> chore/archive-<name>`. The exit
   contract: `0` positioned — the checkout is on `chore/archive-<name>`, cut from a fast-forwarded
   `<base>`; `1` a named refusal — a dirty working tree, **on `<base>` or off it**, a detached
   `HEAD`, an existing archive branch not descended from `origin/<base>`, or a `<base>` or
   `<archive-branch>` name that fails the guard's shape check; `2` an argument is missing,
   `<main-checkout>` is missing, unreadable, or not a git worktree, `HEAD`'s own ref cannot be read,
   or a checkout the guard performs fails; `3` `<base>` cannot be reconciled with `origin` — it
   has diverged, `origin/<base>` does not resolve, or there is no `origin` remote at all.
   Anything but exit `0` stops run 2 here, with nothing staged, committed, pushed or removed.

   **When the script is absent** — a harness whose repository does not carry it — perform the same
   positioning by hand, in this order, and say in the handoff that it was done manually. The guard is
   never skipped for want of the script. The steps are stated in full rather than cited, for the
   reason the resolver's own fallback above gives: the guard being absent takes its header with it.
   Run the wrapped, credential-free fetch first, so an
   unreachable remote refuses quickly rather than hanging. Then read `HEAD`: **refuse a detached
   `HEAD`.** Then read the working tree with `git -C <main-checkout> status --porcelain` — the same
   test the preflight's signal 3 uses — and **refuse a dirty tree wherever it is found, on `<base>` as
   well as off it**, naming both the branch found and `<base>`; uncommitted changes would otherwise
   ride onto the archive branch unremarked. On a clean tree, check out `<base>` if the checkout is not
   already on it, then fast-forward it to `origin/<base>`, **refusing a base that cannot fast-forward**
   rather than merging or resetting it. Finally create `chore/archive-<name>` from that base and check
   it out — or, when it already exists, **reuse it only if it is descended from `origin/<base>`** and
   refuse it otherwise. Apply each refusal in that order and never accept a guess in place of any of
   them.
3. **Archive the change** — `spectre archive <name>` moves it into
   `<project>/spectre/changes/archive/<name>/`. **The archived leaf carries no date prefix**, because
   `spectre archive` adds none: the date-ordered archive OpenSpec gave this pipeline is a real loss,
   accepted, and a prefix re-added here would describe a move the tool does not perform.
   **One call per change, parent and sub-change alike.** A `<name>-fix-N` sub-change is a flat
   sibling under `<project>/spectre/changes/`, never a directory inside its parent — `spectre new`
   refuses an id that is not a single flat directory name — so the parent's call cannot reach it and
   each sub-change is archived by its own call in this same step. Never left behind, never archived
   alone. **There is nothing to sync into
   `<project>/spectre/specs/` first, and no sync step is ever to be added back**: a change edits that
   tree directly on its own branch, so its spec edits reached the base branch with the merge step 1
   proved.
4. **Commit the archive on `chore/archive-<name>` — no push.** There is no merge to do: the change
   branch was already merged, which step 1 proved. Run 2 never merges anything into the base branch,
   and never commits the archive on the base branch itself. Every commit run 2 makes after step 2 —
   this one and the self-review report at step 9 alike — asserts `chore/archive-<name>` rather than
   assuming it: naming the directory with `git -C <main-checkout>` fixes the directory, not the
   branch. A finished change never leaves the archive move uncommitted in the working tree. The push
   happens at step 10, after self-review; step 11, which restores the checkout, closes the run.
5. **Clean up the worktrees, the local branch and the remote branch, then remove the workspace's
   database and bucket** — the worktree half being **Worktree cleanup**
   (`skills/flow-contracts/finish-contract-run2.md`) below.

   **`BASE` is resolved per worktree, inside the cleanup loop below — never once for the whole
   change.** A multi-repo change has one `origin` and one default branch per repository, so a name
   resolved against one worktree can be the wrong ref, or no ref at all, against another
   repository's `origin`. Worktree cleanup's check 3 below is where this resolution actually runs:
   for each worktree in the resolved set it invokes `resolve-base-branch.sh` against that same
   worktree, immediately before that worktree's own removal — the same call and exit contract as
   **Resolve the base branch** under Run 1
   (`skills/flow-contracts/finish-contract-run1.md`), run again here because this is a separate
   invocation and nothing carries `BASE` over from run 1's. Step 4's archive commit runs before this
   step, so every worktree in the set is still present when its own resolution runs — satisfying the
   same "before cleanup removes it" ordering this step has always required, just per worktree rather
   than once for the change. Anything but exit `0` for a given worktree — stop and ask, exactly as
   Run 1 does, and leave every worktree alone, per **Any failed check leaves every worktree alone**
   below.

   The removal runs the project's `remove` command, read from the command table
   **Project configuration** (`skills/flow-contracts/project-configuration.md`) is canonical for,
   with the workspace id substituted into its text by the mechanism that same file defines. **Run 2
   is not handed that id and does not need to be**: it is derived from the change name and from
   nothing else, deterministically and without ever being recorded, per
   **The workspace id** (`skills/flow-contracts/workspace-isolation.md`) — so run 2 re-derives it
   and arrives at the id `/myflow-do` used, in a session that shared nothing with it.

   **A project declaring no `## workspace isolation` section, or no `remove` command in it, has this
   half skipped rather than failed** — a step whose artifact is already absent is a success, which is
   the same re-entrancy rule every other removal in run 2 follows.

   **A failed removal does not stop run 2 here; it is reported, and step 7 decides the verdict** —
   from the project's survivor report and never from this command's exit code, per
   **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).
6. **Remove the proposal artifact source** from the state directory, on the condition its row in
   **Temporary artifacts registry** (`artifacts-registry.md`) gives. That section carries the condition and
   the reason for it; this step does not repeat either.
7. **Verify the cleanup.** Run `check-cleanup-complete.sh <repo> <name> <state-dir>` once
   per repository, **after** every removal above — it is there to judge what the run actually left
   behind, which is the one thing run 2 previously assumed.

   | Verdict | What run 2 does |
   |---------|-----------------|
   | `COMPLETE:` | report the cleanup as verified, **relay every clause the line carries after ` — ` word for word**, and go on to step 8 |
   | `LEFTOVER:` | name what remains, **do not write `FINISHED`**, and stop at `IN_PROGRESS` |

   **A `SKIPPED:` clause on a `COMPLETE:` line is relayed, never dropped, and the two rows are
   symmetric for that reason.** The guard appends its notes to the verdict after ` — `, and a
   `SKIPPED:` note there says a registry row was *not* verified — reached, for instance, as
   `COMPLETE: <repo> — … — SKIPPED: the workspace survivor verification — '<cmd>' exited 7, so the
   service could not be reached`. A run that reported only "cleanup verified" would have told the
   operator the opposite of what the guard said, while following this table to the letter. **A skip
   is never a pass**: `<agents repo>/scripts/check-cleanup-complete.sh`'s own header is canonical for why, and it
   is the reason the clause is quoted rather than summarised — the row it leaves unverified and the
   reason it could not be verified are both inside it. The relay does **not** block step 8; why an
   unreachable service must not strand an already-merged change is stated once under
   **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).

   A non-zero exit with **no verdict line** is the third outcome and not a verdict: report it, leave
   the affected `worktrees` entries in the state file, and treat it exactly as `LEFTOVER` — an
   unverified cleanup is not a verified one. The exit code is checked as well as the line, because a
   caller that greps for `COMPLETE` in empty output finds nothing.

   **None of this is a cue to bring the stack back up.** Once check 5 in **Worktree cleanup** below
   has stopped the project's declared stack, every later run-2 step that touches it — a
   reported-and-continued removal failure at step 5 above, or a `SKIPPED:` clause on a `COMPLETE:`
   line here — is that same stack's absence showing up again, correctly, one step later. Restarting
   it to make one of those steps succeed undoes what check 5 was for and answers a question this
   procedure never asked.

   **A leftover blocks the `FINISHED` write, and that is the whole point of having a verdict.**
   `FINISHED` is terminal: `/myflow-finish` stops at it and `/flow-status` does not list it, so a
   change written `FINISHED` over a known leftover has exactly one record of that leftover — the
   console line — which is the transcript-only record this pipeline refuses everywhere else. Left at
   `IN_PROGRESS` instead, the change stays listed, stays re-runnable, and the state file it already
   has is the durable record; no new field is invented to carry a fact the state itself carries.

   **Run 2 is re-entrant, which is what makes that safe.** Every step is remove-or-move *if present*
   and a step whose artifact is already gone is success, not an error — so a re-run after the
   operator clears the leftover repeats the verification and nothing else. An already-archived change
   directory means step 3 is already done: the archive move is skipped, not repeated, and the run
   continues to cleanup and verification.

   **When the script is absent** — a repository that does not carry it — check the same registry rows
   by hand, in the same order, and say in the handoff that the verification was done manually. The
   check is never skipped for want of the script, and "not verified" is never reported as verified.
8. **Write `FINISHED`**, clearing from `worktrees` **only the entries whose removal actually
   succeeded** — see **Worktree cleanup**
   (`skills/flow-contracts/finish-contract-run2.md`) below — and carry every other field
   forward. This step is reached only on `COMPLETE:`.
9. **Run self-review** — after `FINISHED` is written; a skip, a failure, or a decline never moves
   the change off `FINISHED`. It is skippable per run, with running it the default. It gathers its
   input by invoking `gather-self-review-context.sh` rather than having the reasoning pass
   re-read files inline, and runs **one** combined reasoning pass covering all five angles below,
   together with the operator's 1-5 rating, never as five separate dispatches. **The reasoning pass
   itself runs as a subagent, on the harness-wide `selfReviewModel` setting** (empty inherits
   `defaultModel` — **Model resolution**, `skills/flow/SKILL.md`) rather than inline in the
   session driving `/flow` — the operator's rating and the per-angle filing prompts still run in
   that session, since only it can drive `AskUserQuestion`.

   | # | Angle | Label |
   |---|-------|-------|
   | 1 | Problems encountered, and what pipeline change would avoid them | `flow-fix` |
   | 2 | Token/time cost, and what would reduce it without quality loss | `flow-cost` |
   | 3 | What went well, and how to reproduce it | `flow-improvement` |
   | 4 | What could be automated or moved to a script | `flow-automation` |
   | 5 | What could move to the Go app or its persistent storage | `flow-stats-app` |

   **These labels were renamed from a `myflow-` prefix, and the old names are still in use — on the
   board and in this repository's own history.** Existing Jira issues keep the labels they were filed
   with; nothing relabels them, because that is an outward-facing bulk write over a board this
   pipeline does not own. **So the board carries both taxonomies indefinitely, and a label-based
   search has to match either form.** A reader who searches `flow-improvement`, finds a handful of
   issues and concludes the angle is rarely used has been misled by this table, not by the board.

   The self-review reports under `<project>/docs/self-review/` are the same story one layer down:
   every report written before the rename carries the old labels, and those reports are immutable
   records rather than files to migrate. `check-self-review-report.sh` therefore recognises **both**
   spellings, positionally aligned, and its own header states why. Renaming the labels in that guard
   without the legacy set makes it report 166 violations across all twelve reports it checks —
   measured, not predicted.

   Angle 5's remit covers the records the pipeline writes to files today and the derivation work
   now done in Bash or by the agent — **not** what the SPA should display.

   **Every angle produces zero or more findings, and an angle that produces none says so
   explicitly** — present-but-empty, the way `## Decisions` and `## Open questions` already are,
   rather than omitted. A silent angle and a skipped angle are indistinguishable to a reader, which
   is how KAN-73's cost angle passed unnoticed while its section existed.

   **Every finding is explained in the message body before any prompt fires** — what was observed,
   what breaks, and what the fix would be. A prompt's option text cannot carry that explanation, so
   the prompt records the decision only: a filed issue is durable, and an explanation arriving
   afterward describes something the operator did not agree to.

   The filing ask is **one multi-select prompt per angle**, listing that angle's findings and
   defaulting to filing none of them, shape per **Operator prompts**
   (`skills/flow-contracts/operator-prompts.md`). A finding filed this way carries its angle's
   label on top of the set **Labels on issues the pipeline creates**
   (`skills/flow-contracts/jira-integration.md`) already defines.

   The report is committed onto `chore/archive-<name>` — asserting that branch rather than assuming
   it, and not pushed here; step 10 carries the push — to
   `<project>/docs/self-review/<name>-self-review.md`. It carries one section per angle,
   all five present; each finding is one line naming its angle's label, the finding, and its
   disposition — the issue key when filed, an explicit declined marker when not — and an angle with
   no findings carries an explicit none-marker instead of finding lines. **This procedure is
   canonical here.** Step 9 of `skills/flow/archive.md`'s own run 2 carries only what is
   specific to *executing* it: the script invocation and its arguments, the exact prompt wording,
   and the report-commit shell. It is not a second statement of this rule.

   **There is no requirements layer above this one; change this file.** The procedure was first
   written as **Requirement: Self-review runs only after FINISHED is written**
   (`<agents repo>/openspec/specs/myflow-self-review/spec.md`), alongside sibling requirements in that
   same file for context gathering, the combined pass, the per-angle filing ask, the rating and
   the report path. That capability was frozen with the rest of the `<agents repo>/openspec/` tree at the spectre
   cutover and not migrated, so it records where the rule came from and governs nothing: the
   statement above is both the requirement and the runtime source of the procedure.
10. **Push the archive branch and open its pull request.** Push `chore/archive-<name>`; open a pull
    request against `<base>` via a PR CLI when usable for the host, else print the forge's
    create-PR URL and ask whether it was opened — the same shape Run 1's pull-request route
    (`skills/flow-contracts/finish-contract-run1.md`) uses. This push carries both the archive commit and the self-review report from step 9, so
    there is no window in which the archive pull request merges while the report is still
    unwritten. Run 2 never pushes to `<base>`.

    A failed push, or a failed pull-request creation, is reported with the command's own output. It
    never moves the change off `FINISHED` — the change is already terminal by step 8 — and the
    handoff names the unpushed branch `chore/archive-<name>` and prints the exact push and
    pull-request commands needed to land it by hand. The archive is never reported as landed until
    this step actually lands it.
11. **Restore the main checkout to `<base>`.** Successful or not, run 2 leaves the main checkout on
    `<base>`, never on the archive branch.

**The Jira `Done` transition fires before step 9, not after it.** Per **Jira integration**
(`skills/flow-contracts/jira-integration.md`)'s own timing — the issue moves to `Done` after the
archive move and the state write — that transition has already happened by the time step 9 begins,
so self-review has nothing to delay: there is no Jira write left in run 2 for it to sit in front of.

### Worktree cleanup

Which worktrees those are, and the `git worktree list --porcelain` scan that finds them when
the state file's map is absent or empty, are **Resolving a change's worktrees**
(`skills/flow-contracts/finish-contract-run1.md`).

For each worktree, run **every** check below before removing anything:

```bash
# 1. no uncommitted tracked changes — must be empty
git -C "$WT" status --porcelain --untracked-files=no

# 2. no untracked files that git does not already ignore — must be empty.
#    `--others --exclude-standard` lists exactly the files `--force` would destroy and
#    `.gitignore` does NOT cover. This is the check that makes `--force` safe.
git -C "$WT" ls-files --others --exclude-standard

# 3. no commits that exist only here. Resolve `BASE` fresh for THIS worktree — never reused from
#    another worktree in the set. A multi-repo change has one `origin` and one default branch per
#    repository, so a name resolved against one worktree's `origin` can be the wrong ref, or absent
#    entirely, in another repository. Resolving it here, before this worktree's own removal below,
#    is what "before cleanup removes it" (run 2 step 5, above) means per worktree.
BASE="$(resolve-base-branch.sh "$WT")" || { echo "cannot resolve the base branch for $WT — stop and ask"; false; }
#    `@{upstream}` ERRORS when no upstream is configured, and an empty capture would read as
#    "nothing unpushed" — so resolve it explicitly and never let a failed lookup pass as success.
#
#    Step 1 already proved the branch is an ancestor of the base branch, which is STRICTLY
#    STRONGER evidence than "pushed to its own upstream": the commits are in the base branch.
#    Accept that first. Requiring the upstream regardless would lock out the ordinary
#    squash-merge workflow — GitHub's "delete head branch on merge" plus `fetch.prune=true`
#    removes the tracking ref, after which no upstream can ever resolve and the branch cannot be
#    re-pushed because it no longer exists on the remote.
if git -C "$WT" merge-base --is-ancestor HEAD "origin/$BASE" 2>/dev/null; then
  :                                            # already merged into base — nothing can be lost
elif UP="$(git -C "$WT" rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null)"; then
  git -C "$WT" log --oneline "$UP..HEAD"       # must be empty
else
  echo "not merged, and no upstream — cannot prove these commits exist anywhere else"; false
fi

# 4. what `--force` WILL destroy: ignored files. This does not gate removal — it is shown to
#    the operator, who decides. `--exclude-standard` in check 2 hides everything matched by
#    .gitignore, <project>/.git/info/exclude or the global excludes file, and "ignored" is NOT "disposable":
#    a deliberately-ignored .env, a local override config, or this pipeline's own
#    <abs-worktree>/.superpowers/sdd/ records are all ignored and all irreplaceable.
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
set before it runs.** `check-worktree-processes.sh`'s own header treats a process whose working
directory is at or under the worktree as held, and a shell that is itself `cd`'d into `$WT` (or into
another worktree in the same set) is exactly such a process — it would report `HELD:` against
itself, not against the stack this check exists to catch. `cd` out first, for every worktree, before
this check runs for any of them.

**Check 6 is a gate, and both of its bad outcomes are failures.** `HELD:` names each process
holding the worktree, and exit 2 says the guard could not answer at all — the scanning tool is
absent, or the worktree path is not readable. Neither is a pass: an inability that proceeded to
removal would be the failure this check exists to prevent, arrived at more quietly. On either, stop
at `IN_PROGRESS`, leave every worktree alone, and report each pid, its working directory, and the
command that reaches it:

```bash
ps -o pid,command -p <pid>
```

**There is no confirmation to proceed past check 6**, unlike check 4's disclosure. That disclosure
is safe to confirm because the operator can see what is at stake and decide; a live process is
different in kind. Confirming it destroys the only records that can reach the process afterwards,
and the resulting orphan holds ports shared across every workspace — which is how this incident
reached the operator both times it happened. The remedy is to clear the process and re-run, while
the worktree still exists and the project's own stop command can still read what it started.

**Check 4 is a disclosure, not a gate.** When it lists anything, **stop and show the list**, and
ask for explicit confirmation before removing that worktree. Do not try to classify the entries as
build output: no allowlist of names can be trusted, because the operator decides what they ignore.
Empty list → proceed without asking.

**`/myflow-fast` overrides the ask, and only the ask.** Its own **Guardrails**
(`skills/myflow-fast/SKILL.md`) state that override and why it is safe there: that command reports
what `--force` will destroy and proceeds, having already preserved and committed the records worth
keeping before it reaches cleanup. The override is named here too, so two files cannot silently
disagree about a step that destroys files. It reaches nothing else — checks 1, 2, 3, 5 and 6 stay
gates under every command, and an irreplaceable **unpreserved** entry still stops the run and asks,
under `/myflow-fast` as much as here.

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
  prevent it: it lists exactly what will die, and the operator confirms. Claiming the checks make
  `--force` safe would be false, and was: a gitignored `.env` passes checks 1 and 2 and is
  destroyed silently.
- **Neither check sees a file whose `assume-unchanged` bit is set.** `git status` is blind to it
  by design. Rare, operator-inflicted, and named here so it is a known limit rather than a
  surprise.
- **`git branch -d`, never `-D`.** It must be free to refuse an unmerged branch.
- **An already-removed worktree is success**, not an error.
- **Any failed check leaves every worktree alone** and reports why. There is no partial cleanup.
  This includes check 5: a `## stop` command that **exits non-zero, is not found, or has to be
  interrupted** is a *failed* check, not an absent one — only an undeclared key is skipped. Give it
  a bounded wait rather than letting it hang the run.
- **Verify each removal actually succeeded** before writing state. If any `git worktree remove`
  fails for a reason the checks did not predict — a file lock, a permission error — report it and
  leave that worktree's entry in `worktrees`. Writing `worktrees: {}` regardless would drop it from
  the only authoritative list, and nothing would ever find it again.

Then the change's **remote** branch:

```bash
# `push --delete` exits non-zero BOTH when the branch was already gone and when the push was
# refused, so the two are told apart by git's message and never by the exit code alone. Measured
# against a scratch remote on this machine (git 2.50.1): an already-absent branch prints
# `error: unable to delete 'spectre/<name>': remote ref does not exist` and exits 1, and the
# stale remote-tracking ref SURVIVES that failure.
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

- **The remote branch is deleted without a further prompt.** Run 2 is reached only by proving the
  branch is an ancestor of the base branch, so its commits are in the base branch and nothing can be
  lost — which is why this is not gated the way check 4's disclosure is.
- **An already-absent remote branch is success**, not an error, and the outcome is reported either
  way: deleted, already gone, or refused.
- **A refused push is reported, never swallowed.** A bare `|| true` would make an expired
  credential indistinguishable from a branch the forge already removed, and leave the remote branch
  standing with nothing said about it.
- **The remote delete is not gated on the local one succeeding.** Gating it would leave the remote
  branch behind whenever anything unrelated failed, which is the state this step exists to end.

The stack-stopped check reads the optional `## stop` key from `<project>/.flow/project.md` —
see **Project configuration** in `skills/flow-contracts/project-configuration.md`. When the key
or the file is absent the check is **skipped, not failed**, and cleanup proceeds on the strength of
the other checks — check 6 among them, which is what makes an undeclared `## stop` key survivable:
a project that declares nothing to stop is still checked for a process running from its worktree.

