# self-review-kan-1007-list-mode-workout-mode-row-in-settings

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close kan-1007's eight self-review findings — F1 (Task 1) through F8 (Task 8), as
`proposal.md` names them.

**Architecture:** one contract carve-out; one shared argument walk in the main-checkout hook; a new
Go guard called from **Verify**; a CLI refusal plus a parent-run store repair; three prose rules;
a second check in the sidecar guard.

**Tech Stack:** Python 3 (`hooks/`), Bash, Go (`stats/`), Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-22-self-review-kan-1007-list-mode-workout-mode-row-in-settings-design.md`
in this worktree.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1007-list-mode-workout-mode-row-in-settings` (`<wt>` below), on branch `self-review-kan-1007-list-mode-workout-mode-row-in-settings`; every `**Files:**` path is relative to it. Never touch the main checkout and never create a worktree.
- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Prose added to a run-loaded file is the exact text a step quotes, so `check-verbatim-moves.sh` matches it against this change's `verbatim-moves.txt`; copy it byte for byte.
- Go: `gofmt -w` every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Never widen a timeout, retry or tolerance to make a test pass (`rules/fix-determinism-at-the-source.mdc`).
- Never stop or restart `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container, and never write the flow store: Task 4's Step 5 is the parent's, not the fixer's.

## Review Focus

- Task 2: `git commit -m ''` (an empty word) must not raise in the hook; a push with arguments before a redirect still reads every refspec.
- Task 3: a worktree with no recorded base, or whose base is the tracked branch, is never `BASE-BEHIND`; the printed first command actually syncs the fixture's base.
- Task 4: a back-filled panel row (`-started-at`) and a non-panel reviewer key are admitted over an unticked task; a missing `tasks.md` is admitted.
- Task 7: a sidecar line with one field is not judged on its frame; a mockups root that is not a directory is exit 2 with empty stdout.

- [ ] 1. Carve the missing-base creation out of What still stops
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'The global rules require these confirmed' skills/flow-contracts/operator-prompts-auto-resolution.md` prints line 99 and `grep -c 'BASE-CREATED' skills/flow-contracts/operator-prompts-auto-resolution.md` prints 0, while `grep -n 'push -q origin "refs/remotes/origin/' scripts/integrate-change.sh` prints the unasked push.
    <!-- measured: grep -n 'The global rules require these confirmed' skills/flow-contracts/operator-prompts-auto-resolution.md @ 7d8f01ae → 99 -->
  - [ ] **Step 2: Implement** per decision `base-create-carve-out`. In `skills/flow-contracts/operator-prompts-auto-resolution.md`, after the line `  The global rules require these confirmed.`, add this line, indented two spaces as the bullet's continuation:

