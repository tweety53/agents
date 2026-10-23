# self-review-kan-1027-flow-learning-setting

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the five non-`big` findings of `kan-1027-flow-learning-setting`'s self-review
pass — the learning note's `-C` after `/flow-fast` cleanup, the cwd-dependent commit chain, the
planner's missing verbatim-moves step, the panel-fix back-fill's early start, and the live check's
serial before figures.

**Architecture:** one fence line and one sentence in `skills/flow-contracts/learning.md`; the
COMMIT-PER-TASK chain in `skills/flow/implement.md` anchored with `git -C <abs-worktree>`; one
sentence in `.flow/project.md`'s `## lint` prose; a `panelFixFloor` helper in
`stats/cmd/flow/record.go` floors an inline panel-fix back-fill's derived start, with its test, its
`implement.md` sentence and a repair of kan-1027's row 7260; one sentence each in
`skills/flow/review-panel.md` and `skills/flow/verify-and-handoff.md`.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-23-self-review-kan-1027-flow-learning-setting-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan, listing every run-loaded sentence Tasks 1, 2, 4 and 5 add or reword. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-1027-flow-learning-setting`, branch `spectre/kan-1027-flow-learning-setting`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Each task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root; all five were applied together, in task order, and checked in this worktree at `a68fd756`. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Every prose task's verify step runs `scripts/check-verbatim-moves.sh` beside the other prose guards; it judges `skills/`, `commands-claude/` and `rules/`, and is run on Task 3 too, where it must stay at 0.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container; Task 4's repair writes one row through `flowd`'s repair subcommands, run with `go run`, and stops nothing.

## Review Focus

- Task 4: a `panel-fix-<r>-<chunk>` key floors at round `r-1` like its first chunk; a `-retry` suffix parses the same way.
- Task 4: a record read that fails, or a round with no ended reviewer row, keeps the commit-derived start — an outage is never a refusal.
- Task 4: a floor later than the fix commit itself is ignored, so the derived start never passes the derived end.
- Task 5: the before figure goes to `.superpowers/sdd/`, gitignored, never to the change's tracked directory while the suite's tree gate runs.

- [x] 1. Record the learning note against the main checkout
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `git -C /nonexistent/worktree rev-parse --git-common-dir; echo $?` — the call
    `fallback.ProjectKey` makes for a removed `-C` directory — and
    `grep -n 'flow learning add' skills/flow-contracts/learning.md`: git exits 128, and the
    contract's line ends `-C <worktree>`.
    <!-- measured: the two commands above @ a68fd756 → git exit 128; learning.md:52 ends -C <worktree> -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at a68fd756 together with the other four tasks' diffs, in task order; check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.sh and check-installed-citations.sh exit 0
diff --git a/skills/flow-contracts/learning.md b/skills/flow-contracts/learning.md
index ea9d6bb6..ccdd22bf 100644
--- a/skills/flow-contracts/learning.md
+++ b/skills/flow-contracts/learning.md
@@ -49,9 +49,13 @@ Recorded as the final message prints, never at composition, so a run that stops
 showing of the topic:
 
 ```bash
-flow learning add <topic> -note '<the topic in under 80 characters>' -change <name> -C <worktree>
+flow learning add <topic> -note '<the topic in under 80 characters>' -change <name> -C <project>
 ```
 
+`-C` names the main checkout, never the worktree: `/flow-fast`'s cleanup has removed its worktree
+before the final message prints, and the project key derives from the git common directory, the
+same from either.
+
 Exit 1 with stderr naming `already shown three times` or `rejected` is the store refusing the
 topic, and the run picks another topic. Any other non-zero exit — the store unreachable — prints the
 note unrecorded and never blocks.
````

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow-contracts/learning.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow-contracts): record the learning note against the main checkout
**After:** none

**Decision:** learning-add-project-dir

- [x] 2. Anchor the commit-per-task chain at the worktree root
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In a scratch repository
    outside the worktree holding a committed `stats/f`, modify `stats/f`, `cd stats`, and run the
    chain's unanchored `git add -- stats/f`: it exits 128 with `pathspec 'stats/f' did not match
    any files`; the same chain with `git -C <the scratch repo>` commits.
    <!-- measured: throwaway repo in the session scratchpad @ a68fd756 → unanchored exit 128, anchored exit 0 -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at a68fd756 together with the other four tasks' diffs, in task order; check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.sh, check-installed-citations.sh and check-dispatch-paragraphs.sh exit 0
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index d02423e1..4ce01c20 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -498,13 +500,15 @@ Every implementer dispatch **must** carry:
 > step already staged:
 >
 > ```bash
