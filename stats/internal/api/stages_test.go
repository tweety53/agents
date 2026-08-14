package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// stageRunRecord is fakeStore's in-memory stand-in for a stage_runs row.
// store.StageRun itself carries only a numeric ChangeID (an internal
// database detail this fake never allocates), so the owning project and
// change's public identity is kept alongside it here instead.
type stageRunRecord struct {
	run        store.StageRun
	projectKey string
	changeName string
}

// BeginStage looks the owning change up by its public identity -- exactly
// as store.Store.BeginStage does via its INSERT ... SELECT -- and reports
// store.ErrChangeNotFound when no such change is registered in f.changes,
// so stageHandler.begin's synthetic-change bootstrap has something real to
// react to.
func (f *fakeStore) BeginStage(_ context.Context, in store.BeginStageInput) (store.StageRun, error) {
	if f.beginStageErr != nil {
		return store.StageRun{}, f.beginStageErr
	}
	if _, ok := f.changes[changeKey(in.ProjectKey, in.ChangeName)]; !ok {
		return store.StageRun{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, in.ProjectKey, in.ChangeName)
	}

	attempt := 0
	for _, r := range f.stageRuns {
		if r.projectKey == in.ProjectKey && r.changeName == in.ChangeName &&
			r.run.Command == in.Command && r.run.Stage == in.Stage && r.run.Attempt > attempt {
			attempt = r.run.Attempt
		}
	}
	attempt++

	f.nextStageRunID++
	run := store.StageRun{
		ID:        f.nextStageRunID,
		RepoRoot:  in.RepoRoot,
		Harness:   in.Harness,
		SessionID: in.SessionID,
		Command:   in.Command,
		Stage:     in.Stage,
		Attempt:   attempt,
		StartedAt: in.StartedAt,
		Metrics:   json.RawMessage(`{}`),
	}
	f.stageRuns = append(f.stageRuns, stageRunRecord{run: run, projectKey: in.ProjectKey, changeName: in.ChangeName})
	return run, nil
}

func (f *fakeStore) EndStage(_ context.Context, stageRunID int64, endedAt time.Time, outcome string) error {
	if f.endStageErr != nil {
		return f.endStageErr
	}
	for i := range f.stageRuns {
		if f.stageRuns[i].run.ID == stageRunID {
			ea := endedAt
			oc := outcome
			f.stageRuns[i].run.EndedAt = &ea
			f.stageRuns[i].run.Outcome = &oc
			return nil
		}
	}
	return fmt.Errorf("%w: %d", store.ErrStageRunNotFound, stageRunID)
}

// MergeMetrics merges patch into the recorded run's metrics bag the same
// way store.Store.MergeMetrics documents: recursively, so a sibling key
// under a shared parent that patch does not mention survives.
func (f *fakeStore) MergeMetrics(_ context.Context, stageRunID int64, patch json.RawMessage) error {
	if f.mergeMetricsErr != nil {
		return f.mergeMetricsErr
	}
	for i := range f.stageRuns {
		if f.stageRuns[i].run.ID == stageRunID {
			merged, err := deepMergeJSON(f.stageRuns[i].run.Metrics, patch)
			if err != nil {
				return err
			}
			f.stageRuns[i].run.Metrics = merged
			return nil
		}
	}
	return fmt.Errorf("%w: %d", store.ErrStageRunNotFound, stageRunID)
}

