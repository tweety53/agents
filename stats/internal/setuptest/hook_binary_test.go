package setuptest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// hookNames is every Go hook setup.sh prints a registration snippet for.
var hookNames = []string{"enforce-agent-baseline", "enforce-loop-dispatch", "flow-active-change",
	"flow-session-activity", "flow-skill-load", "await-own-agents", "block-unsafe-commands"}

// snippetCommand matches a printed snippet's `"command": "<json string>"`.
var snippetCommand = regexp.MustCompile(`"command": ("(?:[^"\\]|\\.)*")`)

// TestHookRegistrationsRunTheBinaryDirect pins that every hook snippet setup.sh global prints
// registers the flow-guard binary it built into ~/.claude/hooks directly — no python3 shim, no
// bootstrap — that each such command exits 0 with no decision on a garbage payload while
// recording its hook-runs line, that the binary answers for the checkout it was built from
// (FLOW_GUARD_REPO_ROOT), and that a command whose binary is missing still exits 0.
func TestHookRegistrationsRunTheBinaryDirect(t *testing.T) {
	smoke.Heavy(t)
	g := newGroup(t)
	home := g.newHome()
	g.runSetup(fixture, home, "global", "")
	g.assertRCZero("global succeeds", g.rc, g.log)
	log, err := os.ReadFile(g.log)
	if err != nil {
		t.Fatal(err)
	}
	cmds := map[string]string{}
	for _, m := range snippetCommand.FindAllStringSubmatch(string(log), -1) {
		var c string
		if err := json.Unmarshal([]byte(m[1]), &c); err != nil {
			t.Fatalf("snippet command %s: %v", m[1], err)
		}
		for _, h := range hookNames {
			if strings.Contains(c, h) {
				cmds[h] = c
			}
		}
	}
	bin := filepath.Join(home, ".claude/hooks/flow-guard-hook")
	g.assertExecutable("the hook binary is built", bin)
	state := filepath.Join(g.dir, "state")
	mkdirAll(t, state)
	run := func(cmd string, stdin []byte, env ...string) (string, int) {
		t.Helper()
		c := exec.Command("/bin/sh", "-c", cmd)
		c.Dir = g.dir
		c.Stdin = bytes.NewReader(stdin)
		c.Env = append(withHome(os.Environ(), home), append([]string{"FLOW_STATE_DIR=" + state,
			"FLOW_ADDR=http://127.0.0.1:1", "FLOW_RECORDS_ADDR=http://127.0.0.1:1", "FLOW_HOOK_INCIDENT_CMD=true"}, env...)...)
		out, err := c.Output()
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return string(out), ee.ExitCode()
		} else if err != nil {
			t.Fatalf("sh -c %s: %v", cmd, err)
		}
		return string(out), 0
	}
	for _, h := range hookNames {
		cmd, ok := cmds[h]
		if !ok {
			g.fail("setup.sh prints a snippet for "+h, "none in "+g.log)
			continue
		}
		if want := `"$HOME/.claude/hooks/flow-guard-hook" hook ` + h + " || true"; cmd != want {
			g.fail(h+" registers the hook binary directly", "got "+cmd+", want "+want)
			continue
		}
		if out, rc := run(cmd, []byte("not json {")); rc != 0 || out != "" {
			g.fail(h+" fails open on garbage input", fmt.Sprintf("rc %d, stdout %q", rc, out))
		}
	}
	if b, _ := os.ReadFile(filepath.Join(state, "hook-runs.jsonl")); strings.Count(string(b), "\n") != len(hookNames) {
		g.fail("every direct run records its hook-runs line", string(b))
	}

	// enforce-loop-dispatch hands the checkout it answers for to its checker as the first argument.
	seen := filepath.Join(g.dir, "checker-args")
	checker := filepath.Join(g.dir, "checker")
	g.seedFile(checker, 0o755, "#!/bin/sh\nprintf '%s\\n' \"$1\" > '"+seen+"'\n")
	payload, _ := json.Marshal(map[string]any{"tool_name": "Agent", "cwd": g.dir, "tool_input": map[string]any{
		"prompt": "Key: self-review-x-y\n~/.claude/rules/agent-baseline.md\n"}})
	run(cmds["enforce-loop-dispatch"], payload, "LOOP_DISPATCH_CHECK="+checker, "FLOW_GUARD_REPO_ROOT=/elsewhere")
	if b, _ := os.ReadFile(seen); strings.TrimSpace(string(b)) != fixture {
		g.fail("the hook binary answers for the checkout it was built from", "checker saw "+string(b)+", want "+fixture)
	}

	if err := os.Remove(bin); err != nil {
		t.Fatal(err)
	}
	if _, rc := run(cmds["block-unsafe-commands"], []byte("{}")); rc != 0 {
		g.fail("a registration whose binary is missing exits 0", fmt.Sprintf("rc %d", rc))
	}
}
