// KAN-172 task 6: the test that would have caught the whole defect.
//
// KAN-16 shipped a measurement system that could not measure: every stage
// run was recorded with session_id NULL, so the harvester read 2,988
// transcript offsets and attributed all of it to nothing (proposal.md's
// own measurement, 2026-08-14) -- and no test in this repository noticed,
// because Go tests hand-seeded metrics directly (MergeMetrics,
// internal/store/stageruns_test.go) and every fake in this package's own
// watcher_test.go stands in for *store.Store rather than being one. No
// test had ever driven mark -> transcript -> binding -> attribution as one
// path against a real database, end to end. This file is that test.
//
// Unlike the rest of this package's tests, these drive a real
// *store.Store (internal/store/testsupport_test.go's own pattern,
// duplicated here for the same reason internal/reconcile's
// testsupport_test.go duplicates it rather than importing internal/store's
// unexported test helpers: package store_test's helpers are deliberately
// internal to that package) against the flow-postgres compose stack on
// host port 5433, wired in as harvest.HarvestSink, harvest.WindowSource
// (via a local adapter mirroring cmd/flowd/main.go's storeWindowSource),
// harvest.SessionTokenBinder and harvest.Pricer all at once -- exactly how
// cmd/flowd/main.go wires a real Watcher, and nothing like the
// per-interface fakes the rest of this package's tests use. A test here
// skips cleanly, with a clear message, when the stack is not reachable.
package harvest_test

import (
	"context"
	"encoding/json"
	"fmt"
	dsnutil "github.com/tweety53/agents/stats/internal/dsn"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/store"
)

// --- real-store test support, mirroring internal/store/testsupport_test.go ---

func e2eAdminDSN() string {
	if v := os.Getenv("FLOW_STATS_ADMIN_DSN"); v != "" {
		return v
	}
	return "postgres://flow:flow@localhost:5433/flow?sslmode=disable"
}

// Derives from e2eAdminDSN so FLOW_STATS_ADMIN_DSN moves both connections
// together; see internal/store/testsupport_test.go's testDSN for why repeating
// the credentials here makes the override only half-work.
func e2eTestDSN(dbName string) string {
	out, err := dsnutil.ForDatabase(e2eAdminDSN(), dbName)
	if err != nil {
		// Panic rather than return a best-effort string. This helper's
		// predecessor answered confidently when it could not do the job and
		// silently handed back a corrupted DSN; a test that then connects
		// reports whatever that connection says, which is the wrong question
		// answered convincingly.
		panic(fmt.Sprintf("deriving a per-test DSN: %v", err))
	}
	return out
}

// newEndToEndStore creates a uniquely-named, migrated database against the
// compose stack and returns a *store.Store whose Close and whose database
// drop are both registered as test cleanup. It skips cleanly, with a clear
// message, when the stack is not reachable -- the same guarantee
// internal/store's and internal/reconcile's own newTestStore helpers give
// their packages, duplicated here rather than exported, for the same
// reason internal/reconcile's copy gives (its own doc comment): those
// helpers live in package store_test, deliberately internal to that
// package.
func newEndToEndStore(t *testing.T) *store.Store {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	adminPool, err := pgxpool.New(ctx, e2eAdminDSN())
	if err != nil {
		t.Skipf("flow-postgres compose stack not reachable: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("flow-postgres compose stack not reachable: %v", err)
	}

	dbName := fmt.Sprintf("flow_test_e2e_%d_%d", os.Getpid(), time.Now().UnixNano())
	ident := pgx.Identifier{dbName}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		adminPool.Close()
		t.Fatalf("create test database %s: %v", dbName, err)
	}
	adminPool.Close()

	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer dropCancel()

		dropPool, err := pgxpool.New(dropCtx, e2eAdminDSN())
		if err != nil {
			t.Logf("drop test database %s: reconnect failed: %v", dbName, err)
			return
		}
		defer dropPool.Close()

		if _, err := dropPool.Exec(dropCtx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
			t.Logf("drop test database %s: %v", dbName, err)
		}
	})

	dsn := e2eTestDSN(dbName)

	openCtx, openCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer openCancel()

	st, err := store.Open(openCtx, dsn)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	if err := st.RunMigrations(openCtx); err != nil {
		st.Close()
		t.Fatalf("run migrations: %v", err)
	}
	t.Cleanup(st.Close)

	if err := st.SeedPricing(openCtx); err != nil {
		t.Fatalf("seed pricing: %v", err)
	}

	return st
}

