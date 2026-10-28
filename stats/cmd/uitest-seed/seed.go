package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// fixtureNow anchors every fixture row's timestamps. A fixed instant, not
// time.Now(), keeps the fixture's own contents deterministic across
// bring-ups -- the whole point of resetting the UI-test stack on every
// bring-up (design.md's "Reset on every bring-up") is that the operator
// sees the same thing every time.
var fixtureNow = time.Date(2026, 8, 15, 9, 0, 0, 0, time.UTC)

// weeklyFixtureAt is Thursday 2026-07-23, in the Wednesday week of 2026-07-22.
var weeklyFixtureAt = time.Date(2026, 7, 23, 9, 0, 0, 0, time.UTC)

// stageFixture is one stage run to write against a change, expressed at
// the level seedFixture's callers think in: which model recorded which
// token usage, and what it cost. cost_usd is written directly, alongside
// the token counts, rather than computed through Store.Price -- the
// seeder's own fixture data does not depend on the pricing table being
// populated yet, only on the daemon having seeded it before anyone reads
// the cost views (main.go's own doc comment on why the seeder does not
// call SeedPricing itself).
type stageFixture struct {
	Command string
	Stage   string
	Model   string

	InputTokens  float64
	OutputTokens float64
	CostUSD      float64

	StartedAt time.Time
	Outcome   string

	// ToolErrors is the run's main-session tool-error signal -- what the
	// live state board's Errors column sums. Zero writes no signal.
	ToolErrors int

	// SessionToken is the run correlator this stage was recorded under;
	// empty records none, so the stage counts as a run of its own.
	SessionToken string

	// Dispatches are the run's priced dispatch buckets, keyed by agent id
	// -- what the Weekly costs model·effort chart reads (KAN-1041).
	Dispatches map[string]pricedDispatch

	// SessionID is the harness session that ran the stage -- what the run
	// view groups stage runs by; Effort is the bag's top-level effort.
	SessionID string
	Effort    string

	// Open leaves the run open, with no metrics -- the live state board's
	// Running now card -- and Activity/ActivityDetail are its session's
	// activity (store.SetSessionActivity) when set.
	Open           bool
	Activity       string
	ActivityDetail string
}

// pricedDispatch is one dispatch bucket as the harvester and Store.Price
// leave it: the model and effort that served it, and its cost.
type pricedDispatch struct {
	Model   string
	Effort  string
	CostUSD float64
}

// changeFixture is one change and the stage runs recorded against it.
//
// SessionToken is the run correlator the decision and every dispatch are
// recorded under -- the (change, session_token) pair the Decisions view
// joins a run's totals on, so the stages of that run carry the same token.
// Decision is the planner's recorded JSON, written at DecisionAt;
// Dispatches are review-panel dispatches, each with the findings it raised.
//
// UpdatedAt is the change record's last update, fixtureNow when zero.
type changeFixture struct {
	Name      string
	State     store.State
	UpdatedAt time.Time
	Stage     []stageFixture

	SessionToken string
	Decision     string
	DecisionAt   time.Time
	Dispatches   []dispatchFixture

	// TaskCounts are the plan's task-count observations, first to last;
	// DoneTasks are the task ids an implementer dispatch committed.
	TaskCounts []int
	DoneTasks  []string
}

// dispatchFixture is one review-panel slot's dispatch and the findings it
// raised -- what the Reviewers view and Flow health's C/I/Mi tile count.
type dispatchFixture struct {
	Slot      string
	Model     string
	StartedAt time.Time
	Findings  []findingFixture
}

// findingFixture is one finding a dispatch raised.
type findingFixture struct {
	Ref      string
	Severity string
	Note     string
}

// guardFireFixture is one guard run that fired -- Flow health's
// "Guard fires" tile.
type guardFireFixture struct {
	ProjectKey string
	Guard      string
	Worktree   string
	RecordedAt time.Time
}

// projectFixture is one project and the changes recorded under it.
type projectFixture struct {
	Key              string
	MainCheckoutPath string
	Changes          []changeFixture
}

