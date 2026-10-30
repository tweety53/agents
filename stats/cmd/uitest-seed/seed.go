package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/stages"
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
	if err := seedExperiments(ctx, st); err != nil {
		return err
	}
	if err := seedInteractive(ctx, st); err != nil {
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

// seedExperiments opens the Experiments view's two experiments (KAN-1067)
// over the fixture's existing changes: a closed implementer experiment,
// 2026-07-30 .. fixtureNow, whose before window [07-14, 07-30) holds the
// changes first run in the weeks of weeklyFixtureAt and whose after window
// holds the ones run near fixtureNow; and an open combined experiment
// started 2026-07-01, carrying experiment 4's three planning cells
// (planning-opus-medium-to-xhigh) so the roles panel draws the prototype's
// override text. An open experiment's windows end at the daemon's real
// now, its before window reaching back as far as its after window reaches
// forward, so the start sits where the before window already covers the
// fixture's first stage run (2026-03-09, seedReviewDynamics' week -8) on
// any day from 2026-10-23 on: the figures it draws never move again.
func seedExperiments(ctx context.Context, st *store.Store) error {
	closed, err := st.OpenExperiment(ctx, store.Experiment{Name: "implementer sonnet/flow-low",
		Cells: map[string][]string{"implementer": {"sonnet/flow-low"}}, StartedAt: time.Date(2026, 7, 30, 9, 0, 0, 0, time.UTC)})
	if err != nil {
		return fmt.Errorf("uitest-seed: open the closed experiment: %w", err)
	}
	if err := st.CloseExperimentAt(ctx, closed.ID, "negative",
		"Negative. Sonnet implementers cut tokens 22% but raised review rounds 2.2 → 3.1 and verify reds 0.6 → 1.3; wall-clock rose 26%. Table unchanged; experiment closed.",
		fixtureNow); err != nil {
		return fmt.Errorf("uitest-seed: close experiment %d: %w", closed.ID, err)
	}
	if _, err := st.OpenExperiment(ctx, store.Experiment{Name: "planning on xhigh",
		Cells: map[string][]string{
			"explorer":      {"sonnet/Explore@low"},
			"plan-designer": {"opus/Plan@medium", "opus/Plan@high", "opus/Plan@xhigh"},
			"planner":       {"opus/flow-planner-medium", "opus/flow-planner-high", "opus/flow-planner-xhigh"},
		},
		StartedAt: time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)}); err != nil {
		return fmt.Errorf("uitest-seed: open the combined experiment: %w", err)
	}
	// Explorer dispatches with tokens and duration on both sides of the open
	// experiment's start, so its Tokens/change and Wall-clock panels and
	// charts draw figures, as the prototype does.
	for _, x := range []struct {
		project, change, model string
		at                     time.Time
		tokens, minutes        int
	}{
		{"uitest-dynamics", "kan-300-dynamics-0", "opus", dynamicsFrom.Add(-8*7*24*time.Hour + 9*time.Hour), 52000, 14},
		{"uitest-dynamics", "kan-302-dynamics-0", "opus", dynamicsFrom.Add(9 * time.Hour), 48000, 12},
		{"uitest-dynamics", "kan-306-dynamics-0", "opus", dynamicsFrom.Add(4*7*24*time.Hour + 9*time.Hour), 45000, 11},
		{"uitest-alpha", "kan-103-weekly-fixture", "sonnet", weeklyFixtureAt.Add(-7*24*time.Hour + 30*time.Minute), 61000, 9},
		{"uitest-alpha", "kan-101-add-widget", "sonnet", weeklyFixtureAt.Add(30 * time.Minute), 58000, 8},
	} {
		// Attached to the change's first stage run, so no auto stage run is
		// opened and no other view's run count moves.
		runs, _, err := st.QueryStageRuns(ctx, store.Query{Filters: []store.Filter{
			{Field: "project", Op: store.OpEq, Value: x.project}, {Field: "name", Op: store.OpEq, Value: x.change}}, Limit: store.NoLimit})
		if err != nil || len(runs) == 0 {
			return fmt.Errorf("uitest-seed: stage run for %s: %d runs, %v", x.change, len(runs), err)
		}
		if _, err := st.RecordDispatch(ctx, x.project, x.change, records.Dispatch{Key: "explore-1", Role: "explorer",
			StageRunID: &runs[0].ID, Model: x.model, Effort: "low", Outcome: "completed", StartedAt: x.at,
			EndedAt: timePtr(x.at.Add(time.Duration(x.minutes) * time.Minute)),
			Metrics: json.RawMessage(fmt.Sprintf(`{"tokens":{"main":{"input":%d,"output":%d}}}`, x.tokens*9/10, x.tokens/10))}); err != nil {
			return fmt.Errorf("uitest-seed: record explorer dispatch for %s: %w", x.change, err)
		}
	}
	return nil
}

