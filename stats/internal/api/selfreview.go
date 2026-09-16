package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/selfreview"
	"github.com/tweety53/agents/stats/internal/store"
)

// selfreviewStore is the store dependency the self-review bundle endpoint
// needs, defined here at the consumer per go-interface-design — exactly
// the two reads the handler calls, so a test needs no database. Everything
// else the bundle carries is read from the change's repository through the
// runner, never from the store.
type selfreviewStore interface {
	RunRecord(ctx context.Context, projectKey, change string) (records.Run, error)
	ListChangeRepos(ctx context.Context, projectKey, change string) ([]store.Repo, error)
}

// selfreviewHandler serves GET
// /api/v1/self-review/{project}/{change}/bundle: the whole self-review
// context bundle, assembled server-side — the ledger and panel record
// rendered from the store exactly as the render route renders them, the
// archived change's tasks.md, design.md and narrative.md read out of the
// chore/archive-<name> branch, and the git log of the finish-run commits,
// derived here rather than in any Bash the caller would have to run. The
// CLI transports the result and constructs none of it, the record-render
// rule.
type selfreviewHandler struct {
	store  selfreviewStore
	git    selfreview.Runner
	logger *slog.Logger
}

// bundle answers with text/markdown. A change the store has never heard
// of is not an error — its ledger and panel sections report skipped and
// the rest of the bundle still serves, the gather's own "a missing source
// is never fatal" rule. Only a store read that fails for a real reason is
// a 5xx, and only an invalid change name is a 400: both mean the caller
// asked for something no bundle could answer, not that sources were
// absent.
func (h *selfreviewHandler) bundle(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")

	rec, err := h.store.RunRecord(r.Context(), project, change)
	if err != nil {
		if !errors.Is(err, store.ErrChangeNotFound) {
			status, msg := mapStoreError(h.logger, fmt.Sprintf("read the run record for %s/%s", project, change), err)
			writeError(w, status, msg)
			return
		}
		rec = records.Run{Change: change}
	}

	repos, err := h.store.ListChangeRepos(r.Context(), project, change)
	if err != nil {
		status, msg := mapStoreError(h.logger, fmt.Sprintf("list the change repos for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}

	roots := make([]string, 0, len(repos))
	for _, repo := range repos {
		roots = append(roots, repo.RepoRoot)
	}

	bundle, err := selfreview.Bundle(change, rec, roots, h.git)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(bundle))
}
