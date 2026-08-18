# Self-review — kan-73-install-guard-scripts-alongside-skills

**Date:** 2026-08-18
**Issue:** [KAN-73](https://tweety53.atlassian.net/browse/KAN-73)
**PRs:** [#10](https://github.com/tweety53/agents/pull/10) (implementation), [#11](https://github.com/tweety53/agents/pull/11) (archive)
**Operator rating:** 4 / 5 — good
**Command:** `/myflow-fast`, three invocations (create → integrate → archive)

One combined pass over four angles, per the finish contract. Context gathered by
`scripts/gather-self-review-context.sh` — 4 of 4 sources found.

---

## Problems, and the fixes filed for them

### P1 — the plan-declaration defect class · [KAN-193](https://tweety53.atlassian.net/browse/KAN-193)

Four separate fix rounds, every one a defect in the plan's own `Files:` / `Tests:` declarations and
none in any implementation:

1. Task 2's `Files:` declared four directories; the guard checks individual paths.
2. Task 2's `Tests:` named files belonging to tasks 5 and 6.
3. The correction to (1) still failed — prose plus a blank line after `**Files:**` terminates the
   field, so the bullet list beneath it was never parsed. The guard's message named the missing
   file, not the truncation, so the cause had to be inferred.
4. Tasks 3 and 4's `Tests:` named guard scripts as though those commits carried them.

A fifth surfaced in review: task 3's `Files:` listed three files its commit never touched.
`check-task-commit-fields.sh` cannot catch that direction at all — it checks that a commit touched
nothing undeclared, never that it touched everything declared.

**The structural point:** the guard runs against a commit, so it can only fire *after* a dispatch
has completed. Each of these cost a full implementer round trip to discover something a plan-time
parse would have caught in a second.

### P2 — the guard-to-skill map was derived by grep · [KAN-196](https://tweety53.atlassian.net/browse/KAN-196)

The map of which guards each skill ships was built by grepping skill text for `scripts/<name>.sh`,
which cannot separate an invocation from prose describing a guard. `myflow-status` was given a
directory it should not have; `check-unfinished-work.sh` was shipped to `myflow-do` on the same
evidence — and there the review reached the *opposite* verdict, that the guard genuinely is
reachable and the real defect was its absence from that command's presence-check list.

The classifier that does this correctly landed as **task 5 of this same change**. The plan needed
it and it did not exist yet.

Worth recording as a positive too: when task 5's shape-based classifier and task 3's hand
classification were compared, they agreed on every borderline case. Two independent derivations
agreeing is far better evidence than either alone.

### P3 — the SDD ledger was nearly destroyed · [KAN-192](https://tweety53.atlassian.net/browse/KAN-192)

`preserve-session-records.sh` reads the ledger from exactly `.superpowers/sdd/tasks/progress.md`.
Nothing in `myflow-do` — the skill that writes the ledger — states that path. This run's ledger went
to `.superpowers/sdd/ledger.md`, so the first preservation call emitted
`skipped: ... (absent)` — a **success** line — and run 2 `--force`-removes the worktree.

192 lines of dispatch history were one command from being gone. It was caught only because the
skipped lines were read rather than skimmed. The asymmetry is what makes it dangerous: that line is
indistinguishable from the legitimate case of a change with no ledger.

### P4 — two reviewers mutated the worktree · [KAN-194](https://tweety53.atlassian.net/browse/KAN-194)

A per-task reviewer ran `git clean` and destroyed the worktree's untracked planning artifacts. A
panel reviewer ran `git checkout <sha> -- .`, recovered with `git reset --hard`, and left a stash
behind. Both recovered — the first only because the main checkout held authoritative copies.

Nothing in the dispatch contract forbids a reviewer from writing; the prohibition was added by hand
after the first incident. And the second incident was known only because that agent chose to report
it. A prose promise is not a constraint; a read-only tool set would be.

### P5 — a stale local `main` produced a wrong preflight verdict · [KAN-195](https://tweety53.atlassian.net/browse/KAN-195)

With PR #10 already merged, the preflight returned `RUN1: HEAD is not an ancestor of main`. Correct
for the ref it was given — local `main` was 3 commits behind `origin/main` — but wrong about the
world. `git fetch` updates `origin/main`, not `main`, and the contract's base resolution deliberately
strips the remote prefix.

The run proceeded correctly only because the contradiction against the PR state was investigated
rather than trusted. Note the internal inconsistency this exposes: cleanup check 3 already compares
against `origin/$BASE`, while the preflight compares against the bare name.

---

## Cost

14 subagent dispatches; roughly 1.9M subagent tokens. Seven implementation commits, reshaped to two
at integration. Three `/myflow-fast` invocations.

Expensive, and mostly earned: the review panel surfaced 11 findings, two of them Major gaps in the
change's *own* new guard. The waste was concentrated in P1 — four round trips that a plan-time check
would have removed entirely.

## What went well

**Verification by mutation rather than by report.** Every guard claim was re-tested by breaking the
tree and confirming the guard noticed. This caught that F4 was real (deleting a `myflow-fast`
symlink left the guard reporting OK at exit 0), and it also caught two of *my own* readings as
artifacts — exit codes measured through a `head` pipe, reporting `head`'s status rather than the
guard's. Both times the instinct to re-measure rather than accept the first number was what mattered.

**The `light` roster's cheapest slot was the most productive.** Code review at effort `low` returned
3 Major and 7 Minor, every one with a runnable reproducer, including two gaps neither heavier slot
saw. A useful counterweight to assuming effort correlates with yield.

**Agents disagreeing was informative, not noise.** Task 1's reviewer and the Principles slot reached
opposite conclusions on whether `check-workspace-isolation.sh` should keep its inline `resolve_file`.
Surfacing both, with evidence, and handing the judgement to the fix round produced a better-reasoned
outcome than either verdict alone.

**Honest reporting from agents.** The fix round said plainly that F10's `rc >= 2` path had no
black-box reproducer and that its fixture passes both before and after, rather than dressing it up
as pinned. That is the behaviour worth reinforcing.

## Automation candidates

| Candidate | Filed |
|---|---|
| Plan-time linter for `Files:` / `Tests:` fields | [KAN-193](https://tweety53.atlassian.net/browse/KAN-193) |
| Ledger path stated where it is written, or accepted by the preserver | [KAN-192](https://tweety53.atlassian.net/browse/KAN-192) |
| Read-only tool set for reviewer dispatches | [KAN-194](https://tweety53.atlassian.net/browse/KAN-194) |
| Preflight compares against `origin/<base>`, or reports ref freshness | [KAN-195](https://tweety53.atlassian.net/browse/KAN-195) |
| Guard-to-skill map derived by the guard's own classifier | [KAN-196](https://tweety53.atlassian.net/browse/KAN-196) |

Left unfiled, recorded in PR #10 instead: `resolve_file` still duplicated in
`preserve-session-records.sh` and `gather-self-review-context.sh` (pre-existing, out of scope), and
that helper's cycle and absolute-target branches remain untested in every copy.
