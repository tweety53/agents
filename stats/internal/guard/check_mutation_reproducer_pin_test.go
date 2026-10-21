package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-mutation-reproducer-pin.sh at 75261910,
// one subtest per ok label. The harness pointed the guard at a fixture tree
// with CHECK_MUTATION_REPRODUCER_PIN_ROOT and matched glob patterns over its
// combined output; here the guard runs in-process with the same override on
// Env, and each `*"x"*` pattern is a substring check.

const (
	mrpRunner = "#!/usr/bin/env bash\n# THE MUTATION-REPRODUCER CONVENTION (KAN-568). Such a reproducer declares\n# itself with the exact line `# mutation-reproducer` within its first 10 lines,\n# and this script then reads it under that convention.\nexec flow-guard run-reproducer \"$@\"\n"
	mrpSource = "package guard\n\nconst (\n\trrMarker     = \"# mutation-reproducer\"\n)\n\nfunc declaresMutation(path string) bool {\n\tfor _, l := range headLines(path, 10) {\n\t\tif l == rrMarker {\n\t\t\treturn true\n\t\t}\n\t}\n\treturn false\n}\n"
	mrpPanel  = "A surviving mutant's reproducer carries the exact line\n`# mutation-reproducer` within its first 10 lines — the declaration\n`run-reproducer.sh` reads as the mutation convention.\n"
)

const (
	mrpRunnerPath = "scripts/run-reproducer.sh"
	mrpSourcePath = "stats/internal/guard/runreproducer.go"
	mrpPanelPath  = "skills/flow/review-panel.md"
)

// mrpCheck is one ok label: the exit code when want is "", else want as a
// substring of the combined output.
type mrpCheck struct {
	label string
	rc    int
	want  string
}

