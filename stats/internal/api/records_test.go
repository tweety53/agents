package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// dispatchRecord is fakeStore's in-memory stand-in for a dispatches row.
// records.Dispatch carries no owning project or change (the store resolves
// those to a change_id this fake never allocates), so the owning identity
// is kept alongside the row here -- the same shape stageRunRecord already
// uses for a stage run.
type dispatchRecord struct {
	dispatch   records.Dispatch
	projectKey string
	changeName string
}

// findingRecord is fakeStore's in-memory stand-in for a findings row. See
// dispatchRecord's doc comment for why the owning identity sits beside the
// row rather than inside it.
type findingRecord struct {
	finding    records.Finding
	projectKey string
	changeName string
}

// RecordDispatch mirrors store.Store.RecordDispatch's own contract: it
// allocates seq per change (never globally), ignores any seq the caller
// supplied, and hands back the stored row carrying both the allocated seq
// and the row's own id. A change this fake has never been told about is
// store.ErrChangeNotFound, so the handler's 404 mapping has something real
// to react to.
func (f *fakeStore) RecordDispatch(_ context.Context, projectKey, change string, in records.Dispatch) (records.Dispatch, error) {
	f.recordCalls++
	if f.recordDispatchErr != nil {
		return records.Dispatch{}, f.recordDispatchErr
	}
	if _, ok := f.changes[changeKey(projectKey, change)]; !ok {
		return records.Dispatch{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, change)
	}

	// The store's own idempotency, mirrored: a write carrying a key this
	// change has already recorded under the same session token is the same
	// logical dispatch arriving twice (a replayed journal entry, or a lost
	// response), and returns the row already there rather than allocating a
	// second seq for it.
	if in.Key != "" {
		for i := range f.dispatches {
			d := &f.dispatches[i]
			if d.projectKey == projectKey && d.changeName == change &&
				d.dispatch.Key == in.Key && d.dispatch.SessionToken == in.SessionToken {
				return d.dispatch, nil
			}
		}
	}

	seq := 0
	for _, d := range f.dispatches {
		if d.projectKey == projectKey && d.changeName == change && d.dispatch.Seq > seq {
			seq = d.dispatch.Seq
		}
	}
	seq++

	f.nextDispatchID++
	out := in
	out.ID = f.nextDispatchID
	out.Seq = seq
	if len(out.Metrics) == 0 {
		out.Metrics = json.RawMessage(`{}`)
	}
	f.dispatches = append(f.dispatches, dispatchRecord{dispatch: out, projectKey: projectKey, changeName: change})
	return out, nil
}

// EndDispatch mirrors store.Store.EndDispatch: it closes the row named by
// the session token and key its begin carried, and reports
// store.ErrDispatchNotFound for a key naming none -- the condition the
// handler must answer 404 to rather than 500, and the one 404 the CLI
// journals rather than reports.
//
// A non-empty in.AgentID overwrites the row's identifier -- this is the
// closing half of "a dispatch's identifier may be recorded when it becomes
// known", for a harness that reports it only after launch. An EMPTY
// in.AgentID leaves whatever begin already recorded untouched: an end that
// omits it must never clear an identifier the opening call captured.
func (f *fakeStore) EndDispatch(_ context.Context, projectKey, change string, in records.DispatchEnd) (records.Dispatch, error) {
	f.recordCalls++
	for i := range f.dispatches {
		d := &f.dispatches[i]
		if d.projectKey == projectKey && d.changeName == change &&
			d.dispatch.Key == in.Key && d.dispatch.SessionToken == in.SessionToken {
			endedAt := in.EndedAt
			d.dispatch.CommitSHA = in.CommitSHA
			d.dispatch.Outcome = in.Outcome
			d.dispatch.EndedAt = &endedAt
			if in.AgentID != "" {
				d.dispatch.AgentID = in.AgentID
			}
			return d.dispatch, nil
		}
	}
	return records.Dispatch{}, fmt.Errorf("%w: %s/%s key %q", store.ErrDispatchNotFound, projectKey, change, in.Key)
}

// UpsertFinding mirrors store.Store.UpsertFinding: a ref is unique per
// change, so a second write for the same ref rewrites the row rather than
// appending a second one, and the boolean reports which of the two
// happened -- the fact the handler turns into 201 rather than 200.
func (f *fakeStore) UpsertFinding(_ context.Context, projectKey, change string, in records.Finding) (records.Finding, bool, error) {
	f.recordCalls++
	if f.upsertFindingErr != nil {
		return records.Finding{}, false, f.upsertFindingErr
	}
	if _, ok := f.changes[changeKey(projectKey, change)]; !ok {
		return records.Finding{}, false, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, change)
	}

	for i := range f.findings {
		r := &f.findings[i]
		if r.projectKey == projectKey && r.changeName == change && r.finding.Ref == in.Ref {
			r.finding = in
			return in, false, nil
		}
	}
	f.findings = append(f.findings, findingRecord{finding: in, projectKey: projectKey, changeName: change})
	return in, true, nil
}

// SetFindingStatus rewrites one finding's status and nothing else, and
// reports store.ErrFindingNotFound for a ref the change holds no finding
// under -- the condition the handler must answer 404 to rather than 500.
func (f *fakeStore) SetFindingStatus(_ context.Context, projectKey, change, ref, status, category string) error {
	f.recordCalls++
	if f.setFindingStatusErr != nil {
		return f.setFindingStatusErr
	}
	for i := range f.findings {
		r := &f.findings[i]
		if r.projectKey == projectKey && r.changeName == change && r.finding.Ref == ref {
			r.finding.Status = status
			r.finding.Category = category
			return nil
		}
	}
	return fmt.Errorf("%w: %s in %s/%s", store.ErrFindingNotFound, ref, projectKey, change)
}

// RunRecord returns the change's whole record in the order store.RunRecord
// documents: dispatches by seq, findings by the digits their ref carries.
// The fake reproduces that ordering rather than returning insertion order,
// for the same reason BeginStage above reproduces attempt allocation --
// a fake that answered in a different order than the real store would let
// a handler bug hide behind it.
func (f *fakeStore) RunRecord(_ context.Context, projectKey, change string) (records.Run, error) {
	f.recordCalls++
	if f.runRecordErr != nil {
		return records.Run{}, f.runRecordErr
	}
	if _, ok := f.changes[changeKey(projectKey, change)]; !ok {
		return records.Run{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, change)
	}

	out := records.Run{Change: change}
	for _, d := range f.dispatches {
		if d.projectKey == projectKey && d.changeName == change {
			out.Dispatches = append(out.Dispatches, d.dispatch)
		}
	}
	for _, r := range f.findings {
		if r.projectKey == projectKey && r.changeName == change {
			out.Findings = append(out.Findings, r.finding)
		}
	}
	for _, p := range f.passes {
		if p.projectKey == projectKey && p.changeName == change {
			out.Passes = append(out.Passes, p.pass)
		}
	}
	for _, m := range f.mutations {
		if m.projectKey == projectKey && m.changeName == change {
			out.Mutations = append(out.Mutations, m.mutation)
		}
	}
	sort.Slice(out.Dispatches, func(i, j int) bool { return out.Dispatches[i].Seq < out.Dispatches[j].Seq })
	sort.Slice(out.Findings, func(i, j int) bool { return refLess(out.Findings[i].Ref, out.Findings[j].Ref) })
	sort.Slice(out.Passes, func(i, j int) bool { return passOrderLess(out.Passes[i], out.Passes[j]) })
	sort.Slice(out.Mutations, func(i, j int) bool { return mutationOrderLess(out.Mutations[i], out.Mutations[j]) })
	return out, nil
}

