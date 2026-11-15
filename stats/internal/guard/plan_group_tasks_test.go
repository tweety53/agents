package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPlanGroupTasks pins decisions group-tasks-sibling and
// group-tasks-at-entry: each recorded group's bundle ids map to the task ids
// plan-dispatch-bundles numbers them by, and an unknown bundle id is exit 2.
func TestPlanGroupTasks(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	task := func(id, files string) string {
		s := "- [ ] " + id + ". Task " + id + "\n\n"
		if files != "" {
			s += "**Files:**\n- Create: `" + files + "`\n\n"
		}
		return s + "**After:** none\n\n  - [ ] **Step 1: do it**\n\n"
	}
	plan := func(third string) string {
		return task("1", "a.txt") + task("2", "a.txt") + task("3", third) + task("4", "c.txt") + task("5", "d.txt")
	}
	md := psFixture(t, plan("b.txt"))
	decision := func(body string) string {
		p := filepath.Join(t.TempDir(), "decision.json")
		writeFile(t, p, body)
		return p
	}
	run := func(args ...string) guardResult { return runGuard("plan-group-tasks", args, env) }

	r := run(md, decision(`{"groups":[{"bundles":[1]},{"bundles":[2,4]},{"bundles":[3]}]}`))
	if r.rc != 0 || r.stdout != "tasks 1: 1,2\ntasks 2: 3,5\ntasks 3: 4\n" {
		t.Errorf("(a) rc=%d stdout=%q err=%q", r.rc, r.stdout, r.err)
	}
	r = run(md, decision(`{"groups":[{"bundles":[1,9]}]}`))
	if r.rc != 2 || r.stdout != "" || !strings.Contains(r.err, "9") {
		t.Errorf("(b) unknown bundle: rc=%d stdout=%q err=%q", r.rc, r.stdout, r.err)
	}
	r = run(md, decision(`{"groups":[{"bundles":[1]},{"bundles":[]}]}`))
	if r.rc != 2 || r.stdout != "" || !strings.Contains(r.err, "group 2") {
		t.Errorf("(b) empty group: rc=%d stdout=%q err=%q", r.rc, r.stdout, r.err)
	}
	for _, body := range []string{`{"groups":null}`, `{}`, `not json`} {
		if r = run(md, decision(body)); r.rc != 2 {
			t.Errorf("(c) %s: rc=%d out=%q", body, r.rc, r.out)
		}
	}
	if r = run(md, decision(`{"groups":[{"bundles":[1]}]}`)+".missing"); r.rc != 2 {
		t.Errorf("(c) unreadable decision: rc=%d out=%q", r.rc, r.out)
	}
	for _, args := range [][]string{{md}, {md, md, md}} {
		if r = run(args...); r.rc != 2 || !strings.Contains(r.err, "usage: flow-guard plan-group-tasks") {
			t.Errorf("(d) %d args: rc=%d err=%q", len(args), r.rc, r.err)
		}
	}
	r = run(psFixture(t, plan("")), decision(`{"groups":[{"bundles":[1]}]}`))
	if r.rc != 1 || r.stdout != "" || r.err == "" {
		t.Errorf("(e) bundle violation: rc=%d stdout=%q err=%q", r.rc, r.stdout, r.err)
	}
}
