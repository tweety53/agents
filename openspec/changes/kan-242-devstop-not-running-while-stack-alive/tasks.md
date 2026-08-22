# Tasks — verify the stop before removing a worktree, and stop inferring `not running`

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** make `/myflow-finish` run 2 refuse to remove a worktree that still has a live process, and
make `gymie`'s `devStop` say `not running` only when it observed that, rather than inferring it from
records it could not read.

**Architecture:** one new project-agnostic guard in `agents` answering from `lsof`, wired in as a
worktree-cleanup check; one pure decision function extracted into `gymie`'s `buildSrc` so the
not-running decision becomes testable, following the shape `OrphanScan.kt` already established.

**Spec:** `openspec/changes/kan-242-devstop-not-running-while-stack-alive/design.md`, and
`specs/myflow-finish-cleanup/spec.md` beside it.

## Global constraints

- **Two repositories, two worktrees.** Tasks 1-3 land in `agents`; tasks 4-5 land in `gymie`. Each
  task's `**Files:**` paths are relative to that task's own repository root, named in the task.
- **The gymie half has no delta spec**, per decision `change-directory-in-agents` in `design.md`.
  Its contract is stated in that file and in the code's own documentation, and nothing mechanical
  checks it. Do not invent an `openspec/` delta in `gymie` to compensate — that repository's
  capability set is not this change directory's to write.
- **Never weaken a guard to make a task pass.** `.myflow/project.md`'s `## lint` list is the check
  set for `agents`; there is no auto-fix for `scripts/check-*`. `gymie` has its own lint commands in
  its own `.myflow/project.md`.
- **The new guard stays project-agnostic.** No port list, no build-tool invocation, nothing a project
  must declare. A guard that needed a project to opt in would not have protected the project this
  incident happened in.
- **`lsof` output is parsed, never trusted for existence.** A pid whose working directory cannot be
  read is "cannot prove", reported as an inability, never as a clean result.

The verdict grammar, fixed once here so tasks 1 and 2 cannot disagree:

```text unverified:the guard does not exist yet — this is the contract tasks 1 and 2 build against
CLEAR: <worktree>                       — no process has a working directory at or under it
HELD: <worktree> — <pid> <cwd>[; ...]   — one or more processes do
```

Exit 0 whenever a verdict was reached; exit 2 when the guard cannot answer at all — `lsof` absent,
or the worktree path unreadable. This mirrors `check-cleanup-complete.sh`'s split of verdict from
exit status, and task 3 relies on it: exit 2 is a failed check, never a pass.

---

### 1 Assertion harness for `check-worktree-processes.sh`

**Build:** red

**Squash-with:** Task 2

**Files:**
- Create: `scripts/test-check-worktree-processes.sh` *(repository: `agents`)*

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the executable statement of the verdict grammar above. Task 2 satisfies it.

Written before the guard so the cases assert what the delta spec's **Worktree cleanup verifies the
stack actually stopped, before removing anything** requires, rather than what the implementation
happens to print. Follow `scripts/test-check-workspace-isolation.sh`'s idiom exactly — the same
sandboxed `TMPDIR` fixtures, the same `pass`/`fail` helpers. Do not introduce a second harness shape.

- [x] **Step 1: A worktree with no process in it is `CLEAR`**

Fixture directory nothing is running from; assert `CLEAR:` and exit 0.

- [x] **Step 2: A process whose cwd is the worktree is `HELD`**

Start a `sleep` with its cwd set to the fixture, assert `HELD:` names that pid and that cwd, and that
exit is 0 — a verdict was reached, so the exit status is not the answer. Reap the process in
`trap`/cleanup so the harness never leaks one; this suite must not reproduce the incident it tests.

- [x] **Step 3: A process in a *subdirectory* of the worktree is `HELD`**

The spec says "at or under". A process started in `<worktree>/stats` is the ordinary case — a
backend started from its module root — and a prefix test that only matched the worktree root exactly
would miss every real one.

- [x] **Step 4: A sibling path sharing a name prefix is not `HELD`**

Fixture `<...>/openspec-kan-1` must not match a process running in `<...>/openspec-kan-12`. This is
the string-prefix defect the "at or under" wording exists to exclude; assert `CLEAR:`.

- [x] **Step 5: An unreadable worktree path exits 2**

