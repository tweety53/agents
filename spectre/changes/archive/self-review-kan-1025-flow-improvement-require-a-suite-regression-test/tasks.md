# self-review-kan-1025-flow-improvement-require-a-suite-regression-test

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the three non-`big` findings of `kan-1025-flow-improvement-require-a-suite-regression-test`'s self-review pass.

**Architecture:** two prose edits to `skills/flow/review-panel.md` (the mutation brief, the slot ceiling) and one branch in the main-checkout hook's heredoc stripper.

**Spec:** `design.md` beside this file.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1025-flow-improvement-require-a-suite-regression-test`, branch `spectre/kan-1025-flow-improvement-require-a-suite-regression-test`; `**Files:**` are relative to its root. Never touch the main checkout; never create a worktree.
- One commit per task with its `**Commit:**` subject, no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.
- Go test commands run as `cd stats && make -s web-dist && go test <pkg> -run '<name>' -count=1`, Bash timeout 600000.
- Committed text cites `kan-1025`, never a commit sha of this branch.

## Review Focus

- Task 1: the rendered mutation slot prompt carries the new sentences (the brief is rendered verbatim); `check-mutation-reproducer-pin.sh` stays clean — no new backticked `# mutation-…` span and no "first <N> lines" statement.
- Task 3: a heredoc to `bash`, `sh`, `sudo bash`, `x=$(sh`, `xargs` or `eval` keeps its body scanned; a data opener followed by `bash x.sh` in the same command keeps its body scanned; an unquoted body's `$(rm …)` is still judged.

- [x] 1. Let the mutation brief filter a large package by test-family prefix and state the inverted exits plainly
  - [x] **Step 1: Edit** `skills/flow/review-panel.md`, `### The mutation-testing brief`: after the sentence ending `never selected the killing `TestLoopLandingPromptTuningAdmissionBounds`).`, append, as part of the same paragraph: "When one run of that whole package takes more than a third of the slot's ceiling — measured once in the slot, before the first mutation — the command may instead filter by the name prefix the mutated code's test family shares (`-run '^TestCheckPanel'`), never a full test name, so a fix's new test under that prefix still kills the mutant (kan-1025: the guard package's 332 tests ran 165 s idle and about 455 s contended, the slot ran 55 minutes, and the parent narrowed the reproducer to one test's name). Its `# reproducer-bound:` is that measured run's seconds, doubled. Its exits follow the mutation-reproducer convention, never the generic one: it exits 0 when every test passes with the mutation in place — the mutant survived, the defect present — and non-zero when a test kills it."
    <!-- measured: the guard package's run figures (332 tests, 165 s idle, ~455 s contended) and the 55-minute slot are the self-review pass's evidence for kan-1025, quoted from the dispatch -->
  - [x] **Step 2: Verify.** `scripts/check-mutation-reproducer-pin.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-references.sh` exit 0; `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestRenderSlotPrompt' -count=1` passes.
  - [x] **Step 3: Commit** the file.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** none — prose rendered verbatim into the slot prompt; the pin guard above covers the marker.
**Regression:** none — no test.
**Commit:** docs(flow): let the mutation brief filter a large package by test-family prefix
**After:** none

**Decision:** fa-family-prefix-filter

- [x] 2. Name the wake-up that tracks a panel slot's wall-clock ceiling
  - [x] **Step 1: Edit** `skills/flow/review-panel.md`, `### No forking, and a wall-clock ceiling on every slot`: replace "The dispatcher\ntracks each in-flight slot's elapsed time itself rather than blocking indefinitely on a completion\nnotification." with "The dispatcher tracks each in-flight slot's elapsed time itself rather than blocking indefinitely on a completion notification: it notes each slot's dispatch instant (`date +%s`) when it dispatches, waits on the slot only through **Turn discipline**'s bounded loop (`skills/flow/implement.md`), and on every `still-running` compares the slot's elapsed time against its ceiling before re-issuing — a slot past its ceiling gets no further wait and takes the breach path below at once (kan-1025: with no named wake-up, a breach was caught at about three times the ceiling)." Leave the rest of the paragraph, and `skills/flow/implement.md`, unchanged.
  - [x] **Step 2: Verify.** `scripts/check-dispatch-paragraphs.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-references.sh` exit 0.
  - [x] **Step 3: Commit** the file.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** none — prose.
**Regression:** none — no test.
**Commit:** docs(flow): name the wake-up that tracks a panel slot's ceiling
**After:** Task 1

**Decision:** fb-bounded-wait-elapsed-check