```text verified:authored in-tree for this change — the exact prose this step adds
One push is not asked: creating a `--base` a repository's origin lacks, from that repository's default branch — kickoff's (`skills/flow/SKILL.md`) and integrate's `prepare` (`BASE-CREATED:`, `skills/flow/integrate.md`) — adds a branch at the default branch's own tip, overwrites nothing, and is what the operator's `--base` already asked for.
```
  - [ ] **Step 3: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-verbatim-moves.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow-contracts/operator-prompts-auto-resolution.md`
**Tests:** none — a contract sentence; `check-verbatim-moves.sh` pins its text.
**Regression:** none — prose only.
**Commit:** docs(flow-contracts): carve the missing-base creation out of what still stops
**After:** none

**Decision:** base-create-carve-out

- [ ] 2. Stop the hook's push and commit argument walks at a redirect
  - [ ] **Step 1: Write the cases first.** In `scripts/test-protect-main-checkout.sh`: directly after case 104's `… pass "104 checker args" || fail …` line (the stub checker still answers 1 and `$CWT` is on `spectre/chg2`, which origin holds), add
    `expect "160 a bare change-branch push with its output redirected and piped" deny` over `git -C $CWT push -q 2>&1 | tail -1`;
    directly before case 148's `expect`, add
    `expect "161 commit and push in one command, both redirected, no trailer" deny` over `cd $CWT && git commit -qm 'docs(design): add the handoff' 2>&1 && git push -q >/dev/null 2>&1` and
    `expect "162 a planning commit pathspec-limited to spectre/changes/, redirected" allow` over `git -C $CWT commit -q -m 'docs(spectre): record' -- spectre/changes/chg 2>&1 && git -C $CWT push -q origin spectre/chg3`,
    each in the file's `"$(run "$ROOT" Bash "{\"command\":$(q "…")}")"` shape; add the three to the header's case list after 159.
  - [ ] **Step 2: Reproduce before touching anything; do not guess.** `bash scripts/test-protect-main-checkout.sh` fails exactly cases 160, 161 and 162 (`3 failure(s)`).
    <!-- measured: bash <scratch copy of the harness with these three cases>/scripts/test-protect-main-checkout.sh @ 7d8f01ae → FAIL 160 (allow), FAIL 161 (allow), FAIL 162 (deny), "3 failure(s)" -->
  - [ ] **Step 3: Implement** per decision `git-args-stop-at-redirects` in `hooks/protect-main-checkout.py`: add `def git_args(tokens, start)` just above `def bash_hits`, docstring "A git verb's own arguments: the words from `start` up to the first separator, `)` or redirect — `2>&1` and `>/dev/null` are the shell's, never a remote, refspec or pathspec."; it collects words, breaking on `t in SEPARATORS or t == ")"`, and on `t.startswith(("<", ">", "&>"))` popping a trailing all-digit word first. Replace the `rargs` loop (checkout/restore), the `cargs` loop (commit) and the push `args` loop with `git_args(tokens, j + 1)`.
  - [ ] **Step 4: Verify.** `bash scripts/test-protect-main-checkout.sh` prints `all protect-main-checkout cases pass`; `python3 -c "import importlib.util as u; s=u.spec_from_file_location('h','hooks/protect-main-checkout.py'); h=u.module_from_spec(s); s.loader.exec_module(h); print(h.git_args(['commit','-m','','2','>&','1'], 1))"` prints `['-m', '']`; `scripts/check-python-suppressions.sh` exits 0. Commit.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`
**Tests:** `160 a bare change-branch push with its output redirected and piped`, `161 commit and push in one command, both redirected, no trailer`, `162 a planning commit pathspec-limited to spectre/changes/, redirected`
**Regression:** `160 a bare change-branch push with its output redirected and piped` fails — a redirected push of an untasked commit reaches origin unchecked again.
**Commit:** fix(hooks): stop the push and commit argument walks at a redirect
**After:** none

**Decision:** git-args-stop-at-redirects

- [ ] 3. Compare each recorded base with the branch it tracks at the first verify
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `ls scripts/check-base-behind.sh` fails and `grep -rn 'rev-list --count origin/<base>' skills scripts` names only `skills/flow-contracts/operator-prompts-auto-resolution.md`.
  - [ ] **Step 2: Write the test first** — `TestCheckBaseBehind` in `stats/internal/guard/check_base_behind_test.go`, building its repositories with `fxGit` (`check_base_moved_test.go`): a bare origin seeded with one commit on `main` and a `feature/x` branch at it, cloned so `origin/HEAD` is `main`, a branch `spectre/c` checked out at `origin/feature/x`. Run the guard in-process (`runGuard("check-base-behind", …)`) and assert: (a) no argument → rc 2, empty stdout; (b) a temp directory that is no repository → rc 2, empty stdout; (c) no `branch.spectre/c.flowBase` → rc 0, stdout `BASE-CURRENT: <wt> — no recorded base\n`; (d) `flowBase` `main` → rc 0, `BASE-CURRENT: <wt> — main is the tracked branch\n`; (e) `flowBase` `feature/x` with origin's `main` one commit ahead → rc 1, first line `BASE-BEHIND: <wt> — origin/feature/x lacks 1 commits of origin/main; sync the base, never merge main into the change branch (kan-1007):`, then the two command lines of decision `check-base-behind-guard`, the second ending in the sha `git merge-base HEAD origin/feature/x` prints; (f) after running (e)'s first command line with `bash -c` in the fixture → rc 0, `BASE-CURRENT: <wt> — origin/feature/x holds every commit of origin/main\n`. Run `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckBaseBehind$' -count=1` — it fails (no such guard).
  - [ ] **Step 3: Implement** per decision `check-base-behind-guard`: `stats/internal/guard/basebehind.go` registering `Registry["check-base-behind"]`, with `recordedBase` (`resolvebasebranch.go`) for the base and `git symbolic-ref --short refs/remotes/origin/HEAD` (prefix `origin/` stripped) for the tracked branch; `scripts/check-base-behind.sh` a `flow_guard_exec check-base-behind 2 "check-base-behind:" "$@"` shim shaped as `scripts/check-sidecar-reach.sh`, its header the contract (usage, every line, exit codes, the fetch whose failure is ignored, kan-1007); a relative symlink `skills/flow/scripts/check-base-behind.sh -> ../../../scripts/check-base-behind.sh`.
  - [ ] **Step 4: Wire the call site and the contract.** In `skills/flow/verify-and-handoff.md`, after the paragraph `**Load `skills/flow-contracts/project-configuration-isolation.md`** only when the script cannot be located.`, add:

