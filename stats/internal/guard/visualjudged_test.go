package guard

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// TestVisualJudgedCarriesOnlySpecificVerdicts pins the re-judge rule: a frame
// is carried only while its capture is byte-identical to the one a report
// judged on a line of its own, with no row excused by a decision naming more
// than one frame.
func TestVisualJudgedCarriesOnlySpecificVerdicts(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pairs := []vvmPair{}
	for _, f := range []string{"A1", "A2", "A3", "A4"} {
		path := filepath.Join(dir, f+".png")
		if err := os.WriteFile(path, []byte("png "+f), 0o644); err != nil {
			t.Fatal(err)
		}
		pairs = append(pairs, vvmPair{Frame: f, Capture: path})
	}
	design := []byte("## Decisions\n\n### one\n\n**ID:** shared-bar\n**Chosen:** A2 and A3 keep the old bar\n\n### two\n\n**ID:** own-bar\n**Chosen:** A1 keeps its bar\n**Considered:** A2 too\n")
	report := []byte("## Verdict\n\n- A4: x diff=0.01\n\n## Report\n\n" +
		"- A1: /c/A1.png diff=0.01\n- A1 over-tolerance: 1 — #1: decision own-bar\n" +
		"- A2: /c/A2.png diff=0.01\n- A2 over-tolerance: 1 — #1: decision shared-bar\n" +
		"- A3: /c/A3.png diff=0.01\n")
	judged := judgedAfter(map[string]string{}, pairs, report, design)
	if err := os.WriteFile(pairs[2].Capture, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := vvpCarried(judged, pairs)
	// (a) A1: identical hash, its own line → carried.
	if !slices.Contains(got, "A1") {
		t.Errorf("A1 (identical, specific) not carried: %v", got)
	}
	// (b) A2: excused under a decision naming A2 and A3 → judged.
	if slices.Contains(got, "A2") {
		t.Errorf("A2 (collective decision) carried: %v", got)
	}
	// (c) A3: capture changed since it was judged → judged.
	if slices.Contains(got, "A3") {
		t.Errorf("A3 (changed hash) carried: %v", got)
	}
	// (d) A4: only the Verdict block names it, no Report line → judged.
	if slices.Contains(got, "A4") {
		t.Errorf("A4 (not covered) carried: %v", got)
	}
	// A later report that does not cover A1 keeps it while its capture holds,
	// and drops A3, whose capture changed.
	next := judgedAfter(judged, pairs, []byte("## Report\n"), design)
	if next["A1"] == "" || next["A3"] != "" {
		t.Errorf("after an uncovering report: %v", next)
	}
	sdd := t.TempDir()
	if err := writeJudged(sdd, next); err != nil {
		t.Fatal(err)
	}
	if back := readJudged(sdd); back["A1"] != next["A1"] || len(back) != len(next) {
		t.Errorf("round trip %v, want %v", back, next)
	}
}
