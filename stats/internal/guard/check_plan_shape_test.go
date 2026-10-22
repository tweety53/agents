package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestCheckPlanShape ports scripts/test-check-plan-shape.sh at bbfe685e: one
// subtest per ok: label, the guard run in-process against real fixture files
// (and real git repositories for the base-reading cases). The harness's
// mutation cases sed-disabled one finding's `# F<n>`-tagged decision in a
// copy of check-plan-shape.py; here psCheckOne's off argument disables that
// finding's decision, and "the mutation applied" is proved by the mutant's
// output differing from the guard's on a fixture where the finding fires.
// The harness's three wrapper cases — a missing sibling, the symlinked
// install path, a flow-guard that cannot be built — exec the real shim.
func TestCheckPlanShape(t *testing.T) {
	t.Parallel()
	type check struct{ label, fail string }
	var checks []check
	want := func(label string, ok bool, r guardResult) {
		c := check{label: label}
		if !ok {
			c.fail = fmt.Sprintf("rc=%d out=%q", r.rc, r.out)
		}
		checks = append(checks, c)
	}
	has := func(r guardResult, s string) bool { return strings.Contains(r.out, s) }
	env := Env{Getenv: os.Getenv}
	one := func(path string) guardResult { return runGuard("check-plan-shape", []string{path}, env) }
	scan := func(root string) guardResult {
		return runGuard("check-plan-shape", nil, Env{Getenv: psGetenv(map[string]string{"CHECK_PLAN_SHAPE_ROOT": root})})
	}
	// mutation is assert_mutation_removes_finding: off disables the finding
	// tagged off; proof is a fixture on which that finding fires, so a
	// differing output shows the disable reached the decision.
	mutation := func(off, fixture, proof, msg, label string) {
		m := psMutant(proof, off)
		want(label+" (mutation): the mutation applied", m.out != one(proof).out, m)
		m = psMutant(fixture, off)
		want(label+" (mutation): disabling the check removes the finding", !has(m, msg), m)
	}

	// Case 1: a clean plan -> exit 0, no output.
	md := psFixture(t, "- [ ] 1. A clean task\n\n**Files:** `scripts/foo.sh`\n**Tests:** `scripts/test-foo.sh`\n**Commit:** `feat(scripts): add foo`\n")
	r := one(md)
	want("case 1: a clean plan passes", r.rc == 0, r)
	want("case 1: no output", r.out == "", r)

	// Case 2 (F1): a second **Files:** line, at physical line 7.
	md = psFixture(t, "- [ ] 1. Duplicate Files\n\n**Files:** `a.py`\n\n**Tests:** `test-a.sh`\n\n**Files:** `b.py`\n")
	r = one(md)
	want("case 2 (F1): duplicate Files line fails", r.rc == 1, r)
	want("case 2 (F1): names the task and the field", has(r, "task 1 declares a second **Files:**"), r)
	want("case 2 (F1): the exact file:line prefix is line 7", strings.HasPrefix(r.out, md+":7: task 1 declares a second"), r)
	mutation("F1", md, md, "declares a second **Files:**", "case 2")

	// Case 3 (F2): no **Files:** line at all.
	md = psFixture(t, "- [ ] 1. No Files line\n\n**Tests:** `test-a.sh`\n")
	r = one(md)
	want("case 3 (F2): no Files line fails", r.rc == 1, r)
	want("case 3 (F2): names the task", has(r, "task 1 has no **Files:** line"), r)
	mutation("F2", md, md, "has no **Files:** line", "case 3")

	// Case 4 (F2): a **Files:** line present but empty.
	r = one(psFixture(t, "- [ ] 1. Empty Files line\n\n**Files:**\n\n**Tests:** `test-a.sh`\n"))
	want("case 4 (F2): an empty Files line fails", r.rc == 1, r)
	want("case 4 (F2): names the task", has(r, "task 1 has no **Files:** line"), r)

	// Case 5 (F3a): an indented **Files:** line, at physical line 3.
	md = psFixture(t, "- [ ] 1. Indented Files line\n\n  **Files:** `a.py`\n\n**Tests:** `test-a.sh`\n")
	r = one(md)
	want("case 5 (F3a): an indented Files line fails", r.rc == 1, r)
	want("case 5 (F3a): names the indentation defect", has(r, "indented past column 0"), r)
	want("case 5 (F3a): the exact file:line prefix is line 3", strings.HasPrefix(r.out, md+":3: task 1 carries a"), r)
	mutation("F3a", md, md, "indented past column 0", "case 5")

	// Case 6 (F3b): an unclosed fence names the fence, and suppresses F2/F6.
	md = psFixture(t, "- [ ] 1. Unclosed fence\n\n```\n**Files:** `a.py`\n")
	r = one(md)
	want("case 6 (F3b): an unclosed fence fails", r.rc == 1, r)
	want("case 6 (F3b): names the unclosed fence", has(r, "never closed in its body"), r)
	want("case 6 (F3b): exactly one violation line is printed", !strings.Contains(r.out, "\n"), r)
	want("case 6 (F3b): F2 is suppressed", !has(r, "has no **Files:** line"), r)
	want("case 6 (F3b): F6 is suppressed", !has(r, "declares nothing"), r)
	mutation("F3b", md, md, "never closed in its body", "case 6")

	// Case 7 (F4): `none` Tests over a test-shaped Files entry.
	f4 := psFixture(t, "- [ ] 1. None with a test-shaped file\n\n**Files:** `scripts/test-check-foo.sh`\n\n**Tests:** none added — nothing to verify\n")
	r = one(f4)
	want("case 7 (F4): none-opening Tests with a test-shaped Files entry fails", r.rc == 1, r)
	want("case 7 (F4): names the contradiction", has(r, "contradiction between the opt-out"), r)
	mutation("F4", f4, f4, "contradiction between the opt-out", "case 7")

	// Case 8: `none` over an ordinary file — neither F4 nor F6.
	r = one(psFixture(t, "- [ ] 1. None with an ordinary file\n\n**Files:** `scripts/foo.sh`\n\n**Tests:** none added — nothing to verify\n"))
	want("case 8: none-opening Tests with an ordinary Files entry passes", r.rc == 0, r)

	// Case 9 (F5): zero tasks.
	md = psFixture(t, "# A plan with no tasks\n\nJust prose, no checkbox lines at all.\n")
	r = one(md)
	want("case 9 (F5): zero tasks fails", r.rc == 1, r)
	want("case 9 (F5): names the vacuous-pass defect", has(r, "zero tasks"), r)
	mutation("F5", md, md, "zero tasks", "case 9")

	// Case 10 (F6): a Tests field that declares nothing.
	md = psFixture(t, "- [ ] 1. Vague Tests field\n\n**Files:** `a.py`\n\n**Tests:** the existing harnesses, unchanged, plus a case pinning the fix\n")
	r = one(md)
	want("case 10 (F6): a vague Tests field fails", r.rc == 1, r)
	want("case 10 (F6): names the vacuous field", has(r, "declares nothing"), r)
	mutation("F6", md, md, "declares nothing", "case 10")

	// Cases 11, 11b, 12: a Case N label, a lettered one, a backtick token.
	r = one(psFixture(t, "- [ ] 1. Case-labeled Tests field\n\n**Files:** `a.py`\n\n**Tests:** Case 3 pins the new behaviour\n"))
	want("case 11: a Case N label satisfies Tests:", r.rc == 0, r)
	r = one(psFixture(t, "- [ ] 1. Lettered case-labeled Tests field\n\n**Files:** `a.py`\n\n**Tests:** Case 13b pins the new behaviour\n"))
	want("case 11b: a lettered Case label satisfies Tests:", r.rc == 0, r)
	r = one(psFixture(t, "- [ ] 1. Backtick Tests field\n\n**Files:** `a.py`\n\n**Tests:** `scripts/test-a.sh`\n"))
	want("case 12: a backtick token satisfies Tests:", r.rc == 0, r)

	// Case 13: a duplicate **Files:** inside a fence is no declaration.
	r = one(psFixture(t, "- [ ] 1. Fenced example, not a real duplicate\n\n**Files:** `a.py`\n\n**Tests:** `test-a.sh`\n\nExample of the field grammar:\n\n```\n**Files:** `not-a-real-declaration.py`\n```\n"))
	want("case 13: a fenced example is not read as a second Files line", r.rc == 0, r)

	// Case 14: the no-argument scan over CHECK_PLAN_SHAPE_ROOT; the fixture
	// is no git worktree, so its archived change stays out.
	root := t.TempDir()
	writeFile(t, root+"/spectre/changes/clean-change/tasks.md", "- [ ] 1. Clean\n\n**Files:** `a.py`\n\n**Tests:** `test-a.sh`\n")
	writeFile(t, root+"/spectre/changes/dirty-change/tasks.md", "- [ ] 1. Dirty\n\n**Tests:** `test-a.sh`\n")
	writeFile(t, root+"/spectre/changes/archive/archived-change/tasks.md", "- [ ] 1. Archived, must not be scanned\n\n")
	r = scan(root)
	want("case 14: no-arg scan aggregates a violation from the dirty file", r.rc == 1, r)
	want("case 14: names the dirty file", has(r, "dirty-change/tasks.md:") && has(r, "has no **Files:** line"), r)
	want("case 14: the clean file reports nothing", !has(r, "clean-change"), r)
	want("case 14: archived changes are excluded from the no-arg scan", !has(r, "archived-change"), r)

	// Case 14b: an archived change joins the scan only while the base lacks it.
	root = t.TempDir()
	dirty := "- [ ] 1. Archived, no Files line\n\n**Tests:** `test-a.sh`\n"
	writeFile(t, root+"/spectre/changes/archive/landed-change/tasks.md", dirty)
	gitRun(t, root, "init", "-q", "-b", "main")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-qm", "chore(spectre): archive landed-change")
	writeFile(t, root+"/spectre/changes/archive/new-change/tasks.md", dirty)
	r = scan(root)
	want("case 14b: an archived change the base lacks is scanned", r.rc == 1, r)
	want("case 14b: names the unlanded archived change's violation", has(r, "archive/new-change/tasks.md:") && has(r, "has no **Files:** line"), r)
	want("case 14b: an archived change the base carries stays out of the scan", !has(r, "landed-change"), r)

	// Cases 15, 16: usage and an unreadable file exit 2.
	r = runGuard("check-plan-shape", []string{"a", "b"}, env)
	want("case 15: too many arguments exits 2", r.rc == 2, r)
	r = one(t.TempDir() + "/does-not-exist.md")
	want("case 16: a nonexistent file exits 2", r.rc == 2, r)

	// Case 17: the shim's own sibling missing exits 2 naming it — the
	// sibling is lib/flow-guard.sh now that no Python module is loaded.
	clean := psFixture(t, "- [ ] 1. Clean task\n\n**Files:** `a.py`\n\n**Tests:** `test-a.sh`\n")
	stripped := t.TempDir()
	psCopy(t, psShim, stripped+"/check-plan-shape.sh")
	r = psExec(t, nil, stripped+"/check-plan-shape.sh", clean)
	want("case 17: a shim copy with no lib/flow-guard.sh sibling exits 2", r.rc == 2, r)
	want("case 17: the message names the missing module", has(r, "lib/flow-guard.sh"), r)

	// Case 18: run through a skills/flow/scripts/ symlink, the no-argument
	// scan resolves the root to the sandbox, not to skills/flow.
	sandbox := t.TempDir()
	psCopy(t, psShim, sandbox+"/scripts/check-plan-shape.sh")
	psCopy(t, psScripts+"/lib/flow-guard.sh", sandbox+"/scripts/lib/flow-guard.sh")
	psLink(t, filepath.Dir(psScripts)+"/stats", sandbox+"/stats")
	psLink(t, "../../../scripts/check-plan-shape.sh", sandbox+"/skills/flow/scripts/check-plan-shape.sh")
	psLink(t, "../../../scripts/lib", sandbox+"/skills/flow/scripts/lib")
	writeFile(t, sandbox+"/spectre/changes/dirty-change/tasks.md", "- [ ] 1. Dirty\n\n**Tests:** `test-a.sh`\n")
	r = psExec(t, []string{"FLOW_GUARD_CACHE_DIR=" + guardCache(t)}, sandbox+"/skills/flow/scripts/check-plan-shape.sh")
	want("case 18: the symlinked invocation resolves REPO_ROOT to the sandbox root and finds the dirty fixture", r.rc == 1, r)
	want("case 18: names the dirty fixture", has(r, "dirty-change/tasks.md:") && has(r, "has no **Files:** line"), r)

	// Case 19 (F4/F6): `none` opening on a continuation line.
	r = one(psFixture(t, "- [ ] 1. Continuation-line none, test-shaped file\n\n**Files:** `scripts/test-check-foo.sh`\n\n**Tests:**\nnone added — nothing to verify\n"))
	want("case 19 (F4): a continuation-line none over a test-shaped Files entry fails", r.rc == 1, r)
	want("case 19 (F4): names the contradiction, not F6", has(r, "contradiction between the opt-out"), r)
	r = one(psFixture(t, "- [ ] 1. Continuation-line none, ordinary file\n\n**Files:** `scripts/foo.sh`\n\n**Tests:**\nnone added — nothing to verify\n"))
	want("case 19: a continuation-line none over an ordinary Files entry passes", r.rc == 0, r)

	// Case 20 (F7): two gating **After:** lines.
	r = one(psFixture(t, "- [ ] 1. Duplicate After\n\n**Files:** `a.txt`\n**Tests:** `test_a`\n**Commit:** add a\n**After:** Task 1\n**After:** Task 1\n"))
	want("case 20 (F7): two gating After lines fail", r.rc == 1, r)
	want("case 20 (F7): names the second line and the first's line", has(r, "task 1 declares a second **After:** line (first at line ") && has(r, ")"), r)

	// Case 21 (F8): a free-text **After:** value.
	r = one(psFixture(t, "- [ ] 1. Malformed After\n\n**Files:** `a.txt`\n**Tests:** `test_a`\n**Commit:** add a\n**After:** Task sometime\n"))
	want("case 21 (F8): a non-gating After value fails", r.rc == 1, r)
	want("case 21 (F8): names the offending value", has(r, "task 1 has a malformed **After:** value: Task sometime"), r)
	want("case 21 (F8): names the accepted values", has(r, "`Task <ids>` or `none`"), r)

	// Case 22 (F9): a dangling **After:** id.
	r = one(psFixture(t, psTasks("a", "b")+"**After:** Task 9\n"))
	want("case 22 (F9): a dangling After id fails", r.rc == 1, r)
	want("case 22 (F9): names the absent task", has(r, "task 2's **After:** names task 9, which this plan does not define"), r)

	// Case 23 (F10): a serial-default cycle, and a self-reference.
	r = one(psFixture(t, "- [ ] 1. First task\n\n**Files:** `a.txt`\n**Tests:** `test_a`\n**Commit:** add a\n\n"+
		"- [ ] 2. Second task\n\n**Files:** `b.txt`\n**Tests:** `test_b`\n**Commit:** add b\n**After:** Task 5\n\n"+
		"- [ ] 3. Third task\n\n**Files:** `c.txt`\n**Tests:** `test_c`\n**Commit:** add c\n\n"+
		"- [ ] 4. Fourth task\n\n**Files:** `d.txt`\n**Tests:** `test_d`\n**Commit:** add d\n\n"+
		"- [ ] 5. Fifth task\n\n**Files:** `e.txt`\n**Tests:** `test_e`\n**Commit:** add e\n"))
	want("case 23 (F10): the serial-default cycle fails", r.rc == 1, r)
	want("case 23 (F10): names both tasks of the closed cycle", has(r, "task 2 -> task 5 -> task 2: the resolved After graph has a cycle"), r)
	r = one(psFixture(t, "- [ ] 1. First task\n\n**Files:** `a.txt`\n**Tests:** `test_a`\n**Commit:** add a\n\n"+
		"- [ ] 2. Self-referencing task\n\n**Files:** `b.txt`\n**Tests:** `test_b`\n**Commit:** add b\n**After:** Task 2\n"))
	want("case 23 (F10): a self-reference fails", r.rc == 1, r)
	want("case 23 (F10): the one-node cycle is reported", has(r, "task 2 -> task 2: the resolved After graph has a cycle"), r)

	// Case 24: `Task <ids>` and `none` declarations, clean.
	r = one(psFixture(t, "- [ ] 1. Independent task\n\n**Files:** `a.txt`\n**Tests:** `test_a`\n**Commit:** add a\n**After:** none\n\n"+
		"- [ ] 2. Depends on task 1\n\n**Files:** `b.txt`\n**Tests:** `test_b`\n**Commit:** add b\n**After:** Task 1\n\n"+
		"- [ ] 3. Depends on tasks 1 and 2\n\n**Files:** `c.txt`\n**Tests:** `test_c`\n**Commit:** add c\n**After:** Task 1, 2\n"))
	want("case 24: a clean After-using plan passes", r.rc == 0, r)

	// Case 25: no **After:** fields at all.
	r = one(psFixture(t, psTasks("a", "b")))
	want("case 25: a fieldless plan passes", r.rc == 0, r)
	want("case 25: output byte-identical to before this task", r.out == "", r)

	// Case 26 (F11): one file, two unordered owners, at task 2's Files line 9.
	md = psFixture(t, psTasks("a", "a"))
	r = one(md)
	want("case 26 (F11): a file two unordered tasks own fails", r.rc == 1, r)
	want("case 26 (F11): names the path, both tasks, and the missing ordering",
		has(r, "task 2's **Files:** also names `a.txt`") && has(r, "task 1") && has(r, "no **After:** ordering"), r)
	want("case 26 (F11): the exact file:line prefix is line 9", strings.HasPrefix(r.out, md+":9: task 2's **Files:** also names"), r)
	mutation("F11", md, md, "no **After:** ordering", "case 26")

	// Case 27 (F11): the declared edge passes.
	r = one(psFixture(t, psTasks("a", "a")+"**After:** Task 1\n"))
	want("case 27 (F11): a declared edge over the shared file passes", r.rc == 0, r)
	want("case 27 (F11): no output", r.out == "", r)

	// Case 28 (F11): `none` is independence, not an ordering.
	r = one(psFixture(t, psTasks("a", "a")+"**After:** none\n"))
	want("case 28 (F11): After: none over a shared file fails", r.rc == 1, r)
	want("case 28 (F11): the none declaration is not an ordering", has(r, "no **After:** ordering"), r)

	// Case 29 (F11): a task whose fence never closes is no owner.
	r = one(psFixture(t, psTasks("a", "a")+"\n```\nunterminated\n"))
	want("case 29 (F11): the unclosed-fence task fails on F3b", r.rc == 1, r)
	want("case 29 (F11): exactly one violation line — the fence, no F11", !strings.Contains(r.out, "\n"), r)
	want("case 29 (F11): the fenced owner is skipped", !has(r, "no **After:** ordering"), r)

	// Case 30 (F11): a repeated token within one task is one owner.
	r = one(psFixture(t, "- [ ] 1. Dup token\n\n**Files:** `a.txt`, `a.txt`\n**Tests:** `test_a`\n**Commit:** add a\n"))
	want("case 30 (F11): a duplicate token in one task is not a pair", r.rc == 0, r)
	want("case 30 (F11): no output", r.out == "", r)

	// Cases 31-34 (KAN-794): the repair form, and its negative shapes.
	repair := func(tests string) guardResult {
		return one(psFixture(t, "- [ ] 1. Repair\n\n**Files:** `scripts/test-check-foo.sh`\n\n**Tests:**"+tests+"\n"))
	}
	r = repair(" none — repairs existing tests: the fixture-collision cases")
	want("case 31: the repair form over a test-shaped Files entry passes", r.rc == 0, r)
	want("case 31: no output", r.out == "", r)
	r = repair("\nnone — repairs existing tests: the fixture-collision cases")
	want("case 32: a continuation-line repair form over a test-shaped Files entry passes", r.rc == 0, r)
	want("case 32: no output", r.out == "", r)
	r = repair(" none added — repairs existing tests: the fixture-collision cases")
	want("case 33: prose before the repair phrase still fails F4", r.rc == 1, r)
	want("case 33: names the contradiction", has(r, "contradiction between the opt-out"), r)
	r = repair(" none — repairs existing tests the fixture-collision cases")
	want("case 34: the repair phrase without its colon still fails F4", r.rc == 1, r)
	want("case 34: names the contradiction", has(r, "contradiction between the opt-out"), r)

	// plan-suite-steps, in-process, in both scans.
	suite := "- [ ] 1. Menu\n\n**Files:** `A.kt`\n\n**Tests:** `aRowIsTall`\n\n  - [ ] **Step 1:** run `./gradlew :shared:desktopTest`; migrate failures\n"
	r = one(psFixture(t, suite))
	want("plan-suite-steps: a module-suite step fails (one argument)", r.rc == 1 && has(r, "task 1 runs the suite `./gradlew :shared:desktopTest`"), r)
	root = t.TempDir()
	writeFile(t, root+"/spectre/changes/c/tasks.md", suite)
	r = scan(root)
	want("plan-suite-steps: a module-suite step fails (no-argument scan)", r.rc == 1 && has(r, "task 1 runs the suite"), r)
	r = one(psFixture(t, strings.Replace(suite, "`;", " --tests 'X'`;", 1)))
	want("plan-suite-steps: the same step with --tests passes", r.rc == 0, r)

	// Case 35 (F12): Baseline fields and no convention line.
	f12 := psFixture(t, "> **Execution:** `/flow` implements this plan.\n\n> **Relocation:** no\n\n- [ ] 1. Add a test\n\n"+
		"**Files:** `src/app/profile/src/test/CounterTest.kt`\n**Tests:** `counter renders`\n**Baseline:** before=3 after=4\n**Commit:** `feat(profile): add counter test`\n")
	r = one(f12)
	want("case 35 (F12): a Baseline plan with no convention line fails", r.rc == 1, r)
	want("case 35 (F12): names the missing convention", has(r, "states no **Baseline convention:** header line"), r)
	mutation("F12", f12, f12, "Baseline convention", "case 35 (F12) mutation")

	// Case 36 (F12 absent): no Baseline fields, no convention needed.
	md = psFixture(t, "- [ ] 1. A plan without Baselines\n\n**Files:** `scripts/foo.sh`\n**Tests:** `scripts/test-foo.sh`\n**Commit:** `feat(scripts): add foo`\n")
	r = one(md)
	want("case 36 (F12 absent): a Baseline-free plan needs no convention line", r.rc == 0, r)
	want("case 36 (F12 absent): no output", r.out == "", r)
	mutation("F12", md, f12, "Baseline convention", "case 36 (F12 absent) mutation")

	// Case 37 (F13): a measured: comment counting `@Test` unsafely.
	f13 := func(cmd string) string {
		return psFixture(t, "> **Baseline convention:** chain-position absolutes.\n\n- [ ] 1. Add a test\n\n"+
			"**Files:** `src/app/profile/src/test/CounterTest.kt`\n**Tests:** `counter renders`\n"+
			"**Baseline:** before=3 after=4 <!-- measured: "+cmd+" @ HEAD -->\n**Commit:** `feat(profile): add counter test`\n")
	}
	const kt = "src/app/profile/src/test/CounterTest.kt"
	counts := "counts `@Test` with `grep -c` or `grep -o`"
	md = f13("grep -c '@Test' " + kt)
	r = one(md)
	want("case 37 (F13): grep -c '@Test' fails", r.rc == 1 && has(r, "task 1's measured: command "+counts+" and no `-w`"), r)
	mutation("F13", md, md, "counts `@Test` with `grep -c`", "case 37 (F13) mutation")
	r = one(f13("grep -cw '@Test' " + kt))
	want("case 37 (F13): grep -cw '@Test' passes", r.rc == 0, r)
	r = one(f13("grep -o '@Test' " + kt + " | wc -l"))
	want("case 37 (F13): grep -o '@Test' | wc -l fails", r.rc == 1 && has(r, counts), r)
	r = one(f13("grep -ow '@Test' " + kt + " | wc -l"))
	want("case 37 (F13): grep -ow '@Test' | wc -l passes", r.rc == 0, r)
	r = one(f13("grep -c --include=*.kt '@Test' " + kt))
	want("case 37 (F13): grep -c --include=*.kt '@Test' fails", r.rc == 1 && has(r, counts), r)
	r = one(f13("grep -c -m1 --include=*w* '@Test' " + kt))
	want("case 37 (F13): grep -c -m1 --include=*w* '@Test' fails", r.rc == 1 && has(r, counts), r)
	r = one(f13("grep -cw --include=*.kt '@Test' " + kt))
	want("case 37 (F13): grep -cw --include=*.kt '@Test' passes", r.rc == 0, r)

	// Case 38 (F14): Files all under the planning paths.
	f14 := func(second string) string {
		return psFixture(t, "- [ ] 1. Sweep the frames\n\n**Files:** `spectre/changes/x/visual-sweep.md`, `"+second+"`\n**Tests:** none\n**Commit:** `docs(spectre): sweep`\n")
	}
	md = f14("docs/superpowers/notes.md")
	r = one(md)
	want("case 38 (F14): a planning-paths-only task fails", r.rc == 1 && has(r, "task 1's **Files:** all lie under the planning paths"), r)
	mutation("F14", md, md, "all lie under the planning paths", "case 38 (F14) mutation")
	r = one(f14("docs/visual-sweep.md"))
	want("case 38 (F14): one implementation path passes", r.rc == 0, r)

	// Case 39 (F15): a Tests name the base already declares outside Files.
	f15 := func(files, tests string, more bool) string {
		repo := psF15Repo(t)
		if more {
			writeFile(t, repo+"/pkg/old_test.go", "package pkg\n\nfunc TestFoo(t *testing.T) {}\n")
			gitRun(t, repo, "add", "pkg")
			gitRun(t, repo, "-c", "core.hooksPath=/dev/null", "commit", "-qm", "more")
			gitRun(t, repo, "branch", "-f", "main", "HEAD")
		}
		md := repo + "/tasks.md"
		writeFile(t, md, "- [ ] 1. Port guard-autosquash\n\n**Files:** "+files+"\n**Tests:** `TestAutosquash`, `"+tests+"`\n**Commit:** `feat(pkg): port guard-autosquash`\n")
		return md
	}
	own := "`pkg/autosquash.go`, `pkg/autosquash_test.go`"
	md = f15(own, "TestFoldFixup", false)
	r = one(md)
	want("case 39 (F15): an existing test family outside Files fails", r.rc == 1 &&
		has(r, "names `TestFoldFixup`, which the base already declares outside its **Files:** — pkg/fold_fixup_test.go (`TestFoldFixupFolds`)"), r)
	mutation("F15", md, md, "which the base already declares", "case 39 (F15) mutation")
	r = one(f15(own+", `pkg/fold_fixup_test.go`", "TestFoldFixup", false))
	want("case 39 (F15): the family's file in Files passes", r.rc == 0, r)
	r = one(f15(own+", `pkg/fold_fixup_test.go`", "TestFoo", false))
	want("case 39 (F15): a new TestFoo beside TestFooBar in another file passes", r.rc == 0, r)
	r = one(f15(own+", `pkg/foo_bar_test.go`", "TestFoo", false))
	want("case 39 (F15): a new TestFoo beside TestFooBar in its own file passes", r.rc == 0, r)
	r = one(f15(own+", `pkg/foo_bar_test.go`", "TestFoo", true))
	want("case 39 (F15): an exact name outside Files fails", r.rc == 1 &&
		has(r, "names `TestFoo`, which the base already declares outside its **Files:** — pkg/old_test.go (`TestFoo`)"), r)

	// Case 40 (F16): unbackticked multi-path values.
	md = psFixture(t, "- [ ] 1. One\n\n**Files:** a.py, b.py\n**Tests:** `test_a`\n**After:** none\n**Commit:** `fix: one`\n\n"+
		"- [ ] 2. Two\n\n**Files:** `c.py`\n**Allowed-collateral:** docs/a.md docs/b.md\n**Tests:** `test_c`\n**After:** none\n**Commit:** `fix: two`\n\n"+
		"- [ ] 3. Three\n\n**Files:** d.py\n**Tests:** `test_d`\n**After:** none\n**Commit:** `fix: three`\n\n"+
		"- [ ] 4. Four\n\n**Files:** `e.py`\n**Allowed-collateral:** none, unless step 1 says otherwise\n**Tests:** `test_e`\n**After:** none\n**Commit:** `fix: four`\n")
	r = one(md)
	i := strings.Index(r.out, "task 1's **Files:** value names no backticked path")
	want("case 40 (F16): unbackticked multi-path values fail", r.rc == 1 && i >= 0 &&
		strings.Contains(r.out[i:], "task 2's **Allowed-collateral:** value names no backticked path"), r)
	want("case 40 (F16): a single unbackticked path and a none sentence pass", !has(r, "task 3") && !has(r, "task 4"), r)
	mutation("F16", md, md, "names no backticked path", "case 40")

	// Case 41 (F17): a Baseline chain falling back under one measured: command.
	chain := func(t1, t2 string) string {
		return psFixture(t, "# plan\n\n> **Baseline convention:** chain-position absolutes.\n\n"+
			"- [ ] 1. One\n\n**Files:** `a_test.go`\n**Tests:** `TestA`\n"+t1+"\n**After:** none\n**Commit:** `fix: one`\n\n"+
			"- [ ] 2. Two\n\n**Files:** `b_test.go`\n**Tests:** `TestB`\n"+t2+"\n**After:** Task 1\n**Commit:** `fix: two`\n")
	}
	m := "<!-- measured: grep -c x f @ abc -->"
	md = chain("**Baseline:** before=1 after=2 "+m, "**Baseline:** before=1 after=2 "+m)
	r = one(md)
	want("case 41 (F17): a chain falling back mid-plan fails", r.rc == 1 && has(r, "task 2's **Baseline:** before=1 is below task 1's after=2"), r)
	mutation("F17", md, md, "is below task", "case 41")

	// Case 42 (F17 absent): a continuing chain, a Squash-with pair, two commands.
	r = one(chain("**Baseline:** before=1 after=2 "+m, "**Baseline:** before=2 after=3 "+m))
	want("case 42 (F17): a continuing chain passes", r.rc == 0, r)
	r = one(chain("**Build:** red\n**Squash-with:** Task 2\n**Baseline:** before=1 after=2 "+m, "**Baseline:** before=1 after=2 "+m))
	want("case 42 (F17): a Squash-with pair sharing one pair passes", r.rc == 0, r)
	r = one(chain("**Baseline:** before=1 after=2 "+m, "**Baseline:** before=1 after=2 <!-- measured: grep -c y g @ abc -->"))
	want("case 42 (F17): tasks under different measured: commands pass", r.rc == 0, r)

	// Unlanded archives: a shim copy with no stats/ beside it cannot build
	// flow-guard; the same copy with no lib/flow-guard.sh cannot load it.
	u := t.TempDir()
	mkdir(t, u+"/root/spectre/changes/some-change")
	psCopy(t, psShim, u+"/scripts/check-plan-shape.sh")
	psCopy(t, psScripts+"/lib/flow-guard.sh", u+"/scripts/lib/flow-guard.sh")
	uenv := []string{"CHECK_PLAN_SHAPE_ROOT=" + u + "/root", "FLOW_GUARD_CACHE_DIR=" + t.TempDir()}
	r = psExec(t, uenv, u+"/scripts/check-plan-shape.sh")
	want("unlanded archives: flow-guard that cannot be built exits 2", r.rc == 2 && has(r, "cannot build flow-guard"), r)
	if err := os.Remove(u + "/scripts/lib/flow-guard.sh"); err != nil {
		t.Fatal(err)
	}
	r = psExec(t, uenv, u+"/scripts/check-plan-shape.sh")
	want("unlanded archives: a missing lib/flow-guard.sh exits 2", r.rc == 2 && has(r, "cannot load lib/flow-guard.sh beside "+u+"/scripts/check-plan-shape.sh"), r)

	// Beyond the harness: the outputs it never pinned, as the wrapper
	// printed them at bbfe685e.
	missing := t.TempDir() + "/nope.md"
	r = one(missing)
	want("pin: a missing file prints each check's read error on stderr", r.rc == 2 && r.stdout == "" &&
		strings.HasPrefix(r.err, missing+": cannot read file: [Errno 2] No such file or directory: '"+missing+"'\n"+
			missing+": cannot read file: open "+missing+": no such file or directory\n"), r)
	r = one(filepath.Dir(missing))
	want("pin: a directory reads as Errno 21", r.rc == 2 &&
		strings.HasPrefix(r.err, filepath.Dir(missing)+": cannot read file: [Errno 21] Is a directory: '"+filepath.Dir(missing)+"'\n"), r)
	r = one(psFixture(t, "- [ ] 1. \xff\n"))
	want("pin: a plan that is not UTF-8 exits 2", r.rc == 2 && has(r, "cannot read file: UnicodeDecodeError"), r)
	r = runGuard("check-plan-shape", []string{"a", "b"}, env)
	want("pin: the usage line", r.err == "usage: check-plan-shape.sh [path-to-tasks.md]\n", r)
	r = one(psFixture(t, psTasks("a", "a")))
	want("pin: violations go to stdout, one per line", r.err == "" && strings.HasSuffix(r.stdout, "explicit\n"), r)
	root = t.TempDir()
	writeFile(t, root+"/spectre/changes/dirty-change/tasks.md", "- [ ] 1. Dirty\n\n**Tests:** `test-a.sh`\n")
	r = runGuard("check-plan-shape", nil, Env{Getenv: psGetenv(map[string]string{"FLOW_GUARD_REPO_ROOT": root})})
	want("pin: with no override the root is FLOW_GUARD_REPO_ROOT", r.rc == 1 && has(r, root+"/spectre/changes/dirty-change/tasks.md:1: task 1 has no **Files:** line"), r)
	r = runGuard("check-plan-shape", nil, Env{Getenv: psGetenv(nil)})
	want("pin: with no override and no FLOW_GUARD_REPO_ROOT it exits 2", r.rc == 2 &&
		r.err == "check-plan-shape: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-plan-shape.sh, which sets it\n", r)
	r = one(psFixture(t, "- [ ] 1. Unread\n\n**Files:** `a.py`\n**Tests:** `x`\n\n  - [ ] **Step 1:** run `./gradlew test`\n"))
	want("pin: the sub-guards' findings follow this guard's in one run", r.rc == 1 && has(r, "runs the suite"), r)

	r = one(psFixture(t, "- [ ] 1. Live verification\n\n**Files:** none\n**Tests:** none — measurement task\n"))
	want("pin: a `**Files:** none` task is no F2 — Python's parser reads it as one prose path", r.rc == 0 && r.out == "", r)

	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.fail != "" {
				t.Error(c.fail)
			}
		})
	}
}