````markdown verified:authored in-tree for this change — the exact prose this step adds
**Then compare every recorded base with the branch it tracks, on the run's first `flow.verify` only.** Run

```bash
check-base-behind.sh <worktree>...
```

over the same resolved set; its header is canonical for its lines. Exit 0 → continue. Exit 1 → relay every `BASE-BEHIND:` block and run its two command lines in order — the first pushes the base and is confirmed per **What still stops** (`skills/flow-contracts/operator-prompts-auto-resolution.md`) — then run **Verify** again from its `begin` mark. Exit 2 → name its stderr in the handoff and continue (kan-1007: a stale `--base` was papered over by merging develop into the change branches, and integrate's reshape folded KAN-1010 into the squash).
````

    In `skills/flow-contracts/operator-prompts-auto-resolution.md`, after `cost 61 failures and about an hour).` and before ` Pushing the synced base is still confirmed,`, insert one space and then this sentence:
  <!-- measured: grep -n 'cost 61 failures and about an hour)' skills/flow-contracts/operator-prompts-auto-resolution.md @ 7d8f01ae → line 49 -->

```text verified:authored in-tree for this change — the exact prose this step adds
`check-base-behind.sh` makes that count for every worktree of the change and prints the sync commands (**Verify**, `skills/flow/verify-and-handoff.md`).
```

    After `  per **What still stops**.` add this line, two-space indented:

```text verified:authored in-tree for this change — the exact prose this step adds
**Never merge the tracked branch into a change branch** (kan-1007): the change branch then carries the tracked branch's commits as its own, and integrate's reshape folds them into the change's squash — the base is what gets synced.
```
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckBaseBehind$|TestCheckGuardSymlinks' -count=1` passes; `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-verbatim-moves.sh` exit 0; `ls stats/internal/guard/check_base_behind_test.go` succeeds, the companion `scripts/run-guard-tests.sh` requires. Commit.

**Build:** green
**Files:** `stats/internal/guard/basebehind.go`, `stats/internal/guard/check_base_behind_test.go`, `scripts/check-base-behind.sh`, `skills/flow/scripts/check-base-behind.sh`, `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/operator-prompts-auto-resolution.md`
**Tests:** `TestCheckBaseBehind`
**Regression:** `TestCheckBaseBehind` fails — a base behind its tracked branch is no longer named, and the run merges develop into the change branch again.
**Commit:** feat(guard): compare each recorded base with the branch it tracks at the first verify
**After:** Task 1

**Decision:** check-base-behind-guard

- [ ] 4. Refuse a panel round over an unticked task, and repair dispatch 7043
  - [ ] **Step 1: Write the test first** — `TestRecordDispatchBeginRefusesAPanelRoundOverAnUntickedTask` in `stats/cmd/flow/record_test.go`, shaped as `TestRecordDispatchBeginRefusesAnInlineRowOverEarlierWork`: `recordRepo(t, "kan-258")`, `spectre/changes/kan-258/tasks.md` written `- [x] 1. a\n- [ ] 8. b\n`, a `genuineDaemon` recording whether it was contacted. Assert: `-role reviewer -key panel-4-primary -slot primary` → exit 2, stderr contains `8` and `close-task.sh`, daemon not contacted; the same with `-started-at 2026-10-22T12:18:45Z` → exit 0; `-role reviewer -key self-review-demo-review-1` → exit 0; with `tasks.md` rewritten `- [x] 1. a\n- [x] 8. b\n` → `panel-4-primary` exit 0; with `tasks.md` removed → exit 0. Run `cd stats && make -s web-dist && go test ./cmd/flow -run 'TestRecordDispatchBeginRefusesAPanelRoundOverAnUntickedTask$' -count=1` — it fails at the first assertion (exit 0).
  - [ ] **Step 2: Implement** per decision `panel-refuses-unticked-tasks`: in `stats/cmd/flow/tick.go`, `func untickedTasks(plan []byte) []string` beside `countTasks`, returning the ids of `- [ ] <n>. ` task lines outside fences, with `countTasks`' fence tracking. In `stats/cmd/flow/record.go`'s `runRecordDispatchBegin`, after the slot check, when `records.PanelDispatchRound` accepts the key for role reviewer and `*startedAt == ""`, read `planPath(f.dir, f.change)` (a read error skips the check) and, when `untickedTasks` is non-empty, print `flow: -key %q opens a review-panel round while tasks.md holds unticked task(s) %s — close every task with close-task.sh before the panel reads it (skills/flow/verify-fix-loop.md)\n` (ids comma-joined) and return 2.
  - [ ] **Step 3: Contract line.** In `skills/flow/verify-fix-loop.md` step 2, after `the decision's implementer pair, `close-task.sh` with its guard and its review gate.` (wrapped across two lines), insert one space and then:

```text verified:authored in-tree for this change — the exact prose this step adds
The task closes through `close-task.sh` — an inline-committed one included — before step 3 opens any `panel-<round>-*` dispatch; `flow record dispatch begin` refuses a panel key while `tasks.md` holds an unticked task (kan-1007: panel round 4 began 16 seconds after task 8's inline commit, before its close).
```
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && make -s web-dist && go test ./cmd/flow -run 'TestRecordDispatchBegin|TestTasks' -count=1` passes; `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-verbatim-moves.sh` exit 0. Commit.
  - [ ] **Step 5: Parent step — the fixer does not run this.** After the commit, the parent repairs row 7043 (seq 36) per decision `retime-then-restamp-7043`, in this order, from `<wt>/stats`:

```bash unverified:run by the parent after the commit — design.md's Live check states what success and failure read as
make -s web-dist
FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd retime-dispatch 7043 2026-10-22T11:59:47Z 2026-10-22T12:23:29Z bf6a0acaeae61a94a0545e02fd8a52660e3df2e8
FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd restamp-dispatch 7043 /Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/587a769e-8249-48a8-9769-a13c7ccce16e.jsonl
```
    <!-- unverified: run by the parent after the commit; design.md's Live check states what success and failure read as -->

    then runs `design.md`'s **Live check** and reports its figures.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/tick.go`, `stats/cmd/flow/record_test.go`, `skills/flow/verify-fix-loop.md`
**Tests:** `TestRecordDispatchBeginRefusesAPanelRoundOverAnUntickedTask`
**Regression:** `TestRecordDispatchBeginRefusesAPanelRoundOverAnUntickedTask` fails — a panel round opens over a task `close-task.sh` has not closed.
**Commit:** fix(flow): refuse a panel round while tasks.md holds an unticked task
**After:** none

**Decision:** panel-refuses-unticked-tasks
**Decision:** retime-then-restamp-7043

- [ ] 5. Name a 1dp leftover as a leftover
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c '1dp' skills/flow/verify-fix-loop.md` prints 0.
  - [ ] **Step 2: Implement** per decision `leftover-1dp-not-clean`: in `skills/flow/verify-fix-loop.md`, after `the previous fix had judged by eye, 45–70 minutes a round).`, insert one space and then:
  <!-- measured: grep -n '45–70 minutes a round)' skills/flow/verify-fix-loop.md @ 7d8f01ae → line 81 -->

```text verified:authored in-tree for this change — the exact prose this step adds
A leftover of any size — a 1dp offset included — is closed or named with its cause, never reported clean because it is small (kan-1007: a fix round called a first-try 1dp offset clean).
```
  - [ ] **Step 3: Verify.** `scripts/check-markdown-integrity.sh` and `scripts/check-verbatim-moves.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** none — a contract sentence; `check-verbatim-moves.sh` pins its text.
