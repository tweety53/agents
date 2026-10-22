# self-review-kan-1006-list-mode-step-1-backend-workout-mode-preference

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1006's four self-review findings — F-A (Task 1), F-B (Task 2), F-C (Task 3),
F-D (Task 4), as `proposal.md` names them.

**Architecture:** one new dispatch paragraph rendered into slot prompts and required by the
paragraph guard, one gymie recipe sentence, one PreToolUse denial in the main-checkout hook with
its contract section, and a live store repair.

**Tech Stack:** Go (`stats/`), Python (`hooks/`), Bash, Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-22-self-review-kan-1006-list-mode-step-1-backend-workout-mode-preference-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `patches/` and `verbatim-moves.txt` are already committed with the plan. Never merge or push.
- Tasks 1, 3 and 4 run in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1006-list-mode-step-1-backend-workout-mode-preference` (`<agents-wt>` below); Task 2 in gymie's `/Users/tweety53/Projects/gymie-worktrees/kan-1006-list-mode-step-1-backend-workout-mode-preference` on branch `spectre/kan-1006-list-mode-step-1-backend-workout-mode-preference`. Every `**Files:**` path is relative to its task's worktree root. Never touch a main checkout, and never create a worktree.
- A task's diff is `<agents-wt>/spectre/changes/self-review-kan-1006-list-mode-step-1-backend-workout-mode-preference/patches/task-<n>.diff`, applied with `git apply <that path>` from the task's worktree root, in task order. Each patch was applied and checked in a scratch clone (agents `0e0d8a42`, gymie `894cd84c`), and `git apply --check` passed on each real worktree (task-3 after task-1). A patch that does not apply is reported, never hand-merged by guesswork.
- Prose lands character for character as its patch carries it: `scripts/check-verbatim-moves.sh` matches every added sentence against this change's `verbatim-moves.txt` exactly; with Tasks 1 and 3 applied in the scratch clone, `check-verbatim-moves.sh`, `check-markdown-integrity.py`, `check-references.sh`, `check-installed-citations.sh`, `check-normative-inventory.sh` and `check-dispatch-paragraphs.sh` exited 0.
- Go: run `gofmt -w` on every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Mutation proofs restore with `break-and-prove.sh` or `mutate-and-verify.sh`, or a saved diff re-applied with `git apply` — never `git checkout --` on a file carrying uncommitted edits (Task 3 makes that a hook denial).
- The gymie commit (Task 2) is a commit outside every task of kan-1006's change (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`): add `-m "Flow-Key: self-review-kan-1006-list-mode-step-1-backend-workout-mode-preference-fix"`; never touch that repository's `spectre/`.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container.

## Review Focus

- Task 1: the rendered slot prompt carries the block exactly once and no unresolved placeholder; `check-dispatch-paragraphs.sh` on the real tree reports 33 sites.
- Task 3: `git restore --staged`, a clean path, the main checkout's existing `restore` denial and every pre-existing harness case are unchanged; the hook still fails open on a git error.

