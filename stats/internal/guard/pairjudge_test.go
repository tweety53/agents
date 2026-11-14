package guard

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

// pjJudgeRow is the pair-judge role's seeded row, as 0065 seeds it.
const pjJudgeRow = `"pair-judge":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}`

// pjPolicy is a `flow policy get` answer carrying the pair-judge row beside
// the given role rows.
func pjPolicy(rows ...string) string {
	return "{" + strings.Join(append(rows, pjJudgeRow), ",") + "}"
}

// pjStub is a judge command in a fresh directory: it saves its argv
// (NUL-separated) to argv and its stdin to stdin, then runs answer.
type pjStub struct{ dir string }

func newPJStub(t *testing.T, answer string) pjStub {
	t.Helper()
	s := pjStub{t.TempDir()}
	writeExec(t, s.cmd(), "#!/bin/sh\nprintf '%s\\0' \"$@\" > '"+s.dir+"/argv'\n/bin/cat > '"+s.dir+"/stdin'\n"+answer)
	return s
}

func (s pjStub) cmd() string { return s.dir + "/judge" }

// ran is whether the stub was invoked at all.
func (s pjStub) ran() bool {
	_, err := os.Stat(s.dir + "/argv")
	return err == nil
}

func (s pjStub) argv(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(s.dir + "/argv")
	if err != nil {
		t.Fatalf("the judge was never run: %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(b), "\x00"), "\x00")
}

func (s pjStub) stdin(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(s.dir + "/stdin")
	if err != nil {
		t.Fatalf("the judge's stdin was never saved: %v", err)
	}
	return string(b)
}

// systemPrompt is the value the stub was given after --system-prompt.
func (s pjStub) systemPrompt(t *testing.T) string {
	t.Helper()
	a := s.argv(t)
	i := slices.Index(a, "--system-prompt")
	if i < 0 || i+1 >= len(a) {
		t.Fatalf("argv %q carries no --system-prompt value", a)
	}
	return a[i+1]
}

var pjCandidateLine = regexp.MustCompile(`(?m)^- (\S+)$`)

// candidates are the pairs the stub's system prompt lists, in order.
func (s pjStub) candidates(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, m := range pjCandidateLine.FindAllStringSubmatch(s.systemPrompt(t), -1) {
		out = append(out, m[1])
	}
	return out
}

const pjAnswer = "printf 'PAIR: sonnet/flow-low\\nREASON: renames only\\n'\n"

// pjRepoRoot is the agents checkout these tests run in, whose roles table
// the guard reads the Does line from.
func pjRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// pjEnv is an env whose `flow` answers policy with policy, whose judge
// command is stub, and whose repository root is this checkout.
func pjEnv(t *testing.T, policy string, stub pjStub) Env {
	t.Helper()
	return policyFlowEnv(t, policy, "FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.cmd())
}

// pjDigest writes body to a digest file and returns its path.
func pjDigest(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "digest.md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestPairJudgeCandidates: the candidates are the role's base pairs, the
// overlay applied, less xhigh and max, less high for the implementer and
// every fixer, less every non-opus pair under -opus-only unless that leaves
// none; marked and opt-in pairs are never listed. Each pair shape prints
// parsed into model and effort.
func TestPairJudgeCandidates(t *testing.T) {
	t.Parallel()
	const row = `{"base":["opus/flow-low","opus/flow-medium","opus/flow-high","sonnet/flow-low","opus/flow-xhigh"],"optin":["sonnet/flow-medium"],"marked":[{"pair":"opus/flow-max","marker":"No-progress escalation"}]}`
	cases := []struct {
		name   string
		policy string
		args   []string
		want   []string
	}{
		{"implementer drops high and xhigh", pjPolicy(`"implementer":` + row), []string{"implementer"}, []string{"opus/flow-low", "opus/flow-medium", "sonnet/flow-low"}},
		{"panel-fix drops high", pjPolicy(`"panel-fix":` + row), []string{"panel-fix"}, []string{"opus/flow-low", "opus/flow-medium", "sonnet/flow-low"}},
		{"self-review-fix drops high", pjPolicy(`"self-review-fix":` + row), []string{"self-review-fix"}, []string{"opus/flow-low", "opus/flow-medium", "sonnet/flow-low"}},
		{"pipeline-fix-fix drops high", pjPolicy(`"pipeline-fix-fix":` + row), []string{"pipeline-fix-fix"}, []string{"opus/flow-low", "opus/flow-medium", "sonnet/flow-low"}},
		{"planner keeps high", pjPolicy(`"planner":{"base":["opus/flow-planner-medium","opus/flow-planner-high","opus/flow-planner-xhigh","opus/flow-planner-max","sonnet/flow-planner-low"],"optin":[],"marked":[]}`), []string{"planner"}, []string{"opus/flow-planner-medium", "opus/flow-planner-high", "sonnet/flow-planner-low"}},
		{"-opus-only leaves the opus pairs", pjPolicy(`"implementer":` + row), []string{"implementer", "-opus-only"}, []string{"opus/flow-low", "opus/flow-medium"}},
		{"-opus-only yields on a row with no opus pair", pjPolicy(`"panel-reviewer":{"base":["sonnet/flow-low","fable/flow-medium"],"optin":[],"marked":[]}`), []string{"panel-reviewer", "-opus-only"}, []string{"sonnet/flow-low", "fable/flow-medium"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			stub := newPJStub(t, "printf 'PAIR: %s\\nREASON: r\\n' '"+c.want[0]+"'\n")
			args := append([]string{c.args[0], pjDigest(t, "a plan")}, c.args[1:]...)
			if r := runGuard("pair-judge", args, pjEnv(t, c.policy, stub)); r.rc != 0 {
				t.Fatalf("rc=%d out=%q, want 0", r.rc, r.out)
			}
			if got := stub.candidates(t); !slices.Equal(got, c.want) {
				t.Errorf("candidates = %q, want %q", got, c.want)
			}
		})
	}

	t.Run("an open experiment's override replaces the base row", func(t *testing.T) {
		t.Parallel()
		stub := newPJStub(t, pjAnswer)
		policy := pjPolicy(`"implementer":` + row)
		env := fakeFlowEnvBare(t, "[ \"$1 $2\" = \"policy get\" ] && { printf '%s\\n' '"+policy+"'; exit 0; }\n"+
			"[ \"$1 $2\" = \"experiment overrides\" ] || exit 9\nprintf '%s\\n' '{\"implementer\":[\"sonnet/flow-low\",\"sonnet/flow-medium\"]}'\n",
			"FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.cmd())
		if r := runGuard("pair-judge", []string{"implementer", pjDigest(t, "a plan")}, env); r.rc != 0 {
			t.Fatalf("rc=%d out=%q, want 0", r.rc, r.out)
		}
		if got, want := stub.candidates(t), []string{"sonnet/flow-low", "sonnet/flow-medium"}; !slices.Equal(got, want) {
			t.Errorf("candidates = %q, want %q", got, want)
		}
	})

	for pair, want := range map[string]string{
		"opus/flow-low":            "opus low judge r\n",
		"opus/flow-planner-medium": "opus medium judge r\n",
		"sonnet/Explore@low":       "sonnet low judge r\n",
	} {
		t.Run("prints "+pair+" parsed", func(t *testing.T) {
			t.Parallel()
			stub := newPJStub(t, "printf 'PAIR: %s\\nREASON: r\\n' '"+pair+"'\n")
			policy := pjPolicy(`"explorer":{"base":["` + pair + `","opus/Explore@high"],"optin":[],"marked":[]}`)
			if r := runGuard("pair-judge", []string{"explorer", pjDigest(t, "a plan")}, pjEnv(t, policy, stub)); r.rc != 0 || r.stdout != want {
				t.Errorf("rc=%d stdout=%q, want 0 and %q", r.rc, r.stdout, want)
			}
		})
	}
}

// TestPairJudgeSingle: a role with one candidate prints it as `single` and
// never runs the judge.
func TestPairJudgeSingle(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, pjAnswer)
	policy := pjPolicy(`"panel-fix":{"base":["opus/flow-low","opus/flow-high"],"optin":[],"marked":[]}`)
	r := runGuard("pair-judge", []string{"panel-fix", pjDigest(t, "a plan")}, pjEnv(t, policy, stub))
	if r.rc != 0 || r.stdout != "opus low single the only candidate\n" {
		t.Fatalf("rc=%d stdout=%q, want 0 and `opus low single the only candidate`", r.rc, r.stdout)
	}
	if stub.ran() {
		t.Error("the judge ran for a single candidate")
	}
}

// TestPairJudgeTakesThePick: the judge's answer is printed as `judge`; it
// runs on the pair-judge row's pair with the measured flags, its system
// prompt carries the role, its Does line, every candidate and the answer
// format, and its stdin is the digest.
// smoke: core — the judge's pick reaches the printed line
func TestPairJudgeTakesThePick(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, pjAnswer)
	policy := pjPolicy(`"implementer":{"base":["opus/flow-low","opus/flow-medium","sonnet/flow-low"],"optin":[],"marked":[]}`)
	const digest = "Task 1: rename foo to bar across 40 Markdown files.\n"
	r := runGuard("pair-judge", []string{"implementer", pjDigest(t, digest)}, pjEnv(t, policy, stub))
	if r.rc != 0 || r.stdout != "sonnet low judge renames only\n" {
		t.Fatalf("rc=%d stdout=%q, want 0 and `sonnet low judge renames only`", r.rc, r.stdout)
	}
	argv := strings.Join(stub.argv(t), "\x00")
	for _, want := range [][]string{
		{"-p"}, {"--model", "sonnet"}, {"--effort", "low"}, {"--tools", ""}, {"--setting-sources", ""},
		{"--strict-mcp-config"}, {"--no-session-persistence"},
	} {
		if !strings.Contains("\x00"+argv+"\x00", "\x00"+strings.Join(want, "\x00")+"\x00") {
			t.Errorf("argv %q lacks %q", stub.argv(t), want)
		}
	}
	prompt := stub.systemPrompt(t)
	for _, want := range []string{
		"implementer", "Writes the code for a group of tasks, tests first.",
		"- opus/flow-low\n", "- opus/flow-medium\n", "- sonnet/flow-low",
		"PAIR: <one candidate, verbatim>", "REASON: <one line>",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("system prompt lacks %q:\n%s", want, prompt)
		}
	}
	if got := stub.stdin(t); got != digest {
		t.Errorf("stdin = %q, want the digest %q", got, digest)
	}
}

// TestPairJudgeRunsInEnvDir: the judge runs in the guard's working directory.
func TestPairJudgeRunsInEnvDir(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, "pwd -P > '"+"DIR/pwd'\n"+pjAnswer)
	writeExec(t, stub.cmd(), strings.ReplaceAll(mustRead(t, stub.cmd()), "DIR", stub.dir))
	env := pjEnv(t, pjPolicy(`"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`), stub)
	env.Dir = t.TempDir()
	if r := runGuard("pair-judge", []string{"implementer", pjDigest(t, "a plan")}, env); r.rc != 0 {
		t.Fatalf("rc=%d stdout=%q", r.rc, r.stdout)
	}
	want, _ := filepath.EvalSymlinks(env.Dir)
	if got := strings.TrimSpace(mustRead(t, stub.dir+"/pwd")); got != want {
		t.Errorf("judge ran in %q, want %q", got, want)
	}
}

// TestPairJudgeOrphanHoldingPipes: a judge that exits leaving a child holding
// its stdout returns promptly instead of waiting on the child.
func TestPairJudgeOrphanHoldingPipes(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, "/bin/sleep 15 &\n"+pjAnswer)
	env := pjEnv(t, pjPolicy(`"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`), stub)
	start := time.Now()
	if r := runGuard("pair-judge", []string{"implementer", pjDigest(t, "a plan")}, env); r.rc != 0 {
		t.Fatalf("rc=%d stdout=%q", r.rc, r.stdout)
	}
	if d := time.Since(start); d > 8*time.Second {
		t.Errorf("an orphaned child held the guard for %v", d)
	}
}

