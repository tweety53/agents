package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tweety53/agents/stats/internal/records"
)

// selfReviewBundleURL is the bundle route's URL for one project/change and
// the caller-resolved repository root. The repo travels as a query
// parameter: the archive-derived sources are read out of the repository
// the caller named — the main checkout its own cwd resolves to — never out
// of a path the store guessed at.
func (c *Client) selfReviewBundleURL(project, change, repo string) string {
	u := c.baseURL + "/api/v1/self-review/" + url.PathEscape(project) + "/" + url.PathEscape(change) + "/bundle"
	if repo != "" {
		u += "?repo=" + url.QueryEscape(repo)
	}
	return u
}

// GetSelfReviewBundle fetches a finished change's whole self-review
// context bundle, assembled server-side by flowd. repo is the repository
// root the archive-derived sources are read from — the caller resolves it
// from its own location in the repository, the way every flow command
// resolves its project. The body arrives already rendered and is returned
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