// passOrderLess and mutationOrderLess order the pass-log rows the way
// store.RunRecord's ORDER BY does: by round, then row id.
func passOrderLess(a, b records.Pass) bool {
	if a.Round != b.Round {
		return a.Round < b.Round
	}
	return a.ID < b.ID
}

func mutationOrderLess(a, b records.Mutation) bool {
	if a.Round != b.Round {
		return a.Round < b.Round
	}
	return a.ID < b.ID
}

// refLess orders two finding refs the way store.RunRecord's ORDER BY does:
// by the number the ref's digits spell, so F2 precedes F10, with a ref
// carrying no digits at all sorting last, lexically, rather than crashing
// the comparison.
func refLess(a, b string) bool {
	na, aHasDigits := refDigits(a)
	nb, bHasDigits := refDigits(b)
	switch {
	case aHasDigits && bHasDigits && na != nb:
		return na < nb
	case aHasDigits != bHasDigits:
		return aHasDigits
	default:
		return a < b
	}
}

// refDigits reads the digits a ref carries as one number, reporting
// whether it carried any at all.
func refDigits(ref string) (int, bool) {
	var digits strings.Builder
	for _, r := range ref {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	if digits.Len() == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(digits.String())
	if err != nil {
		return 0, false
	}
	return n, true
}

// verdictRecord is fakeStore's in-memory stand-in for a guard_verdicts row.
// See dispatchRecord's doc comment for why the owning identity sits beside
// the row rather than inside it.
type verdictRecord struct {
	verdict    records.Verdict
	projectKey string
	changeName string
}

// incidentRecord is fakeStore's in-memory stand-in for an incidents row.
// Unlike a verdict, an incident belongs to a project directly (its
// change_id is nullable), so only the project key is kept beside it.
type incidentRecord struct {
	incident   records.Incident
	projectKey string
}

// RecordVerdict mirrors store.Store.RecordVerdict: every call inserts a new
// row (a guard's re-entered run is a distinct verdict, never a replay), and
// an unknown (projectKey, change) pair is store.ErrChangeNotFound, the
// condition the handler must answer 404 to.
func (f *fakeStore) RecordVerdict(_ context.Context, projectKey, change string, in records.Verdict) (records.Verdict, error) {
	f.recordCalls++
	if f.recordVerdictErr != nil {
		return records.Verdict{}, f.recordVerdictErr
	}
	if _, ok := f.changes[changeKey(projectKey, change)]; !ok {
		return records.Verdict{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, change)
	}
	f.nextVerdictID++
	out := in
	out.ID = f.nextVerdictID
	out.Change = change
	f.verdicts = append(f.verdicts, verdictRecord{verdict: out, projectKey: projectKey, changeName: change})
	return out, nil
}

// FlagVerdictFalsePositive mirrors store.Store.FlagVerdictFalsePositive: it
// flags the most recently recorded verdict for (projectKey, change, guard),
// and a pair holding none is store.ErrDispatchNotFound -- the reused
// sentinel the handler must answer 404 to, exactly as the real store does.
func (f *fakeStore) FlagVerdictFalsePositive(_ context.Context, projectKey, change string, in records.VerdictFlag) (records.Verdict, error) {
	f.recordCalls++
	if f.flagVerdictErr != nil {
		return records.Verdict{}, f.flagVerdictErr
	}
	var found *verdictRecord
	for i := range f.verdicts {
		v := &f.verdicts[i]
		if v.projectKey != projectKey || v.changeName != change || v.verdict.Guard != in.Guard {
			continue
		}
		if found == nil || v.verdict.ID > found.verdict.ID {
			found = v
		}
	}
	if found == nil {
		return records.Verdict{}, fmt.Errorf("%w: verdict for guard %q in %s/%s", store.ErrDispatchNotFound, in.Guard, projectKey, change)
	}
	found.verdict.FalsePositive = true
	found.verdict.FalsePositiveReason = in.Reason
	flaggedAt := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	found.verdict.FlaggedAt = &flaggedAt
	return found.verdict, nil
}

// ListVerdicts mirrors store.Store.ListVerdicts' filtering: guard == ""
// means every guard, and falsePositiveOnly restricts to flagged rows. It
// also records the args it was called with, so a test can assert the
// handler parsed and forwarded the query rather than merely that some
// filter fired.
func (f *fakeStore) ListVerdicts(_ context.Context, projectKey, guard string, falsePositiveOnly bool) ([]records.Verdict, error) {
	f.recordCalls++
	f.lastListVerdictsGuard = guard
	f.lastListVerdictsFalsePositiveOnly = falsePositiveOnly
	if f.listVerdictsErr != nil {
		return nil, f.listVerdictsErr
	}
	var out []records.Verdict
	for i := len(f.verdicts) - 1; i >= 0; i-- {
		v := f.verdicts[i]
		if v.projectKey != projectKey {
			continue
		}
		if guard != "" && v.verdict.Guard != guard {
			continue
		}
		if falsePositiveOnly && !v.verdict.FalsePositive {
			continue
		}
		out = append(out, v.verdict)
	}
	return out, nil
}

// RecordIncident mirrors store.Store.RecordIncident: in.Change, when
// non-empty, must name a change this fake already knows about -- an unknown
// name is store.ErrChangeNotFound -- and left empty stores no change at
// all, exactly as the real store's nullable change_id does.
func (f *fakeStore) RecordIncident(_ context.Context, projectKey string, in records.Incident) (records.Incident, error) {
	f.recordCalls++
	if f.recordIncidentErr != nil {
		return records.Incident{}, f.recordIncidentErr
	}
	if in.Change != "" {
		if _, ok := f.changes[changeKey(projectKey, in.Change)]; !ok {
			return records.Incident{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, in.Change)
		}
	}
	f.nextIncidentID++
	out := in
	out.ID = f.nextIncidentID
	if out.OccurredAt.IsZero() {
		out.OccurredAt = time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC).Add(time.Duration(f.nextIncidentID) * time.Second)
	}
	f.incidents = append(f.incidents, incidentRecord{incident: out, projectKey: projectKey})
	return out, nil
}

// ListIncidents mirrors store.Store.ListIncidents' ordering: newest first.
func (f *fakeStore) ListIncidents(_ context.Context, projectKey string) ([]records.Incident, error) {
	f.recordCalls++
	if f.listIncidentsErr != nil {
		return nil, f.listIncidentsErr
	}
	var out []records.Incident
	for i := len(f.incidents) - 1; i >= 0; i-- {
		if f.incidents[i].projectKey == projectKey {
			out = append(out, f.incidents[i].incident)
		}
	}
	return out, nil
}

