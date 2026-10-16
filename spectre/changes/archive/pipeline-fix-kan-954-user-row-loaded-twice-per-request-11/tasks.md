# pipeline-fix-kan-954-user-row-loaded-twice-per-request-11

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a run never reaches FINISHED with an untraced or unverified item: the self-review pass
traces every suspected defect to a verdict, and the landing chain refuses a report that still
carries an untraced finding line.

**Architecture:** one refusal, `lsrrUntraced`, in `stats/internal/guard/landselfreviewreport.go`,
run right after the start-of-run branch assert; prose in four skill files states the rule and the
remedy.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:`
  trailer; never write `spectre/changes/` (this change's `verbatim-moves.txt` is already
  committed with the plan).
- Each task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; both were applied to this worktree at 752f3a28 and checked there. A diff that does
  not apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing.

## Review Focus

- The refusal runs before `git add`, so exit 4 leaves the index and HEAD untouched; every
  existing `TestLandSelfReviewReport` subtest is unchanged and green.
- Only `- **[`-prefixed finding lines are read; a context bundle or report prose saying "not
  traced" elsewhere still lands.
- The prose states each rule once: step 2 canonical for tracing, step 7 for the refusal's remedy,
  integrate.md and flow-fast cite step 7.

- [x] 1. Refuse a self-review report carrying an untraced finding
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf11-task1.diff`, apply its test hunk alone
    (`git apply --include='stats/internal/guard/*_test.go' /tmp/pf11-task1.diff`) and run
    `cd stats && go test ./internal/guard/ -run 'TestLandSelfReviewReportRefusesUntracedFinding' -count=1`:
    it FAILS with `rc=0` and a commit `[main …] docs(self-review): kan-u self-review report` on
    stdout — the untraced report landed.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/*_test.go' /tmp/pf11-task1.diff`):

````diff verified:applied with git apply to this worktree at 752f3a28; the new test fails with the lsrrUntraced call disabled and passes with it; go test ./internal/guard/ -run TestLandSelfReviewReport -count=1 green; gofmt -l and go vet clean
diff --git a/scripts/land-self-review-report.sh b/scripts/land-self-review-report.sh
index 7f5ea101..d2761cd2 100755
--- a/scripts/land-self-review-report.sh
+++ b/scripts/land-self-review-report.sh
@@ -28,6 +28,10 @@
 #   3  foreign staged work — one LAND-FOREIGN-STAGED line naming it; the
 #      commit is refused and nothing is rolled back (the chain's own
 #      add/rm stay staged beside the foreign paths)
+#   4  an untraced finding — one LAND-UNTRACED-FINDING line per finding
+#      line saying "not traced" or "untraced"; checked right after the
+#      start-of-run branch assert, so nothing is added, committed, pulled
+#      or pushed
 #   otherwise git's own exit code, unmasked: a rejected push leaves the
 #   commit local and names itself on stderr; this script never retries it.
 #   A pull --rebase that stops on a conflict leaves <repo> mid-rebase with
diff --git a/stats/internal/guard/land_self_review_report_test.go b/stats/internal/guard/land_self_review_report_test.go
index 92710352..9a3fc3e6 100644
--- a/stats/internal/guard/land_self_review_report_test.go
+++ b/stats/internal/guard/land_self_review_report_test.go
@@ -547,3 +547,41 @@ exec %q "$@"
 		})
 	}
 }
+
+// A finding line whose cause was "not traced" is an open item: the report is
+// refused with exit 4 before anything is staged, and lands once traced.
+func TestLandSelfReviewReportRefusesUntracedFinding(t *testing.T) {
+	t.Parallel()
+	dir := t.TempDir()
+	repo := dir + "/repo"
+	var g fxGit
+	g.git("", "init", "-q", "-b", "main", repo)
+	for _, kv := range [][2]string{{"user.name", "land-test"}, {"user.email", "land-test@example.com"}, {"commit.gpgsign", "false"}} {
+		g.git(repo, "config", kv[0], kv[1])
+	}
+	g.git(repo, "commit", "-q", "--allow-empty", "-m", "base")
+	mkdir(t, repo+"/docs/self-review")
+	const rep = "docs/self-review/kan-u-self-review.md"
+	g.write(repo+"/"+rep, "## 5. Go app — `flow-stats-app`\n\n- **[flow-stats-app]** a row shows no tokens; cause in flowd recording, not traced — noted\n")
+	if g.err != nil {
+		t.Fatal(g.err)
+	}
+	base := g.git(repo, "rev-parse", "HEAD")
+	env := Env{Dir: dir, Getenv: os.Getenv, LookupEnv: hermeticGitLookup}
+	var o, e bytes.Buffer
+	rc := Registry["land-self-review-report"]([]string{repo, "main", "docs(self-review): kan-u self-review report", rep}, env, &o, &e)
+	want := "LAND-UNTRACED-FINDING: " + rep + ":3: - **[flow-stats-app]** a row shows no tokens; cause in flowd recording, not traced — noted — trace it to a verdict before landing; nothing added, committed, pulled or pushed\n"
+	if rc != 4 || e.String() != want || o.String() != "" {
+		t.Fatalf("rc=%d stdout=%q stderr=%q", rc, o.String(), e.String())
+	}
+	if head := g.git(repo, "rev-parse", "HEAD"); head != base || g.git(repo, "diff", "--cached", "--name-only") != "" {
+		t.Fatalf("the refusal moved HEAD or staged a path: %s", g.git(repo, "status", "--short"))
+	}
+
+	g.write(repo+"/"+rep, "## 5. Go app — `flow-stats-app`\n\n- **[flow-stats-app]** a row shows no tokens; disproved — the row was the aborted attempt's — noted\n")
+	o.Reset()
+	e.Reset()
+	if rc := Registry["land-self-review-report"]([]string{repo, "main", "docs(self-review): kan-u self-review report", rep}, env, &o, &e); rc != 0 {
+		t.Fatalf("a traced report did not land: rc=%d stderr=%q", rc, e.String())
+	}
+}
diff --git a/stats/internal/guard/landselfreviewreport.go b/stats/internal/guard/landselfreviewreport.go
index da830167..a7eeccd4 100644
--- a/stats/internal/guard/landselfreviewreport.go
+++ b/stats/internal/guard/landselfreviewreport.go
@@ -4,7 +4,9 @@ import (
 	"errors"
 	"fmt"
 	"io"
+	"os"
 	"os/exec"
+	"path/filepath"
 	"strings"
 )
 
@@ -87,6 +89,9 @@ func landSelfReviewReport(args []string, env Env, stdout, stderr io.Writer) int
 	if rc := assertBranch("— nothing added, committed, pulled or pushed"); rc != 0 {
 		return rc
 	}
+	if rc := lsrrUntraced(env, repo, addPath, stderr); rc != 0 {
+		return rc
+	}
 	if rc := lsrrStatus(g("add", "--", addPath).Run(), stderr); rc != 0 {
 		return rc
 	}
@@ -219,3 +224,31 @@ func lsrrWords(lines []string) string {
 	}
 	return b.String()
 }
+
+// lsrrUntraced refuses a report whose finding line says its cause was not
+// traced: an untraced finding is an open item, and a run whose report
+// carries one must not reach FINISHED (pipeline-fix-kan-954-…-11). Every
+// such line is named, and exit 4 leaves nothing added, committed, pulled or
+// pushed. A path it cannot read is left to `git add`, which names it.
+func lsrrUntraced(env Env, repo, addPath string, stderr io.Writer) int {
+	p := addPath
+	if !filepath.IsAbs(p) {
+		p = filepath.Join(repo, p)
+	}
+	if !filepath.IsAbs(p) {
+		p = filepath.Join(env.Dir, p)
+	}
+	b, err := os.ReadFile(p)
+	if err != nil {
+		return 0
+	}
+	rc := 0
+	for i, line := range strings.Split(string(b), "\n") {
+		l := strings.ToLower(line)
+		if strings.HasPrefix(line, "- **[") && (strings.Contains(l, "not traced") || strings.Contains(l, "untraced")) {
+			fmt.Fprintf(stderr, "LAND-UNTRACED-FINDING: %s:%d: %s — trace it to a verdict before landing; nothing added, committed, pulled or pushed\n", addPath, i+1, strings.TrimSpace(line))
+			rc = 4
+		}
+	}
+	return rc
+}
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestLandSelfReviewReport' -count=1`
    — all green, `gofmt -l` prints nothing.
  - [x] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/landselfreviewreport.go`, `stats/internal/guard/land_self_review_report_test.go`, `scripts/land-self-review-report.sh`
**Tests:** `TestLandSelfReviewReportRefusesUntracedFinding`
**Regression:** reverting the commit lands a report whose finding line says "not traced" —
`TestLandSelfReviewReportRefusesUntracedFinding` fails with `rc=0` and a commit on stdout.
**Baseline:** before=1 after=2
<!-- measured: grep -cE '^func Test[A-Z]' stats/internal/guard/land_self_review_report_test.go (before at 752f3a28; after with this task's diff applied) -->
**Commit:** fix(guard): refuse a self-review report carrying an untraced finding
**After:** none

**Decision:** gate-at-landing-chain
**Decision:** untraced-phrases

- [x] 2. State that a finished run leaves nothing untraced or still open
  - [x] **Step 1: Reproduce before touching anything; do not guess.** On the untouched tree,
    `grep -n 'traced to a verdict' skills/flow-self-review/SKILL.md` and
    `grep -n 'carry no open item' skills/flow/cleanup.md` print nothing: no rule makes the pass
    trace a finding, and the Finished handoff admits an open item.
  - [x] **Step 2: Implement.** Save the diff below to `/tmp/pf11-task2.diff` and apply it
    (`git apply /tmp/pf11-task2.diff`):

````diff verified:applied with git apply to this worktree at 752f3a28; scripts/check-verbatim-moves.sh then FAILs exactly the four sentences this change's verbatim-moves.txt lists, and passes with that file in place
diff --git a/skills/flow-fast/SKILL.md b/skills/flow-fast/SKILL.md
index 6b6eab64..11811464 100644
--- a/skills/flow-fast/SKILL.md
+++ b/skills/flow-fast/SKILL.md
@@ -327,4 +327,5 @@ commits' shas per item 5 of step 3 of `/flow-self-review`, in whichever invocati
 beyond the chain's own path refuses the commit (`LAND-FOREIGN-STAGED`) — clear the staging or land
 from a clean checkout, never around it.
+A `LAND-UNTRACED-FINDING` refusal is not a stop: resolve it per step 7 of `/flow-self-review` (`skills/flow-self-review/SKILL.md`) and land again.
 
 A product defect the pass met stops it at the end of step 2 of `/flow-self-review`, closes the mark below
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index 791a1e02..151c316d 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -46,4 +46,10 @@ counted blast radius and whether it is `big`. This session keeps the filing-and-
 store rows and the report.
 
+**Every suspected defect is traced to a verdict before the pass returns it.** Reproduced, it is
+fixable, `big` or not; disproved, it is `noted`, the evidence that disproved it in its line. "Not
+traced", "unverified" or a cause left unknown is never a classification: such a finding goes to
+step 3's planner as a non-`big` finding whose plan opens on the trace, never into the report or
+into the run's final report as an open item.
+
 The pass runs the seven-angle table over the bundle, every angle explicit —
 present-but-empty for an angle with no findings, never omitted:
@@ -364,4 +370,5 @@ line naming a sha gets no second row when `flow self-review findings -change <na
 lists that sha (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`). A store failure is one
 warning line and the pass continues — the report below is the durable record.
+A row is stated as recorded only once `flow self-review findings -change <name>` lists it: read that listing before step 8, record every row it lacks again, and fix a store that still refuses per **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`).
 
 ### 7. Write the report, delete the bundle, land it
@@ -386,4 +393,5 @@ it in `<project>/.flow/project.md`'s `## lint` section; fix any violation before
 unreadable report inside it, an internal coverage.sh call failing) — is not a violation**: report it and stop before the
 commit, never commit a report the guard could not read.
+`land-self-review-report.sh` refuses a report whose finding line says `not traced` or `untraced` — `LAND-UNTRACED-FINDING`, exit 4, nothing committed: trace that finding to a verdict per step 2 (`skills/flow-self-review/SKILL.md`), rewrite its line, and land again.
 Commit both paths in one commit through the landing chain's one script — the same invocation
 `skills/flow/integrate.md` step 4 lands the report with, differing in the asserted branch,
diff --git a/skills/flow/cleanup.md b/skills/flow/cleanup.md
index 90e3ce1e..f9aa1fbe 100644
--- a/skills/flow/cleanup.md
+++ b/skills/flow/cleanup.md
@@ -61,4 +61,9 @@ merged** (`skills/flow-contracts/finish-contract-run2.md`).
 A run 2 that **completes step 5** is terminal and names **no** next command.
 
+**The Finished handoff, and the run's final report beside it, carry no open item** — nothing
+"still open", unverified or untraced. Each one is resolved before the handoff prints: fixed,
+disproved with its evidence, or a `big` finding filed per step 4 of `/flow-self-review`
+(`skills/flow-self-review/SKILL.md`).
+
 On a leftover — or on no verdict at all — the run stops at step 4 instead:
 
diff --git a/skills/flow/integrate.md b/skills/flow/integrate.md
index a5cbaccd..9162d864 100644
--- a/skills/flow/integrate.md
+++ b/skills/flow/integrate.md
@@ -148,4 +148,5 @@ A branch mismatch or a commit that FAILS is reported and stops the run at `IN_PR
 index holding anything beyond the chain's own path refuses the commit the same way
 (`LAND-FOREIGN-STAGED`).
+A `LAND-UNTRACED-FINDING` refusal is not a stop: resolve it per step 7 of `/flow-self-review` (`skills/flow-self-review/SKILL.md`) and land again.
 
 A product defect the pass met stops it at the end of step 2 of `/flow-self-review`, closes that mark
````

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh && scripts/check-markdown-integrity.py && scripts/check-references.sh`
    — all exit 0 (`verbatim-moves.txt` already lists the four new sentences).
  - [x] **Step 4: Commit** the four files with the subject below.

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`, `skills/flow/cleanup.md`, `skills/flow/integrate.md`, `skills/flow-fast/SKILL.md`
**Tests:** **none** — prose; Task 1's test pins the mechanical half, and `check-verbatim-moves.sh` fails on these sentences without this change's acknowledgement file.
**Regression:** reverting the commit removes the tracing rule, the store re-read, the refusal's remedy and the Finished no-open-item paragraph; `grep -n 'traced to a verdict' skills/flow-self-review/SKILL.md` prints nothing again.
**Commit:** docs(skills): require every finding traced before a run finishes
**After:** Task 1

**Decision:** trace-via-planner
**Decision:** store-rows-reread
**Decision:** finished-no-open-items
