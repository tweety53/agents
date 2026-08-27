// stats.go serves the statistics surface: GET /api/v1/stats/{view} over
// task 3's aggregation methods, and GET /api/v1/stage-runs, a listing
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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
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
// here at the consumer per go-interface-design: exactly task 3's
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
	QueryStageRuns(ctx context.Context, q store.Query) ([]store.StageRun, int, error)
	// CountRunsWithoutModel and ListModels back task 21's model filter:
	// the former only called when a model filter is set (statsResponse's
	// ExcludedNoModel), the latter GET /api/v1/models's only source.
	CountRunsWithoutModel(ctx context.Context, period store.Period, project *string) (int, error)
	ListModels(ctx context.Context, period store.Period, project *string) ([]string, error)
	// AllRecordedRunsUnmeasured backs statsResponse's Unmeasured, below:
	// the third arm between "not recorded" and "recorded as zero"
	// (design.md, "the third arm of the absence distinction"). Called only
	// when Recorded is already true -- a period predating any telemetry
	// cannot also be "all unmeasured", so calling it there would spend a
	// query answering a question that does not apply to that period.
	AllRecordedRunsUnmeasured(ctx context.Context, period store.Period, project *string) (bool, error)
	// projectResolver backs resolveProjectParam, below: every "project"
	// value this handler parses (view and listModels alike) goes through
	// it before reaching any of the methods above. Embedded rather than
	// spelling ProjectKeysByDisplayName's signature out a second time
	// (panel round 1, F2): the contract lives once, in projectResolver
	// itself, and this embedding is what states, in the type rather than
	// only in prose, that StatsStore already satisfies it.
	projectResolver
}

// var _ StatsStore = (*store.Store)(nil) verifies at compile time that the
// real store satisfies the interface this package actually depends on.
var _ StatsStore = (*store.Store)(nil)

// statsHandler serves the statistics endpoints.
type statsHandler struct {
	store  StatsStore
	logger *slog.Logger
}

// viewName is one of the statistics views' URL slugs, taken from
// design.md's "The views" table.
type viewName string

const (
	viewStateBoard       viewName = "state-board"
	viewCostPerChange    viewName = "cost-per-change"
	viewStageLeaderboard viewName = "stage-leaderboard"
	viewTrend            viewName = "trend"
	viewCacheEfficiency  viewName = "cache-efficiency"
)

// knownViews is every accepted {view} path value, used both to dispatch and
// to name the accepted alternatives in an unknown-view error -- the same
// "name the field and the accepted alternatives" posture query.go's
// allowlist already takes for filter and sort fields.
var knownViews = []viewName{
	viewStateBoard, viewCostPerChange, viewStageLeaderboard, viewTrend,
	viewCacheEfficiency,
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
	"from": true, "to": true, "project": true, "change": true, "model": true,
}

// acceptedStatsQueryParamNames is statsQueryParams' keys, named in the
// unrecognised-parameter error message -- kept as one literal alongside the
// map it names, rather than duplicated at each call site, so adding a
// parameter to the map cannot silently leave the error message stale.
const acceptedStatsQueryParamNames = "from, to, project, change, model"

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
//
// Unmeasured is task 5's third arm, added deliberately alongside Recorded
// rather than folded into it: Recorded carries one fact ("this period is
// not entirely before anything this store has ever held"), and Unmeasured
// carries a different one ("stage runs exist for this period, and not one
// of them carries a measurement"). Collapsing the two into one flag would
// make a recording misconfiguration -- runs exist, attribution never
// bound one to a session -- indistinguishable from a genuinely quiet
// period, which is exactly the defect this task exists to close. It is
// only ever true when Recorded is also true (see the handler): a period
// that predates all telemetry is reported through Recorded alone.
type statsResponse struct {
	View               viewName `json:"view"`
	From               string   `json:"from"`
	To                 string   `json:"to"`
	Project            *string  `json:"project,omitempty"`
	Model              *string  `json:"model,omitempty"`
	BoundaryConvention string   `json:"boundaryConvention"`
	Recorded           bool     `json:"recorded"`
	Unmeasured         bool     `json:"unmeasured"`
	ExcludedNoModel    *int     `json:"excludedNoModel,omitempty"`
	Rows               any      `json:"rows"`
}

