// stats.go serves the statistics surface: GET /api/v1/stats/{view} over
// task 3's eight aggregation methods, and GET /api/v1/stage-runs, a listing
// endpoint parallel to GET /api/v1/changes that gives task 3.1's stage-run
// query surface its first external caller (QueryStageRuns had none before
// this task -- only this package's own findOpenStageRun, internal, used
// it). Every dynamic condition in either handler is built by internal/store
// through store.Query's allowlist; this file never assembles SQL, per the
// package's own header and design.md's query-allowlist decision.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// boundaryConvention is the one sentence design.md's "Filtering, searching
// and sorting" and this task's own non-negotiable requirement both demand:
// stated once, applied identically by every view, and carried on every
// response so a reader never has to guess which instant a straddling stage
// run was counted against. A stage run is attributed to the period
// containing its *start* -- the harvester's own [start, end) windows use
// the same convention (internal/harvest/attribute.go) -- and every period
// this file accepts is itself half-open, [from, to), for the same reason:
// adjacent periods then partition every stage run exactly once, with no
// instant double-counted and none silently dropped between them.
const boundaryConvention = "a stage run is attributed to the period containing its start instant (started_at), never its end; periods are half-open [from, to)"

// StatsStore is the store dependency the statistics endpoints need, defined
// here at the consumer per go-interface-design: exactly task 3's eight
// aggregation methods plus QueryStageRuns, which this file also uses for
// the stage-runs listing endpoint and for resolving the earliest recorded
// stage run (see recorded, below). A fake satisfying this needs no
// database.
type StatsStore interface {
	LiveStateBoard(ctx context.Context, period store.Period, project *string) ([]store.LiveStateRow, error)
	CostPerChange(ctx context.Context, period store.Period, project, model *string) ([]store.CostPerChangeRow, error)
	StageLeaderboard(ctx context.Context, period store.Period, project, model *string) ([]store.StageLeaderboardRow, error)
	TrendOverTime(ctx context.Context, period store.Period, project, model *string) ([]store.TrendPoint, error)
	CacheEfficiency(ctx context.Context, period store.Period, project, model *string) ([]store.CacheEfficiencyRow, error)
	PanelEconomics(ctx context.Context, period store.Period, project, model *string) ([]store.PanelEconomicsRow, error)
	ModelComparison(ctx context.Context, period store.Period, project, model *string) ([]store.ModelComparisonRow, error)
	ReworkRate(ctx context.Context, period store.Period, project, model *string) ([]store.ReworkRateRow, error)
	QueryStageRuns(ctx context.Context, q store.Query) ([]store.StageRun, int, error)
	// CountRunsWithoutModel and ListModels back task 21's model filter:
	// the former only called when a model filter is set (statsResponse's
	// ExcludedNoModel), the latter GET /api/v1/models's only source.
	CountRunsWithoutModel(ctx context.Context, period store.Period, project *string) (int, error)
	ListModels(ctx context.Context, period store.Period, project *string) ([]string, error)
}

// var _ StatsStore = (*store.Store)(nil) verifies at compile time that the
// real store satisfies the interface this package actually depends on.
var _ StatsStore = (*store.Store)(nil)

// statsHandler serves the statistics endpoints.
type statsHandler struct {
	store  StatsStore
	logger *slog.Logger
}

// viewName is one of the eight statistics views' URL slugs, taken from
// design.md's "The views" table.
type viewName string

const (
	viewStateBoard       viewName = "state-board"
	viewCostPerChange    viewName = "cost-per-change"
	viewStageLeaderboard viewName = "stage-leaderboard"
	viewTrend            viewName = "trend"
	viewCacheEfficiency  viewName = "cache-efficiency"
	viewPanelEconomics   viewName = "panel-economics"
	viewModelComparison  viewName = "model-comparison"
	viewReworkRate       viewName = "rework-rate"
	breakdownRepo                 = "repo"
)

// knownViews is every accepted {view} path value, used both to dispatch and
// to name the accepted alternatives in an unknown-view error -- the same
// "name the field and the accepted alternatives" posture query.go's
// allowlist already takes for filter and sort fields.
var knownViews = []viewName{
	viewStateBoard, viewCostPerChange, viewStageLeaderboard, viewTrend,
	viewCacheEfficiency, viewPanelEconomics, viewModelComparison, viewReworkRate,
}

func acceptedViewNames() string {
	names := make([]string, len(knownViews))
	for i, v := range knownViews {
		names[i] = string(v)
	}
	return strings.Join(names, ", ")
}

// statsQueryParams are every query parameter the view endpoint interprets
// itself. Anything else present is rejected with 400 naming it -- never
// silently ignored -- per this task's own non-negotiable requirement,
// extended here from "an unknown query *field*" (store.Query's filter/sort
// surface) to "an unknown query *parameter*" on a route that has no
// filter/sort surface of its own to delegate that check to.
var statsQueryParams = map[string]bool{
	"from": true, "to": true, "project": true, "breakdown": true, "change": true,
	"command": true, "stage": true, "model": true,
}

// acceptedStatsQueryParamNames is statsQueryParams' keys, named in the
// unrecognised-parameter error message -- kept as one literal alongside the
// map it names, rather than duplicated at each call site, so adding a
// parameter to the map cannot silently leave the error message stale.
const acceptedStatsQueryParamNames = "from, to, project, breakdown, change, command, stage, model"

