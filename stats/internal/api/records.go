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
	SetFindingStatus(ctx context.Context, projectKey, change, ref, status, category string) error

	// RecordDecision and ListDecisions carry one run's dynamic decision --
	// the whole `## Decision` block a plan return appends -- the same way
	// the four methods above carry a dispatch or a finding: RecordDecision
	// sits here, rather than on RecordStore alone as the guard-log writes
	// do, because a journalled "decision" entry replays through this
	// interface exactly as a journalled dispatch or finding does.
	// ListDecisions has no replay use of its own -- nothing decodes a list
	// read from a journal entry -- but stays beside its write here rather
	// than splitting the pair across two interfaces for no caller's benefit.
	RecordDecision(ctx context.Context, projectKey, change string, in records.Decision) (records.Decision, bool, error)
	ListDecisions(ctx context.Context, projectKey, change string) ([]records.Decision, error)

	// RecordPass and RecordMutation carry the review panel's pass log --
	// the pass-by-pass metadata lines and the fix round's fix-mutation:
	// proof lines (KAN-331). They sit here beside RecordDecision for the
	// same reason it does: a journalled "pass" or "mutation" entry replays
	// through this interface exactly as a journalled dispatch or finding
	// does.
	RecordPass(ctx context.Context, projectKey, change string, in records.Pass) (records.Pass, error)
	RecordMutation(ctx context.Context, projectKey, change string, in records.Mutation) (records.Mutation, error)
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

	// RecordVerdict, FlagVerdictFalsePositive, ListVerdicts, RecordIncident
	// and ListIncidents are KAN-451's guard-log methods -- see
	// design.md's schema section for what each table records. They sit on
	// RecordStore rather than RecordWriter: internal/reconcile's replay
	// path (RecordWriter's own reason for existing, see its doc comment)
	// is a later task's concern, not this one's.
	RecordVerdict(ctx context.Context, projectKey, change string, in records.Verdict) (records.Verdict, error)
	FlagVerdictFalsePositive(ctx context.Context, projectKey, change string, in records.VerdictFlag) (records.Verdict, error)
	ListVerdicts(ctx context.Context, projectKey, guard string, falsePositiveOnly bool) ([]records.Verdict, error)
	RecordIncident(ctx context.Context, projectKey string, in records.Incident) (records.Incident, error)
	ListIncidents(ctx context.Context, projectKey string) ([]records.Incident, error)

	// AddHazard, ListHazards and RetireHazard are KAN-452's hazard
	// methods -- the proactive sibling of the guard-log pair above, and on
	// RecordStore for the same reason.
	AddHazard(ctx context.Context, projectKey string, h records.Hazard) (records.Hazard, error)
	ListHazards(ctx context.Context, projectKey, shape string, includeInactive bool) ([]records.Hazard, error)
	RetireHazard(ctx context.Context, projectKey, name string) (records.Hazard, error)

	// InsertSuiteRun and ListSuiteRuns are KAN-252's suite-runtime
	// methods -- per-run, per-machine figures project.md cites instead of
	// pasting measured durations into prose -- and on RecordStore for the
	// reason the hazard methods state above.
	InsertSuiteRun(ctx context.Context, projectKey string, run records.SuiteRun) (records.SuiteRun, error)
	ListSuiteRuns(ctx context.Context, projectKey, suite string, limit int) ([]records.SuiteRun, error)
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
// network call, and a role this build of flowd has never heard of is not
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

// ApplyDecisionRecord records one run's dynamic decision against rw,
// refusing a record whose session token or decision body is empty before
// the store is touched. It reports whether the write inserted a row,
// exactly as ApplyFindingRecord does -- what the route turns into 201
// versus 200. See ApplyDispatchRecord for why the checks live here rather
// than in the handler.
func ApplyDecisionRecord(ctx context.Context, rw RecordWriter, projectKey, change string, in records.Decision) (records.Decision, bool, error) {
	if in.SessionToken == "" || len(in.Decision) == 0 {
		return records.Decision{}, false, fmt.Errorf("%w: sessionToken and decision are both required", ErrInvalidRecord)
	}
	return rw.RecordDecision(ctx, projectKey, change, in)
}

// ApplyPassRecord records one panel pass-log entry against rw, refusing an
// empty note before the store is touched. Round carries no check: 0 is the
// initial panel and a meaningful value, the same shape Finding.Round
// already takes. See ApplyDispatchRecord for why the checks live here
// rather than in the handler.
func ApplyPassRecord(ctx context.Context, rw RecordWriter, projectKey, change string, in records.Pass) (records.Pass, error) {
	if in.Note == "" {
		return records.Pass{}, fmt.Errorf("%w: note is required: a pass entry with no line records nothing", ErrInvalidRecord)
	}
	return rw.RecordPass(ctx, projectKey, change, in)
}