// The prototype's two /flow-interactive sessions (KAN-1061,
// .superpowers/sdd/planner/prototype-board.html), seeded under its change
// names in uitest-alpha inside every spec's pinned period.
const (
	interactiveRunning = "kan-1072-interactive-session-2026-10-30"
	interactiveWaiting = "kan-1073-interactive-session-2026-10-30"
)

// interactiveRun is one stage run of a seeded session: its request (0 for
// none), its start and duration in minutes before fixtureNow (dur < 0 leaves
// it open), its cost and its input tokens in thousands, 91% of them cache
// reads -- the prototype's run-view figures.
type interactiveRun struct {
	stage      string
	request    int
	start, dur int
	cost       float64
	ktokens    int
}

type interactiveRequest struct {
	title, outcome, commit, reason string
}

// seedInteractive writes the running session -- eight requests, the seventh
// in progress on an open flow.sdd-tdd begun three minutes before fixtureNow,
// 31 stage runs totalling $4.12 -- and the waiting session -- three landed
// requests, its last stage run ended six minutes before fixtureNow, $1.87.
func seedInteractive(ctx context.Context, st *store.Store) error {
	running := []interactiveRun{
		{"flow.kickoff", 0, 112, 1, 0.02, 550},
		{"flow.load-context", 1, 105, 1, 0.02, 600}, {"flow.sdd-tdd", 1, 103, 2, 0.30, 3400}, {"flow.verify", 1, 100, 1, 0.15, 1200},
		{"flow.landing-routes", 1, 98, 1, 0.14, 400}, {"flow.refresh-main-checkout", 1, 97, 1, 0.01, 250},
		{"flow.load-context", 2, 85, 1, 0.01, 600}, {"flow.sdd-tdd", 2, 83, 2, 0.32, 3600}, {"flow.verify", 2, 79, 1, 0.17, 1200},
		{"flow.sdd-tdd", 2, 77, 2, 0.28, 3500}, {"flow.verify", 2, 74, 1, 0.16, 1200}, {"flow.landing-routes", 2, 72, 1, 0.15, 400},
		{"flow.refresh-main-checkout", 2, 71, 1, 0.01, 250}, {"flow.visual-verify", 2, 69, 2, 0.06, 900},
		{"flow.load-context", 3, 60, 2, 0.01, 600},
		{"flow.load-context", 4, 55, 1, 0.01, 600}, {"flow.sdd-tdd", 4, 53, 3, 0.31, 3480}, {"flow.sdd-tdd", 4, 48, 2, 0.29, 3550},
		{"flow.verify", 4, 45, 2, 0.18, 1200}, {"flow.landing-routes", 4, 42, 1, 0.14, 400}, {"flow.refresh-main-checkout", 4, 41, 1, 0.01, 250},
		{"flow.load-context", 5, 40, 1, 0.01, 600},
		{"flow.load-context", 6, 32, 1, 0.01, 600}, {"flow.brainstorm", 6, 30, 3, 0.31, 1500}, {"flow.sdd-tdd", 6, 26, 9, 0.34, 3550},
		{"flow.verify", 6, 16, 5, 0.18, 980}, {"flow.landing-routes", 6, 11, 1, 0.06, 310}, {"flow.refresh-main-checkout", 6, 10, 2, 0.04, 220},
		{"flow.visual-verify", 6, 8, 2, 0.12, 700},
		{"flow.load-context", 7, 4, 1, 0.09, 610}, {"flow.sdd-tdd", 7, 3, -1, 0.21, 1200},
	}
	runningReqs := []interactiveRequest{
		{"Why no effort in stage runs?", "landed", "a41c9e2", ""},
		{`Rename "Running now" to "Active runs"`, "landed", "0b7d113", ""},
		{`What does "Errors" mean?`, "answered", "", ""},
		{"Count cache reads in input tokens", "landed", "9f3c2ab", ""},
		{"Move the cost rail to the left", "dropped", "", ""},
		{"Show the task list and the running agents", "landed", "c18a6d0", "new store fields and two panels"},
		{"Show the stage's model on the Agents row", "", "", ""},
		{"Remove the tool errors column", "", "", ""},
	}
	waiting := []interactiveRun{
		{"flow.kickoff", 0, 60, 1, 0.02, 300},
		{"flow.load-context", 1, 58, 1, 0.03, 500}, {"flow.sdd-tdd", 1, 56, 6, 0.40, 2400}, {"flow.verify", 1, 49, 2, 0.10, 900},
		{"flow.landing-routes", 1, 46, 1, 0.05, 300}, {"flow.refresh-main-checkout", 1, 45, 1, 0.02, 200},
		{"flow.load-context", 2, 42, 1, 0.03, 500}, {"flow.sdd-tdd", 2, 40, 5, 0.38, 2300}, {"flow.verify", 2, 34, 2, 0.10, 900},
		{"flow.landing-routes", 2, 31, 1, 0.05, 300}, {"flow.refresh-main-checkout", 2, 30, 1, 0.02, 200},
		{"flow.load-context", 3, 27, 1, 0.03, 500}, {"flow.sdd-tdd", 3, 25, 8, 0.40, 2400}, {"flow.verify", 3, 16, 3, 0.10, 900},
		{"flow.landing-routes", 3, 12, 1, 0.05, 300}, {"flow.refresh-main-checkout", 3, 9, 3, 0.09, 200},
	}
	waitingReqs := []interactiveRequest{
		{"Show the streak on the profile", "landed", "4d2e9a1", ""},
		{"Widen the weekly chart", "landed", "7b1f0c3", ""},
		{"Hide the empty streak row", "landed", "e5a8d27", ""},
	}
	if err := seedSession(ctx, st, interactiveRunning, "uitest-session-1072", running, runningReqs); err != nil {
		return err
	}
	if err := st.SetSessionActivity(ctx, "uitest-session-1072", "working", "Edit: stats/web/src/lib/phase.ts"); err != nil {
		return fmt.Errorf("uitest-seed: set activity for %s: %w", interactiveRunning, err)
	}
	if err := seedSession(ctx, st, interactiveWaiting, "uitest-session-1073", waiting, waitingReqs); err != nil {
		return err
	}
	return st.SetSessionActivity(ctx, "uitest-session-1073", "idle", "")
}