// --- decisions ---

// decisionRecord is fakeStore's in-memory stand-in for a decisions row. See
// dispatchRecord's doc comment for why the owning identity sits beside the
// row rather than inside it.
type decisionRecord struct {
	decision   records.Decision
	projectKey string
	changeName string
}

// passRecord and mutationRecord are fakeStore's in-memory stand-ins for a
// panel_passes and a panel_mutations row (KAN-331). See dispatchRecord's
// doc comment for why the owning identity sits beside the row rather than
// inside it.
type passRecord struct {
	pass       records.Pass
	projectKey string
	changeName string
}

type mutationRecord struct {
	mutation   records.Mutation
	projectKey string
	changeName string
}

// RecordPass mirrors store.Store.RecordPass: append-only, every call a new
// row, and an unknown (projectKey, change) pair is store.ErrChangeNotFound,
// the condition the handler must answer 404 to.
func (f *fakeStore) RecordPass(_ context.Context, projectKey, change string, in records.Pass) (records.Pass, error) {
	f.recordCalls++
	if f.recordPassErr != nil {
		return records.Pass{}, f.recordPassErr
	}
	if _, ok := f.changes[changeKey(projectKey, change)]; !ok {
		return records.Pass{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, change)
	}
	f.nextPassID++
	out := in
	out.ID = f.nextPassID
	f.passes = append(f.passes, passRecord{pass: out, projectKey: projectKey, changeName: change})
	return out, nil
}

// RecordMutation is RecordPass's mutation-proof counterpart: append-only,
// every call a new row, same ErrChangeNotFound shape.
func (f *fakeStore) RecordMutation(_ context.Context, projectKey, change string, in records.Mutation) (records.Mutation, error) {
	f.recordCalls++
	if f.recordMutationErr != nil {
		return records.Mutation{}, f.recordMutationErr
	}
	if _, ok := f.changes[changeKey(projectKey, change)]; !ok {
		return records.Mutation{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, change)
	}
	f.nextMutationID++
	out := in
	out.ID = f.nextMutationID
	f.mutations = append(f.mutations, mutationRecord{mutation: out, projectKey: projectKey, changeName: change})
	return out, nil
}

// RecordDecision mirrors store.Store.RecordDecision: a write under a
// session token this change already holds a decision for replaces it
// (created=false); any other write inserts a new row (created=true). An
// unknown (projectKey, change) pair is store.ErrChangeNotFound, the
// condition the handler must answer 404 to.
func (f *fakeStore) RecordDecision(_ context.Context, projectKey, change string, in records.Decision) (records.Decision, bool, error) {
	f.recordCalls++
	if f.recordDecisionErr != nil {
		return records.Decision{}, false, f.recordDecisionErr
	}
	if _, ok := f.changes[changeKey(projectKey, change)]; !ok {
		return records.Decision{}, false, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, change)
	}

	for i := range f.decisions {
		d := &f.decisions[i]
		if d.projectKey == projectKey && d.changeName == change && d.decision.SessionToken == in.SessionToken {
			d.decision.Decision = in.Decision
			return d.decision, false, nil
		}
	}
	f.nextDecisionID++
	out := in
	out.ID = f.nextDecisionID
	f.decisions = append(f.decisions, decisionRecord{decision: out, projectKey: projectKey, changeName: change})
	return out, true, nil
}

// ListDecisions mirrors store.Store.ListDecisions' ordering: newest first.
func (f *fakeStore) ListDecisions(_ context.Context, projectKey, change string) ([]records.Decision, error) {
	f.recordCalls++
	if f.listDecisionsErr != nil {
		return nil, f.listDecisionsErr
	}
	var out []records.Decision
	for i := len(f.decisions) - 1; i >= 0; i-- {
		d := f.decisions[i]
		if d.projectKey == projectKey && d.changeName == change {
			out = append(out, d.decision)
		}
	}
	return out, nil
}

// --- test helpers ---

// recordTestServer returns a server backed by a fake that already knows
// about project/change, which every record route resolves against.
func recordTestServer(t *testing.T, project, change string) (*httptest.Server, *fakeStore) {
	t.Helper()
	fs := newFakeStore()
	fs.changes[changeKey(project, change)] = store.Change{ProjectKey: project, Name: change, State: store.StateInProgress}
	return newTestServer(t, fs), fs
}

func recordsPath(project, change string) string {
	return "/api/v1/records/" + project + "/" + change
}

// dispatchBody is the wire body a dispatch POST carries: every field the
// dispatcher knows at close, with seq deliberately absent -- the store
// allocates it.
func dispatchBody(role, model string) map[string]any {
	return map[string]any{
		"taskId":       "3",
		"role":         role,
		"model":        model,
		"key":          "task-3-" + role,
		"commitSha":    "abc1234",
		"outcome":      "completed",
		"sessionToken": "mf-record-api",
		"startedAt":    time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
}

// dispatchEndBody is the wire body a dispatch-end POST carries: the two
// fields that name the row, and the three that close it.
func dispatchEndBody(key string) map[string]any {
	return map[string]any{
		"sessionToken": "mf-record-api",
		"key":          key,
		"commitSha":    "def5678",
		"outcome":      "completed",
		"endedAt":      time.Date(2026, 8, 22, 9, 30, 0, 0, time.UTC).Format(time.RFC3339),
	}
}

// findingBody is the wire body a finding POST carries.
func findingBody(ref string, round int, status string) map[string]any {
	return map[string]any{
		"ref":      ref,
		"round":    round,
		"slot":     "principles",
		"severity": "major",
		"note":     "the handler swallows the decode error",
		"status":   status,
	}
}

// decisionBody is the wire body a decision POST carries: a session token
// and the whole `## Decision` block as opaque JSON.
func decisionBody(sessionToken, decisionJSON string) map[string]any {
	return map[string]any{
		"sessionToken": sessionToken,
		"decision":     json.RawMessage(decisionJSON),
	}
}

// patchJSON sends body as a PATCH and returns the status and raw body --
// the PATCH counterpart to postJSON above, which net/http has no
// one-call helper for.
func patchJSON(t *testing.T, url string, body any) (int, []byte) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	resp, respBody := patchRaw(t, url, string(raw))
	return resp.StatusCode, respBody
}

// patchRaw sends body verbatim as a PATCH, so a test can send something
// json.Marshal would never produce.
func patchRaw(t *testing.T, url, body string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPatch, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("build PATCH %s: %v", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH %s: %v", url, err)
	}
	defer resp.Body.Close()
	return resp, readBody(t, resp)
}

// postRaw sends body verbatim, so a test can send a body that is not valid
// JSON at all -- which json.Marshal could not produce.
func postRaw(t *testing.T, url, body string) (int, []byte) {
	t.Helper()
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode, readBody(t, resp)
}

// getRaw returns the whole response, headers included -- doGet returns
// only the status and body, and the daemon-header case needs the headers.
func getRaw(t *testing.T, url string) (*http.Response, []byte) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	return resp, readBody(t, resp)
}

// --- dispatches ---