// fixtureData is the UI-test stack's committed fixture: two projects,
// changes spanning every pipeline state, and stage runs carrying token
// usage so the cost and statistics views have something to render, per
// the UI-test-stack requirement "The test stack starts populated".
func fixtureData() []projectFixture {
	return []projectFixture{
		{
			Key:              "uitest-alpha",
			MainCheckoutPath: "/tmp/uitest-dynamics",
			Changes: []changeFixture{
				{
					Name:  "kan-101-add-widget",
					State: store.StateStarted,
					// One run per weekly-costs category the other changes do
					// not carry, in the week of 2026-07-22 -- inside the
					// Weekly costs table's 8 weeks, before every spec's
					// pinned period, so no other view's capture moves
					// (KAN-1037).
					// The 2026-07-22 implementation run carries priced
					// dispatches for the Weekly costs model·effort chart
					// (KAN-1041); their cost sits inside the run's own.
					Stage: []stageFixture{
						{Command: "/flow", Stage: "flow.brainstorm", Model: "claude-opus-5", InputTokens: 60000, OutputTokens: 9000, CostUSD: 0.62, StartedAt: weeklyFixtureAt, Outcome: "completed"},
						{Command: "/flow", Stage: "flow.sdd-tdd", Model: "claude-opus-5", InputTokens: 120000, OutputTokens: 20000, CostUSD: 1.1, StartedAt: weeklyFixtureAt.Add(time.Hour), Outcome: "completed",
							Dispatches: map[string]pricedDispatch{
								"agent-w2-a": {Model: "claude-opus-5", Effort: "low", CostUSD: 0.3},
								"agent-w2-b": {Model: "claude-opus-5", Effort: "medium", CostUSD: 0.2},
							}},
						{Command: "/flow", Stage: "flow.self-review", Model: "claude-opus-5", InputTokens: 40000, OutputTokens: 6000, CostUSD: 0.35, StartedAt: weeklyFixtureAt.Add(2 * time.Hour), Outcome: "completed"},
						// The Running now strip's working card (kan-1048).
						{Command: "/flow", Stage: "flow.brainstorm", StartedAt: fixtureNow.Add(-12 * time.Minute), Open: true,
							SessionID: "uitest-session-101", Activity: "working", ActivityDetail: "2 subagents running"},
					},
				},
				{
					// A finished change last updated before every spec's
					// pinned period, so the state board never lists it: a
					// visual-verify run and a run outside every phase
					// (flow.verify) in the week of 2026-07-22, and a run with
					// priced dispatches the week before, so both Weekly costs
					// charts draw lines and labels across two consecutive
					// weeks (KAN-1041).
					Name:      "kan-103-weekly-fixture",
					State:     store.StateFinished,
					UpdatedAt: weeklyFixtureAt,
					Stage: []stageFixture{
						{Command: "/flow", Stage: "flow.sdd-tdd", Model: "claude-opus-5", InputTokens: 100000, OutputTokens: 16000, CostUSD: 0.9, StartedAt: weeklyFixtureAt.Add(-7 * 24 * time.Hour), Outcome: "completed",
							Dispatches: map[string]pricedDispatch{
								"agent-w1-a": {Model: "claude-opus-5", Effort: "low", CostUSD: 0.25},
								"agent-w1-b": {Model: "claude-opus-5", Effort: "medium", CostUSD: 0.3},
							}},
						{Command: "/flow", Stage: "flow.visual-verify", Model: "claude-opus-5", InputTokens: 30000, OutputTokens: 5000, CostUSD: 0.4, StartedAt: weeklyFixtureAt.Add(3 * time.Hour), Outcome: "completed"},
						{Command: "/flow", Stage: "flow.verify", Model: "claude-opus-5", InputTokens: 20000, OutputTokens: 3000, CostUSD: 0.2, StartedAt: weeklyFixtureAt.Add(4 * time.Hour), Outcome: "completed"},
					},
				},
				{
					Name:  "kan-102-fix-widget-bug",
					State: store.StateInProgress,
					Stage: []stageFixture{
						{
							Command:      "/flow",
							Stage:        "SDD + TDD per task",
							Model:        "claude-sonnet-5",
							InputTokens:  42000,
							OutputTokens: 9500,
							CostUSD:      0.179,
							StartedAt:    fixtureNow.Add(-2 * time.Hour),
							Outcome:      "failed",
							ToolErrors:   2,
							SessionToken: "uitest-kan-102",
						},
						{
							Command:      "/flow",
							Stage:        "flow.review-panel",
							Model:        "claude-sonnet-5",
							InputTokens:  90000,
							OutputTokens: 11000,
							CostUSD:      0.41,
							StartedAt:    fixtureNow.Add(-90 * time.Minute),
							Outcome:      "completed",
							SessionToken: "uitest-kan-102",
						},
						// The Running now strip's card awaiting an answer (kan-1048).
						{Command: "/flow", Stage: "flow.sdd-tdd", StartedAt: fixtureNow.Add(-20 * time.Minute), Open: true,
							SessionID: "uitest-session-102", Activity: "question", ActivityDetail: "Proceed with the panel-fix round on 2 findings?"},
					},
					TaskCounts:   []int{4, 5},
					DoneTasks:    []string{"1", "2", "3"},
					SessionToken: "uitest-kan-102",
					DecisionAt:   fixtureNow.Add(-2*time.Hour - 5*time.Minute),
					Decision: `{
						"class": "regular", "execution": "sdd",
						"implementer": {"model": "sonnet", "effort": "medium"},
						"panel": {
							"compact": false, "rerun": "delta",
							"roster": [
								{"slot": "primary", "experimental": false},
								{"slot": "principles", "experimental": false}
							]
						}
					}`,
					Dispatches: []dispatchFixture{
						{
							Slot:      "primary",
							Model:     "opus",
							StartedAt: fixtureNow.Add(-90 * time.Minute),
							Findings: []findingFixture{
								{Ref: "F1", Severity: "Critical", Note: "the fix drops the retry on a closed connection"},
								{Ref: "F2", Severity: "Minor", Note: "a stale comment names the old helper"},
							},
						},
						{
							Slot:      "principles",
							Model:     "opus",
							StartedAt: fixtureNow.Add(-90 * time.Minute),
							Findings: []findingFixture{
								{Ref: "F3", Severity: "Important", Note: "two callers now duplicate the same parse"},
							},
						},
					},
				},
			},
		},
		{
			Key:              "uitest-beta",
			MainCheckoutPath: "/tmp/uitest-beta",
			Changes: []changeFixture{
				{
					Name:  "kan-201-refactor-thing",
					State: store.StateFinished,
					Stage: []stageFixture{
						{
							Command:      "/flow",
							Stage:        "SDD + TDD per task",
							Model:        "claude-opus-5",
							InputTokens:  310000,
							OutputTokens: 48000,
							CostUSD:      2.75,
							StartedAt:    fixtureNow.Add(-26 * time.Hour),
							Outcome:      "completed",
							SessionToken: "uitest-kan-201",
							SessionID:    "uitest-session-201a",
							Effort:       "high",
						},
						{
							Command:      "/flow-fast",
							Stage:        "integrate",
							Model:        "claude-sonnet-5",
							InputTokens:  8000,
							OutputTokens: 1200,
							CostUSD:      0.028,
							StartedAt:    fixtureNow.Add(-time.Hour),
							Outcome:      "completed",
							SessionID:    "uitest-session-201b",
							Effort:       "low",
						},
					},
					SessionToken: "uitest-kan-201",
					DecisionAt:   fixtureNow.Add(-26*time.Hour - 5*time.Minute),
					Decision: `{
						"class": "big", "execution": "sdd",
						"implementer": {"model": "opus", "effort": "high"},
						"panel": "default"
					}`,
				},
			},
		},
	}
}

