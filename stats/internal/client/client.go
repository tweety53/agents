// Package client is the CLI's only means of reaching flowd over HTTP. It
// exists to answer exactly one question for its caller on every request:
// did the store answer, or should the caller fall back? Everything that is
// not the store answering -- a dead port, a timeout, a malformed body, any
// non-2xx status the store's own API does not use as a legitimate refusal
// -- collapses to the single ErrUnavailable signal, because the CLI's
// fallback decision does not distinguish among them (per design.md,
// "Availability and reconciliation", and the task 5 plan's "on any store
// failure whatsoever").
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// maxResponseBytes caps a response body this client will read, matching
// the daemon's own request-body cap (internal/api.maxRequestBodyBytes) --
// a response to a change record is realistically the same small size.
const maxResponseBytes = 1 << 20

// daemonHeaderName and daemonHeaderValue must match internal/api's
// DaemonHeader and DaemonHeaderValue exactly. They are kept as literal
// constants here, rather than imported, for the same reason
// cmd/flow/state.go names flowd's default address again instead of
// importing internal/config: per design.md's "Boundaries", the CLI knows
// only HTTP, never the daemon's internal packages, so the two sides are
// allowed to agree on a string constant without sharing a package.
//
// No response is trusted as a store answer unless it carries this header
// with this exact value -- see internal/api.DaemonHeader's doc comment for
// why: a 409 or a 404 from some other process squatting the port must fall
// back, not be reported as a genuine refusal or a genuine not-found.
const (
	daemonHeaderName  = "Flow-Daemon"
	daemonHeaderValue = "flowd/1"
)

// ErrUnavailable means the caller should treat the store as unreachable:
// a transport failure (dead port, timeout, DNS), a response the store's
// API never legitimately sends as an answer, or a response body that does
// not even parse as JSON. It is the CLI's fallback trigger.
var ErrUnavailable = errors.New("client: store unavailable")

// ErrRefused means the store was reached and correctly refused the write
// because it would move the record's state backwards (HTTP 409, mapping
// store.ErrMonotonicViolation -- see internal/api/changes.go's put doc
// comment for why a benign duplicate retry is folded into the same
// status). This is not a fallback trigger: the store answered, and the
// answer was "no".
var ErrRefused = errors.New("client: store refused the write: state would move backwards")

// ErrNotFound means the store was reached and correctly reports no record
// for the requested project and name (HTTP 404). Like ErrRefused, this is
// the store answering, not the store being unreachable -- so it is not
// wrapped in ErrUnavailable.
var ErrNotFound = errors.New("client: change not found")

// Client talks to flowd's state API. It never retries: exactly one
// request per call, bounded by ctx, so the caller controls how long a
// possibly-unreachable store is allowed to hold up the fallback decision.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New builds a Client against baseURL (e.g. "http://127.0.0.1:4173").
func New(baseURL string, httpClient *http.Client) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient}
}

func (c *Client) changeURL(project, name string) string {
	return c.baseURL + "/api/v1/changes/" + url.PathEscape(project) + "/" + url.PathEscape(name)
}

// GetChange fetches the raw JSON body the store holds for project/name --
// the same wire shape the state file contract describes, passed through
// unmodified so the caller can write it straight to disk or stdout. On any
// outcome other than the store answering 200 with a well-formed JSON body,
// or 404, it returns an error wrapping ErrUnavailable or ErrNotFound. A
// response of any status that does not carry the daemon header is treated
// as ErrUnavailable regardless of its status code -- see daemonHeaderName's
// doc comment.
func (c *Client) GetChange(ctx context.Context, project, name string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.changeURL(project, name), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}

	body, status, fromDaemon, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if !fromDaemon {
		return nil, fmt.Errorf("%w: response missing %s header -- not trusted as a store answer", ErrUnavailable, daemonHeaderName)
	}

	switch status {
	case http.StatusOK:
		if !json.Valid(body) {
			return nil, fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
		}
		return body, nil
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s/%s", ErrNotFound, project, name)
	default:
		return nil, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}

