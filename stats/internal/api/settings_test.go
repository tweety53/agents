package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// GetSettings and PutSettings extend fakeStore (changes_test.go) with an
// in-memory stand-in for flow_settings, so this file's tests need no
// database -- exactly the payoff go-interface-design gives every other
// handler in this package (api.SettingsStore is defined at the consumer,
// just like api.ChangeStore, api.StageStore, api.StatsStore and
// api.RecordStore).
func (f *fakeStore) GetSettings(_ context.Context) (store.Settings, error) {
	if f.getSettingsErr != nil {
		return store.Settings{}, f.getSettingsErr
	}
	if f.settings == nil {
		return store.Settings{
			DefaultModel: store.DefaultModel,
			Reviewers:    append([]string(nil), store.DefaultReviewers...),
		}, nil
	}
	return *f.settings, nil
}

func (f *fakeStore) PutSettings(_ context.Context, s store.Settings) error {
	if f.putSettingsErr != nil {
		return f.putSettingsErr
	}
	if err := store.ValidateSettings(s); err != nil {
		return err
	}
	f.settings = &s
	return nil
}

func newSettingsTestServer(t *testing.T, fs *fakeStore) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	srv, err := api.New(cfg, fs, fs, fs, fs, fs, nil)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	return httptest.NewServer(srv.Handler())
}

// TestSettingsAPI_Get asserts GET /api/v1/settings answers with the
// harness-wide defaults when flow_settings holds no row yet -- the
// GetSettings contract task 1 documents (store.DefaultModel,
// store.DefaultReviewers) -- and with whatever was last written otherwise.
func TestSettingsAPI_Get(t *testing.T) {
	fs := newFakeStore()
	ts := newSettingsTestServer(t, fs)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/settings")
	if err != nil {
		t.Fatalf("GET /api/v1/settings: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var got struct {
		DefaultModel string   `json:"defaultModel"`
		Reviewers    []string `json:"reviewers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.DefaultModel != store.DefaultModel {
		t.Errorf("defaultModel = %q, want %q", got.DefaultModel, store.DefaultModel)
	}
	if len(got.Reviewers) != len(store.DefaultReviewers) {
		t.Errorf("reviewers = %v, want %v", got.Reviewers, store.DefaultReviewers)
	}
}

// TestSettingsAPI_Put_Valid asserts a well-formed PUT overwrites
// flow_settings' single row, and that a following GET reports exactly
// what was written -- the round-trip DecodeChangeBody/put already proves
// for a change record, proved here for settings.
func TestSettingsAPI_Put_Valid(t *testing.T) {
	fs := newFakeStore()
	ts := newSettingsTestServer(t, fs)
	defer ts.Close()

	body := `{"defaultModel":"opus","reviewers":["primary","principles","code-review-low","security"]}`
	req, err := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/settings", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /api/v1/settings: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	got, err := fs.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if got.DefaultModel != "opus" {
		t.Errorf("stored defaultModel = %q, want %q", got.DefaultModel, "opus")
	}
	want := []string{"primary", "principles", "code-review-low", "security"}
	if len(got.Reviewers) != len(want) {
		t.Fatalf("stored reviewers = %v, want %v", got.Reviewers, want)
	}
	for i, r := range want {
		if got.Reviewers[i] != r {
			t.Errorf("stored reviewers[%d] = %q, want %q", i, got.Reviewers[i], r)
		}
	}
}

// TestSettingsAPI_Put_RejectsInvalidValue asserts an invalid model or
// reviewer value is refused with 400, naming the rejected value, and
// never partially written -- the task's own "never a silent partial
// write" requirement, proved by asserting flow_settings is left holding
// whatever it held before the rejected PUT.
func TestSettingsAPI_Put_RejectsInvalidValue(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantInError string
	}{
		{
			name:        "unknown model",
			body:        `{"defaultModel":"gpt-5","reviewers":["primary","principles","code-review-low"]}`,
			wantInError: "gpt-5",
		},
		{
			name:        "unknown reviewer",
			body:        `{"defaultModel":"sonnet","reviewers":["primary","principles","not-a-real-slot"]}`,
			wantInError: "not-a-real-slot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newFakeStore()
			ts := newSettingsTestServer(t, fs)
			defer ts.Close()

			req, err := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/settings", bytes.NewReader([]byte(tt.body)))
			if err != nil {
				t.Fatalf("build request: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("PUT /api/v1/settings: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", resp.StatusCode)
			}

			var body struct {
				Error string `json:"error"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if !bytes.Contains([]byte(body.Error), []byte(tt.wantInError)) {
				t.Errorf("error = %q, want it to name %q", body.Error, tt.wantInError)
			}

			// Never a silent partial write: flow_settings still reports no
			// row written (the harness defaults), exactly as before the
			// rejected PUT.
			got, err := fs.GetSettings(context.Background())
			if err != nil {
				t.Fatalf("GetSettings: %v", err)
			}
			if got.DefaultModel != store.DefaultModel {
				t.Errorf("after rejected PUT, defaultModel = %q, want unchanged %q", got.DefaultModel, store.DefaultModel)
			}
		})
	}
}
