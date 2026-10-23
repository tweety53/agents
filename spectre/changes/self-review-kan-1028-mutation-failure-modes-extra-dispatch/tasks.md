# self-review-kan-1028-mutation-failure-modes-extra-dispatch

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the five non-`big` findings of `kan-1028-mutation-failure-modes-extra-dispatch`'s
self-review pass — a decided bundle keeping a dropped role, the status template's Learning
section, `/flow-settings` offering two always-dropped roles, colliding session tokens, and a
mid-run task leaving the decision's inputs stale.

**Architecture:** one sentence in `skills/flow/review-panel.md`; one template line in
`skills/flow-contracts/handoff-blocks.md`; `ValidReviewers` loses two ids behind migration
`0044`, with tests; a token-shape sentence in `skills/flow-contracts/pipeline.md` cited from the
two run skills; one paragraph in `skills/flow/implement.md` plus a corrected decision row for
kan-1028.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-23-self-review-kan-1028-mutation-failure-modes-extra-dispatch-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan, listing every run-loaded sentence Tasks 1, 2, 3, 4 and 5 add or reword. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-1028-mutation-failure-modes-extra-dispatch`, branch `spectre/kan-1028-mutation-failure-modes-extra-dispatch`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Each task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root; all five were applied together, in task order, and checked in this worktree at `d7933ad6`. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Every prose task's verify step runs `scripts/check-verbatim-moves.sh` beside the other prose guards; it judges `skills/`, `commands-claude/` and `rules/`.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container; Task 5's repair writes one decision row through `flow record decision` and stops nothing.

## Review Focus

- Task 1: a decided bundle left naming no role is not dispatched at all — never dispatched as an empty brief.
- Task 3: the migration strips only `failure-modes` and `mutation` and keeps the surviving ids' order.
- Task 3: a stored list holding only those two ids becomes `[]`, never `NULL`.
- Task 4: the example token in `pipeline.md` stays a literal `check-stage-mark-calls.sh` accepts.
- Task 5: the corrected row changes `inputs` alone — class, panel, rolls and pairs unchanged.

- [ ] 1. Drop a removed role from a decided panel's bundles
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n 'panel.dispatches' skills/flow/review-panel.md` and read the drop rule after
    `**The roster names only`: it drops the roles from the roster and records each, and no
    sentence removes them from a decided `panel.dispatches` bundle, which the dispatch paragraph
    after `write-panel-diff.sh` dispatches as recorded.
    <!-- measured: grep -n 'panel.dispatches' skills/flow/review-panel.md @ d7933ad6 → no hit in the drop rule -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at d7933ad6 together with the other four tasks' diffs, in task order; check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.sh and check-installed-citations.sh exit 0
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 43f8e87c..282109b3 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -158,7 +158,10 @@ overrides:
 **The roster names only `primary` and `principles`.** `failure-modes` and `mutation` run on
 **The always-on extra dispatches** below, so every resolved roster — a `default` panel's, and a decided one recorded before
 kan-1028 — drops any role an always-on extra dispatch carries, each recorded with `flow record pass -round 0 -note 'not
-dispatched — carried by an always-on extra dispatch: <slot>'`.
+dispatched — carried by an always-on extra dispatch: <slot>'`. A dropped role leaves its
+`panel.dispatches` bundle too: that bundle dispatches the roles it still names, and a bundle left
+naming none is not dispatched — so no role an always-on extra dispatch carries runs twice, and no
+two dispatches of one round create the same `<worktree>-mutation-<round>` copy.
 
 **A subagent-facing file is passed by absolute path, never read into this context.** This repository's
 `primary-reviewer-prompt.md` and `reviewer-calibration.md` (Primary), `principles-reviewer-prompt.md` and
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** fix(review-panel): drop a removed role from a decided panel's bundles
**After:** none

**Decision:** dropped-role-leaves-bundle

- [ ] 2. Print the Learning path in the status handoff template
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n 'Learning' skills/flow-contracts/handoff-blocks.md skills/flow-contracts/learning.md`:
    the template carries `## Learning — <topic>`, the contract `**Learning:** <absolute path>`.
    <!-- measured: the grep above @ d7933ad6 → handoff-blocks.md:89 section, learning.md:15-16 line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at d7933ad6 together with the other four tasks' diffs, in task order; check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.sh and check-installed-citations.sh exit 0