// guardFires is the fixture's fired guard runs.
func guardFires() []guardFireFixture {
	return []guardFireFixture{
		{
			ProjectKey: "uitest-beta",
			Guard:      "check-references",
			Worktree:   "/tmp/uitest-beta",
			RecordedAt: fixtureNow.Add(-3 * time.Hour),
		},
	}
}

// mainTokens is the metrics patch's "tokens.main" shape -- written twice
// per stage below (once at the patch's top level, once again nested under
// the stage's own model in "models"), so it is one literal, not two copies
// that could drift apart.
func mainTokens(input, output float64) map[string]any {
	return map[string]any{
		"main": map[string]any{
			"input":  input,
			"output": output,
		},
	}
}

// seedFixture writes fixtureData() through the store's exported types --
// PutChange, BeginStage and MergeMetrics -- never raw SQL, so a migration
// that changes a column stops this compiling instead of loading data into
// a shape the views render wrongly (spec's own "written against the
// application's own persistence types").
func seedFixture(ctx context.Context, st *store.Store) error {
	for _, project := range fixtureData() {
		for _, change := range project.Changes {
			updatedAt := change.UpdatedAt
			if updatedAt.IsZero() {
				updatedAt = fixtureNow
			}
			c := store.Change{
				ProjectKey:       project.Key,
				MainCheckoutPath: project.MainCheckoutPath,
				Name:             change.Name,
				State:            change.State,
				UpdatedAt:        updatedAt,
				UpdatedBy:        "uitest-seed",
			}
			if err := st.PutChange(ctx, c); err != nil {
				return fmt.Errorf("uitest-seed: put change %s/%s: %w", project.Key, change.Name, err)
			}

			for _, stage := range change.Stage {
				var sessionToken *string
				if stage.SessionToken != "" {
					sessionToken = strPtr(stage.SessionToken)
				}
				var sessionID *string
				if stage.SessionID != "" {
					sessionID = strPtr(stage.SessionID)
				}
				run, err := st.BeginStage(ctx, store.BeginStageInput{
					ProjectKey:   project.Key,
					ChangeName:   change.Name,
					Harness:      "claude-code",
					SessionID:    sessionID,
					Command:      stage.Command,
					Stage:        stage.Stage,
					StartedAt:    stage.StartedAt,
					SessionToken: sessionToken,
				})
				if err != nil {
					return fmt.Errorf("uitest-seed: begin stage %s/%s %s: %w", project.Key, change.Name, stage.Stage, err)
				}
				if stage.Activity != "" {
					if err := st.SetSessionActivity(ctx, stage.SessionID, stage.Activity, stage.ActivityDetail); err != nil {
						return fmt.Errorf("uitest-seed: set activity for %s/%s: %w", project.Key, change.Name, err)
					}
				}
				if stage.Open {
					continue
				}

				metrics := map[string]any{
					"cost_usd": stage.CostUSD,
					"tokens":   mainTokens(stage.InputTokens, stage.OutputTokens),
					"models": map[string]any{
						stage.Model: map[string]any{
							"cost_usd": stage.CostUSD,
							"tokens":   mainTokens(stage.InputTokens, stage.OutputTokens),
						},
					},
				}
				if len(stage.Dispatches) > 0 {
					dispatches := map[string]any{}
					for id, d := range stage.Dispatches {
						dispatches[id] = map[string]any{
							"model":    d.Model,
							"cost_usd": d.CostUSD,
							"signals": map[string]any{
								"served_models":  map[string]int{d.Model: 1},
								"served_efforts": map[string]int{d.Effort: 1},
							},
						}
					}
					metrics["dispatches"] = dispatches
				}
				if stage.Effort != "" {
					metrics["effort"] = stage.Effort
				}
				if stage.ToolErrors > 0 {
					metrics["signals"] = map[string]any{"main": map[string]any{"tool_errors": stage.ToolErrors}}
				}
				patch, err := json.Marshal(metrics)
				if err != nil {
					return fmt.Errorf("uitest-seed: marshal metrics for %s/%s %s: %w", project.Key, change.Name, stage.Stage, err)
				}
				if err := st.MergeMetrics(ctx, run.ID, patch); err != nil {
					return fmt.Errorf("uitest-seed: merge metrics for %s/%s %s: %w", project.Key, change.Name, stage.Stage, err)
				}

				endedAt := stage.StartedAt.Add(5 * time.Minute)
				if err := st.EndStage(ctx, run.ID, endedAt, stage.Outcome); err != nil {
					return fmt.Errorf("uitest-seed: end stage for %s/%s %s: %w", project.Key, change.Name, stage.Stage, err)
				}
			}

			if err := seedRunRecords(ctx, st, project.Key, change); err != nil {
				return err
			}
			if err := seedTaskProgress(ctx, st, project.Key, change); err != nil {
				return err
			}
		}
	}

	if err := seedReviewDynamics(ctx, st); err != nil {
		return err
	}

	for _, g := range guardFires() {
		if _, err := st.RecordGuardRun(ctx, g.ProjectKey, records.GuardRun{
			Guard:      g.Guard,
			Worktree:   g.Worktree,
			ExitCode:   1,
			Outcome:    "fired",
			DurationMS: 120,
			RecordedAt: g.RecordedAt,
		}); err != nil {
			return fmt.Errorf("uitest-seed: record guard run %s for %s: %w", g.Guard, g.ProjectKey, err)
		}
	}
	return nil
}

