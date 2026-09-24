# Git boundaries — rationale

This file is the reasoning behind `skills/flow-contracts/git-boundaries.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## Git boundaries

**Why `/flow`'s implement phase commits once a PR exists and not before.** A PR is a remote surface, so a
staged-only fix would be invisible on it. Before that the operator reviews a staged diff in the
IDE, and committing would take that away.

**Both commits are guarded, and an empty one is skipped rather than failed.** `git commit` exits
non-zero when nothing is staged, so an unguarded two-commit sequence dead-ends on three ordinary
cases: a fix touching only the two planning paths leaves the implementation commit empty — which
is exactly what run 1's unfinished-work **Stop** course invites — a fix touching only implementation
leaves the planning commit empty, and a re-run after a rejected push finds both commits already
made. Each commit is therefore preceded by a staged-changes test, and the whole sequence is one
`&&` chain:

**A skipped commit is reported, and a FAILED commit stops the sequence.** Those are different
outcomes: "nothing to commit" is normal and costs one line in the handoff, while a commit a hook
rejects is a git failure and gets the standard treatment — report git's own output and stop. The
chain is what enforces the second, and it is a chain rather than `set -e` deliberately: bash before
4.0 ignores `set -e` inside a subshell whose parent has errexit off, and this block runs through an
agent's shell whose state it does not control. Run it as one command. Without that, a first commit
a hook rejects falls through to the unconstrained second `add`, and the sole resulting commit —
titled `chore(...)` — carries the implementation, silently breaking the very split this section
exists to enforce.

**Why a commit a run instructs defaults to the pathspec-scoped form — the two incidents.** Moved
verbatim from the contract, where it sat between "A plain commit takes the whole staged tree with
it" and "so wherever a run's instructions know the paths a commit should carry, the commit names
them": "— gymie kan-469's visual-verify step swept a ~130-file pre-staged foreign tree into a
baselines commit in a main checkout, and gymie kan-468 lost task commits to staged planning
artifacts the same way —".

**Why planning artifacts are committed at the plan gate rather than held uncommitted until
integrate.** An uncommitted change folder blocked every cross-repo change: `spectre link` refuses
while the canonical change directory carries uncommitted modifications, so each link needed an
operator-approved `--force`, and each link rewrote `link.md`, dirtying the folder again for the
next one. Held uncommitted, the plan was also the one copy a reviewer's `git checkout` could
destroy. Committing the folder once the operator has approved it — and again at each later
boundary where the run itself edits it — removes both, while the task-commit check keeps planning
out of every implementation commit, and integrate's reshape still lands a single planning commit.

**Why a location guard on planning commits.** `spectre link` must run from a repository's primary
checkout to resolve its peers file, so the directory a link commit is made from is one `cd` away
from the landing target. A link commit there lands change-folder content on the base branch, or on
whatever branch the main checkout holds, where no reshape, review or archive step of the change
sees it. `check-planning-commit-location.sh` refuses any planning commit outside the change's own
worktree and branch; `hooks/protect-main-checkout.py` covers only a main checkout on its default
branch, and only under a harness that registers it.