diff --git a/skills/flow-contracts/handoff-blocks.md b/skills/flow-contracts/handoff-blocks.md
index 3978cba2..e8ec5ad4 100644
--- a/skills/flow-contracts/handoff-blocks.md
+++ b/skills/flow-contracts/handoff-blocks.md
@@ -86,7 +86,7 @@ dropped rather than reported missing otherwise. The open-questions entry shape i
 **Change:** <name>
 **Summary:** (run-only) <what this round changed for a user of the app, then one bullet per line another contract requires in the handoff>
 **Decisions:** (run-only) <one line per open question the operator must answer and per prompt the run took on its recommended option, 🤖-marked, question and option taken, or "none">
-## Learning — <topic>  (run-only, when composed — **Learning**, `skills/flow-contracts/learning.md`)
+**Learning:** (run-only) <the learning note's absolute path, when composed — **Learning**, `skills/flow-contracts/learning.md`>
 **Staged:** <completed>/<total> tasks · <staged and uncommitted, committed and pushed to the PR branch, or committed and pushed with no PR — run 1 merged it or handed it over>
 **Records:** <all writes reached the store, "N write(s) journalled — the store was unreachable", or "unknown — the journal could not be counted">
 
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow-contracts/handoff-blocks.md`
**Tests:** **none** — a template line; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow-contracts): print the Learning path in the status handoff template
**After:** none

**Decision:** status-template-learning-line

- [ ] 3. Retire failure-modes and mutation from the reviewer settings
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n '"failure-modes"\|"mutation"' stats/internal/store/settings.go` — both ids sit in
    `ValidReviewers` — and `flow settings get`, which prints the stored list.
    <!-- measured: the grep and flow settings get @ d7933ad6 → settings.go:38-39; {"reviewers":["primary","principles"],"learning":true} -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`) — the vocabulary, the migration,
    the tests and the skill's example line:

````diff verified:applied with git apply to this worktree at d7933ad6 together with the other four tasks' diffs, in task order; with it `cd stats && go test ./internal/store/ ./internal/api/ ./cmd/flow/ -count=1` passes and go vet is clean; with settings.go reverted and the migration removed, TestSettingsStore_MigrationDropsExtraDispatchRoles and TestSettingsStore_RejectsExtraDispatchRoles fail
diff --git a/skills/flow-settings/SKILL.md b/skills/flow-settings/SKILL.md
index 347795df..a192a7e4 100644
--- a/skills/flow-settings/SKILL.md
+++ b/skills/flow-settings/SKILL.md
@@ -86,7 +86,7 @@ conductor depth, and when the type is absent print one line ahead of the write,
 affected slot in roster order:
 
 ```
