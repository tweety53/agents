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
	ConfidenceEstimatesPage(ctx context.Context, status string, limit, offset int) (store.ConfidencePage, error)
	ReviewConfidenceEstimate(ctx context.Context, projectKey string, id int64, status, reason string) error
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

// review serves PATCH /api/v1/confidence/{project}/estimates/{id}/review:
// the operator's approve or decline of an estimate, or its move back to
// to-review, with an optional reason. 204 on success, 404 for an id the
// project has no estimate for, 400 for a status outside the three.
func (h *confidenceHandler) review(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "estimate id must be a number")
		return
	}
	var in struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.ReviewConfidenceEstimate(r.Context(), project, id, in.Status, in.Reason); err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("review confidence estimate %d", id), err)
		writeError(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type confidencePageDTO struct {
	Total    int                          `json:"total"`
	ToReview int                          `json:"toReview"`
	Approved int                          `json:"approved"`
	Declined int                          `json:"declined"`
	Limit    int                          `json:"limit"`
	Offset   int                          `json:"offset"`
	Rows     []records.ConfidenceEstimate `json:"rows"`
}

// page serves GET /api/v1/stats/confidence-estimates?status[&limit][&offset]:
// one page of the dashboard's list -- the estimates carrying the review
// status, newest first -- with the list's total and every status's count for
// the status chips. A status outside the three is 400.
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
	p, err := h.store.ConfidenceEstimatesPage(r.Context(), values.Get("status"), limit, offset)
	if err != nil {
		status, msg := mapStoreError(h.logger, "confidence estimates page", err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, confidencePageDTO{Total: p.Total, ToReview: p.ToReview, Approved: p.Approved, Declined: p.Declined, Limit: limit, Offset: offset, Rows: p.Rows})
}
