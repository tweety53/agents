# self-review-kan-1064-flow-self-review-run-the-coordination-around-the

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the four non-`big` findings (F1–F4) of `kan-1064-flow-self-review-run-the-coordination-around-the`'s self-review pass.

**Architecture:** the self-review SKILL states the pipeline-fix loop's escalation and scopes its own
no-escalation rule; the planner runs the lint list on a scratch copy before tagging `verified:`;
`make restart` marks its restart and the client retries a refused dial while the mark is fresh;
`pipeline-fix.sh end` closes the pass row.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-31-self-review-kan-1064-flow-self-review-run-the-coordination-around-the-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in the worktree `/Users/tweety53/Projects/agents-worktrees/kan-1064-flow-self-review-run-the-coordination-around-the`, branch `spectre/kan-1064-flow-self-review-run-the-coordination-around-the`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- The prose edits land character for character through the script below, saved to a scratch file outside the worktree and run as `python3 -I <saved script> "$PWD" <task number>` from the worktree's root; it asserts each anchor matches exactly once. `scripts/check-verbatim-moves.sh` matches the result against `verbatim-moves.txt` — a hand edit that drifts by one character fails it. A run that asserts is reported, never hand-merged.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container — never run `make restart`; read its recipe with `make -n`.

```python verified:applied with task numbers 1 and 2 to a scratch clone at 6107fdfd with this change's verbatim-moves.txt; scripts/run-verify-list.sh <clone> lint then exited 0 on every line but check-installed-rules.sh, which names the clone not being the installed checkout (exit 0 in this worktree)
import sys
root = sys.argv[1]
def sub(path, old, new):
    p = f"{root}/{path}"
    s = open(p).read()
    assert s.count(old) == 1, (path, old[:60], s.count(old))
    open(p, "w").write(s.replace(old, new))
which = sys.argv[2:]
if "1" in which:
    sub("skills/flow-self-review/SKILL.md",
        "The `fix` row carries no no-progress escalation: a fix that makes no progress is answered by the fresh re-plan above, never by `flow-high`.",
        "The `self-review-fix` row carries no no-progress escalation: a self-review fix that makes no progress is answered by the fresh re-plan above, never by `flow-high`. The `pipeline-fix-fix` row keeps it: a pipeline-fix round that made no progress is dispatched again on that row's `opus/flow-high` escalation, admitted only after `<agents repo>/scripts/pipeline-fix.sh escalate <key>` (**Fewest operator actions**, `skills/flow-contracts/pipeline.md`).")
    sub("skills/flow/implement.md",
        "the self-review loop's own escalation rows included",
        "the pipeline-fix loop's `pipeline-fix-fix` escalation row included")
    sub("skills/flow/SKILL-rationale.md",
        "as in the in-run and self-review fix\nloops,",
        "as in the in-run fix\nloop,")
if "2" in which:
    sub("skills/flow/brainstorm-planner.md",
        "provenance** (`skills/flow-contracts/plan-provenance.md`).\n\n",
        "provenance** (`skills/flow-contracts/plan-provenance.md`).\n\n"
        "**A block verified on a scratch copy has passed the lint list there.** When planning verifies a\n"
        "fenced block by applying it to a scratch copy of the branch — a diff, or a script that edits the\n"
        "tree — it runs `<agents repo>/scripts/run-verify-list.sh <scratch copy> lint` on that copy with\n"
        "the block applied before writing the tag, fixes every red line in the block itself, and the\n"
        "`verified:` tag names that run. A block the list was not run over is tagged\n"
        "`unverified:run-verify-list.sh lint on the applied copy`, never `verified:` (kan-1064\n"
        "self-review: task 2's prose diff, tagged `verified:` after a scratch apply, failed\n"
        "`check-installed-citations.sh` and `check-guard-symlinks.sh` at implementation).\n\n")
```

## Review Focus

- Task 1: no sentence left forbids the `pipeline-fix-fix` escalation, and the self-review loop's own no-escalation rule still stands.
- Task 2: the rule makes the lint run on the scratch copy a precondition of `verified:`, not an extra.
- Task 3: a refused dial with no fresh marker still fails at once; the retry never outlives the request's context; the marker is removed on every exit of the recipe after it is written; the compose and CLI-install steps run before the kill.
- Task 4: the pass is closed under the same change and checkout its begin used, through one shared helper.