// TestRecordDispatchRouteAllocatesSeqAndAnswers201 pins that a recorded
// dispatch is created (201, not 200 -- a row that did not exist before now
// does) and that the allocated seq comes back on the response, which is
// the only way a caller learns where in the change's record its dispatch
// landed.
func TestRecordDispatchRouteAllocatesSeqAndAnswers201(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")

	resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", dispatchBody("implementer", "opus"))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST dispatches = %d (%s), want 201", resp.StatusCode, body)
	}

	var got records.Dispatch
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if got.Seq != 1 {
		t.Errorf("seq = %d, want 1 (the store allocates it and the response reports it)", got.Seq)
	}
	if got.ID == 0 {
		t.Errorf("id = 0, want the stored row's own id so a later metrics merge can name it")
	}
	if got.Role != "implementer" || got.Model != "opus" || got.TaskID != "3" || got.CommitSHA != "abc1234" {
		t.Errorf("response = %+v, want the recorded intent round-tripped", got)
	}

	second, secondBody := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", dispatchBody("reviewer", "sonnet"))
	if second.StatusCode != http.StatusCreated {
		t.Fatalf("second POST dispatches = %d (%s), want 201", second.StatusCode, secondBody)
	}
	var next records.Dispatch
	if err := json.Unmarshal(secondBody, &next); err != nil {
		t.Fatalf("decode second response body %s: %v", secondBody, err)
	}
	if next.Seq != 2 {
		t.Errorf("second seq = %d, want 2", next.Seq)
	}
	if len(fs.dispatches) != 2 {
		t.Errorf("store holds %d dispatches, want 2", len(fs.dispatches))
	}
}

// --- findings ---

// TestRecordFindingRouteAnswers201OnCreateAnd200OnUpdate pins the one
// thing an upsert's status code has to say: whether the write inserted a
// row or replaced one. A fix round restating F1 must be distinguishable
// from the round that first raised it, and the change must still hold
// exactly one row for F1 either way.
func TestRecordFindingRouteAnswers201OnCreateAnd200OnUpdate(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + recordsPath("proj", "kan-1") + "/findings"

	resp, body := postJSON(t, url, findingBody("F1", 0, "open"))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first POST findings = %d (%s), want 201", resp.StatusCode, body)
	}

	restated := findingBody("F1", 1, "open")
	restated["note"] = "restated after the fix round re-read the handler"
	second, secondBody := postJSON(t, url, restated)
	if second.StatusCode != http.StatusOK {
		t.Fatalf("second POST findings for the same ref = %d (%s), want 200", second.StatusCode, secondBody)
	}

	var got records.Finding
	if err := json.Unmarshal(secondBody, &got); err != nil {
		t.Fatalf("decode response body %s: %v", secondBody, err)
	}
	if got.Note != restated["note"] || got.Round != 1 {
		t.Errorf("response = %+v, want the second write's own values", got)
	}
	if len(fs.findings) != 1 {
		t.Errorf("store holds %d findings, want exactly 1 -- the second write appended instead of updating", len(fs.findings))
	}
}

// TestSetFindingStatusRouteAnswers404ForAnUnknownRef pins that
// store.ErrFindingNotFound is mapped, not left to fall through to
// mapStoreError's generic 500. A caller that mistyped a ref must be told
// the ref is unknown, not that the daemon is broken -- and internal/client
// classifies a 500 as the store being unavailable, which would send a
// record write to the journal for a mistake no replay can ever fix.
func TestSetFindingStatusRouteAnswers404ForAnUnknownRef(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	if resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/findings", findingBody("F1", 0, "open")); resp.StatusCode != http.StatusCreated {
		t.Fatalf("seed POST findings = %d (%s), want 201", resp.StatusCode, body)
	}

	status, body := patchJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/findings/F1", map[string]any{"status": "fixed"})
	if status != http.StatusNoContent {
		t.Fatalf("PATCH a known ref = %d (%s), want 204", status, body)
	}

	status, body = patchJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/findings/F9", map[string]any{"status": "fixed"})
	if status != http.StatusNotFound {
		t.Errorf("PATCH an unknown ref = %d (%s), want 404", status, body)
	}
}

// --- the whole record ---

// TestRunRecordRouteReturnsDispatchesAndFindingsInOrder pins the order the
// renderer reads a record in. F10 is the case that matters: a record
// ordered lexically returns F1, F10, F2, so a ten-finding panel renders
// out of order.
func TestRunRecordRouteReturnsDispatchesAndFindingsInOrder(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	for _, role := range []string{"implementer", "reviewer", "panel-fix"} {
		if resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", dispatchBody(role, "opus")); resp.StatusCode != http.StatusCreated {
			t.Fatalf("seed POST dispatches %s = %d (%s), want 201", role, resp.StatusCode, body)
		}
	}
	for _, ref := range []string{"F10", "F2", "F1"} {
		if resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/findings", findingBody(ref, 0, "open")); resp.StatusCode != http.StatusCreated {
			t.Fatalf("seed POST findings %s = %d (%s), want 201", ref, resp.StatusCode, body)
		}
	}

	code, body := doGet(t, ts, recordsPath("proj", "kan-1"))
	if code != http.StatusOK {
		t.Fatalf("GET record = %d (%s), want 200", code, body)
	}

	var got records.Run
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if got.Change != "kan-1" {
		t.Errorf("change = %q, want kan-1", got.Change)
	}

	var seqs []int
	for _, d := range got.Dispatches {
		seqs = append(seqs, d.Seq)
	}
	if fmt.Sprint(seqs) != fmt.Sprint([]int{1, 2, 3}) {
		t.Errorf("dispatch seqs = %v, want [1 2 3]", seqs)
	}

	var refs []string
	for _, f := range got.Findings {
		refs = append(refs, f.Ref)
	}
	if fmt.Sprint(refs) != fmt.Sprint([]string{"F1", "F2", "F10"}) {
		t.Errorf("finding refs = %v, want [F1 F2 F10]", refs)
	}
}

