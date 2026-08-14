package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/stages"
	"github.com/tweety53/agents/stats/internal/store"
)

// StageStore is the store dependency the stage-mark endpoints need,
// defined here at the consumer per go-interface-design -- exactly the
// methods stageHandler calls, so a test needs no database. PutChange is
// part of this surface for exactly one reason: bootstrapping the synthetic
// change row a mark for an unknown change is stored against (see begin's
// own doc comment), not for anything else a stage mark does.
type StageStore interface {
	BeginStage(ctx context.Context, in store.BeginStageInput) (store.StageRun, error)
	EndStage(ctx context.Context, stageRunID int64, endedAt time.Time, outcome string) error
	MergeMetrics(ctx context.Context, stageRunID int64, patch json.RawMessage) error
	QueryStageRuns(ctx context.Context, q store.Query) ([]store.StageRun, int, error)
	PutChange(ctx context.Context, c store.Change) error
}

// var _ StageStore = (*store.Store)(nil) verifies at compile time that the
// real store satisfies the interface this package actually depends on.
var _ StageStore = (*store.Store)(nil)

// stageHandler serves the two stage-mark endpoints. Both are thin: they
// decode the wire request, build a typed mark, and hand it to
// ApplyBeginStageMark/ApplyEndStageMark -- the same two functions
// internal/reconcile calls to replay a journalled mark once the store is
// reachable again. Keeping the actual begin/end logic there, rather than
// in this handler, is what lets replay reuse it instead of growing a
// second implementation of "what a begin mark does" that could drift from
// this one.
type stageHandler struct {
	store  StageStore
	logger *slog.Logger
}

// stageBeginRequest is the wire shape POST /api/v1/stages/begin accepts,
// matching internal/client.BeginStageRequest's JSON field for field.
type stageBeginRequest struct {
	ProjectKey       string  `json:"projectKey"`
	MainCheckoutPath string  `json:"mainCheckoutPath,omitempty"`
	ChangeName       string  `json:"changeName"`
	RepoRoot         *string `json:"repoRoot,omitempty"`
	Harness          string  `json:"harness"`
	SessionID        *string `json:"sessionId,omitempty"`
	// SessionToken is the literal, unique correlator `stage begin` wrote into its
	// own command text (KAN-172, task 1) -- required, exactly as Harness
	// is, so the daemon always has something a later harvest cycle
	// (task 2) can look for in a transcript.
	SessionToken string `json:"sessionToken"`
	Command      string `json:"command"`
	Stage        string `json:"stage"`
	StartedAt    string `json:"startedAt"`
}

// stageEndRequest is the wire shape POST /api/v1/stages/end accepts. It
// deliberately carries no numeric stage run id -- see end's own doc
// comment for why the daemon resolves the open run itself.
type stageEndRequest struct {
	ProjectKey string          `json:"projectKey"`
	ChangeName string          `json:"changeName"`
	Command    string          `json:"command"`
	Stage      string          `json:"stage"`
	EndedAt    string          `json:"endedAt"`
	Outcome    string          `json:"outcome"`
	Metrics    json.RawMessage `json:"metrics,omitempty"`
}

// stageRunResponse is the wire shape both endpoints answer with on
// success: the stage run's identity, which a caller (a future harvester,
// an operator re-running `stage end`) can use to address it.
type stageRunResponse struct {
	StageRunID int64 `json:"stageRunId"`
	Attempt    int   `json:"attempt"`
}

// StageMarkResult is the identity ApplyBeginStageMark/ApplyEndStageMark
// return on success, shared by the HTTP wire response (stageRunResponse)
// and by internal/reconcile's replay path.
type StageMarkResult struct {
	StageRunID int64
	Attempt    int
}