// e2eWindowSource adapts *store.Store to harvest.WindowSource, exactly as
// cmd/flowd/main.go's own storeWindowSource does -- duplicated here
// rather than exported, since that type is unexported in package main and
// this is the one other place a real Watcher is wired against a real
// store in this repository.
type e2eWindowSource struct{ st *store.Store }

func (s e2eWindowSource) WindowsForSession(ctx context.Context, sessionID string) ([]harvest.Window, error) {
	runs, _, err := s.st.QueryStageRuns(ctx, store.Query{
		Filters: []store.Filter{{Field: "session_id", Op: store.OpEq, Value: sessionID}},
		Limit:   store.NoLimit,
	})
	if err != nil {
		return nil, err
	}
	windows := make([]harvest.Window, 0, len(runs))
	for _, r := range runs {
		windows = append(windows, harvest.Window{
			StageRunID: r.ID,
			Attempt:    r.Attempt,
			SessionID:  sessionID,
			StartedAt:  r.StartedAt,
			EndedAt:    r.EndedAt,
		})
	}
	return windows, nil
}

var _ harvest.WindowSource = e2eWindowSource{}

// --- the synthetic transcript, built from a mirrored wire shape rather than by hand ---
//
// internal/harvest/transcript.go's own rawLine/rawMessage/rawContentBlock/
// rawBashInput/rawUsage are the real decode shape -- but they are
// unexported (deliberately: that package only ever reads this format, per
// its own package doc), and this file lives in package harvest_test, so
// they cannot be marshalled directly. wireLine and its nested types below
// mirror those unexported types field-for-field and tag-for-tag, so a
// transcript line is still produced by marshalling a Go value rather than
// hand-assembling a JSON string. Field names were checked against two
// independent sources on 2026-08-14, not assumed: (1)
// internal/harvest/transcript.go's own rawLine et al., and (2) a live
// session transcript
// (~/.claude/projects/-Users-tweety53-Projects-agents/82be9370-fe07-40fc-8f38-f50720db3bfd.jsonl,
// read-only, content not copied) -- confirming "sessionId", "isSidechain",
// message.usage's "cache_creation.ephemeral_5m_input_tokens" /
// "ephemeral_1h_input_tokens", "output_tokens_details.thinking_tokens",
// and the Bash tool_use block's "input.command" all appear exactly as
// transcript.go decodes them. This is the one place in this test file
// that builds transcript JSON at all; every other assertion is against
// real Go values the store and the harvester themselves produced.
type wireLine struct {
	Type        string       `json:"type"`
	Timestamp   string       `json:"timestamp"`
	SessionID   string       `json:"sessionId"`
	IsSidechain bool         `json:"isSidechain"`
	Message     *wireMessage `json:"message"`
}

type wireMessage struct {
	Model   string             `json:"model"`
	Usage   *wireUsage         `json:"usage,omitempty"`
	Content []wireContentBlock `json:"content,omitempty"`
}

type wireContentBlock struct {
	Type  string         `json:"type"`
	Name  string         `json:"name"`
	Input *wireBashInput `json:"input,omitempty"`
}

type wireBashInput struct {
	Command string `json:"command"`
}

type wireUsage struct {
	InputTokens          int64 `json:"input_tokens"`
	CacheReadInputTokens int64 `json:"cache_read_input_tokens"`
	OutputTokens         int64 `json:"output_tokens"`
}

// marshalLine marshals l and appends the trailing newline every JSONL
// record needs -- SplitCompleteLines (transcript.go) only ever treats a
// newline-terminated span as "complete".
func marshalLine(t *testing.T, l wireLine) []byte {
	t.Helper()
	b, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal transcript line: %v", err)
	}
	return append(b, '\n')
}

