# self-review-kan-960-group-session-floor-hidden-tab-polling

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-16-self-review-kan-960-group-session-floor-hidden-tab-polling-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`. Never merge or push either repository.
- **Task 3 commits in the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-960-group-session-floor-hidden-tab-polling`** (branch `spectre/kan-960-group-session-floor-hidden-tab-polling`), never in this agents worktree; Tasks 1 and 2 commit in this agents worktree.
- Each task's change is the unified diff its step gives, applied with `git apply` from the root of the repository its task names — every diff was applied and checked there (`design.md`, **Context**). A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line a diff adds is copied verbatim: this change's `verbatim-moves.txt` acknowledges each new or reworded sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`. Task 1's verify step vets and tests only `./internal/store/`.

## Review focus

- A begin and end recorded in one Bash call must still be refused: the one-second case of `TestEndDispatchRefusesAZeroWidthInlineRow` stays (Task 1).
- The SCRATCH HARNESS paragraph must sit among the implementer dispatch paragraphs the parent copies into every implementer prompt, not in the parent's own prose (Task 2).
- Gymie's `CLAUDE.md` and `AGENTS.md` Frontend sections must stay identical (Task 3).

- [ ] 1. Lower the inline-window floor to three seconds
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'const minInlineWindow' stats/internal/store/records.go` and confirm `const minInlineWindow = 10 * time.Second` — the defect: a `completed` inline row closed 7.6 s after its begin is refused.
  - [ ] **Step 2: Write the failing test.** Apply the test hunks alone (`git apply --include='stats/internal/store/records_test.go' <the diff in step 3>`) and run `cd stats && go test ./internal/store/ -run TestEndDispatchRefusesAZeroWidthInlineRow -count=1`: it FAILS with `EndDispatch seven seconds after an inline begin: store: inline dispatch window too short … closed 7s after its begin`.
  - [ ] **Step 3: Implement the fix.** Apply the rest of this diff (`git apply --exclude='stats/internal/store/records_test.go' <the diff>`):

````diff verified:git apply --check at 7890df95; gofmt -l ./internal/store prints nothing, go vet ./internal/store/ and go test ./internal/store/ -run TestEndDispatchRefusesAZeroWidthInlineRow -count=1 pass with it applied; with records.go reverted the test fails at records_test.go:3990 (closed 7s after its begin)
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index d7d88161..4b527a12 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -189,7 +189,7 @@ these substitutions:
   inline work under the same roles a dispatched run would use. Every inline row's `dispatch
   begin` is recorded when the parent starts that work, before its first edit, and its `dispatch
   end` once the work's commit lands — never the two together, which records the work as taking
-  no time; `flow record dispatch end` refuses a `completed` inline row closed under ten seconds
+  no time; `flow record dispatch end` refuses a `completed` inline row closed under three seconds
   after its begin when it carries no `-tokens` report.
 
 ## 1. Load context and validate the plan
diff --git a/stats/internal/store/records.go b/stats/internal/store/records.go
index 36b22f8f..875b5ea4 100644
--- a/stats/internal/store/records.go
+++ b/stats/internal/store/records.go
@@ -107,9 +107,11 @@ var ErrAgentIDInvalid = errors.New("store: invalid agent id")
 var ErrInlineWindowTooShort = errors.New("store: inline dispatch window too short")
 
 // minInlineWindow is the shortest window an inline row may close with.
-// One Bash call lands its begin and end milliseconds apart; real inline
-// work takes minutes.
-const minInlineWindow = 10 * time.Second
+// One Bash call lands its begin and end milliseconds apart, and that is
+// the case the floor exists to refuse; real inline work takes seconds at
+// the least -- a cached docs-only verify closed in 7.6 s and was refused
+// under the old ten-second floor (kan-960).
+const minInlineWindow = 3 * time.Second
 
 // agentIDPlaceholders is the closed set of stand-in words a dispatcher
 // types when it does not hold the harness's identifier yet. Every word
diff --git a/stats/internal/store/records_test.go b/stats/internal/store/records_test.go
index d78034d0..73db9250 100644
--- a/stats/internal/store/records_test.go
+++ b/stats/internal/store/records_test.go
@@ -3949,8 +3949,9 @@ func TestEndDispatchRejectsInvalidAgentID(t *testing.T) {
 }
 
 // TestEndDispatchRefusesAZeroWidthInlineRow pins the inline window floor:
-// an inline row closed seconds after its begin, with no token report, is
-// refused and stays open; closed a minute later it closes.
+// an inline row closed one second after its begin, with no token report, is
+// refused and stays open; closed seven seconds later -- a cached docs-only
+// verify (kan-960) -- it closes.
 func TestEndDispatchRefusesAZeroWidthInlineRow(t *testing.T) {
 	st, _ := newRecordStore(t)
 	ctx := context.Background()
@@ -3984,9 +3985,9 @@ func TestEndDispatchRefusesAZeroWidthInlineRow(t *testing.T) {
 	if _, err := st.EndDispatch(ctx, projectKey, "kan-1", blocked); err != nil {
 		t.Fatalf("EndDispatch a blocked inline row one second after its begin: %v", err)
 	}
-	end.EndedAt = begin.StartedAt.Add(time.Minute)
+	end.EndedAt = begin.StartedAt.Add(7 * time.Second)
 	if _, err := st.EndDispatch(ctx, projectKey, "kan-1", end); err != nil {
-		t.Fatalf("EndDispatch a minute after an inline begin: %v", err)
+		t.Fatalf("EndDispatch seven seconds after an inline begin: %v", err)
 	}
 }
 
````

  - [ ] **Step 4: Verify.** `cd stats && gofmt -l ./internal/store` prints nothing; `go vet ./internal/store/` and `go test ./internal/store/ -run TestEndDispatchRefusesAZeroWidthInlineRow -count=1` pass; from the repository root `scripts/check-verbatim-moves.sh` and `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `skills/flow/implement.md`
