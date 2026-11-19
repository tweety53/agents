package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/store"
)

// projectsStore is the one read the projects endpoint needs, defined here at
// the consumer per go-interface-design.
type projectsStore interface {
	ProjectRoots(ctx context.Context) ([]store.ProjectRoot, error)
}

// projectsHandler serves GET /api/v1/projects.
type projectsHandler struct {
	store  projectsStore
	logger *slog.Logger
}

// projects serves GET /api/v1/projects: every registered project key, in
// key order. The state board names these as Project chips before its
// finished changes load, so a project whose changes are all finished still
// has a chip.
func (h *projectsHandler) projects(w http.ResponseWriter, r *http.Request) {
	roots, err := h.store.ProjectRoots(r.Context())
	if err != nil {
		status, msg := mapStoreError(h.logger, "read the registered projects", err)
		writeError(w, status, msg)
		return
	}
	keys := make([]string, len(roots))
	for i, root := range roots {
		keys[i] = root.ProjectKey
	}
	writeJSON(w, http.StatusOK, struct {
		Projects []string `json:"projects"`
	}{keys})
}
