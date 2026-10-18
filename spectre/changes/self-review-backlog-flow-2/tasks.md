# self-review-backlog-flow-2

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close the live Backlog Flow findings — KAN-558/KAN-565's uncited-frame gap, KAN-569's
absence sweeps, KAN-570's pre-existing-failure isolation, KAN-603's capture-scope report.

**Architecture:** one new `flow-guard check-mockup-citations` subcommand behind a
`flow_guard_exec` shim, run by writing plans (section D of `skills/flow/brainstorm-planner.md`),
and three prose additions where their readers already look.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-18-self-review-backlog-flow-2-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-2`, branch `self-review-backlog-flow-2`; never touch a main checkout.
- Each task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root — applied and checked in order in a scratch clone at `7d7a1b40` (`design.md`, **Context**). Save it by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- `spectre/changes/self-review-backlog-flow-2/verbatim-moves.txt`, committed with this plan, acknowledges every sentence the four diffs add or reword, so `scripts/check-verbatim-moves.sh` passes once each task lands.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`.

## Review Focus

- `check-mockup-citations` must not read `other/<mockups root>/X.png` or a bare `X.png` as a citation — case "bare ids and other roots are not citations".
- A directory named like a frame is not a frame — case "a directory named like a frame is no file".
- The planner paragraph must not make the guard run when the project declares no `mockups`.
- The Blocking item must bind only views no frame was composed for; a composed frame's sweeps stay step 10's.

- [ ] 1. Refuse a plan citing a mockup frame no file in the mockups root holds (KAN-558, KAN-565)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `ls scripts/check-mockup-citations.sh`
    fails — nothing checks a plan's cited frames against the mockups root. Save Step 2's diff,
    apply its test hunk alone
    (`git apply --include='stats/internal/guard/check_mockup_citations_test.go' <the saved diff>`)
    and run `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckMockupCitations' -count=1`:
    it fails with `rc=-1 out="check-mockup-citations is not registered"`.
    <!-- measured: both commands in a scratch clone @ 7d7a1b40 → ls exit 1; --- FAIL: TestCheckMockupCitations, "check-mockup-citations is not registered" -->
  - [ ] **Step 2: Apply the fix.** Apply the rest
    (`git apply --exclude='stats/internal/guard/check_mockup_citations_test.go' <the saved diff>`);
    it creates `scripts/check-mockup-citations.sh` executable and
    `skills/flow/scripts/check-mockup-citations.sh` as a symlink to
    `../../../scripts/check-mockup-citations.sh`:

````diff verified:applied with git apply to a scratch clone at 7d7a1b40; TestCheckMockupCitations passes all four table cases and the usage cases, and fails with the file-existence check disabled; gofmt -l and go vet ./internal/guard clean; go test ./internal/guard/ and scripts/run-guard-tests.sh pass; check-references.sh, check-guard-symlinks.sh, check-installed-citations.sh and check-verbatim-moves.sh 7d7a1b40 exit 0
diff --git a/.flow/project.md b/.flow/project.md
index 47897120..cbf7be6f 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -14,7 +14,7 @@ halves live in the one repo and are covered below.
 | flow sources | `/Users/tweety53/Projects/agents` | Bash + Python + Go + Markdown | — | The skills/commands/rules half. Verification is the guard scripts below plus a sandboxed `setup.sh` run. |
 | flow stats daemon | `/Users/tweety53/Projects/agents/stats` | Go + React/Vite | `http://127.0.0.1:4173` | `flowd`, loopback-only. Backed by a dedicated `flow-postgres` container on host port 5433, independent of any other Postgres stack on this machine. Also the one application `## apps` names that a fix run's reload rule (`skills/flow/verify-and-handoff.md`) never reloads — the same protection, not a separate one. |
 