Assert exit 2 and that no verdict token is printed, so a caller grepping for `CLEAR` in empty output
cannot read an inability as a pass.

- [x] **Step 6: `lsof` absent exits 2**

Run the guard with a `PATH` holding no `lsof`; assert exit 2 and a message naming the missing tool.

- [x] **Step 7 (KAN-197 mutation): a guard that always prints `CLEAR` fails this suite**

Copy the guard, replace its verdict computation with an unconditional `CLEAR:`, run the suite against
the copy, and assert the suite fails. A suite that cannot detect the guard's own central defect is
not a suite. Assert on the mutant's failure, and leave the real guard untouched.

- [x] **Step 8: Run the harness and confirm it fails**

```bash unverified:the guard does not exist yet, so the harness cannot pass at this task
scripts/test-check-worktree-processes.sh
```

Expected: FAIL — `check-worktree-processes.sh: No such file or directory`.

**Tests:** the cases enumerated in steps 1-7, in `scripts/test-check-worktree-processes.sh`.

**Regression:** Reverting this task leaves the guard with no executable statement of its contract, so
the sibling-prefix case and the exit-2 inability cases in particular become unverifiable — and those
are the two a hand-written guard gets wrong.

**Baseline:** `scripts/test-check-worktree-processes.sh` before=0 after=7 cases; the repository's
`## test` list gains one entry.

**Commit:** `test(scripts): assert the worktree live-process guard's verdicts and inabilities`

---

### 2 `check-worktree-processes.sh`

**Build:** green

**Files:**
- Create: `scripts/check-worktree-processes.sh` *(repository: `agents`)*
- Modify: `.myflow/project.md` — add the harness to `## test` and the guard to `## lint`

**Interfaces:**
- Consumes: task 1's harness, which must pass at the end of this task.
- Produces: the guard task 3 wires into worktree cleanup.

Answer from one `lsof -d cwd -Fpn` pass. Parse the `p<pid>` / `f cwd` / `n<path>` triples: `p` opens a
new record, `n` after an `f cwd` line is that pid's working directory.

```bash verified:measured in this repository on 2026-08-22 — 283 cwd rows, no sudo, on darwin 25.5.0
lsof -d cwd -Fpn 2>/dev/null
```
<!-- measured: lsof -d cwd -Fpn 2>/dev/null | grep -c '^n' @ /Users/tweety53/Projects/agents — 283 on re-measurement during implementation, against 290 when this plan was written. The count is point-in-time: it is however many processes happen to be running, so it varies between runs and between machines. Nothing in the guard depends on it; it is recorded only as evidence that one pass answers the question. -->

- [x] **Step 1: Resolve the worktree path to its physical form before comparing**

`lsof` reports physical paths — `/private/tmp`, not `/tmp`. Compare against the worktree resolved the
same way (`cd … && pwd -P`), or every fixture under `TMPDIR` on macOS reports `CLEAR` while holding a
process. Task 1 step 2 fails until this is right, which is why it is a step rather than a remark.

- [x] **Step 2: Match "at or under", not by string prefix**

A path matches when it equals the worktree or begins with the worktree plus `/`. Task 1 step 4 is the
case this exists for.

- [x] **Step 3: Print the verdict grammar from the global constraints above**

- [x] **Step 4: Exit 2 on every inability**

`lsof` not on `PATH`, a worktree path that is not a readable directory, or an `lsof` invocation that
itself failed. Report on stderr, print no verdict token, exit 2. An empty `lsof` result and a failed
`lsof` invocation must not be the same outcome — the guard reads the exit status, never the empty
output, exactly as `check-cleanup-complete.sh`'s survivor row does.

- [x] **Step 5: Do not require `sudo`, and say so in the header**

The guard sees the invoking user's own processes. That is the correct set: a worktree's stack is
started by the operator. Record the limit in the header rather than leaving it to be discovered —
a root-owned process from another user is outside what this guard can see.

- [x] **Step 6: Run the harness and the repository's lint**

```bash verified:both are .myflow/project.md's own commands for this repository
scripts/test-check-worktree-processes.sh
scripts/check-guard-symlinks.sh
```

Expected: PASS, and `GUARD-SYMLINKS-OK`.

**Tests:** no new test file; task 1's harness must pass in full.

