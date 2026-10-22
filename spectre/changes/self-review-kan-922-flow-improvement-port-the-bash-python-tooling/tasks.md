# self-review-kan-922-flow-improvement-port-the-bash-python-tooling

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close kan-922's four self-review findings — F-A (Task 1), F-B (Task 2), F-C (Task 3),
F-D (Task 4), as `proposal.md` names them.

**Architecture:** `integrate-change.sh prepare` switches a clean drifted main checkout back to its
default branch; the Go land chain runs the repository's report check before staging; the inline
begin's dirt check reads the change's worktree; the fixer brief names where `break-and-prove.sh`
runs its test command.

**Tech Stack:** Bash, Go (`stats/`), Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-22-self-review-kan-922-flow-improvement-port-the-bash-python-tooling-design.md`
in this worktree. `verbatim-moves.txt` beside this file already lists every run-loaded sentence
Tasks 1, 2 and 4 reword or add, exactly as their steps spell them.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-922-flow-improvement-port-the-bash-python-tooling` (`<wt>` below), on branch `spectre/kan-922-flow-improvement-port-the-bash-python-tooling`; every `**Files:**` path is relative to it. Never touch the main checkout and never create a worktree.
- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Prose a step spells in a quote block is inserted exactly as written, line breaks included, so `check-verbatim-moves.sh` matches `verbatim-moves.txt`; it exits 0 after every task.
- Go: `gofmt -w` every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Never widen a timeout, retry or tolerance to make a test pass (`rules/fix-determinism-at-the-source.mdc`).
- Never stop or restart `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container.

## Review Focus

- Task 1: only the clean shape switches; a detached HEAD, a checkout on the change's recorded base, a refused switch and a dirty checkout stay where they were and stop; one repository's switch happens even when another repository stops the run; the drift branch keeps its commits.
- Task 2: a violation or an exit-2 check refuses before `git add`, HEAD and index untouched; a repository without the script and a non-report add-path land as before.
- Task 3: a dirty main checkout beside a clean change worktree admits the begin; an edit inside the change worktree still refuses it; `-C <worktree>` behaves as before.
- Task 4: the sentence sits inside the existing quote block of the fixer brief.

- [ ] 1. Switch a clean drifted main checkout back to its default branch in prepare
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Write the `drift_switch` case of Step 2 first and run `bash scripts/test-integrate-change.sh 2>&1 | grep -E '^(FAIL|ok) - drift_switch'`: its checks fail — prepare stops at `STOP: foreign-staged` and `a` stays on `side`.
  - [ ] **Step 2: Write the cases** in `scripts/test-integrate-change.sh`, each its own `( trap cleanup_work EXIT … end_case ) >"$LOGS/<nn>.log" 2>&1 &` / `PIDS+=($!)` block after case 42 (`rebase_lint_dirty`), and list each in the header's `Cases:` block after `outstanding`:
    - `drift_switch` (log 43): `fixture`; `g -C "$WORK/a" checkout -q -b side`; write and commit `side.txt` there (`-m "side work"`); `side="$(g -C "$WORK/a" rev-parse side)"`; `sut prepare "$WORK/a" demo`. Checks: `drift_switch: prepare exits 0`; `drift_switch: one DRIFT-SWITCHED line names side, its sha and its one commit` — exactly one stdout line equal to `DRIFT-SWITCHED: $WORK/a — side -> main, side kept at <git rev-parse --short of $side> (1 commits ahead of main)`; `drift_switch: a's main checkout is on main`; `drift_switch: side keeps its commit` (`rev-parse side` equals `$side`); `drift_switch: the re-read prints DRIFT-CLEAN` (a line `DRIFT-CLEAN: $WORK/a — main`); `drift_switch: no foreign-staged stop and the preflight ran`.
    - `drift_switch_stops` (log 44): `fixture`; `a` on a new branch `side`; `b` on a new branch `other` with `base.txt` modified, unstaged; `sut prepare "$WORK/a" demo`. Checks: `drift_switch_stops: a dirty b stops prepare at STOP: foreign-staged` (exit 1, last line `STOP: foreign-staged`); `drift_switch_stops: a is switched though b stops the run` (a on `main`, a `DRIFT-SWITCHED: $WORK/a ` line); `drift_switch_stops: dirty b is not switched` (b on `other`, `base.txt` still modified, no `DRIFT-SWITCHED: $WORK/b ` line). Then `fixture`; `g -C "$WORK/a" checkout -q --detach`; `sut prepare "$WORK/a" demo`. Check: `drift_switch_stops: a detached HEAD stops and stays detached` (exit 1, `STOP: foreign-staged`, `branch --show-current` empty, no `DRIFT-SWITCHED` line).
    - `drift_switch_refused` (log 45): `fixture`; `a` on a new branch `side`; `g -C "$WORK/a" worktree add -q "$WORK/a-main" main`; `sut prepare "$WORK/a" demo`. Check: `drift_switch_refused: a switch git refuses stops and leaves a on side` (exit 1, `STOP: foreign-staged`, the line `DRIFT-SWITCH-REFUSED: $WORK/a — git switch main failed; left on side`, a on `side`).
    - In `refresh_on_recorded_base` (case 39), before its `sut prepare "$WORK/a" demo --accept-foreign`, add `sut prepare "$WORK/a" demo` and the check `refresh_on_recorded_base: prepare never switches a checkout on its recorded base` (exit 1, a still on `feature/x`, no `DRIFT-SWITCHED:` line); amend that case's header entry to say prepare leaves the checkout on its recorded base and still asks.
    <!-- measured: grep -c 'LOGS/42.log' scripts/test-integrate-change.sh @ 15542943 → 1, the last numbered case -->
  - [ ] **Step 3: Implement** per decision `prepare-switch-clean-drift` in `scripts/integrate-change.sh`. Add `switch_back` above `cmd_prepare`:

  ```bash unverified:run Step 4's harness; confirm git switch -q DWIMs a missing local default branch from origin
  # switch_back <main-checkout>: a clean checkout on another branch goes back
  # to its default branch (operator instruction, kan-922's integrate run) —
  # never a detached HEAD, whose commits sit on no branch, and never a branch
  # the change records as its base (KAN-911), which run 2 refreshes. The
  # branch left behind keeps every commit; a refused switch changes nothing
  # and leaves the drift for the question.
  switch_back() {
    local b d k base sha n
    b="$(git -C "$1" branch --show-current 2>/dev/null)" && [ -n "$b" ] || return 0
    d="$(git -C "$1" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null)" || return 0
    d="${d#origin/}"
    for k in "${MAINS[@]}"; do
      base="$(git -C "$k" config --get "branch.spectre/$NAME.flowBase" 2>/dev/null)" && [ "$base" = "$b" ] && return 0
    done
    sha="$(git -C "$1" rev-parse --short HEAD)"
    if git -C "$1" switch -q "$d"; then
      n="$(git -C "$1" rev-list --count "$d..$b")"
      say "DRIFT-SWITCHED: $1 — $b -> $d, $b kept at $sha ($n commits ahead of $d)"
    else
      say "DRIFT-SWITCH-REFUSED: $1 — git switch $d failed; left on $b"
    fi
  }
  ```

    and make `cmd_prepare`'s per-checkout loop:

  ```bash unverified:run Step 4's harness
    for m in "${MAINS[@]}"; do
      run "$SCRIPT_DIR/check-foreign-staged.sh" "$m"
      fs=0; { [ "$RC" -eq 0 ] && [ -n "$(verdict STAGED-CLEAN)" ]; } && fs=1
      run "$SCRIPT_DIR/check-main-checkout-drift.sh" "$m"
      if [ "$fs" -eq 1 ] && [ "$RC" -eq 0 ] && [ -n "$(verdict DRIFT-BRANCH)" ] && [ -z "$(verdict DRIFT-DIRTY)" ]; then
        switch_back "$m"
        run "$SCRIPT_DIR/check-main-checkout-drift.sh" "$m"
      fi
      [ "$fs" -eq 1 ] || ask=1
      { [ "$RC" -eq 0 ] && [ -n "$(verdict DRIFT-CLEAN)" ]; } || ask=1
    done
  ```

    adding `fs` to `cmd_prepare`'s `local` line. In the header's `prepare` paragraph, after "and check-main-checkout-drift per distinct main checkout,", add "switching a checkout whose only drift is DRIFT-BRANCH back to its default branch (switch_back below) and re-reading its drift,". Leave `check-main-checkout-drift` untouched.
  - [ ] **Step 4: Write the contract text** per decision `prepare-switch-clean-drift`.
    In `skills/flow-contracts/finish-contract-run1.md`, **Surface foreign staged work before the preflight**'s drift paragraph, replace its two lines opening "repository the run continues with nothing more said." and "line, every repository's findings" with:

    > repository the run continues with nothing more said. **A clean checkout on another branch is
    > switched back, never asked about.** When a repository's only finding is `DRIFT-BRANCH` — both
    > guards exited 0, `check-foreign-staged.sh` printed `STAGED-CLEAN` and no `DRIFT-DIRTY` printed —
    > `integrate-change.sh prepare` runs `git -C <main-checkout> switch <default>`, prints one
    > `DRIFT-SWITCHED: <main-checkout> — <branch> -> <default>, <branch> kept at <sha> (<n> commits
    > ahead of <default>)` line, and re-runs `check-main-checkout-drift.sh`, continuing only on
    > `DRIFT-CLEAN`. The switch moves no commit: `<branch>` keeps every commit it carried, and a live
    > session that had the checkout on it finds its HEAD moved, which the printed line names. Four
    > shapes are never switched, and stop and ask as below: a detached HEAD, whose commits sit on no
    > branch; a checkout on a base the change records (`branch.spectre/<name>.flowBase`), which run 2's
    > refresh brings forward (KAN-911); a switch git refuses — the default branch checked out in
    > another worktree, an untracked file it would overwrite, no default branch to switch to, a locked
    > index — which leaves the checkout as it was and adds a `DRIFT-SWITCH-REFUSED` line to the
    > listing; and a re-run that does not print `DRIFT-CLEAN`. Each repository is judged on its own: one
    > repository's switch never waits on another's question. On any other `DRIFT-BRANCH` line, or any
    > `DRIFT-DIRTY` line, every repository's findings — this guard's and `check-foreign-staged.sh`'s alike — are

    the next line, opening "shown together and the run stops", and everything after it unchanged.
    In `skills/flow-contracts/finish-hand-fallbacks.md`, append to the **When the drift script is absent** paragraph, after "it is never skipped for want of the script.":

    > A checkout whose only finding is its branch is switched back by hand with
    > `git -C <main-checkout> switch <default>` under the same four exclusions, per **Surface foreign
    > staged work before the preflight** (`skills/flow-contracts/finish-contract-run1.md`).

    In `skills/flow/integrate.md`, append to the `STOP: foreign-staged` bullet's last line (after "re-runs `prepare` with `--accept-foreign`."), on the same line:

    > A `DRIFT-SWITCHED:` line is relayed as printed and asks nothing: `prepare` already switched that clean checkout back to its default branch and carried on, per **Surface foreign staged work before the preflight** (`skills/flow-contracts/finish-contract-run1.md`).

    In `skills/flow-contracts/finish-contract-rationale.md`, add after the `## finish-contract-run1.md — one task, one worktree (2026-10-04)` section a section `## finish-contract-run1.md — a clean drift branch is switched back (2026-10-22)` stating: the operator's instruction during kan-922's integrate run ("resolve it yourself, update flow to resolve such things automatically") after prepare stopped on a clean checkout left on another session's `fix/base-missing-on-satellite`; why only the clean shape switches (nothing tracked can be lost); why a detached HEAD is excluded (its commits sit on no branch); why a recorded base is excluded (KAN-911 — run 2 refreshes a checkout on the change's base onto that base); why a plain switch (the branch keeps its commits, no reset); why the switch lives in `integrate-change.sh`, not the guard or a Go subcommand (decision `prepare-switch-clean-drift`'s Considered line, cited by ID, not copied).
  - [ ] **Step 5: Verify.** `bash scripts/test-integrate-change.sh` exits 0 with every `drift_switch` and `refresh_on_recorded_base` check `ok`; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-guard-symlinks.sh` exit 0. Commit.

**Build:** green
**Files:** `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-contracts/finish-hand-fallbacks.md`, `skills/flow/integrate.md`, `skills/flow-contracts/finish-contract-rationale.md`
**Tests:** `drift_switch: one DRIFT-SWITCHED line names side, its sha and its one commit`, `drift_switch_stops: dirty b is not switched`, `drift_switch_refused: a switch git refuses stops and leaves a on side`, `refresh_on_recorded_base: prepare never switches a checkout on its recorded base`
**Regression:** `drift_switch: one DRIFT-SWITCHED line names side, its sha and its one commit` fails — a clean checkout on another branch stops prepare to ask again; `refresh_on_recorded_base: prepare never switches a checkout on its recorded base` fails — prepare moves a checkout off the base run 2 refreshes.
**Commit:** feat(scripts): switch a clean drifted main checkout back to its default branch in prepare
**After:** none

**Decision:** prepare-switch-clean-drift

- [ ] 2. Refuse a self-review report that fails its check at landing
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Write Step 2's test first and run `cd stats && go test ./internal/guard -run 'TestLandSelfReviewReportRefusesAReportFailingItsCheck$' -count=1`: it fails — the chain lands the report (rc 0) where 5 is wanted.
  - [ ] **Step 2: Write the test** `TestLandSelfReviewReportRefusesAReportFailingItsCheck` at the end of `stats/internal/guard/land_self_review_report_test.go`, built as `TestLandSelfReviewReportRefusesUntracedFinding` is (its repo, `fxGit`, `env`). Write `scripts/check-self-review-report.sh` into the repo, untracked: `if [ -e "$1/UNANSWERABLE" ]; then exit 2; fi; if cat "$1"/*-self-review.md 2>/dev/null | grep -q BAD; then echo "$1/kan-c-self-review.md:3: BAD line"; exit 1; fi; exit 0`. Sub-cases, in order: (a) add-path `docs/self-review/kan-c-context.md` holding `BAD` lands, rc 0 (only `*-self-review.md` is checked); (b) `docs/self-review/kan-c-self-review.md` holding `BAD` — rc 5, stdout carries the stub's line, stderr is exactly `LAND-REPORT-INVALID: docs/self-review/kan-c-self-review.md fails check-self-review-report — fix the lines it names; nothing added, committed, pulled or pushed\n`, HEAD unmoved and nothing staged; (c) with the report fixed and `docs/self-review/UNANSWERABLE` present — rc 5, stderr exactly `LAND-REPORT-UNCHECKED: docs/self-review/kan-c-self-review.md — check-self-review-report could not answer (exit 2); nothing added, committed, pulled or pushed\n`, HEAD unmoved; (d) `UNANSWERABLE` removed — rc 0, the report committed. The repositories of the existing tests carry no such script, so they land as before.
  - [ ] **Step 3: Implement** per decision `land-runs-report-check` in `stats/internal/guard/landselfreviewreport.go`: a `lsrrReportCheck(env, repo, addPath, stdout, stderr) int`, called right after `lsrrUntraced` and resolving paths the way `lsrrUntraced` does; it returns 0 unless the add-path's base name ends `-self-review.md` and `<repo>/scripts/check-self-review-report.sh` is a regular file, then runs `exec.Command("bash", <script>, <the report's directory>)` with `Dir` the repository and stdout/stderr the guard's, and maps exit 1 and any other failure to the two lines of the decision, returning 5. Its comment cites kan-922's self-review (KAN-923's report landed with two violations). Update the file's top comment's exit list. In `scripts/land-self-review-report.sh`'s header: the chain list gains "run the repository's own scripts/check-self-review-report.sh over a report's directory when it carries one" after the open-finding check, and the exit codes gain `5  the report fails its check — one LAND-REPORT-INVALID line after the check's own, or LAND-REPORT-UNCHECKED when the check exits other than 0 or 1; checked before anything is added, so nothing is added, committed, pulled or pushed`.
  - [ ] **Step 4: Write the contract text** per decision `land-runs-report-check`. In `skills/flow-self-review/SKILL.md` step 7, replace the two sentences from "Run `<project>/scripts/check-self-review-report.sh` when the project declares" through "never commit a report the guard could not read." with:

    > `land-self-review-report.sh` runs `<repo>/scripts/check-self-review-report.sh` over the
    > report's directory when `<repo>` carries that script, and refuses a violation —
    > `LAND-REPORT-INVALID`, exit 5, nothing committed: fix the lines the check names and land again.
    > **Exit 2 — the guard cannot answer at all (a missing or unreadable target directory, an
    > unreadable report inside it, an internal coverage.sh call failing) — is not a violation**: the
    > chain refuses it as `LAND-REPORT-UNCHECKED`, exit 5; report it and stop, never commit a report
    > the guard could not read.

    In `skills/flow/integrate.md` and in `skills/flow-fast/SKILL.md`, replace the line "A `LAND-UNTRACED-FINDING` or `LAND-DEFERRED-FINDING` refusal is not a stop: resolve it per step 7 of `/flow-self-review` (`skills/flow-self-review/SKILL.md`) and land again." with:

    > A `LAND-UNTRACED-FINDING`, `LAND-DEFERRED-FINDING` or `LAND-REPORT-INVALID` refusal is not a stop: resolve it per step 7 of `/flow-self-review` (`skills/flow-self-review/SKILL.md`) and land again.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestLandSelfReviewReport' -count=1` passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-self-review-report.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/landselfreviewreport.go`, `stats/internal/guard/land_self_review_report_test.go`, `scripts/land-self-review-report.sh`, `skills/flow-self-review/SKILL.md`, `skills/flow-fast/SKILL.md`, `skills/flow/integrate.md`
**Tests:** `TestLandSelfReviewReportRefusesAReportFailingItsCheck`
**Regression:** `TestLandSelfReviewReportRefusesAReportFailingItsCheck` fails — a report `check-self-review-report` refuses lands on the base again.
**Commit:** fix(guard): refuse a self-review report that fails its check at landing
**After:** Task 1

**Decision:** land-runs-report-check

- [ ] 3. Check the change worktree, not the cwd, before an inline begin
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Write Step 2's test first and run `cd stats && go test ./cmd/flow -run 'TestRecordDispatchBeginInlineReadsTheChangeWorktree$' -count=1`: it fails — the begin exits 2 naming the main checkout's `other.txt`.
  - [ ] **Step 2: Write the test** `TestRecordDispatchBeginInlineReadsTheChangeWorktree` after `TestRecordDispatchBeginRefusesAnInlineRowOverEarlierWork` in `stats/cmd/flow/record_test.go`, its `begin` closure that test's with `-change kan-259`, `-key task-6-implementer` and `-C <main>`: `main := gitRepo(t)`; `isolatedStateRoot(t)`; commit `other.txt` in `main`; `git -C main worktree add -q -b spectre/kan-259 <t.TempDir()>/wt`; in `wt` write `spectre/changes/kan-259/tasks.md` and `design.md` and commit them; modify `main/other.txt`. Assert: `begin("-agent-id", "inline")` exits 0 (the main checkout's dirt is not the change's); then modify `wt/spectre/changes/kan-259/design.md` and assert exit 2 with stderr naming `spectre/changes/kan-259/design.md` and not `other.txt`.
  - [ ] **Step 3: Implement** per decision `inline-begin-change-worktree` in `stats/cmd/flow/record.go`: `changeCheckouts(dir, change string) (main string, carrying []string, err error)` holds the worktree scan now inside `requireChangeCheckout`, which keeps its name validation and both refusal messages and calls it. At the inline check, `tree := f.dir`, then the first `carrying` entry that is not `main` replaces it; `inlineDirtyPaths(tree, f.change)`. A `changeCheckouts` error leaves `tree` at `f.dir`. Update `inlineDirtyPaths`' comment to say its caller passes the change's worktree (kan-922's self-review: a /flow parent's cwd is the main checkout).
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && go test ./cmd/flow -run 'TestRecordDispatchBeginInlineReadsTheChangeWorktree$|TestRecordDispatchBeginRefusesAnInlineRowOverEarlierWork$|TestRequireChangeCheckout' -count=1` passes; `cd stats && make -s web-dist && go vet ./cmd/flow` exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBeginInlineReadsTheChangeWorktree`
**Regression:** `TestRecordDispatchBeginInlineReadsTheChangeWorktree` fails — another session's edits in the main checkout refuse an inline begin whose change worktree is clean.
**Commit:** fix(stats): check the change worktree, not the cwd, before an inline begin
**After:** none

**Decision:** inline-begin-change-worktree

- [ ] 4. Name break-and-prove's test-command directory in the fixer brief
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'repository root' skills/flow/review-panel.md` prints no line in the `break-and-prove.sh` paragraph of the fixer brief; `grep -n 'it runs from the repository root' scripts/break-and-prove.sh` names the header's statement.
  - [ ] **Step 2: Implement** per decision `brief-break-and-prove-root`: in `skills/flow/review-panel.md`'s fixer brief, replace the line `> asserts the restored run passes. Each mutation alters one mechanism — where a single` so the passage reads:

    > \> asserts the restored run passes. `<test-command>` runs from the repository root, so a runner
    > \> rooted in a subdirectory names it with its own directory flag — `go -C stats test ./internal/guard` in a
    > \> repository whose `go.mod` sits in `stats/`. Each mutation alters one mechanism — where a single

    each line keeping the brief's `> ` prefix, the rest of the paragraph unchanged.
  - [ ] **Step 3: Verify.** `scripts/check-dispatch-paragraphs.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** none — a brief sentence; `check-verbatim-moves.sh` pins its wording.
**Regression:** none — prose only.
**Commit:** docs(flow): name break-and-prove's test-command directory in the fixer brief
**After:** none

**Decision:** brief-break-and-prove-root
