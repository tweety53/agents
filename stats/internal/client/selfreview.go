package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// selfReviewBundleURL is the bundle route's URL for one project/change.
func (c *Client) selfReviewBundleURL(project, change string) string {
	return c.baseURL + "/api/v1/self-review/" + url.PathEscape(project) + "/" + url.PathEscape(change) + "/bundle"
}

// GetSelfReviewBundle fetches a finished change's whole self-review
// context bundle, assembled server-side by flowd. The body arrives
// already rendered and is returned verbatim -- the CLI constructs none of
// it, the record-render rule, so a bundle's shape is decided in the one
// binary versioned with the store that produces it.
//
// A 404 is ErrNotFound -- the route itself is unknown to this daemon, the
// one case where a caller might still be talking to a flowd that predates
// the endpoint. Every other non-200, and any transport failure, is
// ErrUnavailable: a read that cannot be answered reports that it could
// not, never a partial bundle.
func (c *Client) GetSelfReviewBundle(ctx context.Context, project, change string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.selfReviewBundleURL(project, change), nil)
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
		return body, nil
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s/%s", ErrNotFound, project, change)
	default:
		return nil, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}
