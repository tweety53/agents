package plans

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The security slot (kan-1125): a project declares `## security paths` in
// .flow/project.md — one line of comma-separated globs, matched exactly as
// `ui paths` are (GlobRegexp). plan-check adds a `security` bundle to every
// decision; the decider never picks the slot itself. A task file matching a
// declared glob leaves the bundle without a pair, so its dispatch takes the pair
// judge's; with none matching it runs on the panel-reviewer role's cheapest pair.

// spMatchingFile returns the first task file that matches a declared
// `## security paths` glob, or "" when none does, or the section is absent.
func spMatchingFile(projectFile string, tasks []pdbTask) string {
	var globs []string
	for _, l := range strings.Split(core.ProjectSection(projectFile, "security paths"), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			for _, g := range strings.Split(l, ",") {
				if g = core.TrimGlobElement(g); g != "" {
					globs = append(globs, g)
				}
			}
			break
		}
	}
	for _, t := range tasks {
		for _, f := range t.files {
			p := core.BytesAsRunes(strings.TrimPrefix(filepath.ToSlash(f), "./"))
			for _, g := range globs {
				if core.GlobRegexp(g).MatchString(p) {
					return f
				}
			}
		}
	}
	return ""
}

// spAddSecurity appends the `security` bundle to d.Panel unless a bundle carries the slot already:
// pairless when a task file matches `## security paths`, on base's cheapest pair otherwise.
func spAddSecurity(d *core.PlanDecision, projectFile string, tasks []pdbTask, base []string) {
	for _, b := range d.Panel {
		if slices.Contains(b.Slots, "security") {
			return
		}
	}
	if f := spMatchingFile(projectFile, tasks); f != "" {
		d.Panel = append(d.Panel, core.DecisionEntry{Slots: []string{"security"}, Reason: "task file " + f + " matches `## security paths`"})
		return
	}
	d.Panel = append(d.Panel, core.DecisionEntry{Slots: []string{"security"}, Pair: spCheapest(base), Reason: "no task file matches `## security paths`: the cheapest pair"})
}

// spCheapest is the cheapest of base: the cheapest model, then the lowest effort, the first listed
// on a tie; "" for no pair.
// ponytail: models ranked by name (haiku < sonnet < opus < fable, any other last), a cost table if a model outside them is admitted.
func spCheapest(base []string) string {
	efforts := []string{"low", "medium", "high", "xhigh", "max"}
	models := []string{"haiku", "sonnet", "opus", "fable"}
	rank := func(p string) (int, int) {
		c, _ := pjParse(p)
		m := slices.Index(models, c.model)
		if m < 0 {
			m = len(models)
		}
		return m, slices.Index(efforts, c.effort)
	}
	best := ""
	for _, p := range base {
		m, e := rank(p)
		if bm, be := rank(best); best == "" || m < bm || m == bm && e < be {
			best = p
		}
	}
	return best
}
