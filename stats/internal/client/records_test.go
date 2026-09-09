package client_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/records"
)

// minimalDispatch is the smallest dispatch a caller records: what ran, on
// what model, and when it started.
func minimalDispatch() records.Dispatch {
	return records.Dispatch{
		Role:      "implementer",
		Model:     "opus",
		StartedAt: time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC),
	}
}

// minimalFinding is the smallest finding a panel slot records.
func minimalFinding() records.Finding {
	return records.Finding{
		Ref:      "F1",
		Round:    0,
		Slot:     "principles",
		Severity: "major",
		Note:     "the handler swallows the decode error",
		Status:   "open",
	}
}

// TestRecordDispatchReturnsTheRecordedRow pins that a 201 is read as
// success and that the row the daemon allocated -- its id and its seq,
// neither of which the caller could know -- comes back to the caller
// rather than being discarded with the response body.
func TestRecordDispatchReturnsTheRecordedRow(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/records/proj/kan-1/dispatches" {
			t.Errorf("request = %s %s, want POST /api/v1/records/proj/kan-1/dispatches", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"seq":3,"role":"implementer","model":"opus","startedAt":"2026-08-22T09:00:00Z"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	got, err := c.RecordDispatch(context.Background(), "proj", "kan-1", minimalDispatch())
	if err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}
	if got.ID != 7 || got.Seq != 3 {
		t.Errorf("recorded dispatch = %+v, want id 7 and seq 3 as the daemon allocated them", got)
	}
}

// TestRecordFindingAcceptsBothCreatedAndUpdated pins that the daemon's two
// success codes are both success to this client, and that the client says
// which one it was. An upsert answers 201 when it inserted and 200 when it
// replaced, and a client that recognised only one of them would send every
// fix round's restated finding to the journal for a write that had already
// landed.
//
// The created result is why the split exists at all: `flow record
// finding` prints "recorded: F<n>" on an insert and "updated: F<n>" on a
// replace, so a panel run can tell a new finding from a restated one. A
// client that dropped the flag would make the store's own created result,
// the handler's two status codes and the tests for both dead weight.
func TestRecordFindingAcceptsBothCreatedAndUpdated(t *testing.T) {
	for _, tc := range []struct {
		status      int
		wantCreated bool
	}{
		{http.StatusCreated, true},
		{http.StatusOK, false},
	} {
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(`{"ref":"F1","round":1,"slot":"principles","severity":"major","note":"n","status":"open"}`))
		}))

		c := client.New(srv.URL, srv.Client())
		got, created, err := c.RecordFinding(context.Background(), "proj", "kan-1", minimalFinding())
		if err != nil {
			t.Errorf("RecordFinding against a %d: %v", tc.status, err)
		}
		if got.Round != 1 {
			t.Errorf("RecordFinding against a %d returned %+v, want the daemon's own stored row", tc.status, got)
		}
		if created != tc.wantCreated {
			t.Errorf("RecordFinding against a %d reported created = %t, want %t", tc.status, created, tc.wantCreated)
		}
		srv.Close()
	}
}

// TestSetFindingStatusDistinguishesAnUnknownRefFromAnUnreachableStore is
// the mapping the never-block guarantee rests on: a 404 means the daemon
// answered and the ref is wrong, which no replay can fix, while anything
// that is not the daemon answering is ErrUnavailable and belongs in the
// journal.
func TestSetFindingStatusDistinguishesAnUnknownRefFromAnUnreachableStore(t *testing.T) {
	ok := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/records/proj/kan-1/findings/F1" {
			t.Errorf("request = %s %s, want PATCH /api/v1/records/proj/kan-1/findings/F1", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ok.Close()
	if err := client.New(ok.URL, ok.Client()).SetFindingStatus(context.Background(), "proj", "kan-1", "F1", "fixed"); err != nil {
		t.Fatalf("SetFindingStatus against a 204: %v", err)
	}

	missing := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"store: finding not found: F9 in proj/kan-1"}`))
	}))
	defer missing.Close()
	err := client.New(missing.URL, missing.Client()).SetFindingStatus(context.Background(), "proj", "kan-1", "F9", "fixed")
	if !errors.Is(err, client.ErrNotFound) {
		t.Errorf("SetFindingStatus for an unknown ref = %v, want ErrNotFound", err)
	}

	rejected := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"status is required"}`))
	}))
	defer rejected.Close()
	err = client.New(rejected.URL, rejected.Client()).SetFindingStatus(context.Background(), "proj", "kan-1", "F1", "")
	if !errors.Is(err, client.ErrRecordRejected) {
		t.Errorf("SetFindingStatus refused by the daemon = %v, want ErrRecordRejected -- a caller mistake, never a journalled write", err)
	}
}