- [ ] 1. Scope the no-escalation sentence to self-review-fix and restore the pipeline-fix escalation
  - [ ] **Step 1: Prose** — run the Global constraints script with task number `1`.
  - [ ] **Step 2: Verify** — `git grep -n -F -e 'The `fix` row carries no no-progress escalation' -e "the self-review loop's own escalation rows" -e 'in-run and self-review fix' -- skills` prints nothing; `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/kan-1064-flow-self-review-run-the-coordination-around-the lint` is green.

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`, `skills/flow/implement.md`, `skills/flow/SKILL-rationale.md`
**Tests:** **none** — a contract sentence; `check-verbatim-moves.sh` holds its wording.
**Regression:** none — prose; reverting restores the sentence that forbids the `pipeline-fix-fix` escalation.
**Commit:** fix(flow-self-review): scope the no-escalation rule to self-review-fix and keep the pipeline-fix escalation
**After:** none

**Decision:** escalation-scoped-to-pipeline-fix

- [ ] 2. Make the planner run the lint list on a scratch copy before tagging verified
  - [ ] **Step 1: Prose** — run the Global constraints script with task number `2`.
  - [ ] **Step 2: Verify** — `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/kan-1064-flow-self-review-run-the-coordination-around-the lint` is green.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** **none** — a planner rule; `check-verbatim-moves.sh` holds its wording.
**Regression:** none — prose; reverting drops the rule.
**Commit:** fix(flow): run the lint list on a scratch copy before tagging a plan block verified
**After:** none

**Decision:** scratch-copy-runs-lint

- [ ] 3. Retry a refused dial while make restart swaps flowd
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `stats/internal/client/client_test.go` add three tests, each first calling `t.Setenv("TMPDIR", t.TempDir())`:
    - `TestGetChangeRetriesRefusedDialWhileFlowdRestarts` — reserve a port with `net.Listen("tcp", "127.0.0.1:0")` and close it; write the marker `filepath.Join(os.TempDir(), fmt.Sprintf("flowd-%d.restarting", port))`; start a goroutine that after 300ms listens on that same port and serves `genuineDaemon` answering 200 with `{}` (closed in `t.Cleanup`); `GetChange` under a 2s context succeeds.
    - `TestGetChangeFailsFastOnRefusedDialWithoutRestartMarker` — `deadPortURL`, no marker, a 2s context: `GetChange` returns an error wrapping `ErrUnavailable` in under 500ms.
    - `TestGetChangeIgnoresStaleRestartMarker` — `deadPortURL`, its marker written and its mtime set 3 minutes back with `os.Chtimes`, a 2s context: `ErrUnavailable` in under 500ms.
    <!-- predicted: the 300ms, 2s and 500ms figures are test parameters this plan chooses, not measurements -->

    Run `cd stats && make -s web-dist && go test ./internal/client/ -run 'TestGetChange' -count=1`: `TestGetChangeRetriesRefusedDialWhileFlowdRestarts` fails before the fix.
  - [ ] **Step 2: Retry in `do`** (`stats/internal/client/client.go`) per decision `retry-refused-while-restarting`: on an error from `c.httpClient.Do` that `errors.Is(err, syscall.ECONNREFUSED)`, while `restarting(req.URL.Port())` holds — the marker `filepath.Join(os.TempDir(), "flowd-<port>.restarting")` exists and its mtime is under 2 minutes old — and the request's body can be replayed (`req.Body == nil`, or `req.GetBody != nil`, a fresh body taken from it per attempt), wait 100ms or until `req.Context().Done()`, then send again; a done context returns the last error wrapped in `ErrUnavailable` as today. Rewrite the `Client` doc comment's "It never retries" to state this one retry and why it cannot hold up a dead daemon's fallback; name the marker's writer, `stats/Makefile`'s `restart`, in `restarting`'s comment.
    <!-- predicted: the 100ms interval and 2-minute staleness are this plan's chosen values (decision retry-refused-while-restarting), not measurements -->
  - [ ] **Step 3: The restart recipe** (`stats/Makefile`): add `override LIVE_MARKER = $${TMPDIR:-/tmp}/flowd-$(LIVE_PORT).restarting` beside `LIVE_LOG`; in `restart:`, move the `docker compose up -d` line and the CLI install (`mkdir`, `go build -o $(LIVE_CLI)`, its echo), each with its comment, ahead of the kill line; the kill line first runs `touch "$(LIVE_MARKER)"`, polls `kill -0` with `sleep 0.1` for up to 300 rounds (still 30s), and runs `rm -f "$(LIVE_MARKER)"` before its `exit 1`; the readiness loop runs `rm -f "$(LIVE_MARKER)"` before both its `exit 0` and its `exit 1`. Add a comment paragraph above `LIVE_MARKER` naming its reader (`do`, `stats/internal/client/client.go`) and why the compose and CLI steps now run before the kill (the gap is the daemon swap alone).
  - [ ] **Step 4: Verify** — `gofmt -w` the Go files; `cd stats && make -s web-dist && go test ./internal/client/ -run 'TestGetChange' -count=1` passes; `cd stats && make -s web-dist && go test ./internal/repotest/ -run 'TestMakeBuild' -count=1` passes; `make -n -C stats restart | grep -n -e 'compose up' -e 'go build -o' -e 'kill "' -e restarting` lists `compose up` and `go build -o` before the kill; `cd stats && make -s web-dist && go vet ./internal/client/ && gofmt -l .` is clean.

**Build:** green
**Files:** `stats/internal/client/client.go`, `stats/internal/client/client_test.go`, `stats/Makefile`
**Tests:** `TestGetChangeRetriesRefusedDialWhileFlowdRestarts`, `TestGetChangeFailsFastOnRefusedDialWithoutRestartMarker`, `TestGetChangeIgnoresStaleRestartMarker`
**Regression:** `TestGetChangeRetriesRefusedDialWhileFlowdRestarts` — reverting the commit fails the first refused dial at once, so `GetChange` returns `ErrUnavailable`; the other two fail if the retry ignores the marker or its age.
**Baseline:** before=45 after=48
<!-- measured: grep -cE '^func Test' stats/internal/client/client_test.go @ 6107fdfd (45) -->
**Commit:** fix(stats): retry a refused flowd dial while make restart swaps the daemon
**After:** none

**Decision:** retry-refused-while-restarting

- [ ] 4. Close the self-review pass row in pipeline-fix.sh end
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `stats/internal/guard/pipeline_fix_test.go`, replace `TestPipelineFixEndPassRecordsNothing` and its comment with `TestPipelineFixEndClosesThePass`: `fx := pfNew(t, "[]")`; `fx.pf(t, "end", "self-review-kan-1-demo-pass", "completed")` exits 0 and `fx.calls(t)` contains `record dispatch end` with `-change kan-1-demo -key self-review-kan-1-demo-pass -session-token self-review-kan-1-demo -outcome completed`; a second call with outcome `done` exits 2. Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestPipelineFixEndClosesThePass' -count=1`: it fails before the fix.
  - [ ] **Step 2: Fix** per decision `end-closes-the-pass`: in `stats/internal/guard/loop.go`, extract the pass begin's target into `lpPassTarget(env Env, git string, lp <its loop type>) (change, main string)` — change `strings.TrimSuffix(strings.TrimPrefix(lp.prefix, "self-review-"), "-")`, main `lpMain(env, git, env.Dir)` when `env.Dir != ""` and it resolves, `lp.main` otherwise — and call it from the begin. In `pfEnd` (`stats/internal/guard/pipelinefix.go`) drop the `PIPELINE-FIX-END-NA` branch; for a `pass` key take `-C` and `-change` from `lpPassTarget(env, pe.git, lp)`, every other key keeping `lp.main` and `lp.change`; the existing `completed`/`aborted` check covers the pass.
  - [ ] **Step 3: Verify** — `gofmt -w` the Go files; `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestPipelineFix|TestCheckLoopDispatch' -count=1` passes; `cd stats && make -s web-dist && go vet ./internal/guard/ && gofmt -l .` is clean; `git grep -n -F -e 'PIPELINE-FIX-END-NA' -- stats skills scripts` prints nothing.

**Build:** green
**Files:** `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/loop.go`, `stats/internal/guard/pipeline_fix_test.go`
**Tests:** `TestPipelineFixEndClosesThePass`
**Regression:** `TestPipelineFixEndClosesThePass` — reverting the commit prints `PIPELINE-FIX-END-NA` and calls no `record dispatch end`.
**Commit:** fix(guard): close the self-review pass row in pipeline-fix.sh end
**After:** none

**Decision:** end-closes-the-pass
