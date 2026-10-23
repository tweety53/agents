# pipeline-fix-kan-1021-flow-improvement-move-the-bash-hook-and-setup-27

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `check-plan-shape.sh` refuses an absence sweep whose paths cover a later task's file that
still holds a swept string at the base (F18), so a misfire like kan-1021's Tasks 5 and 6 is caught
at plan time.

**Architecture:** one new finding in `stats/internal/guard/planshape.go`, sharing F15's base
resolution (extracted to `baseRev`); one case in `TestCheckPlanShape`; one sentence in
`skills/flow/brainstorm-planner.md` and its rationale bullet.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or other
  trailer; never write `spectre/changes/` (this change's `verbatim-moves.txt` is already committed
  with the plan).
- The task's change is the unified diff Step 2 gives, applied from the repository root with
  `git apply`; it was applied to this worktree at `e77f951b` and checked there. A diff that does
  not apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing.
- Read-only against the dev store: never stop flowd or flow-postgres, never drop the `flow` DB.

## Review Focus

- A ticked task is never read (`sweep-ticked-skipped`): kan-1021's archived plan, all ticked, exits
  0 under the new guard.
- A later file that does not hold a swept string at the base is not refused
  (`sweep-content-at-base`): no `absence sweep searches` line over
  `spectre/changes/archive/*/tasks.md`.
- A file the sweeping task owns itself is never reported, whoever else lists it.
- A plan outside a git repository, or whose base does not resolve, is not checked — F15's
  behaviour, unchanged by the `baseRev` extraction (cases 39 and 43 both pass).

- [x] 1. Refuse an absence sweep covering a later task's file

**Files:** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`, `skills/flow/brainstorm-planner.md`, `skills/flow/SKILL-rationale.md`
**Tests:** Case 43
**Regression:** Case 43 — reverting the commit drops F18, so the sweep over `pkg` covering task 2's `pkg/fold_fixup_test.go` passes and case 43's first assertion and its mutation proof fail.
**Baseline:** before=35 after=36
<!-- measured: grep -cE '// Case [0-9]+' stats/internal/guard/check_plan_shape_test.go @ branch pipeline-fix-kan-1021-flow-improvement-move-the-bash-hook-and-setup-27, before and after Step 2's diff -->
**After:** none
**Commit:** `fix(guard): refuse an absence sweep covering a later task's file`

**Decision:** sweep-scope-is-f18
**Decision:** sweep-content-at-base
**Decision:** sweep-ticked-skipped
**Decision:** sweep-shape-prescribed
**Decision:** base-rev-shared
**Decision:** sweep-scope-sentence

**Build:** green

  - [x] **Step 1: Reproduce before touching anything; do not guess.** From this worktree:
    `mkdir -p .superpowers/sdd && sed 's/^- \[x\]/- [ ]/' /Users/tweety53/Projects/agents-worktrees/kan-1021-flow-improvement-move-the-bash-hook-and-setup/spectre/changes/archive/kan-1021-flow-improvement-move-the-bash-hook-and-setup/tasks.md > .superpowers/sdd/kan-1021-unticked-tasks.md`
    (the path is gitignored), then `scripts/check-plan-shape.sh .superpowers/sdd/kan-1021-unticked-tasks.md`:
    exit 0 and no `absence sweep searches` line — the misfire passes today.
  - [x] **Step 2: Failing test.** Save the diff below to `.superpowers/sdd/pf27-task1.diff`, apply its
    test hunk alone (`git apply --include='stats/internal/guard/check_plan_shape_test.go' .superpowers/sdd/pf27-task1.diff`)
    and run `cd stats && make -s web-dist && go test ./internal/guard/ -run '^TestCheckPlanShape$' -count=1`:
    FAIL on `case 43 (F18): a sweep over a later task's file holding the string fails` and
    `case 43 (F18) mutation (mutation): the mutation applied`; every other case passes.

