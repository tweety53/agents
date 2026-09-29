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
dangerous state from the correct terminal one, and using it would refuse every legitimate archive.

**Signal 1 precedes signal 2, and that ordering is the point.** A branch with no commits of its own
is an ancestor of every branch, so the ancestor test alone reports *merged* on a branch whose work is
staged and never committed — after which run 2 archives the change and `--force`-removes the worktree
holding all of it.

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

## `classify-untracked.sh` and `prepare-archive-branch.sh` — Run 2, step 2

When the script is absent, classify by hand to the same three classes and say so in the handoff.

**When the script is absent** — a harness whose repository does not carry it — perform the same
positioning by hand, in this order, against `<landing-worktree>`, creating it via `git -C
<main-checkout> worktree add --force <landing-worktree> <base>` when it does not already exist
(`--force` because the main checkout is ordinarily already on `<base>` at this point, per
`check-finish-preflight.sh`'s own main-checkout assertion (**Finish contract**,
`skills/flow-contracts/finish-contract-run1.md`), and git otherwise refuses a second worktree on
a branch already checked out), and say in the handoff that it was done manually. The guard is never
skipped for want of the script.
Run the wrapped, credential-free fetch first, so an
unreachable remote refuses quickly rather than hanging. Then read `HEAD`: **refuse a detached
`HEAD`.** Then read the working tree with `git -C <landing-worktree> status --porcelain` — the
same test the preflight's signal 3 uses — and **refuse a dirty tree wherever it is found, on
`<base>` as well as off it**, naming both the branch found and `<base>`; uncommitted changes
would otherwise ride onto the archive branch unremarked. On a clean tree, check out `<base>` if
the checkout is not already on it, then fast-forward it to `origin/<base>`, **refusing a base
that cannot fast-forward** rather than merging or resetting it. Finally create
`chore/archive-<name>` from that base and check it out — or, when it already exists, **reuse it
only if it is descended from `origin/<base>`** and refuse it otherwise. Apply each refusal in
that order and never accept a guess in place of any of them.

## `check-archive-scope.sh` — Run 2, step 4

**When absent**, run
`git -C <landing-worktree> diff --cached --name-only` by hand and refuse any path outside
`<project>/spectre/changes/`, the prefix step 4's guard call passes.

## `check-cleanup-complete.sh` — Run 2, step 7

**When the script is absent** — a repository that does not carry it — check the same registry rows
by hand, in the same order, and say in the handoff that the verification was done manually. The
check is never skipped for want of the script, and "not verified" is never reported as verified.

**Load `skills/flow-contracts/project-configuration-isolation.md`** only when the script is absent.
