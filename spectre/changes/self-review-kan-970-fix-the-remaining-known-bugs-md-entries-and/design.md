## Context

Found by the self-review pass of `/flow` change `kan-970-fix-the-remaining-known-bugs-md-entries-and`
(report: the gymie worktree's `.superpowers/sdd/self-review-pass.md`). The fixes ride the change's
own branch `spectre/kan-970-fix-the-remaining-known-bugs-md-entries-and`: in this agents worktree,
based at `4a9b5b8e`, and in the gymie worktree
`/Users/tweety53/Projects/gymie-worktrees/kan-970-fix-the-remaining-known-bugs-md-entries-and`,
based at `6070e574`, for Task 7 and Task 3's record half. Each finding is carried by exactly one
task, Task `<n>` carrying finding F`<n>`; F9's tool half was disproved by the pass (the store and
CLI refusals are by design) and only its doc example is planned.

Every fix was applied in its worktree at plan time, its tests run green, its new tests shown
failing without the fix, and then reverted; each task's diff is committed beside this plan under
`patches/`. Facts measured at plan time:

- **F1.** Task 1's test, applied alone at `4a9b5b8e`, fails: the guard answers
  VISUAL-VERIFY-MISSING naming `spectre/changes/archive/demo/link.md` as a path changed since the
  verified HEAD. The live kan-970 satellite now answers OK only because the forced scoped
  re-verify (ledger dispatch 6) recorded a row verifying the archive commit `f747f837`.
- **F2.** `flow record passes` and `flow record dispatches` for kan-970 now hold one pass row and
  implementer and reviewer rows that the pass found absent: they were recorded after the pass read
  the store, after the gate had cleared the change. The guard gap is unchanged: signal five only
  fires when a `panel-<round>-…` reviewer row exists. The pass's appended Tasks 15–17 with no
  `**Tasks appended:**` marker escaped signal four for the same run; signal eight catches that run
  regardless, since no pass row existed. `skills/flow/unfinished-work-gate.md` defers its signal
  list to the script ("The signals that produce that list are the script's own"), so the third
  file is `scripts/check-unfinished-work.sh`'s header, not that doc. `scripts/test-integrate-change.sh`'s
  stub `flow` answered no `record passes`, so its happy path stopped at exit 2 under the new
  signal until the stub gained a pass row.
- **F3.** At plan time `scripts/check-visual-verify-dispatched.sh` over the gymie-frontend
  worktree with the gymie worktree as canonical answered OK, exit 0, while
  `spectre/changes/archive/kan-970-fix-the-remaining-known-bugs-md-entries-and/` in gymie holds no
  `visual-verification.md` — the guard passes on the dispatch row alone. The misplaced record's
  cited files still exist in this session's scratchpad (`…/scratchpad/vv/`); Task 3's record half
  copies them before the session ends. The visual-verify fixture of
  `scripts/test-integrate-change.sh` gains a committed record so its verified-HEAD case stays OK.
- **F4.** Reproduced against `hooks/protect-main-checkout.py` from this agents checkout's cwd:
  an unquoted `<<EOF` body ``see `a` > b`` written into the gymie worktree is denied; `<<'EOF'`
  passes. Planning also found, against the unmodified hook, that `echo` with a backtick span
  removing a file in the main checkout, `echo "$(rm <main>/x)"`, and an unquoted data body holding
  ``see `rm <main>/x` `` are all allowed: `tokenize` never splits a backtick span, and shlex keeps
  a quoted `$(…)` inside one word. Task 4 fixes all three; the harness's four new cases fail on the
  unmodified hook and pass with the patch.
- **F5.** Both new tests fail without the fix: `check-unfinished-work` accepts `"<wtA>=<sha> <wtB>=<sha>"`
  as one range, `gather-dispatch-context` refuses `5 6` naming only the list's shape.
- **F6.** `scripts/lib/plan_grammar.py`'s `AFTER_VALUE_RE` refuses `Task 5, Task 6` and admits
  `Task 5, 6`, checked by running the regex on both values.
- **F7.** In the gymie worktree the running stack is isolated (frontend on `4130`, gateway on
  `9210`): `./gradlew devStatus` at `6070e574` reports "this workspace's recorded block, not its
  deterministic one" — without `-PisolateOAuthPorts` the deterministic frontend and gateway ports
  are the declared, unmoved ones, which a restart would bind. With Task 7's patch the same command
  prints the "Keeping -PisolateOAuthPorts" line and the deterministic block. `.flow/project.md`'s
  `## run` comment "devRestart drops -PisolateOAuthPorts" becomes false with the fix and is cut.