// QueryStageRuns supports exactly the shape stageHandler.end's
// findOpenStageRun builds -- equality filters on project/name/command/stage,
// an isnull filter on ended_at, a single sort key, and a limit -- which is
// the entire query surface this fake's callers exercise.
func (f *fakeStore) QueryStageRuns(_ context.Context, q store.Query) ([]store.StageRun, int, error) {
	if f.queryStageRunsErr != nil {
		return nil, 0, f.queryStageRunsErr
	}

	var matches []stageRunRecord
	for _, r := range f.stageRuns {
		if stageRunMatchesFilters(r, q.Filters) {
			matches = append(matches, r)
		}
	}

	if len(q.Sort) > 0 && q.Sort[0].Field == "attempt" {
		desc := q.Sort[0].Desc
		for i := 1; i < len(matches); i++ {
			for j := i; j > 0; j-- {
				less := matches[j-1].run.Attempt < matches[j].run.Attempt
				if desc {
					less = matches[j-1].run.Attempt > matches[j].run.Attempt
				}
				if less {
					break
				}
				matches[j-1], matches[j] = matches[j], matches[j-1]
			}
		}
	}

	total := len(matches)
	if q.Limit > 0 && len(matches) > q.Limit {
		matches = matches[:q.Limit]
	}

	out := make([]store.StageRun, len(matches))
	for i, r := range matches {
		out[i] = r.run
	}
	return out, total, nil
}

func stageRunMatchesFilters(r stageRunRecord, filters []store.Filter) bool {
	for _, f := range filters {
		switch f.Field {
		case "project":
			if r.projectKey != f.Value {
				return false
			}
		case "name":
			if r.changeName != f.Value {
				return false
			}
		case "command":
			if r.run.Command != f.Value {
				return false
			}
		case "stage":
			if r.run.Stage != f.Value {
				return false
			}
		case "ended_at":
			switch f.Op {
			case store.OpNull:
				if r.run.EndedAt != nil {
					return false
				}
			case store.OpNotNull:
				if r.run.EndedAt == nil {
					return false
				}
			}
		}
	}
	return true
}

// deepMergeJSON merges patch into base recursively -- a JSON object at a
// shared key merges key by key at every depth, and any other value type
// simply replaces what was there -- mirroring jsonb_deep_merge
// (0003_stage_runs.sql) closely enough for this fake's purposes: this
// package's tests assert on merge outcomes, not on the SQL function
// itself, which internal/store's own tests already pin directly.
func deepMergeJSON(base, patch json.RawMessage) (json.RawMessage, error) {
	var baseVal, patchVal any
	if len(base) == 0 {
		base = []byte(`{}`)
	}
	if err := json.Unmarshal(base, &baseVal); err != nil {
		return nil, fmt.Errorf("fake store: decode base metrics: %w", err)
	}
	if err := json.Unmarshal(patch, &patchVal); err != nil {
		return nil, fmt.Errorf("fake store: decode metrics patch: %w", err)
	}
	merged := deepMergeValue(baseVal, patchVal)
	out, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("fake store: encode merged metrics: %w", err)
	}
	return out, nil
}

func deepMergeValue(base, patch any) any {
	baseMap, baseIsMap := base.(map[string]any)
	patchMap, patchIsMap := patch.(map[string]any)
	if !baseIsMap || !patchIsMap {
		return patch
	}
	merged := make(map[string]any, len(baseMap)+len(patchMap))
	for k, v := range baseMap {
		merged[k] = v
	}
	for k, v := range patchMap {
		if existing, ok := merged[k]; ok {
			merged[k] = deepMergeValue(existing, v)
		} else {
			merged[k] = v
		}
	}
	return merged
}

var _ api.StageStore = (*fakeStore)(nil)

// --- test helpers ---

func newStageTestServer(t *testing.T, fs *fakeStore) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	srv, err := api.New(cfg, fs, fs, fs, nil)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	return httptest.NewServer(srv.Handler())
}

func postJSON(t *testing.T, url string, body any) (*http.Response, []byte) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	respBody := readBody(t, resp)
	return resp, respBody
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return buf.Bytes()
}

// --- begin ---

