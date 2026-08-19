> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Base-branch resolution stops being a fenced block that every `/myflow-finish` run retypes,
and becomes one guard every consumer invokes — including `/myflow-status`, which today hand-derives
the same value with no assertion at all.

**Architecture:** One new guard, `scripts/resolve-base-branch.sh`, plus its harness. It prints the
bare branch name on stdout and separates its refusals by exit code. Three relative symlinks ship it
into the skills that invoke it. The finish contract's fenced block becomes a one-line invocation and
keeps its prose. `/myflow-status` acquires its first guard, which reverses three statements in the
repository that currently assert it has none.

**Tech Stack:** Bash (guard + harness, Bash 3.2 floor, as `scripts/test-check-finish-preflight.sh`'s
header records for this repository) and Markdown (the contract and skill text). Nothing is added to
either.

**Spec:** `openspec/changes/kan-236-base-branch-resolution-retyped-by-hand/design.md` and
`specs/myflow-base-branch-resolution/spec.md` in the same directory. The approved design is
`docs/superpowers/specs/2026-08-19-kan-236-base-branch-resolution-retyped-by-hand-design.md`.

## Global Constraints

- **`check-finish-preflight.sh`'s signature does not change.** It keeps taking `<base-ref>` as its
  second argument and resolves no base branch itself. Only its header comment is edited. A task that
  adds a `-` self-resolution mode has exceeded its scope — that route is decision `standalone-guard`
  in `design.md`, explicitly rejected.
- **The third assertion — base ≠ current branch — is unconditional.** No flag, no environment
  variable, no mode disables it. This is decision `assert-always-pass-worktree`.
- **The fetch stays inside the guard**, with its existing wrapping (`-c core.askpass=true`,
  `2>/dev/null`, `|| true`). This is decision `fetch-inside`.
- **The finish contract keeps its prose.** The `HEAD@{upstream}` paragraph and the no-remote
  paragraph are the reasoning, not the copied code. Only the fenced block goes.
- **A guard is named by basename in every invoking position**, never as a repository-relative
  `scripts/<name>` path — `scripts/check-guard-symlinks.sh`'s rule 3 rejects the latter.
- **Every skill symlink is relative** (`../../../scripts/<name>`); rule 1 rejects an absolute
  target.
- **No task edits `openspec/` or `docs/superpowers/`.**
- **Tasks 1 and 2 are one unit.** Task 1's harness tests a guard that does not exist yet.
- **Task 6 does both halves at once.** The moment `skills/myflow-status/SKILL.md` cites the guard,
  `check-guard-symlinks.sh`'s `declare_if_present "myflow-status"` becomes false; leaving them in
  separate tasks would ship an intermediate commit whose coverage report lies.

## Baseline

All measured 2026-08-19 against `6b4021d`.

- `scripts/check-guard-symlinks.sh` reports
  `GUARD-SYMLINKS-OK: /Users/tweety53/Projects/agents — 58 guard(s) across 6 skill(s) validated`,
  with `myflow-finish 6`, `myflow-fast 17`, and `myflow-status 0 (declared: invokes no guard — a
  read-only status report; …)`.
  <!-- measured: scripts/check-guard-symlinks.sh @ 6b4021d -->
- `scripts/check-contract-budget.sh` reports `BUDGET-OK: 27 contract file(s) within budget`.
  <!-- measured: scripts/check-contract-budget.sh @ 6b4021d -->
- `skills/myflow-contracts/finish-contract.md` is 32662 bytes against a budget row of 40724 — 8062
  bytes of headroom.
  <!-- measured: wc -c skills/myflow-contracts/finish-contract.md and grep the budgets table in scripts/check-contract-budget.sh @ 6b4021d -->
- `skills/myflow-finish/SKILL.md` is 34505 bytes against a row of 39965 — 5460 bytes of headroom.
  <!-- measured: wc -c skills/myflow-finish/SKILL.md and grep the budgets table in scripts/check-contract-budget.sh @ 6b4021d -->
- `skills/myflow-status/SKILL.md` is 18041 bytes against a row of 22551 — 4510 bytes of headroom,
  the tightest of the three relative to what task 6 adds.
  <!-- measured: wc -c skills/myflow-status/SKILL.md and grep the budgets table in scripts/check-contract-budget.sh @ 6b4021d -->
