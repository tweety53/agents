package hook

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// TestHookInProcessBoundsAHungGuard: a guard the hook calls in process that
// never returns no longer hangs the hook -- hookInProcess gives up at its
// bound -- while one that returns hands back its output and exit code.
func TestHookInProcessBoundsAHungGuard(t *testing.T) {
	t.Parallel()
	env := core.Env{Getenv: os.Getenv}
	hung := func([]string, core.Env, io.Writer, io.Writer) int { select {} }
	start := time.Now()
	if _, _, _, err := hookInProcess(env, 50*time.Millisecond, t.TempDir(), hung); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err %v, want a deadline", err)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Fatalf("returned after %v", d)
	}
	dir := t.TempDir()
	done := func(args []string, env core.Env, stdout, stderr io.Writer) int {
		_, _ = io.WriteString(stdout, env.Dir+" "+args[0])
		_, _ = io.WriteString(stderr, "e")
		return 3
	}
	out, errb, code, err := hookInProcess(env, time.Minute, dir, done, "a")
	if err != nil || out != dir+" a" || errb != "e" || code != 3 {
		t.Fatalf("out %q err %q code %d %v", out, errb, code, err)
	}
}

// TestHookEnforceLoopDispatchInProcess: with no LOOP_DISPATCH_CHECK the hook
// runs check-loop-dispatch in process, not as scripts/flow-guard: a checkout
// with no scripts/flow-guard still gets the checker's own refusal.
func TestHookEnforceLoopDispatchInProcess(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	stdin := `{"tool_name":"Agent","tool_input":{"prompt":"Key: self-review-a-b\n~/.claude/rules/agent-baseline.md"}}`
	out, errOut, rc := runHookStreams(t, "enforce-loop-dispatch", stdin, "FLOW_GUARD_REPO_ROOT="+repo)
	want := "enforce-loop-dispatch: self-review-a-b not recorded (flow-guard check-loop-dispatch exit 2):\n"
	if rc != 0 || out != "" || !strings.HasPrefix(errOut, want) || !strings.Contains(errOut, "is not a git checkout") {
		t.Fatalf("rc=%d out=%q stderr=%q", rc, out, errOut)
	}
}
