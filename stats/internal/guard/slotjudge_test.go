package guard

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// sjSlots are the three gated slots, in the order slot-judge prints them.
var sjSlots = []string{"test-sanity", "call-efficiency", "failure-modes"}

// TestSlotJudgeTakesTheVerdicts pins decisions
// slot-judge-reuses-pair-judge-call, slot-judge-same-output-shape,
// slot-judge-three-slots and slot-judge-at-plan-time: the judge runs on the
// pair-judge row's first pair with the digest on stdin and a system prompt
// naming the three slots with the ticket's conditions, and its answer --
// whatever its order -- prints in the fixed order, each reason ending
// `; judge` and the spend.
func TestSlotJudgeTakesTheVerdicts(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, pjJSON("failure-modes: skip — no I/O\ntest-sanity: skip — no tests\ncall-efficiency: run — adds a query", `,"total_cost_usd":0.0012`))
	const digest = "inputs: tasks=1 files=1\nTask 1: add a query to the store.\n"
	r := runGuard("slot-judge", []string{pjDigest(t, digest)}, pjEnv(t, pjPolicy(), stub))
	if r.rc != 0 {
		t.Fatalf("rc=%d out=%s", r.rc, r.out)
	}
	lines := strings.Split(strings.TrimSuffix(r.stdout, "\n"), "\n")
	want := []string{"test-sanity: skip — no tests; judge (judge ", "call-efficiency: run — adds a query; judge (judge ", "failure-modes: skip — no I/O; judge (judge "}
	if len(lines) != 3 {
		t.Fatalf("stdout=%q, want three lines", r.stdout)
	}
	for i, w := range want {
		if !strings.HasPrefix(lines[i], w) || !strings.HasSuffix(lines[i], ", $0.0012)") {
			t.Errorf("line %d = %q, want it opening %q and ending with the spend", i, lines[i], w)
		}
	}
	argv := strings.Join(stub.argv(t), "\x00")
	if !strings.Contains(argv, "--model\x00sonnet\x00--effort\x00low") {
		t.Errorf("argv %q lacks --model sonnet --effort low", stub.argv(t))
	}
	if got := stub.stdin(t); got != digest {
		t.Errorf("stdin = %q, want the digest %q", got, digest)
	}
	prompt := stub.systemPrompt(t)
	for _, w := range []string{
		"test-sanity", "tests are added or changed",
		"call-efficiency", "request handlers, queries or client calls",
		"failure-modes", "I/O, persistence or concurrency",
	} {
		if !strings.Contains(prompt, w) {
			t.Errorf("system prompt lacks %q:\n%s", w, prompt)
		}
	}
}

// TestSlotJudgeFallsBack pins decision slot-judge-fail-open: every judge
// failure and every malformed answer prints all three slots as `run —
// fallback: <cause>` and exits 0.
func TestSlotJudgeFallsBack(t *testing.T) {
	t.Parallel()
	full := "test-sanity: run — a\ncall-efficiency: skip — b\nfailure-modes: skip — c"
	cases := []struct {
		name    string
		answer  string
		policy  string
		missing bool
		timeout time.Duration
		cause   string
	}{
		{name: "non-zero exit", answer: "exit 1\n", policy: pjPolicy(), cause: "exit 1"},
		{name: "timeout", answer: "exec sleep 5\n", policy: pjPolicy(), timeout: 200 * time.Millisecond, cause: "timed out"},
		{name: "non-JSON", answer: "echo not json\n", policy: pjPolicy(), cause: "not JSON"},
		{name: "missing command", missing: true, policy: pjPolicy(), cause: "not runnable"},
		{name: "no pair-judge row", answer: pjJSON(full, ""), policy: "{}", cause: "no pair-judge pair"},
		{name: "a slot missing", answer: pjJSON("test-sanity: run — a\ncall-efficiency: skip — b", ""), policy: pjPolicy(), cause: "malformed"},
		{name: "a slot twice", answer: pjJSON(full+"\ntest-sanity: skip — d", ""), policy: pjPolicy(), cause: "malformed"},
		{name: "a verdict outside run|skip", answer: pjJSON("test-sanity: maybe — a\ncall-efficiency: skip — b\nfailure-modes: skip — c", ""), policy: pjPolicy(), cause: "malformed"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			stub := newPJStub(t, c.answer)
			env := pjEnv(t, c.policy, stub)
			if c.missing {
				env = policyFlowEnv(t, c.policy, "FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.dir+"/absent")
			}
			env.PairJudgeTimeout = c.timeout
			r := runGuard("slot-judge", []string{pjDigest(t, "a plan")}, env)
			lines := strings.Split(strings.TrimSuffix(r.stdout, "\n"), "\n")
			if r.rc != 0 || len(lines) != 3 {
				t.Fatalf("rc=%d stdout=%q, want 0 and three lines", r.rc, r.stdout)
			}
			for i, s := range sjSlots {
				if !strings.HasPrefix(lines[i], s+": run — fallback: ") || !strings.Contains(lines[i], c.cause) {
					t.Errorf("line %d = %q, want `%s: run — fallback: …%s…`", i, lines[i], s, c.cause)
				}
			}
			if c.name == "no pair-judge row" && stub.ran() {
				t.Error("the judge ran with no pair-judge pair to run on")
			}
		})
	}
}

// TestSlotJudgeToleratesProse pins the contract's malformed rule: only a
// line opening with a slot name is parsed, so prose around the three
// verdict lines -- a colon in it too -- does not make the answer malformed.
func TestSlotJudgeToleratesProse(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, pjJSON("Here are the verdicts:\ntest-sanity: run — a\ncall-efficiency: skip — b\nfailure-modes: skip — c\nNote: that is all.", ""))
	r := runGuard("slot-judge", []string{pjDigest(t, "a plan")}, pjEnv(t, pjPolicy(), stub))
	lines := strings.Split(strings.TrimSuffix(r.stdout, "\n"), "\n")
	want := []string{"test-sanity: run — a; judge", "call-efficiency: skip — b; judge", "failure-modes: skip — c; judge"}
	if r.rc != 0 || len(lines) != 3 {
		t.Fatalf("rc=%d stdout=%q, want 0 and three lines", r.rc, r.stdout)
	}
	for i, w := range want {
		if !strings.HasPrefix(lines[i], w) {
			t.Errorf("line %d = %q, want it opening %q", i, lines[i], w)
		}
	}
}

// TestSlotJudgeRefuses: usage and an unreadable digest exit 2 with nothing
// on stdout.
func TestSlotJudgeRefuses(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, "exit 1\n")
	env := pjEnv(t, pjPolicy(), stub)
	digest := pjDigest(t, "a plan")
	for _, args := range [][]string{
		nil,
		{digest, digest},
		{filepath.Join(t.TempDir(), "absent.md")},
	} {
		r := runGuard("slot-judge", args, env)
		if r.rc != 2 || r.stdout != "" {
			t.Errorf("%q: rc=%d stdout=%q, want 2 and nothing on stdout", args, r.rc, r.stdout)
		}
	}
	if stub.ran() {
		t.Error("a refused call ran the judge")
	}
}