// PutChange sends body -- the whole change record, already carrying
// mainCheckoutPath where the caller resolved one -- as the record for
// project/name. A 200 response is success; a 409 is ErrRefused; every
// other outcome, transport failure included, is ErrUnavailable. A response
// of any status that does not carry the daemon header is treated as
// ErrUnavailable regardless of its status code -- see daemonHeaderName's
// doc comment.
func (c *Client) PutChange(ctx context.Context, project, name string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.changeURL(project, name), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")

	respBody, status, fromDaemon, err := c.do(req)
	if err != nil {
		return err
	}
	if !fromDaemon {
		return fmt.Errorf("%w: response missing %s header -- not trusted as a store answer", ErrUnavailable, daemonHeaderName)
	}

	switch status {
	case http.StatusOK:
		if !json.Valid(respBody) {
			return fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
		}
		return nil
	case http.StatusConflict:
		return fmt.Errorf("%w: %s", ErrRefused, string(respBody))
	default:
		return fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}

// settingsURL is flowd's harness-wide settings endpoint -- GET/PUT
// /api/v1/settings, per design.md's model-default-sonnet and
// review-panel-fixed-3 decisions. Unlike every other endpoint above, it
// carries no project or change identity: flow_settings holds exactly one
// row for the whole harness (internal/store/settings.go).
func (c *Client) settingsURL() string { return c.baseURL + "/api/v1/settings" }

// Settings is the wire shape GET/PUT /api/v1/settings exchange, copied
// from internal/api's settingsDTO field-for-field rather than imported --
// the same boundary reason StateBoardRow is copied rather than imported:
// the CLI knows only HTTP, never the daemon's internal packages.
type Settings struct {
	DefaultModel    string   `json:"defaultModel"`
	SelfReviewModel string   `json:"selfReviewModel"`
	Reviewers       []string `json:"reviewers"`
}

// ErrSettingsRejected means the store was reached and refused a
// PutSettings call with a 400 -- an unknown model or reviewer value, named
// in the wrapped error text exactly as internal/api/settings.go's put
// handler writes it. Like ErrRefused and ErrNotFound, this is the store
// answering, not the store being unreachable: a caller mistake with no
// fallback value to retry, per this task's own instruction.
var ErrSettingsRejected = errors.New("client: store rejected the settings write")

// GetSettings fetches the harness-wide settings record via GET
// /api/v1/settings. On any outcome other than the store answering 200 with
// a well-formed body carrying the daemon header, it returns an error
// wrapping ErrUnavailable -- the same fallback trigger every other read in
// this package uses.
func (c *Client) GetSettings(ctx context.Context) (Settings, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.settingsURL(), nil)
	if err != nil {
		return Settings{}, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}

	body, status, err := c.send(req)
	if err != nil {
		return Settings{}, err
	}
	if status != http.StatusOK {
		return Settings{}, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}

	var out Settings
	if err := json.Unmarshal(body, &out); err != nil {
		return Settings{}, fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
	}
	return out, nil
}

// PutSettings writes in as the harness-wide settings record via PUT
// /api/v1/settings and returns the record the store echoes back. A 200 is
// success; a 400 is ErrSettingsRejected, carrying the rejected value in
// its wrapped text exactly as the daemon reported it -- this is the store
// answering "no" to a caller mistake, not a reason to fall back. Every
// other outcome, transport failure included, is ErrUnavailable.
func (c *Client) PutSettings(ctx context.Context, in Settings) (Settings, error) {
	respBody, status, err := c.sendJSON(ctx, http.MethodPut, c.settingsURL(), in)
	if err != nil {
		return Settings{}, err
	}

	switch status {
	case http.StatusOK:
		var out Settings
		if err := json.Unmarshal(respBody, &out); err != nil {
			return Settings{}, fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
		}
		return out, nil
	case http.StatusBadRequest:
		var wire errorWireResponse
		if err := json.Unmarshal(respBody, &wire); err != nil || wire.Error == "" {
			return Settings{}, fmt.Errorf("%w: %s", ErrSettingsRejected, string(respBody))
		}
		return Settings{}, fmt.Errorf("%w: %s", ErrSettingsRejected, wire.Error)
	default:
		return Settings{}, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}

// stateBoardEpoch is the "from" instant ListStateBoard's underlying query
// asks for -- before any change this store could ever hold, since the
// store starts empty (state-file.md, "The on-disk file is written, never
// seeded"). It is
// named as a constant, not merely inlined, so the period stays a real,
// finite one rather than reading as "unbounded": GET
// /api/v1/stats/state-board requires both "from" and "to" (design.md,
// "every statistics view is period-parameterised... rather than filtered
// in the client"), and this is the caller-invisible choice that makes
// "list everything" satisfy that requirement instead of asking the daemon
// for a period it does not offer.
const stateBoardEpoch = "1970-01-01T00:00:00Z"

// stateBoardURL builds GET /api/v1/stats/state-board's URL for the period
// [stateBoardEpoch, now), restricted to project when project is not "".
//
// "to" is formatted with RFC3339Nano, not the whole-second RFC3339: the
// view's period is half-open ([from, to), boundaryConvention in
// internal/api/stats.go), so a "to" floored to the start of the current
// second would exclude a change updated earlier in that same second --
// reproduced by hand against a live daemon, where a change written and
// then immediately listed came back with an empty board until the next
// second ticked over. The daemon's own parser accepts either: Go's
// time.Parse recognises an optional fractional-second suffix even when
// the reference layout (time.RFC3339, used server-side) does not include
// one, so this is a client-side fix with no server-side counterpart
// needed.
func (c *Client) stateBoardURL(project string) string {
	q := url.Values{}
	q.Set("from", stateBoardEpoch)
	q.Set("to", time.Now().UTC().Format(time.RFC3339Nano))
	if project != "" {
		q.Set("project", project)
	}
	return c.baseURL + "/api/v1/stats/state-board?" + q.Encode()
}

// StateBoardRow is one row of the store's live state board -- just enough
// of GET /api/v1/stats/state-board's response (internal/api's
// stateBoardRowDTO) for a caller to enumerate change names and their
// current pipeline state. Copied rather than imported, for the same
// boundary reason daemonHeaderName/daemonHeaderValue are: the CLI knows
// only HTTP, never the daemon's internal packages. ProjectKey and
// NextCommand are deliberately not carried here: a caller of this method
// already knows which project it asked for, and the next-command mapping
// is `/flow-status`'s own table (skills/flow-status/SKILL.md), not
// something this package should compute a second copy of.
type StateBoardRow struct {
	Name      string `json:"name"`
	State     string `json:"state"`
	UpdatedAt string `json:"updatedAt"`
	UpdatedBy string `json:"updatedBy"`
}

type stateBoardWireResponse struct {
	Rows []StateBoardRow `json:"rows"`
}

// ListStateBoard fetches every change the store has recorded for project
// ("" for every project), via GET /api/v1/stats/state-board. On any
// outcome other than the store answering 200 with a well-formed body
// carrying the daemon header, it returns an error wrapping ErrUnavailable
// -- the same fallback trigger GetChange and PutChange already use, so a
// caller enumerating changes gets the header check, the ErrUnavailable
// classification and the request timeout for free, exactly as a caller
// reading or writing one record does. Never call this in a way that skips
// it (a hand-written HTTP request in a skill, for instance) -- that is
// precisely the gap design.md's daemon-owns-db decision rejects: "skills
// calling curl directly... puts inline HTTP and error handling into every
// contract file."
func (c *Client) ListStateBoard(ctx context.Context, project string) ([]StateBoardRow, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.stateBoardURL(project), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}

	body, status, fromDaemon, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if !fromDaemon {
		return nil, fmt.Errorf("%w: response missing %s header -- not trusted as a store answer", ErrUnavailable, daemonHeaderName)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}

	var wire stateBoardWireResponse
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
	}
	return wire.Rows, nil
}

