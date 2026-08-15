package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// genuineDaemon wraps handler so every response it writes carries the
// header a real myflowd sets on every response (internal/api's
// withDaemonHeader). Tests asserting client behaviour against a store that
// genuinely answered use this; tests asserting the F1 fallback behaviour
// against a look-alike server deliberately do not.
//
// This reads api.DaemonHeader / api.DaemonHeaderValue directly rather than
// a hand-copied literal -- this package can afford to import internal/api
// in its _test.go files (test code, never linked into the CLI binary), so
// there is no reason for this specific fake to carry its own copy of a
// value the real daemon already exports. internal/client's *production*
// code still keeps its own literal copy of the same two strings (see
// client.go's daemonHeaderName/daemonHeaderValue), per design.md's
// boundary that the CLI binary itself must never import the daemon's
// internal packages -- that copy is exactly what
// TestClientAgreesWithRealDaemonOverDaemonHeader below exists to keep
// honest, since nothing else in this suite would catch it drifting from
// api.DaemonHeaderValue.
func genuineDaemon(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(api.DaemonHeader, api.DaemonHeaderValue)
		handler(w, r)
	}
}

func deadPortURL(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	// Nothing listens on addr any more: connections to it refuse.
	return "http://" + addr
}

func TestGetChangeSucceedsOn200(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"state":"IN_PROGRESS"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	body, err := c.GetChange(context.Background(), "proj", "name")
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if string(body) != `{"state":"IN_PROGRESS"}` {
		t.Errorf("body = %s", body)
	}
}

func TestGetChangeReportsNotFoundDistinctly(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.GetChange(context.Background(), "proj", "name")
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if errors.Is(err, client.ErrUnavailable) {
		t.Errorf("a legitimate 404 must not also read as ErrUnavailable")
	}
}

func TestGetChangeFallsBackOnDeadPort(t *testing.T) {
	c := client.New(deadPortURL(t), &http.Client{Timeout: time.Second})
	_, err := c.GetChange(context.Background(), "proj", "name")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestGetChangeFallsBackOnMalformedBody(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.GetChange(context.Background(), "proj", "name")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestGetChangeFallsBackOnServerError(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.GetChange(context.Background(), "proj", "name")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

// TestGetChangeFallsBackWhenNotFoundLacksDaemonHeader is F1's
// reproduction: a bare foreign server -- no myflowd behind it at all --
// answering 404 on this path (a generic API gateway's default, a stale
// service, a test fixture like Python's http.server) must fall back, not
// be trusted as "the store correctly reports no such record". Only a
// response carrying the daemon header may be treated as a store answer.
func TestGetChangeFallsBackWhenNotFoundLacksDaemonHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`Not Found`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.GetChange(context.Background(), "proj", "name")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable (a look-alike 404 must fall back)", err)
	}
	if errors.Is(err, client.ErrNotFound) {
		t.Errorf("a 404 lacking the daemon header must not be trusted as ErrNotFound")
	}
}

// TestGetChangeFallsBackWhenSuccessLacksDaemonHeader closes the same gap
// on the success path: a foreign 200 must not be trusted as a store
// answer either, since a client that trusted it would report a change as
// found (or, on PutChange, as written) when the real store was never
// reached at all.
func TestGetChangeFallsBackWhenSuccessLacksDaemonHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"state":"IN_PROGRESS"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.GetChange(context.Background(), "proj", "name")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable (a look-alike 200 must fall back)", err)
	}
}

func TestPutChangeSucceedsOn200(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"state":"IN_PROGRESS"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	err := c.PutChange(context.Background(), "proj", "name", []byte(`{"state":"IN_PROGRESS"}`))
	if err != nil {
		t.Fatalf("PutChange: %v", err)
	}
}

func TestPutChangeReportsRefusalDistinctly(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"refused"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	err := c.PutChange(context.Background(), "proj", "name", []byte(`{}`))
	if !errors.Is(err, client.ErrRefused) {
		t.Fatalf("err = %v, want ErrRefused", err)
	}
	if errors.Is(err, client.ErrUnavailable) {
		t.Errorf("a 409 refusal must not also read as ErrUnavailable -- it is not a fallback trigger")
	}
}

