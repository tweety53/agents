package api

// runtimings.go computes one /flow run's time share, parent commands, idle,
// cost and breakdowns from rows flowd already holds, and serves it at
// GET /api/v1/stats/run/{project}/{change} -- the per-run page opened from
// run detail (decision run-costs-separate-page).
//
//   - time-share-sweep: each elementary segment of wall clock is split
//     equally among the dispatch rows active in it; segments no row holds go
//     to parent commands or idle / parent, so every breakdown's rows sum to
//     wall clock exactly. Busy is a row's own union, shown beside it.
//   - group-cost-split-even: a task-13+14-… row's cost and time are split
//     evenly across its task ids, marked estimated.
//   - repo-from-key-suffix: a row's repository is the longest -<repo> key
//     suffix naming one of the change's repositories, else change-wide.
//   - parent-commands-outside-dispatches: parent commands are the union of
//     the change's guard runs minus the union of its dispatch rows; a run
//     nested in another collapses into it.
//
// Every figure is integer: time in milliseconds, cost in micro-dollars
// (usd), so a sum of rows equals its total exactly.

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// RunTimingsStore is what the per-run route reads.
type RunTimingsStore interface {
	RunRecord(ctx context.Context, projectKey, change string) (records.Run, error)
	QueryStageRuns(ctx context.Context, q store.Query) ([]store.StageRun, int, error)
	ChangeGuardRuns(ctx context.Context, projectKey, change string, from, to time.Time) ([]records.GuardRun, int, error)
	ListVerifyRuns(ctx context.Context, projectKey, change string) ([]records.VerifyRun, error)
	FinishedChangesInPeriod(ctx context.Context, period store.Period, project *string) ([]store.Change, error)
	ProjectKeysByDisplayName(ctx context.Context, displayName string) ([]string, error)
}

var _ RunTimingsStore = (*store.Store)(nil)

// WithRunTimings registers GET /api/v1/stats/run/{project}/{change}
// against st.
func WithRunTimings(st RunTimingsStore) Option {
	return func(o *serverOptions) { o.runTimings = st }
}

// WithStateRoot names the state root flowd counts a change's record
// journal under (decision flowd-counts-journal); without it the journal
// reads "unknown".
func WithStateRoot(root string) Option {
	return func(o *serverOptions) { o.stateRoot = root }
}

// usd is an amount in micro-dollars, served as dollars.
type usd int64

func (u usd) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(u)/1e6, 'f', -1, 64)), nil
}

// UnmarshalJSON reads the dollars MarshalJSON writes back to micro-dollars.
func (u *usd) UnmarshalJSON(b []byte) error {
	f, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}
	*u = usd(math.Round(f * 1e6))
	return nil
}

const (
	kindMeasured  = "measured"
	kindDerived   = "derived"
	kindEstimated = "estimated"
)

// metric is one figure with what it means and how it was obtained.
type metric[T any] struct {
	Value      T      `json:"value"`
	Definition string `json:"definition"`
	Kind       string `json:"kind"`
}

// definition names one column's meaning and kind.
type definition struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Kind       string `json:"kind"`
}

// RunTimings is one run's computed result.
type RunTimings struct {
	Project            string               `json:"project"`
	Change             string               `json:"change"`
	From               time.Time            `json:"from"`
	To                 time.Time            `json:"to"`
	Open               bool                 `json:"open"`
	Totals             runTotals            `json:"totals"`
	Lanes              []runLane            `json:"lanes"`
	Breakdowns         map[string]breakdown `json:"breakdowns"`
	CostPerTask        []taskCost           `json:"costPerTask"`
	CostPerVerifyRound []roundCost          `json:"costPerVerifyRound"`
	Counts             runCounts            `json:"counts"`
	DataQuality        []dqItem             `json:"dataQuality"`
	Definitions        []definition         `json:"definitions"`
}

type runTotals struct {
	WallClockMs       metric[int64]   `json:"wallClockMs"`
	AgentTimeMs       metric[int64]   `json:"agentTimeMs"`
	ParentCommandsMs  metric[int64]   `json:"parentCommandsMs"`
	IdleMs            metric[int64]   `json:"idleMs"`
	Cost              metric[usd]     `json:"cost"`
	ParentSessionCost metric[usd]     `json:"parentSessionCost"`
	VisualVerifyShare metric[float64] `json:"visualVerifyShare"`
	Dispatches        int             `json:"dispatches"`
	NoTokens          int             `json:"noTokens"`
	ParentCommandRuns int             `json:"parentCommandRuns"`
	// UnclaimedGuardRuns counts the project's main-checkout guard runs in
	// the window that no change could claim.
	UnclaimedGuardRuns int `json:"unclaimedGuardRuns"`
}

type breakdownRow struct {
	Label    string `json:"label"`
	Rows     int    `json:"rows"`
	ShareMs  int64  `json:"shareMs"`
	BusyMs   int64  `json:"busyMs"`
	Cost     usd    `json:"cost"`
	NoTokens int    `json:"noTokens"`
}

type breakdown struct {
	Columns []definition   `json:"columns"`
	Rows    []breakdownRow `json:"rows"`
	Total   breakdownRow   `json:"total"`
}