func mustRead(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestPairJudgeFallsBack: every judge failure prints the first candidate as
// `fallback` with its cause and exits 0.
func TestPairJudgeFallsBack(t *testing.T) {
	t.Parallel()
	const row = `"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`
	cases := []struct {
		name    string
		answer  string
		policy  string
		missing bool
		timeout time.Duration
		cause   string
	}{
		{name: "missing command", missing: true, policy: pjPolicy(row), cause: "not runnable"},
		{name: "non-zero exit", answer: "exit 1\n", policy: pjPolicy(row), cause: "exit 1"},
		{name: "timeout", answer: "exec sleep 5\n", policy: pjPolicy(row), timeout: 200 * time.Millisecond, cause: "timed out"},
		{name: "no PAIR line", answer: "printf 'REASON: r\\n'\n", policy: pjPolicy(row), cause: "no PAIR"},
		{name: "pick outside the candidates", answer: "printf 'PAIR: opus/flow-high\\nREASON: r\\n'\n", policy: pjPolicy(row), cause: "outside the candidates"},
		{name: "no pair-judge row", answer: pjAnswer, policy: "{" + row + "}", cause: "no pair-judge pair"},
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
			start := time.Now()
			r := runGuard("pair-judge", []string{"implementer", pjDigest(t, "a plan")}, env)
			if r.rc != 0 || !strings.HasPrefix(r.stdout, "opus low fallback ") || !strings.Contains(r.stdout, c.cause) || strings.Count(r.stdout, "\n") != 1 {
				t.Fatalf("rc=%d stdout=%q, want 0 and one `opus low fallback …%s…` line", r.rc, r.stdout, c.cause)
			}
			if c.timeout > 0 && time.Since(start) > 4*time.Second {
				t.Errorf("a %v deadline took %v", c.timeout, time.Since(start))
			}
			if c.name == "no pair-judge row" && stub.ran() {
				t.Error("the judge ran with no pair-judge pair to run on")
			}
		})
	}
}