// stagesBeginURL and stagesEndURL are flowd's stage-mark endpoints --
// POST /api/v1/stages/begin and POST /api/v1/stages/end, per design.md's
// "API" section. Unlike the change endpoints, a mark's identity (project,
// change, command, stage) travels in the request body rather than the URL
// path: design.md's own route shapes name no path parameters for these two
// routes, and the daemon resolves identity from the body exactly as it
// resolves a stage's open run for `stage end` (internal/api/stages.go).
func (c *Client) stagesBeginURL() string { return c.baseURL + "/api/v1/stages/begin" }
func (c *Client) stagesEndURL() string   { return c.baseURL + "/api/v1/stages/end" }

// BeginStageRequest is the wire shape `flow stage begin` sends. Fields
// mirror store.BeginStageInput; MainCheckoutPath bootstraps the project
// row exactly as PutChange's does, for the same reason a mark for an
// unknown change must still be stored (design.md, "A mark for an unknown
// change is stored, not dropped").
type BeginStageRequest struct {
	ProjectKey       string
	MainCheckoutPath string
	ChangeName       string
	RepoRoot         *string
	Harness          string
	SessionID        *string
	// SessionToken is the literal, unique correlator `stage begin` wrote into its
	// own command text (KAN-172, task 1) -- required, alongside Harness,
	// so a later harvest cycle has something to find in the calling
	// session's own transcript.
	SessionToken string
	Command      string
	Stage        string
	StartedAt    time.Time
}

// BeginStageResult identifies the stage run the daemon created, so a later
// `stage end` -- and a future harvester attributing usage to it -- can
// address it.
type BeginStageResult struct {
	StageRunID int64
	Attempt    int
}

// EndStageRequest is the wire shape `flow stage end` sends. It carries
// no numeric stage run id: the daemon resolves the open run for
// (ProjectKey, ChangeName, Command, Stage) itself, exactly as
// design.md's own `flow stage end --change ... --outcome completed`
// example never names one either.
type EndStageRequest struct {
	ProjectKey string
	ChangeName string
	Command    string
	Stage      string
	EndedAt    time.Time
	Outcome    string
	// Metrics, when non-nil, is merged into the stage run's metrics bag
	// (store.MergeMetrics) alongside recording the end instant and
	// outcome -- fix-rounds, panel-rounds and findings-by-severity travel
	// this way from `stage end`'s own flags.
	Metrics json.RawMessage
}

type beginStageWireRequest struct {
	ProjectKey       string  `json:"projectKey"`
	MainCheckoutPath string  `json:"mainCheckoutPath,omitempty"`
	ChangeName       string  `json:"changeName"`
	RepoRoot         *string `json:"repoRoot,omitempty"`
	Harness          string  `json:"harness"`
	SessionID        *string `json:"sessionId,omitempty"`
	SessionToken     string  `json:"sessionToken"`
	Command          string  `json:"command"`
	Stage            string  `json:"stage"`
	StartedAt        string  `json:"startedAt"`
}

type stageRunWireResponse struct {
	StageRunID int64 `json:"stageRunId"`
	Attempt    int   `json:"attempt"`
}

type endStageWireRequest struct {
	ProjectKey string          `json:"projectKey"`
	ChangeName string          `json:"changeName"`
	Command    string          `json:"command"`
	Stage      string          `json:"stage"`
	EndedAt    string          `json:"endedAt"`
	Outcome    string          `json:"outcome"`
	Metrics    json.RawMessage `json:"metrics,omitempty"`
}

