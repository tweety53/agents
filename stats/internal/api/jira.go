package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/jira"
)

// jiraHandler serves POST /api/v1/jira/transition: the pipeline's
// four-position forward-only Jira status transition, performed by this
// daemon rather than in the calling session, so a transient outage is
// ridden out by internal/jira's retry budget instead of stranding a
// landing run (KAN-571).
//
// The handler is thin in the usual way -- decode, validate shape,
// delegate to the JiraTransitioner, map its typed errors -- and its
// dependency is nil exactly when the daemon has no FLOWD_JIRA_* block, a
// normal state reported per request (503), never a startup failure.
type jiraHandler struct {
	transitions JiraTransitioner
	logger      *slog.Logger
}

// JiraTransitioner is the jira dependency the transition endpoint needs,
// defined here at the consumer per go-interface-design: exactly the one
// method the handler calls, so *jira.Client satisfies it and a test can
// exercise the endpoint through api.New by pointing cfg.Jira at a
// scripted Atlassian stand-in.
type JiraTransitioner interface {
	Transition(ctx context.Context, issueKey string, target jira.Position) (jira.Result, error)
}

// var _ JiraTransitioner = (*jira.Client)(nil) verifies at compile time
// that the real client satisfies the interface this package depends on.
var _ JiraTransitioner = (*jira.Client)(nil)

// jiraTransitionRequest is the wire shape of POST /api/v1/jira/transition.
// Target is a position name in any of its accepted spellings ("In
// Progress", "code review", ...); the vocabulary lives in internal/jira
// and is not restated here.
type jiraTransitionRequest struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}

// jiraTransitionResponse is the wire shape of the answer: the status the
// issue actually carries, and whether this call moved it. Moved false is
// the already-at-or-past-target no-op, which is a success here -- the
// caller reports it as "already <status> (no transition)".
type jiraTransitionResponse struct {
	Key    string `json:"key"`
	Status string `json:"status"`
	Moved  bool   `json:"moved"`
}

// transition serves POST /api/v1/jira/transition.
func (h *jiraHandler) transition(w http.ResponseWriter, r *http.Request) {
	if h.transitions == nil {
		writeError(w, http.StatusServiceUnavailable,
			"jira transitions are not configured on this daemon (set FLOWD_JIRA_SITE, FLOWD_JIRA_EMAIL and FLOWD_JIRA_TOKEN)")
		return
	}
	var req jiraTransitionRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Key == "" {
		writeError(w, http.StatusBadRequest, "key is required")
		return
	}
	target, err := jira.ParseTarget(req.Target)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.transitions.Transition(r.Context(), req.Key, target)
	if err != nil {
		status, msg := mapJiraError(h.logger, err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, jiraTransitionResponse{Key: result.Key, Status: result.Status, Moved: result.Moved})
}

// mapJiraError turns a typed jira error into the HTTP status and message
// this daemon reports, with the same reachability logic mapStoreError
// documents: every status below 500 is the upstream having answered, and
// 504 (the retry budget exhausted on 429/5xx/transport failures) plus the
// unmapped default are the failures a caller may treat as "try again or
// give up for now".
//
// The statuses are chosen by what each failure is:
//   - 422 Unprocessable Entity: the issue's current status matches no
//     name mapped onto the four positions. The request and the credentials
//     were fine; the answer needs an operator decision, which is why the
//     message names the status verbatim.
//   - 409 Conflict: the workflow currently offers no transition to the
//     requested position -- the issue sits at a state whose allowed moves
//     cannot reach it, an upstream conflict rather than a malformed
//     request.
//   - 404 Not Found: no such issue.
//   - 502 Bad Gateway: Jira rejected the credentials -- this daemon's
//     configuration problem, reported upstream so the CLI can print it.
//   - 504 Gateway Timeout: internal/jira's retry budget ran out on
//     transient failures.
func mapJiraError(logger *slog.Logger, err error) (status int, msg string) {
	switch {
	case errors.Is(err, jira.ErrUnrecognizedStatus):
		return http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, jira.ErrNoTransition):
		return http.StatusConflict, err.Error()
	case errors.Is(err, jira.ErrIssueNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, jira.ErrAuth):
		return http.StatusBadGateway, err.Error()
	case errors.Is(err, jira.ErrTransientExhausted):
		return http.StatusGatewayTimeout, err.Error()
	default:
		if logger != nil {
			logger.Error("jira transition", "error", err)
		}
		return http.StatusInternalServerError, "internal error"
	}
}