// TestPutChangeFallsBackWhenRefusalLacksDaemonHeader is F1's core
// reproduction: pointing PutChange at a bare foreign server answering 409
// on PUT must fall back, not be reported as a genuine monotonic refusal --
// this exact confusion previously made the CLI exit 1 (blocking) against
// any process squatting the configured port, with no database, no
// myflowd, and no store involved at all.
func TestPutChangeFallsBackWhenRefusalLacksDaemonHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`Conflict`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	err := c.PutChange(context.Background(), "proj", "name", []byte(`{}`))
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable (a look-alike 409 must fall back)", err)
	}
	if errors.Is(err, client.ErrRefused) {
		t.Errorf("a 409 lacking the daemon header must not be trusted as ErrRefused")
	}
}

func TestPutChangeFallsBackOnDeadPort(t *testing.T) {
	c := client.New(deadPortURL(t), &http.Client{Timeout: time.Second})
	err := c.PutChange(context.Background(), "proj", "name", []byte(`{}`))
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestPutChangeFallsBackOnTimeout(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := c.PutChange(ctx, "proj", "name", []byte(`{}`))
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestPutChangeFallsBackOnNon2xxThatIsNotRefusal(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"malformed"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	err := c.PutChange(context.Background(), "proj", "name", []byte(`{}`))
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if errors.Is(err, client.ErrRefused) {
		t.Errorf("a 400 is not the monotonic refusal and must not read as ErrRefused")
	}
}

func TestPutChangeFallsBackOnMalformedResponseBody(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json at all`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	err := c.PutChange(context.Background(), "proj", "name", []byte(`{}`))
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

// --- ListStateBoard: F1's own fix. cmd/myflow's `state list` (and, through
// it, skills/myflow-status/SKILL.md's enumeration) is required to go
// through this method rather than a hand-written curl call, specifically
// so it inherits the header check, the ErrUnavailable classification and
// the request timeout every other client method already has. These tests
// are the proof that inheritance actually holds. ---

func TestListStateBoardSucceedsOn200(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("project"); got != "proj" {
			t.Errorf("project query param = %q, want %q", got, "proj")
		}
		if r.URL.Query().Get("from") == "" || r.URL.Query().Get("to") == "" {
			t.Errorf("from/to must both be set: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"view":"state-board","rows":[
			{"projectKey":"proj","name":"kan-1","state":"STARTED","updatedAt":"2026-08-13T10:00:00Z","updatedBy":"/myflow-start","nextCommand":"/myflow-do"},
			{"projectKey":"proj","name":"kan-2","state":"IN_PROGRESS","updatedAt":"2026-08-13T11:00:00Z","updatedBy":"/myflow-do","nextCommand":"/myflow-finish"}
		]}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	rows, err := c.ListStateBoard(context.Background(), "proj")
	if err != nil {
		t.Fatalf("ListStateBoard: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2: %+v", len(rows), rows)
	}
	if rows[0].Name != "kan-1" || rows[0].State != "STARTED" {
		t.Errorf("rows[0] = %+v", rows[0])
	}
	if rows[1].Name != "kan-2" || rows[1].State != "IN_PROGRESS" {
		t.Errorf("rows[1] = %+v", rows[1])
	}
}

func TestListStateBoardEmptyBoardReturnsEmptyNotError(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"view":"state-board","rows":[]}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	rows, err := c.ListStateBoard(context.Background(), "proj")
	if err != nil {
		t.Fatalf("ListStateBoard: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("rows = %+v, want empty", rows)
	}
}

func TestListStateBoardFallsBackOnDeadPort(t *testing.T) {
	c := client.New(deadPortURL(t), &http.Client{Timeout: time.Second})
	_, err := c.ListStateBoard(context.Background(), "proj")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestListStateBoardFallsBackOnServerError(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.ListStateBoard(context.Background(), "proj")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestListStateBoardFallsBackOnMalformedBody(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json at all`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.ListStateBoard(context.Background(), "proj")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

// TestListStateBoardFallsBackWhenSuccessLacksDaemonHeader is F1's own
// mutation check on this method: a bare look-alike server answering 200
// with a well-formed board body, but never setting Myflow-Daemon, must
// still be treated as unreachable. Mutating the header check in
// ListStateBoard (deleting the `if !fromDaemon` branch, or comparing
// against the wrong constant) is exactly what this test would catch --
// the response body alone is indistinguishable from a genuine one, so the
// header is the only thing standing between a caller and reporting a
// forged, always-empty-or-stale "complete" list as the real one.
func TestListStateBoardFallsBackWhenSuccessLacksDaemonHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"view":"state-board","rows":[{"name":"kan-1","state":"STARTED"}]}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.ListStateBoard(context.Background(), "proj")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable (a look-alike 200 must fall back)", err)
	}
}

