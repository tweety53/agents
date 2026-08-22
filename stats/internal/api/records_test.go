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
func (f *fakeStore) SetFindingStatus(_ context.Context, projectKey, change, ref, status string) error {
	f.recordCalls++
	if f.setFindingStatusErr != nil {
		return f.setFindingStatusErr
	}
	for i := range f.findings {
		r := &f.findings[i]
		if r.projectKey == projectKey && r.changeName == change && r.finding.Ref == ref {
			r.finding.Status = status
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
	sort.Slice(out.Dispatches, func(i, j int) bool { return out.Dispatches[i].Seq < out.Dispatches[j].Seq })
	sort.Slice(out.Findings, func(i, j int) bool { return refLess(out.Findings[i].Ref, out.Findings[j].Ref) })
	return out, nil
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