- `scripts/resolve-base-branch.sh` and `scripts/test-resolve-base-branch.sh` do not exist.
  <!-- measured: ls scripts/resolve-base-branch.sh scripts/test-resolve-base-branch.sh @ 6b4021d -->

---

### 1 `scripts/test-resolve-base-branch.sh` — the harness, before the guard exists

**Build:** red

**Squash-with:** Task 2

**Files:**
- Add: `scripts/test-resolve-base-branch.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the executable definition of `resolve-base-branch.sh`'s contract, which task 2 then
  satisfies. Every later task depends on that contract holding.

The guard's whole value is that it refuses correctly. A harness written after the guard tends to
assert what the guard happens to do; written first, it asserts what the design says it must do.

Model the harness on `scripts/test-check-finish-preflight.sh`: throwaway git repositories under
`TMPDIR`, one `ok:`/`FAIL:` line per assertion, a `resolve-base-branch: all cases pass` line at the
end, and a non-zero exit on any failure. Bash 3.2 is the floor, as that harness's header records for
this repository.

Every case asserts **three things together** — the exit status, stdout, and that stderr names the
failure — because a guard that exits correctly while printing a stale name on stdout is the failure
mode this whole change exists to remove.

- [x] **Step 1: Build the fixture helpers**

A helper that creates a bare "remote" repository with a `main` branch, clones it, sets
`refs/remotes/origin/HEAD`, and checks out a branch named `openspec/fixture`. A second helper that
produces the same clone with `refs/remotes/origin/HEAD` deliberately deleted, so the
`git remote show origin` fallback is the only path left.

```bash unverified:confirm `git remote set-head origin -d` is the deletion form on this machine's git
git -C "$clone" remote set-head origin -d      # force the fallback path
```

- [x] **Step 2: Write the cases**

| Case | Expects |
|------|---------|
| Clone with `origin/HEAD` set | stdout exactly `main`, stderr empty, exit `0` |
| `origin/HEAD` deleted, `git remote show origin` answers | stdout exactly `main`, exit `0` |
| Detached `HEAD` | exit `1`, stderr names the detached `HEAD`, stdout empty |
| Base resolves to the checked-out branch (`main` checked out) | exit `1`, stderr names the self-comparison, stdout empty |
| Neither resolution path answers | exit `1`, stderr names the unresolved base, stdout empty |
| No `origin` remote configured | exit `3`, stderr names the missing remote, stdout empty |
| `<dir>` is a plain directory, not a git worktree | exit `2`, stdout empty |
| `<dir>` does not exist | exit `2`, stdout empty |
| No argument at all | exit `2`, usage line on stderr |
| Remote reports a `HEAD branch:` value of `-x` | exit `1`, stdout empty |
| Remote reports a `HEAD branch:` value carrying a control character | exit `1`, stdout empty |

The last two cases build the hostile value by writing the `HEAD branch:` line into a fake
`git remote show` — the real command will not produce one — so they exercise the validation step and
nothing else.

The unreachable-remote hang is deliberately **not** covered. A wall-clock assertion on a TCP timeout
is flaky by construction; the wrapping is carried forward verbatim and its reason recorded in the
guard's header instead.

- [x] **Step 3: Mutation-prove every case**

Per KAN-197, a case whose assertion still passes with the line it covers removed proves nothing.
Once task 2 lands, delete each of the guard's three assertions and its validation step in turn and
confirm the corresponding case fails; restore, and record the four mutations in the task's own SDD
ledger entry.

- [x] **Step 4: Verify**

```bash verified:run it and read the failure — the guard does not exist yet, so every case must fail for that reason and no other
bash scripts/test-resolve-base-branch.sh; echo "exit=$?"
```

**Tests:** the cases enumerated in step 2, in `scripts/test-resolve-base-branch.sh`.

**Regression:** Reverting this task leaves `resolve-base-branch.sh` with no executable statement of
its contract. Its three assertions are the only thing standing between a misresolved base and run 2
`--force`-removing a worktree holding unmerged work — a rule with no test is exactly how the fenced
block drifted in the first place.

**Baseline:** before=0 after=13 cases / 39 `ok:` assertions in `scripts/test-resolve-base-branch.sh`.
<!-- measured: bash scripts/test-resolve-base-branch.sh @ 16f5324 gave 13 cases and 39 ok: assertions; the harness does not exist at 6b4021d, so before=0 -->

Eleven cases come from step 2's table. Two more were added by review rounds and are kept here rather
than folded into that table, so the table stays the record of what was planned: case 12 covers a
non-ASCII base name under a UTF-8 locale (the per-task review's `LC_ALL=C` finding), and case 13
covers an unreadable current-branch ref (panel finding F4).

**Commit:** `test(kan-236-base-branch-resolution-retyped-by-hand): pin resolve-base-branch.sh's exit contract and its three assertions`