-⚠ roster: no `flow-low` agent type at conductor depth — primary, principles, failure-modes will be substituted (general-purpose) at panel time
+⚠ roster: no `flow-low` agent type at conductor depth — primary, principles will be substituted (general-purpose) at panel time
 ```
 
 The write proceeds either way: never a gate. Never
diff --git a/stats/internal/api/settings_test.go b/stats/internal/api/settings_test.go
index f0319e37..98df3b7d 100644
--- a/stats/internal/api/settings_test.go
+++ b/stats/internal/api/settings_test.go
@@ -97,7 +97,7 @@ func TestSettingsAPI_Put_Valid(t *testing.T) {
 	ts := newSettingsTestServer(t, fs)
 	defer ts.Close()
 
-	body := `{"reviewers":["primary","principles","mutation","failure-modes"]}`
+	body := `{"reviewers":["primary","principles"]}`
 	req, err := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/settings", bytes.NewReader([]byte(body)))
 	if err != nil {
 		t.Fatalf("build request: %v", err)
@@ -115,7 +115,7 @@ func TestSettingsAPI_Put_Valid(t *testing.T) {
 	if err != nil {
 		t.Fatalf("GetSettings: %v", err)
 	}
-	want := []string{"primary", "principles", "mutation", "failure-modes"}
+	want := []string{"primary", "principles"}
 	if len(got.Reviewers) != len(want) {
 		t.Fatalf("stored reviewers = %v, want %v", got.Reviewers, want)
 	}
diff --git a/stats/internal/store/migrations/0044_flow_settings_drop_extra_dispatch_roles.sql b/stats/internal/store/migrations/0044_flow_settings_drop_extra_dispatch_roles.sql
new file mode 100644
index 00000000..cb64c1ff
--- /dev/null
+++ b/stats/internal/store/migrations/0044_flow_settings_drop_extra_dispatch_roles.sql
@@ -0,0 +1,14 @@
+-- 0044_flow_settings_drop_extra_dispatch_roles.sql: strip failure-modes and
+-- mutation from the stored reviewer list.
+--
+-- kan-1028 moved both roles onto the always-on mutation+failure-modes extra
+-- dispatch, which every pass 1 runs, and ValidReviewers (settings.go) no
+-- longer accepts either; a row still listing one would make /flow-settings'
+-- "keep current" send back a list PutSettings now refuses. Order of the
+-- surviving ids is kept, as 0033 keeps it.
+UPDATE flow_settings
+SET reviewers = COALESCE((
+  SELECT jsonb_agg(r ORDER BY ord)
+  FROM jsonb_array_elements_text(reviewers) WITH ORDINALITY AS t(r, ord)
+  WHERE r NOT IN ('failure-modes', 'mutation')
+), '[]'::jsonb);
diff --git a/stats/internal/store/settings.go b/stats/internal/store/settings.go
index c56bbd66..25c1efda 100644
--- a/stats/internal/store/settings.go
+++ b/stats/internal/store/settings.go
@@ -26,17 +26,17 @@ var ValidModels = map[string]bool{
 // "code-review-low" once it proved a shallower copy of primary's
 // code-review angle; "bugbot" and "security" once the flow store showed
 // security raising no finding and bugbot the lowest yield of any slot;
-// "failure-modes" was promoted from the experimental exp-failure-modes
-// prompt. The store rejects every retired id like any other unknown id.
-// Only writes are validated: the record/aggregate paths read historical
-// rows carrying a retired id unchanged, and a stored flow_settings list
-// had its retired ids stripped once by
-// 0033_flow_settings_drop_retired_reviewers.sql.
+// "failure-modes" and "mutation" once kan-1028 moved both onto the
+// always-on mutation+failure-modes extra dispatch, which every pass 1 runs
+// whatever this list holds. The store rejects every retired id like any
+// other unknown id. Only writes are validated: the record/aggregate paths
+// read historical rows carrying a retired id unchanged, and a stored
+// flow_settings list had its retired ids stripped once by
+// 0033_flow_settings_drop_retired_reviewers.sql and
+// 0044_flow_settings_drop_extra_dispatch_roles.sql.
 var ValidReviewers = map[string]bool{
-	"primary":       true,
-	"principles":    true,
-	"failure-modes": true,
-	"mutation":      true,
+	"primary":    true,
+	"principles": true,
 }
 
 // ErrInvalidReviewer is returned by PutSettings when a Reviewers entry is
diff --git a/stats/internal/store/settings_test.go b/stats/internal/store/settings_test.go
index 969a6ab0..5b2733f7 100644
--- a/stats/internal/store/settings_test.go
+++ b/stats/internal/store/settings_test.go
@@ -19,7 +19,7 @@ func TestSettingsStore_RoundTrip(t *testing.T) {
 	ctx := context.Background()
 
 	want := store.Settings{
-		Reviewers: []string{"primary", "principles", "failure-modes", "mutation"},
+		Reviewers: []string{"primary", "principles"},
 	}
 
 	if err := st.PutSettings(ctx, want); err != nil {
@@ -44,7 +44,7 @@ func TestSettingsStore_RoundTripUpdates(t *testing.T) {
 	ctx := context.Background()
 
 	first := store.Settings{Reviewers: []string{"primary"}}
-	second := store.Settings{Reviewers: []string{"primary", "failure-modes"}}
+	second := store.Settings{Reviewers: []string{"primary", "principles"}}
 
 	if err := st.PutSettings(ctx, first); err != nil {
 		t.Fatalf("first PutSettings: %v", err)
@@ -151,7 +151,7 @@ func TestSettingsStore_MigrationDropsRetiredReviewers(t *testing.T) {
 
 	if _, err := pool.Exec(ctx, `
 		INSERT INTO flow_settings (id, reviewers)
