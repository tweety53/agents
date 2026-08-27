// Package stages holds the documented stage vocabulary for every flow
// pipeline command -- the keys a `flow stage begin`/`stage end` mark is
// allowed to carry -- plus the small set of constants the daemon and the
// CLI must agree on to interpret a stage mark's effects, such as
// SyntheticChangeUpdatedBy (synthetic.go). Those constants are not stage
// vocabulary themselves, but they exist only because a stage mark does,
// and one package for both keeps the daemon's writer and the CLI's reader
// from ever drifting onto two copies of the same literal.
//
// The vocabulary is not authored here first: it is a transcription of
// README.md's "Level 1 -- the stages of each command" table, the one place
// this pipeline's stages are documented. TestStagesMatchReadmeLevelOne
// (names_test.go) re-derives the same table mechanically from README.md at
// test time and fails if this package disagrees with it, so the database's
// vocabulary and the documentation stay one list rather than two that can
// drift -- exactly the property task 5's TestProjectKeyMatchesStateFileContract
// already holds for the project-key recipe.
//
// A stage has two identifiers that deliberately do different jobs
// (design.md, "a declared key, and a name that may change"):
//
//   - Key is what a mark carries, what the store groups by, and what never
//     changes. It is declared here, never derived from Name -- a key
//     slugified from the name would change every time the name is
//     reworded, which defeats the entire reason a key exists.
//   - Name is prose for a human reading a dashboard or a mark's rejection
//     message. It may be reworded freely without splitting a stage's
//     recorded history, because nothing is ever keyed on it.
package stages

import (
	"fmt"
	"strings"
)

// Command is a flow pipeline command name, exactly as README.md's Level 1
// table spells it (including the leading slash).
type Command string

const (
	Flow Command = "/flow"
)

// Stage is one documented pipeline stage: a stable key, a human-readable
// name, and every command that runs it.
//
// Commands is namespaced by the command that *defines* the stage.
type Stage struct {
	Key      string
	Name     string
	Commands []Command
}

// Table is the ordered, hand-authored transcription of README.md's
// "Level 1 -- the stages of each command" stages table. Row order here
// matches the README table's row order; TestStagesMatchReadmeLevelOne
// checks the two are identical, not merely set-equal.
var Table = []Stage{
	// /flow -- mints its own flow.* namespace rather than reusing start./do./finish.,
	// per design.md's flow-rename-content-split (see README.md's Level 1 table for why).
	{Key: "flow.kickoff", Name: "Kickoff — write `STARTED`", Commands: []Command{Flow}},
	{Key: "flow.resolve-change", Name: "Resolve the change", Commands: []Command{Flow}},
	{Key: "flow.brainstorm", Name: "Brainstorm ▸", Commands: []Command{Flow}},
	{Key: "flow.design-approval", Name: "Design approval", Commands: []Command{Flow}},
	{Key: "flow.create-artifacts", Name: "Create the spectre artifacts", Commands: []Command{Flow}},
	{Key: "flow.writing-plans", Name: "Writing-plans ▸", Commands: []Command{Flow}},
	{Key: "flow.load-context", Name: "Load context and validate the plan", Commands: []Command{Flow}},
	{Key: "flow.isolate-workspace", Name: "Isolate the workspace (first run only)", Commands: []Command{Flow}},
	{Key: "flow.document-fix", Name: "Document the fix (re-runs only)", Commands: []Command{Flow}},
	{Key: "flow.sdd-tdd", Name: "SDD + TDD per task ▸", Commands: []Command{Flow}},
	{Key: "flow.review-panel", Name: "The review panel ▸", Commands: []Command{Flow}},
	{Key: "flow.verify", Name: "Verify: workspace isolation, lint and test", Commands: []Command{Flow}},
	{Key: "flow.stage-diff", Name: "Stage, excluding the planning paths", Commands: []Command{Flow}},
	{Key: "flow.run-instructions", Name: "Resolve the run instructions", Commands: []Command{Flow}},
	{Key: "flow.write-in-progress", Name: "Write `IN_PROGRESS`", Commands: []Command{Flow}},
	{Key: "flow.preflight", Name: "Preflight verdict (decides run 1 vs run 2) ▸", Commands: []Command{Flow}},
	{Key: "flow.unfinished-work-gate", Name: "Unfinished-work gate (run 1) ▸", Commands: []Command{Flow}},
	{Key: "flow.landing-question", Name: "The landing question (run 1)", Commands: []Command{Flow}},
	{Key: "flow.preserve-sessions", Name: "Preserve the session records (run 1)", Commands: []Command{Flow}},
	{Key: "flow.commit-two", Name: "Two commits, implementation first (run 1)", Commands: []Command{Flow}},
	{Key: "flow.landing-routes", Name: "The landing routes, including moving the issue to In Review (run 1) ▸", Commands: []Command{Flow}},
	{Key: "flow.verify-merge", Name: "Verify the merge (run 2)", Commands: []Command{Flow}},
	{Key: "flow.sync-archive", Name: "Position the checkout and archive (run 2)", Commands: []Command{Flow}},
	{Key: "flow.commit-archive", Name: "Commit the archive (run 2)", Commands: []Command{Flow}},
	{Key: "flow.cleanup", Name: "Cleanup (run 2) ▸", Commands: []Command{Flow}},
	{Key: "flow.verify-cleanup", Name: "Verify the cleanup (run 2)", Commands: []Command{Flow}},
	{Key: "flow.write-finished", Name: "Write `FINISHED` (run 2)", Commands: []Command{Flow}},
	{Key: "flow.self-review", Name: "Self-review (run 2)", Commands: []Command{Flow}},
	{Key: "flow.push-archive", Name: "Push the archive branch and open its PR (run 2)", Commands: []Command{Flow}},
}