// --- BeginStage / EndStage: the same per-failure-mode coverage PutChange
// has above, plus the two distinct 400 cases a stage mark's rejection can
// carry. Before this task's fix, BeginStage/EndStage mapped *every* 400 to
// ErrUndocumentedStage regardless of the daemon's actual reason -- these
// tests are what would have failed against that bug. ---

func minimalBeginReq() client.BeginStageRequest {
	return client.BeginStageRequest{
		ProjectKey:   "proj",
		ChangeName:   "chg",
		Harness:      "claude-code",
		SessionToken: "mf-session-token-minimal",
		Command:      "/myflow-do",
		Stage:        "SDD + TDD per task",
		StartedAt:    time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC),
	}
}

func minimalEndReq() client.EndStageRequest {
	return client.EndStageRequest{
		ProjectKey: "proj",
		ChangeName: "chg",
		Command:    "/myflow-do",
		Stage:      "SDD + TDD per task",
		EndedAt:    time.Date(2026, 8, 13, 10, 5, 0, 0, time.UTC),
		Outcome:    "completed",
	}
}

func TestBeginStageSucceedsOn200(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":42,"attempt":1}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	result, err := c.BeginStage(context.Background(), minimalBeginReq())
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if result.StageRunID != 42 || result.Attempt != 1 {
		t.Errorf("result = %+v, want {StageRunID:42 Attempt:1}", result)
	}
}