-**This repository is Bash + Python + Go, not Bash-only.** Several `scripts/*.py` guards (`check-task-build-green.py`, `check-plan-shape.py`, `check-task-commit-fields.py`, `check-markdown-integrity.py`, among others) run on Python 3, standard library only — `/usr/bin/python3`, no third-party imports, no pip, no network. The guards ported into `flow-guard` — among them `check-plan-provenance`, `check-installed-citations`, `check-references`, `check-unfinished-work`, `check-panel-reproducers`, `check-panel-findings-closed`, `check-stage-mark-calls`, `check-guard-symlinks`, `check-dispatch-paragraphs`, `mutate-and-verify`, `check-base-moved`, `check-panel-fix-single-dispatch`, `prove-reproducer`, `check-finish-preflight`, `check-workspace-isolation`, `check-task-reviewer-single-dispatch`, `recover-guard-incident`, `plan-class`, `check-installed-rules`, `resolve-base-branch`, `check-visual-verify-dispatched`, `check-task-commit-planning-paths`, `check-panel-citation-trigger`, `check-visual-trigger`, `check-visual-verification`, `check-verify-report`, `check-frame-list`, `resolve-visual-screenshots`, `compose-mockup-frames`, `measure-visual-properties`, `commit-split`, `reshape-branch`, `check-foreign-staged`, `check-planning-commit-location`, `check-archive-scope`, `check-main-checkout-drift`, `refresh-main-checkout`, `land-self-review-report`, `check-self-review-report`, `break-and-prove`, `prepare-workspace` and `check-panel-docs-only` — are Go (`stats/internal/guard/`), and their `scripts/<name>.sh` is a `flow_guard_exec` shim over the binary built from this checkout.
+**This repository is Bash + Python + Go, not Bash-only.** Several `scripts/*.py` guards (`check-task-build-green.py`, `check-plan-shape.py`, `check-task-commit-fields.py`, `check-markdown-integrity.py`, among others) run on Python 3, standard library only — `/usr/bin/python3`, no third-party imports, no pip, no network. The guards ported into `flow-guard` — among them `check-plan-provenance`, `check-installed-citations`, `check-references`, `check-unfinished-work`, `check-panel-reproducers`, `check-panel-findings-closed`, `check-stage-mark-calls`, `check-guard-symlinks`, `check-dispatch-paragraphs`, `mutate-and-verify`, `check-base-moved`, `check-panel-fix-single-dispatch`, `prove-reproducer`, `check-finish-preflight`, `check-workspace-isolation`, `check-task-reviewer-single-dispatch`, `recover-guard-incident`, `plan-class`, `check-installed-rules`, `resolve-base-branch`, `check-visual-verify-dispatched`, `check-task-commit-planning-paths`, `check-panel-citation-trigger`, `check-visual-trigger`, `check-visual-verification`, `check-verify-report`, `check-frame-list`, `check-mockup-citations`, `resolve-visual-screenshots`, `compose-mockup-frames`, `measure-visual-properties`, `commit-split`, `reshape-branch`, `check-foreign-staged`, `check-planning-commit-location`, `check-archive-scope`, `check-main-checkout-drift`, `refresh-main-checkout`, `land-self-review-report`, `check-self-review-report`, `break-and-prove`, `prepare-workspace` and `check-panel-docs-only` — are Go (`stats/internal/guard/`), and their `scripts/<name>.sh` is a `flow_guard_exec` shim over the binary built from this checkout.
 
 **Every new guard, and any new logic added to an existing guard, is written in Go in `flow-guard`** — `stats/internal/guard/`, registered in `guard.Registry`, with Go table tests — and its `scripts/<name>.sh` is only a thin `flow_guard_exec` shim onto it. A guard still in bash is ported to Go, byte-for-byte parity, before it is extended; it is never grown in bash. A new `check-*.sh` lands with its `test-check-*.sh` companion in the same change — `run-guard-tests.sh` refuses the suite before anything runs while the companion is missing, a Go test in `stats/internal/guard/<name>_test.go` standing in for it (kan-903).
 
