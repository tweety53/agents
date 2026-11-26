package guard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

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

var pjAnswer = pjJSON("PAIR: sonnet/flow-low\nREASON: renames only", "")

// TestPairJudgeCandidates: the candidates are the role's base pairs, the
// overlay applied, every effort the row admits kept; marked and opt-in pairs
// are never listed. Each pair shape prints parsed into model and effort.
func TestPairJudgeCandidates(t *testing.T) {
	t.Parallel()
	const row = `{"base":["opus/flow-low","opus/flow-medium","opus/flow-high","sonnet/flow-low","opus/flow-xhigh"],"optin":["sonnet/flow-medium"],"marked":[{"pair":"opus/flow-max","marker":"No-progress escalation"}]}`
	cases := []struct {
		name   string
		policy string
		args   []string
		want   []string
	}{
		{"implementer keeps every base pair", pjPolicy(`"implementer":` + row), []string{"implementer"}, []string{"opus/flow-low", "opus/flow-medium", "opus/flow-high", "sonnet/flow-low", "opus/flow-xhigh"}},
		{"planner keeps every base pair", pjPolicy(`"planner":{"base":["opus/flow-planner-medium","opus/flow-planner-max","sonnet/flow-planner-low"],"optin":[],"marked":[]}`), []string{"planner"}, []string{"opus/flow-planner-medium", "opus/flow-planner-max", "sonnet/flow-planner-low"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			stub := newPJStub(t, pjJSON("PAIR: "+c.want[0]+"\nREASON: r", ""))
			args := append([]string{c.args[0], pjDigest(t, "a plan")}, c.args[1:]...)
			if r := runGuard("pair-judge", args, pjEnv(t, c.policy, stub)); r.rc != 0 {
				t.Fatalf("rc=%d out=%q, want 0", r.rc, r.out)
			}
			if got := stub.candidates(t); !slices.Equal(got, c.want) {
				t.Errorf("candidates = %q, want %q", got, c.want)
			}
		})
	}

	for pair, want := range map[string]string{
		"opus/flow-low":            "opus low judge r (judge ",
		"opus/flow-planner-medium": "opus medium judge r (judge ",
		"sonnet/Explore@low":       "sonnet low judge r (judge ",
	} {
		t.Run("prints "+pair+" parsed", func(t *testing.T) {
			t.Parallel()
			stub := newPJStub(t, pjJSON("PAIR: "+pair+"\nREASON: r", ""))
			policy := pjPolicy(`"explorer":{"base":["` + pair + `","opus/Explore@high"],"optin":[],"marked":[]}`)
			if r := runGuard("pair-judge", []string{"explorer", pjDigest(t, "a plan")}, pjEnv(t, policy, stub)); r.rc != 0 || !strings.HasPrefix(r.stdout, want) {
				t.Errorf("rc=%d stdout=%q, want 0 opening %q", r.rc, r.stdout, want)
			}
		})
	}
}

// TestPairJudgeJudgesTheDesigner: a designer dispatch is judged among its role's base pairs, never refused (a small
// adjustment round got the role's first pair every time while the role stood outside the judged set).
func TestPairJudgeJudgesTheDesigner(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, pjJSON("PAIR: opus/flow-low\nREASON: a small adjustment", ""))
	policy := pjPolicy(`"designer":{"base":["opus/flow-medium","opus/flow-low"],"optin":[],"marked":[]}`)
	r := runGuard("pair-judge", []string{"designer", pjDigest(t, "move one icon")}, pjEnv(t, policy, stub))
	if r.rc != 0 || !strings.HasPrefix(r.stdout, "opus low judge ") || !stub.ran() {
		t.Errorf("rc=%d stdout=%q ran=%v, want 0, `opus low judge …` and the judge run", r.rc, r.stdout, stub.ran())
	}
}