type taskCost struct {
	Task      string `json:"task"`
	Rows      int    `json:"rows"`
	ShareMs   int64  `json:"shareMs"`
	Cost      usd    `json:"cost"`
	Estimated bool   `json:"estimated"`
}

type roundCost struct {
	Round   string `json:"round"`
	Rows    int    `json:"rows"`
	ShareMs int64  `json:"shareMs"`
	Cost    usd    `json:"cost"`
}

type laneInterval struct {
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
	Class string    `json:"class"`
}

type runLane struct {
	Label     string         `json:"label"`
	Kind      string         `json:"kind"` // stage | category | parent | idle
	Intervals []laneInterval `json:"intervals"`
}

// runInput is everything computeRun reads, gathered by the handler.
type runInput struct {
	Dispatches []records.Dispatch
	Stages     []store.StageRun
	Guards     []records.GuardRun // the change's claimed guard runs
	Unclaimed  int
	Openers    map[string]string // task id -> **Opener:**
	Repos      []string          // the change's repository names
	Findings   []records.Finding
	VerifyRuns []records.VerifyRun
	Journal    journalCount // the change's record journal, counted by flowd
	PlanRead   bool         // tasks.md was read, live or saved
	Departures []roundDepartures
}

var breakdownColumns = []definition{
	{"Time share", "Each instant split equally among the rows active in it; rows sum to wall clock.", kindDerived},
	{"Busy", "Union of the row's own intervals; not additive.", kindDerived},
	{"Cost", "Sum of the rows' priced cost_usd; the parent session's own cost sits on idle / parent.", kindMeasured},
}

const (
	labelParent = "parent commands"
	labelIdle   = "idle / parent"
	changeWide  = "change-wide"
)

var (
	taskIDs     = regexp.MustCompile(`^tasks?-([\d+]+)-`)
	fixRoundKey = regexp.MustCompile(`^(visual-verify-fix|verify-fix|panel-fix|full-suite-fix)-\d+`)
)

type span struct{ from, to int64 } // [from, to) in Unix ms

// row is one dispatch with what computeRun derives for it.
type row struct {
	d        records.Dispatch
	iv       span
	category string
	catLabel string
	opener   string
	cost     usd
	noTokens bool
	share    int64
	tasks    []string
}

// computeRun is the pure computation behind the per-run route. now closes
// an open stage run and every open dispatch row.
func computeRun(in runInput, now time.Time) RunTimings {
	wall, open := wallClock(in.Stages, in.Dispatches, now)
	rows := make([]row, len(in.Dispatches))
	var dispatchCost usd
	for i, d := range in.Dispatches {
		r := row{d: d, iv: clip(dispatchSpan(d, wall.to), wall)}
		if m := taskIDs.FindStringSubmatch(d.Key); m != nil {
			r.tasks = strings.Split(m[1], "+")
			r.opener = in.Openers[r.tasks[0]]
		}
		r.category, r.catLabel = records.Category(d.Key, d.Role, r.opener)
		r.cost = costOf(d.Metrics)
		r.noTokens = records.CostReason(d) != ""
		dispatchCost += r.cost
		rows[i] = r
	}

	var guards []span
	for _, g := range in.Guards {
		guards = append(guards, clip(span{g.RecordedAt.UnixMilli() - int64(g.DurationMS), g.RecordedAt.UnixMilli()}, wall))
	}
	parentSegs, idleSegs := sweep(rows, guards, wall)
	parentMs, idleMs := length(parentSegs), length(idleSegs)
	parentRuns := countParentRuns(guards, parentSegs)

	var stageCost usd
	for _, s := range in.Stages {
		stageCost += costOf(s.Metrics)
	}
	parentSession := max(stageCost-dispatchCost, 0)
	total := dispatchCost + parentSession
	all := make([]span, len(rows))
	noTokens := 0
	for i, r := range rows {
		all[i] = r.iv
		if r.noTokens {
			noTokens++
		}
	}
	agent := length(union(all))
	wallMs := wall.to - wall.from

	out := RunTimings{
		From: time.UnixMilli(wall.from).UTC(), To: time.UnixMilli(wall.to).UTC(), Open: open,
		Totals: runTotals{
			WallClockMs:        metric[int64]{wallMs, "First stage begin to last stage end (now while a stage is open).", kindDerived},
			AgentTimeMs:        metric[int64]{agent, "Union of every dispatch row's [begin, end) — overlapping dispatches count once.", kindDerived},
			ParentCommandsMs:   metric[int64]{parentMs, "Union of recorded parent-side runs (flow-guard commands) outside every dispatch.", kindMeasured},
			IdleMs:             metric[int64]{idleMs, "Wall clock minus the union of dispatches and parent commands: the parent's own turns, plan edits and waits.", kindDerived},
			Cost:               metric[usd]{total, "Sum of every dispatch row's priced cost_usd plus the parent session's own cost.", kindMeasured},
			ParentSessionCost:  metric[usd]{parentSession, "Stage-run cost minus the dispatch rows' cost, floored at zero.", kindDerived},
			VisualVerifyShare:  metric[float64]{visualShare(rows, wallMs), "Union of visual verifier rows, in-run fixes from visual verify and their gated reviewers, over wall clock.", kindDerived},
			Dispatches:         len(rows),
			NoTokens:           noTokens,
			ParentCommandRuns:  parentRuns,
			UnclaimedGuardRuns: in.Unclaimed,
		},
		Breakdowns:         map[string]breakdown{},
		CostPerTask:        costPerTask(rows),
		CostPerVerifyRound: costPerVerifyRound(rows, in.Repos),
	}
	tail := []breakdownRow{
		{Label: labelParent, Rows: parentRuns, ShareMs: parentMs, BusyMs: parentMs},
		{Label: labelIdle, ShareMs: idleMs, BusyMs: idleMs, Cost: parentSession},
	}
	totalRow := breakdownRow{Label: "total", Rows: len(rows) + parentRuns, ShareMs: wallMs, Cost: total, NoTokens: noTokens}
	for by, label := range map[string]func(row) string{
		"category": func(r row) string { return r.catLabel },
		"role":     func(r row) string { return r.d.Role },
		"model": func(r row) string {
			if r.d.Effort == "" {
				return r.d.Model
			}
			return r.d.Model + " / " + r.d.Effort
		},
		"repo":  func(r row) string { return repoOf(r.d.Key, in.Repos) },
		"round": fixRound,
	} {
		out.Breakdowns[by] = breakdown{Columns: breakdownColumns, Rows: append(group(rows, label), tail...), Total: totalRow}
	}
	out.Lanes = lanes(in.Stages, rows, parentSegs, idleSegs, wall.to)
	out.Counts = counts(in, rows)
	out.DataQuality = dataQuality(in, rows, dispatchCost, stageCost)
	out.Definitions = runDefinitions(out)
	return out
}