- [x] 1. Carry the reproducer authoring contract on every slot prompt
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `<agents-wt>`, `grep -c 'REPRODUCER' stats/internal/guard/renderslotprompt.go` prints `0`, and kan-1006's rendered `/Users/tweety53/Projects/gymie-worktrees/kan-1006-list-mode-step-1-backend-workout-mode-preference/.superpowers/sdd/slot-prompt-0-primary+principles.md` names neither `# demonstrates:` nor `|| true`.
    <!-- measured: grep -c 'REPRODUCER' stats/internal/guard/renderslotprompt.go @ 0e0d8a42 → 0 -->
  - [x] **Step 2: Apply `patches/task-1.diff`** from `<agents-wt>`: it adds the **Every slot's dispatch prompt also carries the REPRODUCER paragraph** block to `skills/flow/review-panel.md` after REPRODUCE, DON'T READ; `REPRODUCER` to `render-slot-prompt`'s shared-label list (`stats/internal/guard/renderslotprompt.go`); the `reproducer` entry and its review-panel.md site (min 1) to `stats/internal/guard/dispatchparagraphs.go`; `dpReproducerBlock`, its place in the clean review-panel fixtures, the site count 33, and cases 100 and 101 to `stats/internal/guard/check_dispatch_paragraphs_test.go`; and `TestRenderSlotPromptCarriesReproducer` to `stats/internal/guard/render_slot_prompt_test.go`.
    <!-- measured: git apply --check patches/task-1.diff @ 0e0d8a42 → exit 0 -->
  - [x] **Step 3: Prove the tests pin the fix.** With `break-and-prove.sh` (or a saved diff re-applied): drop `"REPRODUCER", ` from the label list in `renderslotprompt.go` — `cd stats && go test ./internal/guard -run 'TestRenderSlotPromptCarriesReproducer' -count=1` fails; restore. Drop the `{"reproducer", "skills/flow/review-panel.md", 1, nil},` site row in `dispatchparagraphs.go` — `cd stats && go test ./internal/guard -run 'TestCheckDispatchParagraphs/case_10[01]' -count=1` fails its four subtests; restore.
    <!-- measured: both flips in a scratch clone at 0e0d8a42 with the patch → FAIL (1 test; 4 subtests) -->
  - [x] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckDispatchParagraphs|TestRenderSlotPrompt' -count=1` passes; `scripts/check-dispatch-paragraphs.sh` prints `… — 33 site(s) validated`; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-references.sh` and `scripts/check-installed-citations.sh` exit 0. Commit.
    <!-- measured: those commands in a scratch clone at 0e0d8a42 with the patch → ok; 33 site(s) validated -->

**Build:** green
**Files:** `skills/flow/review-panel.md`, `stats/internal/guard/renderslotprompt.go`, `stats/internal/guard/render_slot_prompt_test.go`, `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`
**Tests:** `TestRenderSlotPromptCarriesReproducer`, `case 100: review-panel.md without the REPRODUCER block exits 1`, `case 101: a REPRODUCER block without the zero-count rule exits 1`
**Regression:** `TestRenderSlotPromptCarriesReproducer` fails — slot prompts lose the reproducer contract again; cases 100 and 101 fail — the paragraph or its zero-count rule can be dropped from review-panel.md unnoticed.
**Baseline:** before=4 after=5 <!-- measured: grep -c -E '^func Test' stats/internal/guard/render_slot_prompt_test.go @ 0e0d8a42 → 4, and 5 with the patch in a scratch clone -->
**Commit:** fix(guard): render and require the REPRODUCER paragraph on every slot prompt
**After:** none

**Decision:** reproducer-dispatch-block

- [x] 2. Stop gymie's before backend by its port, never gradlew --stop
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In the gymie worktree, `grep -c 'stop the before backend\. If' .flow/project.md` prints `1`: the recipe names no way to stop it.
    <!-- measured: that grep in the gymie worktree @ 894cd84c → 1 -->
  - [x] **Step 2: Apply `patches/task-2.diff`** from the gymie worktree root: the sentence names `lsof -ti tcp:<free port> -sTCP:LISTEN | xargs kill` and why `./gradlew --stop` is wrong.
    <!-- measured: git apply --check task-2.diff in the gymie worktree @ 894cd84c → exit 0; lsof -ti tcp:4173 -sTCP:LISTEN printed flowd's one pid, exit 0, on this machine 2026-10-22 -->
  - [x] **Step 3: Verify.** `git diff --stat` names only `.flow/project.md`; `grep -c 'xargs kill' .flow/project.md` prints `1`. Commit in the gymie worktree with the `Flow-Key:` trailer the global constraints name.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** none — a recipe sentence in gymie's project file
**Regression:** none — documentation only
**Commit:** docs(flow): stop the before backend by its port, never gradlew --stop
**After:** none

**Decision:** before-backend-port-kill