var (
	psScripts, _ = filepath.Abs("../../../scripts")
	psShim       = psScripts + "/check-plan-shape.sh"
)

// psMutant runs the one-file check with the finding tagged off disabled.
func psMutant(path, off string) guardResult {
	var all, o, e bytes.Buffer
	rc := psCheckOne(path, Env{Getenv: os.Getenv}, off, &o, &e)
	all.WriteString(o.String() + e.String())
	return guardResult{rc, strings.TrimRight(all.String(), "\n"), o.String(), e.String()}
}

// psTasks is a plan of one task per file letter, in the harness's shape.
func psTasks(files ...string) string {
	names := []string{"First", "Second", "Third"}
	var b strings.Builder
	for i, f := range files {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "- [ ] %d. %s task\n\n**Files:** `%s.txt`\n**Tests:** `test_%s`\n**Commit:** add %s\n", i+1, names[i], f, f, f)
	}
	return b.String()
}

func psFixture(t *testing.T, body string) string {
	t.Helper()
	p := t.TempDir() + "/tasks.md"
	writeFile(t, p, body)
	return p
}

func psGetenv(vars map[string]string) func(string) string {
	return func(k string) string {
		if v, ok := vars[k]; ok {
			return v
		}
		if k == "PATH" {
			return os.Getenv(k)
		}
		return ""
	}
}