**Tests:** `TestEndDispatchRefusesAZeroWidthInlineRow`
**Regression:** reverting `records.go` refuses the seven-second close, and the test fails with `inline dispatch window too short`.
**Baseline:** before=79 after=79
<!-- measured: grep -c '^func Test' stats/internal/store/records_test.go @ 7890df95 (before; after with this task's diff applied) -->
**Commit:** fix(store): lower the inline dispatch window floor to three seconds
**After:** none

**Decision:** inline-floor-three-seconds

- [ ] 2. The implementer dispatch keeps a throwaway harness out of new worktrees
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n 'SCRATCH HARNESS\|scratchpad' skills/flow/implement.md` and confirm no hit — the defect: nothing tells an implementer where a throwaway harness lives.
  - [ ] **Step 2: Apply the paragraph.** Apply this diff from the repository root:

````diff verified:git apply --check at 7890df95; with Tasks 1 and 2 applied check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-installed-citations.sh and check-ask-silence.sh exit 0, and check-verbatim-moves.sh prints only the four sentences verbatim-moves.txt lists
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index d7d88161..19898e77 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -596,6 +596,12 @@ and applies **The handshake** stated above, unchanged.
 > the shape you construct by hand is not the shape the real producer emits. Build the value the way
 > production builds it, or assert against the real boundary.
 
+> **SCRATCH HARNESS:** A throwaway harness — a measuring spec, a repro script, a probe — lives in
+> your session scratchpad or in this change's state dir (the path `flow state dir` prints), never
+> in a new worktree or branch of any repository. Cleanup removes only the change's own worktrees
+> and branches, so one you create outlives the run (kan-960 left a `kan-960-measure` worktree and
+> branch in `gymie-playwright`).
+
 > **REPORT FILE:** write your report to
 > `<abs-worktree>/.superpowers/sdd/implementer-report-<k>.md` as your **last** act — after your
 > commit and your final test run — carrying each commit's sha, the failing RED output you saw,
````

  - [ ] **Step 3: Verify.** From the repository root `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-dispatch-paragraphs.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** **none** — a prose paragraph; `check-verbatim-moves.sh` pins its exact text.
**Regression:** reverting the commit drops the paragraph from every implementer prompt.
**Commit:** docs(flow): keep an implementer's throwaway harness out of new worktrees
**After:** Task 1

**Decision:** scratch-harness-paragraph

- [ ] 3. gymie names Playwright's hidden-tab limits
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** From `/Users/tweety53/Projects/gymie-worktrees/kan-960-group-session-floor-hidden-tab-polling`, run `grep -n 'focus emulation' CLAUDE.md AGENTS.md` and confirm no hit, and `grep -n 'focus emulation' docs/audits/kan-924-call-efficiency.md` and confirm the F16 row — the defect: the measurement lives only in an audit no planner reads.
  - [ ] **Step 2: Apply the note.** From `/Users/tweety53/Projects/gymie-worktrees/kan-960-group-session-floor-hidden-tab-polling`, apply this diff:

````diff verified:git apply --check in the gymie worktree at b4efac39; applied and reverted there, the two files' Frontend sections identical with it applied
diff --git a/AGENTS.md b/AGENTS.md
index 8cc2a2c8..41548c73 100644
--- a/AGENTS.md
+++ b/AGENTS.md
@@ -54,6 +54,11 @@ Home that drew an empty calendar and "Nothing planned" before its data landed.
 - **A green suite is not evidence of a frame.** Verify a swap or a load by recording frames — a CDP
   screencast against the running app, on the phone over USB debugging when the report came from it —
   not by reading the diff.
+- **Playwright cannot hide a tab.** Its focus emulation keeps every page it launches or attaches
+  `visible`, and CDP `/json/activate` fires `visibilitychange` but never the window `blur`/`focus` a
+  user's tab switch fires — the event Compose's lifecycle pauses on. Measure hidden-tab behaviour in
+  a headed Chrome Playwright did not launch, dispatching `blur`/`focus` by hand around each CDP tab
+  switch (kan-960; `docs/audits/kan-924-call-efficiency.md`, F16).
 - **Start `adb` outside every worktree.** The first `adb` call forks a server that keeps its caller's
   working directory, and a server started inside a worktree holds it until killed — `/flow`'s cleanup
   then stops on it. Run `(cd ~ && adb start-server)` before any other `adb` command.
diff --git a/CLAUDE.md b/CLAUDE.md
index 459abd68..fc02cd42 100644
--- a/CLAUDE.md
+++ b/CLAUDE.md
@@ -51,6 +51,11 @@ Home that drew an empty calendar and "Nothing planned" before its data landed.
 - **A green suite is not evidence of a frame.** Verify a swap or a load by recording frames — a CDP
   screencast against the running app, on the phone over USB debugging when the report came from it —
   not by reading the diff.
+- **Playwright cannot hide a tab.** Its focus emulation keeps every page it launches or attaches
+  `visible`, and CDP `/json/activate` fires `visibilitychange` but never the window `blur`/`focus` a
+  user's tab switch fires — the event Compose's lifecycle pauses on. Measure hidden-tab behaviour in
+  a headed Chrome Playwright did not launch, dispatching `blur`/`focus` by hand around each CDP tab
+  switch (kan-960; `docs/audits/kan-924-call-efficiency.md`, F16).
 - **Start `adb` outside every worktree.** The first `adb` call forks a server that keeps its caller's
   working directory, and a server started inside a worktree holds it until killed — `/flow`'s cleanup
   then stops on it. Run `(cd ~ && adb start-server)` before any other `adb` command.
````

  - [ ] **Step 3: Verify.** From `/Users/tweety53/Projects/gymie-worktrees/kan-960-group-session-floor-hidden-tab-polling`: `diff <(sed -n '/^### Frontend/,/^---/p' CLAUDE.md) <(sed -n '/^### Frontend/,/^---/p' AGENTS.md)` prints nothing, and `git diff --cached --stat` shows `AGENTS.md` and `CLAUDE.md` alone. Commit there.

**Build:** green
**Files:** `CLAUDE.md`, `AGENTS.md` (gymie repository, worktree `/Users/tweety53/Projects/gymie-worktrees/kan-960-group-session-floor-hidden-tab-polling`)
**Tests:** **none** — instruction prose.
**Regression:** reverting the commit leaves a gymie planner to plan a hidden-tab harness on Playwright again.
**Commit:** docs: name Playwright's hidden-tab limits for frontend measurement
**After:** none

**Decision:** playwright-note-in-claude-md
