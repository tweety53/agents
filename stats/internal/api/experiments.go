package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
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
	ExperimentStats(ctx context.Context, id int64) (store.ExperimentStatsResult, error)
}

var _ ExperimentStore = (*store.Store)(nil)

// WithExperiments registers the /api/v1/experiments routes against s, and
// /api/v1/roles over table -- the Allowed pairs markdown. The caller passes
// the table (flowd: modelpolicy/embedded.Table) so this package never
// imports the make-copied embed: internal/reconcile imports internal/api,
// and cmd/flow must build without make.
func WithExperiments(s ExperimentStore, table []byte) Option {
	return func(o *serverOptions) { o.experiments, o.rolesTable = s, table }
}

// experimentsHandler serves the experiment endpoints on settingsHandler's
// thin shape: decode, call the store, map its typed errors.
type experimentsHandler struct {
	store  ExperimentStore
	table  []byte
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
	case errors.Is(err, store.ErrExperimentNotOpen), errors.Is(err, store.ErrExperimentNotFound):
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
	if role, ok := h.unknownRole(dto.Cells); !ok {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("role %q is not in the Allowed pairs table", role))
		return
	}
	e, err := h.store.OpenExperiment(r.Context(), store.Experiment{Name: dto.Name, Cells: dto.Cells})
	if err != nil {
		h.experimentError(w, "open experiment", err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

// unknownRole names the first cell role that is neither `main` nor a row of
// the Allowed pairs table: such a cell would claim the role yet govern no
// dispatch, since the guards' overlay ignores roles the table lacks. With no
// table wired, every role passes.
func (h *experimentsHandler) unknownRole(cells map[string][]string) (string, bool) {
	if h.table == nil {
		return "", true
	}
	rows, err := modelpolicy.Parse(h.table)
	if err != nil {
		return "", true
	}
	for role := range cells {
		if _, ok := rows[role]; !ok && role != "main" {
			return role, false
		}
	}
	return "", true
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

// experimentID reads the {id} path value, writing 400 when it is not an
// integer.
func experimentID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "experiment id must be an integer, got "+strconv.Quote(r.PathValue("id")))
		return 0, false
	}
	return id, true
}

// stats serves GET /api/v1/experiments/{id}/stats: the experiment's
// before/after figures (store.ExperimentStats).
func (h *experimentsHandler) stats(w http.ResponseWriter, r *http.Request) {
	id, ok := experimentID(w, r)
	if !ok {
		return
	}
	out, err := h.store.ExperimentStats(r.Context(), id)
	if err != nil {
		h.experimentError(w, "experiment stats", err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// close serves POST /api/v1/experiments/{id}/close.
func (h *experimentsHandler) close(w http.ResponseWriter, r *http.Request) {
	id, ok := experimentID(w, r)
	if !ok {
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
// its own), then the Allowed pairs table one entry per row in
// table order, each unmarked row -- and main -- carrying the open
// experiment cell that replaces it. Marked rows (escalations) never do.
func (h *experimentsHandler) roles(w http.ResponseWriter, r *http.Request) {
	rows, err := modelpolicy.Rows(h.table)
	if err != nil {
		h.experimentError(w, "parse roles table", err)
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
