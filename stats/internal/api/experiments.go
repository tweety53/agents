package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
	"github.com/tweety53/agents/stats/internal/modelpolicy/embedded"
	"github.com/tweety53/agents/stats/internal/store"
)

// ExperimentStore is the store dependency the allowed-pairs experiment
// endpoints need (KAN-1067), defined at the consumer. It is wired through
// WithExperiments rather than New's positional stores, as GuardRunStore is.
type ExperimentStore interface {
	OpenExperiment(ctx context.Context, e store.Experiment) (store.Experiment, error)
	ListExperiments(ctx context.Context, openOnly bool) ([]store.Experiment, error)
	OpenOverrides(ctx context.Context) (map[string][]string, error)
	CloseExperiment(ctx context.Context, id int64, verdict, summary string) error
}

var _ ExperimentStore = (*store.Store)(nil)

// WithExperiments registers the /api/v1/experiments routes against s.
func WithExperiments(s ExperimentStore) Option {
	return func(o *serverOptions) { o.experiments = s }
}

// experimentsHandler serves the experiment endpoints on settingsHandler's
// thin shape: decode, call the store, map its typed errors.
type experimentsHandler struct {
	store  ExperimentStore
	logger *slog.Logger
}

type openExperimentDTO struct {
	Name  string              `json:"name"`
	Cells map[string][]string `json:"cells"`
}

type closeExperimentDTO struct {
	Verdict string `json:"verdict"`
	Summary string `json:"summary"`
}

// experimentError writes err's status: a caller mistake is 400, a role
// another open experiment holds 409, an unknown or closed id 404, anything
// else mapStoreError's answer.
func (h *experimentsHandler) experimentError(w http.ResponseWriter, action string, err error) {
	switch {
	case errors.Is(err, store.ErrInvalidExperiment), errors.Is(err, store.ErrInvalidVerdict):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, store.ErrExperimentConflict):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, store.ErrExperimentNotOpen):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		status, msg := mapStoreError(h.logger, action, err)
		writeError(w, status, msg)
	}
}

// open serves POST /api/v1/experiments: 201 with the opened experiment.
func (h *experimentsHandler) open(w http.ResponseWriter, r *http.Request) {
	var dto openExperimentDTO
	if err := decodeJSONBody(w, r, &dto); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	e, err := h.store.OpenExperiment(r.Context(), store.Experiment{Name: dto.Name, Cells: dto.Cells})
	if err != nil {
		h.experimentError(w, "open experiment", err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

// list serves GET /api/v1/experiments[?open=1], newest first.
func (h *experimentsHandler) list(w http.ResponseWriter, r *http.Request) {
	list, err := h.store.ListExperiments(r.Context(), r.URL.Query().Get("open") == "1")
	if err != nil {
		h.experimentError(w, "list experiments", err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// overrides serves GET /api/v1/experiments/overrides: {role: [pairs]} of
// every open experiment, {} when none is open.
func (h *experimentsHandler) overrides(w http.ResponseWriter, r *http.Request) {
	o, err := h.store.OpenOverrides(r.Context())
	if err != nil {
		h.experimentError(w, "experiment overrides", err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

// close serves POST /api/v1/experiments/{id}/close.
func (h *experimentsHandler) close(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "experiment id must be an integer, got "+strconv.Quote(r.PathValue("id")))
		return
	}
	var dto closeExperimentDTO
	if err := decodeJSONBody(w, r, &dto); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.CloseExperiment(r.Context(), id, dto.Verdict, dto.Summary); err != nil {
		h.experimentError(w, "close experiment", err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}

// roleOverride is the open experiment cell replacing a role's unmarked row.
type roleOverride struct {
	Pairs        []string `json:"pairs"`
	ExperimentID int64    `json:"experimentId"`
}

// roleEntry is one row of the roles table GET /api/v1/roles serves.
type roleEntry struct {
	Role     string        `json:"role"`
	Pairs    []string      `json:"pairs"`
	Marker   string        `json:"marker"`
	Override *roleOverride `json:"override,omitempty"`
}

// roles serves GET /api/v1/roles: a leading `main` entry (no table row of
// its own), then the embedded Allowed pairs table one entry per row in
// table order, each unmarked row -- and main -- carrying the open
// experiment cell that replaces it. Marked rows (escalations) never do.
func (h *experimentsHandler) roles(w http.ResponseWriter, r *http.Request) {
	rows, err := modelpolicy.Rows(embedded.Table)
	if err != nil {
		h.experimentError(w, "parse embedded roles table", err)
		return
	}
	open, err := h.store.ListExperiments(r.Context(), true)
	if err != nil {
		h.experimentError(w, "list roles", err)
		return
	}
	cell := map[string]*roleOverride{}
	for _, e := range open {
		for role, pairs := range e.Cells {
			cell[role] = &roleOverride{Pairs: pairs, ExperimentID: e.ID}
		}
	}
	out := []roleEntry{{Role: "main", Pairs: []string{}, Override: cell["main"]}}
	for _, row := range rows {
		e := roleEntry{Role: row.Role, Pairs: row.Pairs, Marker: row.Marker}
		if row.Marker == "" {
			e.Override = cell[row.Role]
		}
		out = append(out, e)
	}
	writeJSON(w, http.StatusOK, out)
}