diff --git a/scripts/check-mockup-citations.sh b/scripts/check-mockup-citations.sh
new file mode 100755
index 00000000..558f08d3
--- /dev/null
+++ b/scripts/check-mockup-citations.sh
@@ -0,0 +1,40 @@
+#!/usr/bin/env bash
+# check-mockup-citations.sh — refuse a plan citing a mockup frame no file in
+# the project's mockups root holds (KAN-558, KAN-565: kan-469's design.md and
+# tasks.md cited P6.png, no such frame existed, and the gap surfaced only when
+# visual verification skipped the comparison).
+#
+# Usage: check-mockup-citations.sh <project root> <mockups root> <file>...
+#
+# <project root> is the directory `.flow/project.md` lives in (a worktree's
+# root); <mockups root> the `mockups` value of its `## visual verification`
+# table, relative to <project root>, a trailing `/` ignored
+# (skills/flow-contracts/project-configuration-visual.md); each <file> a plan
+# artifact — <changeRoot>/design.md and <changeRoot>/tasks.md.
+#
+# A CITATION is a `<mockups root>/<frame id>.png` token, <frame id> made of
+# letters, digits, `_` and `-`, not preceded by a letter, digit, `_`, `.`,
+# `/` or `-` — so `other/<mockups root>/X.png` is no citation. A bare
+# `<frame id>.png` is not one either: plans name capture screenshots that
+# way too, and nothing in the token tells the two apart. Writing plans
+# (skills/flow/brainstorm-planner.md, section D) has a plan cite each frame
+# by its path for that reason.
+#
+# Prints ONE verdict line to stdout:
+#   MOCKUP-CITATIONS-RESOLVED: <n> cited
+#   MOCKUP-CITATIONS-MISSING: <file>:<line> <citation>; …
+#
+# Exit 0 every citation names a regular file, 1 at least one does not,
+# 2 cannot answer (a usage error, a <mockups root> that is not a directory
+# under <project root>, an unreadable <file>).
+#
+# Go, in flow-guard (stats/internal/guard/mockupcitations.go).
+# flow-guard is built from this checkout, never taken from PATH:
+# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
+# cannot-answer code) with the cause when it cannot.
+set -euo pipefail
+[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
+  echo "check-mockup-citations: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
+  exit 2
+}
+flow_guard_exec check-mockup-citations 2 "check-mockup-citations:" "$@"
diff --git a/skills/flow/SKILL-rationale.md b/skills/flow/SKILL-rationale.md
index 00d6aaa6..78fb8c70 100644
--- a/skills/flow/SKILL-rationale.md
+++ b/skills/flow/SKILL-rationale.md
@@ -164,6 +164,7 @@ text after it is the passage, verbatim.
 
 - *The refusal exits are never repaired by editing the plan* — (KAN-601)
 - *…record its size as the change's first task-count observation — the planned figure every later fix round's growth is measured against* — (KAN-415)
+- *When the project declares `mockups`, also run `check-mockup-citations.sh`* — (KAN-558, KAN-565) kan-469's `design.md` and `tasks.md` cited `P6.png`, which no file in the mockups root held, and the gap surfaced only when visual verification skipped the comparison; kan-30 had planned fidelity specs believing no mockups existed. The guard reads only the path form because a bare `<frame id>.png` cannot be told from a capture screenshot's name: gymie's archived plans carried 178 bare `.png` tokens naming no mockup file, nearly all screenshots, against 100 path-form citations that all resolved.
 
 ### brainstorm-planner.md — D. Writing plans (live check)
 
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index bf076960..8c2acfef 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -294,7 +294,10 @@ paragraph under **The build-green tag** (`skills/flow-contracts/build-green.md`)
 > plus its `<spec>.mockups` sidecar line (**Visual verification**,
 > `skills/flow-contracts/project-configuration-visual.md`). Before finishing, list the frame ids
 > and name the task covering each; a frame no task covers is a plan gap, closed by adding the
-> capture to the task that builds that view or by a capture task of its own.
+> capture to the task that builds that view or by a capture task of its own. **Cite each frame by
+> its project-relative path, the `mockups` root joined to `<frame id>.png`, never as a bare
+> `<frame id>.png`** — a capture screenshot's name takes that shape too, so only the path form is
+> a citation `check-mockup-citations.sh` can resolve.
 
 > **Write a live check when the change touches a running service or persistent state — never a
 > live-verification task.** When the change under plan touches a long-running service, a daemon,
