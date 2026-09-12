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
}

// changeFixture is one change and the stage runs recorded against it.
type changeFixture struct {
	Name  string
	State store.State
	Stage []stageFixture
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
							Outcome:      "completed",
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
				},
			},
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
			c := store.Change{
				ProjectKey:       project.Key,
				MainCheckoutPath: project.MainCheckoutPath,
				Name:             change.Name,
				State:            change.State,
				UpdatedAt:        fixtureNow,
				UpdatedBy:        "uitest-seed",
			}
			if err := st.PutChange(ctx, c); err != nil {
				return fmt.Errorf("uitest-seed: put change %s/%s: %w", project.Key, change.Name, err)
			}

			for _, stage := range change.Stage {
				run, err := st.BeginStage(ctx, store.BeginStageInput{
					ProjectKey: project.Key,
					ChangeName: change.Name,
					Harness:    "claude-code",
					Command:    stage.Command,
					Stage:      stage.Stage,
					StartedAt:  stage.StartedAt,
				})
				if err != nil {
					return fmt.Errorf("uitest-seed: begin stage %s/%s %s: %w", project.Key, change.Name, stage.Stage, err)
				}

				patch, err := json.Marshal(map[string]any{
					"cost_usd": stage.CostUSD,
					"tokens":   mainTokens(stage.InputTokens, stage.OutputTokens),
					"models": map[string]any{
						stage.Model: map[string]any{
							"cost_usd": stage.CostUSD,
							"tokens":   mainTokens(stage.InputTokens, stage.OutputTokens),
						},
					},
				})
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
		}
	}
	if err := seedRunsFixture(ctx, st); err != nil {
		return err
	}
	return nil
}