**Regression:** Reverting this task removes the only mechanism that can observe a live process before
a removal, returning run 2 to the state that produced KAN-129 and KAN-242.

**Baseline:** `scripts/test-check-worktree-processes.sh` before=7 failing after=7 passing.

**Commit:** `feat(scripts): report processes running from a worktree`

---

### 3 Wire check 6 into worktree cleanup

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/finish-contract.md` *(repository: `agents`)*
- Modify: `skills/myflow-finish/SKILL.md` — name the guard in its presence check
- Modify: `skills/myflow-fast/SKILL.md` — state that the `/myflow-fast` check-4 override does not reach check 6
- Create: `skills/myflow-finish/scripts/check-worktree-processes.sh` — a tracked relative symlink to
  `scripts/check-worktree-processes.sh`
- Create: `skills/myflow-fast/scripts/check-worktree-processes.sh` — the same symlink, for the
  command that delegates to `/myflow-finish`'s cleanup

A tracked symlink under `skills/<skill>/scripts/` is how this repository installs a guard into a
skill directory — every entry there is one, checked into the index as mode `120000` and validated by
`scripts/check-guard-symlinks.sh`'s rule 1. Without them the presence check in step 4 names a guard
the installed skill does not carry, so the two paths are declared here rather than covered by a
widened `**Allowed-collateral:**`: a glob over this directory would let a future undeclared file in
it through, which is the opposite of what that field is for.

**Interfaces:**
- Consumes: task 2's guard and its verdict grammar.
- Produces: the contract text the delta spec's requirements describe.

- [x] **Step 1: Add check 6 to the worktree-cleanup check list**

After check 5's `## stop` command, before removal. It runs whether or not `## stop` is declared and
whatever that command exited — the delta spec's fourth scenario is exactly this.

- [x] **Step 2: State that `HELD:` and exit 2 are both failed checks**

They join checks 1, 2, 3 and 5 under "Any failed check leaves every worktree alone". Report the pids,
their working directories, and `ps -o pid,command -p <pid>`.

- [x] **Step 3: Correct the check-number error in the same block**

The bullet reading "This includes check 4: a `## stop` command that exits non-zero…" names the wrong
check — the `## stop` command is check 5; check 4 is the ignored-files disclosure. Fix the number.
Do not reword the sentence around it: the normative content is correct and only the number is wrong.

- [x] **Step 4: Name the guard in `/myflow-finish`'s guard presence list**

So an absent guard prints the guard-presence block and the check is performed by hand, rather than
silently skipped.

- [x] **Step 5: Scope `/myflow-fast`'s check-4 override away from check 6**

Its guardrail already says checks 1, 2, 3 and 5 remain gates. Add check 6 to that list explicitly,
per the delta spec's **A blocked removal is a failure, never a confirmable disclosure**.

- [x] **Step 6: Install and verify the citations resolve**

```bash verified:setup.sh global and the lint list are this repository's own commands
HOME="$(mktemp -d)" ./setup.sh global
scripts/check-references.sh
scripts/check-installed-citations.sh
scripts/check-contract-budget.sh
```

Expected: all exit 0. `check-contract-budget.sh` will fail if `finish-contract.md` outgrows its
budget row — raise that row deliberately, and never narrow the guard.

**Tests:** no new test file; the four guards above are the verification.

**Regression:** Reverting this task leaves the guard present but unreferenced, so run 2 never calls it
and the gate does not exist in practice.

**Baseline:** `scripts/check-references.sh`, `check-installed-citations.sh`, `check-contract-budget.sh`
before=exit 0 after=exit 0.

**Commit:** `feat(finish-contract): gate worktree removal on a stopped stack`

---

### 4 `StopOutcomeTest.kt`

**Build:** red

**Squash-with:** Task 5

**Files:**
- Create: `buildSrc/src/test/kotlin/com/gymie/gradle/StopOutcomeTest.kt` *(repository: `gymie`)*

**Interfaces:**
- Consumes: `Recorded` from `StateFile.kt`, unchanged.
- Produces: the executable statement of the not-running decision. Task 5 satisfies it.

`stopService` lives in `gradle/dev-lifecycle.gradle.kts` and cannot be driven under test. Follow the
split `OrphanScan.kt` already established in this repository — the decision moves into `buildSrc` as
a pure function over values, and the I/O stays in the `.gradle.kts`. Assert the decision, not the
log text.

