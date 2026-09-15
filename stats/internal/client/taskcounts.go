package client

import (
	"context"
	"net/http"
	"net/url"

	"github.com/tweety53/agents/stats/internal/records"
)

// taskCountsURL is the change-scoped task-count observation collection,
// the decisions pair's URL shape with the task-counts segment naming what
// the rows are: plan-growth observations (KAN-415).
func (c *Client) taskCountsURL(project, change string) string {
	return c.baseURL + "/api/v1/records/" + url.PathEscape(project) + "/" + url.PathEscape(change) + "/task-counts"
}

// RecordTaskCount posts one observation of a change plan's total task
// count and returns the row as the store recorded it -- the daemon's id
// and observed_at included, neither of which the caller could know. An
// unreachable or untrustworthy store answers ErrUnavailable, the record
// calls' shared fallback contract: the caller decides what an unrecorded
// observation means, never this layer.
func (c *Client) RecordTaskCount(ctx context.Context, project, change string, in records.TaskCount) (records.TaskCount, error) {
	var out records.TaskCount
	_, err := c.writeRecord(ctx, http.MethodPost, c.taskCountsURL(project, change), in,
		map[int]bool{http.StatusCreated: true}, &out)
	if err != nil {
		return records.TaskCount{}, err
	}
	return out, nil
}
