package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/selfreview"
	"github.com/tweety53/agents/stats/internal/store"
)

// selfreviewStore is the store dependency the self-review bundle endpoint
// needs, defined here at the consumer per go-interface-design — exactly the
// two reads the handler calls, so a test needs no database. The repository
// the archive-derived sources come from is not a store answer:
// change_repos carries no roots for the pipeline's real changes, and the
// caller — running from anywhere inside the repository, the way every flow
// command does — resolves the main checkout itself and passes it as the
// repo query parameter.
type selfreviewStore interface {
	RunRecord(ctx context.Context, projectKey, change string) (records.Run, error)

	// ChangeSummary is the change's recorded summary read — the row the
	// bundle serves as its first source. ErrChangeNotFound from it is the
	// missing-source case the bundle reports skipped, never a failure.
	ChangeSummary(ctx context.Context, projectKey, change string) (records.ChangeSummary, error)
}

// selfreviewHandler serves GET
// /api/v1/self-review/{project}/{change}/bundle?repo=<abs-path>: the whole
// self-review context bundle, assembled server-side — the ledger and panel
// record rendered from the store exactly as the render route renders them,
// the archived change's tasks.md, design.md and narrative.md read out of
// the chore/archive-<name> branch of the repository the caller named, and
// the git log of the finish-run commits, derived here rather than in any
// Bash the caller would have to run. The CLI transports the result and
// constructs none of it, the record-render rule.
type selfreviewHandler struct {
	store  selfreviewStore
	git    selfreview.Runner
	logger *slog.Logger
}

// bundle answers with text/markdown. repo is required: without it the
// route could only report the archive-derived sources skipped, which for
// an archived change reads as a false fact about the change rather than as
// the caller mistake it is — a 400 says so plainly. A change the store has
// never heard of is not an error — its ledger and panel sections report
// skipped and the rest of the bundle still serves, the gather's own "a
// missing source is never fatal" rule. Only a store read that fails for a
// real reason is a 5xx.
func (h *selfreviewHandler) bundle(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")
	repo := r.URL.Query().Get("repo")
	if repo == "" {
		writeError(w, http.StatusBadRequest, "repo is required: the absolute path of the repository the change's archive lives in")
		return
	}
	if !filepath.IsAbs(repo) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("repo %q is not an absolute path", repo))
		return
	}

	rec, err := h.store.RunRecord(r.Context(), project, change)
	if err != nil {
		if !errors.Is(err, store.ErrChangeNotFound) {
			status, msg := mapStoreError(h.logger, fmt.Sprintf("read the run record for %s/%s", project, change), err)
			writeError(w, status, msg)
			return
		}
		rec = records.Run{Change: change}
	}

	// The recorded summary is the bundle's first source, and its read
	// follows the same rule the run record's does: ErrChangeNotFound is
	// the change having no summary — reported skipped inside the bundle —
	// while any other read failure is a 5xx, never a bundle a caller
	// could mistake for the change's own.
	summary, summaryFound := "", false
	cs, err := h.store.ChangeSummary(r.Context(), project, change)
	switch {
	case err == nil:
		summary, summaryFound = cs.Summary, true
	case !errors.Is(err, store.ErrChangeNotFound):
		status, msg := mapStoreError(h.logger, fmt.Sprintf("read the change summary for %s/%s", project, change), err)
		writeError(w, status, msg)
		return
	}

	bundle, err := selfreview.Bundle(change, rec, summary, summaryFound, false, []string{repo}, h.git)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(bundle))
}