// parsePeriodAndProject parses the "from", "to" and "project" parameters
// every view shares. "from" and "to" are both required -- a view with no
// period is not a smaller request, it is a different, unbounded one this
// endpoint deliberately does not offer (design.md: "every statistics view
// is period-parameterised... rather than filtered in the client"). The
// "project" value, if present, is resolved through resolveProjectParam --
// see its own doc comment for what a display name, an exact key, an
// ambiguous name and an unknown one each do.
func parsePeriodAndProject(ctx context.Context, resolver projectResolver, values map[string][]string) (store.Period, *string, error) {
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
		resolved, err := resolveProjectParam(ctx, resolver, p)
		if err != nil {
			return store.Period{}, nil, err
		}
		project = &resolved
	}
	return store.Period{From: from, To: to}, project, nil
}

// projectKeySuffixPattern matches the trailing "-" plus exactly eight
// lowercase hexadecimal characters that stats/web/src/lib/projectLabel.ts
// strips, client-side, to derive a project's display name. Built from
// store.ProjectKeySuffixPattern (panel round 1, F1) rather than its own
// literal, so this Go-side copy and internal/store/changes.go's SQL copy
// cannot drift against each other -- see that constant's own doc comment
// for why the TypeScript copy is the one duplicate that stays a duplicate.
// A "project" query value already carrying that shape cannot itself be a
// display name -- projectLabel.ts never leaves that suffix on -- so
// looksLikeProjectKey treats it as an exact key already, letting
// resolveProjectParam skip resolution (and its query) entirely for the
// common case: a value read straight off a link or a bookmark that still
// carries the full key.
var projectKeySuffixPattern = regexp.MustCompile(store.ProjectKeySuffixPattern)

func looksLikeProjectKey(value string) bool {
	return projectKeySuffixPattern.MatchString(value)
}

// projectResolver is the minimal store dependency resolveProjectParam
// needs, defined at the consumer per go-interface-design. Both ChangeStore
// (server.go) and StatsStore (above) embed projectResolver rather than
// spelling its one method out a second time (panel round 1, F2), so
// h.store already satisfies this interface at either call site below with
// no adapter needed.
type projectResolver interface {
	ProjectKeysByDisplayName(ctx context.Context, displayName string) ([]string, error)
}

// resolveProjectParam is the one place a "project" query parameter is
// turned into the project key a store query actually filters on. It is
// shared by parsePeriodAndProject (every stats view, plus GET
// /api/v1/models) and changes.go's parseChangeQuery (the changes list's
// own "project" filter) so specs/myflow-stats-views/spec.md's "The project
// parameter accepts a display name" is applied identically at both call
// sites rather than reimplemented at either.
//
//   - A value already carrying the derivation suffix (looksLikeProjectKey)
//     is used unchanged, with no query at all: a display name never
//     carries that suffix, so a value that does can only be a key --
//     exact or unknown -- and either way the caller's own store query
//     already handles it (an unknown key matches zero rows, exactly as
//     today).
//   - Otherwise value is resolved against project display names: one
//     match resolves to that project's key; several is refused with an
//     error naming the ambiguity and every candidate key (mapped to 400
//     by both callers); none passes value through unchanged, preserving
//     "an unknown project yields no rows" rather than turning a filter
//     that never errored before into one that suddenly can.
//
// projectResolverError wraps a failure resolveProjectParam propagates
// straight from projectResolver.ProjectKeysByDisplayName -- a store
// failure, not a caller mistake -- so it can be told apart from every
// other error resolveProjectParam and its two callers (parsePeriodAndProject
// above, parseChangeQuery in changes.go) can return: a malformed period, a
// missing "from"/"to", an ambiguous display name. All of those are genuine
// client errors and stay 400; a *projectResolverError is a store call that
// failed one line earlier than every other store call in these handlers,
// and deserves the identical treatment: routed through mapStoreError via
// writeParseOrStoreError below, for the same logged 500 and generic body.
//
// Before this existed, list() (changes.go), view() and listModels()
// (stats.go) classified *any* error out of parseChangeQuery /
// parsePeriodAndProject as a 400, so a database failure here -- as opposed
// to one line later, inside the store call each handler itself makes --
// produced an unlogged 400 whose body carried the store's raw internal
// text, instead of the logged 500 with a generic body every other store
// failure in these handlers gets (post-commit review round 2, F4/F5).
type projectResolverError struct {
	err error
}

func (e *projectResolverError) Error() string { return e.err.Error() }
func (e *projectResolverError) Unwrap() error { return e.err }

// writeParseOrStoreError reports the error parseChangeQuery or
// parsePeriodAndProject returned: a *projectResolverError is unwrapped and
// routed through mapStoreError, exactly as every other store failure in
// these handlers is; anything else is the client error it already was and
// stays a 400 with its own message intact.
func writeParseOrStoreError(w http.ResponseWriter, logger *slog.Logger, action string, err error) {
	var resolverErr *projectResolverError
	if errors.As(err, &resolverErr) {
		status, msg := mapStoreError(logger, action, resolverErr.err)
		writeError(w, status, msg)
		return
	}
	writeError(w, http.StatusBadRequest, err.Error())
}

