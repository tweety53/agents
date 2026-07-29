# KAN-19 — finish safety and session records

**Change:** `kan-19-finish-safety-and-records`
**Issue:** KAN-19 — "KAN-14 follow-up — widen the provenance guard's scope, and make the model
record durable"
**Date:** 2026-07-29

## Scope

KAN-19 records four items across two subsystems that share no file and no requirement:

| Item | Subsystem |
|------|-----------|
| 1. Widen the provenance guard's scan scope beyond `tasks.md` | the provenance guard |
| 2. Make the SDD ledger's model record durable | `/myflow-finish` |
| 3. `/myflow-finish` treats a zero-commit branch as merged | `/myflow-finish` |
| 4. Nothing removes the proposal artifact source at finish | `/myflow-finish` |

**This change is items 2, 3 and 4.** Item 1 is a separate change, brainstormed on its own, because
it touches a different tree entirely — `scripts/check-plan-provenance.py`, its harness, and
`myflow-plan-provenance` — and because item 3 destroys work and should not wait behind a parser
rewrite.

Item 3 is the reason for the ordering. On KAN-14 it very nearly archived and `--force`-deleted a
worktree holding 29 staged entries.

## The problem

### Item 3 — a confident wrong answer where an unknown was available

`pipeline.md`'s Finish contract decides which run happens with
`git merge-base --is-ancestor HEAD origin/$BASE`. On a branch whose work is **staged but never
committed**, `HEAD` is still the merge base — so it is trivially an ancestor of the base branch and
the check reports *merged*. Run 2 then verifies the merge (falsely), syncs delta specs, archives
the change, pushes the archive, and removes the worktree with `--force`.

The contract already guards a *different* version of this failure — never resolve the base branch
from `HEAD@{upstream}`, because inside the apply worktree that compares the branch with itself.
That guard worked on KAN-14: the base resolved to `main`, correctly, and was asserted distinct from
the current branch. The hole is elsewhere and unguarded: **a branch with no commits of its own is
an ancestor of every branch.**

The ticket's first proposed fix — require `git rev-list --count origin/$BASE..HEAD` to exceed zero
— is recorded there as **wrong and not to be implemented**. It was disproved on the next run: once
the branch was genuinely merged it was *also* zero commits ahead, because its commit had become
part of the base branch. Both the dangerous state and the correct terminal state produce zero, so
that test trades a rare destructive failure for a permanent one that blocks every finish.

### Items 2 and 4 — artifacts the pipeline creates and never accounts for

The SDD ledger records which model ran each subagent dispatch. It lives at
`.superpowers/sdd/tasks/progress.md`, `.gitignore` excludes `.superpowers/`, and run 2 removes the
worktree with `--force`. `myflow-model-policy` states this limitation plainly rather than claiming
an audit trail it cannot provide, and defers durability to whichever change owns the archive step.
This is that change.

Separately, `/myflow-start` writes the published proposal's HTML source to
`<state-dir>/<name>-proposal-artifact.html` so a revision round can republish to the same URL.
Nothing in the finish contract accounts for it. Across three finished changes the outcome depended
on whether someone happened to notice: `kan-8` and `kan-10` had it removed by hand, `kan-14` did
not.

## Decisions taken during brainstorming

1. **Split the ticket**, finish-safety first, guard scope second — over one combined change, and
   over a minimal item-3-only hotfix.
2. **A missing recorded merge base refuses and asks**, over falling back to the clean-tree test
   alone and over refusing with no override.
