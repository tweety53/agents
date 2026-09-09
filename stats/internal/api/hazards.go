package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/records"
)

// hazardApplies is the closed applies vocabulary the handlers accept,
// mirrored by the hazards table's own CHECK constraint -- the API's
// validation is UX, the store's is the boundary.
var hazardApplies = map[string]bool{"all": true, "cross-repo": true, "single-repo": true}

// ApplyHazardRecord records one hazard against rw, refusing an empty name
// or body, or an applies outside the closed vocabulary, before the store is
// touched. See ApplyVerdictRecord for why it takes RecordStore.
func ApplyHazardRecord(ctx context.Context, rw RecordStore, projectKey string, in records.Hazard) (records.Hazard, error) {
	if in.Name == "" || in.Body == "" {
		return records.Hazard{}, fmt.Errorf("%w: name and body are required", ErrInvalidRecord)
	}
	if !hazardApplies[in.Applies] {
		return records.Hazard{}, fmt.Errorf("%w: applies must be one of all, cross-repo, single-repo", ErrInvalidRecord)
	}
	return rw.AddHazard(ctx, projectKey, in)
}

// hazardHandler serves the three hazard endpoints. Thin like recordHandler:
// decode, refuse what cannot identify what it describes, hand the typed
// record to the store.
type hazardHandler struct {
	store  RecordStore
	logger *slog.Logger
}

// addHazard serves POST /api/v1/hazards/{project}: a proactive, per-project
// warning a dispatch bundle carries before it costs time. name, body and a
// closed-vocabulary applies are refused before the store is touched; a
// (project, name) the project already holds is a 409, the unique constraint
// refusing a re-add rather than a silent duplicate row.
func (h *hazardHandler) addHazard(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	var in records.Hazard
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := ApplyHazardRecord(r.Context(), h.store, project, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("add hazard for %s", project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// listHazards serves GET /api/v1/hazards/{project}: the project's hazards,
// newest first. The optional shape query filters to applies IN ('all',
// shape) exactly as store.ListHazards does; absent, it is the operator's
// unfiltered listing; present but outside the closed vocabulary, a 400. The
// all=true query lifts the active filter for the -all view.
func (h *hazardHandler) listHazards(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	shape := r.URL.Query().Get("shape")
	if shape != "" && !hazardApplies[shape] {
		writeError(w, http.StatusBadRequest, "shape must be one of all, cross-repo, single-repo")
		return
	}
	includeInactive := r.URL.Query().Get("all") == "true"

	out, err := h.store.ListHazards(r.Context(), project, shape, includeInactive)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list hazards for %s", project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// retireHazard serves PATCH /api/v1/hazards/{project}/{name}: sets
// active=false, never deletes -- the same disposition incidents get. An
// unknown name is a 404.
func (h *hazardHandler) retireHazard(w http.ResponseWriter, r *http.Request) {
	project, name := r.PathValue("project"), r.PathValue("name")

	out, err := h.store.RetireHazard(r.Context(), project, name)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("retire hazard %s/%s", project, name), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
