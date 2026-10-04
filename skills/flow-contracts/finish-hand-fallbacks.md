# Finish contract — hand fallbacks

**Loaded by `finish-contract-run1.md` and `finish-contract-run2.md` only when** the guard presence
check named one of their scripts missing, or a call finds the script absent. Each section is the
by-hand procedure for one script, moved verbatim from the contract that calls it; the script's
verdicts and exit contract stay there.

## `check-finish-preflight.sh` — Finish contract

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

## `check-foreign-staged.sh` and `check-main-checkout-drift.sh` — Surface foreign staged work before the preflight

**When the script is absent** — a repository that
does not carry it — read each main checkout by hand with
`git status --porcelain --untracked-files=no`, keep the lines whose first column is not a space,
together with any intent-to-add entry's ` A <path>` line,
ask the same question over what that finds, and say in the handoff that the surfacing was done
manually; it is never skipped for want of the script.

**When the drift script is absent** — a repository that does not carry it — read each main
checkout by hand in the same shape its header's hand-verification procedure gives: the current
branch against the branch `refs/remotes/origin/HEAD` names with its prefix stripped, and
`git status --porcelain --untracked-files=no`, asking the same question over what those find, and
say in the handoff that the surfacing was done manually; it is never skipped for want of the
script.

## `check-base-moved.sh` — Run 1 — the branch is not merged

**When the script is absent** — a harness whose repository does not carry it — reach the same three
verdicts by hand, in the same order, and say in the handoff that the check was run manually. The
check is never skipped for want of the script.

## `resolve-base-branch.sh` — Run 1 — the branch is not merged

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

## `commit-archive.sh` — Run 1, Archive on the change branch

The rendered ledger and panel record are preserved into this commit first — each
when present, an absent file copying nothing; `<canonical-worktree>` resolves as run 1 resolves
it (**Finish contract**, `skills/flow-contracts/finish-contract-run1.md`):

```bash
bash -c 'for pair in "ledgers/<name>.md ledger.md" "reviews/<name>-panel.md panel.md"; do
  set -- $pair
  [ -f "<canonical-worktree>/.superpowers/sdd/$1" ] \
    && cp "<canonical-worktree>/.superpowers/sdd/$1" \
          "<canonical-worktree>/spectre/changes/archive/<name>/$2"
done'
[ "$(git -C <canonical-worktree> branch --show-current)" = "spectre/<name>" ] \
  && git -C <canonical-worktree> add -A \
  && check-archive-scope.sh <canonical-worktree> "spectre/changes/" \
  && { git -C <canonical-worktree> diff --cached --quiet \
       || git -C <canonical-worktree> commit -m "chore(spectre): archive <name>"; }
```

A branch mismatch is reported, naming the branch found, and stops the commit, leaving the change
at `IN_PROGRESS`. The subject is the fixed literal shown, per **Commit scopes name the module**
(`<agents repo>/rules/commit-scope-is-the-module.mdc`). The self-review bundle (`flow
self-review bundle`) resolves this commit by matching that subject line whole —
**reproduce it exactly.**

## `check-archive-scope.sh` — Run 1, Archive on the change branch

**When absent**, run
`git -C <canonical-worktree> diff --cached --name-only` by hand and refuse any path outside
`<project>/spectre/changes/`, the prefix the archive commit's guard call passes.

## `check-cleanup-complete.sh` — Run 2, step 4

**When the script is absent** — a repository that does not carry it — check the same registry rows
by hand, in the same order, and say in the handoff that the verification was done manually. The
check is never skipped for want of the script, and "not verified" is never reported as verified.

**Load `skills/flow-contracts/project-configuration-isolation.md`** only when the script is absent.

## `remove-change-worktrees.sh` — Worktree cleanup

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
  them — and "ignored" is not "disposable". Check 4 exists to make that visible rather than to
  prevent it: it lists exactly what will die, and asks only about an irreplaceable, unpreserved entry. Claiming the checks make
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
