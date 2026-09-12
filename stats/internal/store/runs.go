package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// RunTotals is one run's (or dispatch's) derived usage figures -- tokens,
// cost, signals and wall-clock/gate time -- folded from the metrics bags
// Task 5's harvester writes.
type RunTotals struct {
	InputTokens, OutputTokens, ThinkingTokens, CacheRead, CacheWrite5m, CacheWrite1h int64
	CostUSD                                                                          *float64 // nil when no window carried cost_usd
	WallClockMs, HumanGateMs                                                         int64
	Compactions, Turns, ToolCalls, ToolErrors, Denials, APIErrors                    int64
	ContextEnd                                                                       *int64
}

// RunDispatchRow is one subagent dispatch inside a run.
type RunDispatchRow struct {
	Seq                                         int
	Role, Slot, AgentID, AgentType, Description string
	Depth                                       *int
	DeclaredModel, DeclaredEffort               string
	ServedModels, ServedEfforts                 map[string]int64
	Mismatch                                    bool
	FindingsRaised                              int            // findings rows whose dispatch_id is this dispatch
	FindingsByStatus                            map[string]int // the same rows counted by their current status
	StartedAt                                   time.Time
	EndedAt                                     *time.Time
	Totals                                      RunTotals
}

// RunRow is one recorded run: a plan session, or a /flow or /flow-fast
// invocation sharing one session token.
type RunRow struct {
	SessionToken, Kind, Command string // Kind: "plan" | "flow" | "flow-fast"
	StartedAt                   time.Time
	EndedAt                     *time.Time
	Totals, Main                RunTotals
	Decision                    *RunDecision // nil for a plan run or no decision row
	FanOutMax                   int          // most dispatches open at one instant
	SuiteRuns                   int          // suite_runs rows for the project inside the run's span
	SuiteFirstPass              *bool        // earliest such row's exit_code == 0; nil when none
	Dispatches                  []RunDispatchRow
}

// RunDecision is the three fields the runs view shows out of a recorded
// decision's JSON.
type RunDecision struct{ Execution, Implementer, Panel string }

// ChangeRuns is one change's (or one unattached plan session's) runs.
type ChangeRuns struct {
	Project           string
	Change            *string
	JiraKey           *string
	Totals            RunTotals
	IdleBetweenRunsMs int64
	FixIterations     int
	Runs              []RunRow
}

// humanGateStages are the stage keys whose whole duration is time spent
// waiting on the operator (spec: "Human gate time").
var humanGateStages = map[string]bool{
	"flow.design-approval": true, "flow.landing-question": true, "flow.unfinished-work-gate": true,
}

type runStageRow struct {
	ID           int64
	ChangeID     int64
	ProjectKey   string
	ChangeName   *string
	JiraKey      *string
	SessionToken string
	Command      string
	Stage        string
	StartedAt    time.Time
	EndedAt      *time.Time
	Metrics      map[string]json.RawMessage
}

type runDispatchRowRaw struct {
	ID            int64
	ChangeID      int64
	SessionToken  string
	Seq           int
	Role, Slot    string
	AgentID       *string
	Model, Effort string
	StartedAt     time.Time
	EndedAt       *time.Time
	Metrics       map[string]json.RawMessage
}

