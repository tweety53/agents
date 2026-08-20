# Run 2 asserts its base branch and lands the archive via a pull request

**Change:** `kan-239-run-2-asserts-base-branch-and-archives-via-pr`
**Jira:** [KAN-239](https://tweety53.atlassian.net/browse/KAN-239)
**Date:** 2026-08-20

## The problem

`/myflow-finish` run 2 archives "on the base branch in the main checkout". Two things are missing
from that sentence, and every observed failure follows from one of them:

1. **Nothing asserts that the main checkout is actually on the base branch.**
2. **Nothing says the archive push must not go to `main` directly.**

Five consequences have been observed across KAN-39, KAN-197, KAN-200, KAN-201 and KAN-236, and the
ticket that merges them (KAN-239) absorbs KAN-214, KAN-246 and KAN-133's first half.

- **The archive is staged on whatever branch happens to be checked out.** On KAN-39 the main
  checkout sat on `main` while the base branch was `develop`, and the `git mv` was already staged
  there before anyone noticed. Recovery was manual.
- **The non-base-branch clause would merge unrelated unmerged work.** The contract says that when
  finish is invoked with a non-base branch checked out, it commits there, merges into the base
  branch, and pushes that. On KAN-201 the main checkout sat on `docs/kan-200-self-review-report`,
  itself carrying unmerged work; followed literally, that clause would have merged KAN-200's
  unmerged self-review report into `main`.
- **The archive push violates the no-direct-push-to-`main` rule.** Followed literally, step 3 is
  `git push origin main`. This is not hypothetical: it happened on KAN-236 (commit `332e593`), and
  again on KAN-102 (commit `7796750`), whose archive sits directly on `main` in this repository's
  history today. The repository's own precedent shows the intended pattern — KAN-201's archive
  landed via a PR, KAN-209's via `chore/archive-kan-209` and PR #20 — but an agent following the
  contract does not arrive at it, because getting it right depends on noticing a conflict between
  the contract and a global rule.
- **The self-review report is committed to the archive branch and stranded.** Step 8 says to commit
  the report "in the main checkout", and the shell that follows fixes the *directory* with `-C` and
  nothing fixes the *branch* — while step 3 has just put that checkout on an archive branch. This
  happened on KAN-197 and again on KAN-200, the second time inside the run that had just recovered
  the first. The failure is silent and looks like success: commit succeeds, push succeeds, the
  handoff prints a path that exists locally, and `FINISHED` is correct. Only the clock decides
  whether it shows.
- **`gather-self-review-context.sh` becomes structurally unrunnable.** It derives its trust anchor
  from `git rev-parse --git-common-dir`, which always resolves to the main repository, and then
  requires the archived path to sit at `<main-repo-root>/openspec/changes/archive/<leaf>`. That is
  deliberate and documented in the script's own header — but it means the archive must exist in the
  main checkout's *working tree*, which it does not when that checkout is on another branch. On
  KAN-201 every attempted invocation exited 0 with every source skipped, and the self-review pass
  ran from the sources by hand.

Both omissions are closed structurally here rather than by asking a reader to be careful.

## The shape of the fix

Run 2 stays a **single terminal run** — this change adds no third run and no new pipeline state.
What changes is where run 2 puts the checkout before it archives, where the archive lands, and when
the push happens.

### Run 2's step order

| # | Step | Status |
|---|------|--------|
| 1 | Verify the merge | unchanged |
| 2 | **Position the checkout on `chore/archive-<name>`, cut from an up-to-date `BASE`** | **new** |
| 3 | Sync delta specs, then move the change into the archive | unchanged in substance; now provably staged on the archive branch |
| 4 | **Commit the archive — no push** | the push moves to step 10 |
| 5 | Clean up worktrees, branches, workspace database and bucket | unchanged |
| 6 | Remove the proposal artifact source | unchanged |
| 7 | Verify the cleanup (`COMPLETE:` gate) | unchanged |
| 8 | Write `FINISHED`, transition Jira to `Done` | unchanged |
| 9 | Self-review — the report is **committed onto the archive branch** | the branch is now asserted, not assumed |
| 10 | **Push the archive branch and open its pull request** | **new** |
| 11 | **Restore the main checkout to `BASE`** | **new** |

Step 3's clause *"when finish is invoked with a non-base branch checked out, it commits there,
merges into the base branch, and pushes that"* is **deleted outright**. It is the clause that would
have merged KAN-200's unmerged work into `main`, and step 2 removes the situation it was written
for.

### Why the push is last, and why `FINISHED` still precedes self-review

The archive branch is pushed once, at step 10, carrying **both** the archive commit and the
self-review report. There is therefore no window in which the archive PR is merged while the report
is still unwritten — the timing race that stranded KAN-197's and KAN-200's reports is not guarded
against, it is absent.

`FINISHED` is still written at step 8, before self-review, so
**Requirement: Self-review runs only after FINISHED is written**
(`openspec/specs/myflow-self-review/spec.md`) is preserved byte for byte and needs no delta. The
alternative — pushing before `FINISHED` so that a failed push blocks the terminal write — was
considered and rejected by the operator; see **Decisions** below.

The consequence is accepted explicitly: **a step-10 failure leaves a `FINISHED` change whose archive
branch exists only locally.** The handoff must therefore name the branch and print the exact push
and PR commands, and must never report the archive as landed. This is the same class as today's
rule that a failed report commit is reported and never reopens the change.

## `prepare-archive-branch.sh` — the new guard

The branch decision moves out of prose and into a script with an exit contract, for the same reason
`check-finish-preflight.sh` and `resolve-base-branch.sh` already exist: the finish contract states
that the decision in front of the irreversible step is made *by a script, not by prose*.

```bash unverified:this is the guard's proposed usage line; task 2 writes it, and its header is the authority once it exists
prepare-archive-branch.sh <main-checkout> <base> <archive-branch>
```

On success the main checkout is on `<archive-branch>`, cut from a `<base>` that has been
fast-forwarded to `origin/<base>`.

| Exit | Meaning |
|------|---------|
| `0` | resolved — the checkout is on `<archive-branch>`, cut from an up-to-date `<base>`; one line on stdout naming the branch it started from and the branch it is on |
| `1` | a named refusal — a dirty working tree, **on `<base>` or off it**; a detached `HEAD`; or an existing `<archive-branch>` that is not descended from `origin/<base>` |
| `2` | `<main-checkout>` is missing, unreadable, or not a git worktree |
| `3` | `<base>` cannot be fast-forwarded to `origin/<base>` — the local branch has diverged |

The clean-versus-dirty split is the operator's chosen behaviour: **clean switches; dirty refuses**,
naming the branch found and the branch required. A dirty tree is refused **wherever it is found — on
`<base>` as well as off it** — because uncommitted changes would otherwise ride onto the archive
branch unremarked. Refusing always would
cost a re-run every time an operator has anything else checked out; switching always would need a
stash-and-restore, which is precisely where KAN-39's manual recovery already went wrong.

Re-entrancy: an existing `<archive-branch>` descended from `origin/<base>` is **reused**, not
refused, so a run 2 that stopped between steps 2 and 10 can be re-run.

`<base>` is never guessed — it is whatever `resolve-base-branch.sh` printed, resolved against the
apply worktree exactly as run 1 resolves it. A repository with no `origin` is that script's exit `3`
and is reported as such.

Per this repository's every-guard-has-a-mutation-test rule (KAN-197), the guard ships with
`test-prepare-archive-branch.sh`, and with symlinks under `skills/myflow-finish/scripts/` and
`skills/myflow-fast/scripts/` so `check-guard-symlinks.sh`'s rule 2 is satisfied.

## `gather-self-review-context.sh` — an explicit repo root

The script gains an **optional fourth positional argument**, `<repo-root>`.

- **Omitted** — behaviour is exactly as today: `TRUSTED_REPO_ROOT` is derived from the process cwd
  via `git rev-parse --git-common-dir`. Every existing caller and every existing test keeps working
  unchanged.
- **Supplied** — it becomes `TRUSTED_REPO_ROOT`, after the same validation the archived path already
  receives: it must be absolute, it must show no divergence between its lexical and its
  symlink-resolved form, and it must be a git repository root. The archived-path containment check
  then runs against it unchanged.

Trusting a caller-supplied root is safe in a way trusting the archived path is not: the root comes
from the skill, while the archived path is the untrusted input being validated. The forty lines of
header defending the derived anchor are about *not deriving the anchor from the thing under test*,
and an explicitly supplied root does not weaken that.

This decouples the script from the caller's cwd and branch, which is KAN-201's structural failure.
Under this design the main checkout is on the archive branch and the archived directory is present
in its working tree, so the derived anchor would also resolve — the argument exists so that the
script's correctness no longer depends on that coincidence.

## Vocabulary, registry and boundaries

- **One new stage key**, `finish.push-archive`, for step 10. `finish.commit-archive` keeps its key
  and now brackets the commit alone.
- **`finish.sync-archive`'s prose name is reworded** to name the positioning it now includes. Keys
  never change; names may be reworded freely, because nothing is keyed on them
  (`stats/internal/stages/names.go`). `README.md`'s Level 1 table is the source the Go vocabulary is
  transcribed from, and `TestStagesMatchReadmeLevelOne` keeps the two in step.
- **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`) gains a row for the
  archive branch. Its removal column says, honestly, that **nothing in this pipeline removes it** —
  run 2 is terminal and the pull request outlives it. This repository already carries five such
  leftovers (`chore/archive-kan-197`, `chore/archive-kan-200`, `chore/archive-kan-209`,
  `chore/self-review-kan-201`, `chore/self-review-kan-236`), so a row claiming a cleanup that does
  not happen would be worse than no row. See **Open questions**.
- **Git boundaries** (`skills/myflow-contracts/pipeline.md`) gains the run-2 rule: run 2 commits
  only on `chore/archive-<name>`, never on the base branch, and never pushes to the base branch at
  all.

## Testing

- `test-prepare-archive-branch.sh` — a mutation test in the shape this repository's other guard
  harnesses use, covering each exit: on-base clean, off-base clean (switches), off-base dirty
  (refuses), detached `HEAD`, an existing archive branch descended from `origin/<base>` (reuses), an
  existing one that is not (refuses), an unreadable checkout, and a diverged local base.
- `test-gather-self-review-context.sh` — new cases for the fourth argument: omitted (existing
  behaviour), supplied and valid, supplied but relative, supplied but reaching through a symlink,
  supplied but not a git root.
- `.myflow/project.md`'s `## test` list gains the new harness, so it runs in this repository's own
  test command.

## Decisions

### Where the archive lands

**ID:** archive-lands-via-pr
**Status:** active
**Chosen:** a `chore/archive-<name>` branch and a pull request — it removes the need for the agent
to reconcile the finish contract against the global no-direct-push-to-`main` rule, which is the
actual failure mode; the repository's own precedent (KAN-201, KAN-209) already does this.
**Considered:** a direct push to the base branch — the status quo; violates the global rule and has
now produced two recorded violations on `main`.

### What run 2 does when the main checkout is on the wrong branch

**ID:** switch-if-clean-refuse-if-dirty
**Status:** active
**Chosen:** switch when the working tree is clean, refuse when it is dirty — KAN-133 §1's suggested
fix, guarded. It never destroys operator state and costs a re-run only when there is state to lose.
**Considered:** refuse always — smallest change, but costs a re-run every time the operator has
anything else checked out, which is most of the time. Switch always with a stash-and-restore —
most convenient, most machinery, and stash/pop is exactly where KAN-39's manual recovery went wrong.

### Number of runs

**ID:** run-2-stays-terminal
**Status:** active
**Chosen:** two runs, as today — run 2 archives, cleans up, writes `FINISHED` and opens the archive
PR. `FINISHED` means the change's branch merged and the pipeline's artifacts were cleaned up, which
step 1 proved and step 7 verified; the archive PR is a follow-on the handoff names.
**Considered:** a third run that waits for the archive PR to merge before writing `FINISHED` — most
faithful to "archived" meaning archived on the base branch, but it turns a two-run finish into
three, leaves the change listed as open in the meantime, and needs a new run-3 trigger that the
existing preflight cannot produce, since by then the worktrees it reads are gone.

### Where the push sits relative to the `FINISHED` write

**ID:** finished-before-push
**Status:** active
**Chosen:** write `FINISHED` first, then self-review, then push and open the PR. This preserves
**Requirement: Self-review runs only after FINISHED is written** byte for byte, so the change needs
no delta against `myflow-self-review`.
**Considered:** push before `FINISHED`, so that a failed push stops at `IN_PROGRESS` and is
re-runnable rather than writing a terminal state over an archive that never left the machine. It is
the safer failure mode, but it forces self-review to run before `FINISHED` and therefore a rewrite
of that requirement's anchor. The residual risk is handled instead by requiring the handoff to name
the unpushed branch and print the exact commands.

### The self-review context script's trust anchor

**ID:** explicit-repo-root-optional
**Status:** active
**Chosen:** an optional fourth positional argument, validated the same way the archived path is, with
the derived `--git-common-dir` anchor unchanged when it is omitted. Decouples the script's
correctness from the caller's cwd and branch without breaking a single existing call site.
**Considered:** leaving the script alone — under this design the derived anchor happens to resolve,
so nothing would break today; rejected because that is a coincidence of the new step order rather
than a property of the script.

## Open questions

### Nothing removes a merged `chore/archive-<name>` branch

**ID:** archive-branch-cleanup
**Status:** open
**Why it is open:** deferred — recording the fact in the **Temporary artifacts registry** is in
scope here, and deciding what should remove the branches is a separate change. Run 2 is terminal, so
the pull request it opens outlives the run and no later run exists to delete the branch it was
opened from; the forge deletes the remote branch on merge only where that is configured, and nothing
deletes the local one. This repository already carries five such leftovers.
**What it affects:** the archive branch's row in the **Temporary artifacts registry**, and whether
some future run — or `check-cleanup-complete.sh` — gains a duty to remove a merged archive branch.

### A supplied `<repo-root>` accepts only the main checkout, never a worktree

**ID:** repo-root-refuses-a-worktree
**Status:** open
**Why it is open:** discovered during implementation, not at design time. The operator chose this
argument from an option worded *"makes the script runnable from a temporary worktree, decoupling it
from the main checkout's branch"*, but the implementation validates a supplied root by resolving it
through `git rev-parse --git-common-dir`, which always lands on the main repository — so a worktree
root is **refused**. That keeps the supplied and derived cases semantically identical, and the
per-task reviewer confirmed the containment boundary is not weakened by the override; it also
delivers less than the option's own wording promised. This design's stated motivation is decoupling
from the caller's *working directory*, which is met in full.
**What it affects:** whether a future caller may invoke `gather-self-review-context.sh` against a
temporary worktree. Nothing in this change needs that, because run 2 now puts the **main** checkout
on the archive branch — the temporary worktree KAN-201 improvised is no longer created. Widening it
would mean accepting a worktree root explicitly and restating the trust argument for that case.
