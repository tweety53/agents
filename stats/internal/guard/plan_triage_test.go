package guard

import (
	"strings"
	"testing"
)

const ptPolicy = `"plan-triage":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}`

// ptFixture is a worktree tracking two files and a ticket beside it.
func ptFixture(t *testing.T) (wt, ticket string) {
	t.Helper()
	wt = t.TempDir()
	gitRun(t, wt, "init", "-q")
	writeFile(t, wt+"/cmd/app/main.go", "package main\n")
	writeFile(t, wt+"/README.md", "readme\n")
	gitRun(t, wt, "add", ".")
	ticket = t.TempDir() + "/ticket.md"
	writeFile(t, ticket, "KAN-9: rename the export button\n")
	return wt, ticket
}

// TestPlanTriage: the verdict line is printed as the model gave it, on the
// plan-triage row's first pair, the ticket and the tracked files on stdin
// and plan-triage-prompt.md as the system prompt.
func TestPlanTriage(t *testing.T) {
	t.Parallel()
	for answer, want := range map[string]string{
		"TRIAGE: short — one button label in one view":           "TRIAGE: short — one button label in one view",
		"Thinking.\n  TRIAGE: full -  changes the  data model\n": "TRIAGE: full — changes the data model",
	} {
		wt, ticket := ptFixture(t)
		stub := newPJStub(t, pjJSON(answer, ""))
		r := runGuard("plan-triage", []string{wt, ticket}, pjEnv(t, pjPolicy(ptPolicy), stub))
		if r.rc != 0 || !strings.HasPrefix(r.stdout, want+" (judge ") || strings.Count(r.stdout, "\n") != 1 || r.err != "" {
			t.Fatalf("rc=%d stdout=%q err=%q, want %q", r.rc, r.stdout, r.err, want)
		}
		if in := stub.stdin(t); !strings.Contains(in, "KAN-9: rename the export button") || !strings.Contains(in, "cmd/app/main.go\n") || !strings.Contains(in, "README.md\n") {
			t.Errorf("stdin %q", in)
		}
		if sp := stub.systemPrompt(t); !strings.Contains(sp, "You decide how much planning a change needs") {
			t.Errorf("system prompt is not plan-triage-prompt.md: %q", sp)
		}
		if a := strings.Join(stub.argv(t), " "); !strings.Contains(a, "--model sonnet --effort low") {
			t.Errorf("argv %q, want the plan-triage row's first pair", a)
		}
	}
}

// TestPlanTriageFallsBack: every failure prints `full` with a `fallback:`
// reason naming the cause, exit 0.
func TestPlanTriageFallsBack(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct{ policy, answer, cause string }{
		"no pair":         {pjPolicy(`"pair-judge":{"base":["opus/flow-low"],"optin":[],"marked":[]}`), pjJSON("TRIAGE: short — x", ""), "no plan-triage pair"},
		"command fails":   {pjPolicy(ptPolicy), "exit 3\n", "exit 3"},
		"no verdict line": {pjPolicy(ptPolicy), pjJSON("short, I think", ""), "no TRIAGE line"},
		"unknown verdict": {pjPolicy(ptPolicy), pjJSON("TRIAGE: medium — x", ""), "no TRIAGE line"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			wt, ticket := ptFixture(t)
			r := runGuard("plan-triage", []string{wt, ticket}, pjEnv(t, c.policy, newPJStub(t, c.answer)))
			if r.rc != 0 || !strings.HasPrefix(r.stdout, "TRIAGE: full — fallback: ") || !strings.Contains(r.stdout, c.cause) || strings.Count(r.stdout, "\n") != 1 {
				t.Fatalf("rc=%d stdout=%q err=%q, want the fallback naming %q", r.rc, r.stdout, r.err, c.cause)
			}
		})
	}
	t.Run("ticket unreadable", func(t *testing.T) {
		t.Parallel()
		wt, _ := ptFixture(t)
		stub := newPJStub(t, pjJSON("TRIAGE: short — x", ""))
		r := runGuard("plan-triage", []string{wt, wt + "/missing.md"}, pjEnv(t, pjPolicy(ptPolicy), stub))
		if r.rc != 0 || !strings.HasPrefix(r.stdout, "TRIAGE: full — fallback: ") || stub.ran() {
			t.Fatalf("rc=%d stdout=%q ran=%v, want the fallback without a call", r.rc, r.stdout, stub.ran())
		}
	})
	t.Run("not a repository", func(t *testing.T) {
		t.Parallel()
		_, ticket := ptFixture(t)
		stub := newPJStub(t, pjJSON("TRIAGE: short — x", ""))
		r := runGuard("plan-triage", []string{t.TempDir(), ticket}, pjEnv(t, pjPolicy(ptPolicy), stub))
		if r.rc != 0 || !strings.HasPrefix(r.stdout, "TRIAGE: full — fallback: git ls-files") || stub.ran() {
			t.Fatalf("rc=%d stdout=%q ran=%v, want the fallback without a call", r.rc, r.stdout, stub.ran())
		}
	})
}

// TestPlanTriageUsage: a wrong argument count exits 2, nothing on stdout.
func TestPlanTriageUsage(t *testing.T) {
	t.Parallel()
	if r := runGuard("plan-triage", []string{"only-one"}, Env{}); r.rc != 2 || r.stdout != "" {
		t.Fatalf("rc=%d stdout=%q, want 2 and nothing", r.rc, r.stdout)
	}
}
