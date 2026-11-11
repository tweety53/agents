package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
	"github.com/tweety53/agents/stats/internal/store"
)

// PolicyStore is the store dependency of the role-pairs endpoints (kan-1103,
// decision pairs-in-db): the store is the pairs' one source, the UI its one
// writer.
type PolicyStore interface {
	RolePairs(ctx context.Context) (map[string]store.RolePolicy, error)
	ReplaceRolePairs(ctx context.Context, role string, base, optin []string) error
	Grant(ctx context.Context, change string, roles []string) error
	Revoke(ctx context.Context, change string, roles []string) error
	Grants(ctx context.Context, change string) ([]string, error)
}

var _ PolicyStore = (*store.Store)(nil)

type rolePairsDTO struct {
	Base  []string `json:"base"`
	Optin []string `json:"optin"`
}

type grantDTO struct {
	Change string   `json:"change"`
	Roles  []string `json:"roles"`
}

// tableRoles is the role ids of the roles table, nil when it is unwired or
// unreadable -- every role then passes.
func tableRoles(table []byte) map[string]bool {
	if table == nil {
		return nil
	}
	rows, err := modelpolicy.Rows(table)
	if err != nil {
		return nil
	}
	out := map[string]bool{}
	for _, r := range rows {
		out[r.Role] = true
	}
	return out
}

// policyError writes err's status: a malformed pair or grant 400, a
// revoke of a grant never made 404, anything else mapStoreError's answer.
func (h *experimentsHandler) policyError(w http.ResponseWriter, action string, err error) {
	switch {
	case errors.Is(err, store.ErrInvalidPair):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, store.ErrGrantNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		status, msg := mapStoreError(h.logger, action, err)
		writeError(w, status, msg)
	}
}

// policy serves GET /api/v1/policy: {role: {base, optin, marked}}.
func (h *experimentsHandler) policy(w http.ResponseWriter, r *http.Request) {
	p, err := h.store.RolePairs(r.Context())
	if err != nil {
		h.policyError(w, "read policy", err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// putRole serves PUT /api/v1/policy/roles/{role}: replaces the role's base
// and opt-in pairs whole. A role the roles table lacks is 400, as is a
// malformed pair.
func (h *experimentsHandler) putRole(w http.ResponseWriter, r *http.Request) {
	role := r.PathValue("role")
	if known := tableRoles(h.table); known != nil && !known[role] {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("role %q is not in the Allowed pairs table", role))
		return
	}
	var dto rolePairsDTO
	if err := decodeJSONBody(w, r, &dto); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.ReplaceRolePairs(r.Context(), role, dto.Base, dto.Optin); err != nil {
		h.policyError(w, "replace role pairs", err)
		return
	}
	h.policy(w, r)
}

// grants serves GET /api/v1/policy/grants?change=<name>: the roles the
// change holds a live grant for.
func (h *experimentsHandler) grants(w http.ResponseWriter, r *http.Request) {
	change := r.URL.Query().Get("change")
	if change == "" {
		writeError(w, http.StatusBadRequest, "grants need ?change=<name>")
		return
	}
	roles, err := h.store.Grants(r.Context(), change)
	if err != nil {
		h.policyError(w, "read grants", err)
		return
	}
	if roles == nil {
		roles = []string{}
	}
	writeJSON(w, http.StatusOK, grantDTO{Change: change, Roles: roles})
}

// putGrant serves PUT /api/v1/policy/grants and DELETE /api/v1/policy/grants:
// grant or revoke {change, roles}. A grant naming a role the roles table
// lacks is 400.
func (h *experimentsHandler) putGrant(w http.ResponseWriter, r *http.Request) {
	var dto grantDTO
	if err := decodeJSONBody(w, r, &dto); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if dto.Change == "" || len(dto.Roles) == 0 {
		writeError(w, http.StatusBadRequest, "a grant names a change and at least one role")
		return
	}
	if known := tableRoles(h.table); known != nil {
		for _, role := range dto.Roles {
			if !known[role] {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("role %q is not in the Allowed pairs table", role))
				return
			}
		}
	}
	write, action := h.store.Grant, "grant"
	if r.Method == http.MethodDelete {
		write, action = h.store.Revoke, "revoke"
	}
	if err := write(r.Context(), dto.Change, dto.Roles); err != nil {
		h.policyError(w, action, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
