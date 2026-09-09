package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tweety53/agents/stats/internal/records"
)

// suitesURL is the project-scoped suite-run collection, the hazards pair's
// URL shape with the runs segment naming what the rows are: timed suite
// executions (KAN-252).
func (c *Client) suitesURL(project string) string {
	return c.baseURL + "/api/v1/suites/" + url.PathEscape(project) + "/runs"
}

// RecordSuiteRun posts one timed suite execution and returns the row as
// the store recorded it -- the daemon's id and ran_at included, neither of
// which the caller could know. An unreachable or untrustworthy store
// answers ErrUnavailable, the record calls' shared fallback contract: the
// caller decides what a wrapped test run does with it, never this layer.
func (c *Client) RecordSuiteRun(ctx context.Context, project string, run records.SuiteRun) (records.SuiteRun, error) {
	var out records.SuiteRun
	_, err := c.writeRecord(ctx, http.MethodPost, c.suitesURL(project), run,
		map[int]bool{http.StatusCreated: true}, &out)
	if err != nil {
		return records.SuiteRun{}, err
	}
	return out, nil
}

// ListSuiteRuns reads a project's recorded suite runs, newest first. An
// empty suite lists every suite; limit caps the row count and must be > 0,
// exactly as store.ListSuiteRuns documents -- the two validations are the
// same rule stated at both ends of one wire call.
func (c *Client) ListSuiteRuns(ctx context.Context, project, suite string, limit int) ([]records.SuiteRun, error) {
	listURL := c.suitesURL(project)
	q := url.Values{}
	if suite != "" {
		q.Set("suite", suite)
	}
	q.Set("limit", fmt.Sprintf("%d", limit))
	listURL += "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	respBody, status, err := c.send(req)
	if err != nil {
		return nil, err
	}
	var out []records.SuiteRun
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
