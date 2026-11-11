package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Policy fetches every role's pairs via GET /api/v1/policy, returned as the
// daemon's JSON ({role: {base, optin, marked}}): its readers print and
// cache it whole.
func (c *Client) Policy(ctx context.Context) (json.RawMessage, error) {
	var out json.RawMessage
	if err := c.getExperimentsJSON(ctx, c.baseURL+"/api/v1/policy", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Grants fetches the roles change holds a live operator grant for, via GET
// /api/v1/policy/grants.
func (c *Client) Grants(ctx context.Context, change string) ([]string, error) {
	var out struct {
		Roles []string `json:"roles"`
	}
	if err := c.getExperimentsJSON(ctx, c.baseURL+"/api/v1/policy/grants?change="+url.QueryEscape(change), &out); err != nil {
		return nil, err
	}
	return out.Roles, nil
}

// Grant (revoke false) or revokes change's operator grant for roles via
// PUT or DELETE /api/v1/policy/grants. A refused call is
// ErrExperimentRejected carrying the daemon's message.
func (c *Client) Grant(ctx context.Context, change string, roles []string, revoke bool) error {
	method := http.MethodPut
	if revoke {
		method = http.MethodDelete
	}
	body, status, err := c.sendJSON(ctx, method, c.baseURL+"/api/v1/policy/grants", map[string]any{"change": change, "roles": roles})
	if err != nil {
		return err
	}
	if err := experimentAnswer(body, status, http.StatusOK); err != nil {
		return fmt.Errorf("policy grant: %w", err)
	}
	return nil
}
