package guard

import (
	"errors"
	"os"
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
	if r := runGuard("pair-judge", []string{"visual-verifier", digest}, pjEnv(t, testPolicy, stub)); r.rc != 2 {
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
	for _, role := range []string{"visual-verifier", "operator-asked", "panel-merger"} {
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
		runGuard("pair-judge", []string{"planner", pjDigest(t, "d")}, env)
		if got, _ := os.ReadFile(stub.dir + "/key"); string(got) != want {
			t.Errorf("%s: key %q, want %q", mode, got, want)
		}
	}
}