---

### 2 `scripts/resolve-base-branch.sh` — the guard

**Build:** green

**Files:**
- Add: `scripts/resolve-base-branch.sh`
- Add: `scripts/test-resolve-base-branch.sh` — task 1's file. Task 1 is `Build: red` with
  `**Squash-with:** Task 2`, so its commit folds into this one and this commit is the squash unit's
  only commit. The field declares what the commit touches, which is both files.

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 1's harness, which defines the contract this task satisfies.
- Produces: `resolve-base-branch.sh <dir>` — the single owner of base-branch resolution, invoked by
  tasks 3-6's call sites.

`resolve-base-branch.sh <dir>` takes exactly one argument, the directory to resolve in. stdout
carries the bare branch name and **nothing else**, so a caller uses it in a command substitution
with no parsing. Every refusal writes its reason to stderr and leaves stdout empty — an empty stdout
must never be readable as a resolved base.

| Exit | Meaning |
|------|---------|
| `0` | Resolved; the name is on stdout |
| `1` | A named refusal — detached `HEAD`, no base resolved, base equal to the current branch, or a base name that fails validation |
| `2` | Cannot answer — the argument is missing, or `<dir>` is absent, unreadable, or not a git worktree |
| `3` | The repository has no `origin` remote at all |

- [x] **Step 1: Write the header comment**

The header carries what the fenced block's comments carried, because that reasoning is the reason
the code is shaped this way and is what stops the next reader "simplifying" it:

- why both network calls are wrapped — `git remote show origin` against an unreachable host blocks
  for roughly 75s on the default TCP timeout, which would turn a correct refusal into a two-minute
  hang;
- why `HEAD@{upstream}` is never consulted — inside an apply worktree `HEAD` *is* `openspec/<name>`,
  so that fallback compares the branch with its own upstream, which is true the moment the branch is
  pushed, and run 2 then archives an unmerged change and `--force`-removes its worktree;
- why exit `3` is separate from exit `2` — the finish contract requires a distinct no-remote message,
  and a caller that could only tell by matching stderr text would stop being able to the first time
  the wording is edited.

- [x] **Step 2: Implement, in this order**

```bash unverified:confirm `git branch --show-current` prints empty (not an error) on a detached HEAD in this machine's git
DIR="${1:-}"                                     # missing  -> exit 2, usage on stderr
[ -d "$DIR" ] || exit 2
git -C "$DIR" rev-parse --git-dir >/dev/null 2>&1 || exit 2
git -C "$DIR" remote get-url origin >/dev/null 2>&1 || exit 3
git -C "$DIR" -c core.askpass=true fetch --quiet origin 2>/dev/null || true
BASE="$(git -C "$DIR" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's#^origin/##')"
[ -n "$BASE" ] || BASE="$(GIT_TERMINAL_PROMPT=0 git -C "$DIR" remote show origin 2>/dev/null | sed -n 's/^ *HEAD branch: //p')"
CUR="$(git -C "$DIR" branch --show-current 2>/dev/null)"
[ -n "$CUR" ]        || exit 1                   # detached HEAD
[ -n "$BASE" ]       || exit 1                   # no base resolved
[ "$BASE" != "$CUR" ] || exit 1                  # refusing to compare a branch with itself
```

Steps 5-7 of that sequence keep the fenced block's own order, so each refusal reports the same
failure a reader of the old block would expect, in the same sequence.

- [x] **Step 3: Validate the name before printing it**

`BASE` arrives from the remote and flows into refs a caller builds (`origin/$BASE`) and into git
commands. Refuse with exit `1`, printing nothing to stdout, unless it matches
`^[A-Za-z0-9._][A-Za-z0-9._/-]*$` — which rejects an empty value, a leading `-` a downstream git
call would read as an option, and any control character.

