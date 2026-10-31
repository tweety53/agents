package api

import (
	"fmt"
	"net/http"
)

// stageCompleted serves GET
// /api/v1/stages/{project}/{change}/{stage}/completed: {"completed": bool},
// whether the change holds a run of the stage that ended completed -- the
// fact integrate's land step checks for flow.self-review (KAN-1073). An
// unknown change has no such run and answers false.
func (h *recordHandler) stageCompleted(w http.ResponseWriter, r *http.Request) {
	project, change, stage := r.PathValue("project"), r.PathValue("change"), r.PathValue("stage")
	done, err := h.store.StageCompleted(r.Context(), project, change, stage)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("read stage %s for %s/%s", stage, project, change), err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"completed": done})
}