// BeginStageMark is the typed, transport-agnostic shape of a begin mark:
// what both the HTTP handler (decoded from stageBeginRequest) and
// internal/reconcile (decoded from a journalled entry) build before
// calling ApplyBeginStageMark.
type BeginStageMark struct {
	ProjectKey       string
	MainCheckoutPath string
	ChangeName       string
	RepoRoot         *string
	Harness          string
	SessionID        *string
	// SessionToken is the literal correlator `stage begin` recorded. See
	// stageBeginRequest.SessionToken's doc comment for what it is and why it is
	// required; ApplyBeginStageMark validates its shape
	// (validateSessionTokenShape) before it ever reaches the store.
	SessionToken string
	Command      string
	Stage        string
	StartedAt    time.Time
}

// EndStageMark is the typed, transport-agnostic shape of an end mark. See
// BeginStageMark's doc comment.
type EndStageMark struct {
	ProjectKey string
	ChangeName string
	Command    string
	Stage      string
	EndedAt    time.Time
	Outcome    string
	Metrics    json.RawMessage
}

// claudeCodeHarness is the only harness value whose sessions write a
// machine-readable transcript for internal/harvest to read -- design.md's
// "Harvesting" section and its "Claude Code gets full telemetry; other
// harnesses degrade honestly" decision.
//
// ApplyEndStageMark reads this against the stage run's own recorded
// Harness -- the value BeginStage stored at `stage begin`, immutable
// afterwards -- rather than anything an end mark supplies, because an end
// mark carries no harness at all (see ApplyEndStageMark's own doc comment
// on this: task 10's post-commit review, finding F1). A field whose only
// job is honesty about whether a metric was measurable must not be
// derived from a value the caller can simply omit or re-resolve
// differently on every call.
const claudeCodeHarness = "claude-code"

// syntheticChangeUpdatedBy marks a change row PutChange creates only to
// give an otherwise-unknown mark somewhere to attach -- never a value a
// real `state set` writes, so a record carrying it is immediately
// recognisable as "nobody ever ran /myflow-start for this", exactly the
// defect design.md says a mark for an unknown change is worth seeing.
const syntheticChangeUpdatedBy = "myflow stage begin (synthetic)"

// ErrInvalidSessionToken is returned by ApplyBeginStageMark when mark.SessionToken is
// empty, or contains a shape that a shell would expand before the
// transcript ever saw it (validateSessionTokenShape). Like ErrUnknownStage, this
// is a caller mistake, never a store failure: the handler answers 400
// carrying the reason (never mapStoreError's generic 500 fallback), and
// internal/client classifies it as ErrStageMarkRejected, never
// ErrUnavailable.
var ErrInvalidSessionToken = errors.New("api: invalid sessionToken")

// sessionTokenShellVarPattern matches a "$" immediately followed by a shell
// variable-name character ($VAR, $HOME, $_x) -- one of the three
// substitution shapes design.md's "the sessionToken is a literal, never a shell
// substitution" names. It deliberately does not match bare "$" or "$$"
// (a PID expansion) on their own: the task names exactly three shapes
// ("$(", a backtick, and "$" followed by a name), and this is the pattern
// for the third.
var sessionTokenShellVarPattern = regexp.MustCompile(`\$[A-Za-z_]`)