-		VALUES (TRUE, '["primary","bugbot","principles","security","code-review-low","simple-reviewer","mutation"]')
+		VALUES (TRUE, '["primary","bugbot","principles","security","code-review-low","simple-reviewer"]')
 		ON CONFLICT (id) DO UPDATE SET reviewers = EXCLUDED.reviewers
 	`); err != nil {
 		t.Fatalf("seed legacy row: %v", err)
@@ -167,7 +167,7 @@ func TestSettingsStore_MigrationDropsRetiredReviewers(t *testing.T) {
 	if err != nil {
 		t.Fatalf("GetSettings: %v", err)
 	}
-	want := []string{"primary", "principles", "mutation"}
+	want := []string{"primary", "principles"}
 	if !reflect.DeepEqual(got.Reviewers, want) {
 		t.Errorf("reviewers = %v, want %v", got.Reviewers, want)
 	}
@@ -176,6 +176,56 @@ func TestSettingsStore_MigrationDropsRetiredReviewers(t *testing.T) {
 	}
 }
 
+// TestSettingsStore_MigrationDropsExtraDispatchRoles asserts
+// 0044_flow_settings_drop_extra_dispatch_roles.sql strips failure-modes and
+// mutation from a flow_settings row written before kan-1028, keeping the
+// surviving ids' order, and that the migrated row is one PutSettings
+// accepts -- the list /flow-settings' "keep current" writes back.
+func TestSettingsStore_MigrationDropsExtraDispatchRoles(t *testing.T) {
+	st, pool := newRecordStore(t)
+	ctx := context.Background()
+
+	if _, err := pool.Exec(ctx, `
+		INSERT INTO flow_settings (id, reviewers)
+		VALUES (TRUE, '["primary","failure-modes","principles","mutation"]')
+		ON CONFLICT (id) DO UPDATE SET reviewers = EXCLUDED.reviewers
+	`); err != nil {
+		t.Fatalf("seed pre-kan-1028 row: %v", err)
+	}
+	if _, err := pool.Exec(ctx, `DELETE FROM schema_migrations WHERE filename = '0044_flow_settings_drop_extra_dispatch_roles.sql'`); err != nil {
+		t.Fatalf("unrecord migration: %v", err)
+	}
+	if err := st.RunMigrations(ctx); err != nil {
+		t.Fatalf("RunMigrations: %v", err)
+	}
+
+	got, err := st.GetSettings(ctx)
+	if err != nil {
+		t.Fatalf("GetSettings: %v", err)
+	}
+	if want := []string{"primary", "principles"}; !reflect.DeepEqual(got.Reviewers, want) {
+		t.Errorf("reviewers = %v, want %v", got.Reviewers, want)
+	}
+	if err := st.PutSettings(ctx, got); err != nil {
+		t.Errorf("PutSettings(migrated row) = %v, want it accepted", err)
+	}
+}
+
+// TestSettingsStore_RejectsExtraDispatchRoles asserts failure-modes and
+// mutation, carried by the always-on extra dispatch since kan-1028, are
+// refused on write like any unknown id.
+func TestSettingsStore_RejectsExtraDispatchRoles(t *testing.T) {
+	st := newTestStore(t)
+	ctx := context.Background()
+
+	for _, id := range []string{"failure-modes", "mutation"} {
+		err := st.PutSettings(ctx, store.Settings{Reviewers: []string{"primary", id}})
+		if !errors.Is(err, store.ErrInvalidReviewer) {
+			t.Errorf("PutSettings(%q) error = %v, want it to wrap ErrInvalidReviewer", id, err)
+		}
+	}
+}
+
 // TestSettingsStore_MigrationDropsModelColumns asserts
 // 0034_flow_settings_drop_model_columns.sql leaves flow_settings holding
 // id, reviewers, updated_at and 0043's learning alone, and that a row written before it -- both model
````

  - [ ] **Step 3: Prove the tests bite.** Save the `stats/internal/store/settings.go` hunk
    (`git diff -- stats/internal/store/settings.go > <scratch>`), `git apply -R <scratch>`, move
    the migration file to the scratch directory, run
    `cd stats && go test ./internal/store/ -run 'TestSettingsStore_(MigrationDropsExtraDispatchRoles|RejectsExtraDispatchRoles)' -count=1`
    and capture both failures, then `git apply <scratch>`, move the migration back, and capture
    the pass.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./... && go test ./internal/store/ ./internal/api/ ./cmd/flow/ -run 'TestSettings' -count=1`
    passes and prints no file; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0.
  - [ ] **Step 5: Commit** the five files with the subject below.