// ErrUndocumentedStage means the store was reached and correctly refused
// the mark because it named a stage absent from README.md's documented
// Level 1 table (HTTP 400 from internal/api/stages.go's own
// internal/stages.Validate call). Like ErrRefused and ErrNotFound, this is
// the store answering, not the store being unreachable -- a caller mistake
// the CLI's own client-side stages.Validate call is expected to catch
// first, but the daemon enforces it too, and this client must not fold
// that refusal into the fallback trigger.
var ErrUndocumentedStage = errors.New("client: store refused the mark: undocumented stage name")

// ErrStageMarkRejected means the store was reached and rejected a stage
// mark with a 400 for a reason *other* than an undocumented stage name --
// a missing required field, a malformed body, or any other validation
// this daemon may reject a mark for. Like ErrUndocumentedStage, this is
// the store answering, not the store being unreachable.
//
// A 400 was, before this type existed, mapped unconditionally to
// ErrUndocumentedStage regardless of *why* the daemon rejected the
// request -- a status code alone cannot carry that distinction, the same
// lesson daemonHeaderName exists to teach for a different ambiguity (a
// look-alike server's status must not be trusted as this daemon's
// answer). See errorWireResponse and classifyBadRequestBody below for how
// the daemon's response body, not its status code, is what this client
// now actually acts on.
var ErrStageMarkRejected = errors.New("client: store rejected the mark")

// undocumentedStageErrorCode must match internal/api's CodeUndocumentedStage
// exactly. Kept as a literal constant here, rather than imported, for the
// same reason daemonHeaderName/daemonHeaderValue are: per design.md's
// "Boundaries", the CLI knows only HTTP, never the daemon's internal
// packages.
const undocumentedStageErrorCode = "undocumented_stage"

// errorWireResponse is the JSON body a stage-mark 400 carries -- the same
// shape internal/api's errorResponse writes, decoded here just far enough
// to read Code, never as a general-purpose error-body parser for every
// endpoint this client calls (GetChange/PutChange's error paths still
// report the raw body text, unchanged by this task).
type errorWireResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// classifyBadRequestBody turns a stage-mark endpoint's 400 response body
// into ErrUndocumentedStage (Code == undocumentedStageErrorCode) or
// ErrStageMarkRejected (anything else, code present or not) -- read from
// the body's Code field, never guessed from the status alone. A body that
// does not even parse as the expected JSON shape still resolves to
// ErrStageMarkRejected rather than ErrUnavailable: the store was reached
// and did answer with a 400, carrying the daemon header, so this is a
// real (if oddly-shaped) rejection, not a transport failure -- the CLI's
// fallback path is not the right response to a store that is unmistakably
// there and unmistakably saying no.
func classifyBadRequestBody(respBody []byte) error {
	var wire errorWireResponse
	if err := json.Unmarshal(respBody, &wire); err != nil {
		return fmt.Errorf("%w: %s", ErrStageMarkRejected, string(respBody))
	}
	if wire.Code == undocumentedStageErrorCode {
		return fmt.Errorf("%w: %s", ErrUndocumentedStage, wire.Error)
	}
	return fmt.Errorf("%w: %s", ErrStageMarkRejected, wire.Error)
}

// BeginStage records the start of one stage run. On any outcome other than
// the store answering 200, it returns an error wrapping ErrUnavailable,
// ErrUndocumentedStage, ErrStageMarkRejected, or ErrRefused -- see
// BeginStageRequest and this package's doc comment for what each means to
// the caller. Which of ErrUndocumentedStage and ErrStageMarkRejected a 400
// maps to is read from the response body's Code field
// (classifyBadRequestBody), never guessed from the 400 status alone. A
// response of any status that does not carry the daemon header is treated
// as ErrUnavailable regardless of its status code, exactly as GetChange
// and PutChange already do.
func (c *Client) BeginStage(ctx context.Context, in BeginStageRequest) (BeginStageResult, error) {
	wire := beginStageWireRequest{
		ProjectKey:       in.ProjectKey,
		MainCheckoutPath: in.MainCheckoutPath,
		ChangeName:       in.ChangeName,
		RepoRoot:         in.RepoRoot,
		Harness:          in.Harness,
		SessionID:        in.SessionID,
		SessionToken:     in.SessionToken,
		Command:          in.Command,
		Stage:            in.Stage,
		StartedAt:        in.StartedAt.UTC().Format(time.RFC3339Nano),
	}
	respBody, status, err := c.sendJSON(ctx, http.MethodPost, c.stagesBeginURL(), wire)
	if err != nil {
		return BeginStageResult{}, err
	}

	switch status {
	case http.StatusOK:
		var resp stageRunWireResponse
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return BeginStageResult{}, fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
		}
		return BeginStageResult{StageRunID: resp.StageRunID, Attempt: resp.Attempt}, nil
	case http.StatusBadRequest:
		return BeginStageResult{}, classifyBadRequestBody(respBody)
	case http.StatusConflict:
		return BeginStageResult{}, fmt.Errorf("%w: %s", ErrRefused, string(respBody))
	default:
		return BeginStageResult{}, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}

