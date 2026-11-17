package guard

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLoadGate pins the gate heavy runs wait behind: off without LoadGate,
// through at once under the core count or on an unreadable load, through
// after a wait once the load drops, and refused when it never does.
func TestLoadGate(t *testing.T) {
	t.Parallel()
	loads := func(l ...float64) func() (float64, error) {
		return func() (float64, error) {
			v := l[0]
			if len(l) > 1 {
				l = l[1:]
			}
			return v, nil
		}
	}
	wait := func(s string) func(string) string {
		return func(k string) string {
			if k == "FLOW_LOAD_WAIT" {
				return s
			}
			return ""
		}
	}
	for _, c := range []struct {
		name   string
		env    Env
		want   bool
		errHas string
	}{
		{"off", Env{Getenv: wait("0"), LoadAvg: loads(1e9)}, true, ""},
		{"under the core count", Env{Getenv: wait("0"), LoadGate: true, LoadAvg: loads(0)}, true, ""},
		{"unreadable", Env{Getenv: wait("0"), LoadGate: true, LoadAvg: func() (float64, error) { return 0, errors.New("x") }}, true, ""},
		{"drops while waiting", Env{Getenv: wait("60"), LoadGate: true, LoadPoll: time.Millisecond, LoadAvg: loads(1e9, 1e9, 0)}, true, "waiting up to 1m0s"},
		{"never drops", Env{Getenv: wait("0"), LoadGate: true, LoadAvg: loads(1e9)}, false, "no command ran"},
	} {
		var errb bytes.Buffer
		if got := loadGate(c.env, "x", &errb); got != c.want || !strings.Contains(errb.String(), c.errHas) || (c.errHas == "") != (errb.Len() == 0) {
			t.Errorf("%s: %v, stderr %q; want %v and %q", c.name, got, errb.String(), c.want, c.errHas)
		}
	}
}

// TestHeavyRunsWaitBehindTheLoadGate: a `test` or `visual` list and a
// recapture run nothing over the core count; a `lint` list is never held.
func TestHeavyRunsWaitBehindTheLoadGate(t *testing.T) {
	t.Parallel()
	env := Env{LoadGate: true, LoadAvg: func() (float64, error) { return 1e9, nil }, Getenv: func(k string) string {
		if k == "FLOW_LOAD_WAIT" {
			return "0"
		}
		return os.Getenv(k)
	}}
	wt := t.TempDir()
	for list, want := range map[string]bool{"test": true, "visual": true, "lint": false} {
		var o, e bytes.Buffer
		if rc := runVerifyList([]string{wt, list}, env, &o, &e); (rc == 5) != want {
			t.Errorf("%s list: exit %d, held=%v wanted\n%s", list, rc, want, e.String())
		}
	}

	root := t.TempDir()
	for p, body := range map[string]string{
		"package.json":              "{}\n",
		"src/Day.tsx":               "export function Day() {}\n",
		"shots/day.spec.ts":         "test('day', () => {})\n",
		"shots/day.spec.ts.sources": "src/Day.tsx\n",
		"plan/tasks.md":             "- [ ] 4. In-run fix 1 — visual-verify: x\n\n**Files:** `src/Day.tsx`\n**Tests:** none\n",
	} {
		writeFile(t, filepath.Join(root, p), body)
	}
	gitRun(t, root, "init", "-q")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	writeFile(t, filepath.Join(root, "src/Day.tsx"), "export function Day() { return 1 }\n")
	ran := filepath.Join(t.TempDir(), "ran")
	r := runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", "echo <spec> >> " + ran}, env)
	if r.rc != 3 || isFile(ran) || isFile(filepath.Join(root, ".git/flow-recapture.jsonl")) {
		t.Errorf("recapture over the core count: rc=%d out=%q, want 3 with nothing run and no receipt", r.rc, r.out)
	}
}