// TestPairJudgeSingle: a role with one candidate prints it as `single` and
// never runs the judge.
func TestPairJudgeSingle(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, pjAnswer)
	policy := pjPolicy(`"panel-fix":{"base":["opus/flow-low"],"optin":["opus/flow-high"],"marked":[]}`)
	r := runGuard("pair-judge", []string{"panel-fix", pjDigest(t, "a plan")}, pjEnv(t, policy, stub))
	if r.rc != 0 || r.stdout != "opus low single the only candidate\n" {
		t.Fatalf("rc=%d stdout=%q, want 0 and `opus low single the only candidate`", r.rc, r.stdout)
	}
	if stub.ran() {
		t.Error("the judge ran for a single candidate")
	}
}

// TestPairJudgeSavesThePickBesideTheDigest: the printed line is saved, byte
// for byte, as <key>.pick beside the <key>.md digest, and a later run under
// the same key overwrites it (decision pair-pick-file).
func TestPairJudgeSavesThePickBesideTheDigest(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "pair-judge")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	digest := filepath.Join(dir, "task-1-implementer.md")
	if err := os.WriteFile(digest, []byte("a plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	pick := filepath.Join(dir, "task-1-implementer.pick")
	stub := newPJStub(t, pjAnswer)
	for _, pair := range []string{"opus/flow-low", "sonnet/flow-medium"} {
		policy := pjPolicy(`"panel-fix":{"base":["` + pair + `"],"optin":[],"marked":[]}`)
		r := runGuard("pair-judge", []string{"panel-fix", digest}, pjEnv(t, policy, stub))
		if r.rc != 0 {
			t.Fatalf("%s: rc=%d out=%q, want 0", pair, r.rc, r.out)
		}
		if got := mustRead(t, pick); got != r.stdout {
			t.Errorf("%s: %s = %q, want the printed line %q", pair, pick, got, r.stdout)
		}
	}
}

// TestPairJudgeRefusalSavesNoPick: a refused call (exit 2) writes no .pick and
// leaves no earlier run's .pick under the key.
func TestPairJudgeRefusalSavesNoPick(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	digest := filepath.Join(dir, "x.md")
	if err := os.WriteFile(digest, []byte("a plan"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "x.pick"), []byte("opus low judge stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stub := newPJStub(t, pjAnswer)
	if r := runGuard("pair-judge", []string{"pair-judge", digest}, pjEnv(t, testPolicy, stub)); r.rc != 2 {
		t.Fatalf("rc=%d out=%q, want 2", r.rc, r.out)
	}
	if _, err := os.Stat(filepath.Join(dir, "x.pick")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("x.pick: stat err = %v, want not exist", err)
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
	if r.rc != 0 || !strings.HasPrefix(r.stdout, "sonnet low judge renames only (judge ") {
		t.Fatalf("rc=%d stdout=%q, want 0 opening `sonnet low judge renames only (judge `", r.rc, r.stdout)
	}
	argv := strings.Join(stub.argv(t), "\x00")
	for _, want := range [][]string{
		{"-p"}, {"--model", "sonnet"}, {"--effort", "low"}, {"--tools", ""}, {"--setting-sources", ""},
		{"--strict-mcp-config"}, {"--no-session-persistence"}, {"--output-format", "json"},
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
	stub := newPJStub(t, "/bin/sleep 15 & echo $! > 'DIR/orphan.pid'\n"+pjAnswer)
	writeExec(t, stub.cmd(), strings.ReplaceAll(mustRead(t, stub.cmd()), "DIR", stub.dir))
	env := pjEnv(t, pjPolicy(`"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`), stub)
	if r := runGuard("pair-judge", []string{"implementer", pjDigest(t, "a plan")}, env); r.rc != 0 {
		t.Fatalf("rc=%d stdout=%q", r.rc, r.stdout)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(mustRead(t, stub.dir+"/orphan.pid")))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	if err := syscall.Kill(pid, 0); err != nil {
		t.Errorf("the guard waited for the orphan %d to exit: %v", pid, err)
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
		{name: "timeout", answer: "exec sleep 5\n", policy: pjPolicy(row), timeout: 200 * time.Millisecond, cause: "timed out after 200ms (judge "},
		{name: "no PAIR line", answer: pjJSON("REASON: r", ""), policy: pjPolicy(row), cause: "no PAIR"},
		{name: "pick outside the candidates", answer: pjJSON("PAIR: opus/flow-high\nREASON: r", ""), policy: pjPolicy(row), cause: "outside the candidates"},
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
		"no arguments":      {nil, env},
		"one argument":      {[]string{"implementer"}, env},
		"a third argument":  {[]string{"implementer", digest, "-opus-only"}, env},
		"unreadable digest": {[]string{"implementer", digest + ".absent"}, env},
		"no policy":         {[]string{"implementer", digest}, fakeFlowEnvBare(t, "exit 1\n", "FLOW_PAIR_JUDGE_CMD="+stub.cmd())},
		"no candidates": {[]string{"implementer", digest}, pjEnv(t,
			pjPolicy(`"implementer":{"base":[],"optin":["opus/flow-high"],"marked":[]}`), stub)},
	}
	for _, role := range []string{"pair-judge", "no-such-role"} {
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

// TestPairJudgeRecordsItsSpend pins kan-1119's self-review F12: a judged
// line's reason ends with the call's time and, when the answer carries it,
// its cost; an answer that is not JSON falls back naming that.
func TestPairJudgeRecordsItsSpend(t *testing.T) {
	t.Parallel()
	policy := pjPolicy(`"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`)
	cases := []struct {
		name, answer, prefix, suffix string
		has, lacks                   []string
	}{
		{"cost", pjJSON("PAIR: sonnet/flow-low\nREASON: renames only", `,"total_cost_usd":0.0123`),
			"sonnet low judge renames only (judge ", "s, $0.0123)\n", nil, nil},
		{"no cost", pjJSON("PAIR: sonnet/flow-low\nREASON: renames only", ""),
			"sonnet low judge renames only (judge ", "s)\n", nil, []string{"$"}},
		{"not JSON", "printf 'PAIR: sonnet/flow-low\\nREASON: r\\n'\n",
			"opus low fallback ", "\n", []string{"not JSON", "(judge "}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			stub := newPJStub(t, c.answer)
			r := runGuard("pair-judge", []string{"implementer", pjDigest(t, "a plan")}, pjEnv(t, policy, stub))
			if r.rc != 0 || !strings.HasPrefix(r.stdout, c.prefix) || !strings.HasSuffix(r.stdout, c.suffix) {
				t.Fatalf("rc=%d stdout=%q, want 0, opening %q and ending %q", r.rc, r.stdout, c.prefix, c.suffix)
			}
			for _, s := range c.has {
				if !strings.Contains(r.stdout, s) {
					t.Errorf("stdout %q lacks %q", r.stdout, s)
				}
			}
			for _, s := range c.lacks {
				if strings.Contains(r.stdout, s) {
					t.Errorf("stdout %q carries %q", r.stdout, s)
				}
			}
		})
	}
}

// TestPairJudgeGroupDigestDropsThePlanInputs: a `decide-group-<g>` digest
// reaches the judge without the plan's `inputs:` line, which counts the whole
// plan and not the group; every other digest keeps it.
func TestPairJudgeGroupDigestDropsThePlanInputs(t *testing.T) {
	t.Parallel()
	policy := pjPolicy(`"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`)
	const tasks, inputs = "Task 15: one UI test.\n", "inputs: tasks=47 files=90 repos=2 migration=yes spec=yes red=no unverified=yes\n"
	for name, want := range map[string]string{
		"decide-group-2.md":     tasks + "after\n",
		"decide-implementer.md": tasks + inputs + "after\n",
	} {
		stub := newPJStub(t, pjAnswer)
		p := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(p, []byte(tasks+inputs+"after\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if r := runGuard("pair-judge", []string{"implementer", p}, pjEnv(t, policy, stub)); r.rc != 0 {
			t.Fatalf("%s: rc=%d err=%q", name, r.rc, r.err)
		}
		if got := stub.stdin(t); got != want {
			t.Errorf("%s: stdin = %q, want %q", name, got, want)
		}
	}
}

// TestPairJudgeRunsWithoutAutoMemory: the judge command runs with
// CLAUDE_CODE_DISABLE_AUTO_MEMORY=1, so the cwd project's memory index never
// reaches it.
func TestPairJudgeRunsWithoutAutoMemory(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, `printf '%s' "$CLAUDE_CODE_DISABLE_AUTO_MEMORY" > "$(dirname "$0")/memory"`+"\n"+pjAnswer)
	policy := pjPolicy(`"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`)
	if r := runGuard("pair-judge", []string{"implementer", pjDigest(t, "Task 1.\n")}, pjEnv(t, policy, stub)); r.rc != 0 {
		t.Fatalf("rc=%d err=%q", r.rc, r.err)
	}
	if got := mustRead(t, stub.dir+"/memory"); got != "1" {
		t.Errorf("CLAUDE_CODE_DISABLE_AUTO_MEMORY = %q in the judge's environment, want 1", got)
	}
}

// TestPairJudgeRunsWithoutTools: the pair judge runs the runner with no tools.
func TestPairJudgeRunsWithoutTools(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, pjJSON("x", ""))
	rows := map[string][]policyRow{"pair-judge": {{Pairs: []string{"sonnet/flow-low"}}}}
	pjAsk(pjEnv(t, "{}", stub), rows, "implementer", []pjCandidate{{pair: "opus/flow-low"}}, []byte("d"))
	a := stub.argv(t)
	if i := slices.Index(a, "--tools"); i < 0 || a[i+1] != "" {
		t.Errorf("argv %q", a)
	}
}

// TestPairJudgeRecordsItsHeadlessCall: the judge's own model call -- role,
// pair, cost and duration -- lands on the run's collector, which flow-guard
// posts with the guard run, so Role costs counts it.
func TestPairJudgeRecordsItsHeadlessCall(t *testing.T) {
	t.Parallel()
	stub := newPJStub(t, pjJSON("PAIR: sonnet/flow-low\nREASON: renames only", `,"total_cost_usd":0.0123`))
	env := pjEnv(t, pjPolicy(`"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`), stub)
	env.Reasons = &Reasons{}
	if r := runGuard("pair-judge", []string{"implementer", pjDigest(t, "a plan")}, env); r.rc != 0 {
		t.Fatalf("rc=%d stdout=%q", r.rc, r.stdout)
	}
	h := env.Reasons.Headless()
	if h == nil || h.Role != "pair-judge" || h.Model != "sonnet" || h.Effort != "low" || h.CostUSD == nil {
		t.Fatalf("headless = %+v, want the pair-judge call on sonnet low with its cost", h)
	}
}

// TestJudgeGetsTheAPIKeyFromItsFile: FLOW_HEADLESS_API_KEY_FILE reaches the
// judge child as ANTHROPIC_API_KEY; an empty file falls back with the cause
// and the child never runs.
func TestJudgeGetsTheAPIKeyFromItsFile(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ name, content string }{{"key", "sk-test\n"}, {"empty", ""}} {
		stub := newPJStub(t, "printf '%s' \"$ANTHROPIC_API_KEY\" > '"+t.TempDir()+"/unused'\nprintf '%s' \"$ANTHROPIC_API_KEY\" > $(dirname $0)/key\n"+pjJSON("PAIR: opus/flow-low\nREASON: r", ""))
		keyFile := filepath.Join(t.TempDir(), "key")
		writeFile(t, keyFile, c.content)
		env := policyFlowEnv(t, pjPolicy(`"planner":{"base":["opus/flow-low","opus/flow-medium"],"optin":[],"marked":[]},"pair-judge":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}`),
			"FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.cmd(), "FLOW_HEADLESS_API_KEY_FILE="+keyFile)
		env.Environ = pjChildEnviron("")
		r := runGuard("pair-judge", []string{"planner", pjDigest(t, "d")}, env)
		got, _ := os.ReadFile(stub.dir + "/key")
		switch c.name {
		case "key":
			if r.rc != 0 || string(got) != "sk-test" {
				t.Errorf("rc=%d key=%q err=%q", r.rc, got, r.err)
			}
		default:
			if stub.ran() || !strings.Contains(r.stdout, "unreadable or empty") {
				t.Errorf("ran=%v stdout=%q", stub.ran(), r.stdout)
			}
		}
	}
}

// TestJudgeKeyFileOffAndDefault: "off" keeps the child on the subscription,
// and the default ~/.config/flow/headless-api-key is used when present.
func TestJudgeKeyFileOffAndDefault(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"off", "default"} {
		stub := newPJStub(t, "printf '%s' \"$ANTHROPIC_API_KEY\" > $(dirname $0)/key\n"+pjJSON("PAIR: opus/flow-low\nREASON: r", ""))
		home := t.TempDir()
		writeFile(t, home+"/.config/flow/headless-api-key", "sk-home\n")
		extra := "HOME=" + home
		want := "sk-home"
		if mode == "off" {
			extra, want = "FLOW_HEADLESS_API_KEY_FILE=off", ""
		}
		env := policyFlowEnv(t, pjPolicy(`"planner":{"base":["opus/flow-low","opus/flow-medium"],"optin":[],"marked":[]},"pair-judge":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}`),
			"FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.cmd(), extra, "HOME="+home)
		env.Environ = pjChildEnviron("")
		runGuard("pair-judge", []string{"planner", pjDigest(t, "d")}, env)
		if got, _ := os.ReadFile(stub.dir + "/key"); string(got) != want {
			t.Errorf("%s: key %q, want %q", mode, got, want)
		}
	}
}

// TestImplementerOpusLowSplitsWithSonnetLow: the implementer's opus low pick
// becomes sonnet low when the coin says so; any other pick, and any other
// role's opus low, stays as picked.
func TestImplementerOpusLowSplitsWithSonnetLow(t *testing.T) {
	t.Parallel()
	pol := func(role string) string {
		return pjPolicy(`"` + role + `":{"base":["opus/flow-low","sonnet/flow-low","opus/flow-high"],"optin":[],"marked":[]},"pair-judge":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}`)
	}
	for _, c := range []struct {
		name, role, pick, want string
		coin                   bool
	}{
		{"heads moves", "implementer", "opus/flow-low", "sonnet low split", true},
		{"tails stays", "implementer", "opus/flow-low", "opus low judge", false},
		{"other pair stays", "implementer", "opus/flow-high", "opus high judge", true},
		{"other role stays", "panel-fix", "opus/flow-low", "opus low judge", true},
	} {
		stub := newPJStub(t, pjJSON("PAIR: "+c.pick+"\nREASON: r", ""))
		env := policyFlowEnv(t, pol(c.role), "FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.cmd())
		env.PairJudgeCoin = func() bool { return c.coin }
		r := runGuard("pair-judge", []string{c.role, pjDigest(t, "d")}, env)
		if !strings.HasPrefix(r.stdout, c.want) {
			t.Errorf("%s: stdout %q, want prefix %q", c.name, r.stdout, c.want)
		}
	}
}

// pjNamedRepo is a git repository holding one file per pair-judge refusal
// beside the two files it sends, and a file outside it; the digest names
// every one and sits in the repository's .superpowers/.
func pjNamedRepo(t *testing.T) (repo, digest string, files map[string]string) {
	t.Helper()
	repo, _ = filepath.EvalSymlinks(t.TempDir())
	outside, _ := filepath.EvalSymlinks(t.TempDir())
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	files = map[string]string{
		repo + "/plan.md":                "the plan\n",
		repo + "/.superpowers/ticket.md": "the ticket\n",
		repo + "/.env.local":             "TOKEN=x\n",
		repo + "/secrets.txt":            "s\n",
		repo + "/ignored.log":            "ignored\n",
		repo + "/bin.dat":                "a\x00b",
		repo + "/.claude/notes.md":       "home claude\n",
		outside + "/outside.md":          "outside\n",
		repo + "/.gitignore":             "ignored.log\n.superpowers/\n",
	}
	for p, body := range files {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(outside+"/outside.md", repo+"/escape.md"); err != nil {
		t.Fatal(err)
	}
	digest = repo + "/.superpowers/brief.md"
	return repo, digest, files
}

// TestPairJudgeAppendsNamedFiles: the judge's stdin is the brief, then each
// part it names once -- bare or in backticks, before a full stop alike -- and nothing outside a git tree, behind a symlink escape, .env,
// secret-named, gitignored outside .superpowers/, binary, under ~/.claude,
// missing, or the brief itself.
func TestPairJudgeAppendsNamedFiles(t *testing.T) {
	t.Parallel()
	repo, digest, files := pjNamedRepo(t)
	var brief strings.Builder
	brief.WriteString("Ask: see `" + repo + "/plan.md:1`, " + repo + "/plan.md:1 and " + repo + "/.superpowers/ticket.md:1.\n")
	for p := range files {
		if strings.HasSuffix(p, ".gitignore") {
			continue // a tracked-tree text file, sent like plan.md
		}
		brief.WriteString("also " + p + ":1\n")
	}
	brief.WriteString("and " + repo + "/escape.md:1, " + repo + "/missing.md:1, " + digest + ":1\n")
	if err := os.WriteFile(digest, []byte(brief.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	policy := pjPolicy(`"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`)
	stub := newPJStub(t, pjAnswer)
	env := policyFlowEnv(t, policy, "FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.cmd(), "HOME="+repo)
	if r := runGuard("pair-judge", []string{"implementer", digest}, env); r.rc != 0 {
		t.Fatalf("rc=%d out=%q, want 0", r.rc, r.out)
	}
	want := brief.String() +
		"\n\n--- " + repo + "/plan.md:1 ---\nthe plan\n" +
		"\n\n--- " + repo + "/.superpowers/ticket.md:1 ---\nthe ticket\n"
	if got := stub.stdin(t); got != want {
		t.Errorf("stdin =\n%s\nwant\n%s", got, want)
	}
}

// TestPairJudgeSecretNames: only .env files, private keys, key stores and
// files named secret, secrets or credentials are kept back.
func TestPairJudgeSecretNames(t *testing.T) {
	t.Parallel()
	repo, _, _ := pjNamedRepo(t)
	for name, sent := range map[string]bool{
		"monkey.go": true, "token.go": true, "tokenizer.ts": true, "secretary.md": true,
		"password.go": true, "deploy-secret.txt": true, "api_key.md": true, "credentials_test.go": true,
		"id_rsa.pub": true, "server.crt": true,
		".env": false, ".env.local": false,
		"server.key": false, "server.PEM": false, "a.p12": false, "a.pfx": false, "a.jks": false, "a.keystore": false,
		"id_rsa": false, "id_dsa": false, "id_ecdsa": false, "id_ed25519": false,
		"secret": false, "secret.yaml": false, "Secrets.json": false, "credentials": false, "credentials.toml": false,
	} {
		p := repo + "/" + name
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, ok := pjSendable(Env{Getenv: func(string) string { return "" }}, p); ok != sent {
			t.Errorf("%s: sent %v, want %v", name, ok, sent)
		}
		if err := os.Remove(p); err != nil {
			t.Fatal(err)
		}
	}
}

// TestPairJudgeCapsEachNamedFile: a named part over 8192 bytes reaches the
// judge as its head -- cut after the last whole line, or else the last whole
// rune -- and a truncation line; one of exactly 8192 bytes is whole, and
// later files still follow.
func TestPairJudgeCapsEachNamedFile(t *testing.T) {
	t.Parallel()
	repo, digest, _ := pjNamedRepo(t)
	lines := strings.Repeat("123456789\n", 1000)
	runes := "x" + strings.Repeat("é", 5000) // byte 8192 falls inside a rune
	exact := strings.Repeat("e", 8192)
	var brief, tail strings.Builder
	for _, f := range [][3]string{
		{"lines.md", lines, lines[:8190] + "[file truncated at 8 KB]\n"},
		{"runes.md", runes, runes[:8191] + "\n[file truncated at 8 KB]\n"},
		{"exact.md", exact, exact},
		{"after.md", "after\n", "after\n"},
	} {
		if err := os.WriteFile(repo+"/"+f[0], []byte(f[1]), 0o644); err != nil {
			t.Fatal(err)
		}
		brief.WriteString("read " + repo + "/" + f[0] + ":1-2000\n")
		tail.WriteString("\n\n--- " + repo + "/" + f[0] + ":1-2000 ---\n" + f[2])
	}
	pjWantStdin(t, digest, brief.String(), brief.String()+tail.String())
}

// TestPairJudgeCapsNamedFiles: the file crossing the 200 KB total is cut
// there and followed by the truncation line; the brief reaches the judge whole.
func TestPairJudgeCapsNamedFiles(t *testing.T) {
	t.Parallel()
	repo, digest, _ := pjNamedRepo(t)
	brief := strings.Repeat("b", 200000) + "\n"
	if err := os.WriteFile(repo+"/big.md", []byte(strings.Repeat("0123456789abcdef", 1024)), 0o644); err != nil {
		t.Fatal(err)
	}
	brief += "read " + repo + "/big.md:1\n"
	head := brief + "\n\n--- " + repo + "/big.md:1 ---\n"
	want := head + strings.Repeat("0123456789abcdef", 1024)[:204800-len(head)] + "\n[digest truncated at 200 KB]\n"
	pjWantStdin(t, digest, brief, want)
}

// TestPairJudgeAppendsNamedTasks: a named tasks.md reaches the judge as the
// sections of the tasks the brief names -- by an implementer's `Tasks: 7, 9
// of 20` line or a reviewer's `PASS task-<n>:` lines -- each capped at 8 KB on
// its own; a brief naming none, or none the file holds, gets nothing of it. The file sits gitignored under spectre/changes/, as real plans do.
func TestPairJudgeAppendsNamedTasks(t *testing.T) {
	t.Parallel()
	pad := strings.Repeat("padding line\n", 700) // ~9 KB, pushes task 7 past the head
	big := strings.Repeat("123456789\n", 1000)
	plan := "# plan\n\n## area\n\n" +
		"- [x] 1. One\n" + pad +
		"- [x] 7. Seven\nseven body\n" +
		"- [ ] 8. Eight\neight body\n" +
		"- [ ] 9. Nine\n" + big +
		"\n## next\nafter\n"
	sec9 := "- [ ] 9. Nine\n" + big
	sec9 = sec9[:strings.LastIndexByte(sec9[:8192], '\n')+1] + "[file truncated at 8 KB]\n"
	for name, c := range map[string]struct{ brief, want string }{
		"implementer": {"Tasks: 7, 9 of 20\n", "- [x] 7. Seven\nseven body\n" + sec9},
		"reviewer":    {"PASS task-7:\nPASS task-8:\n", "- [x] 7. Seven\nseven body\n- [ ] 8. Eight\neight body\n"},
		"none named":  {"Tasks: none\n", ""},
		"not in file": {"Tasks: 42 of 50\n", ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			repo, digest, _ := pjNamedRepo(t)
			p := repo + "/spectre/changes/c/tasks.md" // gitignored, as spectre change dirs are
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(plan), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(repo+"/.gitignore", []byte("ignored.log\n.superpowers/\nspectre/changes/\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			brief := c.brief + "Plan: " + p + "\n"
			want := brief
			if c.want != "" {
				want += "\n\n--- " + p + " ---\n" + c.want
			}
			pjWantStdin(t, digest, brief, want)
		})
	}
}

// TestPairJudgeAppendsNamedParts: only the part a brief names is sent -- a
// backticked `path#heading` its section up to the next heading of the same or
// higher level, fenced lines not headings; `:a-b` and `:a` those lines -- and
// nothing for a file named bare or a part the file lacks.
func TestPairJudgeAppendsNamedParts(t *testing.T) {
	t.Parallel()
	doc := "# Top\nintro\n## The `x` part, here\nbody\n```\n# not a heading\n```\n### Sub\nsub\n## Next\nafter\n"
	section := "## The `x` part, here\nbody\n```\n# not a heading\n```\n### Sub\nsub\n"
	for name, c := range map[string]struct{ ref, want string }{
		"section":         {"`%s#The x part, here`", section},
		"last section":    {"`%s#Next`", "## Next\nafter\n"},
		"lines":           {"%s:2-3.", "intro\n## The `x` part, here\n"},
		"one line":        {"`%s:2`", "intro\n"},
		"line and column": {"%s:2:7", "intro\n"},
		"past the end":    {"%s:11-99", "after\n"},
		"bare":            {"`%s`", ""},
		"no such heading": {"`%s#Absent`", ""},
		"no such line":    {"%s:99", ""},
		"backward range":  {"%s:3-2", ""},
		"line zero":       {"%s:0", ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			repo, digest, _ := pjNamedRepo(t)
			p := repo + "/doc.md"
			if err := os.WriteFile(p, []byte(doc), 0o644); err != nil {
				t.Fatal(err)
			}
			ref := fmt.Sprintf(c.ref, p)
			brief := "Read " + ref + "\n"
			want := brief
			if c.want != "" {
				want += "\n\n--- " + strings.TrimSuffix(strings.Trim(ref, "`."), ":7") + " ---\n" + c.want
			}
			pjWantStdin(t, digest, brief, want)
		})
	}
}

// pjWantStdin judges brief, written to digest, and compares the judge's stdin.
func pjWantStdin(t *testing.T, digest, brief, want string) {
	t.Helper()
	if err := os.WriteFile(digest, []byte(brief), 0o644); err != nil {
		t.Fatal(err)
	}
	policy := pjPolicy(`"implementer":{"base":["opus/flow-low","sonnet/flow-low"],"optin":[],"marked":[]}`)
	stub := newPJStub(t, pjAnswer)
	if r := runGuard("pair-judge", []string{"implementer", digest}, pjEnv(t, policy, stub)); r.rc != 0 {
		t.Fatalf("rc=%d out=%q, want 0", r.rc, r.out)
	}
	if got := stub.stdin(t); got != want {
		t.Errorf("stdin is %d bytes ending %q, want %d ending %q", len(got), got[max(0, len(got)-40):], len(want), want[max(0, len(want)-40):])
	}
}

// TestJudgeKeyOnlyForAnthropicBaseURL: the headless key reaches the judge
// child only when the child's ANTHROPIC_BASE_URL is unset or Anthropic's own
// API host; any other vendor runs on the session's own credentials, so the
// Anthropic key is never sent to it. Hosts compare exactly, never by substring.
func TestJudgeKeyOnlyForAnthropicBaseURL(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ base, want string }{
		{"", "sk-dummy"},
		{"https://api.anthropic.com", "sk-dummy"},
		{"https://API.anthropic.com:443/", "sk-dummy"},
		{"https://example.invalid/apps/anthropic", ""},
		{"https://api.anthropic.com.evil.example", ""},
		{"https://evil.example/api.anthropic.com", ""},
		{"not a url", ""},
	} {
		stub := newPJStub(t, "printf '%s' \"$ANTHROPIC_API_KEY\" > $(dirname $0)/key\n"+pjJSON("PAIR: opus/flow-low\nREASON: r", ""))
		keyFile := filepath.Join(t.TempDir(), "key")
		writeFile(t, keyFile, "sk-dummy\n")
		env := policyFlowEnv(t, pjPolicy(`"planner":{"base":["opus/flow-low","opus/flow-medium"],"optin":[],"marked":[]},"pair-judge":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}`),
			"FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.cmd(), "FLOW_HEADLESS_API_KEY_FILE="+keyFile)
		env.Environ = pjChildEnviron(c.base)
		runGuard("pair-judge", []string{"planner", pjDigest(t, "d")}, env)
		if got, _ := os.ReadFile(stub.dir + "/key"); string(got) != c.want {
			t.Errorf("base %q: key %q, want %q", c.base, got, c.want)
		}
	}
}

// pjChildEnviron is the process environment without any ANTHROPIC_* entry,
// plus ANTHROPIC_BASE_URL=base when base is set: a judge child's environment
// independent of the session running the tests.
func pjChildEnviron(base string) func() []string {
	return func() []string {
		var out []string
		for _, kv := range os.Environ() {
			if !strings.HasPrefix(kv, "ANTHROPIC_") {
				out = append(out, kv)
			}
		}
		if base != "" {
			out = append(out, "ANTHROPIC_BASE_URL="+base)
		}
		return out
	}
}