// validateSessionTokenShape rejects a sessionToken that cannot identify anything: one
// that is empty, or that carries a shell-substitution shape a caller's own
// shell would expand before the transcript is ever written --
// cmd/myflow/stage.go's runStageBegin runs the identical check, client
// side, before the store is ever contacted, for the same reason
// stages.Validate is checked in both places (defence in depth: a mark
// replayed from internal/reconcile's journal reaches ApplyBeginStageMark
// without ever passing back through the CLI's own check). The two copies
// are not shared code -- cmd/myflow knows only HTTP, never this package
// (design.md, "Boundaries") -- so a change to one must be mirrored in the
// other; each has its own test file pinning the same three shapes and the
// same reasoning in its message.
//
// Why each shape is rejected: `tool_use.input.command`, which the
// harvester later reads a transcript for (KAN-172, task 2), records the
// command text exactly as it was handed to the tool -- before the shell
// ever expands it. A sessionToken built from `$(...)`, a backtick, or `$VAR`
// therefore lands in every calling session's transcript as the identical,
// unexpanded literal, and discriminates nothing between them.
func validateSessionTokenShape(sessionToken string) error {
	switch {
	case sessionToken == "":
		return fmt.Errorf("%w: a sessionToken is required", ErrInvalidSessionToken)
	case strings.Contains(sessionToken, "$("):
		return fmt.Errorf("%w: %q contains a command substitution \"$(...)\" -- "+
			"the transcript records the command text before the shell expands it, "+
			"so this value would be recorded identically by every caller and identify nothing; "+
			"write a literal token instead", ErrInvalidSessionToken, sessionToken)
	case strings.Contains(sessionToken, "`"):
		return fmt.Errorf("%w: %q contains a backtick -- backtick command substitution is expanded by "+
			"the shell after the transcript already recorded the unexpanded text, "+
			"so this value would be recorded identically by every caller and identify nothing; "+
			"write a literal token instead", ErrInvalidSessionToken, sessionToken)
	case sessionTokenShellVarPattern.MatchString(sessionToken):
		return fmt.Errorf("%w: %q contains a shell variable reference ($VAR) -- "+
			"the transcript records the command text before the shell expands it, "+
			"so this value would be recorded identically by every caller and identify nothing; "+
			"write a literal token instead", ErrInvalidSessionToken, sessionToken)
	}
	return nil
}

// ErrNoOpenStageRun is ApplyEndStageMark's own not-found condition: no
// stage run matching the requested identity is currently open. It is
// distinct from store.ErrStageRunNotFound (which names a specific,
// already-known id) -- an end mark never carries an id, so this is what it
// reports instead when nothing matches.
var ErrNoOpenStageRun = errors.New("api: no open stage run for that change, command and stage")

// IsDefinitiveMarkOutcome reports whether err represents ApplyBeginStageMark
// or ApplyEndStageMark reaching a *definitive* answer -- success, or a
// refusal that will not change on retry (an undocumented stage name, or no
// open run left to end) -- as opposed to a transient failure (a dropped
// connection, contention on attempt allocation, a context cancellation)
// that a later retry might resolve differently.
//
// internal/reconcile's replay uses this, exactly as it uses
// errors.Is(err, store.ErrMonotonicViolation) for a state journal entry,
// to decide whether a journalled mark is safe to retire: a definitive
// outcome is retired (design.md, "an entry is removed from the journal
// only once the store has accepted or explicitly refused it"), and
// anything else is left for the next replay.
func IsDefinitiveMarkOutcome(err error) bool {
	if err == nil {
		return true
	}
	var unknownStage *stages.ErrUnknownStage
	switch {
	case errors.As(err, &unknownStage):
		// The caller (or a journalled mark replaying against a README
		// table that has since changed) named a stage this build of
		// myflowd does not document. That will not resolve itself by
		// retrying the identical entry.
		return true
	case errors.Is(err, ErrInvalidSessionToken):
		// A caller mistake -- an empty or shell-substitution sessionToken -- that
		// retrying the identical journalled entry can never fix, exactly
		// like an undocumented stage name above.
		return true
	case errors.Is(err, ErrNoOpenStageRun):
		return true
	case errors.Is(err, store.ErrInvalidMainCheckoutPath):
		return true
	case errors.Is(err, store.ErrInvalidState):
		return true
	default:
		// Everything else -- ErrTooManyAttemptCollisions (transient
		// contention), a dropped connection, a context cancellation, an
		// unmapped store error -- is not known to be safe to retire.
		return false
	}
}

// decodeJSONBody reads r's body (capped at maxRequestBodyBytes) and
// decodes it into v, rejecting unknown fields -- the same discipline
// DecodeChangeBody already applies to a change's wire shape.
func decodeJSONBody(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("malformed request body: %w", err)
	}
	return nil
}

