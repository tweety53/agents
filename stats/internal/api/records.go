package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// RecordWriter is the write half of the run-record store dependency:
// exactly the four methods a record write -- live or replayed -- can land
// in. It is what the ApplyRecord* functions below take, which is why it is
// a named interface of its own rather than part of RecordStore: replay
// never reads a record back, so internal/reconcile depends on these four
// methods and no fifth, and used to declare its own identical three-method
// interface to say so. That duplicate is gone -- reconcile.New now takes
// this one -- because two interfaces with the same method set and no
// difference between them is a difference waiting to happen.
type RecordWriter interface {
	RecordDispatch(ctx context.Context, projectKey, change string, in records.Dispatch) (records.Dispatch, error)
	EndDispatch(ctx context.Context, projectKey, change string, in records.DispatchEnd) (records.Dispatch, error)
	UpsertFinding(ctx context.Context, projectKey, change string, in records.Finding) (records.Finding, bool, error)
	SetFindingStatus(ctx context.Context, projectKey, change, ref, status string) error
}

// RecordStore is the store dependency the run-record endpoints need,
// defined here at the consumer per go-interface-design -- exactly the
// methods recordHandler calls, so a test needs no database. It is
// RecordWriter plus the one read the routes add.
//
// store.MergeDispatchMetrics is deliberately absent: it is the harvester's
// way into a dispatch row, not a route, and an interface naming a method
// no handler in this package calls would be an interface its own fake has
// to stub out for nothing.
type RecordStore interface {
	RecordWriter
	RunRecord(ctx context.Context, projectKey, change string) (records.Run, error)
}

// var _ RecordStore = (*store.Store)(nil) verifies at compile time that the
// real store satisfies the interface this package actually depends on.
var _ RecordStore = (*store.Store)(nil)

// ErrInvalidRecord is what the ApplyRecord* functions return when a record
// is missing a field the row cannot be read without. It is a caller
// mistake, not a store failure: the delta spec's "the daemon refuses the
// request" scenario, which the route answers 400 for and which
// internal/reconcile retires rather than requeueing, since a body refused
// on these grounds is refused identically forever.
//
// It deliberately does not cover the checks cmd/flow makes for itself --
// an unrecognised -role, a session token carrying a shell substitution.
// Those are the spec's *first* kind of caller mistake, judged before any
// network call, and a role this build of myflowd has never heard of is not
// a reason for the daemon to reject a row it can store perfectly well.
var ErrInvalidRecord = errors.New("api: invalid record")

// ApplyDispatchRecord records one dispatch against rw, refusing a record
// whose required fields are empty before the store is touched at all.
//
// It exists so that recordDispatch (the live POST route) and
// internal/reconcile's replay of a journalled dispatch cannot answer
// "is this dispatch recordable?" differently -- the same reason
// ApplyBeginStageMark exists on the stage side. That is not a stylistic
// preference: migration 0010 makes these columns NOT NULL, and an empty
// string satisfies NOT NULL, so a replay that skipped these checks would
// insert a row with an empty role and report success.
func ApplyDispatchRecord(ctx context.Context, rw RecordWriter, projectKey, change string, in records.Dispatch) (records.Dispatch, error) {
	if in.Role == "" || in.Model == "" {
		return records.Dispatch{}, fmt.Errorf("%w: role and model are both required: a dispatch row records what ran and on what", ErrInvalidRecord)
	}
	if in.StartedAt.IsZero() {
		return records.Dispatch{}, fmt.Errorf("%w: startedAt is required: the harvester attributes a dispatch's cost to the window it opens", ErrInvalidRecord)
	}
	return rw.RecordDispatch(ctx, projectKey, change, in)
}