3. **Run 1 copies the session records**, over run 2 (the ticket's suggestion) and over `/myflow-do`
   authoring them in-repo from the start.
4. **Preserve the ledger and the review panel record**, over the ledger alone and over everything
   under `.superpowers/sdd/`.
5. **Preserve the artifact source in-repo at run 1 and delete the state-dir copy at run 2**, over
   deleting with no preservation and over retaining it where it is.
6. **Item 3 becomes an executable, tested script**, over prose contract only and over a script for
   the decision with prose for everything else.

Rationales are recorded per decision in the change's `design.md`.

## Design

### Item 3 — an executable run-1/run-2 decision

**New: `scripts/check-finish-preflight.sh`.** Positional arguments: worktree path, resolved base
ref, recorded merge base (or `-` when absent). It prints exactly one verdict word on stdout —
`RUN1`, `RUN2` or `REFUSE`, followed by a reason — and exits `0` whenever it reached a verdict, `2`
on environment failure.

The verdict, not the exit status, carries the answer. A "cannot determine" outcome must not be
indistinguishable from a violation, which is what an exit-code-only protocol would make it.

Ordered checks:

| # | Condition | Verdict |
|---|-----------|---------|
| a | recorded merge base is `-` or empty | `REFUSE` — no recorded merge base |
| b | `HEAD` equals the recorded merge base | `RUN1` — the branch has no commits of its own |
| c | `HEAD` is not an ancestor of the base ref | `RUN1` — not merged |
| d | the worktree has uncommitted tracked changes or untracked-unignored files | `REFUSE` — merged by ancestry, but N entries are uncommitted |
| e | otherwise | `RUN2` |

**Check b runs before check c, and that ordering is the fix.** A branch with no commits of its own
is an ancestor of everything, so c answers "merged" on precisely the most dangerous input.

**Check d runs after check c**, because dirty-and-unmerged is the ordinary in-flight state and must
stay `RUN1` rather than becoming a refusal.

Both sides of check b are normalised through `git rev-parse --verify <ref>^{commit}`, so a
shortened sha in the state file compares correctly against a full `HEAD`.

Base-branch resolution stays where it already is. The script receives the resolved base as an
argument rather than resolving it, so the existing and working `HEAD@{upstream}` guard is not
duplicated into a second copy that can drift.

On a `REFUSE`, run 2 stops before touching anything, reports `HEAD`, the base, and the uncommitted
count, and asks the operator for explicit confirmation. Per decision 2 this is the honest-unknown
path: a change whose state file predates this fix, or whose `worktrees` map was cleared by hand,
still has an in-band route forward that does not involve guessing.

On a multi-repo change the skill runs the script once per `worktrees` key and proceeds to run 2
only if **every** worktree returns `RUN2`.

Run 2's existing cleanup checks 1 and 2 would already have caught KAN-14's 29 staged entries. They
run at step 4 — after the irreversible archive-and-push at step 3. Check d is those same checks
hoisted to a precondition.

### Items 2 and 4 — preserving the session records

**New: `scripts/preserve-session-records.sh`**, invoked from `/myflow-finish` run 1 before its
`git add -A`, and from `/myflow-do`'s commit path — the `prUrl` case, the one place `/myflow-do` is
allowed to commit — so a fix round raised after a PR is open refreshes the records rather than
leaving them a round stale.

| Source | Destination |
|--------|-------------|
| `.superpowers/sdd/tasks/progress.md` | `docs/superpowers/ledgers/<date>-<name>.md` |
| `.superpowers/sdd/final-review-panel.md` | `docs/superpowers/reviews/<date>-<name>-panel.md` |
| `<state-dir>/<name>-proposal-artifact.html` | `docs/superpowers/artifacts/<date>-<name>.html` |

`<date>` is fixed at the first copy. The script globs each destination for an existing `*-<name>.*`
and reuses that path when one is found, so re-copies overwrite in place instead of accumulating one
dated duplicate per fix round.

A missing source is reported in one line and is never a failure. A change may legitimately have no
panel record, and a preservation step that can block an integration would be a worse failure than
the gap it closes.

Both destination shapes match what was preserved by hand on the two prior changes:
`docs/superpowers/ledgers/2026-07-29-kan-14-plan-provenance.md` and
`docs/superpowers/reviews/2026-07-29-kan-14-plan-provenance-panel.md`.

**What is deliberately not preserved.** The remaining files under `.superpowers/sdd/` — fix-wave
briefs, per-task diffs, per-task reports — are let go, as they were by hand on KAN-14. The per-task
diffs duplicate commits that already exist in git history.

**The `unknown (agent-defined)` rule survives untouched.** Panel slots dispatched by
`subagent_type` resolve their model from their own agent definition, which the dispatcher never
reads. Durability must not create pressure to fill those entries in with a plausible value nothing
measured — that is the failure the whole model-record requirement exists to police.

Run 2 then removes `<state-dir>/<name>-proposal-artifact.html`, in the same disclosed sequence as
the worktree removal. The recorded `artifactUrl` stays republishable, because the source now lives
in git history.

### Path provenance

`.superpowers/sdd/final-review-panel.md` is verified — `skills/myflow-do/SKILL.md:192`.

`.superpowers/sdd/tasks/progress.md` is taken from `myflow-model-policy`'s own prose and from the
issue text; no worktree exists to check it against while planning. The plan tags it for
confirmation at implementation time rather than asserting it.

## Spec deltas

- **`myflow-finish-cleanup`**
  - MODIFIED — *Run 2 verifies the merge before changing anything*: the three-signal decision, the
    ordering that makes it correct, and the refusal path.
  - ADDED — *Session records are preserved in the repository*.
  - ADDED — *Run 2 removes the proposal artifact source*.
- **`myflow-model-policy`**
  - MODIFIED — the third requirement's scope paragraph and its *A reader asks which model
    implemented a given task* scenario, both of which currently state the record is gone once the
    change is archived. The `unknown (agent-defined)` requirement text is left byte-for-byte alone.
- **`agents-repo-verification`**
  - MODIFIED — the requirement enumerating `## test`, which gains
    `scripts/test-check-finish-preflight.sh` and `scripts/test-preserve-session-records.sh`.
    Neither new check joins `## lint`: both need a git branch context rather than a tree scan, and
    a lint step that cannot run outside a change would fire on every unrelated invocation.

Contract and skill text follows the specs. `pipeline.md`'s Finish contract points at the preflight
script; `myflow-finish/SKILL.md` gains the copy step in section 1.2 and the artifact removal in run
2, pointing at `pipeline.md` rather than restating it, per its own standing guardrail.

## Testing

`scripts/test-check-finish-preflight.sh` builds throwaway git repositories per case:

- zero-commit branch, staged work present — the KAN-14 shape; must be `RUN1`
- genuinely merged, clean tree — `RUN2`
- merged by ancestry, dirty tree — `REFUSE`
- unmerged, dirty tree — `RUN1`, not a refusal
- absent recorded merge base — `REFUSE`
- shortened sha as the recorded merge base, genuinely merged — `RUN2`

`scripts/test-preserve-session-records.sh` covers the first copy, a re-copy overwriting the same
dated path, and each of the three sources missing independently.

Both harnesses follow the standing warning in `scripts/test-check-plan-provenance.sh`'s header:
assert against the stated contract, never against observed output. That suite four times encoded
one of the guard's own defects as its specification.

## Out of scope

- Item 1, the provenance guard's scan scope — a separate change under this ticket.
- Commit messages, which the issue notes are outside any guard's reach.
- KAN-15, serialization of independent tasks in `/myflow-do`.
