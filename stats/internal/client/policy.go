package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Policy fetches every role's pairs via GET /api/v1/policy, returned as the
// daemon's JSON ({role: {base, optin}}): its readers print and
// cache it whole.
func (c *Client) Policy(ctx context.Context) (json.RawMessage, error) {
	var out json.RawMessage
	if err := c.getPolicyJSON(ctx, c.baseURL+"/api/v1/policy", &out); err != nil {
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
	if err := c.getPolicyJSON(ctx, c.baseURL+"/api/v1/policy/grants?change="+url.QueryEscape(change), &out); err != nil {
		return nil, err
	}
	return out.Roles, nil
}

// Grant (revoke false) or revokes change's operator grant for roles via
// PUT or DELETE /api/v1/policy/grants. A refused call is
// ErrPolicyRejected carrying the daemon's message.
func (c *Client) Grant(ctx context.Context, change string, roles []string, revoke bool) error {
	method := http.MethodPut
	if revoke {
		method = http.MethodDelete
	}
	body, status, err := c.sendJSON(ctx, method, c.baseURL+"/api/v1/policy/grants", map[string]any{"change": change, "roles": roles})
	if err != nil {
		return err
	}
	if err := policyAnswer(body, status, http.StatusOK); err != nil {
		return fmt.Errorf("policy grant: %w", err)
	}
	return nil
}

// ErrPolicyRejected means the daemon refused a policy write as malformed
// (400) or naming a grant never made (404); the wrapped text is its message.
var ErrPolicyRejected = errors.New("client: store rejected the policy call")

// policyAnswer maps a policy write's status onto this package's errors: ok
// is success, 400/404 ErrPolicyRejected, anything else ErrUnavailable.
func policyAnswer(body []byte, status, ok int) error {
	var wire errorWireResponse
	_ = json.Unmarshal(body, &wire)
	switch status {
	case ok:
		return nil
	case http.StatusBadRequest, http.StatusNotFound:
		return fmt.Errorf("%w: %s", ErrPolicyRejected, wire.Error)
	default:
		return fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
}

// getPolicyJSON GETs url and decodes its 200 body into out; any other
// outcome is ErrUnavailable.
func (c *Client) getPolicyJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	body, status, err := c.send(req)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("%w: unexpected status %d", ErrUnavailable, status)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("%w: response body is not valid JSON", ErrUnavailable)
	}
	return nil
}

// Bottlenecks reads GET /api/v1/stats/bottlenecks: the wall-clock ranking
// of the last `last` changes, project "" meaning every project.
func (c *Client) Bottlenecks(ctx context.Context, project string, last int) (json.RawMessage, error) {
	q := url.Values{"last": {strconv.Itoa(last)}}
	if project != "" {
		q.Set("project", project)
	}
	var out json.RawMessage
	if err := c.getPolicyJSON(ctx, c.baseURL+"/api/v1/stats/bottlenecks?"+q.Encode(), &out); err != nil {
		return nil, err
	}
	return out, nil
}
