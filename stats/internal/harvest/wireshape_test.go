package harvest_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// spaMetricsFixturePatch is the one canonical harvest.MetricsPatch value
// both sides of the wire agree to test against: this package's own type,
// not a JSON literal that merely resembles it. Every field of both
// harvest.Bucket instances (Main and Sidechain) -- including the
// omitempty ones -- is distinct and non-zero, so a reader that sums the
// wrong fields, or reads the wrong bucket, produces a value that cannot
// be mistaken for a coincidentally-correct one, and so that every JSON
// key Bucket can produce actually appears in the committed fixture (see
// TestSPAFixtureBucketsCoverEveryBucketField below). MetricsPatch's own
// top-level omitempty fields (Models, Effort, Speed) are deliberately
// left zero and absent from the fixture -- this value exists to pin the
// token-bucket wire shape the SPA's token readers depend on, not every
// field MetricsPatch has ever grown.
var spaMetricsFixturePatch = harvest.MetricsPatch{
	Tokens: harvest.TokenDelta{
		Main: harvest.Bucket{
			Input: 600000, Output: 300000, CacheCreation: 50000,
			CacheCreation5m: 30000, CacheCreation1h: 15000, CacheCreationUnknown: 5000,
			CacheRead: 950000, Thinking: 10000,
		},
		Sidechain: harvest.Bucket{
			Input: 400000, Output: 100000, CacheCreation: 20000,
			CacheCreation5m: 12000, CacheCreation1h: 6000, CacheCreationUnknown: 2000,
			CacheRead: 300000, Thinking: 5000,
		},
	},
}

// spaMetricsFixturePath is stats/web/src/testdata/metricsPatchFixture.json,
// resolved relative to this package -- the one committed copy of the wire
// shape stats/web/src/metrics.test.ts imports directly, rather than
// hand-writing a JSON literal that is free to drift from what
// internal/harvest actually emits.
func spaMetricsFixturePath(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "..", "web", "src", "testdata", "metricsPatchFixture.json"))
	if err != nil {
		t.Fatalf("resolve fixture path: %v", err)
	}
	return path
}

// TestSPAMetricsFixtureMatchesTheWireShape is F2's seam test, one layer
// up from harvestshape_test.go's (internal/store's fixture is Go-native;
// this one is the SPA's, and TypeScript has no way to import a Go type
// directly). It guards stats/web/src/testdata/metricsPatchFixture.json
// against drifting from harvest.MetricsPatch, the same discipline task 24
// established for internal/store: a hand-written TypeScript literal
// standing in for the harvester's real shape is a second, private copy of
// the contract, and this is exactly the failure mode that let
// metrics.test.ts's old flat `{main: 90, sidechain: 60}` fixture pass
// against readers that could never work against a real bag.
//
// What this test actually catches, by comparing marshalled *values*: a
// renamed json tag, a field moved to a different nesting level, and a new
// field added without omitempty (it marshals to something on the fresh
// side and nothing on the stale committed side, so the byte comparison
// below fails). What it does NOT catch on its own: a new field added
// *with* omitempty that spaMetricsFixturePatch leaves at its zero value --
// that marshals to nothing on both sides, so this comparison stays
// byte-identical even though the wire shape has genuinely grown. Closing
// that blind spot is TestSPAFixtureBucketsCoverEveryBucketField below,
// which compares the *field set* harvest.Bucket declares (independent of
// omitempty and independent of any value) against the fixture's actual
// keys, so an omitempty field left at zero is a missing fixture key and
// fails rather than passing silently.
//
// Regenerate the fixture with:
//
//	UPDATE_METRICS_FIXTURE=1 go test ./internal/harvest/ -run TestSPAMetricsFixtureMatchesTheWireShape
//
// which writes the current marshalled shape of spaMetricsFixturePatch and
// always passes; run again without the env var to confirm it now matches.
func TestSPAMetricsFixtureMatchesTheWireShape(t *testing.T) {
	want, err := json.MarshalIndent(spaMetricsFixturePatch, "", "  ")
	if err != nil {
		t.Fatalf("marshal MetricsPatch: %v", err)
	}
	want = append(want, '\n')

	path := spaMetricsFixturePath(t)

	if os.Getenv("UPDATE_METRICS_FIXTURE") != "" {
		if err := os.WriteFile(path, want, 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		return
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read committed fixture %s: %v (run with UPDATE_METRICS_FIXTURE=1 to create it)", path, err)
	}

	var wantAny, gotAny any
	if err := json.Unmarshal(want, &wantAny); err != nil {
		t.Fatalf("unmarshal freshly marshalled patch: %v", err)
	}
	if err := json.Unmarshal(got, &gotAny); err != nil {
		t.Fatalf("unmarshal committed fixture: %v", err)
	}
	if !reflect.DeepEqual(wantAny, gotAny) {
		t.Fatalf(
			"stats/web/src/testdata/metricsPatchFixture.json has drifted from harvest.MetricsPatch's real wire shape.\n"+
				"committed:\n%s\nwant (from harvest.MetricsPatch):\n%s\n"+
				"regenerate with: UPDATE_METRICS_FIXTURE=1 go test ./internal/harvest/ -run TestSPAMetricsFixtureMatchesTheWireShape",
			got, want,
		)
	}
}

// jsonFieldNames returns the full set of JSON key names t's exported
// fields declare via their `json:"..."` tag -- with the omitempty option
// (or any other comma-separated option) stripped, so a field's presence
// here does not depend on whether its zero value happens to marshal to
// nothing. This is deliberately not a marshal-and-inspect approach: it
// reads the struct tags directly, so it reports every key the type CAN
// produce, not just the keys today's fixture value happens to produce.
func jsonFieldNames(t reflect.Type) map[string]bool {
	names := make(map[string]bool, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" {
			continue
		}
		names[name] = true
	}
	return names
}