// ListRuns reads every period-bounded stage run (and, through it, every
// dispatch, finding, suite run and decision the run touched) and folds
// them into one row per run, grouped into change groups of runs.
//
// The three reads share one REPEATABLE READ, read-only transaction
// (queryTxOptions), so a write landing mid-read cannot produce a result
// whose dispatches or findings reference a stage run the stage-run read
// missed.
func (s *Store) ListRuns(ctx context.Context, period Period, project, change *string) ([]ChangeRuns, error) {
	tx, err := s.pool.BeginTx(ctx, queryTxOptions())
	if err != nil {
		return nil, fmt.Errorf("store: list runs: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	stageRows, err := tx.Query(ctx, `
		SELECT sr.id, COALESCE(sr.change_id, 0), COALESCE(c.project_key, sr.project_key), c.name,
		       COALESCE(c.jira_issue, sr.jira_key), COALESCE(sr.session_token, ''), sr.command, sr.stage,
		       sr.started_at, sr.ended_at, sr.metrics
		FROM stage_runs sr
		LEFT JOIN changes c ON c.id = sr.change_id
		WHERE sr.started_at >= $1 AND sr.started_at < $2
		  AND ($3::text IS NULL OR COALESCE(c.project_key, sr.project_key) = $3)
		  AND ($4::text IS NULL OR c.name = $4)
		ORDER BY sr.started_at, sr.id
	`, period.From, period.To, project, change)
	if err != nil {
		return nil, fmt.Errorf("store: list runs: stage runs: %w", err)
	}
	var stages []runStageRow
	for stageRows.Next() {
		var r runStageRow
		var metrics []byte
		if err := stageRows.Scan(&r.ID, &r.ChangeID, &r.ProjectKey, &r.ChangeName, &r.JiraKey, &r.SessionToken, &r.Command, &r.Stage, &r.StartedAt, &r.EndedAt, &metrics); err != nil {
			stageRows.Close()
			return nil, fmt.Errorf("store: list runs: scan stage run: %w", err)
		}
		_ = json.Unmarshal(metrics, &r.Metrics)
		stages = append(stages, r)
	}
	stageRows.Close()
	if err := stageRows.Err(); err != nil {
		return nil, fmt.Errorf("store: list runs: stage runs: %w", err)
	}
	if len(stages) == 0 {
		return []ChangeRuns{}, nil
	}

	dispatchRows, err := tx.Query(ctx, `
		SELECT d.id, d.change_id, COALESCE(d.session_token, ''), d.seq, d.role, COALESCE(d.slot, ''), d.agent_id,
		       d.model, d.effort, d.started_at, d.ended_at, d.metrics
		FROM dispatches d
		JOIN changes c ON c.id = d.change_id
		WHERE d.started_at >= $1 AND d.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		  AND ($4::text IS NULL OR c.name = $4)
		ORDER BY d.started_at, d.seq
	`, period.From, period.To, project, change)
	if err != nil {
		return nil, fmt.Errorf("store: list runs: dispatches: %w", err)
	}
	dispatchesByRun := map[string][]runDispatchRowRaw{}
	for dispatchRows.Next() {
		var d runDispatchRowRaw
		var metrics []byte
		if err := dispatchRows.Scan(&d.ID, &d.ChangeID, &d.SessionToken, &d.Seq, &d.Role, &d.Slot, &d.AgentID, &d.Model, &d.Effort, &d.StartedAt, &d.EndedAt, &metrics); err != nil {
			dispatchRows.Close()
			return nil, fmt.Errorf("store: list runs: scan dispatch: %w", err)
		}
		_ = json.Unmarshal(metrics, &d.Metrics)
		k := runKey(d.ChangeID, d.SessionToken)
		dispatchesByRun[k] = append(dispatchesByRun[k], d)
	}
	dispatchRows.Close()
	if err := dispatchRows.Err(); err != nil {
		return nil, fmt.Errorf("store: list runs: dispatches: %w", err)
	}

	findingRows, err := tx.Query(ctx, `
		SELECT f.dispatch_id, f.status, COUNT(*)
		FROM findings f
		JOIN dispatches d ON d.id = f.dispatch_id
		WHERE d.started_at >= $1 AND d.started_at < $2
		GROUP BY f.dispatch_id, f.status
	`, period.From, period.To)
	if err != nil {
		return nil, fmt.Errorf("store: list runs: findings: %w", err)
	}
	findingsByDispatch := map[int64]map[string]int{}
	for findingRows.Next() {
		var dispatchID int64
		var status string
		var n int
		if err := findingRows.Scan(&dispatchID, &status, &n); err != nil {
			findingRows.Close()
			return nil, fmt.Errorf("store: list runs: scan finding: %w", err)
		}
		if findingsByDispatch[dispatchID] == nil {
			findingsByDispatch[dispatchID] = map[string]int{}
		}
		findingsByDispatch[dispatchID][status] += n
	}
	findingRows.Close()
	if err := findingRows.Err(); err != nil {
		return nil, fmt.Errorf("store: list runs: findings: %w", err)
	}

	suiteRows, err := tx.Query(ctx, `
		SELECT project_key, ran_at, exit_code FROM suite_runs
		WHERE ran_at >= $1 AND ran_at < $2 AND ($3::text IS NULL OR project_key = $3)
		ORDER BY ran_at
	`, period.From, period.To, project)
	if err != nil {
		return nil, fmt.Errorf("store: list runs: suite runs: %w", err)
	}
	var suites []suiteRunRaw
	for suiteRows.Next() {
		var r suiteRunRaw
		if err := suiteRows.Scan(&r.ProjectKey, &r.RanAt, &r.ExitCode); err != nil {
			suiteRows.Close()
			return nil, fmt.Errorf("store: list runs: scan suite run: %w", err)
		}
		suites = append(suites, r)
	}
	suiteRows.Close()
	if err := suiteRows.Err(); err != nil {
		return nil, fmt.Errorf("store: list runs: suite runs: %w", err)
	}

	decisionRows, err := tx.Query(ctx, `
		SELECT DISTINCT ON (change_id, session_token) change_id, COALESCE(session_token, ''), decision
		FROM decisions
		ORDER BY change_id, session_token, recorded_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("store: list runs: decisions: %w", err)
	}
	decisions := map[string]*RunDecision{}
	for decisionRows.Next() {
		var changeID int64
		var token string
		var raw []byte
		if err := decisionRows.Scan(&changeID, &token, &raw); err != nil {
			decisionRows.Close()
			return nil, fmt.Errorf("store: list runs: scan decision: %w", err)
		}
		decisions[runKey(changeID, token)] = summarizeDecision(raw)
	}
	decisionRows.Close()
	if err := decisionRows.Err(); err != nil {
		return nil, fmt.Errorf("store: list runs: decisions: %w", err)
	}

	return groupRuns(stages, dispatchesByRun, findingsByDispatch, suites, decisions), nil
}

type suiteRunRaw struct {
	ProjectKey string
	RanAt      time.Time
	ExitCode   int
}

func runKey(changeID int64, token string) string {
	return strconv.FormatInt(changeID, 10) + "|" + token
}

// groupRuns folds period-bounded stage runs into change groups of runs.
// Group identity is the change id when attached, else (project, jira_key).
// Run identity is the session token within the group; a stage run with no
// token forms its own single-stage run keyed by its id.
func groupRuns(stages []runStageRow, dispatches map[string][]runDispatchRowRaw, findings map[int64]map[string]int, suites []suiteRunRaw, decisions map[string]*RunDecision) []ChangeRuns {
	type groupKey struct {
		changeID int64
		project  string
		jira     string
	}
	groups := map[groupKey]*ChangeRuns{}
	var order []groupKey
	runsByGroup := map[groupKey]map[string]*RunRow{}
	runOrder := map[groupKey][]string{}

	for _, sr := range stages {
		gk := groupKey{changeID: sr.ChangeID, project: sr.ProjectKey}
		if sr.ChangeID == 0 && sr.JiraKey != nil {
			gk.jira = *sr.JiraKey
		}
		g, ok := groups[gk]
		if !ok {
			g = &ChangeRuns{Project: sr.ProjectKey, Change: sr.ChangeName, JiraKey: sr.JiraKey}
			groups[gk] = g
			order = append(order, gk)
			runsByGroup[gk] = map[string]*RunRow{}
		}
		token := sr.SessionToken
		if token == "" {
			token = "stage-" + strconv.FormatInt(sr.ID, 10)
		}
		run, ok := runsByGroup[gk][token]
		if !ok {
			run = &RunRow{SessionToken: sr.SessionToken, Kind: runKind(sr.Command), Command: sr.Command, StartedAt: sr.StartedAt, EndedAt: sr.EndedAt}
			runsByGroup[gk][token] = run
			runOrder[gk] = append(runOrder[gk], token)
			run.Decision = decisions[runKey(sr.ChangeID, sr.SessionToken)]
			raw := dispatches[runKey(sr.ChangeID, sr.SessionToken)]
			run.Dispatches = buildDispatchRows(raw, stages, sr.ChangeID, sr.SessionToken, findings)
			run.FanOutMax = fanOutMax(raw)
		}
		if sr.StartedAt.Before(run.StartedAt) {
			run.StartedAt = sr.StartedAt
		}
		if sr.EndedAt == nil {
			run.EndedAt = nil
		} else if run.EndedAt != nil && sr.EndedAt.After(*run.EndedAt) {
			run.EndedAt = sr.EndedAt
		}
		addStageTotals(&run.Totals, &run.Main, sr)
	}

	out := make([]ChangeRuns, 0, len(order))
	for _, gk := range order {
		g := groups[gk]
		for _, token := range runOrder[gk] {
			run := runsByGroup[gk][token]
			run.Totals.WallClockMs = spanMs(run.StartedAt, run.EndedAt)
			run.Main.WallClockMs = run.Totals.WallClockMs
			run.SuiteRuns, run.SuiteFirstPass = suiteSummary(suites, g.Project, run.StartedAt, run.EndedAt)
			run.Main.HumanGateMs = run.Totals.HumanGateMs
			// Main cost = run cost minus what its dispatches were priced at.
			if run.Totals.CostUSD != nil {
				main := *run.Totals.CostUSD
				for _, d := range run.Dispatches {
					if d.Totals.CostUSD != nil {
						main -= *d.Totals.CostUSD
					}
				}
				run.Main.CostUSD = &main
			}
			g.Runs = append(g.Runs, *run)
		}
		sort.SliceStable(g.Runs, func(i, j int) bool { return g.Runs[i].StartedAt.Before(g.Runs[j].StartedAt) })
		for i, run := range g.Runs {
			sumTotals(&g.Totals, run.Totals)
			if run.Command == "/flow" && !firstStageIsKickoff(stages, gk.changeID, run.SessionToken) {
				g.FixIterations++
			}
			if i > 0 && g.Runs[i-1].EndedAt != nil && run.StartedAt.After(*g.Runs[i-1].EndedAt) {
				g.IdleBetweenRunsMs += run.StartedAt.Sub(*g.Runs[i-1].EndedAt).Milliseconds()
			}
		}
		out = append(out, *g)
	}
	return out
}

func runKind(command string) string {
	switch command {
	case "/flow-plan":
		return "plan"
	case "/flow-fast":
		return "flow-fast"
	default:
		return "flow"
	}
}

func spanMs(from time.Time, to *time.Time) int64 {
	if to == nil {
		return 0
	}
	return to.Sub(from).Milliseconds()
}

// fanOutMax sweeps dispatch intervals and reports the most that were
// open at one instant; an open-ended dispatch counts until the end.
func fanOutMax(raw []runDispatchRowRaw) int {
	type edge struct {
		at    time.Time
		delta int
	}
	var edges []edge
	for _, d := range raw {
		edges = append(edges, edge{d.StartedAt, +1})
		if d.EndedAt != nil {
			edges = append(edges, edge{*d.EndedAt, -1})
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].at.Equal(edges[j].at) {
			return edges[i].delta < edges[j].delta // close before open at the same instant
		}
		return edges[i].at.Before(edges[j].at)
	})
	open, max := 0, 0
	for _, e := range edges {
		open += e.delta
		if open > max {
			max = open
		}
	}
	return max
}