- [x] **Step 4: `set -euo pipefail`, and make it executable**

Match the house style of `scripts/check-finish-preflight.sh`. Every `git` invocation that takes a
ref this script did not choose itself passes it after `--end-of-options`.

- [x] **Step 5: Verify**

```bash verified:run both — the harness goes green, and the guard suite stays green
bash scripts/test-resolve-base-branch.sh
scripts/check-guard-symlinks.sh
```

**Tests:** none of its own — task 1's harness is this task's test, and the two are dispatched and
committed as one unit.

**Regression:** Reverting this task removes the only implementation of the resolution, leaving every
call site with nothing to invoke.

**Baseline:** before=0 after=13 passing cases in `scripts/test-resolve-base-branch.sh`.
<!-- measured: bash scripts/test-resolve-base-branch.sh @ 16f5324 — the same cases task 1 adds, all passing once the guard exists -->

**Commit:** `feat(kan-236-base-branch-resolution-retyped-by-hand): resolve the base branch in one guard instead of a retyped block`

---

### 3 Ship the guard into the three skills that invoke it

**Build:** green

**Files:**
- Add: `skills/myflow-finish/scripts/resolve-base-branch.sh` (symlink)
- Add: `skills/myflow-fast/scripts/resolve-base-branch.sh` (symlink)
- Add: `skills/myflow-status/scripts/resolve-base-branch.sh` (symlink, new directory)
- Modify: `.myflow/project.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: `scripts/resolve-base-branch.sh` from task 2.
- Produces: the symlinks tasks 5 and 6's citations require, and the harness entry that makes the new
  test part of the project's suite.

Shipping the symlinks **before** the citations is deliberate and keeps every commit green:
`check-guard-symlinks.sh`'s rule 2 only ever flags a citation with no corresponding symlink, never a
symlink with no detected citation.

- [x] **Step 1: Create the three relative symlinks**

```bash verified:run it, then confirm each link resolves and its target is relative
ln -s ../../../scripts/resolve-base-branch.sh skills/myflow-finish/scripts/resolve-base-branch.sh
ln -s ../../../scripts/resolve-base-branch.sh skills/myflow-fast/scripts/resolve-base-branch.sh
mkdir -p skills/myflow-status/scripts
ln -s ../../../scripts/resolve-base-branch.sh skills/myflow-status/scripts/resolve-base-branch.sh
```

`/myflow-fast` gets one because it carries the union of `/myflow-do`'s, `/myflow-finish`'s and
`/myflow-status`'s guards, which its own delegating **Check guard presence.** paragraph already
declares.

- [x] **Step 2: Add the harness to the project's test list**

Add `scripts/test-resolve-base-branch.sh` to `.myflow/project.md`'s `## test` section, beside the
other `test-check-*.sh` entries. Do not add a count to the note below that list — it deliberately
cites none.

- [x] **Step 3: Verify**

```bash verified:both must exit clean; the symlink report's total rises by three
scripts/check-guard-symlinks.sh
scripts/check-references.sh
```

**Tests:** none — this task ships an existing guard and edits a configuration list.

**Regression:** Reverting this task makes every citation added by tasks 5 and 6 a rule-2 violation,
and drops the new harness out of the project's test suite, where nothing would run it again.

**Baseline:** before=60 after=60 guards in `check-guard-symlinks.sh`'s verdict line — **unchanged**.
That top-line count is `find "$SCRIPTS_DIR" -mindepth 1 -maxdepth 1` over the **root** `scripts/`
directory, not a count of symlinks under `skills/*/scripts/`; this task adds three symlinks under
skill directories and nothing to root `scripts/`, so it cannot move. It was 58 at `6b4021d` and
became 60 when tasks 1-2 added the guard and its harness to root `scripts/`. What this task changes
is that three skill `scripts/` directories gain an entry, which rule 1 validates; the per-skill
required-guard breakdown moves in tasks 5 and 6, driven by citations rather than by symlinks.
<!-- measured: scripts/check-guard-symlinks.sh @ b535841 gave 60, and again @ 9212ecf gave 60 -->