// seedRunRecords writes a change's decision and its review-panel
// dispatches with their findings, all under the change's session token.
func seedRunRecords(ctx context.Context, st *store.Store, projectKey string, change changeFixture) error {
	if change.Decision != "" {
		if _, _, err := st.RecordDecision(ctx, projectKey, change.Name, records.Decision{
			SessionToken: change.SessionToken,
			RecordedAt:   change.DecisionAt,
			Decision:     json.RawMessage(change.Decision),
		}); err != nil {
			return fmt.Errorf("uitest-seed: record decision for %s/%s: %w", projectKey, change.Name, err)
		}
	}
	for _, d := range change.Dispatches {
		out, err := st.RecordDispatch(ctx, projectKey, change.Name, records.Dispatch{
			Key:          "panel-" + d.Slot,
			Role:         "panel",
			Slot:         d.Slot,
			Model:        d.Model,
			Outcome:      "completed",
			SessionToken: change.SessionToken,
			StartedAt:    d.StartedAt,
			EndedAt:      timePtr(d.StartedAt.Add(4 * time.Minute)),
		})
		if err != nil {
			return fmt.Errorf("uitest-seed: record dispatch %s for %s/%s: %w", d.Slot, projectKey, change.Name, err)
		}
		for _, f := range d.Findings {
			if _, _, err := st.UpsertFinding(ctx, projectKey, change.Name, records.Finding{
				Ref:         f.Ref,
				DispatchSeq: &out.Seq,
				Round:       0,
				Slot:        d.Slot,
				Severity:    f.Severity,
				Note:        f.Note,
				Status:      "fixed",
			}); err != nil {
				return fmt.Errorf("uitest-seed: upsert finding %s for %s/%s: %w", f.Ref, projectKey, change.Name, err)
			}
		}
	}
	return nil
}

