package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/lessons"
	"github.com/tweety53/agents/stats/internal/store"
)

// lessonsStore is the store dependency the lessons-resolve endpoint needs,
// defined here at the consumer per go-interface-design — exactly the one
// read the handler calls. The projects table is the store's one answer to
// where this machine's repositories live; the scan itself is filesystem
// work the daemon does, the way the self-review bundle reads git.
type lessonsStore interface {
	ProjectRoots(ctx context.Context) ([]store.ProjectRoot, error)
}

// lessonsHandler serves GET /api/v1/lessons/resolve?topic=<words>: the
// process-lesson answer assembled server-side — every registered project's
// docs/briefs/ and archived narratives scanned for the topic, briefs
// ranked before narrative mentions, rendered as one Markdown document.
// The CLI transports the result and constructs none of it, the
// record-render rule.
type lessonsHandler struct {
	store  lessonsStore
	logger *slog.Logger
}

// resolve answers with text/markdown. topic is required: an empty one
// would otherwise read as "everything", which is a search, not a resolve.
// A store read failure is a 5xx; an unanswerable scan is never a 5xx —
// a root that cannot be read is reported inside the answer.
func (h *lessonsHandler) resolve(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if lessons.Normalize(topic) == "" {
		writeError(w, http.StatusBadRequest, "topic is required: the words a ticket uses to name the practice or brief")
		return
	}

	roots, err := h.store.ProjectRoots(r.Context())
	if err != nil {
		status, msg := mapStoreError(h.logger, "read the registered project roots", err)
		writeError(w, status, msg)
		return
	}

	ls := make([]lessons.Root, len(roots))
	for i, root := range roots {
		ls[i] = lessons.Root{ProjectKey: root.ProjectKey, Path: root.MainCheckoutPath}
	}
	res, err := lessons.Resolve(ls, topic)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(lessons.Render(res)))
}
