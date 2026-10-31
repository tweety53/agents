package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPlanDispatchBundles ports scripts/test-plan-dispatch-bundles.sh at
// bbfe685e: one subtest per ok: label, the guard run in-process against real
// fixture files (and a real git repository for the unlanded-archive case).
// The harness's wrapper cases -- the no-argument scan through a skill
// directory's symlink, a flow-guard that cannot be built, a missing sibling
// -- exec the real shim.
func TestPlanDispatchBundles(t *testing.T) {
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
	one := func(path string) guardResult { return runGuard("plan-dispatch-bundles", []string{path}, env) }
	scan := func(vars map[string]string) guardResult {
		return runGuard("plan-dispatch-bundles", nil, Env{Getenv: psGetenv(vars)})
	}
	run := func(body string) (string, guardResult) {
		md := psFixture(t, body)
		return md, one(md)
	}
	step := "  - [ ] **Step 1: do it**\n"
	tk := func(id, files string) string {
		return "- [ ] " + id + ". Task " + id + "\n\n**Files:**\n- Create: `" + files + "`\n\n" + step + "\n"
	}
	fieldless := "- [ ] 1. Fieldless task\n\nNo Files field here at all.\n\n" + step

	_, r := run(tk("1", "a.txt") + tk("2", "b.txt") + tk("3", "c.txt"))
	want("case 1: disjoint tasks exit 0", r.rc == 0, r)
	want("case 1: three separate bundles", r.out == "bundle 1: 1\nafter 1: none\nbundle 2: 2\nafter 2: 1\nbundle 3: 3\nafter 3: 1 2", r)

	_, r = run("- [ ] 1. First\n\n**Files:**\n- Modify: `shared.txt`\n\n" + step + "\n- [ ] 2. Second\n\n**Files:**\n- Modify: `shared.txt`\n\n" + step)
	want("case 2: shared path exits 0", r.rc == 0, r)
	want("case 2: one bundle for both tasks", r.out == "bundle 1: 1 2\nafter 1: 1", r)

	_, r = run(tk("1", "x.txt") + "- [ ] 2. Second\n\n**Files:**\n- Create: `x.txt`\n- Create: `y.txt`\n\n" + step + "\n" + tk("3", "y.txt"))
	want("case 3: transitive chain exits 0", r.rc == 0, r)
	want("case 3: transitive closure forms one bundle", r.out == "bundle 1: 1 2 3\nafter 1: 1 2", r)

	_, r = run("- [x] 1. Done task\n\n**Files:**\n- Modify: `shared.txt`\n\n  - [ ] **Step 1: a step left unmarked**\n\n- [ ] 2. Open task\n\n**Files:**\n- Modify: `shared.txt`\n\n  - [x] **Step 1: a step already marked**\n")
	want("case 4: mixed checked/unchecked exits 0", r.rc == 0, r)
	want("case 4: checked task excluded entirely", r.out == "bundle 1: 2\nafter 1: 1", r)

	_, r = run(tk("1", "a.txt") + "- [ ] 2. Fieldless task\n\nNo Files field here at all.\n\n" + step)
	want("case 5: missing Files field exits 1", r.rc == 1, r)
	want("case 5: names task 2", has(r, "task 2 has no **Files:** field"), r)

	_, r = run("- [ ] 1. First\n\n**Files:**\n- Create: `a.txt`\n\n**Allowed-collateral:** `shared.txt`\n\n" + step + "\n" + tk("2", "shared.txt"))
	want("case 6: collateral glob exits 0", r.rc == 0, r)
	want("case 6: collateral does not join bundles", r.out == "bundle 1: 1\nafter 1: none\nbundle 2: 2\nafter 2: 1", r)

	r = one(t.TempDir() + "/does-not-exist.md")
	want("case 7: unreadable path exits 2", r.rc == 2, r)

	_, r = run("- [ ] 1. First\n\n**Files:**\n- Modify: `a.txt`, `b.txt`\n\n" + step + "\n- [ ] 2. Second\n\n**Files:**\n- Modify: `b.txt`\n\n" + step)
	want("case 8: multi-path bullet exits 0", r.rc == 0, r)
	want("case 8: second declared path in a multi-path bullet still joins the bundle", r.out == "bundle 1: 1 2\nafter 1: 1", r)

	_, r = run(tk("2", "x.txt") + tk("10", "y.txt"))
	want("case 9: two-digit ids exit 0", r.rc == 0, r)
	want("case 9: 10 sorts after 2 numerically", r.out == "bundle 1: 2\nafter 1: none\nbundle 2: 10\nafter 2: 2", r)

	// The one fixture whose steps sit at column 0, where BULLET_RE would
	// read them as a Files entry but for its checkbox exclusion.
	_, r = run("- [ ] 1. First\n\n**Files:**\n- Modify: `a.txt`\n- [ ] **Step 1: do it**\n\n- [ ] 2. Second\n\n**Files:**\n- Modify: `b.txt`\n- [ ] **Step 1: do it**\n")
	want("case 10: adjacent step checkbox after Files block exits 0", r.rc == 0, r)
	want("case 10: the step line is not swallowed as a declared path", r.out == "bundle 1: 1\nafter 1: none\nbundle 2: 2\nafter 2: 1", r)

	// Case 11: with no override and no argument, the shim reached through a
	// skills/flow/scripts/ symlink scans its REAL root's changes.
	sandbox := t.TempDir()
	psCopy(t, pdbShim, sandbox+"/scripts/plan-dispatch-bundles.sh")
	psCopy(t, psScripts+"/lib/flow-guard.sh", sandbox+"/scripts/lib/flow-guard.sh")
	psLink(t, filepath.Dir(psScripts)+"/stats", sandbox+"/stats")
	psLink(t, "../../../scripts/plan-dispatch-bundles.sh", sandbox+"/skills/flow/scripts/plan-dispatch-bundles.sh")
	psLink(t, "../../../scripts/lib", sandbox+"/skills/flow/scripts/lib")
	writeFile(t, sandbox+"/spectre/changes/some-change/tasks.md", fieldless)
	r = psExec(t, []string{"PLAN_DISPATCH_BUNDLES_ROOT=", "FLOW_GUARD_CACHE_DIR=" + guardCache(t)}, sandbox+"/skills/flow/scripts/plan-dispatch-bundles.sh")
	want("case 11: no-arg default resolves through a skill-dir symlink to the real repo root", r.rc == 1, r)
	want("case 11: names task 1 from the fixture's own spectre/changes/", has(r, "task 1 has no **Files:** field"), r)
	want("pin: the symlinked shim's violation names the real root's file on stderr", r.stdout == "" &&
		strings.HasSuffix(r.err, sandbox+"/spectre/changes/some-change/tasks.md:1: task 1 has no **Files:** field\n") && strings.Count(r.err, "\n") == 1, r)

	_, r = run("- [ ] 1. The only real task\n\n**Files:**\n- Create: `a.txt`\n\n### 2 A heading, not a task\n\nNo Files field here at all.\n")
	want("case 12: a level-3 heading opens no task", r.rc == 0, r)
	want("case 12: only the checkbox task bundles", r.out == "bundle 1: 1\nafter 1: none", r)

	_, r = run("- [ ] 1. The only real task\n\n**Files:**\n- Create: `a.txt`\n\n" + step + "\n- [ ] 1.1. A dotted task line, which is no task\n\nNo Files field here at all.\n")
	want("case 13: a dotted task line opens no task", r.rc == 0, r)
	want("case 13: only the flat-id task bundles", r.out == "bundle 1: 1\nafter 1: none", r)

	_, r = run("- [ ] 1. The only real task\n\n**Files:**\n- Create: `a.txt`\n\n" + step + "\n  - [ ] 2. An indented task-shaped line, which is no task\n")
	want("case 14: an indented task-shaped line opens no task", r.rc == 0, r)
	want("case 14: only the column-0 task bundles", r.out == "bundle 1: 1\nafter 1: none", r)

	_, r = run("- [ ] 1. Red\n\n**Build:** red\n\n**Squash-with:** Task 2\n\n**Files:**\n- Create: `a_test.go`\n\n" + step +
		"\n- [ ] 2. Green\n\n**Build:** green\n\n**Files:**\n- Modify: `a.go`\n\n" + step +
		"\n- [ ] 3. Unrelated\n\n**Build:** green\n\n**Files:**\n- Modify: `b.go`\n\n" + step)
	want("case 15: a Squash-with pair exits 0", r.rc == 0, r)
	want("case 15: a red task bundles with the partner it names", r.out == "bundle 1: 1 2\nafter 1: 1\nbundle 2: 3\nafter 2: 1 2", r)

	_, r = run("- [ ] 1. Red\n\n**Build:** red\n\n**Squash-with:** Task 2\n\n**Files:**\n- Create: `a_test.go`\n\n" + step +
		"\n- [x] 2. Green, done\n\n**Build:** green\n\n**Files:**\n- Modify: `a.go`\n\n  - [x] **Step 1: make them pass**\n" +
		"\n- [ ] 3. Unrelated\n\n**Build:** green\n\n**Files:**\n- Modify: `b.go`\n\n" + step)
	want("case 16: a red task naming a checked partner exits 0", r.rc == 0, r)
	want("case 16: a checked partner joins nothing", r.out == "bundle 1: 1\nafter 1: none\nbundle 2: 3\nafter 2: 1 2", r)

	_, r = run("- [ ] 1. Red\n\n**Build:** red\n\n**Squash-with:** not-a-valid-id\n\n**Files:**\n- Create: `a_test.go`\n\n" + step +
		"\n- [ ] 2. Unrelated\n\n**Build:** green\n\n**Files:**\n- Modify: `b.go`\n\n" + step)
	want("case 17: an ungated Squash-with value exits 0, not a crash", r.rc == 0, r)
	want("case 17: the malformed value bundles the task alone", r.out == "bundle 1: 1\nafter 1: none\nbundle 2: 2\nafter 2: 1", r)

	_, r = run(tk("1", "a.txt") + "- [ ] 2. Second\n\n**After:** Task 1\n\n**Files:**\n- Create: `b.txt`\n\n" + step)
	want("case 18: declared After exits 0", r.rc == 0, r)
	want("case 18: the declared After set is printed", r.out == "bundle 1: 1\nafter 1: none\nbundle 2: 2\nafter 2: 1", r)

	_, r = run(tk("1", "a.txt") + tk("2", "b.txt") + tk("3", "c.txt"))
	want("case 19: fieldless plan exits 0", r.rc == 0, r)
	want("case 19: the serial default expands to every earlier id", r.out == "bundle 1: 1\nafter 1: none\nbundle 2: 2\nafter 2: 1\nbundle 3: 3\nafter 3: 1 2", r)

	_, r = run(tk("1", "a.txt") + "- [ ] 2. Independent\n\n**After:** none\n\n**Files:**\n- Create: `b.txt`\n\n" + step + "\n" + tk("3", "c.txt"))
	want("case 20: After none exits 0", r.rc == 0, r)
	want("case 20: none lists nobody despite earlier tasks", r.out == "bundle 1: 1\nafter 1: none\nbundle 2: 2\nafter 2: none\nbundle 3: 3\nafter 3: 1 2", r)

	_, r = run(tk("1", "a.txt") + "- [ ] 2. Red\n\n**Build:** red\n\n**After:** Task 1\n\n**Squash-with:** Task 3\n\n**Files:**\n- Create: `t_test.go`\n\n" + step +
		"\n- [ ] 3. Green\n\n**Build:** green\n\n**Files:**\n- Modify: `impl.go`\n\n" + step)
	want("case 21: red pair with After exits 0", r.rc == 0, r)
	want("case 21: the pair's after-set is the union", r.out == "bundle 1: 1\nafter 1: none\nbundle 2: 2 3\nafter 2: 1 2", r)

	_, r = run("- [x] 1. Done task\n\n**Files:**\n- Create: `a.txt`\n\n  - [x] **Step 1: done**\n\n" + tk("2", "b.txt"))
	want("case 22: resumed plan exits 0", r.rc == 0, r)
	want("case 22: the checked task still precedes the open one", r.out == "bundle 1: 2\nafter 1: 1", r)

	_, r = run("- [ ] 1. First\n\n**After:** Task 10\n\n**Files:**\n- Create: `shared.txt`\n\n" + step + "\n" + tk("2", "b.txt") + tk("3", "shared.txt") + tk("10", "j.txt"))
	want("case 23: mixed-width after-set exits 0", r.rc == 0, r)
	want("case 23: the after line sorts numerically, not lexically", r.out == "bundle 1: 1 3\nafter 1: 1 2 10\nbundle 2: 2\nafter 2: 1\nbundle 3: 10\nafter 3: 1 2 3", r)

	// Unlanded archives: a shim copy with no stats/ beside it cannot build
	// flow-guard; the same copy with no lib/flow-guard.sh cannot load it.
	u := t.TempDir()
	mkdir(t, u+"/root/spectre/changes/some-change")
	psCopy(t, pdbShim, u+"/scripts/plan-dispatch-bundles.sh")
	psCopy(t, psScripts+"/lib/flow-guard.sh", u+"/scripts/lib/flow-guard.sh")
	uenv := []string{"PLAN_DISPATCH_BUNDLES_ROOT=" + u + "/root", "FLOW_GUARD_CACHE_DIR=" + t.TempDir()}
	r = psExec(t, uenv, u+"/scripts/plan-dispatch-bundles.sh")
	want("unlanded archives: flow-guard that cannot be built exits 2", r.rc == 2 && has(r, "cannot build flow-guard"), r)
	if err := os.Remove(u + "/scripts/lib/flow-guard.sh"); err != nil {
		t.Fatal(err)
	}
	r = psExec(t, uenv, u+"/scripts/plan-dispatch-bundles.sh")
	want("unlanded archives: a missing lib/flow-guard.sh exits 2", r.rc == 2 && has(r, "cannot load lib/flow-guard.sh beside "+u+"/scripts/plan-dispatch-bundles.sh"), r)

	// Beyond the harness: the outputs it never pinned, as the Python and its
	// wrapper printed them at bbfe685e.
	missing := t.TempDir() + "/nope.md"
	r = one(missing)
	want("pin: a missing file exits 2 with the OSError on stderr", r.rc == 2 && r.stdout == "" &&
		r.err == missing+": cannot read file: [Errno 2] No such file or directory: '"+missing+"'\n", r)
	r = one(t.TempDir())
	want("pin: a directory reads as Errno 21", r.rc == 2 && strings.Contains(r.err, ": cannot read file: [Errno 21] Is a directory: '"), r)
	md, r := run("- [ ] 1. \xff\n")
	want("pin: a plan that is not UTF-8 exits 2 with the codec's message", r.rc == 2 &&
		r.err == md+": cannot read file: 'utf-8' codec can't decode byte 0xff in position 9: invalid start byte\n", r)
	r = runGuard("plan-dispatch-bundles", []string{"a", "b"}, env)
	want("pin: the usage line", r.rc == 2 && r.err == "usage: plan-dispatch-bundles.sh [path-to-tasks.md]\n", r)
	md, r = run("- [ ] 1. A\n\n**After:** Task 2 (soon)\n\n**Files:**\n- `a`\n\n- [ ] 2. B\n\nnothing\n\n- [x] 3. C\n\nnothing\n\n- [ ] 4. D\n\n**After:** sometime\n")
	want("pin: violations go to stderr, every missing field before every malformed After", r.rc == 1 && r.stdout == "" &&
		r.err == md+":8: task 2 has no **Files:** field\n"+
			md+":16: task 4 has no **Files:** field\n"+
			md+":1: task 1 has a malformed **After:** value: Task 2 (soon)\n"+
			md+":16: task 4 has a malformed **After:** value: sometime\n", r)
	_, r = run("- [x] 1. Done\n\n**Files:**\n- `a`\n")
	want("pin: a plan with no unchecked task prints nothing and exits 0", r.rc == 0 && r.out == "", r)
	// Python's BULLET_RE / _extract_paths: the Create:/Modify:/Delete: prefix
	// is stripped, untick text is the path, a fenced bullet adds nothing and
	// does not end the run, and a Files line's own inline value is not read.
	_, r = run("- [ ] 1. A\n\n**Files:** `inline.txt`\n-   Delete:   plain path  \n```\n- `fenced.txt`\n```\n- Create: `x.txt`\n\n" +
		"- [ ] 2. B\n\n**Files:**\n- plain path\n\n- [ ] 3. C\n\n**Files:**\n- `fenced.txt`\n- `inline.txt`\n\n- [ ] 4. D\n\n**Files:**\n- `x.txt`\n")
	want("pin: Files bullets parse as the Python's BULLET_RE did", r.rc == 0 && r.out == "bundle 1: 1 2 4\nafter 1: 1 2 3\nbundle 2: 3\nafter 2: 1 2", r)
	root := t.TempDir()
	writeFile(t, root+"/spectre/changes/a/tasks.md", tk("1", "a.txt"))
	writeFile(t, root+"/spectre/changes/b/tasks.md", "- [ ] 1. \xff\n")
	writeFile(t, root+"/spectre/changes/c/tasks.md", fieldless)
	r = scan(map[string]string{"FLOW_GUARD_REPO_ROOT": root})
	want("pin: with no override the root is FLOW_GUARD_REPO_ROOT and 2 beats 1", r.rc == 2 &&
		r.stdout == "bundle 1: 1\nafter 1: none\n" && has(r, "c/tasks.md:1: task 1 has no **Files:** field"), r)
	r = scan(nil)
	want("pin: with no override and no FLOW_GUARD_REPO_ROOT it exits 2", r.rc == 2 &&
		r.err == "plan-dispatch-bundles: FLOW_GUARD_REPO_ROOT is unset — run scripts/plan-dispatch-bundles.sh, which sets it\n", r)
	r = scan(map[string]string{"PLAN_DISPATCH_BUNDLES_ROOT": t.TempDir()})
	want("pin: a root with no changes directory passes silently", r.rc == 0 && r.out == "", r)

	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.fail != "" {
				t.Error(c.fail)
			}
		})
	}
}

var pdbShim = psScripts + "/plan-dispatch-bundles.sh"
