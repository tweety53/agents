package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
// specs/myflow-ui-test-stack/spec.md's "The test stack starts populated".
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
							Command:      "/myflow-do",
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
							Command:      "/myflow-do",
							Stage:        "SDD + TDD per task",
							Model:        "claude-opus-5",
							InputTokens:  310000,
							OutputTokens: 48000,
							CostUSD:      2.75,
							StartedAt:    fixtureNow.Add(-26 * time.Hour),
							Outcome:      "completed",
						},
						{
							Command:      "/myflow-finish",
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
	return nil
}
