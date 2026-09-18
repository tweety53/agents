// Package jira implements the pipeline's Jira status transitions against
// the Atlassian Cloud REST API (v3): the four-position forward-only table
// with retries, so a transient outage cannot strand a landing run
// (KAN-571). The positions and their accepted status names are the
// pipeline contract's own:
//
//	1  To Do       — "To Do", "TO DO URGENT"
//	2  In Progress — "In Progress", "In-Progress"
//	3  In Review   — "In Review", "Code Review"
//	4  Done        — "Done"
//
// Names match case-insensitively, with ordinary whitespace variance folded
// away and a hyphen standing in for a space. A status matching no name
// above has no position and is never given one — in particular never from
// Jira's statusCategory, which groups custom statuses misleadingly. An
// issue already at or past the target is a successful no-op, never a
// backward move.
//
// The retry unit is the whole read-transition-report sequence, not the
// individual HTTP call: each attempt re-reads the issue's current status
// first, so a POST that landed server-side but whose response was lost
// replays as the already-at-target no-op instead of a duplicate or an
// error. Only transport errors and HTTP 429/5xx are transient; every
// other answer is definitive and returned as its typed error without a
// retry.
package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Position is one slot in the pipeline's four-position forward-only
// order. PositionUnknown is not a slot: it marks a status name that maps
// to nothing, and a transition targeting it or sitting on it is refused
// rather than inferred.
type Position int

const (
	PositionUnknown    Position = 0
	PositionToDo       Position = 1
	PositionInProgress Position = 2
	PositionInReview   Position = 3
	PositionDone       Position = 4
)

// positionByName is the whole name-to-position table, keyed by
// normalizeStatusName's output. This map is the package's one statement of
// the contract's accepted spellings; every match goes through it.
var positionByName = map[string]Position{
	"to do":        PositionToDo,
	"to do urgent": PositionToDo,
	"in progress":  PositionInProgress,
	"in review":    PositionInReview,
	"code review":  PositionInReview,
	"done":         PositionDone,
}

// positionNames names each position, for messages and for ParseTarget's
// error, which must tell the caller what an accepted spelling looks like.
var positionNames = map[Position]string{
	PositionToDo:       "To Do",
	PositionInProgress: "In Progress",
	PositionInReview:   "In Review",
	PositionDone:       "Done",
}

// String returns the position's canonical pipeline name, or "unknown".
func (p Position) String() string {
	if n, ok := positionNames[p]; ok {
		return n
	}
	return "unknown"
}

// normalizeStatusName folds a status name the loose way the contract
// matches by: case-insensitively, hyphen for space, any whitespace run
// collapsed to one space, ends trimmed.
func normalizeStatusName(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(strings.ToLower(s), "-", " ")), " ")
}

// PositionForName maps a status name onto its position, PositionUnknown
// when it matches no accepted spelling. It never consults Jira's
// statusCategory: a name the table does not carry has no position.
func PositionForName(name string) Position {
	return positionByName[normalizeStatusName(name)]
}

// ParseTarget parses a user-supplied target — a position name in any of
// its accepted spellings — into the Position to move an issue to.
func ParseTarget(s string) (Position, error) {
	p := PositionForName(s)
	if p == PositionUnknown {
		return PositionUnknown, fmt.Errorf("jira: %q matches no pipeline position (accepted: To Do, In Progress, In Review, Done)", s)
	}
	return p, nil
}

// Config is the Jira access the Client needs: the site's base URL
// (e.g. "https://example.atlassian.net") and the Atlassian account's
// email and API token, used as HTTP basic auth.
type Config struct {
	Site     string
	Email    string
	APIToken string
}

// sentinel errors, each one a definitive answer the caller distinguishes
// on. errors.Is is the interface; the messages carry the specifics.
var (
	// ErrUnrecognizedStatus: the issue's current status matches no name
	// mapped onto the four positions. Retrying cannot change it.
	ErrUnrecognizedStatus = errors.New("jira: current status matches no name mapped onto the four ordered positions")
	// ErrNoTransition: the workflow currently offers no transition whose
	// target status maps to the requested position.
	ErrNoTransition = errors.New("jira: workflow offers no transition to the requested position")
	// ErrIssueNotFound: Jira reports no such issue (HTTP 404).
	ErrIssueNotFound = errors.New("jira: issue not found")
	// ErrAuth: Jira rejected the credentials (HTTP 401 or 403).
	ErrAuth = errors.New("jira: authentication rejected")
	// ErrTransientExhausted: the operation stayed transient-failing for
	// every attempt.
	ErrTransientExhausted = errors.New("jira: transient failure on every attempt")
)

