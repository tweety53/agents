package guard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fsjFixture is a repository with one commit, the since sha, and an
// uncommitted fix on top; the fix instructions and a decision of three
// bundles beside it.
func fsjFixture(t *testing.T) (wt, since, instr, decision string) {
	t.Helper()
	wt = t.TempDir()
	gitRun(t, wt, "init", "-q")
	writeFile(t, wt+"/a.go", "package a\n")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-qm", "base")
	sha, err := exec.Command("git", "-C", wt, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, wt+"/a.go", "package a\n\nfunc Fixed() {}\n")
	dir := t.TempDir()
	instr = filepath.Join(dir, "fix-instructions-1.md")
	writeFile(t, instr, "rename the helper")
	decision = filepath.Join(dir, "decision.json")
	writeFile(t, decision, `{"panel":[{"slots":["primary","principles"]},{"slots":["mutation"]},{"slots":["test-sanity"]}]}`)
	return wt, strings.TrimSpace(string(sha)), instr, decision
}

const fsjPolicy = `"fix-run-judge":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}`

// TestFixSlotJudge: the judge's answer prints one line per decided slot,
// primary always run and never put to the judge; the lines are saved beside
// the instructions; the judge reads the instructions and the delta, and its
// prompt lists each put slot with its catalog line.
func TestFixSlotJudge(t *testing.T) {
	t.Parallel()
	wt, since, instr, decision := fsjFixture(t)
	stub := newPJStub(t, pjJSON("principles: skip — a rename\nmutation: run — new logic\ntest-sanity: skip — no test changed", ""))
	r := runGuard("fix-slot-judge", []string{wt, since, instr, decision}, pjEnv(t, pjPolicy(fsjPolicy), stub))
	want := "primary: run — always runs\nprinciples: skip — a rename\nmutation: run — new logic\ntest-sanity: skip — no test changed\n"
	if r.rc != 0 || r.stdout != want {
		t.Fatalf("rc=%d stdout=%q err=%q, want %q", r.rc, r.stdout, r.err, want)
	}
	if saved, _ := os.ReadFile(strings.TrimSuffix(instr, ".md") + ".slots"); string(saved) != want {
		t.Errorf("saved %q, want %q", saved, want)
	}
	if in := stub.stdin(t); !strings.Contains(in, "rename the helper") || !strings.Contains(in, "+func Fixed() {}") || !strings.Contains(in, "a.go | ") {
		t.Errorf("stdin %q lacks the instructions, the diff or its stat", in)
	}
	sp := stub.systemPrompt(t)
	if !strings.Contains(sp, "- mutation — sabotage-proofing") || strings.Contains(sp, "- primary") {
		t.Errorf("system prompt %q: want each put slot with its catalog line, primary not put", sp)
	}
	if a := stub.argv(t); !strings.Contains(strings.Join(a, " "), "--model sonnet --effort low") {
		t.Errorf("argv %q, want the fix-run-judge row's first pair", a)
	}
}

// TestFixSlotJudgeFallsBack: any judge failure runs every decided slot,
// exit 0, the cause on one stderr line.
func TestFixSlotJudgeFallsBack(t *testing.T) {
	t.Parallel()
	all := "primary: run — always runs\nprinciples: run — judge fallback\nmutation: run — judge fallback\ntest-sanity: run — judge fallback\n"
	for name, c := range map[string]struct{ policy, answer, cause string }{
		"no pair":         {pjPolicy(`"pair-judge":{"base":["opus/flow-low"],"optin":[],"marked":[]}`), pjJSON("", ""), "no fix-run-judge pair"},
		"not JSON":        {pjPolicy(fsjPolicy), "echo nope\n", "not JSON"},
		"unknown slot":    {pjPolicy(fsjPolicy), pjJSON("principles: skip — x\nmutation: run — x\ntest-sanity: run — x\ncall-efficiency: run — x", ""), "was not put to it"},
		"slot unanswered": {pjPolicy(fsjPolicy), pjJSON("principles: skip — x\nmutation: run — x", ""), "left test-sanity unanswered"},
		"command fails":   {pjPolicy(fsjPolicy), "exit 3\n", "exit 3"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			wt, since, instr, decision := fsjFixture(t)
			r := runGuard("fix-slot-judge", []string{wt, since, instr, decision}, pjEnv(t, c.policy, newPJStub(t, c.answer)))
			if r.rc != 0 || r.stdout != all || !strings.Contains(r.err, c.cause) || strings.Count(r.err, "\n") != 1 {
				t.Fatalf("rc=%d stdout=%q err=%q, want 0, every slot run and one stderr line naming %q", r.rc, r.stdout, r.err, c.cause)
			}
		})
	}
}

// TestFixSlotJudgeRefuses: usage, an unreadable instructions file and a
// decision naming no slot exit 2 with nothing on stdout; a primary-only
// decision runs no judge.
func TestFixSlotJudgeRefuses(t *testing.T) {
	t.Parallel()
	wt, since, instr, decision := fsjFixture(t)
	stub := newPJStub(t, pjJSON("", ""))
	env := pjEnv(t, pjPolicy(fsjPolicy), stub)
	empty := filepath.Join(t.TempDir(), "decision.json")
	writeFile(t, empty, `{"panel":"default"}`)
	for name, args := range map[string][]string{
		"usage":                   {wt, since, instr},
		"instructions unreadable": {wt, since, instr + ".absent", decision},
		"decision names no slot":  {wt, since, instr, empty},
	} {
		if r := runGuard("fix-slot-judge", args, env); r.rc != 2 || r.stdout != "" {
			t.Errorf("%s: rc=%d stdout=%q, want 2 and nothing", name, r.rc, r.stdout)
		}
	}
	writeFile(t, decision, `{"panel":[{"slots":["primary"]}]}`)
	if r := runGuard("fix-slot-judge", []string{wt, since, instr, decision}, env); r.rc != 0 || r.stdout != "primary: run — always runs\n" || stub.ran() {
		t.Errorf("primary only: rc=%d stdout=%q ran=%v, want primary run and no judge", r.rc, r.stdout, stub.ran())
	}
}
