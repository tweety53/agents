package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// lessonsResolveURL is the resolve route's URL for one topic. The topic
// travels as a query parameter: it is words, not a name the store keys.
func (c *Client) lessonsResolveURL(topic string) string {
	return c.baseURL + "/api/v1/lessons/resolve?topic=" + url.QueryEscape(topic)
}

// ResolveLesson fetches the process-lesson answer for topic, assembled
// server-side by flowd — every registered project's docs/briefs/ and
// archived narratives scanned, briefs ranked before narrative mentions.
// The body arrives already rendered and is returned verbatim — the CLI
// constructs none of it, the record-render rule.
//
// A 404 is ErrNotFound — this daemon predates the endpoint. Every other
// non-200, and any transport failure, is ErrUnavailable: a read that
// cannot be answered reports that it could not, never a partial answer.
func (c *Client) ResolveLesson(ctx context.Context, topic string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.lessonsResolveURL(topic), nil)
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
		return nil, fmt.Errorf("%w: lessons resolve", ErrNotFound)
	default:
		return nil, fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}
