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
		DefaultModel: "opus",
		Reviewers:    []string{"primary", "principles", "code-review-low", "bugbot"},
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

	first := store.Settings{DefaultModel: "sonnet", Reviewers: []string{"primary"}}
	second := store.Settings{DefaultModel: "haiku", Reviewers: []string{"primary", "security"}}

	if err := st.PutSettings(ctx, first); err != nil {
		t.Fatalf("first PutSettings: %v", err)
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

// TestSettingsStore_RejectsUnknownModel asserts PutSettings refuses a
// default_model value off the harness's fixed model enum, wrapping
// store.ErrInvalidModel and naming the specific bad value in the error --
// the caller needs to know which value was rejected, not just that
// something was.
func TestSettingsStore_RejectsUnknownModel(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	const badModel = "gpt-5"
	err := st.PutSettings(ctx, store.Settings{
		DefaultModel: badModel,
		Reviewers:    []string{"primary"},
	})
	if err == nil {
		t.Fatal("PutSettings: got nil error, want ErrInvalidModel")
	}
	if !errors.Is(err, store.ErrInvalidModel) {
		t.Errorf("PutSettings error = %v, want it to wrap ErrInvalidModel", err)
	}
	if !strings.Contains(err.Error(), badModel) {
		t.Errorf("PutSettings error = %q, want it to name the rejected value %q", err.Error(), badModel)
	}

	// The rejected write must not have landed.
	if _, getErr := st.GetSettings(ctx); getErr != nil {
		t.Fatalf("GetSettings after rejected write: %v", getErr)
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
		DefaultModel: "sonnet",
		Reviewers:    []string{"primary", badReviewer},
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