// TestGetRunRecordReadsTheWholeRecord pins that the record comes back
// decoded, and that a change the daemon has never heard of is ErrNotFound
// rather than an empty record -- the distinction a render reports as "no
// rows of this kind" versus "no such change".
func TestGetRunRecordReadsTheWholeRecord(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/records/proj/kan-1" {
			t.Errorf("request = %s %s, want GET /api/v1/records/proj/kan-1", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"change":"kan-1","dispatches":[{"id":1,"seq":1,"role":"implementer","model":"opus","startedAt":"2026-08-22T09:00:00Z"}],"findings":[{"ref":"F1","round":0,"slot":"principles","severity":"major","note":"n","status":"open"}]}`))
	}))
	defer srv.Close()

	got, err := client.New(srv.URL, srv.Client()).GetRunRecord(context.Background(), "proj", "kan-1")
	if err != nil {
		t.Fatalf("GetRunRecord: %v", err)
	}
	if got.Change != "kan-1" || len(got.Dispatches) != 1 || len(got.Findings) != 1 {
		t.Errorf("record = %+v, want one dispatch and one finding for kan-1", got)
	}

	missing := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"store: change not found: proj/kan-9"}`))
	}))
	defer missing.Close()
	if _, err := client.New(missing.URL, missing.Client()).GetRunRecord(context.Background(), "proj", "kan-9"); !errors.Is(err, client.ErrNotFound) {
		t.Errorf("GetRunRecord for an unknown change = %v, want ErrNotFound", err)
	}
}