// wallClock is the first stage begin to the last stage end, wall.to = now
// while a stage is open; with no stage run, the dispatch rows' extent.
func wallClock(stages []store.StageRun, ds []records.Dispatch, now time.Time) (span, bool) {
	var w span
	open := false
	first := true
	extend := func(from time.Time, to *time.Time) {
		end := now
		if to != nil {
			end = *to
		} else {
			open = true
		}
		if first || from.UnixMilli() < w.from {
			w.from = from.UnixMilli()
		}
		if first || end.UnixMilli() > w.to {
			w.to = end.UnixMilli()
		}
		first = false
	}
	for _, s := range stages {
		extend(s.StartedAt, s.EndedAt)
	}
	if len(stages) == 0 {
		for _, d := range ds {
			extend(d.StartedAt, d.EndedAt)
		}
	}
	if open {
		w.to = max(w.to, now.UnixMilli())
	}
	return w, open
}

func dispatchSpan(d records.Dispatch, wallEnd int64) span {
	end := wallEnd
	if d.EndedAt != nil {
		end = d.EndedAt.UnixMilli()
	}
	return span{d.StartedAt.UnixMilli(), end}
}

func clip(s, w span) span {
	s.from, s.to = max(s.from, w.from), min(s.to, w.to)
	if s.to < s.from {
		s.to = s.from
	}
	return s
}

func costOf(metrics json.RawMessage) usd {
	var m struct {
		CostUSD *float64 `json:"cost_usd"`
	}
	if json.Unmarshal(metrics, &m) != nil || m.CostUSD == nil {
		return 0
	}
	return usd(*m.CostUSD*1e6 + 0.5)
}

// sweep credits each row its equal share of every elementary segment it is
// active in -- a remainder millisecond to the earliest rows -- and returns
// the segments no row holds, split into parent commands (a guard run covers
// it) and idle / parent.
// ponytail: O(rows × boundaries); an interval tree if a run ever holds tens of thousands of rows.
func sweep(rows []row, guards []span, wall span) (parent, idle []span) {
	cuts := []int64{wall.from, wall.to}
	for _, r := range rows {
		cuts = append(cuts, r.iv.from, r.iv.to)
	}
	for _, g := range guards {
		cuts = append(cuts, g.from, g.to)
	}
	slices.Sort(cuts)
	cuts = slices.Compact(cuts)
	var active []int
	for k := 0; k+1 < len(cuts); k++ {
		a, b := cuts[k], cuts[k+1]
		active = active[:0]
		for i, r := range rows {
			if r.iv.from <= a && b <= r.iv.to {
				active = append(active, i)
			}
		}
		if n := int64(len(active)); n > 0 {
			q, rem := (b-a)/n, (b-a)%n
			for j, i := range active {
				rows[i].share += q
				if int64(j) < rem {
					rows[i].share++
				}
			}
			continue
		}
		seg := span{a, b}
		if slices.ContainsFunc(guards, func(g span) bool { return g.from <= a && b <= g.to }) {
			parent = appendMerged(parent, seg)
		} else {
			idle = appendMerged(idle, seg)
		}
	}
	return parent, idle
}

func appendMerged(ss []span, s span) []span {
	if n := len(ss); n > 0 && ss[n-1].to == s.from {
		ss[n-1].to = s.to
		return ss
	}
	return append(ss, s)
}

// countParentRuns counts the guard runs not nested in another that add
// parent-command time.
func countParentRuns(guards, parent []span) int {
	n := 0
	for i, g := range guards {
		nested := false
		for j, o := range guards {
			if i != j && o.from <= g.from && g.to <= o.to && (o != g || j < i) {
				nested = true
				break
			}
		}
		if !nested && slices.ContainsFunc(parent, func(p span) bool { return p.from < g.to && g.from < p.to }) {
			n++
		}
	}
	return n
}