// suiteSummary counts the project's suite runs inside [from, to] and reports
// whether the earliest one passed. suite_runs carries no change or token, so
// the run's own span is the only link there is.
func suiteSummary(suites []suiteRunRaw, project string, from time.Time, to *time.Time) (int, *bool) {
	n := 0
	var first *bool
	for _, s := range suites { // ordered by ran_at
		if s.ProjectKey != project || s.RanAt.Before(from) || (to != nil && s.RanAt.After(*to)) {
			continue
		}
		if first == nil {
			ok := s.ExitCode == 0
			first = &ok
		}
		n++
	}
	return n, first
}

func firstStageIsKickoff(stages []runStageRow, changeID int64, token string) bool {
	for _, sr := range stages { // stages arrive ordered by started_at
		if sr.ChangeID == changeID && sr.SessionToken == token {
			return sr.Stage == "flow.kickoff"
		}
	}
	return false
}

type tokenBucket struct {
	Input, Output, Thinking, CacheRead, CacheCreation5m, CacheCreation1h int64
}

func readBucket(raw json.RawMessage) tokenBucket {
	var b struct {
		Input           int64 `json:"input"`
		Output          int64 `json:"output"`
		Thinking        int64 `json:"thinking"`
		CacheRead       int64 `json:"cache_read"`
		CacheCreation5m int64 `json:"cache_creation_5m"`
		CacheCreation1h int64 `json:"cache_creation_1h"`
	}
	_ = json.Unmarshal(raw, &b)
	return tokenBucket{b.Input, b.Output, b.Thinking, b.CacheRead, b.CacheCreation5m, b.CacheCreation1h}
}

