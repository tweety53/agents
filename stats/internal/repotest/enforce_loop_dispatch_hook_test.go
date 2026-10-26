package repotest

// TestEnforceLoopDispatchHook ports the retired shell harness,
// the harness for hooks/enforce-loop-dispatch.py. It points the hook at a stub
// checker through LOOP_DISPATCH_CHECK — the guard itself is covered by
// stats/internal/guard/check_loop_dispatch_test.go — and pipes PreToolUse
// payloads to it, asserting the decision and what the checker was handed. Each
// subtest writes its own stub into t.TempDir(). Cases:
//
//	 1. a dispatch with no loop key                    -> allow, checker never run
//	 2. a loop key the checker admits (exit 0)          -> allow, checker got
//	                                                       <repo> <key> <type> <model>
//	 3. a loop key the checker refuses (exit 1)         -> deny, its lines in the reason
//	 4. a loop key the checker cannot answer (exit 2)   -> deny
//	 5. a loop key whose checker cannot run             -> deny (fails closed)
//	 6. malformed JSON                                  -> allow, exit 0
//	 7. a Bash call carrying a loop key                 -> allow, checker never run
//	 8. a prompt naming a loop worktree, no Key line    -> deny, checker never run
//	 9. a loop key with no agent-baseline pointer       -> deny, checker never run
//	10. a Key line with trailing text                   -> deny, checker never run
//	11. a pipeline-fix key with a non-numeric <k>       -> deny, checker never run
//	12. a payload carrying a cwd                        -> the checker runs in it
//	13. a prompt naming a /flow change called self-review-<x>, no Key line, its
//	    spectre/<name> branch in the payload's cwd    -> allow, checker never run;
//	    the same prompt in a cwd with the loop's open branch but no spectre one -> deny
//	14. a prompt naming a closed loop (no branch left), no Key line -> allow
//	15. a planner's Explore/Plan sub-dispatch naming an open loop's worktree -> allow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loopStub is one subtest's stub checker: it writes its argv to log and
// `pwd -P` to cwdFile, prints a refusal line, and exits with the given code.
type loopStub struct{ dir, check, log, cwdFile string }

