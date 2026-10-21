# self-review-kan-1003-one-block-a-gym-workout-can-t-be-saved-without

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1003's seven self-review findings — A (Task 1), B (Tasks 2 and 3, the repair
split out per decision `ticket-repair-own-task`), C (Task 4), D (Task 5), F (Task 6), G (Task 7),
H (Task 8), as `proposal.md` names them.

**Architecture:** one Go guard rule, three prose rules, a gymie ticket line, two hook checks, one
new script, and a store/CLI refusal pair with two `flowd` repair commands run once against the dev
store.

**Tech Stack:** Go (`stats/`), Python 3 (the hook), Bash (scripts and harnesses), Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-22-self-review-kan-1003-one-block-a-gym-workout-can-t-be-saved-without-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Each task names the repository it runs in; every `**Files:**` path is relative to that worktree's root. Tasks 1, 2 and 4–8 run in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1003-one-block-a-gym-workout-can-t-be-saved-without`; Task 3 runs in gymie's `/Users/tweety53/Projects/gymie-worktrees/kan-1003-one-block-a-gym-workout-can-t-be-saved-without` (branch `spectre/kan-1003-one-block-a-gym-workout-can-t-be-saved-without`). Never touch a main checkout, and never create a worktree.
- A diff a task gives is applied with `git apply` from its worktree's root, in task order; Tasks 1, 4 and 5's diffs were applied in that order and checked in a scratch clone at `75261910`. Save a diff by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- A prose block lands character for character where its step says: `scripts/check-verbatim-moves.sh` matches every added or reworded sentence against `verbatim-moves.txt` exactly. All of Tasks 2, 4, 6, 7 and 8's prose was applied in a scratch clone at `75261910` and `check-verbatim-moves.sh`, `check-markdown-integrity.py`, `check-references.sh`, `check-dispatch-paragraphs.sh` and `check-guard-symlinks.sh` exited 0 there.
- Go: run `gofmt -w` on every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0. Store tests use the `flow-postgres` compose stack on port 5433 through `internal/storetest`.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container. Tasks 4 and 8's repairs write the dev store's `flow` database directly through `go run ./cmd/flowd <repair>` with `FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable'` set on the command — the same route `flowd restamp-dispatch` took for kan-995 — and touch only the rows they name.

## Review Focus

- Task 1: a dotless name under a tracked directory, or under a tracked ancestor of the ticket's own directory, is still judged; a dotted name is still judged; a root-absolute or climbing name is never exempted.
- Task 4: a commit whose message the command does not state (`-F`, `--amend`, `$MSG`) is never denied; a commit in one repository and a push from another never pair; a planning commit pathspec-limited to `spectre/changes/` passes.
- Task 5: `2>&1`, `>&2` and `>/dev/null` stay allowed from the main checkout; an unbalanced-quote command still has its glued `>f` judged.
- Task 8: a dispatched (non-inline) row is never refused for its commit time; a blocked or aborted inline row with no commit is accepted; an `end` whose `-commit` no worktree knows sends no `commitAt` and is not refused for it.
- Task 8's repair: every one of the 17 rows ends at its table values, and no other row changes.

- [x] 1. Read a dotless name under no tracked directory as a branch in check-done-when-paths
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In a scratch directory outside every worktree, `git clone -q --single-branch --branch develop /Users/tweety53/Projects/gymie gclone`, then `scripts/check-done-when-paths.sh <that clone>` from this worktree: exit 1 and 8 `DONE-WHEN-PATH` lines, every one `feature/cardio` or `feature/workout-rows`.
    <!-- measured: scripts/check-done-when-paths.sh <gymie develop single-branch clone> @ 75261910 → exit 1, 8 DONE-WHEN-PATH lines -->
  - [x] **Step 2: Apply the diff** (`git apply <the saved diff>`):

````diff verified:git apply in a scratch clone at 75261910; go test ./internal/guard -run TestCheckDoneWhenPaths passes, the new test fails on the unpatched guard, and the gymie clone of Step 1 exits 0
diff --git a/stats/internal/guard/donewhenpaths.go b/stats/internal/guard/donewhenpaths.go
index ee6f2f27..449f3118 100644
--- a/stats/internal/guard/donewhenpaths.go
+++ b/stats/internal/guard/donewhenpaths.go
@@ -60,12 +60,16 @@ func checkDoneWhenPaths(args []string, env Env, stdout, stderr io.Writer) int {
 		return die("cannot read the index of %s", root)
 	}
 	tracked := map[string]bool{}