**Build:** green
**Files:** `stats/internal/store/settings.go`, `stats/internal/store/migrations/0044_flow_settings_drop_extra_dispatch_roles.sql`, `stats/internal/store/settings_test.go`, `stats/internal/api/settings_test.go`, `skills/flow-settings/SKILL.md`
**Tests:** `TestSettingsStore_MigrationDropsExtraDispatchRoles`, `TestSettingsStore_RejectsExtraDispatchRoles`
**Regression:** `TestSettingsStore_MigrationDropsExtraDispatchRoles` — reverting the commit removes the migration, and the seeded `failure-modes` and `mutation` survive; `TestSettingsStore_RejectsExtraDispatchRoles` — reverting restores both ids to `ValidReviewers`, and `PutSettings` accepts them.
**Baseline:** before=9 after=11
<!-- measured: grep -cE '^func Test[A-Za-z0-9_]*\(t \*testing\.T\)' stats/internal/store/settings_test.go @ d7933ad6 -->
**Commit:** fix(store): retire failure-modes and mutation from the reviewer settings
**After:** none

**Decision:** retire-extra-dispatch-reviewers
**Decision:** settings-row-no-repair

- [ ] 4. End every session token in a UTC timestamp
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n 'unique literal string' skills/flow/SKILL.md` and read **Stage marks** in
    `skills/flow-contracts/pipeline.md`: neither states a shape that keeps two runs' tokens apart,
    and kan-1028's archived `narrative.md` records `mf-k1028b` reused across sessions.
    <!-- measured: the grep above @ d7933ad6 → skills/flow/SKILL.md:149 -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at d7933ad6 together with the other four tasks' diffs, in task order; check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.sh and check-installed-citations.sh exit 0; check-stage-mark-calls.sh exits 0
diff --git a/skills/flow-contracts/pipeline.md b/skills/flow-contracts/pipeline.md
index c381c5b4..8f0d0e76 100644
--- a/skills/flow-contracts/pipeline.md
+++ b/skills/flow-contracts/pipeline.md
@@ -353,7 +353,10 @@ identifier is the **key**, never the prose name, from **Level 1 — the stages o
 
 **A `stage begin` mark MUST carry `-session-token` and `-harness`.** Generate the token once, near the start of
 the run, before the first `stage begin`, and reuse that exact value at every later mark site in the
-same run; do not invent a fresh one per mark. `-harness` names the harness actually running the mark
+same run; do not invent a fresh one per mark. **The token ends in the run's start time, UTC to the
+second** — read once with `date -u +%Y%m%dT%H%M%SZ` and typed as a literal, `mf-k1028-20261023T183402Z`
+— so no two runs share one: a name-derived token alone repeats across sessions (kan-1028:
+`mf-k1028b`, reused by a resumed run, collided with the dead session's rows). `-harness` names the harness actually running the mark
 — `claude-code`, `zcode` or `muse`.
 
 **Neither `-session-token` nor `-harness` is ever a hardcoded value in the skill text: both are
diff --git a/skills/flow-fast/SKILL.md b/skills/flow-fast/SKILL.md
index 9a26e9cc..fa3dd092 100644
--- a/skills/flow-fast/SKILL.md
+++ b/skills/flow-fast/SKILL.md
@@ -22,7 +22,8 @@ instructions on a re-run; report anything else rather than ignoring it. One exce
 `--base <branch>` on a creating run, as `/flow` takes it (`skills/flow/SKILL.md`).
 
 **Every mark is a literal call.** `<harness>` is this harness's name, `ff-<literal-token>` is one
-token generated once at the start of the run and typed the same at every `stage begin` and
+token generated once at the start of the run, shaped as **Stage marks**
+(`skills/flow-contracts/pipeline.md`) states, and typed the same at every `stage begin` and
 `stage mark` — never a shell substitution, per the CLI's own usage text. A `flow stage mark` line
 marks each `-stages` key begun and ended `completed`, in order: the stages `/flow-fast` has nothing
 to run for; the mark stays so the run's stage set matches `/flow`'s.
diff --git a/skills/flow/SKILL.md b/skills/flow/SKILL.md
index b4c4f59f..13815ce6 100644
--- a/skills/flow/SKILL.md
+++ b/skills/flow/SKILL.md
@@ -146,7 +146,8 @@ section has produced it, per **Change name resolution (all `/flow*` commands)**
 a guess or the best available name, which is legal for a read; it is never legal for a mark.
 
 **Generate this run's session token once, right here, before the first mark any phase file below
-makes — a short, unique literal string — and reuse that exact same value at every `stage begin` this
+makes — a short literal string shaped as **Stage marks** (`skills/flow-contracts/pipeline.md`)
+states — and reuse that exact same value at every `stage begin` this
 run makes, including inside every phase file it dispatches into.** One run, one token, never a fresh
 one per mark or per phase file. `<literal-token>` in each mark is that string written in place of
 the placeholder, never the placeholder itself, and a later `/flow` invocation — a resume or a fix
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0; `scripts/check-stage-mark-calls.sh` exits 0.
  - [ ] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `skills/flow-contracts/pipeline.md`, `skills/flow/SKILL.md`, `skills/flow-fast/SKILL.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** fix(flow-contracts): end every session token in a UTC timestamp
