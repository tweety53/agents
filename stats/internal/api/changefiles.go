package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/store"
)

// ChangeFileStore is the store dependency the change-file routes need
// (KAN-1073), defined at the consumer. It is wired through WithChangeFiles,
// as GuardRunStore is, so the positional stores' fakes need no new methods.
type ChangeFileStore interface {
	SaveChangeFile(ctx context.Context, projectKey, change, path, body string) error
	GetChangeFile(ctx context.Context, projectKey, change, path string) (string, error)
	HasChangeFiles(ctx context.Context, projectKey, change string) (bool, error)
}

var _ ChangeFileStore = (*store.Store)(nil)

// WithChangeFiles registers the change-file routes against s.
func WithChangeFiles(s ChangeFileStore) Option {
	return func(o *serverOptions) { o.changeFiles = s }
}

// changeFileHandler serves a change's saved files: the plan and the
// self-review bundle that outlive its worktree.
type changeFileHandler struct {
	store  ChangeFileStore
	logger *slog.Logger
}

// changeFileRequest is the PUT body: the file's whole text.
type changeFileRequest struct {
	Body string `json:"body"`
}

// put serves PUT /api/v1/changes/{project}/{name}/files/{path...}: upserts
// one file, 204. An unknown change is 404 through mapStoreError.
func (h *changeFileHandler) put(w http.ResponseWriter, r *http.Request) {
	project, name, path := r.PathValue("project"), r.PathValue("name"), r.PathValue("path")
	var in changeFileRequest
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.SaveChangeFile(r.Context(), project, name, path, in.Body); err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("save change file %s/%s %s", project, name, path), err)
		writeError(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// get serves GET of the same path: the body as text/markdown, 404 when the
// change or the file is absent.
func (h *changeFileHandler) get(w http.ResponseWriter, r *http.Request) {
	project, name, path := r.PathValue("project"), r.PathValue("name"), r.PathValue("path")
	body, err := h.store.GetChangeFile(r.Context(), project, name, path)
	if errors.Is(err, store.ErrChangeFileNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("read change file %s/%s %s", project, name, path), err)
		writeError(w, status, msg)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

// saved serves GET /api/v1/changes/{project}/{name}/files: {"saved": bool},
// whether any file of the change is saved.
func (h *changeFileHandler) saved(w http.ResponseWriter, r *http.Request) {
	project, name := r.PathValue("project"), r.PathValue("name")
	has, err := h.store.HasChangeFiles(r.Context(), project, name)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list change files %s/%s", project, name), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"saved": has})
}
