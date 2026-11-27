package api

import (
	"cmp"
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"github.com/tweety53/agents/stats/internal/store"
)

// BottlenecksStore backs GET /api/v1/stats/bottlenecks.
type BottlenecksStore interface {
	Bottlenecks(ctx context.Context, project *string, last int) ([]store.BottleneckRow, int64, error)
}

// WithBottlenecks registers GET /api/v1/stats/bottlenecks?last=N&project=P
// against s: the wall-clock ranking `flow stats bottlenecks` prints.
func WithBottlenecks(s BottlenecksStore) Option {
	return func(o *serverOptions) { o.bottlenecks = s }
}

// Bottleneck is one ranked row: Share is TotalMs over the window's summed
// wall clock (dimensions overlap, so shares across them do not add to 1).
type Bottleneck struct {
	Dimension string  `json:"dimension"`
	Name      string  `json:"name"`
	Runs      int     `json:"runs"`
	TotalMs   int64   `json:"totalMs"`
	Share     float64 `json:"share"`
	P50Ms     int64   `json:"p50Ms"`
	P95Ms     int64   `json:"p95Ms"`
}

// BottlenecksResponse is the ranking: the top 10 per dimension in
// Dimensions' order, and the top 5 rows by total across all of them.
type BottlenecksResponse struct {
	Last       int                     `json:"last"`
	WallMs     int64                   `json:"wallMs"`
	Dimensions []string                `json:"dimensions"`
	Ranked     map[string][]Bottleneck `json:"ranked"`
	Sinks      []Bottleneck            `json:"sinks"`
}

// bottleneckDimensions is the order the ranking prints its tables in.
var bottleneckDimensions = []string{"stage", "role", "role-class", "guard", "verify"}

type bottlenecksHandler struct {
	store BottlenecksStore
	log   *slog.Logger
}

func (h *bottlenecksHandler) get(w http.ResponseWriter, r *http.Request) {
	last := 20
	if v := r.URL.Query().Get("last"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "last must be a positive integer")
			return
		}
		last = n
	}
	var project *string
	if v := r.URL.Query().Get("project"); v != "" {
		project = &v
	}
	rows, wall, err := h.store.Bottlenecks(r.Context(), project, last)
	if err != nil {
		h.log.Error("bottlenecks", "err", err)
		writeError(w, http.StatusInternalServerError, "bottlenecks query failed")
		return
	}
	writeJSON(w, http.StatusOK, rankBottlenecks(rows, wall, last))
}

// rankBottlenecks keeps each dimension's 10 largest rows (the store orders
// them by total) and the 5 largest across dimensions, role-class excluded
// from the sinks since it re-counts the role rows.
func rankBottlenecks(rows []store.BottleneckRow, wall int64, last int) BottlenecksResponse {
	out := BottlenecksResponse{Last: last, WallMs: wall, Dimensions: bottleneckDimensions, Ranked: map[string][]Bottleneck{}, Sinks: []Bottleneck{}}
	for _, r := range rows {
		b := Bottleneck{Dimension: r.Dimension, Name: r.Name, Runs: r.Runs, TotalMs: r.TotalMs, P50Ms: r.P50Ms, P95Ms: r.P95Ms}
		if wall > 0 {
			b.Share = float64(r.TotalMs) / float64(wall)
		}
		if len(out.Ranked[r.Dimension]) < 10 {
			out.Ranked[r.Dimension] = append(out.Ranked[r.Dimension], b)
		}
		if r.Dimension != "role-class" {
			out.Sinks = append(out.Sinks, b)
		}
	}
	slices.SortStableFunc(out.Sinks, func(a, b Bottleneck) int { return cmp.Compare(b.TotalMs, a.TotalMs) })
	out.Sinks = out.Sinks[:min(5, len(out.Sinks))]
	return out
}