func psCopy(t *testing.T, src, dst string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, dst, string(b))
	if err := os.Chmod(dst, 0o755); err != nil {
		t.Fatal(err)
	}
}

func psLink(t *testing.T, target, link string) {
	t.Helper()
	mkdir(t, filepath.Dir(link))
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

// psExec runs a shim with bash from a directory outside any repository.
func psExec(t *testing.T, extra []string, args ...string) guardResult {
	t.Helper()
	cmd := exec.Command("bash", args...)
	cmd.Dir = t.TempDir()
	cmd.Env = append(append(os.Environ(), "CHECK_PLAN_SHAPE_ROOT="), extra...)
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	rc := 0
	if err := cmd.Run(); err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		rc = ee.ExitCode()
	}
	return guardResult{rc, strings.TrimRight(o.String()+e.String(), "\n"), o.String(), e.String()}
}

// psF15Master is case 39's repository, built once: pkg/ with a
// TestFoldFixupFolds and a TestFooBar on main, then branch work recording
// main as its flowBase.
var psF15Master = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp(execFixtures.dir, "plan-shape-f15")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir+"/pkg", 0o755); err != nil {
		return "", err
	}
	for name, body := range map[string]string{
		"fold_fixup_test.go": "package pkg\n\nfunc TestFoldFixupFolds(t *testing.T) {}\n",
		"foo_bar_test.go":    "package pkg\n\nfunc TestFooBar(t *testing.T) {}\n",
	} {
		if err := os.WriteFile(dir+"/pkg/"+name, []byte(body), 0o644); err != nil {
			return "", err
		}
	}
	g := &fxGit{}
	g.git("", "init", "-q", "-b", "main", dir)
	g.git(dir, "add", "pkg")
	g.git(dir, "-c", "core.hooksPath=/dev/null", "commit", "-qm", "base")
	g.git(dir, "checkout", "-q", "-b", "work")
	g.git(dir, "config", "branch.work.flowBase", "main")
	return dir, g.err
})

func psF15Repo(t *testing.T) string {
	t.Helper()
	master, err := psF15Master()
	if err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir() + "/repo"
	if out, err := exec.Command("cp", "-R", master, repo).CombinedOutput(); err != nil {
		t.Fatalf("cp: %v\n%s", err, out)
	}
	return repo
}
