package web_test

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
	"github.com/tweety53/agents/stats/internal/web"
)

// TestEmbeddedAssetsArePresent asserts the embedded build is genuinely the
// real, compiled React app -- not merely "some bytes exist" (this task's
// own testing standard names that exact vacuous shape as a defect to
// avoid). It cross-checks index.html's own <script> reference against the
// embedded filesystem: a marker that could only be true of a real Vite
// build, which hashes its output filenames from content and rewrites
// index.html to reference the hash it actually produced. A stub dist/
// containing an unrelated index.html plus an unrelated file would satisfy
// "some bytes exist" but fail this cross-check, because the two would not
// name each other.
func TestEmbeddedAssetsArePresent(t *testing.T) {
	fsys, err := web.FS()
	if err != nil {
		t.Fatalf("web.FS: %v", err)
	}

	indexHTML, err := readFileString(fsys, "index.html")
	if err != nil {
		t.Fatalf("reading index.html: %v", err)
	}

	// The build's own React root mount point, per stats/web/index.html --
	// confirms this is the app's index.html, not an arbitrary HTML file
	// that happens to be named that.
	if !strings.Contains(indexHTML, `<div id="root">`) {
		t.Fatalf("index.html does not contain the app's #root mount point:\n%s", indexHTML)
	}

	// Find the built JS entrypoint index.html actually references, and
	// confirm that exact file exists in the embedded tree -- the
	// cross-check a "some bytes exist" assertion would miss entirely,
	// since a mismatched pair (index.html referencing a script that isn't
	// there, or vice versa) trivially satisfies "the directory is
	// non-empty".
	scriptPath := extractModuleScriptSrc(t, indexHTML)
	assetName := strings.TrimPrefix(scriptPath, "/")
	if _, err := readFileString(fsys, assetName); err != nil {
		t.Fatalf("index.html references script %q, but it is not in the embedded build: %v", scriptPath, err)
	}
	if !strings.Contains(assetName, "assets/") {
		t.Fatalf("script src %q is not one of Vite's hashed build outputs under assets/ -- looks unbuilt", scriptPath)
	}
}

