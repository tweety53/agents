package api

// requests.go — the /flow-interactive request routes (KAN-1061): a begin
// and an end, each an upsert by (change, n) in the store, and the list.

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// ApplyRequestBegin records a request's receipt against rw -- what the live
// route and a journalled request-begin's replay both call. The store's
// ErrInvalidRequest is reported as ErrInvalidRecord, the caller mistake the
// route answers 400 for and replay retires.
func ApplyRequestBegin(ctx context.Context, rw RecordStore, projectKey, change string, in records.RequestBegin) (records.Request, error) {
	out, err := rw.BeginRequest(ctx, projectKey, change, in)
	return out, invalidRequestAsRecord(err)
}

// ApplyRequestEnd records a request's end against rw. See ApplyRequestBegin.
func ApplyRequestEnd(ctx context.Context, rw RecordStore, projectKey, change string, in records.RequestEnd) (records.Request, error) {
	out, err := rw.EndRequest(ctx, projectKey, change, in)
	return out, invalidRequestAsRecord(err)
}

func invalidRequestAsRecord(err error) error {
	if errors.Is(err, store.ErrInvalidRequest) {
		return fmt.Errorf("%w: %v", ErrInvalidRecord, err)
	}
	return err
}

// recordRequestBegin serves POST /api/v1/records/{project}/{change}/requests:
// 200 with the stored row.
func (h *recordHandler) recordRequestBegin(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")
	var in records.RequestBegin
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := ApplyRequestBegin(r.Context(), h.store, project, change, in)
	h.writeRequestResult(w, "begin", project, change, out, err)
}

// recordRequestEnd serves POST
// /api/v1/records/{project}/{change}/requests/end: 200 with the stored row.
func (h *recordHandler) recordRequestEnd(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")
	var in records.RequestEnd
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := ApplyRequestEnd(r.Context(), h.store, project, change, in)
	h.writeRequestResult(w, "end", project, change, out, err)
}

func (h *recordHandler) writeRequestResult(w http.ResponseWriter, verb, project, change string, out records.Request, err error) {
	if err != nil {
		if errors.Is(err, ErrInvalidRecord) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, fmt.Sprintf("%s request for %s/%s", verb, project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// listRequests serves GET /api/v1/records/{project}/{change}/requests: the
// change's requests in n order, each with its derived state.
func (h *recordHandler) listRequests(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")
	out, err := h.store.ListRequests(r.Context(), project, change)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list requests for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}
	if out == nil {
		out = []records.Request{}
	}
	writeJSON(w, http.StatusOK, out)
}
