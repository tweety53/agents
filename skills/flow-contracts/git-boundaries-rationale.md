# Git boundaries — rationale

This file is the reasoning behind `skills/flow-contracts/git-boundaries.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## Git boundaries

**Why `/flow`'s implement phase commits once a PR exists and not before.** A PR is a remote surface, so a
staged-only fix would be invisible on it. Before that the operator reviews a staged diff in the
IDE, and committing would take that away.

**The commit is guarded, and an empty one is skipped rather than failed.** `git commit` exits
non-zero when nothing is staged, so an unguarded commit dead-ends on two ordinary cases: a fix
touching only the gitignored change directory leaves the implementation commit empty — which is
exactly what run 1's unfinished-work **Stop** course invites — and a re-run after a rejected push
finds the commit already made. The commit is therefore preceded by a staged-changes test, and the
whole sequence is one `&&` chain:

**A skipped commit is reported, and a FAILED commit stops the sequence.** Those are different
outcomes: "nothing to commit" is normal and costs one line in the handoff, while a commit a hook
rejects is a git failure and gets the standard treatment — report git's own output and stop. The
chain is what enforces the second, and it is a chain rather than `set -e` deliberately: bash before
4.0 ignores `set -e` inside a subshell whose parent has errexit off, and this block runs through an
agent's shell whose state it does not control. Run it as one command.

**Why a commit a run instructs defaults to the pathspec-scoped form — the two incidents.** Moved
verbatim from the contract, where it sat between "A plain commit takes the whole staged tree with
it" and "so wherever a run's instructions know the paths a commit should carry, the commit names
them": "— gymie kan-469's visual-verify step swept a ~130-file pre-staged foreign tree into a
baselines commit in a main checkout, and gymie kan-468 lost task commits to staged planning
artifacts the same way —".

## Moved by the 2026-11-13 prompt audit — incident records

Each passage below sat inline in a run-loaded file beside the rule it motivated; the rule and its
in-sentence reason stay there. The quoted fragment before each dash is the end of that rule, for
attribution; the text after it is the passage, verbatim.

### git-boundaries.md — Commits outside every task

- *judged there — run 1's unfinished-work gate lists it and handles `UNTASKED`* — (kan-941 pushed four such commits before either check existed)
- *and no `chore` subject, and its pathspec is not limited to `<project>/<spec-root>/changes/`* — (kan-1003: the design handoff 0497d943 was committed and pushed in one command, and run 1's gate listed it)

### git-boundaries.md — Revert and back-out commits

- *with the artifact instead of re-deriving the rejected position from indirect evidence* — KAN-703 is the case that paid for it: a run whose design space held a reverted prototype re-derived the prototype's intended polarity from contract wording because the revert's body recorded only `This reverts commit <sha>` — that the code left, never why.

### git-boundaries.md — Revert and back-out commits

- *back it out as two commits, the back-out under the paragraph above* — KAN-539 is the case that paid for it: a guessed icon and its revert were folded into one commit whose message still described the icon its tree did not contain.
