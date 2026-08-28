# kan-88-finish-never-reconciles-the-base-branch

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

`design.md` is canonical for every decision below, for both guards' verdict contracts, and for what
already landed. `proposal.md` is canonical for why the change exists. Nothing here restates them.

**Baseline, measured before any edit:**

- `scripts/` carries 30 guards and 38 test harnesses.
<!-- measured: ls scripts/ | grep -c '^check-'; ls scripts/ | grep -c '^test-' @ c761592 (before this change) -->
- `scripts/test-check-finish-preflight.sh` makes 23 assertions, all passing.
<!-- measured: grep -c 'pass "' scripts/test-check-finish-preflight.sh; scripts/test-check-finish-preflight.sh @ c761592 (before this change) -->
- `skills/flow/integrate.md` is 11500 bytes against a `check-contract-budget.sh` budget of 14370;
  `skills/flow-contracts/finish-contract.md` is 49125 against 54286; `skills/flow/SKILL.md` is
  14271 against 16278.
<!-- measured: wc -c skills/flow/integrate.md skills/flow-contracts/finish-contract.md skills/flow/SKILL.md; grep -n 'integrate.md\|finish-contract.md\|flow/SKILL.md' scripts/check-contract-budget.sh @ c761592 (before this change) -->

**Task order is load-bearing.** Task 3 cites `check-base-moved.sh` from a contract and a skill;
`scripts/check-references.sh` resolves those citations, so landing task 3 before task 2 leaves
`## lint` red on a script that does not exist yet. Task 1 is independent of both.

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` and reconciles
its `budgets()` row in the same commit** — raise a row only where the guard actually fails on that
file, and only to the size its own rule gives. All three files task 3 grows have headroom measured
above, so no row is expected to need raising.

**No task adds a dedicated harness for `scripts/lib/resolve-remote-base.sh`.** It is exercised
through both guards' own harnesses, as `scripts/lib/git-clean.sh` and `scripts/lib/within-root.sh`
already are; `scripts/test-lib-coverage.sh` is `coverage.sh`'s own harness, not a general rule that
every library gets one.

---

- [x] 1. Make the preflight resolve the base ref it was handed

  **Step 1: `scripts/lib/resolve-remote-base.sh`.** A sourceable library defining one function,
  `resolve_remote_base <worktree> <base-ref>`, which prints to stdout `origin/<base-ref>` when
  `refs/remotes/origin/<base-ref>` resolves to a commit in that worktree, and `<base-ref>`
  unchanged otherwise. Resolution is `git -C <worktree> rev-parse --verify --end-of-options
  "refs/remotes/origin/<base-ref>^{commit}"` with stdout and stderr discarded — `--end-of-options`
  so a ref beginning with `-` is read as a ref rather than parsed as a git option. The function
  never fails the caller: an unreadable worktree yields the ref unchanged, because the caller's own
  next step already refuses an unresolvable ref by name. Header comment records `design.md`'s
  `remote-lookup-is-a-preference-not-a-rewrite`: the lookup is a preference, so `origin/main`
  (which resolves `refs/remotes/origin/origin/main`, finding nothing) passes through untouched and
  the contract's own call site is unaffected.

  **Step 2: `scripts/check-finish-preflight.sh`.** Source the library, and compute
  `EFFECTIVE_REF="$(resolve_remote_base "$WORKTREE" "$BASE_REF")"` immediately before the existing
  "base ref that does not resolve" check — after signals 1's two answers, which read only `HEAD`
  and the recorded merge base and must stay answerable when no base ref resolves at all. Replace
  every later use of `$BASE_REF` with `$EFFECTIVE_REF`, in the resolve check, the ancestor test and
  all four verdict lines, so each names the ref the test actually ran against. Extend the usage
  comment to state the substitution and why.

  **Step 3: `scripts/test-check-finish-preflight.sh`.** Add the ticket's own regression case and two
  guards around it, in a new section whose comment names KAN-88:

  - a repository cloned from a bare `origin`, its branch merged into `origin/main`, local `main`
    deliberately left at the recorded merge base, handed the bare name `main` → `RUN2`, and the
    verdict line names `origin/main`;
  - the same repository handed `origin/main` explicitly → `RUN2`, unchanged;
  - the same repository, branch **not** merged, handed the bare name `main` → `RUN1` naming
    `origin/main`, so the substitution is proved not to manufacture a merged verdict.

  Every existing case builds a repository with no `origin` remote and so is unaffected; assert that
  by leaving their expectations byte-identical rather than by editing them.

  Run `scripts/test-check-finish-preflight.sh` and leave it clean.

**Files:** `scripts/lib/resolve-remote-base.sh`, `scripts/check-finish-preflight.sh`, `scripts/test-check-finish-preflight.sh`, `scripts/lib/test-git-shim.sh`, `scripts/test-lib-test-git-shim.sh`, `.flow/project.md`
**Tests:** `scripts/test-check-finish-preflight.sh`, `scripts/test-lib-test-git-shim.sh`
**Regression:** reverting this commit returns the preflight to trusting whatever ref its caller
  hands it, so a caller passing a bare local name feeds a stale local branch into the RUN1/RUN2
  decision — the gate in front of this command's most destructive step, and the failure observed on
  KAN-87, KAN-73 and KAN-201.
**Baseline:** before=23 after=41 assertions in `test-check-finish-preflight.sh`; before=30 after=30 guards and before=38 after=39 harnesses in `scripts/` — this task adds `test-lib-test-git-shim.sh`
<!-- measured: grep -c 'pass "' scripts/test-check-finish-preflight.sh; ls scripts/ | grep -c '^check-' @ d3d762d (after fix round 3) -->
**Commit:** `fix(scripts): resolve the preflight's base ref to its remote-tracking counterpart`
**Build:** green

