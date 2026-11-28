package guard

import (
	"path/filepath"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The security slot (kan-1125): a project declares `## security paths` in
// .flow/project.md — one line of comma-separated globs, matched exactly as
// `ui paths` are (vtGlobRegexp). plan-check adds a `security` bundle to the
// decision when any task's **Files:** matches one; the decider never picks the
// slot itself, so a plan touching no declared path never pays for it.

// spMatchingFile returns the first task file that matches a declared
// `## security paths` glob, or "" when none does, or the section is absent.
func spMatchingFile(projectFile string, tasks []pdbTask) string {
	var globs []string
	for _, l := range strings.Split(core.ProjectSection(projectFile, "security paths"), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			for _, g := range strings.Split(l, ",") {
				if g = trimGlobElement(g); g != "" {
					globs = append(globs, g)
				}
			}
			break
		}
	}
	for _, t := range tasks {
		for _, f := range t.files {
			p := vtBytesAsRunes(strings.TrimPrefix(filepath.ToSlash(f), "./"))
			for _, g := range globs {
				if vtGlobRegexp(g).MatchString(p) {
					return f
				}
			}
		}
	}
	return ""
}

// spAddSecurity appends the `security` bundle to d.Panel when a task file
// matches `## security paths` and no bundle carries the slot already.
func spAddSecurity(d *core.PlanDecision, projectFile string, tasks []pdbTask) {
	for _, b := range d.Panel {
		for _, s := range b.Slots {
			if s == "security" {
				return
			}
		}
	}
	if f := spMatchingFile(projectFile, tasks); f != "" {
		d.Panel = append(d.Panel, core.DecisionEntry{Slots: []string{"security"}, Reason: "task file " + f + " matches `## security paths`"})
	}
}