// seedRunsFixture writes one change with a plan session, a creating run
// carrying two dispatches (one served by a different model than declared)
// and a fix run, so the runs view has every row kind to render.
func seedRunsFixture(ctx context.Context, st *store.Store) error {
	const project, change, jira = "uitest-alpha", "kan-103-runs-view", "KAN-103"
	// t0 is a fixed instant outside both fixtureNow's own window
	// (2026-08-01..2026-08-16, PINNED_QUERY) and the empty-period window
	// (2020-01-01..2020-01-02, EMPTY_QUERY) support.ts's other specs pin
	// to -- this fixture gets its own period (RUNS_QUERY,
	// 2026-07-01..2026-07-16) precisely so the 15 pre-existing visual
	// specs built around the smaller, original fixture's exact row
	// counts and cell values never see these rows.
	t0 := time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)

	plan, err := st.BeginStage(ctx, store.BeginStageInput{
		ProjectKey: project, MainCheckoutPath: "/tmp/uitest-alpha", JiraKey: jira, Harness: "claude-code",
		SessionToken: strPtr("fp-uitest-plan"), Command: "/flow-plan", Stage: "plan.session", StartedAt: t0,
	})
	if err != nil {
		return fmt.Errorf("uitest-seed: plan session: %w", err)
	}
	if err := mergeJSON(ctx, st, plan.ID, `{"tokens":{"main":{"input":12000,"output":3000,"cache_read":90000}},"cost_usd":0.41,
		"signals":{"main":{"compactions":1,"turns":18,"tool_calls_total":40,"tool_errors":2,"denials":1,"context_end":"95000"}}}`); err != nil {
		return err
	}
	if err := st.EndStage(ctx, plan.ID, t0.Add(25*time.Minute), "staged"); err != nil {
		return err
	}

	if err := st.PutChange(ctx, store.Change{ProjectKey: project, MainCheckoutPath: "/tmp/uitest-alpha", Name: change,
		// UpdatedAt uses t0, not fixtureNow: LiveStateBoard (state-board)
		// filters changes by updated_at directly, so a fixtureNow value
		// here would put this change back inside PINNED_QUERY's window
		// despite every stage run below sitting outside it, and reopen
		// exactly the state-board collision this fixture's own period
		// exists to avoid.
		State: store.StateInProgress, JiraIssue: strPtr(jira), UpdatedAt: t0, UpdatedBy: "uitest-seed"}); err != nil {
		return err
	}

	const token = "mf-uitest-run1"
	stage := func(key string, from, to time.Duration, metrics string) error {
		run, err := st.BeginStage(ctx, store.BeginStageInput{ProjectKey: project, ChangeName: change, Harness: "claude-code",
			SessionToken: strPtr(token), Command: "/flow", Stage: key, StartedAt: t0.Add(from)})
		if err != nil {
			return err
		}
		if metrics != "" {
			if err := mergeJSON(ctx, st, run.ID, metrics); err != nil {
				return err
			}
		}
		return st.EndStage(ctx, run.ID, t0.Add(to), "completed")
	}
	if err := stage("flow.kickoff", time.Hour, time.Hour+2*time.Minute, ""); err != nil {
		return err
	}
	if err := stage("flow.design-approval", time.Hour+2*time.Minute, time.Hour+14*time.Minute, ""); err != nil {
		return err
	}
	if err := stage("flow.sdd-tdd", time.Hour+14*time.Minute, time.Hour+70*time.Minute, `{
		"tokens":{"main":{"input":40000,"output":9000,"cache_read":300000},"sidechain":{"input":80000,"output":20000,"cache_read":500000}},
		"cost_usd":3.9,
		"signals":{"main":{"turns":30,"tool_calls_total":80,"tool_errors":3},"sidechain":{"turns":40,"tool_calls_total":120,"tool_errors":5,"denials":2}},
		"dispatches":{
			"agent-uitest-1":{"cost_usd":1.6,"agent_type":"flow-medium","description":"Task 1: store","spawn_depth":"1",
				"tokens":{"sidechain":{"input":50000,"output":12000,"cache_read":300000}},
				"signals":{"turns":25,"tool_calls_total":70,"served_models":{"claude-sonnet-5":25},"served_efforts":{"medium":25}}},
			"agent-uitest-2":{"cost_usd":1.1,"agent_type":"flow-high","description":"review: primary","spawn_depth":"1",
				"tokens":{"sidechain":{"input":30000,"output":8000,"cache_read":200000}},
				"signals":{"turns":15,"tool_calls_total":50,"tool_errors":5,"denials":2,"served_models":{"claude-opus-5":15},"served_efforts":{"high":15}}}}}`); err != nil {
		return err
	}
	dispatch := func(role, agentID, model, effort string, from, to time.Duration, metrics string) error {
		_, err := st.RecordDispatch(ctx, project, change, records.Dispatch{
			Role: role, AgentID: agentID, Model: model, Effort: effort, SessionToken: token,
			StartedAt: t0.Add(from), EndedAt: timePtr(t0.Add(to)), Outcome: "completed", Metrics: json.RawMessage(metrics),
		})
		return err
	}
	if err := dispatch("implementer", "agent-uitest-1", "claude-sonnet-5", "medium", time.Hour+15*time.Minute, time.Hour+45*time.Minute,
		`{"tokens":{"sidechain":{"input":50000,"output":12000,"cache_read":300000}},"signals":{"sidechain":{"turns":25,"tool_calls_total":70,"served_models":{"claude-sonnet-5":25},"served_efforts":{"medium":25}}}}`); err != nil {
		return err
	}
	if err := dispatch("reviewer", "agent-uitest-2", "claude-sonnet-5", "high", time.Hour+46*time.Minute, time.Hour+68*time.Minute,
		`{"tokens":{"sidechain":{"input":30000,"output":8000,"cache_read":200000}},"signals":{"sidechain":{"turns":15,"tool_calls_total":50,"tool_errors":5,"denials":2,"served_models":{"claude-opus-5":15},"served_efforts":{"high":15}}}}`); err != nil {
		return err
	}
	if _, _, err := st.RecordDecision(ctx, project, change, records.Decision{SessionToken: token,
		Decision: json.RawMessage(`{"class":"regular","execution":"subagent","implementer":{"model":"claude-sonnet-5","effort":"medium"},"panel":{"roster":[{"slot":"primary"}],"compact":true}}`)}); err != nil {
		return err
	}

	fix, err := st.BeginStage(ctx, store.BeginStageInput{ProjectKey: project, ChangeName: change, Harness: "claude-code",
		SessionToken: strPtr("mf-uitest-run2"), Command: "/flow", Stage: "flow.document-fix", StartedAt: t0.Add(4 * time.Hour)})
	if err != nil {
		return err
	}
	if err := mergeJSON(ctx, st, fix.ID, `{"tokens":{"main":{"input":9000,"output":2000,"cache_read":60000}},"cost_usd":0.3,"signals":{"main":{"turns":8,"tool_calls_total":20}}}`); err != nil {
		return err
	}
	return st.EndStage(ctx, fix.ID, t0.Add(4*time.Hour+12*time.Minute), "completed")
}

func strPtr(s string) *string        { return &s }
func timePtr(t time.Time) *time.Time { return &t }

func mergeJSON(ctx context.Context, st *store.Store, id int64, patch string) error {
	if err := st.MergeMetrics(ctx, id, json.RawMessage(patch)); err != nil {
		return fmt.Errorf("uitest-seed: merge metrics for stage run %d: %w", id, err)
	}
	return nil
}