// TestStageBeginRecordsIdentityAndInstant pins that a begin mark, once
// accepted, is recorded against the store carrying exactly the identity
// and start instant the request named -- command, stage, harness, session
// id and started-at -- not merely that the request "succeeded".
func TestStageBeginRecordsIdentityAndInstant(t *testing.T) {
	fs := newFakeStore()
	fs.changes[changeKey("proj", "chg")] = store.Change{ProjectKey: "proj", Name: "chg", State: store.StateStarted}
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	session := "sess-1"
	req := map[string]any{
		"projectKey": "proj",
		"changeName": "chg",
		"harness":    "claude-code",
		"sessionId":  session,
		"command":    "/myflow-do",
		"stage":      "SDD + TDD per task",
		"startedAt":  "2026-08-13T10:00:00Z",
	}
	resp, body := postJSON(t, srv.URL+"/api/v1/stages/begin", req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", resp.StatusCode, body)
	}

	if len(fs.stageRuns) != 1 {
		t.Fatalf("len(stageRuns) = %d, want 1", len(fs.stageRuns))
	}
	run := fs.stageRuns[0].run
	if run.Command != "/myflow-do" {
		t.Errorf("Command = %q, want /myflow-do", run.Command)
	}
	if run.Stage != "SDD + TDD per task" {
		t.Errorf("Stage = %q, want %q", run.Stage, "SDD + TDD per task")
	}
	if run.Harness != "claude-code" {
		t.Errorf("Harness = %q, want claude-code", run.Harness)
	}
	if run.SessionID == nil || *run.SessionID != session {
		t.Errorf("SessionID = %v, want %q", run.SessionID, session)
	}
	wantStart, _ := time.Parse(time.RFC3339, "2026-08-13T10:00:00Z")
	if !run.StartedAt.Equal(wantStart) {
		t.Errorf("StartedAt = %v, want %v", run.StartedAt, wantStart)
	}
	if run.Attempt != 1 {
		t.Errorf("Attempt = %d, want 1", run.Attempt)
	}

	var decoded struct {
		StageRunID int64 `json:"stageRunId"`
		Attempt    int   `json:"attempt"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded.StageRunID != run.ID || decoded.Attempt != 1 {
		t.Errorf("response = %+v, want StageRunID=%d Attempt=1", decoded, run.ID)
	}
}

// TestStageBeginRejectsUndocumentedStage is the server-side half of
// task 8's rejection requirement: even though the CLI validates first
// (cmd/myflow/stage_test.go), the daemon itself must never persist a
// stage name absent from internal/stages' documented table.
func TestStageBeginRejectsUndocumentedStage(t *testing.T) {
	fs := newFakeStore()
	fs.changes[changeKey("proj", "chg")] = store.Change{ProjectKey: "proj", Name: "chg", State: store.StateStarted}
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	req := map[string]any{
		"projectKey": "proj",
		"changeName": "chg",
		"harness":    "claude-code",
		"command":    "/myflow-do",
		"stage":      "a stage nobody documented",
		"startedAt":  "2026-08-13T10:00:00Z",
	}
	resp, body := postJSON(t, srv.URL+"/api/v1/stages/begin", req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", resp.StatusCode, body)
	}
	if len(fs.stageRuns) != 0 {
		t.Errorf("len(stageRuns) = %d, want 0 -- an undocumented stage must never be recorded", len(fs.stageRuns))
	}
}

// TestMarkForUnknownChangeIsStored pins design.md's "a mark for an unknown
// change is stored, not dropped": a begin mark naming a change the store
// has never heard of is recorded against a synthetic change row for the
// project, rather than refused.
func TestMarkForUnknownChangeIsStored(t *testing.T) {
	fs := newFakeStore()
	// Deliberately no pre-registered change for "proj"/"chg-nobody-made".
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	req := map[string]any{
		"projectKey":       "proj",
		"mainCheckoutPath": "/repo/proj",
		"changeName":       "chg-nobody-made",
		"harness":          "claude-code",
		"command":          "/myflow-do",
		"stage":            "SDD + TDD per task",
		"startedAt":        "2026-08-13T10:00:00Z",
	}
	resp, body := postJSON(t, srv.URL+"/api/v1/stages/begin", req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (a mark for an unknown change must be stored, not refused); body: %s", resp.StatusCode, body)
	}

	change, ok := fs.changes[changeKey("proj", "chg-nobody-made")]
	if !ok {
		t.Fatal("no synthetic change row was created for the unknown change")
	}
	if change.State != store.StateStarted {
		t.Errorf("synthetic change state = %q, want STARTED", change.State)
	}
	if change.UpdatedBy == "" || change.UpdatedBy == "tester" {
		t.Errorf("synthetic change UpdatedBy = %q, want a value marking it as synthetic", change.UpdatedBy)
	}

	if len(fs.stageRuns) != 1 {
		t.Fatalf("len(stageRuns) = %d, want 1 -- the mark itself must still be recorded", len(fs.stageRuns))
	}
	if fs.stageRuns[0].changeName != "chg-nobody-made" {
		t.Errorf("recorded stage run's change = %q, want chg-nobody-made", fs.stageRuns[0].changeName)
	}
}

// --- end ---

// TestStageEndRecordsOutcomeAndMetrics pins that ending a stage records
// the end instant and outcome on the run `begin` opened, and that metrics
// passed alongside are merged into its metrics bag rather than replacing
// it -- a second end call's metrics must not erase what a first call (or
// a concurrent harvester merge) already wrote.
func TestStageEndRecordsOutcomeAndMetrics(t *testing.T) {
	fs := newFakeStore()
	fs.changes[changeKey("proj", "chg")] = store.Change{ProjectKey: "proj", Name: "chg", State: store.StateStarted}
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	beginReq := map[string]any{
		"projectKey": "proj",
		"changeName": "chg",
		"harness":    "claude-code",
		"command":    "/myflow-do",
		"stage":      "SDD + TDD per task",
		"startedAt":  "2026-08-13T10:00:00Z",
	}
	if resp, body := postJSON(t, srv.URL+"/api/v1/stages/begin", beginReq); resp.StatusCode != http.StatusOK {
		t.Fatalf("begin: status = %d; body: %s", resp.StatusCode, body)
	}

	// Seed a sibling metrics key the way a harvester merge would, before
	// `end` runs -- MergeMetrics' deep-merge guarantee is what this test
	// actually exercises: `end`'s own outcome-adjacent metrics must not
	// erase it.
	if err := fs.MergeMetrics(context.Background(), fs.stageRuns[0].run.ID, json.RawMessage(`{"tokens":{"input":100}}`)); err != nil {
		t.Fatalf("seed metrics: %v", err)
	}

	endReq := map[string]any{
		"projectKey": "proj",
		"changeName": "chg",
		"command":    "/myflow-do",
		"stage":      "SDD + TDD per task",
		"endedAt":    "2026-08-13T10:05:00Z",
		"outcome":    "completed",
		"metrics":    json.RawMessage(`{"fix_rounds":2,"tokens":{"output":50}}`),
	}
	resp, body := postJSON(t, srv.URL+"/api/v1/stages/end", endReq)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("end: status = %d, want 200; body: %s", resp.StatusCode, body)
	}

	if len(fs.stageRuns) != 1 {
		t.Fatalf("len(stageRuns) = %d, want 1", len(fs.stageRuns))
	}
	run := fs.stageRuns[0].run
	if run.EndedAt == nil {
		t.Fatal("EndedAt is nil, want the end instant recorded")
	}
	wantEnd, _ := time.Parse(time.RFC3339, "2026-08-13T10:05:00Z")
	if !run.EndedAt.Equal(wantEnd) {
		t.Errorf("EndedAt = %v, want %v", *run.EndedAt, wantEnd)
	}
	if run.Outcome == nil || *run.Outcome != "completed" {
		t.Errorf("Outcome = %v, want completed", run.Outcome)
	}

	var metrics map[string]any
	if err := json.Unmarshal(run.Metrics, &metrics); err != nil {
		t.Fatalf("decode metrics: %v", err)
	}
	if metrics["fix_rounds"] != float64(2) {
		t.Errorf(`metrics["fix_rounds"] = %v, want 2`, metrics["fix_rounds"])
	}
	tokens, ok := metrics["tokens"].(map[string]any)
	if !ok {
		t.Fatalf("metrics[tokens] = %v, want an object", metrics["tokens"])
	}
	if tokens["input"] != float64(100) {
		t.Errorf(`the harvester-seeded tokens.input key was destroyed by end's merge: got %v, want 100 -- MergeMetrics must merge recursively, never replace the "tokens" object wholesale`, tokens["input"])
	}
	if tokens["output"] != float64(50) {
		t.Errorf(`metrics.tokens.output = %v, want 50`, tokens["output"])
	}
}

