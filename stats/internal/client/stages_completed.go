package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// StageCompleted reports whether project/change holds a run of stage that
// ended completed (KAN-1073). An unreachable or untrustworthy store is
// ErrUnavailable.
func (c *Client) StageCompleted(ctx context.Context, project, change, stage string) (bool, error) {
	u := c.baseURL + "/api/v1/stages/" + url.PathEscape(project) + "/" + url.PathEscape(change) + "/" + url.PathEscape(stage) + "/completed"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return false, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	body, status, err := c.send(req)
	if err != nil {
		return false, err
	}
	var out struct {
		Completed bool `json:"completed"`
	}
	if _, err := classifyRecordResponse(body, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return false, err
	}
	return out.Completed, nil
}
