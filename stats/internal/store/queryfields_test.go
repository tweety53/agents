package store_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// queryFieldsFixture is the committed shape of
// stats/web/src/testdata/queryFields.json: the server's own query
// allowlists, named field-set by field-set, so the SPA has one committed
// copy of the real vocabulary to import rather than a hand-typed guess.
type queryFieldsFixture struct {
	ChangeFields   []string `json:"changeFields"`
	StageRunFields []string `json:"stageRunFields"`
}

// queryFieldsFixturePath is stats/web/src/testdata/queryFields.json,
// resolved relative to this package -- the SPA-facing counterpart of
// task 24's harvestshape fixture and F2's metricsPatchFixture.json
// (internal/harvest/wireshape_test.go), same discipline: a hand-written
// TypeScript field list is a second, private copy of query.go's allowlist,
// and this is exactly the failure mode task 26 exists to close (the
// component call sites sending "updatedAt"/"startedAt" -- DTO field names
// -- against a server allowlist keyed on real column names).
func queryFieldsFixturePath(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "..", "web", "src", "testdata", "queryFields.json"))
	if err != nil {
		t.Fatalf("resolve fixture path: %v", err)
	}
	return path
}

// TestSPAQueryFieldsFixtureMatchesTheAllowlist guards
// stats/web/src/testdata/queryFields.json against drifting from
// store.AllowedChangeFields() and store.AllowedStageRunFields() -- the
// server's own, single source of truth for which sort/filter field names
// a request may use. A field renamed, added or removed in query.go's
// changeFieldColumns/stageRunFieldColumns maps changes what those two
// functions report, which changes what this test marshals, which no
// longer matches the committed fixture until it is regenerated.
//
// Regenerate the fixture with:
//
//	UPDATE_QUERY_FIELDS_FIXTURE=1 go test ./internal/store/ -run TestSPAQueryFieldsFixtureMatchesTheAllowlist
//
// which writes the current allowlists and always passes; run again without
// the env var to confirm it now matches.
func TestSPAQueryFieldsFixtureMatchesTheAllowlist(t *testing.T) {
	want := queryFieldsFixture{
		ChangeFields:   store.AllowedChangeFields(),
		StageRunFields: store.AllowedStageRunFields(),
	}
	wantJSON, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatalf("marshal query fields: %v", err)
	}
	wantJSON = append(wantJSON, '\n')

	path := queryFieldsFixturePath(t)

	if os.Getenv("UPDATE_QUERY_FIELDS_FIXTURE") != "" {
		if err := os.WriteFile(path, wantJSON, 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		return
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read committed fixture %s: %v (run with UPDATE_QUERY_FIELDS_FIXTURE=1 to create it)", path, err)
	}

	var wantAny, gotAny any
	if err := json.Unmarshal(wantJSON, &wantAny); err != nil {
		t.Fatalf("unmarshal freshly marshalled query fields: %v", err)
	}
	if err := json.Unmarshal(got, &gotAny); err != nil {
		t.Fatalf("unmarshal committed fixture: %v", err)
	}
	if !reflect.DeepEqual(wantAny, gotAny) {
		t.Fatalf(
			"stats/web/src/testdata/queryFields.json has drifted from store.AllowedChangeFields()/AllowedStageRunFields().\n"+
				"committed:\n%s\nwant (from the store's own allowlists):\n%s\n"+
				"regenerate with: UPDATE_QUERY_FIELDS_FIXTURE=1 go test ./internal/store/ -run TestSPAQueryFieldsFixtureMatchesTheAllowlist",
			got, wantJSON,
		)
	}
}
