package guard

import (
	"os"
	"strings"
	"testing"
)

// readyPlan builds a plan: each spec is "<id>|<after>" under the heading before it.
func readyPlan(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		if strings.HasPrefix(p, "## ") {
			b.WriteString(p + "\n\n")
			continue
		}
		id, after, _ := strings.Cut(p, "|")
		b.WriteString("- [ ] " + id + ". Task " + id + "\n**Files:** `f" + id + ".go`\n")
		if after != "" {
			b.WriteString("**After:** " + after + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func readyRun(t *testing.T, plan, groups string, rows string) guardResult {
	t.Helper()
	env := Env{Getenv: os.Getenv,
		Dispatches: func(string) ([]byte, error) { return []byte(rows), nil },
		Decisions: func(string) ([]byte, error) {
			return []byte(`[{"id":1,"decision":{"groups":[{"tasks":["9"]}]}},{"id":2,"decision":{"groups":` + groups + `}}]`), nil
		}}
	return runGuard("ready-groups", []string{psFixture(t, plan), t.TempDir(), "c"}, env)
}

func TestReadyGroupsLaunchesOneLanePerRepository(t *testing.T) {
	t.Parallel()
	two := readyPlan("## backend", "1|none", "## frontend", "2|none")
	r := readyRun(t, two, `[{"tasks":["1"]},{"tasks":["2"]}]`, `[]`)
	if r.rc != 0 || r.stdout != "launch 1: 1 backend direct\nlaunch 2: 2 frontend direct\n" {
		t.Errorf("two repos: rc=%d out=%q", r.rc, r.stdout)
	}
	one := readyPlan("## backend", "1|none", "2|none")
	r = readyRun(t, one, `[{"tasks":["1"]},{"tasks":["2"]}]`, `[]`)
	if r.rc != 0 || r.stdout != "launch 1: 1 backend throwaway\nlaunch 2: 2 backend throwaway\n" {
		t.Errorf("one repo: rc=%d out=%q", r.rc, r.stdout)
	}
	four := readyPlan("## a", "1|none", "2|none", "3|none", "## b", "4|none")
	r = readyRun(t, four, `[{"tasks":["1"]},{"tasks":["2"]},{"tasks":["3"]},{"tasks":["4"]}]`, `[]`)
	if r.rc != 0 || r.stdout != "launch 1: 1 a throwaway\nlaunch 2: 2 a throwaway\nlaunch 4: 4 b direct\n" {
		t.Errorf("four groups: rc=%d out=%q", r.rc, r.stdout)
	}
}

func TestReadyGroupsWaitsOnAfterAndInFlight(t *testing.T) {
	t.Parallel()
	plan := readyPlan("## a", "1|none", "## b", "2|Task 1", "## c", "3|none")
	groups := `[{"tasks":["1"]},{"tasks":["2"]},{"tasks":["3"]}]`
	r := readyRun(t, plan, groups, `[]`)
	if r.stdout != "launch 1: 1 a direct\nlaunch 3: 3 c direct\n" {
		t.Errorf("unlanded dependency: out=%q", r.stdout)
	}
	done := `[{"key":"task-1-implementer","role":"implementer","outcome":"completed","endedAt":"x"}]`
	if r = readyRun(t, plan, groups, done); r.stdout != "launch 2: 2 b direct\nlaunch 3: 3 c direct\n" {
		t.Errorf("completed row: out=%q", r.stdout)
	}
	ticked := strings.Replace(plan, "- [ ] 1.", "- [x] 1.", 1)
	if r = readyRun(t, ticked, groups, `[]`); r.stdout != "launch 2: 2 b direct\nlaunch 3: 3 c direct\n" {
		t.Errorf("ticked task: out=%q", r.stdout)
	}
	aborted := `[{"key":"task-1-implementer","role":"implementer","outcome":"aborted","endedAt":"x"}]`
	if r = readyRun(t, plan, groups, aborted); r.stdout != "launch 1: 1 a direct\nlaunch 3: 3 c direct\n" {
		t.Errorf("aborted row: out=%q", r.stdout)
	}
	fix := `[{"key":"task-1-implementer-fix-1","role":"implementer"}]`
	if r = readyRun(t, plan, groups, fix); r.rc != 0 || r.stdout != "launch 3: 3 c direct\n" {
		t.Errorf("open fix row: rc=%d out=%q", r.rc, r.stdout)
	}
}

func TestReadyGroupsRefusesAStuckPlan(t *testing.T) {
	t.Parallel()
	plan := readyPlan("## a", "1|Task 2", "2|Task 1")
	r := readyRun(t, plan, `[{"tasks":["1"]},{"tasks":["2"]}]`, `[]`)
	if r.rc != 1 || r.stdout != "" || !strings.Contains(r.err, "group 1 waits on task 2") || !strings.Contains(r.err, "group 2 waits on task 1") {
		t.Errorf("stuck: rc=%d out=%q err=%q", r.rc, r.stdout, r.err)
	}
	all := `[{"key":"task-1-implementer","role":"implementer","outcome":"completed","endedAt":"x"},{"key":"task-2-implementer","role":"implementer","outcome":"completed","endedAt":"x"}]`
	if r = readyRun(t, plan, `[{"tasks":["1"]},{"tasks":["2"]}]`, all); r.rc != 0 || r.stdout != "" || !strings.Contains(r.err, "every group landed") {
		t.Errorf("all landed: rc=%d out=%q err=%q", r.rc, r.stdout, r.err)
	}
	for _, groups := range []string{`[]`, `[{"tasks":[]}]`} {
		if r = readyRun(t, plan, groups, `[]`); r.rc != 2 {
			t.Errorf("groups %s: rc=%d", groups, r.rc)
		}
	}
	if r = readyRun(t, plan, `[{"tasks":["1"]}]`, `not json`); r.rc != 2 {
		t.Errorf("bad rows: rc=%d", r.rc)
	}
	env := Env{Getenv: os.Getenv}
	if r = runGuard("ready-groups", []string{"a"}, env); r.rc != 2 || !strings.Contains(r.err, "usage") {
		t.Errorf("usage: rc=%d", r.rc)
	}
	if r = runGuard("ready-groups", []string{"/nonexistent/tasks.md", "w", "c"}, env); r.rc != 2 {
		t.Errorf("unreadable plan: rc=%d", r.rc)
	}
	env.Dispatches = func(string) ([]byte, error) { return []byte(`[]`), nil }
	env.Decisions = func(string) ([]byte, error) { return []byte(`[]`), nil }
	if r = runGuard("ready-groups", []string{psFixture(t, plan), "w", "c"}, env); r.rc != 2 {
		t.Errorf("no decision: rc=%d", r.rc)
	}
}

func TestReadyGroupsInFlightHoldsItsRepositoryAndBudget(t *testing.T) {
	t.Parallel()
	groups := `[{"tasks":["1"]},{"tasks":["2"]},{"tasks":["3"]}]`
	plan := readyPlan("## a", "1|none", "2|none", "## b", "3|none")
	open1 := `[{"key":"task-1-implementer","role":"implementer"}]`
	if r := readyRun(t, plan, groups, open1); r.rc != 0 || r.stdout != "launch 3: 3 b direct\n" {
		t.Errorf("one lane per repository: rc=%d out=%q", r.rc, r.stdout)
	}
	five := readyPlan("## a", "1|none", "## b", "2|none", "## c", "3|none", "## d", "4|none", "## e", "5|none")
	two := `[{"key":"task-1-implementer","role":"implementer"},{"key":"task-2-implementer","role":"implementer"}]`
	if r := readyRun(t, five, `[{"tasks":["1"]},{"tasks":["2"]},{"tasks":["3"]},{"tasks":["4"]},{"tasks":["5"]}]`, two); r.rc != 0 || r.stdout != "launch 3: 3 c direct\n" {
		t.Errorf("budget less in flight: rc=%d out=%q", r.rc, r.stdout)
	}
}

func TestReadyGroupsLandsOnlyFirstPassRows(t *testing.T) {
	t.Parallel()
	plan := readyPlan("## a", "1|none", "2|Task 1")
	groups := `[{"tasks":["1"]},{"tasks":["2"]}]`
	for _, key := range []string{"task-1-implementer-fix-1", "task-1-implementer-recommit-1", "task-1-implementer-pick-1"} {
		row := `[{"key":"` + key + `","role":"implementer","outcome":"completed","endedAt":"x"}]`
		if r := readyRun(t, plan, groups, row); r.stdout != "launch 1: 1 a direct\n" {
			t.Errorf("ended %s: out=%q", key, r.stdout)
		}
	}
	retry := `[{"key":"task-1-implementer-retry","role":"implementer","outcome":"completed","endedAt":"x"}]`
	if r := readyRun(t, plan, groups, retry); r.stdout != "launch 2: 2 a direct\n" {
		t.Errorf("ended retry: out=%q", r.stdout)
	}
}

func TestReadyGroupsReturnedRowAwaitingItsEnd(t *testing.T) {
	t.Parallel()
	plan := strings.Replace(readyPlan("## a", "1|none", "2|Task 1"), "- [ ] 1.", "- [x] 1.", 1)
	groups := `[{"tasks":["1"]},{"tasks":["2"]}]`
	for _, key := range []string{"task-1-implementer", "task-1-implementer-retry"} {
		row := `[{"key":"` + key + `","role":"implementer"}]`
		if r := readyRun(t, plan, groups, row); r.rc != 0 || r.stdout != "launch 2: 2 a direct\n" {
			t.Errorf("open %s, task ticked: rc=%d out=%q err=%q", key, r.rc, r.stdout, r.err)
		}
	}
	fix := `[{"key":"task-1-implementer-fix-1","role":"implementer"}]`
	if r := readyRun(t, plan, groups, fix); r.stdout != "" || !strings.Contains(r.err, "1 in flight") {
		t.Errorf("open fix row stays in flight: out=%q err=%q", r.stdout, r.err)
	}
}