// TestStageEndClosesHighestOpenAttempt pins findOpenStageRun's tie-break
// (stages.go: `ORDER BY attempt DESC LIMIT 1`) directly: when two attempts
// for the same (project, change, command, stage) triple are open at once
// -- design.md's "A replayed begin mark can open a second attempt" records
// exactly this path, a `stage begin` that times out client-side after
// committing, is journalled, and replayed, opening a genuine second
// attempt -- `stage end` must close the higher-numbered attempt and leave
// the orphaned, lower-numbered one open for task 10's sweeper. Closing the
// wrong one would attribute that end's outcome and metrics to an attempt
// no live session is actually writing to.
func TestStageEndClosesHighestOpenAttempt(t *testing.T) {
	fs := newFakeStore()
	fs.changes[changeKey("proj", "chg")] = store.Change{ProjectKey: "proj", Name: "chg", State: store.StateStarted}
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	beginReq := map[string]any{
		"projectKey": "proj",
		"changeName": "chg",
		"harness":    "claude-code",
		"command":    "/myflow-do",
		"stage":      "SDD + TDD per task",
	}

	// A normal begin -- attempt 1.
	beginReq["startedAt"] = "2026-08-13T10:00:00Z"
	if resp, body := postJSON(t, srv.URL+"/api/v1/stages/begin", beginReq); resp.StatusCode != http.StatusOK {
		t.Fatalf("first begin: status = %d; body: %s", resp.StatusCode, body)
	}

	// A second begin for the identical triple, simulating a replayed begin
	// mark -- BeginStage has no idempotency key, so this genuinely opens a
	// second attempt exactly as design.md describes.
	beginReq["startedAt"] = "2026-08-13T10:01:00Z"
	if resp, body := postJSON(t, srv.URL+"/api/v1/stages/begin", beginReq); resp.StatusCode != http.StatusOK {
		t.Fatalf("second (replayed) begin: status = %d; body: %s", resp.StatusCode, body)
	}

	if len(fs.stageRuns) != 2 {
		t.Fatalf("len(stageRuns) = %d, want 2 open attempts", len(fs.stageRuns))
	}
	if fs.stageRuns[0].run.Attempt != 1 || fs.stageRuns[1].run.Attempt != 2 {
		t.Fatalf("attempts = %d, %d; want 1, 2", fs.stageRuns[0].run.Attempt, fs.stageRuns[1].run.Attempt)
	}

	endReq := map[string]any{
		"projectKey": "proj",
		"changeName": "chg",
		"command":    "/myflow-do",
		"stage":      "SDD + TDD per task",
		"endedAt":    "2026-08-13T10:05:00Z",
		"outcome":    "completed",
	}
	if resp, body := postJSON(t, srv.URL+"/api/v1/stages/end", endReq); resp.StatusCode != http.StatusOK {
		t.Fatalf("end: status = %d, want 200; body: %s", resp.StatusCode, body)
	}

	attempt1 := fs.stageRuns[0].run
	attempt2 := fs.stageRuns[1].run
	if attempt1.EndedAt != nil {
		t.Errorf("attempt 1 (the orphan) was closed by end; it must stay open for the sweeper -- EndedAt = %v", *attempt1.EndedAt)
	}
	if attempt2.EndedAt == nil {
		t.Fatal("attempt 2 (the live one) was not closed by end -- EndedAt is nil")
	}
	if attempt2.Outcome == nil || *attempt2.Outcome != "completed" {
		t.Errorf("attempt 2's Outcome = %v, want completed", attempt2.Outcome)
	}
}