// sortedKeys is a small formatting helper: a diagnostic listing keys in a
// stable, readable order rather than Go's randomised map iteration order.
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TestSPAFixtureBucketsCoverEveryBucketField is the structural half of
// TestSPAMetricsFixtureMatchesTheWireShape's guard, closing the blind
// spot that test's own doc comment names: a new harvest.Bucket field
// added with `json:",omitempty"` and left at its zero value in
// spaMetricsFixturePatch marshals to nothing on both the freshly
// marshalled side and the committed fixture, so the value-level byte
// comparison stays green even though the wire shape has genuinely grown.
//
// This test does not compare values at all. It compares the *set of JSON
// keys* harvest.Bucket's struct tags declare -- independent of omitempty,
// independent of whatever spaMetricsFixturePatch's fields happen to hold
// -- against the set of keys actually present under "tokens.main" and
// "tokens.sidechain" in the committed fixture. An omitempty field left at
// zero is therefore a missing fixture key, and this test fails on it
// even when the byte comparison above cannot see it.
//
// Before task 25's fix round added CacheCreation5m, CacheCreation1h and
// CacheCreationUnknown to harvest.Bucket, spaMetricsFixturePatch's Main
// and Sidechain buckets left those three (all omitempty) fields at their
// zero value, and the committed fixture correspondingly lacked all three
// keys -- exactly the shape-drift-that-passes case this test exists to
// catch. Running this test against that committed fixture failed with:
//
//	tokens.main fixture keys = [cache_creation cache_read input output thinking],
//	want exactly [cache_creation cache_creation_1h cache_creation_5m cache_creation_unknown cache_read input output thinking]
//	(missing [cache_creation_1h cache_creation_5m cache_creation_unknown])
//
// which is precisely the class of drift the byte-level test above cannot
// see. Regenerating the fixture with spaMetricsFixturePatch's Main and
// Sidechain buckets carrying non-zero values for every field (this
// package's own discipline, see spaMetricsFixturePatch's doc comment)
// closes it.
func TestSPAFixtureBucketsCoverEveryBucketField(t *testing.T) {
	path := spaMetricsFixturePath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read committed fixture %s: %v", path, err)
	}

	var decoded struct {
		Tokens struct {
			Main      map[string]json.RawMessage `json:"main"`
			Sidechain map[string]json.RawMessage `json:"sidechain"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal committed fixture %s: %v", path, err)
	}

	want := jsonFieldNames(reflect.TypeOf(harvest.Bucket{}))

	for _, bucket := range []struct {
		name string
		got  map[string]json.RawMessage
	}{
		{"main", decoded.Tokens.Main},
		{"sidechain", decoded.Tokens.Sidechain},
	} {
		got := make(map[string]bool, len(bucket.got))
		for k := range bucket.got {
			got[k] = true
		}
		if !reflect.DeepEqual(want, got) {
			var missing, extra []string
			for k := range want {
				if !got[k] {
					missing = append(missing, k)
				}
			}
			for k := range got {
				if !want[k] {
					extra = append(extra, k)
				}
			}
			sort.Strings(missing)
			sort.Strings(extra)
			t.Errorf(
				"tokens.%s fixture keys = %v, want exactly %v (missing %v, extra %v) -- "+
					"a harvest.Bucket field with no non-zero value in spaMetricsFixturePatch "+
					"marshals to nothing and so is invisible to the byte-level comparison in "+
					"TestSPAMetricsFixtureMatchesTheWireShape; give it a non-zero value there "+
					"and regenerate the fixture with UPDATE_METRICS_FIXTURE=1",
				bucket.name, sortedKeys(got), sortedKeys(want), fmt.Sprint(missing), fmt.Sprint(extra),
			)
		}
	}
}