// TestPairJudgeCapsTheDigest: a digest over 200 KB reaches the judge as its
// first 204800 bytes and a truncation line; a small one reaches it whole.
func TestPairJudgeCapsTheDigest(t *testing.T) {
	t.Parallel()
	policy := pjPolicy(`"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`)
	big := strings.Repeat("0123456789abcdef", 300*1024/16)
	small := strings.Repeat("x", 1024)
	for digest, want := range map[string]string{
		big:   big[:204800] + "\n[digest truncated at 200 KB]\n",
		small: small,
	} {
		stub := newPJStub(t, pjAnswer)
		if r := runGuard("pair-judge", []string{"implementer", pjDigest(t, digest)}, pjEnv(t, policy, stub)); r.rc != 0 {
			t.Fatalf("rc=%d out=%q, want 0", r.rc, r.out)
		}
		if got := stub.stdin(t); got != want {
			t.Errorf("a %d-byte digest reached the judge as %d bytes ending %q, want %d ending %q",
				len(digest), len(got), got[max(0, len(got)-40):], len(want), want[max(0, len(want)-40):])
		}
	}
}

// TestPairJudgeRefuses: usage errors, an unjudged role, an unreadable digest
// or policy, and a role left with no candidate exit 2 with nothing on stdout.
func TestPairJudgeRefuses(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, pjAnswer)
	env := pjEnv(t, testPolicy, stub)
	digest := pjDigest(t, "a plan")
	cases := map[string]struct {
		args []string
		env  Env
	}{
		"no arguments":         {nil, env},
		"one argument":         {[]string{"implementer"}, env},
		"a bad third argument": {[]string{"implementer", digest, "-sonnet"}, env},
		"four arguments":       {[]string{"implementer", digest, "-opus-only", "x"}, env},
		"unreadable digest":    {[]string{"implementer", digest + ".absent"}, env},
		"no policy":            {[]string{"implementer", digest}, fakeFlowEnvBare(t, "exit 1\n", "FLOW_PAIR_JUDGE_CMD="+stub.cmd())},
		"no candidates": {[]string{"implementer", digest}, pjEnv(t,
			pjPolicy(`"implementer":{"base":["opus/flow-high"],"optin":[],"marked":[]}`), stub)},
	}
	for _, role := range []string{"visual-verifier", "operator-asked", "tooling-analyst", "fix-run-judge", "panel-merger"} {
		cases["role "+role] = struct {
			args []string
			env  Env
		}{[]string{role, digest}, env}
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if r := runGuard("pair-judge", c.args, c.env); r.rc != 2 || r.stdout != "" {
				t.Errorf("rc=%d out=%q, want 2 and no stdout", r.rc, r.out)
			}
		})
	}
	if stub.ran() {
		t.Error("the judge ran on a refused call")
	}
}