**Commit:** `chore(kan-236-base-branch-resolution-retyped-by-hand): ship resolve-base-branch.sh into the skills that invoke it`

---

### 4 The finish contract states the invocation, not the sequence

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/finish-contract.md`
- Modify: `scripts/check-finish-preflight.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: the guard from task 2 and its symlinks from task 3.
- Produces: the contract text tasks 5 and 6 point at for the exit contract.

- [x] **Step 1: Replace the fenced block**

Under **Run 1 — the branch is not merged**, the ten-line block that fetches, resolves and asserts
becomes:

```bash unverified:confirm the surrounding paragraph reads correctly once the block is this short
BASE="$(resolve-base-branch.sh "<abs-worktree>")"
```

State the exit contract beside it — `0` resolved, `1` a named refusal, `2` cannot read the tree, `3`
no `origin` remote — and name which directory a caller passes: **the apply worktree**, where `HEAD`
is the change's own branch by construction, never the main checkout.

Name the guard by **basename**. A repository-relative `scripts/resolve-base-branch.sh` in this
invoking position is a rule-3 violation.

- [x] **Step 2: Keep the prose, re-pointed**

The `HEAD@{upstream}` paragraph and the no-remote paragraph stay. Both are rewritten to name
`resolve-base-branch.sh` as the place the rule is now enforced, rather than describing a block that
is no longer there. The no-remote paragraph additionally says that exit `3` is how a caller
recognises the case.

- [x] **Step 3: Re-point `check-finish-preflight.sh`'s header**

Its comment currently reads that base-branch resolution "deliberately stays in pipeline.md's Finish
contract and is passed in." That is now wrong about *where* and right about *why*. Rewrite it to
name `resolve-base-branch.sh` as the one owner, keeping the reason — a second copy could drift from
it. **Change nothing else in this file**: the signature, the three signals and every exit path stay
exactly as they are.

- [x] **Step 4: Verify**

```bash verified:all three must exit clean; the preflight harness must still pass untouched
scripts/check-references.sh
scripts/check-contract-budget.sh
bash scripts/test-check-finish-preflight.sh
```

**Tests:** none — this task edits a contract and a comment.

**Regression:** Reverting this task restores the fenced block, and with it the retyping this change
exists to remove; the guard would then be shipped but invoked by nobody in the finish path.

**Baseline:** `skills/myflow-contracts/finish-contract.md` has 8062 bytes of headroom against its
budget row; this task removes more than it adds.
<!-- measured: wc -c skills/myflow-contracts/finish-contract.md and grep the budgets table in scripts/check-contract-budget.sh @ 6b4021d -->

**Commit:** `refactor(kan-236-base-branch-resolution-retyped-by-hand): the finish contract invokes the resolver instead of stating the sequence`

---

### 5 `/myflow-finish` declares and invokes the guard

**Build:** green