// ApplyMutationRecord records one fix-mutation: line of the fix round's
// mutation proof against rw, refusing an empty path, mutated or test field
// before the store is touched -- the contract's line has exactly three
// fields, and the exemption form still fills all three ("none" and the
// reason). See ApplyDispatchRecord for why the checks live here rather
// than in the handler.
func ApplyMutationRecord(ctx context.Context, rw RecordWriter, projectKey, change string, in records.Mutation) (records.Mutation, error) {
	if in.Path == "" || in.Mutated == "" || in.Test == "" {
		return records.Mutation{}, fmt.Errorf("%w: path, mutated and test are all required: the contract's fix-mutation line has exactly three fields", ErrInvalidRecord)
	}
	return rw.RecordMutation(ctx, projectKey, change, in)
}

// ApplyFindingStatus rewrites one finding's status, and its deferral
// category, against rw.
//
// ref is deliberately not checked for emptiness, where status is: an empty
// ref is already answered correctly one layer down, by an UPDATE that
// matches zero rows and reports store.ErrFindingNotFound -- a 404 on the
// route, a definitive refusal on replay. A check here would move that
// answer without improving it. An empty *status* has no such backstop: the
// column is NOT NULL, which an empty string satisfies, so it would be
// written. The category is not checked here at all: an empty category is
// the ordinary category-less write, and which non-empty words are legal is
// the CLI validator's judgment, the same split -status's own shape takes.
func ApplyFindingStatus(ctx context.Context, rw RecordWriter, projectKey, change, ref, status, category string) error {
	if status == "" {
		return fmt.Errorf("%w: status is required", ErrInvalidRecord)
	}
	return rw.SetFindingStatus(ctx, projectKey, change, ref, status, category)
}

// ApplyVerdictRecord records one guard's verdict against rw, refusing a
// record whose guard, worktree or verdict line is empty before the store is
// touched. It takes RecordStore rather than RecordWriter -- unlike the four
// checks above, KAN-451's guard-log methods sit on RecordStore alone (see
// RecordStore's own doc comment for why) -- so this is what
// internal/reconcile's replay of a journalled "verdict" entry shares with
// this route, for the identical reason ApplyDispatchRecord exists.
func ApplyVerdictRecord(ctx context.Context, rw RecordStore, projectKey, change string, in records.Verdict) (records.Verdict, error) {
	if in.Guard == "" || in.Worktree == "" || in.Verdict == "" {
		return records.Verdict{}, fmt.Errorf("%w: guard, worktree and verdict are all required", ErrInvalidRecord)
	}
	return rw.RecordVerdict(ctx, projectKey, change, in)
}

// ApplyVerdictFlag flags project/change's most recent verdict for in.Guard
// as a false positive, refusing an empty guard or reason before the store
// is touched. See ApplyVerdictRecord for why it takes RecordStore.
func ApplyVerdictFlag(ctx context.Context, rw RecordStore, projectKey, change string, in records.VerdictFlag) (records.Verdict, error) {
	if in.Guard == "" || in.Reason == "" {
		return records.Verdict{}, fmt.Errorf("%w: guard and reason are both required", ErrInvalidRecord)
	}
	return rw.FlagVerdictFalsePositive(ctx, projectKey, change, in)
}