@@ -466,6 +469,13 @@ are reported and stop the plan. Provenance's exit 4 (content-classification) is
 whose remedy is editing the plan — fix the line it names and run the guard again, its own
 prescribed course, never accepting the plan while it stands.
 
+When the project declares `mockups`, also run
+`check-mockup-citations.sh <worktree> <mockups root> <changeRoot>/design.md <changeRoot>/tasks.md`
+before continuing. Its exit 1 names a cited frame no file in the mockups root holds: a handoff asset
+**The checklist** above still owes, never a frame to plan against — commit the asset or record its
+location, or drop the citation and every task that rests on it. Its exit 2 is reported and stops
+the plan.
+
 With the plan validated, record its size as the change's first task-count observation — the
 planned figure every later fix round's growth is measured against:
 
diff --git a/skills/flow/scripts/check-mockup-citations.sh b/skills/flow/scripts/check-mockup-citations.sh
new file mode 120000
index 00000000..91f1d3e7
--- /dev/null
+++ b/skills/flow/scripts/check-mockup-citations.sh
@@ -0,0 +1 @@
+../../../scripts/check-mockup-citations.sh
\ No newline at end of file
diff --git a/stats/internal/guard/check_mockup_citations_test.go b/stats/internal/guard/check_mockup_citations_test.go
new file mode 100644
index 00000000..fbb08f37
--- /dev/null
+++ b/stats/internal/guard/check_mockup_citations_test.go
@@ -0,0 +1,64 @@
+package guard
+
+import (
+	"os"
+	"path/filepath"
+	"strings"
+	"testing"
+)
+
+// Every case of the check-mockup-citations contract
+// (scripts/check-mockup-citations.sh): a `<mockups root>/<frame id>.png`
+// token resolves to a file under the project root, or it is missing.
+func TestCheckMockupCitations(t *testing.T) {
+	cases := []struct {
+		name string
+		plan string
+		rc   int
+		want string
+	}{
+		{"every cited frame exists",
+			"Match `docs/mockups/P1.png` and docs/mockups/P1.png again.\n",
+			0, "MOCKUP-CITATIONS-RESOLVED: 2 cited"},
+		// KAN-558: kan-469's plan cited P6.png, which no file in the mockups root held.
+		{"a cited frame with no file",
+			"Match `docs/mockups/P1.png`.\nThe row wraps per `docs/mockups/P6.png`.\n",
+			1, "plan.md:2 docs/mockups/P6.png"},
+		{"bare ids and other roots are not citations",
+			"Per `P6.png`, see other/docs/mockups/P6.png and snapshots/P6.png.\n",
+			0, "0 cited"},
+		{"a directory named like a frame is no file",
+			"See docs/mockups/dir.png.\n",
+			1, "docs/mockups/dir.png"},
+	}
+	for _, tc := range cases {
+		t.Run(tc.name, func(t *testing.T) {
+			dir := t.TempDir()
+			mockups := filepath.Join(dir, "docs", "mockups")
+			if err := os.MkdirAll(filepath.Join(mockups, "dir.png"), 0o755); err != nil {
+				t.Fatal(err)
+			}
+			if err := os.WriteFile(filepath.Join(mockups, "P1.png"), nil, 0o644); err != nil {
+				t.Fatal(err)
+			}
+			plan := filepath.Join(dir, "plan.md")
+			if err := os.WriteFile(plan, []byte(tc.plan), 0o644); err != nil {
+				t.Fatal(err)
+			}
+			r := runGuard("check-mockup-citations", []string{dir, "docs/mockups/", plan}, Env{Getenv: os.Getenv})
+			if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
+				t.Fatalf("rc=%d out=%q; want rc=%d containing %q", r.rc, r.out, tc.rc, tc.want)
+			}
+		})
+	}
+	dir := t.TempDir()
+	plan := filepath.Join(dir, "plan.md")
+	if err := os.WriteFile(plan, nil, 0o644); err != nil {
+		t.Fatal(err)
+	}
+	for _, args := range [][]string{nil, {dir, "."}, {dir, "absent", plan}, {dir, ".", filepath.Join(dir, "absent.md")}} {
+		if r := runGuard("check-mockup-citations", args, Env{Getenv: os.Getenv}); r.rc != 2 {
+			t.Fatalf("args %q: rc=%d out=%q; want exit 2", args, r.rc, r.out)
+		}
+	}
+}
diff --git a/stats/internal/guard/mockupcitations.go b/stats/internal/guard/mockupcitations.go
new file mode 100644
index 00000000..740c5f07
--- /dev/null
+++ b/stats/internal/guard/mockupcitations.go
@@ -0,0 +1,52 @@
+package guard
+
+import (
+	"fmt"
+	"io"
+	"os"
+	"path/filepath"
+	"regexp"
+	"strings"
+)
+
+// checkMockupCitations is scripts/check-mockup-citations.sh: that script's
+// header comment is the contract. Every `<mockups root>/<frame id>.png` token
+// a plan file carries must name a file under <project root>.
+func init() { Registry["check-mockup-citations"] = checkMockupCitations }
+
+func checkMockupCitations(args []string, env Env, stdout, stderr io.Writer) int {
+	const prog = "check-mockup-citations"
+	if len(args) < 3 {
+		fmt.Fprintf(stderr, "%s: usage: check-mockup-citations.sh <project root> <mockups root> <file>...\n", prog)
+		return 2
+	}
+	root := strings.TrimSuffix(args[1], "/")
+	if fi, err := os.Stat(filepath.Join(args[0], root)); err != nil || !fi.IsDir() {
+		fmt.Fprintf(stderr, "%s: mockups root %s is not a directory under %s\n", prog, root, args[0])
+		return 2
+	}
+	cite := regexp.MustCompile(`(^|[^A-Za-z0-9_./-])(` + regexp.QuoteMeta(root) + `/[A-Za-z0-9_-]+\.png)`)
+	cited := 0
+	var missing []string
+	for _, file := range args[2:] {
+		data, err := os.ReadFile(file)
+		if err != nil {
+			fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, file, err)
+			return 2
+		}
+		for i, line := range strings.Split(string(data), "\n") {
+			for _, m := range cite.FindAllStringSubmatch(line, -1) {
+				cited++
+				if fi, err := os.Stat(filepath.Join(args[0], m[2])); err != nil || !fi.Mode().IsRegular() {
+					missing = append(missing, fmt.Sprintf("%s:%d %s", file, i+1, m[2]))
+				}
+			}
+		}
+	}
+	if len(missing) > 0 {
+		fmt.Fprintf(stdout, "MOCKUP-CITATIONS-MISSING: %s\n", strings.Join(missing, "; "))
+		return 1
+	}
+	fmt.Fprintf(stdout, "MOCKUP-CITATIONS-RESOLVED: %d cited\n", cited)
+	return 0
+}
````

  - [ ] **Step 3: Verify.** `cd stats && make -s web-dist && go test ./internal/guard -run 'TestCheckMockupCitations' -count=1`
    passes; `cd stats && go vet ./internal/guard && gofmt -l internal/guard` prints nothing;
    `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 4: Commit** the seven files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/mockupcitations.go`, `stats/internal/guard/check_mockup_citations_test.go`, `scripts/check-mockup-citations.sh`, `skills/flow/scripts/check-mockup-citations.sh`, `skills/flow/brainstorm-planner.md`, `skills/flow/SKILL-rationale.md`, `.flow/project.md`