func union(ss []span) []span {
	ss = slices.Clone(ss)
	slices.SortFunc(ss, func(a, b span) int { return cmp.Compare(a.from, b.from) })
	var out []span
	for _, s := range ss {
		if s.to <= s.from {
			continue
		}
		if n := len(out); n > 0 && s.from <= out[n-1].to {
			out[n-1].to = max(out[n-1].to, s.to)
			continue
		}
		out = append(out, s)
	}
	return out
}

func length(ss []span) int64 {
	var n int64
	for _, s := range ss {
		n += s.to - s.from
	}
	return n
}

// group folds rows by label, largest share first.
func group(rows []row, label func(row) string) []breakdownRow {
	idx := map[string]int{}
	var out []breakdownRow
	var spans [][]span
	for _, r := range rows {
		l := label(r)
		i, ok := idx[l]
		if !ok {
			i = len(out)
			idx[l] = i
			out = append(out, breakdownRow{Label: l})
			spans = append(spans, nil)
		}
		out[i].Rows++
		out[i].ShareMs += r.share
		out[i].Cost += r.cost
		if r.noTokens {
			out[i].NoTokens++
		}
		spans[i] = append(spans[i], r.iv)
	}
	for i := range out {
		out[i].BusyMs = length(union(spans[i]))
	}
	slices.SortStableFunc(out, func(a, b breakdownRow) int {
		return cmp.Or(cmp.Compare(b.ShareMs, a.ShareMs), cmp.Compare(a.Label, b.Label))
	})
	return out
}

// fixRound is the loop and round a row's key names, or for an in-run fix
// task's rows its Opener's, else "—".
func fixRound(r row) string {
	if m := fixRoundKey.FindString(r.d.Key); m != "" {
		return m
	}
	if m := fixRoundKey.FindString(r.opener); m != "" {
		return m
	}
	return "—"
}

// repoNames are the repository names a change's paths carry: the main
// checkout's basename of a <name>-worktrees/<change> path, else the path's
// own basename.
func repoNames(paths []string) []string {
	var out []string
	for _, p := range paths {
		name := filepath.Base(p)
		if i := strings.Index(p, "-worktrees/"); i >= 0 {
			name = filepath.Base(p[:i])
		}
		if !slices.Contains(out, name) {
			out = append(out, name)
		}
	}
	return out
}

// repoOf is the longest -<repo> suffix of key among repos, else change-wide.
func repoOf(key string, repos []string) string {
	best := changeWide
	for _, r := range repos {
		if strings.HasSuffix(key, "-"+r) && (best == changeWide || len(r) > len(best)) {
			best = r
		}
	}
	return best
}

// costPerTask splits each task-naming row's cost and time share evenly
// across the task ids it names, a group's marked estimated.
func costPerTask(rows []row) []taskCost {
	idx := map[string]int{}
	out := []taskCost{}
	for _, r := range rows {
		n := int64(len(r.tasks))
		for j, id := range r.tasks {
			i, ok := idx[id]
			if !ok {
				i = len(out)
				idx[id] = i
				out = append(out, taskCost{Task: id})
			}
			c := &out[i]
			c.Rows++
			c.ShareMs += splitPart(r.share, n, j)
			c.Cost += usd(splitPart(int64(r.cost), n, j))
			c.Estimated = c.Estimated || n > 1
		}
	}
	slices.SortFunc(out, func(a, b taskCost) int {
		x, _ := strconv.Atoi(a.Task)
		y, _ := strconv.Atoi(b.Task)
		return cmp.Compare(x, y)
	})
	return out
}

// splitPart is the j-th of n near-equal parts of v; the parts sum to v.
func splitPart(v, n int64, j int) int64 {
	p := v / n
	if int64(j) < v%n {
		p++
	}
	return p
}

// verifyRound is the verify round a row belongs to: a verify or visual
// verifier row's key less its repo suffix, or the round its task's Opener
// names; "" for none.
func verifyRound(r row, repos []string) string {
	strip := func(k string) string {
		if repo := repoOf(k, repos); repo != changeWide {
			return strings.TrimSuffix(k, "-"+repo)
		}
		return k
	}
	switch {
	case r.category == "verify-list" || r.category == "visual-verifier":
		return strip(r.d.Key)
	case strings.HasPrefix(r.opener, "verify") || strings.HasPrefix(r.opener, "visual-verify"):
		return strip(r.opener)
	}
	return ""
}

func costPerVerifyRound(rows []row, repos []string) []roundCost {
	idx := map[string]int{}
	out := []roundCost{}
	for _, r := range rows {
		round := verifyRound(r, repos)
		if round == "" {
			continue
		}
		i, ok := idx[round]
		if !ok {
			i = len(out)
			idx[round] = i
			out = append(out, roundCost{Round: round})
		}
		out[i].Rows++
		out[i].ShareMs += r.share
		out[i].Cost += r.cost
	}
	return out
}

func visualShare(rows []row, wallMs int64) float64 {
	if wallMs == 0 {
		return 0
	}
	var ss []span
	for _, r := range rows {
		if r.category == "visual-verifier" || strings.HasPrefix(r.opener, "visual-verify") {
			ss = append(ss, r.iv)
		}
	}
	return float64(length(union(ss))) / float64(wallMs)
}