- [x] 3. Deny a restore that would discard unstaged edits in a flow worktree
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In a throwaway linked worktree the harness builds, an unstaged edit to `f.txt` and a payload for `git -C <wt> checkout -- f.txt` pass the hook unchanged: Step 3's flip shows it — the harness's cases 151–153 read `allow`.
    <!-- measured: scripts/test-protect-main-checkout.sh in a scratch clone at 0e0d8a42 with the patch's harness cases and the hook's `if discard:` flipped to `if False:` → 3 FAIL (151, 152, 153) -->
  - [x] **Step 2: Apply `patches/task-3.diff`** from `<agents-wt>`, after Task 1's commit: it adds `discards_unstaged` and its call in `bash_hits`'s git-verb scan plus a docstring paragraph to `hooks/protect-main-checkout.py`, cases 151–155 to `scripts/test-protect-main-checkout.sh`, the `## Restoring a path with unstaged edits` section to `skills/flow-contracts/git-boundaries.md`, and its one-sentence citation after the MUTATION PROOF block in `skills/flow/review-panel.md`.
    <!-- measured: git apply --check patches/task-3.diff in a scratch clone at 0e0d8a42 after task-1.diff → exit 0 -->
  - [x] **Step 3: Prove the cases pin the hook.** With `break-and-prove.sh hooks/protect-main-checkout.py --sed 's/                    if discard:/                    if False:/' -- scripts/test-protect-main-checkout.sh`: cases 151, 152 and 153 fail under the flip and pass restored.
    <!-- measured: that flip by hand in a scratch clone at 0e0d8a42 → 3 FAIL; restored → all pass -->
  - [x] **Step 4: Verify.** `scripts/test-protect-main-checkout.sh` ends `all protect-main-checkout cases pass`; `python3 -m py_compile hooks/protect-main-checkout.py` exits 0; `scripts/check-python-suppressions.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-references.sh` and `scripts/check-installed-citations.sh` exit 0. Commit.
    <!-- measured: those commands in a scratch clone at 0e0d8a42 with both patches → ok -->

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`, `skills/flow-contracts/git-boundaries.md`, `skills/flow/review-panel.md`
**Tests:** `151 checkout -- of a path with unstaged edits`, `152 restore of a path with unstaged edits`, `153 checkout HEAD -- after a cd`, `154 restore --staged touches the index only`, `155 checkout -- of a clean path`
**Regression:** cases 151–153 fail — a `git checkout --` / `git restore` discarding an uncommitted fix in a flow worktree passes the hook again.
**Baseline:** before=138 after=143 <!-- measured: grep -c -E '^expect "[0-9]+' scripts/test-protect-main-checkout.sh @ 0e0d8a42 → 138, and 143 with the patch in a scratch clone -->
**Commit:** fix(hooks): deny a restore that would discard unstaged edits in a worktree
**After:** Task 1

**Decision:** restore-denial-in-protect-hook

**Decision:** restore-rule-home

- [x] 4. Restamp kan-1006's panel-1 dispatch rows
  - [x] **Step 1: Reproduce before touching anything; do not guess.** From `/Users/tweety53/Projects/gymie`, `flow record dispatches -change kan-1006-list-mode-step-1-backend-workout-mode-preference | python3 -c 'import json,sys; print([(d["id"], d.get("agentId")) for d in json.load(sys.stdin) if d["id"] in (6966, 6967)])'` prints `[(6966, None), (6967, None)]`.
    <!-- measured: that command @ dev store 2026-10-22 → [(6966, None), (6967, None)] -->
  - [x] **Step 2: Repair the two rows** from `<agents-wt>/stats`: `make -s web-dist`, then `FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd restamp-dispatch 6966 /Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/a228ba0f-353c-4693-9969-641bd2c0d333/subagents/agent-a62f95e81cc0b07c5.jsonl` and the same with `6967` and `…/subagents/agent-a01308edf624bb805.jsonl` in that directory.
    <!-- measured: both transcripts exist (525074 and 535673 bytes); `restamp-dispatch` is a flowd subcommand (stats/cmd/flowd/main.go:87) -->
  - [x] **Step 3: Verify.** Step 1's command prints `[(6966, 'a62f95e81cc0b07c5'), (6967, 'a01308edf624bb805')]`, and both rows' `metrics` are no longer `{}`; report both outputs.

**Build:** green
**Files:** none — a store repair, no file changes
**Tests:** none — a live repair, verified by Step 3's query
**Regression:** none — no code changes; 90efd086 fixed the race forward
**Commit:** none — this task commits nothing; it repairs two flow-store rows
**After:** none

**Decision:** store-repair-no-commit