// TestRecordCallsFallBackWhenNothingTrustworthyAnswers pins the whole
// never-block guarantee for the record verb, on every one of its four
// calls at once: a refused connection and a look-alike 200 from something
// that is not the daemon must both be ErrUnavailable, which is the error
// the CLI journals on. A route that classified either of them as anything
// else would either block the pipeline or silently drop the record.
func TestRecordCallsFallBackWhenNothingTrustworthyAnswers(t *testing.T) {
	lookAlike := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1,"seq":1,"change":"kan-1"}`))
	}))
	defer lookAlike.Close()

	calls := map[string]func(*client.Client) error{
		"RecordDispatch": func(c *client.Client) error {
			_, err := c.RecordDispatch(context.Background(), "proj", "kan-1", minimalDispatch())
			return err
		},
		"RecordFinding": func(c *client.Client) error {
			_, _, err := c.RecordFinding(context.Background(), "proj", "kan-1", minimalFinding())
			return err
		},
		"SetFindingStatus": func(c *client.Client) error {
			return c.SetFindingStatus(context.Background(), "proj", "kan-1", "F1", "fixed")
		},
		"GetRunRecord": func(c *client.Client) error {
			_, err := c.GetRunRecord(context.Background(), "proj", "kan-1")
			return err
		},
	}

	for name, call := range calls {
		t.Run(name+" on a dead port", func(t *testing.T) {
			err := call(client.New(deadPortURL(t), &http.Client{Timeout: time.Second}))
			if !errors.Is(err, client.ErrUnavailable) {
				t.Errorf("err = %v, want ErrUnavailable", err)
			}
		})
		t.Run(name+" against a response with no daemon header", func(t *testing.T) {
			if err := call(client.New(lookAlike.URL, lookAlike.Client())); !errors.Is(err, client.ErrUnavailable) {
				t.Errorf("err = %v, want ErrUnavailable (a look-alike answer must fall back)", err)
			}
		})
	}
}

// TestEndDispatchReadsTheClosedRowAndDistinguishesAnUnknownKey pins both
// halves of the closing call's contract.
//
// A 200 is success and the row comes back closed, so a caller can report
// which dispatch it just ended. A 404 is ErrNotFound and NOT ErrUnavailable
// -- the daemon answered -- which is the distinction the CLI's own handling
// rests on: this one 404 is journalled rather than reported, because the
// begin it closes may still be sitting in the same journal ahead of it, and
// a client that read the answer as a transport failure would be indistinguishable
// from a store that was never reached.
func TestEndDispatchReadsTheClosedRowAndDistinguishesAnUnknownKey(t *testing.T) {
	end := records.DispatchEnd{
		SessionToken: "mf-client-end",
		Key:          "task-6-implementer",
		CommitSHA:    "abc1234",
		Outcome:      "completed",
		EndedAt:      time.Date(2026, 8, 22, 9, 30, 0, 0, time.UTC),
	}

	t.Run("closed", func(t *testing.T) {
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != "/api/v1/records/proj/kan-1/dispatches/end" {
				t.Errorf("request = %s %s, want POST /api/v1/records/proj/kan-1/dispatches/end", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":7,"seq":3,"role":"implementer","model":"opus","key":"task-6-implementer","commitSha":"abc1234","outcome":"completed","startedAt":"2026-08-22T09:00:00Z","endedAt":"2026-08-22T09:30:00Z"}`))
		}))
		defer srv.Close()

		got, err := client.New(srv.URL, srv.Client()).EndDispatch(context.Background(), "proj", "kan-1", end)
		if err != nil {
			t.Fatalf("EndDispatch: %v", err)
		}
		if got.Seq != 3 || got.EndedAt == nil {
			t.Errorf("closed dispatch = %+v, want seq 3 and a non-nil endedAt", got)
		}
	})

	t.Run("unknown key", func(t *testing.T) {
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"no dispatch under that key"}`))
		}))
		defer srv.Close()

		_, err := client.New(srv.URL, srv.Client()).EndDispatch(context.Background(), "proj", "kan-1", end)
		if !errors.Is(err, client.ErrNotFound) {
			t.Fatalf("EndDispatch against a 404 = %v, want ErrNotFound", err)
		}
		if errors.Is(err, client.ErrUnavailable) {
			t.Error("a 404 was classified as an unreachable store -- the daemon answered, and the CLI's journalling decision depends on telling the two apart")
		}
	})
}

// minimalVerdict is the smallest verdict a guard records.
func minimalVerdict() records.Verdict {
	return records.Verdict{
		Guard:      "check-unfinished-work",
		Worktree:   "/wt/kan-1",
		Verdict:    "CLEAR: /wt/kan-1 -- every plan item is checked and no finding is open",
		RecordedAt: time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC),
	}
}

// TestRecordVerdictReturnsTheRecordedRow pins that a 201 is read as success
// and that the row the daemon allocated -- its id -- comes back to the
// caller rather than being discarded with the response body.
func TestRecordVerdictReturnsTheRecordedRow(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/records/proj/kan-1/verdicts" {
			t.Errorf("request = %s %s, want POST /api/v1/records/proj/kan-1/verdicts", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":9,"guard":"check-unfinished-work","worktree":"/wt/kan-1","verdict":"CLEAR: /wt/kan-1","recordedAt":"2026-09-05T09:00:00Z"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	got, err := c.RecordVerdict(context.Background(), "proj", "kan-1", minimalVerdict())
	if err != nil {
		t.Fatalf("RecordVerdict: %v", err)
	}
	if got.ID != 9 {
		t.Errorf("recorded verdict = %+v, want id 9 as the daemon allocated it", got)
	}
}

// TestFlagVerdictDistinguishesNoVerdictFromAnUnreachableStore pins the same
// mapping TestSetFindingStatusDistinguishesAnUnknownRefFromAnUnreachableStore
// pins for findings: a 404 means the daemon answered and no verdict exists
// for that (change, guard) pair, which no replay can fix, while anything
// that is not the daemon answering is ErrUnavailable and belongs in the
// journal.
func TestFlagVerdictDistinguishesNoVerdictFromAnUnreachableStore(t *testing.T) {
	ok := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/records/proj/kan-1/verdicts/false-positive" {
			t.Errorf("request = %s %s, want POST /api/v1/records/proj/kan-1/verdicts/false-positive", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":9,"guard":"check-unfinished-work","worktree":"/wt/kan-1","verdict":"OUTSTANDING","recordedAt":"2026-09-05T09:00:00Z","falsePositive":true,"falsePositiveReason":"verified structural"}`))
	}))
	defer ok.Close()
	got, err := client.New(ok.URL, ok.Client()).FlagVerdictFalsePositive(context.Background(), "proj", "kan-1", records.VerdictFlag{Guard: "check-unfinished-work", Reason: "verified structural"})
	if err != nil {
		t.Fatalf("FlagVerdictFalsePositive against a 200: %v", err)
	}
	if !got.FalsePositive || got.FalsePositiveReason != "verified structural" {
		t.Errorf("flagged verdict = %+v, want falsePositive=true with the reason", got)
	}

	missing := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"store: verdict for guard \"no-such-guard\" in proj/kan-1 not found"}`))
	}))
	defer missing.Close()
	_, err = client.New(missing.URL, missing.Client()).FlagVerdictFalsePositive(context.Background(), "proj", "kan-1", records.VerdictFlag{Guard: "no-such-guard", Reason: "x"})
	if !errors.Is(err, client.ErrNotFound) {
		t.Errorf("FlagVerdictFalsePositive for a guard with no recorded verdict = %v, want ErrNotFound", err)
	}
	if errors.Is(err, client.ErrUnavailable) {
		t.Error("a 404 was classified as an unreachable store -- a genuine \"no verdict\" must never be journalled for a replay that can never succeed")
	}
}

// TestListVerdictsAndIncidentsReadArrays pins that both list reads decode
// the JSON array the daemon answers with, and that a change the daemon has
// never heard of -- an unrelated failure a project-scoped list route cannot
// even ask about -- is not conflated with a transport failure: only the
// daemon-header check and the response's own status decide ErrUnavailable
// here, since neither route resolves a change at all.
func TestListVerdictsAndIncidentsReadArrays(t *testing.T) {
	verdicts := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/verdicts/proj" {
			t.Errorf("request = %s %s, want GET /api/v1/verdicts/proj", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("guard"); got != "check-unfinished-work" {
			t.Errorf("guard query = %q, want check-unfinished-work", got)
		}
		if got := r.URL.Query().Get("falsePositive"); got != "true" {
			t.Errorf("falsePositive query = %q, want true", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"guard":"check-unfinished-work","worktree":"/wt/kan-1","verdict":"OUTSTANDING","recordedAt":"2026-09-05T09:00:00Z","falsePositive":true}]`))
	}))
	defer verdicts.Close()

	gotVerdicts, err := client.New(verdicts.URL, verdicts.Client()).ListVerdicts(context.Background(), "proj", "check-unfinished-work", true)
	if err != nil {
		t.Fatalf("ListVerdicts: %v", err)
	}
	if len(gotVerdicts) != 1 || gotVerdicts[0].ID != 1 {
		t.Errorf("verdicts = %+v, want one row with id 1", gotVerdicts)
	}

	incidents := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/incidents/proj" {
			t.Errorf("request = %s %s, want GET /api/v1/incidents/proj", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer incidents.Close()

	gotIncidents, err := client.New(incidents.URL, incidents.Client()).ListIncidents(context.Background(), "proj")
	if err != nil {
		t.Fatalf("ListIncidents: %v", err)
	}
	if len(gotIncidents) != 0 {
		t.Errorf("incidents = %+v, want an empty slice for an empty array", gotIncidents)
	}
}
