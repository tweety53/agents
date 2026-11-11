package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

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
	PolicyStore
}

var _ ExperimentStore = (*store.Store)(nil)

// WithExperiments registers the /api/v1/experiments and /api/v1/policy
// routes against s, and /api/v1/roles over s's pairs and table -- the roles
// markdown, each role's group and Does line. The caller passes
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
// the roles table: such a cell would claim the role yet govern no dispatch.
// With no table wired, every role passes.
func (h *experimentsHandler) unknownRole(cells map[string][]string) (string, bool) {
	known := tableRoles(h.table)
	if known == nil {
		return "", true
	}
	for role := range cells {
		if !known[role] && role != "main" {
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

// promote serves POST /api/v1/experiments/{id}/promote ("Make permanent",
// decision promote-experiment): each cell of open experiment id but `main`
// -- no role row -- becomes that role's base pairs, its opt-in pairs kept,
// and the experiment closes positive with the body's optional summary, else
// a generated one. An unknown or closed id is 404.
// Replaying a promote whose close failed rewrites the same pairs.
func (h *experimentsHandler) promote(w http.ResponseWriter, r *http.Request) {
	id, ok := experimentID(w, r)
	if !ok {
		return
	}
	var dto struct {
		Summary string `json:"summary"`
	}
	if r.ContentLength != 0 {
		if err := decodeJSONBody(w, r, &dto); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	open, err := h.store.ListExperiments(r.Context(), true)
	if err != nil {
		h.experimentError(w, "promote experiment", err)
		return
	}
	i := slices.IndexFunc(open, func(e store.Experiment) bool { return e.ID == id })
	if i < 0 {
		h.experimentError(w, "promote experiment", fmt.Errorf("%w: %d", store.ErrExperimentNotOpen, id))
		return
	}
	e := open[i]
	pairs, err := h.store.RolePairs(r.Context())
	if err != nil {
		h.policyError(w, "promote experiment", err)
		return
	}
	roles := make([]string, 0, len(e.Cells))
	for role := range e.Cells {
		if role != "main" {
			roles = append(roles, role)
		}
	}
	slices.Sort(roles)
	for _, role := range roles {
		if err := h.store.ReplaceRolePairs(r.Context(), role, e.Cells[role], pairs[role].Optin); err != nil {
			h.policyError(w, "promote experiment", err)
			return
		}
	}
	summary := dto.Summary
	if summary == "" {
		summary = fmt.Sprintf("Made permanent: %s became the base pairs of %s.", e.Name, strings.Join(roles, ", "))
	}
	if err := h.store.CloseExperiment(r.Context(), id, "positive", summary); err != nil {
		h.experimentError(w, "promote experiment", err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}

// roleOverride is the open experiment cell replacing a role's base pairs.
type roleOverride struct {
	Pairs        []string `json:"pairs"`
	ExperimentID int64    `json:"experimentId"`
}

// roleEntry is one row of the roles table GET /api/v1/roles serves.
type roleEntry struct {
	Role     string             `json:"role"`
	Base     []string           `json:"base"`
	Optin    []string           `json:"optin"`
	Marked   []store.MarkedPair `json:"marked"`
	Group    string             `json:"group"`
	Does     string             `json:"does"`
	Override *roleOverride      `json:"override,omitempty"`
}

// roles serves GET /api/v1/roles: a leading `main` entry (no table row of
// its own), then one entry per row of the roles table in table order, its
// pairs from the store, its group and Does line from the table, and the open
// experiment cell that replaces its base pairs -- main's too.
func (h *experimentsHandler) roles(w http.ResponseWriter, r *http.Request) {
	rows, err := modelpolicy.Rows(h.table)
	if err != nil {
		h.experimentError(w, "parse roles table", err)
		return
	}
	pairs, err := h.store.RolePairs(r.Context())
	if err != nil {
		h.policyError(w, "list roles", err)
		return
	}
	open, err := h.store.ListExperiments(r.Context(), true)
	if err != nil {
		h.experimentError(w, "list roles", err)
		return
	}
	cell := map[string]*roleOverride{}
	for _, e := range open {
		for role, p := range e.Cells {
			cell[role] = &roleOverride{Pairs: p, ExperimentID: e.ID}
		}
	}
	out := []roleEntry{{Role: "main", Base: []string{}, Optin: []string{}, Marked: []store.MarkedPair{}, Group: "Session",
		Does: "The session you talk to: runs the pipeline and sends out every other role.", Override: cell["main"]}}
	for _, row := range rows {
		p := pairs[row.Role]
		e := roleEntry{Role: row.Role, Base: p.Base, Optin: p.Optin, Marked: p.Marked, Group: row.Group, Does: row.Does, Override: cell[row.Role]}
		if e.Base == nil {
			e.Base, e.Optin, e.Marked = []string{}, []string{}, []store.MarkedPair{}
		}
		out = append(out, e)
	}
	writeJSON(w, http.StatusOK, out)
}