// lanes are the stage runs, one lane per category in the table's order
// (other:<role> after, by name), parent commands and idle / parent.
func lanes(stages []store.StageRun, rows []row, parent, idle []span, wallEnd int64) []runLane {
	iv := func(s span, class string) laneInterval {
		return laneInterval{time.UnixMilli(s.from).UTC(), time.UnixMilli(s.to).UTC(), class}
	}
	var out []runLane
	stageIdx := map[string]int{}
	for _, s := range stages {
		i, ok := stageIdx[s.Stage]
		if !ok {
			i = len(out)
			stageIdx[s.Stage] = i
			out = append(out, runLane{Label: s.Stage, Kind: "stage"})
		}
		class, end := "open", wallEnd
		if s.EndedAt != nil {
			end = s.EndedAt.UnixMilli()
			if s.Outcome != nil {
				class = *s.Outcome
			}
		}
		out[i].Intervals = append(out[i].Intervals, iv(span{s.StartedAt.UnixMilli(), end}, class))
	}
	order := func(id string) int {
		if i := slices.IndexFunc(records.Categories, func(c records.CategoryRow) bool { return c.ID == id }); i >= 0 {
			return i
		}
		return len(records.Categories)
	}
	byCat := map[string]*runLane{}
	var cats []string
	for _, r := range rows {
		l, ok := byCat[r.category]
		if !ok {
			l = &runLane{Label: r.catLabel, Kind: "category"}
			byCat[r.category] = l
			cats = append(cats, r.category)
		}
		class := r.d.Outcome
		if r.d.EndedAt == nil {
			class = "open"
		}
		l.Intervals = append(l.Intervals, iv(r.iv, class))
	}
	slices.SortFunc(cats, func(a, b string) int { return cmp.Or(cmp.Compare(order(a), order(b)), cmp.Compare(a, b)) })
	for _, c := range cats {
		out = append(out, *byCat[c])
	}
	for _, l := range []struct {
		label, kind string
		ss          []span
	}{{labelParent, "parent", parent}, {labelIdle, "idle", idle}} {
		lane := runLane{Label: l.label, Kind: l.kind, Intervals: []laneInterval{}}
		for _, s := range l.ss {
			lane.Intervals = append(lane.Intervals, iv(s, l.kind))
		}
		out = append(out, lane)
	}
	return out
}

// runTimingsHandler serves the per-run route.
type runTimingsHandler struct {
	store     RunTimingsStore
	stateRoot string
	saved     func(ctx context.Context) savedFile
	now       func() time.Time
	log       *slog.Logger
}