**Regression:** none — prose only.
**Commit:** docs(flow): name a 1dp leftover as a leftover in the fix-round measurement
**After:** Task 4

**Decision:** leftover-1dp-not-clean

- [ ] 6. Measure a plan-time capture against its mockup before it commits
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'compose-mockup-frames' skills/flow/brainstorm-planner.md` prints 0.
  - [ ] **Step 2: Implement** per decision `plan-time-capture-measured`: in `skills/flow/brainstorm-planner.md`, after the blockquote opening `> **Write a capture for every frame on the change's frame list.**` (ending `a citation `check-mockup-citations.sh` can resolve.`) and its following blank line, add this blockquote and one blank line after it:

```markdown verified:authored in-tree for this change — the exact prose this step adds
> **Write a capture task that baselines against a mockup to measure before it commits.** Such a
> task's steps, after its capture, run `compose-mockup-frames.sh` on its spec's `<spec>.mockups`
> sidecar and the band, seam and matrix reads of `measure-visual-properties.sh` on every pair, and
> fix until they come back clean or each leftover is named with its cause — the rule **The
> implementer measures before it hands back** (`skills/flow/verify-fix-loop.md`) states for a fix
> round, applied at plan time (kan-1007: a plan-time capture was judged by eye).
```
  - [ ] **Step 3: Verify.** `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-verbatim-moves.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** none — a planning rule; `check-verbatim-moves.sh` pins its text.
**Regression:** none — prose only.
**Commit:** docs(flow): measure a plan-time capture against its mockup before it commits
**After:** none

**Decision:** plan-time-capture-measured

- [ ] 7. Check that each sidecar line's mockup frame exists
  - [ ] **Step 1: Write the cases first** in `stats/internal/guard/check_sidecar_reach_test.go`: a table field `mockups map[string]string` (mockups-root-relative path → content; nil runs the guard with the capture root alone, as today) and three cases — `a frame present in the mockups root passes` (`fidelity/a.spec.ts` `'a-1.png'`, sidecar `a-1.png A-1\n`, mockups `A-1.png`) → rc 0, `SIDECAR-REACH-OK: 1 lines in 1 sidecars\n`; `a frame absent from the mockups root fails` (the same, mockups `B-1.png` only) → rc 1, stdout `fidelity/a.spec.ts.mockups:1: a-1.png — frame A-1 has no ` followed by the case's mockups root joined to `/A-1.png` and a newline; `a one-field line is not judged on its frame` (sidecar `a-1.png\n`, mockups empty) → rc 0. In the exit-2 table, rename `two arguments` to `three arguments` with three temp dirs, and add `mockups root not a directory` (`{t.TempDir(), file}`). Run `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckSidecarReach$' -count=1` — the absent-frame and three-argument cases fail.
  <!-- measured: grep -n 'SIDECAR-REACH-OK' stats/internal/guard/check_sidecar_reach_test.go @ 7d8f01ae → line 20, the one-sidecar OK line the new passing case reuses -->
  - [ ] **Step 2: Implement** per decision `sidecar-frame-exists` in `stats/internal/guard/sidecarreach.go`: accept one or two arguments (usage `check-sidecar-reach.sh <capture root> [<mockups root>]`), the second a directory or exit 2; for a reached line with a second field and a mockups root, `os.Stat(filepath.Join(mockups, fields[1]+".png"))` failing appends `<sidecar>:<line>: <name> — frame <id> has no <mockups root>/<id>.png`, the root printed as given. Update `scripts/check-sidecar-reach.sh`'s header: usage, the frame rule, the new exit-1 line, and that the second argument that is not a directory is exit 2.
  - [ ] **Step 3: Call site.** In `skills/flow/visual-verify.md`, replace the paragraph opening `   When `mockups` is declared, the same Bash call also runs `check-sidecar-reach.sh <capture` (ending `carrying its stderr.`) with:

```markdown verified:authored in-tree for this change — the exact prose this step adds
   When `mockups` is declared, the same Bash call also runs `check-sidecar-reach.sh <capture
   root> <mockups root>`, `<capture root>` the `regression checkout`'s worktree when one is
   declared, else this worktree, and `<mockups root>` the resolved `mockups` root
   (`<worktree>/<mockups>`); its header is canonical for the rule. Exit 1 → each line names a
   sidecar line whose screenshot no spec beside it takes, or whose frame the mockups root does
   not hold, a failing check: kan-998's stale lines surfaced only at compose, after a 13-minute
   capture. Exit 2 → it cannot answer; treat it as a failing check, carrying its stderr.
