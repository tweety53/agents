package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

func TestPutChangeWritesRepoSet(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-repo-set")
	c.Repos = []store.Repo{
		{RepoRoot: "/Users/tweety53/Projects/agents", MergeBase: ptr("abc1230000000000000000000000000000000000")},
		{RepoRoot: "/Users/tweety53/Projects/other-repo", MergeBase: ptr("def4560000000000000000000000000000000000")},
	}
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	got, err := st.ListChangeRepos(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("ListChangeRepos: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListChangeRepos returned %d repos, want 2: %+v", len(got), got)
	}

	byRoot := make(map[string]store.Repo, len(got))
	for _, r := range got {
		byRoot[r.RepoRoot] = r
	}

	agents, ok := byRoot["/Users/tweety53/Projects/agents"]
	if !ok || agents.MergeBase == nil || *agents.MergeBase != "abc1230000000000000000000000000000000000" {
		t.Errorf("agents repo = %+v, want merge base %q", agents, "abc1230000000000000000000000000000000000")
	}
	other, ok := byRoot["/Users/tweety53/Projects/other-repo"]
	if !ok || other.MergeBase == nil || *other.MergeBase != "def4560000000000000000000000000000000000" {
		t.Errorf("other-repo = %+v, want merge base %q", other, "def4560000000000000000000000000000000000")
	}
}

func TestTwoRepoChangeIsOneRow(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-one-row")
	c.Repos = []store.Repo{
		{RepoRoot: "/repo-a"},
		{RepoRoot: "/repo-b"},
	}
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	got, err := st.ListChanges(ctx, c.ProjectKey)
	if err != nil {
		t.Fatalf("ListChanges: %v", err)
	}

	count := 0
	for _, ch := range got {
		if ch.Name == c.Name {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("ListChanges found %d rows named %q, want exactly 1 (a two-repository change must still be one row)", count, c.Name)
	}
}

func TestPutChangeReplacesRepoSetWholesale(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-replace-repos")
	c.Repos = []store.Repo{
		{RepoRoot: "/repo-a"},
		{RepoRoot: "/repo-b"},
	}
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("first PutChange: %v", err)
	}

	second := c
	second.State = store.StateInProgress
	second.UpdatedAt = c.UpdatedAt.Add(time.Minute)
	second.Repos = []store.Repo{
		{RepoRoot: "/repo-a"},
	}
	if err := st.PutChange(ctx, second); err != nil {
		t.Fatalf("second PutChange: %v", err)
	}

	got, err := st.ListChangeRepos(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("ListChangeRepos: %v", err)
	}
	if len(got) != 1 || got[0].RepoRoot != "/repo-a" {
		t.Fatalf("ListChangeRepos after wholesale replace = %+v, want exactly [/repo-a] (a repository omitted from the write must be removed from the set)", got)
	}
}

func TestNullMergeBaseRoundTrips(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-null-merge-base")
	c.Repos = []store.Repo{
		{RepoRoot: "/repo-no-merge-base", MergeBase: nil},
	}
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	got, err := st.ListChangeRepos(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("ListChangeRepos: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListChangeRepos returned %d repos, want 1", len(got))
	}
	if got[0].MergeBase != nil {
		t.Errorf("MergeBase = %v, want nil (a NULL merge base means none recorded, never a value to infer)", *got[0].MergeBase)
	}
}

func TestListChangeReposOrdersStably(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-stable-order")
	c.Repos = []store.Repo{
		{RepoRoot: "/repo-z"},
		{RepoRoot: "/repo-a"},
		{RepoRoot: "/repo-m"},
	}
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	for i := range 3 {
		got, err := st.ListChangeRepos(ctx, c.ProjectKey, c.Name)
		if err != nil {
			t.Fatalf("ListChangeRepos (call %d): %v", i, err)
		}
		if len(got) != 3 {
			t.Fatalf("ListChangeRepos (call %d) returned %d repos, want 3", i, len(got))
		}
		want := []string{"/repo-a", "/repo-m", "/repo-z"}
		for j, r := range got {
			if r.RepoRoot != want[j] {
				t.Fatalf("ListChangeRepos (call %d)[%d] = %q, want %q", i, j, r.RepoRoot, want[j])
			}
		}
	}
}