-> git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ \
->   && git add -- <this task's files> ':(exclude)spectre/changes/' ':(exclude)openspec/changes/' \
->   ':(exclude)docs/superpowers/' \
->   && git commit -m "<the task's declared subject>" -m "Task-Id: <n>" -- <this task's files>
+> git -C <abs-worktree> reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ \
+>   && git -C <abs-worktree> add -- <this task's files> ':(exclude)spectre/changes/' \
+>   ':(exclude)openspec/changes/' ':(exclude)docs/superpowers/' \
+>   && git -C <abs-worktree> commit -m "<the task's declared subject>" -m "Task-Id: <n>" -- <this task's files>
 > ```
 >
-> Paths are relative to the worktree root. Both spec-tree leaf spellings are named because the
+> Paths are relative to the worktree root, which `-C <abs-worktree>` anchors whatever directory
+> the shell is in — from a subdirectory of it, an unanchored `git add` reads them
+> against that directory and fails with a pathspec error. Both spec-tree leaf spellings are named because the
 > leaf is the project's — `spectre` or `openspec` — and a path the project does not use is a
 > reset that clears nothing and an exclude that matches nothing. **Never a bare
 > `git add -A`, `git add .`, or `git commit -a`**:
````

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0; `scripts/check-dispatch-paragraphs.sh` prints `DISPATCH-PARAGRAPHS-OK`.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow-contracts/git-boundaries.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** fix(flow): anchor the commit-per-task chain at the worktree root
**After:** none

**Decision:** commit-chain-anchored

- [x] 3. Tell a plan to carry the verbatim-moves step for run-loaded prose
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `scripts/project-get.sh . lint | grep -n 'verify step'`: it prints nothing — the section names
    `check-verbatim-moves.sh` as the guard but tells no plan to run it.
    <!-- measured: scripts/project-get.sh . lint | grep -n 'verify step' @ a68fd756 → no output, exit 1; grep -c 'verify step' .flow/project.md → 0 before, 2 with the diff applied -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at a68fd756 together with the other four tasks' diffs, in task order; check-verbatim-moves.sh, check-references.sh and check-markdown-integrity.sh exit 0
diff --git a/.flow/project.md b/.flow/project.md
index 61ffc540..a38873ce 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -166,7 +166,11 @@ sentences — 8 in the whole corpus, against hundreds of rules stated with "neve
 "only" — and a `MUST` moved into a `-rationale.md`, which no run loads, leaves its output unchanged.
 `check-verbatim-moves.sh` is the guard for that: it fails a run-loaded sentence that was deleted or
 reworded rather than moved verbatim, and new run-loaded prose, unless the change lists the sentence
-in its `verbatim-moves.txt`.
+in its `verbatim-moves.txt`. **A plan whose task edits run-loaded prose names
+`check-verbatim-moves.sh` in that task's verify step and commits the change's `verbatim-moves.txt`,
+listing every sentence its tasks add or reword, with the plan** — a `/flow` implementer never writes
+`spectre/changes/`, so a verify step naming only the other prose guards lets the FAIL reach the
+review panel (kan-1027's F1 and F8).
 
 **`check-workspace-isolation.sh` is a lint step where the other `## workspace isolation` guard is
 not.** It takes a project root, defaults to this repository when given none, and answers a question
````

  - [x] **Step 3: Verify.** `grep -c 'verify step' .flow/project.md` prints `2` (it prints `0`
    before the fix; `project-get.sh` resolves from HEAD while the file diverges, so it is run only
    after the commit, where `scripts/project-get.sh . lint | grep -c 'verify step'` prints `2`);
    `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — a prose rule in the project configuration.
**Regression:** none — no test is declared.
**Commit:** docs(flow-config): tell a plan to carry the verbatim-moves step for run-loaded prose
**After:** none

**Decision:** verbatim-moves-in-lint-prose

- [x] 4. Floor an inline panel-fix back-fill at its round's reviewer reports
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `flow record dispatches -change kan-1027-flow-learning-setting -C . | jq -c '.[] | select(.role=="reviewer" or .role=="panel-fix") | [.id,.key,.startedAt,.endedAt,.metrics]'`:
    row 7260 `panel-fix-1` starts `2026-10-23T12:19:34+03:00`, before `panel-0-primary+principles`
    begins (12:20:01) and ends (12:28:18), with `metrics` `{}`; and
    `git log -1 --format='%h %cI %s' cafe4c30~1` prints `70f3d736 2026-10-23T12:19:34+03:00 chore(spectre): plan kan-1027-flow-learning-setting — before review`,
    the commit the back-fill derived the start from.
    <!-- measured: the two commands above @ a68fd756 → as stated -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`) — the helper, its call, the
    derivation comment, the new test and the `implement.md` sentence:

