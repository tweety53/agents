package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// changeFilesURL is a change's saved-file collection (KAN-1073); a path
// names one file in it.
func (c *Client) changeFilesURL(project, change string) string {
	return c.baseURL + "/api/v1/changes/" + url.PathEscape(project) + "/" + url.PathEscape(change) + "/files"
}

// PutChangeFile upserts one file of a change. A refusal is
// ErrRecordRejected, an unknown change ErrNotFound, an unreachable store
// ErrUnavailable.
func (c *Client) PutChangeFile(ctx context.Context, project, change, path, body string) error {
	_, err := c.writeRecord(ctx, http.MethodPut, c.changeFilesURL(project, change)+"/"+url.PathEscape(path),
		map[string]string{"body": body}, map[int]bool{http.StatusNoContent: true}, nil)
	return err
}

// GetChangeFile reads one saved file of a change, verbatim. A change or a
// file the store does not hold is ErrNotFound.
func (c *Client) GetChangeFile(ctx context.Context, project, change, path string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.changeFilesURL(project, change)+"/"+url.PathEscape(path), nil)
	if err != nil {
		return "", fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	body, status, err := c.send(req)
	if err != nil {
		return "", err
	}
	if _, err := classifyRecordResponse(body, status, map[int]bool{http.StatusOK: true}, nil); err != nil {
		return "", err
	}
	return string(body), nil
}

// HasChangeFiles reports whether any file of the change is saved.
func (c *Client) HasChangeFiles(ctx context.Context, project, change string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.changeFilesURL(project, change), nil)
	if err != nil {
		return false, fmt.Errorf("%w: build request: %v", ErrUnavailable, err)
	}
	body, status, err := c.send(req)
	if err != nil {
		return false, err
	}
	var out struct {
		Saved bool `json:"saved"`
	}
	if _, err := classifyRecordResponse(body, status, map[int]bool{http.StatusOK: true}, &out); err != nil {
		return false, err
	}
	return out.Saved, nil
}
