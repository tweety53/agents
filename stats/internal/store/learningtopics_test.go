package store_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

func recordTopic(t *testing.T, st *store.Store, topic, note string) error {
	t.Helper()
	return st.RecordLearningTopic(context.Background(), store.LearningTopic{
		Topic: topic, Note: note, Change: "c", ProjectKey: "p",
	})
}

// TestLearningTopics_ListCounts asserts ListLearningTopics returns each
// topic once, ordered by topic, with its count and newest note.
func TestLearningTopics_ListCounts(t *testing.T) {
	st := newTestStore(t)
	for _, in := range [][2]string{{"go-generics", "first"}, {"go-generics", "second"}, {"cqrs", "only"}} {
		if err := recordTopic(t, st, in[0], in[1]); err != nil {
			t.Fatalf("RecordLearningTopic(%q): %v", in[0], err)
		}
	}

	got, err := st.ListLearningTopics(context.Background())
	if err != nil {
		t.Fatalf("ListLearningTopics: %v", err)
	}
	want := []store.LearningTopicCount{
		{Topic: "cqrs", Count: 1, LastNote: "only"},
		{Topic: "go-generics", Count: 2, LastNote: "second"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListLearningTopics = %+v, want %+v", got, want)
	}
}

// TestLearningTopics_CapAtThree asserts a fourth record of one topic is
// refused with ErrLearningTopicExhausted and writes nothing (cap-in-store).
func TestLearningTopics_CapAtThree(t *testing.T) {
	st := newTestStore(t)
	for i := 0; i < store.MaxLearningTopicShows; i++ {
		if err := recordTopic(t, st, "cqrs", "n"); err != nil {
			t.Fatalf("record %d: %v", i+1, err)
		}
	}

	err := recordTopic(t, st, "cqrs", "n")
	if !errors.Is(err, store.ErrLearningTopicExhausted) {
		t.Fatalf("fourth record err = %v, want ErrLearningTopicExhausted", err)
	}
	got, err := st.ListLearningTopics(context.Background())
	if err != nil {
		t.Fatalf("ListLearningTopics: %v", err)
	}
	if len(got) != 1 || got[0].Count != store.MaxLearningTopicShows {
		t.Errorf("ListLearningTopics = %+v, want cqrs at count %d", got, store.MaxLearningTopicShows)
	}
}

// TestLearningTopics_RejectsBadSlug asserts a topic that is not a
// lowercase kebab slug of at most 64 bytes, or an empty note, is refused
// with ErrInvalidLearningTopic (topic-slug).
func TestLearningTopics_RejectsBadSlug(t *testing.T) {
	st := newTestStore(t)
	for _, topic := range []string{"Go Generics", "", strings.Repeat("a", 65)} {
		if err := recordTopic(t, st, topic, "n"); !errors.Is(err, store.ErrInvalidLearningTopic) {
			t.Errorf("RecordLearningTopic(%q) err = %v, want ErrInvalidLearningTopic", topic, err)
		}
	}
	if err := recordTopic(t, st, "cqrs", ""); !errors.Is(err, store.ErrInvalidLearningTopic) {
		t.Errorf("empty note err = %v, want ErrInvalidLearningTopic", err)
	}
}
