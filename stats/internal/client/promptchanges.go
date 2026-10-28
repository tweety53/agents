package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PromptChange is the wire shape POST /api/v1/prompt-changes accepts,
// copied from internal/api's promptChangeDTO for the same boundary reason
// Settings is.
type PromptChange struct {
	SHA     string    `json:"sha"`
	At      time.Time `json:"at"`
	Kind    string    `json:"kind"`
	Area    string    `json:"area"`
	Slot    *string   `json:"slot"`
	Target  string    `json:"target"`
	Summary string    `json:"summary"`
	CaseID  *string   `json:"caseId"`
}

// RecordPromptChange upserts one prompt change via POST
// /api/v1/prompt-changes. A 204 is success; a 400 carries the daemon's
// message; every other outcome is ErrUnavailable.
func (c *Client) RecordPromptChange(ctx context.Context, in PromptChange) error {
	respBody, status, err := c.sendJSON(ctx, http.MethodPost, c.baseURL+"/api/v1/prompt-changes", in)
	if err != nil {
		return err
	}
	switch status {
	case http.StatusNoContent:
		return nil
	case http.StatusBadRequest:
		var wire errorWireResponse
		_ = json.Unmarshal(respBody, &wire)
		return fmt.Errorf("client: store rejected the prompt change: %s", wire.Error)
	default:
		return fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}