func TestCheckMutationReproducerPin(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		files map[string]string // overrides of the three agreeing fixtures; "" removes
		vars  func(root string) map[string]string
		check []mrpCheck
	}{
		{name: "pin_ok", check: []mrpCheck{
			{label: "pin_ok: exits 0", rc: 0},
			{label: "pin_ok: prints the OK verdict", want: "MUTATION-REPRODUCER-PIN-OK"},
		}},
		{name: "marker_drift_in_prose_fails",
			files: map[string]string{mrpPanelPath: "A surviving mutant's reproducer carries the exact line\n`# mutation-check` within its first 10 lines — the declaration\n`run-reproducer.sh` reads as the mutation convention.\n"},
			check: []mrpCheck{
				{label: "marker_drift_in_prose_fails: exits 1", rc: 1},
				{label: "marker_drift_in_prose_fails: names review-panel.md", want: "review-panel.md"},
			}},
		{name: "window_drift_in_prose_fails",
			files: map[string]string{mrpPanelPath: "A surviving mutant's reproducer carries the exact line\n`# mutation-reproducer` within its first 12 lines — the declaration\n`run-reproducer.sh` reads as the mutation convention.\n"},
			check: []mrpCheck{
				{label: "window_drift_in_prose_fails: exits 1", rc: 1},
				{label: "window_drift_in_prose_fails: names the drifted window statement", want: "no longer states the window"},
			}},
		{name: "comment_drift_fails",
			files: map[string]string{mrpRunnerPath: strings.Replace(mrpRunner, "first 10 lines", "first 12 lines", 1)},
			check: []mrpCheck{
				{label: "comment_drift_fails: exits 1", rc: 1},
				{label: "comment_drift_fails: names run-reproducer.sh", want: "run-reproducer.sh"},
			}},
		{name: "missing_declaration_is_exit_2",
			files: map[string]string{mrpSourcePath: "package guard\n\n// The window is declared elsewhere now.\nconst rrMarker = \"# mutation-reproducer\"\n\nfunc declaresMutation(lines []string) bool {\n\tfor _, l := range lines {\n\t\tif l == rrMarker {\n\t\t\treturn true\n\t\t}\n\t}\n\treturn false\n}\n"},
			check: []mrpCheck{
				{label: "missing_declaration_is_exit_2: exits 2", rc: 2},
				{label: "missing_declaration_is_exit_2: names the declaration line", want: "declaration line"},
			}},
		{name: "empty_root_env_is_exit_2",
			vars:  func(string) map[string]string { return map[string]string{"CHECK_MUTATION_REPRODUCER_PIN_ROOT": ""} },
			check: []mrpCheck{{label: "empty_root_env_is_exit_2: exits 2", rc: 2}}},
		{name: "missing_site_file_is_exit_2",
			files: map[string]string{mrpPanelPath: ""},
			check: []mrpCheck{{label: "missing_site_file_is_exit_2: exits 2", rc: 2}}},
		{name: "multi_site_drift_fails",
			files: map[string]string{mrpRunnerPath: "#!/usr/bin/env bash\n# THE MUTATION-REPRODUCER CONVENTION (KAN-568). Such a reproducer declares\n# itself with the exact line `# mutation-reproducer` within its first 12 lines,\n# and this script then reads it under that convention.\n# Exact line, within the first 10 lines: a line that merely contains the\n# marker is prose, not a declaration.\nexec flow-guard run-reproducer \"$@\"\n"},
			check: []mrpCheck{
				{label: "multi_site_drift_fails: exits 1", rc: 1},
				{label: "multi_site_drift_fails: names the drifted statement at its line", want: `states the window as "first 12 lines"`},
			}},
		{name: "wrapped_window_drift_fails",
			files: map[string]string{mrpPanelPath: mrpPanel + "\nOne `# demonstrates:` line per cited location, within the script's first 12\nlines — the same window the `# mutation-reproducer` declaration reads.\n"},
			check: []mrpCheck{
				{label: "wrapped_window_drift_fails: exits 1", rc: 1},
				{label: "wrapped_window_drift_fails: names the wrapped near-miss", want: `states the window as "first 12 lines"`},
			}},
		{name: "near_miss_marker_in_prose_fails",
			files: map[string]string{mrpPanelPath: mrpPanel + "\nA reproducer declares it with the exact `# mutation-check` line.\n"},
			check: []mrpCheck{
				{label: "near_miss_marker_in_prose_fails: exits 1", rc: 1},
				{label: "near_miss_marker_in_prose_fails: names the near-miss span", want: "`# mutation-check`"},
			}},
		{name: "raw_string_marker_is_exit_2",
			files: map[string]string{mrpSourcePath: "package guard\n\nconst rrMarker = `# mutation-reproducer`\n\nfunc declaresMutation(path string) bool {\n\tfor _, l := range headLines(path, 10) {\n\t\tif l == rrMarker {\n\t\t\treturn true\n\t\t}\n\t}\n\treturn false\n}\n"},
			check: []mrpCheck{
				{label: "raw_string_marker_is_exit_2: exits 2", rc: 2},
				{label: "raw_string_marker_is_exit_2: names the extraction failure", want: "could not be extracted"},
			}},
	} {
		root := t.TempDir()
		files := map[string]string{mrpRunnerPath: mrpRunner, mrpSourcePath: mrpSource, mrpPanelPath: mrpPanel}
		for p, body := range tc.files {
			files[p] = body
		}
		for p, body := range files {
			if body == "" {
				mkdir(t, filepath.Dir(filepath.Join(root, p)))
				continue
			}
			writeFile(t, filepath.Join(root, p), body)
		}
		vars := map[string]string{"CHECK_MUTATION_REPRODUCER_PIN_ROOT": root}
		if tc.vars != nil {
			vars = tc.vars(root)
		}
		fn := Registry["check-mutation-reproducer-pin"]
		if fn == nil {
			t.Fatal("check-mutation-reproducer-pin is not registered")
		}
		var out bytes.Buffer
		rc := fn(nil, crEnv(root, vars), &out, &out)
		for _, c := range tc.check {
			t.Run(c.label, func(t *testing.T) {
				t.Parallel()
				if c.want == "" && rc != c.rc {
					t.Errorf("rc=%d, want %d\n%s", rc, c.rc, out.String())
				}
				if c.want != "" && !strings.Contains(out.String(), c.want) {
					t.Errorf("output lacks %q:\n%s", c.want, out.String())
				}
			})
		}
	}
}