type signalBag struct {
	Compactions    int64            `json:"compactions"`
	Turns          int64            `json:"turns"`
	ToolCallsTotal int64            `json:"tool_calls_total"`
	ToolErrors     int64            `json:"tool_errors"`
	Denials        int64            `json:"denials"`
	APIErrors      int64            `json:"api_errors"`
	ContextEnd     string           `json:"context_end"`
	ServedModels   map[string]int64 `json:"served_models"`
	ServedEfforts  map[string]int64 `json:"served_efforts"`
}

func readSignals(raw json.RawMessage) signalBag {
	var s signalBag
	_ = json.Unmarshal(raw, &s)
	return s
}

// side extracts metrics["tokens"][side] and metrics["signals"][side].
func side(metrics map[string]json.RawMessage, name string) (tokenBucket, signalBag) {
	var tokens, signals map[string]json.RawMessage
	_ = json.Unmarshal(metrics["tokens"], &tokens)
	_ = json.Unmarshal(metrics["signals"], &signals)
	return readBucket(tokens[name]), readSignals(signals[name])
}

func addBucket(t *RunTotals, b tokenBucket) {
	t.InputTokens += b.Input
	t.OutputTokens += b.Output
	t.ThinkingTokens += b.Thinking
	t.CacheRead += b.CacheRead
	t.CacheWrite5m += b.CacheCreation5m
	t.CacheWrite1h += b.CacheCreation1h
}

