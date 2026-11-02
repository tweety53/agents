package harvest_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// TestWatcherParsesFlowGuardRunVerifyList: a verify row begun by
// `flow-guard run-verify-list …` — bare or as `…/scripts/flow-guard` — is
// measured exactly as the older `run-verify-list.sh` spelling, which
// transcripts written before the port still carry.
func TestWatcherParsesFlowGuardRunVerifyList(t *testing.T) {
	for _, call := range []string{
		"flow-guard run-verify-list /wt test c verify mf-s opus medium",
		"/a/scripts/flow-guard run-verify-list /wt test c verify mf-s opus medium > /tmp/x 2>&1",
		"scripts/run-verify-list.sh /wt test c verify mf-s opus medium",
	} {
		deps := runVerifyListSettleFixture(t, call, mainThreadLine("00:02:00", 3), mainThreadLine("00:03:00", 4))
		if p, ok := deps.restamps[7]; !ok || mainOutput(t, p) != 7 {
			t.Errorf("%s: verify restamp = %s (%v), want output 7", call, p, ok)
		}
	}
}

// TestWatcherParsesFlowGuardPanelRoundBegins: `flow-guard
// panel-round-begins …` — bare or as `…/scripts/flow-guard` — yields one
// begin per slot, as the older `panel-round-begins.sh` spelling does.
func TestWatcherParsesFlowGuardPanelRoundBegins(t *testing.T) {
	for _, prefix := range []string{
		"flow-guard panel-round-begins",
		"/a/scripts/flow-guard panel-round-begins",
		"/a/skills/flow/scripts/panel-round-begins.sh",
	} {
		dir := t.TempDir()
		path := filepath.Join(dir, "session.jsonl")
		launch2 := strings.Replace(stampLaunchLine, "a68cee7239419a7e7", "a8884ead3c626d980", 1)
		launch2 = strings.Replace(launch2, "00:00:01Z", "00:00:02Z", 1)
		if err := os.WriteFile(path, stampFixtureLines(stampLaunchLine, launch2), 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		deps := &stampRecordingDeps{}
		w := harvest.NewWatcher([]harvest.Source{harvest.NewClaudeSource(dir)}, newFakeHarvestSink(), harvest.NewAttributor(&fakeWindowSource{}), deps, nil)
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (launch batch): %v", err)
		}
		commandJSON, err := json.Marshal(prefix + " /wt kan-322 4 mf-kan322 primary:opus:high principles:opus:high:origin/main")
		if err != nil {
			t.Fatalf("marshal command: %v", err)
		}
		appendLine(t, path, `{"type":"assistant","timestamp":"2026-01-01T00:00:03Z","sessionId":"s","message":{"model":"m","usage":{"input_tokens":1},"content":[{"type":"tool_use","name":"Bash","input":{"command":`+string(commandJSON)+`}}]}}`)
		if _, err := w.RunOnce(context.Background()); err != nil {
			t.Fatalf("RunOnce (begin batch): %v", err)
		}
		keys := map[string]bool{}
		for _, got := range deps.stamps {
			keys[got.key] = true
		}
		if len(deps.stamps) != 2 || !keys["panel-4-primary"] || !keys["panel-4-principles"] {
			t.Errorf("%s: stamps = %+v, want panel-4-primary and panel-4-principles", prefix, deps.stamps)
		}
	}
}
