package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// confidenceStore is what the confidence-estimate routes (KAN-1130) read and
// write. New registers the routes when its RecordStore also satisfies this,
// which *store.Store does.
type confidenceStore interface {
	RecordConfidenceEstimate(ctx context.Context, projectKey string, in records.ConfidenceEstimate) (records.ConfidenceEstimate, error)
	ConfidenceTags(ctx context.Context) ([]records.ConfidenceTag, error)
	ConfidenceEstimatesPage(ctx context.Context, tag string, limit, offset int) ([]records.ConfidenceEstimate, int, error)
}

var _ confidenceStore = (*store.Store)(nil)

type confidenceHandler struct {
	store  confidenceStore
	logger *slog.Logger
}

// record serves POST /api/v1/confidence/{project}/{change}: one estimate
// below 100%. The change comes from the path, never the body. A store
// validation refusal is 400 through mapStoreError.
func (h *confidenceHandler) record(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")
	var in records.ConfidenceEstimate
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	in.Change = change
	out, err := h.store.RecordConfidenceEstimate(r.Context(), project, in)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("record confidence estimate for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// tags serves GET /api/v1/confidence-tags: every tag in use, most used first.
func (h *confidenceHandler) tags(w http.ResponseWriter, r *http.Request) {
	out, err := h.store.ConfidenceTags(r.Context())
	if err != nil {
		status, msg := mapStoreError(h.logger, "confidence tags", err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type confidencePageDTO struct {
	Total  int                          `json:"total"`
	Limit  int                          `json:"limit"`
	Offset int                          `json:"offset"`
	Tags   []records.ConfidenceTag      `json:"tags"`
	Rows   []records.ConfidenceEstimate `json:"rows"`
}

// page serves GET /api/v1/stats/confidence-estimates[?tag][&limit][&offset]:
// one page of the dashboard's list, newest first, with the list's total and
// every tag's count for the tag chips.
func (h *confidenceHandler) page(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	limit, offset := 25, 0
	for name, dst := range map[string]*int{"limit": &limit, "offset": &offset} {
		if raw := values.Get(name); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 0 || (name == "limit" && (n == 0 || n > selfReviewPageMax)) {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("%s must be a whole number (limit 1-%d, offset 0 or more), got %q", name, selfReviewPageMax, raw))
				return
			}
			*dst = n
		}
	}
	rows, total, err := h.store.ConfidenceEstimatesPage(r.Context(), values.Get("tag"), limit, offset)
	if err == nil {
		var tags []records.ConfidenceTag
		if tags, err = h.store.ConfidenceTags(r.Context()); err == nil {
			writeJSON(w, http.StatusOK, confidencePageDTO{Total: total, Limit: limit, Offset: offset, Tags: tags, Rows: rows})
			return
		}
	}
	status, msg := mapStoreError(h.logger, "confidence estimates page", err)
	writeError(w, status, msg)
}