func TestBeginStageFallsBackOnDeadPort(t *testing.T) {
	c := client.New(deadPortURL(t), &http.Client{Timeout: time.Second})
	_, err := c.BeginStage(context.Background(), minimalBeginReq())
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestBeginStageFallsBackOnTimeout(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := c.BeginStage(ctx, minimalBeginReq())
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestBeginStageFallsBackOnNon2xxThatIsNotAKnownRefusal(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.BeginStage(context.Background(), minimalBeginReq())
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestBeginStageFallsBackWhenResponseLacksDaemonHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.BeginStage(context.Background(), minimalBeginReq())
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable (a look-alike 200 must fall back)", err)
	}
}

func TestBeginStageFallsBackOnMalformedResponseBody(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json at all`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.BeginStage(context.Background(), minimalBeginReq())
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

// TestBeginStageReportsUndocumentedStageDistinctly pins the 400 case a
// daemon actually raises for internal/stages.Validate's own refusal: the
// response body's Code field is "undocumented_stage", which must map to
// ErrUndocumentedStage and *not* the generic ErrStageMarkRejected.
func TestBeginStageReportsUndocumentedStageDistinctly(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"myflow: \"bogus\" is not a documented stage","code":"undocumented_stage"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.BeginStage(context.Background(), minimalBeginReq())
	if !errors.Is(err, client.ErrUndocumentedStage) {
		t.Fatalf("err = %v, want ErrUndocumentedStage", err)
	}
	if errors.Is(err, client.ErrStageMarkRejected) {
		t.Errorf("an undocumented-stage 400 must not also read as the generic ErrStageMarkRejected")
	}
	if errors.Is(err, client.ErrUnavailable) {
		t.Errorf("a 400 carrying the daemon header is not a fallback trigger")
	}
}

// TestBeginStageReportsGenericRejectionDistinctly is F1's core
// reproduction: a 400 raised for a reason *other* than an undocumented
// stage name (here, a missing required field -- no "code" at all in the
// body) must map to ErrStageMarkRejected, not be misreported as
// ErrUndocumentedStage. Before this task's fix, every 400 mapped to
// ErrUndocumentedStage unconditionally, which would have made this test
// fail.
func TestBeginStageReportsGenericRejectionDistinctly(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"projectKey, changeName, command, stage and harness are all required"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.BeginStage(context.Background(), minimalBeginReq())
	if !errors.Is(err, client.ErrStageMarkRejected) {
		t.Fatalf("err = %v, want ErrStageMarkRejected", err)
	}
	if errors.Is(err, client.ErrUndocumentedStage) {
		t.Errorf("a non-stage-name 400 must not be misreported as ErrUndocumentedStage -- this is exactly F1's defect")
	}
	if errors.Is(err, client.ErrUnavailable) {
		t.Errorf("a 400 carrying the daemon header is not a fallback trigger")
	}
}

func TestBeginStageReportsRefusalDistinctly(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"refused"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.BeginStage(context.Background(), minimalBeginReq())
	if !errors.Is(err, client.ErrRefused) {
		t.Fatalf("err = %v, want ErrRefused", err)
	}
}

func TestEndStageSucceedsOn200(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":42,"attempt":2}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	result, err := c.EndStage(context.Background(), minimalEndReq())
	if err != nil {
		t.Fatalf("EndStage: %v", err)
	}
	if result.StageRunID != 42 || result.Attempt != 2 {
		t.Errorf("result = %+v, want {StageRunID:42 Attempt:2}", result)
	}
}

func TestEndStageFallsBackOnDeadPort(t *testing.T) {
	c := client.New(deadPortURL(t), &http.Client{Timeout: time.Second})
	_, err := c.EndStage(context.Background(), minimalEndReq())
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestEndStageFallsBackOnTimeout(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := c.EndStage(ctx, minimalEndReq())
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestEndStageFallsBackOnNon2xxThatIsNotAKnownRefusal(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.EndStage(context.Background(), minimalEndReq())
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestEndStageFallsBackWhenResponseLacksDaemonHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.EndStage(context.Background(), minimalEndReq())
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable (a look-alike 200 must fall back)", err)
	}
}

func TestEndStageFallsBackOnMalformedResponseBody(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json at all`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.EndStage(context.Background(), minimalEndReq())
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestEndStageReportsUndocumentedStageDistinctly(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"myflow: \"bogus\" is not a documented stage","code":"undocumented_stage"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.EndStage(context.Background(), minimalEndReq())
	if !errors.Is(err, client.ErrUndocumentedStage) {
		t.Fatalf("err = %v, want ErrUndocumentedStage", err)
	}
	if errors.Is(err, client.ErrStageMarkRejected) {
		t.Errorf("an undocumented-stage 400 must not also read as the generic ErrStageMarkRejected")
	}
}

// TestEndStageReportsGenericRejectionDistinctly is EndStage's half of F1's
// reproduction -- see TestBeginStageReportsGenericRejectionDistinctly.
func TestEndStageReportsGenericRejectionDistinctly(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"projectKey, changeName, command, stage and outcome are all required"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.EndStage(context.Background(), minimalEndReq())
	if !errors.Is(err, client.ErrStageMarkRejected) {
		t.Fatalf("err = %v, want ErrStageMarkRejected", err)
	}
	if errors.Is(err, client.ErrUndocumentedStage) {
		t.Errorf("a non-stage-name 400 must not be misreported as ErrUndocumentedStage -- this is exactly F1's defect")
	}
}

func TestEndStageReportsNotFoundDistinctly(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"no open stage run"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.EndStage(context.Background(), minimalEndReq())
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestEndStageReportsRefusalDistinctly(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"refused"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.EndStage(context.Background(), minimalEndReq())
	if !errors.Is(err, client.ErrRefused) {
		t.Fatalf("err = %v, want ErrRefused", err)
	}
}

// --- F6: a real api.Server wired to a real client.Client ---

// inMemoryChangeStore is a minimal in-memory api.ChangeStore -- just
// enough to drive a real *api.Server without a database, so this test can
// exercise the actual HTTP wire format both sides speak rather than a
// hand-simulated one.
type inMemoryChangeStore struct {
	changes map[string]store.Change
}

func newInMemoryChangeStore() *inMemoryChangeStore {
	return &inMemoryChangeStore{changes: map[string]store.Change{}}
}

func (s *inMemoryChangeStore) key(projectKey, name string) string { return projectKey + "/" + name }

func (s *inMemoryChangeStore) GetChange(_ context.Context, projectKey, name string) (store.Change, error) {
	c, ok := s.changes[s.key(projectKey, name)]
	if !ok {
		return store.Change{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, name)
	}
	return c, nil
}

func (s *inMemoryChangeStore) PutChange(_ context.Context, c store.Change) error {
	k := s.key(c.ProjectKey, c.Name)
	if existing, ok := s.changes[k]; ok && existing.State == store.StateFinished && c.State != store.StateFinished {
		return store.ErrMonotonicViolation
	}
	s.changes[k] = c
	return nil
}

func (s *inMemoryChangeStore) QueryChanges(_ context.Context, _ store.Query) ([]store.Change, int, error) {
	var all []store.Change
	for _, c := range s.changes {
		all = append(all, c)
	}
	return all, len(all), nil
}

// ProjectKeysByDisplayName is here purely to keep satisfying
// api.ChangeStore -- see stubStageStore.ProjectKeysByDisplayName's own
// doc comment for why: this file's tests never send a display-name
// "project" value either.
func (s *inMemoryChangeStore) ProjectKeysByDisplayName(context.Context, string) ([]string, error) {
	return nil, nil
}

var _ api.ChangeStore = (*inMemoryChangeStore)(nil)

// stubStageStore satisfies api.StageStore with nothing but errors: this
// file's tests exercise the change endpoints only, never the stage-mark
// ones, so there is nothing for a real implementation here to do -- but
// api.New now requires a StageStore alongside a ChangeStore (task 8), and
// a stub that answers everything with a clear "not implemented" is safer
// than silently reusing inMemoryChangeStore's own zero-value behaviour for
// methods it was never asked to have.
type stubStageStore struct{}

var errStageStoreNotImplemented = errors.New("stubStageStore: stage endpoints are not exercised by this test file")

func (stubStageStore) BeginStage(context.Context, store.BeginStageInput) (store.StageRun, error) {
	return store.StageRun{}, errStageStoreNotImplemented
}

func (stubStageStore) EndStage(context.Context, int64, time.Time, string) error {
	return errStageStoreNotImplemented
}

func (stubStageStore) MergeMetrics(context.Context, int64, json.RawMessage) error {
	return errStageStoreNotImplemented
}

func (stubStageStore) QueryStageRuns(context.Context, store.Query) ([]store.StageRun, int, error) {
	return nil, 0, errStageStoreNotImplemented
}

func (stubStageStore) PutChange(context.Context, store.Change) error {
	return errStageStoreNotImplemented
}

var _ api.StageStore = stubStageStore{}

// stubStageStore also stands in as api.StatsStore for this test file: task
// 11's statistics endpoints, like the stage endpoints above, are not what
// TestClientAgreesWithRealDaemonOverDaemonHeader exercises -- it only needs
// a real *api.Server to exist, never a stats route to answer correctly.
func (stubStageStore) LiveStateBoard(context.Context, store.Period, *string) ([]store.LiveStateRow, error) {
	return nil, errStageStoreNotImplemented
}

func (stubStageStore) CostPerChange(context.Context, store.Period, *string, *string) ([]store.CostPerChangeRow, error) {
	return nil, errStageStoreNotImplemented
}

func (stubStageStore) StageLeaderboard(context.Context, store.Period, *string, *string) ([]store.StageLeaderboardRow, error) {
	return nil, errStageStoreNotImplemented
}

func (stubStageStore) TrendOverTime(context.Context, store.Period, *string, *string) ([]store.TrendPoint, error) {
	return nil, errStageStoreNotImplemented
}

func (stubStageStore) CacheEfficiency(context.Context, store.Period, *string, *string) ([]store.CacheEfficiencyRow, error) {
	return nil, errStageStoreNotImplemented
}

func (stubStageStore) PanelEconomics(context.Context, store.Period, *string, *string) ([]store.PanelEconomicsRow, error) {
	return nil, errStageStoreNotImplemented
}

func (stubStageStore) ModelComparison(context.Context, store.Period, *string, *string) ([]store.ModelComparisonRow, error) {
	return nil, errStageStoreNotImplemented
}

func (stubStageStore) ReworkRate(context.Context, store.Period, *string, *string) ([]store.ReworkRateRow, error) {
	return nil, errStageStoreNotImplemented
}

func (stubStageStore) CountRunsWithoutModel(context.Context, store.Period, *string) (int, error) {
	return 0, errStageStoreNotImplemented
}

func (stubStageStore) ListModels(context.Context, store.Period, *string) ([]string, error) {
	return nil, errStageStoreNotImplemented
}

// AllRecordedRunsUnmeasured is here purely to keep satisfying
// api.StatsStore -- this file's stubStageStore never exercises a stats
// route (unanticipated file, task 5: the interface it implements gained
// one method, and every implementer of it must compile; mechanical
// substitution, no logic change).
func (stubStageStore) AllRecordedRunsUnmeasured(context.Context, store.Period, *string) (bool, error) {
	return false, errStageStoreNotImplemented
}

// ProjectKeysByDisplayName is here for the same reason
// AllRecordedRunsUnmeasured's own doc comment gives: api.StatsStore (and
// api.ChangeStore, which inMemoryChangeStore implements separately) gained
// this method for task 3's project display-name resolution, and every
// implementer must keep compiling -- this file's tests never send a
// display-name "project" value, so there is nothing for a real
// implementation here to do.
func (stubStageStore) ProjectKeysByDisplayName(context.Context, string) ([]string, error) {
	return nil, errStageStoreNotImplemented
}

var _ api.StatsStore = stubStageStore{}

// TestClientAgreesWithRealDaemonOverDaemonHeader is F6's fix: the daemon
// header value exists in three places -- api.DaemonHeaderValue (the
// source), internal/client's own literal copy (necessary, per the CLI's
// "never import the daemon's internal packages" boundary), and every
// other test in this file's genuineDaemon fake (now reading
// api.DaemonHeaderValue directly, see genuineDaemon's doc comment). None
// of that catches internal/client's copy drifting from api's real value --
// a fake server that hand-sets the header always agrees with whatever the
// test file believes the value is, never with what a real daemon actually
// sends.
//
// This test is the one thing that does: it wires a real *api.Server
// (internal/api, the same package cmd/myflowd serves) directly to a real
// *client.Client (internal/client, the same package cmd/myflow drives) --
// no fake server standing in for either side -- and asserts a genuine PUT
// round-trips as success, a genuine GET returns it, and a genuine
// monotonic refusal surfaces as client.ErrRefused. If internal/client's
// literal ever drifts from api.DaemonHeaderValue, every response the real
// daemon sends stops carrying a header the client recognises, and this
// test is what turns that into a failure instead of a silent
// always-falls-back CLI.
func TestClientAgreesWithRealDaemonOverDaemonHeader(t *testing.T) {
	cs := newInMemoryChangeStore()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	apiServer, err := api.New(cfg, cs, stubStageStore{}, stubStageStore{}, nil)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	srv := httptest.NewServer(apiServer.Handler())
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())

	// A genuine write round-trips as success.
	body := []byte(`{"state":"IN_PROGRESS","mainCheckoutPath":"/repo","updatedAt":"2026-08-13T10:00:00Z","updatedBy":"tester"}`)
	if err := c.PutChange(context.Background(), "proj", "chg", body); err != nil {
		t.Fatalf("PutChange against a real api.Server: %v", err)
	}

	// A genuine read returns what was just written.
	got, err := c.GetChange(context.Background(), "proj", "chg")
	if err != nil {
		t.Fatalf("GetChange against a real api.Server: %v", err)
	}
	if len(got) == 0 {
		t.Error("GetChange returned an empty body for a change that was just written")
	}

	// A genuine monotonic refusal surfaces as ErrRefused, not
	// ErrUnavailable -- this is the direction that would break silently if
	// internal/client's header literal ever agreed with nothing (the
	// refusal would then be indistinguishable from "unreachable" and the
	// CLI would fall back on a real refusal instead of reporting it).
	finished := []byte(`{"state":"FINISHED","mainCheckoutPath":"/repo","updatedAt":"2026-08-13T11:00:00Z","updatedBy":"tester"}`)
	if err := c.PutChange(context.Background(), "proj", "chg", finished); err != nil {
		t.Fatalf("advance to FINISHED: %v", err)
	}
	regress := []byte(`{"state":"STARTED","mainCheckoutPath":"/repo","updatedAt":"2026-08-13T12:00:00Z","updatedBy":"tester"}`)
	err = c.PutChange(context.Background(), "proj", "chg", regress)
	if !errors.Is(err, client.ErrRefused) {
		t.Fatalf("err = %v, want client.ErrRefused for a genuine monotonic refusal from a real api.Server", err)
	}
	if errors.Is(err, client.ErrUnavailable) {
		t.Errorf("a genuine refusal from a real api.Server must not also read as ErrUnavailable")
	}
}