// errRetryable marks an attempt whose failure a retry may repair:
// a transport error, or an HTTP 429/5xx answer. retryAfter carries the
// 429's Retry-After hint, 0 when absent or inapplicable.
type errRetryable struct {
	cause      error
	retryAfter time.Duration
}

func (e errRetryable) Error() string { return e.cause.Error() }
func (e errRetryable) Unwrap() error { return e.cause }

// Retry budget. maxAttempts and the backoff ladder are sized so the whole
// operation always finishes well inside flowd's 30s writeTimeout, which
// caps the response the CLI is waiting on: four attempts at a 4s
// per-request timeout plus worst-case backoff of 0.5s+1s+2s (jittered up
// no more than 20%) stays under 22s.
const (
	maxAttempts    = 4
	attemptTimeout = 4 * time.Second
	baseDelay      = 500 * time.Millisecond
	// maxRetryAfter caps an HTTP 429's Retry-After hint so a hostile or
	//misconfigured value cannot blow the retry budget on its own.
	maxRetryAfter = 4 * time.Second
)

// Result reports what one Transition call did. Moved false with a Status
// is the already-at-or-past-target no-op; Status is the status the issue
// actually carries (the no-op case) or lands in (the moved case).
type Result struct {
	Key    string
	Status string
	Moved  bool
}

// Client performs pipeline transitions against one Atlassian site.
type Client struct {
	cfg        Config
	httpClient *http.Client

	// pause waits out one inter-attempt delay, ctx-cancellable. nil means
	// the default timer wait; tests replace it to observe the requested
	// delays without real sleeping.
	pause func(ctx context.Context, d time.Duration) error
}

// New builds a Client against cfg. httpClient may be nil for
// http.DefaultClient.
func New(cfg Config, httpClient *http.Client) *Client {
	return &Client{cfg: cfg, httpClient: httpClient}
}

// issueStatusResponse is the slice of GET
// /rest/api/3/issue/{key}?fields=status this package reads.
type issueStatusResponse struct {
	Fields struct {
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
	} `json:"fields"`
}

// transitionsResponse is the slice of GET
// /rest/api/3/issue/{key}/transitions this package reads. The transition's
// own name is ignored — resolution goes by the target status's position,
// never by the transition's label or by its numeric id.
type transitionsResponse struct {
	Transitions []struct {
		ID string `json:"id"`
		To struct {
			Name string `json:"name"`
		} `json:"to"`
	} `json:"transitions"`
}

type transitionRequest struct {
	Transition struct {
		ID string `json:"id"`
	} `json:"transition"`
}

// Transition moves issueKey to target, forward-only, retrying transient
// failures per the package doc comment. An issue already at or past
// target is the Moved=false no-op; an issue whose status has no position
// fails with ErrUnrecognizedStatus rather than moving on a guess.
func (c *Client) Transition(ctx context.Context, issueKey string, target Position) (Result, error) {
	if target == PositionUnknown {
		return Result{}, fmt.Errorf("jira: cannot transition to an unknown position")
	}
	delay := baseDelay
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			if err := c.wait(ctx, delay); err != nil {
				return Result{}, err
			}
		}
		result, err := c.transitionOnce(ctx, issueKey, target)
		if err == nil {
			return result, nil
		}
		var retryable errRetryable
		if !errors.As(err, &retryable) {
			return Result{}, err
		}
		lastErr = err
		delay = backoffDelay(attempt, retryable.retryAfter)
	}
	return Result{}, fmt.Errorf("%w: %s to %s: %w", ErrTransientExhausted, issueKey, target, lastErr)
}

// backoffDelay doubles baseDelay per attempt, jittered ±20%; a 429's
// Retry-After hint replaces the ladder when present, capped at
// maxRetryAfter.
func backoffDelay(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		if retryAfter > maxRetryAfter {
			return maxRetryAfter
		}
		return retryAfter
	}
	d := baseDelay << (attempt - 1)
	jitter := 0.8 + 0.4*rand.Float64()
	return time.Duration(float64(d) * jitter)
}

func (c *Client) wait(ctx context.Context, d time.Duration) error {
	if c.pause == nil {
		t := time.NewTimer(d)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			return nil
		}
	}
	return c.pause(ctx, d)
}