- **F8, F6, F9 prose.** `scripts/check-verbatim-moves.sh` fails each reworded or new run-loaded
  sentence; this change's `verbatim-moves.txt`, committed with this plan, acknowledges exactly the
  sentences Tasks 6, 8 and 9 write, and the guard exits 0 with all three patches applied.

## Decisions

### Drop the change's own records before the visual trigger

**ID:** vvd-drop-change-records
**Status:** active
**Chosen:** `check-visual-verify-dispatched` drops every path under the worktree's `<spec root>/changes/` from the diff before the trigger question and before the freshness intersection, in every worktree — the same exclusion `unfinishedwork.go`'s range walk already makes.
**Considered:** dropping only in a satellite — an own worktree's change records are no UI work either, and one rule is simpler than two; matching only `link.md` — any record a satellite archives (a `ledger.md`, a narrative) trips the same way.

### The verifier's record is required in the change root

**ID:** vvd-record-required
**Status:** active
**Chosen:** a found verifier row counts only when `<spec root>/changes/<name>/visual-verification.md` or its `archive/` twin exists in the worktree the store is read under — the canonical worktree when one is passed — which is the change root `visual-verify-verifier.md` step 11 writes; absent, the verdict is MISSING naming the path. A Decide skip still answers OK with no record.
**Considered:** also checking every evidence file the record cites — a parser of the record's prose for one recorded miss; checking each satellite's own change root — a satellite holds only `link.md`, the verifier writes the canonical root.

### kan-970's record moves with its evidence into the change root

**ID:** f3-evidence-in-change-root
**Status:** active
**Chosen:** the record moves from the gymie worktree's `.superpowers/sdd/` into `spectre/changes/archive/kan-970-fix-the-remaining-known-bugs-md-entries-and/visual-verification.md`, and every scratchpad file it cites is copied into that root's `visual-verification/evidence/`, the location `visual-verify-verifier.md` step 11 names, the citations rewritten to the copies.
**Considered:** the change's state directory, as the planner's dispatch worded it — the verifier contract names the change root, and the self-review bundle reads the repository, not the state directory.

### F3's record half commits in gymie under its own subject

