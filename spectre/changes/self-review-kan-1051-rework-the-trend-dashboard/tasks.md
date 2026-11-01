# self-review-kan-1051-rework-the-trend-dashboard

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** close kan-1051's six non-`big` self-review findings — F7 (Task 1), F4 (Task 2), F5 (Task 3),
F6 (Task 4), F18 (Task 5), F17 (Task 6), as `proposal.md` names them.

**Architecture:** one harvester parser fix plus a store-row repair; four contract edits, one backed by a
guard-table phrase; one new verb in a dev shell tool.

**Tech Stack:** Go (`stats/`), Bash, Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-11-01-self-review-kan-1051-rework-the-trend-dashboard-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never write `spectre/changes/`. Never merge or push.
- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1051-rework-the-trend-dashboard` (`<wt>`), branch `spectre/kan-1051-rework-the-trend-dashboard`; `**Files:**` paths are relative to `<wt>`. Never touch a main checkout, and never create a worktree.
- Go: `gofmt -w` every Go file you change; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Mutation proofs restore with `scripts/break-and-prove.sh` — never `git checkout --` on a file carrying uncommitted edits.
- Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database, and never run tests against the live database. Task 1's one store write goes through `flowd restamp-dispatch`, which leaves the daemon running.
- Every Go test run sets a Bash `timeout` of 600000.

## Review Focus

- Task 1: `&` and newline end a body only when `done` follows; `&&` inside a body, and a quoted `& done`, never end it; row 9234 now carries zero tokens and no other row moved.
- Task 3: the guard phrase and the paragraph sentence match; the body rule and the slot paragraph say the same thing.
- Task 5: the separate verify is gone only where `pipeline-fix.sh land` runs; the change's own branch keeps it.
- Task 6: `restart` kills only the pid its own pidfile names for its own binary; it never copies, builds or drops.

- [ ] 1. End a harvested loop body at `&` or a newline before `done`, and repair row 9234
  - [ ] **Step 1: Confirm the trace; do not guess.** The trace is the brainstorm design file's F7 section. Run `grep -c ') & done' ~/.claude/projects/-Users-tweety53-Projects-agents/82eaa458-f303-4fbb-bbf5-ed58a6d64147.jsonl` — at least 1 (verify-5's `for g in lint test; do ( … verify-5 mf-k1051-20261031T200548Z … ) & done`). Run `flow record dispatches -change kan-1051-rework-the-trend-dashboard` and confirm row 9234 (key `verify-5`) carries `"metrics": {}` and row 9180 (`verify-2`) carries `tokens.main` all 0. Any mismatch: stop and report it.
    <!-- measured: grep -o 'for g in lint test; do[^"]*verify-5[^"]*done' on the 82eaa458 transcript — 2 hits ending ') & done'; flow record dispatches row 9234 metrics {} and row 9180 tokens.main all 0 @ 1f07f9ce -->
  - [ ] **Step 2: Write the test.** `TestJudgeInvocationsEndsLoopBodyAtAnyTerminator` in `stats/internal/harvest/watcher_test.go`, table-driven over `judgeInvocations(<command>, runVerifyListBeginFromInvocation)`: (a) `for g in lint test; do ( set -a; $S/run-verify-list.sh $W $g $N verify-5 mf-x opus default </dev/null > l 2>&1 ) & done` yields two events, key `verify-5`, token `mf-x`; (b) the same with `; done` yields two (unchanged behaviour); (c) the same body with `\n` before `done` yields two; (d) a body containing `a && b` before `; done` still yields two and is not cut at the `&&`; (e) `echo '& done'` inside the body before `; done` is not a terminator (quoted). Run `cd stats && go test ./internal/harvest -run 'TestJudgeInvocationsEndsLoopBodyAtAnyTerminator' -count=1` — (a) and (c) fail.
  - [ ] **Step 3: Implement.** In `scanLoopBody` (`stats/internal/harvest/watcher.go`), per decision `loop-body-terminators`: treat `;`, `\n`, and a lone `&` (not part of `&&`, i.e. neither neighbour is `&`) as a terminator; when `done` follows it after `skipSpaces`, return `s[i:j]` and the index after `done`. Update the function's comment to name the three terminators (kan-1051 self-review, F7).
  - [ ] **Step 4: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, put the check back to `s[j] == ';'` alone — case (a) fails; restore.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l internal/harvest` prints nothing; `cd stats && go test ./internal/harvest -run 'TestJudgeInvocations|TestWatcherStampsKan925LoopBegins|TestWatcherStampsLoopLaunchByItsKey' -count=1` passes. Commit.
  - [ ] **Step 6: Repair row 9234** (decision `repair-row-9234`), after the commit. Run `go -C stats build -o "$TMPDIR/flowd-sr1051" ./cmd/flowd`, then `"$TMPDIR/flowd-sr1051" restamp-dispatch 9234 ~/.claude/projects/-Users-tweety53-Projects-agents/82eaa458-f303-4fbb-bbf5-ed58a6d64147.jsonl`. Then `flow record dispatches -change kan-1051-rework-the-trend-dashboard` shows row 9234 with `tokens.main` all 0, and rows 9180, 9192 and 9206 unchanged from Step 1. Report the before and after `metrics` of row 9234.
    <!-- unverified:restamp-dispatch on an agentless row writes the zero ClosedInlinePatch (tokens all 0, not {}) — read from stats/cmd/flowd/main.go restampDispatch and harvest.ClosedInlinePatch, not run -->