// seedSession writes one interactive session: its change, its requests --
// each received at its first run's start, an ended one ended at its last
// run's end -- and its stage runs, each tagged with its request.
func seedSession(ctx context.Context, st *store.Store, name, sessionID string, runs []interactiveRun, reqs []interactiveRequest) error {
	const project = "uitest-alpha"
	if err := st.PutChange(ctx, store.Change{ProjectKey: project, MainCheckoutPath: "/tmp/uitest-dynamics", Name: name,
		State: store.StateStarted, UpdatedAt: fixtureNow, UpdatedBy: "uitest-seed"}); err != nil {
		return fmt.Errorf("uitest-seed: put change %s: %w", name, err)
	}
	at := func(minutes int) time.Time { return fixtureNow.Add(-time.Duration(minutes) * time.Minute) }
	token := "fi-" + name[:8] + "-" + at(runs[0].start).Format("20060102T150405Z")
	for i, rq := range reqs {
		n := i + 1
		first, last := time.Time{}, time.Time{}
		for _, r := range runs {
			if r.request == n {
				if first.IsZero() {
					first = at(r.start)
				}
				last = at(r.start - r.dur)
			}
		}
		if first.IsZero() {
			first = at(1)
		}
		if _, err := st.BeginRequest(ctx, project, name, records.RequestBegin{N: n, Title: rq.title, BrainstormReason: rq.reason, ReceivedAt: &first}); err != nil {
			return fmt.Errorf("uitest-seed: begin request %d of %s: %w", n, name, err)
		}
		if rq.outcome == "" {
			continue
		}
		end := records.RequestEnd{N: n, Outcome: rq.outcome, EndedAt: &last}
		if rq.outcome == "landed" {
			end.CommitSHA = rq.commit + strings.Repeat("0", 33)
			end.BaseSHA = strings.Repeat("0", 40)
		}
		if _, err := st.EndRequest(ctx, project, name, end); err != nil {
			return fmt.Errorf("uitest-seed: end request %d of %s: %w", n, name, err)
		}
	}
	for _, r := range runs {
		var request *int
		if r.request > 0 {
			request = &r.request
		}
		run, err := st.BeginStage(ctx, store.BeginStageInput{ProjectKey: project, ChangeName: name, Harness: "claude-code",
			SessionID: strPtr(sessionID), SessionToken: strPtr(token), Command: "/flow-interactive", Stage: r.stage,
			StartedAt: at(r.start), RequestN: request})
		if err != nil {
			return fmt.Errorf("uitest-seed: begin %s of %s: %w", r.stage, name, err)
		}
		total := r.ktokens * 1000
		cached := total * 91 / 100
		tokens := map[string]any{"main": map[string]any{"input": total - cached, "cache_read": cached}}
		patch, err := json.Marshal(map[string]any{"cost_usd": r.cost, "effort": "xhigh", "tokens": tokens,
			"models": map[string]any{"claude-opus-5-5": map[string]any{"cost_usd": r.cost, "tokens": tokens}}})
		if err != nil {
			return fmt.Errorf("uitest-seed: marshal metrics for %s of %s: %w", r.stage, name, err)
		}
		if err := st.MergeMetrics(ctx, run.ID, patch); err != nil {
			return fmt.Errorf("uitest-seed: merge metrics for %s of %s: %w", r.stage, name, err)
		}
		if r.dur < 0 {
			continue
		}
		if err := st.EndStage(ctx, run.ID, at(r.start-r.dur), stages.OutcomeCompleted); err != nil {
			return fmt.Errorf("uitest-seed: end %s of %s: %w", r.stage, name, err)
		}
	}
	return nil
}
