> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** `/myflow-finish` run 2 can no longer archive on the wrong branch, and can no longer reach
the base branch by a direct push. Both are closed by mechanism — a guard that positions the checkout,
and a landing route that goes through a pull request — rather than by a clearer sentence in a
contract.

**Architecture:** One new guard, `scripts/prepare-archive-branch.sh`, modelled on
`scripts/resolve-base-branch.sh`, plus its fixture harness. `scripts/gather-self-review-context.sh`
gains an optional fourth positional argument so its trust anchor no longer depends on the caller's
cwd. One new stage key, `finish.push-archive`, added to `README.md`'s Level 1 table and transcribed
into `stats/internal/stages/names.go`. Then the three documents that describe run 2 —
`skills/myflow-contracts/finish-contract.md`, `skills/myflow-finish/SKILL.md` and
`skills/myflow-contracts/pipeline.md` — are brought onto the new step order.

**Tech Stack:** Bash (guard and harness; Bash 3.2 is the floor, as
`scripts/test-check-finish-preflight.sh`'s header records for this repository) and Go for the stage
vocabulary. Nothing is added to either.

**Spec:** `openspec/changes/kan-239-run-2-asserts-base-branch-and-archives-via-pr/design.md` and the
two delta specs beside it under `specs/`. The approved design is
`docs/superpowers/specs/2026-08-20-kan-239-run-2-asserts-base-branch-and-archives-via-pr-design.md`.

## Global Constraints

- **Run 2 stays a single terminal run.** No task adds a third run, a new pipeline state, or a new
  state-file field. A task that introduces one has exceeded its scope.
- **No task changes `openspec/specs/myflow-self-review/spec.md`'s requirement
  `Self-review runs only after FINISHED is written`.** The new step order preserves it deliberately —
  `FINISHED` is still written before self-review runs. This is decision `finished-before-push` in
  `design.md`.
- **A guard is named by basename in every invoking position**, never as a repository-relative
  `scripts/<name>` path — `scripts/check-guard-symlinks.sh`'s rule 3 rejects the latter.
- **A guard does not derive a repository root as a fixed depth above itself**
  (`scripts/check-guard-symlinks.sh`'s rule 4). Derive it from the resolved physical location.
- **Every path citation added to an installed file names its root**, per
  `scripts/check-installed-citations.sh` — `<agents repo>/` for this checkout, `<project>/` for the
  project a command runs against, and installed roots (`skills/`, `rules/`, `commands/`,
  `commands-claude/`, `hooks/`) bare.
- **Every grep in a new guard carries `-a` and `--`, and splits `rc > 1` from `rc == 1`**, per the
  three disciplines `scripts/check-guard-symlinks.sh`'s header records.
- **A refusal writes to stderr and nothing to stdout.** No caller may read an absent report as a
  clean one.
- **The new guard never runs a network operation without the bounded, credential-free fetch wrapper
  `scripts/resolve-base-branch.sh` already uses.** An unreachable remote must refuse quickly rather
  than hang.
- **No task edits `openspec/` or `docs/superpowers/`.** Those are the planning paths `/myflow-do`
  never stages.

## Baseline

All measured 2026-08-20 against `e61603b`.

- `scripts/check-contract-budget.sh` reports `BUDGET-OK: 27 contract file(s) within budget`.
  <!-- measured: scripts/check-contract-budget.sh @ e61603b -->
- `skills/myflow-contracts/finish-contract.md` is 36613 bytes against a budget row of 40724 —
  **4111 bytes of headroom**, the tightest row this change touches.
  <!-- measured: wc -c skills/myflow-contracts/finish-contract.md and the budgets table in scripts/check-contract-budget.sh @ e61603b -->
- `skills/myflow-finish/SKILL.md` is 35173 bytes against 39965 — 4792 bytes of headroom.
  <!-- measured: wc -c skills/myflow-finish/SKILL.md and the budgets table in scripts/check-contract-budget.sh @ e61603b -->
- `skills/myflow-contracts/pipeline.md` is 45990 bytes against 55728 — 9738 bytes of headroom.
  <!-- measured: wc -c skills/myflow-contracts/pipeline.md and the budgets table in scripts/check-contract-budget.sh @ e61603b -->
- `scripts/check-guard-symlinks.sh` reports
  `GUARD-SYMLINKS-OK: /Users/tweety53/Projects/agents — 63 guard(s) across 6 skill(s) validated`,
  with `myflow-do 14 · myflow-fast 18 · myflow-finish 7 · myflow-status 1`.
  <!-- measured: scripts/check-guard-symlinks.sh @ e61603b -->
- `scripts/check-installed-citations.sh` reports
  `INSTALLED-CITATIONS-OK: /Users/tweety53/Projects/agents — 56 file(s) scanned`.
  <!-- measured: scripts/check-installed-citations.sh @ e61603b -->
- `scripts/check-references.sh` reports `check-references: all referenced sections resolve`.
  <!-- measured: scripts/check-references.sh @ e61603b -->
- `scripts/check-vocabulary.sh` reports `✓ Panel-vocabulary guard: clean`.
  <!-- measured: scripts/check-vocabulary.sh @ e61603b -->
- `scripts/test-gather-self-review-context.sh` passes with 95 `ok:` assertions and prints
  `gather-self-review-context: all cases pass`.
  <!-- measured: bash scripts/test-gather-self-review-context.sh | grep -c '^ok:' @ e61603b -->
- `stats/internal/stages/names.go` declares 34 stages, of which 14 begin `finish.`; there is no
  `finish.push-archive`.
  <!-- measured: grep -c '^\s*{Key:' stats/internal/stages/names.go and grep -n 'finish\.' README.md @ e61603b -->
- `scripts/prepare-archive-branch.sh` and `scripts/test-prepare-archive-branch.sh` do not exist.
  <!-- measured: ls scripts/prepare-archive-branch.sh scripts/test-prepare-archive-branch.sh @ e61603b -->
- This repository carries five leftover archive/self-review branches from the hand-recoveries the
  ticket describes: `chore/archive-kan-197`, `chore/archive-kan-200`, `chore/archive-kan-209`,
  `chore/self-review-kan-201`, `chore/self-review-kan-236`.
  <!-- measured: git branch -a @ e61603b -->

---

### 1 `scripts/test-prepare-archive-branch.sh` — the harness, before the guard exists

**Build:** red

**Squash-with:** Task 2

**Files:**
- Add: `scripts/test-prepare-archive-branch.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the executable definition of the guard's contract — its four exit codes and the state it
  leaves the checkout in for each. Task 2 satisfies it; tasks 6 and 7 describe it.

The guard's value is entirely in what it refuses. Written after the guard, a harness asserts what the
guard happens to do; written first, it asserts what
`specs/myflow-finish-cleanup/spec.md`'s **Requirement: Run 2 positions the main checkout before it
archives** says it must do.

Model it on `scripts/test-resolve-base-branch.sh`: real throwaway git repositories under `TMPDIR`
with a local `origin` created by `git init --bare` and cloned from, one `ok:`/`FAIL:` line per
assertion, a `prepare-archive-branch: all cases pass` line at the end, and a non-zero exit on any
failure.

- [x] **Step 1: Build the fixture-repository helper**

A helper that materialises a bare `origin` plus a clone under `TMPDIR`, seeds `<base>` with one
commit, pushes it, and sets `refs/remotes/origin/HEAD`. Every case below starts from it and then
diverges. Follow `scripts/test-resolve-base-branch.sh`'s existing helper rather than inventing a
second fixture idiom.

```bash unverified:confirm scripts/test-resolve-base-branch.sh's fixture helper name and whether it can be reused rather than re-derived
grep -n 'mktemp\|git init\|fixture\|make_repo' scripts/test-resolve-base-branch.sh | head -20
```

- [x] **Step 2: Write the exit-code cases**

| Case | Fixture state | Expects |
|------|---------------|---------|
| `on-base-clean` | checkout on `<base>`, clean, one commit behind `origin/<base>` | exit `0`; `<base>` fast-forwarded; `HEAD` on `<archive-branch>` |
| `off-base-clean` | checkout on an unrelated branch, clean | exit `0`; `HEAD` on `<archive-branch>`, cut from a fast-forwarded `<base>` |
| `off-base-dirty` | checkout on an unrelated branch with an uncommitted modification | exit `1`; stderr names both the branch found and `<base>`; `HEAD` unchanged; nothing staged |
| `on-base-dirty` | checkout on `<base>` with an uncommitted modification | exit `1`; `HEAD` unchanged — a dirty tree is refused wherever it is, because the fast-forward would fail or silently carry the changes onto the archive branch |
| `detached-head` | `git checkout --detach` | exit `1`; stderr says `HEAD` is detached; `HEAD` unchanged |
| `archive-branch-exists-descended` | `<archive-branch>` already exists and is descended from `origin/<base>` | exit `0`; the existing branch is reused, not recreated; its commits survive |
| `archive-branch-exists-unrelated` | `<archive-branch>` exists on an unrelated root commit | exit `1`; stderr says it is not descended from `origin/<base>`; `HEAD` unchanged |
| `not-a-worktree` | a plain directory under `TMPDIR` | exit `2`; nothing on stdout |
| `missing-checkout` | a path that does not exist | exit `2`; nothing on stdout |
| `base-diverged` | local `<base>` carries a commit `origin/<base>` does not | exit `3`; stderr says the fast-forward is impossible; `HEAD` unchanged |
| `no-origin` | a repository with no `origin` remote | exit `2` or `3` per the guard's own header, asserted against whichever the guard documents — never left unasserted |

- [x] **Step 3: Assert the stdout/stderr split**

One case per exit code asserting that a refusal writes nothing to stdout, and that exit `0` writes
exactly one line naming the branch it started from and the branch it left the checkout on.

- [x] **Step 4: Confirm the harness fails against the absent guard**

Run it before task 2 exists and confirm it fails loudly rather than passing vacuously — the
red half of this pair.

```bash verified:run against this tree at e61603b; it exits 2 with the guard absent
bash scripts/test-prepare-archive-branch.sh; echo "exit $?"
```

**Tests:** the cases enumerated in steps 2 and 3, in `scripts/test-prepare-archive-branch.sh`.

**Regression:** Reverting this task leaves the guard with no executable statement of its contract,
so every refusal it is supposed to make becomes unasserted.

**Baseline:** before=0 after=11 cases in `scripts/test-prepare-archive-branch.sh`.

**Commit:** `test(kan-239-run-2-asserts-base-branch-and-archives-via-pr): pin the archive-branch guard's four exit codes and the state each leaves behind`

---

### 2 `scripts/prepare-archive-branch.sh` — the guard

**Build:** green

**Files:**
- Add: `scripts/prepare-archive-branch.sh`
- Add: `scripts/test-prepare-archive-branch.sh` — task 1's file. Task 1 is `Build: red` with
  `**Squash-with:** Task 2`, so its commit folds into this one and this commit is the squash unit's
  only commit. The field declares what the commit touches, which is both files.
- Add: `skills/myflow-finish/scripts/prepare-archive-branch.sh` — a relative symlink to
  `../../../scripts/prepare-archive-branch.sh`
- Add: `skills/myflow-fast/scripts/prepare-archive-branch.sh` — the same relative symlink
- Modify: `.myflow/project.md` — add `scripts/test-prepare-archive-branch.sh` to the `## test` list

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 1's harness, which defines the contract this task satisfies.
- Produces: `prepare-archive-branch.sh <main-checkout> <base> <archive-branch>`, exiting `0`
  positioned / `1` a named refusal / `2` the checkout cannot be read / `3` the base cannot be
  fast-forwarded. Tasks 6 and 7 invoke it by basename.

- [x] **Step 1: Write the header**

The header is the authority for the exact commands and their order, in the shape
`scripts/resolve-base-branch.sh`'s header already uses — so that the by-hand fallback tasks 6 and 7
write can cite it rather than restate it. It states the four exit codes, the clean-versus-dirty
split, the reuse rule for an existing archive branch, and that the fetch is the bounded,
credential-free wrapper.

- [x] **Step 2: Validate the arguments and the checkout**

Three required arguments. `<main-checkout>` must exist and be a git worktree — otherwise exit `2`.
`<base>` and `<archive-branch>` must pass the same branch-name shape assertion
`scripts/resolve-base-branch.sh` applies: the first character is one of `[A-Za-z0-9._]`, and every
character is one of `[A-Za-z0-9._/-]`. A name failing it is exit `1`, never a `git` call with an
argument that could be read as an option.

- [x] **Step 3: Fetch, then decide**

Run the bounded fetch, then read `HEAD`. Detached → exit `1`. Then:

- `HEAD` is `<base>` → fast-forward it to `origin/<base>`; a non-fast-forward is exit `3`.
- `HEAD` is anything else and the working tree is clean → check out `<base>`, then fast-forward as
  above.
- `HEAD` is anything else and the working tree is dirty → exit `1`, naming both branches.
- The working tree is dirty on `<base>` itself → exit `1` as well; a dirty tree is refused wherever
  it is, because the changes would otherwise ride onto the archive branch unremarked.

Cleanliness is `git -C <main-checkout> status --porcelain` being empty — the same test
`scripts/check-finish-preflight.sh` uses for its signal 3, so run 2's two cleanliness judgments
cannot disagree.

- [x] **Step 4: Create or reuse the archive branch**

If `<archive-branch>` does not exist, create it from the fast-forwarded `<base>` and check it out.
If it exists, test whether it is descended from `origin/<base>` with `git merge-base --is-ancestor`;
descended → check it out and reuse; otherwise exit `1`.

- [x] **Step 5: Print the one success line and register the harness**

On success print one line naming the branch the checkout started from and the branch it is on now,
then add `scripts/test-prepare-archive-branch.sh` to `.myflow/project.md`'s `## test` list.

- [x] **Step 6: Verify**

```bash verified:run against this tree at e61603b; both guards pass on the untouched tree
bash scripts/test-prepare-archive-branch.sh
scripts/check-guard-symlinks.sh
```

**Tests:** task 1's cases, now passing, in `scripts/test-prepare-archive-branch.sh`.

**Regression:** Reverting this task removes the only mechanism that stops run 2 archiving on whatever
branch happens to be checked out — KAN-39's failure, and the precondition for KAN-201's.

**Baseline:** before=63 after=65 guards reported by `scripts/check-guard-symlinks.sh`. The
per-skill breakdown stays at `myflow-finish 7 · myflow-fast 18` here: that column counts the guards a
skill's own prose **invokes** (rule 2's required set), not the symlinks present, so it moves only when
task 7 writes the invocation.
<!-- measured: scripts/check-guard-symlinks.sh @ 1485c34, after the symlinks landed; the earlier prediction of 7 → 8 / 18 → 19 at this task was wrong and is corrected here -->

**Commit:** `feat(kan-239-run-2-asserts-base-branch-and-archives-via-pr): put the main checkout on an archive branch cut from an up-to-date base`

---

### 3 `scripts/test-gather-self-review-context.sh` — cases for the optional repo root

**Build:** red

**Squash-with:** Task 4

**Files:**
- Modify: `scripts/test-gather-self-review-context.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: the existing harness's fixture helpers, unchanged.
- Produces: the cases task 4 must satisfy. Nothing else depends on this task.

- [x] **Step 1: Add the four cases**

| Case | Invocation | Expects |
|------|------------|---------|
| `repo-root-omitted` | three arguments, cwd inside the repository | unchanged behaviour — the existing assertions still hold, byte for byte |
| `repo-root-supplied` | four arguments, the fourth an absolute symlink-free repository root, invoked from a cwd **outside** that repository | the bundle is gathered against the supplied root |
| `repo-root-relative` | fourth argument relative | exit `2`, nothing gathered |
| `repo-root-symlinked` | fourth argument reaching through a symlink | exit `2`, nothing gathered |
| `repo-root-not-a-repo` | fourth argument an ordinary directory | exit `2`, nothing gathered |

The `repo-root-supplied` case is the one that proves the point: invoked from outside, the derived
`--git-common-dir` anchor cannot resolve, so the case passes only if the argument is actually being
used.

- [x] **Step 2: Confirm the new cases fail against the current script**

```bash verified:run against this tree at e61603b; it passes with 95 ok: assertions before task 3 adds any
bash scripts/test-gather-self-review-context.sh; echo "exit $?"
```

**Tests:** the five cases in step 1, in `scripts/test-gather-self-review-context.sh`.

**Regression:** Reverting this task leaves the fourth argument's validation unasserted, so a later
change could accept a relative or symlinked root without any test noticing.

**Baseline:** before=95 after=100 `ok:` assertions in `scripts/test-gather-self-review-context.sh`.

**Commit:** `test(kan-239-run-2-asserts-base-branch-and-archives-via-pr): pin the self-review context script's optional repository root`

---

### 4 `scripts/gather-self-review-context.sh` — the optional repo root

**Build:** green

**Files:**
- Modify: `scripts/gather-self-review-context.sh`
- Modify: `scripts/test-gather-self-review-context.sh` — task 3's file, folded in by
  `**Squash-with:** Task 4`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 3's cases.
- Produces: a fourth, optional positional argument. Task 7 passes it from
  `skills/myflow-finish/SKILL.md` step 8.

- [x] **Step 1: Accept and validate the fourth argument**

Optional. When absent, `TRUSTED_REPO_ROOT` is derived exactly as today — the existing
`--git-common-dir` path is not touched. When present, validate it with the mechanisms the script
already owns rather than new ones: it must be absolute; its lexical normalisation and its `cd -P`
resolution must be identical; and it must be a git repository root. Any failure is exit `2`, the
same class as a malformed `<archived-change-path>`.

- [x] **Step 2: Extend the header**

The header is where this script's trust argument lives. Add the paragraph that says why accepting a
caller-supplied root does not weaken it: the prohibition being defended is deriving the anchor from
`<archived-change-path>` — the untrusted input under test — not accepting one from the caller. A
reader who does not find that stated will delete the argument as a regression.

- [x] **Step 3: Update the usage line**

`Usage: gather-self-review-context.sh <archived-change-path> <name> <state-dir> [<repo-root>]`.

- [x] **Step 4: Verify**

```bash verified:run against this tree at e61603b; the same command, re-run after the change
bash scripts/test-gather-self-review-context.sh
```

**Tests:** task 3's cases, now passing.

**Regression:** Reverting this task re-couples the script's correctness to the caller's process
working directory, which is what made it structurally unrunnable on KAN-201.

**Baseline:** before=95 after=100 `ok:` assertions, all passing.

**Commit:** `feat(kan-239-run-2-asserts-base-branch-and-archives-via-pr): accept an explicit repository root for the self-review bundle`

---

### 5 `README.md` and `stats/internal/stages/names.go` — the `finish.push-archive` stage

**Build:** green

**Files:**
- Modify: `README.md` — one new row in the Level 1 table, and `finish.sync-archive`'s Name reworded
- Modify: `stats/internal/stages/names.go` — the transcription of that table

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the stage key `finish.push-archive`, which task 7's mark calls carry.

- [x] **Step 1: Add the row to `README.md`'s Level 1 table**

`| `finish.push-archive` | Push the archive branch and open its PR (run 2) | `/myflow-finish`, `/myflow-fast` |`,
placed after `finish.self-review` — the table's order is the run's order, and this is now run 2's
last step.

- [x] **Step 2: Reword `finish.sync-archive`'s Name**

To name the positioning it now includes. The **Key** is untouched: a key never changes, because a
key slugified from a reworded name would split a stage's recorded history — the property
`stats/internal/stages/names.go`'s package comment states.

- [x] **Step 3: Transcribe into `names.go`**

Add the matching `Stage` entry. `TestStagesMatchReadmeLevelOne` re-derives the table from `README.md`
at test time, so a mismatch fails the build rather than drifting.

- [x] **Step 4: Verify**

```bash verified:run against this tree at e61603b; go vet and gofmt exit clean today
cd stats && go test ./internal/stages/... && go vet ./... && gofmt -l .
```

**Tests:** none added — and none can be. The Go test that covers this task, TestStagesMatchReadmeLevelOne
in the stages package, re-derives the table from README.md mechanically, so it covers the new row
**without being edited**. Declaring it here as a test this task adds is what the task-commit-fields
guard correctly rejects: the field names tests a task's own commit introduces, and this commit
introduces none.

**Regression:** Reverting this task makes every `finish.push-archive` mark task 7 writes a caller
mistake the CLI rejects, so run 2's final step would record nothing. The existing stages test fails
on the revert too, since README.md and the Go table would disagree again.

**Baseline:** before=34 after=35 stages in `stats/internal/stages/names.go`.

**Commit:** `feat(kan-239-run-2-asserts-base-branch-and-archives-via-pr): add the finish.push-archive stage to the documented vocabulary`

---

### 6 `skills/myflow-contracts/finish-contract.md` — run 2's new step order

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/finish-contract.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 2's guard, invoked by basename.
- Produces: the canonical statement of run 2's steps. Task 7 cites it rather than restating it.

This file is canonical for run 2. Task 7's `skills/myflow-finish/SKILL.md` carries only what is
specific to *executing* the procedure.

- [x] **Step 1: Add the positioning step, before the sync**

A new step between "Verify the merge" and "Sync delta specs": resolve `<base>` with
`resolve-base-branch.sh` against the apply worktree, exactly as run 1 does, then invoke
`prepare-archive-branch.sh <main-checkout> <base> chore/archive-<name>`. State the four exit codes
and that anything but `0` stops run 2 with nothing staged, committed, pushed or removed. Add the
by-hand fallback paragraph in the shape the base-branch resolver's already has, citing that guard's
own header as the authority for the commands rather than restating them.

- [x] **Step 2: Rewrite the commit step**

The archive is committed on `chore/archive-<name>` and **not pushed here**. Delete the clause *"When
finish is invoked with a non-base branch checked out, it commits there, merges into the base branch,
and pushes that."* — the deletion is the point of the task, not a side effect. Add the rule that
every commit run 2 makes after the positioning step asserts the branch rather than assuming it,
because `git -C <main-checkout>` fixes the directory and not the branch.

- [x] **Step 3: Add the push step as run 2's last**

After step 8 (self-review): push `chore/archive-<name>`, open a pull request against `<base>` via a
PR CLI when usable for the host and otherwise print the forge's create-PR URL and ask whether it was
opened — the same shape run 1's pull-request route already uses. State that run 2 never pushes to
`<base>`; that a failed push or PR creation is reported with the command's own output, never moves
the change off `FINISHED`, and makes the handoff name the unpushed branch and the exact commands;
and that the checkout is restored to `<base>` afterwards.

- [x] **Step 4: Update step 8's report-commit sentence**

The report is committed on the archive branch and **not pushed by self-review** — the push step
above carries it.

- [x] **Step 5: Verify, watching the budget**

The file has 4111 bytes of headroom, the tightest row this change touches; step 2's deletion returns
some of what steps 1 and 3 spend.

```bash verified:run against this tree at e61603b; all four guards pass on the untouched tree
scripts/check-contract-budget.sh
scripts/check-references.sh
scripts/check-installed-citations.sh
scripts/check-guard-symlinks.sh
```

**Tests:** none added. This task is covered by the four guards in step 5, which check its citations,
its budget row, and that the guard it names by basename is symlinked into the skills that invoke it.

**Regression:** Reverting this task restores both original omissions in the one file that is
canonical for run 2 — after which every other document describing run 2 would be describing a
procedure the contract contradicts.

**Baseline:** before=27 after=27 contract files within budget; `finish-contract.md` stays under
40724 bytes.

**Commit:** `feat(kan-239-run-2-asserts-base-branch-and-archives-via-pr): position the checkout before archiving and land the archive via a pull request`

---

### 7 `skills/myflow-finish/SKILL.md` — run 2's executable form

**Build:** green

**Files:**
- Modify: `skills/myflow-finish/SKILL.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 2's guard, task 4's fourth argument, task 5's stage key, task 6's canonical
  contract.
- Produces: the run-2 procedure `/myflow-finish` and `/myflow-fast` both execute.

- [x] **Step 1: Insert the positioning into the numbered outline**

Between step 1 and step 2, with its own invocation line. It belongs inside the existing
`finish.sync-archive` mark, whose prose name task 5 reworded to say so — no new mark opens for it.

- [x] **Step 2: Update the archive commit and the mark that brackets it**

`finish.commit-archive` now brackets the commit alone. Its shell commits on
`chore/archive-<name>` and asserts that branch in the same `&&` chain, mirroring the guarded shape
**Git boundaries** (`skills/myflow-contracts/pipeline.md`) already documents.

- [x] **Step 3: Add the push step and its mark**

A new numbered step after step 8, bracketed by `finish.push-archive` — begin and end, carrying the
run's own session token and the harness actually running the mark, per **Stage marks**
(`skills/myflow-contracts/pipeline.md`).

- [x] **Step 4: Update step 8's invocation and shell**

Pass the repository root as `gather-self-review-context.sh`'s fourth argument. Change the report
commit shell so it asserts `chore/archive-<name>` and does **not** push — the push step now carries
it. State that a branch mismatch reports and stops the commit, leaving the change `FINISHED`.

- [x] **Step 5: Update the terminal handoff block**

The `## Finished` block gains a line naming the archive pull request, or — when the push failed —
the local archive branch and the exact commands to land it by hand. It must never report the archive
as landed when it was not.

- [x] **Step 6: Add the guard to the presence check**

`prepare-archive-branch.sh` joins the guards this command's own text names, so the guard-presence
check at the start of the run covers it.

- [x] **Step 7: Verify**

```bash verified:run against this tree at e61603b; all five guards pass on the untouched tree
scripts/check-contract-budget.sh
scripts/check-references.sh
scripts/check-installed-citations.sh
scripts/check-guard-symlinks.sh
scripts/check-stage-mark-calls.sh
```

**Tests:** none added; covered by the five guards in step 7 — the stage-mark-calls guard in
particular, which rejects an undocumented stage key and a substituted session token. Its name is
written unbacktracked here on purpose: a backticked token in this field is parsed as a test this
task's own commit adds, which is what the task-commit-fields guard correctly rejects.

**Regression:** Reverting this task leaves the contract describing one procedure and the skill
executing another — and whichever the agent reads first wins, which is non-determinism in the layer
that must be deterministic.

**Baseline:** before=27 after=27 contract files within budget; `myflow-finish/SKILL.md` stays under
39965 bytes.

**Commit:** `feat(kan-239-run-2-asserts-base-branch-and-archives-via-pr): run the archive branch, the push step and the self-review commit from the finish skill`

---

### 8 `skills/myflow-contracts/pipeline.md` — boundaries and the registry

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: tasks 6 and 7, whose procedure these two rules describe from the pipeline's side.
- Produces: the run-2 git boundary and the archive branch's registry row.

- [x] **Step 1: Update the run-2 rows in Git boundaries**

Run 2 commits the archive on `chore/archive-<name>`, never on the base branch, and pushes that
branch — never `<base>`.

- [x] **Step 2: Add the archive branch's registry row**

| Artifact | Created by | Location | Removed by |
|----------|-----------|----------|-----------|
| Archive branch | finish run 2 | the repository and `origin` | nothing in this pipeline — run 2 is terminal and the pull request outlives it |

The removal column says what is true. **An artifact no row accounts for is a defect in the
registry**, and a row claiming a cleanup that does not happen is worse than none — this repository
already carries five such branches. `design.md`'s open question `archive-branch-cleanup` records
that deciding what should remove them is a separate change.

- [x] **Step 3: Verify**

```bash verified:run against this tree at e61603b; all four guards pass on the untouched tree
scripts/check-contract-budget.sh
scripts/check-references.sh
scripts/check-installed-citations.sh
scripts/check-vocabulary.sh
```

**Tests:** none added; covered by the four guards in step 3.

**Regression:** Reverting this task leaves the pipeline's own git boundary table saying run 2 pushes
the base branch, which is the rule this whole change exists to remove, and leaves the archive branch
as an artifact no registry row accounts for.

**Baseline:** before=27 after=27 contract files within budget; `pipeline.md` stays under 55728 bytes.

**Commit:** `docs(kan-239-run-2-asserts-base-branch-and-archives-via-pr): bound run 2's git actions to the archive branch and register it`

---

### 9 `scripts/check-installed-citations.py` — the archive branch is a branch name, not a citation

**Build:** green

**Files:**
- Modify: `scripts/check-installed-citations.py`
- Modify: `scripts/test-check-installed-citations.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: a classifier that recognises `chore/archive-<name>` as a git branch name. Tasks 6, 7 and
  8 cannot report a clean `check-installed-citations.sh` until this lands.

**This task was not in the original plan.** Task 6 discovered it: the guard already carries a
deliberately narrow exemption, `GIT_BRANCH_OPENSPEC_RE`, recognising `` `openspec/<name>` `` as a
**branch name** rather than a path citation — added by KAN-102's own task 10 for exactly this class
of token. This change introduces a second pipeline-created branch shape, `chore/archive-<name>`,
which has no such exemption, so all seven of its backticked occurrences in `finish-contract.md` are
reported as unrooted citations.

**The fix is to widen the exemption, never to strip the backticks.** Rooting a branch name teaches a
prefix that would be wrong wherever it appeared, which is the precise failure the existing exemption
was written to prevent; and de-backticking a branch name to dodge a guard is bypassing it, which this
project's lint policy forbids outright.

- [x] **Step 1: Add the failing cases first**

Two fixture cases in `scripts/test-check-installed-citations.sh`, in the shape its existing cases
use: a backticked `chore/archive-<name>` in an installed fixture file is **not reported**, and — the
bound that keeps the shape from failing open — `chore/archive-<name>/spec.md`, a bracket segment
followed by more path, **is** reported. Run the harness and confirm the first case fails.

- [x] **Step 2: Widen the shape, and keep it narrow**

Add a sibling regex beside `GIT_BRANCH_OPENSPEC_RE`, anchored at both ends exactly as it is, matching
`chore/` followed by a fixed branch-kind word and exactly one `<…>` placeholder segment and nothing
else, with `[^/<>]+` inside the brackets so the placeholder cannot smuggle a second `/`. Cover both
branch shapes this pipeline creates — `chore/archive-<…>` and `chore/self-review-<…>` — since the
repository already carries live branches of both kinds. Document it in the same comment style its
neighbour uses, naming this change as the reason.

- [x] **Step 3: Verify**

```bash verified:run against this tree at 35240f9; it reports the seven violations this task removes
scripts/check-installed-citations.sh
bash scripts/test-check-installed-citations.sh
```

**Tests:** the two fixture cases added in step 1.

**Regression:** Reverting this task makes every document that names the archive branch fail the
citation guard, which would push a later reader toward de-backticking a branch name to get a clean
run — the bypass this task exists to make unnecessary.

**Baseline:** before=7 after=0 violations from `scripts/check-installed-citations.sh`; the harness
goes before=50 after=52 `ok:` assertions.
<!-- measured: scripts/check-installed-citations.sh and bash scripts/test-check-installed-citations.sh @ 35240f9 -->

**Commit:** `fix(kan-239-run-2-asserts-base-branch-and-archives-via-pr): classify the pipeline's chore branch names as branches, not citations`

---

### 10 Stale run-2 step numbers in every file that cites them

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/handoff-blocks.md`
- Modify: `skills/myflow-contracts/jira-integration.md`
- Modify: `README.md`
- Modify: `skills/myflow-finish/SKILL-rationale.md`
- Modify: `skills/myflow-contracts/pipeline-rationale.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `stats/internal/stages/names.go`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 6, which renumbered run 2 by inserting the positioning step.
- Produces: a corpus in which no file cites a run-2 step number that no longer means what it says.

**This task was not in the original plan.** Task 6 inserted a step before the sync, shifting every
later run-2 step number by one, and reported the cross-references outside its own `**Files:**`
boundary that it therefore could not fix. `skills/myflow-finish/SKILL.md`'s own references belong to
task 7 and are **not** in scope here.

A stale step number is worse than a missing one: it resolves to a real step and describes the wrong
one.

- [x] **Step 1: Re-derive the mapping from the committed contract**

Read run 2's numbered steps in `skills/myflow-contracts/finish-contract.md` as task 6 left them, and
write down the old→new mapping. Do **not** apply a blanket `+1`: verify each citation against what
the step it names now actually says, since a citation may already have been correct.

- [x] **Step 2: Update each reference**

Known sites, to be re-verified rather than trusted: `handoff-blocks.md`'s self-review and
`FINISHED`-write references; `jira-integration.md`'s self-review reference; `README.md`'s
self-review reference; `SKILL-rationale.md`'s self-review reference; `pipeline-rationale.md`'s
cleanup-verification references. Report any further site you find that this list does not name.

- [x] **Step 3: Reword `finish.commit-archive`'s prose Name, in both places that carry it**

Task 7 moved the push out of that stage — it now brackets the commit alone — but its Name in
`README.md`'s Level 1 table and in `stats/internal/stages/names.go` still reads
`Commit and push the archive (run 2)`, which describes a step that no longer exists. Reword both to
name the commit only. **The Key is untouched**, exactly as task 5 left `finish.sync-archive`'s alone;
TestStagesMatchReadmeLevelOne requires the two files to agree, so they change together or the build
goes red.

- [x] **Step 4: Add `finish.push-archive` to `/myflow-fast`'s run-2 mark list**

`skills/myflow-fast/SKILL.md`'s **After merge-and-push specifically** section enumerates the run-2
stage marks a fast run fires, and describes the run in prose. Both are now wrong: the list omits
`finish.push-archive`, and the prose still says run 2 will *"commit and push the archive"*. Add the
mark in its run order — after `finish.self-review` — and correct the prose to match the contract's
step order. `/myflow-fast` chains `/myflow-finish`'s stages verbatim, so a stage missing from this
list is a stage a fast run never records.

- [x] **Step 5: Verify**

```bash verified:run against this tree at 9a8c7f0; all four pass before this task and must still pass after
scripts/check-references.sh
scripts/check-installed-citations.sh
scripts/check-contract-budget.sh
cd stats && go test ./internal/stages/... && gofmt -l .
```

**Tests:** none added. Step-number accuracy is not mechanically checkable — no guard resolves a prose
step number to a step — which is exactly why it is stated as its own task rather than left as a
side effect of task 6.

**Regression:** Reverting this task leaves five files pointing at run-2 steps that have moved, so a
reader following any of them lands on the wrong step.

**Baseline:** before=27 after=27 contract files within budget; no violation count changes, because no
guard checks this.

**Commit:** `docs(kan-239-run-2-asserts-base-branch-and-archives-via-pr): repoint every run-2 step citation at the renumbered contract`
