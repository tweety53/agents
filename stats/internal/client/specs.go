package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tweety53/agents/stats/internal/records"
)

// specsRunsURL is the project-scoped spec-run collection, the suites pair's
// URL shape with the specs segment naming what the rows are: one spec
// file's last-run timestamp (KAN-418).
func (c *Client) specsRunsURL(project string) string {
	return c.baseURL + "/api/v1/specs/" + url.PathEscape(project) + "/runs"
}

func (c *Client) specsLastRunsURL(project string) string {
	return c.baseURL + "/api/v1/specs/" + url.PathEscape(project) + "/lastruns"
}

// RecordSpecRun posts one spec file's run and returns the inventory row as
// the store recorded it -- the stamp the caller did not name included. An
// unreachable or untrustworthy store answers ErrUnavailable, the record
// calls' shared fallback contract.
func (c *Client) RecordSpecRun(ctx context.Context, project, spec string) (records.SpecLastRun, error) {
	var out records.SpecLastRun
	_, err := c.writeRecord(ctx, http.MethodPost, c.specsRunsURL(project), map[string]string{"spec": spec},
		map[int]bool{http.StatusCreated: true}, &out)
	if err != nil {
		return records.SpecLastRun{}, err
	}
	return out, nil
}

// ListSpecLastRuns reads a project's per-spec inventory, ordered by spec.
// Named specs answer one row each, recorded or not -- a never-recorded name
// carrying a nil stamp and the project's whole change count; an empty
// specs slice lists the recorded specs alone, exactly as
// store.ListSpecLastRuns documents -- the two validations are the same rule
// stated at both ends of one wire call.
func (c *Client) ListSpecLastRuns(ctx context.Context, project string, specs []string) ([]records.SpecLastRun, error) {
	listURL := c.specsLastRunsURL(project)
	if len(specs) > 0 {
		q := url.Values{}
		for _, spec := range specs {
			q.Add("spec", spec)
		}
		listURL += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	respBody, status, err := c.send(req)
	if err != nil {
		return nil, err
	}
	var out []records.SpecLastRun
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