+	dirs := map[string]bool{}
 	var markdown []string
 	for _, p := range strings.Split(raw, "\x00") {
 		if p == "" {
 			continue
 		}
 		tracked[p] = true
+		for d := path.Dir(p); d != "."; d = path.Dir(d) {
+			dirs[d] = true
+		}
 		if strings.HasSuffix(p, ".md") {
 			markdown = append(markdown, p)
 		}
@@ -120,7 +124,7 @@ func checkDoneWhenPaths(args []string, env Env, stdout, stderr io.Writer) int {
 				continue
 			}
 			for _, p := range dwPaths(line) {
-				if !branch[p] && !dwResolves(tracked, path.Dir(file), p) && !seen[p+"\x00"+file] {
+				if !branch[p] && !dwBranchShaped(dirs, path.Dir(file), p) && !dwResolves(tracked, path.Dir(file), p) && !seen[p+"\x00"+file] {
 					seen[p+"\x00"+file] = true
 					offenders = append(offenders, fmt.Sprintf("%s — %s", p, file))
 				}
@@ -161,6 +165,30 @@ func dwResolves(tracked map[string]bool, dir, named string) bool {
 	}
 }
 
+// dwBranchShaped reports whether named reads as a branch rather than a path
+// even when no ref of that name exists (a merged-and-deleted `feature/cardio`,
+// kan-1003's self-review): its last segment carries no dot, it neither climbs
+// (`../`) nor starts at the root, and its first segment — read relative to dir
+// or any ancestor of dir up to the root, as dwResolves reads the whole name —
+// is no tracked directory. A dotless file under a tracked directory
+// (`src/Makefile`) is still judged.
+func dwBranchShaped(dirs map[string]bool, dir, named string) bool {
+	if strings.HasPrefix(named, "/") || strings.HasPrefix(named, "../") ||
+		strings.Contains(named[strings.LastIndex(named, "/")+1:], ".") {
+		return false
+	}
+	first, _, _ := strings.Cut(named, "/")
+	for {
+		if dirs[path.Join(dir, first)] {
+			return false
+		}
+		if dir == "." {
+			return true
+		}
+		dir = path.Dir(dir)
+	}
+}
+
 // dwPaths extracts the path-like tokens of one line: a whitespace field
 // stripped of its backticks and edge punctuation — markdown link and autolink
 // syntax included, so `[e](shots/27.png)` and `<shots/27.png>` yield the path
````

  - [x] **Step 3: Add the test** at the end of `stats/internal/guard/check_done_when_paths_test.go`:

````go verified:passes in a scratch clone at 75261910 with Step 2 applied; fails without it
// A branch named in a Done-when that no ref carries any more — merged and
// deleted, or never fetched into this clone — is still no path (kan-1003's
// self-review: a gymie clone without feature/cardio refused every archive),
// while a dotless file under a tracked directory, a dotted name and a name
// under a tracked ancestor of the ticket's directory are still judged.
func TestCheckDoneWhenPathsBranchShapedNames(t *testing.T) {
	t.Parallel()
	dir := dwRepo(t, "src/main.go", "docs/x/notes/a.md")
	dwFile(t, dir, "docs/x/tickets/t.md",
		"## Done when\nBranch `feature/one-block` in both repos, `origin/feature/cardio` merged; "+
			"`src/Makefile`, `src/gen/Makefile`, `notes/Makefile` and `feature/notes.md` committed.\n")
	code, out, errb := runDoneWhen(t, dir)
	want := "DONE-WHEN-PATH: src/Makefile — docs/x/tickets/t.md\n" +
		"DONE-WHEN-PATH: src/gen/Makefile — docs/x/tickets/t.md\n" +
		"DONE-WHEN-PATH: notes/Makefile — docs/x/tickets/t.md\n" +
		"DONE-WHEN-PATH: feature/notes.md — docs/x/tickets/t.md\n" +
		"DONE-WHEN-VIOLATION: " + dir + " — 4\n"
	if code != 1 || out != want {
		t.Fatalf("exit %d, stdout %q, stderr %q; want 1 and %q", code, out, errb, want)
	}
}
````

  - [x] **Step 4: Update the shim's header.** In `scripts/check-done-when-paths.sh`, after the sentence ending "is a branch, not a path, and is\n# never judged.", add these comment lines:

```bash verified:comment text only, authored for this change
# A token whose last segment carries no dot, which neither starts with `/`
# nor climbs with `../`, and whose first segment — read relative to the
# naming file's directory or any directory between it and the root — names
# no tracked directory is a branch too, whether or not a ref of that name
# exists: a merged-and-deleted `feature/cardio` (kan-1003's self-review). A
# dotless file under a tracked directory (`src/Makefile`) is still judged.
```

  - [x] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckDoneWhenPaths' -count=1` passes; Step 1's command now exits 0 with `DONE-WHEN-OK`; `scripts/check-done-when-paths.sh "$PWD"` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/donewhenpaths.go`, `stats/internal/guard/check_done_when_paths_test.go`, `scripts/check-done-when-paths.sh`
**Tests:** `TestCheckDoneWhenPathsBranchShapedNames`
**Regression:** `TestCheckDoneWhenPathsBranchShapedNames` fails — `feature/one-block` and `origin/feature/cardio` are reported as missing paths again.
**Baseline:** before=1 after=2 <!-- measured: grep -c -E '^func Test' stats/internal/guard/check_done_when_paths_test.go @ 75261910 → 1, and 2 after the change in a scratch clone -->
**Commit:** fix(guard): read a dotless name under no tracked directory as a branch
**After:** none

**Decision:** done-when-branch-shape

- [x] 2. Amend a handoff ticket's Done-when when a decision overrules it
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'departs from a handoff ticket' skills/flow/brainstorm-planner.md` prints `0`.
    <!-- measured: grep -c 'departs from a handoff ticket' skills/flow/brainstorm-planner.md @ 75261910 → 0 -->
  - [x] **Step 2: Add the rule.** In `skills/flow/brainstorm-planner.md`, after the paragraph ending "enforced by **Archive on\nthe change branch** (`skills/flow-contracts/finish-contract-run1.md`).", add one blank line and this paragraph:

```markdown verified:applied in a scratch clone at 75261910; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt
**A decision that departs from a handoff ticket's Done-when amends that line before the plan-gate
commit.** When a `## Decisions` entry chooses otherwise than a handoff ticket's `## Done when`
states — a branch, a landing target, an acceptance criterion — rewrite that line to what the
decision chose, citing the decision's `**ID:**` and its date, in its own commit carrying
`-m "Flow-Key: flow.brainstorm"`, so integrate never stops on a criterion the plan already
overruled (kan-1003: decision `lands-on-develop` left "Branch `feature/one-block` in both repos,
never develop" standing, and integrate stopped and asked again).
```

  - [x] **Step 3: Verify** — `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-references.sh` each exit 0. Commit.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** none — prose; the lint guards are the check
**Regression:** none — prose only
**Commit:** docs(flow): amend a handoff ticket's Done-when a decision overrules
**After:** none

**Decision:** done-when-amended-by-decision

- [x] 3. Cite the plan decision that landed KAN-1003 on develop
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-worktrees/kan-1003-one-block-a-gym-workout-can-t-be-saved-without`, `git branch --show-current` prints `spectre/kan-1003-one-block-a-gym-workout-can-t-be-saved-without`, and `tail -1 docs/design/one-block/tickets/KAN-1003-step-1.md` prints the line Step 2 replaces.
    <!-- measured: tail -1 docs/design/one-block/tickets/KAN-1003-step-1.md @ gymie e8554100 → the line Step 2 replaces -->
  - [x] **Step 2: Replace the Done-when line.** In `docs/design/one-block/tickets/KAN-1003-step-1.md`, replace

```markdown verified:tail -1 of the file at gymie e8554100
Backend tests per path green; Playwright fidelity OB1, OB3, OB5. Landed on `develop` (operator decision at integrate, 2026-10-22).
```

    with

```markdown verified:the decision lands-on-develop is in gymie's archived kan-1003 design.md, recorded by plan-gate commit 6f3cff23 dated 2026-10-21
Backend tests per path green; Playwright fidelity OB1, OB3, OB5. Landed on `develop` (plan decision `lands-on-develop`, 2026-10-21).
```

  - [x] **Step 3: Verify and commit in that gymie worktree.** `git diff --stat` names only `docs/design/one-block/tickets/KAN-1003-step-1.md`; nothing under `spectre/` is touched. Commit with the `**Commit:**` subject and `-m "Flow-Key: self-review-kan-1003-one-block-a-gym-workout-can-t-be-saved-without-fix-g1"` — a commit outside every task of gymie's change (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`). Never push.

**Build:** green
**Files:** `docs/design/one-block/tickets/KAN-1003-step-1.md`
**Tests:** none — a ticket line
**Regression:** none — documentation only
**Commit:** docs(design): cite the plan decision that landed KAN-1003 on develop
**After:** none

**Decision:** ticket-repair-own-task

- [x] 4. Judge a commit and its push in one command, and clear verdict 646
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run the hook on kan-1003's 10:52 command shape, aimed at gymie's kan-1003 worktree (branch `spectre/kan-1003-one-block-a-gym-workout-can-t-be-saved-without`, fully pushed): `printf '%s' '{"tool_name":"Bash","cwd":"/tmp","tool_input":{"command":"cd /Users/tweety53/Projects/gymie-worktrees/kan-1003-one-block-a-gym-workout-can-t-be-saved-without && git commit -qm \"docs(design): x\" && git push -q"}}' | python3 hooks/protect-main-checkout.py` prints nothing (allowed); the hook only reads the command, it never runs it. `flow record verdicts -C /Users/tweety53/Projects/gymie -guard check-unfinished-work -false-positive` lists verdict 646 (kan-1003) and 436 (kan-941).
    <!-- measured: the hook command above @ 75261910 → no output, and a deny with this task's hook diff in a scratch clone; flow record verdicts -guard check-unfinished-work -false-positive, run from /Users/tweety53/Projects/gymie @ dev store 2026-10-22 → ids 646 and 436 -->
  - [x] **Step 2: Apply the hook diff** (`git apply <the saved diff>`):

````diff verified:git apply in a scratch clone at 75261910; scripts/test-protect-main-checkout.sh passes, and case 143 fails without the hook half
diff --git a/hooks/protect-main-checkout.py b/hooks/protect-main-checkout.py
index 18e3e61c..45980d88 100755
--- a/hooks/protect-main-checkout.py
+++ b/hooks/protect-main-checkout.py
@@ -621,6 +621,49 @@ def untasked_push(args, target):
     return None
 
 
+# A Conventional Commits `chore` subject: the pipeline's planning, link and archive commits.
+CHORE_SUBJECT_RE = re.compile(r"^chore(\([^)]*\))?!?:")
+TRAILER_RE = re.compile(r"^(Task-Id|Flow-Key):[ \t]*\S", re.M)
+# Options that take the message from somewhere the command text does not show.
+UNREAD_MESSAGE_OPTS = ("-F", "--file", "-C", "-c", "--reuse-message", "--reedit-message", "--fixup", "--squash", "--amend")
+
+
+def commit_lacks_trailer(args, target):
+    """True when `git commit <args>` states its message with -m/--message, that message carries
+    neither a `Task-Id:` nor a `Flow-Key:` trailer, its subject is no `chore`, and its pathspec
+    is not limited to `<spec-root>/changes/`. A message the command text does not state — none
+    given, -F, -C, --amend, a shell expansion left in it — is never judged."""
+    msgs, paths, i = [], [], 0
+    while i < len(args):
+        a = args[i]
+        if a == "--":
+            paths = args[i + 1 :]
+            break
+        if a.startswith(UNREAD_MESSAGE_OPTS):
+            return False
+        if a == "--message" and i + 1 < len(args):
+            msgs.append(args[i + 1])
+            i += 2
+            continue
+        if a.startswith("--message="):
+            msgs.append(a.partition("=")[2])
+        elif a.startswith("-") and not a.startswith("--") and "m" in a[1:]:
+            rest = a[a.index("m", 1) + 1 :]
+            if rest:
+                msgs.append(rest)
+            elif i + 1 < len(args):
+                msgs.append(args[i + 1])
+                i += 2
+                continue
+        i += 1
+    if not msgs or any("$" in m or "`" in m for m in msgs):
+        return False
+    if TRAILER_RE.search("\n".join(msgs)) or CHORE_SUBJECT_RE.match(msgs[0]):
+        return False
+    planning = ("/spectre/changes/", "/openspec/changes/")
+    return not (paths and all(any(s in resolve(p, target) + "/" for s in planning) for p in paths))
+
+
 def bash_hits(command, cwd, landings=None):
     """Paths a Bash command writes into or git-mutates, resolved against cwd and `cd`.
 
@@ -635,6 +678,7 @@ def bash_hits(command, cwd, landings=None):
     if target:
         return [target]
     hits = []
+    commits = []
     cur = cwd
     env = {}
     for line in logical_lines(strip_data_heredocs(command)):
@@ -677,6 +721,13 @@ def bash_hits(command, cwd, landings=None):
                     j += 1
                 if j < n and tokens[j] in DENIED_GIT_VERBS:
                     hits.append(target)
+                if j < n and tokens[j] == "commit" and target is not None:
+                    cargs = []
+                    for t in tokens[j + 1 :]:
+                        if t in SEPARATORS:
+                            break
+                        cargs.append(t)
+                    commits.append((target, cargs))
                 if j < n and tokens[j] == "push" and landings is not None:
                     args = []
                     for t in tokens[j + 1 :]:
@@ -695,6 +746,17 @@ def bash_hits(command, cwd, landings=None):
                     untasked = None if unreadable else untasked_push(args, target)
                     if untasked:
                         landings.append(untasked)
+                    dsts = [d for _, d in push_refspecs(args, target) if d.startswith("spectre/")] if target else []
+                    for ct, cargs in commits:
+                        if dsts and os.path.realpath(ct) == os.path.realpath(target) and commit_lacks_trailer(cargs, ct):
+                            landings.append(
+                                f"Blocked: this command commits and then pushes to {dsts[0]}, and that commit's "
+                                "message carries neither a `Task-Id:` nor a `Flow-Key:` trailer — the push check runs "
+                                "before the command does and cannot see the commit. Add `-m \"Flow-Key: <key>\"` "
+                                "(**Commits outside every task**, skills/flow-contracts/git-boundaries.md), or "
+                                "`-m \"Task-Id: <n>\"` on a task's own commit, then run it again."
+                            )
+                            break
                 i = j + 1
                 continue
             if tok == "sed" and any(t.startswith("-i") for t in tokens[i + 1 : i + 4]):
diff --git a/scripts/test-protect-main-checkout.sh b/scripts/test-protect-main-checkout.sh
index 40cb3ba4..a5415e76 100755
--- a/scripts/test-protect-main-checkout.sh
+++ b/scripts/test-protect-main-checkout.sh
@@ -680,6 +680,22 @@ expect "107 a checker that cannot run fails open" allow \
   "$(UNTASKED_COMMITS_CHECK="$ROOT/missing" run "$ROOT" Bash "{\"command\":$(q "git -C $CWT push origin HEAD:spectre/chg2")}")"
 unset UNTASKED_COMMITS_CHECK
 
+# 143-148: a commit and its push in one command (kan-1003's self-review: the design handoff
+# 0497d943 was committed and pushed in one call, before the push check could see it).
+echo h >"$CWT/h.txt"
+expect "143 commit and push in one command, no trailer" deny \
+  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git add h.txt && git commit -qm 'docs(design): add the handoff' && git push -q")}")"
+expect "144 the same with a Flow-Key trailer" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git commit -q -m 'docs(design): add the handoff' -m 'Flow-Key: flow.brainstorm' && git push -q")}")"
+expect "145 a planning commit pathspec-limited to spectre/changes/" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q "git -C $CWT commit -q -m 'docs(spectre): record' -- spectre/changes/chg && git -C $CWT push -q origin spectre/chg3")}")"
+expect "146 a chore subject" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git commit -q -m 'chore(spectre): plan chg — plan gate' && git push -q")}")"
+expect "147 a message the command does not state" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git commit -q -F msg.txt && git push -q")}")"
+expect "148 a trailer-less commit pushed to a branch that is no change branch" allow \
+  "$(run "$ROOT" Bash "{\"command\":$(q "cd $CWT && git commit -qm 'docs(x): y' && git push -q origin HEAD:feature/x")}")"
+
 # 108-109: the real checker after a rebase onto a moved base — the base's own
 # trailer-less commits are on origin already, never the change's to judge.
 RWT="$ROOT/main-worktrees/chg4"
````

  - [x] **Step 3: Document the check.** In the hook's module docstring, after the paragraph opening "It also denies a push of a `/flow` change branch (`spectre/<name>`) whose unpushed commits", add a paragraph: "It also denies a command that commits and then pushes the same repository to a `spectre/` branch when that commit's message, as the command states it (`-m`, `-qm`, `--message`), carries neither trailer, its subject is no `chore`, and its pathspec is not limited to `<spec-root>/changes/` — the push check above runs before the command does and cannot see a commit the command makes first (kan-1003: 0497d943). A message the command does not state (`-F`, `-C`, `--amend`, a shell expansion) is not judged." In the harness header, add one line per case 143–148, in its style.
  - [x] **Step 4: Add the prose.** In `skills/flow-contracts/git-boundaries.md`, after the paragraph ending "(kan-941 pushed four\nsuch commits before either check existed).", add one blank line and:

```markdown verified:applied in a scratch clone at 75261910; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt
**A commit and its push in one command are judged together.** The push check above runs before
the command does, so it cannot see a commit the same command makes first:
`hooks/protect-main-checkout.py` therefore denies a command that commits and pushes a `spectre/`
branch when that commit's own `-m` text carries neither trailer and no `chore` subject, and its
pathspec is not limited to `<spec-root>/changes/` (kan-1003: the design handoff 0497d943 was
committed and pushed in one command, and run 1's gate listed it).
```

    and at the end of `skills/flow/unfinished-work-gate.md`, after "A write that falls back to the journal is one warning line and the run continues.", one blank line and:

```markdown verified:applied in a scratch clone at 75261910; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt
**An `UNTASKED` listing is never flagged this way.** `check-untasked-commits.sh` records no
verdict, so the flag would land on `check-unfinished-work`'s newest row, a `CLEAR` (kan-1003:
verdict 646). A listed commit's author is the identity every session commits under, so the relay
never calls it the operator's own.
```

  - [x] **Step 5: Write the failing store test.** In `stats/internal/store/records_test.go`, `TestUnflagVerdictClearsTheFlag`: record a verdict, flag it with `FlagVerdictFalsePositive`, call `st.UnflagVerdict(ctx, v.ID)`; assert `ListVerdicts(ctx, projectKey, guard, true)` is empty and `ListVerdicts(ctx, projectKey, guard, false)` returns the row with `FalsePositive == false`, `FalsePositiveReason == ""` and a nil `FlaggedAt`; and `st.UnflagVerdict(ctx, <an id no row has>)` returns an error wrapping `store.ErrDispatchNotFound`. Run `cd stats && go test ./internal/store -run TestUnflagVerdictClearsTheFlag -count=1`: fails to compile.
  - [x] **Step 6: Implement.** `func (s *Store) UnflagVerdict(ctx context.Context, id int64) error` in `stats/internal/store/records.go`, beside `FlagVerdictFalsePositive`: `UPDATE guard_verdicts SET false_positive = false, false_positive_reason = NULL, flagged_at = NULL WHERE id = $1`, zero rows → `fmt.Errorf("%w: verdict %d", ErrDispatchNotFound, id)`. In `stats/cmd/flowd/main.go`, an `unflag-verdict` branch beside `restamp-dispatch` in `main` and `func unflagVerdict(args []string) error` beside `restampDispatch`: usage `flowd unflag-verdict <verdict-id>`, the id parsed with `strconv.ParseInt`, `config.FromEnv`, `store.Open`, `st.UnflagVerdict`. Run Step 5's command: passes.
  - [x] **Step 7: Repair verdict 646.** `cd stats && make -s web-dist && FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd unflag-verdict 646` exits 0; Step 1's `flow record verdicts … -false-positive` now lists 436 alone. No kan-994 row is flagged (`design.md`, **Context**), so nothing else is repaired.
  - [x] **Step 8: Verify.** `bash scripts/test-protect-main-checkout.sh` ends `all protect-main-checkout cases pass`; Step 1's hook command now prints a deny; `cd stats && go test ./internal/store -run 'Verdict' -count=1` and `cd stats && make -s web-dist && go vet ./cmd/flowd ./internal/store` pass; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-references.sh` and `scripts/check-python-suppressions.sh` exit 0. Commit.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`, `skills/flow-contracts/git-boundaries.md`, `skills/flow/unfinished-work-gate.md`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/cmd/flowd/main.go`
**Tests:** `143 commit and push in one command, no trailer`, `144 the same with a Flow-Key trailer`, `145 a planning commit pathspec-limited to spectre/changes/`, `146 a chore subject`, `147 a message the command does not state`, `148 a trailer-less commit pushed to a branch that is no change branch`, `TestUnflagVerdictClearsTheFlag`
**Regression:** case 143 is allowed again — a trailer-less commit pushed in its own command reaches origin; `TestUnflagVerdictClearsTheFlag` fails — a flagged verdict cannot be cleared.
**Baseline:** before=215 after=222 <!-- measured: cat scripts/test-protect-main-checkout.sh stats/internal/store/records_test.go | grep -c -E '^(expect "|func Test)' @ 75261910 → 215; 221 after the hook diff in a scratch clone, 222 with the store test -->
**Commit:** fix(hooks): judge a commit and its push made in one command
**After:** none

**Decision:** commit-push-one-command, untasked-never-flagged, flowd-unflag-verdict

- [x] 5. Read only standalone redirect operators in protect-main-checkout
  - [x] **Step 1: Reproduce before touching anything; do not guess.** From this worktree, `printf '%s' '{"tool_name":"Bash","cwd":"/Users/tweety53/Projects/gymie","tool_input":{"command":"python3 grab.py \"/Users/tweety53/Projects/gymie-worktrees/x/a.html\" '"'"'>Save<'"'"' 1 1200"}}' | python3 hooks/protect-main-checkout.py` prints a deny.
    <!-- measured: the command above @ 75261910 → permissionDecision deny -->
  - [x] **Step 2: Apply the diff** (`git apply <the saved diff>`), on top of Task 4's:

````diff verified:git apply in a scratch clone at 75261910 after Task 4's hook diff; scripts/test-protect-main-checkout.sh passes, and cases 138-141 fail on the unpatched hook
diff --git a/hooks/protect-main-checkout.py b/hooks/protect-main-checkout.py
index 45980d88..dec16ec9 100755
--- a/hooks/protect-main-checkout.py
+++ b/hooks/protect-main-checkout.py
@@ -228,7 +228,10 @@ def tokenize(command):
         lexer.whitespace_split = True
         words = list(lexer)
     except ValueError:
-        return command.split()
+        # Unbalanced quotes: raw words, a redirect glued to its target split off it, so `>f`
+        # and `2>>f` read as shlex would have split them; a quoted word keeps its quote and
+        # never matches.
+        return [p for w in command.split() for p in _split_redirect(w)]
     # A backtick span, and a `$(...)` inside a quoted word, still run: each one's command is
     # scanned again as a command of its own, after the line's words.
     spans = [sp for w in words for sp in SUBST_SPAN_RE.findall(w) if sp.startswith("$(")]
@@ -346,6 +349,18 @@ def may_run_code(text, target):
 # that bash runs.
 SUBST_SPAN_RE = re.compile(r"\$\([^()]*\)|`[^`]*`")
 
+# The operators that write a file: shlex (punctuation_chars) yields each as a word of its own, so
+# a word that merely starts with `>` came out of quotes and is data (`'>Save<'`).
+REDIRECT_OPS = (">", ">>", ">|", "&>", "&>>")
+# A redirect glued to its target in an unbalanced-quote fallback word: `>f`, `2>>f`, `&>f`; `>&1`
+# duplicates a descriptor and writes no file.
+ATTACHED_REDIRECT_RE = re.compile(r"^\d*(&>>|&>|>>|>\||>)([^&].*)$")
+
+
+def _split_redirect(word):
+    m = ATTACHED_REDIRECT_RE.match(word)
+    return [m.group(1), m.group(2)] if m else [word]
+
 
 def strip_data_heredocs(command):
     """command with the body of a `cat`/`tee` heredoc to a named file removed — that body is
@@ -777,14 +792,10 @@ def bash_hits(command, cwd, landings=None):
                         hits.append(resolve(t, cur))
                 i += 1
                 continue
-            if tok in (">", ">>") and i + 1 < n:
+            if tok in REDIRECT_OPS and i + 1 < n:
                 hits.append(resolve(tokens[i + 1], cur))
                 i += 2
                 continue
-            if tok.startswith((">", ">>")) and len(tok) > 2 and tok.lstrip(">") not in ("&1", "&2"):
-                hits.append(resolve(tok.lstrip(">"), cur))
-                i += 1
-                continue
             if tok in ("cp", "mv", "rm"):
                 args = []
                 for t in tokens[i + 1 :]:
diff --git a/scripts/test-protect-main-checkout.sh b/scripts/test-protect-main-checkout.sh
index a5415e76..05ae5994 100755
--- a/scripts/test-protect-main-checkout.sh
+++ b/scripts/test-protect-main-checkout.sh
@@ -716,6 +716,21 @@ git_q -C "$RWT" commit -q -m "fix(app): review Minors"
 expect "109 the real checker still lists the change's own untasked commit" deny \
   "$(run "$ROOT" Bash "{\"command\":$(q "git -C $RWT push --force-with-lease origin spectre/chg4")}")"
 
+# 138-142: a `>` inside quotes is data, never a redirect (kan-1003's self-review); every
+# operator that writes a file is still judged.
+expect "138 a quoted '>Save<' argument, cwd = main" allow \
+  "$(run "$MAIN" Bash "{\"command\":$(q "python3 grab.py \"$SIB/a.html\" '>Save<' 1 1200")}")"
+expect "139 a python heredoc body finding '>Save<', cwd = main" allow \
+  "$(run "$MAIN" Bash "{\"command\":$(q "python3 - <<'EOF'
+print(open('$SIB/a.html').read().find('>Save<'))
+EOF")}")"
+expect "140 '>|' into the main checkout" deny \
+  "$(run "$ROOT" Bash "{\"command\":$(q "echo x >| $MAIN/f.txt")}")"
+expect "141 '&>' into the main checkout" deny \
+  "$(run "$ROOT" Bash "{\"command\":$(q "echo x &> $MAIN/f.txt")}")"
+expect "142 an unbalanced quote, then '>' glued to a main-checkout path" deny \
+  "$(run "$ROOT" Bash "{\"command\":$(q "echo it's >$MAIN/f.txt")}")"
+
 # 3 last: moving the main checkout off main lifts the protection
 git_q -C "$MAIN" checkout -q -b feature
 expect "3 Write in main checkout on a feature branch" allow \
````

  - [x] **Step 3: Header lines.** In the harness header, add one line per case 138–142, in its style.
  - [x] **Step 4: Verify.** `bash scripts/test-protect-main-checkout.sh` ends `all protect-main-checkout cases pass`; Step 1's command prints nothing; `scripts/check-python-suppressions.sh` exits 0. Commit.

**Build:** green
**Files:** `hooks/protect-main-checkout.py`, `scripts/test-protect-main-checkout.sh`
**Tests:** `138 a quoted '>Save<' argument, cwd = main`, `139 a python heredoc body finding '>Save<', cwd = main`, `140 '>|' into the main checkout`, `141 '&>' into the main checkout`, `142 an unbalanced quote, then '>' glued to a main-checkout path`
**Regression:** cases 138 and 139 are denied again — a quoted `>` reads as a redirect into the main checkout; 140 and 141 are allowed — `>|` and `&>` write past the hook.
**Baseline:** before=131 after=136 <!-- measured: grep -c -E '^expect "' scripts/test-protect-main-checkout.sh @ 75261910 → 125; 131 after Task 4's diff and 136 after this one, in a scratch clone -->
**Commit:** fix(hooks): read only standalone redirect operators
**After:** Task 4

**Decision:** redirect-operators-only

- [x] 6. Cite the handoff source for a departure's expected value
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'inferred from the pixels' skills/flow/visual-verify-verifier.md` prints `0`.
    <!-- measured: grep -c 'inferred from the pixels' skills/flow/visual-verify-verifier.md @ 75261910 → 0 -->
  - [x] **Step 2: Add the rule.** In `skills/flow/visual-verify-verifier.md`, directly after the line `   beside it — never a bare `departure`.`, add:

```markdown verified:applied in a scratch clone at 75261910; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt
   **A departure's expected value cites the mockup's source wherever the handoff ships one, never a
   rule inferred from the pixels**: where the frame was drawn from HTML or CSS in the tree — beside
   the mockups or cited by `design.md` — the line names the declaration the expected layout, size or
   colour comes from (`grid-template-columns:1fr 1.3fr 1fr; height:32px`), and a value no
   declaration states is named as read from the pixels (kan-1003: an inferred `flex: 1 1 auto`
   cost four fix rounds against the grid the HTML declared).
```

  - [x] **Step 3: Verify** — `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-dispatch-paragraphs.sh` each exit 0. Commit.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** none — prose; the lint guards are the check
**Regression:** none — prose only
**Commit:** docs(flow): cite the handoff source for a departure's expected value
**After:** none

**Decision:** expected-cites-source

- [x] 7. Rebuild the full app suite's zip with one script
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'zip -X -j' skills/flow/visual-verify-verifier.md` prints `1` and `grep -c 'full-app-suite.zip' skills/flow/verify-fix-loop.md` prints `0`: the rebuild lives only in the verifier's step 8, and the in-run fix loop never names it.
    <!-- measured: both greps @ 75261910 → 1 and 0 -->
  - [x] **Step 2: Write the harness** `scripts/test-rebuild-full-app-suite-zip.sh` (executable) and run `bash scripts/test-rebuild-full-app-suite-zip.sh`: it fails, the script being absent.

````bash verified:passes in a scratch clone at 75261910 with Step 3's script; exits non-zero without it
#!/usr/bin/env bash
# Assertion harness for rebuild-full-app-suite-zip.sh: builds a throwaway
# project whose `screenshots` root holds the full app suite's PNGs beside
# another spec's, and runs the REAL script against it.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REBUILD="$SCRIPT_DIR/rebuild-full-app-suite-zip.sh"
FAILURES=0
fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

P="$(mktemp -d "${TMPDIR:-/tmp}/rebuild-full-app-suite-zip.XXXXXX")"
trap 'rm -rf "$P"' EXIT
mkdir -p "$P/.flow" "$P/shots/full-app-suite.spec.ts-snapshots" "$P/shots/other.spec.ts-snapshots"
printf '# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `screenshots` | `shots` |\n' >"$P/.flow/project.md"
printf a >"$P/shots/full-app-suite.spec.ts-snapshots/a.png"
printf b >"$P/shots/full-app-suite.spec.ts-snapshots/b.png"
printf c >"$P/shots/other.spec.ts-snapshots/c.png"
Z="$P/shots/full-app-suite.zip"

# Case 1: the zip holds exactly the suite's PNGs, by base name.
"$REBUILD" "$P" "$P/shots" && rc=0 || rc=$?
names=$(unzip -Z1 "$Z" | sort | tr '\n' ' ')
[ "$rc" -eq 0 ] && [ "$names" = "a.png b.png " ] && pass "case 1 the suite's PNGs only" \
  || fail "case 1: exit $rc, entries '$names'"

# Case 2: a rebuild over unchanged images is byte-identical, whatever their mtimes.
cp "$Z" "$P/first.zip"
touch "$P/shots/full-app-suite.spec.ts-snapshots/a.png"
"$REBUILD" "$P" "$P/shots"
cmp -s "$Z" "$P/first.zip" && pass "case 2 byte-identical rebuild" || fail "case 2: the rebuild differs"

# Case 3: a PNG the suite no longer produces leaves the zip.
rm "$P/shots/full-app-suite.spec.ts-snapshots/b.png"
"$REBUILD" "$P" "$P/shots"
names=$(unzip -Z1 "$Z" | sort | tr '\n' ' ')
[ "$names" = "a.png " ] && pass "case 3 a gone PNG leaves the zip" || fail "case 3: entries '$names'"

# Case 4: no suite PNG is exit 1 with no zip left behind.
rm "$P/shots/full-app-suite.spec.ts-snapshots/a.png"
"$REBUILD" "$P" "$P/shots" && rc=0 || rc=$?
[ "$rc" -eq 1 ] && [ ! -e "$Z" ] && pass "case 4 nothing to zip" || fail "case 4: exit $rc"

# Case 5: a usage error is exit 2.
"$REBUILD" "$P" 2>/dev/null && rc=0 || rc=$?
[ "$rc" -eq 2 ] && pass "case 5 usage" || fail "case 5: exit $rc"

if [ "$FAILURES" -gt 0 ]; then
  echo "$FAILURES failure(s)" >&2
  exit 1
fi
echo "all rebuild-full-app-suite-zip cases pass"
````

  - [x] **Step 3: Write the script** `scripts/rebuild-full-app-suite-zip.sh` (executable) and its link `skills/flow/scripts/rebuild-full-app-suite-zip.sh` → `../../../scripts/rebuild-full-app-suite-zip.sh` (`ln -s`, relative). Run Step 2's harness: passes.

````bash verified:scripts/test-rebuild-full-app-suite-zip.sh passes against it in a scratch clone at 75261910; check-guard-symlinks.sh exits 0 with the link
#!/usr/bin/env bash
# rebuild-full-app-suite-zip.sh — rebuild <screenshots root>/full-app-suite.zip
# from exactly the PNGs full-app-suite.spec.ts produced. The rebuild lived only
# as inline bash in step 8 of skills/flow/visual-verify-verifier.md, so an
# in-run recapture never ran it and kan-1003's zip went stale twice; the
# verifier, the in-run fix loop (skills/flow/verify-fix-loop.md) and the base
# sync's recapture (skills/flow/sync-onto-base.md) all call this one script.
#
# Usage: rebuild-full-app-suite-zip.sh <worktree> <screenshots root>
#
# The PNG list is `resolve-visual-screenshots.sh <worktree> full-app-suite.spec.ts`.
# The old zip is removed first, so the archive never keeps a PNG the suite no
# longer produces. Every PNG's mtime is pinned to 1980-01-01 before zipping,
# since the zip stores it: a rebuild over unchanged images is byte-identical
# and commits nothing.
#
# Exit 0 the zip is written; 1 the resolver matched no PNG (nothing written);
# 2 cannot answer — a usage error, the resolver's own exit 2, or a failing zip.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this script needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
if [ "$#" -ne 2 ]; then
  echo "usage: rebuild-full-app-suite-zip.sh <worktree> <screenshots root>" >&2
  exit 2
fi
zip_path="$2/full-app-suite.zip"
rm -f "$zip_path"
rc=0
list=$("$SCRIPT_DIR/resolve-visual-screenshots.sh" "$1" full-app-suite.spec.ts) || rc=$?
[ "$rc" -eq 0 ] || exit "$rc"
printf '%s\n' "$list" | tr '\n' '\0' | xargs -0 touch -t 198001010000
printf '%s\n' "$list" | zip -q -X -j -@ "$zip_path" || exit 2
````

  - [x] **Step 4: Point the three callers at it.** In `skills/flow/visual-verify-verifier.md` replace

````markdown verified:the verifier's step 8 text at 75261910
   Then rebuild the zip from
   exactly what the suite just produced:

   ```bash
   rm -f <screenshots root>/full-app-suite.zip
   list=$(resolve-visual-screenshots.sh <worktree> full-app-suite.spec.ts)
   printf '%s\n' "$list" | tr '\n' '\0' | xargs -0 touch -t 198001010000
   printf '%s\n' "$list" | zip -X -j -@ <screenshots root>/full-app-suite.zip
   ```

   `<screenshots root>` is `screenshots` resolved the way `resolve-visual-screenshots.sh`
   resolves it. Every PNG's mtime is pinned before zipping because the zip stores it: a rebuild over
   unchanged images is then byte-identical and commits nothing. A resolver exit 1 or 2, or a non-zero `zip`, blocks.
````

    with

```markdown verified:applied in a scratch clone at 75261910; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt
   Then rebuild the zip from exactly what the suite just produced with
   `rebuild-full-app-suite-zip.sh <worktree> <screenshots root>`, the one script every recapture
   that moves the suite runs too (`skills/flow/verify-fix-loop.md`, `skills/flow/sync-onto-base.md`).

   `<screenshots root>` is `screenshots` resolved the way `resolve-visual-screenshots.sh`
   resolves it. Every PNG's mtime is pinned before zipping because the zip stores it: a rebuild over
   unchanged images is then byte-identical and commits nothing. A non-zero exit blocks.
```

    In `skills/flow/sync-onto-base.md` replace step 3's two lines with:

```markdown verified:applied in a scratch clone at 75261910; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt
  3. Rebuild `full-app-suite.zip` when it was named, with
     `rebuild-full-app-suite-zip.sh <worktree> <screenshots root>` as step 8 of
     `skills/flow/visual-verify-verifier.md` runs it.
```

    In `skills/flow/verify-fix-loop.md`, directly after the line ending "45–70 minutes a round).", add:

```markdown verified:applied in a scratch clone at 75261910; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt
   **Every recapture the round runs rebuilds the full app suite's zip after it** —
   `rebuild-full-app-suite-zip.sh <worktree> <screenshots root>` in the checkout the recapture
   wrote to, the rebuilt `full-app-suite.zip` committed with the task's PNGs — so the zip never
   trails the baselines it archives (kan-1003: two rounds left it stale).
```

  - [x] **Step 5: Verify** — `bash scripts/test-rebuild-full-app-suite-zip.sh` passes; `scripts/check-guard-symlinks.sh`, `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-references.sh` and `scripts/check-dispatch-paragraphs.sh` each exit 0; `git grep -n -F -e 'zip -X -j -@' -- skills` prints nothing. Commit.

**Build:** green
**Files:** `scripts/rebuild-full-app-suite-zip.sh`, `scripts/test-rebuild-full-app-suite-zip.sh`, `skills/flow/scripts/rebuild-full-app-suite-zip.sh`, `skills/flow/visual-verify-verifier.md`, `skills/flow/sync-onto-base.md`, `skills/flow/verify-fix-loop.md`
**Tests:** `case 1 the suite's PNGs only`, `case 2 byte-identical rebuild`, `case 3 a gone PNG leaves the zip`, `case 4 nothing to zip`, `case 5 usage`
**Regression:** the harness fails — the script is gone and every caller's rebuild with it.
**Baseline:** before=0 after=5 <!-- measured: cat scripts/test-rebuild-full-app-suite-zip.sh 2>/dev/null | grep -c -E '^# Case [0-9]+:' @ 75261910 → 0, and 5 with the harness in a scratch clone -->
**Commit:** feat(scripts): rebuild the full app suite's zip with one script
**After:** Task 6

**Decision:** zip-rebuild-script

- [x] 8. Refuse an inline row begun after its commit, and retime kan-1003's rows
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `docker exec -i flow-postgres psql -U flow -d flow -At -c "select id, extract(epoch from ended_at-started_at)::int, coalesce(left(commit_sha,8),'-') from dispatches where id in (6799,6842,6903) order by id"` prints `6799|2700|22c9796f`, `6842|11|-` and `6903|3|-`.
    <!-- measured: the psql query above @ dev store 2026-10-22 → 6799|2700|22c9796f, 6842|11|-, 6903|3|- -->
  - [x] **Step 2: Write the failing store tests** in `stats/internal/store/records_test.go`, after `TestEndDispatchRefusesAZeroWidthInlineRow`:
    - `TestEndDispatchRefusesAnInlineTaskRowWithNoCommit` — an inline row keyed `task-3-implementer` ended `completed` with no `CommitSHA`, 10 s after its begin, returns an error wrapping `store.ErrInlineCommitMissing`; the same end with a `CommitSHA` succeeds; an inline row keyed `verify-gymie` ended with no commit succeeds; an inline task row ended `blocked` with no commit succeeds.
    - `TestEndDispatchRefusesAnInlineRowBegunAfterItsCommit` — an inline task row whose end carries `CommitAt` one minute before its begin returns an error wrapping `store.ErrInlineBegunAfterCommit`; `CommitAt` one minute after its begin succeeds; a dispatched row (agent id not `inline`) with `CommitAt` before its begin succeeds.
    - `TestRetimeDispatchSetsInstantsAndCommit` — `st.RetimeDispatch(ctx, id, started, ended, sha)` sets the row's `started_at`, `ended_at` and `commit_sha` exactly; an id no row has returns an error wrapping `store.ErrDispatchNotFound`; `started` after `ended` is an error and changes nothing.
    `TestEndDispatchRefusesAZeroWidthInlineRow`'s closing end (seven seconds after the begin, key `task-1-implementer`) gains a `CommitSHA`, since a completed inline task row with no commit is now refused. Run `cd stats && go test ./internal/store -run 'TestEndDispatchRefusesAnInline|TestRetimeDispatch' -count=1`: fails to compile.
  - [x] **Step 3: Implement the store half.** In `stats/internal/records/types.go`, `records.DispatchEnd` gains a `CommitAt *time.Time` field, JSON key `commitAt` with `omitempty` — the commit's committer time, as the CLI read it. In `stats/internal/store/records.go`: sentinels `ErrInlineCommitMissing` and `ErrInlineBegunAfterCommit` beside `ErrInlineWindowTooShort`, each with a doc comment naming kan-1003; `EndDispatch` reads the row's agent id and `started_at` whenever `in.Outcome == "completed"` (no longer only when `in.Tokens == nil`) and, for an inline row, keeps the window floor for a token-less end and refuses (a) a key that starts `task-` and contains `-implementer` with an empty `CommitSHA`, (b) a `CommitAt` before `started_at`, each error naming the key, the instants and the remedy (`-commit <the task's commit>`; `flow record dispatch backfill` with the real instants). `func (s *Store) RetimeDispatch(ctx context.Context, id int64, started, ended time.Time, commit string) error` beside `RestampDispatch`. In `stats/internal/api/server.go`'s `mapStoreError` both sentinels answer `http.StatusConflict`, as `ErrInlineWindowTooShort` does; in `stats/internal/reconcile/reconcile.go` both are retired on replay beside it. Run Step 2's command: passes.
  - [x] **Step 4: Pin the mapping and the retirement.** Add rows `{"inline commit missing", store.ErrInlineCommitMissing, http.StatusConflict}` and `{"inline begun after commit", store.ErrInlineBegunAfterCommit, http.StatusConflict}` to `TestErrorStatusMapping`'s table in `stats/internal/api/changes_test.go`. In `stats/internal/reconcile/record_test.go`, `TestDispatchEndRefusedForItsCommitIsRetired`, shaped as `TestDispatchEndRefusedAsTooShortIsRetiredAndAppliesEntryBehindIt`, over both sentinels: each journalled end retires and the finding behind it applies (`{Applied:1 Refused:1}`, no pending entry). `cd stats && go test ./internal/api -run TestErrorStatusMapping -count=1` and `cd stats && go test ./internal/reconcile -run TestDispatchEndRefused -count=1` pass.
  - [x] **Step 5: Write the failing CLI tests** in `stats/cmd/flow/record_test.go`, after `TestRecordDispatchBackfillInstants`; each builds its fixture commits with `GIT_COMMITTER_DATE` pinned:
    - `TestRecordDispatchEndSendsCommitAt` — `flow record dispatch end -commit <a fixture commit>` sends `commitAt` equal to that commit's committer time (RFC 3339); `-commit` of a sha no worktree knows sends no `commitAt` key and still exits 0.
    - `TestRecordDispatchBackfillDerivesInlineInstants` — two fixture commits at `2026-10-21T11:02:00Z` and `2026-10-21T11:10:31Z`; `backfill -agent-id inline -commit <the second>` with no instants sends begin `startedAt` `2026-10-21T11:02:00Z` and end `endedAt` `2026-10-21T11:10:31Z`; the same call with `-started-at` or `-ended-at` exits 2 with no store call; `backfill -agent-id none` without instants still exits 2 as today.
    Run `cd stats && make -s web-dist && go test ./cmd/flow -run 'TestRecordDispatchEndSendsCommitAt|TestRecordDispatchBackfillDerivesInlineInstants' -count=1`: fails.
  - [x] **Step 6: Implement the CLI half** in `stats/cmd/flow/record.go`. `func commitTime(sha string, roots []string) (time.Time, bool)` — `git -C <root> show -s --format=%cI <sha>` over roots in order, the first that knows it. `func commitBefore(t time.Time, roots []string) (time.Time, bool)` — the newest `git -C <root> log -1 --format=%cI --before=<t minus one second, RFC 3339> HEAD` across roots. `runRecordDispatchEnd`, given `-commit`, sets `in.CommitAt` from `commitTime` over `-C`'s repository first, then `mutationRoots` only when that misses — so an end whose sha `-C` knows makes no state-record read. `runRecordDispatchBackfill`, given `-agent-id inline` and `-commit`: refuses `-started-at` or `-ended-at` (exit 2, naming that both derive from the commit); resolves the project key with `fallback.ProjectKey(f.dir)` and the roots with `mutationRoots`; derives `-ended-at` from `commitTime` and `-started-at` from `commitBefore` (either missing: exit 2 naming the sha); and passes both on to begin and end as if given. Any existing `cmd/flow` test whose fake daemon now also answers the state-record read of `mutationRoots` (an `end` with a `-commit` its fixture repository lacks) counts POST bodies only. unverified:list which existing cmd/flow tests count every daemon request on an end carrying a -commit the fixture repository lacks — TestRecordDispatchBackfill is one — and narrow each to POSTs. Run Step 5's command, then `cd stats && make -s web-dist && go test ./cmd/flow -run 'RecordDispatch' -count=1`: pass.
  - [x] **Step 7: Add the repair command.** In `stats/cmd/flowd/main.go`, a `retime-dispatch` branch beside `restamp-dispatch` in `main` and `func retimeDispatch(args []string) error`: usage `flowd retime-dispatch <dispatch-id> <started-at> <ended-at> <commit>`, both instants RFC 3339, then `st.RetimeDispatch`. `cd stats && make -s web-dist && go vet ./cmd/flowd` passes.
  - [x] **Step 8: Add the prose.** In `skills/flow/implement.md`, after the line ending "`-tokens` is refused where `CLAUDE_CODE_SESSION_ID` is set." in the **Records:** bullet, add these lines at the bullet's indent:

```markdown verified:applied in a scratch clone at 75261910; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt
  It also refuses a `completed` inline row on a task's key (`task-<n>-implementer…`) that
  carries no `-commit`, and one whose `-commit` — found in any of the change's worktrees — was
  made before the row's begin, a begin recorded after the work it opens (kan-1003: tasks 10–22's
  rows opened after their commits, read 3–53 s and recorded no commit).
```

    and directly after the line opening "**A dispatch recorded only after it finished is a back-fill, and is one call**":

```markdown verified:applied in a scratch clone at 75261910; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt
An inline back-fill (`-agent-id inline`) carrying `-commit` takes neither instant:
`backfill` derives `-ended-at` from that commit's committer time and `-started-at` from the newest
commit before it in any of the change's worktrees, and refuses either flag beside it — a
hand-written instant there is a guess (kan-1003: `-started-at $(date -u -v-45M …)`).
```

  - [x] **Step 9: Retime kan-1003's rows.** For each line of the table below — `<dispatch id> <started-at> <ended-at> <commit>`, derived by **inline-backfill-derives**'s rule (`design.md`, **Context**) — run `cd stats && FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd retime-dispatch <id> <started-at> <ended-at> <commit>`; each exits 0.

```text verified:derived over the four kan-1003 worktrees' HEAD and refs/flow/pre-reshape/kan-1003-one-block-a-gym-workout-can-t-be-saved-without on 2026-10-22, the commit from each row's recorded sha, else the newest Flow-Key:<key> commit, else the newest Task-Id:<n> commit
6799 2026-10-21T11:02:00Z 2026-10-21T11:10:31Z 22c9796f13a2d44a8e8a16397ba13cd338f51743
6804 2026-10-21T11:10:31Z 2026-10-21T11:18:22Z d676328e1b0a7101f8445d0b6bc24c7875c86cfa
6817 2026-10-21T12:00:21Z 2026-10-21T12:08:55Z ac9407a4fe8d19b08de99d4d272e4a5dcc3921bb
6819 2026-10-21T12:10:33Z 2026-10-21T12:30:05Z db7e276058df4955a9b58d36fd6b725aa62737dc
6836 2026-10-21T13:27:29Z 2026-10-21T14:14:20Z 498d8d28c31e3b909a98e9520ceb5f28dcccaf34
6842 2026-10-21T15:25:38Z 2026-10-21T15:30:55Z 34f39e6d8ec46e8bbabaeb28ba63d8d930809859
6847 2026-10-21T16:02:35Z 2026-10-21T16:06:47Z 9692f7636c900ac019ebc492b8d5979faad83325
6859 2026-10-21T16:44:02Z 2026-10-21T17:00:19Z f638052470b8c802a30540b8c66a39abf6f47eaf
6875 2026-10-21T17:46:53Z 2026-10-21T18:18:54Z 28c9eb44878f7461a0b11b38359a2a9059df1aeb
6885 2026-10-21T18:24:50Z 2026-10-21T18:25:19Z 27f72956e853098923b08ffee73cd4e7f1b4ef63
6886 2026-10-21T18:25:19Z 2026-10-21T18:30:16Z 4371a8cfffef28e5d1cdf7f2f8492dedb84fa207
6887 2026-10-21T18:30:16Z 2026-10-21T18:32:56Z b52774a2e8c516f883d0aff406461827ecd43595
6888 2026-10-21T19:26:38Z 2026-10-21T19:32:52Z 28b02cdfff0e9d7290f6564e38465c7ea1201ec6
6889 2026-10-21T18:37:03Z 2026-10-21T18:38:51Z 93161dccbf3fb5425864c5cb91788d1466bd68f2
6890 2026-10-21T18:38:51Z 2026-10-21T18:41:55Z 9e6e0758eb574f9ae0b3d237c921471f0dd743eb
6902 2026-10-21T19:32:52Z 2026-10-21T19:37:51Z 39dca3223a6ea3aab545cb6e1ae361b430ffd1e1
6903 2026-10-21T19:32:52Z 2026-10-21T19:37:51Z a5e8a8b36cddb6eb4000e025813927f24b0c763a
```

  - [x] **Step 10: Restamp them.** unverified:confirm `flowd restamp-dispatch 6799 <transcript>` exits 0 before looping. For each id in Step 9's table, `cd stats && FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd restamp-dispatch <id> /Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/85208c79-6c1b-4be3-9f2a-5b8150f79a68.jsonl` — the parent transcript carrying session token `mf-k1003b`. An id whose restamp exits non-zero is reported with its output, never retried blind.
  - [x] **Step 11: Verify.** The `design.md` **Live check**'s psql query prints every row at its table values. `cd stats && go test ./internal/store -run 'EndDispatch|RetimeDispatch|RestampDispatch' -count=1`, `cd stats && go test ./internal/api -run 'TestErrorStatusMapping|Dispatch' -count=1`, `cd stats && go test ./internal/reconcile -run 'DispatchEnd' -count=1` and `cd stats && make -s web-dist && go test ./cmd/flow -run 'RecordDispatch' -count=1` pass; `cd stats && gofmt -l .` prints nothing; `cd stats && make -s web-dist && go vet ./...` exits 0; `scripts/check-verbatim-moves.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/records/types.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/internal/api/server.go`, `stats/internal/api/changes_test.go`, `stats/internal/reconcile/reconcile.go`, `stats/internal/reconcile/record_test.go`, `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/cmd/flowd/main.go`, `skills/flow/implement.md`
**Tests:** `TestEndDispatchRefusesAnInlineTaskRowWithNoCommit`, `TestEndDispatchRefusesAnInlineRowBegunAfterItsCommit`, `TestRetimeDispatchSetsInstantsAndCommit`, `TestDispatchEndRefusedForItsCommitIsRetired`, `TestRecordDispatchEndSendsCommitAt`, `TestRecordDispatchBackfillDerivesInlineInstants`
**Regression:** the two store refusal tests fail — an inline task row closes with no commit, or with a begin recorded after its commit; the CLI tests fail — `end` sends no `commitAt` and an inline back-fill takes a hand-written instant.
**Baseline:** before=230 after=236 <!-- measured: cat stats/internal/store/records_test.go stats/cmd/flow/record_test.go stats/internal/reconcile/record_test.go | grep -c -E '^func Test' @ 75261910 → 229, 230 after Task 4's store test --> <!-- predicted: 236 after this task's six tests -->
**Commit:** feat(stats): refuse an inline row begun after its commit
**After:** Task 4

**Decision:** inline-commit-checks, inline-backfill-derives, retime-then-restamp
