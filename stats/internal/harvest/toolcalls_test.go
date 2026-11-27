package harvest_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// TestAgentToolCallsPairsUseWithResult: an agent's transcript is found under
// root as AgentTranscriptEnd finds it, each tool_use pairs with the
// tool_result naming its id -- two parallel calls included -- the span named
// by its tool and a short form of its input, a call never answered dropped,
// and a missing transcript reported as not found.
func TestAgentToolCallsPairsUseWithResult(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, "proj", "sess", "subagents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	long := "go test ./internal/store/ -run 'TestTraceSpansOfOneChange|TestBottlenecksRanksEachDimension' -count=1\nsecond line"
	lines := `{"type":"assistant","timestamp":"2026-01-01T00:00:10Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":` + strconv.Quote(long) + `}},{"type":"tool_use","id":"t2","name":"Read","input":{"file_path":"/a/b.go"}}]}}
{"type":"user","timestamp":"2026-01-01T00:00:11Z","message":{"content":[{"type":"tool_result","tool_use_id":"t2"}]}}
not json
{"type":"user","timestamp":"2026-01-01T00:00:14Z","message":{"content":[{"type":"tool_result","tool_use_id":"t1","is_error":true}]}}
{"type":"assistant","timestamp":"2026-01-01T00:00:15Z","message":{"content":[{"type":"tool_use","id":"t3","name":"Grep","input":{"pattern":"x"}}]}}
`
	if err := os.WriteFile(filepath.Join(dir, "agent-a1.jsonl"), []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	got, found, err := harvest.AgentToolCalls(root, "a1")
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	want := []harvest.ToolCall{
		{Name: "Bash: go test ./internal/store/ -run 'TestTraceSpansOfOneChange|Te…", Start: t0.Add(10 * time.Second), End: t0.Add(14 * time.Second), IsError: true},
		{Name: "Read: /a/b.go", Start: t0.Add(10 * time.Second), End: t0.Add(11 * time.Second)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("calls:\n got %+v\nwant %+v", got, want)
	}
	if _, found, err := harvest.AgentToolCalls(root, "absent"); found || err != nil {
		t.Errorf("absent: found=%v err=%v", found, err)
	}
}

// TestTranscriptGapsSplitEachGap: every gap between two consecutive
// timestamped lines is classed by the line that ends it -- an assistant line
// is llm time, a user line answering a tool_use is tool time under that
// tool (Bash by its command word), any other user line is human time; a gap
// is clipped to the window and a line stamped before its predecessor adds
// nothing.
func TestTranscriptGapsSplitEachGap(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, "proj", "sess", "subagents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	lines := `{"type":"user","timestamp":"2026-01-01T00:00:00Z","message":{"content":"do it"}}
{"type":"assistant","timestamp":"2026-01-01T00:00:04Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"go test ./x"}},{"type":"tool_use","id":"t2","name":"Read","input":{"file_path":"/a"}}]}}
{"type":"user","timestamp":"2026-01-01T00:00:05Z","message":{"content":[{"type":"tool_result","tool_use_id":"t2"}]}}
{"type":"user","timestamp":"2026-01-01T00:00:15Z","message":{"content":[{"type":"tool_result","tool_use_id":"t1"}]}}
{"type":"assistant","timestamp":"2026-01-01T00:00:17Z","message":{"content":[{"type":"text","text":"done?"}]}}
{"type":"user","timestamp":"2026-01-01T00:01:17Z","message":{"content":"yes"}}
{"type":"assistant","timestamp":"2026-01-01T00:01:10Z","message":{"content":[{"type":"text","text":"backwards"}]}}
`
	if err := os.WriteFile(filepath.Join(dir, "agent-a1.jsonl"), []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	path, found, err := harvest.AgentTranscriptPath(root, "sess", "a1")
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	gaps, err := harvest.ReadTranscriptGaps(path)
	if err != nil {
		t.Fatal(err)
	}
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if last := gaps[len(gaps)-1].To; !last.Equal(t0.Add(77 * time.Second)) {
		t.Errorf("last gap ends %v, want the latest stamp", last)
	}
	got := harvest.SplitGaps(gaps, t0, t0.Add(time.Hour))
	want := harvest.TimeSplit{LLMMs: 6000, ToolMs: 11000, HumanMs: 60000, ByTool: map[string]int64{"Read": 1000, "Bash go": 10000}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("split:\n got %+v\nwant %+v", got, want)
	}
	clipped := harvest.SplitGaps(gaps, t0.Add(2*time.Second), t0.Add(10*time.Second))
	want = harvest.TimeSplit{LLMMs: 2000, ToolMs: 6000, ByTool: map[string]int64{"Read": 1000, "Bash go": 5000}}
	if !reflect.DeepEqual(clipped, want) {
		t.Errorf("clipped:\n got %+v\nwant %+v", clipped, want)
	}
}

// TestTokenTranscriptPath: a cc- token names its agent's transcript -- the
// session's own file for main, the agent's subagents file otherwise -- and
// an mf-/fi- token names none.
func TestTokenTranscriptPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sess := "75ff2114-8338-4a5e-8d6d-462e72ab97c5"
	main := filepath.Join(root, "proj", sess+".jsonl")
	agent := filepath.Join(root, "proj", sess, "subagents", "agent-a18f.jsonl")
	for _, p := range []string{main, agent} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []struct {
		token, path string
		ok, found   bool
	}{
		{"cc-" + sess + "-main", main, true, true},
		{"cc-" + sess + "-a18f", agent, true, true},
		{"cc-" + sess + "-absent", "", true, false},
		{"mf-7c3a91e2", "", false, false},
		{"fi-7c3a91e2", "", false, false},
	} {
		path, ok, found, err := harvest.TokenTranscriptPath(root, c.token)
		if err != nil || path != c.path || ok != c.ok || found != c.found {
			t.Errorf("%s: path=%q ok=%v found=%v err=%v, want %q %v %v", c.token, path, ok, found, err, c.path, c.ok, c.found)
		}
	}
}

// TestTranscriptGapsKeyBashByItsCommand: a Bash call is filed under the
// first word of the command it runs, past a leading cd and variable
// assignments, so `cd <dir> && go test` is Bash go, not Bash cd.
func TestTranscriptGapsKeyBashByItsCommand(t *testing.T) {
	t.Parallel()
	for cmd, want := range map[string]string{
		"go test ./...":                      "Bash go",
		"cd /w/stats && go test ./...":       "Bash go",
		"cd /w; cd stats\nmake -s web-dist":  "Bash make",
		"S=/tmp/x; D=$S/y; git -C $D status": "Bash git",
		"FOO=1 BAR=2 python3 - <<'EOF'":      "Bash python3",
		"cd /w":                              "Bash",
	} {
		in, _ := json.Marshal(map[string]string{"command": cmd})
		raw := `{"type":"assistant","timestamp":"2026-01-01T00:00:00Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":` + string(in) + `}]}}` + "\n" +
			`{"type":"user","timestamp":"2026-01-01T00:00:05Z","message":{"content":[{"type":"tool_result","tool_use_id":"t1"}]}}` + "\n"
		gaps := harvest.TranscriptGaps([]byte(raw))
		if len(gaps) != 1 || gaps[0].Tool != want {
			t.Errorf("%q: gaps = %+v, want one %s gap", cmd, gaps, want)
		}
	}
}
