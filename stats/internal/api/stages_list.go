package api

import (
	"net/http"

	"github.com/tweety53/agents/stats/internal/stages"
)

// listStages backs GET /api/v1/stages?command=<command>: the command's stage
// keys in the order a run marks them, so the run view can draw the stages a
// change has still to run without a second copy of the stage table.
func listStages(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Query().Get("command")
	if command == "" {
		writeError(w, http.StatusBadRequest, "command is required")
		return
	}
	writeJSON(w, http.StatusOK, map[string][]string{"stages": stages.ForCommand(stages.Command(command))})
}
