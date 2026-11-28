package guard

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// pdgRepositories maps each line number (1-based) to the repository heading
// in force there: a `## <name>` heading whose name pdgDeclaredRepos holds.
func pdgRepositories(env core.Env, path string, lines []string) []string {
	repos := pdgDeclaredRepos(psAbs(env, path))
	repo := make([]string, len(lines)+1)
	cur := ""
	tcfUnfenced(lines, func(i int, line string) bool {
		if h, ok := strings.CutPrefix(line, "## "); ok && repos[strings.TrimSpace(h)] {
			cur = strings.TrimSpace(h)
		}
		repo[i+1] = cur
		return false
	})
	// Fenced lines were skipped: carry the last heading over them.
	for i := 1; i < len(repo); i++ {
		if repo[i] == "" {
			repo[i] = repo[i-1]
		}
	}
	return repo
}

// pdgDeclaredRepos is the basename of every `Repo root` the `## apps` table
// of the nearest `.flow/project.md` above the plan declares, save a root lying
// inside another declared root (agents' `stats`), which is part of that
// repository and no repository of its own. A plan under no project file
// declares none, so every task is under "", one repository.
func pdgDeclaredRepos(plan string) map[string]bool {
	for d := filepath.Dir(plan); ; d = filepath.Dir(d) {
		cfg := filepath.Join(d, ".flow", "project.md")
		if _, err := os.Stat(cfg); err == nil {
			roots := tcfAppRoots(core.ProjectSection(cfg, "apps"))
			repos := map[string]bool{}
			for _, r := range roots {
				if !slices.ContainsFunc(roots, func(o string) bool { return strings.HasPrefix(r, o+"/") }) {
					repos[filepath.Base(r)] = true
				}
			}
			return repos
		}
		if d == filepath.Dir(d) {
			return nil
		}
	}
}