// EndStage records the end of the currently open stage run matching
// in.ProjectKey/ChangeName/Command/Stage. On any outcome other than the
// store answering 200, it returns an error wrapping ErrUnavailable,
// ErrUndocumentedStage, ErrStageMarkRejected, ErrRefused, or ErrNotFound
// (no open stage run matches -- HTTP 404). See BeginStage's own doc
// comment for how a 400 is told apart between the first two.
func (c *Client) EndStage(ctx context.Context, in EndStageRequest) (BeginStageResult, error) {
	wire := endStageWireRequest{
		ProjectKey: in.ProjectKey,
		ChangeName: in.ChangeName,
		Command:    in.Command,
		Stage:      in.Stage,
		EndedAt:    in.EndedAt.UTC().Format(time.RFC3339Nano),
		Outcome:    in.Outcome,
		Metrics:    in.Metrics,
	}
	respBody, status, err := c.sendJSON(ctx, http.MethodPost, c.stagesEndURL(), wire)
	if err != nil {
		return BeginStageResult{}, err
	}

	switch status {
	case http.StatusOK:
		var resp stageRunWireResponse
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return BeginStageResult{}, fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
		}
		return BeginStageResult{StageRunID: resp.StageRunID, Attempt: resp.Attempt}, nil
	case http.StatusBadRequest:
		return BeginStageResult{}, classifyBadRequestBody(respBody)
	case http.StatusNotFound:
		return BeginStageResult{}, fmt.Errorf("%w: %s", ErrNotFound, string(respBody))
	case http.StatusConflict:
		return BeginStageResult{}, fmt.Errorf("%w: %s", ErrRefused, string(respBody))
	default:
		return BeginStageResult{}, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}

// sendJSON marshals body as req's JSON payload, builds the request for
// method/url and executes it through send.
//
// It is the prologue every bodied call to the daemon shares -- marshal,
// build, set Content-Type, execute, refuse an answer that is not the
// daemon's -- stated once. BeginStage, EndStage and every record write hand-
// rolled a copy of it, and the copies had already drifted: the record path
// and the stage path did not agree on what a 409 meant, a difference that
// was invisible precisely because the two prologues were read as separate
// code rather than as one contract with two callers.
//
// What it deliberately does NOT share is the status mapping. A stage mark's
// vocabulary (ErrUndocumentedStage, ErrRefused) and a record write's
// (ErrRecordRejected) are genuinely different answers to genuinely
// different questions, so each caller switches on the status it is given
// rather than being handed a classifier parameterised by an error set only
// one of them uses.
func (c *Client) sendJSON(ctx context.Context, method, url string, body any) ([]byte, int, error) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: encode request: %v", ErrUnavailable, err)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, 0, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.send(req)
}

// send executes req through do and refuses any answer that does not carry
// the daemon header, whatever its status -- a look-alike server on the
// configured port is not a store, and reading its answer as one is the
// failure the header exists to prevent.
func (c *Client) send(req *http.Request) ([]byte, int, error) {
	respBody, status, fromDaemon, err := c.do(req)
	if err != nil {
		return nil, 0, err
	}
	if !fromDaemon {
		return nil, 0, fmt.Errorf("%w: response missing %s header -- not trusted as a store answer", ErrUnavailable, daemonHeaderName)
	}
	return respBody, status, nil
}

// do executes req and returns its body, status and whether the response
// carried the daemon header, or an error already wrapping ErrUnavailable
// for every transport-level failure: connection refused, DNS failure, a
// context deadline (timeout or cancellation), or a body that could not be
// fully read.
func (c *Client) do(req *http.Request) (body []byte, status int, fromDaemon bool, err error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, false, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()

	fromDaemon = resp.Header.Get(daemonHeaderName) == daemonHeaderValue

	body, err = io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, 0, false, fmt.Errorf("%w: read response body: %v", ErrUnavailable, err)
	}
	return body, resp.StatusCode, fromDaemon, nil
}

// --- run records ----------------------------------------------------------

// ErrRecordRejected means the daemon answered and refused a record write
// because the call itself was wrong -- a missing required field, a body
// that did not decode. Like ErrRefused and ErrNotFound, this is the store
// answering, so it is deliberately not wrapped in ErrUnavailable: a
// journalled replay of the identical call would be refused identically,
// and the never-block guarantee covers a store that cannot be reached, not
// a call that cannot be right.
var ErrRecordRejected = errors.New("client: store rejected the record write")

func (c *Client) recordsURL(project, change string) string {
	return c.baseURL + "/api/v1/records/" + url.PathEscape(project) + "/" + url.PathEscape(change)
}

// RecordDispatch opens one subagent dispatch's record for project/change --
// the first of the two calls that record a dispatch, sent as it starts
// rather than as it ends -- and returns the row the daemon stored, carrying
// the id and the seq it allocated -- neither of which the caller could know, and the seq being
// what a rendered record and a finding's DispatchSeq name the dispatch by.
//
// On any outcome other than the daemon answering 201 it returns an error
// wrapping ErrRecordRejected (400: the call is wrong), ErrNotFound (404:
// no such change), or ErrUnavailable for everything else, transport
// failures included. A response of any status that does not carry the
// daemon header is ErrUnavailable regardless of its status code, exactly
// as every other method here treats one -- see daemonHeaderName's doc
// comment.
func (c *Client) RecordDispatch(ctx context.Context, project, change string, in records.Dispatch) (records.Dispatch, error) {
	var out records.Dispatch
	_, err := c.writeRecord(ctx, http.MethodPost, c.recordsURL(project, change)+"/dispatches", in,
		map[int]bool{http.StatusCreated: true}, &out)
	if err != nil {
		return records.Dispatch{}, err
	}
	return out, nil
}

