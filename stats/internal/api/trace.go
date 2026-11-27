package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TraceStore backs GET /api/v1/stats/trace.
type TraceStore interface {
	TraceSpans(ctx context.Context, projectKey, change string) ([]store.TraceSpan, error)
}

// WithTrace registers GET /api/v1/stats/trace?project=P&change=C against
// s: the timed rows `flow stats trace` turns into a Chrome trace.
func WithTrace(s TraceStore) Option {
	return func(o *serverOptions) { o.trace = s }
}

// TraceSpan is store.TraceSpan on the wire; End is null while open.
type TraceSpan struct {
	Kind    string          `json:"kind"`
	Lane    string          `json:"lane"`
	Name    string          `json:"name"`
	AgentID string          `json:"agentId,omitempty"`
	Outcome string          `json:"outcome,omitempty"`
	Start   time.Time       `json:"start"`
	End     *time.Time      `json:"end,omitempty"`
	Split   json.RawMessage `json:"split,omitempty"`
}

type traceHandler struct {
	store TraceStore
	log   *slog.Logger
}

func (h *traceHandler) get(w http.ResponseWriter, r *http.Request) {
	project, change := r.URL.Query().Get("project"), r.URL.Query().Get("change")
	if project == "" || change == "" {
		writeError(w, http.StatusBadRequest, "project and change are required")
		return
	}
	spans, err := h.store.TraceSpans(r.Context(), project, change)
	if errors.Is(err, store.ErrChangeNotFound) {
		writeError(w, http.StatusNotFound, "no such change")
		return
	}
	if err != nil {
		h.log.Error("trace", "err", err)
		writeError(w, http.StatusInternalServerError, "trace query failed")
		return
	}
	out := make([]TraceSpan, len(spans))
	for i, s := range spans {
		out[i] = TraceSpan(s)
	}
	writeJSON(w, http.StatusOK, out)
}