**Files:**
- Modify: `skills/myflow-finish/SKILL.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: the symlinks from task 3 and the contract text from task 4.
- Produces: the rule-2 citation that binds `skills/myflow-finish/scripts/resolve-base-branch.sh`,
  and — through the delegating paragraph — the one in `skills/myflow-fast/scripts/`.

- [x] **Step 1: Add the guard to the presence paragraph**

Its **Check guard presence.** paragraph enumerates the guards this command invokes; add
`resolve-base-branch.sh` to that list. This is what makes rule 2 compute it as required for both
`myflow-finish` and, through delegation, `myflow-fast`.

- [x] **Step 2: Re-point the base-resolution paragraph**

The paragraph beginning *"The base branch is resolved, never assumed, and never derived from the
current branch"* keeps its wording and its warning; it now names the guard rather than sending the
reader to the contract for a block to retype. Say that run 2 resolves the base **from the apply
worktree, before cleanup removes it** — cleanup is run 2's step 4 and the archive commit is step 3,
so the worktree is still there.

- [x] **Step 3: Verify**

```bash verified:both must exit clean, and the coverage line must show myflow-finish and myflow-fast one higher
scripts/check-guard-symlinks.sh
scripts/check-references.sh
```

**Tests:** none — this task edits skill documentation.

**Regression:** Reverting this task leaves `/myflow-finish` with a guard it ships but never declares,
so a missing-guard install would go unreported at exactly the command where a wrong base branch is
destructive.

**Baseline:** before `myflow-finish 6` / `myflow-fast 17`, after `myflow-finish 7` / `myflow-fast 18`
in `check-guard-symlinks.sh`'s coverage line.
<!-- predicted: one guard added to each skill's required set, against the counts measured at 6b4021d -->

**Commit:** `docs(kan-236-base-branch-resolution-retyped-by-hand): /myflow-finish declares and invokes the base-branch resolver`

---

### 6 `/myflow-status` acquires its first guard

**Build:** green

**Files:**
- Modify: `skills/myflow-status/SKILL.md`
- Modify: `scripts/check-guard-symlinks.sh`
- Modify: `scripts/test-check-guard-symlinks.sh` — its F4c fixture reuses the name `myflow-status`
  as a stand-in and relies on the real `declare_if_present` call to keep its legitimately-empty
  required set from tripping the unrelated coverage violation. Removing that declaration breaks
  the fixture, so the stand-in moves to another declared expected-zero skill. Step 4 below is what
  finds this; the field declares it because the break is a consequence of this task, not collateral.

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: the symlink from task 3, the guard from task 2.
- Produces: nothing later in this plan consumes; this is the last task.

Both halves land together. The moment the skill cites the guard, its required set is non-empty and
`declare_if_present "myflow-status"` — which exists to keep a legitimate zero from reading as an
uncomputed empty set — becomes false. Splitting them would ship a commit whose coverage report lies.

- [x] **Step 1: Merge-status step 3 invokes the guard**

Step 3 currently runs `git merge-base --is-ancestor <branch> origin/<base>` and names no source for
`<base>`. It now obtains it from `resolve-base-branch.sh`, run in the worktree the step is already
visiting — which is also what satisfies the unconditional third assertion, since `HEAD` there is the
change's own branch.

- [x] **Step 2: A refusal is inconclusive, never a stop**

Say explicitly that a non-zero exit makes that worktree's merge status **inconclusive** — the same
disposition step 3 already gives a git failure or an unresolvable base ref — that the report
continues, and that the resolver's own stderr message is what the detail view reports. `/myflow-status`
is read-only and never blocks.

- [x] **Step 3: Reverse the three no-guard claims**

1. Replace the "No guard-presence check here — this command invokes no guard" block with a real
   **Check guard presence.** paragraph naming `resolve-base-branch.sh` and
   `skills/myflow-status/scripts/`. Use the same paragraph opening the other skills use — the
   delegation scan in `check-guard-symlinks.sh` keys on `**Check guard presence.**` at the start of
   a line.
2. Replace the "This command therefore carries no `scripts/` directory" sentence; the directory
   exists as of task 3, and the reasoning that a directory nothing resolves against goes stale no
   longer applies to it.
3. Remove `check-guard-symlinks.sh`'s `declare_if_present "myflow-status" "invokes no guard — …"`
   call, and any file-header prose in that script asserting `/myflow-status` invokes nothing. Leave
   the `myflow-start` and `openspec-explore` declarations untouched — both are still true.

- [x] **Step 4: Verify**

```bash verified:every one must exit clean, and myflow-status must appear with a non-zero count and no "(declared)" marker
scripts/check-guard-symlinks.sh
bash scripts/test-check-guard-symlinks.sh
scripts/check-references.sh
scripts/check-contract-budget.sh
```

`scripts/test-check-guard-symlinks.sh` builds its own sandbox fixtures under `TMPDIR` and one of them
is *named* `myflow-status` as a stand-in. That fixture is independent of this repository's real tree
— check whether any of its assertions read the real declaration, and fix only what genuinely breaks.

**Tests:** none of its own — `scripts/test-check-guard-symlinks.sh` already covers the declaration
mechanism and must keep passing.

**Regression:** Reverting this task returns `/myflow-status` to hand-deriving the base branch with no
assertion that it differs from the branch under test — the worse half of the defect, in the command
run most often.

**Baseline:** before `myflow-status 0 (declared: …)`, after `myflow-status 1` in
`check-guard-symlinks.sh`'s coverage line.
<!-- predicted: one required guard for myflow-status, against the declared zero measured at 6b4021d -->

**Commit:** `feat(kan-236-base-branch-resolution-retyped-by-hand): /myflow-status resolves its base branch through the guard`