// ApplyIncidentRecord records one per-project incident against rw, refusing
// an empty guard, symptom or recovery, or a negative minutesLost, before the
// store is touched. See ApplyVerdictRecord for why it takes RecordStore.
func ApplyIncidentRecord(ctx context.Context, rw RecordStore, projectKey string, in records.Incident) (records.Incident, error) {
	if in.Guard == "" || in.Symptom == "" || in.Recovery == "" {
		return records.Incident{}, fmt.Errorf("%w: guard, symptom and recovery are all required", ErrInvalidRecord)
	}
	if in.MinutesLost < 0 {
		return records.Incident{}, fmt.Errorf("%w: minutesLost must not be negative", ErrInvalidRecord)
	}
	return rw.RecordIncident(ctx, projectKey, in)
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
// PATCH .../findings/{ref} accepts: the status a fix round rewrites, plus
// the deferral category that rides beside a `deferred <reason>` status --
// omitted, or empty, for every other status and for a deferral that names
// no category, which is what clears a category an earlier deferral set.
type findingStatusRequest struct {
	Status   string `json:"status"`
	Category string `json:"category,omitempty"`
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

	if err := ApplyFindingStatus(r.Context(), h.store, project, change, ref, req.Status, req.Category); err != nil {
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

// recordDecision serves POST /api/v1/records/{project}/{change}/decisions:
// one run's dynamic decision, the whole `## Decision` block a plan return
// appends. It answers 201 when the write inserted and 200 when it
// replaced a decision already recorded under the same session token --
// the same split RecordFinding's route makes, since a resumed run
// restating its own choice under the token it already holds is a replay,
// never a second decision.
func (h *recordHandler) recordDecision(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	var in records.Decision
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, created, err := ApplyDecisionRecord(r.Context(), h.store, project, change, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("record decision for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	if created {
		writeJSON(w, http.StatusCreated, out)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// listDecisions serves GET /api/v1/records/{project}/{change}/decisions:
// a change's recorded decisions, newest first -- what a resumed run reads
// to recover a stopped run's own choice.
func (h *recordHandler) listDecisions(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	out, err := h.store.ListDecisions(r.Context(), project, change)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list decisions for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// recordPass serves POST /api/v1/records/{project}/{change}/passes: one
// pass-log entry of the review panel's record (KAN-331). Every insert is a
// new row -- a pass entry is recorded once as the fact arises, and the
// 201/200 split the upsert routes make has no meaning here.
func (h *recordHandler) recordPass(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	var in records.Pass
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := ApplyPassRecord(r.Context(), h.store, project, change, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("record pass for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// recordMutation serves POST /api/v1/records/{project}/{change}/mutations:
// one fix-mutation: line of the fix round's mutation proof (KAN-331).
// Append-only like recordPass.
func (h *recordHandler) recordMutation(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	var in records.Mutation
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := ApplyMutationRecord(r.Context(), h.store, project, change, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("record mutation for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, out)
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

// recordVerdict serves POST /api/v1/records/{project}/{change}/verdicts:
// one guard's verdict on one change. Every call inserts a new row -- see
// store.RecordVerdict's own doc comment for why a re-entered run is a
// distinct verdict, never a replay -- so this always answers 201.
func (h *recordHandler) recordVerdict(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	var in records.Verdict
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := ApplyVerdictRecord(r.Context(), h.store, project, change, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("record verdict for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// flagVerdict serves POST
// /api/v1/records/{project}/{change}/verdicts/false-positive: an
// operator's judgment that the most recent verdict the named guard reached
// for this change was wrong, and why -- see design.md's
// flag-latest-verdict decision for why no verdict id is named.
//
// A (change, guard) pair the store holds no verdict for is a 404, through
// mapStoreError's store.ErrDispatchNotFound case (the same reused sentinel
// EndDispatch's unknown key answers) -- not 500, for the reason every other
// 404 in this file is deliberate: internal/client reads a 500 as "the store
// is unavailable", and a mistaken flag would be journalled for a replay
// that could never succeed.
func (h *recordHandler) flagVerdict(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	var in records.VerdictFlag
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := ApplyVerdictFlag(r.Context(), h.store, project, change, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("flag verdict false positive for %s/%s guard %q", project, change, in.Guard), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// listVerdicts serves GET
// /api/v1/verdicts/{project}?guard=&falsePositive=true: a project's guard
// verdicts, newest first. guard absent means every guard; falsePositive,
// when present, must be exactly "true" or "false" -- anything else is a
// caller mistake, refused before the store is touched, since a query
// silently ignored would answer a different question than the one asked.
func (h *recordHandler) listVerdicts(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")
	guard := r.URL.Query().Get("guard")

	falsePositiveOnly := false
	if v := r.URL.Query().Get("falsePositive"); v != "" {
		switch v {
		case "true":
			falsePositiveOnly = true
		case "false":
			falsePositiveOnly = false
		default:
			writeError(w, http.StatusBadRequest, fmt.Sprintf("falsePositive must be true or false, got %q", v))
			return
		}
	}

	out, err := h.store.ListVerdicts(r.Context(), project, guard, falsePositiveOnly)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list verdicts for %s", project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// recordIncident serves POST /api/v1/incidents/{project}: a per-project
// record of one guard actually costing time -- what went wrong, the
// recovery taken and how many minutes it cost. guard, symptom and recovery
// are refused empty, and a negative minutesLost is refused before the
// store is touched -- both caller mistakes a different request would have
// avoided.
func (h *recordHandler) recordIncident(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	var in records.Incident
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := ApplyIncidentRecord(r.Context(), h.store, project, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("record incident for %s", project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// listIncidents serves GET /api/v1/incidents/{project}: a project's
// incidents, newest first, exactly as store.ListIncidents returns them.
func (h *recordHandler) listIncidents(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	out, err := h.store.ListIncidents(r.Context(), project)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list incidents for %s", project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