// statsResponse is the JSON envelope every view answers with: the request
// that was actually served (never merely echoed -- From/To are the parsed
// and reformatted instants, so a caller can see exactly what was applied),
// the boundary convention stated once per this task's own requirement, and
// Recorded, which is false when the requested period lies entirely before
// this store's earliest recorded stage run (or, if no stage run has ever
// been recorded in the requested scope, before anything at all) -- see
// (*statsHandler).recorded's doc comment for exactly what this does and
// does not claim.
// statsResponse's ExcludedNoModel is an *int, not an int with omitempty: an
// int with omitempty would drop a genuinely-zero exclusion count from the
// wire indistinguishably from "no filter was applied", which is exactly
// the absence-is-never-zero confusion task 21 exists to avoid. A *int with
// omitempty omits only a nil pointer -- absent means "no filter" -- and
// still encodes a pointer to zero as an explicit 0, per this file's own
// header comment on why no metric field here uses plain omitempty.
type statsResponse struct {
	View               viewName `json:"view"`
	From               string   `json:"from"`
	To                 string   `json:"to"`
	Project            *string  `json:"project,omitempty"`
	Model              *string  `json:"model,omitempty"`
	BoundaryConvention string   `json:"boundaryConvention"`
	Recorded           bool     `json:"recorded"`
	ExcludedNoModel    *int     `json:"excludedNoModel,omitempty"`
	Rows               any      `json:"rows"`
}

// parsePeriodAndProject parses the "from", "to" and "project" parameters
// every view shares. "from" and "to" are both required -- a view with no
// period is not a smaller request, it is a different, unbounded one this
// endpoint deliberately does not offer (design.md: "every statistics view
// is period-parameterised... rather than filtered in the client").
func parsePeriodAndProject(values map[string][]string) (store.Period, *string, error) {
	fromRaw := firstValue(values, "from")
	toRaw := firstValue(values, "to")
	if fromRaw == "" || toRaw == "" {
		return store.Period{}, nil, fmt.Errorf(`"from" and "to" are both required (RFC 3339 instants)`)
	}
	from, err := time.Parse(time.RFC3339, fromRaw)
	if err != nil {
		return store.Period{}, nil, fmt.Errorf("from: %q is not RFC 3339: %w", fromRaw, err)
	}
	to, err := time.Parse(time.RFC3339, toRaw)
	if err != nil {
		return store.Period{}, nil, fmt.Errorf("to: %q is not RFC 3339: %w", toRaw, err)
	}
	if to.Before(from) {
		return store.Period{}, nil, fmt.Errorf("to (%s) is before from (%s)", toRaw, fromRaw)
	}

	var project *string
	if p := firstValue(values, "project"); p != "" {
		project = &p
	}
	return store.Period{From: from, To: to}, project, nil
}

// parseLimit parses a "limit" query value shared by both list-shaped
// endpoints (this file's parseStageRunQuery and changes.go's
// parseChangeQuery), rejecting a negative number -- store.NoLimit (-1) is
// an internal sentinel that disables paging entirely, documented as such
// on the Query type itself, and was never meant to be reachable by a
// caller-supplied HTTP value. Before this existed, a caller sending
// literal "-1" resolved to that exact sentinel and got every matching row
// back with no LIMIT clause at all -- a real, reproduced gap (post-commit
// review finding F3) in both list endpoints, not merely this task's new
// one: the zero value ("limit unset") was already guarded by
// Query.limit()'s own default, but an *explicit* negative value sent over
// the wire never was.
func parseLimit(v string) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid limit %q: must be an integer", v)
	}
	if n < 0 {
		return 0, fmt.Errorf("invalid limit %q: must not be negative", v)
	}
	return n, nil
}