// ApplyDispatchEnd closes one dispatch against rw, refusing a record that
// cannot name the row it closes, or that carries no end instant, before the
// store is touched at all.
//
// The session token and the key are both required because together they
// ARE the row's name: the caller has no seq to close by, since a `begin`
// whose response was lost never returned one. An empty end instant is
// refused for the reason the call exists: leaving ended_at NULL is exactly
// the open window this half of the pair was added to close, so a request
// that omits it would report success while changing nothing that matters.
//
// It exists beside the live route for the reason ApplyDispatchRecord does:
// replay and the handler must not answer "is this closable?" differently.
func ApplyDispatchEnd(ctx context.Context, rw RecordWriter, projectKey, change string, in records.DispatchEnd) (records.Dispatch, error) {
	if in.SessionToken == "" || in.Key == "" {
		return records.Dispatch{}, fmt.Errorf("%w: sessionToken and key are both required: together they name the dispatch being closed", ErrInvalidRecord)
	}
	if in.EndedAt.IsZero() {
		return records.Dispatch{}, fmt.Errorf("%w: endedAt is required: an unclosed window goes on claiming later usage", ErrInvalidRecord)
	}
	return rw.EndDispatch(ctx, projectKey, change, in)
}

// ApplyFindingRecord upserts one finding against rw, refusing a record
// whose required fields are empty before the store is touched. It reports
// whether the write inserted a row, exactly as UpsertFinding does. See
// ApplyDispatchRecord for why the checks live here rather than in the
// handler.
func ApplyFindingRecord(ctx context.Context, rw RecordWriter, projectKey, change string, in records.Finding) (records.Finding, bool, error) {
	if in.Ref == "" || in.Slot == "" || in.Severity == "" || in.Note == "" || in.Status == "" {
		return records.Finding{}, false, fmt.Errorf("%w: ref, slot, severity, note and status are all required", ErrInvalidRecord)
	}
	return rw.UpsertFinding(ctx, projectKey, change, in)
}

// ApplyFindingStatus rewrites one finding's status against rw.
//
// ref is deliberately not checked for emptiness, where status is: an empty
// ref is already answered correctly one layer down, by an UPDATE that
// matches zero rows and reports store.ErrFindingNotFound -- a 404 on the
// route, a definitive refusal on replay. A check here would move that
// answer without improving it. An empty *status* has no such backstop: the
// column is NOT NULL, which an empty string satisfies, so it would be
// written.
func ApplyFindingStatus(ctx context.Context, rw RecordWriter, projectKey, change, ref, status string) error {
	if status == "" {
		return fmt.Errorf("%w: status is required", ErrInvalidRecord)
	}
	return rw.SetFindingStatus(ctx, projectKey, change, ref, status)
}

// recordHandler serves the four run-record endpoints. Each is thin: decode
// the wire shape, refuse a body that cannot identify what it describes,
// and hand the typed record to the store.
//
// The wire shapes are internal/records' own types rather than a second set
// of request structs beside them. A dispatch and a finding cross four
// components -- the CLI, this daemon, the store and the renderer -- as one
// shape, and a per-layer copy of that shape would be four places to add a
// field to instead of one. That is the opposite trade to stageBeginRequest,
// which exists because a stage mark's wire form (an RFC 3339 string, a
// client-supplied session token) genuinely differs from the typed mark the
// store is handed.
type recordHandler struct {
	store  RecordStore
	logger *slog.Logger
}

// findingStatusRequest is the wire shape
// PATCH .../findings/{ref} accepts: the one column a fix round rewrites.
type findingStatusRequest struct {
	Status string `json:"status"`
}

