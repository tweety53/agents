package api_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestProjectsRoute pins GET /api/v1/projects: every registered project
// key, in the store's key order, so the state board can name a project
// whose changes are all finished before those changes load.
func TestProjectsRoute(t *testing.T) {
	fs := newFakeStore()
	fs.projectRoots = []store.ProjectRoot{
		{ProjectKey: "alpha", MainCheckoutPath: "/a"},
		{ProjectKey: "beta", MainCheckoutPath: "/b"},
	}
	ts := newTestServer(t, fs)

	resp, err := http.Get(ts.URL + "/api/v1/projects")
	if err != nil {
		t.Fatalf("GET projects: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET projects = %d, want 200", resp.StatusCode)
	}
	var body struct {
		Projects []string `json:"projects"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if want := []string{"alpha", "beta"}; !reflect.DeepEqual(body.Projects, want) {
		t.Errorf("projects = %v, want %v", body.Projects, want)
	}
}
