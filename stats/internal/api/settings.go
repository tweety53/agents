package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/store"
)

// ValidateSettings validates s against the harness's fixed model and
// reviewer-slot enums, re-exporting store.ValidateSettings rather than
// redefining the vocabularies here a second time -- so the PUT
// /api/v1/settings handler below and store.PutSettings itself can never
// silently diverge on what counts as a valid value. It returns
// store.ErrInvalidModel or store.ErrInvalidReviewer, wrapped with the
// specific bad value, for the first violation found.
func ValidateSettings(s store.Settings) error {
	return store.ValidateSettings(s)
}

// SettingsStore is the store dependency the settings endpoints need,
// defined here at the consumer per go-interface-design -- exactly the two
// methods settingsHandler calls, so a test needs no database.
type SettingsStore interface {
	GetSettings(ctx context.Context) (store.Settings, error)
	PutSettings(ctx context.Context, s store.Settings) error
}

// var _ SettingsStore = (*store.Store)(nil) verifies at compile time that
// the real store satisfies the interface this package actually depends
// on.
var _ SettingsStore = (*store.Store)(nil)

// settingsHandler serves the two settings endpoints. Both are thin: get
// reads flow_settings' single row (or the harness defaults, when none has
// ever been written) and put validates-then-writes it whole, mirroring
// changeHandler's get/put shape for the same reason -- one settings row,
// no partial update, exactly like one change record.
type settingsHandler struct {
	store  SettingsStore
	logger *slog.Logger
}

// settingsDTO is the wire representation of a store.Settings: the JSON
// shape GET and PUT exchange. Field names match store.Settings' own
// field-for-field, following changeDTO's own precedent.
type settingsDTO struct {
	DefaultModel    string   `json:"defaultModel"`
	SelfReviewModel string   `json:"selfReviewModel"`
	Reviewers       []string `json:"reviewers"`
}

func toSettingsDTO(s store.Settings) settingsDTO {
	return settingsDTO{
		DefaultModel:    s.DefaultModel,
		SelfReviewModel: s.SelfReviewModel,
		Reviewers:       s.Reviewers,
	}
}

// get serves GET /api/v1/settings.
func (h *settingsHandler) get(w http.ResponseWriter, r *http.Request) {
	s, err := h.store.GetSettings(r.Context())
	if err != nil {
		status, msg := mapStoreError(h.logger, "get settings", err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, toSettingsDTO(s))
}

// put serves PUT /api/v1/settings. It sends the whole record to the
// store -- store.PutSettings validates before writing anything, so an
// invalid defaultModel or reviewers entry is refused with 400, naming the
// rejected value, and never partially applied: PutSettings either writes
// the complete row or writes nothing at all.
func (h *settingsHandler) put(w http.ResponseWriter, r *http.Request) {
	var dto settingsDTO
	if err := decodeJSONBody(w, r, &dto); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s := store.Settings{
		DefaultModel:    dto.DefaultModel,
		SelfReviewModel: dto.SelfReviewModel,
		Reviewers:       dto.Reviewers,
	}
	if err := h.store.PutSettings(r.Context(), s); err != nil {
		if errors.Is(err, store.ErrInvalidModel) || errors.Is(err, store.ErrInvalidReviewer) {
			// A caller mistake, not a store failure -- mapStoreError's
			// generic default (500 "internal error") would swallow the
			// rejected value, exactly the failure mode stages.go's
			// ErrInvalidSessionToken handling avoids for a sessionToken.
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		status, msg := mapStoreError(h.logger, "put settings", err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, toSettingsDTO(s))
}