**Tests:** `TestCheckMockupCitations`
**Regression:** `TestCheckMockupCitations` — reverting the commit leaves `check-mockup-citations` unregistered: every case fails with `check-mockup-citations is not registered`, and a plan citing a frame no file holds is accepted again.
**Commit:** feat(guard): refuse a plan citing a mockup frame no file in the mockups root holds
**After:** none

**Decision:** mockup-citations-guard

- [ ] 2. End a removal task on an absence sweep (KAN-569)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -rn "absence sweep" skills; echo $?`: it prints only `1` — no planning rule names the sweep.
    <!-- measured: the grep above in this worktree @ 7d7a1b40 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 7d7a1b40 after Task 1's diff; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-verbatim-moves.sh 7d7a1b40 exit 0
diff --git a/skills/flow/SKILL-rationale.md b/skills/flow/SKILL-rationale.md
index 78fb8c70..93057888 100644
--- a/skills/flow/SKILL-rationale.md
+++ b/skills/flow/SKILL-rationale.md
@@ -156,6 +156,7 @@ text after it is the passage, verbatim.
 - ***Triage the sweep's findings before any fix runs.*** — (KAN-800; kan-741's whole-suite verification found 19 defects and none piled into a mega-fix: six became tasks 7–12 in the creating run, thirteen more tasks 13–25 in the fix run, each with its own field family, while the pre-existing out-of-scope failures went to KNOWN-BUGS.md by a dedicated triage dispatch)
 - ***A task premise about a guard's behaviour is run at plan time, never assumed.*** — kan-542's task 2 shows the cost of leaving the premise to review: planned on a stale premise, it survived until a reviewer struck it with the guard's own exit — a defect this duty kills at planning.
 - ***A task cites the decision it implements, never restates it.** Restated decision prose drifts from its entry the first time either is edited* — which is why gymie kan-468's seeded plan cited instead
+- ***End a removal task on an absence sweep.*** — (KAN-569) every kan-485 copy-removal task ended on a `git grep -n -F` over its cut strings in the shared sources, so each task's done was mechanical rather than a judgment call.
 - *`**Baseline:**` … record a command whose stdout is one integer* — (the `| grep -c` pipelines in kan-271 and kan-298's plans are the shape)
 - *`**Files:**` … is expanded by the writer, never left in the field for the guard to expand* — (KAN-636); kan-579's plan declared gs/core/... behind its own legend and hand-repaired every field back to literal paths at its first task boundary, which is the repair this rule removes.
 - *A block of `**Decision:**` lines sits between blank lines … never glued to the `**Commit:**` subject above it* — (KAN-636): a citation with no blank line rode in the subject's continuation in every reader that had not learned the field's name, and kan-579's run repaired each such field by hand at its task boundary.
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index 8c2acfef..a4a98ce7 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -332,6 +332,11 @@ paragraph under **The build-green tag** (`skills/flow-contracts/build-green.md`)
 > code change, while an edit made before reproducing destroys the evidence the task set out to
 > gather.
 
+> **End a removal task on an absence sweep.** When a task's job is removing text — copy, a setting,
+> an identifier — its last verify step is one `git grep -n -F` over every string it removes, each
+> passed as its own `-e '<string>'`, scoped to the paths the task touches: the sweep prints
+> nothing when the task is done, and any hit means it is not.
+
 > **List the rationale appendix a changed rule's reasoning lives in.** A step that changes a rule
 > whose reasoning a `*-rationale.md` appendix records lists that appendix in its task's
 > `**Files:**` and updates it.
````

  - [ ] **Step 3: Verify.** `grep -c "absence sweep" skills/flow/brainstorm-planner.md skills/flow/SKILL-rationale.md`
    prints `1` for each file; `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 7d7a1b40 with Task 1 applied → 1 and 1 -->
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/SKILL-rationale.md`
**Tests:** **none** — a prose rule; `check-references.sh` and `check-verbatim-moves.sh` cover its citation and its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): end a removal task on an absence sweep
**After:** Task 1

**Decision:** absence-sweep-rule

- [ ] 3. Isolate a suspected pre-existing failure before classifying it (KAN-570)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "pre-date the change" skills/flow-contracts/known-bugs.md; echo $?`: it prints only
    `1` — the sweep says how a pre-existing failure blocks, never how one is told from the change's own.
    <!-- measured: the grep above in this worktree @ 7d7a1b40 → exit 1, no line -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 7d7a1b40 after Tasks 1 and 2; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-verbatim-moves.sh 7d7a1b40 exit 0