func resolveProjectParam(ctx context.Context, resolver projectResolver, value string) (string, error) {
	if value == "" || looksLikeProjectKey(value) {
		return value, nil
	}
	keys, err := resolver.ProjectKeysByDisplayName(ctx, value)
	if err != nil {
		return "", &projectResolverError{err: err}
	}
	switch len(keys) {
	case 0:
		return value, nil
	case 1:
		return keys[0], nil
	default:
		return "", fmt.Errorf("project %q matches more than one project: %s", value, strings.Join(keys, ", "))
	}
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

	period, project, err := parsePeriodAndProject(r.Context(), h.store, values)
	if err != nil {
		writeParseOrStoreError(w, h.logger, "resolve project for "+string(name), err)
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

	recorded, err := h.recorded(r.Context(), period, project)
	if err != nil {
		status, msg := mapStoreError(h.logger, "resolve recorded period for "+string(name), err)
		writeError(w, status, msg)
		return
	}

	// Unmeasured is resolved only once Recorded is true (statsResponse's
	// own doc comment and the StatsStore interface's doc comment on
	// AllRecordedRunsUnmeasured): a period that predates any telemetry is
	// already fully reported by Recorded=false, so asking "did none of its
	// runs carry a measurement" would be a wasted query answering a
	// question that period does not raise.
	var unmeasured bool
	if recorded {
		unmeasured, err = h.store.AllRecordedRunsUnmeasured(r.Context(), period, project)
		if err != nil {
			status, msg := mapStoreError(h.logger, "resolve unmeasured period for "+string(name), err)
			writeError(w, status, msg)
			return
		}
	}

	resp := statsResponse{
		View:               name,
		From:               period.From.UTC().Format(time.RFC3339Nano),
		To:                 period.To.UTC().Format(time.RFC3339Nano),
		Project:            project,
		Model:              model,
		BoundaryConvention: boundaryConvention,
		Recorded:           recorded,
		Unmeasured:         unmeasured,
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
		cost, err := h.store.CostPerChange(ctx, period, project, model)
		if err != nil {
			s, m := mapStoreError(h.logger, "cost per change", err)
			return nil, s, m
		}
		// "change" scopes cost-per-change to one change, server-side --
		// task 21, step 5. Before this, a caller wanting one change's own
		// totals (the run-detail header, useRunDetail.ts) had to fetch
		// every change in the project and filter client-side, which stayed
		// correct only because this view is unpaged; the moment a limit is
		// added here that shape silently under-reports. This is a plain Go
		// slice filter, not a new store method: the result set CostPerChange
		// already returned is small and bounded by construction, so
		// re-filtering it here is the smaller surface.
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

// nextCommandFor maps a change's current pipeline state to the command that
// runs next. The pipeline is ONE command: `/flow` creates a change, resumes it,
// applies a fix and integrates it, choosing what to do from the state it finds.
// So the answer is `/flow` at every state that has a next step, and the value
// of this function is telling a dashboard reader that a step remains at all.
//
// It used to return `/myflow-do` and `/myflow-finish`, which is what a
// three-command pipeline needed and what this function kept returning after
// that pipeline was consolidated into `/flow`. Those commands do not exist:
// a reader following the dashboard's advice would have run nothing.
//
// This is NOT the retired-command-name exclusion that governs `/myflow-*`
// literals elsewhere in this package. Those are stored data -- values in
// `stage_runs` rows and the queries that select on them -- and renaming them
// would make the code disagree with history. This one is forward-looking
// guidance computed from current state, and it was simply wrong.
//
// `/flow-status` reports the same thing from the same record, so the two agree
// for a given change (spec: "the information matches what /flow-status
// reports").
func nextCommandFor(s store.State) string {
	switch s {
	case store.StateStarted, store.StateInProgress:
		return "/flow"
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
// is not the shape this step exists to forbid -- only an unbounded,
// project-blind blend across every same-named change is.
func filterCostPerChangeByName(rows []store.CostPerChangeRow, changeName string) []store.CostPerChangeRow {
	out := make([]store.CostPerChangeRow, 0, len(rows))
	for _, r := range rows {
		if r.ChangeName == changeName {
			out = append(out, r)
		}
	}
	return out
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

	period, project, err := parsePeriodAndProject(r.Context(), h.store, values)
	if err != nil {
		writeParseOrStoreError(w, h.logger, "resolve project for list models", err)
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
