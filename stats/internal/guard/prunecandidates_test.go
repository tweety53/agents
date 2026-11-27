package guard

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// pcServer is a flowd answering prune-candidates' four reads from fixtures:
// guard-health rows, one project, that project's incidents, and the skill
// loads. It records the guard-health query so the window can be checked.
func pcServer(t *testing.T, health, incidents, loads string, query *string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/stats/guard-health", func(w http.ResponseWriter, r *http.Request) {
		*query = r.URL.RawQuery
		fmt.Fprintf(w, `{"view":"guard-health","rows":%s}`, health)
	})
	mux.HandleFunc("GET /api/v1/projects", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"projects":["p-1"]}`)
	})
	mux.HandleFunc("GET /api/v1/incidents/p-1", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, incidents)
	})
	mux.HandleFunc("GET /api/v1/skill-loads", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"rows":%s}`, loads)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func pcEnv(addr string) Env { return pcEnvAt(addr, "") }

// pcEnvAt is pcEnv with FLOW_GUARD_REPO_ROOT, the checkout whose skills/ are listed, set to root.
func pcEnvAt(addr, root string) Env {
	return Env{Getenv: func(k string) string {
		switch k {
		case "FLOW_ADDR":
			return addr
		case "FLOW_GUARD_REPO_ROOT":
			return root
		}
		return ""
	}, Dir: "/"}
}

func TestPruneCandidates(t *testing.T) {
	now := time.Now().UTC()
	recent := now.Add(-48 * time.Hour).Format(time.RFC3339)
	old := now.Add(-90 * 24 * time.Hour).Format(time.RFC3339)
	health := `[
		{"guard":"check-plan-shape","runs":40,"fired":3,"cannotAnswer":0,"helper":false},
		{"guard":"check-fix-revert","runs":12,"fired":0,"cannotAnswer":1,"helper":false},
		{"guard":"pair-judge","runs":90,"fired":0,"cannotAnswer":0,"helper":true}
	]`
	incidents := `[
		{"guard":"enforce-agent-baseline","symptom":"Dispatch blocked: first","minutesLost":2,"occurredAt":"` + recent + `"},
		{"guard":"enforce-agent-baseline","symptom":"Dispatch blocked: second","minutesLost":3,"occurredAt":"` + recent + `"},
		{"guard":"check-plan-shape","symptom":"once only","minutesLost":5,"occurredAt":"` + recent + `"},
		{"guard":"check-plan-shape","symptom":"outside the window","minutesLost":5,"occurredAt":"` + old + `"}
	]`
	var query string
	srv := pcServer(t, health, incidents, "[]", &query)

	r := runGuard("prune-candidates", []string{"-days", "30"}, pcEnv(srv.URL))
	if r.rc != 0 {
		t.Fatalf("rc = %d, want 0; out:\n%s", r.rc, r.out)
	}
	for _, want := range []string{
		"never-fired check-fix-revert runs=12",
		"friction enforce-agent-baseline incidents=2 minutes-lost=5 latest: Dispatch blocked: second",
		"never-ran check-guard-calls",
	} {
		if !strings.Contains(r.out, want) {
			t.Errorf("output lacks %q:\n%s", want, r.out)
		}
	}
	for _, unwanted := range []string{
		"check-plan-shape runs=",    // it fired: no never-fired line
		"friction check-plan-shape", // one incident in the window, the other outside it
		"pair-judge",                // a helper refuses nothing
		"never-ran check-fix-revert",
		"never-ran prune-candidates",
	} {
		if strings.Contains(r.out, unwanted) {
			t.Errorf("output carries %q:\n%s", unwanted, r.out)
		}
	}
	from := now.Add(-30 * 24 * time.Hour).Format("2006-01-02")
	if !strings.Contains(query, "from="+from) || !strings.Contains(query, "to=") {
		t.Errorf("guard-health query %q does not open the 30-day window at %s", query, from)
	}
}

func TestPruneCandidatesCannotAnswer(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	r := runGuard("prune-candidates", nil, pcEnv(srv.URL))
	if r.rc != 2 || !strings.Contains(r.err, "flowd") {
		t.Fatalf("rc = %d, want 2 naming flowd; out:\n%s", r.rc, r.out)
	}
	if r := runGuard("prune-candidates", []string{"-days", "0"}, pcEnv(srv.URL)); r.rc != 2 {
		t.Fatalf("-days 0: rc = %d, want 2", r.rc)
	}
}

// TestPruneCandidatesNeverLoaded pins the never-loaded lines: every .md under the checkout's skills/
// with no load in the window, sorted, a file a conditional load set names annotated with that
// set's condition; a loaded file and a non-.md file are not listed; and a window holding no load
// at all prints one line saying so instead of listing every file.
func TestPruneCandidatesNeverLoaded(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, f := range []string{"skills/flow/SKILL.md", "skills/flow/fix.md", "skills/flow/zeta.md", "skills/flow/run.go"} {
		p := filepath.Join(root, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var query string
	srv := pcServer(t, "[]", "[]", `[{"path":"skills/flow/SKILL.md","loads":3},{"path":"rules/x.md","loads":1}]`, &query)
	r := runGuard("prune-candidates", nil, pcEnvAt(srv.URL, root))
	if r.rc != 0 {
		t.Fatalf("rc = %d, want 0; out:\n%s", r.rc, r.out)
	}
	want := "never-loaded skills/flow/fix.md conditional: fix run, cross-repo, sdd, gated review, live check, fix round, optional slots, isolation, pipeline defects, progress, UI\n" +
		"never-loaded skills/flow/zeta.md\n"
	if !strings.HasSuffix(r.stdout, want) || strings.Contains(r.stdout, "SKILL.md") || strings.Contains(r.stdout, "run.go") {
		t.Errorf("stdout:\n%s\nwant suffix:\n%s", r.stdout, want)
	}

	empty := pcServer(t, "[]", "[]", "[]", &query)
	r = runGuard("prune-candidates", nil, pcEnvAt(empty.URL, root))
	if r.rc != 0 || !strings.HasSuffix(r.stdout, "never-loaded: no load recorded in the window\n") || strings.Contains(r.stdout, "never-loaded skills/") {
		t.Errorf("no loads: rc=%d stdout:\n%s", r.rc, r.stdout)
	}
}
