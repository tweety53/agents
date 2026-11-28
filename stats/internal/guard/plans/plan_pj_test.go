package plans

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
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
	guardtest.WriteExec(t, s.cmd(), "#!/bin/sh\nprintf '%s\\0' \"$@\" > '"+s.dir+"/argv'\n/bin/cat > '"+s.dir+"/stdin'\n"+answer)
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

// pjRepoRoot is the agents checkout these tests run in, whose roles table
// the guard reads the Does line from.
func pjRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// pjEnv is an env whose `flow` answers policy with policy, whose judge
// command is stub, and whose repository root is this checkout.
func pjEnv(t *testing.T, policy string, stub pjStub) core.Env {
	t.Helper()
	return guardtest.PolicyFlowEnv(t, policy, "FLOW_GUARD_REPO_ROOT="+pjRepoRoot(t), "FLOW_PAIR_JUDGE_CMD="+stub.cmd())
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

// pjJSON is a stub answer printing the judge's text as `claude -p
// --output-format json` does, with extra JSON fields (a cost) appended.
func pjJSON(text, extra string) string {
	b, _ := json.Marshal(text)
	return "printf '%s\\n' '{\"type\":\"result\",\"result\":" + string(b) + extra + "}'\n"
}