````diff verified:applied with git apply to this worktree at e77f951b; TestCheckPlanShape passes with it and fails exactly case 43's first assertion and mutation proof with planshape.go reverted; gofmt -l, go vet ./internal/guard/, check-references, check-markdown-integrity, check-installed-citations and check-verbatim-moves (with this change's verbatim-moves.txt) exit clean; the Step 1 call then exits 1 with four lines
diff --git a/skills/flow/SKILL-rationale.md b/skills/flow/SKILL-rationale.md
index e8ab74e6..55b90e25 100644
--- a/skills/flow/SKILL-rationale.md
+++ b/skills/flow/SKILL-rationale.md
@@ -157,6 +157,7 @@ text after it is the passage, verbatim.
 - ***A task premise about a guard's behaviour is run at plan time, never assumed.*** — kan-542's task 2 shows the cost of leaving the premise to review: planned on a stale premise, it survived until a reviewer struck it with the guard's own exit — a defect this duty kills at planning.
 - ***A task cites the decision it implements, never restates it.** Restated decision prose drifts from its entry the first time either is edited* — which is why gymie kan-468's seeded plan cited instead
 - ***End a removal task on an absence sweep.*** — (KAN-569) every kan-485 copy-removal task ended on a `git grep -n -F` over its cut strings in the shared sources, so each task's done was mechanical rather than a judgment call.