// TestRepoSetIsTransactionalWithTheChange proves the repository write
// shares the change row's own transaction, by failing *after* the change
// row's upsert has already run against that transaction (a monotonic
// refusal, by contrast, returns before replaceChangeRepos is ever called,
// so it cannot distinguish a shared transaction from two independently
// committed ones -- verified by moving replaceChangeRepos outside
// PutChange's transaction and confirming this test, unlike a
// monotonic-refusal-based one, fails against that regression).
//
// A duplicate RepoRoot in the payload is the failure: it is rejected by
// replaceChangeRepos only after PutChange's own row upsert has already run
// on tx (id is already known), so the assertions below hold only if that
// upsert's effects are rolled back together with the repo write's.
func TestRepoSetIsTransactionalWithTheChange(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-transactional-repos")
	c.Repos = []store.Repo{
		{RepoRoot: "/repo-original"},
	}
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("first PutChange: %v", err)
	}

	failing := c
	failing.State = store.StateInProgress
	failing.UpdatedAt = c.UpdatedAt.Add(time.Minute)
	failing.Repos = []store.Repo{
		{RepoRoot: "/repo-duplicate"},
		{RepoRoot: "/repo-duplicate"},
	}

	err := st.PutChange(ctx, failing)
	if !errors.Is(err, store.ErrDuplicateRepoRoot) {
		t.Fatalf("PutChange(duplicate repo root) error = %v, want errors.Is(_, store.ErrDuplicateRepoRoot)", err)
	}

	gotChange, err := st.GetChange(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if gotChange.State != store.StateStarted {
		t.Fatalf("stored state = %q after a failed repo write, want %q unchanged (the change row's own upsert must roll back with the repo write, not commit independently of it)", gotChange.State, store.StateStarted)
	}

	got, err := st.ListChangeRepos(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("ListChangeRepos: %v", err)
	}
	if len(got) != 1 || got[0].RepoRoot != "/repo-original" {
		t.Fatalf("ListChangeRepos after a failed write = %+v, want exactly [/repo-original] unchanged", got)
	}
}

// TestPutChangeRejectsDuplicateRepoRoot asserts that a payload naming the
// same RepoRoot twice is refused with a typed ErrDuplicateRepoRoot rather
// than surfacing change_repos' composite-primary-key violation raw.
func TestPutChangeRejectsDuplicateRepoRoot(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-duplicate-repo-root")
	c.Repos = []store.Repo{
		{RepoRoot: "/repo-x"},
		{RepoRoot: "/repo-x"},
	}

	err := st.PutChange(ctx, c)
	if !errors.Is(err, store.ErrDuplicateRepoRoot) {
		t.Fatalf("PutChange(duplicate repo root) error = %v, want errors.Is(_, store.ErrDuplicateRepoRoot)", err)
	}

	if _, getErr := st.GetChange(ctx, c.ProjectKey, c.Name); !errors.Is(getErr, store.ErrChangeNotFound) {
		t.Fatalf("GetChange after a rejected write = %v, want errors.Is(_, store.ErrChangeNotFound) (nothing should have been persisted)", getErr)
	}
}

// TestGetChangeIncludesRepoSet asserts GetChange populates Repos from
// change_repos exactly as PutChange wrote it -- the read side of the same
// whole-object rule the write side follows.
func TestGetChangeIncludesRepoSet(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	c := baseChange("agents", "kan-16-get-change-repos")
	c.Repos = []store.Repo{
		{RepoRoot: "/repo-a", MergeBase: ptr("aaa1110000000000000000000000000000000000")},
		{RepoRoot: "/repo-b"},
	}
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	got, err := st.GetChange(ctx, c.ProjectKey, c.Name)
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if len(got.Repos) != 2 {
		t.Fatalf("GetChange(...).Repos = %+v, want 2 repos", got.Repos)
	}
}

// TestListChangesIncludesRepoSetForEveryChange asserts ListChanges
// populates every returned change's Repos field, not just the first or the
// one queried alone -- the case that would expose an accidental
// single-change assumption in the batched read.
func TestListChangesIncludesRepoSetForEveryChange(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	a := baseChange("proj-list-repos", "change-a")
	a.Repos = []store.Repo{{RepoRoot: "/repo-a"}}
	b := baseChange("proj-list-repos", "change-b")
	b.Repos = []store.Repo{{RepoRoot: "/repo-b1"}, {RepoRoot: "/repo-b2"}}

	for _, c := range []store.Change{a, b} {
		if err := st.PutChange(ctx, c); err != nil {
			t.Fatalf("PutChange(%s): %v", c.Name, err)
		}
	}

	got, err := st.ListChanges(ctx, "proj-list-repos")
	if err != nil {
		t.Fatalf("ListChanges: %v", err)
	}

	byName := make(map[string][]store.Repo, len(got))
	for _, c := range got {
		byName[c.Name] = c.Repos
	}
	if len(byName["change-a"]) != 1 {
		t.Errorf("change-a Repos = %+v, want 1 repo", byName["change-a"])
	}
	if len(byName["change-b"]) != 2 {
		t.Errorf("change-b Repos = %+v, want 2 repos", byName["change-b"])
	}
}