```
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckSidecarReach$' -count=1` passes; `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-verbatim-moves.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/sidecarreach.go`, `stats/internal/guard/check_sidecar_reach_test.go`, `scripts/check-sidecar-reach.sh`, `skills/flow/visual-verify.md`
**Tests:** `a frame absent from the mockups root fails`
**Regression:** `a frame absent from the mockups root fails` fails — a sidecar naming a frame the mockups root lacks passes pre-flight and breaks compose after the capture again.
**Commit:** feat(guard): check that each sidecar line's mockup frame exists
**After:** none

**Decision:** sidecar-frame-exists

- [ ] 8. Compose a mockups-only round instead of verifying it
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'mockups root' skills/flow/verify-fix-loop.md` prints 0.
  - [ ] **Step 2: Implement** per decision `mockups-only-round-composes`: in `skills/flow/verify-fix-loop.md` step 4, after `this run has verified it, or the parent no longer holds the sha — runs the stage unscoped.`, insert one space and then, as a continuation of the same item:

```text verified:authored in-tree for this change — the exact prose this step adds
**A round whose diff touches only the mockups root composes instead of verifying.** When `git diff <round base>..HEAD` matches no `ui paths` entry in any worktree and touches nothing outside the mockups root, no `visual-verify` verifier is dispatched: for each `<spec>.mockups` sidecar with a line naming a changed `<frame id>.png`, the parent runs the compose command and the `measure-visual-properties.sh … --props bands,seams` read of step 10 (`skills/flow/visual-verify-verifier.md`) on every pair; a leftover is a defect that opens the next round, and clean composites close the round with no `-full` run, the last verifier's report standing for every path outside the mockups root (kan-1007: a mockups-only round re-ran the whole suite, 18m28s).
```
  - [ ] **Step 3: Verify.** `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-verbatim-moves.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** none — a loop rule; `check-verbatim-moves.sh` pins its text.
**Regression:** none — prose only.
**Commit:** perf(flow): compose a mockups-only round instead of verifying it
**After:** Task 4, 5

**Decision:** mockups-only-round-composes
