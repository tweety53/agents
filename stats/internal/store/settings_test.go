package store_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestSettingsStore_RoundTrip asserts that writing a well-formed Settings
// value through PutSettings and reading it back through GetSettings
// returns exactly what was written -- the contract /flow-settings depends
// on to both set and later display the harness-wide defaults.
func TestSettingsStore_RoundTrip(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	want := store.Settings{
		Reviewers: []string{"primary", "principles", "failure-modes", "mutation"},
	}

	if err := st.PutSettings(ctx, want); err != nil {
		t.Fatalf("PutSettings: %v", err)
	}

	got, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetSettings = %+v, want %+v", got, want)
	}
}

// TestSettingsStore_RoundTripUpdates asserts a second PutSettings call
// overwrites the first rather than being rejected or silently ignored --
// flow_settings holds exactly one row, and re-running /flow-settings must
// change it.
func TestSettingsStore_RoundTripUpdates(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	first := store.Settings{Reviewers: []string{"primary"}}
	second := store.Settings{Reviewers: []string{"primary", "failure-modes"}}

	if err := st.PutSettings(ctx, first); err != nil {
		t.Fatalf("first PutSettings: %v", err)
	}
	gotFirst, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings after first PutSettings: %v", err)
	}
	if !reflect.DeepEqual(gotFirst, first) {
		t.Errorf("GetSettings = %+v, want %+v (the first write)", gotFirst, first)
	}

	if err := st.PutSettings(ctx, second); err != nil {
		t.Fatalf("second PutSettings: %v", err)
	}

	got, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if !reflect.DeepEqual(got, second) {
		t.Errorf("GetSettings = %+v, want %+v (the second write)", got, second)
	}
}

// TestSettingsStore_RejectsUnknownReviewer asserts PutSettings refuses a
// reviewers entry off the fixed reviewer-slot vocabulary, wrapping
// store.ErrInvalidReviewer and naming the specific bad value in the error.
func TestSettingsStore_RejectsUnknownReviewer(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	const badReviewer = "style-nitpicker"
	err := st.PutSettings(ctx, store.Settings{
		Reviewers: []string{"primary", badReviewer},
	})
	if err == nil {
		t.Fatal("PutSettings: got nil error, want ErrInvalidReviewer")
	}
	if !errors.Is(err, store.ErrInvalidReviewer) {
		t.Errorf("PutSettings error = %v, want it to wrap ErrInvalidReviewer", err)
	}
	if !strings.Contains(err.Error(), badReviewer) {
		t.Errorf("PutSettings error = %q, want it to name the rejected value %q", err.Error(), badReviewer)
	}
}

// TestSettingsStore_RejectsRetiredCodeReviewLow asserts the retired
// "code-review-low" slot id is refused on write like any unknown id, while
// a flow_settings row written before its retirement still loads unchanged:
// validation guards writes only, so a stale row never breaks GetSettings.
func TestSettingsStore_RejectsRetiredCodeReviewLow(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()

	err := st.PutSettings(ctx, store.Settings{
		Reviewers: []string{"primary", "code-review-low"},
	})
	if !errors.Is(err, store.ErrInvalidReviewer) {
		t.Fatalf("PutSettings error = %v, want it to wrap ErrInvalidReviewer", err)
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO flow_settings (id, reviewers)
		VALUES (TRUE, '["primary","principles","code-review-low"]')
		ON CONFLICT (id) DO UPDATE SET reviewers = EXCLUDED.reviewers
	`); err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}
	got, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings on a legacy row: %v", err)
	}
	want := []string{"primary", "principles", "code-review-low"}
	if !reflect.DeepEqual(got.Reviewers, want) {
		t.Errorf("GetSettings reviewers = %v, want %v", got.Reviewers, want)
	}
}

// TestSettingsStore_RejectsRetiredBugbotAndSecurity asserts the retired
// "bugbot" and "security" slot ids are refused on write like any unknown id.
func TestSettingsStore_RejectsRetiredBugbotAndSecurity(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	for _, id := range []string{"bugbot", "security"} {
		err := st.PutSettings(ctx, store.Settings{
			Reviewers: []string{"primary", id},
		})
		if !errors.Is(err, store.ErrInvalidReviewer) {
			t.Errorf("PutSettings(%q) error = %v, want it to wrap ErrInvalidReviewer", id, err)
		}
	}
}

// TestSettingsStore_MigrationDropsRetiredReviewers asserts
// 0033_flow_settings_drop_retired_reviewers.sql strips every retired slot
// id from a flow_settings row written before the retirement, so no run
// resolves a slot with no prompt behind it and /flow-settings' "keep
// current" re-writes a list PutSettings accepts.
func TestSettingsStore_MigrationDropsRetiredReviewers(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, `
		INSERT INTO flow_settings (id, reviewers)
		VALUES (TRUE, '["primary","bugbot","principles","security","code-review-low","simple-reviewer","mutation"]')
		ON CONFLICT (id) DO UPDATE SET reviewers = EXCLUDED.reviewers
	`); err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM schema_migrations WHERE filename = '0033_flow_settings_drop_retired_reviewers.sql'`); err != nil {
		t.Fatalf("unrecord migration: %v", err)
	}
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	got, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	want := []string{"primary", "principles", "mutation"}
	if !reflect.DeepEqual(got.Reviewers, want) {
		t.Errorf("reviewers = %v, want %v", got.Reviewers, want)
	}
	if err := st.PutSettings(ctx, got); err != nil {
		t.Errorf("PutSettings(migrated row) = %v, want it accepted", err)
	}
}

// TestSettingsStore_MigrationDropsModelColumns asserts
// 0034_flow_settings_drop_model_columns.sql leaves flow_settings holding
// id, reviewers and updated_at alone, and that a row written before it -- both model
// columns populated -- keeps its reviewer list through the drop.
func TestSettingsStore_MigrationDropsModelColumns(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, `
		ALTER TABLE flow_settings
			ADD COLUMN default_model TEXT NOT NULL DEFAULT 'opus',
			ADD COLUMN self_review_model TEXT NOT NULL DEFAULT ''
	`); err != nil {
		t.Fatalf("restore pre-0034 columns: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO flow_settings (id, default_model, self_review_model, reviewers)
		VALUES (TRUE, 'sonnet', 'haiku', '["primary","failure-modes"]')
		ON CONFLICT (id) DO UPDATE SET reviewers = EXCLUDED.reviewers
	`); err != nil {
		t.Fatalf("seed pre-0034 row: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM schema_migrations WHERE filename = '0034_flow_settings_drop_model_columns.sql'`); err != nil {
		t.Fatalf("unrecord migration: %v", err)
	}
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	rows, err := pool.Query(ctx, `
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'flow_settings'
		ORDER BY column_name
	`)
	if err != nil {
		t.Fatalf("query columns: %v", err)
	}
	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		cols = append(cols, c)
	}
	rows.Close()
	if want := []string{"id", "reviewers", "updated_at"}; !reflect.DeepEqual(cols, want) {
		t.Errorf("flow_settings columns = %v, want %v", cols, want)
	}

	got, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if want := []string{"primary", "failure-modes"}; !reflect.DeepEqual(got.Reviewers, want) {
		t.Errorf("reviewers = %v, want %v", got.Reviewers, want)
	}
}
