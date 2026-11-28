package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// TestMergePanelFindings: two roles' files merge into findings-<round>.jsonl with round, slot,
// dispatch-seq and status added; a location two roles cite is a duplicate; a missing file is 1.
func TestMergePanelFindings(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	sdd := filepath.Join(wt, ".superpowers/sdd")
	writeFile(t, filepath.Join(sdd, "findings-1-primary.jsonl"), `{"task":"2","severity":"Important","location":"a.go:3","note":"x"}`+"\n")
	writeFile(t, filepath.Join(sdd, "findings-1-principles-unread.jsonl"), `{"severity":"Minor","location":"a.go:3","note":"y"}`+"\n"+`{"severity":"Minor","location":"b.go:9","note":"z"}`+"\n")
	writeFile(t, filepath.Join(sdd, "findings-1-test-sanity-gymie.jsonl"), "")
	r := runGuard("merge-panel-findings", []string{wt, "1", "primary=4", "principles-unread=5", "test-sanity-gymie=6"}, core.Env{})
	if r.rc != 0 {
		t.Fatalf("rc %d: %s", r.rc, r.out)
	}
	for _, want := range []string{
		"finding: primary#1 — Important — a.go:3 — 2 — x",
		"finding: principles#2 — Minor — b.go:9 — none — z",
		"duplicate: a.go:3 — primary#1, principles#1",
	} {
		if !strings.Contains(r.stdout, want) {
			t.Errorf("stdout lacks %q:\n%s", want, r.stdout)
		}
	}
	if got := panelSlotOf("test-sanity-gymie"); got != "test-sanity" {
		t.Errorf("panelSlotOf(test-sanity-gymie) = %q", got)
	}
	if strings.Contains(r.stdout, "duplicate: b.go:9") {
		t.Errorf("single-role location reported as duplicate:\n%s", r.stdout)
	}
	got, _ := os.ReadFile(filepath.Join(sdd, "findings-1.jsonl"))
	want := `{"dispatch-seq":4,"location":"a.go:3","note":"x","round":1,"severity":"Important","slot":"primary","status":"open","task":"2"}`
	if !strings.HasPrefix(string(got), want+"\n") || strings.Count(string(got), "\n") != 3 || !strings.Contains(string(got), `"slot":"principles"`) {
		t.Errorf("findings-1.jsonl:\n%s", got)
	}
	if r := runGuard("merge-panel-findings", []string{wt, "1", "security=6"}, core.Env{}); r.rc != 1 {
		t.Errorf("missing file: rc %d, want 1", r.rc)
	}
}