// parseMarkTime parses s as RFC 3339, defaulting to now (UTC) when s is
// empty -- a mark's CLI caller always supplies an instant in practice, but
// an empty value is treated as "now" rather than a 400, since a caller
// timing a mark from its own clock and one timing it from the request's
// arrival differ by, at most, the fallback timeout's own short budget.
func parseMarkTime(s string) (time.Time, error) {
	if s == "" {
		return time.Now().UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q is not RFC 3339: %w", s, err)
	}
	return t, nil
}

// begin serves POST /api/v1/stages/begin: decode, validate shape, delegate
// to ApplyBeginStageMark.
func (h *stageHandler) begin(w http.ResponseWriter, r *http.Request) {
	var req stageBeginRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.ProjectKey == "" || req.ChangeName == "" || req.Command == "" || req.Stage == "" || req.Harness == "" {
		writeError(w, http.StatusBadRequest, "projectKey, changeName, command, stage and harness are all required")
		return
	}
	startedAt, err := parseMarkTime(req.StartedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "startedAt: "+err.Error())
		return
	}

	result, err := ApplyBeginStageMark(r.Context(), h.store, h.logger, BeginStageMark{
		ProjectKey:       req.ProjectKey,
		MainCheckoutPath: req.MainCheckoutPath,
		ChangeName:       req.ChangeName,
		RepoRoot:         req.RepoRoot,
		Harness:          req.Harness,
		SessionID:        req.SessionID,
		SessionToken:     req.SessionToken,
		Command:          req.Command,
		Stage:            req.Stage,
		StartedAt:        startedAt,
	})
	if err != nil {
		var unknownStage *stages.ErrUnknownStage
		switch {
		case errors.As(err, &unknownStage):
			writeErrorWithCode(w, http.StatusBadRequest, CodeUndocumentedStage, err.Error())
			return
		case errors.Is(err, ErrInvalidSessionToken):
			// A caller mistake, not a store failure -- mapStoreError's
			// generic default (500 "internal error") would swallow the
			// reason this is rejected, exactly the failure mode
			// CodeUndocumentedStage exists to avoid for a stage name.
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("begin stage %s/%s for %s/%s", req.Command, req.Stage, req.ProjectKey, req.ChangeName), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, stageRunResponse{StageRunID: result.StageRunID, Attempt: result.Attempt})
}