- [x] **Step 1: Both halves definitely negative is `NotRunning`**

`Recorded.None` for the pid and a port that resolved with no listener.

- [x] **Step 2: An unreadable pid record is not `NotRunning`**

`Recorded.Unknown` for the pid yields the "cannot tell" outcome naming the pid record. This is the
exact collapse the current code makes and `StateFile.kt`'s own documentation warns against.

- [x] **Step 3: An unresolved port is not `NotRunning`**

A null port — which an isolated workspace produces for an unusable port record, deliberately — yields
"cannot tell" naming the port. This is the second half of the incident: `sweepPortFor` refusing to
guess is correct, and reading its refusal as absence is not.

- [x] **Step 4: Both halves unknown names both**

The operator gets both records, not the first one found.

- [x] **Step 5: A live tracked pid is `Stopped`, not `NotRunning`**

Guards the ordinary success path against a fix that makes everything "cannot tell".

- [x] **Step 6: A resolved port with a listener is not `NotRunning`**

- [x] **Step 7: Run the tests and confirm they fail**

```bash unverified:the type does not exist yet, so the module will not compile at this task
./gradlew :buildSrc:test --tests '*StopOutcomeTest*'
```

Expected: FAIL — unresolved reference to the new type.

**Tests:** the cases enumerated in steps 1-6, in
`buildSrc/src/test/kotlin/com/gymie/gradle/StopOutcomeTest.kt`.

**Regression:** Reverting this task leaves the not-running decision with no executable statement, so
the two unknown-record cases — the ones that produced this incident — become unverifiable again.

**Baseline:** `buildSrc` before=199 after=205 `@Test` functions.
<!-- measured: for f in buildSrc/src/test/kotlin/com/gymie/gradle/*.kt; do grep -c '@Test' $f; done | paste -sd+ | bc @ /Users/tweety53/Projects/gymie, branch main -->

**Commit:** `test(dev-lifecycle): assert that not-running requires a definite negative`

---

### 5 `StopOutcome.kt`, and `stopService` reporting from it

**Build:** green

**Files:**
- Create: `buildSrc/src/main/kotlin/com/gymie/gradle/StopOutcome.kt` *(repository: `gymie`)*
- Modify: `gradle/dev-lifecycle.gradle.kts` — `stopService` calls the new decision

**Interfaces:**
- Consumes: task 4's tests, which must pass at the end of this task.
- Produces: the third outcome the design's `not-running-requires-a-definite-negative` decision names.

- [x] **Step 1: Model the outcome as three states, not two**

`Stopped`, `NotRunning`, and one carrying which halves were unknown. Three, for the same reason
`Recorded` has three: collapsing "I could not tell" into "there is nothing there" is the defect.

- [x] **Step 2: `NotRunning` requires `Recorded.None` for the pid and a resolved, unheld port**

Not `trackedPid == null`, which is satisfied by `Recorded.Unknown` too.

- [x] **Step 3: Take the raw `Recorded` values, not the collapsed ones**

The function must receive `records.pid` itself. Passing `trackedPid` in would carry the collapse
across the new boundary and every test in task 4 would pass over the same defect.

- [x] **Step 4: Wire `stopService` to report from the outcome**

The "cannot tell" branch names each unknown half and the command that answers it —
`ps -o pid,command -p <pid>` and `lsof -i tcp:<port>` — matching the vocabulary
`unusableRecordsReport` already uses. Keep the existing survivors branch and the existing
record-retention behaviour untouched; this task changes what the run *says*, and changes what it
deletes not at all.

- [x] **Step 5: Run the tests and the repository's lint**

```bash unverified:gymie's own .myflow/project.md is canonical for its lint and test commands — read it in the worktree
./gradlew :buildSrc:test --tests '*StopOutcomeTest*'
```

Expected: PASS. Then run gymie's full `## lint` list from its own `.myflow/project.md`; do not assume
this repository's commands apply there.

**Tests:** no new test file; task 4's tests must pass in full.

**Regression:** Reverting this task returns `devStop` to reporting `not running` from records it could
not read, which is the root cause KAN-242 records.

**Baseline:** `buildSrc` before=205 failing after=205 passing.

**Commit:** `fix(dev-lifecycle): report not-running only when it was observed`
