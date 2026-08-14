// Package client is the CLI's only means of reaching myflowd over HTTP. It
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
)

// maxResponseBytes caps a response body this client will read, matching
// the daemon's own request-body cap (internal/api.maxRequestBodyBytes) --
// a response to a change record is realistically the same small size.
const maxResponseBytes = 1 << 20

// daemonHeaderName and daemonHeaderValue must match internal/api's
// DaemonHeader and DaemonHeaderValue exactly. They are kept as literal
// constants here, rather than imported, for the same reason
// cmd/myflow/state.go names myflowd's default address again instead of
// importing internal/config: per design.md's "Boundaries", the CLI knows
// only HTTP, never the daemon's internal packages, so the two sides are
// allowed to agree on a string constant without sharing a package.
//
// No response is trusted as a store answer unless it carries this header
// with this exact value -- see internal/api.DaemonHeader's doc comment for
// why: a 409 or a 404 from some other process squatting the port must fall
// back, not be reported as a genuine refusal or a genuine not-found.
const (
	daemonHeaderName  = "Myflow-Daemon"
	daemonHeaderValue = "myflowd/1"
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

// Client talks to myflowd's state API. It never retries: exactly one
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

// stateBoardEpoch is the "from" instant ListStateBoard's underlying query
// asks for -- before any change this store could ever hold, since the
// store starts empty (state-file.md, "The store starts empty"). It is
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
// is `/myflow-status`'s own table (skills/myflow-status/SKILL.md), not
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

// stagesBeginURL and stagesEndURL are myflowd's stage-mark endpoints --
// POST /api/v1/stages/begin and POST /api/v1/stages/end, per design.md's
// "API" section. Unlike the change endpoints, a mark's identity (project,
// change, command, stage) travels in the request body rather than the URL
// path: design.md's own route shapes name no path parameters for these two
// routes, and the daemon resolves identity from the body exactly as it
// resolves a stage's open run for `stage end` (internal/api/stages.go).
func (c *Client) stagesBeginURL() string { return c.baseURL + "/api/v1/stages/begin" }
func (c *Client) stagesEndURL() string   { return c.baseURL + "/api/v1/stages/end" }

// BeginStageRequest is the wire shape `myflow stage begin` sends. Fields
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

// EndStageRequest is the wire shape `myflow stage end` sends. It carries
// no numeric stage run id: the daemon resolves the open run for
// (ProjectKey, ChangeName, Command, Stage) itself, exactly as
// design.md's own `myflow stage end --change ... --outcome completed`
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
	reqBody, err := json.Marshal(wire)
	if err != nil {
		return BeginStageResult{}, fmt.Errorf("%w: encode request: %v", ErrUnavailable, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.stagesBeginURL(), bytes.NewReader(reqBody))
	if err != nil {
		return BeginStageResult{}, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")

	respBody, status, fromDaemon, err := c.do(req)
	if err != nil {
		return BeginStageResult{}, err
	}
	if !fromDaemon {
		return BeginStageResult{}, fmt.Errorf("%w: response missing %s header -- not trusted as a store answer", ErrUnavailable, daemonHeaderName)
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
	reqBody, err := json.Marshal(wire)
	if err != nil {
		return BeginStageResult{}, fmt.Errorf("%w: encode request: %v", ErrUnavailable, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.stagesEndURL(), bytes.NewReader(reqBody))
	if err != nil {
		return BeginStageResult{}, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")

	respBody, status, fromDaemon, err := c.do(req)
	if err != nil {
		return BeginStageResult{}, err
	}
	if !fromDaemon {
		return BeginStageResult{}, fmt.Errorf("%w: response missing %s header -- not trusted as a store answer", ErrUnavailable, daemonHeaderName)
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
