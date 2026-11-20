package api

import (
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
)

// WithPolicy registers the /api/v1/policy routes against s, and
// /api/v1/roles over s's pairs and table -- the roles markdown, each role's
// group and Does line. The caller passes the table (flowd:
// modelpolicy/embedded.Table) so this package never imports the make-copied
// embed: internal/reconcile imports internal/api, and cmd/flow must build
// without make.
func WithPolicy(s PolicyStore, table []byte) Option {
	return func(o *serverOptions) { o.policy, o.rolesTable = s, table }
}

// policyHandler serves the role-pairs endpoints on settingsHandler's thin
// shape: decode, call the store, map its typed errors.
type policyHandler struct {
	store  PolicyStore
	table  []byte
	logger *slog.Logger
}

// roleEntry is one row of the roles table GET /api/v1/roles serves.
type roleEntry struct {
	Role  string   `json:"role"`
	Base  []string `json:"base"`
	Optin []string `json:"optin"`
	Group string   `json:"group"`
	Does  string   `json:"does"`
}

// roles serves GET /api/v1/roles: a leading `main` entry (no table row of
// its own), then one entry per row of the roles table in table order, its
// pairs from the store, its group and Does line from the table.
func (h *policyHandler) roles(w http.ResponseWriter, r *http.Request) {
	rows, err := modelpolicy.Rows(h.table)
	if err != nil {
		h.policyError(w, "parse roles table", err)
		return
	}
	pairs, err := h.store.RolePairs(r.Context())
	if err != nil {
		h.policyError(w, "list roles", err)
		return
	}
	out := []roleEntry{{Role: "main", Base: []string{}, Optin: []string{}, Group: "Session",
		Does: "The session you talk to: runs the pipeline and sends out every other role."}}
	for _, row := range rows {
		p := pairs[row.Role]
		e := roleEntry{Role: row.Role, Base: p.Base, Optin: p.Optin, Group: row.Group, Does: row.Does}
		if e.Base == nil {
			e.Base, e.Optin = []string{}, []string{}
		}
		out = append(out, e)
	}
	writeJSON(w, http.StatusOK, out)
}
