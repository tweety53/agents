// health.go serves the flow-health dashboard: the guard-health view over
// store/health.go -- one row per guard, helpers excluded -- and POST
// /api/v1/guard-runs/{project}, the write the flow-guard binary makes for
// every guard it runs.
package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// isHealthView reports whether name is the guard-health view or the
// self-review view: neither counts a token measurement nor carries a model.
func isHealthView(name viewName) bool {
	return name == viewGuardHealth || name == viewSelfReview
}

// healthRows answers the guard-health view and the self-review view.
// Neither takes a model restriction: a guard run and a self-review finding
// carry no model, so a model filter is refused rather than silently ignored.
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
	default: // viewGuardHealth
		got, err := h.store.GuardHealth(ctx, period, project)
		if err != nil {
			s, m := mapStoreError(h.logger, "guard health", err)
			return nil, s, m
		}
		return toGuardHealthDTOs(got), 0, ""
	}
}

// guardHealthDTO is one guard-health row (store.GuardHealthRow). LastFiredAt
// is RFC 3339 in UTC, or null when the guard never fired in the window.
type guardHealthDTO struct {
	Guard        string  `json:"guard"`
	Runs         int     `json:"runs"`
	Fired        int     `json:"fired"`
	CannotAnswer int     `json:"cannotAnswer"`
	MedianMs     int64   `json:"medianMs"`
	TotalMs      int64   `json:"totalMs"`
	LastFiredAt  *string `json:"lastFiredAt"`
}

// toGuardHealthDTOs maps rows to their DTOs -- a non-nil slice, so a window
// with no guard runs answers [] rather than null.
func toGuardHealthDTOs(rows []store.GuardHealthRow) []guardHealthDTO {
	out := make([]guardHealthDTO, len(rows))
	for i, r := range rows {
		out[i] = guardHealthDTO{Guard: r.Guard, Runs: r.Runs, Fired: r.Fired, CannotAnswer: r.CannotAnswer,
			MedianMs: r.MedianMs, TotalMs: r.TotalMs}
		if r.LastFiredAt != nil {
			at := r.LastFiredAt.UTC().Format(time.RFC3339)
			out[i].LastFiredAt = &at
		}
	}
	return out
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
