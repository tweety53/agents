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
			MainCheckoutPath: "/tmp/uitest-alpha",
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
					},
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
				run, err := st.BeginStage(ctx, store.BeginStageInput{
					ProjectKey:   project.Key,
					ChangeName:   change.Name,
					Harness:      "claude-code",
					Command:      stage.Command,
					Stage:        stage.Stage,
					StartedAt:    stage.StartedAt,
					SessionToken: sessionToken,
				})
				if err != nil {
					return fmt.Errorf("uitest-seed: begin stage %s/%s %s: %w", project.Key, change.Name, stage.Stage, err)
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
		}
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

func strPtr(s string) *string        { return &s }
func timePtr(t time.Time) *time.Time { return &t }

func mergeJSON(ctx context.Context, st *store.Store, id int64, patch string) error {
	if err := st.MergeMetrics(ctx, id, json.RawMessage(patch)); err != nil {
		return fmt.Errorf("uitest-seed: merge metrics for stage run %d: %w", id, err)
	}
	return nil
}
