# Finish contract — rationale

This file is the reasoning behind `skills/flow-contracts/finish-contract-run1.md` and
`skills/flow-contracts/finish-contract-run2.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## finish-contract-run1.md — Surface foreign staged work before the preflight

Incident: KAN-546.

Moved verbatim from the contract, where it closed "a main checkout's staged residue is exactly what
a resumed run is tempted to clear by hand once a later REFUSE arrives": — the high-judgment surgery
gymie kan-437's run 2 performed inline across three repos, hard reset where it judged a clean
revert and stash where the work was distinct WIP.

## finish-contract-run1.md — Run 1 — the branch is not merged

Sync the branch onto the base — incident behind `aside-planning-artifacts.sh`: KAN-628.

Reshape, moved verbatim from the contract, where it followed "is the chain
**Git boundaries** (`git-boundaries.md`) gives": , and is not written out a second time here.

Resolve the base branch, moved verbatim from the end of the hand-fallback paragraph: That header is
the authority for the exact commands and their order; this paragraph does not repeat what it already
states. The character rule itself is stated here in full, not cited, because this paragraph's own
precondition is the script's absence — a fallback the operator applies by hand when the script is
absent cannot defer them to a file that, by the same precondition, is not there to read.

## finish-contract-run1.md — Resolving a change's worktrees

Moved verbatim from the contract, where it opened the sentence whose call-site list stays there:
"That rule and the commands it binds are not restated here; what follows is specific to bare
`/flow`:".

## finish-contract-run2.md — Run 2 — the branch is merged

Step 2, the pre-flight classifies — incident behind `classify-untracked.sh`: KAN-411.

Step 2, moved verbatim from the hand fallback, where it followed "The guard is never skipped for
want of the script.": The steps are stated in full rather than cited, for the reason the resolver's
own fallback above gives: the guard being absent takes its header with it.

Step 3, moved verbatim from the contract, where it closed "**There is nothing to sync into
`<project>/spectre/specs/` first**": , and no sync step is ever to be added back.

Step 4, moved verbatim from the contract, where it followed "rather than let a stray path land on
`chore/archive-<name>` unremarked.": This has happened: an archive commit made this way reverted
skill-file content another change had shipped minutes earlier.

Step 4, the archive-scope guard's cannot-answer exit — incident: KAN-601.

Step 9, moved verbatim from the contract, where it followed **This procedure is canonical here.**:
Step 9 of `skills/flow/archive.md`'s own run 2 carries only what is specific to *executing* it: the
script invocation and its arguments, the exact prompt wording, and the report-commit shell. It is
not a second statement of this rule.