// TestStageEndWithNoOpenRunIsNotFound pins that `end` reports a clear
// not-found rather than silently succeeding, or panicking, when no
// matching stage run is currently open.
func TestStageEndWithNoOpenRunIsNotFound(t *testing.T) {
	fs := newFakeStore()
	fs.changes[changeKey("proj", "chg")] = store.Change{ProjectKey: "proj", Name: "chg", State: store.StateStarted}
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	endReq := map[string]any{
		"projectKey": "proj",
		"changeName": "chg",
		"command":    "/myflow-do",
		"stage":      "SDD + TDD per task",
		"outcome":    "completed",
	}
	resp, body := postJSON(t, srv.URL+"/api/v1/stages/end", endReq)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body: %s", resp.StatusCode, body)
	}
}

// --- stage end: token unavailability, derived from the recorded harness ---

// beginThenEnd runs a full begin/end round trip against srv: a begin mark
// carrying harness, followed by an end mark carrying no harness at all --
// the wire shape design.md's own canonical example
// (`myflow stage end --change ... --outcome completed`) uses, and the only
// shape `stage end` has sent since task 10's post-commit review (finding
// F1; cmd/myflow/stage.go no longer has a -harness flag on `stage end`).
func beginThenEnd(t *testing.T, srv *httptest.Server, harness string) {
	t.Helper()

	beginReq := map[string]any{
		"projectKey": "proj",
		"changeName": "chg",
		"harness":    harness,
		"command":    "/myflow-do",
		"stage":      "SDD + TDD per task",
		"startedAt":  "2026-08-13T10:00:00Z",
	}
	if resp, body := postJSON(t, srv.URL+"/api/v1/stages/begin", beginReq); resp.StatusCode != http.StatusOK {
		t.Fatalf("begin: status = %d; body: %s", resp.StatusCode, body)
	}

	endReq := map[string]any{
		"projectKey": "proj",
		"changeName": "chg",
		"command":    "/myflow-do",
		"stage":      "SDD + TDD per task",
		"endedAt":    "2026-08-13T10:05:00Z",
		"outcome":    "completed",
	}
	if resp, body := postJSON(t, srv.URL+"/api/v1/stages/end", endReq); resp.StatusCode != http.StatusOK {
		t.Fatalf("end: status = %d, want 200; body: %s", resp.StatusCode, body)
	}
}