func (h *runTimingsHandler) get(w http.ResponseWriter, r *http.Request) {
	project, change := r.PathValue("project"), r.PathValue("change")
	out, err := h.run(r.Context(), project, change)
	if err != nil {
		status, msg := mapStoreError(h.log, fmt.Sprintf("%s for %s/%s", err.what, project, change), err.err)
		writeError(w, status, msg)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// runError names the read that failed computing one run.
type runError struct {
	what string
	err  error
}

// run computes one change's result from the store; the per-run route and
// the run-costs fold both call it, so the two never disagree.
func (h *runTimingsHandler) run(ctx context.Context, project, change string) (RunTimings, *runError) {
	fail := func(what string, err error) (RunTimings, *runError) { return RunTimings{}, &runError{what, err} }
	run, err := h.store.RunRecord(ctx, project, change)
	if err != nil {
		return fail("read the run record", err)
	}
	stages, _, err := h.store.QueryStageRuns(ctx, store.Query{
		Filters: []store.Filter{
			{Field: "project", Op: store.OpEq, Value: project},
			{Field: "name", Op: store.OpEq, Value: change},
		},
		Limit: store.NoLimit,
	})
	if err != nil {
		return fail("read the stage runs", err)
	}
	now := h.now()
	wall, _ := wallClock(stages, run.Dispatches, now)
	guards, unclaimed, err := h.store.ChangeGuardRuns(ctx, project, change, time.UnixMilli(wall.from), time.UnixMilli(wall.to))
	if err != nil {
		return fail("read the guard runs", err)
	}
	var roots, paths []string
	for _, s := range stages {
		if s.RepoRoot != nil && !slices.Contains(roots, *s.RepoRoot) {
			roots = append(roots, *s.RepoRoot)
		}
	}
	paths = append(paths, roots...)
	for _, g := range guards {
		paths = append(paths, g.Worktree)
	}
	verifyRuns, err := h.store.ListVerifyRuns(ctx, project, change)
	if err != nil {
		return fail("read the verify runs", err)
	}
	saved := h.saved(ctx)
	openers := map[string]string{}
	b, planRead := readPlanFile(store.LiveStateRow{ProjectKey: project, Name: change, RepoRoots: roots}, change, saved)
	if planRead {
		openers = records.TaskOpeners(b)
	}
	out := computeRun(runInput{
		Dispatches: run.Dispatches, Stages: stages, Guards: guards, Unclaimed: unclaimed,
		Openers: openers, Repos: repoNames(paths),
		Findings: run.Findings, VerifyRuns: verifyRuns, PlanRead: planRead,
		Journal:    h.journal(project, change),
		Departures: departures(project, change, run.Dispatches, roots, saved),
	}, now)
	out.Project, out.Change = project, change
	return out, nil
}

// journalCount is the change's pending record writes, served as a number,
// or as "unknown" when flowd could not read the journal: a failed count is
// never reported as the clean zero.
type journalCount struct {
	n     int
	known bool
}

func (j journalCount) MarshalJSON() ([]byte, error) {
	if !j.known {
		return []byte(`"unknown"`), nil
	}
	return []byte(strconv.Itoa(j.n)), nil
}

// roundDepartures is one visual round's departures: line.
type roundDepartures struct {
	Round      string `json:"round"`
	Departures int    `json:"departures"`
}

type fixRoundCounts struct {
	Verify int `json:"verify"` // verify-fix-N and visual-verify-fix-N
	Panel  int `json:"panel"`
}

type severityCounts struct {
	Critical  int `json:"critical"`
	Important int `json:"important"`
	Minor     int `json:"minor"`
}

type flakyCounts struct {
	Retries int `json:"retries"`
	Runs    int `json:"runs"` // verify runs that recorded a flaky count
}

type outcomeCounts struct {
	Open    int `json:"open"`
	Aborted int `json:"aborted"`
	Blocked int `json:"blocked"`
}

type runCounts struct {
	FixRounds      metric[fixRoundCounts]    `json:"fixRounds"`
	Departures     metric[[]roundDepartures] `json:"departures"`
	Findings       metric[severityCounts]    `json:"findings"`
	FlakyRetries   metric[flakyCounts]       `json:"flakyRetries"`
	Rows           metric[outcomeCounts]     `json:"rows"`
	JournalPending metric[journalCount]      `json:"journalPending"`
}

// The counts' terms, worded as the prototype's titles.
var (
	defFixRounds  = definition{"Fix rounds", "Distinct verify-fix-N and visual-verify-fix-N rounds, plus panel-fix rounds.", kindDerived}
	defDepartures = definition{"Departures / visual round", "The departures: line of each visual-verify report, by round.", kindMeasured}
	defFindings   = definition{"Findings", "Recorded review findings by canonical severity.", kindMeasured}
	defFlaky      = definition{"Flaky retries", "Tests that passed only on their one retry, summed over recorded verify runs.", kindMeasured}
	defRows       = definition{"Rows open · aborted · blocked", "Dispatch rows by outcome.", kindMeasured}
	defJournal    = definition{"Journal pending", "Record writes for this change still in the local journal.", kindMeasured}
	defPerTask    = definition{"Cost per task", "Every dispatch whose key names the task (implementer, fixer, reviewer); a group key's cost is split evenly across its tasks and marked estimated.", kindDerived}
	defPerRound   = definition{"Cost per verify round", "A round's verify rows plus the in-run fix tasks whose Opener names that round, with their gated reviewers.", kindDerived}
)

// prototypeDefinitions is the Definitions panel's own list, in its order.
var prototypeDefinitions = []definition{
	{"Wall clock", "First stage begin to last stage end; now while a stage is open.", kindDerived},
	{"Agent time", "Union of dispatch rows' [begin, end) — overlap counted once. A category's time is the union of its own rows.", kindDerived},
	{"Parent commands", "Recorded parent-side command runs (flow-guard), union outside every dispatch.", kindMeasured},
	{"Idle / parent", "Wall clock minus the union of dispatches and parent commands.", kindDerived},
	{"Cost", "Priced cost_usd from harvested tokens; a row with no tokens states its reason in Data quality.", kindMeasured},
	{"Apportioned cost", "A row whose tokens were shared across overlapping inline windows and split by time.", kindEstimated},
	{"Category", "Read from the dispatch key and, for a task implementer, the task's Opener — one table; changing it re-labels history.", kindDerived},
}

func measure[T any](v T, d definition) metric[T] {
	return metric[T]{Value: v, Definition: d.Definition, Kind: d.Kind}
}

func counts(in runInput, rows []row) runCounts {
	verify, panel := map[string]bool{}, map[string]bool{}
	var oc outcomeCounts
	for _, r := range rows {
		switch fr := fixRound(r); {
		case strings.HasPrefix(fr, "panel-fix-"):
			panel[fr] = true
		case strings.HasPrefix(fr, "verify-fix-"), strings.HasPrefix(fr, "visual-verify-fix-"):
			verify[fr] = true
		}
		switch {
		case r.d.EndedAt == nil:
			oc.Open++
		case r.d.Outcome == "aborted":
			oc.Aborted++
		case r.d.Outcome == "blocked":
			oc.Blocked++
		}
	}
	var sev severityCounts
	for _, f := range in.Findings {
		switch strings.ToLower(f.Severity) {
		case "critical":
			sev.Critical++
		case "important":
			sev.Important++
		case "minor":
			sev.Minor++
		}
	}
	var fl flakyCounts
	for _, v := range in.VerifyRuns {
		if v.Flaky != nil {
			fl.Retries += *v.Flaky
			fl.Runs++
		}
	}
	deps := in.Departures
	if deps == nil {
		deps = []roundDepartures{}
	}
	return runCounts{
		FixRounds:      measure(fixRoundCounts{len(verify), len(panel)}, defFixRounds),
		Departures:     measure(deps, defDepartures),
		Findings:       measure(sev, defFindings),
		FlakyRetries:   measure(fl, defFlaky),
		Rows:           measure(oc, defRows),
		JournalPending: measure(in.Journal, defJournal),
	}
}

// dqItem is one gap the run still carries. Unknown marks a count flowd
// could not take, which is never reported as zero.
type dqItem struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Count   int      `json:"count"`
	Unknown bool     `json:"unknown,omitempty"`
	Detail  []string `json:"detail"`
}

// unreconciledAt is the gap between the dispatch rows' cost and the stage
// runs' at which the totals stop being believable: one cent.
const unreconciledAt usd = 10_000

func dataQuality(in runInput, rows []row, dispatchCost, stageCost usd) []dqItem {
	item := func(id, label string) dqItem { return dqItem{ID: id, Label: label, Detail: []string{}} }
	add := func(it *dqItem, detail string) { it.Count++; it.Detail = append(it.Detail, detail) }
	open, superseded := item("open-stage-runs", "open stage runs"), item("superseded-stage-runs", "superseded stage runs")
	for _, s := range in.Stages {
		name := fmt.Sprintf("%s (run %d)", s.Stage, s.ID)
		if s.EndedAt == nil {
			add(&open, name)
		}
		if s.Outcome != nil && *s.Outcome == "superseded" {
			add(&superseded, name)
		}
	}
	unlinked := item("unlinked", "dispatch rows with no stage link")
	noTokens := item("no-tokens", "dispatch rows with no tokens — reason stated")
	abortedRan := item("aborted-ran", "aborted rows that ran")
	other := item("uncategorised", "uncategorised dispatch keys")
	openers := item("unreadable-openers", "task rows whose Opener could not be read")
	retimed := item("retimed", "retimed dispatch rows")
	for _, r := range rows {
		key := r.d.Key
		if r.d.StageRunID == nil {
			add(&unlinked, key)
		}
		reason := records.CostReason(r.d)
		if reason != "" {
			add(&noTokens, key+" · "+reason)
		}
		if r.d.Outcome == "aborted" && reason == "" {
			add(&abortedRan, key)
		}
		if strings.HasPrefix(r.category, "other:") {
			add(&other, key+" · "+r.category)
		}
		if r.tasks != nil && !in.PlanRead {
			add(&openers, key)
		}
		if r.d.Retimed != nil {
			add(&retimed, key)
		}
	}
	unreconciled := item("unreconciled-cost", "dispatch cost above stage cost")
	if dispatchCost-stageCost >= unreconciledAt {
		add(&unreconciled, fmt.Sprintf("dispatch rows $%.2f · stage runs $%.2f", float64(dispatchCost)/1e6, float64(stageCost)/1e6))
	}
	journal := item("journal", "record writes pending in the journal")
	journal.Count, journal.Unknown = in.Journal.n, !in.Journal.known
	unclaimed := item("unclaimed-guard-runs", "main-checkout guard runs no change claimed")
	unclaimed.Count = in.Unclaimed
	return []dqItem{open, superseded, unlinked, noTokens, abortedRan, other, openers, unreconciled, journal, retimed, unclaimed}
}

// runDefinitions lists every term the result serves once: the prototype's
// Definitions panel first, in its order, then each other metric's.
func runDefinitions(out RunTimings) []definition {
	t := out.Totals
	more := []definition{
		{"Parent session cost", t.ParentSessionCost.Definition, t.ParentSessionCost.Kind},
		{"Visual-verify loops", t.VisualVerifyShare.Definition, t.VisualVerifyShare.Kind},
	}
	more = append(more, breakdownColumns...)
	more = append(more, defFixRounds, defDepartures, defFindings, defFlaky, defRows, defJournal, defPerTask, defPerRound)
	defs := slices.Clone(prototypeDefinitions)
	for _, d := range more {
		if !slices.ContainsFunc(defs, func(e definition) bool { return e.Name == d.Name }) {
			defs = append(defs, d)
		}
	}
	return defs
}

// departuresLine is the visual verifier report's `- departures: <n>` line
// in its ## Verdict (skills/flow/visual-verify-verifier.md).
var departuresLine = regexp.MustCompile(`(?m)^- departures:\s*(\d+)\s*$`)

// parseDepartures reads a visual verifier report's departures, false when
// the report carries no such line.
func parseDepartures(report []byte) (int, bool) {
	m := departuresLine.FindSubmatch(report)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(string(m[1]))
	return n, err == nil
}

// journal counts the change's record journal at flowd's state root, the
// path the CLI appends to (fallback.RecordJournalPath).
func (h *runTimingsHandler) journal(project, change string) journalCount {
	if h.stateRoot == "" {
		return journalCount{}
	}
	path := fallback.RecordJournalPath(h.stateRoot, project, change)
	n, err := fallback.CountJournalEntries(path)
	if err != nil {
		h.log.Warn("count the record journal", "path", path, "err", err)
		return journalCount{}
	}
	return journalCount{n: n, known: true}
}

// departures reads each visual verifier row's report: the worktree's
// .superpowers/sdd/verify-report-<key>.md first, then the saved
// reports.<repo>.verify-report-<key>.md copy. A round whose report is found
// nowhere, or carries no departures: line, is left out.
func departures(project, change string, ds []records.Dispatch, roots []string, saved savedFile) []roundDepartures {
	var out []roundDepartures
	for _, d := range ds {
		if id, _ := records.Category(d.Key, d.Role, ""); id != "visual-verifier" {
			continue
		}
		file := "verify-report-" + d.Key + ".md"
		var report []byte
		for _, root := range roots {
			if b, err := os.ReadFile(filepath.Join(root, ".superpowers", "sdd", file)); err == nil {
				report = b
				break
			}
		}
		for _, repo := range repoNames(roots) {
			if report != nil || saved == nil {
				break
			}
			report, _ = saved(project, change, "reports."+repo+"."+file)
		}
		if n, ok := parseDepartures(report); ok {
			out = append(out, roundDepartures{Round: d.Key, Departures: n})
		}
	}
	return out
}

// runCostsCategory is one category folded over the period's runs.
type runCostsCategory struct {
	Category    string  `json:"category"`
	Runs        int     `json:"runs"`
	Rows        int     `json:"rows"`
	Seconds     int64   `json:"seconds"` // the runs' busy (union) time, summed
	Cost        usd     `json:"cost"`
	ShareOfCost float64 `json:"shareOfCost"`
}

// runCostsRun is one finished run's line, its values copied from its
// one-run result.
type runCostsRun struct {
	Project      string                 `json:"project"`
	Change       string                 `json:"change"`
	WallClockMs  metric[int64]          `json:"wallClockMs"`
	AgentTimeMs  metric[int64]          `json:"agentTimeMs"`
	IdleMs       metric[int64]          `json:"idleMs"`
	Cost         metric[usd]            `json:"cost"`
	Tasks        metric[int]            `json:"tasks"`
	FixRounds    metric[fixRoundCounts] `json:"fixRounds"`
	CategoryCost map[string]usd         `json:"categoryCost"`
}

type runCosts struct {
	Columns    []definition       `json:"columns"`
	Categories []runCostsCategory `json:"categories"`
	Runs       []runCostsRun      `json:"runs"`
}

var (
	defTasks        = definition{"Tasks", "Task ids with a cost-per-task row.", kindDerived}
	runCostsColumns = []definition{
		{"Runs", "Finished runs in the period with a row of the category.", kindDerived},
		{"Rows", "Dispatch rows of the category, summed over the runs.", kindMeasured},
		{"Time (union)", "Each run's Busy for the category, summed over the runs.", kindDerived},
		{"Cost", "Each run's category cost, summed over the runs.", kindMeasured},
		{"Share of cost", "The category's cost over every category's cost.", kindDerived},
	}
)

// foldRunCosts folds each run's one-run result into the across-runs view.
func foldRunCosts(results []RunTimings) runCosts {
	out := runCosts{Columns: runCostsColumns, Categories: []runCostsCategory{}, Runs: []runCostsRun{}}
	byCat := map[string]*runCostsCategory{}
	var order []string
	var total usd
	for _, r := range results {
		catCost := map[string]usd{}
		for _, row := range r.Breakdowns["category"].Rows {
			if row.Label == labelParent || row.Label == labelIdle {
				continue
			}
			c, ok := byCat[row.Label]
			if !ok {
				c = &runCostsCategory{Category: row.Label}
				byCat[row.Label] = c
				order = append(order, row.Label)
			}
			c.Runs++
			c.Rows += row.Rows
			c.Seconds += row.BusyMs
			c.Cost += row.Cost
			total += row.Cost
			catCost[row.Label] = row.Cost
		}
		out.Runs = append(out.Runs, runCostsRun{
			Project: r.Project, Change: r.Change,
			WallClockMs: r.Totals.WallClockMs, AgentTimeMs: r.Totals.AgentTimeMs, IdleMs: r.Totals.IdleMs,
			Cost: r.Totals.Cost, Tasks: measure(len(r.CostPerTask), defTasks), FixRounds: r.Counts.FixRounds,
			CategoryCost: catCost,
		})
	}
	for _, k := range order {
		c := byCat[k]
		c.Seconds /= 1000
		if total > 0 {
			c.ShareOfCost = float64(c.Cost) / float64(total)
		}
		out.Categories = append(out.Categories, *c)
	}
	return out
}

// costs serves GET /api/v1/stats/run-costs: every finished run in the
// period computed by the per-run function and folded.
func (h *runTimingsHandler) costs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	period, project, err := parsePeriodAndProject(ctx, h.store, r.URL.Query())
	if err != nil {
		writeParseOrStoreError(w, h.log, "resolve project for run-costs", err)
		return
	}
	changes, err := h.store.FinishedChangesInPeriod(ctx, period, project)
	if err != nil {
		status, msg := mapStoreError(h.log, "list finished changes for run-costs", err)
		writeError(w, status, msg)
		return
	}
	results := make([]RunTimings, 0, len(changes))
	for _, c := range changes {
		res, rerr := h.run(ctx, c.ProjectKey, c.Name)
		if rerr != nil {
			status, msg := mapStoreError(h.log, fmt.Sprintf("%s for %s/%s", rerr.what, c.ProjectKey, c.Name), rerr.err)
			writeError(w, status, msg)
			return
		}
		results = append(results, res)
	}
	writeJSON(w, http.StatusOK, foldRunCosts(results))
}
