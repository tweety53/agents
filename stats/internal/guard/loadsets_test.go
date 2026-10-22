package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lsWant is what scripts/load-sets.sh at e111a883 printed for lsFixture.
const lsWant = `
always-on (global block + project CLAUDE.md)            31 bytes  ~     7 tok
router (command + SKILL.md + pipeline.md)              137 bytes  ~    34 tok
--- planning session (creating run) ---
  phase files + load directives                          0 bytes  ~     0 tok
  + cited at point of use                                0 bytes  ~     0 tok
  TOTAL definite                                       168 bytes  ~    42 tok
  + conditional (flow-fix/flow-cost/flow-speed, seeded note, resume)       0 bytes  ~     0 tok
  TOTAL incl. cited                                    168 bytes  ~    42 tok
--- implementation session (enters via Resuming at STARTED) ---
  phase files + load directives                          0 bytes  ~     0 tok
  + cited / reviewer templates                           0 bytes  ~     0 tok
  + conditional (fix run, cross-repo, sdd, gated fix, fix round, optional slots, isolation, UI)       0 bytes  ~     0 tok
  TOTAL definite                                       168 bytes  ~    42 tok
  TOTAL incl. cited                                    168 bytes  ~    42 tok
  TOTAL worst case                                     168 bytes  ~    42 tok
--- finish session (merge-and-push: run 1 chained into run 2) ---
  phase files + load directives                          0 bytes  ~     0 tok
  + cited                                                0 bytes  ~     0 tok
  + conditional (finish contracts, registry, gate, base moved, hand fallbacks, follow-ups, state file, config, verdicts)       0 bytes  ~     0 tok
  TOTAL definite                                       168 bytes  ~    42 tok
  TOTAL incl. cited                                    168 bytes  ~    42 tok
--- subagents (each also inherits whatever always-on context the harness gives it) ---
  reviewer: baseline + primary template                  0 bytes  ~     0 tok
  reviewer: baseline + principles + EP                   0 bytes  ~     0 tok
  reviewer: baseline + failure modes                     0 bytes  ~     0 tok
  verifier: baseline + verifier steps (UI runs)          0 bytes  ~     0 tok
  project-configuration.md (cited everywhere)            0 bytes  ~     0 tok
  state-file.md (load before any state r/w)              0 bytes  ~     0 tok
`

// lsFixture builds a repo holding three of the listed files and a global
// file outside it, of known sizes.
func lsFixture(t *testing.T) (repo, global string) {
	dir := t.TempDir()
	repo = filepath.Join(dir, "r")
	for path, n := range map[string]int{"CLAUDE.md": 10, "skills/flow/SKILL.md": 100, "skills/flow-contracts/pipeline.md": 37, "../g.md": 21} {
		p := filepath.Join(repo, path)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, make([]byte, n), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return repo, filepath.Join(dir, "g.md")
}

// TestLoadSets pins load-sets' report against the bash script's own output
// on the same fixture.
func TestLoadSets(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: func(string) string { return "" }}

	t.Run("rows and section headers match the script", func(t *testing.T) {
		t.Parallel()
		repo, global := lsFixture(t)
		r := runGuard("load-sets", []string{repo, global}, env)
		if r.rc != 0 || r.stdout != lsWant[1:] {
			t.Fatalf("rc=%d stdout=\n%s", r.rc, r.stdout)
		}
	})

	t.Run("each absent file is a repo-relative MISSING line per row", func(t *testing.T) {
		t.Parallel()
		repo, global := lsFixture(t)
		r := runGuard("load-sets", []string{repo, global}, env)
		lines := strings.Split(strings.TrimSuffix(r.err, "\n"), "\n")
		flow := 0
		for _, l := range lines {
			if !strings.HasPrefix(l, "load-sets: MISSING ") || strings.Contains(l, repo) {
				t.Fatalf("bad stderr line %q", l)
			}
			if l == "load-sets: MISSING commands-claude/flow.md" {
				flow++
			}
		}
		if len(lines) != 158 || flow != 8 || strings.Contains(r.err, "SKILL.md") {
			t.Fatalf("lines=%d flow.md=%d stderr=\n%s", len(lines), flow, r.err)
		}
	})

	t.Run("exit 0 with every file missing", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		r := runGuard("load-sets", []string{filepath.Join(dir, "none"), filepath.Join(dir, "g.md")}, env)
		if r.rc != 0 || !strings.Contains(r.err, "load-sets: MISSING "+filepath.Join(dir, "g.md")+"\n") || !strings.Contains(r.err, "load-sets: MISSING CLAUDE.md\n") {
			t.Fatalf("rc=%d stderr=%q", r.rc, r.err)
		}
	})

	t.Run("defaults to FLOW_GUARD_REPO_ROOT and HOME, for absent and empty arguments", func(t *testing.T) {
		t.Parallel()
		repo, _ := lsFixture(t)
		home := t.TempDir()
		vars := map[string]string{"FLOW_GUARD_REPO_ROOT": repo, "HOME": home}
		denv := Env{Getenv: func(k string) string { return vars[k] }}
		for _, args := range [][]string{nil, {"", ""}} {
			r := runGuard("load-sets", args, denv)
			if r.rc != 0 || !strings.HasPrefix(r.stdout, "always-on (global block + project CLAUDE.md)            10 bytes") ||
				!strings.Contains(r.err, "load-sets: MISSING "+home+"/.claude/CLAUDE.md\n") {
				t.Fatalf("rc=%d stdout=%q", r.rc, r.stdout)
			}
		}
	})
}