// recordDispatch serves POST /api/v1/records/{project}/{change}/dispatches:
// the OPENING half of a dispatch's record, written when the dispatch starts
// rather than when it ends. See records.DispatchEnd for why the pair
// exists at all -- in short, the harvester never re-reads past a committed
// offset, so a row that appears only at close is invisible to every harvest
// tick the dispatch ran through.
//
// It answers 201 whether the write inserted a row or matched one already
// recorded under the same session token and key. Unlike a finding, whose
// 201/200 split tells a fix round from the panel that first raised it,
// nothing here needs to distinguish the two: the second write is a replay
// of the first, describing the same dispatch, and reporting it as an
// update would invite a caller to treat a duplicate delivery as a second
// dispatch.
//
// The allocated seq and the row's own id come back on the response body,
// because neither is knowable to the caller: the store allocates both, and
// a caller that could not read them back would have to guess where in the
// change's record its own dispatch landed.
func (h *recordHandler) recordDispatch(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	var in records.Dispatch
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := ApplyDispatchRecord(r.Context(), h.store, project, change, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			// A caller mistake, not a store failure: mapStoreError's
			// generic default (500 "internal error") would both swallow
			// the reason and tell internal/client the daemon is broken,
			// which sends a body that can never be accepted to the
			// journal for a replay that can never succeed.
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("record dispatch for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// endDispatch serves POST
// /api/v1/records/{project}/{change}/dispatches/end: the CLOSING half,
// naming its row by the session token and key its begin carried.
//
// It answers 200, not 201: the row already exists and this call changes
// three of its columns. A key naming no row is a 404 through
// mapStoreError's store.ErrDispatchNotFound case, which matters beyond
// politeness -- internal/client reads a 500 as "the store is unavailable",
// and this one 404 is the answer the CLI deliberately journals rather than
// refuses, the begin it closes possibly still being queued ahead of it.
func (h *recordHandler) endDispatch(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	var in records.DispatchEnd
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := ApplyDispatchEnd(r.Context(), h.store, project, change, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("end dispatch %q for %s/%s", in.Key, project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// recordFinding serves POST /api/v1/records/{project}/{change}/findings.
// It answers 201 when the write inserted a row and 200 when it replaced
// one, so a caller can tell the round that first raised a finding from a
// fix round restating it -- a distinction the response body cannot carry,
// since an upsert's body is identical either way.
func (h *recordHandler) recordFinding(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	var in records.Finding
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, created, err := ApplyFindingRecord(r.Context(), h.store, project, change, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("upsert finding %s for %s/%s", in.Ref, project, change), err)
		writeError(w, status, msg)
		return
	}
	if created {
		writeJSON(w, http.StatusCreated, out)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// setFindingStatus serves PATCH
// /api/v1/records/{project}/{change}/findings/{ref}: the one column a fix
// round rewrites, and nothing else. It answers 204, since the caller
// already holds everything the finding now says.
//
// A ref the change holds no finding under is a 404 rather than a 500 --
// mapStoreError maps store.ErrFindingNotFound. That mapping is load
// bearing beyond politeness: internal/client reads a 500 as the store
// being unavailable, which would send a mistyped ref to the journal for a
// replay that can never succeed.
func (h *recordHandler) setFindingStatus(w http.ResponseWriter, r *http.Request) {
	project, change, ref := r.PathValue("project"), r.PathValue("change"), r.PathValue("ref")

	var req findingStatusRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := ApplyFindingStatus(r.Context(), h.store, project, change, ref, req.Status); err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("set finding %s status for %s/%s", ref, project, change), err)
		writeError(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// runRecord serves GET /api/v1/records/{project}/{change}: the change's
// whole derived record, in the order a render reads it. A change that
// exists and holds no rows is an empty record with a 200; only a change
// the store has never heard of is a 404, so "this change has no findings"
// can never be mistaken for "this change does not exist".
func (h *recordHandler) runRecord(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	rec, err := h.store.RunRecord(r.Context(), project, change)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("read the run record for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

// costStatus serves GET /api/v1/records/{project}/{change}/cost-status: how
// many of a change's dispatches carry no cost figure, and why. It is
// derived from the same run record runRecord serves -- records.CostStatusOf
// applies tokenLine's own precedence to every dispatch -- so this route can
// never disagree with what the rendered ledger would say about the same
// rows, and no second store query is needed to answer it.
func (h *recordHandler) costStatus(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	rec, err := h.store.RunRecord(r.Context(), project, change)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("read cost status for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, records.CostStatusOf(rec))
}
