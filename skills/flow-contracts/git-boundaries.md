# Git boundaries

Which git actions each command may take and the branch backup. The guarded commit chain
that commits the implementation and leaves the change directory out is
`skills/flow-contracts/git-boundaries-commit-chain.md`.

**Loaded by `/flow`'s creating run and implement phase, bare `/flow` and `/flow-plan`** — at the step that commits or stages.

This file is **canonical** for everything in it.

The reasoning behind this file lives in `docs/rationale/flow-contracts/git-boundaries.md`.

## Git boundaries

| Command | Condition | Allowed git actions |
|---------|-----------|---------------------|
| `/flow`'s creating run | — | **Creates the worktree and pushes its empty branch** at kickoff, and nothing else |
| `/flow`'s implement phase | from `STARTED` | Resume the kickoff worktree + **commits each task** (fixups fold in), **pushing after each** — no merge or PR |
| `/flow`'s implement phase | at `IN_PROGRESS`, no `prUrl` | Resume **existing** worktree + **commits fixups** the same way, pushing after each — no merge or PR |
| `/flow`'s implement phase | at `IN_PROGRESS`, `prUrl` recorded | **Commits the implementation and pushes `--force-with-lease`** to the PR branch |
| bare `/flow` | run 1 | **Reshapes** — folds every commit since the merge base — then **commits** the implementation, then each self-review fix, each its own commit, all on `spectre/<name>` in the apply worktree; pushes `--force-with-lease` and opens a PR, or fast-forwards `<base>` to it with one push, by the operator's choice (**The routes**, `skills/flow-contracts/finish-contract-run1.md`) |
| bare `/flow` | run 2 | **Commits nothing and pushes no commit**: removes worktrees and the local and remote branch, then fast-forwards the main checkout (step 5 of **Run 2 — the branch is merged**, `skills/flow-contracts/finish-contract-run2.md`) |
| `/flow-plan` | change captured | **Creates the change worktree and pushes its empty branch** through the kickoff (**Capturing a new change**, `skills/flow-plan/SKILL.md`); commits nothing, ever |

**No command writes the main checkout.** `/flow` creates `<project>-worktrees/<name>` inside
its kickoff, `/flow-plan` through that same kickoff at
capture — it reads the main checkout before then and writes nothing — and every write, stage,
commit and push in the table above happens in a worktree. The main checkout is never checked out,
staged, committed or written, whatever branch it sits on — save run 2's fast-forward of it, which
`flow-guard refresh-main-checkout` refuses whenever it could lose anything.

## The change directory is never committed

**`<project>/spectre/changes/` is gitignored, and nothing under it is ever staged or committed.** A
change's directory lives on disk in its worktree only (**A change's directory**,
`skills/flow-contracts/pipeline.md`).

## Branch backup

A change's branch lives on the remote from the moment its worktree exists: `git worktree add` is
followed by `git push -u origin <branch>`, and every commit a run makes on the branch — a fixup,
a panel fix — is followed by `git push origin <branch>` from the same parent
call that made or verified it. Two pushes are deferred: a task commit `flow-guard close-task`
closed is pushed once at `flow.implement`'s end, and before any stop or handoff inside it (**Push
at the stage boundary**, `skills/flow/implement.md`); a visual-verify commit in a `regression checkout` worktree
(`skills/flow/visual-verify.md`) is carried by integrate's push below. A lost worktree is then rebuilt with `git worktree add <path>
origin/<branch>`. The one rewrite is integrate's reshape (`git reset --soft`, then
the implementation commit), so that run's push is `git push --force-with-lease origin <branch>`, made
before the self-review pass, so a run that stops before its route still
leaves the remote on the rewritten branch; every other push is plain, save a branch a base-sync
rebase rewrote, which pushes `--force-with-lease` (the base-sync step of **1. Load context and
validate the plan**, `skills/flow/implement.md`).

**The planning path** is the one **Handoff output** (`skills/flow-contracts/pipeline.md`) names.
`/flow`'s implement phase clears it from the index and only then stages with it excluded by
pathspec — an exclusion governs what an
`add` adds and cannot retract what an earlier step staged, so the clearing pass is what makes the
rule hold rather than merely assert it. Its staging area therefore carries implementation only.

**`link.md` is never committed either.** It lives in the gitignored
`<project>/spectre/changes/<name>/`, on disk in each worktree, like the rest of the change
directory.

`git add -A` respects `<project>/.gitignore`. Never force-add.

**Every `git add` and `git commit` a run instructs runs as `git -C <abs-worktree>`, written out or not** — a pathspec read against a subdirectory the shell sits in misses the paths it names (`skills/flow/implement.md`).
**A commit a run instructs defaults to the pathspec-scoped form.** `git commit -m "<subject>" -- <paths>`:
the commit carries only the paths it names, whatever else the index holds. A plain commit takes
the whole staged tree with it, so wherever a run's instructions know the paths a commit should carry,
the commit names them. The guarded commit chain (**The guarded commit chain**,
`skills/flow-contracts/git-boundaries-commit-chain.md`) is the deliberate exception: its
`add -A` is what picks up operator edits and new files at the human gate, which a
pathspec-scoped commit would drop.

## Commits outside every task

**Commits outside every task carry a `Flow-Key:` trailer.** A commit a run makes on a change
branch outside every task — the parent's full-suite fix, a fix's implementer (`fix-<k>-implementer`), a panel fix
committed on top, the `minor-fix` subagent's review-Minors commit, `flow.verify`'s capture
commit, the brainstorm's handoff-asset commit, integrate's implementation commit —
carries `-m "Flow-Key: <key>"`, `<key>` the dispatch key its round is recorded under
(`full-suite-fix-<n>`, `panel-fix-<round>`) or, where the stage
records no such row, the stage key (`flow.review`, `flow.verify`, `flow.plan`,
`flow.land`). A task's own commit
carries `Task-Id:` instead.

## Restoring a path with unstaged edits

**A restore that would discard unstaged edits is denied.** `flow-guard hook block-unsafe-commands` denies
a typed `git checkout … -- <path>` or `git restore <path>` in a linked worktree when a path it names
carries unstaged edits: the restore reverts the file to its index state, and an uncommitted fix on
it is lost with the flip a mutation proof made. Restore a flipped file
with `flow-guard break-and-prove`, which restores from a pre-mutation byte
snapshot (**Panel re-runs**, `skills/flow/review-panel.md`); `git restore --staged`, and a path
with no unstaged edits, are restored as before.

## Revert and back-out commits

**A back-out's commit body carries the rejection reason.** When a landed-then-reverted prototype
or a withdrawn task is backed out, the revert/back-out commit's body states what was wrong and
what would have to change for the design to come back, so a later run inherits the verdict with
the artifact instead of re-deriving the rejected position from indirect evidence.

**A commit's message describes the tree it lands, never an attempt that left it.** When work
written inside a task or a fix round is undone before it lands — reverted in the working tree, or
removed by a fixup folded into the commit — the commit's subject and body describe the net tree,
and the body names the abandoned attempt only as abandoned, with why. To keep the attempt in the
history instead, land it and back it out as two commits, the back-out under the paragraph above.
