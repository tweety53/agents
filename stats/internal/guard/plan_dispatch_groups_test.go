package guard

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestPlanDispatchGroups ports scripts/test-plan-dispatch-groups.sh at
// bbfe685e: one subtest per ok: label, the guard run in-process against real
// fixture files; the no-argument case execs the real shim.
func TestPlanDispatchGroups(t *testing.T) {
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
	env := Env{Getenv: os.Getenv}
	one := func(path string) guardResult { return runGuard("plan-dispatch-groups", []string{path}, env) }
	run := func(body string) (string, guardResult) {
		md := psFixture(t, body)
		return md, one(md)
	}
	// task is the harness's task(): one unchecked task, one **Files:**
	// entry, an optional **After:** field.
	task := func(id, files, after string) string {
		s := "- [ ] " + id + ". Task " + id + "\n\n**Files:**\n- Create: `" + files + "`\n\n"
		if after != "" {
			s += "**After:** " + after + "\n\n"
		}
		return s + "  - [ ] **Step 1: do it**\n\n"
	}

	_, r := run(task("1", "a.txt", "") + task("2", "b.txt", "") + task("3", "c.txt", "") + task("4", "d.txt", "") + task("5", "e.txt", ""))
	want("case 1: serial-default exits 0", r.rc == 0, r)
	want("case 1: one group for a fully serial plan", r.out == "group 1: 1 2 3 4 5", r)

	_, r = run(task("1", "a.txt", "none") + task("2", "b.txt", "none") + task("3", "c.txt", "Task 1") + task("4", "d.txt", "Task 2") + task("5", "e.txt", "Task 3, 4"))
	want("case 2: two chains plus a join point exits 0", r.rc == 0, r)
	want("case 2: chains stay split, join point opens its own group", r.out == "group 1: 1 3\ngroup 2: 2 4\ngroup 3: 5", r)

	_, r = run(task("1", "a.txt", "none") + task("2", "b.txt", "none") + task("3", "c.txt", "none") + task("4", "d.txt", "none") + task("5", "e.txt", "Task 1, 2, 3, 4"))
	want("case 3: four singletons plus a join exits 0", r.rc == 0, r)
	want("case 3: fold to two, round-robin, join point untouched", r.out == "group 1: 1 3\ngroup 2: 2 4\ngroup 3: 5", r)

	_, r = run(task("1", "a.txt", "none") + task("2", "b.txt", "none"))
	want("case 3b: two singletons exits 0", r.rc == 0, r)
	want("case 3b: two singletons stay two groups", r.out == "group 1: 1\ngroup 2: 2", r)

	_, r = run(task("1", "a.txt", "") + task("2", "b.txt", "Task 1") + task("3", "c.txt", "Task 1") + task("4", "d.txt", "Task 2, 3"))
	want("case 4: the ceiling case exits 0", r.rc == 0, r)
	want("case 4: one group, the documented ceiling", r.out == "group 1: 1 2 3 4", r)

	md, r := run(task("1", "a.txt", "") + "- [ ] 2. Fieldless task\n\n  - [ ] **Step 1: do it**\n")
	want("case 5: missing Files field exits 1", r.rc == 1, r)
	want("case 5: violation message passed through", strings.Contains(r.out, "task 2 has no **Files:** field"), r)
	want("pin: the bundle violation passes through to stderr byte for byte", r.stdout == "" && r.err == md+":8: task 2 has no **Files:** field\n", r)

	_, r = run("- [x] 1. Done task\n\n**Files:**\n- Modify: `shared.txt`\n\n  - [x] **Step 1: do it**\n")
	want("case 6: zero unchecked tasks prints nothing, exits 0", r.rc == 0 && r.out == "", r)

	r = psExec(t, []string{"FLOW_GUARD_CACHE_DIR=" + guardCache(t)}, pdgShim)
	want("case 7: no arguments exits 2", r.rc == 2, r)
	want("pin: the usage line", r.err == "usage: plan-dispatch-groups.sh <path-to-tasks.md>\n", r)

	// Beyond the harness: the outputs it never pinned, as the Python and its
	// wrapper printed them at bbfe685e.
	missing := t.TempDir() + "/nope.md"
	r = one(missing)
	want("pin: a missing file exits 2 with the OSError on stderr", r.rc == 2 && r.stdout == "" &&
		r.err == missing+": cannot read file: [Errno 2] No such file or directory: '"+missing+"'\n", r)
	// Fold: five after:none singletons deal round-robin into two lanes.
	_, r = run(task("1", "a", "none") + task("2", "b", "none") + task("3", "c", "none") + task("4", "d", "none") + task("5", "e", "none"))
	want("pin: five ready-alike singletons fold into two lanes", r.rc == 0 && r.out == "group 1: 1 3 5\ngroup 2: 2 4", r)
	// A bundle whose member names a bundle-mate in After: its own member is
	// no dependency, so it opens with an empty ready set.
	_, r = run(task("1", "a", "none") + "- [ ] 2. Task 2\n\n**After:** Task 2\n\n**Files:**\n- Create: `b`\n\n" + task("3", "c", "none"))
	want("pin: a bundle's own members are not its dependencies", r.rc == 0 && r.out == "group 1: 1 3\ngroup 2: 2", r)

	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.fail != "" {
				t.Error(c.fail)
			}
		})
	}
}

var pdgShim = psScripts + "/plan-dispatch-groups.sh"
