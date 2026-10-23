package api

import (
	"errors"
	"net/http"

	"github.com/tweety53/agents/stats/internal/store"
)

// learningTopicDTO is the wire shape POST /api/v1/learning accepts: one
// learning note shown by a run (topics-in-flowd).
type learningTopicDTO struct {
	Topic      string `json:"topic"`
	Note       string `json:"note"`
	Change     string `json:"change"`
	ProjectKey string `json:"projectKey"`
}

// learningTopicCountDTO is one element of GET /api/v1/learning's array.
type learningTopicCountDTO struct {
	Topic    string `json:"topic"`
	Count    int    `json:"count"`
	LastNote string `json:"lastNote"`
}

// listLearning serves GET /api/v1/learning: every recorded topic with its
// count and newest note.
func (h *settingsHandler) listLearning(w http.ResponseWriter, r *http.Request) {
	rows, err := h.store.ListLearningTopics(r.Context())
	if err != nil {
		status, msg := mapStoreError(h.logger, "list learning topics", err)
		writeError(w, status, msg)
		return
	}
	out := make([]learningTopicCountDTO, 0, len(rows))
	for _, c := range rows {
		out = append(out, learningTopicCountDTO{Topic: c.Topic, Count: c.Count, LastNote: c.LastNote})
	}
	writeJSON(w, http.StatusOK, out)
}

// addLearning serves POST /api/v1/learning. A malformed topic or empty
// note is 400 naming the value; a topic already shown
// store.MaxLearningTopicShows times is 409 (cap-in-store).
func (h *settingsHandler) addLearning(w http.ResponseWriter, r *http.Request) {
	var dto learningTopicDTO
	if err := decodeJSONBody(w, r, &dto); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	in := store.LearningTopic{Topic: dto.Topic, Note: dto.Note, Change: dto.Change, ProjectKey: dto.ProjectKey}
	if err := h.store.RecordLearningTopic(r.Context(), in); err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidLearningTopic):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, store.ErrLearningTopicExhausted):
			writeError(w, http.StatusConflict, err.Error())
		default:
			status, msg := mapStoreError(h.logger, "record learning topic", err)
			writeError(w, status, msg)
		}
		return
	}
	writeJSON(w, http.StatusCreated, dto)
}