// ApplyBeginStageMark records the start of one stage run against ss:
// exactly what a live begin mark and a replayed one both need.
//
// It validates mark.Stage against internal/stages' documented table before
// touching the store -- defence in depth alongside the CLI's own identical
// check (cmd/myflow/stage.go), since design.md's non-negotiable
// requirement is that no undocumented stage name is ever recorded, not
// merely that the CLI happens to be the only caller today. A journalled
// mark replayed after README.md's table has since changed is caught here
// too, by the same check, rather than a separate one.
//
// mark.SessionToken is validated the same way, by validateSessionTokenShape: empty, or
// carrying a shell-substitution shape, is rejected as ErrInvalidSessionToken
// before the store is ever touched -- again mirroring the CLI's own
// identical check.
//
// A mark for a change the store has never heard of is not dropped: when
// BeginStage reports store.ErrChangeNotFound, this bootstraps a synthetic
// change row -- state STARTED, updated by syntheticChangeUpdatedBy -- and
// retries once. This is deliberately the same PutChange bootstrap path a
// genuine `state set` uses for a brand-new project (store.PutChange's own
// doc comment), so a synthetic row and a real one are indistinguishable to
// every other query except by UpdatedBy -- exactly what makes "a defect
// worth seeing in the data" (design.md) visible rather than a special
// case a reader has to know to look for.
func ApplyBeginStageMark(ctx context.Context, ss StageStore, logger *slog.Logger, mark BeginStageMark) (StageMarkResult, error) {
	if err := stages.Validate(stages.Command(mark.Command), mark.Stage); err != nil {
		return StageMarkResult{}, err
	}
	if err := validateSessionTokenShape(mark.SessionToken); err != nil {
		return StageMarkResult{}, err
	}

	var sessionToken *string
	if mark.SessionToken != "" {
		sessionToken = &mark.SessionToken
	}
	in := store.BeginStageInput{
		ProjectKey:   mark.ProjectKey,
		ChangeName:   mark.ChangeName,
		RepoRoot:     mark.RepoRoot,
		Harness:      mark.Harness,
		SessionID:    mark.SessionID,
		SessionToken: sessionToken,
		Command:      mark.Command,
		Stage:        mark.Stage,
		StartedAt:    mark.StartedAt,
	}

	run, err := ss.BeginStage(ctx, in)
	if errors.Is(err, store.ErrChangeNotFound) {
		bootstrap := store.Change{
			ProjectKey:       mark.ProjectKey,
			MainCheckoutPath: mark.MainCheckoutPath,
			Name:             mark.ChangeName,
			State:            store.StateStarted,
			UpdatedAt:        time.Now().UTC(),
			UpdatedBy:        syntheticChangeUpdatedBy,
		}
		if bootstrapErr := ss.PutChange(ctx, bootstrap); bootstrapErr != nil &&
			!errors.Is(bootstrapErr, store.ErrMonotonicViolation) {
			// A monotonic refusal here means a real change row landed
			// (via a genuine `state set`) between BeginStage's failed
			// lookup and this bootstrap attempt -- a benign race, not a
			// reason to fail the mark: the retry below now finds the
			// change BeginStage originally could not.
			if logger != nil {
				logger.Error("api: bootstrap synthetic change for stage mark failed",
					"project", mark.ProjectKey, "change", mark.ChangeName, "error", bootstrapErr)
			}
			return StageMarkResult{}, bootstrapErr
		}
		run, err = ss.BeginStage(ctx, in)
	}
	if err != nil {
		return StageMarkResult{}, err
	}
	return StageMarkResult{StageRunID: run.ID, Attempt: run.Attempt}, nil
}

// end serves POST /api/v1/stages/end: decode, validate shape, delegate to
// ApplyEndStageMark.
func (h *stageHandler) end(w http.ResponseWriter, r *http.Request) {
	var req stageEndRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.ProjectKey == "" || req.ChangeName == "" || req.Command == "" || req.Stage == "" || req.Outcome == "" {
		writeError(w, http.StatusBadRequest, "projectKey, changeName, command, stage and outcome are all required")
		return
	}
	endedAt, err := parseMarkTime(req.EndedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "endedAt: "+err.Error())
		return
	}

	result, err := ApplyEndStageMark(r.Context(), h.store, EndStageMark{
		ProjectKey: req.ProjectKey,
		ChangeName: req.ChangeName,
		Command:    req.Command,
		Stage:      req.Stage,
		EndedAt:    endedAt,
		Outcome:    req.Outcome,
		Metrics:    req.Metrics,
	})
	if err != nil {
		var unknownStage *stages.ErrUnknownStage
		switch {
		case errors.As(err, &unknownStage):
			writeErrorWithCode(w, http.StatusBadRequest, CodeUndocumentedStage, err.Error())
		case errors.Is(err, ErrNoOpenStageRun):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			status, msg := mapStoreError(h.logger, fmt.Sprintf("end stage %s/%s for %s/%s", req.Command, req.Stage, req.ProjectKey, req.ChangeName), err)
			writeError(w, status, msg)
		}
		return
	}
	writeJSON(w, http.StatusOK, stageRunResponse{StageRunID: result.StageRunID, Attempt: result.Attempt})
}

