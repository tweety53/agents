// skillloads.go serves the skill-load telemetry (0111_skill_loads.sql): POST /api/v1/skill-loads,
// the flow-skill-load hook's write, and GET /api/v1/skill-loads?from&to, the per-path counts
// `flow-guard prune-candidates` reads.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// SkillLoadStore is the store dependency the skill-load routes need, wired through WithSkillLoads
// as GuardRunStore is.
type SkillLoadStore interface {
	RecordSkillLoad(ctx context.Context, path, sessionID string, loadedAt time.Time) error
	SkillLoadCounts(ctx context.Context, from, to time.Time) ([]store.SkillLoadCount, error)
}

var _ SkillLoadStore = (*store.Store)(nil)

// WithSkillLoads registers the skill-load routes against s.
func WithSkillLoads(s SkillLoadStore) Option {
	return func(o *serverOptions) { o.skillLoads = s }
}

type skillLoadHandler struct {
	store  SkillLoadStore
	logger *slog.Logger
}

// record serves POST /api/v1/skill-loads {"path", "sessionId"}: 204, or 400 on an empty path.
func (h *skillLoadHandler) record(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Path      string `json:"path"`
		SessionID string `json:"sessionId"`
	}
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if in.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}
	if err := h.store.RecordSkillLoad(r.Context(), in.Path, in.SessionID, time.Time{}); err != nil {
		status, msg := mapStoreError(h.logger, "record skill load", err)
		writeError(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type skillLoadRowDTO struct {
	Path  string `json:"path"`
	Loads int    `json:"loads"`
}

// counts serves GET /api/v1/skill-loads?from&to (RFC 3339): {"rows": [{"path", "loads"}]}, every
// path loaded in the window, ordered by path.
func (h *skillLoadHandler) counts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	period, _, err := parsePeriodAndProject(r.Context(), nil, map[string][]string{"from": q["from"], "to": q["to"]})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	got, err := h.store.SkillLoadCounts(r.Context(), period.From, period.To)
	if err != nil {
		status, msg := mapStoreError(h.logger, "skill load counts", err)
		writeError(w, status, msg)
		return
	}
	rows := make([]skillLoadRowDTO, len(got))
	for i, c := range got {
		rows[i] = skillLoadRowDTO(c)
	}
	writeJSON(w, http.StatusOK, map[string]any{"rows": rows})
}
