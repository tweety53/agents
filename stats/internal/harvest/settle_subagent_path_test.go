package harvest_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
)

func TestSubagentPathNeverSettlesParentInlineRow(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "session", "subagents")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC)
	end := start.Add(30 * time.Second)
	if err := os.WriteFile(filepath.Join(sub, "agent-abc.jsonl"), []byte(`{"type":"assistant","timestamp":"2026-01-01T00:05:00Z","sessionId":"s","isSidechain":true,"agentId":"abc","message":{"id":"m9","model":"claude-sonnet-5","usage":{"input_tokens":20,"output_tokens":3}}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	deps := &settleDeps{
		rows:     map[string]*settleRow{"verify-fix-1-a": {id: 9, endedAt: end, agentID: harvest.InlineAgentID}},
		windows:  []harvest.DispatchWindow{{DispatchID: 9, AgentID: harvest.InlineAgentID, StartedAt: start, EndedAt: &end}},
		restamps: map[int64]json.RawMessage{},
		inline:   []harvest.InlineRow{{ID: 9, SessionToken: "mf-s", Key: "verify-fix-1-a", StartedAt: start, EndedAt: end}},
	}
	w := harvest.NewWatcher([]harvest.Source{harvest.NewClaudeSource(dir)}, newFakeHarvestSink(), harvest.NewAttributor(&fakeWindowSource{}), deps, nil)
	if _, err := w.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p, ok := deps.restamps[9]; ok {
		t.Fatalf("subagent-only harvest restamped parent inline row 9: %s", p)
	}
}

// realReviewerDeps answers kan-1091's `task-13+14+16+20-reviewer` row as the
// store holds it: one row for the agent, its guessed window ending before
// the agent's last record.
type realReviewerDeps struct {
	backfillAgentDeps
}

func (d *realReviewerDeps) DispatchWindowsForAgent(context.Context, string) ([]harvest.DispatchWindow, error) {
	start := time.Date(2026, 11, 16, 20, 40, 47, 0, time.UTC)
	end := time.Date(2026, 11, 16, 20, 43, 47, 0, time.UTC)
	return []harvest.DispatchWindow{{DispatchID: 26, AgentID: "aa9d310c9b4699bb3", StartedAt: start, EndedAt: &end}}, nil
}

// TestRealKan1091ReviewerPairIsCredited replays kan-1091's real reviewer
// transcript lines, harvested before the parent's real back-fill command:
// the row is credited the file's records.
func TestRealKan1091ReviewerPairIsCredited(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "session", "subagents")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	copyFile := func(from, to string) {
		t.Helper()
		raw, err := os.ReadFile(from)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(to, raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	copyFile("testdata/kan1091-reviewer/session/subagents/agent-aa9d310c9b4699bb3.jsonl", filepath.Join(sub, "agent-aa9d310c9b4699bb3.jsonl"))
	deps := &realReviewerDeps{}
	w := harvest.NewWatcher([]harvest.Source{harvest.NewClaudeSource(dir)}, newFakeHarvestSink(), harvest.NewAttributor(&fakeWindowSource{}), deps, nil)
	runCycles(t, w, 1)
	copyFile("testdata/kan1091-reviewer/parent.jsonl", filepath.Join(dir, "session.jsonl"))
	runCycles(t, w, 2)
	if len(deps.restamps) != 1 || !strings.HasPrefix(deps.restamps[0], "26 aa9d310c9b4699bb3 ") || !strings.Contains(deps.restamps[0], `"input":4`) {
		t.Fatalf("restamps = %v, want row 26 credited input 4", deps.restamps)
	}
}