// ApplyEndStageMark records the end of the currently open stage run
// matching mark's identity: exactly what a live end mark and a replayed
// one both need.
//
// mark carries no harness -- deliberately. An earlier version of this
// task (task 10's own first commit) had `stage end` re-resolve
// $MYFLOW_HARNESS/-harness itself and send tokens_available: false
// whenever that value was not "claude-code". Post-commit review (finding
// F1) found this unsound: nothing ties an end mark's own harness
// resolution to the harness BeginStage actually recorded for this run, so
// design.md's own canonical example --
// `myflow stage end --change ... --outcome completed`, no -harness given
// -- would mark a genuine claude-code run's tokens unavailable the moment
// $MYFLOW_HARNESS was unset in the ending shell, producing exactly the
// false-negative this task exists to prevent: an internally inconsistent
// row carrying both real, harvested tokens.* figures and a flag claiming
// they don't exist.
//
// The fix is to never ask the end mark at all: this reads openRun.Harness
// below -- the value recorded once, at `stage begin`, and never
// afterwards mutated -- and merges tokens_available: false only when that
// recorded harness is not claudeCodeHarness. There is therefore no
// "contradicting harness" case for an end mark to resolve: the row's own
// history is the only source consulted, and an end mark has no field left
// that could disagree with it.
//
// Unlike EndStage/MergeMetrics in internal/store, this never learns a
// stage run id from its caller -- design.md's own example (`myflow stage
// end --change ... --outcome completed`) names none either -- so it
// resolves the run itself: the most recent (highest attempt) stage run
// for mark's project, change, command and stage that has no end mark yet,
// via store's own typed query surface (QueryStageRuns), never a query
// this package assembles itself. See findOpenStageRun's own doc comment
// for why at most one such run can exist for the triple this resolves
// against, so "highest attempt, limit 1" never has to arbitrate between
// two genuinely open candidates.
//
// If no such run exists -- already ended by an earlier `stage end`
// (including a duplicate replay of this exact journalled entry), or
// closed by the sweeper (task 10) as abandoned in the meantime -- this
// returns ErrNoOpenStageRun. That is treated as a *definitive* refusal
// (IsDefinitiveMarkOutcome), not a transient failure: retrying the
// identical entry can never make a closed run open again, so the
// alternative -- leaving it in the journal forever -- would only grow the
// journal without ever making progress. This mirrors exactly how a state
// journal entry's ErrMonotonicViolation covers both "genuinely superseded"
// and "benign duplicate retry" as one retireable outcome (see
// internal/api/changes.go's put doc comment): an end mark that can no
// longer attach to anything is unresolvable in the same sense, and is
// retired the same way, loudly logged by internal/reconcile rather than
// silently dropped.
func ApplyEndStageMark(ctx context.Context, ss StageStore, mark EndStageMark) (StageMarkResult, error) {
	if err := stages.Validate(stages.Command(mark.Command), mark.Stage); err != nil {
		return StageMarkResult{}, err
	}

	openRun, err := findOpenStageRun(ctx, ss, mark.ProjectKey, mark.ChangeName, mark.Command, mark.Stage)
	if err != nil {
		return StageMarkResult{}, err
	}
	if openRun == nil {
		return StageMarkResult{}, fmt.Errorf("%w: %s/%s %s/%s", ErrNoOpenStageRun, mark.ProjectKey, mark.ChangeName, mark.Command, mark.Stage)
	}

	if err := ss.EndStage(ctx, openRun.ID, mark.EndedAt, mark.Outcome); err != nil {
		return StageMarkResult{}, err
	}

	metrics := mark.Metrics
	if openRun.Harness != claudeCodeHarness {
		merged, err := withTokensUnavailable(metrics)
		if err != nil {
			return StageMarkResult{}, fmt.Errorf("api: mark tokens unavailable for stage run %d: %w", openRun.ID, err)
		}
		metrics = merged
	}
	if len(metrics) > 0 {
		if err := ss.MergeMetrics(ctx, openRun.ID, metrics); err != nil {
			return StageMarkResult{}, err
		}
	}
	return StageMarkResult{StageRunID: openRun.ID, Attempt: openRun.Attempt}, nil
}