**Build:** green
**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`
**Tests:** `TestJudgeInvocationsEndsLoopBodyAtAnyTerminator`
**Regression:** `TestJudgeInvocationsEndsLoopBodyAtAnyTerminator` fails — a backgrounded `for` body ending `& done` is not expanded, and its verify row stays unmeasured.
**Commit:** fix(harvest): end a loop body at any list terminator before done
**After:** none

**Decision:** loop-body-terminators

- [ ] 2. Name the restoring planning commit as the since-close sha
  - [ ] **Step 1: Edit.** In `skills/flow/review-panel-late-fix.md`, replace the sentence opening "**After integrate's reshape**" through "a peer worktree's reshaped branch head." with the block below (decision `since-close-after-restore`).

```markdown unverified:run-verify-list.sh lint on the applied copy
**After integrate's reshape** (**3. Commit the staged work**, `skills/flow/integrate.md`) the shas
the last clean close reviewed are gone from the branch, so the since-close sha is the canonical
worktree's drop commit and a peer worktree's reshaped branch head — save where a fix run's planning
commit restored the change directory after that drop (**A change already dropped is never asked**,
`skills/flow/document-fix.md`): the canonical worktree's since-close sha is then that planning commit,
since a range from the drop commit counts the whole restored directory (kan-1051 self-review, F4:
12,397 changed lines, where the fix's own delta was small).
```

  - [ ] **Step 2: Verify.** `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel-late-fix.md`
**Tests:** none
**Regression:** none — prose; the next fix run after a restore reads the right since-close sha.
**Commit:** fix(flow): name the restoring planning commit as the since-close sha
**After:** none

**Decision:** since-close-after-restore

- [ ] 3. Keep panel reproducers out of the shared worktree
  - [ ] **Step 1: Make the guard want the rule.** In `stats/internal/guard/dispatchparagraphs.go`, append `"go test -overlay"` to the `reproducer` row's phrase list in `dpEntries`. Run `scripts/check-dispatch-paragraphs.sh` — it exits 1, naming the missing phrase in `skills/flow/review-panel.md`.
  - [ ] **Step 2: Edit.** In `skills/flow/review-panel.md` (decision `reproducer-writes-nothing-in-tree`): (a) in the body rule, right after the sentence ending "never through shell quoting nested inside another quoted string (kan-1047: … never applied the edit).", insert the first block below; (b) in the **REPRODUCER:** blockquote, right after the sentence ending "never in the recorded line.", insert the second block below, as `> `-prefixed lines.

```markdown unverified:run-verify-list.sh lint on the applied copy
A reproducer writes nothing under `$PWD`: a Go test it adds is written under `mktemp -d` and passed
to `go test -overlay <json>`, whose `Replace` map names the test's path in the package, and every
other file it writes lives in its own `mktemp -d` directory (kan-1051 self-review, F5: a test file a
reproducer wrote into `stats/internal/harvest/` failed a concurrent guard-test run).
```

```markdown unverified:run-verify-list.sh lint on the applied copy
It writes nothing under `$PWD`: a Go test it adds is written under `mktemp -d` and passed to
`go test -overlay <json>`, and any other file it writes lives in its own `mktemp -d` directory.
```

  - [ ] **Step 3: Prove the guard pins the rule.** With `scripts/break-and-prove.sh`, delete the second block from the blockquote — `scripts/check-dispatch-paragraphs.sh` exits 1; restore.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckDispatchParagraphs' -count=1` passes; `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.
    <!-- measured: go -C stats test -overlay <json> -run TestZZOverlaySpike ./internal/harvest with the test under mktemp -d — PASS, git status --porcelain stats/ empty @ 1f07f9ce -->

**Build:** green
**Files:** `skills/flow/review-panel.md`, `stats/internal/guard/dispatchparagraphs.go`
**Tests:** none
**Regression:** `scripts/check-dispatch-paragraphs.sh` fails when the REPRODUCER paragraph loses its overlay rule for a Go reproducer's test file.
**Commit:** fix(flow): keep panel reproducers out of the shared worktree
**After:** none

**Decision:** reproducer-writes-nothing-in-tree

- [ ] 4. Have the self-review pass hand back its files' contents
  - [ ] **Step 1: Edit.** In step 2 of `skills/flow-self-review/SKILL.md` (decision `pass-hands-back-contents`): (a) after "**This session passes file paths, never file contents.**" and its sentence ending "outside every checkout, never committed.", add "The one exception is the pass's own output: the harness refuses a subagent's Write under `<sr-dir>`, so the pass hands back its files' full contents and this session writes them (kan-1051 self-review, F6)."; (b) change "It writes `<sr-dir>/findings.md`:" to "It hands back the full content of `<sr-dir>/findings.md`:" and "It writes `<sr-dir>/report-draft.md`," to "It hands back the full content of `<sr-dir>/report-draft.md`,"; (c) replace the sentence "Every file it writes under `<sr-dir>` or `<state-dir>/flow-eval/<name>/` is its own report file under READ-ONLY REVIEW." with "It hands back every case file under `<sr-dir>/cases/<case-id>/` and every file for `<state-dir>/flow-eval/<name>/` the same way, each under its target path, and writes no file itself."; (d) change "It hands back only its per-angle counts," to "Beside those contents it hands back its per-angle counts,"; (e) after that paragraph's sentence ending "and whether it met a product defect.", add "Before step 2a, this session writes each returned file to its target path, byte for byte."
  - [ ] **Step 2: Verify.** `git -C <wt> grep -n -F -e 'It writes `<sr-dir>/findings.md`' -e 'It writes `<sr-dir>/report-draft.md`' -e 'is its own report file under READ-ONLY REVIEW' -- skills/flow-self-review/SKILL.md` prints nothing; `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`
**Tests:** none
**Regression:** none — prose; the next pass hands back its files instead of failing a Write.
**Commit:** fix(flow-self-review): have the pass hand back its files' contents
**After:** none

**Decision:** pass-hands-back-contents

- [ ] 5. Let land be the loop's verify on merge and push
  - [ ] **Step 1: Edit.** Per decision `land-is-the-verify`: (a) in step 3's item 4 **Verify** of `skills/flow-self-review/SKILL.md`, right after the sentence ending "and reads only its exit code and its `red:` lines.", add the first block below; (b) in `skills/flow-contracts/pipeline.md`'s **Pipeline defects found mid-run** item 4 **Land**, right after "Merge and push is `<agents repo>/scripts/pipeline-fix.sh land <worktree>`, run with a Bash `timeout: 600000` — a landing that moved `<agents repo>/stats/` restarts flowd inside the call, and a kill partway leaves it down.", add the second block below.

```markdown unverified:run-verify-list.sh lint on the applied copy
   On a `self-review-<name>` branch whose route is merge and push, that separate verify is not run:
   step 5's `pipeline-fix.sh land <fix worktree> > <sr-dir>/verify-<v>.log 2>&1` runs the same lint
   lines and tests on the rebased HEAD before it pushes, and its exit code and `red:` lines are read
   as this verify's (kan-1051 self-review, F18: the suite ran twice at one HEAD, about 5 min a loop).
```

```markdown unverified:run-verify-list.sh lint on the applied copy
On that route the loop runs no separate `pipeline-fix.sh verify`: land runs it on the rebased HEAD,
and its `red:` lines feed `pipeline-fix-<k>-fix-v<v>` as a verify's do.
```

  - [ ] **Step 2: Verify.** `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.
    <!-- measured: stats/internal/guard/pipelinefix.go pfLand calls pfVerify after its rebase and before the push; pfVerify prints red:/green: lines and writes no store record @ 1f07f9ce -->

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`, `skills/flow-contracts/pipeline.md`
**Tests:** none
**Regression:** none — prose; a merge-and-push loop runs its suite once per landing attempt.
**Commit:** fix(flow-self-review): let land be the loop's verify on merge and push
**After:** Task 4

**Decision:** land-is-the-verify

- [ ] 6. Add a restart verb to live-store-copy.sh
  - [ ] **Step 1: Write the test cases.** In `scripts/test-live-store-copy.sh`: (a) `restart` with no prior `up` exits 2 and runs nothing (empty `$log`); (b) after `up` with `FLOW_LIVE_COPY_TRANSCRIPTS="$stub/real"`, `restart "$name"` exits 0, the old pid is dead, a new pid is in `/tmp/flowd-$port.pid` and alive, `$log` holds no `createdb`, `pg_dump`, `dropdb`, `make` or `go` line from the restart, the relaunch's `TRANSCRIPTS=` line is `$stub/real`, a sentinel process survives, and `$sdd` still exists; (c) after a restart where the stub flowd leaves a stale pidfile behind (the stub's pidfile is written again with a dead pid before restart's launch), restart still exits 0 — the stale file is removed once its pid is dead; this case's `fail` message reads `restart: stale pidfile …`. Then `down` as today. Run `scripts/test-live-store-copy.sh` — the new cases fail.
  - [ ] **Step 2: Implement.** In `scripts/live-store-copy.sh`, per decision `live-store-copy-restart`: move `down`'s pidfile-checked kill into a `stop_flowd` function and `up`'s `env … nohup "$dir/flowd" …` launch plus the 60 s poll into a `launch_flowd` function, both used unchanged by `up`/`down`; `up` writes the transcripts directory it used to `$dir/transcripts-dir`; add `restart <name>` — exit 2 when `$dir/port` or `$dir/flowd` is missing, else `port_ok`, `stop_flowd`, then remove `/tmp/flowd-<port>.pid` if it still exists and the pid on its first line is dead, then `launch_flowd` with the recorded transcripts directory, printing the address and exports as `up` does and exiting 1 when flowd does not answer within 60 s. Update the usage line and the header comment (a `restart` entry; kan-1051 self-review, F17).
  - [ ] **Step 3: Prove the tests pin the fix.** With `scripts/break-and-prove.sh`, drop the stale-pidfile removal — case (c) fails; restore.
  - [ ] **Step 4: Verify.** `bash -n scripts/live-store-copy.sh` exits 0; `scripts/test-live-store-copy.sh` prints `all cases passed`; `scripts/check-references.sh` exits 0. Commit.

**Build:** green
**Files:** `scripts/live-store-copy.sh`, `scripts/test-live-store-copy.sh`
**Tests:** `restart: stale pidfile`
**Regression:** `scripts/test-live-store-copy.sh` fails — `restart` is missing, recopies or rebuilds, or exits on a stale pidfile.
**Commit:** feat(scripts): add a restart verb to live-store-copy.sh
**After:** none

**Decision:** live-store-copy-restart
