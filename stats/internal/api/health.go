// health.go serves the flow-health dashboard: the guard-health view over
// store/health.go -- one row per guard, helpers excluded. The runs it reads
// reach the store through flowd's ingestion of the file flow-guard appends
// them to (cmd/flowd/hookruns.go).
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	Helper       bool    `json:"helper"`
	Parent       string  `json:"parent,omitempty"`
}

// toGuardHealthDTOs maps rows to their DTOs -- a non-nil slice, so a window
// with no guard runs answers [] rather than null.
func toGuardHealthDTOs(rows []store.GuardHealthRow) []guardHealthDTO {
	out := make([]guardHealthDTO, len(rows))
	for i, r := range rows {
		out[i] = guardHealthDTO{Guard: r.Guard, Runs: r.Runs, Fired: r.Fired, CannotAnswer: r.CannotAnswer,
			MedianMs: r.MedianMs, TotalMs: r.TotalMs, Helper: r.Helper, Parent: r.Parent}
		if r.LastFiredAt != nil {
			at := r.LastFiredAt.UTC().Format(time.RFC3339)
			out[i].LastFiredAt = &at
		}
	}
	return out
}