// TestCheckMutationReproducerPinOutput pins the exact lines the bash printed
// at 75261910 for the clean run, a drift of each kind (anchored, wrapped and
// near-miss) and a symlinked site — captured from that script, not derived.
func TestCheckMutationReproducerPinOutput(t *testing.T) {
	t.Parallel()
	run := func(t *testing.T, panel string) (int, string, string) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, mrpRunnerPath), mrpRunner)
		writeFile(t, filepath.Join(root, mrpSourcePath), mrpSource)
		writeFile(t, filepath.Join(root, mrpPanelPath), panel)
		var stdout, stderr bytes.Buffer
		rc := Registry["check-mutation-reproducer-pin"](nil, crEnv(root, map[string]string{"CHECK_MUTATION_REPRODUCER_PIN_ROOT": root}), &stdout, &stderr)
		return rc, strings.ReplaceAll(stdout.String(), root, "R"), strings.ReplaceAll(stderr.String(), root, "R")
	}
	t.Run("clean", func(t *testing.T) {
		t.Parallel()
		rc, out, errOut := run(t, mrpPanel)
		want := "MUTATION-REPRODUCER-PIN-OK: R — marker '# mutation-reproducer' window 10 pinned across stats/internal/guard/runreproducer.go, scripts/run-reproducer.sh and skills/flow/review-panel.md\n"
		if rc != 0 || out != want || errOut != "" {
			t.Errorf("rc=%d\nstdout=%q\nstderr=%q", rc, out, errOut)
		}
	})
	t.Run("drift", func(t *testing.T) {
		t.Parallel()
		rc, out, _ := run(t, "x first 10\nlines `# mutation-reproducer`\nfirst 9 lines `# mutation-x`\n")
		want := "R/skills/flow/review-panel.md:3: no longer states the window as \"first 10 lines\" — it drifted from stats/internal/guard/runreproducer.go's canonical declaration (window: 10); move every site together\n" +
			"R/skills/flow/review-panel.md:3: carries the backticked mutation-marker span `# mutation-x`, which is not the canonical marker `# mutation-reproducer`\n" +
			"R/skills/flow/review-panel.md:3: states the window as \"first 9 lines\" — a near-miss of the canonical window 10; every window statement must read \"first 10 lines\"\n" +
			"MUTATION-REPRODUCER-PIN-INVALID: R — 3 violation(s)\n"
		if rc != 1 || out != want {
			t.Errorf("rc=%d\nstdout=%q", rc, out)
		}
	})
	t.Run("symlinked site", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		writeFile(t, filepath.Join(root, mrpRunnerPath), mrpRunner)
		writeFile(t, filepath.Join(root, mrpSourcePath), mrpSource)
		writeFile(t, filepath.Join(root, "real.md"), mrpPanel)
		mkdir(t, filepath.Join(root, "skills", "flow"))
		if err := os.Symlink(filepath.Join(root, "real.md"), filepath.Join(root, mrpPanelPath)); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		rc := Registry["check-mutation-reproducer-pin"](nil, crEnv(root, map[string]string{"CHECK_MUTATION_REPRODUCER_PIN_ROOT": root}), &stdout, &stderr)
		if want := root + "/" + mrpPanelPath + ":0: is a symlink — a pinned site must be a real file, never a symlink\n"; rc != 2 || stderr.String() != want {
			t.Errorf("rc=%d stderr=%q", rc, stderr.String())
		}
	})
}
