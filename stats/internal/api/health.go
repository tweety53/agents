// health.go serves the flow-health dashboard: the guards, stage-redo and
// panel-rounds views over store/health.go, and POST
// /api/v1/guard-runs/{project}, the write the flow-guard binary makes for
// every guard it runs.
package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// isHealthView reports whether name is one of the flow-health views.
func isHealthView(name viewName) bool {
	return name == viewGuards || name == viewStageRedo || name == viewPanelRounds
}

// healthRows answers the three flow-health views. None takes a model
// restriction: a guard run carries no model, and a stage's re-entries and a
// change's panel rounds are counted across every model that ran them, so a
// model filter is refused rather than silently ignored -- the runs view's
// posture.
func (h *statsHandler) healthRows(ctx context.Context, name viewName, period store.Period, project, model *string) (rows any, status int, msg string) {
	if model != nil {
		return nil, http.StatusBadRequest, fmt.Sprintf("view %q counts across every model, so a model restriction cannot apply here", name)
	}
	switch name {
	case viewGuards:
		got, err := h.store.GuardActivity(ctx, period, project)
		if err != nil {
			s, m := mapStoreError(h.logger, "guard activity", err)
			return nil, s, m
		}
		return toGuardActivityDTOs(got), 0, ""
	case viewStageRedo:
		got, err := h.store.StageRedo(ctx, period, project)
		if err != nil {
			s, m := mapStoreError(h.logger, "stage redo", err)
			return nil, s, m
		}
		return toStageRedoDTOs(got), 0, ""
	default: // viewPanelRounds
		got, err := h.store.PanelRounds(ctx, period, project)
		if err != nil {
			s, m := mapStoreError(h.logger, "panel rounds", err)
			return nil, s, m
		}
		return toPanelRoundsDTOs(got), 0, ""
	}
}

type guardActivityRowDTO struct {
	Guard            string   `json:"guard"`
	Runs             int      `json:"runs"`
	Fired            int      `json:"fired"`
	CannotAnswer     int      `json:"cannotAnswer"`
	LastRunAt        *string  `json:"lastRunAt"`
	LastFiredAt      *string  `json:"lastFiredAt"`
	MedianDurationMS *float64 `json:"medianDurationMs"`
	Verdicts         int      `json:"verdicts"`
	FalsePositives   int      `json:"falsePositives"`
}

func toGuardActivityDTOs(rows []store.GuardActivityRow) []guardActivityRowDTO {
	out := make([]guardActivityRowDTO, len(rows))
	for i, r := range rows {
		out[i] = guardActivityRowDTO{
			Guard: r.Guard, Runs: r.Runs, Fired: r.Fired, CannotAnswer: r.CannotAnswer,
			LastRunAt: rfc3339Ptr(r.LastRunAt), LastFiredAt: rfc3339Ptr(r.LastFiredAt),
			MedianDurationMS: r.MedianDurationMS,
			Verdicts:         r.Verdicts, FalsePositives: r.FalsePositives,
		}
	}
	return out
}

type stageRedoRowDTO struct {
	Command          string   `json:"command"`
	Stage            string   `json:"stage"`
	Runs             int      `json:"runs"`
	Changes          int      `json:"changes"`
	Reentries        int      `json:"reentries"`
	ReenteredChanges int      `json:"reenteredChanges"`
	MedianSeconds    *float64 `json:"medianSeconds"`
	P90Seconds       *float64 `json:"p90Seconds"`
}

func toStageRedoDTOs(rows []store.StageRedoRow) []stageRedoRowDTO {
	out := make([]stageRedoRowDTO, len(rows))
	for i, r := range rows {
		out[i] = stageRedoRowDTO(r)
	}
	return out
}

type panelRoundsRowDTO struct {
	Project   string `json:"project"`
	Change    string `json:"change"`
	StartedAt string `json:"startedAt"`
	Rounds    *int   `json:"rounds"`
	Findings  int    `json:"findings"`
	Critical  int    `json:"critical"`
	Important int    `json:"important"`
	Minor     int    `json:"minor"`
}

func toPanelRoundsDTOs(rows []store.PanelRoundsRow) []panelRoundsRowDTO {
	out := make([]panelRoundsRowDTO, len(rows))
	for i, r := range rows {
		out[i] = panelRoundsRowDTO{
			Project: r.Project, Change: r.Change,
			StartedAt: r.StartedAt.UTC().Format(time.RFC3339Nano),
			Rounds:    r.Rounds, Findings: r.Findings,
			Critical: r.Critical, Important: r.Important, Minor: r.Minor,
		}
	}
	return out
}

// GuardRunStore is the store dependency POST /api/v1/guard-runs/{project}
// needs, defined at the consumer. It is wired through WithGuardRuns rather
// than New's positional stores, so a server built without it -- every test
// that never exercises the route -- carries no fake for it.
type GuardRunStore interface {
	RecordGuardRun(ctx context.Context, projectKey string, in records.GuardRun) (records.GuardRun, error)
}

var _ GuardRunStore = (*store.Store)(nil)

// WithGuardRuns registers POST /api/v1/guard-runs/{project} against s.
func WithGuardRuns(s GuardRunStore) Option {
	return func(o *serverOptions) { o.guardRuns = s }
}

type guardRunHandler struct {
	store  GuardRunStore
	logger *slog.Logger
}

// record serves POST /api/v1/guard-runs/{project}: one guard invocation.
// An unknown project answers 404 and is not logged as a failure -- it is the
// expected answer for a guard run inside a test fixture's repository.
func (h *guardRunHandler) record(w http.ResponseWriter, r *http.Request) {
	project := r.PathValue("project")

	var in records.GuardRun
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := h.store.RecordGuardRun(r.Context(), project, in)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, out)
	case errors.Is(err, store.ErrProjectNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrInvalidGuardRun):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		status, msg := mapStoreError(h.logger, "record guard run for "+project, err)
		writeError(w, status, msg)
	}
}
