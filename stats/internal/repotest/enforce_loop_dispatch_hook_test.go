package repotest

// TestEnforceLoopDispatchHook ports the retired shell harness,
// the harness for hooks/enforce-loop-dispatch.py. It points the hook at a stub
// checker through LOOP_DISPATCH_CHECK — the guard itself is covered by
// stats/internal/guard/check_loop_dispatch_test.go — and pipes PreToolUse
// payloads to it, asserting the hook never denies and what the checker was
// handed. Each subtest writes its own stub into t.TempDir(). Cases:
//
//	 1. a dispatch with no loop key                    -> allow, checker never run
//	 2. a loop key the checker records (exit 0)        -> allow, checker got
//	                                                       <repo> <key> <type> <model>
//	 3. a loop key the checker cannot record (exit 2)  -> allow, logged on stderr
//	 4. a loop key whose checker cannot run            -> allow, logged on stderr
//	 5. malformed JSON                                  -> allow, exit 0
//	 6. a Bash call carrying a loop key                 -> allow, checker never run
//	 7. a prompt naming a loop worktree, no Key line    -> allow, checker never run
//	 8. a loop key with no agent-baseline pointer       -> allow, checker never run
//	 9. a Key line with trailing text                   -> allow, checker never run
//	10. a pipeline-fix key with a non-numeric <k>       -> allow, checker never run
//	11. a payload carrying a cwd                        -> the checker runs in it

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// loopStub is one subtest's stub checker: it writes its argv to log and
// `pwd -P` to cwdFile, prints a line, and exits with the given code.
type loopStub struct{ dir, check, log, cwdFile string }

func newLoopStub(t *testing.T, rc string) loopStub {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := loopStub{dir: dir, check: dir + "/check", log: dir + "/args", cwdFile: dir + "/cwd"}
	body := "#!/usr/bin/env bash\nprintf '%s\\n' \"$@\" >\"" + s.log + "\"\npwd -P >\"" + s.cwdFile +
		"\"\necho \"stub output\"\nexit " + rc + "\n"
	if err := os.WriteFile(s.check, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestEnforceLoopDispatchHook(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	repo, err := filepath.EvalSymlinks(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	const pointer = "Before anything else, read `~/.claude/rules/agent-baseline.md`.\n"
	const loopPrompt = pointer + "Key: pipeline-fix-6-fix\nDo it."

	payload := func(tool, prompt, cwd string) string {
		p := map[string]any{"tool_name": tool, "tool_input": map[string]string{"prompt": prompt, "subagent_type": "flow-low", "model": "opus"}}
		if cwd != "" {
			p["cwd"] = cwd
		}
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	// hook runs the hook with check as LOOP_DISPATCH_CHECK; it must exit 0
	// and print nothing -- an allow -- and returns its stderr.
	hook := func(t *testing.T, check, stdin string) string {
		t.Helper()
		out, errOut, rc := runHook(t, "enforce-loop-dispatch.py", []byte(stdin), "LOOP_DISPATCH_CHECK="+check)
		if rc != 0 || strings.TrimSpace(out) != "" {
			t.Fatalf("expected an allow, got exit %d: %s", rc, out)
		}
		return errOut
	}

	for _, c := range []struct {
		name, rc, stdin string
		missing, run    bool
		needle          string // "" asserts nothing logged
	}{
		{name: "1 no loop key", rc: "0", stdin: payload("Agent", "Key: task-3-implementer\nwork", "")},
		{name: "2 recorded", rc: "0", stdin: payload("Agent", loopPrompt, ""), run: true},
		{name: "3 cannot record", rc: "2", stdin: payload("Task", loopPrompt, ""), run: true, needle: "exit 2"},
		{name: "4 checker cannot run", rc: "0", missing: true, stdin: payload("Agent", loopPrompt, ""), needle: "could not run"},
		{name: "5 malformed JSON", rc: "2", stdin: "not json"},
		{name: "6 not a dispatch", rc: "0", stdin: `{"tool_name":"Bash","tool_input":{"command":"echo","prompt":"Key: pipeline-fix-6-fix"}}`},
		{name: "7 loop worktree named, no Key line", rc: "0", stdin: payload("Agent", pointer+"Fix it in /x/agents-worktrees/pipeline-fix-kan-1-demo-6 now.", "")},
		{name: "8 no baseline pointer", rc: "0", stdin: payload("Agent", "Key: pipeline-fix-6-fix\nDo it.", "")},
		{name: "9 Key line with trailing text", rc: "0", stdin: payload("Agent", pointer+"Key: pipeline-fix-6-fix (round 2)\nDo it.", "")},
		{name: "10 non-numeric <k>", rc: "0", stdin: payload("Agent", pointer+"Key: pipeline-fix-x-fix\nDo it.", "")},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := newLoopStub(t, c.rc)
			check := s.check
			if c.missing {
				check = s.dir + "/missing"
			}
			errOut := hook(t, check, c.stdin)
			if _, err := os.Stat(s.log); (err == nil) != c.run {
				t.Errorf("checker ran: %v, want %v", err == nil, c.run)
			}
			if c.needle == "" && errOut != "" {
				t.Errorf("logged: %s", errOut)
			}
			if c.needle != "" && (!strings.Contains(errOut, "not recorded") || !strings.Contains(errOut, c.needle)) {
				t.Errorf("stderr carries no %q: %s", c.needle, errOut)
			}
			if c.name == "2 recorded" {
				b, _ := os.ReadFile(s.log)
				if got := strings.ReplaceAll(string(b), "\n", " "); got != repo+" pipeline-fix-6-fix flow-low opus " {
					t.Errorf("checker args: %s", got)
				}
			}
		})
	}
	t.Run("11 checker ran in the payload's cwd", func(t *testing.T) {
		t.Parallel()
		s := newLoopStub(t, "0")
		project := s.dir + "/project"
		if err := os.Mkdir(project, 0o755); err != nil {
			t.Fatal(err)
		}
		hook(t, s.check, payload("Agent", loopPrompt, project))
		if b, _ := os.ReadFile(s.cwdFile); strings.TrimSpace(string(b)) != project {
			t.Errorf("checker ran in %s", b)
		}
	})
}