**After:** none

**Decision:** session-token-utc-suffix

- [ ] 5. Re-decide a task added mid-run, and correct kan-1028's decision row
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `flow record decisions -change kan-1028-mutation-failure-modes-extra-dispatch -C . | jq 'length, .[-1].decision.inputs'`
    and `scripts/plan-class.sh spectre/changes/archive/kan-1028-mutation-failure-modes-extra-dispatch/tasks.md 1 . 53b0dfcd | head -1`:
    one row reading `tasks 5, files 20`, against `inputs: tasks=6 files=23`.
    <!-- measured: the two commands above @ d7933ad6 → length 1, tasks 5 files 20; tasks=6 files=23 -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at d7933ad6 together with the other four tasks' diffs, in task order; check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.sh and check-installed-citations.sh exit 0
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index fbcc9697..fe56f811 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -834,6 +834,14 @@ superseded decision beside its replacement. The amended plan is
 re-validated — `spectre validate` and `check-plan-shape.sh` — before the next task dispatches; a
 change already archived runs `check-plan-shape.sh` alone, as section 1 does.
 
+**A task added after the decision row is re-decided before it runs.** Whenever `tasks.md` gains a
+task once the run's decision is recorded — a pivot, an operator's mid-run request, any append
+outside a fix run's, which re-decides on its own — the parent re-runs **Decide**
+(`skills/flow/brainstorm-planner.md`) and records the new row as **Documenting a fix, before implementing it** (`skills/flow/document-fix.md`) re-decides
+an appended plan, before the added task is implemented, so the newest row's `inputs` count the plan that
+ran (kan-1028: Task 6, added mid-run, left the only row at `tasks=5 files=20` against
+`tasks=6 files=23`).
+
 **The gated per-task reviewer.** One combined review per gate-fired task — spec compliance and
 code quality together — but **one dispatch per bundle of gate-fired tasks, never one per task**,
 the discipline **Bundled dispatch** (`skills/flow/review-panel.md`) applies to panel rounds.
````

  - [ ] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0.
  - [ ] **Step 4: Commit** the file with the subject below.
  - [ ] **Step 5: Repair the decision row.** Read the UTC time once with `date -u +%Y%m%dT%H%M%SZ`
    and type it into the token as a literal (`sr-k1028-<that time>`), then, from this worktree's
    root:

```bash verified:jq over .superpowers/sdd/decision.json @ d7933ad6 and flow decision render -file <that output> -session-model opus -reviewers primary,principles exit 0, printing "tasks 6 · files 23"
jq '.inputs.tasks = 6 | .inputs.files = 23' .superpowers/sdd/decision.json > <scratch>/decision.json
cp <scratch>/decision.json .superpowers/sdd/decision.json
flow record decision -change kan-1028-mutation-failure-modes-extra-dispatch -C . -session-token sr-k1028-<that time> -file .superpowers/sdd/decision.json
```

    <!-- unverified:flow record decision prints recorded: decision <id> against the live store — Step 6 reads it back -->
  - [ ] **Step 6: Read the repair back** — the `## Live check` of `design.md`: the Step 1 `jq`
    prints a length of two and the newest row's inputs `tasks 6, files 23`, class `regular`. A
    record that exits non-zero is reported with its output, never retried blind.

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** fix(flow): re-decide a task added after the decision row
**After:** none

**Decision:** mid-run-task-re-decided
**Decision:** k1028-decision-row-repair
