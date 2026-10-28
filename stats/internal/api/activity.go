package api

import (
	"fmt"
	"net/http"
)

// sessionActivityStates is the closed state vocabulary, mirrored by the
// session_activity table's CHECK constraint -- the API's validation is UX,
// the store's is the boundary.
var sessionActivityStates = map[string]bool{"working": true, "question": true, "permission": true, "idle": true}

// maxActivityDetailBytes caps detail. hooks/flow-session-activity.py
// truncates to it, so a longer detail is a caller bug and refused.
const maxActivityDetailBytes = 500

// putSessionActivity serves PUT /api/v1/sessions/{session}/activity:
// what the session is doing now, reported by
// hooks/flow-session-activity.py. 204 on success; 400 on an unknown state
// or a detail over maxActivityDetailBytes, before the store is touched.
func (h *statsHandler) putSessionActivity(w http.ResponseWriter, r *http.Request) {
	session := r.PathValue("session")

	var in struct {
		State  string `json:"state"`
		Detail string `json:"detail"`
	}
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !sessionActivityStates[in.State] {
		writeError(w, http.StatusBadRequest, "state must be one of working, question, permission, idle")
		return
	}
	if len(in.Detail) > maxActivityDetailBytes {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("detail must be at most %d bytes", maxActivityDetailBytes))
		return
	}

	if err := h.store.SetSessionActivity(r.Context(), session, in.State, in.Detail); err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("set activity for session %s", session), err)
		writeError(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