func newLoopStub(t *testing.T, rc string) loopStub {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := loopStub{dir: dir, check: dir + "/check", log: dir + "/args", cwdFile: dir + "/cwd"}
	if err := os.WriteFile(dir+"/rc", []byte(rc+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := "#!/usr/bin/env bash\nprintf '%s\\n' \"$@\" >\"" + s.log + "\"\npwd -P >\"" + s.cwdFile +
		"\"\necho \"LOOP-DISPATCH-REFUSED: PAIR — stub\"\nexit \"$(cat \"" + dir + "/rc\")\"\n"
	if err := os.WriteFile(s.check, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestEnforceLoopDispatchHook(t *testing.T) {
	t.Parallel()
	repo, err := filepath.EvalSymlinks(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	const loopPrompt = "Before anything else, read `~/.claude/rules/agent-baseline.md`.\nKey: pipeline-fix-6-fix\nDo it."
	const flowPrompt = "Review /x/agents-worktrees/self-review-backlog-demo-5 now."

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
	// hook runs the hook with check as LOOP_DISPATCH_CHECK; a non-zero exit fails the subtest.
	hook := func(t *testing.T, check, stdin string) string {
		t.Helper()
		out, _, rc := runHook(t, "enforce-loop-dispatch.py", []byte(stdin), "LOOP_DISPATCH_CHECK="+check)
		if rc != 0 {
			t.Fatalf("exit %d", rc)
		}
		return strings.TrimRight(out, "\n")
	}
	allow := func(t *testing.T, out string) {
		t.Helper()
		if out != "" {
			t.Errorf("expected allow, got %s", out)
		}
	}
	deny := func(t *testing.T, out, needle string) {
		t.Helper()
		if !strings.Contains(out, `"permissionDecision": "deny"`) || !strings.Contains(out, needle) {
			t.Errorf("expected a deny carrying %q, got %s", needle, out)
		}
	}
	notRun := func(t *testing.T, s loopStub) {
		t.Helper()
		if _, err := os.Stat(s.log); err == nil {
			t.Error("checker ran")
		}
	}
	// flowRepo is a repo whose cwd optionally holds the self-review change's spectre branch.
	flowRepo := func(t *testing.T, withBranch bool, loops ...string) string {
		d := t.TempDir() + "/flowrepo"
		env := []string{"HOME=" + t.TempDir()}
		sh(t, "", env, "git", "init", "-q", "-b", "main", d)
		sh(t, "", env, "git", "-C", d, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "commit", "-q", "--allow-empty", "-m", "init")
		if withBranch {
			sh(t, "", env, "git", "-C", d, "branch", "spectre/self-review-backlog-demo-5")
		}
		for _, l := range loops {
			sh(t, "", env, "git", "-C", d, "branch", l)
		}
		return d
	}
	bashLoop := `{"tool_name":"Bash","tool_input":{"command":"echo","prompt":"Key: pipeline-fix-6-fix"}}`
	pointer := "Before anything else, read `~/.claude/rules/agent-baseline.md`.\n"

	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"1 no loop key", func(t *testing.T) {
			allow(t, hook(t, newLoopStub(t, "0").check, payload("Agent", "Key: task-3-implementer\nwork", "")))
		}},
		{"1 checker not run", func(t *testing.T) {
			s := newLoopStub(t, "0")
			hook(t, s.check, payload("Agent", "Key: task-3-implementer\nwork", ""))
			notRun(t, s)
		}},
		{"2 admitted", func(t *testing.T) {
			allow(t, hook(t, newLoopStub(t, "0").check, payload("Agent", loopPrompt, "")))
		}},
		{"2 checker args", func(t *testing.T) {
			s := newLoopStub(t, "0")
			hook(t, s.check, payload("Agent", loopPrompt, ""))
			b, _ := os.ReadFile(s.log)
			if got := strings.ReplaceAll(string(b), "\n", " "); got != repo+" pipeline-fix-6-fix flow-low opus " {
				t.Errorf("checker args: %s", got)
			}
		}},
		{"3 refused", func(t *testing.T) {
			deny(t, hook(t, newLoopStub(t, "1").check, payload("Task", loopPrompt, "")), "LOOP-DISPATCH-REFUSED: PAIR")
		}},
		{"4 cannot answer", func(t *testing.T) {
			deny(t, hook(t, newLoopStub(t, "2").check, payload("Agent", loopPrompt, "")), "exit 2")
		}},
		{"5 checker cannot run", func(t *testing.T) {
			deny(t, hook(t, newLoopStub(t, "2").dir+"/missing", payload("Agent", loopPrompt, "")), "could not run")
		}},
		{"6 malformed JSON", func(t *testing.T) {
			allow(t, hook(t, newLoopStub(t, "2").check, "not json"))
		}},
		{"7 not a dispatch", func(t *testing.T) {
			allow(t, hook(t, newLoopStub(t, "1").check, bashLoop))
		}},
		{"7 checker not run", func(t *testing.T) {
			s := newLoopStub(t, "1")
			hook(t, s.check, bashLoop)
			notRun(t, s)
		}},
		{"8 loop worktree named, no Key line", func(t *testing.T) {
			deny(t, hook(t, newLoopStub(t, "0").check, payload("Agent", "Fix it in /x/agents-worktrees/pipeline-fix-kan-1-demo-6 now.", flowRepo(t, false, "pipeline-fix-kan-1-demo-6"))), "carries no")
		}},
		{"8 checker not run", func(t *testing.T) {
			s := newLoopStub(t, "0")
			hook(t, s.check, payload("Agent", "Fix it in /x/agents-worktrees/pipeline-fix-kan-1-demo-6 now.", ""))
			notRun(t, s)
		}},
		{"9 no baseline pointer", func(t *testing.T) {
			deny(t, hook(t, newLoopStub(t, "0").check, payload("Agent", "Key: pipeline-fix-6-fix\nDo it.", "")), "agent-baseline.md pointer")
		}},
		{"9 checker not run", func(t *testing.T) {
			s := newLoopStub(t, "0")
			hook(t, s.check, payload("Agent", "Key: pipeline-fix-6-fix\nDo it.", ""))
			notRun(t, s)
		}},
		{"10 Key line with trailing text", func(t *testing.T) {
			deny(t, hook(t, newLoopStub(t, "0").check, payload("Agent", pointer+"Key: pipeline-fix-6-fix (round 2)\nDo it.", "")), "no valid loop key")
		}},
		{"10 checker not run", func(t *testing.T) {
			s := newLoopStub(t, "0")
			hook(t, s.check, payload("Agent", pointer+"Key: pipeline-fix-6-fix (round 2)\nDo it.", ""))
			notRun(t, s)
		}},
		{"11 non-numeric <k>", func(t *testing.T) {
			deny(t, hook(t, newLoopStub(t, "0").check, payload("Agent", pointer+"Key: pipeline-fix-x-fix\nDo it.", "")), "no valid loop key")
		}},
		{"11 checker not run", func(t *testing.T) {
			s := newLoopStub(t, "0")
			hook(t, s.check, payload("Agent", pointer+"Key: pipeline-fix-x-fix\nDo it.", ""))
			notRun(t, s)
		}},
		{"12 admitted with a cwd", func(t *testing.T) {
			s := newLoopStub(t, "0")
			project := s.dir + "/project"
			if err := os.Mkdir(project, 0o755); err != nil {
				t.Fatal(err)
			}
			allow(t, hook(t, s.check, payload("Agent", loopPrompt, project)))
		}},
		{"12 checker ran in the payload's cwd", func(t *testing.T) {
			s := newLoopStub(t, "0")
			project := s.dir + "/project"
			if err := os.Mkdir(project, 0o755); err != nil {
				t.Fatal(err)
			}
			hook(t, s.check, payload("Agent", loopPrompt, project))
			if b, _ := os.ReadFile(s.cwdFile); strings.TrimSpace(string(b)) != project {
				t.Errorf("checker ran in %s", b)
			}
		}},
		{"13 self-review-named change, no spectre branch", func(t *testing.T) {
			deny(t, hook(t, newLoopStub(t, "0").check, payload("Agent", flowPrompt, flowRepo(t, false, "self-review-backlog-demo-5"))), "carries no")
		}},
		{"13 self-review-named /flow change", func(t *testing.T) {
			allow(t, hook(t, newLoopStub(t, "0").check, payload("Agent", flowPrompt, flowRepo(t, true))))
		}},
		{"14 closed loop named in prose", func(t *testing.T) {
			allow(t, hook(t, newLoopStub(t, "0").check, payload("Agent", "The loop self-review-kan-9-demo finished; summarise it.", flowRepo(t, false))))
		}},
		{"15 planner Explore sub-dispatch", func(t *testing.T) {
			p := map[string]any{"tool_name": "Agent", "cwd": flowRepo(t, false, "pipeline-fix-kan-1-demo-6"), "tool_input": map[string]string{
				"prompt": "Map /x/agents-worktrees/pipeline-fix-kan-1-demo-6 for the plan.", "subagent_type": "Explore"}}
			b, _ := json.Marshal(p)
			allow(t, hook(t, newLoopStub(t, "0").check, string(b)))
		}},
		{"13 checker not run", func(t *testing.T) {
			s := newLoopStub(t, "0")
			hook(t, s.check, payload("Agent", flowPrompt, flowRepo(t, true)))
			notRun(t, s)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t)
		})
	}
}