// transitionOnce is one full attempt: read the status, no-op when already
// at or past target, else resolve and fire the transition whose target
// status maps to the requested position.
func (c *Client) transitionOnce(ctx context.Context, issueKey string, target Position) (Result, error) {
	statusName, err := c.currentStatus(ctx, issueKey)
	if err != nil {
		return Result{}, err
	}
	current := PositionForName(statusName)
	if current == PositionUnknown {
		return Result{}, fmt.Errorf("%w: %q", ErrUnrecognizedStatus, statusName)
	}
	if current >= target {
		return Result{Key: issueKey, Status: statusName, Moved: false}, nil
	}

	transitionID, targetName, err := c.transitionTo(ctx, issueKey, target)
	if err != nil {
		return Result{}, err
	}
	if err := c.fireTransition(ctx, issueKey, transitionID); err != nil {
		return Result{}, err
	}
	return Result{Key: issueKey, Status: targetName, Moved: true}, nil
}

func (c *Client) currentStatus(ctx context.Context, issueKey string) (string, error) {
	body, err := c.do(ctx, http.MethodGet, "/rest/api/3/issue/"+url.PathEscape(issueKey)+"?fields=status", nil)
	if err != nil {
		return "", err
	}
	var parsed issueStatusResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("jira: parse status of %s: %w", issueKey, err)
	}
	return parsed.Fields.Status.Name, nil
}

func (c *Client) transitionTo(ctx context.Context, issueKey string, target Position) (id, targetName string, err error) {
	body, err := c.do(ctx, http.MethodGet, "/rest/api/3/issue/"+url.PathEscape(issueKey)+"/transitions", nil)
	if err != nil {
		return "", "", err
	}
	var parsed transitionsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", fmt.Errorf("jira: parse transitions of %s: %w", issueKey, err)
	}
	for _, t := range parsed.Transitions {
		if PositionForName(t.To.Name) == target {
			return t.ID, t.To.Name, nil
		}
	}
	return "", "", fmt.Errorf("%w: %s wants %s", ErrNoTransition, issueKey, target)
}

func (c *Client) fireTransition(ctx context.Context, issueKey, transitionID string) error {
	payload := transitionRequest{}
	payload.Transition.ID = transitionID
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = c.do(ctx, http.MethodPost, "/rest/api/3/issue/"+url.PathEscape(issueKey)+"/transitions", body)
	return err
}

// do performs one authenticated REST call. A transport failure or a
// 429/5xx answer comes back as errRetryable (with Retry-After parsed off
// a 429); 404, 401 and 403 as their sentinels; any other non-2xx as a
// plain error carrying the status and body snippet.
func (c *Client) do(ctx context.Context, method, path string, payload []byte) ([]byte, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
	defer cancel()

	var bodyReader io.Reader
	if payload != nil {
		bodyReader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(attemptCtx, method, strings.TrimRight(c.cfg.Site, "/")+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("jira: build request %s %s: %w", method, path, err)
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.SetBasicAuth(c.cfg.Email, c.cfg.APIToken)

	client := c.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errRetryable{cause: fmt.Errorf("jira: %s %s: %w", method, path, err)}
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, errRetryable{cause: fmt.Errorf("jira: read response %s %s: %w", method, path, err)}
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return body, nil
	case resp.StatusCode == http.StatusTooManyRequests,
		resp.StatusCode >= http.StatusInternalServerError:
		return nil, errRetryable{
			cause:      fmt.Errorf("jira: %s %s: HTTP %d: %s", method, path, resp.StatusCode, snippet(body)),
			retryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
		}
	case resp.StatusCode == http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s %s: HTTP 404", ErrIssueNotFound, method, path)
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return nil, fmt.Errorf("%w: %s %s: HTTP %d", ErrAuth, method, path, resp.StatusCode)
	default:
		return nil, fmt.Errorf("jira: %s %s: HTTP %d: %s", method, path, resp.StatusCode, snippet(body))
	}
}

// parseRetryAfter reads a Retry-After seconds hint; absent or unparsable
// means none.
func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	s, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || s < 0 {
		return 0
	}
	return time.Duration(s) * time.Second
}

// snippet bounds a response body quoted in an error message.
func snippet(body []byte) string {
	s := strings.TrimSpace(string(body))
	const max = 200
	if len(s) > max {
		s = s[:max] + "…"
	}
	return s
}