**ID:** f3-record-half-own-commit
**Status:** active
**Chosen:** Task 3 carries both halves of F3, so each finding stays one task: the guard commit in this worktree under the task's `**Commit:**` subject, and one gymie commit `chore(spectre): move kan-970's visual record into its change root` for the record half — a commit's scope names the module it moved, and a spectre record move is not `guard`.
**Considered:** a separate task for the record half — two tasks for one finding; one subject for both commits — names the wrong module in one of the two repositories.

### A code commit with no pass row is unfinished

**ID:** uw-no-pass-signal
**Status:** active
**Chosen:** `check-unfinished-work` signal eight: where signal seven asks — a non-archived plan and at least one range — a commit outside `<spec root>/changes/` with no `panel-<round>-…` reviewer row and no pass row at all is OUTSTANDING; a change with a panel reviewer row stays signal five's. No exemption for a panel skipped at Decide: no recorded decision skips the panel — `/flow-fast` runs the review-panel step on every class (`skills/flow-fast/SKILL.md`) and keeps no spectre plan this gate reads.
**Considered:** honouring a recorded panel skip as the pass proposed — there is no such decision shape to honour; requiring an implementer row too — plan-time tasks run grouped and inline with no row of their own (signal four's comment), so a missing implementer row is no evidence.

### One hint for an unsplit list

**ID:** unsplit-list-hint
**Status:** active
**Chosen:** one constant `unsplitListHint` in `unfinishedwork.go`, used by both guards: `check-unfinished-work` refuses a range whose merge-base holds whitespace, and `gather-dispatch-context` appends the hint to its existing task-id refusal when the list holds whitespace.
**Considered:** refusing whitespace in the worktree half of a range too — a worktree path may hold a space legitimately; a shared helper function — one string, two callers.

### The hook scans what bash runs, and only that, in a data heredoc

**ID:** hook-subst-spans
**Status:** active
**Chosen:** an unquoted `cat`/`tee` data heredoc whose body holds `$(` or a backtick keeps only its substitution spans scanned, each backtick span rewritten as `$(…)`, or the whole body when a span cannot be closed; and `tokenize` appends, after a line's words, the tokens of every backtick span in the line and every `$(…)` inside a quoted word, each between `;` separators.
**Considered:** leaving the backtick and quoted-substitution gaps, found at plan time, for another change — a noticed defect is fixed, and it is a few lines beside F4's; treating a backtick as a word separator — it breaks case 93's unreadable-landing detection of a backtick refspec. Accepted cost: a single-quoted literal holding a substitution that writes into the main checkout is now denied — over-matching stays on the deny side.

### The After separator is documented, the grammar unchanged

**ID:** after-separator-doc
**Status:** active
**Chosen:** `brainstorm-planner.md`'s `**After:**` sentence names the shape `Task <id>[, <id>…]` with one `Task` prefix, `Task 5, 6`, never `Task 5, Task 6`.
**Considered:** widening `AFTER_VALUE_RE` to admit a repeated prefix — a second accepted spelling of one value, where the strict grammar already refuses clearly at dispatch time.

### The history check lives in section B

**ID:** history-check-in-section-b
**Status:** active
**Chosen:** a bullet in `brainstorm-planner.md` section B's checklist, which every planner runs — `/flow`'s brainstorm, a `/flow-plan` capture, the self-review planner and the in-run fix planner.
**Considered:** `skills/flow/brainstorm.md`, as the pass proposed — only `/flow`'s own brainstorm reads it, and the self-review and in-run fix planners plan defect fixes too.

### The inline verify example separates begin from end

**ID:** inline-verify-example-gap
**Status:** active
**Chosen:** one comment line between the example's `begin` and `end` commands, naming the list run and the `## Report` and that the two are never one call.
**Considered:** splitting the example into two blocks — the existing prose already orders them; the comment is where the reader copying the block looks.

### A restart keeps a running isolated stack isolated

**ID:** oauth-isolation-inherits
**Status:** active
**Chosen:** `isolateOAuthPorts()` is true with the flag, or — decided once per invocation, before any stop, and memoised in the project's extra properties — when a running frontend or gateway's recorded port is not its declared one (`inheritsOAuthIsolation` in buildSrc's `FrontendPortContract.kt`, where it is tested); the inheritance prints one lifecycle line naming the flag and `devStop` as the way to give the ports back. Every caller — the block resolution, `resolveBlock`'s flag and the `ALLOWED_ORIGINS` export — reads the same answer.
**Considered:** refusing a flagless restart of an isolated stack — every restart then needs the flag retyped, the muscle-memory failure this finding is; fixing `devRestart` alone — a bare `devStart` restarts the same way; deciding per call — a restart's stop loop takes down the services the answer is read from, so its starts would export the unisolated origin.

### Each task's diff is a committed patch

**ID:** patches-beside-the-plan
**Status:** active
**Chosen:** each task's diff, applied and checked in its worktree at plan time, is committed under this change's `patches/` as `task-<n>.diff` (plus `task-3-record.sh` for F3's record half), and the task applies it with `git apply`.
**Considered:** inlining the diffs as fenced blocks in `tasks.md` — the plan becomes a transcript of the code; describing the edits for the fixer to write — throws away code already written and tested at plan time.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — two gate guards change what they let through, and the main-checkout hook's scanning changes in a direction where a slip either blocks every agent or lets a write into the main checkout pass.
**Considered:** `flow-low` — every task is small and test-pinned, but the hook's tokenizer and the gate's new refusal reach every later run.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** three fixer dispatches in series: `fix-g1` Tasks 1, 2, 3, 5 — the Go guards, which share `stats/internal/guard/` and `scripts/test-integrate-change.sh`, plus Task 3's gymie record half; `fix-g2` Tasks 4, 6, 8, 9 — the hook and the prose; `fix-g3` Task 7 — gymie's Gradle build, a different repository and build tool.
**Considered:** one fixer for all nine — one context carrying the Go suite, a Python hook harness and a Gradle build; one fixer per task — nine dispatches for patches already written and tested.

## Open questions

## Live check

Task 7 changes the gymie dev stack's restart. In the gymie worktree, with its isolated stack running
(frontend and gateway on their moved ports, as `./gradlew devStatus` reports), run
`./gradlew devRestart -Pservices=frontend` with no `-PisolateOAuthPorts`: it prints the "Keeping
-PisolateOAuthPorts" line and `devStatus` afterwards reports the frontend on the same moved port.
Not working looks like the frontend restarted on its declared port while `devStatus` still names the
moved gateway. The guard and hook tasks have no runtime of their own to check: they read the store
and decide in-process, and their tests run them against real repositories.