// TestNonClaudeHarnessMarksTokensUnavailable pins that a stage run begun
// under any harness but "claude-code" has its metrics.tokens_available
// set to the JSON boolean false once it ends -- design.md's "An
// unavailable metric is recorded as unavailable": a harness with no
// machine-readable transcript can never have its token metrics harvested,
// so `stage end` says so explicitly rather than leaving the field
// silently absent and indistinguishable from "not harvested yet".
//
// The end mark here carries no harness of its own (see beginThenEnd's own
// doc comment) -- the flag is derived entirely from what `stage begin`
// already recorded, which is the fix for finding F1.
func TestNonClaudeHarnessMarksTokensUnavailable(t *testing.T) {
	fs := newFakeStore()
	fs.changes[changeKey("proj", "chg")] = store.Change{ProjectKey: "proj", Name: "chg", State: store.StateStarted}
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	beginThenEnd(t, srv, "cursor")

	if len(fs.stageRuns) != 1 {
		t.Fatalf("len(stageRuns) = %d, want 1", len(fs.stageRuns))
	}
	var metrics map[string]any
	if err := json.Unmarshal(fs.stageRuns[0].run.Metrics, &metrics); err != nil {
		t.Fatalf("decode metrics: %v", err)
	}
	available, present := metrics["tokens_available"]
	if !present {
		t.Fatalf("metrics did not carry tokens_available for a non-claude-code harness: %s", fs.stageRuns[0].run.Metrics)
	}
	if b, ok := available.(bool); !ok || b != false {
		t.Errorf("metrics.tokens_available = %#v (%T), want the JSON boolean false", available, available)
	}
}

