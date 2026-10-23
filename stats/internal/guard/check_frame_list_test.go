package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of the check-frame-list contract (scripts/check-frame-list.sh):
// the declared list is the mockups dir's frames design.md names as whole
// tokens, and each declared frame needs a sidecar line and a report line.
func TestCheckFrameList(t *testing.T) {
	const design = "Frames `Q1` and Q2 (see mockups/M1.png); Q100 and UNCITED-old are other ids.\n"
	cases := []struct {
		name    string
		sidecar string // "" writes no sidecar
		report  string // %SPEC% is replaced by the spec path
		rc      int
		want    string
	}{
		{"every declared frame captured and reported",
			"q1 Q1\nq2 Q2\n# comment\nm1 M1\n",
			"- spec: %SPEC%\n- Q1: /c/q1.png diff=0.1 — match\n- Q2 bands: 3/3 paired\n- M1: /c/m1.png diff=0.2 — match\n- frames: 3/3 — none absent\n",
			0, "FRAME-LIST-COMPLETE"},
		{"a declared frame named absent with its reason",
			"q1 Q1\nq2 Q2\nm1 M1\n",
			"- spec: %SPEC%\n- Q1: /c/q1.png diff=0.1 — match\n- M1: /c/m1.png diff=0.2 — match\n- frames: 2/3 — Q2: reached only by inference from Q1\n",
			0, "3 declared"},
		{"a frame named absent with its reason needs no sidecar line",
			"q1 Q1\nm1 M1\n",
			"- spec: %SPEC%\n- Q1: /c/q1.png diff=0.1 — match\n- M1: /c/m1.png diff=0.2 — match\n- frames: 2/3 — Q2: reached only by inference from Q1\n",
			0, "3 declared"},
		{"a reason citing a declared frame does not mark it absent",
			"m1 M1\n",
			"- spec: %SPEC%\n- M1: /c/m1.png diff=0.2 — match\n- frames: 1/3 — Q2: reached only by inference from Q1\n",
			1, "no sidecar line: Q1; no report line: Q1"},
		// KAN-549: design.md declared more frames than the verifier counted.
		{"verifier counted fewer than declared",
			"q1 Q1\nq2 Q2\nm1 M1\n",
			"- spec: %SPEC%\n- Q1: /c/q1.png diff=0.1 — match\n- frames: 1/1\n",
			1, "frames 1/1 below 3 declared; no report line: M1; no report line: Q2"},
		{"no sidecar line for a declared frame",
			"q1 Q1\nq2 Q2\n",
			"- spec: %SPEC%\n- Q1: x\n- Q2: x\n- M1: x\n- frames: 3/3\n",
			1, "no sidecar line: M1"},
		{"no sidecar file at all",
			"",
			"- spec: %SPEC%\n- Q1: x\n- Q2: x\n- M1: x\n- frames: 3/3\n",
			1, "no sidecar line: M1; no sidecar line: Q1; no sidecar line: Q2"},
		// kan-1008's self-review: one per-frame line naming two frames reports both.
		{"a per-frame line naming frames joined by a slash",
			"q1 Q1\nq2 Q2\nm1 M1\n",
			"- spec: %SPEC%\n- Q1 / Q2 / M1 bands: 3/3 paired — match\n- frames: 3/3 — none absent\n",
			0, "FRAME-LIST-COMPLETE"},
		{"no frames line",
			"q1 Q1\nq2 Q2\nm1 M1\n",
			"- spec: %SPEC%\n- Q1: x\n- Q2: x\n- M1: x\n",
			1, "no frames line"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			mockups := filepath.Join(dir, "mockups")
			if err := os.Mkdir(mockups, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, f := range []string{"Q1.png", "Q2.png", "Q10.png", "M1.png", "UNCITED.png", "notes.txt"} {
				if err := os.WriteFile(filepath.Join(mockups, f), nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			spec := filepath.Join(dir, "kan-1.spec.ts")
			if tc.sidecar != "" {
				if err := os.WriteFile(spec+".mockups", []byte(tc.sidecar), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			designPath := filepath.Join(dir, "design.md")
			report := filepath.Join(dir, "verify-report-visual-verify.md")
			if err := os.WriteFile(designPath, []byte(design), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(report, []byte(strings.ReplaceAll(tc.report, "%SPEC%", spec)), 0o644); err != nil {
				t.Fatal(err)
			}
			r := runGuard("check-frame-list", []string{designPath, mockups, report}, Env{Getenv: os.Getenv})
			if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
				t.Fatalf("rc=%d out=%q; want rc=%d containing %q", r.rc, r.out, tc.rc, tc.want)
			}
		})
	}
	dir := t.TempDir()
	empty := filepath.Join(dir, "design.md")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{nil, {empty, dir}, {filepath.Join(dir, "absent.md"), dir, empty}, {empty, empty, empty}, {empty, dir, filepath.Join(dir, "absent.md")}} {
		if r := runGuard("check-frame-list", args, Env{Getenv: os.Getenv}); r.rc != 2 {
			t.Fatalf("args %q: rc=%d out=%q; want exit 2", args, r.rc, r.out)
		}
	}
}
