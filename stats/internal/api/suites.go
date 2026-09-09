package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/tweety53/agents/stats/internal/records"
)

// defaultSuiteRunLimit is what GET .../runs answers with when the caller
// names no limit: enough recent rows for the CLI's per-(suite, host)
// summary to compute over without the caller having to know a count.
const defaultSuiteRunLimit = 20

// ApplySuiteRunRecord records one suite run against rw, refusing an empty
// suite or host or a negative duration before the store is touched. See
// ApplyHazardRecord for why it takes RecordStore.
func ApplySuiteRunRecord(ctx context.Context, rw RecordStore, projectKey string, in records.SuiteRun) (records.SuiteRun, error) {
	if in.Suite == "" || in.Host == "" || in.DurationMs < 0 {
		return records.SuiteRun{}, fmt.Errorf("%w: suite, host and a non-negative duration are required", ErrInvalidRecord)
	}
	return rw.InsertSuiteRun(ctx, projectKey, in)
}

// suiteHandler serves the two suite-run endpoints. Thin like hazardHandler:
// decode, refuse what cannot identify what it describes, hand the typed
// record to the store.
type suiteHandler struct {
	store  RecordStore
	logger *slog.Logger
}

// recordSuiteRun serves POST /api/v1/suites/{project}/runs: one timed suite
// execution, recorded per run per machine. suite, host and a non-negative
// duration are refused before the store is touched; an unknown project is
// the foreign key answering, mapped like any other store refusal.
func (h *suiteHandler) recordSuiteRun(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	var in records.SuiteRun
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := ApplySuiteRunRecord(r.Context(), h.store, project, in)
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("record suite run for %s", project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// listSuiteRuns serves GET /api/v1/suites/{project}/runs: the project's
// recorded suite runs, newest first. The optional suite query filters to
// one suite exactly as store.ListSuiteRuns does; absent, it is the
// operator's unfiltered listing. limit defaults to defaultSuiteRunLimit; a
// limit that is not a positive integer is a caller mistake, a 400.
func (h *suiteHandler) listSuiteRuns(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	suite := r.URL.Query().Get("suite")
	limit := defaultSuiteRunLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limit = parsed
	}

	out, err := h.store.ListSuiteRuns(r.Context(), project, suite, limit)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list suite runs for %s", project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