// EndDispatch closes the dispatch project/change recorded under
// in.SessionToken and in.Key, writing its commit, its outcome and its end
// instant, and returns the row as it now stands.
//
// It is the second of the two calls that record one dispatch. The first
// exists so the row -- and therefore the attribution window -- is there
// while the dispatch is still running; this one exists so the window closes
// (records.DispatchEnd carries the whole reasoning).
//
// A key the change holds no dispatch under is ErrNotFound (404). That is
// the one record answer the CLI journals rather than reports, because it is
// genuinely retryable here: the `begin` this call closes may still be
// sitting in the same journal ahead of it. See RecordDispatch's doc comment
// for how every other outcome is classified.
func (c *Client) EndDispatch(ctx context.Context, project, change string, in records.DispatchEnd) (records.Dispatch, error) {
	var out records.Dispatch
	_, err := c.writeRecord(ctx, http.MethodPost, c.recordsURL(project, change)+"/dispatches/end", in,
		map[int]bool{http.StatusOK: true}, &out)
	if err != nil {
		return records.Dispatch{}, err
	}
	return out, nil
}

// RecordFinding records one review-panel finding for project/change, or
// replaces the one already recorded under the same ref, and returns the
// row the daemon stored together with which of the two it did.
//
// Both of the daemon's success codes are success here: 201 when the write
// inserted and 200 when it replaced, and created is that distinction --
// the daemon's answer to "did this ref exist already", which no caller can
// work out for itself. `flow record finding` prints "recorded: F<n>" on
// an insert and "updated: F<n>" on a replace, so a panel run can tell a
// finding it has just raised from one a fix round restated. See
// RecordDispatch's doc comment for how every other outcome is classified.
func (c *Client) RecordFinding(ctx context.Context, project, change string, in records.Finding) (out records.Finding, created bool, err error) {
	status, err := c.writeRecord(ctx, http.MethodPost, c.recordsURL(project, change)+"/findings", in,
		map[int]bool{http.StatusCreated: true, http.StatusOK: true}, &out)
	if err != nil {
		return records.Finding{}, false, err
	}
	return out, status == http.StatusCreated, nil
}

// SetFindingStatus rewrites one finding's status and nothing else -- the
// whole of what a fix round changes about a finding it has resolved. A ref
// the change holds no finding under is ErrNotFound, never ErrUnavailable:
// the daemon answered, and journalling a mistyped ref would queue a replay
// that can never succeed. A `deferred <reason>` status against a finding
// whose severity is not Minor answers 409 and is ErrRecordRejected by
// classifyRecordResponse's own doc comment -- the store having been reached
// and having refused, the same reason 404 is not journalled either. See
// RecordDispatch's doc comment for how every other outcome is classified.
func (c *Client) SetFindingStatus(ctx context.Context, project, change, ref, status string) error {
	body := findingStatusWireRequest{Status: status}
	_, err := c.writeRecord(ctx, http.MethodPatch,
		c.recordsURL(project, change)+"/findings/"+url.PathEscape(ref), body,
		map[int]bool{http.StatusNoContent: true}, nil)
	return err
}

// GetRunRecord fetches project/change's whole derived record: its
// dispatches in seq order and its findings in the order their refs' digits
// spell, which is the order the renderer reads them in. A change that
// exists and holds no rows is an empty record and no error; only a change
// the store has never heard of is ErrNotFound, so "this change has no
// findings" can never be read as "this change does not exist".
func (c *Client) GetRunRecord(ctx context.Context, project, change string) (records.Run, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.recordsURL(project, change), nil)
	if err != nil {
		return records.Run{}, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}

	respBody, status, err := c.send(req)
	if err != nil {
		return records.Run{}, err
	}
	var out records.Run
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return records.Run{}, err
	}
	return out, nil
}

// GetCostStatus fetches project/change's cost-status: how many of its
// dispatches carry no cost figure yet, and why. Unlike `flow record
// journal-count`, this genuinely contacts the store -- only the store can
// answer a question about what it holds -- so it is classified exactly as
// GetRunRecord's own read is, ErrNotFound (404, no such change) or
// ErrUnavailable for everything else, transport failures and a response
// missing the daemon header included.
func (c *Client) GetCostStatus(ctx context.Context, project, change string) (records.CostStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.recordsURL(project, change)+"/cost-status", nil)
	if err != nil {
		return records.CostStatus{}, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}

	respBody, status, err := c.send(req)
	if err != nil {
		return records.CostStatus{}, err
	}
	var out records.CostStatus
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return records.CostStatus{}, err
	}
	return out, nil
}