// byKey indexes Table by Key. byCommand indexes Table by every command that
// runs each stage, preserving Table's row order within each command's
// slice. Both are built once, from the literal Table above, and panic on a
// duplicate key: a duplicate is not a state this package can recover from
// meaningfully, since it would silently merge two different stages in
// every statistic (tasks.md, "Uniqueness is enforced, not assumed").
// TestStagesMatchReadmeLevelOne (names_test.go) is the loud, test-time half
// of this same guard, checked against README.md rather than against
// Table's own internal consistency.
var (
	byKey     map[string]Stage
	byCommand map[Command][]Stage
)

func init() {
	byKey = make(map[string]Stage, len(Table))
	byCommand = make(map[Command][]Stage)
	for _, s := range Table {
		if _, dup := byKey[s.Key]; dup {
			panic(fmt.Sprintf("stages: duplicate key %q in Table -- two stages sharing a key would merge two different things in every statistic", s.Key))
		}
		byKey[s.Key] = s
		for _, c := range s.Commands {
			byCommand[c] = append(byCommand[c], s)
		}
	}
}

// Valid reports whether key is a documented stage key of command.
func Valid(command Command, key string) bool {
	s, ok := byKey[key]
	if !ok {
		return false
	}
	for _, c := range s.Commands {
		if c == command {
			return true
		}
	}
	return false
}

// Name returns the documented human-readable name for key, and whether key
// is documented at all. This is the one place the interface should ever go
// from a stored key to prose -- a mark itself never carries a name, only a
// key.
func Name(key string) (string, bool) {
	s, ok := byKey[key]
	if !ok {
		return "", false
	}
	return s.Name, true
}

// ErrUnknownStage is returned by Validate when key is not a documented
// stage key of command -- either because command itself runs no documented
// stages (including `/flow-status`, which marks none at all) or because
// key is not among them.
type ErrUnknownStage struct {
	Command Command
	Stage   string
}

func (e *ErrUnknownStage) Error() string {
	alternatives := byCommand[e.Command]
	if len(alternatives) == 0 {
		return fmt.Sprintf("flow: %q names no documented stage key for %s -- %s has no documented stages in README.md's Level 1 table",
			e.Stage, e.Command, e.Command)
	}
	keys := make([]string, len(alternatives))
	for i, s := range alternatives {
		keys[i] = s.Key
	}
	return fmt.Sprintf("flow: %q is not a documented stage key of %s -- documented stage keys are: %s",
		e.Stage, e.Command, strings.Join(keys, "; "))
}

// Validate returns nil when key is a documented stage key of command, and
// an *ErrUnknownStage naming the documented alternatives otherwise. This is
// the check that keeps a mark from ever recording an undocumented stage: a
// mark naming a key absent from README.md's Level 1 table is rejected
// here, before it ever reaches the store.
func Validate(command Command, key string) error {
	if Valid(command, key) {
		return nil
	}
	return &ErrUnknownStage{Command: command, Stage: key}
}
