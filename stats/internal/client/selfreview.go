package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tweety53/agents/stats/internal/records"
)

// selfReviewBundleURL is the bundle route's URL for one project/change and
// the caller-resolved worktree root. The repo travels as a query parameter:
// the change-directory sources are read from the worktree the caller named
// — the one its own cwd resolves to — never from a path the store guessed
// at.
func (c *Client) selfReviewBundleURL(project, change, repo string) string {
	u := c.baseURL + "/api/v1/self-review/" + url.PathEscape(project) + "/" + url.PathEscape(change) + "/bundle"
	if repo != "" {
		u += "?repo=" + url.QueryEscape(repo)
	}
	return u
}

// GetSelfReviewBundle fetches a change's whole self-review
// context bundle, assembled server-side by flowd. repo is the worktree root
// the change-directory sources are read from — the caller resolves it from
// its own location, the way every flow command resolves its project. The body arrives already rendered and is returned
// verbatim — the CLI constructs none of it, the record-render rule, so a
// bundle's shape is decided in the one binary versioned with the store
// that produces it.
//
// A 404 is ErrNotFound — this daemon predates the endpoint. Every other
// non-200, and any transport failure, is ErrUnavailable: a read that
// cannot be answered reports that it could not, never a partial bundle.
// Truncation at the client's response cap is do()'s own refusal, shared by
// every read.
func (c *Client) GetSelfReviewBundle(ctx context.Context, project, change, repo string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.selfReviewBundleURL(project, change, repo), nil)
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

	switch {
	case status == http.StatusOK:
		return body, nil
	case status == http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s/%s", ErrNotFound, project, change)
	default:
		return nil, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}

// selfReviewFindingsURL is the findings route for one project/change.
func (c *Client) selfReviewFindingsURL(project, change string) string {
	return c.baseURL + "/api/v1/self-review/" + url.PathEscape(project) + "/" + url.PathEscape(change) + "/findings"
}

// RecordSelfReviewFinding records one self-review finding and its outcome
// for project/change and returns the row the daemon stored. A store
// validation refusal is ErrRecordRejected; an unreachable store is
// ErrUnavailable, which the CLI turns into a warning, never a failure.
func (c *Client) RecordSelfReviewFinding(ctx context.Context, project, change string, in records.SelfReviewFinding) (records.SelfReviewFinding, error) {
	in.Change = change
	var out records.SelfReviewFinding
	if _, err := c.writeRecord(ctx, http.MethodPost, c.selfReviewFindingsURL(project, change), in,
		map[int]bool{http.StatusCreated: true}, &out); err != nil {
		return records.SelfReviewFinding{}, err
	}
	return out, nil
}

// ListSelfReviewFindings fetches project/change's self-review findings in
// the order they were recorded.
func (c *Client) ListSelfReviewFindings(ctx context.Context, project, change string) ([]records.SelfReviewFinding, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.selfReviewFindingsURL(project, change), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	respBody, status, err := c.send(req)
	if err != nil {
		return nil, err
	}
	var out []records.SelfReviewFinding
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSelfReviewFinding fetches one of project's self-review findings by its id; an id the
// project has no row for is ErrRecordRejected.
func (c *Client) GetSelfReviewFinding(ctx context.Context, project string, id int64) (records.SelfReviewFinding, error) {
	u := c.baseURL + "/api/v1/self-review-finding/" + url.PathEscape(project) + "/" + strconv.FormatInt(id, 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return records.SelfReviewFinding{}, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	respBody, status, err := c.send(req)
	if err != nil {
		return records.SelfReviewFinding{}, err
	}
	var out records.SelfReviewFinding
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return records.SelfReviewFinding{}, err
	}
	return out, nil
}

// SimilarSelfReviewFindings fetches the n of project's self-review findings most similar to query, best first.
func (c *Client) SimilarSelfReviewFindings(ctx context.Context, project, query string, n int) ([]records.SimilarSelfReviewFinding, error) {
	out, _, _, err := c.SimilarSelfReviewFindingsMode(ctx, project, query, n)
	return out, err
}

// SimilarSelfReviewFindingsMode is SimilarSelfReviewFindings that also returns the daemon's
// X-Similar-Mode and, on bm25-fallback, the reason the embedding half failed.
func (c *Client) SimilarSelfReviewFindingsMode(ctx context.Context, project, query string, n int) (rows []records.SimilarSelfReviewFinding, mode, reason string, err error) {
	u := c.baseURL + "/api/v1/self-review-similar/" + url.PathEscape(project) + "?q=" + url.QueryEscape(query) + "&n=" + strconv.Itoa(n)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, "", "", fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	respBody, status, fromDaemon, h, err := c.doHeader(req)
	if err != nil {
		return nil, "", "", err
	}
	if !fromDaemon {
		return nil, "", "", fmt.Errorf("%w: response missing %s header -- not trusted as a store answer", ErrUnavailable, daemonHeaderName)
	}
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &rows); err != nil {
		return nil, "", "", err
	}
	return rows, h.Get("X-Similar-Mode"), h.Get("X-Similar-Error"), nil
}

// ReviewSelfReviewFinding sets a suggested finding's review status (to-review, approved or
// declined) and reason; an id that is no suggested finding is ErrNotFound.
func (c *Client) ReviewSelfReviewFinding(ctx context.Context, project string, id int64, status, reason string) error {
	u := c.baseURL + "/api/v1/self-review/" + url.PathEscape(project) + "/findings/" + strconv.FormatInt(id, 10) + "/review"
	_, err := c.writeRecord(ctx, http.MethodPatch, u, map[string]string{"status": status, "reason": reason},
		map[int]bool{http.StatusNoContent: true}, nil)
	return err
}