func firstValue(values map[string][]string, key string) string {
	if v := values[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}

// view serves GET /api/v1/stats/{view}.
func (h *statsHandler) view(w http.ResponseWriter, r *http.Request) {
	name := viewName(r.PathValue("view"))

	values := r.URL.Query()
	for param := range values {
		if !statsQueryParams[param] {
			writeError(w, http.StatusBadRequest, fmt.Sprintf(
				"unrecognised query parameter %q; accepted: %s", param, acceptedStatsQueryParamNames))
			return
		}
	}

	period, project, err := parsePeriodAndProject(values)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var model *string
	if m := firstValue(values, "model"); m != "" {
		model = &m
	}
	// The live state board's rows are changes, not stage runs: it has no
	// model to restrict by. Rejecting outright, before this handler does
	// any other work, is the one outcome the round's own spec allows --
	// silently accepting and ignoring the parameter would make the
	// control inert without saying so (specs/myflow-stats-views/spec.md,
	// "A model restriction on the live state board").
	if model != nil && name == viewStateBoard {
		writeError(w, http.StatusBadRequest, fmt.Sprintf(
			"view %q does not accept a model restriction: its rows are changes, not stage runs", name))
		return
	}

	// "breakdown" sat in statsQueryParams so the unknown-parameter guard
	// above passed it through on every view, and only rowsFor's
	// viewCostPerChange case ever read it -- every other view silently
	// discarded it and answered 200 with an ordinary, unbroken-down
	// response (task 25, step 1). Rejecting it here, before any store
	// call, mirrors the model rejection immediately above: name the view,
	// and never let the parameter reach a view that has no way to honour
	// it. cost-per-change's own breakdown=repo pairing rules (change,
	// project, command, stage) are unchanged -- they still run inside
	// rowsFor, downstream of this check.
	if firstValue(values, "breakdown") != "" && name != viewCostPerChange {
		writeError(w, http.StatusBadRequest, fmt.Sprintf(
			"view %q does not accept a breakdown restriction", name))
		return
	}

	recorded, err := h.recorded(r.Context(), period, project)
	if err != nil {
		status, msg := mapStoreError(h.logger, "resolve recorded period for "+string(name), err)
		writeError(w, status, msg)
		return
	}

	resp := statsResponse{
		View:               name,
		From:               period.From.UTC().Format(time.RFC3339Nano),
		To:                 period.To.UTC().Format(time.RFC3339Nano),
		Project:            project,
		Model:              model,
		BoundaryConvention: boundaryConvention,
		Recorded:           recorded,
	}

	// ExcludedNoModel is computed -- and only present on the wire -- when a
	// model filter was actually applied (statsResponse's own doc comment):
	// an unfiltered request pays nothing for this extra query, and a
	// filtered one always reports the count, including the genuinely-zero
	// case, rather than leaving a reader to wonder whether zero was
	// computed or simply omitted.
	if model != nil {
		excluded, err := h.store.CountRunsWithoutModel(r.Context(), period, project)
		if err != nil {
			status, msg := mapStoreError(h.logger, "count runs without model for "+string(name), err)
			writeError(w, status, msg)
			return
		}
		resp.ExcludedNoModel = &excluded
	}

	rows, status, msg := h.rowsFor(r.Context(), name, period, project, model, values)
	if status != 0 {
		writeError(w, status, msg)
		return
	}
	resp.Rows = rows
	writeJSON(w, http.StatusOK, resp)
}

// rowsFor dispatches to the one view name names, returning either the
// view's rows (status == 0) or the status and message to report instead.
// breakdown="repo" is honoured only for cost-per-change -- design.md's
// per-repository scenario is stated in terms of "a cost view", and every
// other view already aggregates across many changes and has no single
// change's repository set to break down in the first place.
//
// model is already known not to be set when name is viewStateBoard --
// view rejects that combination before rowsFor is ever called -- so every
// case below is free to pass it straight through.
func (h *statsHandler) rowsFor(ctx context.Context, name viewName, period store.Period, project, model *string, values map[string][]string) (rows any, status int, msg string) {
	switch name {
	case viewStateBoard:
		board, err := h.store.LiveStateBoard(ctx, period, project)
		if err != nil {
			s, m := mapStoreError(h.logger, "state board", err)
			return nil, s, m
		}
		return toStateBoardDTOs(board), 0, ""

	case viewCostPerChange:
		if breakdown := firstValue(values, "breakdown"); breakdown != "" {
			if breakdown != breakdownRepo {
				return nil, http.StatusBadRequest, fmt.Sprintf(`unrecognised breakdown %q; accepted: %q`, breakdown, breakdownRepo)
			}
			changeName := firstValue(values, "change")
			if changeName == "" {
				return nil, http.StatusBadRequest, `breakdown=repo requires a "change" parameter naming the change to break down`
			}
			// A change is keyed by (project, name) -- store.PutChange's own
			// contract, exercised directly by TestSameNameInTwoProjectsCoexist
			// -- so "change" alone is not enough to identify one: two
			// projects can and do use the same change name. Every SQL-based
			// view avoids this by grouping on c.project_key too; this
			// breakdown does its own grouping in Go (costPerChangeByRepo's
			// own doc comment explains why) and must apply the same
			// discipline explicitly, so project is required here rather than
			// silently blending every same-named change across every
			// project into one breakdown -- a wrong, plausible-looking
			// answer, not merely an incomplete one.
			if project == nil {
				return nil, http.StatusBadRequest, `breakdown=repo requires a "project" parameter: a change name alone does not identify one change across projects`
			}
			// command and stage are required too, and for the identical
			// reason project is: viewCostPerChange's own non-breakdown rows
			// are grouped by (project, name, command, stage) -- one row per
			// stage, per design.md's "broken down by command and stage" --
			// and the interface nests exactly one breakdown toggle under
			// each of those rows (CostPerChange.tsx's RepoBreakdown). A
			// breakdown scoped to the change as a whole, filtered only by
			// project and name, would silently blend every stage's runs
			// into one panel no matter which row's toggle opened it -- a
			// panel that reconciles with none of them (post-commit review
			// finding F1: reproduced live, two rows of one change opening
			// to the identical, unreconciling total). There is no other
			// caller of breakdown=repo that would want the change-wide
			// figure instead, so command and stage are required rather than
			// silently optional -- a request that omits either is rejected,
			// never quietly widened to "the whole change".
			command := firstValue(values, "command")
			stage := firstValue(values, "stage")
			if command == "" || stage == "" {
				return nil, http.StatusBadRequest, `breakdown=repo requires "command" and "stage" parameters naming the row to break down`
			}
			rows, err := h.costPerChangeByRepo(ctx, period, *project, changeName, command, stage)
			if err != nil {
				s, m := mapStoreError(h.logger, "cost per change by repo", err)
				return nil, s, m
			}
			return rows, 0, ""
		}
		cost, err := h.store.CostPerChange(ctx, period, project, model)
		if err != nil {
			s, m := mapStoreError(h.logger, "cost per change", err)
			return nil, s, m
		}
		// "change" alone (no breakdown) scopes cost-per-change to one
		// change, server-side -- task 21, step 5. Before this, only
		// breakdown=repo ever read "change", so a caller wanting one
		// change's own totals (the run-detail header, useRunDetail.ts) had
		// to fetch every change in the project and filter client-side,
		// which stayed correct only because this view is unpaged; the
		// moment a limit is added here that shape silently under-reports.
		// This is a plain Go slice filter, not a new store method, for the
		// same reason costPerChangeByRepo's own doc comment gives: the
		// result set CostPerChange already returned is small and bounded
		// by construction, so re-filtering it here is the smaller surface.
		if changeName := firstValue(values, "change"); changeName != "" {
			cost = filterCostPerChangeByName(cost, changeName)
		}
		return toCostPerChangeDTOs(cost), 0, ""

	case viewStageLeaderboard:
		board, err := h.store.StageLeaderboard(ctx, period, project, model)
		if err != nil {
			s, m := mapStoreError(h.logger, "stage leaderboard", err)
			return nil, s, m
		}
		return toStageLeaderboardDTOs(board), 0, ""

	case viewTrend:
		trend, err := h.store.TrendOverTime(ctx, period, project, model)
		if err != nil {
			s, m := mapStoreError(h.logger, "trend over time", err)
			return nil, s, m
		}
		return toTrendDTOs(trend), 0, ""

	case viewCacheEfficiency:
		eff, err := h.store.CacheEfficiency(ctx, period, project, model)
		if err != nil {
			s, m := mapStoreError(h.logger, "cache efficiency", err)
			return nil, s, m
		}
		return toCacheEfficiencyDTOs(eff), 0, ""

	case viewPanelEconomics:
		panel, err := h.store.PanelEconomics(ctx, period, project, model)
		if err != nil {
			s, m := mapStoreError(h.logger, "panel economics", err)
			return nil, s, m
		}
		return toPanelEconomicsDTOs(panel), 0, ""

	case viewModelComparison:
		cmp, err := h.store.ModelComparison(ctx, period, project, model)
		if err != nil {
			s, m := mapStoreError(h.logger, "model comparison", err)
			return nil, s, m
		}
		return toModelComparisonDTOs(cmp), 0, ""

	case viewReworkRate:
		rework, err := h.store.ReworkRate(ctx, period, project, model)
		if err != nil {
			s, m := mapStoreError(h.logger, "rework rate", err)
			return nil, s, m
		}
		return toReworkRateDTOs(rework), 0, ""

	default:
		return nil, http.StatusBadRequest, fmt.Sprintf("unrecognised view %q; accepted: %s", name, acceptedViewNames())
	}
}

// recorded reports whether period could contain any recorded stage run --
// that is, whether telemetry existed for at least part of it -- scoped to
// project when given. It answers false only when period.To is at or before
// the earliest started_at this store (within that scope) has ever recorded,
// including the case where nothing has ever been recorded there at all.
//
// This is deliberately not "did this specific view return any rows for
// this period": a genuinely quiet period *after* telemetry began (nobody
// ran a stage that week) is a real, measured zero, and every view already
// reports that correctly as an empty Rows slice. What Recorded distinguishes
// is the different fact design.md's "Starting empty" and this task's
// absence-is-not-a-value requirement both name: a period *before* the store
// held anything is not a zero at all, and must not read as one.
//
// The earliest instant is resolved through QueryStageRuns -- sorted by
// started_at ascending, limited to one row -- rather than a bespoke SQL
// method, per the query-allowlist design: this file builds no SQL of its
// own anywhere.
func (h *statsHandler) recorded(ctx context.Context, period store.Period, project *string) (bool, error) {
	var filters []store.Filter
	if project != nil {
		filters = append(filters, store.Filter{Field: "project", Op: store.OpEq, Value: *project})
	}
	runs, _, err := h.store.QueryStageRuns(ctx, store.Query{
		Filters: filters,
		Sort:    []store.SortKey{{Field: "started_at"}},
		Limit:   1,
	})
	if err != nil {
		return false, err
	}
	if len(runs) == 0 {
		return false, nil
	}
	return period.To.After(runs[0].StartedAt), nil
}

// --- view row DTOs ---
//
// Every DTO below mirrors its store row one field at a time, with no
// omitempty on a metric field: a nil pointer is encoded as an explicit JSON
// null rather than an absent key, so a reader can tell "the server said
// this metric was not measured" apart from "the server forgot to mention
// this field", which omitempty could not do -- the same distinction
// design.md's metrics bag draws between absence and a recorded zero,
// carried through to the wire.

type stateBoardRowDTO struct {
	ProjectKey  string `json:"projectKey"`
	Name        string `json:"name"`
	State       string `json:"state"`
	UpdatedAt   string `json:"updatedAt"`
	UpdatedBy   string `json:"updatedBy"`
	NextCommand string `json:"nextCommand"`
}

// nextCommandFor maps a change's current pipeline state to the command
// that runs next, per CLAUDE.md's fixed three-state pipeline
// (/myflow-start -> STARTED, /myflow-do -> IN_PROGRESS, /myflow-finish ->
// FINISHED) -- the same table /myflow-status reads to answer the same
// question, so the two report identically for the same underlying record
// (spec: "the information matches what /myflow-status reports").
func nextCommandFor(s store.State) string {
	switch s {
	case store.StateStarted:
		return "/myflow-do"
	case store.StateInProgress:
		return "/myflow-finish"
	default:
		// FINISHED is terminal: no command comes next.
		return ""
	}
}

func toStateBoardDTOs(rows []store.LiveStateRow) []stateBoardRowDTO {
	out := make([]stateBoardRowDTO, len(rows))
	for i, r := range rows {
		out[i] = stateBoardRowDTO{
			ProjectKey:  r.ProjectKey,
			Name:        r.Name,
			State:       string(r.State),
			UpdatedAt:   r.UpdatedAt.UTC().Format(time.RFC3339Nano),
			UpdatedBy:   r.UpdatedBy,
			NextCommand: nextCommandFor(r.State),
		}
	}
	return out
}

type costPerChangeRowDTO struct {
	ProjectKey string `json:"projectKey"`
	ChangeName string `json:"changeName"`
	Command    string `json:"command"`
	Stage      string `json:"stage"`

	RunCount     int `json:"runCount"`
	MeasuredRuns int `json:"measuredRuns"`

	TotalTokensInput *int64   `json:"totalTokensInput"`
	MeanTokensInput  *float64 `json:"meanTokensInput"`
	TotalCostUSD     *float64 `json:"totalCostUsd"`
	TotalDurationMs  *int64   `json:"totalDurationMs"`

	MainTokens      *int64 `json:"mainTokens"`
	SidechainTokens *int64 `json:"sidechainTokens"`
}

func toCostPerChangeDTOs(rows []store.CostPerChangeRow) []costPerChangeRowDTO {
	out := make([]costPerChangeRowDTO, len(rows))
	for i, r := range rows {
		out[i] = costPerChangeRowDTO{
			ProjectKey:       r.ProjectKey,
			ChangeName:       r.ChangeName,
			Command:          r.Command,
			Stage:            r.Stage,
			RunCount:         r.RunCount,
			MeasuredRuns:     r.MeasuredRuns,
			TotalTokensInput: r.TotalTokensInput,
			MeanTokensInput:  r.MeanTokensInput,
			TotalCostUSD:     r.TotalCostUSD,
			TotalDurationMs:  r.TotalDurationMs,
			MainTokens:       r.MainTokens,
			SidechainTokens:  r.SidechainTokens,
		}
	}
	return out
}

type stageLeaderboardRowDTO struct {
	Command  string `json:"command"`
	Stage    string `json:"stage"`
	RunCount int    `json:"runCount"`

	MeanCostUSD   float64 `json:"meanCostUsd"`
	MedianCostUSD float64 `json:"medianCostUsd"`
	P90CostUSD    float64 `json:"p90CostUsd"`
}

func toStageLeaderboardDTOs(rows []store.StageLeaderboardRow) []stageLeaderboardRowDTO {
	out := make([]stageLeaderboardRowDTO, len(rows))
	for i, r := range rows {
		out[i] = stageLeaderboardRowDTO{
			Command: r.Command, Stage: r.Stage, RunCount: r.RunCount,
			MeanCostUSD: r.MeanCostUSD, MedianCostUSD: r.MedianCostUSD, P90CostUSD: r.P90CostUSD,
		}
	}
	return out
}

type trendPointDTO struct {
	Day          string   `json:"day"`
	RunCount     int      `json:"runCount"`
	TotalCostUSD *float64 `json:"totalCostUsd"`
}

func toTrendDTOs(points []store.TrendPoint) []trendPointDTO {
	out := make([]trendPointDTO, len(points))
	for i, p := range points {
		out[i] = trendPointDTO{
			Day:          p.Day.UTC().Format("2006-01-02"),
			RunCount:     p.RunCount,
			TotalCostUSD: p.TotalCostUSD,
		}
	}
	return out
}

type cacheEfficiencyRowDTO struct {
	Command string `json:"command"`
	Stage   string `json:"stage"`

	CacheReadTotal     *int64   `json:"cacheReadTotal"`
	CacheCreationTotal *int64   `json:"cacheCreationTotal"`
	Ratio              *float64 `json:"ratio"`
}

func toCacheEfficiencyDTOs(rows []store.CacheEfficiencyRow) []cacheEfficiencyRowDTO {
	out := make([]cacheEfficiencyRowDTO, len(rows))
	for i, r := range rows {
		out[i] = cacheEfficiencyRowDTO{
			Command: r.Command, Stage: r.Stage,
			CacheReadTotal: r.CacheReadTotal, CacheCreationTotal: r.CacheCreationTotal, Ratio: r.Ratio,
		}
	}
	return out
}

type panelEconomicsRowDTO struct {
	ReviewPanelRoster string `json:"reviewPanelRoster"`

	FindingsTotal   int64    `json:"findingsTotal"`
	TokensTotal     *int64   `json:"tokensTotal"`
	FindingsPerMTok *float64 `json:"findingsPerMtok"`
}

func toPanelEconomicsDTOs(rows []store.PanelEconomicsRow) []panelEconomicsRowDTO {
	out := make([]panelEconomicsRowDTO, len(rows))
	for i, r := range rows {
		out[i] = panelEconomicsRowDTO{
			ReviewPanelRoster: r.ReviewPanelRoster,
			FindingsTotal:     r.FindingsTotal, TokensTotal: r.TokensTotal, FindingsPerMTok: r.FindingsPerMTok,
		}
	}
	return out
}

type modelComparisonRowDTO struct {
	Model   string `json:"model"`
	Command string `json:"command"`
	Stage   string `json:"stage"`

	RunCount       int      `json:"runCount"`
	MeanCostUSD    *float64 `json:"meanCostUsd"`
	ReworkAttempts int      `json:"reworkAttempts"`
}

func toModelComparisonDTOs(rows []store.ModelComparisonRow) []modelComparisonRowDTO {
	out := make([]modelComparisonRowDTO, len(rows))
	for i, r := range rows {
		out[i] = modelComparisonRowDTO{
			Model: r.Model, Command: r.Command, Stage: r.Stage,
			RunCount: r.RunCount, MeanCostUSD: r.MeanCostUSD, ReworkAttempts: r.ReworkAttempts,
		}
	}
	return out
}

type reworkRateRowDTO struct {
	Command string `json:"command"`
	Stage   string `json:"stage"`

	TotalAttempts  int `json:"totalAttempts"`
	ReworkAttempts int `json:"reworkAttempts"`
	AbandonedCount int `json:"abandonedCount"`
}

func toReworkRateDTOs(rows []store.ReworkRateRow) []reworkRateRowDTO {
	out := make([]reworkRateRowDTO, len(rows))
	for i, r := range rows {
		out[i] = reworkRateRowDTO{
			Command: r.Command, Stage: r.Stage,
			TotalAttempts: r.TotalAttempts, ReworkAttempts: r.ReworkAttempts, AbandonedCount: r.AbandonedCount,
		}
	}
	return out
}

// filterCostPerChangeByName narrows rows to changeName, preserving order.
// Task 21, step 5's own doc comment (rowsFor's viewCostPerChange case)
// explains why this is a Go filter over CostPerChange's already-fetched,
// already-small result rather than a new store-level parameter: unlike
// project and model, "change" alone does not identify one change across
// every project (store.PutChange's own contract -- two projects can and do
// reuse the same change name), so a caller wanting one change's own row
// server-side is expected to send "project" alongside it, exactly as
// useRunDetail.ts does; this filter does not itself enforce that pairing,
// since a caller filtering within one project's own already-scoped result
// (or accepting the small residual risk of a same-named change elsewhere)
// is not the shape this step exists to forbid -- only the unbounded,
// project-blind blend breakdown=repo's own comment describes there.
func filterCostPerChangeByName(rows []store.CostPerChangeRow, changeName string) []store.CostPerChangeRow {
	out := make([]store.CostPerChangeRow, 0, len(rows))
	for _, r := range rows {
		if r.ChangeName == changeName {
			out = append(out, r)
		}
	}
	return out
}

// --- cost-per-change repository breakdown ---

// costPerChangeRepoRowDTO is one repository's contribution to a single
// change's cost -- design.md's "per-repository breakdown available on
// request", requested via breakdown=repo&change=<name>. RepoRoot is nil
// for stage runs recorded against the change as a whole (store.StageRun's
// own nil-repo_root convention: "the whole unit of work", never "unknown
// repository") -- summed here under its own row exactly like any other
// repository, rather than folded into one of the named ones or dropped.
type costPerChangeRepoRowDTO struct {
	RepoRoot *string `json:"repoRoot"`

	RunCount     int `json:"runCount"`
	MeasuredRuns int `json:"measuredRuns"`

	TotalTokensInput *int64   `json:"totalTokensInput"`
	TotalCostUSD     *float64 `json:"totalCostUsd"`
	TotalDurationMs  *int64   `json:"totalDurationMs"`
}

// repoBreakdownAccumulator sums one repository's contribution across the
// stage runs costPerChangeByRepo groups under it. The have* flags are what
// let a repository whose runs never once carried a given metric report that
// field as nil rather than a fabricated zero -- the same
// absence-is-not-a-value rule store.CostPerChange's own SQL already applies
// via COUNT/SUM's native NULL-on-no-rows behaviour, reproduced here in Go
// because this breakdown reads raw stage runs rather than issuing its own
// GROUP BY (see costPerChangeByRepo's doc comment for why).
type repoBreakdownAccumulator struct {
	runCount     int
	measuredRuns int

	haveTokensInput bool
	totalTokensIn   float64

	haveCost  bool
	totalCost float64

	haveDuration    bool
	totalDurationMs int64
}

func (a *repoBreakdownAccumulator) dto(repoRoot *string) costPerChangeRepoRowDTO {
	dto := costPerChangeRepoRowDTO{RepoRoot: repoRoot, RunCount: a.runCount, MeasuredRuns: a.measuredRuns}
	if a.haveTokensInput {
		v := int64(a.totalTokensIn)
		dto.TotalTokensInput = &v
	}
	if a.haveCost {
		v := a.totalCost
		dto.TotalCostUSD = &v
	}
	if a.haveDuration {
		v := a.totalDurationMs
		dto.TotalDurationMs = &v
	}
	return dto
}

// stageRunMetricsBag is the subset of a stage run's metrics bag
// costPerChangeByRepo reads, decoded key-by-key (rather than into a single
// fixed struct) so that a key's mere *presence* -- "tokens" as a whole, for
// MeasuredRuns -- can be told apart from a key that is present but carries
// no numeric value, exactly as store.CostPerChangeRow's own doc comment
// distinguishes RunCount from MeasuredRuns.
type stageRunTokensSubset struct {
	Input *float64 `json:"input"`
}

// costPerChangeByRepo answers design.md's per-repository breakdown scenario
// for one row of the cost-per-change view -- one (project, change, command,
// stage) group, matching exactly what that view's own non-breakdown rows
// are grouped by (store.CostPerChange's GROUP BY). It fetches that group's
// stage runs starting within period via QueryStageRuns -- store's own
// allowlisted query surface, never a bespoke SQL method -- and sums each
// metric per distinct RepoRoot in Go.
//
// This is deliberately not a ninth store aggregation method. Every other
// view aggregates over however many changes and stage runs a whole period
// contains, which is exactly why they are SQL GROUP BYs: the data volume
// makes summing in application code the wrong trade-off. This breakdown is
// scoped to one caller-named row's own stage runs, a small, bounded set by
// construction, so re-deriving store.CostPerChangeRow's arithmetic in Go
// over an already-small result is the smaller surface, not a shortcut
// around it -- and it reuses the one query surface store.go already
// exposes externally (QueryStageRuns) rather than adding a new one.
//
// project is required, not optional: a change is keyed by (project, name),
// and store.TestSameNameInTwoProjectsCoexist is exactly the scenario a
// name-only filter here would silently blend together (post-commit review
// finding F1) -- every SQL-based aggregation avoids this because its own
// GROUP BY includes c.project_key, and this function must apply the same
// discipline explicitly since it builds no SQL of its own.
//
// command and stage are required for the same reason, added by a later
// review round (post-commit review finding F1, task 13): without them this
// summed every stage run of the whole change into one panel, so two
// differently-costed rows of the same change ("SDD + TDD per task" and
// "5. The review panel", say) opened to the identical total, and that total
// reconciled with neither row's own figures -- reproduced live against a
// seeded database, not caught by any fixture, because a fixture with only
// one stage per change cannot exhibit it. rowsFor refuses the request
// before this is ever reached if project, command or stage is missing.
func (h *statsHandler) costPerChangeByRepo(ctx context.Context, period store.Period, project, changeName, command, stage string) ([]costPerChangeRepoRowDTO, error) {
	filters := []store.Filter{
		{Field: "project", Op: store.OpEq, Value: project},
		{Field: "name", Op: store.OpEq, Value: changeName},
		{Field: "command", Op: store.OpEq, Value: command},
		{Field: "stage", Op: store.OpEq, Value: stage},
		{Field: "started_at", Op: store.OpGte, Value: period.From},
		{Field: "started_at", Op: store.OpLt, Value: period.To},
	}

	runs, _, err := h.store.QueryStageRuns(ctx, store.Query{Filters: filters, Limit: store.NoLimit})
	if err != nil {
		return nil, err
	}

	byRepo := map[string]*repoBreakdownAccumulator{}
	var order []string
	keyFor := func(repoRoot *string) string {
		if repoRoot == nil {
			return ""
		}
		return *repoRoot
	}

	for _, run := range runs {
		key := keyFor(run.RepoRoot)
		acc, ok := byRepo[key]
		if !ok {
			acc = &repoBreakdownAccumulator{}
			byRepo[key] = acc
			order = append(order, key)
		}
		acc.runCount++

		var bag map[string]json.RawMessage
		if len(run.Metrics) > 0 {
			if err := json.Unmarshal(run.Metrics, &bag); err != nil {
				return nil, fmt.Errorf("api: decode metrics for stage run %d: %w", run.ID, err)
			}
		}
		if tokensRaw, ok := bag["tokens"]; ok {
			acc.measuredRuns++
			var tok stageRunTokensSubset
			if err := json.Unmarshal(tokensRaw, &tok); err != nil {
				return nil, fmt.Errorf("api: decode tokens for stage run %d: %w", run.ID, err)
			}
			if tok.Input != nil {
				acc.haveTokensInput = true
				acc.totalTokensIn += *tok.Input
			}
		}
		if costRaw, ok := bag["cost_usd"]; ok {
			var cost float64
			if err := json.Unmarshal(costRaw, &cost); err != nil {
				return nil, fmt.Errorf("api: decode cost_usd for stage run %d: %w", run.ID, err)
			}
			acc.haveCost = true
			acc.totalCost += cost
		}
		if run.EndedAt != nil {
			acc.haveDuration = true
			acc.totalDurationMs += run.EndedAt.Sub(run.StartedAt).Milliseconds()
		}
	}

	sort.Strings(order)
	out := make([]costPerChangeRepoRowDTO, 0, len(order))
	for _, key := range order {
		var repoRoot *string
		if key != "" {
			k := key
			repoRoot = &k
		}
		out = append(out, byRepo[key].dto(repoRoot))
	}
	return out, nil
}

// --- GET /api/v1/stage-runs ---

// stageRunDTO is the wire representation of a store.StageRun, matching
// changeDTO's own field-for-field style (internal/api/changes.go).
type stageRunDTO struct {
	StageRunID int64           `json:"stageRunId"`
	RepoRoot   *string         `json:"repoRoot,omitempty"`
	Harness    string          `json:"harness"`
	SessionID  *string         `json:"sessionId,omitempty"`
	Command    string          `json:"command"`
	Stage      string          `json:"stage"`
	Attempt    int             `json:"attempt"`
	StartedAt  string          `json:"startedAt"`
	EndedAt    *string         `json:"endedAt,omitempty"`
	Outcome    *string         `json:"outcome,omitempty"`
	Metrics    json.RawMessage `json:"metrics,omitempty"`
}

func toStageRunDTO(r store.StageRun) stageRunDTO {
	dto := stageRunDTO{
		StageRunID: r.ID,
		RepoRoot:   r.RepoRoot,
		Harness:    r.Harness,
		SessionID:  r.SessionID,
		Command:    r.Command,
		Stage:      r.Stage,
		Attempt:    r.Attempt,
		StartedAt:  r.StartedAt.UTC().Format(time.RFC3339Nano),
		Outcome:    r.Outcome,
		Metrics:    r.Metrics,
	}
	if r.EndedAt != nil {
		endedAt := r.EndedAt.UTC().Format(time.RFC3339Nano)
		dto.EndedAt = &endedAt
	}
	return dto
}

type listStageRunsResponse struct {
	Total     int           `json:"total"`
	StageRuns []stageRunDTO `json:"stageRuns"`
}

// listStageRuns serves GET /api/v1/stage-runs?filter...&q=&sort=&limit=&offset=,
// the stage-run counterpart of (*changeHandler).list: task 3.1's filter,
// search, sort and page surface had store.QueryStageRuns as its only
// implementation but no HTTP caller before this task (see this file's own
// package doc comment). This handler is deliberately parallel to
// (*changeHandler).list/parseChangeQuery rather than sharing code with it:
// the two exist in separate files with separate review histories
// (changes.go predates this task), and query.go's own allowlist -- not this
// handler -- is what actually decides which field names either accepts, so
// the duplication here is a thin, mechanical request-to-store.Query
// translation, not a second copy of the logic that matters.
func (h *statsHandler) listStageRuns(w http.ResponseWriter, r *http.Request) {
	q, err := parseStageRunQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	runs, total, err := h.store.QueryStageRuns(r.Context(), q)
	if err != nil {
		status, msg := mapStoreError(h.logger, "list stage runs", err)
		writeError(w, status, msg)
		return
	}

	dtos := make([]stageRunDTO, len(runs))
	for i, run := range runs {
		dtos[i] = toStageRunDTO(run)
	}
	writeJSON(w, http.StatusOK, listStageRunsResponse{Total: total, StageRuns: dtos})
}

// parseStageRunQuery turns r's query string into a store.Query -- see
// parseChangeQuery (changes.go) for the identical shape this mirrors.
func parseStageRunQuery(r *http.Request) (store.Query, error) {
	values := r.URL.Query()
	q := store.Query{Search: values.Get("q")}

	if sortParam := values.Get("sort"); sortParam != "" {
		for _, field := range strings.Split(sortParam, ",") {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			desc := false
			if strings.HasPrefix(field, "-") {
				desc = true
				field = field[1:]
			}
			q.Sort = append(q.Sort, store.SortKey{Field: field, Desc: desc})
		}
	}

	if v := values.Get("limit"); v != "" {
		n, err := parseLimit(v)
		if err != nil {
			return store.Query{}, err
		}
		q.Limit = n
	}
	if v := values.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return store.Query{}, fmt.Errorf("invalid offset %q: must be an integer", v)
		}
		q.Offset = n
	}

	for field, vals := range values {
		if reservedQueryParams[field] || len(vals) == 0 || vals[0] == "" {
			continue
		}
		q.Filters = append(q.Filters, store.Filter{Field: field, Op: store.OpEq, Value: vals[0]})
	}
	sort.Slice(q.Filters, func(i, j int) bool { return q.Filters[i].Field < q.Filters[j].Field })

	return q, nil
}