+- *Its paths never cover a later task's `**Files:**` while that file still holds a string the sweep removes* — (kan-1021) Tasks 5 and 6 swept `stats`, which covered Task 10's `<agents repo>/stats/internal/guard/pipelinefix.go` and `<agents repo>/stats/internal/guard/pipeline_fix_test.go`, still naming the removed harnesses, so each sweep printed hits at its own commit and needed a `Correction:`; Tasks 1–4's sweeps likewise hit doc comments the plan told them to keep.
 - *`**Baseline:**` … record a command whose stdout is one integer* — (the `| grep -c` pipelines in kan-271 and kan-298's plans are the shape)
 - *`**Files:**` … is expanded by the writer, never left in the field for the guard to expand* — (KAN-636); kan-579's plan declared gs/core/... behind its own legend and hand-repaired every field back to literal paths at its first task boundary, which is the repair this rule removes.
 - *A block of `**Decision:**` lines sits between blank lines … never glued to the `**Commit:**` subject above it* — (KAN-636): a citation with no blank line rode in the subject's continuation in every reader that had not learned the field's name, and kan-579's run repaired each such field by hand at its task boundary.
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index c39208a4..0d0bf7b4 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -380,6 +380,9 @@ paragraph under **The build-green tag** (`skills/flow-contracts/build-green.md`)
 > an identifier — its last verify step is one `git grep -n -F` over every string it removes, each
 > passed as its own `-e '<string>'`, scoped to the paths the task touches: the sweep prints
 > nothing when the task is done, and any hit means it is not.
+> Its paths never cover a later task's `**Files:**` while that file still holds a string the
+> sweep removes — `check-plan-shape.sh` refuses that sweep (F18) — nor a comment or doc line a
+> step of the plan tells the task to keep.
 >
 > A removed UI tag, test id or label that a declared `regression checkout` drives
 > (**Project configuration — visual verification**,
diff --git a/stats/internal/guard/check_plan_shape_test.go b/stats/internal/guard/check_plan_shape_test.go
index c197fbbb..d2dedf46 100644
--- a/stats/internal/guard/check_plan_shape_test.go
+++ b/stats/internal/guard/check_plan_shape_test.go
@@ -398,6 +398,26 @@ func TestCheckPlanShape(t *testing.T) {
 	r = one(chain("**Baseline:** before=1 after=2 "+m, "**Baseline:** before=1 after=2 <!-- measured: grep -c y g @ abc -->"))
 	want("case 42 (F17): tasks under different measured: commands pass", r.rc == 0, r)
 
+	// Case 43 (F18): an absence sweep covering a later task's file that
+	// still holds the swept string at the base (kan-1021's Tasks 5 and 6).
+	f18 := func(mark, paths string) string {
+		repo := psF15Repo(t)
+		md := repo + "/tasks.md"
+		writeFile(t, md, "- ["+mark+"] 1. Drop the fold\n\n**Files:** `pkg/fold.go`\n**Tests:** none\n**After:** none\n**Commit:** `fix(pkg): drop the fold`\n\n"+
+			"  - [ ] **Step 1: Absence sweep.** `git grep -n -F -e 'TestFoldFixupFolds' -- "+paths+"` prints nothing.\n\n"+
+			"- [ ] 2. Rewire the fold test\n\n**Files:** `pkg/fold_fixup_test.go`\n**Tests:** none — repairs existing tests: `TestFoldFixupFolds`\n**After:** Task 1\n**Commit:** `test(pkg): rewire the fold test`\n")
+		return md
+	}
+	md = f18(" ", "pkg")
+	r = one(md)
+	want("case 43 (F18): a sweep over a later task's file holding the string fails", r.rc == 1 &&
+		has(r, "task 1's absence sweep searches pkg/fold_fixup_test.go, which task 2 owns"), r)
+	mutation("F18", md, md, "absence sweep searches", "case 43 (F18) mutation")
+	r = one(f18(" ", "pkg/fold.go"))
+	want("case 43 (F18): a sweep scoped to the task's own Files passes", r.rc == 0, r)
+	r = one(f18("x", "pkg"))
+	want("case 43 (F18): a ticked task's sweep is not read", r.rc == 0, r)
+
 	// Unlanded archives: a shim copy with no stats/ beside it cannot build
 	// flow-guard; the same copy with no lib/flow-guard.sh cannot load it.
 	u := t.TempDir()
diff --git a/stats/internal/guard/planshape.go b/stats/internal/guard/planshape.go
index 480dfc82..7218fc54 100644
--- a/stats/internal/guard/planshape.go
+++ b/stats/internal/guard/planshape.go
@@ -14,10 +14,10 @@ import (
 )
 
 // check-plan-shape checks that a tasks.md is SHAPED so the guard that later
-// judges its commits (check-task-commit-fields) can actually read it. Seventeen
-// findings, F1-F17 (F1-F6's canonical definitions, decisions and rationale:
+// judges its commits (check-task-commit-fields) can actually read it. Eighteen
+// findings, F1-F18 (F1-F6's canonical definitions, decisions and rationale:
 // spectre/changes/kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan/
-// design.md -- not restated here; F7-F17's are the rows below).
+// design.md -- not restated here; F7-F18's are the rows below).
 //
 // The guard reads the plan with the very grammar the commit guard reads it
 // with -- tcfParseTask, tcfTaskIDs, tcfIterTasks, tcfSelectTask,
@@ -35,7 +35,7 @@ import (
 // its own real location -- so a run through the shipped
 // skills/flow/scripts/ symlink judges the repository, never skills/flow.
 //
-// Per file, the F1-F17 checks run, then plan-suite-steps
+// Per file, the F1-F18 checks run, then plan-suite-steps
 // (plansuitesteps.go: a task step running a module or repository suite),
 // plan-script-paths (planscriptpaths.go: a step citing a scripts/ path the
 // repository does not carry) and plan-task-id-rule (plantaskidrule.go: a
@@ -148,6 +148,13 @@ import (
 //	    partner's commit and is exempt. `<` rather than `!=`: a task between
 //	    the two may add tests to the same source set under another command
 //	    (kan-916, kan-939). Reported at the task's Baseline line.
+//	F18 An unticked task's absence sweep (a backticked `git grep ... --
+//	    <paths>` followed by `prints nothing`) whose paths cover a later
+//	    task's Files path the task does not own, while that file holds one
+//	    of the sweep's `-e '<string>'` strings at the base F15 resolves --
+//	    the sweep cannot print nothing at this task's commit (kan-1021's
+//	    Tasks 5 and 6). A plan whose base does not resolve is not checked.
+//	    Reported at the sweep's line, once per such file.
 //
 // A duplicate task id is check-task-build-green's violation, not this
 // guard's: tcfTaskIDs keeps the first occurrence, and every per-task read
@@ -200,7 +207,7 @@ func psAbs(env Env, p string) string {
 	return filepath.Join(env.Dir, p)
 }
 
-// psCheckOne checks one file: F1-F17, then the three plan sub-guards. off
+// psCheckOne checks one file: F1-F18, then the three plan sub-guards. off
 // names one finding whose decision is disabled -- the harness's mutation
 // proof, which sed-disabled that decision's `if` in a copy of the guard;
 // production passes "".
@@ -279,6 +286,10 @@ var (
 	psIdentRE = regexp.MustCompile(`^[A-Za-z_][\p{L}\p{N}_]*$`)
 	psTestPfx = regexp.MustCompile(`(?i)^test_?`)
 	psNonKey  = regexp.MustCompile(`[^a-z0-9]`)
+	// psSweepRE is F18's absence sweep: group 1 the git grep's options and
+	// strings, group 2 its paths. psSweepPatRE is one `-e '<string>'`.
+	psSweepRE    = regexp.MustCompile("`(git grep [^`]*?) -- ([^`]*)`[" + tcfWS + "]*prints nothing")
+	psSweepPatRE = regexp.MustCompile(`-e[` + tcfWS + `]+'([^']*)'`)
 )
 
 func psWord(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' }
@@ -350,7 +361,8 @@ func (p *psPlan) check() []string {
 	v = append(v, p.fileOwnership()...)
 	v = append(v, p.baselineConvention()...)
 	v = append(v, p.baselineChain()...)
-	return append(v, p.baseDeclared(ids)...)
+	v = append(v, p.baseDeclared(ids)...)
+	return append(v, p.sweepScope()...)
 }
 
 func (p *psPlan) checkTask(found tcfTaskBody) []string {
@@ -707,37 +719,45 @@ func psKey(name string) string {
 
 type psDecl struct{ file, name string }
 
-// baseDeclarations is F15's (file, name) test declarations in the
-// test-shaped files of the change's base tree; none when the plan sits
-// outside a git repository or no base resolves.
-func (p *psPlan) baseDeclarations() []psDecl {
+// baseRev is the change's base: the merge base of HEAD with
+// branch.<current>.flowBase (its origin/ copy preferred) or else with
+// refs/remotes/origin/HEAD, and run, a git call in the plan's repository
+// root. rev is empty when the plan sits outside a git repository or no base
+// resolves.
+func (p *psPlan) baseRev() (run func(args ...string) (string, bool), rev string) {
 	git := envGit(p.env)
-	run := func(dir string, args ...string) (string, bool) {
+	inDir := func(dir string, args ...string) (string, bool) {
 		out, err := git(append([]string{"-C", dir, "-c", "core.quotePath=false"}, args...)...).Output()
 		return string(out), err == nil
 	}
-	root, ok := run(psDir(filepath.Clean(psAbs(p.env, p.path))), "rev-parse", "--show-toplevel")
+	root, ok := inDir(psDir(filepath.Clean(psAbs(p.env, p.path))), "rev-parse", "--show-toplevel")
 	if !ok {
-		return nil
+		return nil, ""
 	}
 	root = tcfStrip(root)
-	branch, _ := run(root, "branch", "--show-current")
+	run = func(args ...string) (string, bool) { return inDir(root, args...) }
+	branch, _ := run("branch", "--show-current")
 	recorded := ""
 	if branch = tcfStrip(branch); branch != "" {
-		recorded, _ = run(root, "config", "--get", "branch."+branch+".flowBase")
+		recorded, _ = run("config", "--get", "branch."+branch+".flowBase")
 		recorded = tcfStrip(recorded)
 	}
 	candidates := []string{"refs/remotes/origin/HEAD"}
 	if recorded != "" {
 		candidates = []string{"refs/remotes/origin/" + recorded, recorded}
 	}
-	rev := ""
 	for _, ref := range candidates {
-		if out, ok := run(root, "merge-base", "HEAD", ref); ok {
-			rev = tcfStrip(out)
-			break
+		if out, ok := run("merge-base", "HEAD", ref); ok {
+			return run, tcfStrip(out)
 		}
 	}
+	return run, ""
+}
+
+// baseDeclarations is F15's (file, name) test declarations in the
+// test-shaped files of the change's base tree; none when no base resolves.
+func (p *psPlan) baseDeclarations() []psDecl {
+	run, rev := p.baseRev()
 	if rev == "" {
 		return nil
 	}
@@ -745,7 +765,7 @@ func (p *psPlan) baseDeclarations() []psDecl {
 	for _, g := range psTestShaped {
 		args = append(args, ":(glob)**/"+g)
 	}
-	out, _ := run(root, args...)
+	out, _ := run(args...)
 	var decls []psDecl
 	for _, record := range tcfSplitLines(out) {
 		where, text, _ := strings.Cut(record, "\x00")
@@ -760,6 +780,101 @@ func (p *psPlan) baseDeclarations() []psDecl {
 	return decls
 }
 
+// sweepScope is F18. A ticked task's sweep already ran, and a later commit
+// may have emptied the file, so it is not read. ponytail: one-line sweeps
+// and single-quoted -e strings only -- the shape
+// skills/flow/brainstorm-planner.md prescribes.
+func (p *psPlan) sweepScope() []string {
+	if !p.on("F18") {
+		return nil
+	}
+	tasks := tcfIterTasks(p.lines)
+	type sweep struct {
+		task        tcfTaskBody
+		line        int
+		paths, pats []string
+		own         []string
+	}
+	var sweeps []sweep
+	files := map[string][]string{}
+	for _, task := range tasks {
+		if tcfUnclosedFence(task.lines) >= 0 {
+			continue // F3b: this task's fields are unread
+		}
+		files[task.id] = p.parse(task.id).files
+		if tcfTaskLineRE.FindStringSubmatch(p.lines[task.taskLine-1])[1] == "x" {
+			continue
+		}
+		tcfUnfenced(task.lines, func(off int, l string) bool {
+			for _, m := range psSweepRE.FindAllStringSubmatch(l, -1) {
+				var pats []string
+				for _, e := range psSweepPatRE.FindAllStringSubmatch(m[1], -1) {
+					pats = append(pats, e[1])
+				}
+				if len(pats) > 0 {
+					sweeps = append(sweeps, sweep{task, task.taskLine + off + 1, strings.FieldsFunc(m[2], tcfIsSpace), pats, files[task.id]})
+				}
+			}
+			return false
+		})
+	}
+	if len(sweeps) == 0 {
+		return nil
+	}
+	run, rev := p.baseRev()
+	if rev == "" {
+		return nil
+	}
+	covers := func(paths []string, f string) bool {
+		return slices.ContainsFunc(paths, func(s string) bool {
+			s = strings.TrimSuffix(s, "/")
+			return s == "." || s == ":/" || f == s || strings.HasPrefix(f, s+"/")
+		})
+	}
+	var v []string
+	for _, s := range sweeps {
+		owner := map[string]string{}
+		later := false
+		for _, task := range tasks {
+			if task.id == s.task.id {
+				later = true
+				continue
+			}
+			if !later {
+				continue
+			}
+			for _, f := range files[task.id] {
+				if _, ok := owner[f]; !ok && !slices.Contains(s.own, f) && covers(s.paths, f) {
+					owner[f] = task.id
+				}
+			}
+		}
+		if len(owner) == 0 {
+			continue
+		}
+		args := []string{"grep", "-l", "-F"}
+		for _, pat := range s.pats {
+			args = append(args, "-e", pat)
+		}
+		args = append(args, rev, "--")
+		for f := range owner {
+			args = append(args, f)
+		}
+		out, _ := run(args...)
+		var held []string
+		for _, l := range tcfSplitLines(out) {
+			if f := strings.TrimPrefix(l, rev+":"); owner[f] != "" {
+				held = append(held, f)
+			}
+		}
+		sort.Strings(held)
+		for _, f := range held {
+			v = append(v, p.at(s.line, "task "+s.task.id+"'s absence sweep searches "+f+", which task "+owner[f]+" owns and which still holds a swept string at the base — the sweep cannot print nothing at this task's commit; scope it to the paths this task touches"))
+		}
+	}
+	return v
+}
+
 // baseDeclared is F15.
 func (p *psPlan) baseDeclared(ids []string) []string {
 	decls := p.baseDeclarations()
````

  - [x] **Step 3: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/check_plan_shape_test.go' .superpowers/sdd/pf27-task1.diff`)
    and rerun Step 2's test: PASS.
  - [x] **Step 4: Measure.** Step 1's call exits 1 with exactly four lines — task 5's and task 6's
    sweeps, each naming `stats/internal/guard/pipelinefix.go` and
    `stats/internal/guard/pipeline_fix_test.go` as task 10's. `scripts/check-plan-shape.sh /Users/tweety53/Projects/agents-worktrees/kan-1021-flow-improvement-move-the-bash-hook-and-setup/spectre/changes/archive/kan-1021-flow-improvement-move-the-bash-hook-and-setup/tasks.md` exits 0
    (every task ticked). `for f in spectre/changes/archive/*/tasks.md; do scripts/check-plan-shape.sh "$f"; done | grep -c 'absence sweep searches'`
    prints `0`.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/` prints nothing;
    `scripts/check-plan-shape.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`,
    `scripts/check-installed-citations.sh` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 6: Commit** `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go`,
    `skills/flow/brainstorm-planner.md` and `skills/flow/SKILL-rationale.md` with the subject above.