func addSignals(t *RunTotals, s signalBag) {
	t.Compactions += s.Compactions
	t.Turns += s.Turns
	t.ToolCalls += s.ToolCallsTotal
	t.ToolErrors += s.ToolErrors
	t.Denials += s.Denials
	t.APIErrors += s.APIErrors
	if n, err := strconv.ParseInt(s.ContextEnd, 10, 64); err == nil {
		t.ContextEnd = &n // last stage wins: rows arrive in started_at order
	}
}

func addCost(t *RunTotals, metrics map[string]json.RawMessage) {
	var cost float64
	if err := json.Unmarshal(metrics["cost_usd"], &cost); err != nil {
		return
	}
	if t.CostUSD == nil {
		t.CostUSD = new(float64)
	}
	*t.CostUSD += cost
}

func addStageTotals(total, main *RunTotals, sr runStageRow) {
	mt, ms := side(sr.Metrics, "main")
	st, ss := side(sr.Metrics, "sidechain")
	addBucket(total, mt)
	addBucket(total, st)
	addSignals(total, ms)
	addSignals(total, ss)
	addCost(total, sr.Metrics)
	addBucket(main, mt)
	addSignals(main, ms)
	if humanGateStages[sr.Stage] {
		total.HumanGateMs += spanMs(sr.StartedAt, sr.EndedAt)
	}
}

func sumTotals(dst *RunTotals, src RunTotals) {
	dst.InputTokens += src.InputTokens
	dst.OutputTokens += src.OutputTokens
	dst.ThinkingTokens += src.ThinkingTokens
	dst.CacheRead += src.CacheRead
	dst.CacheWrite5m += src.CacheWrite5m
	dst.CacheWrite1h += src.CacheWrite1h
	dst.WallClockMs += src.WallClockMs
	dst.HumanGateMs += src.HumanGateMs
	dst.Compactions += src.Compactions
	dst.Turns += src.Turns
	dst.ToolCalls += src.ToolCalls
	dst.ToolErrors += src.ToolErrors
	dst.Denials += src.Denials
	dst.APIErrors += src.APIErrors
	if src.CostUSD != nil {
		if dst.CostUSD == nil {
			dst.CostUSD = new(float64)
		}
		*dst.CostUSD += *src.CostUSD
	}
}