// seedTaskProgress writes a change's task-count observations and one
// committed implementer dispatch per done task -- the board's task
// progress (kan-1048).
func seedTaskProgress(ctx context.Context, st *store.Store, projectKey string, change changeFixture) error {
	for _, n := range change.TaskCounts {
		if _, err := st.RecordTaskCount(ctx, projectKey, change.Name, records.TaskCount{TotalTasks: n}); err != nil {
			return fmt.Errorf("uitest-seed: record task count for %s/%s: %w", projectKey, change.Name, err)
		}
	}
	for i, task := range change.DoneTasks {
		start := fixtureNow.Add(-time.Duration(60-10*i) * time.Minute)
		if _, err := st.RecordDispatch(ctx, projectKey, change.Name, records.Dispatch{
			Key:          "task-" + task + "-implementer",
			TaskID:       task,
			Role:         "implementer",
			Model:        "opus",
			CommitSHA:    fmt.Sprintf("%040d", i+1),
			Outcome:      "completed",
			SessionToken: change.SessionToken,
			StartedAt:    start,
			EndedAt:      timePtr(start.Add(8 * time.Minute)),
		}); err != nil {
			return fmt.Errorf("uitest-seed: record implementer dispatch for %s/%s: %w", projectKey, change.Name, err)
		}
	}
	return nil
}

func strPtr(s string) *string        { return &s }
func timePtr(t time.Time) *time.Time { return &t }

func mergeJSON(ctx context.Context, st *store.Store, id int64, patch string) error {
	if err := st.MergeMetrics(ctx, id, json.RawMessage(patch)); err != nil {
		return fmt.Errorf("uitest-seed: merge metrics for stage run %d: %w", id, err)
	}
	return nil
}

// dynamicsFrom is the first Monday of the Review dynamics fixture's eight
// weeks, 2026-05-04 .. 2026-06-29 -- before every other view's pinned
// window, so no other view's baseline moves (review-dynamics.spec.ts).
var dynamicsFrom = time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)

