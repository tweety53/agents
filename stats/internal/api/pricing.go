package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// Repricer is the store dependency the reprice endpoint needs, defined at
// the consumer. It is wired through WithRepricer, as GuardRunStore is.
type Repricer interface {
	RepriceSince(ctx context.Context, since time.Time) (priced, skipped int, err error)
}

var _ Repricer = (*store.Store)(nil)

// WithRepricer registers POST /api/v1/pricing/reprice against s.
func WithRepricer(s Repricer) Option {
	return func(o *serverOptions) { o.repricer = s }
}

type pricingHandler struct {
	store  Repricer
	logger *slog.Logger
}

// reprice serves POST /api/v1/pricing/reprice: Store.RepriceSince over
// every stage run started on or after {"since": "YYYY-MM-DD"} (UTC),
// answering {"priced": N, "skipped": M} (decision reprice-command).
func (h *pricingHandler) reprice(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Since string `json:"since"`
	}
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	since, err := time.Parse(time.DateOnly, in.Since)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("since %q is not YYYY-MM-DD", in.Since))
		return
	}
	priced, skipped, err := h.store.RepriceSince(r.Context(), since)
	if err != nil {
		status, msg := mapStoreError(h.logger, "reprice", err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"priced": priced, "skipped": skipped})
}