// TestClaudeCodeHarnessEndedWithoutHarnessIsNotMarkedUnavailable is
// finding F1's own regression test: a genuine claude-code stage run ended
// exactly the way design.md's canonical example ends one -- no -harness
// flag, no harness field on the wire at all -- must not be marked
// tokens_available: false. The flag must come from what `stage begin`
// recorded (harness: "claude-code"), never from anything the end mark
// itself does or does not carry.
//
// Mutation check: reverting ApplyEndStageMark to resolve tokens_available
// from a freshly-resolved value instead of openRun.Harness makes this
// test fail while TestNonClaudeHarnessMarksTokensUnavailable stays green
// -- confirmed locally -- which is exactly the false-negative finding F1
// reported: a measured run marked unavailable.
func TestClaudeCodeHarnessEndedWithoutHarnessIsNotMarkedUnavailable(t *testing.T) {
	fs := newFakeStore()
	fs.changes[changeKey("proj", "chg")] = store.Change{ProjectKey: "proj", Name: "chg", State: store.StateStarted}
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	beginThenEnd(t, srv, "claude-code")

	if len(fs.stageRuns) != 1 {
		t.Fatalf("len(stageRuns) = %d, want 1", len(fs.stageRuns))
	}
	var metrics map[string]any
	if len(fs.stageRuns[0].run.Metrics) > 0 {
		if err := json.Unmarshal(fs.stageRuns[0].run.Metrics, &metrics); err != nil {
			t.Fatalf("decode metrics: %v", err)
		}
	}
	if _, present := metrics["tokens_available"]; present {
		t.Errorf("metrics carried tokens_available for a claude-code harness ended without one: %v", metrics["tokens_available"])
	}
}

// TestUnavailableIsNeverWrittenAsZero pins the shape design.md's
// absence-is-not-a-value rule requires: a non-claude-code harness's ended
// stage run carries tokens_available: false and nothing else that could
// be misread as a measured value -- in particular, no "tokens" key at all
// (which would invite a reader to see {"tokens":{"input":0}} rather than
// "not measured"), and tokens_available itself is the JSON boolean false,
// never the number 0 (which a looser round-trip could otherwise hide).
func TestUnavailableIsNeverWrittenAsZero(t *testing.T) {
	fs := newFakeStore()
	fs.changes[changeKey("proj", "chg")] = store.Change{ProjectKey: "proj", Name: "chg", State: store.StateStarted}
	srv := newStageTestServer(t, fs)
	defer srv.Close()

	beginThenEnd(t, srv, "codex")

	if len(fs.stageRuns) != 1 {
		t.Fatalf("len(stageRuns) = %d, want 1", len(fs.stageRuns))
	}
	raw := fs.stageRuns[0].run.Metrics

	// The exact bytes matter here, not just the decoded shape: this
	// asserts the stored encoding is the literal token `false`, not `0`
	// or `"false"`, which a looser round-trip through map[string]any
	// could otherwise hide.
	if !bytes.Contains(raw, []byte(`"tokens_available":false`)) {
		t.Fatalf("stored metrics do not carry the literal `\"tokens_available\":false`: %s", raw)
	}

	var metrics map[string]any
	if err := json.Unmarshal(raw, &metrics); err != nil {
		t.Fatalf("decode metrics: %v", err)
	}
	if _, present := metrics["tokens"]; present {
		t.Errorf("metrics carried a \"tokens\" key for an unavailable harness: %v", metrics["tokens"])
	}
}
