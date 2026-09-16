package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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
// A body that fills the client's whole response cap is refused rather than
// returned: a bundle that large was truncated by the cap mid-read, and a
// partial bundle handed to a reasoning pass is the exact defect the
// never-a-partial-bundle contract exists to prevent. A 404 is ErrNotFound
// — this daemon predates the endpoint. Every other non-200, and any
// transport failure, is ErrUnavailable: a read that cannot be answered
// reports that it could not, never a partial bundle.
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
	case status == http.StatusOK && len(body) >= maxResponseBytes:
		return nil, fmt.Errorf("%w: bundle filled the %d-byte response cap -- truncated, not served whole", ErrUnavailable, maxResponseBytes)
	case status == http.StatusOK:
		return body, nil
	case status == http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s/%s", ErrNotFound, project, change)
	default:
		return nil, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}
