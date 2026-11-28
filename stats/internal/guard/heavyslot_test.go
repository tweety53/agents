package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// fakeSlots is flowd's heavy-slot API as HeavySlot calls it: acquire grants
// once grantAfter refusals are spent, release and renewals are counted.
type fakeSlots struct {
	mu                         sync.Mutex
	grantAfter                 int // -1: never grant
	acquires, renews, released int
	held                       string
}

func (f *fakeSlots) serve(t *testing.T) string {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			ID  string `json:"id"`
			PID int    `json:"pid"`
			TTL int    `json:"ttl_seconds"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		f.mu.Lock()
		defer f.mu.Unlock()
		switch r.URL.Path {
		case "/api/v1/slots/acquire":
			if in.ID == "" || in.PID != os.Getpid() || in.TTL <= 0 {
				http.Error(w, "bad acquire", http.StatusBadRequest)
				return
			}
			if f.held == in.ID {
				f.renews++
			} else {
				f.acquires++
			}
			granted := f.held == in.ID || (f.grantAfter >= 0 && f.acquires > f.grantAfter)
			if granted {
				f.held = in.ID
			}
			fmt.Fprintf(w, `{"granted":%v,"max":1,"holders":[]}`, granted)
		case "/api/v1/slots/release":
			if in.ID == f.held {
				f.held = ""
			}
			f.released++
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func slotEnv(addr, wait string) core.Env {
	return core.Env{SlotGate: true, SlotPoll: time.Millisecond, SlotRenew: time.Millisecond, Getenv: func(k string) string {
		switch k {
		case "FLOW_ADDR":
			return addr
		case "FLOW_SLOT_WAIT":
			return wait
		}
		return os.Getenv(k)
	}}
}

// TestHeavySlot pins the gate heavy runs wait behind: off without SlotGate;
// through at once on a free slot, renewed while held and released by the
// returned func; through after a wait once a slot frees; refused when none
// frees within FLOW_SLOT_WAIT; through, never held, when flowd is unreachable.
func TestHeavySlot(t *testing.T) {
	t.Parallel()
	var errb bytes.Buffer
	if release, ok := core.HeavySlot(core.Env{Getenv: func(string) string { return "" }}, "x", "o", &errb); !ok || errb.Len() != 0 {
		t.Errorf("off: %v %q", ok, errb.String())
	} else {
		release()
	}

	f := &fakeSlots{}
	release, ok := core.HeavySlot(slotEnv(f.serve(t), "0"), "x", "o", &errb)
	if !ok || errb.Len() != 0 {
		t.Fatalf("free slot: %v %q", ok, errb.String())
	}
	time.Sleep(20 * time.Millisecond)
	release()
	f.mu.Lock()
	if f.renews == 0 || f.released != 1 || f.held != "" {
		t.Errorf("held slot: renews %d released %d held %q, want renewals, one release, nothing held", f.renews, f.released, f.held)
	}
	f.mu.Unlock()

	f = &fakeSlots{grantAfter: 2}
	if release, ok := core.HeavySlot(slotEnv(f.serve(t), "60"), "x", "o", &errb); !ok || !strings.Contains(errb.String(), "waiting up to 1m0s") {
		t.Errorf("frees while waiting: %v %q", ok, errb.String())
	} else {
		release()
	}

	errb.Reset()
	f = &fakeSlots{grantAfter: -1}
	if _, ok := core.HeavySlot(slotEnv(f.serve(t), "0"), "x", "o", &errb); ok || !strings.Contains(errb.String(), "no command ran") {
		t.Errorf("never frees: %v %q", ok, errb.String())
	}

	errb.Reset()
	if release, ok := core.HeavySlot(slotEnv("http://127.0.0.1:1", "0"), "x", "o", &errb); !ok || !strings.Contains(errb.String(), "unreachable") {
		t.Errorf("unreachable: %v %q", ok, errb.String())
	} else {
		release()
	}
}

// TestHeavyRunsWaitForASlot: a `test` or `visual` list and a recapture run
// nothing while no heavy slot frees; a `lint` list is never held.
func TestHeavyRunsWaitForASlot(t *testing.T) {
	t.Parallel()
	env := slotEnv((&fakeSlots{grantAfter: -1}).serve(t), "0")
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
		guardtest.WriteFile(t, filepath.Join(root, p), body)
	}
	guardtest.GitRun(t, root, "init", "-q")
	guardtest.GitRun(t, root, "add", "-A")
	guardtest.GitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	guardtest.WriteFile(t, filepath.Join(root, "src/Day.tsx"), "export function Day() { return 1 }\n")
	ran := filepath.Join(t.TempDir(), "ran")
	r := runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", "echo <spec> >> " + ran}, env)
	if r.Rc != 3 || core.IsFile(ran) || core.IsFile(filepath.Join(root, ".git/flow-recapture.jsonl")) {
		t.Errorf("recapture with no slot: rc=%d out=%q, want 3 with nothing run and no receipt", r.Rc, r.Out)
	}
}

// TestHeavySlotRecordsItsWait: the time HeavySlot spent waiting for a grant
// lands in the run's collector as wait, so flow-guard posts it as wait_ms;
// a slot granted at once records next to none.
func TestHeavySlotRecordsItsWait(t *testing.T) {
	t.Parallel()
	f := &fakeSlots{grantAfter: 3}
	env := slotEnv(f.serve(t), "60")
	env.SlotPoll = 20 * time.Millisecond
	env.Reasons = &core.Reasons{}
	var errb bytes.Buffer
	release, ok := core.HeavySlot(env, "x", "o", &errb)
	if !ok {
		t.Fatalf("not granted: %q", errb.String())
	}
	release()
	if w := env.Reasons.Wait(); w < 60*time.Millisecond {
		t.Errorf("wait = %s, want at least the three 20 ms polls", w)
	}
}