- [x] 3. Read a heredoc body as data when nothing in the command runs it
  - [x] **Step 1: Write the failing test** `TestProtectMainCheckoutReadsADataHeredocBodyAsData` in `stats/internal/repotest/protect_main_checkout_test.go`, `t.Parallel()`, building its own main checkout the way `TestProtectMainCheckout` does (bare origin, clone to `{ROOT}/main`, one commit, `branch -M main`, `push -u origin main`, `remote set-head origin main`) and a sibling dir `{SIB}` outside it. Cases through `f.bash`:
    - allow, cwd `{MAIN}`: `flow record summary -change x -file - <<'EOF'\nstep 1 -> step 2; rm {MAIN}/f.txt > {MAIN}/g.txt\nEOF`
    - deny, cwd `{ROOT}`: `bash <<'EOF'\nrm {MAIN}/f.txt\nEOF`
    - deny, cwd `{ROOT}`: `sudo sh <<'EOF'\nrm {MAIN}/f.txt\nEOF`
    - deny, cwd `{ROOT}`: `x=$(sh <<'EOF'\nrm {MAIN}/f.txt\nEOF\n)`
    - deny, cwd `{ROOT}`: `flow record summary -change x -file - <<EOF\n$(rm {MAIN}/f.txt)\nEOF`
    - deny, cwd `{ROOT}`: `flow record summary -change x -file - <<'EOF'\nrm {MAIN}/f.txt\nEOF\nbash {SIB}/x.sh`
    - deny, cwd `{ROOT}`: `flow record summary -change x -file - <<'EOF' | sh\nrm {MAIN}/f.txt\nEOF`
    Run it; the first allow case fails (denied today).
    <!-- measured: python3 hooks/protect-main-checkout.py fed {"tool_name":"Bash","tool_input":{"command":"flow record summary -change x -file - <<'EOF2'\nstep 1 -> step 2; rm nothing > here\nEOF2"},"cwd":"/Users/tweety53/Projects/agents"} printed permissionDecision deny @ spectre/kan-1025-flow-improvement-require-a-suite-regression-test -->
  - [x] **Step 2: Fix** `hooks/protect-main-checkout.py`:
    - `may_run_code(text, target=None)`: with `target` None, skip the `names_target` arm; update its docstring by one clause.
    - Extract the `cat`/`tee` branch's unquoted-body span block into `subst_spans(body)` — returns the list of kept spans (a backtick span rewritten as `$(…)`), `[]` when the body has no `$(`/backtick or the delimiter is quoted (caller passes only unquoted bodies), and `None` when a span cannot be closed; the `cat`/`tee` branch calls it with unchanged behaviour.
    - In `strip_data_heredocs`, after the `PROGRAM_RUNNERS` branch and before `if not words or words[0] not in ("cat", "tee")`: when `end is not None`, `words` is non-empty, `words[0]` is not `cat`/`tee`, no token of `seg` split on whitespace, `=`, `$`, `(`, `)`, backtick and both quotes (so `x=$(bash` yields `bash`) has a file name (`rsplit("/", 1)[-1]`) in `RUNNER_WORDS`, and `not may_run_code("\n".join(lines[:k] + lines[end + 1 :]), None)`: for a quoted delimiter drop the body; for an unquoted one with `$(`/backtick, extend `out` with `subst_spans(body)` (escaped backticks removed first, as the `cat`/`tee` branch does), or `continue` with the body scanned when it returns `None`; then `k = end + 1; continue`.
    - Extend the function's docstring by one sentence: any other opener's body is data when neither the opener nor the rest of the command runs code (`RUNNER_WORDS`, `may_run_code`) — kan-1025: `flow record summary -file - <<'EOF'` bodies were denied over their prose.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/repotest/ -run 'ProtectMainCheckout|TestCommitThenPush' -count=1` passes (the existing cases 163-170 included); `python3 -m py_compile hooks/protect-main-checkout.py`; `scripts/check-python-suppressions.sh` and `cd stats && gofmt -l . && go vet ./internal/repotest/` are clean.
  - [x] **Step 4: Commit** the two files.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go`
**Tests:** `TestProtectMainCheckoutReadsADataHeredocBodyAsData`
**Regression:** `TestProtectMainCheckoutReadsADataHeredocBodyAsData` — reverting denies the `flow record summary` heredoc from the main checkout and the allow case fails.
**Baseline:** before=5 after=6
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/repotest/protect_main_checkout_test.go @ spectre/kan-1025-flow-improvement-require-a-suite-regression-test -->
**Commit:** fix(hooks): read a heredoc body as data when nothing in the command runs it
**After:** none

**Decision:** fc-heredoc-body-is-data
