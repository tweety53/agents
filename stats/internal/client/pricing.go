package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// RepriceResult is POST /api/v1/pricing/reprice's answer.
type RepriceResult struct {
	Priced  int `json:"priced"`
	Skipped int `json:"skipped"`
}

// Reprice asks flowd to price every stage run started on or after since
// (YYYY-MM-DD) again (decision reprice-command). A non-200 answer is an
// error carrying the response body.
func (c *Client) Reprice(ctx context.Context, since string) (RepriceResult, error) {
	body, status, err := c.sendJSON(ctx, http.MethodPost, c.baseURL+"/api/v1/pricing/reprice", map[string]string{"since": since})
	if err != nil {
		return RepriceResult{}, err
	}
	if status != http.StatusOK {
		return RepriceResult{}, fmt.Errorf("reprice: status %d: %s", status, body)
	}
	var out RepriceResult
	if err := json.Unmarshal(body, &out); err != nil {
		return RepriceResult{}, fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
	}
	return out, nil
}
