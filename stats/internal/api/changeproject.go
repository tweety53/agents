package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tweety53/agents/stats/internal/store"
)

// changeFinder is the one store call ChangeProject needs.
type changeFinder interface {
	FindChangesByName(ctx context.Context, name string) ([]store.Change, error)
}

// ChangeProject returns the project a stage mark or record write for the
// change named name lands on. A caller sends the project of its own
// checkout, and a /flow-interactive executor working in another of the
// operator's repositories sends that repository's project instead of the
// change's. The rule: projectKey when it holds a row of that name (in any
// state -- the same name legitimately in two projects stays where the
// caller is); else the one other project holding an unfinished row of that
// name; else projectKey -- several unfinished rows elsewhere are never
// guessed between, and an unknown name keeps the begin mark's synthetic
// bootstrap under the caller's project.
func ChangeProject(ctx context.Context, find changeFinder, projectKey, name string) (string, error) {
	if name == "" {
		return projectKey, nil
	}
	rows, err := find.FindChangesByName(ctx, name)
	if err != nil {
		return "", err
	}
	other := ""
	unfinished := 0
	for _, c := range rows {
		if c.ProjectKey == projectKey {
			return projectKey, nil
		}
		if c.State != store.StateFinished {
			other = c.ProjectKey
			unfinished++
		}
	}
	if unfinished == 1 {
		return other, nil
	}
	return projectKey, nil
}

// withChangeProject re-keys a {project}/{change} route's project path value
// by ChangeProject before h runs.
func withChangeProject(find changeFinder, logger *slog.Logger, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		project, err := ChangeProject(r.Context(), find, r.PathValue("project"), r.PathValue("change"))
		if err != nil {
			status, msg := mapStoreError(logger, "resolve the project of "+r.PathValue("change"), err)
			writeError(w, status, msg)
			return
		}
		r.SetPathValue("project", project)
		h(w, r)
	}
}
