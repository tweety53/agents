package api_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// RecordLearningTopic and ListLearningTopics extend fakeStore with an
// in-memory stand-in for flow_learning_topics, enforcing the same slug
// and cap rules the real store does.
func (f *fakeStore) RecordLearningTopic(_ context.Context, in store.LearningTopic) error {
	if in.Topic == "" || strings.ToLower(in.Topic) != in.Topic || strings.Contains(in.Topic, " ") {
		return store.ErrInvalidLearningTopic
	}
	n := 0
	for _, t := range f.learningTopics {
		if t.Topic == in.Topic {
			n++
		}
	}
	if n >= store.MaxLearningTopicShows {
		return store.ErrLearningTopicExhausted
	}
	f.learningTopics = append(f.learningTopics, in)
	return nil
}

func (f *fakeStore) ListLearningTopics(_ context.Context) ([]store.LearningTopicCount, error) {
	out := []store.LearningTopicCount{}
	idx := map[string]int{}
	for _, t := range f.learningTopics {
		i, ok := idx[t.Topic]
		if !ok {
			i = len(out)
			idx[t.Topic] = i
			out = append(out, store.LearningTopicCount{Topic: t.Topic})
		}
		out[i].Count++
		out[i].LastNote = t.Note
	}
	return out, nil
}

func postLearning(t *testing.T, url, body string) int {
	t.Helper()
	resp, err := http.Post(url+"/api/v1/learning", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("POST /api/v1/learning: %v", err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

// TestLearningHandler_AddAndList asserts a POST records a topic and a GET
// lists it with its count and newest note (topics-in-flowd).
func TestLearningHandler_AddAndList(t *testing.T) {
	fs := newFakeStore()
	ts := newSettingsTestServer(t, fs)
	defer ts.Close()

	if got := postLearning(t, ts.URL, `{"topic":"cqrs","note":"n","change":"c","projectKey":"p"}`); got != http.StatusCreated {
		t.Fatalf("POST status = %d, want 201", got)
	}

	resp, err := http.Get(ts.URL + "/api/v1/learning")
	if err != nil {
		t.Fatalf("GET /api/v1/learning: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if want := `[{"topic":"cqrs","count":1,"lastNote":"n"}]`; strings.TrimSpace(string(body)) != want {
		t.Errorf("GET body = %s, want %s", body, want)
	}
}

// TestLearningHandler_ExhaustedIs409 asserts a fourth record of one topic
// answers 409 (cap-in-store).
func TestLearningHandler_ExhaustedIs409(t *testing.T) {
	fs := newFakeStore()
	ts := newSettingsTestServer(t, fs)
	defer ts.Close()

	body := `{"topic":"cqrs","note":"n","change":"c","projectKey":"p"}`
	for i := 0; i < store.MaxLearningTopicShows; i++ {
		postLearning(t, ts.URL, body)
	}
	if got := postLearning(t, ts.URL, body); got != http.StatusConflict {
		t.Errorf("fourth POST status = %d, want 409", got)
	}
}

// TestLearningHandler_BadTopicIs400 asserts a malformed topic answers 400.
func TestLearningHandler_BadTopicIs400(t *testing.T) {
	fs := newFakeStore()
	ts := newSettingsTestServer(t, fs)
	defer ts.Close()

	if got := postLearning(t, ts.URL, `{"topic":"Go Generics","note":"n","change":"c","projectKey":"p"}`); got != http.StatusBadRequest {
		t.Errorf("POST status = %d, want 400", got)
	}
}