// withTokensUnavailable adds "tokens_available": false to metrics -- a
// caller-supplied patch, possibly nil/empty -- as a JSON boolean, never a
// numeric 0: design.md's absence-is-not-a-value rule applies to this flag
// exactly as it does to every other metric. The two are merged here, in
// Go, before either reaches the store, so ApplyEndStageMark still issues
// at most one MergeMetrics call per end mark, rather than one for the
// caller's own metrics and a second, separately racing one for this flag.
func withTokensUnavailable(metrics json.RawMessage) (json.RawMessage, error) {
	patch := map[string]json.RawMessage{}
	if len(metrics) > 0 {
		if err := json.Unmarshal(metrics, &patch); err != nil {
			return nil, fmt.Errorf("decode metrics patch: %w", err)
		}
	}
	patch["tokens_available"] = json.RawMessage("false")
	out, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("encode metrics patch: %w", err)
	}
	return out, nil
}

// findOpenStageRun returns the most recent (highest attempt) stage run
// matching projectKey/changeName/command/stage with no end mark yet, or
// nil if none is open.
//
// In the ordinary calling pattern -- one `stage begin` per stage per
// pipeline run, and a fix re-run only starting after the previous
// attempt's `stage end` has already landed (task 15 wires that sequence
// into each skill) -- at most one attempt for a given triple is open at
// once, so "highest attempt, limit 1" is really "the only open one".
//
// That invariant is *not* enforced anywhere, though, and it can be broken
// by exactly one path: BeginStage always allocates a fresh attempt on
// every call (store/stageruns.go's own doc comment -- it retries on a
// concurrent *attempt-number* collision, but has no concept of "a begin
// for this triple is already open" to de-duplicate against). If a live
// `stage begin` HTTP call times out on the client side *after* the store
// has already committed the insert -- a real, if narrow, race, the same
// shape as any "write succeeded, ack lost" timeout -- the CLI takes the
// fallback path and journals the same begin. internal/reconcile later
// replays that journal entry through this same ApplyBeginStageMark, which
// opens a *second* attempt for the triple, because BeginStage has no way
// to recognise the first one as "this already happened".
//
// When that has occurred, "highest attempt, limit 1" is deterministic and
// always picks a real, currently-open row -- ORDER BY attempt DESC is a
// total order over a UNIQUE column, so there is never an ambiguous tie to
// arbitrate -- and it picks the *newer* attempt, which is the one a
// `stage end` issued after the replay is actually describing. It does not
// pick wrong, and it never blocks or errors because more than one row
// matched. What it does not do is reconcile the two attempts: the earlier,
// now-orphaned one is left open until task 10's sweeper closes it as
// abandoned once its session goes silent. That produces one spurious
// "abandoned" attempt in the data -- a statistics-accuracy cost (an
// inflated attempt/rework count for that stage, caused by network
// flakiness rather than a genuine re-run), not a correctness hazard: no
// data is lost, no state moves backwards, and no request ever fails
// because of it. De-duplicating begin marks (an idempotency key BeginStage
// could check) would close this gap properly; it is not attempted here --
// it needs its own schema and tests, and is a candidate for task 9 or 10
// rather than this one.
func findOpenStageRun(ctx context.Context, ss StageStore, projectKey, changeName, command, stage string) (*store.StageRun, error) {
	runs, _, err := ss.QueryStageRuns(ctx, store.Query{
		Filters: []store.Filter{
			{Field: "project", Op: store.OpEq, Value: projectKey},
			{Field: "name", Op: store.OpEq, Value: changeName},
			{Field: "command", Op: store.OpEq, Value: command},
			{Field: "stage", Op: store.OpEq, Value: stage},
			{Field: "ended_at", Op: store.OpNull},
		},
		Sort:  []store.SortKey{{Field: "attempt", Desc: true}},
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return nil, nil
	}
	return &runs[0], nil
}
