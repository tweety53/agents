package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// ApplySpecRunRecord records one spec file's last run against rw, refusing
// an empty spec before the store is touched. See ApplySuiteRunRecord for
// why it takes RecordStore.
func ApplySpecRunRecord(ctx context.Context, rw RecordStore, projectKey, spec string, ranAt time.Time) (records.SpecLastRun, error) {
	if spec == "" {
		return records.SpecLastRun{}, fmt.Errorf("%w: spec is required", ErrInvalidRecord)
	}
	return rw.RecordSpecRun(ctx, projectKey, spec, ranAt)
}

// specHandler serves the two spec last-run endpoints. Thin like
// suiteHandler: decode, refuse what cannot identify what it describes,
// hand the typed record to the store.
type specHandler struct {
	store  RecordStore
	logger *slog.Logger
}

// specRunRequest is the POST body: the one field a spec run needs to
// identify itself. The timestamp is the daemon's business -- the caller
// that does not name one is stamped with when the run was recorded, the
// store's own now() default doing the stamping.
type specRunRequest struct {
	Spec string `json:"spec"`
}

// recordSpecRun serves POST /api/v1/specs/{project}/runs: one spec file's
// last run, upserted per (project, spec) with the latest timestamp kept.
// An empty spec is refused before the store is touched; an unknown project
// is the foreign key answering, mapped like any other store refusal.
func (h *specHandler) recordSpecRun(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	var in specRunRequest
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := ApplySpecRunRecord(r.Context(), h.store, project, in.Spec, time.Time{})
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("record spec run for %s", project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// listSpecRuns serves GET /api/v1/specs/{project}/lastruns: the per-spec
// inventory, ordered by spec. The repeatable spec query names the specs to
// inventory -- one row each, recorded or not, a never-recorded name
// carrying a nil stamp and the project's whole change count; absent, the
// recorded specs alone.
func (h *specHandler) listSpecRuns(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	out, err := h.store.ListSpecLastRuns(r.Context(), project, r.URL.Query()["spec"])
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list spec last runs for %s", project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