// --- GET /api/v1/models ---

// modelsQueryParams are the only parameters GET /api/v1/models interprets
// -- "from", "to" and "project", exactly parsePeriodAndProject's own
// surface -- rejecting anything else with 400 naming it, the same posture
// statsQueryParams already takes on the view endpoint (this file's own
// header comment).
var modelsQueryParams = map[string]bool{"from": true, "to": true, "project": true}

// modelsResponse is GET /api/v1/models's envelope: the period actually
// applied (echoed back, like statsResponse's own From/To) and the distinct
// models recorded in it, sorted, task 21's own source for "the set of
// models offered" (specs/myflow-stats-views/spec.md) rather than a
// hard-coded list this build would need rebuilding to extend.
type modelsResponse struct {
	From    string   `json:"from"`
	To      string   `json:"to"`
	Project *string  `json:"project,omitempty"`
	Models  []string `json:"models"`
}

// listModels serves GET /api/v1/models?from=&to=&project=. It goes through
// withDaemonHeader like every other route on this mux (server.go wraps the
// whole mux, not each route individually), so a caller can still tell the
// daemon's own answer from a proxy's.
func (h *statsHandler) listModels(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	for param := range values {
		if !modelsQueryParams[param] {
			writeError(w, http.StatusBadRequest, fmt.Sprintf(
				"unrecognised query parameter %q; accepted: from, to, project", param))
			return
		}
	}

	period, project, err := parsePeriodAndProject(values)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	models, err := h.store.ListModels(r.Context(), period, project)
	if err != nil {
		status, msg := mapStoreError(h.logger, "list models", err)
		writeError(w, status, msg)
		return
	}
	if models == nil {
		models = []string{}
	}

	writeJSON(w, http.StatusOK, modelsResponse{
		From:    period.From.UTC().Format(time.RFC3339Nano),
		To:      period.To.UTC().Format(time.RFC3339Nano),
		Project: project,
		Models:  models,
	})
}
