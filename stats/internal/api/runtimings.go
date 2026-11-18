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
	"net/http"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// RunTimingsStore is what the per-run route reads.
type RunTimingsStore interface {
	RunRecord(ctx context.Context, projectKey, change string) (records.Run, error)
	QueryStageRuns(ctx context.Context, q store.Query) ([]store.StageRun, int, error)
	ChangeGuardRuns(ctx context.Context, projectKey, change string, from, to time.Time) ([]records.GuardRun, int, error)
}

var _ RunTimingsStore = (*store.Store)(nil)

// WithRunTimings registers GET /api/v1/stats/run/{project}/{change}
// against st.
func WithRunTimings(st RunTimingsStore) Option {
	return func(o *serverOptions) { o.runTimings = st }
}

// usd is an amount in micro-dollars, served as dollars.
type usd int64

func (u usd) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(u)/1e6, 'f', -1, 64)), nil
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
			WallClockMs:        metric[int64]{wallMs, "First stage begin to last stage end; now while a stage is open.", kindMeasured},
			AgentTimeMs:        metric[int64]{agent, "Union of every dispatch row's interval.", kindDerived},
			ParentCommandsMs:   metric[int64]{parentMs, "Union of recorded parent-side runs (flow-guard commands) outside every dispatch.", kindMeasured},
			IdleMs:             metric[int64]{idleMs, "Wall clock minus the union of dispatches and parent commands: the parent's own turns, plan edits and waits.", kindDerived},
			Cost:               metric[usd]{total, "Sum of every dispatch row's priced cost_usd plus the parent session's own cost.", kindMeasured},
			ParentSessionCost:  metric[usd]{parentSession, "Stage-run cost minus the dispatch rows' cost, floored at zero.", kindDerived},
			VisualVerifyShare:  metric[float64]{visualShare(rows, wallMs), "Union of visual verifier rows, in-run fixes from visual verify and their gated reviewers and fixers, over wall clock.", kindDerived},
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
	store RunTimingsStore
	saved func(ctx context.Context) savedFile
	now   func() time.Time
	log   *slog.Logger
}

func (h *runTimingsHandler) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	project, change := r.PathValue("project"), r.PathValue("change")
	fail := func(what string, err error) {
		status, msg := mapStoreError(h.log, fmt.Sprintf("%s for %s/%s", what, project, change), err)
		writeError(w, status, msg)
	}
	run, err := h.store.RunRecord(ctx, project, change)
	if err != nil {
		fail("read the run record", err)
		return
	}
	stages, _, err := h.store.QueryStageRuns(ctx, store.Query{
		Filters: []store.Filter{
			{Field: "project", Op: store.OpEq, Value: project},
			{Field: "name", Op: store.OpEq, Value: change},
		},
		Limit: store.NoLimit,
	})
	if err != nil {
		fail("read the stage runs", err)
		return
	}
	now := h.now()
	wall, _ := wallClock(stages, run.Dispatches, now)
	guards, unclaimed, err := h.store.ChangeGuardRuns(ctx, project, change, time.UnixMilli(wall.from), time.UnixMilli(wall.to))
	if err != nil {
		fail("read the guard runs", err)
		return
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
	openers := map[string]string{}
	if b, ok := readPlanFile(store.LiveStateRow{ProjectKey: project, Name: change, RepoRoots: roots}, change, h.saved(ctx)); ok {
		openers = records.TaskOpeners(b)
	}
	out := computeRun(runInput{
		Dispatches: run.Dispatches, Stages: stages, Guards: guards, Unclaimed: unclaimed,
		Openers: openers, Repos: repoNames(paths),
	}, now)
	out.Project, out.Change = project, change
	writeJSON(w, http.StatusOK, out)
}