diff --git a/skills/flow-contracts/known-bugs.md b/skills/flow-contracts/known-bugs.md
index ad157473..88a197c4 100644
--- a/skills/flow-contracts/known-bugs.md
+++ b/skills/flow-contracts/known-bugs.md
@@ -17,6 +17,14 @@ file, never filed for later, and never skipped, worked around or
 tolerance-widened (**Fix determinism at the source, never by widening tolerance**,
 `rules/fix-determinism-at-the-source.mdc`).
 
+**A failure suspected to pre-date the change is isolated before it is classified.** Revert every
+file the change touched in the failing area to its merge-base content, re-run the failing tests,
+then restore the files: a failure that persists pre-dates the change, and one that goes away is the
+change's own. The classification decides only whose defect it is and where its fix is recorded —
+never whether it blocks. KAN-570 is the case that set the standard: kan-485 reverted its own
+calculator-area touches, saw the same failures, and so kept a pre-existing defect out of its own
+scope on evidence rather than assertion.
+
 **A stale snapshot baseline on the base is re-baselined, never asked.** A screenshot mismatch is a
 failure whose run produced an `-actual.png` or `-diff.png` attachment; a "waiting for stable
 screenshot" timeout is a timeout, never a mismatch. A mismatch met during a sync's targeted
````

  - [ ] **Step 3: Verify.** `grep -c "pre-date the change" skills/flow-contracts/known-bugs.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow-contracts/known-bugs.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow-contracts): isolate a suspected pre-existing failure before classifying it
**After:** none

**Decision:** isolate-before-classifying

- [ ] 4. Give the capture-scope report its frame-bound slots and a Blocking tripwire (KAN-603)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "bands | seams | matrix" skills/flow/visual-verify-verifier.md; grep -n "capture-scope" skills/flow/visual-verify.md; echo $?`:
    it prints only `1` — the template has no capture-scope slot but sweeps, and Blocking names no
    capture-scope duty.
    <!-- measured: both greps in this worktree @ 7d7a1b40 → no line, exit 1 -->
  - [ ] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 7d7a1b40 after Tasks 1 to 3; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-verbatim-moves.sh 7d7a1b40 exit 0
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index a269e7e7..3123c52d 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -561,6 +561,7 @@ or a measurement the pixels could not give is step 10 undone.
 - per view, no frame composed: <view id> sweeps: text | order | reach | derived | rows |
   ink — the same sweep line at capture scope; `order` carries its containment half's
   outcome beside n/a — no frame; `rows` and `ink` read n/a — no frame