// TestCostStatusRouteDerivesFromTheRealRunRecord drives the real
// costStatus handler (never fakeStore's own bookkeeping) through the HTTP
// route, over dispatches seeded with the wire shapes records.CostStatusOf
// distinguishes: one carrying tokens only, one carrying an unattributed
// stamp only, and one carrying both -- the tokens-outrank-a-stamp
// contradiction case. It pins the route (GET .../cost-status), the status
// (200) and the body shape (records.CostStatus's own JSON tags), so a
// change to any of the three fails this test rather than going unnoticed.
func TestCostStatusRouteDerivesFromTheRealRunRecord(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	seed := []struct {
		key     string
		metrics string
	}{
		{"tokens-only", `{"tokens":{"main":{"input":100,"output":20,"cache_read":0,"cache_creation":0},"sidechain":{"input":0,"output":0,"cache_read":0,"cache_creation":0}}}`},
		{"unattributed-only", `{"unattributed":{"reason":"session never bound"}}`},
		{"both", `{"tokens":{"main":{"input":5,"output":1,"cache_read":0,"cache_creation":0},"sidechain":{"input":0,"output":0,"cache_read":0,"cache_creation":0}},"unattributed":{"reason":"matched more than one dispatch","candidates":2}}`},
	}
	for _, s := range seed {
		body := dispatchBody("implementer", "opus")
		body["key"] = s.key
		body["metrics"] = json.RawMessage(s.metrics)
		resp, respBody := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", body)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("seed POST dispatches %s = %d (%s), want 201", s.key, resp.StatusCode, respBody)
		}
	}

	code, body := doGet(t, ts, recordsPath("proj", "kan-1")+"/cost-status")
	if code != http.StatusOK {
		t.Fatalf("GET cost-status = %d (%s), want 200", code, body)
	}

	var got records.CostStatus
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if got.Unattributed != 1 {
		t.Errorf("Unattributed = %d, want 1 -- only unattributed-only counts, tokens-only and both do not", got.Unattributed)
	}
	if got.Reasons["session never bound"] != 1 {
		t.Errorf("Reasons[session never bound] = %d, want 1", got.Reasons["session never bound"])
	}
	if len(got.Reasons) != 1 {
		t.Errorf("Reasons = %v, want exactly one reason", got.Reasons)
	}
}

// --- rejections ---

