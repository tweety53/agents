package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// TestAwaitSubAgents ports scripts/test-await-sub-agents.sh cases (a)-(e):
// 0 once every named agent's meta.json says finished, 1 at the cap while one
// still runs, 2 on a usage error, an agent it cannot find, or a cap above 240.
func TestAwaitSubAgents(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "proj", "sess", "subagents")
	mkdir(t, dir)
	writeFile(t, filepath.Join(dir, "agent-aaa.meta.json"), `{"agentType":"Plan","finished":true}`+"\n")
	writeFile(t, filepath.Join(dir, "agent-bbb.meta.json"), `{"agentType":"Plan"}`+"\n")
	writeFile(t, filepath.Join(dir, "agent-ccc.meta.json"), `{"agentType":"Plan"}`+"\n")
	// A meta.json outside a subagents/ directory is not an agent's record.
	writeFile(t, filepath.Join(tmp, "proj", "sess", "agent-ddd.meta.json"), `{"agentType":"Plan","finished":true}`+"\n")
	envCap := func(capS string) Env {
		return Env{Getenv: func(k string) string {
			switch k {
			case "CLAUDE_PROJECTS_DIR":
				return tmp
			case "AWAIT_SUB_AGENTS_SECONDS":
				return capS
			}
			return ""
		}}
	}
	timed := func(capS string, args ...string) (guardResult, time.Duration) {
		start := time.Now()
		r := runGuard("await-sub-agents", args, envCap(capS))
		return r, time.Since(start)
	}

	t.Run("a all finished returns 0 at once", func(t *testing.T) {
		t.Parallel()
		if r, el := timed("2", "aaa"); r.rc != 0 || el > time.Second {
			t.Fatalf("rc=%d elapsed=%v out=%q", r.rc, el, r.out)
		}
	})
	t.Run("b one still running returns 1 at the cap naming it", func(t *testing.T) {
		t.Parallel()
		r, el := timed("2", "aaa", "bbb")
		if r.rc != 1 || el < 2*time.Second || el > 5*time.Second || r.stdout != "still running: bbb\n" {
			t.Fatalf("rc=%d elapsed=%v stdout=%q", r.rc, el, r.stdout)
		}
	})
	t.Run("c finishing mid-wait returns 0 before the cap", func(t *testing.T) {
		t.Parallel()
		go func() {
			time.Sleep(500 * time.Millisecond)
			_ = os.WriteFile(filepath.Join(dir, "agent-ccc.meta.json"), []byte(`{"agentType":"Plan","finished":true}`+"\n"), 0o644)
		}()
		if r, el := timed("20", "ccc"); r.rc != 0 || el > 8*time.Second {
			t.Fatalf("rc=%d elapsed=%v out=%q", r.rc, el, r.out)
		}
	})
	t.Run("d unknown agent and no argument return 2", func(t *testing.T) {
		t.Parallel()
		if r, _ := timed("2", "zzz"); r.rc != 2 || !strings.Contains(r.err, "no meta.json for agent zzz under "+tmp) {
			t.Fatalf("unknown: rc=%d stderr=%q", r.rc, r.err)
		}
		if r, _ := timed("2", "ddd"); r.rc != 2 || !strings.Contains(r.err, "no meta.json for agent ddd") {
			t.Fatalf("outside subagents/: rc=%d stderr=%q", r.rc, r.err)
		}
		if r, _ := timed("2"); r.rc != 2 || r.err != "usage: flow-guard await-sub-agents <agent id>...\n" {
			t.Fatalf("no arg: rc=%d stderr=%q", r.rc, r.err)
		}
	})
	t.Run("e a cap above 240 is refused", func(t *testing.T) {
		t.Parallel()
		for _, c := range []string{"300", "241", "x", "-1"} {
			if r, _ := timed(c, "aaa"); r.rc != 2 || r.err != "await-sub-agents: AWAIT_SUB_AGENTS_SECONDS must be 0-240, got "+c+"\n" {
				t.Fatalf("cap %s: rc=%d stderr=%q", c, r.rc, r.err)
			}
		}
	})
}