// RecordDecision records one run's dynamic decision for project/change, or
// replaces the one already recorded under the same session token, and
// returns the row the daemon stored together with which of the two it did.
// See RecordFinding's own doc comment for what the created result means and
// why a client needs it, and RecordDispatch's for how every outcome other
// than the daemon's two success codes is classified.
func (c *Client) RecordDecision(ctx context.Context, project, change string, in records.Decision) (out records.Decision, created bool, err error) {
	status, err := c.writeRecord(ctx, http.MethodPost, c.recordsURL(project, change)+"/decisions", in,
		map[int]bool{http.StatusCreated: true, http.StatusOK: true}, &out)
	if err != nil {
		return records.Decision{}, false, err
	}
	return out, status == http.StatusCreated, nil
}

// ListDecisions fetches project/change's recorded decisions, newest first
// -- what a resumed run reads to recover a stopped run's own choice. See
// GetRunRecord's own doc comment for how every outcome other than 200 is
// classified.
func (c *Client) ListDecisions(ctx context.Context, project, change string) ([]records.Decision, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.recordsURL(project, change)+"/decisions", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	respBody, status, err := c.send(req)
	if err != nil {
		return nil, err
	}
	var out []records.Decision
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RecordPass records one pass-log entry of the review panel's record
// (KAN-331) -- a pass-by-pass metadata line the parent records as it
// arises. Every insert is a new row, so the daemon always answers 201 and
// no created result is reported; see RecordDecision for the classification
// of every outcome other than the daemon's success codes.
func (c *Client) RecordPass(ctx context.Context, project, change string, in records.Pass) (records.Pass, error) {
	var out records.Pass
	if _, err := c.writeRecord(ctx, http.MethodPost, c.recordsURL(project, change)+"/passes", in,
		map[int]bool{http.StatusCreated: true}, &out); err != nil {
		return records.Pass{}, err
	}
	return out, nil
}

// RecordMutation records one fix-mutation: line of the fix round's
// mutation proof (KAN-331). Append-only like RecordPass.
func (c *Client) RecordMutation(ctx context.Context, project, change string, in records.Mutation) (records.Mutation, error) {
	var out records.Mutation
	if _, err := c.writeRecord(ctx, http.MethodPost, c.recordsURL(project, change)+"/mutations", in,
		map[int]bool{http.StatusCreated: true}, &out); err != nil {
		return records.Mutation{}, err
	}
	return out, nil
}

// verdictsURL and incidentsURL are KAN-451's guard-log endpoints -- project
// scoped rather than change scoped, per design.md §4, since a verdict list
// spans every change on the project and an incident can be recorded
// against a project with no change in flight at all.
func (c *Client) verdictsURL(project string) string {
	return c.baseURL + "/api/v1/verdicts/" + url.PathEscape(project)
}

func (c *Client) incidentsURL(project string) string {
	return c.baseURL + "/api/v1/incidents/" + url.PathEscape(project)
}

// RecordVerdict records one guard's verdict for project/change and returns
// the row the daemon stored, carrying the id the store allocated. See
// RecordDispatch's doc comment for how every outcome other than a 201 is
// classified.
func (c *Client) RecordVerdict(ctx context.Context, project, change string, in records.Verdict) (records.Verdict, error) {
	var out records.Verdict
	_, err := c.writeRecord(ctx, http.MethodPost, c.recordsURL(project, change)+"/verdicts", in,
		map[int]bool{http.StatusCreated: true}, &out)
	if err != nil {
		return records.Verdict{}, err
	}
	return out, nil
}

// FlagVerdictFalsePositive flags the most recent verdict project/change
// recorded for in.Guard as a false positive, with the operator's reason,
// and returns the row as it now stands.
//
// A (change, guard) pair the daemon holds no verdict for is ErrNotFound,
// classified exactly as EndDispatch's unknown key is (see its own doc
// comment) -- the daemon answered, and the CLI journals a network failure
// while refusing a genuine "no verdict" outright, since no replay could
// ever make one exist.
func (c *Client) FlagVerdictFalsePositive(ctx context.Context, project, change string, in records.VerdictFlag) (records.Verdict, error) {
	var out records.Verdict
	_, err := c.writeRecord(ctx, http.MethodPost, c.recordsURL(project, change)+"/verdicts/false-positive", in,
		map[int]bool{http.StatusOK: true}, &out)
	if err != nil {
		return records.Verdict{}, err
	}
	return out, nil
}

// ListVerdicts fetches project's guard verdicts, newest first. guard == ""
// means every guard; falsePositiveOnly restricts to rows an operator has
// flagged. See GetRunRecord's own doc comment for how a non-200 outcome is
// classified -- this route names no change, so it can answer only
// ErrUnavailable, never ErrNotFound.
func (c *Client) ListVerdicts(ctx context.Context, project, guard string, falsePositiveOnly bool) ([]records.Verdict, error) {
	q := url.Values{}
	if guard != "" {
		q.Set("guard", guard)
	}
	if falsePositiveOnly {
		q.Set("falsePositive", "true")
	}
	u := c.verdictsURL(project)
	if len(q) > 0 {
		u += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	respBody, status, err := c.send(req)
	if err != nil {
		return nil, err
	}
	var out []records.Verdict
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RecordIncident records one per-project incident -- a guard, what went
// wrong, the recovery taken and how many minutes it cost -- and returns the
// row the daemon stored, carrying the id it allocated.
func (c *Client) RecordIncident(ctx context.Context, project string, in records.Incident) (records.Incident, error) {
	var out records.Incident
	_, err := c.writeRecord(ctx, http.MethodPost, c.incidentsURL(project), in,
		map[int]bool{http.StatusCreated: true}, &out)
	if err != nil {
		return records.Incident{}, err
	}
	return out, nil
}

// ListIncidents fetches project's incidents, newest first. See
// ListVerdicts' own doc comment for why this route can answer only
// ErrUnavailable, never ErrNotFound.
func (c *Client) ListIncidents(ctx context.Context, project string) ([]records.Incident, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.incidentsURL(project), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	respBody, status, err := c.send(req)
	if err != nil {
		return nil, err
	}
	var out []records.Incident
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- hazards (KAN-452) ----------------------------------------------------

// hazardsURL is the project-scoped hazards collection, the incidents pair's
// URL shape with "hazards" in place of "incidents".
func (c *Client) hazardsURL(project string) string {
	return c.baseURL + "/api/v1/hazards/" + url.PathEscape(project)
}

// AddHazard records one proactive, per-project hazard. A 409 -- the
// unique constraint refusing a re-add of a name the project already holds
// -- comes back as ErrRecordRejected, which classifyRecordWrite reports and
// never journals: a replay of it would be refused identically.
func (c *Client) AddHazard(ctx context.Context, project string, h records.Hazard) (records.Hazard, error) {
	var out records.Hazard
	_, err := c.writeRecord(ctx, http.MethodPost, c.hazardsURL(project), h,
		map[int]bool{http.StatusCreated: true}, &out)
	if err != nil {
		return records.Hazard{}, err
	}
	return out, nil
}

// ListHazards reads a project's hazards, newest first. shape, when
// non-empty, filters to applies IN ('all', shape) exactly as the store
// documents; includeInactive lifts the active filter for the -all view.
func (c *Client) ListHazards(ctx context.Context, project, shape string, includeInactive bool) ([]records.Hazard, error) {
	listURL := c.hazardsURL(project)
	q := url.Values{}
	if shape != "" {
		q.Set("shape", shape)
	}
	if includeInactive {
		q.Set("all", "true")
	}
	if len(q) > 0 {
		listURL += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	respBody, status, err := c.send(req)
	if err != nil {
		return nil, err
	}
	var out []records.Hazard
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RetireHazard sets active=false on the named hazard and returns the row.
// A 404 -- no such name -- is ErrNotFound, reported and never journalled.
func (c *Client) RetireHazard(ctx context.Context, project, name string) (records.Hazard, error) {
	var out records.Hazard
	_, err := c.writeRecord(ctx, http.MethodPatch, c.hazardsURL(project)+"/"+url.PathEscape(name), nil,
		map[int]bool{http.StatusOK: true}, &out)
	if err != nil {
		return records.Hazard{}, err
	}
	return out, nil
}

// findingStatusWireRequest is the body PATCH .../findings/{ref} carries:
// the one column a fix round rewrites.
type findingStatusWireRequest struct {
	Status string `json:"status"`
}

// writeRecord sends body as JSON through sendJSON and classifies the answer
// through classifyRecordResponse. The record methods differ only in their
// URL, their method, which status codes mean success and what they decode
// into, so they share this one path rather than five copies of the same
// classification.
//
// It returns the status the daemon answered with, so a method whose success
// codes carry meaning of their own -- RecordFinding's 201-versus-200 -- can
// read it, and a method whose do not can discard it. The status is
// meaningful only when the returned error is nil.
func (c *Client) writeRecord(ctx context.Context, method, url string, body any, success map[int]bool, out any) (int, error) {
	respBody, status, err := c.sendJSON(ctx, method, url, body)
	if err != nil {
		return 0, err
	}
	return classifyRecordResponse(respBody, status, success, out)
}

// classifyRecordResponse maps one record route's answer onto this package's
// error vocabulary and decodes the body into out, which may be nil for a
// response with no body. It returns the status the daemon answered with
// alongside that mapping; the status is meaningful only when the returned
// error is nil.
//
// 400 and 409 are both ErrRecordRejected, and 409 is listed explicitly
// rather than left to the default. `PATCH .../findings/{ref}` answers 409
// for a `deferred <reason>` status set against a finding whose severity is
// not Minor (store.ErrDeferredNotMinor); the default is ErrUnavailable,
// which the CLI journals for a later replay -- and a conflict is the store
// having been reached and having refused, exactly the answer that must
// never be queued for a replay that would be refused identically forever.
// The stage path maps its own 409 to ErrRefused; the two vocabularies
// differ, but neither may now resolve a conflict as a transport failure by
// omission.
func classifyRecordResponse(respBody []byte, status int, success map[int]bool, out any) (int, error) {
	switch {
	case success[status]:
		if out == nil {
			return status, nil
		}
		if err := json.Unmarshal(respBody, out); err != nil {
			return 0, fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
		}
		return status, nil
	case status == http.StatusBadRequest, status == http.StatusConflict:
		return 0, fmt.Errorf("%w: %s", ErrRecordRejected, string(respBody))
	case status == http.StatusNotFound:
		return 0, fmt.Errorf("%w: %s", ErrNotFound, string(respBody))
	default:
		return 0, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}