// TestRecordRouteRejectsAMalformedBodyWithoutReachingTheStore pins that a
// decode failure is a 400 answered before the store is touched at all: a
// partial write from a body that could not be understood would leave a
// record nothing can correct, since a dispatch row has no update path.
func TestRecordRouteRejectsAMalformedBodyWithoutReachingTheStore(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")

	for _, tc := range []struct {
		name string
		path string
		body string
	}{
		{"dispatch, not JSON at all", "/dispatches", `{"role":`},
		{"dispatch, an unknown field", "/dispatches", `{"role":"implementer","model":"opus","startedAt":"2026-08-22T09:00:00Z","spent":3}`},
		{"finding, not JSON at all", "/findings", `{"ref":`},
		{"finding, an unknown field", "/findings", `{"ref":"F1","round":0,"slot":"principles","severity":"major","note":"n","status":"open","blame":"me"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, body := postRaw(t, ts.URL+recordsPath("proj", "kan-1")+tc.path, tc.body)
			if status != http.StatusBadRequest {
				t.Errorf("POST %s = %d (%s), want 400", tc.path, status, body)
			}
		})
	}

	if fs.recordCalls != 0 {
		t.Errorf("the store was called %d times, want 0 -- a body that could not be decoded must never reach it", fs.recordCalls)
	}
}

// TestRecordRoutesCarryTheDaemonHeader pins that every record route
// answers with the header internal/client's do() reads. A route that
// omitted it would be classified as "not the daemon" and take the journal
// fallback on every single call, silently, no matter what it answered.
func TestRecordRoutesCarryTheDaemonHeader(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")
	base := ts.URL + recordsPath("proj", "kan-1")

	dispatchResp, dispatchBody0 := postJSON(t, base+"/dispatches", dispatchBody("implementer", "opus"))
	assertDaemonHeader(t, "POST /dispatches", dispatchResp, http.StatusCreated, dispatchBody0)

	findingResp, findingBody0 := postJSON(t, base+"/findings", findingBody("F1", 0, "open"))
	assertDaemonHeader(t, "POST /findings", findingResp, http.StatusCreated, findingBody0)

	patchResp, patchBody := patchRaw(t, base+"/findings/F1", `{"status":"fixed"}`)
	assertDaemonHeader(t, "PATCH /findings/{ref}", patchResp, http.StatusNoContent, patchBody)

	getResp, getBody := getRaw(t, base)
	assertDaemonHeader(t, "GET the record", getResp, http.StatusOK, getBody)
}

func assertDaemonHeader(t *testing.T, route string, resp *http.Response, wantStatus int, body []byte) {
	t.Helper()
	if resp.StatusCode != wantStatus {
		t.Fatalf("%s = %d (%s), want %d", route, resp.StatusCode, body, wantStatus)
	}
	if got := resp.Header.Get(api.DaemonHeader); got != api.DaemonHeaderValue {
		t.Errorf("%s: %s header = %q, want %q", route, api.DaemonHeader, got, api.DaemonHeaderValue)
	}
}

// TestEndDispatchRouteClosesTheRowItsBeginOpened pins the closing half of
// the dispatch pair: it answers 200 (the row already existed), it writes
// the three facts knowable only at close, and it finds its row by the
// session token and key its begin carried rather than by a seq the caller
// may never have seen.
func TestEndDispatchRouteClosesTheRowItsBeginOpened(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", dispatchBody("implementer", "opus"))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST dispatches = %d (%s), want 201", resp.StatusCode, body)
	}
	var opened records.Dispatch
	if err := json.Unmarshal(body, &opened); err != nil {
		t.Fatalf("decode begin response %s: %v", body, err)
	}

	resp, body = postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/end", dispatchEndBody("task-3-implementer"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST dispatches/end = %d (%s), want 200 -- the row already exists, so 201 would say something untrue", resp.StatusCode, body)
	}
	var closed records.Dispatch
	if err := json.Unmarshal(body, &closed); err != nil {
		t.Fatalf("decode end response %s: %v", body, err)
	}
	if closed.ID != opened.ID || closed.Seq != opened.Seq {
		t.Errorf("end returned dispatch %d/seq %d, want the row begin opened (%d/seq %d)", closed.ID, closed.Seq, opened.ID, opened.Seq)
	}
	if closed.EndedAt == nil {
		t.Fatalf("endedAt is absent from the closed row -- an unclosed window claims later usage forever")
	}
	if closed.CommitSHA != "def5678" || closed.Outcome != "completed" {
		t.Errorf("closed row = commit %q outcome %q, want def5678/completed", closed.CommitSHA, closed.Outcome)
	}
}

// TestDispatchEndAgentIDReachesTheStore pins the store-reaching half of
// the delta spec's "A dispatch's identifier may be recorded when it
// becomes known": begin need not carry the identifier at all -- the
// harness reports it only once the dispatch has launched -- so end must be
// able to carry it, and it must land on the very row begin opened.
func TestDispatchEndAgentIDReachesTheStore(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", dispatchBody("implementer", "opus"))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST dispatches = %d (%s), want 201", resp.StatusCode, body)
	}

	endBody := dispatchEndBody("task-3-implementer")
	endBody["agentId"] = "agent-example0001"
	resp, body = postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/end", endBody)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST dispatches/end = %d (%s), want 200", resp.StatusCode, body)
	}
	var closed records.Dispatch
	if err := json.Unmarshal(body, &closed); err != nil {
		t.Fatalf("decode end response %s: %v", body, err)
	}
	if closed.AgentID != "agent-example0001" {
		t.Errorf("closed.AgentID = %q, want agent-example0001 -- an identifier given only at close must still reach the row begin opened", closed.AgentID)
	}
}

// TestDispatchEndOmittedAgentIDPreservesBeginsIdentifier pins the
// load-bearing half of the same delta spec clause
// TestDispatchEndAgentIDReachesTheStore pins the other half of: an "end"
// call that OMITS agentId must never clear an identifier "begin" already
// recorded. fakeStore.EndDispatch already encodes the guard
// (`if in.AgentID != ""`), but nothing exercised the scenario it exists
// for before this test -- begin sets an identifier, end omits it, the
// identifier survives -- so a regression to an unconditional overwrite in
// either fakeStore or the real handler path would have passed every other
// test in this file.
func TestDispatchEndOmittedAgentIDPreservesBeginsIdentifier(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	beginBody := dispatchBody("implementer", "opus")
	beginBody["agentId"] = "agent-begin00001"
	resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", beginBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST dispatches = %d (%s), want 201", resp.StatusCode, body)
	}

	// dispatchEndBody carries no agentId key at all -- the omitted case.
	resp, body = postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/end", dispatchEndBody("task-3-implementer"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST dispatches/end = %d (%s), want 200", resp.StatusCode, body)
	}
	var closed records.Dispatch
	if err := json.Unmarshal(body, &closed); err != nil {
		t.Fatalf("decode end response %s: %v", body, err)
	}
	if closed.AgentID != "agent-begin00001" {
		t.Errorf("closed.AgentID = %q, want agent-begin00001 -- an end that omits agentId must never clear the identifier begin recorded", closed.AgentID)
	}
}

// TestEndDispatchRouteAnswers404ForAnUnknownKey pins the mapping that
// makes the CLI's own handling possible: a key naming no dispatch is a 404,
// never a 500. internal/client reads a 500 as "the store is unavailable",
// and this is the one record answer the CLI journals for a later replay --
// the begin it closes may still be queued ahead of it -- so a 500 here
// would be indistinguishable from a store that was never reached and the
// distinction would be lost.
func TestEndDispatchRouteAnswers404ForAnUnknownKey(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/end", dispatchEndBody("task-9-nobody"))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("POST dispatches/end for an unknown key = %d (%s), want 404", resp.StatusCode, body)
	}
}

// TestEndDispatchRouteRefusesABodyThatNamesNothing pins that a request
// missing the fields that identify or close the row is refused as the
// caller mistake it is -- a 400, judged before the store is touched, so a
// replay of it is retired rather than queued forever.
func TestEndDispatchRouteRefusesABodyThatNamesNothing(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")

	for _, tc := range []struct {
		name string
		body map[string]any
	}{
		{"no key", map[string]any{"sessionToken": "mf-record-api", "endedAt": time.Date(2026, 8, 22, 9, 30, 0, 0, time.UTC).Format(time.RFC3339)}},
		{"no session token", map[string]any{"key": "task-3-implementer", "endedAt": time.Date(2026, 8, 22, 9, 30, 0, 0, time.UTC).Format(time.RFC3339)}},
		{"no end instant", map[string]any{"sessionToken": "mf-record-api", "key": "task-3-implementer"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := fs.recordCalls
			resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/end", tc.body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("POST dispatches/end = %d (%s), want 400", resp.StatusCode, body)
			}
			if fs.recordCalls != before {
				t.Errorf("the store was reached for a body refused on its face")
			}
		})
	}
}

// TestRecordDispatchRouteIsIdempotentUnderOneKey pins that a begin
// delivered twice records one dispatch. A lost response and an unreachable
// store are indistinguishable to the CLI, so a replay carrying a row the
// store already holds is ordinary -- and a second row for one logical
// dispatch would be counted twice in every cost figure derived from it.
func TestRecordDispatchRouteIsIdempotentUnderOneKey(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")

	var first, replayed records.Dispatch
	for i, into := range []*records.Dispatch{&first, &replayed} {
		resp, body := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", dispatchBody("implementer", "opus"))
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("POST dispatches (attempt %d) = %d (%s), want 201", i+1, resp.StatusCode, body)
		}
		if err := json.Unmarshal(body, into); err != nil {
			t.Fatalf("decode response %s: %v", body, err)
		}
	}
	if replayed.ID != first.ID || replayed.Seq != first.Seq {
		t.Errorf("replay produced dispatch %d/seq %d, want the original %d/seq %d", replayed.ID, replayed.Seq, first.ID, first.Seq)
	}

	status, runBody := doGet(t, ts, recordsPath("proj", "kan-1"))
	if status != http.StatusOK {
		t.Fatalf("GET run record = %d (%s), want 200", status, runBody)
	}
	var run records.Run
	if err := json.Unmarshal([]byte(runBody), &run); err != nil {
		t.Fatalf("decode run record %s: %v", runBody, err)
	}
	if len(run.Dispatches) != 1 {
		t.Fatalf("the change holds %d dispatch rows, want 1", len(run.Dispatches))
	}
}

// --- guard verdicts and incidents (KAN-451) ---

// verdictBody is the wire body a verdict POST carries.
func verdictBody(guard, worktree, verdict string) map[string]any {
	return map[string]any{
		"guard":      guard,
		"worktree":   worktree,
		"verdict":    verdict,
		"recordedAt": time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
}

// incidentBody is the wire body an incident POST carries.
func incidentBody(guard, symptom, recovery string, minutesLost int) map[string]any {
	return map[string]any{
		"guard":       guard,
		"symptom":     symptom,
		"recovery":    recovery,
		"minutesLost": minutesLost,
	}
}

// TestRecordVerdictRouteAnswers201 pins that a recorded verdict is created
// (201) with the store's own row -- its allocated id included -- and that a
// body missing guard, worktree or verdict is refused before the store is
// touched at all.
func TestRecordVerdictRouteAnswers201(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + recordsPath("proj", "kan-1") + "/verdicts"

	resp, body := postJSON(t, url, verdictBody("check-unfinished-work", "/wt/kan-1", "CLEAR: /wt/kan-1 -- every plan item is checked and no finding is open"))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST verdicts = %d (%s), want 201", resp.StatusCode, body)
	}
	var got records.Verdict
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if got.ID == 0 {
		t.Errorf("id = 0, want the stored row's own id")
	}
	if got.Guard != "check-unfinished-work" || got.Worktree != "/wt/kan-1" || got.Verdict == "" {
		t.Errorf("response = %+v, want the recorded verdict round-tripped", got)
	}
	if len(fs.verdicts) != 1 {
		t.Errorf("store holds %d verdicts, want 1", len(fs.verdicts))
	}

	for _, field := range []string{"guard", "worktree", "verdict"} {
		t.Run("missing "+field, func(t *testing.T) {
			before := fs.recordCalls
			b := verdictBody("check-unfinished-work", "/wt/kan-1", "CLEAR: /wt/kan-1")
			delete(b, field)
			resp, respBody := postJSON(t, url, b)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("POST verdicts missing %s = %d (%s), want 400", field, resp.StatusCode, respBody)
			}
			if fs.recordCalls != before {
				t.Errorf("the store was reached for a body missing %s", field)
			}
		})
	}
}

// TestFlagVerdictRouteAnswers200And404 pins the false-positive flag's three
// outcomes: 200 with the flagged row for a real (change, guard) pair, 400
// for an empty reason (a caller mistake, judged before the store), and 404
// -- not 500 -- for a guard the change holds no verdict under, the mapping
// internal/client's classification depends on.
func TestFlagVerdictRouteAnswers200And404(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")
	verdictsURL := ts.URL + recordsPath("proj", "kan-1") + "/verdicts"
	flagURL := ts.URL + recordsPath("proj", "kan-1") + "/verdicts/false-positive"

	resp, body := postJSON(t, verdictsURL, verdictBody("check-unfinished-work", "/wt/kan-1", "OUTSTANDING: /wt/kan-1 -- 2 items unchecked"))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST verdicts = %d (%s), want 201", resp.StatusCode, body)
	}

	resp, body = postJSON(t, flagURL, map[string]any{"guard": "check-unfinished-work", "reason": "verified structural: 19/19 tasks ticked"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST verdicts/false-positive = %d (%s), want 200", resp.StatusCode, body)
	}
	var flagged records.Verdict
	if err := json.Unmarshal(body, &flagged); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if !flagged.FalsePositive || flagged.FalsePositiveReason != "verified structural: 19/19 tasks ticked" {
		t.Errorf("flagged verdict = %+v, want falsePositive=true with the operator's reason", flagged)
	}

	resp, body = postJSON(t, flagURL, map[string]any{"guard": "check-unfinished-work", "reason": ""})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST verdicts/false-positive with an empty reason = %d (%s), want 400", resp.StatusCode, body)
	}

	resp, body = postJSON(t, flagURL, map[string]any{"guard": "no-such-guard", "reason": "x"})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("POST verdicts/false-positive for a guard with no recorded verdict = %d (%s), want 404", resp.StatusCode, body)
	}
}

// TestListVerdictsRouteFiltersByQuery pins that the route parses guard and
// falsePositive from the query string and forwards exactly what it parsed
// to the store -- no query means ("", false); falsePositive=maybe is a
// caller mistake, refused as 400 before the store is ever touched.
func TestListVerdictsRouteFiltersByQuery(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")

	status, body := doGet(t, ts, "/api/v1/verdicts/proj")
	if status != http.StatusOK {
		t.Fatalf("GET verdicts = %d (%s), want 200", status, body)
	}
	if fs.lastListVerdictsGuard != "" || fs.lastListVerdictsFalsePositiveOnly {
		t.Errorf("no query: store called with (%q, %t), want (\"\", false)", fs.lastListVerdictsGuard, fs.lastListVerdictsFalsePositiveOnly)
	}

	status, body = doGet(t, ts, "/api/v1/verdicts/proj?guard=check-unfinished-work&falsePositive=true")
	if status != http.StatusOK {
		t.Fatalf("GET verdicts?guard=...&falsePositive=true = %d (%s), want 200", status, body)
	}
	if fs.lastListVerdictsGuard != "check-unfinished-work" || !fs.lastListVerdictsFalsePositiveOnly {
		t.Errorf("store called with (%q, %t), want (\"check-unfinished-work\", true)", fs.lastListVerdictsGuard, fs.lastListVerdictsFalsePositiveOnly)
	}

	before := fs.recordCalls
	status, body = doGet(t, ts, "/api/v1/verdicts/proj?falsePositive=maybe")
	if status != http.StatusBadRequest {
		t.Fatalf("GET verdicts?falsePositive=maybe = %d (%s), want 400", status, body)
	}
	if fs.recordCalls != before {
		t.Errorf("the store was reached for an unparsable falsePositive value")
	}
}

// TestRecordIncidentRouteAnswers201AndValidatesMinutes pins that a recorded
// incident is created (201) with its allocated id, and that a negative
// minutesLost or a missing guard/symptom/recovery is refused as a caller
// mistake before the store is touched at all.
func TestRecordIncidentRouteAnswers201AndValidatesMinutes(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + "/api/v1/incidents/proj"

	resp, body := postJSON(t, url, incidentBody("check-task-commit-fields", "Baseline revert re-entered mid-flight", "aborted revert, restored dir from stash", 55))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST incidents = %d (%s), want 201", resp.StatusCode, body)
	}
	var got records.Incident
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if got.ID == 0 || got.MinutesLost != 55 {
		t.Errorf("response = %+v, want an allocated id and minutesLost 55", got)
	}
	if len(fs.incidents) != 1 {
		t.Errorf("store holds %d incidents, want 1", len(fs.incidents))
	}

	before := fs.recordCalls
	resp, body = postJSON(t, url, incidentBody("check-task-commit-fields", "s", "r", -1))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST incidents with minutesLost=-1 = %d (%s), want 400", resp.StatusCode, body)
	}
	if fs.recordCalls != before {
		t.Errorf("the store was reached for a negative minutesLost")
	}

	for _, field := range []string{"guard", "symptom", "recovery"} {
		t.Run("missing "+field, func(t *testing.T) {
			before := fs.recordCalls
			b := incidentBody("g", "s", "r", 1)
			delete(b, field)
			resp, respBody := postJSON(t, url, b)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("POST incidents missing %s = %d (%s), want 400", field, resp.StatusCode, respBody)
			}
			if fs.recordCalls != before {
				t.Errorf("the store was reached for a body missing %s", field)
			}
		})
	}
}

// TestListIncidentsRouteReturnsNewestFirst pins that the route hands back
// exactly the array the store returned, in the newest-first order
// ListIncidents documents.
func TestListIncidentsRouteReturnsNewestFirst(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")
	incidentsURL := ts.URL + "/api/v1/incidents/proj"

	first, firstBody := postJSON(t, incidentsURL, incidentBody("check-task-commit-fields", "first", "r1", 10))
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first POST incidents = %d (%s), want 201", first.StatusCode, firstBody)
	}
	second, secondBody := postJSON(t, incidentsURL, incidentBody("check-unfinished-work", "second", "r2", 20))
	if second.StatusCode != http.StatusCreated {
		t.Fatalf("second POST incidents = %d (%s), want 201", second.StatusCode, secondBody)
	}

	status, body := doGet(t, ts, "/api/v1/incidents/proj")
	if status != http.StatusOK {
		t.Fatalf("GET incidents = %d (%s), want 200", status, body)
	}
	var got []records.Incident
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if len(got) != 2 {
		t.Fatalf("incidents = %d, want 2", len(got))
	}
	if got[0].Symptom != "second" || got[1].Symptom != "first" {
		t.Errorf("incidents = %+v, want newest (\"second\") first", got)
	}
}

// --- decisions ---

// TestRecordDecisionRouteCreatesThenReplaces pins that the route follows
// the same 201-then-200 shape RecordFinding's route does, except keyed by
// session token rather than ref: a second decision recorded under the
// same token replaces the first, since it is the same run restating its
// choice, never a second run.
func TestRecordDecisionRouteCreatesThenReplaces(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + recordsPath("proj", "kan-1") + "/decisions"

	resp, body := postJSON(t, url, decisionBody("mf-decide-1", `{"class":"small"}`))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first POST decisions = %d (%s), want 201", resp.StatusCode, body)
	}

	second, secondBody := postJSON(t, url, decisionBody("mf-decide-1", `{"class":"regular"}`))
	if second.StatusCode != http.StatusOK {
		t.Fatalf("second POST decisions for the same token = %d (%s), want 200", second.StatusCode, secondBody)
	}

	var got records.Decision
	if err := json.Unmarshal(secondBody, &got); err != nil {
		t.Fatalf("decode response body %s: %v", secondBody, err)
	}
	if string(got.Decision) != `{"class":"regular"}` {
		t.Errorf("response decision = %s, want the second write's own value", got.Decision)
	}
	if len(fs.decisions) != 1 {
		t.Errorf("store holds %d decisions, want exactly 1 -- the second write appended instead of replacing", len(fs.decisions))
	}
}

// TestListDecisionsRouteNewestFirst pins that the route hands back exactly
// the array the store returned, in the newest-first order ListDecisions
// documents -- what a resumed run reads to recover a stopped run's own
// choice.
func TestListDecisionsRouteNewestFirst(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + recordsPath("proj", "kan-1") + "/decisions"

	first, firstBody := postJSON(t, url, decisionBody("mf-decide-1", `{"class":"small"}`))
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first POST decisions = %d (%s), want 201", first.StatusCode, firstBody)
	}
	second, secondBody := postJSON(t, url, decisionBody("mf-decide-2", `{"class":"big"}`))
	if second.StatusCode != http.StatusCreated {
		t.Fatalf("second POST decisions = %d (%s), want 201", second.StatusCode, secondBody)
	}

	status, body := doGet(t, ts, recordsPath("proj", "kan-1")+"/decisions")
	if status != http.StatusOK {
		t.Fatalf("GET decisions = %d (%s), want 200", status, body)
	}
	var got []records.Decision
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode response body %s: %v", body, err)
	}
	if len(got) != 2 {
		t.Fatalf("decisions = %d, want 2", len(got))
	}
	if got[0].SessionToken != "mf-decide-2" || got[1].SessionToken != "mf-decide-1" {
		t.Errorf("decisions = %+v, want newest (mf-decide-2) first", got)
	}
}

// TestRecordDecisionRejectsEmptyBody pins that an empty session token and
// an empty decision body are both refused with 400 before the store is
// touched -- the same "caller mistake, not a store failure" contract
// ApplyFindingRecord's checks give its own route.
func TestRecordDecisionRejectsEmptyBody(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + recordsPath("proj", "kan-1") + "/decisions"

	missingDecision := decisionBody("mf-decide-1", `{"class":"small"}`)
	delete(missingDecision, "decision")

	for name, body := range map[string]map[string]any{
		"empty sessionToken": decisionBody("", `{"class":"small"}`),
		"empty decision":     missingDecision,
	} {
		t.Run(name, func(t *testing.T) {
			before := fs.recordCalls
			resp, respBody := postJSON(t, url, body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("POST decisions with %s = %d (%s), want 400", name, resp.StatusCode, respBody)
			}
			if fs.recordCalls != before {
				t.Errorf("the store was reached for a body with %s", name)
			}
		})
	}
}

// --- pass log (KAN-331) ---

// TestRecordPanelPassRoute pins that the pass-log write route appends a row
// per call -- never replaces one -- answers 201, and carries the change's
// passes back out on the run-record GET, which is what the renderer reads.
func TestRecordPanelPassRoute(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + recordsPath("proj", "kan-1") + "/passes"

	resp, body := postJSON(t, url, map[string]any{"round": 0, "note": "roster: compact — 60"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST passes = %d (%s), want 201", resp.StatusCode, body)
	}
	second, secondBody := postJSON(t, url, map[string]any{"round": 1, "note": "not re-run — nothing new since its last read"})
	if second.StatusCode != http.StatusCreated {
		t.Fatalf("second POST passes = %d (%s), want 201", second.StatusCode, secondBody)
	}
	if len(fs.passes) != 2 {
		t.Fatalf("store holds %d passes, want 2 -- the write replaced instead of appending", len(fs.passes))
	}

	status, runBody := doGet(t, ts, recordsPath("proj", "kan-1"))
	if status != http.StatusOK {
		t.Fatalf("GET run record = %d (%s), want 200", status, runBody)
	}
	var got records.Run
	if err := json.Unmarshal([]byte(runBody), &got); err != nil {
		t.Fatalf("decode run record %s: %v", runBody, err)
	}
	if len(got.Passes) != 2 {
		t.Fatalf("run record carries %d passes, want 2", len(got.Passes))
	}
	if got.Passes[0].Round != 0 || got.Passes[1].Round != 1 {
		t.Errorf("passes ordered %+v, want round 0 before round 1", got.Passes)
	}
}

// TestRecordPanelMutationRoute pins the mutation-proof write route the same
// way: append-only, 201, and visible on the run-record GET beside the
// passes.
func TestRecordPanelMutationRoute(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")
	url := ts.URL + recordsPath("proj", "kan-1") + "/mutations"

	in := map[string]any{"round": 1, "path": "stats/internal/records/render.go", "mutated": "flipped the guard", "test": "TestRenderPanelPassLog"}
	resp, body := postJSON(t, url, in)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST mutations = %d (%s), want 201", resp.StatusCode, body)
	}
	if len(fs.mutations) != 1 {
		t.Fatalf("store holds %d mutations, want 1", len(fs.mutations))
	}

	status, runBody := doGet(t, ts, recordsPath("proj", "kan-1"))
	if status != http.StatusOK {
		t.Fatalf("GET run record = %d (%s), want 200", status, runBody)
	}
	var got records.Run
	if err := json.Unmarshal([]byte(runBody), &got); err != nil {
		t.Fatalf("decode run record %s: %v", runBody, err)
	}
	if len(got.Mutations) != 1 || got.Mutations[0].Path != in["path"] || got.Mutations[0].Mutated != in["mutated"] || got.Mutations[0].Test != in["test"] {
		t.Fatalf("run record mutations = %+v, want the one recorded row", got.Mutations)
	}
}

// TestRecordPanelPassAndMutationRejectEmptyFields pins the 400 shape: an
// empty note, and any empty mutation field, are refused before the store is
// reached -- a pass-log line with a missing field records nothing, and the
// contract's fix-mutation line has exactly three fields.
func TestRecordPanelPassAndMutationRejectEmptyFields(t *testing.T) {
	ts, fs := recordTestServer(t, "proj", "kan-1")

	for _, tc := range []struct {
		name string
		path string
		body map[string]any
	}{
		{"empty note", "passes", map[string]any{"round": 0, "note": ""}},
		{"missing path", "mutations", map[string]any{"round": 1, "mutated": "x", "test": "y"}},
		{"missing mutated", "mutations", map[string]any{"round": 1, "path": "p", "test": "y"}},
		{"missing test", "mutations", map[string]any{"round": 1, "path": "p", "mutated": "none"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := fs.recordCalls
			resp, respBody := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/"+tc.path, tc.body)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("POST %s with %s = %d (%s), want 400", tc.path, tc.name, resp.StatusCode, respBody)
			}
			if fs.recordCalls != before {
				t.Errorf("the store was reached for a body with an empty %s", tc.name)
			}
		})
	}
}
