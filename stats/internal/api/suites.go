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
// suite or host, a negative duration, or a test row with no package, no
// test name or a negative cost (KAN-1124) before the store is touched. See
// ApplyHazardRecord for why it takes RecordStore.
func ApplySuiteRunRecord(ctx context.Context, rw RecordStore, projectKey string, in records.SuiteRun) (records.SuiteRun, error) {
	if in.Suite == "" || in.Host == "" || in.DurationMs < 0 {
		return records.SuiteRun{}, fmt.Errorf("%w: suite, host and a non-negative duration are required", ErrInvalidRecord)
	}
	for _, test := range in.Tests {
		if test.Package == "" || test.Test == "" || test.CostMs < 0 {
			return records.SuiteRun{}, fmt.Errorf("%w: every test needs a package, a name and a non-negative cost, got %q in %q cost %d", ErrInvalidRecord, test.Test, test.Package, test.CostMs)
		}
	}
	return rw.InsertSuiteRun(ctx, projectKey, in)
}

// suiteHandler serves the three suite-run endpoints. Thin like hazardHandler:
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
	limit, ok := suiteRunLimit(w, r)
	if !ok {
		return
	}

	out, err := h.store.ListSuiteRuns(r.Context(), project, suite, limit)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list suite runs for %s", project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// listSuiteTests serves GET /api/v1/suites/{project}/tests: the newest run
// of one suite that carries per-test costs, its tests heaviest first and
// capped at limit (KAN-1124, decision suite-tests-read). suite is
// required; limit follows listSuiteRuns' rule. No such run answers 200
// with a null body -- an empty answer, never a missing route.
func (h *suiteHandler) listSuiteTests(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	suite := r.URL.Query().Get("suite")
	if suite == "" {
		writeError(w, http.StatusBadRequest, "suite is required")
		return
	}
	limit, ok := suiteRunLimit(w, r)
	if !ok {
		return
	}

	out, err := h.store.LatestSuiteRunTests(r.Context(), project, suite, limit)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("read %s test costs for %s", suite, project), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// suiteRunLimit reads the optional limit query both suite reads take:
// absent is defaultSuiteRunLimit, and anything but a positive integer is
// answered 400 here and reported as not ok.
func suiteRunLimit(w http.ResponseWriter, r *http.Request) (int, bool) {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return defaultSuiteRunLimit, true
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		writeError(w, http.StatusBadRequest, "limit must be a positive integer")
		return 0, false
	}
	return parsed, true
}
