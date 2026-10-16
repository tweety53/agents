// health.go serves the flow-health dashboard: the flow-summary view over
// store/health.go, and POST
// /api/v1/guard-runs/{project}, the write the flow-guard binary makes for
// every guard it runs.
package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// isHealthView reports whether name is the flow-summary view or the
// self-review view: neither counts a token measurement nor carries a model.
func isHealthView(name viewName) bool {
	return name == viewFlowSummary || name == viewSelfReview
}

// healthRows answers the flow-summary view and the self-review view. Neither
// takes a model restriction: a guard run and a self-review finding carry no
// model, and a change's cost, lead time and panel rounds are counted across
// every model that ran them, so a model filter is refused rather than
// silently ignored.
func (h *statsHandler) healthRows(ctx context.Context, name viewName, period store.Period, project, model *string) (rows any, status int, msg string) {
	if model != nil {
		return nil, http.StatusBadRequest, fmt.Sprintf("view %q counts across every model, so a model restriction cannot apply here", name)
	}
	switch name {
	case viewSelfReview:
		got, err := h.store.SelfReviewFindingsInPeriod(ctx, period, project)
		if err != nil {
			s, m := mapStoreError(h.logger, "self-review findings", err)
			return nil, s, m
		}
		return toSelfReviewDTOs(got, h.commitBase), 0, ""
	default: // viewFlowSummary
		got, err := h.store.FlowSummary(ctx, period, project)
		if err != nil {
			s, m := mapStoreError(h.logger, "flow summary", err)
			return nil, s, m
		}
		return flowSummaryDTO(got), 0, ""
	}
}

// flowSummaryDTO is the flow-summary view's one row -- an object, not a
// list: the view is a set of period totals (store.FlowSummaryRow).
type flowSummaryDTO struct {
	ChangesStarted      int      `json:"changesStarted"`
	ChangesFinished     int      `json:"changesFinished"`
	TotalCostUsd        *float64 `json:"totalCostUsd"`
	MeanCostPerFinished *float64 `json:"meanCostPerFinished"`
	MedianLeadTimeMs    *int64   `json:"medianLeadTimeMs"`
	MeanPanelRounds     *float64 `json:"meanPanelRounds"`
	Critical            int      `json:"critical"`
	Important           int      `json:"important"`
	Minor               int      `json:"minor"`
	GuardFires          int      `json:"guardFires"`
}

// GuardRunStore is the store dependency POST /api/v1/guard-runs/{project}
// needs, defined at the consumer. It is wired through WithGuardRuns rather
// than New's positional stores, so a server built without it -- every test
// that never exercises the route -- carries no fake for it.
type GuardRunStore interface {
	RecordGuardRun(ctx context.Context, projectKey string, in records.GuardRun) (records.GuardRun, error)
}

var _ GuardRunStore = (*store.Store)(nil)

// WithGuardRuns registers POST /api/v1/guard-runs/{project} against s.
func WithGuardRuns(s GuardRunStore) Option {
	return func(o *serverOptions) { o.guardRuns = s }
}

type guardRunHandler struct {
	store  GuardRunStore
	logger *slog.Logger
}

// record serves POST /api/v1/guard-runs/{project}: one guard invocation.
// An unknown project answers 404 and is not logged as a failure -- it is the
// expected answer for a guard run inside a test fixture's repository.
func (h *guardRunHandler) record(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	var in records.GuardRun
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := h.store.RecordGuardRun(r.Context(), project, in)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, out)
	case errors.Is(err, store.ErrProjectNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrInvalidGuardRun):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		status, msg := mapStoreError(h.logger, "record guard run for "+project, err)
		writeError(w, status, msg)
	}
}