// buildDispatchRows builds one RunDispatchRow per dispatch belonging to
// (changeID, token). Tokens and signals come from the dispatch row's own
// metrics.tokens.sidechain / metrics.signals.sidechain; cost, agent type
// and depth come from the owning stage run's dispatches.<agentId> bucket
// (that is where store.Price writes per-dispatch cost), falling back to
// the dispatch row's own cost_usd.
func buildDispatchRows(raw []runDispatchRowRaw, stages []runStageRow, changeID int64, token string, findings map[int64]map[string]int) []RunDispatchRow {
	// Every dispatches.<agentId> bucket across this run's stage runs.
	buckets := map[string]map[string]json.RawMessage{}
	for _, sr := range stages {
		if sr.ChangeID != changeID || sr.SessionToken != token {
			continue
		}
		var per map[string]map[string]json.RawMessage
		if json.Unmarshal(sr.Metrics["dispatches"], &per) == nil {
			for id, b := range per {
				buckets[id] = b
			}
		}
	}
	out := make([]RunDispatchRow, 0, len(raw))
	for _, d := range raw {
		row := RunDispatchRow{Seq: d.Seq, Role: d.Role, Slot: d.Slot, DeclaredModel: d.Model, DeclaredEffort: d.Effort, StartedAt: d.StartedAt, EndedAt: d.EndedAt}
		if d.AgentID != nil {
			row.AgentID = *d.AgentID
		}
		tokens, signals := side(d.Metrics, "sidechain")
		addBucket(&row.Totals, tokens)
		addSignals(&row.Totals, signals)
		row.ServedModels, row.ServedEfforts = signals.ServedModels, signals.ServedEfforts
		row.Totals.WallClockMs = spanMs(d.StartedAt, d.EndedAt)
		addCost(&row.Totals, d.Metrics)
		if b, ok := buckets[row.AgentID]; ok && row.AgentID != "" {
			if row.Totals.CostUSD == nil {
				addCost(&row.Totals, b)
			}
			var s string
			if json.Unmarshal(b["agent_type"], &s) == nil {
				row.AgentType = s
			}
			if json.Unmarshal(b["description"], &s) == nil {
				row.Description = s
			}
			if json.Unmarshal(b["spawn_depth"], &s) == nil {
				if n, err := strconv.Atoi(s); err == nil {
					row.Depth = &n
				}
			}
		}
		row.Mismatch = mismatch(d.Model, d.Effort, signals)
		if byStatus := findings[d.ID]; len(byStatus) > 0 {
			row.FindingsByStatus = byStatus
			for _, n := range byStatus {
				row.FindingsRaised += n
			}
		}
		out = append(out, row)
	}
	return out
}

// mismatch flags a dispatch whose transcript never served the declared
// model, or (when an effort was declared and any effort observed) never
// served the declared effort. A dispatch with no served models is
// unmeasured, never mismatched.
func mismatch(model, effort string, s signalBag) bool {
	if len(s.ServedModels) == 0 {
		return false
	}
	if _, ok := s.ServedModels[strings.ToLower(model)]; !ok {
		return true
	}
	if effort != "" && effort != "default" && len(s.ServedEfforts) > 0 {
		if _, ok := s.ServedEfforts[effort]; !ok {
			return true
		}
	}
	return false
}

// summarizeDecision reduces a decision JSON to the three strings the runs
// view shows, using the same key shapes Decisions() reads (aggregate.go).
func summarizeDecision(raw []byte) *RunDecision {
	var d struct {
		Execution   string          `json:"execution"`
		Implementer json.RawMessage `json:"implementer"`
		Panel       *struct {
			Roster  []json.RawMessage `json:"roster"`
			Compact bool              `json:"compact"`
		} `json:"panel"`
	}
	if json.Unmarshal(raw, &d) != nil {
		return nil
	}
	out := &RunDecision{Execution: d.Execution}
	var s string
	if json.Unmarshal(d.Implementer, &s) == nil {
		out.Implementer = s
	} else {
		var obj struct{ Model, Effort string }
		if json.Unmarshal(d.Implementer, &obj) == nil {
			out.Implementer = obj.Model
			if obj.Effort != "" {
				out.Implementer += "/" + obj.Effort
			}
		}
	}
	if d.Panel != nil {
		out.Panel = fmt.Sprintf("%d slot", len(d.Panel.Roster))
		if len(d.Panel.Roster) != 1 {
			out.Panel += "s"
		}
		if d.Panel.Compact {
			out.Panel += ", compact"
		}
	}
	return out
}
