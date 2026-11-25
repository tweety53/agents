package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tweety53/agents/stats/internal/records"
)

// RecordConfidenceEstimate records one confidence estimate below 100% for
// project/change and returns the row the daemon stored. A store validation
// refusal is ErrRecordRejected; an unreachable store is ErrUnavailable, which
// the CLI turns into a warning, never a failure.
func (c *Client) RecordConfidenceEstimate(ctx context.Context, project, change string, in records.ConfidenceEstimate) (records.ConfidenceEstimate, error) {
	var out records.ConfidenceEstimate
	u := c.baseURL + "/api/v1/confidence/" + url.PathEscape(project) + "/" + url.PathEscape(change)
	if _, err := c.writeRecord(ctx, http.MethodPost, u, in, map[int]bool{http.StatusCreated: true}, &out); err != nil {
		return records.ConfidenceEstimate{}, err
	}
	return out, nil
}

// ConfidenceTags fetches every tag in use on the recorded confidence
// estimates, alphabetically.
func (c *Client) ConfidenceTags(ctx context.Context) ([]records.ConfidenceTag, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/confidence-tags", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	respBody, status, err := c.send(req)
	if err != nil {
		return nil, err
	}
	var out []records.ConfidenceTag
	if _, err := classifyRecordResponse(respBody, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