// TestUnknownPathServesIndexNotFound: a path the build did not produce as
// a static file (a client-side route the SPA's own router owns, e.g.
// "/changes/kan-16-myflow-stats-app") is answered with the SPA's
// index.html, 200 OK -- not a 404 -- so the client-side router can take
// over, per design.md's "unknown non-API paths fall through to
// index.html for client-side routing".
func TestUnknownPathServesIndexNotFound(t *testing.T) {
	fsys, err := web.FS()
	if err != nil {
		t.Fatalf("web.FS: %v", err)
	}
	handler, err := web.Handler(fsys)
	if err != nil {
		t.Fatalf("web.Handler: %v", err)
	}

	wantIndex, err := readFileString(fsys, "index.html")
	if err != nil {
		t.Fatalf("reading index.html: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/changes/kan-16-myflow-stats-app", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != wantIndex {
		t.Fatalf("unknown path did not serve index.html verbatim:\ngot:  %q\nwant: %q", got, wantIndex)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}

	// A real static asset, by contrast, must be served as itself -- not
	// silently rewritten to index.html too. Exercising only the unknown
	// path (and never this) is exactly the kind of single-branch coverage
	// a mutation that always returns index.html would still pass.
	scriptPath := extractModuleScriptSrc(t, wantIndex)
	assetReq := httptest.NewRequest(http.MethodGet, scriptPath, nil)
	assetRec := httptest.NewRecorder()
	handler.ServeHTTP(assetRec, assetReq)
	if assetRec.Code != http.StatusOK {
		t.Fatalf("static asset %s: status = %d, want 200", scriptPath, assetRec.Code)
	}
	if assetRec.Body.String() == wantIndex {
		t.Fatalf("static asset %s was served as index.html instead of its own content", scriptPath)
	}
}

// TestApiPathsAreNotSwallowedBySpaFallback exercises the full routing
// split -- internal/api.New wired with WithSPA(web's handler), exactly as
// cmd/myflowd wires it -- in both directions this task's own requirement
// names: a real API route must still be handled as an API route, and an
// API-shaped path that matches no real route must be rejected as an API
// request (JSON, 404, naming the path) rather than "answered" with the
// SPA's index.html. Testing only one direction would leave the other
// unguarded -- exactly the single-branch gap this task's testing standard
// calls out.
func TestApiPathsAreNotSwallowedBySpaFallback(t *testing.T) {
	fsys, err := web.FS()
	if err != nil {
		t.Fatalf("web.FS: %v", err)
	}
	spaHandler, err := web.Handler(fsys)
	if err != nil {
		t.Fatalf("web.Handler: %v", err)
	}

	indexHTML, err := readFileString(fsys, "index.html")
	if err != nil {
		t.Fatalf("reading index.html: %v", err)
	}

	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	srv, err := api.New(cfg, fakeStore{}, fakeStore{}, fakeStore{}, nil, api.WithSPA(spaHandler))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	handler := srv.Handler()

	// Direction 1: a real, registered API route is handled by the API --
	// never by the SPA fallback -- even though it shares the daemon with
	// one. fakeStore.ListChanges returns an empty, successful result, so a
	// 200 with the API's own JSON envelope (never index.html) proves this.
	t.Run("a real API route is handled by the API", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/changes", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json (got the SPA instead?)", ct)
		}
		if rec.Body.String() == indexHTML {
			t.Fatalf("a real API route was answered with the SPA's index.html")
		}
		var body struct {
			Total   int `json:"total"`
			Changes []struct{}
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("response body is not the API's JSON shape: %v (body: %s)", err, rec.Body.String())
		}
	})

	// Direction 2: an API-shaped path that matches no registered route --
	// a typo, a retired endpoint -- is rejected as an API request, never
	// silently handed to the SPA fallback. This is the swallowing this
	// task's non-negotiable requirement names directly.
	t.Run("an unmatched api-shaped path is not swallowed by the SPA fallback", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json (the SPA fallback swallowed this path)", ct)
		}
		if rec.Body.String() == indexHTML {
			t.Fatalf("an unmatched API path was answered with the SPA's index.html")
		}
		var body struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("error body is not JSON: %v (body: %s)", err, rec.Body.String())
		}
		if !strings.Contains(body.Error, "/api/v1/does-not-exist") {
			t.Fatalf("error message %q does not name the unmatched path", body.Error)
		}
	})

	// A genuinely unknown, non-API path is the SPA's job, still -- proven
	// once already by TestUnknownPathServesIndexNotFound against the SPA
	// handler alone, and repeated here against the *combined* mux so a
	// regression that broke only the combined wiring (e.g. registering
	// "/api/" before "/", relying on order instead of specificity) would
	// still be caught.
	t.Run("a non-api path still reaches the SPA", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/changes/kan-16-myflow-stats-app", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if rec.Body.String() != indexHTML {
			t.Fatalf("non-API path was not served the SPA's index.html")
		}
	})

	// F1 (post-commit review, task 12): Go's enhanced http.ServeMux (1.22+)
	// resolves a pattern against r.URL.EscapedPath(), which keeps a
	// percent-encoded slash as three literal characters rather than the
	// "/" it decodes to in r.URL.Path. "/api%2Fv1/changes" therefore has no
	// literal "/" where "/api/" needs one, matches neither the exact route
	// nor the "/api/" catch-all, and -- absent the guard this subtest
	// exercises -- falls through to the SPA and is answered with
	// index.html: HTML with no DaemonHeader, which internal/client reads
	// as "store unreachable" and silently falls back on, on a perfectly
	// healthy daemon. internal/client.go's own url.PathEscape(project) /
	// url.PathEscape(name) produces exactly this shape whenever a project
	// key or change name contains a slash, so this is not only a crafted
	// probe.
	//
	// Every case here must land on one of two outcomes -- reach the real
	// API route, or be cleanly refused as a 400 -- and never on the SPA
	// (200, text/html, index.html's own body). That is the dividing line
	// these subtests assert on, not a specific status code, since a
	// legitimate-shaped request (the real API route, or an encoded slash
	// inside a change name) and a garbled one (a double-encoded slash
	// splitting "api" itself) both satisfy it in different ways.
	t.Run("a percent-encoded slash is never swallowed by the SPA fallback", func(t *testing.T) {
		cases := []struct {
			name string
			path string
		}{
			{"encoded slash splits the api prefix itself (uppercase)", "/api%2Fv1/changes"},
			{"encoded slash splits the api prefix itself (lowercase)", "/api%2fv1/changes"},
			{"double-encoded slash splits the api prefix itself", "/api%252Fv1/changes"},
			{"encoded slash in a legitimate position -- a change name", "/api/v1/changes/proj/name%2Fwith%2Fslash"},
			{"encoded slash reaches the stage-mark route class", "/api/v1/stages%2Fbegin"},
			{"encoded slash reaches the stats route class", "/api/v1/stats/state-board%2Fx"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, tc.path, nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				if rec.Body.String() == indexHTML {
					t.Fatalf("path %q was swallowed by the SPA fallback (200 text/html, index.html body)", tc.path)
				}
				ct := rec.Header().Get("Content-Type")
				if strings.HasPrefix(ct, "text/html") {
					t.Fatalf("path %q got Content-Type %q -- looks like the SPA answered it", tc.path, ct)
				}
				if rec.Code == http.StatusOK && ct != "application/json" {
					t.Fatalf("path %q: 200 response has Content-Type %q, not application/json -- not a real API answer", tc.path, ct)
				}
			})
		}
	})
}