````diff verified:applied with git apply to this worktree at a68fd756 together with the other four tasks' diffs, in task order; with it `cd stats && go test ./cmd/flow/ -count=1` passes and go vet is clean; with the record.go hunks reverted alone TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers fails on panel-fix-1 and panel-fix-1-2
diff --git a/stats/cmd/flow/record.go b/stats/cmd/flow/record.go
index d7b83534..c0da6459 100644
--- a/stats/cmd/flow/record.go
+++ b/stats/cmd/flow/record.go
@@ -1340,7 +1340,8 @@ func runRecordDispatchBackfill(ctx context.Context, args []string, stdout, stder
 	// An inline back-fill carrying -commit takes neither instant: both
 	// derive from the commits (kan-1003: a hand-written -started-at was a
 	// guess). ended is the commit's committer time, started the newest
-	// commit before it in any of the change's worktrees.
+	// commit before it in any of the change's worktrees, floored for a
+	// panel-fix row at its round's reviewer ends (panelFixFloor).
 	var derived []string
 	if *v["agent-id"] == "inline" && *v["commit"] != "" {
 		if *v["started-at"] != "" || *v["ended-at"] != "" {
@@ -1366,6 +1367,9 @@ func runRecordDispatchBackfill(ctx context.Context, args []string, stdout, stder
 			fmt.Fprintf(stderr, "flow: cannot derive the instants of commit %s (or the commit before it) from any of the change's worktrees\n", *v["commit"])
 			return 2
 		}
+		if floor, ok := panelFixFloor(ctx, f, projectKey, *v["role"], *v["key"]); ok && floor.After(startT) && !floor.After(endT) {
+			startT = floor
+		}
 		*v["started-at"] = startT.UTC().Format(time.RFC3339)
 		*v["ended-at"] = endT.UTC().Format(time.RFC3339)
 		derived = []string{"-started-at", *v["started-at"], "-ended-at", *v["ended-at"]}
@@ -3287,6 +3291,37 @@ func commitBefore(t time.Time, roots []string) (time.Time, bool) {
 	return best, found
 }
 
+// panelFixFloor is the latest end among the reviewer rows of the panel
+// round a panel-fix-<r>[-<chunk>] key fixes, round r-1: an inline fix
+// cannot begin before the reports it fixes exist (kan-1027's self-review:
+// panel-fix-1's derived start was the pre-review planning commit, eight
+// minutes before panel-0's reviewers reported). Any other role or key, a
+// failed record read, or a round with no ended reviewer row has no floor.
+func panelFixFloor(ctx context.Context, f recordIdentityFlags, projectKey, role, key string) (time.Time, bool) {
+	rest, ok := strings.CutPrefix(key, "panel-fix-")
+	if role != "panel-fix" || !ok {
+		return time.Time{}, false
+	}
+	digits, _, _ := strings.Cut(rest, "-")
+	r, err := strconv.Atoi(digits)
+	if err != nil || r < 1 {
+		return time.Time{}, false
+	}
+	run, err := callRecord(ctx, f.addr, f.timeout, func(ctx context.Context, cl *client.Client) (records.Run, error) {
+		return cl.GetRunRecord(ctx, projectKey, f.change)
+	})
+	if err != nil {
+		return time.Time{}, false
+	}
+	var floor time.Time
+	for _, d := range run.Dispatches {
+		if n, ok := records.PanelDispatchRound(d); ok && n == r-1 && d.EndedAt != nil && d.EndedAt.After(floor) {
+			floor = *d.EndedAt
+		}
+	}
+	return floor, !floor.IsZero()
+}
+
 // mutationRoots is the -C repository's top level plus every key of the
 // worktrees map in the change's state record.
 func mutationRoots(ctx context.Context, f recordIdentityFlags, projectKey string) []string {
diff --git a/stats/cmd/flow/record_test.go b/stats/cmd/flow/record_test.go
index 4d9ca3d7..fd750bae 100644
--- a/stats/cmd/flow/record_test.go
+++ b/stats/cmd/flow/record_test.go
@@ -5215,6 +5215,58 @@ func TestRecordDispatchBackfillDerivesInlineInstants(t *testing.T) {
 	}
 }
 
+// TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers pins the
+// kan-1027 self-review fix: an inline panel-fix-<r> back-fill never starts
+// before the latest end among round r-1's reviewer rows, however early the
+// commit before its fix commit is; a round with no ended reviewer row, and
+// any other role, keep the commit-derived start.
+func TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers(t *testing.T) {
+	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
+	repo := recordRepo(t, "kan-258")
+	isolatedStateRoot(t)
+	fixtureCommit(t, repo, "2026-10-23T09:19:34Z")
+	sha := fixtureCommit(t, repo, "2026-10-23T09:30:03Z")
+	record := `{"change":"kan-258","dispatches":[` +
+		`{"key":"panel-0-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-23T09:20:01Z","endedAt":"2026-10-23T09:28:18Z"},` +
+		`{"key":"panel-0-tests","role":"reviewer","slot":"tests","startedAt":"2026-10-23T09:20:01Z","endedAt":"2026-10-23T09:20:39Z"},` +
+		`{"key":"panel-1-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-23T09:30:32Z","endedAt":"2026-10-23T09:30:56Z"}]}`
+	var bodies []map[string]any
+	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
+		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
+			_, _ = w.Write([]byte(record))
+			return
+		}
+		if r.Method != http.MethodPost {
+			w.WriteHeader(http.StatusNotFound)
+			return
+		}
+		raw, _ := readAll(r)
+		var b map[string]any
+		_ = json.Unmarshal(raw, &b)
+		bodies = append(bodies, b)
+		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"panel-fix","model":"opus"}`))
+	}))
+	t.Cleanup(srv.Close)
+	for _, tc := range []struct{ role, key, want string }{
+		{"panel-fix", "panel-fix-1", "2026-10-23T09:28:18Z"},
+		{"panel-fix", "panel-fix-1-2", "2026-10-23T09:28:18Z"},
+		{"panel-fix", "panel-fix-3", "2026-10-23T09:19:34Z"},
+		{"implementer", "task-1-implementer", "2026-10-23T09:19:34Z"},
+	} {
+		bodies = nil
+		args := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
+			"-change", "kan-258", "-role", tc.role, "-model", "opus", "-key", tc.key,
+			"-session-token", "mf-record-panel-fix-floor", "-outcome", "completed", "-agent-id", "inline", "-commit", sha}
+		var stdout, stderr bytes.Buffer
+		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
+			t.Fatalf("%s: exit %d; stderr:\n%s", tc.key, code, stderr.String())
+		}
+		if len(bodies) != 2 || bodies[0]["startedAt"] != tc.want {
+			t.Errorf("%s: bodies = %v, want begin startedAt %s", tc.key, bodies, tc.want)
+		}
+	}
+}
+
 // TestRecordDispatchBeginRequiresEffort pins that an omitted -effort is
 // refused before the store is contacted, never recorded as "default" --
 // an explicit -effort default is still accepted.
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index d02423e1..4ce01c20 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -405,7 +405,9 @@ caller inputs.
 **A dispatch recorded only after it finished is a back-fill, and is one call** — `flow record dispatch backfill`, carrying `begin`'s flags and `end`'s together with the dispatch's real `-started-at` and `-ended-at`, both required — never a `begin` and `end` pair, whose instants would read the back-fill's own moment.
 An inline back-fill (`-agent-id inline`) carrying `-commit` takes neither instant:
 `backfill` derives `-ended-at` from that commit's committer time and `-started-at` from the newest
-commit before it in any of the change's worktrees, and refuses either flag beside it — a
+commit before it in any of the change's worktrees — for a `panel-fix-<r>` key never earlier than the
+last end among round `<r-1>`'s reviewer rows, whose reports the fix works from — and refuses either
+flag beside it — a
 hand-written instant there is a guess (kan-1003: `-started-at $(date -u -v-45M …)`).
 
 **`-model` is the model this dispatch was actually given — the group's own `model` from the
````

  - [x] **Step 3: Prove the test bites.** Revert the `stats/cmd/flow/record.go` hunks alone
    (`git diff -- stats/cmd/flow/record.go > <scratch>; git apply -R <scratch>`), run
    `cd stats && go test ./cmd/flow/ -run 'TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers' -count=1`
    and capture its failure on `panel-fix-1` and `panel-fix-1-2`, then `git apply <scratch>` and
    capture the pass.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./... && go test ./cmd/flow/ -run 'TestRecordDispatchBackfill' -count=1`
    passes and prints no file; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0.
  - [x] **Step 5: Commit** the three files with the subject below.
  - [x] **Step 6: Repair row 7260**, the row the defect stamped, from `stats/` with the
    persistent store named explicitly — never this worktree's own derived database:

```bash verified:spiked — go run ./cmd/flowd retime-dispatch and restamp-dispatch with no arguments print their usage @ a68fd756
cd stats && export FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable'
go run ./cmd/flowd retime-dispatch 7260 2026-10-23T09:28:18Z 2026-10-23T09:30:03Z cafe4c30a4be51f9b7438dfc181dc3d762008b82
go run ./cmd/flowd restamp-dispatch 7260 /Users/tweety53/.claude/projects/-Users-tweety53-Projects-agents/c7a0b1c0-5658-47a2-a9ff-542ee90feeef.jsonl
```

    <!-- unverified:restamp-dispatch's inline attribution over the retimed window yields a non-empty tokens figure — the implementer reads it back in Step 7 -->
  - [x] **Step 7: Read the repair back** — the `## Live check` of `design.md`: `startedAt`
    `2026-10-23T12:28:18+03:00`, `endedAt` `2026-10-23T12:30:03+03:00`, `metrics.tokens` present.
    A restamp that exits non-zero or leaves `metrics` empty is reported with its output, never
    retried blind.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/implement.md`
**Tests:** `TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers`
**Regression:** `TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers` — reverting the commit removes the floor, and the `panel-fix-1` and `panel-fix-1-2` cases read the commit-derived `09:19:34Z` where they expect `09:28:18Z`.
**Baseline:** before=128 after=129
<!-- measured: grep -cE '^func Test[A-Za-z0-9_]*\(t \*testing\.T\)' stats/cmd/flow/record_test.go @ a68fd756 -->
**Commit:** fix(stats): floor an inline panel-fix back-fill at its round's reviewer reports
**After:** Task 2

**Decision:** panel-fix-backfill-floor
**Decision:** panel-fix-1-row-repair

- [x] 5. Measure a live check's before figures during pass 1's panel
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n 'before figure' skills/flow/review-panel.md skills/flow/verify-and-handoff.md`: only
    `verify-and-handoff.md`'s **Live check** names one, and it runs after both lint and test
    groups have exited, after the panel.
    <!-- measured: the grep above @ a68fd756 → verify-and-handoff.md hits only -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at a68fd756 together with the other four tasks' diffs, in task order; check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.sh and check-installed-citations.sh exit 0
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index c68cc81b..3c1f8a99 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -435,6 +435,11 @@ runs.**, `skills/flow/implement.md`). A full-suite fix is applied only once the
 ends, never under the reading slots, so once the round's reports exist, every pass-1 role re-runs in the next round over the
 delta that commit adds, under **Panel re-runs** (`skills/flow/review-panel-fix-round.md`) — a
 role that raised nothing included; a suite that passes adds no commit and re-dispatches nothing.
+**A live check's before figures are measured in the same window**, after the suite: when
+`design.md`'s `## Live check` names something to exercise, the parent measures each before figure
+from the merge base the way **Live check** (`skills/flow/verify-and-handoff.md`) states and writes
+it to `<abs-worktree>/.superpowers/sdd/live-before.md` — never into the change's directory, a
+tracked path the suite's tree gate reads — since nothing the panel changes moves the merge base.
 
 The same plan-tree discipline rides every round, re-run rounds included: before the launches the
 parent makes the reviewer-dispatch planning commit (**Planning commits**,
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index cb703340..b38be96b 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -167,7 +167,10 @@ planning path the write-in-progress planning commit carries. **A before figure i
 detached worktree at the merge base** — `git -C <worktree> worktree add --detach
 <project>-worktrees/<name>-before <merge-base>`, the base's stack run from there, then stopped, and the worktree
 removed with `git worktree remove` once the figure is written — never by checking base sources out
-over the change worktree. Then run `check-live-verification.sh
+over the change worktree. A before figure already in
+`<abs-worktree>/.superpowers/sdd/live-before.md`, measured while pass 1's panel read
+(`skills/flow/review-panel.md`), is copied into the record and never measured again. Then run
+`check-live-verification.sh
 <changeRoot>/live-verification.md`; its header is canonical for the record it accepts. Exit 1 → the
 record carries no before figure, or states one was not measured: measure it from the merge base as
 above and rewrite the record, never proceed past it (kan-998). The check is never deferred past this stage:
````

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0.
  - [x] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): measure a live check's before figures during pass 1's panel
**After:** none

**Decision:** live-before-in-panel-window