- [x] 2. Add the base-moved guard

  **Step 1: `scripts/check-base-moved.sh`.** `check-base-moved.sh <worktree> <base-ref>
  <recorded-merge-base|->`, one verdict line to stdout, exit 0 whenever a verdict was reached and
  exit 2 when the tree cannot be read — the protocol `check-finish-preflight.sh` and
  `check-unfinished-work.sh` already use, with the breakdown carried inline on the verdict line as
  `check-unfinished-work.sh` does. `set -euo pipefail` and `export LC_ALL=C`, matching
  `resolve-base-branch.sh`.

  It sources `scripts/lib/resolve-remote-base.sh` from task 1 and resolves its base ref the same
  way, so the two guards can never disagree about which ref answers a question about the base.

  Verdicts:

  ```text unverified:the guard does not exist yet; these are the literals its harness asserts against in step 2
  REFUSE: no merge base recorded for <worktree> — cannot tell whether the base has moved
  REFUSE: recorded merge base '<recorded>' does not resolve in <worktree>
  REFUSE: base ref '<effective-ref>' does not resolve in <worktree> — cannot tell whether the base has moved
  CLEAR: <worktree> — <effective-ref> has not moved since the recorded merge base
  MOVED: <worktree> — <n> commits on <effective-ref> since the recorded merge base; no overlap with this change's paths
  MOVED: <worktree> — <n> commits on <effective-ref> since the recorded merge base; overlaps: <paths>
  ```

  The count is `git rev-list --count --end-of-options <recorded>..<effective-ref>`. The moved paths
  are `git diff --name-only --end-of-options <recorded>..<effective-ref>`. The change's own paths
  are the union of `git diff --name-only --end-of-options <recorded>..HEAD`, `git diff --name-only
  --cached` and `git diff --name-only` — `design.md`'s
  `touched-paths-include-index-and-worktree`, because run 1 is reached with work staged and
  uncommitted by design. The overlap is their intersection, sorted; it is listed inline, comma
  separated, capped at the first 10 with a `(+N more)` tail so one verdict line stays one line.

  Every git invocation whose failure would otherwise be read as an answer is captured on its own
  line and checked, never piped straight into `wc` or `comm` — the reasoning
  `check-finish-preflight.sh`'s signal 3 comment already records, cited rather than restated. A
  failing invocation is exit 2 with a named message, never a `CLEAR`.

  **Step 2: `scripts/test-check-base-moved.sh`.** Same harness shape as
  `scripts/test-check-finish-preflight.sh` — throwaway repositories under a sandboxed `TMPDIR`, an
  indexed `REPOS` array removed by an `EXIT` trap, bash 3.2 as the floor. Cases:

  - an unmoved base → `CLEAR`;
  - a base moved by commits touching only files this change does not → `MOVED` with `no overlap`,
    and the commit count is the number actually added;
  - a base moved by commits touching a file this change also committed → `MOVED` with that path in
    `overlaps:`;
  - the same, where the change's touch is **staged only** → the path still appears, proving the
    index is read;
  - the same, where the change's touch is **unstaged only** → the path still appears, proving the
    working tree is read;
  - overlap of more than 10 paths → the line carries 10 paths and a `(+N more)` tail;
  - `-` as the recorded merge base → `REFUSE`, exit 0;
  - an unresolvable recorded merge base → `REFUSE`, exit 0;
  - an unresolvable base ref → `REFUSE`, exit 0;
  - a bare local name whose `origin/` counterpart exists and has moved → the verdict names
    `origin/<name>`, proving the shared resolution is in force here too;
  - a missing argument → exit 2, no verdict line;
  - a directory that is not a git worktree → exit 2, no verdict line.

  Prove the overlap intersection by mutation rather than asserting it: with the intersection step
  removed, the `no overlap` case must fail. Record that in the harness header, as
  `scripts/test-check-unfinished-work.sh` records its own.

  **Step 3: register the guard.** Create the symlink
  `skills/flow/scripts/check-base-moved.sh -> ../../../scripts/check-base-moved.sh`, matching every
  sibling in that directory. Add `scripts/test-check-base-moved.sh` to `.flow/project.md`'s
  `## test` list.

  Run `scripts/test-check-base-moved.sh`, `scripts/check-guard-symlinks.sh` and
  `scripts/check-references.sh` and leave all three clean.