+- per view, no frame composed: <view id> bands | seams | matrix — n/a — no frame
 - added sweep <name> (sweeps-<n>.md): <`done`, or `n/a — <why it does not apply>` — every view and frame it read, and each departure it found>
 - <frame id> matrix: <n> elements × 11 columns, <k> n/a — <each n/a cell as `<element>.<column>: <why>`; the matrix itself is in visual-verification.md>
 - frames: <n>/<m> — <m> the change's own declared list, then every declared frame id with no line above and why
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index b06d702e..2d48863b 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -285,7 +285,9 @@ an unreadable PNG, **a `compose-mockup-frames.sh` exit 1 or 2 and a departure fr
 verifier reports in a composite**, **a composed frame with a `missing`, `extra` or over-tolerance
 band whose line names no cause — a pre-existing attribution is not one (**the sweep**,
 `skills/flow-contracts/known-bugs.md`) — and a composed frame with a `missing`,
-`extra` or over-tolerance seam or an off-centre line whose line names no cause**, and **a defect the
+`extra` or over-tolerance seam or an off-centre line whose line names no cause**, **a captured view
+no frame was composed for whose report carries no capture-scope `sweeps:` line, or one whose line
+leaves a sweep neither `done` nor n/a — <why it does not apply>**, and **a defect the
 verifier reports in a captured screenshot — even when every assertion passed.** **A named motion
 with no strip under `<changeRoot>/visual-verification/motion/` or no entry in
 `visual-verification.md`, and a frame its strip shows breaking the motion step's rule, block as a
````

  - [ ] **Step 3: Verify.** `grep -c "bands | seams | matrix" skills/flow/visual-verify-verifier.md` and
    `grep -c "capture-scope" skills/flow/visual-verify.md` each print `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`, `skills/flow/visual-verify.md`
**Tests:** **none** — a contract template and a Blocking item; `check-verbatim-moves.sh` covers the acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(visual-verify): give the capture-scope report its frame-bound slots and a Blocking tripwire
**After:** none

**Decision:** capture-scope-report
