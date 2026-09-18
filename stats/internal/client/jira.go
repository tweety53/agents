package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrJiraRefused means flowd was reached and refused the transition, or
// reported Jira refusing it -- a bad target spelling (400), an unknown
// issue (404), no offered transition (409), an unrecognized current
// status (422), rejected credentials (502) or a retry budget that ran
// out (504). The daemon's message is carried verbatim: the pipeline
// prints it as its one Jira skip line. This is never a fallback trigger
// -- the transition path has no journal, and repeating a refused call
// would be refused identically.
var ErrJiraRefused = errors.New("client: jira transition refused")

// JiraTransitionResult mirrors the daemon's answer: the status the issue
// actually carries, and whether this call moved it. Moved false is the
// already-at-or-past-target no-op -- a success the caller reports as
// "already <status> (no transition)".
type JiraTransitionResult struct {
	Key    string `json:"key"`
	Status string `json:"status"`
	Moved  bool   `json:"moved"`
}

// jiraTransitionRequest is the wire shape POST /api/v1/jira/transition
// accepts. Target is a position name in any of its accepted spellings;
// the vocabulary lives in internal/jira on the daemon side and is the
// daemon's to validate, so this client passes the caller's words through
// untouched.
type jiraTransitionRequest struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}

// JiraTransition moves the issue to the named pipeline position via POST
// /api/v1/jira/transition. A 200 is success; every mapped 4xx/5xx refusal
// is ErrJiraRefused carrying the daemon's message; anything else --
// transport failure, a foreign server on the port, a malformed body --
// is ErrUnavailable.
func (c *Client) JiraTransition(ctx context.Context, key, target string) (JiraTransitionResult, error) {
	respBody, status, err := c.sendJSON(ctx, http.MethodPost, c.baseURL+"/api/v1/jira/transition",
		jiraTransitionRequest{Key: key, Target: target})
	if err != nil {
		return JiraTransitionResult{}, err
	}

	switch {
	case status == http.StatusOK:
		var out JiraTransitionResult
		if err := json.Unmarshal(respBody, &out); err != nil {
			return JiraTransitionResult{}, fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
		}
		return out, nil
	case status == http.StatusBadRequest,
		status == http.StatusNotFound,
		status == http.StatusConflict,
		status == http.StatusUnprocessableEntity,
		status == http.StatusBadGateway,
		status == http.StatusServiceUnavailable,
		status == http.StatusGatewayTimeout:
		var wire errorWireResponse
		if err := json.Unmarshal(respBody, &wire); err != nil || wire.Error == "" {
			return JiraTransitionResult{}, fmt.Errorf("%w: HTTP %d: %s", ErrJiraRefused, status, string(respBody))
		}
		return JiraTransitionResult{}, fmt.Errorf("%w: %s", ErrJiraRefused, wire.Error)
	default:
		return JiraTransitionResult{}, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}