**Files:** `scripts/check-base-moved.sh`, `scripts/test-check-base-moved.sh`, `skills/flow/scripts/check-base-moved.sh`, `.flow/project.md`
**Tests:** `scripts/test-check-base-moved.sh`
**Regression:** reverting this commit removes the only mechanical answer to "has the base moved
  under this change, and does the movement touch what this change touched" — returning run 1 to
  discovering it at the merge step, which on KAN-129 was after two of three repositories had
  already merged and pushed.
**Baseline:** before=30 after=31 guards and before=39 after=40 harnesses in `scripts/`; 19 cases in the new harness
<!-- measured: ls scripts/ | grep -c '^check-'; ls scripts/ | grep -c '^test-'; grep -c '^# [0-9]' scripts/test-check-base-moved.sh @ 0b62c42 (after fix round 3) -->
**Commit:** `feat(scripts): report how far the base branch has moved under a change`
**Build:** green

- [x] 3. Run the guard before the landing question, and make the contract canonical for it

  **Step 1: `skills/flow-contracts/finish-contract.md`.** A new subsection under **Run 1 — the
  branch is not merged**, placed with the unfinished-work gate it sits beside, canonical for:

  - the invocation, `check-base-moved.sh <worktree> <base-ref> <recorded-merge-base|->`, with
    `<base-ref>` composed as `origin/$BASE` exactly as the preflight's is, and the recorded merge
    base read from the state file's `worktrees` map for that worktree;
  - the three verdicts and the exit contract, cited from the guard rather than re-derived;
  - that it runs once per worktree in the set found by **Resolving a change's worktrees**, never a
    raw read of the `worktrees` map;
  - that every worktree's verdict is reported, and that the operator is asked **once**, for the
    whole change, and **only** when at least one worktree reported an overlap — `design.md`'s
    `ask-only-on-overlap` and `aggregate-the-multi-repo-ask`;
  - that a `REFUSE`, an exit 2, or a resolved set that comes back empty stops and asks, exactly as
    the preflight's own does;
  - the hand-run fallback for a harness whose repository does not carry the script, matching the
    fallback the preflight section already states.

  **Step 2: `skills/flow/integrate.md`.** In **2. Ask how the branch should land**, between the
  `flow.landing-question` begin mark and the landing prompt, run the guard per worktree and report.
  On an overlap from any worktree, one prompt, shape per **Operator prompts**
  (`skills/flow-contracts/operator-prompts.md`):

  ```markdown unverified:authored in-tree for this change; the wording is what step 2 writes
  > **The base branch has moved and touches paths this change also touched — how should
  > integration proceed?**
  > - **Stop — I'll rebase or reorder first** *(recommended)*
  > - **Continue — land anyway**
  ```

  **Stop** exits leaving the change at `IN_PROGRESS` with nothing staged, committed or pushed, and
  closes the mark `flow stage end … -outcome stopped`; **Continue** carries the reported movement
  into the handoff and asks the landing question. No overlap anywhere → report the counts and ask
  the landing question, with no extra prompt. Cite the contract for the verdicts rather than
  restating them.

  **Step 3: `skills/flow/SKILL.md`.** Add `check-base-moved.sh` to the guard union named under
  **Check guard presence**, keeping the list alphabetical.

  Run `scripts/check-references.sh`, `scripts/check-vocabulary.sh`,
  `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
  `scripts/check-stage-mark-calls.sh` and `scripts/check-installed-citations.sh` and leave all six
  clean.

**Files:** `skills/flow-contracts/finish-contract.md`, `skills/flow/integrate.md`, `skills/flow/SKILL.md`
**Tests:** **none** — this task edits contract and skill prose; what verifies it is
  `scripts/check-references.sh`, `scripts/check-vocabulary.sh`,
  `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
  `scripts/check-stage-mark-calls.sh` and `scripts/check-installed-citations.sh`, all in
  `.flow/project.md`'s `## lint` list.
**Regression:** reverting this commit leaves the guard from task 2 shipped and never invoked — run 1
  asks how the branch should land with no one having looked at whether the base moved, which is
  precisely the state KAN-88 reports.
**Baseline:** before=0 after=0 tests added; `skills/flow/integrate.md` 12496 → under its 14370 budget, `skills/flow-contracts/finish-contract.md` 50798 → under its 54286, `skills/flow/SKILL.md` 14294 → under its 16278
<!-- measured: wc -c skills/flow/integrate.md skills/flow-contracts/finish-contract.md skills/flow/SKILL.md @ d5052d1 (after fix round 3) -->
**Commit:** `feat(finish-contract): check the base branch before the landing question`
**Build:** green