func readFileString(fsys fs.FS, name string) (string, error) {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// extractModuleScriptSrc pulls the src of index.html's module <script>
// tag -- Vite's hashed JS entrypoint -- with a narrow, deliberately
// unforgiving parse: this is a fixed, self-authored file (stats/web's own
// build output), not third-party HTML, so a regex is proportionate and a
// full HTML parser would be doing work nothing here needs.
func extractModuleScriptSrc(t *testing.T, html string) string {
	t.Helper()
	const marker = `<script type="module"`
	tagStart := strings.Index(html, marker)
	if tagStart < 0 {
		t.Fatalf("index.html has no module script tag:\n%s", html)
	}
	tagEnd := strings.Index(html[tagStart:], ">")
	if tagEnd < 0 {
		t.Fatalf("index.html's module script tag is never closed:\n%s", html)
	}
	tag := html[tagStart : tagStart+tagEnd]

	const srcMarker = `src="`
	srcStart := strings.Index(tag, srcMarker)
	if srcStart < 0 {
		t.Fatalf("index.html's module script tag has no src attribute: %s", tag)
	}
	rest := tag[srcStart+len(srcMarker):]
	end := strings.Index(rest, `"`)
	if end < 0 {
		t.Fatalf("index.html's module script src has no closing quote: %s", tag)
	}
	return rest[:end]
}

// fakeStore satisfies api.ChangeStore, api.StageStore and api.StatsStore
// all at once with the minimum this test needs: an empty, successful
// change listing, and every other method returning
// store.ErrChangeNotFound / an empty result, none of which this test's
// two directions ever reach. No database is needed -- these tests are
// entirely about routing, not about what any handler does once it owns a
// request, so a fake this thin is the right amount of test double per
// go-test-quality's "fake only what the test needs".
type fakeStore struct{}

func (fakeStore) GetChange(context.Context, string, string) (store.Change, error) {
	return store.Change{}, store.ErrChangeNotFound
}

func (fakeStore) PutChange(context.Context, store.Change) error { return nil }

func (fakeStore) QueryChanges(context.Context, store.Query) ([]store.Change, int, error) {
	return nil, 0, nil
}

func (fakeStore) BeginStage(context.Context, store.BeginStageInput) (store.StageRun, error) {
	return store.StageRun{}, nil
}

func (fakeStore) EndStage(context.Context, int64, time.Time, string) error { return nil }

func (fakeStore) MergeMetrics(context.Context, int64, json.RawMessage) error { return nil }

func (fakeStore) QueryStageRuns(context.Context, store.Query) ([]store.StageRun, int, error) {
	return nil, 0, nil
}

func (fakeStore) LiveStateBoard(context.Context, store.Period, *string) ([]store.LiveStateRow, error) {
	return nil, nil
}

func (fakeStore) CostPerChange(context.Context, store.Period, *string, *string) ([]store.CostPerChangeRow, error) {
	return nil, nil
}

func (fakeStore) StageLeaderboard(context.Context, store.Period, *string, *string) ([]store.StageLeaderboardRow, error) {
	return nil, nil
}

func (fakeStore) TrendOverTime(context.Context, store.Period, *string, *string) ([]store.TrendPoint, error) {
	return nil, nil
}

func (fakeStore) CacheEfficiency(context.Context, store.Period, *string, *string) ([]store.CacheEfficiencyRow, error) {
	return nil, nil
}

func (fakeStore) PanelEconomics(context.Context, store.Period, *string, *string) ([]store.PanelEconomicsRow, error) {
	return nil, nil
}

func (fakeStore) ModelComparison(context.Context, store.Period, *string, *string) ([]store.ModelComparisonRow, error) {
	return nil, nil
}

func (fakeStore) ReworkRate(context.Context, store.Period, *string, *string) ([]store.ReworkRateRow, error) {
	return nil, nil
}

func (fakeStore) CountRunsWithoutModel(context.Context, store.Period, *string) (int, error) {
	return 0, nil
}

func (fakeStore) ListModels(context.Context, store.Period, *string) ([]string, error) {
	return nil, nil
}

// AllRecordedRunsUnmeasured is here purely to keep satisfying
// api.StatsStore -- this file's fakeStore never exercises a stats route
// (unanticipated file, task 5: the interface it implements gained one
// method, and every implementer of it must compile; mechanical
// substitution, no logic change).
func (fakeStore) AllRecordedRunsUnmeasured(context.Context, store.Period, *string) (bool, error) {
	return false, nil
}

// ProjectKeysByDisplayName is here for the same reason
// AllRecordedRunsUnmeasured's own doc comment gives: api.ChangeStore and
// api.StatsStore both gained this method for task 3's project
// display-name resolution, and every implementer must keep compiling --
// this file's fakeStore never exercises project resolution at all.
func (fakeStore) ProjectKeysByDisplayName(context.Context, string) ([]string, error) {
	return nil, nil
}