// seedReviewDynamics writes the Review dynamics view's fixture: two changes
// a week for the eight weeks from dynamicsFrom plus two in the previous
// range, each with a bundled panel dispatch and a test-sanity one on
// alternating models, versions and efforts, findings over rounds and
// severities, implementer dispatches (some blocked), visual-verify runs
// with fix re-runs, and one prompt change per area.
func seedReviewDynamics(ctx context.Context, st *store.Store) error {
	const project = "uitest-dynamics"
	type pick struct{ model, version, effort string }
	models := []pick{{"opus", "claude-opus-5-5", "medium"}, {"claude-opus-5", "", "low"}, {"sonnet", "claude-sonnet-5-5", "medium"}, {"opus", "", "high"}}
	severities := []string{"Critical", "major", "Important", "Minor", "minor", "Minor", "low", "Minor"}
	weeks := []int{-8, -5, 0, 1, 2, 3, 4, 5, 6, 7}
	for wi, w := range weeks {
		for c := 0; c < 2; c++ {
			name := fmt.Sprintf("kan-3%02d-dynamics-%d", wi, c)
			at := dynamicsFrom.Add(time.Duration(w*7+c*2)*24*time.Hour + 10*time.Hour)
			if err := st.PutChange(ctx, store.Change{ProjectKey: project, MainCheckoutPath: "/tmp/uitest-dynamics",
				Name: name, State: "FINISHED", UpdatedAt: at.Add(48 * time.Hour), UpdatedBy: "uitest-seed"}); err != nil {
				return fmt.Errorf("uitest-seed: put change %s: %w", name, err)
			}
			// The first change of each week is big, the second small, so the
			// page's Task size bar has data to draw (class-buttons).
			class := []string{"big", "small"}[c]
			if _, _, err := st.RecordDecision(ctx, project, name, records.Decision{SessionToken: "mf-" + name,
				RecordedAt: at.Add(-4 * time.Hour), Decision: json.RawMessage(`{"class":"` + class + `"}`)}); err != nil {
				return fmt.Errorf("uitest-seed: record decision for %s: %w", name, err)
			}
			dispatch := func(key, role, slot, model, effort, outcome string, start time.Time, version string) (records.Dispatch, error) {
				d, err := st.RecordDispatch(ctx, project, name, records.Dispatch{Key: key, Role: role, Slot: slot, Model: model,
					Effort: effort, Outcome: outcome, StartedAt: start, EndedAt: timePtr(start.Add(3 * time.Minute))})
				if err != nil {
					return d, fmt.Errorf("uitest-seed: record dispatch %s for %s: %w", key, name, err)
				}
				if version != "" {
					if err := st.MergeDispatchMetrics(ctx, d.ID, json.RawMessage(`{"modelVersion":"`+version+`"}`)); err != nil {
						return d, fmt.Errorf("uitest-seed: stamp %s for %s: %w", key, name, err)
					}
				}
				return d, nil
			}
			m := models[(wi+c)%len(models)]
			for t := 0; t < 2+(wi+c)%2; t++ {
				outcome := "completed"
				if t == 1 && (wi+c)%3 == 0 {
					outcome = "blocked"
				}
				if _, err := dispatch(fmt.Sprintf("task-%d-implementer", t+1), "implementer", "", m.model, "low", outcome, at.Add(-time.Duration(3-t)*time.Hour), m.version); err != nil {
					return err
				}
			}
			panel, err := dispatch("panel-0-primary+principles", "reviewer", "primary+principles", m.model, m.effort, "completed", at, m.version)
			if err != nil {
				return err
			}
			if _, err := dispatch("panel-0-test-sanity", "reviewer", "test-sanity", "sonnet", "low", "completed", at.Add(time.Minute), "claude-sonnet-5-5"); err != nil {
				return err
			}
			// Fewer findings as the weeks go by, so the trend reads.
			n := 8 - wi/2 - c
			rounds := 1 + (wi+c)%3
			for i := 0; i < n; i++ {
				slot := []string{"primary", "principles"}[i%2]
				if _, _, err := st.UpsertFinding(ctx, project, name, records.Finding{Ref: fmt.Sprintf("D%d", i+1),
					DispatchSeq: &panel.Seq, Round: i % (rounds + 1), Slot: slot, Severity: severities[(i+wi)%len(severities)],
					Note: "seeded review dynamics finding", Status: "fixed"}); err != nil {
					return fmt.Errorf("uitest-seed: upsert finding for %s: %w", name, err)
				}
			}
			fixes := (wi + c) % 3
			for f := 0; f <= fixes; f++ {
				key := "visual-verify"
				if f > 0 {
					key = fmt.Sprintf("visual-verify-fix-%d", f)
				}
				if _, err := dispatch(key, "verifier", "", "opus", "low", "completed", at.Add(time.Duration(2+f)*time.Hour), ""); err != nil {
					return err
				}
			}
		}
	}

	slot := "primary"
	for _, pc := range []store.PromptChange{
		{SHA: "2e053aa:skills/flow/primary-reviewer-prompt.md", At: dynamicsFrom.Add(17*24*time.Hour + 9*time.Hour), Kind: "edit", Area: "review",
			Slot: &slot, Target: "skills/flow/primary-reviewer-prompt.md", Summary: "primary reviewer re-reads every writer of state the diff reuses"},
		{SHA: "b1c09e2:skills/flow/visual-verify-verifier.md", At: dynamicsFrom.Add(31*24*time.Hour + 9*time.Hour), Kind: "edit", Area: "vv",
			Target: "skills/flow/visual-verify-verifier.md", Summary: "visual-verify-verifier: compare hover states"},
	} {
		if err := st.UpsertPromptChange(ctx, pc); err != nil {
			return fmt.Errorf("uitest-seed: prompt change %s: %w", pc.SHA, err)
		}
	}
	return nil
}
