# Finish contract — hand fallbacks

**Loaded by `finish-contract-run1.md` and `finish-contract-run2.md` only when** the guard presence
check named one of their scripts missing, or a call finds the script absent. Each section is the
by-hand procedure for one script, moved verbatim from the contract that calls it; the script's
verdicts and exit contract stay there.

## `flow-guard check-finish-preflight` — Finish contract

1. **`HEAD` against the merge base recorded in the state file's `worktrees` map.** No recorded
   value, or one that does not resolve → `REFUSE`; an honest unknown is never inferred. Equal to
   `HEAD` means the branch has no commits of its own → `RUN1`, whatever the ancestor test says.
   Both are answered before the base ref is resolved, so no environmental failure can hide them.
2. **The ancestor test** — `git merge-base --is-ancestor`, and only that. The script consults no PR
   CLI: it must give the same answer on every forge, and git alone already answers this question.
   The base ref is resolved first, so a base ref that does not resolve is its own `REFUSE` rather
   than an accidental `RUN1`.
3. **The worktree's cleanliness.** Merged by ancestry with uncommitted entries → `REFUSE`.
   Merged, distinct from the recorded merge base, and clean → `RUN2`.

**Never substitute a commit count.** `git rev-list --count <base>..HEAD` is zero both for a branch
with no commits and for a branch whose commits have joined the base branch, so it cannot separate the
dangerous state from the correct terminal one, and using it would refuse every legitimate run 2.

**Signal 1 precedes signal 2, and that ordering is the point.** A branch with no commits of its own
is an ancestor of every branch, so the ancestor test alone reports *merged* on a branch whose work is
staged and never committed — after which run 2 `--force`-removes the worktree holding all of it.

**When the script is absent** — a harness whose repository does not carry it — perform the same three
signals by hand in the same order and say in the handoff that the check was run manually. The check is
never skipped for want of the script. Signal 2 may then be answered by a PR CLI when one is usable
for the host, as in run 2
(`skills/flow-contracts/finish-contract-run2.md`) — that option belongs to the human doing this by hand, never to the
script — but signals 1 and 3 still run, and still run in this order.

## `flow-guard check-main-checkout-drift` — Surface main-checkout drift before the preflight

**When the script is absent** — a repository that does not carry it — read each main
checkout by hand in the same shape its header's hand-verification procedure gives: the current
branch against the branch `refs/remotes/origin/HEAD` names with its prefix stripped, and
`git status --porcelain --untracked-files=no`, stopping over what those find, and
say in the handoff that the surfacing was done manually; it is never skipped for want of the
script. A checkout whose only finding is its branch is switched back by hand with
`git -C <main-checkout> switch <default>` under the same four exclusions, per **Surface main-checkout
drift before the preflight** (`skills/flow-contracts/finish-contract-run1.md`).

## `flow-guard check-base-moved` — Run 1 — the branch is not merged

`check-finish-preflight` prints its verdict after a `RUN1`. **When the script is absent** — a harness whose repository does not carry it — reach the same three
verdicts by hand, in the same order, and say in the handoff that the check was run manually. The
check is never skipped for want of the script.

## `flow-guard resolve-base-branch` — Run 1 — the branch is not merged

**When the script is absent** — a harness whose repository does not carry it — resolve the base
branch by hand, in the same order, and say in the handoff that the resolution was done manually. The
guard is never skipped for want of the script. Run the wrapped fetch first — bounded and
credential-free, exactly as the guard's own header describes, so an unreachable remote refuses
quickly rather than hanging — then read `refs/remotes/origin/HEAD`, falling back to `git remote show
origin`'s reported `HEAD branch` only when that is empty. Once a candidate name is in hand, apply the
guard's assertions in order and by hand, never accepting a guess in place of any of them: refuse a
detached `HEAD`, refuse when no name resolved, refuse when the resolved name equals the current
branch, and refuse the resolved name unless it matches this exact shape: the first character is one
of `[A-Za-z0-9._]` — so a leading `-` a downstream git call would read as an option is refused, and
so is a leading `/` — and every character in the name, start to end, is one of `[A-Za-z0-9._/-]`,
which is what rules out control characters and anything else outside that set.

## `flow-guard remove-change-worktrees` — Worktree cleanup

For each worktree, run **every** check below before removing anything:

```bash
# 1. no uncommitted tracked changes — must be empty
git -C "$WT" status --porcelain --untracked-files=no

# 2. no untracked files that git does not already ignore — must be empty.
#    `--others --exclude-standard` lists exactly the files `--force` would destroy and
#    `.gitignore` does NOT cover. It does not make `--force` safe: ignored files go with it.
git -C "$WT" ls-files --others --exclude-standard

# 3. no commits that exist only here. Resolve `BASE` fresh for THIS worktree — never reused from
#    another worktree in the set. A multi-repo change has one `origin` and one default branch per
#    repository, so a name resolved against one worktree's `origin` can be the wrong ref, or absent
#    entirely, in another repository. Resolve it here, before this worktree's own removal below.
BASE="$(flow-guard resolve-base-branch "$WT")" || { echo "cannot resolve the base branch for $WT — stop and ask"; false; }
#    `@{upstream}` ERRORS when no upstream is configured, and an empty capture would read as
#    "nothing unpushed" — so resolve it explicitly and never let a failed lookup pass as success.
if git -C "$WT" merge-base --is-ancestor HEAD "origin/$BASE" 2>/dev/null; then
  :                                            # already merged into base — nothing can be lost
elif UP="$(git -C "$WT" rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null)"; then
  git -C "$WT" log --oneline "$UP..HEAD"       # must be empty
else
  echo "not merged, and no upstream — cannot prove these commits exist anywhere else"; false
fi
```

Then, and only then:

```bash
git -C "$REPO" worktree remove --force "$WT"
git -C "$REPO" branch --set-upstream-to="origin/<base>" "spectre/<name>"
git -C "$REPO" branch -d "spectre/<name>"
git -C "$REPO" worktree prune
```

- **`--force` destroys every ignored file in the worktree, and no check prevents that.** Checks 1
  and 2 establish only that nothing *tracked-and-modified* and nothing *untracked-and-unignored*
  is at risk. They say nothing about ignored files, because `--exclude-standard` is what hides
  them — and "ignored" is not "disposable". Claiming the checks make
  `--force` safe would be false: a gitignored `.env` passes checks 1 and 2 and is
  destroyed silently.
- **`git branch -d`, never `-D`.** It must be free to refuse an unmerged branch.
- **The upstream is pointed at `origin/<base>` before `-d`.** `-d` judges against the branch's upstream, else the main checkout's `HEAD`; a forge merge deletes `origin/spectre/<name>`, and the main checkout is refreshed only after cleanup, so judged against either it refuses a merged branch. **A refused `-d` puts the previous upstream back** — the branch's previous `remote` and `merge` keys written back with `git -C "$REPO" config`, never `--set-upstream-to`, which refuses a pruned ref — or `--unset-upstream` when it had none — so a surviving branch never tracks `origin/<base>`.
- **An already-removed worktree is success**, not an error.

Then the change's **remote** branch:

```bash
OUT="$(git -C "$REPO" push origin --delete "spectre/<name>" 2>&1)"; RC=$?
if [ "$RC" -eq 0 ]; then
  echo "remote branch deleted: origin/spectre/<name>"
elif printf '%s' "$OUT" | grep -q 'remote ref does not exist'; then
  # The forge deleted it on merge. Prune the ref it left behind.
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