// writeLines writes every line to path, creating parent directories as
// needed, appending rather than truncating when the file already exists
// -- a real transcript is append-only, and TestMarkedStageBindsToRealSessionEndToEnd
// relies on that shape when it appends a second batch mid-test.
func writeLines(t *testing.T, path string, lines ...[]byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	for _, l := range lines {
		if _, err := f.Write(l); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
}

// TestMarkedStageBindsToRealSessionEndToEnd is task 6's own test: mark a
// stage with a session token exactly as `flow stage begin` would (a real
// BeginStageInput carrying SessionToken, unbound), write a synthetic
// transcript containing that token in the recorded-command position
// (message.content[].type == "tool_use", name == "Bash",
// input.command carrying the token -- the exact shape
// ParseCommandRecords, transcript.go, reads) alongside a usage entry
// carrying the matching sessionId, drive a real Watcher wired to a real
// *store.Store cycle by cycle, and assert the stage run ends up bound to
// that session and carrying real token counts and a cost_usd -- the two
// halves KAN-16 shipped as permanently empty.
//
// Written and run against 88db0a2 (this change's own base commit, before
// any of tasks 1-5 existed): it fails to compile at all -- SessionToken,
// UnresolvedSessionTokens, BindSession, harvest.WithSessionTokenBinder and
// the -session-token wire shape none existed yet. Run instead against a
// version of this repository with the session_id column and BeginStage
// present but never wired to any nonce/token mechanism (the actual shape
// of the shipped defect, KAN-16 as released), it fails at runtime: the
// stage run's session_id stays NULL, its metrics stay `{}`, and the
// assertions below report exactly that -- "session = <nil>, want bound"
// and zero tokens, zero cost. That is the defect this test exists to
// catch, seen directly rather than reasoned about.
func TestMarkedStageBindsToRealSessionEndToEnd(t *testing.T) {
	st := newEndToEndStore(t)
	ctx := context.Background()

	projectKey := fmt.Sprintf("proj-e2e-%d", time.Now().UnixNano())
	changeName := "kan-1"
	if err := st.PutChange(ctx, store.Change{
		ProjectKey:       projectKey,
		MainCheckoutPath: "/Users/tweety53/Projects/" + projectKey,
		Name:             changeName,
		State:            store.StateInProgress,
		UpdatedAt:        time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC),
		UpdatedBy:        "flow-do",
	}); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	sessionToken := "mf-e2e-e0e1e2e3e4e5"
	const sessionID = "session-e2e-real"
	started := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)

	run, err := st.BeginStage(ctx, store.BeginStageInput{
		ProjectKey:   projectKey,
		ChangeName:   changeName,
		Harness:      "claude-code",
		SessionToken: &sessionToken,
		Command:      "/myflow-do",
		Stage:        "do.tests",
		StartedAt:    started,
	})
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if run.SessionID != nil {
		t.Fatalf("stage run session_id at mark time = %v, want nil (unbound until the harvester finds the token)", *run.SessionID)
	}

	dir := t.TempDir()
	transcriptPath := filepath.Join(dir, "session.jsonl")

	// The mark's own turn: the assistant message a real `flow stage begin
	// -stage do.tests -session-token mf-e2e-... -harness claude-code kan-1`
	// invocation leaves in its own session's transcript (design.md, "bind
	// after the fact, by a correlator the caller writes") -- a Bash
	// tool_use block whose recorded command carries the literal token,
	// plus this same turn's own usage, exactly the shape
	// TestBindMarkAndFirstUsageInSameBatchAreBothAttributed (watcher_test.go)
	// exercises against fakes and this test now exercises against a real
	// store.
	markLine := marshalLine(t, wireLine{
		Type:      "assistant",
		Timestamp: started.Add(time.Second).Format(time.RFC3339Nano),
		SessionID: sessionID,
		Message: &wireMessage{
			Model: "claude-opus-5",
			Usage: &wireUsage{InputTokens: 5, OutputTokens: 1},
			Content: []wireContentBlock{{
				Type:  "tool_use",
				Name:  "Bash",
				Input: &wireBashInput{Command: "flow stage begin -command /myflow-do -stage do.tests -session-token " + sessionToken + " -harness claude-code kan-1"},
			}},
		},
	})
	// A second message in the same session, later in the same turn's
	// flush -- ordinary usage, the kind of message that makes up the bulk
	// of a real stage.
	usageLine := marshalLine(t, wireLine{
		Type:      "assistant",
		Timestamp: started.Add(2 * time.Second).Format(time.RFC3339Nano),
		SessionID: sessionID,
		Message: &wireMessage{
			Model: "claude-opus-5",
			Usage: &wireUsage{InputTokens: 333, CacheReadInputTokens: 40, OutputTokens: 21},
		},
	})
	writeLines(t, transcriptPath, markLine, usageLine)

	windows := e2eWindowSource{st}
	attributor := harvest.NewAttributor(windows)
	w := harvest.NewWatcher(dir, st, attributor, nil, harvest.WithSessionTokenBinder(st), harvest.WithPricer(st))

	// Cycle 1: the batch containing the mark's own turn is read, the
	// token is found and bound, and -- because it revealed a still-pending
	// token -- withheld from commit rather than attributed against a
	// window that does not exist yet (watcher.go's own doc comment on
	// RunOnce, "withholding a batch that revealed a session token").
	if _, err := w.RunOnce(ctx); err != nil {
		t.Fatalf("RunOnce (cycle 1, binding): %v", err)
	}

	bound, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun after cycle 1: %v", err)
	}
	if bound.SessionID == nil || *bound.SessionID != sessionID {
		t.Fatalf("stage run session_id after cycle 1 = %v, want %q -- this is the defect KAN-16 shipped: every stage run's session_id stayed NULL", bound.SessionID, sessionID)
	}

	// Cycle 2: the withheld batch is re-read, now against a window the
	// bound session actually opens, attributed, committed, and priced.
	if _, err := w.RunOnce(ctx); err != nil {
		t.Fatalf("RunOnce (cycle 2, attribution): %v", err)
	}

	final, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun after cycle 2: %v", err)
	}

	var bag struct {
		CostUSD float64 `json:"cost_usd"`
		Tokens  struct {
			Main struct {
				Input     float64 `json:"input"`
				Output    float64 `json:"output"`
				CacheRead float64 `json:"cache_read"`
			} `json:"main"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(final.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics %s: %v", final.Metrics, err)
	}

	// Assert the money, not just the plumbing: a bound session_id alone
	// proves half of it. "Attributed" is what was broken, and cost is what
	// an operator staring at a dashboard actually looks at.
	wantInput := float64(5 + 333)
	if bag.Tokens.Main.Input != wantInput {
		t.Errorf("tokens.main.input = %v, want %v (both messages' usage, neither lost to the withheld-batch mechanism)", bag.Tokens.Main.Input, wantInput)
	}
	if bag.Tokens.Main.Output != 22 {
		t.Errorf("tokens.main.output = %v, want 22", bag.Tokens.Main.Output)
	}
	if bag.Tokens.Main.CacheRead != 40 {
		t.Errorf("tokens.main.cache_read = %v, want 40", bag.Tokens.Main.CacheRead)
	}
	if bag.CostUSD <= 0 {
		t.Errorf("cost_usd = %v, want a positive real cost -- this is the other half of the defect: a bound session with zero cost is still not what an operator can act on", bag.CostUSD)
	}
	// claude-opus-5's seeded rate (pricing_seed.go): $5/Mtok input,
	// $25/Mtok output, $0.50/Mtok cache read.
	wantCost := wantInput/1_000_000*5 + 22.0/1_000_000*25 + 40.0/1_000_000*0.50
	if diff := bag.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("cost_usd = %v, want %v (the published claude-opus-5 rate applied to the attributed tokens)", bag.CostUSD, wantCost)
	}
}

// TestUnmarkedTokenStaysRecordedAndUnattributed is the negative half task
// 6 calls out explicitly: a stage run whose session token never appears in
// any transcript must end up recorded and unattributed -- not bound to
// something, not zero-as-if-measured, but honestly absent, exactly the
// third arm task 5 gave the interface (design.md, "the third arm of the
// absence distinction"). This is the honest-degradation path a harness
// with no transcript at all (Cursor, Codex) takes on every mark, and the
// half most likely to rot unnoticed since it produces no error anywhere.
func TestUnmarkedTokenStaysRecordedAndUnattributed(t *testing.T) {
	st := newEndToEndStore(t)
	ctx := context.Background()

	projectKey := fmt.Sprintf("proj-e2e-neg-%d", time.Now().UnixNano())
	changeName := "kan-1"
	if err := st.PutChange(ctx, store.Change{
		ProjectKey:       projectKey,
		MainCheckoutPath: "/Users/tweety53/Projects/" + projectKey,
		Name:             changeName,
		State:            store.StateInProgress,
		UpdatedAt:        time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC),
		UpdatedBy:        "flow-do",
	}); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	sessionToken := "mf-e2e-never-appears"
	started := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)

	run, err := st.BeginStage(ctx, store.BeginStageInput{
		ProjectKey:   projectKey,
		ChangeName:   changeName,
		Harness:      "codex",
		SessionToken: &sessionToken,
		Command:      "/myflow-do",
		Stage:        "do.tests",
		StartedAt:    started,
	})
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	// No transcript is ever written: the "harness produces no transcript
	// at all" case design.md's rejected-alternatives section calls out
	// for Cursor and Codex. An empty directory is enough to prove the
	// watcher does not invent a session for it.
	dir := t.TempDir()
	windows := e2eWindowSource{st}
	attributor := harvest.NewAttributor(windows)
	w := harvest.NewWatcher(dir, st, attributor, nil, harvest.WithSessionTokenBinder(st), harvest.WithPricer(st))

	if _, err := w.RunOnce(ctx); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if got.SessionID != nil {
		t.Fatalf("session_id = %v, want nil: a session is never guessed (design.md)", *got.SessionID)
	}
	if len(got.Metrics) != 0 && string(got.Metrics) != "{}" {
		t.Fatalf("metrics = %s, want empty/{} -- recorded, not measured, never a guessed zero", got.Metrics)
	}
}
